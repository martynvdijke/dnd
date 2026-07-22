package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/registry"
)

// LinkResponse is the JSON shape for a single link in API responses.
type LinkResponse struct {
	ID         int64  `json:"id"`
	SourceType string `json:"source_type"`
	SourceID   int64  `json:"source_id"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	Context    string `json:"context"`
	CreatedAt  string `json:"created_at"`

	// Resolved display info (populated by GET).
	SourceTitle string `json:"source_title,omitempty"`
	TargetTitle string `json:"target_title,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
	TargetURL   string `json:"target_url,omitempty"`
}

// MentionRef identifies one mentioned entity returned from a saved document.
type MentionRef struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
}

// RegisterLinkRoutes registers link CRUD endpoints.
func RegisterLinkRoutes(auth *gin.RouterGroup) {
	auth.POST("/links", HandleCreateLink)
	auth.DELETE("/links/:id", HandleDeleteLink)
	auth.GET("/links/:type/:id", HandleGetLinks)
}

// HandleCreateLink creates a new manual link between two entities.
func HandleCreateLink(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt, _ := userID.(int64)
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	var req struct {
		SourceType string `json:"source_type"`
		SourceID   int64  `json:"source_id"`
		TargetType string `json:"target_type"`
		TargetID   int64  `json:"target_id"`
		Context    string `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Context == "" {
		req.Context = "manual"
	}
	if req.Context != "manual" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "context must be 'manual' for direct API creation"})
		return
	}

	// Validate both types are linkable.
	if !registry.Linkable(req.SourceType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown or non-linkable source type: %s", req.SourceType)})
		return
	}
	if !registry.Linkable(req.TargetType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown or non-linkable target type: %s", req.TargetType)})
		return
	}

	// Check edit permission on source entity.
	if !registry.Editable(db.DB, req.SourceType, req.SourceID, userIDInt, isAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "no edit permission on source entity"})
		return
	}

	// Check for duplicate.
	var exists int
	db.DB.QueryRow(`SELECT COUNT(*) FROM entity_links WHERE source_type=? AND source_id=? AND target_type=? AND target_id=? AND context=?`,
		req.SourceType, req.SourceID, req.TargetType, req.TargetID, req.Context).Scan(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "link already exists"})
		return
	}

	result, err := db.DB.Exec(`INSERT INTO entity_links(source_type, source_id, target_type, target_id, context) VALUES(?,?,?,?,?)`,
		req.SourceType, req.SourceID, req.TargetType, req.TargetID, req.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, LinkResponse{
		ID:         id,
		SourceType: req.SourceType,
		SourceID:   req.SourceID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Context:    req.Context,
	})
}

// HandleDeleteLink removes a link by ID.
func HandleDeleteLink(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt, _ := userID.(int64)
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	// Fetch the link to check permission on source.
	var sourceType string
	var sourceID int64
	err = db.DB.QueryRow(`SELECT source_type, source_id FROM entity_links WHERE id=?`, id).Scan(&sourceType, &sourceID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !registry.Editable(db.DB, sourceType, sourceID, userIDInt, isAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "no edit permission on source entity"})
		return
	}

	if _, err := db.DB.Exec(`DELETE FROM entity_links WHERE id=?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// HandleGetLinks returns outgoing links and backlinks for a given entity, with
// resolved display info from the search index. Results are permission-scoped:
// a link is only included when the user can see both its source and target.
func HandleGetLinks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt, _ := userID.(int64)
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	entityType := c.Param("type")
	entityID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	// Fetch all links where this entity is source (outgoing) or target (backlinks).
	rows, err := db.DB.Query(`
		SELECT id, source_type, source_id, target_type, target_id, context, created_at
		FROM entity_links
		WHERE (source_type=? AND source_id=?) OR (target_type=? AND target_id=?)
		ORDER BY created_at DESC`, entityType, entityID, entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type linkRow struct {
		id         int64
		sourceType string
		sourceID   int64
		targetType string
		targetID   int64
		context    string
		createdAt  string
	}
	var allLinks []linkRow
	for rows.Next() {
		var l linkRow
		if err := rows.Scan(&l.id, &l.sourceType, &l.sourceID, &l.targetType, &l.targetID, &l.context, &l.createdAt); err != nil {
			continue
		}
		allLinks = append(allLinks, l)
	}

	// Collect unique source and target entity references for permission check
	// and display name resolution.
	type entityRef struct {
		entityType string
		entityID   int64
	}
	seen := map[entityRef]bool{}
	var refs []entityRef
	for _, l := range allLinks {
		for _, ref := range []entityRef{
			{l.sourceType, l.sourceID},
			{l.targetType, l.targetID},
		} {
			if !seen[ref] {
				seen[ref] = true
				refs = append(refs, ref)
			}
		}
	}

	// Permission-scope: for each unique type, get visible IDs.
	typeGroup := map[string][]int64{}
	for _, ref := range refs {
		typeGroup[ref.entityType] = append(typeGroup[ref.entityType], ref.entityID)
	}
	visible := map[string]map[int64]bool{}
	for et, ids := range typeGroup {
		if isAdmin {
			vm := map[int64]bool{}
			for _, id := range ids {
				vm[id] = true
			}
			visible[et] = vm
		} else {
			vm, err := registry.VisibleIDs(db.DB, et, ids, userIDInt, false)
			if err != nil {
				vm = map[int64]bool{}
			}
			visible[et] = vm
		}
	}

	// Resolve display title from entity_search_index for each ref.
	titleMap := map[entityRef]string{}
	for _, ref := range refs {
		var title string
		err := db.DB.QueryRow(`SELECT title FROM entity_search_index WHERE entity_type=? AND entity_id=?`, ref.entityType, ref.entityID).Scan(&title)
		if err != nil {
			title = ref.entityType + " #" + strconv.FormatInt(ref.entityID, 10)
		}
		titleMap[ref] = title
	}

	// Build outgoing and backlinks list, filtering by visibility.
	outgoing := make([]LinkResponse, 0)
	backlinks := make([]LinkResponse, 0)
	for _, l := range allLinks {
		srcRef := entityRef{l.sourceType, l.sourceID}
		tgtRef := entityRef{l.targetType, l.targetID}

		// Both sides must be visible to the user.
		if !visible[srcRef.entityType][srcRef.entityID] || !visible[tgtRef.entityType][tgtRef.entityID] {
			continue
		}

		resp := LinkResponse{
			ID:          l.id,
			SourceType:  l.sourceType,
			SourceID:    l.sourceID,
			TargetType:  l.targetType,
			TargetID:    l.targetID,
			Context:     l.context,
			CreatedAt:   l.createdAt,
			SourceTitle: titleMap[srcRef],
			TargetTitle: titleMap[tgtRef],
			SourceURL:   entityURL(l.sourceType, l.sourceID),
			TargetURL:   entityURL(l.targetType, l.targetID),
		}

		if l.sourceType == entityType && l.sourceID == entityID {
			outgoing = append(outgoing, resp)
		} else {
			backlinks = append(backlinks, resp)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"outgoing":  outgoing,
		"backlinks": backlinks,
	})
}

// ReconcileMentionLinks diffs the current set of mention links for a source
// entity against the provided mentions slice, adding and removing rows as
// needed. This is called when a rich-text document containing @-mentions is
// saved.
func ReconcileMentionLinks(sourceType string, sourceID int64, mentions []MentionRef) error {
	// Build a lookup of the desired mention set.
	want := make(map[string]int64) // entityType+":"+id → entityID
	for _, m := range mentions {
		key := m.EntityType + ":" + fmt.Sprint(m.EntityID)
		want[key] = m.EntityID
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("reconcile begin tx: %w", err)
	}
	defer tx.Rollback()

	// Fetch current mention links for this source.
	rows, err := tx.Query(`SELECT target_type, target_id FROM entity_links WHERE source_type=? AND source_id=? AND context='mention'`,
		sourceType, sourceID)
	if err != nil {
		return fmt.Errorf("reconcile query current: %w", err)
	}

	type key struct{ entityType, entityIDStr string }
	have := map[key]bool{}
	for rows.Next() {
		var tgtType string
		var tgtID int64
		if err := rows.Scan(&tgtType, &tgtID); err != nil {
			rows.Close()
			return fmt.Errorf("reconcile scan: %w", err)
		}
		have[key{tgtType, fmt.Sprint(tgtID)}] = true
	}
	rows.Close()

	// Delete mention links that are no longer wanted.
	delStmt, err := tx.Prepare(`DELETE FROM entity_links WHERE source_type=? AND source_id=? AND target_type=? AND target_id=? AND context='mention'`)
	if err != nil {
		return fmt.Errorf("reconcile prepare delete: %w", err)
	}
	defer delStmt.Close()

	for _, l := range mentions {
		k := key{l.EntityType, fmt.Sprint(l.EntityID)}
		if have[k] {
			delete(have, k) // mark as seen, keep
		}
		// Remaining in `have` after this loop are stale.
	}

	for k := range have {
		id, _ := strconv.ParseInt(k.entityIDStr, 10, 64)
		if _, err := delStmt.Exec(sourceType, sourceID, k.entityType, id); err != nil {
			return fmt.Errorf("reconcile delete stale: %w", err)
		}
	}

	// Insert new mention links.
	insStmt, err := tx.Prepare(`INSERT OR IGNORE INTO entity_links(source_type, source_id, target_type, target_id, context) VALUES(?,?,?,?,'mention')`)
	if err != nil {
		return fmt.Errorf("reconcile prepare insert: %w", err)
	}
	defer insStmt.Close()

	for _, m := range mentions {
		if _, err := insStmt.Exec(sourceType, sourceID, m.EntityType, m.EntityID); err != nil {
			return fmt.Errorf("reconcile insert: %w", err)
		}
	}

	return tx.Commit()
}

// entityURL is defined in search.go.
// entityTypeName is defined in search.go.
// Those functions use the forwarded-by-origin pattern.
