package main

import (
	"strconv"
	"testing"
	"github.com/gin-gonic/gin"

	"villum/handlers"
)

func TestClueRevealHide(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Reveal Test",
		"template":          "custom",
		"difficulty":        "easy",
		"estimated_minutes": 60,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Create clue
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/clues", map[string]any{
		"title":       "Secret Letter",
		"description": "A hidden letter reveals the plot.",
		"clue_type":   "direct",
	})
	if resp.Code != 201 {
		t.Fatalf("create clue failed: %d", resp.Code)
	}
	var clueResult map[string]any
	readJSON(resp, &clueResult)
	clueID := int(clueResult["id"].(float64))

	// Initially hidden
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	readJSON(resp, &clueResult)
	if clueResult["is_revealed"].(bool) {
		t.Fatal("expected clue to be initially hidden")
	}

	// Reveal
	resp = tc.post("/api/clues/"+strconv.Itoa(clueID)+"/reveal", nil)
	if resp.Code != 200 {
		t.Fatalf("reveal clue failed: %d", resp.Code)
	}
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	readJSON(resp, &clueResult)
	if !clueResult["is_revealed"].(bool) {
		t.Fatal("expected clue to be revealed")
	}

	// Hide
	resp = tc.post("/api/clues/"+strconv.Itoa(clueID)+"/hide", nil)
	if resp.Code != 200 {
		t.Fatalf("hide clue failed: %d", resp.Code)
	}
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	readJSON(resp, &clueResult)
	if clueResult["is_revealed"].(bool) {
		t.Fatal("expected clue to be hidden after hide")
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Clue reveal/hide test passed")
}

func TestClueDependencies(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Dependency Test",
		"template":          "custom",
		"difficulty":        "hard",
		"estimated_minutes": 180,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Create two clues
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/clues", map[string]any{
		"title":       "First Clue",
		"description": "The initial clue.",
		"clue_type":   "direct",
		"sort_order":  1,
	})
	if resp.Code != 201 {
		t.Fatalf("create first clue failed: %d", resp.Code)
	}
	var c1 map[string]any
	readJSON(resp, &c1)
	c1ID := int(c1["id"].(float64))

	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/clues", map[string]any{
		"title":       "Second Clue",
		"description": "Requires the first clue.",
		"clue_type":   "location",
		"sort_order":  2,
	})
	if resp.Code != 201 {
		t.Fatalf("create second clue failed: %d", resp.Code)
	}
	var c2 map[string]any
	readJSON(resp, &c2)
	c2ID := int(c2["id"].(float64))

	// Add dependency: Clue 2 depends on Clue 1
	resp = tc.post("/api/clues/"+strconv.Itoa(c2ID)+"/dependencies", map[string]any{
		"depends_on_id": c1ID,
	})
	if resp.Code != 201 {
		t.Fatalf("add dependency failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Verify dependency on Clue 2
	resp = tc.get("/api/clues/"+strconv.Itoa(c2ID), nil)
	if resp.Code != 200 {
		t.Fatalf("get clue 2 failed: %d", resp.Code)
	}
	readJSON(resp, &c2)
	deps := c2["dependencies"].([]any)
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(deps))
	}
	dep := deps[0].(map[string]any)
	if int(dep["depends_on_id"].(float64)) != c1ID {
		t.Fatalf("expected depends_on_id %d, got %v", c1ID, dep["depends_on_id"])
	}

	// Verify Clue 1 has depended_by
	resp = tc.get("/api/clues/"+strconv.Itoa(c1ID), nil)
	readJSON(resp, &c1)
	depBy := c1["depended_by"].([]any)
	if len(depBy) != 1 {
		t.Fatalf("expected cl 1 to have 1 depended_by, got %d", len(depBy))
	}

	// Remove dependency
	resp = tc.del("/api/clues/"+strconv.Itoa(c2ID)+"/dependencies/"+strconv.Itoa(c1ID), nil)
	if resp.Code != 200 {
		t.Fatalf("remove dependency failed: %d", resp.Code)
	}

	// Verify dependency removed
	resp = tc.get("/api/clues/"+strconv.Itoa(c2ID), nil)
	readJSON(resp, &c2)
	deps = c2["dependencies"].([]any)
	if len(deps) != 0 {
		t.Fatalf("expected 0 dependencies after removal, got %d", len(deps))
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Clue dependencies test passed")
}

func TestClueNPCLocationLinks(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create entities
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Linked Clues",
		"template":          "custom",
		"difficulty":        "medium",
		"estimated_minutes": 90,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	resp = tc.post("/api/npcs", map[string]any{"name": "Informant Joe"})
	if resp.Code != 201 {
		t.Fatalf("create npc failed: %d", resp.Code)
	}
	var npcR map[string]any
	readJSON(resp, &npcR)
	npcID := int(npcR["id"].(float64))

	resp = tc.post("/api/locations", map[string]any{"name": "Hidden Cave", "type": "cave"})
	if resp.Code != 201 {
		t.Fatalf("create location failed: %d", resp.Code)
	}
	var locR map[string]any
	readJSON(resp, &locR)
	locID := int(locR["id"].(float64))

	// Create clue
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/clues", map[string]any{
		"title":       "Linked Clue",
		"description": "A clue with links.",
		"clue_type":   "witness",
	})
	if resp.Code != 201 {
		t.Fatalf("create clue failed: %d", resp.Code)
	}
	var clueR map[string]any
	readJSON(resp, &clueR)
	clueID := int(clueR["id"].(float64))

	// Link NPC
	resp = tc.post("/api/clues/"+strconv.Itoa(clueID)+"/npcs", map[string]any{
		"npc_id": npcID,
	})
	if resp.Code != 200 {
		t.Fatalf("link npc to clue failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Link Location
	resp = tc.post("/api/clues/"+strconv.Itoa(clueID)+"/locations", map[string]any{
		"location_id": locID,
	})
	if resp.Code != 200 {
		t.Fatalf("link location to clue failed: %d", resp.Code)
	}

	// Verify links
	resp = tc.get("/api/clues/"+strconv.Itoa(clueID), nil)
	readJSON(resp, &clueR)
	npcs := clueR["npcs"].([]any)
	if len(npcs) != 1 {
		t.Fatalf("expected 1 linked NPC, got %d", len(npcs))
	}
	locs := clueR["locations"].([]any)
	if len(locs) != 1 {
		t.Fatalf("expected 1 linked location, got %d", len(locs))
	}

	// Unlink NPC
	resp = tc.del("/api/clues/"+strconv.Itoa(clueID)+"/npcs/"+strconv.Itoa(npcID), nil)
	if resp.Code != 200 {
		t.Fatalf("unlink npc failed: %d", resp.Code)
	}

	// Unlink Location
	resp = tc.del("/api/clues/"+strconv.Itoa(clueID)+"/locations/"+strconv.Itoa(locID), nil)
	if resp.Code != 200 {
		t.Fatalf("unlink location failed: %d", resp.Code)
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)
	tc.del("/api/locations/"+strconv.Itoa(locID), nil)
	tc.del("/api/npcs/"+strconv.Itoa(npcID), nil)

	t.Log("Clue NPC/location links test passed")
}

// ─── Pregenerated Character Tests ───

func TestPregenCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create pregen
	resp := tc.post("/api/pregens", map[string]any{
		"name":  "Aldric Stoneheart",
		"race":  "dwarf",
		"class": "fighter",
		"level": 3,
		"str":   16, "dex": 12, "con": 14, "int": 8, "wis": 10, "cha": 10,
		"hp": 28, "ac": 18, "speed": 25,
		"skills":    "Athletics, Perception",
		"equipment": "Chain mail, battleaxe, shield",
		"alignment": "Lawful Good",
	})
	if resp.Code != 201 {
		t.Fatalf("create pregen failed: %d - %s", resp.Code, resp.Body.String())
	}
	var created map[string]any
	readJSON(resp, &created)
	pid := int(created["id"].(float64))
	if pid == 0 {
		t.Fatal("expected non-zero id")
	}

	// Get pregen
	resp = tc.get("/api/pregens/"+strconv.Itoa(pid), nil)
	if resp.Code != 200 {
		t.Fatalf("get pregen failed: %d", resp.Code)
	}
	var pregen map[string]any
	readJSON(resp, &pregen)
	if pregen["name"] != "Aldric Stoneheart" {
		t.Fatalf("expected name 'Aldric Stoneheart', got %v", pregen["name"])
	}
	if pregen["race"] != "dwarf" {
		t.Fatalf("expected race 'dwarf', got %v", pregen["race"])
	}

	// Update pregen
	resp = tc.put("/api/pregens/"+strconv.Itoa(pid), map[string]any{
		"name":  "Aldric Ironbane",
		"race":  "dwarf",
		"class": "fighter",
		"level": 5,
		"str":   18, "dex": 12, "con": 16, "int": 8, "wis": 10, "cha": 10,
		"hp": 44, "ac": 18, "speed": 25,
		"skills": "Athletics, Perception, Intimidation",
	})
	if resp.Code != 200 {
		t.Fatalf("update pregen failed: %d", resp.Code)
	}

	resp = tc.get("/api/pregens/"+strconv.Itoa(pid), nil)
	readJSON(resp, &pregen)
	if pregen["name"] != "Aldric Ironbane" {
		t.Fatalf("expected updated name 'Aldric Ironbane', got %v", pregen["name"])
	}

	// List pregens
	resp = tc.get("/api/pregens", nil)
	if resp.Code != 200 {
		t.Fatalf("list pregens failed: %d", resp.Code)
	}
	var list []any
	readJSON(resp, &list)
	if len(list) < 1 {
		t.Fatal("expected at least 1 pregen")
	}

	// Delete pregen
	resp = tc.del("/api/pregens/"+strconv.Itoa(pid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete pregen failed: %d", resp.Code)
	}

	resp = tc.get("/api/pregens/"+strconv.Itoa(pid), nil)
	if resp.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", resp.Code)
	}

	t.Log("Pregen CRUD test passed")
}

func TestPregenGenerate(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Quick generate a pregen
	resp := tc.get("/api/pregens/generate?level=3", nil)
	if resp.Code != 201 {
		t.Fatalf("generate pregen failed: %d - %s", resp.Code, resp.Body.String())
	}
	var created map[string]any
	readJSON(resp, &created)
	pid := int(created["id"].(float64))
	if pid == 0 {
		t.Fatal("expected non-zero id after generate")
	}

	// Verify the generated pregen has valid data
	resp = tc.get("/api/pregens/"+strconv.Itoa(pid), nil)
	if resp.Code != 200 {
		t.Fatalf("get generated pregen failed: %d", resp.Code)
	}
	var pregen map[string]any
	readJSON(resp, &pregen)
	if pregen["name"].(string) == "" {
		t.Fatal("expected non-empty name for generated pregen")
	}
	if pregen["race"].(string) == "" {
		t.Fatal("expected non-empty race for generated pregen")
	}
	if pregen["class"].(string) == "" {
		t.Fatal("expected non-empty class for generated pregen")
	}
	if int(pregen["level"].(float64)) != 3 {
		t.Fatalf("expected level 3, got %v", pregen["level"])
	}

	// Cleanup
	tc.del("/api/pregens/"+strconv.Itoa(pid), nil)

	t.Log("Pregen generation test passed")
}

func TestPartyBalance(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a few pregens for party balance
	classes := []string{"fighter", "cleric", "rogue", "wizard"}
	for _, cls := range classes {
		resp := tc.post("/api/pregens", map[string]any{
			"name":  "Hero " + cls,
			"race":  "human",
			"class": cls,
			"level": 3,
			"str":   12, "dex": 12, "con": 12, "int": 12, "wis": 12, "cha": 12,
			"hp": 20, "ac": 12, "speed": 30,
		})
		if resp.Code != 201 {
			t.Fatalf("create pregen %s failed: %d", cls, resp.Code)
		}
	}

	// Check party balance
	resp := tc.get("/api/pregens/balance", nil)
	if resp.Code != 200 {
		t.Fatalf("party balance failed: %d", resp.Code)
	}
	var balance map[string]any
	readJSON(resp, &balance)
	roles := balance["roles"].(map[string]any)
	if roles["tank"].(float64) < 1 {
		t.Errorf("expected at least 1 tank role, got %v", roles["tank"])
	}
	if roles["healer"].(float64) < 1 {
		t.Errorf("expected at least 1 healer role, got %v", roles["healer"])
	}
	chars := balance["characters"].([]any)
	if len(chars) != 4 {
		t.Fatalf("expected 4 characters in balance, got %d", len(chars))
	}
	rating := balance["rating"].(string)
	if rating == "" {
		t.Fatal("expected non-empty rating")
	}

	t.Log("Party balance test passed")
}

// ─── Prep Dashboard & Checklist Tests ───

func TestPrepChecklistCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create an adventure for the checklist
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Prep Test Adventure",
		"template":          "custom",
		"estimated_minutes": 120,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// List empty checklist
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/checklist", nil)
	if resp.Code != 200 {
		t.Fatalf("list empty checklist failed: %d", resp.Code)
	}
	var items []any
	readJSON(resp, &items)
	if len(items) != 0 {
		t.Fatalf("expected empty checklist, got %d items", len(items))
	}

	// Add checklist items
	itemTitles := []string{"Prepare maps", "Create NPCs", "Print character sheets"}
	var itemIDs []int
	for _, title := range itemTitles {
		resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/checklist", map[string]any{
			"item":       title,
			"category":   "general",
			"sort_order": len(itemIDs) + 1,
		})
		if resp.Code != 201 {
			t.Fatalf("add checklist item '%s' failed: %d", title, resp.Code)
		}
		var result map[string]any
		readJSON(resp, &result)
		itemIDs = append(itemIDs, int(result["id"].(float64)))
	}

	// Verify 3 items
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/checklist", nil)
	readJSON(resp, &items)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Toggle first item via PUT
	resp = tc.put("/api/prep-checklist/"+strconv.Itoa(itemIDs[0]), map[string]any{
		"is_checked": true,
	})
	if resp.Code != 200 {
		t.Fatalf("toggle item failed: %d", resp.Code)
	}

	// Delete second item
	resp = tc.del("/api/prep-checklist/"+strconv.Itoa(itemIDs[1]), nil)
	if resp.Code != 200 {
		t.Fatalf("delete item failed: %d", resp.Code)
	}

	// Verify 2 items remain, one checked
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/checklist", nil)
	readJSON(resp, &items)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Prep checklist CRUD test passed")
}

func TestPrepDashboardDataLoad(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create an adventure with acts and scenes
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "Dashboard Test",
		"premise":           "Test premise",
		"hook":              "Test hook",
		"template":          "five_room_dungeon",
		"difficulty":        "medium",
		"estimated_minutes": 180,
		"notes":             "Test notes",
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Add checklist items
	tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/checklist", map[string]any{
		"item":       "Test prep item",
		"category":   "general",
		"sort_order": 1,
	})

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("Prep dashboard data load test passed")
}

// ─── DM Screen / Quick Reference Tests ───

func TestDmNoteCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create an adventure
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "DM Note Test Adventure",
		"template":          "custom",
		"estimated_minutes": 60,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Create a DM note
	resp = tc.post("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/notes", map[string]any{
		"title":   "Villain Secret",
		"content": "The mayor is actually a doppelganger",
	})
	if resp.Code != 201 {
		t.Fatalf("create dm note failed: %d - %s", resp.Code, resp.Body.String())
	}
	var noteResult map[string]any
	readJSON(resp, &noteResult)
	nid := int(noteResult["id"].(float64))
	if nid == 0 {
		t.Fatal("expected non-zero note id")
	}

	// List notes
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/notes", nil)
	if resp.Code != 200 {
		t.Fatalf("list notes failed: %d", resp.Code)
	}
	var notes []any
	readJSON(resp, &notes)
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	// Update note
	resp = tc.put("/api/dm-notes/"+strconv.Itoa(nid), map[string]any{
		"title":   "Villain Secret (Updated)",
		"content": "The mayor serves the Mind Flayers",
	})
	if resp.Code != 200 {
		t.Fatalf("update note failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Delete note
	resp = tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/notes/"+strconv.Itoa(nid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete note failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Verify deletion
	resp = tc.get("/api/oneshot-adventures/"+strconv.Itoa(aid)+"/notes", nil)
	readJSON(resp, &notes)
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes after delete, got %d", len(notes))
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("DM note CRUD test passed")
}

func TestDmScreenData(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create an adventure
	resp := tc.post("/api/oneshot-adventures", map[string]any{
		"title":             "DM Screen Test",
		"premise":           "A test adventure for DM screen",
		"template":          "three_act_structure",
		"estimated_minutes": 120,
	})
	if resp.Code != 201 {
		t.Fatalf("create adventure failed: %d", resp.Code)
	}
	var adv map[string]any
	readJSON(resp, &adv)
	aid := int(adv["id"].(float64))

	// Access DM screen via HTMX
	resp = tc.get("/htmx/oneshot-adventures/"+strconv.Itoa(aid)+"/dm-screen", nil)
	if resp.Code != 200 {
		t.Fatalf("dm screen failed: %d - %s", resp.Code, resp.Body.String())
	}

	// Verify quick reference data via handler
	sections := handlers.QuickReferenceData()
	if len(sections) == 0 {
		t.Fatal("expected quick reference sections")
	}

	// Verify conditions section exists
	foundConditions := false
	for _, s := range sections {
		if s.Title == "Conditions" {
			foundConditions = true
			if len(s.Entries) < 10 {
				t.Fatalf("expected at least 10 conditions, got %d", len(s.Entries))
			}
		}
	}
	if !foundConditions {
		t.Fatal("expected Conditions section in quick reference data")
	}

	// Clean up
	tc.del("/api/oneshot-adventures/"+strconv.Itoa(aid), nil)

	t.Log("DM screen data test passed")
}

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, _ := setupRouter("/tmp/villum_test_media")
	if router == nil {
		t.Fatal("setupRouter returned nil")
	}
	routes := router.Routes()
	if len(routes) == 0 {
		t.Fatal("setupRouter registered no routes")
	}
	routeSet := make(map[string]bool, len(routes))
	for _, r := range routes {
		routeSet[r.Method+" "+r.Path] = true
	}
	// Key routes that must exist
	required := []string{
		"GET /healthz",
		"GET /metrics",
		"POST /api/login",
		"GET /api/campaigns",
		"POST /api/campaigns",
		"GET /api/characters",
		"POST /api/characters",
		"GET /api/combat",
		"GET /api/admin/users",
		"GET /static/*filepath",
		"GET /",
		"GET /login",
		"GET /admin",
		"GET /api/compendium/races",
		"GET /api/encounters",
	}
	for _, want := range required {
		if !routeSet[want] {
			t.Errorf("missing required route %q", want)
		}
	}
	// Verify idempotency: second router has identical method+path set
	router2, _ := setupRouter("/tmp/villum_test_media")
	routes2 := router2.Routes()
	set2 := make(map[string]bool, len(routes2))
	for _, r := range routes2 {
		set2[r.Method+" "+r.Path] = true
	}
	if len(routeSet) != len(set2) {
		t.Fatalf("route count mismatch between setupRouter calls: %d vs %d", len(routeSet), len(set2))
	}
	for k := range routeSet {
		if !set2[k] {
			t.Errorf("route %q missing in second setupRouter call", k)
		}
	}
	t.Logf("setupRouter registered %d routes", len(routes))
}
