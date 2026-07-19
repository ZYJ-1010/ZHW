package appapi

import "testing"

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
