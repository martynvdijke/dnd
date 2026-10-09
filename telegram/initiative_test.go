package telegram

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestInitiativeAddRenderAndClear(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	const chatID = -100555
	c := &cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: 1}

	if reply := runInitiative(&cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: 1, args: []string{"add", "Goblin", "15", "7"}}); !strings.Contains(reply.Text, "Goblin") {
		t.Fatalf("expected Goblin in render, got %q", reply.Text)
	}
	reply := runInitiative(&cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: 1, args: []string{"add", "Aria", "20"}})
	if reply.Keyboard == nil {
		t.Fatalf("expected initiative keyboard")
	}
	if strings.Index(reply.Text, "Aria") > strings.Index(reply.Text, "Goblin") {
		t.Fatalf("expected Aria (20) before Goblin (15): %q", reply.Text)
	}

	next := handleInitiativeCallback(c, cbInitPrefix+"next")
	if next.Keyboard == nil {
		t.Fatalf("expected keyboard after next")
	}
	var active int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM telegram_initiative WHERE chat_id=? AND turn_order=1`, chatID).Scan(&active); err != nil {
		t.Fatalf("query: %v", err)
	}
	if active != 1 {
		t.Fatalf("expected exactly one active combatant, got %d", active)
	}

	cleared := handleInitiativeCallback(c, cbInitPrefix+"clear")
	if !strings.Contains(cleared.Text, "No combatants") {
		t.Fatalf("expected empty tracker after clear, got %q", cleared.Text)
	}
}

func TestInitiativeRemoveByID(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	const chatID = -100556
	runInitiative(&cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: 1, args: []string{"add", "Goblin", "15"}})
	var id int64
	if err := db.DB.QueryRow(`SELECT id FROM telegram_initiative WHERE chat_id=?`, chatID).Scan(&id); err != nil {
		t.Fatalf("query: %v", err)
	}
	reply := handleInitiativeCallback(&cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: 1}, cbInitPrefix+"rm:"+strconv.FormatInt(id, 10))
	if !strings.Contains(reply.Text, "No combatants") {
		t.Fatalf("expected tracker empty after remove, got %q", reply.Text)
	}
}
