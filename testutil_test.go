package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers"
	"villum/middleware"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	testDB := "/tmp/villum_test.db"
	os.Remove(testDB)
	os.Remove(testDB + "-wal")
	os.Remove(testDB + "-shm")

	if err := db.Init(testDB); err != nil {
		fmt.Printf("Failed to init test db: %v\n", err)
		os.Exit(1)
	}
	db.Seed()

	gin.SetMode(gin.TestMode)
	// Ensure media path exists for setupRouter (mirrors app.go:initMedia).
	mediaPath := "/tmp/villum_test_media"
	_ = os.MkdirAll(mediaPath, 0755)
	handlers.SetMediaPath(mediaPath)
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	middleware.TokenDB = db.DB
	testRouter = buildRouter()

	middleware.StartCleanupTask()
	code := m.Run()
	db.Close()
	os.Remove(testDB)
	os.Remove(testDB + "-wal")
	os.Remove(testDB + "-shm")
	os.Exit(code)
}

func buildRouter() *gin.Engine {
	r, _ := setupRouter("/tmp/villum_test_media")
	return r
}

type testClient struct {
	sessionID string
	csrf      string
	token     string
}

func newTestClient() *testClient { return &testClient{} }

func (tc *testClient) req(method, path string, body any) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if tc.sessionID != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: tc.sessionID})
	}
	if tc.csrf != "" {
		req.Header.Set("X-CSRF-Token", tc.csrf)
	}
	// APITokenRequired exempts GET/HEAD/OPTIONS; for mutating requests attach Bearer if we have a token.
	if tc.token != "" && method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		// Token creation endpoint itself is exempt from APITokenRequired, but sending the header is harmless.
		req.Header.Set("Authorization", "Bearer "+tc.token)
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// Capture session cookie from Set-Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			tc.sessionID = c.Value
		}
	}
	return w
}

func (tc *testClient) get(path string, body any) *httptest.ResponseRecorder {
	return tc.req("GET", path, body)
}
func (tc *testClient) post(path string, body any) *httptest.ResponseRecorder {
	return tc.req("POST", path, body)
}
func (tc *testClient) put(path string, body any) *httptest.ResponseRecorder {
	return tc.req("PUT", path, body)
}
func (tc *testClient) del(path string, body any) *httptest.ResponseRecorder {
	return tc.req("DELETE", path, body)
}

const adminUser = "admin"
const adminPass = "testpass123"

func setupAdmin(t *testing.T, tc *testClient) {
	t.Helper()
	resp := tc.get("/api/check-setup", nil)
	var data map[string]bool
	json.Unmarshal(resp.Body.Bytes(), &data)
	if !data["setup"] {
		resp = tc.post("/api/login", map[string]any{
			"username": adminUser,
			"password": adminPass,
			"setup":    true,
		})
		if resp.Code != 200 {
			t.Fatalf("admin setup failed: %d - %s", resp.Code, resp.Body.String())
		}
	}
	// Always login to ensure fresh session
	resp = tc.post("/api/login", map[string]any{
		"username": adminUser,
		"password": adminPass,
	})
	if resp.Code != 200 {
		t.Fatalf("admin login failed: %d - %s", resp.Code, resp.Body.String())
	}
	// Fetch CSRF token
	resp = tc.get("/api/csrf-token", nil)
	if resp.Code != 200 {
		t.Fatalf("csrf token fetch failed: %d - %s", resp.Code, resp.Body.String())
	}
	var csrfData map[string]string
	json.Unmarshal(resp.Body.Bytes(), &csrfData)
	tc.csrf = csrfData["token"]

	// Create/fetch API token for mutating requests (token endpoint itself is exempt from APITokenRequired).
	// The token is stored in tc.token and automatically sent on POST/PUT/DELETE.
	resp = tc.post("/api/tokens", map[string]any{"name": "test-token"})
	if resp.Code == 201 {
		var tok map[string]any
		json.Unmarshal(resp.Body.Bytes(), &tok)
		if s, ok := tok["token"].(string); ok {
			tc.token = s
		}
	}
}

func login(t *testing.T, tc *testClient, username, password string) {
	t.Helper()
	resp := tc.post("/api/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.Code != 200 {
		t.Fatalf("login failed: %d - %s", resp.Code, resp.Body.String())
	}
	resp = tc.get("/api/csrf-token", nil)
	var csrfData map[string]string
	json.Unmarshal(resp.Body.Bytes(), &csrfData)
	tc.csrf = csrfData["token"]
	// Fetch API token for this user
	resp = tc.post("/api/tokens", map[string]any{"name": "test-token"})
	if resp.Code == 201 {
		var tok map[string]any
		json.Unmarshal(resp.Body.Bytes(), &tok)
		if s, ok := tok["token"].(string); ok {
			tc.token = s
		}
	}
}

func readJSON(w *httptest.ResponseRecorder, v any) {
	json.Unmarshal(w.Body.Bytes(), v)
}
