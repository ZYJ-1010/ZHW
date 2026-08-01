package audit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecordBehaviorKeepsPlanFieldsAndLegacyAliases(t *testing.T) {
	service := NewService()

	log := service.RecordBehavior(BehaviorRequest{
		UserID:       1,
		EventCode:    "view_game_detail",
		BusinessType: "game",
		BusinessID:   10,
		Source:       "app",
		Device:       "android",
		IP:           "127.0.0.1",
	})

	if log.EventCode != "view_game_detail" || log.EventType != "view_game_detail" {
		t.Fatalf("expected event code aliases, got %+v", log)
	}
	if log.BusinessType != "game" || log.TargetType != "game" || log.BusinessID != 10 || log.TargetID != 10 {
		t.Fatalf("expected business aliases, got %+v", log)
	}
	if log.Source != "app" || log.Device != "android" || log.IP == "" || log.OccurredAt.IsZero() {
		t.Fatalf("expected plan fields, got %+v", log)
	}
}

func TestQueryBehaviorLogsSupportsEventCode(t *testing.T) {
	service := NewService()
	service.RecordBehavior(BehaviorRequest{UserID: 1, EventCode: "search", BusinessType: "game", BusinessID: 1})
	service.RecordBehavior(BehaviorRequest{UserID: 1, EventCode: "share", BusinessType: "game", BusinessID: 1})

	items := service.QueryBehaviorLogs(BehaviorQuery{EventCode: "search"})
	if len(items) != 1 || items[0].EventCode != "search" {
		t.Fatalf("expected eventCode filter, got %+v", items)
	}
}

func TestBehaviorRepositoryIsUsedWhenConfigured(t *testing.T) {
	repo := &fakeBehaviorRepository{}
	service := NewServiceWithBehaviorRepository(repo)

	service.RecordBehavior(BehaviorRequest{UserID: 1, EventCode: "search", BusinessType: "game", BusinessID: 1})
	if !repo.saved {
		t.Fatal("expected behavior log saved through repository")
	}
	_ = service.BehaviorLogs()
	if !repo.listed {
		t.Fatal("expected behavior list read through repository")
	}
	_ = service.QueryBehaviorLogs(BehaviorQuery{EventCode: "search"})
	if !repo.queried {
		t.Fatal("expected behavior query read through repository")
	}
}

func TestFunnelAndRetentionSnapshots(t *testing.T) {
	service := NewService()
	day0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	day1 := day0.AddDate(0, 0, 1)
	service.RecordBehavior(BehaviorRequest{UserID: 1, EventCode: "login", BusinessType: "user", BusinessID: 1, OccurredAt: day0})
	service.RecordBehavior(BehaviorRequest{UserID: 1, EventCode: "browse_games", BusinessType: "game", OccurredAt: day0})
	service.RecordBehavior(BehaviorRequest{UserID: 1, EventCode: "login", BusinessType: "user", BusinessID: 1, OccurredAt: day1})
	service.RecordBehavior(BehaviorRequest{UserID: 2, EventCode: "login", BusinessType: "user", BusinessID: 2, OccurredAt: day0})

	funnel := service.Funnel([]string{"login", "browse_games"})
	if len(funnel.Steps) != 2 || funnel.Steps[0].UserCount != 2 || funnel.Steps[1].UserCount != 1 || funnel.Steps[1].ConversionRate != 0.5 || funnel.Steps[1].DropOffRate != 0.5 {
		t.Fatalf("unexpected funnel snapshot: %+v", funnel)
	}

	retention := service.Retention()
	if len(retention.Buckets) != 1 || retention.Buckets[0].NewUsers != 2 || retention.Buckets[0].Day1Retained != 1 || retention.Buckets[0].Day1Rate != 0.5 {
		t.Fatalf("unexpected retention snapshot: %+v", retention)
	}
}

func TestOperationRepositoryIsUsedWhenConfigured(t *testing.T) {
	repo := &fakeOperationRepository{}
	service := NewServiceWithRepositories(nil, repo)

	service.RecordOperation(OperationRequest{AdminUserID: 9, Action: "report:handle", TargetType: "report", TargetID: "1"})
	if !repo.saved {
		t.Fatal("expected operation log saved through repository")
	}
	items := service.OperationLogs()
	if !repo.listed || len(items) != 1 || items[0].Action != "report:handle" {
		t.Fatalf("expected operation list read through repository, listed=%v items=%+v", repo.listed, items)
	}
}

func TestStrictAuditListsDoNotUseStaleCache(t *testing.T) {
	readErr := errors.New("audit database unavailable")
	behaviorRepo := &fakeBehaviorRepository{readErr: readErr}
	operationRepo := &fakeOperationRepository{readErr: readErr}
	service := NewServiceWithRepositories(behaviorRepo, operationRepo)
	service.behaviorLogs = []BehaviorLog{{ID: 1, EventType: "旧行为"}}
	service.operationLogs = []OperationLog{{ID: 1, Action: "旧操作"}}

	if _, err := service.BehaviorLogsStrict(); !errors.Is(err, readErr) {
		t.Fatalf("expected behavior log read error, got %v", err)
	}
	if _, err := service.QueryBehaviorLogsStrict(BehaviorQuery{}); !errors.Is(err, readErr) {
		t.Fatalf("expected behavior query error, got %v", err)
	}
	if _, err := service.OperationLogsStrict(); !errors.Is(err, readErr) {
		t.Fatalf("expected operation log read error, got %v", err)
	}
}

func TestStrictAuditWritesDoNotCreateLocalSuccessOnRepositoryFailure(t *testing.T) {
	writeErr := errors.New("audit persistence unavailable")
	behaviorRepo := &fakeBehaviorRepository{saveErr: writeErr}
	operationRepo := &fakeOperationRepository{saveErr: writeErr}
	service := NewServiceWithRepositories(behaviorRepo, operationRepo)

	if _, err := service.RecordBehaviorStrict(BehaviorRequest{EventType: "login"}); !errors.Is(err, writeErr) {
		t.Fatalf("expected behavior write error, got %v", err)
	}
	if _, err := service.RecordOperationStrict(OperationRequest{Action: "user:update"}); !errors.Is(err, writeErr) {
		t.Fatalf("expected operation write error, got %v", err)
	}
	if len(service.behaviorLogs) != 0 || len(service.operationLogs) != 0 {
		t.Fatalf("failed writes must not create local audit success: behavior=%+v operation=%+v", service.behaviorLogs, service.operationLogs)
	}
}

type fakeBehaviorRepository struct {
	saved   bool
	listed  bool
	queried bool
	items   []BehaviorLog
	readErr error
	saveErr error
}

func (f *fakeBehaviorRepository) SaveBehavior(_ context.Context, log BehaviorLog) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = true
	f.items = append(f.items, log)
	return nil
}

func (f *fakeBehaviorRepository) ListBehavior(_ context.Context) ([]BehaviorLog, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	f.listed = true
	return f.items, nil
}

func (f *fakeBehaviorRepository) QueryBehavior(_ context.Context, _ BehaviorQuery) ([]BehaviorLog, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	f.queried = true
	return f.items, nil
}

type fakeOperationRepository struct {
	saved   bool
	listed  bool
	items   []OperationLog
	readErr error
	saveErr error
}

func (f *fakeOperationRepository) SaveOperation(_ context.Context, log OperationLog) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = true
	f.items = append(f.items, log)
	return nil
}

func (f *fakeOperationRepository) ListOperations(_ context.Context) ([]OperationLog, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	f.listed = true
	return f.items, nil
}
