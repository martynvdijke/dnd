package main

import (
	"fmt"
	"strings"
	"testing"

	"villum/db"
)

func TestCheckSetup(t *testing.T) {
	tc := newTestClient()
	resp := tc.get("/api/check-setup", nil)
	if resp.Code != 200 {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestSetupAndLogin(t *testing.T) {
	tc := newTestClient()

	resp := tc.post("/api/login", map[string]any{
		"username": adminUser,
		"password": adminPass,
		"setup":    true,
	})
	if resp.Code != 200 && resp.Code != 400 {
		t.Fatalf("setup failed: %d - %s", resp.Code, resp.Body.String())
	}
	// If admin already exists (400), setup is already done — continue to login.

	// Login
	resp = tc.post("/api/login", map[string]any{
		"username": adminUser,
		"password": adminPass,
	})
	if resp.Code != 200 {
		t.Fatalf("login failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get CSRF
	resp = tc.get("/api/csrf-token", nil)
	if resp.Code != 200 {
		t.Fatalf("csrf token failed: %d", resp.Code)
	}
	var csrfData map[string]string
	readJSON(resp, &csrfData)
	tc.csrf = csrfData["token"]
	// Fetch API token for mutating requests (logout is POST under APITokenRequired).
	resp = tc.post("/api/tokens", map[string]any{"name": "test-token"})
	if resp.Code == 201 {
		var tok map[string]any
		readJSON(resp, &tok)
		if s, ok := tok["token"].(string); ok {
			tc.token = s
		}
	}

	// Get me
	resp = tc.get("/api/user/me", nil)
	if resp.Code != 200 {
		t.Fatalf("get me failed: %d", resp.Code)
	}
	var me map[string]any
	readJSON(resp, &me)
	if me["role"] != "admin" {
		t.Fatalf("expected admin, got %v", me["role"])
	}

	// Logout
	resp = tc.post("/api/logout", nil)
	if resp.Code != 200 {
		t.Fatalf("logout failed: %d", resp.Code)
	}
}

func TestInvalidLogin(t *testing.T) {
	tc := newTestClient()
	resp := tc.post("/api/login", map[string]any{
		"username": "nobody",
		"password": "wrong",
	})
	if resp.Code != 401 {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
}

// ─── Character CRUD Tests ───

func TestCharacterCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Test Hero", "race": "Elf", "class": "Wizard", "level": 5,
		"str": 10, "dex": 14, "con": 12, "int": 18, "wis": 14, "cha": 8,
		"hp_max": 32, "hp_current": 32,
	})
	if resp.Code != 201 {
		t.Fatalf("create failed: %d - %s", resp.Code, resp.Body.String())
	}
	var char map[string]any
	readJSON(resp, &char)
	id := int(char["id"].(float64))

	resp = tc.get(fmt.Sprintf("/api/characters/%d", id), nil)
	if resp.Code != 200 {
		t.Fatalf("get failed: %d", resp.Code)
	}

	resp = tc.put(fmt.Sprintf("/api/characters/%d", id), map[string]any{
		"name": "Updated Hero", "race": "Elf", "class": "Wizard", "level": 6,
		"hp_max": 40, "hp_current": 36, "str": 10, "dex": 14, "con": 12,
		"int": 18, "wis": 14, "cha": 8, "ac": 12, "initiative": 2, "speed": 30,
	})
	if resp.Code != 200 {
		t.Fatalf("update failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.get("/api/characters", nil)
	if resp.Code != 200 {
		t.Fatalf("list failed: %d", resp.Code)
	}

	resp = tc.del(fmt.Sprintf("/api/characters/%d", id), nil)
	if resp.Code != 200 {
		t.Fatalf("delete failed: %d", resp.Code)
	}
}

func TestCharacterCurrency(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Rich Hero", "race": "Human", "class": "Rogue"})
	var char map[string]any
	readJSON(resp, &char)
	id := int(char["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/characters/%d/currency", id), map[string]any{"gp": 100, "sp": 50, "cp": 25, "pp": 5})
	if resp.Code != 200 {
		t.Fatalf("currency update failed: %d", resp.Code)
	}
}

func TestCharacterInventory(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Packer", "race": "Dwarf", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	id := int(char["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/inventory", id), map[string]any{
		"name": "Longsword", "category": "weapon", "quantity": 1,
		"damage_dice": "1d8", "damage_type": "slashing",
	})
	if resp.Code != 201 {
		t.Fatalf("add item failed: %d", resp.Code)
	}
	var item map[string]any
	readJSON(resp, &item)
	iid := int(item["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/inventory/%d", iid), map[string]any{
		"name": "Longsword+1", "category": "weapon", "quantity": 1,
		"damage_dice": "1d8", "damage_type": "slashing", "is_magical": true,
		"description": "", "weapon_properties": "", "ac_bonus": 0, "armor_type": "",
		"is_equipped": false, "attunement": false, "notes": "",
	})
	if resp.Code != 200 {
		t.Fatalf("update item failed: %d", resp.Code)
	}

	resp = tc.del(fmt.Sprintf("/api/inventory/%d", iid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete item failed: %d", resp.Code)
	}
}

// ─── Dice Tests ───

func TestDiceRoll(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	tests := []struct {
		expr string
		min  int
		max  int
	}{
		{"1d4", 1, 4},
		{"1d6", 1, 6},
		{"1d8", 1, 8},
		{"1d10", 1, 10},
		{"1d12", 1, 12},
		{"1d20", 1, 20},
		{"1d100", 1, 100},
		{"2d6", 2, 12},
		{"1d8+3", 4, 11},
		{"3d4+2", 5, 14},
		{"2d20", 2, 40},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			resp := tc.post("/api/roll", map[string]any{"expression": tt.expr})
			if resp.Code != 200 {
				t.Fatalf("roll %s failed: %d", tt.expr, resp.Code)
			}
			var result map[string]any
			readJSON(resp, &result)
			total := int(result["total"].(float64))
			if total < tt.min || total > tt.max {
				t.Errorf("roll %s = %d, expected [%d,%d]", tt.expr, total, tt.min, tt.max)
			}
			// Verify history
			resp2 := tc.get("/api/dice-rolls", nil)
			if resp2.Code == 200 {
				var rolls []any
				readJSON(resp2, &rolls)
				if len(rolls) > 0 {
					t.Logf("dice history has %d entries", len(rolls))
				}
			}
		})
	}
}

// ─── Compendium Tests ───

func TestCompendium(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	endpoints := []string{
		"/api/compendium/races", "/api/compendium/classes", "/api/compendium/spells",
		"/api/compendium/feats", "/api/compendium/backgrounds", "/api/compendium/equipment",
		"/api/compendium/search?q=fire",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			resp := tc.get(ep, nil)
			if resp.Code != 200 {
				t.Fatalf("%s failed: %d", ep, resp.Code)
			}
		})
	}
}

// ─── Campaign Features Tests ───

func TestCampaigns(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create
	resp := tc.post("/api/campaigns", map[string]any{
		"name": "Curse of Strahd", "party_name": "The Dawnbringers", "description": "Ravenloft campaign",
	})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d - %s", resp.Code, resp.Body.String())
	}
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// List
	resp = tc.get("/api/campaigns", nil)
	if resp.Code != 200 {
		t.Fatalf("list failed: %d", resp.Code)
	}
	var camps []map[string]any
	readJSON(resp, &camps)
	found := false
	for _, c := range camps {
		if int(c["id"].(float64)) == cid {
			found = true
			if c["party_name"].(string) != "The Dawnbringers" {
				t.Fatalf("expected party_name 'The Dawnbringers', got %q", c["party_name"])
			}
			break
		}
	}
	if !found {
		t.Fatal("created campaign not found in list")
	}

	// Update
	resp = tc.put(fmt.Sprintf("/api/campaigns/%d", cid), map[string]any{
		"name": "Curse of Strahd Revised", "party_name": "The Dawnslingers", "description": "Updated campaign",
	})
	if resp.Code != 200 {
		t.Fatalf("update failed: %d", resp.Code)
	}

	// Verify update persisted
	resp = tc.get("/api/campaigns", nil)
	if resp.Code != 200 {
		t.Fatalf("list after update failed: %d", resp.Code)
	}
	readJSON(resp, &camps)
	for _, c := range camps {
		if int(c["id"].(float64)) == cid {
			if c["party_name"].(string) != "The Dawnslingers" {
				t.Fatalf("expected party_name 'The Dawnslingers' after update, got %q", c["party_name"])
			}
			break
		}
	}

	// Create character in campaign
	resp = tc.post("/api/characters", map[string]any{
		"name": "Campaign Hero", "race": "Half-Elf", "class": "Paladin",
		"campaign_id": cid,
	})
	if resp.Code != 201 {
		t.Fatalf("create char in campaign failed: %d", resp.Code)
	}

	// Delete
	resp = tc.del(fmt.Sprintf("/api/campaigns/%d", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete campaign failed: %d", resp.Code)
	}
}

func TestLocationsAndNPCs(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Explorer", "race": "Human", "class": "Ranger"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Location
	resp = tc.post("/api/locations", map[string]any{
		"name": "Waterdeep", "type": "city", "description": "City of Splendors",
	})
	if resp.Code != 201 {
		t.Fatalf("create location failed: %d", resp.Code)
	}
	var loc map[string]any
	readJSON(resp, &loc)
	lid := int(loc["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/locations", cid), map[string]any{
		"location_id": lid, "relationship": "visited",
	})
	if resp.Code != 200 {
		t.Fatalf("link location failed: %d", resp.Code)
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d/locations", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get locations failed: %d", resp.Code)
	}
	var locs []any
	readJSON(resp, &locs)
	if len(locs) != 1 {
		t.Fatalf("expected 1 location link, got %d", len(locs))
	}
	firstLoc := locs[0].(map[string]any)
	lid2 := int(firstLoc["id"].(float64))
	tc.del(fmt.Sprintf("/api/locations/link/%d", lid2), nil)

	tc.del(fmt.Sprintf("/api/locations/%d", lid), nil)

	// NPC
	resp = tc.post("/api/npcs", map[string]any{
		"name": "Elminster", "race": "Human", "class": "Wizard",
		"description": "Sage of Shadowdale",
	})
	if resp.Code != 201 {
		t.Fatalf("create NPC failed: %d", resp.Code)
	}
	var npc map[string]any
	readJSON(resp, &npc)
	nid := int(npc["id"].(float64))

	resp = tc.post(fmt.Sprintf("/api/characters/%d/npcs", cid), map[string]any{
		"npc_id": nid, "relationship": "ally",
	})
	if resp.Code != 200 {
		t.Fatalf("link NPC failed: %d", resp.Code)
	}

	resp = tc.get(fmt.Sprintf("/api/characters/%d/npcs", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get NPCs failed: %d", resp.Code)
	}
	var npcLinks []any
	readJSON(resp, &npcLinks)
	if len(npcLinks) != 1 {
		t.Fatalf("expected 1 NPC link, got %d", len(npcLinks))
	}

	// Log interaction
	firstLink := npcLinks[0].(map[string]any)
	linkID := int(firstLink["id"].(float64))
	resp = tc.post(fmt.Sprintf("/api/npcs/link/%d/interact", linkID), map[string]any{})
	if resp.Code != 200 {
		t.Fatalf("log interaction failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Verify interaction count increased
	resp = tc.get(fmt.Sprintf("/api/characters/%d/npcs", cid), nil)
	var updatedLinks []any
	readJSON(resp, &updatedLinks)
	if len(updatedLinks) == 1 {
		updated := updatedLinks[0].(map[string]any)
		if int(updated["interaction_count"].(float64)) != 1 {
			t.Fatalf("expected interaction_count=1, got %v", updated["interaction_count"])
		}
	}

	tc.del(fmt.Sprintf("/api/npcs/%d", nid), nil)
}

func TestSessionsQuestsJournal(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Adventurer", "race": "Dragonborn", "class": "Sorcerer"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Session
	resp = tc.post(fmt.Sprintf("/api/characters/%d/sessions", cid), map[string]any{
		"title": "Session 1", "notes": "Began journey", "xp_earned": 300, "gold_earned": 50,
	})
	if resp.Code != 201 {
		t.Fatalf("create session failed: %d", resp.Code)
	}
	var sess map[string]any
	readJSON(resp, &sess)
	sid := int(sess["id"].(float64))
	tc.get(fmt.Sprintf("/api/characters/%d/sessions", cid), nil)
	tc.del(fmt.Sprintf("/api/sessions/%d", sid), nil)

	// Quest
	resp = tc.post(fmt.Sprintf("/api/characters/%d/quests", cid), map[string]any{
		"name": "Find the Crown", "description": "Retrieve the lost crown", "status": "active",
	})
	if resp.Code != 201 {
		t.Fatalf("create quest failed: %d", resp.Code)
	}
	var quest map[string]any
	readJSON(resp, &quest)
	qid := int(quest["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/quests/%d", qid), map[string]any{
		"name": "Find the Crown", "description": "Retrieve the lost crown",
		"status": "complete", "objectives": "", "rewards": "", "notes": "",
	})
	if resp.Code != 200 {
		t.Fatalf("update quest failed: %d", resp.Code)
	}
	tc.del(fmt.Sprintf("/api/quests/%d", qid), nil)

	// Journal
	resp = tc.post(fmt.Sprintf("/api/characters/%d/journal", cid), map[string]any{
		"title": "Day 1", "entry": "Today we began...",
	})
	if resp.Code != 201 {
		t.Fatalf("create journal failed: %d", resp.Code)
	}
	var jentry map[string]any
	readJSON(resp, &jentry)
	jid := int(jentry["id"].(float64))
	tc.get(fmt.Sprintf("/api/characters/%d/journal", cid), nil)
	tc.del(fmt.Sprintf("/api/journal/%d", jid), nil)
}

func TestPartyView(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a campaign with party name
	resp := tc.post("/api/campaigns", map[string]any{
		"name": "Test Campaign", "party_name": "Test Party",
	})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d", resp.Code)
	}
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Create characters assigned to campaign
	tc.post("/api/characters", map[string]any{"name": "Party Hero 1", "race": "Human", "class": "Fighter", "hp_max": 30, "hp_current": 25, "campaign_id": cid})
	tc.post("/api/characters", map[string]any{"name": "Party Hero 2", "race": "Elf", "class": "Wizard", "hp_max": 20, "hp_current": 20, "campaign_id": cid})

	resp = tc.get("/api/party", nil)
	if resp.Code != 200 {
		t.Fatalf("party view failed: %d - %s", resp.Code, resp.Body.String())
	}
	var groups []any
	readJSON(resp, &groups)
	if len(groups) < 1 {
		t.Fatalf("expected at least 1 party group, got %d", len(groups))
	}
	// Check party name appears and matches
	found := false
	partyNameFound := false
	for _, g := range groups {
		gm := g.(map[string]any)
		if gm["party_name"] == "Test Party" {
			partyNameFound = true
		}
		members := gm["members"].([]any)
		for _, m := range members {
			mm := m.(map[string]any)
			if mm["name"] == "Party Hero 1" {
				found = true
			}
		}
	}
	if !partyNameFound {
		t.Fatal("expected party_name 'Test Party' in party view")
	}
	if !found {
		t.Fatal("expected Party Hero 1 in party view")
	}
	t.Logf("Party groups: %d", len(groups))
}

func TestStatsEndpoint(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Stat Hero", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Add some data to generate stats
	tc.post(fmt.Sprintf("/api/characters/%d/sessions", cid), map[string]any{"title": "S1", "notes": "First", "xp_earned": 300})
	tc.post(fmt.Sprintf("/api/characters/%d/sessions", cid), map[string]any{"title": "S2", "notes": "Second", "xp_earned": 450})
	tc.post(fmt.Sprintf("/api/characters/%d/quests", cid), map[string]any{"name": "Q1", "status": "active"})
	tc.post(fmt.Sprintf("/api/characters/%d/quests", cid), map[string]any{"name": "Q2", "status": "complete"})

	resp = tc.get(fmt.Sprintf("/api/characters/%d/stats", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("stats failed: %d - %s", resp.Code, resp.Body.String())
	}
	var stats map[string]any
	readJSON(resp, &stats)
	if int(stats["session_count"].(float64)) != 2 {
		t.Fatalf("expected 2 sessions, got %v", stats["session_count"])
	}
	if int(stats["total_xp_earned"].(float64)) != 750 {
		t.Fatalf("expected 750 XP, got %v", stats["total_xp_earned"])
	}
	quests := stats["quests"].(map[string]any)
	if int(quests["active"].(float64)) != 1 || int(quests["complete"].(float64)) != 1 {
		t.Fatalf("quest stats mismatch: %+v", quests)
	}
	t.Logf("Stats: sessions=%v, xp=%v, quests=%+v", stats["session_count"], stats["total_xp_earned"], quests)
}

func TestGraphData(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Graph Hero", "race": "Tiefling", "class": "Warlock"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.get(fmt.Sprintf("/api/characters/%d/graph", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("graph failed: %d - %s", resp.Code, resp.Body.String())
	}
	var gdata map[string]any
	readJSON(resp, &gdata)
	nodes := gdata["nodes"].([]any)
	if len(nodes) < 1 {
		t.Fatal("expected at least 1 graph node")
	}
	t.Logf("Graph: %d nodes, %d edges", len(nodes), len(gdata["edges"].([]any)))
}

func TestCampaignGraphData(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Graph Campaign", "description": "For graph testing"})
	var camp map[string]any
	readJSON(resp, &camp)
	campID := int(camp["id"].(float64))

	resp = tc.post("/api/characters", map[string]any{
		"name": "Graph Hero", "race": "Elf", "class": "Ranger",
		"campaign_id": campID, "level": 3,
	})
	var char map[string]any
	readJSON(resp, &char)
	charID := int(char["id"].(float64))

	resp = tc.post("/api/locations", map[string]any{"name": "Forest", "type": "wilderness"})
	var loc map[string]any
	readJSON(resp, &loc)
	locID := int(loc["id"].(float64))

	tc.post(fmt.Sprintf("/api/characters/%d/locations", charID), map[string]any{
		"location_id": locID, "relationship": "visited",
	})

	resp = tc.post("/api/npcs", map[string]any{"name": "Elara", "race": "Elf", "class": "Druid"})
	var npc map[string]any
	readJSON(resp, &npc)
	npcID := int(npc["id"].(float64))

	tc.post(fmt.Sprintf("/api/characters/%d/npcs", charID), map[string]any{
		"npc_id": npcID, "relationship": "ally",
	})

	tc.post(fmt.Sprintf("/api/characters/%d/quests", charID), map[string]any{
		"name": "Save the Forest", "status": "active",
	})

	tc.post(fmt.Sprintf("/api/campaigns/%d/wiki", campID), map[string]any{
		"title": "The Great Forest", "content": "Lore about the forest...",
		"campaign_id": campID,
	})

	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/graph", campID), nil)
	if resp.Code != 200 {
		t.Fatalf("campaign graph failed: %d - %s", resp.Code, resp.Body.String())
	}
	var gdata map[string]any
	readJSON(resp, &gdata)
	nodes := gdata["nodes"].([]any)
	edges := gdata["edges"].([]any)
	if len(nodes) < 3 {
		t.Fatalf("expected at least 3 graph nodes (campaign + character + wiki), got %d", len(nodes))
	}
	if len(edges) < 2 {
		t.Fatalf("expected at least 2 edges, got %d", len(edges))
	}
	t.Logf("Campaign graph: %d nodes, %d edges", len(nodes), len(edges))
}

func TestRestAndLevelUp(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Tired Hero", "race": "Human", "class": "Fighter",
		"hp_max": 50, "hp_current": 10, "level": 4,
		"str": 16, "dex": 12, "con": 14, "int": 10, "wis": 12, "cha": 10,
		"hit_dice": "1d10", "hit_dice_current": 4,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Short rest with 2 hit dice
	resp = tc.post(fmt.Sprintf("/api/characters/%d/rest", cid), map[string]any{"rest_type": "short", "hit_dice_count": 2})
	if resp.Code != 200 {
		t.Fatalf("short rest failed: %d - %s", resp.Code, resp.Body.String())
	}
	var restResult map[string]any
	readJSON(resp, &restResult)
	hpHealed := int(restResult["hp_healed"].(float64))
	if hpHealed < 2 {
		t.Fatalf("short rest should heal >= 2 HP with 2 HD (CON +2 min 1 each): %d", hpHealed)
	}
	t.Logf("Short rest healed: %d", hpHealed)

	// Verify hit dice were decremented
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["hit_dice_current"].(float64)) != 2 {
		t.Fatalf("expected hit_dice_current=2 after spending 2, got %v", char["hit_dice_current"])
	}

	// Long rest should recover HP and half hit dice
	resp = tc.post(fmt.Sprintf("/api/characters/%d/rest", cid), map[string]any{"rest_type": "long"})
	if resp.Code != 200 {
		t.Fatalf("long rest failed: %d", resp.Code)
	}
	readJSON(resp, &restResult)

	// After long rest at level 4, recover 2 HD (level/2), so total should be 2+2=4
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["hp_current"].(float64)) != 50 {
		t.Fatalf("expected full HP after long rest: %v", char["hp_current"])
	}
	if int(char["hit_dice_current"].(float64)) != 4 {
		t.Fatalf("expected hit_dice_current=4 after long rest (recover 2 of 4), got %v", char["hit_dice_current"])
	}

	// Level up
	resp = tc.post(fmt.Sprintf("/api/characters/%d/levelup", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("level up failed: %d - %s", resp.Code, resp.Body.String())
	}
	var lvl map[string]any
	readJSON(resp, &lvl)
	if int(lvl["new_level"].(float64)) != 5 {
		t.Fatalf("expected level 5, got %v", lvl["new_level"])
	}
	if int(lvl["hp_gain"].(float64)) < 1 {
		t.Fatal("expected HP gain > 0")
	}
	// Level up should increment hit dice
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	if int(char["hit_dice_current"].(float64)) != 5 {
		t.Fatalf("expected hit_dice_current=5 after level up, got %v", char["hit_dice_current"])
	}
	t.Logf("Leveled up: %+v", lvl)
}

func TestImportExport(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Export Test", "race": "Elf", "class": "Ranger",
		"str": 12, "dex": 18, "con": 14, "int": 10, "wis": 16, "cha": 8,
		"hp_max": 44, "hp_current": 44, "level": 5,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Add spellcasting
	tc.put(fmt.Sprintf("/api/characters/%d/spellcasting", cid), map[string]any{
		"ability": "wis", "save_dc": 15, "attack_bonus": 7,
	})
	// Add spell
	tc.post(fmt.Sprintf("/api/characters/%d/spells", cid), map[string]any{
		"name": "Hunter's Mark", "level": 1, "school": "Divination",
	})
	// Add item
	tc.post(fmt.Sprintf("/api/characters/%d/inventory", cid), map[string]any{
		"name": "Longbow", "category": "weapon", "quantity": 1,
		"damage_dice": "1d8", "damage_type": "piercing",
	})
	// Currency
	tc.put(fmt.Sprintf("/api/characters/%d/currency", cid), map[string]any{"gp": 75, "sp": 30, "cp": 0, "ep": 0, "pp": 0})
	// Proficiency
	tc.post("/api/proficiencies", map[string]any{"character_id": cid, "name": "Perception", "type": "skill"})
	// Feature
	tc.post(fmt.Sprintf("/api/characters/%d/features", cid), map[string]any{
		"name": "Favored Enemy", "description": "Bonus vs humanoids", "source": "Class", "level_gained": 1,
	})

	// Export
	resp = tc.get(fmt.Sprintf("/api/characters/%d/export", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("export failed: %d", resp.Code)
	}
	var exportData map[string]any
	readJSON(resp, &exportData)
	if exportData["name"] != "Export Test" {
		t.Fatal("export name mismatch")
	}
	if _, ok := exportData["spells"]; !ok {
		t.Fatal("export missing spells")
	}

	// Print
	resp = tc.get(fmt.Sprintf("/api/characters/%d/print", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("print failed: %d", resp.Code)
	}
	printBody := resp.Body.String()
	if !strings.Contains(printBody, "Export Test") {
		t.Fatal("print output missing character name")
	}
	t.Logf("Print output length: %d", len(printBody))
}

func TestAdminUsers(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create user
	resp := tc.post("/api/admin/users", map[string]any{
		"username": "player1", "password": "playerpass", "role": "user", "email": "player1@example.com",
	})
	if resp.Code != 201 {
		t.Fatalf("create user failed: %d - %s", resp.Code, resp.Body.String())
	}
	var user map[string]any
	readJSON(resp, &user)
	uid := int(user["id"].(float64))

	// List
	resp = tc.get("/api/admin/users", nil)
	if resp.Code != 200 {
		t.Fatalf("list users failed: %d", resp.Code)
	}

	// Verify user has email in list
	var users []map[string]any
	readJSON(resp, &users)
	var found bool
	for _, u := range users {
		if int(u["id"].(float64)) == uid {
			if u["email"] != "player1@example.com" {
				t.Fatalf("expected email player1@example.com, got %v", u["email"])
			}
			found = true
		}
	}
	if !found {
		t.Fatal("created user not found in list")
	}

	// Update
	resp = tc.put(fmt.Sprintf("/api/admin/users/%d", uid), map[string]any{
		"username": "player1_updated", "display_name": "Player One", "role": "user", "email": "player1_new@example.com",
	})
	if resp.Code != 200 {
		t.Fatalf("update user failed: %d", resp.Code)
	}

	// Reset password
	resp = tc.put(fmt.Sprintf("/api/admin/users/%d/password", uid), map[string]any{
		"password": "newpassword",
	})
	if resp.Code != 200 {
		t.Fatalf("reset password failed: %d", resp.Code)
	}

	// Delete
	resp = tc.del(fmt.Sprintf("/api/admin/users/%d", uid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete user failed: %d", resp.Code)
	}
}

func TestCompendiumMonsters(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// List all compendium monsters
	resp := tc.get("/api/compendium-monsters", nil)
	if resp.Code != 200 {
		t.Fatalf("list compendium monsters failed: %d - %s", resp.Code, resp.Body.String())
	}
	var monsters []map[string]any
	readJSON(resp, &monsters)
	if len(monsters) < 10 {
		t.Fatalf("expected >=10 compendium monsters, got %d", len(monsters))
	}
	t.Logf("Total compendium monsters: %d", len(monsters))

	// Verify monster fields
	first := monsters[0]
	requiredFields := []string{"id", "name", "type", "size", "ac", "hp", "str", "dex", "con", "cr"}
	for _, field := range requiredFields {
		if _, ok := first[field]; !ok {
			t.Fatalf("monster missing required field: %s", field)
		}
	}

	// Get a specific monster by ID
	monsterID := int64(first["id"].(float64))
	resp = tc.get(fmt.Sprintf("/api/compendium-monsters/%d", monsterID), nil)
	if resp.Code != 200 {
		t.Fatalf("get compendium monster %d failed: %d - %s", monsterID, resp.Code, resp.Body.String())
	}
	var monster map[string]any
	readJSON(resp, &monster)
	if monster["name"] != first["name"] {
		t.Fatalf("expected monster name %q, got %q", first["name"], monster["name"])
	}
	if _, ok := monster["special_abilities"]; !ok {
		t.Fatal("monster detail missing special_abilities")
	}
	if _, ok := monster["actions"]; !ok {
		t.Fatal("monster detail missing actions")
	}
}

func TestSeedData(t *testing.T) {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM compendium_races").Scan(&count)
	if count < 5 {
		t.Fatalf("expected >=5 races, got %d", count)
	}
	db.DB.QueryRow("SELECT COUNT(*) FROM compendium_classes").Scan(&count)
	if count < 5 {
		t.Fatalf("expected >=5 classes, got %d", count)
	}
	db.DB.QueryRow("SELECT COUNT(*) FROM compendium_spells").Scan(&count)
	if count < 200 {
		t.Fatalf("expected >=200 spells, got %d", count)
	}
	t.Logf("Total spells seeded: %d", count)

	db.DB.QueryRow("SELECT COUNT(*) FROM compendium_monsters").Scan(&count)
	if count < 10 {
		t.Fatalf("expected >=10 monsters, got %d", count)
	}
	t.Logf("Total monsters seeded: %d", count)

	// Verify system/source fields are populated
	var sys, src string
	db.DB.QueryRow("SELECT system, source FROM compendium_races LIMIT 1").Scan(&sys, &src)
	if sys != "dnd5e" || src != "srd" {
		t.Errorf("expected race system=dnd5e source=srd, got system=%q source=%q", sys, src)
	}
	db.DB.QueryRow("SELECT system, source FROM compendium_spells LIMIT 1").Scan(&sys, &src)
	if sys != "dnd5e" || src != "srd" {
		t.Errorf("expected spell system=dnd5e source=srd, got system=%q source=%q", sys, src)
	}
}
