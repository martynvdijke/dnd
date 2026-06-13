package middleware

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"villum/models"
)

// SessionStore defines the interface for session storage backends.
type SessionStore interface {
	Create(userID int64, username, role, ip string) string
	Get(sessionID string) *models.AuthSession
	Delete(sessionID string)
	Cleanup()
}

// MemoryStore is an in-memory session store backed by a map + RWMutex.
// It remains available as a fallback and for testing.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*models.AuthSession
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]*models.AuthSession)}
}

func (s *MemoryStore) Create(userID int64, username, role, ip string) string {
	id := generateSessionID()
	s.mu.Lock()
	s.sessions[id] = &models.AuthSession{
		ID:        id,
		UserID:    userID,
		Username:  username,
		Role:      role,
		IP:        ip,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mu.Unlock()
	return id
}

func (s *MemoryStore) Get(sessionID string) *models.AuthSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil
	}
	return sess
}

func (s *MemoryStore) Delete(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

func (s *MemoryStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// DBSessionStore is a SQLite-backed session store with an in-memory read cache.
type DBSessionStore struct {
	db    *sql.DB
	cache sync.Map // map[string]*models.AuthSession
}

func NewDBSessionStore(db *sql.DB) *DBSessionStore {
	return &DBSessionStore{db: db}
}

func (s *DBSessionStore) Create(userID int64, username, role, ip string) string {
	id := generateSessionID()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	sess := &models.AuthSession{
		ID:        id,
		UserID:    userID,
		Username:  username,
		Role:      role,
		IP:        ip,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	_, err := s.db.Exec(
		`INSERT INTO auth_sessions (id, user_id, username, role, ip, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, userID, username, role, ip, now.UTC().Format(time.RFC3339), expiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		log.Printf("DBSessionStore.Create: %v", err)
		// Fall back to returning the ID even if DB write fails
	}

	s.cache.Store(id, sess)
	return id
}

func (s *DBSessionStore) Get(sessionID string) *models.AuthSession {
	// Check cache first
	if cached, ok := s.cache.Load(sessionID); ok {
		sess := cached.(*models.AuthSession)
		if time.Now().After(sess.ExpiresAt) {
			s.cache.Delete(sessionID)
			return nil
		}
		return sess
	}

	// Cache miss — read from DB
	var createdAt, expiresAt string
	sess := &models.AuthSession{ID: sessionID}
	err := s.db.QueryRow(
		`SELECT user_id, username, role, ip, created_at, expires_at FROM auth_sessions WHERE id = ?`,
		sessionID,
	).Scan(&sess.UserID, &sess.Username, &sess.Role, &sess.IP, &createdAt, &expiresAt)

	if err != nil {
		return nil // not found or DB error
	}

	sess.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	sess.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)

	if time.Now().After(sess.ExpiresAt) {
		s.Delete(sessionID)
		return nil
	}

	s.cache.Store(sessionID, sess)
	return sess
}

func (s *DBSessionStore) Delete(sessionID string) {
	s.cache.Delete(sessionID)
	_, err := s.db.Exec(`DELETE FROM auth_sessions WHERE id = ?`, sessionID)
	if err != nil {
		log.Printf("DBSessionStore.Delete: %v", err)
	}
}

func (s *DBSessionStore) Cleanup() {
	// Remove expired from DB
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`DELETE FROM auth_sessions WHERE expires_at < ?`, now)
	if err != nil {
		log.Printf("DBSessionStore.Cleanup: %v", err)
	}

	// Remove expired from cache
	s.cache.Range(func(key, value interface{}) bool {
		sess := value.(*models.AuthSession)
		if time.Now().After(sess.ExpiresAt) {
			s.cache.Delete(key)
		}
		return true
	})
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
