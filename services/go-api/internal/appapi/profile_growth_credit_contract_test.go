package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

func TestGrowthProfileUsesRequestedActiveRole(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "growth-active-role-contract")
	userID := currentUserIDForTest(t, mux, token)
	server.profiles.GrantRole(userID, "expert")
	server.profiles.GrantRole(userID, "guide")
	server.reviews.AwardTaskReward(userID, "growth-active-role-contract", 0, 200)

	guideBody := getJSON(t, mux, "/api/app/growth/my?roleType=guide", token, http.StatusOK)
	var guideResponse struct {
		Data struct {
			ActiveRoleCode   string `json:"activeRoleCode"`
			ActiveRoleName   string `json:"activeRoleName"`
			ActiveRoleGrowth struct {
				RoleCode string `json:"roleCode"`
			} `json:"activeRoleGrowth"`
		} `json:"data"`
	}
	if err := json.Unmarshal(guideBody, &guideResponse); err != nil {
		t.Fatal(err)
	}
	if guideResponse.Data.ActiveRoleCode != "guide" || guideResponse.Data.ActiveRoleName != "领路人" || guideResponse.Data.ActiveRoleGrowth.RoleCode != "guide" {
		t.Fatalf("expected requested guide growth profile, got %s", string(guideBody))
	}

	playerBody := getJSON(t, mux, "/api/app/growth/my?roleType=player", token, http.StatusOK)
	var playerResponse struct {
		Data struct {
			ActiveRoleCode   string `json:"activeRoleCode"`
			ActiveRoleGrowth struct {
				RoleCode        string `json:"roleCode"`
				Score           int    `json:"score"`
				ProgressPercent int    `json:"progressPercent"`
				Metrics         []struct {
					Score int `json:"score"`
				} `json:"metrics"`
			} `json:"activeRoleGrowth"`
		} `json:"data"`
	}
	if err := json.Unmarshal(playerBody, &playerResponse); err != nil {
		t.Fatal(err)
	}
	player := playerResponse.Data.ActiveRoleGrowth
	if playerResponse.Data.ActiveRoleCode != "player" || player.RoleCode != "player" || player.Score != 200 || player.ProgressPercent != 50 || len(player.Metrics) != 1 || player.Metrics[0].Score != 50 {
		t.Fatalf("expected player growth to use configured level interval progress, got %s", string(playerBody))
	}
}

func TestProfileHomeUsesRequestedActiveRole(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "profile-home-active-role-contract")
	completeIdentityForTest(t, mux, token)
	userID := currentUserIDForTest(t, mux, token)
	server.profiles.GrantRole(userID, "expert")
	server.profiles.GrantRole(userID, "guide")

	body := getJSON(t, mux, "/api/app/profile/home?roleType=expert", token, http.StatusOK)
	var response struct {
		Data struct {
			User struct {
				Role string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.User.Role != "行家" {
		t.Fatalf("expected profile home to display requested expert role, got %s", string(body))
	}
}

func TestGrowthProfileRejectsInactiveRequestedRole(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "growth-inactive-role-contract")
	body := getJSON(t, mux, "/api/app/growth/my?roleType=guide", token, http.StatusOK)
	if !strings.Contains(string(body), `"activeRoleCode":"player"`) || !strings.Contains(string(body), `"activeRoleName":"玩家"`) {
		t.Fatalf("expected inactive requested role to fall back to player, got %s", string(body))
	}
}

func TestGrowthAchievementConfigIsFilteredByCurrentRole(t *testing.T) {
	config := defaultGrowthAchievementConfig()
	config.Catalog = []growthAchievementItemDTO{
		{ID: "player-only", Code: "player-only", Title: "玩家成就", Roles: []string{"player"}, Visible: true},
		{ID: "guide-only", Code: "guide-only", Title: "领路人成就", Roles: []string{"guide"}, Visible: true},
	}
	config.Locked = append([]growthAchievementItemDTO(nil), config.Catalog...)
	trace := reviews.Trace{Achievements: []reviews.Achievement{
		{Code: "player-only", Title: "玩家成就"},
		{Code: "guide-only", Title: "领路人成就"},
	}}

	achievements := achievementDTOsForRole(trace, config, "player")
	filtered := growthAchievementConfigForRole(config, "player")
	if len(achievements) != 2 || achievements[0].Code != "player-only" || achievements[1].Code != "guide-only" {
		t.Fatalf("expected unlocked historical achievements to remain visible, got %+v", achievements)
	}
	if len(filtered.Catalog) != 1 || filtered.Catalog[0].Code != "player-only" || len(filtered.Locked) != 1 || filtered.Locked[0].Code != "player-only" {
		t.Fatalf("expected player-only achievement config, got %+v", filtered)
	}
}

func TestCreditCenterExcludesSubmittedAppealFromAppealableCount(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "credit-center-appeal-contract")
	userID := currentUserIDForTest(t, mux, token)
	creditLog := server.reviews.DeductCredit(userID, 12, "low_review")

	beforeBody := getJSON(t, mux, "/api/app/profile/credit-center", token, http.StatusOK)
	before := decodeCreditCenterContract(t, beforeBody)
	if !before.Data.AppealEntry.Enabled || before.Data.AppealEntry.Count != 1 || !strings.Contains(before.Data.AppealEntry.Route, strconv.FormatInt(creditLog.ID, 10)) {
		t.Fatalf("expected one appealable deduction before submission, got %s", string(beforeBody))
	}

	postJSON(t, mux, "/api/app/profile/credit-appeals", token, `{"creditLogId":`+strconv.FormatInt(creditLog.ID, 10)+`,"content":"申请复核本次扣分"}`, http.StatusOK)
	afterBody := getJSON(t, mux, "/api/app/profile/credit-center", token, http.StatusOK)
	after := decodeCreditCenterContract(t, afterBody)
	if after.Data.AppealEntry.Enabled || after.Data.AppealEntry.Count != 0 || after.Data.AppealEntry.Route != "" {
		t.Fatalf("expected submitted appeal excluded from appealable count, got %s", string(afterBody))
	}
	found := false
	for _, record := range after.Data.Records {
		if record.CreditLogID != creditLog.ID {
			continue
		}
		found = true
		if record.CanAppeal || record.AppealID <= 0 || record.AppealStatusText != "申诉处理中" || !strings.Contains(record.AppealRoute, "appeal-detail") {
			t.Fatalf("expected deduction to link to existing appeal instead of resubmission, got %+v", record)
		}
	}
	if !found {
		t.Fatalf("expected source deduction in credit center, got %s", string(afterBody))
	}
}

type creditCenterContractResponse struct {
	Data struct {
		AppealEntry struct {
			Enabled bool   `json:"enabled"`
			Count   int    `json:"count"`
			Route   string `json:"route"`
		} `json:"appealEntry"`
		Records []struct {
			CreditLogID      int64  `json:"creditLogId"`
			CanAppeal        bool   `json:"canAppeal"`
			AppealID         int64  `json:"appealId"`
			AppealStatusText string `json:"appealStatusText"`
			AppealRoute      string `json:"appealRoute"`
		} `json:"records"`
	} `json:"data"`
}

func decodeCreditCenterContract(t *testing.T, body []byte) creditCenterContractResponse {
	t.Helper()
	var response creditCenterContractResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode credit center contract: %v: %s", err, string(body))
	}
	return response
}

func TestCreditDisplayHelpersKeepZeroNeutralAndChinese(t *testing.T) {
	if got := signedIntText(0); got != "0" {
		t.Fatalf("expected zero credit change to stay 0, got %q", got)
	}
	if got := creditTone(0); got != "neutral" {
		t.Fatalf("expected zero credit change to use neutral tone, got %q", got)
	}
	if got := creditReasonTitle("player_cancel_service"); got != "玩家取消已确认服务" {
		t.Fatalf("expected Chinese credit reason, got %q", got)
	}
}
