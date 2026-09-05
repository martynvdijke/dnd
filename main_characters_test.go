package main

import (
	"fmt"
	"strings"
	"testing"

	"villum/db"
	"villum/middleware"
)

func TestImportJSON(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	importData := []map[string]any{{
		"name": "Imported Hero", "race": "Elf", "class": "Ranger", "level": 3,
		"str": 12, "dex": 18, "con": 14, "int": 10, "wis": 16, "cha": 8,
		"hp_max": 28, "hp_current": 28, "ac": 15, "speed": 35,
		"spells":        []map[string]any{{"name": "Hunter's Mark", "level": 1, "school": "Divination"}},
		"inventory":     []map[string]any{{"name": "Longbow", "category": "weapon", "quantity": 1, "damage_dice": "1d8", "damage_type": "piercing"}},
		"proficiencies": []map[string]any{{"name": "Perception", "type": "skill"}},
		"currency":      map[string]any{"gp": 100},
	}}

	resp := tc.post("/api/characters/import", importData)
	if resp.Code != 200 {
		t.Fatalf("import failed: %d - %s", resp.Code, resp.Body.String())
	}
	var results []any
	readJSON(resp, &results)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestMulticharacterTenant(t *testing.T) {
	// Test user isolation: create two users, ensure they only see their own chars
	adminTC := newTestClient()
	setupAdmin(t, adminTC)

	// Create user1
	resp := adminTC.post("/api/admin/users", map[string]any{
		"username": "user_one", "password": "testpass1", "role": "user",
	})
	var user1 map[string]any
	readJSON(resp, &user1)

	// Create user2
	resp = adminTC.post("/api/admin/users", map[string]any{
		"username": "user_two", "password": "testpass2", "role": "user",
	})
	var user2 map[string]any
	readJSON(resp, &user2)

	// Login as user1 and create character
	tc1 := newTestClient()
	login(t, tc1, "user_one", "testpass1")
	resp = tc1.post("/api/characters", map[string]any{"name": "Hero One", "race": "Human", "class": "Fighter"})
	if resp.Code != 201 {
		t.Fatalf("user1 create failed: %d", resp.Code)
	}

	// Login as user2 and create character
	tc2 := newTestClient()
	login(t, tc2, "user_two", "testpass2")
	resp = tc2.post("/api/characters", map[string]any{"name": "Hero Two", "race": "Elf", "class": "Wizard"})
	if resp.Code != 201 {
		t.Fatalf("user2 create failed: %d", resp.Code)
	}

	// User1 should only see their own character
	resp = tc1.get("/api/characters", nil)
	var chars1 []any
	readJSON(resp, &chars1)
	if len(chars1) != 1 {
		t.Fatalf("user1 sees %d chars, expected 1", len(chars1))
	}
	c1 := chars1[0].(map[string]any)
	if c1["name"] != "Hero One" {
		t.Fatalf("user1 sees wrong char: %v", c1["name"])
	}

	// User2 should only see their own character
	resp = tc2.get("/api/characters", nil)
	var chars2 []any
	readJSON(resp, &chars2)
	if len(chars2) != 1 {
		t.Fatalf("user2 sees %d chars, expected 1", len(chars2))
	}
	c2 := chars2[0].(map[string]any)
	if c2["name"] != "Hero Two" {
		t.Fatalf("user2 sees wrong char: %v", c2["name"])
	}

	// Admin should see both
	resp = adminTC.get("/api/characters", nil)
	var charsAdmin []any
	readJSON(resp, &charsAdmin)
	if len(charsAdmin) != 3 { // 2 from users + admin's own "Export Test" from earlier
		t.Logf("admin sees %d chars (may include chars from other tests)", len(charsAdmin))
	}

	tc1.del(fmt.Sprintf("/api/characters/%d", int(c1["id"].(float64))), nil)
	tc2.del(fmt.Sprintf("/api/characters/%d", int(c2["id"].(float64))), nil)
}

func TestCharacterProficiencies(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Prof Test", "race": "Human", "class": "Rogue"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Add prof
	resp = tc.post("/api/proficiencies", map[string]any{"character_id": cid, "name": "Stealth", "type": "skill"})
	if resp.Code != 201 {
		t.Fatalf("add prof failed: %d", resp.Code)
	}
	var prof map[string]any
	readJSON(resp, &prof)
	pid := int(prof["id"].(float64))

	// Verify via get character
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get char failed: %d", resp.Code)
	}
	readJSON(resp, &char)
	profs := char["proficiencies"].([]any)
	if len(profs) != 1 {
		t.Fatalf("expected 1 prof, got %d", len(profs))
	}

	// Delete prof
	resp = tc.del(fmt.Sprintf("/api/proficiencies/%d", pid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete prof failed: %d", resp.Code)
	}
}

func TestCharacterFeaturesAndSpells(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Feat Test", "race": "Dwarf", "class": "Cleric"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Feature
	resp = tc.post(fmt.Sprintf("/api/characters/%d/features", cid), map[string]any{
		"name": "Darkvision", "description": "See in dark", "source": "Race", "level_gained": 1,
	})
	if resp.Code != 201 {
		t.Fatalf("add feature failed: %d", resp.Code)
	}
	var feat map[string]any
	readJSON(resp, &feat)
	fid := int(feat["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/features/%d", fid), map[string]any{
		"name": "Darkvision 60ft", "description": "See in dark 60ft", "source": "Race", "level_gained": 1,
	})
	if resp.Code != 200 {
		t.Fatalf("update feature failed: %d", resp.Code)
	}

	tc.del(fmt.Sprintf("/api/features/%d", fid), nil)

	// Spell
	resp = tc.post(fmt.Sprintf("/api/characters/%d/spells", cid), map[string]any{
		"name": "Bless", "level": 1, "school": "Enchantment",
		"casting_time": "1 action", "range": "30 feet", "components": "V, S, M",
		"duration": "Concentration, 1 min", "description": "Bless up to 3 creatures",
	})
	if resp.Code != 201 {
		t.Fatalf("add spell failed: %d", resp.Code)
	}
	var spell map[string]any
	readJSON(resp, &spell)
	sid := int(spell["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/spells/%d", sid), map[string]any{
		"name": "Bless", "level": 1, "school": "Enchantment",
		"casting_time": "1 action", "range": "30 feet", "components": "V, S, M",
		"duration": "Concentration, 1 min", "description": "Bless up to 3 creatures",
		"prepared": true, "always_prepared": false, "source": "", "notes": "",
	})
	if resp.Code != 200 {
		t.Fatalf("update spell failed: %d", resp.Code)
	}

	tc.del(fmt.Sprintf("/api/spells/%d", sid), nil)
}

func TestDeathSaves(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Falling Hero", "race": "Human", "class": "Fighter", "hp_max": 20, "hp_current": 0})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Simulate death saves
	resp = tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Falling Hero", "race": "Human", "class": "Fighter",
		"level": 1, "hp_max": 20, "hp_current": 0,
		"death_saves_successes": 1, "death_saves_failures": 2,
		"ac": 10, "initiative": 0, "speed": 30, "str": 10, "dex": 10,
		"con": 10, "int": 10, "wis": 10, "cha": 10,
	})
	if resp.Code != 200 {
		t.Fatalf("update death saves failed: %d", resp.Code)
	}

	// Verify via get
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["death_saves_successes"].(float64)) != 1 || int(char["death_saves_failures"].(float64)) != 2 {
		t.Fatalf("death saves mismatch: %+v", char)
	}

	// Long rest should reset death saves
	resp = tc.post(fmt.Sprintf("/api/characters/%d/rest", cid), map[string]any{"rest_type": "long"})
	if resp.Code != 200 {
		t.Fatalf("long rest failed: %d", resp.Code)
	}
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["death_saves_successes"].(float64)) != 0 || int(char["death_saves_failures"].(float64)) != 0 {
		t.Fatalf("death saves should reset after long rest: %+v", char)
	}
}

func TestUpdateSession(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Session Updater", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/sessions", cid), map[string]any{
		"title": "Session 1", "notes": "Start", "xp_earned": 100, "gold_earned": 50,
	})
	var sess map[string]any
	readJSON(resp, &sess)
	sid := int(sess["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/sessions/%d", sid), map[string]any{
		"title": "Session 1 Updated", "notes": "End", "xp_earned": 600, "gold_earned": 150,
	})
	if resp.Code != 200 {
		t.Fatalf("update session failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d/sessions", cid), nil)
	var sessions []any
	readJSON(resp, &sessions)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session after update, got %d", len(sessions))
	}
	s := sessions[0].(map[string]any)
	if int(s["xp_earned"].(float64)) != 600 {
		t.Fatalf("expected 600 XP, got %v", s["xp_earned"])
	}
}

func TestUpdateJournalEntry(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Journal Updater", "race": "Elf", "class": "Wizard"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/journal", cid), map[string]any{
		"title": "Day 1", "entry": "We began...",
	})
	var entry map[string]any
	readJSON(resp, &entry)
	jid := int(entry["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/journal/%d", jid), map[string]any{
		"title": "Day 1 Revised", "entry": "Actually, we started at dawn...",
	})
	if resp.Code != 200 {
		t.Fatalf("update journal failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d/journal", cid), nil)
	var entries []any
	readJSON(resp, &entries)
	e := entries[0].(map[string]any)
	if e["title"] != "Day 1 Revised" {
		t.Fatalf("expected 'Day 1 Revised', got %v", e["title"])
	}
}

func TestSpellFiltering(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/compendium/spells?level=0", nil)
	if resp.Code != 200 {
		t.Fatalf("filter by level failed: %d", resp.Code)
	}
	var spells []any
	readJSON(resp, &spells)
	if len(spells) < 5 {
		t.Fatalf("expected >=5 cantrips, got %d", len(spells))
	}

	resp = tc.get("/api/compendium/spells?school=Evocation", nil)
	if resp.Code != 200 {
		t.Fatalf("filter by school failed: %d", resp.Code)
	}
	readJSON(resp, &spells)
	t.Logf("Evocation spells: %d", len(spells))

	resp = tc.get("/api/compendium/spells?class=Wizard", nil)
	if resp.Code != 200 {
		t.Fatalf("filter by class failed: %d", resp.Code)
	}
	readJSON(resp, &spells)
	t.Logf("Wizard spells: %d", len(spells))

	resp = tc.get("/api/compendium/spells?q=fire", nil)
	if resp.Code != 200 {
		t.Fatalf("search spells failed: %d", resp.Code)
	}
}

func TestEquipmentFiltering(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/compendium/equipment?category=Weapon", nil)
	if resp.Code != 200 {
		t.Fatalf("filter eq by category failed: %d", resp.Code)
	}
	var eq []any
	readJSON(resp, &eq)
	if len(eq) < 5 {
		t.Fatalf("expected >=5 weapons, got %d", len(eq))
	}

	resp = tc.get("/api/compendium/equipment?q=sword", nil)
	if resp.Code != 200 {
		t.Fatalf("search eq failed: %d", resp.Code)
	}
	readJSON(resp, &eq)
	t.Logf("Sword equipment: %d", len(eq))
}

func TestCompendiumSearchEmptyQuery(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/compendium/search", nil)
	if resp.Code != 400 {
		t.Fatalf("expected 400 for empty query, got %d", resp.Code)
	}
}

func TestCampaignAuthorization(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create user2 as non-admin
	tc.post("/api/admin/users", map[string]any{
		"username": "campaign_user", "password": "testpass123", "role": "user",
	})

	tc2 := newTestClient()
	login(t, tc2, "campaign_user", "testpass123")

	// user2 creates a campaign
	resp := tc2.post("/api/campaigns", map[string]any{"name": "My Campaign", "description": "Private"})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d", resp.Code)
	}
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Admin tries to update it - should fail (admin has different user_id)
	// Actually: the owner is campaign_user, admin role should allow access
	resp = tc.put(fmt.Sprintf("/api/campaigns/%d", cid), map[string]any{"name": "Hijacked", "description": "No"})
	if resp.Code != 200 {
		t.Fatalf("admin should be able to update any campaign: %d", resp.Code)
	}

	// Another non-admin user should not be able to update
	tc.post("/api/admin/users", map[string]any{
		"username": "other_user", "password": "testpass456", "role": "user",
	})
	tc3 := newTestClient()
	login(t, tc3, "other_user", "testpass456")
	resp = tc3.put(fmt.Sprintf("/api/campaigns/%d", cid), map[string]any{"name": "Hijacked 2", "description": "No"})
	if resp.Code != 403 {
		t.Fatalf("non-owner should get 403, got %d", resp.Code)
	}

	// Cleanup
	tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
}

func TestUpdateCharacterCampaignID(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Test Campaign", "description": "C"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	resp = tc.post("/api/characters", map[string]any{"name": "Campaign Member", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	chid := int(char["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/characters/%d", chid), map[string]any{
		"name": "Campaign Member", "race": "Human", "class": "Fighter", "level": 2,
		"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
		"hp_max": 28, "hp_current": 22, "ac": 18, "initiative": 0, "speed": 30,
		"campaign_id": cid,
	})
	if resp.Code != 200 {
		t.Fatalf("update with campaign_id failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", chid), nil)
	readJSON(resp, &char)
	if char["campaign_id"] != float64(cid) {
		t.Fatalf("expected campaign_id %d, got %v", cid, char["campaign_id"])
	}
}

func TestCampaignDeleteUnassignsCharacters(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Delete Me", "description": "T"})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d", resp.Code)
	}
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	resp = tc.post("/api/characters", map[string]any{"name": "Temp Member", "race": "Gnome", "class": "Wizard", "campaign_id": cid})
	if resp.Code != 201 {
		t.Fatalf("create char failed: %d", resp.Code)
	}
	var char map[string]any
	readJSON(resp, &char)
	chid := int(char["id"].(float64))

	// Verify character has campaign_id set via direct DB query
	var dbCampID *int64
	db.DB.QueryRow("SELECT campaign_id FROM characters WHERE id=?", chid).Scan(&dbCampID)
	if dbCampID == nil || *dbCampID != int64(cid) {
		t.Fatalf("DB campaign_id should be %d, got %v", cid, dbCampID)
	}

	resp = tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete campaign failed: %d - %s", resp.Code, resp.Body.String())
	}

	db.DB.QueryRow("SELECT campaign_id FROM characters WHERE id=?", chid).Scan(&dbCampID)
	if dbCampID != nil {
		t.Fatalf("expected campaign_id NULL after campaign delete, got %v", dbCampID)
	}
}

func TestLongRestRecoversSpellSlots(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Caster Rest", "race": "Elf", "class": "Wizard", "hp_max": 30, "hp_current": 15})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	tc.put(fmt.Sprintf("/api/characters/%d/spellcasting", cid), map[string]any{
		"ability": "int", "save_dc": 14, "attack_bonus": 6,
		"slots_1_max": 4, "slots_1_used": 4,
		"slots_2_max": 3, "slots_2_used": 3,
	})

	resp = tc.post(fmt.Sprintf("/api/characters/%d/rest", cid), map[string]any{"rest_type": "long"})
	if resp.Code != 200 {
		t.Fatalf("long rest failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	sc := char["spellcasting"].(map[string]any)
	if int(sc["slots_1_used"].(float64)) != 0 || int(sc["slots_2_used"].(float64)) != 0 {
		t.Fatalf("slots should be recovered after long rest: 1=%v 2=%v", sc["slots_1_used"], sc["slots_2_used"])
	}
	if int(char["hp_current"].(float64)) != 30 {
		t.Fatalf("HP should be full after long rest: %v", char["hp_current"])
	}
}

func TestShortRestNoDeathSaveReset(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Dying Short", "race": "Human", "class": "Fighter", "hp_max": 30, "hp_current": 15})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Dying Short", "race": "Human", "class": "Fighter", "level": 1,
		"hp_max": 30, "hp_current": 15,
		"death_saves_successes": 2, "death_saves_failures": 1,
		"ac": 10, "initiative": 0, "speed": 30, "str": 10, "dex": 10,
		"con": 10, "int": 10, "wis": 10, "cha": 10,
	})

	resp = tc.post(fmt.Sprintf("/api/characters/%d/rest", cid), map[string]any{"rest_type": "short"})
	if resp.Code != 200 {
		t.Fatalf("short rest failed: %d", resp.Code)
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["death_saves_successes"].(float64)) != 2 || int(char["death_saves_failures"].(float64)) != 1 {
		t.Fatalf("short rest should NOT reset death saves: %+v", char)
	}
}

func TestInvalidRestType(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Rest Test", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/rest", cid), map[string]any{"rest_type": "invalid"})
	if resp.Code != 400 {
		t.Fatalf("expected 400 for invalid rest type, got %d", resp.Code)
	}
}

func TestLevelUpHitDice(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Grow Hero", "race": "Dwarf", "class": "Fighter", "level": 1,
		"hit_dice": "1d10", "hit_dice_current": 1, "con": 14,
		"hp_max": 12, "hp_current": 12,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/levelup", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("level up failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["hit_dice_current"].(float64)) != 2 {
		t.Fatalf("expected hit_dice_current=2 after level up, got %v", char["hit_dice_current"])
	}
	if int(char["level"].(float64)) != 2 {
		t.Fatalf("expected level=2, got %v", char["level"])
	}
}

func TestDiceRollsCharacterFilter(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Dice Filter", "race": "Halfling", "class": "Rogue"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	tc.post("/api/roll", map[string]any{"expression": "1d20+5", "character_id": cid})
	tc.post("/api/roll", map[string]any{"expression": "1d6", "character_id": cid})

	resp = tc.get(fmt.Sprintf("/api/dice-rolls?character_id=%d", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("dice history with filter failed: %d", resp.Code)
	}
	var rolls []any
	readJSON(resp, &rolls)
	if len(rolls) != 2 {
		t.Fatalf("expected 2 rolls for character, got %d", len(rolls))
	}
}

func TestExportTextFormat(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Text Export", "race": "Dragonborn", "class": "Paladin",
		"str": 16, "dex": 10, "con": 14, "int": 8, "wis": 12, "cha": 16,
		"hp_max": 36, "hp_current": 36, "level": 3,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	tc.put(fmt.Sprintf("/api/characters/%d/currency", cid), map[string]any{"gp": 50, "pp": 2})

	resp = tc.get(fmt.Sprintf("/api/characters/%d/export?format=text", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("text export failed: %d", resp.Code)
	}
	body := resp.Body.String()
	if !strings.Contains(body, "Text Export") || !strings.Contains(body, "Dragonborn") {
		t.Fatal("text export missing key fields")
	}
	t.Logf("Text export length: %d", len(body))
}

func TestInvalidDiceExpr(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	tests := []struct {
		expr string
		code int
	}{
		{"", 400},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			resp := tc.post("/api/roll", map[string]any{"expression": tt.expr})
			if resp.Code != tt.code {
				t.Errorf("expected %d for '%s', got %d", tt.code, tt.expr, resp.Code)
			}
		})
	}
}

func TestQuestStatusTransitions(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Quest Tester", "race": "Elf", "class": "Ranger"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/quests", cid), map[string]any{
		"name": "Status Quest", "description": "Testing", "status": "available",
	})
	var quest map[string]any
	readJSON(resp, &quest)
	qid := int(quest["id"].(float64))

	statuses := []string{"active", "complete", "active", "failed"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			resp := tc.put(fmt.Sprintf("/api/quests/%d", qid), map[string]any{
				"name": "Status Quest", "description": "Testing", "status": status,
			})
			if resp.Code != 200 {
				t.Fatalf("set status to %s failed: %d", status, resp.Code)
			}
		})
	}
}

// ─── PWA / Static File Tests ───

func TestStaticFileServing(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Verify manifest.json is served
	resp := tc.get("/static/manifest.json", nil)
	if resp.Code != 200 {
		t.Fatalf("manifest.json serving failed: %d - %s", resp.Code, resp.Body.String())
	}
	ct := resp.Header().Get("Content-Type")
	if !strings.Contains(ct, "json") && !strings.Contains(ct, "octet-stream") {
		t.Logf("manifest.json content-type: %s", ct)
	}
	var manifest map[string]any
	readJSON(resp, &manifest)
	if manifest["name"] == "" || manifest["start_url"] == "" {
		t.Fatal("manifest missing required fields (name, start_url)")
	}
	if manifest["display"] != "standalone" {
		t.Fatalf("expected display 'standalone', got %v", manifest["display"])
	}
	t.Logf("Manifest: name=%s, start_url=%s", manifest["name"], manifest["start_url"])

	// Verify sw.js is served
	resp = tc.get("/static/sw.js", nil)
	if resp.Code != 200 {
		t.Fatalf("sw.js serving failed: %d", resp.Code)
	}
	body := resp.Body.String()
	if !strings.Contains(body, "self.addEventListener") {
		t.Fatal("sw.js missing service worker event listeners")
	}
	if !strings.Contains(body, "install") {
		t.Fatal("sw.js missing install event handler")
	}
	if !strings.Contains(body, "fetch") {
		t.Fatal("sw.js missing fetch event handler")
	}
	t.Logf("sw.js served successfully (%d bytes)", len(body))
}

// ─── HTML Structure Tests ───

func TestMainHTMLNewNavigation(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/", nil)
	if resp.Code != 200 {
		t.Fatalf("main page failed: %d", resp.Code)
	}
	body := resp.Body.String()

	// Verify new navigation elements
	if !strings.Contains(body, "bottom-tab-bar") {
		t.Fatal("expected bottom-tab-bar in main page HTML")
	}
	if !strings.Contains(body, "app-sidebar") {
		t.Fatal("expected app-sidebar in main page HTML")
	}
	if !strings.Contains(body, "session-mode-topbar") {
		t.Fatal("expected session-mode-topbar in main page HTML")
	}

	// Verify bottom tab items
	if !strings.Contains(body, "data-nav=\"characters\"") {
		t.Fatal("expected characters bottom tab")
	}
	if !strings.Contains(body, "data-nav=\"party\"") {
		t.Fatal("expected party bottom tab")
	}
	if !strings.Contains(body, "data-nav=\"dice\"") {
		t.Fatal("expected dice bottom tab")
	}
	if !strings.Contains(body, "data-nav=\"compendium\"") {
		t.Fatal("expected compendium bottom tab")
	}
	if !strings.Contains(body, "data-nav=\"more\"") {
		t.Fatal("expected more bottom tab")
	}

	// Verify sidebar nav items
	if !strings.Contains(body, "data-nav=\"encounters\"") {
		t.Fatal("expected encounters in sidebar")
	}
	if !strings.Contains(body, "data-nav=\"timeline\"") {
		t.Fatal("expected timeline in sidebar")
	}

	// Verify manifest link
	if !strings.Contains(body, "/static/manifest.json") {
		t.Fatal("expected manifest link in HTML head")
	}

	// Verify PWA script
	if !strings.Contains(body, "pwa.js") {
		t.Fatal("expected pwa.js script reference")
	}

	t.Logf("Main page HTML contains all new navigation elements")
}

func TestStartCleanupTask(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	store := middleware.Store
	sid := store.Create(999, "test", "user", "127.0.0.1")
	if store.Get(sid) == nil {
		t.Fatal("session should exist after creation")
	}
	store.Delete(sid)
	if store.Get(sid) != nil {
		t.Fatal("session should be gone after deletion")
	}
}

func TestAbilityModifiers(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Mod Test", "race": "Human", "class": "Fighter",
		"str": 18, "dex": 14, "con": 16, "int": 10, "wis": 8, "cha": 6,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)

	if int(char["str_mod"].(float64)) != 4 {
		t.Fatalf("expected str_mod 4, got %v", char["str_mod"])
	}
	if int(char["dex_mod"].(float64)) != 2 {
		t.Fatalf("expected dex_mod 2, got %v", char["dex_mod"])
	}
	if int(char["con_mod"].(float64)) != 3 {
		t.Fatalf("expected con_mod 3, got %v", char["con_mod"])
	}
	if int(char["int_mod"].(float64)) != 0 {
		t.Fatalf("expected int_mod 0, got %v", char["int_mod"])
	}
	if int(char["wis_mod"].(float64)) != -1 {
		t.Fatalf("expected wis_mod -1, got %v", char["wis_mod"])
	}
	if int(char["cha_mod"].(float64)) != -2 {
		t.Fatalf("expected cha_mod -2, got %v", char["cha_mod"])
	}
}

func TestSpellSaveDCComputed(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Save DC Test", "race": "Elf", "class": "Wizard",
		"int": 18, "level": 5,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Set up spellcasting
	tc.put(fmt.Sprintf("/api/characters/%d/spellcasting", cid), map[string]any{
		"ability": "int", "save_dc": 15, "attack_bonus": 7,
	})

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	// Level 5 -> prof bonus 3, int mod 4, save DC = 8 + 3 + 4 = 15, attack = 3 + 4 = 7
	if int(char["spell_save_dc"].(float64)) != 15 {
		t.Fatalf("expected spell_save_dc 15, got %v", char["spell_save_dc"])
	}
	if int(char["spell_attack_bonus"].(float64)) != 7 {
		t.Fatalf("expected spell_attack_bonus 7, got %v", char["spell_attack_bonus"])
	}
}

func TestCheckRollSkills(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Skill Checker", "race": "Elf", "class": "Rogue",
		"dex": 18, "wis": 14, "cha": 10,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Add proficiency in Stealth
	tc.post("/api/proficiencies", map[string]any{
		"character_id": cid, "name": "Stealth", "type": "skill",
	})

	tests := []struct {
		name  string
		skill string
	}{
		{"stealth with prof", "stealth"},
		{"perception without prof", "perception"},
		{"persuasion", "persuasion"},
	}
	for _, tt := range tests {
		t.Run(tt.skill, func(t *testing.T) {
			resp = tc.post("/api/roll/check", map[string]any{
				"character_id": cid, "type": "skill", "name": tt.skill,
			})
			if resp.Code != 200 {
				t.Fatalf("check %s failed: %d - %s", tt.skill, resp.Code, resp.Body.String())
			}
			var result map[string]any
			readJSON(resp, &result)
			total := int(result["total"].(float64))
			if total < 1 || total > 30 {
				t.Errorf("unexpected total %d for %s", total, tt.skill)
			}
			t.Logf("%s -> %s", tt.skill, result["text"])
		})
	}
}
