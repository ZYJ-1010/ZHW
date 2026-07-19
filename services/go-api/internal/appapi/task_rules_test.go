package appapi

import (
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/reviews"
)

func TestValidateTaskRulesRequiresSupportedCategoriesAndUniqueCodes(t *testing.T) {
	if err := validateTaskRules([]taskRuleDTO{{Code: "daily", Category: "daily", Title: "每日任务", Required: 1}}); err != nil {
		t.Fatalf("valid task rule rejected: %v", err)
	}
	if err := validateTaskRules([]taskRuleDTO{{Code: "same", Category: "newbie", Title: "A", Required: 1}, {Code: "same", Category: "activity", Title: "B", Required: 1}}); err == nil {
		t.Fatal("duplicate task code should be rejected")
	}
	if err := validateTaskRules([]taskRuleDTO{{Code: "bad", Category: "season", Title: "活动", Required: 1}}); err == nil {
		t.Fatal("unsupported task category should be rejected")
	}
}

func TestAchievementCatalogFiltersByRole(t *testing.T) {
	trace := reviews.Trace{Achievements: []reviews.Achievement{{Code: "expert_only", AchievedAt: time.Now()}}}
	config := growthAchievementConfigDTO{Catalog: []growthAchievementItemDTO{{Code: "expert_only", ID: "expert_only", Title: "行家成就", Roles: []string{"expert"}}}}
	if got := achievementDTOsForRole(trace, config, "player"); len(got) != 1 || got[0].Title != "expert_only" {
		t.Fatalf("unconfigured achievement should remain visible as historical record, got %+v", got)
	}
	if got := achievementDTOsForRole(trace, config, "expert"); len(got) != 1 || got[0].Title != "行家成就" {
		t.Fatalf("expert achievement should use configured role catalog, got %+v", got)
	}
}
