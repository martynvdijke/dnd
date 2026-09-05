package migrations

var migration042SQL = `
CREATE TABLE IF NOT EXISTS campaign_event_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL DEFAULT 0,
    slug TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    calendar_id TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    color_labels TEXT NOT NULL DEFAULT '',
    filter_mode TEXT NOT NULL DEFAULT 'text',
    cache_ttl_seconds INTEGER NOT NULL DEFAULT 300,
    credentials_json TEXT NOT NULL DEFAULT '',
    auth_method TEXT NOT NULL DEFAULT 'service_account',
    oauth_client_id TEXT NOT NULL DEFAULT '',
    oauth_client_secret TEXT NOT NULL DEFAULT '',
    oauth_refresh_token TEXT NOT NULL DEFAULT '',
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_campaign_event_settings_slug ON campaign_event_settings(slug);
CREATE INDEX IF NOT EXISTS idx_campaign_event_settings_campaign ON campaign_event_settings(campaign_id);
`
