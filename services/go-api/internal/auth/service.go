package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

var (
	ErrInviteRequired     = errors.New("invite required")
	ErrInvalidInvite      = errors.New("invalid invite")
	ErrInviteAlreadyBound = errors.New("invite already bound")
	ErrWechatCodeInvalid  = errors.New("wechat code invalid")
)

type InvitePrecheckRequest struct {
	InviteCode string `json:"inviteCode"`
	EntryType  string `json:"entryType"`
}

type WechatLoginRequest struct {
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
}

type Service struct {
	users    *users.Store
	invites  *invites.Store
	tokens   *TokenStore
	resolver WechatCodeResolver
}

func NewService(userStore *users.Store, inviteStore *invites.Store, tokenStore *TokenStore) *Service {
	return &Service{users: userStore, invites: inviteStore, tokens: tokenStore, resolver: MockWechatCodeResolver{}}
}

func (s *Service) UseWechatCodeResolver(resolver WechatCodeResolver) {
	if resolver == nil {
		return
	}
	s.resolver = resolver
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
		return result, ErrInvalidInvite
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
	return s.invites.UpsertCodeWithEntryType(code, ownerID, maxUses, entryType), nil
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

func (s *Service) AdminInviteRelations(filter invites.RelationFilter) ([]invites.Relation, error) {
	return s.invites.ListRelations(filter)
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
	if strings.TrimSpace(req.InviteCode) == "" {
		return LoginResponse{}, ErrInviteRequired
	}
	entryType := strings.TrimSpace(req.EntryType)
	authPageMode := invites.AuthPageModeLogin
	boundWechat := false

	user, existed, err := s.users.FindByOpenID(openID)
	if err != nil {
		return LoginResponse{}, err
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
		if _, err := s.invites.Bind(invite, user.ID, "invite_"+entryType); err != nil {
			if errors.Is(err, invites.ErrInviteAlreadyBound) {
				return LoginResponse{}, ErrInviteAlreadyBound
			}
			return LoginResponse{}, err
		}
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

func (s *Service) IssueAppToken(userID int64) (Session, error) {
	return s.tokens.IssueApp(userID)
}

func (s *Service) CurrentUser(token string) (users.User, bool) {
	session, ok := s.tokens.Verify(token)
	if !ok {
		return users.User{}, false
	}
	user, ok, err := s.users.FindByID(session.UserID)
	if err != nil {
		return users.User{}, false
	}
	return user, ok
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
	invite, err := s.invites.EnsureCodeForOwner(userID)
	if err != nil {
		return "", err
	}
	return invite.Code, nil
}

func (s *Service) UpdateProfile(userID int64, nickname string, avatarURL string, avatarFileID int64) (users.User, error) {
	return s.users.UpdateProfile(userID, nickname, avatarURL, avatarFileID)
}

func mockOpenID(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	if openID := strings.TrimSpace(os.Getenv("MOCK_WECHAT_OPENID")); openID != "" {
		return openID
	}
	return "mock_openid_local"
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
