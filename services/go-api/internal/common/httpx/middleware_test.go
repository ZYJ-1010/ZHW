package httpx

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLogRedactsSensitiveRequestData(t *testing.T) {
	var output bytes.Buffer
	logger := log.New(&output, "", 0)
	handler := AccessLog(RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OK(w, map[string]string{"status": "ok"})
	})), logger)

	body := strings.NewReader(`{"phone":"13800138000","idCard":"110101199001011234","password":"secret-pass"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/app/identity/phone/verify?token=query-token&phone=13800138000", body)
	req.Header.Set("Authorization", "Bearer header-token")
	req.Header.Set("X-Request-Id", "rid-test")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	logText := output.String()
	for _, forbidden := range []string{
		"header-token",
		"query-token",
		"13800138000",
		"110101199001011234",
		"secret-pass",
		"Authorization",
	} {
		if strings.Contains(logText, forbidden) {
			t.Fatalf("access log leaked %q: %s", forbidden, logText)
		}
	}
	for _, expected := range []string{
		"method=POST",
		"path=/api/app/identity/phone/verify",
		"status=200",
		"request_id=rid-test",
	} {
		if !strings.Contains(logText, expected) {
			t.Fatalf("access log missing %q: %s", expected, logText)
		}
	}
}

func TestRequestIDIsIncludedInJSONResponse(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OK(w, map[string]string{"status": "ok"})
	}))
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, "rid-json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(RequestIDHeader) != "rid-json" {
		t.Fatalf("expected response request id header, got %q", rec.Header().Get(RequestIDHeader))
	}
	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.RequestID != "rid-json" {
		t.Fatalf("expected response body request id, got %q body=%s", resp.RequestID, rec.Body.String())
	}
}
