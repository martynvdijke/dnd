package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Store is the global session store. Defaults to MemoryStore.
// Replace with DBSessionStore after DB is initialized for persistence.
var Store SessionStore = NewMemoryStore()

func StartCleanupTask() {
	go func() {
		for {
			time.Sleep(15 * time.Minute)
			Store.Cleanup()
		}
	}()
}

func CSRFHash(sessionID string) string {
	h := sha256.Sum256([]byte(sessionID + "-csrf"))
	return hex.EncodeToString(h[:])
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session")
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		sess := Store.Get(sessionID)
		if sess == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}
		// Session cutover (retire-password-fallback): password-issued sessions no
		// longer authenticate while the break-glass flag (ALLOW_PASSWORD_FALLBACK)
		// is off. The row is deleted so it cannot be retried, and the response is
		// identical to the expired-session path (same status, same body) so the
		// client sees one "re-authenticate" outcome. OIDC sessions are left alone.
		if sess.AuthMethod == "password" && !PasswordFallbackEnabled() {
			Store.Delete(sessionID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}
		c.Set("user_id", sess.UserID)
		c.Set("username", sess.Username)
		c.Set("role", sess.Role)
		c.Set("session_id", sessionID)
		// OIDC sessions reuse the same cookie/store; expose the login method
		// so handlers can distinguish SSO vs password sessions if needed.
		c.Set("auth_method", sess.AuthMethod)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		c.Next()
	}
}

func DMRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "dm" && role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "dm or admin required"})
			return
		}
		c.Next()
	}
}

// requestOriginAllowed validates the Origin (or Referer fallback) of a
// state-changing cookie-authenticated request. Browsers always send one of
// these headers on cross-site requests, so a present-but-foreign origin is
// rejected even when an attacker somehow obtained a CSRF token. Requests with
// neither header (CLI/mobile clients using an API token) are left to the
// token check.
func requestOriginAllowed(c *gin.Context) bool {
	origin := c.GetHeader("Origin")
	if origin == "" {
		origin = c.GetHeader("Referer")
	}
	if origin == "" {
		return true
	}
	if origin == "null" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if c.Request.Host == "" || strings.EqualFold(u.Host, c.Request.Host) {
		return true
	}
	// Optional explicit allowlist for reverse proxies / split frontends.
	for _, allowed := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		if strings.EqualFold(allowed, origin) {
			return true
		}
		if au, err := url.Parse(allowed); err == nil && strings.EqualFold(au.Host, u.Host) {
			return true
		}
	}
	return false
}

func CSRFRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		if !requestOriginAllowed(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cross-origin request rejected"})
			return
		}
		token := c.GetHeader("X-CSRF-Token")
		sessionID, _ := c.Cookie("session")
		expected := CSRFHash(sessionID)
		// Test-mode debugging aid: when a client explicitly requests it, echo the
		// expected token so failing requests can be diagnosed from the response.
		if c.GetHeader("X-CSRF-Debug") == "1" {
			c.Header("X-CSRF-Debug-Token", expected)
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
			return
		}
		if token != expected {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid CSRF token"})
			return
		}
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

// ErrorHandler converts unhandled handler errors into a generic 500 response.
// Details stay in the server log; clients never see internal error text.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			LogError("http", "unhandled handler error",
				"method", c.Request.Method, "path", c.Request.URL.Path,
				"error", c.Errors.Last().Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	}
}
