package testutil

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
	"villum/middleware"
)

func NewDB(t *testing.T) {
	t.Helper()
	path := fmt.Sprintf("/tmp/villum_%s_test.db", t.Name())
	os.Remove(path)
	os.Remove(path + "-wal")
	os.Remove(path + "-shm")
	if err := db.Init(path); err != nil {
		t.Fatalf("db init: %v", err)
	}
	db.Seed()
	gin.SetMode(gin.TestMode)
}

func CloseDB(t *testing.T) {
	t.Helper()
	if db.DB != nil {
		db.Close()
	}
	os.Remove(fmt.Sprintf("/tmp/villum_%s_test.db", t.Name()))
	os.Remove(fmt.Sprintf("/tmp/villum_%s_test.db-wal", t.Name()))
	os.Remove(fmt.Sprintf("/tmp/villum_%s_test.db-shm", t.Name()))
}

func NewRouter(routes func(*gin.RouterGroup)) *gin.Engine {
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "admin")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	})
	routes(auth)
	return r
}

func NewRouterWithUser(routes func(*gin.RouterGroup), userID int64, role string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("username", fmt.Sprintf("user%d", userID))
		c.Set("role", role)
		c.Set("session_id", "test-session")
		c.Next()
	})
	routes(auth)
	return r
}

func JSONBody(v any) *bytes.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func PostJSON(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, JSONBody(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func Get(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", path, nil)
	r.ServeHTTP(w, req)
	return w
}

func PutJSON(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", path, JSONBody(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func Delete(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", path, nil)
	r.ServeHTTP(w, req)
	return w
}

func ParseJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("json decode: %v (body: %s)", err, w.Body.String())
	}
}

func SeedUser(t *testing.T, id int64, username, role string) {
	t.Helper()
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO users(id, username, password, role) VALUES(?, ?, ?, ?)",
		id, username, "hash", role,
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func SeedCharacter(t *testing.T, id, userID int64, name, race, class string) {
	t.Helper()
	_, err := db.DB.Exec(
		`INSERT OR IGNORE INTO characters(id, user_id, name, race, class, level,
		 str, dex, con, int, wis, cha, hp_max, hp_current, ac, initiative, speed)
		 VALUES(?, ?, ?, ?, ?, 1, 10, 10, 10, 10, 10, 10, 12, 12, 10, 0, 30)`,
		id, userID, name, race, class,
	)
	if err != nil {
		t.Fatalf("seed character: %v", err)
	}
}

func SeedCampaign(t *testing.T, id int64, name, partyName string, userID int64) {
	t.Helper()
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO campaigns(id, name, party_name, user_id) VALUES(?, ?, ?, ?)",
		id, name, partyName, userID,
	)
	if err != nil {
		t.Fatalf("seed campaign: %v", err)
	}
	_, err = db.DB.Exec(
		"INSERT OR IGNORE INTO campaign_members(campaign_id, user_id, role) VALUES(?, ?, ?)",
		id, userID, "dm",
	)
	if err != nil {
		t.Fatalf("seed campaign member: %v", err)
	}
}

func GetCharID(t *testing.T, w *httptest.ResponseRecorder) int64 {
	t.Helper()
	var m map[string]any
	ParseJSON(t, w, &m)
	id, ok := m["id"].(float64)
	if !ok {
		t.Fatalf("response missing id: %s", w.Body.String())
	}
	return int64(id)
}

func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("expected status %d, got %d: %s", expected, w.Code, w.Body.String())
	}
}

func AssertField(t *testing.T, data map[string]any, field string, expected any) {
	t.Helper()
	val, ok := data[field]
	if !ok {
		t.Fatalf("response missing field %q: %+v", field, data)
	}
	if val != expected {
		t.Fatalf("field %q: expected %v, got %v", field, expected, val)
	}
}

func NewTestRequest(method, path string, body any) *http.Request {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	return req
}

func WithTestSession(req *http.Request) *http.Request {
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	req.Header.Set("X-CSRF-Token", "test-csrf")
	return req
}

func CountRows(t *testing.T, table string) int {
	t.Helper()
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	return count
}
