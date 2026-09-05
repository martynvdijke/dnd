package migrations

var migration041SQL = `
ALTER TABLE events_settings ADD COLUMN color_labels TEXT NOT NULL DEFAULT '';
ALTER TABLE events_settings ADD COLUMN filter_mode TEXT NOT NULL DEFAULT 'text';
`
