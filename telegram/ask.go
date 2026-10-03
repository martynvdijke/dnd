package telegram

import (
	"fmt"
	"strings"

	"villum/db"
	"villum/search"
)

// TODO(compendium-and-ai-search 3.1): add AI generation on top of this retrieval fallback. Ceiling: retrieval-only answers. Upgrade path: ai.GenerateText with campaign-first then compendium context.
func runAskQuery(c *cmdContext, query string) botReply {
	q := strings.TrimSpace(query)
	if q == "" {
		return botReply{Text: "Usage: /ask <question>"}
	}
	results, err := search.SearchCompendium(c.ctx, db.DB, search.CompendiumParams{
		Query:         q,
		Limit:         5,
		Reranker:      search.DefaultRerank,
		FuzzyFallback: true,
	})
	if err != nil {
		return botReply{Text: "Search failed, try again."}
	}
	if len(results) == 0 {
		return botReply{Text: fmt.Sprintf("No compendium matches for %s.", escapeHTML(q))}
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("🔎 Top matches for \"%s\"\n", escapeHTML(q)))
	for i, r := range results {
		sub := r.Subtype
		if sub == "" {
			sub = r.Type
		}
		b.WriteString(fmt.Sprintf("\n%d. <b>%s</b> — %s", i+1, escapeHTML(r.Name), escapeHTML(sub)))
	}
	return botReply{Text: b.String()}
}
