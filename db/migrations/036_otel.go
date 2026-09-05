package migrations

var migration036SQL = `
CREATE TABLE IF NOT EXISTS otel_settings (
    id INTEGER PRIMARY KEY,
    endpoint TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO otel_settings (id, endpoint, enabled) VALUES (1, '', 0);
`
