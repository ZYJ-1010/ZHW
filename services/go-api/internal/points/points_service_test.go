package points

import (
	"testing"
	"time"
)

func TestGrantAndDeductCreatePointLogs(t *testing.T) {
	service := NewService()

	account, log, err := service.Grant(1, 100, "review", 10, "review reward")
	if err != nil {
		t.Fatal(err)
	}
	if account.AvailablePoints != 100 || account.TotalEarnedPoints != 100 || log.ChangeValue != 100 {
		t.Fatalf("unexpected grant result account=%+v log=%+v", account, log)
	}

	account, log, err = service.Deduct(1, 30, "redemption", 20, "redeem")
	if err != nil {
		t.Fatal(err)
	}
	if account.AvailablePoints != 70 || log.ChangeValue != -30 || len(service.Logs(1)) != 2 || len(service.AllLogs()) != 2 {
		t.Fatalf("expected deduct log and balance, account=%+v log=%+v logs=%+v", account, log, service.Logs(1))
	}
}

func TestDeductRejectsInsufficientPoints(t *testing.T) {
	service := NewService()

	if _, _, err := service.Deduct(1, 1, "redemption", 1, "redeem"); err != ErrInsufficientPoints {
		t.Fatalf("expected ErrInsufficientPoints, got %v", err)
	}
}

func TestExpireIsIdempotentAndCreatesAuditLog(t *testing.T) {
	service := NewService()
	if _, _, err := service.Grant(1, 40, "review", 1, "old reward"); err != nil {
		t.Fatal(err)
	}
	service.logs[0].CreatedAt = time.Now().Add(-48 * time.Hour)
	account, log, err := service.Expire(1, time.Now().Add(-24*time.Hour))
	if err != nil || account.AvailablePoints != 0 || log.ChangeValue != -40 || log.BizType != "points_expire" {
		t.Fatalf("expected expiry log, account=%+v log=%+v err=%v", account, log, err)
	}
	account, log, err = service.Expire(1, time.Now().Add(-24*time.Hour))
	if err != nil || account.AvailablePoints != 0 || log.ID != 0 {
		t.Fatalf("expected idempotent second expiry, account=%+v log=%+v err=%v", account, log, err)
	}
}

func TestGrantOnceWritesOnlyOneAutomaticReward(t *testing.T) {
	service := NewService()
	firstAccount, firstLog, created, err := service.GrantOnce(7, 15, "completed_game_reward", 99, "完成组局奖励")
	if err != nil || !created || firstAccount.AvailablePoints != 15 || firstLog.ID == 0 {
		t.Fatalf("expected first automatic reward, account=%+v log=%+v created=%v err=%v", firstAccount, firstLog, created, err)
	}
	secondAccount, secondLog, created, err := service.GrantOnce(7, 15, "completed_game_reward", 99, "完成组局奖励")
	if err != nil || created || secondAccount.AvailablePoints != 15 || secondLog.ID != firstLog.ID {
		t.Fatalf("expected idempotent automatic reward, account=%+v log=%+v created=%v err=%v", secondAccount, secondLog, created, err)
	}
}
