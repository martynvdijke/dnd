package migrations

var migration066SQL = `
CREATE TABLE IF NOT EXISTS telegram_initiative (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    initiative INTEGER NOT NULL DEFAULT 0,
    hp INTEGER NOT NULL DEFAULT 0,
    hp_max INTEGER NOT NULL DEFAULT 0,
    turn_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_telegram_initiative_chat ON telegram_initiative(chat_id, turn_order);

CREATE TABLE IF NOT EXISTS telegram_session_polls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    message_id INTEGER NOT NULL DEFAULT 0,
    poll_id TEXT NOT NULL DEFAULT '',
    question TEXT NOT NULL DEFAULT '',
    options TEXT NOT NULL DEFAULT '',
    closes_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_telegram_session_polls_chat ON telegram_session_polls(chat_id);

CREATE TABLE IF NOT EXISTS telegram_forum_topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    campaign_id INTEGER,
    quest_id INTEGER,
    name TEXT NOT NULL,
    thread_id INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_telegram_forum_topics_unique ON telegram_forum_topics(chat_id, name);
`
