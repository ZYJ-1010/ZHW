package appapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"zhw-mini/services/go-api/internal/adminauth"
	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestAdminPendingCountsAggregatesWorkQueue(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	if _, err := server.admins.SubmitAdminApplication(adminauth.SubmitAdminApplicationRequest{
		Name:        "待审核管理员",
		Contact:     "13800138000",
		DesiredRole: "game_manager",
		Reason:      "负责组局审核",
	}); err != nil {
		t.Fatalf("submit admin application: %v", err)
	}
	server.Register(mux)
	adminToken := adminLoginForTest(t, mux)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/pending-counts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected pending counts 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			Counts map[string]int `json:"counts"`
			Total  int            `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Counts["adminApplications"] != 1 || response.Data.Total < 1 {
		t.Fatalf("pending admin application must be included in work queue: %s", rec.Body.String())
	}
}

func TestAdminDashboardAggregatesGamesByPrimaryCategory(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	for _, category := range []string{"task", "explore", "task"} {
		if _, err := server.games.Create(1, games.CreateRequest{
			Title:           category + " game",
			GameType:        "free",
			PrimaryCategory: category,
			MinPlayers:      5,
			MaxPlayers:      8,
			StartAt:         "2026-08-01 10:00",
			EndAt:           "2026-08-01 12:00",
		}); err != nil {
			t.Fatalf("create %s game: %v", category, err)
		}
	}
	server.Register(mux)
	body := getAdminJSON(t, mux, "/api/admin/dashboard", adminLoginForTest(t, mux), http.StatusOK)
	var response struct {
		Data struct {
			Distributions struct {
				GameTypes    map[string]int `json:"gameTypes"`
				GameStatuses map[string]int `json:"gameStatuses"`
			} `json:"distributions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Distributions.GameTypes["task"] != 2 || response.Data.Distributions.GameTypes["explore"] != 1 {
		t.Fatalf("dashboard must aggregate games by primary category: %s", string(body))
	}
	if response.Data.Distributions.GameTypes["free"] != 0 {
		t.Fatalf("dashboard must not aggregate category chart by payment type: %s", string(body))
	}
	for _, status := range []string{
		games.StatusDraft,
		games.StatusPendingAudit,
		games.StatusRecruiting,
		games.StatusFull,
		games.StatusInProgress,
		games.StatusPendingConfirm,
		games.StatusPendingReview,
		games.StatusCompleted,
		games.StatusRejected,
		games.StatusCancelled,
		games.StatusDisputed,
		games.StatusSettling,
		games.StatusClosed,
	} {
		if _, ok := response.Data.Distributions.GameStatuses[status]; !ok {
			t.Fatalf("dashboard game status distribution missing %q: %s", status, string(body))
		}
	}
	if response.Data.Distributions.GameStatuses[games.StatusPendingAudit] != 3 {
		t.Fatalf("dashboard must keep actual status counts: %s", string(body))
	}
}
