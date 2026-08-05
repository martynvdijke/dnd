package handlers

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

// Permission matrix: owner / non-owner / admin / DM × player / linked character.
func TestCharacterPermissionMatrix(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	// user 1 = admin (seeded for admin test); user 2 owns characters + campaign 1 (DM);
	// user 3 = unrelated player.
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "owner", "player")
	testutil.SeedUser(t, 3, "stranger", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 2)

	// char 1: owned by user 2, player type (default)
	testutil.SeedCharacter(t, 1, 2, "Hero", "Elf", "Wizard")
	// char 2: owned by user 2, linked, no campaign
	testutil.SeedCharacter(t, 2, 2, "Linked", "Human", "Fighter")
	if _, err := db.DB.Exec("UPDATE characters SET character_type='linked' WHERE id=2"); err != nil {
		t.Fatalf("update char 2: %v", err)
	}
	// char 3: owned by user 1 (admin), linked, in campaign 1 → user 2 is its DM
	testutil.SeedCharacter(t, 3, 1, "DMChar", "Dwarf", "Cleric")
	if _, err := db.DB.Exec("UPDATE characters SET character_type='linked', campaign_id=1 WHERE id=3"); err != nil {
		t.Fatalf("update char 3: %v", err)
	}

	router := func(auth *gin.RouterGroup) {
		auth.PUT("/characters/:id", UpdateCharacter)
		auth.POST("/characters/:id/inventory", CreateInventory)
	}
	body := map[string]any{"name": "Hero", "race": "Elf", "class": "Wizard"}

	t.Run("owner + player can edit", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 2, "player")
		w := testutil.PutJSON(t, r, "/api/characters/1", body)
		testutil.AssertStatus(t, w, 200)
	})
	t.Run("owner + linked cannot edit (no campaign)", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 2, "player")
		w := testutil.PutJSON(t, r, "/api/characters/2", body)
		testutil.AssertStatus(t, w, 403)
	})
	t.Run("non-owner + player cannot edit", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 3, "player")
		w := testutil.PutJSON(t, r, "/api/characters/1", body)
		testutil.AssertStatus(t, w, 403)
	})
	t.Run("admin can edit linked character", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 1, "admin")
		w := testutil.PutJSON(t, r, "/api/characters/2", body)
		testutil.AssertStatus(t, w, 200)
	})
	t.Run("campaign DM can edit linked character", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 2, "player")
		// campaign_id must be preserved: UpdateCharacter writes it back from the body
		dmBody := map[string]any{"name": "DMChar", "race": "Dwarf", "class": "Cleric", "campaign_id": 1}
		w := testutil.PutJSON(t, r, "/api/characters/3", dmBody)
		testutil.AssertStatus(t, w, 200)
	})
	t.Run("owner + linked cannot add inventory", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 2, "player")
		w := testutil.PostJSON(t, r, "/api/characters/2/inventory", map[string]any{"name": "Sword", "quantity": 1})
		testutil.AssertStatus(t, w, 403)
	})
	t.Run("DM can add inventory to linked character", func(t *testing.T) {
		r := testutil.NewRouterWithUser(router, 2, "player")
		w := testutil.PostJSON(t, r, "/api/characters/3/inventory", map[string]any{"name": "Sword", "quantity": 1})
		testutil.AssertStatus(t, w, 201)
	})
	t.Run("campaign member can view linked character (read-only)", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.GET("/characters/:id", GetCharacter)
		}, 2, "player")
		w := testutil.Get(t, r, "/api/characters/3")
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		testutil.AssertField(t, data, "character_type", "linked")
		testutil.AssertField(t, data, "can_edit", true)
	})
}

// Linking a compendium spell copies its data into the character's spellbook and stores the ref.
func TestLinkCompendiumSpellCopiesData(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	testutil.SeedCompendiumSpell(t, 9001, "Test Spell")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/spells/link", LinkCompendiumSpell)
	})
	w := testutil.PostForm(t, r, "/api/characters/1/spells/link", map[string]string{"compendium_spell_id": "9001"})
	testutil.AssertStatus(t, w, 201)

	var name, school string
	var level int
	var ref any
	err := db.DB.QueryRow(
		"SELECT name, level, school, COALESCE(compendium_spell_id, 0) FROM spells WHERE character_id=1 AND compendium_spell_id=9001",
	).Scan(&name, &level, &school, &ref)
	if err != nil {
		t.Fatalf("linked spell row not found: %v", err)
	}
	if name != "Test Spell" || level != 1 || school != "Evocation" {
		t.Fatalf("data not copied: name=%q level=%d school=%q", name, level, school)
	}
	if fmt.Sprint(ref) != "9001" {
		t.Fatalf("compendium_spell_id not stored: %v", ref)
	}
}

// Linking a compendium equipment copies its data into the inventory and stores the ref.
func TestLinkCompendiumEquipmentCopiesData(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	testutil.SeedCompendiumEquipment(t, 9002, "Test Item")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/inventory/link", LinkCompendiumEquipment)
	})
	w := testutil.PostForm(t, r, "/api/characters/1/inventory/link", map[string]string{"compendium_equipment_id": "9002"})
	testutil.AssertStatus(t, w, 201)

	var name, category string
	var ref any
	err := db.DB.QueryRow(
		"SELECT name, category, COALESCE(compendium_equipment_id, 0) FROM inventory WHERE character_id=1 AND compendium_equipment_id=9002",
	).Scan(&name, &category, &ref)
	if err != nil {
		t.Fatalf("linked item row not found: %v", err)
	}
	if name != "Test Item" || category != "Adventuring Gear" {
		t.Fatalf("data not copied: name=%q category=%q", name, category)
	}
	if fmt.Sprint(ref) != "9002" {
		t.Fatalf("compendium_equipment_id not stored: %v", ref)
	}
}

// Unlinking keeps the character's spell/item (data preserved) but clears the compendium ref.
func TestUnlinkCompendiumPreservesData(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")
	testutil.SeedCompendiumSpell(t, 9001, "Test Spell")
	testutil.SeedCompendiumEquipment(t, 9002, "Test Item")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/characters/:id/spells/link", LinkCompendiumSpell)
		auth.DELETE("/characters/:id/spells/:spellId/link", UnlinkCompendiumSpell)
		auth.POST("/characters/:id/inventory/link", LinkCompendiumEquipment)
		auth.DELETE("/characters/:id/inventory/:itemId/link", UnlinkCompendiumEquipment)
	})

	testutil.PostForm(t, r, "/api/characters/1/spells/link", map[string]string{"compendium_spell_id": "9001"})
	testutil.PostForm(t, r, "/api/characters/1/inventory/link", map[string]string{"compendium_equipment_id": "9002"})

	// spell: unlink → row survives, ref NULL
	w := testutil.Delete(t, r, "/api/characters/1/spells/1/link")
	testutil.AssertStatus(t, w, 200)
	var spellCount int
	var spellRef any
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(compendium_spell_id, 0) FROM spells WHERE id=1").Scan(&spellCount, &spellRef)
	if spellCount != 1 {
		t.Fatalf("spell row deleted on unlink (want preserved): count=%d", spellCount)
	}
	if fmt.Sprint(spellRef) != "0" {
		t.Fatalf("spell compendium ref not cleared: %v", spellRef)
	}

	// item: unlink → row survives, ref NULL
	w = testutil.Delete(t, r, "/api/characters/1/inventory/1/link")
	testutil.AssertStatus(t, w, 200)
	var itemCount int
	var itemRef any
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(compendium_equipment_id, 0) FROM inventory WHERE id=1").Scan(&itemCount, &itemRef)
	if itemCount != 1 {
		t.Fatalf("item row deleted on unlink (want preserved): count=%d", itemCount)
	}
	if fmt.Sprint(itemRef) != "0" {
		t.Fatalf("item compendium ref not cleared: %v", itemRef)
	}
}

// Regression: deleting a character with child rows (spell, item, npc link, session) must succeed.
func TestDeleteCharacterWithChildren(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Hero", "Elf", "Wizard")

	for _, stmt := range []string{
		"INSERT INTO npcs(id, user_id, name, race, class) VALUES(1, 1, 'Bob', 'Human', 'Commoner')",
		"INSERT INTO spells(character_id, name, level, school) VALUES(1, 'Magic Missile', 1, 'Evocation')",
		"INSERT INTO inventory(character_id, name, quantity) VALUES(1, 'Sword', 1)",
		"INSERT INTO character_npcs(character_id, npc_id, relationship) VALUES(1, 1, 'ally')",
		"INSERT INTO sessions(character_id, title) VALUES(1, 'Session 1')",
	} {
		if _, err := db.DB.Exec(stmt); err != nil {
			t.Fatalf("seed child row: %v", err)
		}
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/characters/:id", DeleteCharacter)
	})
	w := testutil.Delete(t, r, "/api/characters/1")
	testutil.AssertStatus(t, w, 200)

	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM characters WHERE id=1").Scan(&count)
	if count != 0 {
		t.Fatalf("character still exists after delete")
	}
}
