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
	mu             sync.RWMutex
	sessions       map[string]Session
	loginAt        map[int64]time.Time
	repo           SessionRepository
	appReauthAfter time.Duration
}

const (
	SessionKindPreAuth = "pre_auth"
	SessionKindApp     = "app"
)

func NewTokenStore() *TokenStore {
	return NewTokenStoreWithRepository(nil)
}

func NewTokenStoreWithRepository(repo SessionRepository) *TokenStore {
	return &TokenStore{sessions: make(map[string]Session), loginAt: make(map[int64]time.Time), repo: repo, appReauthAfter: 60 * 24 * time.Hour}
}

// SetAppReauthAfter controls the inactivity window after which a bound WeChat
// account must use the existing-account login flow rather than silent login.
func (s *TokenStore) SetAppReauthAfter(value time.Duration) {
	if s == nil || value <= 0 {
		return
	}
	s.mu.Lock()
	s.appReauthAfter = value
	s.mu.Unlock()
}

func (s *TokenStore) RequiresAppReauthentication(userID int64) bool {
	if s == nil || userID <= 0 {
		return true
	}
	s.mu.RLock()
	lastLoginAt := s.loginAt[userID]
	window := s.appReauthAfter
	s.mu.RUnlock()
	if lastLoginAt.IsZero() && s.repo != nil {
		if repo, ok := s.repo.(LoginActivityRepository); ok {
			if stored, found, err := repo.LastAppLoginAt(context.Background(), userID); err == nil && found {
				lastLoginAt = stored
				s.mu.Lock()
				s.loginAt[userID] = stored
				s.mu.Unlock()
			}
		}
	}
	if lastLoginAt.IsZero() {
		return true
	}
	return time.Since(lastLoginAt) >= window
}

func (s *TokenStore) Issue(userID int64) (Session, error) {
	return s.IssueApp(userID)
}

func (s *TokenStore) IssuePreAuth(userID int64) (Session, error) {
	return s.issue(userID, SessionKindPreAuth, 2*time.Hour)
}

func (s *TokenStore) IssueApp(userID int64) (Session, error) {
	session, err := s.issue(userID, SessionKindApp, 24*time.Hour)
	if err != nil {
		return Session{}, err
	}
	s.recordAppLogin(userID, time.Now())
	return session, nil
}

func (s *TokenStore) recordAppLogin(userID int64, occurredAt time.Time) {
	if s == nil || userID <= 0 {
		return
	}
	s.mu.Lock()
	s.loginAt[userID] = occurredAt
	s.mu.Unlock()
	if repo, ok := s.repo.(LoginActivityRepository); ok {
		_ = repo.SaveAppLoginAt(context.Background(), userID, occurredAt)
	}
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
			s.mu.Lock()
			delete(s.sessions, token)
			s.mu.Unlock()
			return Session{}, err
		}
	}

	return session, nil
}

func (s *TokenStore) Verify(token string) (Session, bool) {
	if s.repo != nil {
		session, ok, err := s.repo.FindSessionByTokenHash(context.Background(), hashToken(token))
		if err != nil || !ok || time.Now().After(session.ExpiresAt) {
			s.mu.Lock()
			delete(s.sessions, token)
			s.mu.Unlock()
			return Session{}, false
		}
		session.Token = token
		s.mu.Lock()
		s.sessions[token] = session
		s.mu.Unlock()
		return session, true
	}
	s.mu.RLock()
	session, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || time.Now().After(session.ExpiresAt) {
		return Session{}, false
	}
	return session, true
}

func (s *TokenStore) ActiveSessionCount(kind string) int {
	count, _ := s.ActiveSessionCountStrict(kind)
	return count
}

func (s *TokenStore) ActiveSessionCountStrict(kind string) (int, error) {
	if s == nil {
		return 0, nil
	}
	now := time.Now()
	if counter, ok := s.repo.(ActiveSessionCounter); ok {
		return counter.CountActiveSessions(context.Background(), kind, now)
	}
	activeUsers := make(map[int64]struct{})
	expired := make([]string, 0)

	s.mu.RLock()
	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			expired = append(expired, token)
			continue
		}
		if kind == "" || session.Kind == kind {
			activeUsers[session.UserID] = struct{}{}
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

	return len(activeUsers), nil
}

// RevokeUserSessions invalidates every local and persisted session for a user.
func (s *TokenStore) RevokeUserSessions(userID int64) error {
	if s == nil || userID <= 0 {
		return nil
	}
	if s.repo != nil {
		if err := s.repo.RevokeUserSessions(context.Background(), userID); err != nil {
			return err
		}
	}
	s.mu.Lock()
	for token, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, token)
		}
	}
	s.mu.Unlock()
	return nil
}

func (s *TokenStore) RevokeSession(token string) error {
	if s == nil || token == "" {
		return nil
	}
	if revoker, ok := s.repo.(SessionTokenRevoker); ok {
		if err := revoker.RevokeSessionByTokenHash(context.Background(), hashToken(token)); err != nil {
			return err
		}
	}
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
	return nil
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
