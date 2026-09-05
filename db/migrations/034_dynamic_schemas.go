package migrations

var migration034SQL = `
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
`
