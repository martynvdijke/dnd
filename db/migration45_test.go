package db

import "testing"

// TestUnifiedSearchIndex verifies migration v45: the unified FTS5 index,
// its sync triggers, backfill behavior, and the entity_links table.
func TestUnifiedSearchIndex(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Pre-seed a user for FK constraints.
	if _, err := DB.Exec(`INSERT INTO users (username, password, display_name, role) VALUES ('idxuser', 'x', 'Idx', 'admin')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	t.Run("triggers sync on insert update delete", func(t *testing.T) {
		res, err := DB.Exec(`INSERT INTO npcs (user_id, name, description) VALUES (1, 'Emberhold Innkeeper', 'a jolly fellow')`)
		if err != nil {
			t.Fatalf("insert npc: %v", err)
		}
		npcID, _ := res.LastInsertId()

		var title string
		if err := DB.QueryRow(`SELECT title FROM entity_search_index WHERE entity_type='npc' AND entity_id=?`, npcID).Scan(&title); err != nil {
			t.Fatalf("index row after insert: %v", err)
		}
		if title != "Emberhold Innkeeper" {
			t.Errorf("unexpected indexed title %q", title)
		}

		if _, err := DB.Exec(`UPDATE npcs SET name='Renamed Keeper' WHERE id=?`, npcID); err != nil {
			t.Fatalf("update npc: %v", err)
		}
		if err := DB.QueryRow(`SELECT title FROM entity_search_index WHERE entity_type='npc' AND entity_id=?`, npcID).Scan(&title); err != nil {
			t.Fatalf("index row after update: %v", err)
		}
		if title != "Renamed Keeper" {
			t.Errorf("index not updated, got %q", title)
		}

		if _, err := DB.Exec(`DELETE FROM npcs WHERE id=?`, npcID); err != nil {
			t.Fatalf("delete npc: %v", err)
		}
		var n int
		if err := DB.QueryRow(`SELECT COUNT(*) FROM entity_search_index WHERE entity_type='npc' AND entity_id=?`, npcID).Scan(&n); err != nil {
			t.Fatalf("count after delete: %v", err)
		}
		if n != 0 {
			t.Errorf("expected index row removed after delete, got %d", n)
		}
	})

	t.Run("fts match finds body text", func(t *testing.T) {
		if _, err := DB.Exec(`INSERT INTO npcs (user_id, name, description) VALUES (1, 'Grimble', 'a sneaky goblin scout')`); err != nil {
			t.Fatalf("insert npc: %v", err)
		}
		var cnt int
		if err := DB.QueryRow(`SELECT COUNT(*) FROM entity_search_index WHERE entity_search_index MATCH 'goblin'`).Scan(&cnt); err != nil {
			t.Fatalf("fts match: %v", err)
		}
		if cnt < 1 {
			t.Errorf("expected FTS match for 'goblin', got %d", cnt)
		}
	})

	t.Run("index covers all registered entity types", func(t *testing.T) {
		// One row per searchable table; every type must land in the index.
		seed := []struct{ sql, typ string }{
			{`INSERT INTO characters (user_id, name, backstory) VALUES (1, 'Aria', 'orphaned noble')`, "character"},
			{`INSERT INTO character_notes (character_id, title, content) VALUES (1, 'Secret', 'hidden passage')`, "note"},
			{`INSERT INTO quests (character_id, name, description) VALUES (1, 'Find relic', 'long quest')`, "quest"},
			{`INSERT INTO journal (character_id, title, entry) VALUES (1, 'Day 1', 'we set out')`, "journal"},
			{`INSERT INTO sessions (character_id, title, notes) VALUES (1, 'Session 0', 'met party')`, "session"},
			{`INSERT INTO campaigns (user_id, name, description) VALUES (1, 'Dragonfall', 'epic tale')`, "campaign"},
			{`INSERT INTO locations (user_id, name, description) VALUES (1, 'Emberhold', 'river town')`, "location"},
			{`INSERT INTO encounter_templates (user_id, name, description) VALUES (1, 'Ambush', 'goblin ambush')`, "encounter"},
			{`INSERT INTO monster_library (user_id, name, description) VALUES (1, 'Griffon', 'flying beast')`, "monster"},
			{`INSERT INTO shops (user_id, name, description) VALUES (1, 'General Store', 'sells things')`, "shop"},
			{`INSERT INTO factions (campaign_id, name, description) VALUES (1, 'Harpers', 'secret society')`, "faction"},
			{`INSERT INTO oneshot_adventures (user_id, title, premise) VALUES (1, 'Lost Mine', 'a dwarven mystery')`, "adventure"},
			{`INSERT INTO campaign_wiki_pages (campaign_id, user_id, title, content) VALUES (1, 1, 'World lore', 'the realm')`, "wiki"},
			{`INSERT INTO campaign_timeline_events (campaign_id, title, description, event_date) VALUES (1, 'Founding', 'town founded', '101-01-01')`, "timeline"},
			{`INSERT INTO oneshot_items (adventure_id, name, description) VALUES (1, 'Ruby', 'a glowing gem')`, "item"},
		}
		for _, s := range seed {
			if _, err := DB.Exec(s.sql); err != nil {
				t.Fatalf("seed %s: %v", s.typ, err)
			}
		}
		if _, err := DB.Exec(`INSERT INTO compendium_schemas (type_name, display_name, fields) VALUES ('testtype', 'Test Type', '[]')`); err != nil {
			t.Fatalf("seed schema: %v", err)
		}
		if _, err := DB.Exec(`INSERT INTO compendium_entries (schema_id, data) VALUES (1, '{"name":"Widget","desc":"a test widget"}')`); err != nil {
			t.Fatalf("seed compendium: %v", err)
		}

		for _, s := range seed {
			var n int
			if err := DB.QueryRow(`SELECT COUNT(*) FROM entity_search_index WHERE entity_type=?`, s.typ).Scan(&n); err != nil {
				t.Fatalf("count %s: %v", s.typ, err)
			}
			if n < 1 {
				t.Errorf("entity type %q missing from unified index", s.typ)
			}
		}
		var n int
		if err := DB.QueryRow(`SELECT COUNT(*) FROM entity_search_index WHERE entity_type='compendium' AND title='Widget'`).Scan(&n); err != nil {
			t.Fatalf("count compendium: %v", err)
		}
		if n != 1 {
			t.Errorf("expected compendium entry indexed with json name, got %d", n)
		}
	})

	t.Run("link cleanup trigger fires on entity delete", func(t *testing.T) {
		if _, err := DB.Exec(`INSERT INTO entity_links (source_type, source_id, target_type, target_id, context) VALUES ('npc', 2, 'location', 1, 'manual')`); err != nil {
			t.Fatalf("insert link: %v", err)
		}
		if _, err := DB.Exec(`INSERT INTO entity_links (source_type, source_id, target_type, target_id, context) VALUES ('quest', 1, 'npc', 2, 'mention')`); err != nil {
			t.Fatalf("insert link2: %v", err)
		}
		// Delete the NPC with id 2 (Grimble from earlier subtest).
		if _, err := DB.Exec(`DELETE FROM npcs WHERE id=2`); err != nil {
			t.Fatalf("delete npc: %v", err)
		}
		var n int
		if err := DB.QueryRow(`SELECT COUNT(*) FROM entity_links WHERE (source_type='npc' AND source_id=2) OR (target_type='npc' AND target_id=2)`).Scan(&n); err != nil {
			t.Fatalf("count links: %v", err)
		}
		if n != 0 {
			t.Errorf("expected links removed after entity delete, got %d", n)
		}
	})

	t.Run("entity_links table shape", func(t *testing.T) {
		res, err := DB.Exec(`INSERT INTO entity_links (source_type, source_id, target_type, target_id, context) VALUES ('quest', 1, 'location', 1, 'manual')`)
		if err != nil {
			t.Fatalf("insert link: %v", err)
		}
		if id, _ := res.LastInsertId(); id < 1 {
			t.Errorf("expected link id >= 1, got %d", id)
		}
		// Duplicate (source,target,context) must be rejected.
		if _, err := DB.Exec(`INSERT INTO entity_links (source_type, source_id, target_type, target_id, context) VALUES ('quest', 1, 'location', 1, 'manual')`); err == nil {
			t.Error("expected uniqueness violation for duplicate link")
		}
		// Bad context must be rejected.
		if _, err := DB.Exec(`INSERT INTO entity_links (source_type, source_id, target_type, target_id, context) VALUES ('quest', 1, 'location', 1, 'bogus')`); err == nil {
			t.Error("expected CHECK violation for bad context")
		}
	})
}
