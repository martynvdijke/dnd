package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/handlers/testutil"
)

func TestAmbienceBroadcast(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedUser(t, 3, "stranger", "user")
	testutil.SeedCampaign(t, 1, "AmbCamp", "Testers", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	routes := func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/ambience", SetCampaignAmbience)
	}

	member := testutil.NewRouterWithUser(routes, 2, "user")

	w := testutil.PostJSON(t, member, "/api/campaigns/1/ambience", map[string]any{"track": "rain", "action": "play", "volume": 0.5})
	testutil.AssertStatus(t, w, 200)

	w = testutil.PostJSON(t, member, "/api/campaigns/1/ambience", map[string]any{"track": "nope", "action": "play"})
	testutil.AssertStatus(t, w, 400)

	w = testutil.PostJSON(t, member, "/api/campaigns/1/ambience", map[string]any{"track": "rain", "action": "frobnicate"})
	testutil.AssertStatus(t, w, 400)

	w = testutil.PostJSON(t, member, "/api/campaigns/1/ambience", map[string]any{"action": "stop"})
	testutil.AssertStatus(t, w, 200)

	stranger := testutil.NewRouterWithUser(routes, 3, "user")
	w = testutil.PostJSON(t, stranger, "/api/campaigns/1/ambience", map[string]any{"track": "rain", "action": "play"})
	testutil.AssertStatus(t, w, 403)
}
