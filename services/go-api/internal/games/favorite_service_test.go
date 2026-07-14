package games

import (
	"context"
	"testing"
	"time"
)

func TestFavoriteGameIsIdempotent(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)

	first, err := service.FavoriteGame(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.FavoriteGame(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}

	if first.GameID != second.GameID || !first.CreatedAt.Equal(second.CreatedAt) {
		t.Fatalf("expected existing favorite on duplicate call, got first=%+v second=%+v", first, second)
	}
	items := service.FavoriteGames(2)
	if len(items) != 1 {
		t.Fatalf("expected 1 favorite, got %d", len(items))
	}
}

func TestUnfavoriteGameIsIdempotent(t *testing.T) {
	service := newVerifiedGameService()
	game := mustCreateRecruitingGame(t, service)

	if _, err := service.FavoriteGame(2, game.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.UnfavoriteGame(2, game.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.UnfavoriteGame(2, game.ID); err != nil {
		t.Fatal(err)
	}
	if items := service.FavoriteGames(2); len(items) != 0 {
		t.Fatalf("expected no favorites, got %+v", items)
	}
}

func TestFavoriteRepositoryIsUsedWhenConfigured(t *testing.T) {
	repo := &fakeFavoriteRepository{items: make(map[int64]map[int64]Favorite)}
	service := NewServiceWithFavoriteRepository(fakeIdentity{verified: true}, repo)
	game := mustCreateRecruitingGame(t, service)

	favorite, err := service.FavoriteGame(2, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !repo.saved || favorite.GameID != game.ID {
		t.Fatalf("expected repository save, saved=%v favorite=%+v", repo.saved, favorite)
	}

	items := service.FavoriteGames(2)
	if !repo.listedByUser || len(items) != 1 || items[0].GameID != game.ID || items[0].Game.ID != game.ID {
		t.Fatalf("expected repository user favorites hydrated with games, listed=%v items=%+v", repo.listedByUser, items)
	}

	all := service.AllFavorites()
	if !repo.listedAll || len(all) != 1 || all[0].GameID != game.ID {
		t.Fatalf("expected repository all favorites, listed=%v items=%+v", repo.listedAll, all)
	}

	if err := service.UnfavoriteGame(2, game.ID); err != nil {
		t.Fatal(err)
	}
	if !repo.deleted {
		t.Fatal("expected repository delete")
	}
	if items := service.FavoriteGames(2); len(items) != 0 {
		t.Fatalf("expected favorite removed through repository, got %+v", items)
	}
}

type fakeFavoriteRepository struct {
	items        map[int64]map[int64]Favorite
	saved        bool
	deleted      bool
	listedByUser bool
	listedAll    bool
}

func (r *fakeFavoriteRepository) SaveFavorite(ctx context.Context, favorite Favorite) (Favorite, error) {
	r.saved = true
	if favorite.CreatedAt.IsZero() {
		favorite.CreatedAt = time.Now()
	}
	if r.items[favorite.UserID] == nil {
		r.items[favorite.UserID] = make(map[int64]Favorite)
	}
	if existing, ok := r.items[favorite.UserID][favorite.GameID]; ok {
		return existing, nil
	}
	r.items[favorite.UserID][favorite.GameID] = favorite
	return favorite, nil
}

func (r *fakeFavoriteRepository) DeleteFavorite(ctx context.Context, userID int64, gameID int64) error {
	r.deleted = true
	if r.items[userID] != nil {
		delete(r.items[userID], gameID)
	}
	return nil
}

func (r *fakeFavoriteRepository) ListFavoritesByUser(ctx context.Context, userID int64) ([]Favorite, error) {
	r.listedByUser = true
	result := make([]Favorite, 0)
	for _, item := range r.items[userID] {
		result = append(result, item)
	}
	sortFavorites(result)
	return result, nil
}

func (r *fakeFavoriteRepository) ListAllFavorites(ctx context.Context) ([]Favorite, error) {
	r.listedAll = true
	result := make([]Favorite, 0)
	for _, userItems := range r.items {
		for _, item := range userItems {
			result = append(result, item)
		}
	}
	sortFavorites(result)
	return result, nil
}
