package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── HTMX: Entry Table Partial ───

type htmxEntryTableData struct {
	Schema  *models.CompendiumSchema
	Entries []models.CompendiumEntry
	Page    int
	Total   int
	Pages   int
	Query   string
}

func HtmxCompendiumEntryTable(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("schemaId"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid schema id")
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
	q := strings.TrimSpace(c.Query("q"))

	// Get schema
	var schema models.CompendiumSchema
	var fieldsJSON, createdAt, updatedAt string
	err = db.DB.QueryRow(`SELECT id, type_name, display_name, fields, created_at, updated_at FROM compendium_schemas WHERE id=?`, schemaID).
		Scan(&schema.ID, &schema.TypeName, &schema.DisplayName, &fieldsJSON, &createdAt, &updatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "schema not found")
		return
	}
	schema.CreatedAt = createdAt
	schema.UpdatedAt = updatedAt
	json.Unmarshal([]byte(fieldsJSON), &schema.Fields)

	var entries []models.CompendiumEntry
	var total int

	if q != "" {
		db.DB.QueryRow(`SELECT COUNT(*) FROM compendium_entries_fts f
			JOIN compendium_entries e ON f.rowid = e.id
			WHERE e.schema_id=? AND compendium_entries_fts MATCH ?`, schemaID, q).Scan(&total)

		rows, err := db.DB.Query(`SELECT e.id, e.schema_id, e.data, e.created_at, e.updated_at
			FROM compendium_entries e
			JOIN compendium_entries_fts f ON e.id = f.rowid
			WHERE e.schema_id=? AND compendium_entries_fts MATCH ?
			ORDER BY e.created_at DESC LIMIT ? OFFSET ?`, schemaID, q, pageSize, offset)
		if err == nil {
			defer rows.Close()
			entries = scanCompendiumEntries(rows)
		}
	} else {
		db.DB.QueryRow("SELECT COUNT(*) FROM compendium_entries WHERE schema_id=?", schemaID).Scan(&total)

		rows, err := db.DB.Query(`SELECT id, schema_id, data, created_at, updated_at
			FROM compendium_entries WHERE schema_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			schemaID, pageSize, offset)
		if err == nil {
			defer rows.Close()
			entries = scanCompendiumEntries(rows)
		}
	}

	totalPages := max((total+pageSize-1)/pageSize, 1)

	renderTemplate(c, "compendium_entry_table", htmxEntryTableData{
		Schema:  &schema,
		Entries: entries,
		Page:    page,
		Total:   total,
		Pages:   totalPages,
		Query:   q,
	})
}

func scanCompendiumEntries(rows interface {
	Scan(...any) error
	Next() bool
	Close() error
}) []models.CompendiumEntry {
	out := make([]models.CompendiumEntry, 0)
	for rows.Next() {
		var e models.CompendiumEntry
		var dataJSON, createdAt, updatedAt string
		if err := rows.Scan(&e.ID, &e.SchemaID, &dataJSON, &createdAt, &updatedAt); err != nil {
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

// ─── HTMX: Entry Editor Form Partial ───

type htmxEntryEditorData struct {
	Schema *models.CompendiumSchema
	Entry  *models.CompendiumEntry
	Data   map[string]any
}

func HtmxCompendiumEntryEditor(c *gin.Context) {
	entryIDStr := c.Param("id")
	schemaIDStr := c.Query("schema_id")

	var data map[string]any
	var schema models.CompendiumSchema

	if entryIDStr != "" && entryIDStr != "0" {
		entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid entry id")
			return
		}

		var e models.CompendiumEntry
		var dataJSON, createdAt, updatedAt string
		err = db.DB.QueryRow(`SELECT id, schema_id, data, created_at, updated_at FROM compendium_entries WHERE id=?`, entryID).
			Scan(&e.ID, &e.SchemaID, &dataJSON, &createdAt, &updatedAt)
		if err != nil {
			c.String(http.StatusNotFound, "entry not found")
			return
		}
		e.Data = make(map[string]any)
		json.Unmarshal([]byte(dataJSON), &e.Data)
		e.CreatedAt = createdAt
		e.UpdatedAt = updatedAt
		data = e.Data

		// Get schema
		var fieldsJSON, created, updated string
		db.DB.QueryRow(`SELECT id, type_name, display_name, fields, created_at, updated_at FROM compendium_schemas WHERE id=?`, e.SchemaID).
			Scan(&schema.ID, &schema.TypeName, &schema.DisplayName, &fieldsJSON, &created, &updated)
		schema.CreatedAt = created
		schema.UpdatedAt = updated
		json.Unmarshal([]byte(fieldsJSON), &schema.Fields)
	} else {
		// New entry — need schema_id
		schemaID, err := strconv.ParseInt(schemaIDStr, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "schema_id required for new entry")
			return
		}
		var fieldsJSON, created, updated string
		db.DB.QueryRow(`SELECT id, type_name, display_name, fields, created_at, updated_at FROM compendium_schemas WHERE id=?`, schemaID).
			Scan(&schema.ID, &schema.TypeName, &schema.DisplayName, &fieldsJSON, &created, &updated)
		schema.CreatedAt = created
		schema.UpdatedAt = updated
		json.Unmarshal([]byte(fieldsJSON), &schema.Fields)
		data = make(map[string]any)
	}

	renderTemplate(c, "compendium_entry_editor", htmxEntryEditorData{
		Schema: &schema,
		Data:   data,
	})
}

// ─── HTMX: Entry Detail View Partial ───

type htmxEntryDetailData struct {
	Schema *models.CompendiumSchema
	Entry  *models.CompendiumEntry
	Data   map[string]any
}

func HtmxCompendiumEntryDetail(c *gin.Context) {
	entryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid entry id")
		return
	}

	var e models.CompendiumEntry
	var dataJSON, createdAt, updatedAt string
	err = db.DB.QueryRow(`SELECT id, schema_id, data, created_at, updated_at FROM compendium_entries WHERE id=?`, entryID).
		Scan(&e.ID, &e.SchemaID, &dataJSON, &createdAt, &updatedAt)
	if err != nil {
		c.String(http.StatusNotFound, "entry not found")
		return
	}
	e.Data = make(map[string]any)
	json.Unmarshal([]byte(dataJSON), &e.Data)
	e.CreatedAt = createdAt
	e.UpdatedAt = updatedAt

	// Get schema
	var schema models.CompendiumSchema
	var fieldsJSON, created, updated string
	db.DB.QueryRow(`SELECT id, type_name, display_name, fields, created_at, updated_at FROM compendium_schemas WHERE id=?`, e.SchemaID).
		Scan(&schema.ID, &schema.TypeName, &schema.DisplayName, &fieldsJSON, &created, &updated)
	schema.CreatedAt = created
	schema.UpdatedAt = updated
	json.Unmarshal([]byte(fieldsJSON), &schema.Fields)

	renderTemplate(c, "compendium_entry_detail", htmxEntryDetailData{
		Schema: &schema,
		Entry:  &e,
		Data:   e.Data,
	})
}

// ─── HTMX: Duplicate Entry ───

func HtmxCompendiumDuplicateEntry(c *gin.Context) {
	schemaID, err := strconv.ParseInt(c.Param("schemaId"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid schema id")
		return
	}
	entryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid entry id")
		return
	}

	// Fetch original entry
	var dataJSON string
	err = db.DB.QueryRow("SELECT data FROM compendium_entries WHERE id=? AND schema_id=?", entryID, schemaID).Scan(&dataJSON)
	if err != nil {
		c.String(http.StatusNotFound, "entry not found")
		return
	}

	var data map[string]any
	json.Unmarshal([]byte(dataJSON), &data)

	// Append " (copy)" to name field
	if name, ok := data["name"].(string); ok && name != "" {
		data["name"] = name + " (copy)"
	} else if name, ok := data["Name"].(string); ok && name != "" {
		data["Name"] = name + " (copy)"
	}

	newJSON, _ := json.Marshal(data)
	_, err = db.DB.Exec("INSERT INTO compendium_entries(schema_id, data) VALUES(?,?)", schemaID, string(newJSON))
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to duplicate: "+err.Error())
		return
	}

	// Return refreshed table
	HtmxCompendiumEntryTable(c)
}

// ─── HTMX: Check Legacy Migration Status ───

type LegacyMigrationStatus struct {
	NeedsMigration bool   `json:"needs_migration"`
	LegacyCount    int    `json:"legacy_count"`
	MigratedCount  int    `json:"migrated_count"`
	Message        string `json:"message"`
}

func CheckLegacyMigrationStatus(c *gin.Context) {
	// Check if we have the race schema and count its entries
	var raceSchemaID int64
	var raceEntryCount int
	db.DB.QueryRow("SELECT id FROM compendium_schemas WHERE type_name='race'").Scan(&raceSchemaID)
	if raceSchemaID > 0 {
		db.DB.QueryRow("SELECT COUNT(*) FROM compendium_entries WHERE schema_id=?", raceSchemaID).Scan(&raceEntryCount)
	}

	// Check legacy races table
	var legacyCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM compendium_races").Scan(&legacyCount)

	needsMigration := legacyCount > 0 && raceEntryCount == 0
	message := ""
	if needsMigration {
		message = fmt.Sprintf("%d legacy entries found, 0 migrated. Run migration to move legacy data to the new schema system.", legacyCount)
	}

	c.JSON(http.StatusOK, LegacyMigrationStatus{
		NeedsMigration: needsMigration,
		LegacyCount:    legacyCount,
		MigratedCount:  raceEntryCount,
		Message:        message,
	})
}
