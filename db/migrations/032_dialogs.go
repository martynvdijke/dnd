package migrations

var migration032SQL = `
-- Scene Dialog tables
CREATE TABLE IF NOT EXISTS oneshot_scene_dialogs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scene_id INTEGER NOT NULL REFERENCES oneshot_scenes(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    speaker TEXT NOT NULL DEFAULT '',
    dialog_text TEXT NOT NULL DEFAULT '',
    dm_notes TEXT NOT NULL DEFAULT '',
    player_handout TEXT NOT NULL DEFAULT '',
    condition TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_scene_dialogs_scene ON oneshot_scene_dialogs(scene_id);
`
