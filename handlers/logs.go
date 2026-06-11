package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"

	"villum/db"
	"villum/middleware"
)

// ─── Log Query ───

// ListLogs returns log entries from the ring buffer with optional filters.
// Query params: level, source, since (RFC3339), limit (int)
func ListLogs(c *gin.Context) {
	if middleware.AppLog == nil {
		c.JSON(http.StatusOK, []middleware.LogEntry{})
		return
	}

	level := c.Query("level")
	source := c.Query("source")

	var since time.Time
	if s := c.Query("since"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			since = t
		}
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	entries := middleware.AppLog.Buffer().Query(level, source, since, limit)
	if entries == nil {
		entries = []middleware.LogEntry{}
	}
	c.JSON(http.StatusOK, entries)
}

// ─── Log Sources ───

// ListLogSources returns the distinct set of source names from the ring buffer.
func ListLogSources(c *gin.Context) {
	if middleware.AppLog == nil || middleware.AppLog.Buffer() == nil {
		c.JSON(http.StatusOK, []string{})
		return
	}
	sources := middleware.AppLog.Buffer().Sources()
	if sources == nil {
		sources = []string{}
	}
	c.JSON(http.StatusOK, sources)
}

// ─── Log Level ───

// logLevelResponse is the JSON response for log level queries.
type logLevelResponse struct {
	Level     string `json:"level"`
	Effective string `json:"effective"`
}

// GetLogLevel returns the current minimum log level.
func GetLogLevel(c *gin.Context) {
	level := "warn"
	err := db.DB.QueryRow("SELECT value FROM log_settings WHERE key='min_level'").Scan(&level)
	if err != nil {
		level = "warn"
	}
	effective := "warn"
	if middleware.AppLog != nil {
		effective = middleware.StringFromLogLevel(middleware.AppLog.MinLevel())
	}
	c.JSON(http.StatusOK, logLevelResponse{Level: level, Effective: effective})
}

// SetLogLevelRequest is the JSON body for updating the minimum log level.
type SetLogLevelRequest struct {
	Level string `json:"level"`
}

// SetLogLevel updates the minimum log level.
func SetLogLevel(c *gin.Context) {
	var req SetLogLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[req.Level] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "level must be one of: debug, info, warn, error"})
		return
	}

	// Persist to DB
	_, err := db.DB.Exec("INSERT OR REPLACE INTO log_settings (key, value) VALUES ('min_level', ?)", req.Level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save log level"})
		return
	}

	// Update live handler
	if middleware.AppLog != nil {
		middleware.AppLog.SetMinLevel(middleware.LogLevelFromString(req.Level))
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "level": req.Level})
}

// InitLogSettings creates the log_settings table if it doesn't exist and
// loads the persisted log level into AppLog.
func InitLogSettings() {
	_, err := db.DB.Exec("CREATE TABLE IF NOT EXISTS log_settings (key TEXT PRIMARY KEY, value TEXT)")
	if err != nil {
		return
	}
	var level string
	err = db.DB.QueryRow("SELECT value FROM log_settings WHERE key='min_level'").Scan(&level)
	if err != nil || level == "" {
		level = "warn"
	}
	if middleware.AppLog != nil {
		middleware.AppLog.SetMinLevel(slog.Level(middleware.LogLevelFromString(level)))
	}
}
