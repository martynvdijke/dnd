package migrations

var migration044SQL = `
ALTER TABLE events_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api';
ALTER TABLE events_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT '';
ALTER TABLE campaign_event_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api';
ALTER TABLE campaign_event_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT '';
`
