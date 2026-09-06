package db

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"villum/db/migrations"
)

func TestRegistryLength(t *testing.T) {
	files, _ := filepath.Glob("migrations/*.go")
	count := 0
	for _, f := range files {
		base := filepath.Base(f)
		if base == "registry.go" || base == "safe_alters.go" || base == "npc_item_links.go" {
			continue
		}
		count++
	}
	if len(migrations.Registry) != count {
		t.Fatalf("registry len %d != file count %d", len(migrations.Registry), count)
	}
}

func TestMigrateHash(t *testing.T) {
	if err := Init(":memory:"); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer Close()
	rows, err := DB.Query("SELECT sql FROM sqlite_master WHERE sql NOT NULL ORDER BY type, name")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	var parts []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		parts = append(parts, strings.TrimSpace(s))
	}
	sort.Strings(parts)
	h := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	got := fmt.Sprintf("%x", h)
	want := "ad9c09161968b26ba16f59ba96f340174a8bad3c013e14a2983ee08a6e8617bb"
	if got != want {
		t.Fatalf("sqlite_master hash mismatch got %s want %s", got, want)
	}
}
