package games

import (
	"context"
	"database/sql"
)

type SQLFavoriteRepository struct {
	db *sql.DB
}

func NewSQLFavoriteRepository(db *sql.DB) *SQLFavoriteRepository {
	return &SQLFavoriteRepository{db: db}
}

func (r *SQLFavoriteRepository) SaveFavorite(ctx context.Context, favorite Favorite) (Favorite, error) {
	return scanFavorite(r.db.QueryRowContext(ctx, `
insert into game_favorites (user_id, game_id, created_at)
values ($1,$2,$3)
on conflict (user_id, game_id) do update set
  game_id = game_favorites.game_id
returning user_id, game_id, created_at
`, favorite.UserID, favorite.GameID, favorite.CreatedAt))
}

func (r *SQLFavoriteRepository) DeleteFavorite(ctx context.Context, userID int64, gameID int64) error {
	_, err := r.db.ExecContext(ctx, `delete from game_favorites where user_id = $1 and game_id = $2`, userID, gameID)
	return err
}

func (r *SQLFavoriteRepository) ListFavoritesByUser(ctx context.Context, userID int64) ([]Favorite, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, game_id, created_at
from game_favorites
where user_id = $1
order by created_at desc, game_id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFavorites(rows)
}

func (r *SQLFavoriteRepository) ListAllFavorites(ctx context.Context) ([]Favorite, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, game_id, created_at
from game_favorites
order by created_at desc, game_id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFavorites(rows)
}

func scanFavorites(rows *sql.Rows) ([]Favorite, error) {
	items := make([]Favorite, 0)
	for rows.Next() {
		item, err := scanFavorite(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanFavorite(row interface {
	Scan(dest ...any) error
}) (Favorite, error) {
	var favorite Favorite
	if err := row.Scan(&favorite.UserID, &favorite.GameID, &favorite.CreatedAt); err != nil {
		return Favorite{}, err
	}
	return favorite, nil
}
