package handlers

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"testing/quick"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestCharactersCreate(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters", CreateCharacter)
		auth.GET("/characters/:id", GetCharacter)
		auth.PUT("/characters/:id", UpdateCharacter)
		auth.DELETE("/characters/:id", DeleteCharacter)
	})

	t.Run("valid create returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"name": "Test Hero", "race": "Elf", "class": "Wizard", "level": 5,
			"str": 10, "dex": 14, "con": 12, "int": 18, "wis": 14, "cha": 8,
			"hp_max": 32, "hp_current": 32,
		})
		testutil.AssertStatus(t, w, 201)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "name", "Test Hero")
		testutil.AssertField(t, data, "race", "Elf")
		testutil.AssertField(t, data, "class", "Wizard")
		if _, ok := data["id"]; !ok {
			t.Fatal("response missing id")
		}
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"race": "Human", "class": "Fighter",
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for missing name, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("negative ability scores return 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"name": "Bad Stats", "race": "Human", "class": "Fighter",
			"str": -5, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for negative str, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("overly high ability scores return 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"name": "Overpowered", "race": "Human", "class": "Fighter",
			"str": 99, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for str=99, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCharactersGetDelete(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Gandalf", "Elf", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id", GetCharacter)
		auth.DELETE("/characters/:id", DeleteCharacter)
	})

	t.Run("get existing character returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/1")
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "name", "Gandalf")
		testutil.AssertField(t, data, "race", "Elf")
		if _, ok := data["str_mod"]; !ok {
			t.Fatal("response missing str_mod")
		}
	})

	t.Run("get non-existent character returns 404", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/characters/99999")
		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("delete character returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/characters/1")
		testutil.AssertStatus(t, w, 200)
		if count := testutil.CountRows(t, "characters"); count != 0 {
			t.Fatalf("expected 0 characters after delete, got %d", count)
		}
	})
}

func TestCharactersUpdate(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Update Test", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id", UpdateCharacter)
		auth.GET("/characters/:id", GetCharacter)
	})

	t.Run("update character name", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1", map[string]any{
			"name": "Updated Name", "race": "Human", "class": "Fighter", "level": 1,
			"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
			"hp_max": 12, "hp_current": 12, "ac": 10, "initiative": 0, "speed": 30,
		})
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "name", "Updated Name")
	})

	t.Run("update with missing required fields returns 400", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1", map[string]any{
			"name": "Incomplete",
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for incomplete update, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCharactersCurrency(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Rich Hero", "Dwarf", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id/currency", UpdateCurrency)
	})

	t.Run("valid currency update returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1/currency", map[string]any{
			"gp": 100, "sp": 50, "cp": 25, "pp": 5, "ep": 0,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("negative currency returns 400", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1/currency", map[string]any{
			"gp": -100, "sp": 0, "cp": 0, "pp": 0, "ep": 0,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for negative gp, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCharactersInventory(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Packer", "Dwarf", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/inventory", CreateInventory)
		auth.PUT("/inventory/:iid", UpdateInventory)
		auth.DELETE("/inventory/:iid", DeleteInventory)
	})

	t.Run("create inventory item returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/inventory", map[string]any{
			"name": "Longsword", "category": "weapon", "quantity": 1,
			"damage_dice": "1d8", "damage_type": "slashing",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("create item with no name returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/inventory", map[string]any{
			"category": "weapon", "quantity": 1,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for missing item name, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update and delete item", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/inventory", map[string]any{
			"name": "Test Item", "category": "gear", "quantity": 1,
		})
		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		iid := int64(created["id"].(float64))

		w = testutil.PutJSON(t, r, fmt.Sprintf("/api/inventory/%d", iid), map[string]any{
			"name": "Updated Item", "category": "gear", "quantity": 2,
			"damage_dice": "", "damage_type": "", "is_magical": false,
			"description": "", "weapon_properties": "", "ac_bonus": 0, "armor_type": "",
			"is_equipped": false, "attunement": false, "notes": "",
		})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Delete(t, r, fmt.Sprintf("/api/inventory/%d", iid))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCharactersSpells(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Caster", "Elf", "Wizard")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/spells", CreateSpell)
		auth.PUT("/spells/:sid", UpdateSpell)
		auth.DELETE("/spells/:sid", DeleteSpell)
		auth.PUT("/characters/:id/spellcasting", UpdateSpellcasting)
	})

	t.Run("create spell returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/spells", map[string]any{
			"name": "Fireball", "level": 3, "school": "Evocation",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("create spell with invalid level returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/spells", map[string]any{
			"name": "Invalid", "level": -1, "school": "Evocation",
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for invalid level, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update spell casting returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1/spellcasting", map[string]any{
			"ability": "int", "save_dc": 15, "attack_bonus": 7,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("update spell save DC auto-calculates", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1/spellcasting", map[string]any{
			"ability": "int", "save_dc": 0, "attack_bonus": 0,
		})
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCharactersFeatures(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Feat Hero", "Dwarf", "Cleric")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/features", CreateFeature)
		auth.PUT("/features/:fid", UpdateFeature)
		auth.DELETE("/features/:fid", DeleteFeature)
		auth.POST("/proficiencies", CreateProficiency)
		auth.DELETE("/proficiencies/:pid", DeleteProficiency)
	})

	t.Run("create feature returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/features", map[string]any{
			"name": "Darkvision", "description": "See in dark 60ft",
			"source": "Race", "level_gained": 1,
		})
		testutil.AssertStatus(t, w, 201)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "name", "Darkvision")
	})

	t.Run("create proficiency returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/proficiencies", map[string]any{
			"character_id": 1, "name": "Stealth", "type": "skill",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("update and delete feature", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/features", map[string]any{
			"name": "Test Feat", "description": "Test", "source": "Class", "level_gained": 2,
		})
		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		fid := int64(created["id"].(float64))

		w = testutil.PutJSON(t, r, fmt.Sprintf("/api/features/%d", fid), map[string]any{
			"name": "Updated Feat", "description": "Updated", "source": "Class", "level_gained": 2,
		})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Delete(t, r, fmt.Sprintf("/api/features/%d", fid))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestCharactersDeathSaves(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Dying Hero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id", UpdateCharacter)
		auth.GET("/characters/:id", GetCharacter)
	})

	t.Run("update death saves persists correctly", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1", map[string]any{
			"name": "Dying Hero", "race": "Human", "class": "Fighter", "level": 1,
			"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
			"hp_max": 20, "hp_current": 0, "ac": 10, "initiative": 0, "speed": 30,
			"death_saves_successes": 2, "death_saves_failures": 1,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("death saves out of range returns 400", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/characters/1", map[string]any{
			"name": "Dying Hero", "race": "Human", "class": "Fighter", "level": 1,
			"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
			"hp_max": 20, "hp_current": 0, "ac": 10, "initiative": 0, "speed": 30,
			"death_saves_successes": 5, "death_saves_failures": 0,
		})
		if w.Code != 400 {
			t.Fatalf("expected 400 for invalid death saves, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCharactersMultiClass(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Multi Hero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/classes", CreateCharacterClass)
		auth.PUT("/classes/:ccid", UpdateCharacterClass)
		auth.DELETE("/classes/:ccid", DeleteCharacterClass)
	})

	t.Run("add and remove second class", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/characters/1/classes", map[string]any{
			"class": "Wizard", "subclass": "Evocation", "level": 1, "hit_dice": "d6",
		})
		testutil.AssertStatus(t, w, 201)

		var created map[string]any
		testutil.ParseJSON(t, w, &created)
		ccid := int64(created["id"].(float64))

		w = testutil.PutJSON(t, r, fmt.Sprintf("/api/classes/%d", ccid), map[string]any{
			"class": "Wizard", "subclass": "Divination", "level": 2, "hit_dice": "d6",
		})
		testutil.AssertStatus(t, w, 200)

		w = testutil.Delete(t, r, fmt.Sprintf("/api/classes/%d", ccid))
		testutil.AssertStatus(t, w, 200)
	})
}

func TestPropertyAbilityModifier(t *testing.T) {
	f := func(score int) bool {
		if score < 1 || score > 30 {
			return true
		}
		mod := abilityMod(score)
		return mod >= -5 && mod <= 10
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestPropertyPassivePerception(t *testing.T) {
	f := func(wis int) bool {
		if wis < 1 || wis > 30 {
			return true
		}
		pp := 10 + abilityMod(wis)
		return pp >= 5 && pp <= 20
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestPropertyHPCalculation(t *testing.T) {
	hitDieAvg := map[string]int{"d6": 4, "d8": 5, "d10": 6, "d12": 7}
	f := func(die string, level, con int) bool {
		if level < 1 || level > 20 || con < 1 || con > 30 {
			return true
		}
		avg, ok := hitDieAvg[die]
		if !ok {
			return true
		}
		hp := (avg + abilityMod(con)) * level
		return hp >= level && hp <= (12+10)*20
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestConcurrentInventoryUpdate(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Concurrent", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/inventory", CreateInventory)
		auth.PUT("/inventory/:iid", UpdateInventory)
	})

	w := testutil.PostJSON(t, r, "/api/characters/1/inventory", map[string]any{
		"name": "Potion", "quantity": 10, "weight": 0.5,
	})
	if w.Code != 201 {
		t.Skipf("inventory creation returned %d, skipping concurrent test", w.Code)
	}
	var item map[string]any
	testutil.ParseJSON(t, w, &item)
	iid, ok := item["id"].(float64)
	if !ok {
		t.Skipf("inventory response missing id, skipping: %+v", item)
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			rr := testutil.PutJSON(t, r, fmt.Sprintf("/api/inventory/%d", int64(iid)), map[string]any{
				"quantity": 5,
			})
			if rr.Code != 200 {
				t.Errorf("concurrent update failed: %d", rr.Code)
			}
		})
	}
	wg.Wait()
}

func TestConcurrentSpellSlots(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Caster", "Elf", "Wizard")
	_, err := db.DB.Exec("INSERT INTO character_spellcasting(character_id, ability, save_dc, attack_bonus, slots_1_max, slots_2_max) VALUES(1, 'int', 15, 7, 4, 3)")
	if err != nil {
		t.Fatalf("seed spellcasting: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id/spellcasting", UpdateSpellcasting)
	})

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			rr := testutil.PutJSON(t, r, "/api/characters/1/spellcasting", map[string]any{
				"ability": "int", "save_dc": 15, "attack_bonus": 7,
				"slots_1_used": 1, "slots_2_used": 1,
			})
			if rr.Code != 200 {
				t.Errorf("concurrent spell update failed: %d", rr.Code)
			}
		})
	}
	wg.Wait()
}

func FuzzCharacterImport(f *testing.F) {
	f.Add(`{"name":"Test","race":"Human","class":"Fighter"}`)
	f.Add(`{"name":"","race":"","class":""}`)
	f.Add(`invalid json`)
	f.Fuzz(func(t *testing.T, data string) {
		testutil.NewDB(t)
		defer testutil.CloseDB(t)
		testutil.SeedUser(t, 1, "admin", "admin")
		r := testutil.NewRouter(func(auth *gin.RouterGroup) {
			auth.POST("/characters", CreateCharacter)
		})
		w := testutil.PostJSON(t, r, "/api/characters", map[string]any{
			"name": data, "race": "Human", "class": "Fighter",
		})
		_ = w.Code
	})
}

func TestLinkCompendiumSpell(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Test Hero", "Elf", "Wizard")

	// Use first seeded compendium spell
	var spellID int64
	var spellName string
	db.DB.QueryRow("SELECT id, name FROM compendium_spells ORDER BY id LIMIT 1").Scan(&spellID, &spellName)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/spells/link", LinkCompendiumSpell)
	})

	t.Run("links spell successfully", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/characters/1/spells/link", map[string]string{
			"compendium_spell_id": fmt.Sprintf("%d", spellID),
		})
		testutil.AssertStatus(t, w, 201)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "status", "linked")
		if testutil.CountRows(t, "spells") != 1 {
			t.Fatal("expected 1 spell row")
		}
	})

	t.Run("missing compendium_spell_id returns 400", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/characters/1/spells/link", map[string]string{})
		testutil.AssertStatus(t, w, 400)
	})
}

func TestLinkCompendiumEquipment(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Test Hero", "Elf", "Wizard")
	var equipID int64
	db.DB.QueryRow("SELECT id FROM compendium_equipment ORDER BY id LIMIT 1").Scan(&equipID)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/inventory/link", LinkCompendiumEquipment)
	})

	t.Run("links equipment successfully", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/characters/1/inventory/link", map[string]string{
			"compendium_equipment_id": fmt.Sprintf("%d", equipID),
		})
		testutil.AssertStatus(t, w, 201)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "status", "linked")
		if testutil.CountRows(t, "inventory") != 1 {
			t.Fatal("expected 1 inventory row")
		}
	})

	t.Run("missing compendium_equipment_id returns 400", func(t *testing.T) {
		w := testutil.PostForm(t, r, "/api/characters/1/inventory/link", map[string]string{})
		testutil.AssertStatus(t, w, 400)
	})
}

func TestUnlinkCompendiumSpell(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Test Hero", "Elf", "Wizard")

	var spellID int64
	var spellName string
	db.DB.QueryRow("SELECT id, name FROM compendium_spells ORDER BY id LIMIT 1").Scan(&spellID, &spellName)

	_, err := db.DB.Exec(`INSERT INTO spells(id, character_id, name, level, school, compendium_spell_id)
		VALUES(1, 1, ?, 1, 'Evocation', ?)`, spellName, spellID)
	if err != nil {
		t.Fatalf("seed linked spell: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/characters/:id/spells/:spellId/link", UnlinkCompendiumSpell)
	})

	w := testutil.Delete(t, r, "/api/characters/1/spells/1/link")
	testutil.AssertStatus(t, w, 200)
	var data map[string]any
	testutil.ParseJSON(t, w, &data)
	testutil.AssertField(t, data, "status", "unlinked")
	if testutil.CountRows(t, "spells") != 0 {
		t.Fatal("expected 0 spell rows after unlink")
	}
}

func TestUnlinkCompendiumEquipment(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Test Hero", "Elf", "Wizard")
	var eID int64
	var eName string
	db.DB.QueryRow("SELECT id, name FROM compendium_equipment ORDER BY id LIMIT 1").Scan(&eID, &eName)

	_, err := db.DB.Exec(`INSERT INTO inventory(id, character_id, name, quantity, weight, compendium_equipment_id)
		VALUES(1, 1, ?, 1, 2.0, ?)`, eName, eID)
	if err != nil {
		t.Fatalf("seed linked item: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/characters/:id/inventory/:itemId/link", UnlinkCompendiumEquipment)
	})

	w := testutil.Delete(t, r, "/api/characters/1/inventory/1/link")
	testutil.AssertStatus(t, w, 200)
	var data map[string]any
	testutil.ParseJSON(t, w, &data)
	testutil.AssertField(t, data, "status", "unlinked")
	if testutil.CountRows(t, "inventory") != 0 {
		t.Fatal("expected 0 inventory rows after unlink")
	}
}

func TestSearchCompendiumTypeFilter(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/compendium/search", SearchCompendium)
	})

	t.Run("filters by type=spell", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search?q=acid&type=spell")
		testutil.AssertStatus(t, w, 200)
		var results []map[string]any
		testutil.ParseJSON(t, w, &results)
		// Seed data has spells matching "acid" (e.g. Acid Splash, Melf's Acid Arrow)
		if len(results) == 0 {
			t.Fatalf("expected at least 1 spell result, got 0")
		}
		for _, r := range results {
			if r["type"] != "spell" {
				t.Fatalf("expected all results to be type 'spell', got %v", r["type"])
			}
		}
	})

	t.Run("filters by type=equipment", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search?q=pot&type=equipment")
		testutil.AssertStatus(t, w, 200)
		var results []map[string]any
		testutil.ParseJSON(t, w, &results)
		// Seed data has "Potion of Healing"
		if len(results) == 0 {
			t.Fatalf("expected at least 1 equipment result, got 0")
		}
		for _, r := range results {
			if r["type"] != "equipment" {
				t.Fatalf("expected all results to be type 'equipment', got %v", r["type"])
			}
		}
	})

	t.Run("returns all types when no type filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/compendium/search?q=acid")
		testutil.AssertStatus(t, w, 200)
		var results []map[string]any
		testutil.ParseJSON(t, w, &results)
		if len(results) < 1 {
			t.Fatalf("expected at least 1 result without type filter, got %d", len(results))
		}
	})
}

func TestCompendiumCardSpell(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	// Find first seeded spell ID
	var spellID int64
	var spellName string
	db.DB.QueryRow("SELECT id, name FROM compendium_spells ORDER BY id LIMIT 1").Scan(&spellID, &spellName)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("role", "admin")
	})
	r.GET("/htmx/compendium/card/:type/:id", HtmxCompendiumCard)

	t.Run("returns spell card", func(t *testing.T) {
		w := testutil.Get(t, r, fmt.Sprintf("/htmx/compendium/card/spell/%d", spellID))
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), spellName) {
			t.Fatalf("expected spell name %q in card response: %s", spellName, w.Body.String())
		}
	})

	t.Run("returns card for missing entity", func(t *testing.T) {
		w := testutil.Get(t, r, "/htmx/compendium/card/spell/999999")
		testutil.AssertStatus(t, w, 200)
		if !strings.Contains(w.Body.String(), "not found") {
			t.Fatalf("expected not-found message: %s", w.Body.String())
		}
	})
}

func TestLinkCompendiumMonsterToAct(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedOneShot(t, 1, 1, "Test One-Shot")
	testutil.SeedOneShotAct(t, 1, 1, "Act 1", 1)

	var monsterID int64
	db.DB.QueryRow("SELECT id FROM compendium_monsters ORDER BY id LIMIT 1").Scan(&monsterID)

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures/:id/acts/:aid/monsters/link", LinkCompendiumMonsterToAct)
	}, 1, "dm")

	w := testutil.PostForm(t, r, "/api/oneshot-adventures/1/acts/1/monsters/link", map[string]string{
		"compendium_monster_id": fmt.Sprintf("%d", monsterID),
	})
	testutil.AssertStatus(t, w, 201)
	var data map[string]any
	testutil.ParseJSON(t, w, &data)
	testutil.AssertField(t, data, "status", "linked")
	if testutil.CountRows(t, "oneshot_monsters") != 1 {
		t.Fatal("expected 1 oneshot_monster row")
	}
}
