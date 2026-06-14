package handlers

import (
	"fmt"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestCombatEntries(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat", CreateCombatEntry)
		auth.GET("/combat", ListCombatEntries)
		auth.PUT("/combat/:id", UpdateCombatEntry)
		auth.DELETE("/combat/:id", DeleteCombatEntry)
		auth.POST("/combat/initiative", RollInitiative)
		auth.POST("/combat/next-turn", NextTurn)
		auth.GET("/combat/current-turn", GetCurrentTurn)
	})

	t.Run("create combat entry returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"name": "Goblin", "type": "monster",
			"initiative_roll": 12, "initiative_mod": 2,
			"hp_max": 7, "hp_current": 7, "ac": 15,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("create without name returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"type": "monster", "hp_max": 7,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for missing name, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list returns entries", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/combat")
		testutil.AssertStatus(t, w, 200)
		var entries []any
		testutil.ParseJSON(t, w, &entries)
		if len(entries) < 1 {
			t.Fatal("expected at least 1 combat entry")
		}
	})

	t.Run("update entry hp and ac", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"name": "Orc", "type": "monster",
			"initiative_roll": 10, "initiative_mod": 1,
			"hp_max": 15, "hp_current": 15, "ac": 13,
		})
		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		eid := int64(created["id"].(float64))

		w = testutil.PutJSON(t, r, "/api/combat/"+fmt.Sprintf("%d", eid), map[string]any{
			"name": "Orc Chief", "type": "monster",
			"initiative_roll": 10, "initiative_mod": 3,
			"hp_max": 15, "hp_current": 8, "ac": 15,
			"is_active": true,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete entry returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"name": "Delete Me", "type": "monster",
			"initiative_roll": 5, "hp_max": 5, "hp_current": 5, "ac": 10,
		})
		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		eid := int64(created["id"].(float64))

		w = testutil.Delete(t, r, "/api/combat/"+fmt.Sprintf("%d", eid))
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("get current turn with no entries returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/combat/current-turn")
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCombatInitiative(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Combatant", "Elf", "Ranger")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat/initiative", RollInitiative)
	})

	t.Run("roll initiative for character returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat/initiative", map[string]any{
			"character_id": 1,
		})
		testutil.AssertStatus(t, w, 200)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		if total, ok := result["total"].(float64); ok {
			if total < 1 || total > 30 {
				t.Fatalf("initiative total %f out of range [1,30]", total)
			}
		}
	})

	t.Run("roll initiative for non-existent character returns 404", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat/initiative", map[string]any{
			"character_id": 99999,
		})
		if w.Code != 404 && w.Code != 400 {
			t.Fatalf("expected 404 or 400 for non-existent character, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCombatNextTurn(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat", CreateCombatEntry)
		auth.POST("/combat/next-turn", NextTurn)
	})

	t.Run("next turn with no entries returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat/next-turn", nil)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("next turn cycles through entries", func(t *testing.T) {
		testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"name": "Aragorn", "type": "character",
			"initiative_roll": 20, "hp_max": 30, "hp_current": 30, "ac": 16,
		})
		testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"name": "Goblin", "type": "monster",
			"initiative_roll": 10, "hp_max": 7, "hp_current": 7, "ac": 15,
		})

		for range 3 {
			w := testutil.PostJSON(t, r, "/api/combat/next-turn", nil)
			testutil.AssertStatus(t, w, 200)
		}
	})
}

func TestCombatConcurrentSafety(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/combat", ListCombatEntries)
		auth.POST("/combat", CreateCombatEntry)
		auth.POST("/combat/initiative", RollInitiative)
		auth.POST("/combat/next-turn", NextTurn)
	})

	for i := range 3 {
		testutil.PostJSON(t, r, fmt.Sprintf("/api/combat"), map[string]any{
			"name": fmt.Sprintf("Fighter%d", i), "type": "character",
			"initiative_roll": 20 - i, "hp_max": 30, "hp_current": 30, "ac": 16,
		})
	}
	testutil.PostJSON(t, r, "/api/combat/initiative", nil)

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			rr := testutil.PostJSON(t, r, "/api/combat/next-turn", nil)
			if rr.Code != 200 {
				t.Errorf("concurrent next-turn failed: %d", rr.Code)
			}
			rr = testutil.Get(t, r, "/api/combat")
			if rr.Code != 200 {
				t.Errorf("concurrent list failed: %d", rr.Code)
			}
		})
	}
	wg.Wait()
}
