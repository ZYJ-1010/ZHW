package reviews

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

type growthRulesProviderStub struct{ value GrowthRules }

func (p growthRulesProviderStub) Get(key string, target interface{}) bool {
	if key != growthRulesConfigKey {
		return false
	}
	raw, _ := json.Marshal(p.value)
	return json.Unmarshal(raw, target) == nil
}

func TestGrowthRewardsUseAdminRules(t *testing.T) {
	gameProvider := &fakeGameProvider{game: games.Game{ID: 1, Status: "pending_review"}, members: map[int64][]int64{1: {1, 2}}}
	service := NewService(gameProvider)
	service.SetGrowthRulesProvider(growthRulesProviderStub{value: GrowthRules{
		CompletedGameExperience: 20, SubmittedReviewExperience: 7, ReceivedReviewExperience: 11,
		SubmittedReviewPoints: 4, ExperiencePerLevel: 25, InitialLevel: 1, InitialCreditScore: 80, CreditScoreCap: 120,
	}})
	service.MarkGameReviewable(1)
	if rules := service.GrowthRules(); rules.InitialCreditScore != 100 {
		t.Fatalf("expected fixed initial credit score 100, got %+v", rules)
	}
	profiles := service.AwardCompletedGame(1)
	if len(profiles) != 2 || profiles[0].Experience != 20 || profiles[0].Level != 1 {
		t.Fatalf("expected configured completion reward, got %+v", profiles)
	}
	if _, profile, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5}); err != nil || profile.Experience != 27 || profile.Level != 2 {
		t.Fatalf("expected configured review reward and level, profile=%+v err=%v", profile, err)
	}
}

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
	npsScore := 9

	todos, err := service.Todos(1)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.markedReviewable || len(todos) != 1 || todos[0].TargetUserID != 2 {
		t.Fatalf("expected repository review todo, todos=%+v repo=%+v", todos, repo)
	}

	review, profile, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5, Tags: []string{"fun", "organized"}, AgainIntent: "yes", NPSScore: &npsScore})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.savedReview || review.ID == 0 || len(review.Tags) != 2 || review.NPSScore == nil || *review.NPSScore != 9 {
		t.Fatalf("expected repository review save, review=%+v repo=%+v", review, repo)
	}
	if !repo.savedGrowth || !repo.addedExperience || !repo.addedPoints || !repo.addedFootprint || !repo.ensuredAchievement {
		t.Fatalf("expected growth side effects persisted, profile=%+v repo=%+v", profile, repo)
	}

	if _, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5}); err != ErrDuplicateReview {
		t.Fatalf("expected duplicate review from repository, got %v", err)
	}
	invalidNPS := 11
	if _, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5, NPSScore: &invalidNPS}); err != ErrInvalidReview {
		t.Fatalf("expected invalid nps score to be rejected, got %v", err)
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

func TestTodosAndSubmitUseServerGameMemberRoles(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game: games.Game{ID: 1, Status: "pending_review"},
		members: map[int64][]int64{
			1: {1, 2, 3},
		},
		roles: map[int64][]games.MemberRole{
			1: {
				{UserID: 1, Role: "member"},
				{UserID: 2, Role: "expert"},
				{UserID: 3, Role: "main_guide"},
			},
		},
	}
	service := NewService(gameProvider)
	service.MarkGameReviewable(1)

	todos, err := service.Todos(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 2 || todos[0].TargetRole == "member" || todos[1].TargetRole == "member" {
		t.Fatalf("expected expert and guide target roles, got %+v", todos)
	}

	review, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5})
	if err != nil {
		t.Fatal(err)
	}
	if review.TargetRole != "expert" {
		t.Fatalf("expected server-canonical expert role, got %+v", review)
	}
	guideReview, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 3, TargetRole: "main_guide", Score: 5})
	if err != nil {
		t.Fatal(err)
	}
	if guideReview.TargetRole != "main_guide" {
		t.Fatalf("expected server-canonical main guide role, got %+v", guideReview)
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

func TestRepositoryRejectsSelfReview(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game: games.Game{ID: 1, CreatorUserID: 1, Status: "pending_review"},
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
	if len(todos) != 1 || todos[0].TargetUserID != 2 {
		t.Fatalf("expected self excluded from review todos, got %+v", todos)
	}
	if _, _, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 1, TargetRole: "member", Score: 5}); err != ErrInvalidReview {
		t.Fatalf("expected self review rejected, got %v", err)
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
	game        games.Game
	members     map[int64][]int64
	roles       map[int64][]games.MemberRole
	invitations map[int64][]games.Invitation
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

func (f *fakeGameProvider) MemberRoles(gameID int64) []games.MemberRole {
	return append([]games.MemberRole(nil), f.roles[gameID]...)
}

func (f *fakeGameProvider) InvitationsForUser(userID int64) []games.Invitation {
	return append([]games.Invitation(nil), f.invitations[userID]...)
}

func TestReviewTargetsFollowServiceBindingAndOrganizerClosure(t *testing.T) {
	gameID := int64(88)
	pair := games.Invitation{
		ID: 1, GameID: gameID, TargetUserID: 3, PlayerUserID: 2, ExpertUserID: 3,
		Role: "expert", Status: "accepted", InviteGroupID: "pair-88",
	}
	provider := &fakeGameProvider{
		game:    games.Game{ID: gameID, CreatorUserID: 1, Status: "pending_review"},
		members: map[int64][]int64{gameID: {1, 2, 3, 4, 5}},
		roles: map[int64][]games.MemberRole{gameID: {
			{UserID: 1, Role: "member"}, {UserID: 2, Role: "member"},
			{UserID: 3, Role: "expert"}, {UserID: 4, Role: "member"}, {UserID: 5, Role: "guide"},
		}},
		invitations: map[int64][]games.Invitation{2: {pair}, 3: {pair}},
	}
	service := NewService(provider)

	assertTargets := func(reviewerID int64, want ...int64) {
		t.Helper()
		got := service.reviewTargets(gameID, reviewerID)
		if len(got) != len(want) {
			t.Fatalf("reviewer %d targets = %v, want %v", reviewerID, got, want)
		}
		seen := make(map[int64]bool, len(got))
		for _, id := range got {
			seen[id] = true
		}
		for _, id := range want {
			if !seen[id] {
				t.Fatalf("reviewer %d targets = %v, missing %d", reviewerID, got, id)
			}
		}
	}

	assertTargets(1, 2, 3, 4, 5)
	assertTargets(2, 3, 5, 1)
	assertTargets(3, 2, 5, 1)
	assertTargets(4, 1)
	assertTargets(5, 1, 2, 3, 4)
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
