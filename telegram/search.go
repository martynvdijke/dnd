package telegram

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"villum/db"
	"villum/search"
)

const searchPageSize = 5
const searchStateTTL = 15 * time.Minute

type searchState struct {
	query      string
	typeFilter string
	results    []search.CompendiumResult
	page       int
	expires    time.Time
}

var searchStore = struct {
	sync.Mutex
	m map[string]*searchState
}{m: map[string]*searchState{}}

var searchTokenRegexp = regexp.MustCompile(`^[0-9a-f]+$`)

func newSearchToken() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func pruneSearchStates() {
	now := time.Now()
	for k, v := range searchStore.m {
		if now.After(v.expires) {
			delete(searchStore.m, k)
		}
	}
}

func getSearchState(token string) (*searchState, bool) {
	searchStore.Lock()
	defer searchStore.Unlock()
	pruneSearchStates()
	st, ok := searchStore.m[token]
	if !ok {
		return nil, false
	}
	if time.Now().After(st.expires) {
		delete(searchStore.m, token)
		return nil, false
	}
	return st, true
}

func runCompendiumSearch(c *cmdContext, typeFilter string) botReply {
	query := strings.TrimSpace(strings.Join(c.args, " "))
	if query == "" {
		name := c.name
		if name == "" {
			name = "search"
		}
		return botReply{Text: fmt.Sprintf("Usage: /%s <query>", escapeHTML(name))}
	}
	results, err := search.SearchCompendium(c.ctx, db.DB, search.CompendiumParams{
		Query:         query,
		TypeFilter:    typeFilter,
		Limit:         searchPageSize,
		Reranker:      search.DefaultRerank,
		FuzzyFallback: true,
	})
	if err != nil {
		return botReply{Text: "Search failed, try again."}
	}
	if len(results) == 0 {
		return botReply{Text: fmt.Sprintf("No compendium results for %s. Check the spelling or try a type command like /spell.", escapeHTML(query))}
	}
	token := newSearchToken()
	st := &searchState{query: query, typeFilter: typeFilter, results: results, page: 0, expires: time.Now().Add(searchStateTTL)}
	searchStore.Lock()
	pruneSearchStates()
	searchStore.m[token] = st
	searchStore.Unlock()
	return renderSearchPage(query, results, 0, token)
}

func renderSearchPage(query string, results []search.CompendiumResult, page int, token string) botReply {
	total := len(results)
	pages := (total + searchPageSize - 1) / searchPageSize
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start := page * searchPageSize
	end := start + searchPageSize
	if end > total {
		end = total
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>🔍 Compendium results</b> — %s (page %d, %d results)", escapeHTML(query), page+1, total))
	b.WriteString("\n")
	for i := start; i < end; i++ {
		r := results[i]
		sub := r.Subtype
		if sub == "" && r.Type != "" {
			sub = r.Type
		}
		label := fmt.Sprintf("%d. <b>%s</b> — %s", i-start+1, escapeHTML(r.Name), escapeHTML(sub))
		b.WriteString(label + "\n")
	}
	text := b.String()

	var rows [][]inlineButton
	// detail buttons
	for i := start; i < end; i++ {
		r := results[i]
		btnLabel := fmt.Sprintf("%d. %s", i-start+1, truncateRunes(r.Name, 30))
		rows = append(rows, []inlineButton{{Text: btnLabel, CallbackData: cbSearchDetailPrefix + token + ":" + strconv.Itoa(i)}})
	}
	// nav row
	var nav []inlineButton
	if page > 0 {
		nav = append(nav, inlineButton{Text: "◀ Prev", CallbackData: cbSearchPrefix + token + ":" + strconv.Itoa(page-1)})
	}
	if page+1 < pages {
		nav = append(nav, inlineButton{Text: "Next ▶", CallbackData: cbSearchPrefix + token + ":" + strconv.Itoa(page+1)})
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}
	kb := inlineKeyboard(rows...)
	return botReply{Text: text, Keyboard: kb}
}

func handleSearchPageCallback(c *cmdContext, data string) (botReply, bool) {
	rest := strings.TrimPrefix(data, cbSearchPrefix)
	sep := strings.LastIndex(rest, ":")
	if sep < 0 {
		return botReply{}, false
	}
	token := rest[:sep]
	pageStr := rest[sep+1:]
	if !searchTokenRegexp.MatchString(token) {
		return botReply{}, false
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return botReply{}, false
	}
	st, ok := getSearchState(token)
	if !ok {
		return botReply{Text: "This search expired, run it again."}, true
	}
	searchStore.Lock()
	st.page = page
	query := st.query
	results := append([]search.CompendiumResult(nil), st.results...)
	pageSnap := st.page
	searchStore.Unlock()
	return renderSearchPage(query, results, pageSnap, token), true
}

func handleSearchDetailCallback(c *cmdContext, data string) (botReply, bool) {
	rest := strings.TrimPrefix(data, cbSearchDetailPrefix)
	sep := strings.LastIndex(rest, ":")
	if sep < 0 {
		return botReply{}, false
	}
	token := rest[:sep]
	idxStr := rest[sep+1:]
	if !searchTokenRegexp.MatchString(token) {
		return botReply{}, false
	}
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return botReply{}, false
	}
	st, ok := getSearchState(token)
	if !ok {
		return botReply{Text: "This search expired, run it again."}, true
	}
	searchStore.Lock()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(st.results) {
		idx = len(st.results) - 1
	}
	res := st.results[idx]
	pageSnap := st.page
	searchStore.Unlock()
	detail, err := search.LoadCompendiumDetail(c.ctx, db.DB, res.Type, res.ID)
	if err != nil {
		return botReply{Text: "Entry not found."}, true
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>%s</b> — %s\n", escapeHTML(detail.Name), escapeHTML(detail.Type)))
	if detail.Subtype != "" {
		b.WriteString(escapeHTML(detail.Subtype) + "\n")
	}
	desc := detail.Description
	descTruncated := false
	if runeCount(desc) > 3000 {
		desc = truncateRunes(desc, 2999) + "…"
		descTruncated = true
	}
	if desc != "" {
		b.WriteString("\n" + escapeHTML(desc) + "\n")
		if descTruncated {
			b.WriteString("\n… (truncated)\n")
		}
	}
	appendField := func(label, val string) {
		if val != "" {
			b.WriteString(fmt.Sprintf("\n<b>%s:</b> %s", escapeHTML(label), escapeHTML(val)))
		}
	}
	appendField("School", detail.School)
	appendField("Level", func() string {
		if detail.Level != 0 {
			return strconv.Itoa(detail.Level)
		}
		return ""
	}())
	appendField("Casting Time", detail.CastingTime)
	appendField("Range", detail.Range)
	appendField("Components", detail.Components)
	appendField("Duration", detail.Duration)
	appendField("Classes", detail.Classes)
	appendField("Category", detail.Category)
	appendField("Cost", detail.Cost)
	if detail.Weight != 0 {
		b.WriteString(fmt.Sprintf("\n<b>Weight:</b> %s", escapeHTML(strconv.FormatFloat(detail.Weight, 'f', -1, 64))))
	}
	appendField("Size", detail.Size)
	appendField("CR", detail.CR)
	appendField("AC", detail.AC)
	appendField("HP", detail.HP)
	appendField("Source", detail.Source)

	text := b.String()
	if runeCount(text) > 4000 {
		text = truncateRunes(text, 3999) + "…"
	}
	var rows [][]inlineButton
	if res.Type == "equipment" {
		rows = append(rows, []inlineButton{{
			Text:         "➕ Add to inventory",
			CallbackData: cbCompendiumAddPrefix + token + ":" + strconv.Itoa(idx),
		}})
	}
	rows = append(rows, []inlineButton{{Text: "◀ Back", CallbackData: cbSearchPrefix + token + ":" + strconv.Itoa(pageSnap)}})
	kb := inlineKeyboard(rows...)
	return botReply{Text: text, Keyboard: kb}, true
}
