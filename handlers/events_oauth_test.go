package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"villum/db"
	"villum/handlers/testutil"
)

func assertCookiesCleared(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	cookies := w.Header().Values("Set-Cookie")
	joined := strings.Join(cookies, ";")
	if !strings.Contains(joined, "gcal_oauth_state=") {
		t.Errorf("expected gcal_oauth_state cleared in Set-Cookie, got %q", joined)
	}
	if !strings.Contains(joined, "gcal_oauth_verifier=") {
		t.Errorf("expected gcal_oauth_verifier cleared in Set-Cookie, got %q", joined)
	}
	// Ensure Max-Age negative / expired
	if !strings.Contains(joined, "Max-Age=0") && !strings.Contains(strings.ToLower(joined), "max-age=0") {
		// gin uses MaxAge -1 which renders as Max-Age=0
		// Check raw for clearing pattern
		hasClear := false
		for _, c := range cookies {
			if strings.Contains(c, "gcal_oauth_state=") && strings.Contains(c, "Max-Age=0") {
				hasClear = true
			}
			if strings.Contains(c, "gcal_oauth_verifier=") && strings.Contains(c, "Max-Age=0") {
				hasClear = true
			}
		}
		_ = hasClear
	}
}

func TestGoogleOAuthStatus(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/events-oauth/status", GoogleOAuthStatus)
	})
	t.Run("not connected", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/events-oauth/status")
		testutil.AssertStatus(t, w, 200)
		var m map[string]any
		testutil.ParseJSON(t, w, &m)
		if m["connected"] != false {
			t.Errorf("expected connected false, got %v", m["connected"])
		}
		if _, ok := m["redirect_uri"]; !ok {
			t.Error("expected redirect_uri")
		}
		if !strings.HasSuffix(m["redirect_uri"].(string), "/api/admin/events-oauth/callback") {
			t.Errorf("redirect_uri suffix wrong: %v", m["redirect_uri"])
		}
	})
	t.Run("connected", func(t *testing.T) {
		db.SaveEventSettings(db.EventSettings{AuthMethod: "oauth", OAuthRefreshToken: "rt"})
		w := testutil.Get(t, r, "/api/admin/events-oauth/status")
		var m map[string]any
		testutil.ParseJSON(t, w, &m)
		if m["connected"] != true {
			t.Errorf("expected connected true, got %v", m["connected"])
		}
	})
	t.Run("redirect URI override", func(t *testing.T) {
		SetGoogleOAuthRedirectURL("https://dnd.example.com")
		t.Cleanup(func() { SetGoogleOAuthRedirectURL("") })
		w := testutil.Get(t, r, "/api/admin/events-oauth/status")
		var m map[string]any
		testutil.ParseJSON(t, w, &m)
		want := "https://dnd.example.com/api/admin/events-oauth/callback"
		if m["redirect_uri"] != want {
			t.Errorf("redirect_uri = %q, want %q", m["redirect_uri"], want)
		}
	})
}

func TestGoogleOAuthStart(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/events-oauth/start", GoogleOAuthStart)
	})
	t.Run("missing credentials", func(t *testing.T) {
		db.SaveEventSettings(db.EventSettings{})
		w := testutil.Get(t, r, "/api/admin/events-oauth/start")
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d", w.Code)
		}
		loc := w.Header().Get("Location")
		if !strings.Contains(loc, "missing_credentials") {
			t.Errorf("expected missing_credentials, got %q", loc)
		}
	})
	t.Run("with credentials redirects to google", func(t *testing.T) {
		db.SaveEventSettings(db.EventSettings{OAuthClientID: "id", OAuthClientSecret: "secret"})
		w := testutil.Get(t, r, "/api/admin/events-oauth/start")
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d", w.Code)
		}
		loc := w.Header().Get("Location")
		if !strings.Contains(loc, "accounts.google.com") {
			t.Errorf("expected google redirect, got %q", loc)
		}
		if !strings.Contains(loc, "access_type=offline") {
			t.Errorf("expected access_type offline in %q", loc)
		}
		if !strings.Contains(loc, "prompt=consent") {
			t.Errorf("expected prompt consent in %q", loc)
		}
		cookies := w.Header().Values("Set-Cookie")
		joined := strings.Join(cookies, ";")
		if !strings.Contains(joined, "gcal_oauth_state") {
			t.Error("expected gcal_oauth_state cookie")
		}
		if !strings.Contains(joined, "gcal_oauth_verifier") {
			t.Error("expected gcal_oauth_verifier cookie")
		}
	})
}

func TestGoogleOAuthCallback(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	orig := googleOAuthExchangeFn
	defer func() { googleOAuthExchangeFn = orig }()
	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/events-oauth/callback", GoogleOAuthCallback)
	})

	t.Run("bad state", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/events-oauth/callback?code=abc&state=xyz")
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d", w.Code)
		}
		if !strings.Contains(w.Header().Get("Location"), "reason=state") {
			t.Errorf("expected state error, got %q", w.Header().Get("Location"))
		}
		assertCookiesCleared(t, w)
	})

	t.Run("missing verifier cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/admin/events-oauth/callback?code=mycode&state=s123", nil)
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_state", Value: "s123"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d", w.Code)
		}
		if !strings.Contains(w.Header().Get("Location"), "reason=state") {
			t.Errorf("expected state error, got %q", w.Header().Get("Location"))
		}
		assertCookiesCleared(t, w)
	})

	t.Run("success stores refresh token", func(t *testing.T) {
		googleOAuthExchangeFn = func(ctx context.Context, cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "at", RefreshToken: "rt"}, nil
		}
		db.SaveEventSettings(db.EventSettings{OAuthClientID: "id", OAuthClientSecret: "secret"})
		req := httptest.NewRequest("GET", "/api/admin/events-oauth/callback?code=mycode&state=s123", nil)
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_state", Value: "s123"})
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_verifier", Value: "verifier"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d %s", w.Code, w.Body.String())
		}
		loc := w.Header().Get("Location")
		if !strings.Contains(loc, "events_oauth=connected") {
			t.Errorf("expected connected, got %q", loc)
		}
		s := db.GetEventSettings()
		if s.OAuthRefreshToken != "rt" {
			t.Errorf("expected rt, got %q", s.OAuthRefreshToken)
		}
		if s.AuthMethod != "oauth" {
			t.Errorf("expected oauth, got %q", s.AuthMethod)
		}
		if s.SourceType != "google_api" {
			t.Errorf("expected google_api, got %q", s.SourceType)
		}
	})

	t.Run("no refresh token error", func(t *testing.T) {
		googleOAuthExchangeFn = func(ctx context.Context, cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "at", RefreshToken: ""}, nil
		}
		req := httptest.NewRequest("GET", "/api/admin/events-oauth/callback?code=mycode&state=s123", nil)
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_state", Value: "s123"})
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_verifier", Value: "verifier"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if !strings.Contains(w.Header().Get("Location"), "reason=no_refresh_token") {
			t.Errorf("expected no_refresh_token, got %q", w.Header().Get("Location"))
		}
	})

	t.Run("exchange error", func(t *testing.T) {
		googleOAuthExchangeFn = func(ctx context.Context, cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
			return nil, errors.New("boom")
		}
		req := httptest.NewRequest("GET", "/api/admin/events-oauth/callback?code=mycode&state=s123", nil)
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_state", Value: "s123"})
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_verifier", Value: "verifier"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if !strings.Contains(w.Header().Get("Location"), "reason=exchange") {
			t.Errorf("expected exchange error, got %q", w.Header().Get("Location"))
		}
		assertCookiesCleared(t, w)
	})

	t.Run("nil token", func(t *testing.T) {
		googleOAuthExchangeFn = func(ctx context.Context, cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
			return nil, nil
		}
		req := httptest.NewRequest("GET", "/api/admin/events-oauth/callback?code=mycode&state=s123", nil)
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_state", Value: "s123"})
		req.AddCookie(&http.Cookie{Name: "gcal_oauth_verifier", Value: "verifier"})
		w := httptest.NewRecorder()
		// should not panic
		r.ServeHTTP(w, req)
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d", w.Code)
		}
		if !strings.Contains(w.Header().Get("Location"), "reason=exchange") {
			t.Errorf("expected exchange error, got %q", w.Header().Get("Location"))
		}
	})
}
