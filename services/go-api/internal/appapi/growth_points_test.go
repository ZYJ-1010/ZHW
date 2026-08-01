package appapi

import "testing"

func TestTaskRewardBizIDSeparatesDailyOccurrences(t *testing.T) {
	first := taskRewardBizID("daily_join_game@2026-07-30")
	second := taskRewardBizID("daily_join_game@2026-07-31")
	if first == second {
		t.Fatal("daily task rewards from different dates must not share an idempotency key")
	}
	if first != taskRewardBizID("daily_join_game@2026-07-30") {
		t.Fatal("the same daily task occurrence must keep a stable idempotency key")
	}
}
