package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestTimelineEventCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Timeline Campaign", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/timeline", ListTimelineEvents)
		auth.POST("/timeline", CreateTimelineEvent)
		auth.PUT("/timeline/:id", UpdateTimelineEvent)
		auth.DELETE("/timeline/:id", DeleteTimelineEvent)
	})

	t.Run("list timeline events returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/timeline?campaign_id=1")
		testutil.AssertStatus(t, w, 200)
		var events []any
		testutil.ParseJSON(t, w, &events)
	})

	t.Run("create timeline event returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/timeline", map[string]any{
			"campaign_id": 1,
			"title":       "Start of Adventure",
			"description": "The party set out",
			"event_date":  "2026-06-01",
			"event_type":  "story",
			"importance":  3,
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("update timeline event returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/timeline/1", map[string]any{
			"title":      "Updated Event",
			"importance": 2,
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete timeline event returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/timeline/1")
		testutil.AssertStatus(t, w, 200)
	})
}
