package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent/user"
	"villum/handlers/testutil"
	"villum/middleware"
	"villum/models"
	"villum/telegram"
)

// setupAuthLinkTest prepares a full DB, a local Telegram API stub, a real
// session store, and a router exposing the public redemption endpoint.
func setupAuthLinkTest(t *testing.T) *gin.Engine {
	t.Helper()
	testutil.NewDB(t)
	t.Cleanup(func() { testutil.CloseDB(t) })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test","username":"testbot","message_id":1,"date":0,"chat":{"id":1}}}`)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("TELEGRAM_API_BASE", srv.URL)
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")

	oldStore := middleware.Store
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	t.Cleanup(func() { middleware.Store = oldStore })

	oldBase := BaseURL
	BaseURL = "https://example.test"
	t.Cleanup(func() { BaseURL = oldBase })

	r := gin.New()
	r.GET("/telegram/auth", HandleTelegramAuthLink)
	return r
}

func issueAuthToken(t *testing.T, tgID, chatID int64, username, email string, expires time.Time) string {
	t.Helper()
	raw, err := telegram.GenerateAuthToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if err := telegram.InsertAuthToken(tgID, chatID, username, email, telegram.HashAuthToken(raw), expires.UTC().Format("2006-01-02 15:04:05")); err != nil {
		t.Fatalf("insert token: %v", err)
	}
	return raw
}

func referrerPolicy(w *httptest.ResponseRecorder) string {
	return w.Result().Header.Get("Referrer-Policy")
}

func sessionCookie(w *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			return c
		}
	}
	return nil
}

func TestTelegramAuthLinkCreatesPlayerAndSession(t *testing.T) {
	r := setupAuthLinkTest(t)
	raw := issueAuthToken(t, 555, 555, "newbie", "newplayer@example.com", time.Now().Add(15*time.Minute))

	w := testutil.Get(t, r, "/telegram/auth?token="+raw)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", w.Code, w.Body.String())
	}
	if got := referrerPolicy(w); got != "no-referrer" {
		t.Fatalf("expected no-referrer, got %q", got)
	}
	if cookie := sessionCookie(w); cookie == nil || cookie.Value == "" {
		t.Fatal("expected a session cookie to be set")
	} else if middleware.Store.Get(cookie.Value) == nil {
		t.Fatal("session cookie should resolve in the store")
	}

	ctx := t.Context()
	u, err := db.Client.User.Query().Where(user.Username("newplayer")).Only(ctx)
	if err != nil {
		t.Fatalf("expected player account: %v", err)
	}
	if u.Role != "player" || u.Email != "newplayer@example.com" {
		t.Fatalf("unexpected player: role=%q email=%q", u.Role, u.Email)
	}

	var linked int64
	if err := db.DB.QueryRow("SELECT user_id FROM telegram_identities WHERE telegram_user_id=555").Scan(&linked); err != nil {
		t.Fatalf("identity not upserted: %v", err)
	}
	if linked != u.ID {
		t.Fatalf("identity linked to %d, expected %d", linked, u.ID)
	}
}

func TestTelegramAuthLinkExistingEmailDoesNotCreate(t *testing.T) {
	r := setupAuthLinkTest(t)
	testutil.SeedUser(t, 2, "bob", "player")
	if _, err := db.DB.Exec("UPDATE users SET email=? WHERE id=2", "bob@example.com"); err != nil {
		t.Fatalf("set email: %v", err)
	}
	before := testutil.CountRows(t, "users")

	raw := issueAuthToken(t, 556, 556, "bob_tg", "BOB@example.com", time.Now().Add(15*time.Minute))
	w := testutil.Get(t, r, "/telegram/auth?token="+raw)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", w.Code, w.Body.String())
	}
	if after := testutil.CountRows(t, "users"); after != before {
		t.Fatalf("no account should be created: before=%d after=%d", before, after)
	}
	var linked int64
	if err := db.DB.QueryRow("SELECT user_id FROM telegram_identities WHERE telegram_user_id=556").Scan(&linked); err != nil {
		t.Fatalf("identity not upserted: %v", err)
	}
	if linked != 2 {
		t.Fatalf("expected link to existing user 2, got %d", linked)
	}
}

func TestTelegramAuthLinkRejectsUsedAndExpired(t *testing.T) {
	r := setupAuthLinkTest(t)

	used := issueAuthToken(t, 557, 557, "u", "used@example.com", time.Now().Add(15*time.Minute))
	if w := testutil.Get(t, r, "/telegram/auth?token="+used); w.Code != http.StatusOK {
		t.Fatalf("first use should succeed, got %d", w.Code)
	}
	if w := testutil.Get(t, r, "/telegram/auth?token="+used); w.Code == http.StatusOK {
		t.Fatalf("reused token should be refused, got %d", w.Code)
	}

	expired := issueAuthToken(t, 558, 558, "u", "expired@example.com", time.Now().Add(-time.Minute))
	if w := testutil.Get(t, r, "/telegram/auth?token="+expired); w.Code == http.StatusOK {
		t.Fatalf("expired token should be refused, got %d", w.Code)
	}

	if w := testutil.Get(t, r, "/telegram/auth?token=nonsense"); w.Code == http.StatusOK {
		t.Fatalf("unknown token should be refused, got %d", w.Code)
	}
}

func TestTelegramAuthLinkSelfRegisterDisabled(t *testing.T) {
	r := setupAuthLinkTest(t)
	t.Setenv("TELEGRAM_SELF_REGISTER", "off")
	before := testutil.CountRows(t, "users")

	raw := issueAuthToken(t, 559, 559, "u", "nobody@example.com", time.Now().Add(15*time.Minute))
	w := testutil.Get(t, r, "/telegram/auth?token="+raw)
	if w.Code == http.StatusOK {
		t.Fatalf("self-registration off should refuse, got %d", w.Code)
	}
	if after := testutil.CountRows(t, "users"); after != before {
		t.Fatalf("no account should be created: before=%d after=%d", before, after)
	}
	var n int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_identities WHERE telegram_user_id=559").Scan(&n)
	if n != 0 {
		t.Fatalf("no identity should be linked when refused, got %d", n)
	}
}

func TestTelegramAuthLinkMissingToken(t *testing.T) {
	r := setupAuthLinkTest(t)
	w := testutil.Get(t, r, "/telegram/auth")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing token, got %d", w.Code)
	}
	if got := referrerPolicy(w); got != "no-referrer" {
		t.Fatalf("expected no-referrer on refusal, got %q", got)
	}
	if !strings.Contains(w.Body.String(), "invalid") {
		t.Fatalf("expected an explanatory page, got %q", w.Body.String())
	}
}

func TestSendTelegramAuthLinkEmailBuildsURL(t *testing.T) {
	testutil.NewDB(t)
	t.Cleanup(func() { testutil.CloseDB(t) })

	oldBase := BaseURL
	BaseURL = "https://example.test/"
	t.Cleanup(func() { BaseURL = oldBase })

	if _, err := db.DB.Exec(`INSERT OR REPLACE INTO email_settings (id, smtp_host, smtp_port, username, password, from_addr, enabled) VALUES (1,'smtp.example.com',587,'u','p','from@example.com',1)`); err != nil {
		t.Fatalf("seed email settings: %v", err)
	}

	orig := SendEmailFunc
	var to, body string
	SendEmailFunc = func(s *models.EmailSettings, recipient, subject, b string) error {
		to, body = recipient, b
		return nil
	}
	t.Cleanup(func() { SendEmailFunc = orig })

	if err := SendTelegramAuthLinkEmail("dest@example.com", "tok+en"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if to != "dest@example.com" {
		t.Fatalf("unexpected recipient %q", to)
	}
	if !strings.Contains(body, "https://example.test/telegram/auth?token=tok%2Ben") {
		t.Fatalf("expected escaped link in body, got %q", body)
	}
}
