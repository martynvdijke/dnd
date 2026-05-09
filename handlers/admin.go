package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/models"
)

func AdminListUsers(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, username, display_name, role, created_at FROM users ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Role, &u.CreatedAt)
		users = append(users, u)
	}
	c.JSON(http.StatusOK, users)
}

func AdminCreateUser(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
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
	result, err := db.DB.Exec("INSERT INTO users(username,password,display_name,role) VALUES(?,?,?,?)",
		req.Username, string(hash), req.DisplayName, req.Role)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username may already exist"})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func AdminUpdateUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("UPDATE users SET username=?, display_name=?, role=? WHERE id=?",
		req.Username, req.DisplayName, req.Role, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminDeleteUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// Don't allow deleting self
	currentUserID, _ := c.Get("user_id")
	if currentUserID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}
	db.DB.Exec("DELETE FROM users WHERE id=?", id)
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
	db.DB.Exec("UPDATE users SET password=? WHERE id=?", string(hash), id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Backup settings

func GetBackupSettings(c *gin.Context) {
	var enabled bool
	var interval int
	var lastBackup string
	err := db.DB.QueryRow("SELECT enabled, interval_hours, last_backup FROM backup_settings WHERE id=1").Scan(&enabled, &interval, &lastBackup)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"enabled": true, "interval_hours": 168, "last_backup": ""})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"enabled": true, "interval_hours": 168, "last_backup": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": enabled, "interval_hours": interval, "last_backup": lastBackup})
}

func SaveBackupSettings(c *gin.Context) {
	var req struct {
		Enabled       bool `json:"enabled"`
		IntervalHours int  `json:"interval_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec(`INSERT INTO backup_settings(id,enabled,interval_hours,last_backup) VALUES(1,?,?,'')
		ON CONFLICT(id) DO UPDATE SET enabled=excluded.enabled, interval_hours=excluded.interval_hours`,
		req.Enabled, req.IntervalHours)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TriggerBackup(c *gin.Context) {
	backupPath, err := CreateBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": backupPath})
}

func ListBackups(c *gin.Context) {
	// List backup files from the backups directory
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
