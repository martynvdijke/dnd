package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"villum/db"
	googlecalendar "villum/google_calendar"
)

var googleOAuthConfigFn = googlecalendar.OAuthConfig
var googleOAuthExchangeFn = googlecalendar.ExchangeCode

var googleOAuthRedirectURL string

func SetGoogleOAuthRedirectURL(u string) {
	googleOAuthRedirectURL = strings.TrimRight(strings.TrimSpace(u), "/")
}

func clearGoogleOAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("gcal_oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("gcal_oauth_verifier", "", -1, "/", "", false, true)
}

func googleOAuthRedirectURI(c *gin.Context) string {
	if googleOAuthRedirectURL != "" {
		if strings.Contains(googleOAuthRedirectURL, "/api/admin/events-oauth/callback") {
			return googleOAuthRedirectURL
		}
		return googleOAuthRedirectURL + "/api/admin/events-oauth/callback"
	}
	var scheme, host string
	if BaseURL != "" {
		if u, err := url.Parse(BaseURL); err == nil && u.Scheme != "" && u.Host != "" {
			scheme = u.Scheme
			host = u.Host
		}
	}
	if scheme == "" || host == "" {
		proto := c.GetHeader("X-Forwarded-Proto")
		if proto == "http" || proto == "https" {
			scheme = proto
		} else {
			scheme = "https"
		}
		host = c.Request.Host
	}
	return scheme + "://" + host + "/api/admin/events-oauth/callback"
}

func GoogleOAuthStatus(c *gin.Context) {
	s := db.GetEventSettings()
	connected := s.AuthMethod == "oauth" && s.OAuthRefreshToken != ""
	c.JSON(http.StatusOK, gin.H{
		"connected":    connected,
		"redirect_uri": googleOAuthRedirectURI(c),
	})
}

func GoogleOAuthStart(c *gin.Context) {
	s := db.GetEventSettings()
	if s.OAuthClientID == "" || s.OAuthClientSecret == "" {
		c.Redirect(http.StatusFound, "/admin?events_oauth=missing_credentials")
		return
	}
	// Generate state
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		c.Redirect(http.StatusFound, "/admin?events_oauth=error&reason=state")
		return
	}
	state := hex.EncodeToString(b)
	verifier := oauth2.GenerateVerifier()
	setOIDCCookie(c, "gcal_oauth_state", state, 300)
	setOIDCCookie(c, "gcal_oauth_verifier", verifier, 300)
	cfg := googleOAuthConfigFn(s.OAuthClientID, s.OAuthClientSecret, googleOAuthRedirectURI(c))
	u := googlecalendar.AuthCodeURL(cfg, state, verifier)
	c.Redirect(http.StatusFound, u)
}

func GoogleOAuthCallback(c *gin.Context) {
	stateCookie, err1 := c.Cookie("gcal_oauth_state")
	verifier, err2 := c.Cookie("gcal_oauth_verifier")
	clearGoogleOAuthCookies(c)
	code := c.Query("code")
	stateQuery := c.Query("state")
	if err1 != nil || err2 != nil || code == "" || stateQuery == "" || stateCookie == "" || subtle.ConstantTimeCompare([]byte(stateCookie), []byte(stateQuery)) != 1 {
		c.Redirect(http.StatusFound, "/admin?events_oauth=error&reason=state")
		return
	}
	s := db.GetEventSettings()
	cfg := googleOAuthConfigFn(s.OAuthClientID, s.OAuthClientSecret, googleOAuthRedirectURI(c))
	tok, err := googleOAuthExchangeFn(c.Request.Context(), cfg, code, verifier)
	if err != nil || tok == nil {
		c.Redirect(http.StatusFound, "/admin?events_oauth=error&reason=exchange")
		return
	}
	if tok.RefreshToken == "" {
		c.Redirect(http.StatusFound, "/admin?events_oauth=error&reason=no_refresh_token")
		return
	}
	s.AuthMethod = "oauth"
	s.SourceType = "google_api"
	s.OAuthRefreshToken = tok.RefreshToken
	_ = db.SaveEventSettings(s)
	_ = db.ClearCache("")
	c.Redirect(http.StatusFound, "/admin?events_oauth=connected")
}
