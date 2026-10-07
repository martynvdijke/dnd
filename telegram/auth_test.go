package telegram

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"villum/db"
	"villum/handlers/testutil"
)

func resetAuthState() {
	authFlowsMu.Lock()
	authFlows = map[int64]*authFlow{}
	authFlowsMu.Unlock()
	authRateMu.Lock()
	authRate = map[int64][]time.Time{}
	authRateMu.Unlock()
	SetAuthLinkSender(nil)
}

func TestValidEmailAddress(t *testing.T) {
	valid := []string{"a@b.com", "Alice@Example.com", "a.b+c@sub.domain.io"}
	for _, s := range valid {
		if !validEmailAddress(s) {
			t.Errorf("expected %q valid", s)
		}
	}
	invalid := []string{"", "no-at", "@domain.com", "local@", "a@@b.com", "a b@c.com", "a@b c.com"}
	for _, s := range invalid {
		if validEmailAddress(s) {
			t.Errorf("expected %q invalid", s)
		}
	}
}

func TestAuthFlowStoreAndExpiry(t *testing.T) {
	resetAuthState()
	putAuthFlow(1)
	if !authFlowActive(1) {
		t.Fatal("flow should be active")
	}
	authFlowsMu.Lock()
	authFlows[1].UpdatedAt = time.Now().Add(-authFlowTTL - time.Minute)
	authFlowsMu.Unlock()
	if authFlowActive(1) {
		t.Fatal("flow should have expired")
	}
	putAuthFlow(2)
	if !abortAuthFlow(2) {
		t.Fatal("abort should report the flow was present")
	}
	if abortAuthFlow(2) {
		t.Fatal("second abort should report nothing")
	}
}

func TestAuthTokenConsumeSingleUseAndExpiry(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)

	hash := HashAuthToken("token-one")
	exp := time.Now().UTC().Add(authTokenTTL).Format("2006-01-02 15:04:05")
	if err := InsertAuthToken(10, 20, "tguser", "a@example.com", hash, exp); err != nil {
		t.Fatalf("insert: %v", err)
	}
	at, ok := ConsumeAuthToken(hash)
	if !ok {
		t.Fatal("first consume should succeed")
	}
	if at.TgUserID != 10 || at.ChatID != 20 || at.Username != "tguser" || at.Email != "a@example.com" {
		t.Fatalf("unexpected payload: %+v", at)
	}
	if _, ok := ConsumeAuthToken(hash); ok {
		t.Fatal("second consume should fail")
	}

	expired := HashAuthToken("token-expired")
	past := time.Now().UTC().Add(-time.Minute).Format("2006-01-02 15:04:05")
	if err := InsertAuthToken(11, 21, "u", "b@example.com", expired, past); err != nil {
		t.Fatalf("insert expired: %v", err)
	}
	if _, ok := ConsumeAuthToken(expired); ok {
		t.Fatal("expired token should fail")
	}
	if _, ok := ConsumeAuthToken(HashAuthToken("nope")); ok {
		t.Fatal("unknown token should fail")
	}
}

func TestAllowAuthLinkRateLimit(t *testing.T) {
	resetAuthState()
	for i := 0; i < authRateMax; i++ {
		if !allowAuthLink(5) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if allowAuthLink(5) {
		t.Fatal("request beyond the limit should be throttled")
	}
	if !allowAuthLink(6) {
		t.Fatal("another Telegram user must be unaffected")
	}
}

func TestRunLoginUnavailable(t *testing.T) {
	resetAuthState()
	os.Unsetenv("BASE_URL")
	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	reply := runLogin(c)
	if !strings.Contains(reply.Text, "unavailable") {
		t.Fatalf("expected unavailability message, got %q", reply.Text)
	}
	if authFlowActive(100) {
		t.Fatal("no flow should start when unavailable")
	}
}

func TestRunLoginAlreadyLinked(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	resetAuthState()
	os.Setenv("BASE_URL", "https://example.test")
	t.Cleanup(func() { os.Unsetenv("BASE_URL") })
	testutil.SeedUser(t, 1, "alice", "user")
	if err := UpsertIdentity(1, 100, 100, "alice_tg"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	reply := runLogin(c)
	if !strings.Contains(reply.Text, "already linked") || !strings.Contains(reply.Text, "alice") {
		t.Fatalf("expected already-linked reply naming the user, got %q", reply.Text)
	}
	if authFlowActive(100) {
		t.Fatal("no flow should start for a linked user")
	}
}

func TestAuthFlowSendsLinkAndValidates(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	resetAuthState()
	os.Setenv("BASE_URL", "https://example.test")
	t.Cleanup(func() { os.Unsetenv("BASE_URL") })

	var gotEmail, gotToken string
	SetAuthLinkSender(func(email, token string) error {
		gotEmail, gotToken = email, token
		return nil
	})
	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100, username: "tester"}

	if reply := runLogin(c); !strings.Contains(reply.Text, "email") {
		t.Fatalf("expected email prompt, got %q", reply.Text)
	}
	if !authFlowActive(100) {
		t.Fatal("flow should be active")
	}

	feedAuthFlow(c, "not-an-email")
	if gotToken != "" || gotEmail != "" {
		t.Fatal("malformed address must not send mail")
	}
	if !authFlowActive(100) {
		t.Fatal("flow stays active after malformed input")
	}

	feedAuthFlow(c, "Alice@Example.com")
	if gotEmail != "Alice@Example.com" || gotToken == "" {
		t.Fatalf("expected captured email and token, got %q %q", gotEmail, gotToken)
	}
	var n int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM telegram_auth_tokens WHERE token_hash=?", HashAuthToken(gotToken)).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected hashed token row, got %d", n)
	}
	if authFlowActive(100) {
		t.Fatal("flow should end after a valid address")
	}
	if !strings.Contains(strings.Join(sent, "\n"), "Check your inbox") {
		t.Fatalf("expected acknowledgement, got %v", sent)
	}
}

func TestAuthAckEnumerationSafe(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	resetAuthState()
	os.Setenv("BASE_URL", "https://example.test")
	t.Cleanup(func() { os.Unsetenv("BASE_URL") })
	SetAuthLinkSender(func(email, token string) error { return nil })

	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	putAuthFlow(100)
	feedAuthFlow(c, "known@example.com")
	putAuthFlow(100)
	feedAuthFlow(c, "unknown@example.com")

	if len(sent) < 2 {
		t.Fatalf("expected two replies, got %v", sent)
	}
	last, prev := sent[len(sent)-1], sent[len(sent)-2]
	if last != prev || !strings.Contains(last, "Check your inbox") {
		t.Fatalf("replies must be identical acknowledgements, got %q and %q", prev, last)
	}
}

func TestFeedAuthFlowRateLimited(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	resetAuthState()
	os.Setenv("BASE_URL", "https://example.test")
	t.Cleanup(func() { os.Unsetenv("BASE_URL") })
	sends := 0
	SetAuthLinkSender(func(email, token string) error { sends++; return nil })

	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	for i := 0; i < authRateMax; i++ {
		putAuthFlow(100)
		feedAuthFlow(c, "a@example.com")
	}
	putAuthFlow(100)
	feedAuthFlow(c, "a@example.com")
	if sends != authRateMax {
		t.Fatalf("expected %d sends, got %d", authRateMax, sends)
	}
	if !strings.Contains(sent[len(sent)-1], "Too many") {
		t.Fatalf("expected throttling reply, got %v", sent)
	}
}

func TestDispatcherAuthFlow(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	resetAuthState()
	os.Setenv("BASE_URL", "https://example.test")
	t.Cleanup(func() { os.Unsetenv("BASE_URL") })
	var gotEmail string
	SetAuthLinkSender(func(email, token string) error { gotEmail = email; return nil })

	HandleUpdate(context.Background(), messageUpdate(1, 100, 100, "/login"))
	if !authFlowActive(100) {
		t.Fatalf("login should start the flow, got %v", sent)
	}
	HandleUpdate(context.Background(), messageUpdate(2, 100, 100, "me@example.com"))
	if gotEmail != "me@example.com" {
		t.Fatalf("flow reply should be treated as the address, got %q", gotEmail)
	}
	if authFlowActive(100) {
		t.Fatal("flow should end after a valid address")
	}

	HandleUpdate(context.Background(), messageUpdate(3, 100, 100, "/login"))
	if !authFlowActive(100) {
		t.Fatal("login should restart the flow")
	}
	HandleUpdate(context.Background(), messageUpdate(4, 100, 100, "/help"))
	if authFlowActive(100) {
		t.Fatal("another command should abort the auth flow")
	}
}
