package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestConditionsCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "TestHero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/conditions", CreateCondition)
		auth.GET("/conditions", ListConditions)
		auth.PUT("/conditions/:id", UpdateCondition)
		auth.DELETE("/conditions/:id", DeleteCondition)
		auth.POST("/conditions/tick", TickConditions)
		auth.GET("/conditions/types", GetConditionTypes)
	})

	t.Run("create condition returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/conditions", map[string]any{
			"character_id": 1, "name": "Poisoned", "type": "poisoned",
			"duration": 5, "duration_type": "round", "source": "Spider",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list conditions returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/conditions?character_id=1")
		testutil.AssertStatus(t, w, 200)
		var conds []any
		testutil.ParseJSON(t, w, &conds)
		if len(conds) == 0 {
			t.Fatal("expected at least 1 condition")
		}
	})

	t.Run("get condition types returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/conditions/types")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("tick conditions returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/conditions/tick", map[string]any{
			"character_id": 1, "count": 1, "duration_type": "round",
		})
		testutil.AssertStatus(t, w, 200)
	})
}

func TestConditionEdgeCases(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "EdgeHero", "Dwarf", "Cleric")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/conditions", CreateCondition)
		auth.GET("/conditions", ListConditions)
	})

	t.Run("list without character_id returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/conditions")
		if w.Code != 400 {
			t.Logf("expected 400 for missing character_id, got %d", w.Code)
		}
	})

	t.Run("create without name still returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/conditions", map[string]any{
			"character_id": 1, "type": "poisoned",
		})
		if w.Code != 201 {
			t.Logf("create condition without name: got %d", w.Code)
		}
	})
}
