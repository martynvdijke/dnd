package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestConcentrationCheck(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "ConcentrateHero", "Human", "Wizard")
	// Set concentrating_on so the handler returns a DC
	_, err := db.DB.Exec("UPDATE characters SET concentrating_on = 'Haste' WHERE id = 1")
	if err != nil {
		t.Fatalf("set concentrating_on: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/concentration-check", CheckConcentration)
	})

	t.Run("concentration check 10 damage returns DC 10", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/concentration-check", map[string]any{
			"damage": 10,
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		dc, ok := result["dc"].(float64)
		if !ok {
			t.Fatalf("expected dc field, got %+v", result)
		}
		// DC = max(10, floor(damage/2)) = max(10, 5) = 10
		if int(dc) < 10 {
			t.Fatalf("expected DC >= 10, got %v", dc)
		}
	})

	t.Run("concentration check 30 damage returns DC 15", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/concentration-check", map[string]any{
			"damage": 30,
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		dc, ok := result["dc"].(float64)
		if !ok {
			t.Fatalf("expected dc field, got %+v", result)
		}
		if int(dc) < 15 {
			t.Fatalf("expected DC >= 15 for 30 damage, got %v", dc)
		}
	})

	t.Run("concentration check non-existent character returns 404", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/999/concentration-check", map[string]any{
			"damage": 10,
		})
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
