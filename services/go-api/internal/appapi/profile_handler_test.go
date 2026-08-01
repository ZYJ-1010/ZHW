package appapi

import "testing"

func TestNormalizeRoleBenefitConfigSupportsLegacyLeaderField(t *testing.T) {
	config, err := normalizeRoleBenefitConfig(map[string]interface{}{
		"roleComparison": map[string]interface{}{
			"benefits": []interface{}{
				map[string]interface{}{"name": "关系网络", "leader": "✓"},
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeRoleBenefitConfig returned error: %v", err)
	}
	comparison := config["roleComparison"].(map[string]interface{})
	benefits := comparison["benefits"].([]interface{})
	benefit := benefits[0].(map[string]interface{})
	if benefit["guide"] != "✓" {
		t.Fatalf("expected legacy leader value copied to guide, got %#v", benefit)
	}
}

func TestDefaultRoleBenefitConfigUsesGuideField(t *testing.T) {
	comparison := defaultRoleBenefitConfig()["roleComparison"].(map[string]interface{})
	benefits := comparison["benefits"].([]map[string]interface{})
	for _, benefit := range benefits {
		if _, hasLegacy := benefit["leader"]; hasLegacy {
			t.Fatalf("default benefit must not use legacy leader field: %#v", benefit)
		}
		if _, hasGuide := benefit["guide"]; !hasGuide {
			t.Fatalf("default benefit must include guide field: %#v", benefit)
		}
	}
}
