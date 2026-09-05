package migrations

var migration015SQL = `
CREATE TABLE IF NOT EXISTS crafting_recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'other' CHECK(category IN ('potion','scroll','magic-item','poison','other')),
    difficulty_dc INTEGER NOT NULL DEFAULT 10,
    crafting_time_hours REAL NOT NULL DEFAULT 1,
    required_tools TEXT NOT NULL DEFAULT '[]',
    required_materials TEXT NOT NULL DEFAULT '[]',
    result_item_name TEXT NOT NULL,
    result_item_category TEXT NOT NULL DEFAULT 'other',
    result_quantity INTEGER NOT NULL DEFAULT 1,
    result_description TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS character_crafting (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    recipe_id INTEGER REFERENCES crafting_recipes(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    progress_hours REAL NOT NULL DEFAULT 0,
    total_hours_required REAL NOT NULL DEFAULT 1,
    dc INTEGER NOT NULL DEFAULT 10,
    status TEXT NOT NULL DEFAULT 'in-progress' CHECK(status IN ('in-progress','complete','abandoned')),
    materials_allocated TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    started_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_crafting_recipes_user ON crafting_recipes(user_id);
CREATE INDEX IF NOT EXISTS idx_char_crafting_char ON character_crafting(character_id);
`
