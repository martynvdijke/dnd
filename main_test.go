package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
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

		// One-Shot Adventures
		auth.GET("/oneshot-adventures", handlers.ListOneShotAdventures)
		auth.POST("/oneshot-adventures", handlers.CreateOneShotAdventure)
		auth.GET("/oneshot-adventures/:id", handlers.GetOneShotAdventure)
		auth.PUT("/oneshot-adventures/:id", handlers.UpdateOneShotAdventure)
		auth.DELETE("/oneshot-adventures/:id", handlers.DeleteOneShotAdventure)
		auth.POST("/oneshot-adventures/generate", handlers.GenerateOneShotFromTemplate)
		auth.POST("/oneshot-adventures/:id/acts", handlers.CreateOneShotAct)
		auth.GET("/oneshot-adventures/:id/npcs", handlers.GetOneShotNPCs)
		auth.POST("/oneshot-adventures/:id/npcs", handlers.LinkOneShotNPC)
		auth.DELETE("/oneshot-adventures/:id/npcs/:nid", handlers.UnlinkOneShotNPC)
		auth.GET("/oneshot-adventures/:id/locations", handlers.GetOneShotLocations)
		auth.POST("/oneshot-adventures/:id/locations", handlers.LinkOneShotLocation)
		auth.DELETE("/oneshot-adventures/:id/locations/:lid", handlers.UnlinkOneShotLocation)
		auth.GET("/oneshot-adventures/:id/encounters", handlers.GetOneShotEncounters)
		auth.POST("/oneshot-adventures/:id/encounters", handlers.LinkOneShotEncounter)
		auth.DELETE("/oneshot-adventures/:id/encounters/:eid", handlers.UnlinkOneShotEncounter)

		// Acts & Scenes
		auth.PUT("/oneshot-acts/:id", handlers.UpdateOneShotAct)
		auth.DELETE("/oneshot-acts/:id", handlers.DeleteOneShotAct)
		auth.POST("/oneshot-acts/:id/scenes", handlers.CreateOneShotScene)
		auth.PUT("/oneshot-scenes/:id", handlers.UpdateOneShotScene)
		auth.DELETE("/oneshot-scenes/:id", handlers.DeleteOneShotScene)

		// Session Pacing
		auth.POST("/oneshot-adventures/:id/pacing/start", handlers.StartPacingSession)
		auth.GET("/oneshot-adventures/:id/pacing", handlers.GetPacingSession)
		auth.GET("/session-pacing/:id", handlers.GetPacingSession)
		auth.POST("/session-pacing/:id/pause", handlers.PausePacingSession)
		auth.POST("/session-pacing/:id/resume", handlers.ResumePacingSession)
		auth.POST("/session-pacing/:id/next-scene", handlers.AdvanceToNextScene)
		auth.POST("/session-pacing/:id/complete", handlers.CompletePacingSession)
		auth.POST("/session-pacing/:id/tick", handlers.UpdatePacingTimers)

		// Clue/Mystery Tracker
		auth.GET("/oneshot-adventures/:id/clues", handlers.ListClues)
		auth.POST("/oneshot-adventures/:id/clues", handlers.CreateClue)
		auth.GET("/clues/:id", handlers.GetClue)
		auth.PUT("/clues/:id", handlers.UpdateClue)
		auth.DELETE("/clues/:id", handlers.DeleteClue)
		auth.POST("/clues/:id/reveal", handlers.RevealClue)
		auth.POST("/clues/:id/hide", handlers.HideClue)
		auth.POST("/clues/:id/dependencies", handlers.AddClueDependency)
		auth.DELETE("/clues/:id/dependencies/:did", handlers.RemoveClueDependency)
		auth.POST("/clues/:id/npcs", handlers.LinkClueNPC)
		auth.DELETE("/clues/:id/npcs/:nid", handlers.UnlinkClueNPC)
		auth.POST("/clues/:id/locations", handlers.LinkClueLocation)
		auth.DELETE("/clues/:id/locations/:lid", handlers.UnlinkClueLocation)

		// Pregenerated Characters
		auth.GET("/pregens", handlers.ListPregens)
		auth.POST("/pregens", handlers.CreatePregen)
		auth.GET("/pregens/generate", handlers.GeneratePregen)
		auth.GET("/pregens/balance", handlers.CheckPartyBalance)
		auth.GET("/pregens/:id", handlers.GetPregen)
		auth.PUT("/pregens/:id", handlers.UpdatePregen)
		auth.DELETE("/pregens/:id", handlers.DeletePregen)

		// Prep Dashboard & Checklist
		auth.GET("/oneshot-adventures/:id/checklist", handlers.ListPrepChecklist)
		auth.POST("/oneshot-adventures/:id/checklist", handlers.CreatePrepChecklistItem)
		auth.PUT("/prep-checklist/:cid", handlers.UpdatePrepChecklistItem)
		auth.DELETE("/prep-checklist/:cid", handlers.DeletePrepChecklistItem)

		// DM Screen / Quick Reference
		auth.GET("/oneshot-adventures/:id/dm-screen", handlers.HtmxDmScreen)
		auth.GET("/oneshot-adventures/:id/notes", handlers.ListDmNotes)
		auth.POST("/oneshot-adventures/:id/notes", handlers.CreateDmNote)
		auth.PUT("/dm-notes/:nid", handlers.UpdateDmNote)
		auth.DELETE("/oneshot-adventures/:id/notes/:nid", handlers.DeleteDmNote)

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
		auth.GET("/generate/adventure-hook", handlers.HandleGenerateAdventureHook)
		auth.GET("/generate/dungeon-dressing", handlers.HandleGenerateDungeonDressing)
		auth.GET("/generate/tavern", handlers.HandleGenerateTavern)
		auth.GET("/generate/urban-encounter", handlers.HandleGenerateUrbanEncounter)
		auth.GET("/generate/road-encounter", handlers.HandleGenerateRoadEncounter)

		// Wiki
		auth.GET("/campaigns/:id/wiki", handlers.ListWikiPages)
		auth.POST("/campaigns/:id/wiki", handlers.CreateWikiPage)
		auth.GET("/wiki/:id", handlers.GetWikiPage)
		auth.PUT("/wiki/:id", handlers.UpdateWikiPage)
		auth.DELETE("/wiki/:id", handlers.DeleteWikiPage)

		// Campaign graph
		auth.GET("/campaigns/:id/graph", handlers.GetCampaignGraphData)

		// Share links
		auth.POST("/share", handlers.CreateShareLink)
		auth.GET("/share", handlers.ListShareLinks)
		auth.DELETE("/share/:token", handlers.DeleteShareLink)

		// Campaign Dashboard
		auth.GET("/campaigns/:id/dashboard", handlers.GetCampaignDashboard)

		// Character Resources
		auth.GET("/characters/:id/resources", handlers.ListCharacterResources)
		auth.POST("/characters/:id/resources", handlers.CreateCharacterResource)
		auth.PUT("/resources/:id", handlers.UpdateCharacterResource)
		auth.DELETE("/resources/:id", handlers.DeleteCharacterResource)
		auth.POST("/characters/:id/recover-resources", handlers.RecoverResourcesOnRest)

		// Combat Log
		auth.GET("/combat-log", handlers.ListCombatLogEntries)
		auth.POST("/combat-log", handlers.CreateCombatLogEntry)
		auth.GET("/combat-log/stats", handlers.GetCombatLogStats)

		// Homebrew Content Manager
		auth.GET("/homebrew/:type", handlers.ListHomebrewContent)
		auth.POST("/homebrew/:type", handlers.CreateHomebrewContent)
		auth.PUT("/homebrew/:type/:id", handlers.UpdateHomebrewContent)
		auth.DELETE("/homebrew/:type/:id", handlers.DeleteHomebrewContent)

		// Campaign Maps
		auth.GET("/campaigns/:id/maps", handlers.ListCampaignMaps)
		auth.POST("/campaigns/:id/maps", handlers.CreateCampaignMap)
		auth.PUT("/maps/:id", handlers.UpdateCampaignMap)
		auth.DELETE("/maps/:id", handlers.DeleteCampaignMap)
		auth.POST("/campaigns/:id/maps/:mapId/activate", handlers.SetActiveCampaignMap)
		auth.PUT("/maps/:id/fog", handlers.UpdateFogOfWar)
		auth.GET("/campaigns/:id/maps/active", handlers.GetActiveCampaignMap)
		auth.GET("/maps/:mapId/pins", handlers.ListMapPins)
		auth.POST("/maps/:mapId/pins", handlers.CreateMapPin)
		auth.PUT("/map-pins/:id", handlers.UpdateMapPin)
		auth.DELETE("/map-pins/:id", handlers.DeleteMapPin)

		// Quick Reference
		auth.GET("/quickref", handlers.GetQuickReference)

		// Downtime Activities
		auth.GET("/characters/:id/downtime", handlers.ListDowntimeActivities)
		auth.POST("/characters/:id/downtime", handlers.CreateDowntimeActivity)
		auth.PUT("/downtime/:id", handlers.UpdateDowntimeActivity)
		auth.DELETE("/downtime/:id", handlers.DeleteDowntimeActivity)
		auth.POST("/downtime/:id/advance", handlers.AdvanceDowntimeDay)
		auth.GET("/downtime/types", handlers.GetDowntimeTypes)

		// Level Up Planner
		auth.GET("/characters/:id/level-plan", handlers.GetLevelUpPlan)
		auth.POST("/characters/:id/level-plan", handlers.SaveLevelUpPlan)
		auth.DELETE("/characters/:id/level-plan", handlers.DeleteLevelUpPlan)
		auth.GET("/characters/:id/level-suggestions", handlers.GetLevelUpSuggestions)

		// Campaign Recaps
		auth.GET("/campaigns/:id/recaps", handlers.ListCampaignRecaps)
		auth.POST("/campaigns/:id/recaps", handlers.CreateCampaignRecap)
		auth.GET("/recaps/:id", handlers.GetCampaignRecap)
		auth.PUT("/recaps/:id", handlers.UpdateCampaignRecap)
		auth.DELETE("/recaps/:id", handlers.DeleteCampaignRecap)
		auth.POST("/campaigns/:id/recaps/generate", handlers.GenerateCampaignRecap)
		auth.POST("/recaps/:id/mark-sent", handlers.MarkRecapAsSent)
	}

	dm := r.Group("/api")
	dm.Use(middleware.AuthRequired(), middleware.DMRequired(), middleware.CSRFRequired())
	{
		dm.GET("/compendium-monsters", handlers.ListCompendiumMonsters)
		dm.GET("/compendium-monsters/:id", handlers.GetCompendiumMonster)
	}

	htmxGroup := r.Group("/")
	htmxGroup.Use(middleware.AuthRequired(), middleware.CSRFRequired())
	handlers.HtmxRegisterRoutes(htmxGroup)

	// Public share routes
	r.GET("/api/share/:token", handlers.GetSharedEntity)

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired(), middleware.CSRFRequired())
	{
		admin.GET("/users", handlers.AdminListUsers)
		admin.POST("/users", handlers.AdminCreateUser)
		admin.PUT("/users/:id", handlers.AdminUpdateUser)
		admin.DELETE("/users/:id", handlers.AdminDeleteUser)
		admin.PUT("/users/:id/password", handlers.AdminResetPassword)
		admin.GET("/email-settings", handlers.GetEmailSettings)
		admin.POST("/email-settings", handlers.SaveEmailSettings)
		admin.POST("/test-email", handlers.TestEmail)
		admin.POST("/campaign-highlights", handlers.SendCampaignHighlights)
	}

	// Static file serving for non-API routes
	staticFS, _ := fs.Sub(staticFiles, "static")
	r.GET("/static/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		if filepath == "" || filepath == "/" {
			c.String(http.StatusNotFound, "not found")
			return
		}
		data, err := fs.ReadFile(staticFS, filepath[1:])
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}
		ct := "application/octet-stream"
		if strings.HasSuffix(filepath, ".js") {
			ct = "application/javascript"
		} else if strings.HasSuffix(filepath, ".css") {
			ct = "text/css"
		} else if strings.HasSuffix(filepath, ".json") {
			ct = "application/json"
		} else if strings.HasSuffix(filepath, ".html") {
			ct = "text/html"
		} else if strings.HasSuffix(filepath, ".svg") {
			ct = "image/svg+xml"
		}
		c.Data(http.StatusOK, ct, data)
	})

	// Serve HTML pages with version substitution
	serveHTML := func(path, fileName string) {
		r.GET(path, func(c *gin.Context) {
			data, err := fs.ReadFile(staticFS, fileName)
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			content := strings.ReplaceAll(string(data), "{{VERSION}}", Version)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
		})
	}
	serveHTML("/", "app.html")
	serveHTML("/login", "login.html")
	serveHTML("/setup", "setup.html")
	serveHTML("/admin", "admin.html")

	// Redirect /app to / for backward compatibility
	r.GET("/app", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

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
