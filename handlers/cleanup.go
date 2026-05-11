package handlers

import (
	"log"
	"time"

	"villum/db"
)

func StartDBCleanupTask() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)

			// Delete old dice roll history (older than 90 days)
			result, err := db.DB.Exec("DELETE FROM dice_rolls WHERE timestamp < datetime('now', '-90 days')")
			if err != nil {
				log.Printf("Cleanup: failed to delete old dice rolls: %v", err)
			} else if n, _ := result.RowsAffected(); n > 0 {
				log.Printf("Cleanup: removed %d old dice rolls", n)
			}

			// Delete old rest logs (older than 180 days)
			result, err = db.DB.Exec("DELETE FROM rest_log WHERE timestamp < datetime('now', '-180 days')")
			if err != nil {
				log.Printf("Cleanup: failed to delete old rest logs: %v", err)
			} else if n, _ := result.RowsAffected(); n > 0 {
				log.Printf("Cleanup: removed %d old rest logs", n)
			}

			// Delete old NPC interactions (older than 365 days)
			result, err = db.DB.Exec("DELETE FROM character_npcs WHERE last_interacted != '' AND last_interacted < datetime('now', '-365 days') AND interaction_count = 0")
			if err != nil {
				log.Printf("Cleanup: failed to delete stale NPC links: %v", err)
			} else if n, _ := result.RowsAffected(); n > 0 {
				log.Printf("Cleanup: removed %d stale NPC links", n)
			}

			// Manually trigger WAL checkpoint to keep DB size small
			db.DB.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
		}
	}()
}
