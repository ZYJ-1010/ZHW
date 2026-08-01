package revenue

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

func (r *SQLRepository) SaveTemplate(ctx context.Context, template Template) (Template, error) {
	return scanTemplate(r.db.QueryRowContext(ctx, `
insert into revenue_templates (name, game_type, platform_bps, creator_bps, member_bps, status, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,now())
returning id, name, game_type, platform_bps, creator_bps, member_bps, status, created_at
`, template.Name, template.GameType, template.PlatformBps, template.CreatorBps, template.MemberBps, template.Status, template.CreatedAt))
}

func (r *SQLRepository) ListTemplates(ctx context.Context) ([]Template, error) {
	rows, err := r.db.QueryContext(ctx, templateSelect()+` order by created_at desc, id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTemplates(rows)
}

func (r *SQLRepository) FindTemplate(ctx context.Context, templateID int64) (Template, bool, error) {
	template, err := scanTemplate(r.db.QueryRowContext(ctx, templateSelect()+` where id = $1`, templateID))
	if errors.Is(err, sql.ErrNoRows) {
		return Template{}, false, nil
	}
	return template, err == nil, err
}

func (r *SQLRepository) SaveRule(ctx context.Context, rule Rule) (Rule, error) {
	return scanRule(r.db.QueryRowContext(ctx, `
insert into revenue_rules (template_id, rule_code, rule_value, created_at)
values ($1,$2,$3,$4)
on conflict (template_id, rule_code) do update set
  rule_value = excluded.rule_value
returning id, template_id, rule_code, rule_value, created_at
`, rule.TemplateID, rule.RuleCode, rule.RuleValue, rule.CreatedAt))
}

func (r *SQLRepository) ListRules(ctx context.Context, templateID int64) ([]Rule, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, template_id, rule_code, rule_value, created_at
from revenue_rules
where template_id = $1
order by id asc
`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Rule, 0)
	for rows.Next() {
		item, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *SQLRepository) SaveRecord(ctx context.Context, record Record) (Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Record{}, err
	}
	defer tx.Rollback()

	saved, err := scanRecord(tx.QueryRowContext(ctx, `
insert into revenue_records (revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,now())
returning id, revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at
`, record.RecordNo, record.GameID, record.TemplateID, record.Status, record.AmountCent, nullString(record.FrozenReason), nullTimeString(record.SettledAt), record.CreatedAt))
	if err != nil {
		return Record{}, err
	}
	if err := r.replaceRecordItems(ctx, tx, saved.ID, record.Items); err != nil {
		return Record{}, err
	}
	if err := tx.Commit(); err != nil {
		return Record{}, err
	}
	saved.Items = record.Items
	return saved, nil
}

func (r *SQLRepository) SaveRecordWithIncome(ctx context.Context, record Record) (Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Record{}, err
	}
	defer tx.Rollback()
	saved, err := scanRecord(tx.QueryRowContext(ctx, `
insert into revenue_records (revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,now())
returning id, revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at
`, record.RecordNo, record.GameID, record.TemplateID, record.Status, record.AmountCent, nullString(record.FrozenReason), nullTimeString(record.SettledAt), record.CreatedAt))
	if err != nil {
		return Record{}, err
	}
	if err = r.replaceRecordItems(ctx, tx, saved.ID, record.Items); err != nil {
		return Record{}, err
	}
	if err = syncIncomeStateTx(ctx, tx, record.Items, saved.ID, "record_generated", record.CreatedAt); err != nil {
		return Record{}, err
	}
	if err = tx.Commit(); err != nil {
		return Record{}, err
	}
	saved.Items = record.Items
	return saved, nil
}

func (r *SQLRepository) ListRecords(ctx context.Context) ([]Record, error) {
	rows, err := r.db.QueryContext(ctx, recordSelect()+` order by created_at desc, id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records, err := scanRecords(rows)
	if err != nil {
		return nil, err
	}
	return r.withItems(ctx, records)
}

func (r *SQLRepository) FindRecord(ctx context.Context, recordID int64) (Record, bool, error) {
	record, err := scanRecord(r.db.QueryRowContext(ctx, recordSelect()+` where id = $1`, recordID))
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	items, err := r.listRecordItems(ctx, record.ID)
	if err != nil {
		return Record{}, false, err
	}
	record.Items = items
	return record, true, nil
}

func (r *SQLRepository) FindRecordByGame(ctx context.Context, gameID int64) (Record, bool, error) {
	record, err := scanRecord(r.db.QueryRowContext(ctx, recordSelect()+` where game_id = $1`, gameID))
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	items, err := r.listRecordItems(ctx, record.ID)
	if err != nil {
		return Record{}, false, err
	}
	record.Items = items
	return record, true, nil
}

func (r *SQLRepository) UpdateRecord(ctx context.Context, record Record) (Record, error) {
	saved, err := scanRecord(r.db.QueryRowContext(ctx, `
update revenue_records set
  status = $2,
  frozen_reason = $3,
  settled_at = $4,
  updated_at = now()
where id = $1
returning id, revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at
`, record.ID, record.Status, nullString(record.FrozenReason), nullTimeString(record.SettledAt)))
	if err != nil {
		return Record{}, err
	}
	items, err := r.listRecordItems(ctx, saved.ID)
	if err != nil {
		return Record{}, err
	}
	saved.Items = items
	return saved, nil
}

func (r *SQLRepository) SaveSettlement(ctx context.Context, settlement Settlement) (Settlement, error) {
	return scanSettlement(r.db.QueryRowContext(ctx, `
insert into settlement_records (revenue_record_id, method, proof_no, amount_cent, created_at)
values ($1,$2,$3,$4,$5)
returning id, revenue_record_id, method, proof_no, amount_cent, created_at
`, settlement.RecordID, settlement.Method, nullString(settlement.ProofNo), settlement.AmountCent, settlement.CreatedAt))
}

// SettleRecord commits the settlement, record state, income accounts and
// income ledgers together. The row lock also prevents concurrent settlement.
func (r *SQLRepository) SettleRecord(ctx context.Context, recordID int64, settlement Settlement) (Record, Settlement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Record{}, Settlement{}, err
	}
	defer tx.Rollback()
	record, err := scanRecord(tx.QueryRowContext(ctx, recordSelect()+` where id = $1 for update`, recordID))
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, Settlement{}, ErrRecordNotFound
	}
	if err != nil {
		return Record{}, Settlement{}, err
	}
	if record.Status == "frozen" {
		return Record{}, Settlement{}, ErrRecordFrozen
	}
	if record.Status != "pending_settlement" {
		return Record{}, Settlement{}, ErrRecordNotSettleable
	}
	items, err := listRecordItemsTx(ctx, tx, recordID)
	if err != nil {
		return Record{}, Settlement{}, err
	}
	settlement.RecordID = recordID
	settlement.AmountCent = record.AmountCent
	savedSettlement, err := scanSettlement(tx.QueryRowContext(ctx, `
insert into settlement_records (revenue_record_id, method, proof_no, amount_cent, created_at)
values ($1,$2,$3,$4,$5)
returning id, revenue_record_id, method, proof_no, amount_cent, created_at
`, settlement.RecordID, settlement.Method, nullString(settlement.ProofNo), settlement.AmountCent, settlement.CreatedAt))
	if err != nil {
		return Record{}, Settlement{}, err
	}
	savedRecord, err := scanRecord(tx.QueryRowContext(ctx, `
update revenue_records
set status = 'settled', frozen_reason = null, settled_at = $2, updated_at = now()
where id = $1
returning id, revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at
`, recordID, savedSettlement.CreatedAt))
	if err != nil {
		return Record{}, Settlement{}, err
	}
	for _, userID := range affectedUserIDs(items) {
		var summary IncomeSummary
		summary.UserID = userID
		if err = tx.QueryRowContext(ctx, `
select
  coalesce(sum(i.amount_cent), 0),
  coalesce(sum(case when rr.status = 'settled' then 0 else i.amount_cent end), 0),
  coalesce(sum(case when rr.status = 'settled' then i.amount_cent else 0 end), 0)
from revenue_record_items i
join revenue_records rr on rr.id = i.revenue_record_id
where i.user_id = $1
`, userID).Scan(&summary.TotalCent, &summary.PendingCent, &summary.SettledCent); err != nil {
			return Record{}, Settlement{}, err
		}
		if _, err = tx.ExecContext(ctx, `
insert into user_income_accounts (user_id, total_cent, pending_cent, settled_cent, updated_at)
values ($1,$2,$3,$4,now())
on conflict (user_id) do update set
  total_cent = excluded.total_cent,
  pending_cent = excluded.pending_cent,
  settled_cent = excluded.settled_cent,
  updated_at = now()
`, userID, summary.TotalCent, summary.PendingCent, summary.SettledCent); err != nil {
			return Record{}, Settlement{}, err
		}
	}
	for _, item := range items {
		if item.UserID <= 0 {
			continue
		}
		if _, err = tx.ExecContext(ctx, `
insert into income_logs (user_id, revenue_record_id, change_value_cent, reason, created_at)
values ($1,$2,$3,$4,$5)
`, item.UserID, recordID, item.AmountCent, "record_settled", savedSettlement.CreatedAt); err != nil {
			return Record{}, Settlement{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Record{}, Settlement{}, err
	}
	savedRecord.Items = items
	return savedRecord, savedSettlement, nil
}

func (r *SQLRepository) TransitionRecord(ctx context.Context, recordID int64, expectedStatus string, nextStatus string, frozenReason string, logReason string, changedAt time.Time) (Record, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Record{}, false, err
	}
	defer tx.Rollback()
	record, err := scanRecord(tx.QueryRowContext(ctx, recordSelect()+` where id = $1 for update`, recordID))
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, false, ErrRecordNotFound
	}
	if err != nil {
		return Record{}, false, err
	}
	items, err := listRecordItemsTx(ctx, tx, recordID)
	if err != nil {
		return Record{}, false, err
	}
	record.Items = items
	if (expectedStatus != "" && record.Status != expectedStatus) || record.Status == nextStatus {
		return record, false, nil
	}
	settledAt := nullTimeString(record.SettledAt)
	saved, err := scanRecord(tx.QueryRowContext(ctx, `
update revenue_records
set status = $2, frozen_reason = $3, settled_at = $4, updated_at = now()
where id = $1
returning id, revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at
`, recordID, nextStatus, nullString(frozenReason), settledAt))
	if err != nil {
		return Record{}, false, err
	}
	if err = syncIncomeStateTx(ctx, tx, items, recordID, logReason, changedAt); err != nil {
		return Record{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return Record{}, false, err
	}
	saved.Items = items
	return saved, true, nil
}

func syncIncomeStateTx(ctx context.Context, tx *sql.Tx, items []Item, recordID int64, logReason string, changedAt time.Time) error {
	for _, userID := range affectedUserIDs(items) {
		var summary IncomeSummary
		if err := tx.QueryRowContext(ctx, `
select
  coalesce(sum(i.amount_cent), 0),
  coalesce(sum(case when rr.status = 'settled' then 0 else i.amount_cent end), 0),
  coalesce(sum(case when rr.status = 'settled' then i.amount_cent else 0 end), 0)
from revenue_record_items i
join revenue_records rr on rr.id = i.revenue_record_id
where i.user_id = $1
`, userID).Scan(&summary.TotalCent, &summary.PendingCent, &summary.SettledCent); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
insert into user_income_accounts (user_id, total_cent, pending_cent, settled_cent, updated_at)
values ($1,$2,$3,$4,now())
on conflict (user_id) do update set
  total_cent = excluded.total_cent,
  pending_cent = excluded.pending_cent,
  settled_cent = excluded.settled_cent,
  updated_at = now()
`, userID, summary.TotalCent, summary.PendingCent, summary.SettledCent); err != nil {
			return err
		}
	}
	for _, item := range items {
		if item.UserID <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
insert into income_logs (user_id, revenue_record_id, change_value_cent, reason, created_at)
values ($1,$2,$3,$4,$5)
`, item.UserID, recordID, item.AmountCent, logReason, changedAt); err != nil {
			return err
		}
	}
	return nil
}

func listRecordItemsTx(ctx context.Context, tx *sql.Tx, recordID int64) ([]Item, error) {
	rows, err := tx.QueryContext(ctx, `
select user_id, role, amount_cent
from revenue_record_items
where revenue_record_id = $1
order by id asc
`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		var userID sql.NullInt64
		if err := rows.Scan(&userID, &item.Role, &item.AmountCent); err != nil {
			return nil, err
		}
		item.UserID = userID.Int64
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ListSettlements(ctx context.Context) ([]Settlement, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, revenue_record_id, method, proof_no, amount_cent, created_at
from settlement_records
order by created_at desc, id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Settlement, 0)
	for rows.Next() {
		item, err := scanSettlement(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *SQLRepository) UpsertIncomeAccount(ctx context.Context, account IncomeAccount) (IncomeAccount, error) {
	return scanIncomeAccount(r.db.QueryRowContext(ctx, `
insert into user_income_accounts (user_id, total_cent, pending_cent, settled_cent, updated_at)
values ($1,$2,$3,$4,now())
on conflict (user_id) do update set
  total_cent = excluded.total_cent,
  pending_cent = excluded.pending_cent,
  settled_cent = excluded.settled_cent,
  updated_at = now()
returning id, user_id, total_cent, pending_cent, settled_cent, updated_at
`, account.UserID, account.TotalCent, account.PendingCent, account.SettledCent))
}

func (r *SQLRepository) AppendIncomeLog(ctx context.Context, log IncomeLogEntry) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	if log.Reason == "" {
		log.Reason = "revenue_record"
	}
	_, err := r.db.ExecContext(ctx, `
insert into income_logs (user_id, revenue_record_id, change_value_cent, reason, created_at)
values ($1,$2,$3,$4,$5)
`, log.UserID, nullInt64(log.RevenueRecordID), log.ChangeValueCent, log.Reason, log.CreatedAt)
	return err
}

func (r *SQLRepository) IncomeSummary(ctx context.Context, userID int64) (IncomeSummary, error) {
	var summary IncomeSummary
	summary.UserID = userID
	err := r.db.QueryRowContext(ctx, `
select
  coalesce(sum(i.amount_cent), 0),
  coalesce(sum(case when rr.status = 'settled' then 0 else i.amount_cent end), 0),
  coalesce(sum(case when rr.status = 'settled' then i.amount_cent else 0 end), 0)
from revenue_record_items i
join revenue_records rr on rr.id = i.revenue_record_id
where i.user_id = $1
`, userID).Scan(&summary.TotalCent, &summary.PendingCent, &summary.SettledCent)
	return summary, err
}

func (r *SQLRepository) IncomeLogs(ctx context.Context, userID int64, status string) ([]IncomeLog, error) {
	args := []any{userID}
	query := `
select rr.id, rr.revenue_record_no, rr.game_id, i.role, i.amount_cent, rr.status, rr.created_at, rr.settled_at
from revenue_record_items i
join revenue_records rr on rr.id = i.revenue_record_id
where i.user_id = $1
`
	if status != "" {
		query += ` and rr.status = $2`
		args = append(args, status)
	}
	query += ` order by rr.created_at desc, rr.id desc`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]IncomeLog, 0)
	for rows.Next() {
		var item IncomeLog
		var settledAt sql.NullTime
		if err := rows.Scan(&item.RecordID, &item.RecordNo, &item.GameID, &item.Role, &item.AmountCent, &item.Status, &item.CreatedAt, &settledAt); err != nil {
			return nil, err
		}
		if settledAt.Valid {
			item.SettledAt = settledAt.Time.Format(time.RFC3339)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *SQLRepository) replaceRecordItems(ctx context.Context, tx *sql.Tx, recordID int64, items []Item) error {
	if _, err := tx.ExecContext(ctx, `delete from revenue_record_items where revenue_record_id = $1`, recordID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `
insert into revenue_record_items (revenue_record_id, user_id, role, amount_cent, created_at)
values ($1,$2,$3,$4,now())
`, recordID, nullInt64(item.UserID), item.Role, item.AmountCent); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLRepository) listRecordItems(ctx context.Context, recordID int64) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, role, amount_cent
from revenue_record_items
where revenue_record_id = $1
order by id asc
`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		var userID sql.NullInt64
		if err := rows.Scan(&userID, &item.Role, &item.AmountCent); err != nil {
			return nil, err
		}
		item.UserID = userID.Int64
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) withItems(ctx context.Context, records []Record) ([]Record, error) {
	for index := range records {
		items, err := r.listRecordItems(ctx, records[index].ID)
		if err != nil {
			return nil, err
		}
		records[index].Items = items
	}
	return records, nil
}

func templateSelect() string {
	return `select id, name, game_type, platform_bps, creator_bps, member_bps, status, created_at from revenue_templates`
}

func recordSelect() string {
	return `select id, revenue_record_no, game_id, template_id, status, total_amount_cent, frozen_reason, settled_at, created_at from revenue_records`
}

func scanTemplates(rows *sql.Rows) ([]Template, error) {
	items := make([]Template, 0)
	for rows.Next() {
		item, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanTemplate(row interface {
	Scan(dest ...any) error
}) (Template, error) {
	var item Template
	err := row.Scan(&item.ID, &item.Name, &item.GameType, &item.PlatformBps, &item.CreatorBps, &item.MemberBps, &item.Status, &item.CreatedAt)
	return item, err
}

func scanRule(row interface {
	Scan(dest ...any) error
}) (Rule, error) {
	var item Rule
	err := row.Scan(&item.ID, &item.TemplateID, &item.RuleCode, &item.RuleValue, &item.CreatedAt)
	return item, err
}

func scanRecords(rows *sql.Rows) ([]Record, error) {
	items := make([]Record, 0)
	for rows.Next() {
		item, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanRecord(row interface {
	Scan(dest ...any) error
}) (Record, error) {
	var item Record
	var frozenReason sql.NullString
	var settledAt sql.NullTime
	err := row.Scan(&item.ID, &item.RecordNo, &item.GameID, &item.TemplateID, &item.Status, &item.AmountCent, &frozenReason, &settledAt, &item.CreatedAt)
	item.FrozenReason = frozenReason.String
	if settledAt.Valid {
		item.SettledAt = settledAt.Time.Format(time.RFC3339)
	}
	return item, err
}

func scanSettlement(row interface {
	Scan(dest ...any) error
}) (Settlement, error) {
	var item Settlement
	var proofNo sql.NullString
	if err := row.Scan(&item.ID, &item.RecordID, &item.Method, &proofNo, &item.AmountCent, &item.CreatedAt); err != nil {
		return Settlement{}, err
	}
	item.ProofNo = proofNo.String
	return item, nil
}

func scanIncomeAccount(row interface {
	Scan(dest ...any) error
}) (IncomeAccount, error) {
	var account IncomeAccount
	err := row.Scan(&account.ID, &account.UserID, &account.TotalCent, &account.PendingCent, &account.SettledCent, &account.UpdatedAt)
	return account, err
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
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
