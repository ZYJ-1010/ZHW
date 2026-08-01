package identity

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) SaveRecord(ctx context.Context, record Record) error {
	createdAt := parseTimeOrNow(record.CreatedAt)
	updatedAt := parseTimeOrNow(record.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `
insert into identity_verification_records (
  user_id, status, phone_masked, phone_encrypted, real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext, sms_verified, phone_verified, face_verified,
  wechat_realname_consistency, failure_reason, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
on conflict (user_id) do update set
  status = excluded.status,
  phone_masked = excluded.phone_masked,
  phone_encrypted = excluded.phone_encrypted,
  real_name_masked = excluded.real_name_masked,
  real_name_ciphertext = excluded.real_name_ciphertext,
  real_name_initials = excluded.real_name_initials,
  id_card_masked = excluded.id_card_masked,
  id_card_ciphertext = excluded.id_card_ciphertext,
  sms_verified = excluded.sms_verified,
  phone_verified = excluded.phone_verified,
  face_verified = excluded.face_verified,
  wechat_realname_consistency = excluded.wechat_realname_consistency,
  failure_reason = excluded.failure_reason,
  updated_at = excluded.updated_at
`, record.UserID, string(record.Status), nullString(record.PhoneMasked), nullString(record.PhoneEncrypted), nullString(record.RealNameMasked), nullString(record.RealNameCiphertext), nullString(record.RealNameInitials), nullString(record.IDCardMasked), nullString(record.IDCardCiphertext), record.SMSVerified, record.PhoneVerified, record.FaceVerified, record.WechatRealnameConsistency, nullString(record.FailureReason), createdAt, updatedAt)
	return err
}

func (r *SQLRepository) UpdateRecordIfStatus(ctx context.Context, record Record, expected Status) (Record, bool, error) {
	updatedAt := parseTimeOrNow(record.UpdatedAt)
	row := r.db.QueryRowContext(ctx, `
update identity_verification_records
set status = $2,
    phone_verified = $3,
    failure_reason = $4,
    updated_at = $5
where user_id = $1 and status = $6
returning user_id, status, phone_masked, phone_encrypted, sms_verified, phone_verified, face_verified,
  real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext,
  wechat_realname_consistency, failure_reason, created_at, updated_at
`, record.UserID, string(record.Status), record.PhoneVerified, nullString(record.FailureReason), updatedAt, string(expected))
	updated, err := scanRecord(row)
	if err == sql.ErrNoRows {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	return updated, true, nil
}

func (r *SQLRepository) FindRecord(ctx context.Context, userID int64) (Record, bool, error) {
	rows, err := r.queryRecords(ctx, "where user_id = $1", userID)
	if err != nil {
		return Record{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return Record{}, false, rows.Err()
	}
	record, err := scanRecord(rows)
	if err != nil {
		return Record{}, false, err
	}
	return record, true, rows.Err()
}

func (r *SQLRepository) ListRecords(ctx context.Context) ([]Record, error) {
	rows, err := r.queryRecords(ctx, "order by updated_at desc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Record, 0)
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	return items, rows.Err()
}

func (r *SQLRepository) SaveSMSCode(ctx context.Context, record SMSCodeRecord) error {
	_, err := r.db.ExecContext(ctx, `
insert into identity_sms_code_records (
  user_id, scene, phone_masked, code_hash, sent_at, expires_at, verified_at, verify_failed_count
) values ($1,$2,$3,$4,$5,$6,$7,$8)
`, record.UserID, record.Scene, nullString(record.PhoneMasked), record.CodeHash, record.SentAt, record.ExpiresAt, record.VerifiedAt, record.VerifyFailedCount)
	return err
}

type smsCodeDecision uint8

const (
	smsCodeInvalid smsCodeDecision = iota
	smsCodeValid
	smsCodeExpired
	smsCodeConsumed
	smsCodeLocked
)

func evaluateSMSCodeHash(storedHash string, candidateHash string, expiresAt time.Time, verified bool, failedCount int, maxFailures int, currentTime time.Time) smsCodeDecision {
	if verified {
		return smsCodeConsumed
	}
	if !currentTime.Before(expiresAt) {
		return smsCodeExpired
	}
	if maxFailures > 0 && failedCount >= maxFailures {
		return smsCodeLocked
	}
	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidateHash)) != 1 {
		return smsCodeInvalid
	}
	return smsCodeValid
}

// VerifyAndConsumeSMSCode serializes verification on the latest code for a
// user and scene. The plaintext code never crosses the repository boundary.
func (r *SQLRepository) VerifyAndConsumeSMSCode(ctx context.Context, userID int64, scene string, codeHash string, verifiedAt time.Time, maxFailures int) (Record, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Record{}, false, err
	}
	defer tx.Rollback()

	var codeID int64
	var storedHash string
	var expiresAt time.Time
	var consumedAt sql.NullTime
	var failedCount int
	err = tx.QueryRowContext(ctx, `
select id, code_hash, expires_at, verified_at, verify_failed_count
from identity_sms_code_records
where user_id = $1 and scene = $2
order by sent_at desc, id desc
limit 1
for update
`, userID, scene).Scan(&codeID, &storedHash, &expiresAt, &consumedAt, &failedCount)
	if err == sql.ErrNoRows {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}

	decision := evaluateSMSCodeHash(storedHash, codeHash, expiresAt, consumedAt.Valid, failedCount, maxFailures, verifiedAt)
	if decision == smsCodeInvalid {
		if _, err := tx.ExecContext(ctx, `
update identity_sms_code_records
set verify_failed_count = least(verify_failed_count + 1, $2)
where id = $1 and verified_at is null
`, codeID, maxFailures); err != nil {
			return Record{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return Record{}, false, err
		}
		return Record{}, false, nil
	}
	if decision != smsCodeValid {
		return Record{}, false, nil
	}

	record, err := scanRecord(tx.QueryRowContext(ctx, `
update identity_verification_records
set sms_verified = true,
    status = case when status in ($2, $3) then $4 else status end,
    updated_at = $5
where user_id = $1
returning user_id, status, phone_masked, phone_encrypted, sms_verified, phone_verified, face_verified,
  real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext,
  wechat_realname_consistency, failure_reason, created_at, updated_at
`, userID, string(StatusWechatLoggedIn), string(StatusPhoneBound), string(StatusSMSVerified), verifiedAt))
	if err == sql.ErrNoRows {
		return Record{}, false, ErrRecordNotFound
	}
	if err != nil {
		return Record{}, false, err
	}
	result, err := tx.ExecContext(ctx, `
update identity_sms_code_records
set verified_at = $2
where id = $1
  and verified_at is null
  and expires_at > $2
  and verify_failed_count < $3
`, codeID, verifiedAt, maxFailures)
	if err != nil {
		return Record{}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Record{}, false, err
	}
	if affected != 1 {
		return Record{}, false, nil
	}
	if err := tx.Commit(); err != nil {
		return Record{}, false, err
	}
	return record, true, nil
}

func (r *SQLRepository) SaveFaceIDSession(ctx context.Context, record FaceIDSessionRecord) error {
	_, err := r.db.ExecContext(ctx, `
insert into identity_faceid_sessions (
  user_id, face_token_hash, status, detect_auth_payload_digest, callback_payload_digest,
  failure_reason, created_at, completed_at
) values ($1,$2,$3,$4,$5,$6,$7,$8)
on conflict (face_token_hash) do update set
  status = excluded.status,
  callback_payload_digest = excluded.callback_payload_digest,
  failure_reason = excluded.failure_reason,
  completed_at = excluded.completed_at
`, record.UserID, record.FaceTokenHash, record.Status, nullString(record.DetectAuthPayloadDigest), nullString(record.CallbackPayloadDigest), nullString(record.FailureReason), record.CreatedAt, record.CompletedAt)
	return err
}

// CompleteFaceIDSession atomically consumes a provider token and promotes the
// matching identity record. A completed token is idempotent so a provider retry
// can still repair downstream user-status synchronization.
func (r *SQLRepository) CompleteFaceIDSession(ctx context.Context, tokenHash string, expectedUserID int64, completedAt time.Time) (Record, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Record{}, false, err
	}
	defer tx.Rollback()
	var userID int64
	var sessionStatus string
	err = tx.QueryRowContext(ctx, `
select user_id, status
from identity_faceid_sessions
where face_token_hash = $1
for update
`, tokenHash).Scan(&userID, &sessionStatus)
	if err == sql.ErrNoRows {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	if expectedUserID > 0 && expectedUserID != userID {
		return Record{}, false, nil
	}
	if sessionStatus == string(StatusVerified) {
		record, found, findErr := findRecordWithQueryer(ctx, tx, userID)
		if findErr != nil || !found {
			return Record{}, false, findErr
		}
		if err := tx.Commit(); err != nil {
			return Record{}, false, err
		}
		return record, true, nil
	}
	if sessionStatus != string(StatusFaceIDProcessing) {
		return Record{}, false, nil
	}
	record, err := scanRecord(tx.QueryRowContext(ctx, `
update identity_verification_records
set status = $2,
    face_verified = true,
    wechat_realname_consistency = 'not_supported',
    updated_at = $3
where user_id = $1 and status = $4
returning user_id, status, phone_masked, phone_encrypted, sms_verified, phone_verified, face_verified,
  real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext,
  wechat_realname_consistency, failure_reason, created_at, updated_at
`, userID, string(StatusVerified), completedAt, string(StatusFaceIDProcessing)))
	if err == sql.ErrNoRows {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `
update identity_faceid_sessions
set status = $2, completed_at = $3
where face_token_hash = $1
`, tokenHash, string(StatusVerified), completedAt); err != nil {
		return Record{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Record{}, false, err
	}
	return record, true, nil
}

type identityRecordQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func findRecordWithQueryer(ctx context.Context, queryer identityRecordQueryer, userID int64) (Record, bool, error) {
	record, err := scanRecord(queryer.QueryRowContext(ctx, `
select user_id, status, phone_masked, phone_encrypted, sms_verified, phone_verified, face_verified,
  real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext,
  wechat_realname_consistency, failure_reason, created_at, updated_at
from identity_verification_records
where user_id = $1
`, userID))
	if err == sql.ErrNoRows {
		return Record{}, false, nil
	}
	return record, err == nil, err
}

func (r *SQLRepository) queryRecords(ctx context.Context, suffix string, args ...any) (*sql.Rows, error) {
	query := `
select user_id, status, phone_masked, phone_encrypted, sms_verified, phone_verified, face_verified,
  real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext, wechat_realname_consistency, failure_reason, created_at, updated_at
from identity_verification_records ` + suffix
	return r.db.QueryContext(ctx, query, args...)
}

type recordScanner interface {
	Scan(dest ...any) error
}

func scanRecord(rows recordScanner) (Record, error) {
	var record Record
	var status string
	var phoneMasked sql.NullString
	var phoneEncrypted sql.NullString
	var realNameMasked sql.NullString
	var realNameCiphertext sql.NullString
	var realNameInitials sql.NullString
	var idCardMasked sql.NullString
	var idCardCiphertext sql.NullString
	var consistency sql.NullString
	var failureReason sql.NullString
	var createdAt time.Time
	var updatedAt time.Time
	if err := rows.Scan(&record.UserID, &status, &phoneMasked, &phoneEncrypted, &record.SMSVerified, &record.PhoneVerified, &record.FaceVerified, &realNameMasked, &realNameCiphertext, &realNameInitials, &idCardMasked, &idCardCiphertext, &consistency, &failureReason, &createdAt, &updatedAt); err != nil {
		return Record{}, err
	}
	record.Status = Status(status)
	record.PhoneMasked = phoneMasked.String
	record.PhoneEncrypted = phoneEncrypted.String
	record.RealNameMasked = realNameMasked.String
	record.RealNameCiphertext = realNameCiphertext.String
	record.RealNameInitials = realNameInitials.String
	record.IDCardMasked = idCardMasked.String
	record.IDCardCiphertext = idCardCiphertext.String
	record.WechatRealnameConsistency = consistency.String
	record.FailureReason = failureReason.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func parseTimeOrNow(value string) time.Time {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	if parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", value); err == nil {
		return parsed
	}
	return time.Now()
}
