package adminauth

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestVerifyPassword(t *testing.T) {
	hash := "pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5"
	if !verifyPassword("admin123", hash) {
		t.Fatal("expected password to match hash")
	}
	if verifyPassword("wrong-password", hash) {
		t.Fatal("expected wrong password to fail")
	}
	if verifyPassword("admin123", "admin123") {
		t.Fatal("expected plaintext password storage to be rejected")
	}
	if verifyPassword("admin123", "pbkdf2-sha256$bad$00$00") {
		t.Fatal("expected invalid iteration count to fail")
	}
}

func TestAdminLoginUsesPasswordHash(t *testing.T) {
	service := NewService()
	if _, err := service.Login(LoginRequest{Username: "admin", Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	resp, err := service.Login(LoginRequest{Username: "admin", Password: "admin123"})
	if err != nil {
		t.Fatalf("expected login success: %v", err)
	}
	if resp.Token == "" || !hasPermission(resp.Permissions, "operation_log:view_full") {
		t.Fatalf("expected token and permissions: %+v", resp)
	}
	if !hasPermission(resp.Permissions, "admin_user:update") {
		t.Fatalf("expected super admin to update admin users: %+v", resp.Permissions)
	}
}

func TestBuiltinRolePermissionBoundaries(t *testing.T) {
	service := NewService()
	analyst, err := service.Login(LoginRequest{Username: "data_analyst", Password: "admin123"})
	if err != nil {
		t.Fatalf("expected data analyst login success: %v", err)
	}
	if !hasPermission(analyst.Permissions, "data:behavior:read") || !hasPermission(analyst.Permissions, "testcase:read") {
		t.Fatalf("expected data analyst read permissions: %+v", analyst.Permissions)
	}
	for _, forbidden := range []string{"admin_user:update", "redemption:manage", "system_config:update", "operation_log:view_full", "testcase:manage"} {
		if hasPermission(analyst.Permissions, forbidden) {
			t.Fatalf("data analyst must not have %s: %+v", forbidden, analyst.Permissions)
		}
	}
	if hasPermission(analyst.Permissions, "identity:sensitive:read") || hasPermission(analyst.Permissions, "profile:sensitive:read") {
		t.Fatalf("data analyst must not have sensitive identity/profile permissions: %+v", analyst.Permissions)
	}

	userManager, err := service.Login(LoginRequest{Username: "user_manager", Password: "admin123"})
	if err != nil {
		t.Fatalf("expected user manager login success: %v", err)
	}
	for _, required := range []string{"identity:read", "identity:sensitive:read", "profile:read", "profile:sensitive:read"} {
		if !hasPermission(userManager.Permissions, required) {
			t.Fatalf("user manager must have %s: %+v", required, userManager.Permissions)
		}
	}

	operator, err := service.Login(LoginRequest{Username: "operator", Password: "admin123"})
	if err != nil {
		t.Fatalf("expected operator login success: %v", err)
	}
	for _, forbidden := range []string{"admin_user:update", "profile:read", "operation_log:view_full", "testcase:read", "testcase:manage"} {
		if hasPermission(operator.Permissions, forbidden) {
			t.Fatalf("operator must not have %s: %+v", forbidden, operator.Permissions)
		}
	}
}

func TestSubmitAdminApplication(t *testing.T) {
	service := NewService()
	item, err := service.SubmitAdminApplication(SubmitAdminApplicationRequest{
		Name:        "运营同事",
		Contact:     "wx-admin",
		DesiredRole: "operation_manager",
		Reason:      "处理日常运营",
	})
	if err != nil {
		t.Fatalf("expected submit admin application: %v", err)
	}
	if item.ID == 0 || item.Status != "pending" || item.DesiredRole != "operation_manager" {
		t.Fatalf("unexpected application: %+v", item)
	}
	items, err := service.AdminApplications()
	if err != nil {
		t.Fatalf("expected list admin applications: %v", err)
	}
	if len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("expected submitted application in list: %+v", items)
	}
	if _, err := service.SubmitAdminApplication(SubmitAdminApplicationRequest{Name: "root", Contact: "wx", DesiredRole: "super_admin"}); !errors.Is(err, ErrInvalidAdminInput) {
		t.Fatalf("expected super admin application to be rejected, got %v", err)
	}
}

func TestAdminSeedGrantsLateSuperAdminPermissions(t *testing.T) {
	body, err := os.ReadFile("../../../../db/seeds/admin_roles_permissions.sql")
	if err != nil {
		t.Fatalf("read admin seed: %v", err)
	}
	seed := string(body)
	latePermissions := []string{
		"im:room:read",
		"im:room:archive",
		"im:room:retry_create",
		"system_config:read",
		"invite_code:read",
		"invite_code:manage",
	}

	blocks := strings.Split(seed, "insert into admin_role_permissions(role_id, permission_id)")
	for _, permission := range latePermissions {
		granted := false
		needle := "'" + permission + "'"
		for _, block := range blocks {
			if strings.Contains(block, "where r.role_code = 'super_admin'") && strings.Contains(block, needle) {
				granted = true
				break
			}
		}
		if !granted {
			t.Fatalf("expected admin seed to grant %s to super_admin after late permission inserts", permission)
		}
	}
}

func TestServiceWithRepositoryUsesPersistentAccounts(t *testing.T) {
	repo := newFakeAdminRepository()
	service := NewServiceWithRepository(repo)

	resp, err := service.Login(LoginRequest{Username: "admin", Password: "admin123"})
	if err != nil {
		t.Fatalf("expected repository admin login success: %v", err)
	}
	if resp.AdminUser.ID != 10 || !hasPermission(resp.Permissions, "admin_user:create") {
		t.Fatalf("expected repository account and permissions: %+v", resp)
	}

	created, err := service.CreateAdminUser(CreateAdminUserRequest{
		Username: "finance_ops",
		Password: "finance123",
		Roles:    []string{"finance_manager"},
		Status:   "active",
	})
	if err != nil {
		t.Fatalf("expected create through repository: %v", err)
	}
	if created.ID == 0 || !hasPermission(created.Permissions, "revenue:template:update") {
		t.Fatalf("expected finance permissions: %+v", created)
	}

	login, err := service.Login(LoginRequest{Username: "finance_ops", Password: "finance123"})
	if err != nil {
		t.Fatalf("expected created admin login success: %v", err)
	}
	if _, _, ok := service.Session(login.Token); !ok {
		t.Fatal("expected session to be active before update")
	}

	updated, err := service.UpdateAdminUser(created.ID, UpdateAdminUserRequest{Roles: []string{"customer_manager"}, Status: "disabled"})
	if err != nil {
		t.Fatalf("expected update through repository: %v", err)
	}
	if updated.Status != "disabled" || !hasPermission(updated.Permissions, "report:handle") {
		t.Fatalf("expected customer manager disabled account: %+v", updated)
	}
	if _, _, ok := service.Session(login.Token); ok {
		t.Fatal("expected updating admin to expire old sessions")
	}
	if _, err := service.Login(LoginRequest{Username: "finance_ops", Password: "finance123"}); !errors.Is(err, ErrAdminDisabled) {
		t.Fatalf("expected disabled login error, got %v", err)
	}
}

func TestSuperAdminAlwaysHasAllPermissionsWhenDatabaseBindingsAreIncomplete(t *testing.T) {
	repo := newFakeAdminRepository()
	account := repo.accounts[10]
	account.Permissions = nil
	repo.accounts[10] = account
	service := NewServiceWithRepository(repo)

	login, err := service.Login(LoginRequest{Username: "admin", Password: "admin123"})
	if err != nil {
		t.Fatalf("expected super admin login success: %v", err)
	}
	if !hasPermission(login.Permissions, "game:create_admin") {
		t.Fatalf("expected login response to expose complete super admin permissions: %+v", login.Permissions)
	}
	if adminID, ok := service.HasPermission(login.Token, "game:create_admin"); !ok || adminID != account.User.ID {
		t.Fatalf("expected super admin permission bypass, adminID=%d ok=%v", adminID, ok)
	}
	tree, err := service.Permissions(login.Token)
	if err != nil {
		t.Fatalf("expected permission tree: %v", err)
	}
	if !hasPermission(tree.Permissions, "game:create_admin") {
		t.Fatalf("expected permission endpoint to expose complete super admin permissions: %+v", tree.Permissions)
	}
}

func hasPermission(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type fakeAdminRepository struct {
	accounts map[int64]StoredAdminAccount
	nextID   int64
}

func newFakeAdminRepository() *fakeAdminRepository {
	passwordHash := "pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5"
	return &fakeAdminRepository{
		accounts: map[int64]StoredAdminAccount{
			10: {
				User:         AdminUser{ID: 10, Username: "admin", Status: "active", Roles: []string{"super_admin"}},
				PasswordHash: passwordHash,
				Permissions:  defaultPermissions(),
			},
		},
		nextID: 11,
	}
}

func (r *fakeAdminRepository) FindAccountByUsername(ctx context.Context, username string) (StoredAdminAccount, bool, error) {
	for _, account := range r.accounts {
		if account.User.Username == username {
			return cloneStoredAccount(account), true, nil
		}
	}
	return StoredAdminAccount{}, false, nil
}

func (r *fakeAdminRepository) FindAccountByID(ctx context.Context, id int64) (StoredAdminAccount, bool, error) {
	account, ok := r.accounts[id]
	return cloneStoredAccount(account), ok, nil
}

func (r *fakeAdminRepository) ListAccounts(ctx context.Context) ([]StoredAdminAccount, error) {
	items := make([]StoredAdminAccount, 0, len(r.accounts))
	for _, account := range r.accounts {
		items = append(items, cloneStoredAccount(account))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].User.ID < items[j].User.ID
	})
	return items, nil
}

func (r *fakeAdminRepository) CreateAccount(ctx context.Context, account StoredAdminAccount) (StoredAdminAccount, error) {
	if _, exists, _ := r.FindAccountByUsername(ctx, account.User.Username); exists {
		return StoredAdminAccount{}, ErrAdminExists
	}
	account.User.ID = r.nextID
	r.nextID++
	r.accounts[account.User.ID] = cloneStoredAccount(account)
	return cloneStoredAccount(account), nil
}

func (r *fakeAdminRepository) UpdateAccount(ctx context.Context, account StoredAdminAccount) (StoredAdminAccount, error) {
	if _, ok := r.accounts[account.User.ID]; !ok {
		return StoredAdminAccount{}, ErrAdminNotFound
	}
	r.accounts[account.User.ID] = cloneStoredAccount(account)
	return cloneStoredAccount(account), nil
}

func (r *fakeAdminRepository) RoleSummaries(ctx context.Context) ([]RoleSummary, error) {
	roles := roleDefinitions()
	for i := range roles {
		for _, account := range r.accounts {
			for _, role := range account.User.Roles {
				if role == roles[i].Code {
					roles[i].AdminCount++
				}
			}
		}
		roles[i].PermissionCount = len(roles[i].Permissions)
	}
	return roles, nil
}

func (r *fakeAdminRepository) PermissionCatalog(ctx context.Context) ([]PermissionSummary, error) {
	values := make([]string, 0)
	for _, role := range roleDefinitions() {
		values = append(values, role.Permissions...)
	}
	values = uniqueSortedStrings(values)
	items := make([]PermissionSummary, 0, len(values))
	for _, value := range values {
		module, action := splitPermissionCode(value)
		items = append(items, PermissionSummary{Code: value, Module: module, Action: action})
	}
	return items, nil
}

func cloneStoredAccount(account StoredAdminAccount) StoredAdminAccount {
	return StoredAdminAccount{
		User: AdminUser{
			ID:       account.User.ID,
			Username: account.User.Username,
			Status:   account.User.Status,
			Roles:    append([]string(nil), account.User.Roles...),
		},
		PasswordHash: account.PasswordHash,
		Permissions:  append([]string(nil), account.Permissions...),
	}
}
