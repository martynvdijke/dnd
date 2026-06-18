package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
)

var startTime = time.Now()

func HandleHealth(c *gin.Context) {
	err := db.DB.Ping()
	dbStatus := "ok"
	if err != nil {
		dbStatus = "error: " + err.Error()
	}
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"database":   dbStatus,
		"uptime":     time.Since(startTime).String(),
		"go_version": runtime.Version(),
		"version":    "2.18.1",
	})
}

func HandleMetrics(c *gin.Context) {
	var userCount, charCount, campCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM characters").Scan(&charCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM campaigns").Scan(&campCount)

	c.JSON(http.StatusOK, gin.H{
		"users":          userCount,
		"characters":     charCount,
		"campaigns":      campCount,
		"uptime_seconds": int(time.Since(startTime).Seconds()),
	})
}
