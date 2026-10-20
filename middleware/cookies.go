package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// cookieSecure reports whether session cookies must carry the Secure flag.
// It is true when the request arrived over TLS, when a reverse proxy reports
// HTTPS via X-Forwarded-Proto, or when COOKIE_SECURE explicitly enables it.
func cookieSecure(c *gin.Context) bool {
	if v := strings.TrimSpace(os.Getenv("COOKIE_SECURE")); v != "" {
		return v == "1" || strings.EqualFold(v, "true")
	}
	if c.Request != nil && c.Request.TLS != nil {
		return true
	}
	return strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
}

// SetSessionCookie writes the session cookie with HttpOnly, SameSite=Lax and
// Secure when the deployment is reached over TLS (directly or via a proxy).
func SetSessionCookie(c *gin.Context, sessionID string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session", sessionID, maxAge, "/", "", cookieSecure(c), true)
}

// ClearSessionCookie expires the session cookie using the same attributes as
// SetSessionCookie.
func ClearSessionCookie(c *gin.Context) {
	SetSessionCookie(c, "", -1)
}

// SetOIDCCookie writes a short-lived OIDC flow cookie with the same hardened
// attributes as the session cookie.
func SetOIDCCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", cookieSecure(c), true)
}
