package redemption

import (
	"sync"
	"testing"

	"zhw-mini/services/go-api/internal/points"
)

func TestCreateOrderRejectsInsufficientPointsAndRestoresStock(t *testing.T) {
	ledger := points.NewService()
	service := NewService(ledger)
	item, err := service.CreateItem(CreateItemRequest{Name: "coupon", PointsCost: 100, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.CreateOrder(1, CreateOrderRequest{ItemID: item.ID}); err != ErrInsufficientPoints {
		t.Fatalf("expected ErrInsufficientPoints, got %v", err)
	}
	items := service.AdminItems()
	if len(items) != 1 || items[0].Stock != 1 {
		t.Fatalf("expected stock restored after insufficient points, got %+v", items)
	}
}

func TestCreateOrderRejectsInsufficientStock(t *testing.T) {
	ledger := points.NewService()
	service := NewService(ledger)
	item, err := service.CreateItem(CreateItemRequest{Name: "coupon", PointsCost: 10, Stock: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = ledger.Grant(1, 100, "seed", 1, "seed"); err != nil {
		t.Fatal(err)
	}

	if _, err = service.CreateOrder(1, CreateOrderRequest{ItemID: item.ID}); err != ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestConcurrentCreateOrderDoesNotOversell(t *testing.T) {
	ledger := points.NewService()
	service := NewService(ledger)
	item, err := service.CreateItem(CreateItemRequest{Name: "limited coupon", PointsCost: 10, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}
	for userID := int64(1); userID <= 2; userID++ {
		if _, _, err = ledger.Grant(userID, 100, "seed", userID, "seed"); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	successes := 0
	var mu sync.Mutex
	for userID := int64(1); userID <= 2; userID++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			if _, err := service.CreateOrder(userID, CreateOrderRequest{ItemID: item.ID}); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}(userID)
	}
	wg.Wait()

	if successes != 1 || len(service.AdminOrders()) != 1 || service.AdminItems()[0].Stock != 0 {
		t.Fatalf("expected exactly one successful order, successes=%d orders=%+v items=%+v", successes, service.AdminOrders(), service.AdminItems())
	}
}

func TestRejectedOrderRefundsPoints(t *testing.T) {
	ledger := points.NewService()
	service := NewService(ledger)
	item, err := service.CreateItem(CreateItemRequest{Name: "coupon", PointsCost: 30, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = ledger.Grant(1, 100, "seed", 1, "seed"); err != nil {
		t.Fatal(err)
	}
	order, err := service.CreateOrder(1, CreateOrderRequest{ItemID: item.ID})
	if err != nil {
		t.Fatal(err)
	}
	if ledger.Summary(1).AvailablePoints != 70 {
		t.Fatalf("expected points deducted, got %+v", ledger.Summary(1))
	}
	if _, err = service.ReviewOrder(order.ID, 99, ReviewOrderRequest{Approve: false, Reason: "invalid"}); err != nil {
		t.Fatal(err)
	}
	if ledger.Summary(1).AvailablePoints != 100 {
		t.Fatalf("expected rejected order refund, got %+v", ledger.Summary(1))
	}
}

func TestReviewOrderRequiresReason(t *testing.T) {
	ledger := points.NewService()
	service := NewService(ledger)
	item, err := service.CreateItem(CreateItemRequest{Name: "coupon", PointsCost: 10, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = ledger.Grant(1, 100, "seed", 1, "seed"); err != nil {
		t.Fatal(err)
	}
	order, err := service.CreateOrder(1, CreateOrderRequest{ItemID: item.ID})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = service.ReviewOrder(order.ID, 99, ReviewOrderRequest{Approve: true}); err != ErrInvalidOrder {
		t.Fatalf("expected ErrInvalidOrder when reason is missing, got %v", err)
	}
	if orders := service.AdminOrders(); len(orders) != 1 || orders[0].Status != "pending" {
		t.Fatalf("expected order to stay pending, got %+v", orders)
	}
}
