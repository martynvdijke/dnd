package migrations

var migration026SQL = `
-- Restructure App v26: Role system, Party, Items, Monsters, NPC tiers, Uploads

-- Update users table to allow 'dm' role
PRAGMA foreign_keys=OFF;
CREATE TABLE IF NOT EXISTS users_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin','user','dm')),
    email TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
INSERT OR IGNORE INTO users_v2 SELECT id, username, password, display_name, role, email, created_at FROM users;
DROP TABLE IF EXISTS users;
ALTER TABLE users_v2 RENAME TO users;
PRAGMA foreign_keys=ON;

-- Parties
CREATE TABLE IF NOT EXISTS parties (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- One-Shot Items
CREATE TABLE IF NOT EXISTS oneshot_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'gear',
    quantity INTEGER NOT NULL DEFAULT 1,
    weight REAL NOT NULL DEFAULT 0,
    price_gp REAL NOT NULL DEFAULT 0,
    is_magical INTEGER NOT NULL DEFAULT 0,
    attunement INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Inline Monsters per act/scene
CREATE TABLE IF NOT EXISTS oneshot_monsters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE,
    scene_id INTEGER REFERENCES oneshot_scenes(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    ac INTEGER NOT NULL DEFAULT 10,
    hp INTEGER NOT NULL DEFAULT 1,
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int_ INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    cr TEXT NOT NULL DEFAULT '0',
    source TEXT NOT NULL DEFAULT 'homebrew',
    is_full INTEGER NOT NULL DEFAULT 0,
    saves TEXT NOT NULL DEFAULT '',
    skills TEXT NOT NULL DEFAULT '',
    damage_vulnerabilities TEXT NOT NULL DEFAULT '',
    damage_resistances TEXT NOT NULL DEFAULT '',
    damage_immunities TEXT NOT NULL DEFAULT '',
    condition_immunities TEXT NOT NULL DEFAULT '',
    senses TEXT NOT NULL DEFAULT '',
    languages TEXT NOT NULL DEFAULT '',
    special_abilities TEXT NOT NULL DEFAULT '',
    actions TEXT NOT NULL DEFAULT '',
    legendary_actions TEXT NOT NULL DEFAULT '',
    library_id INTEGER DEFAULT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Monster Library
CREATE TABLE IF NOT EXISTS monster_library (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    ac INTEGER NOT NULL DEFAULT 10,
    hp INTEGER NOT NULL DEFAULT 1,
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int_ INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    cr TEXT NOT NULL DEFAULT '0',
    source TEXT NOT NULL DEFAULT 'homebrew',
    is_full INTEGER NOT NULL DEFAULT 0,
    saves TEXT NOT NULL DEFAULT '',
    skills TEXT NOT NULL DEFAULT '',
    damage_vulnerabilities TEXT NOT NULL DEFAULT '',
    damage_resistances TEXT NOT NULL DEFAULT '',
    damage_immunities TEXT NOT NULL DEFAULT '',
    condition_immunities TEXT NOT NULL DEFAULT '',
    senses TEXT NOT NULL DEFAULT '',
    languages TEXT NOT NULL DEFAULT '',
    special_abilities TEXT NOT NULL DEFAULT '',
    actions TEXT NOT NULL DEFAULT '',
    legendary_actions TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- NPC-Item Links
CREATE TABLE IF NOT EXISTS npc_item_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    npc_id INTEGER NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    item_id INTEGER REFERENCES oneshot_items(id) ON DELETE CASCADE,
    compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL,
    relationship_type TEXT NOT NULL DEFAULT 'owns',
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE(npc_id, item_id)
);

-- Linked Player Characters
CREATE TABLE IF NOT EXISTS oneshot_player_characters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE(adventure_id, character_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_parties_user ON parties(user_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_items_adventure ON oneshot_items(adventure_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_monsters_act ON oneshot_monsters(act_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_monsters_scene ON oneshot_monsters(scene_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_monsters_adventure ON oneshot_monsters(adventure_id);
CREATE INDEX IF NOT EXISTS idx_monster_library_user ON monster_library(user_id);
CREATE INDEX IF NOT EXISTS idx_npc_item_links_npc ON npc_item_links(npc_id);
CREATE INDEX IF NOT EXISTS idx_npc_item_links_item ON npc_item_links(item_id);
CREATE INDEX IF NOT EXISTS idx_npc_item_links_adventure ON npc_item_links(adventure_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_pc_adventure ON oneshot_player_characters(adventure_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_pc_character ON oneshot_player_characters(character_id);
`
