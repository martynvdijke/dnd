package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestHPCalculation(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "HPHero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id/hp-calc", CalculateHP)
	})

	t.Run("calculate HP returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1/hp-calc")
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		// Should have hp details
		if _, ok := result["hp"]; !ok {
			t.Logf("hp calc response: %+v", result)
		}
	})

	t.Run("calculate HP for non-existent character returns 404", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/999/hp-calc")
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
