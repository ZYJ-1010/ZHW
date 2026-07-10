package games

import (
	"context"
	"testing"
	"time"
)

func TestServiceConfirmRepositoryPersistsConfirmAndItems(t *testing.T) {
	service := newVerifiedGameService()
	repo := newFakeServiceConfirmRepository()
	service.UseServiceConfirmRepository(repo)
	game, members := mustCreateStartedGame(t, service)

	confirm, items, updatedGame, err := service.ConfirmService(members[0], game.ID, "creator confirmed", 11, 12)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.savedConfirm || !repo.savedItem || confirm.Status != "pending" || updatedGame.Status != "pending_confirm" {
		t.Fatalf("expected first confirm persisted as pending, confirm=%+v items=%+v game=%+v repo=%+v", confirm, items, updatedGame, repo)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after first confirm, got %+v", items)
	}
	if len(items[0].FileIDs) != 2 || items[0].FileIDs[0] != 11 || items[0].FileIDs[1] != 12 {
		t.Fatalf("expected confirm proof file ids, got %+v", items[0])
	}

	for _, userID := range members[1:] {
		confirm, items, updatedGame, err = service.ConfirmService(userID, game.ID, "member confirmed")
		if err != nil {
			t.Fatal(err)
		}
	}
	if confirm.Status != "completed" || updatedGame.Status != "pending_review" || len(items) != len(members) {
		t.Fatalf("expected completed confirm after all members confirmed, confirm=%+v items=%+v game=%+v", confirm, items, updatedGame)
	}

	confirm, items, ok := service.ServiceConfirmForGame(game.ID)
	if !ok || !repo.loadedConfirm || len(items) != len(members) || len(confirm.ConfirmedBy) != len(members) {
		t.Fatalf("expected confirm query from repository, ok=%v confirm=%+v items=%+v repo=%+v", ok, confirm, items, repo)
	}
	if !hasConfirmItemFileIDs(items, members[0], []int64{11, 12}) {
		t.Fatalf("expected persisted confirm proof file ids, got %+v", items)
	}

	_, items, _, err = service.ConfirmService(members[1], game.ID, "duplicate ignored")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != len(members) {
		t.Fatalf("expected duplicate confirm to remain idempotent, got %+v", items)
	}
}

func hasConfirmItemFileIDs(items []ServiceConfirmItem, userID int64, fileIDs []int64) bool {
	for _, item := range items {
		if item.UserID != userID || len(item.FileIDs) != len(fileIDs) {
			continue
		}
		matched := true
		for index := range fileIDs {
			if item.FileIDs[index] != fileIDs[index] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

type fakeServiceConfirmRepository struct {
	nextConfirmID int64
	nextItemID    int64
	confirms      map[int64]ServiceConfirm
	items         map[int64]map[int64]ServiceConfirmItem
	savedConfirm  bool
	savedItem     bool
	loadedConfirm bool
}

func newFakeServiceConfirmRepository() *fakeServiceConfirmRepository {
	return &fakeServiceConfirmRepository{
		nextConfirmID: 1,
		nextItemID:    1,
		confirms:      make(map[int64]ServiceConfirm),
		items:         make(map[int64]map[int64]ServiceConfirmItem),
	}
}

func (r *fakeServiceConfirmRepository) GetConfirm(ctx context.Context, gameID int64) (ServiceConfirm, []ServiceConfirmItem, bool, error) {
	r.loadedConfirm = true
	confirm, ok := r.confirms[gameID]
	if !ok {
		return ServiceConfirm{}, nil, false, nil
	}
	items, err := r.ListConfirmItems(ctx, gameID)
	if err != nil {
		return ServiceConfirm{}, nil, false, err
	}
	confirm.ConfirmedBy = confirmedBy(items)
	return confirm, items, true, nil
}

func (r *fakeServiceConfirmRepository) SaveConfirm(ctx context.Context, confirm ServiceConfirm) (ServiceConfirm, error) {
	r.savedConfirm = true
	if confirm.ID == 0 {
		confirm.ID = r.nextConfirmID
		r.nextConfirmID++
	}
	if confirm.CreatedAt.IsZero() {
		confirm.CreatedAt = time.Now()
	}
	r.confirms[confirm.GameID] = confirm
	return confirm, nil
}

func (r *fakeServiceConfirmRepository) SaveConfirmItem(ctx context.Context, item ServiceConfirmItem) (ServiceConfirmItem, error) {
	r.savedItem = true
	if r.items[item.GameID] == nil {
		r.items[item.GameID] = make(map[int64]ServiceConfirmItem)
	}
	if existing, ok := r.items[item.GameID][item.UserID]; ok {
		return existing, nil
	}
	item.ID = r.nextItemID
	r.nextItemID++
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	r.items[item.GameID][item.UserID] = item
	return item, nil
}

func (r *fakeServiceConfirmRepository) ListConfirmItems(ctx context.Context, gameID int64) ([]ServiceConfirmItem, error) {
	result := make([]ServiceConfirmItem, 0, len(r.items[gameID]))
	for _, item := range r.items[gameID] {
		result = append(result, item)
	}
	return result, nil
}
