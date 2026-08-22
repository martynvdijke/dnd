package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func trmnlPublicRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/trmnl/character-stats", GetTRMNLCharacterStats)
		auth.GET("/trmnl/campaign-stats", GetTRMNLCampaignStats)
		auth.GET("/trmnl/characters", GetTRMNLCharacterRoster)
		auth.GET("/trmnl/combat", GetTRMNLCombatStatus)
	})
}

func TestTRMNLCharacterStatsPublic(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Thorgar", "Dwarf", "Fighter")

	r := trmnlPublicRouter()

	// No credentials -> 200 (public read)
	w := testutil.Get(t, r, "/api/trmnl/character-stats?character_id=1")
	testutil.AssertStatus(t, w, 200)
	var stats map[string]any
	testutil.ParseJSON(t, w, &stats)
	testutil.AssertField(t, stats, "name", "Thorgar")
	testutil.AssertField(t, stats, "race", "Dwarf")
	testutil.AssertField(t, stats, "class", "Fighter")
	testutil.AssertField(t, stats, "level", float64(1))
	testutil.AssertField(t, stats, "hp_current", float64(12))
	testutil.AssertField(t, stats, "hp_max", float64(12))
	testutil.AssertField(t, stats, "ac", float64(10))
	testutil.AssertField(t, stats, "initiative", float64(0))
	testutil.AssertField(t, stats, "str", float64(10))
	testutil.AssertField(t, stats, "str_mod", float64(0))

	// Unknown character -> 404
	w2 := testutil.Get(t, r, "/api/trmnl/character-stats?character_id=999")
	testutil.AssertStatus(t, w2, 404)

	// No mutation: character row count unchanged by polling
	if n := testutil.CountRows(t, "characters"); n != 1 {
		t.Fatalf("expected 1 character row after polling, got %d", n)
	}
}

func TestTRMNLCampaignStatsPublic(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Thorgar", "Dwarf", "Fighter")

	r := trmnlPublicRouter()

	// No credentials -> 200 (public read)
	w := testutil.Get(t, r, "/api/trmnl/campaign-stats?character_id=1")
	testutil.AssertStatus(t, w, 200)
	var stats map[string]any
	testutil.ParseJSON(t, w, &stats)
	testutil.AssertField(t, stats, "name", "Thorgar")
	testutil.AssertField(t, stats, "session_count", float64(0))
	testutil.AssertField(t, stats, "total_xp_earned", float64(0))
	testutil.AssertField(t, stats, "total_gold_earned", float64(0))
	if _, ok := stats["quests"]; !ok {
		t.Fatalf("expected quests breakdown in campaign stats: %+v", stats)
	}
	if _, ok := stats["rests"]; !ok {
		t.Fatalf("expected rests breakdown in campaign stats: %+v", stats)
	}
	if _, ok := stats["dice_rolls"]; !ok {
		t.Fatalf("expected dice_rolls in campaign stats: %+v", stats)
	}
	if _, ok := stats["top_npcs"]; !ok {
		t.Fatalf("expected top_npcs in campaign stats: %+v", stats)
	}

	// Unknown character -> 404
	w2 := testutil.Get(t, r, "/api/trmnl/campaign-stats?character_id=999")
	testutil.AssertStatus(t, w2, 404)

	// No mutation: GET-only polling
	if n := testutil.CountRows(t, "sessions"); n != 0 {
		t.Fatalf("expected 0 session rows after polling, got %d", n)
	}
}

func TestTRMNLCharacterRosterPublic(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Thorgar", "Dwarf", "Fighter")
	testutil.SeedCharacter(t, 2, 1, "Aldric", "Human", "Wizard")

	r := trmnlPublicRouter()

	// No credentials -> 200 with JSON array, ordered by name
	w := testutil.Get(t, r, "/api/trmnl/characters")
	testutil.AssertStatus(t, w, 200)
	var roster []map[string]any
	testutil.ParseJSON(t, w, &roster)
	if len(roster) != 2 {
		t.Fatalf("expected 2 characters in roster, got %d: %s", len(roster), w.Body.String())
	}
	// ORDER BY name: Aldric (id 2) before Thorgar (id 1)
	testutil.AssertField(t, roster[0], "id", float64(2))
	testutil.AssertField(t, roster[0], "name", "Aldric")
	testutil.AssertField(t, roster[0], "race", "Human")
	testutil.AssertField(t, roster[0], "class", "Wizard")
	testutil.AssertField(t, roster[0], "character_type", "player")
	testutil.AssertField(t, roster[1], "id", float64(1))
	testutil.AssertField(t, roster[1], "name", "Thorgar")
	testutil.AssertField(t, roster[1], "race", "Dwarf")
	testutil.AssertField(t, roster[1], "class", "Fighter")
	testutil.AssertField(t, roster[1], "character_type", "player")

	// Spot-check stat fields on the first row (seeded defaults: level 1,
	// str 10 -> str_mod 0, hp 12/12, ac 10, initiative 0)
	for _, row := range roster {
		testutil.AssertField(t, row, "level", float64(1))
		testutil.AssertField(t, row, "hp_current", float64(12))
		testutil.AssertField(t, row, "hp_max", float64(12))
		testutil.AssertField(t, row, "ac", float64(10))
		testutil.AssertField(t, row, "initiative", float64(0))
		testutil.AssertField(t, row, "str", float64(10))
		testutil.AssertField(t, row, "str_mod", float64(0))
		if _, ok := row["subclass"]; !ok {
			t.Fatalf("expected subclass field in roster row: %+v", row)
		}
	}

	// No mutation: GET-only polling
	if n := testutil.CountRows(t, "characters"); n != 2 {
		t.Fatalf("expected 2 character rows after polling, got %d", n)
	}
}

func TestAbilityModifier(t *testing.T) {
	cases := []struct {
		score, want int
	}{
		{10, 0}, {11, 0}, {12, 1}, {18, 4}, {20, 5}, {8, -1}, {7, -2}, {3, -4}, {1, -5},
	}
	for _, c := range cases {
		if got := abilityModifier(c.score); got != c.want {
			t.Errorf("abilityModifier(%d) = %d, want %d", c.score, got, c.want)
		}
	}
}

func TestTRMNLCombatStatusPublic(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Curse of Strahd", "Brigade", 1)

	seedCombat := func(name, typ string, roll, mod, hpCur, hpMax, ac, turnOrder int, active bool, conds string) int64 {
		t.Helper()
		a := 0
		if active {
			a = 1
		}
		res, err := db.DB.Exec(`INSERT INTO combat_entries
			(campaign_id, name, type, initiative_roll, initiative_mod, hp_max, hp_current, ac, is_active, turn_order, condition_ids, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '')`,
			1, name, typ, roll, mod, hpMax, hpCur, ac, a, turnOrder, conds)
		if err != nil {
			t.Fatalf("seed combat entry: %v", err)
		}
		id, _ := res.LastInsertId()
		return id
	}

	r := trmnlPublicRouter()

	// Seed out of order on purpose: Thorgar wins outright (25); Orc and the
	// Goblin Boss tie at roll 20 but the Orc has the lower turn_order.
	seedCombat("Goblin Boss", "monster", 20, 0, 30, 44, 17, 2, true, "paralyzed, Poisoned ,bogus-123")
	seedCombat("Thorgar", "character", 25, 0, 12, 12, 16, 0, false, "")
	seedCombat("Orc", "monster", 20, 2, 8, 15, 13, 1, false, "")

	// No credentials -> 200 (public read), ordered like the web tracker.
	w := testutil.Get(t, r, "/api/trmnl/combat?campaign_id=1")
	testutil.AssertStatus(t, w, 200)

	var payload struct {
		CampaignID int64            `json:"campaign_id"`
		Entries    []map[string]any `json:"entries"`
	}
	testutil.ParseJSON(t, w, &payload)

	if payload.CampaignID != 1 {
		t.Fatalf("campaign_id = %d, want 1", payload.CampaignID)
	}
	if len(payload.Entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(payload.Entries))
	}

	wantOrder := []string{"Thorgar", "Orc", "Goblin Boss"}
	for i, name := range wantOrder {
		if got := payload.Entries[i]["name"]; got != name {
			t.Errorf("entry[%d] = %v, want %s", i, got, name)
		}
	}

	// initiative_total = roll + modifier
	testutil.AssertField(t, payload.Entries[0], "initiative_total", float64(25))
	testutil.AssertField(t, payload.Entries[1], "initiative_total", float64(22))
	testutil.AssertField(t, payload.Entries[2], "initiative_total", float64(20))

	// Active-turn flag lands on the right combatant.
	testutil.AssertField(t, payload.Entries[2], "is_active", true)
	testutil.AssertField(t, payload.Entries[0], "is_active", false)

	// Conditions resolve to canonical names; unknown tokens are skipped;
	// whitespace around tokens is tolerated.
	conds, ok := payload.Entries[2]["conditions"].([]any)
	if !ok || len(conds) != 2 || conds[0] != "paralyzed" || conds[1] != "poisoned" {
		t.Fatalf("conditions = %#v, want [paralyzed poisoned]", payload.Entries[2]["conditions"])
	}
	// Empty condition list serializes as [] not null.
	if c, ok := payload.Entries[0]["conditions"].([]any); !ok || len(c) != 0 {
		t.Fatalf("conditions = %#v, want empty array", payload.Entries[0]["conditions"])
	}

	// Unknown campaign -> 200 with an empty list (idle screen).
	w2 := testutil.Get(t, r, "/api/trmnl/combat?campaign_id=999")
	testutil.AssertStatus(t, w2, 200)
	var idle struct {
		CampaignID int64            `json:"campaign_id"`
		Entries    []map[string]any `json:"entries"`
	}
	testutil.ParseJSON(t, w2, &idle)
	if len(idle.Entries) != 0 {
		t.Fatalf("got %d entries for unknown campaign, want 0", len(idle.Entries))
	}

	// Missing campaign_id -> 200 with an empty list too.
	w3 := testutil.Get(t, r, "/api/trmnl/combat")
	testutil.AssertStatus(t, w3, 200)
	var missing struct {
		CampaignID int64            `json:"campaign_id"`
		Entries    []map[string]any `json:"entries"`
	}
	testutil.ParseJSON(t, w3, &missing)
	if len(missing.Entries) != 0 || missing.CampaignID != 0 {
		t.Fatalf("missing campaign_id: got campaign=%d entries=%d, want 0/0",
			missing.CampaignID, len(missing.Entries))
	}

	// Malformed campaign_id -> 400.
	w4 := testutil.Get(t, r, "/api/trmnl/combat?campaign_id=abc")
	testutil.AssertStatus(t, w4, 400)

	// No mutation: GET-only polling.
	if n := testutil.CountRows(t, "combat_entries"); n != 3 {
		t.Fatalf("expected 3 combat_entries rows after polling, got %d", n)
	}
}
