package migrations

var migration012SQL = `
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
`
