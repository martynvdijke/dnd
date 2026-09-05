package migrations

var migration045SQL = `
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
`
