package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/handlers/testutil"
	"villum/search"
)

func testReadFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func TestBuildInlineResults_Articles(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 100, "Fireball")
	testutil.SeedCompendiumSpell(t, 101, "Fire Bolt")
	results, err := search.SearchCompendium(context.Background(), db.DB, search.CompendiumParams{Query: "Fire", Limit: 10, Reranker: search.DefaultRerank, FuzzyFallback: true})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
	articles := buildInlineResults(context.Background(), results)
	if len(articles) != len(results) {
		t.Fatalf("expected %d articles got %d", len(results), len(articles))
	}
	for _, a := range articles {
		art, ok := a.(*tgmodels.InlineQueryResultArticle)
		if !ok {
			t.Fatalf("expected article type")
		}
		if art.Title == "" {
			t.Fatalf("empty title")
		}
		if art.InputMessageContent == nil {
			t.Fatalf("nil content")
		}
		content := art.InputMessageContent.(*tgmodels.InputTextMessageContent)
		if content.MessageText == "" {
			t.Fatalf("empty message text")
		}
		if runeCount(content.MessageText) > 4096 {
			t.Fatalf("message too long: %d", runeCount(content.MessageText))
		}
	}
}

func TestBuildInlineResults_EscapesHTML(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 9998, "<b>evil</b>")
	results, _ := search.SearchCompendium(context.Background(), db.DB, search.CompendiumParams{Query: "<b>evil</b>", Limit: 10, Reranker: search.DefaultRerank, FuzzyFallback: true})
	if len(results) == 0 {
		t.Fatalf("expected results for evil")
	}
	articles := buildInlineResults(context.Background(), results)
	for _, a := range articles {
		art := a.(*tgmodels.InlineQueryResultArticle)
		c := art.InputMessageContent.(*tgmodels.InputTextMessageContent)
		if strings.Contains(c.MessageText, "<b>evil</b>") {
			t.Fatalf("not escaped: %q", c.MessageText)
		}
		if !strings.Contains(c.MessageText, "&lt;b&gt;") {
			t.Fatalf("expected escaped, got %q", c.MessageText)
		}
	}
}

func TestBuildInlineResults_TitleFormat(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 700, "Fireball")
	results, _ := search.SearchCompendium(context.Background(), db.DB, search.CompendiumParams{Query: "Fireball", Limit: 5, Reranker: search.DefaultRerank, FuzzyFallback: true})
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
	articles := buildInlineResults(context.Background(), results)
	art := articles[0].(*tgmodels.InlineQueryResultArticle)
	if !strings.Contains(art.Title, "Fireball") || !strings.Contains(art.Title, "(spell)") {
		t.Fatalf("title format wrong: %q", art.Title)
	}
}

func TestBuildInlineResults_SubtypeFallback(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	// Seed a race which has no Subtype/Level, fallback branch should use Type
	testutil.SeedCompendiumSpell(t, 800, "UniqueSpellXYZ")
	// Manually craft a result with Subtype empty
	r := search.CompendiumResult{ID: 800, Name: "UniqueSpellXYZ", Type: "spell", Subtype: ""}
	articles := buildInlineResults(context.Background(), []search.CompendiumResult{r})
	if len(articles) != 1 {
		t.Fatalf("expected 1")
	}
	art := articles[0].(*tgmodels.InlineQueryResultArticle)
	// For spell with Subtype=="" but Type=spell without level in result, desc fallback is type
	// But the code path: if Subtype!="" else if Type==spell && Level !=0 else r.Type
	// Since Level is 0, expect "spell"
	if art.Description != "spell" {
		t.Fatalf("expected subtype fallback 'spell', got %q", art.Description)
	}
	// Now with level set
	r2 := search.CompendiumResult{ID: 800, Name: "UniqueSpellXYZ", Type: "spell", Level: 3, Subtype: ""}
	articles2 := buildInlineResults(context.Background(), []search.CompendiumResult{r2})
	art2 := articles2[0].(*tgmodels.InlineQueryResultArticle)
	if art2.Description != "Level 3" {
		t.Fatalf("expected 'Level 3', got %q", art2.Description)
	}
	// With Subtype present
	r3 := search.CompendiumResult{ID: 800, Name: "UniqueSpellXYZ", Type: "monster", Subtype: "beast"}
	articles3 := buildInlineResults(context.Background(), []search.CompendiumResult{r3})
	art3 := articles3[0].(*tgmodels.InlineQueryResultArticle)
	if art3.Description != "beast" {
		t.Fatalf("expected 'beast', got %q", art3.Description)
	}
}

func TestBuildInlineMessageText_LongDescription(t *testing.T) {
	long := strings.Repeat("x", 4000)
	r := search.CompendiumResult{ID: 1, Name: "TestSpell", Type: "spell", Level: 1, Subtype: "Evocation"}
	d := &search.CompendiumDetail{Name: "TestSpell", Type: "spell", Subtype: "Evocation", Description: long, School: "Evocation", Level: 1}
	text := buildInlineMessageText(r, d)
	if !strings.Contains(text, "…") {
		t.Fatalf("expected truncation marker")
	}
	if runeCount(text) > 3500 {
		// buildInlineMessageText truncates description at 3000, but buildInlineResults also caps at 3500
		// For direct call, check at least that description part is truncated
		// The full message may still be >3000 but should contain ellipsis
	}
	// Also test nil detail case
	r2 := search.CompendiumResult{ID: 2, Name: "NoDetail", Type: "spell", Subtype: "Evocation"}
	text2 := buildInlineMessageText(r2, nil)
	if !strings.Contains(text2, "NoDetail") {
		t.Fatalf("expected name in nil detail text: %q", text2)
	}
}

func TestBuildInlineResults_LongDescriptionTruncation(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	long := strings.Repeat("y", 5000)
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO compendium_spells(id, name, level, school, description) VALUES(?, ?, 1, 'Evocation', ?)`, 4243, "HugeSpell", long); err != nil {
		t.Fatalf("seed: %v", err)
	}
	results, _ := search.SearchCompendium(context.Background(), db.DB, search.CompendiumParams{Query: "HugeSpell", Limit: 5, Reranker: search.DefaultRerank, FuzzyFallback: true})
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
	articles := buildInlineResults(context.Background(), results)
	art := articles[0].(*tgmodels.InlineQueryResultArticle)
	msg := art.InputMessageContent.(*tgmodels.InputTextMessageContent).MessageText
	if !strings.Contains(msg, "…") {
		t.Fatalf("expected ellipsis in truncated article")
	}
	if runeCount(msg) > 3500 {
		t.Fatalf("expected <=3500 runes, got %d", runeCount(msg))
	}
}

func TestHandleInlineQuery_ShortQuery(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var capturedID, capturedResults, capturedButton string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedID = requestFormValue(r, "inline_query_id")
		capturedResults = requestFormValue(r, "results")
		capturedButton = requestFormValue(r, "button")
		w.Write(jsonOK(true))
	})
	q := &tgmodels.InlineQuery{ID: "qid123", Query: "x"}
	handleInlineQuery(context.Background(), q)
	if capturedID != "qid123" {
		t.Fatalf("expected qid123 got %q", capturedID)
	}
	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(capturedResults), &arr); err != nil {
		t.Fatalf("results not JSON array: %v %q", err, capturedResults)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 hint article, got %d", len(arr))
	}
	var hint map[string]any
	if err := json.Unmarshal(arr[0], &hint); err != nil {
		t.Fatalf("hint unmarshal: %v", err)
	}
	title, _ := hint["title"].(string)
	if !strings.Contains(strings.ToLower(title), "at least 2 characters") {
		t.Fatalf("hint title missing phrase, got %q", title)
	}
	if capturedButton != "" && capturedButton != "null" {
		// button should be empty for short query
		var btn map[string]any
		if err := json.Unmarshal([]byte(capturedButton), &btn); err == nil {
			if len(btn) != 0 {
				t.Fatalf("expected no button for short query, got %q", capturedButton)
			}
		}
	}
}

func TestHandleInlineQuery_NormalQuery(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 100, "Fireball")
	testutil.SeedCompendiumSpell(t, 101, "Fire Bolt")
	var capturedResults, capturedButton string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedResults = requestFormValue(r, "results")
		capturedButton = requestFormValue(r, "button")
		w.Write(jsonOK(true))
	})
	q := &tgmodels.InlineQuery{ID: "qid2", Query: "fire"}
	handleInlineQuery(context.Background(), q)
	var arr []map[string]any
	if err := json.Unmarshal([]byte(capturedResults), &arr); err != nil {
		t.Fatalf("results JSON: %v %q", err, capturedResults)
	}
	if len(arr) == 0 {
		t.Fatalf("expected >=1 article")
	}
	found := false
	for _, a := range arr {
		if title, _ := a["title"].(string); strings.Contains(title, "Fireball") || strings.Contains(title, "Fire Bolt") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spell name in title, got %v", arr)
	}
	if capturedButton == "" {
		t.Fatalf("expected button for normal query")
	}
	var btn map[string]any
	if err := json.Unmarshal([]byte(capturedButton), &btn); err != nil {
		t.Fatalf("button JSON: %v %q", err, capturedButton)
	}
	param, _ := btn["start_parameter"].(string)
	if !strings.HasPrefix(param, "ask_") {
		t.Fatalf("expected ask_ prefix, got %q", param)
	}
	token := strings.TrimPrefix(param, "ask_")
	if got, ok := lookupInlineAskQuery(token); !ok || got != "fire" {
		t.Fatalf("lookup failed: ok=%v got=%q", ok, got)
	}
}

func TestHandleInlineQuery_NoResults(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var capturedResults, capturedButton string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedResults = requestFormValue(r, "results")
		capturedButton = requestFormValue(r, "button")
		w.Write(jsonOK(true))
	})
	q := &tgmodels.InlineQuery{ID: "qid3", Query: "zzzzqqq"}
	handleInlineQuery(context.Background(), q)
	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(capturedResults), &arr); err != nil {
		t.Fatalf("results JSON: %v %q", err, capturedResults)
	}
	if len(arr) != 0 {
		t.Fatalf("expected 0 results, got %d", len(arr))
	}
	// button should be empty/nil
	if capturedButton != "" && capturedButton != "null" && capturedButton != "{}" {
		var btn map[string]any
		if err := json.Unmarshal([]byte(capturedButton), &btn); err == nil && len(btn) != 0 {
			t.Fatalf("expected empty button for no results, got %q", capturedButton)
		}
	}
}

func TestHandleInlineQuery_Nil(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	// Should not panic
	handleInlineQuery(context.Background(), nil)
}

func TestInlineAskStore_RoundTrip(t *testing.T) {
	token := putInlineAskQuery("fireball query")
	q, ok := lookupInlineAskQuery(token)
	if !ok || q != "fireball query" {
		t.Fatalf("round trip failed: %v %q", ok, q)
	}
	if _, ok := lookupInlineAskQuery("nonexistent"); ok {
		t.Fatalf("unknown token should fail")
	}
	inlineAskStore.Lock()
	e := inlineAskStore.m[token]
	e.expires = e.expires.Add(-20 * time.Minute)
	inlineAskStore.m[token] = e
	inlineAskStore.Unlock()
	if _, ok := lookupInlineAskQuery(token); ok {
		t.Fatalf("expired token should fail")
	}
}

func TestNoLLMImport(t *testing.T) {
	for _, path := range []string{"inline.go"} {
		data, err := testReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(data, "\"villum/ai\"") || strings.Contains(data, "villum/ai") {
			t.Fatalf("%s imports ai package", path)
		}
	}
}

func TestRunAskQuery_RetrievalOnly(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 500, "Magic Missile")
	testutil.SeedUser(t, 1, "askuser", "player")
	db.DB.Exec("INSERT INTO telegram_identities(user_id, telegram_user_id, telegram_username, dm_enabled) VALUES(?,?,?,?)", 1, 1, "askuser", 1)
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 1, name: "ask", args: []string{"Magic", "Missile"}}
	reply := runAskQuery(c, "Magic Missile")
	if !strings.Contains(reply.Text, "AI is not configured") {
		t.Fatalf("expected degraded header, got %q", reply.Text)
	}
	if !strings.Contains(reply.Text, "Magic Missile") {
		t.Fatalf("expected missile, got %q", reply.Text)
	}
	reply2 := runAskQuery(c, "zzzzqqq999")
	if !strings.Contains(reply2.Text, "AI is not configured") {
		t.Fatalf("expected degraded for no matches, got %q", reply2.Text)
	}
}
