package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"villum/db"
	"villum/handlers/testutil"

	"github.com/gin-gonic/gin"
)

func setupTelegramDB(t *testing.T) {
	t.Helper()
	testutil.NewDB(t)
	// ensure env cleanup
	t.Cleanup(func() {
		ResetBotUsernameCache()
		os.Unsetenv("TELEGRAM_API_BASE")
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_BOT_USERNAME")
	})
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
}

// mockTelegramServer returns httptest.Server that handles Bot API methods.
func mockTelegramServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(srv.Close)
	os.Setenv("TELEGRAM_API_BASE", srv.URL)
	return srv
}

func jsonOK(v any) []byte {
	b, _ := json.Marshal(map[string]any{"ok": true, "result": v})
	return b
}
func jsonErr(code int, desc string) []byte {
	b, _ := json.Marshal(map[string]any{"ok": false, "error_code": code, "description": desc})
	return b
}

// --- chunkMessage tests ---

func TestChunkMessagePassthrough(t *testing.T) {
	txt := "hello world"
	chunks := chunkMessage(txt, 4096)
	if len(chunks) != 1 || chunks[0] != txt {
		t.Fatalf("expected passthrough, got %v", chunks)
	}
}

func TestChunkMessageSplitParagraph(t *testing.T) {
	para := strings.Repeat("a", 3000)
	txt := para + "\n\n" + para
	chunks := chunkMessage(txt, 4096)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	for _, c := range chunks {
		if len([]rune(c)) > 4096 {
			t.Fatalf("chunk too large: %d", len([]rune(c)))
		}
	}
}

func TestChunkMessageSplitLineBoundary(t *testing.T) {
	line := strings.Repeat("b", 100)
	var b strings.Builder
	for i := 0; i < 50; i++ {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(line)
	}
	txt := b.String()
	chunks := chunkMessage(txt, 4096)
	if len(chunks) < 2 {
		t.Fatalf("expected split, got %d chunks", len(chunks))
	}
	for _, c := range chunks {
		if len([]rune(c)) > 4096 {
			t.Fatalf("chunk too large")
		}
	}
}

func TestChunkMessageUTF8NotSplit(t *testing.T) {
	// rune "😀" is 4 bytes, ensure not split mid-rune
	token := strings.Repeat("😀", 5000)
	chunks := chunkMessage(token, 4096)
	for _, c := range chunks {
		if !strings.Contains(c, "😀") && c != "" {
			// ok
		}
		// verify valid UTF-8 by checking rune count matches len after range
		for _, r := range c {
			if r == 0xFFFD {
				t.Fatalf("invalid UTF-8 rune found")
			}
		}
		if len([]rune(c)) > 4096 {
			t.Fatalf("chunk exceeds limit")
		}
	}
	// reconstruct equals original
	joined := strings.Join(chunks, "")
	// for hard-cut case join should reconstruct? For line-split case, newlines preserved? For pure hard-cut, join == original
	// Here we hard-cut a single long line without separators, so chunks concatenated == original
	if joined != token {
		t.Fatalf("reconstructed mismatch: got %d runes want %d", len([]rune(joined)), len([]rune(token)))
	}
}

func TestChunkMessageDocThreshold(t *testing.T) {
	txt := strings.Repeat("x", 4096*4+10)
	chunks := chunkMessage(txt, 4096)
	if len(chunks) <= 3 {
		t.Fatalf("expected >3 chunks, got %d", len(chunks))
	}
}

// --- Link codes ---

func TestLinkCodeConsume(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	code, _, err := CreateLinkCode(1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	h := hashCode(strings.ToUpper(code))
	uid, ok := ConsumeLinkCode(h)
	if !ok || uid != 1 {
		t.Fatalf("consume failed: %v %d", ok, uid)
	}
	// second consume must fail
	_, ok2 := ConsumeLinkCode(h)
	if ok2 {
		t.Fatalf("second consume should fail")
	}
	// unknown code fails
	_, ok3 := ConsumeLinkCode("nonexistenthash")
	if ok3 {
		t.Fatalf("unknown should fail")
	}
}

func TestLinkCodeExpired(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	// expired code
	h := hashCode("EXPIRED1")
	past := time.Now().Add(-1 * time.Hour).UTC().Format("2006-01-02 15:04:05")
	if err := InsertLinkCode(1, h, past); err != nil {
		t.Fatalf("insert: %v", err)
	}
	_, ok := ConsumeLinkCode(h)
	if ok {
		t.Fatalf("expired should fail")
	}
}

// --- Webhook secret ---

func TestWebhookSecret(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	_ = SetWebhookSecret("mysecret")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/telegram/webhook", WebhookHandler)

	// missing secret -> 403
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/telegram/webhook", strings.NewReader(`{"update_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("expected 403 got %d", w.Code)
	}
	// wrong secret -> 403
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/telegram/webhook", strings.NewReader(`{"update_id":2}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong")
	r.ServeHTTP(w2, req2)
	if w2.Code != 403 {
		t.Fatalf("expected 403 wrong secret got %d", w2.Code)
	}
	// correct secret -> 200 (need to mock send? no message -> just returns ok)
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/telegram/webhook", strings.NewReader(`{"update_id":3}`))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-Telegram-Bot-Api-Secret-Token", "mysecret")
	r.ServeHTTP(w3, req3)
	if w3.Code != 200 {
		t.Fatalf("expected 200 got %d %s", w3.Code, w3.Body.String())
	}
}

// --- Dispatcher ---

func TestDispatcherUnlinkedHelps(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if txt, ok := body["text"].(string); ok {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	HandleUpdate(nil, Update{Message: &Message{Chat: &Chat{ID: 100}, From: &User{ID: 999}, Text: "/recap"}})
	if len(sent) == 0 || !strings.Contains(sent[0], "/help") && !strings.Contains(sent[0], "Villum bot") {
		t.Fatalf("expected help text, got %v", sent)
	}
}

func TestDispatcherRecapNotFoundNoLeak(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "other", "user")
	testutil.SeedCampaign(t, 10, "Secret", "Party", 2)
	testutil.SeedCampaign(t, 11, "Other", "Party2", 1)
	// add recap to campaign 10
	if _, err := db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content) VALUES(1,10,'Secret Recap','hidden')"); err != nil {
		t.Fatalf("seed recap: %v", err)
	}
	// link user 1 to telegram 111
	if err := UpsertIdentity(1, 111, 111, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if txt, ok := body["text"].(string); ok {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	// user 1 is NOT member of campaign 10, tries /recap 10
	HandleUpdate(nil, Update{Message: &Message{Chat: &Chat{ID: 111}, From: &User{ID: 111}, Text: "/recap 10"}})
	if len(sent) == 0 {
		t.Fatalf("no reply")
	}
	if sent[0] != "Recap not found." {
		t.Fatalf("expected 'Recap not found.' got %q", sent[0])
	}
	// user 2 is member, should get recap
	sent = nil
	if err := UpsertIdentity(2, 222, 222, "other"); err != nil {
		t.Fatalf("upsert2: %v", err)
	}
	HandleUpdate(nil, Update{Message: &Message{Chat: &Chat{ID: 222}, From: &User{ID: 222}, Text: "/recap 10"}})
	if len(sent) == 0 || !strings.Contains(sent[0], "Secret Recap") {
		t.Fatalf("member should get recap, got %v", sent)
	}
}

// --- Delivery idempotency ---

func TestDeliverRecapIdempotency(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "C1", "P1", 1)
	// campaign chat enabled
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled) VALUES(10,777,1)"); err != nil {
		t.Fatalf("settings: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content) VALUES(100,10,'T','hello')"); err != nil {
		t.Fatalf("recap: %v", err)
	}
	count := 0
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "sendMessage") || strings.Contains(r.URL.Path, "sendDocument") {
			count++
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	DeliverRecap(100, "manual")
	time.Sleep(400 * time.Millisecond)
	DeliverRecap(100, "manual")
	time.Sleep(400 * time.Millisecond)
	// should have exactly one sendMessage
	if count != 1 {
		t.Fatalf("expected 1 send, got %d", count)
	}
	var n int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_deliveries WHERE recap_id=100").Scan(&n)
	if n != 1 {
		t.Fatalf("expected 1 delivery row, got %d", n)
	}
}

// --- Auto-post selection ---

func TestAutoPostSelection(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "C1", "P1", 1)
	// enabled campaign
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled,auto_post_enabled) VALUES(10,777,1,1)"); err != nil {
		t.Fatalf("settings: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled,auto_post_enabled) VALUES(20,888,0,1)"); err != nil {
		// 20 not exist campaign but setting still; we'll test differently: create campaign 20 disabled
		testutil.SeedCampaign(t, 20, "C2", "P2", 1)
		db.DB.Exec("UPDATE campaign_telegram_settings SET is_enabled=0 WHERE campaign_id=20")
	}
	// also test auto_post disabled case: campaign 30
	testutil.SeedCampaign(t, 30, "C3", "P3", 1)
	db.DB.Exec("INSERT OR REPLACE INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled,auto_post_enabled) VALUES(30,999,1,0)")

	now := time.Now().UTC()
	past := now.Add(-2 * time.Hour).Format("2006-01-02 15:04:05")
	withinGrace := now.Add(-10 * time.Minute).Format("2006-01-02 15:04:05")

	// null session_end_date never selected
	db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content,session_end_date) VALUES(1,10,'R1','c',NULL)")
	// empty string also
	db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content,session_end_date) VALUES(2,10,'R2','c','')")
	// disabled campaign (is_enabled 0) -> need recap there but should not be selected; create campaign 20 recap
	db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content,session_end_date) VALUES(3,20,'R3','c',?)", past)
	// auto_post disabled -> recap 4
	db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content,session_end_date) VALUES(4,30,'R4','c',?)", past)
	// enabled + past grace -> should be delivered (recap 5)
	db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content,session_end_date) VALUES(5,10,'R5','c',?)", past)
	// within grace -> not yet (recap 6)
	db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content,session_end_date) VALUES(6,10,'R6','c',?)", withinGrace)

	// set grace to 30
	_ = SetGraceMinutes(30)

	var sends int
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "sendMessage") {
			sends++
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	scanAutoPost(now)
	time.Sleep(400 * time.Millisecond)
	// check deliveries
	var has5, has6, has1 int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_deliveries WHERE recap_id=5").Scan(&has5)
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_deliveries WHERE recap_id=6").Scan(&has6)
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_deliveries WHERE recap_id=1").Scan(&has1)
	if has5 != 1 {
		t.Fatalf("recap 5 should be delivered, got %d sends=%d", has5, sends)
	}
	if has6 != 0 {
		t.Fatalf("recap 6 within grace should not be delivered, got %d", has6)
	}
	if has1 != 0 {
		t.Fatalf("null date should not be selected")
	}
	var has3, has4 int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_deliveries WHERE recap_id=3").Scan(&has3)
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_deliveries WHERE recap_id=4").Scan(&has4)
	if has3 != 0 || has4 != 0 {
		t.Fatalf("disabled should not be selected: has3=%d has4=%d", has3, has4)
	}
	// second scan should not duplicate due to EXISTS
	sends = 0
	scanAutoPost(now)
	time.Sleep(300 * time.Millisecond)
	if sends != 0 {
		t.Fatalf("should not re-deliver, got %d sends", sends)
	}
}

// --- 429 RetryAfter ---

func TestRetryAfterError(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":5}}`))
	})
	err := doPost("sendMessage", map[string]any{"chat_id": 1, "text": "hi"}, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	ra, ok := err.(*RetryAfterError)
	if !ok {
		t.Fatalf("expected RetryAfterError, got %T %v", err, err)
	}
	if ra.After != 5*time.Second {
		t.Fatalf("expected 5s got %v", ra.After)
	}
}

// --- GetMe username resolution ---

func TestGetMeUsername(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getMe") {
			w.Write(jsonOK(map[string]any{"id": 123, "is_bot": true, "first_name": "Test", "username": "MyBot"}))
			return
		}
		w.Write(jsonOK(map[string]any{}))
	})
	ResetBotUsernameCache()
	u := GetBotUsername()
	if u != "MyBot" {
		t.Fatalf("expected MyBot got %q", u)
	}
	// env wins
	os.Setenv("TELEGRAM_BOT_USERNAME", "EnvBot")
	if got := GetBotUsername(); got != "EnvBot" {
		t.Fatalf("env should win got %q", got)
	}
}

// --- MaskToken ---

func TestMaskToken(t *testing.T) {
	m := MaskToken("1234567890ABCDEF")
	if strings.Contains(m, "4567890") {
		t.Fatalf("mask reveals too much: %q", m)
	}
	if !strings.HasPrefix(m, "123") || !strings.HasSuffix(m, "EF") {
		t.Fatalf("unexpected mask %q", m)
	}
	if MaskToken("short") != "***" {
		t.Fatalf("short should be ***")
	}
}
