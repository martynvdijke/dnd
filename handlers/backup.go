package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"villum/db"
	"villum/middleware"
)

func getBackupDir() string {
	dbPath := getDBPath()
	if dbPath == "" {
		return "backups"
	}
	base := filepath.Dir(dbPath)
	if base == "." {
		return "backups"
	}
	return filepath.Join(base, "backups")
}

func CreateBackup() (string, error) {
	backupDir := getBackupDir()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	// Get the database path from the DB connection
	dbPath := getDBPath()
	if dbPath == "" {
		return "", fmt.Errorf("could not determine database path")
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("villum_%s.db", timestamp)
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
	var intervalDays int
	var keepCount int
	var lastBackup string

	err := db.DB.QueryRow("SELECT enabled, COALESCE(interval_days, 7), COALESCE(keep_count, 7), last_backup FROM backup_settings WHERE id=1").Scan(&enabled, &intervalDays, &keepCount, &lastBackup)
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
			if hoursSince >= float64(intervalDays*24) {
				shouldBackup = true
			}
		}
	}

	if shouldBackup {
		path, err := CreateBackup()
		if err != nil {
			middleware.LogError("backup", "auto backup failed", "error", err)
			return
		}
		now := time.Now().Format("2006-01-02 15:04:05")
		db.DB.Exec("UPDATE backup_settings SET last_backup=? WHERE id=1", now)
		middleware.LogInfo("backup", "auto backup created", "path", path)
		PruneBackups()
	}
}

func PruneBackups() {
	var keepCount int
	err := db.DB.QueryRow("SELECT COALESCE(keep_count, 7) FROM backup_settings WHERE id=1").Scan(&keepCount)
	if err != nil || keepCount < 1 {
		keepCount = 7
	}

	backupDir := getBackupDir()
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	type backupFile struct {
		name    string
		modTime time.Time
	}

	var backups []backupFile
	for _, e := range entries {
		if info, err := e.Info(); err == nil && !info.IsDir() && strings.HasPrefix(e.Name(), "villum_") && strings.HasSuffix(e.Name(), ".db") {
			backups = append(backups, backupFile{name: e.Name(), modTime: info.ModTime()})
		}
	}

	if len(backups) <= keepCount {
		return
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})

	removed := 0
	for _, b := range backups[keepCount:] {
		p := filepath.Join(backupDir, b.name)
		if err := os.Remove(p); err == nil {
			removed++
		}
	}
	if removed > 0 {
		middleware.LogInfo("backup", "pruned old backups", "removed", removed, "keep_count", keepCount)
	}
}
