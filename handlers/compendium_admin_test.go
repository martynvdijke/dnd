package handlers

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

// formatInt formats a float64 as an integer string (no decimals)
func formatInt(v float64) string {
	return strconv.FormatInt(int64(v), 10)
}

func TestCompendiumAdminSchemaCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas", CreateCompendiumSchema)
		auth.GET("/admin/compendium-schemas/:id", GetCompendiumSchema)
		auth.PUT("/admin/compendium-schemas/:id", UpdateCompendiumSchema)
		auth.DELETE("/admin/compendium-schemas/:id", DeleteCompendiumSchema)
	})

	t.Run("create schema", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-schemas", map[string]any{
			"type_name":    "test_schema",
			"display_name": "Test Schema",
			"fields": []map[string]any{
				{"name": "name", "label": "Name", "type": "string", "required": true, "sortable": true, "searchable": true},
				{"name": "description", "label": "Description", "type": "text", "searchable": true},
				{"name": "number", "label": "Number", "type": "integer", "sortable": true},
			},
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if _, ok := result["id"]; !ok {
			t.Fatal("create response missing id")
		}
	})

	t.Run("duplicate type_name returns 409", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-schemas", map[string]any{
			"type_name":    "test_schema",
			"display_name": "Duplicate",
			"fields":       []map[string]any{{"name": "name", "label": "Name", "type": "string"}},
		})
		testutil.AssertStatus(t, w, 409)
	})

	t.Run("list schemas includes built-in and custom", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		testutil.AssertStatus(t, w, 200)
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		if len(schemas) < 8 {
			t.Fatalf("expected at least 8 schemas (7 built-in + 1 custom), got %d", len(schemas))
		}
		found := false
		for _, s := range schemas {
			if s["type_name"] == "test_schema" {
				found = true
				if s["display_name"] != "Test Schema" {
					t.Fatalf("expected display_name 'Test Schema', got %v", s["display_name"])
				}
				break
			}
		}
		if !found {
			t.Fatal("test_schema not found in list")
		}
	})

	t.Run("get schema by id", func(t *testing.T) {
		// Fetch test_schema id from list
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		var schemaID float64
		for _, s := range schemas {
			if s["type_name"] == "test_schema" {
				schemaID = s["id"].(float64)
				break
			}
		}
		if schemaID == 0 {
			t.Fatal("test_schema not found")
		}

		w2 := testutil.Get(t, r, "/api/admin/compendium-schemas/"+formatInt(schemaID))
		testutil.AssertStatus(t, w2, 200)
		var schema map[string]any
		testutil.ParseJSON(t, w2, &schema)
		if schema["type_name"] != "test_schema" {
			t.Fatalf("expected type_name test_schema, got %v", schema["type_name"])
		}
	})

	t.Run("update schema", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		var schemaID float64
		for _, s := range schemas {
			if s["type_name"] == "test_schema" {
				schemaID = s["id"].(float64)
				break
			}
		}
		if schemaID == 0 {
			t.Fatal("test_schema not found")
		}

		w2 := testutil.PutJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(schemaID), map[string]any{
			"display_name": "Updated Schema",
			"fields": []map[string]any{
				{"name": "name", "label": "Name", "type": "string", "required": true},
				{"name": "new_field", "label": "New Field", "type": "text"},
			},
		})
		testutil.AssertStatus(t, w2, 200)

		// Verify
		w3 := testutil.Get(t, r, "/api/admin/compendium-schemas/"+formatInt(schemaID))
		var schema map[string]any
		testutil.ParseJSON(t, w3, &schema)
		if schema["display_name"] != "Updated Schema" {
			t.Fatalf("expected Updated Schema, got %v", schema["display_name"])
		}
	})

	t.Run("delete schema", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		var schemaID float64
		for _, s := range schemas {
			if s["type_name"] == "test_schema" {
				schemaID = s["id"].(float64)
				break
			}
		}
		if schemaID == 0 {
			t.Fatal("test_schema not found")
		}

		w2 := testutil.Delete(t, r, "/api/admin/compendium-schemas/"+formatInt(schemaID))
		testutil.AssertStatus(t, w2, 200)

		// Verify deletion
		w3 := testutil.Get(t, r, "/api/admin/compendium-schemas/"+formatInt(schemaID))
		testutil.AssertStatus(t, w3, 404)
	})
}

func TestCompendiumAdminEntryCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.GET("/admin/compendium-schemas/:id/entries", ListCompendiumEntries)
		auth.POST("/admin/compendium-schemas/:id/entries", CreateCompendiumEntry)
		auth.GET("/admin/compendium-entries/:id", GetCompendiumEntry)
		auth.PUT("/admin/compendium-entries/:id", UpdateCompendiumEntry)
		auth.DELETE("/admin/compendium-entries/:id", DeleteCompendiumEntry)
	})

	// Get race schema id
	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
		if raceID == 0 {
			t.Fatal("race schema not found from seed")
		}
	}

	var entryID float64

	t.Run("create entry", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries", map[string]any{
			"data": map[string]any{
				"name":        "Test Race",
				"description": "A test race for testing",
				"speed":       30,
				"size":        "Medium",
				"system":      "dnd5e",
				"source":      "srd",
			},
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		id, ok := result["id"].(float64)
		if !ok {
			t.Fatal("create entry response missing id")
		}
		entryID = id
	})

	t.Run("list entries", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries")
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		entries, ok := result["entries"].([]any)
		if !ok {
			t.Fatal("expected entries array in response")
		}
		if len(entries) < 1 {
			t.Fatal("expected at least 1 entry")
		}
	})

	t.Run("get entry", func(t *testing.T) {
		if entryID == 0 {
			t.Skip("no entry id from previous test")
		}
		w := testutil.Get(t, r, "/api/admin/compendium-entries/"+formatInt(entryID))
		testutil.AssertStatus(t, w, 200)
		var entry map[string]any
		testutil.ParseJSON(t, w, &entry)
		data, ok := entry["data"].(map[string]any)
		if !ok {
			t.Fatal("response missing data field")
		}
		if data["name"] != "Test Race" {
			t.Fatalf("expected name 'Test Race', got %v", data["name"])
		}
	})

	t.Run("update entry", func(t *testing.T) {
		if entryID == 0 {
			t.Skip("no entry id from previous test")
		}
		w := testutil.PutJSON(t, r, "/api/admin/compendium-entries/"+formatInt(entryID), map[string]any{
			"data": map[string]any{
				"name":        "Updated Race",
				"description": "Updated description",
				"speed":       35,
				"size":        "Large",
			},
		})
		testutil.AssertStatus(t, w, 200)

		// Verify
		w2 := testutil.Get(t, r, "/api/admin/compendium-entries/"+formatInt(entryID))
		var entry map[string]any
		testutil.ParseJSON(t, w2, &entry)
		data := entry["data"].(map[string]any)
		if data["name"] != "Updated Race" {
			t.Fatalf("expected 'Updated Race', got '%v'", data["name"])
		}
	})

	t.Run("delete entry", func(t *testing.T) {
		if entryID == 0 {
			t.Skip("no entry id from previous test")
		}
		w := testutil.Delete(t, r, "/api/admin/compendium-entries/"+formatInt(entryID))
		testutil.AssertStatus(t, w, 200)

		// Verify deletion
		w2 := testutil.Get(t, r, "/api/admin/compendium-entries/"+formatInt(entryID))
		testutil.AssertStatus(t, w2, 404)
	})
}

func TestCompendiumAdminSearch(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas/:id/entries", CreateCompendiumEntry)
		auth.GET("/admin/compendium-search", SearchCompendiumEntries)
	})

	// Get race schema id
	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
	}

	// Create an entry with searchable content
	testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries", map[string]any{
		"data": map[string]any{
			"name":        "Dragonborn",
			"description": "Proud dragon-kin with elemental breath weapons and draconic resilience",
			"speed":       30,
			"size":        "Medium",
		},
	})

	t.Run("search finds matching entry", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-search?q=dragonborn")
		testutil.AssertStatus(t, w, 200)
		var results []map[string]any
		testutil.ParseJSON(t, w, &results)
		if len(results) < 1 {
			t.Fatal("expected at least 1 search result")
		}
		found := false
		for _, r := range results {
			if r["name"] == "Dragonborn" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Dragonborn not found in search results")
		}
	})

	t.Run("search no results returns empty array", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-search?q=xyznonexistent")
		testutil.AssertStatus(t, w, 200)
		var results []map[string]any
		testutil.ParseJSON(t, w, &results)
		if len(results) != 0 {
			t.Fatalf("expected empty array, got %d results", len(results))
		}
	})

	t.Run("search without query returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-search")
		testutil.AssertStatus(t, w, 400)
	})
}

func TestCompendiumAdminImport(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas/:id/import/with-mapping", ImportCompendiumEntriesWithMapping)
	})

	// Get race schema id
	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
	}

	t.Run("import entries with field mapping", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/import/with-mapping", map[string]any{
			"entries": []map[string]any{
				{
					"name":        "Elf",
					"description": "Graceful and long-lived fey beings",
					"speed":       30,
					"size":        "Medium",
				},
				{
					"name":        "Dwarf",
					"description": "Sturdy mountain dwellers",
					"speed":       25,
					"size":        "Medium",
				},
			},
			"field_mapping": []map[string]any{
				{"source": "name", "target": "name"},
				{"source": "description", "target": "description"},
				{"source": "speed", "target": "speed"},
				{"source": "size", "target": "size"},
			},
			"duplicate_action": "skip",
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		inserted, ok := result["inserted"].(float64)
		if !ok || inserted < 1 {
			t.Fatalf("expected at least 1 inserted, got %v", result)
		}
	})

	t.Run("import with duplicate skip does not duplicate", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/import/with-mapping", map[string]any{
			"entries": []map[string]any{
				{
					"name":        "Elf",
					"description": "Another elf",
					"speed":       30,
					"size":        "Medium",
				},
			},
			"field_mapping": []map[string]any{
				{"source": "name", "target": "name"},
			},
			"duplicate_action": "skip",
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		skipped, ok := result["skipped"].(float64)
		if !ok || skipped < 1 {
			t.Fatalf("expected at least 1 skipped for duplicate, got %v", result)
		}
	})
}

func TestCompendiumAdminExport(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas/:id/entries", CreateCompendiumEntry)
		auth.GET("/admin/compendium-schemas/:id/export", ExportCompendiumEntries)
	})

	// Get race schema id and create an entry
	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
	}

	testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries", map[string]any{
		"data": map[string]any{
			"name": "Exportable Race",
			"description": "A race for export testing",
		},
	})

	t.Run("export returns entries with metadata", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/export")
		testutil.AssertStatus(t, w, 200)
		var export map[string]any
		testutil.ParseJSON(t, w, &export)
		if export["schema"] != "race" {
			t.Fatalf("expected schema 'race', got %v", export["schema"])
		}
		entries, ok := export["entries"].([]any)
		if !ok {
			t.Fatal("expected entries array in export")
		}
		if len(entries) < 1 {
			t.Fatal("expected at least 1 entry in export")
		}
		entry := entries[0].(map[string]any)
		if _, ok := entry["id"]; !ok {
			t.Fatal("export entry missing id")
		}
		if _, ok := entry["data"]; !ok {
			t.Fatal("export entry missing data")
		}
	})
}

func TestCompendiumAdminImportLogs(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas/:id/import/with-mapping", ImportCompendiumEntriesWithMapping)
		auth.GET("/admin/compendium-import-logs", ListCompendiumImportLogs)
		auth.POST("/admin/compendium-import-logs/:id/rollback", RollbackCompendiumImport)
	})

	// Get race schema id
	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
	}

	// Perform an import to generate a log entry
	testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/import/with-mapping", map[string]any{
		"entries": []map[string]any{
			{"name": "Halfling", "description": "Small and nimble"},
		},
		"field_mapping": []map[string]any{
			{"source": "name", "target": "name"},
			{"source": "description", "target": "description"},
		},
		"duplicate_action": "skip",
	})

	t.Run("list import logs", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/compendium-import-logs")
		testutil.AssertStatus(t, w, 200)
		var logs []map[string]any
		testutil.ParseJSON(t, w, &logs)
		if len(logs) < 1 {
			t.Fatal("expected at least 1 import log")
		}
		if logs[0]["status"] != "completed" {
			t.Fatalf("expected status 'completed', got %v", logs[0]["status"])
		}
	})

	t.Run("rollback import", func(t *testing.T) {
		// Get log id
		w := testutil.Get(t, r, "/api/admin/compendium-import-logs")
		var logs []map[string]any
		testutil.ParseJSON(t, w, &logs)
		if len(logs) < 1 {
			t.Fatal("no import logs to rollback")
		}
		logID := logs[0]["id"].(float64)

		w2 := testutil.PostJSON(t, r, "/api/admin/compendium-import-logs/"+formatInt(logID)+"/rollback", nil)
		testutil.AssertStatus(t, w2, 200)

		// Verify rollback
		w3 := testutil.Get(t, r, "/api/admin/compendium-import-logs")
		var updatedLogs []map[string]any
		testutil.ParseJSON(t, w3, &updatedLogs)
		found := false
		for _, l := range updatedLogs {
			if l["id"].(float64) == logID {
				if l["status"] != "rolled_back" {
					t.Fatalf("expected status 'rolled_back', got %v", l["status"])
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatal("rolled back log not found")
		}
	})
}

func TestCompendiumAdminDetectFields(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas/:id/import/detect", DetectImportFields)
	})

	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
	}

	t.Run("detect fields from sample data", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/import/detect", map[string]any{
			"source_prefix": "",
			"sample": map[string]any{
				"name":        "Test",
				"description": "Desc",
				"speed":       30,
				"traits":      "Darkvision",
				"properties": map[string]any{
					"nested": "value",
				},
			},
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		suggestions, ok := result["suggestions"].([]any)
		if !ok {
			t.Fatal("expected suggestions array")
		}
		if len(suggestions) < 1 {
			t.Fatal("expected at least 1 suggestion")
		}
		// Check that name got matched
		foundName := false
		for _, s := range suggestions {
			entry := s.(map[string]any)
			if entry["target"] == "name" {
				foundName = true
				break
			}
		}
		if !foundName {
			t.Logf("suggestions: %+v", suggestions)
			t.Fatal("expected name target in suggestions")
		}
	})
}

func TestCompendiumAdminBatchOps(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/compendium-schemas", ListCompendiumSchemas)
		auth.POST("/admin/compendium-schemas/:id/entries", CreateCompendiumEntry)
		auth.GET("/admin/compendium-entries/:id", GetCompendiumEntry)
		auth.POST("/admin/compendium-entries/batch-delete", BatchDeleteCompendiumEntries)
		auth.POST("/admin/compendium-entries/batch-update", BatchUpdateCompendiumEntries)
	})

	// Get race schema id
	var raceID float64
	{
		w := testutil.Get(t, r, "/api/admin/compendium-schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
		if raceID == 0 {
			t.Fatal("race schema not found")
		}
	}

	// Create 2 entries for batch ops
	var id1, id2 float64
	testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries", map[string]any{
		"data": map[string]any{"name": "BatchDel1", "description": "To delete", "speed": 30, "size": "Medium"},
	})
	w1 := testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries", map[string]any{
		"data": map[string]any{"name": "BatchDel2", "description": "Also delete", "speed": 25, "size": "Small"},
	})
	var r1 map[string]any
	testutil.ParseJSON(t, w1, &r1)
	id1 = r1["id"].(float64)

	w2 := testutil.PostJSON(t, r, "/api/admin/compendium-schemas/"+formatInt(raceID)+"/entries", map[string]any{
		"data": map[string]any{"name": "BatchUpd1", "description": "To update", "speed": 30, "size": "Medium"},
	})
	var r2 map[string]any
	testutil.ParseJSON(t, w2, &r2)
	id2 = r2["id"].(float64)

	t.Run("batch delete removes entries", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-entries/batch-delete", map[string]any{
			"ids": []float64{id1},
		})
		testutil.AssertStatus(t, w, 200)
		// Verify deletion
		w2 := testutil.Get(t, r, "/api/admin/compendium-entries/"+formatInt(id1))
		testutil.AssertStatus(t, w2, 404)
	})

	t.Run("batch update modifies entries", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/compendium-entries/batch-update", map[string]any{
			"ids":  []float64{id2},
			"data": map[string]any{"source": "batch_updated"},
		})
		testutil.AssertStatus(t, w, 200)
		// Verify update
		w2 := testutil.Get(t, r, "/api/admin/compendium-entries/"+formatInt(id2))
		var entry map[string]any
		testutil.ParseJSON(t, w2, &entry)
		data := entry["data"].(map[string]any)
		if data["source"] != "batch_updated" {
			t.Fatalf("expected source 'batch_updated', got %v", data["source"])
		}
	})
}
