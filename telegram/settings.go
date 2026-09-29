package telegram

import (
	"database/sql"
	"errors"
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

	defaultGraceMinutes = 30
	defaultAPIBase      = "https://api.telegram.org"
)

// appSetting reads a raw value from the app_settings key-value store.
func appSetting(key string) (string, bool) {
	var v string
	if err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key=?", key).Scan(&v); err != nil {
		return "", false
	}
	return v, true
}

// setAppSetting writes a raw value to the app_settings key-value store.
func setAppSetting(key, value string) error {
	_, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings(key,value) VALUES(?,?)", key, value)
	return err
}

// deleteAppSetting removes a key from the app_settings key-value store.
func deleteAppSetting(key string) error {
	_, err := db.DB.Exec("DELETE FROM app_settings WHERE key=?", key)
	return err
}

// Settings is the single resolved view of the bot's stored configuration.
// Every package reads settings through LoadSettings (or its typed accessors)
// instead of querying app_settings directly.
type Settings struct {
	Token         string
	WebhookSecret string
	Mode          string // stored mode: off | polling | webhook | auto
	EffectiveMode string // resolved mode: off | polling | webhook
	UpdateOffset  int64
	GraceMinutes  int
	APIBase       string
}

// LoadSettings resolves every bot setting, including environment overrides.
func LoadSettings() Settings {
	return Settings{
		Token:         GetBotToken(),
		WebhookSecret: GetWebhookSecret(),
		Mode:          GetMode(),
		EffectiveMode: EffectiveMode(),
		UpdateOffset:  GetUpdateOffset(),
		GraceMinutes:  GetGraceMinutes(),
		APIBase:       GetAPIBase(),
	}
}

// CampaignTelegramSettings is the typed view of one campaign's Telegram
// delivery configuration (campaign_telegram_settings).
type CampaignTelegramSettings struct {
	CampaignID      int64
	ChatID          *int64
	ChatType        string
	TitleCache      string
	IsEnabled       bool
	AutoPostEnabled bool
	BoundAt         *string
	BoundByUserID   *int64
}

// GetCampaignTelegramSettings reads a campaign's Telegram settings. The bool
// reports whether a row exists; a missing row is not an error.
func GetCampaignTelegramSettings(campaignID int64) (CampaignTelegramSettings, bool, error) {
	s := CampaignTelegramSettings{CampaignID: campaignID}
	var chatID, boundBy sql.NullInt64
	var chatType, titleCache, boundAt sql.NullString
	var isEnabled, autoPost int
	err := db.DB.QueryRow(`
		SELECT chat_id, chat_type, title_cache, is_enabled, auto_post_enabled, bound_at, bound_by_user_id
		FROM campaign_telegram_settings WHERE campaign_id = ?`, campaignID).
		Scan(&chatID, &chatType, &titleCache, &isEnabled, &autoPost, &boundAt, &boundBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, false, nil
		}
		return s, false, err
	}
	if chatID.Valid {
		s.ChatID = &chatID.Int64
	}
	if boundBy.Valid {
		s.BoundByUserID = &boundBy.Int64
	}
	if boundAt.Valid {
		s.BoundAt = &boundAt.String
	}
	s.ChatType = chatType.String
	s.TitleCache = titleCache.String
	s.IsEnabled = isEnabled != 0
	s.AutoPostEnabled = autoPost != 0
	return s, true, nil
}

// UpsertCampaignTelegramSettings writes a campaign's Telegram settings row.
func UpsertCampaignTelegramSettings(s CampaignTelegramSettings) error {
	var chatID, boundBy any
	if s.ChatID != nil {
		chatID = *s.ChatID
	}
	if s.BoundByUserID != nil {
		boundBy = *s.BoundByUserID
	}
	enabled := 0
	if s.IsEnabled {
		enabled = 1
	}
	autoPost := 0
	if s.AutoPostEnabled {
		autoPost = 1
	}
	var boundAt any
	if s.BoundAt != nil {
		boundAt = *s.BoundAt
	}
	_, err := db.DB.Exec(`
		INSERT INTO campaign_telegram_settings
			(campaign_id, chat_id, chat_type, title_cache, is_enabled, auto_post_enabled, bound_at, bound_by_user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(campaign_id) DO UPDATE SET
			chat_id = excluded.chat_id,
			chat_type = excluded.chat_type,
			title_cache = excluded.title_cache,
			is_enabled = excluded.is_enabled,
			auto_post_enabled = excluded.auto_post_enabled,
			bound_at = excluded.bound_at,
			bound_by_user_id = excluded.bound_by_user_id`,
		s.CampaignID, chatID, s.ChatType, s.TitleCache, enabled, autoPost, boundAt, boundBy)
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
		return deleteAppSetting(settingBotToken)
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
		return deleteAppSetting(settingWebhookSecret)
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
	return defaultAPIBase
}

func MaskToken(t string) string {
	if len(t) <= 5 {
		return "***"
	}
	// reveal only first 3 and last 2
	return t[:3] + "***" + t[len(t)-2:]
}

// EffectiveMode resolves the configured mode against the public base URL:
// auto becomes webhook when BASE_URL is public HTTPS, otherwise polling.
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

// EnsureWebhookState is kept for callers that expect the old startup hook.
// Webhook registration now converges inside the bot supervisor, so this is a
// no-op and exists only to avoid breaking startup wiring.
func EnsureWebhookState() {}

// WebhookURL returns the public webhook endpoint derived from BASE_URL, or ""
// when no public HTTPS base URL is configured.
func WebhookURL() string {
	base := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	if !strings.HasPrefix(base, "https://") {
		return ""
	}
	return base + "/api/telegram/webhook"
}
