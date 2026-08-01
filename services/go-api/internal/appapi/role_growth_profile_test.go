package appapi

import (
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/games"
)

func TestRoleMetricWindowsExcludeOlderGames(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	starts := roleMetricWindowStarts([]roleMetricRuleDTO{{MetricCode: "completion_rate", WindowDays: 90}}, now)
	start := starts["completion_rate"]
	if start.Format("2006-01-02") != "2026-05-03" {
		t.Fatalf("unexpected 90-day window start: %s", start)
	}
	if gameWithinGrowthWindow(games.Game{CreatedAt: start.Add(-time.Second)}, start) {
		t.Fatal("game before the configured window must not contribute to the metric")
	}
	if !gameWithinGrowthWindow(games.Game{CreatedAt: start}, start) {
		t.Fatal("game at the configured window boundary must contribute to the metric")
	}
}
