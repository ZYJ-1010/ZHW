package appapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestAdminPendingCountsAggregatesWorkQueue(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)
	adminToken := adminLoginForTest(t, mux)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/pending-counts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected pending counts 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
