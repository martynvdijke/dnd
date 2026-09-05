package migrations

var migration010SQL = `
ALTER TABLE campaign_members ADD COLUMN role TEXT NOT NULL DEFAULT 'player' CHECK(role IN ('dm','player'));
`
