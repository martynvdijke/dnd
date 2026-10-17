package telegram

import (
	"context"
	"strings"
	"testing"
	"time"

	"villum/db"
	"villum/handlers/testutil"
	"villum/search"
)

func TestRunCompendiumSearch(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 1, "Fireball")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"fire"}}
	reply := runCompendiumSearch(c, "")
	if !strings.Contains(reply.Text, "Fireball") {
		t.Fatalf("expected Fireball in reply %q", reply.Text)
	}
	if reply.Keyboard == nil {
		t.Fatalf("expected keyboard")
	}
}

func TestRunCompendiumSearchEmpty(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{}}
	reply := runCompendiumSearch(c, "")
	if !strings.Contains(reply.Text, "Usage:") {
		t.Fatalf("expected usage, got %q", reply.Text)
	}
}

func TestRunCompendiumSearchMisspelling(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 10, "Fireball")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"firebal"}}
	reply := runCompendiumSearch(c, "")
	if !strings.Contains(reply.Text, "Fireball") {
		t.Fatalf("expected fuzzy Fireball, got %q", reply.Text)
	}
}

func TestHandleCallbackDataSearch(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 20, "Fireball")
	testutil.SeedCompendiumSpell(t, 21, "Fire Bolt")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"fire"}}
	reply := runCompendiumSearch(c, "")
	if reply.Keyboard == nil || len(reply.Keyboard.InlineKeyboard) == 0 {
		t.Fatalf("no keyboard")
	}
	// extract token from first detail button
	var token string
	var idx int
	found := false
	for _, row := range reply.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, cbSearchDetailPrefix) {
				rest := strings.TrimPrefix(btn.CallbackData, cbSearchDetailPrefix)
				sep := strings.LastIndex(rest, ":")
				token = rest[:sep]
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("no detail button found")
	}
	// search page callback
	_, ok := handleCallbackData(c, cbSearchPrefix+token+":0")
	if !ok {
		t.Fatalf("page callback not handled")
	}
	// detail callback
	_ = idx
	reply2, ok := handleCallbackData(c, cbSearchDetailPrefix+token+":0")
	if !ok {
		t.Fatalf("detail callback not handled")
	}
	if !strings.Contains(reply2.Text, "Fire") {
		t.Fatalf("detail reply missing, got %q", reply2.Text)
	}
	// malformed token rejected
	if _, ok := handleCallbackData(c, cbSearchPrefix+"BAD!:0"); ok {
		t.Fatalf("malformed token should be rejected")
	}
	if _, ok := handleCallbackData(c, cbSearchDetailPrefix+"zzz:abc"); ok {
		t.Fatalf("malformed detail should be rejected")
	}
}

func TestDetailEscapesHTML(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 9999, "<b>evil</b>")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"<b>evil</b>"}}
	reply := runCompendiumSearch(c, "")
	if !strings.Contains(reply.Text, "&lt;b&gt;") {
		t.Fatalf("expected escaped list, got %q", reply.Text)
	}
	// detail also escapes
	if reply.Keyboard == nil {
		t.Fatalf("expected keyboard for detail test, reply=%q", reply.Text)
	}
	var token string
	for _, row := range reply.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, cbSearchDetailPrefix) {
				rest := strings.TrimPrefix(btn.CallbackData, cbSearchDetailPrefix)
				sep := strings.LastIndex(rest, ":")
				token = rest[:sep]
			}
		}
	}
	if token != "" {
		reply2, _ := handleCallbackData(c, cbSearchDetailPrefix+token+":0")
		if strings.Contains(reply2.Text, "<b>evil</b>") {
			t.Fatalf("detail not escaped: %q", reply2.Text)
		}
		if !strings.Contains(reply2.Text, "&lt;b&gt;") {
			t.Fatalf("detail should escape, got %q", reply2.Text)
		}
	}
}

func TestDetailTruncatesLongDescription(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	long := strings.Repeat("x", 5000)
	if _, err := db.DB.Exec(
		`INSERT OR IGNORE INTO compendium_spells(id, name, level, school, description) VALUES(?, ?, 1, 'Evocation', ?)`,
		4242, "Longspell", long,
	); err != nil {
		t.Fatalf("seed long spell: %v", err)
	}
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"Longspell"}}
	reply := runCompendiumSearch(c, "")
	if reply.Keyboard == nil {
		t.Fatalf("expected keyboard, reply=%q", reply.Text)
	}
	var token string
	for _, row := range reply.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, cbSearchDetailPrefix) {
				rest := strings.TrimPrefix(btn.CallbackData, cbSearchDetailPrefix)
				sep := strings.LastIndex(rest, ":")
				token = rest[:sep]
			}
		}
	}
	if token == "" {
		t.Fatalf("no detail token, reply=%q", reply.Text)
	}
	detail, ok := handleCallbackData(c, cbSearchDetailPrefix+token+":0")
	if !ok {
		t.Fatalf("detail callback not handled")
	}
	if n := runeCount(detail.Text); n > 4096 {
		t.Fatalf("detail text too long: %d runes", n)
	}
	if !strings.Contains(detail.Text, "…") {
		t.Fatalf("expected truncation marker, got %q", detail.Text)
	}
	if !strings.Contains(detail.Text, "(truncated)") {
		t.Fatalf("expected truncated note, got %q", detail.Text)
	}
}

func TestRenderSearchPage_Pagination(t *testing.T) {
	// directly test pagination with synthetic results
	results := make([]search.CompendiumResult, 6)
	for i := range results {
		results[i] = search.CompendiumResult{ID: int64(i + 1), Name: "PagSpell", Type: "spell", Subtype: "Evocation"}
	}
	token := "abc123"
	// render page 1 (last page) - should have Prev but no Next
	reply2 := renderSearchPage("PagSpell", results, 1, token)
	hasPrev, hasNext := false, false
	for _, row := range reply2.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.Contains(btn.Text, "Prev") {
				hasPrev = true
			}
			if strings.Contains(btn.Text, "Next") {
				hasNext = true
			}
		}
	}
	if !hasPrev {
		t.Fatalf("expected Prev on page 1")
	}
	if hasNext {
		t.Fatalf("expected no Next on last page")
	}
	// page 0 should have Next, no Prev
	reply0 := renderSearchPage("PagSpell", results, 0, token)
	hasPrev, hasNext = false, false
	for _, row := range reply0.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.Contains(btn.Text, "Prev") {
				hasPrev = true
			}
			if strings.Contains(btn.Text, "Next") {
				hasNext = true
			}
		}
	}
	if hasPrev {
		t.Fatalf("page 0 should not have Prev")
	}
	if !hasNext {
		t.Fatalf("page 0 should have Next")
	}
}

func TestHandleSearchPageCallback_ExpiredAndMalformed(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 3100, "Fireball")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"Fireball"}}
	reply := runCompendiumSearch(c, "")
	var token string
	for _, row := range reply.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, cbSearchDetailPrefix) {
				rest := strings.TrimPrefix(btn.CallbackData, cbSearchDetailPrefix)
				sep := strings.LastIndex(rest, ":")
				token = rest[:sep]
			}
		}
	}
	if token == "" {
		t.Fatalf("no token")
	}
	// expire token
	searchStore.Lock()
	if st, ok := searchStore.m[token]; ok {
		st.expires = st.expires.Add(-20 * 60 * 1000000000)
	}
	searchStore.Unlock()
	reply2, ok := handleCallbackData(c, cbSearchPrefix+token+":0")
	if !ok {
		t.Fatalf("expired should still be ok=true")
	}
	if !strings.Contains(reply2.Text, "expired") {
		t.Fatalf("expected expired message, got %q", reply2.Text)
	}
	// malformed token
	if _, ok := handleCallbackData(c, cbSearchPrefix+"BAD!:0"); ok {
		t.Fatalf("malformed token should be false")
	}
	// malformed page
	if _, ok := handleCallbackData(c, cbSearchPrefix+token+":abc"); ok {
		t.Fatalf("malformed page should be false")
	}
	// missing colon
	if _, ok := handleCallbackData(c, cbSearchPrefix+token); ok {
		t.Fatalf("missing colon should be false")
	}
}

func TestHandleSearchDetailCallback_EdgeCases(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 3200, "Fireball")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"Fireball"}}
	reply := runCompendiumSearch(c, "")
	var token string
	for _, row := range reply.Keyboard.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, cbSearchDetailPrefix) {
				rest := strings.TrimPrefix(btn.CallbackData, cbSearchDetailPrefix)
				sep := strings.LastIndex(rest, ":")
				token = rest[:sep]
			}
		}
	}
	if token == "" {
		t.Fatalf("no token")
	}
	// out-of-range index clamped, no panic
	reply2, ok := handleCallbackData(c, cbSearchDetailPrefix+token+":999")
	if !ok {
		t.Fatalf("oob index should be ok")
	}
	if reply2.Text == "" {
		t.Fatalf("expected text for oob")
	}
	// negative index clamped
	reply3, ok := handleCallbackData(c, cbSearchDetailPrefix+token+":-5")
	if !ok {
		t.Fatalf("negative index should be ok")
	}
	if reply3.Text == "" {
		t.Fatalf("expected text for negative index")
	}
	// expired token - insert expired entry then let getSearchState prune it
	searchStore.Lock()
	expiredToken := "abcdef12"
	searchStore.m[expiredToken] = &searchState{query: "x", results: nil, expires: time.Now().Add(-20 * time.Minute)}
	searchStore.Unlock()
	// also test the stored expired token path
	if _, ok := handleCallbackData(c, cbSearchDetailPrefix+expiredToken+":0"); !ok {
		t.Fatalf("stored expired detail should be true")
	}
	// create a valid-looking token that will be expired via getSearchState check
	// Use a hex token that is not in store -> expired path
	reply4, ok := handleCallbackData(c, cbSearchDetailPrefix+"aaaaaaaaaaaaaaaa:0")
	if !ok {
		t.Fatalf("expired detail should be true")
	}
	if !strings.Contains(reply4.Text, "expired") {
		t.Fatalf("expected expired, got %q", reply4.Text)
	}
	_ = expiredToken
	// malformed detail
	if _, ok := handleCallbackData(c, cbSearchDetailPrefix+"BAD!:0"); ok {
		t.Fatalf("malformed detail token should be false")
	}
}

func TestRunCompendiumSearch_TypeFilterAndNoResults(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedCompendiumSpell(t, 3300, "Fireball")
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "spell", args: []string{"Fireball"}}
	reply := runCompendiumSearch(c, "spell")
	if !strings.Contains(reply.Text, "Fireball") {
		t.Fatalf("type filter spell should find Fireball, got %q", reply.Text)
	}
	// no results with filter
	c2 := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "spell", args: []string{"zzzzqqq999nope"}}
	reply2 := runCompendiumSearch(c2, "spell")
	if !strings.Contains(reply2.Text, "No compendium results") {
		t.Fatalf("expected no results message, got %q", reply2.Text)
	}
	// no results without filter
	c3 := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 999, name: "search", args: []string{"zzzzqqq999nope"}}
	reply3 := runCompendiumSearch(c3, "")
	if !strings.Contains(reply3.Text, "No compendium results") {
		t.Fatalf("expected no results, got %q", reply3.Text)
	}
}
