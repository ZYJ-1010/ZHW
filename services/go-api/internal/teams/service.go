package teams

import (
	"context"
	"errors"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/revenue"
)

var ErrTeamForbidden = errors.New("team forbidden")

type RevenueProvider interface {
	IncomeSummary(userID int64) revenue.IncomeSummary
}

type Repository interface {
	SaveTeam(ctx context.Context, team Team) (Team, error)
	FindTeamByLeader(ctx context.Context, leaderUserID int64) (Team, bool, error)
	FindTeamByID(ctx context.Context, teamID int64) (Team, bool, error)
	ListTeams(ctx context.Context) ([]Team, error)
	SaveMember(ctx context.Context, leaderUserID int64, member Member) (Member, error)
	ListMembers(ctx context.Context, leaderUserID int64) ([]Member, error)
}

type Team struct {
	ID           int64     `json:"id"`
	LeaderUserID int64     `json:"leaderUserId"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Member struct {
	UserID        int64     `json:"userId"`
	RelationLevel int       `json:"relationLevel"`
	Source        string    `json:"source"`
	Status        string    `json:"status"`
	JoinedAt      time.Time `json:"joinedAt"`
}

type RevenueSummary struct {
	TeamID       int64 `json:"teamId"`
	LeaderUserID int64 `json:"leaderUserId"`
	MemberCount  int   `json:"memberCount"`
	TotalCent    int64 `json:"totalCent"`
	PendingCent  int64 `json:"pendingCent"`
	SettledCent  int64 `json:"settledCent"`
}

type Detail struct {
	Team           Team           `json:"team"`
	Members        []Member       `json:"members"`
	RevenueSummary RevenueSummary `json:"revenueSummary"`
}

type Service struct {
	mu      sync.RWMutex
	nextID  int64
	teams   map[int64]Team
	members map[int64][]Member
	revenue RevenueProvider
	repo    Repository
}

func NewService(revenue RevenueProvider) *Service {
	return NewServiceWithRepository(revenue, nil)
}

func NewServiceWithRepository(revenue RevenueProvider, repo Repository) *Service {
	return &Service{
		nextID:  1,
		teams:   make(map[int64]Team),
		members: make(map[int64][]Member),
		revenue: revenue,
		repo:    repo,
	}
}

func (s *Service) GrantLeader(userID int64, name string) Team {
	if name == "" {
		name = "My Team"
	}
	if s.repo != nil {
		if team, ok, err := s.repo.FindTeamByLeader(context.Background(), userID); err == nil && ok && team.Status == "active" {
			return team
		}
		team, err := s.repo.SaveTeam(context.Background(), Team{LeaderUserID: userID, Name: name, Status: "active", CreatedAt: time.Now()})
		if err == nil {
			return team
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if team, ok := s.teams[userID]; ok {
		return team
	}
	team := Team{ID: s.nextID, LeaderUserID: userID, Name: name, Status: "active", CreatedAt: time.Now()}
	s.nextID++
	s.teams[userID] = team
	return team
}

func (s *Service) AddMember(leaderUserID int64, memberUserID int64, source string) (Member, error) {
	if source == "" {
		source = "invite"
	}
	if s.repo != nil {
		team, ok, err := s.repo.FindTeamByLeader(context.Background(), leaderUserID)
		if err != nil {
			return Member{}, err
		}
		if !ok || team.Status != "active" {
			return Member{}, ErrTeamForbidden
		}
		return s.repo.SaveMember(context.Background(), leaderUserID, Member{UserID: memberUserID, RelationLevel: 1, Source: source, Status: "active", JoinedAt: time.Now()})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[leaderUserID]; !ok {
		return Member{}, ErrTeamForbidden
	}
	for _, item := range s.members[leaderUserID] {
		if item.UserID == memberUserID && item.Status == "active" {
			return item, nil
		}
	}
	member := Member{UserID: memberUserID, RelationLevel: 1, Source: source, Status: "active", JoinedAt: time.Now()}
	s.members[leaderUserID] = append(s.members[leaderUserID], member)
	return member, nil
}

func (s *Service) My(userID int64) (Team, error) {
	if s.repo != nil {
		team, ok, err := s.repo.FindTeamByLeader(context.Background(), userID)
		if err != nil {
			return Team{}, err
		}
		if !ok || team.Status != "active" {
			return Team{}, ErrTeamForbidden
		}
		return team, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	team, ok := s.teams[userID]
	if !ok || team.Status != "active" {
		return Team{}, ErrTeamForbidden
	}
	return team, nil
}

func (s *Service) All() []Team {
	if s.repo != nil {
		if teams, err := s.repo.ListTeams(context.Background()); err == nil {
			result := make([]Team, 0, len(teams))
			for _, team := range teams {
				if team.Status == "active" {
					result = append(result, team)
				}
			}
			return result
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Team, 0, len(s.teams))
	for _, team := range s.teams {
		if team.Status == "active" {
			result = append(result, team)
		}
	}
	return result
}

func (s *Service) Detail(teamID int64) (Detail, error) {
	if s.repo != nil {
		team, ok, err := s.repo.FindTeamByID(context.Background(), teamID)
		if err != nil {
			return Detail{}, err
		}
		if !ok || team.Status != "active" {
			return Detail{}, ErrTeamForbidden
		}
		members, err := s.repo.ListMembers(context.Background(), team.LeaderUserID)
		if err != nil {
			return Detail{}, err
		}
		members = activeMembers(members)
		return Detail{Team: team, Members: members, RevenueSummary: s.revenueSummaryFor(team, members)}, nil
	}
	s.mu.RLock()
	var team Team
	found := false
	for _, item := range s.teams {
		if item.ID == teamID && item.Status == "active" {
			team = item
			found = true
			break
		}
	}
	if !found {
		s.mu.RUnlock()
		return Detail{}, ErrTeamForbidden
	}
	members := activeMembers(s.members[team.LeaderUserID])
	s.mu.RUnlock()

	return Detail{Team: team, Members: members, RevenueSummary: s.revenueSummaryFor(team, members)}, nil
}

func (s *Service) Members(userID int64) ([]Member, error) {
	if s.repo != nil {
		team, ok, err := s.repo.FindTeamByLeader(context.Background(), userID)
		if err != nil {
			return nil, err
		}
		if !ok || team.Status != "active" {
			return nil, ErrTeamForbidden
		}
		members, err := s.repo.ListMembers(context.Background(), userID)
		if err != nil {
			return nil, err
		}
		return activeMembers(members), nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if team, ok := s.teams[userID]; !ok || team.Status != "active" {
		return nil, ErrTeamForbidden
	}
	return activeMembers(s.members[userID]), nil
}

func (s *Service) RevenueSummary(userID int64) (RevenueSummary, error) {
	team, err := s.My(userID)
	if err != nil {
		return RevenueSummary{}, err
	}
	members, err := s.Members(userID)
	if err != nil {
		return RevenueSummary{}, err
	}
	return s.revenueSummaryFor(team, members), nil
}

func (s *Service) revenueSummaryFor(team Team, members []Member) RevenueSummary {
	summary := RevenueSummary{TeamID: team.ID, LeaderUserID: team.LeaderUserID, MemberCount: len(members)}
	for _, member := range members {
		income := s.revenue.IncomeSummary(member.UserID)
		summary.TotalCent += income.TotalCent
		summary.PendingCent += income.PendingCent
		summary.SettledCent += income.SettledCent
	}
	return summary
}

func activeMembers(members []Member) []Member {
	result := make([]Member, 0, len(members))
	for _, item := range members {
		if item.Status == "active" {
			result = append(result, item)
		}
	}
	return result
}
