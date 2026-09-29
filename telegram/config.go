package telegram

import (
	"os"
	"strconv"
	"strings"

	"villum/crypto"
	"villum/db"
	"villum/middleware"
)

const (
	settingBotToken      = "telegram_bot_token"
	settingWebhookSecret = "telegram_webhook_secret"
	settingMode          = "telegram_mode"
	settingUpdateOffset  = "telegram_update_offset"
	settingGraceMinutes  = "telegram_grace_minutes"
	defaultGraceMinutes  = 30
)

func appSetting(key string) (string, bool) {
	var v string
	if err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key=?", key).Scan(&v); err != nil {
		return "", false
	}
	return v, true
}
func setAppSetting(key, value string) error {
	_, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings(key,value) VALUES(?,?)", key, value)
	return err
}

func GetBotToken() string {
	if v := os.Getenv("TELEGRAM_BOT_TOKEN"); v != "" {
		return v
	}
	val, ok := appSetting(settingBotToken)
	if !ok || val == "" {
		return ""
	}
	dec, err := crypto.Decrypt(val)
	if err != nil {
		middleware.LogWarn("telegram", "failed to decrypt bot token", "error", err)
		return val
	}
	return dec
}
func SetBotToken(token string) error {
	if token == "" {
		_, err := db.DB.Exec("DELETE FROM app_settings WHERE key=?", settingBotToken)
		return err
	}
	enc, err := crypto.Encrypt(token)
	if err != nil {
		return err
	}
	return setAppSetting(settingBotToken, enc)
}
func GetWebhookSecret() string {
	if v := os.Getenv("TELEGRAM_WEBHOOK_SECRET"); v != "" {
		return v
	}
	val, ok := appSetting(settingWebhookSecret)
	if !ok || val == "" {
		return ""
	}
	dec, err := crypto.Decrypt(val)
	if err != nil {
		middleware.LogWarn("telegram", "failed to decrypt webhook secret", "error", err)
		return val
	}
	return dec
}
func SetWebhookSecret(s string) error {
	if s == "" {
		_, err := db.DB.Exec("DELETE FROM app_settings WHERE key=?", settingWebhookSecret)
		return err
	}
	enc, err := crypto.Encrypt(s)
	if err != nil {
		return err
	}
	return setAppSetting(settingWebhookSecret, enc)
}
func GetMode() string {
	v, ok := appSetting(settingMode)
	if !ok || v == "" {
		return "off"
	}
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "off", "polling", "webhook", "auto":
		return v
	default:
		return "off"
	}
}
func SetMode(m string) error { return setAppSetting(settingMode, m) }

func GetUpdateOffset() int64 {
	v, ok := appSetting(settingUpdateOffset)
	if !ok {
		return 0
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}
func SetUpdateOffset(n int64) error {
	return setAppSetting(settingUpdateOffset, strconv.FormatInt(n, 10))
}

func GetGraceMinutes() int {
	v, ok := appSetting(settingGraceMinutes)
	if !ok {
		return defaultGraceMinutes
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultGraceMinutes
	}
	return n
}
func SetGraceMinutes(n int) error { return setAppSetting(settingGraceMinutes, strconv.Itoa(n)) }

func GetAPIBase() string {
	if v := os.Getenv("TELEGRAM_API_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://api.telegram.org"
}
func MaskToken(t string) string {
	if len(t) <= 5 {
		return "***"
	}
	// reveal only first 3 and last 2
	return t[:3] + "***" + t[len(t)-2:]
}
func EffectiveMode() string {
	m := GetMode()
	if m != "auto" {
		return m
	}
	baseURL := os.Getenv("BASE_URL")
	if strings.HasPrefix(baseURL, "https://") {
		return "webhook"
	}
	return "polling"
}

// EnsureWebhookState aligns webhook registration with effective mode on startup.
func EnsureWebhookState() {
	mode := EffectiveMode()
	if mode == "webhook" {
		base := os.Getenv("BASE_URL")
		secret := GetWebhookSecret()
		if base != "" && secret != "" && GetBotToken() != "" && strings.HasPrefix(base, "https://") {
			_ = SetWebhook(strings.TrimRight(base, "/")+"/api/telegram/webhook", secret)
		}
	} else if mode == "polling" {
		if GetBotToken() != "" {
			_ = DeleteWebhook()
		}
	}
}
