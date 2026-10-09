package telegram

import (
	"context"
	"fmt"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestQuestToggleCallback(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "admin")
	testutil.SeedCharacter(t, 1, 1, "Aria", "Elf", "Ranger")
	testutil.SeedCampaign(t, 10, "Test Campaign", "Party", 1)
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO campaign_characters(campaign_id, character_id) VALUES(10, 1)`); err != nil {
		t.Fatalf("campaign_characters: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR REPLACE INTO campaign_telegram_settings(campaign_id, chat_id, chat_type, is_enabled) VALUES(10, -100, 'supergroup', 1)`); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	res, err := db.DB.Exec(`INSERT INTO quests(character_id, name, status) VALUES(1, 'Find the relic', 'active')`)
	if err != nil {
		t.Fatalf("seed quest: %v", err)
	}
	id, _ := res.LastInsertId()

	c := &cmdContext{ctx: context.Background(), chatID: -100, tgUserID: 1}
	handleQuestCallback(c, fmt.Sprintf("quest:done:%d", id))

	var status string
	if err := db.DB.QueryRow(`SELECT status FROM quests WHERE id=?`, id).Scan(&status); err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != "complete" {
		t.Fatalf("expected quest complete, got %q", status)
	}

	handleQuestCallback(c, fmt.Sprintf("quest:open:%d", id))
	if err := db.DB.QueryRow(`SELECT status FROM quests WHERE id=?`, id).Scan(&status); err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != "active" {
		t.Fatalf("expected quest active after reopen, got %q", status)
	}
}
