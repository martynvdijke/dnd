package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/search"
)

// Re-export search types for backward compat (handlers tests import handlers.UnifiedResult)
type SearchResultItem = search.SearchResultItem
type SearchResults = search.SearchResults
type UnifiedResult = search.UnifiedResult

var entityTypeLegacyBucket = search.EntityTypeLegacyBucket

func entityTypeName(et string) string      { return search.EntityTypeName(et) }
func entityURL(et string, id int64) string { return search.EntityURL(et, id) }
func buildFTS5Query(q string) string       { return search.BuildFTS5Query(q) }
func emptyLegacy() SearchResults           { return search.EmptyLegacy() }
func legacyBucketPtr(legacy *SearchResults, et string) *[]SearchResultItem {
	return search.LegacyBucketPtr(legacy, et)
}

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
	if ftsQuery := buildFTS5Query(q); ftsQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
		return
	}
	results, err := search.SearchUnified(c.Request.Context(), db.DB, search.UnifiedParams{
		Query: q, TypesFilter: typesFilter, Limit: limit, UserID: userIDInt, IsAdmin: isAdmin,
	})
	if err != nil {
		// e.g. invalid query already handled; fallback to empty like before
		results = []UnifiedResult{}
	}
	// Build legacy buckets from unified results (preserve exact JSON shape)
	legacy := emptyLegacy()
	for _, r := range results {
		if bucket := legacyBucketPtr(&legacy, r.EntityType); bucket != nil {
			*bucket = append(*bucket, SearchResultItem{Type: r.EntityType, ID: r.EntityID, Name: r.Title, Snippet: r.Snippet, Subtype: r.Subtitle})
		}
	}
	resp := gin.H{"results": results}
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
