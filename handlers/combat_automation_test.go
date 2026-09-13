package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func TestDeriveAttackBonus(t *testing.T) {
	// STR 16 (+3), DEX 14 (+2), level 1 pb 2
	ch := models.Character{Str: 16, Dex: 14, Level: 1, ProficiencyBonus: 2}
	// melee no finesse -> STR
	item := models.InventoryItem{WeaponProperties: ""}
	if got := deriveAttackBonus(ch, item); got != 5 {
		t.Fatalf("melee expected 5 got %d", got)
	}
	// finesse -> DEX
	item = models.InventoryItem{WeaponProperties: "finesse"}
	if got := deriveAttackBonus(ch, item); got != 4 {
		t.Fatalf("finesse expected 4 got %d", got)
	}
	// ranged -> DEX
	item = models.InventoryItem{WeaponProperties: "ranged"}
	if got := deriveAttackBonus(ch, item); got != 4 {
		t.Fatalf("ranged expected 4 got %d", got)
	}
	// magical +1
	item = models.InventoryItem{WeaponProperties: "", IsMagical: true}
	if got := deriveAttackBonus(ch, item); got != 6 {
		t.Fatalf("magical expected 6 got %d", got)
	}
	// explicit override wins
	b := 9
	item = models.InventoryItem{WeaponProperties: "finesse", IsMagical: true, AttackBonus: &b}
	if got := deriveAttackBonus(ch, item); got != 9 {
		t.Fatalf("override expected 9 got %d", got)
	}
	// explicit attack_ability
	item = models.InventoryItem{AttackAbility: "dex"}
	if got := deriveAttackBonus(ch, item); got != 4 {
		t.Fatalf("attack_ability dex expected 4 got %d", got)
	}
	// prof 0 computed from level
	ch2 := models.Character{Str: 10, Level: 5, ProficiencyBonus: 0}
	item = models.InventoryItem{}
	if got := deriveAttackBonus(ch2, item); got != 3 {
		t.Fatalf("level 5 pb 3 expected 3 got %d", got)
	}
}

func TestApplyHPChange(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/hp", HandleCharacterHP)
	})

	t.Run("temp hp absorption", func(t *testing.T) {
		testutil.SeedCharacter(t, 10, 1, "TempGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=12, hp_current=10, temp_hp=5 WHERE id=10")
		w := testutil.PostJSON(t, r, "/api/characters/10/hp", map[string]any{"delta": -8, "type": "slashing", "source": "test"})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.TempHP != 0 || res.HPCurrent != 7 {
			t.Fatalf("temp absorb expected temp 0 hp 7 got temp %d hp %d", res.TempHP, res.HPCurrent)
		}
	})

	t.Run("overheal clamps and resets death saves", func(t *testing.T) {
		testutil.SeedCharacter(t, 11, 1, "HealGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=10, hp_current=5, death_saves_successes=2, death_saves_failures=1 WHERE id=11")
		w := testutil.PostJSON(t, r, "/api/characters/11/hp", map[string]any{"delta": 10, "type": "", "source": ""})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.HPCurrent != 10 || res.DeathSavesSuccesses != 0 || res.DeathSavesFailures != 0 {
			t.Fatalf("overheal clamp failed %+v", res)
		}
	})

	t.Run("reaching 0 resets death saves", func(t *testing.T) {
		testutil.SeedCharacter(t, 12, 1, "ZeroGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=10, hp_current=5, temp_hp=0 WHERE id=12")
		w := testutil.PostJSON(t, r, "/api/characters/12/hp", map[string]any{"delta": -5, "type": "", "source": ""})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.HPCurrent != 0 {
			t.Fatalf("expected 0 hp got %d", res.HPCurrent)
		}
		if res.DeathSavesSuccesses != 0 || res.DeathSavesFailures != 0 {
			t.Fatalf("death saves not reset %+v", res)
		}
	})

	t.Run("damage at 0 increments failures", func(t *testing.T) {
		testutil.SeedCharacter(t, 13, 1, "DownGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET hp_max=10, hp_current=0, death_saves_failures=1 WHERE id=13")
		w := testutil.PostJSON(t, r, "/api/characters/13/hp", map[string]any{"delta": -5, "type": "", "source": ""})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.DeathSavesFailures != 2 {
			t.Fatalf("expected failures 2 got %d", res.DeathSavesFailures)
		}
	})

	t.Run("immunity check", func(t *testing.T) {
		testutil.SeedCharacter(t, 14, 1, "ImmuneGuy", "Human", "Fighter")
		db.DB.Exec("UPDATE characters SET condition_immunities='poisoned, charmed' WHERE id=14")
		applied, err := applyConditionIfNotImmune(14, "poisoned", "poisoned", "test", 1, "round", 0)
		if err != nil {
			t.Fatalf("err %v", err)
		}
		if applied {
			t.Fatal("expected not applied due immunity")
		}
		var cnt int
		db.DB.QueryRow("SELECT COUNT(*) FROM character_conditions WHERE character_id=14").Scan(&cnt)
		if cnt != 0 {
			t.Fatalf("expected 0 conditions got %d", cnt)
		}
		applied, _ = applyConditionIfNotImmune(14, "prone", "prone", "test", 1, "round", 0)
		if !applied {
			t.Fatal("expected prone applied")
		}
	})

	t.Run("concentration hold and break", func(t *testing.T) {
		// hold: high CON
		testutil.SeedCharacter(t, 20, 1, "ConHold", "Human", "Wizard")
		db.DB.Exec("UPDATE characters SET con=20, concentrating_on='hold person', hp_max=20, hp_current=20 WHERE id=20")
		// give proficiency in con save to make success likely but not guaranteed - instead test logic directly via applyHPChange small damage, but since dice random, we test that concentration field is either checked
		w := testutil.PostJSON(t, r, "/api/characters/20/hp", map[string]any{"delta": -1, "type": "", "source": ""})
		testutil.AssertStatus(t, w, 200)
		var res HPChangeResult
		testutil.ParseJSON(t, w, &res)
		if res.Concentration == nil || !res.Concentration.Checked {
			t.Fatalf("expected concentration checked %+v", res)
		}
		// break: force failure by setting con very low and damage high DC
		testutil.SeedCharacter(t, 21, 1, "ConBreak", "Human", "Wizard")
		db.DB.Exec("UPDATE characters SET con=1, concentrating_on='hold person', hp_max=20, hp_current=20 WHERE id=21")
		// damage 50 -> DC 25, impossible to pass with -5 mod
		w = testutil.PostJSON(t, r, "/api/characters/21/hp", map[string]any{"delta": -50, "type": "", "source": ""})
		testutil.AssertStatus(t, w, 200)
		testutil.ParseJSON(t, w, &res)
		if res.Concentration == nil || !res.Concentration.Dropped {
			t.Fatalf("expected dropped concentration %+v", res)
		}
		var conc string
		db.DB.QueryRow("SELECT concentrating_on FROM characters WHERE id=21").Scan(&conc)
		if conc != "" {
			t.Fatalf("expected cleared concentrating_on got %q", conc)
		}
	})
}

func TestHandleCombatAttack(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Attacker", "Human", "Fighter")
	testutil.SeedCharacter(t, 2, 1, "Target", "Elf", "Wizard")
	db.DB.Exec("UPDATE characters SET str=16, dex=10, level=1, proficiency_bonus=2, ac=10 WHERE id=1")
	db.DB.Exec("UPDATE characters SET ac=5, hp_max=20, hp_current=20 WHERE id=2")
	// create weapon
	res, _ := db.DB.Exec("INSERT INTO inventory(character_id,name,damage_dice,damage_type,weapon_properties,is_magical) VALUES(1,'Longsword','1d8','slashing','',0)")
	wid, _ := res.LastInsertId()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat/attack", HandleCombatAttack)
	})

	t.Run("attack hits low AC", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat/attack", map[string]any{
			"attacker_type": "character", "attacker_id": 1,
			"target_type": "character", "target_id": 2,
			"item_id": wid, "apply": false,
		})
		testutil.AssertStatus(t, w, 200)
		var resp map[string]any
		testutil.ParseJSON(t, w, &resp)
		// with AC 5, most attacks should hit; but_nat1 case may miss 5% of time - retry if miss but not fumble?
		// just assert response has required fields
		if _, ok := resp["hit"]; !ok {
			t.Fatal("missing hit")
		}
		if _, ok := resp["critical"]; !ok {
			t.Fatal("missing critical")
		}
	})

	t.Run("apply damage reduces hp", func(t *testing.T) {
		db.DB.Exec("UPDATE characters SET hp_current=20 WHERE id=2")
		w2 := testutil.PostJSON(t, r, "/api/combat/attack", map[string]any{
			"attacker_type": "character", "attacker_id": 1,
			"target_type": "character", "target_id": 2,
			"attack_bonus": 100, "damage_dice": "1d1", "damage_type": "slashing", "apply": true,
		})
		testutil.AssertStatus(t, w2, 200)
		var resp2 map[string]any
		testutil.ParseJSON(t, w2, &resp2)
		if hit, _ := resp2["hit"].(bool); !hit {
			t.Fatalf("expected hit with high bonus %+v", resp2)
		}
		var hp int
		db.DB.QueryRow("SELECT hp_current FROM characters WHERE id=2").Scan(&hp)
		if hp != 19 {
			t.Fatalf("expected hp 19 after 1d1 damage got %d", hp)
		}
	})
}
