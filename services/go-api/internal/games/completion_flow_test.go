package games

import (
	"errors"
	"testing"
)

func TestRequestCompletionAdminGameGoesDirectlyToReview(t *testing.T) {
	service := newVerifiedGameService()
	game, members := mustCreateStartedGame(t, service)
	service.mu.Lock()
	game.GameSource = "admin"
	service.games[game.ID] = game
	service.mu.Unlock()

	updated, err := service.RequestCompletion(members[0], game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "pending_review" {
		t.Fatalf("admin game status = %q, want pending_review", updated.Status)
	}
	repeated, err := service.RequestCompletion(members[0], game.ID)
	if err != nil {
		t.Fatalf("repeated admin completion request: %v", err)
	}
	if repeated.Status != "pending_review" {
		t.Fatalf("repeated admin game status = %q, want pending_review", repeated.Status)
	}
	if _, _, _, err := service.ConfirmService(members[0], game.ID, "should not confirm"); !errors.Is(err, ErrGameNotConfirmable) {
		t.Fatalf("admin game confirmation error = %v, want ErrGameNotConfirmable", err)
	}
}

func TestAppGameRequiresExpertsBeforePlayersAndIgnoresGuide(t *testing.T) {
	service := newVerifiedGameService()
	game, members := mustCreateStartedGame(t, service)
	service.mu.Lock()
	service.memberRoles[game.ID][members[1]] = "expert"
	service.memberRoles[game.ID][members[4]] = "guide"
	service.mu.Unlock()

	updated, err := service.RequestCompletion(members[0], game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "pending_confirm" {
		t.Fatalf("app game status = %q, want pending_confirm", updated.Status)
	}
	repeated, err := service.RequestCompletion(members[0], game.ID)
	if err != nil {
		t.Fatalf("repeated app completion request: %v", err)
	}
	if repeated.Status != "pending_confirm" {
		t.Fatalf("repeated app game status = %q, want pending_confirm", repeated.Status)
	}
	if _, _, _, err := service.ConfirmService(members[0], game.ID, "player first"); !errors.Is(err, ErrExpertConfirmRequired) {
		t.Fatalf("player-first error = %v, want ErrExpertConfirmRequired", err)
	}
	if _, _, _, err := service.ConfirmService(members[4], game.ID, "guide confirm"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("guide error = %v, want ErrForbidden", err)
	}

	confirm, _, updated, err := service.ConfirmService(members[1], game.ID, "expert confirmed")
	if err != nil {
		t.Fatal(err)
	}
	if confirm.Status != "pending" || updated.Status != "pending_confirm" {
		t.Fatalf("after expert confirm=%+v game=%+v", confirm, updated)
	}
	for _, playerID := range []int64{members[0], members[2], members[3]} {
		confirm, _, updated, err = service.ConfirmService(playerID, game.ID, "player confirmed")
		if err != nil {
			t.Fatal(err)
		}
	}
	if confirm.Status != "completed" || updated.Status != "pending_review" {
		t.Fatalf("final confirm=%+v game=%+v", confirm, updated)
	}
}

func TestGroupedInvitationCompletionOnlyAllowsBoundPlayer(t *testing.T) {
	service := newVerifiedGameService()
	game, members := mustCreateStartedGame(t, service)
	expertID := members[1]
	boundPlayerID := members[2]
	unboundPlayerID := members[3]
	service.mu.Lock()
	service.memberRoles[game.ID][expertID] = "expert"
	service.invitations[101] = Invitation{
		ID: 101, GameID: game.ID, InviterID: members[0], TargetUserID: expertID,
		PlayerUserID: boundPlayerID, ExpertUserID: expertID, InviteGroupID: "pair-1",
		Role: "expert", Status: "accepted",
	}
	service.mu.Unlock()

	if _, err := service.RequestCompletion(members[0], game.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := service.ConfirmService(expertID, game.ID, "expert confirmed"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := service.ConfirmService(unboundPlayerID, game.ID, "unbound player"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("unbound player error = %v, want ErrForbidden", err)
	}
	confirm, _, updated, err := service.ConfirmService(boundPlayerID, game.ID, "bound player confirmed")
	if err != nil {
		t.Fatal(err)
	}
	if confirm.Status != "completed" || updated.Status != "pending_review" {
		t.Fatalf("bound confirmation did not complete service: confirm=%+v game=%+v", confirm, updated)
	}
}
