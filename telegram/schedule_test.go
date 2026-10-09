package telegram

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

func TestScheduleUsage(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: -1001, tgUserID: 1, args: []string{"onlyone"}}
	reply := runSchedule(c)
	if !strings.Contains(strings.ToLower(reply.Text), "usage") {
		t.Fatalf("expected usage, got %q", reply.Text)
	}
}

func TestScheduleWithBoundCampaign(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 7, "Test Campaign", "Party", 1)
	chatID := int64(-100123)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 7, ChatID: &chatID, IsEnabled: true}); err != nil {
		t.Fatalf("bind: %v", err)
	}
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		// handle poll and follow-up message
		w.Write(jsonOK(map[string]any{"message_id": 42}))
	})
	c := &cmdContext{ctx: context.Background(), chatID: chatID, tgUserID: 1, args: []string{"Friday", "|", "Saturday"}}
	reply := runSchedule(c)
	if reply.Text != "" {
		// runSchedule returns botReply{} on success, but may return error text
		t.Fatalf("expected empty reply on success, got %q", reply.Text)
	}
	var n int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM telegram_session_polls WHERE chat_id=?", chatID).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 poll row, got %d", n)
	}
}

func TestHandleScheduleCallback(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: -1001, tgUserID: 1}
	r := handleScheduleCallback(c, "sched:unknown")
	if r.Text != "Unknown action." {
		t.Fatalf("expected unknown, got %q", r.Text)
	}
}
