package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

type SearchResultItem struct {
	Type    string `json:"type"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Snippet string `json:"snippet,omitempty"`
	Subtype string `json:"subtype,omitempty"`
}

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

func searchLike(rows *sql.Rows, err error, source string) []SearchResultItem {
	items := make([]SearchResultItem, 0)
	if err != nil {
		middleware.LogWarn("search", "query failed", "source", source, "error", err)
		return items
	}
	if rows == nil {
		return items
	}
	defer rows.Close()
	for rows.Next() {
		var r SearchResultItem
		var snippet sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &snippet); err != nil {
			middleware.LogWarn("search", "scan failed, skipping row", "source", source, "error", err)
			continue
		}
		if snippet.Valid {
			r.Snippet = snippet.String
		}
		items = append(items, r)
	}
	if err := rows.Err(); err != nil {
		middleware.LogWarn("search", "row iteration error", "source", source, "error", err)
	}
	return items
}

// searchQuery runs a LIKE query and returns search result items, logging any
// errors with the given source label.
func searchQuery(source, query string, args ...any) []SearchResultItem {
	rows, err := db.DB.Query(query, args...)
	return searchLike(rows, err, source)
}

func HandleSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	like := "%" + q + "%"

	results := SearchResults{
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

	// Characters via LIKE (all tables for consistency)
	results.Characters = searchQuery("characters", `
		SELECT id, name, CASE WHEN backstory LIKE ? THEN substr(backstory, max(0, instr(backstory, ?)-20), 60) ELSE NULL END
		FROM characters WHERE name LIKE ? OR race LIKE ? OR class LIKE ? OR background LIKE ? LIMIT 10`,
		like, q, like, like, like, like)

	// NPCs
	results.NPCs = searchQuery("npcs", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM npcs WHERE name LIKE ? OR race LIKE ? OR class LIKE ? OR description LIKE ? LIMIT 10`,
		like, q, like, like, like, like)

	// Notes
	results.Notes = searchQuery("notes", `
		SELECT id, title, CASE WHEN content LIKE ? THEN substr(content, max(0, instr(content, ?)-20), 60) ELSE NULL END
		FROM character_notes WHERE title LIKE ? OR content LIKE ? LIMIT 10`,
		like, q, like, like)

	// Quests
	results.Quests = searchQuery("quests", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM quests WHERE name LIKE ? OR description LIKE ? OR objectives LIKE ? LIMIT 10`,
		like, q, like, like, like)

	// Journal
	results.Journal = searchQuery("journal", `
		SELECT id, title, CASE WHEN entry LIKE ? THEN substr(entry, max(0, instr(entry, ?)-20), 60) ELSE NULL END
		FROM journal WHERE title LIKE ? OR entry LIKE ? LIMIT 10`,
		like, q, like, like)

	// Sessions
	results.Sessions = searchQuery("sessions", `
		SELECT id, title, CASE WHEN notes LIKE ? THEN substr(notes, max(0, instr(notes, ?)-20), 60) ELSE NULL END
		FROM sessions WHERE title LIKE ? OR notes LIKE ? OR important_events LIKE ? LIMIT 10`,
		like, q, like, like, like)

	// Compendium spells
	results.Spells = searchQuery("spells", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM compendium_spells WHERE name LIKE ? OR description LIKE ? LIMIT 10`,
		like, q, like, like)

	// Compendium equipment
	results.Equipment = searchQuery("equipment", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM compendium_equipment WHERE name LIKE ? OR description LIKE ? LIMIT 10`,
		like, q, like, like)

	// Compendium races
	results.Races = searchQuery("races", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM compendium_races WHERE name LIKE ? OR description LIKE ? LIMIT 5`,
		like, q, like, like)

	// Compendium classes
	results.Classes = searchQuery("classes", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM compendium_classes WHERE name LIKE ? OR description LIKE ? LIMIT 5`,
		like, q, like, like)

	// Compendium feats
	results.Feats = searchQuery("feats", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM compendium_feats WHERE name LIKE ? OR description LIKE ? LIMIT 5`,
		like, q, like, like)

	// Compendium backgrounds
	results.Backgrounds = searchQuery("backgrounds", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM compendium_backgrounds WHERE name LIKE ? OR description LIKE ? LIMIT 5`,
		like, q, like, like)

	// Campaigns
	results.Campaigns = searchQuery("campaigns", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60) ELSE NULL END
		FROM campaigns WHERE name LIKE ? OR description LIKE ? OR party_name LIKE ? LIMIT 10`,
		like, q, like, like, like)

	// Compendium monsters
	results.Monsters = searchQuery("monsters", `
		SELECT id, name, CASE WHEN description LIKE ? THEN substr(description, max(0, instr(description, ?)-20), 60)
			WHEN type LIKE ? THEN type ELSE NULL END
		FROM compendium_monsters WHERE name LIKE ? OR type LIKE ? OR description LIKE ? LIMIT 10`,
		like, q, like, like, like, like)

	c.JSON(http.StatusOK, results)
}
