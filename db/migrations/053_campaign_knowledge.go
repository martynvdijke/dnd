package migrations

var migration053SQL = `
CREATE TABLE IF NOT EXISTS campaign_knowledge (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'rumor' CHECK(status IN ('rumor','confirmed','revealed','false')),
    shared INTEGER NOT NULL DEFAULT 0,
    status_history TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_campaign_knowledge_campaign ON campaign_knowledge(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_knowledge_status ON campaign_knowledge(status);

CREATE TABLE IF NOT EXISTS campaign_knowledge_known_by (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    knowledge_id INTEGER NOT NULL REFERENCES campaign_knowledge(id) ON DELETE CASCADE,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(knowledge_id, character_id)
);
CREATE INDEX IF NOT EXISTS idx_knowledge_known_by_knowledge ON campaign_knowledge_known_by(knowledge_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_known_by_character ON campaign_knowledge_known_by(character_id);
`
