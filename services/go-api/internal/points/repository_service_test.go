package points

import (
	"context"
	"testing"
)

func TestRepositoryBackedGrantAndDeduct(t *testing.T) {
	repo := newFakePointRepository()
	service := NewServiceWithRepository(repo)

	account, log, err := service.Grant(1, 50, "review_reward", 7, "reward")
	if err != nil {
		t.Fatal(err)
	}
	if account.AvailablePoints != 50 || account.TotalEarnedPoints != 50 || log.BeforePoints != 0 || log.AfterPoints != 50 {
		t.Fatalf("unexpected grant result account=%+v log=%+v", account, log)
	}

	account, log, err = service.Deduct(1, 20, "redemption_order", 8, "redeem")
	if err != nil {
		t.Fatal(err)
	}
	if account.AvailablePoints != 30 || log.ChangeValue != -20 || log.BizType != "redemption_order" {
		t.Fatalf("unexpected deduct result account=%+v log=%+v", account, log)
	}
	if len(service.Logs(1)) != 2 || len(service.AllLogs()) != 2 {
		t.Fatalf("expected repository logs, got user=%+v all=%+v", service.Logs(1), service.AllLogs())
	}
}

type fakePointRepository struct {
	nextID   int64
	accounts map[int64]Account
	logs     []Log
}

func newFakePointRepository() *fakePointRepository {
	return &fakePointRepository{nextID: 1, accounts: make(map[int64]Account)}
}

func (r *fakePointRepository) GetAccount(ctx context.Context, userID int64) (Account, bool, error) {
	account, ok := r.accounts[userID]
	return account, ok, nil
}

func (r *fakePointRepository) SaveAccount(ctx context.Context, account Account) (Account, error) {
	r.accounts[account.UserID] = account
	return account, nil
}

func (r *fakePointRepository) ListLogsByUser(ctx context.Context, userID int64) ([]Log, error) {
	result := make([]Log, 0)
	for _, item := range r.logs {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakePointRepository) ListLogs(ctx context.Context) ([]Log, error) {
	return append([]Log(nil), r.logs...), nil
}

func (r *fakePointRepository) AddLog(ctx context.Context, log Log) (Log, error) {
	log.ID = r.nextID
	r.nextID++
	r.logs = append(r.logs, log)
	return log, nil
}
