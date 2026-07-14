package auth

import "context"

type SessionRepository interface {
	SaveSession(ctx context.Context, session Session) error
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (Session, bool, error)
}
