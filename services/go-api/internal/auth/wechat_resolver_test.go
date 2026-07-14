package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWechatAPIResolverResolvesOpenID(t *testing.T) {
	var gotAppID string
	var gotSecret string
	var gotCode string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAppID = r.URL.Query().Get("appid")
		gotSecret = r.URL.Query().Get("secret")
		gotCode = r.URL.Query().Get("js_code")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"openid":"openid-123","unionid":"union-123","session_key":"session-123"}`))
	}))
	defer server.Close()

	resolver := NewWechatAPIResolver("wx-app", "wx-secret")
	resolver.Endpoint = server.URL

	session, err := resolver.Resolve(context.Background(), "login-code")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if session.OpenID != "openid-123" || session.UnionID != "union-123" || session.SessionKey != "session-123" {
		t.Fatalf("unexpected session: %+v", session)
	}
	if gotAppID != "wx-app" || gotSecret != "wx-secret" || gotCode != "login-code" {
		t.Fatalf("unexpected query: appid=%s secret=%s code=%s", gotAppID, gotSecret, gotCode)
	}
}

func TestWechatAPIResolverMapsWechatError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer server.Close()

	resolver := NewWechatAPIResolver("wx-app", "wx-secret")
	resolver.Endpoint = server.URL

	if _, err := resolver.Resolve(context.Background(), "bad-code"); err != ErrWechatCodeInvalid {
		t.Fatalf("expected ErrWechatCodeInvalid, got %v", err)
	}
}

func TestWechatLoginUsesInjectedResolver(t *testing.T) {
	service := newTestService()
	service.UseWechatCodeResolver(staticWechatResolver{openid: "real-openid"})
	invite, err := service.IssueInviteEntry(0, "link")
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}

	resp, err := service.WechatLogin(WechatLoginRequest{Code: "real-code", InviteCode: invite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp.User.OpenID != "real-openid" {
		t.Fatalf("expected injected openid, got %+v", resp.User)
	}
}

type staticWechatResolver struct {
	openid string
}

func (r staticWechatResolver) Resolve(context.Context, string) (WechatSession, error) {
	return WechatSession{OpenID: r.openid}, nil
}
