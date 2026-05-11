package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

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
