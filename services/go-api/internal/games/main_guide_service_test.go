package games

import "testing"

func TestManualStartRequiresMinPlayers(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)
	mustApproveMembers(t, service, game.ID, 2, 3, 4)

	if _, err := service.ManualStart(1, game.ID); err != ErrGameNotStartable {
		t.Fatalf("expected ErrGameNotStartable below min players, got %v", err)
	}

	mustApproveMembers(t, service, game.ID, 5)
	started, err := service.ManualStart(1, game.ID)
	if err != nil {
		t.Fatalf("manual start with min players failed: %v", err)
	}
	if started.Status != "in_progress" {
		t.Fatalf("expected in_progress, got %+v", started)
	}
}

func TestCreateRejectsPlayersOutsideFiveToEight(t *testing.T) {
	service := newVerifiedGameService()
	cases := []CreateRequest{
		{Title: "too few", GameType: "free", MinPlayers: 4, MaxPlayers: 8},
		{Title: "too many", GameType: "free", MinPlayers: 5, MaxPlayers: 9},
		{Title: "invalid range", GameType: "free", MinPlayers: 6, MaxPlayers: 5},
	}

	for _, req := range cases {
		if _, err := service.Create(1, req); err != ErrInvalidPlayers {
			t.Fatalf("expected ErrInvalidPlayers for %+v, got %v", req, err)
		}
	}
}

func TestReviewApplicationRejectsMemberAboveMaxPlayers(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)
	mustApproveMembers(t, service, game.ID, 2, 3, 4, 5, 6, 7, 8)

	fullGame, err := service.Get(game.ID)
	if err != nil {
		t.Fatalf("get full game failed: %v", err)
	}
	if fullGame.CurrentPlayers != MaxGamePlayers || fullGame.Status != "full" {
		t.Fatalf("expected full game at max players, got %+v", fullGame)
	}
	if app, err := service.Apply(9, game.ID, ApplyRequest{Reason: "join"}); err != ErrGameNotRecruiting {
		t.Fatalf("expected ErrGameNotRecruiting after full game closes recruiting, got app=%+v err=%v", app, err)
	}
}

func TestInvitedMainGuideCanManageProgress(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGameWithInvitedMainGuide(t, service)

	if game.MainGuideUserID != 2 {
		t.Fatalf("expected invited user 2 as main guide, got %+v", game)
	}
	if _, err := service.AddProgressFeedback(3, game.ID, ProgressFeedbackRequest{Progress: 20, Content: "member update"}); err != ErrForbidden {
		t.Fatalf("expected member progress forbidden, got %v", err)
	}
	feedback, err := service.AddProgressFeedback(2, game.ID, ProgressFeedbackRequest{Progress: 30, Content: "main guide update"})
	if err != nil {
		t.Fatalf("main guide progress failed: %v", err)
	}
	if feedback.Progress != 30 || feedback.UserID != 2 {
		t.Fatalf("expected main guide progress feedback, got %+v", feedback)
	}
}
