package handlers

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func itoa(n int) string { return fmt.Sprintf("%d", n) }

func TestGenerators(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	genRoutes := map[string]gin.HandlerFunc{
		"/generate/npc":              HandleGenerateNPC,
		"/generate/name":             HandleGenerateName,
		"/generate/encounter":        HandleGenerateEncounter,
		"/generate/loot":             HandleGenerateLoot,
		"/generate/character":        HandleGenerateRandomCharacter,
		"/generate/adventure-hook":   HandleGenerateAdventureHook,
		"/generate/dungeon-dressing": HandleGenerateDungeonDressing,
		"/generate/tavern":           HandleGenerateTavern,
		"/generate/urban-encounter":  HandleGenerateUrbanEncounter,
		"/generate/road-encounter":   HandleGenerateRoadEncounter,
		"/generate/weather":          HandleGenerateWeather,
	}

	for name, h := range genRoutes {
		t.Run(name+" returns 200", func(t *testing.T) {
			testutil.NewDB(t)
			defer testutil.CloseDB(t)
			testutil.SeedUser(t, 1, "admin", "admin")

			r := testutil.NewRouter(func(auth *gin.RouterGroup) {
				auth.GET(name, h)
			})
			w := testutil.Get(t, r, "/api"+name)
			testutil.AssertStatus(t, w, 200)
		})
	}
}

func TestGeneratorFilters(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/generate/name", HandleGenerateName)
		auth.GET("/generate/encounter", HandleGenerateEncounter)
		auth.GET("/generate/loot", HandleGenerateLoot)
		auth.GET("/generate/weather", HandleGenerateWeather)
	})

	t.Run("name with race filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/generate/name?race=dwarf")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("encounter with terrain filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/generate/encounter?terrain=forest&level=5")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("loot with cr filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/generate/loot?cr=1")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("weather with season filter", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/generate/weather?season=Winter")
		testutil.AssertStatus(t, w, 200)
		var weather map[string]any
		testutil.ParseJSON(t, w, &weather)
		if weather["season"] != "Winter" {
			t.Fatalf("expected Winter, got %v", weather["season"])
		}
	})
}
