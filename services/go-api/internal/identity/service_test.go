package identity

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fixedSMSSender struct{}

func (fixedSMSSender) GenerateCode() string { return "123456" }

func (fixedSMSSender) Send(_ context.Context, _ SMSDispatchRequest) (SMSDispatchResult, error) {
	return SMSDispatchResult{Provider: "test"}, nil
}

type failingSMSSender struct{}

func (failingSMSSender) GenerateCode() string { return "123456" }

func (failingSMSSender) Send(_ context.Context, _ SMSDispatchRequest) (SMSDispatchResult, error) {
	return SMSDispatchResult{}, errors.New("provider unavailable")
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
	initial := service.Ensure(userID)
	if initial.CreatedAt == "" || initial.UpdatedAt != initial.CreatedAt {
		t.Fatalf("new identity record must have a stable creation timestamp, got %+v", initial)
	}

	record, err := service.BindPhone(userID, "13800138000")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != StatusPhoneBound {
		t.Fatalf("expected phone_bound, got %s", record.Status)
	}
	if record.CreatedAt != initial.CreatedAt {
		t.Fatalf("identity updates must preserve createdAt: before=%q after=%q", initial.CreatedAt, record.CreatedAt)
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
	repeated, err := service.CompleteFaceID(userID, token)
	if err != nil || repeated.Status != StatusVerified {
		t.Fatalf("provider retry should be idempotent, record=%+v err=%v", repeated, err)
	}
}

func TestStrictIdentityListDoesNotReturnCacheOnRepositoryFailure(t *testing.T) {
	repo := &recordingRepository{records: map[int64]Record{1: {UserID: 1, Status: StatusPendingManualReview}}, listRecordsErr: errors.New("identity database unavailable")}
	service := NewServiceWithRepository(repo)
	service.records[1] = Record{UserID: 1, Status: StatusVerified}

	items, err := service.AllRecordsStrict()
	if err == nil || items != nil {
		t.Fatalf("expected strict identity list error without cache fallback, items=%+v err=%v", items, err)
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

func TestIdentitySMSFailureDoesNotTriggerRetryThrottle(t *testing.T) {
	service := NewService()
	service.UseSMSSender(failingSMSSender{})
	if _, err := service.BindPhone(12, "13800138012"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(12); err == nil {
		t.Fatal("expected provider failure")
	}
	service.UseSMSSender(fixedSMSSender{})
	if _, err := service.SendSMSCode(12); err != nil {
		t.Fatalf("retry after provider failure must not be rate limited: %v", err)
	}
}

func TestSMSCodePersistenceFailureDoesNotTriggerRetryThrottle(t *testing.T) {
	base := newRecordingRepository()
	repo := &failingSMSCodeSaveRepository{recordingRepository: base, failSave: true}
	service := NewServiceWithRepository(repo)
	service.UseSMSSender(fixedSMSSender{})
	userID := int64(13013)
	if _, err := service.BindPhone(userID, "13800138064"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err == nil {
		t.Fatal("expected sms code persistence failure")
	}

	repo.failSave = false
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatalf("retry after sms code persistence failure must not be rate limited: %v", err)
	}
	if _, err := service.VerifySMSCode(userID, "123456"); err != nil {
		t.Fatalf("persisted retry code must remain verifiable: %v", err)
	}

	base.mu.Lock()
	storedCount := len(base.smsCodes)
	base.mu.Unlock()
	if storedCount != 1 {
		t.Fatalf("repository stored %d codes, want only the successful retry", storedCount)
	}
}

func TestIdentitySMSCodeExpiresAndCannotBeReplayed(t *testing.T) {
	service := NewService()
	service.UseSMSSender(fixedSMSSender{})
	if _, err := service.BindPhone(13, "13800138013"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(13); err != nil {
		t.Fatal(err)
	}
	service.mu.Lock()
	state := service.smsCodes[13]
	state.sentAt = time.Now().Add(-6 * time.Minute)
	service.smsCodes[13] = state
	service.mu.Unlock()
	if _, err := service.VerifySMSCode(13, "123456"); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("expired SMS code error = %v, want ErrCodeInvalid", err)
	}
	if _, err := service.SendSMSCode(13); err != nil {
		t.Fatalf("send replacement SMS code: %v", err)
	}
	if _, err := service.VerifySMSCode(13, "123456"); err != nil {
		t.Fatalf("verify replacement SMS code: %v", err)
	}
	if _, err := service.VerifySMSCode(13, "123456"); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("replayed SMS code error = %v, want ErrCodeInvalid", err)
	}
}

func TestRepositorySMSCodeCanBeVerifiedAcrossInstances(t *testing.T) {
	repo := newRecordingRepository()
	sender := NewServiceWithRepository(repo)
	sender.UseSMSSender(fixedSMSSender{})
	userID := int64(13010)
	if _, err := sender.BindPhone(userID, "13800138061"); err != nil {
		t.Fatal(err)
	}
	if _, err := sender.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}

	verifier := NewServiceWithRepository(repo)
	verifier.UseSMSSender(fixedSMSSender{})
	record, err := verifier.VerifySMSCode(userID, "123456")
	if err != nil {
		t.Fatalf("second instance must verify repository-backed code: %v", err)
	}
	if !record.SMSVerified || record.Status != StatusSMSVerified {
		t.Fatalf("repository verification must update identity atomically: %+v", record)
	}
	if _, err := sender.VerifySMSCode(userID, "123456"); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("consumed code must not replay on the sending instance: %v", err)
	}

	repo.mu.Lock()
	stored := repo.smsCodes[len(repo.smsCodes)-1]
	repo.mu.Unlock()
	if stored.VerifiedAt == nil || stored.CodeHash == "123456" {
		t.Fatalf("repository must mark the hashed code consumed without plaintext: %+v", stored)
	}
}

func TestRepositorySMSCodeHasSingleConcurrentConsumer(t *testing.T) {
	repo := newRecordingRepository()
	sender := NewServiceWithRepository(repo)
	sender.UseSMSSender(fixedSMSSender{})
	userID := int64(13012)
	if _, err := sender.BindPhone(userID, "13800138063"); err != nil {
		t.Fatal(err)
	}
	if _, err := sender.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}

	firstVerifier := NewServiceWithRepository(repo)
	secondVerifier := NewServiceWithRepository(repo)
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, verifier := range []*Service{firstVerifier, secondVerifier} {
		go func(service *Service) {
			<-start
			_, err := service.VerifySMSCode(userID, "123456")
			results <- err
		}(verifier)
	}
	close(start)

	successes := 0
	replays := 0
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrCodeInvalid):
			replays++
		default:
			t.Fatalf("unexpected concurrent verification error: %v", err)
		}
	}
	if successes != 1 || replays != 1 {
		t.Fatalf("concurrent consumers: successes=%d replays=%d, want one each", successes, replays)
	}
}

func TestRepositorySMSFailuresAccumulateAcrossInstances(t *testing.T) {
	repo := newRecordingRepository()
	sender := NewServiceWithRepository(repo)
	sender.UseSMSSender(fixedSMSSender{})
	userID := int64(13011)
	if _, err := sender.BindPhone(userID, "13800138062"); err != nil {
		t.Fatal(err)
	}
	if _, err := sender.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	firstVerifier := NewServiceWithRepository(repo)
	secondVerifier := NewServiceWithRepository(repo)
	for attempt := 1; attempt <= smsMaxFailures; attempt++ {
		service := firstVerifier
		if attempt%2 == 0 {
			service = secondVerifier
		}
		if _, err := service.VerifySMSCode(userID, "111111"); !errors.Is(err, ErrCodeInvalid) {
			t.Fatalf("wrong attempt %d error = %v", attempt, err)
		}
	}
	if _, err := sender.VerifySMSCode(userID, "123456"); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("correct code must be locked after five shared failures: %v", err)
	}
	repo.mu.Lock()
	failedCount := repo.smsCodes[len(repo.smsCodes)-1].VerifyFailedCount
	repo.mu.Unlock()
	if failedCount != smsMaxFailures {
		t.Fatalf("repository failed count=%d want %d", failedCount, smsMaxFailures)
	}
}

func TestPhoneLoginSyncPreservesVerifiedIdentityAfterColdStart(t *testing.T) {
	repo := newRecordingRepository()
	service := NewServiceWithRepository(repo)
	userID := int64(13001)
	if _, err := service.BindPhone(userID, "13800138031"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, temporarySMSCode); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitManualRealname(userID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	approved, err := service.ReviewManualRealname(userID, true, "")
	if err != nil {
		t.Fatal(err)
	}

	restarted := NewServiceWithRepository(repo)
	synced, err := restarted.SyncPhoneLoginVerification(userID, "13800138031")
	if err != nil {
		t.Fatal(err)
	}
	if synced.Status != StatusVerified || !synced.PhoneVerified || synced.RealNameCiphertext != approved.RealNameCiphertext || synced.IDCardCiphertext != approved.IDCardCiphertext {
		t.Fatalf("same-phone login must preserve approved identity after cold start: %+v", synced)
	}
}

func TestBindingDifferentPhoneClearsOldVerificationFlags(t *testing.T) {
	service := NewService()
	userID := int64(13002)
	if _, err := service.BindPhone(userID, "13800138032"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, temporarySMSCode); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitManualRealname(userID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReviewManualRealname(userID, true, ""); err != nil {
		t.Fatal(err)
	}
	changed, err := service.BindPhone(userID, "13800138033")
	if err != nil {
		t.Fatal(err)
	}
	if changed.Status != StatusPhoneBound || changed.SMSVerified || changed.PhoneVerified || changed.FaceVerified {
		t.Fatalf("changing phone must clear old verification flags: %+v", changed)
	}
}

func TestConcurrentManualReviewHasSingleWinner(t *testing.T) {
	repo := newRecordingRepository()
	service := NewServiceWithRepository(repo)
	userID := int64(13003)
	if _, err := service.BindPhone(userID, "13800138034"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, temporarySMSCode); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitManualRealname(userID, "Test User", "110101199001011234"); err != nil {
		t.Fatal(err)
	}

	approveService := NewServiceWithRepository(repo)
	rejectService := NewServiceWithRepository(repo)
	errorsCh := make(chan error, 2)
	start := make(chan struct{})
	go func() {
		<-start
		_, err := approveService.ReviewManualRealname(userID, true, "")
		errorsCh <- err
	}()
	go func() {
		<-start
		_, err := rejectService.ReviewManualRealname(userID, false, "资料不一致")
		errorsCh <- err
	}()
	close(start)
	firstErr := <-errorsCh
	secondErr := <-errorsCh
	successes := 0
	conflicts := 0
	for _, err := range []error{firstErr, secondErr} {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrReviewStateInvalid) {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent review error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("expected one review winner and one conflict, success=%d conflict=%d", successes, conflicts)
	}
}

func TestSMSVerificationPersistenceFailureDoesNotConsumeCode(t *testing.T) {
	base := newRecordingRepository()
	repo := &failingRecordRepository{recordingRepository: base}
	service := NewServiceWithRepository(repo)
	service.UseSMSSender(fixedSMSSender{})
	userID := int64(13004)
	if _, err := service.BindPhone(userID, "13800138035"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	repo.failSave = true
	if _, err := service.VerifySMSCode(userID, "123456"); err == nil {
		t.Fatal("expected record persistence failure")
	}
	repo.failSave = false
	if _, err := service.VerifySMSCode(userID, "123456"); err != nil {
		t.Fatalf("verification code must remain usable after persistence failure: %v", err)
	}
}

func TestManualRealnamePersistenceFailureDoesNotPolluteCachedStatus(t *testing.T) {
	base := newRecordingRepository()
	repo := &failingRecordRepository{recordingRepository: base}
	service := NewServiceWithRepository(repo)
	service.UseSMSSender(fixedSMSSender{})
	userID := int64(13005)
	if _, err := service.BindPhone(userID, "13800138036"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendSMSCode(userID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, "123456"); err != nil {
		t.Fatal(err)
	}
	repo.failSave = true
	if _, err := service.SubmitManualRealname(userID, "测试用户", "110101199001011234"); err == nil {
		t.Fatal("expected manual realname persistence failure")
	}
	repo.failSave = false
	record, err := service.StatusStrict(userID)
	if err != nil {
		t.Fatal(err)
	}
	if record.Status == StatusPendingManualReview || record.RealNameMasked != "" || record.IDCardMasked != "" {
		t.Fatalf("failed persistence must not change cached identity: %+v", record)
	}
}

func TestStatusStrictReturnsRepositoryFailure(t *testing.T) {
	repo := &failingRecordRepository{recordingRepository: newRecordingRepository(), failSave: true}
	service := NewServiceWithRepository(repo)
	if _, err := service.StatusStrict(13006); err == nil {
		t.Fatal("expected identity repository read failure")
	}
}

func TestManualRealnameReviewFlow(t *testing.T) {
	service := NewService()
	userID := int64(10001)

	if _, err := service.SubmitManualRealname(userID, "Test User", "bad-id"); err != ErrIDCardInvalid {
		t.Fatalf("expected invalid id card, got %v", err)
	}
	if _, err := service.SubmitManualRealname(userID, "Test User", "110101199001011234"); err != ErrCodeInvalid {
		t.Fatalf("expected manual realname to require a verified phone, got %v", err)
	}
	if _, err := service.BindPhone(userID, "13800138001"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifySMSCode(userID, temporarySMSCode); err != nil {
		t.Fatal(err)
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
	if _, err := service.ReviewManualRealname(userID, true, ""); !errors.Is(err, ErrReviewStateInvalid) {
		t.Fatalf("completed identity review must not be repeated, got %v", err)
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
	if _, err := service.BindPhone(userID, "13800138000"); err != nil {
		t.Fatal(err)
	}

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
	mu             sync.Mutex
	records        map[int64]Record
	smsCodes       []SMSCodeRecord
	faceSessions   []FaceIDSessionRecord
	listRecordsErr error
}

type failingRecordRepository struct {
	*recordingRepository
	failSave bool
}

type failingSMSCodeSaveRepository struct {
	*recordingRepository
	failSave bool
}

func (r *failingRecordRepository) SaveRecord(ctx context.Context, record Record) error {
	if r.failSave {
		return errors.New("record persistence unavailable")
	}
	return r.recordingRepository.SaveRecord(ctx, record)
}

func (r *failingRecordRepository) VerifyAndConsumeSMSCode(ctx context.Context, userID int64, scene string, codeHash string, verifiedAt time.Time, maxFailures int) (Record, bool, error) {
	if r.failSave {
		return Record{}, false, errors.New("record persistence unavailable")
	}
	return r.recordingRepository.VerifyAndConsumeSMSCode(ctx, userID, scene, codeHash, verifiedAt, maxFailures)
}

func (r *failingSMSCodeSaveRepository) SaveSMSCode(ctx context.Context, record SMSCodeRecord) error {
	if r.failSave {
		return errors.New("sms code persistence unavailable")
	}
	return r.recordingRepository.SaveSMSCode(ctx, record)
}

func newRecordingRepository() *recordingRepository {
	return &recordingRepository{records: make(map[int64]Record)}
}

func (r *recordingRepository) SaveRecord(_ context.Context, record Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[record.UserID] = record
	return nil
}

func (r *recordingRepository) UpdateRecordIfStatus(_ context.Context, record Record, expected Status) (Record, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.records[record.UserID]
	if !ok || current.Status != expected {
		return Record{}, false, nil
	}
	current.Status = record.Status
	current.PhoneVerified = record.PhoneVerified
	current.FailureReason = record.FailureReason
	current.UpdatedAt = record.UpdatedAt
	r.records[record.UserID] = current
	return current, true, nil
}

func (r *recordingRepository) FindRecord(_ context.Context, userID int64) (Record, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[userID]
	return record, ok, nil
}

func (r *recordingRepository) ListRecords(_ context.Context) ([]Record, error) {
	if r.listRecordsErr != nil {
		return nil, r.listRecordsErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]Record, 0, len(r.records))
	for _, record := range r.records {
		items = append(items, record)
	}
	return items, nil
}

func (r *recordingRepository) SaveSMSCode(_ context.Context, record SMSCodeRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.smsCodes = append(r.smsCodes, record)
	return nil
}

func (r *recordingRepository) VerifyAndConsumeSMSCode(_ context.Context, userID int64, scene string, codeHash string, verifiedAt time.Time, maxFailures int) (Record, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := -1
	for candidate := len(r.smsCodes) - 1; candidate >= 0; candidate-- {
		if r.smsCodes[candidate].UserID == userID && r.smsCodes[candidate].Scene == scene {
			index = candidate
			break
		}
	}
	if index < 0 {
		return Record{}, false, nil
	}
	codeRecord := r.smsCodes[index]
	decision := evaluateSMSCodeHash(codeRecord.CodeHash, codeHash, codeRecord.ExpiresAt, codeRecord.VerifiedAt != nil, codeRecord.VerifyFailedCount, maxFailures, verifiedAt)
	if decision == smsCodeInvalid {
		codeRecord.VerifyFailedCount++
		if codeRecord.VerifyFailedCount > maxFailures {
			codeRecord.VerifyFailedCount = maxFailures
		}
		r.smsCodes[index] = codeRecord
		return Record{}, false, nil
	}
	if decision != smsCodeValid {
		return Record{}, false, nil
	}
	record, ok := r.records[userID]
	if !ok {
		return Record{}, false, ErrRecordNotFound
	}
	record.SMSVerified = true
	if record.Status == StatusWechatLoggedIn || record.Status == StatusPhoneBound || record.Status == "" {
		record.Status = StatusSMSVerified
	}
	record.UpdatedAt = verifiedAt.Format(time.RFC3339)
	consumedAt := verifiedAt
	codeRecord.VerifiedAt = &consumedAt
	r.smsCodes[index] = codeRecord
	r.records[userID] = record
	return record, true, nil
}

func (r *recordingRepository) SaveFaceIDSession(_ context.Context, record FaceIDSessionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.faceSessions = append(r.faceSessions, record)
	return nil
}
