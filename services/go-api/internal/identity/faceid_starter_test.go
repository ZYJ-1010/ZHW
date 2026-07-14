package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPFaceIDStarterStartsVerification(t *testing.T) {
	var gotSecret string
	var gotPayload struct {
		UserID         int64  `json:"userId"`
		PhoneMasked    string `json:"phoneMasked"`
		RealNameMasked string `json:"realNameMasked"`
		IDCardMasked   string `json:"idCardMasked"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSecret = r.Header.Get("X-FaceID-Secret")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"faceToken":"face-token-1","requestId":"req-1"}`))
	}))
	defer server.Close()

	starter := NewHTTPFaceIDStarter(server.URL, "faceid-secret")
	service := NewService()
	service.UseFaceIDStarter(starter)
	userID := int64(7)

	if _, err := service.BindPhone(userID, "13800138000"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, "000000"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifyPhone(userID, "User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	token, err := service.StartFaceID(userID)
	if err != nil {
		t.Fatal(err)
	}
	if token != "face-token-1" {
		t.Fatalf("expected gateway face token, got %s", token)
	}
	if gotSecret != "faceid-secret" || gotPayload.UserID != userID || gotPayload.PhoneMasked != "138****8000" || gotPayload.RealNameMasked != "U***" || gotPayload.IDCardMasked != "110***********1234" {
		t.Fatalf("unexpected faceid request secret=%s payload=%+v", gotSecret, gotPayload)
	}
	if _, err := service.CompleteFaceID(userID, token); err != nil {
		t.Fatalf("expected gateway token to complete faceid: %v", err)
	}
}

func TestHTTPFaceIDStarterAcceptsBizToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"bizToken":"biz-token-1"}`))
	}))
	defer server.Close()

	starter := NewHTTPFaceIDStarter(server.URL, "faceid-secret")
	result, err := starter.Start(context.Background(), FaceIDStartRequest{UserID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.FaceToken != "biz-token-1" || result.Provider != "http" {
		t.Fatalf("unexpected faceid result: %+v", result)
	}
}

func TestHTTPFaceIDStarterMapsFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusBadGateway)
	}))
	defer server.Close()

	starter := NewHTTPFaceIDStarter(server.URL, "faceid-secret")
	_, err := starter.Start(context.Background(), FaceIDStartRequest{UserID: 1})
	if err != ErrFaceIDStartFailed {
		t.Fatalf("expected ErrFaceIDStartFailed, got %v", err)
	}
}
