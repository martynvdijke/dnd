package migrations

import (
	"database/sql"
	"villum/middleware"
)

// RebuildNPCItemLinksIfNeeded recreates npc_item_links with a nullable item_id.
func RebuildNPCItemLinksIfNeeded(db *sql.DB) error {
	var notNull int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('npc_item_links') WHERE name='item_id' AND "notnull"=1`).Scan(&notNull)
	if err != nil {
		return err
	}
	if notNull == 0 {
		return nil
	}
	middleware.LogInfo("migration", "rebuilding npc_item_links for compendium links")
	statements := []string{
		`CREATE TABLE npc_item_links_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			npc_id INTEGER NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
			adventure_id INTEGER NOT NULL REFERENCES oneshot_adventures(id) ON DELETE CASCADE,
			item_id INTEGER REFERENCES oneshot_items(id) ON DELETE CASCADE,
			compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL,
			relationship_type TEXT NOT NULL DEFAULT 'owns',
			notes TEXT NOT NULL DEFAULT '',
			UNIQUE(npc_id, item_id)
		)`,
		`INSERT INTO npc_item_links_new(id, npc_id, adventure_id, item_id, compendium_equipment_id, relationship_type, notes)
			SELECT id, npc_id, adventure_id, item_id, NULL, relationship_type, notes FROM npc_item_links`,
		`DROP TABLE npc_item_links`,
		`ALTER TABLE npc_item_links_new RENAME TO npc_item_links`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
