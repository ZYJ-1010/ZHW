package games

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGameRepositoryPersistsCoreFlow(t *testing.T) {
	repo := newFakeGameRepository()
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)

	game, err := service.Create(1, CreateRequest{Title: "repo game", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.createdGame || !repo.addedMember {
		t.Fatalf("expected repository create and creator member, repo=%+v", repo)
	}

	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	app, err := service.Apply(2, game.ID, ApplyRequest{Reason: "join", FileIDs: []int64{7, 8}})
	if err != nil {
		t.Fatal(err)
	}
	if len(app.FileIDs) != 2 || app.FileIDs[0] != 7 || app.FileIDs[1] != 8 {
		t.Fatalf("expected application file ids to persist, got %+v", app)
	}
	if !repo.createdApplication {
		t.Fatal("expected repository application create")
	}
	approved, err := service.ReviewApplication(1, app.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != "approved" || !repo.atomicApproval || !repo.updatedApplication || !repo.updatedGame {
		t.Fatalf("expected repository review updates, approved=%+v repo=%+v", approved, repo)
	}
	if err := service.RecordExitCredit(game.ID, 2, 11); err != nil {
		t.Fatal(err)
	}
	if members := service.Members(game.ID); len(members) != 2 || !repo.listedMembers {
		t.Fatalf("expected repository members, got %+v repo=%+v", members, repo)
	}

	invitation, err := service.CreateInvitation(1, game.ID, InvitationRequest{TargetUserID: 3, Message: "join us"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.createdInvitation || invitation.Status != "pending" {
		t.Fatalf("expected repository invitation create, invitation=%+v repo=%+v", invitation, repo)
	}
	responded, inviteApp, err := service.RespondInvitation(3, invitation.ID, InvitationRespondRequest{Accept: true, Reason: "accepted"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.updatedInvitation || responded.Status != "accepted" || responded.ApplicationID != inviteApp.ID || inviteApp.Status != "pending" {
		t.Fatalf("expected invitation response to create pending application, invitation=%+v app=%+v repo=%+v", responded, inviteApp, repo)
	}
	if freshMembers := service.Members(game.ID); len(freshMembers) != 2 {
		t.Fatalf("expected accepted invitation not to add member before audit, got %+v", freshMembers)
	}

	fresh := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	loaded, err := fresh.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != game.ID || loaded.CurrentPlayers != 2 {
		t.Fatalf("expected repository game after fresh service, got %+v", loaded)
	}
	if !fresh.IsMember(game.ID, 2) {
		t.Fatal("expected repository member lookup after fresh service")
	}
}

func TestExitWithCreditMutationUsesAtomicRepositoryCapability(t *testing.T) {
	repo := newFakeGameRepository()
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, err := service.Create(1, CreateRequest{Title: "atomic exit", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ExitWithCreditMutation(1, game.ID, "quit_after_started", "quit_after_started", ExitCreditMutation{ChangeValue: -10, Reason: "quit_after_started"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ChangeValue != -10 || result.BeforeScore != 100 || result.AfterScore != 90 || result.CreditLogID == 0 {
		t.Fatalf("unexpected atomic credit result: %+v", result)
	}
	if result.Game.CurrentPlayers != 0 || service.IsMember(game.ID, 1) {
		t.Fatalf("atomic exit did not update membership and capacity: %+v", result)
	}
}

func TestStatsForUserUsesPersistentRepositoryAfterRestart(t *testing.T) {
	repo := &statsGameRepository{
		fakeGameRepository: newFakeGameRepository(),
		stats:              UserStats{UserID: 42, Participated: 3, Completed: 2},
	}
	// 新建服务模拟进程重启；内存 games/members 缓存为空。
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)

	stats := service.StatsForUser(42)
	if stats.UserID != 42 || stats.Participated != 3 || stats.Completed != 2 {
		t.Fatalf("expected repository-backed stats after restart, got %+v", stats)
	}
}

type statsGameRepository struct {
	*fakeGameRepository
	stats                UserStats
	participatedInWindow bool
	participationErr     error
}

func (r *statsGameRepository) StatsForUser(ctx context.Context, userID int64) (UserStats, error) {
	result := r.stats
	result.UserID = userID
	return result, nil
}

func (r *statsGameRepository) ParticipatedBetween(context.Context, int64, time.Time, time.Time) (bool, error) {
	return r.participatedInWindow, r.participationErr
}

func TestParticipatedBetweenUsesMembershipTimestampRepository(t *testing.T) {
	repo := &statsGameRepository{
		fakeGameRepository:   newFakeGameRepository(),
		participatedInWindow: true,
	}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	start := time.Date(2026, 7, 31, 0, 0, 0, 0, time.Local)
	if !service.ParticipatedBetween(42, start, start.Add(24*time.Hour)) {
		t.Fatal("expected repository membership time to drive daily participation")
	}
}

func TestParticipatedBetweenStrictReturnsRepositoryFailure(t *testing.T) {
	repo := &statsGameRepository{
		fakeGameRepository: newFakeGameRepository(),
		participationErr:   errors.New("membership time database unavailable"),
	}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	if _, err := service.ParticipatedBetweenStrict(42, start, start.Add(24*time.Hour)); !errors.Is(err, repo.participationErr) {
		t.Fatalf("expected strict participation read error, got %v", err)
	}
}

func TestCreateFailsClosedWhenDailyCreationCountCannotBeRead(t *testing.T) {
	repo := newFakeGameRepository()
	repo.countTodayErr = errors.New("daily creation counter unavailable")
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	_, err := service.Create(1, CreateRequest{
		Title: "每日限额读取失败", GameType: "free", MinPlayers: 5, MaxPlayers: 8,
		StartAt: "2026-08-02 10:00", EndAt: "2026-08-02 12:00",
	})
	if !errors.Is(err, repo.countTodayErr) {
		t.Fatalf("expected daily-count repository failure, got %v", err)
	}
	if len(repo.games) != 0 {
		t.Fatalf("creation must not proceed after daily-count read failure: %+v", repo.games)
	}
}

func TestAtomicApplicationApprovalDoesNotMutateServiceCacheOnRepositoryFailure(t *testing.T) {
	repo := newFakeGameRepository()
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, err := service.Create(1, CreateRequest{Title: "审批原子性", GameType: "free", MinPlayers: 5, MaxPlayers: 5, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	application, err := service.Apply(2, game.ID, ApplyRequest{Reason: "报名"})
	if err != nil {
		t.Fatal(err)
	}
	repo.approveErr = errors.New("database unavailable")
	if _, err := service.ReviewApplication(1, application.ID, true); err == nil {
		t.Fatal("expected repository approval failure")
	}
	current, err := service.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.CurrentPlayers != 1 || current.Status != StatusRecruiting || service.IsMember(game.ID, 2) {
		t.Fatalf("failed approval must not update game capacity or membership: %+v", current)
	}
	pending := service.applications[application.ID]
	if pending.Status != "pending" {
		t.Fatalf("failed approval must keep application pending: %+v", pending)
	}
}

func TestAtomicGameCreationDoesNotLeaveGameWithoutCreator(t *testing.T) {
	repo := newFakeGameRepository()
	repo.createErr = errors.New("member insert failed")
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	_, err := service.Create(1, CreateRequest{Title: "创建原子性", GameType: "free", MinPlayers: 5, MaxPlayers: 5, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err == nil {
		t.Fatal("expected atomic create failure")
	}
	if len(repo.games) != 0 || len(service.games) != 0 {
		t.Fatalf("failed create must not leave game without creator: repo=%+v cache=%+v", repo.games, service.games)
	}
}

func TestAtomicInvitationAcceptanceDoesNotCreateOrphanApplication(t *testing.T) {
	repo := newFakeGameRepository()
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, err := service.Create(1, CreateRequest{Title: "邀请原子性", GameType: "free", MinPlayers: 5, MaxPlayers: 5, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveGame(game.ID); err != nil {
		t.Fatal(err)
	}
	invitation, err := service.CreateInvitation(1, game.ID, InvitationRequest{TargetUserID: 2, RoleType: "player"})
	if err != nil {
		t.Fatal(err)
	}
	repo.acceptInviteErr = errors.New("invitation update failed")
	if _, _, err := service.RespondInvitation(2, invitation.ID, InvitationRespondRequest{Accept: true}); err == nil {
		t.Fatal("expected atomic invitation acceptance failure")
	}
	if len(repo.applications) != 0 || len(service.applications) != 0 {
		t.Fatalf("failed acceptance must not create orphan application: repo=%+v cache=%+v", repo.applications, service.applications)
	}
	if current := repo.invitations[invitation.ID]; current.Status != "pending" || current.ApplicationID != 0 {
		t.Fatalf("failed acceptance must keep invitation pending: %+v", current)
	}
}

func TestMembershipAndGameReadsFailClosedOnRepositoryError(t *testing.T) {
	repo := newFakeGameRepository()
	repo.getErr = errors.New("game database unavailable")
	repo.listMembersErr = repo.getErr
	repo.listGamesErr = repo.getErr
	repo.listApplicationsErr = repo.getErr
	service := NewServiceWithRepositories(nil, repo, nil)
	service.games[7] = Game{ID: 7, CreatorUserID: 1, Status: StatusRecruiting}
	service.members[7] = map[int64]bool{2: true}
	service.applications[3] = Application{ID: 3, GameID: 7, UserID: 2, Status: "pending"}

	if _, err := service.Get(7); !errors.Is(err, repo.getErr) {
		t.Fatalf("expected game read error without cache fallback, got %v", err)
	}
	if service.IsMember(7, 2) {
		t.Fatal("membership read failure must not grant member access from cache")
	}
	if members := service.Members(7); len(members) != 0 {
		t.Fatalf("membership read failure must not expose cached members, got %+v", members)
	}
	if items, err := service.ListStrict(); !errors.Is(err, repo.getErr) || items != nil {
		t.Fatalf("expected strict game list error without cache fallback, items=%+v err=%v", items, err)
	}
	if items, err := service.ApplicationsForUserStrict(2); !errors.Is(err, repo.getErr) || items != nil {
		t.Fatalf("expected strict user application list error, items=%+v err=%v", items, err)
	}
	if items, err := service.ApplicationsForCreatorStrict(1); !errors.Is(err, repo.getErr) || items != nil {
		t.Fatalf("expected strict creator application list error, items=%+v err=%v", items, err)
	}
}

func TestMemberRolesAndInvitationsFailClosedOnRepositoryError(t *testing.T) {
	readErr := errors.New("membership relationship database unavailable")
	repo := &relationshipReadFailingRepository{fakeGameRepository: newFakeGameRepository(), err: readErr}
	service := NewServiceWithRepositories(nil, repo, nil)
	service.members[7] = map[int64]bool{2: true}
	service.memberRoles[7] = map[int64]string{2: "expert"}
	service.invitations[8] = Invitation{ID: 8, GameID: 7, InviterID: 1, TargetUserID: 2, Status: "pending"}

	if roles := service.MemberRoles(7); len(roles) != 0 {
		t.Fatalf("member-role read failure must not expose cached roles: %+v", roles)
	}
	if invitations := service.InvitationsForUser(2); len(invitations) != 0 {
		t.Fatalf("invitation read failure must not expose cached invitations: %+v", invitations)
	}
}

func TestRepositoryMembershipRefreshRemovesStaleMemberCache(t *testing.T) {
	repo := newFakeGameRepository()
	repo.games[7] = Game{ID: 7, CreatorUserID: 1, Status: StatusRecruiting}
	repo.members[7] = map[int64]bool{}
	service := NewServiceWithRepositories(nil, repo, nil)
	service.members[7] = map[int64]bool{2: true}

	service.mu.Lock()
	defer service.mu.Unlock()
	if service.memberLocked(7, 2) {
		t.Fatal("removed member must not retain access from process cache")
	}
	if len(service.members[7]) != 0 {
		t.Fatalf("membership cache must be refreshed from repository: %+v", service.members[7])
	}
	repo.listMembersErr = errors.New("member database unavailable")
	service.members[7] = map[int64]bool{2: true}
	if service.memberLocked(7, 2) {
		t.Fatal("membership read failure must fail closed instead of granting cached access")
	}
}

type relationshipReadFailingRepository struct {
	*fakeGameRepository
	err error
}

func (r *relationshipReadFailingRepository) ListMemberRoles(context.Context, int64) ([]MemberRole, error) {
	return nil, r.err
}

func (r *relationshipReadFailingRepository) ListInvitationsForUser(context.Context, int64) ([]Invitation, error) {
	return nil, r.err
}

type fakeGameRepository struct {
	nextGameID          int64
	nextApplicationID   int64
	nextInvitationID    int64
	games               map[int64]Game
	applications        map[int64]Application
	invitations         map[int64]Invitation
	members             map[int64]map[int64]bool
	createdGame         bool
	updatedGame         bool
	addedMember         bool
	listedMembers       bool
	createdApplication  bool
	updatedApplication  bool
	atomicApproval      bool
	approveErr          error
	acceptInviteErr     error
	createdInvitation   bool
	updatedInvitation   bool
	updatedExitCredit   bool
	createErr           error
	getErr              error
	listMembersErr      error
	listGamesErr        error
	listApplicationsErr error
	countTodayErr       error
}

func newFakeGameRepository() *fakeGameRepository {
	return &fakeGameRepository{
		nextGameID:        1,
		nextApplicationID: 1,
		nextInvitationID:  1,
		games:             make(map[int64]Game),
		applications:      make(map[int64]Application),
		invitations:       make(map[int64]Invitation),
		members:           make(map[int64]map[int64]bool),
	}
}

func (r *fakeGameRepository) CreateGame(ctx context.Context, game Game) (Game, error) {
	r.createdGame = true
	game.ID = r.nextGameID
	r.nextGameID++
	r.games[game.ID] = game
	return game, nil
}

func (r *fakeGameRepository) CreateGameWithCreator(ctx context.Context, game Game, creatorRole string) (Game, error) {
	if r.createErr != nil {
		return Game{}, r.createErr
	}
	saved, err := r.CreateGame(ctx, game)
	if err != nil {
		return Game{}, err
	}
	if err := r.AddMember(ctx, saved.ID, saved.CreatorUserID, creatorRole); err != nil {
		delete(r.games, saved.ID)
		return Game{}, err
	}
	return saved, nil
}

func (r *fakeGameRepository) UpdateGame(ctx context.Context, game Game) (Game, error) {
	r.updatedGame = true
	r.games[game.ID] = game
	return game, nil
}

func (r *fakeGameRepository) GetGame(ctx context.Context, gameID int64) (Game, error) {
	if r.getErr != nil {
		return Game{}, r.getErr
	}
	game, ok := r.games[gameID]
	if !ok {
		return Game{}, ErrGameNotFound
	}
	return game, nil
}

func (r *fakeGameRepository) ListGames(ctx context.Context) ([]Game, error) {
	if r.listGamesErr != nil {
		return nil, r.listGamesErr
	}
	result := make([]Game, 0, len(r.games))
	for _, game := range r.games {
		result = append(result, game)
	}
	return result, nil
}

func (r *fakeGameRepository) CountGamesCreatedToday(ctx context.Context, userID int64, now time.Time) (int, error) {
	if r.countTodayErr != nil {
		return 0, r.countTodayErr
	}
	count := 0
	for _, game := range r.games {
		if game.CreatorUserID == userID && sameDay(now, game.CreatedAt) {
			count++
		}
	}
	return count, nil
}

func (r *fakeGameRepository) AddMember(ctx context.Context, gameID int64, userID int64, role string) error {
	r.addedMember = true
	if r.members[gameID] == nil {
		r.members[gameID] = make(map[int64]bool)
	}
	r.members[gameID][userID] = true
	return nil
}

func (r *fakeGameRepository) DeleteMember(ctx context.Context, gameID int64, userID int64, status string, reason string) error {
	if r.members[gameID] != nil {
		delete(r.members[gameID], userID)
	}
	return nil
}

func (r *fakeGameRepository) UpdateMemberExitCredit(ctx context.Context, gameID int64, userID int64, creditDeducted bool, creditLogID int64) error {
	r.updatedExitCredit = true
	return nil
}

func (r *fakeGameRepository) ExitGameWithCredit(ctx context.Context, gameID int64, userID int64, memberStatus string, reason string, mutation ExitCreditMutation) (ExitCreditResult, error) {
	game, ok := r.games[gameID]
	if !ok {
		return ExitCreditResult{}, ErrGameNotFound
	}
	if !r.members[gameID][userID] {
		return ExitCreditResult{}, ErrForbidden
	}
	delete(r.members[gameID], userID)
	if game.CurrentPlayers > 0 {
		game.CurrentPlayers--
	}
	r.games[gameID] = game
	return ExitCreditResult{
		ExitResult:  ExitResult{Game: game, GameID: gameID, UserID: userID, Reason: reason, CreditDeduct: true, CreditDeducted: true, CreditLogID: 1, MemberStatus: memberStatus},
		CreditLogID: 1,
		BeforeScore: 100,
		AfterScore:  100 + mutation.ChangeValue,
		ChangeValue: mutation.ChangeValue,
		CreatedAt:   time.Now(),
	}, nil
}

func (r *fakeGameRepository) ListMembers(ctx context.Context, gameID int64) ([]int64, error) {
	r.listedMembers = true
	if r.listMembersErr != nil {
		return nil, r.listMembersErr
	}
	result := make([]int64, 0, len(r.members[gameID]))
	for userID := range r.members[gameID] {
		result = append(result, userID)
	}
	return result, nil
}

func (r *fakeGameRepository) CreateApplication(ctx context.Context, application Application) (Application, error) {
	r.createdApplication = true
	application.ID = r.nextApplicationID
	r.nextApplicationID++
	r.applications[application.ID] = application
	return application, nil
}

func (r *fakeGameRepository) UpdateApplication(ctx context.Context, application Application) (Application, error) {
	r.updatedApplication = true
	r.applications[application.ID] = application
	return application, nil
}

func (r *fakeGameRepository) ApproveApplication(ctx context.Context, game Game, expectedCurrentPlayers int, application Application, memberRole string, statusLog *StatusLog) (Game, Application, error) {
	if r.approveErr != nil {
		return Game{}, Application{}, r.approveErr
	}
	current, ok := r.games[game.ID]
	if !ok || current.Status != "recruiting" || current.CurrentPlayers != expectedCurrentPlayers {
		return Game{}, Application{}, ErrFull
	}
	pending, ok := r.applications[application.ID]
	if !ok || pending.Status != "pending" {
		return Game{}, Application{}, ErrApplicationNotPending
	}
	if memberRole == "" {
		memberRole = "member"
	}
	r.atomicApproval = true
	r.updatedGame = true
	r.updatedApplication = true
	r.games[game.ID] = game
	r.applications[application.ID] = application
	if r.members[game.ID] == nil {
		r.members[game.ID] = make(map[int64]bool)
	}
	r.members[game.ID][application.UserID] = true
	r.addedMember = true
	return game, application, nil
}

func (r *fakeGameRepository) GetApplication(ctx context.Context, applicationID int64) (Application, error) {
	application, ok := r.applications[applicationID]
	if !ok {
		return Application{}, ErrApplicationNotFound
	}
	return application, nil
}

func (r *fakeGameRepository) ListApplicationsByUser(ctx context.Context, userID int64) ([]Application, error) {
	if r.listApplicationsErr != nil {
		return nil, r.listApplicationsErr
	}
	result := make([]Application, 0)
	for _, application := range r.applications {
		if application.UserID == userID {
			result = append(result, application)
		}
	}
	return result, nil
}

func (r *fakeGameRepository) ListApplicationsForCreator(ctx context.Context, creatorUserID int64) ([]Application, error) {
	if r.listApplicationsErr != nil {
		return nil, r.listApplicationsErr
	}
	result := make([]Application, 0)
	for _, application := range r.applications {
		game, ok := r.games[application.GameID]
		if ok && game.CreatorUserID == creatorUserID {
			result = append(result, application)
		}
	}
	return result, nil
}

func (r *fakeGameRepository) PendingApplicationExists(ctx context.Context, gameID int64, userID int64) (bool, error) {
	for _, application := range r.applications {
		if application.GameID == gameID && application.UserID == userID && application.Status == "pending" {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeGameRepository) CreateInvitation(ctx context.Context, invitation Invitation) (Invitation, error) {
	r.createdInvitation = true
	invitation.ID = r.nextInvitationID
	r.nextInvitationID++
	r.invitations[invitation.ID] = invitation
	return invitation, nil
}

func (r *fakeGameRepository) UpdateInvitation(ctx context.Context, invitation Invitation) (Invitation, error) {
	r.updatedInvitation = true
	r.invitations[invitation.ID] = invitation
	return invitation, nil
}

func (r *fakeGameRepository) AcceptInvitation(ctx context.Context, invitation Invitation, application Application) (Invitation, Application, error) {
	if r.acceptInviteErr != nil {
		return Invitation{}, Application{}, r.acceptInviteErr
	}
	savedApplication, err := r.CreateApplication(ctx, application)
	if err != nil {
		return Invitation{}, Application{}, err
	}
	invitation.ApplicationID = savedApplication.ID
	savedInvitation, err := r.UpdateInvitation(ctx, invitation)
	if err != nil {
		delete(r.applications, savedApplication.ID)
		return Invitation{}, Application{}, err
	}
	return savedInvitation, savedApplication, nil
}

func (r *fakeGameRepository) GetInvitation(ctx context.Context, invitationID int64) (Invitation, error) {
	invitation, ok := r.invitations[invitationID]
	if !ok {
		return Invitation{}, ErrInvitationNotFound
	}
	return invitation, nil
}
