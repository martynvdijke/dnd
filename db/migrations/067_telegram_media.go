package migrations

var migration067SQL = `
-- NOTE: uploads.kind / source / telegram_file_unique_id and their indexes are
-- added by ApplySafeAlters (db/migrations/safe_alters.go), NOT here. The
-- uploads table is ent-owned, and ent.Schema.Create runs after migrations and
-- rebuilds that table, which drops columns added before it runs. Only
-- raw-owned tables and columns belong in this migration.

ALTER TABLE campaign_telegram_settings ADD COLUMN auto_save_media INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS telegram_media_posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    upload_id INTEGER REFERENCES uploads(id) ON DELETE SET NULL,
    campaign_id INTEGER,
    target_type TEXT NOT NULL CHECK(target_type IN ('chat','dm')),
    target_chat_id INTEGER NOT NULL,
    telegram_message_id INTEGER,
    kind TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    recap_id INTEGER REFERENCES campaign_recaps(id) ON DELETE SET NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    sent_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_telegram_media_posts_status ON telegram_media_posts(status);
CREATE INDEX IF NOT EXISTS idx_telegram_media_posts_campaign ON telegram_media_posts(campaign_id);
CREATE INDEX IF NOT EXISTS idx_telegram_media_posts_upload ON telegram_media_posts(upload_id);
CREATE INDEX IF NOT EXISTS idx_telegram_media_posts_recap ON telegram_media_posts(recap_id);

CREATE TABLE IF NOT EXISTS telegram_pending_captures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_user_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL,
    file_id TEXT NOT NULL,
    file_unique_id TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT '',
    file_name TEXT NOT NULL DEFAULT '',
    caption TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_telegram_pending_captures_user ON telegram_pending_captures(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_telegram_pending_captures_unique ON telegram_pending_captures(file_unique_id);
`
