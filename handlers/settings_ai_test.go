package handlers

import (
	"context"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
)

// TestAIEnabledDelegation pins the behaviour of the aiEnabled wrapper that now
// delegates to ai.IsEnabled: a missing setting defaults to enabled, and the
// stored value decides afterwards.
func TestAIEnabledDelegation(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	ctx := context.Background()

	if !aiEnabled(ctx) {
		t.Fatal("missing ai_enabled setting must default to enabled")
	}

	set := func(value string) {
		t.Helper()
		if _, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('ai_enabled', ?)", value); err != nil {
			t.Fatalf("seed ai_enabled=%s: %v", value, err)
		}
	}

	set("0")
	if aiEnabled(ctx) {
		t.Fatal("ai_enabled=0 must report disabled")
	}

	set("1")
	if !aiEnabled(ctx) {
		t.Fatal("ai_enabled=1 must report enabled")
	}
}
