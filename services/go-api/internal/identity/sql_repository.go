package identity

import (
	"context"
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
	updatedAt := parseTimeOrNow(record.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `
insert into identity_verification_records (
  user_id, status, phone_masked, phone_encrypted, real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext, sms_verified, phone_verified, face_verified,
  wechat_realname_consistency, failure_reason, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
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
`, record.UserID, string(record.Status), nullString(record.PhoneMasked), nullString(record.PhoneEncrypted), nullString(record.RealNameMasked), nullString(record.RealNameCiphertext), nullString(record.RealNameInitials), nullString(record.IDCardMasked), nullString(record.IDCardCiphertext), record.SMSVerified, record.PhoneVerified, record.FaceVerified, record.WechatRealnameConsistency, nullString(record.FailureReason), updatedAt)
	return err
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

func (r *SQLRepository) queryRecords(ctx context.Context, suffix string, args ...any) (*sql.Rows, error) {
	query := `
select user_id, status, phone_masked, phone_encrypted, sms_verified, phone_verified, face_verified,
  real_name_masked, real_name_ciphertext, real_name_initials, id_card_masked, id_card_ciphertext, wechat_realname_consistency, failure_reason, updated_at
from identity_verification_records ` + suffix
	return r.db.QueryContext(ctx, query, args...)
}

func scanRecord(rows *sql.Rows) (Record, error) {
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
	var updatedAt time.Time
	if err := rows.Scan(&record.UserID, &status, &phoneMasked, &phoneEncrypted, &record.SMSVerified, &record.PhoneVerified, &record.FaceVerified, &realNameMasked, &realNameCiphertext, &realNameInitials, &idCardMasked, &idCardCiphertext, &consistency, &failureReason, &updatedAt); err != nil {
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
