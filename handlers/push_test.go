package handlers

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func pushRouter() *gin.Engine {
	return testutil.NewRouter(func(r *gin.RouterGroup) {
		r.POST("/push/subscribe", SubscribePush)
		r.POST("/push/unsubscribe", UnsubscribePush)
		r.GET("/campaigns/:id/push-mute", GetCampaignPushMute)
		r.PUT("/campaigns/:id/push-mute", SetCampaignPushMute)
		r.GET("/push-settings", GetPushSettings)
		r.POST("/push-settings", SavePushSettings)
		r.POST("/test-push", TestPush)
	})
}

// testKeys returns client keys shaped like real browser subscriptions: the
// library decodes p256dh as an uncompressed P-256 point before encrypting.
func testKeys() (string, string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	p256dh := base64.RawURLEncoding.EncodeToString(elliptic.Marshal(elliptic.P256(), priv.X, priv.Y))
	auth := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 16))
	return p256dh, auth
}

func validSubscription(endpoint string) map[string]any {
	p256dh, auth := testKeys()
	return map[string]any{
		"endpoint": endpoint,
		"keys":     map[string]any{"p256dh": p256dh, "auth": auth},
	}
}

// configureTestVAPID stores generated keys so delivery paths run.
func configureTestVAPID(t *testing.T) {
	t.Helper()
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("generate vapid: %v", err)
	}
	if err := setAppSetting(settingVapidPrivateKey, priv); err != nil {
		t.Fatalf("store private key: %v", err)
	}
	if err := setAppSetting(settingVapidPublicKey, pub); err != nil {
		t.Fatalf("store public key: %v", err)
	}
}

func insertSubscription(t *testing.T, userID int64, endpoint string) int64 {
	t.Helper()
	p256dh, auth := testKeys()
	res, err := db.DB.Exec(
		"INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth) VALUES (?, ?, ?, ?)",
		userID, endpoint, p256dh, auth)
	if err != nil {
		t.Fatalf("insert subscription: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// pushTargetServer is a stand-in push service: it records hits and replies
// with the given status.
func pushTargetServer(t *testing.T, status int, hits chan<- struct{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits <- struct{}{}
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestPushSubscribeUpsertUnsubscribe(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := pushRouter()

	w := testutil.PostJSON(t, r, "/api/push/subscribe", validSubscription("https://push.example/s/1"))
	testutil.AssertStatus(t, w, http.StatusCreated)
	if got := testutil.CountRows(t, "push_subscriptions"); got != 1 {
		t.Fatalf("expected 1 subscription, got %d", got)
	}

	// Same endpoint re-subscribes in place with fresh keys.
	updated := validSubscription("https://push.example/s/1")
	updated["keys"] = map[string]any{"p256dh": "new-p256dh", "auth": "new-auth"}
	w = testutil.PostJSON(t, r, "/api/push/subscribe", updated)
	testutil.AssertStatus(t, w, http.StatusCreated)
	if got := testutil.CountRows(t, "push_subscriptions"); got != 1 {
		t.Fatalf("duplicate endpoint created a second row: %d", got)
	}
	var p256dh string
	db.DB.QueryRow("SELECT p256dh FROM push_subscriptions WHERE endpoint = ?", "https://push.example/s/1").Scan(&p256dh)
	if p256dh != "new-p256dh" {
		t.Fatalf("upsert did not update keys: %q", p256dh)
	}

	w = testutil.PostJSON(t, r, "/api/push/unsubscribe", map[string]any{"endpoint": "https://push.example/s/1"})
	testutil.AssertStatus(t, w, http.StatusOK)
	if got := testutil.CountRows(t, "push_subscriptions"); got != 0 {
		t.Fatalf("unsubscribe left %d rows", got)
	}
}

func TestPushSubscribeInvalidPayload(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	r := pushRouter()

	w := testutil.PostJSON(t, r, "/api/push/subscribe", map[string]any{"keys": map[string]any{"p256dh": "x", "auth": "y"}})
	testutil.AssertStatus(t, w, http.StatusBadRequest)

	w = testutil.PostJSON(t, r, "/api/push/unsubscribe", map[string]any{})
	testutil.AssertStatus(t, w, http.StatusBadRequest)
}

func TestCampaignPushMuteRoundTrip(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 7, "Campaign", "Party", 1)
	r := pushRouter()

	var m map[string]any
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/campaigns/7/push-mute"), &m)
	testutil.AssertField(t, m, "muted", false)

	testutil.ParseJSON(t, testutil.PutJSON(t, r, "/api/campaigns/7/push-mute", map[string]any{"muted": true}), &m)
	testutil.AssertField(t, m, "muted", true)
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/campaigns/7/push-mute"), &m)
	testutil.AssertField(t, m, "muted", true)
	if got := testutil.CountRows(t, "push_mutes"); got != 1 {
		t.Fatalf("expected 1 mute row, got %d", got)
	}

	testutil.PutJSON(t, r, "/api/campaigns/7/push-mute", map[string]any{"muted": false})
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/campaigns/7/push-mute"), &m)
	testutil.AssertField(t, m, "muted", false)
	if got := testutil.CountRows(t, "push_mutes"); got != 0 {
		t.Fatalf("unmute left %d rows", got)
	}
}

func TestSavePushSettingsAutoGeneratesKeys(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	r := pushRouter()

	var m map[string]any
	testutil.ParseJSON(t, testutil.Get(t, r, "/api/push-settings"), &m)
	testutil.AssertField(t, m, "has_public_key", false)

	body := map[string]any{"subject": "mailto:dm@example.com", "lead_minutes": 30, "session_lead_days": 2}
	w := testutil.PostJSON(t, r, "/api/push-settings", body)
	testutil.AssertStatus(t, w, http.StatusOK)
	testutil.ParseJSON(t, w, &m)
	testutil.AssertField(t, m, "has_private_key", true)
	testutil.AssertField(t, m, "lead_minutes", float64(30))
	testutil.AssertField(t, m, "session_lead_days", float64(2))

	// The private key must never appear in a response.
	priv, _ := appSetting(settingVapidPrivateKey)
	if priv == "" || strings.Contains(w.Body.String(), priv) {
		t.Fatal("private key missing from storage or leaked in response")
	}
	if got := pushReminderLeadMinutes(); got != 30 {
		t.Fatalf("lead minutes not persisted: %d", got)
	}
	if got := pushSessionReminderLeadDays(); got != 2 {
		t.Fatalf("session lead days not persisted: %d", got)
	}
}

func TestTestPushRequiresConfiguration(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	r := pushRouter()

	w := testutil.PostJSON(t, r, "/api/test-push", nil)
	testutil.AssertStatus(t, w, http.StatusBadRequest)

	configureTestVAPID(t)
	w = testutil.PostJSON(t, r, "/api/test-push", nil)
	testutil.AssertStatus(t, w, http.StatusOK)
	var m map[string]any
	testutil.ParseJSON(t, w, &m)
	testutil.AssertField(t, m, "sent", float64(0)) // no subscriptions yet
}

func TestSendPushToUserPrunesGoneEndpoint(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	configureTestVAPID(t)

	hits := make(chan struct{}, 8)
	srv := pushTargetServer(t, http.StatusGone, hits)
	insertSubscription(t, 1, srv.URL+"/gone")

	if sent := SendPushToUser(1, PushPayload{Title: "t"}); sent != 0 {
		t.Fatalf("expected 0 sent, got %d", sent)
	}
	select {
	case <-hits:
	default:
		t.Fatal("push service was never contacted")
	}
	if got := testutil.CountRows(t, "push_subscriptions"); got != 0 {
		t.Fatalf("410 endpoint not pruned: %d rows left", got)
	}
}

func TestSendPushToUserKeepsTransientFailures(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	configureTestVAPID(t)

	okHits := make(chan struct{}, 8)
	okSrv := pushTargetServer(t, http.StatusCreated, okHits)
	deadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := deadSrv.URL
	deadSrv.Close() // nothing listens anymore → transport error

	insertSubscription(t, 1, okSrv.URL+"/ok")
	insertSubscription(t, 1, deadURL+"/dead")

	sent := SendPushToUser(1, PushPayload{Title: "t"})
	if sent != 1 {
		t.Fatalf("expected exactly 1 successful send, got %d", sent)
	}
	if got := testutil.CountRows(t, "push_subscriptions"); got != 2 {
		t.Fatalf("transient failure should keep the row: %d left", got)
	}
}

func TestNotifyCampaignRecapPublishedSkipsMuted(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	configureTestVAPID(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "player", "player")
	testutil.SeedCampaign(t, 10, "Campaign", "Party", 1)
	testutil.SeedCampaignMember(t, 10, 2, "player")

	hits := make(chan struct{}, 16)
	srv := pushTargetServer(t, http.StatusCreated, hits)
	insertSubscription(t, 1, srv.URL+"/u1")
	insertSubscription(t, 2, srv.URL+"/u2")

	db.DB.Exec("INSERT OR IGNORE INTO push_mutes (user_id, campaign_id) VALUES (2, 10)")

	NotifyCampaignRecapPublished(10, 5, "The goblin king falls")

	timeout := time.After(5 * time.Second)
	received := 0
	for received < 1 {
		select {
		case <-hits:
			received++
		case <-timeout:
			t.Fatalf("fan-out timed out; received %d pushes, want exactly 1 (muted user skipped)", received)
		}
	}
	// Give any illegal second delivery a moment to arrive.
	select {
	case <-hits:
		t.Fatal("more than one push delivered")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestScanFeedEventRemindersHybridRules(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	configureTestVAPID(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 3, "Campaign", "Party", 1)
	hits := make(chan struct{}, 16)
	srv := pushTargetServer(t, http.StatusCreated, hits)
	insertSubscription(t, 1, srv.URL+"/u1")

	// campaign_event_settings row linking slug → campaign.
	db.DB.Exec(`INSERT OR IGNORE INTO campaign_event_settings (id, campaign_id, slug, display_name, calendar_id)
		VALUES (1, 3, 'waterdeep', 'Waterdeep Nights', 'cal@group.calendar.google.com')`)

	now := time.Now()
	inWindow := now.Add(30 * time.Minute).Format(time.RFC3339)
	outOfWindow := now.Add(3 * time.Hour).Format(time.RFC3339)
	past := now.Add(-time.Hour).Format(time.RFC3339)
	seedFeedEvent := func(id, start string, allDay int) {
		db.DB.Exec(`INSERT OR IGNORE INTO google_events_cache (event_id, title, start_time, all_day, campaign_slug)
			VALUES (?, ?, ?, ?, 'waterdeep')`, id, "Event "+id, start, allDay)
	}
	seedFeedEvent("evt-in", inWindow, 0)
	seedFeedEvent("evt-out", outOfWindow, 0)
	seedFeedEvent("evt-allday", inWindow, 1)
	seedFeedEvent("evt-past", past, 0)

	scanFeedEventReminders(now)

	// Exactly one reminder delivered: only evt-in qualifies.
	select {
	case <-hits:
	default:
		t.Fatal("in-window feed event produced no push")
	}
	if n := len(hits); n != 0 {
		t.Fatalf("unexpected extra pushes: %d", n+1)
	}
	var logged int
	db.DB.QueryRow("SELECT COUNT(*) FROM push_reminder_log WHERE source = 'feed_event'").Scan(&logged)
	if logged != 1 {
		t.Fatalf("expected 1 dedup log entry, got %d", logged)
	}

	// Re-scan: dedup must prevent a second send.
	scanFeedEventReminders(now.Add(time.Minute))
	if n := len(hits); n != 0 {
		t.Fatal("dedup failed: second scan delivered again")
	}
	logged = 0
	db.DB.QueryRow("SELECT COUNT(*) FROM push_reminder_log WHERE source = 'feed_event'").Scan(&logged)
	if logged != 1 {
		t.Fatalf("re-scan added log entries: %d", logged)
	}
}

func TestScanLocalSessionRemindersDayLead(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	configureTestVAPID(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 4, "Campaign", "Party", 1)
	hits := make(chan struct{}, 16)
	srv := pushTargetServer(t, http.StatusCreated, hits)
	insertSubscription(t, 1, srv.URL+"/u1")

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	nextWeek := time.Now().AddDate(0, 0, 6).Format("2006-01-02")
	seedLocalSession := func(id int64, date string) {
		db.DB.Exec(`INSERT OR IGNORE INTO campaign_calendar_events (id, campaign_id, title, event_date, event_type)
			VALUES (?, 4, ?, ?, 'session')`, id, "Session "+fmt.Sprint(id), date)
	}
	seedLocalSession(1, tomorrow)
	seedLocalSession(2, nextWeek)

	scanLocalSessionReminders(time.Now())

	select {
	case <-hits:
	default:
		t.Fatal("tomorrow's session produced no 'Session coming up' push")
	}
	if n := len(hits); n != 0 {
		t.Fatalf("next week's session reminded too early: %d extra", n+1)
	}

	// Dedup on re-scan.
	scanLocalSessionReminders(time.Now().Add(time.Hour))
	if n := len(hits); n != 0 {
		t.Fatal("local session dedup failed")
	}
}

func TestRecapCreateAndMarkSentTriggerPush(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	configureTestVAPID(t)

	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "player", "player")
	testutil.SeedCampaign(t, 10, "Campaign", "Party", 1)
	testutil.SeedCampaignMember(t, 10, 2, "player")

	hits := make(chan struct{}, 16)
	srv := pushTargetServer(t, http.StatusCreated, hits)
	insertSubscription(t, 1, srv.URL+"/u1")
	insertSubscription(t, 2, srv.URL+"/u2")

	r := testutil.NewRouter(func(rg *gin.RouterGroup) {
		rg.POST("/campaigns/:id/recaps", CreateCampaignRecap)
		rg.POST("/recaps/:id/mark-sent", MarkRecapAsSent)
	})

	w := testutil.PostJSON(t, r, "/api/campaigns/10/recaps", map[string]any{
		"title":   "The goblin king falls",
		"content": "We fought a dragon and lived.",
	})
	testutil.AssertStatus(t, w, http.StatusCreated)

	// Both subscribed members get exactly one push from recap creation.
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 2 {
		select {
		case <-hits:
			received++
		case <-timeout:
			t.Fatalf("recap create fan-out timed out; received %d pushes, want 2", received)
		}
	}
	select {
	case <-hits:
		t.Fatal("more than two pushes delivered on create")
	case <-time.After(200 * time.Millisecond):
	}

	// Marking the recap as sent notifies again (explicit publish moment).
	w = testutil.PostJSON(t, r, "/api/recaps/1/mark-sent", nil)
	testutil.AssertStatus(t, w, http.StatusOK)

	received = 0
	timeout = time.After(5 * time.Second)
	for received < 2 {
		select {
		case <-hits:
			received++
		case <-timeout:
			t.Fatalf("mark-sent fan-out timed out; received %d pushes, want 2", received)
		}
	}
}
