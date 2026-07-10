package identity

import (
	"context"
	"testing"
)

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
