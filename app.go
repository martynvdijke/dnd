package main

import (
	"log"
	"os"
	"path/filepath"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
	"villum/telegram"
)

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

func registerSchedulers() chan struct{} {
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	middleware.TokenDB = db.DB
	middleware.StartCleanupTask()
	handlers.StartBackupScheduler()
	handlers.StartDBCleanupTask()
	pushStop := make(chan struct{})
	handlers.StartPushReminderScheduler(pushStop)
	telegram.EnsureWebhookState()
	telegramStop := make(chan struct{})
	telegram.StartBot(telegramStop)
	telegram.SetCharacterCreator(handlers.BotCharacterCreator)
	telegram.SetAuthLinkSender(handlers.SendTelegramAuthLinkEmail)
	telegram.StartTelegramAutoPostScheduler(telegramStop)
	_ = telegramStop
	return pushStop
}
