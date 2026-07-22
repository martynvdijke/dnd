package handlers

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestSearchV2(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Gandalf", "Maia", "Wizard")
	testutil.SeedUser(t, 2, "other", "user")

	// Also seed an NPC
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(1, 1, 'Grimble', 'Gnome', 'Tinker', 'A clever gnome inventor')`); err != nil {
		t.Fatalf("seed npc: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/search", HandleSearch)
	})

	t.Run("basic search returns results", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=Gandalf")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Results []UnifiedResult `json:"results"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if len(resp.Results) == 0 {
			t.Fatalf("expected results, got empty: body=%s", w.Body.String())
		}
		found := false
		for _, r := range resp.Results {
			if r.EntityType == "character" && r.EntityID == 1 {
				found = true
				if r.Title != "Gandalf" {
					t.Errorf("expected title 'Gandalf', got %q", r.Title)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected to find character/1 in results")
		}
	})

	t.Run("empty query returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("partial match via FTS5 prefix", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=Grim")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Results []UnifiedResult `json:"results"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if len(resp.Results) == 0 {
			t.Fatalf("expected results for 'Grim', got none: body=%s", w.Body.String())
		}
	})

	t.Run("type filter narrows results", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=G&types=npc")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Results []UnifiedResult `json:"results"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		for _, r := range resp.Results {
			if r.EntityType != "npc" {
				t.Errorf("expected only npc results with types=npc, got %q", r.EntityType)
			}
		}
	})

	t.Run("no match returns empty results", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=Zxqwooper")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Results []UnifiedResult `json:"results"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if len(resp.Results) != 0 {
			t.Errorf("expected empty results, got %d", len(resp.Results))
		}
	})

	t.Run("legacy fields present in response", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=Gandalf")
		testutil.AssertStatus(t, w, 200)
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if _, ok := resp["characters"]; !ok {
			t.Errorf("expected 'characters' legacy field")
		}
		if _, ok := resp["results"]; !ok {
			t.Errorf("expected 'results' field")
		}
	})

	t.Run("permission scoping: other user sees own content only", func(t *testing.T) {
		r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/search", HandleSearch)
		}, 2, "user")
		w := testutil.Get(t, r2, "/api/search?q=Gandalf")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Results []UnifiedResult `json:"results"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		for _, r := range resp.Results {
			if r.EntityType == "character" {
				t.Errorf("user 2 should not see user 1's character: %+v", r)
			}
		}
	})

	t.Run("admin sees everything", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/search?q=Gandalf")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Results []UnifiedResult `json:"results"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		// Admin should see character/1
		found := false
		for _, r := range resp.Results {
			if r.EntityType == "character" && r.EntityID == 1 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("admin should see all entities, character/1 not found")
		}
	})
}
