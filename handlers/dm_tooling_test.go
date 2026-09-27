package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func TestGenerateTreasure(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/generate/treasure", HandleGenerateTreasure)
	})

	w := testutil.Get(t, r, "/api/generate/treasure?tier=2")
	testutil.AssertStatus(t, w, 200)
	var hoard map[string]any
	testutil.ParseJSON(t, w, &hoard)
	if hoard["tier"].(float64) != 2 {
		t.Fatalf("tier = %v, want 2", hoard["tier"])
	}
	coins, ok := hoard["coins"].(map[string]any)
	if !ok {
		t.Fatalf("coins missing: %v", hoard["coins"])
	}
	for _, denom := range []string{"cp", "sp", "ep", "gp", "pp"} {
		if _, ok := coins[denom]; !ok {
			t.Fatalf("coins missing %s", denom)
		}
	}
	for _, key := range []string{"gems", "art_objects", "magic_items"} {
		if _, ok := hoard[key].([]any); !ok {
			t.Fatalf("%s is not an array: %v", key, hoard[key])
		}
	}

	w = testutil.Get(t, r, "/api/generate/treasure?tier=99")
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &hoard)
	if hoard["tier"].(float64) != 1 {
		t.Fatalf("unknown tier should fall back to 1, got %v", hoard["tier"])
	}
}

func TestGenerateWeatherBiome(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/generate/weather", HandleGenerateWeather)
	})

	for range 40 {
		w := testutil.Get(t, r, "/api/generate/weather?biome=desert")
		testutil.AssertStatus(t, w, 200)
		var weather map[string]any
		testutil.ParseJSON(t, w, &weather)
		if weather["biome"] != "desert" {
			t.Fatalf("biome = %v, want desert", weather["biome"])
		}
		if p := weather["precipitation"]; p == "Snow" || p == "Heavy Snow" {
			t.Fatalf("desert produced %v", p)
		}
	}

	w := testutil.Get(t, r, "/api/generate/weather?biome=volcano")
	testutil.AssertStatus(t, w, 200)
	var weather map[string]any
	testutil.ParseJSON(t, w, &weather)
	if weather["biome"] != "temperate" {
		t.Fatalf("unknown biome should fall back to temperate, got %v", weather["biome"])
	}
}

func TestGenerateNameCultures(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/generate/name", HandleGenerateName)
	})

	for _, race := range []string{"orc", "dragonborn", "tiefling", "gnome"} {
		w := testutil.Get(t, r, "/api/generate/name?race="+race)
		testutil.AssertStatus(t, w, 200)
		var name map[string]any
		testutil.ParseJSON(t, w, &name)
		if name["race"] != race {
			t.Fatalf("race = %v, want %s", name["race"], race)
		}
		if s, _ := name["name"].(string); s == "" {
			t.Fatalf("empty name for %s", race)
		}
	}
}

func TestDailyBudget(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/encounters/daily-budget", HandleDailyBudget)
	})

	w := testutil.Get(t, r, "/api/encounters/daily-budget?levels=1,1,1,1")
	testutil.AssertStatus(t, w, 200)
	var budget map[string]any
	testutil.ParseJSON(t, w, &budget)
	if budget["party_budget"].(float64) != 1200 {
		t.Fatalf("party_budget = %v, want 1200", budget["party_budget"])
	}
	if s, _ := budget["suggested_encounters"].(string); s == "" {
		t.Fatal("suggested_encounters should be non-empty")
	}

	w = testutil.Get(t, r, "/api/encounters/daily-budget?levels=99")
	testutil.AssertStatus(t, w, 200)
	testutil.ParseJSON(t, w, &budget)
	if budget["party_budget"].(float64) != 0 {
		t.Fatalf("invalid levels should give 0, got %v", budget["party_budget"])
	}
	if len(budget["per_level"].([]any)) != 0 {
		t.Fatalf("per_level should be empty, got %v", budget["per_level"])
	}
}

func TestCampaignDMScreen(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedCampaign(t, 1, "DMScreenCamp", "The Bold", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	testutil.SeedCharacterInCampaign(t, 10, 1, 1, "Hero", "Elf", "Ranger")

	if _, err := db.DB.Exec(`INSERT INTO combat_entries (campaign_id, name, type, hp_current, hp_max, ac, is_active, turn_order) VALUES (1, 'Goblin', 'monster', 7, 7, 15, 1, 0)`); err != nil {
		t.Fatalf("seed combat: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO character_conditions (character_id, name, type) VALUES (10, 'Poisoned', 'poisoned')`); err != nil {
		t.Fatalf("seed condition: %v", err)
	}

	routes := func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/dm-screen", HandleCampaignDMScreen)
	}

	dm := testutil.NewRouterWithUser(routes, 1, "admin")
	w := testutil.Get(t, dm, "/api/campaigns/1/dm-screen")
	testutil.AssertStatus(t, w, 200)
	var screen map[string]any
	testutil.ParseJSON(t, w, &screen)
	for _, key := range []string{"campaign", "session", "combat", "party", "ambience"} {
		if _, ok := screen[key]; !ok {
			t.Fatalf("missing key %s", key)
		}
	}
	combat := screen["combat"].(map[string]any)
	if combat["current_turn"] != "Goblin" {
		t.Fatalf("current_turn = %v, want Goblin", combat["current_turn"])
	}
	party := screen["party"].([]any)
	if len(party) != 1 {
		t.Fatalf("party = %v, want 1", party)
	}
	conds := party[0].(map[string]any)["conditions"].([]any)
	if len(conds) != 1 || conds[0] != "Poisoned" {
		t.Fatalf("conditions = %v, want [Poisoned]", conds)
	}

	player := testutil.NewRouterWithUser(routes, 2, "user")
	w = testutil.Get(t, player, "/api/campaigns/1/dm-screen")
	testutil.AssertStatus(t, w, 403)
}
