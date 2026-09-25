package handlers

import (
	"testing"

	"strconv"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/handlers/testutil"
)

func battlemapRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/battlemap", GetCampaignBattlemap)
		auth.POST("/campaigns/:id/battlemap/tokens", CreateBattlemapToken)
		auth.POST("/campaigns/:id/battlemap/sync", SyncBattlemapTokens)
		auth.PUT("/battlemap-tokens/:id", UpdateBattlemapToken)
		auth.DELETE("/battlemap-tokens/:id", DeleteBattlemapToken)
	})
}

func TestBattlemapTokens(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedCampaign(t, 1, "MapCamp", "The Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	if _, err := db.DB.Exec(`INSERT INTO combat_entries (campaign_id, name, type, hp_max, hp_current, ac, is_active) VALUES (1, 'Goblin', 'monster', 30, 30, 14, 1)`); err != nil {
		t.Fatalf("seed combat: %v", err)
	}
	var entryID int64
	db.DB.QueryRow("SELECT id FROM combat_entries WHERE campaign_id=1").Scan(&entryID)

	r := battlemapRouter()

	// Sync creates one token per active combatant, named after the entry.
	w := testutil.PostJSON(t, r, "/api/campaigns/1/battlemap/sync", map[string]any{})
	testutil.AssertStatus(t, w, 200)
	var sync struct {
		Created int `json:"created"`
	}
	testutil.ParseJSON(t, w, &sync)
	if sync.Created != 1 {
		t.Fatalf("sync created = %d, want 1", sync.Created)
	}
	if n := testutil.CountRows(t, "battlemap_tokens"); n != 1 {
		t.Fatalf("token rows = %d, want 1", n)
	}

	// Token mirrors the linked combat entry's HP/AC.
	w = testutil.Get(t, r, "/api/campaigns/1/battlemap")
	testutil.AssertStatus(t, w, 200)
	var bm struct {
		Map    map[string]any   `json:"map"`
		Tokens []battlemapToken `json:"tokens"`
	}
	testutil.ParseJSON(t, w, &bm)
	if len(bm.Tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(bm.Tokens))
	}
	tok := bm.Tokens[0]
	if tok.CombatEntryID == nil || *tok.CombatEntryID != entryID {
		t.Fatalf("token not linked to entry: %+v", tok)
	}
	if tok.HPMax == nil || *tok.HPMax != 30 || tok.HPCurrent == nil || *tok.HPCurrent != 30 {
		t.Fatalf("token HP = %v/%v, want 30/30", tok.HPCurrent, tok.HPMax)
	}
	if tok.Name != "Goblin" {
		t.Fatalf("token name = %q, want Goblin", tok.Name)
	}

	// Move clamps to the unit square.
	w = testutil.PutJSON(t, r, "/api/battlemap-tokens/"+strconv.FormatInt(tok.ID, 10), map[string]any{"x": 1.7, "y": -0.2})
	testutil.AssertStatus(t, w, 200)
	var moved struct {
		Token battlemapToken `json:"token"`
	}
	testutil.ParseJSON(t, w, &moved)
	if moved.Token.X != 1 || moved.Token.Y != 0 {
		t.Fatalf("moved token = %v,%v want 1,0", moved.Token.X, moved.Token.Y)
	}

	// A campaign member (non-DM) may move but not delete or sync.
	memberRouter := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/battlemap", GetCampaignBattlemap)
		auth.PUT("/battlemap-tokens/:id", UpdateBattlemapToken)
		auth.DELETE("/battlemap-tokens/:id", DeleteBattlemapToken)
	}, 2, "user")
	testutil.AssertStatus(t, testutil.Get(t, memberRouter, "/api/campaigns/1/battlemap"), 200)
	testutil.AssertStatus(t, testutil.Delete(t, memberRouter, "/api/battlemap-tokens/"+strconv.FormatInt(tok.ID, 10)), 403)

	// A stranger sees nothing.
	strangerRouter := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/battlemap", GetCampaignBattlemap)
		auth.POST("/campaigns/:id/battlemap/tokens", CreateBattlemapToken)
	}, 99, "user")
	testutil.AssertStatus(t, testutil.Get(t, strangerRouter, "/api/campaigns/1/battlemap"), 403)
	testutil.AssertStatus(t, testutil.PostJSON(t, strangerRouter, "/api/campaigns/1/battlemap/tokens", map[string]any{"name": "X"}), 403)

	// DM deletes the token.
	testutil.AssertStatus(t, testutil.Delete(t, r, "/api/battlemap-tokens/"+strconv.FormatInt(tok.ID, 10)), 200)
	if n := testutil.CountRows(t, "battlemap_tokens"); n != 0 {
		t.Fatalf("token rows after delete = %d, want 0", n)
	}
}
