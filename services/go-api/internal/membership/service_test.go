package membership

import "testing"

func TestMembershipPlansAndGrant(t *testing.T) {
	service := NewService()
	plans := service.Plans()
	if len(plans) < 2 {
		t.Fatalf("expected default membership plans, got %#v", plans)
	}

	empty := service.My(1)
	if empty.Status != "none" || empty.PlanCode != "none" {
		t.Fatalf("expected empty membership state, got %#v", empty)
	}

	member, err := service.Grant(1, "basic", nil)
	if err != nil {
		t.Fatal(err)
	}
	if member.Status != "active" || member.PlanCode != "basic" || member.PlanName == "" {
		t.Fatalf("expected active basic membership, got %#v", member)
	}

	current := service.My(1)
	if current.Status != "active" || current.PlanCode != "basic" || current.ID == 0 {
		t.Fatalf("expected current membership, got %#v", current)
	}
}

func TestMembershipGrantRejectsInvalidPlan(t *testing.T) {
	service := NewService()
	if _, err := service.Grant(1, "missing", nil); err != ErrPlanNotFound {
		t.Fatalf("expected ErrPlanNotFound, got %v", err)
	}
}
