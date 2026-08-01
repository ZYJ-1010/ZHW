package reports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) SaveReport(ctx context.Context, report Report) (Report, error) {
	if report.ReportType == "credit_appeal" {
		return r.saveCreditAppeal(ctx, report)
	}
	return insertReport(ctx, r.db, report)
}

type reportQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func insertReport(ctx context.Context, queryer reportQueryer, report Report) (Report, error) {
	return scanReport(queryer.QueryRowContext(ctx, `
insert into reports (
  game_id, reporter_user_id, target_user_id, report_type, content, status,
  chat_message_id, file_id, review_id, credit_log_id, revenue_record_id, revenue_frozen,
  revenue_freeze_note, handler_admin_id, handle_result, handled_at, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
returning id, game_id, reporter_user_id, target_user_id, report_type, content, status,
  chat_message_id, file_id, review_id, credit_log_id, revenue_record_id, revenue_frozen,
  revenue_freeze_note, handler_admin_id, handle_result, handled_at, created_at
`, report.GameID, report.ReporterUserID, nullInt64(report.TargetUserID), report.ReportType, nullString(report.Content), report.Status, nullInt64(report.ChatMessageID), nullInt64(report.FileID), nullInt64(report.ReviewID), nullInt64(report.CreditLogID), nullInt64(report.RevenueRecordID), report.RevenueFrozen, nullString(report.RevenueFreezeNote), nullInt64(report.HandlerAdminID), nullString(encodeHandleResult(report)), nullTimeString(report.HandledAt), report.CreatedAt))
}

func (r *SQLRepository) saveCreditAppeal(ctx context.Context, report Report) (Report, error) {
	if report.ReporterUserID <= 0 || report.CreditLogID <= 0 {
		return Report{}, ErrInvalidReport
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Report{}, err
	}
	defer tx.Rollback()

	var linkedAppealID sql.NullInt64
	err = tx.QueryRowContext(ctx, `
select appeal_id
from credit_logs
where id = $1 and user_id = $2 and change_value < 0
for update
`, report.CreditLogID, report.ReporterUserID).Scan(&linkedAppealID)
	if errors.Is(err, sql.ErrNoRows) {
		return Report{}, ErrInvalidReport
	}
	if err != nil {
		return Report{}, err
	}
	if linkedAppealID.Valid {
		return Report{}, ErrCreditAppealExists
	}

	saved, err := insertReport(ctx, tx, report)
	if err != nil {
		return Report{}, err
	}
	result, err := tx.ExecContext(ctx, `
update credit_logs
set appeal_id = $2
where id = $1 and appeal_id is null
`, report.CreditLogID, saved.ID)
	if err != nil {
		return Report{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Report{}, err
	}
	if rows != 1 {
		return Report{}, ErrCreditAppealExists
	}
	if err = tx.Commit(); err != nil {
		return Report{}, err
	}
	return saved, nil
}

func (r *SQLRepository) ListReports(ctx context.Context) ([]Report, error) {
	rows, err := r.queryReports(ctx, "order by created_at desc, id desc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReports(rows)
}

func (r *SQLRepository) ListReportsByUser(ctx context.Context, userID int64) ([]Report, error) {
	rows, err := r.queryReports(ctx, "where reporter_user_id = $1 order by created_at desc, id desc", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReports(rows)
}

func (r *SQLRepository) FindReport(ctx context.Context, reportID int64) (Report, bool, error) {
	report, err := scanReport(r.db.QueryRowContext(ctx, `
select id, game_id, reporter_user_id, target_user_id, report_type, content, status,
  chat_message_id, file_id, review_id, credit_log_id, revenue_record_id, revenue_frozen,
  revenue_freeze_note, handler_admin_id, handle_result, handled_at, created_at
from reports
where id = $1
`, reportID))
	if err == sql.ErrNoRows {
		return Report{}, false, nil
	}
	if err != nil {
		return Report{}, false, err
	}
	return report, true, nil
}

func (r *SQLRepository) UpdateReport(ctx context.Context, report Report) (Report, error) {
	return scanReport(r.db.QueryRowContext(ctx, `
update reports set
  status = $2,
  revenue_frozen = $3,
  revenue_freeze_note = $4,
  handler_admin_id = $5,
  handle_result = $6,
  handled_at = $7,
  file_id = $8
where id = $1
returning id, game_id, reporter_user_id, target_user_id, report_type, content, status,
  chat_message_id, file_id, review_id, credit_log_id, revenue_record_id, revenue_frozen,
  revenue_freeze_note, handler_admin_id, handle_result, handled_at, created_at
`, report.ID, report.Status, report.RevenueFrozen, nullString(report.RevenueFreezeNote), nullInt64(report.HandlerAdminID), nullString(encodeHandleResult(report)), nullTimeString(report.HandledAt), nullInt64(report.FileID)))
}

func (r *SQLRepository) queryReports(ctx context.Context, suffix string, args ...any) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, `
select id, game_id, reporter_user_id, target_user_id, report_type, content, status,
  chat_message_id, file_id, review_id, credit_log_id, revenue_record_id, revenue_frozen,
  revenue_freeze_note, handler_admin_id, handle_result, handled_at, created_at
from reports `+suffix, args...)
}

func scanReports(rows *sql.Rows) ([]Report, error) {
	items := make([]Report, 0)
	for rows.Next() {
		item, err := scanReportRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanReport(row interface {
	Scan(dest ...any) error
}) (Report, error) {
	var report Report
	var targetUserID sql.NullInt64
	var content sql.NullString
	var chatMessageID sql.NullInt64
	var fileID sql.NullInt64
	var reviewID sql.NullInt64
	var creditLogID sql.NullInt64
	var revenueRecordID sql.NullInt64
	var revenueFreezeNote sql.NullString
	var handlerAdminID sql.NullInt64
	var handleResult sql.NullString
	var handledAt sql.NullTime
	if err := row.Scan(&report.ID, &report.GameID, &report.ReporterUserID, &targetUserID, &report.ReportType, &content, &report.Status, &chatMessageID, &fileID, &reviewID, &creditLogID, &revenueRecordID, &report.RevenueFrozen, &revenueFreezeNote, &handlerAdminID, &handleResult, &handledAt, &report.CreatedAt); err != nil {
		return Report{}, err
	}
	report.TargetUserID = targetUserID.Int64
	report.Content = content.String
	report.ChatMessageID = chatMessageID.Int64
	report.FileID = fileID.Int64
	report.ReviewID = reviewID.Int64
	report.CreditLogID = creditLogID.Int64
	report.RevenueRecordID = revenueRecordID.Int64
	report.RevenueFreezeNote = revenueFreezeNote.String
	report.HandlerAdminID = handlerAdminID.Int64
	report.HandleResult = handleResult.String
	decodeHandleResult(&report)
	if handledAt.Valid {
		report.HandledAt = handledAt.Time.Format(time.RFC3339)
	}
	return report, nil
}

func scanReportRows(rows *sql.Rows) (Report, error) {
	return scanReport(rows)
}

type encodedHandleResult struct {
	Result             string  `json:"result"`
	Outcome            string  `json:"outcome,omitempty"`
	RewardPoints       int     `json:"rewardPoints,omitempty"`
	CreditChange       int     `json:"creditChange,omitempty"`
	CreditTargetUserID int64   `json:"creditTargetUserId,omitempty"`
	AppealFileIDs      []int64 `json:"appealFileIds,omitempty"`
}

func encodeHandleResult(report Report) string {
	if report.HandleOutcome == "" && report.RewardPoints == 0 && report.CreditChange == 0 && report.CreditTargetUserID == 0 && len(report.AppealFileIDs) == 0 {
		return report.HandleResult
	}
	payload := encodedHandleResult{
		Result:             report.HandleResult,
		Outcome:            report.HandleOutcome,
		RewardPoints:       report.RewardPoints,
		CreditChange:       report.CreditChange,
		CreditTargetUserID: report.CreditTargetUserID,
		AppealFileIDs:      report.AppealFileIDs,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return report.HandleResult
	}
	return string(data)
}

func decodeHandleResult(report *Report) {
	value := strings.TrimSpace(report.HandleResult)
	if !strings.HasPrefix(value, "{") {
		return
	}
	var payload encodedHandleResult
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return
	}
	report.HandleResult = payload.Result
	report.HandleOutcome = payload.Outcome
	report.RewardPoints = payload.RewardPoints
	report.CreditChange = payload.CreditChange
	report.CreditTargetUserID = payload.CreditTargetUserID
	report.AppealFileIDs = payload.AppealFileIDs
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
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return sql.NullTime{Time: parsed, Valid: true}
	}
	if parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", value); err == nil {
		return sql.NullTime{Time: parsed, Valid: true}
	}
	return sql.NullTime{}
}
