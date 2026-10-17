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
		SELECT r.id, r.campaign_id FROM campaign_recaps r
		WHERE r.is_sent=0 AND r.session_end_date IS NOT NULL AND r.session_end_date != ''
		  AND datetime(r.session_end_date) <= datetime(?)
		  AND NOT EXISTS (SELECT 1 FROM telegram_deliveries d WHERE d.recap_id=r.id)
	`, cutoff)
	if err != nil {
		middleware.LogWarn("telegram", "auto-post scan failed", "error", err)
		return
	}
	type pendingRecap struct {
		id         int64
		campaignID int64
	}
	var pending []pendingRecap
	func() {
		defer rows.Close()
		for rows.Next() {
			var p pendingRecap
			if err := rows.Scan(&p.id, &p.campaignID); err != nil {
				continue
			}
			pending = append(pending, p)
		}
	}()
	for _, p := range pending {
		settings, found, err := GetCampaignTelegramSettings(p.campaignID)
		if err != nil || !found || !settings.IsEnabled || !settings.AutoPostEnabled {
			continue
		}
		DeliverRecap(p.id, "auto")
	}
}
