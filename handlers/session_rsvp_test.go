package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/handlers/testutil"
)

func rsvpRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/session-plans/:id/rsvps", ListSessionRSVPs)
		auth.PUT("/session-plans/:id/rsvp", SetSessionRSVP)
		auth.PUT("/session-plans/:id/attendance", SetSessionAttendance)
	})
}

func TestSessionRSVP(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	testutil.SeedUser(t, 1, "dmuser", "admin")
	testutil.SeedUser(t, 2, "player1", "user")
	testutil.SeedUser(t, 3, "outsider", "user")
	testutil.SeedCampaign(t, 1, "RSVPCamp", "The Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	res, err := db.DB.Exec(`INSERT INTO session_plans(campaign_id, title, session_date, status, dm_notes, planned_encounters, npc_ids, player_goals, expected_duration, created_at, updated_at)
		VALUES(1, 'Session One', '2026-10-01', 'planned', '', '[]', '[]', '[]', '', datetime('now'), datetime('now'))`)
	if err != nil {
		t.Fatalf("seed session plan: %v", err)
	}
	planID, _ := res.LastInsertId()

	r := rsvpRouter()

	// DM responds yes.
	w := testutil.PutJSON(t, r, "/api/session-plans/1/rsvp", map[string]any{"status": "yes", "note": "bring snacks"})
	testutil.AssertStatus(t, w, 200)

	// DM sees own response reflected.
	w = testutil.Get(t, r, "/api/session-plans/1/rsvps")
	testutil.AssertStatus(t, w, 200)
	var listResp struct {
		RSVPs    []sessionRSVP  `json:"rsvps"`
		Counts   map[string]int `json:"counts"`
		MyStatus string         `json:"my_status"`
	}
	testutil.ParseJSON(t, w, &listResp)
	if listResp.MyStatus != "yes" {
		t.Fatalf("my_status = %q, want yes", listResp.MyStatus)
	}
	if listResp.Counts["yes"] != 1 || listResp.Counts["total"] != 2 {
		t.Fatalf("counts = %+v, want yes=1 total=2", listResp.Counts)
	}

	// Member (user 2) responds no; their view reports my_status=no.
	r2 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/session-plans/:id/rsvps", ListSessionRSVPs)
		auth.PUT("/session-plans/:id/rsvp", SetSessionRSVP)
		auth.PUT("/session-plans/:id/attendance", SetSessionAttendance)
	}, 2, "user")
	w = testutil.PutJSON(t, r2, "/api/session-plans/1/rsvp", map[string]any{"status": "no"})
	testutil.AssertStatus(t, w, 200)

	// Invalid status rejected.
	w = testutil.PutJSON(t, r, "/api/session-plans/1/rsvp", map[string]any{"status": "perhaps"})
	testutil.AssertStatus(t, w, 400)

	// Non-member cannot respond or read.
	r3 := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/session-plans/:id/rsvps", ListSessionRSVPs)
		auth.PUT("/session-plans/:id/rsvp", SetSessionRSVP)
		auth.PUT("/session-plans/:id/attendance", SetSessionAttendance)
	}, 3, "user")
	w = testutil.PutJSON(t, r3, "/api/session-plans/1/rsvp", map[string]any{"status": "yes"})
	testutil.AssertStatus(t, w, 403)
	w = testutil.Get(t, r3, "/api/session-plans/1/rsvps")
	testutil.AssertStatus(t, w, 403)

	// Attendance is DM-only.
	w = testutil.PutJSON(t, r2, "/api/session-plans/1/attendance", map[string]any{"user_id": 2, "attended": true})
	testutil.AssertStatus(t, w, 403)
	w = testutil.PutJSON(t, r, "/api/session-plans/1/attendance", map[string]any{"user_id": 2, "attended": true})
	testutil.AssertStatus(t, w, 200)

	// Non-existent plan 404s.
	w = testutil.Get(t, r, "/api/session-plans/999/rsvps")
	testutil.AssertStatus(t, w, 404)

	_ = planID
}
