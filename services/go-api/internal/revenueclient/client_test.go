package revenueclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCallsFundsServicePlaceholders(t *testing.T) {
	seen := make(map[string]bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.Method+" "+r.URL.Path] = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data": map[string]interface{}{
				"path": r.URL.Path,
			},
		})
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()
	calls := []struct {
		name string
		fn   func() (Response, error)
	}{
		{"health", func() (Response, error) { return client.Health(ctx) }},
		{"payment", func() (Response, error) {
			return client.PaymentPrecreatePlaceholder(ctx, map[string]interface{}{"gameId": 1})
		}},
		{"callback", func() (Response, error) {
			return client.PaymentCallbackPlaceholder(ctx, map[string]interface{}{"orderNo": "o1"})
		}},
		{"profit sharing", func() (Response, error) {
			return client.ProfitSharingOrderPlaceholder(ctx, map[string]interface{}{"outOrderNo": "PS-1"})
		}},
		{"profit sharing return", func() (Response, error) {
			return client.ProfitSharingReturnPlaceholder(ctx, map[string]interface{}{"outReturnNo": "PR-1"})
		}},
		{"template", func() (Response, error) {
			return client.CreateTemplate(ctx, TemplateRequest{Name: "default", GameType: "free", PlatformBps: 1000, CreatorBps: 3000, MemberBps: 6000})
		}},
		{"simulate", func() (Response, error) {
			return client.Simulate(ctx, CalculateRequest{GameID: 1, AmountCent: 10000, TemplateID: 1, MemberIDs: []int64{1, 2}, CreatorID: 1})
		}},
		{"generate", func() (Response, error) {
			return client.Generate(ctx, CalculateRequest{GameID: 1, AmountCent: 10000, TemplateID: 1})
		}},
		{"settle", func() (Response, error) {
			return client.SettleOffline(ctx, SettlementRequest{RecordID: 1, Method: "offline", ProofNo: "P001"})
		}},
	}
	for _, call := range calls {
		if _, err := call.fn(); err != nil {
			t.Fatalf("%s failed: %v", call.name, err)
		}
	}
	expected := []string{
		"GET /health",
		"POST /api/funds/payment-precreate-placeholder",
		"POST /api/funds/payment-callback-placeholder",
		"POST /api/funds/profit-sharing/orders",
		"POST /api/funds/profit-sharing/return-orders",
		"POST /api/funds/revenue/templates",
		"POST /api/funds/revenue/simulate",
		"POST /api/funds/revenue/generate",
		"POST /api/funds/settlements/offline",
	}
	for _, key := range expected {
		if !seen[key] {
			t.Fatalf("expected request %s, seen=%v", key, seen)
		}
	}
}

func TestClientReturnsFundsServiceErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    40941,
			"message": "review incomplete",
			"data":    nil,
		})
	}))
	defer server.Close()

	_, err := New(server.URL).Generate(context.Background(), CalculateRequest{GameID: 1})
	if err == nil {
		t.Fatal("expected error")
	}
}
