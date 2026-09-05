package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func newTokenDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS api_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, name TEXT NOT NULL DEFAULT '',
		token_hash TEXT NOT NULL UNIQUE, prefix TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		expires_at TEXT NOT NULL DEFAULT '', revoked_at TEXT NOT NULL DEFAULT '',
		last_used_at TEXT NOT NULL DEFAULT '')`); err != nil {
		t.Fatalf("create api_tokens: %v", err)
	}
	return db
}

func seedAPIToken(t *testing.T, db *sql.DB, userID int64, expiresAt, revokedAt string) (int64, string) {
	t.Helper()
	secret, err := GenerateTokenSecret()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	res, err := db.Exec(`INSERT INTO api_tokens (user_id, name, token_hash, prefix, expires_at, revoked_at) VALUES (?,?,?,?,?,?)`,
		userID, "test", HashTokenSecret(secret), TokenPrefix(secret), expiresAt, revokedAt)
	if err != nil {
		t.Fatalf("seed token: %v", err)
	}
	id, _ := res.LastInsertId()
	return id, secret
}

func tokenMiddlewareRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// stub session user_id set by AuthRequired normally
		if _, exists := c.Get("user_id"); !exists {
			c.Set("user_id", int64(1))
		}
		c.Next()
	})
	r.Use(APITokenRequired())
	r.POST("/api/mutate", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/api/read", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.PUT("/api/mutate2", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestGenerateTokenSecret(t *testing.T) {
	s, err := GenerateTokenSecret()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasPrefix(s, "vlt_") || len(s) != 68 {
		t.Fatalf("bad secret %q len %d", s, len(s))
	}
	s2, _ := GenerateTokenSecret()
	if s == s2 {
		t.Fatal("expected distinct secrets")
	}
}

func TestHashTokenSecret(t *testing.T) {
	h := HashTokenSecret("vlt_abc")
	if len(h) != 64 {
		t.Fatalf("expected 64 hex, got %d", len(h))
	}
	if h != HashTokenSecret("vlt_abc") {
		t.Fatal("deterministic")
	}
	if h == HashTokenSecret("vlt_different") {
		t.Fatal("different inputs should differ")
	}
}

func TestTokenPrefix(t *testing.T) {
	if TokenPrefix("vlt_abcdefghijklmnop") != "vlt_abcdef" {
		t.Fatalf("prefix mismatch: %q", TokenPrefix("vlt_abcdefghijklmnop"))
	}
	if TokenPrefix("short") != "short" {
		t.Fatal("short should return itself")
	}
	if TokenPrefix("") != "" {
		t.Fatal("empty")
	}
}

func TestAPITokenRequired_ValidAndInvalid(t *testing.T) {
	db := newTokenDB(t)
	TokenDB = db
	t.Cleanup(func() { TokenDB = nil })

	_, valid := seedAPIToken(t, db, 1, "", "")
	_, revoked := seedAPIToken(t, db, 1, "", time.Now().UTC().Format(time.RFC3339))
	_, expired := seedAPIToken(t, db, 1, time.Now().Add(-time.Hour).UTC().Format(time.RFC3339), "")
	_, otherUser := seedAPIToken(t, db, 2, "", "")

	r := tokenMiddlewareRouter()

	tests := []struct {
		name   string
		method string
		path   string
		secret string
		want   int
	}{
		{"GET exempt", "GET", "/api/read", "", 200},
		{"missing token", "POST", "/api/mutate", "", 401},
		{"valid token accepted", "POST", "/api/mutate", valid, 200},
		{"revoked rejected", "POST", "/api/mutate", revoked, 401},
		{"expired rejected", "POST", "/api/mutate", expired, 401},
		{"other user rejected", "POST", "/api/mutate", otherUser, 401},
		{"unknown rejected", "POST", "/api/mutate", "vlt_doesnotexist00000000000000000000000000000000000000000000000000", 401},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			if tc.secret != "" {
				req.Header.Set("Authorization", "Bearer "+tc.secret)
			}
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("want %d got %d body %s", tc.want, w.Code, w.Body.String())
			}
		})
	}
}

func TestAPITokenRequired_MalformedHeaders(t *testing.T) {
	db := newTokenDB(t)
	TokenDB = db
	t.Cleanup(func() { TokenDB = nil })
	r := tokenMiddlewareRouter()
	for _, hdr := range []string{"Basic abc", "Bearer", "Bearer ", "bearer abc", ""} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
		if hdr != "" {
			req.Header.Set("Authorization", hdr)
		}
		r.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Fatalf("header %q expected 401 got %d", hdr, w.Code)
		}
	}
}

func TestAPITokenRequired_NilDB(t *testing.T) {
	TokenDB = nil
	r := tokenMiddlewareRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer vlt_anything")
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401 with nil DB, got %d", w.Code)
	}
}

func TestAPITokenRequired_ExpiredAndRevokedPaths(t *testing.T) {
	db := newTokenDB(t)
	TokenDB = db
	t.Cleanup(func() { TokenDB = nil })
	// token with future expiry should pass
	_, future := seedAPIToken(t, db, 1, time.Now().Add(time.Hour).UTC().Format(time.RFC3339), "")
	r := tokenMiddlewareRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+future)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("future expiry: want 200 got %d %s", w.Code, w.Body.String())
	}
}

func TestAPITokenRequired_UpdatesLastUsed(t *testing.T) {
	db := newTokenDB(t)
	TokenDB = db
	t.Cleanup(func() { TokenDB = nil })
	id, secret := seedAPIToken(t, db, 1, "", "")
	r := tokenMiddlewareRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+secret)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("want 200 got %d", w.Code)
	}
	var lastUsed string
	if err := db.QueryRow(`SELECT last_used_at FROM api_tokens WHERE id=?`, id).Scan(&lastUsed); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if lastUsed == "" {
		t.Fatal("expected last_used_at set")
	}
}

func TestAPITokenRequired_PUTRequiresToken(t *testing.T) {
	db := newTokenDB(t)
	TokenDB = db
	t.Cleanup(func() { TokenDB = nil })
	_, valid := seedAPIToken(t, db, 1, "", "")
	r := tokenMiddlewareRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/mutate2", strings.NewReader(`{}`))
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("PUT without token should be 401 got %d", w.Code)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/mutate2", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+valid)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("PUT with valid token want 200 got %d", w.Code)
	}
}

func TestAPITokenRequired_MissingUserID(t *testing.T) {
	db := newTokenDB(t)
	TokenDB = db
	t.Cleanup(func() { TokenDB = nil })
	_, secret := seedAPIToken(t, db, 1, "", "")
	// router that does not set user_id at all
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APITokenRequired())
	r.POST("/api/mutate", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+secret)
	// explicitly not setting user_id
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("missing user_id want 401 got %d", w.Code)
	}
	if w.Code == http.StatusOK {
		t.Fatal("should not succeed")
	}
}
