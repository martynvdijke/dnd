package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestFactionsCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/factions", CreateFaction)
		auth.GET("/factions", ListFactions)
		auth.PUT("/factions/:id", UpdateFaction)
		auth.DELETE("/factions/:id", DeleteFaction)
	})

	t.Run("create faction returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/factions", map[string]any{
			"name": "Harpers", "type": "organization",
			"description": "Secret society",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list factions returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/factions")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("update faction returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/factions/1", map[string]any{
			"name": "Harpers Updated", "type": "organization",
			"description": "Updated description",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete faction returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/factions/1")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestFactionReputation(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "RepHero", "Human", "Rogue")
	testutil.SeedCampaign(t, 1, "RepCamp", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/factions", CreateFaction)
		auth.POST("/faction-reputation", SetFactionReputation)
		auth.GET("/faction-reputation", GetFactionReputations)
		auth.DELETE("/faction-reputation/:id", DeleteFactionReputation)
	})

	t.Run("set and get reputation returns correct values", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/factions", map[string]any{
			"name": "Zhentarim", "type": "guild",
		})
		testutil.AssertStatus(t, w, 201)

		w = testutil.PostJSON(t, r, "/api/faction-reputation", map[string]any{
			"character_id": 1, "faction_id": 1, "standing": 50,
			"rank": "Member", "notes": "Joined recently",
		})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Get(t, r, "/api/faction-reputation?character_id=1")
		testutil.AssertStatus(t, w, 200)
		var reps []map[string]any
		testutil.ParseJSON(t, w, &reps)
		if len(reps) < 1 {
			t.Fatal("expected at least 1 reputation entry")
		}
		if reps[0]["standing"].(float64) != 50 {
			t.Fatalf("expected standing 50, got %v", reps[0]["standing"])
		}
	})
}
