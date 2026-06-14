package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

func TestCampaignPartyItemsCRUD(t *testing.T) {
	testDB := "/tmp/villum_party_items_test.db"
	os.Remove(testDB)

	if err := db.Init(testDB); err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDB)
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	auth.GET("/campaigns/:id/party-items", ListCampaignPartyItems)
	auth.POST("/campaigns/:id/party-items", CreateCampaignPartyItem)
	auth.DELETE("/party-items/:id", DeleteCampaignPartyItem)

	db.DB.Exec("INSERT INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")
	db.DB.Exec("INSERT INTO campaigns(id, name, user_id) VALUES(1, 'Test Campaign', 1)")

	t.Run("List returns empty initially", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/party-items", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var items []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("expected empty list, got %d items", len(items))
		}
	})

	t.Run("Create party item returns 201", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"name":     "Potion of Healing",
			"quantity": 5,
			"notes":    "Found in dungeon",
		})
		req, _ := http.NewRequest("POST", "/api/campaigns/1/party-items", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		id, ok := resp["id"].(float64)
		if !ok || id <= 0 {
			t.Errorf("expected valid id, got %v", resp["id"])
		}
	})

	t.Run("List returns created item", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/party-items", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var items []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0]["name"] != "Potion of Healing" {
			t.Errorf("expected 'Potion of Healing', got '%v'", items[0]["name"])
		}
		if int(items[0]["quantity"].(float64)) != 5 {
			t.Errorf("expected quantity 5, got %v", items[0]["quantity"])
		}
	})

	t.Run("Delete party item", func(t *testing.T) {
		// Get the item ID first
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/party-items", nil)
		r.ServeHTTP(w, req)
		var items []map[string]any
		json.Unmarshal(w.Body.Bytes(), &items)
		if len(items) == 0 {
			t.Fatal("no items to delete")
		}
		itemID := int(items[0]["id"].(float64))

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/party-items/%d", itemID), nil)
		r.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
		}
	})

	t.Run("List returns empty after delete", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/party-items", nil)
		r.ServeHTTP(w, req)

		var items []map[string]any
		json.Unmarshal(w.Body.Bytes(), &items)
		if len(items) != 0 {
			t.Errorf("expected empty list after delete, got %d items", len(items))
		}
	})
}

func TestCampaignSessionPlansCRUD(t *testing.T) {
	testDB := "/tmp/villum_session_plans_test.db"
	os.Remove(testDB)

	if err := db.Init(testDB); err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDB)
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	auth.GET("/campaigns/:id/session-plans", ListSessionPlans)
	auth.POST("/campaigns/:id/session-plans", CreateSessionPlan)
	auth.PUT("/session-plans/:id", UpdateSessionPlan)
	auth.DELETE("/session-plans/:id", DeleteSessionPlan)

	db.DB.Exec("INSERT INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")
	db.DB.Exec("INSERT INTO campaigns(id, name, user_id) VALUES(1, 'Test Campaign', 1)")

	var createdID float64

	t.Run("List returns empty initially", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/session-plans", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var plans []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &plans); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if len(plans) != 0 {
			t.Errorf("expected empty list, got %d items", len(plans))
		}
	})

	t.Run("Create session plan", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"title":        "Session 1",
			"session_date": "2025-06-01",
			"status":       "planned",
			"dm_notes":     "Prepare dragon encounter",
		})
		req, _ := http.NewRequest("POST", "/api/campaigns/1/session-plans", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		id, ok := resp["id"].(float64)
		if !ok || id <= 0 {
			t.Errorf("expected valid id, got %v", resp["id"])
		}
		createdID = id
	})

	t.Run("List returns created plan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/session-plans", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var plans []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &plans); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if len(plans) != 1 {
			t.Fatalf("expected 1 plan, got %d", len(plans))
		}
		if plans[0]["title"] != "Session 1" {
			t.Errorf("expected 'Session 1', got '%v'", plans[0]["title"])
		}
	})

	t.Run("Update session plan", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("no plan ID from create step")
		}
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"title":        "Session 1 - Updated",
			"session_date": "2025-06-01",
			"status":       "completed",
		})
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/session-plans/%d", int(createdID)), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify update
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/campaigns/1/session-plans", nil)
		r.ServeHTTP(w2, req2)
		var plans []map[string]any
		json.Unmarshal(w2.Body.Bytes(), &plans)
		if len(plans) > 0 && plans[0]["title"] != "Session 1 - Updated" {
			t.Errorf("expected updated title, got '%v'", plans[0]["title"])
		}
	})

	t.Run("Delete session plan", func(t *testing.T) {
		if createdID == 0 {
			t.Fatal("no plan ID from create step")
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/session-plans/%d", int(createdID)), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/campaigns/1/session-plans", nil)
		r.ServeHTTP(w2, req2)
		var plans []map[string]any
		json.Unmarshal(w2.Body.Bytes(), &plans)
		if len(plans) != 0 {
			t.Errorf("expected empty list after delete, got %d", len(plans))
		}
	})
}

func TestCampaignDashboard(t *testing.T) {
	testDB := "/tmp/villum_dashboard_test.db"
	os.Remove(testDB)

	if err := db.Init(testDB); err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDB)
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	auth.GET("/campaigns/:id/dashboard", GetCampaignDashboard)

	// Seed data
	db.DB.Exec("INSERT INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")
	db.DB.Exec("INSERT INTO campaigns(id, name, user_id, party_name) VALUES(1, 'Test Campaign', 1, 'The Heroes')")
	db.DB.Exec("INSERT INTO characters(id, user_id, name, race, class, level, hp_max, hp_current, campaign_id, str, dex, con, int, wis, cha, ac, speed) VALUES(1, 1, 'Hero', 'Human', 'Fighter', 5, 50, 50, 1, 15, 14, 13, 12, 10, 8, 17, 30)")

	t.Run("Dashboard returns campaign data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/1/dashboard", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var dash map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &dash); err != nil {
			t.Fatalf("failed to parse dashboard: %v", err)
		}

		if dash["name"] != "Test Campaign" {
			t.Errorf("expected campaign name, got '%v'", dash["name"])
		}
		if dash["party_name"] != "The Heroes" {
			t.Errorf("expected party name, got '%v'", dash["party_name"])
		}

		chars, ok := dash["characters"].([]any)
		if !ok {
			t.Fatal("expected characters array")
		}
		if len(chars) != 1 {
			t.Fatalf("expected 1 character, got %d", len(chars))
		}
		char := chars[0].(map[string]any)
		if char["name"] != "Hero" {
			t.Errorf("expected character name 'Hero', got '%v'", char["name"])
		}
		if int(char["level"].(float64)) != 5 {
			t.Errorf("expected level 5, got %v", char["level"])
		}
	})
}

func TestExhaustionAPI(t *testing.T) {
	testDB := "/tmp/villum_exhaustion_test.db"
	os.Remove(testDB)

	if err := db.Init(testDB); err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDB)
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	auth.PATCH("/characters/:id/exhaustion", UpdateExhaustion)

	// Need Ent schema tables created by migration, plus seed data
	db.DB.Exec("INSERT INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")
	db.DB.Exec("INSERT INTO campaigns(id, user_id, name) VALUES(1, 1, 'Test Campaign')")
	db.DB.Exec("INSERT INTO characters(id, user_id, name, race, class, level, hp_max, hp_current, campaign_id, str, dex, con, int, wis, cha, ac, speed) VALUES(1, 1, 'Hero', 'Human', 'Fighter', 5, 50, 50, 1, 15, 14, 13, 12, 10, 8, 17, 30)")

	t.Run("Set exhaustion to 3", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"level": 3,
		})
		req, _ := http.NewRequest("PATCH", "/api/characters/1/exhaustion", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Invalid level 7 returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"level": 7,
		})
		req, _ := http.NewRequest("PATCH", "/api/characters/1/exhaustion", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Negative level returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"level": -1,
		})
		req, _ := http.NewRequest("PATCH", "/api/characters/1/exhaustion", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Reset exhaustion to 0", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"level": 0,
		})
		req, _ := http.NewRequest("PATCH", "/api/characters/1/exhaustion", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestBatchSpellPrep(t *testing.T) {
	testDB := "/tmp/villum_spell_prep_test.db"
	os.Remove(testDB)

	if err := db.Init(testDB); err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDB)
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	auth.PUT("/characters/:id/spells/prepare", BatchPrepareSpells)

	// Seed data
	db.DB.Exec("INSERT INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")
	db.DB.Exec("INSERT INTO campaigns(id, user_id, name) VALUES(1, 1, 'Test Campaign')")
	db.DB.Exec("INSERT INTO characters(id, user_id, name, race, class, level, hp_max, hp_current, campaign_id, str, dex, con, int, wis, cha, ac, speed) VALUES(1, 1, 'Hero', 'Human', 'Fighter', 5, 50, 50, 1, 15, 14, 13, 12, 10, 8, 17, 30)")
	db.DB.Exec("INSERT INTO spells(id, character_id, name, level, school, prepared) VALUES(1, 1, 'Fireball', 3, 'evocation', true)")
	db.DB.Exec("INSERT INTO spells(id, character_id, name, level, school, prepared) VALUES(2, 1, 'Shield', 1, 'abjuration', true)")

	t.Run("Unprepares all spells when empty list", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"spell_ids": []int64{},
		})
		req, _ := http.NewRequest("PUT", "/api/characters/1/spells/prepare", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify all are unprepared
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM spells WHERE character_id=1 AND prepared=1").Scan(&count)
		if count != 0 {
			t.Errorf("expected 0 prepared spells, got %d", count)
		}
	})

	t.Run("Prepares specific spells", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"spell_ids": []int64{1},
		})
		req, _ := http.NewRequest("PUT", "/api/characters/1/spells/prepare", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM spells WHERE character_id=1 AND prepared=1").Scan(&count)
		if count != 1 {
			t.Errorf("expected 1 prepared spell, got %d", count)
		}

		var name string
		db.DB.QueryRow("SELECT name FROM spells WHERE character_id=1 AND prepared=1").Scan(&name)
		if name != "Fireball" {
			t.Errorf("expected Fireball to be prepared, got '%s'", name)
		}
	})
}

func TestDashboardNonExistentCampaign(t *testing.T) {
	testDB := "/tmp/villum_dash_404_test.db"
	os.Remove(testDB)

	if err := db.Init(testDB); err != nil {
		t.Fatalf("Failed to init test db: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(testDB)
	}()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())

	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	auth.GET("/campaigns/:id/dashboard", GetCampaignDashboard)

	t.Run("Non-existent campaign returns empty dashboard", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/campaigns/999/dashboard", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var dash map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &dash); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		// Dashboard should still return a valid JSON with empty fields
		if dash["name"] == nil {
			t.Error("expected dashboard to be returned even for non-existent campaign")
		}
	})
}
