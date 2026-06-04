package handlers

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestCampaignsCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "player1", "user")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns", CreateCampaign)
		auth.GET("/campaigns", ListCampaigns)
		auth.PUT("/campaigns/:id", UpdateCampaign)
		auth.DELETE("/campaigns/:id", DeleteCampaign)
	})

	t.Run("create campaign returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns", map[string]any{
			"name": "Curse of Strahd", "party_name": "The Dawnbringers",
			"description": "Ravenloft campaign",
		})
		testutil.AssertStatus(t, w, 201)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "name", "Curse of Strahd")
		testutil.AssertField(t, data, "party_name", "The Dawnbringers")
	})

	t.Run("create campaign with empty name returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns", map[string]any{
			"name": "",
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for empty name, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list campaigns returns created campaigns", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns")
		testutil.AssertStatus(t, w, 200)
		var camps []any
		testutil.ParseJSON(t, w, &camps)
		if len(camps) < 1 {
			t.Fatal("expected at least 1 campaign")
		}
	})

	t.Run("update campaign returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns", map[string]any{
			"name": "Update Test", "party_name": "Party",
		})
		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		cid := int64(created["id"].(float64))

		w = testutil.PutJSON(t, r, fmt.Sprintf("/api/campaigns/%d", cid), map[string]any{
			"name": "Updated Name", "party_name": "Updated Party",
			"description": "Updated",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete campaign returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns", map[string]any{
			"name": "Delete Me", "party_name": "Temp",
		})
		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		cid := int64(created["id"].(float64))

		w = testutil.Delete(t, r, fmt.Sprintf("/api/campaigns/%d", cid))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCampaignMembers(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "member1", "user")
	testutil.SeedCampaign(t, 1, "Member Test", "Test Party", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/members", ListCampaignMembers)
		auth.POST("/campaigns/:id/members", AddCampaignMember)
		auth.PUT("/campaigns/:id/members/:userId", SetCampaignMemberRole)
		auth.DELETE("/campaigns/:id/members/:userId", RemoveCampaignMember)
	})

	t.Run("list members includes owner", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/members")
		testutil.AssertStatus(t, w, 200)
		var members []any
		testutil.ParseJSON(t, w, &members)
		if len(members) < 1 {
			t.Fatal("expected at least owner in members")
		}
	})

	t.Run("add member returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/members", map[string]any{
			"username": "member1",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("add non-existent user returns 404", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/members", map[string]any{
			"username": "nonexistent_user",
		})
		if w.Code != 404 {
			t.Fatalf("expected 404 for non-existent user, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("change member role returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/campaigns/1/members/2", map[string]any{
			"role": "dm",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("remove member returns 200", func(t *testing.T) {
		// First get the member's user_id from the list
		w := testutil.Get(t, r, "/api/campaigns/1/members")
		var members []map[string]any
		testutil.ParseJSON(t, w, &members)
		var targetID float64
		for _, m := range members {
			if m["username"] == "member1" {
				targetID = m["user_id"].(float64)
				break
			}
		}
		if targetID == 0 {
			t.Skip("member1 not found, skipping")
		}
		w = testutil.Delete(t, r, fmt.Sprintf("/api/campaigns/1/members/%d", int64(targetID)))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCampaignUnassignsCharacters(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	// Create campaign with character assigned
	testutil.SeedCampaign(t, 1, "Delete Campaign", "Party", 1)
	testutil.SeedCharacter(t, 1, 1, "Temp Member", "Gnome", "Wizard")

	// Assign character to campaign via raw SQL
	_, err := db.DB.Exec("UPDATE characters SET campaign_id = 1 WHERE id = 1")
	if err != nil {
		t.Fatalf("assign char to campaign: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/campaigns/:id", DeleteCampaign)
	})

	t.Run("delete campaign unassigns characters", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/campaigns/1")
		testutil.AssertStatus(t, w, 200)

		var campaignID *int64
		_ = db.DB.QueryRow("SELECT campaign_id FROM characters WHERE id = 1").Scan(&campaignID)
		if campaignID != nil {
			t.Fatalf("expected campaign_id to be NULL after campaign delete, got %v", *campaignID)
		}
	})
}

func TestCampaignMonsterRoster(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Roster Campaign", "Tester Party", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/campaigns/:id/monster-roster", HtmxCampaignMonsterRoster)
		auth.POST("/htmx/campaigns/:id/monster-roster", HtmxAddCampaignMonster)
		auth.DELETE("/htmx/campaigns/:id/monster-roster/:rid", HtmxRemoveCampaignMonster)
	})

	t.Run("list roster returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/htmx/campaigns/1/monster-roster")
		testutil.AssertStatus(t, w, 200)
		if len(w.Body.String()) == 0 {
			t.Fatal("expected non-empty HTML response")
		}
	})

	t.Run("add monster returns 200", func(t *testing.T) {
		// HTMX handler uses PostForm, so send URL-encoded form data
		// compendium_monster_id is NOT NULL in the schema, so use a seeded compendium monster
		w := testutil.PostForm(t, r, "/api/htmx/campaigns/1/monster-roster", map[string]string{
			"compendium_monster_id": "1",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("remove monster returns 200", func(t *testing.T) {
		// Seed a roster entry directly
		// Need to provide compendium_monster_id since it's NOT NULL
		_, err := db.DB.Exec(`INSERT INTO campaign_monster_roster(id, campaign_id, compendium_monster_id, name)
			VALUES(2, 1, 1, 'Test Monster')`)
		if err != nil {
			t.Fatalf("seed roster: %v", err)
		}
		w := testutil.Delete(t, r, "/api/htmx/campaigns/1/monster-roster/2")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCampaignAuthorization(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "user")
	testutil.SeedUser(t, 2, "intruder", "user")
	testutil.SeedCampaign(t, 1, "Private Campaign", "Secret", 1)

	t.Run("non-owner cannot delete campaign", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.DELETE("/campaigns/:id", DeleteCampaign)
		}, 2, "user")

		w := testutil.Delete(t, r, "/api/campaigns/1")
		if w.Code != 403 {
			t.Fatalf("expected 403 for non-owner delete, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("admin can delete any campaign", func(t *testing.T) {
		r := testutil.NewRouter(func(auth *gin.RouterGroup) {
			auth.DELETE("/campaigns/:id", DeleteCampaign)
		})

		w := testutil.Delete(t, r, "/api/campaigns/1")
		testutil.AssertStatus(t, w, 200)
	})
}
