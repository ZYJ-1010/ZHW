package gamedrafts

import (
	"context"
	"database/sql"
	"encoding/json"
)

type SQLRepository struct{ db *sql.DB }

func NewSQLRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) Create(ctx context.Context, draft Draft) (Draft, error) {
	return scanDraft(r.db.QueryRowContext(ctx, `
insert into game_create_drafts (creator_user_id, title, payload, created_at, updated_at)
values ($1,$2,$3,$4,$5)
returning id, creator_user_id, title, payload, created_at, updated_at
`, draft.CreatorUserID, draft.Title, []byte(draft.Payload), draft.CreatedAt, draft.UpdatedAt))
}

func (r *SQLRepository) Update(ctx context.Context, draft Draft) (Draft, error) {
	return scanDraft(r.db.QueryRowContext(ctx, `
update game_create_drafts set title=$3, payload=$4, updated_at=$5
where id=$1 and creator_user_id=$2
returning id, creator_user_id, title, payload, created_at, updated_at
`, draft.ID, draft.CreatorUserID, draft.Title, []byte(draft.Payload), draft.UpdatedAt))
}

func (r *SQLRepository) GetByUser(ctx context.Context, userID int64, draftID int64) (Draft, error) {
	draft, err := scanDraft(r.db.QueryRowContext(ctx, `
select id, creator_user_id, title, payload, created_at, updated_at
from game_create_drafts where id=$1 and creator_user_id=$2
`, draftID, userID))
	if err == sql.ErrNoRows {
		return Draft{}, ErrNotFound
	}
	return draft, err
}

func (r *SQLRepository) ListByUser(ctx context.Context, userID int64) ([]Draft, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, creator_user_id, title, payload, created_at, updated_at
from game_create_drafts where creator_user_id=$1 order by updated_at desc, id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Draft, 0)
	for rows.Next() {
		draft, scanErr := scanDraft(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, draft)
	}
	return items, rows.Err()
}

func (r *SQLRepository) DeleteByUser(ctx context.Context, userID int64, draftID int64) error {
	result, err := r.db.ExecContext(ctx, `delete from game_create_drafts where id=$1 and creator_user_id=$2`, draftID, userID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrNotFound
	}
	return nil
}

func scanDraft(row interface{ Scan(...any) error }) (Draft, error) {
	var draft Draft
	var payload []byte
	err := row.Scan(&draft.ID, &draft.CreatorUserID, &draft.Title, &payload, &draft.CreatedAt, &draft.UpdatedAt)
	if err != nil {
		return Draft{}, err
	}
	if !json.Valid(payload) {
		return Draft{}, ErrInvalid
	}
	draft.Payload = append(json.RawMessage(nil), payload...)
	return draft, nil
}
