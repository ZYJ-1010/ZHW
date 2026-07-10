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

func TestIssueTokenAfterSMSVerificationForPhaseOne(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	preAuthToken := loginForTestWithCode(t, mux, "sms-phase-one")
	postJSON(t, mux, "/api/app/identity/phone/bind", preAuthToken, `{"phone":"13800138000"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/send-code", preAuthToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/verify-code", preAuthToken, `{"code":"000000"}`, http.StatusOK)

	body := postJSON(t, mux, "/api/app/auth/issue-token-after-identity", preAuthToken, `{}`, http.StatusOK)
	var resp struct {
		Data struct {
			Token              string `json:"token"`
			IdentityBindStatus string `json:"identityBindStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Token == "" || resp.Data.IdentityBindStatus != "sms_verified" {
		t.Fatalf("expected sms verified user to receive app token: %s", string(body))
	}
}
