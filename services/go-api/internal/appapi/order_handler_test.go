package appapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestFreeGameCreatesPaymentOrderAndPrecreateDoesNotTriggerWechatPay(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "order-creator")
	completeIdentityForTest(t, mux, token)
	token = issueFormalTokenForTest(t, mux, token)

	postJSON(t, mux, "/api/app/games", token, `{"title":"free order game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	orderBody := getJSON(t, mux, "/api/app/orders/1", token, http.StatusOK)
	var createdOrderResp struct {
		Data struct {
			ID            int64  `json:"id"`
			PayStatus     string `json:"payStatus"`
			NeedWechatPay bool   `json:"needWechatPay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(orderBody, &createdOrderResp); err != nil {
		t.Fatal(err)
	}
	if createdOrderResp.Data.ID != 1 || createdOrderResp.Data.PayStatus != "free_no_pay" || createdOrderResp.Data.NeedWechatPay {
		t.Fatalf("expected create game to auto create free_no_pay order: %s", string(orderBody))
	}

	precreateBody := postJSON(t, mux, "/api/app/payment/precreate-placeholder", token, `{"gameId":1}`, http.StatusOK)
	var precreateResp struct {
		Data struct {
			Order struct {
				ID            int64  `json:"id"`
				OrderNo       string `json:"orderNo"`
				PayStatus     string `json:"payStatus"`
				NeedWechatPay bool   `json:"needWechatPay"`
			} `json:"order"`
			OrderNo       string `json:"orderNo"`
			PayStatus     string `json:"payStatus"`
			NeedWechatPay bool   `json:"needWechatPay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(precreateBody, &precreateResp); err != nil {
		t.Fatal(err)
	}
	if precreateResp.Data.Order.ID != 1 || precreateResp.Data.OrderNo == "" || precreateResp.Data.PayStatus != "free_no_pay" || precreateResp.Data.NeedWechatPay {
		t.Fatalf("unexpected precreate placeholder: %s", string(precreateBody))
	}

	var orderResp struct {
		Data struct {
			ID            int64  `json:"id"`
			PayStatus     string `json:"payStatus"`
			NeedWechatPay bool   `json:"needWechatPay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(orderBody, &orderResp); err != nil {
		t.Fatal(err)
	}
	if orderResp.Data.ID != 1 || orderResp.Data.PayStatus != "free_no_pay" || orderResp.Data.NeedWechatPay {
		t.Fatalf("unexpected app order detail: %s", string(orderBody))
	}
}

func TestPaymentCallbackPlaceholderIsIdempotentHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "order-callback")
	completeIdentityForTest(t, mux, token)
	token = issueFormalTokenForTest(t, mux, token)

	postJSON(t, mux, "/api/internal/pay/callback-placeholder", "", `{"eventId":"evt-1"}`, http.StatusNotFound)
}

func TestGuidePaymentPlaceholderUpdatesQualificationHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "guide-payment")
	completeIdentityForTest(t, mux, token)
	token = issueFormalTokenForTest(t, mux, token)

	postJSON(t, mux, "/api/app/guides/payment/precreate-placeholder", token, `{}`, http.StatusNotFound)
}

func TestProfitSharingPlaceholdersDoNotTriggerWechatPayHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	postJSON(t, mux, "/api/funds/profit-sharing/orders", "", `{"outOrderNo":"PS-1","orderNo":"ORD-1","amountCent":1000}`, http.StatusNotFound)
	queryReq, _ := http.NewRequest(http.MethodGet, "/api/funds/profit-sharing/orders/PS-1", nil)
	queryRec := httptest.NewRecorder()
	mux.ServeHTTP(queryRec, queryReq)
	if queryRec.Code != http.StatusNotFound {
		t.Fatalf("expected phase-one profit sharing query disabled, got %d: %s", queryRec.Code, queryRec.Body.String())
	}
	postJSON(t, mux, "/api/funds/profit-sharing/return-orders", "", `{"outReturnNo":"PR-1","outOrderNo":"PS-1","reason":"refund"}`, http.StatusNotFound)
}

func issueFormalTokenForTest(t *testing.T, mux *http.ServeMux, preAuthToken string) string {
	t.Helper()
	body := postJSON(t, mux, "/api/app/auth/issue-token-after-identity", preAuthToken, `{}`, http.StatusOK)
	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Token == "" {
		t.Fatalf("expected formal token: %s", string(body))
	}
	return resp.Data.Token
}
