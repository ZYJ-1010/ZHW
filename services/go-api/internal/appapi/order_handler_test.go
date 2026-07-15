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

	postJSON(t, mux, "/api/app/games", token, `{"title":"callback game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	precreateBody := postJSON(t, mux, "/api/app/payment/precreate-placeholder", token, `{"gameId":1}`, http.StatusOK)
	var precreateResp struct {
		Data struct {
			OrderNo string `json:"orderNo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(precreateBody, &precreateResp); err != nil {
		t.Fatal(err)
	}

	body := `{"orderNo":"` + precreateResp.Data.OrderNo + `","eventId":"evt-1","payload":{"trade_state":"SUCCESS"}}`
	firstBody := postJSON(t, mux, "/api/internal/pay/callback-placeholder", "", body, http.StatusOK)
	secondBody := postJSON(t, mux, "/api/internal/pay/callback-placeholder", "", body, http.StatusOK)
	var firstResp, secondResp struct {
		Data struct {
			Received      bool `json:"received"`
			Verified      bool `json:"verified"`
			IdempotentHit bool `json:"idempotentHit"`
		} `json:"data"`
	}
	if err := json.Unmarshal(firstBody, &firstResp); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(secondBody, &secondResp); err != nil {
		t.Fatal(err)
	}
	if !firstResp.Data.Received || !firstResp.Data.Verified || firstResp.Data.IdempotentHit || !secondResp.Data.IdempotentHit {
		t.Fatalf("unexpected callback results: first=%s second=%s", string(firstBody), string(secondBody))
	}
}

func TestGuidePaymentPlaceholderUpdatesQualificationHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "guide-payment")
	completeIdentityForTest(t, mux, token)
	token = issueFormalTokenForTest(t, mux, token)

	body := postJSON(t, mux, "/api/app/guides/payment/precreate-placeholder", token, `{}`, http.StatusOK)
	var resp struct {
		Data struct {
			Payment struct {
				OrderNo       string `json:"orderNo"`
				PayStatus     string `json:"payStatus"`
				NeedWechatPay bool   `json:"needWechatPay"`
				Mode          string `json:"mode"`
			} `json:"payment"`
			Qualification struct {
				UserID          int64  `json:"userId"`
				PaymentMet      bool   `json:"paymentMet"`
				GuideOpenStatus string `json:"guideOpenStatus"`
			} `json:"qualification"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Payment.OrderNo == "" || resp.Data.Payment.PayStatus != "guide_fee_placeholder" || resp.Data.Payment.NeedWechatPay || resp.Data.Payment.Mode != "guide_fee_placeholder" {
		t.Fatalf("unexpected guide payment placeholder: %s", string(body))
	}
	if resp.Data.Qualification.UserID != 1 || !resp.Data.Qualification.PaymentMet || resp.Data.Qualification.GuideOpenStatus != "waiting_condition" {
		t.Fatalf("expected guide payment qualification update: %s", string(body))
	}

	againBody := postJSON(t, mux, "/api/app/guides/payment/precreate-placeholder", token, `{}`, http.StatusOK)
	var againResp struct {
		Data struct {
			Payment struct {
				OrderNo string `json:"orderNo"`
			} `json:"payment"`
		} `json:"data"`
	}
	if err := json.Unmarshal(againBody, &againResp); err != nil {
		t.Fatal(err)
	}
	if againResp.Data.Payment.OrderNo != resp.Data.Payment.OrderNo {
		t.Fatalf("expected idempotent guide payment placeholder: first=%s second=%s", string(body), string(againBody))
	}
}

func TestProfitSharingPlaceholdersDoNotTriggerWechatPayHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "profit-sharing")
	completeIdentityForTest(t, mux, token)
	token = issueFormalTokenForTest(t, mux, token)

	postJSON(t, mux, "/api/app/games", token, `{"title":"profit sharing game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	precreateBody := postJSON(t, mux, "/api/app/payment/precreate-placeholder", token, `{"gameId":1}`, http.StatusOK)
	var precreateResp struct {
		Data struct {
			OrderNo string `json:"orderNo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(precreateBody, &precreateResp); err != nil {
		t.Fatal(err)
	}

	shareBody := postJSON(t, mux, "/api/funds/profit-sharing/orders", "", `{"outOrderNo":"PS-1","orderNo":"`+precreateResp.Data.OrderNo+`","amountCent":1000}`, http.StatusOK)
	var shareResp struct {
		Data struct {
			Status        string `json:"status"`
			NeedWechatPay bool   `json:"needWechatPay"`
			Placeholder   bool   `json:"placeholder"`
			Mode          string `json:"mode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(shareBody, &shareResp); err != nil {
		t.Fatal(err)
	}
	if shareResp.Data.Status != "share_placeholder" || shareResp.Data.NeedWechatPay || !shareResp.Data.Placeholder || shareResp.Data.Mode != "profit_sharing_placeholder" {
		t.Fatalf("unexpected profit sharing placeholder: %s", string(shareBody))
	}

	queryReq, _ := http.NewRequest(http.MethodGet, "/api/funds/profit-sharing/orders/PS-1", nil)
	queryRec := httptest.NewRecorder()
	mux.ServeHTTP(queryRec, queryReq)
	if queryRec.Code != http.StatusOK {
		t.Fatalf("expected profit sharing query 200, got %d: %s", queryRec.Code, queryRec.Body.String())
	}
	returnBody := postJSON(t, mux, "/api/funds/profit-sharing/return-orders", "", `{"outReturnNo":"PR-1","outOrderNo":"PS-1","reason":"refund"}`, http.StatusOK)
	var returnResp struct {
		Data struct {
			Status        string `json:"status"`
			NeedWechatPay bool   `json:"needWechatPay"`
			Placeholder   bool   `json:"placeholder"`
			Mode          string `json:"mode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(returnBody, &returnResp); err != nil {
		t.Fatal(err)
	}
	if returnResp.Data.Status != "return_placeholder" || returnResp.Data.NeedWechatPay || !returnResp.Data.Placeholder || returnResp.Data.Mode != "profit_sharing_return_placeholder" {
		t.Fatalf("unexpected profit sharing return placeholder: %s", string(returnBody))
	}
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
