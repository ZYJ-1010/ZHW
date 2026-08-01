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

func TestMyGamesCategoryTabsKeepCreatedManagementSeparate(t *testing.T) {
	tabs := normalizeMyGamesCategoryTabs([]interface{}{
		map[string]interface{}{"key": "joined", "text": "我参与的"},
		map[string]interface{}{"key": "created", "text": "我受邀的"},
	})
	got := map[string]string{}
	for _, tab := range tabs {
		got[tab["key"].(string)] = tab["text"].(string)
	}
	if len(tabs) != 4 || got["joined"] == "" || got["created"] != "我发起/管理的" || got["invited"] == "" || got["favorite"] == "" {
		t.Fatalf("expected independent category tabs, got %+v", tabs)
	}
}

func TestCollaborationProgressForCanceledGameIsTerminal(t *testing.T) {
	progress := collaborationProgress(games.Game{Status: "canceled"})
	if progress["percent"] != 0 || progress["title"] != "组局已取消" {
		t.Fatalf("expected canceled collaboration progress, got %+v", progress)
	}
}
