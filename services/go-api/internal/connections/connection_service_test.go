package connections

import "testing"

func TestUpsertPairCreatesBidirectionalConnectionAndFollowLog(t *testing.T) {
	service := NewService()
	service.UpsertPair(1, 2, "co_game", "game", 10, 2)
	service.UpsertPair(1, 2, "co_game", "game", 10, 3)

	userConnections := service.My(1)
	if len(userConnections) != 1 {
		t.Fatalf("expected one user connection, got %+v", userConnections)
	}
	if userConnections[0].ConnectedUserID != 2 || userConnections[0].StrengthScore != 5 {
		t.Fatalf("expected merged connection strength, got %+v", userConnections[0])
	}
	if len(service.My(2)) != 1 || len(service.All()) != 2 {
		t.Fatalf("expected bidirectional connection, got all=%+v", service.All())
	}

	log, err := service.AddFollowLog(1, userConnections[0].ID, FollowRequest{FollowType: "call", Content: "followed"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if log.OperatorUserID != 1 || log.ConnectionID != userConnections[0].ID {
		t.Fatalf("unexpected follow log: %+v", log)
	}
	if service.My(1)[0].StrengthScore != 6 {
		t.Fatalf("expected follow log to increase strength, got %+v", service.My(1)[0])
	}
}

func TestFollowLogRequiresParticipantOrGuide(t *testing.T) {
	service := NewService()
	service.UpsertPair(1, 2, "co_game", "game", 10, 1)
	connection := service.My(1)[0]

	if _, err := service.AddFollowLog(3, connection.ID, FollowRequest{FollowType: "call", Content: "call"}, false); err != ErrConnectionForbidden {
		t.Fatalf("expected ErrConnectionForbidden, got %v", err)
	}
	if _, err := service.AddFollowLog(3, connection.ID, FollowRequest{FollowType: "call", Content: "call"}, true); err != nil {
		t.Fatalf("expected guide follow log allowed, got %v", err)
	}
}
