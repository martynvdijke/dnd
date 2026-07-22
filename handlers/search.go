package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/registry"
)

// SearchResultItem is a legacy flat search result (kept for backward compat).
type SearchResultItem struct {
	Type    string `json:"type"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Snippet string `json:"snippet,omitempty"`
	Subtype string `json:"subtype,omitempty"`
}

// SearchResults is the legacy per-type response shape (kept for backward compat).
type SearchResults struct {
	Characters  []SearchResultItem `json:"characters"`
	NPCs        []SearchResultItem `json:"npcs"`
	Notes       []SearchResultItem `json:"notes"`
	Quests      []SearchResultItem `json:"quests"`
	Journal     []SearchResultItem `json:"journal"`
	Sessions    []SearchResultItem `json:"sessions"`
	Spells      []SearchResultItem `json:"spells"`
	Equipment   []SearchResultItem `json:"equipment"`
	Races       []SearchResultItem `json:"races"`
	Classes     []SearchResultItem `json:"classes"`
	Feats       []SearchResultItem `json:"feats"`
	Backgrounds []SearchResultItem `json:"backgrounds"`
	Campaigns   []SearchResultItem `json:"campaigns"`
	Monsters    []SearchResultItem `json:"monsters"`
}

// UnifiedResult is the new unified search result format.
type UnifiedResult struct {
	EntityType string  `json:"entity_type"`
	EntityID   int64   `json:"entity_id"`
	Title      string  `json:"title"`
	Subtitle   string  `json:"subtitle"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
	URL        string  `json:"url"`
}

// entityTypeLegacyBucket maps unified entity_type to legacy field name.
var entityTypeLegacyBucket = map[string]string{
	"character": "Characters",
	"npc":       "NPCs",
	"note":      "Notes",
	"quest":     "Quests",
	"journal":   "Journal",
	"session":   "Sessions",
	"campaign":  "Campaigns",
	"monster":   "Monsters",
}

// entityTypeName returns a human-readable display name for an entity type.
func entityTypeName(et string) string {
	switch et {
	case "character":
		return "Character"
	case "npc":
		return "NPC"
	case "note":
		return "Note"
	case "quest":
		return "Quest"
	case "journal":
		return "Journal"
	case "session":
		return "Session"
	case "campaign":
		return "Campaign"
	case "location":
		return "Location"
	case "encounter":
		return "Encounter"
	case "monster":
		return "Monster"
	case "shop":
		return "Shop"
	case "faction":
		return "Faction"
	case "adventure":
		return "Adventure"
	case "wiki":
		return "Wiki Page"
	case "timeline":
		return "Timeline Event"
	case "item":
		return "Item"
	case "compendium":
		return "Compendium"
	}
	return et
}

// entityURL builds a hash-fragment deep link for a given entity.
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
	case "compendium":
		return fmt.Sprintf("#/compendium/%d", id)
	}
	return ""
}

// buildFTS5Query converts a plain-text query into a safe FTS5 MATCH string.
// It escapes double-quotes, splits on whitespace, and appends * to each term
// for prefix matching.
func buildFTS5Query(q string) string {
	terms := strings.Fields(q)
	if len(terms) == 0 {
		return ""
	}
	escaped := make([]string, len(terms))
	for i, t := range terms {
		s := strings.ReplaceAll(t, `"`, `""`)
		// Wrap multi-character terms for prefix matching
		if len(s) > 0 {
			escaped[i] = `"` + s + `"*`
		}
	}
	return strings.Join(escaped, " AND ")
}

func init() {
	// Ensure empty slices not null in JSON for legacy fields
}

func emptyLegacy() SearchResults {
	return SearchResults{
		Characters:  []SearchResultItem{},
		NPCs:        []SearchResultItem{},
		Notes:       []SearchResultItem{},
		Quests:      []SearchResultItem{},
		Journal:     []SearchResultItem{},
		Sessions:    []SearchResultItem{},
		Spells:      []SearchResultItem{},
		Equipment:   []SearchResultItem{},
		Races:       []SearchResultItem{},
		Classes:     []SearchResultItem{},
		Feats:       []SearchResultItem{},
		Backgrounds: []SearchResultItem{},
		Campaigns:   []SearchResultItem{},
		Monsters:    []SearchResultItem{},
	}
}

// legacyBucketPtr returns a pointer to the correct legacy slice for a given
// entity type, or nil if the type has no legacy bucket.
func legacyBucketPtr(legacy *SearchResults, et string) *[]SearchResultItem {
	switch et {
	case "character":
		return &legacy.Characters
	case "npc":
		return &legacy.NPCs
	case "note":
		return &legacy.Notes
	case "quest":
		return &legacy.Quests
	case "journal":
		return &legacy.Journal
	case "session":
		return &legacy.Sessions
	case "campaign":
		return &legacy.Campaigns
	case "monster":
		return &legacy.Monsters
	}
	return nil
}

// HandleSearch performs FTS5 search across all entity types via the unified
// entity_search_index. It supports:
//   - q (required): the search query
//   - types (optional): comma-separated entity type filter
//   - limit (optional): max results, default 20, max 100
//
// Response includes both the new unified "results" array and the legacy per-type
// fields for backward compatibility.
func HandleSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt, _ := userID.(int64)
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := fmt.Sscanf(l, "%d", &limit); err != nil || n != 1 || limit < 1 || limit > 100 {
			limit = 20
		}
	}

	typesFilter := strings.TrimSpace(c.Query("types"))

	ftsQuery := buildFTS5Query(q)
	if ftsQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}

	// Build the SQL — query FTS5 with rank, snippet, and type filter.
	sql := `SELECT entity_type, entity_id, title, subtitle,
	           snippet(entity_search_index, 2, '<b>', '</b>', '...', 32),
	           bm25(entity_search_index, 10.0, 5.0, 1.0) AS score
	        FROM entity_search_index
	        WHERE entity_search_index MATCH ?`
	args := []any{ftsQuery}

	if typesFilter != "" {
		parts := strings.Split(typesFilter, ",")
		phs := make([]string, len(parts))
		for i, p := range parts {
			phs[i] = "?"
			args = append(args, strings.TrimSpace(p))
		}
		sql += " AND entity_type IN (" + strings.Join(phs, ",") + ")"
	}

	sql += " ORDER BY score DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.DB.Query(sql, args...)
	if err != nil {
		// FTS5 parse errors or empty index
		c.JSON(http.StatusOK, gin.H{
			"results":    []UnifiedResult{},
			"characters": []SearchResultItem{},
			"npcs":       []SearchResultItem{},
			"notes":      []SearchResultItem{},
			"quests":     []SearchResultItem{},
			"journal":    []SearchResultItem{},
			"sessions":   []SearchResultItem{},
			"spells":     []SearchResultItem{},
			"equipment":  []SearchResultItem{},
			"races":      []SearchResultItem{},
			"classes":    []SearchResultItem{},
			"feats":      []SearchResultItem{},
			"backgrounds": []SearchResultItem{},
			"campaigns":  []SearchResultItem{},
			"monsters":   []SearchResultItem{},
		})
		return
	}
	defer rows.Close()

	// Collect raw results and group IDs by type for permission check.
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

	// Permission check — get visible entity IDs per type.
	visible := map[string]map[int64]bool{}
	for et, ids := range typeGroup {
		if isAdmin {
			vis := map[int64]bool{}
			for _, id := range ids {
				vis[id] = true
			}
			visible[et] = vis
		} else {
			visIDs, err := registry.VisibleIDs(db.DB, et, ids, userIDInt, false)
			vis := map[int64]bool{}
			if err == nil {
				for id := range visIDs {
					vis[id] = true
				}
			}
			visible[et] = vis
		}
	}

	// Build response.
	results := make([]UnifiedResult, 0, len(raws))
	legacy := emptyLegacy()

	for _, r := range raws {
		vis := visible[r.entityType]
		if !vis[r.entityID] {
			continue
		}
		u := UnifiedResult{
			EntityType: r.entityType,
			EntityID:   r.entityID,
			Title:      r.title,
			Subtitle:   r.subtitle,
			Snippet:    r.snippet,
			Score:      r.score,
			URL:        entityURL(r.entityType, r.entityID),
		}
		results = append(results, u)

		// Legacy bucket
		if bucket := legacyBucketPtr(&legacy, r.entityType); bucket != nil {
			*bucket = append(*bucket, SearchResultItem{
				Type:    r.entityType,
				ID:      r.entityID,
				Name:    r.title,
				Snippet: r.snippet,
				Subtype: r.subtitle,
			})
		}
	}

	// Merge legacy + new into a single JSON response.
	resp := gin.H{
		"results": results,
	}
	// Flatten legacy fields into response.
	resp["characters"] = legacy.Characters
	resp["npcs"] = legacy.NPCs
	resp["notes"] = legacy.Notes
	resp["quests"] = legacy.Quests
	resp["journal"] = legacy.Journal
	resp["sessions"] = legacy.Sessions
	resp["spells"] = legacy.Spells
	resp["equipment"] = legacy.Equipment
	resp["races"] = legacy.Races
	resp["classes"] = legacy.Classes
	resp["feats"] = legacy.Feats
	resp["backgrounds"] = legacy.Backgrounds
	resp["campaigns"] = legacy.Campaigns
	resp["monsters"] = legacy.Monsters

	c.JSON(http.StatusOK, resp)
}
