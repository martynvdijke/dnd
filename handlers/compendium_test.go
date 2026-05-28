package handlers

import (
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
