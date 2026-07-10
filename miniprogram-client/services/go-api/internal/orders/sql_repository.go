package orders

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) EnsureFreeNoPayOrder(ctx context.Context, userID int64, gameID int64) (Order, error) {
	orderNo := "FREE-" + int64String(gameID)
	return r.scanOrder(r.db.QueryRowContext(ctx, `
insert into payment_orders (order_no, user_id, game_id, amount_cent, pay_status)
values ($1,$2,$3,0,'free_no_pay')
on conflict (order_no) do update set
  updated_at = payment_orders.updated_at
returning id, order_no, user_id, game_id, amount_cent, pay_status, created_at, updated_at
`, orderNo, userID, gameID))
}

func (r *SQLRepository) EnsureGuideFeePlaceholderOrder(ctx context.Context, userID int64) (Order, error) {
	orderNo := "GUIDE-FEE-" + int64String(userID)
	return r.scanOrder(r.db.QueryRowContext(ctx, `
insert into payment_orders (order_no, user_id, game_id, amount_cent, pay_status)
values ($1,$2,null,0,'guide_fee_placeholder')
on conflict (order_no) do update set
  updated_at = payment_orders.updated_at
returning id, order_no, user_id, game_id, amount_cent, pay_status, created_at, updated_at
`, orderNo, userID))
}

func (r *SQLRepository) FindOrderForUser(ctx context.Context, userID int64, orderID int64) (Order, bool, error) {
	order, err := r.scanOrder(r.db.QueryRowContext(ctx, `
select id, order_no, user_id, game_id, amount_cent, pay_status, created_at, updated_at
from payment_orders
where id = $1 and user_id = $2
`, orderID, userID))
	if err == sql.ErrNoRows {
		return Order{}, false, nil
	}
	if err != nil {
		return Order{}, false, err
	}
	return order, true, nil
}

func (r *SQLRepository) FindOrderByNo(ctx context.Context, orderNo string) (Order, bool, error) {
	order, err := r.scanOrder(r.db.QueryRowContext(ctx, `
select id, order_no, user_id, game_id, amount_cent, pay_status, created_at, updated_at
from payment_orders
where order_no = $1
`, orderNo))
	if err == sql.ErrNoRows {
		return Order{}, false, nil
	}
	if err != nil {
		return Order{}, false, err
	}
	return order, true, nil
}

func (r *SQLRepository) SaveCallback(ctx context.Context, req CallbackRequest, payloadDigest string) (bool, error) {
	rawBody, _ := json.Marshal(req.Payload)
	var inserted bool
	err := r.db.QueryRowContext(ctx, `
with inserted as (
  insert into payment_callbacks (
    provider, callback_type, event_id, order_no, raw_body, payload_digest,
    verify_status, process_status, processed_at
  ) values ('wechat_pay','payment_notify',$1,$2,$3,$4,'placeholder_verified','received',now())
  on conflict (order_no, event_id) where order_no is not null and event_id is not null do nothing
  returning id
)
select exists(select 1 from inserted)
`, req.EventID, req.OrderNo, string(rawBody), payloadDigest).Scan(&inserted)
	if err != nil {
		return false, err
	}
	return !inserted, nil
}

func (r *SQLRepository) scanOrder(row *sql.Row) (Order, error) {
	var order Order
	var gameID sql.NullInt64
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(&order.ID, &order.OrderNo, &order.UserID, &gameID, &order.AmountCent, &order.PayStatus, &createdAt, &updatedAt); err != nil {
		return Order{}, err
	}
	order.GameID = gameID.Int64
	order.PayChannel = "none"
	order.NeedWechatPay = order.PayStatus != "free_no_pay" && order.PayStatus != "guide_fee_placeholder"
	order.CreatedAt = createdAt
	order.UpdatedAt = updatedAt
	return order, nil
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}
