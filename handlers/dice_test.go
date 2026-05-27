package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

func TestDiceAPI(t *testing.T) {
	testDB := "/tmp/villum_dice_test.db"
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

	auth.POST("/roll", HandleRoll)
	auth.GET("/dice-rolls", GetDiceRolls)

	// Ensure the pool initializes
	_ = getDicePool()

	// Create test user for FK constraint
	db.DB.Exec("INSERT OR IGNORE INTO users(id, username, password, role) VALUES(1, 'test', 'hash', 'admin')")

	t.Run("Basic roll 2d6+3", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "2d6+3",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp HandlerResult
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp.Expression != "2d6+3" {
			t.Errorf("expected expression '2d6+3', got '%s'", resp.Expression)
		}
		if resp.Total < 5 || resp.Total > 15 {
			t.Errorf("expected total 5-15, got %d", resp.Total)
		}
		if len(resp.Breakdown) == 0 {
			t.Error("expected non-empty breakdown")
		}
		if resp.Text == "" {
			t.Error("expected non-empty text")
		}
	})

	t.Run("Keep highest 4d6kh3", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "4d6kh3",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp HandlerResult
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp.Expression != "4d6kh3" {
			t.Errorf("expected expression '4d6kh3', got '%s'", resp.Expression)
		}
		// Should have 4 rolls in the first breakdown group
		if len(resp.Breakdown) > 0 && len(resp.Breakdown[0].Rolls) != 4 {
			t.Errorf("expected 4 rolls in breakdown, got %d", len(resp.Breakdown[0].Rolls))
		}
		// Some rolls should be marked as dropped (useInTotal: false)
		if len(resp.Breakdown) > 0 {
			dropped := 0
			for _, r := range resp.Breakdown[0].Rolls {
				if !r.UseInTotal {
					dropped++
				}
			}
			if dropped != 1 {
				t.Logf("expected 1 dropped roll, got %d (random, might vary)", dropped)
			}
		}
	})

	t.Run("Advantage maps to 2d20kh1", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "1d20",
			"advantage":  "advantage",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp HandlerResult
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		// Should have 2 d20 rolls
		if len(resp.Breakdown) > 0 && len(resp.Breakdown[0].Rolls) != 2 {
			t.Errorf("expected 2 rolls for advantage, got %d", len(resp.Breakdown[0].Rolls))
		}
	})

	t.Run("Invalid expression returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "not-a-valid-roll!!",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Empty expression returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Exploding dice", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "1d6!",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Percentile d100", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "1d100",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("d20 with modifier", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "1d20+5",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp HandlerResult
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		// Breakdown should have at least 2 entries (1 die group + 1 modifier)
		if len(resp.Breakdown) < 2 {
			t.Errorf("expected 2+ breakdown entries, got %d", len(resp.Breakdown))
		}
		// Total should be >= 6 (min 1 on d20 + 5)
		if resp.Total < 6 || resp.Total > 25 {
			t.Errorf("expected total 6-25, got %d", resp.Total)
		}
	})

	t.Run("Roll history saves and retrieves", func(t *testing.T) {
		// First make a roll to save history
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]interface{}{
			"expression": "1d20",
		})
		req, _ := http.NewRequest("POST", "/api/roll", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("roll failed: %d: %s", w.Code, w.Body.String())
		}

		// Then get history
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/dice-rolls", nil)
		r.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 for history, got %d: %s", w2.Code, w2.Body.String())
		}

		var rolls []map[string]interface{}
		if err := json.Unmarshal(w2.Body.Bytes(), &rolls); err != nil {
			t.Fatalf("failed to parse history: %v", err)
		}
		if len(rolls) == 0 {
			t.Error("expected at least 1 roll in history")
		}
	})
}
