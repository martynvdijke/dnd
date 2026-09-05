package migrations

var migration013SQL = `
CREATE TABLE IF NOT EXISTS character_conditions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'other' CHECK(type IN ('blinded','charmed','deafened','exhaustion','frightened','grappled','incapacitated','invisible','paralyzed','petrified','poisoned','prone','restrained','stunned','unconscious','concentration','other')),
    source TEXT NOT NULL DEFAULT '',
    duration INTEGER NOT NULL DEFAULT 0,
    duration_type TEXT NOT NULL DEFAULT 'round' CHECK(duration_type IN ('round','minute','hour','day','permanent')),
    saving_throw TEXT NOT NULL DEFAULT '',
    save_dc INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    started_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS character_feats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    prerequisites TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    level_gained INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS companions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'companion' CHECK(type IN ('familiar','mount','companion','summoned','pet')),
    race TEXT NOT NULL DEFAULT '',
    hp_max INTEGER NOT NULL DEFAULT 1,
    hp_current INTEGER NOT NULL DEFAULT 1,
    ac INTEGER NOT NULL DEFAULT 10,
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    speed INTEGER NOT NULL DEFAULT 30,
    abilities TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    portrait_url TEXT NOT NULL DEFAULT '',
    is_alive INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS factions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'organization' CHECK(type IN ('organization','guild','government','religion','cult','military','other')),
    headquarters TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS faction_reputation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    faction_id INTEGER NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    standing INTEGER NOT NULL DEFAULT 0 CHECK(standing >= -100 AND standing <= 100),
    rank TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE(character_id, faction_id)
);

CREATE TABLE IF NOT EXISTS character_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    visibility TEXT NOT NULL DEFAULT 'player' CHECK(visibility IN ('player','dm','both')),
    category TEXT NOT NULL DEFAULT 'general' CHECK(category IN ('general','backstory','quest','lore','dm','other')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_conditions_char ON character_conditions(character_id);
CREATE INDEX IF NOT EXISTS idx_feats_char ON character_feats(character_id);
CREATE INDEX IF NOT EXISTS idx_companions_char ON companions(character_id);
CREATE INDEX IF NOT EXISTS idx_factions_campaign ON factions(campaign_id);
CREATE INDEX IF NOT EXISTS idx_faction_rep_char ON faction_reputation(character_id);
CREATE INDEX IF NOT EXISTS idx_faction_rep_faction ON faction_reputation(faction_id);
CREATE INDEX IF NOT EXISTS idx_notes_char ON character_notes(character_id);
`
