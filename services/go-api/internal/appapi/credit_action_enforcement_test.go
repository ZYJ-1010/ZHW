package appapi

import (
	"strings"
	"testing"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestGameDetailRelationReflectsCreditJoinRestriction(t *testing.T) {
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	userID := int64(880001)
	for index := 0; index < 7; index++ {
		server.reviews.DeductCredit(userID, 0, "test_credit_restriction")
	}

	relation := server.buildGameRelation(userID, games.Game{
		ID:             900001,
		Status:         games.StatusRecruiting,
		GameSource:     "app",
		AllowedRoles:   []string{"player"},
		MinPlayers:     5,
		MaxPlayers:     8,
		CurrentPlayers: 1,
	})
	if relation.CanApply {
		t.Fatal("credit-restricted user must not see an enabled apply action")
	}
	if !strings.Contains(relation.ApplyDisabledReason, "信用分") {
		t.Fatalf("expected credit reason in game detail, got %q", relation.ApplyDisabledReason)
	}
}

func TestFrozenCreditBlocksInvitationAction(t *testing.T) {
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	userID := int64(880002)
	for index := 0; index < 9; index++ {
		server.reviews.DeductCredit(userID, 0, "test_credit_restriction")
	}

	if allowed, message := server.canUseCreditAction(userID, "invite_game"); allowed || !strings.Contains(message, "冻结") {
		t.Fatalf("expected frozen account to be blocked from invitations, allowed=%v message=%q", allowed, message)
	}
}
