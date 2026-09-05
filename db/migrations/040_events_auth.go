package migrations

var migration040SQL = `
ALTER TABLE events_settings ADD COLUMN auth_method TEXT NOT NULL DEFAULT 'service_account';
ALTER TABLE events_settings ADD COLUMN oauth_client_id TEXT NOT NULL DEFAULT '';
ALTER TABLE events_settings ADD COLUMN oauth_client_secret TEXT NOT NULL DEFAULT '';
ALTER TABLE events_settings ADD COLUMN oauth_refresh_token TEXT NOT NULL DEFAULT '';
`
