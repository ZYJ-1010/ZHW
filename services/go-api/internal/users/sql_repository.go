package users

import (
	"context"
	"database/sql"
	"errors"
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
select u.id, wa.openid, coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
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

func (r *SQLRepository) FindByPhoneHash(ctx context.Context, phoneHash string) (User, bool, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, `
select u.id, coalesce(wa.openid, ''), coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
from users u
left join user_wechat_accounts wa on wa.user_id = u.id
where u.mobile_hash = $1
order by wa.id asc
limit 1
`, phoneHash))
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
select u.id, coalesce(wa.openid, ''), coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
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
or lower(coalesce(u.mobile_masked, '')) like $%d
or cast(u.id as text) = $%d
)`, keywordArg, keywordArg, keywordArg, idArg))
	}

	query := `
select u.id, coalesce(wa.openid, ''), coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
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
select u.id, wa.openid, coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
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

func (r *SQLRepository) CreateWithPhone(ctx context.Context, phoneHash string, phoneMasked string) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRowContext(ctx, `
insert into users (status, realname_status, mobile_hash, mobile_masked, created_at, updated_at)
values ('active','pending',$1,$2,now(),now())
returning id
`, phoneHash, phoneMasked).Scan(&userID)
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
select u.id, '', coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
from users u
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

func (r *SQLRepository) UpdatePhoneAuth(ctx context.Context, userID int64, phoneHash string, phoneMasked string) (User, error) {
	_, err := r.db.ExecContext(ctx, `
update users
set mobile_hash = $2, mobile_masked = $3, updated_at = now()
where id = $1
`, userID, phoneHash, phoneMasked)
	if err != nil {
		return User{}, err
	}
	user, ok, err := r.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrInvalidProfile
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
select u.id, coalesce(wa.openid, ''), coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
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

func (r *SQLRepository) UpdateRealnameStatus(ctx context.Context, userID int64, status string) (User, error) {
	_, err := r.db.ExecContext(ctx, `
update users
set realname_status = $2, updated_at = now()
where id = $1
`, userID, strings.TrimSpace(status))
	if err != nil {
		return User{}, err
	}
	user, ok, err := r.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrInvalidProfile
	}
	return user, nil
}

func (r *SQLRepository) PasswordHash(ctx context.Context, userID int64) (string, bool, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `select password_hash from user_login_passwords where user_id = $1`, userID).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return hash, hash != "", nil
}

func (r *SQLRepository) UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
insert into user_login_passwords (user_id, password_hash, updated_at)
values ($1,$2,now())
on conflict (user_id) do update set password_hash = excluded.password_hash, updated_at = now()
`, userID, passwordHash)
	return err
}

func (r *SQLRepository) BindWechat(ctx context.Context, userID int64, openID string) (User, error) {
	var existingOpenID string
	err := r.db.QueryRowContext(ctx, `
select openid
from user_wechat_accounts
where user_id = $1
order by id asc
limit 1
`, userID).Scan(&existingOpenID)
	if err != nil && err != sql.ErrNoRows {
		return User{}, err
	}
	if err == nil && existingOpenID != openID {
		return User{}, ErrWechatConflict
	}
	result, err := r.db.ExecContext(ctx, `
insert into user_wechat_accounts (user_id, openid, created_at, updated_at)
values ($1,$2,now(),now())
on conflict (openid) do nothing
`, userID, openID)
	if err != nil {
		return User{}, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return User{}, err
	} else if affected == 0 {
		var ownerUserID int64
		if err := r.db.QueryRowContext(ctx, `select user_id from user_wechat_accounts where openid = $1`, openID).Scan(&ownerUserID); err != nil {
			return User{}, err
		}
		if ownerUserID != userID {
			return User{}, ErrWechatAlreadyUsed
		}
	}
	user, ok, err := r.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrInvalidProfile
	}
	return user, nil
}

func (r *SQLRepository) DeactivateAndClearLoginBindings(ctx context.Context, userID int64) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
update users
set status = 'deleted',
    nickname = null,
    avatar_url = null,
    avatar_file_id = null,
    mobile_encrypted = null,
    mobile_hash = null,
    mobile_masked = null,
    updated_at = now()
where id = $1
`, userID)
	if err != nil {
		return User{}, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return User{}, err
	} else if affected == 0 {
		return User{}, ErrInvalidProfile
	}

	_, err = tx.ExecContext(ctx, `
delete from user_wechat_accounts
where user_id = $1
`, userID)
	if err != nil {
		return User{}, err
	}
	_, err = tx.ExecContext(ctx, `
update user_profiles
set gender = null,
    bio = null,
    interest_tags = '[]'::jsonb,
    city_code = null,
    city_name = null,
    updated_at = now()
where user_id = $1
`, userID)
	if err != nil {
		return User{}, err
	}

	user, err := scanUser(tx.QueryRowContext(ctx, `
select u.id, '', coalesce(u.mobile_masked, ''), coalesce(u.nickname, ''), coalesce(u.avatar_url, ''), coalesce(u.avatar_file_id, 0), coalesce(u.account_type, 'player'), u.realname_status, u.status, u.created_at
from users u
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

func scanUser(row interface {
	Scan(dest ...any) error
}) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.OpenID, &user.PhoneMasked, &user.Nickname, &user.AvatarURL, &user.AvatarFileID, &user.AccountType, &user.RealnameStatus, &user.Status, &user.CreatedAt)
	return user, err
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}
