package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"villum/registry"
)

func SearchUnified(ctx context.Context, db *sql.DB, p UnifiedParams) ([]UnifiedResult, error) {
	ftsQuery := BuildFTS5Query(p.Query)
	if ftsQuery == "" {
		return nil, fmt.Errorf("invalid query")
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	sqlStr := `SELECT entity_type, entity_id, title, subtitle,
	           snippet(entity_search_index, 2, '<b>', '</b>', '...', 32),
	           bm25(entity_search_index, 10.0, 5.0, 1.0) AS score
	        FROM entity_search_index
	        WHERE entity_search_index MATCH ?`
	args := []any{ftsQuery}
	if strings.TrimSpace(p.TypesFilter) != "" {
		parts := strings.Split(p.TypesFilter, ",")
		phs := make([]string, len(parts))
		for i, p2 := range parts {
			phs[i] = "?"
			args = append(args, strings.TrimSpace(p2))
		}
		sqlStr += " AND entity_type IN (" + strings.Join(phs, ",") + ")"
	}
	sqlStr += " ORDER BY score DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return []UnifiedResult{}, nil
	}
	defer rows.Close()

	type raw struct {
		entityType string
		entityID   int64
		title      string
		subtitle   string
		snippet    string
		score      float64
	}
	var raws []raw
	typeGroup := map[string][]int64{}
	for rows.Next() {
		var r raw
		var snip *string
		if err := rows.Scan(&r.entityType, &r.entityID, &r.title, &r.subtitle, &snip, &r.score); err != nil {
			continue
		}
		if snip != nil {
			r.snippet = *snip
		}
		raws = append(raws, r)
		typeGroup[r.entityType] = append(typeGroup[r.entityType], r.entityID)
	}
	visible := map[string]map[int64]bool{}
	for et, ids := range typeGroup {
		if p.IsAdmin {
			m := map[int64]bool{}
			for _, id := range ids {
				m[id] = true
			}
			visible[et] = m
		} else {
			vis, err := registry.VisibleIDs(db, et, ids, p.UserID, false)
			m := map[int64]bool{}
			if err == nil {
				for id := range vis {
					m[id] = true
				}
			}
			visible[et] = m
		}
	}
	results := make([]UnifiedResult, 0, len(raws))
	for _, r := range raws {
		if vis, ok := visible[r.entityType]; !ok || !vis[r.entityID] {
			continue
		}
		results = append(results, UnifiedResult{
			EntityType: r.entityType,
			EntityID:   r.entityID,
			Title:      r.title,
			Subtitle:   r.subtitle,
			Snippet:    r.snippet,
			Score:      r.score,
			URL:        EntityURL(r.entityType, r.entityID),
		})
	}
	return results, nil
}

func entityURL(et string, id int64) string {
	switch et {
	case "character":
		return fmt.Sprintf("#/characters/%d", id)
	case "npc":
		return fmt.Sprintf("#/npcs/%d", id)
	case "note":
		return fmt.Sprintf("#/notes/%d", id)
	case "quest":
		return fmt.Sprintf("#/quests/%d", id)
	case "journal":
		return fmt.Sprintf("#/journal/%d", id)
	case "session":
		return fmt.Sprintf("#/sessions/%d", id)
	case "campaign":
		return fmt.Sprintf("#/campaigns/%d", id)
	case "location":
		return fmt.Sprintf("#/locations/%d", id)
	case "encounter":
		return fmt.Sprintf("#/encounters/%d", id)
	case "monster":
		return fmt.Sprintf("#/monsters/%d", id)
	case "shop":
		return fmt.Sprintf("#/shops/%d", id)
	case "faction":
		return fmt.Sprintf("#/factions/%d", id)
	case "adventure":
		return fmt.Sprintf("#/adventures/%d", id)
	case "wiki":
		return fmt.Sprintf("#/wiki/%d", id)
	case "timeline":
		return fmt.Sprintf("#/timeline/%d", id)
	case "item":
		return fmt.Sprintf("#/items/%d", id)
	case "recap":
		return "#/recaps"
	case "knowledge":
		return fmt.Sprintf("#/campaigns/knowledge/%d", id)
	case "compendium":
		return fmt.Sprintf("#/compendium/%d", id)
	}
	return ""
}
