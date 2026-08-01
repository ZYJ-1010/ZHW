package identity

import (
	"database/sql"
	"testing"
	"time"
)

func TestEvaluateSMSCodeHash(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(5 * time.Minute)

	tests := []struct {
		name        string
		storedHash  string
		candidate   string
		expiresAt   time.Time
		verified    bool
		failedCount int
		want        smsCodeDecision
	}{
		{
			name:       "matching hash is valid",
			storedHash: "hash-123456",
			candidate:  "hash-123456",
			expiresAt:  expiresAt,
			want:       smsCodeValid,
		},
		{
			name:       "different hash is invalid",
			storedHash: "hash-123456",
			candidate:  "hash-000000",
			expiresAt:  expiresAt,
			want:       smsCodeInvalid,
		},
		{
			name:       "expiry boundary is expired",
			storedHash: "hash-123456",
			candidate:  "hash-123456",
			expiresAt:  now,
			want:       smsCodeExpired,
		},
		{
			name:       "consumed code cannot be reused",
			storedHash: "hash-123456",
			candidate:  "hash-123456",
			expiresAt:  expiresAt,
			verified:   true,
			want:       smsCodeConsumed,
		},
		{
			name:        "failure limit locks code",
			storedHash:  "hash-123456",
			candidate:   "hash-123456",
			expiresAt:   expiresAt,
			failedCount: smsMaxFailures,
			want:        smsCodeLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateSMSCodeHash(tt.storedHash, tt.candidate, tt.expiresAt, tt.verified, tt.failedCount, smsMaxFailures, now); got != tt.want {
				t.Fatalf("evaluateSMSCodeHash()=%v want %v", got, tt.want)
			}
		})
	}
}

type fixedIdentityRecordScanner struct {
	createdAt time.Time
	updatedAt time.Time
}

func (s fixedIdentityRecordScanner) Scan(dest ...any) error {
	*dest[0].(*int64) = 10001
	*dest[1].(*string) = string(StatusVerified)
	*dest[2].(*sql.NullString) = sql.NullString{String: "138****0001", Valid: true}
	*dest[3].(*sql.NullString) = sql.NullString{String: "encrypted-phone", Valid: true}
	*dest[4].(*bool) = true
	*dest[5].(*bool) = true
	*dest[6].(*bool) = false
	*dest[7].(*sql.NullString) = sql.NullString{String: "测*", Valid: true}
	*dest[8].(*sql.NullString) = sql.NullString{String: "encrypted-name", Valid: true}
	*dest[9].(*sql.NullString) = sql.NullString{String: "CS", Valid: true}
	*dest[10].(*sql.NullString) = sql.NullString{String: "110***********1234", Valid: true}
	*dest[11].(*sql.NullString) = sql.NullString{String: "encrypted-id-card", Valid: true}
	*dest[12].(*sql.NullString) = sql.NullString{String: "pending", Valid: true}
	*dest[13].(*sql.NullString) = sql.NullString{}
	*dest[14].(*time.Time) = s.createdAt
	*dest[15].(*time.Time) = s.updatedAt
	return nil
}

func TestScanRecordReadsDatabaseCreationTime(t *testing.T) {
	createdAt := time.Date(2026, 7, 20, 8, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	updatedAt := createdAt.Add(48 * time.Hour)
	record, err := scanRecord(fixedIdentityRecordScanner{createdAt: createdAt, updatedAt: updatedAt})
	if err != nil {
		t.Fatal(err)
	}
	if record.CreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("createdAt=%q want %q", record.CreatedAt, createdAt.Format(time.RFC3339))
	}
	if record.UpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Fatalf("updatedAt=%q want %q", record.UpdatedAt, updatedAt.Format(time.RFC3339))
	}
	if record.CreatedAt == record.UpdatedAt {
		t.Fatalf("database creation and update times must remain distinct: %+v", record)
	}
}
