package db

import (
	"fmt"
	"strings"
	"testing"

	"villum/ent/migrate"
)

// rawOnlyTables lists every non-ent table that is expected to exist in a
// fully migrated DB. Each entry is raw-managed (migrations or FTS) or is an
// FTS5 shadow table. Any live table not in migrate.Tables and not here fails
// the test — unregistered raw DDL is a defect (ADR docs/adr-data-access.md).
var rawOnlyTables = map[string]string{
	"_migration_v4_check":            "migration bookkeeping table created by early migrations",
	"_migration_v5_check":            "migration bookkeeping table created by early migrations",
	"api_tokens":                     "raw-managed via db/migrations — API token store",
	"app_settings":                   "raw-managed via db/migrations — app-wide settings",
	"auth_sessions":                  "raw-managed via db/migrations — session store",
	"campaign_event_settings":        "raw-managed via db/migrations — calendar event settings",
	"campaign_invitations":           "raw-managed via db/migrations — campaign invite tokens",
	"campaign_knowledge":             "raw-managed via db/migrations — knowledge base",
	"campaign_knowledge_known_by":    "raw-managed via db/migrations — many-to-many for knowledge visibility",
	"campaign_monster_roster":        "raw-managed via db/migrations — monster roster for campaigns",
	"campaign_npcs":                  "raw-managed via db/migrations — campaign-NPC join",
	"characters_fts":                 "FTS5 virtual table for characters (ent has no FTS support)",
	"characters_fts_config":          "FTS5 shadow table for characters_fts",
	"characters_fts_data":            "FTS5 shadow table for characters_fts",
	"characters_fts_docsize":         "FTS5 shadow table for characters_fts",
	"characters_fts_idx":             "FTS5 shadow table for characters_fts",
	"clue_dependencies":              "raw-managed via db/migrations — clue dependency graph",
	"clue_locations":                 "raw-managed via db/migrations — clue-location links",
	"clue_npcs":                      "raw-managed via db/migrations — clue-NPC links",
	"clues":                          "raw-managed via db/migrations — investigation clues",
	"compendium_entries":             "raw-managed via db/migrations — dynamic compendium entries",
	"compendium_entries_fts":         "FTS5 virtual table for compendium_entries",
	"compendium_entries_fts_config":  "FTS5 shadow table for compendium_entries_fts",
	"compendium_entries_fts_data":    "FTS5 shadow table for compendium_entries_fts",
	"compendium_entries_fts_docsize": "FTS5 shadow table for compendium_entries_fts",
	"compendium_entries_fts_idx":     "FTS5 shadow table for compendium_entries_fts",
	"compendium_import_logs":         "raw-managed via db/migrations — compendium import history",
	"compendium_monsters":            "raw-managed via db/migrations — compendium monster data",
	"compendium_schemas":             "raw-managed via db/migrations — dynamic compendium schemas",
	"copilot_conversations":          "raw-managed via db/migrations — copilot conversations",
	"copilot_messages":               "raw-managed via db/migrations — copilot messages",
	"dm_notes":                       "raw-managed via db/migrations — DM notes",
	"entity_links":                   "raw-managed via db/migrations/045_entity_links.go — polymorphic cross-entity links",
	"entity_search_index":            "FTS5 virtual table for unified search (migration 045 + db/search_index.go triggers)",
	"entity_search_index_config":     "FTS5 shadow table for entity_search_index",
	"entity_search_index_content":    "FTS5 shadow table for entity_search_index",
	"entity_search_index_data":       "FTS5 shadow table for entity_search_index",
	"entity_search_index_docsize":    "FTS5 shadow table for entity_search_index",
	"entity_search_index_idx":        "FTS5 shadow table for entity_search_index",
	"events_settings":                "raw-managed via db/migrations — events settings",
	"external_import_logs":           "raw-managed via db/migrations — external import logs",
	"google_events_cache":            "raw-managed via db/migrations — Google calendar cache",
	"log_settings":                   "raw-managed via db/migrations — persisted log level",
	"monster_library":                "raw-managed via db/migrations — monster library",
	"npc_item_links":                 "raw-managed via db/migrations/npc_item_links.go — NPC-item join",
	"oneshot_adventure_locations":    "raw-managed via db/migrations — oneshot adventure locations",
	"oneshot_adventure_npcs":         "raw-managed via db/migrations — oneshot adventure NPCs",
	"oneshot_monsters":               "raw-managed via db/migrations — oneshot monsters",
	"oneshot_player_characters":      "raw-managed via db/migrations — oneshot player characters",
	"oneshot_scene_dialogs":          "raw-managed via db/migrations — oneshot scene dialogs",
	"parties":                        "raw-managed via db/migrations — parties",
	"pregen_characters":              "raw-managed via db/migrations — pre-generated characters",
	"prep_checklist":                 "raw-managed via db/migrations — DM prep checklist",
	"push_mutes":                     "raw-managed via db/migrations — push notification mutes",
	"push_reminder_log":              "raw-managed via db/migrations — push reminder log",
	"push_subscriptions":             "raw-managed via db/migrations — push subscriptions",
	"race_colors":                    "raw-managed via db/migrations — race color config",
	"scene_timings":                  "raw-managed via db/migrations — scene timing data",
	"schema_version":                 "raw-managed via db/migrations — schema version tracking",
	"session_pacing":                 "raw-managed via db/migrations — session pacing",
	"transfer_import_logs":           "raw-managed via db/migrations — transfer import logs",
	"umami_settings":                 "raw-managed via db/migrations — analytics settings",
}

func TestSchemaConsistency(t *testing.T) {
	if err := Init(":memory:"); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer Close()

	// Build ent table set.
	entTables := make(map[string]bool)
	for _, tbl := range migrate.Tables {
		entTables[tbl.Name] = true
	}
	if len(entTables) == 0 {
		t.Fatalf("migrate.Tables is empty")
	}

	// Collect live user tables from sqlite_master.
	rows, err := DB.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	defer rows.Close()
	var live []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if strings.HasPrefix(name, "sqlite_") {
			continue
		}
		live = append(live, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	liveSet := make(map[string]bool, len(live))
	for _, n := range live {
		liveSet[n] = true
	}

	// Forward: every ent table/column exists in live DB.
	for _, tbl := range migrate.Tables {
		if !liveSet[tbl.Name] {
			t.Errorf("ent table %q missing in live DB", tbl.Name)
			continue
		}
		cols, err := pragmaColumns(tbl.Name)
		if err != nil {
			t.Errorf("pragma table_info %q: %v", tbl.Name, err)
			continue
		}
		for _, col := range tbl.Columns {
			if !cols[col.Name] {
				t.Errorf("ent column %q.%q missing in live DB", tbl.Name, col.Name)
			}
		}
	}

	// Inverse: every live table is ent-managed or allowlisted.
	for _, name := range live {
		if entTables[name] {
			continue
		}
		if _, ok := rawOnlyTables[name]; ok {
			continue
		}
		// FTS5 shadow tables that may appear with varying suffixes — treat any
		// table whose name starts with a known FTS base + "_" as allowed.
		if isFTSShadow(name) {
			continue
		}
		t.Errorf("live table %q is neither ent-managed nor in rawOnlyTables allowlist — unregistered raw DDL", name)
	}
}

func pragmaColumns(table string) (map[string]bool, error) {
	// table names are from migrate.Tables / sqlite_master, not user input, but
	// still quote with double quotes to handle any reserved names.
	q := fmt.Sprintf("PRAGMA table_info(%q)", table)
	rows, err := DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}

func isFTSShadow(name string) bool {
	// Known FTS5 bases produce shadow tables with these suffixes.
	bases := []string{"characters_fts", "compendium_entries_fts", "entity_search_index"}
	for _, b := range bases {
		if strings.HasPrefix(name, b+"_") {
			return true
		}
	}
	return false
}
