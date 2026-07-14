package appapi

import (
	"testing"

	"zhw-mini/services/go-api/internal/notifications"
)

func TestGameApprovedNotificationRoutesToGameDetail(t *testing.T) {
	server := &Server{}
	route, routeKey := server.notificationGameGroupRoute(notifications.Notification{
		NotifyType: "game_approved",
		BizType:    "game",
		BizID:      92018,
	})

	if route != "pages/game/detail/index?id=92018" {
		t.Fatalf("unexpected game approved route: %s", route)
	}
	if routeKey != "gameDetail" {
		t.Fatalf("unexpected game approved route key: %s", routeKey)
	}
}
