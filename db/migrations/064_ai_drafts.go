package migrations

var migration064SQL = `
CREATE TABLE IF NOT EXISTS ai_draft_sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL,
    campaign_id INTEGER,
    parent_id INTEGER,
    status TEXT NOT NULL DEFAULT 'drafting',
    messages TEXT NOT NULL DEFAULT '[]',
    draft TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_ai_draft_sessions_user ON ai_draft_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_draft_sessions_entity ON ai_draft_sessions(entity_type);
`
