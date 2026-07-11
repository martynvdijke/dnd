package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"villum/models"
)

func GetOTelSettings(c *gin.Context) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	c.JSON(http.StatusOK, models.OTelSettings{
		ID:       1,
		Enabled:  endpoint != "",
		Endpoint: endpoint,
	})
}

func SaveOTelSettings(c *gin.Context) {
	// OTel settings are now configured via environment variables
	// (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_PROTOCOL, etc.).
	// The admin UI displays current configuration; changes must be made via env vars.
	c.JSON(http.StatusOK, gin.H{"status": "read_only", "message": "OTel configuration is managed via environment variables"})
}
