package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

// 7.1: concurrent combat-tracker mutations (HP updates, next-turn, current-turn)
// must not panic or return 5xx (SQLite busy_timeout=10000 + max 4 open conns).
func TestConcurrentCombatTracker(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 50, "Concurrent", "Party", 1)
	testutil.SeedCharacter(t, 100, 1, "Combat A", "Human", "Fighter")
	testutil.SeedCharacter(t, 101, 1, "Combat B", "Human", "Cleric")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/combat", CreateCombatEntry)
		auth.PUT("/combat/:id", UpdateCombatEntry)
		auth.POST("/combat/next-turn", NextTurn)
		auth.GET("/combat/current-turn", GetCurrentTurn)
	})

	fullEntry := func(name string, init int) map[string]any {
		return map[string]any{
			"character_id":    100,
			"campaign_id":     50,
			"name":            name,
			"type":            "monster",
			"initiative_roll": init,
			"initiative_mod":  0,
			"hp_max":          10,
			"hp_current":      10,
			"ac":              12,
			"turn_order":      0,
			"condition_ids":   "[]",
			"notes":           "",
			"is_active":       true,
		}
	}

	var ids []int64
	for i, name := range []string{"Goblin A", "Goblin B", "Goblin C"} {
		body, _ := json.Marshal(fullEntry(name, i+1))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/combat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create combat entry %s: expected 201, got %d: %s", name, rec.Code, rec.Body.String())
		}
		var out struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.ID == 0 {
			t.Fatalf("create combat entry %s: no id: %v %s", name, err, rec.Body.String())
		}
		ids = append(ids, out.ID)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 200)
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				id := ids[(g+i)%len(ids)]
				switch i % 3 {
				case 0:
					body, _ := json.Marshal(fullEntry("Goblin", 1))
					body, _ = json.Marshal(map[string]any{"hp_current": 10 - (i % 10), "hp_max": 10, "ac": 12})
					rec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/combat/%d", id), bytes.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					r.ServeHTTP(rec, req)
					if rec.Code >= http.StatusInternalServerError {
						errCh <- fmt.Errorf("PUT /combat/%d: %d", id, rec.Code)
						return
					}
				case 1:
					nb, _ := json.Marshal(map[string]any{"campaign_id": 50})
					rec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodPost, "/api/combat/next-turn", bytes.NewReader(nb))
					req.Header.Set("Content-Type", "application/json")
					r.ServeHTTP(rec, req)
					if rec.Code >= http.StatusInternalServerError {
						errCh <- fmt.Errorf("next-turn: %d", rec.Code)
						return
					}
				case 2:
					rec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/api/combat/current-turn?campaign_id=50", nil)
					r.ServeHTTP(rec, req)
					if rec.Code >= http.StatusInternalServerError {
						errCh <- fmt.Errorf("current-turn: %d", rec.Code)
						return
					}
				}
			}
		}(g)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent combat tracker: %v", err)
	}
}
