package points

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) GetAccount(ctx context.Context, userID int64) (Account, bool, error) {
	account, err := scanAccount(r.db.QueryRowContext(ctx, `
select user_id, available_points, frozen_points, total_earned_points, updated_at
from points_accounts
where user_id = $1
`, userID))
	if err == sql.ErrNoRows {
		return Account{UserID: userID}, false, nil
	}
	if err != nil {
		return Account{}, false, err
	}
	return account, true, nil
}

func (r *SQLRepository) SaveAccount(ctx context.Context, account Account) (Account, error) {
	return scanAccount(r.db.QueryRowContext(ctx, `
insert into points_accounts (user_id, available_points, frozen_points, total_earned_points, updated_at)
values ($1,$2,$3,$4,now())
on conflict (user_id) do update set
  available_points = excluded.available_points,
  frozen_points = excluded.frozen_points,
  total_earned_points = excluded.total_earned_points,
  updated_at = now()
returning user_id, available_points, frozen_points, total_earned_points, updated_at
`, account.UserID, account.AvailablePoints, account.FrozenPoints, account.TotalEarnedPoints))
}

func (r *SQLRepository) ListLogsByUser(ctx context.Context, userID int64) ([]Log, error) {
	rows, err := r.db.QueryContext(ctx, pointLogSelect()+` where user_id = $1 order by created_at asc, id asc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLogs(rows)
}

func (r *SQLRepository) ListLogs(ctx context.Context) ([]Log, error) {
	rows, err := r.db.QueryContext(ctx, pointLogSelect()+` order by created_at asc, id asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLogs(rows)
}

func (r *SQLRepository) AddLog(ctx context.Context, log Log) (Log, error) {
	return scanLog(r.db.QueryRowContext(ctx, `
insert into points_logs (user_id, game_id, change_value, before_points, after_points, biz_type, biz_id, reason, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
returning id, user_id, change_value, before_points, after_points, biz_type, biz_id, reason, created_at
`, log.UserID, nullInt64(log.BizID), log.ChangeValue, log.BeforePoints, log.AfterPoints, log.BizType, nullInt64(log.BizID), log.Reason, log.CreatedAt))
}

func pointLogSelect() string {
	return `select id, user_id, change_value, coalesce(before_points, 0), coalesce(after_points, 0), coalesce(biz_type, ''), biz_id, reason, created_at from points_logs`
}

func scanAccount(row interface {
	Scan(dest ...any) error
}) (Account, error) {
	var account Account
	if err := row.Scan(&account.UserID, &account.AvailablePoints, &account.FrozenPoints, &account.TotalEarnedPoints, &account.UpdatedAt); err != nil {
		return Account{}, err
	}
	return account, nil
}

func scanLogs(rows *sql.Rows) ([]Log, error) {
	items := make([]Log, 0)
	for rows.Next() {
		item, err := scanLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanLog(row interface {
	Scan(dest ...any) error
}) (Log, error) {
	var item Log
	var bizID sql.NullInt64
	if err := row.Scan(&item.ID, &item.UserID, &item.ChangeValue, &item.BeforePoints, &item.AfterPoints, &item.BizType, &bizID, &item.Reason, &item.CreatedAt); err != nil {
		return Log{}, err
	}
	item.BizID = bizID.Int64
	return item, nil
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}
