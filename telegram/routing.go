package telegram

import (
	"fmt"
	"strings"
	"villum/db"
	"villum/middleware"
)

func DeliverRecap(recapID int64, kind string) {
	if kind != "manual" && kind != "auto" {
		kind = "manual"
	}
	go func() {
		var campaignID int64
		var title, content string
		err := db.DB.QueryRow("SELECT campaign_id, title, content FROM campaign_recaps WHERE id=?", recapID).Scan(&campaignID, &title, &content)
		if err != nil {
			return
		}
		if GetBotToken() == "" {
			return
		}
		var targets []struct {
			chatID int64
			typ    string
		}
		chat, found, err := GetCampaignTelegramSettings(campaignID)
		if err == nil && found && chat.ChatID != nil && chat.IsEnabled {
			targets = append(targets, struct {
				chatID int64
				typ    string
			}{*chat.ChatID, "chat"})
		}
		rows, _ := db.DB.Query(`SELECT telegram_chat_id FROM telegram_identities ti
			JOIN campaign_members cm ON cm.user_id=ti.user_id
			WHERE cm.campaign_id=? AND ti.dm_enabled=1 AND ti.telegram_chat_id IS NOT NULL
			UNION
			SELECT telegram_chat_id FROM telegram_identities ti
			JOIN campaigns c ON c.user_id=ti.user_id
			WHERE c.id=? AND ti.dm_enabled=1 AND ti.telegram_chat_id IS NOT NULL`, campaignID, campaignID)
		if rows != nil {
			defer rows.Close()
			seen := map[int64]bool{}
			for _, t := range targets {
				seen[t.chatID] = true
			}
			for rows.Next() {
				var cid *int64
				_ = rows.Scan(&cid)
				if cid != nil && !seen[*cid] {
					seen[*cid] = true
					targets = append(targets, struct {
						chatID int64
						typ    string
					}{*cid, "dm"})
				}
			}
		}
		if len(targets) == 0 {
			return
		}
		text := title + "\n\n" + content
		chunks := chunkMessage(text, 4000)
		useDoc := len(chunks) > 3
		for _, tgt := range targets {
			res, err := db.DB.Exec(`INSERT OR IGNORE INTO telegram_deliveries(recap_id,campaign_id,target_type,target_chat_id,kind,status) VALUES(?,?,?,?,?,?)`,
				recapID, campaignID, tgt.typ, tgt.chatID, kind, "pending")
			if err != nil {
				continue
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				continue
			}
			var lastErr string
			var success bool
			var msgID *int64
			if useDoc {
				id, err := SendDocument(tgt.chatID, fmt.Sprintf("recap-%d.txt", recapID), text)
				if err != nil {
					lastErr = err.Error()
				} else {
					success = true
					msgID = &id
				}
			} else {
				for _, ch := range chunks {
					ch = escapeHTML(strings.TrimSpace(ch))
					if ch == "" {
						continue
					}
					id, err := SendMessage(tgt.chatID, ch)
					if err != nil {
						if ra, ok := err.(*RetryAfterError); ok {
							lastErr = ra.Error()
						} else {
							lastErr = err.Error()
						}
						break
					}
					if msgID == nil {
						msgID = &id
					}
				}
				if lastErr == "" {
					success = true
				}
			}
			if success {
				_, _ = db.DB.Exec(`UPDATE telegram_deliveries SET status='sent', telegram_message_id=?, sent_at=datetime('now'), attempt_count=attempt_count+1 WHERE recap_id=? AND target_chat_id=?`, msgID, recapID, tgt.chatID)
			} else {
				_, _ = db.DB.Exec(`UPDATE telegram_deliveries SET status='failed', last_error=?, attempt_count=attempt_count+1 WHERE recap_id=? AND target_chat_id=?`, lastErr, recapID, tgt.chatID)
				middleware.LogWarn("telegram", "delivery failed", "recap", recapID, "chat", tgt.chatID, "error", lastErr)
			}
		}
		retryFailed()
	}()
}
func retryFailed() {
	rows, err := db.DB.Query(`SELECT recap_id, target_chat_id FROM telegram_deliveries WHERE status='failed' AND attempt_count < 3 LIMIT 10`)
	if err != nil {
		return
	}
	defer rows.Close()
	type r struct{ recap, chat int64 }
	var list []r
	for rows.Next() {
		var a, b int64
		_ = rows.Scan(&a, &b)
		list = append(list, r{a, b})
	}
	for _, it := range list {
		var content, title string
		var cid int64
		if err := db.DB.QueryRow("SELECT campaign_id, title, content FROM campaign_recaps WHERE id=?", it.recap).Scan(&cid, &title, &content); err != nil {
			continue
		}
		text := title + "\n\n" + content
		chunks := chunkMessage(text, 4000)
		var lastErr string
		var success bool
		for _, ch := range chunks {
			ch = escapeHTML(strings.TrimSpace(ch))
			if ch == "" {
				continue
			}
			_, err := SendMessage(it.chat, ch)
			if err != nil {
				lastErr = err.Error()
				break
			}
			success = lastErr == ""
		}
		if success && lastErr == "" {
			_, _ = db.DB.Exec(`UPDATE telegram_deliveries SET status='sent', sent_at=datetime('now'), attempt_count=attempt_count+1 WHERE recap_id=? AND target_chat_id=?`, it.recap, it.chat)
		} else {
			_, _ = db.DB.Exec(`UPDATE telegram_deliveries SET last_error=?, attempt_count=attempt_count+1 WHERE recap_id=? AND target_chat_id=?`, lastErr, it.recap, it.chat)
		}
	}
}
