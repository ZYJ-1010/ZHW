package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
	Kind      string
}

type TokenStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
	repo     SessionRepository
}

const (
	SessionKindPreAuth = "pre_auth"
	SessionKindApp     = "app"
)

func NewTokenStore() *TokenStore {
	return NewTokenStoreWithRepository(nil)
}

func NewTokenStoreWithRepository(repo SessionRepository) *TokenStore {
	return &TokenStore{sessions: make(map[string]Session), repo: repo}
}

func (s *TokenStore) Issue(userID int64) (Session, error) {
	return s.IssueApp(userID)
}

func (s *TokenStore) IssuePreAuth(userID int64) (Session, error) {
	return s.issue(userID, SessionKindPreAuth, 2*time.Hour)
}

func (s *TokenStore) IssueApp(userID int64) (Session, error) {
	return s.issue(userID, SessionKindApp, 24*time.Hour)
}

func (s *TokenStore) issue(userID int64, kind string, ttl time.Duration) (Session, error) {
	token, err := randomToken(24)
	if err != nil {
		return Session{}, err
	}
	session := Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
		Kind:      kind,
	}

	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()
	if s.repo != nil {
		if err := s.repo.SaveSession(context.Background(), session); err != nil {
			return Session{}, err
		}
	}

	return session, nil
}

func (s *TokenStore) Verify(token string) (Session, bool) {
	s.mu.RLock()
	session, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || time.Now().After(session.ExpiresAt) {
		if s.repo == nil {
			return Session{}, false
		}
		session, ok, err := s.repo.FindSessionByTokenHash(context.Background(), hashToken(token))
		if err != nil || !ok {
			return Session{}, false
		}
		session.Token = token
		s.mu.Lock()
		s.sessions[token] = session
		s.mu.Unlock()
		return session, true
	}
	return session, true
}

func (s *TokenStore) ActiveSessionCount(kind string) int {
	if s == nil {
		return 0
	}
	now := time.Now()
	count := 0
	expired := make([]string, 0)

	s.mu.RLock()
	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			expired = append(expired, token)
			continue
		}
		if kind == "" || session.Kind == kind {
			count++
		}
	}
	s.mu.RUnlock()

	if len(expired) > 0 {
		s.mu.Lock()
		for _, token := range expired {
			if session, ok := s.sessions[token]; ok && now.After(session.ExpiresAt) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}

	return count
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
