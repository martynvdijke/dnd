package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/middleware"
	"villum/models"

	_ "modernc.org/sqlite"
)

func setupResetTestDB(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	// Create isolated memory DB
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	// assign to global db.DB for handlers to use
	oldDB := db.DB
	db.DB = sqlDB
	t.Cleanup(func() { db.DB = oldDB })

	sqlDB.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE, password TEXT NOT NULL DEFAULT '', display_name TEXT NOT NULL DEFAULT '', role TEXT NOT NULL DEFAULT 'user', email TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL DEFAULT '')`)
	sqlDB.Exec(`CREATE TABLE email_settings (id INTEGER PRIMARY KEY, smtp_host TEXT, smtp_port INTEGER, username TEXT, password TEXT, from_addr TEXT, enabled INTEGER)`)
	sqlDB.Exec(`INSERT INTO email_settings (id, smtp_host, smtp_port, username, password, from_addr, enabled) VALUES (1,'smtp.example.com',587,'user','pass','from@example.com',1)`)
	sqlDB.Exec(`CREATE TABLE password_reset_tokens(id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, token_hash TEXT UNIQUE NOT NULL, expires_at INTEGER NOT NULL, used_at INTEGER, created_at INTEGER NOT NULL DEFAULT 0)`)
	sqlDB.Exec(`CREATE TABLE auth_sessions (id TEXT PRIMARY KEY, user_id INTEGER, username TEXT, role TEXT, ip TEXT, created_at TEXT, expires_at TEXT, auth_method TEXT, oidc_sub TEXT)`)

	r := gin.New()
	r.POST("/api/forgot-password", RequestPasswordReset)
	r.POST("/api/reset-password", ResetPassword)
	r.POST("/api/login", HandleLogin)
	return r
}

func createResetUser(t *testing.T, username, email, password string) int64 {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	res, err := db.DB.Exec(`INSERT INTO users (username, password, display_name, role, email, created_at) VALUES (?,?,?,?,?,?)`, username, string(hash), username, "user", email, time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestPasswordResetFlow(t *testing.T) {
	r := setupResetTestDB(t)

	var capturedLink string
	orig := SendEmailFunc
	SendEmailFunc = func(s *models.EmailSettings, to, subject, body string) error {
		capturedLink = body
		return nil
	}
	defer func() { SendEmailFunc = orig }()

	username := "reset_user1"
	email := "reset1@example.com"
	createResetUser(t, username, email, "oldpassword123")

	t.Run("enumeration safe nonexistent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/forgot-password", bytes.NewReader([]byte(`{"email":"nope@example.com"}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["ok"] != true {
			t.Fatalf("expected ok true %v", resp)
		}
	})

	t.Run("successful request", func(t *testing.T) {
		capturedLink = ""
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/forgot-password", bytes.NewReader([]byte(`{"email":"`+email+`"}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("code %d %s", w.Code, w.Body.String())
		}
		var count int
		db.DB.QueryRow(`SELECT COUNT(*) FROM password_reset_tokens WHERE user_id=(SELECT id FROM users WHERE username=?)`, username).Scan(&count)
		if count == 0 {
			t.Fatal("expected token inserted")
		}
		if capturedLink == "" {
			t.Fatal("expected email sent")
		}
	})

	knownToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	knownHash := func(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }(knownToken)
	db.DB.Exec(`DELETE FROM password_reset_tokens`)
	db.DB.Exec(`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, created_at) VALUES ((SELECT id FROM users WHERE username=?), ?, ?, ?)`, username, knownHash, time.Now().Add(1*time.Hour).Unix(), time.Now().Unix())

	t.Run("weak password rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"token": knownToken, "new_password": "short"})
		req := httptest.NewRequest("POST", "/api/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400 got %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("successful reset then login", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"token": knownToken, "new_password": "newpassword123"})
		req := httptest.NewRequest("POST", "/api/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("reset failed %d %s", w.Code, w.Body.String())
		}
		var pw string
		db.DB.QueryRow(`SELECT password FROM users WHERE username=?`, username).Scan(&pw)
		if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte("newpassword123")); err != nil {
			t.Fatalf("new password hash mismatch: %v", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte("oldpassword123")); err == nil {
			t.Fatal("old password should not work")
		}
	})

	t.Run("reused token rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"token": knownToken, "new_password": "anotherpass123"})
		req := httptest.NewRequest("POST", "/api/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			t.Fatalf("reused token should fail got %d", w.Code)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expToken := "expiredtoken1234567890expiredtoken1234567890expired1234567890abcd"
		expHash := func(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }(expToken)
		db.DB.Exec(`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, created_at) VALUES ((SELECT id FROM users WHERE username=?), ?, ?, ?)`, username, expHash, time.Now().Add(-2*time.Hour).Unix(), time.Now().Unix())
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"token": expToken, "new_password": "validpass123"})
		req := httptest.NewRequest("POST", "/api/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			t.Fatalf("expired token should fail")
		}
	})
}

// TestPasswordResetRetiredByDefault proves that when the password break-glass
// flag is off, the reset surface is refused exactly like POST /api/login: no
// token is minted, no token is consumed, and no password is written. It must
// not call t.Parallel because it uses t.Setenv (which requires a clean
// environment and is illegal in parallel tests).
func TestPasswordResetRetiredByDefault(t *testing.T) {
	r := setupResetTestDB(t)
	// The handlers package runs with break-glass on (see main_test.go TestMain);
	// this test owns the retired behaviour explicitly.
	t.Setenv(middleware.PasswordFallbackEnvVar, "false")

	var capturedLink string
	orig := SendEmailFunc
	SendEmailFunc = func(s *models.EmailSettings, to, subject, body string) error {
		capturedLink = body
		return nil
	}
	defer func() { SendEmailFunc = orig }()

	username := "reset_retired"
	email := "retired@example.com"
	createResetUser(t, username, email, "oldpassword123")

	t.Run("forgot-password refused and mints no token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/forgot-password", bytes.NewReader([]byte(`{"email":"`+email+`"}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 403 {
			t.Fatalf("expected 403 got %d %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"] != "password login is disabled" {
			t.Fatalf("expected retired message, got %v", resp["error"])
		}
		var count int
		db.DB.QueryRow(`SELECT COUNT(*) FROM password_reset_tokens WHERE user_id=(SELECT id FROM users WHERE username=?)`, username).Scan(&count)
		if count != 0 {
			t.Fatalf("expected no token minted, got %d", count)
		}
		if capturedLink != "" {
			t.Fatalf("expected no email sent, got %q", capturedLink)
		}
	})

	knownToken := "retiredtoken1234567890retiredtoken1234567890retired1234567890abcd"
	knownHash := func(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }(knownToken)
	db.DB.Exec(`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, created_at) VALUES ((SELECT id FROM users WHERE username=?), ?, ?, ?)`, username, knownHash, time.Now().Add(1*time.Hour).Unix(), time.Now().Unix())
	var hashBefore string
	db.DB.QueryRow(`SELECT password FROM users WHERE username=?`, username).Scan(&hashBefore)

	t.Run("reset-password refused and leaves password and token untouched", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"token": knownToken, "new_password": "newpassword123"})
		req := httptest.NewRequest("POST", "/api/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 403 {
			t.Fatalf("expected 403 got %d %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"] != "password login is disabled" {
			t.Fatalf("expected retired message, got %v", resp["error"])
		}
		var hashAfter string
		db.DB.QueryRow(`SELECT password FROM users WHERE username=?`, username).Scan(&hashAfter)
		if hashAfter != hashBefore {
			t.Fatalf("password hash changed while password login is retired")
		}
		var usedAt sql.NullInt64
		db.DB.QueryRow(`SELECT used_at FROM password_reset_tokens WHERE token_hash=?`, knownHash).Scan(&usedAt)
		if usedAt.Valid {
			t.Fatalf("token was consumed while password login is retired")
		}
	})

	t.Run("token works once break-glass is set", func(t *testing.T) {
		t.Setenv(middleware.PasswordFallbackEnvVar, "true")
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{"token": knownToken, "new_password": "newpassword123"})
		req := httptest.NewRequest("POST", "/api/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200 once fallback enabled, got %d %s", w.Code, w.Body.String())
		}
		var pw string
		db.DB.QueryRow(`SELECT password FROM users WHERE username=?`, username).Scan(&pw)
		if err := bcrypt.CompareHashAndPassword([]byte(pw), []byte("newpassword123")); err != nil {
			t.Fatalf("new password hash mismatch: %v", err)
		}
	})
}
