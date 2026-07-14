package games

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SQLServiceConfirmRepository struct {
	db *sql.DB
}

func NewSQLServiceConfirmRepository(db *sql.DB) *SQLServiceConfirmRepository {
	return &SQLServiceConfirmRepository{db: db}
}

func (r *SQLServiceConfirmRepository) GetConfirm(ctx context.Context, gameID int64) (ServiceConfirm, []ServiceConfirmItem, bool, error) {
	confirm, err := scanServiceConfirm(r.db.QueryRowContext(ctx, `
select id, game_id, status, completed_at, created_at
from game_service_confirms
where game_id = $1
`, gameID))
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceConfirm{}, nil, false, nil
	}
	if err != nil {
		return ServiceConfirm{}, nil, false, err
	}
	items, err := r.ListConfirmItems(ctx, gameID)
	if err != nil {
		return ServiceConfirm{}, nil, false, err
	}
	confirm.ConfirmedBy = confirmedBy(items)
	return confirm, items, true, nil
}

func (r *SQLServiceConfirmRepository) SaveConfirm(ctx context.Context, confirm ServiceConfirm) (ServiceConfirm, error) {
	completedAt := parseCompletedAt(confirm.CompletedAt)
	saved, err := scanServiceConfirm(r.db.QueryRowContext(ctx, `
insert into game_service_confirms (game_id, status, completed_at, created_at, updated_at)
values ($1,$2,$3,$4,now())
on conflict (game_id) do update set
  status = excluded.status,
  completed_at = excluded.completed_at,
  updated_at = now()
returning id, game_id, status, completed_at, created_at
`, confirm.GameID, confirm.Status, completedAt, confirm.CreatedAt))
	if err != nil {
		return ServiceConfirm{}, err
	}
	saved.ConfirmedBy = append([]int64(nil), confirm.ConfirmedBy...)
	return saved, nil
}

func (r *SQLServiceConfirmRepository) SaveConfirmItem(ctx context.Context, item ServiceConfirmItem) (ServiceConfirmItem, error) {
	fileIDs, _ := json.Marshal(item.FileIDs)
	return scanServiceConfirmItem(r.db.QueryRowContext(ctx, `
insert into game_service_confirm_items (confirm_id, game_id, user_id, note, file_ids, created_at)
values ($1,$2,$3,$4,$5,$6)
on conflict (game_id, user_id) do update set
  note = game_service_confirm_items.note
returning id, confirm_id, game_id, user_id, note, file_ids, created_at
`, item.ConfirmID, item.GameID, item.UserID, item.Note, fileIDs, item.CreatedAt))
}

func (r *SQLServiceConfirmRepository) ListConfirmItems(ctx context.Context, gameID int64) ([]ServiceConfirmItem, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, confirm_id, game_id, user_id, note, file_ids, created_at
from game_service_confirm_items
where game_id = $1
order by created_at asc, id asc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ServiceConfirmItem, 0)
	for rows.Next() {
		item, err := scanServiceConfirmItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanServiceConfirm(row interface {
	Scan(dest ...any) error
}) (ServiceConfirm, error) {
	var confirm ServiceConfirm
	var completedAt sql.NullTime
	if err := row.Scan(&confirm.ID, &confirm.GameID, &confirm.Status, &completedAt, &confirm.CreatedAt); err != nil {
		return ServiceConfirm{}, err
	}
	if completedAt.Valid {
		confirm.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}
	return confirm, nil
}

func scanServiceConfirmItem(row interface {
	Scan(dest ...any) error
}) (ServiceConfirmItem, error) {
	var item ServiceConfirmItem
	var note sql.NullString
	var rawFileIDs []byte
	if err := row.Scan(&item.ID, &item.ConfirmID, &item.GameID, &item.UserID, &note, &rawFileIDs, &item.CreatedAt); err != nil {
		return ServiceConfirmItem{}, err
	}
	item.Note = note.String
	if len(rawFileIDs) > 0 {
		_ = json.Unmarshal(rawFileIDs, &item.FileIDs)
	}
	return item, nil
}

func confirmedBy(items []ServiceConfirmItem) []int64 {
	result := make([]int64, 0, len(items))
	for _, item := range items {
		result = append(result, item.UserID)
	}
	return result
}

func parseCompletedAt(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
