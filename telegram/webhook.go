package telegram

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"villum/middleware"

	"github.com/gin-gonic/gin"
	tgmodels "github.com/go-telegram/bot/models"
)

func WebhookHandler(c *gin.Context) {
	secret := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
	expected := GetWebhookSecret()
	if expected == "" || subtle.ConstantTimeCompare([]byte(secret), []byte(expected)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var upd tgmodels.Update
	if err := json.NewDecoder(c.Request.Body).Decode(&upd); err != nil {
		middleware.LogWarn("telegram", "webhook decode failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	ProcessUpdate(&upd)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
