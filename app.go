package main

import (
	"log"
	"os"
	"path/filepath"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
)

// initMedia resolves the media directory from env/DB path, creates it,
// and configures the handlers package. Returns the resolved mediaPath.
func initMedia(dbPath string) string {
	mediaPath := os.Getenv("MEDIA_PATH")
	if mediaPath == "" {
		basePath := filepath.Dir(dbPath)
		if basePath == "." {
			basePath = "."
		}
		mediaPath = filepath.Join(basePath, "media")
	}
	if err := os.MkdirAll(mediaPath, 0755); err != nil {
		log.Printf("Warning: could not create media directory: %v", err)
	}
	handlers.SetMediaPath(mediaPath)
	return mediaPath
}

// registerSchedulers wires session/DB stores and starts background schedulers.
// Returns a channel that controls the push reminder scheduler lifetime.
func registerSchedulers() chan struct{} {
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	middleware.TokenDB = db.DB
	middleware.StartCleanupTask()
	handlers.StartBackupScheduler()
	handlers.StartDBCleanupTask()
	pushStop := make(chan struct{})
	handlers.StartPushReminderScheduler(pushStop)
	return pushStop
}
