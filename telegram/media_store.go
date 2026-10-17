package telegram

import (
	"database/sql"
	"villum/db"
)

// MediaPostRow mirrors telegram_media_posts for read/write helpers.
type MediaPostRow struct {
	ID                int64
	UploadID          int64
	CampaignID        sql.NullInt64
	TargetType        string
	TargetChatID      int64
	TelegramMessageID sql.NullInt64
	Kind              string
	Status            string
	RecapID           sql.NullInt64
	AttemptCount      int
	LastError         string
	CreatedAt         string
	SentAt            sql.NullString
}

// InsertMediaPostPending inserts a pending row and returns its id.
func InsertMediaPostPending(uploadID int64, campaignID *int64, targetType string, targetChatID int64, kind string, recapID *int64) (int64, error) {
	var cid any
	if campaignID != nil {
		cid = *campaignID
	}
	var rid any
	if recapID != nil {
		rid = *recapID
	}
	res, err := db.DB.Exec(
		`INSERT INTO telegram_media_posts(upload_id, campaign_id, target_type, target_chat_id, kind, status, recap_id, attempt_count, last_error) VALUES(?,?,?,?,?,'pending',?,0,'')`,
		uploadID, cid, targetType, targetChatID, kind, rid,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateMediaPostSent marks a row as sent.
func UpdateMediaPostSent(id int64, telegramMessageID int64) error {
	_, err := db.DB.Exec(
		`UPDATE telegram_media_posts SET status='sent', telegram_message_id=?, sent_at=datetime('now'), attempt_count=attempt_count+1 WHERE id=?`,
		telegramMessageID, id,
	)
	return err
}

// UpdateMediaPostFailed marks a row as failed.
func UpdateMediaPostFailed(id int64, errMsg string) error {
	_, err := db.DB.Exec(
		`UPDATE telegram_media_posts SET status='failed', last_error=?, attempt_count=attempt_count+1 WHERE id=?`,
		errMsg, id,
	)
	return err
}
