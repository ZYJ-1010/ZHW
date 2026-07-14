package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWriteFormatsRFC3339TimeToSeconds(t *testing.T) {
	recorder := httptest.NewRecorder()
	value := time.Date(2026, 7, 12, 9, 14, 0, 365326000, time.UTC)

	OK(recorder, map[string]interface{}{"createdAt": value})

	body := recorder.Body.String()
	if !strings.Contains(body, `"createdAt":"2026-07-12T09:14:00Z"`) {
		t.Fatalf("expected second precision time, got %s", body)
	}
	if strings.Contains(body, ".365326") {
		t.Fatalf("unexpected fractional seconds in %s", body)
	}
}
