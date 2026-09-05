package migrations

var migration031SQL = `
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
`
