package auth

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"time"
)

type SQLSessionRepository struct {
	db *sql.DB
}

var _ PhoneCodeRepository = (*SQLSessionRepository)(nil)

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

func (r *SQLSessionRepository) RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `
update app_auth_sessions
set revoked_at = now()
where token_hash = $1 and revoked_at is null
`, tokenHash)
	return err
}

func (r *SQLSessionRepository) CountActiveSessions(ctx context.Context, kind string, now time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
select count(distinct user_id)
from app_auth_sessions
where session_kind = $1
  and revoked_at is null
  and expires_at > $2
`, kind, now).Scan(&count)
	return count, err
}

func (r *SQLSessionRepository) SaveAppLoginAt(ctx context.Context, userID int64, occurredAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
insert into user_login_activity (user_id, last_app_login_at, updated_at)
values ($1,$2,$2)
on conflict (user_id) do update set
  last_app_login_at = excluded.last_app_login_at,
  updated_at = excluded.updated_at
`, userID, occurredAt)
	return err
}

func (r *SQLSessionRepository) LastAppLoginAt(ctx context.Context, userID int64) (time.Time, bool, error) {
	var occurredAt time.Time
	err := r.db.QueryRowContext(ctx, `select last_app_login_at from user_login_activity where user_id=$1`, userID).Scan(&occurredAt)
	if err == sql.ErrNoRows {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return occurredAt, true, nil
}

func (r *SQLSessionRepository) ReservePhoneCodeSend(ctx context.Context, reservation PhoneCodeSendReservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
insert into auth_phone_code_states (phone_key, phone_masked, day_key, day_count)
values ($1,$2,$3,0)
on conflict (phone_key) do nothing
`, reservation.PhoneKey, reservation.PhoneMasked, reservation.DayKey); err != nil {
		return err
	}
	var pendingToken sql.NullString
	var pendingStartedAt sql.NullTime
	var sentAt sql.NullTime
	var dayKey sql.NullString
	var dayCount int
	if err := tx.QueryRowContext(ctx, `
select pending_token, pending_started_at, sent_at, day_key, day_count
from auth_phone_code_states
where phone_key = $1
for update
`, reservation.PhoneKey).Scan(&pendingToken, &pendingStartedAt, &sentAt, &dayKey, &dayCount); err != nil {
		return err
	}
	if pendingToken.Valid && pendingStartedAt.Valid && reservation.StartedAt.Before(pendingStartedAt.Time.Add(reservation.ResendInterval)) {
		return ErrPhoneCodeRateLimit
	}
	if sentAt.Valid && reservation.StartedAt.Before(sentAt.Time.Add(reservation.ResendInterval)) {
		return ErrPhoneCodeRateLimit
	}
	if dayKey.String != reservation.DayKey {
		dayCount = 0
	}
	if reservation.DailyLimit > 0 && dayCount >= reservation.DailyLimit {
		return ErrPhoneCodeDailyLimit
	}
	if _, err := tx.ExecContext(ctx, `
update auth_phone_code_states
set phone_masked = $2,
    pending_token = $3,
    pending_scene = $4,
    pending_code_hash = $5,
    pending_expires_at = $6,
    pending_started_at = $7,
    updated_at = $7
where phone_key = $1
`, reservation.PhoneKey, reservation.PhoneMasked, reservation.ReservationToken, reservation.Scene, reservation.CodeHash, reservation.ExpiresAt, reservation.StartedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLSessionRepository) ConfirmPhoneCodeSend(ctx context.Context, phoneKey string, reservationToken string, sentAt time.Time, dayKey string) error {
	result, err := r.db.ExecContext(ctx, `
update auth_phone_code_states
set scene = pending_scene,
    code_hash = pending_code_hash,
    sent_at = $3,
    expires_at = pending_expires_at,
    failed_attempts = 0,
    consumed_at = null,
    claim_token = null,
    day_count = case when day_key = $4 then day_count + 1 else 1 end,
    day_key = $4,
    pending_token = null,
    pending_scene = null,
    pending_code_hash = null,
    pending_expires_at = null,
    pending_started_at = null,
    updated_at = $3
where phone_key = $1 and pending_token = $2
`, phoneKey, reservationToken, sentAt, dayKey)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrPhoneCodeSendFailed
	}
	return nil
}

func (r *SQLSessionRepository) CancelPhoneCodeSend(ctx context.Context, phoneKey string, reservationToken string) error {
	_, err := r.db.ExecContext(ctx, `
update auth_phone_code_states
set pending_token = null,
    pending_scene = null,
    pending_code_hash = null,
    pending_expires_at = null,
    pending_started_at = null,
    updated_at = now()
where phone_key = $1 and pending_token = $2
`, phoneKey, reservationToken)
	return err
}

func (r *SQLSessionRepository) ClaimPhoneCode(ctx context.Context, phoneKey string, candidateHashes map[string]string, allowedScenes map[string]bool, claimToken string, nowTime time.Time, maxFailures int, pendingTimeout time.Duration) (PhoneCodeClaim, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return PhoneCodeClaim{}, false, err
	}
	defer tx.Rollback()
	var scene sql.NullString
	var storedHash sql.NullString
	var sentAt sql.NullTime
	var expiresAt sql.NullTime
	var consumedAt sql.NullTime
	var failedAttempts int
	var pendingToken sql.NullString
	var pendingStartedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
select scene, code_hash, sent_at, expires_at, consumed_at, failed_attempts,
       pending_token, pending_started_at
from auth_phone_code_states
where phone_key = $1
for update
`, phoneKey).Scan(&scene, &storedHash, &sentAt, &expiresAt, &consumedAt, &failedAttempts, &pendingToken, &pendingStartedAt)
	if err == sql.ErrNoRows {
		return PhoneCodeClaim{}, false, nil
	}
	if err != nil {
		return PhoneCodeClaim{}, false, err
	}
	if pendingToken.Valid && pendingStartedAt.Valid && nowTime.Before(pendingStartedAt.Time.Add(pendingTimeout)) {
		return PhoneCodeClaim{}, false, nil
	}
	if !scene.Valid || !storedHash.Valid || !sentAt.Valid || !expiresAt.Valid || consumedAt.Valid || !nowTime.Before(expiresAt.Time) || failedAttempts >= maxFailures || !allowedScenes[scene.String] {
		return PhoneCodeClaim{}, false, nil
	}
	candidateHash, ok := candidateHashes[scene.String]
	if !ok || subtle.ConstantTimeCompare([]byte(storedHash.String), []byte(candidateHash)) != 1 {
		if _, err := tx.ExecContext(ctx, `
update auth_phone_code_states
set failed_attempts = least(failed_attempts + 1, $2), updated_at = $3
where phone_key = $1 and consumed_at is null
`, phoneKey, maxFailures, nowTime); err != nil {
			return PhoneCodeClaim{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return PhoneCodeClaim{}, false, err
		}
		return PhoneCodeClaim{}, false, nil
	}
	result, err := tx.ExecContext(ctx, `
update auth_phone_code_states
set consumed_at = $2, claim_token = $3, updated_at = $2
where phone_key = $1 and consumed_at is null and failed_attempts < $4 and expires_at > $2
`, phoneKey, nowTime, claimToken, maxFailures)
	if err != nil {
		return PhoneCodeClaim{}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return PhoneCodeClaim{}, false, err
	}
	if affected != 1 {
		return PhoneCodeClaim{}, false, nil
	}
	if err := tx.Commit(); err != nil {
		return PhoneCodeClaim{}, false, err
	}
	return PhoneCodeClaim{PhoneKey: phoneKey, Scene: scene.String, ClaimToken: claimToken, SentAt: sentAt.Time, ExpiresAt: expiresAt.Time}, true, nil
}

func (r *SQLSessionRepository) RestorePhoneCodeClaim(ctx context.Context, phoneKey string, claimToken string) error {
	_, err := r.db.ExecContext(ctx, `
update auth_phone_code_states
set consumed_at = null, claim_token = null, updated_at = now()
where phone_key = $1 and claim_token = $2
`, phoneKey, claimToken)
	return err
}
