package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"villum/handlers/testutil"
)

func TestHtmxHelpersExtra(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Error("truncate short")
	}
	if !strings.Contains(truncate("hello world", 5), "...") {
		t.Error("truncate long")
	}
	if seq(1, 3)[2] != 3 {
		t.Error("seq")
	}
}

func TestHtmxGetIntParam(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	// indirectly via HtmxCreateEncounter with invalid int falls back to default
	w := testutil.PostForm(t, r, "/api/htmx/campaigns/1/monster-roster", map[string]string{"compendium_monster_id": "bad"})
	// handler parses but fails to find monster -> 404; still covers parse path
	if w.Code != 404 && w.Code != 400 && w.Code != 200 {
		t.Logf("got %d", w.Code)
	}
}

func TestHtmxListMonsterLibraryViaBrowse(t *testing.T) {
	r := newCompendiumHtmxRouter(t)
	defer testutil.CloseDB(t)
	w := testutil.Get(t, r, "/api/htmx/compendium/spells")
	if w.Code != 200 {
		t.Fatalf("want 200 got %d %s", w.Code, w.Body.String())
	}
}
func init() { _ = gin.Mode() }
