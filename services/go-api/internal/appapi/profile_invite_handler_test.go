package appapi

import (
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

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

func TestInviteCodeBindingIsNotCountedAsSuccessfulConversion(t *testing.T) {
	server := newTestAppServer(
		auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()),
		identity.NewService(),
	)

	relations := []invites.Relation{{InviterUserID: 1, InviteeUserID: 2}}
	if got := server.countConvertedInvitees(relations); got != 0 {
		t.Fatalf("invite code binding alone must not count as conversion, got %d", got)
	}
}

func TestInviteRecordsDoNotInventBindingTime(t *testing.T) {
	server := newTestAppServer(
		auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()),
		identity.NewService(),
	)
	relations := []invites.Relation{{InviteCodeID: 9, InviterUserID: 1, InviteeUserID: 2}}
	connections := server.inviteConnections(1, nil, relations)
	if len(connections) != 1 || !connections[0].CreatedAt.IsZero() || !connections[0].UpdatedAt.IsZero() {
		t.Fatalf("legacy invite relation must not receive a fabricated timestamp: %#v", connections)
	}
	records := server.inviteRecords(1, connections)
	if len(records) != 1 || records[0]["statusKey"] != "progress" || records[0]["time"] != "绑定时间待补" {
		t.Fatalf("missing binding time must not be treated as an active or timeout timestamp: %#v", records)
	}
}
