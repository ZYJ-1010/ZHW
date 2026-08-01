package appapi

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"zhw-mini/services/go-api/internal/systemconfig"
)

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

func TestNormalizeOperationRulesProvidesNewbieProfileReminderDefaults(t *testing.T) {
	normalized := normalizeOperationRules(operationRulesDTO{})
	if !normalized.NewbieGuide.Enabled {
		t.Fatal("newbie guide must remain enabled in phase one")
	}
	if normalized.NewbieGuide.ProfileReminderLimit != 3 {
		t.Fatalf("expected default reminder limit 3, got %d", normalized.NewbieGuide.ProfileReminderLimit)
	}
	if normalized.NewbieGuide.ProfileReminderIntervalHours != 24 {
		t.Fatalf("expected default reminder interval 24 hours, got %d", normalized.NewbieGuide.ProfileReminderIntervalHours)
	}
}

func TestNormalizeOperationRulesProvidesReauthWindowDefault(t *testing.T) {
	normalized := normalizeOperationRules(operationRulesDTO{})
	if normalized.Login.ReauthAfterDays != 60 {
		t.Fatalf("expected 60-day reauth window, got %d", normalized.Login.ReauthAfterDays)
	}
}

func TestStrictOperationRulesDoNotSilentlyUseDefaultsOnRepositoryFailure(t *testing.T) {
	repoErr := errors.New("operation rules repository unavailable")
	server := &Server{systemConfig: systemconfig.NewServiceWithRepository(failingOperationRulesRepository{err: repoErr})}
	if _, err := server.currentOperationRulesStrict(); !errors.Is(err, repoErr) {
		t.Fatalf("expected strict rules read error, got %v", err)
	}
	if fallback := server.currentOperationRules(); fallback.Game.DailyCreateLimit != defaultOperationRules().Game.DailyCreateLimit {
		t.Fatalf("display compatibility fallback must retain defaults, got %+v", fallback.Game)
	}
}

func TestStrictCreditRestrictionConfigDoesNotSilentlyUseDefaultsOnRepositoryFailure(t *testing.T) {
	repoErr := errors.New("credit restriction repository unavailable")
	server := &Server{systemConfig: systemconfig.NewServiceWithRepository(failingOperationRulesRepository{err: repoErr})}
	if _, err := server.currentCreditRestrictionConfigStrict(); !errors.Is(err, repoErr) {
		t.Fatalf("expected strict credit config read error, got %v", err)
	}
	if fallback := server.currentCreditRestrictionConfig(); fallback.CreateRestrictedBelow != defaultCreditRestrictionConfig().CreateRestrictedBelow {
		t.Fatalf("display compatibility fallback must retain defaults, got %+v", fallback)
	}
}

type failingOperationRulesRepository struct{ err error }

func (r failingOperationRulesRepository) Get(context.Context, string) (json.RawMessage, error) {
	return nil, r.err
}

func (r failingOperationRulesRepository) Set(context.Context, string, json.RawMessage) error {
	return r.err
}
