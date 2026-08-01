package appapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestGameDraftIDFromPathRejectsInvalidPaths(t *testing.T) {
	tests := map[string]int64{
		"/api/app/game-drafts/12":        12,
		"/api/app/game-drafts/12/":       12,
		"/api/app/game-drafts/":          0,
		"/api/app/game-drafts/not-id":    0,
		"/api/app/game-drafts/12/detail": 0,
		"/api/app/game-drafts/-1":        0,
	}
	for path, expected := range tests {
		if actual := gameDraftIDFromPath(path); actual != expected {
			t.Fatalf("path %q: expected %d, got %d", path, expected, actual)
		}
	}
}

func TestGameDraftSaveRequiresJSONObjectPayload(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTest(t, mux)

	postJSON(t, mux, "/api/app/game-drafts", token, `{"title":"数组草稿","payload":[]}`, http.StatusUnprocessableEntity)
	postJSON(t, mux, "/api/app/game-drafts", token, `{"title":"空草稿","payload":null}`, http.StatusUnprocessableEntity)
	postJSON(t, mux, "/api/app/game-drafts", token, `{"title":"正常草稿","payload":{"form":{}}}`, http.StatusOK)

	req := httptest.NewRequest(http.MethodPost, "/api/app/game-drafts", strings.NewReader(`{"title":"尾随内容","payload":{"form":{}}}{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected trailing JSON to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
