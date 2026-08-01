package appapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestRoleHomeRequiresActiveBackendRole(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "role-home-access-user")
	completeIdentityForTest(t, mux, token)
	userID := currentUserIDForTest(t, mux, token)

	getJSON(t, mux, "/api/app/home?roleType=expert", token, http.StatusForbidden)
	getJSON(t, mux, "/api/app/home?roleType=guide", token, http.StatusForbidden)

	for _, roleType := range []string{"expert", "guide"} {
		server.profiles.GrantRole(userID, roleType)
		body := getJSON(t, mux, "/api/app/home?roleType="+roleType, token, http.StatusOK)
		var response struct {
			Data struct {
				Hero struct {
					RoleType string `json:"roleType"`
				} `json:"hero"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		if response.Data.Hero.RoleType != roleType {
			t.Fatalf("expected %s home, got %s: %s", roleType, response.Data.Hero.RoleType, string(body))
		}
	}
}

func TestAdminGrantRoleRequiresRealnameAndReconcilesGuide(t *testing.T) {
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	mux := http.NewServeMux()
	server.Register(mux)
	token := loginForTestWithCode(t, mux, "role-grant-realname-user")
	userID := currentUserIDForTest(t, mux, token)

	grant := func() (int, map[string]interface{}) {
		request := httptest.NewRequest(http.MethodPost, "/api/admin/roles/grant", bytes.NewBufferString(`{"userIds":[`+strconv.FormatInt(userID, 10)+`],"roleCode":"guide"}`))
		response := httptest.NewRecorder()
		server.adminGrantRole(response, request)
		var payload map[string]interface{}
		_ = json.Unmarshal(response.Body.Bytes(), &payload)
		return response.Code, payload
	}
	if _, err := server.identity.BindPhone(userID, "13800000003"); err != nil {
		t.Fatal(err)
	}
	if _, err := server.identity.MarkSMSVerified(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := server.identity.VerifyPhone(userID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	if code, payload := grant(); code != http.StatusOK || payload["data"] == nil {
		t.Fatalf("expected per-item realname rejection response, code=%d payload=%v", code, payload)
	}
	if server.profiles.IsGuide(userID) {
		t.Fatal("unverified user must not receive guide role")
	}

	completeIdentityForTest(t, mux, token)
	if code, payload := grant(); code != http.StatusOK || payload["data"] == nil {
		t.Fatalf("expected verified grant response, code=%d payload=%v", code, payload)
	}
	if !server.profiles.IsGuide(userID) {
		t.Fatal("verified user should receive guide role")
	}
	qualification, err := server.profiles.GuideQualification(userID)
	if err != nil || qualification.GuideOpenStatus != "opened" {
		t.Fatalf("guide qualification not opened: %+v err=%v", qualification, err)
	}

	phoneToken := loginForTestWithCode(t, mux, "role-grant-phone-user")
	completeIdentityForTest(t, mux, phoneToken)
	phoneUserID := currentUserIDForTest(t, mux, phoneToken)
	plain, revealErr := server.identity.RevealRecord(server.identity.Status(phoneUserID))
	if revealErr != nil || plain.Phone == "" {
		t.Fatalf("expected realname phone for whitelist import, record=%+v err=%v", server.identity.Status(phoneUserID), revealErr)
	}
	phoneRequest := httptest.NewRequest(http.MethodPost, "/api/admin/roles/grant", bytes.NewBufferString(`{"phoneNumbers":["`+plain.Phone+`"],"roleCode":"expert"}`))
	phoneResponse := httptest.NewRecorder()
	server.adminGrantRole(phoneResponse, phoneRequest)
	if phoneResponse.Code != http.StatusOK || server.profiles.RoleSnapshot(phoneUserID).RoleStatusMap["expert"] != "approved" {
		t.Fatalf("expected phone whitelist import to grant expert, code=%d body=%s", phoneResponse.Code, phoneResponse.Body.String())
	}
}

func TestAdminPermissionUsesAuthenticatedAdminIDInsteadOfRequestHeader(t *testing.T) {
	server := newTestAppServer(auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()), identity.NewService())
	mux := http.NewServeMux()
	server.Register(mux)
	token := adminLoginForTest(t, mux)
	expectedID, permitted := server.admins.HasPermission(token, "role:update")
	if !permitted || expectedID <= 0 {
		t.Fatalf("test administrator must have role update permission, id=%d permitted=%v", expectedID, permitted)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/roles/grant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Admin-ID", "999999")
	rec := httptest.NewRecorder()
	actualID, ok := server.requireAdminPermissionID(rec, req, "role:update")
	if !ok || actualID != expectedID {
		t.Fatalf("expected authenticated admin id %d, got %d ok=%v", expectedID, actualID, ok)
	}
	if req.Header.Get("X-Admin-ID") != strconv.FormatInt(expectedID, 10) {
		t.Fatalf("request header must be normalized to authenticated admin id, got %q", req.Header.Get("X-Admin-ID"))
	}
}
