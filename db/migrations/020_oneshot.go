package migrations

var migration020SQL = `
-- One-Shot Adventure Builder tables
CREATE TABLE IF NOT EXISTS oneshot_adventures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL,
    title TEXT NOT NULL DEFAULT '',
    premise TEXT NOT NULL DEFAULT '',
    hook TEXT NOT NULL DEFAULT '',
    template TEXT NOT NULL DEFAULT 'custom',
    estimated_minutes INTEGER NOT NULL DEFAULT 180,
    difficulty TEXT NOT NULL DEFAULT 'medium',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS oneshot_acts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    number INTEGER NOT NULL DEFAULT 1,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    estimated_minutes INTEGER NOT NULL DEFAULT 30
);

CREATE TABLE IF NOT EXISTS oneshot_scenes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    act_id INTEGER NOT NULL REFERENCES oneshot_acts(id) ON DELETE CASCADE,
    number INTEGER NOT NULL DEFAULT 1,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    scene_type TEXT NOT NULL DEFAULT 'roleplay',
    location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    encounter_id INTEGER REFERENCES encounter_templates(id) ON DELETE SET NULL,
    estimated_minutes INTEGER NOT NULL DEFAULT 15,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS oneshot_adventure_npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    npc_id INTEGER NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT '',
    UNIQUE(adventure_id, npc_id)
);

CREATE TABLE IF NOT EXISTS oneshot_adventure_locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    UNIQUE(adventure_id, location_id)
);

CREATE TABLE IF NOT EXISTS oneshot_adventure_encounters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    encounter_id INTEGER NOT NULL REFERENCES encounter_templates(id) ON DELETE CASCADE,
    UNIQUE(adventure_id, encounter_id)
);
`
