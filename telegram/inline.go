// Package telegram — inline mode.
//
// Telegram inline mode must be enabled in BotFather via /setinline and a
// placeholder (e.g. "Search compendium…"). Without it, inline queries are
// never delivered even when AllowedUpdateInlineQuery is set.
package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/middleware"
	"villum/search"
)

const inlineMaxResults = 20

const inlineAskTTL = 15 * time.Minute

var inlineAskStore = struct {
	sync.Mutex
	m map[string]inlineAskEntry
}{m: map[string]inlineAskEntry{}}

type inlineAskEntry struct {
	query   string
	expires time.Time
}

func newInlineAskToken() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func putInlineAskQuery(query string) string {
	token := newInlineAskToken()
	inlineAskStore.Lock()
	defer inlineAskStore.Unlock()
	pruneInlineAskStoreLocked()
	inlineAskStore.m[token] = inlineAskEntry{query: query, expires: time.Now().Add(inlineAskTTL)}
	return token
}

func pruneInlineAskStoreLocked() {
	now := time.Now()
	for k, v := range inlineAskStore.m {
		if now.After(v.expires) {
			delete(inlineAskStore.m, k)
		}
	}
}

func lookupInlineAskQuery(token string) (string, bool) {
	inlineAskStore.Lock()
	defer inlineAskStore.Unlock()
	pruneInlineAskStoreLocked()
	e, ok := inlineAskStore.m[token]
	if !ok {
		return "", false
	}
	if time.Now().After(e.expires) {
		delete(inlineAskStore.m, token)
		return "", false
	}
	return e.query, true
}

func handleInlineQuery(ctx context.Context, q *tgmodels.InlineQuery) {
	if q == nil {
		return
	}
	trimmed := strings.TrimSpace(q.Query)
	if runeCount(trimmed) < 2 {
		results := []tgmodels.InlineQueryResult{
			&tgmodels.InlineQueryResultArticle{
				ID:    "hint",
				Title: "Type at least 2 characters to search the compendium",
				InputMessageContent: &tgmodels.InputTextMessageContent{
					MessageText: "Type at least 2 characters to search the compendium",
				},
			},
		}
		answerInline(ctx, q.ID, results, nil)
		return
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	results, err := search.SearchCompendium(timeoutCtx, db.DB, search.CompendiumParams{
		Query:         trimmed,
		Limit:         inlineMaxResults,
		Reranker:      search.DefaultRerank,
		FuzzyFallback: true,
	})
	if err != nil {
		answerInline(ctx, q.ID, nil, nil)
		return
	}
	articles := buildInlineResults(timeoutCtx, results)
	var button *tgmodels.InlineQueryResultsButton
	if len(results) > 0 {
		token := putInlineAskQuery(trimmed)
		button = &tgmodels.InlineQueryResultsButton{
			Text:           "Ask AI about this",
			StartParameter: "ask_" + token,
		}
	}
	answerInline(ctx, q.ID, articles, button)
}

// answerInline sends an inline answer with a personal cache and logs failures
// instead of discarding them, so rate limits and transport errors are visible.
func answerInline(ctx context.Context, queryID string, results []tgmodels.InlineQueryResult, button *tgmodels.InlineQueryResultsButton) {
	if err := AnswerInlineQuery(ctx, queryID, results, 30, true, button); err != nil {
		middleware.LogWarn("telegram", "answer inline query failed", "error", err, "query_id", queryID)
	}
}

func buildInlineResults(ctx context.Context, results []search.CompendiumResult) []tgmodels.InlineQueryResult {
	out := make([]tgmodels.InlineQueryResult, 0, len(results))
	for i, r := range results {
		title := r.Name + " (" + r.Type + ")"
		var desc string
		if r.Subtype != "" {
			desc = r.Subtype
		} else if r.Type == "spell" && r.Level != 0 {
			desc = fmt.Sprintf("Level %d", r.Level)
		} else {
			desc = r.Type
		}
		detail, _ := compendiumDetailForInline(ctx, r)
		msgText := buildInlineMessageText(r, detail)
		// Cap at ~3500 runes
		if runeCount(msgText) > 3500 {
			msgText = truncateRunes(msgText, 3499) + "…"
		}
		id := fmt.Sprintf("%s:%d:%d", r.Type, r.ID, i)
		article := &tgmodels.InlineQueryResultArticle{
			ID:          id,
			Title:       truncateRunes(title, 80),
			Description: truncateRunes(desc, 80),
			InputMessageContent: &tgmodels.InputTextMessageContent{
				MessageText:        msgText,
				ParseMode:          tgmodels.ParseModeHTML,
				LinkPreviewOptions: &tgmodels.LinkPreviewOptions{IsDisabled: boolPtr(true)},
			},
		}
		out = append(out, article)
	}
	return out
}

func compendiumDetailForInline(ctx context.Context, r search.CompendiumResult) (*search.CompendiumDetail, error) {
	// Bound each detail load by the parent inline budget (handleInlineQuery's 8s
	// deadline) so a slow query cannot exceed Telegram's answer window.
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	d, err := search.LoadCompendiumDetail(ctx, db.DB, r.Type, r.ID)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func buildInlineMessageText(r search.CompendiumResult, d *search.CompendiumDetail) string {
	var b strings.Builder
	b.WriteString("<b>" + escapeHTML(r.Name) + "</b> — " + escapeHTML(r.Type))
	if d != nil {
		if d.Subtype != "" {
			b.WriteString("\n" + escapeHTML(d.Subtype))
		}
		if d.Description != "" {
			desc := d.Description
			if runeCount(desc) > 3000 {
				desc = truncateRunes(desc, 2999) + "…"
			}
			b.WriteString("\n\n" + escapeHTML(desc))
		}
		// Append a few fields if present
		append := func(label, val string) {
			if val != "" {
				b.WriteString("\n<b>" + escapeHTML(label) + ":</b> " + escapeHTML(val))
			}
		}
		append("School", d.School)
		if d.Level != 0 {
			append("Level", strconv.Itoa(d.Level))
		}
		append("Casting Time", d.CastingTime)
		append("Range", d.Range)
		append("Components", d.Components)
		append("Duration", d.Duration)
		append("Classes", d.Classes)
		append("Category", d.Category)
		append("Cost", d.Cost)
		append("Size", d.Size)
		append("CR", d.CR)
		append("AC", d.AC)
		append("HP", d.HP)
		append("Source", d.Source)
	} else {
		if r.Subtype != "" {
			b.WriteString("\n" + escapeHTML(r.Subtype))
		}
	}
	return b.String()
}
