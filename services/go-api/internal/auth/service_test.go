package auth

import (
	"testing"

	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func newTestService() *Service {
	return NewService(users.NewStore(), invites.NewStore(), NewTokenStore())
}

func TestWechatLoginRequiresInviteForNewUser(t *testing.T) {
	service := newTestService()

	_, err := service.WechatLogin(WechatLoginRequest{Code: "u1"})
	if err != ErrInviteRequired {
		t.Fatalf("expected ErrInviteRequired, got %v", err)
	}
}

func TestWechatLoginCreatesUserWithValidInvite(t *testing.T) {
	service := newTestService()
	invite, err := service.IssueInviteEntry(0, "qrcode")
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}

	resp, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: invite.Code, EntryType: "qrcode"})
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if resp.Token != "" || resp.PreAuthToken == "" {
		t.Fatal("expected pre-auth token only")
	}
	if !resp.RequiresIdentityBinding || resp.IdentityBindStatus != "wechat_logged_in" {
		t.Fatalf("expected identity binding hint, got %+v", resp)
	}
	if resp.User.ID == 0 {
		t.Fatal("expected user id")
	}
	if resp.InviteRelation == nil {
		t.Fatal("expected invite relation")
	}
	if resp.EntryType != "qrcode" || resp.AuthPageMode != "register" {
		t.Fatalf("expected qrcode register mode, got %+v", resp)
	}
}

func TestInvitePrecheckSplitsRegisterAndLogin(t *testing.T) {
	service := newTestService()
	invite, err := service.IssueInviteEntry(0, "poster")
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}

	before, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: invite.Code, EntryType: "poster"})
	if err != nil {
		t.Fatalf("precheck before bind failed: %v", err)
	}
	if !before.Valid || before.EntryType != "poster" || before.AuthPageMode != "register" || before.BoundWechat {
		t.Fatalf("expected poster register mode before bind, got %+v", before)
	}

	login, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: invite.Code, EntryType: "poster"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	after, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: invite.Code, EntryType: "poster"})
	if err != nil {
		t.Fatalf("precheck after bind failed: %v", err)
	}
	if !after.BoundWechat || after.AuthPageMode != "login" || after.BoundUserID != login.User.ID {
		t.Fatalf("expected login mode after bind, got %+v", after)
	}
}

func TestWechatLoginRejectsBoundInviteForDifferentWechat(t *testing.T) {
	service := newTestService()
	invite, err := service.IssueInviteEntry(0, "link")
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}

	if _, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: invite.Code, EntryType: "link"}); err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	relogin, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: invite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("same wechat should reuse its bound invite: %v", err)
	}
	if relogin.AuthPageMode != invites.AuthPageModeLogin || !relogin.BoundWechat {
		t.Fatalf("expected same wechat login mode, got %+v", relogin)
	}
	_, err = service.WechatLogin(WechatLoginRequest{Code: "u2", InviteCode: invite.Code, EntryType: "link"})
	if err != ErrInviteAlreadyBound {
		t.Fatalf("expected ErrInviteAlreadyBound, got %v", err)
	}
}

func TestWechatLoginBindsNewEntryCodeForExistingWechat(t *testing.T) {
	service := newTestService()
	firstInvite, err := service.IssueInviteEntry(0, "poster")
	if err != nil {
		t.Fatalf("issue first invite: %v", err)
	}
	first, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: firstInvite.Code, EntryType: "poster"})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}

	linkInvite, err := service.IssueInviteEntry(first.User.ID, "link")
	if err != nil {
		t.Fatalf("issue link invite: %v", err)
	}
	before, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: linkInvite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("precheck before existing login failed: %v", err)
	}
	if before.AuthPageMode != invites.AuthPageModeRegister || before.BoundWechat {
		t.Fatalf("expected fresh link invite before existing login, got %+v", before)
	}

	second, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: linkInvite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("existing user login with new link failed: %v", err)
	}
	if second.User.ID != first.User.ID || second.AuthPageMode != invites.AuthPageModeLogin || !second.BoundWechat {
		t.Fatalf("expected existing wechat login mode, got %+v", second)
	}
	after, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: linkInvite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("precheck after existing login failed: %v", err)
	}
	if !after.BoundWechat || after.BoundUserID != first.User.ID || after.AuthPageMode != invites.AuthPageModeLogin {
		t.Fatalf("expected link invite bound to existing user, got %+v", after)
	}
	if _, err := service.WechatLogin(WechatLoginRequest{Code: "u2", InviteCode: linkInvite.Code, EntryType: "link"}); err != ErrInviteAlreadyBound {
		t.Fatalf("expected bound link to reject another wechat, got %v", err)
	}
}

func TestWechatLoginRequiresInviteForExistingUserEntry(t *testing.T) {
	service := newTestService()

	first, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	second, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: ""})
	if err != ErrInviteRequired {
		t.Fatalf("expected second login to require invite, got %v", err)
	}
	second, err = service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("second login with invite failed: %v", err)
	}
	if first.User.ID != second.User.ID {
		t.Fatalf("expected same user id, got %d and %d", first.User.ID, second.User.ID)
	}
}
