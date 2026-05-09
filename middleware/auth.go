package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"vellum/models"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*models.AuthSession
}

var Store = &SessionStore{
	sessions: make(map[string]*models.AuthSession),
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *SessionStore) Create(userID int64, username, role, ip string) string {
	id := generateSessionID()
	s.mu.Lock()
	s.sessions[id] = &models.AuthSession{
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IP:        ip,
	}
	s.mu.Unlock()
	return id
}

func (s *SessionStore) Get(sessionID string) *models.AuthSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil
	}
	return sess
}

func (s *SessionStore) Delete(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

func (s *SessionStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

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

func CSRFRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
			return
		}
		sessionID, _ := c.Cookie("session")
		expected := CSRFHash(sessionID)
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
