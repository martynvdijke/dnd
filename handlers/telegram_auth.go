package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/ent"
	"villum/ent/user"
	"villum/middleware"
	"villum/telegram"
)

// telegramSelfRegisterEnabled reports whether unknown emails may create a
// player account. TELEGRAM_SELF_REGISTER defaults to on.
func telegramSelfRegisterEnabled() bool {
	v := strings.TrimSpace(os.Getenv("TELEGRAM_SELF_REGISTER"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "0", "false", "off", "no":
		return false
	}
	return true
}

// SendTelegramAuthLinkEmail builds the redemption URL and sends it by email.
// It is injected into the telegram package via telegram.SetAuthLinkSender.
func SendTelegramAuthLinkEmail(email, token string) error {
	base := strings.TrimRight(BaseURL, "/")
	if base == "" {
		return fmt.Errorf("BASE_URL not configured")
	}
	settings, err := getEmailSettings()
	if err != nil {
		return err
	}
	link := base + "/telegram/auth?token=" + url.QueryEscape(token)
	body := "Follow this link to sign in to Villum and link your Telegram account:\n\n" +
		link + "\n\n" +
		"This link works once and expires in 15 minutes. If you did not request it, you can ignore this email."
	return SendEmailFunc(settings, email, "Your Villum sign-in link", body)
}

// HandleTelegramAuthLink redeems an emailed sign-in token: it links the
// Telegram identity, opens a real web session, and (when enabled) self-registers
// an unknown email as a player.
func HandleTelegramAuthLink(c *gin.Context) {
	c.Header("Referrer-Policy", "no-referrer")

	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		renderTelegramAuthPage(c, http.StatusBadRequest, "Link unavailable", "This sign-in link is invalid or has expired. Return to Telegram and send /login for a new one.")
		return
	}
	t, ok := telegram.ConsumeAuthToken(telegram.HashAuthToken(token))
	if !ok {
		renderTelegramAuthPage(c, http.StatusBadRequest, "Link unavailable", "This sign-in link is invalid, already used, or expired. Return to Telegram and send /login for a new one.")
		return
	}

	ctx := c.Request.Context()
	u, err := resolveOrCreateTelegramUser(ctx, t.Email, t.Username)
	if err != nil {
		renderTelegramAuthPage(c, http.StatusForbidden, "Link unavailable", "We couldn't find or create an account for that email. Ask your DM for an invite, or link from the web app.")
		return
	}

	if err := telegram.UpsertIdentity(u.ID, t.TgUserID, t.ChatID, t.Username); err != nil {
		middleware.LogWarn("telegram", "auth link upsert identity failed", "error", err)
		renderTelegramAuthPage(c, http.StatusInternalServerError, "Something went wrong", "Your account was found but the link could not be saved. Try again.")
		return
	}

	sessionID := middleware.Store.Create(u.ID, u.Username, u.Role, c.ClientIP())
	c.SetCookie("session", sessionID, 86400, "/", "", false, true)

	display := u.Username
	if display == "" {
		display = t.Username
	}
	if t.ChatID != 0 {
		if _, err := telegram.SendMessage(t.ChatID, "✅ Linked as "+html.EscapeString(display)+". Send /help to see what you can do."); err != nil {
			middleware.LogWarn("telegram", "auth link confirmation DM failed", "error", err)
		}
	}

	renderTelegramAuthPage(c, http.StatusOK, "You're signed in", "Your Telegram account is linked and you're signed in to Villum. You can close this tab and return to Telegram.")
}

// resolveOrCreateTelegramUser finds an account by case-insensitive, non-empty
// email, or creates a player when self-registration is enabled.
func resolveOrCreateTelegramUser(ctx context.Context, email, tgUsername string) (*ent.User, error) {
	if u, ok := findUserByEmail(ctx, email); ok {
		return u, nil
	}
	if !telegramSelfRegisterEnabled() {
		return nil, fmt.Errorf("self-registration disabled")
	}
	return createTelegramPlayer(ctx, email, tgUsername)
}

func findUserByEmail(ctx context.Context, email string) (*ent.User, bool) {
	var id int64
	err := db.DB.QueryRow("SELECT id FROM users WHERE email <> '' AND lower(email)=lower(?) ORDER BY id LIMIT 1", email).Scan(&id)
	if err != nil {
		return nil, false
	}
	u, err := db.Client.User.Get(ctx, id)
	if err != nil {
		return nil, false
	}
	return u, true
}

func createTelegramPlayer(ctx context.Context, email, tgUsername string) (*ent.User, error) {
	username := uniqueUsername(ctx, emailLocalPart(email))
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	// The random password is discarded, so password login can never match.
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(raw)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	display := strings.TrimSpace(tgUsername)
	if display == "" {
		display = emailLocalPart(email)
	}
	return db.Client.User.Create().
		SetUsername(username).
		SetPassword(string(hash)).
		SetDisplayName(display).
		SetRole("player").
		SetEmail(email).
		Save(ctx)
}

func emailLocalPart(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return "player"
	}
	return email[:at]
}

func sanitizeUsername(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			b.WriteRune(r)
		}
	}
	out := strings.ToLower(strings.Trim(b.String(), "._-"))
	if out == "" {
		out = "player"
	}
	if len(out) > 24 {
		out = out[:24]
	}
	return out
}

func uniqueUsername(ctx context.Context, base string) string {
	base = sanitizeUsername(base)
	name := base
	for i := 2; ; i++ {
		exists, err := db.Client.User.Query().Where(user.Username(name)).Exist(ctx)
		if err != nil || !exists {
			return name
		}
		name = fmt.Sprintf("%s%d", base, i)
	}
}

func renderTelegramAuthPage(c *gin.Context, status int, title, message string) {
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("Content-Type", "text/html; charset=utf-8")
	page := "<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\">" +
		"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">" +
		"<meta name=\"referrer\" content=\"no-referrer\">" +
		"<title>" + html.EscapeString(title) + " · Villum</title>" +
		"<style>body{font-family:system-ui,sans-serif;background:#1b1b1f;color:#eee;display:flex;min-height:100vh;align-items:center;justify-content:center;margin:0}" +
		".card{max-width:26rem;padding:2rem;background:#26262c;border-radius:12px;text-align:center}" +
		"h1{font-size:1.3rem;margin:0 0 .75rem}p{margin:0;color:#bbb;line-height:1.5}</style></head>" +
		"<body><div class=\"card\"><h1>" + html.EscapeString(title) + "</h1><p>" + html.EscapeString(message) + "</p></div></body></html>"
	c.String(status, page)
}
