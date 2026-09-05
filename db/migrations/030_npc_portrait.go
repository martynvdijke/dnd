package migrations

var migration030SQL = `
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
`
