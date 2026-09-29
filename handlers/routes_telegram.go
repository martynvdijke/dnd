package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/middleware"
	"villum/telegram"
)

func RegisterTelegramRoutes(r *gin.RouterGroup) {
	r.POST("/telegram/link-code", CreateTelegramLinkCode)
	r.GET("/telegram/status", GetTelegramStatus)
	r.DELETE("/telegram/unlink", UnlinkTelegram)
	r.PUT("/telegram/prefs", SetTelegramPrefs)
	r.GET("/campaigns/:id/telegram", GetCampaignTelegram)
	r.PUT("/campaigns/:id/telegram", SetCampaignTelegram)
}
func RegisterTelegramAdminRoutes(r *gin.RouterGroup) {
	r.GET("/telegram-settings", GetTelegramSettings)
	r.POST("/telegram-settings", SaveTelegramSettings)
	r.POST("/telegram-test", TestTelegram)
}
func RegisterTelegramPublicRoutes(r *gin.Engine) {
	r.POST("/api/telegram/webhook", telegram.WebhookHandler)
}

func CreateTelegramLinkCode(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)
	var cnt int
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM telegram_link_codes WHERE user_id=? AND created_at > datetime('now','-15 minutes')", userID).Scan(&cnt)
	if cnt >= 5 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many codes"})
		return
	}
	code, exp, err := telegram.CreateLinkCode(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	bot := telegram.GetBotUsername()
	if bot == "" {
		bot = os.Getenv("TELEGRAM_BOT_USERNAME")
	}
	url := ""
	if bot != "" {
		url = "https://t.me/" + bot + "?start=" + code
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "url": url, "expires_at": exp})
}
func GetTelegramStatus(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)
	tgID, username, dm, ok := telegram.GetIdentityByUserID(userID)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"linked": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"linked": true, "telegram_user_id": tgID, "telegram_username": username, "dm_enabled": dm == 1})
}
func UnlinkTelegram(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)
	_ = telegram.DeleteIdentityByUserID(userID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
func SetTelegramPrefs(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)
	var req struct {
		DMEnabled *bool `json:"dm_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.DMEnabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	if err := telegram.SetDMEnabled(userID, *req.DMEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "dm_enabled": *req.DMEnabled})
}
func GetCampaignTelegram(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isCampaignMember(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var chatID *int64
	var chatType, titleCache string
	var isEnabled, autoPost int
	var boundAt *string
	err := db.DB.QueryRow("SELECT chat_id, chat_type, title_cache, is_enabled, auto_post_enabled, bound_at FROM campaign_telegram_settings WHERE campaign_id=?", id).Scan(&chatID, &chatType, &titleCache, &isEnabled, &autoPost, &boundAt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"chat_id": nil, "is_enabled": false, "auto_post_enabled": false, "bound_at": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chat_id": chatID, "chat_type": chatType, "title_cache": titleCache, "is_enabled": isEnabled == 1, "auto_post_enabled": autoPost == 1, "bound_at": boundAt})
}
func SetCampaignTelegram(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !isCampaignDM(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var req struct {
		ChatID          *int64 `json:"chat_id"`
		IsEnabled       *bool  `json:"is_enabled"`
		AutoPostEnabled *bool  `json:"auto_post_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	uid, _ := c.Get("user_id")
	userID := uid.(int64)
	var isEnabled, autoPost int
	if req.IsEnabled != nil && *req.IsEnabled {
		isEnabled = 1
	}
	if req.AutoPostEnabled != nil && *req.AutoPostEnabled {
		autoPost = 1
	}
	var exChatID *int64
	var exType, exTitle string
	var exEn, exAuto int
	_ = db.DB.QueryRow("SELECT chat_id, chat_type, title_cache, is_enabled, auto_post_enabled FROM campaign_telegram_settings WHERE campaign_id=?", id).Scan(&exChatID, &exType, &exTitle, &exEn, &exAuto)
	if req.ChatID == nil {
		req.ChatID = exChatID
	}
	if req.IsEnabled == nil {
		isEnabled = exEn
	}
	if req.AutoPostEnabled == nil {
		autoPost = exAuto
	}
	chatType := exType
	titleCache := exTitle
	if req.ChatID != nil {
		if ch, err := telegram.GetChat(*req.ChatID); err == nil {
			chatType = ch.Type
			titleCache = ch.Title
		}
	}
	_, err := db.DB.Exec(`INSERT INTO campaign_telegram_settings(campaign_id,chat_id,chat_type,title_cache,is_enabled,auto_post_enabled,bound_at,bound_by_user_id)
		VALUES(?,?,?,?,?,?,datetime('now'),?)
		ON CONFLICT(campaign_id) DO UPDATE SET chat_id=excluded.chat_id, chat_type=excluded.chat_type, title_cache=excluded.title_cache, is_enabled=excluded.is_enabled, auto_post_enabled=excluded.auto_post_enabled, bound_at=excluded.bound_at, bound_by_user_id=excluded.bound_by_user_id`,
		id, req.ChatID, chatType, titleCache, isEnabled, autoPost, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func telegramSettingsPayload() gin.H {
	token := telegram.GetBotToken()
	hasToken := token != ""
	masked := ""
	if hasToken {
		masked = telegram.MaskToken(token)
	}
	mode := telegram.GetMode()
	effective := telegram.EffectiveMode()
	base := os.Getenv("BASE_URL")
	webhookURL := ""
	if base != "" {
		webhookURL = strings.TrimRight(base, "/") + "/api/telegram/webhook"
	}
	off := telegram.GetUpdateOffset()
	grace := telegram.GetGraceMinutes()
	payload := gin.H{
		"has_token": hasToken, "token_masked": masked,
		"mode": mode, "effective_mode": effective,
		"webhook_url":   webhookURL,
		"update_offset": off,
		"grace_minutes": grace,
		"status": gin.H{
			"mode":          effective,
			"webhook_url":   webhookURL,
			"update_offset": off,
		},
	}
	return payload
}

func GetTelegramSettings(c *gin.Context) {
	c.JSON(http.StatusOK, telegramSettingsPayload())
}
func SaveTelegramSettings(c *gin.Context) {
	var req struct {
		Token         *string `json:"token"`
		Mode          *string `json:"mode"`
		WebhookSecret *string `json:"webhook_secret"`
		GraceMinutes  *int    `json:"grace_minutes"`
		ClearToken    *bool   `json:"clear_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	if req.ClearToken != nil && *req.ClearToken {
		_ = telegram.SetBotToken("")
		_ = telegram.DeleteWebhook()
	} else if req.Token != nil {
		if *req.Token == "" {
			_ = telegram.SetBotToken("")
		} else {
			_ = telegram.SetBotToken(*req.Token)
		}
	}
	if req.Mode != nil {
		_ = telegram.SetMode(*req.Mode)
	}
	if req.WebhookSecret != nil {
		if *req.WebhookSecret == "" {
			_ = telegram.SetWebhookSecret("")
		} else {
			_ = telegram.SetWebhookSecret(*req.WebhookSecret)
		}
	}
	if req.GraceMinutes != nil {
		_ = telegram.SetGraceMinutes(*req.GraceMinutes)
	}
	mode := telegram.EffectiveMode()
	if req.ClearToken != nil && *req.ClearToken {
		// already deleted webhook
	} else if mode == "webhook" {
		base := os.Getenv("BASE_URL")
		secret := telegram.GetWebhookSecret()
		if base != "" && secret != "" {
			_ = telegram.SetWebhook(strings.TrimRight(base, "/")+"/api/telegram/webhook", secret)
		}
	} else if mode == "polling" {
		_ = telegram.DeleteWebhook()
	}
	c.JSON(http.StatusOK, telegramSettingsPayload())
}
func TestTelegram(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)
	tgChatID, _, _, ok := telegram.GetIdentityByUserID(userID)
	if !ok || tgChatID == 0 {
		c.JSON(http.StatusOK, gin.H{"sent": false, "error": "not linked"})
		return
	}
	if _, err := telegram.SendMessage(tgChatID, "Villum test message"); err != nil {
		middleware.LogWarn("telegram", "test send failed", "error", err)
		c.JSON(http.StatusOK, gin.H{"sent": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": true})
}
