package reports

import (
	"context"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/revenue"
)

type fakeFreezer struct {
	called bool
	gameID int64
	reason string
}

func (f *fakeFreezer) FreezeByGame(gameID int64, reason string) (revenue.Record, bool, error) {
	f.called = true
	f.gameID = gameID
	f.reason = reason
	return revenue.Record{ID: 3, GameID: gameID, Status: "frozen"}, true, nil
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

type fakeReportRepository struct {
	items        map[int64]Report
	saved        bool
	listed       bool
	listedByUser bool
	found        bool
	updated      bool
}

func (r *fakeReportRepository) SaveReport(ctx context.Context, report Report) (Report, error) {
	r.saved = true
	report.ID = 10
	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	r.items[report.ID] = report
	return report, nil
}

func (r *fakeReportRepository) ListReports(ctx context.Context) ([]Report, error) {
	r.listed = true
	return []Report{r.items[9]}, nil
}

func (r *fakeReportRepository) ListReportsByUser(ctx context.Context, userID int64) ([]Report, error) {
	r.listedByUser = true
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
