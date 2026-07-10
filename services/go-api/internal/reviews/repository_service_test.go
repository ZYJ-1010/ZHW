package reviews

import (
	"context"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

func TestRepositoryPersistsReviewGrowthCreditFootprintsAndAchievements(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game: games.Game{ID: 1, Status: "pending_review"},
		members: map[int64][]int64{
			1: {1, 2},
		},
	}
	repo := newFakeRepository()
	service := NewServiceWithRepository(gameProvider, repo)
	service.MarkGameReviewable(1)

	todos, err := service.Todos(1)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.markedReviewable || len(todos) != 1 || todos[0].TargetUserID != 2 {
		t.Fatalf("expected repository review todo, todos=%+v repo=%+v", todos, repo)
	}

	review, profile, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5, Tags: []string{"fun", "organized"}, AgainIntent: "yes"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.savedReview || review.ID == 0 || len(review.Tags) != 2 {
		t.Fatalf("expected repository review save, review=%+v repo=%+v", review, repo)
	}
	if !repo.savedGrowth || !repo.addedExperience || !repo.addedPoints || !repo.addedFootprint || !repo.ensuredAchievement {
		t.Fatalf("expected growth side effects persisted, profile=%+v repo=%+v", profile, repo)
	}

	if _, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5}); err != ErrDuplicateReview {
		t.Fatalf("expected duplicate review from repository, got %v", err)
	}
	if service.GameReviewComplete(1) {
		t.Fatal("expected game review incomplete before all member pairs review")
	}
	if _, _, err := service.Submit(2, SubmitRequest{GameID: 1, TargetUserID: 1, TargetRole: "member", Score: 5}); err != nil {
		t.Fatal(err)
	}
	if !service.GameReviewComplete(1) {
		t.Fatal("expected repository reviews to complete game review")
	}
	if items := service.MyIntents(1); len(items) != 1 || items[0].AgainIntent != "yes" || len(items[0].Tags) != 2 {
		t.Fatalf("expected my intents from repository, got %+v", items)
	}
	if profile := service.Profile(1); profile.Experience != 15 || profile.AvailablePoints != 2 || len(profile.Achievements) != 2 {
		t.Fatalf("expected profile from repository, got %+v", profile)
	}

	credit := service.DeductCredit(1, 1, "quit_after_started")
	if credit.AfterScore != 90 || !repo.addedCreditLog {
		t.Fatalf("expected credit log from repository, credit=%+v repo=%+v", credit, repo)
	}
	trace := service.TraceByUser(1)
	if len(trace.Reviews) != 2 || len(trace.CreditLogs) != 1 || len(trace.Footprints) == 0 || len(trace.Achievements) != 2 {
		t.Fatalf("expected trace from repository, got %+v", trace)
	}
}

func TestRepositoryReviewDeadlineBlocksExpiredTodosAndSubmit(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game: games.Game{ID: 1, Status: "pending_review"},
		members: map[int64][]int64{
			1: {1, 2},
		},
	}
	repo := newFakeRepository()
	repo.reviewable[1] = map[int64]time.Time{1: time.Now().Add(-time.Hour)}
	repo.reviewable[2] = map[int64]time.Time{1: time.Now().Add(-time.Hour)}
	service := NewServiceWithRepository(gameProvider, repo)

	todos, err := service.Todos(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 0 {
		t.Fatalf("expected expired review todo hidden, got %+v", todos)
	}
	if _, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, Score: 5}); err != ErrGameNotReviewable {
		t.Fatalf("expected expired review blocked, got %v", err)
	}
}

func TestRepositoryCreditDeductionUsesRuleWhenConfigured(t *testing.T) {
	repo := newFakeRepository()
	repo.creditRules["quit_after_started"] = -25
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	credit := service.DeductCredit(1, 2, "quit_after_started")
	if credit.BeforeScore != 100 || credit.AfterScore != 75 || credit.ChangeValue != -25 {
		t.Fatalf("expected configured deduction, got %+v", credit)
	}

	fallback := service.DeductCredit(2, 2, "quit_after_confirm")
	if fallback.BeforeScore != 100 || fallback.AfterScore != 90 || fallback.ChangeValue != -10 {
		t.Fatalf("expected default deduction, got %+v", fallback)
	}
}

func TestRepositoryProfileInitializesDefaultGrowth(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	profile := service.Profile(99)
	if profile.UserID != 99 || profile.Level != 1 || profile.CreditScore != 100 || profile.TodayCreditScore != 100 {
		t.Fatalf("expected default growth profile, got %+v", profile)
	}
	if !repo.savedGrowth {
		t.Fatal("expected default growth profile persisted")
	}
	if saved, ok := repo.growth[99]; !ok || saved.UserID != 99 || saved.Level != 1 {
		t.Fatalf("expected persisted default growth profile, got %+v ok=%v", saved, ok)
	}
}

type fakeGameProvider struct {
	game    games.Game
	members map[int64][]int64
}

func (f *fakeGameProvider) Get(id int64) (games.Game, error) {
	return f.game, nil
}

func (f *fakeGameProvider) Members(gameID int64) []int64 {
	return append([]int64(nil), f.members[gameID]...)
}

func (f *fakeGameProvider) IsMember(gameID int64, userID int64) bool {
	for _, item := range f.members[gameID] {
		if item == userID {
			return true
		}
	}
	return false
}

type fakeRepository struct {
	nextReviewID       int64
	nextCreditID       int64
	reviewable         map[int64]map[int64]time.Time
	reviews            []Review
	growth             map[int64]GrowthProfile
	credits            []CreditLog
	footprints         []Footprint
	achievements       map[int64][]Achievement
	todayCredit        map[int64]int
	creditRules        map[string]int
	markedReviewable   bool
	savedReview        bool
	savedGrowth        bool
	addedExperience    bool
	addedPoints        bool
	addedFootprint     bool
	ensuredAchievement bool
	addedCreditLog     bool
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		nextReviewID: 1,
		nextCreditID: 1,
		reviewable:   make(map[int64]map[int64]time.Time),
		reviews:      make([]Review, 0),
		growth:       make(map[int64]GrowthProfile),
		credits:      make([]CreditLog, 0),
		footprints:   make([]Footprint, 0),
		achievements: make(map[int64][]Achievement),
		todayCredit:  make(map[int64]int),
		creditRules:  make(map[string]int),
	}
}

func (r *fakeRepository) MarkReviewable(ctx context.Context, gameID int64, userID int64, deadline time.Time) error {
	r.markedReviewable = true
	if r.reviewable[userID] == nil {
		r.reviewable[userID] = make(map[int64]time.Time)
	}
	r.reviewable[userID][gameID] = deadline
	return nil
}

func (r *fakeRepository) ListReviewable(ctx context.Context, userID int64) (map[int64]time.Time, error) {
	result := make(map[int64]time.Time)
	for gameID, deadline := range r.reviewable[userID] {
		result[gameID] = deadline
	}
	return result, nil
}

func (r *fakeRepository) ReviewExists(ctx context.Context, gameID int64, reviewerID int64, targetID int64, targetRole string) (bool, error) {
	for _, item := range r.reviews {
		if item.GameID == gameID && item.ReviewerUserID == reviewerID && item.TargetUserID == targetID && item.TargetRole == targetRole {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepository) SaveReview(ctx context.Context, review Review) (Review, error) {
	r.savedReview = true
	review.ID = r.nextReviewID
	r.nextReviewID++
	r.reviews = append(r.reviews, review)
	return review, nil
}

func (r *fakeRepository) ListReviews(ctx context.Context) ([]Review, error) {
	return append([]Review(nil), r.reviews...), nil
}

func (r *fakeRepository) ListReviewsByUser(ctx context.Context, userID int64) ([]Review, error) {
	result := make([]Review, 0)
	for _, item := range r.reviews {
		if item.ReviewerUserID == userID || item.TargetUserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRepository) ListReviewsByGame(ctx context.Context, gameID int64) ([]Review, error) {
	result := make([]Review, 0)
	for _, item := range r.reviews {
		if item.GameID == gameID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRepository) GetGrowthProfile(ctx context.Context, userID int64) (GrowthProfile, bool, error) {
	profile, ok := r.growth[userID]
	return profile, ok, nil
}

func (r *fakeRepository) SaveGrowthProfile(ctx context.Context, profile GrowthProfile) (GrowthProfile, error) {
	r.savedGrowth = true
	r.growth[profile.UserID] = profile
	return profile, nil
}

func (r *fakeRepository) AddExperienceLog(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, createdAt time.Time) error {
	r.addedExperience = true
	return nil
}

func (r *fakeRepository) AddPointsLog(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, createdAt time.Time) error {
	r.addedPoints = true
	return nil
}

func (r *fakeRepository) AddFootprint(ctx context.Context, footprint Footprint) (Footprint, error) {
	r.addedFootprint = true
	r.footprints = append(r.footprints, footprint)
	return footprint, nil
}

func (r *fakeRepository) ListFootprintsByUser(ctx context.Context, userID int64) ([]Footprint, error) {
	result := make([]Footprint, 0)
	for _, item := range r.footprints {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRepository) ListFootprints(ctx context.Context) ([]Footprint, error) {
	return append([]Footprint(nil), r.footprints...), nil
}

func (r *fakeRepository) GetTodayCredit(ctx context.Context, userID int64, now time.Time) (int, error) {
	score, ok := r.todayCredit[userID]
	if !ok {
		score = 100
		r.todayCredit[userID] = score
	}
	return score, nil
}

func (r *fakeRepository) SaveTodayCredit(ctx context.Context, userID int64, now time.Time, score int) error {
	r.todayCredit[userID] = score
	return nil
}

func (r *fakeRepository) CreditDeductionValue(ctx context.Context, ruleCode string) (int, bool, error) {
	change, ok := r.creditRules[ruleCode]
	return change, ok, nil
}

func (r *fakeRepository) ListCreditDeductionRules(ctx context.Context) ([]CreditDeductionRule, error) {
	items := make([]CreditDeductionRule, 0, len(r.creditRules))
	for code, change := range r.creditRules {
		items = append(items, CreditDeductionRule{RuleCode: code, ChangeValue: change, Enabled: true})
	}
	return items, nil
}

func (r *fakeRepository) UpsertCreditDeductionRule(ctx context.Context, rule CreditDeductionRule) (CreditDeductionRule, error) {
	r.creditRules[rule.RuleCode] = rule.ChangeValue
	return rule, nil
}

func (r *fakeRepository) AddCreditLog(ctx context.Context, log CreditLog) (CreditLog, error) {
	r.addedCreditLog = true
	log.ID = r.nextCreditID
	r.nextCreditID++
	r.credits = append(r.credits, log)
	return log, nil
}

func (r *fakeRepository) ListCreditLogsByUser(ctx context.Context, userID int64) ([]CreditLog, error) {
	result := make([]CreditLog, 0)
	for _, item := range r.credits {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRepository) ListCreditLogsByGame(ctx context.Context, gameID int64) ([]CreditLog, error) {
	result := make([]CreditLog, 0)
	for _, item := range r.credits {
		if item.GameID == gameID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRepository) EnsureAchievement(ctx context.Context, achievement Achievement) (Achievement, bool, error) {
	r.ensuredAchievement = true
	for _, item := range r.achievements[achievement.UserID] {
		if item.Code == achievement.Code {
			return item, false, nil
		}
	}
	r.achievements[achievement.UserID] = append(r.achievements[achievement.UserID], achievement)
	return achievement, true, nil
}

func (r *fakeRepository) ListAchievementsByUser(ctx context.Context, userID int64) ([]Achievement, error) {
	return append([]Achievement(nil), r.achievements[userID]...), nil
}

func (r *fakeRepository) ListAchievements(ctx context.Context) ([]Achievement, error) {
	result := make([]Achievement, 0)
	for _, items := range r.achievements {
		result = append(result, items...)
	}
	return result, nil
}
