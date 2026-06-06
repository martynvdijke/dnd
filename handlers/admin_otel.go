package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func GetOTelSettings(c *gin.Context) {
	var s models.OTelSettings
	var enabled int
	err := db.DB.QueryRow(`SELECT id, COALESCE(endpoint, ''), COALESCE(enabled, 0) FROM otel_settings WHERE id = 1`).
		Scan(&s.ID, &s.Endpoint, &enabled)
	if err != nil {
		s = models.OTelSettings{ID: 1, Enabled: false}
		c.JSON(http.StatusOK, s)
		return
	}
	s.Enabled = enabled == 1
	c.JSON(http.StatusOK, s)
}

func SaveOTelSettings(c *gin.Context) {
	var s models.OTelSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := 0
	if s.Enabled {
		enabled = 1
	}

	_, err := db.DB.Exec(`INSERT OR REPLACE INTO otel_settings (id, endpoint, enabled) VALUES (1, ?, ?)`,
		s.Endpoint, enabled)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
