package auth

import (
	"context"
	"database/sql"
	"time"
)

type SQLSessionRepository struct {
	db *sql.DB
}

func NewSQLSessionRepository(db *sql.DB) *SQLSessionRepository {
	return &SQLSessionRepository{db: db}
}

func (r *SQLSessionRepository) SaveSession(ctx context.Context, session Session) error {
	_, err := r.db.ExecContext(ctx, `
insert into app_auth_sessions (token_hash, user_id, session_kind, expires_at)
values ($1,$2,$3,$4)
on conflict (token_hash) do update set
  user_id = excluded.user_id,
  session_kind = excluded.session_kind,
  expires_at = excluded.expires_at,
  revoked_at = null
`, hashToken(session.Token), session.UserID, session.Kind, session.ExpiresAt)
	return err
}

func (r *SQLSessionRepository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (Session, bool, error) {
	var session Session
	var revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
select token_hash, user_id, session_kind, expires_at, revoked_at
from app_auth_sessions
where token_hash = $1
`, tokenHash).Scan(&session.Token, &session.UserID, &session.Kind, &session.ExpiresAt, &revokedAt)
	if err == sql.ErrNoRows {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	if revokedAt.Valid || time.Now().After(session.ExpiresAt) {
		return Session{}, false, nil
	}
	return session, true, nil
}

func (r *SQLSessionRepository) RevokeUserSessions(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
update app_auth_sessions
set revoked_at = now()
where user_id = $1 and revoked_at is null
`, userID)
	return err
}
