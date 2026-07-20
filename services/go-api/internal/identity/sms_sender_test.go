package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSMSSenderSendsCodeWithoutReturningMockCode(t *testing.T) {
	var gotSecret string
	var gotPayload struct {
		UserID      int64  `json:"userId"`
		Phone       string `json:"phone"`
		PhoneMasked string `json:"phoneMasked"`
		Scene       string `json:"scene"`
		Code        string `json:"code"`
		ExpiresAt   string `json:"expiresAt"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSecret = r.Header.Get("X-SMS-Secret")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messageId":"msg-1"}`))
	}))
	defer server.Close()

	sender := NewHTTPSMSSender(server.URL, "sms-secret")
	service := NewService()
	service.UseSMSSender(sender)

	if _, err := service.BindPhone(7, "13800138000"); err != nil {
		t.Fatal(err)
	}
	result, err := service.SendSMSCode(7)
	if err != nil {
		t.Fatal(err)
	}
	if result.Provider != "http" || result.MessageID != "msg-1" || result.MockCode != "" {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
	if gotSecret != "sms-secret" || gotPayload.UserID != 7 || len(gotPayload.Code) != 6 || gotPayload.PhoneMasked != "138****8000" || gotPayload.Phone != "" {
		t.Fatalf("unexpected sms request secret=%s payload=%+v", gotSecret, gotPayload)
	}
	if _, err := service.VerifySMSCode(7, gotPayload.Code); err != nil {
		t.Fatalf("expected generated code to verify: %v", err)
	}
}

func TestHTTPSMSSenderMapsNon2xxToSendFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusBadGateway)
	}))
	defer server.Close()

	sender := NewHTTPSMSSender(server.URL, "sms-secret")
	_, err := sender.Send(context.Background(), SMSDispatchRequest{UserID: 1, Code: "123456"})
	if err != ErrSMSSendFailed {
		t.Fatalf("expected ErrSMSSendFailed, got %v", err)
	}
}
