package redemption

import (
	"context"
	"database/sql"
	"errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateItem(ctx context.Context, item Item) (Item, error) {
	return scanItem(r.db.QueryRowContext(ctx, `
insert into redemption_items (name, description, image_url, points_cost, stock, status, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id, name, description, image_url, points_cost, stock, status, created_at, updated_at
`, item.Name, item.Description, item.ImageURL, item.PointsCost, item.Stock, item.Status, item.CreatedAt, item.UpdatedAt))
}

func (r *SQLRepository) ListItems(ctx context.Context, includeInactive bool) ([]Item, error) {
	query := `select id, name, description, image_url, points_cost, stock, status, created_at, updated_at from redemption_items`
	if !includeInactive {
		query += ` where status = 'active'`
	}
	query += ` order by created_at desc, id desc`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (r *SQLRepository) FindItem(ctx context.Context, itemID int64) (Item, bool, error) {
	item, err := scanItem(r.db.QueryRowContext(ctx, `
select id, name, description, image_url, points_cost, stock, status, created_at, updated_at
from redemption_items
where id = $1
`, itemID))
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, false, nil
	}
	if err != nil {
		return Item{}, false, err
	}
	return item, true, nil
}

func (r *SQLRepository) UpdateItem(ctx context.Context, item Item) (Item, error) {
	return scanItem(r.db.QueryRowContext(ctx, `
update redemption_items
set name = $2, description = $3, image_url = $4, points_cost = $5, stock = $6, status = $7, updated_at = $8
where id = $1
returning id, name, description, image_url, points_cost, stock, status, created_at, updated_at
`, item.ID, item.Name, item.Description, item.ImageURL, item.PointsCost, item.Stock, item.Status, item.UpdatedAt))
}

func (r *SQLRepository) ReserveItemStock(ctx context.Context, itemID int64) (Item, error) {
	item, err := scanItem(r.db.QueryRowContext(ctx, `
update redemption_items
set stock = stock - 1, updated_at = now()
where id = $1 and status = 'active' and stock > 0
returning id, name, description, image_url, points_cost, stock, status, created_at, updated_at
`, itemID))
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, ErrInsufficientStock
	}
	return item, err
}

func (r *SQLRepository) RestoreItemStock(ctx context.Context, itemID int64) error {
	_, err := r.db.ExecContext(ctx, `
update redemption_items
set stock = stock + 1, updated_at = now()
where id = $1
`, itemID)
	return err
}

func (r *SQLRepository) NextOrderID(ctx context.Context) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `select nextval(pg_get_serial_sequence('redemption_orders','id'))`).Scan(&id)
	return id, err
}

func (r *SQLRepository) CreateOrder(ctx context.Context, order Order) (Order, error) {
	return scanOrder(r.db.QueryRowContext(ctx, `
insert into redemption_orders (id, order_no, user_id, item_id, points_cost, status, review_admin_id, review_reason, created_at, updated_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
returning id, order_no, user_id, item_id,
  (select name from redemption_items where id = redemption_orders.item_id),
  (select coalesce(image_url, '') from redemption_items where id = redemption_orders.item_id),
  points_cost, status, review_admin_id, review_reason, created_at, updated_at
`, order.ID, order.OrderNo, order.UserID, order.ItemID, order.PointsCost, order.Status, nullInt64(order.ReviewAdminID), nullString(order.ReviewReason), order.CreatedAt, order.UpdatedAt))
}

func (r *SQLRepository) ListOrdersByUser(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, orderSelect()+` where o.user_id = $1 order by o.created_at desc, o.id desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *SQLRepository) ListOrders(ctx context.Context) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, orderSelect()+` order by o.created_at desc, o.id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *SQLRepository) FindOrder(ctx context.Context, orderID int64) (Order, bool, error) {
	order, err := scanOrder(r.db.QueryRowContext(ctx, orderSelect()+` where o.id = $1`, orderID))
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, false, nil
	}
	if err != nil {
		return Order{}, false, err
	}
	return order, true, nil
}

func (r *SQLRepository) UpdateOrder(ctx context.Context, order Order) (Order, error) {
	return scanOrder(r.db.QueryRowContext(ctx, `
update redemption_orders
set status = $2, review_admin_id = $3, review_reason = $4, updated_at = $5
where id = $1
returning id, order_no, user_id, item_id,
  (select name from redemption_items where id = redemption_orders.item_id),
  (select coalesce(image_url, '') from redemption_items where id = redemption_orders.item_id),
  points_cost, status, review_admin_id, review_reason, created_at, updated_at
`, order.ID, order.Status, nullInt64(order.ReviewAdminID), nullString(order.ReviewReason), order.UpdatedAt))
}

func orderSelect() string {
	return `
select o.id, o.order_no, o.user_id, o.item_id, coalesce(i.name, ''), coalesce(i.image_url, ''), o.points_cost, o.status,
  o.review_admin_id, o.review_reason, o.created_at, o.updated_at
from redemption_orders o
left join redemption_items i on i.id = o.item_id`
}

func scanItems(rows *sql.Rows) ([]Item, error) {
	items := make([]Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanItem(row interface {
	Scan(dest ...any) error
}) (Item, error) {
	var item Item
	var description sql.NullString
	var imageURL sql.NullString
	if err := row.Scan(&item.ID, &item.Name, &description, &imageURL, &item.PointsCost, &item.Stock, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Item{}, err
	}
	item.Description = description.String
	item.ImageURL = imageURL.String
	return item, nil
}

func scanOrders(rows *sql.Rows) ([]Order, error) {
	items := make([]Order, 0)
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanOrder(row interface {
	Scan(dest ...any) error
}) (Order, error) {
	var order Order
	var reviewAdminID sql.NullInt64
	var reviewReason sql.NullString
	if err := row.Scan(&order.ID, &order.OrderNo, &order.UserID, &order.ItemID, &order.ItemName, &order.ItemImageURL, &order.PointsCost, &order.Status, &reviewAdminID, &reviewReason, &order.CreatedAt, &order.UpdatedAt); err != nil {
		return Order{}, err
	}
	order.ReviewAdminID = reviewAdminID.Int64
	order.ReviewReason = reviewReason.String
	return order, nil
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
