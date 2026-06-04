package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestCompanionCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "CompanionHero", "Human", "Ranger")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/companions", ListCompanions)
		auth.POST("/companions", CreateCompanion)
		auth.PUT("/companions/:id", UpdateCompanion)
		auth.DELETE("/companions/:id", DeleteCompanion)
	})

	t.Run("list companions returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/companions?character_id=1")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create companion returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/companions", map[string]any{
			"character_id": 1,
			"name":         "Wolf Companion",
			"type":         "beast",
			"race":         "Wolf",
			"hp_max":       30,
			"hp_current":   30,
			"ac":           14,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("create companion without name returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/companions", map[string]any{
			"character_id": 1,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for missing name, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update companion returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/companions/1", map[string]any{
			"name":   "Dire Wolf Companion",
			"hp_max": 50,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete companion returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/companions/1")
		testutil.AssertStatus(t, w, 200)
	})
}
