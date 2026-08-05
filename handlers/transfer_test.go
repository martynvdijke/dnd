package handlers

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestTransferExport(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")

	// Seed entities owned by user 1.
	testutil.SeedCharacter(t, 1, 1, "Hero", "Human", "Fighter")
	testutil.SeedCampaign(t, 1, "Test Campaign", "The Party", 1)
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(1, 1, 'Grimble', 'Gnome', 'Tinker', 'A clever gnome inventor')`); err != nil {
		t.Fatalf("seed npc: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO locations(id, user_id, name, type, description)
		VALUES(1, 1, 'Emberhold', 'town', 'A small town')`); err != nil {
		t.Fatalf("seed location: %v", err)
	}
	// Adventure owned by user 1.
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO oneshot_adventures(id, user_id, title, premise, template)
		VALUES(1, 1, 'The Lost Mine', 'A classic adventure', 'standard')`); err != nil {
		t.Fatalf("seed adventure: %v", err)
	}
	// NPC owned by user 2 (should not appear in user 1's export when non-admin).
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(2, 2, 'User2 NPC', 'Elf', 'Wizard', 'Should be hidden')`); err != nil {
		t.Fatalf("seed npc 2: %v", err)
	}

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 1, "user")

	t.Run("export single type", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/export", map[string]any{
			"types": []string{"npc"},
		})
		testutil.AssertStatus(t, w, 200)
		var env TransferEnvelope
		testutil.ParseJSON(t, w, &env)
		if env.VillumTransfer.Version != 1 {
			t.Errorf("expected version 1, got %d", env.VillumTransfer.Version)
		}
		if len(env.Entities) != 1 {
			t.Fatalf("expected 1 npc, got %d", len(env.Entities))
		}
		if env.Entities[0].Type != "npc" || env.Entities[0].OriginalID != 1 {
			t.Errorf("unexpected entity: type=%s id=%d", env.Entities[0].Type, env.Entities[0].OriginalID)
		}
		if name, ok := env.Entities[0].Data["name"].(string); !ok || name != "Grimble" {
			t.Errorf("expected name Grimble, got %v", env.Entities[0].Data["name"])
		}
	})

	t.Run("export all types returns only visible entities", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/export", map[string]any{})
		testutil.AssertStatus(t, w, 200)
		var env TransferEnvelope
		testutil.ParseJSON(t, w, &env)
		// User 1 should see their own entities but not user 2's NPC.
		for _, e := range env.Entities {
			if e.Type == "npc" && e.OriginalID == 2 {
				t.Errorf("user 1 should not see user 2's npc")
			}
		}
		if len(env.Entities) == 0 {
			t.Error("expected at least some entities")
		}
	})

	t.Run("admin sees all entities", func(t *testing.T) {
		adminRouter := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			RegisterTransferRoutes(auth)
		}, 1, "admin")
		w := testutil.PostJSON(t, adminRouter, "/api/transfer/export", map[string]any{})
		testutil.AssertStatus(t, w, 200)
		var env TransferEnvelope
		testutil.ParseJSON(t, w, &env)
		// Admin should see user 2's NPC.
		found := false
		for _, e := range env.Entities {
			if e.Type == "npc" && e.OriginalID == 2 {
				found = true
				break
			}
		}
		if !found {
			t.Error("admin should see user 2's npc")
		}
	})

	t.Run("export campaign filters correctly", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/transfer/export/campaign/1")
		testutil.AssertStatus(t, w, 200)
		var env TransferEnvelope
		testutil.ParseJSON(t, w, &env)
		// Campaign 1 should be included.
		foundCampaign := false
		for _, e := range env.Entities {
			if e.Type == "campaign" {
				foundCampaign = true
			}
		}
		if !foundCampaign {
			t.Error("expected campaign in campaign export")
		}
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/export", "not a json object")
		testutil.AssertStatus(t, w, 400)
	})
}

func TestTransferImport(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")

	// Seed a campaign and character owned by user 1.
	testutil.SeedCampaign(t, 1, "Source Campaign", "Old Party", 1)
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO characters(id, user_id, name, race, class, level,
		str, dex, con, int, wis, cha, hp_max, hp_current, ac, initiative, speed)
		VALUES(1, 1, 'Warrior', 'Dwarf', 'Fighter', 5,
		16, 14, 16, 10, 12, 10, 55, 55, 18, 10, 25)`); err != nil {
		t.Fatalf("seed character: %v", err)
	}

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 2, "user")

	// Build a transfer envelope that imports campaign+character to user 2.
	// old_id=1 campaign should get remapped to a new ID, and the character's
	// campaign_id should follow.
	envelope := TransferEnvelope{
		VillumTransfer: TransferMeta{
			Version:    1,
			ExportedAt: "2026-01-01T00:00:00Z",
			Source:     "villum",
		},
		Entities: []TransferEntity{
			{
				Type: "campaign", OriginalID: 1,
				Data: map[string]any{
					"name":        "Imported Campaign",
					"party_name":  "New Party",
					"description": "Brought from another instance",
				},
			},
			{
				Type: "character", OriginalID: 1,
				Data: map[string]any{
					"name":        "Imported Hero",
					"race":        "Elf",
					"class":       "Ranger",
					"level":       3,
					"campaign_id": int64(1), // references old campaign ID
					"str":         int64(12), "dex": int64(16), "con": int64(14),
					"int": int64(10), "wis": int64(14), "cha": int64(10),
					"hp_max": int64(30), "hp_current": int64(30),
					"ac": int64(15), "initiative": int64(3), "speed": int64(30),
				},
			},
		},
	}

	t.Run("dry run returns ok without inserting", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/import?dry_run=true", envelope)
		testutil.AssertStatus(t, w, 200)
		var result TransferImportResult
		testutil.ParseJSON(t, w, &result)
		if !result.DryRun {
			t.Error("expected dry_run=true in response")
		}
		// Verify nothing was actually inserted.
		var campaignCount int
		db.DB.QueryRow("SELECT COUNT(*) FROM campaigns WHERE name='Imported Campaign'").Scan(&campaignCount)
		if campaignCount > 0 {
			t.Error("campaign was inserted despite dry run")
		}
	})

	t.Run("import creates entities with remapped IDs", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/import", envelope)
		testutil.AssertStatus(t, w, 200)
		var result TransferImportResult
		testutil.ParseJSON(t, w, &result)
		if result.Status != "completed" {
			t.Fatalf("expected completed, got %s: %s", result.Status, result.Error)
		}
		if result.ImportID == 0 {
			t.Error("expected non-zero import_id")
		}
		if len(result.Results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(result.Results))
		}
		// Both should be imported.
		for _, r := range result.Results {
			if r.Status != "imported" {
				t.Errorf("expected imported, got %s for %s id=%d", r.Status, r.Type, r.OriginalID)
			}
		}
		// Verify the campaign was inserted with a new ID.
		var newCampaignID int64
		var campaignName, partyName string
		err := db.DB.QueryRow("SELECT id, name, party_name FROM campaigns WHERE name='Imported Campaign'").Scan(&newCampaignID, &campaignName, &partyName)
		if err != nil {
			t.Fatalf("campaign not found: %v", err)
		}
		if newCampaignID == 1 {
			t.Error("expected new campaign ID (not 1 which belongs to user 1)")
		}
		if partyName != "New Party" {
			t.Errorf("expected party_name 'New Party', got %s", partyName)
		}
		// Verify character was inserted with campaign_id remapped.
		var charCampaignID int64
		var charName string
		err = db.DB.QueryRow("SELECT campaign_id, name FROM characters WHERE name='Imported Hero'").Scan(&charCampaignID, &charName)
		if err != nil {
			t.Fatalf("character not found: %v", err)
		}
		if charCampaignID != newCampaignID {
			t.Errorf("expected character campaign_id %d (remapped), got %d", newCampaignID, charCampaignID)
		}
		// Character should be owned by user 2.
		var userID int64
		db.DB.QueryRow("SELECT user_id FROM characters WHERE name='Imported Hero'").Scan(&userID)
		if userID != 2 {
			t.Errorf("expected character owned by user 2, got %d", userID)
		}
	})

	t.Run("rejects unknown entity type", func(t *testing.T) {
		bad := envelope
		bad.Entities = []TransferEntity{
			{Type: "nonexistent", OriginalID: 1, Data: map[string]any{"name": "test"}},
		}
		w := testutil.PostJSON(t, r, "/api/transfer/import", bad)
		testutil.AssertStatus(t, w, 400)
	})
}

func TestTransferLogs(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 2, "user")

	t.Run("list logs returns empty for new user", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/transfer/logs")
		testutil.AssertStatus(t, w, 200)
		var logs []any
		testutil.ParseJSON(t, w, &logs)
		if len(logs) != 0 {
			t.Errorf("expected empty logs, got %d", len(logs))
		}
	})

	t.Run("log appears after import", func(t *testing.T) {
		envelope := TransferEnvelope{
			VillumTransfer: TransferMeta{Version: 1, ExportedAt: "2026-01-01T00:00:00Z", Source: "villum"},
			Entities: []TransferEntity{
				{Type: "npc", OriginalID: 1, Data: map[string]any{"name": "Test NPC", "race": "Human", "class": "Commoner", "description": "Test"}},
			},
		}
		testutil.PostJSON(t, r, "/api/transfer/import", envelope)

		w := testutil.Get(t, r, "/api/transfer/logs")
		testutil.AssertStatus(t, w, 200)
		var logs []map[string]any
		testutil.ParseJSON(t, w, &logs)
		if len(logs) == 0 {
			t.Fatal("expected at least 1 log after import")
		}
		if logs[0]["status"] != "completed" {
			t.Errorf("expected status completed, got %v", logs[0]["status"])
		}
		if logs[0]["doc_type"] != "villum-transfer" {
			t.Errorf("expected doc_type 'villum-transfer', got %v", logs[0]["doc_type"])
		}
	})
}

func TestTransferImportEdgeCases(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 1, "user")

	t.Run("rejects empty entities", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/import", TransferEnvelope{
			VillumTransfer: TransferMeta{Version: 1, ExportedAt: "now", Source: "villum"},
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("rejects null data", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/import", TransferEnvelope{
			VillumTransfer: TransferMeta{Version: 1, ExportedAt: "now", Source: "villum"},
			Entities: []TransferEntity{
				{Type: "npc", OriginalID: 1, Data: nil},
			},
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("non-transferable type fails", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/transfer/import", TransferEnvelope{
			VillumTransfer: TransferMeta{Version: 1, ExportedAt: "now", Source: "villum"},
			Entities: []TransferEntity{
				{Type: "note", OriginalID: 1, Data: map[string]any{"title": "test"}},
			},
		})
		testutil.AssertStatus(t, w, 400)
	})
}

// TestTransferImportFKChain tests a deeper dependency chain:
// campaign -> adventure -> shop (references campaign AND adventure)
func TestTransferImportFKChain(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 1, "user")

	envelope := TransferEnvelope{
		VillumTransfer: TransferMeta{Version: 1, ExportedAt: "2026-01-01T00:00:00Z", Source: "villum"},
		Entities: []TransferEntity{
			{
				Type: "campaign", OriginalID: 10,
				Data: map[string]any{"name": "FK Chain Campaign", "party_name": "Test"},
			},
			{
				Type: "adventure", OriginalID: 20,
				Data: map[string]any{
					"title": "Test Adventure", "premise": "Testing FK chain",
					"template": "standard", "campaign_id": int64(10),
				},
			},
			{
				Type: "shop", OriginalID: 30,
				Data: map[string]any{
					"name":                 "Campaign Shop",
					"description":          "A shop in the campaign",
					"campaign_id":          int64(10),
					"oneshot_adventure_id": int64(20),
				},
			},
		},
	}

	w := testutil.PostJSON(t, r, "/api/transfer/import", envelope)
	testutil.AssertStatus(t, w, 200)
	var result TransferImportResult
	testutil.ParseJSON(t, w, &result)
	if result.Status != "completed" {
		t.Fatalf("expected completed, got %s: %v", result.Status, result.Results)
	}
	if len(result.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result.Results))
	}
	for _, r := range result.Results {
		if r.Status != "imported" {
			t.Errorf("expected imported, got %s for %s id=%d", r.Status, r.Type, r.OriginalID)
		}
	}

	// Verify FK remapping: shop's adventure_id should point to the new adventure ID.
	var shopAdventureID int64
	err := db.DB.QueryRow(`SELECT oneshot_adventure_id FROM shops WHERE name='Campaign Shop'`).Scan(&shopAdventureID)
	if err != nil {
		t.Fatalf("shop not found: %v", err)
	}
	if shopAdventureID == 20 {
		t.Error("expected shop.oneshot_adventure_id to be remapped (should not be 20)")
	}
	// The adventure should have the new campaign ID.
	var advCampaignID int64
	err = db.DB.QueryRow(`SELECT campaign_id FROM oneshot_adventures WHERE title='Test Adventure'`).Scan(&advCampaignID)
	if err != nil {
		t.Fatalf("adventure not found: %v", err)
	}
	if advCampaignID == 10 {
		t.Error("expected adventure.campaign_id to be remapped (should not be 10)")
	}
}

// TestTransferRoundTrip exports data then re-imports it as a different user.
func TestTransferRoundTrip(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "user1", "user")
	testutil.SeedUser(t, 2, "user2", "user")

	// Seed data for user 1.
	testutil.SeedCharacter(t, 1, 1, "RoundTrip Hero", "Half-Elf", "Paladin")
	testutil.SeedCampaign(t, 2, "RoundTrip Campaign", "The Circle", 1)
	if _, err := db.DB.Exec(`INSERT OR IGNORE INTO npcs(id, user_id, name, race, class, description)
		VALUES(100, 1, 'RoundTrip NPC', 'Dwarf', 'Blacksmith', 'A friendly smith')`); err != nil {
		t.Fatalf("seed npc: %v", err)
	}

	// Export as user 1 (admin so we can see campaign even though not owned directly by user_id).
	exportRouter := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 1, "admin") // admin to see everything

	w := testutil.PostJSON(t, exportRouter, "/api/transfer/export", map[string]any{
		"types": []string{"character", "npc", "campaign"},
	})
	testutil.AssertStatus(t, w, 200)
	var exportData TransferEnvelope
	testutil.ParseJSON(t, w, &exportData)

	if len(exportData.Entities) != 3 {
		t.Fatalf("expected 3 entities in export, got %d", len(exportData.Entities))
	}

	// Now import as user 2.
	importRouter := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterTransferRoutes(auth)
	}, 2, "user")

	w = testutil.PostJSON(t, importRouter, "/api/transfer/import", exportData)
	testutil.AssertStatus(t, w, 200)
	var importResult TransferImportResult
	testutil.ParseJSON(t, w, &importResult)

	if importResult.Status != "completed" {
		t.Fatalf("round-trip import failed: %s - %v", importResult.Error, importResult.Results)
	}

	// Verify user 2 now owns the entities.
	for _, r := range importResult.Results {
		if r.Status != "imported" {
			t.Errorf("expected imported, got %s for %s", r.Status, r.Type)
		}
		if r.NewID == 0 {
			t.Errorf("expected non-zero new_id for %s", r.Type)
		}
	}
}

// TestTransferToInt64 tests the toInt64 helper.
func TestTransferToInt64(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
		ok       bool
	}{
		{int64(42), 42, true},
		{float64(42), 42, true},
		{int(42), 42, true},
		{json.Number("42"), 42, true},
		{"42", 42, true},
		{"not a number", 0, false},
		{nil, 0, false},
		{true, 0, false},
	}
	for _, tt := range tests {
		result, ok := toInt64(tt.input)
		if ok != tt.ok {
			t.Errorf("toInt64(%v) ok=%v, want %v", tt.input, ok, tt.ok)
		}
		if ok && result != tt.expected {
			t.Errorf("toInt64(%v) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

// TestTransferRemapFK tests the remapFK helper.
func TestTransferRemapFK(t *testing.T) {
	idMapping := map[string]map[int64]int64{
		"campaign": {1: 100, 2: 200},
	}
	t.Run("remaps existing FK", func(t *testing.T) {
		data := map[string]any{"campaign_id": int64(1)}
		remapFK(data, "campaign_id", "campaign", idMapping)
		if data["campaign_id"] != int64(100) {
			t.Errorf("expected 100, got %v", data["campaign_id"])
		}
	})
	t.Run("ignores non-existent FK", func(t *testing.T) {
		data := map[string]any{"campaign_id": int64(999)}
		remapFK(data, "campaign_id", "campaign", idMapping)
		if data["campaign_id"] != int64(999) {
			t.Errorf("expected unchanged 999, got %v", data["campaign_id"])
		}
	})
	t.Run("ignores nil FK", func(t *testing.T) {
		data := map[string]any{"campaign_id": nil}
		remapFK(data, "campaign_id", "campaign", idMapping)
		if data["campaign_id"] != nil {
			t.Errorf("expected nil, got %v", data["campaign_id"])
		}
	})
}
