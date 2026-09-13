package migrations

var migration055SQL = `
CREATE TABLE IF NOT EXISTS external_import_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed',
    created_ids TEXT NOT NULL DEFAULT '[]',
    counts TEXT NOT NULL DEFAULT '{}',
    filename TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    rolled_back_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_external_import_logs_user ON external_import_logs(user_id);
`
