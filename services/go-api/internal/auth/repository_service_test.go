package auth

import (
	"context"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func TestWechatLoginWithUserAndInviteRepositories(t *testing.T) {
	userRepo := newFakeUserRepository()
	inviteRepo := newFakeInviteRepository()
	_, _ = inviteRepo.UpsertCode(context.Background(), invites.InviteCode{Code: "TEST2026", Status: "active", MaxUses: 100})
	service := NewService(users.NewStoreWithRepository(userRepo), invites.NewStoreWithRepository(inviteRepo), NewTokenStore())

	first, err := service.WechatLogin(WechatLoginRequest{Code: "repo-user", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	if first.User.ID == 0 || first.InviteRelation == nil {
		t.Fatalf("expected persisted user and relation: %+v", first)
	}
	if inviteRepo.codes["TEST2026"].UsedCount != 1 {
		t.Fatalf("expected one invite use, got %d", inviteRepo.codes["TEST2026"].UsedCount)
	}

	second, err := service.WechatLogin(WechatLoginRequest{Code: "repo-user"})
	if err != nil {
		t.Fatalf("bound repository user should login without invite: %v", err)
	}
	if second.User.ID != first.User.ID || second.InviteRelation == nil || second.AuthPageMode != invites.AuthPageModeLogin || !second.BoundWechat {
		t.Fatalf("expected bound repository user login mode, got %+v", second)
	}
	second, err = service.WechatLogin(WechatLoginRequest{Code: "repo-user", InviteCode: "TEST2026"})
	if err != nil {
		t.Fatalf("second login with invite: %v", err)
	}
	if second.User.ID != first.User.ID {
		t.Fatalf("expected same user id, got %d and %d", first.User.ID, second.User.ID)
	}
	if inviteRepo.codes["TEST2026"].UsedCount != 1 {
		t.Fatalf("expected no duplicate invite use, got %d", inviteRepo.codes["TEST2026"].UsedCount)
	}
	if current, ok := service.CurrentUser(second.PreAuthToken); !ok || current.ID != first.User.ID {
		t.Fatalf("expected current user from repository, got %+v ok=%v", current, ok)
	}
}

func TestInvitePrecheckWithRepository(t *testing.T) {
	userRepo := newFakeUserRepository()
	inviteRepo := newFakeInviteRepository()
	code, _ := inviteRepo.UpsertCode(context.Background(), invites.InviteCode{Code: "QR2026", Status: "active", MaxUses: 1, EntryType: invites.EntryTypeQRCode})
	service := NewService(users.NewStoreWithRepository(userRepo), invites.NewStoreWithRepository(inviteRepo), NewTokenStore())

	before, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: code.Code, EntryType: "qrcode"})
	if err != nil {
		t.Fatalf("precheck before bind failed: %v", err)
	}
	if before.AuthPageMode != invites.AuthPageModeRegister || before.EntryType != invites.EntryTypeQRCode {
		t.Fatalf("expected register precheck from repo, got %+v", before)
	}

	_, err = service.WechatLogin(WechatLoginRequest{Code: "repo-user", InviteCode: code.Code, EntryType: "qrcode"})
	if err != nil {
		t.Fatalf("repo login failed: %v", err)
	}
	after, err := service.InvitePrecheck(InvitePrecheckRequest{InviteCode: code.Code, EntryType: "qrcode"})
	if err != nil {
		t.Fatalf("precheck after bind failed: %v", err)
	}
	if !after.BoundWechat || after.AuthPageMode != invites.AuthPageModeLogin {
		t.Fatalf("expected login precheck after bind, got %+v", after)
	}
}

type fakeUserRepository struct {
	nextID  int64
	byID    map[int64]users.User
	byOpen  map[string]int64
	byPhone map[string]int64
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{nextID: 1, byID: make(map[int64]users.User), byOpen: make(map[string]int64), byPhone: make(map[string]int64)}
}

func (r *fakeUserRepository) FindByOpenID(ctx context.Context, openID string) (users.User, bool, error) {
	id, ok := r.byOpen[openID]
	if !ok {
		return users.User{}, false, nil
	}
	return r.byID[id], true, nil
}

func (r *fakeUserRepository) FindByPhoneHash(ctx context.Context, phoneHash string) (users.User, bool, error) {
	id, ok := r.byPhone[phoneHash]
	if !ok {
		return users.User{}, false, nil
	}
	return r.byID[id], true, nil
}

func (r *fakeUserRepository) FindByID(ctx context.Context, id int64) (users.User, bool, error) {
	user, ok := r.byID[id]
	return user, ok, nil
}

func (r *fakeUserRepository) List(ctx context.Context, filter users.Filter) ([]users.User, error) {
	items := make([]users.User, 0, len(r.byID))
	for _, user := range r.byID {
		items = append(items, user)
	}
	return items, nil
}

func (r *fakeUserRepository) CreateWithOpenID(ctx context.Context, openID string) (users.User, error) {
	user := users.User{
		ID:             r.nextID,
		OpenID:         openID,
		RealnameStatus: "pending",
		Status:         "active",
		CreatedAt:      time.Now(),
	}
	r.nextID++
	r.byID[user.ID] = user
	r.byOpen[openID] = user.ID
	return user, nil
}

func (r *fakeUserRepository) CreateWithPhone(ctx context.Context, phoneHash string, phoneMasked string) (users.User, error) {
	user := users.User{
		ID:             r.nextID,
		PhoneMasked:    phoneMasked,
		RealnameStatus: "pending",
		Status:         "active",
		CreatedAt:      time.Now(),
	}
	r.nextID++
	r.byID[user.ID] = user
	r.byPhone[phoneHash] = user.ID
	return user, nil
}

func (r *fakeUserRepository) UpdatePhoneAuth(ctx context.Context, userID int64, phoneHash string, phoneMasked string) (users.User, error) {
	user := r.byID[userID]
	user.PhoneMasked = phoneMasked
	r.byID[user.ID] = user
	r.byPhone[phoneHash] = user.ID
	return user, nil
}

func (r *fakeUserRepository) UpdateProfile(ctx context.Context, userID int64, nickname string, avatarURL string, avatarFileID int64) (users.User, error) {
	user := r.byID[userID]
	user.Nickname = nickname
	user.AvatarURL = avatarURL
	user.AvatarFileID = avatarFileID
	r.byID[user.ID] = user
	return user, nil
}

func (r *fakeUserRepository) UpdateRealnameStatus(ctx context.Context, userID int64, status string) (users.User, error) {
	user := r.byID[userID]
	user.RealnameStatus = status
	r.byID[user.ID] = user
	return user, nil
}

func (r *fakeUserRepository) PasswordHash(ctx context.Context, userID int64) (string, bool, error) {
	user, ok := r.byID[userID]
	return user.PasswordHash, ok && user.PasswordHash != "", nil
}

func (r *fakeUserRepository) UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error {
	user, ok := r.byID[userID]
	if !ok {
		return users.ErrInvalidProfile
	}
	user.PasswordHash = passwordHash
	r.byID[userID] = user
	return nil
}

func (r *fakeUserRepository) BindWechat(ctx context.Context, userID int64, openID string) (users.User, error) {
	user, ok := r.byID[userID]
	if !ok {
		return users.User{}, users.ErrInvalidProfile
	}
	user.OpenID = openID
	r.byID[userID] = user
	return user, nil
}

func (r *fakeUserRepository) DeactivateAndClearLoginBindings(ctx context.Context, userID int64) (users.User, error) {
	user, ok := r.byID[userID]
	if !ok {
		return users.User{}, users.ErrInvalidProfile
	}
	for openID, id := range r.byOpen {
		if id == userID {
			delete(r.byOpen, openID)
		}
	}
	for phoneHash, id := range r.byPhone {
		if id == userID {
			delete(r.byPhone, phoneHash)
		}
	}
	user.OpenID = ""
	user.PhoneMasked = ""
	user.Nickname = ""
	user.AvatarURL = ""
	user.AvatarFileID = 0
	user.Status = "deleted"
	r.byID[userID] = user
	return user, nil
}

type fakeInviteRepository struct {
	nextID        int64
	codes         map[string]invites.InviteCode
	relations     map[int64]invites.Relation
	boundCode     map[int64]int64
	quotaRequests map[int64]invites.QuotaRequest
	nextQuotaID   int64
}

func newFakeInviteRepository() *fakeInviteRepository {
	return &fakeInviteRepository{nextID: 1, codes: make(map[string]invites.InviteCode), relations: make(map[int64]invites.Relation), boundCode: make(map[int64]int64), quotaRequests: make(map[int64]invites.QuotaRequest), nextQuotaID: 1}
}

func (r *fakeInviteRepository) UpsertCode(ctx context.Context, invite invites.InviteCode) (invites.InviteCode, error) {
	if existing, ok := r.codes[invite.Code]; ok {
		existing.OwnerID = invite.OwnerID
		existing.Status = invite.Status
		existing.MaxUses = invite.MaxUses
		existing.EntryType = invite.EntryType
		existing.ExpiresAt = invite.ExpiresAt
		r.codes[invite.Code] = existing
		return existing, nil
	}
	if invite.ID == 0 {
		invite.ID = r.nextID
		r.nextID++
	} else if invite.ID >= r.nextID {
		r.nextID = invite.ID + 1
	}
	r.codes[invite.Code] = invite
	return invite, nil
}

func (r *fakeInviteRepository) CreateQuotaRequest(ctx context.Context, request invites.QuotaRequest) (invites.QuotaRequest, error) {
	request.ID = r.nextQuotaID
	r.nextQuotaID++
	r.quotaRequests[request.ID] = request
	return request, nil
}

func (r *fakeInviteRepository) ListQuotaRequests(ctx context.Context, ownerUserID int64, status string) ([]invites.QuotaRequest, error) {
	items := make([]invites.QuotaRequest, 0)
	for _, item := range r.quotaRequests {
		if ownerUserID > 0 && item.OwnerUserID != ownerUserID {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeInviteRepository) ReviewQuotaRequest(ctx context.Context, id int64, status string, auditReason string, reviewedBy int64) (invites.QuotaRequest, error) {
	item, ok := r.quotaRequests[id]
	if !ok {
		return invites.QuotaRequest{}, context.Canceled
	}
	item.Status, item.AuditReason, item.ReviewedBy = status, auditReason, reviewedBy
	r.quotaRequests[id] = item
	return item, nil
}

func (r *fakeInviteRepository) ClearInviteeBindings(ctx context.Context, userID int64) error {
	if relation, ok := r.relations[userID]; ok {
		delete(r.boundCode, relation.InviteCodeID)
	}
	delete(r.relations, userID)
	return nil
}

func (r *fakeInviteRepository) FindCode(ctx context.Context, code string) (invites.InviteCode, bool, error) {
	invite, ok := r.codes[code]
	return invite, ok, nil
}

func (r *fakeInviteRepository) ListCodes(ctx context.Context, filter invites.CodeFilter) ([]invites.InviteCode, error) {
	result := make([]invites.InviteCode, 0)
	for _, invite := range r.codes {
		if filter.Status != "" && invite.Status != filter.Status {
			continue
		}
		if filter.EntryType != "" && invite.EntryType != filter.EntryType {
			continue
		}
		if filter.OwnerID > 0 && invite.OwnerID != filter.OwnerID {
			continue
		}
		result = append(result, invite)
	}
	return result, nil
}

func (r *fakeInviteRepository) Bind(ctx context.Context, invite invites.InviteCode, inviteeUserID int64, source string) (invites.Relation, error) {
	if invite.MaxUses == 1 {
		if boundUserID, ok := r.boundCode[invite.ID]; ok {
			if boundUserID != inviteeUserID {
				return invites.Relation{}, invites.ErrInviteAlreadyBound
			}
			if relation, ok := r.relations[inviteeUserID]; ok && relation.InviteCodeID == invite.ID {
				return relation, nil
			}
		}
	} else {
		if relation, ok := r.relations[inviteeUserID]; ok {
			return relation, nil
		}
	}
	relation := invites.Relation{InviteCodeID: invite.ID, InviterUserID: invite.OwnerID, InviteeUserID: inviteeUserID, BindSource: source}
	if invite.MaxUses == 1 {
		r.boundCode[invite.ID] = inviteeUserID
	}
	invite.UsedCount++
	if _, ok := r.relations[inviteeUserID]; !ok {
		r.relations[inviteeUserID] = relation
	}
	r.codes[invite.Code] = invite
	return relation, nil
}

func (r *fakeInviteRepository) SetRelationInviter(ctx context.Context, inviteCodeID int64, inviteeUserID int64, inviterUserID int64, source string) (invites.Relation, error) {
	relation := invites.Relation{
		InviteCodeID:  inviteCodeID,
		InviterUserID: inviterUserID,
		InviteeUserID: inviteeUserID,
		BindSource:    source,
	}
	r.relations[inviteeUserID] = relation
	if _, ok := r.boundCode[inviteCodeID]; !ok {
		r.boundCode[inviteCodeID] = inviteeUserID
	}
	return relation, nil
}

func (r *fakeInviteRepository) RelationForUser(ctx context.Context, userID int64) (invites.Relation, bool, error) {
	relation, ok := r.relations[userID]
	return relation, ok, nil
}

func (r *fakeInviteRepository) ListRelations(ctx context.Context, filter invites.RelationFilter) ([]invites.Relation, error) {
	result := make([]invites.Relation, 0)
	for _, relation := range r.relations {
		if filter.InviterUserID > 0 && relation.InviterUserID != filter.InviterUserID {
			continue
		}
		if filter.InviteeUserID > 0 && relation.InviteeUserID != filter.InviteeUserID {
			continue
		}
		if filter.InviteCodeID > 0 && relation.InviteCodeID != filter.InviteCodeID {
			continue
		}
		result = append(result, relation)
	}
	return result, nil
}

func (r *fakeInviteRepository) FindBoundCode(ctx context.Context, inviteCodeID int64) (invites.InviteCode, bool, error) {
	for _, invite := range r.codes {
		if invite.ID != inviteCodeID {
			continue
		}
		if userID, ok := r.boundCode[invite.ID]; ok {
			invite.BoundWechatUserID = userID
			return invite, true, nil
		}
		return invites.InviteCode{}, false, nil
	}
	return invites.InviteCode{}, false, nil
}
