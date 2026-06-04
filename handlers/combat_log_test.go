package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestCombatLogCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Combat Campaign", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/combat-log", ListCombatLogEntries)
		auth.POST("/combat-log", CreateCombatLogEntry)
		auth.GET("/combat-log/stats", GetCombatLogStats)
	})

	t.Run("list combat log returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/combat-log?campaign_id=1")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create combat log entry returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/combat-log", map[string]any{
			"campaign_id":   1,
			"actor_name":    "Goblin",
			"action":        "attack",
			"target_name":   "Hero",
			"damage":        8,
			"damage_type":   "slashing",
			"roll_expression": "1d20+4",
			"roll_total":    15,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("get combat log stats returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/combat-log/stats?campaign_id=1")
		testutil.AssertStatus(t, w, 200)
		var stats map[string]any
		testutil.ParseJSON(t, w, &stats)
	})
}
