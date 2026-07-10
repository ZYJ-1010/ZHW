package orders

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrOrderNotFound   = errors.New("payment order not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidOrder    = errors.New("invalid payment order")
	ErrCallbackInvalid = errors.New("invalid payment callback")
)

type Order struct {
	ID            int64     `json:"id"`
	OrderNo       string    `json:"orderNo"`
	UserID        int64     `json:"userId"`
	GameID        int64     `json:"gameId"`
	AmountCent    int64     `json:"amountCent"`
	PayStatus     string    `json:"payStatus"`
	PayChannel    string    `json:"payChannel"`
	NeedWechatPay bool      `json:"needWechatPay"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type PrecreateResult struct {
	Order         Order  `json:"order"`
	OrderNo       string `json:"orderNo"`
	PayStatus     string `json:"payStatus"`
	AmountCent    int64  `json:"amountCent"`
	NeedWechatPay bool   `json:"needWechatPay"`
	Mode          string `json:"mode"`
}

type CallbackRequest struct {
	OrderNo string                 `json:"orderNo"`
	EventID string                 `json:"eventId"`
	Payload map[string]interface{} `json:"payload"`
}

type CallbackResult struct {
	OrderNo       string `json:"orderNo"`
	EventID       string `json:"eventId"`
	Received      bool   `json:"received"`
	Verified      bool   `json:"verified"`
	IdempotentHit bool   `json:"idempotentHit"`
	Mode          string `json:"mode"`
}

type ProfitSharingOrderRequest struct {
	OutOrderNo string `json:"outOrderNo"`
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
}

type ProfitSharingOrder struct {
	OutOrderNo    string    `json:"outOrderNo"`
	OrderNo       string    `json:"orderNo"`
	AmountCent    int64     `json:"amountCent"`
	Status        string    `json:"status"`
	NeedWechatPay bool      `json:"needWechatPay"`
	Placeholder   bool      `json:"placeholder"`
	Mode          string    `json:"mode"`
	CreatedAt     time.Time `json:"createdAt"`
	LastQueriedAt time.Time `json:"lastQueriedAt,omitempty"`
}

type ProfitSharingReturnRequest struct {
	OutReturnNo string `json:"outReturnNo"`
	OutOrderNo  string `json:"outOrderNo"`
	Reason      string `json:"reason"`
}

type ProfitSharingReturn struct {
	OutReturnNo   string    `json:"outReturnNo"`
	OutOrderNo    string    `json:"outOrderNo"`
	Status        string    `json:"status"`
	NeedWechatPay bool      `json:"needWechatPay"`
	Placeholder   bool      `json:"placeholder"`
	Mode          string    `json:"mode"`
	Reason        string    `json:"reason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type callbackRecord struct {
	eventID       string
	orderNo       string
	payloadDigest string
	createdAt     time.Time
}

type Repository interface {
	EnsureFreeNoPayOrder(ctx context.Context, userID int64, gameID int64) (Order, error)
	EnsureGuideFeePlaceholderOrder(ctx context.Context, userID int64) (Order, error)
	FindOrderForUser(ctx context.Context, userID int64, orderID int64) (Order, bool, error)
	FindOrderByNo(ctx context.Context, orderNo string) (Order, bool, error)
	SaveCallback(ctx context.Context, req CallbackRequest, payloadDigest string) (idempotentHit bool, err error)
}

type Service struct {
	mu             sync.RWMutex
	nextID         int64
	orders         map[int64]Order
	ordersByNo     map[string]int64
	ordersByGame   map[int64]int64
	callbackEvents map[string]callbackRecord
	shareOrders    map[string]ProfitSharingOrder
	shareReturns   map[string]ProfitSharingReturn
	repo           Repository
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{
		nextID:         1,
		orders:         make(map[int64]Order),
		ordersByNo:     make(map[string]int64),
		ordersByGame:   make(map[int64]int64),
		callbackEvents: make(map[string]callbackRecord),
		shareOrders:    make(map[string]ProfitSharingOrder),
		shareReturns:   make(map[string]ProfitSharingReturn),
		repo:           repo,
	}
}

func (s *Service) EnsureFreeNoPayOrder(userID int64, gameID int64) (Order, error) {
	if userID <= 0 || gameID <= 0 {
		return Order{}, ErrInvalidOrder
	}
	if s.repo != nil {
		return s.repo.EnsureFreeNoPayOrder(context.Background(), userID, gameID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if orderID, ok := s.ordersByGame[gameID]; ok {
		return s.orders[orderID], nil
	}
	now := time.Now()
	order := Order{
		ID:            s.nextID,
		OrderNo:       fmt.Sprintf("FREE-%d-%d", gameID, s.nextID),
		UserID:        userID,
		GameID:        gameID,
		AmountCent:    0,
		PayStatus:     "free_no_pay",
		PayChannel:    "none",
		NeedWechatPay: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.nextID++
	s.orders[order.ID] = order
	s.ordersByNo[order.OrderNo] = order.ID
	s.ordersByGame[gameID] = order.ID
	return order, nil
}

func (s *Service) EnsureGuideFeePlaceholderOrder(userID int64) (Order, error) {
	if userID <= 0 {
		return Order{}, ErrInvalidOrder
	}
	if s.repo != nil {
		return s.repo.EnsureGuideFeePlaceholderOrder(context.Background(), userID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	orderNo := fmt.Sprintf("GUIDE-FEE-%d", userID)
	if orderID, ok := s.ordersByNo[orderNo]; ok {
		return s.orders[orderID], nil
	}
	now := time.Now()
	order := Order{
		ID:            s.nextID,
		OrderNo:       orderNo,
		UserID:        userID,
		AmountCent:    0,
		PayStatus:     "guide_fee_placeholder",
		PayChannel:    "none",
		NeedWechatPay: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.nextID++
	s.orders[order.ID] = order
	s.ordersByNo[order.OrderNo] = order.ID
	return order, nil
}

func (s *Service) GetForUser(userID int64, orderID int64) (Order, error) {
	if s.repo != nil {
		order, ok, err := s.repo.FindOrderForUser(context.Background(), userID, orderID)
		if err != nil {
			return Order{}, err
		}
		if !ok {
			return Order{}, ErrOrderNotFound
		}
		return order, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[orderID]
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	if order.UserID != userID {
		return Order{}, ErrForbidden
	}
	return order, nil
}

func (s *Service) PrecreateFreeNoPay(userID int64, gameID int64) (PrecreateResult, error) {
	order, err := s.EnsureFreeNoPayOrder(userID, gameID)
	if err != nil {
		return PrecreateResult{}, err
	}
	return PrecreateResult{
		Order:         order,
		OrderNo:       order.OrderNo,
		PayStatus:     order.PayStatus,
		AmountCent:    order.AmountCent,
		NeedWechatPay: order.NeedWechatPay,
		Mode:          "free_no_pay_placeholder",
	}, nil
}

func (s *Service) PrecreateGuideFeePlaceholder(userID int64) (PrecreateResult, error) {
	order, err := s.EnsureGuideFeePlaceholderOrder(userID)
	if err != nil {
		return PrecreateResult{}, err
	}
	return PrecreateResult{
		Order:         order,
		OrderNo:       order.OrderNo,
		PayStatus:     order.PayStatus,
		AmountCent:    order.AmountCent,
		NeedWechatPay: order.NeedWechatPay,
		Mode:          "guide_fee_placeholder",
	}, nil
}

func (s *Service) RecordCallback(req CallbackRequest) (CallbackResult, error) {
	if req.OrderNo == "" || req.EventID == "" {
		return CallbackResult{}, ErrCallbackInvalid
	}
	digest := digestPayload(req.Payload)
	if s.repo != nil {
		if _, ok, err := s.repo.FindOrderByNo(context.Background(), req.OrderNo); err != nil {
			return CallbackResult{}, err
		} else if !ok {
			return CallbackResult{}, ErrOrderNotFound
		}
		idempotentHit, err := s.repo.SaveCallback(context.Background(), req, digest)
		if err != nil {
			return CallbackResult{}, err
		}
		return CallbackResult{OrderNo: req.OrderNo, EventID: req.EventID, Received: true, Verified: true, IdempotentHit: idempotentHit, Mode: "placeholder"}, nil
	}
	key := req.OrderNo + ":" + req.EventID
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ordersByNo[req.OrderNo]; !ok {
		return CallbackResult{}, ErrOrderNotFound
	}
	if existing, ok := s.callbackEvents[key]; ok {
		return CallbackResult{
			OrderNo:       existing.orderNo,
			EventID:       existing.eventID,
			Received:      true,
			Verified:      true,
			IdempotentHit: true,
			Mode:          "placeholder",
		}, nil
	}
	s.callbackEvents[key] = callbackRecord{eventID: req.EventID, orderNo: req.OrderNo, payloadDigest: digest, createdAt: time.Now()}
	return CallbackResult{OrderNo: req.OrderNo, EventID: req.EventID, Received: true, Verified: true, Mode: "placeholder"}, nil
}

func (s *Service) CreateProfitSharingPlaceholder(req ProfitSharingOrderRequest) (ProfitSharingOrder, error) {
	if req.OutOrderNo == "" || req.OrderNo == "" || req.AmountCent < 0 {
		return ProfitSharingOrder{}, ErrInvalidOrder
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ordersByNo[req.OrderNo]; !ok && s.repo == nil {
		return ProfitSharingOrder{}, ErrOrderNotFound
	}
	if item, ok := s.shareOrders[req.OutOrderNo]; ok {
		return item, nil
	}
	item := ProfitSharingOrder{
		OutOrderNo:    req.OutOrderNo,
		OrderNo:       req.OrderNo,
		AmountCent:    req.AmountCent,
		Status:        "share_placeholder",
		NeedWechatPay: false,
		Placeholder:   true,
		Mode:          "profit_sharing_placeholder",
		CreatedAt:     time.Now(),
	}
	s.shareOrders[item.OutOrderNo] = item
	return item, nil
}

func (s *Service) ProfitSharingPlaceholder(outOrderNo string) (ProfitSharingOrder, error) {
	if outOrderNo == "" {
		return ProfitSharingOrder{}, ErrInvalidOrder
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.shareOrders[outOrderNo]
	if !ok {
		return ProfitSharingOrder{}, ErrOrderNotFound
	}
	item.LastQueriedAt = time.Now()
	s.shareOrders[outOrderNo] = item
	return item, nil
}

func (s *Service) CreateProfitSharingReturnPlaceholder(req ProfitSharingReturnRequest) (ProfitSharingReturn, error) {
	if req.OutReturnNo == "" || req.OutOrderNo == "" {
		return ProfitSharingReturn{}, ErrInvalidOrder
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.shareOrders[req.OutOrderNo]; !ok {
		return ProfitSharingReturn{}, ErrOrderNotFound
	}
	if item, ok := s.shareReturns[req.OutReturnNo]; ok {
		return item, nil
	}
	item := ProfitSharingReturn{
		OutReturnNo:   req.OutReturnNo,
		OutOrderNo:    req.OutOrderNo,
		Status:        "return_placeholder",
		NeedWechatPay: false,
		Placeholder:   true,
		Mode:          "profit_sharing_return_placeholder",
		Reason:        req.Reason,
		CreatedAt:     time.Now(),
	}
	s.shareReturns[item.OutReturnNo] = item
	return item, nil
}

func digestPayload(payload map[string]interface{}) string {
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
