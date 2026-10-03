package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func aiSearchRoutes(rg *gin.RouterGroup) {
	rg.POST("/search/ai", HandleAISearch)
}

func TestAISearch_Unauthenticated(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	r := gin.New()
	rg := r.Group("/api")
	rg.POST("/search/ai", HandleAISearch)
	w := testutil.PostJSON(t, r, "/api/search/ai", map[string]any{"query": "fireball"})
	if w.Code != 401 {
		t.Fatalf("expected 401 got %d %s", w.Code, w.Body.String())
	}
}

func TestAISearch_EmptyQuery(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	r := testutil.NewRouterWithUser(aiSearchRoutes, 1, "player")
	testutil.SeedUser(t, 1, "u1", "player")
	w := testutil.PostJSON(t, r, "/api/search/ai", map[string]any{"query": ""})
	if w.Code != 400 {
		t.Fatalf("expected 400 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil {
		t.Fatalf("expected error key, got %s", w.Body.String())
	}
}

func TestAISearch_NonMemberCampaign(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "other", "player")
	testutil.SeedCampaign(t, 10, "Camp", "Party", 1)
	r := testutil.NewRouterWithUser(aiSearchRoutes, 2, "player")
	cid := int64(10)
	w := testutil.PostJSON(t, r, "/api/search/ai", map[string]any{"query": "hello", "campaign_id": cid})
	if w.Code != 403 {
		t.Fatalf("expected 403 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "not a campaign member" {
		t.Fatalf("expected not a campaign member, got %v", resp)
	}
}

func TestAISearch_AIDisabled(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "u1", "player")
	testutil.SeedCompendiumSpell(t, 9001, "Fireball")
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	r := testutil.NewRouterWithUser(aiSearchRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/search/ai", map[string]any{"query": "Fireball"})
	if w.Code != 503 {
		t.Fatalf("expected 503 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil {
		t.Fatalf("expected error key, got %s", w.Body.String())
	}
	if resp["scope"] != "compendium" {
		t.Fatalf("expected scope compendium, got %v", resp["scope"])
	}
	srcs, ok := resp["sources"].([]any)
	if !ok || len(srcs) == 0 {
		t.Fatalf("expected non-empty sources, got %v", resp["sources"])
	}
}

func TestAISearch_MockEndpoint(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	_, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 500, 1, 1, "WikiFire", "Fireball is explosive uniqueFireballWord", "public")
	if err != nil {
		t.Fatalf("seed wiki: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "mock answer"}, "finish_reason": "stop"}}})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	_, err = db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "t", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	if err != nil {
		t.Fatalf("seed ep: %v", err)
	}
	r := testutil.NewRouterWithUser(aiSearchRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/search/ai", map[string]any{"query": "uniqueFireballWord", "campaign_id": 1})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["answer"] == nil || resp["answer"] == "" {
		t.Fatalf("expected answer, got %v", resp)
	}
	if _, ok := resp["sources"]; !ok {
		t.Fatalf("expected sources, got %v", resp)
	}
}

func TestAISearch_GenerationFailure(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	_, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 501, 1, 1, "WikiFire2", "Fireball is explosive uniqueFireballWord2", "public")
	if err != nil {
		t.Fatalf("seed wiki: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("provider error"))
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	_, err = db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "t", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	if err != nil {
		t.Fatalf("seed ep: %v", err)
	}
	r := testutil.NewRouterWithUser(aiSearchRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/search/ai", map[string]any{"query": "uniqueFireballWord2", "campaign_id": 1})
	if w.Code != 503 {
		t.Fatalf("expected 503 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil || resp["error"] == "" {
		t.Fatalf("expected non-empty error, got %v", resp)
	}
	srcs, ok := resp["sources"].([]any)
	if !ok || len(srcs) == 0 {
		t.Fatalf("expected non-empty sources, got %v", resp["sources"])
	}
	if resp["scope"] == nil || resp["scope"] == "" {
		t.Fatalf("expected non-empty scope, got %v", resp["scope"])
	}
}
