package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

// 5.1-5.6: unified monster picker — tab rendering per context
func TestHtmxMonsterPickerRendersTabs(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/monster-picker/:context/:contextId", HtmxMonsterPicker)
	})

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rec, req)
		return rec
	}

	// oneshot context: compendium + library tabs, no roster tab
	rec := get("/api/htmx/monster-picker/oneshot/1")
	if rec.Code != http.StatusOK {
		t.Fatalf("oneshot: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "mpCompendiumTab") || !strings.Contains(body, "mpLibraryTab") {
		t.Fatalf("oneshot: missing compendium/library tabs: %s", body[:min(len(body), 300)])
	}
	if strings.Contains(body, "mpRosterTab") {
		t.Fatalf("oneshot: roster tab should not render: %s", body[:min(len(body), 300)])
	}

	// campaign context: roster tab present
	rec = get("/api/htmx/monster-picker/campaign/5")
	if rec.Code != http.StatusOK {
		t.Fatalf("campaign: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "mpRosterTab") {
		t.Fatalf("campaign: roster tab missing: %s", rec.Body.String()[:min(len(rec.Body.String()), 300)])
	}

	// encounter context: Create Custom wiring to encounter monster form
	rec = get("/api/htmx/monster-picker/encounter/3")
	if rec.Code != http.StatusOK {
		t.Fatalf("encounter: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body = rec.Body.String()
	if !strings.Contains(body, "Create Custom") || !strings.Contains(body, "/htmx/encounters/3/monsters/new") {
		t.Fatalf("encounter: Create Custom wiring missing: %s", body[:min(len(body), 400)])
	}

	// invalid context → 400
	rec = get("/api/htmx/monster-picker/bogus/1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bogus context: expected 400, got %d", rec.Code)
	}
}

// 5.7: add from compendium tab into a one-shot
func TestMonsterPickerOneShotAdd(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	// seed a legacy compendium monster row
	if _, err := db.DB.Exec("INSERT INTO compendium_monsters (id, name, type) VALUES (9001, 'Test Orc', 'humanoid')"); err != nil {
		t.Fatalf("seed compendium_monsters: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/oneshot-adventures", CreateOneShotAdventure)
		auth.POST("/oneshot-adventures/:id/import/compendium-entry", ImportCompendiumMonsterToOneShot)
	})

	// create one-shot
	osBody, _ := json.Marshal(map[string]any{"name": "Picker OS", "dm_notes": ""})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/oneshot-adventures", bytes.NewReader(osBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("create oneshot: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}
	var os struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &os); err != nil || os.ID == 0 {
		t.Fatalf("create oneshot: no id: %v body=%s", err, rec.Body.String())
	}

	// import compendium monster
	impBody, _ := json.Marshal(map[string]any{"compendium_monster_id": 9001, "adventure_id": os.ID})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/oneshot-adventures/"+itoa64(os.ID)+"/import/compendium-entry", bytes.NewReader(impBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("import: expected 2xx, got %d: %s", rec.Code, rec.Body.String())
	}

	// verify oneshot_monsters row
	var cnt int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_monsters WHERE adventure_id = ? AND compendium_monster_id = 9001", os.ID).Scan(&cnt); err != nil {
		t.Fatalf("query oneshot_monsters: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 oneshot_monsters row, got %d", cnt)
	}
}

// 5.8: campaign roster tab lists roster entries
func TestMonsterPickerCampaignRoster(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 5, "Picker Camp", "Party", 1)

	// seed a compendium monster + roster row
	if _, err := db.DB.Exec("INSERT INTO compendium_monsters (id, name, type) VALUES (9001, 'Test Orc', 'humanoid')"); err != nil {
		t.Fatalf("seed compendium_monsters: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_monster_roster (campaign_id, compendium_monster_id, name) VALUES (5, 9001, 'Test Orc')"); err != nil {
		t.Fatalf("seed roster: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/monster-picker/:context/:contextId", HtmxMonsterPicker)
		auth.GET("/htmx/campaigns/:id/monster-roster", HtmxCampaignMonsterRoster)
	})

	// picker campaign context shows roster tab
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/htmx/monster-picker/campaign/5", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("picker: expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "mpRosterTab") {
		t.Fatalf("picker: roster tab missing")
	}

	// roster pane content (what hx-get /htmx/campaigns/5/monster-roster loads) lists the entry
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/htmx/campaigns/5/monster-roster", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("roster: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Test Orc") {
		t.Fatalf("roster: entry missing from roster list: %s", rec.Body.String()[:min(len(rec.Body.String()), 400)])
	}
}
