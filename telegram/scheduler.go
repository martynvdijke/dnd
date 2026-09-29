package telegram

import (
	"time"
	"villum/db"
	"villum/middleware"
)

func StartTelegramAutoPostScheduler(stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case now := <-ticker.C:
				if GetBotToken() == "" {
					continue
				}
				scanAutoPost(now)
				retryFailed()
			}
		}
	}()
}
func scanAutoPost(now time.Time) {
	grace := GetGraceMinutes()
	cutoff := now.UTC().Add(-time.Duration(grace) * time.Minute).Format("2006-01-02 15:04:05")
	rows, err := db.DB.Query(`
		SELECT r.id FROM campaign_recaps r
		JOIN campaign_telegram_settings s ON s.campaign_id=r.campaign_id
		WHERE r.is_sent=0 AND r.session_end_date IS NOT NULL AND r.session_end_date != ''
		  AND s.auto_post_enabled=1 AND s.is_enabled=1
		  AND datetime(r.session_end_date) <= datetime(?)
		  AND NOT EXISTS (SELECT 1 FROM telegram_deliveries d WHERE d.recap_id=r.id)
	`, cutoff)
	if err != nil {
		middleware.LogWarn("telegram", "auto-post scan failed", "error", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		_ = rows.Scan(&id)
		DeliverRecap(id, "auto")
	}
}
