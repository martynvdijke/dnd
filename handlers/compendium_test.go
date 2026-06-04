package handlers

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
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
