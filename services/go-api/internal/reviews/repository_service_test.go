package reviews

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

type growthRulesProviderStub struct{ value GrowthRules }

type failingGrowthRulesStrictProvider struct{}

func (failingGrowthRulesStrictProvider) GetStrict(string, interface{}) (bool, error) {
	return false, errors.New("growth rules unavailable")
}

func (p growthRulesProviderStub) Get(key string, target interface{}) bool {
	if key != growthRulesConfigKey {
		return false
	}
	raw, _ := json.Marshal(p.value)
	return json.Unmarshal(raw, target) == nil
}

func TestGrowthRewardsUseAdminRules(t *testing.T) {
	gameProvider := &fakeGameProvider{game: games.Game{ID: 1, CreatorUserID: 1, Status: "pending_review"}, members: map[int64][]int64{1: {1, 2}}}
	service := NewService(gameProvider)
	service.SetGrowthRulesProvider(growthRulesProviderStub{value: GrowthRules{
		CompletedGameExperience: 20, SubmittedReviewExperience: 7, ReceivedReviewExperience: 11,
		ExperiencePerLevel: 25, InitialLevel: 1, InitialCreditScore: 80, CreditScoreCap: 120,
	}})
	service.MarkGameReviewable(1)
	if rules := service.GrowthRules(); rules.InitialCreditScore != 100 {
		t.Fatalf("expected fixed initial credit score 100, got %+v", rules)
	}
	profiles := service.AwardCompletedGame(1)
	if len(profiles) != 2 || profiles[0].Experience != 20 || profiles[0].Level != 1 || profiles[1].Experience != 20 || profiles[1].Level != 1 {
		t.Fatalf("expected configured completion experience for all occupied seats, got %+v", profiles)
	}
	if _, profile, err := service.Submit(1, SubmitRequest{GameID: 1, TargetUserID: 2, TargetRole: "member", Score: 5}); err != nil || profile.Experience != 27 || profile.Level != 2 {
		t.Fatalf("expected configured review reward and level, profile=%+v err=%v", profile, err)
	}
}

func TestCreditMutationFailsClosedWhenConfiguredLimitsCannotBeRead(t *testing.T) {
	service := NewService(&fakeGameProvider{})
	service.SetCreditAccountRulesStrictProvider(func() (int, int, error) {
		return 0, 0, errors.New("credit config unavailable")
	})
	if _, err := service.DeductCreditStrict(1, 1, "low_review"); err == nil {
		t.Fatal("credit mutation must not use fallback limits when configured limits are unavailable")
	}
	if profile := service.Profile(1); profile.CreditScore != 100 {
		t.Fatalf("failed credit mutation must not change the account: %+v", profile)
	}
}

func TestGrowthMutationFailsClosedWhenConfiguredRewardsCannotBeRead(t *testing.T) {
	gameProvider := &fakeGameProvider{game: games.Game{ID: 1, CreatorUserID: 1, Status: "pending_review"}, members: map[int64][]int64{1: {1}}}
	service := NewService(gameProvider)
	service.SetGrowthRulesStrictProvider(failingGrowthRulesStrictProvider{})
	if _, err := service.AwardCompletedGameStrict(1); err == nil {
		t.Fatal("growth mutation must not use default rewards when configured rewards are unavailable")
	}
	if profile := service.Profile(1); profile.Experience != 0 {
		t.Fatalf("failed growth mutation must not change the profile: %+v", profile)
	}
}

func TestCompletedGameExperienceExcludesGuideEscort(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game:    games.Game{ID: 1, CreatorUserID: 1, Status: "pending_review"},
		members: map[int64][]int64{1: {1, 2, 3}},
		roles: map[int64][]games.MemberRole{1: {
			{UserID: 1, Role: "member"},
			{UserID: 2, Role: "member"},
			{UserID: 3, Role: "guide_escort"},
		}},
	}
	service := NewService(gameProvider)
	profiles := service.AwardCompletedGame(1)
	if len(profiles) != 2 || service.Profile(1).Experience != 100 || service.Profile(2).Experience != 100 || service.Profile(3).Experience != 0 {
		t.Fatalf("expected only creator and actual member to receive completion experience, profiles=%+v", profiles)
	}
	if again := service.AwardCompletedGame(1); len(again) != 0 {
		t.Fatalf("expected completion experience to remain idempotent, got %+v", again)
	}
}

func TestCompletedGameWritesIdempotentGrowthEvents(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game:    games.Game{ID: 8, CreatorUserID: 1, Status: "pending_review"},
		members: map[int64][]int64{8: {1, 2}},
		roles: map[int64][]games.MemberRole{8: {
			{UserID: 1, Role: "member"},
			{UserID: 2, Role: "expert"},
		}},
	}
	repo := newFakeRepository()
	service := NewServiceWithRepository(gameProvider, repo)
	service.AwardCompletedGame(8)
	if len(repo.growthEvents) != 2 {
		t.Fatalf("expected one factual event per completed member, got %+v", repo.growthEvents)
	}
	if event := repo.growthEvents["completed_game:8:2"]; event.RoleCode != "expert" || event.EventCode != "game_completed_member" {
		t.Fatalf("expected expert completion event, got %+v", event)
	}
	service.AwardCompletedGame(8)
	if len(repo.growthEvents) != 2 {
		t.Fatalf("expected duplicate completion to keep ledger idempotent, got %+v", repo.growthEvents)
	}
}

func TestCompletedGameGrowthMutationIsAtomicAndRetryable(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game:    games.Game{ID: 18, CreatorUserID: 1, Status: "pending_review"},
		members: map[int64][]int64{18: {1, 2}},
	}
	repo := newFakeRepository()
	repo.growthMutationErr = errors.New("growth transaction failed")
	service := NewServiceWithRepository(gameProvider, repo)
	if _, err := service.AwardCompletedGameStrict(18); err == nil {
		t.Fatal("expected growth mutation failure")
	}
	if len(repo.growth) != 0 || len(repo.footprints) != 0 {
		t.Fatalf("failed growth mutation must not change profile or footprint, growth=%+v footprints=%+v", repo.growth, repo.footprints)
	}
	repo.growthMutationErr = nil
	profiles, err := service.AwardCompletedGameStrict(18)
	if err != nil || len(profiles) != 2 {
		t.Fatalf("expected retry to award both members, profiles=%+v err=%v", profiles, err)
	}
	profiles, err = service.AwardCompletedGameStrict(18)
	if err != nil || len(profiles) != 0 || service.Profile(1).Experience != 100 {
		t.Fatalf("expected idempotent retry, profiles=%+v err=%v profile=%+v", profiles, err, service.Profile(1))
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
	if !repo.savedGrowth || !repo.addedExperience || repo.addedPoints || !repo.addedFootprint || !repo.ensuredAchievement {
		t.Fatalf("expected growth side effects persisted, profile=%+v repo=%+v", profile, repo)
	}
	if len(repo.growthEvents) != 2 || !hasGrowthEventCode(repo.growthEvents, "review_submitted") || !hasGrowthEventCode(repo.growthEvents, "positive_review_received") {
		t.Fatalf("expected review facts recorded for future role metrics, got %+v", repo.growthEvents)
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
	if profile := service.Profile(1); profile.Experience != 15 || profile.AvailablePoints != 0 || len(profile.Achievements) != 2 {
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

func hasGrowthEventCode(events map[string]GrowthEvent, code string) bool {
	for _, event := range events {
		if event.EventCode == code {
			return true
		}
	}
	return false
}

func TestMarkGameReviewableBatchFailureCanBeRetried(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game:    games.Game{ID: 31, Status: "pending_review"},
		members: map[int64][]int64{31: {1, 2}},
	}
	repo := newFakeRepository()
	repo.markReviewableErr = errors.New("review reminder transaction failed")
	service := NewServiceWithRepository(gameProvider, repo)

	if err := service.MarkGameReviewableStrict(31); err == nil {
		t.Fatal("expected batch persistence failure")
	}
	if len(repo.reviewable) != 0 || len(service.reviewable) != 0 {
		t.Fatalf("failed batch must not leave partial reviewable state, repo=%+v memory=%+v", repo.reviewable, service.reviewable)
	}

	repo.markReviewableErr = nil
	if _, err := service.Todos(1); err != nil {
		t.Fatalf("expected task-center read to repair review reminders: %v", err)
	}
	if len(repo.reviewable[1]) != 1 || len(repo.reviewable[2]) != 1 {
		t.Fatalf("expected reminders for all members, got %+v", repo.reviewable)
	}
}

func TestReviewAndGrowthTransactionFailureLeavesNoPartialReview(t *testing.T) {
	gameProvider := &fakeGameProvider{
		game:    games.Game{ID: 41, Status: "pending_review"},
		members: map[int64][]int64{41: {1, 2}},
	}
	repo := newFakeRepository()
	service := NewServiceWithRepository(gameProvider, repo)
	service.MarkGameReviewable(41)
	repo.reviewGrowthErr = errors.New("review growth transaction failed")
	request := SubmitRequest{GameID: 41, TargetUserID: 2, TargetRole: "member", Score: 5, Content: "很好"}
	if _, _, err := service.Submit(1, request); err == nil {
		t.Fatal("expected review growth transaction failure")
	}
	if len(repo.reviews) != 0 || len(repo.growth) != 0 || len(repo.footprints) != 0 || len(repo.achievements) != 0 {
		t.Fatalf("failed review transaction must leave no partial state: reviews=%+v growth=%+v footprints=%+v achievements=%+v", repo.reviews, repo.growth, repo.footprints, repo.achievements)
	}
	repo.reviewGrowthErr = nil
	review, profile, err := service.Submit(1, request)
	if err != nil || review.ID == 0 || profile.Experience == 0 || len(repo.achievements) != 2 {
		t.Fatalf("expected clean retry to persist all review effects, review=%+v profile=%+v err=%v", review, profile, err)
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

func TestRepositoryCreditDeductionDisabledRuleDoesNotFallback(t *testing.T) {
	repo := newFakeRepository()
	// 仓库以 found=true、change=0 表示规则存在但已停用；服务层不得再
	// 回退到默认 -10。
	repo.creditRules["low_review"] = 0
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	credit := service.DeductCredit(1, 2, "low_review")
	if credit.BeforeScore != 100 || credit.AfterScore != 100 || credit.ChangeValue != 0 {
		t.Fatalf("expected disabled rule to keep credit unchanged, got %+v", credit)
	}
}

func TestCreditAccountIsPermanentAndDoesNotFollowDailyScore(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	if log := service.DeductCredit(88, 0, "quit_after_confirm"); log.BeforeScore != 100 || log.AfterScore != 90 {
		t.Fatalf("expected permanent credit deduction, got %+v", log)
	}
	// 模拟旧的按日表在次日生成了新的 100 分记录；永久账户不能被其覆盖。
	repo.todayCredit[88] = 100
	if profile := service.Profile(88); profile.CreditScore != 90 || profile.TodayCreditScore != 90 {
		t.Fatalf("expected profile to use permanent credit account, got %+v", profile)
	}
}

func TestCreditAccountUsesConfiguredInitialScoreAndCap(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)
	service.SetCreditAccountRulesProvider(func() (int, int) { return 80, 120 })

	if profile := service.Profile(77); profile.CreditScore != 80 || profile.TodayCreditScore != 80 {
		t.Fatalf("expected configured initial permanent credit score, got %+v", profile)
	}
	if log := service.RestoreCredit(77, 0, "appeal_passed", 100); log.AfterScore != 120 {
		t.Fatalf("expected configured permanent credit cap, got %+v", log)
	}
}

func TestStrictCreditMutationDoesNotChangeStateWhenRepositoryFails(t *testing.T) {
	repo := newFakeRepository()
	repo.creditAccounts[44] = CreditAccount{UserID: 44, CurrentScore: 100}
	repo.creditMutationErr = errors.New("credit transaction failed")
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	if log, err := service.DeductCreditStrict(44, 9, "quit_after_confirm"); err == nil || log.ID != 0 {
		t.Fatalf("expected strict mutation failure, log=%+v err=%v", log, err)
	}
	if account := repo.creditAccounts[44]; account.CurrentScore != 100 {
		t.Fatalf("repository failure must not change account, got %+v", account)
	}
	if len(repo.credits) != 0 || len(service.creditLogs) != 0 {
		t.Fatalf("repository failure must not create ledger, repo=%+v memory=%+v", repo.credits, service.creditLogs)
	}
}

func TestStrictGrowthReadDoesNotOverwriteProfileOnRepositoryFailure(t *testing.T) {
	repo := newFakeRepository()
	repo.growthReadErr = errors.New("growth profile read failed")
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	if profile, err := service.ProfileStrict(44); err == nil || profile.UserID != 0 {
		t.Fatalf("expected strict growth read failure, profile=%+v err=%v", profile, err)
	}
	if repo.savedGrowth {
		t.Fatal("growth read failure must not create or overwrite a default profile")
	}
}

func TestCreditDeductionOnceUsesBusinessIdempotencyKey(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	first, applied, err := service.DeductCreditOnceStrict(44, 9, "report_confirmed", "report_confirmed:501")
	if err != nil || !applied || first.BeforeScore != 100 || first.AfterScore != 90 {
		t.Fatalf("expected first business deduction, log=%+v applied=%v err=%v", first, applied, err)
	}
	second, applied, err := service.DeductCreditOnceStrict(44, 9, "report_confirmed", "report_confirmed:501")
	if err != nil || applied || second.ID != first.ID {
		t.Fatalf("expected retry to reuse the original deduction, log=%+v applied=%v err=%v", second, applied, err)
	}
	if account := repo.creditAccounts[44]; account.CurrentScore != 90 {
		t.Fatalf("expected credit deducted only once, got %+v", account)
	}
	if len(repo.credits) != 1 {
		t.Fatalf("expected one credit ledger, got %+v", repo.credits)
	}
}

func TestStrictCreditPathsDoNotFallbackOnRepositoryReadFailure(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)

	repo.creditRuleErr = errors.New("credit rule unavailable")
	if _, err := service.DeductCreditStrict(45, 9, "quit_after_confirm"); err == nil {
		t.Fatal("expected deduction to stop when configured rule cannot be read")
	}
	if len(repo.credits) != 0 {
		t.Fatalf("rule read failure must not apply fallback deduction, got %+v", repo.credits)
	}

	repo.creditRuleErr = nil
	repo.creditAccountErr = errors.New("credit account unavailable")
	if _, err := service.CreditScoreStrict(45); err == nil {
		t.Fatal("expected authorization credit read to fail closed")
	}
}

func TestCreditAppealRestoreIsIdempotentAndLinked(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)
	source := service.DeductCredit(77, 9, "quit_after_confirm")
	if source.ChangeValue != -10 {
		t.Fatalf("expected source deduction, got %+v", source)
	}

	first, applied, err := service.RestoreCreditForAppeal(77, 9, source.ID, 501, 50)
	if err != nil || !applied || first.ChangeValue != 10 || first.AppealID != 501 {
		t.Fatalf("expected first approval to restore and link credit, log=%+v applied=%v err=%v", first, applied, err)
	}
	second, applied, err := service.RestoreCreditForAppeal(77, 9, source.ID, 501, 50)
	if err != nil || applied || second.ID != first.ID {
		t.Fatalf("expected repeated approval to return the original restore log, log=%+v applied=%v err=%v", second, applied, err)
	}
	if profile := service.Profile(77); profile.CreditScore != 100 {
		t.Fatalf("expected credit restored only once, got %+v", profile)
	}
	if _, _, err := service.RestoreCreditForAppeal(77, 9, source.ID, 502, 10); err != ErrCreditAppealConflict {
		t.Fatalf("expected another appeal to be rejected, got %v", err)
	}
	logs, _ := repo.ListCreditLogsByUser(context.Background(), 77)
	if len(logs) != 2 || logs[0].AppealID != 501 || logs[1].AppealID != 501 {
		t.Fatalf("expected source and restore ledgers linked to one appeal, got %+v", logs)
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

func (f *fakeGameProvider) List() []games.Game {
	if f.game.ID <= 0 {
		return nil
	}
	return []games.Game{f.game}
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

func TestStrictReviewListsDoNotUseStaleProcessCache(t *testing.T) {
	repo := newFakeRepository()
	service := NewServiceWithRepository(&fakeGameProvider{}, repo)
	service.reviews = []Review{{ID: 9, ReviewerUserID: 1, AgainIntent: "yes"}}
	service.footprints = []Footprint{{UserID: 1, GameID: 8, Action: "旧足迹"}}
	repo.listReadErr = errors.New("review database unavailable")

	if _, err := service.MyIntentsStrict(1); !errors.Is(err, repo.listReadErr) {
		t.Fatalf("expected intent read error, got %v", err)
	}
	if _, err := service.AllReviewsStrict(); !errors.Is(err, repo.listReadErr) {
		t.Fatalf("expected review list error, got %v", err)
	}
	if _, err := service.FootprintsStrict(1); !errors.Is(err, repo.listReadErr) {
		t.Fatalf("expected footprint read error, got %v", err)
	}
	if _, err := service.AllFootprintsStrict(); !errors.Is(err, repo.listReadErr) {
		t.Fatalf("expected all footprint read error, got %v", err)
	}
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
	creditAccounts     map[int64]CreditAccount
	creditMutationKeys map[string]CreditLog
	growthEvents       map[string]GrowthEvent
	growthMutationKeys map[string]bool
	creditRules        map[string]int
	markedReviewable   bool
	markReviewableErr  error
	savedReview        bool
	savedGrowth        bool
	addedExperience    bool
	addedPoints        bool
	addedFootprint     bool
	ensuredAchievement bool
	addedCreditLog     bool
	creditMutationErr  error
	creditAccountErr   error
	creditRuleErr      error
	growthMutationErr  error
	reviewGrowthErr    error
	growthReadErr      error
	listReadErr        error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		nextReviewID:       1,
		nextCreditID:       1,
		reviewable:         make(map[int64]map[int64]time.Time),
		reviews:            make([]Review, 0),
		growth:             make(map[int64]GrowthProfile),
		credits:            make([]CreditLog, 0),
		footprints:         make([]Footprint, 0),
		achievements:       make(map[int64][]Achievement),
		todayCredit:        make(map[int64]int),
		creditAccounts:     make(map[int64]CreditAccount),
		creditMutationKeys: make(map[string]CreditLog),
		growthEvents:       make(map[string]GrowthEvent),
		growthMutationKeys: make(map[string]bool),
		creditRules:        make(map[string]int),
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

func (r *fakeRepository) MarkReviewableBatch(ctx context.Context, gameID int64, userIDs []int64, deadline time.Time) error {
	if r.markReviewableErr != nil {
		return r.markReviewableErr
	}
	for _, userID := range userIDs {
		if err := r.MarkReviewable(ctx, gameID, userID, deadline); err != nil {
			return err
		}
	}
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

func (r *fakeRepository) SaveReviewWithGrowth(ctx context.Context, bundle ReviewGrowthBundle, initialLevel int) (Review, GrowthProfile, GrowthProfile, []Footprint, []Achievement, error) {
	if r.reviewGrowthErr != nil {
		return Review{}, GrowthProfile{}, GrowthProfile{}, nil, nil, r.reviewGrowthErr
	}
	review, err := r.SaveReview(ctx, bundle.Review)
	if err != nil {
		return Review{}, GrowthProfile{}, GrowthProfile{}, nil, nil, err
	}
	reviewerProfile, reviewerFootprint, _, err := r.ApplyGrowthMutation(ctx, bundle.ReviewerMutation, initialLevel)
	if err != nil {
		return Review{}, GrowthProfile{}, GrowthProfile{}, nil, nil, err
	}
	targetProfile, targetFootprint, _, err := r.ApplyGrowthMutation(ctx, bundle.TargetMutation, initialLevel)
	if err != nil {
		return Review{}, GrowthProfile{}, GrowthProfile{}, nil, nil, err
	}
	for _, achievement := range bundle.Achievements {
		found := false
		for _, current := range r.achievements[achievement.UserID] {
			if current.Code == achievement.Code {
				found = true
				break
			}
		}
		if !found {
			r.achievements[achievement.UserID] = append(r.achievements[achievement.UserID], achievement)
		}
	}
	r.ensuredAchievement = true
	return review, reviewerProfile, targetProfile, []Footprint{reviewerFootprint, targetFootprint}, bundle.Achievements, nil
}

func (r *fakeRepository) ListReviews(ctx context.Context) ([]Review, error) {
	if r.listReadErr != nil {
		return nil, r.listReadErr
	}
	return append([]Review(nil), r.reviews...), nil
}

func (r *fakeRepository) ListReviewsByUser(ctx context.Context, userID int64) ([]Review, error) {
	if r.listReadErr != nil {
		return nil, r.listReadErr
	}
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
	if r.growthReadErr != nil {
		return GrowthProfile{}, false, r.growthReadErr
	}
	profile, ok := r.growth[userID]
	return profile, ok, nil
}

func (r *fakeRepository) SaveGrowthProfile(ctx context.Context, profile GrowthProfile) (GrowthProfile, error) {
	r.savedGrowth = true
	r.growth[profile.UserID] = profile
	return profile, nil
}

func (r *fakeRepository) ApplyGrowthMutation(ctx context.Context, mutation GrowthMutation, initialLevel int) (GrowthProfile, Footprint, bool, error) {
	if r.growthMutationErr != nil {
		return GrowthProfile{}, Footprint{}, false, r.growthMutationErr
	}
	if r.growthMutationKeys[mutation.IdempotencyKey] {
		return r.growth[mutation.UserID], Footprint{}, false, nil
	}
	profile, ok := r.growth[mutation.UserID]
	if !ok {
		profile = GrowthProfile{UserID: mutation.UserID, Level: initialLevel}
	}
	profile.Experience += mutation.ExperienceDelta
	profile.ReviewCount += mutation.ReviewCountDelta
	profile.UpdatedAt = mutation.CreatedAt.Format(time.RFC3339)
	r.growth[mutation.UserID] = profile
	footprint := Footprint{UserID: mutation.UserID, GameID: mutation.GameID, Action: mutation.Action, CreatedAt: mutation.CreatedAt}
	r.footprints = append(r.footprints, footprint)
	if mutation.Event != nil {
		r.growthEvents[mutation.Event.IdempotencyKey] = *mutation.Event
	}
	r.growthMutationKeys[mutation.IdempotencyKey] = true
	r.savedGrowth = true
	r.addedExperience = true
	r.addedFootprint = true
	return profile, footprint, true, nil
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
	if r.listReadErr != nil {
		return nil, r.listReadErr
	}
	result := make([]Footprint, 0)
	for _, item := range r.footprints {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeRepository) ListFootprints(ctx context.Context) ([]Footprint, error) {
	if r.listReadErr != nil {
		return nil, r.listReadErr
	}
	return append([]Footprint(nil), r.footprints...), nil
}

func (r *fakeRepository) GetCreditAccount(ctx context.Context, userID int64) (CreditAccount, bool, error) {
	account, ok := r.creditAccounts[userID]
	return account, ok, nil
}

func (r *fakeRepository) EnsureCreditAccount(ctx context.Context, userID int64, initialScore int) (CreditAccount, error) {
	if r.creditAccountErr != nil {
		return CreditAccount{}, r.creditAccountErr
	}
	account, ok := r.creditAccounts[userID]
	if !ok {
		account = CreditAccount{UserID: userID, CurrentScore: initialScore, UpdatedAt: time.Now()}
		r.creditAccounts[userID] = account
	}
	return account, nil
}

func (r *fakeRepository) SaveCreditAccount(ctx context.Context, account CreditAccount) (CreditAccount, error) {
	account.UpdatedAt = time.Now()
	r.creditAccounts[account.UserID] = account
	return account, nil
}

func (r *fakeRepository) ApplyCreditChange(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, initialScore int, scoreCap int, createdAt time.Time) (CreditLog, error) {
	if r.creditMutationErr != nil {
		return CreditLog{}, r.creditMutationErr
	}
	account, ok := r.creditAccounts[userID]
	if !ok {
		account = CreditAccount{UserID: userID, CurrentScore: initialScore}
	}
	before := account.CurrentScore
	after := before + changeValue
	if after < 0 {
		after = 0
	}
	if after > scoreCap {
		after = scoreCap
	}
	account.CurrentScore = after
	account.UpdatedAt = createdAt
	r.creditAccounts[userID] = account
	log := CreditLog{ID: r.nextCreditID, UserID: userID, GameID: gameID, ChangeValue: after - before, BeforeScore: before, AfterScore: after, Reason: reason, CreatedAt: createdAt}
	r.nextCreditID++
	r.credits = append(r.credits, log)
	r.addedCreditLog = true
	return log, nil
}

func (r *fakeRepository) ApplyCreditChangeOnce(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, idempotencyKey string, initialScore int, scoreCap int, createdAt time.Time) (CreditLog, bool, error) {
	if r.creditMutationErr != nil {
		return CreditLog{}, false, r.creditMutationErr
	}
	key := strconv.FormatInt(userID, 10) + ":" + idempotencyKey
	if log, ok := r.creditMutationKeys[key]; ok {
		return log, false, nil
	}
	log, err := r.ApplyCreditChange(ctx, userID, gameID, changeValue, reason, initialScore, scoreCap, createdAt)
	if err != nil {
		return CreditLog{}, false, err
	}
	r.creditMutationKeys[key] = log
	return log, true, nil
}

func (r *fakeRepository) RecordGrowthEvent(ctx context.Context, event GrowthEvent) (bool, error) {
	if _, exists := r.growthEvents[event.IdempotencyKey]; exists {
		return false, nil
	}
	r.growthEvents[event.IdempotencyKey] = event
	return true, nil
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
	if r.creditRuleErr != nil {
		return 0, false, r.creditRuleErr
	}
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

func (r *fakeRepository) RestoreCreditForAppeal(ctx context.Context, userID int64, gameID int64, sourceCreditLogID int64, appealID int64, amount int, initialScore int, scoreCap int, createdAt time.Time) (CreditLog, bool, error) {
	for i := range r.credits {
		source := &r.credits[i]
		if source.ID != sourceCreditLogID || source.UserID != userID || source.ChangeValue >= 0 {
			continue
		}
		if source.AppealID != 0 && source.AppealID != appealID {
			return CreditLog{}, false, ErrCreditAppealConflict
		}
		for _, existing := range r.credits {
			if existing.AppealID == appealID && existing.Reason == "appeal_passed" {
				return existing, false, nil
			}
		}
		source.AppealID = appealID
		account, ok := r.creditAccounts[userID]
		if !ok {
			account = CreditAccount{UserID: userID, CurrentScore: initialScore}
		}
		if amount < 0 {
			amount = -amount
		}
		if maximum := -source.ChangeValue; amount > maximum {
			amount = maximum
		}
		after := account.CurrentScore + amount
		if after > scoreCap {
			after = scoreCap
		}
		log := CreditLog{
			ID:          r.nextCreditID,
			UserID:      userID,
			GameID:      gameID,
			ChangeValue: after - account.CurrentScore,
			BeforeScore: account.CurrentScore,
			AfterScore:  after,
			Reason:      "appeal_passed",
			AppealID:    appealID,
			CreatedAt:   createdAt,
		}
		r.nextCreditID++
		account.CurrentScore = after
		r.creditAccounts[userID] = account
		r.credits = append(r.credits, log)
		return log, true, nil
	}
	return CreditLog{}, false, ErrCreditLogNotFound
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
