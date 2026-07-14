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

	createBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"notification lifecycle game","gameType":"free","minPlayers":5,"maxPlayers":8}`, http.StatusOK)
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
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(rejectedApplicationID, 10)+"/audit", creatorToken, `{"approve":false}`, http.StatusOK)

	approvedNotices := getJSON(t, mux, "/api/app/notifications", applicantToken, http.StatusOK)
	if countNotificationsByType(t, approvedNotices, "application_approved") != 1 {
		t.Fatalf("expected application_approved notification: %s", string(approvedNotices))
	}
	rejectedNotices := getJSON(t, mux, "/api/app/notifications", rejectedToken, http.StatusOK)
	if countNotificationsByType(t, rejectedNotices, "application_rejected") != 1 {
		t.Fatalf("expected application_rejected notification: %s", string(rejectedNotices))
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
