package exports

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type SQLRepository struct{ db *sql.DB }

func NewSQLRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) CreateTask(ctx context.Context, task Task) (Task, error) {
	filters, _ := json.Marshal(task.Filters)
	return scanTask(r.db.QueryRowContext(ctx, `
insert into export_tasks (task_no, template_code, export_type, status, created_by, filters_json, created_at)
values ($1,$2,$3,$4,$5,$6,$7)
returning id, task_no, template_code, export_type, status, file_id, created_by,
  filters_json, fail_reason, created_at, finished_at
`, task.TaskNo, task.TemplateCode, task.ExportType, task.Status, nullInt64(task.CreatedBy), nullJSON(filters), task.CreatedAt))
}

func (r *SQLRepository) ListTasks(ctx context.Context) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, task_no, template_code, export_type, status, file_id, created_by,
  filters_json, fail_reason, created_at, finished_at
from export_tasks order by created_at desc, id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Task, 0)
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetTask(ctx context.Context, taskID int64) (Task, bool, error) {
	item, err := scanTask(r.db.QueryRowContext(ctx, `
select id, task_no, template_code, export_type, status, file_id, created_by,
  filters_json, fail_reason, created_at, finished_at
from export_tasks where id = $1
`, taskID))
	if err == sql.ErrNoRows {
		return Task{}, false, nil
	}
	return item, err == nil, err
}

func (r *SQLRepository) UpdateTask(ctx context.Context, task Task) (Task, error) {
	filters, _ := json.Marshal(task.Filters)
	return scanTask(r.db.QueryRowContext(ctx, `
update export_tasks set status=$2, file_id=$3, filters_json=$4, fail_reason=$5, finished_at=$6
where id=$1
returning id, task_no, template_code, export_type, status, file_id, created_by,
  filters_json, fail_reason, created_at, finished_at
`, task.ID, task.Status, nullInt64(task.FileID), nullJSON(filters), nullString(task.FailReason), nullTimeString(task.FinishedAt)))
}

func (r *SQLRepository) SaveContent(ctx context.Context, fileID int64, content []byte) error {
	_, err := r.db.ExecContext(ctx, `
insert into export_file_contents (file_id, content, created_at)
values ($1,$2,now())
on conflict (file_id) do update set content=excluded.content, created_at=now()
`, fileID, content)
	return err
}

func (r *SQLRepository) LoadContent(ctx context.Context, fileID int64) ([]byte, bool, error) {
	var content []byte
	err := r.db.QueryRowContext(ctx, `select content from export_file_contents where file_id=$1`, fileID).Scan(&content)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	return content, err == nil, err
}

func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var item Task
	var fileID, createdBy sql.NullInt64
	var filters []byte
	var failReason sql.NullString
	var finishedAt sql.NullTime
	err := row.Scan(&item.ID, &item.TaskNo, &item.TemplateCode, &item.ExportType, &item.Status,
		&fileID, &createdBy, &filters, &failReason, &item.CreatedAt, &finishedAt)
	if err != nil {
		return Task{}, err
	}
	item.FileID = fileID.Int64
	item.CreatedBy = createdBy.Int64
	item.FailReason = failReason.String
	if len(filters) > 0 {
		_ = json.Unmarshal(filters, &item.Filters)
	}
	if finishedAt.Valid {
		item.FinishedAt = finishedAt.Time.Format(time.RFC3339)
	}
	return item, nil
}

func nullInt64(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullJSON(value []byte) interface{} {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}

func nullTimeString(value string) interface{} {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return parsed
}
