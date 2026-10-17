package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
	"villum/telegram"
)

func seedUpload(t *testing.T, id int64, kind, ext, name string, data []byte) int64 {
	t.Helper()
	if data == nil {
		data = []byte("test-data-" + fmt.Sprint(id))
	}
	// Ensure extension
	if ext == "" {
		ext = ".jpg"
	}
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])
	// Write file to media path if set
	if MediaPath != "" {
		sub := hashStr[:2]
		dir := filepath.Join(MediaPath, sub)
		_ = os.MkdirAll(dir, 0755)
		filename := hashStr + ext
		_ = os.WriteFile(filepath.Join(dir, filename), data, 0644)
	}
	url := fmt.Sprintf("/media/%s/%s%s", hashStr[:2], hashStr, ext)
	// Insert upload row
	_, err := db.DB.Exec(
		`INSERT OR IGNORE INTO uploads(id, hash, ext, url, resized_url, thumbnail_url, owner_type, owner_id, kind, source, telegram_file_unique_id) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		id, hashStr, ext, url, url, "", "campaign", 10, kind, "upload", "",
	)
	if err != nil {
		t.Fatalf("seed upload: %v", err)
	}
	// Install loader for this upload
	return id
}

func setupMediaLoader(t *testing.T) {
	t.Helper()
	telegram.SetMediaLoader(func(uploadID int64) (*models.MediaFile, error) {
		var hash, ext, kind string
		err := db.DB.QueryRow("SELECT hash, ext, COALESCE(kind,'image') FROM uploads WHERE id=?", uploadID).Scan(&hash, &ext, &kind)
		if err != nil {
			return nil, fmt.Errorf("upload not found")
		}
		sub := hash[:2]
		// Normalize ext like upload code does? Keep as stored.
		saveExt := ext
		if saveExt == ".jpeg" {
			saveExt = ".jpg"
		}
		path := filepath.Join(MediaPath, sub, hash+saveExt)
		data, err := os.ReadFile(path)
		if err != nil {
			// fallback: return synthetic data
			data = []byte("fallback-" + fmt.Sprint(uploadID))
		}
		filename := fmt.Sprintf("file%d%s", uploadID, ext)
		return &models.MediaFile{Data: data, Filename: filename, Kind: kind, Ext: ext}, nil
	})
	t.Cleanup(func() { telegram.SetMediaLoader(nil) })
}

func TestTelegramTargets(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "Camp One", "P1", 1)
	testutil.SeedCampaign(t, 11, "Camp Two", "P2", 1)
	// Bound and enabled for 10, disabled for 11
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled, title_cache) VALUES(10, 12345, 1, 'ChatOne')"); err != nil {
		t.Fatalf("insert cts: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled, title_cache) VALUES(11, 99999, 0, 'ChatTwo')"); err != nil {
		t.Fatalf("insert cts2: %v", err)
	}
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/telegram/targets", GetTelegramTargets)
	})
	w := testutil.Get(t, r, "/api/telegram/targets")
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	camps, _ := resp["campaigns"].([]any)
	if len(camps) != 1 {
		t.Fatalf("expected 1 enabled campaign, got %d: %v", len(camps), resp)
	}
	dm, _ := resp["dm"].(map[string]any)
	if dm["linked"] != false {
		t.Fatalf("expected dm linked false")
	}
	// Link identity
	if _, err := db.DB.Exec("INSERT INTO telegram_identities(user_id, telegram_user_id, telegram_chat_id) VALUES(1,111,111)"); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	w2 := testutil.Get(t, r, "/api/telegram/targets")
	var resp2 map[string]any
	testutil.ParseJSON(t, w2, &resp2)
	dm2, _ := resp2["dm"].(map[string]any)
	if dm2["linked"] != true {
		t.Fatalf("expected dm linked true, got %v", resp2)
	}
}

func TestTelegramMediaCampaignSuccess(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "Camp One", "P1", 1)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled, title_cache) VALUES(10, 55555, 1, 'ChatOne')"); err != nil {
		t.Fatalf("insert cts: %v", err)
	}
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	seedUpload(t, 100, "image", ".jpg", "a.jpg", []byte("imgdata1"))
	seedUpload(t, 101, "image", ".jpg", "b.jpg", []byte("imgdata2"))
	setupMediaLoader(t)
	var sendCalls int
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		sendCalls++
		return int64(1000 + sendCalls), nil
	}
	telegram.SendMediaGroupImpl = func(chatID int64, caption string, items []telegram.MediaItem) ([]int64, error) {
		sendCalls++
		ids := make([]int64, len(items))
		for i := range items {
			ids[i] = int64(2000 + i)
		}
		return ids, nil
	}
	t.Cleanup(func() {
		telegram.SendMediaImpl = nil
		telegram.SendMediaGroupImpl = nil
	})

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	})
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{100},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 10},
	})
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	results, _ := resp["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %v", resp)
	}
	first, _ := results[0].(map[string]any)
	if first["status"] != "sent" {
		t.Fatalf("expected sent, got %v", first)
	}
	if _, ok := first["telegram_message_id"]; !ok {
		t.Fatalf("missing telegram_message_id")
	}
	// Check DB row
	var status string
	var msgID *int64
	_ = db.DB.QueryRow("SELECT status, telegram_message_id FROM telegram_media_posts WHERE upload_id=100").Scan(&status, &msgID)
	if status != "sent" {
		t.Fatalf("expected db status sent, got %q", status)
	}
}

func TestTelegramMediaDM(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if _, err := db.DB.Exec("INSERT INTO telegram_identities(user_id, telegram_user_id, telegram_chat_id) VALUES(1,777,777)"); err != nil {
		t.Fatalf("seed id: %v", err)
	}
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	seedUpload(t, 200, "image", ".jpg", "c.jpg", []byte("imgdata3"))
	setupMediaLoader(t)
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		if chatID != 777 {
			t.Fatalf("expected dm chat 777, got %d", chatID)
		}
		return 999, nil
	}
	t.Cleanup(func() { telegram.SendMediaImpl = nil })
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	})
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{200},
		"target":     map[string]any{"kind": "dm"},
	})
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	results, _ := resp["results"].([]any)
	if results[0].(map[string]any)["status"] != "sent" {
		t.Fatalf("expected sent dm")
	}
}

func TestTelegramMediaNonDMForbidden(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 20, "Camp20", "P20", 1)
	// Make user2 a player not dm
	if _, err := db.DB.Exec("INSERT INTO campaign_members(campaign_id, user_id, role) VALUES(20, 2, 'player')"); err != nil {
		t.Fatalf("seed member: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(20, 555, 1)"); err != nil {
		t.Fatalf("insert cts: %v", err)
	}
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	seedUpload(t, 300, "image", ".jpg", "d.jpg", []byte("imgdata4"))
	setupMediaLoader(t)
	var called bool
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		called = true
		return 1, nil
	}
	t.Cleanup(func() { telegram.SendMediaImpl = nil })
	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	}, 2, "user")
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{300},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 20},
	})
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d %s", w.Code, w.Body.String())
	}
	if called {
		t.Fatalf("should not have called Telegram")
	}
	cnt := testutil.CountRows(t, "telegram_media_posts")
	if cnt != 0 {
		t.Fatalf("expected no rows, got %d", cnt)
	}
}

func TestTelegramMediaOversizedPreReject(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "Camp", "P", 1)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(10, 55555, 1)"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	// Create oversized image (>10MB)
	big := bytes.Repeat([]byte("a"), 10*1024*1024+1)
	seedUpload(t, 400, "image", ".jpg", "big.jpg", big)
	setupMediaLoader(t)
	var called bool
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		called = true
		return 1, nil
	}
	telegram.SendMediaGroupImpl = func(chatID int64, caption string, items []telegram.MediaItem) ([]int64, error) {
		called = true
		return []int64{1}, nil
	}
	t.Cleanup(func() {
		telegram.SendMediaImpl = nil
		telegram.SendMediaGroupImpl = nil
	})
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	})
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{400},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 10},
	})
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	results, _ := resp["results"].([]any)
	if len(results) != 1 || results[0].(map[string]any)["status"] != "failed" {
		t.Fatalf("expected failed for oversized, got %v", resp)
	}
	if called {
		t.Fatalf("oversized should not have called Telegram")
	}
	var status string
	_ = db.DB.QueryRow("SELECT status FROM telegram_media_posts WHERE upload_id=400").Scan(&status)
	if status != "failed" {
		t.Fatalf("expected db failed, got %q", status)
	}
}

func TestTelegramMediaPerFileFailure(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "Camp", "P", 1)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(10, 55555, 1)"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	seedUpload(t, 500, "document", ".pdf", "a.pdf", []byte("pdf1"))
	seedUpload(t, 501, "document", ".pdf", "b.pdf", []byte("pdf2"))
	setupMediaLoader(t)
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		if filename == "file500.pdf" {
			return 0, fmt.Errorf("telegram error for 500")
		}
		return 1234, nil
	}
	t.Cleanup(func() { telegram.SendMediaImpl = nil })
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	})
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{500, 501},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 10},
	})
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	results, _ := resp["results"].([]any)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %v", resp)
	}
	m0 := results[0].(map[string]any)
	m1 := results[1].(map[string]any)
	if m0["status"] != "failed" {
		t.Fatalf("expected first failed, got %v", m0)
	}
	if m1["status"] != "sent" {
		t.Fatalf("expected second sent, got %v", m1)
	}
}

// IDOR: unauthorized upload access is forbidden.

func seedUploadOwned(t *testing.T, id int64, ownerType string, ownerID int64) {
	t.Helper()
	hash := fmt.Sprintf("%064x", id)
	_, err := db.DB.Exec(`INSERT OR IGNORE INTO uploads(id, hash, ext, url, owner_type, owner_id, kind) VALUES(?,?,?,?,?,?,?)`, id, hash, ".jpg", "/media/"+hash+".jpg", ownerType, ownerID, "image")
	if err != nil {
		t.Fatalf("seed owned upload: %v", err)
	}
}

func TestTelegramMediaIDOR_DMBypassBlocked(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 10, "Camp10", "P10", 1)
	testutil.SeedCampaign(t, 20, "Camp20", "P20", 1)
	// user2 has DM identity but no access to campaign 10's upload
	if _, err := db.DB.Exec("INSERT INTO telegram_identities(user_id, telegram_user_id, telegram_chat_id) VALUES(2,777,777)"); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	seedUploadOwned(t, 900, "campaign", 10)
	seedUploadOwned(t, 901, "campaign", 20)
	var called bool
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		called = true
		return 1, nil
	}
	telegram.SendMediaGroupImpl = func(chatID int64, caption string, items []telegram.MediaItem) ([]int64, error) {
		called = true
		return []int64{1}, nil
	}
	t.Cleanup(func() {
		telegram.SendMediaImpl = nil
		telegram.SendMediaGroupImpl = nil
	})
	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	}, 2, "user")
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{900},
		"target":     map[string]any{"kind": "dm"},
	})
	if w.Code != 403 {
		t.Fatalf("expected 403 for DM IDOR, got %d %s", w.Code, w.Body.String())
	}
	if called {
		t.Fatalf("should not have called Telegram on forbidden")
	}
}

func TestTelegramMediaIDOR_CampaignTargetForbidden(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 10, "Camp10", "P10", 1)
	testutil.SeedCampaign(t, 20, "Camp20", "P20", 2)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(20, 55555, 1)"); err != nil {
		t.Fatalf("insert cts: %v", err)
	}
	// upload owned by campaign 10, user2 is DM of campaign 20
	seedUploadOwned(t, 910, "campaign", 10)
	var called bool
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		called = true
		return 1, nil
	}
	t.Cleanup(func() { telegram.SendMediaImpl = nil })
	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	}, 2, "user")
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{910},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 20},
	})
	if w.Code != 403 {
		t.Fatalf("expected 403 for campaign IDOR, got %d %s", w.Code, w.Body.String())
	}
	if called {
		t.Fatalf("should not have called Telegram")
	}
	if testutil.CountRows(t, "telegram_media_posts") != 0 {
		t.Fatalf("expected no media posts")
	}
}

func TestTelegramMediaAdminBypass(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 10, "Camp10", "P10", 2)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(10, 55555, 1)"); err != nil {
		t.Fatalf("insert cts: %v", err)
	}
	seedUploadOwned(t, 920, "campaign", 10)
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	// need file on disk for loader
	seedUpload(t, 920, "image", ".jpg", "x.jpg", []byte("data920"))
	setupMediaLoader(t)
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		return 999, nil
	}
	t.Cleanup(func() { telegram.SendMediaImpl = nil })
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	})
	// user 1 is admin, should bypass ownership check and succeed (200 sent)
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{920},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 10},
	})
	testutil.AssertStatus(t, w, 200)
}

func TestTelegramMediaHappyPathCampaignOwner(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 2, "user2", "user")
	testutil.SeedCampaign(t, 10, "Camp10", "P10", 2)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled) VALUES(10, 55555, 1)"); err != nil {
		t.Fatalf("insert cts: %v", err)
	}
	dir := t.TempDir()
	SetMediaPath(dir)
	defer SetMediaPath("/app/media")
	seedUploadOwned(t, 930, "campaign", 10)
	// overwrite file for loader
	seedUpload(t, 930, "image", ".jpg", "y.jpg", []byte("data930"))
	setupMediaLoader(t)
	telegram.SendMediaImpl = func(chatID int64, kind, filename, caption string, data []byte) (int64, error) {
		return 1111, nil
	}
	t.Cleanup(func() { telegram.SendMediaImpl = nil })
	r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/media", PostTelegramMedia)
	}, 2, "user")
	w := testutil.PostJSON(t, r, "/api/telegram/media", map[string]any{
		"upload_ids": []int64{930},
		"target":     map[string]any{"kind": "campaign", "campaign_id": 10},
	})
	testutil.AssertStatus(t, w, 200)
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	results, _ := resp["results"].([]any)
	if len(results) == 0 || results[0].(map[string]any)["status"] != "sent" {
		t.Fatalf("expected sent, got %v", resp)
	}
}
