package games

import "testing"

func TestGroupedInvitationRequiresPlayerBeforeExpertAndStillNeedsFinalAudit(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)

	playerInvitation, err := service.CreateInvitation(1, game.ID, InvitationRequest{
		TargetUserID:  2,
		ExpertUserID:  3,
		InviteGroupID: "grouped-confirmation",
		RoleType:      "player",
		Message:       "confirm first",
	})
	if err != nil {
		t.Fatalf("create player invitation: %v", err)
	}
	expertInvitation, err := service.CreateInvitation(1, game.ID, InvitationRequest{
		TargetUserID:  3,
		PlayerUserID:  2,
		InviteGroupID: "grouped-confirmation",
		RoleType:      "expert",
		Message:       "review after player",
	})
	if err != nil {
		t.Fatalf("create expert invitation: %v", err)
	}

	if _, _, err := service.RespondInvitation(3, expertInvitation.ID, InvitationRespondRequest{Accept: true}); err != ErrInvitationPlayerPending {
		t.Fatalf("expert response before player = %v, want ErrInvitationPlayerPending", err)
	}

	_, playerApplication, err := service.RespondInvitation(2, playerInvitation.ID, InvitationRespondRequest{Accept: true})
	if err != nil {
		t.Fatalf("player response: %v", err)
	}
	if playerApplication.Status != "pending" || service.IsMember(game.ID, 2) {
		t.Fatalf("player should remain pending and not be added before audit: %+v", playerApplication)
	}

	_, expertApplication, err := service.RespondInvitation(3, expertInvitation.ID, InvitationRespondRequest{Accept: true})
	if err != nil {
		t.Fatalf("expert response after player: %v", err)
	}
	if expertApplication.Status != "pending" || service.IsMember(game.ID, 3) {
		t.Fatalf("expert should remain pending and not be added before audit: %+v", expertApplication)
	}
	if _, err := service.ReviewApplication(1, playerApplication.ID, true); err != nil || !service.IsMember(game.ID, 2) {
		t.Fatalf("player final audit failed: err=%v members=%v", err, service.Members(game.ID))
	}
	if _, err := service.ReviewApplication(1, expertApplication.ID, true); err != nil || !service.IsMember(game.ID, 3) {
		t.Fatalf("expert final audit failed: err=%v members=%v", err, service.Members(game.ID))
	}
}
