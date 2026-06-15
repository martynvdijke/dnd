package handlers

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

func TestCompendiumRaces(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/races", ListCompendiumRaces)
		auth.GET("/compendium/classes", ListCompendiumClasses)
		auth.GET("/compendium/spells", ListCompendiumSpells)
		auth.GET("/compendium/feats", ListCompendiumFeats)
		auth.GET("/compendium/backgrounds", ListCompendiumBackgrounds)
		auth.GET("/compendium/equipment", ListCompendiumEquipment)
		auth.GET("/compendium/search", SearchCompendium)
	})

	t.Run("list races returns 200 with system/source", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/races")
		testutil.AssertStatus(t, w, 200)
		var races []map[string]any
		testutil.ParseJSON(t, w, &races)
		if len(races) > 0 {
			if _, ok := races[0]["system"]; !ok {
				t.Fatal("response missing system field")
			}
			if _, ok := races[0]["source"]; !ok {
				t.Fatal("response missing source field")
			}
		}
	})

	t.Run("list classes returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/classes")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("list spells returns 200 with at least 200 entries", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/spells")
		testutil.AssertStatus(t, w, 200)
		var spells []any
		testutil.ParseJSON(t, w, &spells)
		if len(spells) < 200 {
			t.Fatalf("expected >=200 spells, got %d", len(spells))
		}
	})

	t.Run("list feats returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/feats")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("list backgrounds returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/backgrounds")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("list equipment returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/equipment")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCompendiumSearch(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/search", SearchCompendium)
	})

	t.Run("search with query returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search?q=fire")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("search without query returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search")
		if w.Code != 400 {
			t.Fatalf("expected 400 for empty query, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("search with empty string returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search?q=")
		if w.Code != 400 {
			t.Fatalf("expected 400 for empty q, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCompendiumFilters(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/spells", ListCompendiumSpells)
		auth.GET("/compendium/equipment", ListCompendiumEquipment)
	})

	t.Run("filter spells by level returns correct level", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/spells?level=0")
		testutil.AssertStatus(t, w, 200)
		var spells []map[string]any
		testutil.ParseJSON(t, w, &spells)
		for _, s := range spells {
			if s["level"].(float64) != 0 {
				t.Fatalf("expected all cantrips, got level %v: %v", s["level"], s["name"])
			}
		}
		if len(spells) < 5 {
			t.Fatalf("expected >=5 cantrips, got %d", len(spells))
		}
	})

	t.Run("filter spells by school returns correct school", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/spells?school=Evocation")
		testutil.AssertStatus(t, w, 200)
		var spells []map[string]any
		testutil.ParseJSON(t, w, &spells)
		for _, s := range spells {
			if school, ok := s["school"].(string); ok && school != "Evocation" {
				t.Fatalf("expected school Evocation, got %s: %v", school, s["name"])
			}
		}
	})

	t.Run("filter equipment by category returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/equipment?category=Weapon")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestPlayerCompendiumAccess(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	SeedCompendiumSchemas()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/schemas", ListCompendiumSchemas)
		auth.GET("/compendium/schemas/:id/entries", ListCompendiumEntries)
		auth.GET("/compendium/schemas/:id/entries/:eid", GetCompendiumEntry)
		auth.GET("/compendium/search", SearchCompendium)
	})

	t.Run("list schemas returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/schemas")
		testutil.AssertStatus(t, w, 200)
		var schemas []any
		testutil.ParseJSON(t, w, &schemas)
		if len(schemas) < 7 {
			t.Fatalf("expected at least 7 schemas, got %d", len(schemas))
		}
	})

	t.Run("list entries for race schema returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/schemas")
		var schemas []map[string]any
		testutil.ParseJSON(t, w, &schemas)
		var raceID float64
		for _, s := range schemas {
			if s["type_name"] == "race" {
				raceID = s["id"].(float64)
				break
			}
		}
		if raceID == 0 {
			t.Fatal("race schema not found")
		}
		w2 := testutil.Get(t, r, "/api/compendium/schemas/"+formatInt(raceID)+"/entries")
		testutil.AssertStatus(t, w2, 200)
	})

	t.Run("search returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search?q=fire")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestHandleEnabledAIEndpoints(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/ai/endpoints", HandleListEnabledAIEndpoints)
	})

	t.Run("list text endpoints returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/ai/endpoints?type=text")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("list image endpoints returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/ai/endpoints?type=image")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("list without type returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/ai/endpoints")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("list with invalid type returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/ai/endpoints?type=video")
		testutil.AssertStatus(t, w, 400)
	})
}

func TestListUserCompendiumEntriesBySchema(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "user", "user")

	// Seed a compendium schema with some entries
	_, err := db.DB.Exec(`INSERT INTO compendium_schemas(id, type_name, display_name, fields)
		VALUES(100, 'magic-item', 'Magic Item', '[{"name":"name","label":"Name","type":"string","required":true}]')`)
	if err != nil {
		t.Fatalf("seed schema: %v", err)
	}
	_, err = db.DB.Exec(`INSERT INTO compendium_entries(id, schema_id, data) VALUES(1, 100, '{"name":"Staff of Power"}')`)
	if err != nil {
		t.Fatalf("seed entry 1: %v", err)
	}
	_, err = db.DB.Exec(`INSERT INTO compendium_entries(id, schema_id, data) VALUES(2, 100, '{"name":"Cloak of Invisibility"}')`)
	if err != nil {
		t.Fatalf("seed entry 2: %v", err)
	}

	// Seed an empty schema (should be excluded from response)
	_, err = db.DB.Exec(`INSERT INTO compendium_schemas(id, type_name, display_name, fields)
		VALUES(101, 'empty-type', 'Empty Type', '[{"name":"name","label":"Name","type":"string","required":true}]')`)
	if err != nil {
		t.Fatalf("seed empty schema: %v", err)
	}

	t.Run("authenticated user sees schema entries", func(t *testing.T) {
		r := testutil.NewRouter(func(auth *gin.RouterGroup) {
			auth.GET("/compendium/entries-by-schema", ListUserCompendiumEntriesBySchema)
		})
		w := testutil.Get(t, r, "/api/compendium/entries-by-schema")
		testutil.AssertStatus(t, w, 200)

		var resp struct {
			Schemas []map[string]any `json:"schemas"`
		}
		testutil.ParseJSON(t, w, &resp)
		if len(resp.Schemas) != 1 {
			t.Fatalf("expected 1 schema (empty excluded), got %d", len(resp.Schemas))
		}
		if resp.Schemas[0]["type_name"] != "magic-item" {
			t.Fatalf("expected type_name 'magic-item', got %v", resp.Schemas[0]["type_name"])
		}
		entries := resp.Schemas[0]["entries"].([]any)
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
	})

	t.Run("unauthorized request returns 401", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.SecurityHeaders())
		noAuth := r.Group("/api")
		noAuth.Use(middleware.AuthRequired())
		noAuth.GET("/compendium/entries-by-schema", ListUserCompendiumEntriesBySchema)
		w := testutil.Get(t, r, "/api/compendium/entries-by-schema")
		if w.Code != 401 {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

func FuzzCompendiumSearch(f *testing.F) {
	f.Add("fire")
	f.Add("heal")
	f.Add("")
	f.Add("a")
	f.Add("' OR '1'='1")
	f.Fuzz(func(t *testing.T, q string) {
		testutil.NewDB(t)
		defer testutil.CloseDB(t)
		testutil.SeedUser(t, 1, "admin", "admin")
		r := testutil.NewRouter(func(auth *gin.RouterGroup) {
			auth.GET("/compendium/search", SearchCompendium)
		})
		req := httptest.NewRequest("GET", "/api/compendium/search", nil)
		urlQ := url.QueryEscape(q)
		req.URL.RawQuery = "q=" + urlQ
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		_ = w.Code
	})
}
