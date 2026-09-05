package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func newCompendiumHtmxRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCompendiumMonster(t, 1, "Goblin")
	testutil.SeedCompendiumMonster(t, 2, "Orc")
	testutil.SeedCompendiumSpell(t, 1, "Fireball")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		HtmxRegisterRoutes(auth)
	})
	return r
}

func TestHtmxCompendiumMonsterSearch(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	t.Run("search all", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/search")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
			t.Errorf("want html ct")
		}
	})
	t.Run("search with query", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/search?q=Goblin")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Goblin") {
			t.Logf("body: %s", w.Body.String())
		}
	})
	t.Run("search with cr", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/search?cr=1")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d", w.Code)
		}
	})
}

func TestHtmxCompendiumMonsterDetail(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	t.Run("found", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/1")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d %s", w.Code, w.Body.String())
		}
	})
	t.Run("not found", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/9999")
		if w.Code != 404 {
			t.Fatalf("want 404 got %d", w.Code)
		}
	})
}

func TestHtmxAPIImportModalAndSearch(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	t.Run("modal", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/import-modal")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d", w.Code)
		}
	})
	t.Run("api search empty query", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium-monsters/api-search")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d", w.Code)
		}
	})
}

func TestHtmxImportAPIMonsterValidation(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	t.Run("missing url", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/htmx/compendium-monsters/import", map[string]string{"name": "Goblin"})
		if w.Code != 400 {
			t.Fatalf("want 400 got %d %s", w.Code, w.Body.String())
		}
	})
}

func TestHtmxCompendiumBrowserAndPickers(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	paths := []string{
		"/api/htmx/compendium-monsters",
		"/api/htmx/compendium-monsters/picker/1?campaign_id=1",
		"/api/htmx/compendium-monsters/oneshot/1",
	}
	for _, p := range paths {
		w := testutil.Get(t, r, p)
		if w.Code != 200 {
			t.Errorf("path %s want 200 got %d", p, w.Code)
		}
	}
}

func TestHtmxCompendiumSpellDetailAndModal(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	t.Run("spell detail found", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium/spells/1/detail")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d %s", w.Code, w.Body.String())
		}
	})
	t.Run("spell detail not found", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium/spells/9999/detail")
		if w.Code != 404 {
			t.Fatalf("want 404 got %d", w.Code)
		}
	})
	t.Run("spell modal found", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium/spells/1/modal")
		if w.Code != 200 {
			t.Fatalf("want 200 got %d", w.Code)
		}
	})
	t.Run("spell detail bad id", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/compendium/spells/abc/detail")
		if w.Code != 400 {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
}

func TestHtmxHelpers(t *testing.T) {
	if got := float64ToCR(0); got != "0" {
		t.Errorf("0")
	}
	if got := float64ToCR(0.5); got != "1/2" {
		t.Errorf("0.5 got %s", got)
	}
	if got := jsonArrayToString(nil); got != "" {
		t.Errorf("nil")
	}
	if got := getStrFromMap(map[string]any{"x": "y"}, "x", "def"); got != "y" {
		t.Errorf("getStr")
	}
	if got := getStrFromMap(map[string]any{}, "x", "def"); got != "def" {
		t.Errorf("def")
	}
}
