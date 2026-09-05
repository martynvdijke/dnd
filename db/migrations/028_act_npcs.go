package migrations

var migration028SQL = `
-- Act-level planning: notes, NPCs, and DM notes per act
CREATE TABLE IF NOT EXISTS oneshot_act_npcs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    act_id INTEGER NOT NULL REFERENCES oneshot_acts(id) ON DELETE CASCADE,
    npc_id INTEGER REFERENCES npcs(id) ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    is_inline INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_oneshot_act_npcs_act ON oneshot_act_npcs(act_id);
CREATE INDEX IF NOT EXISTS idx_oneshot_act_npcs_npc ON oneshot_act_npcs(npc_id);
`
