package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/middleware"
)

// SendEmailFunc is injectable for tests.
var SendEmailFunc = sendEmail

// sendPasswordResetEmail sends the reset link email.
func SendPasswordResetEmail(to, link string) error {
	settings, err := getEmailSettings()
	if err != nil {
		return err
	}
	subject := "Villum — Password Reset"
	body := fmt.Sprintf(`<p>You requested a password reset for your Villum account.</p><p><a href="%s">Reset your password</a></p><p>This link expires in 1 hour. If you didn't request this, you can ignore this email.</p><p>Link: %s</p>`, link, link)
	return SendEmailFunc(settings, to, subject, body)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// rate limit: 5/min per IP
var (
	rateMu   sync.Mutex
	rateHits = map[string][]time.Time{}
)

func rateLimited(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()
	now := time.Now()
	hits := rateHits[ip]
	// prune >1m
	var fresh []time.Time
	for _, t := range hits {
		if now.Sub(t) < time.Minute {
			fresh = append(fresh, t)
		}
	}
	if len(fresh) >= 5 {
		rateHits[ip] = fresh
		return true
	}
	fresh = append(fresh, now)
	rateHits[ip] = fresh
	return false
}

// RequestPasswordReset PUBLIC POST /api/forgot-password
func RequestPasswordReset(c *gin.Context) {
	if rateLimited(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		return
	}
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	_ = c.ShouldBindJSON(&req)
	email := strings.TrimSpace(req.Email)
	username := strings.TrimSpace(req.Username)

	// Always return 200 to avoid enumeration
	respondOK := func() {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}

	lookup := ""
	isEmail := false
	if email != "" {
		lookup = email
		isEmail = true
	} else if username != "" {
		lookup = username
	} else {
		respondOK()
		return
	}

	var userID int64
	var userEmail string
	var err error
	if isEmail {
		err = db.DB.QueryRow("SELECT id, COALESCE(email,'') FROM users WHERE LOWER(email)=LOWER(?) AND email != '' LIMIT 1", lookup).Scan(&userID, &userEmail)
	} else {
		err = db.DB.QueryRow("SELECT id, COALESCE(email,'') FROM users WHERE LOWER(username)=LOWER(?) LIMIT 1", lookup).Scan(&userID, &userEmail)
		// Also try email fallback if username looks like email? Keep enumeration-safe: only one lookup.
	}
	if err != nil || userEmail == "" {
		respondOK()
		return
	}

	// Support lookup by username where email field may match username input: already handled above with username branch.
	// If request provided email but user not found via email, try username as fallback for generic input handling
	if err != nil {
		respondOK()
		return
	}

	// Generate token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		middleware.LogError("password_reset", "token generation failed", "error", err)
		respondOK()
		return
	}
	token := hex.EncodeToString(b)
	tokenHash := hashToken(token)
	now := time.Now().Unix()
	expires := time.Now().Add(1 * time.Hour).Unix()

	_, err = db.DB.Exec("INSERT INTO password_reset_tokens(user_id, token_hash, expires_at, created_at) VALUES(?,?,?,?)", userID, tokenHash, expires, now)
	if err != nil {
		middleware.LogError("password_reset", "insert token failed", "error", err)
		respondOK()
		return
	}

	link := publicBaseURL(c) + "/reset-password?token=" + token
	if err := SendPasswordResetEmail(userEmail, link); err != nil {
		middleware.LogError("password_reset", "send email failed", "error", err)
		// Still return ok to avoid enumeration; optionally delete token? Keep it so user can retry.
	}
	respondOK()
}

// ResetPassword PUBLIC POST /api/reset-password
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
		Password    string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token := strings.TrimSpace(req.Token)
	newPass := req.NewPassword
	if newPass == "" {
		newPass = req.Password
	}
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}
	if len(newPass) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}
	tokenHash := hashToken(token)
	var id, userID, expiresAt int64
	var usedAt *int64
	err := db.DB.QueryRow("SELECT id, user_id, expires_at, used_at FROM password_reset_tokens WHERE token_hash=?", tokenHash).Scan(&id, &userID, &expiresAt, &usedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}
	if usedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}
	if time.Now().Unix() > expiresAt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	// Update password via ent-compatible column name is "password"
	_, err = db.DB.Exec("UPDATE users SET password=? WHERE id=?", string(hash), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}
	now := time.Now().Unix()
	db.DB.Exec("UPDATE password_reset_tokens SET used_at=? WHERE id=?", now, id)
	// Invalidate remaining tokens for user
	db.DB.Exec("UPDATE password_reset_tokens SET used_at=? WHERE user_id=? AND used_at IS NULL AND id != ?", now, userID, id)
	// Delete sessions for user
	db.DB.Exec("DELETE FROM auth_sessions WHERE user_id=?", userID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
