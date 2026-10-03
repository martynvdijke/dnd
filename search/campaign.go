package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"villum/registry"
)

// RetrieveCampaignContext retrieves campaign-scoped search context.
func RetrieveCampaignContext(ctx context.Context, db *sql.DB, campaignID, userID int64, isAdmin bool, query string, limit int) ([]Source, string, error) {
	if limit <= 0 {
		limit = 8
	}
	campaignSet := map[string]map[int64]bool{}
	add := func(et string, ids []int64) {
		if len(ids) == 0 {
			return
		}
		m := map[int64]bool{}
		for _, id := range ids {
			m[id] = true
		}
		campaignSet[et] = m
	}
	queryIDs := func(q string, args ...any) []int64 {
		rows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil
		}
		defer rows.Close()
		var out []int64
		for rows.Next() {
			var id int64
			rows.Scan(&id)
			out = append(out, id)
		}
		return out
	}
	add("character", queryIDs("SELECT c.id FROM characters c JOIN campaign_characters cm ON cm.character_id=c.id WHERE cm.campaign_id=?", campaignID))
	add("campaign", queryIDs("SELECT id FROM campaigns WHERE id=?", campaignID))
	add("wiki", queryIDs("SELECT id FROM campaign_wiki_pages WHERE campaign_id=?", campaignID))
	add("timeline", queryIDs("SELECT id FROM campaign_timeline_events WHERE campaign_id=?", campaignID))
	add("knowledge", queryIDs("SELECT id FROM campaign_knowledge WHERE campaign_id=?", campaignID))
	add("faction", queryIDs("SELECT id FROM factions WHERE campaign_id=?", campaignID))
	add("shop", queryIDs("SELECT id FROM shops WHERE campaign_id=?", campaignID))
	add("adventure", queryIDs("SELECT id FROM oneshot_adventures WHERE campaign_id=?", campaignID))
	add("encounter", queryIDs("SELECT id FROM encounter_templates WHERE campaign_id=?", campaignID))
	add("npc", queryIDs("SELECT npc_id FROM campaign_npcs WHERE campaign_id=?", campaignID))
	add("session", queryIDs("SELECT s.id FROM sessions s JOIN characters c ON s.character_id=c.id JOIN campaign_characters cm ON cm.character_id=c.id WHERE cm.campaign_id=?", campaignID))
	add("quest", queryIDs("SELECT q.id FROM quests q JOIN characters c ON q.character_id=c.id JOIN campaign_characters cm ON cm.character_id=c.id WHERE cm.campaign_id=?", campaignID))
	add("journal", queryIDs("SELECT j.id FROM journal j JOIN characters c ON j.character_id=c.id JOIN campaign_characters cm ON cm.character_id=c.id WHERE cm.campaign_id=?", campaignID))
	add("note", queryIDs("SELECT n.id FROM character_notes n JOIN characters c ON n.character_id=c.id JOIN campaign_characters cm ON cm.character_id=c.id WHERE cm.campaign_id=?", campaignID))
	add("recap", queryIDs("SELECT id FROM campaign_recaps WHERE campaign_id=?", campaignID))

	ftsQ := BuildFTS5Query(query)
	if ftsQ == "" {
		return []Source{}, "", nil
	}
	sqlStr := `SELECT entity_type, entity_id, title, subtitle, snippet(entity_search_index,2,'<b>','</b>','...',32), bm25(entity_search_index,10.0,5.0,1.0) AS score FROM entity_search_index WHERE entity_search_index MATCH ? ORDER BY score DESC LIMIT ?`
	rows, err := db.QueryContext(ctx, sqlStr, ftsQ, 100)
	if err != nil {
		return []Source{}, "", nil
	}
	defer rows.Close()
	type raw struct {
		et, title, subtitle, snippet string
		id                           int64
		score                        float64
	}
	var raws []raw
	for rows.Next() {
		var r raw
		var snip *string
		if err := rows.Scan(&r.et, &r.id, &r.title, &r.subtitle, &snip, &r.score); err != nil {
			continue
		}
		if snip != nil {
			r.snippet = *snip
		}
		if m, ok := campaignSet[r.et]; !ok || !m[r.id] {
			continue
		}
		raws = append(raws, r)
	}
	typeGroup := map[string][]int64{}
	for _, r := range raws {
		typeGroup[r.et] = append(typeGroup[r.et], r.id)
	}
	visible := map[string]map[int64]bool{}
	for et, ids := range typeGroup {
		if isAdmin {
			m := map[int64]bool{}
			for _, id := range ids {
				m[id] = true
			}
			visible[et] = m
		} else {
			vis, err := registry.VisibleIDs(db, et, ids, userID, false)
			if err != nil {
				visible[et] = map[int64]bool{}
			} else {
				visible[et] = vis
			}
		}
	}
	var filtered []raw
	for _, r := range raws {
		if vis, ok := visible[r.et]; ok && vis[r.id] {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	sources := make([]Source, 0, len(filtered))
	var blocks []string
	for i, r := range filtered {
		sources = append(sources, Source{
			EntityType: r.et,
			EntityID:   r.id,
			Title:      r.title,
			Subtitle:   r.subtitle,
			URL:        EntityURL(r.et, r.id),
			Snippet:    r.snippet,
		})
		block := fmt.Sprintf("[%d] %s/%d %s\n%s\n%s", i+1, r.et, r.id, r.title, r.subtitle, r.snippet)
		blocks = append(blocks, block)
	}
	ctxStr := strings.Join(blocks, "\n\n")
	if len(ctxStr) > 8000 {
		ctxStr = ctxStr[:8000] + "\n(context truncated)"
	}
	if sources == nil {
		sources = []Source{}
	}
	return sources, ctxStr, nil
}
