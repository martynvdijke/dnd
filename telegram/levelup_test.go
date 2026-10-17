package telegram

import (
	"context"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestLevelupNoClaim(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if err := UpsertIdentity(1, 111, 111, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: nil}
	reply := runLevelup(c)
	if !strings.Contains(reply.Text, "No character claimed") {
		t.Fatalf("expected no-claim refusal, got %q", reply.Text)
	}
}

func TestLevelupConfirm(t *testing.T) {
	tgUID, chatID := setupActionCharacter(t, 20, 20, 0, 3, 14, "1d8", 3)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: tgUID}

	reply := runLevelup(c)
	if reply.Keyboard == nil {
		t.Fatalf("expected confirmation keyboard, got %q", reply.Text)
	}
	confirm := handleLevelupCallback(c, cbLevelupPrefix+"confirm")
	if !strings.Contains(confirm.Text, "level 4") {
		t.Fatalf("expected level 4, got %q", confirm.Text)
	}

	var level, hpMax, hdCurrent int
	if err := db.DB.QueryRow(`SELECT level, hp_max, hit_dice_current FROM characters WHERE id=1`).Scan(&level, &hpMax, &hdCurrent); err != nil {
		t.Fatalf("query: %v", err)
	}
	// d8 average 5 + con(14) modifier 2 = +7 → 20+7=27
	if level != 4 || hpMax != 27 || hdCurrent != 4 {
		t.Fatalf("unexpected after level up: level=%d hp_max=%d hd_current=%d", level, hpMax, hdCurrent)
	}

	if cancel := handleLevelupCallback(c, cbLevelupPrefix+"cancel"); !strings.Contains(cancel.Text, "cancelled") {
		t.Fatalf("expected cancel text, got %q", cancel.Text)
	}
}
