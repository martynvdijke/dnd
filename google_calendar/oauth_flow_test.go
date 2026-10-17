package googlecalendar

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func TestOAuthConfig(t *testing.T) {
	cfg := OAuthConfig("id", "secret", "https://example.com/cb")
	if cfg.ClientID != "id" || cfg.ClientSecret != "secret" || cfg.RedirectURL != "https://example.com/cb" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Endpoint != google.Endpoint {
		t.Errorf("expected google endpoint")
	}
	found := false
	for _, s := range cfg.Scopes {
		if s == calendar.CalendarReadonlyScope {
			found = true
		}
	}
	if !found {
		t.Errorf("expected calendar readonly scope, got %v", cfg.Scopes)
	}
}

func TestAuthCodeURL(t *testing.T) {
	cfg := OAuthConfig("id", "secret", "https://example.com/cb")
	verifier := oauth2.GenerateVerifier()
	u := AuthCodeURL(cfg, "mystate", verifier)
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	q := parsed.Query()
	if q.Get("state") != "mystate" {
		t.Errorf("state got %q", q.Get("state"))
	}
	if q.Get("access_type") != "offline" {
		t.Errorf("access_type got %q", q.Get("access_type"))
	}
	if q.Get("prompt") != "consent" {
		t.Errorf("prompt got %q", q.Get("prompt"))
	}
	if q.Get("code_challenge") == "" {
		t.Error("expected code_challenge")
	}
	if q.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method got %q", q.Get("code_challenge_method"))
	}
}

func TestExchangeCode(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"at","refresh_token":"rt","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()
	cfg := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Endpoint:     oauth2.Endpoint{AuthURL: srv.URL + "/auth", TokenURL: srv.URL + "/token"},
		RedirectURL:  "https://example.com/cb",
		Scopes:       []string{calendar.CalendarReadonlyScope},
	}
	verifier := oauth2.GenerateVerifier()
	tok, err := ExchangeCode(context.Background(), cfg, "mycode", verifier)
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if tok.AccessToken != "at" {
		t.Errorf("access token got %q", tok.AccessToken)
	}
	if tok.RefreshToken != "rt" {
		t.Errorf("refresh token got %q", tok.RefreshToken)
	}
	if !strings.Contains(body, "code=mycode") {
		t.Errorf("body missing code, got %q", body)
	}
	if !strings.Contains(body, "code_verifier=") {
		t.Errorf("body missing code_verifier, got %q", body)
	}
}
