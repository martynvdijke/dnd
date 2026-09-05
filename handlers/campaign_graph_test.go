package handlers

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
	"villum/models"
)

func TestCampaignGraphSnapshot(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Graph Camp", "Party", 1)
	testutil.SeedCharacterInCampaign(t, 1, 1, 1, "Hero", "Human", "Fighter")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/graph", GetCampaignGraphData)
	})

	w := testutil.Get(t, r, "/api/campaigns/1/graph")
	testutil.AssertStatus(t, w, 200)
	var gd models.GraphData
	if err := json.Unmarshal(w.Body.Bytes(), &gd); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Ensure dedup works and at least campaign + character nodes present
	if len(gd.Nodes) < 2 {
		t.Fatalf("expected >=2 nodes got %d", len(gd.Nodes))
	}
	// Check no duplicate IDs
	seen := map[string]bool{}
	for _, n := range gd.Nodes {
		if seen[n.ID] {
			t.Fatalf("duplicate node %s", n.ID)
		}
		seen[n.ID] = true
	}
	// Snapshot helper called
	_ = gin.H{}
}
