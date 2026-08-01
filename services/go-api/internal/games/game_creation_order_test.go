package games

import (
	"context"
	"testing"
)

type atomicGameOrderRepository struct {
	*fakeGameRepository
	called bool
}

func (r *atomicGameOrderRepository) CreateGameWithCreatorAndFreeOrder(ctx context.Context, game Game, creatorRole string) (Game, error) {
	r.called = true
	return r.fakeGameRepository.CreateGameWithCreator(ctx, game, creatorRole)
}

func TestAppGameCreationPrefersAtomicFreeOrderRepository(t *testing.T) {
	repo := &atomicGameOrderRepository{fakeGameRepository: newFakeGameRepository()}
	service := NewServiceWithRepositories(fakeIdentity{verified: true}, repo, nil)
	game, err := service.Create(7, CreateRequest{
		Title: "原子创建免费局", GameType: "free", MinPlayers: 5, MaxPlayers: 8,
		StartAt: "2026-08-01 10:00", EndAt: "2026-08-01 12:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.called {
		t.Fatal("app game creation must use the game/member/free-order transaction when available")
	}
	if game.ID <= 0 || !repo.members[game.ID][7] {
		t.Fatalf("atomic creation must persist the game and creator membership: game=%+v members=%+v", game, repo.members)
	}
}
