package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/handlers/testutil"
)

func vttVisionRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/maps", CreateCampaignMap)
		auth.GET("/maps/:id/walls", GetMapWalls)
		auth.PUT("/maps/:id/walls", UpdateMapWalls)
		auth.GET("/campaigns/:id/battlemap", GetCampaignBattlemap)
		auth.POST("/campaigns/:id/battlemap/tokens", CreateBattlemapToken)
		auth.PUT("/battlemap-tokens/:id", UpdateBattlemapToken)
	})
}

func TestMapWalls(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedCampaign(t, 1, "WallCamp", "The Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	r := vttVisionRouter()
	testutil.AssertStatus(t, testutil.PostJSON(t, r, "/api/campaigns/1/maps", map[string]any{
		"name": "Dungeon", "width": 1000, "height": 800, "grid_size": 50,
	}), 201)

	// DM saves walls.
	w := testutil.PutJSON(t, r, "/api/maps/1/walls", map[string]any{
		"walls": []map[string]float64{{"x1": 0.1, "y1": 0.1, "x2": 0.1, "y2": 0.9}},
	})
	testutil.AssertStatus(t, w, 200)

	// Walls round-trip.
	w = testutil.Get(t, r, "/api/maps/1/walls")
	testutil.AssertStatus(t, w, 200)
	var got struct {
		Walls []map[string]float64 `json:"walls"`
	}
	testutil.ParseJSON(t, w, &got)
	if len(got.Walls) != 1 || got.Walls[0]["x2"] != 0.1 {
		t.Fatalf("walls = %+v, want one segment ending at x2=0.1", got.Walls)
	}

	// Malformed payload is rejected and leaves walls unchanged.
	w = testutil.PutJSON(t, r, "/api/maps/1/walls", map[string]any{"walls": "not-an-array"})
	testutil.AssertStatus(t, w, 400)
	w = testutil.Get(t, r, "/api/maps/1/walls")
	testutil.ParseJSON(t, w, &got)
	if len(got.Walls) != 1 {
		t.Fatalf("walls after malformed = %d, want 1", len(got.Walls))
	}

	// A non-DM member may read but not write.
	memberRouter := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/maps/:id/walls", GetMapWalls)
		auth.PUT("/maps/:id/walls", UpdateMapWalls)
	}, 2, "user")
	testutil.AssertStatus(t, testutil.Get(t, memberRouter, "/api/maps/1/walls"), 200)
	testutil.AssertStatus(t, testutil.PutJSON(t, memberRouter, "/api/maps/1/walls", map[string]any{
		"walls": []map[string]float64{{"x1": 0, "y1": 0, "x2": 1, "y2": 1}},
	}), 403)
}

func TestBattlemapTokenAuraVision(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedCampaign(t, 1, "AuraCamp", "The Party", 1)

	r := vttVisionRouter()
	testutil.AssertStatus(t, testutil.PostJSON(t, r, "/api/campaigns/1/maps", map[string]any{
		"name": "Arena", "width": 1000, "height": 800, "grid_size": 50,
	}), 201)

	w := testutil.PostJSON(t, r, "/api/campaigns/1/battlemap/tokens", map[string]any{
		"name": "Paladin", "aura_radius": 2, "vision_radius": 6,
	})
	testutil.AssertStatus(t, w, 201)

	w = testutil.Get(t, r, "/api/campaigns/1/battlemap")
	testutil.AssertStatus(t, w, 200)
	var bm struct {
		Tokens []battlemapToken `json:"tokens"`
	}
	testutil.ParseJSON(t, w, &bm)
	if len(bm.Tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(bm.Tokens))
	}
	tok := bm.Tokens[0]
	if tok.AuraRadius != 2 || tok.VisionRadius != 6 {
		t.Fatalf("aura/vision = %v/%v, want 2/6", tok.AuraRadius, tok.VisionRadius)
	}

	// Update changes the aura radius.
	w = testutil.PutJSON(t, r, "/api/battlemap-tokens/1", map[string]any{"aura_radius": 3})
	testutil.AssertStatus(t, w, 200)
	var updated struct {
		Token battlemapToken `json:"token"`
	}
	testutil.ParseJSON(t, w, &updated)
	if updated.Token.AuraRadius != 3 {
		t.Fatalf("aura after update = %v, want 3", updated.Token.AuraRadius)
	}
}
