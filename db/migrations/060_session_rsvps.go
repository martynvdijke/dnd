package migrations

var migration060SQL = `
CREATE TABLE IF NOT EXISTS session_rsvps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_plan_id INTEGER NOT NULL REFERENCES session_plans(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'maybe' CHECK(status IN ('yes','maybe','no')),
    note TEXT NOT NULL DEFAULT '',
    attended INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(session_plan_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_session_rsvps_plan ON session_rsvps(session_plan_id);
CREATE INDEX IF NOT EXISTS idx_session_rsvps_user ON session_rsvps(user_id);
`
