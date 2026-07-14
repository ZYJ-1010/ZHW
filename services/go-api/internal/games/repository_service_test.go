package games

import (
	"context"
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
	if approved.Status != "approved" || !repo.updatedApplication || !repo.updatedGame {
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

type fakeGameRepository struct {
	nextGameID         int64
	nextApplicationID  int64
	nextInvitationID   int64
	games              map[int64]Game
	applications       map[int64]Application
	invitations        map[int64]Invitation
	members            map[int64]map[int64]bool
	createdGame        bool
	updatedGame        bool
	addedMember        bool
	listedMembers      bool
	createdApplication bool
	updatedApplication bool
	createdInvitation  bool
	updatedInvitation  bool
	updatedExitCredit  bool
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

func (r *fakeGameRepository) UpdateGame(ctx context.Context, game Game) (Game, error) {
	r.updatedGame = true
	r.games[game.ID] = game
	return game, nil
}

func (r *fakeGameRepository) GetGame(ctx context.Context, gameID int64) (Game, error) {
	game, ok := r.games[gameID]
	if !ok {
		return Game{}, ErrGameNotFound
	}
	return game, nil
}

func (r *fakeGameRepository) ListGames(ctx context.Context) ([]Game, error) {
	result := make([]Game, 0, len(r.games))
	for _, game := range r.games {
		result = append(result, game)
	}
	return result, nil
}

func (r *fakeGameRepository) CountGamesCreatedToday(ctx context.Context, userID int64, now time.Time) (int, error) {
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

func (r *fakeGameRepository) ListMembers(ctx context.Context, gameID int64) ([]int64, error) {
	r.listedMembers = true
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

func (r *fakeGameRepository) GetApplication(ctx context.Context, applicationID int64) (Application, error) {
	application, ok := r.applications[applicationID]
	if !ok {
		return Application{}, ErrApplicationNotFound
	}
	return application, nil
}

func (r *fakeGameRepository) ListApplicationsByUser(ctx context.Context, userID int64) ([]Application, error) {
	result := make([]Application, 0)
	for _, application := range r.applications {
		if application.UserID == userID {
			result = append(result, application)
		}
	}
	return result, nil
}

func (r *fakeGameRepository) ListApplicationsForCreator(ctx context.Context, creatorUserID int64) ([]Application, error) {
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

func (r *fakeGameRepository) GetInvitation(ctx context.Context, invitationID int64) (Invitation, error) {
	invitation, ok := r.invitations[invitationID]
	if !ok {
		return Invitation{}, ErrInvitationNotFound
	}
	return invitation, nil
}
