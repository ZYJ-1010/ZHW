package appapi

import (
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

func TestPublicGamesKeepsSameDayFullOrAutoStartedCards(t *testing.T) {
	now := time.Now()
	items := publicGames([]games.Game{
		{ID: 1, Status: "recruiting"},
		{ID: 2, Status: "full", CreatedAt: now},
		{ID: 3, Status: "in_progress", CurrentPlayers: 5, MaxPlayers: 5, StartedAt: now.Format(time.RFC3339)},
		{ID: 4, Status: "pending_confirm"},
		{ID: 5, Status: "completed"},
	})
	if len(items) != 3 || items[0].ID != 1 || items[1].ID != 2 || items[2].ID != 3 {
		t.Fatalf("public games = %+v, want recruiting and same-day full cards", items)
	}
}

func TestGameDetailShowsEndedWhileDeliveryConfirmationIsPending(t *testing.T) {
	if got := gameDetailStatusText("pending_confirm"); got != "已结束" {
		t.Fatalf("pending_confirm detail status = %q, want 已结束", got)
	}
	if got := gameDetailStatusText("in_progress"); got != "进行中" {
		t.Fatalf("in_progress detail status = %q, want 进行中", got)
	}
}

func TestGameStatusTextCoversTerminalAndExceptionalStates(t *testing.T) {
	tests := map[string]string{
		"rejected": "\u5ba1\u6838\u672a\u901a\u8fc7",
		"canceled": "\u5df2\u53d6\u6d88",
		"disputed": "\u4e89\u8bae\u4e2d",
		"settling": "\u7ed3\u7b97\u4e2d",
		"closed":   "\u5df2\u5173\u95ed",
		"unknown":  "\u72b6\u6001\u5904\u7406\u4e2d",
	}
	for status, expected := range tests {
		if got := homeGameStatusText(status); got != expected {
			t.Fatalf("status %q text = %q, want %q", status, got, expected)
		}
	}
}

func TestPendingConfirmationIsCompletedInServiceLists(t *testing.T) {
	statusType, statusText := serviceOrderStatus("pending_confirm")
	if statusType != "complete" || statusText != "已完成" {
		t.Fatalf("pending_confirm service status = %q/%q, want complete/已完成", statusType, statusText)
	}
}

func TestManagedGameViewerRolesIncludeExpertAndGuide(t *testing.T) {
	for _, role := range []string{"guide", "main_guide"} {
		if !isManagedGameRole(role) {
			t.Fatalf("managed game role %q must be included", role)
		}
	}
	if isManagedGameRole("member") {
		t.Fatal("member must not be treated as a managed game role")
	}
}

func TestReadableIMMessageSummaryHidesStructuredPayload(t *testing.T) {
	raw := `{"actionRoute":"pages/game/delivery/index","memberDesc":"请行家先确认完成，随后由玩家确认","title":"本局进入完成确认"}`
	if got := readableIMMessageSummary(raw); got != "请行家先确认完成，随后由玩家确认" {
		t.Fatalf("summary = %q", got)
	}
}
