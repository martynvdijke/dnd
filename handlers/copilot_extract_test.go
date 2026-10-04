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

func copilotExtractRoutes(rg *gin.RouterGroup) {
	rg.POST("/campaigns/:id/copilot/transcript", handleCopilotTranscript)
	rg.POST("/campaigns/:id/copilot/transcript/summarize", handleCopilotTranscriptSummarize)
	rg.POST("/campaigns/:id/copilot/transcript/:tid/extract", handleCopilotTranscriptExtract)
	rg.POST("/ai/draft/:id/commit", CommitAIDraft)
	rg.GET("/ai/draft/:id", GetAIDraft)
}

func TestCopilotExtract_Success(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	// seed transcript
	_, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 10, 1, 1, "Session 1", "transcript content long enough", "dm-only")
	if err != nil {
		t.Fatalf("seed transcript: %v", err)
	}
	// provider returning valid envelope
	draftObj := map[string]any{
		"recap": map[string]any{"title": "Recap Title", "content": "recap content here with key events"},
		"entities": map[string]any{
			"npcs":       []any{map[string]any{"name": "Gandrel", "race": "elf", "class": "wizard", "description": "wise", "notes": ""}},
			"locations":  []any{map[string]any{"name": "Old Mill", "type": "building", "description": "creaky mill"}},
			"encounters": []any{map[string]any{"name": "Wolf Ambush", "description": "wolves attack", "environment": "forest", "difficulty": "medium", "notes": ""}},
			"factions":   []any{map[string]any{"name": "Red Cloaks", "description": "bandits", "type": "gang", "headquarters": "forest"}},
		},
	}
	envelope := map[string]any{"status": "ready", "message": "done", "draft": draftObj}
	envelopeJSON, _ := json.Marshal(envelope)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": string(envelopeJSON)}, "finish_reason": "stop"}},
		})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	_, err = db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	if err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	r := testutil.NewRouterWithUser(copilotExtractRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript/10/extract", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Recap struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"recap"`
		Drafts []struct {
			ID         string `json:"id"`
			EntityType string `json:"entity_type"`
			Name       string `json:"name"`
		} `json:"drafts"`
		Source struct {
			EntityType string `json:"entity_type"`
			EntityID   int64  `json:"entity_id"`
			Title      string `json:"title"`
		} `json:"source"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v body %s", err, w.Body.String())
	}
	if resp.Recap.Title != "Recap Title" {
		t.Fatalf("recap title %q", resp.Recap.Title)
	}
	if len(resp.Drafts) != 4 {
		t.Fatalf("expected 4 drafts got %d %+v", len(resp.Drafts), resp.Drafts)
	}
	if resp.Source.EntityID != 10 || resp.Source.EntityType != "wiki" {
		t.Fatalf("source %+v", resp.Source)
	}
	// each persisted
	for _, d := range resp.Drafts {
		sess, err := loadAIDraftSession(d.ID)
		if err != nil {
			t.Fatalf("load draft %s: %v", d.ID, err)
		}
		if sess.Status != "ready" {
			t.Fatalf("status %q", sess.Status)
		}
		if sess.UserID != 1 {
			t.Fatalf("owner %d", sess.UserID)
		}
		if sess.CampaignID == nil || *sess.CampaignID != 1 {
			t.Fatalf("campaign %v", sess.CampaignID)
		}
	}
	// committable via existing commit endpoint
	first := resp.Drafts[0]
	w2 := testutil.PostJSON(t, r, "/api/ai/draft/"+first.ID+"/commit", map[string]any{})
	if w2.Code != 201 {
		t.Fatalf("commit failed %d %s", w2.Code, w2.Body.String())
	}
}

func TestCopilotExtract_Unparseable(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 11, 1, 1, "S", "content", "dm-only")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "not json at all <<<"}, "finish_reason": "stop"}},
		})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	r := testutil.NewRouterWithUser(copilotExtractRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript/11/extract", map[string]any{})
	if w.Code != 422 {
		t.Fatalf("expected 422 got %d %s", w.Code, w.Body.String())
	}
}

func TestCopilotExtract_AIDisabled(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 12, 1, 1, "S", "content", "dm-only")
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	r := testutil.NewRouterWithUser(copilotExtractRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript/12/extract", map[string]any{})
	if w.Code != 503 {
		t.Fatalf("expected 503 got %d %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] == nil {
		t.Fatalf("expected error")
	}
}

func TestCopilotExtract_NonMember(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "other", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 13, 1, 1, "S", "content", "dm-only")
	// need endpoint but should be blocked before
	enc, _ := crypto.Encrypt("sk-test")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "{}"}, "finish_reason": "stop"}}})
	}))
	defer srv.Close()
	db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	r := testutil.NewRouterWithUser(copilotExtractRoutes, 2, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript/13/extract", map[string]any{})
	if w.Code != 403 {
		t.Fatalf("expected 403 got %d %s", w.Code, w.Body.String())
	}
}

func TestCopilotExtract_ProviderError(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 14, 1, 1, "S", "content", "dm-only")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	r := testutil.NewRouterWithUser(copilotExtractRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript/14/extract", map[string]any{})
	if w.Code != 502 {
		t.Fatalf("expected 502 got %d %s", w.Code, w.Body.String())
	}
}
