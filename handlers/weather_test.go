package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestWeatherGeneration(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/weather", HandleGenerateWeather)
	})

	t.Run("generate weather returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/weather")
		testutil.AssertStatus(t, w, 200)
		var weather map[string]any
		testutil.ParseJSON(t, w, &weather)
		if _, ok := weather["season"]; !ok {
			t.Fatal("expected season in weather response")
		}
	})

	t.Run("generate weather with season returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/weather?season=winter")
		testutil.AssertStatus(t, w, 200)
		var weather map[string]any
		testutil.ParseJSON(t, w, &weather)
	})

	t.Run("generate weather with invalid season still returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/weather?season=invalid")
		testutil.AssertStatus(t, w, 200)
	})
}
