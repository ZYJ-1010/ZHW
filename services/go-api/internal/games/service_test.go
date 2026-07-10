package games

import "testing"

type fakeIdentity struct {
	verified  bool
	byUserIDs map[int64]bool
}

func (f fakeIdentity) IsVerified(userID int64) bool {
	if f.byUserIDs != nil {
		return f.byUserIDs[userID]
	}
	return f.verified
}

func TestCreateAllowsPlayerWithoutVerifiedIdentity(t *testing.T) {
	service := NewService(fakeIdentity{verified: false})
	game, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8})
	if err != nil {
		t.Fatalf("expected player create without identity, got %v", err)
	}
	if game.ID == 0 || game.CurrentPlayers != 1 {
		t.Fatalf("expected created game, got %+v", game)
	}
}

func TestCreateAllowsOnlyFreeGameFromApp(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	_, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "paid", MinPlayers: 5, MaxPlayers: 8})
	if err != ErrInvalidGameType {
		t.Fatalf("expected ErrInvalidGameType, got %v", err)
	}
}

func TestCreateFromAppIgnoresMainGuideUserID(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{Title: "app free", GameType: "free", MainGuideUserID: 2, MinPlayers: 5, MaxPlayers: 8})
	if err != nil {
		t.Fatalf("app create failed: %v", err)
	}
	if game.MainGuideUserID != 0 {
		t.Fatalf("app create should not set main guide directly, got %+v", game)
	}
}

func TestCreateKeepsTrimmedAddress(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{Title: "address game", GameType: "free", MinPlayers: 5, MaxPlayers: 8, Address: "  北京市朝阳区测试地点  "})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if game.Address != "北京市朝阳区测试地点" {
		t.Fatalf("expected trimmed address, got %q", game.Address)
	}

	longAddress := make([]rune, 256)
	for i := range longAddress {
		longAddress[i] = 'a'
	}
	if _, err := service.Create(2, CreateRequest{Title: "long address", GameType: "free", MinPlayers: 5, MaxPlayers: 8, Address: string(longAddress)}); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput for long address, got %v", err)
	}
}

func TestCreateEnforcesDailyLimit(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	for i := 0; i < 3; i++ {
		if _, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8}); err != nil {
			t.Fatalf("create %d failed: %v", i, err)
		}
	}
	_, err := service.Create(1, CreateRequest{Title: "测试局", GameType: "free", MinPlayers: 5, MaxPlayers: 8})
	if err != ErrDailyLimit {
		t.Fatalf("expected ErrDailyLimit, got %v", err)
	}
}

func TestCreateFromAdminAllowsDocumentedGameTypes(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})

	adminTypes := []string{"free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition"}
	for i, gameType := range adminTypes {
		creatorUserID := int64(11 + i)
		game, err := service.CreateFromAdmin(CreateRequest{Title: "admin " + gameType, CreatorUserID: creatorUserID, GameType: gameType, MinPlayers: 5, MaxPlayers: 8})
		if err != nil {
			t.Fatalf("admin %s create failed: %v", gameType, err)
		}
		if game.GameType != gameType || game.GameSource != "admin" || game.Status != "recruiting" || game.CreatorUserID != creatorUserID {
			t.Fatalf("unexpected %s game: %+v", gameType, game)
		}
	}

	condition, err := service.CreateFromAdmin(CreateRequest{Title: "admin condition guide", CreatorUserID: 31, MainGuideUserID: 41, GameType: "condition", MinPlayers: 5, MaxPlayers: 8})
	if err != nil {
		t.Fatalf("admin condition create failed: %v", err)
	}
	if condition.GameType != "condition" || condition.GameSource != "admin" || condition.Status != "recruiting" || condition.CreatorUserID != 31 || condition.MainGuideUserID != 41 || condition.CurrentPlayers != 2 {
		t.Fatalf("unexpected condition game: %+v", condition)
	}
	if !service.members[condition.ID][41] {
		t.Fatalf("expected admin main guide to be added as a game member")
	}
}

func TestCreateFromAdminRejectsInvalidTypeAndCreator(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})

	if _, err := service.CreateFromAdmin(CreateRequest{Title: "missing creator", GameType: "standard", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput, got %v", err)
	}
	if _, err := service.CreateFromAdmin(CreateRequest{Title: "invalid type", CreatorUserID: 11, GameType: "paid", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameType {
		t.Fatalf("expected ErrInvalidGameType, got %v", err)
	}
	if _, err := service.CreateFromAdmin(CreateRequest{Title: "same guide", CreatorUserID: 11, MainGuideUserID: 11, GameType: "condition", MinPlayers: 5, MaxPlayers: 8}); err != ErrInvalidGameInput {
		t.Fatalf("expected ErrInvalidGameInput for creator as main guide, got %v", err)
	}
}

func TestCreateFromAdminRequiresVerifiedMainGuide(t *testing.T) {
	service := NewService(fakeIdentity{byUserIDs: map[int64]bool{11: true, 22: false}})

	if _, err := service.CreateFromAdmin(CreateRequest{Title: "unverified main guide", CreatorUserID: 11, MainGuideUserID: 22, GameType: "condition", MinPlayers: 5, MaxPlayers: 8}); err != ErrRealnameRequired {
		t.Fatalf("expected ErrRealnameRequired for unverified main guide, got %v", err)
	}
}

func TestCreateUsesConfiguredDailyLimit(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	service.SetDailyCreateLimit(1)

	if _, err := service.Create(1, CreateRequest{Title: "configured limit", GameType: "free", MinPlayers: 5, MaxPlayers: 8}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if _, err := service.Create(1, CreateRequest{Title: "configured limit", GameType: "free", MinPlayers: 5, MaxPlayers: 8}); err != ErrDailyLimit {
		t.Fatalf("expected ErrDailyLimit from configured limit, got %v", err)
	}
}

func TestCreatePersistsCategoryFields(t *testing.T) {
	service := NewService(fakeIdentity{verified: true})
	game, err := service.Create(1, CreateRequest{
		Title:                 "category game",
		GameType:              "free",
		PrimaryCategory:       "task",
		PrimaryCategoryText:   "Task",
		SecondaryCategory:     "project",
		SecondaryCategoryText: "Project",
		Type:                  "free",
		MinPlayers:            5,
		MaxPlayers:            8,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if game.PrimaryCategory != "task" || game.SecondaryCategory != "project" || game.Type != "free" {
		t.Fatalf("expected category fields to persist, got %+v", game)
	}
}
