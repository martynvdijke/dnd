package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Schema CRUD ───

func ListCompendiumSchemas(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT cs.id, cs.type_name, cs.display_name, cs.fields, cs.created_at, cs.updated_at,
		(SELECT COUNT(*) FROM compendium_entries ce WHERE ce.schema_id = cs.id) AS entry_count
		FROM compendium_schemas cs ORDER BY cs.display_name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	schemas := make([]models.CompendiumSchema, 0)
	for rows.Next() {
		var s models.CompendiumSchema
		var fieldsJSON string
		var createdAt, updatedAt string
		if err := rows.Scan(&s.ID, &s.TypeName, &s.DisplayName, &fieldsJSON, &createdAt, &updatedAt, &s.EntryCount); err != nil {
			continue
		}
		s.CreatedAt = createdAt
		s.UpdatedAt = updatedAt
		json.Unmarshal([]byte(fieldsJSON), &s.Fields)
		schemas = append(schemas, s)
	}
	c.JSON(http.StatusOK, schemas)
}

func GetCompendiumSchema(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema id"})
		return
	}

	var s models.CompendiumSchema
	var fieldsJSON string
	var createdAt, updatedAt string
	err = db.DB.QueryRow(`SELECT cs.id, cs.type_name, cs.display_name, cs.fields, cs.created_at, cs.updated_at,
		(SELECT COUNT(*) FROM compendium_entries ce WHERE ce.schema_id = cs.id) AS entry_count
		FROM compendium_schemas cs WHERE cs.id=?`, id).
		Scan(&s.ID, &s.TypeName, &s.DisplayName, &fieldsJSON, &createdAt, &updatedAt, &s.EntryCount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	s.CreatedAt = createdAt
	s.UpdatedAt = updatedAt
	json.Unmarshal([]byte(fieldsJSON), &s.Fields)
	c.JSON(http.StatusOK, s)
}

func CreateCompendiumSchema(c *gin.Context) {
	var req struct {
		TypeName    string              `json:"type_name" binding:"required"`
		DisplayName string              `json:"display_name" binding:"required"`
		Fields      []models.SchemaField `json:"fields" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize fields"})
		return
	}

	result, err := db.DB.Exec(`INSERT INTO compendium_schemas(type_name, display_name, fields) VALUES(?,?,?)`,
		strings.ToLower(strings.TrimSpace(req.TypeName)), strings.TrimSpace(req.DisplayName), string(fieldsJSON))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "schema type_name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateCompendiumSchema(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema id"})
		return
	}

	var req struct {
		DisplayName string              `json:"display_name"`
		Fields      []models.SchemaField `json:"fields"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build UPDATE dynamically
	setClauses := []string{}
	args := []interface{}{}

	if req.DisplayName != "" {
		setClauses = append(setClauses, "display_name=?")
		args = append(args, strings.TrimSpace(req.DisplayName))
	}
	if req.Fields != nil {
		fieldsJSON, err := json.Marshal(req.Fields)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize fields"})
			return
		}
		setClauses = append(setClauses, "fields=?")
		args = append(args, string(fieldsJSON))
	}
	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
		return
	}

	setClauses = append(setClauses, "updated_at=datetime('now')")
	args = append(args, id)

	_, err = db.DB.Exec("UPDATE compendium_schemas SET "+strings.Join(setClauses, ", ")+" WHERE id=?", args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCompendiumSchema(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema id"})
		return
	}
	// CASCADE will delete entries
	_, err = db.DB.Exec("DELETE FROM compendium_schemas WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ─── Entry CRUD ───

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
			ORDER BY e.created_at DESC LIMIT ? OFFSET ?`, schemaID, q, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		entries = scanEntries(rows)
	} else {
		db.DB.QueryRow("SELECT COUNT(*) FROM compendium_entries WHERE schema_id=?", schemaID).Scan(&total)

		rows, err := db.DB.Query(`SELECT id, schema_id, data, created_at, updated_at
			FROM compendium_entries WHERE schema_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			schemaID, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		entries = scanEntries(rows)
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, models.CompendiumEntryList{
		Entries:    entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func scanEntries(rows interface{ Scan(...interface{}) error }) []models.CompendiumEntry {
	out := make([]models.CompendiumEntry, 0)
	// rows is a *sql.Rows but we use duck typing
	type rowScanner interface {
		Scan(...interface{}) error
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
		e.Data = make(map[string]interface{})
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
	e.Data = make(map[string]interface{})
	json.Unmarshal([]byte(dataJSON), &e.Data)
	e.CreatedAt = createdAt
	e.UpdatedAt = updatedAt
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
		Data map[string]interface{} `json:"data" binding:"required"`
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
		Data map[string]interface{} `json:"data" binding:"required"`
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
	args := make([]interface{}, len(req.IDs))
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
		IDs  []int64                `json:"ids" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
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
	var errs []map[string]interface{}

	for _, id := range req.IDs {
		// Fetch current data
		var dataJSON string
		err := db.DB.QueryRow("SELECT data FROM compendium_entries WHERE id=?", id).Scan(&dataJSON)
		if err != nil {
			errs = append(errs, map[string]interface{}{"id": id, "error": "not found"})
			continue
		}

		// Merge update into existing data
		var entryData map[string]interface{}
		json.Unmarshal([]byte(dataJSON), &entryData)
		for k, v := range req.Data {
			entryData[k] = v
		}

		newJSON, _ := json.Marshal(entryData)
		_, err = db.DB.Exec("UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=?", string(newJSON), id)
		if err != nil {
			errs = append(errs, map[string]interface{}{"id": id, "error": err.Error()})
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
	args := []interface{}{q}

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
		var data map[string]interface{}
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

func ImportCompendiumEntries(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}

	userID, _ := c.Get("user_id")

	// Read uploaded file or JSON body
	var entries []map[string]interface{}

	// Try multipart form first
	file, _, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
			return
		}
		if err := json.Unmarshal(body, &entries); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			return
		}
	} else {
		// Try JSON body
		var body []byte
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}
		if err := json.Unmarshal(body, &entries); err != nil {
			// Try single object
			var single map[string]interface{}
			if err2 := json.Unmarshal(body, &single); err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON array or object"})
				return
			}
			entries = []map[string]interface{}{single}
		}
	}

	if len(entries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entries to import"})
		return
	}

	// Validate required fields
	var schema models.CompendiumSchema
	var fieldsJSON string
	err = db.DB.QueryRow("SELECT type_name, display_name, fields FROM compendium_schemas WHERE id=?", schemaID).
		Scan(&schema.TypeName, &schema.DisplayName, &fieldsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	json.Unmarshal([]byte(fieldsJSON), &schema.Fields)

	// Validate
	requiredFields := make(map[string]bool)
	for _, f := range schema.Fields {
		if f.Required {
			requiredFields[f.Name] = true
		}
	}

	fieldErrors := []models.CompendiumImportError{}
	for i, entry := range entries {
		for fName := range requiredFields {
			if val, ok := entry[fName]; !ok || val == nil || fmt.Sprintf("%v", val) == "" {
				fieldErrors = append(fieldErrors, models.CompendiumImportError{
					Index:   i,
					Field:   fName,
					Message: fmt.Sprintf("missing required field: %s", fName),
				})
			}
		}
	}

	if len(fieldErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation errors",
			"errors": fieldErrors,
		})
		return
	}

	// Check duplicates (by name field)
	var existingNames []string
	if nameField := c.Query("name_field"); nameField != "" {
		rows, _ := db.DB.Query(`SELECT json_extract(data, '$."`+nameField+`"') FROM compendium_entries WHERE schema_id=?`, schemaID)
		if rows != nil {
			for rows.Next() {
				var name string
				rows.Scan(&name)
				if name != "" {
					existingNames = append(existingNames, strings.ToLower(name))
				}
			}
			rows.Close()
		}
	} else {
		// Auto-detect: look for a "name" field in data
		rows, _ := db.DB.Query(`SELECT json_extract(data, '$.name') FROM compendium_entries WHERE schema_id=?`, schemaID)
		if rows != nil {
			for rows.Next() {
				var name string
				rows.Scan(&name)
				if name != "" {
					existingNames = append(existingNames, strings.ToLower(name))
				}
			}
			rows.Close()
		}
	}

	existingSet := make(map[string]bool, len(existingNames))
	for _, n := range existingNames {
		existingSet[n] = true
	}

	duplicates := []models.CompendiumImportDuplicate{}
	cleanEntries := []map[string]interface{}{}
	skipCount := 0

	for i, entry := range entries {
		var entryName string
		if n, ok := entry["name"].(string); ok && n != "" {
			entryName = n
		}

		if entryName != "" && existingSet[strings.ToLower(entryName)] {
			// Look up existing entry
			var existingID int64
			var existingData string
			db.DB.QueryRow(`SELECT id, data FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`,
				schemaID, entryName).Scan(&existingID, &existingData)

			existingMap := make(map[string]interface{})
			if existingData != "" {
				json.Unmarshal([]byte(existingData), &existingMap)
			}

			duplicates = append(duplicates, models.CompendiumImportDuplicate{
				Index:      i,
				ExistingID: existingID,
				Existing:   existingMap,
				Incoming:   entry,
				Resolved:   "skip", // default: skip
			})
			skipCount++
			continue
		}
		cleanEntries = append(cleanEntries, entry)
	}

	// Insert clean entries
	inserted := 0
	for _, entry := range cleanEntries {
		dataJSON, _ := json.Marshal(entry)
		_, err := db.DB.Exec(`INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`, schemaID, string(dataJSON))
		if err == nil {
			inserted++
		}
	}

	// Log import
	filesJSON, _ := json.Marshal([]map[string]interface{}{
		{"filename": "upload", "entries": len(entries)},
	})
	summary := map[string]interface{}{
		"total":      len(entries),
		"inserted":   inserted,
		"duplicates": skipCount,
		"errors":     len(fieldErrors),
	}
	summaryJSON, _ := json.Marshal(summary)
	mappingJSON, _ := json.Marshal(map[string]string{})

	var logID int64
	if inserted > 0 || skipCount > 0 {
		result, err := db.DB.Exec(`INSERT INTO compendium_import_logs(user_id, status, files, mapping, summary) VALUES(?,?,?,?,?)`,
			userID, "completed", string(filesJSON), string(mappingJSON), string(summaryJSON))
		if err == nil {
			logID, _ = result.LastInsertId()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"import_log_id": logID,
		"total":         len(entries),
		"inserted":      inserted,
		"duplicates":    duplicates,
		"skipped":       skipCount,
		"errors":        fieldErrors,
	})
}

// ─── Export ───

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

	var rows interface{ Scan(...interface{}) error; Close() error; Next() bool }

	if filter != "" {
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
		ID        int64                  `json:"id"`
		Data      map[string]interface{} `json:"data"`
		CreatedAt string                 `json:"created_at"`
	}

	entries := make([]exportEntry, 0)
	for rows.Next() {
		var e exportEntry
		var dataJSON, createdAt string
		if err := rows.Scan(&e.ID, &dataJSON, &createdAt); err != nil {
			continue
		}
		e.Data = make(map[string]interface{})
		json.Unmarshal([]byte(dataJSON), &e.Data)
		e.CreatedAt = createdAt
		entries = append(entries, e)
	}

	switch format {
	case "json":
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-export.json"`, schemaType))
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"schema":  schemaType,
			"name":    displayName,
			"exported": time.Now().UTC().Format(time.RFC3339),
			"count":   len(entries),
			"entries": entries,
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
// hardcoded compendium types. Safe to call multiple times (INSERT OR IGNORE via type_name UNIQUE).
func SeedCompendiumSchemas() {
	schemas := []struct {
		TypeName    string
		DisplayName string
		Fields      []models.SchemaField
	}{
		{
			TypeName: "race", DisplayName: "Races",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "speed", Label: "Speed", Type: models.FieldTypeInteger, Sortable: true},
				{Name: "size", Label: "Size", Type: models.FieldTypeSelect, Options: []string{"Tiny", "Small", "Medium", "Large", "Huge", "Gargantuan"}},
				{Name: "ability_bonuses", Label: "Ability Bonuses", Type: models.FieldTypeText},
				{Name: "traits", Label: "Traits", Type: models.FieldTypeText},
				{Name: "languages", Label: "Languages", Type: models.FieldTypeText},
				{Name: "source_page", Label: "Source Page", Type: models.FieldTypeString},
				{Name: "system", Label: "System", Type: models.FieldTypeSelect, Options: []string{"dnd5e", "pf2e", "homebrew"}, Default: "dnd5e"},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "custom"}, Default: "srd"},
				{Name: "category", Label: "Category", Type: models.FieldTypeString},
				{Name: "expansion", Label: "Expansion", Type: models.FieldTypeString},
				{Name: "publisher", Label: "Publisher", Type: models.FieldTypeString},
			},
		},
		{
			TypeName: "class", DisplayName: "Classes",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "hit_die", Label: "Hit Die", Type: models.FieldTypeInteger},
				{Name: "primary_ability", Label: "Primary Ability", Type: models.FieldTypeString},
				{Name: "saving_throws", Label: "Saving Throws", Type: models.FieldTypeString},
				{Name: "proficiencies", Label: "Proficiencies", Type: models.FieldTypeText},
				{Name: "spellcasting_ability", Label: "Spellcasting Ability", Type: models.FieldTypeString},
				{Name: "source_page", Label: "Source Page", Type: models.FieldTypeString},
				{Name: "system", Label: "System", Type: models.FieldTypeSelect, Options: []string{"dnd5e", "pf2e", "homebrew"}, Default: "dnd5e"},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "custom"}, Default: "srd"},
				{Name: "category", Label: "Category", Type: models.FieldTypeString},
				{Name: "expansion", Label: "Expansion", Type: models.FieldTypeString},
				{Name: "publisher", Label: "Publisher", Type: models.FieldTypeString},
			},
		},
		{
			TypeName: "spell", DisplayName: "Spells",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "level", Label: "Level", Type: models.FieldTypeInteger, Sortable: true},
				{Name: "school", Label: "School", Type: models.FieldTypeSelect, Options: []string{"Abjuration", "Conjuration", "Divination", "Enchantment", "Evocation", "Illusion", "Necromancy", "Transmutation"}},
				{Name: "casting_time", Label: "Casting Time", Type: models.FieldTypeString},
				{Name: "range", Label: "Range", Type: models.FieldTypeString},
				{Name: "components", Label: "Components", Type: models.FieldTypeString},
				{Name: "duration", Label: "Duration", Type: models.FieldTypeString},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "higher_levels", Label: "Higher Levels", Type: models.FieldTypeText},
				{Name: "classes", Label: "Classes", Type: models.FieldTypeString, Searchable: true},
				{Name: "source_page", Label: "Source Page", Type: models.FieldTypeString},
				{Name: "system", Label: "System", Type: models.FieldTypeSelect, Options: []string{"dnd5e", "pf2e", "homebrew"}, Default: "dnd5e"},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "custom"}, Default: "srd"},
				{Name: "publisher", Label: "Publisher", Type: models.FieldTypeString},
			},
		},
		{
			TypeName: "feat", DisplayName: "Feats",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "prerequisites", Label: "Prerequisites", Type: models.FieldTypeText},
				{Name: "source_page", Label: "Source Page", Type: models.FieldTypeString},
				{Name: "system", Label: "System", Type: models.FieldTypeSelect, Options: []string{"dnd5e", "pf2e", "homebrew"}, Default: "dnd5e"},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "custom"}, Default: "srd"},
			},
		},
		{
			TypeName: "background", DisplayName: "Backgrounds",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "feature_name", Label: "Feature Name", Type: models.FieldTypeString},
				{Name: "feature_description", Label: "Feature Description", Type: models.FieldTypeText},
				{Name: "proficiencies", Label: "Proficiencies", Type: models.FieldTypeText},
				{Name: "source_page", Label: "Source Page", Type: models.FieldTypeString},
				{Name: "system", Label: "System", Type: models.FieldTypeSelect, Options: []string{"dnd5e", "pf2e", "homebrew"}, Default: "dnd5e"},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "custom"}, Default: "srd"},
				{Name: "category", Label: "Category", Type: models.FieldTypeString},
				{Name: "data_list", Label: "Data List", Type: models.FieldTypeBoolean},
				{Name: "data_bonds", Label: "Bonds", Type: models.FieldTypeText},
				{Name: "data_flaws", Label: "Flaws", Type: models.FieldTypeText},
				{Name: "data_ideals", Label: "Ideals", Type: models.FieldTypeText},
				{Name: "data_equipment", Label: "Equipment", Type: models.FieldTypeText},
				{Name: "data_starting_gold", Label: "Starting Gold", Type: models.FieldTypeInteger},
				{Name: "data_personality_traits", Label: "Personality Traits", Type: models.FieldTypeText},
				{Name: "publisher", Label: "Publisher", Type: models.FieldTypeString},
			},
		},
		{
			TypeName: "equipment", DisplayName: "Equipment",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "category", Label: "Category", Type: models.FieldTypeSelect, Options: []string{"weapon", "armor", "potion", "scroll", "wand", "ring", "wondrous", "gear", "tool", "other"}},
				{Name: "cost", Label: "Cost", Type: models.FieldTypeString},
				{Name: "weight", Label: "Weight", Type: models.FieldTypeFloat},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "source_page", Label: "Source Page", Type: models.FieldTypeString},
				{Name: "system", Label: "System", Type: models.FieldTypeSelect, Options: []string{"dnd5e", "pf2e", "homebrew"}, Default: "dnd5e"},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "custom"}, Default: "srd"},
				{Name: "item_type", Label: "Item Type", Type: models.FieldTypeString},
				{Name: "item_rarity", Label: "Item Rarity", Type: models.FieldTypeSelect, Options: []string{"common", "uncommon", "rare", "very rare", "legendary", "artifact"}},
				{Name: "publisher", Label: "Publisher", Type: models.FieldTypeString},
			},
		},
		{
			TypeName: "monster", DisplayName: "Monsters",
			Fields: []models.SchemaField{
				{Name: "name", Label: "Name", Type: models.FieldTypeString, Required: true, Sortable: true, Searchable: true},
				{Name: "type", Label: "Type", Type: models.FieldTypeString, Sortable: true},
				{Name: "size", Label: "Size", Type: models.FieldTypeSelect, Options: []string{"Tiny", "Small", "Medium", "Large", "Huge", "Gargantuan"}},
				{Name: "ac", Label: "Armor Class", Type: models.FieldTypeInteger},
				{Name: "hp", Label: "Hit Points", Type: models.FieldTypeInteger},
				{Name: "str", Label: "Strength", Type: models.FieldTypeInteger},
				{Name: "dex", Label: "Dexterity", Type: models.FieldTypeInteger},
				{Name: "con", Label: "Constitution", Type: models.FieldTypeInteger},
				{Name: "int", Label: "Intelligence", Type: models.FieldTypeInteger},
				{Name: "wis", Label: "Wisdom", Type: models.FieldTypeInteger},
				{Name: "cha", Label: "Charisma", Type: models.FieldTypeInteger},
				{Name: "cr", Label: "Challenge Rating", Type: models.FieldTypeString, Sortable: true},
				{Name: "source", Label: "Source", Type: models.FieldTypeSelect, Options: []string{"srd", "homebrew", "dnd5eapi", "custom"}, Default: "srd"},
				{Name: "saves", Label: "Saving Throws", Type: models.FieldTypeText},
				{Name: "skills", Label: "Skills", Type: models.FieldTypeText},
				{Name: "damage_vulnerabilities", Label: "Damage Vulnerabilities", Type: models.FieldTypeText},
				{Name: "damage_resistances", Label: "Damage Resistances", Type: models.FieldTypeText},
				{Name: "damage_immunities", Label: "Damage Immunities", Type: models.FieldTypeText},
				{Name: "condition_immunities", Label: "Condition Immunities", Type: models.FieldTypeText},
				{Name: "senses", Label: "Senses", Type: models.FieldTypeText},
				{Name: "languages", Label: "Languages", Type: models.FieldTypeText},
				{Name: "special_abilities", Label: "Special Abilities", Type: models.FieldTypeJSON},
				{Name: "actions", Label: "Actions", Type: models.FieldTypeJSON},
				{Name: "legendary_actions", Label: "Legendary Actions", Type: models.FieldTypeJSON},
				{Name: "description", Label: "Description", Type: models.FieldTypeText, Searchable: true},
				{Name: "alignment", Label: "Alignment", Type: models.FieldTypeSelect, Options: []string{"lawful good", "neutral good", "chaotic good", "lawful neutral", "neutral", "chaotic neutral", "lawful evil", "neutral evil", "chaotic evil", "unaligned"}},
				{Name: "expansion", Label: "Expansion", Type: models.FieldTypeString},
				{Name: "publisher", Label: "Publisher", Type: models.FieldTypeString},
			},
		},
	}

	for _, s := range schemas {
		fieldsJSON, _ := json.Marshal(s.Fields)
		_, err := db.DB.Exec(`INSERT OR IGNORE INTO compendium_schemas(type_name, display_name, fields) VALUES(?,?,?)`,
			s.TypeName, s.DisplayName, string(fieldsJSON))
		if err != nil {
			// Log but continue — schema may already exist
			fmt.Printf("Warning: seed schema %s: %v\n", s.TypeName, err)
		}
	}
}

// AutoSyncCompendiumEntry is called when syncing a legacy entry into the generic system.
// For now it's a no-op placeholder; full sync happens in Phase 2.
func AutoSyncCompendiumEntry(schemaID, entryID int64, data map[string]interface{}) {
	// Reserved for future use — writes metadata or triggers cross-system sync
}

// ─── Legacy Data Migration ───

// MigrateLegacyCompendiumData copies data from all 7 legacy compendium tables
// into the new compendium_entries table. Safe to call multiple times — skips
// entries whose name already exists in the target schema.
// Returns a summary of what was migrated.
func MigrateLegacyCompendiumData() map[string]interface{} {
	result := map[string]interface{}{}
	typeNameToID := map[string]int64{}
	rows, err := db.DB.Query("SELECT id, type_name FROM compendium_schemas")
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var tn string
		rows.Scan(&id, &tn)
		typeNameToID[tn] = id
	}

	legacyTables := []struct {
		Table     string
		Schema    string
		Columns   []string
		ColumnMap map[string]string // legacy_col → schema_field name override
	}{
		{
			Table: "compendium_races", Schema: "race",
			Columns: []string{"name", "description", "speed", "size", "ability_bonuses", "traits", "languages", "source_page"},
		},
		{
			Table: "compendium_classes", Schema: "class",
			Columns: []string{"name", "description", "hit_die", "primary_ability", "saving_throws", "proficiencies", "spellcasting_ability", "source_page"},
		},
		{
			Table: "compendium_spells", Schema: "spell",
			Columns: []string{"name", "level", "school", "casting_time", "range", "components", "duration", "description", "higher_levels", "classes", "source_page"},
		},
		{
			Table: "compendium_feats", Schema: "feat",
			Columns: []string{"name", "description", "prerequisites", "source_page"},
		},
		{
			Table: "compendium_backgrounds", Schema: "background",
			Columns: []string{"name", "description", "feature_name", "feature_description", "proficiencies", "source_page"},
		},
		{
			Table: "compendium_equipment", Schema: "equipment",
			Columns: []string{"name", "category", "cost", "weight", "description", "source_page"},
		},
		{
			Table: "compendium_monsters", Schema: "monster",
			Columns: []string{"name", "type", "size", "ac", "hp", "str", "dex", "con", "wis", "cha", "cr", "source",
				"saves", "skills", "damage_vulnerabilities", "damage_resistances", "damage_immunities",
				"condition_immunities", "senses", "languages", "special_abilities", "actions", "legendary_actions", "description"},
			ColumnMap: map[string]string{"int_": "int"},
		},
	}

	totalMigrated := 0
	totalSkipped := 0
	perSchema := map[string]int{}

	for _, lt := range legacyTables {
		schemaID, ok := typeNameToID[lt.Schema]
		if !ok {
			continue
		}

		colList := strings.Join(lt.Columns, ", ")
		query := fmt.Sprintf("SELECT %s FROM %s ORDER BY name", colList, lt.Table)
		sqlRows, err := db.DB.Query(query)
		if err != nil {
			result[lt.Schema+"_error"] = err.Error()
			continue
		}

		// Build scan targets dynamically
		colCount := len(lt.Columns)
		for sqlRows.Next() {
			scanTargets := make([]interface{}, colCount)
			scanStrings := make([]string, colCount)
			for i := range scanTargets {
				scanTargets[i] = &scanStrings[i]
			}
			if err := sqlRows.Scan(scanTargets...); err != nil {
				continue
			}

			// Build data map
			data := map[string]interface{}{}
			name := ""
			for i, col := range lt.Columns {
				val := strings.TrimSpace(scanStrings[i])
				// Map legacy column name to schema field name
				fieldName := col
				if lt.ColumnMap != nil {
					if mapped, ok := lt.ColumnMap[col]; ok {
						fieldName = mapped
					}
				}
				if col == "name" {
					name = val
				}
				// Try numeric conversion for numeric-looking values
				if val == "" {
					continue // skip empty fields
				}
				data[fieldName] = val
			}

			if name == "" {
				continue
			}

			// Check for duplicate by name
			var existing int
			// We approximate name matching by checking if any entry data contains "name":"<name>"
			// Since data is JSON, we use json_extract
			err := db.DB.QueryRow(
				`SELECT COUNT(*) FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`,
				schemaID, name).Scan(&existing)
			if err == nil && existing > 0 {
				totalSkipped++
				continue
			}

			dataJSON, _ := json.Marshal(data)
			_, err = db.DB.Exec(
				`INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`,
				schemaID, string(dataJSON))
			if err != nil {
				continue
			}
			totalMigrated++
			perSchema[lt.Schema]++
		}
		sqlRows.Close()
	}

	result["total_migrated"] = totalMigrated
	result["total_skipped"] = totalSkipped
	result["per_schema"] = perSchema
	return result
}

// HandleMigrateLegacy is an admin-only endpoint to trigger legacy data migration.
func HandleMigrateLegacy(c *gin.Context) {
	summary := MigrateLegacyCompendiumData()
	c.JSON(http.StatusOK, summary)
}

// ─── Configurable Import: Field Mapping ───

// DetectImportFields analyzes a JSON payload and returns all discovered fields
// (including nested keys via dot notation) plus auto-suggested mapping to schema fields.
func DetectImportFields(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}

	// Read the uploaded data
	var rawEntries []map[string]interface{}
	file, _, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		body, _ := io.ReadAll(file)
		json.Unmarshal(body, &rawEntries)
	} else {
		body, _ := io.ReadAll(c.Request.Body)
		json.Unmarshal(body, &rawEntries)
		// Try single object
		if len(rawEntries) == 0 {
			var single map[string]interface{}
			if json.Unmarshal(body, &single) == nil {
				rawEntries = []map[string]interface{}{single}
			}
		}
	}

	if len(rawEntries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entries found"})
		return
	}

	// Get schema fields for auto-suggest
	var fieldsJSON string
	err = db.DB.QueryRow("SELECT fields FROM compendium_schemas WHERE id=?", schemaID).Scan(&fieldsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	var schemaFields []models.SchemaField
	json.Unmarshal([]byte(fieldsJSON), &schemaFields)

	// Build schema field name set for matching
	schemaFieldNames := make([]string, len(schemaFields))
	for i, f := range schemaFields {
		schemaFieldNames[i] = f.Name
	}

	// Discover all fields from the first few entries (up to 5)
	discoveredFields := discoverJSONFields(rawEntries)

	// Auto-suggest mappings: case-insensitive match + fuzzy
	suggestions := make([]map[string]interface{}, 0)
	matched := map[string]bool{}
	for _, sf := range schemaFieldNames {
		sfLower := strings.ToLower(sf)
		for _, df := range discoveredFields {
			dfLower := strings.ToLower(df)
			if dfLower == sfLower || strings.ReplaceAll(dfLower, "_", "") == strings.ReplaceAll(sfLower, "_", "") {
				suggestions = append(suggestions, map[string]interface{}{
					"source":         df,
					"target":         sf,
					"auto_matched":   true,
					"confidence":     "high",
				})
				matched[df] = true
				break
			}
		}
	}

	// Add unmatched discovered fields
	for _, df := range discoveredFields {
		if !matched[df] {
			// Try partial match
			bestMatch := ""
			bestScore := 0
			dfLower := strings.ToLower(df)
			for _, sf := range schemaFieldNames {
				sfLower := strings.ToLower(sf)
				// Simple substring/length ratio scoring
				if strings.Contains(dfLower, sfLower) || strings.Contains(sfLower, dfLower) {
					score := len(sfLower)
					if score > bestScore {
						bestScore = score
						bestMatch = sf
					}
				}
			}
			suggestions = append(suggestions, map[string]interface{}{
				"source":         df,
				"target":         bestMatch,
				"auto_matched":   bestMatch != "",
				"confidence":     "medium",
			})
		}
	}

	// Add schema fields not matched at all
	matchedTargets := map[string]bool{}
	for _, s := range suggestions {
		if t, ok := s["target"].(string); ok && t != "" {
			matchedTargets[t] = true
		}
	}
	for _, sf := range schemaFieldNames {
		if !matchedTargets[sf] {
			// Check if it has a special field like name which might be in the data
			suggestions = append(suggestions, map[string]interface{}{
				"source":       "",
				"target":       sf,
				"auto_matched": false,
				"confidence":   "unmapped",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entry_count":    len(rawEntries),
		"sample_entry":   rawEntries[0],
		"discovered_fields": discoveredFields,
		"suggestions":    suggestions,
		"schema_fields":  schemaFieldNames,
	})
}

// discoverJSONFields recursively extracts all keys from a set of JSON objects
// using dot notation for nested objects (e.g., "properties.Level").
func discoverJSONFields(entries []map[string]interface{}) []string {
	fieldSet := map[string]bool{}
	limit := 5
	if len(entries) < limit {
		limit = len(entries)
	}
	for _, entry := range entries[:limit] {
		flattenKeys("", entry, fieldSet)
	}
	fields := make([]string, 0, len(fieldSet))
	for f := range fieldSet {
		fields = append(fields, f)
	}
	return fields
}

func flattenKeys(prefix string, data map[string]interface{}, out map[string]bool) {
	for k, v := range data {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		if sub, ok := v.(map[string]interface{}); ok {
			// Nested object: record the parent key and descend
			out[fullKey] = true
			flattenKeys(fullKey, sub, out)
		} else {
			out[fullKey] = true
		}
	}
}

// getNestedValue retrieves a value from a nested map using dot notation.
func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 1 {
		return data[parts[0]]
	}
	if sub, ok := data[parts[0]].(map[string]interface{}); ok {
		return getNestedValue(sub, parts[1])
	}
	return nil
}

// ImportCompendiumEntriesWithMapping is the enhanced import that accepts field mappings.
func ImportCompendiumEntriesWithMapping(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}

	userID, _ := c.Get("user_id")

	// Parse multipart form
	var rawEntries []map[string]interface{}
	var fieldMapping []struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	// Accept JSON body with entries + field_mapping
	var req struct {
		Entries      []map[string]interface{} `json:"entries"`
		FieldMapping []struct {
			Source string `json:"source"`
			Target string `json:"target"`
		} `json:"field_mapping"`
		DuplicateAction string `json:"duplicate_action"` // skip (default), overwrite, create-new
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Try file upload
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON body with 'entries' and 'field_mapping' or file upload"})
			return
		}
		defer file.Close()
		body, _ := io.ReadAll(file)
		json.Unmarshal(body, &rawEntries)
		if len(rawEntries) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no entries found in file"})
			return
		}
		// Parse field_mapping from form value
		mappingStr := c.PostForm("field_mapping")
		if mappingStr != "" {
			json.Unmarshal([]byte(mappingStr), &fieldMapping)
		}
	} else {
		rawEntries = req.Entries
		fieldMapping = req.FieldMapping
	}

	if len(rawEntries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entries to import"})
		return
	}

	// Get schema fields for validation
	var fieldsJSON string
	err = db.DB.QueryRow("SELECT fields FROM compendium_schemas WHERE id=?", schemaID).Scan(&fieldsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	var schemaFields []models.SchemaField
	json.Unmarshal([]byte(fieldsJSON), &schemaFields)

	requiredFieldNames := make(map[string]bool)
	for _, f := range schemaFields {
		if f.Required {
			requiredFieldNames[f.Name] = true
		}
	}

	// Apply field mapping to each entry
	duplicateAction := req.DuplicateAction
	if duplicateAction == "" {
		duplicateAction = c.DefaultQuery("duplicate_action", "skip")
	}

	mappedEntries := make([]map[string]interface{}, 0, len(rawEntries))
	fieldErrors := make([]models.CompendiumImportError, 0)

	for i, entry := range rawEntries {
		mapped := make(map[string]interface{})
		if len(fieldMapping) > 0 {
			for _, m := range fieldMapping {
				if m.Source == "" {
					continue
				}
				val := getNestedValue(entry, m.Source)
				if val != nil {
					mapped[m.Target] = val
				}
			}
		} else {
			// No mapping: try direct copy of all fields
			for k, v := range entry {
				// Skip non-data fields
				if k == "publisher" || k == "book" || k == "properties" {
					continue
				}
				if sub, ok := v.(map[string]interface{}); ok {
					// Flatten nested objects
					for sk, sv := range sub {
						mapped[sk] = sv
					}
				} else {
					mapped[k] = v
				}
			}
		}

		// Validate required fields
		for rf := range requiredFieldNames {
			if val, ok := mapped[rf]; !ok || val == nil || fmt.Sprintf("%v", val) == "" {
				fieldErrors = append(fieldErrors, models.CompendiumImportError{
					Index:   i,
					Field:   rf,
					Message: fmt.Sprintf("missing required field: %s", rf),
				})
			}
		}

		mappedEntries = append(mappedEntries, mapped)
	}

	// If there are validation errors, return them immediately
	if len(fieldErrors) > 0 && c.Query("validate_only") == "" {
		// But only abort if ?validate_only is NOT set (preview mode continues)
		if duplicateAction != "force" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "validation errors",
				"errors": fieldErrors,
			})
			return
		}
	}

	// Check duplicates (by name field)
	duplicates := make([]models.CompendiumImportDuplicate, 0)
	cleanEntries := make([]map[string]interface{}, 0)
	skipCount := 0

	existingNames := loadExistingNames(schemaID)
	existingSet := make(map[string]bool, len(existingNames))
	for _, n := range existingNames {
		existingSet[strings.ToLower(n)] = true
	}

	for i, entry := range mappedEntries {
		var entryName string
		if n, ok := entry["name"].(string); ok && n != "" {
			entryName = n
		}

		if entryName != "" && existingSet[strings.ToLower(entryName)] {
			var existingID int64
			var existingData string
			db.DB.QueryRow(`SELECT id, data FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`,
				schemaID, entryName).Scan(&existingID, &existingData)

			existingMap := make(map[string]interface{})
			if existingData != "" {
				json.Unmarshal([]byte(existingData), &existingMap)
			}

			if duplicateAction == "skip" {
				duplicates = append(duplicates, models.CompendiumImportDuplicate{
					Index:      i,
					ExistingID: existingID,
					Existing:   existingMap,
					Incoming:   entry,
					Resolved:   "skip",
				})
				skipCount++
				continue
			} else if duplicateAction == "overwrite" {
				dataJSON, _ := json.Marshal(entry)
				db.DB.Exec(`UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=?`,
					string(dataJSON), existingID)
				duplicates = append(duplicates, models.CompendiumImportDuplicate{
					Index:      i,
					ExistingID: existingID,
					Resolved:   "overwrite",
				})
				continue
			}
			// create-new: just falls through to insert
		}

		cleanEntries = append(cleanEntries, entry)
	}

	// Insert new entries
	inserted := 0
	for _, entry := range cleanEntries {
		dataJSON, _ := json.Marshal(entry)
		_, err := db.DB.Exec(`INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`, schemaID, string(dataJSON))
		if err == nil {
			inserted++
		}
	}

	// Log the import
	totalEntries := len(rawEntries)
	filesJSON, _ := json.Marshal([]map[string]interface{}{
		{"filename": "upload", "entries": totalEntries},
	})
	summary := map[string]interface{}{
		"total":         totalEntries,
		"mapped":        len(mappedEntries),
		"inserted":      inserted,
		"duplicates":    skipCount,
		"overwritten":   len(duplicates) - skipCount,
		"validation_errors": len(fieldErrors),
	}
	summaryJSON, _ := json.Marshal(summary)
	mappingJSON, _ := json.Marshal(fieldMapping)

	var logID int64
	if inserted > 0 || skipCount > 0 {
		result, err := db.DB.Exec(`INSERT INTO compendium_import_logs(user_id, status, files, mapping, summary) VALUES(?,?,?,?,?)`,
			userID, "completed", string(filesJSON), string(mappingJSON), string(summaryJSON))
		if err == nil {
			logID, _ = result.LastInsertId()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"import_log_id": logID,
		"total":         totalEntries,
		"mapped":        len(mappedEntries),
		"inserted":      inserted,
		"duplicates":    duplicates,
		"skipped":       skipCount,
		"errors":        fieldErrors,
		"summary":       summary,
	})
}

// ImportCompendiumBatchJSON handles the frontend admin.ts POST /api/admin/compendium-import.
// The frontend sends schema_id in the body and uses different field names than the
// existing /compendium-schemas/:id/import/with-mapping route.
func ImportCompendiumBatchJSON(c *gin.Context) {
	var req struct {
		SchemaID       int64 `json:"schema_id"`
		Entries        []map[string]interface{} `json:"entries"`
		DedupAction    string `json:"dedup_action"`
		FieldMapping   []struct {
			SourceField string `json:"source_field"`
			SchemaField string `json:"schema_field"`
		} `json:"field_mapping"`
		Filename string `json:"filename"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if req.SchemaID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schema_id is required"})
		return
	}
	if len(req.Entries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entries to import"})
		return
	}

	userID, _ := c.Get("user_id")

	// Get schema fields for validation
	var fieldsJSON string
	err := db.DB.QueryRow("SELECT fields FROM compendium_schemas WHERE id=?", req.SchemaID).Scan(&fieldsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	var schemaFields []models.SchemaField
	json.Unmarshal([]byte(fieldsJSON), &schemaFields)

	requiredFieldNames := make(map[string]bool)
	for _, f := range schemaFields {
		if f.Required {
			requiredFieldNames[f.Name] = true
		}
	}

	// Convert frontend mapping format (source_field/schema_field) to internal (source/target)
	type fieldMap struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	var fieldMapping []fieldMap
	for _, m := range req.FieldMapping {
		if m.SourceField == "" || m.SchemaField == "" {
			continue
		}
		fieldMapping = append(fieldMapping, fieldMap{Source: m.SourceField, Target: m.SchemaField})
	}

	duplicateAction := req.DedupAction
	if duplicateAction == "" {
		duplicateAction = "skip"
	}

	// Apply field mapping to each entry
	mappedEntries := make([]map[string]interface{}, 0, len(req.Entries))
	fieldErrors := make([]models.CompendiumImportError, 0)

	for i, entry := range req.Entries {
		mapped := make(map[string]interface{})
		if len(fieldMapping) > 0 {
			for _, m := range fieldMapping {
				val := getNestedValue(entry, m.Source)
				if val != nil {
					mapped[m.Target] = val
				}
			}
		} else {
			for k, v := range entry {
				if k == "publisher" || k == "book" || k == "properties" {
					continue
				}
				if sub, ok := v.(map[string]interface{}); ok {
					for sk, sv := range sub {
						mapped[sk] = sv
					}
				} else {
					mapped[k] = v
				}
			}
		}

		// Validate required fields
		for rf := range requiredFieldNames {
			if val, ok := mapped[rf]; !ok || val == nil || fmt.Sprintf("%v", val) == "" {
				fieldErrors = append(fieldErrors, models.CompendiumImportError{
					Index:   i,
					Field:   rf,
					Message: fmt.Sprintf("missing required field: %s", rf),
				})
			}
		}

		mappedEntries = append(mappedEntries, mapped)
	}

	// Check duplicates (by name field)
	skipCount := 0
	cleanEntries := make([]map[string]interface{}, 0)

	existingNames := loadExistingNames(req.SchemaID)
	existingSet := make(map[string]bool, len(existingNames))
	for _, n := range existingNames {
		existingSet[strings.ToLower(n)] = true
	}

	for _, entry := range mappedEntries {
		var entryName string
		if n, ok := entry["name"].(string); ok && n != "" {
			entryName = n
		}

		if entryName != "" && existingSet[strings.ToLower(entryName)] {
			if duplicateAction == "skip" {
				skipCount++
				continue
			} else if duplicateAction == "overwrite" {
				var existingID int64
				db.DB.QueryRow(`SELECT id FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`,
					req.SchemaID, entryName).Scan(&existingID)
				if existingID > 0 {
					dataJSON, _ := json.Marshal(entry)
					db.DB.Exec(`UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=?`,
						string(dataJSON), existingID)
				}
				continue
			}
			// create-new: falls through
		}
		cleanEntries = append(cleanEntries, entry)
	}

	// Insert new entries
	inserted := 0
	for _, entry := range cleanEntries {
		dataJSON, _ := json.Marshal(entry)
		_, err := db.DB.Exec(`INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`, req.SchemaID, string(dataJSON))
		if err == nil {
			inserted++
		}
	}

	// Log the import
	totalEntries := len(req.Entries)
	filesJSON, _ := json.Marshal([]map[string]interface{}{
		{"filename": req.Filename, "entries": totalEntries},
	})
	summary := map[string]interface{}{
		"total":             totalEntries,
		"mapped":            len(mappedEntries),
		"inserted":          inserted,
		"duplicates_skipped": skipCount,
		"validation_errors": len(fieldErrors),
	}
	summaryJSON, _ := json.Marshal(summary)

	// Build mapping JSON for the log using the converted mapping
	mappingJSON, _ := json.Marshal(fieldMapping)

	var logID int64
	if inserted > 0 || skipCount > 0 {
		result, err := db.DB.Exec(`INSERT INTO compendium_import_logs(user_id, status, files, mapping, summary) VALUES(?,?,?,?,?)`,
			userID, "completed", string(filesJSON), string(mappingJSON), string(summaryJSON))
		if err == nil {
			logID, _ = result.LastInsertId()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"import_log_id": logID,
		"imported":      inserted,
		"total":         totalEntries,
		"mapped":        len(mappedEntries),
		"skipped":       skipCount,
		"errors":        fieldErrors,
		"summary":       summary,
	})
}

func loadExistingNames(schemaID int64) []string {
	rows, err := db.DB.Query(`SELECT json_extract(data, '$.name') FROM compendium_entries WHERE schema_id=?`, schemaID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// ─── Helpers ───

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
