package migrations

var migration048SQL = `
-- Persist Google Calendar event color labels through the events cache so
-- swatches survive cache-served page loads.
ALTER TABLE google_events_cache ADD COLUMN color_id TEXT NOT NULL DEFAULT '';
`
