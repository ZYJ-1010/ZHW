package games

import "testing"

func TestRetrospectiveRequiresMemberAndReviewableGame(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGame(t, service)

	if _, err := service.CreateRetrospective(99, game.ID, RetrospectiveRequest{Content: "done"}); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden for non-member, got %v", err)
	}
	if _, err := service.CreateRetrospective(2, game.ID, RetrospectiveRequest{Content: "done"}); err != ErrGameNotConfirmable {
		t.Fatalf("expected ErrGameNotConfirmable before confirmation, got %v", err)
	}
}

func TestRetrospectiveRejectsDuplicateSubmission(t *testing.T) {
	service := newVerifiedGameService()
	game, members := mustCreateStartedGame(t, service)
	mustMoveGameToPendingReview(t, service, game.ID, members)

	if _, err := service.CreateRetrospective(2, game.ID, RetrospectiveRequest{Content: "done", AgainIntent: "yes"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateRetrospective(2, game.ID, RetrospectiveRequest{Content: "again", AgainIntent: "yes"}); err != ErrDuplicateRetrospective {
		t.Fatalf("expected ErrDuplicateRetrospective, got %v", err)
	}
}

func TestContinueDraftOnlyCreatesDraftGame(t *testing.T) {
	service := newVerifiedGameService()
	game, members := mustCreateStartedGame(t, service)
	reviewable := mustMoveGameToPendingReview(t, service, game.ID, members)

	continued, err := service.ContinueDraft(1, game.ID, ContinueDraftRequest{Title: "续局草稿"})
	if err != nil {
		t.Fatal(err)
	}
	if continued.Draft.Status != "draft" || continued.Draft.GameSource != "continue" {
		t.Fatalf("expected continue draft, got %+v", continued.Draft)
	}
	original, err := service.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if original.Status != reviewable.Status {
		t.Fatalf("expected original game unchanged, got before=%s after=%s", reviewable.Status, original.Status)
	}
	records, err := service.AdminContinueDrafts(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].DraftGameID != continued.Draft.ID {
		t.Fatalf("expected continue draft record, got %+v", records)
	}
}

func TestCreateReplayGameCreatesRecruitingGameAndInvitesFormerMembers(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGameWithInvitedMainGuide(t, service)
	price := 88.5

	result, err := service.CreateReplayGame(2, game.ID, ReplayGameRequest{
		PlayerUserIDs: []int64{1},
		ExpertUserIDs: []int64{3},
		Message:       "again",
		Overrides: ReplayGameOverrides{
			Title: "next game",
			Price: &price,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ReplayGameID == game.ID || result.Game.Status != "recruiting" || result.Game.GameSource != "replay" {
		t.Fatalf("expected independent recruiting replay game, got %+v", result.Game)
	}
	if result.Game.CreatorUserID != 2 || result.Game.Title != "next game" || result.Game.Price != price {
		t.Fatalf("expected replay overrides and current user as creator, got %+v", result.Game)
	}
	if result.InvitationCount != 2 || len(result.Invitations) != 2 || result.PrimaryInvitationID == 0 {
		t.Fatalf("expected two replay invitations, got %+v", result)
	}
	roles := map[int64]string{}
	for _, invitation := range result.Invitations {
		if invitation.GameID != result.ReplayGameID {
			t.Fatalf("expected invitation on replay game, got %+v", invitation)
		}
		roles[invitation.TargetUserID] = invitation.Role
	}
	if roles[1] != "player" || roles[3] != "expert" {
		t.Fatalf("expected preserved replay roles, got %+v", roles)
	}
	original, err := service.Get(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if original.Status != game.Status {
		t.Fatalf("expected original game unchanged, got before=%s after=%s", game.Status, original.Status)
	}
}

func TestCreateReplayGameRequiresCreatorOrMainGuide(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGameWithInvitedMainGuide(t, service)
	request := ReplayGameRequest{
		PlayerUserIDs: []int64{4},
		Message:       "again",
	}

	for _, userID := range []int64{3, 99} {
		if _, err := service.CreateReplayGame(userID, game.ID, request); err != ErrForbidden {
			t.Fatalf("expected ErrForbidden for user %d, got %v", userID, err)
		}
	}
}
