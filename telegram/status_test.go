package telegram

import (
	"context"
	"strings"
	"testing"

	"villum/handlers/testutil"
)

func TestRunStatusReportsBuildVersion(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if err := UpsertIdentity(1, 100, 100, "admin"); err != nil {
		t.Fatalf("upsert identity: %v", err)
	}

	prev := buildVersion
	SetBuildVersion("v9.9.9-test")
	defer SetBuildVersion(prev)

	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	reply := runStatus(c)
	if !strings.Contains(reply.Text, "Build: v9.9.9-test") {
		t.Fatalf("status reply missing build version: %q", reply.Text)
	}
	if !strings.Contains(reply.Text, "Linked:") {
		t.Fatalf("status reply missing linked line: %q", reply.Text)
	}
}

func TestRunStatusOmitsBuildWhenUnset(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if err := UpsertIdentity(1, 100, 100, "admin"); err != nil {
		t.Fatalf("upsert identity: %v", err)
	}

	prev := buildVersion
	SetBuildVersion("")
	defer SetBuildVersion(prev)

	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	reply := runStatus(c)
	if strings.Contains(reply.Text, "Build:") {
		t.Fatalf("status reply unexpectedly contains build version: %q", reply.Text)
	}
}
