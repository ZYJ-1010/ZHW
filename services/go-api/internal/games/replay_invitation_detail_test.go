package games

import "testing"

func TestCreateReplayGamePersistsInvitationDetail(t *testing.T) {
	service := newVerifiedGameService()
	source := mustCreateRecruitingGame(t, service)

	result, err := service.CreateReplayGame(1, source.ID, ReplayGameRequest{
		PlayerUserIDs:    []int64{2},
		ExpertUserIDs:    []int64{3},
		Message:          "请确认本次引荐",
		InviteGroupID:    "replay-group-1",
		ServiceType:      "产品咨询",
		ServiceDuration:  "2小时",
		DemandDetail:     "梳理产品架构",
		BudgetAmountCent: 80000,
		ExpectedTime:     "本周内",
		Overrides: ReplayGameOverrides{
			Title: "产品咨询续局",
		},
	})
	if err != nil {
		t.Fatalf("create replay game failed: %v", err)
	}
	if result.PrimaryInvitationID <= 0 || len(result.Invitations) != 2 {
		t.Fatalf("unexpected replay result: %+v", result)
	}
	if result.Game.StartAt != source.StartAt || result.Game.EndAt != source.EndAt {
		t.Fatalf("replay game should inherit source schedule: source=%+v replay=%+v", source, result.Game)
	}

	for _, invitation := range result.Invitations {
		if invitation.InviteGroupID != "replay-group-1" || invitation.ServiceType != "产品咨询" || invitation.ServiceDuration != "2小时" || invitation.DemandDetail != "梳理产品架构" || invitation.BudgetAmountCent != 80000 || invitation.ExpectedTime != "本周内" {
			t.Fatalf("invitation detail not preserved: %+v", invitation)
		}
		if invitation.Role == "expert" && invitation.ExpertUserID != invitation.TargetUserID {
			t.Fatalf("expert relation missing: %+v", invitation)
		}
		if invitation.Role == "player" && (invitation.PlayerUserID != invitation.TargetUserID || invitation.ExpertUserID != 3) {
			t.Fatalf("player relation missing: %+v", invitation)
		}
	}
}
