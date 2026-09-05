package main

import (
	"fmt"
	"testing"
)

func TestShareLinkExpiration(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a character
	resp := tc.post("/api/characters", map[string]any{
		"name": "Expiring Hero", "race": "Human", "class": "Fighter",
	})
	if resp.Code != 201 {
		t.Fatalf("create character failed: %d", resp.Code)
	}
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create share link with 1 hour expiration
	resp = tc.post("/api/share", map[string]any{
		"entity_type": "character",
		"entity_id":   cid,
		"expires_in":  1,
	})
	if resp.Code != 201 {
		t.Fatalf("create expiring share link failed: %d - %s", resp.Code, resp.Body.String())
	}
	var shareResp map[string]any
	readJSON(resp, &shareResp)
	token := shareResp["token"].(string)
	if shareResp["expires_at"] == "" {
		t.Fatal("expected expires_at for expiring share link")
	}
	t.Logf("Expiring share link: token=%s, expires_at=%s", token, shareResp["expires_at"])

	// Token should work immediately
	tc2 := newTestClient()
	resp = tc2.get("/api/share/"+token, nil)
	if resp.Code != 200 {
		t.Fatalf("expiring share link should work immediately: %d", resp.Code)
	}
}

// ─── Edge Cases & Permission Tests ───

func TestDashboardEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Dashboard for non-existent campaign
	resp := tc.get("/api/campaigns/99999/dashboard", nil)
	if resp.Code != 200 {
		t.Logf("Dashboard for non-existent campaign: %d (expected 200 with empty data)", resp.Code)
	}

	// Dashboard for campaign with no data
	resp = tc.post("/api/campaigns", map[string]any{"name": "Empty Dash Camp"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/dashboard", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("dashboard for empty campaign failed: %d", resp.Code)
	}
	var dash map[string]any
	readJSON(resp, &dash)
	if int(dash["active_quests"].(float64)) != 0 {
		t.Fatal("expected 0 active quests for empty campaign")
	}
	t.Logf("Empty dashboard: 0 quests, 0 chars, 0 events")
}

func TestResourceEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Resource Edge", "race": "Human", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create resource with zero values
	resp = tc.post(fmt.Sprintf("/api/characters/%d/resources", cid), map[string]any{
		"name": "Empty Resource", "current": 0, "max": 0,
		"short_rest_recovery": 0, "long_rest_recovery": 0,
		"icon": "fa-bolt", "sort_order": 0,
	})
	if resp.Code != 201 {
		t.Fatalf("create zero resource failed: %d", resp.Code)
	}

	// Invalid recover-resources
	resp = tc.post(fmt.Sprintf("/api/characters/%d/recover-resources", cid), map[string]any{"rest_type": "invalid"})
	if resp.Code == 200 {
		t.Log("Invalid rest type accepted")
	}

	// Delete non-existent resource
	resp = tc.del("/api/resources/99999", nil)
	if resp.Code != 200 {
		t.Fatalf("delete non-existent resource should return 200, got %d", resp.Code)
	}
}

func TestHomebrewEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Invalid type
	resp := tc.get("/api/homebrew/invalid", nil)
	if resp.Code != 400 {
		t.Fatalf("expected 400 for invalid homebrew type, got %d", resp.Code)
	}

	// Create then verify source='homebrew'
	resp = tc.post("/api/homebrew/races", map[string]any{
		"name": "Edge Homebrew Race", "speed": 30, "size": "Medium",
	})
	if resp.Code != 201 {
		t.Fatalf("create homebrew race failed: %d", resp.Code)
	}
	var entry map[string]any
	readJSON(resp, &entry)
	eid := int(entry["id"].(float64))

	// Verify it appears in homebrew listing
	resp = tc.get("/api/homebrew/races", nil)
	var items []any
	readJSON(resp, &items)
	found := false
	for _, item := range items {
		it := item.(map[string]any)
		if int(it["id"].(float64)) == eid {
			if it["source"] != "homebrew" {
				t.Fatalf("expected source='homebrew', got %v", it["source"])
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("homebrew entry not found in listing")
	}

	tc.del(fmt.Sprintf("/api/homebrew/races/%d", eid), nil)
}

func TestMapEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Map Edge Camp"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Activate non-existent map
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/maps/99999/activate", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("activate non-existent map should return 200, got %d", resp.Code)
	}

	// Get active map when none exists
	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/maps/active", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get active map with none should return 200, got %d", resp.Code)
	}
	var active map[string]any
	readJSON(resp, &active)
	if active["map"] != nil {
		t.Fatal("expected no active map when none created")
	}

	// Create map then update fog with valid JSON
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/maps", cid), map[string]any{
		"name": "Test Map", "width": 1000, "height": 800, "grid_size": 50,
	})
	var m map[string]any
	readJSON(resp, &m)
	mid := int(m["id"].(float64))

	resp = tc.put(fmt.Sprintf("/api/maps/%d/fog", mid), map[string]any{
		"fog_of_war": `[]`,
	})
	if resp.Code != 200 {
		t.Fatalf("update fog failed: %d", resp.Code)
	}

	tc.del(fmt.Sprintf("/api/maps/%d", mid), nil)
}

func TestCombatLogEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "CL Edge Camp"})
	var camp map[string]any
	readJSON(resp, &camp)

	// Stats for own campaign (should succeed with zero entries)
	resp = tc.get(fmt.Sprintf("/api/combat-log/stats?campaign_id=%d", int(camp["id"].(float64))), nil)
	if resp.Code != 200 {
		t.Fatalf("stats for own campaign should return 200, got %d", resp.Code)
	}

	// Stats for non-existent campaign
	resp = tc.get(fmt.Sprintf("/api/combat-log/stats?campaign_id=%d", 99999), nil)
	if resp.Code != 200 {
		t.Fatalf("stats for non-existent campaign should return 200, got %d", resp.Code)
	}
	var stats map[string]any
	readJSON(resp, &stats)
	if int(stats["total_entries"].(float64)) != 0 {
		t.Fatal("expected 0 entries for non-existent campaign")
	}

	// Stats without campaign_id
	resp = tc.get("/api/combat-log/stats", nil)
	if resp.Code != 400 {
		t.Fatalf("expected 400 for stats without campaign_id, got %d", resp.Code)
	}

	// Log entry without required fields
	resp = tc.post("/api/combat-log", map[string]any{"damage": 5})
	if resp.Code != 201 {
		t.Fatalf("log entry with minimum fields should succeed: %d", resp.Code)
	}
}

func TestQuickRefEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Non-existent section
	resp := tc.get("/api/quickref?section=invalid-section", nil)
	if resp.Code != 404 {
		t.Fatalf("expected 404 for invalid section, got %d", resp.Code)
	}

	// Empty section parameter should return all
	resp = tc.get("/api/quickref", nil)
	var sections []any
	readJSON(resp, &sections)
	if len(sections) < 4 {
		t.Fatalf("expected >=4 sections, got %d", len(sections))
	}
}

func TestDowntimeEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "DT Edge", "race": "Human", "class": "Rogue"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create activity with invalid type (should default to 'other')
	resp = tc.post(fmt.Sprintf("/api/characters/%d/downtime", cid), map[string]any{
		"activity_type": "invalid_type", "name": "Weird Activity", "days_required": 5,
	})
	if resp.Code != 201 {
		t.Fatalf("create with invalid type failed: %d", resp.Code)
	}
	var act map[string]any
	readJSON(resp, &act)

	resp = tc.get(fmt.Sprintf("/api/characters/%d/downtime", cid), nil)
	var acts []any
	readJSON(resp, &acts)
	if len(acts) > 0 {
		a := acts[0].(map[string]any)
		if a["activity_type"] != "other" {
			t.Fatalf("expected activity_type='other' for invalid type, got %v", a["activity_type"])
		}
	}

	// Advance completed activity
	tc.del(fmt.Sprintf("/api/downtime/%d", int(act["id"].(float64))), nil)

	// Advance non-existent
	resp = tc.post("/api/downtime/99999/advance", nil)
	if resp.Code != 400 {
		t.Fatalf("expected 400 for advancing non-existent activity, got %d", resp.Code)
	}
}

// ─── Feature 1: Campaign Dashboard ───

func TestCampaignDashboard(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{
		"name": "Dashboard Campaign", "party_name": "The Testers",
		"description": "Testing dashboard",
	})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Add a character
	resp = tc.post("/api/characters", map[string]any{
		"name": "Dash Hero", "race": "Human", "class": "Fighter",
		"hp_max": 30, "hp_current": 25, "campaign_id": cid,
	})
	var char map[string]any
	readJSON(resp, &char)

	// Add a quest
	tc.post(fmt.Sprintf("/api/characters/%d/quests", int(char["id"].(float64))),
		map[string]any{"name": "Test Quest", "status": "active"})

	// Add calendar event
	tc.post("/api/calendar", map[string]any{
		"campaign_id": cid, "title": "Upcoming Session",
		"event_date": "2026-06-01", "event_type": "session",
	})

	// Get dashboard
	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/dashboard", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("dashboard failed: %d - %s", resp.Code, resp.Body.String())
	}
	var dash map[string]any
	readJSON(resp, &dash)
	if dash["name"] != "Dashboard Campaign" {
		t.Fatalf("expected campaign name 'Dashboard Campaign', got %v", dash["name"])
	}
	if int(dash["active_quests"].(float64)) < 1 {
		t.Fatal("expected active quests >= 1")
	}
	t.Logf("Dashboard: %d quests, %d chars", int(dash["active_quests"].(float64)), len(dash["characters"].([]any)))
}

// ─── Feature 2: Character Resources ───

func TestCharacterResources(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Resource Test", "race": "Elf", "class": "Sorcerer"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Create resource
	resp = tc.post(fmt.Sprintf("/api/characters/%d/resources", cid), map[string]any{
		"name": "Sorcery Points", "current": 3, "max": 5,
		"short_rest_recovery": 0, "long_rest_recovery": 1,
		"icon": "fa-star", "sort_order": 1,
	})
	if resp.Code != 201 {
		t.Fatalf("create resource failed: %d - %s", resp.Code, resp.Body.String())
	}

	// List resources
	resp = tc.get(fmt.Sprintf("/api/characters/%d/resources", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list resources failed: %d", resp.Code)
	}
	var resources []any
	readJSON(resp, &resources)
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	r := resources[0].(map[string]any)
	if int(r["current"].(float64)) != 3 || int(r["max"].(float64)) != 5 {
		t.Fatalf("resource values mismatch: %+v", r)
	}

	// Update resource
	resp = tc.put(fmt.Sprintf("/api/resources/%d", int(r["id"].(float64))), map[string]any{
		"name": "Sorcery Points", "current": 5, "max": 5,
		"short_rest_recovery": 0, "long_rest_recovery": 1,
		"icon": "fa-star", "sort_order": 1,
	})
	if resp.Code != 200 {
		t.Fatalf("update resource failed: %d", resp.Code)
	}

	// Recover on long rest
	tc.post(fmt.Sprintf("/api/characters/%d/recover-resources", cid), map[string]any{"rest_type": "long"})

	resp = tc.get(fmt.Sprintf("/api/characters/%d/resources", cid), nil)
	readJSON(resp, &resources)
	r = resources[0].(map[string]any)
	if int(r["current"].(float64)) != 5 {
		t.Fatalf("expected current=5 after long rest, got %v", r["current"])
	}

	tc.del(fmt.Sprintf("/api/resources/%d", int(r["id"].(float64))), nil)
	t.Log("Resource lifecycle test passed")
}

// ─── Feature 3: Homebrew Content ───

func TestHomebrewContentCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Test each type
	types := []string{"races", "classes", "spells", "feats", "backgrounds", "equipment"}
	for _, ctype := range types {
		t.Run(ctype, func(t *testing.T) {
			// Create
			body := map[string]any{
				"name": "Test Homebrew " + ctype,
			}
			if ctype == "races" {
				body["speed"] = 30
				body["size"] = "Medium"
			}
			if ctype == "spells" {
				body["level"] = 1
				body["school"] = "Evocation"
			}
			resp := tc.post("/api/homebrew/"+ctype, body)
			if resp.Code != 201 {
				t.Fatalf("create %s failed: %d - %s", ctype, resp.Code, resp.Body.String())
			}
			var entry map[string]any
			readJSON(resp, &entry)
			eid := int(entry["id"].(float64))

			// List
			resp = tc.get("/api/homebrew/"+ctype, nil)
			if resp.Code != 200 {
				t.Fatalf("list %s failed: %d", ctype, resp.Code)
			}
			var items []any
			readJSON(resp, &items)
			found := false
			for _, item := range items {
				it := item.(map[string]any)
				if int(it["id"].(float64)) == eid {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("created %s entry not found in list", ctype)
			}

			// Update
			resp = tc.put(fmt.Sprintf("/api/homebrew/%s/%d", ctype, eid), map[string]any{
				"name": "Updated " + ctype,
			})
			if resp.Code != 200 {
				t.Fatalf("update %s failed: %d", ctype, resp.Code)
			}

			// Delete
			resp = tc.del(fmt.Sprintf("/api/homebrew/%s/%d", ctype, eid), nil)
			if resp.Code != 200 {
				t.Fatalf("delete %s failed: %d", ctype, resp.Code)
			}
			t.Logf("%s CRUD test passed", ctype)
		})
	}
}

// ─── Feature 4: Campaign Maps ───

func TestCampaignMaps(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Map Test Camp"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Create map
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/maps", cid), map[string]any{
		"name": "World Map", "image_url": "/media/world.jpg",
		"width": 2000, "height": 1500, "grid_size": 50,
	})
	if resp.Code != 201 {
		t.Fatalf("create map failed: %d - %s", resp.Code, resp.Body.String())
	}
	var m map[string]any
	readJSON(resp, &m)
	mid := int(m["id"].(float64))

	// List maps
	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/maps", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list maps failed: %d", resp.Code)
	}
	var maps []any
	readJSON(resp, &maps)
	if len(maps) != 1 {
		t.Fatalf("expected 1 map, got %d", len(maps))
	}

	// Create pin
	resp = tc.post(fmt.Sprintf("/api/maps/%d/pins", mid), map[string]any{
		"name": "Waterdeep", "type": "city", "x": 500, "y": 300,
		"icon": "fa-city", "color": "#b8963e", "description": "City of Splendors",
	})
	if resp.Code != 201 {
		t.Fatalf("create pin failed: %d - %s", resp.Code, resp.Body.String())
	}

	// List pins
	resp = tc.get(fmt.Sprintf("/api/maps/%d/pins", mid), nil)
	if resp.Code != 200 {
		t.Fatalf("list pins failed: %d", resp.Code)
	}
	var pins []any
	readJSON(resp, &pins)
	if len(pins) != 1 {
		t.Fatalf("expected 1 pin, got %d", len(pins))
	}

	// Activate map
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/maps/%d/activate", cid, mid), nil)
	if resp.Code != 200 {
		t.Fatalf("activate map failed: %d", resp.Code)
	}

	// Get active map
	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/maps/active", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get active map failed: %d", resp.Code)
	}
	var active map[string]any
	readJSON(resp, &active)
	if active["map"] == nil {
		t.Fatal("expected active map data")
	}

	// Update fog of war
	resp = tc.put(fmt.Sprintf("/api/maps/%d/fog", mid), map[string]any{
		"fog_of_war": `[[0,0,500,400]]`,
	})
	if resp.Code != 200 {
		t.Fatalf("update fog failed: %d", resp.Code)
	}

	// Cleanup
	pin := pins[0].(map[string]any)
	tc.del(fmt.Sprintf("/api/map-pins/%d", int(pin["id"].(float64))), nil)
	tc.del(fmt.Sprintf("/api/maps/%d", mid), nil)
	t.Log("Campaign maps test passed")
}

// ─── Feature 5: Combat Log ───

func TestCombatLog(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a campaign
	resp := tc.post("/api/campaigns", map[string]any{"name": "Combat Log Camp"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Create combat log entries
	entries := []map[string]any{
		{"campaign_id": cid, "actor_name": "Gandalf", "action": "Attack", "target_name": "Orc", "damage": 12, "damage_type": "fire", "roll_total": 18, "description": "Fireball!"},
		{"campaign_id": cid, "actor_name": "Legolas", "action": "Attack", "target_name": "Orc", "damage": 8, "damage_type": "piercing", "roll_total": 15, "description": "Arrow shot"},
		{"campaign_id": cid, "actor_name": "Aragorn", "action": "Heal", "healing": 10, "description": "Lay on hands"},
	}
	for _, entry := range entries {
		resp = tc.post("/api/combat-log", entry)
		if resp.Code != 201 {
			t.Fatalf("create combat log entry failed: %d - %s", resp.Code, resp.Body.String())
		}
	}

	// List
	resp = tc.get(fmt.Sprintf("/api/combat-log?campaign_id=%d", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list combat log failed: %d", resp.Code)
	}
	var logEntries []any
	readJSON(resp, &logEntries)
	if len(logEntries) < 3 {
		t.Fatalf("expected >=3 log entries, got %d", len(logEntries))
	}

	// Stats
	resp = tc.get(fmt.Sprintf("/api/combat-log/stats?campaign_id=%d", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("combat log stats failed: %d", resp.Code)
	}
	var stats map[string]any
	readJSON(resp, &stats)
	totalDmg := int(stats["total_damage"].(float64))
	if totalDmg != 20 {
		t.Fatalf("expected 20 total damage, got %d", totalDmg)
	}
	t.Logf("Combat log: %d entries, %d damage, %d healing",
		int(stats["total_entries"].(float64)),
		int(stats["total_damage"].(float64)),
		int(stats["total_healing"].(float64)))
}

// ─── Feature 6: Quick Reference ───

func TestQuickReference(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// List all sections
	resp := tc.get("/api/quickref", nil)
	if resp.Code != 200 {
		t.Fatalf("quickref failed: %d", resp.Code)
	}
	var sections []any
	readJSON(resp, &sections)
	if len(sections) < 3 {
		t.Fatalf("expected >=3 quick reference sections, got %d", len(sections))
	}

	// Get specific section
	sectionTests := []string{"conditions", "actions", "damage-types", "skills"}
	for _, s := range sectionTests {
		t.Run(s, func(t *testing.T) {
			resp := tc.get("/api/quickref?section="+s, nil)
			if resp.Code != 200 {
				t.Fatalf("section %s failed: %d", s, resp.Code)
			}
			var sec map[string]any
			readJSON(resp, &sec)
			entries := sec["entries"].([]any)
			if len(entries) < 3 {
				t.Fatalf("section %s has <3 entries", s)
			}
			t.Logf("%s: %d entries", s, len(entries))
		})
	}
}

// ─── Feature 8: Downtime Activities ───

func TestDowntimeActivities(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Downtime Hero", "race": "Dwarf", "class": "Fighter"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Get types
	resp = tc.get("/api/downtime/types", nil)
	if resp.Code != 200 {
		t.Fatalf("downtime types failed: %d", resp.Code)
	}
	var types []any
	readJSON(resp, &types)
	if len(types) < 5 {
		t.Fatalf("expected >=5 downtime types, got %d", len(types))
	}

	// Create activity
	resp = tc.post(fmt.Sprintf("/api/characters/%d/downtime", cid), map[string]any{
		"activity_type": "training", "name": "Learn Dwarven", "description": "Practice Dwarven language",
		"dc": 12, "days_required": 30, "cost_per_day": 2,
	})
	if resp.Code != 201 {
		t.Fatalf("create downtime failed: %d - %s", resp.Code, resp.Body.String())
	}
	var activity map[string]any
	readJSON(resp, &activity)
	aid := int(activity["id"].(float64))

	// List
	resp = tc.get(fmt.Sprintf("/api/characters/%d/downtime", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list downtime failed: %d", resp.Code)
	}
	var list []any
	readJSON(resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(list))
	}

	// Advance day
	resp = tc.post(fmt.Sprintf("/api/downtime/%d/advance", aid), nil)
	if resp.Code != 200 {
		t.Fatalf("advance day failed: %d - %s", resp.Code, resp.Body.String())
	}
	var adv map[string]any
	readJSON(resp, &adv)
	if int(adv["days_completed"].(float64)) != 1 {
		t.Fatalf("expected days_completed=1, got %v", adv["days_completed"])
	}

	// Update
	resp = tc.put(fmt.Sprintf("/api/downtime/%d", aid), map[string]any{
		"activity_type": "training", "name": "Learn Dwarven Updated",
		"description": "Practice Dwarven", "dc": 12, "days_required": 30,
		"days_completed": 1, "cost_per_day": 2, "total_cost": 60,
		"reward": "Dwarven language proficiency", "status": "in-progress", "notes": "",
	})
	if resp.Code != 200 {
		t.Fatalf("update downtime failed: %d", resp.Code)
	}

	// Delete
	tc.del(fmt.Sprintf("/api/downtime/%d", aid), nil)
	t.Log("Downtime activity test passed")
}

// ─── Feature 9: Campaign Recaps ───

func TestCampaignRecaps(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Recap Campaign"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Create a recap
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/recaps", cid), map[string]any{
		"title": "Session 1 Recap", "content": "The party defeated the goblin king and found the ancient artifact.",
	})
	if resp.Code != 201 {
		t.Fatalf("create recap failed: %d - %s", resp.Code, resp.Body.String())
	}
	var recap map[string]any
	readJSON(resp, &recap)
	rid := int(recap["id"].(float64))

	// List recaps
	resp = tc.get(fmt.Sprintf("/api/campaigns/%d/recaps", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("list recaps failed: %d", resp.Code)
	}
	var recaps []any
	readJSON(resp, &recaps)
	if len(recaps) != 1 {
		t.Fatalf("expected 1 recap, got %d", len(recaps))
	}

	// Get single recap
	resp = tc.get(fmt.Sprintf("/api/recaps/%d", rid), nil)
	if resp.Code != 200 {
		t.Fatalf("get recap failed: %d", resp.Code)
	}
	readJSON(resp, &recap)
	if int(recap["word_count"].(float64)) < 5 {
		t.Fatalf("expected word_count >= 5, got %v", recap["word_count"])
	}

	// Update recap
	resp = tc.put(fmt.Sprintf("/api/recaps/%d", rid), map[string]any{
		"title": "Session 1 Recap (Edited)", "content": "The party defeated the goblin king and found the ancient artifact. Updated.",
	})
	if resp.Code != 200 {
		t.Fatalf("update recap failed: %d", resp.Code)
	}

	// Generate recap
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/recaps/generate", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("generate recap failed: %d - %s", resp.Code, resp.Body.String())
	}
	readJSON(resp, &recap)
	t.Logf("Generated recap: %d words, title=%s", int(recap["word_count"].(float64)), recap["title"])

	// Mark as sent
	resp = tc.post(fmt.Sprintf("/api/recaps/%d/mark-sent", rid), nil)
	if resp.Code != 200 {
		t.Fatalf("mark sent failed: %d", resp.Code)
	}

	// Delete
	tc.del(fmt.Sprintf("/api/recaps/%d", rid), nil)
	t.Log("Campaign recaps test passed")
}

func TestCampaignRecapEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/campaigns", map[string]any{"name": "Edge Recap Camp"})
	var camp map[string]any
	readJSON(resp, &camp)
	cid := int(camp["id"].(float64))

	// Get non-existent recap
	resp = tc.get("/api/recaps/99999", nil)
	if resp.Code != 404 {
		t.Fatalf("expected 404 for non-existent recap, got %d", resp.Code)
	}

	// Create recap with empty content
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/recaps", cid), map[string]any{
		"title": "Empty Recap", "content": "",
	})
	if resp.Code != 201 {
		t.Fatalf("create empty recap failed: %d", resp.Code)
	}
	var created map[string]any
	readJSON(resp, &created)
	rid := int(created["id"].(float64))

	// Fetch it back to verify word_count
	resp = tc.get(fmt.Sprintf("/api/recaps/%d", rid), nil)
	var recap map[string]any
	readJSON(resp, &recap)
	if recap["word_count"] != nil {
		wc := int(recap["word_count"].(float64))
		if wc != 0 {
			t.Fatalf("expected word_count=0 for empty content, got %d", wc)
		}
	} else {
		t.Log("word_count is nil for empty recap")
	}

	// Generate recap with no data (empty campaign)
	resp = tc.post(fmt.Sprintf("/api/campaigns/%d/recaps/generate", 99998), nil)
	if resp.Code == 200 {
		readJSON(resp, &recap)
		if recap["title"] != nil {
			t.Logf("Generated recap for empty campaign: title=%s", recap["title"])
		}
	}

	tc.del(fmt.Sprintf("/api/recaps/%d", rid), nil)
	t.Log("Recap edge case tests passed")
}

// ─── Feature 10: Level Up Planner ───

func TestLevelUpPlanner(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{
		"name": "Planning Hero", "race": "Half-Elf", "class": "Warlock",
		"level": 3, "str": 8, "dex": 14, "con": 14, "int": 10, "wis": 10, "cha": 18,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Get level suggestions
	resp = tc.get(fmt.Sprintf("/api/characters/%d/level-suggestions", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("level suggestions failed: %d - %s", resp.Code, resp.Body.String())
	}
	var suggestions []any
	readJSON(resp, &suggestions)
	if len(suggestions) < 1 {
		t.Fatalf("expected suggestions, got %d", len(suggestions))
	}
	t.Logf("Level suggestions: %d levels", len(suggestions))

	// Save a level plan
	planData := []map[string]any{
		{"level": 4, "feat": "War Caster", "class": "Warlock"},
		{"level": 5, "class": "Warlock", "class_feature": "Pact Boon improvement"},
		{"level": 8, "feat": "ASI (CHA +2)", "class": "Warlock"},
	}
	resp = tc.post(fmt.Sprintf("/api/characters/%d/level-plan", cid), map[string]any{
		"character_id": cid, "target_level": 20, "plan_data": planData, "notes": "Focus on CHA",
	})
	if resp.Code != 201 {
		t.Fatalf("save level plan failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get plan
	resp = tc.get(fmt.Sprintf("/api/characters/%d/level-plan", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get level plan failed: %d", resp.Code)
	}
	var plan map[string]any
	readJSON(resp, &plan)
	if int(plan["target_level"].(float64)) != 20 {
		t.Fatalf("expected target_level=20, got %v", plan["target_level"])
	}
	entries := plan["plan_data"].([]any)
	if len(entries) != 3 {
		t.Fatalf("expected 3 plan entries, got %d", len(entries))
	}

	// Update plan
	resp = tc.post(fmt.Sprintf("/api/characters/%d/level-plan", cid), map[string]any{
		"character_id": cid, "target_level": 20, "plan_data": planData, "notes": "Updated plan",
	})
	if resp.Code != 200 {
		t.Fatalf("update level plan failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Delete plan
	resp = tc.del(fmt.Sprintf("/api/characters/%d/level-plan", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete level plan failed: %d", resp.Code)
	}
	t.Log("Level up planner test passed")
}
