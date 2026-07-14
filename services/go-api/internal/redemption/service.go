package redemption

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/points"
)

var (
	ErrInvalidOrder       = errors.New("invalid redemption order")
	ErrItemNotFound       = errors.New("redemption item not found")
	ErrItemInactive       = errors.New("redemption item inactive")
	ErrInsufficientStock  = errors.New("insufficient redemption stock")
	ErrInsufficientPoints = errors.New("insufficient redemption points")
	ErrOrderNotFound      = errors.New("redemption order not found")
	ErrInvalidStatus      = errors.New("invalid redemption status")
)

type PointsLedger interface {
	Deduct(userID int64, value int, bizType string, bizID int64, reason string) (points.Account, points.Log, error)
	Grant(userID int64, value int, bizType string, bizID int64, reason string) (points.Account, points.Log, error)
}

type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	PointsCost  int       `json:"pointsCost"`
	Stock       int       `json:"stock"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Order struct {
	ID            int64     `json:"id"`
	OrderNo       string    `json:"orderNo"`
	UserID        int64     `json:"userId"`
	ItemID        int64     `json:"itemId"`
	ItemName      string    `json:"itemName"`
	ItemImageURL  string    `json:"imageUrl,omitempty"`
	PointsCost    int       `json:"pointsCost"`
	Status        string    `json:"status"`
	ReviewAdminID int64     `json:"reviewAdminId,omitempty"`
	ReviewReason  string    `json:"reviewReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	PointsCost  int    `json:"pointsCost"`
	Stock       int    `json:"stock"`
}

type CreateOrderRequest struct {
	ItemID int64 `json:"itemId"`
}

type UpdateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	PointsCost  int    `json:"pointsCost"`
	Stock       *int   `json:"stock"`
	Status      string `json:"status"`
}

type ReviewOrderRequest struct {
	Approve bool   `json:"approve"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type Repository interface {
	CreateItem(ctx context.Context, item Item) (Item, error)
	ListItems(ctx context.Context, includeInactive bool) ([]Item, error)
	FindItem(ctx context.Context, itemID int64) (Item, bool, error)
	UpdateItem(ctx context.Context, item Item) (Item, error)
	ReserveItemStock(ctx context.Context, itemID int64) (Item, error)
	RestoreItemStock(ctx context.Context, itemID int64) error
	NextOrderID(ctx context.Context) (int64, error)
	CreateOrder(ctx context.Context, order Order) (Order, error)
	ListOrdersByUser(ctx context.Context, userID int64) ([]Order, error)
	ListOrders(ctx context.Context) ([]Order, error)
	FindOrder(ctx context.Context, orderID int64) (Order, bool, error)
	UpdateOrder(ctx context.Context, order Order) (Order, error)
}

type Service struct {
	mu          sync.Mutex
	nextItemID  int64
	nextOrderID int64
	items       map[int64]Item
	orders      map[int64]Order
	points      PointsLedger
	repo        Repository
}

func NewService(points PointsLedger) *Service {
	return NewServiceWithRepository(points, nil)
}

func NewServiceWithRepository(points PointsLedger, repo Repository) *Service {
	return &Service{
		nextItemID:  1,
		nextOrderID: 1,
		items:       make(map[int64]Item),
		orders:      make(map[int64]Order),
		points:      points,
		repo:        repo,
	}
}

func (s *Service) CreateItem(req CreateItemRequest) (Item, error) {
	if req.Name == "" || req.PointsCost <= 0 || req.Stock < 0 {
		return Item{}, ErrInvalidOrder
	}
	if s.repo != nil {
		now := time.Now()
		return s.repo.CreateItem(context.Background(), Item{
			Name:        req.Name,
			Description: req.Description,
			ImageURL:    req.ImageURL,
			PointsCost:  req.PointsCost,
			Stock:       req.Stock,
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	item := Item{
		ID:          s.nextItemID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		PointsCost:  req.PointsCost,
		Stock:       req.Stock,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.nextItemID++
	s.items[item.ID] = item
	return item, nil
}

func (s *Service) Items() []Item {
	if s.repo != nil {
		if items, err := s.repo.ListItems(context.Background(), false); err == nil {
			return items
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Item, 0)
	for _, item := range s.items {
		if item.Status == "active" {
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) AdminItems() []Item {
	if s.repo != nil {
		if items, err := s.repo.ListItems(context.Background(), true); err == nil {
			return items
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

func (s *Service) UpdateItem(itemID int64, req UpdateItemRequest) (Item, error) {
	if s.repo != nil {
		item, ok, err := s.repo.FindItem(context.Background(), itemID)
		if err != nil {
			return Item{}, err
		}
		if !ok {
			return Item{}, ErrItemNotFound
		}
		updated, err := applyItemUpdate(item, req)
		if err != nil {
			return Item{}, err
		}
		return s.repo.UpdateItem(context.Background(), updated)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[itemID]
	if !ok {
		return Item{}, ErrItemNotFound
	}
	item, err := applyItemUpdate(item, req)
	if err != nil {
		return Item{}, err
	}
	s.items[item.ID] = item
	return item, nil
}

func (s *Service) CreateOrder(userID int64, req CreateOrderRequest) (Order, error) {
	if req.ItemID <= 0 {
		return Order{}, ErrInvalidOrder
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.repo != nil {
		item, ok, err := s.repo.FindItem(context.Background(), req.ItemID)
		if err != nil {
			return Order{}, err
		}
		if !ok {
			return Order{}, ErrItemNotFound
		}
		if item.Status != "active" {
			return Order{}, ErrItemInactive
		}
		if item.Stock <= 0 {
			return Order{}, ErrInsufficientStock
		}
		item, err = s.repo.ReserveItemStock(context.Background(), req.ItemID)
		if err != nil {
			if errors.Is(err, ErrInsufficientStock) {
				return Order{}, ErrInsufficientStock
			}
			return Order{}, err
		}
		orderID, err := s.repo.NextOrderID(context.Background())
		if err != nil {
			_ = s.repo.RestoreItemStock(context.Background(), item.ID)
			return Order{}, err
		}
		order := Order{
			ID:           orderID,
			OrderNo:      fmt.Sprintf("RDM-%d", orderID),
			UserID:       userID,
			ItemID:       item.ID,
			ItemName:     item.Name,
			ItemImageURL: item.ImageURL,
			PointsCost:   item.PointsCost,
			Status:       "pending",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if _, _, err := s.points.Deduct(userID, order.PointsCost, "redemption_order", order.ID, "兑换商品"); err != nil {
			_ = s.repo.RestoreItemStock(context.Background(), item.ID)
			if errors.Is(err, points.ErrInsufficientPoints) {
				return Order{}, ErrInsufficientPoints
			}
			return Order{}, err
		}
		order, err = s.repo.CreateOrder(context.Background(), order)
		if err != nil {
			_, _, _ = s.points.Grant(userID, item.PointsCost, "redemption_refund", order.ID, "兑换建单失败退回积分")
			_ = s.repo.RestoreItemStock(context.Background(), item.ID)
			return Order{}, err
		}
		return order, nil
	}
	item, ok := s.items[req.ItemID]
	if !ok {
		return Order{}, ErrItemNotFound
	}
	if item.Status != "active" {
		return Order{}, ErrItemInactive
	}
	if item.Stock <= 0 {
		return Order{}, ErrInsufficientStock
	}
	orderID := s.nextOrderID
	now := time.Now()
	order := Order{
		ID:           orderID,
		OrderNo:      fmt.Sprintf("RDM-%d", orderID),
		UserID:       userID,
		ItemID:       item.ID,
		ItemName:     item.Name,
		ItemImageURL: item.ImageURL,
		PointsCost:   item.PointsCost,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.nextOrderID++
	item.Stock--
	item.UpdatedAt = now
	s.items[item.ID] = item

	if _, _, err := s.points.Deduct(userID, order.PointsCost, "redemption_order", order.ID, "兑换商品"); err != nil {
		item := s.items[order.ItemID]
		item.Stock++
		item.UpdatedAt = time.Now()
		s.items[item.ID] = item
		if errors.Is(err, points.ErrInsufficientPoints) {
			return Order{}, ErrInsufficientPoints
		}
		return Order{}, err
	}

	s.orders[order.ID] = order
	return order, nil
}

func (s *Service) OrdersForUser(userID int64) []Order {
	if s.repo != nil {
		if items, err := s.repo.ListOrdersByUser(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Order, 0)
	for _, order := range s.orders {
		if order.UserID == userID {
			result = append(result, order)
		}
	}
	return result
}

func (s *Service) OrderForUser(userID int64, orderID int64) (Order, error) {
	if s.repo != nil {
		order, ok, err := s.repo.FindOrder(context.Background(), orderID)
		if err != nil {
			return Order{}, err
		}
		if !ok || order.UserID != userID {
			return Order{}, ErrOrderNotFound
		}
		return order, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	order, ok := s.orders[orderID]
	if !ok || order.UserID != userID {
		return Order{}, ErrOrderNotFound
	}
	return order, nil
}

func (s *Service) AdminOrders() []Order {
	if s.repo != nil {
		if items, err := s.repo.ListOrders(context.Background()); err == nil {
			return items
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Order, 0, len(s.orders))
	for _, order := range s.orders {
		result = append(result, order)
	}
	return result
}

func (s *Service) CancelOrder(userID int64, orderID int64, reason string) (Order, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "用户主动取消"
	}
	if s.repo != nil {
		order, ok, err := s.repo.FindOrder(context.Background(), orderID)
		if err != nil {
			return Order{}, err
		}
		if !ok || order.UserID != userID {
			return Order{}, ErrOrderNotFound
		}
		if order.Status != "pending" {
			return Order{}, ErrInvalidStatus
		}
		order.Status = "canceled"
		order.ReviewReason = reason
		order.UpdatedAt = time.Now()
		order, err = s.repo.UpdateOrder(context.Background(), order)
		if err != nil {
			return Order{}, err
		}
		if err := s.repo.RestoreItemStock(context.Background(), order.ItemID); err != nil {
			return Order{}, err
		}
		if _, _, err := s.points.Grant(order.UserID, order.PointsCost, "redemption_cancel_refund", order.ID, "兑换取消退回积分"); err != nil {
			return Order{}, err
		}
		return order, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	order, ok := s.orders[orderID]
	if !ok || order.UserID != userID {
		return Order{}, ErrOrderNotFound
	}
	if order.Status != "pending" {
		return Order{}, ErrInvalidStatus
	}
	now := time.Now()
	order.Status = "canceled"
	order.ReviewReason = reason
	order.UpdatedAt = now
	s.orders[order.ID] = order
	if item, ok := s.items[order.ItemID]; ok {
		item.Stock++
		item.UpdatedAt = now
		s.items[item.ID] = item
	}
	if _, _, err := s.points.Grant(order.UserID, order.PointsCost, "redemption_cancel_refund", order.ID, "兑换取消退回积分"); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s *Service) ReviewOrder(orderID int64, adminID int64, req ReviewOrderRequest) (Order, error) {
	if s.repo != nil {
		order, ok, err := s.repo.FindOrder(context.Background(), orderID)
		if err != nil {
			return Order{}, err
		}
		if !ok {
			return Order{}, ErrOrderNotFound
		}
		order, targetStatus, err := applyOrderReview(order, adminID, req)
		if err != nil {
			return Order{}, err
		}
		order, err = s.repo.UpdateOrder(context.Background(), order)
		if err != nil {
			return Order{}, err
		}
		if targetStatus == "rejected" {
			if err := s.repo.RestoreItemStock(context.Background(), order.ItemID); err != nil {
				return Order{}, err
			}
			if _, _, err := s.points.Grant(order.UserID, order.PointsCost, "redemption_refund", order.ID, "兑换驳回退回积分"); err != nil {
				return Order{}, err
			}
		}
		return order, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	order, ok := s.orders[orderID]
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	order, targetStatus, err := applyOrderReview(order, adminID, req)
	if err != nil {
		return Order{}, err
	}
	s.orders[order.ID] = order
	if targetStatus == "rejected" {
		if item, ok := s.items[order.ItemID]; ok {
			item.Stock++
			item.UpdatedAt = time.Now()
			s.items[item.ID] = item
		}
		if _, _, err := s.points.Grant(order.UserID, order.PointsCost, "redemption_refund", order.ID, "兑换驳回退回积分"); err != nil {
			return Order{}, err
		}
	}
	return order, nil
}

func applyItemUpdate(item Item, req UpdateItemRequest) (Item, error) {
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.ImageURL != "" {
		item.ImageURL = strings.TrimSpace(req.ImageURL)
	}
	if req.PointsCost > 0 {
		item.PointsCost = req.PointsCost
	}
	if req.Stock != nil {
		if *req.Stock < 0 {
			return Item{}, ErrInvalidOrder
		}
		item.Stock = *req.Stock
	}
	if req.Status != "" {
		if req.Status != "active" && req.Status != "inactive" {
			return Item{}, ErrInvalidStatus
		}
		item.Status = req.Status
	}
	item.UpdatedAt = time.Now()
	return item, nil
}

func applyOrderReview(order Order, adminID int64, req ReviewOrderRequest) (Order, string, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	if order.Status != "pending" && order.Status != "approved" {
		return Order{}, "", ErrInvalidStatus
	}
	targetStatus := req.Status
	if targetStatus == "" {
		if req.Approve {
			targetStatus = "approved"
		} else {
			targetStatus = "rejected"
		}
	}
	if targetStatus != "approved" && targetStatus != "rejected" && targetStatus != "fulfilled" {
		return Order{}, "", ErrInvalidStatus
	}
	if req.Reason == "" {
		return Order{}, "", ErrInvalidOrder
	}
	if targetStatus == "fulfilled" && order.Status != "approved" {
		return Order{}, "", ErrInvalidStatus
	}
	order.Status = targetStatus
	order.ReviewAdminID = adminID
	order.ReviewReason = req.Reason
	order.UpdatedAt = time.Now()
	return order, targetStatus, nil
}
