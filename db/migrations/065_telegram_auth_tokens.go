package migrations

var migration065SQL = `
CREATE TABLE IF NOT EXISTS telegram_auth_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token_hash TEXT NOT NULL UNIQUE,
    tg_user_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL,
    tg_username TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_telegram_auth_tokens_hash ON telegram_auth_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_telegram_auth_tokens_tg_user ON telegram_auth_tokens(tg_user_id);
`
