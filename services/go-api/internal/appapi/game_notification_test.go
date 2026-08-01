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

func TestGameAuditAndApplicationNotificationsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "notification-game-creator")
	applicantToken := loginForTestWithCode(t, mux, "notification-game-applicant")
	rejectedToken := loginForTestWithCode(t, mux, "notification-game-rejected")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, applicantToken)
	completeIdentityForTest(t, mux, rejectedToken)

	createBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"notification lifecycle game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createBody, &created); err != nil {
		t.Fatal(err)
	}
	gameIDText := strconv.FormatInt(created.Data.ID, 10)
	postAdminJSONWithPermission(t, mux, "/api/admin/games/"+gameIDText+"/audit", "game:update_status", `{"approve":true}`, http.StatusOK)

	creatorNotices := getJSON(t, mux, "/api/app/notifications", creatorToken, http.StatusOK)
	if countNotificationsByType(t, creatorNotices, "game_approved") != 1 {
		t.Fatalf("expected one game_approved notification: %s", string(creatorNotices))
	}

	approvedApplicationID := applyForNotificationTest(t, mux, applicantToken, gameIDText)
	rejectedApplicationID := applyForNotificationTest(t, mux, rejectedToken, gameIDText)
	creatorNotices = getJSON(t, mux, "/api/app/notifications", creatorToken, http.StatusOK)
	if countNotificationsByType(t, creatorNotices, "game_apply") != 2 {
		t.Fatalf("expected two game_apply notifications: %s", string(creatorNotices))
	}

	centerBody := getJSON(t, mux, "/api/app/messages/center", creatorToken, http.StatusOK)
	var center struct {
		Data struct {
			Sections []struct {
				Key   string `json:"key"`
				Items []struct {
					ID int64 `json:"id"`
				} `json:"items"`
			} `json:"sections"`
		} `json:"data"`
	}
	if err := json.Unmarshal(centerBody, &center); err != nil {
		t.Fatal(err)
	}
	groupCount := 0
	for _, section := range center.Data.Sections {
		if section.Key == "group" {
			groupCount = len(section.Items)
		}
	}
	if groupCount != 3 {
		t.Fatalf("expected audit and application notifications in group section: %s", string(centerBody))
	}

	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(approvedApplicationID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(rejectedApplicationID, 10)+"/audit", creatorToken, `{"approve":false,"rejectReason":"资料不完整"}`, http.StatusOK)

	approvedNotices := getJSON(t, mux, "/api/app/notifications", applicantToken, http.StatusOK)
	if countNotificationsByType(t, approvedNotices, "application_approved") != 1 {
		t.Fatalf("expected application_approved notification: %s", string(approvedNotices))
	}
	if !strings.Contains(string(approvedNotices), "notification lifecycle game") ||
		!strings.Contains(string(approvedNotices), "已入局") ||
		!strings.Contains(string(approvedNotices), "审核时间") {
		t.Fatalf("approved notification must include game name, result and review time: %s", string(approvedNotices))
	}
	rejectedNotices := getJSON(t, mux, "/api/app/notifications", rejectedToken, http.StatusOK)
	if countNotificationsByType(t, rejectedNotices, "application_rejected") != 1 {
		t.Fatalf("expected application_rejected notification: %s", string(rejectedNotices))
	}
	if !strings.Contains(string(rejectedNotices), "资料不完整") ||
		!strings.Contains(string(rejectedNotices), "未通过") ||
		!strings.Contains(string(rejectedNotices), "审核时间") {
		t.Fatalf("rejection notification must include result, time and reason: %s", string(rejectedNotices))
	}
	approvedCenter := getJSON(t, mux, "/api/app/messages/center", applicantToken, http.StatusOK)
	if !strings.Contains(string(approvedCenter), `"notifyType":"application_approved"`) ||
		!strings.Contains(string(approvedCenter), `"text":"查看组局"`) {
		t.Fatalf("approved result must appear in the message center with a game action: %s", string(approvedCenter))
	}
	approvedMyGames := getJSON(t, mux, "/api/app/games/player/manage", applicantToken, http.StatusOK)
	if !strings.Contains(string(approvedMyGames), `"applicationStatus":"approved"`) ||
		!strings.Contains(string(approvedMyGames), `"statusText":"已入局"`) {
		t.Fatalf("approved result must be synchronized to my joined games: %s", string(approvedMyGames))
	}
	rejectedMyGames := getJSON(t, mux, "/api/app/games/player/manage", rejectedToken, http.StatusOK)
	if !strings.Contains(string(rejectedMyGames), `"applicationStatus":"rejected"`) ||
		!strings.Contains(string(rejectedMyGames), `"statusText":"未通过"`) ||
		!strings.Contains(string(rejectedMyGames), "资料不完整") {
		t.Fatalf("rejected result must be synchronized to my joined games: %s", string(rejectedMyGames))
	}
	received := getJSON(t, mux, "/api/app/game-applications/received?gameId="+gameIDText, creatorToken, http.StatusOK)
	if !strings.Contains(string(received), `"rejectReason":"资料不完整"`) {
		t.Fatalf("rejection reason must be returned in audit detail: %s", string(received))
	}
}

func applyForNotificationTest(t *testing.T, mux *http.ServeMux, token string, gameID string) int64 {
	t.Helper()
	body := postJSON(t, mux, "/api/app/games/"+gameID+"/applications", token, `{"reason":"notification test"}`, http.StatusOK)
	var response struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.ID <= 0 {
		t.Fatalf("expected application id: %s", string(body))
	}
	return response.Data.ID
}
