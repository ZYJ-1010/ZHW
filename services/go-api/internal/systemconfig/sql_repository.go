package systemconfig

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Get(ctx context.Context, key string) (json.RawMessage, error) {
	var value []byte
	err := r.db.QueryRowContext(ctx, `
select config_value
from system_configs
where config_key = $1 and status = 'active'
limit 1
`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(value), nil
}

func (r *SQLRepository) Set(ctx context.Context, key string, value json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
insert into system_configs (config_key, config_value, status, updated_at)
values ($1, $2::jsonb, 'active', now())
on conflict (config_key) do update set
	config_value = excluded.config_value,
	status = 'active',
	updated_at = now()
`, key, string(value))
	return err
}
