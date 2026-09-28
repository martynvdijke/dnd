package db

import (
	"context"
	"path/filepath"
	"testing"
)

// legacyCampaignFixture builds a database that looks like a pre-membership
// install: characters.campaign_id exists (added by ALTER TABLE, like
// safe_alters did) with an index, and joins are on that column. It returns the
// database path.
func legacyCampaignFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy.db")
	if err := Init(path); err != nil {
		t.Fatalf("init: %v", err)
	}

	if _, err := DB.Exec("INSERT INTO users(username, password) VALUES('dm', 'x')"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := DB.Exec("INSERT INTO campaigns(user_id, name) VALUES(1, 'Legacy Campaign')"); err != nil {
		t.Fatalf("insert campaign: %v", err)
	}
	if _, err := DB.Exec("INSERT INTO characters(user_id, name) VALUES(1, 'Mira')"); err != nil {
		t.Fatalf("insert character: %v", err)
	}
	if _, err := DB.Exec("ALTER TABLE characters ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL"); err != nil {
		t.Fatalf("add legacy campaign_id: %v", err)
	}
	if _, err := DB.Exec("CREATE INDEX idx_characters_campaign_name ON characters (campaign_id, name)"); err != nil {
		t.Fatalf("add legacy index: %v", err)
	}
	if _, err := DB.Exec("UPDATE characters SET campaign_id=1 WHERE id=1"); err != nil {
		t.Fatalf("link character: %v", err)
	}
	// Child row that must survive the column drop and the rebuild fallback.
	if _, err := DB.Exec("INSERT INTO character_currency(character_id, gp) VALUES(1, 42)"); err != nil {
		t.Fatalf("insert currency: %v", err)
	}
	// A real legacy install has no campaign_characters table yet.
	if _, err := DB.Exec("DROP TABLE campaign_characters"); err != nil {
		t.Fatalf("drop join table: %v", err)
	}
	Close()
	return path
}

func columnExists(t *testing.T, table, column string) bool {
	t.Helper()
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?", table, column).Scan(&count); err != nil {
		t.Fatalf("pragma_table_info(%s): %v", table, err)
	}
	return count > 0
}

func TestCharacterCampaignMembershipMigration(t *testing.T) {
	path := legacyCampaignFixture(t)

	// Re-open: Init runs the membership migration.
	if err := Init(path); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	t.Cleanup(func() { Close() })

	if columnExists(t, "characters", "campaign_id") {
		t.Fatal("characters.campaign_id still exists after migration")
	}
	var memberships int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM campaign_characters WHERE campaign_id=1 AND character_id=1`).Scan(&memberships); err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if memberships != 1 {
		t.Fatalf("expected 1 membership, got %d", memberships)
	}
	var legacyIndex int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_characters_campaign_name'`).Scan(&legacyIndex); err != nil {
		t.Fatalf("count legacy index: %v", err)
	}
	if legacyIndex != 0 {
		t.Fatal("legacy campaign index still exists")
	}
	var gp int
	if err := DB.QueryRow(`SELECT gp FROM character_currency WHERE character_id=1`).Scan(&gp); err != nil {
		t.Fatalf("read currency: %v", err)
	}
	if gp != 42 {
		t.Fatalf("currency row lost: gp=%d", gp)
	}

	// Idempotent: another startup does not duplicate memberships.
	if err := Init(path); err != nil {
		t.Fatalf("second re-init: %v", err)
	}
	if err := DB.QueryRow(`SELECT COUNT(*) FROM campaign_characters WHERE campaign_id=1 AND character_id=1`).Scan(&memberships); err != nil {
		t.Fatalf("recount memberships: %v", err)
	}
	if memberships != 1 {
		t.Fatalf("expected 1 membership after second start, got %d", memberships)
	}
}

// TestRebuildCharactersWithoutCampaignID exercises the table-rebuild fallback
// directly, since DROP COLUMN succeeds in normal operation.
func TestRebuildCharactersWithoutCampaignID(t *testing.T) {
	path := legacyCampaignFixture(t)

	if err := Init(path); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	t.Cleanup(func() { Close() })

	// Re-add the legacy shape without running the membership migration.
	if _, err := DB.Exec("ALTER TABLE characters ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL"); err != nil {
		t.Fatalf("re-add campaign_id: %v", err)
	}
	if _, err := DB.Exec("UPDATE characters SET campaign_id=1 WHERE id=1"); err != nil {
		t.Fatalf("relink character: %v", err)
	}

	ctx := context.Background()
	conn, err := DB.Conn(ctx)
	if err != nil {
		t.Fatalf("conn: %v", err)
	}
	defer conn.Close()
	if err := rebuildCharactersWithoutCampaignID(ctx, conn); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	if columnExists(t, "characters", "campaign_id") {
		t.Fatal("campaign_id still exists after rebuild")
	}
	var name string
	if err := DB.QueryRow(`SELECT name FROM characters WHERE id=1`).Scan(&name); err != nil {
		t.Fatalf("read character: %v", err)
	}
	if name != "Mira" {
		t.Fatalf("character row lost: %q", name)
	}
	var gp int
	if err := DB.QueryRow(`SELECT gp FROM character_currency WHERE character_id=1`).Scan(&gp); err != nil {
		t.Fatalf("read currency: %v", err)
	}
	if gp != 42 {
		t.Fatalf("currency row lost after rebuild: gp=%d", gp)
	}
	var violations int
	if err := DB.QueryRow("SELECT COUNT(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil {
		t.Fatalf("foreign key check: %v", err)
	}
	if violations != 0 {
		t.Fatalf("foreign key violations after rebuild: %d", violations)
	}
	// Hot-path index from safe_alters must have been recreated.
	var idx int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_characters_user_name_level'`).Scan(&idx); err != nil {
		t.Fatalf("count index: %v", err)
	}
	if idx != 1 {
		t.Fatal("idx_characters_user_name_level missing after rebuild")
	}
}
