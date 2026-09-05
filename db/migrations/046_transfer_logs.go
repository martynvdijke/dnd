package migrations

var migration046SQL = `
-- Audit log for villum-transfer imports
CREATE TABLE IF NOT EXISTS transfer_import_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL DEFAULT '',
    doc_type TEXT NOT NULL DEFAULT '',
    counts TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'completed' CHECK(status IN ('completed','failed','rolled_back')),
    error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_transfer_import_logs_user ON transfer_import_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_transfer_import_logs_created ON transfer_import_logs(created_at);
`
