package redemption

import (
	"context"
	"testing"

	"zhw-mini/services/go-api/internal/points"
)

func TestRepositoryBackedRedemptionOrderAndRejectRefund(t *testing.T) {
	ledger := points.NewService()
	service := NewServiceWithRepository(ledger, newFakeRedemptionRepository())
	if _, _, err := ledger.Grant(1, 30, "seed", 0, "seed"); err != nil {
		t.Fatal(err)
	}
	item, err := service.CreateItem(CreateItemRequest{Name: "coupon", PointsCost: 10, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}

	order, err := service.CreateOrder(1, CreateOrderRequest{ItemID: item.ID})
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != "pending" || ledger.Summary(1).AvailablePoints != 20 {
		t.Fatalf("expected pending order and deducted points, order=%+v points=%+v", order, ledger.Summary(1))
	}
	if items := service.AdminItems(); len(items) != 1 || items[0].Stock != 0 {
		t.Fatalf("expected reserved stock, got %+v", items)
	}

	order, err = service.ReviewOrder(order.ID, 99, ReviewOrderRequest{Approve: false, Reason: "invalid"})
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != "rejected" || order.ReviewAdminID != 99 || ledger.Summary(1).AvailablePoints != 30 {
		t.Fatalf("expected rejected order refund, order=%+v points=%+v", order, ledger.Summary(1))
	}
	if items := service.AdminItems(); len(items) != 1 || items[0].Stock != 1 {
		t.Fatalf("expected rejected order to restore stock, got %+v", items)
	}
}

func TestRepositoryBackedRedemptionCancelRefund(t *testing.T) {
	ledger := points.NewService()
	service := NewServiceWithRepository(ledger, newFakeRedemptionRepository())
	if _, _, err := ledger.Grant(1, 30, "seed", 0, "seed"); err != nil {
		t.Fatal(err)
	}
	item, err := service.CreateItem(CreateItemRequest{Name: "coupon", PointsCost: 10, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}

	order, err := service.CreateOrder(1, CreateOrderRequest{ItemID: item.ID})
	if err != nil {
		t.Fatal(err)
	}
	order, err = service.CancelOrder(1, order.ID, "change mind")
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != "canceled" || ledger.Summary(1).AvailablePoints != 30 {
		t.Fatalf("expected canceled order refund, order=%+v points=%+v", order, ledger.Summary(1))
	}
	if _, err := service.CancelOrder(1, order.ID, "again"); err != ErrInvalidStatus {
		t.Fatalf("expected invalid status on repeated cancel, got %v", err)
	}
	if items := service.AdminItems(); len(items) != 1 || items[0].Stock != 1 {
		t.Fatalf("expected canceled order to restore stock, got %+v", items)
	}
}

type fakeRedemptionRepository struct {
	nextItemID  int64
	nextOrderID int64
	items       map[int64]Item
	orders      map[int64]Order
}

func newFakeRedemptionRepository() *fakeRedemptionRepository {
	return &fakeRedemptionRepository{
		nextItemID:  1,
		nextOrderID: 1,
		items:       make(map[int64]Item),
		orders:      make(map[int64]Order),
	}
}

func (r *fakeRedemptionRepository) CreateItem(ctx context.Context, item Item) (Item, error) {
	item.ID = r.nextItemID
	r.nextItemID++
	r.items[item.ID] = item
	return item, nil
}

func (r *fakeRedemptionRepository) ListItems(ctx context.Context, includeInactive bool) ([]Item, error) {
	result := make([]Item, 0)
	for _, item := range r.items {
		if includeInactive || item.Status == "active" {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRedemptionRepository) FindItem(ctx context.Context, itemID int64) (Item, bool, error) {
	item, ok := r.items[itemID]
	return item, ok, nil
}

func (r *fakeRedemptionRepository) UpdateItem(ctx context.Context, item Item) (Item, error) {
	r.items[item.ID] = item
	return item, nil
}

func (r *fakeRedemptionRepository) ReserveItemStock(ctx context.Context, itemID int64) (Item, error) {
	item, ok := r.items[itemID]
	if !ok || item.Stock <= 0 || item.Status != "active" {
		return Item{}, ErrInsufficientStock
	}
	item.Stock--
	r.items[item.ID] = item
	return item, nil
}

func (r *fakeRedemptionRepository) RestoreItemStock(ctx context.Context, itemID int64) error {
	item := r.items[itemID]
	item.Stock++
	r.items[item.ID] = item
	return nil
}

func (r *fakeRedemptionRepository) NextOrderID(ctx context.Context) (int64, error) {
	id := r.nextOrderID
	r.nextOrderID++
	return id, nil
}

func (r *fakeRedemptionRepository) CreateOrder(ctx context.Context, order Order) (Order, error) {
	if item, ok := r.items[order.ItemID]; ok {
		order.ItemName = item.Name
	}
	r.orders[order.ID] = order
	return order, nil
}

func (r *fakeRedemptionRepository) ListOrdersByUser(ctx context.Context, userID int64) ([]Order, error) {
	result := make([]Order, 0)
	for _, order := range r.orders {
		if order.UserID == userID {
			result = append(result, order)
		}
	}
	return result, nil
}

func (r *fakeRedemptionRepository) ListOrders(ctx context.Context) ([]Order, error) {
	result := make([]Order, 0, len(r.orders))
	for _, order := range r.orders {
		result = append(result, order)
	}
	return result, nil
}

func (r *fakeRedemptionRepository) FindOrder(ctx context.Context, orderID int64) (Order, bool, error) {
	order, ok := r.orders[orderID]
	return order, ok, nil
}

func (r *fakeRedemptionRepository) UpdateOrder(ctx context.Context, order Order) (Order, error) {
	r.orders[order.ID] = order
	return order, nil
}
