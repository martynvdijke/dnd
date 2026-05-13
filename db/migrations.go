package db

import (
	"fmt"
	"log"
	"strings"
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
    email TEXT NOT NULL DEFAULT '',
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
	{
		version: 2,
		sql: `
CREATE TABLE IF NOT EXISTS locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'region',
    description TEXT NOT NULL DEFAULT '',
    parent_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    latitude REAL,
    longitude REAL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS character_locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    relationship TEXT NOT NULL DEFAULT 'visited' CHECK(relationship IN ('current','hometown','visited','headquarters','quest','other')),
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE(character_id, location_id)
);

CREATE TABLE IF NOT EXISTS npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    race TEXT NOT NULL DEFAULT '',
    class TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    hp_max INTEGER NOT NULL DEFAULT 10,
    hp_current INTEGER NOT NULL DEFAULT 10,
    is_alive INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS character_npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    npc_id INTEGER NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    relationship TEXT NOT NULL DEFAULT 'acquaintance' CHECK(relationship IN ('ally','enemy','family','contact','acquaintance','pet','deity','other')),
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE(character_id, npc_id)
);

CREATE INDEX IF NOT EXISTS idx_locations_user_id ON locations(user_id);
CREATE INDEX IF NOT EXISTS idx_locations_parent_id ON locations(parent_id);
CREATE INDEX IF NOT EXISTS idx_character_locations_character ON character_locations(character_id);
CREATE INDEX IF NOT EXISTS idx_npcs_user_id ON npcs(user_id);
CREATE INDEX IF NOT EXISTS idx_character_npcs_character ON character_npcs(character_id);
`,
	},
	{
		version: 3,
		sql: `
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    session_date TEXT NOT NULL DEFAULT (date('now')),
    title TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    xp_earned INTEGER NOT NULL DEFAULT 0,
    gold_earned INTEGER NOT NULL DEFAULT 0,
    important_events TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS quests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('available','active','complete','failed','abandoned')),
    objectives TEXT NOT NULL DEFAULT '',
    rewards TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS journal (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    entry TEXT NOT NULL DEFAULT '',
    entry_date TEXT NOT NULL DEFAULT (date('now')),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_character ON sessions(character_id);
CREATE INDEX IF NOT EXISTS idx_quests_character ON quests(character_id);
CREATE INDEX IF NOT EXISTS idx_journal_character ON journal(character_id);
`,
	},
	{
		version: 4,
		sql: `
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
`,
	},
	{
		version: 5,
		sql: `
CREATE TABLE IF NOT EXISTS _migration_v5_check (dummy INTEGER);
`,
	},
	{
		version: 6,
		sql: `
CREATE TABLE IF NOT EXISTS combat_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER REFERENCES characters(id) ON DELETE CASCADE,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'character' CHECK(type IN ('character','monster','npc')),
    initiative_roll INTEGER NOT NULL DEFAULT 0,
    initiative_mod INTEGER NOT NULL DEFAULT 0,
    hp_max INTEGER NOT NULL DEFAULT 1,
    hp_current INTEGER NOT NULL DEFAULT 1,
    ac INTEGER NOT NULL DEFAULT 10,
    is_active INTEGER NOT NULL DEFAULT 1,
    turn_order INTEGER NOT NULL DEFAULT 0,
    condition_ids TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_combat_campaign ON combat_entries(campaign_id);
CREATE INDEX IF NOT EXISTS idx_combat_character ON combat_entries(character_id);
`,
	},
	{
		version: 7,
		sql: `
CREATE TABLE IF NOT EXISTS uploads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT UNIQUE,
    ext TEXT,
    url TEXT,
    resized_url TEXT,
    thumbnail_url TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS email_settings (
    id INTEGER PRIMARY KEY,
    smtp_host TEXT,
    smtp_port INTEGER DEFAULT 587,
    username TEXT,
    password TEXT,
    from_addr TEXT,
    enabled INTEGER DEFAULT 0
);

INSERT OR IGNORE INTO email_settings (id, enabled) VALUES (1, 0);
`,
	},
	{
		version: 8,
		sql: `
ALTER TABLE backup_settings ADD COLUMN interval_days INTEGER NOT NULL DEFAULT 7;
ALTER TABLE backup_settings ADD COLUMN keep_count INTEGER NOT NULL DEFAULT 7;
UPDATE backup_settings SET interval_days = MAX(1, COALESCE(interval_hours, 168) / 24) WHERE id = 1;
`,
	},
	{
		version: 9,
		sql: `
CREATE TABLE IF NOT EXISTS campaign_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(campaign_id, user_id)
);
`,
	},
	{
		version: 10,
		sql: `
ALTER TABLE campaign_members ADD COLUMN role TEXT NOT NULL DEFAULT 'player' CHECK(role IN ('dm','player'));
`,
	},
	{
		version: 11,
		sql: `
ALTER TABLE compendium_races ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_races ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_classes ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_classes ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_spells ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_spells ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_feats ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_feats ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_backgrounds ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_backgrounds ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_equipment ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_equipment ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
`,
	},
	{
		version: 12,
		sql: `
CREATE TABLE IF NOT EXISTS character_classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    class TEXT NOT NULL,
    subclass TEXT NOT NULL DEFAULT '',
    level INTEGER NOT NULL DEFAULT 1,
    hit_dice TEXT NOT NULL DEFAULT 'd10',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS encounter_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    environment TEXT NOT NULL DEFAULT '',
    difficulty TEXT NOT NULL DEFAULT 'medium' CHECK(difficulty IN ('easy','medium','hard','deadly')),
    xp_budget INTEGER NOT NULL DEFAULT 0,
    total_xp INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS encounter_monsters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    encounter_id INTEGER NOT NULL REFERENCES encounter_templates(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    count INTEGER NOT NULL DEFAULT 1,
    cr TEXT NOT NULL DEFAULT '0',
    xp INTEGER NOT NULL DEFAULT 0,
    ac INTEGER NOT NULL DEFAULT 10,
    hp INTEGER NOT NULL DEFAULT 1,
    initiative_mod INTEGER NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'homebrew',
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS campaign_calendar_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    event_date TEXT NOT NULL,
    event_type TEXT NOT NULL DEFAULT 'other' CHECK(event_type IN ('session','quest','holiday','weather','combat','festival','other')),
    color TEXT NOT NULL DEFAULT '#b8963e',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS campaign_timeline_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    event_date TEXT NOT NULL,
    event_type TEXT NOT NULL DEFAULT 'other' CHECK(event_type IN ('session','quest','combat','discovery','npc','location','milestone','other')),
    importance INTEGER NOT NULL DEFAULT 1 CHECK(importance BETWEEN 1 AND 5),
    icon TEXT NOT NULL DEFAULT 'fa-star',
    linked_entity_type TEXT NOT NULL DEFAULT '',
    linked_entity_id INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_char_classes_char ON character_classes(character_id);
CREATE INDEX IF NOT EXISTS idx_encounter_campaign ON encounter_templates(campaign_id);
CREATE INDEX IF NOT EXISTS idx_encounter_monsters_enc ON encounter_monsters(encounter_id);
CREATE INDEX IF NOT EXISTS idx_calendar_campaign ON campaign_calendar_events(campaign_id);
CREATE INDEX IF NOT EXISTS idx_timeline_campaign ON campaign_timeline_events(campaign_id);
`,
	},
	{
		version: 13,
		sql: `
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
`,
	},
	{
		version: 14,
		sql: `
ALTER TABLE campaigns ADD COLUMN party_name TEXT NOT NULL DEFAULT '';
`,
	},
}

func Migrate() error {
	var current int
	err := DB.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_version").Scan(&current)
	if err != nil {
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

	// Run safe ALTER TABLE additions (ignore if column already exists)
	alterStatements := []string{
		"ALTER TABLE characters ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL",
		"ALTER TABLE character_npcs ADD COLUMN interaction_count INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE character_npcs ADD COLUMN last_interacted TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE characters ADD COLUMN death_saves_successes INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE characters ADD COLUMN death_saves_failures INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE characters ADD COLUMN concentrating_on TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE backup_settings ADD COLUMN interval_days INTEGER NOT NULL DEFAULT 7",
		"ALTER TABLE backup_settings ADD COLUMN keep_count INTEGER NOT NULL DEFAULT 7",
		"ALTER TABLE campaign_members ADD COLUMN role TEXT NOT NULL DEFAULT 'player'",
		"ALTER TABLE characters ADD COLUMN portrait_url TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE characters ADD COLUMN dm_notes TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE characters ADD COLUMN hp_auto_calc INTEGER NOT NULL DEFAULT 0",
	}
	for _, stmt := range alterStatements {
		if _, err := DB.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("alter table: %w", err)
			}
		}
	}

	return nil
}
