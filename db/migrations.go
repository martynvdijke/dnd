package db

import (
	"fmt"
	"log"
)

var migrations = []struct {
	version int
	sql     string
}{
	{
		version: 1,
		sql: `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin','user')),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS characters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    race TEXT NOT NULL DEFAULT '',
    class TEXT NOT NULL DEFAULT '',
    subclass TEXT NOT NULL DEFAULT '',
    level INTEGER NOT NULL DEFAULT 1,
    xp INTEGER NOT NULL DEFAULT 0,
    background TEXT NOT NULL DEFAULT '',
    alignment TEXT NOT NULL DEFAULT '',
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    ac INTEGER NOT NULL DEFAULT 10,
    initiative INTEGER NOT NULL DEFAULT 0,
    speed INTEGER NOT NULL DEFAULT 30,
    hp_max INTEGER NOT NULL DEFAULT 10,
    hp_current INTEGER NOT NULL DEFAULT 10,
    temp_hp INTEGER NOT NULL DEFAULT 0,
    hit_dice TEXT NOT NULL DEFAULT '1d10',
    hit_dice_current INTEGER NOT NULL DEFAULT 1,
    proficiency_bonus INTEGER NOT NULL DEFAULT 2,
    inspiration INTEGER NOT NULL DEFAULT 0,
    passive_perception INTEGER NOT NULL DEFAULT 10,
    personality_traits TEXT NOT NULL DEFAULT '',
    ideals TEXT NOT NULL DEFAULT '',
    bonds TEXT NOT NULL DEFAULT '',
    flaws TEXT NOT NULL DEFAULT '',
    appearance TEXT NOT NULL DEFAULT '',
    backstory TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS character_currency (
    character_id INTEGER PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    cp INTEGER NOT NULL DEFAULT 0,
    sp INTEGER NOT NULL DEFAULT 0,
    ep INTEGER NOT NULL DEFAULT 0,
    gp INTEGER NOT NULL DEFAULT 0,
    pp INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS character_proficiencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('skill','save','tool','weapon','armor','language','other')),
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS character_features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    level_gained INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS character_spellcasting (
    character_id INTEGER PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    ability TEXT NOT NULL DEFAULT '',
    save_dc INTEGER NOT NULL DEFAULT 10,
    attack_bonus INTEGER NOT NULL DEFAULT 0,
    slots_1_max INTEGER NOT NULL DEFAULT 0,
    slots_1_used INTEGER NOT NULL DEFAULT 0,
    slots_2_max INTEGER NOT NULL DEFAULT 0,
    slots_2_used INTEGER NOT NULL DEFAULT 0,
    slots_3_max INTEGER NOT NULL DEFAULT 0,
    slots_3_used INTEGER NOT NULL DEFAULT 0,
    slots_4_max INTEGER NOT NULL DEFAULT 0,
    slots_4_used INTEGER NOT NULL DEFAULT 0,
    slots_5_max INTEGER NOT NULL DEFAULT 0,
    slots_5_used INTEGER NOT NULL DEFAULT 0,
    slots_6_max INTEGER NOT NULL DEFAULT 0,
    slots_6_used INTEGER NOT NULL DEFAULT 0,
    slots_7_max INTEGER NOT NULL DEFAULT 0,
    slots_7_used INTEGER NOT NULL DEFAULT 0,
    slots_8_max INTEGER NOT NULL DEFAULT 0,
    slots_8_used INTEGER NOT NULL DEFAULT 0,
    slots_9_max INTEGER NOT NULL DEFAULT 0,
    slots_9_used INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS spells (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 0,
    school TEXT NOT NULL DEFAULT '',
    casting_time TEXT NOT NULL DEFAULT '',
    range TEXT NOT NULL DEFAULT '',
    components TEXT NOT NULL DEFAULT '',
    duration TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    prepared INTEGER NOT NULL DEFAULT 0,
    always_prepared INTEGER NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS inventory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    weight REAL NOT NULL DEFAULT 0,
    category TEXT NOT NULL DEFAULT 'gear' CHECK(category IN ('weapon','armor','gear','tool','consumable','magic','ammunition')),
    damage_dice TEXT NOT NULL DEFAULT '',
    damage_type TEXT NOT NULL DEFAULT '',
    weapon_properties TEXT NOT NULL DEFAULT '',
    ac_bonus INTEGER NOT NULL DEFAULT 0,
    armor_type TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    is_equipped INTEGER NOT NULL DEFAULT 0,
    is_magical INTEGER NOT NULL DEFAULT 0,
    attunement INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS dice_rolls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id INTEGER REFERENCES characters(id) ON DELETE SET NULL,
    expression TEXT NOT NULL,
    result TEXT NOT NULL,
    total INTEGER NOT NULL,
    timestamp TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS backup_settings (
    id INTEGER PRIMARY KEY CHECK(id=1),
    enabled INTEGER NOT NULL DEFAULT 1,
    interval_hours INTEGER NOT NULL DEFAULT 168,
    last_backup TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS compendium_races (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    speed INTEGER NOT NULL DEFAULT 30,
    size TEXT NOT NULL DEFAULT 'Medium',
    ability_bonuses TEXT NOT NULL DEFAULT '{}',
    traits TEXT NOT NULL DEFAULT '{}',
    languages TEXT NOT NULL DEFAULT '',
    source_page TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS compendium_classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    hit_die INTEGER NOT NULL DEFAULT 8,
    primary_ability TEXT NOT NULL DEFAULT '',
    saving_throws TEXT NOT NULL DEFAULT '[]',
    proficiencies TEXT NOT NULL DEFAULT '{}',
    spellcasting_ability TEXT NOT NULL DEFAULT '',
    source_page TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS compendium_spells (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 0,
    school TEXT NOT NULL DEFAULT '',
    casting_time TEXT NOT NULL DEFAULT '',
    range TEXT NOT NULL DEFAULT '',
    components TEXT NOT NULL DEFAULT '',
    duration TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    higher_levels TEXT NOT NULL DEFAULT '',
    classes TEXT NOT NULL DEFAULT '[]',
    source_page TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS compendium_feats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    prerequisites TEXT NOT NULL DEFAULT '[]',
    source_page TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS compendium_backgrounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    feature_name TEXT NOT NULL DEFAULT '',
    feature_description TEXT NOT NULL DEFAULT '',
    proficiencies TEXT NOT NULL DEFAULT '{}',
    source_page TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS compendium_equipment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    cost TEXT NOT NULL DEFAULT '{}',
    weight REAL NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    source_page TEXT NOT NULL DEFAULT ''
);

CREATE VIRTUAL TABLE IF NOT EXISTS characters_fts USING fts5(
    name, race, class, background, backstory,
    content='characters',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS characters_ai AFTER INSERT ON characters BEGIN
    INSERT INTO characters_fts(rowid, name, race, class, background, backstory)
    VALUES (new.id, new.name, new.race, new.class, new.background, new.backstory);
END;

CREATE TRIGGER IF NOT EXISTS characters_ad AFTER DELETE ON characters BEGIN
    INSERT INTO characters_fts(characters_fts, rowid, name, race, class, background, backstory)
    VALUES ('delete', old.id, old.name, old.race, old.class, old.background, old.backstory);
END;

CREATE TRIGGER IF NOT EXISTS characters_au AFTER UPDATE ON characters BEGIN
    INSERT INTO characters_fts(characters_fts, rowid, name, race, class, background, backstory)
    VALUES ('delete', old.id, old.name, old.race, old.class, old.background, old.backstory);
    INSERT INTO characters_fts(rowid, name, race, class, background, backstory)
    VALUES (new.id, new.name, new.race, new.class, new.background, new.backstory);
END;

CREATE INDEX IF NOT EXISTS idx_characters_user_id ON characters(user_id);
CREATE INDEX IF NOT EXISTS idx_spells_character_id ON spells(character_id);
CREATE INDEX IF NOT EXISTS idx_inventory_character_id ON inventory(character_id);
CREATE INDEX IF NOT EXISTS idx_features_character_id ON character_features(character_id);
CREATE INDEX IF NOT EXISTS idx_proficiencies_character_id ON character_proficiencies(character_id);
CREATE INDEX IF NOT EXISTS idx_dice_rolls_user_id ON dice_rolls(user_id);
CREATE INDEX IF NOT EXISTS idx_dice_rolls_character_id ON dice_rolls(character_id);
`,
	},
}

func Migrate() error {
	var current int
	err := DB.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_version").Scan(&current)
	if err != nil {
		// table doesn't exist yet
		current = 0
	}

	for _, m := range migrations {
		if m.version > current {
			log.Printf("Running migration v%d", m.version)
			tx, err := DB.Begin()
			if err != nil {
				return fmt.Errorf("begin migration v%d: %w", m.version, err)
			}
			if _, err := tx.Exec(m.sql); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration v%d: %w", m.version, err)
			}
			if _, err := tx.Exec("INSERT OR REPLACE INTO schema_version(version) VALUES(?)", m.version); err != nil {
				tx.Rollback()
				return fmt.Errorf("record migration v%d: %w", m.version, err)
			}
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit migration v%d: %w", m.version, err)
			}
		}
	}

	return nil
}
