package ai

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"villum/db"
	"villum/search"
)

type QuestionRequest struct {
	Query      string
	CampaignID int64
	UserID     int64
	IsAdmin    bool
	TypeFilter string
	MaxTokens  int
	Timeout    time.Duration
}

type QuestionResult struct {
	Answer   string
	Sources  []search.Source
	Scope    string
	Degraded bool
	Message  string
}

func AnswerQuestion(ctx context.Context, passed *sql.DB, req QuestionRequest) QuestionResult {
	d := db.DB
	if passed != nil {
		d = passed
	}
	q := strings.TrimSpace(req.Query)
	if q == "" {
		return QuestionResult{Degraded: true, Message: "A question is required.", Scope: "compendium", Sources: []search.Source{}}
	}

	if !IsEnabled(ctx, d) {
		res, _ := search.SearchCompendium(ctx, d, search.CompendiumParams{Query: q, TypeFilter: req.TypeFilter, Limit: 8, Reranker: search.DefaultRerank, FuzzyFallback: true})
		srcs := compendiumToSources(q, res)
		if srcs == nil {
			srcs = []search.Source{}
		}
		return QuestionResult{Sources: srcs, Scope: "compendium", Degraded: true, Message: "AI is not configured; showing direct matches instead."}
	}

	var sources []search.Source
	var blocks string
	scope := "compendium"

	if req.CampaignID > 0 {
		srcs, blks, err := search.RetrieveCampaignContext(ctx, d, req.CampaignID, req.UserID, req.IsAdmin, q, 8)
		if err == nil && len(srcs) > 0 {
			sources = srcs
			blocks = blks
			scope = "campaign"
		}
	}
	var system string
	if scope == "campaign" {
		system = SystemPromptForCampaignQA(blocks)
	} else {
		res, _ := search.SearchCompendium(ctx, d, search.CompendiumParams{Query: q, TypeFilter: req.TypeFilter, Limit: 8, Reranker: search.DefaultRerank, FuzzyFallback: true})
		compSources, compBlocks := compendiumToSourcesAndBlocks(ctx, d, res)
		sources = compSources
		blocks = compBlocks
		scope = "compendium"
		system = SystemPromptForCompendiumQA(blocks)
	}

	if len(sources) == 0 {
		if sources == nil {
			sources = []search.Source{}
		}
		return QuestionResult{Degraded: true, Scope: scope, Message: "AI could not find relevant context for that question.", Sources: sources}
	}

	ep, err := ResolveEndpoint(ctx, d, 0)
	if err != nil {
		if sources == nil {
			sources = []search.Source{}
		}
		return QuestionResult{Sources: sources, Scope: scope, Degraded: true, Message: "AI is unavailable; showing direct matches instead."}
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	genCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	mt := req.MaxTokens
	if mt <= 0 {
		mt = 800
	}
	answer, _, err := GenerateText(genCtx, d, GenerateRequest{EndpointID: ep.ID, Prompt: q, System: system, MaxTokens: &mt})
	if err != nil {
		if sources == nil {
			sources = []search.Source{}
		}
		return QuestionResult{Sources: sources, Scope: scope, Degraded: true, Message: "AI generation failed; showing direct matches instead."}
	}
	return QuestionResult{Answer: answer, Sources: sources, Scope: scope}
}

func compendiumToSources(query string, res []search.CompendiumResult) []search.Source {
	out := make([]search.Source, 0, len(res))
	for _, r := range res {
		url := search.EntityURL(r.Type, r.ID)
		if url == "" {
			url = search.EntityURL("compendium", r.ID)
		}
		out = append(out, search.Source{
			EntityType: r.Type,
			EntityID:   r.ID,
			Title:      r.Name,
			Subtitle:   r.Subtype,
			URL:        url,
		})
	}
	return out
}

func compendiumToSourcesAndBlocks(ctx context.Context, d *sql.DB, res []search.CompendiumResult) ([]search.Source, string) {
	sources := compendiumToSources("", res)
	var blocks []string
	for i, r := range res {
		detail, err := search.LoadCompendiumDetail(ctx, d, r.Type, r.ID)
		desc := ""
		name := r.Name
		subtype := r.Subtype
		if err == nil && detail != nil {
			if detail.Name != "" {
				name = detail.Name
			}
			if detail.Subtype != "" {
				subtype = detail.Subtype
			} else if detail.Category != "" && subtype == "" {
				subtype = detail.Category
			}
			desc = detail.Description
		}
		// truncate description to ~1200 runes
		if len([]rune(desc)) > 1200 {
			desc = string([]rune(desc)[:1200])
		}
		subLine := subtype
		if subLine == "" {
			subLine = r.Type
		}
		// level info if available
		if detail != nil && detail.Level != 0 && r.Type == "spell" {
			subLine = fmt.Sprintf("%s level %d", subLine, detail.Level)
		}
		block := fmt.Sprintf("[%d] %s (%s)\n%s\n%s", i+1, name, r.Type, subLine, desc)
		blocks = append(blocks, block)
	}
	return sources, strings.Join(blocks, "\n\n")
}
