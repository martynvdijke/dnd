package migrations

var migration011SQL = `
ALTER TABLE compendium_races ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_races ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_classes ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_classes ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_spells ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_spells ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_feats ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_feats ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_backgrounds ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_backgrounds ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
ALTER TABLE compendium_equipment ADD COLUMN system TEXT NOT NULL DEFAULT 'dnd5e';
ALTER TABLE compendium_equipment ADD COLUMN source TEXT NOT NULL DEFAULT 'srd';
`
