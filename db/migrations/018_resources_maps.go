package migrations

var migration018SQL = `
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
`
