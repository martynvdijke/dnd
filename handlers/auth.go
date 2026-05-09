package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/middleware"
)

func CheckSetup(c *gin.Context) {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role='admin'").Scan(&count)
	c.JSON(http.StatusOK, gin.H{"setup": count > 0})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Setup    bool   `json:"setup"`
}

func HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Setup {
		// First-time setup - create admin user
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role='admin'").Scan(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "admin already exists"})
			return
		}
		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		result, err := db.DB.Exec("INSERT INTO users(username,password,display_name,role) VALUES(?,?,'Admin','admin')",
			req.Username, string(hash))
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		userID, _ := result.LastInsertId()
		sessionID := middleware.Store.Create(userID, req.Username, "admin", c.ClientIP())
		c.SetCookie("session", sessionID, 86400, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"session": sessionID, "user": gin.H{"id": userID, "username": req.Username, "role": "admin"}})
		return
	}

	// Normal login
	var userID int64
	var username, password, role string
	err := db.DB.QueryRow("SELECT id, username, password, role FROM users WHERE username=?", req.Username).Scan(&userID, &username, &password, &role)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	sessionID := middleware.Store.Create(userID, username, role, c.ClientIP())
	c.SetCookie("session", sessionID, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"session": sessionID, "user": gin.H{"id": userID, "username": username, "role": role}})
}

func HandleLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("session")
	if sessionID != "" {
		middleware.Store.Delete(sessionID)
	}
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
		"role":     role,
	})
}

func GetCSRFToken(c *gin.Context) {
	sessionID, _ := c.Cookie("session")
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": middleware.CSRFHash(sessionID)})
}
