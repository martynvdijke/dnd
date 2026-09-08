package migrations

var migration054SQL = `
ALTER TABLE auth_sessions ADD COLUMN auth_method TEXT NOT NULL DEFAULT 'password';
ALTER TABLE auth_sessions ADD COLUMN oidc_sub TEXT NOT NULL DEFAULT '';
`
