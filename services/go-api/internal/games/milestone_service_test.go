package games

import "testing"

func TestMilestoneRequiresCreatorAndValidStatus(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)

	if _, err := service.CreateMilestone(2, game.ID, MilestoneRequest{Title: "阶段一"}); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden for non-creator, got %v", err)
	}
	milestone, err := service.CreateMilestone(1, game.ID, MilestoneRequest{Title: "阶段一"})
	if err != nil {
		t.Fatal(err)
	}
	if milestone.Status != "pending" {
		t.Fatalf("expected pending milestone, got %+v", milestone)
	}
	if _, err := service.UpdateMilestone(1, game.ID, milestone.ID, MilestoneRequest{Status: "done"}); err != ErrInvalidMilestone {
		t.Fatalf("expected ErrInvalidMilestone, got %v", err)
	}
	inProgress, err := service.UpdateMilestone(1, game.ID, milestone.ID, MilestoneRequest{Status: "in_progress"})
	if err != nil {
		t.Fatal(err)
	}
	if inProgress.Status != "in_progress" {
		t.Fatalf("expected in_progress milestone, got %+v", inProgress)
	}
	updated, err := service.UpdateMilestone(1, game.ID, milestone.ID, MilestoneRequest{Status: "completed"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "completed" {
		t.Fatalf("expected completed milestone, got %+v", updated)
	}
}

func TestMilestonesRequireMembership(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGame(t, service)
	if _, err := service.CreateMilestone(1, game.ID, MilestoneRequest{Title: "阶段一"}); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Milestones(99, game.ID); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden for non-member, got %v", err)
	}
	items, err := service.Milestones(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 milestone, got %d", len(items))
	}
}
