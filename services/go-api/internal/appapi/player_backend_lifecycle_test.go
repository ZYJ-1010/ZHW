package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/im"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/lbs"
	"zhw-mini/services/go-api/internal/users"
)

// TestPlayerBackendIsolatedLifecycleHTTP intentionally uses only fresh in-memory
// stores. It must never be wired to SQL repositories or existing user data.
func TestPlayerBackendIsolatedLifecycleHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	gameService := games.NewService(identityService)
	server := New(authService, identityService, gameService, lbs.NewService(), im.NewService(gameService))
	server.Register(mux)

	adminToken := adminLoginForTest(t, mux)
	ownerToken := loginForTestWithCode(t, mux, "isolated-owner")
	completeIdentityForTest(t, mux, ownerToken)
	ownerID := currentUserIDForTest(t, mux, ownerToken)

	inviteBody := postJSON(t, mux, "/api/app/invites/entries", ownerToken, `{"entryType":"link","title":"isolated lifecycle invite"}`, http.StatusOK)
	var inviteResp struct {
		Data struct {
			InviteCode string `json:"inviteCode"`
			Path       string `json:"path"`
		} `json:"data"`
	}
	mustDecodeLifecycle(t, inviteBody, &inviteResp)
	if inviteResp.Data.InviteCode == "" || inviteResp.Data.Path == "" {
		t.Fatalf("expected generated invite entry: %s", string(inviteBody))
	}

	invitedToken := loginForTestWithCodeAndInvite(t, mux, "isolated-invited-player", inviteResp.Data.InviteCode)
	completeIdentityForTest(t, mux, invitedToken)
	invitedID := currentUserIDForTest(t, mux, invitedToken)
	if invitedID == ownerID {
		t.Fatal("invited registration must create a distinct isolated user")
	}
	precheckBody := postJSON(t, mux, "/api/app/invites/precheck", "", `{"inviteCode":"`+inviteResp.Data.InviteCode+`","entryType":"link"}`, http.StatusOK)
	if !jsonContainsLifecycle(precheckBody, `"boundWechat":true`) || !jsonContainsLifecycle(precheckBody, `"authPageMode":"login"`) {
		t.Fatalf("expected invite to become bound after registration: %s", string(precheckBody))
	}

	guideToken := loginForTestWithCode(t, mux, "isolated-guide")
	memberAToken := loginForTestWithCode(t, mux, "isolated-member-a")
	memberBToken := loginForTestWithCode(t, mux, "isolated-member-b")
	for _, token := range []string{guideToken, memberAToken, memberBToken} {
		completeIdentityForTest(t, mux, token)
	}
	guideID := currentUserIDForTest(t, mux, guideToken)

	// Exercise both role applications before the game flow.
	expertAppBody := postJSON(t, mux, "/api/app/role-applications", ownerToken, `{"roleCode":"expert","reason":"isolated expert application","abilityDescription":"lifecycle test"}`, http.StatusOK)
	expertAppID := lifecycleResponseID(t, expertAppBody)
	postAdminJSON(t, mux, "/api/admin/audits/role-applications/"+strconv.FormatInt(expertAppID, 10)+"/review", adminToken, `{"approve":true,"remark":"isolated approval"}`, http.StatusOK)

	postAdminJSON(t, mux, "/api/admin/guide-qualification-rules", adminToken, `{"userId":`+strconv.FormatInt(guideID, 10)+`,"conditionMet":true,"paymentMet":true}`, http.StatusOK)
	guideAppBody := postJSON(t, mux, "/api/app/guides/apply", guideToken, `{"reason":"isolated guide application","abilityDescription":"lead isolated games"}`, http.StatusOK)
	guideAppID := lifecycleResponseID(t, guideAppBody)
	postAdminJSON(t, mux, "/api/admin/audits/role-applications/"+strconv.FormatInt(guideAppID, 10)+"/review", adminToken, `{"approve":true,"remark":"isolated approval"}`, http.StatusOK)

	firstGameID := lifecycleCreateAndApproveGame(t, mux, ownerToken, adminToken, "isolated owner game", "2026-08-01 10:00", "2026-08-01 12:00")
	lifecycleApplyAndApprove(t, mux, invitedToken, ownerToken, firstGameID)
	lifecycleInviteGuideAndApprove(t, mux, ownerToken, guideToken, guideID, firstGameID)
	lifecycleApplyAndApprove(t, mux, memberAToken, ownerToken, firstGameID)
	lifecycleApplyAndApprove(t, mux, memberBToken, ownerToken, firstGameID)

	startBody := postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(firstGameID, 10)+"/manual-start", guideToken, `{}`, http.StatusOK)
	if !jsonContainsLifecycle(startBody, `"status":"in_progress"`) {
		t.Fatalf("expected guide to start first game: %s", string(startBody))
	}
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(firstGameID, 10)+"/chat/messages", invitedToken, `{"messageType":"text","content":"isolated lifecycle hello"}`, http.StatusOK)

	pointsBefore := lifecyclePoints(t, mux, invitedToken)
	creditBefore := lifecycleCredit(t, mux, guideToken)
	lifecycleFinishGame(t, mux, firstGameID, ownerToken, []string{invitedToken, memberAToken, memberBToken})
	postJSON(t, mux, "/api/app/reviews", invitedToken, `{"gameId":`+strconv.FormatInt(firstGameID, 10)+`,"targetUserId":`+strconv.FormatInt(ownerID, 10)+`,"targetRole":"member","score":5,"content":"isolated good review","againIntent":"yes"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/reviews", memberAToken, `{"gameId":`+strconv.FormatInt(firstGameID, 10)+`,"targetUserId":`+strconv.FormatInt(guideID, 10)+`,"targetRole":"member","score":1,"content":"isolated low review","againIntent":"no"}`, http.StatusOK)
	if after := lifecyclePoints(t, mux, invitedToken); after <= pointsBefore {
		t.Errorf("expected review points to increase, before=%d after=%d", pointsBefore, after)
	}
	if after := lifecycleCredit(t, mux, guideToken); after >= creditBefore {
		t.Errorf("expected low review to reduce guide credit, before=%d after=%d", creditBefore, after)
	}

	// The user who registered through the generated invite repeats the group flow.
	secondGameID := lifecycleCreateAndApproveGame(t, mux, invitedToken, adminToken, "isolated invited user game", "2026-08-02 10:00", "2026-08-02 12:00")
	lifecycleApplyAndApprove(t, mux, ownerToken, invitedToken, secondGameID)
	lifecycleInviteGuideAndApprove(t, mux, invitedToken, guideToken, guideID, secondGameID)
	lifecycleApplyAndApprove(t, mux, memberAToken, invitedToken, secondGameID)
	lifecycleApplyAndApprove(t, mux, memberBToken, invitedToken, secondGameID)
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(secondGameID, 10)+"/manual-start", guideToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(secondGameID, 10)+"/chat/messages", ownerToken, `{"messageType":"text","content":"isolated second lifecycle hello"}`, http.StatusOK)
	lifecycleFinishGame(t, mux, secondGameID, invitedToken, []string{ownerToken, memberAToken, memberBToken})
}

func lifecycleCreateAndApproveGame(t *testing.T, mux *http.ServeMux, creatorToken, adminToken, title, startAt, endAt string) int64 {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"`+title+`","gameType":"free","minPlayers":5,"maxPlayers":8,"cityCode":"330100","cityName":"Hangzhou","startAt":"`+startAt+`","endAt":"`+endAt+`"}`, http.StatusOK)
	gameID := lifecycleResponseID(t, body)
	postAdminJSON(t, mux, "/api/admin/games/"+strconv.FormatInt(gameID, 10)+"/audit", adminToken, `{"approve":true,"remark":"isolated approval"}`, http.StatusOK)
	return gameID
}

func lifecycleApplyAndApprove(t *testing.T, mux *http.ServeMux, playerToken, ownerToken string, gameID int64) {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/applications", playerToken, `{"reason":"isolated join"}`, http.StatusOK)
	applicationID := lifecycleResponseID(t, body)
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(applicationID, 10)+"/audit", ownerToken, `{"approve":true}`, http.StatusOK)
}

func lifecycleInviteGuideAndApprove(t *testing.T, mux *http.ServeMux, ownerToken, guideToken string, guideID, gameID int64) {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/guide-invitations", ownerToken, `{"targetUserId":`+strconv.FormatInt(guideID, 10)+`,"message":"lead isolated game"}`, http.StatusOK)
	var response struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	mustDecodeLifecycle(t, body, &response)
	respondBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(response.Data.Invitation.ID, 10)+"/respond", guideToken, `{"accept":true,"reason":"isolated accept"}`, http.StatusOK)
	var respond struct {
		Data struct {
			Application struct {
				ID int64 `json:"id"`
			} `json:"application"`
		} `json:"data"`
	}
	mustDecodeLifecycle(t, respondBody, &respond)
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(respond.Data.Application.ID, 10)+"/audit", ownerToken, `{"approve":true}`, http.StatusOK)
}

func lifecycleFinishGame(t *testing.T, mux *http.ServeMux, gameID int64, ownerToken string, memberTokens []string) {
	t.Helper()
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/service-confirm", ownerToken, `{"note":"isolated finished"}`, http.StatusOK)
	var last []byte
	for _, token := range memberTokens {
		last = postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/service-confirm-items", token, `{"note":"isolated confirmed"}`, http.StatusOK)
	}
	if !jsonContainsLifecycle(last, `"status":"pending_review"`) {
		t.Fatalf("expected pending_review after all confirmations: %s", string(last))
	}
}

func lifecycleResponseID(t *testing.T, body []byte) int64 {
	t.Helper()
	var response struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	mustDecodeLifecycle(t, body, &response)
	if response.Data.ID == 0 {
		t.Fatalf("expected response id: %s", string(body))
	}
	return response.Data.ID
}

func lifecyclePoints(t *testing.T, mux *http.ServeMux, token string) int {
	t.Helper()
	body := getJSON(t, mux, "/api/app/points/summary", token, http.StatusOK)
	var response struct {
		Data struct {
			AvailablePoints int `json:"availablePoints"`
		} `json:"data"`
	}
	mustDecodeLifecycle(t, body, &response)
	return response.Data.AvailablePoints
}

func lifecycleCredit(t *testing.T, mux *http.ServeMux, token string) int {
	t.Helper()
	body := getJSON(t, mux, "/api/app/profile/credit-center", token, http.StatusOK)
	var response struct {
		Data struct {
			Score int `json:"score"`
		} `json:"data"`
	}
	mustDecodeLifecycle(t, body, &response)
	return response.Data.Score
}

func mustDecodeLifecycle(t *testing.T, body []byte, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("decode response: %v: %s", err, string(body))
	}
}

func jsonContainsLifecycle(body []byte, value string) bool {
	return strings.Contains(string(body), value)
}
