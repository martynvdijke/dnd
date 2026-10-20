package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

// newLoginTestRouter wires the real DB-backed session store so auth_sessions
// rows are observable, restoring the previous store on cleanup.
func newLoginTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	old := middleware.Store
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	t.Cleanup(func() { middleware.Store = old })

	r := gin.New()
	r.POST("/api/login", HandleLogin)
	r.GET("/api/auth/oidc/status", OIDCStatus)
	return r
}

func seedPasswordUser(t *testing.T, username, password, role string) int64 {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	u, err := db.Client.User.Create().
		SetUsername(username).
		SetPassword(string(hash)).
		SetDisplayName(username).
		SetRole(role).
		Save(context.Background())
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u.ID
}

func postLogin(t *testing.T, r *gin.Engine, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func countAuthSessions(t *testing.T) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM auth_sessions`).Scan(&n); err != nil {
		t.Fatalf("count auth_sessions: %v", err)
	}
	return n
}

func countAdmins(t *testing.T) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&n); err != nil {
		t.Fatalf("count admins: %v", err)
	}
	return n
}

func hasSessionCookie(w *httptest.ResponseRecorder) bool {
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.Value != "" && c.MaxAge >= 0 {
			return true
		}
	}
	return false
}

func TestLoginDisabledByDefault(t *testing.T) {
	t.Setenv(middleware.PasswordFallbackEnvVar, "")
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	seedPasswordUser(t, "alice", "correct-horse-battery", "user")
	r := newLoginTestRouter(t)

	t.Run("correct password rejected with 403 and no session", func(t *testing.T) {
		before := countAuthSessions(t)
		w := postLogin(t, r, map[string]any{"username": "alice", "password": "correct-horse-battery"})
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
		if hasSessionCookie(w) {
			t.Fatal("password login must not set a session cookie while disabled")
		}
		if got := countAuthSessions(t); got != before {
			t.Fatalf("expected no new auth_sessions row, before=%d after=%d", before, got)
		}
	})

	t.Run("setup rejected with 403 and creates no admin", func(t *testing.T) {
		if countAdmins(t) != 0 {
			t.Fatal("precondition: no admin should exist")
		}
		before := countAuthSessions(t)
		w := postLogin(t, r, map[string]any{
			"setup":    true,
			"username": "admin",
			"password": "longenoughpw",
		})
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
		if got := countAdmins(t); got != 0 {
			t.Fatalf("expected no admin created, got %d", got)
		}
		if hasSessionCookie(w) {
			t.Fatal("setup must not set a session cookie while password auth is disabled")
		}
		if got := countAuthSessions(t); got != before {
			t.Fatalf("expected no new auth_sessions row, before=%d after=%d", before, got)
		}
	})
}

func TestLoginEnabledWithBreakGlass(t *testing.T) {
	t.Setenv(middleware.PasswordFallbackEnvVar, "true")
	testutil.NewDB(t)
	defer testutil.CloseDB(t)

	seedPasswordUser(t, "bob", "correct-horse-battery", "user")
	r := newLoginTestRouter(t)

	t.Run("correct password returns 200 and a password session", func(t *testing.T) {
		w := postLogin(t, r, map[string]any{"username": "bob", "password": "correct-horse-battery"})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !hasSessionCookie(w) {
			t.Fatal("expected a session cookie")
		}
		var method string
		if err := db.DB.QueryRow(`SELECT auth_method FROM auth_sessions ORDER BY created_at DESC LIMIT 1`).Scan(&method); err != nil {
			t.Fatalf("read auth_sessions: %v", err)
		}
		if method != "password" {
			t.Fatalf("expected auth_method=password, got %q", method)
		}
	})

	t.Run("wrong password still returns 401 invalid credentials", func(t *testing.T) {
		w := postLogin(t, r, map[string]any{"username": "bob", "password": "wrong-password"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
		var resp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if resp.Error != "invalid credentials" {
			t.Fatalf("expected invalid credentials, got %q", resp.Error)
		}
	})

	t.Run("setup creates the admin", func(t *testing.T) {
		if countAdmins(t) != 0 {
			t.Fatal("precondition: no admin should exist")
		}
		w := postLogin(t, r, map[string]any{
			"setup":    true,
			"username": "newadmin",
			"password": "longenoughpw",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if got := countAdmins(t); got != 1 {
			t.Fatalf("expected 1 admin created, got %d", got)
		}
	})
}

func TestOIDCStatusReportsPasswordPolicy(t *testing.T) {
	t.Setenv(middleware.PasswordFallbackEnvVar, "")
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	r := newLoginTestRouter(t)

	readPasswordEnabled := func(t *testing.T) bool {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/auth/oidc/status", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var st struct {
			Enabled         bool `json:"enabled"`
			PasswordEnabled bool `json:"password_enabled"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil {
			t.Fatalf("decode status: %v", err)
		}
		return st.PasswordEnabled
	}

	if readPasswordEnabled(t) {
		t.Fatal("expected password_enabled=false with break-glass unset")
	}
	t.Setenv(middleware.PasswordFallbackEnvVar, "true")
	if !readPasswordEnabled(t) {
		t.Fatal("expected password_enabled=true with break-glass set")
	}
}
