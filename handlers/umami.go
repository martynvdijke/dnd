package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

func GetUmamiSettings(c *gin.Context) {
	var s models.UmamiSettings
	var enabled, shareData, enableAdminTracking int
	err := db.DB.QueryRow(`SELECT id, COALESCE(enabled, 0), COALESCE(tracker_hostname, ''), COALESCE(website_id, ''), COALESCE(share_data, 0), COALESCE(enable_admin_tracking, 0) FROM umami_settings WHERE id = 1`).
		Scan(&s.ID, &enabled, &s.TrackerHostname, &s.WebsiteID, &shareData, &enableAdminTracking)
	if err != nil {
		s = models.UmamiSettings{ID: 1, Enabled: false, ShareData: false, EnableAdminTracking: false}
		c.JSON(http.StatusOK, s)
		return
	}
	s.Enabled = enabled == 1
	s.ShareData = shareData == 1
	s.EnableAdminTracking = enableAdminTracking == 1
	c.JSON(http.StatusOK, s)
}

func SaveUmamiSettings(c *gin.Context) {
	var s models.UmamiSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	shareData := 0
	if s.ShareData {
		shareData = 1
	}
	enableAdminTracking := 0
	if s.EnableAdminTracking {
		enableAdminTracking = 1
	}

	_, err := db.DB.Exec(`INSERT OR REPLACE INTO umami_settings (id, enabled, tracker_hostname, website_id, share_data, enable_admin_tracking) VALUES (1, ?, ?, ?, ?, ?)`,
		enabled, s.TrackerHostname, s.WebsiteID, shareData, enableAdminTracking)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// InjectUmamiScript reads Umami settings and returns the tracking script tag
// if analytics is enabled and configured. Returns empty string otherwise.
func InjectUmamiScript() string {
	var enabled, shareData int
	var trackerHostname, websiteID string
	err := db.DB.QueryRow(`SELECT enabled, tracker_hostname, website_id, share_data FROM umami_settings WHERE id = 1`).
		Scan(&enabled, &trackerHostname, &websiteID, &shareData)
	if err != nil {
		return ""
	}
	if enabled != 1 || trackerHostname == "" || websiteID == "" {
		return ""
	}
	attrs := fmt.Sprintf(`async defer src="%s/script.js" data-website-id="%s"`, trackerHostname, websiteID)
	if shareData == 1 {
		attrs += ` data-share-data="true"`
	}
	return fmt.Sprintf(`<script %s></script>`, attrs)
}
