package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestEncountersCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/encounters", CreateEncounter)
		auth.GET("/encounters", ListEncounters)
		auth.GET("/encounters/:id", GetEncounter)
		auth.PUT("/encounters/:id", UpdateEncounter)
		auth.DELETE("/encounters/:id", DeleteEncounter)
		auth.POST("/encounters/calculate-xp", CalculateEncounterXP)
		auth.POST("/encounters/:id/monsters", AddEncounterMonster)
		auth.PUT("/encounter-monsters/:mid", UpdateEncounterMonster)
		auth.DELETE("/encounter-monsters/:mid", DeleteEncounterMonster)
		auth.GET("/monster-xp", GetMonsterXP)
	})

	t.Run("create encounter returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/encounters", map[string]any{
			"name": "Goblin Ambush", "difficulty": "medium",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list encounters returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/encounters")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get monster xp table returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/monster-xp")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("calculate xp returns difficulty assessment", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/encounters/calculate-xp", map[string]any{
			"party_levels": []int{1, 1, 1, 1},
			"monsters": []map[string]any{
				{"name": "Goblin", "cr": "1/4", "count": 3, "xp": 50},
			},
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if result["total_xp"].(float64) != 150 {
			t.Fatalf("expected 150 XP, got %v", result["total_xp"])
		}
	})
}

func TestEncountersMonsterCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/encounters", CreateEncounter)
		auth.POST("/encounters/:id/monsters", AddEncounterMonster)
		auth.PUT("/encounter-monsters/:mid", UpdateEncounterMonster)
		auth.DELETE("/encounter-monsters/:mid", DeleteEncounterMonster)
	})

	t.Run("full monster lifecycle", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/encounters", map[string]any{
			"name": "Test", "difficulty": "easy",
		})
		var enc map[string]any
		testutil.ParseJSON(t, w, &enc)
		eid := int(enc["id"].(float64))

		w = testutil.PostJSON(t, r, "/api/encounters/"+itoa(eid)+"/monsters", map[string]any{
			"name": "Goblin", "count": 3, "cr": "1/4", "xp": 50, "ac": 15, "hp": 7,
		})
		testutil.AssertStatus(t, w, 201)
		var mon map[string]any
		testutil.ParseJSON(t, w, &mon)
		mid := int(mon["id"].(float64))

		w = testutil.PutJSON(t, r, "/api/encounter-monsters/"+itoa(mid), map[string]any{
			"name": "Goblin Boss", "count": 1, "cr": "1", "xp": 200, "ac": 17, "hp": 30,
		})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Delete(t, r, "/api/encounter-monsters/"+itoa(mid))
		testutil.AssertStatus(t, w, 200)
	})
}
