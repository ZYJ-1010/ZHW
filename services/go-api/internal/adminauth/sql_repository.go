package adminauth

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) FindAccountByUsername(ctx context.Context, username string) (StoredAdminAccount, bool, error) {
	account, err := scanStoredAdminAccount(r.db.QueryRowContext(ctx, adminAccountSelect()+` where u.username = $1 group by u.id`, username))
	if errors.Is(err, sql.ErrNoRows) {
		return StoredAdminAccount{}, false, nil
	}
	return account, err == nil, err
}

func (r *SQLRepository) FindAccountByID(ctx context.Context, id int64) (StoredAdminAccount, bool, error) {
	account, err := scanStoredAdminAccount(r.db.QueryRowContext(ctx, adminAccountSelect()+` where u.id = $1 group by u.id`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return StoredAdminAccount{}, false, nil
	}
	return account, err == nil, err
}

func (r *SQLRepository) ListAccounts(ctx context.Context) ([]StoredAdminAccount, error) {
	rows, err := r.db.QueryContext(ctx, adminAccountSelect()+` group by u.id order by u.id asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]StoredAdminAccount, 0)
	for rows.Next() {
		item, err := scanStoredAdminAccount(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) CreateAccount(ctx context.Context, account StoredAdminAccount) (StoredAdminAccount, error) {
	if _, exists, err := r.FindAccountByUsername(ctx, account.User.Username); err != nil {
		return StoredAdminAccount{}, err
	} else if exists {
		return StoredAdminAccount{}, ErrAdminExists
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return StoredAdminAccount{}, err
	}
	defer tx.Rollback()
	var id int64
	if err := tx.QueryRowContext(ctx, `
insert into admin_users (username, password_hash, status, created_at, updated_at)
values ($1,$2,$3,now(),now())
returning id
`, account.User.Username, account.PasswordHash, account.User.Status).Scan(&id); err != nil {
		return StoredAdminAccount{}, err
	}
	if err := replaceAccountRoles(ctx, tx, id, account.User.Roles); err != nil {
		return StoredAdminAccount{}, err
	}
	if err := tx.Commit(); err != nil {
		return StoredAdminAccount{}, err
	}
	saved, ok, err := r.FindAccountByID(ctx, id)
	if err != nil {
		return StoredAdminAccount{}, err
	}
	if !ok {
		return StoredAdminAccount{}, ErrAdminNotFound
	}
	return saved, nil
}

func (r *SQLRepository) UpdateAccount(ctx context.Context, account StoredAdminAccount) (StoredAdminAccount, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return StoredAdminAccount{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
update admin_users
set status = $2, updated_at = now()
where id = $1
`, account.User.ID, account.User.Status)
	if err != nil {
		return StoredAdminAccount{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return StoredAdminAccount{}, err
	}
	if affected == 0 {
		return StoredAdminAccount{}, ErrAdminNotFound
	}
	if err := replaceAccountRoles(ctx, tx, account.User.ID, account.User.Roles); err != nil {
		return StoredAdminAccount{}, err
	}
	if err := tx.Commit(); err != nil {
		return StoredAdminAccount{}, err
	}
	saved, ok, err := r.FindAccountByID(ctx, account.User.ID)
	if err != nil {
		return StoredAdminAccount{}, err
	}
	if !ok {
		return StoredAdminAccount{}, ErrAdminNotFound
	}
	return saved, nil
}

func (r *SQLRepository) RoleSummaries(ctx context.Context) ([]RoleSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
select r.role_code, r.role_name,
       count(distinct ur.admin_user_id) as admin_count,
       coalesce(string_agg(distinct p.permission_code, ','), '') as permissions
from admin_roles r
left join admin_user_roles ur on ur.role_id = r.id
left join admin_role_permissions rp on rp.role_id = r.id
left join admin_permissions p on p.id = rp.permission_id
group by r.id, r.role_code, r.role_name
order by r.role_code asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	descriptions := roleDescriptionMap()
	items := make([]RoleSummary, 0)
	for rows.Next() {
		var item RoleSummary
		var rawPermissions string
		if err := rows.Scan(&item.Code, &item.Name, &item.AdminCount, &rawPermissions); err != nil {
			return nil, err
		}
		item.Description = descriptions[item.Code]
		item.Permissions = csvToSortedStrings(rawPermissions)
		item.PermissionCount = len(item.Permissions)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) PermissionCatalog(ctx context.Context) ([]PermissionSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
select permission_code
from admin_permissions
order by permission_code asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PermissionSummary, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		module, action := splitPermissionCode(code)
		items = append(items, PermissionSummary{Code: code, Module: module, Action: action})
	}
	return items, rows.Err()
}

func adminAccountSelect() string {
	return `
select u.id, u.username, u.password_hash, u.status,
       coalesce(string_agg(distinct r.role_code, ','), '') as roles,
       coalesce(string_agg(distinct p.permission_code, ','), '') as permissions
from admin_users u
left join admin_user_roles ur on ur.admin_user_id = u.id
left join admin_roles r on r.id = ur.role_id
left join admin_role_permissions rp on rp.role_id = r.id
left join admin_permissions p on p.id = rp.permission_id`
}

func replaceAccountRoles(ctx context.Context, tx *sql.Tx, adminUserID int64, roles []string) error {
	if _, err := tx.ExecContext(ctx, `delete from admin_user_roles where admin_user_id = $1`, adminUserID); err != nil {
		return err
	}
	for _, role := range roles {
		if _, err := tx.ExecContext(ctx, `
insert into admin_user_roles (admin_user_id, role_id, created_at)
select $1, id, now()
from admin_roles
where role_code = $2
on conflict do nothing
`, adminUserID, role); err != nil {
			return err
		}
	}
	return nil
}

func scanStoredAdminAccount(row interface {
	Scan(dest ...any) error
}) (StoredAdminAccount, error) {
	var account StoredAdminAccount
	var rawRoles string
	var rawPermissions string
	if err := row.Scan(&account.User.ID, &account.User.Username, &account.PasswordHash, &account.User.Status, &rawRoles, &rawPermissions); err != nil {
		return StoredAdminAccount{}, err
	}
	account.User.Roles = csvToSortedStrings(rawRoles)
	account.Permissions = csvToSortedStrings(rawPermissions)
	return account, nil
}

func csvToSortedStrings(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && !seen[part] {
			items = append(items, part)
			seen[part] = true
		}
	}
	sort.Strings(items)
	return items
}

func roleDescriptionMap() map[string]string {
	result := make(map[string]string)
	for _, role := range roleDefinitions() {
		result[role.Code] = role.Description
	}
	return result
}
