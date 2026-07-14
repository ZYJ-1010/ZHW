package appapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestProductionPaymentPreviewComesFromBackend(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)
	token := loginForTestWithCode(t, mux, "payment-preview-owner")
	completeIdentityForTest(t, mux, token)

	created := postJSON(t, mux, "/api/app/games", token, `{"title":"payment preview","gameType":"free","minPlayers":5,"maxPlayers":8}`, http.StatusOK)
	var createdResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created, &createdResp); err != nil || createdResp.Data.ID <= 0 {
		t.Fatalf("invalid create response: %s", created)
	}
	body := getJSON(t, mux, "/api/app/games/1/payment-preview", token, http.StatusOK)
	var response struct {
		Data struct {
			Payment struct {
				GameID        int64 `json:"gameId"`
				AmountCent    int64 `json:"amountCent"`
				NeedWechatPay bool  `json:"needWechatPay"`
			} `json:"payment"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Payment.GameID != createdResp.Data.ID || response.Data.Payment.AmountCent != 0 || response.Data.Payment.NeedWechatPay {
		t.Fatalf("unexpected payment preview: %s", body)
	}
}

func TestProductionReportRejectsSelfAndOutsider(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)
	ownerToken := loginForTestWithCode(t, mux, "report-owner")
	completeIdentityForTest(t, mux, ownerToken)
	outsiderToken := loginForTestWithCode(t, mux, "report-outsider")
	completeIdentityForTest(t, mux, outsiderToken)
	postJSON(t, mux, "/api/app/games", ownerToken, `{"title":"report boundary","gameType":"free","minPlayers":5,"maxPlayers":8}`, http.StatusOK)

	postJSON(t, mux, "/api/app/reports", ownerToken, `{"gameId":1,"targetUserId":1,"reportType":"other","content":"self"}`, http.StatusUnprocessableEntity)
	postJSON(t, mux, "/api/app/reports", outsiderToken, `{"gameId":1,"targetUserId":1,"reportType":"other","content":"outsider"}`, http.StatusForbidden)
}

func TestProductionCancelPagesReadBackendDetail(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)
	ownerToken := loginForTestWithCode(t, mux, "cancel-owner")
	completeIdentityForTest(t, mux, ownerToken)
	playerToken := loginForTestWithCode(t, mux, "cancel-player")
	completeIdentityForTest(t, mux, playerToken)
	postJSON(t, mux, "/api/app/games", ownerToken, `{"title":"cancel detail","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", ownerToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", ownerToken, `{"approve":true}`, http.StatusOK)
	approveExtraMembersForHTTP(t, mux, ownerToken, 1, "cancel-detail", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", ownerToken, `{}`, http.StatusOK)

	for _, item := range []struct {
		path  string
		token string
		role  string
	}{
		{path: "/api/app/games/1/player-cancel-detail", token: ownerToken, role: "player"},
		{path: "/api/app/games/1/player-cancel-detail", token: playerToken, role: "player"},
		{path: "/api/app/games/1/expert-cancel-detail", token: ownerToken, role: "expert"},
	} {
		body := getJSON(t, mux, item.path, item.token, http.StatusOK)
		var response struct {
			Data struct {
				Role           string `json:"role"`
				GameID         int64  `json:"gameId"`
				ServiceOrderID string `json:"serviceOrderId"`
				ServiceTitle   string `json:"serviceTitle"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		if response.Data.Role != item.role || response.Data.GameID != 1 || response.Data.ServiceOrderID == "" || response.Data.ServiceTitle == "" {
			t.Fatalf("incomplete cancel detail: %s", body)
		}
	}
}
