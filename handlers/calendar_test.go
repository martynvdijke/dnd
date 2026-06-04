package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestCalendarEventCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Calendar Campaign", "Testers", 1)

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/calendar", ListCalendarEvents)
		auth.POST("/calendar", CreateCalendarEvent)
		auth.PUT("/calendar/:id", UpdateCalendarEvent)
		auth.DELETE("/calendar/:id", DeleteCalendarEvent)
	})

	t.Run("list calendar events returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/calendar?campaign_id=1")
		testutil.AssertStatus(t, w, 200)
		var events []any
		testutil.ParseJSON(t, w, &events)
	})

	t.Run("list without campaign_id returns 400", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/calendar")
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("create calendar event returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/calendar", map[string]any{
			"campaign_id": 1,
			"title":       "Session 1",
			"description": "First session",
			"event_date":  "2026-06-01",
			"event_type":  "session",
			"color":       "#ff0000",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("update calendar event returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/calendar/1", map[string]any{
			"title":      "Updated Session",
			"event_date": "2026-06-02",
			"event_type": "session",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete calendar event returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/calendar/1")
		testutil.AssertStatus(t, w, 200)
	})
}
