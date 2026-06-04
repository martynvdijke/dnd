package handlers

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestWikiPageCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Wiki Campaign", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/wiki", ListWikiPages)
		auth.POST("/campaigns/:id/wiki", CreateWikiPage)
		auth.GET("/wiki/:id", GetWikiPage)
		auth.PUT("/wiki/:id", UpdateWikiPage)
		auth.DELETE("/wiki/:id", DeleteWikiPage)
	})

	var pageID int64

	t.Run("list wiki pages returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/wiki")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create wiki page returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/wiki", map[string]any{
			"campaign_id": 1,
			"title":       "Campaign Notes",
			"content":     "Important information about the campaign world.",
			"visibility":  "party",
			"tags":        "lore, world",
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		pageID = int64(result["id"].(float64))
	})

	t.Run("get wiki page returns 200", func(t *testing.T) {
		if pageID == 0 {
			t.Skip("no page created")
		}
		w := testutil.Get(t, r, "/api/wiki/"+strconv.FormatInt(pageID, 10))
		testutil.AssertStatus(t, w, 200)
		var page map[string]any
		testutil.ParseJSON(t, w, &page)
		if page["title"] != "Campaign Notes" {
			t.Fatalf("expected 'Campaign Notes', got %v", page["title"])
		}
	})

	t.Run("update wiki page returns 200", func(t *testing.T) {
		if pageID == 0 {
			t.Skip("no page created")
		}
		w := testutil.PutJSON(t, r, "/api/wiki/"+strconv.FormatInt(pageID, 10), map[string]any{
			"title":      "Updated Notes",
			"content":    "Updated content",
			"visibility": "public",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete wiki page returns 200", func(t *testing.T) {
		if pageID == 0 {
			t.Skip("no page created")
		}
		w := testutil.Delete(t, r, "/api/wiki/"+strconv.FormatInt(pageID, 10))
		testutil.AssertStatus(t, w, 200)
	})
}
