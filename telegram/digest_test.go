package telegram

import (
	"context"
	"strings"
	"testing"
	"time"

	"villum/db"
	"villum/handlers/testutil"
)

func seedDigestMember(t *testing.T, userID, campaignID int64, tgUserID int64) {
	t.Helper()
	testutil.SeedUser(t, userID, "player", "user")
	testutil.SeedCampaign(t, campaignID, "Sunken Crown", "The Party", userID)
	if err := UpsertIdentity(userID, tgUserID, tgUserID, "player"); err != nil {
		t.Fatalf("upsert identity: %v", err)
	}
}

func TestDigestSubscribe(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 111)
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111}
	reply := runDigest(c)
	if !strings.Contains(reply.Text, "enabled") {
		t.Fatalf("expected subscription confirmation, got %q", reply.Text)
	}
	found, lastSent, err := GetDigestSubscription(1, 10)
	if err != nil || !found {
		t.Fatalf("expected subscription: found=%v err=%v", found, err)
	}
	if lastSent != "" {
		t.Fatalf("fresh subscription must have no last_sent_at, got %q", lastSent)
	}
	var frequency, createdAt string
	var lastSentCol any
	if err := db.DB.QueryRow(`SELECT frequency, created_at, last_sent_at FROM telegram_digest_subscriptions
		WHERE user_id = 1 AND campaign_id = 10`).Scan(&frequency, &createdAt, &lastSentCol); err != nil {
		t.Fatalf("read subscription: %v", err)
	}
	if frequency != "weekly" {
		t.Fatalf("expected weekly frequency, got %q", frequency)
	}
	if createdAt == "" {
		t.Fatalf("expected created_at to be set")
	}
	if lastSentCol != nil {
		t.Fatalf("expected last_sent_at NULL, got %v", lastSentCol)
	}
}

func TestDigestUnsubscribe(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 111)
	if err := UpsertDigestSubscription(1, 10, "weekly"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	recordingMock(t, &[]string{})

	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: []string{"off"}}
	reply := runDigest(c)
	if !strings.Contains(reply.Text, "disabled") {
		t.Fatalf("expected unsubscribe confirmation, got %q", reply.Text)
	}
	if found, _, _ := GetDigestSubscription(1, 10); found {
		t.Fatalf("expected subscription to be removed")
	}
}

func TestDigestUnsubscribeWhenNotSubscribed(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 111)
	recordingMock(t, &[]string{})

	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: []string{"off"}}
	reply := runDigest(c)
	if !strings.Contains(reply.Text, "not subscribed") {
		t.Fatalf("expected a not-subscribed reply, got %q", reply.Text)
	}
}

func TestDigestMembershipRequired(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "player", "user")
	testutil.SeedUser(t, 2, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 2)
	if err := UpsertIdentity(1, 111, 111, "player"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	recordingMock(t, &[]string{})

	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: []string{"10"}}
	reply := runDigest(c)
	if !strings.Contains(reply.Text, "Campaign not found") {
		t.Fatalf("expected membership refusal, got %q", reply.Text)
	}
	if found, _, _ := GetDigestSubscription(1, 10); found {
		t.Fatalf("non-member subscription must not be stored")
	}
}

func TestDigestDueSend(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 222)
	if _, err := db.DB.Exec(`INSERT INTO campaign_recaps(campaign_id, title, content, created_at) VALUES(10, 'The Sunken Crown', 'body', datetime('now'))`); err != nil {
		t.Fatalf("seed recap: %v", err)
	}
	if err := UpsertDigestSubscription(1, 10, "weekly"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	now := time.Now().UTC()
	scanDigests(now)
	if len(sent) == 0 {
		t.Fatalf("expected a due digest to be sent")
	}
	if !strings.Contains(sent[0], "Weekly digest") || !strings.Contains(sent[0], "The Sunken Crown") {
		t.Fatalf("unexpected digest text: %q", sent[0])
	}
	_, lastSent, _ := GetDigestSubscription(1, 10)
	if lastSent == "" {
		t.Fatalf("expected last_sent_at to be recorded after sending")
	}

	sent = nil
	scanDigests(now)
	if len(sent) != 0 {
		t.Fatalf("a second scan in the same week must not resend, got %v", sent)
	}
}

func TestDigestNotYetDueSkipped(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 222)
	if _, err := db.DB.Exec(`INSERT INTO campaign_recaps(campaign_id, title, content) VALUES(10, 'The Sunken Crown', 'body')`); err != nil {
		t.Fatalf("seed recap: %v", err)
	}
	if err := UpsertDigestSubscription(1, 10, "weekly"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if _, err := db.DB.Exec(`UPDATE telegram_digest_subscriptions SET last_sent_at = datetime('now', '-1 day')
		WHERE user_id = 1 AND campaign_id = 10`); err != nil {
		t.Fatalf("set last sent: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	scanDigests(time.Now().UTC())
	if len(sent) != 0 {
		t.Fatalf("a digest sent less than a week ago must be skipped, got %v", sent)
	}
}

func TestDigestDueAgainAfterAWeek(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 222)
	if _, err := db.DB.Exec(`INSERT INTO campaign_recaps(campaign_id, title, content) VALUES(10, 'The Sunken Crown', 'body')`); err != nil {
		t.Fatalf("seed recap: %v", err)
	}
	if err := UpsertDigestSubscription(1, 10, "weekly"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if _, err := db.DB.Exec(`UPDATE telegram_digest_subscriptions SET last_sent_at = datetime('now', '-8 days')
		WHERE user_id = 1 AND campaign_id = 10`); err != nil {
		t.Fatalf("set last sent: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	scanDigests(time.Now().UTC())
	if len(sent) == 0 {
		t.Fatalf("expected a digest eight days after the last one")
	}
}

func TestDigestScanSkipsNonMember(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "player", "user")
	testutil.SeedUser(t, 2, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 2)
	if err := UpsertIdentity(1, 111, 111, "player"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// A stale subscription row for a user who is not a member of the campaign.
	if _, err := db.DB.Exec(`INSERT INTO telegram_digest_subscriptions(user_id, campaign_id, frequency, created_at)
		VALUES(1, 10, 'weekly', datetime('now'))`); err != nil {
		t.Fatalf("insert subscription: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	scanDigests(time.Now().UTC())
	if len(sent) != 0 {
		t.Fatalf("a non-member must not receive a digest, got %v", sent)
	}
	_, lastSent, _ := GetDigestSubscription(1, 10)
	if lastSent != "" {
		t.Fatalf("skipped subscription must not be claimed, got %q", lastSent)
	}
}

func TestDigestScanSkipsZeroChat(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	seedDigestMember(t, 1, 10, 0)
	if err := UpsertDigestSubscription(1, 10, "weekly"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if _, err := db.DB.Exec(`UPDATE telegram_identities SET telegram_chat_id = NULL WHERE user_id = 1`); err != nil {
		t.Fatalf("clear chat: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	scanDigests(time.Now().UTC())
	if len(sent) != 0 {
		t.Fatalf("unlinked chat must not receive a digest, got %v", sent)
	}
}
