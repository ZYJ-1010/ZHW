package points

import (
	"context"
	"errors"
	"testing"
	"time"
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

func TestStrictPointReadsDoNotReturnProcessFallback(t *testing.T) {
	repo := newFakePointRepository()
	repo.getErr = errors.New("points account unavailable")
	repo.logsErr = errors.New("points logs unavailable")
	service := NewServiceWithRepository(repo)
	if account, err := service.SummaryStrict(1); err == nil || account.UserID != 0 {
		t.Fatalf("expected strict account read failure, account=%+v err=%v", account, err)
	}
	if logs, err := service.LogsStrict(1); err == nil || logs != nil {
		t.Fatalf("expected strict logs read failure, logs=%+v err=%v", logs, err)
	}
}

func TestExpireIsIdempotentForSameCutoffDay(t *testing.T) {
	service := NewService()
	if _, _, err := service.Grant(1, 50, "test_reward", 1, "测试奖励"); err != nil {
		t.Fatal(err)
	}
	cutoff := time.Now().Add(24 * time.Hour)
	first, log, err := service.Expire(1, cutoff)
	if err != nil || first.AvailablePoints != 0 || log.ChangeValue != -50 {
		t.Fatalf("expected first expiry deduction, account=%+v log=%+v err=%v", first, log, err)
	}
	second, _, err := service.Expire(1, cutoff)
	if err != nil || second.AvailablePoints != 0 || len(service.Logs(1)) != 2 {
		t.Fatalf("expected repeated expiry not to deduct again, account=%+v logs=%+v err=%v", second, service.Logs(1), err)
	}
}

type fakePointRepository struct {
	nextID   int64
	accounts map[int64]Account
	logs     []Log
	getErr   error
	logsErr  error
}

func newFakePointRepository() *fakePointRepository {
	return &fakePointRepository{nextID: 1, accounts: make(map[int64]Account)}
}

func (r *fakePointRepository) GetAccount(ctx context.Context, userID int64) (Account, bool, error) {
	if r.getErr != nil {
		return Account{}, false, r.getErr
	}
	account, ok := r.accounts[userID]
	return account, ok, nil
}

func (r *fakePointRepository) SaveAccount(ctx context.Context, account Account) (Account, error) {
	r.accounts[account.UserID] = account
	return account, nil
}

func (r *fakePointRepository) ListLogsByUser(ctx context.Context, userID int64) ([]Log, error) {
	if r.logsErr != nil {
		return nil, r.logsErr
	}
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
