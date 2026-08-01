package appapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestGameMemberRoleAdminGameAlwaysPlayer(t *testing.T) {
	game := games.Game{ID: 1, GameSource: "admin", CreatorUserID: 1, MainGuideUserID: 2}
	roles := map[int64]string{1: "expert", 2: "main_guide", 3: "guide"}
	for _, userID := range []int64{1, 2, 3} {
		if role := gameMemberRole(game, userID, true, roles); role != "member" {
			t.Fatalf("admin game user %d must be player, got %q", userID, role)
		}
	}
}

func TestGameMemberRoleAppGameKeepsPerGameRole(t *testing.T) {
	game := games.Game{ID: 1, GameSource: "app", CreatorUserID: 1, MainGuideUserID: 3}
	roles := map[int64]string{1: "member", 2: "expert", 3: "main_guide", 4: "guide"}
	wants := map[int64]string{1: "member", 2: "expert", 3: "main_guide", 4: "guide"}
	for userID, want := range wants {
		if role := gameMemberRole(game, userID, true, roles); role != want {
			t.Fatalf("app game user %d role = %q, want %q", userID, role, want)
		}
	}
	if role := gameMemberRole(game, 2, false, roles); role != "expert" {
		t.Fatalf("historical canceled role = %q, want expert", role)
	}
}

func TestServiceOrderSummaryCountsCanceledCards(t *testing.T) {
	summary := serviceOrderSummary([]map[string]interface{}{
		{"statusType": "active"},
		{"statusType": "complete"},
		{"statusType": "canceled"},
	}, "summary")
	if summary["activeCount"] != 1 || summary["completeCount"] != 1 || summary["canceledCount"] != 1 || summary["refundCount"] != 1 {
		t.Fatalf("unexpected service order summary: %+v", summary)
	}
}

func TestDisputeManagementTabAlsoReturnsCanceledCards(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/app/games/my/manage?status=dispute", nil)
	items := []map[string]interface{}{
		{"id": "canceled", "statusType": "canceled"},
		{"id": "disputed", "statusType": "dispute"},
		{"id": "active", "statusType": "active"},
	}
	filtered, _, _, total := paginateRoleItems(request, items, "statusType")
	if total != 2 || len(filtered) != 2 {
		t.Fatalf("expected canceled and disputed cards, got total=%d items=%+v", total, filtered)
	}
}

func TestManagementEndpointsUsePerGameRole(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "role-manage-creator")
	expertToken := loginForTestWithCode(t, mux, "role-manage-expert")
	playerToken := loginForTestWithCode(t, mux, "role-manage-player")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, expertToken)
	completeIdentityForTest(t, mux, playerToken)
	expertUser, _ := authService.CurrentUser(expertToken)
	server.profiles.GrantRole(expertUser.ID, "expert")

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"role management","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	applyAndApproveRoleForTest(t, mux, creatorToken, expertToken, "expert")
	applyAndApproveRoleForTest(t, mux, creatorToken, playerToken, "player")
	approveExtraMembersForHTTP(t, mux, creatorToken, 1, "role-management", 2)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	assertManageItemCountForTest(t, mux, expertToken, "/api/app/games/my/manage", 1)
	assertManageItemCountForTest(t, mux, expertToken, "/api/app/games/player/manage", 0)
	assertManageItemCountForTest(t, mux, playerToken, "/api/app/games/player/manage", 1)
	assertManageItemCountForTest(t, mux, playerToken, "/api/app/games/my/manage", 0)

	if _, err := server.games.CancelService(1, "test canceled card"); err != nil {
		t.Fatal(err)
	}
	assertManageItemCountForTest(t, mux, expertToken, "/api/app/games/my/manage?status=dispute", 1)
}

func TestGameApplicationRequiresActivePlatformRole(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "role-apply-creator")
	guideToken := loginForTestWithCode(t, mux, "role-apply-guide")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, guideToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"guide role application","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", guideToken, `{"reason":"join","roleType":"guide"}`, http.StatusForbidden)

	guideUser, _ := authService.CurrentUser(guideToken)
	server.profiles.GrantRole(guideUser.ID, "guide")
	creatorUser, _ := authService.CurrentUser(creatorToken)
	if _, err := authService.SetInviteRelationInviter(creatorUser.ID, guideUser.ID, "test_direct_guide"); err != nil {
		t.Fatal(err)
	}
	body := postJSON(t, mux, "/api/app/games/1/applications", guideToken, `{"reason":"join","roleType":"guide"}`, http.StatusOK)
	var application struct {
		Data struct {
			ID   int64  `json:"id"`
			Role string `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &application); err != nil {
		t.Fatal(err)
	}
	if application.Data.Role != "guide" {
		t.Fatalf("expected guide application role, got %s: %s", application.Data.Role, string(body))
	}

	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(application.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	detailBody := getJSON(t, mux, "/api/app/games/1", guideToken, http.StatusOK)
	if !strings.Contains(string(detailBody), `"role":"main_guide"`) {
		t.Fatalf("expected approved guide to become main guide: %s", string(detailBody))
	}
}

func TestGuideEscortApplicationIsUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "escort-creator")
	guideToken := loginForTestWithCode(t, mux, "escort-guide")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, guideToken)
	guideUser, _ := authService.CurrentUser(guideToken)
	server.profiles.GrantRole(guideUser.ID, "guide")

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"escort game","gameType":"free","minPlayers":5,"maxPlayers":5,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00","allowedRoles":["player"],"allowGuideEscort":true}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", guideToken, `{"reason":"escort","roleType":"guide_escort"}`, http.StatusUnprocessableEntity)
}

func applyAndApproveRoleForTest(t *testing.T, mux *http.ServeMux, creatorToken string, entrantToken string, roleType string) {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games/1/applications", entrantToken, `{"reason":"join","roleType":"`+roleType+`"}`, http.StatusOK)
	var response struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(response.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
}

func assertManageItemCountForTest(t *testing.T, mux *http.ServeMux, token string, path string, want int) {
	t.Helper()
	body := getJSON(t, mux, path, token, http.StatusOK)
	var response struct {
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data.Items) != want {
		t.Fatalf("%s item count = %d, want %d: %s", path, len(response.Data.Items), want, string(body))
	}
}
