package migrations

var migration063SQL = `
CREATE TABLE IF NOT EXISTS telegram_character_claims (
    telegram_user_id INTEGER PRIMARY KEY,
    character_id INTEGER NOT NULL UNIQUE REFERENCES characters(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`
