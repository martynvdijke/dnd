package migrations

var migration039SQL = `
CREATE TABLE IF NOT EXISTS google_events_cache (
    event_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    start_time TEXT NOT NULL DEFAULT '',
    end_time TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    all_day INTEGER NOT NULL DEFAULT 0,
    cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events_settings (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    calendar_id TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    cache_ttl_seconds INTEGER NOT NULL DEFAULT 300,
    credentials_json TEXT NOT NULL DEFAULT ''
);
`
