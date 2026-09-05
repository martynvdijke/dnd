package migrations

var migration014SQL = `
ALTER TABLE campaigns ADD COLUMN party_name TEXT NOT NULL DEFAULT '';
`
