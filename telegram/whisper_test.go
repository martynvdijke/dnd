package telegram

import (
	"context"
	"testing"
)

func TestRunWhisperDM(t *testing.T) {
	setupTelegramDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: 123, tgUserID: 999, args: []string{"1d20"}}
	reply := runWhisper(c)
	if reply.Text == "" {
		t.Fatalf("expected refusal text in DM")
	}
	want := "Whispers only work in a group chat."
	if reply.Text != want {
		t.Fatalf("expected %q got %q", want, reply.Text)
	}
}

func TestRunWhisperGroup(t *testing.T) {
	setupTelegramDB(t)
	var sent []string
	recordingMock(t, &sent)
	c := &cmdContext{ctx: context.Background(), chatID: -100123, tgUserID: 999, args: []string{"1d20"}}
	reply := runWhisper(c)
	if reply.Text != "" {
		t.Fatalf("expected empty Text on success, got %q", reply.Text)
	}
}
