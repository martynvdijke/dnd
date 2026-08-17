package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

const defaultAutoSaveInterval = 12

func normalizedAutoSaveInterval(value any) int {
	n, ok := value.(float64)
	if !ok || n != float64(int(n)) {
		return defaultAutoSaveInterval
	}
	interval := int(n)
	if interval < 5 {
		return 5
	}
	if interval > 300 {
		return 300
	}
	return interval
}

func GetAutoSaveSetting(c *gin.Context) {
	interval := defaultAutoSaveInterval
	var value string
	if err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key = 'autosave_interval'").Scan(&value); err == nil {
		var decoded any
		if json.Unmarshal([]byte(value), &decoded) == nil {
			interval = normalizedAutoSaveInterval(decoded)
		}
	}
	c.JSON(http.StatusOK, gin.H{"interval": interval})
}

func SetAutoSaveSetting(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	interval := defaultAutoSaveInterval
	if raw, ok := req["interval"]; ok {
		interval = normalizedAutoSaveInterval(raw)
	}
	if _, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('autosave_interval', ?)", interval); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interval": interval})
}

func GetEmailSettings(c *gin.Context) {
	var s models.EmailSettings
	var enabled int
	var password string
	err := db.DB.QueryRow("SELECT id, COALESCE(smtp_host, ''), COALESCE(smtp_port, 587), COALESCE(username, ''), COALESCE(password, ''), COALESCE(from_addr, ''), COALESCE(enabled, 0) FROM email_settings WHERE id = 1").
		Scan(&s.ID, &s.SMTPHost, &s.SMTPPort, &s.Username, &password, &s.FromAddr, &enabled)
	if err != nil {
		s = models.EmailSettings{ID: 1, SMTPPort: 587, Enabled: false}
		c.JSON(http.StatusOK, s)
		return
	}
	s.Enabled = enabled == 1
	c.JSON(http.StatusOK, gin.H{"id": s.ID, "smtp_host": s.SMTPHost, "smtp_port": s.SMTPPort, "username": s.Username, "has_password": password != "", "from_addr": s.FromAddr, "enabled": s.Enabled})
}

func SaveEmailSettings(c *gin.Context) {
	var s models.EmailSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if s.Password == "" {
		var existingPw string
		db.DB.QueryRow("SELECT COALESCE(password, '') FROM email_settings WHERE id = 1").Scan(&existingPw)
		s.Password = existingPw
	}

	enabled := 0
	if s.Enabled {
		enabled = 1
	}

	_, err := db.DB.Exec(`INSERT OR REPLACE INTO email_settings (id, smtp_host, smtp_port, username, password, from_addr, enabled) VALUES (1, ?, ?, ?, ?, ?, ?)`,
		s.SMTPHost, s.SMTPPort, s.Username, s.Password, s.FromAddr, enabled)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetEinkSetting returns the site-wide e-ink mode setting.
func GetEinkSetting(c *gin.Context) {
	var value string
	err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key = 'eink'").Scan(&value)
	if err != nil {
		value = "0"
	}
	c.JSON(http.StatusOK, gin.H{"enabled": value == "1"})
}

// SetEinkSetting persists the site-wide e-ink mode setting.
func SetEinkSetting(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	value := "0"
	if req.Enabled {
		value = "1"
	}
	if _, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES ('eink', ?)", value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": req.Enabled})
}
