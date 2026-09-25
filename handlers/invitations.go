package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

// Campaign invitations: a DM/owner mints a token, optionally emails it, and the
// invited user accepts it from a public page to join the campaign. The token is
// the secret; acceptance additionally requires an authenticated session whose
// email matches the invite (when both are known).

const invitationExpiryDays = 7

type invitationRow struct {
	ID           int64
	CampaignID   int64
	Email        string
	Role         string
	CampaignName string
	ExpiresAt    sql.NullString
	AcceptedAt   sql.NullString
}

func loadInvitation(token string) (*invitationRow, bool) {
	var inv invitationRow
	err := db.DB.QueryRow(`
		SELECT i.id, i.campaign_id, i.email, i.role, COALESCE(c.name,''),
			i.expires_at, i.accepted_at
		FROM campaign_invitations i
		LEFT JOIN campaigns c ON c.id = i.campaign_id
		WHERE i.token=?`, token).
		Scan(&inv.ID, &inv.CampaignID, &inv.Email, &inv.Role, &inv.CampaignName, &inv.ExpiresAt, &inv.AcceptedAt)
	return &inv, err == nil
}

func invitationExpired(inv *invitationRow) bool {
	if !inv.ExpiresAt.Valid || inv.ExpiresAt.String == "" {
		return false
	}
	var n int
	db.DB.QueryRow("SELECT CASE WHEN ? < datetime('now') THEN 1 ELSE 0 END", inv.ExpiresAt.String).Scan(&n)
	return n == 1
}

// publicBaseURL derives the externally reachable origin, honoring BASE_URL and
// reverse-proxy headers (same approach as resolvePublicURL).
func publicBaseURL(c *gin.Context) string {
	if b := strings.TrimRight(BaseURL, "/"); b != "" {
		return b
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func invitationURL(c *gin.Context, token string) string {
	return publicBaseURL(c) + "/invite/" + token
}

type createInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateCampaignInvitation mints an invite token for a campaign. Owner/DM only.
func CreateCampaignInvitation(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, strErr("only campaign DMs can invite members"))
		return
	}
	var req createInvitationRequest
	if !BindOr400(c, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		WriteError(c, http.StatusBadRequest, strErr("valid email required"))
		return
	}
	if req.Role == "" {
		req.Role = "player"
	}
	if req.Role != "player" && req.Role != "dm" {
		WriteError(c, http.StatusBadRequest, strErr("role must be 'dm' or 'player'"))
		return
	}
	invitedBy, _ := c.Get("user_id")
	uid, _ := invitedBy.(int64)
	token := generateToken()
	res, err := db.DB.Exec(`
		INSERT INTO campaign_invitations (token, campaign_id, email, role, invited_by, expires_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', ?))`,
		token, campaignID, req.Email, req.Role, uid, fmt.Sprintf("+%d days", invitationExpiryDays))
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	id, _ := res.LastInsertId()

	var campaignName string
	db.DB.QueryRow("SELECT COALESCE(name,'') FROM campaigns WHERE id=?", campaignID).Scan(&campaignName)
	link := invitationURL(c, token)

	// Best-effort email; the response always carries the link so DM invites work
	// even when SMTP is not configured.
	emailSent := false
	if settings, err := getEmailSettings(); err == nil {
		inviter, _ := c.Get("username")
		body := fmt.Sprintf(`
<h2>You're invited to %s</h2>
<p><strong>%s</strong> invited you to join the <strong>%s</strong> campaign on Villum as a %s.</p>
<p><a href="%s">Accept the invitation</a></p>
<p>Or paste this link into your browser:<br><code>%s</code></p>
<p>This invitation expires in %d days.</p>
<hr><p style="color:#888;">Sent from Villum</p>`,
			campaignName, inviter, campaignName, req.Role, link, link, invitationExpiryDays)
		if err := sendEmail(settings, req.Email, "Invitation to join "+campaignName, body); err != nil {
			middleware.LogWarn("invitations", "invite email failed", "error", err)
		} else {
			emailSent = true
		}
	}
	middleware.LogInfo("invitations", "created", "campaign_id", campaignID, "invitation_id", id, "email_sent", emailSent)
	WriteJSON(c, http.StatusCreated, gin.H{
		"id":         id,
		"token":      token,
		"url":        link,
		"email":      req.Email,
		"role":       req.Role,
		"expires_at": fmt.Sprintf("%d days", invitationExpiryDays),
		"email_sent": emailSent,
	})
}

// ListCampaignInvitations lists pending (unaccepted) invites. Owner/DM only.
func ListCampaignInvitations(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, strErr("only campaign DMs can view invitations"))
		return
	}
	rows, err := db.DB.Query(`
		SELECT id, email, role, COALESCE(created_at,''), COALESCE(expires_at,'')
		FROM campaign_invitations
		WHERE campaign_id=? AND accepted_at IS NULL
		ORDER BY id DESC`, campaignID)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	type inviteOut struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at"`
	}
	out := []inviteOut{}
	for rows.Next() {
		var v inviteOut
		rows.Scan(&v.ID, &v.Email, &v.Role, &v.CreatedAt, &v.ExpiresAt)
		out = append(out, v)
	}
	WriteJSON(c, http.StatusOK, out)
}

// RevokeCampaignInvitation deletes a pending invite. Owner/DM only.
func RevokeCampaignInvitation(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isCampaignDM(c, campaignID) {
		WriteError(c, http.StatusForbidden, strErr("only campaign DMs can revoke invitations"))
		return
	}
	inviteID, _ := strconv.ParseInt(c.Param("inviteId"), 10, 64)
	if _, err := db.DB.Exec("DELETE FROM campaign_invitations WHERE id=? AND campaign_id=?", inviteID, campaignID); err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

type invitePageView struct {
	Token        string
	CampaignName string
	CampaignID   int64
	Email        string
	Role         string
	LoggedIn     bool
	Username     string
	UserEmail    string
	LoginURL     string
	Accepted     bool
	Expired      bool
	Mismatch     bool
	Already      bool
}

func sessionUser(c *gin.Context) (userID int64, username string, ok bool) {
	sid, err := c.Cookie("session")
	if err != nil || sid == "" {
		return 0, "", false
	}
	sess := middleware.Store.Get(sid)
	if sess == nil {
		return 0, "", false
	}
	return sess.UserID, sess.Username, true
}

// InviteAcceptPage is the public landing page for an invite link.
func InviteAcceptPage(c *gin.Context) {
	token := c.Param("token")
	inv, ok := loadInvitation(token)
	if !ok {
		c.String(http.StatusNotFound, "Invitation not found.")
		return
	}
	view := invitePageView{
		Token:        token,
		CampaignName: inv.CampaignName,
		CampaignID:   inv.CampaignID,
		Email:        inv.Email,
		Role:         inv.Role,
		Accepted:     inv.AcceptedAt.Valid && inv.AcceptedAt.String != "",
		Expired:      invitationExpired(inv),
	}
	if uid, username, ok := sessionUser(c); ok {
		view.LoggedIn = true
		view.Username = username
		db.DB.QueryRow("SELECT COALESCE(email,'') FROM users WHERE id=?", uid).Scan(&view.UserEmail)
		var n int
		db.DB.QueryRow("SELECT COUNT(*) FROM campaign_members WHERE campaign_id=? AND user_id=?", inv.CampaignID, uid).Scan(&n)
		view.Already = n > 0
	} else {
		view.LoginURL = "/login?next=" + url.QueryEscape("/invite/"+token)
	}
	renderSharePage(c, "invite.html", view)
}

// InviteAccept is the public accept action. Requires a session; joins the
// campaign and marks the invite used.
func InviteAccept(c *gin.Context) {
	token := c.Param("token")
	inv, ok := loadInvitation(token)
	if !ok {
		c.String(http.StatusNotFound, "Invitation not found.")
		return
	}
	uid, _, loggedIn := sessionUser(c)
	if !loggedIn {
		c.Redirect(http.StatusSeeOther, "/login?next="+url.QueryEscape("/invite/"+token))
		return
	}
	if inv.AcceptedAt.Valid && inv.AcceptedAt.String != "" {
		c.Redirect(http.StatusSeeOther, "/invite/"+token)
		return
	}
	if invitationExpired(inv) {
		c.String(http.StatusGone, "This invitation has expired.")
		return
	}
	// Email binding: if both sides are known, they must match. A secret token
	// plus a session still warrants this guard so a leaked link can't be
	// redeemed by an unrelated logged-in user.
	var userEmail string
	db.DB.QueryRow("SELECT COALESCE(email,'') FROM users WHERE id=?", uid).Scan(&userEmail)
	if inv.Email != "" && userEmail != "" && !strings.EqualFold(inv.Email, userEmail) {
		c.String(http.StatusForbidden, "This invitation was sent to a different email address.")
		return
	}
	var exists int
	db.DB.QueryRow("SELECT COUNT(*) FROM campaigns WHERE id=?", inv.CampaignID).Scan(&exists)
	if exists == 0 {
		c.String(http.StatusNotFound, "Campaign no longer exists.")
		return
	}
	var memberCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM campaign_members WHERE campaign_id=? AND user_id=?", inv.CampaignID, uid).Scan(&memberCount)
	if memberCount == 0 {
		if _, err := db.DB.Exec("INSERT INTO campaign_members (campaign_id, user_id, role) VALUES (?, ?, ?)", inv.CampaignID, uid, inv.Role); err != nil {
			c.String(http.StatusInternalServerError, "Failed to join campaign.")
			return
		}
	}
	if _, err := db.DB.Exec("UPDATE campaign_invitations SET accepted_at=datetime('now'), accepted_by=? WHERE id=?", uid, inv.ID); err != nil {
		middleware.LogWarn("invitations", "failed to mark accepted", "error", err)
	}
	middleware.LogInfo("invitations", "accepted", "campaign_id", inv.CampaignID, "user_id", uid)
	c.Redirect(http.StatusSeeOther, "/")
}
