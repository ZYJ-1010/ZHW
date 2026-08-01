package appapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/orders"
	"zhw-mini/services/go-api/internal/users"
)

type failingFreeOrderRepository struct{}

func (failingFreeOrderRepository) EnsureFreeNoPayOrder(context.Context, int64, int64) (orders.Order, error) {
	return orders.Order{}, errors.New("order repository unavailable")
}

func (failingFreeOrderRepository) EnsureGuideFeePlaceholderOrder(context.Context, int64) (orders.Order, error) {
	return orders.Order{}, errors.New("order repository unavailable")
}

func (failingFreeOrderRepository) FindOrderForUser(context.Context, int64, int64) (orders.Order, bool, error) {
	return orders.Order{}, false, errors.New("order repository unavailable")
}

func (failingFreeOrderRepository) FindOrderByNo(context.Context, string) (orders.Order, bool, error) {
	return orders.Order{}, false, errors.New("order repository unavailable")
}

func (failingFreeOrderRepository) SaveCallback(context.Context, orders.CallbackRequest, string) (bool, error) {
	return false, errors.New("order repository unavailable")
}

func TestCreatedGameIsNotReportedFailedWhenOrderRecheckIsUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.orders = orders.NewServiceWithRepository(failingFreeOrderRepository{})
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "game-order-recheck")
	completeIdentityForTest(t, mux, token)
	token = issueFormalTokenForTest(t, mux, token)
	body := postJSON(t, mux, "/api/app/games", token, `{"title":"订单复核降级","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var response struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.ID <= 0 {
		t.Fatalf("created game response is missing its id: %s", string(body))
	}
	if _, err := server.games.Get(response.Data.ID); err != nil {
		t.Fatalf("the committed game must remain available after order recheck degradation: %v", err)
	}
}
