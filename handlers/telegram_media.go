package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/telegram"
)

// GetTelegramTargets returns bound campaigns and DM link state.
func GetTelegramTargets(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	rows, err := db.DB.Query(`
		SELECT c.id, c.name, COALESCE(cts.title_cache,'')
		FROM campaigns c
		JOIN campaign_telegram_settings cts ON cts.campaign_id = c.id
		WHERE cts.is_enabled = 1 AND cts.chat_id IS NOT NULL
		  AND (c.user_id = ? OR EXISTS (SELECT 1 FROM campaign_members cm WHERE cm.campaign_id = c.id AND cm.user_id = ?))
		ORDER BY c.name`, userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	type camp struct {
		ID        int64  `json:"id"`
		Title     string `json:"title"`
		ChatTitle string `json:"chat_title"`
	}
	camps := []camp{}
	for rows.Next() {
		var id int64
		var title, chatTitle string
		if err := rows.Scan(&id, &title, &chatTitle); err != nil {
			continue
		}
		camps = append(camps, camp{ID: id, Title: title, ChatTitle: chatTitle})
	}
	_, _, _, linked := telegram.GetIdentityByUserID(userID)
	// Ensure empty slice marshals as []
	if camps == nil {
		camps = []camp{}
	}
	c.JSON(http.StatusOK, gin.H{
		"campaigns": camps,
		"dm":        gin.H{"linked": linked},
	})
}

// PostTelegramMedia handles POST /api/telegram/media
func PostTelegramMedia(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	var req struct {
		UploadIDs []int64 `json:"upload_ids"`
		Target    *struct {
			Kind       string `json:"kind"`
			CampaignID *int64 `json:"campaign_id"`
		} `json:"target"`
		RecapID *int64 `json:"recap_id"`
		Caption string `json:"caption"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	if len(req.UploadIDs) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "upload_ids required"})
		return
	}
	if req.Target == nil || (req.Target.Kind != "campaign" && req.Target.Kind != "dm") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	// Validate all upload IDs exist and are accessible before sending.
	if !canAccessUploadIDs(c, req.UploadIDs) {
		return
	}

	var targetChatID int64
	var targetType string
	var campaignID *int64

	if req.Target.Kind == "campaign" {
		if req.Target.CampaignID == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "campaign_id required"})
			return
		}
		cid := *req.Target.CampaignID
		campaignID = &cid
		// Must be DM of that campaign.
		if !isCampaignDM(c, cid) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		s, found, err := telegram.GetCampaignTelegramSettings(cid)
		if err != nil || !found || s.ChatID == nil || !s.IsEnabled {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "campaign not bound"})
			return
		}
		targetChatID = *s.ChatID
		targetType = "chat"
	} else {
		// dm
		tgID, _, _, ok := telegram.GetIdentityByUserID(userID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "not linked"})
			return
		}
		// Use telegram_user_id as chat id for DM (matches TestTelegram).
		targetChatID = tgID
		targetType = "dm"
	}

	results, err := telegram.PostMedia(targetChatID, targetType, campaignID, req.UploadIDs, req.RecapID, req.Caption)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
