package adminauth

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"
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
	rows, err := tx.QueryContext(ctx, `
select u.id
from admin_users u
join admin_user_roles ur on ur.admin_user_id = u.id
join admin_roles ar on ar.id = ur.role_id
where u.status = 'active' and ar.role_code = 'super_admin'
for update of u
`)
	if err != nil {
		return StoredAdminAccount{}, err
	}
	activeSuperIDs := make([]int64, 0)
	for rows.Next() {
		var id int64
		if scanErr := rows.Scan(&id); scanErr != nil {
			rows.Close()
			return StoredAdminAccount{}, scanErr
		}
		activeSuperIDs = append(activeSuperIDs, id)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		rows.Close()
		return StoredAdminAccount{}, rowsErr
	}
	rows.Close()
	targetIsLastSuper := len(activeSuperIDs) == 1 && activeSuperIDs[0] == account.User.ID
	if targetIsLastSuper && (account.User.Status != "active" || !containsRole(account.User.Roles, "super_admin")) {
		return StoredAdminAccount{}, ErrLastSuperAdmin
	}
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

func (r *SQLRepository) CreateApplication(ctx context.Context, application AdminApplication) (AdminApplication, error) {
	var id int64
	if err := r.db.QueryRowContext(ctx, `
insert into admin_applications (name, contact, desired_role, reason, status, created_at, updated_at)
values ($1,$2,$3,$4,'pending',now(),now())
returning id
`, application.Name, application.Contact, application.DesiredRole, application.Reason).Scan(&id); err != nil {
		return AdminApplication{}, err
	}
	return r.findApplicationByID(ctx, r.db, id)
}

func (r *SQLRepository) ListApplications(ctx context.Context) ([]AdminApplication, error) {
	rows, err := r.db.QueryContext(ctx, adminApplicationSelect()+` order by a.id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AdminApplication, 0)
	for rows.Next() {
		item, scanErr := scanAdminApplication(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ReviewApplication(ctx context.Context, id int64, review AdminApplicationReview) (AdminApplication, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return AdminApplication{}, err
	}
	defer tx.Rollback()

	var currentStatus string
	var desiredRole string
	if err := tx.QueryRowContext(ctx, `
select status, desired_role
from admin_applications
where id = $1
for update
`, id).Scan(&currentStatus, &desiredRole); errors.Is(err, sql.ErrNoRows) {
		return AdminApplication{}, ErrApplicationNotFound
	} else if err != nil {
		return AdminApplication{}, err
	}
	if currentStatus != "pending" {
		return AdminApplication{}, ErrApplicationProcessed
	}

	linkedAdminUserID := review.LinkedAdminUserID
	if review.Status == "approved" {
		if linkedAdminUserID > 0 {
			if err := tx.QueryRowContext(ctx, `select id from admin_users where id = $1 and status = 'active' for update`, linkedAdminUserID).Scan(&linkedAdminUserID); errors.Is(err, sql.ErrNoRows) {
				return AdminApplication{}, ErrAdminNotFound
			} else if err != nil {
				return AdminApplication{}, err
			}
		} else {
			if review.NewAccount == nil {
				return AdminApplication{}, ErrInvalidAdminInput
			}
			if err := tx.QueryRowContext(ctx, `
insert into admin_users (username, password_hash, status, created_at, updated_at)
values ($1,$2,'active',now(),now())
on conflict (username) do nothing
returning id
`, review.NewAccount.User.Username, review.NewAccount.PasswordHash).Scan(&linkedAdminUserID); errors.Is(err, sql.ErrNoRows) {
				return AdminApplication{}, ErrAdminExists
			} else if err != nil {
				return AdminApplication{}, err
			}
		}
		if _, err := tx.ExecContext(ctx, `
insert into admin_user_roles (admin_user_id, role_id, created_at)
select $1, id, now()
from admin_roles
where role_code = $2
on conflict do nothing
`, linkedAdminUserID, desiredRole); err != nil {
			return AdminApplication{}, err
		}
		var roleLinked bool
		if err := tx.QueryRowContext(ctx, `
select exists (
  select 1
  from admin_user_roles ur
  join admin_roles r on r.id = ur.role_id
  where ur.admin_user_id = $1 and r.role_code = $2
)
`, linkedAdminUserID, desiredRole).Scan(&roleLinked); err != nil {
			return AdminApplication{}, err
		}
		if !roleLinked {
			return AdminApplication{}, ErrInvalidAdminInput
		}
	} else {
		linkedAdminUserID = 0
	}

	result, err := tx.ExecContext(ctx, `
update admin_applications
set status = $2,
    review_remark = $3,
    reviewed_by = $4,
    reviewed_at = now(),
    linked_admin_user_id = $5,
    updated_at = now()
where id = $1 and status = 'pending'
`, id, review.Status, review.Remark, review.ReviewerID, nullAdminID(linkedAdminUserID))
	if err != nil {
		return AdminApplication{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return AdminApplication{}, err
	}
	if affected != 1 {
		return AdminApplication{}, ErrApplicationProcessed
	}
	if err := tx.Commit(); err != nil {
		return AdminApplication{}, err
	}
	return r.findApplicationByID(ctx, r.db, id)
}

func (r *SQLRepository) findApplicationByID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, id int64) (AdminApplication, error) {
	item, err := scanAdminApplication(queryer.QueryRowContext(ctx, adminApplicationSelect()+` where a.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return AdminApplication{}, ErrApplicationNotFound
	}
	return item, err
}

func adminApplicationSelect() string {
	return `
select a.id, a.name, a.contact, a.desired_role, a.reason, a.status,
       coalesce(a.review_remark, ''), coalesce(a.reviewed_by, 0),
       coalesce(reviewer.username, ''), a.reviewed_at,
       coalesce(a.linked_admin_user_id, 0), coalesce(linked.username, ''),
       a.created_at, a.updated_at
from admin_applications a
left join admin_users reviewer on reviewer.id = a.reviewed_by
left join admin_users linked on linked.id = a.linked_admin_user_id`
}

func scanAdminApplication(row interface {
	Scan(dest ...any) error
}) (AdminApplication, error) {
	var item AdminApplication
	var reviewedAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(
		&item.ID, &item.Name, &item.Contact, &item.DesiredRole, &item.Reason, &item.Status,
		&item.ReviewRemark, &item.ReviewedBy, &item.ReviewerUsername, &reviewedAt,
		&item.LinkedAdminUserID, &item.LinkedAdminUsername, &createdAt, &updatedAt,
	); err != nil {
		return AdminApplication{}, err
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	if reviewedAt.Valid {
		item.ReviewedAt = reviewedAt.Time.Format(time.RFC3339)
	}
	return item, nil
}

func nullAdminID(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
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
