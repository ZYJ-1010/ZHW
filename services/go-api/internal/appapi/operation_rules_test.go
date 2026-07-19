package appapi

import "testing"

func TestNormalizeOperationRulesKeepsPhaseOneCapabilitiesDisabled(t *testing.T) {
	rules := defaultOperationRules()
	rules.Roles.MembershipRequired = true
	rules.Invite.PlayerEnabled = true
	rules.Revenue.Enabled = true

	normalized := normalizeOperationRules(rules)
	if normalized.Roles.MembershipRequired {
		t.Fatal("phase one must keep membership requirement disabled")
	}
	if normalized.Invite.PlayerEnabled {
		t.Fatal("players must not gain invitation capability through configuration")
	}
	if normalized.Revenue.Enabled {
		t.Fatal("phase one must not enable real revenue through configuration")
	}
}
