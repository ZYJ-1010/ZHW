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

func TestNewbieTaskGuideAndCompletionRouteHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "newbie-guide-route")
	completeIdentityForTest(t, mux, token)
	// The client calls /{code}/complete. This must resolve to the configured
	// task code instead of treating "complete" as part of the code.
	postJSON(t, mux, "/api/app/newbie-tasks/complete_identity/complete", token, `{}`, http.StatusOK)

	body := postJSON(t, mux, "/api/app/newbie-tasks/guide-profile-reminder", token, `{}`, http.StatusOK)
	var response struct {
		Data struct {
			Guide struct {
				ProfileReminderCount int  `json:"profileReminderCount"`
				ProfileComplete      bool `json:"profileComplete"`
			} `json:"guide"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Guide.ProfileComplete {
		t.Fatal("user without an avatar must still receive the profile-completion guide")
	}
	if response.Data.Guide.ProfileReminderCount != 1 {
		t.Fatalf("expected first guide display to be recorded once, got %s", string(body))
	}

	// 玩家身份是所有账户的基础身份，不能让它自动完成“申请行家或领路人”。
	postJSON(t, mux, "/api/app/newbie-tasks/apply_role/complete", token, `{}`, http.StatusUnprocessableEntity)
	userID := currentUserIDForTest(t, mux, token)
	server.profiles.GrantRole(userID, "expert")
	postJSON(t, mux, "/api/app/newbie-tasks/apply_role/complete", token, `{}`, http.StatusOK)
}
