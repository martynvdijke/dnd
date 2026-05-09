package handlers

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vellum/db"
)

func CreateBackup() (string, error) {
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	// Get the database path from the DB connection
	dbPath := getDBPath()
	if dbPath == "" {
		return "", fmt.Errorf("could not determine database path")
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("vellum_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Perform VACUUM INTO for a consistent snapshot (escape single quotes in path)
	safePath := strings.ReplaceAll(backupPath, "'", "''")
	_, err := db.DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", safePath))
	if err != nil {
		// Fallback: copy the file directly
		src, err := os.Open(dbPath)
		if err != nil {
			return "", fmt.Errorf("open source db: %w", err)
		}
		defer src.Close()

		dst, err := os.Create(backupPath)
		if err != nil {
			return "", fmt.Errorf("create backup: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return "", fmt.Errorf("copy db: %w", err)
		}
	}

	return backupPath, nil
}

var dbPathCache string

func getDBPath() string {
	if dbPathCache != "" {
		return dbPathCache
	}

	// Try to extract the database path from the PRAGMA
	row := db.DB.QueryRow("PRAGMA database_list")
	var seq int
	var name, file string
	if err := row.Scan(&seq, &name, &file); err == nil && file != "" {
		dbPathCache = file
	}

	return dbPathCache
}

// SetDBPath allows main.go to set the db path for backup
func SetDBPath(path string) {
	dbPathCache = path
}

func StartBackupScheduler() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			checkAndBackup()
		}
	}()
}

func checkAndBackup() {
	var enabled bool
	var interval int
	var lastBackup string

	err := db.DB.QueryRow("SELECT enabled, interval_hours, last_backup FROM backup_settings WHERE id=1").Scan(&enabled, &interval, &lastBackup)
	if err != nil {
		return
	}
	if !enabled {
		return
	}

	shouldBackup := false
	if lastBackup == "" {
		shouldBackup = true
	} else {
		lastTime, err := time.Parse("2006-01-02 15:04:05", lastBackup)
		if err != nil {
			shouldBackup = true
		} else {
			hoursSince := time.Since(lastTime).Hours()
			if hoursSince >= float64(interval) {
				shouldBackup = true
			}
		}
	}

	if shouldBackup {
		path, err := CreateBackup()
		if err != nil {
			log.Printf("Auto backup failed: %v", err)
			return
		}
		now := time.Now().Format("2006-01-02 15:04:05")
		db.DB.Exec("UPDATE backup_settings SET last_backup=? WHERE id=1", now)
		log.Printf("Auto backup created: %s", path)
	}
}
