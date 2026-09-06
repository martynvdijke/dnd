package migrations

var migration052SQL = `
ALTER TABLE campaign_maps ADD COLUMN grid_units TEXT NOT NULL DEFAULT 'ft';
ALTER TABLE campaign_map_pins ADD COLUMN snap_to_grid INTEGER NOT NULL DEFAULT 0;
`
