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
	ErrGameNotReviewable = errors.New("game not reviewable")
	ErrForbidden         = errors.New("forbidden")
	ErrDuplicateReview   = errors.New("duplicate review")
	ErrInvalidReview     = errors.New("invalid review")
)

// SubmittedReviewPoints is kept for backwards compatibility with callers that
// still use the historical default. Runtime rewards are read from GrowthRules.
const SubmittedReviewPoints = 2

const growthRulesConfigKey = "growth.reward_rules"

// GrowthRules controls the rewards and level calculation used by the growth
// ledger. The values are persisted in system configuration and can be changed
// by an administrator without rebuilding the service.
type GrowthRules struct {
	CompletedGameExperience   int `json:"completedGameExperience"`
	CompletedGamePoints       int `json:"completedGamePoints"`
	SubmittedReviewExperience int `json:"submittedReviewExperience"`
	ReceivedReviewExperience  int `json:"receivedReviewExperience"`
	SubmittedReviewPoints     int `json:"submittedReviewPoints"`
	ExperiencePerLevel        int `json:"experiencePerLevel"`
	InitialLevel              int `json:"initialLevel"`
	InitialCreditScore        int `json:"initialCreditScore"`
	CreditScoreCap            int `json:"creditScoreCap"`
}

type GrowthRulesProvider interface {
	Get(key string, target interface{}) bool
}

type GameProvider interface {
	Get(id int64) (games.Game, error)
	Members(gameID int64) []int64
	IsMember(gameID int64, userID int64) bool
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
	CreatedAt   time.Time `json:"createdAt"`
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
	ListReviewable(ctx context.Context, userID int64) (map[int64]time.Time, error)
	ReviewExists(ctx context.Context, gameID int64, reviewerID int64, targetID int64, targetRole string) (bool, error)
	SaveReview(ctx context.Context, review Review) (Review, error)
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
	GetTodayCredit(ctx context.Context, userID int64, now time.Time) (int, error)
	SaveTodayCredit(ctx context.Context, userID int64, now time.Time, score int) error
	CreditDeductionValue(ctx context.Context, ruleCode string) (int, bool, error)
	ListCreditDeductionRules(ctx context.Context) ([]CreditDeductionRule, error)
	UpsertCreditDeductionRule(ctx context.Context, rule CreditDeductionRule) (CreditDeductionRule, error)
	AddCreditLog(ctx context.Context, log CreditLog) (CreditLog, error)
	ListCreditLogsByUser(ctx context.Context, userID int64) ([]CreditLog, error)
	ListCreditLogsByGame(ctx context.Context, gameID int64) ([]CreditLog, error)
	EnsureAchievement(ctx context.Context, achievement Achievement) (Achievement, bool, error)
	ListAchievementsByUser(ctx context.Context, userID int64) ([]Achievement, error)
	ListAchievements(ctx context.Context) ([]Achievement, error)
}

type Service struct {
	mu            sync.RWMutex
	nextID        int64
	nextCreditID  int64
	games         GameProvider
	reviews       []Review
	reviewed      map[string]bool
	reviewable    map[int64]time.Time
	growth        map[int64]GrowthProfile
	dailyCredit   map[string]int
	creditLogs    []CreditLog
	achievements  map[int64][]Achievement
	footprints    []Footprint
	repo          Repository
	rulesProvider GrowthRulesProvider
}

func NewService(games GameProvider) *Service {
	return &Service{
		nextID:       1,
		nextCreditID: 1,
		games:        games,
		reviews:      make([]Review, 0),
		reviewed:     make(map[string]bool),
		reviewable:   make(map[int64]time.Time),
		growth:       make(map[int64]GrowthProfile),
		dailyCredit:  make(map[string]int),
		creditLogs:   make([]CreditLog, 0),
		achievements: make(map[int64][]Achievement),
		footprints:   make([]Footprint, 0),
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

func DefaultGrowthRules() GrowthRules {
	return GrowthRules{
		CompletedGameExperience:   10,
		CompletedGamePoints:       0,
		SubmittedReviewExperience: 5,
		ReceivedReviewExperience:  10,
		SubmittedReviewPoints:     SubmittedReviewPoints,
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

func normalizeGrowthRules(rules GrowthRules) GrowthRules {
	defaults := DefaultGrowthRules()
	if rules.CompletedGameExperience <= 0 {
		rules.CompletedGameExperience = defaults.CompletedGameExperience
	}
	if rules.CompletedGamePoints < 0 {
		rules.CompletedGamePoints = defaults.CompletedGamePoints
	}
	if rules.SubmittedReviewExperience <= 0 {
		rules.SubmittedReviewExperience = defaults.SubmittedReviewExperience
	}
	if rules.ReceivedReviewExperience <= 0 {
		rules.ReceivedReviewExperience = defaults.ReceivedReviewExperience
	}
	if rules.SubmittedReviewPoints < 0 {
		rules.SubmittedReviewPoints = defaults.SubmittedReviewPoints
	}
	if rules.ExperiencePerLevel <= 0 {
		rules.ExperiencePerLevel = defaults.ExperiencePerLevel
	}
	if rules.InitialLevel <= 0 {
		rules.InitialLevel = defaults.InitialLevel
	}
	if rules.InitialCreditScore <= 0 {
		rules.InitialCreditScore = defaults.InitialCreditScore
	}
	if rules.CreditScoreCap <= 0 {
		rules.CreditScoreCap = defaults.CreditScoreCap
	}
	if rules.CreditScoreCap < rules.InitialCreditScore {
		rules.CreditScoreCap = rules.InitialCreditScore
	}
	return rules
}

func (s *Service) MarkGameReviewable(gameID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	deadline := time.Now().Add(7 * 24 * time.Hour)
	if _, ok := s.reviewable[gameID]; ok {
		return
	}
	s.reviewable[gameID] = deadline
	if s.repo != nil {
		for _, userID := range s.games.Members(gameID) {
			_ = s.repo.MarkReviewable(context.Background(), gameID, userID, deadline)
		}
	}
}

func (s *Service) AwardCompletedGame(gameID int64) []GrowthProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles := make([]GrowthProfile, 0)
	for _, userID := range s.games.Members(gameID) {
		if s.hasFootprintLocked(userID, gameID, "completed_game") {
			continue
		}
		rules := s.growthRulesLocked()
		profiles = append(profiles, s.addExperienceOnlyLocked(userID, rules.CompletedGameExperience, rules.CompletedGamePoints, gameID, "completed_game"))
	}
	return profiles
}

func (s *Service) Todos(userID int64) ([]Todo, error) {
	if s.repo != nil {
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

func (s *Service) Submit(userID int64, req SubmitRequest) (Review, GrowthProfile, error) {
	return s.submit(userID, req, s.GrowthRules().SubmittedReviewPoints)
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
	if s.repo != nil {
		saved, err := s.repo.SaveReview(context.Background(), review)
		if err != nil {
			return Review{}, GrowthProfile{}, err
		}
		review = saved
	}
	s.nextID++
	s.reviews = append(s.reviews, review)
	s.reviewed[key] = true
	rules := s.growthRulesLocked()
	profile := s.addGrowthLocked(userID, rules.SubmittedReviewExperience, rewardPoints, req.GameID, "submitted_review")
	s.addGrowthLocked(req.TargetUserID, rules.ReceivedReviewExperience, 0, req.GameID, "received_review")
	s.ensureAchievementLocked(userID, "first_review", "首次评价")
	s.ensureAchievementLocked(req.TargetUserID, "first_received_review", "首次收到评价")
	return review, profile, nil
}

func (s *Service) MyIntents(userID int64) []Review {
	if s.repo != nil {
		if items, err := s.repo.ListReviewsByUser(context.Background(), userID); err == nil {
			result := make([]Review, 0)
			for _, review := range items {
				if review.ReviewerUserID == userID && review.AgainIntent != "" {
					result = append(result, review)
				}
			}
			return result
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Review, 0)
	for _, review := range s.reviews {
		if review.ReviewerUserID == userID && review.AgainIntent != "" {
			result = append(result, review)
		}
	}
	return result
}

func (s *Service) AllReviews() []Review {
	if s.repo != nil {
		if items, err := s.repo.ListReviews(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Review(nil), s.reviews...)
}

func (s *Service) Profile(userID int64) GrowthProfile {
	if s.repo != nil {
		if profile, ok, err := s.repo.GetGrowthProfile(context.Background(), userID); err == nil && ok {
			rules := s.GrowthRules()
			profile.Level = profile.Experience/rules.ExperiencePerLevel + rules.InitialLevel
			profile.TodayCreditScore = s.todayCreditFromRepository(userID)
			profile.CreditScore = profile.TodayCreditScore
			if achievements, err := s.repo.ListAchievementsByUser(context.Background(), userID); err == nil {
				for _, item := range achievements {
					profile.Achievements = append(profile.Achievements, item.Code)
				}
			}
			return profile
		}
		profile := defaultGrowthProfile(userID)
		rules := s.GrowthRules()
		profile.Level = profile.Experience/rules.ExperiencePerLevel + rules.InitialLevel
		profile.TodayCreditScore = s.todayCreditFromRepository(userID)
		profile.CreditScore = profile.TodayCreditScore
		if saved, err := s.repo.SaveGrowthProfile(context.Background(), profile); err == nil {
			return saved
		}
		return profile
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profileWithCreditLocked(userID)
}

func (s *Service) Footprints(userID int64) []Footprint {
	if s.repo != nil {
		if items, err := s.repo.ListFootprintsByUser(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Footprint, 0)
	for _, item := range s.footprints {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) AllFootprints() []Footprint {
	if s.repo != nil {
		if items, err := s.repo.ListFootprints(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Footprint(nil), s.footprints...)
}

func (s *Service) DeductCredit(userID int64, gameID int64, reason string) CreditLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	before := s.todayCreditLocked(userID)
	change := s.creditDeductionValueLocked(reason)
	after := before + change
	if after < 0 {
		after = 0
	}
	s.dailyCredit[dailyCreditKey(userID, time.Now())] = after
	if s.repo != nil {
		_ = s.repo.SaveTodayCredit(context.Background(), userID, time.Now(), after)
	}
	log := CreditLog{
		ID:          s.nextCreditID,
		UserID:      userID,
		GameID:      gameID,
		ChangeValue: after - before,
		BeforeScore: before,
		AfterScore:  after,
		Reason:      reason,
		CreatedAt:   time.Now(),
	}
	if s.repo != nil {
		if saved, err := s.repo.AddCreditLog(context.Background(), log); err == nil {
			log = saved
		}
	}
	s.nextCreditID++
	s.creditLogs = append(s.creditLogs, log)
	profile := s.profileLocked(userID)
	profile.CreditScore = after
	profile.TodayCreditScore = after
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	s.growth[userID] = profile
	footprint := Footprint{UserID: userID, GameID: gameID, Action: "credit_deducted", CreatedAt: time.Now()}
	if s.repo != nil {
		if saved, err := s.repo.AddFootprint(context.Background(), footprint); err == nil {
			footprint = saved
		}
	}
	s.footprints = append(s.footprints, footprint)
	return log
}

func (s *Service) RestoreCredit(userID int64, gameID int64, reason string, amount int) CreditLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	if amount < 0 {
		amount = -amount
	}
	before := s.todayCreditLocked(userID)
	after := before + amount
	if after > s.growthRulesLocked().CreditScoreCap {
		after = s.growthRulesLocked().CreditScoreCap
	}
	change := after - before
	s.dailyCredit[dailyCreditKey(userID, time.Now())] = after
	if s.repo != nil {
		_ = s.repo.SaveTodayCredit(context.Background(), userID, time.Now(), after)
	}
	log := CreditLog{
		ID:          s.nextCreditID,
		UserID:      userID,
		GameID:      gameID,
		ChangeValue: change,
		BeforeScore: before,
		AfterScore:  after,
		Reason:      reason,
		CreatedAt:   time.Now(),
	}
	if s.repo != nil {
		if saved, err := s.repo.AddCreditLog(context.Background(), log); err == nil {
			log = saved
		}
	}
	s.nextCreditID++
	s.creditLogs = append(s.creditLogs, log)
	profile := s.profileLocked(userID)
	profile.CreditScore = after
	profile.TodayCreditScore = after
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	s.growth[userID] = profile
	footprint := Footprint{UserID: userID, GameID: gameID, Action: "credit_restored", CreatedAt: time.Now()}
	if s.repo != nil {
		if saved, err := s.repo.AddFootprint(context.Background(), footprint); err == nil {
			footprint = saved
		}
	}
	s.footprints = append(s.footprints, footprint)
	return log
}

func (s *Service) creditDeductionValueLocked(reason string) int {
	if s.repo != nil {
		change, ok, err := s.repo.CreditDeductionValue(context.Background(), reason)
		if err == nil && ok {
			return change
		}
	}
	switch reason {
	case "player_cancel_service":
		return -3
	case "expert_cancel_service":
		return -5
	}
	return -10
}

func (s *Service) TraceByUser(userID int64) Trace {
	if s.repo != nil {
		trace := Trace{UserID: userID, Profile: s.Profile(userID)}
		if items, err := s.repo.ListReviewsByUser(context.Background(), userID); err == nil {
			trace.Reviews = items
		}
		if items, err := s.repo.ListCreditLogsByUser(context.Background(), userID); err == nil {
			trace.CreditLogs = items
		}
		if items, err := s.repo.ListFootprintsByUser(context.Background(), userID); err == nil {
			trace.Footprints = items
		}
		if items, err := s.repo.ListAchievementsByUser(context.Background(), userID); err == nil {
			trace.Achievements = items
		}
		return trace
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
	return trace
}

func (s *Service) TraceByGame(gameID int64) Trace {
	if s.repo != nil {
		trace := Trace{GameID: gameID}
		if items, err := s.repo.ListReviewsByGame(context.Background(), gameID); err == nil {
			trace.Reviews = items
		}
		if items, err := s.repo.ListCreditLogsByGame(context.Background(), gameID); err == nil {
			trace.CreditLogs = items
		}
		if items, err := s.repo.ListFootprints(context.Background()); err == nil {
			for _, item := range items {
				if item.GameID == gameID {
					trace.Footprints = append(trace.Footprints, item)
				}
			}
		}
		if items, err := s.repo.ListAchievements(context.Background()); err == nil {
			trace.Achievements = items
		}
		return trace
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
	return trace
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
	profile.Level = profile.Experience/rules.ExperiencePerLevel + rules.InitialLevel
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
	profile.Level = profile.Experience/rules.ExperiencePerLevel + rules.InitialLevel
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

func (s *Service) profileWithCreditLocked(userID int64) GrowthProfile {
	profile := s.profileLocked(userID)
	rules := s.growthRulesLocked()
	profile.Level = profile.Experience/rules.ExperiencePerLevel + rules.InitialLevel
	profile.TodayCreditScore = s.todayCreditLocked(userID)
	profile.CreditScore = profile.TodayCreditScore
	for _, item := range s.achievements[userID] {
		profile.Achievements = append(profile.Achievements, item.Code)
	}
	return profile
}

func (s *Service) todayCreditLocked(userID int64) int {
	key := dailyCreditKey(userID, time.Now())
	if s.repo != nil {
		if score, err := s.repo.GetTodayCredit(context.Background(), userID, time.Now()); err == nil {
			s.dailyCredit[key] = score
			return score
		}
	}
	score, ok := s.dailyCredit[key]
	if !ok {
		score = s.growthRulesLocked().InitialCreditScore
		s.dailyCredit[key] = score
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

func (s *Service) todayCreditFromRepository(userID int64) int {
	score, err := s.repo.GetTodayCredit(context.Background(), userID, time.Now())
	if err != nil {
		return s.GrowthRules().InitialCreditScore
	}
	return score
}

func dailyCreditKey(userID int64, now time.Time) string {
	return strconv.FormatInt(userID, 10) + ":" + now.Format("2006-01-02")
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
