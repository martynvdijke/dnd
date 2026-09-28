package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"villum/middleware"
)

// campaignCharactersDDL creates the campaign membership join table. It matches
// what ent generates for the CampaignCharacter schema so that running it before
// ent's schema migration leaves ent nothing to change. Any drift is reconciled
// by ent on the next startup.
const campaignCharactersDDL = `
CREATE TABLE IF NOT EXISTS campaign_characters (
	id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	campaign_id integer NOT NULL,
	character_id integer NOT NULL,
	CONSTRAINT campaign_characters_campaigns_character_links FOREIGN KEY (campaign_id) REFERENCES campaigns (id) ON DELETE NO ACTION,
	CONSTRAINT campaign_characters_characters_campaign_links FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS campaigncharacter_campaign_id_character_id ON campaign_characters (campaign_id, character_id);
CREATE INDEX IF NOT EXISTS campaigncharacter_character_id ON campaign_characters (character_id);`

// migrateCharacterCampaignMemberships backfills campaign_characters from the
// legacy characters.campaign_id column, then drops the column. The column only
// exists on databases created before campaign memberships became many-to-many;
// fresh databases never have it, so the whole migration is a no-op there.
//
// It runs after numbered migrations and BEFORE ent's schema migration: ent
// rebuilds tables whose columns it no longer knows about, which would drop
// characters.campaign_id (and its data) before a backfill could read it.
func migrateCharacterCampaignMemberships() error {
	ctx := context.Background()
	conn, err := DB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Close()

	var count int
	if err := conn.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pragma_table_info('characters') WHERE name='campaign_id'").Scan(&count); err != nil {
		return fmt.Errorf("inspect characters table: %w", err)
	}
	if count == 0 {
		return nil
	}

	if _, err := conn.ExecContext(ctx, campaignCharactersDDL); err != nil {
		return fmt.Errorf("create campaign_characters: %w", err)
	}

	// Backfill first, in its own statement, so the data move survives even if
	// dropping the column needs the rebuild fallback.
	if _, err := conn.ExecContext(ctx, `
		INSERT OR IGNORE INTO campaign_characters (campaign_id, character_id)
		SELECT campaign_id, id FROM characters
		WHERE campaign_id IS NOT NULL AND campaign_id != 0`); err != nil {
		return fmt.Errorf("backfill campaign memberships: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "DROP INDEX IF EXISTS idx_characters_campaign_name"); err != nil {
		return fmt.Errorf("drop legacy character campaign index: %w", err)
	}

	if _, err := conn.ExecContext(ctx, "ALTER TABLE characters DROP COLUMN campaign_id"); err == nil {
		middleware.LogInfo("migration", "dropped legacy characters.campaign_id column")
		return nil
	} else {
		middleware.LogWarn("migration", "ALTER TABLE DROP COLUMN failed; rebuilding characters table", "error", err)
	}

	if err := rebuildCharactersWithoutCampaignID(ctx, conn); err != nil {
		return fmt.Errorf("rebuild characters table: %w", err)
	}
	return nil
}

// rebuildCharactersWithoutCampaignID recreates the characters table without the
// campaign_id column following the SQLite table-rebuild procedure. It keeps
// legacy_alter_table on so child foreign keys continue to reference
// "characters" instead of the temporary table name, and turns foreign keys off
// for the duration. It must run in autocommit mode.
func rebuildCharactersWithoutCampaignID(ctx context.Context, conn *sql.Conn) error {
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys=OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA legacy_alter_table=ON"); err != nil {
		return fmt.Errorf("enable legacy alter table: %w", err)
	}
	defer func() {
		conn.ExecContext(ctx, "PRAGMA legacy_alter_table=OFF")
		conn.ExecContext(ctx, "PRAGMA foreign_keys=ON")
	}()

	cols, err := characterColumnNames(ctx, conn)
	if err != nil {
		return err
	}
	keep := make([]string, 0, len(cols))
	for _, c := range cols {
		if c != "campaign_id" {
			keep = append(keep, c)
		}
	}
	if len(keep) == 0 || len(keep) == len(cols) {
		return fmt.Errorf("unexpected characters columns: %v", cols)
	}

	var createSQL string
	if err := conn.QueryRowContext(ctx,
		"SELECT sql FROM sqlite_master WHERE type='table' AND name='characters'").Scan(&createSQL); err != nil {
		return fmt.Errorf("read characters DDL: %w", err)
	}
	newDDL, err := replaceCharactersTableName(removeColumnFromCreateTable(createSQL, "campaign_id"), "characters_new")
	if err != nil {
		return err
	}

	// Collect index DDL before the table is dropped. Indexes on campaign_id
	// disappear with the column.
	indexSQLs, err := characterIndexDDL(ctx, conn)
	if err != nil {
		return err
	}

	quoted := make([]string, len(keep))
	for i, c := range keep {
		quoted[i] = `"` + c + `"`
	}
	colList := strings.Join(quoted, ", ")

	steps := []string{
		`ALTER TABLE "characters" RENAME TO "characters_old"`,
		newDDL,
		fmt.Sprintf("INSERT INTO %q (%s) SELECT %s FROM %q", "characters_new", colList, colList, "characters_old"),
		`DROP TABLE "characters_old"`,
		`ALTER TABLE "characters_new" RENAME TO "characters"`,
	}
	for _, step := range steps {
		if _, err := conn.ExecContext(ctx, step); err != nil {
			return fmt.Errorf("step %q: %w", truncate(step, 80), err)
		}
	}
	for _, idx := range indexSQLs {
		if _, err := conn.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("recreate index %q: %w", truncate(idx, 80), err)
		}
	}

	var violations int
	if err := conn.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pragma_foreign_key_check('characters')").Scan(&violations); err != nil {
		return fmt.Errorf("foreign key check: %w", err)
	}
	if violations > 0 {
		return fmt.Errorf("rebuild left %d foreign key violations", violations)
	}
	middleware.LogInfo("migration", "rebuilt characters table without campaign_id")
	return nil
}

// characterColumnNames returns the column names of the characters table in
// declaration order.
func characterColumnNames(ctx context.Context, conn *sql.Conn) ([]string, error) {
	rows, err := conn.QueryContext(ctx, "SELECT name FROM pragma_table_info('characters') ORDER BY cid")
	if err != nil {
		return nil, fmt.Errorf("read characters columns: %w", err)
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan characters column: %w", err)
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

// characterIndexDDL returns CREATE INDEX statements for the characters table,
// skipping any index that references campaign_id.
func characterIndexDDL(ctx context.Context, conn *sql.Conn) ([]string, error) {
	rows, err := conn.QueryContext(ctx,
		`SELECT sql FROM sqlite_master WHERE type='index' AND tbl_name='characters' AND sql IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("read characters indexes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var stmt string
		if err := rows.Scan(&stmt); err != nil {
			return nil, fmt.Errorf("scan characters index: %w", err)
		}
		if strings.Contains(strings.ToLower(stmt), "campaign_id") {
			continue
		}
		out = append(out, stmt)
	}
	return out, rows.Err()
}

// removeColumnFromCreateTable strips a column definition from a CREATE TABLE
// statement, splitting the column list on commas at paren depth 1 so inline
// ALTER TABLE additions and multi-line definitions both work.
func removeColumnFromCreateTable(ddl, column string) string {
	open := strings.Index(ddl, "(")
	close := strings.LastIndex(ddl, ")")
	if open < 0 || close < open {
		return ddl
	}
	body := ddl[open+1 : close]
	var parts []string
	depth, start := 0, 0
	for i, r := range body {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, body[start:])

	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		isColumn := false
		for _, name := range []string{column, "`" + column + "`", `"` + column + `"`} {
			if trimmed == name || strings.HasPrefix(trimmed, name+" ") {
				isColumn = true
				break
			}
		}
		if !isColumn {
			kept = append(kept, part)
		}
	}
	return ddl[:open+1] + strings.Join(kept, ",") + ddl[close:]
}

// replaceCharactersTableName rewrites the table name in a CREATE TABLE
// statement, handling backtick, double-quote, and bare identifiers.
func replaceCharactersTableName(createSQL, newName string) (string, error) {
	for _, old := range []string{"`characters`", `"characters"`, "characters"} {
		prefix := "CREATE TABLE " + old
		if strings.Contains(createSQL, prefix) {
			return strings.Replace(createSQL, prefix, "CREATE TABLE "+newName, 1), nil
		}
	}
	return "", fmt.Errorf("unrecognized characters DDL: %s", truncate(createSQL, 120))
}
