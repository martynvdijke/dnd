package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newAuthRequiredRouter wires the real AuthRequired in front of a tiny handler
// so the tests exercise the middleware, not a stub.
func newAuthRequiredRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"auth_method": c.GetString("auth_method")})
	})
	return r
}

// serveWithSession issues a GET /protected request carrying the session cookie.
func serveWithSession(r *gin.Engine, sessionID string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if sessionID != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	}
	r.ServeHTTP(w, req)
	return w
}

// useIsolatedStore swaps the package-global Store for a fresh MemoryStore so
// these tests never depend on (or leak into) another test's sessions.
func useIsolatedStore(t *testing.T) {
	t.Helper()
	old := Store
	Store = NewMemoryStore()
	t.Cleanup(func() { Store = old })
}

func TestAuthRequiredRejectsPasswordSessionWhenFallbackOff(t *testing.T) {
	useIsolatedStore(t)
	t.Setenv(PasswordFallbackEnvVar, "false")

	id := Store.Create(1, "alice", "user", "127.0.0.1")
	r := newAuthRequiredRouter()

	w := serveWithSession(r, id)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for password session with flag off, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body != `{"error":"session expired"}` {
		t.Fatalf("expected indistinguishable expired-session body, got %s", body)
	}
	if Store.Get(id) != nil {
		t.Fatal("rejected password session should be deleted from the store")
	}
}

func TestAuthRequiredAcceptsPasswordSessionWhenFallbackOn(t *testing.T) {
	useIsolatedStore(t)
	t.Setenv(PasswordFallbackEnvVar, "true")

	id := Store.Create(1, "alice", "user", "127.0.0.1")
	r := newAuthRequiredRouter()

	w := serveWithSession(r, id)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for password session with flag on, got %d: %s", w.Code, w.Body.String())
	}
	if Store.Get(id) == nil {
		t.Fatal("accepted password session should still exist")
	}
}

func TestAuthRequiredAcceptsOIDCSessionWhenFallbackOff(t *testing.T) {
	useIsolatedStore(t)
	t.Setenv(PasswordFallbackEnvVar, "false")

	id := Store.CreateWithAuth(1, "admin", "admin", "127.0.0.1", "oidc", "https://idp.example|x")
	r := newAuthRequiredRouter()

	w := serveWithSession(r, id)
	if w.Code != http.StatusOK {
		t.Fatalf("OIDC session must survive cutover, got %d: %s", w.Code, w.Body.String())
	}
	if Store.Get(id) == nil {
		t.Fatal("OIDC session should not be touched")
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["auth_method"] != "oidc" {
		t.Fatalf("expected auth_method oidc on context, got %q", got["auth_method"])
	}
}

func TestAuthRequiredNoCookieUnchanged(t *testing.T) {
	useIsolatedStore(t)
	t.Setenv(PasswordFallbackEnvVar, "false")

	r := newAuthRequiredRouter()
	w := serveWithSession(r, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without cookie, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body != `{"error":"not authenticated"}` {
		t.Fatalf("expected unchanged not-authenticated body, got %s", body)
	}
}
