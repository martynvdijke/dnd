package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"villum/handlers/testutil"
)

// Route-level CRUD coverage for endpoints added across the recent feature
// phases that were not yet exercised through their HTTP handlers.

func wledAdminRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/wled-settings", GetWLEDSettings)
		auth.POST("/wled-settings", SaveWLEDSettings)
		auth.POST("/test-wled", TestWLED)
	})
}

func TestWLEDTestEndpoint(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := wledAdminRouter()

	// No devices configured yet.
	w := testutil.PostJSON(t, r, "/api/test-wled", map[string]any{})
	testutil.AssertStatus(t, w, 200)
	var res struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	testutil.ParseJSON(t, w, &res)
	if res.Success {
		t.Fatalf("test-wled with no devices = %+v, want success=false", res)
	}

	// Save a device pointing at an unreachable address; the test endpoint must
	// still answer 200 with success=false (never 500 / panic).
	testutil.AssertStatus(t, testutil.PostJSON(t, r, "/api/wled-settings", map[string]any{
		"enabled": true,
		"devices": []map[string]any{
			{"name": "TV", "base_url": "127.0.0.1:1", "brightness": 120, "enabled": true},
		},
	}), 200)

	var got struct {
		Enabled    bool         `json:"enabled"`
		Configured bool         `json:"configured"`
		Devices    []wledDevice `json:"devices"`
	}
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/wled-settings"), &got)
	if !got.Configured || len(got.Devices) != 1 {
		t.Fatalf("settings = %+v, want one configured device", got)
	}
	if got.Devices[0].BaseURL != "http://127.0.0.1:1" {
		t.Fatalf("base_url = %q, want normalized http://127.0.0.1:1", got.Devices[0].BaseURL)
	}

	w = testutil.PostJSON(t, r, "/api/test-wled", map[string]any{})
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &res)
	if res.Success {
		t.Fatalf("test-wled with unreachable device = %+v, want success=false", res)
	}
}

func TestHtmxActEncounterCrud(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedOneShot(t, 10, 1, "Adventure")
	testutil.SeedOneShotAct(t, 100, 10, "Act One", 1)
	seedEncounterTemplate(t, 500, 0, 1, "Goblin Ambush")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/oneshot-acts/:id/encounters", HtmxActEncounters)
		auth.POST("/htmx/oneshot-acts/:id/encounters", HtmxLinkActEncounter)
		auth.DELETE("/htmx/oneshot-acts/:id/encounters/:eid", HtmxUnlinkActEncounter)
	})

	// Read: the picker panel lists the candidate encounter.
	w := testutil.Get(t, r, "/api/htmx/oneshot-acts/100/encounters")
	testutil.AssertStatus(t, w, 200)
	if body := w.Body.String(); !strings.Contains(body, `name="encounter_id"`) || !strings.Contains(body, "Goblin Ambush") {
		t.Fatalf("encounter panel missing picker/candidate:\n%s", body)
	}
	if n := testutil.CountRows(t, "oneshot_adventure_encounters"); n != 0 {
		t.Fatalf("links before create = %d, want 0", n)
	}

	// Create: link via the form, link row persists.
	w = testutil.PostForm(t, r, "/api/htmx/oneshot-acts/100/encounters", map[string]string{"encounter_id": "500"})
	testutil.AssertStatus(t, w, 200)
	if n := testutil.CountRows(t, "oneshot_adventure_encounters"); n != 1 {
		t.Fatalf("links after create = %d, want 1", n)
	}

	// Delete: unlink by encounter id.
	w = testutil.Delete(t, r, "/api/htmx/oneshot-acts/100/encounters/500")
	testutil.AssertStatus(t, w, 200)
	if n := testutil.CountRows(t, "oneshot_adventure_encounters"); n != 0 {
		t.Fatalf("links after delete = %d, want 0", n)
	}
}

func TestBattlemapCreateTokenDirect(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedCampaign(t, 1, "MapCamp", "The Party", 1)

	r := battlemapRouter()

	// Create a free-standing token (not linked to a combat entry).
	w := testutil.PostJSON(t, r, "/api/campaigns/1/battlemap/tokens", map[string]any{
		"name": "Rock", "x": 0.25, "y": 0.75, "color": "#112233",
	})
	testutil.AssertStatus(t, w, 201)
	var created struct {
		Token battlemapToken `json:"token"`
	}
	testutil.ParseJSON(t, w, &created)
	if created.Token.ID == 0 || created.Token.Name != "Rock" {
		t.Fatalf("created token = %+v", created.Token)
	}
	if created.Token.X != 0.25 || created.Token.Y != 0.75 {
		t.Fatalf("token position = %v,%v want 0.25,0.75", created.Token.X, created.Token.Y)
	}
	if n := testutil.CountRows(t, "battlemap_tokens"); n != 1 {
		t.Fatalf("token rows = %d, want 1", n)
	}

	// Reading a non-existent campaign is a 404.
	testutil.AssertStatus(t, testutil.Get(t, r, "/api/campaigns/0/battlemap"), 404)
	testutil.AssertStatus(t, testutil.PostJSON(t, r, "/api/campaigns/0/battlemap/tokens", map[string]any{"name": "X"}), 404)
}
