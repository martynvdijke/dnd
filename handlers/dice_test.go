package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestDiceRoll(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/roll", HandleRoll)
		auth.GET("/dice-rolls", GetDiceRolls)
	})
	_ = getDicePool()

	t.Run("valid dice expression returns 200", func(t *testing.T) {
		exprs := []string{"1d4", "1d6", "1d8", "1d10", "1d12", "1d20", "1d100", "2d6", "1d8+3", "3d4+2"}
		for _, expr := range exprs {
			w := testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": expr})
			testutil.AssertStatus(t, w, 200)
			var result map[string]any
			testutil.ParseJSON(t, w, &result)
			if _, ok := result["total"]; !ok {
				t.Fatalf("roll %s: missing total", expr)
			}
		}
	})

	t.Run("empty expression returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": ""})
		if w.Code != 400 {
			t.Fatalf("expected 400 for empty expression, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid expression returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": "invalid!!"})
		if w.Code != 400 {
			t.Fatalf("expected 400 for invalid expression, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("dice history returns entries", func(t *testing.T) {
		testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": "1d20"})
		testutil.PostJSON(t, r, "/api/roll", map[string]any{"expression": "1d6"})

		w := testutil.Get(t, r, "/api/dice-rolls")
		testutil.AssertStatus(t, w, 200)
		var rolls []any
		testutil.ParseJSON(t, w, &rolls)
		if len(rolls) < 1 {
			t.Fatal("expected at least 1 dice roll in history")
		}
	})
}

func TestDiceAdvantageDisadvantage(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/roll", HandleRoll)
	})
	_ = getDicePool()

	t.Run("advantage returns two rolls with max", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{
			"expression": "1d20",
			"advantage":  "advantage",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("disadvantage returns two rolls with min", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{
			"expression": "1d20",
			"advantage":  "disadvantage",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("roll with character id stores history", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll", map[string]any{
			"expression":   "1d20+5",
			"character_id": 1,
		})
		testutil.AssertStatus(t, w, 200)
	})
}

func TestDiceCheckRoll(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Skill Checker", "Elf", "Rogue")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/roll/check", HandleCheckRoll)
	})
	_ = getDicePool()

	t.Run("skill check returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll/check", map[string]any{
			"character_id": 1, "type": "skill", "name": "perception",
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if _, ok := result["total"]; !ok {
			t.Fatal("check roll missing total")
		}
	})

	t.Run("unknown skill returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll/check", map[string]any{
			"character_id": 1, "type": "skill", "name": "nonexistent",
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for unknown skill, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid type returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/roll/check", map[string]any{
			"character_id": 1, "type": "invalid", "name": "str",
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for invalid type, got %d: %s", w.Code, w.Body.String())
		}
	})
}
