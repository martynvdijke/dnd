package telegram

import (
	"context"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
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
