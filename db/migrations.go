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
	{
		version: 15,
		sql: `
CREATE TABLE IF NOT EXISTS crafting_recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'other' CHECK(category IN ('potion','scroll','magic-item','poison','other')),
    difficulty_dc INTEGER NOT NULL DEFAULT 10,
    crafting_time_hours REAL NOT NULL DEFAULT 1,
    required_tools TEXT NOT NULL DEFAULT '[]',
    required_materials TEXT NOT NULL DEFAULT '[]',
    result_item_name TEXT NOT NULL,
    result_item_category TEXT NOT NULL DEFAULT 'other',
    result_quantity INTEGER NOT NULL DEFAULT 1,
    result_description TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS character_crafting (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    recipe_id INTEGER REFERENCES crafting_recipes(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    progress_hours REAL NOT NULL DEFAULT 0,
    total_hours_required REAL NOT NULL DEFAULT 1,
    dc INTEGER NOT NULL DEFAULT 10,
    status TEXT NOT NULL DEFAULT 'in-progress' CHECK(status IN ('in-progress','complete','abandoned')),
    materials_allocated TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    started_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_crafting_recipes_user ON crafting_recipes(user_id);
CREATE INDEX IF NOT EXISTS idx_char_crafting_char ON character_crafting(character_id);
`,
	},
	{
		version: 16,
		sql: `
CREATE TABLE IF NOT EXISTS shops (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    markup_percent REAL NOT NULL DEFAULT 100,
    markup_buy_percent REAL NOT NULL DEFAULT 50,
    location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS shop_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    item_name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'gear',
    price_gp REAL NOT NULL DEFAULT 0,
    quantity_available INTEGER NOT NULL DEFAULT -1,
    description TEXT NOT NULL DEFAULT '',
    is_magical INTEGER NOT NULL DEFAULT 0,
    attunement_required INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS shop_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_name TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    price_gp REAL NOT NULL DEFAULT 0,
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('buy','sell')),
    timestamp TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS campaign_wiki_pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES campaign_wiki_pages(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    visibility TEXT NOT NULL DEFAULT 'public' CHECK(visibility IN ('public','dm-only')),
    tags TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_shops_campaign ON shops(campaign_id);
CREATE INDEX IF NOT EXISTS idx_shop_items_shop ON shop_items(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_transactions_shop ON shop_transactions(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_transactions_char ON shop_transactions(character_id);
CREATE INDEX IF NOT EXISTS idx_wiki_campaign ON campaign_wiki_pages(campaign_id);
CREATE INDEX IF NOT EXISTS idx_wiki_parent ON campaign_wiki_pages(parent_id);
`,
	},
	{
		version: 17,
		sql: `
CREATE TABLE IF NOT EXISTS share_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    entity_type TEXT NOT NULL CHECK(entity_type IN ('character','party')),
    entity_id INTEGER NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);
CREATE INDEX IF NOT EXISTS idx_share_links_entity ON share_links(entity_type, entity_id);
`,
	},
	{
		version: 18,
		sql: `
CREATE TABLE IF NOT EXISTS character_resources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    current INTEGER NOT NULL DEFAULT 0,
    max INTEGER NOT NULL DEFAULT 0,
    short_rest_recovery INTEGER NOT NULL DEFAULT 0,
    long_rest_recovery INTEGER NOT NULL DEFAULT 1,
    icon TEXT NOT NULL DEFAULT 'fa-bolt',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS combat_log_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE,
    combat_entry_id INTEGER REFERENCES combat_entries(id) ON DELETE SET NULL,
    actor_name TEXT NOT NULL,
    action TEXT NOT NULL,
    target_name TEXT NOT NULL DEFAULT '',
    damage INTEGER NOT NULL DEFAULT 0,
    damage_type TEXT NOT NULL DEFAULT '',
    healing INTEGER NOT NULL DEFAULT 0,
    condition_applied TEXT NOT NULL DEFAULT '',
    roll_expression TEXT NOT NULL DEFAULT '',
    roll_total INTEGER NOT NULL DEFAULT 0,
    is_critical INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS campaign_maps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    width INTEGER NOT NULL DEFAULT 1000,
    height INTEGER NOT NULL DEFAULT 800,
    grid_size INTEGER NOT NULL DEFAULT 50,
    is_active INTEGER NOT NULL DEFAULT 0,
    fog_of_war TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS campaign_map_pins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    map_id INTEGER NOT NULL REFERENCES campaign_maps(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'poi' CHECK(type IN ('poi','city','dungeon','town','landmark','spawn','custom')),
    x REAL NOT NULL DEFAULT 0,
    y REAL NOT NULL DEFAULT 0,
    icon TEXT NOT NULL DEFAULT 'fa-map-pin',
    color TEXT NOT NULL DEFAULT '#b8963e',
    description TEXT NOT NULL DEFAULT '',
    linked_entity_type TEXT NOT NULL DEFAULT '',
    linked_entity_id INTEGER,
    is_hidden INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS downtime_activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    activity_type TEXT NOT NULL CHECK(activity_type IN ('training','crafting','research','carousing','pit_fighting','crime','religious','scribing','gambling','other')),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    dc INTEGER NOT NULL DEFAULT 10,
    days_required INTEGER NOT NULL DEFAULT 10,
    days_completed INTEGER NOT NULL DEFAULT 0,
    cost_per_day REAL NOT NULL DEFAULT 0,
    total_cost REAL NOT NULL DEFAULT 0,
    reward TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'in-progress' CHECK(status IN ('in-progress','complete','failed')),
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS level_up_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    target_level INTEGER NOT NULL DEFAULT 20,
    plan_data TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS campaign_recaps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    session_start_date TEXT,
    session_end_date TEXT,
    word_count INTEGER NOT NULL DEFAULT 0,
    is_edited INTEGER NOT NULL DEFAULT 0,
    is_sent INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_char_resources_char ON character_resources(character_id);
CREATE INDEX IF NOT EXISTS idx_combat_log_campaign ON combat_log_entries(campaign_id);
CREATE INDEX IF NOT EXISTS idx_combat_log_created ON combat_log_entries(created_at);
CREATE INDEX IF NOT EXISTS idx_campaign_maps_campaign ON campaign_maps(campaign_id);
CREATE INDEX IF NOT EXISTS idx_map_pins_map ON campaign_map_pins(map_id);
CREATE INDEX IF NOT EXISTS idx_downtime_char ON downtime_activities(character_id);
CREATE INDEX IF NOT EXISTS idx_level_plans_char ON level_up_plans(character_id);
CREATE INDEX IF NOT EXISTS idx_recaps_campaign ON campaign_recaps(campaign_id);
`,
	},
	{
		version: 19,
		sql: `
-- Add id column to character_spellcasting (currently character_id is PK, no id column)
CREATE TABLE IF NOT EXISTS character_spellcasting_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
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
    slots_9_used INTEGER NOT NULL DEFAULT 0,
    UNIQUE(character_id)
);
INSERT OR IGNORE INTO character_spellcasting_v2 (character_id, ability, save_dc, attack_bonus,
    slots_1_max, slots_1_used, slots_2_max, slots_2_used, slots_3_max, slots_3_used,
    slots_4_max, slots_4_used, slots_5_max, slots_5_used, slots_6_max, slots_6_used,
    slots_7_max, slots_7_used, slots_8_max, slots_8_used, slots_9_max, slots_9_used)
SELECT character_id, ability, save_dc, attack_bonus,
    slots_1_max, slots_1_used, slots_2_max, slots_2_used, slots_3_max, slots_3_used,
    slots_4_max, slots_4_used, slots_5_max, slots_5_used, slots_6_max, slots_6_used,
    slots_7_max, slots_7_used, slots_8_max, slots_8_used, slots_9_max, slots_9_used
FROM character_spellcasting;
DROP TABLE IF EXISTS character_spellcasting;
ALTER TABLE character_spellcasting_v2 RENAME TO character_spellcasting;

-- Add id column to character_currency (currently character_id is PK, no id column)
CREATE TABLE IF NOT EXISTS character_currency_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    cp INTEGER NOT NULL DEFAULT 0,
    sp INTEGER NOT NULL DEFAULT 0,
    ep INTEGER NOT NULL DEFAULT 0,
    gp INTEGER NOT NULL DEFAULT 0,
    pp INTEGER NOT NULL DEFAULT 0
);
INSERT OR IGNORE INTO character_currency_v2 (character_id, cp, sp, ep, gp, pp)
SELECT character_id, cp, sp, ep, gp, pp FROM character_currency;
DROP TABLE IF EXISTS character_currency;
ALTER TABLE character_currency_v2 RENAME TO character_currency;
`,
	},
	{
		version: 20,
		sql: `
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
`,
	},
	{
		version: 21,
		sql: `
-- Session Pacing tables
CREATE TABLE IF NOT EXISTS session_pacing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    current_act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE SET NULL,
    current_scene_id INTEGER REFERENCES oneshot_scenes(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'running' CHECK(status IN ('running','paused','completed')),
    elapsed_seconds INTEGER NOT NULL DEFAULT 0,
    started_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT
);

CREATE TABLE IF NOT EXISTS scene_timings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES session_pacing(id) ON DELETE CASCADE,
    scene_id INTEGER NOT NULL REFERENCES oneshot_scenes(id) ON DELETE CASCADE,
    elapsed_seconds INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','completed')),
    started_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT
);
`,
	},
	{
		version: 22,
		sql: `
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
`,
	},
	{
		version: 23,
		sql: `
CREATE TABLE IF NOT EXISTS pregen_characters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    race TEXT NOT NULL DEFAULT '',
    class TEXT NOT NULL DEFAULT '',
    subclass TEXT NOT NULL DEFAULT '',
    level INTEGER NOT NULL DEFAULT 1,
    background TEXT NOT NULL DEFAULT '',
    alignment TEXT NOT NULL DEFAULT '',
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    hp INTEGER NOT NULL DEFAULT 10,
    ac INTEGER NOT NULL DEFAULT 10,
    speed INTEGER NOT NULL DEFAULT 30,
    skills TEXT NOT NULL DEFAULT '',
    equipment TEXT NOT NULL DEFAULT '',
    spells TEXT NOT NULL DEFAULT '',
    features TEXT NOT NULL DEFAULT '',
    personality TEXT NOT NULL DEFAULT '',
    backstory TEXT NOT NULL DEFAULT '',
    portrait_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`,
	},
	{
		version: 24,
		sql: `
CREATE TABLE IF NOT EXISTS prep_checklist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    item TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general',
    is_checked INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0
);
`,
	},
	{
		version: 25,
		sql: `
CREATE TABLE IF NOT EXISTS dm_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`,
	},
	{
		version: 26,
		sql: `
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
    item_id INTEGER NOT NULL REFERENCES oneshot_items(id) ON DELETE CASCADE,
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
`,
	},
	{
		version: 27,
		sql: `
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
	`,
	},
	{
		version: 28,
		sql: `
-- Act-level planning: notes, NPCs, and DM notes per act
CREATE TABLE IF NOT EXISTS oneshot_act_npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    act_id INTEGER NOT NULL REFERENCES oneshot_acts(id) ON DELETE CASCADE,
    npc_id INTEGER REFERENCES npcs(id) ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    is_inline INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_oneshot_act_npcs_act ON oneshot_act_npcs(act_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_act_npcs_npc ON oneshot_act_npcs(npc_id);
`,
	},
	{
		version: 29,
		sql: `
-- Act tree: parent_act_id for nested sub-acts, sort_order for ordering
ALTER TABLE oneshot_acts ADD COLUMN parent_act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;
ALTER TABLE oneshot_acts ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Scene ordering with explicit sort_order
ALTER TABLE oneshot_scenes ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Act-level shops, items, encounters (NULL = adventure-level)
ALTER TABLE shops ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;
ALTER TABLE oneshot_items ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;
ALTER TABLE oneshot_adventure_encounters ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;

-- Copy existing number values into sort_order for acts and scenes
UPDATE oneshot_acts SET sort_order = number;
UPDATE oneshot_scenes SET sort_order = number;
`,
	},
	{
		version: 30,
		sql: `
-- Add portrait_url to NPCs for character images
ALTER TABLE npcs ADD COLUMN portrait_url TEXT NOT NULL DEFAULT '';

-- Upload links table for many-to-many entity-file associations
CREATE TABLE IF NOT EXISTS upload_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    upload_id INTEGER NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    field_name TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(upload_id, entity_type, entity_id)
);
CREATE INDEX IF NOT EXISTS idx_upload_links_entity ON upload_links(entity_type, entity_id);
`,
	},
	{
		version: 31,
		sql: `
-- Global monster compendium (SRD monsters)
CREATE TABLE IF NOT EXISTS compendium_monsters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT '',
    size TEXT NOT NULL DEFAULT 'Medium',
    ac INTEGER NOT NULL DEFAULT 10,
    hp INTEGER NOT NULL DEFAULT 1,
    str INTEGER NOT NULL DEFAULT 10,
    dex INTEGER NOT NULL DEFAULT 10,
    con INTEGER NOT NULL DEFAULT 10,
    int_ INTEGER NOT NULL DEFAULT 10,
    wis INTEGER NOT NULL DEFAULT 10,
    cha INTEGER NOT NULL DEFAULT 10,
    cr TEXT NOT NULL DEFAULT '0',
    source TEXT NOT NULL DEFAULT 'SRD',
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
    description TEXT NOT NULL DEFAULT ''
);

-- Campaign NPC linking
CREATE TABLE IF NOT EXISTS campaign_npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    npc_id INTEGER NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(campaign_id, npc_id)
);

CREATE INDEX IF NOT EXISTS idx_campaign_npcs_campaign ON campaign_npcs(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_npcs_npc ON campaign_npcs(npc_id);
CREATE INDEX IF NOT EXISTS idx_compendium_monsters_name ON compendium_monsters(name);
CREATE INDEX IF NOT EXISTS idx_compendium_monsters_cr ON compendium_monsters(cr);
`,
	},
	{
		version: 32,
		sql: `
-- Scene Dialog tables
CREATE TABLE IF NOT EXISTS oneshot_scene_dialogs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scene_id INTEGER NOT NULL REFERENCES oneshot_scenes(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    speaker TEXT NOT NULL DEFAULT '',
    dialog_text TEXT NOT NULL DEFAULT '',
    dm_notes TEXT NOT NULL DEFAULT '',
    player_handout TEXT NOT NULL DEFAULT '',
    condition TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_scene_dialogs_scene ON oneshot_scene_dialogs(scene_id);
`,
	},
	{
		version: 33,
		sql: `
CREATE TABLE IF NOT EXISTS campaign_monster_roster (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    compendium_monster_id INTEGER NOT NULL REFERENCES compendium_monsters(id) ON DELETE CASCADE,
    library_monster_id INTEGER REFERENCES monster_library(id) ON DELETE SET NULL,
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
    source TEXT NOT NULL DEFAULT 'compendium',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(campaign_id, compendium_monster_id, library_monster_id)
);

CREATE INDEX IF NOT EXISTS idx_campaign_roster_campaign ON campaign_monster_roster(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_roster_compendium ON campaign_monster_roster(compendium_monster_id);
`,
	},
	{
		version: 34,
		sql: `
-- Dynamic compendium schema system
CREATE TABLE IF NOT EXISTS compendium_schemas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type_name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    fields TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS compendium_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schema_id INTEGER NOT NULL REFERENCES compendium_schemas(id) ON DELETE CASCADE,
    data TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS compendium_import_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','mapping','validated','previewed','completed','rolled_back','failed')),
    files TEXT NOT NULL DEFAULT '[]',
    mapping TEXT NOT NULL DEFAULT '{}',
    summary TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    rolled_back_at TEXT
);

CREATE VIRTUAL TABLE IF NOT EXISTS compendium_entries_fts USING fts5(
    data,
    content='compendium_entries',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS compendium_entries_ai AFTER INSERT ON compendium_entries BEGIN
    INSERT INTO compendium_entries_fts(rowid, data)
    VALUES (new.id, new.data);
END;

CREATE TRIGGER IF NOT EXISTS compendium_entries_ad AFTER DELETE ON compendium_entries BEGIN
    INSERT INTO compendium_entries_fts(compendium_entries_fts, rowid, data)
    VALUES ('delete', old.id, old.data);
END;

CREATE TRIGGER IF NOT EXISTS compendium_entries_au AFTER UPDATE ON compendium_entries BEGIN
    INSERT INTO compendium_entries_fts(compendium_entries_fts, rowid, data)
    VALUES ('delete', old.id, old.data);
    INSERT INTO compendium_entries_fts(rowid, data)
    VALUES (new.id, new.data);
END;

CREATE INDEX IF NOT EXISTS idx_compendium_entries_schema ON compendium_entries(schema_id);
CREATE INDEX IF NOT EXISTS idx_compendium_entries_created ON compendium_entries(created_at);
CREATE INDEX IF NOT EXISTS idx_compendium_import_logs_user ON compendium_import_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_compendium_import_logs_status ON compendium_import_logs(status);
`,
	},
	{
		version: 35,
		sql: `
CREATE TABLE IF NOT EXISTS umami_settings (
    id INTEGER PRIMARY KEY,
    enabled INTEGER NOT NULL DEFAULT 0,
    tracker_hostname TEXT NOT NULL DEFAULT '',
    website_id TEXT NOT NULL DEFAULT '',
    share_data INTEGER NOT NULL DEFAULT 0,
    enable_admin_tracking INTEGER NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO umami_settings (id, enabled) VALUES (1, 0);
`,
	},
	{
		version: 36,
		sql: `
CREATE TABLE IF NOT EXISTS otel_settings (
    id INTEGER PRIMARY KEY,
    endpoint TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO otel_settings (id, endpoint, enabled) VALUES (1, '', 0);
`,
	},
	{
		version: 37,
		sql: `
CREATE TABLE IF NOT EXISTS auth_sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    username TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'user',
    ip TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT NOT NULL
);
`,
	},
	{
		version: 38,
		sql: `
CREATE TABLE IF NOT EXISTS race_colors (
    race_name TEXT PRIMARY KEY,
    color TEXT NOT NULL DEFAULT '#6c757d'
);

INSERT OR IGNORE INTO race_colors (race_name, color) VALUES
    ('Human', '#4a90d9'),
    ('Elf', '#43b581'),
    ('Dwarf', '#f04747'),
    ('Halfling', '#faa61a'),
    ('Dragonborn', '#7289da'),
    ('Gnome', '#e91e63'),
    ('Half-Elf', '#9b59b6'),
    ('Half-Orc', '#e67e22'),
    ('Tiefling', '#c0392b'),
    ('Aarakocra', '#1abc9c'),
    ('Aasimar', '#f1c40f'),
    ('Bugbear', '#2c3e50'),
    ('Centaur', '#8e44ad'),
    ('Changeling', '#d35400'),
    ('Deep Gnome', '#7f8c8d'),
    ('Duergar', '#95a5a6'),
    ('Eladrin', '#3498db'),
    ('Fairy', '#e91e63'),
    ('Firbolg', '#27ae60'),
    ('Genasi', '#2980b9'),
    ('Gith', '#2c3e50'),
    ('Goblin', '#e74c3c'),
    ('Goliath', '#bdc3c7'),
    ('Harengon', '#f48fb1'),
    ('Kenku', '#795548'),
    ('Kobold', '#ff5722'),
    ('Leonin', '#ff9800'),
    ('Lizardfolk', '#4caf50'),
    ('Minotaur', '#3f51b5'),
    ('Orc', '#d32f2f'),
    ('Satyr', '#9c27b0'),
    ('Sea Elf', '#00bcd4'),
    ('Shadar-Kai', '#607d8b'),
    ('Shifter', '#ff6f00'),
    ('Tabaxi', '#ffc107'),
    ('Tortle', '#8d6e63'),
    ('Triton', '#00acc1'),
    ('Verdan', '#43a047'),
    ('Warforged', '#546e7a'),
    ('Yuan-Ti', '#1b5e20');
`,
	},
	{
		version: 39,
		sql: `
CREATE TABLE IF NOT EXISTS google_events_cache (
    event_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    start_time TEXT NOT NULL DEFAULT '',
    end_time TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    all_day INTEGER NOT NULL DEFAULT 0,
    cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events_settings (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    calendar_id TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    cache_ttl_seconds INTEGER NOT NULL DEFAULT 300,
    credentials_json TEXT NOT NULL DEFAULT ''
);
`,
	},
	{
		version: 40,
		sql: `
ALTER TABLE events_settings ADD COLUMN auth_method TEXT NOT NULL DEFAULT 'service_account';
ALTER TABLE events_settings ADD COLUMN oauth_client_id TEXT NOT NULL DEFAULT '';
ALTER TABLE events_settings ADD COLUMN oauth_client_secret TEXT NOT NULL DEFAULT '';
ALTER TABLE events_settings ADD COLUMN oauth_refresh_token TEXT NOT NULL DEFAULT '';
`,
	},
	{
		version: 41,
		sql: `
ALTER TABLE events_settings ADD COLUMN color_labels TEXT NOT NULL DEFAULT '';
ALTER TABLE events_settings ADD COLUMN filter_mode TEXT NOT NULL DEFAULT 'text';
`,
	},
	{
		version: 42,
		sql: `
CREATE TABLE IF NOT EXISTS campaign_event_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL DEFAULT 0,
    slug TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    calendar_id TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    color_labels TEXT NOT NULL DEFAULT '',
    filter_mode TEXT NOT NULL DEFAULT 'text',
    cache_ttl_seconds INTEGER NOT NULL DEFAULT 300,
    credentials_json TEXT NOT NULL DEFAULT '',
    auth_method TEXT NOT NULL DEFAULT 'service_account',
    oauth_client_id TEXT NOT NULL DEFAULT '',
    oauth_client_secret TEXT NOT NULL DEFAULT '',
    oauth_refresh_token TEXT NOT NULL DEFAULT '',
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_campaign_event_settings_slug ON campaign_event_settings(slug);
CREATE INDEX IF NOT EXISTS idx_campaign_event_settings_campaign ON campaign_event_settings(campaign_id);
`,
	},
	{
		version: 43,
		sql: `
ALTER TABLE google_events_cache ADD COLUMN campaign_slug TEXT NOT NULL DEFAULT '';
`,
	},
	{
		version: 44,
		sql: `
ALTER TABLE events_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api';
ALTER TABLE events_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT '';
ALTER TABLE campaign_event_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api';
ALTER TABLE campaign_event_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT '';
`,
	},
	{
		version: 45,
		sql: `
-- Universal entity links (polymorphic cross-entity references)
CREATE TABLE IF NOT EXISTS entity_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type TEXT NOT NULL,
    source_id INTEGER NOT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    context TEXT NOT NULL DEFAULT 'manual' CHECK(context IN ('manual','mention','import')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(source_type, source_id, target_type, target_id, context)
);
CREATE INDEX IF NOT EXISTS idx_entity_links_source ON entity_links(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_entity_links_target ON entity_links(target_type, target_id);

-- Unified full-text search index across all searchable entity types.
-- Standalone FTS5 table (not external-content). Sync triggers and backfill
-- are NOT created here: ent's auto-migration (Client.Schema.Create) recreates
-- ent-managed tables on startup, which would drop triggers on those tables.
-- db.EnsureSearchIndex() runs after ent migration and creates them instead.
CREATE VIRTUAL TABLE IF NOT EXISTS entity_search_index USING fts5(
    entity_type,
    entity_id UNINDEXED,
    title,
    subtitle,
    body,
    tokenize='porter unicode61'
);
`,
	},
	{
		version: 46,
		sql: `
-- Audit log for villum-transfer imports
CREATE TABLE IF NOT EXISTS transfer_import_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL DEFAULT '',
    doc_type TEXT NOT NULL DEFAULT '',
    counts TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'completed' CHECK(status IN ('completed','failed','rolled_back')),
    error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_transfer_import_logs_user ON transfer_import_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_transfer_import_logs_created ON transfer_import_logs(created_at);
`,
	},
	{
		version: 47,
		sql: `
-- Site-wide settings (key/value)
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '0'
);

INSERT OR IGNORE INTO app_settings (key, value) VALUES ('eink', '0');
`,
	},
	{
		version: 48,
		sql: `
-- Persist Google Calendar event color labels through the events cache so
-- swatches survive cache-served page loads.
ALTER TABLE google_events_cache ADD COLUMN color_id TEXT NOT NULL DEFAULT '';
`,
	},
}

func Migrate() error {
	var current int
	err := DB.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_version").Scan(&current)
	if err != nil {
		current = 0
	}

	migrated := false
	for _, m := range migrations {
		if m.version > current {
			migrated = true
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

	// Compact the database file after applying any pending migrations and
	// re-report page statistics (VACUUM cannot run inside a transaction).
	if migrated {
		if _, err := DB.Exec("VACUUM"); err != nil {
			return fmt.Errorf("vacuum after migrations: %w", err)
		}
		LogPageStats()
	}

	return nil
}

// LogPageStats reports current SQLite page statistics (shared with db.Init).
func LogPageStats() {
	var pageCount, freelistCount, pageSize int
	if err := DB.QueryRow("PRAGMA page_count").Scan(&pageCount); err != nil {
		return
	}
	_ = DB.QueryRow("PRAGMA freelist_count").Scan(&freelistCount)
	_ = DB.QueryRow("PRAGMA page_size").Scan(&pageSize)
	log.Printf("db: page_count=%d freelist_count=%d page_size=%d", pageCount, freelistCount, pageSize)
}

// ApplySafeAlters runs ALTER TABLE statements that safely add columns if they don't exist.
// This must run AFTER ent.Schema.Create() to avoid ent recreating tables and dropping extra columns.
func ApplySafeAlters() error {
	alterStatements := []string{
		"ALTER TABLE characters ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL",
		"ALTER TABLE characters ADD COLUMN character_type TEXT NOT NULL DEFAULT 'player'",
		// Composite indexes for hot query patterns (must run after Schema.Create
		// since ent owns these tables/columns).
		"CREATE INDEX IF NOT EXISTS idx_characters_user_name_level ON characters (user_id, name, level)",
		"CREATE INDEX IF NOT EXISTS idx_characters_campaign_name ON characters (campaign_id, name)",
		"CREATE INDEX IF NOT EXISTS idx_spells_char_level_name ON spells (character_id, level, name)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_char_category_name ON inventory (character_id, category, name)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_char_date ON sessions (character_id, session_date)",
		"CREATE INDEX IF NOT EXISTS idx_quests_char_status_updated ON quests (character_id, status, updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_journal_char_entry_date ON journal (character_id, entry_date)",
		"CREATE INDEX IF NOT EXISTS idx_combat_campaign_turn ON combat_entries (campaign_id, turn_order)",
		"CREATE INDEX IF NOT EXISTS idx_dice_rolls_user_timestamp ON dice_rolls (user_id, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_campaigns_user_name ON campaigns (user_id, name)",
		"UPDATE characters SET character_type='linked' WHERE campaign_id IS NOT NULL AND campaign_id != 0 AND EXISTS (SELECT 1 FROM campaign_members cm WHERE cm.campaign_id = characters.campaign_id AND cm.user_id != characters.user_id)",
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
		// Restructure app columns
		"ALTER TABLE uploads ADD COLUMN owner_type TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE uploads ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE factions ADD COLUMN party_id INTEGER REFERENCES parties(id) ON DELETE CASCADE",
		"ALTER TABLE shops ADD COLUMN oneshot_adventure_id INTEGER REFERENCES oneshot_adventures(id) ON DELETE SET NULL",
		"ALTER TABLE npcs ADD COLUMN is_full INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE npcs ADD COLUMN ac INTEGER NOT NULL DEFAULT 10",
		"ALTER TABLE npcs ADD COLUMN speed INTEGER NOT NULL DEFAULT 30",
		"ALTER TABLE npcs ADD COLUMN skills TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN saves TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN features TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN actions TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN backstory TEXT NOT NULL DEFAULT ''",
		// Campaign completeness columns
		"ALTER TABLE characters ADD COLUMN exhaustion_level INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE inventory ADD COLUMN is_identified INTEGER NOT NULL DEFAULT 1",
		"ALTER TABLE oneshot_acts ADD COLUMN notes TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE dm_notes ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE",
		// NPC linking for one-shots
		"ALTER TABLE oneshot_adventure_npcs ADD COLUMN story_hook TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE oneshot_adventure_npcs ADD COLUMN combat_ready INTEGER NOT NULL DEFAULT 0",
		// Mini-campaign mode for one-shots
		"ALTER TABLE oneshot_adventures ADD COLUMN is_mini_campaign INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE oneshot_adventures ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0",
		// One-shot scoping
		"ALTER TABLE campaign_timeline_events ADD COLUMN oneshot_adventure_id INTEGER REFERENCES oneshot_adventures(id) ON DELETE CASCADE",
		"ALTER TABLE factions ADD COLUMN oneshot_adventure_id INTEGER REFERENCES oneshot_adventures(id) ON DELETE CASCADE",
		// Campaign NPC linking
		"ALTER TABLE oneshot_monsters ADD COLUMN compendium_monster_id INTEGER DEFAULT NULL",
		// Compendium schema field additions
		"ALTER TABLE compendium_monsters ADD COLUMN alignment TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_monsters ADD COLUMN expansion TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_monsters ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN category TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_list INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_bonds TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_flaws TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_ideals TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_equipment TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_starting_gold INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_personality_traits TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_classes ADD COLUMN category TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_classes ADD COLUMN expansion TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_classes ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_equipment ADD COLUMN item_type TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_equipment ADD COLUMN item_rarity TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_equipment ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_races ADD COLUMN category TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_races ADD COLUMN expansion TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_races ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_spells ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		// Shop location support
		"ALTER TABLE shops ADD COLUMN location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL",
		// Compendium legacy table linking
		"ALTER TABLE spells ADD COLUMN compendium_spell_id INTEGER REFERENCES compendium_spells(id) ON DELETE SET NULL",
		"ALTER TABLE inventory ADD COLUMN compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL",
		// Compendium entry linking
		"ALTER TABLE encounter_monsters ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE oneshot_monsters ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE spells ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE inventory ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE oneshot_items ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE oneshot_items ADD COLUMN compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL",
		"ALTER TABLE character_features ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		// Indexes for compendium reference columns
		"CREATE INDEX IF NOT EXISTS idx_spells_compendium_spell_id ON spells(compendium_spell_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_compendium_equipment_id ON inventory(compendium_equipment_id)",
		"CREATE INDEX IF NOT EXISTS idx_oneshot_monsters_compendium_monster_id ON oneshot_monsters(compendium_monster_id)",
		"CREATE INDEX IF NOT EXISTS idx_oneshot_items_compendium_equipment_id ON oneshot_items(compendium_equipment_id)",
		// Events iCal URL source support
		"ALTER TABLE events_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api'",
		"ALTER TABLE events_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE campaign_event_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api'",
		"ALTER TABLE campaign_event_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT ''",
	}
	for _, stmt := range alterStatements {
		log.Printf("ALTER: %s", stmt)
		if _, err := DB.Exec(stmt); err != nil {
			log.Printf("ALTER error: %v", err)
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("alter table: %w", err)
			}
		}
	}
	return nil
}
