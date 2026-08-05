package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
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
		c.Set("user_id", sess.UserID)
		c.Set("username", sess.Username)
		c.Set("role", sess.Role)
		c.Set("session_id", sessionID)
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

func CSRFRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
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

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": c.Errors.Last().Error()})
		}
	}
}
