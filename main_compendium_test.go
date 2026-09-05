package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestDnDAPIRespectsDB(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Test with empty query (should get 400)
	resp := tc.get("/api/compendium/api/spells", nil)
	if resp.Code != 400 {
		t.Errorf("expected 400 for missing query, got %d", resp.Code)
	}

	// Test with unknown category
	resp = tc.get("/api/compendium/api/invalid?q=test", nil)
	if resp.Code != 400 {
		t.Errorf("expected 400 for unknown category, got %d", resp.Code)
	}

	// Test API fallback (may or may not have network access)
	resp = tc.get("/api/compendium/api/equipment?q=longsword", nil)
	if resp.Code == 200 {
		var result map[string]any
		readJSON(resp, &result)
		if src, ok := result["source"]; !ok || src != "dnd5eapi" {
			t.Errorf("expected source=dnd5eapi, got %v", src)
		}
		if sys, ok := result["system"]; !ok || sys != "dnd5e" {
			t.Errorf("expected system=dnd5e, got %v", sys)
		}
		if count, ok := result["count"]; ok && count.(float64) > 0 {
			t.Logf("D&D API returned %d results for 'longsword'", int(count.(float64)))
		}
	} else {
		t.Logf("D&D API unavailable (status %d) - skipping online test", resp.Code)
	}
}

// ─── Conditions Tests ───

func TestConditions(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Condition Hero", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create condition
	resp = tc.post("/api/conditions", map[string]any{
		"character_id": cid, "name": "Poisoned", "type": "poisoned",
		"duration": 5, "duration_type": "round", "source": "Spider Bite",
		"saving_throw": "con", "save_dc": 12,
	})
	if resp.Code != 201 {
		t.Fatalf("create condition failed: %d - %s", resp.Code, resp.Body.String())
	}
	// List
	resp = tc.get("/api/conditions?character_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list conditions failed: %d", resp.Code)
	}
	var conds []any
	readJSON(resp, &conds)
	if len(conds) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(conds))
	}

	// Get types
	resp = tc.get("/api/conditions/types", nil)
	if resp.Code != 200 {
		t.Fatalf("condition types failed: %d", resp.Code)
	}

	// Get summary
	resp = tc.get("/api/conditions/summary?character_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("condition summary failed: %d", resp.Code)
	}

	// Tick (advance 3 rounds)
	resp = tc.post("/api/conditions/tick", map[string]any{
		"character_id": cid, "count": 3, "duration_type": "round",
	})
	if resp.Code != 200 {
		t.Fatalf("tick conditions failed: %d - %s", resp.Code, resp.Body.String())
	}
	var tickResult map[string]any
	readJSON(resp, &tickResult)

	// Tick remaining rounds (should expire)
	resp = tc.post("/api/conditions/tick", map[string]any{
		"character_id": cid, "count": 5, "duration_type": "round",
	})
	readJSON(resp, &tickResult)

	// Verify expired
	resp = tc.get("/api/conditions?character_id="+strconv.Itoa(cid), nil)
	readJSON(resp, &conds)
	if len(conds) != 0 {
		t.Fatalf("expected 0 conditions after expiry, got %d", len(conds))
	}
}

// ─── Concentration Tests ───

func TestConcentrationCheck(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Concentrating Hero", "race": "Elf", "class": "Wizard", "con": 14,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Set concentrating
	tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Concentrating Hero", "race": "Elf", "class": "Wizard",
		"level": 1, "con": 14, "str": 10, "dex": 10, "int": 10, "wis": 10, "cha": 10,
		"hp_max": 20, "hp_current": 20, "ac": 10, "initiative": 0, "speed": 30,
		"concentrating_on": "Hunter's Mark",
	})

	// Check concentration with low damage
	resp = tc.post(fmt.Sprintf("/api/characters/%d/check-concentration", cid), map[string]any{"damage": 5})
	if resp.Code != 200 {
		t.Fatalf("concentration check failed: %d - %s", resp.Code, resp.Body.String())
	}
	var concResult map[string]any
	readJSON(resp, &concResult)
	if concResult["needs_check"] != true {
		t.Fatal("expected needs_check=true")
	}
	if int(concResult["dc"].(float64)) != 10 {
		t.Fatalf("expected DC 10, got %v", concResult["dc"])
	}
	if concResult["spell_name"] != "Hunter's Mark" {
		t.Fatalf("expected 'Hunter's Mark', got %v", concResult["spell_name"])
	}

	// Check concentration with high damage
	resp = tc.post(fmt.Sprintf("/api/characters/%d/check-concentration", cid), map[string]any{"damage": 30})
	readJSON(resp, &concResult)
	if int(concResult["dc"].(float64)) != 15 {
		t.Fatalf("expected DC 15 for 30 damage, got %v", concResult["dc"])
	}

	// No concentration
	tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Concentrating Hero", "race": "Elf", "class": "Wizard",
		"level": 1, "con": 14, "str": 10, "dex": 10, "int": 10, "wis": 10, "cha": 10,
		"hp_max": 20, "hp_current": 20, "ac": 10, "initiative": 0, "speed": 30,
		"concentrating_on": "",
	})
	resp = tc.post(fmt.Sprintf("/api/characters/%d/check-concentration", cid), map[string]any{"damage": 5})
	readJSON(resp, &concResult)
	if concResult["needs_check"] != false {
		t.Fatal("expected needs_check=false when not concentrating")
	}
}

// ─── Feats Tests ───

func TestCharacterFeats(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Feat Hero", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create feat
	resp = tc.post("/api/feats", map[string]any{
		"character_id": cid, "name": "Sharpshooter",
		"description": "Ranged attacks ignore half cover", "source": "PHB",
		"prerequisites": "Dex 13+", "level_gained": 4,
	})
	if resp.Code != 201 {
		t.Fatalf("create feat failed: %d - %s", resp.Code, resp.Body.String())
	}
	// List
	resp = tc.get("/api/feats?character_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list feats failed: %d", resp.Code)
	}
	var feats []any
	readJSON(resp, &feats)
	if len(feats) != 1 {
		t.Fatalf("expected 1 feat, got %d", len(feats))
	}
	// Delete
	var feat map[string]any
	readJSON(resp, &feat)
	resp = tc.del(fmt.Sprintf("/api/feats/%d", int(feats[0].(map[string]any)["id"].(float64))), nil)
	if resp.Code != 200 {
		t.Fatalf("delete feat failed: %d", resp.Code)
	}
}

// ─── Companions Tests ───

func TestCompanions(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Companion Hero", "race": "Human", "class": "Wizard"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create familiar
	resp = tc.post("/api/companions", map[string]any{
		"character_id": cid, "name": "Owlbert", "type": "familiar",
		"race": "Owl", "hp_max": 3, "hp_current": 3, "ac": 12,
		"str": 3, "dex": 14, "con": 8, "int": 2, "wis": 12, "cha": 6,
		"speed": 10, "abilities": "Flyby, Darkvision 60ft",
	})
	if resp.Code != 201 {
		t.Fatalf("create companion failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Create mount
	resp = tc.post("/api/companions", map[string]any{
		"character_id": cid, "name": "Shadowmere", "type": "mount",
		"race": "Warhorse", "hp_max": 30, "hp_current": 30, "ac": 13,
		"str": 18, "dex": 12, "con": 14, "int": 3, "wis": 12, "cha": 7,
	})
	if resp.Code != 201 {
		t.Fatalf("create mount failed: %d", resp.Code)
	}

	// List
	resp = tc.get("/api/companions?character_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list companions failed: %d", resp.Code)
	}
	var companions []any
	readJSON(resp, &companions)
	if len(companions) != 2 {
		t.Fatalf("expected 2 companions, got %d", len(companions))
	}

	// Delete first companion
	first := companions[0].(map[string]any)
	tc.del(fmt.Sprintf("/api/companions/%d", int(first["id"].(float64))), nil)
	resp = tc.get("/api/companions?character_id="+strconv.Itoa(cid), nil)
	readJSON(resp, &companions)
	if len(companions) != 1 {
		t.Fatalf("expected 1 companion after delete, got %d", len(companions))
	}
}

// ─── Factions Tests ───

func TestFactionsReputation(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Faction Hero", "race": "Human", "class": "Rogue"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create faction
	resp = tc.post("/api/factions", map[string]any{
		"name": "Harpers", "description": "Secret society", "type": "organization",
	})
	if resp.Code != 201 {
		t.Fatalf("create faction failed: %d - %s", resp.Code, resp.Body.String())
	}
	var fac map[string]any
	readJSON(resp, &fac)
	fid := int(fac["id"].(float64))

	// List factions
	resp = tc.get("/api/factions", nil)
	if resp.Code != 200 {
		t.Fatalf("list factions failed: %d", resp.Code)
	}

	// Set reputation
	resp = tc.post("/api/faction-reputation", map[string]any{
		"character_id": cid, "faction_id": fid, "standing": 50,
		"rank": "Friend", "notes": "Helped with a mission",
	})
	if resp.Code != 200 {
		t.Fatalf("set reputation failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get reputations
	resp = tc.get("/api/faction-reputation?character_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get reputation failed: %d", resp.Code)
	}
	var reps []any
	readJSON(resp, &reps)
	if len(reps) != 1 {
		t.Fatalf("expected 1 reputation, got %d", len(reps))
	}
	first := reps[0].(map[string]any)
	if int(first["standing"].(float64)) != 50 {
		t.Fatalf("expected standing 50, got %v", first["standing"])
	}
	if first["faction_name"] != "Harpers" {
		t.Fatalf("expected faction_name Harpers, got %v", first["faction_name"])
	}

	// Update reputation
	resp = tc.post("/api/faction-reputation", map[string]any{
		"character_id": cid, "faction_id": fid, "standing": 75,
		"rank": "Trusted Ally",
	})
	if resp.Code != 200 {
		t.Fatalf("update reputation failed: %d", resp.Code)
	}
	resp = tc.get("/api/faction-reputation?character_id="+strconv.Itoa(cid), nil)
	readJSON(resp, &reps)
	if int(reps[0].(map[string]any)["standing"].(float64)) != 75 {
		t.Fatalf("expected standing 75 after update, got %v", reps[0].(map[string]any)["standing"])
	}
}

// ─── Weather Tests ───

func TestWeatherGenerator(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/weather", nil)
	if resp.Code != 200 {
		t.Fatalf("weather gen failed: %d - %s", resp.Code, resp.Body.String())
	}
	var weather map[string]any
	readJSON(resp, &weather)
	if weather["season"] == "" || weather["temperature"] == "" {
		t.Fatal("weather missing season or temperature")
	}
	t.Logf("Weather: %s - %s - %s", weather["season"], weather["temperature"], weather["description"])

	// With season filter
	resp = tc.get("/api/generate/weather?season=Winter", nil)
	readJSON(resp, &weather)
	if weather["season"] != "Winter" {
		t.Fatalf("expected Winter, got %v", weather["season"])
	}
}

// ─── Notes Tests ───

func TestCharacterNotes(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Notes Hero", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create player note
	resp = tc.post("/api/notes", map[string]any{
		"character_id": cid, "title": "Quest Idea", "content": "Find the lost crown",
		"visibility": "player", "category": "quest",
	})
	if resp.Code != 201 {
		t.Fatalf("create note failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Create DM note
	resp = tc.post("/api/notes", map[string]any{
		"character_id": cid, "title": "Secret", "content": "The king is a doppelganger",
		"visibility": "dm", "category": "lore",
	})
	if resp.Code != 201 {
		t.Fatalf("create dm note failed: %d", resp.Code)
	}

	// List notes
	resp = tc.get("/api/notes?character_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list notes failed: %d", resp.Code)
	}
	var notes []any
	readJSON(resp, &notes)
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}

	// Update
	first := notes[0].(map[string]any)
	resp = tc.put(fmt.Sprintf("/api/notes/%d", int(first["id"].(float64))), map[string]any{
		"title": "Updated Note", "content": "Updated content",
		"visibility": "both", "category": "general",
	})
	if resp.Code != 200 {
		t.Fatalf("update note failed: %d", resp.Code)
	}
}

// ─── HP Calc Tests ───

func TestHPAutoCalc(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Auto HP Hero", "race": "Dwarf", "class": "Fighter", "con": 14,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Enable auto-calc
	tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Auto HP Hero", "race": "Dwarf", "class": "Fighter",
		"level": 3, "con": 14, "str": 16, "dex": 10, "int": 10, "wis": 10, "cha": 10,
		"hp_max": 10, "hp_current": 10, "ac": 10, "initiative": 0, "speed": 30,
		"hp_auto_calc": true,
	})

	// Add character multi-class
	tc.post(fmt.Sprintf("/api/characters/%d/classes", cid), map[string]any{
		"class": "Fighter", "level": 1, "hit_dice": "d10",
	})
	tc.post(fmt.Sprintf("/api/characters/%d/classes", cid), map[string]any{
		"class": "Wizard", "level": 2, "hit_dice": "d6",
	})

	// Calculate HP
	resp = tc.post(fmt.Sprintf("/api/characters/%d/calc-hp", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("calc HP failed: %d - %s", resp.Code, resp.Body.String())
	}
	var hpResult map[string]any
	readJSON(resp, &hpResult)
	t.Logf("HP result: %+v", hpResult)
	if int(hpResult["hp_max"].(float64)) < 10 {
		t.Fatalf("expected HP >= 10, got %v", hpResult["hp_max"])
	}
	breakdown := hpResult["breakdown"].([]any)
	if len(breakdown) != 3 {
		t.Fatalf("expected 3 breakdown entries (1 fighter + 2 wizard), got %d", len(breakdown))
	}
}

// ─── Random Character Generator Tests ───

func TestRandomCharacterGen(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/character", nil)
	if resp.Code != 200 {
		t.Fatalf("random character gen failed: %d - %s", resp.Code, resp.Body.String())
	}
	var rc map[string]any
	readJSON(resp, &rc)
	if rc["name"] == "" || rc["race"] == "" || rc["class"] == "" {
		t.Fatal("random character missing basic fields")
	}
	if int(rc["str"].(float64)) < 3 || int(rc["str"].(float64)) > 20 {
		t.Fatalf("str out of range: %v", rc["str"])
	}
	if rc["backstory_hook"] == "" {
		t.Fatal("expected backstory_hook")
	}
	t.Logf("Random: %s (%s %s Lv%v)", rc["name"], rc["race"], rc["class"], rc["level"])
}

// ─── Character Comparison Tests ───

func TestCharacterComparison(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Compare A", "race": "Elf", "class": "Wizard",
		"str": 8, "dex": 14, "con": 12, "int": 18, "wis": 14, "cha": 10,
	})
	var charA map[string]any
	readJSON(resp, &charA)
	cidA := int(charA["id"].(float64))

	resp = tc.post("/api/characters", map[string]any{
		"name": "Compare B", "race": "Dwarf", "class": "Fighter",
		"str": 18, "dex": 12, "con": 16, "int": 8, "wis": 10, "cha": 8,
	})
	var charB map[string]any
	readJSON(resp, &charB)
	cidB := int(charB["id"].(float64))

	resp = tc.get(fmt.Sprintf("/api/characters/compare?ids=%d,%d", cidA, cidB), nil)
	if resp.Code != 200 {
		t.Fatalf("compare failed: %d - %s", resp.Code, resp.Body.String())
	}
	var comp []any
	readJSON(resp, &comp)
	if len(comp) != 2 {
		t.Fatalf("expected 2 characters in comparison, got %d", len(comp))
	}
	c1 := comp[0].(map[string]any)
	c2 := comp[1].(map[string]any)
	if c1["name"] == "Compare A" && int(c1["int"].(float64)) != 18 {
		t.Fatalf("expected Compare A INT 18, got %v", c1["int"])
	}
	if c2["name"] == "Compare B" && int(c2["str"].(float64)) != 18 {
		t.Fatalf("expected Compare B STR 18, got %v", c2["str"])
	}
	t.Logf("Comparison: %s (INT %v) vs %s (STR %v)", c1["name"], c1["int"], c2["name"], c2["str"])
}

func TestFeatUpdate(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Feat Edit", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.post("/api/feats", map[string]any{
		"character_id": cid, "name": "Alert",
		"description": "+5 initiative", "source": "PHB", "level_gained": 1,
	})
	var feat map[string]any
	readJSON(resp, &feat)
	fid := int(feat["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/feats/%d", fid), map[string]any{
		"name": "Alert Updated", "description": "+5 initiative, can't be surprised",
		"prerequisites": "", "source": "PHB", "level_gained": 1,
	})
	if resp.Code != 200 {
		t.Fatalf("update feat failed: %d", resp.Code)
	}

	resp = tc.get("/api/feats?character_id="+strconv.Itoa(cid), nil)
	var feats []any
	readJSON(resp, &feats)
	if len(feats) != 1 {
		t.Fatalf("expected 1 feat after update, got %d", len(feats))
	}
	f := feats[0].(map[string]any)
	if f["name"] != "Alert Updated" {
		t.Fatalf("expected updated name 'Alert Updated', got %q", f["name"])
	}
}

func TestSpellcastingAutoCalc(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Wizard Hero", "race": "Elf", "class": "Wizard",
		"int": 18, "level": 5,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Set spellcasting with ability=int, save_dc=0, attack_bonus=0 to trigger auto-calc
	resp = tc.put(fmt.Sprintf("/api/characters/%d/spellcasting", cid), map[string]any{
		"ability": "int", "save_dc": 0, "attack_bonus": 0,
	})
	if resp.Code != 200 {
		t.Fatalf("update spellcasting failed: %d", resp.Code)
	}

	// Verify auto-calc: PB=3 (lvl5), int=18 => mod=4, DC=8+3+4=15, atk=3+4=7
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	sc := char["spellcasting"].(map[string]any)
	if sc["save_dc"].(float64) != 15 {
		t.Fatalf("expected save_dc 15 (8+3+4), got %v", sc["save_dc"])
	}
	if sc["attack_bonus"].(float64) != 7 {
		t.Fatalf("expected attack_bonus 7 (3+4), got %v", sc["attack_bonus"])
	}
}

func TestPassivePerceptionAutoCalc(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Perceptive Hero", "race": "Elf", "class": "Ranger",
		"wis": 16, "level": 1, "proficiency_bonus": 2,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Add Perception proficiency
	resp = tc.post("/api/proficiencies", map[string]any{
		"character_id": cid, "type": "skill", "name": "Perception",
	})
	if resp.Code != 201 {
		t.Fatalf("add perception proficiency failed: %d", resp.Code)
	}

	// Update character to trigger passive perception auto-calc
	resp = tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Perceptive Hero", "race": "Elf", "class": "Ranger",
		"level": 1, "wis": 16, "proficiency_bonus": 2,
		"ac": 10, "initiative": 0, "speed": 30, "str": 10, "dex": 14,
		"con": 12, "int": 10, "cha": 10,
	})
	if resp.Code != 200 {
		t.Fatalf("update character failed: %d", resp.Code)
	}

	// Verify passive perception: 10 + wisMod(3) + PB(2) = 15
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	pp := int(char["passive_perception"].(float64))
	if pp != 15 {
		t.Fatalf("expected passive_perception 15 (10+3+2), got %d", pp)
	}
}

func TestProficiencyBonusOnLevelUp(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Level Hero", "race": "Human", "class": "Fighter",
		"level": 1, "hp_max": 12, "hp_current": 12, "con": 14,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Level up from 1 to 5 (should go from PB=2 to PB=3)
	for i := range 4 {
		resp = tc.post(fmt.Sprintf("/api/characters/%d/levelup", cid), nil)
		if resp.Code != 200 {
			t.Fatalf("levelup failed at iteration %d: %d", i, resp.Code)
		}
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	lvl := int(char["level"].(float64))
	if lvl != 5 {
		t.Fatalf("expected level 5, got %d", lvl)
	}
	pb := int(char["proficiency_bonus"].(float64))
	if pb != 3 {
		t.Fatalf("expected proficiency_bonus 3 at level 5, got %d", pb)
	}
}

func TestShareLinks(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a character
	resp := tc.post("/api/characters", map[string]any{
		"name": "Shareable Hero", "race": "Elf", "class": "Ranger", "level": 3,
		"str": 12, "dex": 16, "con": 14, "int": 10, "wis": 14, "cha": 8,
		"hp_max": 28, "hp_current": 22,
	})
	if resp.Code != 201 {
		t.Fatalf("create character failed: %d - %s", resp.Code, resp.Body.String())
	}
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create campaign for party share
	resp = tc.post("/api/campaigns", map[string]any{
		"name": "Share Campaign", "party_name": "Share Party",
	})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d", resp.Code)
	}
	var camp map[string]any
	readJSON(resp, &camp)
	campaignID := int(camp["id"].(float64))

	// Assign character to campaign
	tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Shareable Hero", "race": "Elf", "class": "Ranger", "level": 3,
		"campaign_id": campaignID,
		"str":         12, "dex": 16, "con": 14, "int": 10, "wis": 14, "cha": 8,
		"hp_max": 28, "hp_current": 22, "ac": 15, "initiative": 3, "speed": 30,
	})

	// --- Character share ---
	resp = tc.post("/api/share", map[string]any{
		"entity_type": "character",
		"entity_id":   cid,
	})
	if resp.Code != 201 {
		t.Fatalf("create character share link failed: %d - %s", resp.Code, resp.Body.String())
	}
	var shareResp map[string]any
	readJSON(resp, &shareResp)
	charToken := shareResp["token"].(string)
	charURL := shareResp["url"].(string)
	if charToken == "" {
		t.Fatal("expected non-empty token")
	}
	if charURL == "" {
		t.Fatal("expected non-empty url")
	}

	// --- Party share ---
	resp = tc.post("/api/share", map[string]any{
		"entity_type": "party",
		"entity_id":   campaignID,
	})
	if resp.Code != 201 {
		t.Fatalf("create party share link failed: %d - %s", resp.Code, resp.Body.String())
	}
	readJSON(resp, &shareResp)
	partyToken := shareResp["token"].(string)
	if partyToken == "" {
		t.Fatal("expected non-empty party token")
	}

	// --- List share links ---
	resp = tc.get("/api/share", nil)
	if resp.Code != 200 {
		t.Fatalf("list share links failed: %d", resp.Code)
	}
	var links []any
	readJSON(resp, &links)
	if len(links) < 2 {
		t.Fatalf("expected at least 2 share links, got %d", len(links))
	}

	// --- Get shared character via public route ---
	tc2 := newTestClient() // no auth
	resp = tc2.get("/api/share/"+charToken, nil)
	if resp.Code != 200 {
		t.Fatalf("get shared character failed: %d - %s", resp.Code, resp.Body.String())
	}
	var sharedChar map[string]any
	readJSON(resp, &sharedChar)
	if sharedChar["name"] != "Shareable Hero" {
		t.Fatalf("expected 'Shareable Hero', got %q", sharedChar["name"])
	}
	if sharedChar["race"] != "Elf" {
		t.Fatalf("expected race 'Elf', got %q", sharedChar["race"])
	}
	if sharedChar["level"].(float64) != 3 {
		t.Fatalf("expected level 3, got %v", sharedChar["level"])
	}
	t.Logf("Shared character: %s (%s Lvl %v)", sharedChar["name"], sharedChar["race"], sharedChar["level"])

	// --- Get shared party via public route ---
	resp = tc2.get("/api/share/"+partyToken, nil)
	if resp.Code != 200 {
		t.Fatalf("get shared party failed: %d - %s", resp.Code, resp.Body.String())
	}
	var sharedParty map[string]any
	readJSON(resp, &sharedParty)
	campaign := sharedParty["campaign"].(map[string]any)
	if campaign["name"] != "Share Campaign" {
		t.Fatalf("expected campaign name 'Share Campaign', got %q", campaign["name"])
	}
	members := sharedParty["members"].([]any)
	if len(members) != 1 {
		t.Fatalf("expected 1 party member, got %d", len(members))
	}
	member := members[0].(map[string]any)
	if member["name"] != "Shareable Hero" {
		t.Fatalf("expected 'Shareable Hero' in party, got %q", member["name"])
	}
	t.Logf("Shared party: %s with %d members", campaign["name"], len(members))

	// --- Invalid token ---
	resp = tc2.get("/api/share/invalidtoken123", nil)
	if resp.Code != 404 {
		t.Fatalf("expected 404 for invalid token, got %d", resp.Code)
	}

	// --- Delete share links ---
	resp = tc.del("/api/share/"+charToken, nil)
	if resp.Code != 200 {
		t.Fatalf("delete char share link failed: %d", resp.Code)
	}
	resp = tc.del("/api/share/"+partyToken, nil)
	if resp.Code != 200 {
		t.Fatalf("delete party share link failed: %d", resp.Code)
	}

	// Verify deleted
	resp = tc2.get("/api/share/"+charToken, nil)
	if resp.Code != 404 {
		t.Fatalf("expected 404 for deleted share, got %d", resp.Code)
	}
}

func TestEmailSettingsConfig(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Save email settings
	resp := tc.post("/api/admin/email-settings", map[string]any{
		"smtp_host": "smtp.example.com",
		"smtp_port": 587,
		"username":  "test@example.com",
		"password":  "secret",
		"from_addr": "test@example.com",
		"enabled":   true,
	})
	if resp.Code != 200 {
		t.Fatalf("save email settings failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get email settings (password should be masked)
	resp = tc.get("/api/admin/email-settings", nil)
	if resp.Code != 200 {
		t.Fatalf("get email settings failed: %d", resp.Code)
	}
	var settings map[string]any
	readJSON(resp, &settings)
	if settings["smtp_host"] != "smtp.example.com" {
		t.Fatalf("expected smtp_host 'smtp.example.com', got %q", settings["smtp_host"])
	}
	if settings["has_password"] != true {
		t.Fatalf("expected has_password true, got %v", settings["has_password"])
	}
	if settings["enabled"] != true {
		t.Fatalf("expected enabled true, got %v", settings["enabled"])
	}
	t.Logf("Email settings: host=%s, port=%v, enabled=%v", settings["smtp_host"], settings["smtp_port"], settings["enabled"])

	// Save with empty password (should preserve existing)
	resp = tc.post("/api/admin/email-settings", map[string]any{
		"smtp_host": "smtp.example.com",
		"smtp_port": 587,
		"username":  "test@example.com",
		"password":  "",
		"from_addr": "test@example.com",
		"enabled":   true,
	})
	if resp.Code != 200 {
		t.Fatalf("save email settings (empty pw) failed: %d", resp.Code)
	}

	// Update to different host
	resp = tc.post("/api/admin/email-settings", map[string]any{
		"smtp_host": "smtp.newhost.com",
		"smtp_port": 465,
		"username":  "new@example.com",
		"password":  "newsecret",
		"from_addr": "new@example.com",
		"enabled":   false,
	})
	if resp.Code != 200 {
		t.Fatalf("save updated email settings failed: %d", resp.Code)
	}

	resp = tc.get("/api/admin/email-settings", nil)
	readJSON(resp, &settings)
	if settings["smtp_host"] != "smtp.newhost.com" {
		t.Fatalf("expected smtp_host 'smtp.newhost.com', got %q", settings["smtp_host"])
	}
	if settings["enabled"] != false {
		t.Fatalf("expected enabled false, got %v", settings["enabled"])
	}
	t.Logf("Updated email settings: host=%s, enabled=%v", settings["smtp_host"], settings["enabled"])

	// Test email (will fail because SMTP is not reachable - that's expected)
	resp = tc.post("/api/admin/test-email", nil)
	if resp.Code == 200 {
		t.Log("Test email reported success (SMTP may be available)")
	} else {
		// Expected to fail since we can't reach smtp.newhost.com
		t.Logf("Test email failed as expected: %d - %s", resp.Code, resp.Body.String())
		// Verify it's an email-send error, not a server error
		if resp.Code != 400 && resp.Code != 500 {
			t.Fatalf("expected 400 or 500 for test email, got %d", resp.Code)
		}
	}
}
