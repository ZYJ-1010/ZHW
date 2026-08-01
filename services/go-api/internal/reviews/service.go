package reviews

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

var (
	ErrGameNotReviewable      = errors.New("game not reviewable")
	ErrForbidden              = errors.New("forbidden")
	ErrDuplicateReview        = errors.New("duplicate review")
	ErrInvalidReview          = errors.New("invalid review")
	ErrCreditLogNotFound      = errors.New("credit log not found")
	ErrCreditAppealConflict   = errors.New("credit appeal conflict")
	ErrCreditRulesUnavailable = errors.New("credit account rules unavailable")
)

const growthRulesConfigKey = "growth.reward_rules"

// GrowthRules controls the rewards and level calculation used by the growth
// ledger. The values are persisted in system configuration and can be changed
// by an administrator without rebuilding the service.
type GrowthRules struct {
	CompletedGameExperience   int `json:"completedGameExperience"`
	SubmittedReviewExperience int `json:"submittedReviewExperience"`
	ReceivedReviewExperience  int `json:"receivedReviewExperience"`
	ExperiencePerLevel        int `json:"experiencePerLevel"`
	InitialLevel              int `json:"initialLevel"`
	InitialCreditScore        int `json:"initialCreditScore"`
	CreditScoreCap            int `json:"creditScoreCap"`
}

type GrowthRulesProvider interface {
	Get(key string, target interface{}) bool
}

type GrowthRulesStrictProvider interface {
	GetStrict(key string, target interface{}) (bool, error)
}

// LevelResolver lets the application layer own the role-aware, operator
// configured player ladder without making this domain package depend on the
// HTTP or system-config packages.
type LevelResolver func(experience int) int

// CreditAccountRulesProvider returns the permanent-credit initial score and
// cap maintained by operations, without coupling this package to app config.
type CreditAccountRulesProvider func() (int, int)

// CreditAccountRulesStrictProvider is used by mutations. Credit-score limits
// must not silently fall back to defaults when the operator configuration
// store is unavailable.
type CreditAccountRulesStrictProvider func() (int, int, error)

type GameProvider interface {
	Get(id int64) (games.Game, error)
	Members(gameID int64) []int64
	IsMember(gameID int64, userID int64) bool
}

type GameListProvider interface {
	List() []games.Game
}

type GameRoleProvider interface {
	MemberRoles(gameID int64) []games.MemberRole
}

type GameInvitationProvider interface {
	InvitationsForUser(userID int64) []games.Invitation
}

type Review struct {
	ID             int64     `json:"id"`
	GameID         int64     `json:"gameId"`
	ReviewerUserID int64     `json:"reviewerUserId"`
	TargetUserID   int64     `json:"targetUserId"`
	TargetRole     string    `json:"targetRole"`
	Score          int       `json:"score"`
	Content        string    `json:"content,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	AgainIntent    string    `json:"againIntent,omitempty"`
	NPSScore       *int      `json:"npsScore,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type SubmitRequest struct {
	GameID       int64    `json:"gameId"`
	TargetUserID int64    `json:"targetUserId"`
	TargetRole   string   `json:"targetRole"`
	Score        int      `json:"score"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags"`
	AgainIntent  string   `json:"againIntent"`
	NPSScore     *int     `json:"npsScore"`
}

type Todo struct {
	GameID       int64  `json:"gameId"`
	TargetUserID int64  `json:"targetUserId"`
	TargetRole   string `json:"targetRole"`
	DeadlineAt   string `json:"deadlineAt"`
}

type GrowthProfile struct {
	UserID           int64    `json:"userId"`
	Level            int      `json:"level"`
	Experience       int      `json:"experience"`
	CreditScore      int      `json:"creditScore"`
	TodayCreditScore int      `json:"todayCreditScore"`
	AvailablePoints  int      `json:"availablePoints"`
	ReviewCount      int      `json:"reviewCount"`
	Achievements     []string `json:"achievements"`
	UpdatedAt        string   `json:"updatedAt"`
}

type Footprint struct {
	UserID    int64     `json:"userId"`
	GameID    int64     `json:"gameId"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreditLog struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userId"`
	GameID      int64     `json:"gameId"`
	ChangeValue int       `json:"changeValue"`
	BeforeScore int       `json:"beforeScore"`
	AfterScore  int       `json:"afterScore"`
	Reason      string    `json:"reason"`
	AppealID    int64     `json:"appealId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreditAccount is the user's permanent platform credit balance. The legacy
// todayCreditScore response field is retained only for older clients and
// always mirrors CurrentScore.
type CreditAccount struct {
	UserID       int64     `json:"userId"`
	CurrentScore int       `json:"currentScore"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// GrowthEvent is an append-only, idempotent source record for role metrics.
// It stores factual business evidence rather than a precomputed score, so
// future rule changes can be recalculated against the original event.
type GrowthEvent struct {
	UserID         int64     `json:"userId"`
	GameID         int64     `json:"gameId,omitempty"`
	RoleCode       string    `json:"roleCode"`
	EventCode      string    `json:"eventCode"`
	IdempotencyKey string    `json:"idempotencyKey"`
	OccurredAt     time.Time `json:"occurredAt"`
}

type GrowthMutation struct {
	UserID           int64
	GameID           int64
	ExperienceDelta  int
	ReviewCountDelta int
	Action           string
	IdempotencyKey   string
	Event            *GrowthEvent
	CreatedAt        time.Time
}

type ReviewGrowthBundle struct {
	Review           Review
	ReviewerMutation GrowthMutation
	TargetMutation   GrowthMutation
	Achievements     []Achievement
}

type CreditDeductionRule struct {
	ID          int64     `json:"id"`
	RuleCode    string    `json:"ruleCode"`
	ChangeValue int       `json:"changeValue"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Achievement struct {
	UserID     int64     `json:"userId"`
	Code       string    `json:"code"`
	Title      string    `json:"title"`
	AchievedAt time.Time `json:"achievedAt"`
}

type Trace struct {
	UserID       int64         `json:"userId,omitempty"`
	GameID       int64         `json:"gameId,omitempty"`
	Profile      GrowthProfile `json:"profile,omitempty"`
	Reviews      []Review      `json:"reviews"`
	CreditLogs   []CreditLog   `json:"creditLogs"`
	Footprints   []Footprint   `json:"footprints"`
	Achievements []Achievement `json:"achievements"`
}

type Repository interface {
	MarkReviewable(ctx context.Context, gameID int64, userID int64, deadline time.Time) error
	MarkReviewableBatch(ctx context.Context, gameID int64, userIDs []int64, deadline time.Time) error
	ListReviewable(ctx context.Context, userID int64) (map[int64]time.Time, error)
	ReviewExists(ctx context.Context, gameID int64, reviewerID int64, targetID int64, targetRole string) (bool, error)
	SaveReview(ctx context.Context, review Review) (Review, error)
	SaveReviewWithGrowth(ctx context.Context, bundle ReviewGrowthBundle, initialLevel int) (Review, GrowthProfile, GrowthProfile, []Footprint, []Achievement, error)
	ListReviews(ctx context.Context) ([]Review, error)
	ListReviewsByUser(ctx context.Context, userID int64) ([]Review, error)
	ListReviewsByGame(ctx context.Context, gameID int64) ([]Review, error)
	GetGrowthProfile(ctx context.Context, userID int64) (GrowthProfile, bool, error)
	SaveGrowthProfile(ctx context.Context, profile GrowthProfile) (GrowthProfile, error)
	AddExperienceLog(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, createdAt time.Time) error
	AddPointsLog(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, createdAt time.Time) error
	AddFootprint(ctx context.Context, footprint Footprint) (Footprint, error)
	ListFootprintsByUser(ctx context.Context, userID int64) ([]Footprint, error)
	ListFootprints(ctx context.Context) ([]Footprint, error)
	GetCreditAccount(ctx context.Context, userID int64) (CreditAccount, bool, error)
	EnsureCreditAccount(ctx context.Context, userID int64, initialScore int) (CreditAccount, error)
	SaveCreditAccount(ctx context.Context, account CreditAccount) (CreditAccount, error)
	ApplyCreditChange(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, initialScore int, scoreCap int, createdAt time.Time) (CreditLog, error)
	ApplyCreditChangeOnce(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, idempotencyKey string, initialScore int, scoreCap int, createdAt time.Time) (CreditLog, bool, error)
	RecordGrowthEvent(ctx context.Context, event GrowthEvent) (bool, error)
	ApplyGrowthMutation(ctx context.Context, mutation GrowthMutation, initialLevel int) (GrowthProfile, Footprint, bool, error)
	// GetTodayCredit and SaveTodayCredit remain temporarily for repository
	// compatibility. New credit mutations must use the permanent account APIs.
	GetTodayCredit(ctx context.Context, userID int64, now time.Time) (int, error)
	SaveTodayCredit(ctx context.Context, userID int64, now time.Time, score int) error
	CreditDeductionValue(ctx context.Context, ruleCode string) (int, bool, error)
	ListCreditDeductionRules(ctx context.Context) ([]CreditDeductionRule, error)
	UpsertCreditDeductionRule(ctx context.Context, rule CreditDeductionRule) (CreditDeductionRule, error)
	AddCreditLog(ctx context.Context, log CreditLog) (CreditLog, error)
	RestoreCreditForAppeal(ctx context.Context, userID int64, gameID int64, sourceCreditLogID int64, appealID int64, amount int, initialScore int, scoreCap int, createdAt time.Time) (CreditLog, bool, error)
	ListCreditLogsByUser(ctx context.Context, userID int64) ([]CreditLog, error)
	ListCreditLogsByGame(ctx context.Context, gameID int64) ([]CreditLog, error)
	EnsureAchievement(ctx context.Context, achievement Achievement) (Achievement, bool, error)
	ListAchievementsByUser(ctx context.Context, userID int64) ([]Achievement, error)
	ListAchievements(ctx context.Context) ([]Achievement, error)
}

type Service struct {
	mu                        sync.RWMutex
	nextID                    int64
	nextCreditID              int64
	games                     GameProvider
	reviews                   []Review
	reviewed                  map[string]bool
	reviewable                map[int64]time.Time
	growth                    map[int64]GrowthProfile
	creditScores              map[int64]int
	creditMutationKeys        map[string]CreditLog
	growthEvents              map[string]GrowthEvent
	creditLogs                []CreditLog
	achievements              map[int64][]Achievement
	footprints                []Footprint
	repo                      Repository
	rulesProvider             GrowthRulesProvider
	rulesStrictProvider       GrowthRulesStrictProvider
	levelResolver             LevelResolver
	creditRulesProvider       CreditAccountRulesProvider
	creditRulesStrictProvider CreditAccountRulesStrictProvider
}

func NewService(games GameProvider) *Service {
	return &Service{
		nextID:             1,
		nextCreditID:       1,
		games:              games,
		reviews:            make([]Review, 0),
		reviewed:           make(map[string]bool),
		reviewable:         make(map[int64]time.Time),
		growth:             make(map[int64]GrowthProfile),
		creditScores:       make(map[int64]int),
		creditMutationKeys: make(map[string]CreditLog),
		growthEvents:       make(map[string]GrowthEvent),
		creditLogs:         make([]CreditLog, 0),
		achievements:       make(map[int64][]Achievement),
		footprints:         make([]Footprint, 0),
	}
}

func NewServiceWithRepository(games GameProvider, repo Repository) *Service {
	service := NewService(games)
	service.repo = repo
	return service
}

func (s *Service) Repository() Repository {
	return s.repo
}

// SetGrowthRulesProvider connects the service to the system configuration
// store. It is intentionally a small interface so the reviews package does
// not depend on the app API package.
func (s *Service) SetGrowthRulesProvider(provider GrowthRulesProvider) {
	s.mu.Lock()
	s.rulesProvider = provider
	s.mu.Unlock()
}

func (s *Service) SetGrowthRulesStrictProvider(provider GrowthRulesStrictProvider) {
	s.mu.Lock()
	s.rulesStrictProvider = provider
	s.mu.Unlock()
}

func (s *Service) SetLevelResolver(resolver LevelResolver) {
	s.mu.Lock()
	s.levelResolver = resolver
	s.mu.Unlock()
}

func (s *Service) SetCreditAccountRulesProvider(provider CreditAccountRulesProvider) {
	s.mu.Lock()
	s.creditRulesProvider = provider
	s.mu.Unlock()
}

func (s *Service) SetCreditAccountRulesStrictProvider(provider CreditAccountRulesStrictProvider) {
	s.mu.Lock()
	s.creditRulesStrictProvider = provider
	s.mu.Unlock()
}

func DefaultGrowthRules() GrowthRules {
	return GrowthRules{
		CompletedGameExperience:   100,
		SubmittedReviewExperience: 5,
		ReceivedReviewExperience:  10,
		ExperiencePerLevel:        100,
		InitialLevel:              1,
		InitialCreditScore:        100,
		CreditScoreCap:            100,
	}
}

func (s *Service) GrowthRules() GrowthRules {
	s.mu.RLock()
	provider := s.rulesProvider
	s.mu.RUnlock()
	rules := DefaultGrowthRules()
	if provider != nil {
		var stored GrowthRules
		if provider.Get(growthRulesConfigKey, &stored) {
			rules = normalizeGrowthRules(stored)
		}
	}
	return rules
}

func (s *Service) growthRulesStrict() (GrowthRules, error) {
	s.mu.RLock()
	provider := s.rulesStrictProvider
	s.mu.RUnlock()
	if provider != nil {
		var stored GrowthRules
		found, err := provider.GetStrict(growthRulesConfigKey, &stored)
		if err != nil {
			return GrowthRules{}, err
		}
		if found {
			return normalizeGrowthRules(stored), nil
		}
	}
	return s.GrowthRules(), nil
}

func normalizeGrowthRules(rules GrowthRules) GrowthRules {
	defaults := DefaultGrowthRules()
	if rules.CompletedGameExperience <= 0 {
		rules.CompletedGameExperience = defaults.CompletedGameExperience
	}
	if rules.SubmittedReviewExperience <= 0 {
		rules.SubmittedReviewExperience = defaults.SubmittedReviewExperience
	}
	if rules.ReceivedReviewExperience <= 0 {
		rules.ReceivedReviewExperience = defaults.ReceivedReviewExperience
	}
	if rules.ExperiencePerLevel <= 0 {
		rules.ExperiencePerLevel = defaults.ExperiencePerLevel
	}
	if rules.InitialLevel <= 0 {
		rules.InitialLevel = defaults.InitialLevel
	}
	// The phase-one credit baseline is fixed for every user. It is deliberately
	// not an operational knob so historical accounts and new accounts share the
	// same starting score.
	rules.InitialCreditScore = defaults.InitialCreditScore
	if rules.CreditScoreCap <= 0 {
		rules.CreditScoreCap = defaults.CreditScoreCap
	}
	if rules.CreditScoreCap < rules.InitialCreditScore {
		rules.CreditScoreCap = rules.InitialCreditScore
	}
	return rules
}

func (s *Service) MarkGameReviewable(gameID int64) {
	_ = s.MarkGameReviewableStrict(gameID)
}

func (s *Service) MarkGameReviewableStrict(gameID int64) error {
	members := s.games.Members(gameID)
	s.mu.Lock()
	defer s.mu.Unlock()
	deadline := time.Now().Add(7 * 24 * time.Hour)
	if _, ok := s.reviewable[gameID]; ok {
		return nil
	}
	if s.repo != nil {
		if err := s.repo.MarkReviewableBatch(context.Background(), gameID, members, deadline); err != nil {
			return err
		}
	}
	s.reviewable[gameID] = deadline
	return nil
}

func (s *Service) AwardCompletedGame(gameID int64) []GrowthProfile {
	profiles, _ := s.AwardCompletedGameStrict(gameID)
	return profiles
}

func (s *Service) AwardCompletedGameStrict(gameID int64) ([]GrowthProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rules, rulesErr := s.growthRulesStrictLocked()
	if rulesErr != nil {
		return nil, rulesErr
	}
	profiles := make([]GrowthProfile, 0)
	game, err := s.games.Get(gameID)
	if err != nil {
		return nil, err
	}
	roles := map[int64]string{}
	if provider, ok := s.games.(GameRoleProvider); ok {
		for _, item := range provider.MemberRoles(gameID) {
			roles[item.UserID] = strings.ToLower(strings.TrimSpace(item.Role))
		}
	}
	for _, userID := range s.games.Members(gameID) {
		// 护航领路人是观察者，不占席位、不参与交付，因此不获得完成局经验。
		if isGrowthObserverRole(roles[userID]) {
			continue
		}
		// 积分由 points.Service 作为唯一账本发放；成长档案这里只记录经验和
		// 完成足迹，避免出现“成长积分”和“积分账户”两套余额。
		experience := rules.CompletedGameExperience
		eventCode := "game_completed_member"
		if userID == game.CreatorUserID {
			eventCode = "game_completed_creator"
		}
		event := GrowthEvent{
			UserID: userID, GameID: gameID, RoleCode: growthEventRole(roles[userID]), EventCode: eventCode,
			IdempotencyKey: "completed_game:" + strconv.FormatInt(gameID, 10) + ":" + strconv.FormatInt(userID, 10), OccurredAt: time.Now(),
		}
		profile, applied, err := s.applyExperienceMutationLocked(GrowthMutation{
			UserID: userID, GameID: gameID, ExperienceDelta: experience, Action: "completed_game",
			IdempotencyKey: event.IdempotencyKey, Event: &event, CreatedAt: event.OccurredAt,
		})
		if err != nil {
			return profiles, err
		}
		if applied {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

func isGrowthObserverRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "guide_escort", "guide-escort", "escort", "observer":
		return true
	default:
		return false
	}
}

func growthEventRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "expert", "master":
		return "expert"
	case "guide", "leader", "main_guide":
		return "guide"
	default:
		return "player"
	}
}

func (s *Service) recordGrowthEventLocked(event GrowthEvent) {
	if event.UserID <= 0 || event.EventCode == "" || event.IdempotencyKey == "" {
		return
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	if _, exists := s.growthEvents[event.IdempotencyKey]; exists {
		return
	}
	if s.repo != nil {
		created, err := s.repo.RecordGrowthEvent(context.Background(), event)
		if err != nil || !created {
			return
		}
	}
	s.growthEvents[event.IdempotencyKey] = event
}

// RecordGrowthEvent exposes the same idempotent ledger to adjacent
// application workflows, such as invitation binding, without exposing the
// service's internal maps.
func (s *Service) RecordGrowthEvent(event GrowthEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordGrowthEventLocked(event)
}

// AwardTaskReward records task experience and the completion footprint. Actual
// redeemable points are written by points.Service, which is the single points
// ledger used by the user asset centre.
func (s *Service) AwardTaskReward(userID int64, taskCode string, points int, experience int) GrowthProfile {
	profile, _ := s.AwardTaskRewardStrict(userID, taskCode, points, experience)
	return profile
}

func (s *Service) AwardTaskRewardStrict(userID int64, taskCode string, points int, experience int) (GrowthProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userID <= 0 || taskCode == "" {
		return GrowthProfile{}, ErrInvalidReview
	}
	if _, err := s.growthRulesStrictLocked(); err != nil {
		return GrowthProfile{}, err
	}
	action := "task:" + taskCode
	profile, _, err := s.applyExperienceMutationLocked(GrowthMutation{
		UserID: userID, ExperienceDelta: experience, Action: action,
		IdempotencyKey: "task_reward:" + strconv.FormatInt(userID, 10) + ":" + taskCode, CreatedAt: time.Now(),
	})
	return profile, err
}

func (s *Service) applyExperienceMutationLocked(mutation GrowthMutation) (GrowthProfile, bool, error) {
	if s.repo != nil {
		rules := s.growthRulesLocked()
		profile, footprint, applied, err := s.repo.ApplyGrowthMutation(context.Background(), mutation, rules.InitialLevel)
		if err != nil {
			return GrowthProfile{}, false, err
		}
		profile.Level = s.levelForExperienceLocked(profile.Experience, rules)
		s.growth[mutation.UserID] = profile
		if applied {
			s.footprints = append(s.footprints, footprint)
			if mutation.Event != nil {
				s.growthEvents[mutation.Event.IdempotencyKey] = *mutation.Event
			}
		}
		return profile, applied, nil
	}
	if _, exists := s.growthEvents[mutation.IdempotencyKey]; exists {
		return s.profileLocked(mutation.UserID), false, nil
	}
	profile := s.addExperienceOnlyLocked(mutation.UserID, mutation.ExperienceDelta, 0, mutation.GameID, mutation.Action)
	s.growthEvents[mutation.IdempotencyKey] = GrowthEvent{UserID: mutation.UserID, GameID: mutation.GameID, EventCode: mutation.Action, IdempotencyKey: mutation.IdempotencyKey, OccurredAt: mutation.CreatedAt}
	if mutation.Event != nil {
		s.growthEvents[mutation.Event.IdempotencyKey] = *mutation.Event
	}
	return profile, true, nil
}

func (s *Service) Todos(userID int64) ([]Todo, error) {
	if s.repo != nil {
		if err := s.repairPendingReviewReminders(userID); err != nil {
			return nil, err
		}
		return s.todosFromRepository(userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Todo, 0)
	for gameID, deadline := range s.reviewable {
		if !s.games.IsMember(gameID, userID) {
			continue
		}
		if time.Now().After(deadline) {
			continue
		}
		for _, targetID := range s.reviewTargets(gameID, userID) {
			targetRole := s.targetRole(gameID, targetID)
			if targetID == userID || s.reviewed[reviewKey(gameID, userID, targetID, targetRole)] {
				continue
			}
			result = append(result, Todo{
				GameID:       gameID,
				TargetUserID: targetID,
				TargetRole:   targetRole,
				DeadlineAt:   deadline.Format(time.RFC3339),
			})
		}
	}
	return result, nil
}

func (s *Service) repairPendingReviewReminders(userID int64) error {
	provider, ok := s.games.(GameListProvider)
	if !ok {
		return nil
	}
	existing, err := s.repo.ListReviewable(context.Background(), userID)
	if err != nil {
		return err
	}
	for _, game := range provider.List() {
		if game.Status != "pending_review" || !s.games.IsMember(game.ID, userID) {
			continue
		}
		if _, ok := existing[game.ID]; ok {
			continue
		}
		if err := s.MarkGameReviewableStrict(game.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Submit(userID int64, req SubmitRequest) (Review, GrowthProfile, error) {
	return s.submit(userID, req, 0)
}

// SubmitWithPoints lets the app API use the central points ledger as the only
// reward writer while keeping Submit's package-level behavior backwards compatible.
func (s *Service) SubmitWithPoints(userID int64, req SubmitRequest, rewardPoints int) (Review, GrowthProfile, error) {
	if rewardPoints < 0 {
		return Review{}, GrowthProfile{}, ErrInvalidReview
	}
	return s.submit(userID, req, rewardPoints)
}

func (s *Service) submit(userID int64, req SubmitRequest, rewardPoints int) (Review, GrowthProfile, error) {
	req.TargetRole = strings.TrimSpace(req.TargetRole)
	req.Content = strings.TrimSpace(req.Content)
	req.AgainIntent = strings.TrimSpace(req.AgainIntent)
	req.Tags = normalizeReviewTags(req.Tags)
	if req.Score < 1 || req.Score > 5 || req.GameID <= 0 || req.TargetUserID <= 0 || req.TargetUserID == userID {
		return Review{}, GrowthProfile{}, ErrInvalidReview
	}
	if len(req.Content) > 500 || !validTargetRole(req.TargetRole) || !validAgainIntent(req.AgainIntent) {
		return Review{}, GrowthProfile{}, ErrInvalidReview
	}
	if req.NPSScore != nil && (*req.NPSScore < 0 || *req.NPSScore > 10) {
		return Review{}, GrowthProfile{}, ErrInvalidReview
	}
	game, err := s.games.Get(req.GameID)
	if err != nil {
		return Review{}, GrowthProfile{}, err
	}
	if game.Status != "pending_review" && game.Status != "completed" {
		return Review{}, GrowthProfile{}, ErrGameNotReviewable
	}
	if !s.games.IsMember(req.GameID, userID) || !s.games.IsMember(req.GameID, req.TargetUserID) {
		return Review{}, GrowthProfile{}, ErrForbidden
	}
	allowed := false
	for _, targetID := range s.reviewTargets(req.GameID, userID) {
		if targetID == req.TargetUserID {
			allowed = true
			break
		}
	}
	if !allowed {
		return Review{}, GrowthProfile{}, ErrForbidden
	}
	req.TargetRole = s.targetRole(req.GameID, req.TargetUserID)

	s.mu.Lock()
	defer s.mu.Unlock()
	targetRole := defaultTargetRole(req.TargetRole)
	key := reviewKey(req.GameID, userID, req.TargetUserID, targetRole)
	if s.reviewed[key] {
		return Review{}, GrowthProfile{}, ErrDuplicateReview
	}
	if s.repo != nil {
		exists, err := s.repo.ReviewExists(context.Background(), req.GameID, userID, req.TargetUserID, targetRole)
		if err != nil {
			return Review{}, GrowthProfile{}, err
		}
		if exists {
			return Review{}, GrowthProfile{}, ErrDuplicateReview
		}
	}
	deadline, ok, err := s.reviewDeadlineLocked(userID, req.GameID)
	if err != nil {
		return Review{}, GrowthProfile{}, err
	}
	if !ok || time.Now().After(deadline) {
		return Review{}, GrowthProfile{}, ErrGameNotReviewable
	}

	review := Review{
		ID:             s.nextID,
		GameID:         req.GameID,
		ReviewerUserID: userID,
		TargetUserID:   req.TargetUserID,
		TargetRole:     targetRole,
		Score:          req.Score,
		Content:        req.Content,
		Tags:           append([]string(nil), req.Tags...),
		AgainIntent:    req.AgainIntent,
		NPSScore:       cloneInt(req.NPSScore),
		CreatedAt:      time.Now(),
	}
	rules, rulesErr := s.growthRulesStrictLocked()
	if rulesErr != nil {
		return Review{}, GrowthProfile{}, rulesErr
	}
	targetEventCode := "review_received"
	if review.Score >= 4 {
		targetEventCode = "positive_review_received"
	}
	if s.repo != nil {
		mutationKey := "review_growth:" + key
		reviewerEvent := GrowthEvent{
			UserID: userID, GameID: req.GameID, RoleCode: growthEventRole(s.targetRole(req.GameID, userID)), EventCode: "review_submitted",
			IdempotencyKey: "review_submitted:" + key, OccurredAt: review.CreatedAt,
		}
		targetEvent := GrowthEvent{
			UserID: req.TargetUserID, GameID: req.GameID, RoleCode: growthEventRole(targetRole), EventCode: targetEventCode,
			IdempotencyKey: targetEventCode + ":" + key, OccurredAt: review.CreatedAt,
		}
		achievements := []Achievement{
			{UserID: userID, Code: "first_review", Title: "首次评价", AchievedAt: review.CreatedAt},
			{UserID: req.TargetUserID, Code: "first_received_review", Title: "首次收到评价", AchievedAt: review.CreatedAt},
		}
		saved, reviewerProfile, targetProfile, footprints, savedAchievements, err := s.repo.SaveReviewWithGrowth(context.Background(), ReviewGrowthBundle{
			Review: review,
			ReviewerMutation: GrowthMutation{
				UserID: userID, GameID: req.GameID, ExperienceDelta: rules.SubmittedReviewExperience, ReviewCountDelta: 1,
				Action: "submitted_review", IdempotencyKey: mutationKey + ":reviewer", Event: &reviewerEvent, CreatedAt: review.CreatedAt,
			},
			TargetMutation: GrowthMutation{
				UserID: req.TargetUserID, GameID: req.GameID, ExperienceDelta: rules.ReceivedReviewExperience, ReviewCountDelta: 1,
				Action: "received_review", IdempotencyKey: mutationKey + ":target", Event: &targetEvent, CreatedAt: review.CreatedAt,
			},
			Achievements: achievements,
		}, rules.InitialLevel)
		if err != nil {
			return Review{}, GrowthProfile{}, err
		}
		review = saved
		reviewerProfile.Level = s.levelForExperienceLocked(reviewerProfile.Experience, rules)
		targetProfile.Level = s.levelForExperienceLocked(targetProfile.Experience, rules)
		reviewerProfile.CreditScore = s.creditScoreLocked(userID)
		reviewerProfile.TodayCreditScore = reviewerProfile.CreditScore
		targetProfile.CreditScore = s.creditScoreLocked(req.TargetUserID)
		targetProfile.TodayCreditScore = targetProfile.CreditScore
		s.growth[userID] = reviewerProfile
		s.growth[req.TargetUserID] = targetProfile
		s.footprints = append(s.footprints, footprints...)
		s.growthEvents[reviewerEvent.IdempotencyKey] = reviewerEvent
		s.growthEvents[targetEvent.IdempotencyKey] = targetEvent
		for _, achievement := range savedAchievements {
			exists := false
			for _, current := range s.achievements[achievement.UserID] {
				if current.Code == achievement.Code {
					exists = true
					break
				}
			}
			if !exists {
				s.achievements[achievement.UserID] = append(s.achievements[achievement.UserID], achievement)
			}
		}
		if review.ID >= s.nextID {
			s.nextID = review.ID + 1
		}
		s.reviews = append(s.reviews, review)
		s.reviewed[key] = true
		return review, reviewerProfile, nil
	}
	s.nextID++
	s.reviews = append(s.reviews, review)
	s.reviewed[key] = true
	profile := s.addGrowthLocked(userID, rules.SubmittedReviewExperience, rewardPoints, req.GameID, "submitted_review")
	s.addGrowthLocked(req.TargetUserID, rules.ReceivedReviewExperience, 0, req.GameID, "received_review")
	s.ensureAchievementLocked(userID, "first_review", "首次评价")
	s.ensureAchievementLocked(req.TargetUserID, "first_received_review", "首次收到评价")
	reviewKeyText := strconv.FormatInt(review.ID, 10)
	s.recordGrowthEventLocked(GrowthEvent{
		UserID: userID, GameID: req.GameID, RoleCode: growthEventRole(s.targetRole(req.GameID, userID)), EventCode: "review_submitted",
		IdempotencyKey: "review_submitted:" + reviewKeyText, OccurredAt: review.CreatedAt,
	})
	s.recordGrowthEventLocked(GrowthEvent{
		UserID: req.TargetUserID, GameID: req.GameID, RoleCode: growthEventRole(targetRole), EventCode: targetEventCode,
		IdempotencyKey: targetEventCode + ":" + reviewKeyText, OccurredAt: review.CreatedAt,
	})
	return review, profile, nil
}

func (s *Service) MyIntents(userID int64) []Review {
	items, _ := s.MyIntentsStrict(userID)
	return items
}

func (s *Service) MyIntentsStrict(userID int64) ([]Review, error) {
	if s.repo != nil {
		items, err := s.repo.ListReviewsByUser(context.Background(), userID)
		if err != nil {
			return nil, err
		}
		result := make([]Review, 0)
		for _, review := range items {
			if review.ReviewerUserID == userID && review.AgainIntent != "" {
				result = append(result, review)
			}
		}
		return result, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Review, 0)
	for _, review := range s.reviews {
		if review.ReviewerUserID == userID && review.AgainIntent != "" {
			result = append(result, review)
		}
	}
	return result, nil
}

func (s *Service) AllReviews() []Review {
	items, _ := s.AllReviewsStrict()
	return items
}

func (s *Service) AllReviewsStrict() ([]Review, error) {
	if s.repo != nil {
		return s.repo.ListReviews(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Review(nil), s.reviews...), nil
}

func (s *Service) Profile(userID int64) GrowthProfile {
	profile, err := s.ProfileStrict(userID)
	if err == nil {
		return profile
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profileWithCreditLocked(userID)
}

func (s *Service) ProfileStrict(userID int64) (GrowthProfile, error) {
	if s.repo != nil {
		profile, ok, err := s.repo.GetGrowthProfile(context.Background(), userID)
		if err != nil {
			return GrowthProfile{}, err
		}
		rules, rulesErr := s.growthRulesStrict()
		if rulesErr != nil {
			return GrowthProfile{}, rulesErr
		}
		if !ok {
			profile = defaultGrowthProfile(userID)
			profile.Level = s.levelForExperience(profile.Experience, rules)
			profile, err = s.repo.SaveGrowthProfile(context.Background(), profile)
			if err != nil {
				return GrowthProfile{}, err
			}
		}
		profile.Level = s.levelForExperience(profile.Experience, rules)
		profile.CreditScore, err = s.CreditScoreStrict(userID)
		if err != nil {
			return GrowthProfile{}, err
		}
		profile.TodayCreditScore = profile.CreditScore
		achievements, err := s.repo.ListAchievementsByUser(context.Background(), userID)
		if err != nil {
			return GrowthProfile{}, err
		}
		for _, item := range achievements {
			profile.Achievements = append(profile.Achievements, item.Code)
		}
		return profile, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profileWithCreditLocked(userID), nil
}

func (s *Service) Footprints(userID int64) []Footprint {
	items, _ := s.FootprintsStrict(userID)
	return items
}

func (s *Service) FootprintsStrict(userID int64) ([]Footprint, error) {
	if s.repo != nil {
		return s.repo.ListFootprintsByUser(context.Background(), userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Footprint, 0)
	for _, item := range s.footprints {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Service) AllFootprints() []Footprint {
	items, _ := s.AllFootprintsStrict()
	return items
}

func (s *Service) AllFootprintsStrict() ([]Footprint, error) {
	if s.repo != nil {
		return s.repo.ListFootprints(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Footprint(nil), s.footprints...), nil
}

func (s *Service) DeductCredit(userID int64, gameID int64, reason string) CreditLog {
	log, _ := s.DeductCreditStrict(userID, gameID, reason)
	return log
}

// DeductCreditStrict persists the permanent account and its ledger entry atomically.
// Callers that have already committed their business action may deliberately
// observe the returned error and mark the response as degraded for repair.
func (s *Service) DeductCreditStrict(userID int64, gameID int64, reason string) (CreditLog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	change, err := s.creditDeductionValueStrictLocked(reason)
	if err != nil {
		return CreditLog{}, err
	}
	return s.applyCreditChangeLocked(userID, gameID, reason, change, "credit_deducted")
}

// DeductCreditOnceStrict applies one business deduction at most once. It is
// used when the surrounding business operation may be retried after a partial
// failure, such as report handling.
func (s *Service) DeductCreditOnceStrict(userID int64, gameID int64, reason string, idempotencyKey string) (CreditLog, bool, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if userID <= 0 || idempotencyKey == "" {
		return CreditLog{}, false, ErrInvalidReview
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	change, err := s.creditDeductionValueStrictLocked(reason)
	if err != nil {
		return CreditLog{}, false, err
	}
	initial, cap, rulesErr := s.creditAccountRulesStrictLocked()
	if rulesErr != nil {
		return CreditLog{}, false, rulesErr
	}
	if s.repo != nil {
		log, applied, err := s.repo.ApplyCreditChangeOnce(context.Background(), userID, gameID, change, reason, idempotencyKey, initial, cap, time.Now())
		if err != nil {
			return CreditLog{}, false, err
		}
		if applied {
			s.syncCreditStateLocked(log)
			footprint := Footprint{UserID: userID, GameID: gameID, Action: "credit_deducted", CreatedAt: time.Now()}
			if saved, saveErr := s.repo.AddFootprint(context.Background(), footprint); saveErr == nil {
				footprint = saved
			}
			s.footprints = append(s.footprints, footprint)
		}
		return log, applied, nil
	}
	key := strconv.FormatInt(userID, 10) + ":" + idempotencyKey
	if existing, ok := s.creditMutationKeys[key]; ok {
		return existing, false, nil
	}
	before := s.creditScoreLocked(userID)
	after := before + change
	if after < 0 {
		after = 0
	}
	if after > cap {
		after = cap
	}
	log := CreditLog{ID: s.nextCreditID, UserID: userID, GameID: gameID, ChangeValue: after - before, BeforeScore: before, AfterScore: after, Reason: reason, CreatedAt: time.Now()}
	s.creditMutationKeys[key] = log
	s.syncCreditStateLocked(log)
	s.footprints = append(s.footprints, Footprint{UserID: userID, GameID: gameID, Action: "credit_deducted", CreatedAt: time.Now()})
	return log, true, nil
}

func (s *Service) applyCreditChangeLocked(userID int64, gameID int64, reason string, change int, footprintAction string) (CreditLog, error) {
	initial, cap, rulesErr := s.creditAccountRulesStrictLocked()
	if rulesErr != nil {
		return CreditLog{}, rulesErr
	}
	if s.repo != nil {
		log, err := s.repo.ApplyCreditChange(context.Background(), userID, gameID, change, reason, initial, cap, time.Now())
		if err != nil {
			return CreditLog{}, err
		}
		s.syncCreditStateLocked(log)
		footprint := Footprint{UserID: userID, GameID: gameID, Action: footprintAction, CreatedAt: time.Now()}
		if saved, err := s.repo.AddFootprint(context.Background(), footprint); err == nil {
			footprint = saved
		}
		s.footprints = append(s.footprints, footprint)
		return log, nil
	}
	before := s.creditScoreLocked(userID)
	after := before + change
	if after < 0 {
		after = 0
	}
	if after > cap {
		after = cap
	}
	log := CreditLog{ID: s.nextCreditID, UserID: userID, GameID: gameID, ChangeValue: after - before, BeforeScore: before, AfterScore: after, Reason: reason, CreatedAt: time.Now()}
	s.syncCreditStateLocked(log)
	s.footprints = append(s.footprints, Footprint{UserID: userID, GameID: gameID, Action: footprintAction, CreatedAt: time.Now()})
	return log, nil
}

func (s *Service) syncCreditStateLocked(log CreditLog) {
	s.creditScores[log.UserID] = log.AfterScore
	if log.ID >= s.nextCreditID {
		s.nextCreditID = log.ID + 1
	}
	s.creditLogs = append(s.creditLogs, log)
	profile := s.profileLocked(log.UserID)
	profile.CreditScore = log.AfterScore
	profile.TodayCreditScore = log.AfterScore
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	s.growth[log.UserID] = profile
}

func (s *Service) RestoreCredit(userID int64, gameID int64, reason string, amount int) CreditLog {
	log, _ := s.RestoreCreditStrict(userID, gameID, reason, amount)
	return log
}

func (s *Service) RestoreCreditStrict(userID int64, gameID int64, reason string, amount int) (CreditLog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if amount < 0 {
		amount = -amount
	}
	return s.applyCreditChangeLocked(userID, gameID, reason, amount, "credit_restored")
}

// RestoreCreditForAppeal restores one deduction at most once for a concrete
// appeal. The source deduction and the restore ledger are linked by appeal ID,
// so repeated or concurrent approvals cannot increase the user's credit twice.
func (s *Service) RestoreCreditForAppeal(userID int64, gameID int64, sourceCreditLogID int64, appealID int64, amount int) (CreditLog, bool, error) {
	if userID <= 0 || sourceCreditLogID <= 0 || appealID <= 0 || amount == 0 {
		return CreditLog{}, false, ErrInvalidReview
	}
	if amount < 0 {
		amount = -amount
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	initialScore, scoreCap, rulesErr := s.creditAccountRulesStrictLocked()
	if rulesErr != nil {
		return CreditLog{}, false, rulesErr
	}
	if s.repo != nil {
		log, applied, err := s.repo.RestoreCreditForAppeal(
			context.Background(), userID, gameID, sourceCreditLogID, appealID,
			amount, initialScore, scoreCap, time.Now(),
		)
		if err != nil {
			return CreditLog{}, false, err
		}
		if !applied {
			return log, false, nil
		}
		s.creditScores[userID] = log.AfterScore
		s.nextCreditID++
		s.creditLogs = append(s.creditLogs, log)
		profile := s.profileLocked(userID)
		profile.CreditScore = log.AfterScore
		profile.TodayCreditScore = log.AfterScore
		profile.UpdatedAt = time.Now().Format(time.RFC3339)
		s.growth[userID] = profile
		footprint := Footprint{UserID: userID, GameID: gameID, Action: "credit_restored", CreatedAt: time.Now()}
		if saved, err := s.repo.AddFootprint(context.Background(), footprint); err == nil {
			footprint = saved
		}
		s.footprints = append(s.footprints, footprint)
		return log, true, nil
	}

	for i := range s.creditLogs {
		source := &s.creditLogs[i]
		if source.ID != sourceCreditLogID || source.UserID != userID || source.ChangeValue >= 0 {
			continue
		}
		if source.AppealID != 0 && source.AppealID != appealID {
			return CreditLog{}, false, ErrCreditAppealConflict
		}
		for _, existing := range s.creditLogs {
			if existing.AppealID == appealID && existing.ChangeValue >= 0 && existing.Reason == "appeal_passed" {
				return existing, false, nil
			}
		}
		source.AppealID = appealID
		if maximum := -source.ChangeValue; amount > maximum {
			amount = maximum
		}
		before := s.creditScoreLocked(userID)
		after := before + amount
		if after > scoreCap {
			after = scoreCap
		}
		log := CreditLog{
			ID:          s.nextCreditID,
			UserID:      userID,
			GameID:      gameID,
			ChangeValue: after - before,
			BeforeScore: before,
			AfterScore:  after,
			Reason:      "appeal_passed",
			AppealID:    appealID,
			CreatedAt:   time.Now(),
		}
		s.nextCreditID++
		s.creditScores[userID] = after
		s.creditLogs = append(s.creditLogs, log)
		profile := s.profileLocked(userID)
		profile.CreditScore = after
		profile.TodayCreditScore = after
		profile.UpdatedAt = time.Now().Format(time.RFC3339)
		s.growth[userID] = profile
		footprint := Footprint{UserID: userID, GameID: gameID, Action: "credit_restored", CreatedAt: time.Now()}
		s.footprints = append(s.footprints, footprint)
		return log, true, nil
	}
	return CreditLog{}, false, ErrCreditLogNotFound
}

// CreditDeductionValue exposes the configured deduction without mutating any
// state. The exit coordinator uses it to pass the rule into the single SQL
// transaction that also updates membership and game capacity.
func (s *Service) CreditDeductionValue(reason string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.creditDeductionValueLocked(reason)
}

func (s *Service) CreditDeductionValueStrict(reason string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.creditDeductionValueStrictLocked(reason)
}

func (s *Service) creditDeductionValueLocked(reason string) int {
	change, err := s.creditDeductionValueStrictLocked(reason)
	if err == nil {
		return change
	}
	return defaultCreditDeductionValue(reason)
}

func (s *Service) creditDeductionValueStrictLocked(reason string) (int, error) {
	if s.repo != nil {
		change, ok, err := s.repo.CreditDeductionValue(context.Background(), reason)
		if err != nil {
			return 0, err
		}
		if ok {
			return change, nil
		}
	}
	return defaultCreditDeductionValue(reason), nil
}

func defaultCreditDeductionValue(reason string) int {
	switch reason {
	case "player_cancel_service":
		return -3
	case "expert_cancel_service":
		return -5
	}
	return -10
}

// CreditScoreStrict is used by authorization checks. A repository failure must
// not be converted into a permissive default score.
func (s *Service) CreditScoreStrict(userID int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.repo != nil {
		initial, _, rulesErr := s.creditAccountRulesStrictLocked()
		if rulesErr != nil {
			return 0, rulesErr
		}
		account, err := s.repo.EnsureCreditAccount(context.Background(), userID, initial)
		if err != nil {
			return 0, err
		}
		s.creditScores[userID] = account.CurrentScore
		return account.CurrentScore, nil
	}
	return s.creditScoreLocked(userID), nil
}

func (s *Service) TraceByUser(userID int64) Trace {
	trace, _ := s.TraceByUserStrict(userID)
	return trace
}

func (s *Service) TraceByUserStrict(userID int64) (Trace, error) {
	if s.repo != nil {
		profile, err := s.ProfileStrict(userID)
		if err != nil {
			return Trace{}, err
		}
		trace := Trace{UserID: userID, Profile: profile}
		trace.Reviews, err = s.repo.ListReviewsByUser(context.Background(), userID)
		if err != nil {
			return Trace{}, err
		}
		trace.CreditLogs, err = s.repo.ListCreditLogsByUser(context.Background(), userID)
		if err != nil {
			return Trace{}, err
		}
		trace.Footprints, err = s.repo.ListFootprintsByUser(context.Background(), userID)
		if err != nil {
			return Trace{}, err
		}
		trace.Achievements, err = s.repo.ListAchievementsByUser(context.Background(), userID)
		if err != nil {
			return Trace{}, err
		}
		return trace, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	trace := Trace{UserID: userID, Profile: s.profileWithCreditLocked(userID)}
	for _, review := range s.reviews {
		if review.ReviewerUserID == userID || review.TargetUserID == userID {
			trace.Reviews = append(trace.Reviews, review)
		}
	}
	for _, log := range s.creditLogs {
		if log.UserID == userID {
			trace.CreditLogs = append(trace.CreditLogs, log)
		}
	}
	for _, item := range s.footprints {
		if item.UserID == userID {
			trace.Footprints = append(trace.Footprints, item)
		}
	}
	trace.Achievements = append(trace.Achievements, s.achievements[userID]...)
	return trace, nil
}

func (s *Service) TraceByGame(gameID int64) Trace {
	trace, _ := s.TraceByGameStrict(gameID)
	return trace
}

func (s *Service) TraceByGameStrict(gameID int64) (Trace, error) {
	if s.repo != nil {
		trace := Trace{GameID: gameID}
		var err error
		trace.Reviews, err = s.repo.ListReviewsByGame(context.Background(), gameID)
		if err != nil {
			return Trace{}, err
		}
		trace.CreditLogs, err = s.repo.ListCreditLogsByGame(context.Background(), gameID)
		if err != nil {
			return Trace{}, err
		}
		footprints, err := s.repo.ListFootprints(context.Background())
		if err != nil {
			return Trace{}, err
		}
		for _, item := range footprints {
			if item.GameID == gameID {
				trace.Footprints = append(trace.Footprints, item)
			}
		}
		trace.Achievements, err = s.repo.ListAchievements(context.Background())
		if err != nil {
			return Trace{}, err
		}
		return trace, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	trace := Trace{GameID: gameID}
	for _, review := range s.reviews {
		if review.GameID == gameID {
			trace.Reviews = append(trace.Reviews, review)
		}
	}
	for _, log := range s.creditLogs {
		if log.GameID == gameID {
			trace.CreditLogs = append(trace.CreditLogs, log)
		}
	}
	for _, item := range s.footprints {
		if item.GameID == gameID {
			trace.Footprints = append(trace.Footprints, item)
		}
	}
	for _, items := range s.achievements {
		for _, item := range items {
			trace.Achievements = append(trace.Achievements, item)
		}
	}
	return trace, nil
}

func (s *Service) GameReviewComplete(gameID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.repo != nil {
		return s.gameReviewCompleteFromRepository(gameID)
	}
	if _, ok := s.reviewable[gameID]; !ok {
		return false
	}
	members := s.games.Members(gameID)
	requiredCount := 0
	for _, reviewerID := range members {
		for _, targetID := range s.reviewTargets(gameID, reviewerID) {
			requiredCount++
			if !s.reviewed[reviewKey(gameID, reviewerID, targetID, s.targetRole(gameID, targetID))] {
				return false
			}
		}
	}
	return requiredCount > 0
}

func (s *Service) gameReviewCompleteFromRepository(gameID int64) bool {
	members := s.games.Members(gameID)
	if len(members) <= 1 {
		return false
	}
	reviewable := false
	requiredCount := 0
	for _, reviewerID := range members {
		reviewableGames, err := s.repo.ListReviewable(context.Background(), reviewerID)
		if err != nil {
			return false
		}
		if _, ok := reviewableGames[gameID]; ok {
			reviewable = true
		}
		for _, targetID := range s.reviewTargets(gameID, reviewerID) {
			requiredCount++
			exists, err := s.repo.ReviewExists(context.Background(), gameID, reviewerID, targetID, s.targetRole(gameID, targetID))
			if err != nil || !exists {
				return false
			}
		}
	}
	return reviewable && requiredCount > 0
}

func (s *Service) addGrowthLocked(userID int64, exp int, points int, gameID int64, action string) GrowthProfile {
	profile := s.profileLocked(userID)
	profile.Experience += exp
	profile.AvailablePoints += points
	profile.ReviewCount++
	rules := s.growthRulesLocked()
	profile.Level = s.levelForExperienceLocked(profile.Experience, rules)
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	if s.repo != nil {
		if saved, err := s.repo.SaveGrowthProfile(context.Background(), profile); err == nil {
			profile = saved
		}
		_ = s.repo.AddExperienceLog(context.Background(), userID, gameID, exp, action, time.Now())
		if points != 0 {
			_ = s.repo.AddPointsLog(context.Background(), userID, gameID, points, action, time.Now())
		}
	}
	s.growth[userID] = profile
	footprint := Footprint{
		UserID:    userID,
		GameID:    gameID,
		Action:    action,
		CreatedAt: time.Now(),
	}
	if s.repo != nil {
		if saved, err := s.repo.AddFootprint(context.Background(), footprint); err == nil {
			footprint = saved
		}
	}
	s.footprints = append(s.footprints, footprint)
	return profile
}

func (s *Service) addExperienceOnlyLocked(userID int64, exp int, points int, gameID int64, action string) GrowthProfile {
	profile := s.profileLocked(userID)
	profile.Experience += exp
	profile.AvailablePoints += points
	rules := s.growthRulesLocked()
	profile.Level = s.levelForExperienceLocked(profile.Experience, rules)
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	if s.repo != nil {
		if saved, err := s.repo.SaveGrowthProfile(context.Background(), profile); err == nil {
			profile = saved
		}
		_ = s.repo.AddExperienceLog(context.Background(), userID, gameID, exp, action, time.Now())
		if points != 0 {
			_ = s.repo.AddPointsLog(context.Background(), userID, gameID, points, action, time.Now())
		}
	}
	s.growth[userID] = profile
	footprint := Footprint{
		UserID:    userID,
		GameID:    gameID,
		Action:    action,
		CreatedAt: time.Now(),
	}
	if s.repo != nil {
		if saved, err := s.repo.AddFootprint(context.Background(), footprint); err == nil {
			footprint = saved
		}
	}
	s.footprints = append(s.footprints, footprint)
	return profile
}

func (s *Service) hasFootprintLocked(userID int64, gameID int64, action string) bool {
	if s.repo != nil {
		if items, err := s.repo.ListFootprintsByUser(context.Background(), userID); err == nil {
			for _, item := range items {
				if item.GameID == gameID && item.Action == action {
					return true
				}
			}
		}
	}
	for _, item := range s.footprints {
		if item.UserID == userID && item.GameID == gameID && item.Action == action {
			return true
		}
	}
	return false
}

func (s *Service) profileLocked(userID int64) GrowthProfile {
	if s.repo != nil {
		if profile, ok, err := s.repo.GetGrowthProfile(context.Background(), userID); err == nil && ok {
			return profile
		}
	}
	profile, ok := s.growth[userID]
	if !ok {
		profile = defaultGrowthProfile(userID)
		rules := s.growthRulesLocked()
		profile.Level = rules.InitialLevel
		profile.CreditScore = rules.InitialCreditScore
		profile.TodayCreditScore = rules.InitialCreditScore
	}
	return profile
}

func defaultGrowthProfile(userID int64) GrowthProfile {
	rules := DefaultGrowthRules()
	return GrowthProfile{
		UserID:           userID,
		Level:            rules.InitialLevel,
		CreditScore:      rules.InitialCreditScore,
		TodayCreditScore: rules.InitialCreditScore,
		UpdatedAt:        time.Now().Format(time.RFC3339),
	}
}

// growthRulesLocked is used by mutation paths that already hold s.mu. It
// avoids taking a nested lock while still reading the latest admin config.
func (s *Service) growthRulesLocked() GrowthRules {
	rules := DefaultGrowthRules()
	if s.rulesProvider != nil {
		var stored GrowthRules
		if s.rulesProvider.Get(growthRulesConfigKey, &stored) {
			rules = normalizeGrowthRules(stored)
		}
	}
	return rules
}

func (s *Service) growthRulesStrictLocked() (GrowthRules, error) {
	if s.rulesStrictProvider != nil {
		var stored GrowthRules
		found, err := s.rulesStrictProvider.GetStrict(growthRulesConfigKey, &stored)
		if err != nil {
			return GrowthRules{}, err
		}
		if found {
			return normalizeGrowthRules(stored), nil
		}
	}
	return s.growthRulesLocked(), nil
}

// levelForExperience resolves a player level from the application-owned
// ladder. Older callers that do not inject a ladder keep the historical
// experience-per-level calculation, so package-level consumers remain
// compatible while the HTTP application can use its configured thresholds.
func (s *Service) levelForExperience(experience int, rules GrowthRules) int {
	if experience < 0 {
		experience = 0
	}
	s.mu.RLock()
	resolver := s.levelResolver
	s.mu.RUnlock()
	if resolver != nil {
		return resolver(experience)
	}
	return experience/rules.ExperiencePerLevel + rules.InitialLevel
}

// levelForExperienceLocked is the equivalent helper for mutation paths that
// already hold s.mu. It must not take a nested read lock.
func (s *Service) levelForExperienceLocked(experience int, rules GrowthRules) int {
	if experience < 0 {
		experience = 0
	}
	if s.levelResolver != nil {
		return s.levelResolver(experience)
	}
	return experience/rules.ExperiencePerLevel + rules.InitialLevel
}

func (s *Service) profileWithCreditLocked(userID int64) GrowthProfile {
	profile := s.profileLocked(userID)
	rules := s.growthRulesLocked()
	profile.Level = s.levelForExperienceLocked(profile.Experience, rules)
	profile.CreditScore = s.creditScoreLocked(userID)
	profile.TodayCreditScore = profile.CreditScore
	for _, item := range s.achievements[userID] {
		profile.Achievements = append(profile.Achievements, item.Code)
	}
	return profile
}

func (s *Service) creditScoreLocked(userID int64) int {
	if s.repo != nil {
		initial, _ := s.creditAccountRulesLocked()
		if account, err := s.repo.EnsureCreditAccount(context.Background(), userID, initial); err == nil {
			s.creditScores[userID] = account.CurrentScore
			return account.CurrentScore
		}
	}
	score, ok := s.creditScores[userID]
	if !ok {
		score, _ = s.creditAccountRulesLocked()
		s.creditScores[userID] = score
	}
	return score
}

func (s *Service) ensureAchievementLocked(userID int64, code string, title string) {
	for _, item := range s.achievements[userID] {
		if item.Code == code {
			return
		}
	}
	achievement := Achievement{
		UserID:     userID,
		Code:       code,
		Title:      title,
		AchievedAt: time.Now(),
	}
	if s.repo != nil {
		saved, created, err := s.repo.EnsureAchievement(context.Background(), achievement)
		if err != nil || !created {
			return
		}
		achievement = saved
	}
	s.achievements[userID] = append(s.achievements[userID], achievement)
}

func (s *Service) todosFromRepository(userID int64) ([]Todo, error) {
	reviewable, err := s.repo.ListReviewable(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	result := make([]Todo, 0)
	for gameID, deadline := range reviewable {
		if time.Now().After(deadline) {
			continue
		}
		if !s.games.IsMember(gameID, userID) {
			continue
		}
		for _, targetID := range s.reviewTargets(gameID, userID) {
			targetRole := s.targetRole(gameID, targetID)
			exists, err := s.repo.ReviewExists(context.Background(), gameID, userID, targetID, targetRole)
			if err != nil {
				return nil, err
			}
			if exists {
				continue
			}
			result = append(result, Todo{
				GameID:       gameID,
				TargetUserID: targetID,
				TargetRole:   targetRole,
				DeadlineAt:   deadline.Format(time.RFC3339),
			})
		}
	}
	return result, nil
}

func (s *Service) reviewTargets(gameID int64, reviewerID int64) []int64 {
	members := s.games.Members(gameID)
	memberSet := make(map[int64]bool, len(members))
	for _, memberID := range members {
		memberSet[memberID] = true
	}
	game, _ := s.games.Get(gameID)
	organizerID := game.CreatorUserID
	if reviewerID == organizerID {
		return reviewTargetsFallback(members, reviewerID)
	}
	provider, ok := s.games.(GameInvitationProvider)
	if !ok {
		return reviewTargetsFallback(members, reviewerID)
	}

	pairedGame := false
	boundTargets := make([]int64, 0, 1)
	seen := make(map[int64]bool)
	appendTarget := func(targetID int64) {
		if targetID > 0 && targetID != reviewerID && memberSet[targetID] && !seen[targetID] {
			seen[targetID] = true
			boundTargets = append(boundTargets, targetID)
		}
	}
	for _, memberID := range members {
		for _, invitation := range provider.InvitationsForUser(memberID) {
			if invitation.GameID != gameID || invitation.Status != "accepted" || invitation.PlayerUserID <= 0 || invitation.ExpertUserID <= 0 {
				continue
			}
			pairedGame = true
			var targetID int64
			switch reviewerID {
			case invitation.PlayerUserID:
				targetID = invitation.ExpertUserID
			case invitation.ExpertUserID:
				targetID = invitation.PlayerUserID
			}
			appendTarget(targetID)
		}
	}
	role := s.targetRole(gameID, reviewerID)
	if len(boundTargets) > 0 {
		if role == "member" || role == "player" || role == "expert" {
			for _, memberID := range members {
				memberRole := s.targetRole(gameID, memberID)
				if memberRole == "guide" || memberRole == "main_guide" {
					appendTarget(memberID)
				}
			}
			appendTarget(organizerID)
		}
		return boundTargets
	}
	if pairedGame {
		if (role == "member" || role == "player") && organizerID > 0 && organizerID != reviewerID && memberSet[organizerID] {
			return []int64{organizerID}
		}
		if role == "guide" || role == "main_guide" {
			return reviewTargetsFallback(members, reviewerID)
		}
		if role == "expert" {
			for _, memberID := range members {
				memberRole := s.targetRole(gameID, memberID)
				if memberRole == "guide" || memberRole == "main_guide" {
					appendTarget(memberID)
				}
			}
			appendTarget(organizerID)
			return boundTargets
		}
		return nil
	}
	return reviewTargetsFallback(members, reviewerID)
}

func reviewTargetsFallback(members []int64, reviewerID int64) []int64 {
	result := make([]int64, 0, len(members))
	for _, targetID := range members {
		if targetID != reviewerID {
			result = append(result, targetID)
		}
	}
	return result
}

func (s *Service) reviewDeadlineLocked(userID int64, gameID int64) (time.Time, bool, error) {
	if s.repo != nil {
		reviewable, err := s.repo.ListReviewable(context.Background(), userID)
		if err != nil {
			return time.Time{}, false, err
		}
		deadline, ok := reviewable[gameID]
		return deadline, ok, nil
	}
	deadline, ok := s.reviewable[gameID]
	return deadline, ok, nil
}

func (s *Service) creditScoreFromRepository(userID int64) int {
	initial, _ := s.creditAccountRules()
	account, err := s.repo.EnsureCreditAccount(context.Background(), userID, initial)
	if err != nil {
		return initial
	}
	return account.CurrentScore
}

func (s *Service) creditAccountRules() (int, int) {
	s.mu.RLock()
	provider := s.creditRulesProvider
	s.mu.RUnlock()
	if provider != nil {
		initial, cap := provider()
		if initial >= 0 && cap >= initial {
			return initial, cap
		}
	}
	rules := s.GrowthRules()
	return rules.InitialCreditScore, rules.CreditScoreCap
}

func (s *Service) creditAccountRulesLocked() (int, int) {
	if s.creditRulesProvider != nil {
		initial, cap := s.creditRulesProvider()
		if initial >= 0 && cap >= initial {
			return initial, cap
		}
	}
	rules := s.growthRulesLocked()
	return rules.InitialCreditScore, rules.CreditScoreCap
}

func (s *Service) creditAccountRulesStrictLocked() (int, int, error) {
	if s.creditRulesStrictProvider != nil {
		initial, cap, err := s.creditRulesStrictProvider()
		if err != nil {
			return 0, 0, err
		}
		if initial >= 0 && cap >= initial {
			return initial, cap, nil
		}
		return 0, 0, ErrCreditRulesUnavailable
	}
	initial, cap := s.creditAccountRulesLocked()
	return initial, cap, nil
}

func (s *Service) creditScoreCapLocked() int {
	_, cap := s.creditAccountRulesLocked()
	return cap
}

func reviewKey(gameID int64, reviewerID int64, targetID int64, targetRole string) string {
	return strconv.FormatInt(gameID, 10) + ":" + strconv.FormatInt(reviewerID, 10) + ":" + strconv.FormatInt(targetID, 10) + ":" + targetRole
}

func defaultTargetRole(value string) string {
	if value == "" {
		return "member"
	}
	return value
}

func validTargetRole(value string) bool {
	switch value {
	case "", "member", "player", "expert", "guide", "main_guide":
		return true
	default:
		return false
	}
}

func (s *Service) targetRole(gameID int64, targetUserID int64) string {
	if provider, ok := s.games.(GameRoleProvider); ok {
		for _, item := range provider.MemberRoles(gameID) {
			if item.UserID == targetUserID {
				return normalizeTargetRole(item.Role)
			}
		}
	}
	game, err := s.games.Get(gameID)
	if err == nil && game.MainGuideUserID == targetUserID {
		return "main_guide"
	}
	return "member"
}

func normalizeTargetRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "expert", "master":
		return "expert"
	case "guide", "leader":
		return "guide"
	case "main_guide":
		return "main_guide"
	default:
		return "member"
	}
}

func validAgainIntent(value string) bool {
	switch value {
	case "", "yes", "no", "maybe":
		return true
	default:
		return false
	}
}

func normalizeReviewTags(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > 32 || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
		if len(result) >= 8 {
			break
		}
	}
	return result
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
