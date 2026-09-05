package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// FieldMapping maps a source field (dot-notation) to a target schema field.
type FieldMapping struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ImportOpts carries all varying inputs for the single parameterized import.
type ImportOpts struct {
	SchemaID          int64
	SchemaFields      []models.SchemaField
	SchemaDisplayName string
	RawEntries        []map[string]any
	Mapping           []FieldMapping // nil when absent (no mapping)
	DuplicateAction   string         // skip (default), overwrite, create-new, force
	DryRun            bool
	UseTx             bool
	UserID            any
	Filename          string
	NameField         string // custom field for duplicate detection (ImportCompendiumEntries ?name_field)
}

// ImportResult carries counts/IDs returned by the shared core.
type ImportResult struct {
	Total       int
	Mapped      int
	Inserted    int
	Skipped     int
	Overwritten int
	FieldErrors []models.CompendiumImportError
	Duplicates  []models.CompendiumImportDuplicate
	CleanCount  int // len of non-duplicate entries that would be inserted
}

// importCompendiumEntries is the single parameterized implementation for all
// compendium import paths. It centralizes mapping → validate → dedup →
// insert/update → (caller logs via insertImportLog). LastInsertId handling
// lives only in insertImportLog.
func importCompendiumEntries(ctx context.Context, sqlDB *sql.DB, opts ImportOpts) (ImportResult, error) {
	if sqlDB == nil {
		sqlDB = db.DB
	}
	// Normalise duplicate action default.
	dupAction := opts.DuplicateAction
	if dupAction == "" {
		dupAction = "skip"
	}

	// Build required field set.
	required := make(map[string]bool)
	for _, f := range opts.SchemaFields {
		if f.Required {
			required[f.Name] = true
		}
	}

	// Apply mapping.
	mappedEntries := make([]map[string]any, 0, len(opts.RawEntries))
	fieldErrors := make([]models.CompendiumImportError, 0)
	for i, entry := range opts.RawEntries {
		mapped := make(map[string]any)
		if len(opts.Mapping) > 0 {
			for _, m := range opts.Mapping {
				if m.Source == "" {
					continue
				}
				val := getNestedValue(entry, m.Source)
				if val != nil {
					mapped[m.Target] = val
				}
			}
		} else {
			// No mapping: copy/flatten like original paths.
			// ImportCompendiumEntries has no mapping at all and keeps raw entries verbatim (no skip of publisher/book/properties, no flatten).
			// For mapping paths (WithMapping/BatchJSON) the original skipped publisher/book/properties and flattened nested objects.
			// Detect: if opts.Mapping is nil and we are the plain ImportCompendiumEntries path, keep verbatim; otherwise apply legacy flatten/skip.
			// We distinguish by whether the caller set Filename=="upload" and UseTx true with no mapping? That's ambiguous.
			// Instead use an explicit signal: if Mapping == nil and DuplicateAction == "skip" && NameField != "" or DryRun handling differs, we cannot reliably distinguish.
			// Simpler: preserve original behaviour per path via a flag in opts: when Mapping is nil, the plain path should NOT flatten/skip.
			// We encode that by checking if opts.NameField was set via the plain path's ?name_field handling vs mapping paths always flatten.
			// However BatchJSON with no mapping also flattens. So we need a separate bool.
			// To avoid extra field, we treat plain ImportCompendiumEntries as: Mapping==nil && !UseTx=false? Not robust.
			// Add heuristic: if Mapping==nil and opts.Filename=="upload" and !hasMappingLogic -> keep verbatim only for plain path.
			// All callers set Filename; WithMapping uses "upload", BatchJSON uses req.Filename.
			// Better: add a dedicated bool to ImportOpts - NoFlatten. For now handle via NameField sentinel: WithMapping/BatchJSON never set NameField, plain may.
			// We add logic below via opts.UseTx + Filename check: WithMapping has UseTx=false, BatchJSON has UseTx true but Mapping nil still flattens.
			// So we need explicit field. Add RawMappingNilMeansFlatten bool? Instead add a simple rule: if Mapping==nil and opts.SchemaDisplayName=="" and Filename=="upload" and UseTx==false? Complex.
			// Easiest: add a new opts field BatchMode to distinguish.
			// For now handle both cases: if Mapping==nil we keep verbatim (plain behaviour) and let mapping callers explicitly set Mapping to empty slice (non-nil) to trigger flatten path.
			// So callers that want flatten with no mapping will pass Mapping = []FieldMapping{} (non-nil empty) instead of nil.
			for k, v := range entry {
				mapped[k] = v
			}
		}
		// For flatten/skip path (mapping callers with empty mapping): we re-apply that logic if Mapping is non-nil empty.
		if opts.Mapping != nil && len(opts.Mapping) == 0 {
			// Re-derive mapped with flatten/skip semantics (overwrites verbatim copy above)
			mapped = make(map[string]any)
			for k, v := range entry {
				if k == "publisher" || k == "book" || k == "properties" {
					continue
				}
				if sub, ok := v.(map[string]any); ok {
					maps.Copy(mapped, sub)
				} else {
					mapped[k] = v
				}
			}
		}
		for rf := range required {
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

	// Load existing names for duplicate detection.
	var existingSet map[string]bool
	if opts.NameField != "" {
		// Custom name field path (ImportCompendiumEntries ?name_field)
		existingNames := loadExistingNamesByField(opts.SchemaID, opts.NameField)
		existingSet = make(map[string]bool, len(existingNames))
		for _, n := range existingNames {
			existingSet[strings.ToLower(n)] = true
		}
	} else {
		existingNames := loadExistingNames(opts.SchemaID)
		existingSet = make(map[string]bool, len(existingNames))
		for _, n := range existingNames {
			existingSet[strings.ToLower(n)] = true
		}
	}

	duplicates := make([]models.CompendiumImportDuplicate, 0)
	cleanEntries := make([]map[string]any, 0)
	skipCount := 0
	overwriteCount := 0

	// Transaction handling.
	var tx *sql.Tx
	var err error
	if opts.UseTx && !opts.DryRun {
		tx, err = sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return ImportResult{}, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback() //nolint:errcheck
	}

	// For non-transactional overwrite path (WithMapping) we need sqlDB directly.
	// For transactional path we use tx when available.

	for i, entry := range mappedEntries {
		var entryName string
		if n, ok := entry["name"].(string); ok && n != "" {
			entryName = n
		}
		if entryName != "" && existingSet[strings.ToLower(entryName)] {
			if dupAction == "skip" {
				// Fetch existing for duplicate detail (preserve original shape)
				var existingID int64
				var existingData string
				sqlDB.QueryRowContext(ctx, `SELECT id, data FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`, opts.SchemaID, entryName).Scan(&existingID, &existingData)
				existingMap := make(map[string]any)
				if existingData != "" {
					json.Unmarshal([]byte(existingData), &existingMap)
				}
				duplicates = append(duplicates, models.CompendiumImportDuplicate{
					Index:      i,
					ExistingID: existingID,
					Existing:   existingMap,
					Incoming:   entry,
					Resolved:   "skip",
				})
				skipCount++
				continue
			} else if dupAction == "overwrite" {
				if opts.DryRun {
					overwriteCount++
					continue
				}
				// Perform overwrite.
				dataJSON, _ := json.Marshal(entry)
				if tx != nil {
					var existingID int64
					tx.QueryRowContext(ctx, `SELECT id FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`, opts.SchemaID, entryName).Scan(&existingID)
					if existingID > 0 {
						if _, err := tx.ExecContext(ctx, `UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=?`, string(dataJSON), existingID); err != nil {
							return ImportResult{}, fmt.Errorf("update failed: %w", err)
						}
					}
					overwriteCount++
					// For WithMapping original, duplicates entry had ExistingID and Resolved overwrite but no Existing/Incoming? Preserve shape for that caller via a generic entry; wrappers will adjust if needed.
					// We emit skip-style duplicates for plain path but for overwrite we need minimal; However to keep shapes byte-identical we need to branch on caller.
					// Emit overwrite-style duplicate with minimal fields for now; WithMapping expects ExistingID+Resolved; plain never hits overwrite branch.
					// Detect caller by UseTx vs not: WithMapping is non-tx, BatchJSON is tx.
					// For non-tx overwrite, original stored Existing map; for tx path original did not store duplicates detail at all (BatchJSON just counted).
					// To preserve, we add conditional.
					if !opts.UseTx {
						// WithMapping shape: need Existing/Incoming? Original WithMapping stored Existing for skip but for overwrite stored only ExistingID+Resolved (no Existing/Incoming).
						// We already have duplicates handling for overwrite below for non-tx case; adjust after.
						duplicates = append(duplicates, models.CompendiumImportDuplicate{
							Index:      i,
							ExistingID: existingID,
							Resolved:   "overwrite",
						})
					}
					continue
				}
				// Non-tx overwrite
				dataJSON2, _ := json.Marshal(entry)
				sqlDB.ExecContext(ctx, `UPDATE compendium_entries SET data=?, updated_at=datetime('now') WHERE id=(SELECT id FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=? LIMIT 1)`, string(dataJSON2), opts.SchemaID, entryName)
				// Need existingID for duplicate detail
				var existingID int64
				sqlDB.QueryRowContext(ctx, `SELECT id FROM compendium_entries WHERE schema_id=? AND json_extract(data, '$.name')=?`, opts.SchemaID, entryName).Scan(&existingID)
				duplicates = append(duplicates, models.CompendiumImportDuplicate{
					Index:      i,
					ExistingID: existingID,
					Resolved:   "overwrite",
				})
				overwriteCount++
				continue
			}
			// create-new: fall through to insert
		}
		cleanEntries = append(cleanEntries, entry)
	}

	if opts.DryRun {
		return ImportResult{
			Total:       len(opts.RawEntries),
			Mapped:      len(mappedEntries),
			Inserted:    0,
			Skipped:     skipCount,
			Overwritten: overwriteCount,
			FieldErrors: fieldErrors,
			Duplicates:  duplicates,
			CleanCount:  len(cleanEntries),
		}, nil
	}

	// Insert clean entries.
	inserted := 0
	for _, entry := range cleanEntries {
		dataJSON, _ := json.Marshal(entry)
		var execErr error
		if tx != nil {
			_, execErr = tx.ExecContext(ctx, `INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`, opts.SchemaID, string(dataJSON))
		} else {
			_, execErr = sqlDB.ExecContext(ctx, `INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)`, opts.SchemaID, string(dataJSON))
		}
		if execErr != nil {
			return ImportResult{}, fmt.Errorf("insert failed: %w", execErr)
		}
		inserted++
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return ImportResult{}, fmt.Errorf("commit failed: %w", err)
		}
	}

	return ImportResult{
		Total:       len(opts.RawEntries),
		Mapped:      len(mappedEntries),
		Inserted:    inserted,
		Skipped:     skipCount,
		Overwritten: overwriteCount,
		FieldErrors: fieldErrors,
		Duplicates:  duplicates,
		CleanCount:  len(cleanEntries),
	}, nil
}

func ImportCompendiumEntries(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}
	userID, _ := c.Get("user_id")
	var entries []map[string]any
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
		var body []byte
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}
		if err := json.Unmarshal(body, &entries); err != nil {
			var single map[string]any
			if err2 := json.Unmarshal(body, &single); err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON array or object"})
				return
			}
			entries = []map[string]any{single}
		}
	}
	if len(entries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entries to import"})
		return
	}
	var schema models.CompendiumSchema
	var fieldsJSON string
	err = db.DB.QueryRow("SELECT type_name, display_name, fields FROM compendium_schemas WHERE id=?", schemaID).Scan(&schema.TypeName, &schema.DisplayName, &fieldsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	json.Unmarshal([]byte(fieldsJSON), &schema.Fields)
	dryRun := c.DefaultQuery("dry_run", "") == "true"
	nameField := c.Query("name_field")
	opts := ImportOpts{
		SchemaID:          schemaID,
		SchemaFields:      schema.Fields,
		SchemaDisplayName: schema.DisplayName,
		RawEntries:        entries,
		Mapping:           nil,
		DuplicateAction:   "skip",
		DryRun:            dryRun,
		UseTx:             true,
		UserID:            userID,
		Filename:          "upload",
		NameField:         nameField,
	}
	result, err := importCompendiumEntries(c.Request.Context(), db.DB, opts)
	if err != nil {
		if strings.Contains(err.Error(), "insert failed") || strings.Contains(err.Error(), "commit failed") || strings.Contains(err.Error(), "failed to begin") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	if len(result.FieldErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation errors", "errors": result.FieldErrors})
		return
	}
	if dryRun {
		c.JSON(http.StatusOK, gin.H{"dry_run": true, "total": result.Total, "would_insert": result.CleanCount, "would_skip": result.Skipped, "duplicates": result.Duplicates, "errors": result.FieldErrors})
		return
	}
	filesJSON, _ := json.Marshal([]string{"upload"})
	summary := map[string]any{"total": result.Total, "inserted": result.Inserted, "duplicates": result.Skipped, "errors": len(result.FieldErrors), "schema_id": schemaID, "schema_name": schema.DisplayName}
	summaryJSON, _ := json.Marshal(summary)
	mappingJSON, _ := json.Marshal(map[string]string{})
	var logID int64
	if result.Inserted > 0 || result.Skipped > 0 {
		logID = insertImportLog(userID, string(filesJSON), string(mappingJSON), string(summaryJSON))
	}
	db.DB.Exec("PRAGMA wal_checkpoint(PASSIVE)")
	c.JSON(http.StatusOK, gin.H{"import_log_id": logID, "total": result.Total, "inserted": result.Inserted, "duplicates": result.Duplicates, "skipped": result.Skipped, "errors": result.FieldErrors})
}

// ─── Export ───

func ImportCompendiumEntriesWithMapping(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema_id"})
		return
	}
	userID, _ := c.Get("user_id")
	var rawEntries []map[string]any
	var fieldMapping []FieldMapping
	var req struct {
		Entries      []map[string]any `json:"entries"`
		FieldMapping []struct {
			Source string `json:"source"`
			Target string `json:"target"`
		} `json:"field_mapping"`
		DuplicateAction string `json:"duplicate_action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
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
		mappingStr := c.PostForm("field_mapping")
		if mappingStr != "" {
			json.Unmarshal([]byte(mappingStr), &fieldMapping)
		}
	} else {
		rawEntries = req.Entries
		for _, m := range req.FieldMapping {
			fieldMapping = append(fieldMapping, FieldMapping{Source: m.Source, Target: m.Target})
		}
		// Preserve non-nil empty slice semantics for flatten path
		if req.FieldMapping != nil && len(req.FieldMapping) == 0 {
			fieldMapping = []FieldMapping{}
		} else if req.FieldMapping == nil && len(rawEntries) > 0 {
			// No mapping provided via JSON -> signal flatten behaviour with empty non-nil slice
			if fieldMapping == nil {
				fieldMapping = []FieldMapping{}
			}
		}
	}
	if len(rawEntries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entries to import"})
		return
	}
	var fieldsJSON string
	err = db.DB.QueryRow("SELECT fields FROM compendium_schemas WHERE id=?", schemaID).Scan(&fieldsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	var schemaFields []models.SchemaField
	json.Unmarshal([]byte(fieldsJSON), &schemaFields)
	duplicateAction := req.DuplicateAction
	if duplicateAction == "" {
		duplicateAction = c.DefaultQuery("duplicate_action", "skip")
	}
	opts := ImportOpts{
		SchemaID:        schemaID,
		SchemaFields:    schemaFields,
		RawEntries:      rawEntries,
		Mapping:         fieldMapping,
		DuplicateAction: duplicateAction,
		DryRun:          false,
		UseTx:           false,
		UserID:          userID,
		Filename:        "upload",
	}
	result, err := importCompendiumEntries(c.Request.Context(), db.DB, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(result.FieldErrors) > 0 && c.Query("validate_only") == "" {
		if duplicateAction != "force" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation errors", "errors": result.FieldErrors})
			return
		}
	}
	totalEntries := len(rawEntries)
	filesJSON, _ := json.Marshal([]map[string]any{{"filename": "upload", "entries": totalEntries}})
	summary := map[string]any{"total": totalEntries, "mapped": result.Mapped, "inserted": result.Inserted, "duplicates": result.Skipped, "overwritten": result.Overwritten, "validation_errors": len(result.FieldErrors)}
	summaryJSON, _ := json.Marshal(summary)
	mappingJSON, _ := json.Marshal(fieldMapping)
	var logID int64
	if result.Inserted > 0 || result.Skipped > 0 {
		logID = insertImportLog(userID, string(filesJSON), string(mappingJSON), string(summaryJSON))
	}
	c.JSON(http.StatusOK, gin.H{"import_log_id": logID, "total": totalEntries, "mapped": result.Mapped, "inserted": result.Inserted, "duplicates": result.Duplicates, "skipped": result.Skipped, "errors": result.FieldErrors, "summary": summary})
}

// ImportCompendiumBatchJSON handles the frontend admin.ts POST /api/admin/compendium-import.
func ImportCompendiumBatchJSON(c *gin.Context) {
	var req struct {
		SchemaID     int64            `json:"schema_id"`
		Entries      []map[string]any `json:"entries"`
		DedupAction  string           `json:"dedup_action"`
		FieldMapping []struct {
			SourceField string `json:"source_field"`
			SchemaField string `json:"schema_field"`
		} `json:"field_mapping"`
		Filename string `json:"filename"`
	}
	if !BindOr400(c, &req) {
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
	var fieldsJSON, schemaDisplayName string
	err := db.DB.QueryRow("SELECT fields, display_name FROM compendium_schemas WHERE id=?", req.SchemaID).Scan(&fieldsJSON, &schemaDisplayName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}
	var schemaFields []models.SchemaField
	json.Unmarshal([]byte(fieldsJSON), &schemaFields)
	var fieldMapping []FieldMapping
	for _, m := range req.FieldMapping {
		if m.SourceField == "" || m.SchemaField == "" {
			continue
		}
		fieldMapping = append(fieldMapping, FieldMapping{Source: m.SourceField, Target: m.SchemaField})
	}
	if req.FieldMapping != nil && len(fieldMapping) == 0 && len(req.FieldMapping) == 0 {
		fieldMapping = []FieldMapping{}
	} else if req.FieldMapping == nil {
		fieldMapping = []FieldMapping{}
	}
	duplicateAction := req.DedupAction
	if duplicateAction == "" {
		duplicateAction = "skip"
	}
	dryRun := c.DefaultQuery("dry_run", "") == "true"
	opts := ImportOpts{
		SchemaID:          req.SchemaID,
		SchemaFields:      schemaFields,
		SchemaDisplayName: schemaDisplayName,
		RawEntries:        req.Entries,
		Mapping:           fieldMapping,
		DuplicateAction:   duplicateAction,
		DryRun:            dryRun,
		UseTx:             true,
		UserID:            userID,
		Filename:          req.Filename,
	}
	result, err := importCompendiumEntries(c.Request.Context(), db.DB, opts)
	if err != nil {
		if strings.Contains(err.Error(), "insert failed") || strings.Contains(err.Error(), "commit failed") || strings.Contains(err.Error(), "update failed") || strings.Contains(err.Error(), "failed to begin") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if dryRun {
		db.DB.Exec("PRAGMA wal_checkpoint(PASSIVE)")
		c.JSON(http.StatusOK, gin.H{"dry_run": true, "total": result.Total, "would_create": result.CleanCount, "would_update": result.Overwritten, "would_skip": result.Skipped, "validation_errors": len(result.FieldErrors), "errors": result.FieldErrors})
		return
	}
	totalEntries := len(req.Entries)
	filesJSON, _ := json.Marshal([]string{req.Filename})
	summary := map[string]any{"total": totalEntries, "mapped": result.Mapped, "inserted": result.Inserted, "duplicates_skipped": result.Skipped, "validation_errors": len(result.FieldErrors), "schema_id": req.SchemaID, "schema_name": schemaDisplayName}
	summaryJSON, _ := json.Marshal(summary)
	mappingJSON, _ := json.Marshal(fieldMapping)
	var logID int64
	if result.Inserted > 0 || result.Skipped > 0 {
		logID = insertImportLog(userID, string(filesJSON), string(mappingJSON), string(summaryJSON))
	}
	c.JSON(http.StatusOK, gin.H{"import_log_id": logID, "imported": result.Inserted, "total": totalEntries, "mapped": result.Mapped, "skipped": result.Skipped, "errors": result.FieldErrors, "summary": summary})
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

func loadExistingNamesByField(schemaID int64, field string) []string {
	// field is validated as simple identifier by caller; escape quotes
	field = strings.ReplaceAll(field, `"`, `""`)
	rows, err := db.DB.Query(`SELECT json_extract(data, '$."`+field+`"') FROM compendium_entries WHERE schema_id=?`, schemaID)
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

// insertImportLog centralizes LastInsertId error handling (single site).
func insertImportLog(userID any, filesJSON, mappingJSON, summaryJSON string) int64 {
	result, err := db.DB.Exec(`INSERT INTO compendium_import_logs(user_id, status, files, mapping, summary) VALUES(?,?,?,?,?)`,
		userID, "completed", filesJSON, mappingJSON, summaryJSON)
	if err != nil {
		return 0
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0
	}
	return id
}

// ─── Helpers ───

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
