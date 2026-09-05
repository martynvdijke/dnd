package migrations

import (
	"database/sql"
	"fmt"
	"villum/middleware"
)

// Migration is an ordered DB migration.
type Migration struct {
	Version int
	SQL     string
}

// Registry is the ordered list of all migrations.
var Registry = []Migration{
	{Version: 1, SQL: migration001SQL},
	{Version: 2, SQL: migration002SQL},
	{Version: 3, SQL: migration003SQL},
	{Version: 4, SQL: migration004SQL},
	{Version: 5, SQL: migration005SQL},
	{Version: 6, SQL: migration006SQL},
	{Version: 7, SQL: migration007SQL},
	{Version: 8, SQL: migration008SQL},
	{Version: 9, SQL: migration009SQL},
	{Version: 10, SQL: migration010SQL},
	{Version: 11, SQL: migration011SQL},
	{Version: 12, SQL: migration012SQL},
	{Version: 13, SQL: migration013SQL},
	{Version: 14, SQL: migration014SQL},
	{Version: 15, SQL: migration015SQL},
	{Version: 16, SQL: migration016SQL},
	{Version: 17, SQL: migration017SQL},
	{Version: 18, SQL: migration018SQL},
	{Version: 19, SQL: migration019SQL},
	{Version: 20, SQL: migration020SQL},
	{Version: 21, SQL: migration021SQL},
	{Version: 22, SQL: migration022SQL},
	{Version: 23, SQL: migration023SQL},
	{Version: 24, SQL: migration024SQL},
	{Version: 25, SQL: migration025SQL},
	{Version: 26, SQL: migration026SQL},
	{Version: 27, SQL: migration027SQL},
	{Version: 28, SQL: migration028SQL},
	{Version: 29, SQL: migration029SQL},
	{Version: 30, SQL: migration030SQL},
	{Version: 31, SQL: migration031SQL},
	{Version: 32, SQL: migration032SQL},
	{Version: 33, SQL: migration033SQL},
	{Version: 34, SQL: migration034SQL},
	{Version: 35, SQL: migration035SQL},
	{Version: 36, SQL: migration036SQL},
	{Version: 37, SQL: migration037SQL},
	{Version: 38, SQL: migration038SQL},
	{Version: 39, SQL: migration039SQL},
	{Version: 40, SQL: migration040SQL},
	{Version: 41, SQL: migration041SQL},
	{Version: 42, SQL: migration042SQL},
	{Version: 43, SQL: migration043SQL},
	{Version: 44, SQL: migration044SQL},
	{Version: 45, SQL: migration045SQL},
	{Version: 46, SQL: migration046SQL},
	{Version: 47, SQL: migration047SQL},
	{Version: 48, SQL: migration048SQL},
	{Version: 49, SQL: migration049SQL},
	{Version: 50, SQL: migration050SQL},
	{Version: 51, SQL: migration051SQL},
}

// Migrate applies pending migrations to the given DB.
func Migrate(db *sql.DB) error {
	var current int
	err := db.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_version").Scan(&current)
	if err != nil {
		current = 0
	}
	migrated := false
	for _, m := range Registry {
		if m.Version > current {
			migrated = true
			middleware.LogInfo("migration", "Running migration", "version", m.Version)
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("begin migration v%d: %w", m.Version, err)
			}
			if _, err := tx.Exec(m.SQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration v%d: %w", m.Version, err)
			}
			if _, err := tx.Exec("INSERT OR REPLACE INTO schema_version(version) VALUES(?)", m.Version); err != nil {
				tx.Rollback()
				return fmt.Errorf("record migration v%d: %w", m.Version, err)
			}
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit migration v%d: %w", m.Version, err)
			}
		}
	}
	if migrated {
		if _, err := db.Exec("VACUUM"); err != nil {
			return fmt.Errorf("vacuum after migrations: %w", err)
		}
		// LogPageStats equivalent
		var pageCount, freelistCount, pageSize int
		if err := db.QueryRow("PRAGMA page_count").Scan(&pageCount); err == nil {
			_ = db.QueryRow("PRAGMA freelist_count").Scan(&freelistCount)
			_ = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
			middleware.LogInfo("db", "page stats", "page_count", pageCount, "freelist_count", freelistCount, "page_size", pageSize)
		}
	}
	return nil
}
