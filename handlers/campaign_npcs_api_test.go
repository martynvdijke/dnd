package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/handlers/testutil"
)

func TestListCampaignNPCs(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "P", 1)
	testutil.SeedNPC(t, 5, "Mordenkainen", "Human", "Wizard")
	db.DB.Exec("INSERT INTO campaign_npcs(campaign_id,npc_id,role,notes) VALUES(1,5,'ally','notes')")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/npcs", ListCampaignNPCs)
	})
	w := testutil.Get(t, r, "/api/campaigns/1/npcs")
	testutil.AssertStatus(t, w, 200)
}

func TestLinkNPCCampaign(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "P", 1)
	testutil.SeedNPC(t, 6, "Tasha", "Human", "Witch")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/npcs", LinkNPCCampaign)
	})
	t.Run("success", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/npcs", map[string]any{"npc_id": 6, "role": "ally"})
		testutil.AssertStatus(t, w, 201)
	})
	t.Run("duplicate", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/npcs", map[string]any{"npc_id": 6, "role": "ally"})
		if w.Code != 409 {
			t.Fatalf("want 409 got %d %s", w.Code, w.Body.String())
		}
	})
	t.Run("bad json", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/npcs", "bad")
		if w.Code != 400 {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
}

func TestUpdateCampaignNPC(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "P", 1)
	testutil.SeedNPC(t, 7, "Vecna", "Lich", "Wizard")
	db.DB.Exec("INSERT INTO campaign_npcs(id,campaign_id,npc_id,role) VALUES(10,1,7,'enemy')")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/campaign-npcs/:id", UpdateCampaignNPC)
	})
	t.Run("success", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/campaign-npcs/10", map[string]any{"role": "ally", "notes": "hi"})
		testutil.AssertStatus(t, w, 200)
	})
	t.Run("bad json", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/campaign-npcs/10", "bad")
		if w.Code != 400 {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
}

func TestUnlinkCampaignNPC(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "P", 1)
	testutil.SeedNPC(t, 8, "Strahd", "Vampire", "Wizard")
	db.DB.Exec("INSERT INTO campaign_npcs(id,campaign_id,npc_id) VALUES(11,1,8)")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.DELETE("/campaign-npcs/:id", UnlinkCampaignNPC)
	})
	w := testutil.Delete(t, r, "/api/campaign-npcs/11")
	testutil.AssertStatus(t, w, 200)
}

func TestCreateAndLinkCampaignNPC(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "P", 1)
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/campaigns/:id/npcs/create-and-link", CreateAndLinkCampaignNPC)
	})
	t.Run("success", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/npcs/create-and-link", map[string]any{"name": "NewNPC", "race": "Elf", "class": "Rogue", "role": "ally"})
		testutil.AssertStatus(t, w, 201)
	})
	t.Run("bad json", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/npcs/create-and-link", "bad")
		if w.Code != 400 {
			t.Fatalf("want 400 got %d", w.Code)
		}
	})
}
