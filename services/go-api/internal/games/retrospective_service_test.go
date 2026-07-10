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
