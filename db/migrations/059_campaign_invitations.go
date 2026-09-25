package migrations

var migration059SQL = `
CREATE TABLE IF NOT EXISTS campaign_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    email TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'player' CHECK(role IN ('player','dm')),
    invited_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT,
    accepted_at TEXT,
    accepted_by INTEGER
);

CREATE INDEX IF NOT EXISTS idx_campaign_invitations_token ON campaign_invitations(token);
CREATE INDEX IF NOT EXISTS idx_campaign_invitations_campaign ON campaign_invitations(campaign_id);
`
