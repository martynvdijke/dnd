package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
	"villum/telegram"
)

//go:embed templates/legal/*.html
var legalTemplatesFS embed.FS

// LegalData is the operator identity plus the deployment facts the published
// policy is allowed to assert. Every processor flag is read from live settings,
// so an instance cannot claim analytics, mail, push, Telegram or AI processing
// that it does not actually run.
type LegalData struct {
	Controller     string
	ContactEmail   string
	ContactAddress string
	HostingRegion  string
	BaseURL        string
	SessionCookie  string
	Year           int
	Analytics      bool
	AnalyticsURL   string
	Traces         bool
	Mail           bool
	Push           bool
	OIDC           bool
	Telegram       bool
	GoogleCalendar bool
	GoogleOAuth    bool
	AI             bool
	WLED           bool
}

// LegalPageData is the template payload for the public pages and /settings.
type LegalPageData struct {
	Title    string
	Legal    LegalData
	User     string
	SignedIn bool
	CSRF     string
}

func legalData() LegalData {
	d := LegalData{
		Controller:    "the operator of this Villum instance",
		HostingRegion: "the European Union",
		SessionCookie: "session",
		Year:          time.Now().Year(),
		BaseURL:       strings.TrimRight(os.Getenv("BASE_URL"), "/"),
	}
	if v := strings.TrimSpace(os.Getenv("LEGAL_CONTROLLER_NAME")); v != "" {
		d.Controller = v
	}
	d.ContactEmail = strings.TrimSpace(os.Getenv("LEGAL_CONTACT_EMAIL"))
	d.ContactAddress = strings.TrimSpace(os.Getenv("LEGAL_CONTACT_ADDRESS"))
	if v := strings.TrimSpace(os.Getenv("LEGAL_HOSTING_REGION")); v != "" {
		d.HostingRegion = v
	}

	if db.DB != nil {
		var enabled int
		var host, website string
		if err := db.DB.QueryRow(`SELECT enabled, tracker_hostname, website_id FROM umami_settings WHERE id = 1`).
			Scan(&enabled, &host, &website); err == nil {
			d.Analytics = enabled == 1 && host != "" && website != ""
			d.AnalyticsURL = host
		}
		var smtpEnabled int
		var smtpHost string
		if err := db.DB.QueryRow(`SELECT enabled, COALESCE(smtp_host, '') FROM email_settings WHERE id = 1`).
			Scan(&smtpEnabled, &smtpHost); err == nil {
			d.Mail = smtpEnabled == 1 && smtpHost != ""
		}
		var aiCount int
		if err := db.DB.QueryRow(`SELECT COUNT(*) FROM ai_endpoints WHERE enabled = 1`).Scan(&aiCount); err == nil {
			d.AI = aiCount > 0
		}
	}
	if _, ok := loadVAPID(); ok {
		d.Push = true
	}
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != "" {
		d.Traces = true
	}
	if cfg := loadOIDCConfig(); cfg.valid() {
		d.OIDC = true
	}
	if telegram.GetBotToken() != "" {
		d.Telegram = true
	}
	d.WLED = wledEnabled()
	if db.DB != nil {
		var sourceType, authMethod, refreshToken, creds, ical string
		if err := db.DB.QueryRow(`SELECT COALESCE(source_type, 'google_api'), COALESCE(auth_method, ''),
			COALESCE(oauth_refresh_token, ''), COALESCE(credentials_json, ''), COALESCE(ical_url, '')
			FROM events_settings WHERE id = 1`).
			Scan(&sourceType, &authMethod, &refreshToken, &creds, &ical); err == nil {
			d.GoogleCalendar = sourceType == "google_api" && (creds != "" || refreshToken != "" || strings.TrimSpace(os.Getenv("GOOGLE_CALENDAR_CREDENTIALS")) != "")
			d.GoogleOAuth = authMethod == "oauth" && refreshToken != ""
		}
	}
	return d
}

func renderLegalPage(c *gin.Context, page, title string) {
	data := LegalPageData{Title: title, Legal: legalData()}
	if u, ok := c.Get("username"); ok {
		data.User, _ = u.(string)
		data.SignedIn = data.User != ""
	}
	if sid, err := c.Cookie("session"); err == nil && sid != "" {
		data.CSRF = middleware.CSRFHash(sid)
	}
	tmpl, err := template.ParseFS(legalTemplatesFS, "templates/legal/layout.html", "templates/legal/"+page+".html")
	if err != nil {
		InternalErrorText(c, err)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "layout", data); err != nil {
		InternalErrorText(c, err)
	}
}

// PrivacyPage is public: store reviewers and prospective users must be able to
// read the policy before they have an account.
func PrivacyPage(c *gin.Context) { renderLegalPage(c, "privacy", "Privacy policy") }
func TermsPage(c *gin.Context)   { renderLegalPage(c, "terms", "Terms of use") }
func AccountDeletePage(c *gin.Context) {
	renderLegalPage(c, "account_delete", "Delete your account")
}

// RegisterLegalRoutes registers the public policy pages.
func RegisterLegalRoutes(r *gin.Engine) {
	r.GET("/privacy", PrivacyPage)
	r.GET("/terms", TermsPage)
	r.GET("/account/delete", AccountDeletePage)
}

// AccountSettingsPage is the signed-in page that owns export and deletion.
func AccountSettingsPage(c *gin.Context) {
	renderLegalPage(c, "settings", "Account settings")
}

// RegisterAccountRoutes registers the HTML account routes. The group in the
// router supplies AuthRequired + CSRFRequired; the JSON API lives under /api.
func RegisterAccountRoutes(r *gin.RouterGroup) {
	r.GET("/settings", AccountSettingsPage)
	r.POST("/settings/account/delete", HandleAccountDeleteForm)
}

// HandleAccountDeleteForm deletes the signed-in account after the user types
// their username. It is deliberately not reachable with only a session: the
// router also enforces CSRF, and the confirmation token is checked here.
func HandleAccountDeleteForm(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")
	confirm := strings.TrimSpace(c.PostForm("confirm"))
	if confirm == "" || confirm != username {
		c.String(http.StatusBadRequest, "Type your username exactly to confirm deletion.")
		return
	}
	if err := deleteAccountAndSessions(c, userID); err != nil {
		writeAccountDeleteError(c, err)
		return
	}
	middleware.ClearSessionCookie(c)
	c.Redirect(http.StatusSeeOther, "/login?deleted=1")
}

// HandleDeleteMe is the JSON API used by the Flutter app. DELETE /api/me
func HandleDeleteMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")
	var req struct {
		Confirm string `json:"confirm"`
	}
	_ = c.ShouldBindJSON(&req)
	if strings.TrimSpace(req.Confirm) == "" || req.Confirm != username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type your username exactly to confirm deletion"})
		return
	}
	if err := deleteAccountAndSessions(c, userID); err != nil {
		writeAccountDeleteError(c, err)
		return
	}
	middleware.ClearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": true})
}

func deleteAccountAndSessions(c *gin.Context, userID int64) error {
	if err := deleteAccount(c.Request.Context(), userID); err != nil {
		return err
	}
	// The database rows are gone, but DBSessionStore keeps a read cache. Clear
	// it so an already-open browser cannot keep using a deleted account.
	if middleware.Store != nil {
		middleware.Store.DeleteUserSessions(userID)
	}
	return nil
}

func writeAccountDeleteError(c *gin.Context, err error) {
	switch {
	case err == errLastAdmin:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case err == errAccountNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		InternalError(c, err)
	}
}
