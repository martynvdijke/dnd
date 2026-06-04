package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestLevelPlannerCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "PlannerHero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/level-plan", GetLevelUpPlan)
		auth.POST("/characters/:id/level-plan", SaveLevelUpPlan)
		auth.DELETE("/characters/:id/level-plan", DeleteLevelUpPlan)
		auth.GET("/characters/:id/level-suggestions", GetLevelUpSuggestions)
	})

	t.Run("get level plan empty returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/level-plan")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("save level plan returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/level-plan", map[string]any{
			"target_level": 5,
			"plan_data": []map[string]any{
				{"level": 4, "choice": "ASI", "detail": "Strength +2"},
				{"level": 5, "choice": "Extra Attack", "detail": "Fighter feature"},
			},
			"notes": "Plan to reach level 5",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("get level plan after save returns plan", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/level-plan")
		testutil.AssertStatus(t, w, 200)
		var plan map[string]any
		testutil.ParseJSON(t, w, &plan)
		if plan["target_level"] == nil {
			t.Fatal("expected target_level in plan")
		}
	})

	t.Run("get level suggestions returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/level-suggestions")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete level plan returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/characters/1/level-plan")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get after delete returns empty or 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/level-plan")
		testutil.AssertStatus(t, w, 200)
	})
}
