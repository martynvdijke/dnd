package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/handlers/testutil"
)

// Guards the dashboard recent-activity queries: combat_entries has no `round`
// column and dice_rolls uses `timestamp`, not `created_at`.
func TestCampaignDashboardRecentActivity(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedCampaign(t, 1, "DashCamp", "Testers", 1)
	testutil.SeedCharacterInCampaign(t, 10, 1, 1, "Hero", "Elf", "Ranger")

	if _, err := db.DB.Exec(`INSERT INTO combat_entries (campaign_id, name, created_at) VALUES (1, 'Goblin Fight', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("seed combat: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO dice_rolls (user_id, character_id, expression, result, total, timestamp) VALUES (1, 10, '1d20', '[15]', 15, '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("seed roll: %v", err)
	}

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/dashboard", GetCampaignDashboard)
	})
	w := testutil.Get(t, r, "/api/campaigns/1/dashboard")
	testutil.AssertStatus(t, w, 200)

	var resp CampaignDashboard
	testutil.ParseJSON(t, w, &resp)
	if len(resp.RecentCombats) != 1 || resp.RecentCombats[0].Name != "Goblin Fight" {
		t.Fatalf("recent combats = %+v, want one Goblin Fight", resp.RecentCombats)
	}
	if len(resp.RecentDiceRolls) != 1 || resp.RecentDiceRolls[0].Total != 15 {
		t.Fatalf("recent dice rolls = %+v, want one total 15", resp.RecentDiceRolls)
	}
}
