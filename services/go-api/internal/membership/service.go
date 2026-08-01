package membership

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrPlanNotFound      = errors.New("membership plan not found")
	ErrInvalidMembership = errors.New("invalid membership")
)

type Plan struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type Membership struct {
	ID        int64      `json:"id,omitempty"`
	UserID    int64      `json:"userId"`
	PlanCode  string     `json:"planCode"`
	PlanName  string     `json:"planName"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"startedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type Repository interface {
	ListPlans(ctx context.Context) ([]Plan, error)
	UpsertPlan(ctx context.Context, plan Plan) (Plan, error)
	GetMembership(ctx context.Context, userID int64) (Membership, bool, error)
	SaveMembership(ctx context.Context, membership Membership) (Membership, error)
}

type Service struct {
	mu           sync.RWMutex
	nextPlanID   int64
	nextMemberID int64
	plans        map[string]Plan
	memberships  map[int64]Membership
	repo         Repository
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	service := &Service{
		nextPlanID:   1,
		nextMemberID: 1,
		plans:        make(map[string]Plan),
		memberships:  make(map[int64]Membership),
		repo:         repo,
	}
	service.seedDefaultPlans()
	return service
}

func (s *Service) Plans() []Plan {
	plans, _ := s.PlansStrict()
	return plans
}

func (s *Service) PlansStrict() ([]Plan, error) {
	if s.repo != nil {
		return s.repo.ListPlans(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Plan, 0, len(s.plans))
	for _, plan := range s.plans {
		items = append(items, plan)
	}
	return items, nil
}

func (s *Service) My(userID int64) Membership {
	membership, _ := s.MyStrict(userID)
	return membership
}

func (s *Service) MyStrict(userID int64) (Membership, error) {
	if s.repo != nil {
		membership, ok, err := s.repo.GetMembership(context.Background(), userID)
		if err != nil {
			return Membership{}, err
		}
		if ok {
			return membership, nil
		}
		return Membership{UserID: userID, PlanCode: "none", PlanName: "none", Status: "none"}, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if membership, ok := s.memberships[userID]; ok {
		return membership, nil
	}
	return Membership{UserID: userID, PlanCode: "none", PlanName: "none", Status: "none"}, nil
}

func (s *Service) Grant(userID int64, planCode string, expiresAt *time.Time) (Membership, error) {
	planCode = strings.TrimSpace(planCode)
	if userID <= 0 || planCode == "" {
		return Membership{}, ErrInvalidMembership
	}
	plan, ok, err := s.findPlanStrict(planCode)
	if err != nil {
		return Membership{}, err
	}
	if !ok || plan.Status != "active" {
		return Membership{}, ErrPlanNotFound
	}
	membership := Membership{
		UserID:    userID,
		PlanCode:  plan.Code,
		PlanName:  plan.Name,
		Status:    "active",
		StartedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	if s.repo != nil {
		saved, err := s.repo.SaveMembership(context.Background(), membership)
		if err != nil {
			return Membership{}, err
		}
		membership = saved
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if membership.ID == 0 {
		membership.ID = s.nextMemberID
		s.nextMemberID++
	}
	s.memberships[userID] = membership
	return membership, nil
}

func (s *Service) seedDefaultPlans() {
	_, _ = s.upsertPlan(Plan{Code: "basic", Name: "Basic Member", Status: "active", CreatedAt: time.Now()})
	_, _ = s.upsertPlan(Plan{Code: "pro", Name: "Pro Member", Status: "active", CreatedAt: time.Now()})
}

func (s *Service) findPlan(code string) (Plan, bool) {
	plan, ok, _ := s.findPlanStrict(code)
	return plan, ok
}

func (s *Service) findPlanStrict(code string) (Plan, bool, error) {
	if s.repo != nil {
		plans, err := s.repo.ListPlans(context.Background())
		if err == nil {
			for _, plan := range plans {
				if plan.Code == code {
					return plan, true, nil
				}
			}
			return Plan{}, false, nil
		}
		return Plan{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	plan, ok := s.plans[code]
	return plan, ok, nil
}

func (s *Service) upsertPlan(plan Plan) (Plan, error) {
	plan.Code = strings.TrimSpace(plan.Code)
	plan.Name = strings.TrimSpace(plan.Name)
	plan.Status = strings.TrimSpace(plan.Status)
	if plan.Code == "" || plan.Name == "" {
		return Plan{}, ErrInvalidMembership
	}
	if plan.Status == "" {
		plan.Status = "active"
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = time.Now()
	}
	if s.repo != nil {
		return s.repo.UpsertPlan(context.Background(), plan)
	}
	if plan.ID == 0 {
		plan.ID = s.nextPlanID
		s.nextPlanID++
	}
	s.plans[plan.Code] = plan
	return plan, nil
}
