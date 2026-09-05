package migrations

var migration008SQL = `
ALTER TABLE backup_settings ADD COLUMN interval_days INTEGER NOT NULL DEFAULT 7;
ALTER TABLE backup_settings ADD COLUMN keep_count INTEGER NOT NULL DEFAULT 7;
UPDATE backup_settings SET interval_days = MAX(1, COALESCE(interval_hours, 168) / 24) WHERE id = 1;
`
