package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func copilotRoutes(rg *gin.RouterGroup) {
	rg.POST("/campaigns/:id/copilot/chat", handleCopilotChat)
	rg.GET("/campaigns/:id/copilot/conversations", handleCopilotListConversations)
	rg.GET("/campaigns/:id/copilot/conversations/:cid", handleCopilotGetConversation)
	rg.POST("/campaigns/:id/copilot/prep", handleCopilotPrep)
	rg.POST("/campaigns/:id/copilot/transcript", handleCopilotTranscript)
	rg.POST("/campaigns/:id/copilot/transcript/summarize", handleCopilotTranscriptSummarize)
}

func TestRetrieveCampaignContext_Visibility(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedCampaign(t, 1, "Camp1", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	testutil.SeedCampaign(t, 2, "Camp2", "Party2", 1)
	// wiki page in camp1
	_, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 10, 1, 1, "WikiOne", "dragons are here uniqueWordWiki", "public")
	if err != nil {
		t.Fatalf("seed wiki: %v", err)
	}
	// shared knowledge
	_, err = db.DB.Exec("INSERT INTO campaign_knowledge(id,campaign_id,title,content,source,status,shared,status_history) VALUES(?,?,?,?,?,?,?,?)", 20, 1, "SharedKnow", "shared content uniqueWordShared", "src", "rumor", 1, "[]")
	if err != nil {
		t.Fatalf("seed shared: %v", err)
	}
	// unshared
	_, err = db.DB.Exec("INSERT INTO campaign_knowledge(id,campaign_id,title,content,source,status,shared,status_history) VALUES(?,?,?,?,?,?,?,?)", 21, 1, "SecretKnow", "secret content uniqueWordSecret", "src", "rumor", 0, "[]")
	if err != nil {
		t.Fatalf("seed unshared: %v", err)
	}
	// other campaign knowledge
	_, err = db.DB.Exec("INSERT INTO campaign_knowledge(id,campaign_id,title,content,source,status,shared,status_history) VALUES(?,?,?,?,?,?,?,?)", 30, 2, "OtherCamp", "other uniqueWordOther", "src", "rumor", 1, "[]")
	if err != nil {
		t.Fatalf("seed other: %v", err)
	}
	// as member B - should see shared+wiki, not unshared, not other camp
	sources, _, err := retrieveCampaignContext(1, 2, false, "uniqueWord", 8)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	foundShared := false
	foundWiki := false
	for _, s := range sources {
		if s.EntityType == "knowledge" && s.EntityID == 20 {
			foundShared = true
		}
		if s.EntityType == "wiki" && s.EntityID == 10 {
			foundWiki = true
		}
		if s.EntityType == "knowledge" && s.EntityID == 21 {
			t.Fatalf("member should not see unshared")
		}
		if s.EntityID == 30 {
			t.Fatalf("cross campaign leakage")
		}
	}
	if !foundShared {
		t.Fatalf("member should see shared, got %+v", sources)
	}
	if !foundWiki {
		t.Fatalf("member should see wiki, got %+v", sources)
	}
	// owner sees both
	sources2, _, _ := retrieveCampaignContext(1, 1, false, "uniqueWordSecret", 8)
	foundSecret := false
	for _, s := range sources2 {
		if s.EntityID == 21 {
			foundSecret = true
		}
	}
	if !foundSecret {
		t.Fatalf("owner should see unshared, got %+v", sources2)
	}
	// admin sees both
	sources3, _, _ := retrieveCampaignContext(1, 2, true, "uniqueWordSecret", 8)
	found := false
	for _, s := range sources3 {
		if s.EntityID == 21 {
			found = true
		}
	}
	if !found {
		t.Fatalf("admin should see unshared")
	}
	// no cross campaign
	sources4, _, _ := retrieveCampaignContext(1, 1, false, "uniqueWordOther", 8)
	for _, s := range sources4 {
		if s.EntityID == 30 {
			t.Fatalf("no cross campaign")
		}
	}
	if len(sources4) != 0 {
		// may be empty
	}
}

func TestGenerateText_MockProvider(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "hello"}, "finish_reason": "stop"}},
		})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	ep, err := db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "t", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "gpt-4o", Enabled: true})
	if err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}
	text, finish, err := generateText(context.Background(), ep.ID, "hi", "", nil, "")
	if err != nil {
		t.Fatalf("generateText err: %v", err)
	}
	if text != "hello" || finish != "stop" {
		t.Fatalf("got %q %q", text, finish)
	}
	_, _, err = generateText(context.Background(), 0, "hi", "", nil, "")
	if err == nil {
		t.Fatalf("expected error for missing endpoint")
	}
	if ae, ok := err.(*aiGenError); !ok || ae.Status != 400 {
		t.Fatalf("expected 400, got %v", err)
	}
}

func TestCopilotChat_PersistsAndCites(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	_, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 100, 1, 1, "AlphaWiki", "the dragon hoard is hidden uniqueDragonWord", "public")
	if err != nil {
		t.Fatalf("seed wiki: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "answer with cite"}, "finish_reason": "stop"}}})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	_, err = db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	if err != nil {
		t.Fatalf("seed ep: %v", err)
	}
	r := testutil.NewRouterWithUser(copilotRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/chat", map[string]any{"query": "uniqueDragonWord"})
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	sources, _ := resp["sources"].([]any)
	if len(sources) == 0 {
		t.Fatalf("expected sources, got %s", w.Body.String())
	}
	found := false
	for _, s := range sources {
		m, _ := s.(map[string]any)
		if m["entity_type"] == "wiki" && int(m["entity_id"].(float64)) == 100 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected wiki source, got %v", sources)
	}
	if resp["conversation_id"] == nil || resp["conversation_id"].(float64) == 0 {
		t.Fatalf("expected conversation_id, got %v", resp)
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(*) FROM copilot_conversations").Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected 1 conversation, got %d", cnt)
	}
	db.DB.QueryRow("SELECT COUNT(*) FROM copilot_messages").Scan(&cnt)
	if cnt != 2 {
		t.Fatalf("expected 2 messages, got %d", cnt)
	}
}

func TestCopilotChat_AIDisabledNoWrites(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	r := testutil.NewRouterWithUser(copilotRoutes, 1, "player")
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/chat", map[string]any{"query": "hello"})
	if w.Code != 503 {
		t.Fatalf("expected 503 got %d %s", w.Code, w.Body.String())
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(*) FROM copilot_conversations").Scan(&cnt)
	if cnt != 0 {
		t.Fatalf("expected no conversations, got %d", cnt)
	}
}

func TestCopilotConversations_OwnerIsolation(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "member", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "hi"}, "finish_reason": "stop"}}})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	_, err := db.DB.Exec("INSERT INTO campaign_wiki_pages(id,campaign_id,user_id,title,content,visibility) VALUES(?,?,?,?,?,?)", 200, 1, 1, "W", "content uniqueIsolation", "public")
	if err != nil {
		t.Fatalf("seed wiki: %v", err)
	}
	ownerRouter := testutil.NewRouterWithUser(copilotRoutes, 1, "player")
	w := testutil.PostJSON(t, ownerRouter, "/api/campaigns/1/copilot/chat", map[string]any{"query": "uniqueIsolation"})
	if w.Code != 200 {
		t.Fatalf("chat failed %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	cid := int(resp["conversation_id"].(float64))
	memberRouter := testutil.NewRouterWithUser(copilotRoutes, 2, "player")
	w2 := testutil.Get(t, memberRouter, fmt.Sprintf("/api/campaigns/1/copilot/conversations/%d", cid))
	if w2.Code != 404 {
		t.Fatalf("expected 404 got %d %s", w2.Code, w2.Body.String())
	}
}

func TestCopilotTranscript_IndexedAndSummarized(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	r := testutil.NewRouterWithUser(copilotRoutes, 1, "player")
	// transcript
	w := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript", map[string]any{"text": "the secret ritual was performed under moonlight uniqueTranscriptWord", "title": "Session 1 Transcript"})
	if w.Code != 200 {
		t.Fatalf("transcript failed %d %s", w.Code, w.Body.String())
	}
	var tres map[string]any
	json.Unmarshal(w.Body.Bytes(), &tres)
	id := int64(tres["id"].(float64))
	var cnt int
	db.DB.QueryRow("SELECT COUNT(*) FROM entity_search_index WHERE entity_type='wiki' AND entity_id=?", id).Scan(&cnt)
	if cnt < 1 {
		t.Fatalf("expected indexed wiki, got %d", cnt)
	}
	// MATCH query - need built query
	rows, _ := db.DB.Query("SELECT COUNT(*) FROM entity_search_index WHERE entity_search_index MATCH ?", buildFTS5Query("uniqueTranscriptWord"))
	if rows != nil {
		defer rows.Close()
		if rows.Next() {
			var c int
			rows.Scan(&c)
			if c < 1 {
				t.Fatalf("MATCH should return transcript")
			}
		}
	}
	// summarize
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "summary answer"}, "finish_reason": "stop"}}})
	}))
	defer srv.Close()
	enc, _ := crypto.Encrypt("sk-test")
	db.CreateAIEndpoint(context.Background(), &models.AIEndpoint{Name: "e2", Type: "text", BaseURL: srv.URL, EncryptedAPIKey: enc, Model: "m", Enabled: true})
	w2 := testutil.PostJSON(t, r, "/api/campaigns/1/copilot/transcript/summarize", map[string]any{"id": id})
	if w2.Code != 200 {
		t.Fatalf("summarize failed %d %s", w2.Code, w2.Body.String())
	}
	var sresp map[string]any
	json.Unmarshal(w2.Body.Bytes(), &sresp)
	if sresp["answer"] == nil || sresp["answer"] == "" {
		t.Fatalf("expected answer, got %v", sresp)
	}
	_ = gin.H{}
}
