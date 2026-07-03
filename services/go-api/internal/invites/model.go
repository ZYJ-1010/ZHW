package invites

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

var ErrInviteAlreadyBound = errors.New("invite already bound")
var ErrInvalidEntryType = errors.New("invalid invite entry type")

const (
	EntryTypePoster = "poster"
	EntryTypeQRCode = "qrcode"
	EntryTypeLink   = "link"

	AuthPageModeRegister = "register"
	AuthPageModeLogin    = "login"
)

type InviteCode struct {
	ID                  int64  `json:"id"`
	Code                string `json:"code"`
	OwnerID             int64  `json:"ownerUserId"`
	Status              string `json:"status"`
	MaxUses             int    `json:"maxUses"`
	UsedCount           int    `json:"usedCount"`
	EntryType           string `json:"entryType"`
	BoundWechatUserID   int64  `json:"boundWechatUserId,omitempty"`
	BoundWechatOpenID   string `json:"-"`
	BoundWechatNickname string `json:"boundWechatNickname,omitempty"`
}

type Relation struct {
	InviteCodeID  int64  `json:"inviteCodeId"`
	InviterUserID int64  `json:"inviterUserId"`
	InviteeUserID int64  `json:"inviteeUserId"`
	BindSource    string `json:"bindSource"`
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
		repo:           repo,
	}
	store.UpsertCode("TEST2026", 0, 100)
	return store
}

type Repository interface {
	UpsertCode(ctx context.Context, invite InviteCode) (InviteCode, error)
	FindCode(ctx context.Context, code string) (InviteCode, bool, error)
	Bind(ctx context.Context, invite InviteCode, inviteeUserID int64, source string) (Relation, error)
	RelationForUser(ctx context.Context, userID int64) (Relation, bool, error)
	FindBoundCode(ctx context.Context, inviteCodeID int64) (InviteCode, bool, error)
	ListCodes(ctx context.Context, filter CodeFilter) ([]InviteCode, error)
	ListRelations(ctx context.Context, filter RelationFilter) ([]Relation, error)
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
	if s.repo != nil {
		items, err := s.repo.ListCodes(context.Background(), filter)
		if err == nil {
			return items, nil
		}
		return nil, err
	}
	result := make([]InviteCode, 0, len(s.codes))
	for _, invite := range s.codes {
		if filter.Status != "" && invite.Status != filter.Status {
			continue
		}
		if filter.EntryType != "" && invite.EntryType != filter.EntryType {
			continue
		}
		if filter.OwnerID > 0 && invite.OwnerID != filter.OwnerID {
			continue
		}
		if invite.MaxUses == 1 {
			if bound, ok, _ := s.FindBoundCode(invite.ID); ok {
				invite.BoundWechatUserID = bound.BoundWechatUserID
				invite.BoundWechatNickname = bound.BoundWechatNickname
			}
		}
		result = append(result, invite)
	}
	return result, nil
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
			result.Valid = invite.Status == "active"
			if !result.Valid {
				result.FailureReason = "inactive"
			}
			return result, nil
		}
	}
	result.Valid = invite.Status == "active" && (invite.MaxUses == 0 || invite.UsedCount < invite.MaxUses)
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
	invite.Status = "disabled"
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
	key := relationKey(invite.ID, inviteeUserID)
	if relation, ok := s.relations[inviteeUserID]; ok {
		return relation, nil
	}
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
	if _, ok := s.relations[inviteeUserID]; !ok {
		s.relations[inviteeUserID] = relation
	}
	s.codes[invite.Code] = invite
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
	case EntryTypePoster:
		return EntryTypePoster, true
	case EntryTypeQRCode, "qr_code", "qr":
		return EntryTypeQRCode, true
	case EntryTypeLink:
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
