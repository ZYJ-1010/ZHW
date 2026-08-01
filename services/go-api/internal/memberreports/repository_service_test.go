package memberreports

import (
	"context"
	"errors"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/revenue"
)

func TestServiceWithRepositoryBuildsMemberSnapshots(t *testing.T) {
	repo := newFakeMemberReportRepository()
	service := NewServiceWithRepository(fakeMemberReportGames{}, fakeMemberReportRevenue{}, repo)

	if _, err := service.Me(10); err != ErrMemberReportForbidden {
		t.Fatalf("expected forbidden before membership grant, got %v", err)
	}

	service.GrantMembership(10, "pro", 3)
	snapshot, err := service.Me(10)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if snapshot.ID == 0 || snapshot.MembershipPlan != "pro" || snapshot.InvitedCount != 3 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	if snapshot.Participated != 4 || snapshot.Completed != 2 || snapshot.IncomeSummary.TotalCent != 1000 {
		t.Fatalf("unexpected metrics: %+v", snapshot)
	}

	adminSnapshots := service.AdminSnapshots()
	if len(adminSnapshots) != 1 {
		t.Fatalf("unexpected admin snapshots: %+v", adminSnapshots)
	}
	if len(repo.snapshots) != 2 {
		t.Fatalf("expected persisted personal and admin snapshots, got %d", len(repo.snapshots))
	}
}

func TestRepositoryFailureDoesNotCreateLocalMembershipOrSnapshots(t *testing.T) {
	service := NewServiceWithRepository(fakeMemberReportGames{}, fakeMemberReportRevenue{}, failingMemberReportRepository{})
	if err := service.GrantMembershipStrict(10, "pro", 3); err == nil {
		t.Fatal("expected membership persistence failure")
	}
	if _, err := service.Me(10); err == nil {
		t.Fatal("membership must not appear after persistence failure")
	}
	if _, err := service.AdminSnapshotsStrict(); err == nil {
		t.Fatal("expected strict report list to return repository failure")
	}
}

type failingMemberReportRepository struct{}

func (failingMemberReportRepository) SaveMembership(context.Context, Membership) (Membership, error) {
	return Membership{}, errors.New("repository unavailable")
}

func (failingMemberReportRepository) GetMembership(context.Context, int64) (Membership, bool, error) {
	return Membership{}, false, errors.New("repository unavailable")
}

func (failingMemberReportRepository) ListMemberships(context.Context) ([]Membership, error) {
	return nil, errors.New("repository unavailable")
}

func (failingMemberReportRepository) SaveSnapshot(context.Context, Snapshot) (Snapshot, error) {
	return Snapshot{}, errors.New("repository unavailable")
}

type fakeMemberReportGames struct{}

func (fakeMemberReportGames) StatsForUser(userID int64) games.UserStats {
	return games.UserStats{UserID: userID, Participated: 4, Completed: 2, AverageScore: 4.5}
}

type fakeMemberReportRevenue struct{}

func (fakeMemberReportRevenue) IncomeSummary(userID int64) revenue.IncomeSummary {
	return revenue.IncomeSummary{UserID: userID, TotalCent: 1000, PendingCent: 200, SettledCent: 800}
}

type fakeMemberReportRepository struct {
	memberships map[int64]Membership
	snapshots   []Snapshot
	nextID      int64
}

func newFakeMemberReportRepository() *fakeMemberReportRepository {
	return &fakeMemberReportRepository{memberships: make(map[int64]Membership), snapshots: make([]Snapshot, 0), nextID: 1}
}

func (r *fakeMemberReportRepository) SaveMembership(ctx context.Context, membership Membership) (Membership, error) {
	if membership.UpdatedAt.IsZero() {
		membership.UpdatedAt = time.Now()
	}
	r.memberships[membership.UserID] = membership
	return membership, nil
}

func (r *fakeMemberReportRepository) GetMembership(ctx context.Context, userID int64) (Membership, bool, error) {
	membership, ok := r.memberships[userID]
	return membership, ok, nil
}

func (r *fakeMemberReportRepository) ListMemberships(ctx context.Context) ([]Membership, error) {
	items := make([]Membership, 0, len(r.memberships))
	for _, membership := range r.memberships {
		items = append(items, membership)
	}
	return items, nil
}

func (r *fakeMemberReportRepository) SaveSnapshot(ctx context.Context, snapshot Snapshot) (Snapshot, error) {
	snapshot.ID = r.nextID
	r.nextID++
	r.snapshots = append(r.snapshots, snapshot)
	return snapshot, nil
}
