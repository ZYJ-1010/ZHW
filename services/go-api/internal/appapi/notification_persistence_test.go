package appapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/users"
)

func TestCriticalGameNotificationPersistenceFailureIsObservable(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "notification-persist-creator")
	applicantToken := loginForTestWithCode(t, mux, "notification-persist-applicant")
	completeIdentityForTest(t, mux, creatorToken)
	completeIdentityForTest(t, mux, applicantToken)
	adminToken := adminLoginForTest(t, mux)

	createdBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"notification persistence game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createdBody, &created); err != nil {
		t.Fatal(err)
	}
	gameID := strconv.FormatInt(created.Data.ID, 10)

	persistErr := errors.New("notification database unavailable")
	server.notices = notifications.NewServiceWithRepository(alwaysFailNotificationRepository{err: persistErr})

	auditRec := serveNotificationPersistencePost(mux, "/api/admin/games/"+gameID+"/audit", adminToken, `{"approve":true}`)
	assertNotificationDegradedResponse(t, auditRec, "game audit")
	var auditResult struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(auditRec.Body.Bytes(), &auditResult); err != nil {
		t.Fatal(err)
	}
	if auditResult.Data.Status != "recruiting" {
		t.Fatalf("audit mutation must remain successful, got %s", auditRec.Body.String())
	}

	applyPath := "/api/app/games/" + gameID + "/applications"
	applyBody := completeApplicationBodyForTest(applyPath, `{"reason":"notification persistence test"}`, http.StatusOK)
	applyRec := serveNotificationPersistencePost(mux, applyPath, applicantToken, applyBody)
	assertNotificationDegradedResponse(t, applyRec, "game application")
	var application struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(applyRec.Body.Bytes(), &application); err != nil {
		t.Fatal(err)
	}
	if application.Data.ID <= 0 {
		t.Fatalf("application mutation must remain successful, got %s", applyRec.Body.String())
	}

	reviewPath := "/api/app/game-applications/" + strconv.FormatInt(application.Data.ID, 10) + "/audit"
	reviewRec := serveNotificationPersistencePost(mux, reviewPath, creatorToken, `{"approve":true}`)
	assertNotificationDegradedResponse(t, reviewRec, "application review")
	var reviewed struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reviewRec.Body.Bytes(), &reviewed); err != nil {
		t.Fatal(err)
	}
	if reviewed.Data.Status != "approved" {
		t.Fatalf("application review mutation must remain successful, got %s", reviewRec.Body.String())
	}

	reportRec := serveNotificationPersistencePost(mux, "/api/app/reports", creatorToken, `{"gameId":`+gameID+`,"reportType":"other","content":"notification persistence report"}`)
	assertNotificationDegradedResponse(t, reportRec, "report creation")
	var report struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reportRec.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Data.ID <= 0 {
		t.Fatalf("report mutation must remain successful, got %s", reportRec.Body.String())
	}
	reportHandleRec := serveNotificationPersistencePost(mux, "/api/admin/reports/"+strconv.FormatInt(report.Data.ID, 10)+"/handle", adminToken, `{"result":"已核实","outcome":"unconfirmed"}`)
	assertNotificationDegradedResponse(t, reportHandleRec, "report handling")

	notificationReq := httptest.NewRequest(http.MethodGet, "/api/app/notifications", nil)
	notificationReq.Header.Set("Authorization", "Bearer "+creatorToken)
	notificationRec := httptest.NewRecorder()
	mux.ServeHTTP(notificationRec, notificationReq)
	if notificationRec.Code != http.StatusInternalServerError || !bytes.Contains(notificationRec.Body.Bytes(), []byte("读取消息失败")) {
		t.Fatalf("message center must not expose process-local fallback after database failure, got %d: %s", notificationRec.Code, notificationRec.Body.String())
	}

	for _, endpoint := range []string{"/api/app/home", "/api/app/profile/home", "/api/app/users/me/summary"} {
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req.Header.Set("Authorization", "Bearer "+creatorToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("%s must not expose a false zero unread count, got %d: %s", endpoint, rec.Code, rec.Body.String())
		}
	}
}

func serveNotificationPersistencePost(mux *http.ServeMux, path string, token string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func assertNotificationDegradedResponse(t *testing.T, rec *httptest.ResponseRecorder, operation string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("%s must preserve committed business result, got %d: %s", operation, rec.Code, rec.Body.String())
	}
	if rec.Header().Get(notificationPersistenceHeader) != "degraded" {
		t.Fatalf("%s must expose notification persistence degradation, headers=%v", operation, rec.Header())
	}
}

type alwaysFailNotificationRepository struct {
	err error
}

var _ notifications.Repository = alwaysFailNotificationRepository{}

func (r alwaysFailNotificationRepository) SaveNotification(context.Context, notifications.Notification, *notifications.WechatTask) (notifications.Notification, error) {
	return notifications.Notification{}, r.err
}

func (r alwaysFailNotificationRepository) ListNotifications(context.Context, int64) ([]notifications.Notification, error) {
	return nil, r.err
}

func (r alwaysFailNotificationRepository) FindNotification(context.Context, int64) (notifications.Notification, bool, error) {
	return notifications.Notification{}, false, r.err
}

func (r alwaysFailNotificationRepository) UpdateNotification(context.Context, notifications.Notification) (notifications.Notification, error) {
	return notifications.Notification{}, r.err
}

func (r alwaysFailNotificationRepository) ListWechatTasks(context.Context) ([]notifications.WechatTask, error) {
	return nil, r.err
}

func (r alwaysFailNotificationRepository) ListWechatTemplates(context.Context) ([]notifications.WechatTemplate, error) {
	return nil, r.err
}

func (r alwaysFailNotificationRepository) MarkWechatTaskSent(context.Context, int64, string, string, string) (notifications.WechatTask, error) {
	return notifications.WechatTask{}, r.err
}
