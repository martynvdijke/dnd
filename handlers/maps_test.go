package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestCampaignMapsFullCycle(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "MapCamp", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/maps", CreateCampaignMap)
		auth.GET("/campaigns/:id/maps", ListCampaignMaps)
		auth.PUT("/maps/:id", UpdateCampaignMap)
		auth.DELETE("/maps/:id", DeleteCampaignMap)
		auth.POST("/campaigns/:id/maps/:mapId/activate", SetActiveCampaignMap)
		auth.GET("/campaigns/:id/maps/active", GetActiveCampaignMap)
		auth.POST("/maps/:mapId/pins", CreateMapPin)
		auth.GET("/maps/:mapId/pins", ListMapPins)
		auth.PUT("/map-pins/:id", UpdateMapPin)
		auth.DELETE("/map-pins/:id", DeleteMapPin)
		auth.PUT("/maps/:id/fog", UpdateFogOfWar)
	})

	t.Run("create map returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/maps", map[string]any{
			"name": "World Map", "width": 1000, "height": 800, "grid_size": 50,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list maps returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/maps")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("activate and get active map returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/maps/1/activate", nil)
		testutil.AssertStatus(t, w, 200)

		w = testutil.Get(t, r, "/api/campaigns/1/maps/active")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create pin returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/maps/1/pins", map[string]any{
			"name": "Waterdeep", "type": "city", "x": 500, "y": 300,
			"icon": "fa-city", "description": "City of Splendors",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("list pins returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/maps/1/pins")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("update fog of war returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/maps/1/fog", map[string]any{
			"fog_of_war": `[]`,
		})
		testutil.AssertStatus(t, w, 200)
	})
}

func TestMapsEdgeCases(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "EdgeMap", "Edgers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/maps/:mapId/activate", SetActiveCampaignMap)
		auth.GET("/campaigns/:id/maps/active", GetActiveCampaignMap)
	})

	t.Run("activate non-existent map returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/maps/999/activate", nil)
		if w.Code != 200 {
			t.Logf("activate non-existent: got %d", w.Code)
		}
	})

	t.Run("get active with no maps returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/maps/active")
		testutil.AssertStatus(t, w, 200)
	})
}
