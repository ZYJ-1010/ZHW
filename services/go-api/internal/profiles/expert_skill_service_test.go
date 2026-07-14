package profiles

import "testing"

func TestExpertSkillRequiresGrantedExpertRole(t *testing.T) {
	service := NewService()

	if _, err := service.UpdateExpertSkill(1, ExpertSkillRequest{SkillTree: []string{"boardgame"}}); err != ErrExpertForbidden {
		t.Fatalf("expected ErrExpertForbidden, got %v", err)
	}

	service.GrantRole(1, "expert")
	profile, err := service.UpdateExpertSkill(1, ExpertSkillRequest{
		SkillTree:   []string{"boardgame"},
		ServiceTags: []string{"host"},
		CaseFileIDs: []int64{11},
	})
	if err != nil {
		t.Fatal(err)
	}
	if profile.UserID != 1 || profile.Completeness != 100 || len(service.AllExpertSkills()) != 1 {
		t.Fatalf("expected complete expert profile, got %+v", profile)
	}
}

func TestGuideResourceRequiresGrantedGuideRole(t *testing.T) {
	service := NewService()

	if _, err := service.UpdateGuideResource(2, GuideResourceRequest{ResourceTags: []string{"venue"}}); err != ErrGuideForbidden {
		t.Fatalf("expected ErrGuideForbidden, got %v", err)
	}

	service.GrantRole(2, "guide")
	profile, err := service.UpdateGuideResource(2, GuideResourceRequest{
		ResourceTags:    []string{"venue"},
		IndustryTags:    []string{"entertainment"},
		CityCodes:       []string{"110100"},
		ConnectionScale: "100-500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if profile.UserID != 2 || profile.Completeness != 100 || len(service.AllGuideResources()) != 1 {
		t.Fatalf("expected complete guide profile, got %+v", profile)
	}
}
