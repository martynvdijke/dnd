package handlers

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func TestComputeArmorClass(t *testing.T) {
	eq := func(armorType string, acBonus int) models.InventoryItem {
		return models.InventoryItem{ArmorType: armorType, ACBonus: acBonus, IsEquipped: true}
	}
	cases := []struct {
		name     string
		dexMod   int
		items    []models.InventoryItem
		wantAC   int
		wantComp bool
	}{
		{"light adds full dex", 2, []models.InventoryItem{eq("light", 11)}, 13, true},
		{"light falls back to 11", 3, []models.InventoryItem{eq("light", 0)}, 14, true},
		{"medium caps dex at 2", 4, []models.InventoryItem{eq("medium", 13)}, 15, true},
		{"heavy ignores dex", 4, []models.InventoryItem{eq("heavy", 16)}, 16, true},
		{"shield adds", 1, []models.InventoryItem{eq("light", 12), eq("", 2)}, 15, true},
		{"highest body armor wins", 0, []models.InventoryItem{eq("light", 11), eq("heavy", 16)}, 16, true},
		{"no body armor keeps manual", 3, []models.InventoryItem{eq("", 2)}, 10, false},
		{"unequipped light ignored", 3, []models.InventoryItem{{ArmorType: "light", ACBonus: 11}}, 10, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, comp := ComputeArmorClass(tc.dexMod, tc.items)
			if got != tc.wantAC || comp != tc.wantComp {
				t.Fatalf("got (%d,%v) want (%d,%v)", got, comp, tc.wantAC, tc.wantComp)
			}
		})
	}
}

func TestAttackSituationalModifiers(t *testing.T) {
	if b, adv, dis, cannot := AttackSituationalModifiers("half", false, 0, 0, 0); b != 2 || adv || dis || cannot {
		t.Fatalf("half cover got (%d,%v,%v,%v)", b, adv, dis, cannot)
	}
	if b, _, _, _ := AttackSituationalModifiers("three_quarters", false, 0, 0, 0); b != 5 {
		t.Fatalf("three quarters expected 5 got %d", b)
	}
	if _, _, _, cannot := AttackSituationalModifiers("total", false, 0, 0, 0); !cannot {
		t.Fatal("total cover should block")
	}
	if _, adv, _, _ := AttackSituationalModifiers("", true, 0, 0, 0); !adv {
		t.Fatal("flanking should grant advantage")
	}
	// within normal range: nothing
	if _, adv, dis, cannot := AttackSituationalModifiers("", false, 30, 60, 120); adv || dis || cannot {
		t.Fatalf("within range should be clean got (%v,%v,%v)", adv, dis, cannot)
	}
	// between normal and long: disadvantage
	if _, _, dis, cannot := AttackSituationalModifiers("", false, 90, 60, 120); !dis || cannot {
		t.Fatalf("long range should give disadvantage got (%v,%v)", dis, cannot)
	}
	// beyond long: blocked
	if _, _, _, cannot := AttackSituationalModifiers("", false, 150, 60, 120); !cannot {
		t.Fatal("beyond long range should block")
	}
	// no distance supplied: range ignored
	if _, adv, dis, cannot := AttackSituationalModifiers("", false, 0, 60, 120); adv || dis || cannot {
		t.Fatalf("no distance should skip range got (%v,%v,%v)", adv, dis, cannot)
	}
}

func TestRecomputeCharacterAC(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 30, 1, "Armored", "Human", "Fighter")
	db.DB.Exec("UPDATE characters SET dex=14, ac=10 WHERE id=30")

	// No armor: manual AC preserved.
	if _, comp := recomputeCharacterAC(30); comp {
		t.Fatal("expected uncomputed without armor")
	}

	// Equip leather armor (light, implicit 11) -> 11 + 2 = 13.
	db.DB.Exec("INSERT INTO inventory(character_id,name,category,armor_type,ac_bonus,is_equipped) VALUES(30,'Leather','armor','light',0,1)")
	ac, comp := recomputeCharacterAC(30)
	if !comp || ac != 13 {
		t.Fatalf("expected (13,true) got (%d,%v)", ac, comp)
	}
	var stored int
	db.DB.QueryRow("SELECT ac FROM characters WHERE id=30").Scan(&stored)
	if stored != 13 {
		t.Fatalf("expected persisted AC 13 got %d", stored)
	}

	// Add a shield -> +2.
	db.DB.Exec("INSERT INTO inventory(character_id,name,category,ac_bonus,is_equipped) VALUES(30,'Shield','armor',2,1)")
	if ac, _ := recomputeCharacterAC(30); ac != 15 {
		t.Fatalf("expected 15 with shield got %d", ac)
	}
}

func TestHandleCombatEconomy(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	db.DB.Exec("INSERT INTO combat_entries(name,type,ac,hp_max,hp_current,is_active,turn_order,movement_max) VALUES('Goblin','monster',15,7,7,1,0,30)")
	var id int64
	db.DB.QueryRow("SELECT id FROM combat_entries WHERE name='Goblin'").Scan(&id)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat/:id/economy", HandleCombatEconomy)
	})
	economyPath := "/api/combat/" + strconv.FormatInt(id, 10) + "/economy"

	w := testutil.PostJSON(t, r, economyPath, map[string]any{"action": true, "movement": 10})
	testutil.AssertStatus(t, w, 200)
	var e CombatEntry
	testutil.ParseJSON(t, w, &e)
	if !e.ActionUsed || e.MovementUsed != 10 {
		t.Fatalf("expected action used and 10ft got %+v", e)
	}

	// Spending the action again is rejected.
	w = testutil.PostJSON(t, r, economyPath, map[string]any{"action": true})
	testutil.AssertStatus(t, w, 400)

	// Overspending movement is rejected.
	w = testutil.PostJSON(t, r, economyPath, map[string]any{"movement": 25})
	testutil.AssertStatus(t, w, 400)

	// Remaining movement works.
	w = testutil.PostJSON(t, r, economyPath, map[string]any{"movement": 20})
	testutil.AssertStatus(t, w, 200)
}

func TestHandleCombatAttackCover(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Attacker", "Human", "Fighter")
	testutil.SeedCharacter(t, 2, 1, "Target", "Elf", "Wizard")
	db.DB.Exec("UPDATE characters SET str=16, level=1, proficiency_bonus=2 WHERE id=1")
	db.DB.Exec("UPDATE characters SET ac=15, hp_max=20, hp_current=20 WHERE id=2")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat/attack", HandleCombatAttack)
	})

	// Half cover raises effective AC by 2 for the comparison.
	w := testutil.PostJSON(t, r, "/api/combat/attack", map[string]any{
		"attacker_type": "character", "attacker_id": 1,
		"target_type": "character", "target_id": 2,
		"attack_bonus": 1, "damage_dice": "1d1", "apply": false,
		"cover": "half",
	})
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if got, _ := resp["target_ac"].(float64); got != 17 {
		t.Fatalf("expected effective AC 17 got %v", resp["target_ac"])
	}
	if got, _ := resp["cover_bonus"].(float64); got != 2 {
		t.Fatalf("expected cover bonus 2 got %v", resp["cover_bonus"])
	}

	// Total cover blocks regardless of attack bonus.
	w = testutil.PostJSON(t, r, "/api/combat/attack", map[string]any{
		"attacker_type": "character", "attacker_id": 1,
		"target_type": "character", "target_id": 2,
		"attack_bonus": 100, "damage_dice": "1d1", "apply": false,
		"cover": "total",
	})
	testutil.AssertStatus(t, w, 200)
	resp = nil
	testutil.ParseJSON(t, w, &resp)
	if hit, _ := resp["hit"].(bool); hit {
		t.Fatalf("expected blocked attack got %+v", resp)
	}
	if cannot, _ := resp["cannot_target"].(bool); !cannot {
		t.Fatalf("expected cannot_target true got %+v", resp)
	}
	if dmg, _ := resp["damage"].(float64); dmg != 0 {
		t.Fatalf("expected 0 damage got %v", resp["damage"])
	}
}
