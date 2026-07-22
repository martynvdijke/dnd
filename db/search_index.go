package db

import (
	"fmt"
	"log"
)

// searchIndexTriggers syncs the unified entity_search_index FTS5 table with
// all searchable entity tables. These MUST be created after ent's
// auto-migration (Client.Schema.Create), which recreates ent-managed tables
// and would silently drop any triggers attached to them. EnsureSearchIndex
// runs at every startup (idempotent), so the index self-heals.
const searchIndexTriggers = `
-- characters
CREATE TRIGGER IF NOT EXISTS esi_characters_ai AFTER INSERT ON characters BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('character', new.id, new.name, trim(new.race || ' ' || new.class), new.backstory);
END;
CREATE TRIGGER IF NOT EXISTS esi_characters_ad AFTER DELETE ON characters BEGIN
    DELETE FROM entity_search_index WHERE entity_type='character' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_characters_au AFTER UPDATE ON characters BEGIN
    DELETE FROM entity_search_index WHERE entity_type='character' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('character', new.id, new.name, trim(new.race || ' ' || new.class), new.backstory);
END;

-- npcs
CREATE TRIGGER IF NOT EXISTS esi_npcs_ai AFTER INSERT ON npcs BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('npc', new.id, new.name, trim(new.race || ' ' || new.class), new.description || ' ' || new.notes);
END;
CREATE TRIGGER IF NOT EXISTS esi_npcs_ad AFTER DELETE ON npcs BEGIN
    DELETE FROM entity_search_index WHERE entity_type='npc' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_npcs_au AFTER UPDATE ON npcs BEGIN
    DELETE FROM entity_search_index WHERE entity_type='npc' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('npc', new.id, new.name, trim(new.race || ' ' || new.class), new.description || ' ' || new.notes);
END;

-- character_notes (entity type: note)
CREATE TRIGGER IF NOT EXISTS esi_notes_ai AFTER INSERT ON character_notes BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('note', new.id, new.title, new.category, new.content);
END;
CREATE TRIGGER IF NOT EXISTS esi_notes_ad AFTER DELETE ON character_notes BEGIN
    DELETE FROM entity_search_index WHERE entity_type='note' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_notes_au AFTER UPDATE ON character_notes BEGIN
    DELETE FROM entity_search_index WHERE entity_type='note' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('note', new.id, new.title, new.category, new.content);
END;

-- quests
CREATE TRIGGER IF NOT EXISTS esi_quests_ai AFTER INSERT ON quests BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('quest', new.id, new.name, new.status, new.description || ' ' || new.objectives || ' ' || new.notes);
END;
CREATE TRIGGER IF NOT EXISTS esi_quests_ad AFTER DELETE ON quests BEGIN
    DELETE FROM entity_search_index WHERE entity_type='quest' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_quests_au AFTER UPDATE ON quests BEGIN
    DELETE FROM entity_search_index WHERE entity_type='quest' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('quest', new.id, new.name, new.status, new.description || ' ' || new.objectives || ' ' || new.notes);
END;

-- journal
CREATE TRIGGER IF NOT EXISTS esi_journal_ai AFTER INSERT ON journal BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('journal', new.id, new.title, new.entry_date, new.entry);
END;
CREATE TRIGGER IF NOT EXISTS esi_journal_ad AFTER DELETE ON journal BEGIN
    DELETE FROM entity_search_index WHERE entity_type='journal' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_journal_au AFTER UPDATE ON journal BEGIN
    DELETE FROM entity_search_index WHERE entity_type='journal' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('journal', new.id, new.title, new.entry_date, new.entry);
END;

-- sessions
CREATE TRIGGER IF NOT EXISTS esi_sessions_ai AFTER INSERT ON sessions BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('session', new.id, new.title, new.session_date, new.notes || ' ' || new.important_events);
END;
CREATE TRIGGER IF NOT EXISTS esi_sessions_ad AFTER DELETE ON sessions BEGIN
    DELETE FROM entity_search_index WHERE entity_type='session' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_sessions_au AFTER UPDATE ON sessions BEGIN
    DELETE FROM entity_search_index WHERE entity_type='session' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('session', new.id, new.title, new.session_date, new.notes || ' ' || new.important_events);
END;

-- campaigns
CREATE TRIGGER IF NOT EXISTS esi_campaigns_ai AFTER INSERT ON campaigns BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('campaign', new.id, new.name, new.party_name, new.description);
END;
CREATE TRIGGER IF NOT EXISTS esi_campaigns_ad AFTER DELETE ON campaigns BEGIN
    DELETE FROM entity_search_index WHERE entity_type='campaign' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_campaigns_au AFTER UPDATE ON campaigns BEGIN
    DELETE FROM entity_search_index WHERE entity_type='campaign' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('campaign', new.id, new.name, new.party_name, new.description);
END;

-- locations
CREATE TRIGGER IF NOT EXISTS esi_locations_ai AFTER INSERT ON locations BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('location', new.id, new.name, new.type, new.description);
END;
CREATE TRIGGER IF NOT EXISTS esi_locations_ad AFTER DELETE ON locations BEGIN
    DELETE FROM entity_search_index WHERE entity_type='location' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_locations_au AFTER UPDATE ON locations BEGIN
    DELETE FROM entity_search_index WHERE entity_type='location' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('location', new.id, new.name, new.type, new.description);
END;

-- encounter_templates (entity type: encounter)
CREATE TRIGGER IF NOT EXISTS esi_encounters_ai AFTER INSERT ON encounter_templates BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('encounter', new.id, new.name, new.difficulty, new.description || ' ' || new.environment || ' ' || new.notes);
END;
CREATE TRIGGER IF NOT EXISTS esi_encounters_ad AFTER DELETE ON encounter_templates BEGIN
    DELETE FROM entity_search_index WHERE entity_type='encounter' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_encounters_au AFTER UPDATE ON encounter_templates BEGIN
    DELETE FROM entity_search_index WHERE entity_type='encounter' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('encounter', new.id, new.name, new.difficulty, new.description || ' ' || new.environment || ' ' || new.notes);
END;

-- monster_library (entity type: monster)
CREATE TRIGGER IF NOT EXISTS esi_monsters_ai AFTER INSERT ON monster_library BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('monster', new.id, new.name, 'CR ' || new.cr, new.description || ' ' || new.special_abilities || ' ' || new.actions);
END;
CREATE TRIGGER IF NOT EXISTS esi_monsters_ad AFTER DELETE ON monster_library BEGIN
    DELETE FROM entity_search_index WHERE entity_type='monster' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_monsters_au AFTER UPDATE ON monster_library BEGIN
    DELETE FROM entity_search_index WHERE entity_type='monster' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('monster', new.id, new.name, 'CR ' || new.cr, new.description || ' ' || new.special_abilities || ' ' || new.actions);
END;

-- shops
CREATE TRIGGER IF NOT EXISTS esi_shops_ai AFTER INSERT ON shops BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('shop', new.id, new.name, '', new.description);
END;
CREATE TRIGGER IF NOT EXISTS esi_shops_ad AFTER DELETE ON shops BEGIN
    DELETE FROM entity_search_index WHERE entity_type='shop' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_shops_au AFTER UPDATE ON shops BEGIN
    DELETE FROM entity_search_index WHERE entity_type='shop' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('shop', new.id, new.name, '', new.description);
END;

-- factions
CREATE TRIGGER IF NOT EXISTS esi_factions_ai AFTER INSERT ON factions BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('faction', new.id, new.name, new.type, new.description);
END;
CREATE TRIGGER IF NOT EXISTS esi_factions_ad AFTER DELETE ON factions BEGIN
    DELETE FROM entity_search_index WHERE entity_type='faction' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_factions_au AFTER UPDATE ON factions BEGIN
    DELETE FROM entity_search_index WHERE entity_type='faction' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('faction', new.id, new.name, new.type, new.description);
END;

-- oneshot_adventures (entity type: adventure)
CREATE TRIGGER IF NOT EXISTS esi_adventures_ai AFTER INSERT ON oneshot_adventures BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('adventure', new.id, new.title, new.difficulty, new.premise || ' ' || new.hook || ' ' || new.notes);
END;
CREATE TRIGGER IF NOT EXISTS esi_adventures_ad AFTER DELETE ON oneshot_adventures BEGIN
    DELETE FROM entity_search_index WHERE entity_type='adventure' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_adventures_au AFTER UPDATE ON oneshot_adventures BEGIN
    DELETE FROM entity_search_index WHERE entity_type='adventure' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('adventure', new.id, new.title, new.difficulty, new.premise || ' ' || new.hook || ' ' || new.notes);
END;

-- campaign_wiki_pages (entity type: wiki)
CREATE TRIGGER IF NOT EXISTS esi_wiki_ai AFTER INSERT ON campaign_wiki_pages BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('wiki', new.id, new.title, '', new.content);
END;
CREATE TRIGGER IF NOT EXISTS esi_wiki_ad AFTER DELETE ON campaign_wiki_pages BEGIN
    DELETE FROM entity_search_index WHERE entity_type='wiki' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_wiki_au AFTER UPDATE ON campaign_wiki_pages BEGIN
    DELETE FROM entity_search_index WHERE entity_type='wiki' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('wiki', new.id, new.title, '', new.content);
END;

-- campaign_timeline_events (entity type: timeline)
CREATE TRIGGER IF NOT EXISTS esi_timeline_ai AFTER INSERT ON campaign_timeline_events BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('timeline', new.id, new.title, new.event_type, new.description);
END;
CREATE TRIGGER IF NOT EXISTS esi_timeline_ad AFTER DELETE ON campaign_timeline_events BEGIN
    DELETE FROM entity_search_index WHERE entity_type='timeline' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_timeline_au AFTER UPDATE ON campaign_timeline_events BEGIN
    DELETE FROM entity_search_index WHERE entity_type='timeline' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('timeline', new.id, new.title, new.event_type, new.description);
END;

-- oneshot_items (entity type: item)
CREATE TRIGGER IF NOT EXISTS esi_items_ai AFTER INSERT ON oneshot_items BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('item', new.id, new.name, new.category, new.description);
END;
CREATE TRIGGER IF NOT EXISTS esi_items_ad AFTER DELETE ON oneshot_items BEGIN
    DELETE FROM entity_search_index WHERE entity_type='item' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_items_au AFTER UPDATE ON oneshot_items BEGIN
    DELETE FROM entity_search_index WHERE entity_type='item' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('item', new.id, new.name, new.category, new.description);
END;

-- compendium_entries (entity type: compendium)
CREATE TRIGGER IF NOT EXISTS esi_compendium_ai AFTER INSERT ON compendium_entries BEGIN
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('compendium', new.id,
        COALESCE(json_extract(new.data, '$.name'), ''),
        COALESCE((SELECT display_name FROM compendium_schemas WHERE id = new.schema_id), ''),
        new.data);
END;
CREATE TRIGGER IF NOT EXISTS esi_compendium_ad AFTER DELETE ON compendium_entries BEGIN
    DELETE FROM entity_search_index WHERE entity_type='compendium' AND entity_id=old.id;
END;
CREATE TRIGGER IF NOT EXISTS esi_compendium_au AFTER UPDATE ON compendium_entries BEGIN
    DELETE FROM entity_search_index WHERE entity_type='compendium' AND entity_id=old.id;
    INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
    VALUES ('compendium', new.id,
        COALESCE(json_extract(new.data, '$.name'), ''),
        COALESCE((SELECT display_name FROM compendium_schemas WHERE id = new.schema_id), ''),
        new.data);
END;
`

// searchIndexBackfill repopulates the unified index from all source tables.
const searchIndexBackfill = `
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'character', id, name, trim(race || ' ' || class), backstory FROM characters;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'npc', id, name, trim(race || ' ' || class), description || ' ' || notes FROM npcs;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'note', id, title, category, content FROM character_notes;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'quest', id, name, status, description || ' ' || objectives || ' ' || notes FROM quests;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'journal', id, title, entry_date, entry FROM journal;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'session', id, title, session_date, notes || ' ' || important_events FROM sessions;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'campaign', id, name, party_name, description FROM campaigns;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'location', id, name, type, description FROM locations;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'encounter', id, name, difficulty, description || ' ' || environment || ' ' || notes FROM encounter_templates;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'monster', id, name, 'CR ' || cr, description || ' ' || special_abilities || ' ' || actions FROM monster_library;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'shop', id, name, '', description FROM shops;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'faction', id, name, type, description FROM factions;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'adventure', id, title, difficulty, premise || ' ' || hook || ' ' || notes FROM oneshot_adventures;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'wiki', id, title, '', content FROM campaign_wiki_pages;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'timeline', id, title, event_type, description FROM campaign_timeline_events;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'item', id, name, category, description FROM oneshot_items;
INSERT INTO entity_search_index(entity_type, entity_id, title, subtitle, body)
SELECT 'compendium', e.id, COALESCE(json_extract(e.data, '$.name'), ''),
    COALESCE((SELECT display_name FROM compendium_schemas WHERE id = e.schema_id), ''), e.data
FROM compendium_entries e;
`

// linkCleanupTriggers removes entity_links rows whenever a linked entity is
// deleted. Polymorphic columns can't use FK constraints, so triggers (like the
// search index) enforce integrity across every delete path.
const linkCleanupTriggers = `
CREATE TRIGGER IF NOT EXISTS elc_characters AFTER DELETE ON characters BEGIN
    DELETE FROM entity_links WHERE (source_type='character' AND source_id=old.id) OR (target_type='character' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_npcs AFTER DELETE ON npcs BEGIN
    DELETE FROM entity_links WHERE (source_type='npc' AND source_id=old.id) OR (target_type='npc' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_notes AFTER DELETE ON character_notes BEGIN
    DELETE FROM entity_links WHERE (source_type='note' AND source_id=old.id) OR (target_type='note' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_quests AFTER DELETE ON quests BEGIN
    DELETE FROM entity_links WHERE (source_type='quest' AND source_id=old.id) OR (target_type='quest' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_journal AFTER DELETE ON journal BEGIN
    DELETE FROM entity_links WHERE (source_type='journal' AND source_id=old.id) OR (target_type='journal' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_sessions AFTER DELETE ON sessions BEGIN
    DELETE FROM entity_links WHERE (source_type='session' AND source_id=old.id) OR (target_type='session' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_campaigns AFTER DELETE ON campaigns BEGIN
    DELETE FROM entity_links WHERE (source_type='campaign' AND source_id=old.id) OR (target_type='campaign' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_locations AFTER DELETE ON locations BEGIN
    DELETE FROM entity_links WHERE (source_type='location' AND source_id=old.id) OR (target_type='location' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_encounters AFTER DELETE ON encounter_templates BEGIN
    DELETE FROM entity_links WHERE (source_type='encounter' AND source_id=old.id) OR (target_type='encounter' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_monsters AFTER DELETE ON monster_library BEGIN
    DELETE FROM entity_links WHERE (source_type='monster' AND source_id=old.id) OR (target_type='monster' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_shops AFTER DELETE ON shops BEGIN
    DELETE FROM entity_links WHERE (source_type='shop' AND source_id=old.id) OR (target_type='shop' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_factions AFTER DELETE ON factions BEGIN
    DELETE FROM entity_links WHERE (source_type='faction' AND source_id=old.id) OR (target_type='faction' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_adventures AFTER DELETE ON oneshot_adventures BEGIN
    DELETE FROM entity_links WHERE (source_type='adventure' AND source_id=old.id) OR (target_type='adventure' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_wiki AFTER DELETE ON campaign_wiki_pages BEGIN
    DELETE FROM entity_links WHERE (source_type='wiki' AND source_id=old.id) OR (target_type='wiki' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_timeline AFTER DELETE ON campaign_timeline_events BEGIN
    DELETE FROM entity_links WHERE (source_type='timeline' AND source_id=old.id) OR (target_type='timeline' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_items AFTER DELETE ON oneshot_items BEGIN
    DELETE FROM entity_links WHERE (source_type='item' AND source_id=old.id) OR (target_type='item' AND target_id=old.id);
END;
CREATE TRIGGER IF NOT EXISTS elc_compendium AFTER DELETE ON compendium_entries BEGIN
    DELETE FROM entity_links WHERE (source_type='compendium' AND source_id=old.id) OR (target_type='compendium' AND target_id=old.id);
END;
`

// EnsureSearchIndex creates the unified-index sync triggers (idempotently) and
// backfills the index when it is empty. It MUST run after ent's
// Client.Schema.Create, because ent recreates managed tables on startup,
// dropping any triggers attached to them.
func EnsureSearchIndex() error {
	if _, err := DB.Exec(searchIndexTriggers); err != nil {
		return fmt.Errorf("create search index triggers: %w", err)
	}
	if _, err := DB.Exec(linkCleanupTriggers); err != nil {
		return fmt.Errorf("create link cleanup triggers: %w", err)
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM entity_search_index").Scan(&count); err != nil {
		return fmt.Errorf("count search index: %w", err)
	}
	if count == 0 {
		log.Printf("db: backfilling unified search index")
		if _, err := DB.Exec(searchIndexBackfill); err != nil {
			return fmt.Errorf("backfill search index: %w", err)
		}
	}
	return nil
}

// ResyncSearchIndex rebuilds the unified search index from scratch.
// Used by the admin resync action to repair index drift.
func ResyncSearchIndex() error {
	if _, err := DB.Exec(searchIndexTriggers); err != nil {
		return fmt.Errorf("ensure triggers: %w", err)
	}
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM entity_search_index"); err != nil {
		tx.Rollback()
		return fmt.Errorf("clear search index: %w", err)
	}
	if _, err := tx.Exec(searchIndexBackfill); err != nil {
		tx.Rollback()
		return fmt.Errorf("rebackfill search index: %w", err)
	}
	return tx.Commit()
}
