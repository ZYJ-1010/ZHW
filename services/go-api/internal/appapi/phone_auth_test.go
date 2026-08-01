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

func TestWechatPreAuthBindsPhoneOnlyAfterSMSVerification(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	body := postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"preauth-phone","inviteCode":"TEST2026"}`, http.StatusOK)
	var login struct {
		Data struct {
			PreAuthToken string `json:"preAuthToken"`
			User         struct {
				ID int64 `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &login); err != nil {
		t.Fatal(err)
	}
	if login.Data.PreAuthToken == "" || login.Data.User.ID == 0 {
		t.Fatalf("expected pre-auth login: %s", string(body))
	}
	getJSON(t, mux, "/api/app/home", login.Data.PreAuthToken, http.StatusUnauthorized)
	postJSON(t, mux, "/api/app/identity/phone/bind", login.Data.PreAuthToken, `{"phone":"13800138041"}`, http.StatusOK)
	before, ok := authService.UserByID(login.Data.User.ID)
	if !ok || before.PhoneMasked != "" {
		t.Fatalf("phone must not be written before SMS verification: %+v", before)
	}
	postJSON(t, mux, "/api/app/sms/send-code", login.Data.PreAuthToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/verify-code", login.Data.PreAuthToken, `{"code":"000000"}`, http.StatusOK)
	after, ok := authService.UserByID(login.Data.User.ID)
	if !ok || after.PhoneMasked != "138****8041" {
		t.Fatalf("verified phone must be written to the account: %+v", after)
	}
	formalBody := postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"preauth-phone"}`, http.StatusOK)
	var formal struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(formalBody, &formal); err != nil {
		t.Fatal(err)
	}
	if formal.Data.Token == "" {
		t.Fatalf("verified binding must allow formal WeChat login: %s", string(formalBody))
	}
	getJSON(t, mux, "/api/app/home", formal.Data.Token, http.StatusOK)
}

func TestPhoneLoginPreservesApprovedRealname(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	registered, err := authService.PhoneLogin(auth.PhoneLoginRequest{
		Phone: "13800138042", Code: "000000", InviteCode: "TEST2026", EntryType: invites.EntryTypeLink,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := identityService.SyncPhoneLoginVerification(registered.User.ID, "13800138042"); err != nil {
		t.Fatal(err)
	}
	if _, err := identityService.SubmitManualRealname(registered.User.ID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	approved, err := identityService.ReviewManualRealname(registered.User.ID, true, "")
	if err != nil {
		t.Fatal(err)
	}
	server := newTestAppServer(authService, identityService)
	server.Register(mux)
	postJSON(t, mux, "/api/app/auth/phone-login", "", `{"phone":"13800138042","code":"000000"}`, http.StatusOK)
	after := identityService.Status(registered.User.ID)
	if after.Status != identity.StatusVerified || after.RealNameCiphertext != approved.RealNameCiphertext || after.IDCardCiphertext != approved.IDCardCiphertext {
		t.Fatalf("same-phone login must preserve approved realname: %+v", after)
	}
}
