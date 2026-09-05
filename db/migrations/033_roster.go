package migrations

var migration033SQL = `
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
`
