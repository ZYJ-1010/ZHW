package identity

import "testing"

func TestSMSVerificationSatisfiesPhaseOneIdentity(t *testing.T) {
	service := NewService()
	userID := int64(7)

	record := service.Status(userID)
	if record.Status != StatusWechatLoggedIn {
		t.Fatalf("expected default wechat_logged_in, got %+v", record)
	}
	if service.IsVerified(userID) {
		t.Fatal("expected user not verified before sms")
	}

	if _, err := service.BindPhone(userID, "13800138000"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, "000000"); err != nil {
		t.Fatal(err)
	}
	if !service.IsVerified(userID) {
		t.Fatal("expected user verified for phase one after sms")
	}
	record, err := service.VerifyPhone(userID, "User", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if record.RealNameMasked != "U***" || record.IDCardMasked != "110***********1234" {
		t.Fatalf("expected masked identity fields, got %+v", record)
	}
	if !service.IsVerified(userID) {
		t.Fatal("expected user still verified before FaceID completion")
	}
	token, err := service.StartFaceID(userID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteFaceID(userID, token); err != nil {
		t.Fatal(err)
	}
	if !service.IsVerified(userID) {
		t.Fatal("expected user verified after FaceID completion")
	}
}

func TestStrongIdentityRejectsInvalidPhoneAndIDCard(t *testing.T) {
	service := NewService()
	userID := int64(9)

	if _, err := service.BindPhone(userID, "12345"); err != ErrPhoneInvalid {
		t.Fatalf("expected ErrPhoneInvalid, got %v", err)
	}
	if _, err := service.BindPhone(userID, "13800138000"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, "000000"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifyPhone(userID, "User", "bad-id"); err != ErrIDCardInvalid {
		t.Fatalf("expected ErrIDCardInvalid, got %v", err)
	}
	if _, err := service.StartFaceID(userID); err != ErrPhoneNotVerified {
		t.Fatalf("expected ErrPhoneNotVerified, got %v", err)
	}
}
