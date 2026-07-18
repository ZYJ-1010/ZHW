package audit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type SQLBehaviorRepository struct {
	db *sql.DB
}

type SQLOperationRepository struct {
	db *sql.DB
}

func NewSQLBehaviorRepository(db *sql.DB) *SQLBehaviorRepository {
	return &SQLBehaviorRepository{db: db}
}

func NewSQLOperationRepository(db *sql.DB) *SQLOperationRepository {
	return &SQLOperationRepository{db: db}
}

func (r *SQLBehaviorRepository) SaveBehavior(ctx context.Context, log BehaviorLog) error {
	_, err := r.db.ExecContext(ctx, `
insert into user_behavior_logs (
  user_id, event_type, event_code, target_type, business_type, target_id, business_id,
  page_path, keyword, source, device, ip, extra, created_at, occurred_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
`, nullInt64(log.UserID), log.EventType, log.EventCode, log.TargetType, log.BusinessType, nullInt64(log.TargetID), nullInt64(log.BusinessID), nullString(log.PagePath), nullString(log.Keyword), log.Source, nullString(log.Device), nullString(log.IP), nullJSON(log.Extra), log.CreatedAt, log.OccurredAt)
	return err
}

func (r *SQLBehaviorRepository) ListBehavior(ctx context.Context) ([]BehaviorLog, error) {
	rows, err := r.queryBehavior(ctx, BehaviorQuery{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBehaviorRows(rows)
}

func (r *SQLOperationRepository) SaveOperation(ctx context.Context, log OperationLog) error {
	_, err := r.db.ExecContext(ctx, `
insert into operation_logs (
  admin_user_id, action, target_type, target_id, request_id, ip, detail_json, before_json, after_json, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
`, nullInt64(log.AdminUserID), log.Action, nullString(log.TargetType), nullString(log.TargetID), nullString(log.RequestID), nullString(log.IP), nullJSON(log.Detail), nullJSON(log.Before), nullJSON(log.After), log.CreatedAt)
	return err
}

func (r *SQLOperationRepository) ListOperations(ctx context.Context) ([]OperationLog, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, admin_user_id, action, target_type, target_id, request_id, ip, detail_json, before_json, after_json, created_at
from operation_logs
order by created_at desc, id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OperationLog, 0)
	for rows.Next() {
		item, err := scanOperationLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLBehaviorRepository) QueryBehavior(ctx context.Context, query BehaviorQuery) ([]BehaviorLog, error) {
	rows, err := r.queryBehavior(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBehaviorRows(rows)
}

func (r *SQLBehaviorRepository) queryBehavior(ctx context.Context, query BehaviorQuery) (*sql.Rows, error) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if query.UserID > 0 {
		add("user_id = $%d", query.UserID)
	}
	if query.EventType != "" {
		add("event_type = $%d", query.EventType)
	}
	if query.EventCode != "" {
		add("event_code = $%d", query.EventCode)
	}
	sqlText := `
select id, user_id, event_type, event_code, target_type, business_type, target_id, business_id,
  page_path, keyword, source, device, ip, extra, created_at, occurred_at
from user_behavior_logs`
	if len(clauses) > 0 {
		sqlText += " where " + strings.Join(clauses, " and ")
	}
	sqlText += " order by occurred_at desc, id desc"
	return r.db.QueryContext(ctx, sqlText, args...)
}

func scanBehaviorRows(rows *sql.Rows) ([]BehaviorLog, error) {
	items := make([]BehaviorLog, 0)
	for rows.Next() {
		item, err := scanBehaviorLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanBehaviorLog(rows *sql.Rows) (BehaviorLog, error) {
	var item BehaviorLog
	var userID sql.NullInt64
	var targetID sql.NullInt64
	var businessID sql.NullInt64
	var pagePath sql.NullString
	var keyword sql.NullString
	var device sql.NullString
	var ip sql.NullString
	var extra []byte
	var createdAt time.Time
	var occurredAt time.Time
	if err := rows.Scan(&item.ID, &userID, &item.EventType, &item.EventCode, &item.TargetType, &item.BusinessType, &targetID, &businessID, &pagePath, &keyword, &item.Source, &device, &ip, &extra, &createdAt, &occurredAt); err != nil {
		return BehaviorLog{}, err
	}
	item.UserID = userID.Int64
	item.TargetID = targetID.Int64
	item.BusinessID = businessID.Int64
	item.PagePath = pagePath.String
	item.Keyword = keyword.String
	item.Device = device.String
	item.IP = ip.String
	item.Extra = append(item.Extra, extra...)
	item.CreatedAt = createdAt
	item.OccurredAt = occurredAt
	return item, nil
}

func scanOperationLog(rows *sql.Rows) (OperationLog, error) {
	var item OperationLog
	var adminUserID sql.NullInt64
	var targetType sql.NullString
	var targetID sql.NullString
	var requestID sql.NullString
	var ip sql.NullString
	var detail []byte
	var before []byte
	var after []byte
	if err := rows.Scan(&item.ID, &adminUserID, &item.Action, &targetType, &targetID, &requestID, &ip, &detail, &before, &after, &item.CreatedAt); err != nil {
		return OperationLog{}, err
	}
	item.AdminUserID = adminUserID.Int64
	item.TargetType = targetType.String
	item.TargetID = targetID.String
	item.RequestID = requestID.String
	item.IP = ip.String
	item.Detail = append(item.Detail, detail...)
	item.Before = append(item.Before, before...)
	item.After = append(item.After, after...)
	return item, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}

func nullJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
