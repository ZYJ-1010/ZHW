package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

var (
	ErrInviteRequired      = errors.New("invite required")
	ErrInvalidInvite       = errors.New("invalid invite")
	ErrInviteAlreadyBound  = errors.New("invite already bound")
	ErrInviteExpired       = errors.New("invite expired")
	ErrWechatCodeInvalid   = errors.New("wechat code invalid")
	ErrPhoneRequired       = errors.New("phone required")
	ErrPhoneInvalid        = errors.New("phone invalid")
	ErrPhoneCodeInvalid    = errors.New("phone code invalid")
	ErrInviteQuotaExceeded = errors.New("invite quota exhausted")
	ErrPhoneCodeRateLimit  = errors.New("phone code send rate limited")
	ErrPhoneCodeDailyLimit = errors.New("phone code daily limit exceeded")
	ErrPhoneCodeSendFailed = errors.New("phone code send failed")
	ErrPasswordInvalid     = errors.New("password invalid")
	ErrPasswordNotSet      = errors.New("password not set")
	ErrWechatAlreadyBound  = errors.New("wechat already bound")
)

const temporaryPhoneCode = "000000"

const (
	phoneCodeResendInterval = 30 * time.Second
	phoneCodeExpiry         = 5 * time.Minute
	phoneCodeDailyLimit     = 5
	phoneCodeMaxFailures    = 5
)

const (
	InviteBindingStatusAlreadyBound = "already_bound"
	inviteAlreadyBoundMessage       = "该微信已绑定邀请码，将继续使用原邀请码进入小程序"
)

type InvitePrecheckRequest struct {
	InviteCode string `json:"inviteCode"`
	EntryType  string `json:"entryType"`
}

type PasswordLoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type WechatLoginRequest struct {
	Code       string `json:"code"`
	InviteCode string `json:"inviteCode"`
	EntryType  string `json:"entryType"`
}

// WechatEntryPrecheckResponse is intentionally limited to routing data. It
// must not create users, consume invite codes, or issue a login token.
type WechatEntryPrecheckResponse struct {
	BoundWechat    bool `json:"boundWechat"`
	RequiresInvite bool `json:"requiresInvite"`
	RequiresReauth bool `json:"requiresReauth"`
}

type PhoneLoginRequest struct {
	Phone      string `json:"phone"`
	Code       string `json:"code"`
	InviteCode string `json:"inviteCode"`
	EntryType  string `json:"entryType"`
}

type LoginResponse struct {
	Token                   string            `json:"token,omitempty"`
	PreAuthToken            string            `json:"preAuthToken,omitempty"`
	ExpiresAt               string            `json:"expiresAt"`
	User                    users.User        `json:"user"`
	InviteRelation          *invites.Relation `json:"inviteRelation,omitempty"`
	NeedProfile             bool              `json:"needProfile"`
	RequiresIdentityBinding bool              `json:"requiresIdentityBinding"`
	IdentityBindStatus      string            `json:"identityBindStatus,omitempty"`
	EntryType               string            `json:"entryType"`
	AuthPageMode            string            `json:"authPageMode"`
	BoundWechat             bool              `json:"boundWechat"`
	InviteBindingStatus     string            `json:"inviteBindingStatus,omitempty"`
	InviteBindingMessage    string            `json:"inviteBindingMessage,omitempty"`
	phoneCodeClaim          phoneCodeState
}

type phoneCodeState struct {
	code           string
	scene          string
	sentAt         time.Time
	dayKey         string
	dayCount       int
	sending        bool
	failedAttempts int
	phoneKey       string
	claimToken     string
	repository     bool
}

type Service struct {
	users              *users.Store
	invites            *invites.Store
	tokens             *TokenStore
	resolver           WechatCodeResolver
	phoneLookupSecret  string
	phoneSMSSender     identity.SMSSender
	phoneCodeRepo      PhoneCodeRepository
	phoneCodeMu        sync.Mutex
	phoneCodes         map[string]phoneCodeState
	relationListenerMu sync.RWMutex
	relationListener   func(invites.Relation)
	phoneVerifierMu    sync.RWMutex
	phoneVerifier      func(userID int64) bool
}

// SetInviteRelationListener lets the application layer observe newly bound
// invitation relations without coupling authentication to growth services.
func (s *Service) SetInviteRelationListener(listener func(invites.Relation)) {
	s.relationListenerMu.Lock()
	s.relationListener = listener
	s.relationListenerMu.Unlock()
}

// SetPhoneBindingVerifier lets the application layer confirm that a stored
// phone number has completed SMS verification. A masked phone number alone is
// not proof of ownership because it may have been entered during onboarding.
func (s *Service) SetPhoneBindingVerifier(verifier func(userID int64) bool) {
	s.phoneVerifierMu.Lock()
	s.phoneVerifier = verifier
	s.phoneVerifierMu.Unlock()
}

func (s *Service) phoneBindingVerified(userID int64, phoneMasked string) bool {
	if strings.TrimSpace(phoneMasked) == "" {
		return false
	}
	s.phoneVerifierMu.RLock()
	verifier := s.phoneVerifier
	s.phoneVerifierMu.RUnlock()
	if verifier == nil {
		// Standalone auth-service users do not have an identity service. The
		// application server always installs the verifier during construction.
		return true
	}
	return verifier(userID)
}

func (s *Service) SetAppReauthAfterDays(days int) {
	if s == nil || s.tokens == nil || days <= 0 {
		return
	}
	s.tokens.SetAppReauthAfter(time.Duration(days) * 24 * time.Hour)
}

func (s *Service) notifyInviteRelationBound(relation invites.Relation) {
	s.relationListenerMu.RLock()
	listener := s.relationListener
	s.relationListenerMu.RUnlock()
	if listener != nil {
		listener(relation)
	}
}

func NewService(userStore *users.Store, inviteStore *invites.Store, tokenStore *TokenStore) *Service {
	service := &Service{
		users:             userStore,
		invites:           inviteStore,
		tokens:            tokenStore,
		resolver:          MockWechatCodeResolver{},
		phoneLookupSecret: "local-phone-lookup-secret",
		phoneSMSSender:    identity.LocalSMSSender{},
		phoneCodes:        make(map[string]phoneCodeState),
	}
	if tokenStore != nil && tokenStore.repo != nil {
		service.phoneCodeRepo, _ = tokenStore.repo.(PhoneCodeRepository)
	}
	return service
}

func (s *Service) UseWechatCodeResolver(resolver WechatCodeResolver) {
	if resolver == nil {
		return
	}
	s.resolver = resolver
}

func (s *Service) UsePhoneLookupSecret(secret string) {
	secret = strings.TrimSpace(secret)
	if secret != "" {
		s.phoneLookupSecret = secret
	}
}

func (s *Service) UsePhoneSMSSender(sender identity.SMSSender) {
	if sender != nil {
		s.phoneSMSSender = sender
	}
}

func (s *Service) SendPhoneCode(ctx context.Context, phone string, scene string) (identity.SMSDispatchResult, error) {
	phone = strings.TrimSpace(phone)
	if !validMainlandPhone(phone) {
		return identity.SMSDispatchResult{}, ErrPhoneInvalid
	}
	phoneKey := s.phoneLookupHash(phone)
	scene = normalizePhoneCodeScene(scene)
	nowTime := time.Now()
	if s.phoneCodeRepo != nil {
		return s.sendPhoneCodeWithRepository(ctx, phone, phoneKey, scene, nowTime)
	}
	s.phoneCodeMu.Lock()
	state := s.phoneCodes[phoneKey]
	if state.sending {
		s.phoneCodeMu.Unlock()
		return identity.SMSDispatchResult{}, ErrPhoneCodeRateLimit
	}
	if !state.sentAt.IsZero() && nowTime.Sub(state.sentAt) < phoneCodeResendInterval {
		s.phoneCodeMu.Unlock()
		return identity.SMSDispatchResult{}, ErrPhoneCodeRateLimit
	}
	dayKey := nowTime.Format("20060102")
	dayCount := state.dayCount
	if state.dayKey != dayKey {
		dayCount = 0
	}
	if dayCount >= phoneCodeDailyLimit {
		s.phoneCodeMu.Unlock()
		return identity.SMSDispatchResult{}, ErrPhoneCodeDailyLimit
	}
	sender := s.phoneSMSSender
	code := sender.GenerateCode()
	previousState := state
	inFlightState := state
	inFlightState.code = code
	inFlightState.scene = scene
	inFlightState.sending = true
	s.phoneCodes[phoneKey] = inFlightState
	s.phoneCodeMu.Unlock()

	result, err := sender.Send(ctx, identity.SMSDispatchRequest{
		Phone:       phone,
		PhoneMasked: maskPhone(phone),
		Scene:       scene,
		Code:        code,
		ExpiresAt:   nowTime.Add(phoneCodeExpiry),
	})
	if err != nil {
		s.phoneCodeMu.Lock()
		current, found := s.phoneCodes[phoneKey]
		if found && current.sending && current.code == code {
			if previousState.code == "" && previousState.sentAt.IsZero() && previousState.dayCount == 0 {
				delete(s.phoneCodes, phoneKey)
			} else {
				s.phoneCodes[phoneKey] = previousState
			}
		}
		s.phoneCodeMu.Unlock()
		return identity.SMSDispatchResult{}, ErrPhoneCodeSendFailed
	}

	s.phoneCodeMu.Lock()
	confirmedState := previousState
	if confirmedState.dayKey != dayKey {
		confirmedState.dayKey = dayKey
		confirmedState.dayCount = 0
	}
	confirmedState.code = code
	confirmedState.scene = scene
	confirmedState.sentAt = nowTime
	confirmedState.dayCount++
	confirmedState.sending = false
	confirmedState.failedAttempts = 0
	s.phoneCodes[phoneKey] = confirmedState
	s.phoneCodeMu.Unlock()
	return result, nil
}

func (s *Service) sendPhoneCodeWithRepository(ctx context.Context, phone string, phoneKey string, scene string, nowTime time.Time) (identity.SMSDispatchResult, error) {
	sender := s.phoneSMSSender
	code := sender.GenerateCode()
	reservationToken, err := randomToken(16)
	if err != nil {
		return identity.SMSDispatchResult{}, err
	}
	dayKey := nowTime.Format("20060102")
	err = s.phoneCodeRepo.ReservePhoneCodeSend(ctx, PhoneCodeSendReservation{
		PhoneKey:         phoneKey,
		PhoneMasked:      maskPhone(phone),
		Scene:            scene,
		CodeHash:         s.phoneCodeHash(phoneKey, scene, code),
		ReservationToken: reservationToken,
		StartedAt:        nowTime,
		ExpiresAt:        nowTime.Add(phoneCodeExpiry),
		DayKey:           dayKey,
		ResendInterval:   phoneCodeResendInterval,
		DailyLimit:       phoneCodeDailyLimit,
	})
	if err != nil {
		return identity.SMSDispatchResult{}, err
	}
	result, err := sender.Send(ctx, identity.SMSDispatchRequest{
		Phone:       phone,
		PhoneMasked: maskPhone(phone),
		Scene:       scene,
		Code:        code,
		ExpiresAt:   nowTime.Add(phoneCodeExpiry),
	})
	if err != nil {
		_ = s.phoneCodeRepo.CancelPhoneCodeSend(context.Background(), phoneKey, reservationToken)
		return identity.SMSDispatchResult{}, ErrPhoneCodeSendFailed
	}
	if err := s.phoneCodeRepo.ConfirmPhoneCodeSend(context.Background(), phoneKey, reservationToken, nowTime, dayKey); err != nil {
		_ = s.phoneCodeRepo.CancelPhoneCodeSend(context.Background(), phoneKey, reservationToken)
		return identity.SMSDispatchResult{}, ErrPhoneCodeSendFailed
	}
	return result, nil
}

func (s *Service) VerifyPhoneCode(phone string, code string) error {
	_, err := s.consumePhoneCode(phone, code, map[string]bool{"login": true, "invite_register": true})
	return err
}

func (s *Service) consumePhoneCode(phone string, code string, allowedScenes map[string]bool) (phoneCodeState, error) {
	phone = strings.TrimSpace(phone)
	code = strings.TrimSpace(code)
	if !validMainlandPhone(phone) {
		return phoneCodeState{}, ErrPhoneInvalid
	}
	phoneKey := s.phoneLookupHash(phone)
	if s.phoneCodeRepo != nil {
		claimToken, err := randomToken(16)
		if err != nil {
			return phoneCodeState{}, err
		}
		candidateHashes := make(map[string]string, len(allowedScenes))
		for scene, allowed := range allowedScenes {
			if allowed {
				candidateHashes[scene] = s.phoneCodeHash(phoneKey, scene, code)
			}
		}
		claim, verified, err := s.phoneCodeRepo.ClaimPhoneCode(
			context.Background(), phoneKey, candidateHashes, allowedScenes, claimToken,
			time.Now(), phoneCodeMaxFailures, phoneCodeResendInterval,
		)
		if err != nil {
			return phoneCodeState{}, err
		}
		if !verified {
			return phoneCodeState{}, ErrPhoneCodeInvalid
		}
		return phoneCodeState{
			scene: claim.Scene, sentAt: claim.SentAt, phoneKey: claim.PhoneKey,
			claimToken: claim.ClaimToken, repository: true,
		}, nil
	}
	s.phoneCodeMu.Lock()
	defer s.phoneCodeMu.Unlock()
	state, ok := s.phoneCodes[phoneKey]
	allowTemporaryCode := false
	if sender, ok := s.phoneSMSSender.(interface{ AllowsTemporaryCode() bool }); ok {
		allowTemporaryCode = sender.AllowsTemporaryCode()
	}
	if ok && state.sending {
		return phoneCodeState{}, ErrPhoneCodeInvalid
	}
	if !ok || state.sentAt.IsZero() || time.Since(state.sentAt) > phoneCodeExpiry {
		if allowTemporaryCode && code == temporaryPhoneCode {
			return phoneCodeState{code: code, scene: "local"}, nil
		}
		delete(s.phoneCodes, phoneKey)
		return phoneCodeState{}, ErrPhoneCodeInvalid
	}
	if len(allowedScenes) > 0 && !allowedScenes[state.scene] {
		return phoneCodeState{}, ErrPhoneCodeInvalid
	}
	if state.code != code && !(allowTemporaryCode && code == temporaryPhoneCode) {
		state.failedAttempts++
		if state.failedAttempts >= phoneCodeMaxFailures {
			delete(s.phoneCodes, phoneKey)
		} else {
			s.phoneCodes[phoneKey] = state
		}
		return phoneCodeState{}, ErrPhoneCodeInvalid
	}
	delete(s.phoneCodes, phoneKey)
	state.phoneKey = phoneKey
	return state, nil
}

func (s *Service) restorePhoneCode(phone string, state phoneCodeState) {
	_ = s.restorePhoneCodeWithError(phone, state)
}

func (s *Service) restorePhoneCodeWithError(phone string, state phoneCodeState) error {
	if state.repository {
		return s.phoneCodeRepo.RestorePhoneCodeClaim(context.Background(), state.phoneKey, state.claimToken)
	}
	if state.code == "" || state.scene == "local" {
		return nil
	}
	phoneKey := state.phoneKey
	if phoneKey == "" {
		phoneKey = s.phoneLookupHash(strings.TrimSpace(phone))
	}
	s.phoneCodeMu.Lock()
	if _, exists := s.phoneCodes[phoneKey]; !exists {
		s.phoneCodes[phoneKey] = state
	}
	s.phoneCodeMu.Unlock()
	return nil
}

func normalizePhoneCodeScene(scene string) string {
	switch strings.ToLower(strings.TrimSpace(scene)) {
	case "invite_register", "register":
		return "invite_register"
	case "password_reset":
		return "password_reset"
	default:
		return "login"
	}
}

func (s *Service) InvitePrecheck(req InvitePrecheckRequest) (invites.PrecheckResult, error) {
	if strings.TrimSpace(req.InviteCode) == "" {
		return invites.PrecheckResult{}, ErrInviteRequired
	}
	result, err := s.invites.Precheck(req.InviteCode, req.EntryType)
	if err != nil {
		return invites.PrecheckResult{}, err
	}
	if !result.Valid {
		if result.FailureReason == "inactive" || result.FailureReason == "inactive_or_exhausted" {
			return result, ErrInviteExpired
		}
		return result, ErrInvalidInvite
	}
	if result.BoundWechat && result.BoundUserID > 0 {
		boundUser, found, findErr := s.users.FindByID(result.BoundUserID)
		if findErr != nil {
			return invites.PrecheckResult{}, findErr
		}
		if !found || boundUser.Status == "deleted" {
			result.Valid = false
			result.AuthPageMode = invites.AuthPageModeRegister
			result.FailureReason = "account_deleted"
			return result, ErrInviteExpired
		}
	}
	return result, nil
}

func (s *Service) IssueInviteEntry(ownerID int64, entryType string) (invites.InviteCode, error) {
	entryType, ok := invites.ParseEntryType(entryType)
	if !ok || strings.TrimSpace(entryType) == "" {
		return invites.InviteCode{}, invites.ErrInvalidEntryType
	}
	return s.invites.IssueEntryCode(ownerID, entryType)
}

// IssueAssignedInviteEntry consumes an invitation already allocated by the
// administrator. It deliberately never creates a new code on the C end.
func (s *Service) IssueAssignedInviteEntry(ownerID int64, entryType string) (invites.InviteCode, error) {
	entryType, ok := invites.ParseEntryType(entryType)
	if !ok || strings.TrimSpace(entryType) == "" {
		return invites.InviteCode{}, invites.ErrInvalidEntryType
	}
	items, err := s.AdminInviteCodes(invites.CodeFilter{OwnerID: ownerID, EntryType: entryType})
	if err != nil {
		return invites.InviteCode{}, err
	}
	for _, item := range items {
		if item.Status == invites.StatusActive && item.MaxUses == 1 && item.UsedCount == 0 && item.BoundWechatUserID == 0 {
			return item, nil
		}
	}
	return invites.InviteCode{}, ErrInviteQuotaExceeded
}

func (s *Service) AdminCreateInviteCode(code string, ownerID int64, maxUses int, entryType string) (invites.InviteCode, error) {
	entryType, ok := invites.ParseEntryType(entryType)
	if !ok {
		return invites.InviteCode{}, invites.ErrInvalidEntryType
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return s.invites.IssueEntryCode(ownerID, entryType)
	}
	if maxUses < 0 {
		maxUses = 0
	}
	return s.invites.UpsertCodeWithEntryTypeResult(code, ownerID, maxUses, entryType)
}

func (s *Service) AdminCreateInviteCodes(codes []string, ownerID int64, maxUses int, entryType string, expiresAt time.Time) ([]invites.InviteCode, error) {
	entryType, ok := invites.ParseEntryType(entryType)
	if !ok || entryType == "" {
		return nil, invites.ErrInvalidEntryType
	}
	if maxUses < 0 {
		maxUses = 0
	}
	items := make([]invites.InviteCode, 0, len(codes))
	for _, code := range codes {
		items = append(items, invites.InviteCode{
			Code:      strings.TrimSpace(code),
			OwnerID:   ownerID,
			MaxUses:   maxUses,
			EntryType: entryType,
			ExpiresAt: expiresAt,
		})
	}
	return s.invites.CreateCodes(items)
}

func (s *Service) AdminInviteCodes(filter invites.CodeFilter) ([]invites.InviteCode, error) {
	if filter.EntryType != "" {
		entryType, ok := invites.ParseEntryType(filter.EntryType)
		if !ok {
			return nil, invites.ErrInvalidEntryType
		}
		filter.EntryType = entryType
	}
	return s.invites.ListCodes(filter)
}

func (s *Service) AdminDisableInviteCode(code string) (invites.InviteCode, error) {
	return s.invites.DisableCode(code)
}

func (s *Service) AdminEnableInviteCode(code string) (invites.InviteCode, error) {
	return s.invites.EnableCode(code)
}

func (s *Service) AdminVoidInviteCode(code string) (invites.InviteCode, error) {
	return s.invites.VoidCode(code)
}

func (s *Service) AdminUpdateUnusedInviteCode(code string, ownerID int64, entryType string, expiresAt time.Time) (invites.InviteCode, error) {
	return s.invites.UpdateUnusedCode(code, ownerID, entryType, expiresAt)
}

func (s *Service) CreateInviteQuotaRequest(ownerID int64, quantity int, reason string) (invites.QuotaRequest, error) {
	return s.invites.CreateQuotaRequest(ownerID, quantity, reason)
}

func (s *Service) InviteQuotaRequests(ownerID int64, status string) ([]invites.QuotaRequest, error) {
	return s.invites.ListQuotaRequests(ownerID, status)
}

func (s *Service) ReviewInviteQuotaRequest(id int64, status string, auditReason string, reviewedBy int64) (invites.QuotaRequest, error) {
	return s.invites.ReviewQuotaRequest(id, status, auditReason, reviewedBy)
}

func (s *Service) ApproveInviteQuotaRequest(id int64, auditReason string, reviewedBy int64, entryType string, expiresAt time.Time) (invites.QuotaRequest, []invites.InviteCode, error) {
	entryType, ok := invites.ParseEntryType(entryType)
	if !ok || entryType == "" {
		return invites.QuotaRequest{}, nil, invites.ErrInvalidEntryType
	}
	requests, err := s.invites.ListQuotaRequests(0, "pending")
	if err != nil {
		return invites.QuotaRequest{}, nil, err
	}
	var target invites.QuotaRequest
	for _, request := range requests {
		if request.ID == id {
			target = request
			break
		}
	}
	if target.ID == 0 {
		return invites.QuotaRequest{}, nil, errors.New("invite quota request not found or reviewed")
	}
	items, err := s.invites.PrepareEntryCodes(target.OwnerUserID, target.Quantity, entryType, expiresAt)
	if err != nil {
		return invites.QuotaRequest{}, nil, err
	}
	return s.invites.ApproveQuotaRequestWithCodes(id, auditReason, reviewedBy, items)
}

func (s *Service) AdminInviteRelations(filter invites.RelationFilter) ([]invites.Relation, error) {
	return s.invites.ListRelations(filter)
}

func (s *Service) InviteRelationForUser(userID int64) (invites.Relation, bool, error) {
	return s.invites.RelationForUser(userID)
}

func (s *Service) SetInviteRelationInviter(inviteeUserID int64, inviterUserID int64, source string) (invites.Relation, error) {
	relation, err := s.invites.SetRelationInviter(inviteeUserID, inviterUserID, source)
	if err == nil {
		s.notifyInviteRelationBound(relation)
	}
	return relation, err
}

func (s *Service) WechatEntryPrecheck(code string) (WechatEntryPrecheckResponse, error) {
	sessionInfo, err := s.resolver.Resolve(context.Background(), strings.TrimSpace(code))
	if err != nil {
		return WechatEntryPrecheckResponse{}, err
	}
	openID := strings.TrimSpace(sessionInfo.OpenID)
	if openID == "" {
		return WechatEntryPrecheckResponse{}, ErrWechatCodeInvalid
	}
	user, found, err := s.users.FindByOpenID(openID)
	if err != nil {
		return WechatEntryPrecheckResponse{}, err
	}
	return WechatEntryPrecheckResponse{
		BoundWechat:    found,
		RequiresInvite: !found,
		RequiresReauth: found && s.tokens.RequiresAppReauthentication(user.ID),
	}, nil
}

func (s *Service) WechatLogin(req WechatLoginRequest) (LoginResponse, error) {
	sessionInfo, err := s.resolver.Resolve(context.Background(), strings.TrimSpace(req.Code))
	if err != nil {
		return LoginResponse{}, err
	}
	openID := strings.TrimSpace(sessionInfo.OpenID)
	if openID == "" {
		return LoginResponse{}, ErrWechatCodeInvalid
	}
	entryType := strings.TrimSpace(req.EntryType)
	authPageMode := invites.AuthPageModeLogin
	boundWechat := false

	user, existed, err := s.users.FindByOpenID(openID)
	if err != nil {
		return LoginResponse{}, err
	}
	var existingRelation invites.Relation
	hasExistingRelation := false
	if existed {
		existingRelation, hasExistingRelation, err = s.invites.RelationForUser(user.ID)
		if err != nil {
			return LoginResponse{}, err
		}
	}
	if strings.TrimSpace(req.InviteCode) == "" {
		if !existed {
			return LoginResponse{}, ErrInviteRequired
		}
		if !hasExistingRelation {
			// An existing WeChat account remains an existing account even when
			// legacy invite relations have been cleared by operations.
			return s.boundWechatLoginResponse(user, nil, "", "", "")
		}
		entryType = entryTypeFromBindSource(existingRelation.BindSource)
		return s.boundWechatLoginResponse(user, &existingRelation, entryType, "", "")
	}
	if existed && !hasExistingRelation {
		return s.boundWechatLoginResponse(user, nil, "", "", "")
	}
	if existed && hasExistingRelation {
		entryType = entryTypeFromBindSource(existingRelation.BindSource)
		status := ""
		message := ""
		sameInvite, err := s.requestInviteMatchesRelation(req.InviteCode, existingRelation)
		if err != nil {
			return LoginResponse{}, err
		}
		if !sameInvite {
			status = InviteBindingStatusAlreadyBound
			message = inviteAlreadyBoundMessage
		}
		return s.boundWechatLoginResponse(user, &existingRelation, entryType, status, message)
	}
	precheck, err := s.InvitePrecheck(InvitePrecheckRequest{InviteCode: req.InviteCode, EntryType: entryType})
	if err != nil {
		return LoginResponse{}, err
	}
	entryType = precheck.EntryType
	authPageMode = precheck.AuthPageMode
	boundWechat = precheck.BoundWechat
	if precheck.BoundWechat && precheck.BoundUserID != user.ID {
		return LoginResponse{}, ErrInviteAlreadyBound
	}
	if !existed {
		user, err = s.users.Create(openID)
		if err != nil {
			return LoginResponse{}, err
		}
	}
	if !precheck.BoundWechat || precheck.BoundUserID != user.ID {
		invite := precheck.InviteCode
		relation, err := s.invites.Bind(invite, user.ID, "invite_"+entryType)
		if err != nil {
			if errors.Is(err, invites.ErrInviteAlreadyBound) {
				return LoginResponse{}, ErrInviteAlreadyBound
			}
			if errors.Is(err, invites.ErrInviteInactive) {
				return LoginResponse{}, ErrInvalidInvite
			}
			return LoginResponse{}, err
		}
		s.notifyInviteRelationBound(relation)
	}
	if existed && !precheck.BoundWechat {
		authPageMode = invites.AuthPageModeLogin
		boundWechat = true
	}

	session, err := s.tokens.IssuePreAuth(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}

	var relationPtr *invites.Relation
	relation, ok, err := s.invites.RelationForUser(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}
	if ok {
		relationPtr = &relation
	}

	return LoginResponse{
		PreAuthToken:            session.Token,
		ExpiresAt:               session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User:                    user,
		InviteRelation:          relationPtr,
		NeedProfile:             user.Nickname == "",
		RequiresIdentityBinding: true,
		IdentityBindStatus:      "wechat_logged_in",
		EntryType:               entryType,
		AuthPageMode:            authPageMode,
		BoundWechat:             boundWechat,
	}, nil
}

func (s *Service) boundWechatLoginResponse(user users.User, relation *invites.Relation, entryType string, bindingStatus string, bindingMessage string) (LoginResponse, error) {
	// A user who has already completed phone binding is an existing account.
	// Reissuing a pre-auth session here would make every WeChat return look
	// unfinished and, importantly, would not refresh the 60-day login record.
	identityBound := s.phoneBindingVerified(user.ID, user.PhoneMasked)
	var session Session
	var err error
	if identityBound {
		session, err = s.tokens.IssueApp(user.ID)
	} else {
		session, err = s.tokens.IssuePreAuth(user.ID)
	}
	if err != nil {
		return LoginResponse{}, err
	}
	response := LoginResponse{
		ExpiresAt:               session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User:                    user,
		InviteRelation:          relation,
		NeedProfile:             user.Nickname == "",
		RequiresIdentityBinding: !identityBound,
		IdentityBindStatus:      "sms_verified",
		EntryType:               entryType,
		AuthPageMode:            invites.AuthPageModeLogin,
		BoundWechat:             true,
		InviteBindingStatus:     bindingStatus,
		InviteBindingMessage:    bindingMessage,
	}
	if identityBound {
		response.Token = session.Token
	} else {
		response.PreAuthToken = session.Token
		response.IdentityBindStatus = "wechat_logged_in"
	}
	return response, nil
}

func (s *Service) requestInviteMatchesRelation(inviteCode string, relation invites.Relation) (bool, error) {
	invite, ok, err := s.invites.FindCode(inviteCode)
	if err != nil {
		return false, err
	}
	return ok && invite.ID == relation.InviteCodeID, nil
}

func entryTypeFromBindSource(source string) string {
	entryType := strings.TrimPrefix(strings.TrimSpace(source), "invite_")
	if normalized, ok := invites.ParseEntryType(entryType); ok {
		return normalized
	}
	return ""
}

func (s *Service) PhoneLogin(req PhoneLoginRequest) (response LoginResponse, resultErr error) {
	phone := strings.TrimSpace(req.Phone)
	code := strings.TrimSpace(req.Code)
	if phone == "" {
		return LoginResponse{}, ErrPhoneRequired
	}
	if !validMainlandPhone(phone) {
		return LoginResponse{}, ErrPhoneInvalid
	}
	phoneHash := s.phoneLookupHash(phone)
	user, existed, err := s.users.FindByPhoneHash(phoneHash)
	if err != nil {
		return LoginResponse{}, err
	}
	authPageMode := invites.AuthPageModeLogin
	entryType := strings.TrimSpace(req.EntryType)
	var precheck invites.PrecheckResult
	if strings.TrimSpace(req.InviteCode) != "" {
		precheck, err = s.InvitePrecheck(InvitePrecheckRequest{
			InviteCode: req.InviteCode,
			EntryType:  entryType,
		})
		if err != nil {
			return LoginResponse{}, err
		}
		entryType = precheck.EntryType
		if precheck.BoundWechat && (!existed || precheck.BoundUserID != user.ID) {
			return LoginResponse{}, ErrInviteAlreadyBound
		}
	} else if !existed {
		return LoginResponse{}, ErrInviteRequired
	}
	// Registration prerequisites are checked before consuming the one-time SMS
	// code. A missing or invalid invite must not force the user to request a new
	// verification code.
	consumedCode, err := s.consumePhoneCode(phone, code, map[string]bool{"login": true, "invite_register": true})
	if err != nil {
		if errors.Is(err, ErrPhoneCodeInvalid) || errors.Is(err, ErrPhoneInvalid) {
			return LoginResponse{}, ErrPhoneCodeInvalid
		}
		return LoginResponse{}, err
	}
	defer func() {
		if resultErr != nil {
			s.restorePhoneCode(phone, consumedCode)
		}
	}()

	if !existed {
		user, err = s.users.CreateWithPhone(phoneHash, maskPhone(phone))
		if err != nil {
			return LoginResponse{}, err
		}
		authPageMode = invites.AuthPageModeRegister
	}

	if strings.TrimSpace(req.InviteCode) != "" && (!precheck.BoundWechat || precheck.BoundUserID != user.ID) {
		relation, err := s.invites.Bind(precheck.InviteCode, user.ID, "invite_"+entryType)
		if err != nil {
			if !existed {
				_ = s.users.DeleteAccount(user.ID)
			}
			if errors.Is(err, invites.ErrInviteAlreadyBound) {
				return LoginResponse{}, ErrInviteAlreadyBound
			}
			if errors.Is(err, invites.ErrInviteInactive) {
				return LoginResponse{}, ErrInvalidInvite
			}
			return LoginResponse{}, err
		}
		s.notifyInviteRelationBound(relation)
	}

	session, err := s.tokens.IssueApp(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}
	var relationPtr *invites.Relation
	relation, ok, err := s.invites.RelationForUser(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}
	if ok {
		relationPtr = &relation
	}

	return LoginResponse{
		Token:                   session.Token,
		ExpiresAt:               session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User:                    user,
		InviteRelation:          relationPtr,
		NeedProfile:             user.Nickname == "",
		RequiresIdentityBinding: false,
		IdentityBindStatus:      "sms_verified",
		EntryType:               entryType,
		AuthPageMode:            authPageMode,
		BoundWechat:             user.OpenID != "",
		phoneCodeClaim:          consumedCode,
	}, nil
}

// RollbackPhoneLogin is used when the application layer cannot complete the
// post-login identity synchronization. The client never received the token, so
// the issued session is revoked and the one-time code is made retryable.
func (s *Service) RollbackPhoneLogin(response LoginResponse) error {
	restoreErr := s.restorePhoneCodeWithError("", response.phoneCodeClaim)
	revokeErr := s.tokens.RevokeSession(response.Token)
	if restoreErr != nil {
		return restoreErr
	}
	return revokeErr
}

func (s *Service) BindPhoneAuth(userID int64, phone string) (users.User, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return users.User{}, ErrPhoneRequired
	}
	if !validMainlandPhone(phone) {
		return users.User{}, ErrPhoneInvalid
	}
	return s.users.BindPhoneAuth(userID, s.phoneLookupHash(phone), maskPhone(phone))
}

func (s *Service) SetPassword(userID int64, password string) (users.User, error) {
	if userID <= 0 || !validLoginPassword(password) {
		return users.User{}, ErrPasswordInvalid
	}
	hash, err := passwordHash(password)
	if err != nil {
		return users.User{}, err
	}
	return s.users.UpdatePasswordHash(userID, hash)
}

// ResetPasswordByPhone is deliberately a single verification-and-update
// transaction at the service boundary.  The SMS code is consumed here, so a
// client cannot first consume it in a "verify" step and then fail to reset the
// password with the same code.
func (s *Service) ResetPasswordByPhone(phone string, code string, password string) (users.User, error) {
	phone = strings.TrimSpace(phone)
	if !validMainlandPhone(phone) {
		return users.User{}, ErrPhoneInvalid
	}
	// 先校验新密码，避免用户因一次格式输入错误就消耗短信验证码。
	if !validLoginPassword(password) {
		return users.User{}, ErrPasswordInvalid
	}
	user, found, err := s.users.FindByPhoneHash(s.phoneLookupHash(phone))
	if err != nil {
		return users.User{}, err
	}
	if !found {
		return users.User{}, ErrPasswordInvalid
	}
	consumedCode, err := s.consumePhoneCode(phone, code, map[string]bool{"password_reset": true})
	if err != nil {
		if errors.Is(err, ErrPhoneCodeInvalid) || errors.Is(err, ErrPhoneInvalid) {
			return users.User{}, ErrPhoneCodeInvalid
		}
		return users.User{}, err
	}
	updated, err := s.SetPassword(user.ID, password)
	if err != nil {
		s.restorePhoneCode(phone, consumedCode)
		return users.User{}, err
	}
	return updated, nil
}

func (s *Service) HasPassword(userID int64) bool {
	_, ok, err := s.users.PasswordHash(userID)
	return err == nil && ok
}

func (s *Service) PasswordLogin(req PasswordLoginRequest) (LoginResponse, error) {
	phone := strings.TrimSpace(req.Phone)
	if !validMainlandPhone(phone) {
		return LoginResponse{}, ErrPhoneInvalid
	}
	user, found, err := s.users.FindByPhoneHash(s.phoneLookupHash(phone))
	if err != nil {
		return LoginResponse{}, err
	}
	if !found {
		return LoginResponse{}, ErrPasswordInvalid
	}
	hash, configured, err := s.users.PasswordHash(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}
	if !configured {
		return LoginResponse{}, ErrPasswordNotSet
	}
	if !verifyPassword(hash, req.Password) {
		return LoginResponse{}, ErrPasswordInvalid
	}
	session, err := s.tokens.IssueApp(user.ID)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{Token: session.Token, ExpiresAt: session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"), User: user, NeedProfile: user.Nickname == "", IdentityBindStatus: "sms_verified", AuthPageMode: invites.AuthPageModeLogin, BoundWechat: user.OpenID != ""}, nil
}

func (s *Service) BindWechat(userID int64, code string) (users.User, error) {
	session, err := s.resolver.Resolve(context.Background(), strings.TrimSpace(code))
	if err != nil || strings.TrimSpace(session.OpenID) == "" {
		return users.User{}, ErrWechatCodeInvalid
	}
	if owner, found, err := s.users.FindByOpenID(session.OpenID); err != nil {
		return users.User{}, err
	} else if found && owner.ID != userID {
		return users.User{}, ErrWechatAlreadyBound
	}
	user, err := s.users.BindWechat(userID, session.OpenID)
	if errors.Is(err, users.ErrWechatAlreadyUsed) || errors.Is(err, users.ErrWechatConflict) {
		return users.User{}, ErrWechatAlreadyBound
	}
	return user, err
}

func validLoginPassword(value string) bool {
	if len(value) < 8 || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func passwordHash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := derivePasswordKey([]byte(password), salt, 180000, 32)
	return base64.RawStdEncoding.EncodeToString(salt) + ":" + base64.RawStdEncoding.EncodeToString(key), nil
}

func verifyPassword(encoded string, password string) bool {
	parts := strings.Split(encoded, ":")
	if len(parts) != 2 || !validLoginPassword(password) {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	actual := derivePasswordKey([]byte(password), salt, 180000, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

// derivePasswordKey implements PBKDF2-HMAC-SHA256 without relying on a Go
// standard-library package introduced after the production build image's Go 1.22.
func derivePasswordKey(password, salt []byte, iterations, keyLen int) []byte {
	if iterations <= 0 || keyLen <= 0 {
		return nil
	}
	hashLen := sha256.Size
	blockCount := (keyLen + hashLen - 1) / hashLen
	output := make([]byte, 0, blockCount*hashLen)
	for block := 1; block <= blockCount; block++ {
		input := append(append([]byte(nil), salt...), byte(block>>24), byte(block>>16), byte(block>>8), byte(block))
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(input)
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for round := 1; round < iterations; round++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(u)
			u = mac.Sum(nil)
			for index := range t {
				t[index] ^= u[index]
			}
		}
		output = append(output, t...)
	}
	return output[:keyLen]
}

func (s *Service) IssueAppToken(userID int64) (Session, error) {
	return s.tokens.IssueApp(userID)
}

// DeleteAccount releases all account login bindings. Invite usage is never
// restored, so a deleted account must obtain a new invitation to register.
func (s *Service) DeleteAccount(userID int64) error {
	if userID <= 0 {
		return users.ErrInvalidProfile
	}
	if err := s.invites.ClearInviteeBindings(userID); err != nil {
		return err
	}
	if err := s.users.DeleteAccount(userID); err != nil {
		return err
	}
	return s.tokens.RevokeUserSessions(userID)
}

func (s *Service) ActiveAppSessionCount() int {
	count, _ := s.ActiveAppSessionCountStrict()
	return count
}

func (s *Service) ActiveAppSessionCountStrict() (int, error) {
	if s == nil || s.tokens == nil {
		return 0, nil
	}
	return s.tokens.ActiveSessionCountStrict(SessionKindApp)
}

func (s *Service) CurrentUser(token string) (users.User, bool) {
	return s.currentUserForSessionKinds(token, SessionKindApp)
}

// CurrentIdentityUser accepts the short-lived pre-auth session only for the
// explicitly whitelisted onboarding and identity endpoints.
func (s *Service) CurrentIdentityUser(token string) (users.User, bool) {
	return s.currentUserForSessionKinds(token, SessionKindApp, SessionKindPreAuth)
}

func (s *Service) currentUserForSessionKinds(token string, kinds ...string) (users.User, bool) {
	session, ok := s.tokens.Verify(token)
	if !ok {
		return users.User{}, false
	}
	allowed := false
	for _, kind := range kinds {
		if session.Kind == kind {
			allowed = true
			break
		}
	}
	if !allowed {
		return users.User{}, false
	}
	user, ok, err := s.users.FindByID(session.UserID)
	if err != nil {
		return users.User{}, false
	}
	return user, ok && user.Status != "deleted"
}

func (s *Service) UserByID(userID int64) (users.User, bool) {
	user, ok, err := s.users.FindByID(userID)
	if err != nil {
		return users.User{}, false
	}
	return user, ok
}

func (s *Service) AdminUsers(filter users.Filter) ([]users.User, error) {
	return s.users.List(filter)
}

func (s *Service) InviteCodeForUser(userID int64) (string, error) {
	items, err := s.invites.ListCodes(invites.CodeFilter{OwnerID: userID})
	if err != nil {
		return "", err
	}
	for _, invite := range items {
		if invite.Status == invites.StatusActive && invite.MaxUses == 1 && invite.UsedCount == 0 && invite.BoundWechatUserID == 0 {
			return invite.Code, nil
		}
	}
	return "", nil
}

func (s *Service) UpdateProfile(userID int64, nickname string, avatarURL string, avatarFileID int64) (users.User, error) {
	return s.users.UpdateProfile(userID, nickname, avatarURL, avatarFileID)
}

func (s *Service) UpdateRealnameStatus(userID int64, status string) (users.User, error) {
	return s.users.UpdateRealnameStatus(userID, status)
}

func mockOpenID(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	return "mock_openid_" + code
}

func validMainlandPhone(phone string) bool {
	if len(phone) != 11 || phone[0] != '1' {
		return false
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

func (s *Service) phoneLookupHash(phone string) string {
	mac := hmac.New(sha256.New, []byte(s.phoneLookupSecret))
	_, _ = mac.Write([]byte(strings.TrimSpace(phone)))
	return fmt.Sprintf("hmac-sha256:%x", mac.Sum(nil))
}

func (s *Service) phoneCodeHash(phoneKey string, scene string, code string) string {
	mac := hmac.New(sha256.New, []byte(s.phoneLookupSecret))
	_, _ = mac.Write([]byte("phone-code:" + phoneKey + ":" + scene + ":" + strings.TrimSpace(code)))
	return fmt.Sprintf("hmac-sha256:%x", mac.Sum(nil))
}

type WechatSession struct {
	OpenID     string
	UnionID    string
	SessionKey string
}

type WechatCodeResolver interface {
	Resolve(ctx context.Context, code string) (WechatSession, error)
}

type MockWechatCodeResolver struct{}

func (MockWechatCodeResolver) Resolve(_ context.Context, code string) (WechatSession, error) {
	openID := mockOpenID(code)
	if openID == "" {
		return WechatSession{}, ErrWechatCodeInvalid
	}
	return WechatSession{OpenID: openID}, nil
}

type WechatAPIResolver struct {
	AppID      string
	AppSecret  string
	HTTPClient *http.Client
	Endpoint   string
}

func NewWechatAPIResolver(appID string, appSecret string) *WechatAPIResolver {
	return &WechatAPIResolver{
		AppID:      strings.TrimSpace(appID),
		AppSecret:  strings.TrimSpace(appSecret),
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		Endpoint:   "https://api.weixin.qq.com/sns/jscode2session",
	}
}

func (r *WechatAPIResolver) Resolve(ctx context.Context, code string) (WechatSession, error) {
	code = strings.TrimSpace(code)
	if code == "" || strings.TrimSpace(r.AppID) == "" || strings.TrimSpace(r.AppSecret) == "" {
		return WechatSession{}, ErrWechatCodeInvalid
	}
	endpoint := strings.TrimSpace(r.Endpoint)
	if endpoint == "" {
		endpoint = "https://api.weixin.qq.com/sns/jscode2session"
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return WechatSession{}, err
	}
	q := parsed.Query()
	q.Set("appid", r.AppID)
	q.Set("secret", r.AppSecret)
	q.Set("js_code", code)
	q.Set("grant_type", "authorization_code")
	parsed.RawQuery = q.Encode()

	client := r.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return WechatSession{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return WechatSession{}, err
	}
	defer resp.Body.Close()

	var payload struct {
		OpenID     string `json:"openid"`
		UnionID    string `json:"unionid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return WechatSession{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WechatSession{}, fmt.Errorf("wechat jscode2session status %d", resp.StatusCode)
	}
	if payload.ErrCode != 0 {
		return WechatSession{}, ErrWechatCodeInvalid
	}
	if strings.TrimSpace(payload.OpenID) == "" {
		return WechatSession{}, ErrWechatCodeInvalid
	}
	return WechatSession{OpenID: payload.OpenID, UnionID: payload.UnionID, SessionKey: payload.SessionKey}, nil
}
