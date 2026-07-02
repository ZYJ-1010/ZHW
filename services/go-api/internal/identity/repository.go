package identity

import (
	"context"
	"time"
)

type Repository interface {
	SaveRecord(ctx context.Context, record Record) error
	FindRecord(ctx context.Context, userID int64) (Record, bool, error)
	ListRecords(ctx context.Context) ([]Record, error)
	SaveSMSCode(ctx context.Context, record SMSCodeRecord) error
	SaveFaceIDSession(ctx context.Context, record FaceIDSessionRecord) error
}

type SMSCodeRecord struct {
	UserID            int64
	Scene             string
	PhoneMasked       string
	CodeHash          string
	SentAt            time.Time
	ExpiresAt         time.Time
	VerifiedAt        *time.Time
	VerifyFailedCount int
}

type FaceIDSessionRecord struct {
	UserID                  int64
	FaceTokenHash           string
	Status                  string
	DetectAuthPayloadDigest string
	CallbackPayloadDigest   string
	FailureReason           string
	CreatedAt               time.Time
	CompletedAt             *time.Time
}
