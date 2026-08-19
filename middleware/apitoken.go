package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenDB is the database used for API token lookups. It is set by main after
// db.Init (mirroring the middleware.Store pattern) to avoid an import cycle
// between the db and middleware packages.
var TokenDB *sql.DB

// TokenPrefixLen is the number of leading characters of a token secret that
// are safe to display in listings. The full secret is never stored or shown
// again after creation.
const TokenPrefixLen = 10

// GenerateTokenSecret creates a new API token secret: "vlt_" followed by 32
// random bytes hex-encoded (68 characters total).
func GenerateTokenSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "vlt_" + hex.EncodeToString(b), nil
}

// HashTokenSecret returns the SHA-256 hex digest of a token secret. Only this
// digest is persisted; the plaintext secret is shown once at creation.
func HashTokenSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}

// TokenPrefix returns the display prefix of a token secret.
func TokenPrefix(secret string) string {
	if len(secret) <= TokenPrefixLen {
		return secret
	}
	return secret[:TokenPrefixLen]
}

// APITokenRequired enforces a valid bearer API token on mutating requests.
// It MUST run after AuthRequired so the session user_id is available.
//
// GET/HEAD/OPTIONS requests are exempt: the three TRMNL reads are public and
// CSRF preflight must not require a token.
//
// Failures return a generic "invalid API token" message so a rejected request
// never reveals whether a token exists or which user owns it.
func APITokenRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}
		if TokenDB == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
			return
		}
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || len(auth) <= len("Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
			return
		}
		secret := strings.TrimSpace(auth[len("Bearer "):])
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
			return
		}

		userID, _ := c.Get("user_id")
		uid, ok := userID.(int64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
			return
		}

		hash := HashTokenSecret(secret)
		var (
			tokenID   int64
			tokenName string
			expiresAt string
			revokedAt string
		)
		err := TokenDB.QueryRow(
			`SELECT id, name, expires_at, revoked_at FROM api_tokens WHERE token_hash = ? AND user_id = ?`,
			hash, uid,
		).Scan(&tokenID, &tokenName, &expiresAt, &revokedAt)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API token"})
			return
		}
		if revokedAt != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API token"})
			return
		}
		if expiresAt != "" {
			exp, perr := time.Parse(time.RFC3339, expiresAt)
			if perr != nil || time.Now().After(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API token"})
				return
			}
		}
		// Best-effort last_used_at update; failures must not block the request.
		_, _ = TokenDB.Exec(
			`UPDATE api_tokens SET last_used_at = ? WHERE id = ?`,
			time.Now().UTC().Format(time.RFC3339), tokenID,
		)
		c.Set("api_token_id", tokenID)
		c.Set("api_token_name", tokenName)
		c.Next()
	}
}
