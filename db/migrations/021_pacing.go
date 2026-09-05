package migrations

var migration021SQL = `
-- Session Pacing tables
CREATE TABLE IF NOT EXISTS session_pacing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
    current_act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE SET NULL,
    current_scene_id INTEGER REFERENCES oneshot_scenes(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'running' CHECK(status IN ('running','paused','completed')),
    elapsed_seconds INTEGER NOT NULL DEFAULT 0,
    started_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT
);

CREATE TABLE IF NOT EXISTS scene_timings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES session_pacing(id) ON DELETE CASCADE,
    scene_id INTEGER NOT NULL REFERENCES oneshot_scenes(id) ON DELETE CASCADE,
    elapsed_seconds INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','completed')),
    started_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT
);
`
