package points

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

// ApplyChange updates the account and its ledger in one PostgreSQL
// transaction. The account row is locked before calculating the new balance.
func (r *SQLRepository) ApplyChange(ctx context.Context, userID int64, changeValue int, bizType string, bizID int64, reason string) (Account, Log, error) {
	account, log, _, err := r.applyChange(ctx, userID, changeValue, bizType, bizID, reason, false)
	return account, log, err
}

// ApplyChangeOnce is the idempotent form for automatic rewards.
func (r *SQLRepository) ApplyChangeOnce(ctx context.Context, userID int64, changeValue int, bizType string, bizID int64, reason string) (Account, Log, bool, error) {
	return r.applyChange(ctx, userID, changeValue, bizType, bizID, reason, true)
}

func (r *SQLRepository) applyChange(ctx context.Context, userID int64, changeValue int, bizType string, bizID int64, reason string, once bool) (Account, Log, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Account{}, Log{}, false, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
insert into points_accounts (user_id, available_points, frozen_points, total_earned_points, updated_at)
values ($1,0,0,0,now()) on conflict (user_id) do nothing
`, userID); err != nil {
		return Account{}, Log{}, false, err
	}
	account, err := scanAccount(tx.QueryRowContext(ctx, `
select user_id, available_points, frozen_points, total_earned_points, updated_at
from points_accounts where user_id=$1 for update
`, userID))
	if err != nil {
		return Account{}, Log{}, false, err
	}
	if once {
		existing, err := scanLog(tx.QueryRowContext(ctx, pointLogSelect()+` where user_id=$1 and biz_type=$2 and biz_id=$3 order by id asc limit 1`, userID, bizType, bizID))
		if err == nil {
			if commitErr := tx.Commit(); commitErr != nil {
				return Account{}, Log{}, false, commitErr
			}
			return account, existing, false, nil
		}
		if err != sql.ErrNoRows {
			return Account{}, Log{}, false, err
		}
	}
	after := account.AvailablePoints + changeValue
	if after < 0 {
		return account, Log{}, false, ErrInsufficientPoints
	}
	account.AvailablePoints = after
	if changeValue > 0 {
		account.TotalEarnedPoints += changeValue
	}
	account.UpdatedAt = time.Now()
	account, err = scanAccount(tx.QueryRowContext(ctx, `
update points_accounts set available_points=$2, frozen_points=$3, total_earned_points=$4, updated_at=now()
where user_id=$1
returning user_id, available_points, frozen_points, total_earned_points, updated_at
`, account.UserID, account.AvailablePoints, account.FrozenPoints, account.TotalEarnedPoints))
	if err != nil {
		return Account{}, Log{}, false, err
	}
	log, err := scanLog(tx.QueryRowContext(ctx, `
insert into points_logs (user_id, game_id, change_value, before_points, after_points, biz_type, biz_id, reason, created_at)
values ($1,null,$2,$3,$4,$5,$6,$7,now())
returning id, user_id, change_value, before_points, after_points, coalesce(biz_type, ''), biz_id, reason, created_at
`, userID, changeValue, account.AvailablePoints-changeValue, account.AvailablePoints, bizType, nullInt64(bizID), reason))
	if err != nil {
		return Account{}, Log{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Account{}, Log{}, false, err
	}
	return account, log, true, nil
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
