package telegram

import (
	"context"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestAsk_Unlinked(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "u1", "player")
	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 9999, args: []string{"hi"}}
	reply := runAskQuery(c, "hi")
	if !strings.Contains(reply.Text, "/start") {
		t.Fatalf("expected linking help containing /start, got %q", reply.Text)
	}
}

func TestAsk_LinkedAIDisabled(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "u1", "player")
	testutil.SeedCompendiumSpell(t, 9010, "Lightning Bolt")
	// link identity
	db.DB.Exec("INSERT INTO telegram_identities(user_id, telegram_user_id, telegram_username, dm_enabled) VALUES(?,?,?,?)", 1, 12345, "tester", 1)
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	c := &cmdContext{ctx: context.Background(), chatID: 200, tgUserID: 12345, args: []string{"Lightning"}}
	reply := runAskQuery(c, "Lightning")
	// degraded message expected, plus compendium match
	if !strings.Contains(reply.Text, "AI is not configured") {
		t.Fatalf("expected degraded message, got %q", reply.Text)
	}
	if !strings.Contains(reply.Text, "Lightning") {
		t.Fatalf("expected compendium match, got %q", reply.Text)
	}
}

func TestAsk_BoundGroupNonMember(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "player")
	testutil.SeedUser(t, 2, "intruder", "player")
	testutil.SeedCampaign(t, 10, "CampBound", "Party", 1)
	// link intruder (not a campaign member) to Telegram
	db.DB.Exec("INSERT INTO telegram_identities(user_id, telegram_user_id, telegram_username, dm_enabled) VALUES(?,?,?,?)", 2, 99999, "intruder", 1)
	// bind campaign to group chat (negative chat ID = group)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(?,?,?)", 10, -12345, 1); err != nil {
		t.Fatalf("bind group: %v", err)
	}
	// keep AI disabled to force degraded/direct-matches path; still exercises bound-group scope
	db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled','0')")
	testutil.SeedCompendiumSpell(t, 9020, "Magic Missile")
	c := &cmdContext{ctx: context.Background(), chatID: -12345, tgUserID: 99999, args: []string{"hi"}}
	reply := runAskQuery(c, "hi")
	if reply.Text == "" {
		t.Fatalf("expected non-empty reply, got empty")
	}
	if strings.Contains(reply.Text, "error") && strings.Contains(strings.ToLower(reply.Text), "not a campaign member") {
		t.Fatalf("unexpected membership error in bound group, got %q", reply.Text)
	}
	// degraded path should mention AI not configured / direct matches
	if !strings.Contains(reply.Text, "AI is not configured") && !strings.Contains(reply.Text, "AI is unavailable") && !strings.Contains(reply.Text, "Top matches") {
		t.Fatalf("expected degraded/direct-matches reply, got %q", reply.Text)
	}
}
