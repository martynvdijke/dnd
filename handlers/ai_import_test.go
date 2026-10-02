package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

const testOneShotDraft = `{"title":"The Wolves of Welton","premise":"Sheep are missing","hook":"A farmer pleads",` +
	`"difficulty":"easy","estimated_minutes":180,"notes":"","acts":[{"title":"Arrival","description":"d",` +
	`"estimated_minutes":30,"scenes":[{"title":"The Farm","description":"s","scene_type":"roleplay","estimated_minutes":15}]}],` +
	`"npcs":[{"name":"Marla","race":"human","description":"guide","role":"ally"}],` +
	`"locations":[{"name":"Welton","type":"village","description":"sleepy"}],` +
	`"encounters":[{"name":"Wolf Pack","description":"e","difficulty":"easy"}],` +
	`"clues":[{"title":"Tracks","description":"c","clue_type":"object"}]}`

func countRows(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	return n
}

func TestAIDraftTokenBudget(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	epID := seedAIEndpoint(t, "budget", "text", "http://127.0.0.1:9", "m", true)
	defer cleanupAIEndpoints()
	ctx := context.Background()

	if got := *aiDraftTokenBudget(ctx, epID); got != defaultAIDraftMaxTokens {
		t.Fatalf("default budget = %d, want %d", got, defaultAIDraftMaxTokens)
	}
	set := func(v int) {
		t.Helper()
		if _, err := db.DB.Exec("UPDATE ai_endpoints SET max_tokens=? WHERE id=?", v, epID); err != nil {
			t.Fatalf("set max_tokens: %v", err)
		}
	}
	set(100)
	if got := *aiDraftTokenBudget(ctx, epID); got != minAIDraftMaxTokens {
		t.Fatalf("clamped low budget = %d, want %d", got, minAIDraftMaxTokens)
	}
	set(8000)
	if got := *aiDraftTokenBudget(ctx, epID); got != 8000 {
		t.Fatalf("configured budget = %d, want 8000", got)
	}
	set(400000000)
	if got := *aiDraftTokenBudget(ctx, epID); got != maxAIDraftMaxTokens {
		t.Fatalf("clamped high budget = %d, want %d", got, maxAIDraftMaxTokens)
	}
}

func TestValidateAIDraftReply(t *testing.T) {
	cases := []struct {
		name    string
		reply   string
		finish  string
		wantErr bool
	}{
		{"empty reply", "", "stop", true},
		{"whitespace reply", "   \n", "stop", true},
		{"truncated reply", `{"status":"ready","draft":{`, "length", true},
		{"good reply", `{"status":"ready","draft":{}}`, "stop", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAIDraftReply(tc.reply, tc.finish)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestAIDraftSystemPromptContract(t *testing.T) {
	prompt := aiDraftSystemPrompt("oneshot")
	for _, want := range []string{"2024", "One-shot drafting contract", "3-5", "scene_type", "clue_type", "never truncate"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("one-shot system prompt is missing %q", want)
		}
	}
	if strings.Contains(aiDraftSystemPrompt("npc"), "One-shot drafting contract") {
		t.Fatalf("non-oneshot prompt must not include the one-shot contract")
	}
}

// finishProvider returns a fake provider with a fixed finish_reason.
func finishProvider(t *testing.T, content, finish string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": content}, "finish_reason": finish},
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestAIDraftStartRejectsBrokenReplies(t *testing.T) {
	cases := []struct {
		name    string
		content string
		finish  string
	}{
		{"empty reply", "", "stop"},
		{"truncated reply", `{"status":"ready","message":"x","draft":{"title":`, "length"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := setupAIDraftRouter(t)
			defer cleanupAIEndpoints()
			defer testutil.CloseDB(t)
			epID := seedAIEndpoint(t, "broken", "text", finishProvider(t, tc.content, tc.finish).URL, "m", true)

			w := testutil.PostJSON(t, r, "/api/ai/draft", map[string]any{
				"endpoint_id": epID, "entity_type": "oneshot", "message": "go",
			})
			if w.Code != http.StatusBadGateway {
				t.Fatalf("status = %d, want 502; body %s", w.Code, w.Body.String())
			}
			if n := countRows(t, "SELECT COUNT(*) FROM ai_draft_sessions"); n != 0 {
				t.Fatalf("broken reply persisted %d sessions, want 0", n)
			}
		})
	}
}

func TestImportOneShotJSONCreates(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer testutil.CloseDB(t)

	w := testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"entity_type": "oneshot",
		"json":        "```json\n" + testOneShotDraft + "\n```",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", w.Code, w.Body.String())
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res["name"] != "The Wolves of Welton" || res["updated"] != false {
		t.Fatalf("unexpected result: %v", res)
	}
	counts, _ := res["counts"].(map[string]any)
	for key, want := range map[string]float64{"acts": 1, "scenes": 1, "npcs": 1, "locations": 1, "encounters": 1, "clues": 1} {
		if counts[key] != want {
			t.Fatalf("counts[%s] = %v, want %v", key, counts[key], want)
		}
	}
	for _, q := range []string{
		"SELECT COUNT(*) FROM oneshot_adventures",
		"SELECT COUNT(*) FROM oneshot_acts",
		"SELECT COUNT(*) FROM oneshot_scenes",
		"SELECT COUNT(*) FROM clues",
	} {
		if n := countRows(t, q); n != 1 {
			t.Fatalf("%s = %d, want 1", q, n)
		}
	}
	if n := countRows(t, "SELECT COUNT(*) FROM npcs"); n != 1 {
		t.Fatalf("npcs = %d, want 1", n)
	}
}

func TestImportOneShotJSONEnvelope(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer testutil.CloseDB(t)

	// Full assistant envelope with the draft as a string-encoded object.
	envelope := map[string]any{"status": "ready", "message": "Here you go", "draft": testOneShotDraft}
	raw, _ := json.Marshal(envelope)
	w := testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"json": string(raw),
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", w.Code, w.Body.String())
	}
	if n := countRows(t, "SELECT COUNT(*) FROM oneshot_adventures"); n != 1 {
		t.Fatalf("adventures = %d, want 1", n)
	}
}

func TestImportOneShotJSONRejects(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer testutil.CloseDB(t)

	cases := []struct {
		name string
		body map[string]any
		want int
	}{
		{"malformed json", map[string]any{"json": "not json at all"}, http.StatusBadRequest},
		{"missing title", map[string]any{"json": `{"premise":"no title"}`}, http.StatusBadRequest},
		{"unknown type", map[string]any{"entity_type": "dragon", "json": testOneShotDraft}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := testutil.PostJSON(t, r, "/api/ai/import", tc.body)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d; body %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
	if n := countRows(t, "SELECT COUNT(*) FROM oneshot_adventures"); n != 0 {
		t.Fatalf("rejected imports created %d adventures", n)
	}
}

func TestImportCampaignAccessDenied(t *testing.T) {
	testutil.NewDB(t)
	testutil.SeedUser(t, 1, "player", "player")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/ai")
	g.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "player")
		c.Set("role", "user")
		c.Set("session_id", "test")
		c.Next()
	})
	g.POST("/import", HandleImportDraftJSON)
	defer testutil.CloseDB(t)

	w := testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"campaign_id": 4242, "json": testOneShotDraft,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body %s", w.Code, w.Body.String())
	}
	if n := countRows(t, "SELECT COUNT(*) FROM oneshot_adventures"); n != 0 {
		t.Fatalf("denied import created %d adventures", n)
	}
}

func TestImportOneShotJSONReplaces(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer testutil.CloseDB(t)

	w := testutil.PostJSON(t, r, "/api/ai/import", map[string]any{"json": testOneShotDraft})
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := int64(created["id"].(float64))
	oldEncID := countRows(t, "SELECT id FROM encounter_templates LIMIT 1")

	revised := `{"title":"The Wolves of Welton (Revised)","premise":"p2","hook":"h2","difficulty":"hard",` +
		`"estimated_minutes":240,"notes":"n2","acts":[` +
		`{"title":"A1","description":"d","estimated_minutes":30,"scenes":[]},` +
		`{"title":"A2","description":"d","estimated_minutes":30,"scenes":[]}],` +
		`"npcs":[{"name":"Marla","race":"human","description":"updated","role":"ally"}],` +
		`"locations":[{"name":"Welton","type":"village","description":"updated"}],` +
		`"encounters":[{"name":"Alpha Wolf","description":"e2","difficulty":"hard"}],` +
		`"clues":[{"title":"Tracks","description":"c2","clue_type":"object"}]}`
	w = testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"entity_type": "oneshot", "entity_id": id, "json": revised,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("replace status = %d, body %s", w.Code, w.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["updated"] != true || int64(res["id"].(float64)) != id {
		t.Fatalf("unexpected replace result: %v", res)
	}
	counts, _ := res["counts"].(map[string]any)
	if counts["acts"] != float64(2) {
		t.Fatalf("acts after replace = %v, want 2", counts["acts"])
	}
	if n := countRows(t, "SELECT COUNT(*) FROM oneshot_adventures"); n != 1 {
		t.Fatalf("adventures = %d, want 1 (replace must not create)", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM oneshot_acts"); n != 2 {
		t.Fatalf("acts = %d, want 2", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM npcs"); n != 1 {
		t.Fatalf("npcs = %d, want 1 (same-named NPC reused)", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM locations"); n != 1 {
		t.Fatalf("locations = %d, want 1 (same-named location reused)", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM encounter_templates"); n != 1 {
		t.Fatalf("encounter templates = %d, want 1 (old template pruned)", n)
	}
	if n := countRows(t, "SELECT COUNT(*) FROM encounter_templates WHERE id=?", oldEncID); n != 0 {
		t.Fatalf("old encounter template %d still present", oldEncID)
	}
	var title string
	if err := db.DB.QueryRow("SELECT title FROM oneshot_adventures WHERE id=?", id).Scan(&title); err != nil || title != "The Wolves of Welton (Revised)" {
		t.Fatalf("title = %q err %v", title, err)
	}
}

func TestImportNPCReplacesAndChecksOwner(t *testing.T) {
	r := setupAIDraftRouter(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 2, "other", "other")

	w := testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"entity_type": "npc", "json": `{"name":"Marla","race":"human","description":"guide"}`,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := int64(created["id"].(float64))

	w = testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"entity_type": "npc", "entity_id": id,
		"json": `{"name":"Marla the Bold","race":"human","description":"updated"}`,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("replace status = %d, body %s", w.Code, w.Body.String())
	}
	var name string
	if err := db.DB.QueryRow("SELECT name FROM npcs WHERE id=?", id).Scan(&name); err != nil || name != "Marla the Bold" {
		t.Fatalf("name = %q err %v", name, err)
	}

	foreign, err := db.Client.NPC.Create().SetUserID(2).SetName("Foreign").Save(context.Background())
	if err != nil {
		t.Fatalf("seed foreign npc: %v", err)
	}
	w = testutil.PostJSON(t, r, "/api/ai/import", map[string]any{
		"entity_type": "npc", "entity_id": foreign.ID, "json": `{"name":"Stolen"}`,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("foreign replace status = %d, want 403; body %s", w.Code, w.Body.String())
	}
}

func TestReviseAIDraft(t *testing.T) {
	content, _ := json.Marshal(map[string]any{
		"status": "ready", "message": "Updated",
		"draft": json.RawMessage(`{"title":"New","premise":"p2"}`),
	})
	srv, seen := draftProvider(t, string(content))
	defer srv.Close()

	r := setupAIDraftRouter(t)
	defer cleanupAIEndpoints()
	defer testutil.CloseDB(t)
	seedAIEndpoint(t, "revise", "text", srv.URL, "m", true)

	w := testutil.PostJSON(t, r, "/api/ai/revise", map[string]any{
		"entity_type": "oneshot",
		"json":        `{"title":"Old","premise":"p"}`,
		"instruction": "rename the title to New",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["status"] != "ready" {
		t.Fatalf("status = %v, want ready", res["status"])
	}
	if draft, _ := json.Marshal(res["draft"]); !strings.Contains(string(draft), "New") {
		t.Fatalf("revised draft = %v", res["draft"])
	}
	msgs := (*seen)[0]
	if len(msgs) != 2 {
		t.Fatalf("provider got %d messages, want 2", len(msgs))
	}
	if !strings.Contains(msgs[0]["content"], "Revision mode") {
		t.Fatalf("system prompt missing revision addendum")
	}
	if !strings.Contains(msgs[1]["content"], "rename the title to New") || !strings.Contains(msgs[1]["content"], `"title":"Old"`) {
		t.Fatalf("user message missing instruction or current draft: %s", msgs[1]["content"])
	}
}

func TestReviseAIDraftErrors(t *testing.T) {
	t.Run("empty reply", func(t *testing.T) {
		r := setupAIDraftRouter(t)
		defer cleanupAIEndpoints()
		defer testutil.CloseDB(t)
		seedAIEndpoint(t, "revise-empty", "text", finishProvider(t, "", "stop").URL, "m", true)
		w := testutil.PostJSON(t, r, "/api/ai/revise", map[string]any{
			"entity_type": "oneshot", "json": `{"title":"Old"}`, "instruction": "change it",
		})
		if w.Code != http.StatusBadGateway {
			t.Fatalf("status = %d, want 502; body %s", w.Code, w.Body.String())
		}
	})
	t.Run("no endpoint", func(t *testing.T) {
		r := setupAIDraftRouter(t)
		defer testutil.CloseDB(t)
		w := testutil.PostJSON(t, r, "/api/ai/revise", map[string]any{
			"entity_type": "oneshot", "json": `{"title":"Old"}`, "instruction": "change it",
		})
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503; body %s", w.Code, w.Body.String())
		}
	})
	t.Run("missing instruction", func(t *testing.T) {
		r := setupAIDraftRouter(t)
		defer testutil.CloseDB(t)
		w := testutil.PostJSON(t, r, "/api/ai/revise", map[string]any{
			"entity_type": "oneshot", "json": `{"title":"Old"}`,
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body %s", w.Code, w.Body.String())
		}
	})
}
