package identity

import (
	"context"
	"strings"
	"testing"
)

type fixedSMSSender struct{}

func (fixedSMSSender) GenerateCode() string { return "123456" }

func (fixedSMSSender) Send(_ context.Context, _ SMSDispatchRequest) (SMSDispatchResult, error) {
	return SMSDispatchResult{Provider: "test"}, nil
}

func TestRealNameInitialsUsesFirstTwoChineseCharacters(t *testing.T) {
	cases := map[string]string{"张三": "ZS", "王小明": "WX", "欧阳娜娜": "OY"}
	for name, want := range cases {
		if got := RealNameInitials(name); got != want {
			t.Fatalf("RealNameInitials(%q)=%q want %q", name, got, want)
		}
	}
}

func TestRealNameEncryptionRoundTrip(t *testing.T) {
	service := NewService()
	service.UseDataEncryptionKey("test-identity-key")
	ciphertext, err := service.encryptRealName("张三")
	if err != nil || ciphertext == "" || strings.Contains(ciphertext, "张三") {
		t.Fatalf("expected encrypted real name, ciphertext=%q err=%v", ciphertext, err)
	}
	plain, err := service.decryptRealName(ciphertext)
	if err != nil || plain != "张三" {
		t.Fatalf("decrypt real name=%q err=%v", plain, err)
	}
}

func TestInGameIdentityDoesNotExposeMaskedNameAsFullIdentity(t *testing.T) {
	repo := newRecordingRepository()
	repo.records[10005] = Record{
		UserID:         10005,
		Status:         StatusVerified,
		PhoneVerified:  true,
		FaceVerified:   true,
		RealNameMasked: "测****",
	}
	service := NewServiceWithRepository(repo)

	if profile, ok := service.InGameIdentity(10005); ok {
		t.Fatalf("masked-only legacy identity must fall back to the user's display name, got %+v", profile)
	}
}

func TestIdentityFlow(t *testing.T) {
	service := NewService()
	userID := int64(1)

	record, err := service.BindPhone(userID, "13800138000")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusPhoneBound {
		t.Fatalf("expected phone_bound, got %s", record.Status)
	}

	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	record, err = service.VerifySMSCode(userID, "000000")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusSMSVerified {
		t.Fatalf("expected sms_verified, got %s", record.Status)
	}

	record, err = service.VerifyPhone(userID, "张三", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusPhoneVerified {
		t.Fatalf("expected phone_verified, got %s", record.Status)
	}

	token, err := service.StartFaceID(userID)
	if err != nil {
		t.Fatal(err)
	}
	record, err = service.CompleteFaceID(userID, token)
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusVerified {
		t.Fatalf("expected verified, got %s", record.Status)
	}
}

func TestRealSMSSenderDoesNotAllowTemporaryBypassCode(t *testing.T) {
	service := NewService()
	service.UseSMSSender(fixedSMSSender{})
	if _, err := service.BindPhone(11, "13800138000"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(11); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(11, temporarySMSCode); err != ErrCodeInvalid {
		t.Fatalf("temporary code must be rejected for non-local sender, got %v", err)
	}
	if _, err := service.VerifySMSCode(11, "123456"); err != nil {
		t.Fatalf("actual provider code must verify: %v", err)
	}
}

func TestManualRealnameReviewFlow(t *testing.T) {
	service := NewService()
	userID := int64(10001)

	if _, err := service.SubmitManualRealname(userID, "Test User", "bad-id"); err != ErrIDCardInvalid {
		t.Fatalf("expected invalid id card, got %v", err)
	}

	record, err := service.SubmitManualRealname(userID, "Test User", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusPendingManualReview || record.PhoneVerified || record.RealNameMasked == "" || record.IDCardMasked == "" {
		t.Fatalf("expected pending manual realname record, got %+v", record)
	}
	if service.IsVerified(userID) {
		t.Fatal("pending manual realname must not be treated as verified")
	}
	if _, err := service.ReviewManualRealname(userID, false, " "); err != ErrReviewReasonRequired {
		t.Fatalf("expected rejected review to require a reason, got %v", err)
	}

	record, err = service.ReviewManualRealname(userID, false, "retry")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusRejected || record.FailureReason != "retry" || service.IsVerified(userID) {
		t.Fatalf("expected rejected manual realname record, got %+v", record)
	}

	record, err = service.SubmitManualRealname(userID, "Test User", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusPendingManualReview {
		t.Fatalf("expected pending after resubmit, got %+v", record)
	}

	record, err = service.ReviewManualRealname(userID, true, "")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusVerified || !record.PhoneVerified || !service.IsVerified(userID) {
		t.Fatalf("expected approved manual realname to verify identity, got %+v", record)
	}
	record, err = service.SubmitManualRealname(userID, "Test User", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusPendingManualReview || record.PhoneVerified || service.IsVerified(userID) {
		t.Fatalf("expected resubmitted realname to return to pending review, got %+v", record)
	}
}

func TestRealnameVerificationDoesNotTreatPhoneVerificationAsApprovedIdentity(t *testing.T) {
	service := NewService()
	userID := int64(10002)
	if _, err := service.BindPhone(userID, "13800000002"); err != nil {
		t.Fatal(err)
	}
	dispatch, err := service.SendSMSCode(userID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.VerifySMSCode(userID, dispatch.MockCode); err != nil {
		t.Fatal(err)
	}
	if _, err = service.VerifyPhone(userID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	if !service.IsVerified(userID) {
		t.Fatal("phone verification should remain available for account security")
	}
	if service.IsRealnameVerified(userID) {
		t.Fatal("phone verification must not satisfy the role realname requirement")
	}
	if _, err = service.SubmitManualRealname(userID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.ReviewManualRealname(userID, true, ""); err != nil {
		t.Fatal(err)
	}
	if !service.IsRealnameVerified(userID) {
		t.Fatal("approved personal identity should satisfy the role realname requirement")
	}
}

func TestSMSCodeRateLimit(t *testing.T) {
	service := NewService()
	userID := int64(1)

	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != ErrSMSRateLimited {
		t.Fatalf("expected ErrSMSRateLimited, got %v", err)
	}
}

func TestTemporarySMSCodeCanVerifyWithoutDispatch(t *testing.T) {
	service := NewService()
	userID := int64(1)

	if _, err := service.BindPhone(userID, "13800138000"); err != nil {
		t.Fatal(err)
	}
	record, err := service.VerifySMSCode(userID, "000000")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusSMSVerified {
		t.Fatalf("expected sms_verified, got %s", record.Status)
	}
}

func TestIdentityRepositoryPersistsStrongIdentityFlow(t *testing.T) {
	repo := newRecordingRepository()
	service := NewServiceWithRepository(repo)
	userID := int64(1)

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
	if _, err := service.CompleteFaceID(userID, token); err != nil {
		t.Fatal(err)
	}

	if repo.records[userID].Status != StatusVerified {
		t.Fatalf("expected persisted verified record, got %+v", repo.records[userID])
	}
	if repo.records[userID].RealNameMasked != "U***" || repo.records[userID].IDCardMasked != "110***********1234" {
		t.Fatalf("expected persisted masked identity fields, got %+v", repo.records[userID])
	}
	if len(repo.smsCodes) != 1 || repo.smsCodes[0].CodeHash == "" || repo.smsCodes[0].CodeHash == "000000" {
		t.Fatalf("expected hashed sms code record, got %+v", repo.smsCodes)
	}
	if len(repo.faceSessions) != 2 || repo.faceSessions[0].Status != string(StatusFaceIDProcessing) || repo.faceSessions[1].Status != string(StatusVerified) {
		t.Fatalf("expected faceid session lifecycle, got %+v", repo.faceSessions)
	}
}

type recordingRepository struct {
	records      map[int64]Record
	smsCodes     []SMSCodeRecord
	faceSessions []FaceIDSessionRecord
}

func newRecordingRepository() *recordingRepository {
	return &recordingRepository{records: make(map[int64]Record)}
}

func (r *recordingRepository) SaveRecord(_ context.Context, record Record) error {
	r.records[record.UserID] = record
	return nil
}

func (r *recordingRepository) FindRecord(_ context.Context, userID int64) (Record, bool, error) {
	record, ok := r.records[userID]
	return record, ok, nil
}

func (r *recordingRepository) ListRecords(_ context.Context) ([]Record, error) {
	items := make([]Record, 0, len(r.records))
	for _, record := range r.records {
		items = append(items, record)
	}
	return items, nil
}

func (r *recordingRepository) SaveSMSCode(_ context.Context, record SMSCodeRecord) error {
	r.smsCodes = append(r.smsCodes, record)
	return nil
}

func (r *recordingRepository) SaveFaceIDSession(_ context.Context, record FaceIDSessionRecord) error {
	r.faceSessions = append(r.faceSessions, record)
	return nil
}
