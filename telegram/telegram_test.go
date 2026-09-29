package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/handlers/testutil"
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

// messageUpdate builds a private-chat text update for dispatcher tests.
func messageUpdate(updateID, chatID, userID int64, text string) *tgmodels.Update {
	return &tgmodels.Update{
		ID: updateID,
		Message: &tgmodels.Message{
			ID:   int(updateID),
			Chat: tgmodels.Chat{ID: chatID, Type: "private"},
			From: &tgmodels.User{ID: userID, FirstName: "Tester", Username: "tester"},
			Text: text,
		},
	}
}

// recordingMock records sent message texts and answers every Bot API call.
func recordingMock(t *testing.T, sent *[]string) *httptest.Server {
	t.Helper()
	return mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if txt := requestFormValue(r, "text"); txt != "" {
			*sent = append(*sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
}

// requestFormValue reads one parameter from a Bot API request. The library
// always sends multipart/form-data, but JSON is supported for hand-built
// requests.
func requestFormValue(r *http.Request, key string) string {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return ""
		}
		return r.FormValue(key)
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return ""
	}
	if v, ok := body[key].(string); ok {
		return v
	}
	return ""
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
	// hard-cut a single long line without separators, so chunks concatenated == original
	joined := strings.Join(chunks, "")
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

func TestEscapeHTML(t *testing.T) {
	got := escapeHTML(`<b>A & B</b>`)
	want := "&lt;b&gt;A &amp; B&lt;/b&gt;"
	if got != want {
		t.Fatalf("escapeHTML = %q, want %q", got, want)
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
	// correct secret -> 200; with no running client the update is dropped
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
	recordingMock(t, &sent)
	HandleUpdate(context.Background(), messageUpdate(1, 100, 999, "/recap"))
	if len(sent) == 0 || !strings.Contains(sent[0], "/start") {
		t.Fatalf("expected linking instructions, got %v", sent)
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
	recordingMock(t, &sent)
	// user 1 is NOT member of campaign 10, tries /recap 10
	HandleUpdate(context.Background(), messageUpdate(2, 111, 111, "/recap 10"))
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
	HandleUpdate(context.Background(), messageUpdate(3, 222, 222, "/recap 10"))
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
	testutil.SeedCampaign(t, 20, "C2", "P2", 1)
	db.DB.Exec("INSERT OR REPLACE INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled,auto_post_enabled) VALUES(20,888,0,1)")
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

// --- 429 RetryAfter translation ---

func TestTranslateBotError(t *testing.T) {
	err := translateBotError(&tgbot.TooManyRequestsError{RetryAfter: 5})
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
	sentinel := errors.New("boom")
	if got := translateBotError(sentinel); got != sentinel {
		t.Fatalf("expected passthrough, got %v", got)
	}
	if translateBotError(nil) != nil {
		t.Fatalf("nil should stay nil")
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

// --- Registry, help and native menu ---

func TestHelpTextFromRegistry(t *testing.T) {
	help := helpText()
	for _, want := range []string{"/characters", "/sheet", "/stats", "/overview", "/recap", "/status", "Characters", "Campaigns", "Notifications"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
	// every non-hidden registry command appears in help
	for _, cmd := range commandRegistry {
		if cmd.hidden {
			continue
		}
		if !strings.Contains(help, "/"+cmd.name) {
			t.Fatalf("help missing command %q", cmd.name)
		}
	}
}

func TestFindCommand(t *testing.T) {
	if _, ok := findCommand("HELP"); !ok {
		t.Fatalf("findCommand should be case-insensitive")
	}
	if _, ok := findCommand("nope"); ok {
		t.Fatalf("unknown command should not resolve")
	}
}

func TestBotCommandList(t *testing.T) {
	list := botCommandList()
	visible := 0
	for _, cmd := range commandRegistry {
		if !cmd.hidden {
			visible++
		}
	}
	if len(list) != visible {
		t.Fatalf("expected %d commands, got %d", visible, len(list))
	}
	for _, c := range list {
		if strings.HasPrefix(c.Command, "/") {
			t.Fatalf("command must not have a leading slash: %q", c.Command)
		}
		if len([]rune(c.Description)) > 256 {
			t.Fatalf("description too long for %q", c.Command)
		}
	}
}

// --- Callback routing ---

func TestHandleCallbackDataNavigation(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if err := UpsertIdentity(1, 999, 999, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)
	c := &cmdContext{ctx: context.Background(), chatID: 999, tgUserID: 999}
	reply, ok := handleCallbackData(c, cbNavPrefix+"characters")
	if !ok {
		t.Fatalf("nav callback not handled")
	}
	if reply.Keyboard == nil {
		t.Fatalf("navigation keyboard missing")
	}
	if !strings.Contains(reply.Text, "no characters") {
		t.Fatalf("unexpected reply: %q", reply.Text)
	}
	if _, ok := handleCallbackData(c, "bogus:1"); ok {
		t.Fatalf("unknown callback should not be handled")
	}
}

// --- Claims ---

func TestClaimLifecycle(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "dm", "user")
	testutil.SeedCharacter(t, 1, 1, "Aria", "Elf", "Ranger")
	// campaign owned by user 2 contains character 1, so user 2 may edit it too
	testutil.SeedCampaign(t, 10, "Table", "Party", 2)
	if _, err := db.DB.Exec("INSERT OR IGNORE INTO campaign_characters(campaign_id,character_id) VALUES(10,1)"); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if err := UpsertIdentity(1, 100, 100, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := UpsertIdentity(2, 200, 200, "dm"); err != nil {
		t.Fatalf("upsert2: %v", err)
	}
	owner := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	reply := claimByID(owner, 1)
	if !strings.Contains(reply.Text, "now playing") {
		t.Fatalf("claim failed: %q", reply.Text)
	}
	claim, found, err := getClaim(100)
	if err != nil || !found || claim.CharacterID != 1 {
		t.Fatalf("claim not stored: found=%v err=%v", found, err)
	}
	// another user who can edit the character still cannot take the claim
	other := &cmdContext{ctx: context.Background(), chatID: 200, tgUserID: 200}
	if reply := claimByID(other, 1); !strings.Contains(reply.Text, "already claimed") {
		t.Fatalf("expected refusal, got %q", reply.Text)
	}
	// candidates exclude already claimed characters
	candidates, err := claimCandidates(1)
	if err != nil {
		t.Fatalf("candidates: %v", err)
	}
	for _, ch := range candidates {
		if ch.ID == 1 {
			t.Fatalf("claimed character must not be a candidate")
		}
	}
	// unclaim releases it
	if reply := runUnclaim(owner); !strings.Contains(reply.Text, "Released") {
		t.Fatalf("unclaim failed: %q", reply.Text)
	}
	if _, found, _ := getClaim(100); found {
		t.Fatalf("claim should be gone")
	}
	// claiming a foreign character is refused
	if reply := claimByID(owner, 999); !strings.Contains(reply.Text, "not found") {
		t.Fatalf("expected not found, got %q", reply.Text)
	}
}

// --- Create flow ---

func TestCreateFlow(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 42, 1, "Aria", "Elf", "Ranger")
	if err := UpsertIdentity(1, 100, 100, "admin"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	var gotInput CreateCharacterInput
	var gotUser int64
	SetCharacterCreator(func(ctx context.Context, userID int64, in CreateCharacterInput) (CreatedCharacter, error) {
		gotInput = in
		gotUser = userID
		return CreatedCharacter{ID: 42, Name: in.Name, Race: in.Race, Class: in.Class, Level: in.Level, HpMax: 10, HpCurrent: 10, Ac: 10}, nil
	})
	t.Cleanup(func() { SetCharacterCreator(nil) })

	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	if reply := runCreate(c); !strings.Contains(reply.Text, "name") {
		t.Fatalf("expected name prompt, got %q", reply.Text)
	}
	if !flowActive(100) {
		t.Fatalf("flow should be active")
	}
	feedCreateFlow(c, "Aria")
	feedCreateFlow(c, "Elf")
	feedCreateFlow(c, "Ranger")
	feedCreateFlow(c, "3")
	feedCreateFlow(c, "none")

	if gotUser != 1 {
		t.Fatalf("creator got user %d", gotUser)
	}
	if gotInput.Name != "Aria" || gotInput.Race != "Elf" || gotInput.Class != "Ranger" || gotInput.Level != 3 || gotInput.CampaignID != 0 {
		t.Fatalf("unexpected input: %+v", gotInput)
	}
	if !strings.Contains(strings.Join(sent, "\n"), "Created") {
		t.Fatalf("expected creation confirmation, got %v", sent)
	}
	if claim, found, _ := getClaim(100); !found || claim.CharacterID != 42 {
		t.Fatalf("created character should be claimed")
	}
	if flowActive(100) {
		t.Fatalf("flow should be finished")
	}
}

func TestCreateFlowValidationAndAbort(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	_ = UpsertIdentity(1, 100, 100, "admin")
	var sent []string
	recordingMock(t, &sent)
	SetCharacterCreator(func(ctx context.Context, userID int64, in CreateCharacterInput) (CreatedCharacter, error) {
		t.Fatalf("creator must not be called for invalid input")
		return CreatedCharacter{}, nil
	})
	t.Cleanup(func() { SetCharacterCreator(nil) })

	c := &cmdContext{ctx: context.Background(), chatID: 100, tgUserID: 100}
	start := runCreate(c)
	if !strings.Contains(start.Text, "name") {
		t.Fatalf("expected name prompt, got %q", start.Text)
	}
	feedCreateFlow(c, "Aria")
	feedCreateFlow(c, "Elf")
	feedCreateFlow(c, "Ranger")
	feedCreateFlow(c, "99")
	if !strings.Contains(strings.Join(sent, "\n"), "between 1 and 20") {
		t.Fatalf("expected level validation reply, got %v", sent)
	}
	if !flowActive(100) {
		t.Fatalf("flow should still be active after invalid level")
	}
	abortCreateFlow(100)
	if flowActive(100) {
		t.Fatalf("flow should be aborted")
	}
	if reply := runCancel(c); !strings.Contains(reply.Text, "Nothing to cancel") {
		t.Fatalf("unexpected cancel reply: %q", reply.Text)
	}
}

// --- Settings ---

func TestCampaignTelegramSettingsAccessors(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 7, "Table", "Party", 1)
	chatID := int64(555)
	s := CampaignTelegramSettings{
		CampaignID:      7,
		ChatID:          &chatID,
		ChatType:        "supergroup",
		TitleCache:      "Table",
		IsEnabled:       true,
		AutoPostEnabled: true,
	}
	if err := UpsertCampaignTelegramSettings(s); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, found, err := GetCampaignTelegramSettings(7)
	if err != nil || !found {
		t.Fatalf("get: found=%v err=%v", found, err)
	}
	if got.ChatID == nil || *got.ChatID != 555 || got.TitleCache != "Table" || !got.IsEnabled || !got.AutoPostEnabled {
		t.Fatalf("unexpected settings: %+v", got)
	}
	got.IsEnabled = false
	if err := UpsertCampaignTelegramSettings(got); err != nil {
		t.Fatalf("upsert update: %v", err)
	}
	again, _, _ := GetCampaignTelegramSettings(7)
	if again.IsEnabled {
		t.Fatalf("is_enabled update lost")
	}
	if _, found, err := GetCampaignTelegramSettings(99); err != nil || found {
		t.Fatalf("missing row should return found=false, got found=%v err=%v", found, err)
	}
}

func TestLoadSettingsEnvOverride(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	_ = SetBotToken("stored-token")
	os.Setenv("TELEGRAM_BOT_TOKEN", "env-token")
	if got := LoadSettings(); got.Token != "env-token" {
		t.Fatalf("env token should win, got %q", got.Token)
	}
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	if got := LoadSettings(); got.Token != "stored-token" {
		t.Fatalf("stored token should be used, got %q", got.Token)
	}
}

// --- Supervisor ---

func TestSupervisorReconcile(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getMe") {
			w.Write(jsonOK(map[string]any{"id": 123, "is_bot": true, "first_name": "T", "username": "tbot"}))
			return
		}
		w.Write(jsonOK(true))
	})
	bot.mu.Lock()
	bot.stopClientLocked()
	bot.started = false
	bot.mu.Unlock()
	t.Cleanup(func() {
		bot.mu.Lock()
		bot.stopClientLocked()
		bot.started = false
		bot.mu.Unlock()
	})

	_ = SetMode("webhook")
	bot.reconcile()
	if runningClient() == nil {
		t.Fatalf("expected a running client after reconcile with a token")
	}

	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	_ = SetBotToken("")
	_ = SetMode("off")
	bot.reconcile()
	if runningClient() != nil {
		t.Fatalf("expected client to stop when settings are cleared")
	}
}

func TestGroupCommandsBoundChat(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 7, "Sunken Crown", "The Party", 1)
	testutil.SeedCharacterInCampaign(t, 11, 1, 7, "Aria", "Elf", "Ranger")
	if _, err := db.DB.Exec(`INSERT INTO party_items (campaign_id, name, quantity, notes) VALUES (7, 'Rope', 3, ''), (7, '<b>Bomb</b>', 1, '')`); err != nil {
		t.Fatalf("seed party items: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO quests (character_id, name, status, objectives) VALUES (11, 'Find the Sunken Crown', 'active', 'Search the harbor')`); err != nil {
		t.Fatalf("seed quest: %v", err)
	}
	res, err := db.DB.Exec(`INSERT INTO locations (user_id, name, type) VALUES (1, 'Waterdeep', 'city')`)
	if err != nil {
		t.Fatalf("seed location: %v", err)
	}
	locID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("location id: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO character_locations (character_id, location_id, relationship) VALUES (11, ?, 'visited')`, locID); err != nil {
		t.Fatalf("seed character location: %v", err)
	}
	groupChat := int64(-100123)
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 7, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	var sent []string
	recordingMock(t, &sent)

	cases := []struct{ text, want string }{
		{"/items", "Rope"},
		{"/quests", "Find the Sunken Crown"},
		{"/visits", "Waterdeep"},
		{"/stats", "Statistics"},
	}
	for i, tc := range cases {
		HandleUpdate(context.Background(), messageUpdate(int64(900+i), groupChat, 999999, tc.text))
		if len(sent) == 0 || !strings.Contains(sent[len(sent)-1], tc.want) {
			t.Fatalf("%s reply does not contain %q: %v", tc.text, tc.want, sent)
		}
	}
	joined := strings.Join(sent, "\n")
	if strings.Contains(joined, "<b>Bomb</b>") || !strings.Contains(joined, "&lt;b&gt;Bomb&lt;/b&gt;") {
		t.Fatalf("hostile item text was not escaped: %q", joined)
	}
}

func TestGroupCommandsUnboundChat(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)
	groupChat := int64(-100777)
	HandleUpdate(context.Background(), messageUpdate(910, groupChat, 999999, "/items"))
	if len(sent) == 0 || !strings.Contains(sent[len(sent)-1], "not connected to a campaign") {
		t.Fatalf("expected binding instructions, got %v", sent)
	}
}

func TestStartWelcomeAndLink(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 5, "player", "user")
	var sent []string
	recordingMock(t, &sent)

	HandleUpdate(context.Background(), messageUpdate(920, 555, 555, "/start"))
	if len(sent) == 0 || !strings.Contains(sent[len(sent)-1], "Welcome") {
		t.Fatalf("expected welcome, got %v", sent)
	}

	code, _, err := CreateLinkCode(5)
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	HandleUpdate(context.Background(), messageUpdate(921, 555, 555, "/start "+code))
	last := sent[len(sent)-1]
	if !strings.Contains(last, "Linked") || !strings.Contains(last, "/claim") {
		t.Fatalf("expected post-link guidance, got %q", last)
	}
}

func TestGroupWelcome(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	var sent []string
	recordingMock(t, &sent)

	chat := tgmodels.Chat{ID: -100999, Type: tgmodels.ChatTypeSupergroup}
	added := &tgmodels.ChatMemberUpdated{
		Chat:          chat,
		From:          tgmodels.User{ID: 1},
		NewChatMember: tgmodels.ChatMember{Type: tgmodels.ChatMemberTypeMember},
		OldChatMember: tgmodels.ChatMember{Type: tgmodels.ChatMemberTypeLeft},
	}
	HandleUpdate(context.Background(), &tgmodels.Update{ID: 930, MyChatMember: added})
	if len(sent) == 0 || !strings.Contains(sent[len(sent)-1], "Thanks for adding me") {
		t.Fatalf("expected group welcome, got %v", sent)
	}

	before := len(sent)
	HandleUpdate(context.Background(), &tgmodels.Update{ID: 931, MyChatMember: &tgmodels.ChatMemberUpdated{
		Chat:          chat,
		NewChatMember: tgmodels.ChatMember{Type: tgmodels.ChatMemberTypeMember},
		OldChatMember: tgmodels.ChatMember{Type: tgmodels.ChatMemberTypeMember},
	}})
	if len(sent) != before {
		t.Fatalf("expected no reply for an ordinary member update, got %v", sent[before:])
	}

	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedCampaign(t, 7, "Sunken Crown", "The Party", 1)
	groupChat := chat.ID
	if err := UpsertCampaignTelegramSettings(CampaignTelegramSettings{CampaignID: 7, ChatID: &groupChat, IsEnabled: true}); err != nil {
		t.Fatalf("bind chat: %v", err)
	}
	HandleUpdate(context.Background(), &tgmodels.Update{ID: 932, MyChatMember: added})
	if last := sent[len(sent)-1]; !strings.Contains(last, "Sunken Crown") {
		t.Fatalf("expected connected campaign name, got %q", last)
	}
}
