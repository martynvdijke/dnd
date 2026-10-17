package telegram

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"villum/db"
	"villum/middleware"
)

// authTokenTTL bounds how long an emailed sign-in link stays valid.
const authTokenTTL = 15 * time.Minute

// authFlowTTL bounds how long the in-chat email prompt waits for a reply.
const authFlowTTL = 10 * time.Minute

const (
	authRateWindow = 15 * time.Minute
	authRateMax    = 3
)

// AuthToken is the payload stored with an emailed sign-in token.
type AuthToken struct {
	TgUserID int64
	ChatID   int64
	Username string
	Email    string
}

// HashAuthToken returns the SHA-256 hex digest used to store tokens at rest.
func HashAuthToken(token string) string { return hashCode(token) }

// GenerateAuthToken returns 32 random bytes, hex-encoded.
func GenerateAuthToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// InsertAuthToken stores a hashed sign-in token for later redemption.
func InsertAuthToken(tgUserID, chatID int64, username, email, tokenHash, expiresAt string) error {
	_, err := db.DB.Exec(
		`INSERT INTO telegram_auth_tokens(tg_user_id,chat_id,tg_username,email,token_hash,expires_at) VALUES(?,?,?,?,?,?)`,
		tgUserID, chatID, username, email, tokenHash, expiresAt)
	return err
}

// ConsumeAuthToken atomically marks a fresh token used and returns its payload.
// Expired or already-used tokens are rejected.
func ConsumeAuthToken(tokenHash string) (AuthToken, bool) {
	res, err := db.DB.Exec(
		"UPDATE telegram_auth_tokens SET used_at=datetime('now') WHERE token_hash=? AND used_at IS NULL AND datetime('now') <= expires_at",
		tokenHash)
	if err != nil {
		return AuthToken{}, false
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return AuthToken{}, false
	}
	var t AuthToken
	if err := db.DB.QueryRow(
		"SELECT tg_user_id,chat_id,tg_username,email FROM telegram_auth_tokens WHERE token_hash=?",
		tokenHash).Scan(&t.TgUserID, &t.ChatID, &t.Username, &t.Email); err != nil {
		return AuthToken{}, false
	}
	return t, true
}

// --- In-chat email flow ---------------------------------------------------

type authFlow struct {
	UpdatedAt time.Time
}

var (
	authFlowsMu sync.Mutex
	authFlows   = map[int64]*authFlow{}
)

func getAuthFlow(chatID int64) (*authFlow, bool) {
	authFlowsMu.Lock()
	defer authFlowsMu.Unlock()
	f, ok := authFlows[chatID]
	if !ok {
		return nil, false
	}
	if time.Since(f.UpdatedAt) > authFlowTTL {
		delete(authFlows, chatID)
		return nil, false
	}
	return f, true
}

func putAuthFlow(chatID int64) {
	authFlowsMu.Lock()
	defer authFlowsMu.Unlock()
	authFlows[chatID] = &authFlow{UpdatedAt: time.Now()}
}

func authFlowActive(chatID int64) bool {
	_, ok := getAuthFlow(chatID)
	return ok
}

func abortAuthFlow(chatID int64) bool {
	authFlowsMu.Lock()
	defer authFlowsMu.Unlock()
	if _, ok := authFlows[chatID]; !ok {
		return false
	}
	delete(authFlows, chatID)
	return true
}

// --- Rate limiting --------------------------------------------------------

var (
	authRateMu sync.Mutex
	authRate   = map[int64][]time.Time{}
)

// allowAuthLink reports whether the Telegram user may request another link,
// recording the attempt when allowed.
func allowAuthLink(tgUserID int64) bool {
	authRateMu.Lock()
	defer authRateMu.Unlock()
	cutoff := time.Now().Add(-authRateWindow)
	hits := authRate[tgUserID][:0]
	for _, t := range authRate[tgUserID] {
		if t.After(cutoff) {
			hits = append(hits, t)
		}
	}
	if len(hits) >= authRateMax {
		authRate[tgUserID] = hits
		return false
	}
	authRate[tgUserID] = append(hits, time.Now())
	return true
}

// --- Injected email sender ------------------------------------------------

// AuthLinkSender emails a sign-in link for the given address and raw token.
type AuthLinkSender func(email, token string) error

var (
	authLinkSenderMu sync.RWMutex
	authLinkSender   AuthLinkSender
)

// SetAuthLinkSender wires the outbound email path (see app.go).
func SetAuthLinkSender(fn AuthLinkSender) {
	authLinkSenderMu.Lock()
	authLinkSender = fn
	authLinkSenderMu.Unlock()
}

func getAuthLinkSender() AuthLinkSender {
	authLinkSenderMu.RLock()
	defer authLinkSenderMu.RUnlock()
	return authLinkSender
}

// --- Validation and command ----------------------------------------------

// validEmailAddress accepts a single-@ address with non-empty local and domain
// parts and no whitespace.
func validEmailAddress(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at != strings.LastIndexByte(s, '@') {
		return false
	}
	local, domain := s[:at], s[at+1:]
	if local == "" || domain == "" || strings.HasPrefix(domain, "@") {
		return false
	}
	return !strings.ContainsAny(s, " \t\r\n")
}

// authEmailAvailable reports whether the email path is configured.
func authEmailAvailable() bool {
	return strings.TrimSpace(os.Getenv("BASE_URL")) != "" && getAuthLinkSender() != nil
}

// linkedVillumUsername returns the Villum username linked to a Telegram user.
func linkedVillumUsername(tgID int64) (string, bool) {
	var name string
	err := db.DB.QueryRow(
		"SELECT u.username FROM users u JOIN telegram_identities ti ON ti.user_id=u.id WHERE ti.telegram_user_id=?",
		tgID).Scan(&name)
	return name, err == nil
}

// runLogin starts the in-chat email sign-in flow.
func runLogin(c *cmdContext) botReply {
	if _, _, _, ok := LookupIdentityByTelegramID(c.tgUserID); ok {
		name, _ := linkedVillumUsername(c.tgUserID)
		if name == "" {
			name = "your Villum account"
		} else {
			name = "<b>" + escapeHTML(name) + "</b>"
		}
		return botReply{Text: fmt.Sprintf("✅ You're already linked as %s. Send /unlink first if you want to switch accounts.", name)}
	}
	if !authEmailAvailable() {
		return botReply{Text: "📭 Email sign-in is unavailable right now.\n\nLink with the web code flow: open Villum → <b>Settings → Telegram</b>, generate a code, then send <code>/start &lt;code&gt;</code> here."}
	}
	putAuthFlow(c.chatID)
	return botReply{Text: "📧 <b>Email sign-in</b>\n\nSend the email address for your Villum account and I'll email you a one-time link to sign in and link this chat. Send /cancel to stop."}
}

// feedAuthFlow handles the reply to the email prompt.
func feedAuthFlow(c *cmdContext, text string) {
	email := strings.TrimSpace(text)
	if !validEmailAddress(email) {
		putAuthFlow(c.chatID)
		c.reply("That doesn't look like an email address. Send a valid address, or /cancel to stop.")
		return
	}
	abortAuthFlow(c.chatID)

	if !allowAuthLink(c.tgUserID) {
		c.reply("⏳ Too many sign-in requests. Wait a few minutes, then try again.")
		return
	}

	// Every well-formed address gets the same acknowledgement, whether or not a
	// matching account exists and whether or not delivery succeeds.
	ack := "📧 Check your inbox."

	token, err := GenerateAuthToken()
	if err != nil {
		middleware.LogError("telegram", "generate auth token", "err", err)
		c.reply(ack)
		return
	}
	expires := time.Now().UTC().Add(authTokenTTL).Format("2006-01-02 15:04:05")
	if err := InsertAuthToken(c.tgUserID, c.chatID, c.username, email, HashAuthToken(token), expires); err != nil {
		middleware.LogError("telegram", "store auth token", "err", err)
		c.reply(ack)
		return
	}
	if sender := getAuthLinkSender(); sender != nil {
		if err := sender(email, token); err != nil {
			middleware.LogError("telegram", "send auth link", "err", err)
		}
	}
	c.reply(ack)
}
