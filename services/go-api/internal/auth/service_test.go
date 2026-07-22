package auth

import (
	"context"
	"errors"
	"testing"

	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

type failingPhoneSMSSender struct{}

func (failingPhoneSMSSender) GenerateCode() string {
	return "123456"
}

func (failingPhoneSMSSender) Send(context.Context, identity.SMSDispatchRequest) (identity.SMSDispatchResult, error) {
	return identity.SMSDispatchResult{}, errors.New("provider unavailable")
}

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

func TestPhoneLoginRegistersThenLogsInWithTemporaryCode(t *testing.T) {
	service := newTestService()
	invite, err := service.IssueInviteEntry(0, "qrcode")
	if err != nil {
		t.Fatalf("issue invite: %v", err)
	}

	registered, err := service.PhoneLogin(PhoneLoginRequest{
		Phone:      "13800138000",
		Code:       "000000",
		InviteCode: invite.Code,
		EntryType:  "qrcode",
	})
	if err != nil {
		t.Fatalf("phone register failed: %v", err)
	}
	if registered.Token == "" || registered.AuthPageMode != invites.AuthPageModeRegister {
		t.Fatalf("expected registered phone session: %+v", registered)
	}
	if registered.User.PhoneMasked != "138****8000" || registered.User.OpenID != "" {
		t.Fatalf("expected phone-only user: %+v", registered.User)
	}
	if registered.InviteRelation == nil {
		t.Fatal("expected invite relation")
	}

	loggedIn, err := service.PhoneLogin(PhoneLoginRequest{
		Phone:      "13800138000",
		Code:       "000000",
		InviteCode: invite.Code,
		EntryType:  "qrcode",
	})
	if err != nil {
		t.Fatalf("phone login failed: %v", err)
	}
	if loggedIn.User.ID != registered.User.ID || loggedIn.AuthPageMode != invites.AuthPageModeLogin {
		t.Fatalf("expected existing phone login: %+v", loggedIn)
	}
}

func TestPhoneLoginValidatesCodeAndInvite(t *testing.T) {
	service := newTestService()
	if _, err := service.PhoneLogin(PhoneLoginRequest{Phone: "13800138000", Code: "123456", InviteCode: "TEST2026"}); err != ErrPhoneCodeInvalid {
		t.Fatalf("expected ErrPhoneCodeInvalid, got %v", err)
	}
	if _, err := service.PhoneLogin(PhoneLoginRequest{Phone: "13800138000", Code: "000000"}); err != ErrInviteRequired {
		t.Fatalf("expected ErrInviteRequired, got %v", err)
	}
}

func TestSendPhoneCodeDoesNotThrottleAfterProviderFailure(t *testing.T) {
	service := newTestService()
	service.UsePhoneSMSSender(failingPhoneSMSSender{})

	if _, err := service.SendPhoneCode(context.Background(), "13800138000", "login"); !errors.Is(err, ErrPhoneCodeSendFailed) {
		t.Fatalf("first failed send error = %v, want ErrPhoneCodeSendFailed", err)
	}
	if _, err := service.SendPhoneCode(context.Background(), "13800138000", "login"); !errors.Is(err, ErrPhoneCodeSendFailed) {
		t.Fatalf("retry after failed send error = %v, want ErrPhoneCodeSendFailed instead of rate limit", err)
	}
}

func TestBoundPhoneCanLoginExistingWechatUserWithoutInvite(t *testing.T) {
	service := newTestService()
	wechat, err := service.WechatLogin(WechatLoginRequest{Code: "phone-bind-user", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("wechat login failed: %v", err)
	}
	if _, err := service.BindPhoneAuth(wechat.User.ID, "13900139000"); err != nil {
		t.Fatalf("bind phone auth failed: %v", err)
	}

	phone, err := service.PhoneLogin(PhoneLoginRequest{Phone: "13900139000", Code: "000000"})
	if err != nil {
		t.Fatalf("phone login failed: %v", err)
	}
	if phone.User.ID != wechat.User.ID || phone.AuthPageMode != invites.AuthPageModeLogin {
		t.Fatalf("expected existing wechat user phone login: %+v", phone)
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

func TestWechatLoginKeepsExistingInviteWhenBoundWechatScansAnotherInvite(t *testing.T) {
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
		t.Fatalf("existing user login with another invite failed: %v", err)
	}
	if second.User.ID != first.User.ID || second.AuthPageMode != invites.AuthPageModeLogin || !second.BoundWechat || second.InviteRelation == nil {
		t.Fatalf("expected existing wechat login mode with original relation, got %+v", second)
	}
	if second.InviteRelation.InviteCodeID != first.InviteRelation.InviteCodeID || second.InviteBindingStatus != InviteBindingStatusAlreadyBound || second.InviteBindingMessage == "" {
		t.Fatalf("expected original invite relation with already-bound notice, got %+v", second)
	}
	after, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: linkInvite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("precheck after existing login failed: %v", err)
	}
	if after.BoundWechat || after.AuthPageMode != invites.AuthPageModeRegister {
		t.Fatalf("expected new link invite to remain unbound, got %+v", after)
	}
	third, err := service.WechatLogin(WechatLoginRequest{Code: "u2", InviteCode: linkInvite.Code, EntryType: "link"})
	if err != nil {
		t.Fatalf("expected another wechat to bind still-unbound link invite, got %v", err)
	}
	if third.User.ID == first.User.ID || third.InviteRelation == nil || third.InviteRelation.InviteCodeID != linkInvite.ID {
		t.Fatalf("expected another wechat to bind link invite, got %+v", third)
	}
}

func TestWechatLoginAllowsBoundWechatWithoutInvite(t *testing.T) {
	service := newTestService()

	first, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	second, err := service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: ""})
	if err != nil {
		t.Fatalf("bound wechat should login without invite: %v", err)
	}
	if first.User.ID != second.User.ID || second.AuthPageMode != invites.AuthPageModeLogin || !second.BoundWechat || second.InviteRelation == nil {
		t.Fatalf("expected bound wechat login mode, got %+v", second)
	}
	second, err = service.WechatLogin(WechatLoginRequest{Code: "u1", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("second login with invite failed: %v", err)
	}
	if first.User.ID != second.User.ID {
		t.Fatalf("expected same user id, got %d and %d", first.User.ID, second.User.ID)
	}
}

func TestDeleteAccountReleasesLoginBindingsButExpiresOldInvite(t *testing.T) {
	service := newTestService()
	oldInvite, err := service.IssueInviteEntry(0, invites.EntryTypeQRCode)
	if err != nil {
		t.Fatalf("issue old invite: %v", err)
	}
	first, err := service.WechatLogin(WechatLoginRequest{Code: "deleted-account", InviteCode: oldInvite.Code, EntryType: invites.EntryTypeQRCode})
	if err != nil {
		t.Fatalf("first wechat login: %v", err)
	}
	if _, err := service.BindPhoneAuth(first.User.ID, "13900139001"); err != nil {
		t.Fatalf("bind phone: %v", err)
	}
	if _, ok := service.CurrentUser(first.PreAuthToken); !ok {
		t.Fatal("expected first session to be valid before deletion")
	}

	if err := service.DeleteAccount(first.User.ID); err != nil {
		t.Fatalf("delete account: %v", err)
	}
	if _, ok := service.CurrentUser(first.PreAuthToken); ok {
		t.Fatal("deleted account session must be revoked")
	}
	entry, err := service.WechatEntryPrecheck("deleted-account")
	if err != nil {
		t.Fatalf("wechat entry precheck: %v", err)
	}
	if entry.BoundWechat || !entry.RequiresInvite {
		t.Fatalf("deleted account must no longer retain wechat binding: %+v", entry)
	}
	if _, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: oldInvite.Code, EntryType: invites.EntryTypeQRCode}); err != ErrInviteExpired {
		t.Fatalf("old invite must be expired after account deletion, got %v", err)
	}

	newInvite, err := service.IssueInviteEntry(0, invites.EntryTypeQRCode)
	if err != nil {
		t.Fatalf("issue new invite: %v", err)
	}
	second, err := service.WechatLogin(WechatLoginRequest{Code: "deleted-account", InviteCode: newInvite.Code, EntryType: invites.EntryTypeQRCode})
	if err != nil {
		t.Fatalf("same wechat should register again with a new invite: %v", err)
	}
	if second.User.ID == first.User.ID {
		t.Fatalf("expected a new user after account deletion, got reused id %d", second.User.ID)
	}

	phoneInvite, err := service.IssueInviteEntry(0, invites.EntryTypeQRCode)
	if err != nil {
		t.Fatalf("issue phone invite: %v", err)
	}
	phoneLogin, err := service.PhoneLogin(PhoneLoginRequest{Phone: "13900139001", Code: temporaryPhoneCode, InviteCode: phoneInvite.Code, EntryType: invites.EntryTypeQRCode})
	if err != nil {
		t.Fatalf("same phone should register again with a new invite: %v", err)
	}
	if phoneLogin.User.ID == first.User.ID {
		t.Fatalf("expected a new phone user after account deletion, got reused id %d", phoneLogin.User.ID)
	}
}

func TestIssueAssignedInviteEntryOnlyUsesAllocatedUnusedCode(t *testing.T) {
	service := newTestService()
	allocated, err := service.IssueInviteEntry(77, "qrcode")
	if err != nil {
		t.Fatalf("allocate invite code: %v", err)
	}
	entry, err := service.IssueAssignedInviteEntry(77, "qrcode")
	if err != nil || entry.Code != allocated.Code {
		t.Fatalf("expected allocated code, entry=%+v err=%v", entry, err)
	}
	if _, err := service.WechatLogin(WechatLoginRequest{Code: "assigned-owner", InviteCode: allocated.Code, EntryType: "qrcode"}); err != nil {
		t.Fatalf("bind allocated code: %v", err)
	}
	if _, err := service.IssueAssignedInviteEntry(77, "qrcode"); !errors.Is(err, ErrInviteQuotaExceeded) {
		t.Fatalf("expected exhausted allocation, got %v", err)
	}
}

func TestWechatLoginAllowsExistingWechatWithoutInviteRelation(t *testing.T) {
	service := newTestService()
	user, err := service.users.Create(mockOpenID("existing-no-relation"))
	if err != nil {
		t.Fatalf("create existing user: %v", err)
	}

	login, err := service.WechatLogin(WechatLoginRequest{Code: "existing-no-relation"})
	if err != nil {
		t.Fatalf("existing WeChat user should login without invite relation: %v", err)
	}
	if login.User.ID != user.ID || !login.BoundWechat || login.InviteRelation != nil || login.AuthPageMode != invites.AuthPageModeLogin {
		t.Fatalf("unexpected existing WeChat login response: %+v", login)
	}
}
