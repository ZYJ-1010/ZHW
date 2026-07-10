package orders

import (
	"context"
	"testing"
	"time"
)

func TestEnsureFreeNoPayOrderIsIdempotent(t *testing.T) {
	service := NewService()

	first, err := service.EnsureFreeNoPayOrder(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.EnsureFreeNoPayOrder(1, 10)
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != second.ID || first.OrderNo != second.OrderNo {
		t.Fatalf("expected same placeholder order, got first=%+v second=%+v", first, second)
	}
	if first.PayStatus != "free_no_pay" || first.NeedWechatPay {
		t.Fatalf("expected free_no_pay without wechat pay, got %+v", first)
	}
}

func TestGuideFeePlaceholderOrderIsIdempotent(t *testing.T) {
	service := NewService()

	first, err := service.EnsureGuideFeePlaceholderOrder(7)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.EnsureGuideFeePlaceholderOrder(7)
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != second.ID || first.OrderNo != second.OrderNo {
		t.Fatalf("expected same guide fee placeholder order, got first=%+v second=%+v", first, second)
	}
	if first.PayStatus != "guide_fee_placeholder" || first.NeedWechatPay || first.GameID != 0 {
		t.Fatalf("expected guide fee placeholder without wechat pay, got %+v", first)
	}
}

func TestGetForUserRejectsOtherUser(t *testing.T) {
	service := NewService()
	order, err := service.EnsureFreeNoPayOrder(1, 10)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.GetForUser(2, order.ID); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if got, err := service.GetForUser(1, order.ID); err != nil || got.ID != order.ID {
		t.Fatalf("expected owner can read order, got order=%+v err=%v", got, err)
	}
}

func TestCallbackPlaceholderIsIdempotent(t *testing.T) {
	service := NewService()
	order, err := service.EnsureFreeNoPayOrder(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	req := CallbackRequest{OrderNo: order.OrderNo, EventID: "evt-1", Payload: map[string]interface{}{"status": "free_no_pay"}}

	first, err := service.RecordCallback(req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.RecordCallback(req)
	if err != nil {
		t.Fatal(err)
	}

	if first.IdempotentHit || !second.IdempotentHit {
		t.Fatalf("expected second callback idempotent hit, first=%+v second=%+v", first, second)
	}
}

func TestProfitSharingPlaceholderLifecycle(t *testing.T) {
	service := NewService()
	order, err := service.EnsureFreeNoPayOrder(1, 10)
	if err != nil {
		t.Fatal(err)
	}

	share, err := service.CreateProfitSharingPlaceholder(ProfitSharingOrderRequest{OutOrderNo: "PS-1", OrderNo: order.OrderNo, AmountCent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if share.Mode != "profit_sharing_placeholder" || share.NeedWechatPay || !share.Placeholder {
		t.Fatalf("unexpected share placeholder: %+v", share)
	}
	again, err := service.CreateProfitSharingPlaceholder(ProfitSharingOrderRequest{OutOrderNo: "PS-1", OrderNo: order.OrderNo, AmountCent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if again.OutOrderNo != share.OutOrderNo {
		t.Fatalf("expected idempotent share placeholder, got %+v %+v", share, again)
	}
	got, err := service.ProfitSharingPlaceholder("PS-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "share_placeholder" || got.LastQueriedAt.IsZero() {
		t.Fatalf("unexpected profit sharing query: %+v", got)
	}
	ret, err := service.CreateProfitSharingReturnPlaceholder(ProfitSharingReturnRequest{OutReturnNo: "PR-1", OutOrderNo: "PS-1", Reason: "refund"})
	if err != nil {
		t.Fatal(err)
	}
	if ret.Mode != "profit_sharing_return_placeholder" || ret.NeedWechatPay || !ret.Placeholder {
		t.Fatalf("unexpected return placeholder: %+v", ret)
	}
}

func TestServiceUsesRepositoryWhenConfigured(t *testing.T) {
	repo := &fakeRepository{
		order: Order{
			ID:            9,
			OrderNo:       "FREE-10",
			UserID:        1,
			GameID:        10,
			PayStatus:     "free_no_pay",
			PayChannel:    "none",
			NeedWechatPay: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	service := NewServiceWithRepository(repo)

	order, err := service.EnsureFreeNoPayOrder(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != 9 || !repo.ensureCalled {
		t.Fatalf("expected repository order, got order=%+v repo=%+v", order, repo)
	}
	got, err := service.GetForUser(1, 9)
	if err != nil {
		t.Fatal(err)
	}
	if got.OrderNo != "FREE-10" || !repo.findForUserCalled {
		t.Fatalf("expected repository get, got order=%+v repo=%+v", got, repo)
	}
	result, err := service.RecordCallback(CallbackRequest{OrderNo: "FREE-10", EventID: "evt-1", Payload: map[string]interface{}{"trade_state": "SUCCESS"}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IdempotentHit || !repo.findByNoCalled || !repo.callbackCalled {
		t.Fatalf("expected repository callback idempotent hit, got result=%+v repo=%+v", result, repo)
	}
}

type fakeRepository struct {
	order             Order
	ensureCalled      bool
	findForUserCalled bool
	findByNoCalled    bool
	callbackCalled    bool
}

func (f *fakeRepository) EnsureFreeNoPayOrder(_ context.Context, _ int64, _ int64) (Order, error) {
	f.ensureCalled = true
	return f.order, nil
}

func (f *fakeRepository) EnsureGuideFeePlaceholderOrder(_ context.Context, _ int64) (Order, error) {
	f.ensureCalled = true
	return f.order, nil
}

func (f *fakeRepository) FindOrderForUser(_ context.Context, _ int64, _ int64) (Order, bool, error) {
	f.findForUserCalled = true
	return f.order, true, nil
}

func (f *fakeRepository) FindOrderByNo(_ context.Context, _ string) (Order, bool, error) {
	f.findByNoCalled = true
	return f.order, true, nil
}

func (f *fakeRepository) SaveCallback(_ context.Context, _ CallbackRequest, _ string) (bool, error) {
	f.callbackCalled = true
	return true, nil
}
