package telegram

import (
	"context"
	"strings"
	"testing"

	"villum/handlers/testutil"
)

func TestAnnouncePostsToBoundChat(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	groupChat := int64(-100111)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 10, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	if err := UpsertIdentity(1, 111, 111, "dm"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: groupChat, tgUserID: 111, args: []string{"Session", "moved", "to", "Thursday", "<b>sharp</b>"}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "posted") {
		t.Fatalf("expected a posted confirmation, got %q", reply.Text)
	}
	if len(sent) == 0 {
		t.Fatalf("expected the announcement to be sent")
	}
	if !strings.Contains(sent[0], "Session moved to Thursday") {
		t.Fatalf("announcement text missing, got %q", sent[0])
	}
	if strings.Contains(sent[0], "<b>sharp</b>") || !strings.Contains(sent[0], "&lt;b&gt;sharp&lt;/b&gt;") {
		t.Fatalf("announcement HTML was not escaped, got %q", sent[0])
	}
}

func TestAnnounceFromDirectMessage(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	groupChat := int64(-100112)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 10, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	if err := UpsertIdentity(1, 111, 111, "dm"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: []string{"Gather", "at", "dawn"}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "posted") {
		t.Fatalf("expected a posted confirmation, got %q", reply.Text)
	}
	if len(sent) == 0 || !strings.Contains(sent[0], "Gather at dawn") {
		t.Fatalf("expected announcement to the bound chat, got %v", sent)
	}
}

func TestAnnounceEmptyRejected(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	groupChat := int64(-100113)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 10, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	if err := UpsertIdentity(1, 111, 111, "dm"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: groupChat, tgUserID: 111, args: []string{"  "}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "Usage:") {
		t.Fatalf("expected usage reply, got %q", reply.Text)
	}
	if len(sent) != 0 {
		t.Fatalf("empty announcement must not be sent, got %v", sent)
	}
}

func TestAnnounceNonDMRejected(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	testutil.SeedCampaignMember(t, 10, 2, "player")
	groupChat := int64(-100114)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 10, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	if err := UpsertIdentity(2, 222, 222, "player"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: groupChat, tgUserID: 222, args: []string{"hello"}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "Only the campaign DM") {
		t.Fatalf("expected DM-only refusal, got %q", reply.Text)
	}
	if len(sent) != 0 {
		t.Fatalf("non-DM announcement must not be sent, got %v", sent)
	}
}

func TestAnnounceAdminAllowed(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedUser(t, 9, "admin", "admin")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	groupChat := int64(-100115)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 10, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	if err := UpsertIdentity(9, 999, 999, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: groupChat, tgUserID: 999, args: []string{"Table", "cancelled"}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "posted") {
		t.Fatalf("expected admin announcement to post, got %q", reply.Text)
	}
	if len(sent) == 0 {
		t.Fatalf("admin announcement did not send")
	}
}

func TestAnnounceUnboundChatRejected(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	if err := UpsertIdentity(1, 111, 111, "dm"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	recordingMock(t, &[]string{})

	c := &cmdContext{ctx: context.Background(), chatID: -100116, tgUserID: 111, args: []string{"hello"}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "not connected to a campaign") {
		t.Fatalf("expected binding instructions, got %q", reply.Text)
	}
}

func TestAnnounceDisabledRejected(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	groupChat := int64(-100117)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 10, ChatID: &groupChat, IsEnabled: false}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	if err := UpsertIdentity(1, 111, 111, "dm"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	c := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: []string{"hello"}}
	reply := runAnnounce(c)
	if !strings.Contains(reply.Text, "disabled") {
		t.Fatalf("expected disabled refusal, got %q", reply.Text)
	}
	if len(sent) != 0 {
		t.Fatalf("disabled campaign must not receive an announcement, got %v", sent)
	}
}

func TestAnnounceUnboundCampaignChatRejected(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 10, "Sunken Crown", "The Party", 1)
	if err := UpsertIdentity(1, 111, 111, "dm"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	ctx := &cmdContext{ctx: context.Background(), chatID: 111, tgUserID: 111, args: []string{"hello"}}
	reply := runAnnounce(ctx)
	if !strings.Contains(reply.Text, "No Telegram chat is bound") {
		t.Fatalf("expected no-bound-chat refusal, got %q", reply.Text)
	}
	if len(sent) != 0 {
		t.Fatalf("unbound campaign must not receive an announcement, got %v", sent)
	}
}
