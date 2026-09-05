package migrations

var migration004SQL = `
CREATE TABLE IF NOT EXISTS campaigns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    dm_notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS rest_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    rest_type TEXT NOT NULL CHECK(rest_type IN ('short','long')),
    hp_healed INTEGER NOT NULL DEFAULT 0,
    slots_recovered TEXT NOT NULL DEFAULT '[]',
    hit_dice_spent INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    timestamp TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS _migration_v4_check (dummy INTEGER);
`
