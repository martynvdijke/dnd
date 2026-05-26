package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

func TestHtmxCRUD(t *testing.T) {
	testDB := "/tmp/villum_htmx_test.db"
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

	htmxGroup := r.Group("/")
	htmxGroup.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	HtmxRegisterRoutes(htmxGroup)

	// Create a test user and character
	db.DB.Exec("INSERT INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")
	db.DB.Exec("INSERT INTO characters(id, user_id, name, race, class, level, hp_max, hp_current, str, dex, con, int, wis, cha, ac, speed) VALUES(1, 1, 'TestChar', 'Human', 'Fighter', 1, 10, 10, 10, 10, 10, 10, 10, 10, 10, 30)")
	db.DB.Exec("INSERT INTO campaigns(id, name, user_id) VALUES(1, 'Test Campaign', 1)")

	t.Run("ListNotes - returns HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/notes?character_id=1", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "No notes yet") {
			t.Errorf("expected empty state message, got: %s", body)
		}
		if !strings.Contains(body, "New Note") {
			t.Errorf("expected New Note button, got: %s", body)
		}
	})

	t.Run("CreateNote - creates and returns HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "title": {"Test Note"}, "content": {"Test content"}, "visibility": {"player"}, "category": {"general"}}
		req, _ := http.NewRequest("POST", "/htmx/notes", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "Test Note") {
			t.Errorf("expected new note in list, got: %s", body)
		}
	})

	t.Run("ListNotes - shows created note", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/notes?character_id=1", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Test Note") {
			t.Errorf("expected note in list")
		}
	})

	t.Run("EditNoteForm - returns edit form", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/notes/1/edit", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "Test Note") {
			t.Errorf("expected note title in form, got: %s", body)
		}
		if !strings.Contains(body, "Save Note") {
			t.Errorf("expected Save Note button, got: %s", body)
		}
	})

	t.Run("UpdateNote - updates and returns HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "title": {"Updated Note"}, "content": {"Updated content"}, "visibility": {"dm"}, "category": {"lore"}}
		req, _ := http.NewRequest("PUT", "/htmx/notes/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Updated Note") {
			t.Errorf("expected updated note in list")
		}
		if !strings.Contains(w.Body.String(), "DM only") {
			t.Errorf("expected DM only icon for dm visibility")
		}
	})

	t.Run("DeleteNote - deletes and returns HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/htmx/notes/1", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "Updated Note") {
			t.Errorf("expected note to be removed from list")
		}
	})

	t.Run("FeatsCRUD - full cycle", func(t *testing.T) {
		// Create
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "name": {"Great Weapon Master"}, "description": {"Power attack"}, "prerequisites": {"Str 13+"}, "source": {"PHB"}, "level_gained": {"4"}}
		req, _ := http.NewRequest("POST", "/htmx/feats", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create feats: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Great Weapon Master") {
			t.Errorf("expected feat in list after create")
		}

		// List
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/htmx/feats?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("list feats: expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Great Weapon Master") {
			t.Errorf("expected feat in list")
		}

		// Edit form
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/htmx/feats/1/edit", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("edit feats: expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Great Weapon Master") {
			t.Errorf("expected feat name in form")
		}

		// Update
		w = httptest.NewRecorder()
		form = url.Values{"character_id": {"1"}, "name": {"GWM"}, "description": {""}, "prerequisites": {"Str 13+"}, "source": {"PHB"}, "level_gained": {"4"}}
		req, _ = http.NewRequest("PUT", "/htmx/feats/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("update feats: expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "GWM") {
			t.Errorf("expected updated feat name")
		}

		// Delete
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/htmx/feats/1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("delete feats: expected 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "GWM") {
			t.Errorf("expected feat to be removed")
		}
	})

	t.Run("ConditionsCRUD - full cycle", func(t *testing.T) {
		// Create
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "name": {"Poisoned"}, "type": {"poisoned"}, "source": {"Spider"}, "duration": {"10"}, "duration_type": {"round"}, "save_dc": {"12"}, "description": {"Turned green"}}
		req, _ := http.NewRequest("POST", "/htmx/conditions", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create condition: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "Poisoned") {
			t.Errorf("expected condition in list, got: %s", body)
		}
		if !strings.Contains(body, "10 round") {
			t.Errorf("expected duration display")
		}

		// List
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/htmx/conditions?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("list conditions: expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Poisoned") {
			t.Errorf("expected condition in list")
		}
	})

	t.Run("LocationsCRUD - create location and link to character", func(t *testing.T) {
		// Create location and link
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "name": {"Waterdeep"}, "type": {"city"}, "description": {"City of Splendors"}}
		req, _ := http.NewRequest("POST", "/htmx/locations", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create location: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Waterdeep") {
			t.Errorf("expected location in list after create")
		}
	})

	t.Run("FactionsCRUD - create and delete", func(t *testing.T) {
		// Create
		w := httptest.NewRecorder()
		form := url.Values{"name": {"Harpers"}, "type": {"guild"}, "description": {"Secret organization"}, "headquarters": {"Waterdeep"}}
		req, _ := http.NewRequest("POST", "/htmx/factions", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create faction: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Harpers") {
			t.Errorf("expected faction in list after create")
		}

		// List
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/htmx/factions", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("list factions: expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Harpers") {
			t.Errorf("expected faction in list")
		}

		// Delete
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/htmx/factions/1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("delete faction: expected 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "Harpers") {
			t.Errorf("expected faction to be removed")
		}
	})

	t.Run("CompanionsCRUD - create companion", func(t *testing.T) {
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "name": {"Shadow"}, "type": {"companion"}, "race": {"Wolf"}, "hp_max": {"20"}, "ac": {"13"}, "str": {"14"}, "dex": {"15"}, "con": {"12"}, "int": {"3"}, "wis": {"12"}, "cha": {"6"}, "speed": {"50"}, "abilities": {"Keen senses"}, "notes": {"Good boy"}, "is_alive": {"true"}}
		req, _ := http.NewRequest("POST", "/htmx/companions", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create companion: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Shadow") {
			t.Errorf("expected companion in list after create, got: %s", w.Body.String())
		}
	})

	t.Run("JournalCRUD - create entry", func(t *testing.T) {
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "title": {"Day 1"}, "entry": {"Today we fought a dragon"}, "entry_date": {"2025-01-01"}}
		req, _ := http.NewRequest("POST", "/htmx/journal", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create journal: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Day 1") {
			t.Errorf("expected journal entry in list")
		}
	})

	t.Run("SessionsCRUD - create session", func(t *testing.T) {
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "title": {"Session 1"}, "session_date": {"2025-01-01"}, "xp_earned": {"300"}, "gold_earned": {"50"}, "important_events": {"Defeated goblins"}, "notes": {"Fun session"}}
		req, _ := http.NewRequest("POST", "/htmx/sessions", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create session: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Session 1") {
			t.Errorf("expected session in list")
		}
	})

	t.Run("QuestsCRUD - create quest", func(t *testing.T) {
		w := httptest.NewRecorder()
		form := url.Values{"character_id": {"1"}, "name": {"Save the village"}, "description": {"Goblins attacking"}, "status": {"active"}, "objectives": {"Kill goblin chief"}, "rewards": {"500gp"}, "notes": {"Important"}}
		req, _ := http.NewRequest("POST", "/htmx/quests", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create quest: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "Save the village") {
			t.Errorf("expected quest in list")
		}
	})

	t.Run("ServiceUnavailable - 404 for unknown route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/nonexistent", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestHtmxNewForms(t *testing.T) {
	testDB := "/tmp/villum_htmx_form_test.db"
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

	htmxGroup := r.Group("/")
	htmxGroup.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	HtmxRegisterRoutes(htmxGroup)

	t.Run("NewNoteForm", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/notes/new?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Create Note") {
			t.Errorf("expected Create Note button, got: %s", body)
		}
		if !strings.Contains(body, `hx-post="/htmx/notes"`) {
			t.Errorf("expected hx-post for create, got: %s", body)
		}
	})

	t.Run("NewFeatForm", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/feats/new?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Add Feat") {
			t.Errorf("expected Add Feat button, got: %s", body)
		}
	})

	t.Run("NewConditionForm", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/conditions/new?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Add Condition") {
			t.Errorf("expected Add Condition button, got: %s", body)
		}
	})

	t.Run("NewInventoryForm", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/inventory/new?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Add Item") {
			t.Errorf("expected Add Item button, got: %s", body)
		}
	})

	t.Run("NewSpellForm", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/spells/new?character_id=1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Add Spell") {
			t.Errorf("expected Add Spell button, got: %s", body)
		}
	})
}

func TestHtmxErrorHandling(t *testing.T) {
	testDB := "/tmp/villum_htmx_err_test.db"
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

	htmxGroup := r.Group("/")
	htmxGroup.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	HtmxRegisterRoutes(htmxGroup)

	t.Run("MissingCharacterID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/notes", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for missing character_id, got %d", w.Code)
		}
	})

	t.Run("NotFoundItem", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/notes/999/edit", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 for non-existent item, got %d", w.Code)
		}
	})

	t.Run("ContentType", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/htmx/factions", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		ct := w.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("expected text/html content type, got: %s", ct)
		}
	})
}

func TestHtmxContentTypeAndTemplate(t *testing.T) {
	testDB := "/tmp/villum_htmx_ct_test.db"
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

	htmxGroup := r.Group("/")
	htmxGroup.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "testuser")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})

	HtmxRegisterRoutes(htmxGroup)

	tests := []struct {
		name string
		path string
	}{
		{"Timeline List", "/htmx/timeline"},
		{"Factions List", "/htmx/factions"},
		{"Notes List", "/htmx/notes?character_id=1"},
		{"Feats List", "/htmx/feats?character_id=1"},
		{"Conditions List", "/htmx/conditions?character_id=1"},
		{"Companions List", "/htmx/companions?character_id=1"},
		{"Inventory List", "/htmx/inventory?character_id=1"},
		{"Spells List", "/htmx/spells?character_id=1"},
		{"Sessions List", "/htmx/sessions?character_id=1"},
		{"Quests List", "/htmx/quests?character_id=1"},
		{"Journal List", "/htmx/journal?character_id=1"},
		{"NPCs List", "/htmx/npcs?character_id=1"},
		{"Locations List", "/htmx/locations?character_id=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
				return
			}
			ct := w.Header().Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("expected text/html content type, got: %s", ct)
			}
			body := w.Body.String()
			if len(body) < 10 {
				t.Errorf("expected non-empty response body, got length %d", len(body))
			}
			// Verify it looks like HTML
			if !strings.Contains(body, "<") || !strings.Contains(body, ">") {
				t.Errorf("response doesn't look like HTML: %s", body)
			}
		})
	}
}
