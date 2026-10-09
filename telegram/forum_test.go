package telegram

import (
	"context"
	"strings"
	"testing"

	"villum/handlers/testutil"
)

func TestForumTopicsEmpty(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: -1001}
	reply := runTopics(c)
	if !strings.Contains(reply.Text, "No forum topics") {
		t.Fatalf("expected friendly empty, got %q", reply.Text)
	}
}

func TestForumTopicUsage(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	c := &cmdContext{ctx: context.Background(), chatID: -1001, args: []string{}}
	reply := runTopic(c)
	if !strings.Contains(strings.ToLower(reply.Text), "usage") {
		t.Fatalf("expected usage, got %q", reply.Text)
	}
}
