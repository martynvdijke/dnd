package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

func newOIDCTestRouter() *gin.Engine {
	r := gin.New()
	r.GET("/api/auth/oidc/status", OIDCStatus)
	r.GET("/api/auth/oidc/login", OIDCLogin)
	r.GET("/api/auth/oidc/callback", OIDCCallback)
	r.GET("/api/auth/oidc/logout", OIDCLogout)
	return r
}

func clearOIDCEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"OIDC_ENABLED", "OIDC_ISSUER_URL", "OIDC_CLIENT_ID",
		"OIDC_CLIENT_SECRET", "OIDC_CLIENT_SECRET_FILE", "OIDC_REDIRECT_URL",
		"OIDC_SCOPES", "OIDC_LOGOUT_URL"} {
		t.Setenv(k, "")
	}
}

func TestLoadOIDCConfigDefaults(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	clearOIDCEnv(t)
	cfg := loadOIDCConfig()
	if cfg.Enabled {
		t.Fatal("expected disabled by default")
	}
	if len(cfg.Scopes) != 4 || cfg.Scopes[0] != "openid" {
		t.Fatalf("unexpected default scopes: %v", cfg.Scopes)
	}
}

func TestLoadOIDCConfigSecretFile(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	clearOIDCEnv(t)
	path := filepath.Join(t.TempDir(), "oidc-secret")
	if err := os.WriteFile(path, []byte("  file-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_ENABLED", "true")
	t.Setenv("OIDC_ISSUER_URL", "https://authelia.example")
	t.Setenv("OIDC_CLIENT_ID", "dnd")
	t.Setenv("OIDC_CLIENT_SECRET", "env-secret")
	t.Setenv("OIDC_CLIENT_SECRET_FILE", path)
	t.Setenv("OIDC_REDIRECT_URL", "https://dnd.example/api/auth/oidc/callback")
	cfg := loadOIDCConfig()
	if !cfg.valid() {
		t.Fatal("expected valid config")
	}
	if cfg.Secret != "file-secret" {
		t.Fatalf("file secret should win, got %q", cfg.Secret)
	}
}

func TestOIDCRoleForGroups(t *testing.T) {
	if got := oidcRoleForGroups([]string{"users", "admins"}); got != "admin" {
		t.Fatalf("expected admin, got %q", got)
	}
	if got := oidcRoleForGroups([]string{"users"}); got != "user" {
		t.Fatalf("expected user, got %q", got)
	}
	if got := oidcRoleForGroups(nil); got != "user" {
		t.Fatalf("expected user, got %q", got)
	}
}

func TestOIDCMigrationColumns(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	cols := map[string]bool{}
	rows, err := db.DB.Query(`PRAGMA table_info(users)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var cid int
	var name, ctype string
	var notnull, pk int
	var dflt any
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	if !cols["oidc_sub"] {
		t.Fatal("users.oidc_sub missing after migrate + safe alters")
	}
	for _, want := range []string{"auth_method", "oidc_sub"} {
		var n int
		if err := db.DB.QueryRow(
			`SELECT COUNT(*) FROM pragma_table_info('auth_sessions') WHERE name = ?`, want).Scan(&n); err != nil || n != 1 {
			t.Fatalf("auth_sessions.%s missing: %v", want, err)
		}
	}
}

func oidcTestClaims(email string) oidcClaims {
	return oidcClaims{Sub: "user-1", Email: email, EmailVerified: true, Name: "Test User", PreferredName: "testuser"}
}

func TestLinkOrProvisionOIDCUser_NewThenLink(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	ctx := context.Background()
	u, err := linkOrProvisionOIDCUser(ctx, "https://idp|x1", oidcTestClaims("sso@example.com"), []string{"users"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if u.username != "testuser" || u.role != "user" {
		t.Fatalf("unexpected provisioned user: %+v", u)
	}
	var pw, sub string
	if err := db.DB.QueryRow(`SELECT password, COALESCE(oidc_sub,'') FROM users WHERE id = ?`, u.id).Scan(&pw, &sub); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(pw, "oidc$") {
		t.Fatalf("OIDC password must be unusable, got %q", pw)
	}
	if sub != "https://idp|x1" {
		t.Fatalf("oidc_sub not stored, got %q", sub)
	}
	// Second login links by sub even if the email changed at the IdP.
	changed := oidcTestClaims("new@example.com")
	changed.Sub = "user-1"
	u2, err := linkOrProvisionOIDCUser(ctx, "https://idp|x1", changed, []string{"users"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if u2.id != u.id {
		t.Fatalf("expected link to user %d, got %d", u.id, u2.id)
	}
}

func TestLinkOrProvisionOIDCUser_ExistingEmailGrantsAdmin(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	if _, err := db.DB.Exec(`INSERT INTO users(username, password, display_name, role, email) VALUES('local','hash','Local','user','sso@example.com')`); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	u, err := linkOrProvisionOIDCUser(ctx, "https://idp|x2", oidcTestClaims("sso@example.com"), []string{"admins"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if u.username != "local" || u.role != "admin" {
		t.Fatalf("expected link + admin sync, got %+v", u)
	}
}

func TestLinkOrProvisionOIDCUser_NoDemoteWithoutGroupsClaim(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	if _, err := db.DB.Exec(`INSERT INTO users(username, password, display_name, role, email) VALUES('boss','hash','Boss','admin','boss@example.com')`); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	u, err := linkOrProvisionOIDCUser(ctx, "https://idp|x3", oidcTestClaims("boss@example.com"), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if u.role != "admin" {
		t.Fatalf("absent groups claim must not demote, got %q", u.role)
	}
}

func TestUniquifyUsername(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 7, "alice", "user")
	ctx := context.Background()
	if got := uniquifyUsername(ctx, "alice"); got != "alice-2" {
		t.Fatalf("expected alice-2, got %q", got)
	}
	if got := uniquifyUsername(ctx, "bob"); got != "bob" {
		t.Fatalf("expected bob, got %q", got)
	}
}

func TestOIDCRoutesDisabled(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	clearOIDCEnv(t)
	r := newOIDCTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/status", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d", w.Code)
	}
	var st struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil || st.Enabled {
		t.Fatalf("expected enabled=false, got %q / %v", w.Body.String(), err)
	}

	for _, path := range []string{"/api/auth/oidc/login", "/api/auth/oidc/callback"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404 when disabled, got %d", path, w.Code)
		}
	}
}

func TestOIDCLogoutClearsSession(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	clearOIDCEnv(t)
	r := newOIDCTestRouter()

	middleware.Store = middleware.NewMemoryStore()
	sid := middleware.Store.CreateWithAuth(1, "admin", "admin", "127.0.0.1", "oidc", "https://idp|x")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sid})
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 fallback when OIDC disabled, got %d", w.Code)
	}
	if middleware.Store.Get(sid) != nil {
		t.Fatal("expected session to be deleted")
	}
	cleared := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("expected session cookie to be cleared")
	}
}
