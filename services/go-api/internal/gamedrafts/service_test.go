package gamedrafts

import (
	"encoding/json"
	"testing"
)

func TestDraftsAreScopedToCreatorAndPersistInService(t *testing.T) {
	service := NewService()
	payload := json.RawMessage(`{"form":{"theme":"测试局"}}`)

	created, err := service.Save(101, 0, "测试局", payload)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if created.ID <= 0 || created.CreatorUserID != 101 {
		t.Fatalf("unexpected created draft: %+v", created)
	}

	updated, err := service.Save(101, created.ID, "更新后的局", json.RawMessage(`{"form":{"theme":"更新后的局"}}`))
	if err != nil {
		t.Fatalf("update draft: %v", err)
	}
	if updated.Title != "更新后的局" {
		t.Fatalf("expected updated title, got %+v", updated)
	}

	if _, err := service.Get(202, created.ID); err != ErrForbidden {
		t.Fatalf("expected cross-user access forbidden, got %v", err)
	}
	items, err := service.List(101)
	if err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("expected one creator draft, items=%+v err=%v", items, err)
	}
	if err := service.Delete(101, created.ID); err != nil {
		t.Fatalf("delete draft: %v", err)
	}
	if _, err := service.Get(101, created.ID); err != ErrNotFound {
		t.Fatalf("expected deleted draft missing, got %v", err)
	}
}
