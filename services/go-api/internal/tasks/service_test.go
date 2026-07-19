package tasks

import (
	"testing"
	"time"
)

func TestCompletedTaskPersistsInService(t *testing.T) {
	service := NewService()
	first, err := service.MarkCompleted(7, "complete_identity")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.MarkCompleted(7, "complete_identity")
	if err != nil {
		t.Fatal(err)
	}
	if first.CompletedAt.IsZero() || second.CompletedAt != first.CompletedAt {
		t.Fatalf("expected idempotent completion, first=%+v second=%+v", first, second)
	}
	if !service.CompletedCodes(7)["complete_identity"] {
		t.Fatal("completed task missing")
	}
}

func TestDailyTaskCompletionIsScopedToDate(t *testing.T) {
	service := NewService()
	today := time.Date(2026, 7, 19, 12, 0, 0, 0, time.Local)
	if _, err := service.MarkCompletedForDate(7, "daily_join_game", today); err != nil {
		t.Fatal(err)
	}
	if !service.CompletedCodesForDate(7, today)["daily_join_game"] {
		t.Fatal("today task should be completed")
	}
	tomorrow := today.Add(24 * time.Hour)
	if service.CompletedCodesForDate(7, tomorrow)["daily_join_game"] {
		t.Fatal("daily task should reset on the next day")
	}
}
