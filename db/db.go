package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"villum/ent"
)

var DB *sql.DB
var Client *ent.Client

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_fk_on=1&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(4)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	// Enable WAL and foreign keys
	if _, err := DB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enable wal: %w", err)
	}
	if _, err := DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enable fk: %w", err)
	}

	// Initialize ent client
	drv := entsql.OpenDB(dialect.SQLite, DB)
	Client = ent.NewClient(ent.Driver(drv))

	return Migrate()
}

func Close() {
	if Client != nil {
		Client.Close()
	}
	if DB != nil {
		DB.Close()
	}
}
