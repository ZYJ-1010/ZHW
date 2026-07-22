package invites

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

var ErrInviteAlreadyBound = errors.New("invite already bound")
var ErrInvalidEntryType = errors.New("invalid invite entry type")
var ErrInviteInactive = errors.New("invite inactive")

const (
	EntryTypePoster = "poster"
	EntryTypeQRCode = "qrcode"
	EntryTypeLink   = "link"

	AuthPageModeRegister = "register"
	AuthPageModeLogin    = "login"
)

type InviteCode struct {
	ID                     int64     `json:"id"`
	Code                   string    `json:"code"`
	OwnerID                int64     `json:"ownerUserId"`
	OwnerNickname          string    `json:"ownerNickname,omitempty"`
	OwnerPhoneMasked       string    `json:"ownerPhoneMasked,omitempty"`
	Status                 string    `json:"status"`
	DisplayStatus          string    `json:"displayStatus,omitempty"`
	UseStatus              string    `json:"useStatus,omitempty"`
	MaxUses                int       `json:"maxUses"`
	UsedCount              int       `json:"usedCount"`
	EntryType              string    `json:"entryType"`
	BoundWechatUserID      int64     `json:"boundWechatUserId,omitempty"`
	BoundWechatOpenID      string    `json:"-"`
	BoundWechatNickname    string    `json:"boundWechatNickname,omitempty"`
	BoundWechatPhoneMasked string    `json:"boundWechatPhoneMasked,omitempty"`
	ExpiresAt              time.Time `json:"expiresAt,omitempty"`
	CreatedAt              time.Time `json:"createdAt,omitempty"`
	UpdatedAt              time.Time `json:"updatedAt,omitempty"`
}

const (
	StatusActive    = "active"
	StatusDisabled  = "disabled"
	StatusVoided    = "voided"
	StatusExhausted = "exhausted"
	StatusExpired   = "expired"
)

type Relation struct {
	InviteCodeID  int64  `json:"inviteCodeId"`
	InviterUserID int64  `json:"inviterUserId"`
	InviteeUserID int64  `json:"inviteeUserId"`
	BindSource    string `json:"bindSource"`
}

type QuotaRequest struct {
	ID          int64     `json:"id"`
	OwnerUserID int64     `json:"ownerUserId"`
	Quantity    int       `json:"quantity"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	AuditReason string    `json:"auditReason,omitempty"`
	ReviewedBy  int64     `json:"reviewedBy,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ReviewedAt  time.Time `json:"reviewedAt,omitempty"`
}

type PrecheckResult struct {
	Valid         bool       `json:"valid"`
	InviteCode    InviteCode `json:"inviteCode"`
	EntryType     string     `json:"entryType"`
	AuthPageMode  string     `json:"authPageMode"`
	BoundWechat   bool       `json:"boundWechat"`
	BoundUserID   int64      `json:"boundUserId,omitempty"`
	FailureReason string     `json:"failureReason,omitempty"`
}

type Store struct {
	nextID         int64
	codes          map[string]InviteCode
	relations      map[int64]Relation
	entryRelations map[string]Relation
	boundCode      map[int64]int64
	quotaRequests  map[int64]QuotaRequest
	nextQuotaID    int64
	repo           Repository
}

func NewStore() *Store {
	return NewStoreWithRepository(nil)
}

func NewStoreWithRepository(repo Repository) *Store {
	store := &Store{
		nextID:         1,
		codes:          make(map[string]InviteCode),
		relations:      make(map[int64]Relation),
		entryRelations: make(map[string]Relation),
		boundCode:      make(map[int64]int64),
		quotaRequests:  make(map[int64]QuotaRequest),
		nextQuotaID:    1,
		repo:           repo,
	}
	store.UpsertCode("TEST2026", 0, 100)
	return store
}

type Repository interface {
	UpsertCode(ctx context.Context, invite InviteCode) (InviteCode, error)
	FindCode(ctx context.Context, code string) (InviteCode, bool, error)
	Bind(ctx context.Context, invite InviteCode, inviteeUserID int64, source string) (Relation, error)
	SetRelationInviter(ctx context.Context, inviteCodeID int64, inviteeUserID int64, inviterUserID int64, source string) (Relation, error)
	RelationForUser(ctx context.Context, userID int64) (Relation, bool, error)
	FindBoundCode(ctx context.Context, inviteCodeID int64) (InviteCode, bool, error)
	ListCodes(ctx context.Context, filter CodeFilter) ([]InviteCode, error)
	ListRelations(ctx context.Context, filter RelationFilter) ([]Relation, error)
	CreateQuotaRequest(ctx context.Context, request QuotaRequest) (QuotaRequest, error)
	ListQuotaRequests(ctx context.Context, ownerUserID int64, status string) ([]QuotaRequest, error)
	ReviewQuotaRequest(ctx context.Context, id int64, status string, auditReason string, reviewedBy int64) (QuotaRequest, error)
	ClearInviteeBindings(ctx context.Context, userID int64) error
}

func (s *Store) CreateQuotaRequest(ownerUserID int64, quantity int, reason string) (QuotaRequest, error) {
	if ownerUserID <= 0 || quantity <= 0 || quantity > 1000 {
		return QuotaRequest{}, errors.New("invalid invite quota request")
	}
	request := QuotaRequest{ID: s.nextQuotaID, OwnerUserID: ownerUserID, Quantity: quantity, Reason: strings.TrimSpace(reason), Status: "pending", CreatedAt: time.Now()}
	if s.repo != nil {
		created, err := s.repo.CreateQuotaRequest(context.Background(), request)
		if err != nil {
			return QuotaRequest{}, err
		}
		return created, nil
	}
	s.nextQuotaID++
	s.quotaRequests[request.ID] = request
	return request, nil
}

func (s *Store) ListQuotaRequests(ownerUserID int64, status string) ([]QuotaRequest, error) {
	if s.repo != nil {
		return s.repo.ListQuotaRequests(context.Background(), ownerUserID, status)
	}
	items := make([]QuotaRequest, 0)
	for _, item := range s.quotaRequests {
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

func (s *Store) ReviewQuotaRequest(id int64, status string, auditReason string, reviewedBy int64) (QuotaRequest, error) {
	if status != "approved" && status != "rejected" {
		return QuotaRequest{}, errors.New("invalid invite quota review status")
	}
	if s.repo != nil {
		return s.repo.ReviewQuotaRequest(context.Background(), id, status, auditReason, reviewedBy)
	}
	request, ok := s.quotaRequests[id]
	if !ok {
		return QuotaRequest{}, errors.New("invite quota request not found")
	}
	if request.Status != "pending" {
		return QuotaRequest{}, errors.New("invite quota request already reviewed")
	}
	request.Status, request.AuditReason, request.ReviewedBy, request.ReviewedAt = status, strings.TrimSpace(auditReason), reviewedBy, time.Now()
	s.quotaRequests[id] = request
	return request, nil
}

type CodeFilter struct {
	Status    string
	EntryType string
	OwnerID   int64
}

type RelationFilter struct {
	InviterUserID int64
	InviteeUserID int64
	InviteCodeID  int64
}

func (s *Store) UpsertCode(code string, ownerID int64, maxUses int) InviteCode {
	return s.UpsertCodeWithEntryType(code, ownerID, maxUses, EntryTypeLink)
}

func (s *Store) UpsertCodeWithEntryType(code string, ownerID int64, maxUses int, entryType string) InviteCode {
	entryType = NormalizeEntryType(entryType)
	if s.repo != nil {
		invite, err := s.repo.UpsertCode(context.Background(), InviteCode{Code: code, OwnerID: ownerID, Status: "active", MaxUses: maxUses, EntryType: entryType})
		if err == nil {
			s.codes[code] = invite
			return invite
		}
	}
	invite := InviteCode{
		ID:        s.nextID,
		Code:      code,
		OwnerID:   ownerID,
		Status:    "active",
		MaxUses:   maxUses,
		EntryType: entryType,
	}
	s.nextID++
	s.codes[code] = invite
	return invite
}

func (s *Store) FindCode(code string) (InviteCode, bool, error) {
	if s.repo != nil {
		invite, ok, err := s.repo.FindCode(context.Background(), code)
		if err == nil && ok {
			return invite, true, nil
		}
		if err != nil {
			return InviteCode{}, false, err
		}
	}
	invite, ok := s.codes[code]
	return invite, ok, nil
}

func (s *Store) ListCodes(filter CodeFilter) ([]InviteCode, error) {
	queryFilter := filter
	// 使用状态与过期状态由邀请码的绑定关系和有效期共同计算，不能直接
	// 交给数据库按原始 status 过滤，否则已过期的邀请码会被误显示为正常。
	queryFilter.Status = ""
	var items []InviteCode
	if s.repo != nil {
		storedItems, err := s.repo.ListCodes(context.Background(), queryFilter)
		if err != nil {
			return nil, err
		}
		items = storedItems
	} else {
		items = make([]InviteCode, 0, len(s.codes))
		for _, invite := range s.codes {
			items = append(items, invite)
		}
	}

	result := make([]InviteCode, 0, len(items))
	for _, invite := range items {
		if invite.MaxUses == 1 && invite.BoundWechatUserID == 0 {
			if bound, ok, _ := s.FindBoundCode(invite.ID); ok {
				invite.BoundWechatUserID = bound.BoundWechatUserID
				invite.BoundWechatNickname = bound.BoundWechatNickname
				invite.BoundWechatPhoneMasked = bound.BoundWechatPhoneMasked
			}
		}
		invite = decorateInviteStatus(invite, time.Now())
		if !matchesCodeFilter(invite, filter) {
			continue
		}
		result = append(result, invite)
	}
	return result, nil
}

func decorateInviteStatus(invite InviteCode, now time.Time) InviteCode {
	invite.UseStatus = "unused"
	if invite.UsedCount > 0 || invite.BoundWechatUserID > 0 {
		invite.UseStatus = "used"
	}
	invite.DisplayStatus = invite.Status
	if invite.Status == StatusActive && inviteExpiredAt(invite, now) {
		invite.DisplayStatus = StatusExpired
	}
	return invite
}

func matchesCodeFilter(invite InviteCode, filter CodeFilter) bool {
	if filter.EntryType != "" && invite.EntryType != filter.EntryType {
		return false
	}
	if filter.OwnerID > 0 && invite.OwnerID != filter.OwnerID {
		return false
	}
	status := strings.TrimSpace(filter.Status)
	if status == "" {
		return true
	}
	if status == "used" || status == "unused" {
		return invite.UseStatus == status
	}
	return invite.DisplayStatus == status
}

func inviteExpiredAt(invite InviteCode, now time.Time) bool {
	return !invite.ExpiresAt.IsZero() && !invite.ExpiresAt.After(now)
}

func (s *Store) Precheck(code string, entryType string) (PrecheckResult, error) {
	entryType, ok := ParseEntryType(entryType)
	if !ok {
		return PrecheckResult{Valid: false, AuthPageMode: AuthPageModeRegister, FailureReason: "invalid_entry_type"}, nil
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return PrecheckResult{Valid: false, EntryType: entryType, AuthPageMode: AuthPageModeRegister, FailureReason: "missing_invite_code"}, nil
	}
	invite, ok, err := s.FindCode(code)
	if err != nil {
		return PrecheckResult{}, err
	}
	if !ok {
		return PrecheckResult{Valid: false, EntryType: entryType, AuthPageMode: AuthPageModeRegister, FailureReason: "not_found"}, nil
	}
	if invite.EntryType == "" {
		invite.EntryType = entryType
	}
	if !invite.ExpiresAt.IsZero() && !time.Now().Before(invite.ExpiresAt) {
		invite.Status = StatusExpired
	}
	if entryType != "" && invite.EntryType != entryType {
		return PrecheckResult{
			Valid:         false,
			InviteCode:    invite,
			EntryType:     invite.EntryType,
			AuthPageMode:  AuthPageModeRegister,
			FailureReason: "entry_type_mismatch",
		}, nil
	}
	result := PrecheckResult{
		InviteCode:   invite,
		EntryType:    invite.EntryType,
		AuthPageMode: AuthPageModeRegister,
	}
	if invite.MaxUses == 1 {
		bound, ok, err := s.FindBoundCode(invite.ID)
		if err != nil {
			return PrecheckResult{}, err
		}
		if ok {
			result.InviteCode = bound
			result.BoundWechat = true
			result.BoundUserID = bound.BoundWechatUserID
			result.AuthPageMode = AuthPageModeLogin
			result.Valid = invite.Status == StatusActive
			if !result.Valid {
				result.FailureReason = "inactive"
			}
			return result, nil
		}
	}
	result.Valid = invite.Status == StatusActive && (invite.MaxUses == 0 || invite.UsedCount < invite.MaxUses)
	if !result.Valid {
		result.FailureReason = "inactive_or_exhausted"
	}
	return result, nil
}

func (s *Store) FindBoundCode(inviteCodeID int64) (InviteCode, bool, error) {
	if s.repo != nil {
		invite, ok, err := s.repo.FindBoundCode(context.Background(), inviteCodeID)
		if err == nil && ok {
			return invite, true, nil
		}
		if err != nil {
			return InviteCode{}, false, err
		}
	}
	userID, ok := s.boundCode[inviteCodeID]
	if !ok {
		return InviteCode{}, false, nil
	}
	for _, invite := range s.codes {
		if invite.ID == inviteCodeID {
			invite.BoundWechatUserID = userID
			return invite, true, nil
		}
	}
	return InviteCode{}, false, nil
}

func (s *Store) EnsureCodeForOwner(ownerID int64) (InviteCode, error) {
	for _, invite := range s.codes {
		if invite.OwnerID == ownerID && invite.Status == "active" {
			return invite, nil
		}
	}
	code := "ZH" + formatOwnerCode(ownerID)
	if s.repo != nil {
		invite, err := s.repo.UpsertCode(context.Background(), InviteCode{Code: code, OwnerID: ownerID, Status: "active", MaxUses: 0, EntryType: EntryTypeLink})
		if err != nil {
			return InviteCode{}, err
		}
		s.codes[code] = invite
		return invite, nil
	}
	return s.UpsertCode(code, ownerID, 0), nil
}

func (s *Store) DisableCode(code string) (InviteCode, error) {
	invite, ok, err := s.FindCode(code)
	if err != nil {
		return InviteCode{}, err
	}
	if !ok {
		return InviteCode{}, errors.New("invite code not found")
	}
	if inviteExpiredAt(invite, time.Now()) {
		return InviteCode{}, errors.New("expired invite code cannot be disabled")
	}
	if !inviteEditable(invite) {
		return InviteCode{}, errors.New("used invite code cannot be disabled")
	}
	invite.Status = StatusDisabled
	if s.repo != nil {
		updated, err := s.repo.UpsertCode(context.Background(), invite)
		if err != nil {
			return InviteCode{}, err
		}
		s.codes[updated.Code] = updated
		return updated, nil
	}
	s.codes[invite.Code] = invite
	return invite, nil
}

// UpdateUnusedCode only changes an invite before it has established a user
// relationship. Once used, its owner and distribution parameters are frozen.
func (s *Store) UpdateUnusedCode(code string, ownerID int64, entryType string, expiresAt time.Time) (InviteCode, error) {
	invite, ok, err := s.FindCode(code)
	if err != nil {
		return InviteCode{}, err
	}
	if !ok {
		return InviteCode{}, errors.New("invite code not found")
	}
	if inviteExpiredAt(invite, time.Now()) {
		return InviteCode{}, errors.New("expired invite code cannot be edited")
	}
	if !inviteEditable(invite) {
		return InviteCode{}, errors.New("used invite code cannot be edited")
	}
	entryType, valid := ParseEntryType(entryType)
	if !valid || entryType == "" {
		return InviteCode{}, ErrInvalidEntryType
	}
	if !expiresAt.IsZero() && !expiresAt.After(time.Now()) {
		return InviteCode{}, errors.New("invite expiry must be in the future")
	}
	invite.OwnerID = ownerID
	invite.EntryType = entryType
	invite.ExpiresAt = expiresAt
	if invite.Status == StatusExpired && (expiresAt.IsZero() || expiresAt.After(time.Now())) {
		invite.Status = StatusActive
	}
	return s.saveInvite(invite)
}

func (s *Store) EnableCode(code string) (InviteCode, error) {
	invite, ok, err := s.FindCode(code)
	if err != nil {
		return InviteCode{}, err
	}
	if !ok {
		return InviteCode{}, errors.New("invite code not found")
	}
	if inviteExpiredAt(invite, time.Now()) {
		return InviteCode{}, errors.New("expired invite code cannot be enabled")
	}
	if !inviteEditable(invite) {
		return InviteCode{}, errors.New("used invite code cannot be enabled")
	}
	if !invite.ExpiresAt.IsZero() && !invite.ExpiresAt.After(time.Now()) {
		return InviteCode{}, errors.New("expired invite code cannot be enabled")
	}
	if invite.Status != StatusDisabled {
		return InviteCode{}, errors.New("invite code cannot be enabled from current status")
	}
	invite.Status = StatusActive
	return s.saveInvite(invite)
}

func (s *Store) VoidCode(code string) (InviteCode, error) {
	invite, ok, err := s.FindCode(code)
	if err != nil {
		return InviteCode{}, err
	}
	if !ok {
		return InviteCode{}, errors.New("invite code not found")
	}
	if inviteExpiredAt(invite, time.Now()) {
		return InviteCode{}, errors.New("expired invite code cannot be voided")
	}
	if !inviteEditable(invite) {
		return InviteCode{}, errors.New("used invite code cannot be voided")
	}
	if invite.Status != StatusActive && invite.Status != StatusDisabled {
		return InviteCode{}, errors.New("invite code cannot be voided from current status")
	}
	invite.Status = StatusVoided
	return s.saveInvite(invite)
}

func inviteEditable(invite InviteCode) bool {
	return invite.UsedCount == 0 && invite.BoundWechatUserID == 0 && invite.Status != StatusVoided && invite.Status != StatusExhausted
}

func (s *Store) saveInvite(invite InviteCode) (InviteCode, error) {
	if s.repo != nil {
		updated, err := s.repo.UpsertCode(context.Background(), invite)
		if err != nil {
			return InviteCode{}, err
		}
		s.codes[updated.Code] = updated
		return updated, nil
	}
	invite.UpdatedAt = time.Now()
	s.codes[invite.Code] = invite
	return invite, nil
}

func (s *Store) IssueEntryCode(ownerID int64, entryType string) (InviteCode, error) {
	entryType, ok := ParseEntryType(entryType)
	if !ok || entryType == "" {
		return InviteCode{}, ErrInvalidEntryType
	}
	for i := 0; i < 5; i++ {
		code, err := randomInviteCode(entryType)
		if err != nil {
			return InviteCode{}, err
		}
		if _, exists, err := s.FindCode(code); err != nil {
			return InviteCode{}, err
		} else if exists {
			continue
		}
		return s.UpsertCodeWithEntryType(code, ownerID, 1, entryType), nil
	}
	return InviteCode{}, errors.New("failed to generate unique invite code")
}

func (s *Store) Bind(invite InviteCode, inviteeUserID int64, source string) (Relation, error) {
	if s.repo != nil {
		relation, err := s.repo.Bind(context.Background(), invite, inviteeUserID, source)
		if err != nil {
			return Relation{}, err
		}
		s.relations[inviteeUserID] = relation
		return relation, nil
	}
	if current, found := s.codes[invite.Code]; found {
		invite = current
	}
	if invite.Status != StatusActive || (!invite.ExpiresAt.IsZero() && !invite.ExpiresAt.After(time.Now())) || (invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses) {
		return Relation{}, ErrInviteInactive
	}
	key := relationKey(invite.ID, inviteeUserID)
	if relation, ok := s.entryRelations[key]; ok {
		return relation, nil
	}
	if invite.MaxUses == 1 {
		if boundUserID, ok := s.boundCode[invite.ID]; ok {
			if boundUserID != inviteeUserID {
				return Relation{}, ErrInviteAlreadyBound
			}
			if relation, ok := s.entryRelations[key]; ok {
				return relation, nil
			}
		}
	}
	if relation, ok := s.relations[inviteeUserID]; ok {
		return relation, nil
	}

	relation := Relation{
		InviteCodeID:  invite.ID,
		InviterUserID: invite.OwnerID,
		InviteeUserID: inviteeUserID,
		BindSource:    source,
	}
	if invite.MaxUses == 1 {
		s.boundCode[invite.ID] = inviteeUserID
	}
	invite.UsedCount++
	if invite.BoundWechatUserID == 0 {
		invite.BoundWechatUserID = inviteeUserID
	}
	s.entryRelations[key] = relation
	s.relations[inviteeUserID] = relation
	s.codes[invite.Code] = invite
	return relation, nil
}

// ClearInviteeBindings clears the deleted account's invite relationship but
// intentionally does not decrease used_count. A consumed invitation cannot be
// reused after the account is deleted.
func (s *Store) ClearInviteeBindings(userID int64) error {
	if userID <= 0 {
		return nil
	}
	if s.repo != nil {
		if err := s.repo.ClearInviteeBindings(context.Background(), userID); err != nil {
			return err
		}
	}
	if relation, ok := s.relations[userID]; ok {
		delete(s.boundCode, relation.InviteCodeID)
		delete(s.entryRelations, relationKey(relation.InviteCodeID, userID))
	}
	delete(s.relations, userID)
	for key, relation := range s.entryRelations {
		if relation.InviteeUserID == userID {
			delete(s.entryRelations, key)
			if s.boundCode[relation.InviteCodeID] == userID {
				delete(s.boundCode, relation.InviteCodeID)
			}
		}
	}
	return nil
}

func (s *Store) SetRelationInviter(inviteeUserID int64, inviterUserID int64, source string) (Relation, error) {
	if inviterUserID <= 0 {
		return Relation{}, errors.New("invalid inviter user id")
	}
	code, err := s.EnsureCodeForOwner(inviterUserID)
	if err != nil {
		return Relation{}, err
	}
	if s.repo != nil {
		relation, err := s.repo.SetRelationInviter(context.Background(), code.ID, inviteeUserID, inviterUserID, source)
		if err != nil {
			return Relation{}, err
		}
		relation.InviteCodeID = code.ID
		relation.InviterUserID = inviterUserID
		relation.InviteeUserID = inviteeUserID
		relation.BindSource = source
		s.relations[inviteeUserID] = relation
		s.entryRelations[relationKey(code.ID, inviteeUserID)] = relation
		return relation, nil
	}
	relation, ok := s.relations[inviteeUserID]
	if !ok {
		relation = Relation{
			InviteCodeID:  code.ID,
			InviterUserID: inviterUserID,
			InviteeUserID: inviteeUserID,
			BindSource:    source,
		}
		s.entryRelations[relationKey(code.ID, inviteeUserID)] = relation
		s.relations[inviteeUserID] = relation
		return relation, nil
	}
	relation.InviterUserID = inviterUserID
	relation.BindSource = source
	relation.InviteCodeID = code.ID
	s.entryRelations[relationKey(code.ID, inviteeUserID)] = relation
	s.relations[inviteeUserID] = relation
	return relation, nil
}

func relationKey(inviteCodeID int64, userID int64) string {
	return strconv.FormatInt(inviteCodeID, 10) + ":" + strconv.FormatInt(userID, 10)
}

func (s *Store) RelationForUser(userID int64) (Relation, bool, error) {
	if s.repo != nil {
		relation, ok, err := s.repo.RelationForUser(context.Background(), userID)
		if err == nil && ok {
			return relation, true, nil
		}
		if err != nil {
			return Relation{}, false, err
		}
	}
	relation, ok := s.relations[userID]
	return relation, ok, nil
}

func (s *Store) ListRelations(filter RelationFilter) ([]Relation, error) {
	if s.repo != nil {
		return s.repo.ListRelations(context.Background(), filter)
	}
	result := make([]Relation, 0, len(s.entryRelations))
	for _, relation := range s.entryRelations {
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

func NormalizeEntryType(entryType string) string {
	if normalized, ok := ParseEntryType(entryType); ok && normalized != "" {
		return normalized
	}
	return EntryTypeLink
}

func ParseEntryType(entryType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(entryType)) {
	case "":
		return "", true
	case EntryTypePoster, "card", "share_card", "小程序卡片", "海报", "卡片":
		return EntryTypePoster, true
	case EntryTypeQRCode, "qr_code", "qr", "二维码", "小程序码", "扫码":
		return EntryTypeQRCode, true
	case EntryTypeLink, "url", "urllink", "链接", "邀请链接":
		return EntryTypeLink, true
	}
	return "", false
}

func formatOwnerCode(ownerID int64) string {
	if ownerID < 0 {
		ownerID = -ownerID
	}
	digits := "000000" + strconv.FormatInt(ownerID, 10)
	return digits[len(digits)-6:]
}

func randomInviteCode(entryType string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	prefix := "L"
	switch entryType {
	case EntryTypePoster:
		prefix = "P"
	case EntryTypeQRCode:
		prefix = "Q"
	}
	return prefix + strings.ToUpper(hex.EncodeToString(buf)), nil
}
