package main

import (
	"fmt"
	"slices"
	"strconv"
	"testing"
)

func TestLevelUpPlannerEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Get plan for character with no plan
	resp := tc.post("/api/characters", map[string]any{
		"name": "No Plan Hero", "race": "Human", "class": "Fighter", "level": 1,
		"str": 15, "dex": 14, "con": 13, "int": 10, "wis": 12, "cha": 8,
	})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	resp = tc.get(fmt.Sprintf("/api/characters/%d/level-plan", cid), nil)
	if resp.Code != 200 {
		t.Fatalf("get plan for no-plan char failed: %d", resp.Code)
	}
	var plan map[string]any
	readJSON(resp, &plan)
	if plan["plan_data"] == nil {
		t.Fatal("expected empty plan_data for char with no plan")
	}

	// ASI suggestions at appropriate levels
	resp = tc.get(fmt.Sprintf("/api/characters/%d/level-suggestions", cid), nil)
	var suggestions []any
	readJSON(resp, &suggestions)
	asiLevels := []int{}
	for _, s := range suggestions {
		sug := s.(map[string]any)
		if sug["has_asi"] == true {
			asiLevels = append(asiLevels, int(sug["level"].(float64)))
		}
	}
	if len(asiLevels) == 0 {
		t.Fatal("expected at least 1 ASI suggestion")
	}
	t.Logf("ASI suggestion levels: %v", asiLevels)
	expectedASIs := []int{4, 8, 12, 16, 19}
	for _, exp := range expectedASIs {
		found := slices.Contains(asiLevels, exp)
		if !found {
			t.Fatalf("expected ASI at level %d, got levels %v", exp, asiLevels)
		}
	}

	// Save plan with empty data
	resp = tc.post(fmt.Sprintf("/api/characters/%d/level-plan", cid), map[string]any{
		"character_id": cid, "target_level": 20, "plan_data": []any{}, "notes": "",
	})
	if resp.Code != 201 && resp.Code != 200 {
		t.Fatalf("save empty plan failed: %d", resp.Code)
	}
	t.Log("Level up planner edge case tests passed")
}

// ─── One-Shot Generator Tests ───

func TestGenerateAdventureHook(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/adventure-hook", nil)
	if resp.Code != 200 {
		t.Fatalf("generate adventure hook failed: %d", resp.Code)
	}
	var data map[string]any
	readJSON(resp, &data)
	required := []string{"hook_name", "hook_type", "villain", "macguffin", "stakes", "location_hint", "twist"}
	for _, k := range required {
		if _, ok := data[k]; !ok {
			t.Errorf("missing field: %s", k)
		}
	}
	t.Logf("Adventure hook: %v", data["hook_name"])
}

func TestGenerateDungeonDressing(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/dungeon-dressing", nil)
	if resp.Code != 200 {
		t.Fatalf("generate dungeon dressing failed: %d", resp.Code)
	}
	var data map[string]any
	readJSON(resp, &data)
	required := []string{"room_type", "size", "sound", "smell", "light", "debris"}
	for _, k := range required {
		if _, ok := data[k]; !ok {
			t.Errorf("missing field: %s", k)
		}
	}
	t.Logf("Dungeon dressing: %v / %v", data["room_type"], data["size"])
}

func TestGenerateTavern(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/tavern", nil)
	if resp.Code != 200 {
		t.Fatalf("generate tavern failed: %d", resp.Code)
	}
	var data map[string]any
	readJSON(resp, &data)
	required := []string{"name", "proprietor", "clientele", "specialty_drink", "atmosphere", "prices", "rumors"}
	for _, k := range required {
		if _, ok := data[k]; !ok {
			t.Errorf("missing field: %s", k)
		}
	}
	if _, ok := data["name"].(string); !ok || data["name"].(string) == "" {
		t.Error("expected non-empty tavern name")
	}
	clientele := data["clientele"].([]any)
	if len(clientele) < 2 {
		t.Error("expected at least 2 clientele entries")
	}
	rumors := data["rumors"].([]any)
	if len(rumors) < 1 {
		t.Error("expected at least 1 rumor")
	}
	t.Logf("Tavern: %v", data["name"])
}

func TestGenerateUrbanEncounter(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/urban-encounter", nil)
	if resp.Code != 200 {
		t.Fatalf("generate urban encounter failed: %d", resp.Code)
	}
	var data map[string]any
	readJSON(resp, &data)
	required := []string{"theme", "npc", "description", "complication", "possible_resolution"}
	for _, k := range required {
		if _, ok := data[k]; !ok {
			t.Errorf("missing field: %s", k)
		}
	}
	t.Logf("Urban encounter: %v / %v", data["theme"], data["npc"])
}

func TestGenerateRoadEncounter(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.get("/api/generate/road-encounter", nil)
	if resp.Code != 200 {
		t.Fatalf("generate road encounter failed: %d", resp.Code)
	}
	var data map[string]any
	readJSON(resp, &data)
	required := []string{"terrain", "encounter_type", "description", "creatures", "loot_hint", "complication"}
	for _, k := range required {
		if _, ok := data[k]; !ok {
			t.Errorf("missing field: %s", k)
		}
	}
	t.Logf("Road encounter: %v / %v", data["terrain"], data["encounter_type"])
}

// ─── One-Shot Adventure Tests ───

func TestOneShotAdventureCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a one-shot adventure
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "The Lost Temple",
		"premise":           "An ancient temple has been discovered in the jungle.",
		"hook":              "The party is hired by a historian to explore the ruins.",
		"template":          "custom",
		"estimated_minutes": 180,
		"difficulty":        "medium",
		"notes":             "Prepare jungle encounters",
	})
	if resp.Code != 201 {
		t.Fatalf("create oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}
	var created map[string]any
	readJSON(resp, &created)
	aid := int(created["id"].(float64))
	if aid == 0 {
		t.Fatal("expected non-zero id")
	}

	// Get the one-shot adventure
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	if resp.Code != 200 {
		t.Fatalf("get oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}
	var adventure map[string]any
	readJSON(resp, &adventure)
	if adventure["title"] != "The Lost Temple" {
		t.Fatalf("expected title 'The Lost Temple', got %v", adventure["title"])
	}

	// Update the one-shot adventure
	resp = tc.put("/api/oneshot-adventures/"+strconv.Itoa(aid), map[string]any{
		"title":             "The Lost Temple - Updated",
		"premise":           "Updated premise",
		"hook":              "Updated hook",
		"template":          "five_room_dungeon",
		"estimated_minutes": 240,
		"difficulty":        "hard",
		"notes":             "Updated notes",
	})
	if resp.Code != 200 {
		t.Fatalf("update oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get updated adventure
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	readJSON(resp, &adventure)
	if adventure["title"] != "The Lost Temple - Updated" {
		t.Fatalf("expected updated title")
	}

	// List one-shot adventures
	resp = tc.get("/api/oneshot-adventures", nil)
	if resp.Code != 200 {
		t.Fatalf("list oneshots failed: %d - %s", resp.Code, resp.Body.String())
	}
	var list []any
	readJSON(resp, &list)
	if len(list) < 1 {
		t.Fatal("expected at least 1 one-shot adventure")
	}

	// Add an act
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/acts", map[string]any{
		"title":             "Act 1: Discovery",
		"description":       "The party arrives at the temple.",
		"estimated_minutes": 60,
		"number":            1,
	})
	if resp.Code != 201 {
		t.Fatalf("create act failed: %d - %s", resp.Code, resp.Body.String())
	}
	var actResult map[string]any
	readJSON(resp, &actResult)
	actID := int(actResult["id"].(float64))
	if actID == 0 {
		t.Fatal("expected non-zero act id")
	}

	// Add scenes to act
	resp = tc.post("/api/oneshot-acts/"+strconv.Itoa(actID)+"/scenes", map[string]any{
		"title":             "Scene 1: The Entrance",
		"description":       "The party approaches the temple entrance.",
		"scene_type":        "exploration",
		"estimated_minutes": 20,
		"number":            1,
	})
	if resp.Code != 201 {
		t.Fatalf("create scene failed: %d - %s", resp.Code, resp.Body.String())
	}

	resp = tc.post("/api/oneshot-acts/"+strconv.Itoa(actID)+"/scenes", map[string]any{
		"title":             "Scene 2: Temple Guardians",
		"description":       "Golem guardians awaken.",
		"scene_type":        "combat",
		"estimated_minutes": 30,
		"number":            2,
	})
	if resp.Code != 201 {
		t.Fatalf("create scene 2 failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Get full adventure with acts and scenes
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	if resp.Code != 200 {
		t.Fatalf("get detailed oneshot failed: %d", resp.Code)
	}
	readJSON(resp, &adventure)
	acts := adventure["acts"].([]any)
	if len(acts) != 1 {
		t.Fatalf("expected 1 act, got %d", len(acts))
	}
	act := acts[0].(map[string]any)
	scenes := act["scenes"].([]any)
	if len(scenes) != 2 {
		t.Fatalf("expected 2 scenes, got %d", len(scenes))
	}

	// Update act
	resp = tc.put("/api/oneshot-acts/"+strconv.Itoa(actID), map[string]any{
		"title":             "Act 1: Discovery (Updated)",
		"description":       "Updated description",
		"estimated_minutes": 90,
		"number":            1,
	})
	if resp.Code != 200 {
		t.Fatalf("update act failed: %d", resp.Code)
	}

	// Delete scene
	var sceneID int
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	readJSON(resp, &adventure)
	acts = adventure["acts"].([]any)
	if len(acts) > 0 {
		act = acts[0].(map[string]any)
		scenes = act["scenes"].([]any)
		if len(scenes) > 0 {
			sceneID = int(scenes[0].(map[string]any)["id"].(float64))
		}
	}
	if sceneID > 0 {
		resp = tc.del("/api/oneshot-scenes/"+strconv.Itoa(sceneID), nil)
		if resp.Code != 200 {
			t.Fatalf("delete scene failed: %d", resp.Code)
		}
	}

	// Delete act
	resp = tc.del("/api/oneshot-acts/"+strconv.Itoa(actID), nil)
	if resp.Code != 200 {
		t.Fatalf("delete act failed: %d", resp.Code)
	}

	// Delete the one-shot
	resp = tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Verify deletion
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	if resp.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", resp.Code)
	}

	t.Log("One-shot adventure CRUD test passed")
}

func TestOneShotAdventureGeneration(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Generate a 5-room dungeon
	resp := tc.post("/api/oneshot-adventures/generate", map[string]any{
		"title":             "Generated Dungeon",
		"template":          "five_room_dungeon",
		"difficulty":        "medium",
		"estimated_minutes": 120,
	})
	if resp.Code != 201 {
		t.Fatalf("generate oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}
	var created map[string]any
	readJSON(resp, &created)
	aid := int(created["id"].(float64))
	if aid == 0 {
		t.Fatal("expected non-zero id")
	}

	// Verify the generated structure
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	if resp.Code != 200 {
		t.Fatalf("get generated oneshot failed: %d", resp.Code)
	}
	var adventure map[string]any
	readJSON(resp, &adventure)
	acts := adventure["acts"].([]any)
	if len(acts) != 5 {
		t.Fatalf("expected 5 acts (rooms) in five_room_dungeon, got %d", len(acts))
	}
	if adventure["template"] != "five_room_dungeon" {
		t.Fatalf("expected template 'five_room_dungeon', got %v", adventure["template"])
	}

	t.Log("One-shot adventure generation test passed")
}

func TestOneShotAdventureLinks(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a campaign for context
	resp := tc.post("/api/campaigns", map[string]any{
		"name": "Test Campaign for One-Shot",
	})
	if resp.Code != 201 {
		t.Fatalf("create campaign failed: %d", resp.Code)
	}
	var camp map[string]any
	readJSON(resp, &camp)
	campaignID := int(camp["id"].(float64))

	// Create a location
	resp = tc.post("/api/locations", map[string]any{
		"name": "One-Shot Temple",
		"type": "dungeon",
	})
	if resp.Code != 201 {
		t.Fatalf("create location failed: %d", resp.Code)
	}
	var locResult map[string]any
	readJSON(resp, &locResult)
	locID := int(locResult["id"].(float64))

	// Create an NPC
	resp = tc.post("/api/npcs", map[string]any{
		"name": "Historian Marcus",
	})
	if resp.Code != 201 {
		t.Fatalf("create npc failed: %d", resp.Code)
	}
	var npcResult map[string]any
	readJSON(resp, &npcResult)
	npcID := int(npcResult["id"].(float64))

	// Create an encounter
	resp = tc.post("/api/encounters", map[string]any{
		"name":       "Temple Guardians Encounter",
		"difficulty": "medium",
	})
	if resp.Code != 201 {
		t.Fatalf("create encounter failed: %d", resp.Code)
	}
	var encResult map[string]any
	readJSON(resp, &encResult)
	encID := int(encResult["id"].(float64))

	// Create a one-shot linked to the campaign
	resp = tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Linked Adventure",
		"campaign_id":       campaignID,
		"template":          "custom",
		"estimated_minutes": 180,
	})
	if resp.Code != 201 {
		t.Fatalf("create oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}
	var created map[string]any
	readJSON(resp, &created)
	aid := int(created["id"].(float64))

	// Link NPC
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/npcs", map[string]any{
		"npc_id": npcID,
		"role":   "quest giver",
	})
	if resp.Code != 200 {
		t.Fatalf("link npc failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Link Location
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/locations", map[string]any{
		"location_id": locID,
	})
	if resp.Code != 200 {
		t.Fatalf("link location failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Link Encounter
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/encounters", map[string]any{
		"encounter_id": encID,
	})
	if resp.Code != 200 {
		t.Fatalf("link encounter failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Verify links in detail
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	if resp.Code != 200 {
		t.Fatalf("get linked oneshot failed: %d", resp.Code)
	}
	var adventure map[string]any
	readJSON(resp, &adventure)
	npcs := adventure["npcs"].([]any)
	if len(npcs) != 1 {
		t.Fatalf("expected 1 linked NPC, got %d", len(npcs))
	}
	npc := npcs[0].(map[string]any)
	if npc["role"] != "quest giver" {
		t.Fatalf("expected NPC role 'quest giver', got %v", npc["role"])
	}

	locs := adventure["locations"].([]any)
	if len(locs) != 1 {
		t.Fatalf("expected 1 linked location, got %d", len(locs))
	}

	encs := adventure["encounters"].([]any)
	if len(encs) != 1 {
		t.Fatalf("expected 1 linked encounter, got %d", len(encs))
	}

	// Unlink NPC
	resp = tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/npcs/"+strconv.Itoa(npcID), nil)
	if resp.Code != 200 {
		t.Fatalf("unlink npc failed: %d", resp.Code)
	}

	// Unlink Location
	resp = tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/locations/"+strconv.Itoa(locID), nil)
	if resp.Code != 200 {
		t.Fatalf("unlink location failed: %d", resp.Code)
	}

	// Unlink Encounter
	resp = tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/encounters/"+strconv.Itoa(encID), nil)
	if resp.Code != 200 {
		t.Fatalf("unlink encounter failed: %d", resp.Code)
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	tc.del("/api/encounters/"+strconv.Itoa(encID), nil)
	tc.del("/api/npcs/"+strconv.Itoa(npcID), nil)
	tc.del("/api/locations/"+strconv.Itoa(locID), nil)
	tc.del("/api/campaigns/"+strconv.Itoa(campaignID), nil)

	t.Log("One-shot adventure links test passed")
}

func TestOneShotAdventureEdgeCases(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Filter by campaign
	resp := tc.get("/api/oneshot-adventures?campaign_id=9999", nil)
	if resp.Code != 200 {
		t.Fatalf("list by campaign filter failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Generate without specifying fields
	resp = tc.post("/api/oneshot-adventures/generate", map[string]any{
		"template": "five_room_dungeon",
	})
	if resp.Code != 201 {
		t.Fatalf("generate minimal oneshot failed: %d - %s", resp.Code, resp.Body.String())
	}
	var created map[string]any
	readJSON(resp, &created)
	aid := int(created["id"].(float64))
	if aid == 0 {
		t.Fatal("expected non-zero id for minimal generate")
	}

	// Get NPC links for empty adventure
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/npcs", nil)
	if resp.Code != 200 {
		t.Fatalf("get npcs for empty adventure failed: %d", resp.Code)
	}
	var npcs []any
	readJSON(resp, &npcs)
	if len(npcs) != 0 {
		t.Fatalf("expected 0 NPCs for empty adventure, got %d", len(npcs))
	}

	// Get location links for empty adventure
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/locations", nil)
	if resp.Code != 200 {
		t.Fatalf("get locations for empty adventure failed: %d", resp.Code)
	}

	// Get encounter links for empty adventure
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/encounters", nil)
	if resp.Code != 200 {
		t.Fatalf("get encounters for empty adventure failed: %d", resp.Code)
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("One-shot adventure edge case tests passed")
}

// ─── Session Pacing Tests ───

func TestSessionPacingLifecycle(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create an adventure first
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Pacing Test Adventure",
		"template":          "five_room_dungeon",
		"difficulty":        "medium",
		"estimated_minutes": 120,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Start pacing session
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/pacing/start", nil)
	if resp.Code != 201 {
		t.Fatalf("start pacing failed: %d - %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	readJSON(resp, &result)
	sessionID := int(result["id"].(float64))
	if sessionID == 0 {
		t.Fatal("expected non-zero session id")
	}

	// Get pacing session
	resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
	if resp.Code != 200 {
		t.Fatalf("get pacing session failed: %d", resp.Code)
	}
	var session map[string]any
	readJSON(resp, &session)
	if session["status"] != "running" {
		t.Fatalf("expected running status, got %v", session["status"])
	}
	if session["adventure_title"] != "Pacing Test Adventure" {
		t.Fatalf("expected adventure title 'Pacing Test Adventure', got %v", session["adventure_title"])
	}

	// Pause session
	resp = tc.post("/api/session-pacing/"+strconv.Itoa(sessionID)+"/pause", nil)
	if resp.Code != 200 {
		t.Fatalf("pause pacing failed: %d", resp.Code)
	}

	resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
	readJSON(resp, &session)
	if session["status"] != "paused" {
		t.Fatalf("expected paused status, got %v", session["status"])
	}

	// Resume session
	resp = tc.post("/api/session-pacing/"+strconv.Itoa(sessionID)+"/resume", nil)
	if resp.Code != 200 {
		t.Fatalf("resume pacing failed: %d", resp.Code)
	}

	resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
	readJSON(resp, &session)
	if session["status"] != "running" {
		t.Fatalf("expected running status after resume, got %v", session["status"])
	}

	// Tick timer
	resp = tc.post("/api/session-pacing/"+strconv.Itoa(sessionID)+"/tick", nil)
	if resp.Code != 200 {
		t.Fatalf("tick pacing failed: %d", resp.Code)
	}

	// Complete session
	resp = tc.post("/api/session-pacing/"+strconv.Itoa(sessionID)+"/complete", nil)
	if resp.Code != 200 {
		t.Fatalf("complete pacing failed: %d", resp.Code)
	}

	resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
	readJSON(resp, &session)
	if session["status"] != "completed" {
		t.Fatalf("expected completed status, got %v", session["status"])
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Session pacing lifecycle test passed")
}

func TestSessionPacingSceneAdvance(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create adventure with acts and scenes
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Multi-Act Adventure",
		"template":          "five_room_dungeon",
		"difficulty":        "easy",
		"estimated_minutes": 180,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Start pacing session
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/pacing/start", nil)
	if resp.Code != 201 {
		t.Fatalf("start pacing failed: %d", resp.Code)
	}
	var startResult map[string]any
	readJSON(resp, &startResult)
	sessionID := int(startResult["id"].(float64))

	// Get initial state
	resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
	if resp.Code != 200 {
		t.Fatalf("get session failed: %d", resp.Code)
	}
	var session map[string]any
	readJSON(resp, &session)

	advanceCount := 0
	sceneTimings, _ := session["scene_timings"].([]any)
	_ = sceneTimings

	// Advance through scenes
	for i := range 8 {
		// Check session detail before advancing to see what we have
		resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
		readJSON(resp, &session)
		if session["status"] == "completed" {
			break
		}

		resp = tc.post("/api/session-pacing/"+strconv.Itoa(sessionID)+"/next-scene", nil)
		if resp.Code != 200 {
			t.Fatalf("advance scene failed at step %d: %d - %s", i, resp.Code, resp.Body.String())
		}
		advanceCount++
	}

	if advanceCount < 3 {
		t.Fatalf("expected at least 3 scene advances, got %d", advanceCount)
	}

	// Verify session eventually completed or has advanced
	resp = tc.get("/api/session-pacing/"+strconv.Itoa(sessionID), nil)
	readJSON(resp, &session)

	if st, ok := session["scene_timings"].([]any); ok {
		t.Logf("Session advanced %d times, final status: %v, scenes tracked: %d",
			advanceCount, session["status"], len(st))
	} else {
		t.Logf("Session advanced %d times, final status: %v, no scene timings",
			advanceCount, session["status"])
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Session pacing scene advance test passed")
}

func TestSessionPacingResumeExisting(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create adventure
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Resume Test",
		"template":          "five_room_dungeon",
		"difficulty":        "medium",
		"estimated_minutes": 90,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Start session
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/pacing/start", nil)
	if resp.Code != 201 {
		t.Fatalf("start pacing failed: %d", resp.Code)
	}
	var result map[string]any
	readJSON(resp, &result)
	sessionID := int(result["id"].(float64))

	// Pause session
	tc.post("/api/session-pacing/"+strconv.Itoa(sessionID)+"/pause", nil)

	// Start again - should resume existing
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/pacing/start", nil)
	if resp.Code != 200 {
		t.Fatalf("resume existing pacing failed: %d - expected 200", resp.Code)
	}
	readJSON(resp, &result)
	if int(result["id"].(float64)) != sessionID {
		t.Fatalf("expected same session id %d, got %v", sessionID, result["id"])
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Session pacing resume existing test passed")
}

// ─── Clue/Mystery Tracker Tests ───

func TestClueCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create adventure
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Mystery Adventure",
		"template":          "custom",
		"difficulty":        "medium",
		"estimated_minutes": 120,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Create a clue
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/clues", map[string]any{
		"title":       "Bloodstained Dagger",
		"description": "A dagger with dried blood found at the scene.",
		"clue_type":   "object",
		"sort_order":  1,
		"notes":       "The blood is elven.",
	})
	if resp.Code != 201 {
		t.Fatalf("create clue failed: %d - %s", resp.Code, resp.Body.String())
	}
	var clueResult map[string]any
	readJSON(resp, &clueResult)
	clueID := int(clueResult["id"].(float64))
	if clueID == 0 {
		t.Fatal("expected non-zero clue id")
	}

	// Get clue
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	if resp.Code != 200 {
		t.Fatalf("get clue failed: %d", resp.Code)
	}
	var clue map[string]any
	readJSON(resp, &clue)
	if clue["title"] != "Bloodstained Dagger" {
		t.Fatalf("expected 'Bloodstained Dagger', got %v", clue["title"])
	}

	// Update clue
	resp = tc.put("/api/clues/"+strconv.Itoa(clueID), map[string]any{
		"title":          "Bloodstained Dagger (Updated)",
		"description":    "Updated description",
		"clue_type":      "object",
		"is_red_herring": true,
		"sort_order":     2,
		"notes":          "Updated notes",
	})
	if resp.Code != 200 {
		t.Fatalf("update clue failed: %d", resp.Code)
	}

	// Verify update
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	readJSON(resp, &clue)
	if clue["title"] != "Bloodstained Dagger (Updated)" {
		t.Fatalf("expected updated title")
	}

	// List clues
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/clues", nil)
	if resp.Code != 200 {
		t.Fatalf("list clues failed: %d", resp.Code)
	}
	var clues []any
	readJSON(resp, &clues)
	if len(clues) < 1 {
		t.Fatal("expected at least 1 clue")
	}

	// Delete clue
	resp = tc.del("/api/clues/"+strconv.Itoa(clueID), nil)
	if resp.Code != 200 {
		t.Fatalf("delete clue failed: %d", resp.Code)
	}
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	if resp.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", resp.Code)
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Clue CRUD test passed")
}
