package migrations

var migration024SQL = `
CREATE TABLE IF NOT EXISTS prep_checklist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    item TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general',
    is_checked INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0
);
`
