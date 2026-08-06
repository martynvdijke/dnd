package db

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestCompositeIndexesUsed verifies the v49 composite indexes are used by the
// planner for the hot query patterns (EXPLAIN QUERY PLAN mentions the index).
func TestCompositeIndexesUsed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "indexes.db")
	if err := Init(path); err != nil {
		t.Fatalf("init: %v", err)
	}
	t.Cleanup(func() { Close() })

	cases := []struct {
		pattern string
		index   string
	}{
		{"SELECT id FROM characters WHERE user_id=1 ORDER BY name, level", "idx_characters_user_name_level"},
		{"SELECT id FROM characters WHERE campaign_id=1 ORDER BY name", "idx_characters_campaign_name"},
		{"SELECT id FROM spells WHERE character_id=1 ORDER BY level, name", "idx_spells_char_level_name"},
		{"SELECT id FROM inventory WHERE character_id=1 ORDER BY category, name", "idx_inventory_char_category_name"},
		{"SELECT id FROM sessions WHERE character_id=1 ORDER BY session_date", "idx_sessions_char_date"},
		{"SELECT id FROM quests WHERE character_id=1 ORDER BY status, updated_at", "idx_quests_char_status_updated"},
		{"SELECT id FROM journal WHERE character_id=1 ORDER BY entry_date", "idx_journal_char_entry_date"},
		{"SELECT id FROM combat_entries WHERE campaign_id=1 ORDER BY turn_order", "idx_combat_campaign_turn"},
		{"SELECT id FROM dice_rolls WHERE user_id=1 ORDER BY timestamp", "idx_dice_rolls_user_timestamp"},
		{"SELECT id FROM campaigns WHERE user_id=1 ORDER BY name", "idx_campaigns_user_name"},
	}

	for _, tc := range cases {
		t.Run(tc.index, func(t *testing.T) {
			rows, err := DB.Query("EXPLAIN QUERY PLAN " + tc.pattern)
			if err != nil {
				t.Fatalf("explain: %v", err)
			}
			defer rows.Close()
			var plan strings.Builder
			for rows.Next() {
				var id, parent, notUsed int
				var detail string
				if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
					t.Fatalf("scan: %v", err)
				}
				plan.WriteString(detail)
				plan.WriteString("\n")
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows: %v", err)
			}
			out := plan.String()
			if !strings.Contains(out, tc.index) {
				t.Fatalf("plan for %q does not use %s:\n%s", tc.pattern, tc.index, out)
			}
		})
	}
}
