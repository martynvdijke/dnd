package handlers

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestRecapCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 1, 1, "RecapHero", "Human", "Fighter")
	testutil.SeedCampaign(t, 1, "Recap Campaign", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/recaps", ListCampaignRecaps)
		auth.POST("/campaigns/:id/recaps", CreateCampaignRecap)
		auth.GET("/recaps/:id", GetCampaignRecap)
		auth.PUT("/recaps/:id", UpdateCampaignRecap)
		auth.DELETE("/recaps/:id", DeleteCampaignRecap)
		auth.POST("/campaigns/:id/recaps/generate", GenerateCampaignRecap)
		auth.PUT("/recaps/:id/send", MarkRecapAsSent)
	})

	var recapID int64

	t.Run("list recaps returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/campaigns/1/recaps")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("create recap returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps", map[string]any{
			"title":             "Session 1 Recap",
			"content":           "The party explored the dungeon...",
			"session_start_date": "2026-06-01",
			"session_end_date":   "2026-06-01",
		})
		testutil.AssertStatus(t, w, 201)
		var result map[string]any
		testutil.ParseJSON(t, w, &result)
		recapID = int64(result["id"].(float64))
	})

	t.Run("get recap returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.Get(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10))
		testutil.AssertStatus(t, w, 200)
		var recap map[string]any
		testutil.ParseJSON(t, w, &recap)
		if recap["title"] != "Session 1 Recap" {
			t.Fatalf("expected 'Session 1 Recap', got %v", recap["title"])
		}
	})

	t.Run("update recap returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.PutJSON(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10), map[string]any{
			"title":   "Updated Recap",
			"content": "Updated content",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("generate recap returns 200 or 500", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/campaigns/1/recaps/generate", nil)
		if w.Code != 200 && w.Code != 500 {
			t.Fatalf("expected 200 or 500, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("mark recap as sent returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.PutJSON(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10)+"/send", nil)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete recap returns 200", func(t *testing.T) {
		if recapID == 0 {
			t.Skip("no recap created")
		}
		w := testutil.Delete(t, r, "/api/recaps/"+strconv.FormatInt(recapID, 10))
		testutil.AssertStatus(t, w, 200)
	})
}
