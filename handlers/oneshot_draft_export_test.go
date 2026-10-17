package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

// setupOneShotExportRouter builds a router with the export endpoint and a
// fixed user identity, so ownership checks can be exercised per user.
func setupOneShotExportRouter(t *testing.T, userID int64) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", "tester")
		c.Set("role", "user")
		c.Set("session_id", "test")
		c.Next()
	})
	g.GET("/oneshot-adventures/:id/draft", ExportOneShotDraft)
	return r
}

func TestExportOneShotDraft(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "other", "user")
	ctx := context.Background()

	now := "2026-10-02 12:00:00"
	res, err := db.DB.Exec(`INSERT INTO oneshot_adventures
		(user_id, title, premise, hook, template, estimated_minutes, difficulty, notes, created_at, updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`,
		1, "The Wolves of Welton", "Sheep are missing", "A farmer pleads", "custom", 180, "easy", "n", now, now)
	if err != nil {
		t.Fatalf("seed adventure: %v", err)
	}
	adventureID, _ := res.LastInsertId()

	actRes, err := db.DB.Exec("INSERT INTO oneshot_acts(adventure_id, number, title, description, estimated_minutes) VALUES(?,?,?,?,?)",
		adventureID, 1, "Arrival", "d", 30)
	if err != nil {
		t.Fatalf("seed act: %v", err)
	}
	actID, _ := actRes.LastInsertId()
	if _, err := db.DB.Exec("INSERT INTO oneshot_scenes(act_id, number, title, description, scene_type, estimated_minutes, notes) VALUES(?,?,?,?,?,?,'')",
		actID, 1, "The Farm", "s", "roleplay", 15); err != nil {
		t.Fatalf("seed scene: %v", err)
	}

	npc, err := db.Client.NPC.Create().SetUserID(1).SetName("Marla").SetRace("human").SetDescription("guide").Save(ctx)
	if err != nil {
		t.Fatalf("seed npc: %v", err)
	}
	db.DB.Exec("INSERT INTO oneshot_adventure_npcs(adventure_id, npc_id, role, story_hook, combat_ready) VALUES(?,?,?,'',0)",
		adventureID, npc.ID, "ally")

	loc, err := db.Client.Location.Create().SetUserID(1).SetName("Welton").SetType("village").SetDescription("sleepy").Save(ctx)
	if err != nil {
		t.Fatalf("seed location: %v", err)
	}
	db.DB.Exec("INSERT INTO oneshot_adventure_locations(adventure_id, location_id) VALUES(?,?)", adventureID, loc.ID)

	encRes, err := db.DB.Exec("INSERT INTO encounter_templates(campaign_id,user_id,name,description,environment,difficulty,xp_budget,total_xp,notes) VALUES(NULL,?,?,?,'','easy',0,0,'')",
		1, "Wolf Pack", "e")
	if err != nil {
		t.Fatalf("seed encounter: %v", err)
	}
	encID, _ := encRes.LastInsertId()
	db.DB.Exec("INSERT INTO oneshot_adventure_encounters(adventure_id, encounter_id) VALUES(?,?)", adventureID, encID)

	if _, err := db.DB.Exec("INSERT INTO clues(adventure_id, title, description, clue_type, is_red_herring, sort_order, notes) VALUES(?,?,?,?,0,0,'')",
		adventureID, "Tracks", "c", "object"); err != nil {
		t.Fatalf("seed clue: %v", err)
	}

	// Owner gets the full serialized draft.
	r := setupOneShotExportRouter(t, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/oneshot-adventures/%d/draft", adventureID), nil))
	if w.Code != http.StatusOK {
		t.Fatalf("export status = %d, body %s", w.Code, w.Body.String())
	}
	var d aiOneShotDraft
	if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
		t.Fatalf("unmarshal draft: %v", err)
	}
	if d.Title != "The Wolves of Welton" || d.Difficulty != "easy" || d.EstimatedMinutes != 180 {
		t.Fatalf("main fields wrong: %+v", d)
	}
	if len(d.Acts) != 1 || len(d.Acts[0].Scenes) != 1 || d.Acts[0].Scenes[0].Title != "The Farm" {
		t.Fatalf("acts/scenes wrong: %+v", d.Acts)
	}
	if len(d.NPCs) != 1 || d.NPCs[0].Name != "Marla" || d.NPCs[0].Role != "ally" {
		t.Fatalf("npcs wrong: %+v", d.NPCs)
	}
	if len(d.Locations) != 1 || d.Locations[0].Name != "Welton" {
		t.Fatalf("locations wrong: %+v", d.Locations)
	}
	if len(d.Encounters) != 1 || d.Encounters[0].Name != "Wolf Pack" {
		t.Fatalf("encounters wrong: %+v", d.Encounters)
	}
	if len(d.Clues) != 1 || d.Clues[0].Title != "Tracks" {
		t.Fatalf("clues wrong: %+v", d.Clues)
	}

	// A different user must not be able to read it.
	other := setupOneShotExportRouter(t, 2)
	w2 := httptest.NewRecorder()
	other.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/oneshot-adventures/%d/draft", adventureID), nil))
	if w2.Code != http.StatusNotFound {
		t.Fatalf("foreign export status = %d, want 404", w2.Code)
	}
}

func TestReviseAIDraftSectionPrompt(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)

	srv, seen := draftProvider(t, `{"status":"ready","message":"done","draft":{"title":"Scarier"}}`)
	defer srv.Close()
	epID := seedAIEndpoint(t, "revise", "text", srv.URL, "m", true)

	w := testutil.PostJSON(t, r, "/api/ai/revise", map[string]any{
		"entity_type": "oneshot",
		"json":        `{"title":"X"}`,
		"instruction": "make it scarier",
		"section":     "Act 1: Arrival",
		"endpoint_id": epID,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("revise status = %d, body %s", w.Code, w.Body.String())
	}
	if len(*seen) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(*seen))
	}
	msgs := (*seen)[0]
	if len(msgs) != 2 {
		t.Fatalf("messages = %d, want 2", len(msgs))
	}
	system, user := msgs[0]["content"], msgs[1]["content"]
	if !strings.Contains(system, "Act 1: Arrival") || !strings.Contains(system, "selected one part") {
		t.Fatalf("system prompt missing section scope: %q", system)
	}
	if !strings.Contains(user, "Only revise this part: Act 1: Arrival") || !strings.Contains(user, "make it scarier") {
		t.Fatalf("user prompt missing section scope: %q", user)
	}
}
