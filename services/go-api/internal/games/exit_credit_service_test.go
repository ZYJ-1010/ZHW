package games

import "testing"

func TestExitAfterAdmissionRequiresCreditDeduction(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)
	mustApproveMembers(t, service, game.ID, 2)

	result, err := service.Exit(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Reason != "quit_after_admitted" {
		t.Fatalf("expected admitted exit reason, got %+v", result)
	}
	if !result.CreditDeduct || !result.CreditDeducted {
		t.Fatalf("expected admitted exit to require a credit deduction, got %+v", result)
	}
}
