package appapi

import (
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

func TestPublicGamesFiltersExpiredAndFullGames(t *testing.T) {
	now := time.Now()
	items := []games.Game{
		{ID: 1, Status: "recruiting", CurrentPlayers: 2, MaxPlayers: 5, SignupStartAt: now.Add(-time.Hour).Format(time.RFC3339), SignupEndAt: now.Add(time.Hour).Format(time.RFC3339)},
		{ID: 2, Status: "recruiting", CurrentPlayers: 5, MaxPlayers: 5, SignupStartAt: now.Add(-time.Hour).Format(time.RFC3339), SignupEndAt: now.Add(time.Hour).Format(time.RFC3339)},
		{ID: 3, Status: "recruiting", CurrentPlayers: 2, MaxPlayers: 5, SignupStartAt: now.Add(-2 * time.Hour).Format(time.RFC3339), SignupEndAt: now.Add(-time.Hour).Format(time.RFC3339)},
		{ID: 4, Status: "completed", CurrentPlayers: 2, MaxPlayers: 5},
	}
	visible := publicGames(items)
	if len(visible) != 1 || visible[0].ID != 1 {
		t.Fatalf("expected only joinable game to remain, got %+v", visible)
	}
}

func TestHomeGameCategoryTextUsesPrimaryCategory(t *testing.T) {
	game := games.Game{GameType: "free", PrimaryCategory: "growth"}
	if got := homeGameCategoryText(game); got != "成长局" {
		t.Fatalf("expected primary category text, got %q", got)
	}
}
