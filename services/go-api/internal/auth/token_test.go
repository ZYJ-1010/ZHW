package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTokenStorePersistsAndReloadsSession(t *testing.T) {
	repo := newRecordingSessionRepository()
	issuer := NewTokenStoreWithRepository(repo)
	session, err := issuer.IssuePreAuth(42)
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.sessions) != 1 {
		t.Fatalf("expected persisted session, got %d", len(repo.sessions))
	}
	if _, ok := repo.sessions[session.Token]; ok {
		t.Fatal("expected repository to store token hash, not raw token")
	}

	reloader := NewTokenStoreWithRepository(repo)
	reloaded, ok := reloader.Verify(session.Token)
	if !ok {
		t.Fatal("expected session reload from repository")
	}
	if reloaded.UserID != 42 || reloaded.Kind != SessionKindPreAuth || reloaded.Token != session.Token {
		t.Fatalf("unexpected reloaded session: %+v", reloaded)
	}
}

func TestTokenStoreActiveSessionCountOnlyCountsValidAppSessions(t *testing.T) {
	store := NewTokenStore()
	appSession, err := store.IssueApp(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssuePreAuth(2); err != nil {
		t.Fatal(err)
	}
	expiredSession, err := store.IssueApp(3)
	if err != nil {
		t.Fatal(err)
	}

	store.mu.Lock()
	expiredSession.ExpiresAt = time.Now().Add(-time.Minute)
	store.sessions[expiredSession.Token] = expiredSession
	store.mu.Unlock()

	if got := store.ActiveSessionCount(SessionKindApp); got != 1 {
		t.Fatalf("expected one active app session, got %d", got)
	}
	if got := store.ActiveSessionCount(""); got != 2 {
		t.Fatalf("expected app and preauth sessions, got %d", got)
	}
	if _, ok := store.sessions[appSession.Token]; !ok {
		t.Fatal("active app session should remain cached")
	}
	if _, ok := store.sessions[expiredSession.Token]; ok {
		t.Fatal("expired app session should be removed")
	}
}

func TestTokenStoreActiveSessionCountUsesDistinctUsers(t *testing.T) {
	store := NewTokenStore()
	if _, err := store.IssueApp(1); err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssueApp(1); err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssueApp(2); err != nil {
		t.Fatal(err)
	}

	if got := store.ActiveSessionCount(SessionKindApp); got != 2 {
		t.Fatalf("expected two online users, got %d", got)
	}
}

type recordingSessionRepository struct {
	sessions map[string]Session
	saveErr  error
	countErr error
}

func newRecordingSessionRepository() *recordingSessionRepository {
	return &recordingSessionRepository{sessions: make(map[string]Session)}
}

func (r *recordingSessionRepository) SaveSession(_ context.Context, session Session) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	session.Token = hashToken(session.Token)
	r.sessions[session.Token] = session
	return nil
}

func TestTokenStoreDoesNotRetainSessionWhenPersistenceFails(t *testing.T) {
	repo := newRecordingSessionRepository()
	repo.saveErr = errors.New("database unavailable")
	store := NewTokenStoreWithRepository(repo)
	if _, err := store.IssueApp(42); err == nil {
		t.Fatal("expected session persistence failure")
	}
	store.mu.RLock()
	count := len(store.sessions)
	store.mu.RUnlock()
	if count != 0 {
		t.Fatalf("failed session must not remain locally valid, got %d cached session(s)", count)
	}
}

func TestPersistedSessionRevocationInvalidatesOtherInstanceCache(t *testing.T) {
	repo := newRecordingSessionRepository()
	issuer := NewTokenStoreWithRepository(repo)
	session, err := issuer.IssueApp(42)
	if err != nil {
		t.Fatal(err)
	}
	otherInstance := NewTokenStoreWithRepository(repo)
	if _, ok := otherInstance.Verify(session.Token); !ok {
		t.Fatal("expected persisted session to load before revocation")
	}
	if err := issuer.RevokeUserSessions(42); err != nil {
		t.Fatal(err)
	}
	if _, ok := otherInstance.Verify(session.Token); ok {
		t.Fatal("revoked session must not remain valid in another instance cache")
	}
}

func TestActiveSessionCountDoesNotUseLocalCacheWhenRepositoryFails(t *testing.T) {
	repo := newRecordingSessionRepository()
	store := NewTokenStoreWithRepository(repo)
	if _, err := store.IssueApp(42); err != nil {
		t.Fatal(err)
	}
	repo.countErr = errors.New("session database unavailable")
	count, err := store.ActiveSessionCountStrict(SessionKindApp)
	if !errors.Is(err, repo.countErr) || count != 0 {
		t.Fatalf("expected strict online count failure without local fallback, count=%d err=%v", count, err)
	}
}

func (r *recordingSessionRepository) FindSessionByTokenHash(_ context.Context, tokenHash string) (Session, bool, error) {
	session, ok := r.sessions[tokenHash]
	return session, ok, nil
}

func (r *recordingSessionRepository) RevokeUserSessions(_ context.Context, userID int64) error {
	for tokenHash, session := range r.sessions {
		if session.UserID == userID {
			delete(r.sessions, tokenHash)
		}
	}
	return nil
}

func (r *recordingSessionRepository) CountActiveSessions(_ context.Context, kind string, now time.Time) (int, error) {
	if r.countErr != nil {
		return 0, r.countErr
	}
	users := make(map[int64]struct{})
	for _, session := range r.sessions {
		if !now.Before(session.ExpiresAt) || (kind != "" && session.Kind != kind) {
			continue
		}
		users[session.UserID] = struct{}{}
	}
	return len(users), nil
}
