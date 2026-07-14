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
