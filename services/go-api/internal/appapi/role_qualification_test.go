package appapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestRoleApplicationEligibilityRejectsUnmetConditionsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "role-qualification-user")
	completeIdentityForTest(t, mux, token)

	configBody := getJSON(t, mux, "/api/app/role-applications/expert/config", token, http.StatusOK)
	var config struct {
		Data struct {
			BaseEligible bool `json:"baseEligible"`
			Requirements []struct {
				Key     string `json:"key"`
				Checked bool   `json:"checked"`
			} `json:"requirements"`
		} `json:"data"`
	}
	if err := json.Unmarshal(configBody, &config); err != nil {
		t.Fatal(err)
	}
	if config.Data.BaseEligible || len(config.Data.Requirements) != 6 {
		t.Fatalf("expected unmet expert conditions in config: %s", string(configBody))
	}
	for _, requirement := range config.Data.Requirements {
		if (requirement.Key == "level" || requirement.Key == "enterprise" || requirement.Key == "created_games") && requirement.Checked {
			t.Fatalf("unmet requirement should not be checked: %+v", requirement)
		}
	}

	expertBody := postJSON(t, mux, "/api/app/role-applications", token, `{"roleCode":"expert","reason":"申请行家","abilityDescription":"申请说明"}`, http.StatusConflict)
	if !strings.Contains(string(expertBody), "未满足申请条件") || !strings.Contains(string(expertBody), "玩家等级") {
		t.Fatalf("expected expert eligibility error: %s", string(expertBody))
	}
	guideBody := postJSON(t, mux, "/api/app/role-applications", token, `{"roleCode":"guide","reason":"申请领路人","abilityDescription":"申请说明"}`, http.StatusConflict)
	if !strings.Contains(string(guideBody), "未满足申请条件") || !strings.Contains(string(guideBody), "玩家等级") {
		t.Fatalf("expected guide eligibility error: %s", string(guideBody))
	}
}
