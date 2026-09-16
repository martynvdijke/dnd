package migrations

var migration058SQL = `
CREATE TABLE IF NOT EXISTS log_settings (key TEXT PRIMARY KEY, value TEXT);
`
