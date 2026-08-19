package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/middleware"
)

// TokenInfo is the metadata returned for a token listing. The secret itself is
// never included — only its display prefix.
type TokenInfo struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
	RevokedAt  string `json:"revoked_at"`
	LastUsedAt string `json:"last_used_at"`
}

// CreateTokenRequest is the body for POST /api/tokens.
type CreateTokenRequest struct {
	Name          string `json:"name"`
	ExpiresInDays int    `json:"expires_in_days"` // 0 = never expires
}

// CreateTokenResponse includes the one-time-visible secret.
type CreateTokenResponse struct {
	TokenInfo
	Token string `json:"token"`
}

// CreateToken creates a new API token for the current user. The secret is
// returned exactly once; only its SHA-256 hash is persisted.
func CreateToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(req.Name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name too long"})
		return
	}
	secret, err := middleware.GenerateTokenSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	hash := middleware.HashTokenSecret(secret)
	prefix := middleware.TokenPrefix(secret)
	expiresAt := ""
	if req.ExpiresInDays > 0 {
		expiresAt = time.Now().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour).UTC().Format(time.RFC3339)
	}
	res, err := db.DB.Exec(
		`INSERT INTO api_tokens (user_id, name, token_hash, prefix, expires_at) VALUES (?, ?, ?, ?, ?)`,
		uid, req.Name, hash, prefix, expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, CreateTokenResponse{
		TokenInfo: TokenInfo{
			ID:        id,
			Name:      req.Name,
			Prefix:    prefix,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			ExpiresAt: expiresAt,
		},
		Token: secret,
	})
}

// ListTokens returns metadata for the current user's tokens. Secrets are never
// included in listings.
func ListTokens(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	rows, err := db.DB.Query(
		`SELECT id, name, prefix, created_at, expires_at, revoked_at, last_used_at
		 FROM api_tokens WHERE user_id = ? ORDER BY id DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
		return
	}
	defer rows.Close()
	tokens := make([]TokenInfo, 0)
	for rows.Next() {
		var t TokenInfo
		if err := rows.Scan(&t.ID, &t.Name, &t.Prefix, &t.CreatedAt, &t.ExpiresAt, &t.RevokedAt, &t.LastUsedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
			return
		}
		tokens = append(tokens, t)
	}
	c.JSON(http.StatusOK, tokens)
}

// RevokeToken revokes one of the current user's tokens. Revocation is
// owner-scoped: another user's token id returns 404 without revealing it.
func RevokeToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}
	res, err := db.DB.Exec(
		`UPDATE api_tokens SET revoked_at = ? WHERE id = ? AND user_id = ? AND revoked_at = ''`,
		time.Now().UTC().Format(time.RFC3339), id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RotateToken replaces a token's secret while keeping its metadata. The new
// secret is returned once; the old secret is immediately invalid. Rotation is
// owner-scoped.
func RotateToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}
	secret, err := middleware.GenerateTokenSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	hash := middleware.HashTokenSecret(secret)
	prefix := middleware.TokenPrefix(secret)
	res, err := db.DB.Exec(
		`UPDATE api_tokens SET token_hash = ?, prefix = ? WHERE id = ? AND user_id = ? AND revoked_at = ''`,
		hash, prefix, id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate token"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "prefix": prefix, "token": secret})
}

// RegisterTokenRoutes registers user API token lifecycle routes. These are
// session + CSRF protected but intentionally NOT API-token protected: a user
// must be able to create their first token without already holding one.
func RegisterTokenRoutes(r *gin.RouterGroup) {
	r.POST("/tokens", CreateToken)
	r.GET("/tokens", ListTokens)
	r.DELETE("/tokens/:id", RevokeToken)
	r.POST("/tokens/:id/rotate", RotateToken)
}
