package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestCheckRollAdvantageDisadvantage(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Adv Test", "race": "Human", "class": "Fighter",
		"str": 16,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Normal
	resp = tc.post("/api/roll/check", map[string]any{
		"character_id": cid, "type": "check", "name": "str",
	})
	if resp.Code != 200 {
		t.Fatalf("normal check failed: %d", resp.Code)
	}
	var normal map[string]any
	readJSON(resp, &normal)
	rolls := normal["rolls"].([]any)
	if len(rolls) != 1 {
		t.Fatalf("expected 1 roll for normal, got %d", len(rolls))
	}

	// Advantage
	resp = tc.post("/api/roll/check", map[string]any{
		"character_id": cid, "type": "check", "name": "str", "advantage": "advantage",
	})
	if resp.Code != 200 {
		t.Fatalf("advantage check failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	rolls = adv["rolls"].([]any)
	if len(rolls) != 2 {
		t.Fatalf("expected 2 rolls for advantage, got %d", len(rolls))
	}
	t.Logf("Advantage: %v", adv["text"])

	// Disadvantage
	resp = tc.post("/api/roll/check", map[string]any{
		"character_id": cid, "type": "check", "name": "str", "advantage": "disadvantage",
	})
	if resp.Code != 200 {
		t.Fatalf("disadvantage check failed: %d", resp.Code)
	}
	var dis map[string]any
	readJSON(resp, &dis)
	rolls = dis["rolls"].([]any)
	if len(rolls) != 2 {
		t.Fatalf("expected 2 rolls for disadvantage, got %d", len(rolls))
	}
	t.Logf("Disadvantage: %v", dis["text"])
}

func TestCheckRollErrors(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a character to test with
	resp := tc.post("/api/characters", map[string]any{
		"name": "Error Test", "race": "Human", "class": "Fighter",
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Unknown skill
	resp = tc.post("/api/roll/check", map[string]any{
		"character_id": cid, "type": "skill", "name": "nonexistent",
	})
	if resp.Code != 400 {
		t.Fatalf("expected 400 for unknown skill, got %d - %s", resp.Code, resp.Body.String())
	}

	// Invalid type
	resp = tc.post("/api/roll/check", map[string]any{
		"character_id": cid, "type": "invalid", "name": "str",
	})
	if resp.Code != 400 {
		t.Fatalf("expected 400 for invalid type, got %d", resp.Code)
	}

	// Character not found
	resp = tc.post("/api/roll/check", map[string]any{
		"character_id": 99999, "type": "check", "name": "str",
	})
	if resp.Code != 404 {
		t.Fatalf("expected 404 for missing character, got %d", resp.Code)
	}
}

func TestCombatTracker(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create character for initiative
	resp := tc.post("/api/characters", map[string]any{
		"name": "Combatant", "race": "Elf", "class": "Ranger",
		"dex": 18,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Roll initiative
	resp = tc.post("/api/combat/initiative", map[string]any{"character_id": cid})
	if resp.Code != 200 {
		t.Fatalf("initiative roll failed: %d", resp.Code)
	}
	var init map[string]any
	readJSON(resp, &init)
	if int(init["total"].(float64)) < 1 {
		t.Fatal("initiative roll too low")
	}
	t.Logf("Initiative: %+v", init)

	// Create combat entry
	resp = tc.post("/api/combat", map[string]any{
		"name": "Goblin", "type": "monster",
		"initiative_roll": 12, "initiative_mod": 2,
		"hp_max": 7, "hp_current": 7, "ac": 15,
	})
	if resp.Code != 201 {
		t.Fatalf("create combat entry failed: %d", resp.Code)
	}
	var entry map[string]any
	readJSON(resp, &entry)
	eid := int(entry["id"].(float64))

	// List combat entries
	resp = tc.get("/api/combat", nil)
	if resp.Code != 200 {
		t.Fatalf("list combat failed: %d", resp.Code)
	}
	var entries []any
	readJSON(resp, &entries)
	if len(entries) < 1 {
		t.Fatal("expected at least 1 combat entry")
	}

	// Update entry
	resp = tc.put(fmt.Sprintf("/api/combat/%d", eid), map[string]any{
		"name": "Goblin Archer", "type": "monster",
		"initiative_roll": 12, "initiative_mod": 3,
		"hp_max": 7, "hp_current": 5, "ac": 15,
		"is_active": true,
	})
	if resp.Code != 200 {
		t.Fatalf("update combat entry failed: %d", resp.Code)
	}

	// Delete
	resp = tc.del(fmt.Sprintf("/api/combat/%d", eid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete combat entry failed: %d", resp.Code)
	}
}

func TestGenerators(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	gens := []string{"/api/generate/npc", "/api/generate/name", "/api/generate/encounter", "/api/generate/loot"}
	for _, g := range gens {
		t.Run(g, func(t *testing.T) {
			resp := tc.get(g, nil)
			if resp.Code != 200 {
				t.Fatalf("%s failed: %d", g, resp.Code)
			}
		})
	}

	// With filters
	resp := tc.get("/api/generate/name?race=dwarf", nil)
	if resp.Code != 200 {
		t.Fatalf("dwarf name gen failed: %d", resp.Code)
	}
	resp = tc.get("/api/generate/encounter?terrain=forest&level=5", nil)
	if resp.Code != 200 {
		t.Fatalf("forest encounter gen failed: %d", resp.Code)
	}
	resp = tc.get("/api/generate/loot?cr=5-10", nil)
	if resp.Code != 200 {
		t.Fatalf("loot gen failed: %d", resp.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	tc := newTestClient()
	resp := tc.get("/healthz", nil)
	if resp.Code != 200 {
		t.Fatalf("healthz failed: %d", resp.Code)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	tc := newTestClient()
	resp := tc.get("/metrics", nil)
	if resp.Code != 200 {
		t.Fatalf("metrics failed: %d", resp.Code)
	}
}

func TestSpellcasting(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Spellcaster", "race": "Elf", "class": "Wizard",
		"int": 18, "level": 5,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/characters/%d/spellcasting", cid), map[string]any{
		"ability": "int", "save_dc": 15, "attack_bonus": 7,
		"slots_1_max": 4, "slots_1_used": 2,
		"slots_2_max": 3, "slots_2_used": 1,
		"slots_3_max": 2, "slots_3_used": 0,
	})
	if resp.Code != 200 {
		t.Fatalf("update spellcasting failed: %d", resp.Code)
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	sc := char["spellcasting"].(map[string]any)
	if int(sc["slots_1_max"].(float64)) != 4 {
		t.Fatalf("expected slots_1_max=4, got %v", sc["slots_1_max"])
	}
	if int(sc["slots_1_used"].(float64)) != 2 {
		t.Fatalf("expected slots_1_used=2, got %v", sc["slots_1_used"])
	}
	if int(sc["slots_3_max"].(float64)) != 2 {
		t.Fatalf("expected slots_3_max=2, got %v", sc["slots_3_max"])
	}
	if int(char["spell_save_dc"].(float64)) != 15 {
		t.Fatalf("expected spell_save_dc 15, got %v", char["spell_save_dc"])
	}
	t.Logf("Spellcasting: save_dc=%v, attack_bonus=%v, slots=%+v",
		char["spell_save_dc"], char["spell_attack_bonus"], sc)
}

// ─── Campaign Member Tests ───

func TestCampaignMemberAddListRemove(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/admin/users", map[string]any{
		"username": "playertest", "password": "testpass123", "role": "user", "display_name": "Player",
	})
	if resp.Code != 201 {
		t.Fatalf("create player user failed: %d", resp.Code)
	}

	resp = tc.post("/api/campaigns", map[string]any{"name": "Member Campaign", "description": "Test"})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d - %s", resp.Code, resp.Body.String())
	}
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/members", cid), map[string]any{"username": "playertest"})
	if resp.Code != 200 {
		t.Fatalf("add member failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/members", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list members failed: %d", resp.Code)
	}
	var members []map[string]any
	readJSON(resp, &members)
	if len(members) < 2 {
		t.Fatalf("expected at least 2 members (owner + player), got %d", len(members))
	}
	found := false
	for _, m := range members {
		if m["username"] == "playertest" {
			found = true
			if m["role"] != "player" {
				t.Fatalf("expected role 'player', got %v", m["role"])
			}
			break
		}
	}
	if !found {
		t.Fatalf("player member not found in campaign members list")
	}

	ownerFound := false
	for _, m := range members {
		if m["username"] == adminUser {
			ownerFound = true
			if m["role"] != "dm" {
				t.Fatalf("expected owner role 'dm', got %v", m["role"])
			}
			break
		}
	}
	if !ownerFound {
		t.Fatalf("owner not found in campaign members")
	}

	targetID := 0
	for _, m := range members {
		if m["username"] == "playertest" {
			targetID = int(m["user_id"].(float64))
		}
	}
	if targetID == 0 {
		t.Fatalf("could not find player user_id")
	}
	resp = tc.del(fmt.Sprintf("/api/campaigns/%d/members/%d", cid, targetID), nil)
	if resp.Code != 200 {
		t.Fatalf("remove member failed: %d", resp.Code)
	}

	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/members", cid), nil)
	readJSON(resp, &members)
	if len(members) != 1 {
		t.Fatalf("expected 1 member after removal, got %d", len(members))
	}

	tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
	tc.del(fmt.Sprintf("/api/admin/users/%d", targetID), nil)
}

func TestCampaignDMRoleAllowsCharacterAccess(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/admin/users", map[string]any{
		"username": "dmtestplayer", "password": "testpass123", "role": "user", "display_name": "DM Player",
	})
	if resp.Code != 201 {
		t.Fatalf("create player failed: %d", resp.Code)
	}
	var userResp map[string]any
	readJSON(resp, &userResp)
	playerID := int(userResp["id"].(float64))

	resp = tc.post("/api/admin/users", map[string]any{
		"username": "codm", "password": "testpass123", "role": "user", "display_name": "Co-DM",
	})
	if resp.Code != 201 {
		t.Fatalf("create co-dm failed: %d", resp.Code)
	}
	var coDMResp map[string]any
	readJSON(resp, &coDMResp)
	coDMID := int(coDMResp["id"].(float64))

	resp = tc.post("/api/campaigns", map[string]any{"name": "DM Access Campaign", "description": "Test"})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d", resp.Code)
	}
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	tc.post(fmt.Sprintf("/api/campaigns/%d/members", cid), map[string]any{"username": "dmtestplayer"})
	tc.post(fmt.Sprintf("/api/campaigns/%d/members", cid), map[string]any{"username": "codm"})

	resp = tc.put(fmt.Sprintf("/api/campaigns/%d/members/%d", cid, coDMID), map[string]any{"role": "dm"})
	if resp.Code != 200 {
		t.Fatalf("set co-dm role failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/members", cid), nil)
	var members []map[string]any
	readJSON(resp, &members)
	for _, m := range members {
		if m["username"] == "codm" && m["role"] != "dm" {
			t.Fatalf("expected co-dm role to be 'dm', got %v", m["role"])
		}
	}

	login(t, tc, "dmtestplayer", "testpass123")
	resp = tc.post("/api/characters", map[string]any{
		"name": "PlayerChar", "race": "Human", "class": "Fighter",
		"campaign_id": cid,
	})
	if resp.Code != 201 {
		t.Fatalf("player create char failed: %d", resp.Code)
	}
	var playerChar map[string]any
	readJSON(resp, &playerChar)
	playerCharID := int(playerChar["id"].(float64))

	login(t, tc, "codm", "testpass123")
	resp = tc.get(fmt.Sprintf("/api/characters/%d", playerCharID), nil)
	if resp.Code != 200 {
		t.Fatalf("co-DM cannot view player character: %d - %s", resp.Code, resp.Body.String())
	}
	var charData map[string]any
	readJSON(resp, &charData)
	if charData["name"] != "PlayerChar" {
		t.Fatalf("expected 'PlayerChar', got %v", charData["name"])
	}
	t.Logf("Co-DM viewed player character: %s", charData["name"])

	login(t, tc, adminUser, adminPass)

	resp = tc.put(fmt.Sprintf("/api/campaigns/%d/members/%d", cid, coDMID), map[string]any{"role": "player"})
	if resp.Code != 200 {
		t.Fatalf("demote co-dm failed: %d", resp.Code)
	}

	login(t, tc, "codm", "testpass123")
	resp = tc.get(fmt.Sprintf("/api/characters/%d", playerCharID), nil)
	if resp.Code != 200 {
		t.Fatalf("demoted user cannot view player character (still a member, read-only): %d - %s", resp.Code, resp.Body.String())
	}
	readJSON(resp, &charData)
	if canEdit, _ := charData["can_edit"].(bool); canEdit {
		t.Fatal("demoted user should not be able to edit the player character")
	}

	// Writes must still be denied for non-owners even though viewing is allowed
	resp = tc.put(fmt.Sprintf("/api/characters/%d", playerCharID), map[string]any{"name": "Hijacked"})
	if resp.Code != 403 {
		t.Fatalf("expected 403 for demoted user write, got %d - %s", resp.Code, resp.Body.String())
	}
	t.Logf("Demoted user can view read-only but cannot edit (got %d)", resp.Code)

	login(t, tc, adminUser, adminPass)
	tc.del(fmt.Sprintf("/api/characters/%d", playerCharID), nil)
	tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
	tc.del(fmt.Sprintf("/api/admin/users/%d", playerID), nil)
	tc.del(fmt.Sprintf("/api/admin/users/%d", coDMID), nil)
}

func TestUserSearch(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/admin/users", map[string]any{
		"username": "searchableuser", "password": "testpass123", "role": "user",
	})
	if resp.Code != 201 {
		t.Fatalf("create searchable user failed: %d", resp.Code)
	}
	var userResp map[string]any
	readJSON(resp, &userResp)
	uid := int(userResp["id"].(float64))

	resp = tc.get("/api/users/search?q=searchableuser", nil)
	if resp.Code != 200 {
		t.Fatalf("user search failed: %d - %s", resp.Code, resp.Body.String())
	}
	var results []map[string]any
	readJSON(resp, &results)
	if len(results) == 0 {
		t.Fatal("expected at least 1 search result")
	}
	if results[0]["username"] != "searchableuser" {
		t.Fatalf("expected 'searchableuser', got %v", results[0]["username"])
	}

	resp = tc.get("/api/users/search?q=search", nil)
	readJSON(resp, &results)
	if len(results) == 0 {
		t.Fatal("expected partial search results")
	}

	resp = tc.get("/api/users/search?q=", nil)
	readJSON(resp, &results)
	if len(results) != 0 {
		t.Fatalf("expected empty for empty query, got %d", len(results))
	}

	resp = tc.get("/api/users/search?q=zzzznonexistent", nil)
	readJSON(resp, &results)
	if len(results) != 0 {
		t.Fatalf("expected 0 for nonexistent, got %d", len(results))
	}

	tc.del(fmt.Sprintf("/api/admin/users/%d", uid), nil)
}

func TestDiceRollAdvantageDisadvantage(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/roll", map[string]any{
		"expression": "1d20",
		"advantage":  "advantage",
	})
	if resp.Code != 200 {
		t.Fatalf("advantage roll failed: %d - %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	readJSON(resp, &result)
	breakdown := result["breakdown"].([]any)
	rolls := breakdown[0].(map[string]any)["rolls"].([]any)
	if len(rolls) != 2 {
		t.Fatalf("expected 2 rolls for advantage, got %d", len(rolls))
	}
	total := int(result["total"].(float64))
	rollVal := func(r any) int {
		if v, ok := r.(float64); ok {
			return int(v)
		}
		if m, ok := r.(map[string]any); ok {
			if v, ok := m["value"].(float64); ok {
				return int(v)
			}
		}
		return 0
	}
	r1, r2 := rollVal(rolls[0]), rollVal(rolls[1])
	if total != max(r1, r2) {
		t.Fatalf("advantage total %d should be max(%d,%d)=%d", total, r1, r2, max(r1, r2))
	}
	t.Logf("Advantage: [%d,%d] total=%d", r1, r2, total)

	resp = tc.post("/api/roll", map[string]any{
		"expression": "1d20",
		"advantage":  "disadvantage",
	})
	if resp.Code != 200 {
		t.Fatalf("disadvantage roll failed: %d - %s", resp.Code, resp.Body.String())
	}
	readJSON(resp, &result)
	breakdown = result["breakdown"].([]any)
	rolls = breakdown[0].(map[string]any)["rolls"].([]any)
	if len(rolls) != 2 {
		t.Fatalf("expected 2 rolls for disadvantage, got %d", len(rolls))
	}
	total = int(result["total"].(float64))
	r1, r2 = rollVal(rolls[0]), rollVal(rolls[1])
	if total != min(r1, r2) {
		t.Fatalf("disadvantage total %d should be min(%d,%d)=%d", total, r1, r2, min(r1, r2))
	}
	t.Logf("Disadvantage: [%d,%d] total=%d", r1, r2, total)

	resp = tc.post("/api/roll", map[string]any{"expression": "1d20"})
	if resp.Code != 200 {
		t.Fatalf("normal roll failed: %d", resp.Code)
	}
	readJSON(resp, &result)
	breakdown = result["breakdown"].([]any)
	rolls = breakdown[0].(map[string]any)["rolls"].([]any)
	if len(rolls) != 1 {
		t.Fatalf("expected 1 roll for normal, got %d", len(rolls))
	}
	total = int(result["total"].(float64))
	if total != rollVal(rolls[0]) {
		t.Fatalf("normal total %d != roll %d", total, rollVal(rolls[0]))
	}
	t.Logf("Normal: %d", total)

	resp = tc.post("/api/roll", map[string]any{
		"expression": "1d20+5",
		"advantage":  "advantage",
	})
	if resp.Code != 200 {
		t.Fatalf("advantage+mod failed: %d", resp.Code)
	}
	readJSON(resp, &result)
	breakdown = result["breakdown"].([]any)
	if len(breakdown) != 2 {
		t.Fatalf("expected 2 breakdown items for adv+mod, got %d", len(breakdown))
	}
	rolls = breakdown[0].(map[string]any)["rolls"].([]any)
	chosen := rollVal(breakdown[0].(map[string]any)["total"])
	total = int(result["total"].(float64))
	if total != chosen+5 {
		t.Fatalf("adv+5 total %d != chosen %d + 5 = %d", total, chosen, chosen+5)
	}
	t.Logf("Advantage+5: rolls=[%d,%d] chosen=%d total=%d",
		rollVal(rolls[0]), rollVal(rolls[1]), chosen, total)

	// ─── Compendium System/Source Fields ───

	resp = tc.get("/api/compendium/races", nil)
	if resp.Code == 200 {
		var races []map[string]any
		readJSON(resp, &races)
		if len(races) > 0 {
			r := races[0]
			if _, ok := r["system"]; !ok {
				t.Error("expected 'system' field in compendium race response")
			}
			if _, ok := r["source"]; !ok {
				t.Error("expected 'source' field in compendium race response")
			}
			t.Logf("Race system=%v source=%v", r["system"], r["source"])
		}
	}
}

// ─── Portrait Tests ───

func TestCharacterPortrait(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Portrait Hero", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Set portrait URL
	resp = tc.put(fmt.Sprintf("/api/characters/%d", cid), map[string]any{
		"name": "Portrait Hero", "race": "Human", "class": "Fighter", "level": 1,
		"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10,
		"hp_max": 10, "hp_current": 10, "ac": 10, "initiative": 0, "speed": 30,
		"portrait_url": "/media/test/portrait.jpg",
	})
	if resp.Code != 200 {
		t.Fatalf("update portrait failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if char["portrait_url"] != "/media/test/portrait.jpg" {
		t.Fatalf("expected portrait_url='/media/test/portrait.jpg', got %v", char["portrait_url"])
	}
}

// ─── Multi-class Tests ───

func TestMultiClass(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Multi Hero", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Add second class
	resp = tc.post(fmt.Sprintf("/api/characters/%d/classes", cid), map[string]any{
		"class": "Wizard", "subclass": "Evocation", "level": 3, "hit_dice": "d6",
	})
	if resp.Code != 201 {
		t.Fatalf("add class failed: %d - %s", resp.Code, resp.Body.String())
	}
	var cc map[string]any
	readJSON(resp, &cc)
	ccid := int(cc["id"].(float64))

	// Verify via get character
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	classes := char["classes"].([]any)
	if len(classes) != 1 {
		t.Fatalf("expected 1 class entry, got %d", len(classes))
	}

	// Update class
	resp = tc.put(fmt.Sprintf("/api/classes/%d", ccid), map[string]any{
		"class": "Wizard", "subclass": "Divination", "level": 4, "hit_dice": "d6",
	})
	if resp.Code != 200 {
		t.Fatalf("update class failed: %d", resp.Code)
	}

	// Delete class
	resp = tc.del(fmt.Sprintf("/api/classes/%d", ccid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete class failed: %d", resp.Code)
	}
}

// ─── Encounter Builder Tests ───

func TestEncounterBuilder(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create encounter
	resp := tc.post("/api/encounters", map[string]any{
		"name": "Goblin Ambush", "description": "A goblin attack", "difficulty": "medium",
		"environment": "forest",
	})
	if resp.Code != 201 {
		t.Fatalf("create encounter failed: %d - %s", resp.Code, resp.Body.String())
	}
	var enc map[string]any
	readJSON(resp, &enc)
	eid := int(enc["id"].(float64))

	// Add monsters
	resp = tc.post(fmt.Sprintf("/api/encounters/%d/monsters", eid), map[string]any{
		"name": "Goblin", "count": 3, "cr": "1/4", "xp": 50, "ac": 15, "hp": 7,
	})
	if resp.Code != 201 {
		t.Fatalf("add monster failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get encounter
	resp = tc.get(fmt.Sprintf("/api/encounters/%d", eid), nil)
	if resp.Code != 200 {
		t.Fatalf("get encounter failed: %d", resp.Code)
	}
	readJSON(resp, &enc)
	if enc["name"] != "Goblin Ambush" {
		t.Fatalf("expected 'Goblin Ambush', got %v", enc["name"])
	}

	// List encounters
	resp = tc.get("/api/encounters", nil)
	if resp.Code != 200 {
		t.Fatalf("list encounters failed: %d", resp.Code)
	}

	// Calculate XP
	resp = tc.post("/api/encounters/calculate-xp", map[string]any{
		"party_levels": []int{1, 1, 1, 1},
		"monsters":     []map[string]any{{"name": "Goblin", "cr": "1/4", "count": 3, "xp": 50}},
	})
	if resp.Code != 200 {
		t.Fatalf("calculate xp failed: %d - %s", resp.Code, resp.Body.String())
	}
	var xpResult map[string]any
	readJSON(resp, &xpResult)
	if int(xpResult["total_xp"].(float64)) != 150 {
		t.Fatalf("expected 150 XP, got %v", xpResult["total_xp"])
	}

	// Get monster XP table
	resp = tc.get("/api/monster-xp", nil)
	if resp.Code != 200 {
		t.Fatalf("monster xp failed: %d", resp.Code)
	}

	// Delete
	resp = tc.del(fmt.Sprintf("/api/encounters/%d", eid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete encounter failed: %d", resp.Code)
	}
}

// ─── Calendar Tests ───

func TestCalendarEvents(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Calendar Campaign", "description": "Test"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	resp = tc.post("/api/calendar", map[string]any{
		"campaign_id": cid, "title": "Full Moon", "event_date": "1491-04-15",
		"event_type": "holiday", "description": "A full moon rises",
	})
	if resp.Code != 201 {
		t.Fatalf("create calendar event failed: %d - %s", resp.Code, resp.Body.String())
	}
	var ev map[string]any
	readJSON(resp, &ev)
	eid := int(ev["id"].(float64))

	resp = tc.get("/api/calendar?campaign_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list calendar events failed: %d", resp.Code)
	}
	var events []any
	readJSON(resp, &events)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	resp = tc.put(fmt.Sprintf("/api/calendar/%d", eid), map[string]any{
		"campaign_id": cid, "title": "Blood Moon", "event_date": "1491-04-15",
		"event_type": "holiday", "color": "#ff0000",
	})
	if resp.Code != 200 {
		t.Fatalf("update calendar event failed: %d", resp.Code)
	}

	resp = tc.del(fmt.Sprintf("/api/calendar/%d", eid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete calendar event failed: %d", resp.Code)
	}

	tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
}

// ─── Timeline Tests ───

func TestTimelineEvents(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Timeline Campaign", "description": "Test"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Create timeline event
	resp = tc.post("/api/timeline", map[string]any{
		"campaign_id": cid, "title": "Party meets in tavern",
		"event_date": "1491-03-01", "event_type": "milestone",
		"importance": 5, "description": "The adventure begins",
	})
	if resp.Code != 201 {
		t.Fatalf("create timeline event failed: %d - %s", resp.Code, resp.Body.String())
	}
	var ev map[string]any
	readJSON(resp, &ev)
	eid := int(ev["id"].(float64))

	// List
	resp = tc.get("/api/timeline?campaign_id="+strconv.Itoa(cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list timeline failed: %d", resp.Code)
	}
	var events []any
	readJSON(resp, &events)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// Update
	resp = tc.put(fmt.Sprintf("/api/timeline/%d", eid), map[string]any{
		"campaign_id": cid, "title": "Party meets at crossroads",
		"event_date": "1491-03-01", "event_type": "milestone",
		"importance": 4, "icon": "fa-crossroads",
	})
	if resp.Code != 200 {
		t.Fatalf("update timeline failed: %d", resp.Code)
	}

	// Delete
	resp = tc.del(fmt.Sprintf("/api/timeline/%d", eid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete timeline failed: %d", resp.Code)
	}

	tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
}

func TestCompendiumJSONSeed(t *testing.T) {
	// Test that JSON-seeded data includes system/source fields
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/compendium/races", nil)
	if resp.Code != 200 {
		t.Fatalf("get races failed: %d", resp.Code)
	}
	var races []map[string]any
	readJSON(resp, &races)
	if len(races) == 0 {
		t.Fatal("no races found")
	}
	// Check system/source on first race
	r := races[0]
	sys, ok := r["system"]
	if !ok || sys != "dnd5e" {
		t.Errorf("expected system='dnd5e', got %v", sys)
	}
	src, ok := r["source"]
	if !ok || src != "srd" {
		t.Errorf("expected source='srd', got %v", src)
	}

	// Verify classes also have fields
	resp = tc.get("/api/compendium/classes", nil)
	var classes []map[string]any
	readJSON(resp, &classes)
	if len(classes) > 0 {
		if _, ok := classes[0]["system"]; !ok {
			t.Error("expected 'system' in classes")
		}
		if _, ok := classes[0]["source"]; !ok {
			t.Error("expected 'source' in classes")
		}
	}
}
