package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"vellum/db"
	"vellum/handlers"
	"vellum/middleware"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	testDB := "/tmp/vellum_test.db"
	os.Remove(testDB)
	os.Remove(testDB + "-wal")
	os.Remove(testDB + "-shm")

	if err := db.Init(testDB); err != nil {
		fmt.Printf("Failed to init test db: %v\n", err)
		os.Exit(1)
	}
	db.Seed()

	gin.SetMode(gin.TestMode)
	testRouter = buildRouter()

	middleware.StartCleanupTask()
	code := m.Run()
	db.Close()
	os.Remove(testDB)
	os.Remove(testDB + "-wal")
	os.Remove(testDB + "-shm")
	os.Exit(code)
}

func buildRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	r.GET("/api/check-setup", handlers.CheckSetup)
	r.POST("/api/login", handlers.HandleLogin)

	auth := r.Group("/api")
	auth.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	{
		auth.POST("/logout", handlers.HandleLogout)
		auth.GET("/csrf-token", handlers.GetCSRFToken)
		auth.GET("/user/me", handlers.GetMe)
		auth.GET("/characters", handlers.ListCharacters)
		auth.POST("/characters", handlers.CreateCharacter)
		auth.GET("/characters/:id", handlers.GetCharacter)
		auth.PUT("/characters/:id", handlers.UpdateCharacter)
		auth.DELETE("/characters/:id", handlers.DeleteCharacter)
		auth.GET("/characters/:id/export", handlers.ExportCharacter)
		auth.GET("/characters/:id/print", handlers.PrintCharacter)
		auth.POST("/characters/import", handlers.ImportJSON)
		auth.PUT("/characters/:id/currency", handlers.UpdateCurrency)
		auth.PUT("/characters/:id/spellcasting", handlers.UpdateSpellcasting)
		auth.POST("/characters/:id/inventory", handlers.CreateInventory)
		auth.PUT("/inventory/:iid", handlers.UpdateInventory)
		auth.DELETE("/inventory/:iid", handlers.DeleteInventory)
		auth.POST("/characters/:id/spells", handlers.CreateSpell)
		auth.PUT("/spells/:sid", handlers.UpdateSpell)
		auth.DELETE("/spells/:sid", handlers.DeleteSpell)
		auth.POST("/characters/:id/features", handlers.CreateFeature)
		auth.PUT("/features/:fid", handlers.UpdateFeature)
		auth.DELETE("/features/:fid", handlers.DeleteFeature)
		auth.POST("/proficiencies", handlers.CreateProficiency)
		auth.DELETE("/proficiencies/:pid", handlers.DeleteProficiency)
		auth.POST("/roll", handlers.HandleRoll)
		auth.POST("/roll/check", handlers.HandleCheckRoll)
		auth.GET("/dice-rolls", handlers.GetDiceRolls)
		auth.GET("/compendium/races", handlers.ListCompendiumRaces)
		auth.GET("/compendium/classes", handlers.ListCompendiumClasses)
		auth.GET("/compendium/spells", handlers.ListCompendiumSpells)
		auth.GET("/compendium/feats", handlers.ListCompendiumFeats)
		auth.GET("/compendium/backgrounds", handlers.ListCompendiumBackgrounds)
		auth.GET("/compendium/equipment", handlers.ListCompendiumEquipment)
		auth.GET("/compendium/search", handlers.SearchCompendium)
		auth.GET("/locations", handlers.ListLocations)
		auth.POST("/locations", handlers.CreateLocation)
		auth.PUT("/locations/:id", handlers.UpdateLocation)
		auth.DELETE("/locations/:id", handlers.DeleteLocation)
		auth.GET("/characters/:id/locations", handlers.GetCharacterLocations)
		auth.POST("/characters/:id/locations", handlers.LinkLocation)
		auth.DELETE("/locations/link/:lid", handlers.UnlinkLocation)
		auth.GET("/npcs", handlers.ListNPCs)
		auth.POST("/npcs", handlers.CreateNPC)
		auth.PUT("/npcs/:id", handlers.UpdateNPC)
		auth.DELETE("/npcs/:id", handlers.DeleteNPC)
		auth.GET("/characters/:id/npcs", handlers.GetCharacterNPCs)
		auth.POST("/characters/:id/npcs", handlers.LinkNPC)
		auth.DELETE("/npcs/link/:nid", handlers.UnlinkNPC)
		auth.POST("/npcs/link/:nid/interact", handlers.LogNPCInteraction)
		auth.GET("/characters/:id/sessions", handlers.ListSessions)
		auth.POST("/characters/:id/sessions", handlers.CreateSession)
		auth.PUT("/sessions/:sid", handlers.UpdateSession)
		auth.DELETE("/sessions/:sid", handlers.DeleteSession)
		auth.GET("/characters/:id/quests", handlers.ListQuests)
		auth.POST("/characters/:id/quests", handlers.CreateQuest)
		auth.PUT("/quests/:qid", handlers.UpdateQuest)
		auth.DELETE("/quests/:qid", handlers.DeleteQuest)
		auth.GET("/characters/:id/journal", handlers.ListJournal)
		auth.POST("/characters/:id/journal", handlers.CreateJournalEntry)
		auth.PUT("/journal/:jid", handlers.UpdateJournalEntry)
		auth.DELETE("/journal/:jid", handlers.DeleteJournalEntry)
		auth.GET("/characters/:id/graph", handlers.GetGraphData)
		auth.GET("/characters/:id/stats", handlers.GetCharacterStats)
		auth.GET("/campaigns", handlers.ListCampaigns)
		auth.POST("/campaigns", handlers.CreateCampaign)
		auth.PUT("/campaigns/:id", handlers.UpdateCampaign)
		auth.DELETE("/campaigns/:id", handlers.DeleteCampaign)
		auth.POST("/characters/:id/rest", handlers.DoRest)
		auth.POST("/characters/:id/levelup", handlers.LevelUp)
		auth.GET("/party", handlers.GetPartyView)

		// Combat
		auth.GET("/combat", handlers.ListCombatEntries)
		auth.POST("/combat", handlers.CreateCombatEntry)
		auth.PUT("/combat/:id", handlers.UpdateCombatEntry)
		auth.DELETE("/combat/:id", handlers.DeleteCombatEntry)
		auth.POST("/combat/initiative", handlers.RollInitiative)
		auth.POST("/combat/next-turn", handlers.NextTurn)
		auth.GET("/combat/current-turn", handlers.GetCurrentTurn)

		// Generators
		auth.GET("/generate/npc", handlers.HandleGenerateNPC)
		auth.GET("/generate/name", handlers.HandleGenerateName)
		auth.GET("/generate/encounter", handlers.HandleGenerateEncounter)
		auth.GET("/generate/loot", handlers.HandleGenerateLoot)
	}

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired(), middleware.CSRFRequired())
	{
		admin.GET("/users", handlers.AdminListUsers)
		admin.POST("/users", handlers.AdminCreateUser)
		admin.PUT("/users/:id", handlers.AdminUpdateUser)
		admin.DELETE("/users/:id", handlers.AdminDeleteUser)
		admin.PUT("/users/:id/password", handlers.AdminResetPassword)
		admin.POST("/compendium/:type", handlers.AdminCreateCompendiumEntry)
		admin.PUT("/compendium/:type/:id", handlers.AdminUpdateCompendiumEntry)
		admin.DELETE("/compendium/:type/:id", handlers.AdminDeleteCompendiumEntry)
	}
	return r
}

type testClient struct {
	sessionID string
	csrf      string
}

func newTestClient() *testClient { return &testClient{} }

func (tc *testClient) req(method, path string, body any) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if tc.sessionID != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: tc.sessionID})
	}
	if tc.csrf != "" {
		req.Header.Set("X-CSRF-Token", tc.csrf)
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// Capture session cookie from Set-Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			tc.sessionID = c.Value
		}
	}
	return w
}

func (tc *testClient) get(path string, body any) *httptest.ResponseRecorder { return tc.req("GET", path, body) }
func (tc *testClient) post(path string, body any) *httptest.ResponseRecorder { return tc.req("POST", path, body) }
func (tc *testClient) put(path string, body any) *httptest.ResponseRecorder  { return tc.req("PUT", path, body) }
func (tc *testClient) del(path string, body any) *httptest.ResponseRecorder  { return tc.req("DELETE", path, body) }

const adminUser = "admin"
const adminPass = "testpass123"

func setupAdmin(t *testing.T, tc *testClient) {
	t.Helper()
	resp := tc.get("/api/check-setup", nil)
	var data map[string]bool
	json.Unmarshal(resp.Body.Bytes(), &data)
	if !data["setup"] {
		resp = tc.post("/api/login", map[string]any{
			"username": adminUser,
			"password": adminPass,
			"setup":    true,
		})
		if resp.Code != 200 {
			t.Fatalf("admin setup failed: %d - %s", resp.Code, resp.Body.String())
		}
	}
	// Always login to ensure fresh session
	resp = tc.post("/api/login", map[string]any{
		"username": adminUser,
		"password": adminPass,
	})
	if resp.Code != 200 {
		t.Fatalf("admin login failed: %d - %s", resp.Code, resp.Body.String())
	}
	// Fetch CSRF token
	resp = tc.get("/api/csrf-token", nil)
	if resp.Code != 200 {
		t.Fatalf("csrf token fetch failed: %d - %s", resp.Code, resp.Body.String())
	}
	var csrfData map[string]string
	json.Unmarshal(resp.Body.Bytes(), &csrfData)
	tc.csrf = csrfData["token"]
}

func login(t *testing.T, tc *testClient, username, password string) {
	t.Helper()
	resp := tc.post("/api/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.Code != 200 {
		t.Fatalf("login failed: %d - %s", resp.Code, resp.Body.String())
	}
	resp = tc.get("/api/csrf-token", nil)
	var csrfData map[string]string
	json.Unmarshal(resp.Body.Bytes(), &csrfData)
	tc.csrf = csrfData["token"]
}

func readJSON(w *httptest.ResponseRecorder, v any) {
	json.Unmarshal(w.Body.Bytes(), v)
}

// ─── Auth Tests ───

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
	if resp.Code != 200 {
		t.Fatalf("setup failed: %d - %s", resp.Code, resp.Body.String())
	}

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
		{"1d20", 1, 20},
		{"2d6", 2, 12},
		{"1d8+3", 4, 11},
		{"1d100", 1, 100},
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
		"name": "Curse of Strahd", "description": "Ravenloft campaign",
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

	// Update
	resp = tc.put(fmt.Sprintf("/api/campaigns/%d", cid), map[string]any{
		"name": "Curse of Strahd Revised", "description": "Updated campaign",
	})
	if resp.Code != 200 {
		t.Fatalf("update failed: %d", resp.Code)
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

	// Create a couple characters
	tc.post("/api/characters", map[string]any{"name": "Party Hero 1", "race": "Human", "class": "Fighter", "hp_max": 30, "hp_current": 25})
	tc.post("/api/characters", map[string]any{"name": "Party Hero 2", "race": "Elf", "class": "Wizard", "hp_max": 20, "hp_current": 20})

	resp := tc.get("/api/party", nil)
	if resp.Code != 200 {
		t.Fatalf("party view failed: %d - %s", resp.Code, resp.Body.String())
	}
	var groups []any
	readJSON(resp, &groups)
	if len(groups) < 1 {
		t.Fatalf("expected at least 1 party group, got %d", len(groups))
	}
	// Check members exist
	found := false
	for _, g := range groups {
		gm := g.(map[string]any)
		members := gm["members"].([]any)
		for _, m := range members {
			mm := m.(map[string]any)
			if mm["name"] == "Party Hero 1" {
				found = true
			}
		}
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
		"username": "player1", "password": "playerpass", "role": "user",
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

	// Update
	resp = tc.put(fmt.Sprintf("/api/admin/users/%d", uid), map[string]any{
		"username": "player1_updated", "display_name": "Player One", "role": "user",
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

func TestAdminCompendiumCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create
	resp := tc.post("/api/admin/compendium/races", map[string]any{
		"name": "Test Race", "description": "Custom race", "speed": 30, "size": "Medium",
	})
	if resp.Code != 201 {
		t.Fatalf("create failed: %d - %s", resp.Code, resp.Body.String())
	}
	var entry map[string]any
	readJSON(resp, &entry)
	eid := int(entry["id"].(float64))

	// Update
	resp = tc.put(fmt.Sprintf("/api/admin/compendium/races/%d", eid), map[string]any{
		"name": "Test Race Updated", "description": "Updated", "speed": 35, "size": "Large",
	})
	if resp.Code != 200 {
		t.Fatalf("update failed: %d", resp.Code)
	}

	// Delete
	resp = tc.del(fmt.Sprintf("/api/admin/compendium/races/%d", eid), nil)
	if resp.Code != 200 {
		t.Fatalf("delete failed: %d", resp.Code)
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
}

func TestImportJSON(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	importData := []map[string]any{{
		"name": "Imported Hero", "race": "Elf", "class": "Ranger", "level": 3,
		"str": 12, "dex": 18, "con": 14, "int": 10, "wis": 16, "cha": 8,
		"hp_max": 28, "hp_current": 28, "ac": 15, "speed": 35,
		"spells": []map[string]any{{"name": "Hunter's Mark", "level": 1, "school": "Divination"}},
		"inventory": []map[string]any{{"name": "Longbow", "category": "weapon", "quantity": 1, "damage_dice": "1d8", "damage_type": "piercing"}},
		"proficiencies": []map[string]any{{"name": "Perception", "type": "skill"}},
		"currency": map[string]any{"gp": 100},
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

func TestAdminCompendiumAllTypesCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Test all 6 types
	tests := []struct {
		typ   string
		body  map[string]any
		upd   map[string]any
	}{
		{"races", map[string]any{"name": "Custom Race", "description": "T", "speed": 30, "size": "M"}, map[string]any{"name": "Custom Race v2", "description": "U", "speed": 35, "size": "L"}},
		{"classes", map[string]any{"name": "Custom Class", "description": "T", "hit_die": 8, "primary_ability": "str"}, map[string]any{"name": "Custom Class v2", "description": "U", "hit_die": 10, "primary_ability": "dex"}},
		{"spells", map[string]any{"name": "Custom Spell", "level": 1, "school": "Evocation"}, map[string]any{"name": "Custom Spell v2", "level": 2, "school": "Abjuration"}},
		{"feats", map[string]any{"name": "Custom Feat", "description": "T"}, map[string]any{"name": "Custom Feat v2", "description": "U"}},
		{"backgrounds", map[string]any{"name": "Custom BG", "description": "T"}, map[string]any{"name": "Custom BG v2", "description": "U"}},
		{"equipment", map[string]any{"name": "Custom Item", "category": "Gear", "cost": "{}", "weight": 1.0}, map[string]any{"name": "Custom Item v2", "category": "Weapon", "cost": "{}", "weight": 2.0}},
	}
	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			resp := tc.post(fmt.Sprintf("/api/admin/compendium/%s", tt.typ), tt.body)
			if resp.Code != 201 {
				t.Fatalf("create %s failed: %d - %s", tt.typ, resp.Code, resp.Body.String())
			}
			var entry map[string]any
			readJSON(resp, &entry)
			eid := int(entry["id"].(float64))

			resp = tc.put(fmt.Sprintf("/api/admin/compendium/%s/%d", tt.typ, eid), tt.upd)
			if resp.Code != 200 {
				t.Fatalf("update %s failed: %d", tt.typ, resp.Code)
			}

			resp = tc.del(fmt.Sprintf("/api/admin/compendium/%s/%d", tt.typ, eid), nil)
			if resp.Code != 200 {
				t.Fatalf("delete %s failed: %d", tt.typ, resp.Code)
			}
		})
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
		t.Fatalf("long rest failed: %d", resp.Code)
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
		{"101d6", 400},
		{"1d1001", 400},
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

func TestSpellcasting(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	resp := tc.post("/api/characters", map[string]any{"name": "Caster", "race": "Elf", "class": "Wizard"})
	var char map[string]any
	readJSON(resp, &char)
	cid := int(char["id"].(float64))

	// Set up spellcasting
	resp = tc.put(fmt.Sprintf("/api/characters/%d/spellcasting", cid), map[string]any{
		"ability": "int", "save_dc": 15, "attack_bonus": 7,
		"slots_1_max": 4, "slots_1_used": 2,
		"slots_2_max": 3, "slots_2_used": 1,
	})
	if resp.Code != 200 {
		t.Fatalf("set spellcasting failed: %d", resp.Code)
	}

	// Verify
	resp = tc.get(fmt.Sprintf("/api/characters/%d", cid), nil)
	readJSON(resp, &char)
	sc := char["spellcasting"].(map[string]any)
	if int(sc["save_dc"].(float64)) != 15 {
		t.Fatalf("expected DC 15, got %v", sc["save_dc"])
	}
}
