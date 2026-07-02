package aidata

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) IMExportConfig(ctx context.Context) (IMExportConfig, error) {
	var enabled bool
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, `
select im_export_enabled, created_at
from ai_data_snapshots
order by created_at desc, id desc
limit 1
`).Scan(&enabled, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return IMExportConfig{Enabled: false}, nil
	}
	if err != nil {
		return IMExportConfig{}, err
	}
	return IMExportConfig{Enabled: enabled, UpdatedAt: createdAt.Format(time.RFC3339)}, nil
}

func (r *SQLRepository) SetIMExportEnabled(ctx context.Context, enabled bool) (IMExportConfig, error) {
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, `
insert into ai_data_snapshots (im_export_enabled, created_at)
values ($1, now())
returning created_at
`, enabled).Scan(&createdAt)
	if err != nil {
		return IMExportConfig{}, err
	}
	return IMExportConfig{Enabled: enabled, UpdatedAt: createdAt.Format(time.RFC3339)}, nil
}
