package tasks

import "testing"

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
