package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestIssueTokenAfterManualReviewForPhaseOne(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	preAuthToken := loginForTestWithCode(t, mux, "sms-phase-one")
	postJSON(t, mux, "/api/app/identity/phone/bind", preAuthToken, `{"phone":"13800138000"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/send-code", preAuthToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/verify-code", preAuthToken, `{"code":"000000"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/auth/issue-token-after-identity", preAuthToken, `{}`, http.StatusForbidden)

	submitBody := postJSON(t, mux, "/api/app/identity/phone/verify", preAuthToken, `{"realName":"Test User","idCard":"110101199001011234"}`, http.StatusOK)
	var submitResp struct {
		Data struct {
			UserID int64  `json:"userId"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(submitBody, &submitResp); err != nil {
		t.Fatal(err)
	}
	if submitResp.Data.UserID == 0 || submitResp.Data.Status != "pending" {
		t.Fatalf("expected pending manual review submission: %s", string(submitBody))
	}

	adminToken := adminLoginForTest(t, mux)
	postAdminJSON(t, mux, "/api/admin/identity-verifications/"+strconv.FormatInt(submitResp.Data.UserID, 10)+"/review", adminToken, `{"approve":true,"reason":"ok"}`, http.StatusOK)

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
	if resp.Data.Token == "" || resp.Data.IdentityBindStatus != "verified" {
		t.Fatalf("expected manually reviewed user to receive app token: %s", string(body))
	}
}

func TestManualRealnameSubmissionAndAdminReviewForPhaseOne(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	token := loginForTestWithCode(t, mux, "manual-realname-phase-one")
	body := postJSON(t, mux, "/api/app/identity/phone/verify", token, `{"realName":"Test User","idCard":"110101199001011234"}`, http.StatusOK)
	var submitResp struct {
		Data struct {
			UserID        int64  `json:"userId"`
			Status        string `json:"status"`
			PhoneVerified bool   `json:"phoneVerified"`
			IDCardMasked  string `json:"idCardMasked"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &submitResp); err != nil {
		t.Fatal(err)
	}
	if submitResp.Data.UserID == 0 || submitResp.Data.Status != "pending" || submitResp.Data.PhoneVerified || submitResp.Data.IDCardMasked == "" {
		t.Fatalf("expected pending manual realname submission: %s", string(body))
	}

	postJSON(t, mux, "/api/app/auth/issue-token-after-identity", token, `{}`, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	reviewBody := postAdminJSON(t, mux, "/api/admin/identity-verifications/"+strconv.FormatInt(submitResp.Data.UserID, 10)+"/review", adminToken, `{"approve":true,"reason":"ok"}`, http.StatusOK)
	var reviewResp struct {
		Data struct {
			Status        string `json:"status"`
			PhoneVerified bool   `json:"phoneVerified"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reviewBody, &reviewResp); err != nil {
		t.Fatal(err)
	}
	if reviewResp.Data.Status != "verified" || !reviewResp.Data.PhoneVerified {
		t.Fatalf("expected admin approval to verify identity: %s", string(reviewBody))
	}
	user, ok := authService.UserByID(submitResp.Data.UserID)
	if !ok || user.RealnameStatus != "verified" {
		t.Fatalf("expected user realname status to sync after approval, got ok=%v user=%+v", ok, user)
	}

	issuedBody := postJSON(t, mux, "/api/app/auth/issue-token-after-identity", token, `{}`, http.StatusOK)
	var issuedResp struct {
		Data struct {
			Token              string `json:"token"`
			IdentityBindStatus string `json:"identityBindStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(issuedBody, &issuedResp); err != nil {
		t.Fatal(err)
	}
	if issuedResp.Data.Token == "" || issuedResp.Data.IdentityBindStatus != "verified" {
		t.Fatalf("expected approved manual realname user to receive app token: %s", string(issuedBody))
	}
}
