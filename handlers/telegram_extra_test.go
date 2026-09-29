package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"villum/db"
	"villum/handlers/testutil"
	"villum/telegram"
)

func mockTelegramForHandlers(t *testing.T, h func(w http.ResponseWriter, r *http.Request)) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(srv.Close)
	os.Setenv("TELEGRAM_API_BASE", srv.URL)
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Cleanup(func() {
		os.Unsetenv("TELEGRAM_API_BASE")
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_BOT_USERNAME")
		telegram.ResetBotUsernameCache()
	})
}
func jsonOK2(v any) []byte {
	b, _ := json.Marshal(map[string]any{"ok": true, "result": v})
	return b
}

func TestCreateLinkCodeURLViaGetMe(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	mockTelegramForHandlers(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getMe") {
			w.Write(jsonOK2(map[string]any{"id": 1, "is_bot": true, "first_name": "b", "username": "TestBot123"}))
			return
		}
		w.Write(jsonOK2(map[string]any{}))
	})
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram/link-code", CreateTelegramLinkCode)
	})
	w := testutil.PostJSON(t, r, "/api/telegram/link-code", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	urlStr, _ := resp["url"].(string)
	if !strings.Contains(urlStr, "TestBot123") {
		t.Fatalf("url should contain bot username, got %q", urlStr)
	}
	if !strings.Contains(urlStr, "?start=") {
		t.Fatalf("url missing start param %q", urlStr)
	}
}

func TestAdminTelegramSettingsStatusShape(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	os.Setenv("BASE_URL", "https://example.com")
	t.Cleanup(func() { os.Unsetenv("BASE_URL") })
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	_ = telegram.SetBotToken("tok1234567890_clear_test_token")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/telegram-settings", GetTelegramSettings)
		auth.POST("/telegram-settings", SaveTelegramSettings)
	})
	w := testutil.Get(t, r, "/api/telegram-settings")
	if w.Code != 200 {
		t.Fatalf("get %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if _, ok := resp["status"]; !ok {
		t.Fatalf("missing status object: %v", resp)
	}
	status, _ := resp["status"].(map[string]any)
	if _, ok := status["mode"]; !ok {
		t.Fatalf("status missing mode")
	}
	if _, ok := status["webhook_url"]; !ok {
		t.Fatalf("status missing webhook_url")
	}
	if _, ok := status["update_offset"]; !ok {
		t.Fatalf("status missing update_offset")
	}
	// POST clear_token
	w2 := testutil.PostJSON(t, r, "/api/telegram-settings", map[string]any{"clear_token": true})
	if w2.Code != 200 {
		t.Fatalf("post clear %d %s", w2.Code, w2.Body.String())
	}
	var resp2 map[string]any
	testutil.ParseJSON(t, w2, &resp2)
	if has, _ := resp2["has_token"].(bool); has {
		t.Fatalf("token should be cleared")
	}
}

func TestAdminTelegramTestNoBody(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/telegram-test", TestTelegram)
	})
	w := testutil.PostJSON(t, r, "/api/telegram-test", map[string]any{})
	if w.Code != 200 {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if resp["sent"] != false {
		t.Fatalf("expected sent false, got %v", resp)
	}
	if resp["error"] != "not linked" {
		t.Fatalf("expected not linked error, got %v", resp)
	}
	if _, err := db.DB.Exec("INSERT INTO telegram_identities(user_id,telegram_user_id,telegram_chat_id) VALUES(1,111,222)"); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	mockTelegramForHandlers(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonOK2(map[string]any{"message_id": 1}))
	})
	w2 := testutil.PostJSON(t, r, "/api/telegram-test", map[string]any{})
	if w2.Code != 200 {
		t.Fatalf("code %d %s", w2.Code, w2.Body.String())
	}
	var resp2 map[string]any
	testutil.ParseJSON(t, w2, &resp2)
	if resp2["sent"] != true {
		t.Fatalf("expected sent true, got %v %s", resp2, w2.Body.String())
	}
}

func TestCampaignTelegramBoundAt(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "C1", "P1", 1)
	if _, err := db.DB.Exec("INSERT INTO campaign_telegram_settings(campaign_id,chat_id,is_enabled,bound_at) VALUES(10,123,1,datetime('now'))"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/campaigns/:id/telegram", GetCampaignTelegram)
	})
	w := testutil.Get(t, r, "/api/campaigns/10/telegram")
	if w.Code != 200 {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if _, ok := resp["bound_at"]; !ok {
		t.Fatalf("missing bound_at, got %v", resp)
	}
}

func TestSetPrefsIncludesOK(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	if _, err := db.DB.Exec("INSERT INTO telegram_identities(user_id,telegram_user_id,telegram_chat_id) VALUES(1,111,222)"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/telegram/prefs", SetTelegramPrefs)
	})
	w := testutil.PutJSON(t, r, "/api/telegram/prefs", map[string]any{"dm_enabled": true})
	if w.Code != 200 {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	testutil.ParseJSON(t, w, &resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok true, got %v", resp)
	}
}
