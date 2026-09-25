package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

func TestTableScreenShare(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedUser(t, 2, "outsider", "user")
	testutil.SeedCampaign(t, 1, "TableCamp", "The Party", 1)
	testutil.SeedCharacterInCampaign(t, 10, 1, 1, "Aria", "Elf", "Wizard")

	if _, err := db.DB.Exec(`INSERT INTO combat_entries (campaign_id, name, type, initiative_roll, initiative_mod, hp_current, hp_max, ac, is_active, turn_order) VALUES (1, 'Goblin Boss', 'monster', 15, 2, 20, 30, 14, 1, 1)`); err != nil {
		t.Fatalf("seed combat: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO character_conditions (character_id, name, type) VALUES (10, 'Prone', 'condition')`); err != nil {
		t.Fatalf("seed condition: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO campaign_knowledge (campaign_id, title, content, status, shared) VALUES (1, 'The Prophecy', 'Beware the ninth', 'revealed', 1)`); err != nil {
		t.Fatalf("seed knowledge: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO dice_rolls (user_id, character_id, expression, result, total, timestamp) VALUES (1, 10, '1d20', '[18]', 18, '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("seed roll: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/share", CreateShareLink)
		auth.GET("/campaigns/:id/table-state", GetCampaignTableState)
	})
	public := gin.New()
	public.Use(middleware.SecurityHeaders())
	public.GET("/api/share/:token", GetSharedEntity)
	public.GET("/share/:token", GetSharedPage)

	var token string
	t.Run("create table share link", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/share", map[string]any{
			"entity_type": "table", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		token, _ = result["token"].(string)
		if token == "" {
			t.Fatal("expected non-empty token")
		}
		if result["label"] != "TableCamp" {
			t.Fatalf("expected label 'TableCamp', got %v", result["label"])
		}
	})

	t.Run("public JSON returns live table state", func(t *testing.T) {
		w := testutil.Get(t, public, "/api/share/"+token)
		testutil.AssertStatus(t, w, 200)
		body := w.Body.String()
		for _, want := range []string{"TableCamp", "The Party", "Goblin Boss", "Aria", "The Prophecy", "Prone"} {
			if !strings.Contains(body, want) {
				t.Fatalf("table state missing %q: %s", want, body)
			}
		}
	})

	t.Run("public HTML renders the table page", func(t *testing.T) {
		w := testutil.Get(t, public, "/share/"+token)
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), token) {
			t.Fatal("table page should carry its polling token")
		}
	})

	t.Run("authenticated table-state for member", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/table-state")
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), "Goblin Boss") {
			t.Fatal("authenticated table-state missing combatant")
		}
	})

	t.Run("non-member cannot read table-state", func(t *testing.T) {
		stranger := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/campaigns/:id/table-state", GetCampaignTableState)
		}, 2, "user")
		w := testutil.Get(t, stranger, "/api/campaigns/1/table-state")
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("non-member cannot mint a table link", func(t *testing.T) {
		stranger := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.POST("/share", CreateShareLink)
		}, 2, "user")
		w := testutil.PostJSON(t, stranger, "/api/share", map[string]any{
			"entity_type": "table", "entity_id": 1,
		})
		testutil.AssertStatus(t, w, 403)
	})
}
