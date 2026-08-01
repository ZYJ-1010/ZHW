package auth

import (
	"context"
	"time"
)

type SessionRepository interface {
	SaveSession(ctx context.Context, session Session) error
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (Session, bool, error)
	RevokeUserSessions(ctx context.Context, userID int64) error
}

// LoginActivityRepository is deliberately optional so custom session
// repositories remain compatible while SQL deployments persist inactivity.
type LoginActivityRepository interface {
	SaveAppLoginAt(ctx context.Context, userID int64, occurredAt time.Time) error
	LastAppLoginAt(ctx context.Context, userID int64) (time.Time, bool, error)
}

// ActiveSessionCounter is optional so existing custom session repositories
// continue to work. Production uses it to count sessions from the shared
// database instead of only the current process memory.
type ActiveSessionCounter interface {
	CountActiveSessions(ctx context.Context, kind string, now time.Time) (int, error)
}

type SessionTokenRevoker interface {
	RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error
}

type PhoneCodeSendReservation struct {
	PhoneKey         string
	PhoneMasked      string
	Scene            string
	CodeHash         string
	ReservationToken string
	StartedAt        time.Time
	ExpiresAt        time.Time
	DayKey           string
	ResendInterval   time.Duration
	DailyLimit       int
}

type PhoneCodeClaim struct {
	PhoneKey   string
	Scene      string
	ClaimToken string
	SentAt     time.Time
	ExpiresAt  time.Time
}

// PhoneCodeRepository is optional so local and unit-test token stores can keep
// the in-memory implementation. Production's SQL session repository also
// implements this interface, making login and password-reset codes shared by
// every API instance.
type PhoneCodeRepository interface {
	ReservePhoneCodeSend(ctx context.Context, reservation PhoneCodeSendReservation) error
	ConfirmPhoneCodeSend(ctx context.Context, phoneKey string, reservationToken string, sentAt time.Time, dayKey string) error
	CancelPhoneCodeSend(ctx context.Context, phoneKey string, reservationToken string) error
	ClaimPhoneCode(ctx context.Context, phoneKey string, candidateHashes map[string]string, allowedScenes map[string]bool, claimToken string, now time.Time, maxFailures int, pendingTimeout time.Duration) (PhoneCodeClaim, bool, error)
	RestorePhoneCodeClaim(ctx context.Context, phoneKey string, claimToken string) error
}
