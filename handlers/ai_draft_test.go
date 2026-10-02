package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func setupAIDraftRouter(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/ai")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "admin")
		c.Set("role", "admin")
		c.Set("session_id", "test")
		c.Next()
	})
	g.POST("/draft", StartAIDraft)
	g.GET("/draft", ListAIDrafts)
	g.GET("/draft/:id", GetAIDraft)
	g.POST("/draft/:id/turn", AIDraftTurn)
	g.POST("/draft/:id/commit", CommitAIDraft)
	g.DELETE("/draft/:id", DiscardAIDraft)
	g.POST("/import", HandleImportDraftJSON)
	g.POST("/revise", ReviseAIDraft)
	return r
}

// draftProvider starts a fake OpenAI-compatible server returning the given
// assistant content, and records every message array it receives.
func draftProvider(t *testing.T, content string) (*httptest.Server, *[][]map[string]string) {
	t.Helper()
	var mu sync.Mutex
	seen := &[][]map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Messages []map[string]string `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		*seen = append(*seen, payload.Messages)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": content}, "finish_reason": "stop"},
			},
		})
	}))
	return srv, seen
}

func TestParseAIDraftReply(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantStatus string
		wantDraft  bool
	}{
		{"chatting", `{"status":"chatting","message":"What tone?","draft":null}`, "chatting", false},
		{"ready", `{"status":"ready","message":"Done","draft":{"title":"X"}}`, "ready", true},
		{"fenced", "```json\n{\"status\":\"ready\",\"message\":\"Done\",\"draft\":{\"title\":\"X\"}}\n```", "ready", true},
		{"prose wrapped", `Here you go: {"status":"chatting","message":"Hi","draft":null} hope that helps`, "chatting", false},
		{"draft as string", `{"status":"ready","message":"Done","draft":"{\"title\":\"Y\"}"}`, "ready", true},
		{"fallback plain text", "Sure, tell me more about the setting.", "chatting", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _, draft := parseAIDraftReply(tc.in)
			if status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", status, tc.wantStatus)
			}
			if (len(draft) > 0) != tc.wantDraft {
				t.Fatalf("draft present = %v, want %v (draft=%q)", len(draft) > 0, tc.wantDraft, string(draft))
			}
		})
	}
}

func TestAIDraftStartAndCommitOneShot(t *testing.T) {
	draft := `{"title":"The Sunken Vault","premise":"A flooded crypt","hook":"A dying sailor's map",` +
		`"difficulty":"hard","estimated_minutes":240,"notes":"","acts":[{"title":"Descent","description":"d",` +
		`"estimated_minutes":40,"scenes":[{"title":"The Rope","description":"s","scene_type":"exploration","estimated_minutes":15}]}],` +
		`"npcs":[{"name":"Marla","race":"human","description":"guide","role":"ally"}],` +
		`"locations":[{"name":"Sunken Vault","type":"dungeon","description":"flooded"}],` +
		`"encounters":[{"name":"Drowned Dead","description":"e","difficulty":"hard"}],` +
		`"clues":[{"title":"Salt Map","description":"c","clue_type":"object"}]}`
	content, _ := json.Marshal(map[string]any{"status": "ready", "message": "Here's the draft", "draft": json.RawMessage(draft)})

	srv, _ := draftProvider(t, string(content))
	defer srv.Close()

	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	epID := seedAIEndpoint(t, "draft-text", "text", srv.URL, "gpt-4o", true)

	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": epID, "entity_type": "oneshot", "message": "a sunken vault",
	})
	if w.Code != 201 {
		t.Fatalf("start status = %d, body %s", w.Code, w.Body.String())
	}
	var start map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &start); err != nil {
		t.Fatalf("unmarshal start: %v", err)
	}
	if start["status"] != "ready" {
		t.Fatalf("status = %v, want ready", start["status"])
	}
	if start["draft"] == nil {
		t.Fatalf("expected a draft in start response: %v", start)
	}
	id, _ := start["id"].(string)
	if id == "" {
		t.Fatalf("expected session id")
	}

	wc := testutil.PostJSON(t, r, "/api/ai/draft/"+id+"/commit", map[string]any{})
	if wc.Code != 201 {
		t.Fatalf("commit status = %d, body %s", wc.Code, wc.Body.String())
	}
	var commit map[string]any
	if err := json.Unmarshal(wc.Body.Bytes(), &commit); err != nil {
		t.Fatalf("unmarshal commit: %v", err)
	}
	advID := int64(commit["entity_id"].(float64))
	if advID <= 0 {
		t.Fatalf("expected positive adventure id, got %v", commit["entity_id"])
	}

	var title string
	if err := db.DB.QueryRow("SELECT title FROM oneshot_adventures WHERE id=?", advID).Scan(&title); err != nil {
		t.Fatalf("query adventure: %v", err)
	}
	if title != "The Sunken Vault" {
		t.Fatalf("title = %q", title)
	}
	var actCount, sceneCount, npcCount, clueCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_acts WHERE adventure_id=?", advID).Scan(&actCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_scenes s JOIN oneshot_acts a ON s.act_id=a.id WHERE a.adventure_id=?", advID).Scan(&sceneCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM oneshot_adventure_npcs WHERE adventure_id=?", advID).Scan(&npcCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM clues WHERE adventure_id=?", advID).Scan(&clueCount)
	if actCount != 1 || sceneCount != 1 || npcCount != 1 || clueCount != 1 {
		t.Fatalf("created counts acts=%d scenes=%d npcs=%d clues=%d", actCount, sceneCount, npcCount, clueCount)
	}
}

func TestAIDraftTurnSendsHistory(t *testing.T) {
	content := `{"status":"chatting","message":"Tell me more","draft":null}`
	srv, seen := draftProvider(t, content)
	defer srv.Close()

	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	epID := seedAIEndpoint(t, "draft-mem", "text", srv.URL, "gpt-4o", true)

	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": epID, "entity_type": "npc", "message": "a grumpy blacksmith",
	})
	if w.Code != 201 {
		t.Fatalf("start status = %d %s", w.Code, w.Body.String())
	}
	var start map[string]any
	json.Unmarshal(w.Body.Bytes(), &start)
	id := start["id"].(string)

	w2 := testutil.PostJSON(t, r, "/api/ai/draft/"+id+"/turn", map[string]any{
		"endpoint_id": epID, "message": "make her a dwarf",
	})
	if w2.Code != 200 {
		t.Fatalf("turn status = %d %s", w2.Code, w2.Body.String())
	}

	if len(*seen) != 2 {
		t.Fatalf("expected 2 provider calls, got %d", len(*seen))
	}
	second := (*seen)[1]
	// system + 2 user turns + 1 assistant turn = 4 messages.
	if len(second) != 4 {
		t.Fatalf("expected full history in second call, got %d messages: %+v", len(second), second)
	}
	if second[0]["role"] != "system" {
		t.Fatalf("first message should be system, got %q", second[0]["role"])
	}
	if second[1]["content"] != "a grumpy blacksmith" || second[3]["content"] != "make her a dwarf" {
		t.Fatalf("history not preserved: %+v", second)
	}
}

func TestAIDraftCommitRejectedWhenNotReady(t *testing.T) {
	content := `{"status":"chatting","message":"Any preferences?","draft":null}`
	srv, _ := draftProvider(t, content)
	defer srv.Close()

	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	epID := seedAIEndpoint(t, "draft-chat", "text", srv.URL, "gpt-4o", true)

	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": epID, "entity_type": "location", "message": "a swamp",
	})
	var start map[string]any
	json.Unmarshal(w.Body.Bytes(), &start)
	id := start["id"].(string)

	wc := testutil.PostJSON(t, r, "/api/ai/draft/"+id+"/commit", map[string]any{})
	if wc.Code != 400 {
		t.Fatalf("expected 400 for not-ready commit, got %d %s", wc.Code, wc.Body.String())
	}
}

func TestAIDraftRejectsUnsupportedEntity(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer testutil.CloseDB(t)
	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": 1, "entity_type": "dragon", "message": "hi",
	})
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}

func TestAIDraftListAndDiscard(t *testing.T) {
	content := `{"status":"chatting","message":"ok","draft":null}`
	srv, _ := draftProvider(t, content)
	defer srv.Close()

	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	epID := seedAIEndpoint(t, "draft-list", "text", srv.URL, "gpt-4o", true)

	w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
		"endpoint_id": epID, "entity_type": "faction", "message": "a thieves guild",
	})
	var start map[string]any
	json.Unmarshal(w.Body.Bytes(), &start)
	id := start["id"].(string)

	wl := testutil.Get(t, r, "/api/ai/draft")
	if wl.Code != 200 {
		t.Fatalf("list status = %d", wl.Code)
	}
	var list []map[string]any
	json.Unmarshal(wl.Body.Bytes(), &list)
	if len(list) != 1 || list[0]["id"] != id {
		t.Fatalf("expected the session in the list, got %v", list)
	}

	req, _ := http.NewRequest("DELETE", "/api/ai/draft/"+id, nil)
	wd := httptest.NewRecorder()
	r.ServeHTTP(wd, req)
	if wd.Code != 200 {
		t.Fatalf("discard status = %d %s", wd.Code, wd.Body.String())
	}
	if _, err := loadAIDraftSession(id); err == nil {
		t.Fatalf("expected session to be deleted")
	}
	_ = context.Background()
}
