package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWechatSubscribeSenderHTTP(t *testing.T) {
	var gotTokenPath bool
	var gotSendPayload struct {
		ToUser     string                       `json:"touser"`
		TemplateID string                       `json:"template_id"`
		Page       string                       `json:"page"`
		Data       map[string]map[string]string `json:"data"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			gotTokenPath = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access-1","expires_in":7200}`))
			return
		}
		if r.URL.Path == "/cgi-bin/message/subscribe/send" {
			if err := json.NewDecoder(r.Body).Decode(&gotSendPayload); err != nil {
				t.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	sender := NewWechatSubscribeSender("wx-appid", "wx-secret")
	sender.TokenEndpoint = server.URL + "/cgi-bin/token"
	sender.SendEndpoint = server.URL + "/cgi-bin/message/subscribe/send"
	result, err := sender.Send(context.Background(), WechatSubscribeSendRequest{
		UserID:     1,
		OpenID:     "openid-1",
		TemplateID: "tpl-1",
		Page:       "pages/index/index",
		Data: map[string]string{
			"thing1": "hello",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !gotTokenPath || result.Code != "0" || result.Message != "ok" {
		t.Fatalf("unexpected send result: gotToken=%v result=%+v payload=%+v", gotTokenPath, result, gotSendPayload)
	}
	if gotSendPayload.ToUser != "openid-1" || gotSendPayload.TemplateID != "tpl-1" || gotSendPayload.Page != "pages/index/index" || gotSendPayload.Data["thing1"]["value"] != "hello" {
		t.Fatalf("unexpected payload: %+v", gotSendPayload)
	}
}

func TestWechatSubscribeSenderHTTPRejectsBadToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	sender := NewWechatSubscribeSender("wx-appid", "wx-secret")
	sender.TokenEndpoint = server.URL + "/cgi-bin/token"
	_, err := sender.Send(context.Background(), WechatSubscribeSendRequest{OpenID: "openid-1", TemplateID: "tpl-1"})
	if err != ErrWechatSubscribeSend {
		t.Fatalf("expected ErrWechatSubscribeSend, got %v", err)
	}
}
