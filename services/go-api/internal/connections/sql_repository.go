package connections

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

func (r *SQLRepository) UpsertConnection(ctx context.Context, connection Connection, strengthDelta int) (Connection, error) {
	return scanConnection(r.db.QueryRowContext(ctx, `
insert into user_connections (user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
on conflict (user_id, connected_user_id, relation_type, source_type, source_id) do update set
  strength_score = user_connections.strength_score + $9,
  updated_at = now()
returning id, user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at
`, connection.UserID, connection.ConnectedUserID, connection.RelationType, connection.SourceType, connection.SourceID, connection.StrengthScore, connection.CreatedAt, connection.UpdatedAt, strengthDelta))
}

func (r *SQLRepository) UpsertConnectionPair(ctx context.Context, first Connection, second Connection, strengthDelta int) (Connection, Connection, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Connection{}, Connection{}, err
	}
	defer tx.Rollback()
	firstSaved, err := upsertConnectionRow(ctx, tx, first, strengthDelta)
	if err != nil {
		return Connection{}, Connection{}, err
	}
	secondSaved, err := upsertConnectionRow(ctx, tx, second, strengthDelta)
	if err != nil {
		return Connection{}, Connection{}, err
	}
	if err = tx.Commit(); err != nil {
		return Connection{}, Connection{}, err
	}
	return firstSaved, secondSaved, nil
}

func upsertConnectionRow(ctx context.Context, tx *sql.Tx, connection Connection, strengthDelta int) (Connection, error) {
	return scanConnection(tx.QueryRowContext(ctx, `
insert into user_connections (user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
on conflict (user_id, connected_user_id, relation_type, source_type, source_id) do update set
  strength_score = user_connections.strength_score + $9,
  updated_at = now()
returning id, user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at
`, connection.UserID, connection.ConnectedUserID, connection.RelationType, connection.SourceType, connection.SourceID, connection.StrengthScore, connection.CreatedAt, connection.UpdatedAt, strengthDelta))
}

func (r *SQLRepository) ListConnections(ctx context.Context, userID int64) ([]Connection, error) {
	rows, err := r.db.QueryContext(ctx, connectionSelect()+` where user_id = $1 order by updated_at desc, id desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConnections(rows)
}

func (r *SQLRepository) ListAllConnections(ctx context.Context) ([]Connection, error) {
	rows, err := r.db.QueryContext(ctx, connectionSelect()+` order by updated_at desc, id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConnections(rows)
}

func (r *SQLRepository) FindConnection(ctx context.Context, connectionID int64) (Connection, bool, error) {
	item, err := scanConnection(r.db.QueryRowContext(ctx, connectionSelect()+` where id = $1`, connectionID))
	if errors.Is(err, sql.ErrNoRows) {
		return Connection{}, false, nil
	}
	if err != nil {
		return Connection{}, false, err
	}
	return item, true, nil
}

func (r *SQLRepository) AddFollowLog(ctx context.Context, log FollowLog) (FollowLog, error) {
	return scanFollowLog(r.db.QueryRowContext(ctx, `
insert into connection_follow_logs (connection_id, operator_user_id, follow_type, content, next_follow_at, created_at)
values ($1,$2,$3,$4,$5,$6)
returning id, connection_id, operator_user_id, follow_type, content, next_follow_at, created_at
`, log.ConnectionID, log.OperatorUserID, log.FollowType, log.Content, nullFollowTime(log.NextFollowAt), log.CreatedAt))
}

func (r *SQLRepository) IncrementStrength(ctx context.Context, connectionID int64, delta int) (Connection, error) {
	return scanConnection(r.db.QueryRowContext(ctx, `
update user_connections
set strength_score = strength_score + $2, updated_at = now()
where id = $1
returning id, user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at
`, connectionID, delta))
}

func (r *SQLRepository) AddFollowLogAndIncrement(ctx context.Context, log FollowLog, strengthDelta int) (FollowLog, Connection, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return FollowLog{}, Connection{}, err
	}
	defer tx.Rollback()
	savedLog, err := scanFollowLog(tx.QueryRowContext(ctx, `
insert into connection_follow_logs (connection_id, operator_user_id, follow_type, content, next_follow_at, created_at)
values ($1,$2,$3,$4,$5,$6)
returning id, connection_id, operator_user_id, follow_type, content, next_follow_at, created_at
`, log.ConnectionID, log.OperatorUserID, log.FollowType, log.Content, nullFollowTime(log.NextFollowAt), log.CreatedAt))
	if err != nil {
		return FollowLog{}, Connection{}, err
	}
	connection, err := scanConnection(tx.QueryRowContext(ctx, `
update user_connections
set strength_score = strength_score + $2, updated_at = now()
where id = $1
returning id, user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at
`, log.ConnectionID, strengthDelta))
	if errors.Is(err, sql.ErrNoRows) {
		return FollowLog{}, Connection{}, ErrConnectionNotFound
	}
	if err != nil {
		return FollowLog{}, Connection{}, err
	}
	if err = tx.Commit(); err != nil {
		return FollowLog{}, Connection{}, err
	}
	return savedLog, connection, nil
}

func connectionSelect() string {
	return `select id, user_id, connected_user_id, relation_type, source_type, source_id, strength_score, created_at, updated_at from user_connections`
}

func scanConnections(rows *sql.Rows) ([]Connection, error) {
	items := make([]Connection, 0)
	for rows.Next() {
		item, err := scanConnection(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanConnection(row interface {
	Scan(dest ...any) error
}) (Connection, error) {
	var item Connection
	var sourceID sql.NullInt64
	if err := row.Scan(&item.ID, &item.UserID, &item.ConnectedUserID, &item.RelationType, &item.SourceType, &sourceID, &item.StrengthScore, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Connection{}, err
	}
	item.SourceID = sourceID.Int64
	return item, nil
}

func scanFollowLog(row interface {
	Scan(dest ...any) error
}) (FollowLog, error) {
	var item FollowLog
	var nextFollowAt sql.NullTime
	if err := row.Scan(&item.ID, &item.ConnectionID, &item.OperatorUserID, &item.FollowType, &item.Content, &nextFollowAt, &item.CreatedAt); err != nil {
		return FollowLog{}, err
	}
	if nextFollowAt.Valid {
		item.NextFollowAt = nextFollowAt.Time.Format(time.RFC3339)
	}
	return item, nil
}
