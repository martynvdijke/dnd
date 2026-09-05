package migrations

import (
	"database/sql"
	"fmt"
	"strings"
	"villum/middleware"
)

// ApplySafeAlters runs ALTER TABLE statements that safely add columns if they don't exist.
func ApplySafeAlters(db *sql.DB) error {
	alterStatements := []string{
		"ALTER TABLE characters ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL",
		"ALTER TABLE characters ADD COLUMN character_type TEXT NOT NULL DEFAULT 'player'",
		// Composite indexes for hot query patterns (must run after Schema.Create
		// since ent owns these tables/columns).
		"CREATE INDEX IF NOT EXISTS idx_characters_user_name_level ON characters (user_id, name, level)",
		"CREATE INDEX IF NOT EXISTS idx_characters_campaign_name ON characters (campaign_id, name)",
		"CREATE INDEX IF NOT EXISTS idx_spells_char_level_name ON spells (character_id, level, name)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_char_category_name ON inventory (character_id, category, name)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_char_date ON sessions (character_id, session_date)",
		"CREATE INDEX IF NOT EXISTS idx_quests_char_status_updated ON quests (character_id, status, updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_journal_char_entry_date ON journal (character_id, entry_date)",
		"CREATE INDEX IF NOT EXISTS idx_combat_campaign_turn ON combat_entries (campaign_id, turn_order)",
		"CREATE INDEX IF NOT EXISTS idx_dice_rolls_user_timestamp ON dice_rolls (user_id, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_campaigns_user_name ON campaigns (user_id, name)",
		"UPDATE characters SET character_type='linked' WHERE campaign_id IS NOT NULL AND campaign_id != 0 AND EXISTS (SELECT 1 FROM campaign_members cm WHERE cm.campaign_id = characters.campaign_id AND cm.user_id != characters.user_id)",
		"ALTER TABLE character_npcs ADD COLUMN interaction_count INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE character_npcs ADD COLUMN last_interacted TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE characters ADD COLUMN death_saves_successes INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE characters ADD COLUMN death_saves_failures INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE characters ADD COLUMN concentrating_on TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE backup_settings ADD COLUMN interval_days INTEGER NOT NULL DEFAULT 7",
		"ALTER TABLE backup_settings ADD COLUMN keep_count INTEGER NOT NULL DEFAULT 7",
		"ALTER TABLE campaign_members ADD COLUMN role TEXT NOT NULL DEFAULT 'player'",
		"ALTER TABLE characters ADD COLUMN portrait_url TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE characters ADD COLUMN dm_notes TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE characters ADD COLUMN hp_auto_calc INTEGER NOT NULL DEFAULT 0",
		// Restructure app columns
		"ALTER TABLE uploads ADD COLUMN owner_type TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE uploads ADD COLUMN owner_id INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE factions ADD COLUMN party_id INTEGER REFERENCES parties(id) ON DELETE CASCADE",
		"ALTER TABLE shops ADD COLUMN oneshot_adventure_id INTEGER REFERENCES oneshot_adventures(id) ON DELETE SET NULL",
		"ALTER TABLE npcs ADD COLUMN is_full INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE npcs ADD COLUMN ac INTEGER NOT NULL DEFAULT 10",
		"ALTER TABLE npcs ADD COLUMN speed INTEGER NOT NULL DEFAULT 30",
		"ALTER TABLE npcs ADD COLUMN skills TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN saves TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN features TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN actions TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npcs ADD COLUMN backstory TEXT NOT NULL DEFAULT ''",
		// Campaign completeness columns
		"ALTER TABLE characters ADD COLUMN exhaustion_level INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE inventory ADD COLUMN is_identified INTEGER NOT NULL DEFAULT 1",
		"ALTER TABLE oneshot_acts ADD COLUMN notes TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE dm_notes ADD COLUMN act_id INTEGER REFERENCES oneshot_acts(id) ON DELETE CASCADE",
		// NPC linking for one-shots
		"ALTER TABLE oneshot_adventure_npcs ADD COLUMN story_hook TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE oneshot_adventure_npcs ADD COLUMN combat_ready INTEGER NOT NULL DEFAULT 0",
		// Mini-campaign mode for one-shots
		"ALTER TABLE oneshot_adventures ADD COLUMN is_mini_campaign INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE oneshot_adventures ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0",
		// One-shot scoping
		"ALTER TABLE campaign_timeline_events ADD COLUMN oneshot_adventure_id INTEGER REFERENCES oneshot_adventures(id) ON DELETE CASCADE",
		"ALTER TABLE factions ADD COLUMN oneshot_adventure_id INTEGER REFERENCES oneshot_adventures(id) ON DELETE CASCADE",
		// Campaign NPC linking
		"ALTER TABLE oneshot_monsters ADD COLUMN compendium_monster_id INTEGER DEFAULT NULL",
		// Compendium schema field additions
		"ALTER TABLE compendium_monsters ADD COLUMN alignment TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_monsters ADD COLUMN expansion TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_monsters ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN category TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_list INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_bonds TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_flaws TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_ideals TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_equipment TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_starting_gold INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE compendium_backgrounds ADD COLUMN data_personality_traits TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_backgrounds ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_classes ADD COLUMN category TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_classes ADD COLUMN expansion TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_classes ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_equipment ADD COLUMN item_type TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_equipment ADD COLUMN item_rarity TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_equipment ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_races ADD COLUMN category TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_races ADD COLUMN expansion TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_races ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE compendium_spells ADD COLUMN publisher TEXT NOT NULL DEFAULT ''",
		// Shop location support
		"ALTER TABLE shops ADD COLUMN location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL",
		// Compendium legacy table linking
		"ALTER TABLE spells ADD COLUMN compendium_spell_id INTEGER REFERENCES compendium_spells(id) ON DELETE SET NULL",
		"ALTER TABLE inventory ADD COLUMN compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL",
		// Compendium entry linking
		"ALTER TABLE encounter_monsters ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE oneshot_monsters ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE spells ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE inventory ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE oneshot_items ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		"ALTER TABLE oneshot_items ADD COLUMN compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL",
		"ALTER TABLE character_features ADD COLUMN compendium_entry_id INTEGER REFERENCES compendium_entries(id) ON DELETE SET NULL",
		// Indexes for compendium reference columns
		"CREATE INDEX IF NOT EXISTS idx_spells_compendium_spell_id ON spells(compendium_spell_id)",
		"CREATE INDEX IF NOT EXISTS idx_inventory_compendium_equipment_id ON inventory(compendium_equipment_id)",
		"CREATE INDEX IF NOT EXISTS idx_oneshot_monsters_compendium_monster_id ON oneshot_monsters(compendium_monster_id)",
		"CREATE INDEX IF NOT EXISTS idx_oneshot_items_compendium_equipment_id ON oneshot_items(compendium_equipment_id)",
		// Shop/NPC/character compendium linking
		"ALTER TABLE shop_items ADD COLUMN compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL",
		"ALTER TABLE shop_items ADD COLUMN weight REAL NOT NULL DEFAULT 0",
		"ALTER TABLE shop_items ADD COLUMN item_rarity TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE npc_item_links ADD COLUMN compendium_equipment_id INTEGER REFERENCES compendium_equipment(id) ON DELETE SET NULL",
		"ALTER TABLE characters ADD COLUMN compendium_race_id INTEGER REFERENCES compendium_races(id) ON DELETE SET NULL",
		"ALTER TABLE characters ADD COLUMN compendium_class_id INTEGER REFERENCES compendium_classes(id) ON DELETE SET NULL",
		"ALTER TABLE characters ADD COLUMN compendium_background_id INTEGER REFERENCES compendium_backgrounds(id) ON DELETE SET NULL",
		"CREATE INDEX IF NOT EXISTS idx_shop_items_compendium_equipment_id ON shop_items(compendium_equipment_id)",
		"CREATE INDEX IF NOT EXISTS idx_npc_item_links_compendium_equipment_id ON npc_item_links(compendium_equipment_id)",
		"CREATE INDEX IF NOT EXISTS idx_characters_compendium_race_id ON characters(compendium_race_id)",
		"CREATE INDEX IF NOT EXISTS idx_characters_compendium_class_id ON characters(compendium_class_id)",
		"CREATE INDEX IF NOT EXISTS idx_characters_compendium_background_id ON characters(compendium_background_id)",
		// Events iCal URL source support
		"ALTER TABLE events_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api'",
		"ALTER TABLE events_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE campaign_event_settings ADD COLUMN source_type TEXT NOT NULL DEFAULT 'google_api'",
		"ALTER TABLE campaign_event_settings ADD COLUMN ical_url TEXT NOT NULL DEFAULT ''",
	}

	for _, stmt := range alterStatements {
		middleware.LogInfo("migration", "ALTER", "statement", stmt)
		if _, err := db.Exec(stmt); err != nil {
			middleware.LogWarn("migration", "ALTER error", "error", err)
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("alter table: %w", err)
			}
		}
	}
	if err := RebuildNPCItemLinksIfNeeded(db); err != nil {
		return fmt.Errorf("rebuild npc_item_links: %w", err)
	}
	return nil
}
