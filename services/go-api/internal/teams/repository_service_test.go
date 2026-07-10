package teams

import (
	"context"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/revenue"
)

func TestServiceWithRepositoryPersistsTeamAndMembers(t *testing.T) {
	repo := newFakeTeamRepository()
	service := NewServiceWithRepository(fakeTeamRevenue{}, repo)

	team := service.GrantLeader(10, "leader team")
	if team.ID == 0 || team.LeaderUserID != 10 {
		t.Fatalf("unexpected team: %+v", team)
	}

	member, err := service.AddMember(10, 20, "invite")
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	if member.UserID != 20 || member.Status != "active" {
		t.Fatalf("unexpected member: %+v", member)
	}

	if _, err := service.AddMember(99, 21, "invite"); err != ErrTeamForbidden {
		t.Fatalf("expected forbidden for unknown leader, got %v", err)
	}

	members, err := service.Members(10)
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 1 || members[0].UserID != 20 {
		t.Fatalf("unexpected members: %+v", members)
	}

	detail, err := service.Detail(team.ID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detail.RevenueSummary.TotalCent != 2000 || detail.RevenueSummary.PendingCent != 500 || detail.RevenueSummary.SettledCent != 1500 {
		t.Fatalf("unexpected revenue summary: %+v", detail.RevenueSummary)
	}
}

type fakeTeamRevenue struct{}

func (fakeTeamRevenue) IncomeSummary(userID int64) revenue.IncomeSummary {
	return revenue.IncomeSummary{UserID: userID, TotalCent: userID * 100, PendingCent: 500, SettledCent: userID*100 - 500}
}

type fakeTeamRepository struct {
	nextID  int64
	teams   map[int64]Team
	members map[int64][]Member
}

func newFakeTeamRepository() *fakeTeamRepository {
	return &fakeTeamRepository{nextID: 1, teams: make(map[int64]Team), members: make(map[int64][]Member)}
}

func (r *fakeTeamRepository) SaveTeam(ctx context.Context, team Team) (Team, error) {
	if existing, ok := r.teams[team.LeaderUserID]; ok {
		existing.Name = team.Name
		existing.Status = team.Status
		r.teams[team.LeaderUserID] = existing
		return existing, nil
	}
	team.ID = r.nextID
	r.nextID++
	if team.CreatedAt.IsZero() {
		team.CreatedAt = time.Now()
	}
	r.teams[team.LeaderUserID] = team
	return team, nil
}

func (r *fakeTeamRepository) FindTeamByLeader(ctx context.Context, leaderUserID int64) (Team, bool, error) {
	team, ok := r.teams[leaderUserID]
	return team, ok, nil
}

func (r *fakeTeamRepository) FindTeamByID(ctx context.Context, teamID int64) (Team, bool, error) {
	for _, team := range r.teams {
		if team.ID == teamID {
			return team, true, nil
		}
	}
	return Team{}, false, nil
}

func (r *fakeTeamRepository) ListTeams(ctx context.Context) ([]Team, error) {
	items := make([]Team, 0, len(r.teams))
	for _, team := range r.teams {
		items = append(items, team)
	}
	return items, nil
}

func (r *fakeTeamRepository) SaveMember(ctx context.Context, leaderUserID int64, member Member) (Member, error) {
	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}
	for index, item := range r.members[leaderUserID] {
		if item.UserID == member.UserID && item.RelationLevel == member.RelationLevel {
			r.members[leaderUserID][index] = member
			return member, nil
		}
	}
	r.members[leaderUserID] = append(r.members[leaderUserID], member)
	return member, nil
}

func (r *fakeTeamRepository) ListMembers(ctx context.Context, leaderUserID int64) ([]Member, error) {
	return append([]Member(nil), r.members[leaderUserID]...), nil
}
