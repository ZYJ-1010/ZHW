package reports

import (
	"context"
	"errors"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/revenue"
)

type fakeFreezer struct {
	called        bool
	gameID        int64
	reason        string
	frozen        bool
	restoreCalled bool
	restoreReason string
	restoreErr    error
}

func (f *fakeFreezer) FreezeByGame(gameID int64, reason string) (revenue.Record, bool, error) {
	f.called = true
	f.gameID = gameID
	f.reason = reason
	if f.frozen {
		return revenue.Record{ID: 3, GameID: gameID, Status: "frozen"}, false, nil
	}
	f.frozen = true
	return revenue.Record{ID: 3, GameID: gameID, Status: "frozen"}, true, nil
}

func (f *fakeFreezer) RestoreFrozenByGame(gameID int64, reason string) (revenue.Record, bool, error) {
	f.restoreCalled = true
	f.restoreReason = reason
	if f.restoreErr != nil {
		return revenue.Record{}, false, f.restoreErr
	}
	if !f.frozen {
		return revenue.Record{ID: 3, GameID: gameID, Status: "pending_settlement"}, false, nil
	}
	f.frozen = false
	return revenue.Record{ID: 3, GameID: gameID, Status: "pending_settlement"}, true, nil
}

func TestCreateReportStoresEvidenceAndFreezesRevenue(t *testing.T) {
	freezer := &fakeFreezer{}
	service := NewService(freezer)

	report, err := service.Create(2, CreateRequest{
		GameID:          10,
		TargetUserID:    1,
		ReportType:      "service_dispute",
		Content:         "dispute",
		ChatMessageID:   20,
		FileID:          30,
		ReviewID:        40,
		RevenueRecordID: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !freezer.called || freezer.gameID != 10 || freezer.reason != "report_created" {
		t.Fatalf("expected freezer called, got %+v", freezer)
	}
	if !report.RevenueFrozen || report.RevenueFreezeNote != "report_created" {
		t.Fatalf("expected frozen report, got %+v", report)
	}
	if report.ChatMessageID != 20 || report.FileID != 30 || report.ReviewID != 40 || report.RevenueRecordID != 50 {
		t.Fatalf("expected evidence ids stored, got %+v", report)
	}
}

func TestCreateReportRestoresNewFreezeWhenRepositorySaveFails(t *testing.T) {
	freezer := &fakeFreezer{}
	repoErr := errors.New("save report failed")
	repo := &fakeReportRepository{items: map[int64]Report{}, saveErr: repoErr}
	service := NewServiceWithRepository(freezer, repo)

	_, err := service.Create(2, CreateRequest{GameID: 10, ReportType: "service_dispute", Content: "dispute"})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
	if !freezer.restoreCalled || freezer.restoreReason != "report_create_rollback" || freezer.frozen {
		t.Fatalf("expected newly frozen revenue restored, freezer=%+v", freezer)
	}
}

func TestCreateReportReturnsRollbackErrorWhenRevenueRestoreFails(t *testing.T) {
	rollbackErr := errors.New("restore revenue failed")
	freezer := &fakeFreezer{restoreErr: rollbackErr}
	repo := &fakeReportRepository{items: map[int64]Report{}, saveErr: errors.New("save report failed")}
	service := NewServiceWithRepository(freezer, repo)

	_, err := service.Create(2, CreateRequest{GameID: 10, ReportType: "service_dispute", Content: "dispute"})
	if !errors.Is(err, ErrRevenueRollback) || !errors.Is(err, rollbackErr) {
		t.Fatalf("expected identifiable rollback failure, got %v", err)
	}
}

func TestCreateReportDoesNotRestorePreexistingFreezeWhenSaveFails(t *testing.T) {
	freezer := &fakeFreezer{frozen: true}
	repo := &fakeReportRepository{items: map[int64]Report{}, saveErr: errors.New("save report failed")}
	service := NewServiceWithRepository(freezer, repo)

	if _, err := service.Create(2, CreateRequest{GameID: 10, ReportType: "service_dispute", Content: "dispute"}); err == nil {
		t.Fatal("expected repository error")
	}
	if freezer.restoreCalled || !freezer.frozen {
		t.Fatalf("expected preexisting freeze preserved, freezer=%+v", freezer)
	}
}

func TestHandleAndCloseReportUpdateStatus(t *testing.T) {
	service := NewService(nil)
	report, err := service.Create(2, CreateRequest{GameID: 10, ReportType: "service_dispute", Content: "dispute"})
	if err != nil {
		t.Fatal(err)
	}

	assigned, err := service.Assign(report.ID, AssignRequest{AdminID: 88, HandlerAdminID: 99})
	if err != nil {
		t.Fatal(err)
	}
	if assigned.Status != "assigned" || assigned.HandlerAdminID != 99 || assigned.HandleResult != "" || assigned.HandledAt != "" {
		t.Fatalf("expected assigned report, got %+v", assigned)
	}
	if _, err := service.Handle(report.ID, HandleRequest{AdminID: 99, Outcome: "appeal_rejected"}); err != ErrInvalidReport {
		t.Fatalf("expected rejected appeal to require a reason, got %v", err)
	}

	handled, err := service.Handle(report.ID, HandleRequest{AdminID: 99, Result: "notified"})
	if err != nil {
		t.Fatal(err)
	}
	if handled.Status != "handled" || handled.HandlerAdminID != 99 || handled.HandleResult != "notified" || handled.HandledAt == "" {
		t.Fatalf("expected handled report, got %+v", handled)
	}

	closed, err := service.Close(report.ID, HandleRequest{AdminID: 99, Result: "closed"})
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != "closed" || closed.HandleResult != "closed" || closed.HandledAt == "" {
		t.Fatalf("expected closed report, got %+v", closed)
	}
}

func TestAppealStoresStructuredFileIDs(t *testing.T) {
	service := NewService(nil)
	report, err := service.Create(2, CreateRequest{GameID: 10, TargetUserID: 1, ReportType: "service_dispute", Content: "dispute"})
	if err != nil {
		t.Fatal(err)
	}

	appealed, err := service.Appeal(1, report.ID, AppealRequest{Content: "not true", FileIDs: []int64{30, 31, 30}})
	if err != nil {
		t.Fatal(err)
	}
	if appealed.Status != "appealed" || appealed.HandleResult != "not true" || appealed.FileID != 30 || len(appealed.AppealFileIDs) != 2 || appealed.AppealFileIDs[1] != 31 {
		t.Fatalf("expected structured appeal files, got %+v", appealed)
	}
}

func TestAssignAppealedReportPreservesAppealEvidence(t *testing.T) {
	service := NewService(nil)
	report, err := service.Create(2, CreateRequest{GameID: 10, TargetUserID: 1, ReportType: "service_dispute", Content: "dispute"})
	if err != nil {
		t.Fatal(err)
	}
	appealed, err := service.Appeal(1, report.ID, AppealRequest{Content: "appeal evidence", FileIDs: []int64{30, 31}})
	if err != nil {
		t.Fatal(err)
	}
	assigned, err := service.Assign(report.ID, AssignRequest{AdminID: 88, HandlerAdminID: 99})
	if err != nil {
		t.Fatal(err)
	}
	if assigned.Status != "appealed" || assigned.HandlerAdminID != 99 || assigned.HandleResult != appealed.HandleResult || assigned.HandledAt != appealed.HandledAt || len(assigned.AppealFileIDs) != 2 {
		t.Fatalf("expected assignment to preserve appeal state and evidence, got %+v", assigned)
	}
}

func TestAssignTerminalReportIsRejected(t *testing.T) {
	service := NewService(nil)
	report, err := service.Create(2, CreateRequest{GameID: 10, ReportType: "service_dispute", Content: "dispute"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Handle(report.ID, HandleRequest{AdminID: 88, Result: "done"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Assign(report.ID, AssignRequest{AdminID: 88, HandlerAdminID: 99}); err != ErrInvalidReport {
		t.Fatalf("expected handled report assignment rejected, got %v", err)
	}
}

func TestCreditAppealCanOnlyBeSubmittedOncePerCreditLog(t *testing.T) {
	service := NewService(nil)
	request := CreditAppealRequest{GameID: 10, CreditLogID: 88, Content: "申请复核"}
	first, err := service.CreateCreditAppeal(2, request)
	if err != nil || first.ReportType != "credit_appeal" {
		t.Fatalf("expected first credit appeal accepted, report=%+v err=%v", first, err)
	}
	if _, err := service.CreateCreditAppeal(2, request); err != ErrCreditAppealExists {
		t.Fatalf("expected duplicate credit appeal rejected, got %v", err)
	}
	if _, err := service.CreateCreditAppeal(3, request); err != nil {
		t.Fatalf("expected uniqueness to include the user, got %v", err)
	}
}

func TestServiceUsesRepositoryWhenConfigured(t *testing.T) {
	repo := &fakeReportRepository{
		items: map[int64]Report{
			9: {
				ID:             9,
				GameID:         10,
				ReporterUserID: 2,
				ReportType:     "service_dispute",
				Status:         "pending",
				CreatedAt:      time.Now(),
			},
		},
	}
	service := NewServiceWithRepository(nil, repo)

	report, err := service.Create(2, CreateRequest{GameID: 10, ReportType: "service_dispute", Content: "dispute"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.saved || report.ID != 10 {
		t.Fatalf("expected repository save, saved=%v report=%+v", repo.saved, report)
	}

	if items := service.My(2); !repo.listedByUser || len(items) != 1 || items[0].ID != 9 {
		t.Fatalf("expected repository user list, listed=%v items=%+v", repo.listedByUser, items)
	}
	if items := service.List(); !repo.listed || len(items) != 1 || items[0].ID != 9 {
		t.Fatalf("expected repository list, listed=%v items=%+v", repo.listed, items)
	}
	got, err := service.Get(9)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.found || got.ID != 9 {
		t.Fatalf("expected repository get, found=%v report=%+v", repo.found, got)
	}
	assigned, err := service.Assign(9, AssignRequest{AdminID: 88, HandlerAdminID: 99})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.updated || assigned.Status != "assigned" || assigned.HandlerAdminID != 99 {
		t.Fatalf("expected repository assign, updated=%v report=%+v", repo.updated, assigned)
	}
	handled, err := service.Handle(9, HandleRequest{AdminID: 99, Result: "notified"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.updated || handled.Status != "handled" || handled.HandlerAdminID != 99 {
		t.Fatalf("expected repository update, updated=%v report=%+v", repo.updated, handled)
	}
}

func TestStrictReportListsDoNotFallbackToProcessMemory(t *testing.T) {
	repoErr := errors.New("report database unavailable")
	repo := &fakeReportRepository{items: map[int64]Report{}, listErr: repoErr}
	service := NewServiceWithRepository(nil, repo)
	service.reports[1] = Report{ID: 1, ReporterUserID: 2, ReportType: "service_dispute"}

	if items, err := service.MyStrict(2); !errors.Is(err, repoErr) || items != nil {
		t.Fatalf("expected user list database error without memory fallback, items=%+v err=%v", items, err)
	}
	if items, err := service.ListStrict(); !errors.Is(err, repoErr) || items != nil {
		t.Fatalf("expected admin list database error without memory fallback, items=%+v err=%v", items, err)
	}
	if items, err := service.AppealsStrict(2); !errors.Is(err, repoErr) || items != nil {
		t.Fatalf("expected appeal list database error without memory fallback, items=%+v err=%v", items, err)
	}
}

type fakeReportRepository struct {
	items        map[int64]Report
	saved        bool
	listed       bool
	listedByUser bool
	found        bool
	updated      bool
	saveErr      error
	listErr      error
}

func (r *fakeReportRepository) SaveReport(ctx context.Context, report Report) (Report, error) {
	r.saved = true
	if r.saveErr != nil {
		return Report{}, r.saveErr
	}
	report.ID = 10
	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	r.items[report.ID] = report
	return report, nil
}

func (r *fakeReportRepository) ListReports(ctx context.Context) ([]Report, error) {
	r.listed = true
	if r.listErr != nil {
		return nil, r.listErr
	}
	return []Report{r.items[9]}, nil
}

func (r *fakeReportRepository) ListReportsByUser(ctx context.Context, userID int64) ([]Report, error) {
	r.listedByUser = true
	if r.listErr != nil {
		return nil, r.listErr
	}
	return []Report{r.items[9]}, nil
}

func (r *fakeReportRepository) FindReport(ctx context.Context, reportID int64) (Report, bool, error) {
	r.found = true
	report, ok := r.items[reportID]
	return report, ok, nil
}

func (r *fakeReportRepository) UpdateReport(ctx context.Context, report Report) (Report, error) {
	r.updated = true
	r.items[report.ID] = report
	return report, nil
}
