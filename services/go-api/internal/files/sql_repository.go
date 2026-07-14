package files

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

func (r *SQLRepository) SaveFile(ctx context.Context, file File) (File, error) {
	return scanFile(r.db.QueryRowContext(ctx, `
insert into files (
  uploader_user_id, biz_type, object_id, file_name, mime_type, size_bytes,
  sha256, storage_key, access_level, expires_at, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
returning id, uploader_user_id, biz_type, object_id, file_name, mime_type,
  size_bytes, sha256, storage_key, access_level, expires_at, created_at
`, nullInt64(file.UploaderID), file.BizType, nullInt64(file.ObjectID), file.FileName, nullString(file.MimeType), file.Size, nullString(file.SHA256), file.StorageKey, file.AccessLevel, nullTimeString(file.ExpiresAt), file.CreatedAt))
}

func (r *SQLRepository) FindFile(ctx context.Context, fileID int64) (File, bool, error) {
	file, err := scanFile(r.db.QueryRowContext(ctx, `
select id, uploader_user_id, biz_type, object_id, file_name, mime_type,
  size_bytes, sha256, storage_key, access_level, expires_at, created_at
from files
where id = $1
`, fileID))
	if err == sql.ErrNoRows {
		return File{}, false, nil
	}
	if err != nil {
		return File{}, false, err
	}
	return file, true, nil
}

func scanFile(row interface {
	Scan(dest ...any) error
}) (File, error) {
	var file File
	var uploaderID sql.NullInt64
	var objectID sql.NullInt64
	var mimeType sql.NullString
	var sha256 sql.NullString
	var expiresAt sql.NullTime
	if err := row.Scan(&file.ID, &uploaderID, &file.BizType, &objectID, &file.FileName, &mimeType, &file.Size, &sha256, &file.StorageKey, &file.AccessLevel, &expiresAt, &file.CreatedAt); err != nil {
		return File{}, err
	}
	file.UploaderID = uploaderID.Int64
	file.ObjectID = objectID.Int64
	file.MimeType = mimeType.String
	file.SHA256 = sha256.String
	if expiresAt.Valid {
		file.ExpiresAt = expiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return file, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}

func nullTimeString(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}
