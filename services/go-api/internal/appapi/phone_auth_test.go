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

func TestPhoneLoginRegistersAndReturnsFormalTokenHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	body := postJSON(t, mux, "/api/app/auth/phone-login", "", `{
  "phone":"13800138000",
  "code":"000000",
  "inviteCode":"TEST2026",
  "entryType":"link"
}`, http.StatusOK)
	var response struct {
		Code int `json:"code"`
		Data struct {
			Token                   string `json:"token"`
			AuthPageMode            string `json:"authPageMode"`
			RequiresIdentityBinding bool   `json:"requiresIdentityBinding"`
			IdentityBindStatus      string `json:"identityBindStatus"`
			User                    struct {
				ID          int64  `json:"id"`
				PhoneMasked string `json:"phoneMasked"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != 0 || response.Data.Token == "" || response.Data.User.ID == 0 {
		t.Fatalf("expected phone login token: %s", string(body))
	}
	if response.Data.AuthPageMode != "register" || response.Data.RequiresIdentityBinding || response.Data.IdentityBindStatus != "sms_verified" {
		t.Fatalf("expected completed phone registration: %s", string(body))
	}
	if response.Data.User.PhoneMasked != "138****8000" {
		t.Fatalf("expected masked phone: %s", string(body))
	}
	getJSON(t, mux, "/api/app/users/me", response.Data.Token, http.StatusOK)

	second := postJSON(t, mux, "/api/app/auth/phone-login", "", `{
  "phone":"13800138000",
  "code":"000000",
  "inviteCode":"TEST2026",
  "entryType":"link"
}`, http.StatusOK)
	var secondResponse struct {
		Data struct {
			AuthPageMode string `json:"authPageMode"`
			User         struct {
				ID int64 `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(second, &secondResponse); err != nil {
		t.Fatal(err)
	}
	if secondResponse.Data.AuthPageMode != "login" || secondResponse.Data.User.ID != response.Data.User.ID {
		t.Fatalf("expected existing phone login: %s", string(second))
	}

	third := postJSON(t, mux, "/api/app/auth/phone-login", "", `{
  "phone":"13800138000",
  "code":"000000"
}`, http.StatusOK)
	var thirdResponse struct {
		Data struct {
			AuthPageMode string `json:"authPageMode"`
			User         struct {
				ID int64 `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(third, &thirdResponse); err != nil {
		t.Fatal(err)
	}
	if thirdResponse.Data.AuthPageMode != "login" || thirdResponse.Data.User.ID != response.Data.User.ID {
		t.Fatalf("expected existing phone login without invite: %s", string(third))
	}
}

func TestPhoneLoginRejectsInvalidTemporaryCodeHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)
	postJSON(t, mux, "/api/app/auth/phone-login", "", `{
  "phone":"13800138000",
  "code":"123456",
  "inviteCode":"TEST2026"
}`, http.StatusUnprocessableEntity)
}
