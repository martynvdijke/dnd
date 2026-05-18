package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/ent"
	"villum/ent/user"
	"villum/middleware"
)

func CheckSetup(c *gin.Context) {
	count, err := db.Client.User.Query().Where(user.Role("admin")).Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
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
		count, err := db.Client.User.Query().Where(user.Role("admin")).Count(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
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
		u, err := db.Client.User.Create().
			SetUsername(req.Username).
			SetPassword(string(hash)).
			SetDisplayName("Admin").
			SetRole("admin").
			Save(c.Request.Context())
		if err != nil {
			if ent.IsConstraintError(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
		sessionID := middleware.Store.Create(u.ID, req.Username, "admin", c.ClientIP())
		c.SetCookie("session", sessionID, 86400, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"session": sessionID, "user": gin.H{"id": u.ID, "username": req.Username, "role": "admin"}})
		return
	}

	u, err := db.Client.User.Query().Where(user.Username(req.Username)).Only(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	sessionID := middleware.Store.Create(u.ID, u.Username, u.Role, c.ClientIP())
	c.SetCookie("session", sessionID, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"session": sessionID, "user": gin.H{"id": u.ID, "username": u.Username, "role": u.Role}})
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
