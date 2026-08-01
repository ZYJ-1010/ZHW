package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) SaveNotification(ctx context.Context, notification Notification, task *WechatTask) (Notification, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Notification{}, err
	}
	defer tx.Rollback()

	saved, err := scanNotification(tx.QueryRowContext(ctx, `
insert into notifications (
  user_id, notify_type, title, content, biz_type, biz_id, status,
  need_wechat, wechat_state, wechat_template_id, created_at, read_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
returning id, user_id, notify_type, title, content, biz_type, biz_id, status,
  need_wechat, wechat_state, wechat_template_id, wechat_task_id, created_at, read_at
`, notification.UserID, notification.NotifyType, notification.Title, nullString(notification.Content), nullString(notification.BizType), nullInt64(notification.BizID), notification.Status, notification.NeedWechat, nullString(notification.WechatState), nullString(notification.WechatTemplateID), notification.CreatedAt, nullTimeString(notification.ReadAt)))
	if err != nil {
		return Notification{}, err
	}

	if task != nil {
		payload, _ := json.Marshal(task.RequestPayload)
		var taskID int64
		err = tx.QueryRowContext(ctx, `
insert into wechat_subscribe_tasks (
  notification_id, user_id, scene, template_id, status, request_payload, created_at, sent_at
) values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id
`, saved.ID, task.UserID, task.Scene, task.TemplateID, task.Status, nullJSON(payload), task.CreatedAt, nullTimeString(task.SentAt)).Scan(&taskID)
		if err != nil {
			return Notification{}, err
		}
		saved, err = scanNotification(tx.QueryRowContext(ctx, `
update notifications
set wechat_task_id = $2, wechat_state = 'pending'
where id = $1
returning id, user_id, notify_type, title, content, biz_type, biz_id, status,
  need_wechat, wechat_state, wechat_template_id, wechat_task_id, created_at, read_at
`, saved.ID, taskID))
		if err != nil {
			return Notification{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Notification{}, err
	}
	return saved, nil
}

func (r *SQLRepository) ListNotifications(ctx context.Context, userID int64) ([]Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, user_id, notify_type, title, content, biz_type, biz_id, status,
  need_wechat, wechat_state, wechat_template_id, wechat_task_id, created_at, read_at
from notifications
where user_id = $1
order by created_at desc, id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func (r *SQLRepository) FindNotification(ctx context.Context, notificationID int64) (Notification, bool, error) {
	notification, err := scanNotification(r.db.QueryRowContext(ctx, `
select id, user_id, notify_type, title, content, biz_type, biz_id, status,
  need_wechat, wechat_state, wechat_template_id, wechat_task_id, created_at, read_at
from notifications
where id = $1
`, notificationID))
	if err == sql.ErrNoRows {
		return Notification{}, false, nil
	}
	if err != nil {
		return Notification{}, false, err
	}
	return notification, true, nil
}

func (r *SQLRepository) UpdateNotification(ctx context.Context, notification Notification) (Notification, error) {
	return scanNotification(r.db.QueryRowContext(ctx, `
update notifications set
	notify_type = $2,
	title = $3,
	content = $4,
	biz_type = $5,
	biz_id = $6,
	status = $7,
	need_wechat = $8,
	wechat_state = $9,
	wechat_template_id = $10,
	wechat_task_id = $11,
	created_at = $12,
	read_at = $13
where id = $1
returning id, user_id, notify_type, title, content, biz_type, biz_id, status,
  need_wechat, wechat_state, wechat_template_id, wechat_task_id, created_at, read_at
`, notification.ID, notification.NotifyType, notification.Title, nullString(notification.Content), nullString(notification.BizType), nullInt64(notification.BizID), notification.Status, notification.NeedWechat, nullString(notification.WechatState), nullString(notification.WechatTemplateID), nullInt64(notification.WechatTaskID), notification.CreatedAt, nullTimeString(notification.ReadAt)))
}

func (r *SQLRepository) ListWechatTasks(ctx context.Context) ([]WechatTask, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, notification_id, user_id, scene, template_id, status, request_payload,
  result_code, result_message, created_at, sent_at
from wechat_subscribe_tasks
order by created_at desc, id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWechatTasks(rows)
}

func (r *SQLRepository) ListWechatTemplates(ctx context.Context) ([]WechatTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `
select scene, template_id, title, status
from wechat_subscribe_templates
order by id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]WechatTemplate, 0)
	for rows.Next() {
		var item WechatTemplate
		if err := rows.Scan(&item.Scene, &item.TemplateID, &item.Title, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) MarkWechatTaskSent(ctx context.Context, taskID int64, resultCode string, resultMessage string, sentAt string) (WechatTask, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return WechatTask{}, err
	}
	defer tx.Rollback()
	task, err := scanWechatTask(tx.QueryRowContext(ctx, `
update wechat_subscribe_tasks set
  status = 'sent',
  result_code = $2,
  result_message = $3,
  sent_at = $4
where id = $1
returning id, notification_id, user_id, scene, template_id, status, request_payload,
  result_code, result_message, created_at, sent_at
`, taskID, nullString(resultCode), nullString(resultMessage), nullTimeString(sentAt)))
	if err == sql.ErrNoRows {
		return WechatTask{}, ErrNotificationNotFound
	}
	if err != nil {
		return WechatTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `update notifications set wechat_state = 'sent' where id = $1`, task.NotificationID); err != nil {
		return WechatTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return WechatTask{}, err
	}
	return task, nil
}

func scanNotifications(rows *sql.Rows) ([]Notification, error) {
	items := make([]Notification, 0)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanNotification(row interface {
	Scan(dest ...any) error
}) (Notification, error) {
	var item Notification
	var content sql.NullString
	var bizType sql.NullString
	var bizID sql.NullInt64
	var wechatState sql.NullString
	var wechatTemplateID sql.NullString
	var wechatTaskID sql.NullInt64
	var readAt sql.NullTime
	if err := row.Scan(&item.ID, &item.UserID, &item.NotifyType, &item.Title, &content, &bizType, &bizID, &item.Status, &item.NeedWechat, &wechatState, &wechatTemplateID, &wechatTaskID, &item.CreatedAt, &readAt); err != nil {
		return Notification{}, err
	}
	item.Content = content.String
	item.BizType = bizType.String
	item.BizID = bizID.Int64
	item.WechatState = wechatState.String
	item.WechatTemplateID = wechatTemplateID.String
	item.WechatTaskID = wechatTaskID.Int64
	if readAt.Valid {
		item.ReadAt = readAt.Time.Format(time.RFC3339)
	}
	return item, nil
}

func scanWechatTasks(rows *sql.Rows) ([]WechatTask, error) {
	items := make([]WechatTask, 0)
	for rows.Next() {
		item, err := scanWechatTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanWechatTask(row interface {
	Scan(dest ...any) error
}) (WechatTask, error) {
	var item WechatTask
	var payload []byte
	var resultCode sql.NullString
	var resultMessage sql.NullString
	var sentAt sql.NullTime
	if err := row.Scan(&item.ID, &item.NotificationID, &item.UserID, &item.Scene, &item.TemplateID, &item.Status, &payload, &resultCode, &resultMessage, &item.CreatedAt, &sentAt); err != nil {
		return WechatTask{}, err
	}
	_ = json.Unmarshal(payload, &item.RequestPayload)
	item.ResultCode = resultCode.String
	item.ResultMessage = resultMessage.String
	if sentAt.Valid {
		item.SentAt = sentAt.Time.Format(time.RFC3339)
	}
	return item, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}

func nullJSON(value []byte) any {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
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
