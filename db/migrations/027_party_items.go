package migrations

var migration027SQL = `
-- Campaign completeness features: party_items, session_plans
CREATE TABLE IF NOT EXISTS party_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 1,
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS session_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    session_date TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','planned','ready','completed','cancelled')),
    dm_notes TEXT NOT NULL DEFAULT '',
    planned_encounters TEXT NOT NULL DEFAULT '[]',
    npc_ids TEXT NOT NULL DEFAULT '[]',
    player_goals TEXT NOT NULL DEFAULT '[]',
    expected_duration INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_party_items_campaign ON party_items(campaign_id);
CREATE INDEX IF NOT EXISTS idx_session_plans_campaign ON session_plans(campaign_id);
`
