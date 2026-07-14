package appapi

import (
	"testing"

	"zhw-mini/services/go-api/internal/games"
)

func TestPublicGamesOnlyReturnsRecruitingGames(t *testing.T) {
	items := publicGames([]games.Game{
		{ID: 1, Status: "recruiting"},
		{ID: 2, Status: "full"},
		{ID: 3, Status: "in_progress"},
		{ID: 4, Status: "pending_confirm"},
		{ID: 5, Status: "completed"},
	})
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("public games = %+v, want recruiting game only", items)
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
