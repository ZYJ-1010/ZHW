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

func TestCurrentGameInvitePermissionMatchesCreateRequirements(t *testing.T) {
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	game := games.Game{ID: 1, CreatorUserID: 101, Status: games.StatusRecruiting}

	if allowed, reason := server.currentGameInvitePermission(101, game); allowed || !strings.Contains(reason, "行家或领路人") {
		t.Fatalf("expected player-only creator to be rejected consistently, allowed=%v reason=%q", allowed, reason)
	}

	server.profiles.GrantRole(101, "guide")
	if allowed, reason := server.currentGameInvitePermission(101, game); !allowed || reason != "" {
		t.Fatalf("expected active guide creator to be allowed, allowed=%v reason=%q", allowed, reason)
	}

	server.profiles.GrantRole(202, "guide")
	if allowed, reason := server.currentGameInvitePermission(202, game); allowed || !strings.Contains(reason, "局创建者或主领路人") {
		t.Fatalf("expected unrelated guide to be rejected, allowed=%v reason=%q", allowed, reason)
	}

	game.Status = games.StatusInProgress
	if allowed, reason := server.currentGameInvitePermission(101, game); allowed || !strings.Contains(reason, "不在招募中") {
		t.Fatalf("expected non-recruiting game to be rejected, allowed=%v reason=%q", allowed, reason)
	}

	game.Status = games.StatusRecruiting
	for index := 0; index < 9; index++ {
		server.reviews.DeductCredit(101, 0, "test_invite_freeze")
	}
	if allowed, reason := server.currentGameInvitePermission(101, game); allowed || !strings.Contains(reason, "冻结") {
		t.Fatalf("expected frozen guide creator to be rejected, allowed=%v reason=%q", allowed, reason)
	}
}

func TestGameRelationExposesCreditApplicationRestriction(t *testing.T) {
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	game, err := server.games.Create(301, games.CreateRequest{
		Title:      "信用资格前置测试局",
		GameType:   "free",
		MinPlayers: 5,
		MaxPlayers: 8,
		StartAt:    "2026-08-12 14:00",
		EndAt:      "2026-08-12 16:00",
	})
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	game, err = server.games.ApproveGame(game.ID)
	if err != nil {
		t.Fatalf("approve game: %v", err)
	}

	const applicantID int64 = 302
	for index := 0; index < 7; index++ {
		server.reviews.DeductCredit(applicantID, 0, "test_join_restriction")
	}
	relation := server.buildGameRelation(applicantID, game)
	if relation.CanApply {
		t.Fatal("expected credit-restricted user not to be eligible for application")
	}
	if !strings.Contains(relation.ApplyDisabledReason, "报名阈值") {
		t.Fatalf("expected relation to expose credit restriction, got %q", relation.ApplyDisabledReason)
	}
}
