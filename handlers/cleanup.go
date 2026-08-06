package handlers

import (
	"time"

	"villum/db"
	"villum/middleware"
)

func StartDBCleanupTask() {
	go func() {
		for {
			time.Sleep(4 * time.Hour)

			// Delete old dice roll history (older than 90 days)
			result, err := db.DB.Exec("DELETE FROM dice_rolls WHERE timestamp < datetime('now', '-90 days')")
			if err != nil {
				middleware.LogError("cleanup", "failed to delete old dice rolls", "error", err)
			} else if n, _ := result.RowsAffected(); n > 0 {
				middleware.LogInfo("cleanup", "removed old dice rolls", "count", n)
			}

			// Delete old rest logs (older than 180 days)
			result, err = db.DB.Exec("DELETE FROM rest_log WHERE timestamp < datetime('now', '-180 days')")
			if err != nil {
				middleware.LogError("cleanup", "failed to delete old rest logs", "error", err)
			} else if n, _ := result.RowsAffected(); n > 0 {
				middleware.LogInfo("cleanup", "removed old rest logs", "count", n)
			}

			// Delete old NPC interactions (older than 365 days)
			result, err = db.DB.Exec("DELETE FROM character_npcs WHERE last_interacted != '' AND last_interacted < datetime('now', '-365 days') AND interaction_count = 0")
			if err != nil {
				middleware.LogError("cleanup", "failed to delete stale NPC links", "error", err)
			} else if n, _ := result.RowsAffected(); n > 0 {
				middleware.LogInfo("cleanup", "removed stale NPC links", "count", n)
			}

			// Manually trigger WAL checkpoint to keep DB size small
			var busy, logPages, checkpointed int
			if err := db.DB.QueryRow("PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &logPages, &checkpointed); err != nil {
				middleware.LogWarn("cleanup", "wal checkpoint failed", "error", err)
			} else {
				middleware.LogInfo("cleanup", "wal checkpoint", "busy", busy, "log_pages", logPages, "checkpointed", checkpointed)
			}
		}
	}()
}
