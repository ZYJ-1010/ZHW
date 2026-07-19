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
}
