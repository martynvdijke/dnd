package handlers

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

var _ = regexp.MustCompile
var _ = time.Now
var _ = maps.Copy[map[string]any, map[string]any]

func ListCompendiumEntries(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Optional search
	q := strings.TrimSpace(c.Query("q"))
	var total int
	var entries []models.CompendiumEntry

	// Optional sorting: ?sort=<field>&order=asc|desc (field allowlist to prevent SQL injection)
	orderBy := "e.created_at DESC"
	if sortField := strings.TrimSpace(c.Query("sort")); sortField != "" {
		if fieldNameRe.MatchString(sortField) {
			dir := "ASC"
			if strings.ToLower(c.DefaultQuery("order", "asc")) == "desc" {
				dir = "DESC"
			}
			orderBy = "json_extract(e.data, '$.\"" + sortField + "\"') COLLATE NOCASE " + dir + ", e.id DESC"
		}
	}

	if q != "" {
		// FTS5 search scoped to schema
		err = db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_entries_fts f
			JOIN compendium_entries e ON f.rowid = e.id
			WHERE e.schema_id=? AND compendium_entries_fts MATCH ?`, schemaID, q).Scan(&total)
		if err != nil {
			total = 0
		}

		rows, err := db.DB.Query(`SELECT e.id, e.schema_id, e.data, e.created_at, e.updated_at
			FROM compendium_entries e
			JOIN compendium_entries_fts f ON e.id = f.rowid
			WHERE e.schema_id=? AND compendium_entries_fts MATCH ?
			ORDER BY `+orderBy+` LIMIT ? OFFSET ?`, schemaID, q, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		entries = scanEntries(rows)
	} else {
		db.DB.QueryRow("SELECT COUNT(*) FROM compendium_entries WHERE schema_id=?", schemaID).Scan(&total)

		rows, err := db.DB.Query(`SELECT e.id, e.schema_id, e.data, e.created_at, e.updated_at
			FROM compendium_entries e WHERE e.schema_id=? ORDER BY `+orderBy+` LIMIT ? OFFSET ?`,
			schemaID, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		entries = scanEntries(rows)
	}

	totalPages := max((total+pageSize-1)/pageSize, 1)

	c.JSON(http.StatusOK, models.CompendiumEntryList{
		Entries:    entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func scanEntries(rows interface{ Scan(...any) error }) []models.CompendiumEntry {
	out := make([]models.CompendiumEntry, 0)
	// rows is a *sql.Rows but we use duck typing
	type rowScanner interface {
		Scan(...any) error
		Next() bool
		Close() error
	}
	rs, ok := rows.(rowScanner)
	if !ok {
		return out
	}
	for rs.Next() {
		var e models.CompendiumEntry
		var dataJSON string
		var createdAt, updatedAt string
		if err := rs.Scan(&e.ID, &e.SchemaID, &dataJSON, &createdAt, &updatedAt); err != nil {
			continue
		}
		e.Data = make(map[string]any)
		json.Unmarshal([]byte(dataJSON), &e.Data)
		e.CreatedAt = createdAt
		e.UpdatedAt = updatedAt
		out = append(out, e)
	}
	return out
}

func GetCompendiumEntry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry id"})
		return
	}

	var e models.CompendiumEntry
	var dataJSON, createdAt, updatedAt string
	err = db.DB.QueryRow(`SELECT id, schema_id, data, created_at, updated_at FROM compendium_entries WHERE id=?`, id).
		Scan(&e.ID, &e.SchemaID, &dataJSON, &createdAt, &updatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	e.Data = make(map[string]any)
	json.Unmarshal([]byte(dataJSON), &e.Data)
	e.CreatedAt = createdAt
	e.UpdatedAt = updatedAt
	c.JSON(http.StatusOK, e)
}

// GetCompendiumEntryBySchema returns a single compendium entry to authenticated
// users (read-only), verifying the entry belongs to the requested schema.
func GetCompendiumEntryBySchema(c *gin.Context) {
	schemaID, err1 := strconv.ParseInt(c.Param("id"), 10, 64)
	entryID, err2 := strconv.ParseInt(c.Param("entryId"), 10, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var e models.CompendiumEntry
	var dataJSON string
	err := db.DB.QueryRow("SELECT id, schema_id, data, created_at, updated_at FROM compendium_entries WHERE id = ?", entryID).
		Scan(&e.ID, &e.SchemaID, &dataJSON, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	if e.SchemaID != schemaID {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found in schema"})
		return
	}
	if err := json.Unmarshal([]byte(dataJSON), &e.Data); err != nil {
		e.Data = map[string]any{}
	}
	c.JSON(http.StatusOK, e)
}

func CreateCompendiumEntry(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}

	// Verify schema exists
	var typeName string
	err = db.DB.QueryRow("SELECT type_name FROM compendium_schemas WHERE id=?", schemaID).Scan(&typeName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}

	var req struct {
		Data map[string]any `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize data"})
		return
	}

	result, err := db.DB.Exec(`INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`, schemaID, string(dataJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()

	// Also sync migrated legacy entries by adding a name hint for search
	AutoSyncCompendiumEntry(schemaID, id, req.Data)

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCompendiumEntry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry id"})
		return
	}

	var req struct {
		Data map[string]any `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize data"})
		return
	}

	result, err := db.DB.Exec(`UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=?`, string(dataJSON), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCompendiumEntry(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry id"})
		return
	}
	result, err := db.DB.Exec("DELETE FROM compendium_entries WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func BatchDeleteCompendiumEntries(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no ids provided"})
		return
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(req.IDs))
	args := make([]any, len(req.IDs))
	for i, id := range req.IDs {
		placeholders[i] = "?"
		args[i] = id
	}

	result, err := db.DB.Exec("DELETE FROM compendium_entries WHERE id IN ("+strings.Join(placeholders, ",")+")", args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	deleted, _ := result.RowsAffected()

	c.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func BatchUpdateCompendiumEntries(c *gin.Context) {
	var req struct {
		IDs  []int64        `json:"ids" binding:"required"`
		Data map[string]any `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no ids provided"})
		return
	}
	if len(req.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no data provided"})
		return
	}

	updated := int64(0)
	var errs []map[string]any

	for _, id := range req.IDs {
		// Fetch current data
		var dataJSON string
		err := db.DB.QueryRow("SELECT data FROM compendium_entries WHERE id=?", id).Scan(&dataJSON)
		if err != nil {
			errs = append(errs, map[string]any{"id": id, "error": "not found"})
			continue
		}

		// Merge update into existing data
		var entryData map[string]any
		json.Unmarshal([]byte(dataJSON), &entryData)
		maps.Copy(entryData, req.Data)

		newJSON, _ := json.Marshal(entryData)
		_, err = db.DB.Exec("UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=?", string(newJSON), id)
		if err != nil {
			errs = append(errs, map[string]any{"id": id, "error": err.Error()})
			continue
		}
		updated++
	}

	resp := gin.H{"updated": updated}
	if len(errs) > 0 {
		resp["errors"] = errs
	}
	c.JSON(http.StatusOK, resp)
}

// ─── Search (Cross-type FTS5) ───

func SearchCompendiumEntries(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	schemaFilter := c.Query("schema_id") // optional

	query := `SELECT e.id, e.schema_id, e.data, cs.type_name, cs.display_name
		FROM compendium_entries e
		JOIN compendium_entries_fts f ON e.id = f.rowid
		JOIN compendium_schemas cs ON e.schema_id = cs.id
		WHERE compendium_entries_fts MATCH ?`
	args := []any{q}

	if schemaFilter != "" {
		query += " AND e.schema_id=?"
		args = append(args, schemaFilter)
	}
	query += " ORDER BY rank LIMIT 50"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := make([]models.CompendiumSearchResult, 0)
	for rows.Next() {
		var r models.CompendiumSearchResult
		var dataJSON, typeName, displayName string
		if err := rows.Scan(&r.ID, &r.Type, &dataJSON, &typeName, &displayName); err != nil {
			continue
		}
		r.TypeName = displayName
		// Extract name from data JSON
		var data map[string]any
		if json.Unmarshal([]byte(dataJSON), &data) == nil {
			if name, ok := data["name"].(string); ok {
				r.Name = name
			}
			// Create snippet from first text field
			for _, v := range data {
				if s, ok := v.(string); ok && len(s) > len(r.Snippet) && s != r.Name {
					r.Snippet = truncateStr(s, 150)
					break
				}
			}
		}
		results = append(results, r)
	}
	c.JSON(http.StatusOK, results)
}

// ─── Import ───

func ExportCompendiumEntries(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}

	var schemaType, displayName string
	err = db.DB.QueryRow("SELECT type_name, display_name FROM compendium_schemas WHERE id=?", schemaID).
		Scan(&schemaType, &displayName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}

	format := c.DefaultQuery("format", "json")
	filter := c.Query("q")

	// POST mode: export a specific set of entry IDs (5.3)
	var ids []int64
	if c.Request.Method == http.MethodPost {
		var req struct {
			IDs []int64 `json:"ids"`
		}
		if err := c.ShouldBindJSON(&req); err == nil && len(req.IDs) > 0 {
			ids = req.IDs
		}
	}

	var rows interface {
		Scan(...any) error
		Close() error
		Next() bool
	}

	if len(ids) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
		args := make([]any, 0, len(ids)+1)
		args = append(args, schemaID)
		for _, id := range ids {
			args = append(args, id)
		}
		r, err := db.DB.Query(`SELECT id, data, created_at FROM compendium_entries
			WHERE schema_id=? AND id IN (`+placeholders+`) ORDER BY id`, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rows = r
	} else if filter != "" {
		r, err := db.DB.Query(`SELECT e.id, e.data, e.created_at
			FROM compendium_entries e
			JOIN compendium_entries_fts f ON e.id = f.rowid
			WHERE e.schema_id=? AND compendium_entries_fts MATCH ?
			ORDER BY e.created_at DESC`, schemaID, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rows = r
	} else {
		r, err := db.DB.Query(`SELECT id, data, created_at FROM compendium_entries WHERE schema_id=? ORDER BY created_at DESC`, schemaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rows = r
	}
	defer rows.Close()

	type exportEntry struct {
		ID        int64          `json:"id"`
		Data      map[string]any `json:"data"`
		CreatedAt string         `json:"created_at"`
	}

	entries := make([]exportEntry, 0)
	for rows.Next() {
		var e exportEntry
		var dataJSON, createdAt string
		if err := rows.Scan(&e.ID, &dataJSON, &createdAt); err != nil {
			continue
		}
		e.Data = make(map[string]any)
		json.Unmarshal([]byte(dataJSON), &e.Data)
		e.CreatedAt = createdAt
		entries = append(entries, e)
	}

	switch format {
	case "json":
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-export.json"`, schemaType))
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"schema":   schemaType,
			"name":     displayName,
			"exported": time.Now().UTC().Format(time.RFC3339),
			"count":    len(entries),
			"entries":  entries,
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"schema":  schemaType,
			"name":    displayName,
			"count":   len(entries),
			"entries": entries,
		})
	}
}

// ─── Import Logs ───

func ListCompendiumImportLogs(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT id, user_id, status, files, mapping, summary, created_at, rolled_back_at
		FROM compendium_import_logs ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	logs := make([]models.CompendiumImportLog, 0)
	for rows.Next() {
		var l models.CompendiumImportLog
		var filesJSON, mappingJSON, summaryJSON, createdAt string
		var rolledBackAt *string
		if err := rows.Scan(&l.ID, &l.UserID, &l.Status, &filesJSON, &mappingJSON, &summaryJSON, &createdAt, &rolledBackAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(filesJSON), &l.Files)
		json.Unmarshal([]byte(mappingJSON), &l.Mapping)
		json.Unmarshal([]byte(summaryJSON), &l.Summary)
		l.CreatedAt = createdAt
		l.RolledBackAt = rolledBackAt
		// Enrich with derived display fields (filename from files list, schema from summary)
		l.Filename = strings.Join(l.Files, ", ")
		if sid, ok := l.Summary["schema_id"]; ok {
			switch v := sid.(type) {
			case float64:
				l.SchemaID = int64(v)
			case int64:
				l.SchemaID = v
			}
		}
		if sn, ok := l.Summary["schema_name"].(string); ok {
			l.SchemaName = sn
		}
		logs = append(logs, l)
	}
	c.JSON(http.StatusOK, logs)
}

func RollbackCompendiumImport(c *gin.Context) {
	logID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}

	// For now, just mark as rolled_back.
	// Full rollback (removing inserted entries) requires storing entry IDs.
	_, err = db.DB.Exec(`UPDATE compendium_import_logs SET status='rolled_back', rolled_back_at=datetime('now') WHERE id=?`, logID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Seed Built-in Schemas ───

// SeedCompendiumSchemas creates the 7 default schema definitions from the existing
