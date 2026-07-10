package points

import "testing"

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
