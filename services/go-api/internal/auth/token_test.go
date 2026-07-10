package auth

import (
	"context"
	"testing"
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

type recordingSessionRepository struct {
	sessions map[string]Session
}

func newRecordingSessionRepository() *recordingSessionRepository {
	return &recordingSessionRepository{sessions: make(map[string]Session)}
}

func (r *recordingSessionRepository) SaveSession(_ context.Context, session Session) error {
	session.Token = hashToken(session.Token)
	r.sessions[session.Token] = session
	return nil
}

func (r *recordingSessionRepository) FindSessionByTokenHash(_ context.Context, tokenHash string) (Session, bool, error) {
	session, ok := r.sessions[tokenHash]
	return session, ok, nil
}
