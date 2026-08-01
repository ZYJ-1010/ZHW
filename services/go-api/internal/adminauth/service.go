package adminauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCredentials   = errors.New("invalid admin credentials")
	ErrAdminDisabled        = errors.New("admin disabled")
	ErrInvalidToken         = errors.New("invalid admin token")
	ErrAdminNotFound        = errors.New("admin not found")
	ErrAdminExists          = errors.New("admin exists")
	ErrInvalidAdminInput    = errors.New("invalid admin input")
	ErrApplicationNotFound  = errors.New("admin application not found")
	ErrApplicationProcessed = errors.New("admin application already processed")
	ErrAdminSelfMutation    = errors.New("admin cannot change own roles or status")
	ErrLastSuperAdmin       = errors.New("last active super admin must be preserved")
)

type AdminUser struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

type AdminUserSummary struct {
	ID              int64    `json:"id"`
	Username        string   `json:"username"`
	Status          string   `json:"status"`
	Roles           []string `json:"roles"`
	PermissionCount int      `json:"permissionCount"`
	Permissions     []string `json:"permissions"`
}

type AdminApplication struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Contact             string `json:"contact"`
	DesiredRole         string `json:"desiredRole"`
	Reason              string `json:"reason"`
	Status              string `json:"status"`
	ReviewRemark        string `json:"reviewRemark,omitempty"`
	ReviewedBy          int64  `json:"reviewedBy,omitempty"`
	ReviewerUsername    string `json:"reviewerUsername,omitempty"`
	ReviewedAt          string `json:"reviewedAt,omitempty"`
	LinkedAdminUserID   int64  `json:"linkedAdminUserId,omitempty"`
	LinkedAdminUsername string `json:"linkedAdminUsername,omitempty"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type RoleSummary struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	AdminCount      int      `json:"adminCount"`
	PermissionCount int      `json:"permissionCount"`
	Permissions     []string `json:"permissions"`
}

type PermissionSummary struct {
	Code   string `json:"code"`
	Module string `json:"module"`
	Action string `json:"action"`
}

type Session struct {
	Token       string   `json:"token"`
	AdminUserID int64    `json:"adminUserId"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	ExpiresAt   string   `json:"expiresAt"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token       string    `json:"token"`
	ExpiresAt   string    `json:"expiresAt"`
	AdminUser   AdminUser `json:"adminUser"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
}

type CreateAdminUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
	Status   string   `json:"status"`
}

type UpdateAdminUserRequest struct {
	Roles  []string `json:"roles"`
	Status string   `json:"status"`
}

type SubmitAdminApplicationRequest struct {
	Name        string `json:"name"`
	Contact     string `json:"contact"`
	DesiredRole string `json:"desiredRole"`
	Reason      string `json:"reason"`
}

type ReviewAdminApplicationRequest struct {
	Action      string `json:"action"`
	Remark      string `json:"remark"`
	AdminUserID int64  `json:"adminUserId"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type AdminApplicationReview struct {
	Status            string
	Remark            string
	ReviewerID        int64
	LinkedAdminUserID int64
	NewAccount        *StoredAdminAccount
}

type PermissionNode struct {
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Type     string           `json:"type"`
	Children []PermissionNode `json:"children,omitempty"`
}

type PermissionTree struct {
	AdminUser   AdminUser        `json:"adminUser"`
	Roles       []string         `json:"roles"`
	Permissions []string         `json:"permissions"`
	Menus       []PermissionNode `json:"menus"`
	Buttons     []string         `json:"buttons"`
	Apis        []string         `json:"apis"`
}

type adminAccount struct {
	user          AdminUser
	passwordHash  string
	permissions   []string
	permissionSet map[string]bool
}

type StoredAdminAccount struct {
	User         AdminUser
	PasswordHash string
	Permissions  []string
}

type Repository interface {
	FindAccountByUsername(ctx context.Context, username string) (StoredAdminAccount, bool, error)
	FindAccountByID(ctx context.Context, id int64) (StoredAdminAccount, bool, error)
	ListAccounts(ctx context.Context) ([]StoredAdminAccount, error)
	CreateAccount(ctx context.Context, account StoredAdminAccount) (StoredAdminAccount, error)
	UpdateAccount(ctx context.Context, account StoredAdminAccount) (StoredAdminAccount, error)
	CreateApplication(ctx context.Context, application AdminApplication) (AdminApplication, error)
	ListApplications(ctx context.Context) ([]AdminApplication, error)
	ReviewApplication(ctx context.Context, id int64, review AdminApplicationReview) (AdminApplication, error)
	RoleSummaries(ctx context.Context) ([]RoleSummary, error)
	PermissionCatalog(ctx context.Context) ([]PermissionSummary, error)
}

type Service struct {
	mu                sync.RWMutex
	accounts          map[string]adminAccount
	sessions          map[string]Session
	nextAdminID       int64
	applications      map[int64]AdminApplication
	nextApplicationID int64
	repository        Repository
}

func NewService() *Service {
	passwordHash := "pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5"
	return &Service{
		accounts: map[string]adminAccount{
			"admin":        newAdminAccount(1, "admin", []string{"super_admin"}, passwordHash, defaultPermissions()),
			"data_analyst": newAdminAccount(2, "data_analyst", []string{"data_analyst"}, passwordHash, dataAnalystPermissions()),
			"operator":     newAdminAccount(3, "operator", []string{"operation_manager"}, passwordHash, operatorPermissions()),
			"user_manager": newAdminAccount(4, "user_manager", []string{"user_manager"}, passwordHash, userManagerPermissions()),
			"finance":      newAdminAccount(5, "finance", []string{"finance_manager"}, passwordHash, financeManagerPermissions()),
			"customer":     newAdminAccount(6, "customer", []string{"customer_manager"}, passwordHash, customerManagerPermissions()),
			"game_manager": newAdminAccount(7, "game_manager", []string{"game_manager"}, passwordHash, gameManagerPermissions()),
		},
		sessions:          make(map[string]Session),
		nextAdminID:       8,
		applications:      make(map[int64]AdminApplication),
		nextApplicationID: 1,
	}
}

func NewServiceWithRepository(repository Repository) *Service {
	service := NewService()
	service.repository = repository
	return service
}

func newAdminAccount(id int64, username string, roles []string, passwordHash string, permissions []string) adminAccount {
	return adminAccount{
		user: AdminUser{
			ID:       id,
			Username: username,
			Status:   "active",
			Roles:    append([]string(nil), roles...),
		},
		passwordHash:  passwordHash,
		permissions:   append([]string(nil), permissions...),
		permissionSet: toSet(permissions),
	}
}

func (s *Service) Login(req LoginRequest) (LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	account, ok, err := s.accountByUsername(context.Background(), username)
	if err != nil {
		return LoginResponse{}, err
	}
	if !ok || !verifyPassword(req.Password, account.passwordHash) {
		return LoginResponse{}, ErrInvalidCredentials
	}
	if account.user.Status != "active" {
		return LoginResponse{}, ErrAdminDisabled
	}
	token, err := randomAdminToken(24)
	if err != nil {
		return LoginResponse{}, err
	}
	expiresAt := time.Now().Add(8 * time.Hour).Format(time.RFC3339)
	permissions := effectivePermissions(account.user.Roles, account.permissions)
	session := Session{
		Token:       token,
		AdminUserID: account.user.ID,
		Roles:       append([]string(nil), account.user.Roles...),
		Permissions: append([]string(nil), permissions...),
		ExpiresAt:   expiresAt,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = session
	return LoginResponse{
		Token:       token,
		ExpiresAt:   expiresAt,
		AdminUser:   account.user,
		Roles:       append([]string(nil), account.user.Roles...),
		Permissions: append([]string(nil), permissions...),
	}, nil
}

func (s *Service) Session(token string) (Session, AdminUser, bool) {
	token = strings.TrimSpace(token)
	s.mu.RLock()
	session, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return Session{}, AdminUser{}, false
	}
	expiresAt, err := time.Parse(time.RFC3339, session.ExpiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return Session{}, AdminUser{}, false
	}
	account, ok, err := s.accountByID(context.Background(), session.AdminUserID)
	if err != nil || !ok {
		return Session{}, AdminUser{}, false
	}
	return session, account.user, true
}

func (s *Service) HasPermission(token string, permission string) (int64, bool) {
	session, user, ok := s.Session(token)
	if !ok {
		return 0, false
	}
	if containsRole(user.Roles, "super_admin") || containsRole(session.Roles, "super_admin") {
		return session.AdminUserID, true
	}
	for _, item := range session.Permissions {
		if item == permission {
			return session.AdminUserID, true
		}
	}
	return session.AdminUserID, false
}

func (s *Service) Permissions(token string) (PermissionTree, error) {
	session, user, ok := s.Session(token)
	if !ok {
		return PermissionTree{}, ErrInvalidToken
	}
	permissions := effectivePermissions(session.Roles, session.Permissions)
	sort.Strings(permissions)
	return PermissionTree{
		AdminUser:   user,
		Roles:       append([]string(nil), session.Roles...),
		Permissions: permissions,
		Menus:       menuNodesForPermissions(permissions),
		Buttons:     permissions,
		Apis:        permissions,
	}, nil
}

func effectivePermissions(roles []string, permissions []string) []string {
	if containsRole(roles, "super_admin") {
		return append([]string(nil), defaultPermissions()...)
	}
	return append([]string(nil), permissions...)
}

func containsRole(roles []string, want string) bool {
	for _, role := range roles {
		if role == want {
			return true
		}
	}
	return false
}

func (s *Service) AdminUsers() ([]AdminUserSummary, error) {
	if s.repository != nil {
		accounts, err := s.repository.ListAccounts(context.Background())
		if err != nil {
			return nil, err
		}
		items := make([]AdminUserSummary, 0, len(accounts))
		for _, account := range accounts {
			items = append(items, summaryFromAccount(adminAccountFromStored(account)))
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].ID < items[j].ID
		})
		return items, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]AdminUserSummary, 0, len(s.accounts))
	for _, account := range s.accounts {
		permissions := cloneSortedStrings(account.permissions)
		items = append(items, AdminUserSummary{
			ID:              account.user.ID,
			Username:        account.user.Username,
			Status:          account.user.Status,
			Roles:           cloneSortedStrings(account.user.Roles),
			PermissionCount: len(permissions),
			Permissions:     permissions,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (s *Service) SubmitAdminApplication(req SubmitAdminApplicationRequest) (AdminApplication, error) {
	name := strings.TrimSpace(req.Name)
	contact := strings.TrimSpace(req.Contact)
	role := strings.TrimSpace(req.DesiredRole)
	reason := strings.TrimSpace(req.Reason)
	if name == "" || contact == "" || role == "" {
		return AdminApplication{}, ErrInvalidAdminInput
	}
	if len([]rune(name)) > 64 || len([]rune(contact)) > 120 || len([]rune(reason)) > 500 {
		return AdminApplication{}, ErrInvalidAdminInput
	}
	roles, err := normalizeRoles([]string{role})
	if err != nil || len(roles) != 1 || roles[0] == "super_admin" {
		return AdminApplication{}, ErrInvalidAdminInput
	}
	now := time.Now().Format(time.RFC3339)
	item := AdminApplication{
		Name:        name,
		Contact:     contact,
		DesiredRole: roles[0],
		Reason:      reason,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if s.repository != nil {
		return s.repository.CreateApplication(context.Background(), item)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = s.nextApplicationID
	s.nextApplicationID++
	s.applications[item.ID] = item
	return item, nil
}

func (s *Service) AdminApplications() ([]AdminApplication, error) {
	if s.repository != nil {
		return s.repository.ListApplications(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]AdminApplication, 0, len(s.applications))
	for _, item := range s.applications {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID > items[j].ID
	})
	return items, nil
}

func (s *Service) ReviewAdminApplication(id int64, reviewerID int64, req ReviewAdminApplicationRequest) (AdminApplication, error) {
	if id <= 0 || reviewerID <= 0 {
		return AdminApplication{}, ErrInvalidAdminInput
	}
	status, err := normalizeApplicationAction(req.Action)
	if err != nil {
		return AdminApplication{}, err
	}
	remark := strings.TrimSpace(req.Remark)
	if len([]rune(remark)) > 500 || ((status == "rejected" || status == "closed") && remark == "") {
		return AdminApplication{}, ErrInvalidAdminInput
	}

	review := AdminApplicationReview{Status: status, Remark: remark, ReviewerID: reviewerID}
	if status == "approved" {
		username := strings.TrimSpace(req.Username)
		password := strings.TrimSpace(req.Password)
		hasExistingAccount := req.AdminUserID > 0
		hasNewAccount := username != "" || password != ""
		if hasExistingAccount == hasNewAccount {
			return AdminApplication{}, ErrInvalidAdminInput
		}
		if hasExistingAccount {
			if req.AdminUserID == reviewerID {
				return AdminApplication{}, ErrAdminSelfMutation
			}
			account, found, findErr := s.accountByID(context.Background(), req.AdminUserID)
			if findErr != nil {
				return AdminApplication{}, findErr
			}
			if !found {
				return AdminApplication{}, ErrAdminNotFound
			}
			if account.user.Status != "active" {
				return AdminApplication{}, ErrAdminDisabled
			}
			review.LinkedAdminUserID = req.AdminUserID
		} else {
			if username == "" || !validAdminPassword(req.Password) || len([]rune(username)) > 64 || strings.ContainsAny(username, " \t\r\n") {
				return AdminApplication{}, ErrInvalidAdminInput
			}
			passwordHash, hashErr := hashPassword(req.Password)
			if hashErr != nil {
				return AdminApplication{}, hashErr
			}
			review.NewAccount = &StoredAdminAccount{
				User:         AdminUser{Username: username, Status: "active"},
				PasswordHash: passwordHash,
			}
		}
	} else if req.AdminUserID > 0 || strings.TrimSpace(req.Username) != "" || strings.TrimSpace(req.Password) != "" {
		return AdminApplication{}, ErrInvalidAdminInput
	}

	if s.repository != nil {
		item, reviewErr := s.repository.ReviewApplication(context.Background(), id, review)
		if reviewErr != nil {
			return AdminApplication{}, reviewErr
		}
		if item.LinkedAdminUserID > 0 {
			s.mu.Lock()
			s.expireSessionsForAdminLocked(item.LinkedAdminUserID)
			s.mu.Unlock()
		}
		return item, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.applications[id]
	if !ok {
		return AdminApplication{}, ErrApplicationNotFound
	}
	if item.Status != "pending" {
		return AdminApplication{}, ErrApplicationProcessed
	}
	if status == "approved" {
		linkedID, linkedUsername, reviewErr := s.applyApplicationAccountLocked(item.DesiredRole, review)
		if reviewErr != nil {
			return AdminApplication{}, reviewErr
		}
		item.LinkedAdminUserID = linkedID
		item.LinkedAdminUsername = linkedUsername
		s.expireSessionsForAdminLocked(linkedID)
	}
	now := time.Now().Format(time.RFC3339)
	item.Status = status
	item.ReviewRemark = remark
	item.ReviewedBy = reviewerID
	for _, account := range s.accounts {
		if account.user.ID == reviewerID {
			item.ReviewerUsername = account.user.Username
			break
		}
	}
	item.ReviewedAt = now
	item.UpdatedAt = now
	s.applications[id] = item
	return item, nil
}

func (s *Service) applyApplicationAccountLocked(desiredRole string, review AdminApplicationReview) (int64, string, error) {
	if review.LinkedAdminUserID > 0 {
		for username, account := range s.accounts {
			if account.user.ID != review.LinkedAdminUserID {
				continue
			}
			account.user.Roles = uniqueSortedStrings(append(account.user.Roles, desiredRole))
			account.permissions = permissionsForRoles(account.user.Roles)
			account.permissionSet = toSet(account.permissions)
			s.accounts[username] = account
			return account.user.ID, username, nil
		}
		return 0, "", ErrAdminNotFound
	}
	if review.NewAccount == nil {
		return 0, "", ErrInvalidAdminInput
	}
	username := review.NewAccount.User.Username
	if _, exists := s.accounts[username]; exists {
		return 0, "", ErrAdminExists
	}
	account := newAdminAccount(s.nextAdminID, username, []string{desiredRole}, review.NewAccount.PasswordHash, permissionsForRoles([]string{desiredRole}))
	s.nextAdminID++
	s.accounts[username] = account
	return account.user.ID, username, nil
}

func normalizeApplicationAction(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "approve", "approved":
		return "approved", nil
	case "reject", "rejected":
		return "rejected", nil
	case "close", "closed":
		return "closed", nil
	default:
		return "", ErrInvalidAdminInput
	}
}

func (s *Service) CreateAdminUser(req CreateAdminUserRequest) (AdminUserSummary, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || strings.ContainsAny(username, " \t\r\n") || !validAdminPassword(req.Password) {
		return AdminUserSummary{}, ErrInvalidAdminInput
	}
	roles, err := normalizeRoles(req.Roles)
	if err != nil {
		return AdminUserSummary{}, err
	}
	status, err := normalizeAdminStatus(firstNonEmptyText(req.Status, "active"))
	if err != nil {
		return AdminUserSummary{}, err
	}
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return AdminUserSummary{}, err
	}
	if s.repository != nil {
		account, err := s.repository.CreateAccount(context.Background(), StoredAdminAccount{
			User: AdminUser{
				Username: username,
				Status:   status,
				Roles:    roles,
			},
			PasswordHash: passwordHash,
			Permissions:  permissionsForRoles(roles),
		})
		if err != nil {
			return AdminUserSummary{}, err
		}
		return summaryFromAccount(adminAccountFromStored(account)), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.accounts[username]; exists {
		return AdminUserSummary{}, ErrAdminExists
	}
	account := newAdminAccount(s.nextAdminID, username, roles, passwordHash, permissionsForRoles(roles))
	account.user.Status = status
	s.nextAdminID++
	s.accounts[username] = account
	return summaryFromAccount(account), nil
}

func (s *Service) UpdateAdminUser(id int64, req UpdateAdminUserRequest) (AdminUserSummary, error) {
	return s.UpdateAdminUserByActor(0, id, req)
}

func (s *Service) UpdateAdminUserByActor(actorID int64, id int64, req UpdateAdminUserRequest) (AdminUserSummary, error) {
	if id <= 0 {
		return AdminUserSummary{}, ErrInvalidAdminInput
	}
	roles, err := normalizeRoles(req.Roles)
	if err != nil {
		return AdminUserSummary{}, err
	}
	status, err := normalizeAdminStatus(req.Status)
	if err != nil {
		return AdminUserSummary{}, err
	}
	if s.repository != nil {
		current, ok, err := s.repository.FindAccountByID(context.Background(), id)
		if err != nil {
			return AdminUserSummary{}, err
		}
		if !ok {
			return AdminUserSummary{}, ErrAdminNotFound
		}
		if actorID == id && (current.User.Status != status || !sameAdminRoles(current.User.Roles, roles)) {
			return AdminUserSummary{}, ErrAdminSelfMutation
		}
		current.User.Roles = roles
		current.User.Status = status
		current.Permissions = permissionsForRoles(roles)
		updated, err := s.repository.UpdateAccount(context.Background(), current)
		if err != nil {
			return AdminUserSummary{}, err
		}
		s.mu.Lock()
		s.expireSessionsForAdminLocked(id)
		s.mu.Unlock()
		return summaryFromAccount(adminAccountFromStored(updated)), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for username, account := range s.accounts {
		if account.user.ID != id {
			continue
		}
		if actorID == id && (account.user.Status != status || !sameAdminRoles(account.user.Roles, roles)) {
			return AdminUserSummary{}, ErrAdminSelfMutation
		}
		if account.user.Status == "active" && containsRole(account.user.Roles, "super_admin") &&
			(status != "active" || !containsRole(roles, "super_admin")) && s.activeSuperAdminCountLocked() <= 1 {
			return AdminUserSummary{}, ErrLastSuperAdmin
		}
		account.user.Roles = roles
		account.user.Status = status
		account.permissions = permissionsForRoles(roles)
		account.permissionSet = toSet(account.permissions)
		s.accounts[username] = account
		s.expireSessionsForAdminLocked(id)
		return summaryFromAccount(account), nil
	}
	return AdminUserSummary{}, ErrAdminNotFound
}

func (s *Service) activeSuperAdminCountLocked() int {
	count := 0
	for _, account := range s.accounts {
		if account.user.Status == "active" && containsRole(account.user.Roles, "super_admin") {
			count++
		}
	}
	return count
}

func sameAdminRoles(left []string, right []string) bool {
	left = uniqueSortedStrings(left)
	right = uniqueSortedStrings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validAdminPassword(value string) bool {
	length := len([]rune(value))
	return length >= 8 && length <= 128 && strings.TrimSpace(value) != ""
}

func (s *Service) Roles() ([]RoleSummary, error) {
	if s.repository != nil {
		return s.repository.RoleSummaries(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := roleDefinitions()
	for i := range items {
		items[i].Permissions = cloneSortedStrings(items[i].Permissions)
		items[i].PermissionCount = len(items[i].Permissions)
		for _, account := range s.accounts {
			for _, role := range account.user.Roles {
				if role == items[i].Code {
					items[i].AdminCount++
				}
			}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Code < items[j].Code
	})
	return items, nil
}

func (s *Service) PermissionCatalog() ([]PermissionSummary, error) {
	if s.repository != nil {
		return s.repository.PermissionCatalog(context.Background())
	}
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

func menuNodesForPermissions(permissions []string) []PermissionNode {
	permissionSet := toSet(permissions)
	definitions := []struct {
		code        string
		name        string
		permissions []string
	}{
		{code: "dashboard", name: "数据看板", permissions: []string{
			"analytics:funnel:view", "analytics:retention:view", "analytics:timeline:view", "user:view", "game:read",
			"identity:read", "role:view", "report:view", "redemption:manage", "invite_code:read", "feedback:view",
			"notification:wechat:view", "admin_user:view",
		}},
		{code: "users", name: "用户管理", permissions: []string{"user:view"}},
		{code: "invites", name: "邀请管理", permissions: []string{"invite_code:read", "invite_code:manage"}},
		{code: "games", name: "组局管理", permissions: []string{"game:read"}},
		{code: "audits", name: "审核中心", permissions: []string{"identity:read", "identity:update", "role:view", "role:update"}},
		{code: "revenue", name: "分润结算", permissions: []string{"revenue:template:view", "revenue:record:view", "revenue:simulate", "revenue:generate", "settlement:offline:create"}},
		{code: "redemption", name: "积分兑换", permissions: []string{"points:read", "redemption:manage"}},
		{code: "members", name: "会员团队", permissions: []string{"team:read", "member_report:read"}},
		{code: "profiles", name: "画像关系", permissions: []string{"connection:read", "profile:read"}},
		{code: "growth", name: "评价成长", permissions: []string{"user:view", "game:view"}},
		{code: "reports", name: "举报申诉", permissions: []string{"report:view", "report:assign", "report:handle", "report:close"}},
		{code: "exports", name: "导出中心", permissions: []string{"report_export:create"}},
		{code: "analytics", name: "数据分析", permissions: []string{"analytics:funnel:view", "analytics:retention:view", "analytics:timeline:view", "data:behavior:read", "ai:data:read", "ai:data:seed"}},
		{code: "delivery", name: "通知交付", permissions: []string{"notification:wechat:view", "delivery:manage", "testcase:read", "testcase:manage"}},
		{code: "admins", name: "管理员权限", permissions: []string{"admin_user:view"}},
		{code: "system", name: "系统配置", permissions: []string{"system_config:read", "system_config:update", "content:sensitive_word:view", "content:risk_log:view", "ai:data:read"}},
		{code: "im", name: "IM 证据", permissions: []string{"im:room:read", "im:message:view_dispute"}},
		{code: "logs", name: "操作日志", permissions: []string{"operation_log:view_self", "operation_log:view_full"}},
	}
	nodes := make([]PermissionNode, 0, len(definitions))
	for _, definition := range definitions {
		for _, permission := range definition.permissions {
			if permissionSet[permission] {
				nodes = append(nodes, PermissionNode{Code: definition.code, Name: definition.name, Type: "menu"})
				break
			}
		}
	}
	return nodes
}

func (s *Service) accountByUsername(ctx context.Context, username string) (adminAccount, bool, error) {
	if s.repository != nil {
		account, ok, err := s.repository.FindAccountByUsername(ctx, username)
		return adminAccountFromStored(account), ok, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	account, ok := s.accounts[username]
	return account, ok, nil
}

func (s *Service) accountByID(ctx context.Context, id int64) (adminAccount, bool, error) {
	if s.repository != nil {
		account, ok, err := s.repository.FindAccountByID(ctx, id)
		return adminAccountFromStored(account), ok, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, account := range s.accounts {
		if account.user.ID == id {
			return account, true, nil
		}
	}
	return adminAccount{}, false, nil
}

func summaryFromAccount(account adminAccount) AdminUserSummary {
	permissions := cloneSortedStrings(account.permissions)
	return AdminUserSummary{
		ID:              account.user.ID,
		Username:        account.user.Username,
		Status:          account.user.Status,
		Roles:           cloneSortedStrings(account.user.Roles),
		PermissionCount: len(permissions),
		Permissions:     permissions,
	}
}

func storedFromAdminAccount(account adminAccount) StoredAdminAccount {
	return StoredAdminAccount{
		User: AdminUser{
			ID:       account.user.ID,
			Username: account.user.Username,
			Status:   account.user.Status,
			Roles:    cloneSortedStrings(account.user.Roles),
		},
		PasswordHash: account.passwordHash,
		Permissions:  cloneSortedStrings(account.permissions),
	}
}

func adminAccountFromStored(account StoredAdminAccount) adminAccount {
	result := adminAccount{
		user: AdminUser{
			ID:       account.User.ID,
			Username: account.User.Username,
			Status:   account.User.Status,
			Roles:    cloneSortedStrings(account.User.Roles),
		},
		passwordHash: account.PasswordHash,
		permissions:  cloneSortedStrings(account.Permissions),
	}
	result.permissionSet = toSet(result.permissions)
	return result
}

func defaultPermissions() []string {
	return []string{
		"admin_user:create",
		"admin_user:view",
		"admin_user:update",
		"ai:data:export",
		"ai:data:read",
		"ai:data:seed",
		"analytics:funnel:view",
		"analytics:retention:view",
		"analytics:timeline:view",
		"connection:read",
		"content:risk_log:view",
		"content:sensitive_word:create",
		"content:sensitive_word:import",
		"content:sensitive_word:view",
		"content:sensitive_word:update",
		"data:behavior:read",
		"delivery:manage",
		"feedback:reply",
		"feedback:view",
		"game:create_admin",
		"game:progress:manage",
		"game:read",
		"game:update_status",
		"game:view",
		"identity:read",
		"identity:sensitive:read",
		"identity:update",
		"im:message:hide",
		"im:message:view_dispute",
		"im:room:archive",
		"im:room:read",
		"im:room:retry_create",
		"invite_code:manage",
		"invite_code:read",
		"member_report:read",
		"notification:wechat:view",
		"operation_log:view_full",
		"operation_log:view_self",
		"points:read",
		"profile:read",
		"profile:sensitive:read",
		"redemption:manage",
		"report:assign",
		"report:handle",
		"report:close",
		"report:view",
		"report_export:create",
		"revenue:freeze",
		"revenue:generate",
		"revenue:record:view",
		"revenue:simulate",
		"revenue:template:update",
		"revenue:template:view",
		"role:view",
		"role:update",
		"settlement:offline:create",
		"system_config:read",
		"system_config:update",
		"team:read",
		"testcase:read",
		"testcase:manage",
		"user:read",
		"user:view",
	}
}

func roleDefinitions() []RoleSummary {
	return []RoleSummary{
		{
			Code:        "super_admin",
			Name:        "超级管理员",
			Description: "全部权限、管理员管理、高风险配置、完整操作日志查看",
			Permissions: defaultPermissions(),
		},
		{
			Code:        "operation_manager",
			Name:        "运营管理",
			Description: "审核、组局、IM 争议、举报申诉和内容安全处理",
			Permissions: operatorPermissions(),
		},
		{
			Code:        "user_manager",
			Name:        "用户管理",
			Description: "用户资料、实名状态、角色状态、邀请绑定关系",
			Permissions: userManagerPermissions(),
		},
		{
			Code:        "finance_manager",
			Name:        "财务管理",
			Description: "分润模板、收益记录、冻结和线下结算登记",
			Permissions: financeManagerPermissions(),
		},
		{
			Code:        "customer_manager",
			Name:        "客户管理",
			Description: "客户跟进、投诉协助、争议证据查看",
			Permissions: customerManagerPermissions(),
		},
		{
			Code:        "data_analyst",
			Name:        "数据分析员",
			Description: "行为、漏斗、留存、固定报表和 AI 数据只读/导出",
			Permissions: dataAnalystPermissions(),
		},
		{
			Code:        "game_manager",
			Name:        "组局管理",
			Description: "组局审核、局状态、后台开局和进度管理",
			Permissions: gameManagerPermissions(),
		},
	}
}

func dataAnalystPermissions() []string {
	return []string{
		"ai:data:read",
		"analytics:funnel:view",
		"analytics:retention:view",
		"analytics:timeline:view",
		"content:risk_log:view",
		"content:sensitive_word:update",
		"data:behavior:read",
		"member_report:read",
		"operation_log:view_self",
		"report:view",
		"report_export:create",
		"testcase:read",
	}
}

func userManagerPermissions() []string {
	return []string{
		"connection:read",
		"identity:read",
		"identity:sensitive:read",
		"identity:update",
		"invite_code:manage",
		"invite_code:read",
		"member_report:read",
		"points:read",
		"operation_log:view_self",
		"profile:read",
		"profile:sensitive:read",
		"redemption:manage",
		"role:update",
		"role:view",
		"team:read",
		"user:read",
		"user:view",
	}
}

func financeManagerPermissions() []string {
	return []string{
		"operation_log:view_self",
		"report_export:create",
		"revenue:freeze",
		"revenue:generate",
		"revenue:record:view",
		"revenue:simulate",
		"revenue:template:update",
		"revenue:template:view",
		"settlement:offline:create",
	}
}

func customerManagerPermissions() []string {
	return []string{
		"im:message:view_dispute",
		"im:room:read",
		"operation_log:view_self",
		"feedback:reply",
		"feedback:view",
		"report:assign",
		"report:close",
		"report:handle",
		"report:view",
		"user:read",
		"user:view",
	}
}

func gameManagerPermissions() []string {
	return []string{
		"game:create_admin",
		"game:progress:manage",
		"game:read",
		"game:update_status",
		"game:view",
		"im:room:read",
		"operation_log:view_self",
	}
}

func permissionsForRoles(roles []string) []string {
	values := make([]string, 0)
	for _, role := range roleDefinitions() {
		for _, code := range roles {
			if code == role.Code {
				values = append(values, role.Permissions...)
			}
		}
	}
	return uniqueSortedStrings(values)
}

func normalizeRoles(values []string) ([]string, error) {
	allowed := make(map[string]bool)
	for _, role := range roleDefinitions() {
		allowed[role.Code] = true
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if !allowed[value] {
			return nil, ErrInvalidAdminInput
		}
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	if len(result) == 0 {
		return nil, ErrInvalidAdminInput
	}
	sort.Strings(result)
	return result, nil
}

func normalizeAdminStatus(value string) (string, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "active", "disabled":
		return value, nil
	default:
		return "", ErrInvalidAdminInput
	}
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (s *Service) expireSessionsForAdminLocked(adminID int64) {
	for token, session := range s.sessions {
		if session.AdminUserID == adminID {
			delete(s.sessions, token)
		}
	}
}

func operatorPermissions() []string {
	return []string{
		"game:progress:manage",
		"game:read",
		"game:view",
		"content:risk_log:view",
		"content:sensitive_word:view",
		"content:sensitive_word:update",
		"notification:wechat:view",
		"operation_log:view_self",
		"feedback:reply",
		"feedback:view",
		"report:assign",
		"report:handle",
		"report:close",
		"report:view",
		"user:read",
		"user:view",
	}
}

func cloneSortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func uniqueSortedStrings(values []string) []string {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = true
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func splitPermissionCode(value string) (string, string) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func toSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func randomAdminToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func verifyPassword(password string, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[3])
	if err != nil || len(expected) == 0 {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return hmac.Equal(actual, expected)
}

func hashPassword(password string) (string, error) {
	salt, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	iterations := 60000
	hash := pbkdf2SHA256([]byte(password), salt, iterations, sha256.Size)
	return "pbkdf2-sha256$" + strconv.Itoa(iterations) + "$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(hash), nil
}

func randomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func pbkdf2SHA256(password []byte, salt []byte, iterations int, keyLen int) []byte {
	hashLen := sha256.Size
	numBlocks := (keyLen + hashLen - 1) / hashLen
	output := make([]byte, 0, numBlocks*hashLen)
	for block := 1; block <= numBlocks; block++ {
		u := hmacSHA256(password, appendBlockIndex(salt, block))
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			u = hmacSHA256(password, u)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		output = append(output, t...)
	}
	return output[:keyLen]
}

func hmacSHA256(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

func appendBlockIndex(salt []byte, block int) []byte {
	result := append([]byte(nil), salt...)
	return append(result, byte(block>>24), byte(block>>16), byte(block>>8), byte(block))
}
