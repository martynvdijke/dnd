package migrations

var migration035SQL = `
CREATE TABLE IF NOT EXISTS umami_settings (
    id INTEGER PRIMARY KEY,
    enabled INTEGER NOT NULL DEFAULT 0,
    tracker_hostname TEXT NOT NULL DEFAULT '',
    website_id TEXT NOT NULL DEFAULT '',
    share_data INTEGER NOT NULL DEFAULT 0,
    enable_admin_tracking INTEGER NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO umami_settings (id, enabled) VALUES (1, 0);
`
