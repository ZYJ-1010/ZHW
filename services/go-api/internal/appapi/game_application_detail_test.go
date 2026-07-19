package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestReceivedApplicationIncludesAuditDetailDisplayHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	creatorToken := loginForTestWithCode(t, mux, "audit-detail-creator")
	playerToken := loginForTestWithCode(t, mux, "audit-detail-player")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, playerToken)

	createdBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"真实审核详情局","gameType":"free","address":"真实地点","startAt":"2026-07-12T10:00:00+08:00","endAt":"2026-07-12T12:00:00+08:00","minPlayers":5,"maxPlayers":8}`, http.StatusOK)
	var created struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createdBody, &created); err != nil {
		t.Fatal(err)
	}
	gameID := strconv.FormatInt(created.Data.ID, 10)
	postJSON(t, mux, "/api/app/games/"+gameID+"/approve-local", creatorToken, `{}`, http.StatusOK)
	applyForNotificationTest(t, mux, playerToken, gameID)

	body := getJSON(t, mux, "/api/app/game-applications/received?gameId="+gameID, creatorToken, http.StatusOK)
	var response struct {
		Data struct {
			Items []struct {
				Nickname      string `json:"nickname"`
				GameTitle     string `json:"gameTitle"`
				LocationText  string `json:"locationText"`
				DetailDisplay struct {
					PageTitle string `json:"pageTitle"`
					Scenario  string `json:"scenario"`
					GameInfo  struct {
						Topic    string `json:"topic"`
						Time     string `json:"time"`
						Location string `json:"location"`
					} `json:"gameInfo"`
					SessionInfo []map[string]interface{} `json:"sessionInfo"`
					ConfirmRows []map[string]interface{} `json:"confirmRows"`
				} `json:"detailDisplay"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("expected one received application: %s", string(body))
	}
	item := response.Data.Items[0]
	if item.Nickname == "" || item.GameTitle != "真实审核详情局" || item.LocationText != "真实地点" {
		t.Fatalf("expected enriched applicant and game fields: %s", string(body))
	}
	if item.DetailDisplay.PageTitle == "" || item.DetailDisplay.Scenario != "direct_application" || item.DetailDisplay.GameInfo.Topic != "真实审核详情局" || item.DetailDisplay.GameInfo.Time == "" || item.DetailDisplay.GameInfo.Location != "真实地点" || len(item.DetailDisplay.SessionInfo) != 3 || len(item.DetailDisplay.ConfirmRows) != 3 {
		t.Fatalf("expected complete audit detail display: %s", string(body))
	}
}

func TestInvitationProgressSeparatesExpertAndPlayerViewsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	creatorToken := loginForTestWithCode(t, mux, "invite-view-creator")
	expertToken := loginForTestWithCode(t, mux, "invite-view-expert")
	playerToken := loginForTestWithCode(t, mux, "invite-view-player")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, expertToken)
	completeIdentityForTest(t, mux, playerToken)
	gameID := createApprovedGameForInvitationViewTest(t, mux, creatorToken, "角色邀请详情局")
	// 普通用户没有邀请功能，后端必须拒绝生成邀请。
	postJSON(t, mux, "/api/app/games/"+gameID+"/guide-invitations", creatorToken, `{"targetUserId":1,"role":"expert"}`, http.StatusForbidden)
	postJSON(t, mux, "/api/app/games/"+gameID+"/guide-invitations", playerToken, `{"targetUserId":1,"role":"player"}`, http.StatusForbidden)
}

func createApprovedGameForInvitationViewTest(t *testing.T, mux *http.ServeMux, token string, title string) string {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games", token, `{"title":"`+title+`","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var response struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	id := strconv.FormatInt(response.Data.ID, 10)
	postJSON(t, mux, "/api/app/games/"+id+"/approve-local", token, `{}`, http.StatusOK)
	return id
}

func createRoleInvitationForTest(t *testing.T, mux *http.ServeMux, token string, gameID string, targetUserID int64, role string) int64 {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games/"+gameID+"/guide-invitations", token, `{"targetUserId":`+strconv.FormatInt(targetUserID, 10)+`,"role":"`+role+`","message":"角色确认"}`, http.StatusOK)
	var response struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	return response.Data.Invitation.ID
}

func assertInvitationViewForTest(t *testing.T, mux *http.ServeMux, token string, invitationID int64, expectedRole string, expectedPageTitle string) {
	t.Helper()
	body := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(invitationID, 10), token, http.StatusOK)
	var response struct {
		Data struct {
			ActiveParties []struct {
				InvitationID  int64  `json:"invitationId"`
				TargetRole    string `json:"targetRole"`
				DetailDisplay struct {
					PageTitle string `json:"pageTitle"`
					Scenario  string `json:"scenario"`
				} `json:"detailDisplay"`
			} `json:"activeParties"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	for _, item := range response.Data.ActiveParties {
		if item.InvitationID == invitationID {
			if item.TargetRole != expectedRole || item.DetailDisplay.PageTitle != expectedPageTitle || item.DetailDisplay.Scenario != "invitation_"+expectedRole {
				t.Fatalf("unexpected role-specific invitation view: %s", string(body))
			}
			return
		}
	}
	t.Fatalf("invitation not found: %s", string(body))
}
