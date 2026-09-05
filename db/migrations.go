package db

import (
	"villum/db/migrations"
	"villum/middleware"
)

// Migrate applies pending migrations.
func Migrate() error { return migrations.Migrate(DB) }

// ApplySafeAlters runs safe ALTER TABLE additions.
func ApplySafeAlters() error { return migrations.ApplySafeAlters(DB) }

// rebuildNPCItemLinksIfNeeded is kept for backward compat (unexported caller).
func rebuildNPCItemLinksIfNeeded() error { return migrations.RebuildNPCItemLinksIfNeeded(DB) }

// LogPageStats reports page stats.
func LogPageStats() {
	var pageCount, freelistCount, pageSize int
	if err := DB.QueryRow("PRAGMA page_count").Scan(&pageCount); err != nil {
		return
	}
	_ = DB.QueryRow("PRAGMA freelist_count").Scan(&freelistCount)
	_ = DB.QueryRow("PRAGMA page_size").Scan(&pageSize)
	middleware.LogInfo("db", "page stats", "page_count", pageCount, "freelist_count", freelistCount, "page_size", pageSize)
}
