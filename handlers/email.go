package handlers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
	"villum/models"
)

func getEmailSettings() (*models.EmailSettings, error) {
	var s models.EmailSettings
	var enabled int
	var password string
	err := db.DB.QueryRow("SELECT id, COALESCE(smtp_host, ''), COALESCE(smtp_port, 587), COALESCE(username, ''), COALESCE(password, ''), COALESCE(from_addr, ''), COALESCE(enabled, 0) FROM email_settings WHERE id = 1").
		Scan(&s.ID, &s.SMTPHost, &s.SMTPPort, &s.Username, &password, &s.FromAddr, &enabled)
	if err != nil {
		return nil, fmt.Errorf("email settings not found")
	}
	s.Password = password
	s.Enabled = enabled == 1
	if !s.Enabled {
		return nil, fmt.Errorf("email not enabled")
	}
	if s.SMTPHost == "" || s.FromAddr == "" {
		return nil, fmt.Errorf("incomplete email settings")
	}
	if s.SMTPPort == 0 {
		s.SMTPPort = 587
	}
	return &s, nil
}

func sendEmail(settings *models.EmailSettings, to, subject, body string) error {
	auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.SMTPHost)

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		settings.FromAddr, to, subject, body))

	addr := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)

	if settings.SMTPPort == 465 {
		tlsConfig := &tls.Config{ServerName: settings.SMTPHost}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, settings.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		if err = client.Mail(settings.FromAddr); err != nil {
			return fmt.Errorf("mail from: %w", err)
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("rcpt to: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("data: %w", err)
		}
		_, err = w.Write(msg)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, settings.FromAddr, []string{to}, msg)
}

func TestEmail(c *gin.Context) {
	settings, err := getEmailSettings()
	if err != nil {
		middleware.LogError("email", "test email failed to load settings", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userEmail, _ := c.Get("username")
	to := ""
	db.DB.QueryRow("SELECT COALESCE(email, '') FROM users WHERE username=?", userEmail).Scan(&to)
	if to == "" {
		to = settings.FromAddr
	}

	subject := "Test Email from Villum"
	body := fmt.Sprintf(`<h2>Test Email</h2><p>This is a test email from your Villum instance.</p><p>If you received this, your email settings are configured correctly.</p>`)

	if err := sendEmail(settings, to, subject, body); err != nil {
		middleware.LogError("email", "test email send failed", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("email send failed: %v", err)})
		return
	}

	middleware.LogInfo("email", "test email sent", "to", to)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Test email sent successfully"})
}

type CampaignHighlightsRequest struct {
	CampaignID int64 `json:"campaign_id"`
}

func SendCampaignHighlights(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var req CampaignHighlightsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ownerID int64
	err := db.DB.QueryRow("SELECT user_id FROM campaigns WHERE id=?", req.CampaignID).Scan(&ownerID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}
	if role != "admin" && ownerID != userID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	settings, err := getEmailSettings()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("email not configured: %v", err)})
		return
	}

	var campaignName, partyName string
	db.DB.QueryRow("SELECT name, COALESCE(party_name, '') FROM campaigns WHERE id=?", req.CampaignID).Scan(&campaignName, &partyName)

	charRows, _ := db.DB.Query(`
		SELECT c.name, c.race, c.class, c.level, c.hp_max, c.hp_current, COALESCE(u.username, '')
		FROM characters c LEFT JOIN users u ON u.id = c.user_id
		WHERE c.campaign_id=? ORDER BY c.name`, req.CampaignID)
	var chars []map[string]string
	for charRows.Next() {
		var name, race, class, owner string
		var level, hpMax, hpCurrent int
		charRows.Scan(&name, &race, &class, &level, &hpMax, &hpCurrent, &owner)
		status := "alive"
		if hpCurrent <= 0 {
			status = "down"
		} else if float64(hpCurrent)/float64(hpMax) < 0.25 {
			status = "injured"
		}
		chars = append(chars, map[string]string{
			"name": name, "race": race, "class": class, "level": strconv.Itoa(level),
			"hp": fmt.Sprintf("%d/%d", hpCurrent, hpMax), "status": status, "owner": owner,
		})
	}
	charRows.Close()

	sessRows, _ := db.DB.Query(`
		SELECT s.title, s.session_date, s.important_events
		FROM sessions s JOIN characters c ON c.id = s.character_id
		WHERE c.campaign_id=? ORDER BY s.session_date DESC LIMIT 5`, req.CampaignID)
	var sessions []map[string]string
	for sessRows.Next() {
		var title, date, events string
		sessRows.Scan(&title, &date, &events)
		sessions = append(sessions, map[string]string{"title": title, "date": date, "events": events})
	}
	sessRows.Close()

	questRows, _ := db.DB.Query(`
		SELECT q.name, q.status
		FROM quests q JOIN characters c ON c.id = q.character_id
		WHERE c.campaign_id=? AND q.status IN ('active','available')
		ORDER BY q.name LIMIT 10`, req.CampaignID)
	var quests []map[string]string
	for questRows.Next() {
		var name, status string
		questRows.Scan(&name, &status)
		quests = append(quests, map[string]string{"name": name, "status": status})
	}
	questRows.Close()

	memberRows, _ := db.DB.Query(`
		SELECT u.email, u.username FROM campaign_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.campaign_id=? AND u.email != ''
		UNION
		SELECT u.email, u.username FROM campaigns c
		JOIN users u ON u.id = c.user_id
		WHERE c.id=? AND u.email != ''`, req.CampaignID, req.CampaignID)
	var recipients []struct{ email, username string }
	for memberRows.Next() {
		var email, username string
		memberRows.Scan(&email, &username)
		if email != "" {
			recipients = append(recipients, struct{ email, username string }{email, username})
		}
	}
	memberRows.Close()

	if len(recipients) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no recipients with email addresses found"})
		return
	}

	shareURL := fmt.Sprintf("%s/api/share/party/%d", c.Request.Host, req.CampaignID)

	charHTML := ""
	for _, ch := range chars {
		charHTML += fmt.Sprintf(`<tr><td>%s</td><td>%s %s</td><td>Lvl %s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			ch["name"], ch["race"], ch["class"], ch["level"], ch["hp"], ch["status"], ch["owner"])
	}
	sessHTML := ""
	for _, s := range sessions {
		sessHTML += fmt.Sprintf(`<li><strong>%s</strong> (%s): %s</li>`, s["title"], s["date"], s["events"])
	}
	if sessHTML == "" {
		sessHTML = "<li>No recent sessions</li>"
	}
	questHTML := ""
	for _, q := range quests {
		questHTML += fmt.Sprintf(`<li>%s [%s]</li>`, q["name"], q["status"])
	}
	if questHTML == "" {
		questHTML = "<li>No active quests</li>"
	}

	subject := fmt.Sprintf("Campaign Highlights: %s", campaignName)
	body := fmt.Sprintf(`
<h2>%s</h2>
%s
<h3>Party Roster</h3>
<table border="1" cellpadding="6" cellspacing="0" style="border-collapse:collapse;width:100%%">
<tr style="background:#f0f0f0"><th>Name</th><th>Class</th><th>Level</th><th>HP</th><th>Status</th><th>Player</th></tr>
%s
</table>
<h3>Recent Sessions</h3>
<ul>%s</ul>
<h3>Active Quests</h3>
<ul>%s</ul>
<p><a href="%s">View Party Online</a></p>
<hr><p style="color:#888;">Sent from your Villum campaign manager</p>`,
		campaignName, partyName, charHTML, sessHTML, questHTML, shareURL)

	sentCount := 0
	var errors []string
	for _, r := range recipients {
		to := r.email
		if to == "" {
			continue
		}
		if err := sendEmail(settings, to, subject, body); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", r.username, err))
		} else {
			sentCount++
		}
	}

	if sentCount == 0 && len(errors) > 0 {
		middleware.LogError("email", "campaign highlights all sends failed", "campaign_id", req.CampaignID, "errors", strings.Join(errors, "; "))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("all sends failed: %s", strings.Join(errors, "; "))})
		return
	}

	middleware.LogInfo("email", "campaign highlights sent", "campaign_id", req.CampaignID, "recipient_count", sentCount)
	resp := gin.H{"status": "ok", "sent": sentCount}
	if len(errors) > 0 {
		middleware.LogWarn("email", "campaign highlights partial failures", "campaign_id", req.CampaignID, "errors", strings.Join(errors, "; "))
		resp["errors"] = errors
	}
	c.JSON(http.StatusOK, resp)
}
