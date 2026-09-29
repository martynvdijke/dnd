package migrations

var migration062SQL = `
CREATE TABLE IF NOT EXISTS telegram_identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telegram_user_id INTEGER NOT NULL UNIQUE,
    telegram_chat_id INTEGER,
    telegram_username TEXT NOT NULL DEFAULT '',
    dm_enabled INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_telegram_identities_user ON telegram_identities(user_id);

CREATE TABLE IF NOT EXISTS telegram_link_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_telegram_link_codes_user ON telegram_link_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_telegram_link_codes_hash ON telegram_link_codes(code_hash);

CREATE TABLE IF NOT EXISTS campaign_telegram_settings (
    campaign_id INTEGER PRIMARY KEY REFERENCES campaigns(id) ON DELETE CASCADE,
    chat_id INTEGER,
    chat_type TEXT NOT NULL DEFAULT '',
    title_cache TEXT NOT NULL DEFAULT '',
    is_enabled INTEGER NOT NULL DEFAULT 0,
    auto_post_enabled INTEGER NOT NULL DEFAULT 0,
    bound_at TEXT,
    bound_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS telegram_deliveries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recap_id INTEGER NOT NULL REFERENCES campaign_recaps(id) ON DELETE CASCADE,
    campaign_id INTEGER NOT NULL,
    target_type TEXT NOT NULL CHECK(target_type IN ('chat','dm')),
    target_chat_id INTEGER NOT NULL,
    telegram_message_id INTEGER,
    kind TEXT NOT NULL DEFAULT 'manual' CHECK(kind IN ('manual','auto')),
    status TEXT NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    sent_at TEXT,
    UNIQUE(recap_id, target_chat_id)
);
CREATE INDEX IF NOT EXISTS idx_telegram_deliveries_status ON telegram_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_telegram_deliveries_recap ON telegram_deliveries(recap_id);
CREATE INDEX IF NOT EXISTS idx_telegram_deliveries_campaign ON telegram_deliveries(campaign_id);
`
