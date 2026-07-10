package games

import "testing"

func newVerifiedGameService() *Service {
	return NewService(fakeIdentity{verified: true})
}

func mustCreateRecruitingGame(t *testing.T, service *Service) Game {
	t.Helper()
	game, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8})
	if err != nil {
		t.Fatalf("create game failed: %v", err)
	}
	game, err = service.ApproveGame(game.ID)
	if err != nil {
		t.Fatalf("approve game failed: %v", err)
	}
	return game
}

func mustCreateStartedGame(t *testing.T, service *Service) (Game, []int64) {
	t.Helper()
	game := mustCreateRecruitingGame(t, service)
	members := []int64{1, 2, 3, 4, 5}
	mustApproveMembers(t, service, game.ID, members[1:]...)
	started, err := service.ManualStart(1, game.ID)
	if err != nil {
		t.Fatalf("manual start failed: %v", err)
	}
	return started, members
}

func mustApproveMembers(t *testing.T, service *Service, gameID int64, userIDs ...int64) {
	t.Helper()
	for _, userID := range userIDs {
		app, err := service.Apply(userID, gameID, ApplyRequest{Reason: "join"})
		if err != nil {
			t.Fatalf("apply user %d failed: %v", userID, err)
		}
		if _, err := service.ReviewApplication(1, app.ID, true); err != nil {
			t.Fatalf("approve application for user %d failed: %v", userID, err)
		}
	}
}

func mustInviteAndApproveMember(t *testing.T, service *Service, inviterID int64, gameID int64, targetUserID int64) Invitation {
	t.Helper()
	invitation, err := service.CreateInvitation(inviterID, gameID, InvitationRequest{TargetUserID: targetUserID, Message: "join"})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	invitation, app, err := service.RespondInvitation(targetUserID, invitation.ID, InvitationRespondRequest{Accept: true, Reason: "accept"})
	if err != nil {
		t.Fatalf("respond invitation failed: %v", err)
	}
	if _, err := service.ReviewApplication(1, app.ID, true); err != nil {
		t.Fatalf("approve invited application failed: %v", err)
	}
	return invitation
}

func mustCreateStartedGameWithInvitedMainGuide(t *testing.T, service *Service) (Game, []int64) {
	t.Helper()
	game := mustCreateRecruitingGame(t, service)
	members := []int64{1, 2, 3, 4, 5}
	mustInviteAndApproveMember(t, service, 1, game.ID, 2)
	mustApproveMembers(t, service, game.ID, members[2:]...)
	started, err := service.ManualStart(1, game.ID)
	if err != nil {
		t.Fatalf("manual start failed: %v", err)
	}
	return started, members
}

func mustMoveGameToPendingReview(t *testing.T, service *Service, gameID int64, members []int64) Game {
	t.Helper()
	var game Game
	for _, userID := range members {
		_, _, nextGame, err := service.ConfirmService(userID, gameID, "ok")
		if err != nil {
			t.Fatalf("confirm service for user %d failed: %v", userID, err)
		}
		game = nextGame
	}
	if game.Status != "pending_review" {
		t.Fatalf("expected pending_review, got %s", game.Status)
	}
	return game
}
