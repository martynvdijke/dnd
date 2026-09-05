package middleware

import (
	"database/sql"
	"testing"
	"time"

	"villum/models"

	_ "modernc.org/sqlite"
)

func TestMemoryStoreCreateGetRoundTrip(t *testing.T) {
	s := NewMemoryStore()
	id := s.Create(42, "alice", "user", "1.2.3.4")
	if id == "" {
		t.Fatal("expected non-empty session id")
	}
	sess := s.Get(id)
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
	if sess.UserID != 42 || sess.Username != "alice" || sess.Role != "user" || sess.IP != "1.2.3.4" {
		t.Fatalf("unexpected session fields: %+v", sess)
	}
}

func TestMemoryStoreGetMissing(t *testing.T) {
	s := NewMemoryStore()
	if got := s.Get("nonexistent"); got != nil {
		t.Fatalf("expected nil for missing, got %+v", got)
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	s := NewMemoryStore()
	id := s.Create(1, "bob", "admin", "127.0.0.1")
	s.Delete(id)
	if got := s.Get(id); got != nil {
		t.Fatal("expected nil after delete")
	}
	// delete non-existent should not panic
	s.Delete("does-not-exist")
}

func TestMemoryStoreExpiredGetReturnsNil(t *testing.T) {
	s := NewMemoryStore()
	id := s.Create(1, "u", "user", "127.0.0.1")
	// expire via direct map mutation (deterministic, no sleep)
	s.mu.Lock()
	s.sessions[id].ExpiresAt = time.Now().Add(-time.Hour)
	s.mu.Unlock()
	if got := s.Get(id); got != nil {
		t.Fatal("expected nil for expired session")
	}
}

func TestMemoryStoreCleanup(t *testing.T) {
	s := NewMemoryStore()
	idActive := s.Create(1, "a", "user", "127.0.0.1")
	idExpired := s.Create(2, "b", "user", "127.0.0.1")
	s.mu.Lock()
	s.sessions[idExpired].ExpiresAt = time.Now().Add(-time.Hour)
	s.mu.Unlock()
	s.Cleanup()
	if got := s.Get(idActive); got == nil {
		t.Fatal("active session should survive cleanup")
	}
	// expired should be removed from map directly
	s.mu.RLock()
	_, exists := s.sessions[idExpired]
	s.mu.RUnlock()
	if exists {
		t.Fatal("expired session should be removed by cleanup")
	}
}

func TestMemoryStoreCleanupRemovesExpiredFromGet(t *testing.T) {
	s := NewMemoryStore()
	// ensure cache-style expiry also works after Cleanup
	id := s.Create(1, "u", "user", "127.0.0.1")
	s.mu.Lock()
	s.sessions[id] = &models.AuthSession{
		ID: id, UserID: 1, Username: "u", Role: "user", IP: "127.0.0.1",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	s.mu.Unlock()
	s.Cleanup()
	if s.Get(id) != nil {
		t.Fatal("expired session should be nil after cleanup")
	}
}

func TestNewMemoryStoreEmpty(t *testing.T) {
	s := NewMemoryStore()
	if s == nil || s.sessions == nil {
		t.Fatal("expected initialized store")
	}
	if len(s.sessions) != 0 {
		t.Fatalf("expected empty store, got %d", len(s.sessions))
	}
}

// ─── DBSessionStore ───

func newDBStoreForTest(t *testing.T) *DBSessionStore {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if _, err := sqlDB.Exec(`CREATE TABLE IF NOT EXISTS auth_sessions (
		id TEXT PRIMARY KEY, user_id INTEGER NOT NULL, username TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT 'user', ip TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL, expires_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create auth_sessions: %v", err)
	}
	return NewDBSessionStore(sqlDB)
}

func TestDBSessionStoreCreateGetRoundTrip(t *testing.T) {
	s := newDBStoreForTest(t)
	id := s.Create(99, "carol", "dm", "10.0.0.1")
	if id == "" {
		t.Fatal("empty id")
	}
	sess := s.Get(id)
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
	if sess.UserID != 99 || sess.Username != "carol" {
		t.Fatalf("wrong fields: %+v", sess)
	}
}

func TestDBSessionStoreGetMissing(t *testing.T) {
	s := newDBStoreForTest(t)
	if got := s.Get("nope"); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestDBSessionStoreDelete(t *testing.T) {
	s := newDBStoreForTest(t)
	id := s.Create(1, "u", "user", "127.0.0.1")
	s.Delete(id)
	// clear cache to force DB read
	s.cache.Delete(id)
	if got := s.Get(id); got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestDBSessionStoreGetFromDBAfterCacheEvict(t *testing.T) {
	s := newDBStoreForTest(t)
	id := s.Create(1, "u", "user", "127.0.0.1")
	// evict cache, should hit DB path
	s.cache.Delete(id)
	sess := s.Get(id)
	if sess == nil {
		t.Fatal("expected session from DB after cache evict")
	}
	if sess.Username != "u" {
		t.Fatalf("wrong username: %s", sess.Username)
	}
}

func TestDBSessionStoreExpiredViaSQL(t *testing.T) {
	s := newDBStoreForTest(t)
	id := s.Create(1, "u", "user", "127.0.0.1")
	// expire in DB and cache
	expired := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if _, err := s.db.Exec(`UPDATE auth_sessions SET expires_at = ? WHERE id = ?`, expired, id); err != nil {
		t.Fatalf("update expires_at: %v", err)
	}
	// also expire cache entry deterministically
	if v, ok := s.cache.Load(id); ok {
		v.(*models.AuthSession).ExpiresAt = time.Now().Add(-time.Hour)
	}
	if got := s.Get(id); got != nil {
		t.Fatal("expected nil for expired (cache path)")
	}
	// create another, evict cache, expire only in DB
	id2 := s.Create(2, "v", "user", "127.0.0.1")
	s.cache.Delete(id2)
	if _, err := s.db.Exec(`UPDATE auth_sessions SET expires_at = ? WHERE id = ?`, expired, id2); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got := s.Get(id2); got != nil {
		t.Fatal("expected nil for DB-expired")
	}
}

func TestDBSessionStoreCleanup(t *testing.T) {
	s := newDBStoreForTest(t)
	idActive := s.Create(1, "a", "user", "127.0.0.1")
	idExpired := s.Create(2, "b", "user", "127.0.0.1")
	expired := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	if _, err := s.db.Exec(`UPDATE auth_sessions SET expires_at = ? WHERE id = ?`, expired, idExpired); err != nil {
		t.Fatalf("update: %v", err)
	}
	// expire cache entry for that id
	if v, ok := s.cache.Load(idExpired); ok {
		v.(*models.AuthSession).ExpiresAt = time.Now().Add(-2 * time.Hour)
	}
	s.Cleanup()
	// active survives (via cache or DB)
	s.cache.Delete(idActive)
	if got := s.Get(idActive); got == nil {
		t.Fatal("active should survive cleanup")
	}
	// expired removed from DB
	s.cache.Delete(idExpired)
	if got := s.Get(idExpired); got != nil {
		t.Fatal("expired should be removed by cleanup")
	}
	var cnt int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM auth_sessions WHERE id = ?`, idExpired).Scan(&cnt); err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 0 {
		t.Fatal("expired row should be deleted from DB")
	}
}

func TestDBSessionStoreCorruptTimestamp(t *testing.T) {
	s := newDBStoreForTest(t)
	id := s.Create(1, "u", "user", "127.0.0.1")
	s.cache.Delete(id)
	if _, err := s.db.Exec(`UPDATE auth_sessions SET created_at = 'not-a-time' WHERE id = ?`, id); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got := s.Get(id); got != nil {
		t.Fatal("expected nil for corrupt created_at")
	}
}

func TestGenerateSessionID(t *testing.T) {
	a := generateSessionID()
	b := generateSessionID()
	if len(a) != 64 || len(b) != 64 {
		t.Fatalf("expected 64 hex chars, got %d %d", len(a), len(b))
	}
	if a == b {
		t.Fatal("expected distinct ids")
	}
}

func TestNewDBSessionStore(t *testing.T) {
	s := newDBStoreForTest(t)
	if s == nil || s.db == nil {
		t.Fatal("expected db store")
	}
	if got := s.Get("missing"); got != nil {
		t.Fatal("expected nil for missing")
	}
}
