package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestRollCombatInitiativeBatch(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat", CreateCombatEntry)
		auth.GET("/combat", ListCombatEntries)
		auth.POST("/combat/roll-initiative", RollCombatInitiative)
	})

	for _, name := range []string{"Goblin", "Orc", "Wolf"} {
		w := testutil.PostJSON(t, r, "/api/combat", map[string]any{
			"name": name, "type": "monster", "initiative_mod": 2,
			"hp_max": 7, "hp_current": 7, "ac": 12,
		})
		testutil.AssertStatus(t, w, 201)
	}

	w := testutil.PostJSON(t, r, "/api/combat/roll-initiative", map[string]any{})
	testutil.AssertStatus(t, w, 200)

	w = testutil.Get(t, r, "/api/combat")
	testutil.AssertStatus(t, w, 200)
	var entries []struct {
		Name           string `json:"name"`
		InitiativeRoll int    `json:"initiative_roll"`
		TurnOrder      int    `json:"turn_order"`
	}
	testutil.ParseJSON(t, w, &entries)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.InitiativeRoll < 3 || e.InitiativeRoll > 22 {
			t.Fatalf("initiative roll %d out of range for %s", e.InitiativeRoll, e.Name)
		}
		if e.TurnOrder < 1 || e.TurnOrder > 3 {
			t.Fatalf("turn order %d out of range for %s", e.TurnOrder, e.Name)
		}
	}
}

func TestClearCombat(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat", CreateCombatEntry)
		auth.GET("/combat", ListCombatEntries)
		auth.POST("/combat/clear", ClearCombat)
	})

	testutil.AssertStatus(t, testutil.PostJSON(t, r, "/api/combat", map[string]any{
		"name": "Goblin", "type": "monster", "hp_max": 7, "hp_current": 0, "ac": 12,
	}), 201)

	testutil.AssertStatus(t, testutil.PostJSON(t, r, "/api/combat/clear", map[string]any{}), 200)

	w := testutil.Get(t, r, "/api/combat")
	testutil.AssertStatus(t, w, 200)
	var entries []any
	testutil.ParseJSON(t, w, &entries)
	if len(entries) != 0 {
		t.Fatalf("expected the tracker to be empty, got %d entries", len(entries))
	}
}

func TestMonsterAttackerUsesProvidedName(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat", CreateCombatEntry)
		auth.POST("/combat/attack", HandleCombatAttack)
	})

	var target struct {
		ID int64 `json:"id"`
	}
	w := testutil.PostJSON(t, r, "/api/combat", map[string]any{
		"name": "Target Dummy", "type": "monster", "hp_max": 30, "hp_current": 30, "ac": 5,
	})
	testutil.AssertStatus(t, w, 201)
	testutil.ParseJSON(t, w, &target)

	var res struct {
		Hit      bool `json:"hit"`
		LogID    int  `json:"log_id"`
		TargetHP *int `json:"target_hp"`
	}
	// A natural 1 always misses, so a single roll would flake at 5%. Retry until
	// the attack lands; 20 consecutive misses is a ~1e-26 event.
	for attempt := 0; attempt < 20; attempt++ {
		w = testutil.PostJSON(t, r, "/api/combat/attack", map[string]any{
			"attacker_type": "combat",
			"attacker_name": "Goblin Boss",
			"target_type":   "combat",
			"target_id":     target.ID,
			"attack_bonus":  100,
			"damage_dice":   "1d6",
			"apply":         true,
		})
		testutil.AssertStatus(t, w, 200)
		testutil.ParseJSON(t, w, &res)
		if res.Hit {
			break
		}
	}
	if !res.Hit {
		t.Fatal("expected the monster attack to hit a low-AC target")
	}
	if res.TargetHP == nil || *res.TargetHP >= 30 {
		t.Fatalf("expected target HP to drop, got %v", res.TargetHP)
	}
	if res.LogID == 0 {
		t.Fatal("expected a combat log entry")
	}
	var actor string
	if err := db.DB.QueryRow("SELECT actor_name FROM combat_log_entries WHERE id=?", res.LogID).Scan(&actor); err != nil {
		t.Fatalf("query combat log: %v", err)
	}
	if actor != "Goblin Boss" {
		t.Fatalf("expected attacker name Goblin Boss, got %q", actor)
	}
}

func TestAwardCharacterXP(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.POST("/characters/:id/xp", AwardCharacterXP)
	})

	var ch struct {
		ID int64 `json:"id"`
	}
	w := testutil.PostJSON(t, r, "/api/characters", map[string]any{"name": "XP Taker", "level": 1, "xp": 10})
	testutil.AssertStatus(t, w, 201)
	testutil.ParseJSON(t, w, &ch)

	w = testutil.PostJSON(t, r, "/api/characters/1/xp", map[string]any{"amount": 200})
	testutil.AssertStatus(t, w, 200)
	var res struct {
		XP int `json:"xp"`
	}
	testutil.ParseJSON(t, w, &res)
	if res.XP != 210 {
		t.Fatalf("expected xp 210, got %d", res.XP)
	}

	w = testutil.PostJSON(t, r, "/api/characters/1/xp", map[string]any{"amount": -5})
	testutil.AssertStatus(t, w, 400)
}
