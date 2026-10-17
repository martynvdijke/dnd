package telegram

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func TestDeliverRecapWithMedia(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "C1", "P1", 1)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled) VALUES(10,777,1)"); err != nil {
		t.Fatalf("settings: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content) VALUES(100,10,'T','hello world')"); err != nil {
		t.Fatalf("recap: %v", err)
	}
	for _, uid := range []int64{1, 2} {
		db.DB.Exec("INSERT OR IGNORE INTO uploads(id,hash,ext,url, resized_url, thumbnail_url, kind) VALUES(?,?,?, ?, '', '', 'image')", uid, "h"+string(rune('0'+uid)), ".jpg", "u"+string(rune('0'+uid)))
	}
	SetMediaLoader(func(id int64) (*models.MediaFile, error) {
		return &models.MediaFile{Data: []byte("fake"), Filename: "pic.jpg", Kind: "image", Ext: ".jpg"}, nil
	})
	t.Cleanup(func() { SetMediaLoader(nil) })
	var calls []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	DeliverRecapWithMedia(100, "manual", []int64{1, 2})
	for i := 0; i < 20; i++ {
		time.Sleep(200 * time.Millisecond)
		var c int
		db.DB.QueryRow("SELECT COUNT(*) FROM telegram_media_posts WHERE recap_id=100").Scan(&c)
		if c == 2 {
			break
		}
	}
	// check text delivery status sent
	var status string
	if err := db.DB.QueryRow("SELECT status FROM telegram_deliveries WHERE recap_id=100 AND target_chat_id=777").Scan(&status); err != nil {
		t.Fatalf("delivery row: %v", err)
	}
	if status != "sent" {
		t.Fatalf("expected sent, got %q", status)
	}
	// media rows
	var cnt int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_media_posts WHERE recap_id=100").Scan(&cnt)
	if cnt != 2 {
		t.Fatalf("expected 2 media rows, got %d", cnt)
	}
	// verify PostMedia used chat type 'chat' and recap_id set
	var targetType string
	db.DB.QueryRow("SELECT target_type FROM telegram_media_posts WHERE recap_id=100 LIMIT 1").Scan(&targetType)
	if targetType != "chat" {
		t.Fatalf("expected target_type chat, got %q", targetType)
	}
	// ensure both sendMessage and media send called
	joined := strings.Join(calls, ",")
	if !strings.Contains(joined, "sendMessage") {
		t.Fatalf("expected sendMessage, got %v", calls)
	}
	if !strings.Contains(joined, "sendPhoto") && !strings.Contains(joined, "sendMediaGroup") {
		t.Fatalf("expected sendPhoto or sendMediaGroup for media, got %v", calls)
	}
}

func TestDeliverRecapMediaFailureDoesNotAffectText(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "C1", "P1", 1)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled) VALUES(10,777,1)"); err != nil {
		t.Fatalf("settings: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_recaps(id,campaign_id,title,content) VALUES(101,10,'T','hello')"); err != nil {
		t.Fatalf("recap: %v", err)
	}
	db.DB.Exec("INSERT OR IGNORE INTO uploads(id,hash,ext,url, resized_url, thumbnail_url, kind) VALUES(?,?,?,?, '', '', 'image')", int64(10), "hx10", ".jpg", "ux")
	db.DB.Exec("INSERT OR IGNORE INTO uploads(id,hash,ext,url, resized_url, thumbnail_url, kind) VALUES(?,?,?,?, '', '', 'document')", int64(11), "hx11", ".pdf", "ux")
	SetMediaLoader(func(id int64) (*models.MediaFile, error) {
		if id == 11 {
			return &models.MediaFile{Data: []byte("fake"), Filename: "doc.pdf", Kind: "document", Ext: ".pdf"}, nil
		}
		return &models.MediaFile{Data: []byte("fake"), Filename: "pic.jpg", Kind: "image", Ext: ".jpg"}, nil
	})
	t.Cleanup(func() { SetMediaLoader(nil) })
	// Make second upload fail
	origImpl := SendMediaImpl
	SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		// First call succeeds, but second fails based on count
		// Use a counter: we send two uploads, second fails
		// Since SendMediaImpl is called per upload, fail for upload 2
		// We don't have upload id here, so fail on second invocation overall
		// Use closure counter
		return 0, nil // placeholder; will override with stateful
	}
	// stateful wrapper
	calls := 0
	SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		calls++
		if calls == 2 {
			return 0, &RetryAfterError{Desc: "fail"}
		}
		return 1, nil
	}
	t.Cleanup(func() { SendMediaImpl = origImpl })
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	DeliverRecapWithMedia(101, "manual", []int64{10, 11})
	time.Sleep(600 * time.Millisecond)
	var status string
	db.DB.QueryRow("SELECT status FROM telegram_deliveries WHERE recap_id=101").Scan(&status)
	if status != "sent" {
		t.Fatalf("text delivery should remain sent despite media failure, got %q", status)
	}
	var failed int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_media_posts WHERE recap_id=101 AND status='failed'").Scan(&failed)
	if failed != 1 {
		t.Fatalf("expected 1 failed media row, got %d", failed)
	}
	var sent int
	db.DB.QueryRow("SELECT COUNT(*) FROM telegram_media_posts WHERE recap_id=101 AND status='sent'").Scan(&sent)
	if sent != 1 {
		t.Fatalf("expected 1 sent media row, got %d", sent)
	}
}
