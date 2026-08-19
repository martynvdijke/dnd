package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
	"villum/middleware"
)

// seedToken inserts an api_tokens row directly and returns its id and the
// plaintext secret (which the test then uses to authenticate requests).
func seedToken(t *testing.T, userID int64, name, expiresAt, revokedAt string) (int64, string) {
	t.Helper()
	secret, err := middleware.GenerateTokenSecret()
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}
	res, err := db.DB.Exec(
		`INSERT INTO api_tokens (user_id, name, token_hash, prefix, expires_at, revoked_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, name, middleware.HashTokenSecret(secret), middleware.TokenPrefix(secret), expiresAt, revokedAt,
	)
	if err != nil {
		t.Fatalf("seed token: %v", err)
	}
	id, _ := res.LastInsertId()
	return id, secret
}

// postWithAuth issues a POST with an optional bearer token.
func postWithAuth(t *testing.T, r *gin.Engine, path, secret string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	r.ServeHTTP(w, req)
	return w
}

func tokenRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		RegisterTokenRoutes(auth)
	})
}

// ─── Token lifecycle (task 2.3 flows) ───

func TestCreateToken(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := tokenRouter()

	t.Run("creates token and returns secret once", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/tokens", map[string]any{"name": "ci"})
		testutil.AssertStatus(t, w, 201)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		token, _ := res["token"].(string)
		if !strings.HasPrefix(token, "vlt_") {
			t.Fatalf("expected token to start with vlt_, got %q", token)
		}
		if len(token) != 68 {
			t.Fatalf("expected 68-char token, got %d", len(token))
		}
		prefix, _ := res["prefix"].(string)
		if prefix != token[:10] {
			t.Fatalf("expected prefix %q, got %q", token[:10], prefix)
		}
		if res["id"].(float64) <= 0 {
			t.Fatalf("expected positive id, got %v", res["id"])
		}
		// Only the hash is persisted, never the secret.
		var storedHash string
		err := db.DB.QueryRow(`SELECT token_hash FROM api_tokens WHERE id = ?`, int64(res["id"].(float64))).Scan(&storedHash)
		if err != nil {
			t.Fatalf("read stored hash: %v", err)
		}
		if storedHash != middleware.HashTokenSecret(token) {
			t.Fatal("stored hash does not match token secret")
		}
		if strings.Contains(storedHash, token) {
			t.Fatal("stored hash must not contain the plaintext secret")
		}
	})

	t.Run("creates token with expiry", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/tokens", map[string]any{"name": "short", "expires_in_days": 30})
		testutil.AssertStatus(t, w, 201)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		exp, _ := res["expires_at"].(string)
		if exp == "" {
			t.Fatal("expected expires_at to be set")
		}
		if _, err := time.Parse(time.RFC3339, exp); err != nil {
			t.Fatalf("expires_at not RFC3339: %v", err)
		}
	})

	t.Run("rejects name longer than 100 chars", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/tokens", map[string]any{"name": strings.Repeat("x", 101)})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("rejects invalid body", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/tokens", "not-json")
		testutil.AssertStatus(t, w, 400)
	})
}

func TestListTokens(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := tokenRouter()

	// Create two tokens via the API so we know their secrets.
	secrets := make([]string, 0, 2)
	for _, name := range []string{"one", "two"} {
		w := testutil.PostJSON(t, r, "/api/tokens", map[string]any{"name": name})
		testutil.AssertStatus(t, w, 201)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		secrets = append(secrets, res["token"].(string))
	}

	w := testutil.Get(t, r, "/api/tokens")
	testutil.AssertStatus(t, w, 200)
	var tokens []map[string]any
	testutil.ParseJSON(t, w, &tokens)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %s", len(tokens), w.Body.String())
	}
	body := w.Body.String()
	for _, s := range secrets {
		if strings.Contains(body, s) {
			t.Fatal("list response must never contain a token secret")
		}
	}
	for _, tok := range tokens {
		if _, has := tok["token"]; has {
			t.Fatal("list response must not include a token field")
		}
		if _, has := tok["token_hash"]; has {
			t.Fatal("list response must not include token_hash")
		}
		if _, has := tok["prefix"]; !has {
			t.Fatal("expected prefix field in listing")
		}
	}
}

func TestRevokeToken(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "other", "user")

	r := tokenRouter()

	t.Run("revokes own token", func(t *testing.T) {
		id, _ := seedToken(t, 1, "ci", "", "")
		w := testutil.Delete(t, r, fmt.Sprintf("/api/tokens/%d", id))
		testutil.AssertStatus(t, w, 200)
		var revokedAt string
		err := db.DB.QueryRow(`SELECT revoked_at FROM api_tokens WHERE id = ?`, id).Scan(&revokedAt)
		if err != nil {
			t.Fatalf("read revoked_at: %v", err)
		}
		if revokedAt == "" {
			t.Fatal("expected revoked_at to be set after revoke")
		}
		// Revoking again is a 404 (already revoked).
		w2 := testutil.Delete(t, r, fmt.Sprintf("/api/tokens/%d", id))
		testutil.AssertStatus(t, w2, 404)
	})

	t.Run("cannot revoke another user's token", func(t *testing.T) {
		id, _ := seedToken(t, 2, "other-token", "", "")
		w := testutil.Delete(t, r, fmt.Sprintf("/api/tokens/%d", id))
		testutil.AssertStatus(t, w, 404)
		var revokedAt string
		err := db.DB.QueryRow(`SELECT revoked_at FROM api_tokens WHERE id = ?`, id).Scan(&revokedAt)
		if err != nil {
			t.Fatalf("read revoked_at: %v", err)
		}
		if revokedAt != "" {
			t.Fatal("other user's token must not be revoked")
		}
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/tokens/abc")
		testutil.AssertStatus(t, w, 400)
	})
}

func TestRotateToken(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "other", "user")

	r := tokenRouter()

	t.Run("rotates own token", func(t *testing.T) {
		id, oldSecret := seedToken(t, 1, "ci", "", "")
		w := testutil.PostJSON(t, r, fmt.Sprintf("/api/tokens/%d/rotate", id), nil)
		testutil.AssertStatus(t, w, 200)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		newSecret, _ := res["token"].(string)
		if newSecret == "" || newSecret == oldSecret {
			t.Fatal("expected a new secret after rotation")
		}
		if res["prefix"] != newSecret[:10] {
			t.Fatalf("expected prefix %q, got %v", newSecret[:10], res["prefix"])
		}
		// Old secret must no longer validate.
		var storedHash string
		err := db.DB.QueryRow(`SELECT token_hash FROM api_tokens WHERE id = ?`, id).Scan(&storedHash)
		if err != nil {
			t.Fatalf("read stored hash: %v", err)
		}
		if storedHash == middleware.HashTokenSecret(oldSecret) {
			t.Fatal("old secret must be invalid after rotation")
		}
		if storedHash != middleware.HashTokenSecret(newSecret) {
			t.Fatal("new secret must be stored hashed")
		}
	})

	t.Run("cannot rotate another user's token", func(t *testing.T) {
		id, _ := seedToken(t, 2, "other-token", "", "")
		w := testutil.PostJSON(t, r, fmt.Sprintf("/api/tokens/%d/rotate", id), nil)
		testutil.AssertStatus(t, w, 404)
	})

	t.Run("cannot rotate revoked token", func(t *testing.T) {
		id, _ := seedToken(t, 1, "ci", "", time.Now().UTC().Format(time.RFC3339))
		w := testutil.PostJSON(t, r, fmt.Sprintf("/api/tokens/%d/rotate", id), nil)
		testutil.AssertStatus(t, w, 404)
	})
}

// ─── APITokenRequired middleware (task 2.1) ───

// apiTokenRouter builds a router with the stub session auth followed by
// APITokenRequired and a single mutation endpoint.
func apiTokenRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.Use(middleware.APITokenRequired())
		auth.POST("/mutate", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		auth.GET("/read", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	})
}

func TestAPITokenRequired(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "other", "user")
	middleware.TokenDB = db.DB

	_, validSecret := seedToken(t, 1, "ci", "", "")
	_, revokedSecret := seedToken(t, 1, "revoked", "", time.Now().UTC().Format(time.RFC3339))
	_, expiredSecret := seedToken(t, 1, "expired", time.Now().Add(-time.Hour).UTC().Format(time.RFC3339), "")
	_, otherUserSecret := seedToken(t, 2, "other", "", "")

	r := apiTokenRouter()

	t.Run("GET is exempt", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/read")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("missing token rejected", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", "")
		testutil.AssertStatus(t, w, 401)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		if res["error"] != "API token required" {
			t.Fatalf("expected 'API token required', got %v", res["error"])
		}
	})

	t.Run("malformed authorization header rejected", func(t *testing.T) {
		for _, hdr := range []string{"Basic abc", "Bearer", "Bearer ", "bearer abc"} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
			req.Header.Set("Authorization", hdr)
			r.ServeHTTP(w, req)
			testutil.AssertStatus(t, w, 401)
		}
	})

	t.Run("unknown token rejected", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", "vlt_doesnotexist")
		testutil.AssertStatus(t, w, 401)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		if res["error"] != "invalid API token" {
			t.Fatalf("expected 'invalid API token', got %v", res["error"])
		}
	})

	t.Run("revoked token rejected", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", revokedSecret)
		testutil.AssertStatus(t, w, 401)
	})

	t.Run("expired token rejected", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", expiredSecret)
		testutil.AssertStatus(t, w, 401)
	})

	t.Run("another user's token rejected without revealing ownership", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", otherUserSecret)
		testutil.AssertStatus(t, w, 401)
		var res map[string]any
		testutil.ParseJSON(t, w, &res)
		// Same generic message as an unknown token: no ownership leak.
		if res["error"] != "invalid API token" {
			t.Fatalf("expected generic 'invalid API token', got %v", res["error"])
		}
	})

	t.Run("valid token accepted", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", validSecret)
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("valid token updates last_used_at", func(t *testing.T) {
		id, secret := seedToken(t, 1, "usage", "", "")
		w := postWithAuth(t, r, "/api/mutate", secret)
		testutil.AssertStatus(t, w, 200)
		var lastUsed string
		err := db.DB.QueryRow(`SELECT last_used_at FROM api_tokens WHERE id = ?`, id).Scan(&lastUsed)
		if err != nil {
			t.Fatalf("read last_used_at: %v", err)
		}
		if lastUsed == "" {
			t.Fatal("expected last_used_at to be updated after use")
		}
	})
}

// TestTokenWithoutSessionRejected covers the spec scenario: a token presented
// without a user session must be rejected and change nothing.
func TestTokenWithoutSessionRejected(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	middleware.Store = middleware.NewDBSessionStore(db.DB)
	middleware.TokenDB = db.DB

	_, secret := seedToken(t, 1, "ci", "", "")

	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	auth := r.Group("/api")
	auth.Use(middleware.AuthRequired(), middleware.APITokenRequired())
	auth.POST("/mutate", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	t.Run("token without session rejected", func(t *testing.T) {
		w := postWithAuth(t, r, "/api/mutate", secret)
		testutil.AssertStatus(t, w, 401)
	})

	t.Run("session without token rejected", func(t *testing.T) {
		sessID := middleware.Store.Create(1, "admin", "admin", "127.0.0.1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
		req.AddCookie(&http.Cookie{Name: "session", Value: sessID})
		r.ServeHTTP(w, req)
		testutil.AssertStatus(t, w, 401)
	})

	t.Run("session and token accepted", func(t *testing.T) {
		sessID := middleware.Store.Create(1, "admin", "admin", "127.0.0.1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/mutate", strings.NewReader(`{}`))
		req.AddCookie(&http.Cookie{Name: "session", Value: sessID})
		req.Header.Set("Authorization", "Bearer "+secret)
		r.ServeHTTP(w, req)
		testutil.AssertStatus(t, w, 200)
	})
}

// ─── Admin boundaries (task 3.2) ───

func TestAdminBoundaryRequiresToken(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedUser(t, 2, "user", "user")
	middleware.TokenDB = db.DB

	_, adminSecret := seedToken(t, 1, "admin-token", "", "")
	_, userSecret := seedToken(t, 2, "user-token", "", "")

	t.Run("non-admin with valid token is forbidden", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.Use(middleware.AdminRequired(), middleware.APITokenRequired())
			auth.POST("/mutate", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
		}, 2, "user")
		w := postWithAuth(t, r, "/api/mutate", userSecret)
		testutil.AssertStatus(t, w, 403)
	})

	t.Run("admin without token is rejected", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.Use(middleware.AdminRequired(), middleware.APITokenRequired())
			auth.POST("/mutate", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
		}, 1, "admin")
		w := postWithAuth(t, r, "/api/mutate", "")
		testutil.AssertStatus(t, w, 401)
	})

	t.Run("admin with token is accepted", func(t *testing.T) {
		r := testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
			auth.Use(middleware.AdminRequired(), middleware.APITokenRequired())
			auth.POST("/mutate", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
		}, 1, "admin")
		w := postWithAuth(t, r, "/api/mutate", adminSecret)
		testutil.AssertStatus(t, w, 200)
	})
}

// ─── Secrets never logged (task 3.2) ───

func TestTokenSecretNeverLogged(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	middleware.TokenDB = db.DB

	_, secret := seedToken(t, 1, "ci", "", "")

	appLog := middleware.InitAppLogger(100, slog.LevelDebug)
	defer func() { middleware.AppLog = nil }()

	r := gin.New()
	r.Use(middleware.RequestLogger())
	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "admin")
		c.Set("role", "admin")
		c.Set("session_id", "test-session")
		c.Next()
	}, middleware.APITokenRequired())
	auth.POST("/mutate", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := postWithAuth(t, r, "/api/mutate", secret)
	testutil.AssertStatus(t, w, 200)

	entries := appLog.Buffer().Query("", "", time.Time{}, 0)
	if len(entries) == 0 {
		t.Fatal("expected request log entries")
	}
	for _, e := range entries {
		if strings.Contains(e.Message, secret) {
			t.Fatalf("token secret found in log message: %q", e.Message)
		}
		for k, v := range e.Attributes {
			if strings.Contains(k, secret) || strings.Contains(fmt.Sprintf("%v", v), secret) {
				t.Fatalf("token secret found in log attribute %s=%v", k, v)
			}
		}
	}
}
