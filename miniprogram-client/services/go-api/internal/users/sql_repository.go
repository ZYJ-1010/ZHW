package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) FindByOpenID(ctx context.Context, openID string) (User, bool, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, `
select u.id, wa.openid, coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), u.realname_status, u.status, u.created_at
from user_wechat_accounts wa
join users u on u.id = wa.user_id
where wa.openid = $1
`, openID))
	if err == sql.ErrNoRows {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	return user, true, nil
}

func (r *SQLRepository) FindByID(ctx context.Context, id int64) (User, bool, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, `
select u.id, coalesce(wa.openid, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), u.realname_status, u.status, u.created_at
from users u
left join user_wechat_accounts wa on wa.user_id = u.id
where u.id = $1
order by wa.id asc
limit 1
`, id))
	if err == sql.ErrNoRows {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	return user, true, nil
}

func (r *SQLRepository) List(ctx context.Context, filter Filter) ([]User, error) {
	filter = normalizeFilter(filter)
	args := make([]any, 0, 3)
	clauses := make([]string, 0, 3)
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("u.status = $%d", len(args)))
	}
	if filter.RealnameStatus != "" {
		args = append(args, filter.RealnameStatus)
		clauses = append(clauses, fmt.Sprintf("u.realname_status = $%d", len(args)))
	}
	if filter.Keyword != "" {
		args = append(args, "%"+filter.Keyword+"%")
		keywordArg := len(args)
		args = append(args, filter.Keyword)
		idArg := len(args)
		clauses = append(clauses, fmt.Sprintf(`(
lower(coalesce(u.nickname, '')) like $%d
or lower(coalesce(wa.openid, '')) like $%d
or cast(u.id as text) = $%d
)`, keywordArg, keywordArg, idArg))
	}

	query := `
select u.id, coalesce(wa.openid, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), u.realname_status, u.status, u.created_at
from users u
left join user_wechat_accounts wa on wa.user_id = u.id
`
	if len(clauses) > 0 {
		query += "where " + strings.Join(clauses, " and ") + "\n"
	}
	query += "order by u.created_at desc, u.id desc"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *SQLRepository) CreateWithOpenID(ctx context.Context, openID string) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRowContext(ctx, `
insert into users (status, realname_status, created_at, updated_at)
values ('active','pending',now(),now())
returning id
`).Scan(&userID)
	if err != nil {
		return User{}, err
	}

	_, err = tx.ExecContext(ctx, `
insert into user_wechat_accounts (user_id, openid, created_at, updated_at)
values ($1,$2,now(),now())
`, userID, openID)
	if err != nil {
		return User{}, err
	}

	_, err = tx.ExecContext(ctx, `
insert into user_profiles (user_id, created_at, updated_at)
values ($1,now(),now())
on conflict (user_id) do nothing
`, userID)
	if err != nil {
		return User{}, err
	}

	user, err := scanUser(tx.QueryRowContext(ctx, `
select u.id, wa.openid, coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), u.realname_status, u.status, u.created_at
from users u
join user_wechat_accounts wa on wa.user_id = u.id
where u.id = $1
`, userID))
	if err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) UpdateProfile(ctx context.Context, userID int64, nickname string, avatarURL string, avatarFileID int64) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
update users
set nickname = $2, avatar_url = $3, avatar_file_id = $4, updated_at = now()
where id = $1
`, userID, nullString(nickname), nullString(avatarURL), nullInt64(avatarFileID))
	if err != nil {
		return User{}, err
	}

	_, err = tx.ExecContext(ctx, `
insert into user_profiles (user_id, created_at, updated_at)
values ($1,now(),now())
on conflict (user_id) do update set updated_at = now()
`, userID)
	if err != nil {
		return User{}, err
	}

	user, err := scanUser(tx.QueryRowContext(ctx, `
select u.id, coalesce(wa.openid, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), u.realname_status, u.status, u.created_at
from users u
left join user_wechat_accounts wa on wa.user_id = u.id
where u.id = $1
order by wa.id asc
limit 1
`, userID))
	if err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return user, nil
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.OpenID, &user.Nickname, &user.AvatarURL, &user.AvatarFileID, &user.RealnameStatus, &user.Status, &user.CreatedAt)
	return user, err
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}
