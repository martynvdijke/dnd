package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func trmnlPublicRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/trmnl/character-stats", GetTRMNLCharacterStats)
		auth.GET("/trmnl/campaign-stats", GetTRMNLCampaignStats)
		auth.GET("/trmnl/characters", GetTRMNLCharacterRoster)
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
