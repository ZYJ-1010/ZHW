package games

import (
	"testing"
	"time"
)

type fakeIdentity struct {
	verified  bool
	byUserIDs map[int64]bool
}

type recordingRoomEnsurer struct {
	gameIDs []int64
}

func (r *recordingRoomEnsurer) EnsureRoom(gameID int64) {
	r.gameIDs = append(r.gameIDs, gameID)
}

func (f fakeIdentity) IsVerified(userID int64) bool {
	if f.byUserIDs != nil {
		return f.byUserIDs[userID]
	}
	return f.verified
}

func TestCreateAllowsPlayerWithoutVerifiedIdentity(t *testing.T) {
	service := NewService(fakeIdentity{verified: false})
	game, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatalf("expected player create without identity, got %v", err)
	}
	if game.ID == 0 || game.CurrentPlayers != 1 {
		t.Fatalf("expected created game, got %+v", game)
	}
}

func TestCreateUsesConfiguredPlayerLimits(t *testing.T) {
	service := NewService(fakeIdentity{verified: false})
	service.SetPlayerLimits(4, 6)
	if _, err := service.Create(1, CreateRequest{Title: "配置人数局", GameType: "free", MinPlayers: 5, MaxPlayers: 6, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"}); err != nil {
		t.Fatalf("expected configured min player limit to allow request, got %v", err)
	}
	if _, err := service.Create(2, CreateRequest{Title: "超出人数局", GameType: "free", MinPlayers: 4, MaxPlayers: 7, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"}); err != ErrInvalidPlayers {
		t.Fatalf("expected configured max player limit to reject request, got %v", err)
	}
}

func TestReviewApplicationCreatesRoomWhenGameBecomesFull(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	ensurer := &recordingRoomEnsurer{}
	service.UseRoomEnsurer(ensurer)
	game, err := service.Create(1, CreateRequest{Title: "满员建房", GameType: "free", MinPlayers: 5, MaxPlayers: 5, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []int64{2, 3, 4, 5} {
		app, applyErr := service.Apply(userID, game.ID, ApplyRequest{Reason: "join"})
		if applyErr != nil {
			t.Fatal(applyErr)
		}
		if _, reviewErr := service.ReviewApplication(1, app.ID, true); reviewErr != nil {
			t.Fatal(reviewErr)
		}
	}
	if len(ensurer.gameIDs) != 1 || ensurer.gameIDs[0] != game.ID {
		t.Fatalf("expected one room ensure for full game %d, got %+v", game.ID, ensurer.gameIDs)
	}
}

func TestCreateRequiresValidGameTime(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	base := CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8}
	if _, err := service.Create(1, base); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput, got %v", err)
	}
	base.StartAt = "2026-07-12 16:00"
	base.EndAt = "2026-07-12 14:00"
	if _, err := service.Create(1, base); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput, got %v", err)
	}
}

func TestCreateAllowsOnlyFreeGameFromApp(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	_, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "paid", MinPlayers: 5, MaxPlayers: 8})
	if err != ErrInvalidGameType {
		t.Fatalf("expected ErrInvalidGameType, got %v", err)
	}
}

func TestCreateFromAppIgnoresMainGuideUserID(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{Title: "app free", GameType: "free", MainGuideUserID: 2, MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatalf("app create failed: %v", err)
	}
	if game.MainGuideUserID != 0 {
		t.Fatalf("app create should not set main guide directly, got %+v", game)
	}
}

func TestCreateKeepsTrimmedAddress(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{Title: "address game", GameType: "free", MinPlayers: 5, MaxPlayers: 8, Address: "  北京市朝阳区测试地点  ", StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if game.Address != "北京市朝阳区测试地点" {
		t.Fatalf("expected trimmed address, got %q", game.Address)
	}

	longAddress := make([]rune, 256)
	for i := range longAddress {
		longAddress[i] = 'a'
	}
	if _, err := service.Create(2, CreateRequest{Title: "long address", GameType: "free", MinPlayers: 5, MaxPlayers: 8, Address: string(longAddress)}); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput for long address, got %v", err)
	}
}

func TestCreateEnforcesDailyLimit(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	for i := 0; i < 3; i++ {
		if _, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"}); err != nil {
			t.Fatalf("create %d failed: %v", i, err)
		}
	}
	_, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != ErrDailyLimit {
		t.Fatalf("expected ErrDailyLimit, got %v", err)
	}
}

func TestCreateFromAdminAllowsDocumentedGameTypes(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})

	adminTypes := []string{"free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition"}
	for i, gameType := range adminTypes {
		creatorUserID := int64(11 + i)
		game, err := service.CreateFromAdmin(CreateRequest{Title: "admin " + gameType, CreatorUserID: creatorUserID, GameType: gameType, MinPlayers: 5, MaxPlayers: 8, SignupStartAt: "2026-07-10 09:00", SignupEndAt: "2026-07-12 14:00", StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
		if err != nil {
			t.Fatalf("admin %s create failed: %v", gameType, err)
		}
		if game.GameType != gameType || game.GameSource != "admin" || game.Status != "recruiting" || game.CreatorUserID != creatorUserID {
			t.Fatalf("unexpected %s game: %+v", gameType, game)
		}
	}

	if _, err := service.CreateFromAdmin(CreateRequest{Title: "admin condition guide", CreatorUserID: 31, MainGuideUserID: 41, GameType: "condition", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameInput {
		t.Fatalf("expected admin game to reject non-player preset role, got %v", err)
	}
}

func TestCreateFromAdminRejectsInvalidTypeAndCreator(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})

	if _, err := service.CreateFromAdmin(CreateRequest{Title: "missing creator", GameType: "standard", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput, got %v", err)
	}
	if _, err := service.CreateFromAdmin(CreateRequest{Title: "invalid type", CreatorUserID: 11, GameType: "paid", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameType {
		t.Fatalf("expected ErrInvalidGameType, got %v", err)
	}
	if _, err := service.CreateFromAdmin(CreateRequest{Title: "same guide", CreatorUserID: 11, MainGuideUserID: 11, GameType: "condition", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput for creator as main guide, got %v", err)
	}
}

func TestCreateFromAdminRequiresValidGameTime(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	base := CreateRequest{Title: "admin timed game", CreatorUserID: 11, GameType: "free", MinPlayers: 5, MaxPlayers: 8}
	if _, err := service.CreateFromAdmin(base); err != ErrInvalidGameInput {
		t.Fatalf("expected missing schedule to fail, got %v", err)
	}
	base.StartAt = "2026-07-12 16:00"
	base.EndAt = "2026-07-12 14:00"
	if _, err := service.CreateFromAdmin(base); err != ErrInvalidGameInput {
		t.Fatalf("expected reversed schedule to fail, got %v", err)
	}
}

func TestCreateFromAdminRequiresValidSignupTime(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	base := CreateRequest{
		Title: "admin signup time", CreatorUserID: 11, GameType: "free", MinPlayers: 5, MaxPlayers: 8,
		StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00",
	}
	if _, err := service.CreateFromAdmin(base); err != ErrInvalidGameInput {
		t.Fatalf("expected missing signup time to fail, got %v", err)
	}

	base.SignupStartAt = "2026-07-11 10:00"
	base.SignupEndAt = "2026-07-11 09:00"
	if _, err := service.CreateFromAdmin(base); err != ErrInvalidGameInput {
		t.Fatalf("expected reversed signup time to fail, got %v", err)
	}

	base.SignupEndAt = "2026-07-12 15:00"
	if _, err := service.CreateFromAdmin(base); err != ErrInvalidGameInput {
		t.Fatalf("expected signup deadline after game start to fail, got %v", err)
	}

	base.SignupEndAt = base.StartAt
	game, err := service.CreateFromAdmin(base)
	if err != nil {
		t.Fatalf("expected signup deadline at game start to succeed, got %v", err)
	}
	if game.SignupStartAt != base.SignupStartAt || game.SignupEndAt != base.SignupEndAt {
		t.Fatalf("expected signup times to persist, got %+v", game)
	}
}

func TestCreateFromAdminRequiresVerifiedMainGuide(t *testing.T) {
	service := NewService(fakeIdentity{byUserIDs: map[int64]bool{11: true, 22: false}})

	if _, err := service.CreateFromAdmin(CreateRequest{Title: "unverified main guide", CreatorUserID: 11, MainGuideUserID: 22, GameType: "condition", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameInput {
		t.Fatalf("expected admin game to reject main guide role, got %v", err)
	}
}

func TestAdminGameForcesApprovedEntrantsToPlayer(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	format := func(value time.Time) string { return value.Format("2006-01-02 15:04") }
	now := time.Now()
	game, err := service.CreateFromAdmin(CreateRequest{
		Title:         "admin players only",
		CreatorUserID: 1,
		GameType:      "free",
		MinPlayers:    5,
		MaxPlayers:    8,
		SignupStartAt: format(now.Add(-time.Hour)),
		SignupEndAt:   format(now.Add(time.Hour)),
		StartAt:       format(now.Add(2 * time.Hour)),
		EndAt:         format(now.Add(3 * time.Hour)),
	})
	if err != nil {
		t.Fatal(err)
	}
	application, err := service.Apply(2, game.ID, ApplyRequest{RoleType: "expert", Reason: "join as expert"})
	if err != nil {
		t.Fatal(err)
	}
	if application.Role != "member" {
		t.Fatalf("expected admin game application role member, got %q", application.Role)
	}
	if _, err := service.ReviewApplication(1, application.ID, true); err != nil {
		t.Fatal(err)
	}
	roles := service.MemberRoles(game.ID)
	for _, item := range roles {
		if item.Role != "member" {
			t.Fatalf("admin game must contain players only, got %+v", roles)
		}
	}
	invitation, err := service.CreateInvitation(1, game.ID, InvitationRequest{TargetUserID: 3, RoleType: "guide", Message: "join as guide"})
	if err != nil {
		t.Fatal(err)
	}
	if invitation.Role != "member" {
		t.Fatalf("expected admin invitation role member, got %q", invitation.Role)
	}
	_, invitedApplication, err := service.RespondInvitation(3, invitation.ID, InvitationRespondRequest{Accept: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReviewApplication(1, invitedApplication.ID, true); err != nil {
		t.Fatal(err)
	}
	for _, item := range service.MemberRoles(game.ID) {
		if item.Role != "member" {
			t.Fatalf("admin game invitation must still create players only, got %+v", service.MemberRoles(game.ID))
		}
	}
}

func TestRespondInvitationAllowsUnverifiedUserDuringPhaseOne(t *testing.T) {
	service := NewService(fakeIdentity{byUserIDs: map[int64]bool{1: true, 2: false}})
	game, err := service.Create(1, CreateRequest{Title: "invite without realname", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	invitation, err := service.CreateInvitation(1, game.ID, InvitationRequest{TargetUserID: 2, RoleType: "player", Message: "join"})
	if err != nil {
		t.Fatal(err)
	}
	responded, app, err := service.RespondInvitation(2, invitation.ID, InvitationRespondRequest{Accept: true})
	if err != nil {
		t.Fatalf("expected phase-one invitation accept without verified identity, got %v", err)
	}
	if responded.Status != "accepted" || app.Status != "pending" || app.UserID != 2 {
		t.Fatalf("expected pending application from accepted invitation, got invitation=%+v app=%+v", responded, app)
	}
}

func TestAppGameKeepsSelectedEntryRole(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{Title: "app role game", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	for userID, role := range map[int64]string{2: "expert", 3: "guide", 4: "player"} {
		application, err := service.Apply(userID, game.ID, ApplyRequest{RoleType: role, Reason: "join"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.ReviewApplication(1, application.ID, true); err != nil {
			t.Fatal(err)
		}
	}
	roleByUser := map[int64]string{}
	for _, item := range service.MemberRoles(game.ID) {
		roleByUser[item.UserID] = item.Role
	}
	if roleByUser[1] != "member" || roleByUser[2] != "expert" || roleByUser[3] != "main_guide" || roleByUser[4] != "member" {
		t.Fatalf("unexpected app game roles: %+v", roleByUser)
	}
}

func TestApplyRejectsBeforeSignupWindow(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	format := func(value time.Time) string { return value.Format("2006-01-02 15:04") }
	start := time.Now().Add(4 * time.Hour)
	game, err := service.Create(1, CreateRequest{
		Title:         "future signup",
		GameType:      "free",
		MinPlayers:    5,
		MaxPlayers:    8,
		SignupStartAt: format(start),
		SignupEndAt:   format(start.Add(30 * time.Minute)),
		StartAt:       format(start.Add(time.Hour)),
		EndAt:         format(start.Add(2 * time.Hour)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Apply(2, game.ID, ApplyRequest{Reason: "join"}); err != ErrSignupClosed {
		t.Fatalf("expected ErrSignupClosed before signup window, got %v", err)
	}
}

func TestRejectedApplicationsDoNotConsumePlayerCapacity(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	format := func(value time.Time) string { return value.Format("2006-01-02 15:04") }
	now := time.Now()
	game, err := service.Create(1, CreateRequest{
		Title:         "capacity by members",
		GameType:      "free",
		MinPlayers:    5,
		MaxPlayers:    5,
		SignupStartAt: format(now.Add(-time.Hour)),
		SignupEndAt:   format(now.Add(time.Hour)),
		StartAt:       format(now.Add(2 * time.Hour)),
		EndAt:         format(now.Add(4 * time.Hour)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if game, err = service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}

	first, err := service.Apply(2, game.ID, ApplyRequest{Reason: "first"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.ReviewApplication(1, first.ID, false); err != nil {
		t.Fatal(err)
	}
	afterReject, err := service.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if afterReject.CurrentPlayers != 1 || afterReject.Status != "recruiting" {
		t.Fatalf("rejected application must not consume capacity: %+v", afterReject)
	}

	second, err := service.Apply(2, game.ID, ApplyRequest{Reason: "retry"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.ReviewApplication(1, second.ID, true); err != nil {
		t.Fatal(err)
	}
	for _, userID := range []int64{3, 4} {
		app, applyErr := service.Apply(userID, game.ID, ApplyRequest{Reason: "join"})
		if applyErr != nil {
			t.Fatal(applyErr)
		}
		if _, reviewErr := service.ReviewApplication(1, app.ID, true); reviewErr != nil {
			t.Fatal(reviewErr)
		}
	}

	// 创建者 + A/B/C = 4 人，D 仍可报名；申请次数（含 A 的 rejected 记录）不应占位。
	d, err := service.Apply(5, game.ID, ApplyRequest{Reason: "join"})
	if err != nil {
		t.Fatalf("fourth applicant must still be allowed with one slot left: %v", err)
	}
	current, err := service.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.CurrentPlayers != 4 || current.Status != "recruiting" {
		t.Fatalf("pending application must not consume capacity: %+v", current)
	}
	if _, err = service.ReviewApplication(1, d.ID, true); err != nil {
		t.Fatal(err)
	}
	full, err := service.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if full.CurrentPlayers != 5 || full.Status != "full" {
		t.Fatalf("only approved members should fill the game: %+v", full)
	}
	if _, err = service.Apply(6, game.ID, ApplyRequest{Reason: "too late"}); err != ErrGameNotRecruiting {
		t.Fatalf("full game should reject new applications, got %v", err)
	}
}

func TestSignupWindowUsesShanghaiLocalTime(t *testing.T) {
	game := Game{
		SignupStartAt: "2026-07-15 22:00",
		SignupEndAt:   "2026-07-15 23:00",
	}
	if CanApplyWithinSignupWindow(game, time.Date(2026, 7, 15, 13, 59, 0, 0, time.UTC)) {
		t.Fatal("signup window must stay closed before Beijing signup start")
	}
	if !CanApplyWithinSignupWindow(game, time.Date(2026, 7, 15, 14, 30, 0, 0, time.UTC)) {
		t.Fatal("signup window must open when Beijing signup start has passed")
	}
	if CanApplyWithinSignupWindow(game, time.Date(2026, 7, 15, 15, 1, 0, 0, time.UTC)) {
		t.Fatal("signup window must close after Beijing signup end")
	}
}

func TestApplicationCanBeRejectedMoreThanOnce(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{Title: "repeat reject", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}

	for attempt := 1; attempt <= 2; attempt++ {
		application, err := service.Apply(2, game.ID, ApplyRequest{Reason: "join"})
		if err != nil {
			t.Fatalf("attempt %d apply failed: %v", attempt, err)
		}
		rejected, err := service.ReviewApplication(1, application.ID, false)
		if err != nil {
			t.Fatalf("attempt %d reject failed: %v", attempt, err)
		}
		if rejected.Status != "rejected" {
			t.Fatalf("attempt %d status = %q, want rejected", attempt, rejected.Status)
		}
	}
}

func TestCanceledGameKeepsHistoricalMemberRoles(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)
	application, err := service.Apply(2, game.ID, ApplyRequest{RoleType: "expert", Reason: "join"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReviewApplication(1, application.ID, true); err != nil {
		t.Fatal(err)
	}
	mustApproveMembers(t, service, game.ID, 3, 4, 5)
	game, err = service.ManualStart(1, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CancelService(game.ID, "canceled"); err != nil {
		t.Fatal(err)
	}
	roleByUser := map[int64]string{}
	for _, item := range service.MemberRoles(game.ID) {
		roleByUser[item.UserID] = item.Role
	}
	if roleByUser[1] != "member" || roleByUser[2] != "expert" {
		t.Fatalf("expected canceled game roles to remain queryable, got %+v", roleByUser)
	}
}

func TestCancelServiceRequiresInProgressGame(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)

	if _, err := service.CancelService(game.ID, "canceled"); err != ErrGameNotConfirmable {
		t.Fatalf("expected ErrGameNotConfirmable before start, got %v", err)
	}
}

func TestCreateUsesConfiguredDailyLimit(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	service.SetDailyCreateLimit(1)

	if _, err := service.Create(1, CreateRequest{Title: "configured limit", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if _, err := service.Create(1, CreateRequest{Title: "configured limit", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"}); err != ErrDailyLimit {
		t.Fatalf("expected ErrDailyLimit from configured limit, got %v", err)
	}
}

func TestCreatePersistsCategoryFields(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{
		Title:                 "category game",
		GameType:              "free",
		PrimaryCategory:       "task",
		PrimaryCategoryText:   "Task",
		SecondaryCategory:     "project",
		SecondaryCategoryText: "Project",
		Type:                  "free",
		MinPlayers:            5,
		MaxPlayers:            8,
		StartAt:               "2026-07-12 14:00",
		EndAt:                 "2026-07-12 16:00",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if game.PrimaryCategory != "task" || game.SecondaryCategory != "project" || game.Type != "free" {
		t.Fatalf("expected category fields to persist, got %+v", game)
	}
}
