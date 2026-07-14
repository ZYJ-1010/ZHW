package appapi

import (
	"testing"

	"zhw-mini/services/go-api/internal/games"
)

func TestInvitationGameScheduleSupportsMiniProgramFormat(t *testing.T) {
	game := games.Game{StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:30"}
	timeText, durationText := invitationGameSchedule(game)
	if timeText != "2026-07-12 14:00 - 16:30" {
		t.Fatalf("unexpected time text: %q", timeText)
	}
	if durationText != "2小时30分钟" {
		t.Fatalf("unexpected duration text: %q", durationText)
	}
}

func TestInviteSourceGameConfigSeparatesScheduleAndDuration(t *testing.T) {
	game := games.Game{ID: 1, Title: "测试局", StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"}
	config := inviteSourceGameConfig(game)
	source, ok := config["sourceGame"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing sourceGame: %+v", config)
	}
	if source["expectedTime"] != "2026-07-12 14:00 - 16:00" || source["serviceDuration"] != "2小时" {
		t.Fatalf("unexpected schedule fields: %+v", source)
	}
	if source["startAt"] != game.StartAt || source["endAt"] != game.EndAt {
		t.Fatalf("missing raw schedule fields: %+v", source)
	}
}
