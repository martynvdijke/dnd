package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func trmnlAdminRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/settings/trmnl", GetTRMNLSettings)
		auth.PUT("/admin/settings/trmnl", SetTRMNLSettings)
	})
}

func trmnlPublicRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/trmnl/character-stats", GetTRMNLCharacterStats)
		auth.GET("/trmnl/campaign-stats", GetTRMNLCampaignStats)
	})
}

func TestTRMNLTokenAutoCreate(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	r := trmnlAdminRouter()
	w := testutil.Get(t, r, "/api/admin/settings/trmnl")
	testutil.AssertStatus(t, w, 200)

	var first map[string]any
	testutil.ParseJSON(t, w, &first)
	token, ok := first["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected non-empty token, got %+v", first)
	}

	// Second read returns the same stored token.
	w2 := testutil.Get(t, r, "/api/admin/settings/trmnl")
	testutil.AssertStatus(t, w2, 200)
	var second map[string]any
	testutil.ParseJSON(t, w2, &second)
	if second["token"] != token {
		t.Fatalf("expected same token on second read, got %v then %v", token, second["token"])
	}
}

func TestTRMNLTokenRegenerate(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	r := trmnlAdminRouter()
	w := testutil.Get(t, r, "/api/admin/settings/trmnl")
	testutil.AssertStatus(t, w, 200)
	var before map[string]any
	testutil.ParseJSON(t, w, &before)
	oldToken := before["token"].(string)

	w2 := testutil.PutJSON(t, r, "/api/admin/settings/trmnl", map[string]any{"regenerate": true})
	testutil.AssertStatus(t, w2, 200)
	var after map[string]any
	testutil.ParseJSON(t, w2, &after)
	newToken, ok := after["token"].(string)
	if !ok || newToken == "" {
		t.Fatalf("expected non-empty regenerated token, got %+v", after)
	}
	if newToken == oldToken {
		t.Fatalf("expected regenerated token to differ, both %q", newToken)
	}

	// Stored value replaced: GET now returns the new token.
	w3 := testutil.Get(t, r, "/api/admin/settings/trmnl")
	testutil.AssertStatus(t, w3, 200)
	var current map[string]any
	testutil.ParseJSON(t, w3, &current)
	if current["token"] != newToken {
		t.Fatalf("expected stored token %q, got %v", newToken, current["token"])
	}
}

func TestTRMNLCharacterStats(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Thorgar", "Dwarf", "Fighter")

	var settings map[string]any
	w := testutil.Get(t, trmnlAdminRouter(), "/api/admin/settings/trmnl")
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &settings)
	token := settings["token"].(string)

	r := trmnlPublicRouter()

	// No token -> 401
	w2 := testutil.Get(t, r, "/api/trmnl/character-stats?character_id=1")
	testutil.AssertStatus(t, w2, 401)

	// Wrong token -> 401
	w3 := testutil.Get(t, r, "/api/trmnl/character-stats?character_id=1&token=wrongtoken")
	testutil.AssertStatus(t, w3, 401)

	// Valid token + known character -> 200 with stat fields
	w4 := testutil.Get(t, r, "/api/trmnl/character-stats?character_id=1&token="+token)
	testutil.AssertStatus(t, w4, 200)
	var stats map[string]any
	testutil.ParseJSON(t, w4, &stats)
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
	w5 := testutil.Get(t, r, "/api/trmnl/character-stats?character_id=999&token="+token)
	testutil.AssertStatus(t, w5, 404)

	// No mutation: character row count unchanged by polling
	if n := testutil.CountRows(t, "characters"); n != 1 {
		t.Fatalf("expected 1 character row after polling, got %d", n)
	}
}

func TestTRMNLCampaignStats(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "Thorgar", "Dwarf", "Fighter")

	var settings map[string]any
	w := testutil.Get(t, trmnlAdminRouter(), "/api/admin/settings/trmnl")
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &settings)
	token := settings["token"].(string)

	r := trmnlPublicRouter()

	// Missing token -> 401
	w2 := testutil.Get(t, r, "/api/trmnl/campaign-stats?character_id=1")
	testutil.AssertStatus(t, w2, 401)

	// Unknown character -> 404
	w3 := testutil.Get(t, r, "/api/trmnl/campaign-stats?character_id=999&token="+token)
	testutil.AssertStatus(t, w3, 404)

	// Valid token + known character -> 200 with campaign fields
	w4 := testutil.Get(t, r, "/api/trmnl/campaign-stats?character_id=1&token="+token)
	testutil.AssertStatus(t, w4, 200)
	var stats map[string]any
	testutil.ParseJSON(t, w4, &stats)
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

	// No mutation: GET-only polling
	if n := testutil.CountRows(t, "sessions"); n != 0 {
		t.Fatalf("expected 0 session rows after polling, got %d", n)
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
