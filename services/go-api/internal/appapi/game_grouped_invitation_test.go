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
	"zhw-mini/services/go-api/internal/users"
)

func TestGroupedInvitationPlayerThenExpertConfirmationHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	guideToken := loginForTestWithCode(t, mux, "grouped-guide")
	completeIdentityForTest(t, mux, guideToken)
	playerToken := loginForTestWithCode(t, mux, "grouped-player")
	completeIdentityForTest(t, mux, playerToken)
	expertToken := loginForTestWithCode(t, mux, "grouped-expert")
	completeIdentityForTest(t, mux, expertToken)
	server.profiles.GrantRole(3, "expert")

	postJSON(t, mux, "/api/app/games", guideToken, `{"title":"grouped confirmation","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2030-01-01 10:00","endAt":"2030-01-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", guideToken, `{}`, http.StatusOK)

	create := func(payload string) int64 {
		body := postJSON(t, mux, "/api/app/games/1/guide-invitations", guideToken, payload, http.StatusOK)
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

	playerInvitationID := create(`{"targetUserId":2,"expertUserId":3,"inviteGroupId":"paired-http","roleType":"player","message":"player first"}`)
	expertInvitationID := create(`{"targetUserId":3,"playerUserId":2,"inviteGroupId":"paired-http","roleType":"expert","message":"expert second"}`)

	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(expertInvitationID, 10)+"/respond", expertToken, `{"accept":true}`, http.StatusConflict)
	playerBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(playerInvitationID, 10)+"/respond", playerToken, `{"accept":true}`, http.StatusOK)
	if !strings.Contains(string(playerBody), `"status":"approved"`) {
		t.Fatalf("player confirmation should approve directly: %s", string(playerBody))
	}
	expertNotices := getJSON(t, mux, "/api/app/notifications?type=game_invitation_progress", expertToken, http.StatusOK)
	if countNotificationsByType(t, expertNotices, "game_invitation_progress") != 1 ||
		!strings.Contains(string(expertNotices), `"title":"玩家已确认，等待你的确认"`) ||
		!strings.Contains(string(expertNotices), `pages/game/audit-detail/index?invitationId=`+strconv.FormatInt(expertInvitationID, 10)) {
		t.Fatalf("expert should receive a confirmation-ready notification: %s", string(expertNotices))
	}

	progressBody := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(expertInvitationID, 10), expertToken, http.StatusOK)
	if !strings.Contains(string(progressBody), `"canConfirm":true`) ||
		!strings.Contains(string(progressBody), `"requirementConfirmed":true`) ||
		!strings.Contains(string(progressBody), `"requirementStatusText":"已确认"`) ||
		strings.Contains(string(progressBody), "请等待玩家先确认组局") {
		t.Fatalf("expert should be enabled after player confirmation: %s", string(progressBody))
	}

	expertBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(expertInvitationID, 10)+"/respond", expertToken, `{"accept":true}`, http.StatusOK)
	if !strings.Contains(string(expertBody), `"status":"approved"`) || !strings.Contains(string(expertBody), `pages/game/success-expert/index?gameId=1`) {
		t.Fatalf("expert confirmation should complete the group: %s", string(expertBody))
	}
}

func TestGroupedInvitationMainGuideReadsPlayerAutoApprovalHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "grouped-owner")
	completeIdentityForTest(t, mux, creatorToken)
	guideToken := loginForTestWithCode(t, mux, "grouped-main-guide")
	completeIdentityForTest(t, mux, guideToken)
	playerToken := loginForTestWithCode(t, mux, "grouped-guide-player")
	completeIdentityForTest(t, mux, playerToken)
	expertToken := loginForTestWithCode(t, mux, "grouped-guide-expert")
	completeIdentityForTest(t, mux, expertToken)
	server.profiles.GrantRole(4, "expert")

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"main guide grouped confirmation","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2030-01-01 10:00","endAt":"2030-01-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	guideInviteBody := postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":2,"message":"lead this game"}`, http.StatusOK)
	var guideInviteResponse struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(guideInviteBody, &guideInviteResponse); err != nil {
		t.Fatal(err)
	}
	guideAcceptBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(guideInviteResponse.Data.Invitation.ID, 10)+"/respond", guideToken, `{"accept":true}`, http.StatusOK)
	var guideAcceptResponse struct {
		Data struct {
			Application struct {
				ID int64 `json:"id"`
			} `json:"application"`
		} `json:"data"`
	}
	if err := json.Unmarshal(guideAcceptBody, &guideAcceptResponse); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(guideAcceptResponse.Data.Application.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)

	create := func(payload string) int64 {
		body := postJSON(t, mux, "/api/app/games/1/guide-invitations", guideToken, payload, http.StatusOK)
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
	playerInvitationID := create(`{"targetUserId":3,"expertUserId":4,"inviteGroupId":"main-guide-paired","roleType":"player","message":"player first"}`)
	expertInvitationID := create(`{"targetUserId":4,"playerUserId":3,"inviteGroupId":"main-guide-paired","roleType":"expert","message":"expert second"}`)

	playerBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(playerInvitationID, 10)+"/respond", playerToken, `{"accept":true}`, http.StatusOK)
	if !strings.Contains(string(playerBody), `"status":"approved"`) {
		t.Fatalf("main guide invitation should auto-approve the player: %s", string(playerBody))
	}
	progressBody := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(expertInvitationID, 10), expertToken, http.StatusOK)
	if !strings.Contains(string(progressBody), `"canConfirm":true`) ||
		!strings.Contains(string(progressBody), `"requirementStatusText":"已确认"`) ||
		strings.Contains(string(progressBody), "请等待玩家先确认组局") {
		t.Fatalf("expert should see the main guide's player as confirmed: %s", string(progressBody))
	}
	assertAvatarCount := func(expected int) {
		listBody := getJSON(t, mux, "/api/app/games", playerToken, http.StatusOK)
		var listResponse struct {
			Data struct {
				Items []struct {
					ID            int64 `json:"id"`
					PlayerAvatars []struct {
						Name      string `json:"name"`
						AvatarURL string `json:"avatarUrl"`
					} `json:"playerAvatars"`
				} `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(listBody, &listResponse); err != nil {
			t.Fatal(err)
		}
		if len(listResponse.Data.Items) != 1 || len(listResponse.Data.Items[0].PlayerAvatars) != expected {
			t.Fatalf("expected %d nickname avatars: %s", expected, string(listBody))
		}
		for _, avatar := range listResponse.Data.Items[0].PlayerAvatars {
			if avatar.Name == "" {
				t.Fatalf("joined member must include nickname: %s", string(listBody))
			}
		}
	}
	assertAvatarCount(3)
	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(expertInvitationID, 10)+"/respond", expertToken, `{"accept":true}`, http.StatusOK)
	assertAvatarCount(3)
}
