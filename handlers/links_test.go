package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// Seed users.
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")

	// Seed entities owned by user 1.
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(1, 1, 'Grimble', 'Gnome', 'Tinker', 'A clever gnome inventor')`); err != nil {
		t.Fatalf("seed npc: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO locations(id, user_id, name, type, description)
		VALUES(1, 1, 'Emberhold Tavern', 'tavern', 'A cozy tavern')`); err != nil {
		t.Fatalf("seed location: %v", err)
	}

	// Seed an NPC owned by user 2 for permission tests.
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(2, 2, 'Merlin', 'Human', 'Wizard', 'A friendly wizard')`); err != nil {
		t.Fatalf("seed npc 2: %v", err)
	}
	// NPC 3 for delete test (avoid test isolation issues).
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(3, 1, 'DeleteTest', 'Human', 'Fighter', 'For delete test')`); err != nil {
		t.Fatalf("seed npc 3: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		RegisterLinkRoutes(auth)
	})

	t.Run("create link succeeds", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/links", map[string]any{
			"source_type": "npc",
			"source_id":   1,
			"target_type": "location",
			"target_id":   1,
			"context":     "manual",
		})
		testutil.AssertStatus(t, w, 201)
		var resp LinkResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if resp.ID == 0 {
			t.Errorf("expected non-zero link id")
		}
		if resp.SourceType != "npc" || resp.TargetType != "location" {
			t.Errorf("unexpected types: source=%s target=%s", resp.SourceType, resp.TargetType)
		}
	})

	t.Run("duplicate link returns 409", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/links", map[string]any{
			"source_type": "npc",
			"source_id":   1,
			"target_type": "location",
			"target_id":   1,
			"context":     "manual",
		})
		testutil.AssertStatus(t, w, 409)
	})

	t.Run("create link unauthorized returns 403", func(t *testing.T) {
		r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			RegisterLinkRoutes(auth)
		}, 2, "user")
		// User 2 tries to create a link whose source is user 1's NPC.
		w := testutil.PostJSON(t, r2, "/api/links", map[string]any{
			"source_type": "npc",
			"source_id":   1,
			"target_type": "location",
			"target_id":   1,
			"context":     "manual",
		})
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("get links returns outgoing and backlinks", func(t *testing.T) {
		// Create a backlink: location → npc (so npc has a backlink)
		w2 := testutil.PostJSON(t, r, "/api/links", map[string]any{
			"source_type": "location",
			"source_id":   1,
			"target_type": "npc",
			"target_id":   1,
			"context":     "manual",
		})
		testutil.AssertStatus(t, w2, 201)

		w := testutil.Get(t, r, "/api/links/npc/1")
		testutil.AssertStatus(t, w, 200)
		var resp struct {
			Outgoing  []LinkResponse `json:"outgoing"`
			Backlinks []LinkResponse `json:"backlinks"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if len(resp.Outgoing) == 0 {
			t.Errorf("expected outgoing links, got none")
		}
		if len(resp.Backlinks) == 0 {
			t.Errorf("expected backlinks, got none")
		}
		// Clean up second link for test isolation.
		for _, l := range resp.Backlinks {
			if l.SourceType == "location" && l.SourceID == 1 {
				w3 := testutil.Delete(t, r, "/api/links/"+jsonNumber(l.ID))
				testutil.AssertStatus(t, w3, 200)
				break
			}
		}
	})

	t.Run("delete link succeeds", func(t *testing.T) {
		id, _ := db.DB.Exec(`INSERT INTO entity_links(source_type, source_id, target_type, target_id, context) VALUES('npc', 3, 'location', 1, 'manual')`)
		linkID, _ := id.LastInsertId()

		del := testutil.Delete(t, r, "/api/links/"+strconv.FormatInt(linkID, 10))
		testutil.AssertStatus(t, del, 200)

		var count int
		db.DB.QueryRow(`SELECT COUNT(*) FROM entity_links WHERE id=?`, linkID).Scan(&count)
		if count != 0 {
			t.Errorf("expected link %d to be deleted", linkID)
		}
	})

	t.Run("delete non-existent link returns 404", func(t *testing.T) {
		del := testutil.Delete(t, r, "/api/links/99999")
		testutil.AssertStatus(t, del, 404)
	})
}

func TestReconcileMentionLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(1, 1, 'Grimble', 'Gnome', 'Tinker', 'A clever gnome inventor')`); err != nil {
		t.Fatalf("seed npc: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO locations(id, user_id, name, type, description)
		VALUES(1, 1, 'Emberhold Tavern', 'tavern', 'A cozy tavern')`); err != nil {
		t.Fatalf("seed location: %v", err)
	}

	t.Run("adds new mention links", func(t *testing.T) {
		err := ReconcileMentionLinks("note", 1, []MentionRef{
			{EntityType: "npc", EntityID: 1},
			{EntityType: "location", EntityID: 1},
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}

		var count int
		db.DB.QueryRow(`SELECT COUNT(*) FROM entity_links WHERE source_type='note' AND source_id=1 AND context='mention'`).Scan(&count)
		if count != 2 {
			t.Errorf("expected 2 mention links, got %d", count)
		}
	})

	t.Run("removes stale mention links", func(t *testing.T) {
		// Start with 2 mention links, reconcile with only 1.
		err := ReconcileMentionLinks("note", 1, []MentionRef{
			{EntityType: "npc", EntityID: 1},
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}

		var count int
		db.DB.QueryRow(`SELECT COUNT(*) FROM entity_links WHERE source_type='note' AND source_id=1 AND context='mention'`).Scan(&count)
		if count != 1 {
			t.Errorf("expected 1 mention link after removal, got %d", count)
		}

		var targetType string
		db.DB.QueryRow(`SELECT target_type FROM entity_links WHERE source_type='note' AND source_id=1 AND context='mention'`).Scan(&targetType)
		if targetType != "npc" {
			t.Errorf("expected remaining mention to be npc, got %s", targetType)
		}
	})

	t.Run("empty mentions removes all", func(t *testing.T) {
		err := ReconcileMentionLinks("note", 1, []MentionRef{})
		if err != nil {
			t.Fatalf("reconcile empty: %v", err)
		}
		var count int
		db.DB.QueryRow(`SELECT COUNT(*) FROM entity_links WHERE source_type='note' AND source_id=1 AND context='mention'`).Scan(&count)
		if count != 0 {
			t.Errorf("expected 0 mention links after empty reconcile, got %d", count)
		}
	})
}

// jsonNumber returns an int64 as a decimal string for use in URL paths.
func jsonNumber(id int64) string {
	return fmt.Sprint(id)
}
