package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_fk_on=1")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(1)

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

	return Migrate()
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
