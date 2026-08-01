package tasks

import (
	"context"
	"errors"
	"sync"
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

func TestStrictProgressReadsReturnRepositoryFailure(t *testing.T) {
	service := NewServiceWithRepository(failingTaskRepository{})
	if _, err := service.CompletedCodesStrict(7); err == nil {
		t.Fatal("expected strict completed-code read to return repository failure")
	}
	if got := service.CompletedCodes(7); len(got) != 0 {
		t.Fatalf("compatibility read must not invent completed tasks: %+v", got)
	}
	if _, err := service.CompletedCodesForDateStrict(7, time.Now()); err == nil {
		t.Fatal("expected strict daily completed-code read to return repository failure")
	}
	if _, err := service.GuideProgressStrict(7); err == nil {
		t.Fatal("expected strict guide-progress read to return repository failure")
	}
}

type failingTaskRepository struct{}

func (failingTaskRepository) UpsertProgress(context.Context, Progress) (Progress, bool, error) {
	return Progress{}, false, errors.New("repository unavailable")
}

func (failingTaskRepository) ListProgress(context.Context, int64) ([]Progress, error) {
	return nil, errors.New("repository unavailable")
}

func (failingTaskRepository) GetGuideProgress(context.Context, int64) (GuideProgress, error) {
	return GuideProgress{}, errors.New("repository unavailable")
}

func (failingTaskRepository) RecordProfileReminder(context.Context, int64, time.Time) (GuideProgress, error) {
	return GuideProgress{}, errors.New("repository unavailable")
}

func TestMarkCompletedOnceHasSingleWinnerUnderConcurrency(t *testing.T) {
	service := NewService()
	const workers = 20
	var wait sync.WaitGroup
	wait.Add(workers)
	created := make(chan bool, workers)
	errors := make(chan error, workers)
	for index := 0; index < workers; index++ {
		go func() {
			defer wait.Done()
			_, first, err := service.MarkCompletedOnce(7, "complete_identity")
			created <- first
			errors <- err
		}()
	}
	wait.Wait()
	close(created)
	close(errors)
	createdCount := 0
	for first := range created {
		if first {
			createdCount++
		}
	}
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if createdCount != 1 {
		t.Fatalf("expected exactly one newly completed result, got %d", createdCount)
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

func TestProfileGuideReminderProgressIsIndependentAndPersistentInService(t *testing.T) {
	service := NewService()
	first, err := service.RecordProfileReminder(7, time.Date(2026, 7, 27, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.RecordProfileReminder(7, time.Date(2026, 7, 28, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if first.ProfileReminderCount != 1 || second.ProfileReminderCount != 2 {
		t.Fatalf("expected sequential reminder counts, first=%+v second=%+v", first, second)
	}
	if got := service.GuideProgress(7); got.ProfileReminderCount != 2 || got.LastProfileReminderAt != second.LastProfileReminderAt {
		t.Fatalf("expected guide progress to be retained, got=%+v", got)
	}
	if service.CompletedCodes(7)["complete_identity"] {
		t.Fatal("displaying a reminder must not complete any task")
	}
}
