package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"villum/db"
)

// ─── Session RSVPs & attendance ───

// sessionPlanCampaignID resolves a session plan's campaign, returning false
// when the plan does not exist or is orphaned (campaign_id NULL).
func sessionPlanCampaignID(planID int64) (int64, bool) {
	var campaignID *int64
	if err := db.DB.QueryRow("SELECT campaign_id FROM session_plans WHERE id=?", planID).Scan(&campaignID); err != nil {
		return 0, false
	}
	if campaignID == nil {
		return 0, false
	}
	return *campaignID, true
}

type sessionRSVP struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Note     string `json:"note"`
	Attended bool   `json:"attended"`
	IsMe     bool   `json:"is_me"`
}

// rsvpRows returns every campaign member (plus the owner) with their RSVP for
// a plan. Members who have not responded get an empty status.
func rsvpRows(planID, campaignID, meID int64) ([]sessionRSVP, error) {
	rows, err := db.DB.Query(`
		SELECT cm.user_id, u.username, COALESCE(r.status,''), COALESCE(r.note,''), COALESCE(r.attended,0)
		FROM campaign_members cm
		JOIN users u ON u.id = cm.user_id
		LEFT JOIN session_rsvps r ON r.session_plan_id=? AND r.user_id = cm.user_id
		WHERE cm.campaign_id=?
		UNION
		SELECT c.user_id, u.username, COALESCE(r.status,''), COALESCE(r.note,''), COALESCE(r.attended,0)
		FROM campaigns c
		JOIN users u ON u.id = c.user_id
		LEFT JOIN session_rsvps r ON r.session_plan_id=? AND r.user_id = c.user_id
		WHERE c.id=?`, planID, campaignID, planID, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]sessionRSVP, 0)
	for rows.Next() {
		var r sessionRSVP
		rows.Scan(&r.UserID, &r.Username, &r.Status, &r.Note, &r.Attended)
		r.IsMe = r.UserID == meID
		out = append(out, r)
	}
	return out, rows.Err()
}

func rsvpCounts(list []sessionRSVP) gin.H {
	var yes, maybe, no, attended int
	for _, r := range list {
		switch r.Status {
		case "yes":
			yes++
		case "maybe":
			maybe++
		case "no":
			no++
		}
		if r.Attended {
			attended++
		}
	}
	return gin.H{"yes": yes, "maybe": maybe, "no": no, "attended": attended, "total": len(list)}
}

// ListSessionRSVPs GET /session-plans/:id/rsvps
func ListSessionRSVPs(c *gin.Context) {
	planID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	campaignID, ok := sessionPlanCampaignID(planID)
	if !ok {
		WriteNotFound(c, "session plan not found")
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	meID := c.GetInt64("user_id")
	list, err := rsvpRows(planID, campaignID, meID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	myStatus := ""
	for _, r := range list {
		if r.IsMe {
			myStatus = r.Status
		}
	}
	WriteJSON(c, http.StatusOK, gin.H{"rsvps": list, "counts": rsvpCounts(list), "my_status": myStatus})
}

// SetSessionRSVP PUT /session-plans/:id/rsvp
func SetSessionRSVP(c *gin.Context) {
	planID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	campaignID, ok := sessionPlanCampaignID(planID)
	if !ok {
		WriteNotFound(c, "session plan not found")
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var body struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if !BindOr400(c, &body) {
		return
	}
	if body.Status != "yes" && body.Status != "maybe" && body.Status != "no" {
		WriteError(c, http.StatusBadRequest, strErr("status must be yes, maybe or no"))
		return
	}
	meID := c.GetInt64("user_id")
	_, err := db.DB.Exec(`
		INSERT INTO session_rsvps(session_plan_id, user_id, campaign_id, status, note, updated_at)
		VALUES(?,?,?,?,?, datetime('now'))
		ON CONFLICT(session_plan_id, user_id)
		DO UPDATE SET status=excluded.status, note=excluded.note, updated_at=datetime('now')`,
		planID, meID, campaignID, body.Status, body.Note)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	list, err := rsvpRows(planID, campaignID, meID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "counts": rsvpCounts(list)})
}

// SetSessionAttendance PUT /session-plans/:id/attendance  (DM only)
func SetSessionAttendance(c *gin.Context) {
	planID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	campaignID, ok := sessionPlanCampaignID(planID)
	if !ok {
		WriteNotFound(c, "session plan not found")
		return
	}
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var body struct {
		UserID   int64 `json:"user_id"`
		Attended bool  `json:"attended"`
	}
	if !BindOr400(c, &body) {
		return
	}
	if body.UserID == 0 {
		WriteError(c, http.StatusBadRequest, strErr("user_id required"))
		return
	}
	attended := 0
	if body.Attended {
		attended = 1
	}
	_, err := db.DB.Exec(`
		INSERT INTO session_rsvps(session_plan_id, user_id, campaign_id, status, attended, updated_at)
		VALUES(?,?,?, 'yes', ?, datetime('now'))
		ON CONFLICT(session_plan_id, user_id)
		DO UPDATE SET attended=excluded.attended, updated_at=datetime('now')`,
		planID, body.UserID, campaignID, attended)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}
