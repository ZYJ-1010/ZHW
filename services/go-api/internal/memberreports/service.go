package memberreports

import (
	"context"
	"errors"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/revenue"
)

var ErrMemberReportForbidden = errors.New("member report forbidden")

type GameStatsProvider interface {
	StatsForUser(userID int64) games.UserStats
}

type RevenueProvider interface {
	IncomeSummary(userID int64) revenue.IncomeSummary
}

type Repository interface {
	SaveMembership(ctx context.Context, membership Membership) (Membership, error)
	GetMembership(ctx context.Context, userID int64) (Membership, bool, error)
	ListMemberships(ctx context.Context) ([]Membership, error)
	SaveSnapshot(ctx context.Context, snapshot Snapshot) (Snapshot, error)
}

type Membership struct {
	UserID       int64
	PlanName     string
	InvitedCount int
	UpdatedAt    time.Time
}

type Snapshot struct {
	ID             int64                 `json:"id"`
	UserID         int64                 `json:"userId"`
	ReportType     string                `json:"reportType"`
	Period         string                `json:"period"`
	Participated   int                   `json:"participatedGames"`
	Completed      int                   `json:"completedGames"`
	AverageScore   float64               `json:"averageScore"`
	IncomeSummary  revenue.IncomeSummary `json:"incomeSummary"`
	InvitedCount   int                   `json:"invitedCount"`
	MembershipPlan string                `json:"membershipPlan"`
	CreatedAt      time.Time             `json:"createdAt"`
}

type Service struct {
	mu          sync.RWMutex
	nextID      int64
	memberships map[int64]string
	invites     map[int64]int
	snapshots   map[int64]Snapshot
	games       GameStatsProvider
	revenue     RevenueProvider
	repo        Repository
}

func NewService(games GameStatsProvider, revenue RevenueProvider) *Service {
	return NewServiceWithRepository(games, revenue, nil)
}

func NewServiceWithRepository(games GameStatsProvider, revenue RevenueProvider, repo Repository) *Service {
	return &Service{
		nextID:      1,
		memberships: make(map[int64]string),
		invites:     make(map[int64]int),
		snapshots:   make(map[int64]Snapshot),
		games:       games,
		revenue:     revenue,
		repo:        repo,
	}
}

func (s *Service) GrantMembership(userID int64, planName string, invitedCount int) {
	if planName == "" {
		planName = "Basic Member"
	}
	if s.repo != nil {
		if membership, err := s.repo.SaveMembership(context.Background(), Membership{UserID: userID, PlanName: planName, InvitedCount: invitedCount, UpdatedAt: time.Now()}); err == nil {
			planName = membership.PlanName
			invitedCount = membership.InvitedCount
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memberships[userID] = planName
	s.invites[userID] = invitedCount
}

func (s *Service) Me(userID int64) (Snapshot, error) {
	if s.repo != nil {
		membership, ok, err := s.repo.GetMembership(context.Background(), userID)
		if err != nil {
			return Snapshot{}, err
		}
		if !ok || membership.PlanName == "" {
			return Snapshot{}, ErrMemberReportForbidden
		}
		return s.saveRepositorySnapshot(s.buildSnapshot(userID, membership.PlanName, membership.InvitedCount))
	}
	s.mu.RLock()
	planName := s.memberships[userID]
	s.mu.RUnlock()
	if planName == "" {
		return Snapshot{}, ErrMemberReportForbidden
	}
	return s.BuildSnapshot(userID), nil
}

func (s *Service) BuildSnapshot(userID int64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	planName := s.memberships[userID]
	invitedCount := s.invites[userID]
	snapshot := s.buildSnapshot(userID, planName, invitedCount)
	snapshot.ID = s.nextID
	s.nextID++
	s.snapshots[userID] = snapshot
	return snapshot
}

func (s *Service) AdminSnapshots() []Snapshot {
	if s.repo != nil {
		memberships, err := s.repo.ListMemberships(context.Background())
		if err == nil {
			result := make([]Snapshot, 0, len(memberships))
			for _, membership := range memberships {
				snapshot, err := s.saveRepositorySnapshot(s.buildSnapshot(membership.UserID, membership.PlanName, membership.InvitedCount))
				if err == nil {
					result = append(result, snapshot)
				}
			}
			return result
		}
	}
	s.mu.RLock()
	userIDs := make([]int64, 0, len(s.memberships))
	for userID := range s.memberships {
		userIDs = append(userIDs, userID)
	}
	s.mu.RUnlock()

	result := make([]Snapshot, 0, len(userIDs))
	for _, userID := range userIDs {
		result = append(result, s.BuildSnapshot(userID))
	}
	return result
}

func (s *Service) buildSnapshot(userID int64, planName string, invitedCount int) Snapshot {
	stats := s.games.StatsForUser(userID)
	income := s.revenue.IncomeSummary(userID)
	return Snapshot{
		UserID:         userID,
		ReportType:     "monthly",
		Period:         time.Now().Format("2006-01"),
		Participated:   stats.Participated,
		Completed:      stats.Completed,
		AverageScore:   stats.AverageScore,
		IncomeSummary:  income,
		InvitedCount:   invitedCount,
		MembershipPlan: planName,
		CreatedAt:      time.Now(),
	}
}

func (s *Service) saveRepositorySnapshot(snapshot Snapshot) (Snapshot, error) {
	snapshot, err := s.repo.SaveSnapshot(context.Background(), snapshot)
	if err != nil {
		return Snapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snapshot.UserID] = snapshot
	return snapshot, nil
}
