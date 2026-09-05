package migrations

var migration029SQL = `
-- Act tree: parent_act_id for nested sub-acts, sort_order for ordering
ALTER TABLE oneshot_acts ADD COLUMN parent_act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;
ALTER TABLE oneshot_acts ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Scene ordering with explicit sort_order
ALTER TABLE oneshot_scenes ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Act-level shops, items, encounters (NULL = adventure-level)
ALTER TABLE shops ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;
ALTER TABLE oneshot_items ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;
ALTER TABLE oneshot_adventure_encounters ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE;

-- Copy existing number values into sort_order for acts and scenes
UPDATE oneshot_acts SET sort_order = number;
UPDATE oneshot_scenes SET sort_order = number;
`
