package migrations

var migration061SQL = `
CREATE TABLE IF NOT EXISTS battlemap_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    combat_entry_id INTEGER REFERENCES combat_entries(id) ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    x REAL NOT NULL DEFAULT 0.5,
    y REAL NOT NULL DEFAULT 0.5,
    color TEXT NOT NULL DEFAULT '#b8963e',
    size REAL NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_battlemap_tokens_campaign ON battlemap_tokens(campaign_id);
CREATE INDEX IF NOT EXISTS idx_battlemap_tokens_entry ON battlemap_tokens(combat_entry_id);
`
