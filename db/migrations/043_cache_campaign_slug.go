package migrations

var migration043SQL = `
ALTER TABLE google_events_cache ADD COLUMN campaign_slug TEXT NOT NULL DEFAULT '';
`
