package handlers

// Web Push notifications: VAPID configuration (admin), browser subscription
// lifecycle, per-campaign mute, and the two trigger fan-outs (recap published,
// session reminders). Delivery uses github.com/SherClockHolmes/webpush-go.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"

	"villum/db"
)

// app_settings keys for push configuration.
const (
	settingVapidPublicKey      = "vapid_public_key"
	settingVapidPrivateKey     = "vapid_private_key"
	settingVapidSubject        = "vapid_subject"
	settingPushLeadMinutes     = "push_reminder_lead_minutes"
	settingPushSessionLeadDays = "push_session_reminder_lead_days"
)

const (
	defaultPushLeadMinutes     = 60
	defaultPushSessionLeadDays = 1
	pushHTTPTimeout            = 10 * time.Second
	pushTTLSeconds             = 3600
	fallbackVapidSubject       = "mailto:admin@villum.local"
)

var (
	errPushGone    = errors.New("push endpoint gone")
	errPushExpired = errors.New("push subscription expired")
)

// PushPayload is the JSON delivered to the service worker.
type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag,omitempty"`
}

// ─── Settings accessors ───

func appSetting(key string) (string, bool) {
	var value string
	if err := db.DB.QueryRow("SELECT value FROM app_settings WHERE key = ?", key).Scan(&value); err != nil {
		return "", false
	}
	return value, true
}

func setAppSetting(key, value string) error {
	_, err := db.DB.Exec("INSERT OR REPLACE INTO app_settings (key, value) VALUES (?, ?)", key, value)
	return err
}

func deleteAppSetting(key string) {
	db.DB.Exec("DELETE FROM app_settings WHERE key = ?", key)
}

// pushReminderLeadMinutes is the minute-precision lead for timed external
// feed events.
func pushReminderLeadMinutes() int {
	value, ok := appSetting(settingPushLeadMinutes)
	if !ok {
		return defaultPushLeadMinutes
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return defaultPushLeadMinutes
	}
	return n
}

// pushSessionReminderLeadDays is the day-granular lead for local calendar
// sessions ("Session coming up").
func pushSessionReminderLeadDays() int {
	value, ok := appSetting(settingPushSessionLeadDays)
	if !ok {
		return defaultPushSessionLeadDays
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return defaultPushSessionLeadDays
	}
	return n
}

type vapidConfig struct {
	Public  string
	Private string
	Subject string
}

// loadVAPID reads the stored VAPID config; ok is false when keys were never
// saved. The private key never leaves this struct except into webpush.Options.
func loadVAPID() (vapidConfig, bool) {
	pub, okPub := appSetting(settingVapidPublicKey)
	priv, okPriv := appSetting(settingVapidPrivateKey)
	if !okPub || !okPriv || pub == "" || priv == "" {
		return vapidConfig{}, false
	}
	subject, _ := appSetting(settingVapidSubject)
	if subject == "" {
		subject = fallbackVapidSubject
	}
	return vapidConfig{Public: pub, Private: priv, Subject: subject}, true
}

// ─── Delivery ───

type pushSubscriptionRow struct {
	ID             int64
	UserID         int64
	Endpoint       string
	P256dh         string
	Auth           string
	ExpirationTime int64 // unix seconds; 0 = no expiration
}

func subscriptionsForUser(userID int64) ([]pushSubscriptionRow, error) {
	rows, err := db.DB.Query(
		"SELECT id, user_id, endpoint, p256dh, auth, expiration_time FROM push_subscriptions WHERE user_id = ?",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []pushSubscriptionRow
	for rows.Next() {
		var s pushSubscriptionRow
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.ExpirationTime); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func deleteSubscription(id int64) {
	db.DB.Exec("DELETE FROM push_subscriptions WHERE id = ?", id)
}

// deliverToSubscription sends one payload to one subscription. It returns
// errPushGone/errPushExpired when the row must be pruned.
func deliverToSubscription(v vapidConfig, s pushSubscriptionRow, payload PushPayload) error {
	if s.ExpirationTime > 0 && time.Now().Unix() >= s.ExpirationTime {
		return errPushExpired
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := webpush.SendNotification(data, &webpush.Subscription{
		Endpoint: s.Endpoint,
		Keys:     webpush.Keys{P256dh: s.P256dh, Auth: s.Auth},
	}, &webpush.Options{
		Subscriber:      v.Subject,
		VAPIDPublicKey:  v.Public,
		VAPIDPrivateKey: v.Private,
		TTL:             pushTTLSeconds,
		Urgency:         webpush.UrgencyNormal,
		HTTPClient:      &http.Client{Timeout: pushHTTPTimeout},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return errPushGone
	case resp.StatusCode >= 400:
		return fmt.Errorf("push service returned %d", resp.StatusCode)
	}
	return nil
}

// SendPushToUser delivers a payload to every subscription of a user,
// pruning dead/expired rows. Returns how many deliveries succeeded.
func SendPushToUser(userID int64, payload PushPayload) int {
	v, ok := loadVAPID()
	if !ok {
		return 0
	}
	subs, err := subscriptionsForUser(userID)
	if err != nil {
		return 0
	}
	sent := 0
	for _, s := range subs {
		err := deliverToSubscription(v, s, payload)
		switch {
		case errors.Is(err, errPushExpired), errors.Is(err, errPushGone):
			deleteSubscription(s.ID)
		case err != nil:
			// Transient failure (network, 4xx/5xx): keep the subscription.
		default:
			sent++
		}
	}
	return sent
}

// subscribedUnmutedMemberIDs returns campaign members that have at least one
// subscription and have not muted the campaign.
func subscribedUnmutedMemberIDs(campaignID int64) ([]int64, error) {
	rows, err := db.DB.Query(`
		SELECT DISTINCT cm.user_id FROM campaign_members cm
		WHERE cm.campaign_id = ?
		  AND NOT EXISTS (SELECT 1 FROM push_mutes m WHERE m.user_id = cm.user_id AND m.campaign_id = cm.campaign_id)
		  AND EXISTS (SELECT 1 FROM push_subscriptions s WHERE s.user_id = cm.user_id)`,
		campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// NotifyCampaignRecapPublished fans out a recap notification to subscribed,
// unmuted members. Fire-and-forget: callers must not block on delivery.
func NotifyCampaignRecapPublished(campaignID, recapID int64, recapTitle string) {
	go func() {
		members, err := subscribedUnmutedMemberIDs(campaignID)
		if err != nil || len(members) == 0 {
			return
		}
		payload := PushPayload{
			Title: "Recap published",
			Body:  recapTitle,
			URL:   "/#/campaignOverview",
			Tag:   fmt.Sprintf("recap-%d", recapID),
		}
		for _, memberID := range members {
			SendPushToUser(memberID, payload)
		}
	}()
}

// SendTestPushToUser sends an admin test push to the caller's own devices.
func SendTestPushToUser(userID int64) (int, error) {
	if _, ok := loadVAPID(); !ok {
		return 0, errors.New("push is not configured")
	}
	payload := PushPayload{
		Title: "Villum test push",
		Body:  "If you can read this, push works.",
		URL:   "/",
		Tag:   "villum-test",
	}
	return SendPushToUser(userID, payload), nil
}

// ─── Subscription & mute endpoints ───

type subscribeRequest struct {
	Endpoint       string `json:"endpoint"`
	ExpirationTime int64  `json:"expirationTime"`
	Keys           struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// SubscribePush registers or updates a browser push subscription for the
// authenticated user. The endpoint is unique; re-subscribing upserts in place.
func SubscribePush(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription payload"})
		return
	}
	_, err := db.DB.Exec(`
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, expiration_time)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(endpoint) DO UPDATE SET
			user_id = excluded.user_id,
			p256dh = excluded.p256dh,
			auth = excluded.auth,
			expiration_time = excluded.expiration_time`,
		uid, req.Endpoint, req.Keys.P256dh, req.Keys.Auth, req.ExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store subscription"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "subscribed"})
}

// UnsubscribePush removes the given endpoint if it belongs to the caller.
func UnsubscribePush(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	db.DB.Exec("DELETE FROM push_subscriptions WHERE endpoint = ? AND user_id = ?", req.Endpoint, uid)
	c.JSON(http.StatusOK, gin.H{"status": "unsubscribed"})
}

// GetCampaignPushMute reports whether the authenticated user muted pushes
// for this campaign.
func GetCampaignPushMute(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || campaignID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	var muted int
	db.DB.QueryRow("SELECT COUNT(*) FROM push_mutes WHERE user_id = ? AND campaign_id = ?", uid, campaignID).Scan(&muted)
	c.JSON(http.StatusOK, gin.H{"muted": muted > 0})
}

// SetCampaignPushMute toggles the caller's per-campaign push mute.
func SetCampaignPushMute(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || campaignID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	var req struct {
		Muted bool `json:"muted"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Muted {
		if _, err := db.DB.Exec("INSERT OR IGNORE INTO push_mutes (user_id, campaign_id) VALUES (?, ?)", uid, campaignID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save mute"})
			return
		}
	} else {
		if _, err := db.DB.Exec("DELETE FROM push_mutes WHERE user_id = ? AND campaign_id = ?", uid, campaignID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save mute"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"muted": req.Muted})
}

// GetVapidPublicKey exposes the VAPID public key to any authenticated user —
// browsers need it to subscribe. Only the public half is ever returned.
func GetVapidPublicKey(c *gin.Context) {
	pub, ok := appSetting(settingVapidPublicKey)
	if !ok || pub == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "push is not configured"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"public_key": pub})
}

// ─── Admin VAPID configuration ───

// GetPushSettings returns the admin-facing push configuration. The private
// key is only reported as a boolean.
func GetPushSettings(c *gin.Context) {
	pub, hasPub := appSetting(settingVapidPublicKey)
	_, hasPriv := appSetting(settingVapidPrivateKey)
	subject, _ := appSetting(settingVapidSubject)
	c.JSON(http.StatusOK, gin.H{
		"public_key":          pub,
		"has_public_key":      hasPub && pub != "",
		"has_private_key":     hasPriv,
		"subject":             subject,
		"lead_minutes":        pushReminderLeadMinutes(),
		"session_lead_days":   pushSessionReminderLeadDays(),
	})
}

type savePushSettingsRequest struct {
	Subject         string `json:"subject"`
	LeadMinutes     *int   `json:"lead_minutes"`
	SessionLeadDays *int   `json:"session_lead_days"`
	GenerateKeys    bool   `json:"generate_keys"`
}

// SavePushSettings persists subject + lead times and auto-generates VAPID
// keys when absent (or when generate_keys is set). The private key is stored,
// never logged, never returned.
func SavePushSettings(c *gin.Context) {
	var req savePushSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.GenerateKeys {
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate VAPID keys"})
			return
		}
		if err := setAppSetting(settingVapidPrivateKey, privateKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
		if err := setAppSetting(settingVapidPublicKey, publicKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
	} else if _, ok := loadVAPID(); !ok {
		// First save with no keys configured yet: auto-generate so the
		// feature works without a manual generation step.
		privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate VAPID keys"})
			return
		}
		if err := setAppSetting(settingVapidPrivateKey, privateKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
		if err := setAppSetting(settingVapidPublicKey, publicKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
	}

	if req.Subject != "" {
		if err := setAppSetting(settingVapidSubject, req.Subject); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
	}
	if req.LeadMinutes != nil && *req.LeadMinutes >= 1 {
		if err := setAppSetting(settingPushLeadMinutes, strconv.Itoa(*req.LeadMinutes)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
	}
	if req.SessionLeadDays != nil && *req.SessionLeadDays >= 0 {
		if err := setAppSetting(settingPushSessionLeadDays, strconv.Itoa(*req.SessionLeadDays)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
			return
		}
	}

	GetPushSettings(c)
}

// TestPush sends a test notification to the calling admin's own devices.
func TestPush(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sent, err := SendTestPushToUser(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": sent})
}

// ─── Session reminder scheduler (hybrid sources) ───

type feedEventRow struct {
	EventID     string
	CampaignID  int64
	Title       string
	StartTime   string
	AllDay      int
}

type localSessionRow struct {
	ID        int64
	CampaignID int64
	Title     string
	EventDate string
}

// recordReminderSent marks an event as reminded (dedup marker). Returns true
// if this call was the one that claimed it.
func recordReminderSent(source, eventKey string) bool {
	res, err := db.DB.Exec("INSERT OR IGNORE INTO push_reminder_log (source, event_key) VALUES (?, ?)", source, eventKey)
	if err != nil {
		return false
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false
	}
	return n > 0
}

// scanFeedEventReminders sends minute-precision reminders for timed external
// feed events entering their lead window. The cache stores a campaign slug,
// so rows are resolved to campaigns through campaign_event_settings.
func scanFeedEventReminders(now time.Time) {
	rows, err := db.DB.Query(`
		SELECT gec.event_id, ces.campaign_id, COALESCE(gec.title, ''), COALESCE(gec.start_time, ''), COALESCE(gec.all_day, 0)
		FROM google_events_cache gec
		JOIN campaign_event_settings ces ON ces.slug = gec.campaign_slug
		WHERE gec.start_time != ''`)
	if err != nil {
		return
	}
	defer rows.Close()
	lead := time.Duration(pushReminderLeadMinutes()) * time.Minute
	for rows.Next() {
		var e feedEventRow
		if err := rows.Scan(&e.EventID, &e.CampaignID, &e.Title, &e.StartTime, &e.AllDay); err != nil {
			continue
		}
		if e.AllDay == 1 {
			continue
		}
		start, ok := parseFeedTime(e.StartTime)
		if !ok || start.Before(now) {
			continue
		}
		if now.Before(start.Add(-lead)) {
			continue
		}
		key := fmt.Sprintf("%d:%s", e.CampaignID, e.EventID)
		if !recordReminderSent("feed_event", key) {
			continue
		}
		notifyEventMembers(e.CampaignID, PushPayload{
			Title: "Session starting soon",
			Body:  e.Title,
			URL:   fmt.Sprintf("/events/%s", e.EventID),
			Tag:   "reminder-" + key,
		})
	}
}

// scanLocalSessionReminders sends day-granular "Session coming up" reminders
// for local calendar sessions within their day lead.
func scanLocalSessionReminders(now time.Time) {
	rows, err := db.DB.Query(`
		SELECT id, campaign_id, COALESCE(title, ''), event_date
		FROM campaign_calendar_events
		WHERE event_type = 'session' AND event_date >= date('now')`)
	if err != nil {
		return
	}
	defer rows.Close()
	days := pushSessionReminderLeadDays()
	for rows.Next() {
		var e localSessionRow
		if err := rows.Scan(&e.ID, &e.CampaignID, &e.Title, &e.EventDate); err != nil {
			continue
		}
		eventDate, err := time.ParseInLocation("2006-01-02", e.EventDate, time.Local)
		if err != nil {
			continue
		}
		// Reminder window opens at the start of (eventDate - days).
		windowStart := eventDate.AddDate(0, 0, -days)
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		if todayStart.Before(windowStart) {
			continue
		}
		key := fmt.Sprintf("%d:%d", e.CampaignID, e.ID)
		if !recordReminderSent("local_session", key) {
			continue
		}
		notifyEventMembers(e.CampaignID, PushPayload{
			Title: "Session coming up",
			Body:  e.Title,
			URL:   "/#/timeline",
			Tag:   "reminder-" + key,
		})
	}
}

// notifyEventMembers fans a reminder out to subscribed, unmuted members of a
// campaign (synchronous variant used by the scheduler loop).
func notifyEventMembers(campaignID int64, payload PushPayload) {
	members, err := subscribedUnmutedMemberIDs(campaignID)
	if err != nil {
		return
	}
	for _, memberID := range members {
		SendPushToUser(memberID, payload)
	}
}

// parseFeedTime parses the datetime formats found in google_events_cache
// start_time (RFC3339 from the Google API, or "2006-01-02 15:04").
func parseFeedTime(value string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// StartPushReminderScheduler runs the hybrid reminder scans once a minute,
// mirroring StartBackupScheduler's lifecycle.
func StartPushReminderScheduler(stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case now := <-ticker.C:
				if _, ok := loadVAPID(); !ok {
					continue
				}
				scanFeedEventReminders(now)
				scanLocalSessionReminders(now)
			}
		}
	}()
}
