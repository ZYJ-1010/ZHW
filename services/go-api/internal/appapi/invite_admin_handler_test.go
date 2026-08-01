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

func TestInviteQuotaAuditValidatesEntryTypeBeforeChangingStatus(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	ownerToken := loginForTestWithCode(t, mux, "invite-quota-owner")
	ownerID := currentUserIDForTest(t, mux, ownerToken)
	request, err := authService.CreateInviteQuotaRequest(ownerID, 2, "需要邀请体验用户")
	if err != nil {
		t.Fatalf("create quota request: %v", err)
	}

	adminToken := adminLoginForTest(t, mux)
	postJSON(t, mux, "/api/admin/invite-quota-requests/"+strconv.FormatInt(request.ID, 10)+"/audit", adminToken, `{"approve":true,"entryType":"invalid"}`, http.StatusUnprocessableEntity)

	items, err := authService.InviteQuotaRequests(ownerID, "")
	if err != nil {
		t.Fatalf("list quota requests: %v", err)
	}
	if len(items) != 1 || items[0].Status != "pending" {
		payload, _ := json.Marshal(items)
		t.Fatalf("invalid entry type must leave request pending, got %s", payload)
	}
}

func TestInviteQuotaApprovalAtomicallyCreatesAssignedCodes(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	ownerToken := loginForTestWithCode(t, mux, "invite-quota-approved-owner")
	ownerID := currentUserIDForTest(t, mux, ownerToken)
	server.profiles.GrantRole(ownerID, "guide")
	request, err := authService.CreateInviteQuotaRequest(ownerID, 2, "需要邀请体验用户")
	if err != nil {
		t.Fatal(err)
	}
	adminToken := adminLoginForTest(t, mux)
	postJSON(t, mux, "/api/admin/invite-quota-requests/"+strconv.FormatInt(request.ID, 10)+"/audit", adminToken, `{"approve":true,"entryType":"qrcode"}`, http.StatusOK)

	requests, err := authService.InviteQuotaRequests(ownerID, "")
	if err != nil || len(requests) != 1 || requests[0].Status != "approved" {
		t.Fatalf("expected approved request, items=%+v err=%v", requests, err)
	}
	codes, err := authService.AdminInviteCodes(invites.CodeFilter{OwnerID: ownerID, EntryType: invites.EntryTypeQRCode})
	if err != nil || len(codes) != 2 {
		t.Fatalf("expected two assigned codes, items=%+v err=%v", codes, err)
	}
	for _, code := range codes {
		if code.MaxUses != 1 || code.UsedCount != 0 || code.Status != invites.StatusActive {
			t.Fatalf("unexpected generated code: %+v", code)
		}
	}
}

func TestInviteQuotaApprovalRejectsOwnerWhoseRoleIsNotActive(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	ownerToken := loginForTestWithCode(t, mux, "invite-quota-player-owner")
	ownerID := currentUserIDForTest(t, mux, ownerToken)
	request, err := authService.CreateInviteQuotaRequest(ownerID, 1, "无有效身份")
	if err != nil {
		t.Fatal(err)
	}
	adminToken := adminLoginForTest(t, mux)
	postJSON(t, mux, "/api/admin/invite-quota-requests/"+strconv.FormatInt(request.ID, 10)+"/audit", adminToken, `{"approve":true,"entryType":"link"}`, http.StatusUnprocessableEntity)

	requests, err := authService.InviteQuotaRequests(ownerID, "")
	if err != nil || len(requests) != 1 || requests[0].Status != "pending" {
		t.Fatalf("failed approval must leave request pending, items=%+v err=%v", requests, err)
	}
}
