package appapi

import (
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/revenue"
	"zhw-mini/services/go-api/internal/users"
)

func TestInviteIncomeTrendSeriesDoesNotGenerateSyntheticValues(t *testing.T) {
	if got := inviteIncomeTrendSeries(nil); len(got) != 0 {
		t.Fatalf("expected empty trend without real logs, got %#v", got)
	}
	created := time.Date(2026, 7, 19, 12, 0, 0, 0, time.Local)
	got := inviteIncomeTrendSeries([]revenue.IncomeLog{
		{CreatedAt: created, AmountCent: 1200},
		{CreatedAt: created.Add(24 * time.Hour), AmountCent: 300},
	})
	if len(got) != 1 || got[0]["amount"] != int64(1500) {
		t.Fatalf("expected one real monthly aggregate, got %#v", got)
	}
}

func TestIncomeDisplayTextHonorsPhaseOneRevenueSwitch(t *testing.T) {
	if got := incomeDisplayText(false, 1600); got != "一期未启用" {
		t.Fatalf("expected disabled revenue label, got %q", got)
	}
	if got := incomeDisplayText(true, 1600); got != "¥16.00" {
		t.Fatalf("expected formatted enabled revenue, got %q", got)
	}
}

func TestInviteNetworkNodesUseRealMemberLabels(t *testing.T) {
	server := newTestAppServer(
		auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()),
		identity.NewService(),
	)
	items := []connections.Connection{{ConnectedUserID: 42}}
	nodes := inviteNetworkNodes(server, items)
	if len(nodes) != 1 || nodes[0] == "1" {
		t.Fatalf("expected real member label instead of numeric placeholder, got %#v", nodes)
	}
}
