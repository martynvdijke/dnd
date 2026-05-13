package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	testDB := "/tmp/villum_test.db"
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

	r.GET("/healthz", handlers.HandleHealth)
	r.GET("/metrics", handlers.HandleMetrics)
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
		auth.GET("/compendium/api/:category", handlers.FetchFromDnDApi)
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
		auth.GET("/campaigns/:id/members", handlers.ListCampaignMembers)
		auth.POST("/campaigns/:id/members", handlers.AddCampaignMember)
		auth.PUT("/campaigns/:id/members/:userId", handlers.SetCampaignMemberRole)
		auth.DELETE("/campaigns/:id/members/:userId", handlers.RemoveCampaignMember)
		auth.GET("/users/search", handlers.SearchUsers)
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

		// Multi-class
		auth.POST("/characters/:id/classes", handlers.CreateCharacterClass)
		auth.PUT("/classes/:ccid", handlers.UpdateCharacterClass)
		auth.DELETE("/classes/:ccid", handlers.DeleteCharacterClass)

		// Encounter Builder
		auth.GET("/encounters", handlers.ListEncounters)
		auth.POST("/encounters", handlers.CreateEncounter)
		auth.GET("/encounters/:id", handlers.GetEncounter)
		auth.PUT("/encounters/:id", handlers.UpdateEncounter)
		auth.DELETE("/encounters/:id", handlers.DeleteEncounter)
		auth.POST("/encounters/:id/monsters", handlers.AddEncounterMonster)
		auth.PUT("/encounter-monsters/:mid", handlers.UpdateEncounterMonster)
		auth.DELETE("/encounter-monsters/:mid", handlers.DeleteEncounterMonster)
		auth.POST("/encounters/calculate-xp", handlers.CalculateEncounterXP)
		auth.GET("/monster-xp", handlers.GetMonsterXP)

		// Calendar
		auth.GET("/calendar", handlers.ListCalendarEvents)
		auth.POST("/calendar", handlers.CreateCalendarEvent)
		auth.PUT("/calendar/:id", handlers.UpdateCalendarEvent)
		auth.DELETE("/calendar/:id", handlers.DeleteCalendarEvent)

		// Timeline
		auth.GET("/timeline", handlers.ListTimelineEvents)
		auth.POST("/timeline", handlers.CreateTimelineEvent)
		auth.PUT("/timeline/:id", handlers.UpdateTimelineEvent)
		auth.DELETE("/timeline/:id", handlers.DeleteTimelineEvent)

		// Conditions
		auth.GET("/conditions", handlers.ListConditions)
		auth.POST("/conditions", handlers.CreateCondition)
		auth.PUT("/conditions/:id", handlers.UpdateCondition)
		auth.DELETE("/conditions/:id", handlers.DeleteCondition)
		auth.POST("/conditions/tick", handlers.TickConditions)
		auth.GET("/conditions/types", handlers.GetConditionTypes)
		auth.GET("/conditions/summary", handlers.GetActiveConditionSummary)
		auth.POST("/characters/:id/check-concentration", handlers.CheckConcentration)

		// Feats
		auth.GET("/feats", handlers.ListFeats)
		auth.POST("/feats", handlers.CreateFeat)
		auth.PUT("/feats/:id", handlers.UpdateFeat)
		auth.DELETE("/feats/:id", handlers.DeleteFeat)

		// Companions
		auth.GET("/companions", handlers.ListCompanions)
		auth.POST("/companions", handlers.CreateCompanion)
		auth.PUT("/companions/:id", handlers.UpdateCompanion)
		auth.DELETE("/companions/:id", handlers.DeleteCompanion)

		// Factions
		auth.GET("/factions", handlers.ListFactions)
		auth.POST("/factions", handlers.CreateFaction)
		auth.PUT("/factions/:id", handlers.UpdateFaction)
		auth.DELETE("/factions/:id", handlers.DeleteFaction)
		auth.GET("/faction-reputation", handlers.GetFactionReputations)
		auth.POST("/faction-reputation", handlers.SetFactionReputation)
		auth.DELETE("/faction-reputation/:id", handlers.DeleteFactionReputation)
		auth.GET("/generate/weather", handlers.HandleGenerateWeather)

		// Notes
		auth.GET("/notes", handlers.ListCharacterNotes)
		auth.POST("/notes", handlers.CreateCharacterNote)
		auth.PUT("/notes/:id", handlers.UpdateCharacterNote)
		auth.DELETE("/notes/:id", handlers.DeleteCharacterNote)

		// HP
		auth.POST("/characters/:id/calc-hp", handlers.CalculateHP)

		// Comparison
		auth.GET("/characters/compare", handlers.CompareCharacters)

		// Generators
		auth.GET("/generate/npc", handlers.HandleGenerateNPC)
		auth.GET("/generate/name", handlers.HandleGenerateName)
		auth.GET("/generate/encounter", handlers.HandleGenerateEncounter)
		auth.GET("/generate/loot", handlers.HandleGenerateLoot)
		auth.GET("/generate/character", handlers.HandleGenerateRandomCharacter)
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

func (tc *testClient) get(path string, body any) *httptest.ResponseRecorder {
	return tc.req("GET", path, body)
}
func (tc *testClient) post(path string, body any) *httptest.ResponseRecorder {
	return tc.req("POST", path, body)
}
func (tc *testClient) put(path string, body any) *httptest.ResponseRecorder {
	return tc.req("PUT", path, body)
}
func (tc *testClient) del(path string, body any) *httptest.ResponseRecorder {
	return tc.req("DELETE", path, body)
}

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

func TestAdminCompendiumAllTypesCRUD(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Test all 6 types
	tests := []struct {
		typ  string
		body map[string]any
		upd  map[string]any
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
	if resp.Code != 403 {
		t.Fatalf("expected 403 for demoted user, got %d - %s", resp.Code, resp.Body.String())
	}
	t.Logf("Demoted user correctly denied (got %d)", resp.Code)

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
	r1, r2 := int(rolls[0].(float64)), int(rolls[1].(float64))
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
	r1, r2 = int(rolls[0].(float64)), int(rolls[1].(float64))
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
	if total != int(rolls[0].(float64)) {
		t.Fatalf("normal total %d != roll %d", total, int(rolls[0].(float64)))
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
	chosen := int(breakdown[0].(map[string]any)["total"].(float64))
	total = int(result["total"].(float64))
	if total != chosen+5 {
		t.Fatalf("adv+5 total %d != chosen %d + 5 = %d", total, chosen, chosen+5)
	}
	t.Logf("Advantage+5: rolls=[%d,%d] chosen=%d total=%d",
		int(rolls[0].(float64)), int(rolls[1].(float64)), chosen, total)

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

func TestCompendiumSystemFilter(t *testing.T) {
	tc := newTestClient()
	setupAdmin(t, tc)

	// Create a custom compendium entry with a different system
	resp := tc.post("/api/admin/compendium/races", map[string]any{
		"name": "Test Race PF2e", "description": "A PF2e test race",
		"speed": 25, "size": "Medium", "system": "pf2e", "source": "custom",
	})
	if resp.Code != 201 {
		t.Fatalf("create pf2e race failed: %d", resp.Code)
	}

	// Verify it appears in the listing
	resp = tc.get("/api/compendium/races", nil)
	var races []map[string]any
	readJSON(resp, &races)
	found := false
	for _, r := range races {
		if r["system"] == "pf2e" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pf2e race in listing, but none found")
	}
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
	for i := 0; i < 4; i++ {
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


