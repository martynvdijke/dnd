package migrations

var migration022SQL = `
-- Clue/Mystery Tracker tables
CREATE TABLE IF NOT EXISTS clues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    clue_type TEXT NOT NULL DEFAULT 'direct' CHECK(clue_type IN ('direct','witness','object','location')),
    is_red_herring INTEGER NOT NULL DEFAULT 0,
    is_revealed INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS clue_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    clue_id INTEGER NOT NULL REFERENCES clues(id) ON DELETE CASCADE,
    depends_on_id INTEGER NOT NULL REFERENCES clues(id) ON DELETE CASCADE,
    UNIQUE(clue_id, depends_on_id)
);

CREATE TABLE IF NOT EXISTS clue_npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    clue_id INTEGER NOT NULL REFERENCES clues(id) ON DELETE CASCADE,
    npc_id INTEGER NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    UNIQUE(clue_id, npc_id)
);

CREATE TABLE IF NOT EXISTS clue_locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    clue_id INTEGER NOT NULL REFERENCES clues(id) ON DELETE CASCADE,
    location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    UNIQUE(clue_id, location_id)
);
`
