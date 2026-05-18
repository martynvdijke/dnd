package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/ent"
	"villum/ent/user"
	"villum/models"
)

func AdminListUsers(c *gin.Context) {
	users, err := db.Client.User.Query().Order(ent.Asc(user.FieldID)).All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]models.User, len(users))
	for i, u := range users {
		result[i] = models.User{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Role:        u.Role,
			Email:       u.Email,
			CreatedAt:   u.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, result)
}

func AdminCreateUser(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Email       string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	u, err := db.Client.User.Create().
		SetUsername(req.Username).
		SetPassword(string(hash)).
		SetDisplayName(req.DisplayName).
		SetRole(req.Role).
		SetEmail(req.Email).
		Save(c.Request.Context())
	if err != nil {
		if ent.IsConstraintError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "username may already exist"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "username may already exist"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID})
}

func AdminUpdateUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Email       string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Client.User.UpdateOneID(id).
		SetUsername(req.Username).
		SetDisplayName(req.DisplayName).
		SetRole(req.Role).
		SetEmail(req.Email).
		Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminDeleteUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	currentUserID, _ := c.Get("user_id")
	if currentUserID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}
	db.Client.User.DeleteOneID(id).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminResetPassword(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
		return
	}
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash"})
		return
	}
	db.Client.User.UpdateOneID(id).SetPassword(string(hash)).Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetBackupSettings(c *gin.Context) {
	s, err := db.Client.BackupSetting.Query().Only(c.Request.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"enabled": true, "interval_days": 7, "keep_count": 7, "last_backup": ""})
			return
		}
		c.JSON(http.StatusOK, gin.H{"enabled": true, "interval_days": 7, "keep_count": 7, "last_backup": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": s.Enabled, "interval_days": s.IntervalDays, "keep_count": s.KeepCount, "last_backup": s.LastBackup})
}

func SaveBackupSettings(c *gin.Context) {
	var req struct {
		Enabled      bool `json:"enabled"`
		IntervalDays int  `json:"interval_days"`
		KeepCount    int  `json:"keep_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IntervalDays < 1 {
		req.IntervalDays = 7
	}
	if req.KeepCount < 1 {
		req.KeepCount = 7
	}
	db.Client.BackupSetting.Create().
		SetID(1).
		SetEnabled(req.Enabled).
		SetIntervalDays(req.IntervalDays).
		SetKeepCount(req.KeepCount).
		OnConflict().
		UpdateEnabled().
		UpdateIntervalDays().
		UpdateKeepCount().
		Exec(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TriggerBackup(c *gin.Context) {
	backupPath, err := CreateBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	PruneBackups()
	c.JSON(http.StatusOK, gin.H{"path": backupPath})
}

func ListBackups(c *gin.Context) {
	backupDir := "backups"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusOK, []string{})
		return
	}
	type BackupEntry struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	var backups []BackupEntry
	for _, e := range entries {
		if info, err := e.Info(); err == nil {
			backups = append(backups, BackupEntry{Name: e.Name(), Size: info.Size()})
		}
	}
	c.JSON(http.StatusOK, backups)
}
