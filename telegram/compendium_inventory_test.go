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

func TestCompendiumAddInvalidToken(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: 1, tgUserID: 1}
	reply := handleCompendiumInventoryCallback(c, cbCompendiumAddPrefix+"zzzz:0")
	if !strings.Contains(reply.Text, "expired") && !strings.Contains(reply.Text, "invalid") {
		t.Fatalf("expected graceful failure, got %q", reply.Text)
	}
}

func TestCompendiumAddEquipment(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 3, 14, "1d8", 3)
	defer testutil.CloseDB(t)

	token := "abcd1234"
	searchStore.Lock()
	searchStore.m[token] = &searchState{
		results: []search.CompendiumResult{{Type: "equipment", ID: 7, Name: "Hempen Rope"}},
		expires: time.Now().Add(time.Minute),
	}
	searchStore.Unlock()
	defer func() {
		searchStore.Lock()
		delete(searchStore.m, token)
		searchStore.Unlock()
	}()

	c := &cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: tgUID}
	reply := handleCompendiumInventoryCallback(c, cbCompendiumAddPrefix+token+":0")
	if !strings.Contains(reply.Text, "Hempen Rope") {
		t.Fatalf("expected confirmation, got %q", reply.Text)
	}

	var name, category string
	var qty int
	if err := db.DB.QueryRow(`SELECT name, category, quantity FROM inventory WHERE character_id=1 ORDER BY id DESC LIMIT 1`).Scan(&name, &category, &qty); err != nil {
		t.Fatalf("query: %v", err)
	}
	if name != "Hempen Rope" || category != "equipment" || qty != 1 {
		t.Fatalf("unexpected inventory row: name=%q category=%q qty=%d", name, category, qty)
	}
}
