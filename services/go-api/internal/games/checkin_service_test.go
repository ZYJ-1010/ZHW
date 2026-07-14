package games

import "testing"

func TestCheckinRequiresMemberAndValidType(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGame(t, service)

	if _, err := service.CreateCheckin(99, game.ID, CheckinRequest{CheckinType: "progress"}); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden for non-member, got %v", err)
	}
	if _, err := service.CreateCheckin(2, game.ID, CheckinRequest{}); err != ErrInvalidCheckin {
		t.Fatalf("expected ErrInvalidCheckin, got %v", err)
	}
	if _, err := service.CreateCheckin(2, game.ID, CheckinRequest{CheckinType: "photo", Content: "legacy type"}); err != ErrInvalidCheckin {
		t.Fatalf("expected ErrInvalidCheckin for legacy type, got %v", err)
	}
	checkin, err := service.CreateCheckin(2, game.ID, CheckinRequest{CheckinType: "arrival", Content: "arrived"})
	if err != nil {
		t.Fatal(err)
	}
	if checkin.Status != "valid" {
		t.Fatalf("expected valid checkin, got %+v", checkin)
	}
}

func TestInvalidCheckinIsHiddenFromMemberList(t *testing.T) {
	service := newVerifiedGameService()
	game, _ := mustCreateStartedGame(t, service)
	checkin, err := service.CreateCheckin(2, game.ID, CheckinRequest{CheckinType: "proof", Content: "arrived"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.MarkCheckinInvalid(checkin.ID); err != nil {
		t.Fatal(err)
	}
	items, err := service.Checkins(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected invalid checkin hidden from member list, got %+v", items)
	}
	adminItems, err := service.AdminCheckins(game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(adminItems) != 1 || adminItems[0].Status != "invalid" {
		t.Fatalf("expected admin to see invalid checkin, got %+v", adminItems)
	}
}
