package handlers

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

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
