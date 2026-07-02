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

type GameProvider interface {
	Get(id int64) (games.Game, error)
	Members(gameID int64) []int64
	IsMember(gameID int64, userID int64) bool
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
	mu           sync.RWMutex
	nextID       int64
	nextCreditID int64
	games        GameProvider
	reviews      []Review
	reviewed     map[string]bool
	reviewable   map[int64]time.Time
	growth       map[int64]GrowthProfile
	dailyCredit  map[string]int
	creditLogs   []CreditLog
	achievements map[int64][]Achievement
	footprints   []Footprint
	repo         Repository
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
		profiles = append(profiles, s.addExperienceOnlyLocked(userID, 10, gameID, "completed_game"))
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
		for _, targetID := range s.games.Members(gameID) {
			if targetID == userID || s.reviewed[reviewKey(gameID, userID, targetID, "member")] {
				continue
			}
			result = append(result, Todo{
				GameID:       gameID,
				TargetUserID: targetID,
				TargetRole:   "member",
				DeadlineAt:   deadline.Format(time.RFC3339),
			})
		}
	}
	return result, nil
}

func (s *Service) Submit(userID int64, req SubmitRequest) (Review, GrowthProfile, error) {
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
	profile := s.addGrowthLocked(userID, 5, 2, req.GameID, "submitted_review")
	s.addGrowthLocked(req.TargetUserID, 10, 0, req.GameID, "received_review")
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
	if after > 100 {
		after = 100
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
	for _, reviewerID := range members {
		for _, targetID := range members {
			if reviewerID == targetID {
				continue
			}
			if !s.reviewed[reviewKey(gameID, reviewerID, targetID, "member")] {
				return false
			}
		}
	}
	return len(members) > 1
}

func (s *Service) gameReviewCompleteFromRepository(gameID int64) bool {
	members := s.games.Members(gameID)
	if len(members) <= 1 {
		return false
	}
	reviewable := false
	for _, reviewerID := range members {
		reviewableGames, err := s.repo.ListReviewable(context.Background(), reviewerID)
		if err != nil {
			return false
		}
		if _, ok := reviewableGames[gameID]; ok {
			reviewable = true
		}
		for _, targetID := range members {
			if reviewerID == targetID {
				continue
			}
			exists, err := s.repo.ReviewExists(context.Background(), gameID, reviewerID, targetID, "member")
			if err != nil || !exists {
				return false
			}
		}
	}
	return reviewable
}

func (s *Service) addGrowthLocked(userID int64, exp int, points int, gameID int64, action string) GrowthProfile {
	profile := s.profileLocked(userID)
	profile.Experience += exp
	profile.AvailablePoints += points
	profile.ReviewCount++
	profile.Level = profile.Experience/100 + 1
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

func (s *Service) addExperienceOnlyLocked(userID int64, exp int, gameID int64, action string) GrowthProfile {
	profile := s.profileLocked(userID)
	profile.Experience += exp
	profile.Level = profile.Experience/100 + 1
	profile.UpdatedAt = time.Now().Format(time.RFC3339)
	if s.repo != nil {
		if saved, err := s.repo.SaveGrowthProfile(context.Background(), profile); err == nil {
			profile = saved
		}
		_ = s.repo.AddExperienceLog(context.Background(), userID, gameID, exp, action, time.Now())
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
	}
	return profile
}

func defaultGrowthProfile(userID int64) GrowthProfile {
	return GrowthProfile{
		UserID:           userID,
		Level:            1,
		CreditScore:      100,
		TodayCreditScore: 100,
		UpdatedAt:        time.Now().Format(time.RFC3339),
	}
}

func (s *Service) profileWithCreditLocked(userID int64) GrowthProfile {
	profile := s.profileLocked(userID)
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
		score = 100
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
		for _, targetID := range s.games.Members(gameID) {
			if targetID == userID {
				continue
			}
			exists, err := s.repo.ReviewExists(context.Background(), gameID, userID, targetID, "member")
			if err != nil {
				return nil, err
			}
			if exists {
				continue
			}
			result = append(result, Todo{
				GameID:       gameID,
				TargetUserID: targetID,
				TargetRole:   "member",
				DeadlineAt:   deadline.Format(time.RFC3339),
			})
		}
	}
	return result, nil
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
		return 100
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
	case "", "member":
		return true
	default:
		return false
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
