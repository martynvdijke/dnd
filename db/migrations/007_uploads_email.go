package migrations

var migration007SQL = `
CREATE TABLE IF NOT EXISTS uploads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT UNIQUE,
    ext TEXT,
    url TEXT,
    resized_url TEXT,
    thumbnail_url TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS email_settings (
    id INTEGER PRIMARY KEY,
    smtp_host TEXT,
    smtp_port INTEGER DEFAULT 587,
    username TEXT,
    password TEXT,
    from_addr TEXT,
    enabled INTEGER DEFAULT 0
);

INSERT OR IGNORE INTO email_settings (id, enabled) VALUES (1, 0);
`
