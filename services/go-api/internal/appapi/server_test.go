package appapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/common/config"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/im"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/lbs"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/redemption"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/systemconfig"
	"zhw-mini/services/go-api/internal/users"
)

type fakeMapProvider struct {
	searchReq  lbs.MapSearchRequest
	geocodeReq lbs.MapGeocodeRequest
	reverseReq lbs.MapReverseGeocodeRequest
	routeReq   lbs.MapRouteRequest
}

func (f *fakeMapProvider) Search(ctx context.Context, req lbs.MapSearchRequest) (lbs.MapSearchResult, error) {
	f.searchReq = req
	if req.Keyword == "bad" {
		return lbs.MapSearchResult{}, lbs.ErrMapRequestInvalid
	}
	return lbs.MapSearchResult{
		Provider: "tencent",
		Total:    1,
		Items: []lbs.MapPlace{{
			ID:            "poi-1",
			Title:         "西湖",
			Address:       "杭州市西湖区",
			City:          "杭州",
			Longitude:     120.1551,
			Latitude:      30.2741,
			DistanceMeter: 1200,
		}},
	}, nil
}

func (f *fakeMapProvider) Geocode(ctx context.Context, req lbs.MapGeocodeRequest) (lbs.MapPlace, error) {
	f.geocodeReq = req
	if req.Address == "" {
		return lbs.MapPlace{}, lbs.ErrMapRequestInvalid
	}
	return lbs.MapPlace{Title: req.Address, Address: req.Address, City: req.City, Longitude: 120.1551, Latitude: 30.2741}, nil
}

func (f *fakeMapProvider) ReverseGeocode(ctx context.Context, req lbs.MapReverseGeocodeRequest) (lbs.MapPlace, error) {
	f.reverseReq = req
	return lbs.MapPlace{Title: "杭州市西湖区", Address: "杭州市西湖区", City: "杭州", Longitude: req.Longitude, Latitude: req.Latitude}, nil
}

func (f *fakeMapProvider) Route(ctx context.Context, req lbs.MapRouteRequest) (lbs.MapRoute, error) {
	f.routeReq = req
	return lbs.MapRoute{Provider: "tencent", Mode: req.Mode, DistanceMeter: 1500, DurationSecond: 600}, nil
}

var identityPhoneSeq int

func newTestAppServer(authService *auth.Service, identityService *identity.Service) *Server {
	gameService := games.NewService(identityService)
	return New(authService, identityService, gameService, lbs.NewService(), im.NewService(gameService))
}

type failingWechatResolver struct{}

func (failingWechatResolver) Resolve(context.Context, string) (auth.WechatSession, error) {
	return auth.WechatSession{}, auth.ErrWechatCodeInvalid
}

func TestAppBusinessRoutesRequireTokenMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{LBS: config.LBSConfig{DefaultRadiusMeter: 1}})
	server.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/app/games", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected games list to require token, got %d: %s", rec.Code, rec.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/app/auth/wechat-login", bytes.NewBufferString(`{"code":"auth-route","inviteCode":"TEST2026"}`))
	loginRec := httptest.NewRecorder()
	mux.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login route to stay public, got %d: %s", loginRec.Code, loginRec.Body.String())
	}
}

func TestExpertApplyConfigReadsSystemConfigHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(expertApplyConfigKey, map[string]interface{}{
		"skillOptions": []map[string]interface{}{
			{"name": "数据库技能", "active": true},
		},
		"fields": []map[string]interface{}{
			{"type": "input", "key": "dbField", "label": "数据库字段", "required": true},
		},
		"uploadField": map[string]interface{}{
			"label":    "数据库证明",
			"required": true,
			"maxCount": 2,
		},
		"validationRules": map[string]interface{}{
			"intro": map[string]interface{}{"minLength": 10, "maxLength": 200},
		},
		"yearOptions":  []string{"1年", "2年"},
		"serviceCount": 2,
		"priceHint":    "数据库价格提示",
	}); err != nil {
		t.Fatalf("seed expert apply config failed: %v", err)
	}
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "expert-apply-config")
	completeIdentityForTest(t, mux, token)
	body := getJSON(t, mux, "/api/app/role-applications/expert/config", token, http.StatusOK)
	var resp struct {
		Data struct {
			SkillOptions []struct {
				Name string `json:"name"`
			} `json:"skillOptions"`
			Fields []struct {
				Key string `json:"key"`
			} `json:"fields"`
			UploadField struct {
				Label    string `json:"label"`
				MaxCount int    `json:"maxCount"`
			} `json:"uploadField"`
			ServiceCount int    `json:"serviceCount"`
			PriceHint    string `json:"priceHint"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.SkillOptions) != 1 || resp.Data.SkillOptions[0].Name != "数据库技能" || len(resp.Data.Fields) != 1 || resp.Data.Fields[0].Key != "dbField" || resp.Data.UploadField.Label != "数据库证明" || resp.Data.UploadField.MaxCount != 2 || resp.Data.ServiceCount != 2 || resp.Data.PriceHint != "数据库价格提示" {
		t.Fatalf("expected expert apply config from system config: %s", string(body))
	}
}

func TestGuideApplyConfigFromSystemConfigHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(guideApplyConfigKey, map[string]interface{}{
		"applyRoleType": "guide",
		"applyRoleName": "数据库领路人",
		"requirements": []map[string]interface{}{
			{"title": "数据库条件", "text": "数据库条件说明", "done": true},
		},
		"fields": []map[string]interface{}{
			{"type": "input", "key": "dbGuideField", "label": "数据库领路字段", "required": true},
		},
		"uploadField": map[string]interface{}{
			"label":    "数据库领路证明",
			"required": true,
			"maxCount": 3,
		},
		"serviceCount": 1,
		"priceHint":    "数据库领路价格提示",
	}); err != nil {
		t.Fatalf("seed guide apply config failed: %v", err)
	}
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "guide-apply-config")
	completeIdentityForTest(t, mux, token)
	body := getJSON(t, mux, "/api/app/role-applications/guide/config", token, http.StatusOK)
	var resp struct {
		Data struct {
			ApplyRoleName string `json:"applyRoleName"`
			Requirements  []struct {
				Title string `json:"title"`
			} `json:"requirements"`
			Fields []struct {
				Key string `json:"key"`
			} `json:"fields"`
			UploadField struct {
				Label    string `json:"label"`
				MaxCount int    `json:"maxCount"`
			} `json:"uploadField"`
			ServiceCount int    `json:"serviceCount"`
			PriceHint    string `json:"priceHint"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ApplyRoleName != "数据库领路人" || len(resp.Data.Requirements) != 1 || resp.Data.Requirements[0].Title != "数据库条件" || len(resp.Data.Fields) != 1 || resp.Data.Fields[0].Key != "dbGuideField" || resp.Data.UploadField.Label != "数据库领路证明" || resp.Data.UploadField.MaxCount != 3 || resp.Data.ServiceCount != 1 || resp.Data.PriceHint != "数据库领路价格提示" {
		t.Fatalf("expected guide apply config from system config: %s", string(body))
	}
}

func TestApproveLocalDisabledInProduction(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{AppEnv: "production"})
	server.Register(mux)
	token := loginForTestWithCode(t, mux, "approve-local-prod")

	req := httptest.NewRequest(http.MethodPost, "/api/app/games/1/approve-local", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected approve-local disabled in production, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRegisteredRoutesReturnRequestID(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{LBS: config.LBSConfig{DefaultRadiusMeter: 1}})
	server.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/app/auth/wechat-login", bytes.NewBufferString(`{"code":"request-id","inviteCode":"TEST2026"}`))
	req.Header.Set("X-Request-Id", "rid-appapi")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Request-Id") != "rid-appapi" {
		t.Fatalf("expected request id header, got %q", rec.Header().Get("X-Request-Id"))
	}
	var resp struct {
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.RequestID != "rid-appapi" {
		t.Fatalf("expected response body request id, got %q body=%s", resp.RequestID, rec.Body.String())
	}
}

func TestInvitePrecheckAndWechatLoginFlowHTTP(t *testing.T) {
	mux := http.NewServeMux()
	inviteStore := invites.NewStore()
	inviteStore.UpsertCodeWithEntryType("POSTER2026", 0, 1, invites.EntryTypePoster)
	authService := auth.NewService(users.NewStore(), inviteStore, auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(gameInviteConfigKey, gameInviteConfigDTO{
		MinPlayerCount:      1,
		MaxPlayerCount:      2,
		BudgetMaxAmount:     6000,
		DefaultBudget:       "1200",
		DefaultTitle:        "数据库邀请测试标题",
		DefaultDetail:       "数据库邀请测试详情",
		PlayerIntroTemplate: "数据库模板-{expertName}",
		ActivityTypes: []inviteActivityTypeDTO{
			{Key: "db", Name: "数据库类型"},
		},
		RewardRateConfig: inviteRewardRateConfigDTO{
			PlatformServiceRate:   8,
			SystemGuideRewardRate: 12,
			InviteRewardRate:      30,
		},
	}); err != nil {
		t.Fatalf("seed invite config failed: %v", err)
	}
	server.Register(mux)

	precheckBody := postJSON(t, mux, "/api/app/invites/precheck", "", `{"inviteCode":"POSTER2026","entryType":"poster"}`, http.StatusOK)
	var precheck struct {
		Code int `json:"code"`
		Data struct {
			Valid        bool   `json:"valid"`
			EntryType    string `json:"entryType"`
			AuthPageMode string `json:"authPageMode"`
			BoundWechat  bool   `json:"boundWechat"`
		} `json:"data"`
	}
	if err := json.Unmarshal(precheckBody, &precheck); err != nil {
		t.Fatal(err)
	}
	if !precheck.Data.Valid || precheck.Data.EntryType != "poster" || precheck.Data.AuthPageMode != "register" || precheck.Data.BoundWechat {
		t.Fatalf("unexpected precheck response: %s", string(precheckBody))
	}

	loginBody := postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"precheck-http","inviteCode":"POSTER2026","entryType":"poster"}`, http.StatusOK)
	var login struct {
		Data struct {
			EntryType    string `json:"entryType"`
			AuthPageMode string `json:"authPageMode"`
			BoundWechat  bool   `json:"boundWechat"`
			PreAuthToken string `json:"preAuthToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginBody, &login); err != nil {
		t.Fatal(err)
	}
	if login.Data.EntryType != "poster" || login.Data.AuthPageMode != "register" || login.Data.BoundWechat || login.Data.PreAuthToken == "" {
		t.Fatalf("unexpected login response: %s", string(loginBody))
	}

	precheckBody = postJSON(t, mux, "/api/app/invites/precheck", "", `{"inviteCode":"POSTER2026","entryType":"poster"}`, http.StatusOK)
	if err := json.Unmarshal(precheckBody, &precheck); err != nil {
		t.Fatal(err)
	}
	if !precheck.Data.BoundWechat || precheck.Data.AuthPageMode != "login" {
		t.Fatalf("expected login mode after bind: %s", string(precheckBody))
	}
	postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"precheck-http-other","inviteCode":"POSTER2026","entryType":"poster"}`, http.StatusForbidden)
}

func TestWechatLoginInvalidCodeHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	authService.UseWechatCodeResolver(failingWechatResolver{})
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"bad-code","inviteCode":"TEST2026"}`, http.StatusUnprocessableEntity)
}

func TestCreateInviteEntryHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(gameInviteConfigKey, gameInviteConfigDTO{
		MinPlayerCount:      1,
		MaxPlayerCount:      2,
		BudgetMaxAmount:     6000,
		DefaultBudget:       "1200",
		DefaultTitle:        "数据库邀请测试标题",
		DefaultDetail:       "数据库邀请测试详情",
		PlayerIntroTemplate: "数据库模板-{expertName}",
		ActivityTypes:       []inviteActivityTypeDTO{{Key: "db", Name: "数据库类型"}},
		RewardRateConfig: inviteRewardRateConfigDTO{
			PlatformServiceRate:   8,
			SystemGuideRewardRate: 12,
			InviteRewardRate:      30,
		},
	}); err != nil {
		t.Fatalf("seed invite config failed: %v", err)
	}
	server.Register(mux)
	token := loginForTestWithCode(t, mux, "entry-owner")
	completeIdentityForTest(t, mux, token)
	postJSON(t, mux, "/api/app/games", token, `{"title":"周末城市探索","gameType":"free","minPlayers":5,"maxPlayers":8,"cityName":"杭州","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)

	firstBody := postJSON(t, mux, "/api/app/invites/entries", token, `{"entryType":"link","title":"邀请你加入真好玩"}`, http.StatusOK)
	posterBody := postJSON(t, mux, "/api/app/invites/entries", token, `{"entryType":"poster","gameId":1}`, http.StatusOK)
	secondBody := postJSON(t, mux, "/api/app/invites/entries", token, `{"entryType":"link"}`, http.StatusOK)
	var first struct {
		Data struct {
			InviteCode   string `json:"inviteCode"`
			EntryType    string `json:"entryType"`
			Path         string `json:"path"`
			Scene        string `json:"scene"`
			UrlLink      string `json:"urlLink"`
			UrlLinkError string `json:"urlLinkError"`
		} `json:"data"`
	}
	var poster struct {
		Data struct {
			InviteCode string `json:"inviteCode"`
			EntryType  string `json:"entryType"`
			Path       string `json:"path"`
			Scene      string `json:"scene"`
			Title      string `json:"title"`
			Game       struct {
				ID             int64  `json:"id"`
				Title          string `json:"title"`
				CityName       string `json:"cityName"`
				CurrentPlayers int    `json:"currentPlayers"`
				MinPlayers     int    `json:"minPlayers"`
				MaxPlayers     int    `json:"maxPlayers"`
			} `json:"game"`
		} `json:"data"`
	}
	var second struct {
		Data struct {
			InviteCode string `json:"inviteCode"`
			EntryType  string `json:"entryType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(firstBody, &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(secondBody, &second); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(posterBody, &poster); err != nil {
		t.Fatal(err)
	}
	if first.Data.InviteCode == "" || second.Data.InviteCode == "" || first.Data.InviteCode == second.Data.InviteCode {
		t.Fatalf("expected unique invite codes: first=%s second=%s", string(firstBody), string(secondBody))
	}
	if first.Data.EntryType != "link" || first.Data.Path == "" || first.Data.Scene == "" || first.Data.UrlLink == "" {
		t.Fatalf("expected link entry payload: %s", string(firstBody))
	}
	if first.Data.UrlLink != first.Data.Path || first.Data.UrlLinkError == "" {
		t.Fatalf("expected internal path fallback with url link error: %s", string(firstBody))
	}
	if poster.Data.EntryType != "poster" || poster.Data.InviteCode == "" || poster.Data.Title != "杭州 · 周末城市探索" || poster.Data.Game.ID != 1 || poster.Data.Game.CurrentPlayers != 1 || poster.Data.Game.MinPlayers != 5 || poster.Data.Game.MaxPlayers != 8 {
		t.Fatalf("expected poster share card game payload: %s", string(posterBody))
	}
	if !strings.Contains(poster.Data.Path, "/pages/game/detail/index?id=1") || !strings.Contains(poster.Data.Path, "inviteCode="+poster.Data.InviteCode) || !strings.Contains(poster.Data.Scene, "gameId=1") {
		t.Fatalf("expected game detail poster path and scene: %s", string(posterBody))
	}
	postJSON(t, mux, "/api/app/invites/entries", token, `{"entryType":"search"}`, http.StatusUnprocessableEntity)
}

func TestInviteEntryBindsExistingWechatHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(gameInviteConfigKey, gameInviteConfigDTO{
		MinPlayerCount:      1,
		MaxPlayerCount:      2,
		BudgetMaxAmount:     6000,
		DefaultBudget:       "1200",
		DefaultTitle:        "数据库邀请测试标题",
		DefaultDetail:       "数据库邀请测试详情",
		PlayerIntroTemplate: "数据库模板-{expertName}",
		ActivityTypes:       []inviteActivityTypeDTO{{Key: "db", Name: "数据库类型"}},
		RewardRateConfig: inviteRewardRateConfigDTO{
			PlatformServiceRate:   8,
			SystemGuideRewardRate: 12,
			InviteRewardRate:      30,
		},
	}); err != nil {
		t.Fatalf("seed invite config failed: %v", err)
	}
	server.Register(mux)
	token := loginForTestWithCode(t, mux, "entry-existing-user")

	entryBody := postJSON(t, mux, "/api/app/invites/entries", token, `{"entryType":"link"}`, http.StatusOK)
	var entry struct {
		Data struct {
			InviteCode string `json:"inviteCode"`
			EntryType  string `json:"entryType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(entryBody, &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Data.InviteCode == "" || entry.Data.EntryType != "link" {
		t.Fatalf("expected link entry: %s", string(entryBody))
	}

	loginBody := postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"entry-existing-user","inviteCode":"`+entry.Data.InviteCode+`","entryType":"link"}`, http.StatusOK)
	var login struct {
		Data struct {
			AuthPageMode         string `json:"authPageMode"`
			BoundWechat          bool   `json:"boundWechat"`
			InviteBindingStatus  string `json:"inviteBindingStatus"`
			InviteBindingMessage string `json:"inviteBindingMessage"`
			InviteRelation       struct {
				InviteCodeID int64 `json:"inviteCodeId"`
			} `json:"inviteRelation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginBody, &login); err != nil {
		t.Fatal(err)
	}
	if login.Data.AuthPageMode != "login" || !login.Data.BoundWechat || login.Data.InviteBindingStatus != auth.InviteBindingStatusAlreadyBound || login.Data.InviteBindingMessage == "" || login.Data.InviteRelation.InviteCodeID == 0 {
		t.Fatalf("expected existing wechat login mode: %s", string(loginBody))
	}
	reopenBody := postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"entry-existing-user"}`, http.StatusOK)
	var reopen struct {
		Data struct {
			AuthPageMode   string `json:"authPageMode"`
			BoundWechat    bool   `json:"boundWechat"`
			InviteRelation struct {
				InviteeUserID int64 `json:"inviteeUserId"`
			} `json:"inviteRelation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reopenBody, &reopen); err != nil {
		t.Fatal(err)
	}
	if reopen.Data.AuthPageMode != "login" || !reopen.Data.BoundWechat || reopen.Data.InviteRelation.InviteeUserID == 0 {
		t.Fatalf("expected bound wechat to reopen without invite: %s", string(reopenBody))
	}

	precheckBody := postJSON(t, mux, "/api/app/invites/precheck", "", `{"inviteCode":"`+entry.Data.InviteCode+`","entryType":"link"}`, http.StatusOK)
	var precheck struct {
		Data struct {
			AuthPageMode string `json:"authPageMode"`
			BoundWechat  bool   `json:"boundWechat"`
		} `json:"data"`
	}
	if err := json.Unmarshal(precheckBody, &precheck); err != nil {
		t.Fatal(err)
	}
	if precheck.Data.AuthPageMode != "register" || precheck.Data.BoundWechat {
		t.Fatalf("expected scanned invite to remain unbound: %s", string(precheckBody))
	}
	otherLoginBody := postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"entry-other-user","inviteCode":"`+entry.Data.InviteCode+`","entryType":"link"}`, http.StatusOK)
	var otherLogin struct {
		Data struct {
			AuthPageMode string `json:"authPageMode"`
			BoundWechat  bool   `json:"boundWechat"`
		} `json:"data"`
	}
	if err := json.Unmarshal(otherLoginBody, &otherLogin); err != nil {
		t.Fatal(err)
	}
	if otherLogin.Data.AuthPageMode != "register" || otherLogin.Data.BoundWechat {
		t.Fatalf("expected another wechat to bind the still-unbound invite: %s", string(otherLoginBody))
	}
	postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"entry-third-user","inviteCode":"`+entry.Data.InviteCode+`","entryType":"link"}`, http.StatusForbidden)
}

func TestAdminUserInviteRelationHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	inviterToken := loginForTestWithCode(t, mux, "invite-rel-inviter")
	completeIdentityForTest(t, mux, inviterToken)
	inviterID := currentUserIDForTest(t, mux, inviterToken)

	inviteeToken := loginForTestWithCode(t, mux, "invite-rel-invitee")
	completeIdentityForTest(t, mux, inviteeToken)
	inviteeID := currentUserIDForTest(t, mux, inviteeToken)

	_, err := authService.AdminCreateInviteCode("", inviterID, 1, "link")
	if err != nil {
		t.Fatal(err)
	}
	_, err = authService.SetInviteRelationInviter(inviteeID, inviterID, "test")
	if err != nil {
		t.Fatal(err)
	}

	adminToken := adminLoginForTest(t, mux)
	listBody := getAdminJSON(t, mux, "/api/admin/users", adminToken, http.StatusOK)
	var listResp struct {
		Data struct {
			Items []struct {
				ID            int64  `json:"id"`
				InviterUserID int64  `json:"inviterUserId"`
				InviteCodeID  int64  `json:"inviteCodeId"`
				BindSource    string `json:"bindSource"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listBody, &listResp); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range listResp.Data.Items {
		if item.ID == inviteeID {
			found = true
			if item.InviterUserID != inviterID {
				t.Fatalf("expected inviter user id %d, got %d", inviterID, item.InviterUserID)
			}
			if item.InviteCodeID == 0 || item.BindSource == "" {
				t.Fatal("expected invite relation metadata in list")
			}
		}
	}
	if !found {
		t.Fatalf("expected invitee in admin user list: %s", string(listBody))
	}

	newInviterToken := loginForTestWithCode(t, mux, "invite-rel-new-inviter")
	completeIdentityForTest(t, mux, newInviterToken)
	newInviterID := currentUserIDForTest(t, mux, newInviterToken)
	updateBody := putJSON(t, mux, "/api/admin/users/"+strconv.FormatInt(inviteeID, 10)+"/invite-relation", adminToken, `{"inviterUserId":`+strconv.FormatInt(newInviterID, 10)+`}`, http.StatusOK)
	var updateResp struct {
		Data struct {
			InviteRelation struct {
				InviterUserID int64 `json:"inviterUserId"`
			} `json:"inviteRelation"`
			Inviter struct {
				ID int64 `json:"id"`
			} `json:"inviter"`
		} `json:"data"`
	}
	if err := json.Unmarshal(updateBody, &updateResp); err != nil {
		t.Fatal(err)
	}
	if updateResp.Data.InviteRelation.InviterUserID != newInviterID || updateResp.Data.Inviter.ID != newInviterID {
		t.Fatalf("expected updated inviter in response: %s", string(updateBody))
	}
}

func TestReplayContextReadsQuickActionsFromSystemConfigHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(gameReplayQuickActionsConfigKey, []replayQuickActionDTO{
		{ID: "create-new", Theme: "pink", IconType: "plus", Title: "数据库创建新局", Desc: "数据库描述", Route: "create", Order: 20, Visible: true},
		{ID: "same-friends", Theme: "green", IconText: "A", Title: "数据库再来一局", Desc: "数据库好友", Route: "confirm", Order: 10, Visible: true},
		{ID: "hidden", Theme: "blue", Title: "隐藏项", Desc: "不展示", Route: "create", Order: 30, Visible: false},
	}); err != nil {
		t.Fatalf("seed replay quick actions failed: %v", err)
	}
	if err := server.systemConfig.Set(gameReplayQuickMessagesConfigKey, []replayQuickMessageDTO{
		{Text: "数据库第一句", Order: 20, Visible: true},
		{Text: "数据库置顶句", Order: 10, Visible: true},
		{Text: "隐藏句", Order: 30, Visible: false},
	}); err != nil {
		t.Fatalf("seed replay quick messages failed: %v", err)
	}
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "replay-quick-actions")
	completeIdentityForTest(t, mux, token)
	createBody := postJSON(t, mux, "/api/app/games", token, `{"title":"再玩一局配置测试","cityName":"杭州","locationName":"西湖","address":"西湖","longitude":120.1,"latitude":30.2,"startTime":"2026-07-01T10:00:00Z","endTime":"2026-07-01T12:00:00Z","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var createResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createBody, &createResp); err != nil {
		t.Fatal(err)
	}

	body := getJSON(t, mux, "/api/app/game-invites/replay-context?sourceGameId="+strconv.FormatInt(createResp.Data.ID, 10), token, http.StatusOK)
	var resp struct {
		Data struct {
			QuickActions []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				Route string `json:"route"`
			} `json:"quickActions"`
			QuickMessages []string `json:"quickMessages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.QuickActions) != 2 || resp.Data.QuickActions[0].ID != "same-friends" || resp.Data.QuickActions[0].Title != "数据库再来一局" || resp.Data.QuickActions[1].Route != "create" {
		t.Fatalf("expected replay quick actions from system config: %s", string(body))
	}
	if len(resp.Data.QuickMessages) != 2 || resp.Data.QuickMessages[0] != "数据库置顶句" {
		t.Fatalf("expected replay quick messages from system config: %s", string(body))
	}
}

func TestReviewCompleteConfigFeedsAppAndAdminHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	userToken := loginForTestWithCode(t, mux, "review-complete-config-user")
	body := getJSON(t, mux, "/api/app/reviews/complete-config", userToken, http.StatusOK)
	var appResp struct {
		Data struct {
			Benefits []struct {
				Title string `json:"title"`
			} `json:"benefits"`
			PlayOptions []struct {
				ID     string `json:"id"`
				Intent string `json:"intent"`
			} `json:"playOptions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &appResp); err != nil {
		t.Fatal(err)
	}
	if len(appResp.Data.Benefits) == 0 || len(appResp.Data.PlayOptions) == 0 || appResp.Data.PlayOptions[0].Intent == "" {
		t.Fatalf("expected default review complete config: %s", string(body))
	}

	adminToken := adminLoginForTest(t, mux)
	putJSON(t, mux, "/api/admin/reviews/complete-config", adminToken, `{"benefits":[{"iconText":"R","theme":"blue","title":"数据库权益","desc":"数据库描述","order":10,"visible":true}],"playOptions":[{"id":"again","theme":"green","title":"数据库再玩","desc":"数据库选项","intent":"yes","route":"play_again","order":10,"visible":true}],"version":"db-test"}`, http.StatusOK)
	body = getJSON(t, mux, "/api/app/reviews/complete-config", userToken, http.StatusOK)
	if err := json.Unmarshal(body, &appResp); err != nil {
		t.Fatal(err)
	}
	if len(appResp.Data.Benefits) != 1 || appResp.Data.Benefits[0].Title != "数据库权益" || appResp.Data.PlayOptions[0].ID != "again" {
		t.Fatalf("expected updated review complete config: %s", string(body))
	}
}

func TestAdminCreditDeductionRulesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(growthAchievementConfigKey, growthAchievementConfigDTO{
		Filters: []growthAchievementFilterDTO{
			{Key: "all", Label: "全部", Order: 10},
			{Key: "city", Label: "点亮城市", Order: 20},
		},
		Catalog: []growthAchievementItemDTO{
			{ID: "first_review", Code: "first_review", Title: "首次评价", Desc: "完成第一次服务评价", Category: "city", Order: 10, Visible: true},
		},
		Locked: []growthAchievementItemDTO{
			{ID: "credit_keeper", Code: "credit_keeper", Title: "信用守护", Desc: "信用分保持 80 分以上", Category: "city", ProgressPercent: 80, Order: 10, Visible: true},
		},
		Season:           growthAchievementSeasonDTO{Title: "成长赛季", Status: "进行中", RemainTpl: "已沉淀 {footprintCount} 条足迹"},
		OnlineSuffix:     "人成长中",
		LevelTitlePrefix: "探索行家 Lv.",
		Version:          "test",
	}); err != nil {
		t.Fatal(err)
	}
	server.Register(mux)

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	getAdminJSON(t, mux, "/api/admin/credit-deduction-rules", operatorToken, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	body := getAdminJSON(t, mux, "/api/admin/credit-deduction-rules", adminToken, http.StatusOK)
	var resp struct {
		Data struct {
			Items []struct {
				RuleCode    string `json:"ruleCode"`
				ChangeValue int    `json:"changeValue"`
				Enabled     bool   `json:"enabled"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.Items) == 0 || resp.Data.Items[0].ChangeValue >= 0 {
		t.Fatalf("expected default credit deduction rules: %s", string(body))
	}

	putJSON(t, mux, "/api/admin/credit-deduction-rules", adminToken, `{"items":[{"ruleCode":"low_review","changeValue":5,"enabled":true}]}`, http.StatusUnprocessableEntity)
	body = putJSON(t, mux, "/api/admin/credit-deduction-rules", adminToken, `{"items":[{"ruleCode":"low_review","changeValue":-8,"enabled":true,"description":"low review test"}]}`, http.StatusOK)
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.Items) != 1 || resp.Data.Items[0].RuleCode != "low_review" || resp.Data.Items[0].ChangeValue != -8 || !resp.Data.Items[0].Enabled {
		t.Fatalf("expected saved credit deduction rule: %s", string(body))
	}
}

func TestCreditAppealFromCreditLogHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "credit-appeal-user")
	completeIdentityForTest(t, mux, token)
	otherToken := loginForTestWithCode(t, mux, "credit-appeal-other")
	completeIdentityForTest(t, mux, otherToken)
	userID := currentUserIDForTest(t, mux, token)
	credit := server.reviews.DeductCredit(userID, 12, "low_review")
	appealFileBody := postJSON(t, mux, "/api/app/files/upload-token", token, `{"bizType":"report_attachment","objectId":`+strconv.FormatInt(credit.ID, 10)+`,"fileName":"credit-appeal.pdf","mimeType":"application/pdf","size":128}`, http.StatusOK)
	var appealFileResp struct {
		Data struct {
			File struct {
				ID int64 `json:"fileId"`
			} `json:"file"`
		} `json:"data"`
	}
	if err := json.Unmarshal(appealFileBody, &appealFileResp); err != nil {
		t.Fatal(err)
	}

	postJSON(t, mux, "/api/app/profile/credit-appeals", otherToken, `{"creditLogId":`+strconv.FormatInt(credit.ID, 10)+`,"content":"not mine"}`, http.StatusNotFound)
	body := postJSON(t, mux, "/api/app/profile/credit-appeals", token, `{"creditLogId":`+strconv.FormatInt(credit.ID, 10)+`,"content":"有特殊情况","fileIds":[`+strconv.FormatInt(appealFileResp.Data.File.ID, 10)+`]}`, http.StatusOK)
	var resp struct {
		Data struct {
			ID            int64   `json:"id"`
			ReportType    string  `json:"reportType"`
			Status        string  `json:"status"`
			CreditLogID   int64   `json:"creditLogId"`
			AppealFileIDs []int64 `json:"appealFileIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ID == 0 || resp.Data.ReportType != "credit_appeal" || resp.Data.Status != "appealed" || resp.Data.CreditLogID != credit.ID || len(resp.Data.AppealFileIDs) != 1 || resp.Data.AppealFileIDs[0] != appealFileResp.Data.File.ID {
		t.Fatalf("expected credit appeal report: %s", string(body))
	}
	adminToken := adminLoginForTest(t, mux)
	appealsBody := getJSON(t, mux, "/api/app/reports/appeals/my", token, http.StatusOK)
	if !strings.Contains(string(appealsBody), "credit_appeal") {
		t.Fatalf("expected credit appeal in my appeals: %s", string(appealsBody))
	}

	handleBody := postAdminJSON(t, mux, "/api/admin/reports/"+strconv.FormatInt(resp.Data.ID, 10)+"/handle", adminToken, `{"adminId":99,"result":"appeal approved","outcome":"appeal_approved"}`, http.StatusOK)
	var handled struct {
		Data struct {
			Status             string `json:"status"`
			HandleOutcome      string `json:"handleOutcome"`
			CreditChange       int    `json:"creditChange"`
			CreditTargetUserID int64  `json:"creditTargetUserId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(handleBody, &handled); err != nil {
		t.Fatal(err)
	}
	if handled.Data.Status != "handled" || handled.Data.HandleOutcome != "appeal_approved" || handled.Data.CreditChange != 10 || handled.Data.CreditTargetUserID != userID {
		t.Fatalf("expected approved credit appeal to restore credit: %s", string(handleBody))
	}
	creditBody := getJSON(t, mux, "/api/app/profile/credit-center", token, http.StatusOK)
	var creditResp struct {
		Data struct {
			Score   int `json:"score"`
			Records []struct {
				Reason string `json:"reason"`
				Score  string `json:"score"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(creditBody, &creditResp); err != nil {
		t.Fatal(err)
	}
	foundAppealPassed := false
	for _, record := range creditResp.Data.Records {
		if record.Reason == "appeal_passed" && record.Score == "+10" {
			foundAppealPassed = true
			break
		}
	}
	if creditResp.Data.Score != 100 || !foundAppealPassed {
		t.Fatalf("expected restored credit center record: %s", string(creditBody))
	}
	appealsAfterBody := getJSON(t, mux, "/api/app/reports/appeals/my", token, http.StatusOK)
	if !strings.Contains(string(appealsAfterBody), "credit_appeal") || !strings.Contains(string(appealsAfterBody), "appeal_approved") {
		t.Fatalf("expected handled credit appeal retained in appeal history: %s", string(appealsAfterBody))
	}
}

func TestAdminInviteCodeManagementHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	getAdminJSON(t, mux, "/api/admin/invite-codes", operatorToken, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	ownerToken := loginForTestWithCode(t, mux, "admin-invite-owner")
	ownerUserID := currentUserIDForTest(t, mux, ownerToken)
	ownerUserIDJSON := strconv.FormatInt(ownerUserID, 10)
	postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"entryType":"qrcode"}`, http.StatusUnprocessableEntity)
	postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"ownerUserId":999999,"entryType":"qrcode"}`, http.StatusNotFound)
	body := postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"code":"ADMINQR001","ownerUserId":`+ownerUserIDJSON+`,"maxUses":99,"entryType":"qrcode"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID        int64  `json:"id"`
			Code      string `json:"code"`
			OwnerID   int64  `json:"ownerUserId"`
			MaxUses   int    `json:"maxUses"`
			EntryType string `json:"entryType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == 0 || created.Data.OwnerID != ownerUserID || created.Data.Code != "ADMINQR001" || created.Data.MaxUses != 1 || created.Data.EntryType != "qrcode" {
		t.Fatalf("expected created qrcode invite code: %s", string(body))
	}
	postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"code":"ADMINQR001","ownerUserId":`+ownerUserIDJSON+`,"entryType":"qrcode"}`, http.StatusUnprocessableEntity)
	postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"code":"邀请码中文","ownerUserId":`+ownerUserIDJSON+`,"entryType":"qrcode"}`, http.StatusUnprocessableEntity)
	materialsBody := getAdminJSON(t, mux, "/api/admin/invite-codes/ADMINQR001/materials", adminToken, http.StatusOK)
	var materialsResp struct {
		Data struct {
			InviteCode string `json:"inviteCode"`
			EntryType  string `json:"entryType"`
			Query      string `json:"query"`
			Scene      string `json:"scene"`
		} `json:"data"`
	}
	if err := json.Unmarshal(materialsBody, &materialsResp); err != nil {
		t.Fatal(err)
	}
	if materialsResp.Data.InviteCode != "ADMINQR001" || materialsResp.Data.EntryType != "qrcode" || materialsResp.Data.Scene != "ADMINQR001" || !strings.Contains(materialsResp.Data.Query, "entryType=qrcode") {
		t.Fatalf("expected invite material scene to be raw invite code: %s", string(materialsBody))
	}
	batchBody := postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"ownerUserId":`+ownerUserIDJSON+`,"maxUses":99,"entryType":"link","batchCount":3}`, http.StatusOK)
	var batchResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				Code      string `json:"code"`
				EntryType string `json:"entryType"`
				MaxUses   int    `json:"maxUses"`
				OwnerID   int64  `json:"ownerUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(batchBody, &batchResp); err != nil {
		t.Fatal(err)
	}
	seenBatchCodes := map[string]bool{}
	if batchResp.Data.Total != 3 || len(batchResp.Data.Items) != 3 {
		t.Fatalf("expected 3 batched invite codes: %s", string(batchBody))
	}
	for _, item := range batchResp.Data.Items {
		if item.Code == "" || item.EntryType != "link" || item.MaxUses != 1 || item.OwnerID != ownerUserID || seenBatchCodes[item.Code] {
			t.Fatalf("expected unique link invite batch item: %#v in %s", item, string(batchBody))
		}
		seenBatchCodes[item.Code] = true
	}
	postAdminJSON(t, mux, "/api/admin/invite-codes", adminToken, `{"code":"MANUAL-BATCH","entryType":"poster","batchCount":2}`, http.StatusUnprocessableEntity)

	listBody := getAdminJSON(t, mux, "/api/admin/invite-codes?entryType=qrcode&status=active", adminToken, http.StatusOK)
	var listResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				Code      string `json:"code"`
				EntryType string `json:"entryType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listBody, &listResp); err != nil {
		t.Fatal(err)
	}
	if listResp.Data.Total != 1 || listResp.Data.Items[0].Code != "ADMINQR001" {
		t.Fatalf("expected admin invite code list: %s", string(listBody))
	}

	postJSON(t, mux, "/api/app/auth/wechat-login", "", `{"code":"admin-invite-user","inviteCode":"ADMINQR001","entryType":"qrcode"}`, http.StatusOK)
	relationsBody := getAdminJSON(t, mux, "/api/admin/invite-relations?inviterUserId="+strconv.FormatInt(created.Data.OwnerID, 10), adminToken, http.StatusOK)
	var relationsResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				InviteCodeID  int64  `json:"inviteCodeId"`
				InviterUserID int64  `json:"inviterUserId"`
				InviteeUserID int64  `json:"inviteeUserId"`
				BindSource    string `json:"bindSource"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(relationsBody, &relationsResp); err != nil {
		t.Fatal(err)
	}
	if relationsResp.Data.Total != 1 || relationsResp.Data.Items[0].InviteCodeID != created.Data.ID || relationsResp.Data.Items[0].InviterUserID != created.Data.OwnerID || relationsResp.Data.Items[0].BindSource != "invite_qrcode" {
		t.Fatalf("expected invite relation list: %s", string(relationsBody))
	}

	detailBody := getAdminJSON(t, mux, "/api/admin/invite-codes/ADMINQR001", adminToken, http.StatusOK)
	var detailResp struct {
		Data struct {
			InviteCode struct {
				Code              string `json:"code"`
				BoundWechatUserID int64  `json:"boundWechatUserId"`
			} `json:"inviteCode"`
			Relations []struct {
				InviteCodeID int64 `json:"inviteCodeId"`
			} `json:"relations"`
			BoundUser struct {
				ID int64 `json:"id"`
			} `json:"boundUser"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.InviteCode.Code != "ADMINQR001" || detailResp.Data.InviteCode.BoundWechatUserID == 0 || len(detailResp.Data.Relations) != 1 || detailResp.Data.BoundUser.ID == 0 {
		t.Fatalf("expected invite code detail with bound user and relation: %s", string(detailBody))
	}

	postAdminJSON(t, mux, "/api/admin/invite-codes/ADMINQR001/disable", adminToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/invites/precheck", "", `{"inviteCode":"ADMINQR001","entryType":"qrcode"}`, http.StatusForbidden)
}

func TestWechatLoginAndCurrentUserHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.profiles.GrantRole(1, "expert")
	if _, _, err := server.points.Grant(1, 12, "test_seed", 0, "seed points"); err != nil {
		t.Fatal(err)
	}
	server.Register(mux)

	body := bytes.NewBufferString(`{"code":"u1","inviteCode":"TEST2026"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/app/auth/wechat-login", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var login struct {
		Code int `json:"code"`
		Data struct {
			Token                   string `json:"token"`
			PreAuthToken            string `json:"preAuthToken"`
			RequiresIdentityBinding bool   `json:"requiresIdentityBinding"`
			IdentityBindStatus      string `json:"identityBindStatus"`
			User                    struct {
				ID            int64             `json:"id"`
				Roles         []string          `json:"roles"`
				RoleStatusMap map[string]string `json:"roleStatusMap"`
				InviteCode    string            `json:"inviteCode"`
				Membership    struct {
					Status   string `json:"status"`
					PlanCode string `json:"planCode"`
					PlanName string `json:"planName"`
				} `json:"membership"`
				Growth struct {
					Level            int `json:"level"`
					ExperienceValue  int `json:"experienceValue"`
					CreditScore      int `json:"creditScore"`
					TodayCreditScore int `json:"todayCreditScore"`
					Points           int `json:"points"`
				} `json:"growth"`
				PointsSummary struct {
					AvailablePoints int `json:"availablePoints"`
				} `json:"pointsSummary"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	if login.Data.Token != "" || login.Data.PreAuthToken == "" {
		t.Fatalf("expected pre-auth token only, got %s", rec.Body.String())
	}
	if !login.Data.RequiresIdentityBinding || login.Data.IdentityBindStatus != "wechat_logged_in" {
		t.Fatalf("expected identity binding hint: %s", rec.Body.String())
	}
	if login.Data.User.ID != 1 || login.Data.User.InviteCode == "" || login.Data.User.Membership.Status != "none" {
		t.Fatalf("expected login CurrentUserDTO basics: %s", rec.Body.String())
	}
	if len(login.Data.User.Roles) != 2 || login.Data.User.RoleStatusMap["expert"] != "approved" || login.Data.User.RoleStatusMap["guide"] != "none" {
		t.Fatalf("expected role snapshot in login CurrentUserDTO: %s", rec.Body.String())
	}
	if login.Data.User.Growth.Level != 1 || login.Data.User.Growth.CreditScore != 100 || login.Data.User.Growth.TodayCreditScore != 100 || login.Data.User.Growth.Points != 12 || login.Data.User.PointsSummary.AvailablePoints != 12 {
		t.Fatalf("expected growth and points in login CurrentUserDTO: %s", rec.Body.String())
	}
	if login.Data.User.Membership.Status != "none" || login.Data.User.Membership.PlanCode != "none" {
		t.Fatalf("expected default membership state: %s", rec.Body.String())
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/app/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+login.Data.PreAuthToken)
	meRec := httptest.NewRecorder()
	mux.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", meRec.Code, meRec.Body.String())
	}
	membershipBody := getJSON(t, mux, "/api/app/membership/my", login.Data.PreAuthToken, http.StatusOK)
	var membershipResp struct {
		Data struct {
			Status   string `json:"status"`
			PlanCode string `json:"planCode"`
			PlanName string `json:"planName"`
		} `json:"data"`
	}
	if err := json.Unmarshal(membershipBody, &membershipResp); err != nil {
		t.Fatal(err)
	}
	if membershipResp.Data.Status != "none" || membershipResp.Data.PlanCode != "none" {
		t.Fatalf("expected default membership response: %s", string(membershipBody))
	}
	plansBody := getJSON(t, mux, "/api/app/membership/plans", login.Data.PreAuthToken, http.StatusOK)
	if !bytes.Contains(plansBody, []byte("basic")) || !bytes.Contains(plansBody, []byte("pro")) {
		t.Fatalf("expected membership plans: %s", string(plansBody))
	}
	radarBody := getJSON(t, mux, "/api/app/membership/radar-config", login.Data.PreAuthToken, http.StatusOK)
	if !bytes.Contains(radarBody, []byte(`"radarNodes"`)) || !bytes.Contains(radarBody, []byte(`"formRows"`)) {
		t.Fatalf("expected membership radar config: %s", string(radarBody))
	}

	postJSON(t, mux, "/api/app/files/upload-token", login.Data.PreAuthToken, `{"bizType":"chat_file","objectId":1,"fileName":"a.png","mimeType":"image/png","size":128}`, http.StatusForbidden)
	postJSON(t, mux, "/api/app/files/upload-token", login.Data.PreAuthToken, `{"bizType":"realname_material","fileName":"id-card.png","mimeType":"image/png","size":128}`, http.StatusOK)
	postJSON(t, mux, "/api/app/files/upload-token", login.Data.PreAuthToken, `{"bizType":"report_attachment","objectId":1,"fileName":"report.pdf","mimeType":"application/pdf","size":128}`, http.StatusOK)
	postJSON(t, mux, "/api/app/files/upload-token", login.Data.PreAuthToken, `{"bizType":"game_application","objectId":1,"fileName":"case.pdf","mimeType":"application/pdf","size":128}`, http.StatusOK)

	uploadBody := postJSON(t, mux, "/api/app/files/upload-token", login.Data.PreAuthToken, `{"bizType":"avatar","fileName":"avatar.png","mimeType":"image/png","size":128}`, http.StatusOK)
	var uploadResp struct {
		Data struct {
			File struct {
				ID int64 `json:"fileId"`
			} `json:"file"`
		} `json:"data"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResp); err != nil {
		t.Fatal(err)
	}
	if uploadResp.Data.File.ID == 0 {
		t.Fatalf("expected avatar file id: %s", string(uploadBody))
	}

	putJSON(t, mux, "/api/app/users/me/profile", login.Data.PreAuthToken, `{"nickname":"Alice","avatarFileId":`+strconv.FormatInt(uploadResp.Data.File.ID, 10)+`}`, http.StatusUnprocessableEntity)
	updateBody := putJSON(t, mux, "/api/app/users/me/profile", login.Data.PreAuthToken, `{"nickname":"Alice"}`, http.StatusOK)
	var updated struct {
		Data struct {
			Nickname     string `json:"nickname"`
			AvatarURL    string `json:"avatarUrl"`
			AvatarFileID int64  `json:"avatarFileId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(updateBody, &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Data.Nickname != "Alice" || updated.Data.AvatarURL != "" || updated.Data.AvatarFileID != 0 {
		t.Fatalf("expected updated profile: %s", string(updateBody))
	}

	otherToken := loginForTestWithCode(t, mux, "avatar-other")
	completeIdentityForTest(t, mux, login.Data.PreAuthToken)
	formalToken := issueFormalTokenForTest(t, mux, login.Data.PreAuthToken)
	putJSON(t, mux, "/api/app/users/me/profile", otherToken, `{"nickname":"Bob","avatarFileId":1}`, http.StatusUnprocessableEntity)
	avatarDownloadPath := "/api/app/files/" + strconv.FormatInt(uploadResp.Data.File.ID, 10) + "/download-url"
	getJSON(t, mux, avatarDownloadPath, formalToken, http.StatusOK)
	getJSON(t, mux, avatarDownloadPath, otherToken, http.StatusForbidden)

	meReq = httptest.NewRequest(http.MethodGet, "/api/app/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+login.Data.PreAuthToken)
	meRec = httptest.NewRecorder()
	mux.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("expected 200 after profile update, got %d: %s", meRec.Code, meRec.Body.String())
	}
	var meResp struct {
		Data struct {
			Nickname      string            `json:"nickname"`
			Roles         []string          `json:"roles"`
			RoleStatusMap map[string]string `json:"roleStatusMap"`
			InviteCode    string            `json:"inviteCode"`
			Growth        struct {
				Points int `json:"points"`
			} `json:"growth"`
			Identity struct {
				Status string `json:"status"`
			} `json:"identity"`
		} `json:"data"`
	}
	if err := json.Unmarshal(meRec.Body.Bytes(), &meResp); err != nil {
		t.Fatal(err)
	}
	if meResp.Data.Nickname != "Alice" || meResp.Data.InviteCode == "" || meResp.Data.Growth.Points != 12 {
		t.Fatalf("expected current user aggregate updated: %s", meRec.Body.String())
	}
	if len(meResp.Data.Roles) != 2 || meResp.Data.RoleStatusMap["expert"] != "approved" || meResp.Data.Identity.Status != "verified" {
		t.Fatalf("expected current user role and identity aggregate: %s", meRec.Body.String())
	}
	if _, err := server.membership.Grant(1, "basic", nil); err != nil {
		t.Fatal(err)
	}
	memberBody := getJSON(t, mux, "/api/app/membership/my", login.Data.PreAuthToken, http.StatusOK)
	if !bytes.Contains(memberBody, []byte(`"status":"active"`)) || !bytes.Contains(memberBody, []byte(`"planCode":"basic"`)) {
		t.Fatalf("expected active membership after grant: %s", string(memberBody))
	}
	meReq = httptest.NewRequest(http.MethodGet, "/api/app/users/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+login.Data.PreAuthToken)
	meRec = httptest.NewRecorder()
	mux.ServeHTTP(meRec, meReq)
	if !bytes.Contains(meRec.Body.Bytes(), []byte(`"membership"`)) || !bytes.Contains(meRec.Body.Bytes(), []byte(`"planCode":"basic"`)) {
		t.Fatalf("expected current user membership aggregate after grant: %s", meRec.Body.String())
	}
}

func TestMembershipRadarActionsRecordStateAndConnections(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "radar-actions")
	postJSON(t, mux, "/api/app/membership/radar/actions", token, `{"action":"unknown"}`, http.StatusUnprocessableEntity)

	saveBody := postJSON(t, mux, "/api/app/membership/radar/actions", token, `{"action":"save","matchMode":"criteria","matchCriteria":{"industry":"TMT"},"formRows":[{"key":"industry","value":"TMT"}]}`, http.StatusOK)
	if !bytes.Contains(saveBody, []byte(`"industry"`)) {
		t.Fatalf("expected saved radar profile: %s", string(saveBody))
	}

	followBody := postJSON(t, mux, "/api/app/membership/radar/actions", token, `{"action":"follow","targetId":"lu-yi","targetUserId":201}`, http.StatusOK)
	if !bytes.Contains(followBody, []byte(`"status":"ok"`)) {
		t.Fatalf("expected follow action ok: %s", string(followBody))
	}
	if len(server.connections.My(1)) == 0 {
		t.Fatal("expected radar follow to create a connection")
	}

	profileBody := postJSON(t, mux, "/api/app/membership/radar/actions", token, `{"action":"profile","targetId":"lu-yi","targetUserId":201}`, http.StatusOK)
	if !bytes.Contains(profileBody, []byte(`/pages/profile/service-center/invite/member-detail/index?id=201`)) {
		t.Fatalf("expected profile route: %s", string(profileBody))
	}
	if !hasBehaviorEvent(server.audit.BehaviorLogs(), "membership_radar_follow", 201) {
		t.Fatalf("expected radar follow behavior log: %#v", server.audit.BehaviorLogs())
	}
}

func TestUpdateProfileUsesConfiguredAvatarDownloadURL(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{Storage: config.StorageConfig{
		UploadBaseURL:   "https://upload.example.com/private",
		DownloadBaseURL: "https://download.example.com/private",
	}})
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "avatar-public-url")
	uploadBody := postJSON(t, mux, "/api/app/files/upload-token", token, `{"bizType":"avatar","fileName":"my avatar.png","mimeType":"image/png","size":128}`, http.StatusOK)
	var uploadResp struct {
		Data struct {
			File struct {
				ID int64 `json:"fileId"`
			} `json:"file"`
		} `json:"data"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResp); err != nil {
		t.Fatal(err)
	}

	updateBody := putJSON(t, mux, "/api/app/profile/system-management/profile-info", token, `{"personalInfo":{"name":"Avatar","avatarFileId":`+strconv.FormatInt(uploadResp.Data.File.ID, 10)+`}}`, http.StatusOK)
	var updated struct {
		Data struct {
			PersonalInfo struct {
				AvatarURL           string `json:"avatarUrl"`
				PendingAvatarURL    string `json:"pendingAvatarUrl"`
				AvatarAuditStatus   string `json:"avatarAuditStatus"`
				PendingAvatarFileID int64  `json:"pendingAvatarFileId"`
			} `json:"personalInfo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(updateBody, &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Data.PersonalInfo.AvatarURL != "" || updated.Data.PersonalInfo.PendingAvatarFileID != uploadResp.Data.File.ID || updated.Data.PersonalInfo.AvatarAuditStatus != "pending" || updated.Data.PersonalInfo.PendingAvatarURL != "https://download.example.com/private/avatar/0/1-my%20avatar.png" {
		t.Fatalf("expected configured pending avatar download url, got %+v body=%s", updated.Data.PersonalInfo, string(updateBody))
	}
}

func TestProductionFileUploadRequiresConfiguredStorageURLs(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{AppEnv: "production"})
	server.Register(mux)

	token := loginForTest(t, mux)
	completeIdentityForTest(t, mux, token)
	postJSON(t, mux, "/api/app/files/upload-token", token, `{"bizType":"avatar","fileName":"avatar.png","mimeType":"image/png","size":128}`, http.StatusServiceUnavailable)
}

func TestProductionFileUploadUsesConfiguredStorageURLs(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{
		AppEnv: "production",
		Storage: config.StorageConfig{
			UploadBaseURL:   "https://upload.example.com/private",
			DownloadBaseURL: "https://download.example.com/private",
		},
	})
	server.Register(mux)

	token := loginForTest(t, mux)
	completeIdentityForTest(t, mux, token)
	body := postJSON(t, mux, "/api/app/files/upload-token", token, `{"bizType":"avatar","fileName":"avatar.png","mimeType":"image/png","size":128}`, http.StatusOK)
	var resp struct {
		Data struct {
			Upload struct {
				UploadURL string `json:"uploadUrl"`
			} `json:"upload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Upload.UploadURL == "" || strings.HasPrefix(resp.Data.Upload.UploadURL, "mock://") {
		t.Fatalf("expected configured production upload url: %s", string(body))
	}
}

func TestIdentityFlowHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{LBS: config.LBSConfig{DefaultRadiusMeter: 1}})
	server.Register(mux)
	token := loginForTest(t, mux)

	postJSON(t, mux, "/api/app/identity/phone/bind", token, `{"phone":"13800138000"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/send-code", token, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/send-code", token, `{}`, http.StatusTooManyRequests)
	postJSON(t, mux, "/api/app/sms/verify-code", token, `{"code":"000000"}`, http.StatusOK)
	body := postJSON(t, mux, "/api/app/identity/phone/verify", token, `{"realName":"User","idCard":"110101199001011234"}`, http.StatusOK)
	var submitResp struct {
		Data struct {
			UserID int64 `json:"userId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &submitResp); err != nil {
		t.Fatal(err)
	}
	if submitResp.Data.UserID == 0 {
		t.Fatalf("expected manual realname user id: %s", string(body))
	}
	adminToken := adminLoginForTest(t, mux)
	postAdminJSON(t, mux, "/api/admin/identity-verifications/"+strconv.FormatInt(submitResp.Data.UserID, 10)+"/review", adminToken, `{"approve":true,"reason":"test"}`, http.StatusOK)

	faceResp := postJSON(t, mux, "/api/app/identity/faceid/detect-auth", token, `{}`, http.StatusOK)
	var face struct {
		Data struct {
			FaceToken string `json:"faceToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(faceResp, &face); err != nil {
		t.Fatal(err)
	}
	if face.Data.FaceToken == "" {
		t.Fatal("expected face token")
	}

	postJSON(t, mux, "/api/app/identity/faceid/callback", token, `{"faceToken":"`+face.Data.FaceToken+`"}`, http.StatusOK)

	statusReq := httptest.NewRequest(http.MethodGet, "/api/app/identity/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusRec := httptest.NewRecorder()
	mux.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", statusRec.Code, statusRec.Body.String())
	}

	issueResp := postJSON(t, mux, "/api/app/auth/issue-token-after-identity", token, `{}`, http.StatusOK)
	var issued struct {
		Data struct {
			Token              string `json:"token"`
			IdentityBindStatus string `json:"identityBindStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(issueResp, &issued); err != nil {
		t.Fatal(err)
	}
	if issued.Data.Token == "" || issued.Data.IdentityBindStatus != "verified" {
		t.Fatalf("expected formal token after identity: %s", string(issueResp))
	}

	user, ok := authService.CurrentUser(token)
	if !ok {
		t.Fatal("expected login user")
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/identity-verifications", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected admin identity list 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
	filterReq := httptest.NewRequest(http.MethodGet, "/api/admin/identity-verifications?status=verified&userId="+strconv.FormatInt(user.ID, 10), nil)
	filterReq.Header.Set("Authorization", "Bearer "+adminToken)
	filterRec := httptest.NewRecorder()
	mux.ServeHTTP(filterRec, filterReq)
	if filterRec.Code != http.StatusOK {
		t.Fatalf("expected filtered admin identity list 200, got %d: %s", filterRec.Code, filterRec.Body.String())
	}
	var filterResp struct {
		Data struct {
			Items []struct {
				UserID int64  `json:"userId"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(filterRec.Body.Bytes(), &filterResp); err != nil {
		t.Fatal(err)
	}
	if len(filterResp.Data.Items) != 1 || filterResp.Data.Items[0].UserID != user.ID || filterResp.Data.Items[0].Status != "verified" {
		t.Fatalf("expected filtered identity record: %s", filterRec.Body.String())
	}
	emptyFilterReq := httptest.NewRequest(http.MethodGet, "/api/admin/identity-verifications?status=phone_bound&userId="+strconv.FormatInt(user.ID, 10), nil)
	emptyFilterReq.Header.Set("Authorization", "Bearer "+adminToken)
	emptyFilterRec := httptest.NewRecorder()
	mux.ServeHTTP(emptyFilterRec, emptyFilterReq)
	if emptyFilterRec.Code != http.StatusOK {
		t.Fatalf("expected empty filtered admin identity list 200, got %d: %s", emptyFilterRec.Code, emptyFilterRec.Body.String())
	}
	var emptyFilterResp struct {
		Data struct {
			Items []struct {
				UserID int64 `json:"userId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(emptyFilterRec.Body.Bytes(), &emptyFilterResp); err != nil {
		t.Fatal(err)
	}
	if len(emptyFilterResp.Data.Items) != 0 {
		t.Fatalf("expected empty filtered identity list: %s", emptyFilterRec.Body.String())
	}
	detailReq := httptest.NewRequest(http.MethodGet, "/api/admin/identity-verifications/"+strconv.FormatInt(user.ID, 10), nil)
	detailReq.Header.Set("Authorization", "Bearer "+adminToken)
	detailRec := httptest.NewRecorder()
	mux.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected admin identity detail 200, got %d: %s", detailRec.Code, detailRec.Body.String())
	}
	var detailResp struct {
		Data struct {
			PhoneMasked    string `json:"phoneMasked"`
			PhoneFull      string `json:"phoneFull"`
			RealNameMasked string `json:"realNameMasked"`
			RealNameFull   string `json:"realNameFull"`
			IDCardMasked   string `json:"idCardMasked"`
			IDCardFull     string `json:"idCardFull"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.PhoneMasked != "138****8000" || detailResp.Data.RealNameMasked != "U***" || detailResp.Data.IDCardMasked != "110***********1234" {
		t.Fatalf("expected admin identity detail to keep masked fields: %s", detailRec.Body.String())
	}
	if detailResp.Data.PhoneFull != "13800138000" || detailResp.Data.RealNameFull != "User" || detailResp.Data.IDCardFull != "110101199001011234" {
		t.Fatalf("expected admin identity detail to expose full review fields: %s", detailRec.Body.String())
	}
}

func TestRestartRealnameResetsIdentityAndExposesPhoneFullHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "restart-realname")
	completeIdentityForTest(t, mux, token)
	userID := currentUserIDForTest(t, mux, token)
	if !identityService.IsVerified(userID) {
		t.Fatal("expected identity to start verified")
	}

	restartBody := postJSON(t, mux, "/api/app/identity/realname/restart", token, `{"phone":"13900139011"}`, http.StatusOK)
	var restartResp struct {
		Data struct {
			Status      string `json:"status"`
			PhoneMasked string `json:"phoneMasked"`
		} `json:"data"`
	}
	if err := json.Unmarshal(restartBody, &restartResp); err != nil {
		t.Fatal(err)
	}
	if restartResp.Data.Status != "phone_bound" || restartResp.Data.PhoneMasked != "139****9011" {
		t.Fatalf("expected phone_bound restart response: %s", string(restartBody))
	}
	if identityService.IsVerified(userID) {
		t.Fatal("expected restart realname to clear verified identity")
	}
	user, ok := authService.CurrentUser(token)
	if !ok || user.RealnameStatus != "phone_bound" || user.PhoneMasked != "139****9011" {
		t.Fatalf("expected user realname status and phone to be reset: %+v", user)
	}

	adminToken := adminLoginForTest(t, mux)
	detailBody := getAdminJSON(t, mux, "/api/admin/identity-verifications/"+strconv.FormatInt(userID, 10), adminToken, http.StatusOK)
	var detailResp struct {
		Data struct {
			Status      string `json:"status"`
			PhoneMasked string `json:"phoneMasked"`
			PhoneFull   string `json:"phoneFull"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.Status != "phone_bound" || detailResp.Data.PhoneMasked != "139****9011" || detailResp.Data.PhoneFull != "13900139011" {
		t.Fatalf("expected admin identity detail to expose restarted full phone: %s", string(detailBody))
	}

	submitBody := postJSON(t, mux, "/api/app/identity/phone/verify", token, `{"realName":"Restart User","idCard":"110101199001011235"}`, http.StatusOK)
	var submitResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(submitBody, &submitResp); err != nil {
		t.Fatal(err)
	}
	if submitResp.Data.Status != "pending" {
		t.Fatalf("expected manual review pending after re-submit: %s", string(submitBody))
	}
	user, ok = authService.CurrentUser(token)
	if !ok || user.RealnameStatus != "pending" {
		t.Fatalf("expected user realname status pending after re-submit: %+v", user)
	}
}

func TestAdminUsersListSupportsInviteLoggedUsers(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "admin-user-list")
	user, ok := authService.CurrentUser(token)
	if !ok {
		t.Fatal("expected login user")
	}
	completeIdentityForTest(t, mux, token)

	adminToken := adminLoginForTest(t, mux)
	body := getAdminJSON(t, mux, "/api/admin/users?keyword=mock_openid_admin-user-list&realnameStatus=verified&status=active", adminToken, http.StatusOK)
	var resp struct {
		Data struct {
			Total int          `json:"total"`
			Items []users.User `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Total != 1 || len(resp.Data.Items) != 1 {
		t.Fatalf("expected one admin user result: %s", string(body))
	}
	got := resp.Data.Items[0]
	if got.ID != user.ID || got.OpenID != user.OpenID || got.RealnameStatus != "verified" || got.Status != "active" {
		t.Fatalf("unexpected admin user payload: %+v body=%s", got, string(body))
	}
	idCardBody := getAdminJSON(t, mux, "/api/admin/users?keyword=110101199001011234&realnameStatus=verified&status=active", adminToken, http.StatusOK)
	var idCardResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(idCardBody, &idCardResp); err != nil {
		t.Fatal(err)
	}
	if idCardResp.Data.Total != 1 || len(idCardResp.Data.Items) != 1 || idCardResp.Data.Items[0].ID != user.ID {
		t.Fatalf("expected identity id card keyword to find user: %s", string(idCardBody))
	}
	realNameBody := getAdminJSON(t, mux, "/api/admin/users?keyword=User&realnameStatus=verified&status=active", adminToken, http.StatusOK)
	var realNameResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(realNameBody, &realNameResp); err != nil {
		t.Fatal(err)
	}
	if realNameResp.Data.Total != 1 || len(realNameResp.Data.Items) != 1 || realNameResp.Data.Items[0].ID != user.ID {
		t.Fatalf("expected identity real name keyword to find user: %s", string(realNameBody))
	}

	detailBody := getAdminJSON(t, mux, "/api/admin/users/"+strconv.FormatInt(user.ID, 10), adminToken, http.StatusOK)
	if !bytes.Contains(detailBody, []byte(`"user"`)) || !bytes.Contains(detailBody, []byte(`"identity"`)) {
		t.Fatalf("expected admin user detail with identity bundle: %s", string(detailBody))
	}
}

func TestAdminUserOptionsForGameCreatorSelector(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "game-creator-selector")
	user, ok := authService.CurrentUser(token)
	if !ok {
		t.Fatal("expected login user")
	}
	if _, err := authService.UpdateProfile(user.ID, "发起人昵称", "", 0); err != nil {
		t.Fatal(err)
	}

	gameManagerToken := adminLoginForTestAs(t, mux, "game_manager", "admin123")
	body := getAdminJSON(t, mux, "/api/admin/users/options?keyword=发起人昵称", gameManagerToken, http.StatusOK)
	var resp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				ID       int64  `json:"id"`
				Nickname string `json:"nickname"`
				Label    string `json:"label"`
				Status   string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Total != 1 || len(resp.Data.Items) != 1 {
		t.Fatalf("expected one user option: %s", string(body))
	}
	item := resp.Data.Items[0]
	if item.ID != user.ID || item.Nickname != "发起人昵称" || !strings.Contains(item.Label, "用户 "+strconv.FormatInt(user.ID, 10)) || item.Status != "active" {
		t.Fatalf("unexpected user option: %+v body=%s", item, string(body))
	}
	if bytes.Contains(body, []byte("mock_openid_game-creator-selector")) {
		t.Fatalf("user option leaked openid: %s", string(body))
	}

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	getAdminJSON(t, mux, "/api/admin/users/options", operatorToken, http.StatusForbidden)
}

func TestFaceIDResultAliasCompletesIdentity(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "face-result")

	postJSON(t, mux, "/api/app/identity/phone/bind", token, `{"phone":"13800138000"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/send-code", token, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/verify-code", token, `{"code":"000000"}`, http.StatusOK)
	body := postJSON(t, mux, "/api/app/identity/phone/verify", token, `{"realName":"User","idCard":"110101199001011234"}`, http.StatusOK)
	var submitResp struct {
		Data struct {
			UserID int64 `json:"userId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &submitResp); err != nil {
		t.Fatal(err)
	}
	if submitResp.Data.UserID == 0 {
		t.Fatalf("expected manual realname user id: %s", string(body))
	}
	adminToken := adminLoginForTest(t, mux)
	postAdminJSON(t, mux, "/api/admin/identity-verifications/"+strconv.FormatInt(submitResp.Data.UserID, 10)+"/review", adminToken, `{"approve":true,"reason":"test"}`, http.StatusOK)
	faceResp := postJSON(t, mux, "/api/app/identity/faceid/detect-auth", token, `{}`, http.StatusOK)
	var face struct {
		Data struct {
			FaceToken string `json:"faceToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(faceResp, &face); err != nil {
		t.Fatal(err)
	}
	resultBody := postJSON(t, mux, "/api/app/identity/faceid/result", token, `{"faceToken":"`+face.Data.FaceToken+`"}`, http.StatusOK)
	var result struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resultBody, &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Status != "verified" {
		t.Fatalf("expected verified via result alias: %s", string(resultBody))
	}
}

func TestFaceIDCallbackSignatureVerifier(t *testing.T) {
	server := &Server{}
	payload := []byte(`{"faceToken":"mock_face_token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/app/identity/faceid/callback", bytes.NewReader(payload))
	if !server.verifyFaceIDCallbackSignature(req, payload) {
		t.Fatal("expected unsigned callback to pass when verifier is disabled")
	}

	server.UseFaceIDCallbackVerifier("faceid-prod-32-random-bytes-value", true)
	if server.verifyFaceIDCallbackSignature(req, payload) {
		t.Fatal("expected missing signature to be rejected")
	}

	signedReq := httptest.NewRequest(http.MethodPost, "/api/app/identity/faceid/callback", bytes.NewReader(payload))
	signedReq.Header.Set("X-FaceID-Signature", "sha256="+faceIDCallbackTestSignature("faceid-prod-32-random-bytes-value", payload))
	if !server.verifyFaceIDCallbackSignature(signedReq, payload) {
		t.Fatal("expected valid signature to pass")
	}

	tamperedReq := httptest.NewRequest(http.MethodPost, "/api/app/identity/faceid/callback", bytes.NewReader(payload))
	tamperedReq.Header.Set("X-FaceID-Signature", faceIDCallbackTestSignature("faceid-prod-32-random-bytes-value", []byte(`{"faceToken":"other"}`)))
	if server.verifyFaceIDCallbackSignature(tamperedReq, payload) {
		t.Fatal("expected mismatched signature to be rejected")
	}
}

func TestPlayerCanCreateGameWithoutVerifiedIdentity(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTest(t, mux)

	body := postJSON(t, mux, "/api/app/games", token, `{"title":"test game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID             int64  `json:"id"`
			RealnameStatus string `json:"realnameStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == 0 {
		t.Fatalf("expected game create success for player: %s", string(body))
	}
}

func TestCreateFreeGameAfterIdentityVerified(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTest(t, mux)
	completeIdentityForTest(t, mux, token)

	body := postJSON(t, mux, "/api/app/games", token, `{"title":"test game","gameType":"free","minPlayers":5,"maxPlayers":8,"cityCode":"110100","cityName":"Beijing","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID       int64  `json:"id"`
			GameType string `json:"gameType"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == 0 || created.Data.GameType != "free" || created.Data.Status != "pending_audit" {
		t.Fatalf("unexpected game response: %s", string(body))
	}
}

func TestCreateGamePersistsMiniProgramDetailFieldsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTest(t, mux)

	payload := `{
		"title":"完整字段局",
		"gameType":"free",
		"coverImage":"mock://cover.jpg",
		"description":"局介绍",
		"highlights":"局亮点",
		"notice":"活动须知",
		"audience":"企业主",
		"participation":"offline",
		"price":0,
		"profitTemplate":"",
		"startAt":"2026-07-10 14:00",
		"endAt":"2026-07-10 16:00",
		"signupStartAt":"2026-07-02 09:00",
		"signupEndAt":"2026-07-09 18:00",
		"tags":["同城","商务"],
		"completionRules":["checkin","review"],
		"primaryCategory":"task",
		"primaryCategoryText":"任务局",
		"secondaryCategory":"project",
		"secondaryCategoryText":"项目交流",
		"type":"free",
		"minPlayers":5,
		"maxPlayers":8,
		"cityCode":"110100",
		"cityName":"北京",
		"address":"朝阳区",
		"longitude":116.3972,
		"latitude":39.9166
	}`
	body := postJSON(t, mux, "/api/app/games", token, payload, http.StatusOK)
	var created struct {
		Data struct {
			ID              int64    `json:"id"`
			CoverImage      string   `json:"coverImage"`
			Description     string   `json:"description"`
			Highlights      string   `json:"highlights"`
			Notice          string   `json:"notice"`
			Audience        string   `json:"audience"`
			Participation   string   `json:"participation"`
			StartAt         string   `json:"startAt"`
			EndAt           string   `json:"endAt"`
			SignupStartAt   string   `json:"signupStartAt"`
			SignupEndAt     string   `json:"signupEndAt"`
			Tags            []string `json:"tags"`
			CompletionRules []string `json:"completionRules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == 0 || created.Data.CoverImage != "mock://cover.jpg" || created.Data.Description != "局介绍" || created.Data.StartAt != "2026-07-10 14:00" {
		t.Fatalf("expected create response to keep mini program detail fields: %s", string(body))
	}
	if len(created.Data.Tags) != 2 || len(created.Data.CompletionRules) != 2 {
		t.Fatalf("expected tags and completion rules in create response: %s", string(body))
	}

	detailBody := getJSON(t, mux, "/api/app/games/"+strconv.FormatInt(created.Data.ID, 10), token, http.StatusOK)
	var detail struct {
		Data struct {
			ID              int64    `json:"id"`
			Description     string   `json:"description"`
			Highlights      string   `json:"highlights"`
			Notice          string   `json:"notice"`
			Audience        string   `json:"audience"`
			Participation   string   `json:"participation"`
			StartAt         string   `json:"startAt"`
			EndAt           string   `json:"endAt"`
			Tags            []string `json:"tags"`
			CompletionRules []string `json:"completionRules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Data.ID != created.Data.ID || detail.Data.Description != "局介绍" || detail.Data.Highlights != "局亮点" || detail.Data.Notice != "活动须知" || detail.Data.Audience != "企业主" || detail.Data.Participation != "offline" || detail.Data.StartAt != "2026-07-10 14:00" || detail.Data.EndAt != "2026-07-10 16:00" {
		t.Fatalf("expected detail response to keep mini program detail fields: %s", string(detailBody))
	}
	if len(detail.Data.Tags) != 2 || len(detail.Data.CompletionRules) != 2 {
		t.Fatalf("expected tags and completion rules in detail response: %s", string(detailBody))
	}
}

func TestGameCategoryConfigAndCreateCategoryFieldsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTest(t, mux)

	configBody := getJSON(t, mux, "/api/app/games/category-config", token, http.StatusOK)
	var config struct {
		Data struct {
			PrimaryCategories []struct {
				Key      string `json:"key"`
				Children []struct {
					Key string `json:"key"`
				} `json:"children"`
			} `json:"primaryCategories"`
			DefaultPrimaryCategory   string `json:"defaultPrimaryCategory"`
			DefaultSecondaryCategory string `json:"defaultSecondaryCategory"`
			SortOptions              []struct {
				Key     string `json:"key"`
				SortKey string `json:"sortKey"`
			} `json:"sortOptions"`
			EventActions []string `json:"eventActions"`
			CreateForm   struct {
				Capacity struct {
					Min int `json:"min"`
					Max int `json:"max"`
				} `json:"capacity"`
				ParticipationModes []struct {
					Key string `json:"key"`
				} `json:"participationModes"`
				FeeTypes []struct {
					Key string `json:"key"`
				} `json:"feeTypes"`
			} `json:"createForm"`
		} `json:"data"`
	}
	if err := json.Unmarshal(configBody, &config); err != nil {
		t.Fatal(err)
	}
	if len(config.Data.PrimaryCategories) == 0 || config.Data.DefaultPrimaryCategory == "" || config.Data.DefaultSecondaryCategory == "" {
		t.Fatalf("unexpected category config: %s", string(configBody))
	}
	if config.Data.CreateForm.Capacity.Min != 5 || config.Data.CreateForm.Capacity.Max != 8 || len(config.Data.CreateForm.ParticipationModes) == 0 || len(config.Data.CreateForm.FeeTypes) == 0 {
		t.Fatalf("expected create form config for mini program: %s", string(configBody))
	}
	if len(config.Data.SortOptions) != 5 || len(config.Data.EventActions) != 4 {
		t.Fatalf("expected hall sort and event action config: %s", string(configBody))
	}

	body := postJSON(t, mux, "/api/app/games", token, `{"title":"category game","gameType":"free","primaryCategory":"task","primaryCategoryText":"Task","secondaryCategory":"project","secondaryCategoryText":"Project","type":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			PrimaryCategory   string `json:"primaryCategory"`
			SecondaryCategory string `json:"secondaryCategory"`
			Type              string `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.PrimaryCategory != "task" || created.Data.SecondaryCategory != "project" || created.Data.Type != "free" {
		t.Fatalf("expected category fields in create response: %s", string(body))
	}
}

func TestAdminGameCategoryConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	adminToken := adminLoginForTestAs(t, mux, "admin", "admin123")
	appToken := loginForTest(t, mux)

	payload := `{
		"primaryCategories":[
			{"key":"custom_admin","name":"后台配置局","order":1,"children":[{"key":"custom_child","name":"后台二级","order":1}]}
		],
		"typeFilters":[
			{"key":"all","name":"全部","order":0},
			{"key":"free","name":"免费局","order":1}
		],
		"locationFilters":[{"key":"all","name":"全国","order":0}],
		"defaultPrimaryCategory":"custom_admin",
		"defaultSecondaryCategory":"custom_child",
		"defaultType":"free",
		"version":"admin-test"
	}`
	putJSON(t, mux, "/api/admin/games/category-config", adminToken, payload, http.StatusOK)

	body := getJSON(t, mux, "/api/app/games/category-config", appToken, http.StatusOK)
	var got struct {
		Data struct {
			PrimaryCategories []struct {
				Key      string `json:"key"`
				Name     string `json:"name"`
				Children []struct {
					Key string `json:"key"`
				} `json:"children"`
			} `json:"primaryCategories"`
			DefaultPrimaryCategory   string `json:"defaultPrimaryCategory"`
			DefaultSecondaryCategory string `json:"defaultSecondaryCategory"`
			Version                  string `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Data.Version != "admin-test" || got.Data.DefaultPrimaryCategory != "custom_admin" || len(got.Data.PrimaryCategories) != 1 {
		t.Fatalf("expected app category config from admin update: %s", string(body))
	}
	if got.Data.PrimaryCategories[0].Key != "custom_admin" || got.Data.PrimaryCategories[0].Children[0].Key != "custom_child" {
		t.Fatalf("expected custom category tree: %s", string(body))
	}

	invalidPayload := `{
		"primaryCategories":[
			{"key":"custom_admin","name":"custom","order":1,"children":[{"key":"custom_child","name":"child","order":1}]}
		],
		"typeFilters":[
			{"key":"all","name":"all","order":0},
			{"key":"unknown_paid","name":"unknown","order":1}
		],
		"defaultPrimaryCategory":"custom_admin",
		"defaultSecondaryCategory":"custom_child",
		"defaultType":"unknown_paid"
	}`
	putJSON(t, mux, "/api/admin/games/category-config", adminToken, invalidPayload, http.StatusUnprocessableEntity)
}

func TestAdminGameRuleConfigsFeedAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	adminToken := adminLoginForTestAs(t, mux, "admin", "admin123")
	appToken := loginForTest(t, mux)

	applicationPayload := `{
		"agreementTitle":"测试入局协议",
		"agreementText":"请确认你已经理解本局规则。",
		"requireRealname":true,
		"requireIntro":true,
		"requireAgreement":true,
		"allowDuplicateApply":false,
		"uploadRequired":true,
		"maxUploadCount":2,
		"allowedUploadTypes":["jpg","PDF","jpg"],
		"minIntroLength":6,
		"maxIntroLength":180,
		"searchEnabled":true,
		"recommendationHint":"后台配置推荐提示",
		"version":"application-test"
	}`
	putJSON(t, mux, "/api/admin/games/application-config", adminToken, applicationPayload, http.StatusOK)
	applicationBody := getJSON(t, mux, "/api/app/games/application-config", appToken, http.StatusOK)
	var applicationResp struct {
		Data gameApplicationConfigDTO `json:"data"`
	}
	if err := json.Unmarshal(applicationBody, &applicationResp); err != nil {
		t.Fatal(err)
	}
	if applicationResp.Data.Version != "application-test" || applicationResp.Data.MaxUploadCount != 2 || len(applicationResp.Data.AllowedUploadTypes) != 2 {
		t.Fatalf("expected app application config from admin update: %s", string(applicationBody))
	}
	if applicationResp.Data.MaxMessageLength != 120 || applicationResp.Data.Texts["submitText"] == "" || applicationResp.Data.Texts["introPlaceholder"] == "" {
		t.Fatalf("expected application page texts and message length defaults: %s", string(applicationBody))
	}
	if applicationResp.Data.AuditPage.PageTitle == "" || len(applicationResp.Data.AuditPage.Filters) == 0 || applicationResp.Data.AuditPage.Texts["approveText"] == "" {
		t.Fatalf("expected application audit page config defaults: %s", string(applicationBody))
	}
	if applicationResp.Data.AuditPage.Detail.PageTitle == "" || len(applicationResp.Data.AuditPage.Detail.SessionItems) == 0 || applicationResp.Data.AuditPage.Detail.Texts["confirmText"] == "" {
		t.Fatalf("expected application audit detail config defaults: %s", string(applicationBody))
	}

	auditPayload := `{
		"autoApproveFreeGames":true,
		"requireManualAuditTypes":["condition","deposit"],
		"requiredRejectReason":true,
		"allowUserResubmitAfterReject":true,
		"batchAuditMaxCount":20,
		"applicationAuditMode":"admin_only",
		"reviewerRoles":["super_admin","audit_admin"],
		"version":"audit-test"
	}`
	putJSON(t, mux, "/api/admin/games/audit-config", adminToken, auditPayload, http.StatusOK)
	auditBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/audit-config", "system_config:read", http.StatusOK)
	var auditResp struct {
		Data struct {
			Config gameAuditConfigDTO `json:"config"`
		} `json:"data"`
	}
	if err := json.Unmarshal(auditBody, &auditResp); err != nil {
		t.Fatal(err)
	}
	if auditResp.Data.Config.Version != "audit-test" || auditResp.Data.Config.ApplicationAuditMode != "admin_only" || auditResp.Data.Config.BatchAuditMaxCount != 20 {
		t.Fatalf("expected admin audit config from update: %s", string(auditBody))
	}

	conditionPayload := `{
		"enabled":true,
		"visibleInMiniProgram":true,
		"adminOnlyCreate":true,
		"ruleItems":[
			{"key":"credit_min_90","name":"信用分不低于90","required":true,"order":1}
		],
		"defaultVisibility":"invite_only",
		"reviewRequired":true,
		"paymentRequired":false,
		"version":"condition-test"
	}`
	putJSON(t, mux, "/api/admin/games/condition-rule-config", adminToken, conditionPayload, http.StatusOK)
	conditionBody := getJSON(t, mux, "/api/app/games/condition-rule-config", appToken, http.StatusOK)
	var conditionResp struct {
		Data gameConditionRuleConfigDTO `json:"data"`
	}
	if err := json.Unmarshal(conditionBody, &conditionResp); err != nil {
		t.Fatal(err)
	}
	if conditionResp.Data.Version != "condition-test" || conditionResp.Data.DefaultVisibility != "invite_only" || len(conditionResp.Data.RuleItems) != 1 {
		t.Fatalf("expected app condition config from admin update: %s", string(conditionBody))
	}

	roleBenefitPayload := `{
		"permissionPrompts":{
			"expert":{"roleType":"expert","title":"测试行家权益提示","primary":"申请行家","secondary":"权益对比"},
			"guide":{"roleType":"guide","title":"测试领路人权益提示","primary":"申请领路人","secondary":"权益对比"}
		},
		"roleComparison":{
			"title":"测试角色权益",
			"roles":[{"key":"player","name":"玩家","level":"Lv.1+"}],
			"benefits":[{"name":"测试权益","player":"✓","leader":"—","expert":"✓"}],
			"primary":"申请角色"
		},
		"version":"role-benefit-test"
	}`
	putJSON(t, mux, "/api/admin/roles/benefit-config", adminToken, roleBenefitPayload, http.StatusOK)
	roleBenefitBody := getJSON(t, mux, "/api/app/role-applications/benefit-config", appToken, http.StatusOK)
	var roleBenefitResp struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(roleBenefitBody, &roleBenefitResp); err != nil {
		t.Fatal(err)
	}
	if roleBenefitResp.Data["version"] != "role-benefit-test" {
		t.Fatalf("expected app role benefit config from admin update: %s", string(roleBenefitBody))
	}
	comparison, _ := roleBenefitResp.Data["roleComparison"].(map[string]interface{})
	if comparison["title"] != "测试角色权益" {
		t.Fatalf("expected role comparison title from admin update: %s", string(roleBenefitBody))
	}

	putJSON(t, mux, "/api/admin/games/audit-config", adminToken, `{"requireManualAuditTypes":["unknown"],"batchAuditMaxCount":20}`, http.StatusUnprocessableEntity)
	putJSON(t, mux, "/api/admin/games/condition-rule-config", adminToken, `{"enabled":true,"ruleItems":[]}`, http.StatusUnprocessableEntity)
	putJSON(t, mux, "/api/admin/roles/benefit-config", adminToken, `{"permissionPrompts":{}}`, http.StatusUnprocessableEntity)
}

func TestReportCenterConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	adminToken := adminLoginForTestAs(t, mux, "admin", "admin123")
	appToken := loginForTest(t, mux)

	payload := `{
		"types":[
			{"key":"custom-dispute","label":"测试争议","reportType":"service_dispute","order":1,"visible":true},
			{"key":"custom-other","label":"其他问题","reportType":"other","order":2,"visible":true}
		],
		"defaultType":"custom-dispute",
		"maxEvidenceCount":4,
		"allowedUploadTypes":["jpg","png"],
		"tips":["后台配置提示A","后台配置提示B"],
		"appealReasons":[{"key":"custom","label":"自定义申诉","order":1,"visible":true}],
		"appealPlaceholder":"后台配置申诉说明",
		"appealUploadNote":"后台配置上传要求",
		"appealFileMaxCount":3,
		"appealUploadFullText":"后台配置满额提示",
		"appealUploadSelectedTemplate":"后台配置已选 {selected}/{max}",
		"appealReviewTitle":"后台配置处理时效",
		"appealReviewRules":["后台配置初审规则"],
		"version":"report-test"
	}`
	putJSON(t, mux, "/api/admin/reports/config", adminToken, payload, http.StatusOK)
	body := getJSON(t, mux, "/api/app/reports/config", appToken, http.StatusOK)
	var resp struct {
		Data reportCenterConfigDTO `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Version != "report-test" || resp.Data.DefaultType != "custom-dispute" || resp.Data.MaxEvidenceCount != 4 || len(resp.Data.Types) != 2 || len(resp.Data.Tips) != 2 {
		t.Fatalf("expected report config from admin update: %s", string(body))
	}
	if len(resp.Data.AppealReasons) != 1 || resp.Data.AppealReasons[0].Key != "custom" || resp.Data.AppealPlaceholder != "后台配置申诉说明" || resp.Data.AppealFileMaxCount != 3 || resp.Data.AppealUploadFullText != "后台配置满额提示" || resp.Data.AppealUploadSelectedTemplate == "" || len(resp.Data.AppealReviewRules) != 1 {
		t.Fatalf("expected report config from admin update: %s", string(body))
	}

	putJSON(t, mux, "/api/admin/reports/config", adminToken, `{"types":[],"maxEvidenceCount":4}`, http.StatusUnprocessableEntity)
	putJSON(t, mux, "/api/admin/reports/config", adminToken, `{"types":[{"key":"bad","label":"bad","reportType":"other"}],"maxEvidenceCount":12}`, http.StatusUnprocessableEntity)
}

func TestGameCancelConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	appToken := loginForTest(t, mux)

	body := getJSON(t, mux, "/api/app/games/cancel-config", appToken, http.StatusOK)
	var resp struct {
		Data gameCancelConfigDTO `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data.Player.ReasonOptions) == 0 || resp.Data.Player.DefaultReason == "" || resp.Data.Player.AgreementText == "" || len(resp.Data.Expert.ReasonOptions) == 0 || resp.Data.Expert.DefaultReason == "" {
		t.Fatalf("expected cancel config for player and expert: %s", string(body))
	}
}

func TestUseSystemConfigRepositorySeedsDefaultGameCategoryConfig(t *testing.T) {
	server := newTestAppServer(auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()), identity.NewService())
	repository := &appTestSystemConfigRepository{values: map[string]json.RawMessage{}}

	server.UseSystemConfigRepository(repository)

	var stored gameCategoryConfigDTO
	if !server.systemConfig.Get(gameCategoryConfigKey, &stored) {
		t.Fatal("expected default game category config to be written into repository")
	}
	if len(stored.PrimaryCategories) == 0 || stored.DefaultType == "" {
		t.Fatalf("unexpected seeded category config: %#v", stored)
	}

	stored.Version = "custom-admin-version"
	if err := server.systemConfig.Set(gameCategoryConfigKey, stored); err != nil {
		t.Fatal(err)
	}
	server.UseSystemConfigRepository(repository)
	var preserved gameCategoryConfigDTO
	if !server.systemConfig.Get(gameCategoryConfigKey, &preserved) {
		t.Fatal("expected preserved category config")
	}
	if preserved.Version != "custom-admin-version" {
		t.Fatalf("expected existing admin config to be preserved, got %q", preserved.Version)
	}

	var applicationConfig gameApplicationConfigDTO
	if !server.systemConfig.Get(gameApplicationConfigKey, &applicationConfig) || applicationConfig.AgreementTitle == "" {
		t.Fatal("expected default application config to be written into repository")
	}
	var auditConfig gameAuditConfigDTO
	if !server.systemConfig.Get(gameAuditConfigKey, &auditConfig) || auditConfig.ApplicationAuditMode == "" {
		t.Fatal("expected default audit config to be written into repository")
	}
	var conditionConfig gameConditionRuleConfigDTO
	if !server.systemConfig.Get(gameConditionRuleConfigKey, &conditionConfig) || len(conditionConfig.RuleItems) == 0 {
		t.Fatal("expected default condition rule config to be written into repository")
	}
	var cancelConfig gameCancelConfigDTO
	if !server.systemConfig.Get(gameCancelConfigKey, &cancelConfig) || len(cancelConfig.Player.ReasonOptions) == 0 || len(cancelConfig.Expert.ReasonOptions) == 0 {
		t.Fatal("expected default cancel config to be written into repository")
	}
	var reportConfig reportCenterConfigDTO
	if !server.systemConfig.Get(reportCenterConfigKey, &reportConfig) || len(reportConfig.Types) == 0 {
		t.Fatal("expected default report center config to be written into repository")
	}
	var pointsConfig pointsPageConfigDTO
	if !server.systemConfig.Get(pointsPageConfigKey, &pointsConfig) || len(pointsConfig.Stats) == 0 || len(pointsConfig.Filters) == 0 {
		t.Fatal("expected default points page config to be written into repository")
	}
	var assetConfig profileAssetManageConfigDTO
	if !server.systemConfig.Get(profileAssetManageConfigKey, &assetConfig) || len(assetConfig.AssetStats) == 0 || len(assetConfig.OrderStatuses) == 0 {
		t.Fatal("expected default profile asset manage config to be written into repository")
	}
	var redemptionOrderConfig redemptionOrderPageConfigDTO
	if !server.systemConfig.Get(redemptionOrderPageConfigKey, &redemptionOrderConfig) || len(redemptionOrderConfig.Tabs) == 0 || redemptionOrderConfig.EmptyText == "" {
		t.Fatal("expected default redemption order page config to be written into repository")
	}
	var reviewPageConfig reviewPageConfigDTO
	if !server.systemConfig.Get(reviewPageConfigKey, &reviewPageConfig) || reviewPageConfig.NavTitle == "" || len(reviewPageConfig.SatisfactionOptions) == 0 || len(reviewPageConfig.RoleConfigs) == 0 {
		t.Fatal("expected default review page config to be written into repository")
	}
	var mapMyCityConfig mapMyCityConfigDTO
	if !server.systemConfig.Get(mapMyCityConfigKey, &mapMyCityConfig) || len(mapMyCityConfig.StoryGroups) == 0 || len(mapMyCityConfig.SummaryStats) == 0 {
		t.Fatal("expected default my city map config to be written into repository")
	}
	var systemRecommendationsConfig gameSystemRecommendationsConfigDTO
	if !server.systemConfig.Get(gameSystemRecommendationsConfigKey, &systemRecommendationsConfig) || systemRecommendationsConfig.Title == "" || len(systemRecommendationsConfig.Categories) == 0 {
		t.Fatal("expected default system recommendations config to be written into repository")
	}
}

type appTestSystemConfigRepository struct {
	values map[string]json.RawMessage
}

func (r *appTestSystemConfigRepository) Get(ctx context.Context, key string) (json.RawMessage, error) {
	value, ok := r.values[key]
	if !ok {
		return nil, systemconfig.ErrNotFound
	}
	return append(json.RawMessage(nil), value...), nil
}

func (r *appTestSystemConfigRepository) Set(ctx context.Context, key string, value json.RawMessage) error {
	r.values[key] = append(json.RawMessage(nil), value...)
	return nil
}

func TestAdminAuditGameApprovesPendingGame(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "admin-audit-game")
	completeIdentityForTest(t, mux, token)

	body := postJSON(t, mux, "/api/app/games", token, `{"title":"audit game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == 0 || created.Data.Status != "pending_audit" {
		t.Fatalf("expected pending audit game: %s", string(body))
	}

	pendingBody := getAdminJSONWithPermission(t, mux, "/api/admin/games?status=pending_audit", "game:read", http.StatusOK)
	var pendingResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
			Total  int    `json:"total"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pendingBody, &pendingResp); err != nil {
		t.Fatal(err)
	}
	if pendingResp.Data.Total != 1 || pendingResp.Data.Status != "pending_audit" || pendingResp.Data.Items[0].ID != created.Data.ID {
		t.Fatalf("expected admin pending audit list to include created game: %s", string(pendingBody))
	}

	publicPendingBody := getJSON(t, mux, "/api/app/games", token, http.StatusOK)
	var publicPendingResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(publicPendingBody, &publicPendingResp); err != nil {
		t.Fatal(err)
	}
	for _, item := range publicPendingResp.Data.Items {
		if item.ID == created.Data.ID {
			t.Fatalf("pending audit game must not appear in app public list: %s", string(publicPendingBody))
		}
	}

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	postAdminJSON(t, mux, "/api/admin/games/"+strconv.FormatInt(created.Data.ID, 10)+"/audit", operatorToken, `{"approve":true}`, http.StatusForbidden)

	auditBody := postAdminJSONWithPermission(t, mux, "/api/admin/games/"+strconv.FormatInt(created.Data.ID, 10)+"/audit", "game:update_status", `{"approve":true,"remark":"ok"}`, http.StatusOK)
	var audited struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(auditBody, &audited); err != nil {
		t.Fatal(err)
	}
	if audited.Data.ID != created.Data.ID || audited.Data.Status != "recruiting" {
		t.Fatalf("expected audited recruiting game: %s", string(auditBody))
	}

	pendingAfterBody := getAdminJSONWithPermission(t, mux, "/api/admin/games?status=pending_audit", "game:read", http.StatusOK)
	var pendingAfterResp struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pendingAfterBody, &pendingAfterResp); err != nil {
		t.Fatal(err)
	}
	if pendingAfterResp.Data.Total != 0 {
		t.Fatalf("expected no pending audit games after audit: %s", string(pendingAfterBody))
	}

	recruitingBody := getAdminJSONWithPermission(t, mux, "/api/admin/games?status=recruiting", "game:read", http.StatusOK)
	var recruitingResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recruitingBody, &recruitingResp); err != nil {
		t.Fatal(err)
	}
	if recruitingResp.Data.Total != 1 || recruitingResp.Data.Items[0].ID != created.Data.ID || recruitingResp.Data.Items[0].Status != "recruiting" {
		t.Fatalf("expected recruiting list to include audited game: %s", string(recruitingBody))
	}

	publicAuditedBody := getJSON(t, mux, "/api/app/games", token, http.StatusOK)
	var publicAuditedResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(publicAuditedBody, &publicAuditedResp); err != nil {
		t.Fatal(err)
	}
	foundAudited := false
	for _, item := range publicAuditedResp.Data.Items {
		if item.ID == created.Data.ID && item.Status == "recruiting" {
			foundAudited = true
		}
	}
	if !foundAudited {
		t.Fatalf("audited recruiting game must appear in app public list: %s", string(publicAuditedBody))
	}

	logBody := getAdminJSONWithPermission(t, mux, "/api/admin/operation-logs", "operation_log:view_full", http.StatusOK)
	var logResp struct {
		Data struct {
			Items []struct {
				Action      string `json:"action"`
				TargetType  string `json:"targetType"`
				TargetID    string `json:"targetId"`
				AdminUserID int64  `json:"adminUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(logBody, &logResp); err != nil {
		t.Fatal(err)
	}
	if !hasOperationLog(logResp.Data.Items, "game:audit", "game", strconv.FormatInt(created.Data.ID, 10), 1) {
		t.Fatalf("expected game audit operation log: %s", string(logBody))
	}
}

func TestAdminBatchAuditGames(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "admin-batch-audit")
	completeIdentityForTest(t, mux, token)

	firstBody := postJSON(t, mux, "/api/app/games", token, `{"title":"batch game 1","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	secondBody := postJSON(t, mux, "/api/app/games", token, `{"title":"batch game 2","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var first, second struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(firstBody, &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(secondBody, &second); err != nil {
		t.Fatal(err)
	}

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	postAdminJSON(t, mux, "/api/admin/games/batch-audit", operatorToken, `{"gameIds":[1],"approve":true}`, http.StatusForbidden)

	body := postAdminJSONWithPermission(t, mux, "/api/admin/games/batch-audit", "game:update_status", `{"gameIds":[`+strconv.FormatInt(first.Data.ID, 10)+`,`+strconv.FormatInt(second.Data.ID, 10)+`],"approve":true,"remark":"batch ok"}`, http.StatusOK)
	var resp struct {
		Data struct {
			Success int `json:"success"`
			Failed  int `json:"failed"`
			Items   []struct {
				ID      int64  `json:"id"`
				Success bool   `json:"success"`
				Status  string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Success != 2 || resp.Data.Failed != 0 || len(resp.Data.Items) != 2 {
		t.Fatalf("expected two batch audit successes: %s", string(body))
	}

	recruitingBody := getAdminJSONWithPermission(t, mux, "/api/admin/games?status=recruiting", "game:read", http.StatusOK)
	var recruitingResp struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recruitingBody, &recruitingResp); err != nil {
		t.Fatal(err)
	}
	if recruitingResp.Data.Total != 2 {
		t.Fatalf("expected two recruiting games after batch audit: %s", string(recruitingBody))
	}
}

func TestAdminCreateConditionGameWhileAppCreateStaysFreeOnly(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "admin-create-condition-game")

	postJSON(t, mux, "/api/app/games", token, `{"title":"app standard","gameType":"standard","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusUnprocessableEntity)
	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	postAdminJSON(t, mux, "/api/admin/games", operatorToken, `{"title":"admin condition","creatorUserId":1,"gameType":"condition","minPlayers":5,"maxPlayers":8,"signupStartAt":"2026-07-01 10:00","signupEndAt":"2026-08-01 09:00","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusForbidden)
	postAdminJSONWithPermission(t, mux, "/api/admin/games", "game:create_admin", `{"title":"same guide","creatorUserId":1,"mainGuideUserId":1,"gameType":"condition","minPlayers":5,"maxPlayers":8,"signupStartAt":"2026-07-01 10:00","signupEndAt":"2026-08-01 09:00","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusUnprocessableEntity)
	postAdminJSONWithPermission(t, mux, "/api/admin/games", "game:create_admin", `{"title":"preset guide is phase two","creatorUserId":1,"mainGuideUserId":2,"gameType":"condition","minPlayers":5,"maxPlayers":8,"signupStartAt":"2026-07-01 10:00","signupEndAt":"2026-08-01 09:00","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusUnprocessableEntity)
	guideToken := loginForTestWithCode(t, mux, "admin-create-condition-guide")
	completeIdentityForTest(t, mux, guideToken)

	body := postAdminJSONWithPermission(t, mux, "/api/admin/games", "game:create_admin", `{"title":"admin condition","creatorUserId":1,"gameType":"condition","minPlayers":5,"maxPlayers":8,"signupStartAt":"2026-07-01 10:00","signupEndAt":"2026-08-01 09:00","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID              int64  `json:"id"`
			CreatorUserID   int64  `json:"creatorUserId"`
			MainGuideUserID int64  `json:"mainGuideUserId"`
			GameType        string `json:"gameType"`
			GameSource      string `json:"gameSource"`
			Status          string `json:"status"`
			CurrentPlayers  int    `json:"currentPlayers"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == 0 || created.Data.CreatorUserID != 1 || created.Data.MainGuideUserID != 0 || created.Data.GameType != "condition" || created.Data.GameSource != "admin" || created.Data.Status != "recruiting" || created.Data.CurrentPlayers != 1 {
		t.Fatalf("unexpected admin created condition game: %s", string(body))
	}
	detailBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/"+strconv.FormatInt(created.Data.ID, 10), "game:read", http.StatusOK)
	var detail struct {
		Data struct {
			Game struct {
				MainGuideUserID int64 `json:"mainGuideUserId"`
				CurrentPlayers  int   `json:"currentPlayers"`
			} `json:"game"`
			MemberIDs []int64 `json:"memberIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Data.Game.MainGuideUserID != 0 || detail.Data.Game.CurrentPlayers != 1 || containsInt64(detail.Data.MemberIDs, 2) {
		t.Fatalf("expected admin detail to keep phase-two guide out of members: %s", string(detailBody))
	}

	standardBody := postAdminJSONWithPermission(t, mux, "/api/admin/games/", "game:create_admin", `{"title":"admin standard","creatorUserId":1,"gameType":"standard","minPlayers":5,"maxPlayers":8,"signupStartAt":"2026-07-01 10:00","signupEndAt":"2026-08-01 09:00","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var standard struct {
		Data struct {
			GameType   string `json:"gameType"`
			GameSource string `json:"gameSource"`
		} `json:"data"`
	}
	if err := json.Unmarshal(standardBody, &standard); err != nil {
		t.Fatal(err)
	}
	if standard.Data.GameType != "standard" || standard.Data.GameSource != "admin" {
		t.Fatalf("unexpected admin created standard game: %s", string(standardBody))
	}

	for _, gameType := range []string{"public_welfare", "aa", "crowdfund", "deposit"} {
		extraBody := postAdminJSONWithPermission(t, mux, "/api/admin/games", "game:create_admin", `{"title":"admin `+gameType+`","creatorUserId":1,"gameType":"`+gameType+`","minPlayers":5,"maxPlayers":8,"signupStartAt":"2026-07-01 10:00","signupEndAt":"2026-08-01 09:00","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
		var extra struct {
			Data struct {
				GameType   string `json:"gameType"`
				GameSource string `json:"gameSource"`
				Status     string `json:"status"`
			} `json:"data"`
		}
		if err := json.Unmarshal(extraBody, &extra); err != nil {
			t.Fatal(err)
		}
		if extra.Data.GameType != gameType || extra.Data.GameSource != "admin" || extra.Data.Status != "recruiting" {
			t.Fatalf("unexpected admin created %s game: %s", gameType, string(extraBody))
		}
	}

	adminListBody := getAdminJSONWithPermission(t, mux, "/api/admin/games?gameType=condition&keyword=condition", "game:read", http.StatusOK)
	var adminListResp struct {
		Data struct {
			Total    int    `json:"total"`
			GameType string `json:"gameType"`
			Keyword  string `json:"keyword"`
			Items    []struct {
				ID       int64  `json:"id"`
				GameType string `json:"gameType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminListBody, &adminListResp); err != nil {
		t.Fatal(err)
	}
	if adminListResp.Data.Total != 1 || adminListResp.Data.GameType != "condition" || adminListResp.Data.Keyword != "condition" || adminListResp.Data.Items[0].ID != created.Data.ID {
		t.Fatalf("expected filtered admin game list: %s", string(adminListBody))
	}

	adminDetailBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/"+strconv.FormatInt(created.Data.ID, 10), "game:read", http.StatusOK)
	var adminDetailResp struct {
		Data struct {
			Game struct {
				ID       int64  `json:"id"`
				GameType string `json:"gameType"`
			} `json:"game"`
			MemberIDs    []int64       `json:"memberIds"`
			Milestones   []interface{} `json:"milestones"`
			Checkins     []interface{} `json:"checkins"`
			ConfirmItems []interface{} `json:"confirmItems"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminDetailBody, &adminDetailResp); err != nil {
		t.Fatal(err)
	}
	if adminDetailResp.Data.Game.ID != created.Data.ID || adminDetailResp.Data.Game.GameType != "condition" || len(adminDetailResp.Data.MemberIDs) != 1 {
		t.Fatalf("expected admin game detail chain: %s", string(adminDetailBody))
	}

	listBody := getJSON(t, mux, "/api/app/games", token, http.StatusOK)
	var listResp struct {
		Data struct {
			Items []struct {
				ID         int64  `json:"id"`
				GameType   string `json:"gameType"`
				GameSource string `json:"gameSource"`
				Status     string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listBody, &listResp); err != nil {
		t.Fatal(err)
	}
	foundCondition := false
	for _, item := range listResp.Data.Items {
		if item.ID == created.Data.ID && item.GameType == "condition" && item.GameSource == "admin" && item.Status == "recruiting" {
			foundCondition = true
		}
	}
	if !foundCondition {
		t.Fatalf("expected app list to show admin condition game: %s", string(listBody))
	}
}

func TestCreateGameIdempotencyKeyReplaysFirstResponse(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	token := loginForTestWithCode(t, mux, "game-idempotency")
	completeIdentityForTest(t, mux, token)

	payload := `{"title":"same submit","gameType":"free","minPlayers":5,"maxPlayers":8,"cityCode":"110100","cityName":"Beijing","startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`
	firstBody := postJSONWithIdempotencyKey(t, mux, "/api/app/games", token, "create-game-1", payload, http.StatusOK)
	secondBody := postJSONWithIdempotencyKey(t, mux, "/api/app/games", token, "create-game-1", payload, http.StatusOK)
	if !bytes.Equal(firstBody, secondBody) {
		t.Fatalf("expected duplicate submit to replay first response\nfirst=%s\nsecond=%s", string(firstBody), string(secondBody))
	}

	postJSON(t, mux, "/api/app/games/1/approve-local", token, `{}`, http.StatusOK)
	listBody := getJSON(t, mux, "/api/app/games", token, http.StatusOK)
	var listResp struct {
		Data struct {
			Items []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listBody, &listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Data.Items) != 1 || listResp.Data.Items[0].Title != "same submit" {
		t.Fatalf("expected one created game after duplicate submit: %s", string(listBody))
	}
}

func TestGameApplicationAndManualStartFlow(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "player")
	completeIdentityForTest(t, mux, playerToken)

	gameBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"test game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var created struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(gameBody, &created); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/games/999/favorite", playerToken, `{}`, http.StatusNotFound)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	favoriteBody := postJSON(t, mux, "/api/app/games/1/favorite", playerToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/favorite", playerToken, `{}`, http.StatusOK)
	var favoriteResp struct {
		Data struct {
			GameID int64 `json:"gameId"`
			Game   struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"game"`
		} `json:"data"`
	}
	if err := json.Unmarshal(favoriteBody, &favoriteResp); err != nil {
		t.Fatal(err)
	}
	if favoriteResp.Data.GameID != 1 || favoriteResp.Data.Game.Status != "recruiting" {
		t.Fatalf("expected favorite game: %s", string(favoriteBody))
	}
	favoriteDetailBody := getJSON(t, mux, "/api/app/games/1", playerToken, http.StatusOK)
	var favoriteDetailResp struct {
		Data struct {
			IsFavorited   bool `json:"isFavorited"`
			FavoriteCount int  `json:"favoriteCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(favoriteDetailBody, &favoriteDetailResp); err != nil {
		t.Fatal(err)
	}
	if !favoriteDetailResp.Data.IsFavorited || favoriteDetailResp.Data.FavoriteCount != 1 {
		t.Fatalf("expected favorite state in game detail: %s", string(favoriteDetailBody))
	}
	favoritesReq := httptest.NewRequest(http.MethodGet, "/api/app/games/favorites/my", nil)
	favoritesReq.Header.Set("Authorization", "Bearer "+playerToken)
	favoritesRec := httptest.NewRecorder()
	mux.ServeHTTP(favoritesRec, favoritesReq)
	if favoritesRec.Code != http.StatusOK {
		t.Fatalf("expected favorites 200, got %d: %s", favoritesRec.Code, favoritesRec.Body.String())
	}
	var favoritesResp struct {
		Data struct {
			Items []struct {
				GameID int64 `json:"gameId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(favoritesRec.Body.Bytes(), &favoritesResp); err != nil {
		t.Fatal(err)
	}
	if len(favoritesResp.Data.Items) != 1 || favoritesResp.Data.Items[0].GameID != 1 {
		t.Fatalf("expected one favorite: %s", favoritesRec.Body.String())
	}

	cancelBody := postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"maybe"}`, http.StatusOK)
	var cancelCandidate struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cancelBody, &cancelCandidate); err != nil {
		t.Fatal(err)
	}
	cancelledBody := postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(cancelCandidate.Data.ID, 10)+"/cancel", playerToken, `{}`, http.StatusOK)
	var cancelledResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cancelledBody, &cancelledResp); err != nil {
		t.Fatal(err)
	}
	if cancelledResp.Data.Status != "cancelled" {
		t.Fatalf("expected cancelled application: %s", string(cancelledBody))
	}

	applicationBody := postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join","fileIds":[7,8]}`, http.StatusOK)
	var application struct {
		Data struct {
			ID      int64   `json:"id"`
			FileIDs []int64 `json:"fileIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(applicationBody, &application); err != nil {
		t.Fatal(err)
	}
	if len(application.Data.FileIDs) != 2 || application.Data.FileIDs[0] != 7 || application.Data.FileIDs[1] != 8 {
		t.Fatalf("expected application file ids: %s", string(applicationBody))
	}
	receivedBody := getJSON(t, mux, "/api/app/game-applications/received?status=pending", creatorToken, http.StatusOK)
	var receivedResp struct {
		Data struct {
			Items []struct {
				ID      int64   `json:"id"`
				GameID  int64   `json:"gameId"`
				Status  string  `json:"status"`
				FileIDs []int64 `json:"fileIds"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(receivedBody, &receivedResp); err != nil {
		t.Fatal(err)
	}
	if len(receivedResp.Data.Items) != 1 || receivedResp.Data.Items[0].ID != application.Data.ID || receivedResp.Data.Items[0].GameID != 1 {
		t.Fatalf("expected creator received pending application: %s", string(receivedBody))
	}
	if len(receivedResp.Data.Items[0].FileIDs) != 2 || receivedResp.Data.Items[0].FileIDs[0] != 7 || receivedResp.Data.Items[0].FileIDs[1] != 8 {
		t.Fatalf("expected received application file ids: %s", string(receivedBody))
	}
	adminApplicationsBody := getAdminJSONWithPermission(t, mux, "/api/admin/game-applications?status=pending", "game:read", http.StatusOK)
	var adminApplicationsResp struct {
		Data struct {
			Items []struct {
				ID      int64   `json:"id"`
				GameID  int64   `json:"gameId"`
				UserID  int64   `json:"userId"`
				FileIDs []int64 `json:"fileIds"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminApplicationsBody, &adminApplicationsResp); err != nil {
		t.Fatal(err)
	}
	if len(adminApplicationsResp.Data.Items) != 1 || adminApplicationsResp.Data.Items[0].ID != application.Data.ID || adminApplicationsResp.Data.Items[0].GameID != 1 || adminApplicationsResp.Data.Items[0].UserID != 2 {
		t.Fatalf("expected admin game application list: %s", string(adminApplicationsBody))
	}
	if len(adminApplicationsResp.Data.Items[0].FileIDs) != 2 || adminApplicationsResp.Data.Items[0].FileIDs[0] != 7 || adminApplicationsResp.Data.Items[0].FileIDs[1] != 8 {
		t.Fatalf("expected admin application file ids: %s", string(adminApplicationsBody))
	}
	adminApplicationsByGameBody := getAdminJSONWithPermission(t, mux, "/api/admin/game-applications?gameId=1&userId=2&status=pending", "game:read", http.StatusOK)
	var adminApplicationsByGameResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminApplicationsByGameBody, &adminApplicationsByGameResp); err != nil {
		t.Fatal(err)
	}
	if len(adminApplicationsByGameResp.Data.Items) != 1 || adminApplicationsByGameResp.Data.Items[0].ID != application.Data.ID {
		t.Fatalf("expected filtered admin application list: %s", string(adminApplicationsByGameBody))
	}
	adminApplicationsEmptyBody := getAdminJSONWithPermission(t, mux, "/api/admin/game-applications?status=approved", "game:read", http.StatusOK)
	var adminApplicationsEmptyResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminApplicationsEmptyBody, &adminApplicationsEmptyResp); err != nil {
		t.Fatal(err)
	}
	if len(adminApplicationsEmptyResp.Data.Items) != 0 {
		t.Fatalf("expected no approved admin applications before audit: %s", string(adminApplicationsEmptyBody))
	}

	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(application.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "game-flow", 3)
	startBody := postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)
	var started struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(startBody, &started); err != nil {
		t.Fatal(err)
	}
	if started.Data.Status != "in_progress" {
		t.Fatalf("expected in_progress, got %s", started.Data.Status)
	}

	progressBody := postJSON(t, mux, "/api/app/games/1/progress-feedbacks", creatorToken, `{"progress":30,"content":"started","fileIds":[1,2]}`, http.StatusOK)
	var progressResp struct {
		Data struct {
			ID       int64   `json:"id"`
			Progress int     `json:"progress"`
			FileIDs  []int64 `json:"fileIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(progressBody, &progressResp); err != nil {
		t.Fatal(err)
	}
	if progressResp.Data.ID == 0 || progressResp.Data.Progress != 30 || len(progressResp.Data.FileIDs) != 2 {
		t.Fatalf("expected progress feedback: %s", string(progressBody))
	}

	progressReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/progress-feedbacks", nil)
	progressReq.Header.Set("Authorization", "Bearer "+playerToken)
	progressRec := httptest.NewRecorder()
	mux.ServeHTTP(progressRec, progressReq)
	if progressRec.Code != http.StatusOK {
		t.Fatalf("expected progress feedbacks 200, got %d: %s", progressRec.Code, progressRec.Body.String())
	}
	var progressListResp struct {
		Data struct {
			LatestProgress int `json:"latestProgress"`
			Items          []struct {
				Progress int `json:"progress"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(progressRec.Body.Bytes(), &progressListResp); err != nil {
		t.Fatal(err)
	}
	if progressListResp.Data.LatestProgress != 30 || len(progressListResp.Data.Items) != 1 {
		t.Fatalf("expected progress list with latest 30: %s", progressRec.Body.String())
	}
	detailBody := getJSON(t, mux, "/api/app/games/1", playerToken, http.StatusOK)
	var detailResp struct {
		Data struct {
			ID            int64 `json:"id"`
			DetailDisplay struct {
				Organizer struct {
					UserID    int64  `json:"userId"`
					Name      string `json:"name"`
					Role      string `json:"role"`
					RoleLabel string `json:"roleLabel"`
				} `json:"organizer"`
			} `json:"detailDisplay"`
			MyRelation struct {
				Role       string `json:"role"`
				IsMember   bool   `json:"isMember"`
				CanEnterIM bool   `json:"canEnterIM"`
				CanConfirm bool   `json:"canConfirm"`
			} `json:"myRelation"`
			MemberIDs []int64 `json:"memberIds"`
			Members   []struct {
				UserID    int64  `json:"userId"`
				Role      string `json:"role"`
				RoleLabel string `json:"roleLabel"`
			} `json:"members"`
			Progress struct {
				Feedbacks []struct {
					Progress int `json:"progress"`
				} `json:"feedbacks"`
			} `json:"progress"`
			IM struct {
				Available bool  `json:"available"`
				RoomID    int64 `json:"roomId"`
			} `json:"im"`
			Review struct {
				Reviewable bool `json:"reviewable"`
			} `json:"review"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.ID != 1 || detailResp.Data.MyRelation.Role != "member" || !detailResp.Data.MyRelation.IsMember || !detailResp.Data.MyRelation.CanEnterIM || !detailResp.Data.MyRelation.CanConfirm {
		t.Fatalf("expected member relation in game detail: %s", string(detailBody))
	}
	if detailResp.Data.DetailDisplay.Organizer.UserID <= 0 || detailResp.Data.DetailDisplay.Organizer.Name == "" || detailResp.Data.DetailDisplay.Organizer.Role == "" || detailResp.Data.DetailDisplay.Organizer.RoleLabel == "" {
		t.Fatalf("expected organizer identity and in-game role from backend: %s", string(detailBody))
	}
	if len(detailResp.Data.MemberIDs) != 5 || len(detailResp.Data.Members) != 5 || len(detailResp.Data.Progress.Feedbacks) != 1 || detailResp.Data.Progress.Feedbacks[0].Progress != 30 {
		t.Fatalf("expected progress and members in game detail: %s", string(detailBody))
	}
	for _, member := range detailResp.Data.Members {
		if member.UserID <= 0 || member.Role == "" || member.RoleLabel == "" {
			t.Fatalf("expected every detail member to include its in-game role: %s", string(detailBody))
		}
	}
	if !detailResp.Data.IM.Available || detailResp.Data.IM.RoomID == 0 || detailResp.Data.Review.Reviewable {
		t.Fatalf("expected active im and no review before confirm in game detail: %s", string(detailBody))
	}
	membersBody := getJSON(t, mux, "/api/app/games/1/members", playerToken, http.StatusOK)
	var membersResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				UserID        int64  `json:"userId"`
				Role          string `json:"role"`
				IsCreator     bool   `json:"isCreator"`
				IsCurrentUser bool   `json:"isCurrentUser"`
				Confirmed     bool   `json:"confirmed"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(membersBody, &membersResp); err != nil {
		t.Fatal(err)
	}
	if membersResp.Data.Total != 5 || len(membersResp.Data.Items) != 5 {
		t.Fatalf("expected five game members: %s", string(membersBody))
	}
	seenCreator := false
	seenCurrent := false
	for _, item := range membersResp.Data.Items {
		if item.UserID == 1 && item.Role == "member" && item.IsCreator {
			seenCreator = true
		}
		if item.UserID == 2 && item.Role == "member" && item.IsCurrentUser {
			seenCurrent = true
		}
	}
	if !seenCreator || !seenCurrent {
		t.Fatalf("expected creator and current member markers: %s", string(membersBody))
	}
	postJSON(t, mux, "/api/app/games/1/progress-feedbacks", creatorToken, `{"progress":20,"content":"backward"}`, http.StatusUnprocessableEntity)

	milestoneBody := postJSON(t, mux, "/api/app/games/1/milestones", creatorToken, `{"title":"first stage"}`, http.StatusOK)
	var milestoneResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(milestoneBody, &milestoneResp); err != nil {
		t.Fatal(err)
	}
	if milestoneResp.Data.ID == 0 || milestoneResp.Data.Status != "pending" {
		t.Fatalf("expected milestone: %s", string(milestoneBody))
	}
	updatedMilestoneBody := putJSON(t, mux, "/api/app/games/1/milestones/1", creatorToken, `{"title":"first stage done","status":"completed"}`, http.StatusOK)
	var updatedMilestoneResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(updatedMilestoneBody, &updatedMilestoneResp); err != nil {
		t.Fatal(err)
	}
	if updatedMilestoneResp.Data.ID != 1 || updatedMilestoneResp.Data.Title != "first stage done" || updatedMilestoneResp.Data.Status != "completed" {
		t.Fatalf("expected updated milestone: %s", string(updatedMilestoneBody))
	}
	checkinBody := postJSON(t, mux, "/api/app/games/1/checkins", playerToken, `{"milestoneId":1,"checkinType":"progress","content":"done","fileIds":[1]}`, http.StatusOK)
	var checkinResp struct {
		Data struct {
			ID          int64   `json:"id"`
			MilestoneID int64   `json:"milestoneId"`
			FileIDs     []int64 `json:"fileIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(checkinBody, &checkinResp); err != nil {
		t.Fatal(err)
	}
	if checkinResp.Data.ID == 0 || checkinResp.Data.MilestoneID != 1 || len(checkinResp.Data.FileIDs) != 1 {
		t.Fatalf("expected checkin: %s", string(checkinBody))
	}
	milestonesReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/milestones", nil)
	milestonesReq.Header.Set("Authorization", "Bearer "+playerToken)
	milestonesRec := httptest.NewRecorder()
	mux.ServeHTTP(milestonesRec, milestonesReq)
	if milestonesRec.Code != http.StatusOK {
		t.Fatalf("expected milestones 200, got %d: %s", milestonesRec.Code, milestonesRec.Body.String())
	}
	checkinsReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/checkins", nil)
	checkinsReq.Header.Set("Authorization", "Bearer "+creatorToken)
	checkinsRec := httptest.NewRecorder()
	mux.ServeHTTP(checkinsRec, checkinsReq)
	if checkinsRec.Code != http.StatusOK {
		t.Fatalf("expected checkins 200, got %d: %s", checkinsRec.Code, checkinsRec.Body.String())
	}
	postAdminJSONWithPermission(t, mux, "/api/admin/game-checkins/1/mark-invalid", "game:progress:manage", `{}`, http.StatusUnprocessableEntity)
	postAdminJSONWithPermission(t, mux, "/api/admin/game-checkins/1/mark-invalid", "game:progress:manage", `{"reason":"duplicate proof"}`, http.StatusOK)
	filteredCheckinsReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/checkins", nil)
	filteredCheckinsReq.Header.Set("Authorization", "Bearer "+creatorToken)
	filteredCheckinsRec := httptest.NewRecorder()
	mux.ServeHTTP(filteredCheckinsRec, filteredCheckinsReq)
	if filteredCheckinsRec.Code != http.StatusOK {
		t.Fatalf("expected filtered checkins 200, got %d: %s", filteredCheckinsRec.Code, filteredCheckinsRec.Body.String())
	}
	var filteredCheckinsResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(filteredCheckinsRec.Body.Bytes(), &filteredCheckinsResp); err != nil {
		t.Fatal(err)
	}
	if len(filteredCheckinsResp.Data.Items) != 0 {
		t.Fatalf("expected invalid checkin filtered out: %s", filteredCheckinsRec.Body.String())
	}
	postJSON(t, mux, "/api/app/games/1/retrospectives", playerToken, `{"content":"too early","againIntent":"yes"}`, http.StatusConflict)
	postJSON(t, mux, "/api/app/games/1/continue", creatorToken, `{"title":"too early continue"}`, http.StatusConflict)
	postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	retroBody := postJSON(t, mux, "/api/app/games/1/retrospectives", playerToken, `{"content":"great game","againIntent":"yes"}`, http.StatusOK)
	var retroResp struct {
		Data struct {
			ID          int64  `json:"id"`
			AgainIntent string `json:"againIntent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(retroBody, &retroResp); err != nil {
		t.Fatal(err)
	}
	if retroResp.Data.ID == 0 || retroResp.Data.AgainIntent != "yes" {
		t.Fatalf("expected retrospective: %s", string(retroBody))
	}
	continueBody := postJSON(t, mux, "/api/app/games/1/continue", creatorToken, `{"title":"next round"}`, http.StatusOK)
	var continueResp struct {
		Data struct {
			OriginalGameID int64 `json:"originalGameId"`
			Draft          struct {
				ID         int64  `json:"id"`
				GameSource string `json:"gameSource"`
				Status     string `json:"status"`
				Title      string `json:"title"`
			} `json:"draft"`
		} `json:"data"`
	}
	if err := json.Unmarshal(continueBody, &continueResp); err != nil {
		t.Fatal(err)
	}
	if continueResp.Data.OriginalGameID != 1 || continueResp.Data.Draft.ID == 0 || continueResp.Data.Draft.Status != "draft" || continueResp.Data.Draft.GameSource != "continue" {
		t.Fatalf("expected continue draft: %s", string(continueBody))
	}

	lateToken := loginForTestWithCode(t, mux, "late")
	completeIdentityForTest(t, mux, lateToken)
	postJSON(t, mux, "/api/app/games/1/applications", lateToken, `{"reason":"late"}`, http.StatusConflict)
	deleteFavoriteReq := httptest.NewRequest(http.MethodDelete, "/api/app/games/1/favorite", nil)
	deleteFavoriteReq.Header.Set("Authorization", "Bearer "+playerToken)
	deleteFavoriteRec := httptest.NewRecorder()
	mux.ServeHTTP(deleteFavoriteRec, deleteFavoriteReq)
	if deleteFavoriteRec.Code != http.StatusOK {
		t.Fatalf("expected delete favorite 200, got %d: %s", deleteFavoriteRec.Code, deleteFavoriteRec.Body.String())
	}
	deleteFavoriteAgainReq := httptest.NewRequest(http.MethodDelete, "/api/app/games/1/favorite", nil)
	deleteFavoriteAgainReq.Header.Set("Authorization", "Bearer "+playerToken)
	deleteFavoriteAgainRec := httptest.NewRecorder()
	mux.ServeHTTP(deleteFavoriteAgainRec, deleteFavoriteAgainReq)
	if deleteFavoriteAgainRec.Code != http.StatusOK {
		t.Fatalf("expected idempotent delete favorite 200, got %d: %s", deleteFavoriteAgainRec.Code, deleteFavoriteAgainRec.Body.String())
	}
	outsiderProgressReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/progress-feedbacks", nil)
	outsiderProgressReq.Header.Set("Authorization", "Bearer "+lateToken)
	outsiderProgressRec := httptest.NewRecorder()
	mux.ServeHTTP(outsiderProgressRec, outsiderProgressReq)
	if outsiderProgressRec.Code != http.StatusForbidden {
		t.Fatalf("expected outsider progress 403, got %d: %s", outsiderProgressRec.Code, outsiderProgressRec.Body.String())
	}
	outsiderMilestonesReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/milestones", nil)
	outsiderMilestonesReq.Header.Set("Authorization", "Bearer "+lateToken)
	outsiderMilestonesRec := httptest.NewRecorder()
	mux.ServeHTTP(outsiderMilestonesRec, outsiderMilestonesReq)
	if outsiderMilestonesRec.Code != http.StatusForbidden {
		t.Fatalf("expected outsider milestones 403, got %d: %s", outsiderMilestonesRec.Code, outsiderMilestonesRec.Body.String())
	}
}

func TestGameInvitationRespondCreatesApplication(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "invite-creator")
	completeIdentityForTest(t, mux, creatorToken)
	inviteeToken := loginForTestWithCode(t, mux, "invitee")
	completeIdentityForTest(t, mux, inviteeToken)
	outsiderToken := loginForTestWithCode(t, mux, "invite-outsider")
	completeIdentityForTest(t, mux, outsiderToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"invite game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	inviteBody := postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":2,"message":"join us"}`, http.StatusOK)
	var inviteResp struct {
		Data struct {
			Invitation struct {
				ID           int64  `json:"id"`
				GameID       int64  `json:"gameId"`
				TargetUserID int64  `json:"targetUserId"`
				Status       string `json:"status"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(inviteBody, &inviteResp); err != nil {
		t.Fatal(err)
	}
	if inviteResp.Data.Invitation.ID == 0 || inviteResp.Data.Invitation.GameID != 1 || inviteResp.Data.Invitation.TargetUserID != 2 || inviteResp.Data.Invitation.Status != "pending" {
		t.Fatalf("expected pending invitation: %s", string(inviteBody))
	}

	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+"/respond", outsiderToken, `{"accept":true}`, http.StatusForbidden)
	respondBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+"/respond", inviteeToken, `{"accept":true,"reason":"accept invite"}`, http.StatusOK)
	var respondResp struct {
		Data struct {
			Invitation struct {
				Status        string `json:"status"`
				ApplicationID int64  `json:"applicationId"`
			} `json:"invitation"`
			Application struct {
				ID     int64  `json:"id"`
				GameID int64  `json:"gameId"`
				UserID int64  `json:"userId"`
				Status string `json:"status"`
			} `json:"application"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respondBody, &respondResp); err != nil {
		t.Fatal(err)
	}
	if respondResp.Data.Invitation.Status != "accepted" || respondResp.Data.Invitation.ApplicationID != respondResp.Data.Application.ID {
		t.Fatalf("expected accepted invitation linked to application: %s", string(respondBody))
	}
	if respondResp.Data.Application.GameID != 1 || respondResp.Data.Application.UserID != 2 || respondResp.Data.Application.Status != "pending" {
		t.Fatalf("expected pending application from invitation: %s", string(respondBody))
	}
	getJSON(t, mux, "/api/app/games/1/members", inviteeToken, http.StatusForbidden)

	receivedBody := getJSON(t, mux, "/api/app/game-applications/received?status=pending", creatorToken, http.StatusOK)
	var receivedResp struct {
		Data struct {
			Items []struct {
				ID     int64 `json:"id"`
				UserID int64 `json:"userId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(receivedBody, &receivedResp); err != nil {
		t.Fatal(err)
	}
	if len(receivedResp.Data.Items) != 1 || receivedResp.Data.Items[0].ID != respondResp.Data.Application.ID || receivedResp.Data.Items[0].UserID != 2 {
		t.Fatalf("expected creator received pending invited application: %s", string(receivedBody))
	}
	connectionsBody := getJSON(t, mux, "/api/app/connections/my", creatorToken, http.StatusOK)
	var connectionsResp struct {
		Data struct {
			Items []connectionTestItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(connectionsBody, &connectionsResp); err != nil {
		t.Fatal(err)
	}
	if !hasConnectionStrength(connectionsResp.Data.Items, 2, "guide_match", 3) {
		t.Fatalf("expected accepted invitation to create guide_match connection: %s", string(connectionsBody))
	}
	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+"/respond", inviteeToken, `{"accept":false}`, http.StatusConflict)
}

func TestNotificationInvitationRequiresDetailPageResponseHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "notice-invite-creator")
	completeIdentityForTest(t, mux, creatorToken)
	inviteeToken := loginForTestWithCode(t, mux, "notice-invitee")
	completeIdentityForTest(t, mux, inviteeToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"notice invite game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":2,"message":"join from notice"}`, http.StatusOK)

	noticesBody := getJSON(t, mux, "/api/app/notifications", inviteeToken, http.StatusOK)
	var noticesResp struct {
		Data struct {
			Items []struct {
				ID         int64  `json:"id"`
				NotifyType string `json:"notifyType"`
				BizType    string `json:"bizType"`
				Status     string `json:"status"`
			} `json:"items"`
			Sections []struct {
				Key   string `json:"key"`
				Items []struct {
					ID          int64  `json:"id"`
					Route       string `json:"route"`
					DetailRoute string `json:"detailRoute"`
					Actions     []struct {
						Key  string `json:"key"`
						Text string `json:"text"`
					} `json:"actions"`
				} `json:"items"`
			} `json:"sections"`
			QuickActions []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"quickActions"`
			Tabs []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"tabs"`
			Texts map[string]string `json:"texts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesBody, &noticesResp); err != nil {
		t.Fatal(err)
	}
	var invitationNoticeID int64
	for _, item := range noticesResp.Data.Items {
		if item.NotifyType == "game_invitation" && item.BizType == "game_invitation" && item.Status == "unread" {
			invitationNoticeID = item.ID
			break
		}
	}
	if invitationNoticeID == 0 {
		t.Fatalf("expected game invitation notification: %s", string(noticesBody))
	}
	if len(noticesResp.Data.Sections) == 0 || len(noticesResp.Data.Sections[0].Items) == 0 || len(noticesResp.Data.Sections[0].Items[0].Actions) != 1 || noticesResp.Data.Sections[0].Items[0].Actions[0].Key != "detail" {
		t.Fatalf("expected invitation notification to expose detail entry only: %s", string(noticesBody))
	}
	if !strings.Contains(noticesResp.Data.Sections[0].Items[0].Route, "pages/game/player-confirm/index") || noticesResp.Data.Sections[0].Items[0].DetailRoute == "" {
		t.Fatalf("expected player invitation detail route: %s", string(noticesBody))
	}
	if len(noticesResp.Data.QuickActions) == 0 || len(noticesResp.Data.Tabs) == 0 || noticesResp.Data.Texts["actionSuccessText"] == "" || noticesResp.Data.Sections[0].Items[0].Actions[0].Text == "" {
		t.Fatalf("expected configured message center payload: %s", string(noticesBody))
	}

	messageCenterBody := getJSON(t, mux, "/api/app/messages/center", inviteeToken, http.StatusOK)
	if !strings.Contains(string(messageCenterBody), `"quickActions"`) || !strings.Contains(string(messageCenterBody), `"texts"`) {
		t.Fatalf("expected message center compatible route: %s", string(messageCenterBody))
	}
	if !strings.Contains(string(messageCenterBody), `"key":"friend"`) || !strings.Contains(string(messageCenterBody), `"iconSrc"`) {
		t.Fatalf("expected message center quick action icons from config: %s", string(messageCenterBody))
	}

	messageMyConfigBody := getJSON(t, mux, "/api/app/messages/my-config", inviteeToken, http.StatusOK)
	if !strings.Contains(string(messageMyConfigBody), `"quickActions"`) || !strings.Contains(string(messageMyConfigBody), `"messageText"`) || !strings.Contains(string(messageMyConfigBody), `"justNowText"`) {
		t.Fatalf("expected message my page config: %s", string(messageMyConfigBody))
	}

	detailBody := postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(invitationNoticeID, 10)+"/actions", inviteeToken, `{"action":"detail"}`, http.StatusOK)
	if !strings.Contains(string(detailBody), `"handled":true`) || !strings.Contains(string(detailBody), `pages/game/player-confirm/index`) {
		t.Fatalf("expected invitation detail action target: %s", string(detailBody))
	}

	postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(invitationNoticeID, 10)+"/actions", inviteeToken, `{"action":"accept","reason":"ok"}`, http.StatusUnprocessableEntity)
	noticesBody = getJSON(t, mux, "/api/app/notifications", inviteeToken, http.StatusOK)
	if !strings.Contains(string(noticesBody), `"status":"read"`) {
		t.Fatalf("expected notification read after action: %s", string(noticesBody))
	}
}

func TestInvitationPairCreatesRoleSpecificSuccessNotificationsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	guideToken := loginForTestWithCode(t, mux, "success-notice-guide")
	completeIdentityForTest(t, mux, guideToken)
	playerToken := loginForTestWithCode(t, mux, "success-notice-player")
	completeIdentityForTest(t, mux, playerToken)
	expertToken := loginForTestWithCode(t, mux, "success-notice-expert")
	completeIdentityForTest(t, mux, expertToken)
	server.profiles.GrantRole(3, "expert")

	postJSON(t, mux, "/api/app/games", guideToken, `{"title":"success notice game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2030-01-01 10:00","endAt":"2030-01-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", guideToken, `{}`, http.StatusOK)

	createInvitation := func(targetUserID int64, roleType string) int64 {
		body := postJSON(t, mux, "/api/app/games/1/guide-invitations", guideToken, `{"targetUserId":`+strconv.FormatInt(targetUserID, 10)+`,"roleType":"`+roleType+`","message":"confirm the group"}`, http.StatusOK)
		var response struct {
			Data struct {
				Invitation struct {
					ID int64 `json:"id"`
				} `json:"invitation"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		return response.Data.Invitation.ID
	}

	playerInvitationID := createInvitation(2, "player")
	expertInvitationID := createInvitation(3, "expert")
	acceptInvitation := func(invitationID int64, token string) int64 {
		body := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(invitationID, 10)+"/respond", token, `{"accept":true}`, http.StatusOK)
		var response struct {
			Data struct {
				Application struct {
					ID     int64  `json:"id"`
					Status string `json:"status"`
				} `json:"application"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		if response.Data.Application.ID <= 0 || response.Data.Application.Status != "pending" {
			t.Fatalf("expected accepted invitation to create pending application: %s", string(body))
		}
		return response.Data.Application.ID
	}
	playerApplicationID := acceptInvitation(playerInvitationID, playerToken)
	expertApplicationID := acceptInvitation(expertInvitationID, expertToken)

	reviewNotices := getJSON(t, mux, "/api/app/notifications?type=game_application", guideToken, http.StatusOK)
	if !strings.Contains(string(reviewNotices), `"title":"邀请已接受，待你审核"`) ||
		!strings.Contains(string(reviewNotices), `pages/game/audit/index?gameId=1`) {
		t.Fatalf("accepted invitations must notify the guide to review applications: %s", string(reviewNotices))
	}
	pendingReviewProgress := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(playerInvitationID, 10), guideToken, http.StatusOK)
	if !strings.Contains(string(pendingReviewProgress), `"statusText":"待审核"`) ||
		!strings.Contains(string(pendingReviewProgress), `"primaryActionText":"前往审核"`) ||
		strings.Contains(string(pendingReviewProgress), `"successRoute":"pages/game/success-guide`) {
		t.Fatalf("accepted invitations must remain pending review without a success route: %s", string(pendingReviewProgress))
	}
	beforeAudit := getJSON(t, mux, "/api/app/notifications?type=game_invitation_success", guideToken, http.StatusOK)
	if strings.Contains(string(beforeAudit), `"notifyType":"game_invitation_success"`) {
		t.Fatalf("success notification must wait for application audits: %s", string(beforeAudit))
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(playerApplicationID, 10)+"/audit", guideToken, `{"approve":true}`, http.StatusOK)
	afterFirstAudit := getJSON(t, mux, "/api/app/notifications?type=game_invitation_success", guideToken, http.StatusOK)
	if strings.Contains(string(afterFirstAudit), `"notifyType":"game_invitation_success"`) {
		t.Fatalf("success notification must wait for both application audits: %s", string(afterFirstAudit))
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(expertApplicationID, 10)+"/audit", guideToken, `{"approve":true}`, http.StatusOK)
	approvedProgress := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(playerInvitationID, 10), guideToken, http.StatusOK)
	if !strings.Contains(string(approvedProgress), `"statusText":"组局成功"`) ||
		!strings.Contains(string(approvedProgress), `"successRoute":"pages/game/success-guide/index?gameId=1"`) {
		t.Fatalf("both approved applications must expose the success progress state: %s", string(approvedProgress))
	}

	guideNotices := getJSON(t, mux, "/api/app/notifications?type=game_invitation_success", guideToken, http.StatusOK)
	if !strings.Contains(string(guideNotices), `"notifyType":"game_invitation_success"`) || !strings.Contains(string(guideNotices), `pages/game/success-guide/index?gameId=1`) {
		t.Fatalf("expected guide success notification and route: %s", string(guideNotices))
	}
	readyToStartNotices := getJSON(t, mux, "/api/app/notifications?type=game_ready_to_start", guideToken, http.StatusOK)
	if countNotificationsByType(t, readyToStartNotices, "game_ready_to_start") != 0 {
		t.Fatalf("group success must not send a ready-to-start message before the game is full: %s", string(readyToStartNotices))
	}
	approveExtraMembersForHTTP(t, mux, guideToken, 1, "ready-to-start", 5)
	readyToStartNotices = getJSON(t, mux, "/api/app/notifications?type=game_ready_to_start", guideToken, http.StatusOK)
	if countNotificationsByType(t, readyToStartNotices, "game_ready_to_start") != 1 ||
		!strings.Contains(string(readyToStartNotices), `"title":"组局成功，可以开局"`) ||
		!strings.Contains(string(readyToStartNotices), `pages/game/detail/index?id=1`) {
		t.Fatalf("expected one ready-to-start message for the initiator after the game is full: %s", string(readyToStartNotices))
	}
	expertNotices := getJSON(t, mux, "/api/app/notifications?type=game_invitation_success", expertToken, http.StatusOK)
	if !strings.Contains(string(expertNotices), `"notifyType":"game_invitation_success"`) || !strings.Contains(string(expertNotices), `pages/game/success-expert/index?gameId=1`) {
		t.Fatalf("expected expert success notification and route: %s", string(expertNotices))
	}
	getJSON(t, mux, "/api/app/games/1/success-detail", expertToken, http.StatusOK)
	guideSuccess := getJSON(t, mux, "/api/app/games/1/guide-success-detail", guideToken, http.StatusOK)
	if !strings.Contains(string(guideSuccess), `"userId":2`) || !strings.Contains(string(guideSuccess), `"userId":3`) {
		t.Fatalf("expected real player and expert in guide success detail: %s", string(guideSuccess))
	}
}

func TestNotificationActionsReturnConcreteTargetsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "notice-action-target")
	completeIdentityForTest(t, mux, token)
	userID := currentUserIDForTest(t, mux, token)

	reviewNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "review_remind",
		Title:      "review",
		Content:    "review",
		BizType:    "game",
		BizID:      7,
	})
	routeNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "progress_feedback_remind",
		Title:      "route",
		Content:    "route",
		BizType:    "game",
		BizID:      8,
	})
	reportNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "report_handled",
		Title:      "report",
		Content:    "report",
		BizType:    "report",
		BizID:      9,
	})

	type actionResp struct {
		Data struct {
			Handled bool `json:"handled"`
			Target  struct {
				Route    string `json:"route"`
				RouteKey string `json:"routeKey"`
				BizType  string `json:"bizType"`
				BizID    int64  `json:"bizId"`
			} `json:"target"`
		} `json:"data"`
	}

	processBody := postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(reviewNotice.ID, 10)+"/actions", token, `{"action":"process"}`, http.StatusOK)
	var processResp actionResp
	if err := json.Unmarshal(processBody, &processResp); err != nil {
		t.Fatal(err)
	}
	if !processResp.Data.Handled || processResp.Data.Target.Route != "pages/game/review/index?gameId=7" || processResp.Data.Target.RouteKey != "gameReview" {
		t.Fatalf("expected review process target: %s", string(processBody))
	}

	routeBody := postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(routeNotice.ID, 10)+"/actions", token, `{"action":"route"}`, http.StatusOK)
	var routeResp actionResp
	if err := json.Unmarshal(routeBody, &routeResp); err != nil {
		t.Fatal(err)
	}
	if !routeResp.Data.Handled || routeResp.Data.Target.Route != "pages/map/index?gameId=8&mode=route" || routeResp.Data.Target.RouteKey != "mapRoute" {
		t.Fatalf("expected route target: %s", string(routeBody))
	}

	contactNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "progress_feedback_remind",
		Title:      "contact",
		Content:    "contact",
		BizType:    "game",
		BizID:      10,
	})
	contactBody := postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(contactNotice.ID, 10)+"/actions", token, `{"action":"contact"}`, http.StatusOK)
	var contactResp actionResp
	if err := json.Unmarshal(contactBody, &contactResp); err != nil {
		t.Fatal(err)
	}
	if !contactResp.Data.Handled || contactResp.Data.Target.Route != "pages/im/room/index?gameId=10" || contactResp.Data.Target.RouteKey != "imRoom" {
		t.Fatalf("expected contact target: %s", string(contactBody))
	}

	detailBody := postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(reportNotice.ID, 10)+"/actions", token, `{"action":"detail"}`, http.StatusOK)
	var detailResp actionResp
	if err := json.Unmarshal(detailBody, &detailResp); err != nil {
		t.Fatal(err)
	}
	if !detailResp.Data.Handled || detailResp.Data.Target.Route != "pages/profile/system-management/report-record-detail/index?id=9" || detailResp.Data.Target.BizType != "report" || detailResp.Data.Target.BizID != 9 {
		t.Fatalf("expected report detail target: %s", string(detailBody))
	}

	badNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "system",
		Title:      "bad",
		Content:    "bad",
		BizType:    "system",
		BizID:      11,
	})
	postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(badNotice.ID, 10)+"/actions", token, `{"action":"unknown"}`, http.StatusUnprocessableEntity)
}

func TestMessageDetailEndpointsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "message-detail")
	completeIdentityForTest(t, mux, token)
	userID := currentUserIDForTest(t, mux, token)

	tradeNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "trade_warning",
		Title:      "trade",
		Content:    "trade",
		BizType:    "game",
		BizID:      8,
	})
	tradeBody := getJSON(t, mux, "/api/app/messages/trade-warning?warningId="+strconv.FormatInt(tradeNotice.ID, 10)+"&orderId=GD2024062901", token, http.StatusOK)
	var tradeResp struct {
		Data struct {
			Warning struct {
				HighlightText string `json:"highlightText"`
			} `json:"warning"`
			Order struct {
				OrderNo string `json:"orderNo"`
			} `json:"order"`
			Texts map[string]string `json:"texts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(tradeBody, &tradeResp); err != nil {
		t.Fatal(err)
	}
	if tradeResp.Data.Warning.HighlightText == "" || tradeResp.Data.Order.OrderNo != "GD2024062901" {
		t.Fatalf("expected trade warning detail, got: %s", string(tradeBody))
	}
	if tradeResp.Data.Texts["countdownTitle"] == "" || tradeResp.Data.Texts["deliveryTitle"] == "" {
		t.Fatalf("expected trade warning page texts, got: %s", string(tradeBody))
	}

	systemNotice := server.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "report_created",
		Title:      "system",
		Content:    "system",
		BizType:    "report",
		BizID:      9,
	})
	systemBody := getJSON(t, mux, "/api/app/messages/system-notification?messageId="+strconv.FormatInt(systemNotice.ID, 10), token, http.StatusOK)
	var systemResp struct {
		Data struct {
			MessageID string `json:"messageId"`
			Article   struct {
				Title string `json:"title"`
			} `json:"article"`
			Texts map[string]string `json:"texts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(systemBody, &systemResp); err != nil {
		t.Fatal(err)
	}
	if systemResp.Data.MessageID != strconv.FormatInt(systemNotice.ID, 10) || systemResp.Data.Article.Title == "" {
		t.Fatalf("expected system notification detail, got: %s", string(systemBody))
	}
	if systemResp.Data.Texts["feedbackSuccessText"] == "" || systemResp.Data.Texts["loadFailedText"] == "" {
		t.Fatalf("expected system notification page texts, got: %s", string(systemBody))
	}
}

func TestGameInviteAggregatesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(gameInviteConfigKey, gameInviteConfigDTO{
		MinPlayerCount:      1,
		MaxPlayerCount:      2,
		BudgetMaxAmount:     6000,
		DefaultBudget:       "1200",
		DefaultTitle:        "数据库邀请测试标题",
		DefaultDetail:       "数据库邀请测试详情",
		PlayerIntroTemplate: "数据库模板-{expertName}",
		ActivityTypes: []inviteActivityTypeDTO{
			{Key: "db", Name: "数据库类型"},
		},
		RewardRateConfig: inviteRewardRateConfigDTO{
			PlatformServiceRate:   8,
			SystemGuideRewardRate: 12,
			InviteRewardRate:      30,
		},
	}); err != nil {
		t.Fatalf("seed invite config failed: %v", err)
	}
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "invite-aggregate-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "invite-aggregate-player")
	completeIdentityForTest(t, mux, playerToken)
	expertToken := loginForTestWithCode(t, mux, "invite-aggregate-expert")
	completeIdentityForTest(t, mux, expertToken)

	server.profiles.GrantRole(3, "expert")
	if _, err := server.profiles.UpdateExpertSkill(3, profiles.ExpertSkillRequest{
		SkillTree:   []string{"product", "strategy"},
		ServiceTags: []string{"architecture"},
	}); err != nil {
		t.Fatal(err)
	}

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"aggregate game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	inviteBody := postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":2,"message":"please join"}`, http.StatusOK)
	var inviteResp struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(inviteBody, &inviteResp); err != nil {
		t.Fatal(err)
	}

	configBody := getJSON(t, mux, "/api/app/game-invites/player-config", creatorToken, http.StatusOK)
	var configResp struct {
		Data struct {
			MinPlayerCount      int    `json:"minPlayerCount"`
			MaxPlayerCount      int    `json:"maxPlayerCount"`
			BudgetMaxAmount     int    `json:"budgetMaxAmount"`
			DefaultBudget       string `json:"defaultBudget"`
			DefaultTitle        string `json:"defaultTitle"`
			DefaultDetail       string `json:"defaultDetail"`
			PlayerIntroTemplate string `json:"playerIntroTemplate"`
			ActivityTypes       []struct {
				Key  string `json:"key"`
				Name string `json:"name"`
			} `json:"activityTypes"`
			RewardRateConfig struct {
				PlatformServiceRate   int `json:"platformServiceRate"`
				SystemGuideRewardRate int `json:"systemGuideRewardRate"`
				InviteRewardRate      int `json:"inviteRewardRate"`
			} `json:"rewardRateConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(configBody, &configResp); err != nil {
		t.Fatal(err)
	}
	if configResp.Data.MinPlayerCount != 1 || configResp.Data.MaxPlayerCount != 2 || configResp.Data.BudgetMaxAmount != 6000 || configResp.Data.DefaultBudget != "1200" || configResp.Data.DefaultTitle != "数据库邀请测试标题" || configResp.Data.DefaultDetail != "数据库邀请测试详情" || len(configResp.Data.ActivityTypes) != 1 || configResp.Data.ActivityTypes[0].Key != "db" || configResp.Data.RewardRateConfig.InviteRewardRate != 30 {
		t.Fatalf("unexpected invite config: %s", string(configBody))
	}
	if !strings.Contains(configResp.Data.PlayerIntroTemplate, "数据库模板-") || strings.Contains(configResp.Data.PlayerIntroTemplate, "{expertName}") {
		t.Fatalf("expected rendered invite intro template: %s", string(configBody))
	}

	playersBody := getJSON(t, mux, "/api/app/game-invites/players", creatorToken, http.StatusOK)
	var playersResp struct {
		Data struct {
			List []struct {
				UserID int64 `json:"userId"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(playersBody, &playersResp); err != nil {
		t.Fatal(err)
	}
	if len(playersResp.Data.List) == 0 || playersResp.Data.List[0].UserID == 1 {
		t.Fatalf("expected invite player candidates excluding self: %s", string(playersBody))
	}

	recommendBody := getJSON(t, mux, "/api/app/game-invites/system-recommendations", creatorToken, http.StatusOK)
	var recommendResp struct {
		Data struct {
			Title           string `json:"title"`
			EmptyText       string `json:"emptyText"`
			SummaryTemplate string `json:"summaryTemplate"`
			Categories      []struct {
				Key string `json:"key"`
			} `json:"categories"`
			Experts []struct {
				UserID   int64  `json:"userId"`
				RoleType string `json:"roleType"`
			} `json:"experts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recommendBody, &recommendResp); err != nil {
		t.Fatal(err)
	}
	if len(recommendResp.Data.Experts) != 1 || recommendResp.Data.Experts[0].UserID != 3 || recommendResp.Data.Experts[0].RoleType != "expert" {
		t.Fatalf("expected expert recommendation: %s", string(recommendBody))
	}
	if recommendResp.Data.Title == "" || recommendResp.Data.EmptyText == "" || recommendResp.Data.SummaryTemplate == "" || len(recommendResp.Data.Categories) == 0 {
		t.Fatalf("expected system recommendation display config: %s", string(recommendBody))
	}

	progressBody := getJSON(t, mux, "/api/app/game-invites/guide-progress", creatorToken, http.StatusOK)
	var progressResp struct {
		Data struct {
			ActiveCount   int `json:"activeCount"`
			ActiveParties []struct {
				ID int64 `json:"id"`
			} `json:"activeParties"`
		} `json:"data"`
	}
	if err := json.Unmarshal(progressBody, &progressResp); err != nil {
		t.Fatal(err)
	}
	if progressResp.Data.ActiveCount != 1 || len(progressResp.Data.ActiveParties) != 1 || progressResp.Data.ActiveParties[0].ID != inviteResp.Data.Invitation.ID {
		t.Fatalf("expected active invite progress: %s", string(progressBody))
	}

	referralBody := getJSON(t, mux, "/api/app/game-invites/referral-records", creatorToken, http.StatusOK)
	var referralResp struct {
		Data struct {
			Total      int `json:"total"`
			PageConfig struct {
				PageTitle string            `json:"pageTitle"`
				Texts     map[string]string `json:"texts"`
			} `json:"pageConfig"`
			Tabs []struct {
				Key   string `json:"key"`
				Count int    `json:"count"`
			} `json:"tabs"`
			Records []struct {
				InvitationID int64  `json:"invitationId"`
				GameID       int64  `json:"gameId"`
				State        string `json:"state"`
				Actions      []struct {
					Key string `json:"key"`
				} `json:"actions"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(referralBody, &referralResp); err != nil {
		t.Fatal(err)
	}
	if referralResp.Data.Total != 1 || len(referralResp.Data.Records) != 1 || referralResp.Data.Records[0].InvitationID != inviteResp.Data.Invitation.ID || referralResp.Data.Records[0].GameID != 1 || referralResp.Data.Records[0].State != "processing" || len(referralResp.Data.Records[0].Actions) != 2 {
		t.Fatalf("expected referral record from invitation: %s", string(referralBody))
	}
	if referralResp.Data.PageConfig.PageTitle == "" || referralResp.Data.PageConfig.Texts["emptyText"] == "" || len(referralResp.Data.Tabs) == 0 {
		t.Fatalf("expected referral records page config: %s", string(referralBody))
	}

	getJSON(t, mux, "/api/app/games/1/success-detail", expertToken, http.StatusForbidden)
	successDetailBody := getJSON(t, mux, "/api/app/games/1/success-detail", creatorToken, http.StatusOK)
	var successDetailResp struct {
		Data struct {
			Viewer struct {
				UserID int64  `json:"userId"`
				Role   string `json:"role"`
			} `json:"viewer"`
			Group struct {
				GameID int64 `json:"gameId"`
				RoomID int64 `json:"roomId"`
			} `json:"group"`
			Participants []struct {
				UserID    int64  `json:"userId"`
				RoleLabel string `json:"roleLabel"`
			} `json:"participants"`
			Fund struct {
				Status     string `json:"status"`
				AmountText string `json:"amountText"`
			} `json:"fund"`
			NextSteps []struct {
				Action string `json:"action"`
			} `json:"nextSteps"`
			DeliveryProof struct {
				MaxCount         int    `json:"maxCount"`
				EmptyText        string `json:"emptyText"`
				SelectedTemplate string `json:"selectedTemplate"`
				TimelinePrefill  string `json:"timelinePrefill"`
			} `json:"deliveryProof"`
			DeliveryPage struct {
				Paid struct {
					PageTitle    string `json:"pageTitle"`
					ConfirmItems []struct {
						ID string `json:"id"`
					} `json:"confirmItems"`
				} `json:"paid"`
				QuickActions []struct {
					Key string `json:"key"`
				} `json:"quickActions"`
			} `json:"deliveryPage"`
		} `json:"data"`
	}
	if err := json.Unmarshal(successDetailBody, &successDetailResp); err != nil {
		t.Fatal(err)
	}
	if successDetailResp.Data.Viewer.UserID != 1 || successDetailResp.Data.Viewer.Role != "member" || successDetailResp.Data.Group.GameID != 1 || successDetailResp.Data.Group.RoomID == 0 {
		t.Fatalf("expected success detail viewer and group: %s", string(successDetailBody))
	}
	if len(successDetailResp.Data.Participants) != 1 || successDetailResp.Data.Participants[0].UserID != 1 || successDetailResp.Data.Participants[0].RoleLabel != "玩家" {
		t.Fatalf("expected success detail participants: %s", string(successDetailBody))
	}
	if successDetailResp.Data.Fund.Status != "free_no_pay" || successDetailResp.Data.Fund.AmountText == "" || len(successDetailResp.Data.NextSteps) != 3 || successDetailResp.Data.NextSteps[1].Action != "contact_player" {
		t.Fatalf("expected success detail fund and next steps: %s", string(successDetailBody))
	}
	if successDetailResp.Data.DeliveryProof.MaxCount != 4 || successDetailResp.Data.DeliveryProof.EmptyText == "" || successDetailResp.Data.DeliveryProof.SelectedTemplate == "" || successDetailResp.Data.DeliveryProof.TimelinePrefill == "" {
		t.Fatalf("expected delivery proof config: %s", string(successDetailBody))
	}
	if successDetailResp.Data.DeliveryPage.Paid.PageTitle == "" || len(successDetailResp.Data.DeliveryPage.Paid.ConfirmItems) != 3 || len(successDetailResp.Data.DeliveryPage.QuickActions) != 3 {
		t.Fatalf("expected delivery page config: %s", string(successDetailBody))
	}

	postJSON(t, mux, "/api/app/game-invites/reminders", expertToken, `{"invitationId":`+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+`,"message":"outsider"}`, http.StatusNotFound)
	reminderBody := postJSON(t, mux, "/api/app/game-invites/reminders", creatorToken, `{"invitationId":`+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+`,"message":"please confirm"}`, http.StatusOK)
	var reminderResp struct {
		Data struct {
			InvitationID int64 `json:"invitationId"`
			TargetUserID int64 `json:"targetUserId"`
			Notification struct {
				NotifyType string `json:"notifyType"`
				BizType    string `json:"bizType"`
				BizID      int64  `json:"bizId"`
			} `json:"notification"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reminderBody, &reminderResp); err != nil {
		t.Fatal(err)
	}
	if reminderResp.Data.InvitationID != inviteResp.Data.Invitation.ID || reminderResp.Data.TargetUserID != 2 || reminderResp.Data.Notification.NotifyType != "game_invitation_remind" || reminderResp.Data.Notification.BizID != inviteResp.Data.Invitation.ID {
		t.Fatalf("expected reminder notification: %s", string(reminderBody))
	}
	playerNoticesBody := getJSON(t, mux, "/api/app/notifications?type=game_invitation_remind", playerToken, http.StatusOK)
	if countNotificationsByType(t, playerNoticesBody, "game_invitation_remind") != 1 {
		t.Fatalf("expected player reminder notification: %s", string(playerNoticesBody))
	}

	replayContextBody := getJSON(t, mux, "/api/app/game-invites/replay-context?sourceGameId=1", creatorToken, http.StatusOK)
	var replayContextResp struct {
		Data struct {
			SourceGameID int64 `json:"sourceGameId"`
			Invitees     []struct {
				UserID int64 `json:"userId"`
			} `json:"invitees"`
		} `json:"data"`
	}
	if err := json.Unmarshal(replayContextBody, &replayContextResp); err != nil {
		t.Fatal(err)
	}
	if replayContextResp.Data.SourceGameID != 1 || len(replayContextResp.Data.Invitees) != 0 {
		t.Fatalf("expected replay context invitees: %s", string(replayContextBody))
	}

	replayBody := postJSON(t, mux, "/api/app/game-invites/replay", creatorToken, `{"sourceGameId":1,"expertUserIds":[3],"message":"again"}`, http.StatusOK)
	var replayResp struct {
		Data struct {
			ReplayGameID        int64 `json:"replayGameId"`
			PrimaryInvitationID int64 `json:"primaryInvitationId"`
			InvitationCount     int   `json:"invitationCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(replayBody, &replayResp); err != nil {
		t.Fatal(err)
	}
	if replayResp.Data.ReplayGameID == 0 || replayResp.Data.ReplayGameID == 1 || replayResp.Data.PrimaryInvitationID == 0 || replayResp.Data.InvitationCount != 1 {
		t.Fatalf("expected replay invitation created: %s", string(replayBody))
	}

	rejectInviteBody := postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":3,"message":"reject this"}`, http.StatusOK)
	var rejectInviteResp struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rejectInviteBody, &rejectInviteResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(rejectInviteResp.Data.Invitation.ID, 10)+"/respond", expertToken, `{"accept":false,"reason":"busy"}`, http.StatusOK)
	cancelBody := getJSON(t, mux, "/api/app/game-invites/guide-cancel-detail?id="+strconv.FormatInt(rejectInviteResp.Data.Invitation.ID, 10), creatorToken, http.StatusOK)
	var cancelResp struct {
		Data struct {
			ID       int64 `json:"id"`
			Timeline []struct {
				Key string `json:"key"`
			} `json:"timeline"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cancelBody, &cancelResp); err != nil {
		t.Fatal(err)
	}
	if cancelResp.Data.ID != rejectInviteResp.Data.Invitation.ID || len(cancelResp.Data.Timeline) == 0 {
		t.Fatalf("expected cancel detail: %s", string(cancelBody))
	}
}

func TestGuideProgressNotificationCountdownAndDetailsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	if err := server.systemConfig.Set(gameInviteConfigKey, gameInviteConfigDTO{InvitationTimeoutMinutes: 90}); err != nil {
		t.Fatalf("seed invitation timeout failed: %v", err)
	}
	server.Register(mux)

	guideToken := loginForTestWithCode(t, mux, "guide-progress-notification-guide")
	completeIdentityForTest(t, mux, guideToken)
	expertToken := loginForTestWithCode(t, mux, "guide-progress-notification-expert")
	completeIdentityForTest(t, mux, expertToken)
	server.profiles.GrantRole(2, "expert")

	postJSON(t, mux, "/api/app/games", guideToken, `{"title":"进度详情测试局","gameType":"free","cityName":"杭州","address":"西湖","startTime":"2026-07-15T14:00:00+08:00","endTime":"2026-07-15T15:30:00+08:00","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", guideToken, `{}`, http.StatusOK)
	replayBody := postJSON(t, mux, "/api/app/game-invites/replay", guideToken, `{"sourceGameId":1,"expertUserIds":[2],"message":"请确认方案","serviceType":"产品咨询","serviceDuration":"90分钟","demandDetail":"梳理产品方案","budgetAmountCent":128800,"expectedTime":"2026-07-15 14:00"}`, http.StatusOK)
	var replayResp struct {
		Data struct {
			ReplayGameID        int64 `json:"replayGameId"`
			PrimaryInvitationID int64 `json:"primaryInvitationId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(replayBody, &replayResp); err != nil {
		t.Fatal(err)
	}
	if replayResp.Data.ReplayGameID <= 0 || replayResp.Data.PrimaryInvitationID <= 0 {
		t.Fatalf("expected replay invitation: %s", string(replayBody))
	}

	guideNotices := getJSON(t, mux, "/api/app/notifications?type=game_invitation", guideToken, http.StatusOK)
	if countNotificationsByType(t, guideNotices, "game_invitation") != 1 || !strings.Contains(string(guideNotices), `"title":"组局动态"`) || !strings.Contains(string(guideNotices), `"bizId":`+strconv.FormatInt(replayResp.Data.PrimaryInvitationID, 10)) {
		t.Fatalf("expected guide progress notification: %s", string(guideNotices))
	}

	progressBody := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(replayResp.Data.PrimaryInvitationID, 10), guideToken, http.StatusOK)
	for _, expected := range []string{
		`"timeoutSeconds":5400`,
		`"label":"服务类型","value":"产品咨询"`,
		`"label":"咨询时长","value":"90分钟"`,
		`"label":"预算金额","value":"¥1288.00"`,
		`"label":"预计时间","value":"2026-07-15 14:00"`,
		`"badgeIcon":"https://static.haowan.net.cn/miniprogram/pages/game/guide-progress-detail/assets/participant-waiting.svg"`,
		`"iconSrc":"https://static.haowan.net.cn/miniprogram/pages/game/referral-record/assets/action-bell-blue.svg"`,
	} {
		if !strings.Contains(string(progressBody), expected) {
			t.Fatalf("expected %s in guide progress payload: %s", expected, string(progressBody))
		}
	}
}

func TestCurrentGameInviteKeepsOriginalGameAndNotifiesInviteesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "current-game-invite-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "current-game-invite-player")
	completeIdentityForTest(t, mux, playerToken)
	expertToken := loginForTestWithCode(t, mux, "current-game-invite-expert")
	completeIdentityForTest(t, mux, expertToken)

	game, err := server.games.Create(1, games.CreateRequest{Title: "Current game", GameType: "free", MinPlayers: 5, MaxPlayers: 8, StartAt: "2026-07-12 14:00", EndAt: "2026-07-12 16:00"})
	if err != nil {
		t.Fatalf("seed current game: %v", err)
	}
	if _, err := server.games.ApproveGame(game.ID); err != nil {
		t.Fatalf("approve current game: %v", err)
	}

	body := postJSON(t, mux, "/api/app/game-invites/current", creatorToken, `{"sourceGameId":1,"playerUserIds":[2],"expertUserIds":[3],"message":"Join this game"}`, http.StatusOK)
	var response struct {
		Data struct {
			GameID          int64 `json:"gameId"`
			InvitationCount int   `json:"invitationCount"`
			Invitations     []struct {
				ID           int64  `json:"id"`
				GameID       int64  `json:"gameId"`
				TargetUserID int64  `json:"targetUserId"`
				Role         string `json:"role"`
			} `json:"invitations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.GameID != 1 || response.Data.InvitationCount != 2 || len(response.Data.Invitations) != 2 {
		t.Fatalf("expected two invitations for original game: %s", string(body))
	}
	invitationByUser := map[int64]int64{}
	for _, invitation := range response.Data.Invitations {
		if invitation.GameID != 1 {
			t.Fatalf("invitation created for wrong game: %s", string(body))
		}
		invitationByUser[invitation.TargetUserID] = invitation.ID
	}
	if invitationByUser[2] <= 0 || invitationByUser[3] <= 0 {
		t.Fatalf("missing player or expert invitation: %s", string(body))
	}

	playerNotices := getJSON(t, mux, "/api/app/notifications?type=game_invitation", playerToken, http.StatusOK)
	expertNotices := getJSON(t, mux, "/api/app/notifications?type=game_invitation", expertToken, http.StatusOK)
	if countNotificationsByType(t, playerNotices, "game_invitation") != 1 || countNotificationsByType(t, expertNotices, "game_invitation") != 1 {
		t.Fatalf("expected both invitees to receive notifications: player=%s expert=%s", string(playerNotices), string(expertNotices))
	}

	playerResponse := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(invitationByUser[2], 10)+"/respond", playerToken, `{"accept":true}`, http.StatusOK)
	expertResponse := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(invitationByUser[3], 10)+"/respond", expertToken, `{"accept":true}`, http.StatusOK)
	if !strings.Contains(string(playerResponse), `"status":"approved"`) || !strings.Contains(string(expertResponse), `"status":"approved"`) {
		t.Fatalf("expected accepted grouped invitations to join directly: player=%s expert=%s", string(playerResponse), string(expertResponse))
	}
	detail := getJSON(t, mux, "/api/app/games/1", creatorToken, http.StatusOK)
	if !strings.Contains(string(detail), `"currentPlayers":3`) {
		t.Fatalf("expected original game to contain creator, player, and expert: %s", string(detail))
	}
}

func TestGuideProgressRejectedPlayerAndExpertUsesBackendPayload(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	guideToken := loginForTestWithCode(t, mux, "guide-progress-rejected-guide")
	completeIdentityForTest(t, mux, guideToken)
	playerToken := loginForTestWithCode(t, mux, "guide-progress-rejected-player")
	completeIdentityForTest(t, mux, playerToken)
	expertToken := loginForTestWithCode(t, mux, "guide-progress-rejected-expert")
	completeIdentityForTest(t, mux, expertToken)
	outsiderToken := loginForTestWithCode(t, mux, "guide-progress-rejected-outsider")
	completeIdentityForTest(t, mux, outsiderToken)
	server.profiles.GrantRole(3, "expert")

	postJSON(t, mux, "/api/app/games", guideToken, `{"title":"backend rejected progress","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", guideToken, `{}`, http.StatusOK)

	createInvitation := func(targetUserID int64, roleType string, message string) int64 {
		body := postJSON(t, mux, "/api/app/games/1/guide-invitations", guideToken, `{"targetUserId":`+strconv.FormatInt(targetUserID, 10)+`,"roleType":"`+roleType+`","message":"`+message+`"}`, http.StatusOK)
		var response struct {
			Data struct {
				Invitation struct {
					ID int64 `json:"id"`
				} `json:"invitation"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		if response.Data.Invitation.ID == 0 {
			t.Fatalf("expected invitation id: %s", string(body))
		}
		return response.Data.Invitation.ID
	}

	playerInvitationID := createInvitation(2, "player", "玩家邀请后台说明")
	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(playerInvitationID, 10)+"/respond", playerToken, `{"accept":false,"reason":"玩家时间冲突"}`, http.StatusOK)
	expertInvitationID := createInvitation(3, "expert", "行家邀请后台说明")
	postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(expertInvitationID, 10)+"/respond", expertToken, `{"accept":false,"reason":"行家档期冲突"}`, http.StatusOK)

	progressBody := getJSON(t, mux, "/api/app/game-invites/guide-progress?gameId=1", guideToken, http.StatusOK)
	var progressResp struct {
		Data struct {
			PageTexts        map[string]string `json:"pageTexts"`
			CompletedParties []struct {
				ID            int64  `json:"id"`
				GameID        int64  `json:"gameId"`
				TargetRole    string `json:"targetRole"`
				Title         string `json:"title"`
				StatusTitle   string `json:"statusTitle"`
				DetailRoute   string `json:"detailRoute"`
				IsCanceled    bool   `json:"isCanceled"`
				RejectName    string `json:"rejectName"`
				RejectRole    string `json:"rejectRole"`
				RejectText    string `json:"rejectText"`
				Reason        string `json:"reason"`
				DetailDisplay struct {
					PageTitle      string `json:"pageTitle"`
					ConfirmText    string `json:"confirmText"`
					ShowActionBar  bool   `json:"showActionBar"`
					ReviewReadonly string `json:"reviewReadonlyText"`
				} `json:"detailDisplay"`
				Player struct {
					Status     string `json:"status"`
					StatusText string `json:"statusText"`
					StateClass string `json:"stateClass"`
				} `json:"player"`
				Expert struct {
					Status     string `json:"status"`
					StatusText string `json:"statusText"`
					StateClass string `json:"stateClass"`
				} `json:"expert"`
				Steps []struct {
					Key   string `json:"key"`
					Title string `json:"title"`
				} `json:"steps"`
			} `json:"completedParties"`
		} `json:"data"`
	}
	if err := json.Unmarshal(progressBody, &progressResp); err != nil {
		t.Fatal(err)
	}
	if progressResp.Data.PageTexts["pageTitle"] == "" || progressResp.Data.PageTexts["timelineTitle"] == "" || len(progressResp.Data.CompletedParties) != 2 {
		t.Fatalf("expected backend page texts and two rejected invitations: %s", string(progressBody))
	}
	if !strings.Contains(string(progressBody), `"cancelRoute":"/pages/game/guide-cancel/index?`) || !strings.Contains(string(progressBody), `"chatPageTitle":"小程序通知"`) || !strings.Contains(string(progressBody), `"chatDeclineButtonText":"婉拒"`) {
		t.Fatalf("expected backend-owned action routes and chat texts: %s", string(progressBody))
	}
	byRole := make(map[string]struct {
		ID          int64
		Title       string
		StatusTitle string
		DetailRoute string
		IsCanceled  bool
		RejectName  string
		RejectRole  string
		RejectText  string
		Reason      string
		Status      string
		StatusText  string
		StateClass  string
		StepTitle   string
		PageTitle   string
		ConfirmText string
		Readonly    string
	})
	for _, item := range progressResp.Data.CompletedParties {
		status, statusText, stateClass := item.Player.Status, item.Player.StatusText, item.Player.StateClass
		stepTitle := ""
		stepKey := "player"
		if item.TargetRole == "expert" {
			status, statusText, stateClass = item.Expert.Status, item.Expert.StatusText, item.Expert.StateClass
			stepKey = "expert"
		}
		for _, step := range item.Steps {
			if step.Key == stepKey {
				stepTitle = step.Title
			}
		}
		byRole[item.TargetRole] = struct {
			ID          int64
			Title       string
			StatusTitle string
			DetailRoute string
			IsCanceled  bool
			RejectName  string
			RejectRole  string
			RejectText  string
			Reason      string
			Status      string
			StatusText  string
			StateClass  string
			StepTitle   string
			PageTitle   string
			ConfirmText string
			Readonly    string
		}{item.ID, item.Title, item.StatusTitle, item.DetailRoute, item.IsCanceled, item.RejectName, item.RejectRole, item.RejectText, item.Reason, status, statusText, stateClass, stepTitle, item.DetailDisplay.PageTitle, item.DetailDisplay.ConfirmText, item.DetailDisplay.ReviewReadonly}
	}
	playerItem := byRole["player"]
	if playerItem.ID != playerInvitationID || playerItem.Title != "组局已取消" || playerItem.StatusTitle != "组局已取消" || !playerItem.IsCanceled || playerItem.RejectName == "" || playerItem.RejectRole != "玩家" || playerItem.RejectText != "已拒绝邀请" || playerItem.Reason != "玩家邀请后台说明" || playerItem.Status != "rejected" || playerItem.StatusText != "已拒绝" || playerItem.StateClass != "canceled" || playerItem.StepTitle != "玩家已拒绝" || playerItem.PageTitle != "玩家确认组局" || playerItem.ConfirmText != "确认参加" || playerItem.Readonly != "已拒绝" || !strings.Contains(playerItem.DetailRoute, "invitationId=") {
		t.Fatalf("unexpected player rejected payload: %s", string(progressBody))
	}
	expertItem := byRole["expert"]
	if expertItem.ID != expertInvitationID || expertItem.Title != "组局已取消" || expertItem.StatusTitle != "组局已取消" || !expertItem.IsCanceled || expertItem.RejectName == "" || expertItem.RejectRole != "行家" || expertItem.RejectText != "已拒绝邀请" || expertItem.Reason != "行家邀请后台说明" || expertItem.Status != "rejected" || expertItem.StatusText != "已拒绝" || expertItem.StateClass != "canceled" || expertItem.StepTitle != "行家已拒绝" || expertItem.PageTitle != "行家审核组局" || expertItem.ConfirmText != "确认通过" || expertItem.Readonly != "已拒绝" || !strings.Contains(expertItem.DetailRoute, "invitationId=") {
		t.Fatalf("unexpected expert rejected payload: %s", string(progressBody))
	}

	filteredBody := getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(expertInvitationID, 10), guideToken, http.StatusOK)
	if !strings.Contains(string(filteredBody), `"targetRole":"expert"`) || !strings.Contains(string(filteredBody), `"行家已拒绝"`) {
		t.Fatalf("expected filtered expert rejection detail: %s", string(filteredBody))
	}
	getJSON(t, mux, "/api/app/game-invites/guide-progress?invitationId="+strconv.FormatInt(expertInvitationID, 10), outsiderToken, http.StatusNotFound)
}

func TestManagedGameActionTextsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "managed-game-action-texts-creator")
	completeIdentityForTest(t, mux, creatorToken)
	guideToken := loginForTestWithCode(t, mux, "managed-game-action-texts-guide")
	completeIdentityForTest(t, mux, guideToken)
	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"action text game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	inviteBody := postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":2,"message":"lead action text game"}`, http.StatusOK)
	var inviteResp struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(inviteBody, &inviteResp); err != nil {
		t.Fatal(err)
	}
	respondBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+"/respond", guideToken, `{"accept":true,"reason":"ok"}`, http.StatusOK)
	var respondResp struct {
		Data struct {
			Application struct {
				ID int64 `json:"id"`
			} `json:"application"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respondBody, &respondResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(respondResp.Data.Application.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)

	body := getJSON(t, mux, "/api/app/games/my/manage", guideToken, http.StatusOK)
	var response struct {
		Data struct {
			Orders []struct {
				PrimaryActionText   string `json:"primaryActionText"`
				SecondaryActionText string `json:"secondaryActionText"`
				PlayerActionText    string `json:"playerActionText"`
				GuideActionText     string `json:"guideActionText"`
			} `json:"orders"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data.Orders) != 1 {
		t.Fatalf("expected one managed order: %s", string(body))
	}
	order := response.Data.Orders[0]
	if order.PrimaryActionText != "提前结束交付" || order.SecondaryActionText != "取消并赔付" || order.PlayerActionText != "联系玩家" || order.GuideActionText != "联系领路人" {
		t.Fatalf("unexpected managed order action texts: %s", string(body))
	}
}

func TestPlayerGameActionTextsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "player-action-texts-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "player-action-texts-player")
	completeIdentityForTest(t, mux, playerToken)
	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"player action text game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	applicationBody := postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	var application struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(applicationBody, &application); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(application.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)

	body := getJSON(t, mux, "/api/app/games/player/manage", playerToken, http.StatusOK)
	var response struct {
		Data struct {
			Orders []struct {
				PrimaryActionText   string `json:"primaryActionText"`
				SecondaryActionText string `json:"secondaryActionText"`
				NoticeText          string `json:"noticeText"`
			} `json:"orders"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Data.Orders) != 1 {
		t.Fatalf("expected one player order: %s", string(body))
	}
	order := response.Data.Orders[0]
	if order.PrimaryActionText != "暂无行家" || order.SecondaryActionText != "申请取消" || order.NoticeText != "本局暂未分配行家，无法联系行家。" {
		t.Fatalf("unexpected player order action texts: %s", string(body))
	}
}

func TestSystemProfileAvatarUsesOwnedFileAndReturnsURL(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	token := loginForTestWithCode(t, mux, "system-profile-avatar")
	userID := currentUserIDForTest(t, mux, token)
	completeIdentityForTest(t, mux, token)
	otherToken := loginForTestWithCode(t, mux, "system-profile-avatar-other")
	completeIdentityForTest(t, mux, otherToken)

	putJSON(t, mux, "/api/app/users/me/profile", token, `{"nickname":"敏感词"}`, http.StatusUnavailableForLegalReasons)
	putJSON(t, mux, "/api/app/profile/system-management/profile-info", token, `{"personalInfo":{"name":"敏感词"}}`, http.StatusUnavailableForLegalReasons)

	type uploadResponse struct {
		Data struct {
			File struct {
				ID int64 `json:"fileId"`
			} `json:"file"`
		} `json:"data"`
	}
	uploadAvatar := func(ownerToken, fileName string) uploadResponse {
		t.Helper()
		body := postJSON(t, mux, "/api/app/files/upload-token", ownerToken, `{"bizType":"avatar","fileName":"`+fileName+`","mimeType":"image/png","size":128}`, http.StatusOK)
		var resp uploadResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Data.File.ID == 0 {
			t.Fatalf("expected uploaded avatar file: %s", string(body))
		}
		return resp
	}

	otherAvatar := uploadAvatar(otherToken, "other.png")
	putJSON(t, mux, "/api/app/profile/system-management/profile-info", token, `{"personalInfo":{"name":"Alice","avatarFileId":`+strconv.FormatInt(otherAvatar.Data.File.ID, 10)+`}}`, http.StatusForbidden)

	ownedAvatar := uploadAvatar(token, "owned.png")
	body := putJSON(t, mux, "/api/app/profile/system-management/profile-info", token, `{"personalInfo":{"name":"Alice","avatarFileId":`+strconv.FormatInt(ownedAvatar.Data.File.ID, 10)+`}}`, http.StatusOK)
	var saved struct {
		Data struct {
			PersonalInfo struct {
				AvatarFileID        int64  `json:"avatarFileId"`
				AvatarURL           string `json:"avatarUrl"`
				PendingAvatarFileID int64  `json:"pendingAvatarFileId"`
				PendingAvatarURL    string `json:"pendingAvatarUrl"`
				AvatarAuditStatus   string `json:"avatarAuditStatus"`
			} `json:"personalInfo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Data.PersonalInfo.PendingAvatarFileID != ownedAvatar.Data.File.ID || saved.Data.PersonalInfo.PendingAvatarURL == "" || saved.Data.PersonalInfo.AvatarAuditStatus != "pending" {
		t.Fatalf("expected pending system profile avatar audit: %s", string(body))
	}
	if saved.Data.PersonalInfo.AvatarFileID != 0 || saved.Data.PersonalInfo.AvatarURL != "" {
		t.Fatalf("expected avatar to wait for admin review before taking effect: %s", string(body))
	}

	body = getJSON(t, mux, "/api/app/profile/system-management/profile-info", token, http.StatusOK)
	if err := json.Unmarshal(body, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Data.PersonalInfo.PendingAvatarFileID != ownedAvatar.Data.File.ID || saved.Data.PersonalInfo.PendingAvatarURL == "" || saved.Data.PersonalInfo.AvatarAuditStatus != "pending" {
		t.Fatalf("expected loaded pending system profile avatar audit: %s", string(body))
	}

	homeBody := getJSON(t, mux, "/api/app/profile/home", token, http.StatusOK)
	var homeResp struct {
		Data struct {
			User struct {
				AvatarFileID int64  `json:"avatarFileId"`
				AvatarURL    string `json:"avatarUrl"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(homeBody, &homeResp); err != nil {
		t.Fatal(err)
	}
	if homeResp.Data.User.AvatarFileID != 0 || homeResp.Data.User.AvatarURL != "" {
		t.Fatalf("expected profile home avatar to remain unchanged before review: %s", string(homeBody))
	}

	adminToken := adminLoginForTest(t, mux)
	listBody := getAdminJSON(t, mux, "/api/admin/avatar-audits", adminToken, http.StatusOK)
	if !strings.Contains(string(listBody), `"status":"pending"`) {
		t.Fatalf("expected pending avatar audit in admin list: %s", string(listBody))
	}
	postAdminJSON(t, mux, "/api/admin/avatar-audits/"+strconv.FormatInt(userID, 10)+"/review", adminToken, `{"approve":true,"reason":"ok"}`, http.StatusOK)

	homeBody = getJSON(t, mux, "/api/app/profile/home", token, http.StatusOK)
	if err := json.Unmarshal(homeBody, &homeResp); err != nil {
		t.Fatal(err)
	}
	if homeResp.Data.User.AvatarFileID != ownedAvatar.Data.File.ID || homeResp.Data.User.AvatarURL == "" {
		t.Fatalf("expected profile home to reuse approved avatar: %s", string(homeBody))
	}
}

func TestProfileSystemManagementHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)
	token := loginForTestWithCode(t, mux, "profile-system")
	completeIdentityForTest(t, mux, token)
	server.profiles.GrantRole(1, "expert")
	if _, err := server.profiles.UpdateExpertSkill(1, profiles.ExpertSkillRequest{SkillTree: []string{"strategy"}, ServiceTags: []string{"board-game"}, CaseFileIDs: []int64{1}}); err != nil {
		t.Fatalf("seed expert skill failed: %v", err)
	}

	profileBody := getJSON(t, mux, "/api/app/profile/system-management/profile-info", token, http.StatusOK)
	var profileResp struct {
		Data struct {
			PersonalInfo struct {
				Name string `json:"name"`
			} `json:"personalInfo"`
			Certifications []struct {
				Key         string `json:"key"`
				StatusClass string `json:"statusClass"`
			} `json:"certifications"`
			VisibilityOptions []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"visibilityOptions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(profileBody, &profileResp); err != nil {
		t.Fatal(err)
	}
	if profileResp.Data.PersonalInfo.Name == "" || len(profileResp.Data.Certifications) == 0 || profileResp.Data.Certifications[0].StatusClass != "verified" || len(profileResp.Data.VisibilityOptions) != 3 {
		t.Fatalf("expected default profile info from user and identity: %s", string(profileBody))
	}

	putJSON(t, mux, "/api/app/profile/system-management/profile-info", token, `{"personalInfo":{"name":"Alice","avatarText":"AL","phoneMasked":"138****0001","contactVisibility":"member","hobby":"board games"},"enterpriseInfo":{"company":"Acme","jobTitle":"PM","businessCountText":"2","resources":"venue","publicBusinessInfo":true},"certifications":[{"key":"personal","status":"ok","statusClass":"verified"}]}`, http.StatusOK)
	profileBody = getJSON(t, mux, "/api/app/profile/system-management/profile-info", token, http.StatusOK)
	if err := json.Unmarshal(profileBody, &profileResp); err != nil {
		t.Fatal(err)
	}
	if profileResp.Data.PersonalInfo.Name != "Alice" {
		t.Fatalf("expected saved profile info, got: %s", string(profileBody))
	}
	userBody := getJSON(t, mux, "/api/app/users/me", token, http.StatusOK)
	if !strings.Contains(string(userBody), `"nickname":"Alice"`) {
		t.Fatalf("expected profile save to sync nickname: %s", string(userBody))
	}

	skillBody := getJSON(t, mux, "/api/app/profile/system-management/skill-config", token, http.StatusOK)
	var skillResp struct {
		Data struct {
			RoleSummary struct {
				ConfiguredCount int `json:"configuredCount"`
			} `json:"roleSummary"`
			SkillGroups map[string][]map[string]interface{} `json:"skillGroups"`
		} `json:"data"`
	}
	if err := json.Unmarshal(skillBody, &skillResp); err != nil {
		t.Fatal(err)
	}
	if skillResp.Data.RoleSummary.ConfiguredCount == 0 || len(skillResp.Data.SkillGroups["visible"]) == 0 {
		t.Fatalf("expected default skill config from expert profile: %s", string(skillBody))
	}

	putJSON(t, mux, "/api/app/profile/system-management/skill-config", token, `{"activeTab":"hidden","roleSummary":{"roleName":"expert","maxSkillCount":3,"monthlyLimit":3,"usedCount":1,"remainingCount":2,"configuredCount":1},"skillSlots":[{"id":"custom","title":"custom","empty":false}],"skillGroups":{"visible":[{"id":"custom","title":"custom"}],"hidden":[],"cases":[]},"unlockSuggestion":{"title":"done"}}`, http.StatusOK)
	skillBody = getJSON(t, mux, "/api/app/profile/system-management/skill-config", token, http.StatusOK)
	if err := json.Unmarshal(skillBody, &skillResp); err != nil {
		t.Fatal(err)
	}
	if skillResp.Data.RoleSummary.ConfiguredCount != 1 || len(skillResp.Data.SkillGroups["visible"]) != 1 {
		t.Fatalf("expected saved skill config, got: %s", string(skillBody))
	}

	putJSON(t, mux, "/api/app/profile/system-management/skill-config", token, `{"activeTab":"cases","roleSummary":{"roleName":"expert","maxSkillCount":3,"monthlyLimit":3,"usedCount":1,"remainingCount":2,"configuredCount":1},"skillSlots":[{"id":"case-custom","title":"custom","empty":false}],"skillGroups":{"visible":[],"hidden":[],"cases":[{"id":"case-custom","title":"custom case","caseTitle":"真实服务案例","caseDesc":"后端返回的服务案例详情","caseDate":"2026-06-30","casePlayers":"5人局","rating":4.8,"iconText":"★","tone":"blue","sourceText":"历史组局服务","badge":"已绑定","tags":["交付稳定"],"players":[{"name":"真实玩家A","desc":"参与玩家"}]},{"id":"case-empty","title":"empty case","caseTitle":"空参与案例","caseDate":"2026-06-30","casePlayers":"待绑定","rating":4.2,"iconText":"★","tone":"blue"}]},"unlockSuggestion":{"title":"done"}}`, http.StatusOK)
	caseBody := getJSON(t, mux, "/api/app/profile/system-management/service-cases/case-custom", token, http.StatusOK)
	var caseResp struct {
		Data struct {
			CaseInfo struct {
				Title       string `json:"title"`
				PlayersText string `json:"playersText"`
			} `json:"caseInfo"`
			Rating struct {
				Score string   `json:"score"`
				Tags  []string `json:"tags"`
			} `json:"rating"`
			DetailSections []struct {
				Text string `json:"text"`
			} `json:"detailSections"`
			Players []struct {
				Name string `json:"name"`
			} `json:"players"`
		} `json:"data"`
	}
	if err := json.Unmarshal(caseBody, &caseResp); err != nil {
		t.Fatal(err)
	}
	if caseResp.Data.CaseInfo.Title != "真实服务案例" || caseResp.Data.CaseInfo.PlayersText != "5人局" || caseResp.Data.Rating.Score != "4.8" || len(caseResp.Data.Rating.Tags) == 0 || len(caseResp.Data.DetailSections) == 0 || !strings.Contains(caseResp.Data.DetailSections[0].Text, "后端返回") || len(caseResp.Data.Players) != 1 || caseResp.Data.Players[0].Name != "真实玩家A" {
		t.Fatalf("expected backend service case detail, got: %s", string(caseBody))
	}
	emptyCaseBody := getJSON(t, mux, "/api/app/profile/system-management/service-cases/case-empty", token, http.StatusOK)
	if err := json.Unmarshal(emptyCaseBody, &caseResp); err != nil {
		t.Fatal(err)
	}
	if len(caseResp.Data.Players) != 0 || len(caseResp.Data.Rating.Tags) != 0 {
		t.Fatalf("expected empty service case to avoid synthetic players and tags: %s", string(emptyCaseBody))
	}

	feedbackHomeBody := getJSON(t, mux, "/api/app/profile/system-management/feedback", token, http.StatusOK)
	var feedbackHomeResp struct {
		Data struct {
			FeedbackTypes []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"feedbackTypes"`
			Limits struct {
				ContentMaxLength int    `json:"contentMaxLength"`
				FileMaxCount     int    `json:"fileMaxCount"`
				UploadNote       string `json:"uploadNote"`
			} `json:"limits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(feedbackHomeBody, &feedbackHomeResp); err != nil {
		t.Fatal(err)
	}
	if len(feedbackHomeResp.Data.FeedbackTypes) == 0 || feedbackHomeResp.Data.Limits.ContentMaxLength != 500 || feedbackHomeResp.Data.Limits.FileMaxCount != 9 || feedbackHomeResp.Data.Limits.UploadNote == "" {
		t.Fatalf("expected feedback home config, got: %s", string(feedbackHomeBody))
	}

	feedbackBody := postJSON(t, mux, "/api/app/profile/system-management/feedback", token, `{"typeKey":"problem","sessionKey":"general","content":"cannot open map","contact":"13800000000","fileIds":[1],"quick":true}`, http.StatusOK)
	var feedbackResp struct {
		Data struct {
			Record struct {
				ID          string `json:"id"`
				TypeKey     string `json:"typeKey"`
				StatusClass string `json:"statusClass"`
				Content     string `json:"content"`
			} `json:"record"`
			SuccessPage struct {
				SuccessTitle string `json:"successTitle"`
				Reward       struct {
					Value string `json:"value"`
				} `json:"reward"`
				Rating struct {
					ScoreFrom int `json:"scoreFrom"`
					ScoreTo   int `json:"scoreTo"`
				} `json:"rating"`
				Reasons []struct {
					Value string `json:"value"`
				} `json:"reasons"`
			} `json:"successPage"`
		} `json:"data"`
	}
	if err := json.Unmarshal(feedbackBody, &feedbackResp); err != nil {
		t.Fatal(err)
	}
	if feedbackResp.Data.Record.ID == "" || feedbackResp.Data.Record.TypeKey != "problem" || feedbackResp.Data.Record.StatusClass != "pending" {
		t.Fatalf("expected submitted feedback record, got: %s", string(feedbackBody))
	}
	if feedbackResp.Data.SuccessPage.SuccessTitle == "" || feedbackResp.Data.SuccessPage.Reward.Value == "" || feedbackResp.Data.SuccessPage.Rating.ScoreTo != 10 || len(feedbackResp.Data.SuccessPage.Reasons) == 0 {
		t.Fatalf("expected feedback success page config, got: %s", string(feedbackBody))
	}
	postJSON(t, mux, "/api/app/profile/system-management/feedback", token, `{"typeKey":"problem","content":""}`, http.StatusUnprocessableEntity)
	attachmentFeedbackBody := postJSON(t, mux, "/api/app/profile/system-management/feedback", token, `{"typeKey":"problem","sessionKey":"general","content":"","fileIds":[2]}`, http.StatusOK)
	var attachmentFeedbackResp struct {
		Data struct {
			Record struct {
				ID      string `json:"id"`
				Content string `json:"content"`
			} `json:"record"`
		} `json:"data"`
	}
	if err := json.Unmarshal(attachmentFeedbackBody, &attachmentFeedbackResp); err != nil {
		t.Fatal(err)
	}
	if attachmentFeedbackResp.Data.Record.ID == "" || attachmentFeedbackResp.Data.Record.Content != "附件反馈" {
		t.Fatalf("expected attachment-only feedback, got: %s", string(attachmentFeedbackBody))
	}
	messageBody := postJSON(t, mux, "/api/app/profile/system-management/feedback-records/"+attachmentFeedbackResp.Data.Record.ID+"/messages", token, `{"content":"","fileIds":[3]}`, http.StatusOK)
	var messageResp struct {
		Data struct {
			Message struct {
				Content string  `json:"content"`
				FileIDs []int64 `json:"fileIds"`
			} `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(messageBody, &messageResp); err != nil {
		t.Fatal(err)
	}
	if messageResp.Data.Message.Content != "附件补充" || len(messageResp.Data.Message.FileIDs) != 1 {
		t.Fatalf("expected attachment-only feedback message, got: %s", string(messageBody))
	}

	feedbackRecordsBody := getJSON(t, mux, "/api/app/profile/system-management/feedback-records?tab=processing", token, http.StatusOK)
	var feedbackRecordsResp struct {
		Data struct {
			ActiveTab string `json:"activeTab"`
			Tabs      []struct {
				Key   string `json:"key"`
				Count int    `json:"count"`
			} `json:"tabs"`
			Records []struct {
				ID          string `json:"id"`
				StatusClass string `json:"statusClass"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(feedbackRecordsBody, &feedbackRecordsResp); err != nil {
		t.Fatal(err)
	}
	hasFeedbackRecord := false
	for _, record := range feedbackRecordsResp.Data.Records {
		if record.ID == feedbackResp.Data.Record.ID {
			hasFeedbackRecord = true
			break
		}
	}
	if feedbackRecordsResp.Data.ActiveTab != "processing" || len(feedbackRecordsResp.Data.Records) == 0 || !hasFeedbackRecord {
		t.Fatalf("expected processing feedback records, got: %s", string(feedbackRecordsBody))
	}

	adminFeedbackBody := getAdminJSONWithPermission(t, mux, "/api/admin/feedback-records", "feedback:view", http.StatusOK)
	var adminFeedbackResp struct {
		Data struct {
			Items []struct {
				ID     string `json:"id"`
				UserID int64  `json:"userId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminFeedbackBody, &adminFeedbackResp); err != nil {
		t.Fatal(err)
	}
	userID := currentUserIDForTest(t, mux, token)
	hasAdminFeedbackRecord := false
	for _, record := range adminFeedbackResp.Data.Items {
		if record.ID == feedbackResp.Data.Record.ID && record.UserID == userID {
			hasAdminFeedbackRecord = true
			break
		}
	}
	if !hasAdminFeedbackRecord {
		t.Fatalf("expected admin feedback list to include submitted record, got: %s", string(adminFeedbackBody))
	}
	adminReplyBody := postAdminJSONWithPermission(t, mux, "/api/admin/feedback-records/"+feedbackResp.Data.Record.ID+"/reply", "feedback:reply", `{"userId":`+strconv.FormatInt(userID, 10)+`,"content":"map issue accepted","status":"processing"}`, http.StatusOK)
	var adminReplyResp struct {
		Data struct {
			Record struct {
				StatusClass string `json:"statusClass"`
				ReplyText   string `json:"replyText"`
			} `json:"record"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminReplyBody, &adminReplyResp); err != nil {
		t.Fatal(err)
	}
	if adminReplyResp.Data.Record.StatusClass != "processing" || adminReplyResp.Data.Record.ReplyText != "map issue accepted" || !hasFeedbackServiceMessage(adminReplyResp.Data.Messages, "map issue accepted") {
		t.Fatalf("expected admin reply to persist service message, got: %s", string(adminReplyBody))
	}
	feedbackDetailBody := getJSON(t, mux, "/api/app/profile/system-management/feedback-records/"+feedbackResp.Data.Record.ID, token, http.StatusOK)
	var feedbackDetailResp struct {
		Data struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(feedbackDetailBody, &feedbackDetailResp); err != nil {
		t.Fatal(err)
	}
	if !hasFeedbackServiceMessage(feedbackDetailResp.Data.Messages, "map issue accepted") {
		t.Fatalf("expected app feedback detail to include admin service reply, got: %s", string(feedbackDetailBody))
	}

	blockBody := getJSON(t, mux, "/api/app/profile/system-management/block-settings", token, http.StatusOK)
	var blockResp struct {
		Data struct {
			Enabled  bool     `json:"enabled"`
			Keywords []string `json:"keywords"`
			Scenes   []struct {
				Key  string `json:"key"`
				Mode string `json:"mode"`
			} `json:"scenes"`
			Whitelist       []map[string]interface{} `json:"whitelist"`
			ProtectionMode  string                   `json:"protectionMode"`
			RenewalDays     int                      `json:"renewalDays"`
			ProtectionModes []map[string]interface{} `json:"protectionModes"`
			RenewalOptions  []map[string]interface{} `json:"renewalOptions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(blockBody, &blockResp); err != nil {
		t.Fatal(err)
	}
	if !blockResp.Data.Enabled || len(blockResp.Data.Keywords) == 0 || len(blockResp.Data.Scenes) == 0 || blockResp.Data.ProtectionMode == "" || blockResp.Data.RenewalDays == 0 || len(blockResp.Data.ProtectionModes) == 0 || len(blockResp.Data.RenewalOptions) == 0 {
		t.Fatalf("expected default block settings, got: %s", string(blockBody))
	}

	putJSON(t, mux, "/api/app/profile/system-management/block-settings", token, `{"enabled":false,"keywords":["spam","private"],"whitelist":[{"id":"u2","name":"Bob"}],"scenes":[{"key":"message","mode":"hard","enabled":false}],"protectionMode":"soft","renewalDays":60}`, http.StatusOK)
	blockBody = getJSON(t, mux, "/api/app/profile/system-management/block-settings", token, http.StatusOK)
	if err := json.Unmarshal(blockBody, &blockResp); err != nil {
		t.Fatal(err)
	}
	if blockResp.Data.Enabled || len(blockResp.Data.Keywords) != 2 || blockResp.Data.Keywords[0] != "spam" || len(blockResp.Data.Whitelist) != 1 || len(blockResp.Data.Scenes) != 1 || blockResp.Data.Scenes[0].Key != "message" || blockResp.Data.ProtectionMode != "soft" || blockResp.Data.RenewalDays != 60 {
		t.Fatalf("expected saved block settings, got: %s", string(blockBody))
	}
	putJSON(t, mux, "/api/app/profile/system-management/block-settings", token, `{"keywords":["keep"]}`, http.StatusOK)
	blockBody = getJSON(t, mux, "/api/app/profile/system-management/block-settings", token, http.StatusOK)
	if err := json.Unmarshal(blockBody, &blockResp); err != nil {
		t.Fatal(err)
	}
	if blockResp.Data.ProtectionMode != "soft" || blockResp.Data.RenewalDays != 60 || len(blockResp.Data.Scenes) != 1 || blockResp.Data.Keywords[0] != "keep" {
		t.Fatalf("expected partial block settings update to preserve other pages, got: %s", string(blockBody))
	}

	settingsBody := getJSON(t, mux, "/api/app/profile/settings", token, http.StatusOK)
	var settingsResp struct {
		Data struct {
			Sections []struct {
				Title string `json:"title"`
				Rows  []struct {
					ID      string `json:"id"`
					Enabled bool   `json:"enabled"`
				} `json:"rows"`
			} `json:"sections"`
		} `json:"data"`
	}
	if err := json.Unmarshal(settingsBody, &settingsResp); err != nil {
		t.Fatal(err)
	}
	if len(settingsResp.Data.Sections) == 0 || len(settingsResp.Data.Sections[0].Rows) == 0 {
		t.Fatalf("expected default profile settings, got: %s", string(settingsBody))
	}

	putJSON(t, mux, "/api/app/profile/settings", token, `{"sections":[{"title":"通知设置","rows":[{"id":"gamePush","label":"局消息推送","iconKey":"gamePush","switch":true,"enabled":false}]}]}`, http.StatusOK)
	settingsBody = getJSON(t, mux, "/api/app/profile/settings", token, http.StatusOK)
	if err := json.Unmarshal(settingsBody, &settingsResp); err != nil {
		t.Fatal(err)
	}
	if len(settingsResp.Data.Sections) != 1 || len(settingsResp.Data.Sections[0].Rows) != 1 || settingsResp.Data.Sections[0].Rows[0].Enabled {
		t.Fatalf("expected saved profile settings, got: %s", string(settingsBody))
	}

	agreementsBody := getJSON(t, mux, "/api/app/profile/agreements", token, http.StatusOK)
	var agreementsResp struct {
		Data struct {
			Items []struct {
				Key            string `json:"key"`
				Signed         bool   `json:"signed"`
				Version        string `json:"version"`
				SignedVersion  string `json:"signedVersion"`
				RequiresResign bool   `json:"requiresResign"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(agreementsBody, &agreementsResp); err != nil {
		t.Fatal(err)
	}
	if len(agreementsResp.Data.Items) < 3 || agreementsResp.Data.Items[2].Key != "settlement" || agreementsResp.Data.Items[2].Signed {
		t.Fatalf("expected unsigned settlement agreement, got: %s", string(agreementsBody))
	}
	userID = currentUserIDForTest(t, mux, token)
	server.profiles.SaveSystemManagementConfig(userID, "agreements", map[string]interface{}{"items": []map[string]interface{}{
		{"key": "privacy", "title": "隐私政策", "desc": "新版隐私政策", "signed": true, "signedVersion": "2026-06-30", "version": "2026-07-01", "sections": defaultAgreementSections()},
	}})
	agreementsBody = getJSON(t, mux, "/api/app/profile/agreements", token, http.StatusOK)
	if err := json.Unmarshal(agreementsBody, &agreementsResp); err != nil {
		t.Fatal(err)
	}
	if len(agreementsResp.Data.Items) != 1 || agreementsResp.Data.Items[0].Signed || !agreementsResp.Data.Items[0].RequiresResign {
		t.Fatalf("expected changed agreement to require resign: %s", string(agreementsBody))
	}
	postJSON(t, mux, "/api/app/profile/agreements/privacy/sign", token, `{}`, http.StatusOK)
	agreementsBody = getJSON(t, mux, "/api/app/profile/agreements", token, http.StatusOK)
	if err := json.Unmarshal(agreementsBody, &agreementsResp); err != nil {
		t.Fatal(err)
	}
	if !agreementsResp.Data.Items[0].Signed || agreementsResp.Data.Items[0].RequiresResign || agreementsResp.Data.Items[0].SignedVersion != agreementsResp.Data.Items[0].Version {
		t.Fatalf("expected resign to store current version: %s", string(agreementsBody))
	}
	server.profiles.SaveSystemManagementConfig(userID, "agreements", map[string]interface{}{"items": defaultProfileAgreements()})

	agreementBody := getJSON(t, mux, "/api/app/profile/agreements/settlement", token, http.StatusOK)
	var agreementResp struct {
		Data struct {
			Key             string `json:"key"`
			Signed          bool   `json:"signed"`
			SignedAt        string `json:"signedAt"`
			SignActionText  string `json:"signActionText"`
			SignSuccessText string `json:"signSuccessText"`
			SignConfirm     struct {
				Title       string `json:"title"`
				Desc        string `json:"desc"`
				CancelText  string `json:"cancelText"`
				ConfirmText string `json:"confirmText"`
			} `json:"signConfirm"`
			Sections []struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			} `json:"sections"`
		} `json:"data"`
	}
	if err := json.Unmarshal(agreementBody, &agreementResp); err != nil {
		t.Fatal(err)
	}
	if agreementResp.Data.Key != "settlement" || agreementResp.Data.Signed || len(agreementResp.Data.Sections) == 0 {
		t.Fatalf("expected settlement agreement detail, got: %s", string(agreementBody))
	}
	if agreementResp.Data.SignActionText == "" || agreementResp.Data.SignSuccessText == "" || agreementResp.Data.SignConfirm.Title == "" || agreementResp.Data.SignConfirm.CancelText == "" || agreementResp.Data.SignConfirm.ConfirmText == "" {
		t.Fatalf("expected agreement sign texts from api: %s", string(agreementBody))
	}

	postJSON(t, mux, "/api/app/profile/agreements/settlement/sign", token, `{}`, http.StatusOK)
	agreementBody = getJSON(t, mux, "/api/app/profile/agreements/settlement", token, http.StatusOK)
	if err := json.Unmarshal(agreementBody, &agreementResp); err != nil {
		t.Fatal(err)
	}
	if !agreementResp.Data.Signed || agreementResp.Data.SignedAt == "" {
		t.Fatalf("expected signed settlement agreement, got: %s", string(agreementBody))
	}
}

func TestInvitedMainGuideCanStartAndManageProgressHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "main-guide-creator")
	completeIdentityForTest(t, mux, creatorToken)
	guideToken := loginForTestWithCode(t, mux, "main-guide")
	completeIdentityForTest(t, mux, guideToken)
	memberTokens := []string{
		loginForTestWithCode(t, mux, "main-guide-member-3"),
		loginForTestWithCode(t, mux, "main-guide-member-4"),
		loginForTestWithCode(t, mux, "main-guide-member-5"),
	}
	for _, token := range memberTokens {
		completeIdentityForTest(t, mux, token)
	}

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"main guide game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	inviteBody := postJSON(t, mux, "/api/app/games/1/guide-invitations", creatorToken, `{"targetUserId":2,"message":"lead this game"}`, http.StatusOK)
	var inviteResp struct {
		Data struct {
			Invitation struct {
				ID int64 `json:"id"`
			} `json:"invitation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(inviteBody, &inviteResp); err != nil {
		t.Fatal(err)
	}
	respondBody := postJSON(t, mux, "/api/app/game-invitations/"+strconv.FormatInt(inviteResp.Data.Invitation.ID, 10)+"/respond", guideToken, `{"accept":true,"reason":"ok"}`, http.StatusOK)
	var respondResp struct {
		Data struct {
			Application struct {
				ID int64 `json:"id"`
			} `json:"application"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respondBody, &respondResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(respondResp.Data.Application.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	for index, token := range memberTokens {
		appBody := postJSON(t, mux, "/api/app/games/1/applications", token, `{"reason":"join"}`, http.StatusOK)
		var appResp struct {
			Data struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(appBody, &appResp); err != nil {
			t.Fatal(err)
		}
		if appResp.Data.ID == 0 {
			t.Fatalf("expected application id for member %d: %s", index+3, string(appBody))
		}
		postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(appResp.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	}

	detailBody := getJSON(t, mux, "/api/app/games/1", guideToken, http.StatusOK)
	var detailResp struct {
		Data struct {
			MainGuideUserID int64 `json:"mainGuideUserId"`
			MyRelation      struct {
				Role     string `json:"role"`
				CanStart bool   `json:"canStart"`
			} `json:"myRelation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.MainGuideUserID != 2 || detailResp.Data.MyRelation.Role != "main_guide" || !detailResp.Data.MyRelation.CanStart {
		t.Fatalf("expected invited main guide relation: %s", string(detailBody))
	}
	membersBody := getJSON(t, mux, "/api/app/games/1/members", guideToken, http.StatusOK)
	var membersResp struct {
		Data struct {
			Items []struct {
				UserID int64  `json:"userId"`
				Role   string `json:"role"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(membersBody, &membersResp); err != nil {
		t.Fatal(err)
	}
	foundGuide := false
	for _, item := range membersResp.Data.Items {
		if item.UserID == 2 && item.Role == "main_guide" {
			foundGuide = true
		}
	}
	if !foundGuide {
		t.Fatalf("expected main guide in members: %s", string(membersBody))
	}
	postJSON(t, mux, "/api/app/games/1/manual-start", guideToken, `{}`, http.StatusOK)
	collaborationBody := getJSON(t, mux, "/api/app/games/1/collaboration", guideToken, http.StatusOK)
	var collaborationResp struct {
		Data struct {
			GameID int64  `json:"gameId"`
			Title  string `json:"title"`
		} `json:"data"`
	}
	if err := json.Unmarshal(collaborationBody, &collaborationResp); err != nil {
		t.Fatal(err)
	}
	if collaborationResp.Data.GameID != 1 || collaborationResp.Data.Title != "main guide game" {
		t.Fatalf("expected main guide collaboration detail: %s", string(collaborationBody))
	}
	manageBody := getJSON(t, mux, "/api/app/games/my/manage", guideToken, http.StatusOK)
	if !strings.Contains(string(manageBody), `"main guide game"`) || !strings.Contains(string(manageBody), `"gameId":1`) {
		t.Fatalf("expected main guide to manage invited game: %s", string(manageBody))
	}
	followBody := postJSON(t, mux, "/api/app/games/1/guide-follow-ups", guideToken, `{"action":"feedback"}`, http.StatusOK)
	var followResp struct {
		Data struct {
			Action string `json:"action"`
			Status string `json:"status"`
			Target struct {
				Type string `json:"type"`
			} `json:"target"`
		} `json:"data"`
	}
	if err := json.Unmarshal(followBody, &followResp); err != nil {
		t.Fatal(err)
	}
	if followResp.Data.Action != "feedback" || followResp.Data.Status != "recorded" || followResp.Data.Target.Type != "im_room" {
		t.Fatalf("expected guide follow-up response: %s", string(followBody))
	}
	postJSON(t, mux, "/api/app/games/1/guide-follow-ups", memberTokens[0], `{"action":"feedback"}`, http.StatusForbidden)
	postJSON(t, mux, "/api/app/games/1/guide-follow-ups", guideToken, `{"action":"unknown"}`, http.StatusUnprocessableEntity)
	creatorNoticesBody := getJSON(t, mux, "/api/app/notifications?type=guide_follow_up", creatorToken, http.StatusOK)
	if countNotificationsByType(t, creatorNoticesBody, "guide_follow_up") != 1 {
		t.Fatalf("expected guide follow-up notification: %s", string(creatorNoticesBody))
	}
	postJSON(t, mux, "/api/app/games/1/progress-feedbacks", memberTokens[0], `{"progress":10,"content":"member update"}`, http.StatusForbidden)
	progressBody := postJSON(t, mux, "/api/app/games/1/progress-feedbacks", guideToken, `{"progress":30,"content":"guide update"}`, http.StatusOK)
	var progressResp struct {
		Data struct {
			UserID   int64 `json:"userId"`
			Progress int   `json:"progress"`
		} `json:"data"`
	}
	if err := json.Unmarshal(progressBody, &progressResp); err != nil {
		t.Fatal(err)
	}
	if progressResp.Data.UserID != 2 || progressResp.Data.Progress != 30 {
		t.Fatalf("expected main guide progress: %s", string(progressBody))
	}
}

func TestAdminGameProgressRetrospectiveAndContinueDraftsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "admin-progress-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "admin-progress-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"progress target","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "admin-progress", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	noMilestonePermReq := httptest.NewRequest(http.MethodGet, "/api/admin/games/1/milestones", nil)
	noMilestonePermRec := httptest.NewRecorder()
	mux.ServeHTTP(noMilestonePermRec, noMilestonePermReq)
	if noMilestonePermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin milestones 403 without permission, got %d: %s", noMilestonePermRec.Code, noMilestonePermRec.Body.String())
	}

	milestoneBody := postAdminJSONWithPermission(t, mux, "/api/admin/games/1/milestones", "game:progress:manage", `{"title":"admin stage"}`, http.StatusOK)
	var milestoneResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(milestoneBody, &milestoneResp); err != nil {
		t.Fatal(err)
	}
	if milestoneResp.Data.ID != 1 || milestoneResp.Data.Status != "pending" {
		t.Fatalf("unexpected admin milestone: %s", string(milestoneBody))
	}

	milestonesBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/1/milestones", "game:read", http.StatusOK)
	var milestonesResp struct {
		Data struct {
			Items []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(milestonesBody, &milestonesResp); err != nil {
		t.Fatal(err)
	}
	if len(milestonesResp.Data.Items) != 1 || milestonesResp.Data.Items[0].Title != "admin stage" {
		t.Fatalf("expected admin milestones: %s", string(milestonesBody))
	}

	postJSON(t, mux, "/api/app/games/1/checkins", playerToken, `{"milestoneId":1,"checkinType":"progress","content":"proof","fileIds":[1]}`, http.StatusOK)
	postAdminJSONWithPermission(t, mux, "/api/admin/game-checkins/1/mark-invalid", "game:progress:manage", `{"reason":"invalid evidence"}`, http.StatusOK)
	checkinsBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/1/checkins", "game:read", http.StatusOK)
	var checkinsResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(checkinsBody, &checkinsResp); err != nil {
		t.Fatal(err)
	}
	if len(checkinsResp.Data.Items) != 1 || checkinsResp.Data.Items[0].Status != "invalid" {
		t.Fatalf("expected admin checkins include invalid item: %s", string(checkinsBody))
	}

	postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	postJSON(t, mux, "/api/app/games/1/retrospectives", playerToken, `{"content":"done","againIntent":"yes"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/continue", creatorToken, `{"title":"next progress target"}`, http.StatusOK)

	retroBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/1/retrospectives", "game:read", http.StatusOK)
	var retroResp struct {
		Data struct {
			Items []struct {
				Content     string `json:"content"`
				AgainIntent string `json:"againIntent"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(retroBody, &retroResp); err != nil {
		t.Fatal(err)
	}
	if len(retroResp.Data.Items) != 1 || retroResp.Data.Items[0].Content != "done" || retroResp.Data.Items[0].AgainIntent != "yes" {
		t.Fatalf("expected admin retrospectives: %s", string(retroBody))
	}

	draftsBody := getAdminJSONWithPermission(t, mux, "/api/admin/games/1/continue-drafts", "game:read", http.StatusOK)
	var draftsResp struct {
		Data struct {
			Items []struct {
				OriginalGameID int64  `json:"originalGameId"`
				DraftGameID    int64  `json:"draftGameId"`
				Status         string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(draftsBody, &draftsResp); err != nil {
		t.Fatal(err)
	}
	if len(draftsResp.Data.Items) != 1 || draftsResp.Data.Items[0].OriginalGameID != 1 || draftsResp.Data.Items[0].DraftGameID != 2 || draftsResp.Data.Items[0].Status != "draft" {
		t.Fatalf("expected admin continue drafts: %s", string(draftsBody))
	}
}

func TestIMFlow(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "im-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "im-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"IM game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	approveExtraMembersForHTTP(t, mux, creatorToken, 1, "im-flow", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	postJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, `{"messageType":"text","content":"hello"}`, http.StatusOK)
	creatorIMNotices := getJSON(t, mux, "/api/app/notifications?type=im_message", creatorToken, http.StatusOK)
	if countNotificationsByType(t, creatorIMNotices, "im_message") != 1 ||
		!strings.Contains(string(creatorIMNotices), `"title":"收到局内消息"`) ||
		!strings.Contains(string(creatorIMNotices), `pages/im/room/index?gameId=1`) {
		t.Fatalf("expected creator to receive IM notification: %s", string(creatorIMNotices))
	}
	creatorMessageCenter := getJSON(t, mux, "/api/app/messages/center", creatorToken, http.StatusOK)
	var creatorCenterResp struct {
		Data struct {
			Sections []struct {
				Key       string `json:"key"`
				CountText string `json:"countText"`
				Items     []struct {
					Title    string `json:"title"`
					RouteKey string `json:"routeKey"`
					Unread   bool   `json:"unread"`
				} `json:"items"`
			} `json:"sections"`
		} `json:"data"`
	}
	if err := json.Unmarshal(creatorMessageCenter, &creatorCenterResp); err != nil {
		t.Fatal(err)
	}
	friendCards := []struct {
		Title    string `json:"title"`
		RouteKey string `json:"routeKey"`
		Unread   bool   `json:"unread"`
	}{}
	friendCountText := ""
	for _, section := range creatorCenterResp.Data.Sections {
		if section.Key == "friend" {
			friendCards = section.Items
			friendCountText = section.CountText
			break
		}
	}
	if len(friendCards) != 1 || friendCards[0].Title != "IM gameIM" || friendCards[0].RouteKey != "im_room" || !friendCards[0].Unread || friendCountText != "1/1" {
		t.Fatalf("expected one aggregated IM room card in friend section: %s", string(creatorMessageCenter))
	}
	playerIMNotices := getJSON(t, mux, "/api/app/notifications?type=im_message", playerToken, http.StatusOK)
	if countNotificationsByType(t, playerIMNotices, "im_message") != 0 {
		t.Fatalf("sender must not receive own IM notification: %s", string(playerIMNotices))
	}
	postJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, `{"messageType":"text","content":"敏感词"}`, http.StatusUnavailableForLegalReasons)
	createdWordBody := postAdminJSONWithPermission(t, mux, "/api/admin/sensitive-words", "content:sensitive_word:create", `{"word":"blocked-word","level":"high","action":"block"}`, http.StatusOK)
	var createdWordResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createdWordBody, &createdWordResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, `{"messageType":"text","content":"contains blocked-word"}`, http.StatusUnavailableForLegalReasons)
	putJSON(t, mux, "/api/admin/sensitive-words/"+strconv.FormatInt(createdWordResp.Data.ID, 10), adminLoginForTest(t, mux), `{"status":"disabled"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, `{"messageType":"text","content":"contains blocked-word after disabled"}`, http.StatusOK)
	postAdminJSONWithPermission(t, mux, "/api/admin/sensitive-words", "content:sensitive_word:create", `{"word":"flag-word","level":"medium","action":"flag"}`, http.StatusOK)
	flaggedBody := postJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, `{"messageType":"text","content":"contains flag-word"}`, http.StatusOK)
	var flaggedResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(flaggedBody, &flaggedResp); err != nil {
		t.Fatal(err)
	}
	if flaggedResp.Data.Status != "risk_flagged" {
		t.Fatalf("expected flagged message status: %s", string(flaggedBody))
	}

	riskLogsBody := getAdminJSONWithPermission(t, mux, "/api/admin/content-risk/logs", "content:risk_log:view", http.StatusOK)
	var riskLogsResp struct {
		Data struct {
			Items []struct {
				Status string `json:"status"`
				Word   string `json:"word"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(riskLogsBody, &riskLogsResp); err != nil {
		t.Fatal(err)
	}
	if len(riskLogsResp.Data.Items) == 0 {
		t.Fatalf("expected content risk logs: %s", string(riskLogsBody))
	}
	aiRiskBody := postJSON(t, mux, "/api/internal/ai/content-risk/check-placeholder", "", `{"content":"blocked-word"}`, http.StatusOK)
	var aiRiskResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(aiRiskBody, &aiRiskResp); err != nil {
		t.Fatal(err)
	}
	if aiRiskResp.Data.Status != "risk_flagged" {
		t.Fatalf("expected ai content risk placeholder flag: %s", string(aiRiskBody))
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/chat-session", nil)
	sessionReq.Header.Set("Authorization", "Bearer "+playerToken)
	sessionRec := httptest.NewRecorder()
	mux.ServeHTTP(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("expected chat session 200, got %d: %s", sessionRec.Code, sessionRec.Body.String())
	}

	collaborationBody := getJSON(t, mux, "/api/app/games/1/collaboration", playerToken, http.StatusOK)
	var collaborationResp struct {
		Data struct {
			GameID   int64 `json:"gameId"`
			Progress struct {
				Percent int `json:"percent"`
				Tasks   []struct {
					Title string `json:"title"`
				} `json:"tasks"`
			} `json:"progress"`
			Members []struct {
				UserID   int64  `json:"userId"`
				RoleText string `json:"roleText"`
			} `json:"members"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
			Actions struct {
				CanManageMembers bool   `json:"canManageMembers"`
				EndConfirmRoute  string `json:"endConfirmRoute"`
			} `json:"actions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(collaborationBody, &collaborationResp); err != nil {
		t.Fatal(err)
	}
	hasHelloMessage := false
	for _, message := range collaborationResp.Data.Messages {
		if message.Content == "hello" {
			hasHelloMessage = true
			break
		}
	}
	if collaborationResp.Data.GameID != 1 || collaborationResp.Data.Progress.Percent == 0 || len(collaborationResp.Data.Progress.Tasks) != 3 || len(collaborationResp.Data.Members) != 5 || !hasHelloMessage || collaborationResp.Data.Actions.CanManageMembers || collaborationResp.Data.Actions.EndConfirmRoute == "" {
		t.Fatalf("expected collaboration aggregate for player member: %s", string(collaborationBody))
	}

	roomReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/chat-room", nil)
	roomReq.Header.Set("Authorization", "Bearer "+playerToken)
	roomRec := httptest.NewRecorder()
	mux.ServeHTTP(roomRec, roomReq)
	if roomRec.Code != http.StatusOK {
		t.Fatalf("expected chat room 200, got %d: %s", roomRec.Code, roomRec.Body.String())
	}
	var roomResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(roomRec.Body.Bytes(), &roomResp); err != nil {
		t.Fatal(err)
	}
	if roomResp.Data.ID == 0 {
		t.Fatal("expected room id")
	}

	roomMessageBody := postJSON(t, mux, "/api/app/chat/rooms/1/messages", playerToken, `{"messageType":"text","content":"room message"}`, http.StatusOK)
	var roomMessageResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(roomMessageBody, &roomMessageResp); err != nil {
		t.Fatal(err)
	}
	if roomMessageResp.Data.ID == 0 {
		t.Fatal("expected room message id")
	}
	postJSON(t, mux, "/api/app/chat/rooms/1/messages/2/ack", playerToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/chat/rooms/1/messages/2/read", playerToken, `{}`, http.StatusOK)

	uploadBody := postJSON(t, mux, "/api/app/files/upload-token", playerToken, `{"bizType":"chat_file","objectId":1,"fileName":"a.png","mimeType":"image/png","size":128}`, http.StatusOK)
	var uploadResp struct {
		Data struct {
			Upload struct {
				FileID int64 `json:"fileId"`
			} `json:"upload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResp); err != nil {
		t.Fatal(err)
	}
	if uploadResp.Data.Upload.FileID == 0 {
		t.Fatal("expected file id")
	}
	imageBody := postJSON(t, mux, "/api/app/chat/rooms/1/messages", playerToken, `{"messageType":"image","fileId":1,"content":"a.png"}`, http.StatusOK)
	var imageResp struct {
		Data struct {
			FileID int64 `json:"fileId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(imageBody, &imageResp); err != nil {
		t.Fatal(err)
	}
	if imageResp.Data.FileID != uploadResp.Data.Upload.FileID {
		t.Fatalf("expected chat image message to bind file id: %s", string(imageBody))
	}
	postJSON(t, mux, "/api/app/chat/rooms/1/messages", playerToken, `{"messageType":"image","fileId":999,"content":"bad.png"}`, http.StatusNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/app/games/1/chat/messages", nil)
	req.Header.Set("Authorization", "Bearer "+playerToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected history 200, got %d: %s", rec.Code, rec.Body.String())
	}

	roomHistoryReq := httptest.NewRequest(http.MethodGet, "/api/app/chat/rooms/1/messages", nil)
	roomHistoryReq.Header.Set("Authorization", "Bearer "+playerToken)
	roomHistoryRec := httptest.NewRecorder()
	mux.ServeHTTP(roomHistoryRec, roomHistoryReq)
	if roomHistoryRec.Code != http.StatusOK {
		t.Fatalf("expected room history 200, got %d: %s", roomHistoryRec.Code, roomHistoryRec.Body.String())
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/app/files/1/download-url", nil)
	downloadReq.Header.Set("Authorization", "Bearer "+playerToken)
	downloadRec := httptest.NewRecorder()
	mux.ServeHTTP(downloadRec, downloadReq)
	if downloadRec.Code != http.StatusOK {
		t.Fatalf("expected download url 200, got %d: %s", downloadRec.Code, downloadRec.Body.String())
	}

	outsiderToken := loginForTestWithCode(t, mux, "im-outsider")
	completeIdentityForTest(t, mux, outsiderToken)
	outsiderReq := httptest.NewRequest(http.MethodGet, "/api/app/files/1/download-url", nil)
	outsiderReq.Header.Set("Authorization", "Bearer "+outsiderToken)
	outsiderRec := httptest.NewRecorder()
	mux.ServeHTTP(outsiderRec, outsiderReq)
	if outsiderRec.Code != http.StatusForbidden {
		t.Fatalf("expected outsider download 403, got %d: %s", outsiderRec.Code, outsiderRec.Body.String())
	}
	outsiderSessionReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/chat-session", nil)
	outsiderSessionReq.Header.Set("Authorization", "Bearer "+outsiderToken)
	outsiderSessionRec := httptest.NewRecorder()
	mux.ServeHTTP(outsiderSessionRec, outsiderSessionReq)
	if outsiderSessionRec.Code != http.StatusForbidden {
		t.Fatalf("expected outsider chat session 403, got %d: %s", outsiderSessionRec.Code, outsiderSessionRec.Body.String())
	}
	outsiderHistoryReq := httptest.NewRequest(http.MethodGet, "/api/app/games/1/chat/messages", nil)
	outsiderHistoryReq.Header.Set("Authorization", "Bearer "+outsiderToken)
	outsiderHistoryRec := httptest.NewRecorder()
	mux.ServeHTTP(outsiderHistoryRec, outsiderHistoryReq)
	if outsiderHistoryRec.Code != http.StatusForbidden {
		t.Fatalf("expected outsider game history 403, got %d: %s", outsiderHistoryRec.Code, outsiderHistoryRec.Body.String())
	}
	outsiderRoomHistoryReq := httptest.NewRequest(http.MethodGet, "/api/app/chat/rooms/1/messages", nil)
	outsiderRoomHistoryReq.Header.Set("Authorization", "Bearer "+outsiderToken)
	outsiderRoomHistoryRec := httptest.NewRecorder()
	mux.ServeHTTP(outsiderRoomHistoryRec, outsiderRoomHistoryReq)
	if outsiderRoomHistoryRec.Code != http.StatusForbidden {
		t.Fatalf("expected outsider room history 403, got %d: %s", outsiderRoomHistoryRec.Code, outsiderRoomHistoryRec.Body.String())
	}

	postJSON(t, mux, "/api/app/games/1/completion-request", creatorToken, `{}`, http.StatusOK)
	repeatedCompletionBody := postJSON(t, mux, "/api/app/games/1/completion-request", creatorToken, `{}`, http.StatusOK)
	if !strings.Contains(string(repeatedCompletionBody), `"alreadyRequested":true`) {
		t.Fatalf("expected repeated completion request to be idempotent: %s", string(repeatedCompletionBody))
	}
	completionHistoryBody := getJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, http.StatusOK)
	if count := strings.Count(string(completionHistoryBody), `"messageType":"service_confirm_remind"`); count != 1 {
		t.Fatalf("expected exactly one completion reminder, got %d: %s", count, string(completionHistoryBody))
	}

	archiveBody := postJSON(t, mux, "/api/app/chat/rooms/1/archive", creatorToken, `{"reason":"completed_without_dispute"}`, http.StatusOK)
	var archiveResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(archiveBody, &archiveResp); err != nil {
		t.Fatal(err)
	}
	if archiveResp.Data.Status != "archived" {
		t.Fatalf("expected archived room: %s", string(archiveBody))
	}
	postJSON(t, mux, "/api/app/chat/rooms/1/messages", playerToken, `{"messageType":"text","content":"after archive"}`, http.StatusUnprocessableEntity)

	webhookReq := httptest.NewRequest(http.MethodPost, "/api/internal/openim/webhooks", bytes.NewBufferString(`{"callbackCommand":"callbackAfterSendSingleMsgCommand","groupID":"zhw_game_1"}`))
	webhookRec := httptest.NewRecorder()
	mux.ServeHTTP(webhookRec, webhookReq)
	if webhookRec.Code != http.StatusOK {
		t.Fatalf("expected openim webhook 200, got %d: %s", webhookRec.Code, webhookRec.Body.String())
	}
	var webhookResp struct {
		Data struct {
			Command string `json:"command"`
			GameID  int64  `json:"gameId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(webhookRec.Body.Bytes(), &webhookResp); err != nil {
		t.Fatal(err)
	}
	if webhookResp.Data.Command == "" || webhookResp.Data.GameID != 1 {
		t.Fatalf("expected webhook event record: %s", webhookRec.Body.String())
	}
	if webhookRec.Code != http.StatusOK {
		t.Fatalf("expected webhook 200, got %d: %s", webhookRec.Code, webhookRec.Body.String())
	}
}

func TestIMMessageDTOIncludesShanghaiDisplayTime(t *testing.T) {
	app := newTestAppServer(auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()), identity.NewService())
	createdAt := time.Date(2026, 7, 15, 9, 40, 0, 0, time.UTC)

	inGame := app.inGameMessageDTO(im.Message{
		ID:        1,
		RoomID:    1,
		GameID:    1,
		SenderID:  0,
		Type:      "text",
		Content:   "hello",
		Status:    "sent",
		CreatedAt: createdAt,
	})
	if inGame["createdAtText"] != "2026-07-15 17:40" {
		t.Fatalf("expected in-game message Shanghai display time, got %+v", inGame)
	}

	private := app.privateChatMessageDTO(im.PrivateMessage{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		TargetID:       2,
		Type:           "text",
		Content:        "hello",
		Status:         "sent",
		CreatedAt:      createdAt,
	})
	if private["createdAtText"] != "2026-07-15 17:40" {
		t.Fatalf("expected private message Shanghai display time, got %+v", private)
	}
}

func TestPrivateChatMessagesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	aliceToken := loginForTestWithCode(t, mux, "private-chat-alice")
	completeIdentityForTest(t, mux, aliceToken)
	bobToken := loginForTestWithCode(t, mux, "private-chat-bob")
	completeIdentityForTest(t, mux, bobToken)

	aliceID := currentUserIDForTest(t, mux, aliceToken)
	bobID := currentUserIDForTest(t, mux, bobToken)
	if aliceID == 0 || bobID == 0 || aliceID == bobID {
		t.Fatalf("expected distinct private chat users, alice=%d bob=%d", aliceID, bobID)
	}

	sendBody := postJSON(t, mux, "/api/app/private-chat/messages", aliceToken, `{"targetUserId":`+strconv.FormatInt(bobID, 10)+`,"sourceGameId":9,"messageType":"text","content":"你好，想咨询这个局"}`, http.StatusOK)
	var sendResp struct {
		Data struct {
			Message struct {
				ID           int64  `json:"id"`
				SenderUserID int64  `json:"senderUserId"`
				TargetUserID int64  `json:"targetUserId"`
				Content      string `json:"content"`
			} `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(sendBody, &sendResp); err != nil {
		t.Fatal(err)
	}
	if sendResp.Data.Message.ID == 0 || sendResp.Data.Message.SenderUserID != aliceID || sendResp.Data.Message.TargetUserID != bobID || sendResp.Data.Message.Content != "你好，想咨询这个局" {
		t.Fatalf("expected sent private message: %s", string(sendBody))
	}
	bobNotices := getJSON(t, mux, "/api/app/notifications?type=private_message", bobToken, http.StatusOK)
	if countNotificationsByType(t, bobNotices, "private_message") != 1 ||
		!strings.Contains(string(bobNotices), `"title":"收到私聊消息"`) ||
		!strings.Contains(string(bobNotices), `targetUserId=`+strconv.FormatInt(aliceID, 10)) {
		t.Fatalf("expected bob to receive private chat notification: %s", string(bobNotices))
	}
	aliceNotices := getJSON(t, mux, "/api/app/notifications?type=private_message", aliceToken, http.StatusOK)
	if countNotificationsByType(t, aliceNotices, "private_message") != 0 {
		t.Fatalf("sender must not receive own private chat notification: %s", string(aliceNotices))
	}

	bobHistory := getJSON(t, mux, "/api/app/private-chat/messages?targetUserId="+strconv.FormatInt(aliceID, 10), bobToken, http.StatusOK)
	var historyResp struct {
		Data struct {
			CurrentUserID int64 `json:"currentUserId"`
			TargetUser    struct {
				UserID int64 `json:"userId"`
			} `json:"targetUser"`
			Items []struct {
				ID           int64  `json:"id"`
				SenderUserID int64  `json:"senderUserId"`
				Content      string `json:"content"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bobHistory, &historyResp); err != nil {
		t.Fatal(err)
	}
	if historyResp.Data.CurrentUserID != bobID || historyResp.Data.TargetUser.UserID != aliceID || len(historyResp.Data.Items) != 1 || historyResp.Data.Items[0].SenderUserID != aliceID || historyResp.Data.Items[0].Content != "你好，想咨询这个局" {
		t.Fatalf("expected bob to read alice private message: %s", string(bobHistory))
	}

	postJSON(t, mux, "/api/app/private-chat/messages", aliceToken, `{"targetUserId":`+strconv.FormatInt(bobID, 10)+`,"messageType":"text","content":"敏感词"}`, http.StatusUnavailableForLegalReasons)
	postJSON(t, mux, "/api/app/private-chat/messages", aliceToken, `{"targetUserId":`+strconv.FormatInt(aliceID, 10)+`,"messageType":"text","content":"self"}`, http.StatusUnprocessableEntity)
}

func TestIMArchiveJobSkipsDisputedRooms(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "im-archive-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "im-archive-player")
	completeIdentityForTest(t, mux, playerToken)

	archiveGameID, archiveRoomID := createPendingReviewGameForIMArchive(t, mux, creatorToken, playerToken, "archive ok")
	disputedGameID, disputedRoomID := createPendingReviewGameForIMArchive(t, mux, creatorToken, playerToken, "archive disputed")
	postJSON(t, mux, "/api/app/reports", playerToken, `{"gameId":`+strconv.FormatInt(disputedGameID, 10)+`,"targetUserId":1,"reportType":"service_dispute","content":"keep evidence"}`, http.StatusOK)

	body := postJSON(t, mux, "/api/internal/im/archive-expired-rooms", "", `{}`, http.StatusOK)
	var resp struct {
		Data struct {
			ArchivedCount int `json:"archivedCount"`
			ArchivedRooms []struct {
				GameID int64  `json:"gameId"`
				Status string `json:"status"`
			} `json:"archivedRooms"`
			SkippedCount int `json:"skippedCount"`
			SkippedRooms []struct {
				GameID int64  `json:"gameId"`
				Reason string `json:"reason"`
			} `json:"skippedRooms"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ArchivedCount != 1 || len(resp.Data.ArchivedRooms) != 1 || resp.Data.ArchivedRooms[0].GameID != archiveGameID || resp.Data.ArchivedRooms[0].Status != "archived" {
		t.Fatalf("expected only undisputed room archived: %s", string(body))
	}
	if resp.Data.SkippedCount != 1 || len(resp.Data.SkippedRooms) != 1 || resp.Data.SkippedRooms[0].GameID != disputedGameID || resp.Data.SkippedRooms[0].Reason != "open_report" {
		t.Fatalf("expected disputed room skipped: %s", string(body))
	}
	postJSON(t, mux, "/api/app/chat/rooms/"+strconv.FormatInt(archiveRoomID, 10)+"/messages", playerToken, `{"messageType":"text","content":"after archive job"}`, http.StatusUnprocessableEntity)
	postJSON(t, mux, "/api/app/chat/rooms/"+strconv.FormatInt(disputedRoomID, 10)+"/messages", playerToken, `{"messageType":"text","content":"after skipped archive"}`, http.StatusUnprocessableEntity)
}

func TestAdminIMRoomsListAndDetail(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "im-admin-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "im-admin-player")
	completeIdentityForTest(t, mux, playerToken)
	gameID, roomID, _ := createActiveGameRoomForIMTest(t, mux, creatorToken, playerToken, "im admin room")

	uploadBody := postJSON(t, mux, "/api/app/files/upload-token", playerToken, `{"bizType":"chat_file","objectId":`+strconv.FormatInt(gameID, 10)+`,"fileName":"proof.pdf","mimeType":"application/pdf","size":128}`, http.StatusOK)
	var uploadResp struct {
		Data struct {
			Upload struct {
				FileID int64 `json:"fileId"`
			} `json:"upload"`
		} `json:"data"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/chat/rooms/"+strconv.FormatInt(roomID, 10)+"/messages", playerToken, `{"messageType":"file","fileId":`+strconv.FormatInt(uploadResp.Data.Upload.FileID, 10)+`,"content":"proof.pdf"}`, http.StatusOK)

	unauthReq := httptest.NewRequest(http.MethodGet, "/api/admin/im/rooms", nil)
	unauthRec := httptest.NewRecorder()
	mux.ServeHTTP(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusForbidden {
		t.Fatalf("expected rooms list 403 without admin token, got %d: %s", unauthRec.Code, unauthRec.Body.String())
	}

	adminToken := adminLoginForTest(t, mux)
	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/im/rooms?status=active", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected rooms list 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
	var listResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				ID               int64   `json:"id"`
				GameID           int64   `json:"gameId"`
				Status           string  `json:"status"`
				MemberIDs        []int64 `json:"memberIds"`
				MessageCount     int     `json:"messageCount"`
				FileMessageCount int     `json:"fileMessageCount"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatal(err)
	}
	if listResp.Data.Total != 1 || len(listResp.Data.Items) != 1 {
		t.Fatalf("expected one im room: %s", listRec.Body.String())
	}
	item := listResp.Data.Items[0]
	if item.ID != roomID || item.GameID != gameID || item.Status != "active" || len(item.MemberIDs) != 5 || item.MessageCount != 3 || item.FileMessageCount != 1 {
		t.Fatalf("unexpected im room summary: %s", listRec.Body.String())
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/admin/im/rooms/"+strconv.FormatInt(roomID, 10), nil)
	detailReq.Header.Set("Authorization", "Bearer "+adminToken)
	detailRec := httptest.NewRecorder()
	mux.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected room detail 200, got %d: %s", detailRec.Code, detailRec.Body.String())
	}
	var detailResp struct {
		Data struct {
			MessageCount     int `json:"messageCount"`
			FileMessageCount int `json:"fileMessageCount"`
			FileMessages     []struct {
				FileID int64  `json:"fileId"`
				Type   string `json:"messageType"`
			} `json:"fileMessages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.MessageCount != 3 || detailResp.Data.FileMessageCount != 1 || len(detailResp.Data.FileMessages) != 1 || detailResp.Data.FileMessages[0].FileID != uploadResp.Data.Upload.FileID {
		t.Fatalf("unexpected room detail file messages: %s", detailRec.Body.String())
	}
}

func TestAdminIMRoomArchiveRetryAndHideMessageHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "im-admin-action-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "im-admin-action-player")
	completeIdentityForTest(t, mux, playerToken)
	_, roomID := createPendingReviewGameForIMArchive(t, mux, creatorToken, playerToken, "im admin actions")

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	postAdminJSON(t, mux, "/api/admin/im/messages/1/hide", operatorToken, `{"reason":"no permission"}`, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	retryBody := postAdminJSON(t, mux, "/api/admin/im/rooms/"+strconv.FormatInt(roomID, 10)+"/retry-create", adminToken, `{}`, http.StatusOK)
	var retryResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
			Engine string `json:"engine"`
		} `json:"data"`
	}
	if err := json.Unmarshal(retryBody, &retryResp); err != nil {
		t.Fatal(err)
	}
	if retryResp.Data.ID != roomID || retryResp.Data.Status != "active" || retryResp.Data.Engine == "" {
		t.Fatalf("expected retry-create room response: %s", string(retryBody))
	}

	hideBody := postAdminJSON(t, mux, "/api/admin/im/messages/1/hide", adminToken, `{"reason":"violation"}`, http.StatusOK)
	var hideResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(hideBody, &hideResp); err != nil {
		t.Fatal(err)
	}
	if hideResp.Data.ID != 1 || hideResp.Data.Status != "hidden" {
		t.Fatalf("expected hidden message: %s", string(hideBody))
	}

	historyBody := getJSON(t, mux, "/api/app/chat/rooms/"+strconv.FormatInt(roomID, 10)+"/messages", playerToken, http.StatusOK)
	var historyResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(historyBody, &historyResp); err != nil {
		t.Fatal(err)
	}
	for _, item := range historyResp.Data.Items {
		if item.ID == 1 || item.Status == "hidden" {
			t.Fatalf("expected hidden message filtered from app history: %s", string(historyBody))
		}
	}

	detailBody := getAdminJSON(t, mux, "/api/admin/im/rooms/"+strconv.FormatInt(roomID, 10), adminToken, http.StatusOK)
	var detailResp struct {
		Data struct {
			Messages []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detailResp); err != nil {
		t.Fatal(err)
	}
	foundHidden := false
	for _, item := range detailResp.Data.Messages {
		if item.ID == 1 && item.Status == "hidden" {
			foundHidden = true
		}
	}
	if !foundHidden {
		t.Fatalf("expected admin room detail to keep hidden message as evidence: %s", string(detailBody))
	}

	archiveBody := postAdminJSON(t, mux, "/api/admin/im/rooms/"+strconv.FormatInt(roomID, 10)+"/archive", adminToken, `{"reason":"admin close"}`, http.StatusOK)
	var archiveResp struct {
		Data struct {
			ID            int64  `json:"id"`
			Status        string `json:"status"`
			ArchiveReason string `json:"archiveReason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(archiveBody, &archiveResp); err != nil {
		t.Fatal(err)
	}
	if archiveResp.Data.ID != roomID || archiveResp.Data.Status != "archived" || archiveResp.Data.ArchiveReason != "admin close" {
		t.Fatalf("expected admin archived room: %s", string(archiveBody))
	}
	postJSON(t, mux, "/api/app/chat/rooms/"+strconv.FormatInt(roomID, 10)+"/messages", playerToken, `{"messageType":"text","content":"after admin archive"}`, http.StatusUnprocessableEntity)
}

func createPendingReviewGameForIMArchive(t *testing.T, mux *http.ServeMux, creatorToken string, playerToken string, title string) (int64, int64) {
	t.Helper()
	gameID, roomID, extraTokens := createActiveGameRoomForIMTest(t, mux, creatorToken, playerToken, title)
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	return gameID, roomID
}

func createActiveGameRoomForIMTest(t *testing.T, mux *http.ServeMux, creatorToken string, playerToken string, title string) (int64, int64, []string) {
	t.Helper()
	gameBody := postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"`+title+`","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	var gameResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(gameBody, &gameResp); err != nil {
		t.Fatal(err)
	}
	gameID := gameResp.Data.ID
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/approve-local", creatorToken, `{}`, http.StatusOK)
	applicationBody := postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	var applicationResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(applicationBody, &applicationResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/games/applications/"+strconv.FormatInt(applicationResp.Data.ID, 10)+"/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, gameID, "im-archive-"+strconv.FormatInt(gameID, 10), 3)
	postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/manual-start", creatorToken, `{}`, http.StatusOK)
	roomBody := getJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/chat-room", playerToken, http.StatusOK)
	var roomResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(roomBody, &roomResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/chat/rooms/"+strconv.FormatInt(roomResp.Data.ID, 10)+"/messages", playerToken, `{"messageType":"text","content":"before archive"}`, http.StatusOK)
	return gameID, roomResp.Data.ID, extraTokens
}

func TestServiceConfirmAndReviewFlow(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "review-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "review-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"review game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "service-review", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	postJSON(t, mux, "/api/app/reviews", creatorToken, `{"gameId":1,"targetUserId":2,"score":5,"againIntent":"yes"}`, http.StatusConflict)
	creatorConfirmBody := postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"note":"done","fileIds":[101,102],"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	var creatorConfirmResp struct {
		Data struct {
			Items []struct {
				UserID  int64   `json:"userId"`
				FileIDs []int64 `json:"fileIds"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(creatorConfirmBody, &creatorConfirmResp); err != nil {
		t.Fatal(err)
	}
	if !hasHTTPConfirmItemFileIDs(creatorConfirmResp.Data.Items, 1, []int64{101, 102}) {
		t.Fatalf("expected service confirm proof file ids: %s", string(creatorConfirmBody))
	}
	confirmBody := postJSON(t, mux, "/api/app/games/1/service-confirm-items", playerToken, `{"note":"confirmed","confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		confirmBody = postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"note":"confirmed","confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	var confirmed struct {
		Data struct {
			Game struct {
				Status string `json:"status"`
			} `json:"game"`
		} `json:"data"`
	}
	if err := json.Unmarshal(confirmBody, &confirmed); err != nil {
		t.Fatal(err)
	}
	if confirmed.Data.Game.Status != "pending_review" {
		t.Fatalf("expected pending_review, got %s", confirmed.Data.Game.Status)
	}
	growthAfterConfirmBody := getJSON(t, mux, "/api/app/users/me/growth", creatorToken, http.StatusOK)
	var growthAfterConfirmResp struct {
		Data struct {
			Profile struct {
				Experience int `json:"experience"`
			} `json:"profile"`
			Footprints []struct {
				Action string `json:"action"`
			} `json:"footprints"`
		} `json:"data"`
	}
	if err := json.Unmarshal(growthAfterConfirmBody, &growthAfterConfirmResp); err != nil {
		t.Fatal(err)
	}
	if growthAfterConfirmResp.Data.Profile.Experience != 10 || !hasFootprintAction(growthAfterConfirmResp.Data.Footprints, "completed_game") {
		t.Fatalf("expected completed game growth after service confirm: %s", string(growthAfterConfirmBody))
	}
	creatorNoticesAfterConfirmBody := getJSON(t, mux, "/api/app/notifications", creatorToken, http.StatusOK)
	playerNoticesAfterConfirmBody := getJSON(t, mux, "/api/app/notifications", playerToken, http.StatusOK)
	if countNotificationsByType(t, creatorNoticesAfterConfirmBody, "review_remind") != 1 {
		t.Fatalf("expected creator review reminder after service confirm: %s", string(creatorNoticesAfterConfirmBody))
	}
	if countNotificationsByType(t, playerNoticesAfterConfirmBody, "review_remind") != 1 {
		t.Fatalf("expected player review reminder after service confirm: %s", string(playerNoticesAfterConfirmBody))
	}
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", playerToken, `{"note":"duplicate","confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	growthAfterDuplicateConfirmBody := getJSON(t, mux, "/api/app/users/me/growth", creatorToken, http.StatusOK)
	var growthAfterDuplicateConfirmResp struct {
		Data struct {
			Profile struct {
				Experience int `json:"experience"`
			} `json:"profile"`
		} `json:"data"`
	}
	if err := json.Unmarshal(growthAfterDuplicateConfirmBody, &growthAfterDuplicateConfirmResp); err != nil {
		t.Fatal(err)
	}
	if growthAfterDuplicateConfirmResp.Data.Profile.Experience != 10 {
		t.Fatalf("expected completed game growth to be idempotent: %s", string(growthAfterDuplicateConfirmBody))
	}
	creatorNoticesAfterDuplicateBody := getJSON(t, mux, "/api/app/notifications", creatorToken, http.StatusOK)
	if countNotificationsByType(t, creatorNoticesAfterDuplicateBody, "review_remind") != 1 {
		t.Fatalf("expected duplicate service confirm not to duplicate review reminder: %s", string(creatorNoticesAfterDuplicateBody))
	}

	playerManageBody := getJSON(t, mux, "/api/app/games/player/manage", playerToken, http.StatusOK)
	var playerManageResp struct {
		Data struct {
			Items []struct {
				GameID             int64  `json:"gameId"`
				ExpertID           int64  `json:"expertId"`
				StatusType         string `json:"statusType"`
				ReviewStatus       string `json:"reviewStatus"`
				CanReview          bool   `json:"canReview"`
				ContactExpertRoute string `json:"contactExpertRoute"`
				PlayerCancelRoute  string `json:"playerCancelRoute"`
				ReviewRoute        string `json:"reviewRoute"`
				Actions            struct {
					ContactExpertRoute string `json:"contactExpertRoute"`
					PlayerCancelRoute  string `json:"playerCancelRoute"`
					ReviewRoute        string `json:"reviewRoute"`
				} `json:"actions"`
			} `json:"items"`
			Summary struct {
				CompleteCount int `json:"completeCount"`
			} `json:"summary"`
			PageConfig struct {
				PageTitle    string `json:"pageTitle"`
				CategoryTabs []struct {
					Key  string `json:"key"`
					Text string `json:"text"`
				} `json:"categoryTabs"`
			} `json:"pageConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(playerManageBody, &playerManageResp); err != nil {
		t.Fatal(err)
	}
	if len(playerManageResp.Data.Items) != 1 || playerManageResp.Data.Items[0].GameID != 1 || playerManageResp.Data.Items[0].ExpertID != 0 || playerManageResp.Data.Items[0].StatusType != "complete" || playerManageResp.Data.Items[0].ReviewStatus != "pending" || !playerManageResp.Data.Items[0].CanReview || playerManageResp.Data.Summary.CompleteCount != 1 {
		t.Fatalf("expected player manage service order to expose review target: %s", string(playerManageBody))
	}
	if playerManageResp.Data.Items[0].ContactExpertRoute != "" || !strings.Contains(playerManageResp.Data.Items[0].PlayerCancelRoute, "/pages/game/player-cancel/index") || !strings.Contains(playerManageResp.Data.Items[0].PlayerCancelRoute, "freeCancel=1") || !strings.Contains(playerManageResp.Data.Items[0].PlayerCancelRoute, "warningTitle=") || !strings.Contains(playerManageResp.Data.Items[0].ReviewRoute, "/pages/game/review/index") {
		t.Fatalf("expected player manage action routes: %s", string(playerManageBody))
	}
	if playerManageResp.Data.Items[0].Actions.ContactExpertRoute != "" || playerManageResp.Data.Items[0].Actions.PlayerCancelRoute == "" || playerManageResp.Data.Items[0].Actions.ReviewRoute == "" {
		t.Fatalf("expected player manage actions payload: %s", string(playerManageBody))
	}
	if playerManageResp.Data.PageConfig.PageTitle == "" || len(playerManageResp.Data.PageConfig.CategoryTabs) != 3 {
		t.Fatalf("expected my games page config: %s", string(playerManageBody))
	}
	managedBody := getJSON(t, mux, "/api/app/games/my/manage", creatorToken, http.StatusOK)
	var managedResp struct {
		Data struct {
			Items []struct {
				GameID             int64  `json:"gameId"`
				PlayerID           int64  `json:"playerId"`
				StatusType         string `json:"statusType"`
				ReviewStatus       string `json:"reviewStatus"`
				CanReview          bool   `json:"canReview"`
				DeliveryRoute      string `json:"deliveryRoute"`
				ExpertCancelRoute  string `json:"expertCancelRoute"`
				ContactPlayerRoute string `json:"contactPlayerRoute"`
				ContactGuideRoute  string `json:"contactGuideRoute"`
				ReviewRoute        string `json:"reviewRoute"`
				Actions            struct {
					FinishDeliveryRoute string `json:"finishDeliveryRoute"`
					ExpertCancelRoute   string `json:"expertCancelRoute"`
					ContactPlayerRoute  string `json:"contactPlayerRoute"`
					ContactGuideRoute   string `json:"contactGuideRoute"`
					ReviewRoute         string `json:"reviewRoute"`
				} `json:"actions"`
			} `json:"items"`
			PageConfig struct {
				StatusTabs []struct {
					Key  string `json:"key"`
					Text string `json:"text"`
				} `json:"statusTabs"`
			} `json:"pageConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(managedBody, &managedResp); err != nil {
		t.Fatal(err)
	}
	if len(managedResp.Data.Items) != 0 {
		t.Fatalf("expected player-role creator to stay out of expert managed orders: %s", string(managedBody))
	}
	if len(managedResp.Data.PageConfig.StatusTabs) != 5 {
		t.Fatalf("expected my games status tabs config: %s", string(managedBody))
	}

	todosReq := httptest.NewRequest(http.MethodGet, "/api/app/reviews/todos", nil)
	todosReq.Header.Set("Authorization", "Bearer "+creatorToken)
	todosRec := httptest.NewRecorder()
	mux.ServeHTTP(todosRec, todosReq)
	if todosRec.Code != http.StatusOK {
		t.Fatalf("expected review todos 200, got %d: %s", todosRec.Code, todosRec.Body.String())
	}
	availableReq := httptest.NewRequest(http.MethodGet, "/api/app/reviews/available", nil)
	availableReq.Header.Set("Authorization", "Bearer "+creatorToken)
	availableRec := httptest.NewRecorder()
	mux.ServeHTTP(availableRec, availableReq)
	if availableRec.Code != http.StatusOK {
		t.Fatalf("expected review available 200, got %d: %s", availableRec.Code, availableRec.Body.String())
	}
	var availableResp struct {
		Data struct {
			Items []struct {
				TargetUserID int64  `json:"targetUserId"`
				TargetName   string `json:"targetName"`
				AvatarText   string `json:"avatarText"`
			} `json:"items"`
			Reward struct {
				Show bool `json:"show"`
			} `json:"reward"`
			ReviewPage struct {
				NavTitle            string `json:"navTitle"`
				SatisfactionOptions []struct {
					ID string `json:"id"`
				} `json:"satisfactionOptions"`
				RoleConfigs map[string]struct {
					TagTitle string   `json:"tagTitle"`
					Tags     []string `json:"tags"`
				} `json:"roleConfigs"`
			} `json:"reviewPage"`
		} `json:"data"`
	}
	if err := json.Unmarshal(availableRec.Body.Bytes(), &availableResp); err != nil {
		t.Fatal(err)
	}
	if availableResp.Data.Reward.Show {
		t.Fatalf("expected default review reward hidden in phase one: %s", availableRec.Body.String())
	}
	if availableResp.Data.ReviewPage.NavTitle == "" || len(availableResp.Data.ReviewPage.SatisfactionOptions) == 0 || len(availableResp.Data.ReviewPage.RoleConfigs["expert"].Tags) == 0 {
		t.Fatalf("expected review page config: %s", availableRec.Body.String())
	}
	if len(availableResp.Data.Items) == 0 || availableResp.Data.Items[0].TargetName == "" || availableResp.Data.Items[0].AvatarText == "" {
		t.Fatalf("expected review todo identity display fields: %s", availableRec.Body.String())
	}

	reviewBody := postJSON(t, mux, "/api/app/reviews", creatorToken, `{"gameId":1,"targetUserId":2,"targetRole":"member","score":5,"content":"good","tags":["fun","organized"],"againIntent":"yes","npsScore":9}`, http.StatusOK)
	var reviewResp struct {
		Data struct {
			Review struct {
				Tags        []string `json:"tags"`
				AgainIntent string   `json:"againIntent"`
				NPSScore    *int     `json:"npsScore"`
			} `json:"review"`
			Profile struct {
				Experience      int `json:"experience"`
				AvailablePoints int `json:"availablePoints"`
			} `json:"profile"`
			Reward struct {
				Points       int `json:"points"`
				BeforePoints int `json:"beforePoints"`
				AfterPoints  int `json:"afterPoints"`
				LogID        int `json:"logId"`
			} `json:"reward"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reviewBody, &reviewResp); err != nil {
		t.Fatal(err)
	}
	if reviewResp.Data.Profile.Experience == 0 || reviewResp.Data.Profile.AvailablePoints == 0 {
		t.Fatalf("expected growth and points after review: %s", string(reviewBody))
	}
	if reviewResp.Data.Reward.Points != reviews.SubmittedReviewPoints || reviewResp.Data.Reward.AfterPoints-reviewResp.Data.Reward.BeforePoints != reviews.SubmittedReviewPoints || reviewResp.Data.Reward.LogID == 0 {
		t.Fatalf("expected response reward to match persisted point change: %s", string(reviewBody))
	}
	if reviewResp.Data.Profile.AvailablePoints != reviewResp.Data.Reward.AfterPoints {
		t.Fatalf("expected review reward to be credited exactly once: %s", string(reviewBody))
	}
	if len(reviewResp.Data.Review.Tags) != 2 || reviewResp.Data.Review.AgainIntent != "yes" || reviewResp.Data.Review.NPSScore == nil || *reviewResp.Data.Review.NPSScore != 9 {
		t.Fatalf("expected review tags and again intent: %s", string(reviewBody))
	}
	managedAfterFirstReviewBody := getJSON(t, mux, "/api/app/games/player/manage", creatorToken, http.StatusOK)
	var managedAfterFirstReviewResp struct {
		Data struct {
			Items []struct {
				ReviewStatus string `json:"reviewStatus"`
				CanReview    bool   `json:"canReview"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(managedAfterFirstReviewBody, &managedAfterFirstReviewResp); err != nil {
		t.Fatal(err)
	}
	if len(managedAfterFirstReviewResp.Data.Items) != 1 || managedAfterFirstReviewResp.Data.Items[0].ReviewStatus != "pending" || !managedAfterFirstReviewResp.Data.Items[0].CanReview {
		t.Fatalf("expected remaining member reviews to keep manage entry pending: %s", string(managedAfterFirstReviewBody))
	}
	receivedReviewsBody := getJSON(t, mux, "/api/app/profile/service-center/reviews", playerToken, http.StatusOK)
	var receivedReviewsResp struct {
		Data struct {
			PendingCount int `json:"pendingCount"`
			Reviews      []struct {
				ID             string `json:"id"`
				GameID         int64  `json:"gameId"`
				ReviewerUserID int64  `json:"reviewerUserId"`
				Reply          string `json:"reply"`
				StatusType     string `json:"statusType"`
				DetailURL      string `json:"detailUrl"`
				ReportURL      string `json:"reportUrl"`
			} `json:"reviews"`
		} `json:"data"`
	}
	if err := json.Unmarshal(receivedReviewsBody, &receivedReviewsResp); err != nil {
		t.Fatal(err)
	}
	if receivedReviewsResp.Data.PendingCount != 1 || len(receivedReviewsResp.Data.Reviews) != 1 || receivedReviewsResp.Data.Reviews[0].StatusType != "pending" {
		t.Fatalf("expected one pending received review: %s", string(receivedReviewsBody))
	}
	if receivedReviewsResp.Data.Reviews[0].GameID != 1 || receivedReviewsResp.Data.Reviews[0].ReviewerUserID == 0 || receivedReviewsResp.Data.Reviews[0].DetailURL == "" || !strings.Contains(receivedReviewsResp.Data.Reviews[0].ReportURL, "reviewId=") {
		t.Fatalf("expected received review action urls: %s", string(receivedReviewsBody))
	}
	reviewID := receivedReviewsResp.Data.Reviews[0].ID
	reviewDetailBody := getJSON(t, mux, "/api/app/profile/service-center/reviews/"+reviewID, playerToken, http.StatusOK)
	var reviewDetailResp struct {
		Data struct {
			Review struct {
				ID    string `json:"id"`
				Reply string `json:"reply"`
			} `json:"review"`
			Templates []string `json:"templates"`
			History   []struct {
				Side string `json:"side"`
			} `json:"history"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reviewDetailBody, &reviewDetailResp); err != nil {
		t.Fatal(err)
	}
	if reviewDetailResp.Data.Review.ID != reviewID || len(reviewDetailResp.Data.Templates) == 0 || len(reviewDetailResp.Data.History) != 1 {
		t.Fatalf("expected received review detail with templates and history: %s", string(reviewDetailBody))
	}
	likeBody := postJSON(t, mux, "/api/app/profile/service-center/reviews/"+reviewID+"/like", playerToken, `{}`, http.StatusOK)
	var likeResp struct {
		Data struct {
			ReviewID string `json:"reviewId"`
			Liked    bool   `json:"liked"`
		} `json:"data"`
	}
	if err := json.Unmarshal(likeBody, &likeResp); err != nil {
		t.Fatal(err)
	}
	if likeResp.Data.ReviewID != reviewID || !likeResp.Data.Liked {
		t.Fatalf("expected review like response: %s", string(likeBody))
	}
	unlikeBody := postJSON(t, mux, "/api/app/profile/service-center/reviews/"+reviewID+"/like", playerToken, `{}`, http.StatusOK)
	if err := json.Unmarshal(unlikeBody, &likeResp); err != nil {
		t.Fatal(err)
	}
	if likeResp.Data.ReviewID != reviewID || likeResp.Data.Liked {
		t.Fatalf("expected review unlike response: %s", string(unlikeBody))
	}
	actionsBody := postJSON(t, mux, "/api/app/profile/service-center/reviews/"+reviewID+"/actions", playerToken, `{}`, http.StatusOK)
	var actionsResp struct {
		Data struct {
			Actions []struct {
				Key   string `json:"key"`
				Route string `json:"route"`
			} `json:"actions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(actionsBody, &actionsResp); err != nil {
		t.Fatal(err)
	}
	if len(actionsResp.Data.Actions) < 2 || actionsResp.Data.Actions[0].Route == "" {
		t.Fatalf("expected review action menu: %s", string(actionsBody))
	}
	postJSON(t, mux, "/api/app/profile/service-center/reviews/"+reviewID+"/reply", creatorToken, `{"content":"not owner"}`, http.StatusNotFound)
	replyBody := postJSON(t, mux, "/api/app/profile/service-center/reviews/"+reviewID+"/reply", playerToken, `{"content":"thanks for feedback"}`, http.StatusOK)
	var replyResp struct {
		Data struct {
			StatusType string `json:"statusType"`
			Reply      struct {
				Content string `json:"content"`
			} `json:"reply"`
		} `json:"data"`
	}
	if err := json.Unmarshal(replyBody, &replyResp); err != nil {
		t.Fatal(err)
	}
	if replyResp.Data.StatusType != "replied" || replyResp.Data.Reply.Content != "thanks for feedback" {
		t.Fatalf("expected reply status: %s", string(replyBody))
	}
	receivedAfterReplyBody := getJSON(t, mux, "/api/app/profile/service-center/reviews", playerToken, http.StatusOK)
	var receivedAfterReplyResp struct {
		Data struct {
			PendingCount int `json:"pendingCount"`
			Reviews      []struct {
				Reply      string `json:"reply"`
				StatusType string `json:"statusType"`
			} `json:"reviews"`
		} `json:"data"`
	}
	if err := json.Unmarshal(receivedAfterReplyBody, &receivedAfterReplyResp); err != nil {
		t.Fatal(err)
	}
	if receivedAfterReplyResp.Data.PendingCount != 0 || len(receivedAfterReplyResp.Data.Reviews) != 1 || receivedAfterReplyResp.Data.Reviews[0].Reply != "thanks for feedback" || receivedAfterReplyResp.Data.Reviews[0].StatusType != "replied" {
		t.Fatalf("expected received review replied state: %s", string(receivedAfterReplyBody))
	}
	for _, todo := range availableResp.Data.Items {
		if todo.TargetUserID == 2 {
			continue
		}
		postJSON(t, mux, "/api/app/reviews", creatorToken, fmt.Sprintf(`{"gameId":1,"targetUserId":%d,"targetRole":"member","score":5,"npsScore":9}`, todo.TargetUserID), http.StatusOK)
	}
	managedAfterReviewBody := getJSON(t, mux, "/api/app/games/player/manage", creatorToken, http.StatusOK)
	var managedAfterReviewResp struct {
		Data struct {
			Items []struct {
				ReviewStatus string `json:"reviewStatus"`
				CanReview    bool   `json:"canReview"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(managedAfterReviewBody, &managedAfterReviewResp); err != nil {
		t.Fatal(err)
	}
	if len(managedAfterReviewResp.Data.Items) != 1 || managedAfterReviewResp.Data.Items[0].ReviewStatus != "reviewed" || managedAfterReviewResp.Data.Items[0].CanReview {
		t.Fatalf("expected managed service order to be reviewed after submit: %s", string(managedAfterReviewBody))
	}
	intentsBody := getJSON(t, mux, "/api/app/reviews/my-intents", creatorToken, http.StatusOK)
	var intentsResp struct {
		Data struct {
			Items []struct {
				Tags        []string `json:"tags"`
				AgainIntent string   `json:"againIntent"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(intentsBody, &intentsResp); err != nil {
		t.Fatal(err)
	}
	if len(intentsResp.Data.Items) != 1 || intentsResp.Data.Items[0].AgainIntent != "yes" || len(intentsResp.Data.Items[0].Tags) != 2 {
		t.Fatalf("expected my intents to include tags and again intent: %s", string(intentsBody))
	}
	connectionsBody := getJSON(t, mux, "/api/app/connections/my", creatorToken, http.StatusOK)
	var connectionsResp struct {
		Data struct {
			Items []connectionTestItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(connectionsBody, &connectionsResp); err != nil {
		t.Fatal(err)
	}
	if !hasConnectionSource(connectionsResp.Data.Items, "review") || !hasConnectionStrength(connectionsResp.Data.Items, 2, "review", 1) {
		t.Fatalf("expected review to strengthen connection: %s", string(connectionsBody))
	}
	postJSON(t, mux, "/api/app/reviews", creatorToken, `{"gameId":1,"targetUserId":2,"targetRole":"member","score":5}`, http.StatusConflict)

	lowReviewBody := postJSON(t, mux, "/api/app/reviews", extraTokens[0], `{"gameId":1,"targetUserId":2,"targetRole":"member","score":1,"content":"late"}`, http.StatusOK)
	var lowReviewResp struct {
		Data struct {
			Credit struct {
				ChangeValue int    `json:"changeValue"`
				Reason      string `json:"reason"`
			} `json:"credit"`
		} `json:"data"`
	}
	if err := json.Unmarshal(lowReviewBody, &lowReviewResp); err != nil {
		t.Fatal(err)
	}
	if lowReviewResp.Data.Credit.Reason != "low_review" || lowReviewResp.Data.Credit.ChangeValue >= 0 {
		t.Fatalf("expected low score review to deduct credit: %s", string(lowReviewBody))
	}
	playerCreditBody := getJSON(t, mux, "/api/app/profile/credit-center", playerToken, http.StatusOK)
	var playerCreditResp struct {
		Data struct {
			Score   int `json:"score"`
			Records []struct {
				Reason string `json:"reason"`
				Tone   string `json:"tone"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(playerCreditBody, &playerCreditResp); err != nil {
		t.Fatal(err)
	}
	if playerCreditResp.Data.Score >= 100 || len(playerCreditResp.Data.Records) == 0 || playerCreditResp.Data.Records[0].Reason != "low_review" || playerCreditResp.Data.Records[0].Tone != "minus" {
		t.Fatalf("expected low review credit center record: %s", string(playerCreditBody))
	}
	playerLowReviewNoticesBody := getJSON(t, mux, "/api/app/notifications", playerToken, http.StatusOK)
	if countNotificationsByType(t, playerLowReviewNoticesBody, "low_review_credit_deducted") != 1 {
		t.Fatalf("expected low review credit notification: %s", string(playerLowReviewNoticesBody))
	}

	outsiderToken := loginForTestWithCode(t, mux, "review-outsider")
	completeIdentityForTest(t, mux, outsiderToken)
	postJSON(t, mux, "/api/app/reviews", outsiderToken, `{"gameId":1,"targetUserId":1,"score":5}`, http.StatusForbidden)

	growthReq := httptest.NewRequest(http.MethodGet, "/api/app/users/me/growth", nil)
	growthReq.Header.Set("Authorization", "Bearer "+creatorToken)
	growthRec := httptest.NewRecorder()
	mux.ServeHTTP(growthRec, growthReq)
	if growthRec.Code != http.StatusOK {
		t.Fatalf("expected growth 200, got %d: %s", growthRec.Code, growthRec.Body.String())
	}
	var growthResp struct {
		Data struct {
			Profile struct {
				TodayCreditScore int      `json:"todayCreditScore"`
				Achievements     []string `json:"achievements"`
			} `json:"profile"`
			Footprints []struct {
				Action string `json:"action"`
			} `json:"footprints"`
			Achievements []struct {
				Code       string `json:"code"`
				Title      string `json:"title"`
				Category   string `json:"category"`
				StatusText string `json:"statusText"`
				Unlocked   bool   `json:"unlocked"`
			} `json:"achievements"`
			AchievementConfig struct {
				Filters []struct {
					Key   string `json:"key"`
					Label string `json:"label"`
				} `json:"filters"`
				Locked []struct {
					Code     string `json:"code"`
					Unlocked bool   `json:"unlocked"`
				} `json:"locked"`
				Season struct {
					Title string `json:"title"`
				} `json:"season"`
			} `json:"achievementConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(growthRec.Body.Bytes(), &growthResp); err != nil {
		t.Fatal(err)
	}
	if growthResp.Data.Profile.TodayCreditScore != 100 {
		t.Fatalf("expected initial daily credit 100, got %d", growthResp.Data.Profile.TodayCreditScore)
	}
	if len(growthResp.Data.Profile.Achievements) == 0 || len(growthResp.Data.Footprints) == 0 {
		t.Fatalf("expected achievements and footprints: %s", growthRec.Body.String())
	}
	if len(growthResp.Data.Achievements) == 0 || growthResp.Data.Achievements[0].Title == "" || !growthResp.Data.Achievements[0].Unlocked {
		t.Fatalf("expected expanded achievement display data: %s", growthRec.Body.String())
	}
	if len(growthResp.Data.AchievementConfig.Filters) == 0 || growthResp.Data.AchievementConfig.Season.Title == "" {
		t.Fatalf("expected achievement display config: %s", growthRec.Body.String())
	}
	footprintsReq := httptest.NewRequest(http.MethodGet, "/api/app/footprints/my", nil)
	footprintsReq.Header.Set("Authorization", "Bearer "+creatorToken)
	footprintsRec := httptest.NewRecorder()
	mux.ServeHTTP(footprintsRec, footprintsReq)
	if footprintsRec.Code != http.StatusOK {
		t.Fatalf("expected footprints 200, got %d: %s", footprintsRec.Code, footprintsRec.Body.String())
	}
	var footprintsResp struct {
		Data struct {
			Items []struct {
				Action string `json:"action"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(footprintsRec.Body.Bytes(), &footprintsResp); err != nil {
		t.Fatal(err)
	}
	if len(footprintsResp.Data.Items) == 0 {
		t.Fatalf("expected footprints after review: %s", footprintsRec.Body.String())
	}
	summaryReq := httptest.NewRequest(http.MethodGet, "/api/app/users/me/summary", nil)
	summaryReq.Header.Set("Authorization", "Bearer "+creatorToken)
	summaryRec := httptest.NewRecorder()
	mux.ServeHTTP(summaryRec, summaryReq)
	if summaryRec.Code != http.StatusOK {
		t.Fatalf("expected user summary 200, got %d: %s", summaryRec.Code, summaryRec.Body.String())
	}
	var summaryResp struct {
		Data struct {
			User struct {
				ID int64 `json:"id"`
			} `json:"user"`
			ReviewTodoCount         int `json:"reviewTodoCount"`
			UnreadNotificationCount int `json:"unreadNotificationCount"`
			IncomeSummary           struct {
				TotalCent int64 `json:"totalCent"`
			} `json:"incomeSummary"`
			PointsSummary struct {
				AvailablePoints int `json:"availablePoints"`
			} `json:"pointsSummary"`
			RecentFootprints []struct {
				Action string `json:"action"`
			} `json:"recentFootprints"`
		} `json:"data"`
	}
	if err := json.Unmarshal(summaryRec.Body.Bytes(), &summaryResp); err != nil {
		t.Fatal(err)
	}
	if summaryResp.Data.User.ID != 1 || summaryResp.Data.ReviewTodoCount < 0 || summaryResp.Data.UnreadNotificationCount == 0 || summaryResp.Data.PointsSummary.AvailablePoints < 0 || len(summaryResp.Data.RecentFootprints) == 0 {
		t.Fatalf("expected populated user summary: %s", summaryRec.Body.String())
	}

	reviewRemindBody := postJSON(t, mux, "/api/internal/jobs/review-remind", "", `{}`, http.StatusOK)
	var reviewRemindResp struct {
		Data struct {
			Created int `json:"created"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reviewRemindBody, &reviewRemindResp); err != nil {
		t.Fatal(err)
	}
	if reviewRemindResp.Data.Created == 0 {
		t.Fatalf("expected review remind notifications: %s", string(reviewRemindBody))
	}

	exitBody := postJSON(t, mux, "/api/app/games/1/exit", creatorToken, `{}`, http.StatusOK)
	var exitResp struct {
		Data struct {
			Credit struct {
				BeforeScore int `json:"beforeScore"`
				AfterScore  int `json:"afterScore"`
			} `json:"credit"`
		} `json:"data"`
	}
	if err := json.Unmarshal(exitBody, &exitResp); err != nil {
		t.Fatal(err)
	}
	if exitResp.Data.Credit.BeforeScore != 100 || exitResp.Data.Credit.AfterScore != 90 {
		t.Fatalf("expected credit deducted from 100 to 90: %s", string(exitBody))
	}

	creditCenterBody := getJSON(t, mux, "/api/app/profile/credit-center", creatorToken, http.StatusOK)
	var creditCenterResp struct {
		Data struct {
			Score   int `json:"score"`
			Summary []struct {
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"summary"`
			Records []struct {
				GameID int64  `json:"gameId"`
				Score  string `json:"score"`
				Tone   string `json:"tone"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(creditCenterBody, &creditCenterResp); err != nil {
		t.Fatal(err)
	}
	if creditCenterResp.Data.Score != 90 || len(creditCenterResp.Data.Summary) != 3 || len(creditCenterResp.Data.Records) == 0 || creditCenterResp.Data.Records[0].Score != "-10" || creditCenterResp.Data.Records[0].Tone != "minus" {
		t.Fatalf("expected credit center with deduction record: %s", string(creditCenterBody))
	}
	noticesBody := getJSON(t, mux, "/api/app/notifications", creatorToken, http.StatusOK)
	var noticesResp struct {
		Data struct {
			Items []struct {
				NotifyType string `json:"notifyType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesBody, &noticesResp); err != nil {
		t.Fatal(err)
	}
	if len(noticesResp.Data.Items) == 0 {
		t.Fatalf("expected exit notifications: %s", string(noticesBody))
	}

	adminUserReq := httptest.NewRequest(http.MethodGet, "/api/admin/users/1/growth", nil)
	adminUserReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	adminUserRec := httptest.NewRecorder()
	mux.ServeHTTP(adminUserRec, adminUserReq)
	if adminUserRec.Code != http.StatusOK {
		t.Fatalf("expected admin user trace 200, got %d: %s", adminUserRec.Code, adminUserRec.Body.String())
	}
	adminGameReq := httptest.NewRequest(http.MethodGet, "/api/admin/games/1/review-trace", nil)
	adminGameReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	adminGameRec := httptest.NewRecorder()
	mux.ServeHTTP(adminGameRec, adminGameReq)
	if adminGameRec.Code != http.StatusOK {
		t.Fatalf("expected admin game trace 200, got %d: %s", adminGameRec.Code, adminGameRec.Body.String())
	}
}

func TestPlayerCancelRequestRequiresPlayerMemberAndNotifiesCreator(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "player-cancel-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "player-cancel-player")
	completeIdentityForTest(t, mux, playerToken)
	outsiderToken := loginForTestWithCode(t, mux, "player-cancel-outsider")
	completeIdentityForTest(t, mux, outsiderToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"player cancel game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	approveExtraMembersForHTTP(t, mux, creatorToken, 1, "player-cancel-extra", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	payload := `{"reasonKey":"need_changed","reasonText":"schedule changed","compensationRate":0,"payAmountText":"免费"}`
	postJSON(t, mux, "/api/app/games/1/player-cancel", outsiderToken, payload, http.StatusForbidden)
	body := postJSON(t, mux, "/api/app/games/1/player-cancel", playerToken, payload, http.StatusOK)
	var resp struct {
		Data struct {
			Status string `json:"status"`
			Game   struct {
				Status string `json:"status"`
			} `json:"game"`
			Credit struct {
				Reason      string `json:"reason"`
				ChangeValue int    `json:"changeValue"`
			} `json:"credit"`
			CancelRequest struct {
				GameID           int64   `json:"gameId"`
				ReasonKey        string  `json:"reasonKey"`
				CompensationRate float64 `json:"compensationRate"`
			} `json:"cancelRequest"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Status != "exited" || resp.Data.Game.Status == "canceled" || resp.Data.Credit.Reason != "player_cancel_service" || resp.Data.Credit.ChangeValue != -3 || resp.Data.CancelRequest.GameID != 1 || resp.Data.CancelRequest.ReasonKey != "need_changed" || resp.Data.CancelRequest.CompensationRate != 0 {
		t.Fatalf("expected player to exit without canceling whole game: %s", string(body))
	}

	noticesBody := getJSON(t, mux, "/api/app/notifications", creatorToken, http.StatusOK)
	var noticesResp struct {
		Data struct {
			Items []struct {
				ID         int64  `json:"id"`
				NotifyType string `json:"notifyType"`
				Title      string `json:"title"`
				Content    string `json:"content"`
				BizID      int64  `json:"bizId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesBody, &noticesResp); err != nil {
		t.Fatal(err)
	}
	foundID := int64(0)
	for _, item := range noticesResp.Data.Items {
		if item.NotifyType == "player_cancel_request" && item.BizID == 1 {
			if item.Title != "玩家取消组局" || !strings.Contains(item.Content, "原因：schedule changed") {
				t.Fatalf("expected localized player cancel notification: %s", string(noticesBody))
			}
			foundID = item.ID
			break
		}
	}
	if foundID <= 0 {
		t.Fatalf("expected player cancel notification: %s", string(noticesBody))
	}

	detailBody := getJSON(t, mux, "/api/app/game-invites/guide-cancel-detail?notificationId="+strconv.FormatInt(foundID, 10), creatorToken, http.StatusOK)
	if !strings.Contains(string(detailBody), `"roleLabel":"玩家"`) || !strings.Contains(string(detailBody), `"desc":"schedule changed"`) || !strings.Contains(string(detailBody), `"gameId":1`) {
		t.Fatalf("expected player cancel notification detail: %s", string(detailBody))
	}
	actionBody := postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(foundID, 10)+"/actions", creatorToken, `{"action":"detail"}`, http.StatusOK)
	if !strings.Contains(string(actionBody), `pages/game/guide-cancel/index?notificationId=`) {
		t.Fatalf("expected cancel notification detail route: %s", string(actionBody))
	}

	creatorBody := postJSON(t, mux, "/api/app/games/1/player-cancel", creatorToken, payload, http.StatusOK)
	var creatorResp struct {
		Data struct {
			Status string `json:"status"`
			Game   struct {
				Status string `json:"status"`
			} `json:"game"`
		} `json:"data"`
	}
	if err := json.Unmarshal(creatorBody, &creatorResp); err != nil {
		t.Fatal(err)
	}
	if creatorResp.Data.Status != "canceled" || creatorResp.Data.Game.Status != "canceled" {
		t.Fatalf("expected creator cancel to close whole game: %s", string(creatorBody))
	}
}

func TestExpertCancelRequestRequiresManagerAndNotifiesMembers(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "expert-cancel-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "expert-cancel-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"expert cancel game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	approveExtraMembersForHTTP(t, mux, creatorToken, 1, "expert-cancel-extra", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	payload := `{"reasonKey":"schedule_conflict","reasonText":"unexpected schedule","compensationRate":0,"payAmountText":"免费"}`
	postJSON(t, mux, "/api/app/games/1/expert-cancel", playerToken, payload, http.StatusForbidden)
	body := postJSON(t, mux, "/api/app/games/1/expert-cancel", creatorToken, payload, http.StatusOK)
	var resp struct {
		Data struct {
			Status string `json:"status"`
			Game   struct {
				Status string `json:"status"`
			} `json:"game"`
			Credit struct {
				Reason      string `json:"reason"`
				ChangeValue int    `json:"changeValue"`
			} `json:"credit"`
			CancelRequest struct {
				GameID           int64   `json:"gameId"`
				ReasonKey        string  `json:"reasonKey"`
				CompensationRate float64 `json:"compensationRate"`
			} `json:"cancelRequest"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Status != "canceled" || resp.Data.Game.Status != "canceled" || resp.Data.Credit.Reason != "expert_cancel_service" || resp.Data.Credit.ChangeValue != -5 || resp.Data.CancelRequest.GameID != 1 || resp.Data.CancelRequest.ReasonKey != "schedule_conflict" || resp.Data.CancelRequest.CompensationRate != 0 {
		t.Fatalf("expected canceled expert service: %s", string(body))
	}

	noticesBody := getJSON(t, mux, "/api/app/notifications", playerToken, http.StatusOK)
	var noticesResp struct {
		Data struct {
			Items []struct {
				NotifyType string `json:"notifyType"`
				BizID      int64  `json:"bizId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesBody, &noticesResp); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range noticesResp.Data.Items {
		if item.NotifyType == "expert_cancel_request" && item.BizID == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected expert cancel notification: %s", string(noticesBody))
	}
}

func TestProfitTemplateConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	adminToken := adminLoginForTest(t, mux)
	appToken := loginForTest(t, mux)

	putJSON(t, mux, "/api/admin/revenue/profit-template-config", adminToken, `{"depositRuleText":"后台规则","depositNoticeText":"后台提示","version":"profit-test"}`, http.StatusOK)
	body := getJSON(t, mux, "/api/app/games/profit-templates", appToken, http.StatusOK)
	var got struct {
		Data struct {
			DepositRuleText   string `json:"depositRuleText"`
			DepositNoticeText string `json:"depositNoticeText"`
			Deposit           struct {
				RuleText   string `json:"ruleText"`
				NoticeText string `json:"noticeText"`
			} `json:"deposit"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Data.DepositRuleText != "后台规则" || got.Data.DepositNoticeText != "后台提示" {
		t.Fatalf("expected app profit template config from admin update: %s", string(body))
	}
	if got.Data.Deposit.RuleText != "后台规则" || got.Data.Deposit.NoticeText != "后台提示" {
		t.Fatalf("expected nested deposit config: %s", string(body))
	}
}

func TestRevenuePreviewGenerateFreezeAndSettlementFlow(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "revenue-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "revenue-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"revenue game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "revenue-flow", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}

	adminToken := adminLoginForTest(t, mux)

	templateBody := postAdminJSON(t, mux, "/api/admin/revenue/templates", adminToken, `{"name":"default","gameType":"free","platformBps":1000,"creatorBps":3000,"memberBps":6000}`, http.StatusOK)
	var templateResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(templateBody, &templateResp); err != nil {
		t.Fatal(err)
	}
	if templateResp.Data.ID == 0 {
		t.Fatal("expected template id")
	}

	previewBody := postAdminJSON(t, mux, "/api/admin/revenue/preview", adminToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusOK)
	var previewResp struct {
		Data struct {
			CanGenerateRecord bool     `json:"canGenerateRecord"`
			BlockReasons      []string `json:"blockReasons"`
			Items             []struct {
				Role       string `json:"role"`
				AmountCent int64  `json:"amountCent"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(previewBody, &previewResp); err != nil {
		t.Fatal(err)
	}
	if len(previewResp.Data.Items) < 3 {
		t.Fatalf("expected platform creator member items: %s", string(previewBody))
	}
	if previewResp.Data.CanGenerateRecord || !hasString(previewResp.Data.BlockReasons, "review_incomplete") {
		t.Fatalf("expected review_incomplete preview block: %s", string(previewBody))
	}

	ruleBody := postAdminJSON(t, mux, "/api/admin/revenue/rules", adminToken, `{"templateId":1,"ruleCode":"guide_bps","ruleValue":"500"}`, http.StatusOK)
	var ruleResp struct {
		Data struct {
			ID        int64  `json:"id"`
			RuleCode  string `json:"ruleCode"`
			RuleValue string `json:"ruleValue"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ruleBody, &ruleResp); err != nil {
		t.Fatal(err)
	}
	if ruleResp.Data.ID == 0 || ruleResp.Data.RuleCode != "guide_bps" || ruleResp.Data.RuleValue != "500" {
		t.Fatalf("unexpected revenue rule create: %s", string(ruleBody))
	}
	rulesBody := getAdminJSON(t, mux, "/api/admin/revenue/rules?templateId=1", adminToken, http.StatusOK)
	var rulesResp struct {
		Data struct {
			Items []struct {
				RuleCode  string `json:"ruleCode"`
				RuleValue string `json:"ruleValue"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rulesBody, &rulesResp); err != nil {
		t.Fatal(err)
	}
	if len(rulesResp.Data.Items) != 1 || rulesResp.Data.Items[0].RuleCode != "guide_bps" {
		t.Fatalf("unexpected revenue rules list: %s", string(rulesBody))
	}
	ruleUpdateBody := postAdminJSON(t, mux, "/api/admin/revenue/rules", adminToken, `{"templateId":1,"ruleCode":"guide_bps","ruleValue":"700"}`, http.StatusOK)
	var ruleUpdateResp struct {
		Data struct {
			RuleValue string `json:"ruleValue"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ruleUpdateBody, &ruleUpdateResp); err != nil {
		t.Fatal(err)
	}
	if ruleUpdateResp.Data.RuleValue != "700" {
		t.Fatalf("expected revenue rule upsert, got %s", string(ruleUpdateBody))
	}
	rulePreviewBody := postAdminJSON(t, mux, "/api/admin/revenue/preview", adminToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusOK)
	var rulePreviewResp struct {
		Data struct {
			Items []struct {
				Role       string `json:"role"`
				AmountCent int64  `json:"amountCent"`
			} `json:"items"`
			Rules []struct {
				RuleCode  string `json:"ruleCode"`
				RuleValue string `json:"ruleValue"`
			} `json:"rules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rulePreviewBody, &rulePreviewResp); err != nil {
		t.Fatal(err)
	}
	if len(rulePreviewResp.Data.Rules) != 1 || rulePreviewResp.Data.Rules[0].RuleCode != "guide_bps" || rulePreviewResp.Data.Rules[0].RuleValue != "700" {
		t.Fatalf("expected revenue preview to include active template rules: %s", string(rulePreviewBody))
	}
	if !hasRevenueItemRole(rulePreviewResp.Data.Items, "guide", 700) {
		t.Fatalf("expected revenue preview to apply guide_bps rule: %s", string(rulePreviewBody))
	}
	simulateOutsiderToken := loginForTestWithCode(t, mux, "revenue-simulate-outsider")
	completeIdentityForTest(t, mux, simulateOutsiderToken)
	postJSON(t, mux, "/api/app/revenues/simulate", simulateOutsiderToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusForbidden)
	appPreviewBody := postJSON(t, mux, "/api/app/revenues/simulate", playerToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusOK)
	var appPreviewResp struct {
		Data struct {
			GameID            int64    `json:"gameId"`
			CanGenerateRecord bool     `json:"canGenerateRecord"`
			BlockReasons      []string `json:"blockReasons"`
			Items             []struct {
				Role       string `json:"role"`
				AmountCent int64  `json:"amountCent"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(appPreviewBody, &appPreviewResp); err != nil {
		t.Fatal(err)
	}
	if appPreviewResp.Data.GameID != 1 || appPreviewResp.Data.CanGenerateRecord || !hasString(appPreviewResp.Data.BlockReasons, "review_incomplete") || !hasRevenueItemRole(appPreviewResp.Data.Items, "guide", 700) {
		t.Fatalf("expected app revenue simulate preview only: %s", string(appPreviewBody))
	}
	recordsBeforeBody := getAdminJSON(t, mux, "/api/admin/revenue/records", adminToken, http.StatusOK)
	var recordsBeforeResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recordsBeforeBody, &recordsBeforeResp); err != nil {
		t.Fatal(err)
	}
	if len(recordsBeforeResp.Data.Items) != 0 {
		t.Fatalf("expected app revenue simulate not to generate records: %s", string(recordsBeforeBody))
	}
	revenueOperationLogsBody := getAdminJSON(t, mux, "/api/admin/operation-logs", adminToken, http.StatusOK)
	var revenueOperationLogsResp struct {
		Data struct {
			Items []struct {
				Action      string `json:"action"`
				TargetType  string `json:"targetType"`
				TargetID    string `json:"targetId"`
				AdminUserID int64  `json:"adminUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(revenueOperationLogsBody, &revenueOperationLogsResp); err != nil {
		t.Fatal(err)
	}
	if !hasOperationLog(revenueOperationLogsResp.Data.Items, "revenue:template:create", "revenue_template", "1", 1) {
		t.Fatalf("expected revenue template operation log: %s", string(revenueOperationLogsBody))
	}
	if !hasOperationLog(revenueOperationLogsResp.Data.Items, "revenue:rule:upsert", "revenue_rule", "1", 1) {
		t.Fatalf("expected revenue rule operation log: %s", string(revenueOperationLogsBody))
	}

	postAdminJSON(t, mux, "/api/admin/revenue/records/generate", adminToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusConflict)
	completeGameReviewsForHTTP(t, mux, 1, append([]string{creatorToken, playerToken}, extraTokens...))
	recordBody := postAdminJSON(t, mux, "/api/admin/revenue/records/generate", adminToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusOK)
	var recordResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recordBody, &recordResp); err != nil {
		t.Fatal(err)
	}
	if recordResp.Data.ID == 0 || recordResp.Data.Status != "pending_settlement" {
		t.Fatalf("unexpected revenue record: %s", string(recordBody))
	}

	postAdminJSON(t, mux, "/api/admin/revenue/records/1/freeze", adminToken, `{"reason":"report"}`, http.StatusOK)
	postAdminJSON(t, mux, "/api/admin/revenue/records/1/settle", adminToken, `{"method":"offline","proofNo":"P001"}`, http.StatusConflict)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"settle game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/2/approve-local", creatorToken, `{}`, http.StatusOK)
	settleAppBody := postJSON(t, mux, "/api/app/games/2/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	var settleAppResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(settleAppBody, &settleAppResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(settleAppResp.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens = approveExtraMembersForHTTP(t, mux, creatorToken, 2, "connections-second", 3)
	postJSON(t, mux, "/api/app/games/2/manual-start", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/2/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/2/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/2/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	completeGameReviewsForHTTP(t, mux, 2, append([]string{creatorToken, playerToken}, extraTokens...))
	postAdminJSON(t, mux, "/api/admin/revenue/records/generate", adminToken, `{"gameId":2,"amountCent":10000,"templateId":1}`, http.StatusOK)
	settleBody := postAdminJSON(t, mux, "/api/admin/revenue/records/2/settle", adminToken, `{"method":"offline","proofNo":"P002"}`, http.StatusOK)
	var settleResp struct {
		Data struct {
			Record struct {
				Status string `json:"status"`
			} `json:"record"`
		} `json:"data"`
	}
	if err := json.Unmarshal(settleBody, &settleResp); err != nil {
		t.Fatal(err)
	}
	if settleResp.Data.Record.Status != "settled" {
		t.Fatalf("expected settled record: %s", string(settleBody))
	}
	settlementsBody := getAdminJSON(t, mux, "/api/admin/revenue/settlements", adminToken, http.StatusOK)
	var settlementsResp struct {
		Data struct {
			Items []struct {
				RecordID   int64  `json:"recordId"`
				Method     string `json:"method"`
				ProofNo    string `json:"proofNo"`
				AmountCent int64  `json:"amountCent"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(settlementsBody, &settlementsResp); err != nil {
		t.Fatal(err)
	}
	if len(settlementsResp.Data.Items) != 1 || settlementsResp.Data.Items[0].RecordID != 2 || settlementsResp.Data.Items[0].ProofNo != "P002" || settlementsResp.Data.Items[0].AmountCent != 10000 {
		t.Fatalf("expected settlement record list: %s", string(settlementsBody))
	}

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"reported before revenue","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/3/approve-local", creatorToken, `{}`, http.StatusOK)
	frozenAppBody := postJSON(t, mux, "/api/app/games/3/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	var frozenAppResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(frozenAppBody, &frozenAppResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(frozenAppResp.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens = approveExtraMembersForHTTP(t, mux, creatorToken, 3, "connections-third", 3)
	postJSON(t, mux, "/api/app/games/3/manual-start", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/3/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/3/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/3/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	completeGameReviewsForHTTP(t, mux, 3, append([]string{creatorToken, playerToken}, extraTokens...))
	postJSON(t, mux, "/api/app/reports", playerToken, `{"gameId":3,"targetUserId":1,"reportType":"service_dispute","content":"before revenue"}`, http.StatusOK)
	disputedPreviewBody := postAdminJSON(t, mux, "/api/admin/revenue/preview", adminToken, `{"gameId":3,"amountCent":10000,"templateId":1}`, http.StatusOK)
	var disputedPreviewResp struct {
		Data struct {
			CanGenerateRecord bool     `json:"canGenerateRecord"`
			BlockReasons      []string `json:"blockReasons"`
		} `json:"data"`
	}
	if err := json.Unmarshal(disputedPreviewBody, &disputedPreviewResp); err != nil {
		t.Fatal(err)
	}
	if disputedPreviewResp.Data.CanGenerateRecord || !hasString(disputedPreviewResp.Data.BlockReasons, "disputed") {
		t.Fatalf("expected disputed preview block: %s", string(disputedPreviewBody))
	}
	frozenBody := postAdminJSON(t, mux, "/api/admin/revenue/records/generate", adminToken, `{"gameId":3,"amountCent":10000,"templateId":1}`, http.StatusOK)
	var frozenResp struct {
		Data struct {
			ID           int64  `json:"id"`
			Status       string `json:"status"`
			FrozenReason string `json:"frozenReason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(frozenBody, &frozenResp); err != nil {
		t.Fatal(err)
	}
	if frozenResp.Data.ID == 0 || frozenResp.Data.Status != "frozen" || frozenResp.Data.FrozenReason != "report_open" {
		t.Fatalf("expected open report to freeze generated record: %s", string(frozenBody))
	}
	postAdminJSON(t, mux, "/api/admin/revenue/records/3/settle", adminToken, `{"method":"offline","proofNo":"P003"}`, http.StatusConflict)

	incomeReq := httptest.NewRequest(http.MethodGet, "/api/app/incomes/summary", nil)
	incomeReq.Header.Set("Authorization", "Bearer "+playerToken)
	incomeRec := httptest.NewRecorder()
	mux.ServeHTTP(incomeRec, incomeReq)
	if incomeRec.Code != http.StatusOK {
		t.Fatalf("expected income summary 200, got %d: %s", incomeRec.Code, incomeRec.Body.String())
	}
	accountReq := httptest.NewRequest(http.MethodGet, "/api/app/incomes/account", nil)
	accountReq.Header.Set("Authorization", "Bearer "+playerToken)
	accountRec := httptest.NewRecorder()
	mux.ServeHTTP(accountRec, accountReq)
	if accountRec.Code != http.StatusOK {
		t.Fatalf("expected income account 200, got %d: %s", accountRec.Code, accountRec.Body.String())
	}
	var accountResp struct {
		Data struct {
			UserID      int64 `json:"userId"`
			TotalCent   int64 `json:"totalCent"`
			PendingCent int64 `json:"pendingCent"`
			SettledCent int64 `json:"settledCent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(accountRec.Body.Bytes(), &accountResp); err != nil {
		t.Fatal(err)
	}
	if accountResp.Data.UserID == 0 || accountResp.Data.TotalCent == 0 || accountResp.Data.PendingCent == 0 || accountResp.Data.SettledCent == 0 {
		t.Fatalf("expected income account totals with pending and settled amounts: %s", accountRec.Body.String())
	}
	singularIncomeReq := httptest.NewRequest(http.MethodGet, "/api/app/income/summary", nil)
	singularIncomeReq.Header.Set("Authorization", "Bearer "+playerToken)
	singularIncomeRec := httptest.NewRecorder()
	mux.ServeHTTP(singularIncomeRec, singularIncomeReq)
	if singularIncomeRec.Code != http.StatusOK {
		t.Fatalf("expected singular income summary 200, got %d: %s", singularIncomeRec.Code, singularIncomeRec.Body.String())
	}
	incomeLogsReq := httptest.NewRequest(http.MethodGet, "/api/app/incomes/logs", nil)
	incomeLogsReq.Header.Set("Authorization", "Bearer "+playerToken)
	incomeLogsRec := httptest.NewRecorder()
	mux.ServeHTTP(incomeLogsRec, incomeLogsReq)
	if incomeLogsRec.Code != http.StatusOK {
		t.Fatalf("expected income logs 200, got %d: %s", incomeLogsRec.Code, incomeLogsRec.Body.String())
	}
	var incomeLogsResp struct {
		Data struct {
			Items []struct {
				GameID     int64  `json:"gameId"`
				Role       string `json:"role"`
				AmountCent int64  `json:"amountCent"`
				Status     string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(incomeLogsRec.Body.Bytes(), &incomeLogsResp); err != nil {
		t.Fatal(err)
	}
	if len(incomeLogsResp.Data.Items) < 3 {
		t.Fatalf("expected player income logs including frozen dispute records, got %s", incomeLogsRec.Body.String())
	}
	seenSettled := false
	seenFrozen := false
	for _, item := range incomeLogsResp.Data.Items {
		if item.AmountCent == 0 {
			t.Fatalf("expected non-zero income log amounts: %s", incomeLogsRec.Body.String())
		}
		if item.Status == "settled" {
			seenSettled = true
		}
		if item.Status == "frozen" {
			seenFrozen = true
		}
	}
	if !seenSettled || !seenFrozen {
		t.Fatalf("expected settled and frozen income logs: %s", incomeLogsRec.Body.String())
	}

	singularIncomeLogsReq := httptest.NewRequest(http.MethodGet, "/api/app/income/logs", nil)
	singularIncomeLogsReq.Header.Set("Authorization", "Bearer "+playerToken)
	singularIncomeLogsRec := httptest.NewRecorder()
	mux.ServeHTTP(singularIncomeLogsRec, singularIncomeLogsReq)
	if singularIncomeLogsRec.Code != http.StatusOK {
		t.Fatalf("expected singular income logs 200, got %d: %s", singularIncomeLogsRec.Code, singularIncomeLogsRec.Body.String())
	}

	settledLogsReq := httptest.NewRequest(http.MethodGet, "/api/app/incomes/logs?status=settled", nil)
	settledLogsReq.Header.Set("Authorization", "Bearer "+playerToken)
	settledLogsRec := httptest.NewRecorder()
	mux.ServeHTTP(settledLogsRec, settledLogsReq)
	if settledLogsRec.Code != http.StatusOK {
		t.Fatalf("expected settled income logs 200, got %d: %s", settledLogsRec.Code, settledLogsRec.Body.String())
	}
	var settledLogsResp struct {
		Data struct {
			Items []struct {
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(settledLogsRec.Body.Bytes(), &settledLogsResp); err != nil {
		t.Fatal(err)
	}
	if len(settledLogsResp.Data.Items) != 1 || settledLogsResp.Data.Items[0].Status != "settled" {
		t.Fatalf("expected one settled income log: %s", settledLogsRec.Body.String())
	}

	outsiderToken := loginForTestWithCode(t, mux, "revenue-outsider")
	completeIdentityForTest(t, mux, outsiderToken)
	outsiderAccountBody := getJSON(t, mux, "/api/app/incomes/account", outsiderToken, http.StatusOK)
	var outsiderAccountResp struct {
		Data struct {
			UserID      int64 `json:"userId"`
			TotalCent   int64 `json:"totalCent"`
			PendingCent int64 `json:"pendingCent"`
			SettledCent int64 `json:"settledCent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(outsiderAccountBody, &outsiderAccountResp); err != nil {
		t.Fatal(err)
	}
	if outsiderAccountResp.Data.UserID == accountResp.Data.UserID || outsiderAccountResp.Data.TotalCent != 0 || outsiderAccountResp.Data.PendingCent != 0 || outsiderAccountResp.Data.SettledCent != 0 {
		t.Fatalf("expected outsider income account isolated: %s", string(outsiderAccountBody))
	}
	outsiderLogsBody := getJSON(t, mux, "/api/app/incomes/logs", outsiderToken, http.StatusOK)
	var outsiderLogsResp struct {
		Data struct {
			Items []struct {
				RecordID int64 `json:"recordId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(outsiderLogsBody, &outsiderLogsResp); err != nil {
		t.Fatal(err)
	}
	if len(outsiderLogsResp.Data.Items) != 0 {
		t.Fatalf("expected outsider income logs isolated: %s", string(outsiderLogsBody))
	}
}

func TestMemberReportAndTeamHTTP(t *testing.T) {
	const leaderUserIDForTest int64 = 1
	const memberUserIDForTest int64 = 2

	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	leaderToken := loginForTestWithCode(t, mux, "team-leader")
	completeIdentityForTest(t, mux, leaderToken)
	memberToken := loginForTestWithCode(t, mux, "team-member")
	completeIdentityForTest(t, mux, memberToken)

	memberReportReq := httptest.NewRequest(http.MethodGet, "/api/app/member-reports/me", nil)
	memberReportReq.Header.Set("Authorization", "Bearer "+leaderToken)
	memberReportRec := httptest.NewRecorder()
	mux.ServeHTTP(memberReportRec, memberReportReq)
	if memberReportRec.Code != http.StatusForbidden {
		t.Fatalf("expected member report forbidden, got %d: %s", memberReportRec.Code, memberReportRec.Body.String())
	}
	var memberReportForbidden struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(memberReportRec.Body.Bytes(), &memberReportForbidden); err != nil {
		t.Fatal(err)
	}
	if memberReportForbidden.Code != 40351 {
		t.Fatalf("expected code 40351, got %s", memberReportRec.Body.String())
	}

	teamReq := httptest.NewRequest(http.MethodGet, "/api/app/teams/my/members", nil)
	teamReq.Header.Set("Authorization", "Bearer "+leaderToken)
	teamRec := httptest.NewRecorder()
	mux.ServeHTTP(teamRec, teamReq)
	if teamRec.Code != http.StatusForbidden {
		t.Fatalf("expected team forbidden, got %d: %s", teamRec.Code, teamRec.Body.String())
	}
	var teamForbidden struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(teamRec.Body.Bytes(), &teamForbidden); err != nil {
		t.Fatal(err)
	}
	if teamForbidden.Code != 40352 {
		t.Fatalf("expected code 40352, got %s", teamRec.Body.String())
	}

	postJSON(t, mux, "/api/app/games", leaderToken, `{"title":"team revenue game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", leaderToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", memberToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", leaderToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, leaderToken, 1, "member-report", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", leaderToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm", leaderToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", memberToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}

	adminToken := adminLoginForTest(t, mux)
	postAdminJSON(t, mux, "/api/admin/revenue/templates", adminToken, `{"name":"default","gameType":"free","platformBps":1000,"creatorBps":3000,"memberBps":6000}`, http.StatusOK)
	completeGameReviewsForHTTP(t, mux, 1, append([]string{leaderToken, memberToken}, extraTokens...))
	postAdminJSON(t, mux, "/api/admin/revenue/records/generate", adminToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusOK)

	server.memberReports.GrantMembership(leaderUserIDForTest, "basic member", 1)
	team := server.teams.GrantLeader(leaderUserIDForTest, "partner team")
	if team.ID == 0 {
		t.Fatal("expected team id")
	}
	if _, err := server.teams.AddMember(leaderUserIDForTest, memberUserIDForTest, "invite"); err != nil {
		t.Fatal(err)
	}

	memberReportBody := getJSON(t, mux, "/api/app/member-reports/me", leaderToken, http.StatusOK)
	var memberReportResp struct {
		Data struct {
			UserID        int64 `json:"userId"`
			Participated  int   `json:"participatedGames"`
			Completed     int   `json:"completedGames"`
			IncomeSummary struct {
				TotalCent int64 `json:"totalCent"`
			} `json:"incomeSummary"`
			InvitedCount int `json:"invitedCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(memberReportBody, &memberReportResp); err != nil {
		t.Fatal(err)
	}
	if memberReportResp.Data.UserID != leaderUserIDForTest || memberReportResp.Data.Participated != 1 || memberReportResp.Data.Completed != 1 || memberReportResp.Data.IncomeSummary.TotalCent == 0 || memberReportResp.Data.InvitedCount != 1 {
		t.Fatalf("unexpected member report: %s", string(memberReportBody))
	}

	teamBody := getJSON(t, mux, "/api/app/teams/my", leaderToken, http.StatusOK)
	var teamResp struct {
		Data struct {
			ID           int64  `json:"id"`
			LeaderUserID int64  `json:"leaderUserId"`
			Name         string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(teamBody, &teamResp); err != nil {
		t.Fatal(err)
	}
	if teamResp.Data.ID != team.ID || teamResp.Data.LeaderUserID != leaderUserIDForTest || teamResp.Data.Name != "partner team" {
		t.Fatalf("unexpected team: %s", string(teamBody))
	}

	membersBody := getJSON(t, mux, "/api/app/teams/my/members", leaderToken, http.StatusOK)
	var membersResp struct {
		Data struct {
			Items []struct {
				UserID        int64  `json:"userId"`
				RelationLevel int    `json:"relationLevel"`
				Source        string `json:"source"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(membersBody, &membersResp); err != nil {
		t.Fatal(err)
	}
	if len(membersResp.Data.Items) != 1 || membersResp.Data.Items[0].UserID != memberUserIDForTest || membersResp.Data.Items[0].RelationLevel != 1 {
		t.Fatalf("unexpected team members: %s", string(membersBody))
	}

	summaryBody := getJSON(t, mux, "/api/app/teams/my/revenue-summary", leaderToken, http.StatusOK)
	var summaryResp struct {
		Data struct {
			TeamID      int64 `json:"teamId"`
			MemberCount int   `json:"memberCount"`
			TotalCent   int64 `json:"totalCent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(summaryBody, &summaryResp); err != nil {
		t.Fatal(err)
	}
	if summaryResp.Data.TeamID != team.ID || summaryResp.Data.MemberCount != 1 || summaryResp.Data.TotalCent == 0 {
		t.Fatalf("unexpected team revenue summary: %s", string(summaryBody))
	}

	noTeamPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/teams", nil)
	noTeamPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noTeamPermRec, noTeamPermReq)
	if noTeamPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin teams 403 without permission, got %d: %s", noTeamPermRec.Code, noTeamPermRec.Body.String())
	}

	adminTeamsBody := getAdminJSONWithPermission(t, mux, "/api/admin/teams", "team:read", http.StatusOK)
	var adminTeamsResp struct {
		Data struct {
			Items []struct {
				ID           int64  `json:"id"`
				LeaderUserID int64  `json:"leaderUserId"`
				Name         string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminTeamsBody, &adminTeamsResp); err != nil {
		t.Fatal(err)
	}
	if len(adminTeamsResp.Data.Items) != 1 || adminTeamsResp.Data.Items[0].ID != team.ID || adminTeamsResp.Data.Items[0].LeaderUserID != leaderUserIDForTest {
		t.Fatalf("unexpected admin teams: %s", string(adminTeamsBody))
	}

	adminTeamBody := getAdminJSONWithPermission(t, mux, "/api/admin/teams/1", "team:read", http.StatusOK)
	var adminTeamResp struct {
		Data struct {
			Team struct {
				ID int64 `json:"id"`
			} `json:"team"`
			Members []struct {
				UserID int64 `json:"userId"`
			} `json:"members"`
			RevenueSummary struct {
				MemberCount int   `json:"memberCount"`
				TotalCent   int64 `json:"totalCent"`
			} `json:"revenueSummary"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminTeamBody, &adminTeamResp); err != nil {
		t.Fatal(err)
	}
	if adminTeamResp.Data.Team.ID != team.ID || len(adminTeamResp.Data.Members) != 1 || adminTeamResp.Data.RevenueSummary.MemberCount != 1 || adminTeamResp.Data.RevenueSummary.TotalCent == 0 {
		t.Fatalf("unexpected admin team detail: %s", string(adminTeamBody))
	}

	noReportPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/member-reports", nil)
	noReportPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noReportPermRec, noReportPermReq)
	if noReportPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin member reports 403 without permission, got %d: %s", noReportPermRec.Code, noReportPermRec.Body.String())
	}

	adminReportsBody := getAdminJSONWithPermission(t, mux, "/api/admin/member-reports", "member_report:read", http.StatusOK)
	var adminReportsResp struct {
		Data struct {
			Items []struct {
				UserID         int64  `json:"userId"`
				MembershipPlan string `json:"membershipPlan"`
				InvitedCount   int    `json:"invitedCount"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminReportsBody, &adminReportsResp); err != nil {
		t.Fatal(err)
	}
	if len(adminReportsResp.Data.Items) != 1 || adminReportsResp.Data.Items[0].UserID != leaderUserIDForTest || adminReportsResp.Data.Items[0].InvitedCount != 1 {
		t.Fatalf("unexpected admin member reports: %s", string(adminReportsBody))
	}
}

func TestPointsAndRedemptionHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "points-user")
	completeIdentityForTest(t, mux, token)
	otherToken := loginForTestWithCode(t, mux, "points-other")
	completeIdentityForTest(t, mux, otherToken)

	summaryBody := getJSON(t, mux, "/api/app/points/summary", token, http.StatusOK)
	var emptySummary struct {
		Data struct {
			UserID          int64 `json:"userId"`
			AvailablePoints int   `json:"availablePoints"`
			Stats           []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"stats"`
			Rules []struct {
				Text string `json:"text"`
			} `json:"rules"`
			Filters []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"filters"`
		} `json:"data"`
	}
	if err := json.Unmarshal(summaryBody, &emptySummary); err != nil {
		t.Fatal(err)
	}
	if emptySummary.Data.UserID != 1 || emptySummary.Data.AvailablePoints != 0 {
		t.Fatalf("unexpected empty points summary: %s", string(summaryBody))
	}
	if len(emptySummary.Data.Stats) == 0 || len(emptySummary.Data.Rules) == 0 || len(emptySummary.Data.Filters) != 3 {
		t.Fatalf("expected points page config in summary: %s", string(summaryBody))
	}

	item, err := server.redemption.CreateItem(redemption.CreateItemRequest{Name: "coffee", PointsCost: 10, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/redemption/orders", token, `{"itemId":1}`, http.StatusConflict)

	if _, _, err := server.points.Grant(1, 20, "test_seed", 0, "seed points"); err != nil {
		t.Fatal(err)
	}

	itemsBody := getJSON(t, mux, "/api/app/redemption/items", token, http.StatusOK)
	var itemsResp struct {
		Data struct {
			Items []struct {
				ID         int64 `json:"id"`
				PointsCost int   `json:"pointsCost"`
				Stock      int   `json:"stock"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(itemsBody, &itemsResp); err != nil {
		t.Fatal(err)
	}
	if len(itemsResp.Data.Items) != 1 || itemsResp.Data.Items[0].ID != item.ID || itemsResp.Data.Items[0].Stock != 1 {
		t.Fatalf("unexpected redemption items: %s", string(itemsBody))
	}

	orderBody := postJSON(t, mux, "/api/app/redemption/orders", token, `{"itemId":1}`, http.StatusOK)
	var orderResp struct {
		Data struct {
			ID         int64  `json:"id"`
			OrderNo    string `json:"orderNo"`
			PointsCost int    `json:"pointsCost"`
		} `json:"data"`
	}
	if err := json.Unmarshal(orderBody, &orderResp); err != nil {
		t.Fatal(err)
	}
	if orderResp.Data.ID == 0 || orderResp.Data.OrderNo == "" || orderResp.Data.PointsCost != 10 {
		t.Fatalf("unexpected redemption order: %s", string(orderBody))
	}
	postJSON(t, mux, "/api/app/redemption/orders", token, `{"itemId":1}`, http.StatusConflict)

	logsBody := getJSON(t, mux, "/api/app/points/logs", token, http.StatusOK)
	var logsResp struct {
		Data struct {
			Items []struct {
				ChangeValue int    `json:"changeValue"`
				BizType     string `json:"bizType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(logsBody, &logsResp); err != nil {
		t.Fatal(err)
	}
	if len(logsResp.Data.Items) != 2 || logsResp.Data.Items[1].ChangeValue != -10 || logsResp.Data.Items[1].BizType != "redemption_order" {
		t.Fatalf("expected grant and deduct logs: %s", string(logsBody))
	}

	myOrdersBody := getJSON(t, mux, "/api/app/redemption/orders/my", token, http.StatusOK)
	var myOrdersResp struct {
		Data struct {
			Items []struct {
				ID        int64  `json:"id"`
				StatusKey string `json:"statusKey"`
			} `json:"items"`
			Orders []struct {
				ID        int64  `json:"id"`
				StatusKey string `json:"statusKey"`
			} `json:"orders"`
			Tabs []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
			} `json:"tabs"`
			EmptyText  string `json:"emptyText"`
			PageConfig struct {
				DetailEmptyText    string `json:"detailEmptyText"`
				LogisticsEmptyText string `json:"logisticsEmptyText"`
				CancelConfirm      struct {
					Title  string `json:"title"`
					Reason string `json:"reason"`
				} `json:"cancelConfirm"`
			} `json:"pageConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(myOrdersBody, &myOrdersResp); err != nil {
		t.Fatal(err)
	}
	if len(myOrdersResp.Data.Items) != 1 || len(myOrdersResp.Data.Orders) != 1 || myOrdersResp.Data.Items[0].ID != orderResp.Data.ID || myOrdersResp.Data.Items[0].StatusKey != "pending" {
		t.Fatalf("unexpected redemption orders: %s", string(myOrdersBody))
	}
	if len(myOrdersResp.Data.Tabs) == 0 || myOrdersResp.Data.EmptyText == "" || myOrdersResp.Data.PageConfig.CancelConfirm.Title == "" || myOrdersResp.Data.PageConfig.LogisticsEmptyText == "" || myOrdersResp.Data.PageConfig.DetailEmptyText == "" || myOrdersResp.Data.PageConfig.CancelConfirm.Reason == "" {
		t.Fatalf("expected redemption order page config: %s", string(myOrdersBody))
	}
	approvedFilterBody := getJSON(t, mux, "/api/app/redemption/orders/my?status=approved", token, http.StatusOK)
	var approvedFilterResp struct {
		Data struct {
			Orders []struct {
				ID int64 `json:"id"`
			} `json:"orders"`
		} `json:"data"`
	}
	if err := json.Unmarshal(approvedFilterBody, &approvedFilterResp); err != nil {
		t.Fatal(err)
	}
	if len(approvedFilterResp.Data.Orders) != 0 {
		t.Fatalf("expected no approved orders before admin review: %s", string(approvedFilterBody))
	}

	orderPath := fmt.Sprintf("/api/app/redemption/orders/%d", orderResp.Data.ID)
	cancelPath := fmt.Sprintf("%s/cancel", orderPath)
	orderDetailBody := getJSON(t, mux, orderPath, token, http.StatusOK)
	var orderDetailResp struct {
		Data struct {
			Order struct {
				ID        int64  `json:"id"`
				StatusKey string `json:"statusKey"`
			} `json:"order"`
			DetailRows []struct {
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"detailRows"`
			EmptyText string `json:"emptyText"`
		} `json:"data"`
	}
	if err := json.Unmarshal(orderDetailBody, &orderDetailResp); err != nil {
		t.Fatal(err)
	}
	if orderDetailResp.Data.Order.ID != orderResp.Data.ID || orderDetailResp.Data.Order.StatusKey != "pending" || len(orderDetailResp.Data.DetailRows) == 0 {
		t.Fatalf("unexpected redemption order detail: %s", string(orderDetailBody))
	}
	if orderDetailResp.Data.EmptyText == "" {
		t.Fatalf("expected redemption order detail empty text from page config: %s", string(orderDetailBody))
	}
	getJSON(t, mux, orderPath, otherToken, http.StatusNotFound)

	cancelBody := postJSON(t, mux, cancelPath, token, `{"reason":"change mind"}`, http.StatusOK)
	var cancelResp struct {
		Data struct {
			Order struct {
				ID        int64  `json:"id"`
				StatusKey string `json:"statusKey"`
			} `json:"order"`
			PointsSummary struct {
				AvailablePoints int `json:"availablePoints"`
			} `json:"pointsSummary"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cancelBody, &cancelResp); err != nil {
		t.Fatal(err)
	}
	if cancelResp.Data.Order.ID != orderResp.Data.ID || cancelResp.Data.Order.StatusKey != "canceled" || cancelResp.Data.PointsSummary.AvailablePoints != 20 {
		t.Fatalf("unexpected redemption cancel response: %s", string(cancelBody))
	}
	postJSON(t, mux, cancelPath, token, `{}`, http.StatusConflict)
	if items := server.redemption.AdminItems(); len(items) == 0 || items[0].Stock != 1 {
		t.Fatalf("expected canceled order to restore stock, got %+v", items)
	}

	concurrentItem, err := server.redemption.CreateItem(redemption.CreateItemRequest{Name: "ticket", PointsCost: 5, Stock: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.points.Grant(1, 10, "test_seed", 0, "seed points"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.points.Grant(2, 10, "test_seed", 0, "seed points"); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for _, userToken := range []string{token, otherToken} {
		wg.Add(1)
		go func(userToken string) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/app/redemption/orders", bytes.NewBufferString(`{"itemId":2}`))
			req.Header.Set("Authorization", "Bearer "+userToken)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			statuses <- rec.Code
		}(userToken)
	}
	wg.Wait()
	close(statuses)

	successCount := 0
	conflictCount := 0
	for status := range statuses {
		if status == http.StatusOK {
			successCount++
		}
		if status == http.StatusConflict {
			conflictCount++
		}
	}
	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one success and one conflict for item %d, got success=%d conflict=%d", concurrentItem.ID, successCount, conflictCount)
	}
}

func TestAdminPointsAndRedemptionHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	userToken := loginForTestWithCode(t, mux, "admin-redemption-user")
	completeIdentityForTest(t, mux, userToken)
	if _, _, err := server.points.Grant(1, 30, "test_seed", 0, "seed"); err != nil {
		t.Fatal(err)
	}

	noPointsPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/points/logs", nil)
	noPointsPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noPointsPermRec, noPointsPermReq)
	if noPointsPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin points logs 403 without permission, got %d: %s", noPointsPermRec.Code, noPointsPermRec.Body.String())
	}

	pointsBody := getAdminJSONWithPermission(t, mux, "/api/admin/points/logs", "points:read", http.StatusOK)
	var pointsResp struct {
		Data struct {
			Items []struct {
				UserID      int64  `json:"userId"`
				ChangeValue int    `json:"changeValue"`
				BizType     string `json:"bizType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pointsBody, &pointsResp); err != nil {
		t.Fatal(err)
	}
	if len(pointsResp.Data.Items) != 1 || pointsResp.Data.Items[0].UserID != 1 || pointsResp.Data.Items[0].ChangeValue != 30 {
		t.Fatalf("expected admin points logs: %s", string(pointsBody))
	}

	noRedemptionPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/redemption/items", nil)
	noRedemptionPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noRedemptionPermRec, noRedemptionPermReq)
	if noRedemptionPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin redemption items 403 without permission, got %d: %s", noRedemptionPermRec.Code, noRedemptionPermRec.Body.String())
	}

	itemBody := postAdminJSONWithPermission(t, mux, "/api/admin/redemption/items", "redemption:manage", `{"name":"Coffee","pointsCost":10,"stock":2}`, http.StatusOK)
	var itemResp struct {
		Data struct {
			ID         int64 `json:"id"`
			PointsCost int   `json:"pointsCost"`
			Stock      int   `json:"stock"`
		} `json:"data"`
	}
	if err := json.Unmarshal(itemBody, &itemResp); err != nil {
		t.Fatal(err)
	}
	if itemResp.Data.ID != 1 || itemResp.Data.PointsCost != 10 || itemResp.Data.Stock != 2 {
		t.Fatalf("expected admin item create: %s", string(itemBody))
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/admin/redemption/items/1", bytes.NewBufferString(`{"stock":3,"status":"active"}`))
	updateReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	updateRec := httptest.NewRecorder()
	mux.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected admin item update 200, got %d: %s", updateRec.Code, updateRec.Body.String())
	}

	orderBody := postJSON(t, mux, "/api/app/redemption/orders", userToken, `{"itemId":1}`, http.StatusOK)
	var orderResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(orderBody, &orderResp); err != nil {
		t.Fatal(err)
	}
	if orderResp.Data.ID != 1 || orderResp.Data.Status != "pending" {
		t.Fatalf("expected pending redemption order: %s", string(orderBody))
	}

	ordersBody := getAdminJSONWithPermission(t, mux, "/api/admin/redemption/orders", "redemption:manage", http.StatusOK)
	var ordersResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ordersBody, &ordersResp); err != nil {
		t.Fatal(err)
	}
	if len(ordersResp.Data.Items) != 1 || ordersResp.Data.Items[0].ID != 1 {
		t.Fatalf("expected admin redemption orders: %s", string(ordersBody))
	}

	postAdminJSONWithPermission(t, mux, "/api/admin/redemption/orders/1/review", "redemption:manage", `{"approve":true}`, http.StatusUnprocessableEntity)

	approveBody := postAdminJSONWithPermission(t, mux, "/api/admin/redemption/orders/1/review", "redemption:manage", `{"approve":true,"reason":"ok"}`, http.StatusOK)
	var approveResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(approveBody, &approveResp); err != nil {
		t.Fatal(err)
	}
	if approveResp.Data.ID != 1 || approveResp.Data.Status != "approved" {
		t.Fatalf("expected approved order: %s", string(approveBody))
	}
	approvedOrdersBody := getJSON(t, mux, "/api/app/redemption/orders/my?status=approved", userToken, http.StatusOK)
	var approvedOrdersResp struct {
		Data struct {
			Orders []struct {
				ID        int64  `json:"id"`
				StatusKey string `json:"statusKey"`
				Actions   []struct {
					Key string `json:"key"`
				} `json:"actions"`
			} `json:"orders"`
		} `json:"data"`
	}
	if err := json.Unmarshal(approvedOrdersBody, &approvedOrdersResp); err != nil {
		t.Fatal(err)
	}
	if len(approvedOrdersResp.Data.Orders) != 1 || approvedOrdersResp.Data.Orders[0].StatusKey != "approved" || !hasActionKey(approvedOrdersResp.Data.Orders[0].Actions, "detail") || !hasActionKey(approvedOrdersResp.Data.Orders[0].Actions, "logistics") {
		t.Fatalf("expected approved redemption order for app page: %s", string(approvedOrdersBody))
	}
	logisticsBody := getJSON(t, mux, "/api/app/profile/points/orders/1/logistics", userToken, http.StatusOK)
	var logisticsResp struct {
		Data struct {
			OrderID  int64 `json:"orderId"`
			Timeline []struct {
				ID string `json:"id"`
			} `json:"timeline"`
		} `json:"data"`
	}
	if err := json.Unmarshal(logisticsBody, &logisticsResp); err != nil {
		t.Fatal(err)
	}
	if logisticsResp.Data.OrderID != 1 || len(logisticsResp.Data.Timeline) == 0 || logisticsResp.Data.Timeline[0].ID != "approved" {
		t.Fatalf("expected redemption logistics timeline: %s", string(logisticsBody))
	}

	fulfillBody := postAdminJSONWithPermission(t, mux, "/api/admin/redemption/orders/1/review", "redemption:manage", `{"status":"fulfilled","reason":"sent"}`, http.StatusOK)
	var fulfillResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(fulfillBody, &fulfillResp); err != nil {
		t.Fatal(err)
	}
	if fulfillResp.Data.Status != "fulfilled" {
		t.Fatalf("expected fulfilled order: %s", string(fulfillBody))
	}

	rejectItemBody := postAdminJSONWithPermission(t, mux, "/api/admin/redemption/items", "redemption:manage", `{"name":"Ticket","pointsCost":10,"stock":1}`, http.StatusOK)
	var rejectItemResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rejectItemBody, &rejectItemResp); err != nil {
		t.Fatal(err)
	}
	rejectOrderBody := postJSON(t, mux, "/api/app/redemption/orders", userToken, `{"itemId":2}`, http.StatusOK)
	var rejectOrderResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rejectOrderBody, &rejectOrderResp); err != nil {
		t.Fatal(err)
	}
	postAdminJSONWithPermission(t, mux, "/api/admin/redemption/orders/2/review", "redemption:manage", `{"approve":false,"reason":"bad"}`, http.StatusOK)
	summaryBody := getJSON(t, mux, "/api/app/points/summary", userToken, http.StatusOK)
	var summaryResp struct {
		Data struct {
			AvailablePoints int `json:"availablePoints"`
		} `json:"data"`
	}
	if err := json.Unmarshal(summaryBody, &summaryResp); err != nil {
		t.Fatal(err)
	}
	if summaryResp.Data.AvailablePoints != 20 {
		t.Fatalf("expected rejected order refund points to 20: %s", string(summaryBody))
	}
}

func TestAdminDeliveryAndTestCaseHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	noDeliveryPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/delivery-documents", nil)
	noDeliveryPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noDeliveryPermRec, noDeliveryPermReq)
	if noDeliveryPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected delivery documents 403 without permission, got %d: %s", noDeliveryPermRec.Code, noDeliveryPermRec.Body.String())
	}

	postAdminJSONWithPermission(t, mux, "/api/admin/delivery-documents", "delivery:manage", `{"docType":"openapi","title":"Backend API handoff","status":"ready"}`, http.StatusUnprocessableEntity)
	docBody := postAdminJSONWithPermission(t, mux, "/api/admin/delivery-documents", "delivery:manage", `{"docType":"openapi","title":"Backend API handoff","status":"ready","reason":"reviewed"}`, http.StatusOK)
	var docResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(docBody, &docResp); err != nil {
		t.Fatal(err)
	}
	if docResp.Data.ID != 1 || docResp.Data.Status != "ready" {
		t.Fatalf("expected delivery document: %s", string(docBody))
	}

	docsBody := getAdminJSONWithPermission(t, mux, "/api/admin/delivery-documents", "delivery:manage", http.StatusOK)
	var docsResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(docsBody, &docsResp); err != nil {
		t.Fatal(err)
	}
	if len(docsResp.Data.Items) != 1 || docsResp.Data.Items[0].ID != 1 {
		t.Fatalf("expected delivery document list: %s", string(docsBody))
	}

	noCasePermReq := httptest.NewRequest(http.MethodGet, "/api/admin/test-cases", nil)
	noCasePermRec := httptest.NewRecorder()
	mux.ServeHTTP(noCasePermRec, noCasePermReq)
	if noCasePermRec.Code != http.StatusForbidden {
		t.Fatalf("expected test cases 403 without permission, got %d: %s", noCasePermRec.Code, noCasePermRec.Body.String())
	}

	postAdminJSONWithPermission(t, mux, "/api/admin/test-cases", "testcase:manage", `{"module":"E3.1","caseName":"bad priority","priority":"P3","expectedResult":"fail"}`, http.StatusUnprocessableEntity)
	caseBody := postAdminJSONWithPermission(t, mux, "/api/admin/test-cases", "testcase:manage", `{"module":"E3.1","caseName":"admin delivery smoke","priority":"P0","expectedResult":"pass"}`, http.StatusOK)
	var caseResp struct {
		Data struct {
			ID       int64  `json:"id"`
			Module   string `json:"module"`
			Priority string `json:"priority"`
		} `json:"data"`
	}
	if err := json.Unmarshal(caseBody, &caseResp); err != nil {
		t.Fatal(err)
	}
	if caseResp.Data.ID != 1 || caseResp.Data.Module != "E3.1" || caseResp.Data.Priority != "P0" {
		t.Fatalf("expected test case: %s", string(caseBody))
	}

	postAdminJSONWithPermission(t, mux, "/api/admin/test-runs", "testcase:manage", `{"caseId":1,"result":"passed"}`, http.StatusUnprocessableEntity)
	postAdminJSONWithPermission(t, mux, "/api/admin/test-runs", "testcase:manage", `{"caseId":99,"result":"passed","requestId":"req-missing-case"}`, http.StatusNotFound)
	runBody := postAdminJSONWithPermission(t, mux, "/api/admin/test-runs", "testcase:manage", `{"caseId":1,"result":"passed","actualResult":"ok","requestId":"req-test-1"}`, http.StatusOK)
	var runResp struct {
		Data struct {
			ID             int64  `json:"id"`
			CaseID         int64  `json:"caseId"`
			Result         string `json:"result"`
			RequestID      string `json:"requestId"`
			EvidenceFileID int64  `json:"evidenceFileId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(runBody, &runResp); err != nil {
		t.Fatal(err)
	}
	if runResp.Data.ID != 1 || runResp.Data.CaseID != 1 || runResp.Data.Result != "passed" || runResp.Data.RequestID != "req-test-1" {
		t.Fatalf("expected test run: %s", string(runBody))
	}

	evidenceRunBody := postAdminJSONWithPermission(t, mux, "/api/admin/test-runs", "testcase:manage", `{"caseId":1,"result":"blocked","actualResult":"see evidence","evidenceFileId":7}`, http.StatusOK)
	var evidenceRunResp struct {
		Data struct {
			ID             int64  `json:"id"`
			Result         string `json:"result"`
			EvidenceFileID int64  `json:"evidenceFileId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(evidenceRunBody, &evidenceRunResp); err != nil {
		t.Fatal(err)
	}
	if evidenceRunResp.Data.ID != 2 || evidenceRunResp.Data.Result != "blocked" || evidenceRunResp.Data.EvidenceFileID != 7 {
		t.Fatalf("expected test run with evidence file: %s", string(evidenceRunBody))
	}

	runsBody := getAdminJSONWithPermission(t, mux, "/api/admin/test-runs", "testcase:manage", http.StatusOK)
	var runsResp struct {
		Data struct {
			Items []struct {
				ID             int64  `json:"id"`
				Result         string `json:"result"`
				EvidenceFileID int64  `json:"evidenceFileId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(runsBody, &runsResp); err != nil {
		t.Fatal(err)
	}
	if len(runsResp.Data.Items) != 2 || runsResp.Data.Items[0].Result != "passed" || runsResp.Data.Items[1].EvidenceFileID != 7 {
		t.Fatalf("expected test run list: %s", string(runsBody))
	}
}

func TestAdminBehaviorEventsAndUserFavoritesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "admin-fav-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "admin-fav-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"favorite target","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/favorite", playerToken, `{}`, http.StatusOK)

	noFavoritePermReq := httptest.NewRequest(http.MethodGet, "/api/admin/users/2/favorites", nil)
	noFavoritePermRec := httptest.NewRecorder()
	mux.ServeHTTP(noFavoritePermRec, noFavoritePermReq)
	if noFavoritePermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin favorites 403 without permission, got %d: %s", noFavoritePermRec.Code, noFavoritePermRec.Body.String())
	}

	favoritesBody := getAdminJSONWithPermission(t, mux, "/api/admin/users/2/favorites", "user:read", http.StatusOK)
	var favoritesResp struct {
		Data struct {
			Items []struct {
				UserID int64 `json:"userId"`
				GameID int64 `json:"gameId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(favoritesBody, &favoritesResp); err != nil {
		t.Fatal(err)
	}
	if len(favoritesResp.Data.Items) != 1 || favoritesResp.Data.Items[0].UserID != 2 || favoritesResp.Data.Items[0].GameID != 1 {
		t.Fatalf("expected admin user favorites: %s", string(favoritesBody))
	}

	postJSON(t, mux, "/api/app/behavior/events", playerToken, `{"eventType":"search","targetType":"game","targetId":1,"keyword":"favorite"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/behavior/events", playerToken, `{"eventType":"share","targetType":"game","targetId":1}`, http.StatusOK)

	noBehaviorPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/behavior/events?userId=2&eventType=search", nil)
	noBehaviorPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noBehaviorPermRec, noBehaviorPermReq)
	if noBehaviorPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin behavior events 403 without permission, got %d: %s", noBehaviorPermRec.Code, noBehaviorPermRec.Body.String())
	}

	behaviorBody := getAdminJSONWithPermission(t, mux, "/api/admin/behavior/events?userId=2&eventType=search", "data:behavior:read", http.StatusOK)
	var behaviorResp struct {
		Data struct {
			Items []struct {
				UserID    int64  `json:"userId"`
				EventType string `json:"eventType"`
				TargetID  int64  `json:"targetId"`
				Keyword   string `json:"keyword"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(behaviorBody, &behaviorResp); err != nil {
		t.Fatal(err)
	}
	if len(behaviorResp.Data.Items) != 1 || behaviorResp.Data.Items[0].UserID != 2 || behaviorResp.Data.Items[0].EventType != "search" || behaviorResp.Data.Items[0].TargetID != 1 || behaviorResp.Data.Items[0].Keyword != "favorite" {
		t.Fatalf("expected filtered behavior events: %s", string(behaviorBody))
	}
}

func TestConnectionsAndProfilesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	inviteStore := invites.NewStore()
	authService := auth.NewService(users.NewStore(), inviteStore, auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "connection-creator")
	completeIdentityForTest(t, mux, creatorToken)
	inviteStore.UpsertCode("OWNER1", 1, 10)
	inviteeToken := loginForTestWithCodeAndInvite(t, mux, "connection-invitee", "OWNER1")
	if inviteeToken == "" {
		t.Fatal("expected invitee token")
	}
	memberToken := loginForTestWithCode(t, mux, "connection-member")
	completeIdentityForTest(t, mux, memberToken)
	guideToken := loginForTestWithCode(t, mux, "connection-guide")
	completeIdentityForTest(t, mux, guideToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"connection game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", memberToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "profile-flow", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", memberToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}

	postJSON(t, mux, "/api/app/experts/me/skills", creatorToken, `{"skillTree":["boardgame"],"serviceTags":["host"]}`, http.StatusForbidden)
	server.profiles.GrantRole(1, "expert")
	expertBody := putJSON(t, mux, "/api/app/experts/me/skills", creatorToken, `{"skillTree":["boardgame","social"],"serviceTags":["host"],"caseFileIds":[1]}`, http.StatusOK)
	var expertResp struct {
		Data struct {
			UserID       int64    `json:"userId"`
			SkillTree    []string `json:"skillTree"`
			Completeness int      `json:"completeness"`
		} `json:"data"`
	}
	if err := json.Unmarshal(expertBody, &expertResp); err != nil {
		t.Fatal(err)
	}
	if expertResp.Data.UserID != 1 || len(expertResp.Data.SkillTree) != 2 || expertResp.Data.Completeness != 100 {
		t.Fatalf("unexpected expert profile: %s", string(expertBody))
	}
	getJSON(t, mux, "/api/app/experts/me/skills", creatorToken, http.StatusOK)
	rolesBody := getJSON(t, mux, "/api/app/roles/my", creatorToken, http.StatusOK)
	var rolesResp struct {
		Data struct {
			Roles         []string          `json:"roles"`
			RoleStatusMap map[string]string `json:"roleStatusMap"`
			Applications  []struct {
				RoleCode string `json:"roleCode"`
				Status   string `json:"status"`
			} `json:"applications"`
			GuideQualification struct {
				GuideOpenStatus string `json:"guideOpenStatus"`
			} `json:"guideQualification"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rolesBody, &rolesResp); err != nil {
		t.Fatal(err)
	}
	if len(rolesResp.Data.Roles) != 2 || rolesResp.Data.RoleStatusMap["expert"] != "approved" || rolesResp.Data.RoleStatusMap["guide"] != "none" || rolesResp.Data.GuideQualification.GuideOpenStatus == "" {
		t.Fatalf("unexpected roles snapshot: %s", string(rolesBody))
	}

	putJSON(t, mux, "/api/app/guides/me/resources", guideToken, `{"resourceTags":["venue"],"industryTags":["entertainment"],"cityCodes":["110100"],"connectionScale":"100-500"}`, http.StatusForbidden)
	server.profiles.GrantRole(4, "guide")
	guideBody := putJSON(t, mux, "/api/app/guides/me/resources", guideToken, `{"resourceTags":["venue"],"industryTags":["entertainment"],"cityCodes":["110100"],"connectionScale":"100-500"}`, http.StatusOK)
	var guideResp struct {
		Data struct {
			UserID          int64    `json:"userId"`
			ResourceTags    []string `json:"resourceTags"`
			ConnectionScale string   `json:"connectionScale"`
			Completeness    int      `json:"completeness"`
		} `json:"data"`
	}
	if err := json.Unmarshal(guideBody, &guideResp); err != nil {
		t.Fatal(err)
	}
	if guideResp.Data.UserID != 4 || len(guideResp.Data.ResourceTags) != 1 || guideResp.Data.ConnectionScale == "" || guideResp.Data.Completeness != 100 {
		t.Fatalf("unexpected guide profile: %s", string(guideBody))
	}

	server.connections.UpsertPair(4, 1, "guide_match", "guide_match", 88, 3)
	connectionsBody := getJSON(t, mux, "/api/app/connections/my", creatorToken, http.StatusOK)
	var connectionsResp struct {
		Data struct {
			Items []connectionTestItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(connectionsBody, &connectionsResp); err != nil {
		t.Fatal(err)
	}
	if !hasConnectionSource(connectionsResp.Data.Items, "invite") || !hasConnectionSource(connectionsResp.Data.Items, "co_game") || !hasConnectionSource(connectionsResp.Data.Items, "guide_match") {
		t.Fatalf("expected invite, co_game and guide_match connections without sensitive fields: %s", string(connectionsBody))
	}
	if bytes.Contains(connectionsBody, []byte("phone")) || bytes.Contains(connectionsBody, []byte("idCard")) {
		t.Fatalf("connections leaked sensitive fields: %s", string(connectionsBody))
	}

	connectionID := firstConnectionID(connectionsResp.Data.Items)
	followBody := postJSON(t, mux, "/api/app/connections/"+int64String(connectionID)+"/follow-up", creatorToken, `{"followType":"note","content":"follow note"}`, http.StatusOK)
	var followResp struct {
		Data struct {
			ConnectionID   int64  `json:"connectionId"`
			OperatorUserID int64  `json:"operatorUserId"`
			Content        string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(followBody, &followResp); err != nil {
		t.Fatal(err)
	}
	if followResp.Data.ConnectionID != connectionID || followResp.Data.OperatorUserID != 1 || followResp.Data.Content == "" {
		t.Fatalf("unexpected follow log: %s", string(followBody))
	}

	postJSON(t, mux, "/api/app/connections/"+int64String(connectionID)+"/follow-logs", creatorToken, `{"followType":"note","content":"follow logs alias"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/connections/"+int64String(connectionID)+"/follow-up", guideToken, `{"followType":"guide","content":"guide follow"}`, http.StatusOK)
}

func TestProfileInviteCenterHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "profile-invite-owner")
	completeIdentityForTest(t, mux, creatorToken)
	memberToken := loginForTestWithCode(t, mux, "profile-invite-member")
	completeIdentityForTest(t, mux, memberToken)
	creatorID := currentUserIDForTest(t, mux, creatorToken)
	memberID := currentUserIDForTest(t, mux, memberToken)
	server.connections.UpsertPair(creatorID, memberID, "invite", "invite", 1, 4)

	overviewBody := getJSON(t, mux, "/api/app/profile/service-center/invite/overview", creatorToken, http.StatusOK)
	var overviewResp struct {
		Data struct {
			Profile struct {
				Name string `json:"name"`
				Desc string `json:"desc"`
			} `json:"profile"`
			Metrics []struct {
				Label string `json:"label"`
			} `json:"metrics"`
			Actions []struct {
				Key        string `json:"key"`
				InviteCode string `json:"inviteCode"`
			} `json:"actions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(overviewBody, &overviewResp); err != nil {
		t.Fatal(err)
	}
	if overviewResp.Data.Profile.Name == "" || len(overviewResp.Data.Metrics) == 0 || len(overviewResp.Data.Actions) != 3 || overviewResp.Data.Actions[0].InviteCode == "" {
		t.Fatalf("expected invite overview payload: %s", string(overviewBody))
	}

	networkBody := getJSON(t, mux, "/api/app/profile/service-center/invite/network", creatorToken, http.StatusOK)
	var networkResp struct {
		Data struct {
			Summary []struct {
				Value string `json:"value"`
			} `json:"summary"`
			Members []struct {
				ID string `json:"id"`
			} `json:"members"`
		} `json:"data"`
	}
	if err := json.Unmarshal(networkBody, &networkResp); err != nil {
		t.Fatal(err)
	}
	if len(networkResp.Data.Summary) == 0 || len(networkResp.Data.Members) != 1 || networkResp.Data.Members[0].ID == "" {
		t.Fatalf("expected invite network payload: %s", string(networkBody))
	}

	recordsBody := getJSON(t, mux, "/api/app/profile/service-center/invite/records?role=referred&status=all", creatorToken, http.StatusOK)
	var recordsResp struct {
		Data struct {
			Records []struct {
				ID        string `json:"id"`
				StatusKey string `json:"statusKey"`
			} `json:"records"`
			Filters []struct {
				Key string `json:"key"`
			} `json:"filters"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recordsBody, &recordsResp); err != nil {
		t.Fatal(err)
	}
	if len(recordsResp.Data.Records) != 1 || recordsResp.Data.Records[0].ID == "" || len(recordsResp.Data.Filters) == 0 {
		t.Fatalf("expected invite records payload: %s", string(recordsBody))
	}

	rankingBody := getJSON(t, mux, "/api/app/profile/service-center/invite/ranking?period=week&type=inviteCount", creatorToken, http.StatusOK)
	var rankingResp struct {
		Data struct {
			Members []struct {
				Rank int `json:"rank"`
			} `json:"members"`
			RankTypes []struct {
				Key string `json:"key"`
			} `json:"rankTypes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rankingBody, &rankingResp); err != nil {
		t.Fatal(err)
	}
	if len(rankingResp.Data.Members) != 1 || rankingResp.Data.Members[0].Rank != 1 || len(rankingResp.Data.RankTypes) == 0 {
		t.Fatalf("expected invite ranking payload: %s", string(rankingBody))
	}

	incomeBody := getJSON(t, mux, "/api/app/profile/service-center/invite/income", creatorToken, http.StatusOK)
	var incomeResp struct {
		Data struct {
			TrendSeries []struct {
				Month string `json:"month"`
			} `json:"trendSeries"`
			Metrics []struct {
				Label string `json:"label"`
			} `json:"metrics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(incomeBody, &incomeResp); err != nil {
		t.Fatal(err)
	}
	if len(incomeResp.Data.TrendSeries) == 0 || len(incomeResp.Data.Metrics) == 0 {
		t.Fatalf("expected invite income payload: %s", string(incomeBody))
	}

	memberBody := getJSON(t, mux, "/api/app/profile/service-center/invite/member-detail?memberId="+strconv.FormatInt(memberID, 10), creatorToken, http.StatusOK)
	var memberResp struct {
		Data struct {
			Member struct {
				Name string `json:"name"`
			} `json:"member"`
			Stats []struct {
				Label string `json:"label"`
			} `json:"stats"`
		} `json:"data"`
	}
	if err := json.Unmarshal(memberBody, &memberResp); err != nil {
		t.Fatal(err)
	}
	if memberResp.Data.Member.Name == "" || len(memberResp.Data.Stats) == 0 {
		t.Fatalf("expected invite member detail payload: %s", string(memberBody))
	}
	getJSON(t, mux, "/api/app/profile/service-center/invite/member-detail?memberId=999", creatorToken, http.StatusNotFound)
}

func TestProfileHomeHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "profile-home-user")
	completeIdentityForTest(t, mux, token)

	homeBody := getJSON(t, mux, "/api/app/profile/home", token, http.StatusOK)
	var homeResp struct {
		Data struct {
			User struct {
				Nickname     string `json:"nickname"`
				MemberLevel  string `json:"memberLevel"`
				RoleLevel    string `json:"roleLevel"`
				AvatarText   string `json:"avatarText"`
				AvatarURL    string `json:"avatarUrl"`
				AvatarFileID int64  `json:"avatarFileId"`
			} `json:"user"`
			Stats []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"stats"`
			Assets []struct {
				Key   string `json:"key"`
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"assets"`
			ServiceSections []struct {
				Title string `json:"title"`
				Items []struct {
					Title   string `json:"title"`
					Route   string `json:"route"`
					Enabled bool   `json:"enabled"`
				} `json:"items"`
			} `json:"serviceSections"`
			Summary CurrentUserSummaryDTO `json:"summary"`
		} `json:"data"`
	}
	if err := json.Unmarshal(homeBody, &homeResp); err != nil {
		t.Fatal(err)
	}
	if homeResp.Data.User.Nickname == "" || homeResp.Data.User.MemberLevel == "" || homeResp.Data.User.RoleLevel == "" || homeResp.Data.User.AvatarText == "" {
		t.Fatalf("expected profile home user block: %s", string(homeBody))
	}
	if homeResp.Data.User.MemberLevel == "none" {
		t.Fatalf("expected user-facing member level instead of none: %s", string(homeBody))
	}
	if len(homeResp.Data.Stats) != 4 || len(homeResp.Data.Assets) != 3 || len(homeResp.Data.ServiceSections) == 0 || len(homeResp.Data.ServiceSections[0].Items) == 0 {
		t.Fatalf("expected profile home summary blocks: %s", string(homeBody))
	}
	if homeResp.Data.ServiceSections[0].Items[0].Route == "" || !homeResp.Data.ServiceSections[0].Items[0].Enabled {
		t.Fatalf("expected enabled service route in profile home: %s", string(homeBody))
	}
	if homeResp.Data.Summary.User.ID == 0 {
		t.Fatalf("expected embedded legacy summary for compatibility: %s", string(homeBody))
	}
}

func TestCountInProgressUserGamesIncludesCreatedAndJoinedWithoutDuplicates(t *testing.T) {
	items := []games.Game{
		{ID: 1, CreatorUserID: 42, Status: "in_progress"},
		{ID: 2, CreatorUserID: 7, Status: "pending_confirm"},
		{ID: 3, CreatorUserID: 7, Status: "completed"},
		{ID: 4, CreatorUserID: 7, Status: "in_progress"},
	}
	memberships := map[int64]bool{1: true, 2: true, 3: true}

	got := countInProgressUserGames(42, items, func(gameID int64, userID int64) bool {
		return userID == 42 && memberships[gameID]
	})
	if got != 1 {
		t.Fatalf("expected 1 unique in-progress game, got %d", got)
	}
}

func TestProfileAssetsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "profile-assets-user")
	completeIdentityForTest(t, mux, token)
	_, _, err := server.points.Grant(1, 1000, "seed", 1, "测试积分")
	if err != nil {
		t.Fatal(err)
	}
	item, err := server.redemption.CreateItem(redemption.CreateItemRequest{Name: "测试权益", PointsCost: 100, Stock: 3})
	if err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/redemption/orders", token, `{"itemId":`+strconv.FormatInt(item.ID, 10)+`}`, http.StatusOK)

	body := getJSON(t, mux, "/api/app/profile/assets", token, http.StatusOK)
	var resp struct {
		Data struct {
			Overview struct {
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"overview"`
			AssetStats []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"assetStats"`
			OrderStatuses []struct {
				Key   string `json:"key"`
				Count int    `json:"count"`
				Route string `json:"route"`
			} `json:"orderStatuses"`
			RecentOrders []struct {
				ID string `json:"id"`
			} `json:"recentOrders"`
			BankCards struct {
				SummaryText string `json:"summaryText"`
			} `json:"bankCards"`
			FAQLinks []struct {
				Key string `json:"key"`
			} `json:"faqLinks"`
			ConfigVersion string `json:"configVersion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Overview.Value == "" || len(resp.Data.AssetStats) != 3 || len(resp.Data.OrderStatuses) != 5 || len(resp.Data.FAQLinks) == 0 {
		t.Fatalf("expected profile assets aggregate: %s", string(body))
	}
	if resp.Data.Overview.Label != "总资产（元）" || resp.Data.ConfigVersion == "" {
		t.Fatalf("expected profile asset manage config fields: %s", string(body))
	}
	if len(resp.Data.RecentOrders) != 1 || resp.Data.BankCards.SummaryText == "" {
		t.Fatalf("expected recent orders and bank card state: %s", string(body))
	}
	if resp.Data.OrderStatuses[1].Key != "processing" || resp.Data.OrderStatuses[1].Count != 1 || resp.Data.OrderStatuses[1].Route == "" {
		t.Fatalf("expected processing order status count and route: %s", string(body))
	}
}

func TestAdminConnectionsAndProfilesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	expertToken := loginForTestWithCode(t, mux, "admin-profile-expert")
	completeIdentityForTest(t, mux, expertToken)
	guideToken := loginForTestWithCode(t, mux, "admin-profile-guide")
	completeIdentityForTest(t, mux, guideToken)
	peerToken := loginForTestWithCode(t, mux, "admin-profile-peer")
	completeIdentityForTest(t, mux, peerToken)
	if _, err := authService.UpdateProfile(1, "profile-expert-alpha", "", 0); err != nil {
		t.Fatalf("seed expert nickname failed: %v", err)
	}
	if _, err := authService.UpdateProfile(2, "profile-guide-beta", "", 0); err != nil {
		t.Fatalf("seed guide nickname failed: %v", err)
	}

	server.profiles.GrantRole(1, "expert")
	server.profiles.GrantRole(2, "guide")
	putJSON(t, mux, "/api/app/experts/me/skills", expertToken, `{"skillTree":["boardgame","social"],"serviceTags":["host"],"caseFileIds":[1]}`, http.StatusOK)
	putJSON(t, mux, "/api/app/guides/me/resources", guideToken, `{"resourceTags":["venue"],"industryTags":["entertainment"],"cityCodes":["110100"],"connectionScale":"100-500"}`, http.StatusOK)
	server.connections.UpsertPair(1, 2, "guide_match", "guide_match", 88, 3)
	server.connections.UpsertPair(1, 3, "co_game", "co_game", 1, 2)

	noConnectionPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/connections", nil)
	noConnectionPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noConnectionPermRec, noConnectionPermReq)
	if noConnectionPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin connections 403 without permission, got %d: %s", noConnectionPermRec.Code, noConnectionPermRec.Body.String())
	}

	adminConnectionsBody := getAdminJSONWithPermission(t, mux, "/api/admin/connections", "connection:read", http.StatusOK)
	var adminConnectionsResp struct {
		Data struct {
			Items []struct {
				UserID     int64  `json:"userId"`
				SourceType string `json:"sourceType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminConnectionsBody, &adminConnectionsResp); err != nil {
		t.Fatal(err)
	}
	if len(adminConnectionsResp.Data.Items) < 4 {
		t.Fatalf("expected bidirectional admin connections: %s", string(adminConnectionsBody))
	}

	adminUserConnectionsBody := getAdminJSONWithPermission(t, mux, "/api/admin/users/1/connections", "connection:read", http.StatusOK)
	var adminUserConnectionsResp struct {
		Data struct {
			Items []connectionTestItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminUserConnectionsBody, &adminUserConnectionsResp); err != nil {
		t.Fatal(err)
	}
	if !hasConnectionSource(adminUserConnectionsResp.Data.Items, "guide_match") || !hasConnectionSource(adminUserConnectionsResp.Data.Items, "co_game") {
		t.Fatalf("expected admin user connections by source: %s", string(adminUserConnectionsBody))
	}

	noProfilePermReq := httptest.NewRequest(http.MethodGet, "/api/admin/experts/1/skills", nil)
	noProfilePermRec := httptest.NewRecorder()
	mux.ServeHTTP(noProfilePermRec, noProfilePermReq)
	if noProfilePermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin expert skills 403 without permission, got %d: %s", noProfilePermRec.Code, noProfilePermRec.Body.String())
	}
	noProfileSearchPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/profile-users?keyword=profile", nil)
	noProfileSearchPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noProfileSearchPermRec, noProfileSearchPermReq)
	if noProfileSearchPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin profile user search 403 without permission, got %d: %s", noProfileSearchPermRec.Code, noProfileSearchPermRec.Body.String())
	}

	adminProfileUsersByNicknameBody := getAdminJSONWithPermission(t, mux, "/api/admin/profile-users?keyword=profile-expert", "profile:read", http.StatusOK)
	var adminProfileUsersByNicknameResp struct {
		Data struct {
			Items []struct {
				ID          int64    `json:"id"`
				Nickname    string   `json:"nickname"`
				MatchFields []string `json:"matchFields"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminProfileUsersByNicknameBody, &adminProfileUsersByNicknameResp); err != nil {
		t.Fatal(err)
	}
	if len(adminProfileUsersByNicknameResp.Data.Items) != 1 || adminProfileUsersByNicknameResp.Data.Items[0].ID != 1 || !hasString(adminProfileUsersByNicknameResp.Data.Items[0].MatchFields, "昵称") {
		t.Fatalf("expected profile user search by nickname: %s", string(adminProfileUsersByNicknameBody))
	}

	adminProfileUsersByPhoneBody := getAdminJSONWithPermission(t, mux, "/api/admin/profile-users?keyword="+identityService.Status(1).PhoneMasked, "profile:read", http.StatusOK)
	var adminProfileUsersByPhoneResp struct {
		Data struct {
			Items []struct {
				ID          int64    `json:"id"`
				MatchFields []string `json:"matchFields"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminProfileUsersByPhoneBody, &adminProfileUsersByPhoneResp); err != nil {
		t.Fatal(err)
	}
	if len(adminProfileUsersByPhoneResp.Data.Items) != 1 || adminProfileUsersByPhoneResp.Data.Items[0].ID != 1 || !hasString(adminProfileUsersByPhoneResp.Data.Items[0].MatchFields, "手机号") {
		t.Fatalf("expected profile user search by phone: %s", string(adminProfileUsersByPhoneBody))
	}

	adminProfileUsersByIDCardBody := getAdminJSONWithPermission(t, mux, "/api/admin/profile-users?keyword=110101199001011234", "profile:read", http.StatusOK)
	var adminProfileUsersByIDCardResp struct {
		Data struct {
			Items []struct {
				IDCardMasked string   `json:"idCardMasked"`
				MatchFields  []string `json:"matchFields"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminProfileUsersByIDCardBody, &adminProfileUsersByIDCardResp); err != nil {
		t.Fatal(err)
	}
	if len(adminProfileUsersByIDCardResp.Data.Items) < 3 || !hasString(adminProfileUsersByIDCardResp.Data.Items[0].MatchFields, "身份证号") {
		t.Fatalf("expected profile user search by id card: %s", string(adminProfileUsersByIDCardBody))
	}
	if !hasString(profileUserMatchFields(users.User{ID: 99, PhoneMasked: "138****8000"}, "", "13800138000"), "手机号") {
		t.Fatal("expected full phone keyword to match masked phone")
	}

	adminExpertBody := getAdminJSONWithPermission(t, mux, "/api/admin/experts/1/skills", "profile:read", http.StatusOK)
	var adminExpertResp struct {
		Data struct {
			UserID       int64    `json:"userId"`
			SkillTree    []string `json:"skillTree"`
			Completeness int      `json:"completeness"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminExpertBody, &adminExpertResp); err != nil {
		t.Fatal(err)
	}
	if adminExpertResp.Data.UserID != 1 || len(adminExpertResp.Data.SkillTree) != 2 || adminExpertResp.Data.Completeness != 100 {
		t.Fatalf("expected admin expert skill profile: %s", string(adminExpertBody))
	}

	adminGuideBody := getAdminJSONWithPermission(t, mux, "/api/admin/guides/2/resources", "profile:read", http.StatusOK)
	var adminGuideResp struct {
		Data struct {
			UserID          int64  `json:"userId"`
			ConnectionScale string `json:"connectionScale"`
			Completeness    int    `json:"completeness"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminGuideBody, &adminGuideResp); err != nil {
		t.Fatal(err)
	}
	if adminGuideResp.Data.UserID != 2 || adminGuideResp.Data.ConnectionScale != "100-500" || adminGuideResp.Data.Completeness != 100 {
		t.Fatalf("expected admin guide resource profile: %s", string(adminGuideBody))
	}

	_ = peerToken
}

func TestProfileUserMatchFieldsFuzzyInputs(t *testing.T) {
	user := users.User{
		ID:          99,
		Nickname:    "profile-expert-alpha",
		PhoneMasked: "138****8000",
	}
	if !hasString(profileUserMatchFields(user, "110***********1234", "profile-expert"), "昵称") {
		t.Fatal("expected nickname keyword to match")
	}
	if !hasString(profileUserMatchFields(user, "110***********1234", "13800138000"), "手机号") {
		t.Fatal("expected full phone keyword to match masked phone")
	}
	if !hasString(profileUserMatchFields(user, "110***********1234", "110101199001011234"), "身份证号") {
		t.Fatal("expected full id card keyword to match masked id card")
	}
	if !hasString(profileUserMatchFields(user, "110***********1234", "9"), "用户编号") {
		t.Fatal("expected user id keyword to match")
	}
}

func TestRoleApplicationAndGuideQualificationHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	userToken := loginForTestWithCode(t, mux, "role-app-user")
	completeIdentityForTest(t, mux, userToken)
	adminToken := adminLoginForTest(t, mux)

	postJSON(t, mux, "/api/app/role-applications", userToken, `{"roleCode":"guide","reason":"want to help"}`, http.StatusConflict)
	statusConfigBody := getJSON(t, mux, "/api/app/role-applications/status-config", userToken, http.StatusOK)
	var statusConfigResp struct {
		Data struct {
			RoleMeta map[string]struct {
				RoleName string `json:"roleName"`
			} `json:"roleMeta"`
			Texts           map[string]string `json:"texts"`
			PendingTimeline []struct {
				Title string `json:"title"`
			} `json:"pendingTimeline"`
			ApprovedActions []struct {
				RouteKey string `json:"routeKey"`
			} `json:"approvedActions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(statusConfigBody, &statusConfigResp); err != nil {
		t.Fatal(err)
	}
	if statusConfigResp.Data.RoleMeta["guide"].RoleName == "" || statusConfigResp.Data.Texts["pendingTitle"] == "" || len(statusConfigResp.Data.PendingTimeline) == 0 || len(statusConfigResp.Data.ApprovedActions) == 0 {
		t.Fatalf("expected role status page config: %s", string(statusConfigBody))
	}
	qualificationBody := getJSON(t, mux, "/api/app/guides/qualification/me", userToken, http.StatusOK)
	var qualificationResp struct {
		Data struct {
			Qualification struct {
				GuideOpenStatus string `json:"guideOpenStatus"`
			} `json:"qualification"`
			Rules []struct {
				ID int64 `json:"id"`
			} `json:"rules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(qualificationBody, &qualificationResp); err != nil {
		t.Fatal(err)
	}
	if qualificationResp.Data.Qualification.GuideOpenStatus != "waiting_condition" || len(qualificationResp.Data.Rules) != 1 {
		t.Fatalf("expected waiting condition and one guide rule: %s", string(qualificationBody))
	}
	postAdminJSONWithPermission(t, mux, "/api/admin/guide-qualification-rules", "role:update", `{"userId":1,"conditionMet":true,"paymentMet":false}`, http.StatusOK)
	postJSON(t, mux, "/api/app/guides/apply", userToken, `{"reason":"want to help"}`, http.StatusConflict)
	postJSON(t, mux, "/api/app/role-applications", userToken, `{"roleCode":"guide","reason":"want to help"}`, http.StatusConflict)
	postAdminJSONWithPermission(t, mux, "/api/admin/guide-qualification-rules", "role:update", `{"userId":1,"conditionMet":true,"paymentMet":true}`, http.StatusOK)
	appBody := postJSON(t, mux, "/api/app/guides/apply", userToken, `{"reason":"want to help","abilityDescription":"hosted games","proofFileIds":[7,8]}`, http.StatusOK)
	var appResp struct {
		Data struct {
			ID                 int64   `json:"id"`
			RoleCode           string  `json:"roleCode"`
			Status             string  `json:"status"`
			AbilityDescription string  `json:"abilityDescription"`
			ProofFileIDs       []int64 `json:"proofFileIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(appBody, &appResp); err != nil {
		t.Fatal(err)
	}
	if appResp.Data.ID == 0 || appResp.Data.RoleCode != "guide" || appResp.Data.Status != "pending" || appResp.Data.AbilityDescription != "hosted games" || len(appResp.Data.ProofFileIDs) != 2 {
		t.Fatalf("expected pending guide application: %s", string(appBody))
	}

	myBody := getJSON(t, mux, "/api/app/role-applications/my", userToken, http.StatusOK)
	var myResp struct {
		Data struct {
			Items []struct {
				RoleCode string `json:"roleCode"`
				Status   string `json:"status"`
			} `json:"items"`
			PageConfig struct {
				PageTitle string            `json:"pageTitle"`
				Texts     map[string]string `json:"texts"`
			} `json:"pageConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(myBody, &myResp); err != nil {
		t.Fatal(err)
	}
	if len(myResp.Data.Items) != 1 || myResp.Data.Items[0].RoleCode != "guide" || myResp.Data.Items[0].Status != "pending" {
		t.Fatalf("expected my role applications: %s", string(myBody))
	}
	if myResp.Data.PageConfig.PageTitle == "" || myResp.Data.PageConfig.Texts["submitButtonText"] == "" {
		t.Fatalf("expected role application page config: %s", string(myBody))
	}

	adminListBody := getAdminJSONWithPermission(t, mux, "/api/admin/audits/role-applications", "role:view", http.StatusOK)
	var adminListResp struct {
		Data struct {
			Items []struct {
				RoleCode string `json:"roleCode"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminListBody, &adminListResp); err != nil {
		t.Fatal(err)
	}
	if len(adminListResp.Data.Items) == 0 {
		t.Fatalf("expected admin role applications list: %s", string(adminListBody))
	}

	rulesBody := getAdminJSONWithPermission(t, mux, "/api/admin/guides/qualification-rules", "role:view", http.StatusOK)
	var rulesResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rulesBody, &rulesResp); err != nil {
		t.Fatal(err)
	}
	if len(rulesResp.Data.Items) == 0 {
		t.Fatalf("expected guide qualification rules: %s", string(rulesBody))
	}
	putJSON(t, mux, "/api/admin/guides/qualification-rules/1", adminToken, `{"minCreditScore":80,"paymentRequired":true}`, http.StatusOK)

	reviewBody := postAdminJSONWithPermission(t, mux, "/api/admin/audits/role-applications/1/review", "role:update", `{"approve":true,"remark":"ok"}`, http.StatusOK)
	var reviewResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reviewBody, &reviewResp); err != nil {
		t.Fatal(err)
	}
	if reviewResp.Data.Status != "approved" {
		t.Fatalf("expected approved application: %s", string(reviewBody))
	}

	adminGuideBody := getAdminJSONWithPermission(t, mux, "/api/admin/guide-qualification-rules?userId=1", "role:view", http.StatusOK)
	var guideResp struct {
		Data struct {
			ConditionMet    bool   `json:"conditionMet"`
			PaymentMet      bool   `json:"paymentMet"`
			GuideOpenStatus string `json:"guideOpenStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminGuideBody, &guideResp); err != nil {
		t.Fatal(err)
	}
	if !guideResp.Data.ConditionMet || !guideResp.Data.PaymentMet || guideResp.Data.GuideOpenStatus != "opened" {
		t.Fatalf("expected opened guide qualification: %s", string(adminGuideBody))
	}

	postJSON(t, mux, "/api/app/role-applications", userToken, `{"roleCode":"guide","reason":"duplicate active guide"}`, http.StatusConflict)
	expertBody := postJSON(t, mux, "/api/app/role-applications", userToken, `{"roleCode":"expert","reason":"expert application"}`, http.StatusOK)
	var expertResp struct {
		Data struct {
			ID       int64  `json:"id"`
			RoleCode string `json:"roleCode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(expertBody, &expertResp); err != nil {
		t.Fatal(err)
	}
	if expertResp.Data.ID == 0 || expertResp.Data.RoleCode != "expert" {
		t.Fatalf("expected separate expert application: %s", string(expertBody))
	}
	postAdminJSONWithPermission(t, mux, "/api/admin/audits/role-applications/"+strconv.FormatInt(expertResp.Data.ID, 10)+"/review", "role:update", `{"approve":false,"remark":"retry later"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/role-applications", userToken, `{"roleCode":"expert","reason":"retry too soon"}`, http.StatusConflict)
}

func TestBehaviorEventHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	token := loginForTestWithCode(t, mux, "behavior-user")
	postJSON(t, mux, "/api/app/behavior/events", token, `{"targetType":"game","targetId":1}`, http.StatusUnprocessableEntity)
	completeIdentityForTest(t, mux, token)
	postJSON(t, mux, "/api/app/games", token, `{"title":"behavior game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	getJSON(t, mux, "/api/app/games", token, http.StatusOK)

	postJSON(t, mux, "/api/app/behavior/events", token, `{"eventCode":"search","businessType":"game","businessId":1,"pagePath":"/pages/home/index","keyword":"party","source":"app","device":"ios","extra":{"cityCode":"110100"}}`, http.StatusUnprocessableEntity)
	body := postJSON(t, mux, "/api/app/behavior/events", token, `{"eventType":"search","eventCode":"search","businessType":"game","businessId":1,"pagePath":"/pages/home/index","keyword":"party","source":"app","device":"ios","extra":{"cityCode":"110100"}}`, http.StatusOK)
	var resp struct {
		Data struct {
			UserID       int64  `json:"userId"`
			EventType    string `json:"eventType"`
			EventCode    string `json:"eventCode"`
			TargetType   string `json:"targetType"`
			BusinessType string `json:"businessType"`
			TargetID     int64  `json:"targetId"`
			BusinessID   int64  `json:"businessId"`
			PagePath     string `json:"pagePath"`
			Keyword      string `json:"keyword"`
			Source       string `json:"source"`
			Device       string `json:"device"`
			IP           string `json:"ip"`
			OccurredAt   string `json:"occurredAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.UserID != 1 || resp.Data.EventType != "search" || resp.Data.EventCode != "search" || resp.Data.TargetType != "game" || resp.Data.BusinessType != "game" || resp.Data.TargetID != 1 || resp.Data.BusinessID != 1 || resp.Data.PagePath == "" || resp.Data.Keyword != "party" || resp.Data.Source != "app" || resp.Data.Device != "ios" || resp.Data.IP == "" || resp.Data.OccurredAt == "" {
		t.Fatalf("unexpected behavior event: %s", string(body))
	}

	anonymousBody := postJSON(t, mux, "/api/app/behavior/events", "", `{"eventType":"share","targetType":"game","targetId":2}`, http.StatusOK)
	var anonymousResp struct {
		Data struct {
			UserID    int64  `json:"userId"`
			EventType string `json:"eventType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(anonymousBody, &anonymousResp); err != nil {
		t.Fatal(err)
	}
	if anonymousResp.Data.UserID != 0 || anonymousResp.Data.EventType != "share" {
		t.Fatalf("unexpected anonymous behavior event: %s", string(anonymousBody))
	}

	adminToken := adminLoginForTest(t, mux)
	behaviorReq := httptest.NewRequest(http.MethodGet, "/api/admin/behavior-logs", nil)
	behaviorReq.Header.Set("Authorization", "Bearer "+adminToken)
	behaviorRec := httptest.NewRecorder()
	mux.ServeHTTP(behaviorRec, behaviorReq)
	if behaviorRec.Code != http.StatusOK {
		t.Fatalf("expected behavior logs 200, got %d: %s", behaviorRec.Code, behaviorRec.Body.String())
	}
	var behaviorResp struct {
		Data struct {
			Items []struct {
				EventType string `json:"eventType"`
				TargetID  int64  `json:"targetId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(behaviorRec.Body.Bytes(), &behaviorResp); err != nil {
		t.Fatal(err)
	}
	if !hasBehaviorEvent(behaviorResp.Data.Items, "search", 1) || !hasBehaviorEvent(behaviorResp.Data.Items, "share", 2) || !hasBehaviorEvent(behaviorResp.Data.Items, "browse_games", 0) {
		t.Fatalf("expected search and share behavior logs: %s", behaviorRec.Body.String())
	}

	filteredBody := getAdminJSONWithPermission(t, mux, "/api/admin/behavior/events?eventCode=search", "data:behavior:read", http.StatusOK)
	var filteredResp struct {
		Data struct {
			Items []struct {
				EventCode    string `json:"eventCode"`
				BusinessType string `json:"businessType"`
				BusinessID   int64  `json:"businessId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(filteredBody, &filteredResp); err != nil {
		t.Fatal(err)
	}
	if len(filteredResp.Data.Items) != 1 || filteredResp.Data.Items[0].EventCode != "search" || filteredResp.Data.Items[0].BusinessType != "game" || filteredResp.Data.Items[0].BusinessID != 1 {
		t.Fatalf("expected filtered eventCode behavior logs: %s", string(filteredBody))
	}

	postJSON(t, mux, "/api/app/behavior/events", token, `{"eventType":"login","eventCode":"login","businessType":"user","businessId":1}`, http.StatusOK)
	dashboardBody := getAdminJSONWithPermission(t, mux, "/api/admin/dashboard", "analytics:funnel:view", http.StatusOK)
	var dashboardResp struct {
		Data struct {
			Funnel struct {
				Steps []struct {
					EventCode string `json:"eventCode"`
					UserCount int    `json:"userCount"`
				} `json:"steps"`
			} `json:"funnel"`
			Retention struct {
				Buckets []struct {
					CohortDate string `json:"cohortDate"`
					NewUsers   int    `json:"newUsers"`
				} `json:"buckets"`
			} `json:"retention"`
		} `json:"data"`
	}
	if err := json.Unmarshal(dashboardBody, &dashboardResp); err != nil {
		t.Fatal(err)
	}
	if len(dashboardResp.Data.Funnel.Steps) == 0 || len(dashboardResp.Data.Retention.Buckets) == 0 {
		t.Fatalf("expected dashboard funnel and retention: %s", string(dashboardBody))
	}

	funnelBody := getAdminJSONWithPermission(t, mux, "/api/admin/analytics/funnel?eventCodes=login,browse_games,search", "analytics:funnel:view", http.StatusOK)
	var funnelResp struct {
		Data struct {
			Steps []struct {
				EventCode string `json:"eventCode"`
				UserCount int    `json:"userCount"`
			} `json:"steps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(funnelBody, &funnelResp); err != nil {
		t.Fatal(err)
	}
	if len(funnelResp.Data.Steps) != 3 || funnelResp.Data.Steps[0].EventCode != "login" || funnelResp.Data.Steps[1].EventCode != "browse_games" || funnelResp.Data.Steps[1].UserCount == 0 {
		t.Fatalf("expected custom funnel steps: %s", string(funnelBody))
	}

	retentionBody := getAdminJSONWithPermission(t, mux, "/api/admin/analytics/retention", "analytics:retention:view", http.StatusOK)
	var retentionResp struct {
		Data struct {
			Buckets []struct {
				NewUsers int `json:"newUsers"`
			} `json:"buckets"`
		} `json:"data"`
	}
	if err := json.Unmarshal(retentionBody, &retentionResp); err != nil {
		t.Fatal(err)
	}
	if len(retentionResp.Data.Buckets) == 0 || retentionResp.Data.Buckets[0].NewUsers == 0 {
		t.Fatalf("expected retention buckets: %s", string(retentionBody))
	}
}

func TestAIDataSnapshotAndIMExportGateHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "ai-creator")
	completeIdentityForTest(t, mux, creatorToken)
	memberToken := loginForTestWithCode(t, mux, "ai-member")
	completeIdentityForTest(t, mux, memberToken)
	guideToken := loginForTestWithCode(t, mux, "ai-guide")
	completeIdentityForTest(t, mux, guideToken)

	postJSON(t, mux, "/api/app/behavior/events", memberToken, `{"eventType":"search","targetType":"game","targetId":1,"keyword":"ai"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"ai data game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/favorite", memberToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", memberToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "aidata-export", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/chat/messages", creatorToken, `{"messageType":"text","content":"hello member"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", memberToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	postJSON(t, mux, "/api/app/reviews", creatorToken, `{"gameId":1,"targetUserId":2,"targetRole":"member","score":5,"content":"great"}`, http.StatusOK)

	server.connections.UpsertPair(1, 2, "co_game", "co_game", 1, 1)
	server.profiles.GrantRole(1, "expert")
	putJSON(t, mux, "/api/app/experts/me/skills", creatorToken, `{"skillTree":["boardgame"],"serviceTags":["host"]}`, http.StatusOK)
	server.profiles.GrantRole(3, "guide")
	putJSON(t, mux, "/api/app/guides/me/resources", guideToken, `{"resourceTags":["venue"],"industryTags":["entertainment"],"cityCodes":["110100"],"connectionScale":"100-500"}`, http.StatusOK)

	snapshotBody := getAdminJSONWithPermission(t, mux, "/api/admin/ai-data/snapshot", "ai:data:read", http.StatusOK)
	var snapshotResp struct {
		Data struct {
			UserCount          int  `json:"userCount"`
			GameCount          int  `json:"gameCount"`
			BehaviorLogCount   int  `json:"behaviorLogCount"`
			FavoriteCount      int  `json:"favoriteCount"`
			ReviewCount        int  `json:"reviewCount"`
			FootprintCount     int  `json:"footprintCount"`
			ConnectionCount    int  `json:"connectionCount"`
			ExpertProfileCount int  `json:"expertProfileCount"`
			GuideProfileCount  int  `json:"guideProfileCount"`
			IMMessageCount     int  `json:"imMessageCount"`
			DataReady          bool `json:"dataReady"`
			AcceptanceReady    bool `json:"acceptanceReady"`
			AcceptanceChecks   struct {
				Users        struct{ Current, Required int } `json:"users"`
				Games        struct{ Current, Required int } `json:"games"`
				BehaviorLogs struct{ Current, Required int } `json:"behaviorLogs"`
				Favorites    struct{ Current, Required int } `json:"favorites"`
				Reviews      struct{ Current, Required int } `json:"reviews"`
			} `json:"acceptanceChecks"`
			BehaviorStats struct {
				ByDate      map[string]int `json:"byDate"`
				ByUser      map[string]int `json:"byUser"`
				ByEventType map[string]int `json:"byEventType"`
			} `json:"behaviorStats"`
			FavoriteStats struct {
				ByUser     map[string]int `json:"byUser"`
				ByGameType map[string]int `json:"byGameType"`
			} `json:"favoriteStats"`
			ProfileStats struct {
				ExpertCompletenessAverage float64 `json:"expertCompletenessAverage"`
				GuideCompletenessAverage  float64 `json:"guideCompletenessAverage"`
			} `json:"profileStats"`
			ConnectionStats struct {
				BySourceType map[string]int `json:"bySourceType"`
			} `json:"connectionStats"`
			DataReadinessSections map[string]struct {
				Count int  `json:"count"`
				Ready bool `json:"ready"`
			} `json:"dataReadinessSections"`
			IMExportEnabled bool `json:"imExportEnabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(snapshotBody, &snapshotResp); err != nil {
		t.Fatal(err)
	}
	if snapshotResp.Data.BehaviorLogCount == 0 || snapshotResp.Data.FavoriteCount == 0 || snapshotResp.Data.ReviewCount == 0 || snapshotResp.Data.FootprintCount == 0 || snapshotResp.Data.ConnectionCount == 0 || snapshotResp.Data.ExpertProfileCount == 0 || snapshotResp.Data.GuideProfileCount == 0 || snapshotResp.Data.IMMessageCount == 0 {
		t.Fatalf("expected populated ai data snapshot: %s", string(snapshotBody))
	}
	if !snapshotResp.Data.DataReady || snapshotResp.Data.IMExportEnabled {
		t.Fatalf("expected data ready with im export disabled: %s", string(snapshotBody))
	}
	if snapshotResp.Data.AcceptanceReady {
		t.Fatalf("expected acceptance readiness to stay false for single-game fixture: %s", string(snapshotBody))
	}
	if snapshotResp.Data.AcceptanceChecks.Users.Required != 5 || snapshotResp.Data.AcceptanceChecks.Games.Required != 3 || snapshotResp.Data.AcceptanceChecks.BehaviorLogs.Required != 20 || snapshotResp.Data.AcceptanceChecks.Favorites.Required != 5 || snapshotResp.Data.AcceptanceChecks.Reviews.Required != 5 {
		t.Fatalf("expected E5 acceptance thresholds in snapshot: %s", string(snapshotBody))
	}
	if snapshotResp.Data.UserCount == 0 || snapshotResp.Data.GameCount == 0 || snapshotResp.Data.AcceptanceChecks.Users.Current != snapshotResp.Data.UserCount || snapshotResp.Data.AcceptanceChecks.Games.Current != snapshotResp.Data.GameCount {
		t.Fatalf("expected user and game counts in readiness checks: %s", string(snapshotBody))
	}
	if len(snapshotResp.Data.BehaviorStats.ByDate) == 0 || snapshotResp.Data.BehaviorStats.ByUser["2"] == 0 || snapshotResp.Data.BehaviorStats.ByEventType["search"] == 0 {
		t.Fatalf("expected behavior stats by date, user and event type: %s", string(snapshotBody))
	}
	if snapshotResp.Data.FavoriteStats.ByUser["2"] == 0 || snapshotResp.Data.FavoriteStats.ByGameType["free"] == 0 {
		t.Fatalf("expected favorite stats by user and game type: %s", string(snapshotBody))
	}
	if snapshotResp.Data.ProfileStats.ExpertCompletenessAverage == 0 || snapshotResp.Data.ProfileStats.GuideCompletenessAverage == 0 {
		t.Fatalf("expected expert and guide profile completeness stats: %s", string(snapshotBody))
	}
	if snapshotResp.Data.ConnectionStats.BySourceType["co_game"] == 0 || !snapshotResp.Data.DataReadinessSections["connections"].Ready || !snapshotResp.Data.DataReadinessSections["expertProfiles"].Ready {
		t.Fatalf("expected connection and readiness section stats: %s", string(snapshotBody))
	}

	postAdminJSONWithPermission(t, mux, "/api/admin/ai-data/im-export", "ai:data:export", `{}`, http.StatusForbidden)

	configBody := getAdminJSONWithPermission(t, mux, "/api/admin/ai-data/im-export-config", "ai:data:read", http.StatusOK)
	var configResp struct {
		Data struct {
			Config struct {
				Enabled bool `json:"enabled"`
			} `json:"config"`
		} `json:"data"`
	}
	if err := json.Unmarshal(configBody, &configResp); err != nil {
		t.Fatal(err)
	}
	if configResp.Data.Config.Enabled {
		t.Fatalf("expected im export disabled by default: %s", string(configBody))
	}

	putJSON(t, mux, "/api/admin/ai-data/im-export-config", adminLoginForTestAs(t, mux, "operator", "admin123"), `{"enabled":true}`, http.StatusForbidden)
	putJSON(t, mux, "/api/admin/ai-data/im-export-config", adminLoginForTestAs(t, mux, "admin", "admin123"), `{"enabled":true}`, http.StatusOK)

	enabledSnapshotBody := getAdminJSONWithPermission(t, mux, "/api/admin/ai-data/snapshot", "ai:data:read", http.StatusOK)
	var enabledSnapshotResp struct {
		Data struct {
			IMExportEnabled bool `json:"imExportEnabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(enabledSnapshotBody, &enabledSnapshotResp); err != nil {
		t.Fatal(err)
	}
	if !enabledSnapshotResp.Data.IMExportEnabled {
		t.Fatalf("expected im export enabled after config update: %s", string(enabledSnapshotBody))
	}

	exportBody := postAdminJSONWithPermission(t, mux, "/api/admin/ai-data/im-export", "ai:data:export", `{}`, http.StatusOK)
	var exportResp struct {
		Data struct {
			Items []struct {
				MessageID int64 `json:"messageId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(exportBody, &exportResp); err != nil {
		t.Fatal(err)
	}
	if len(exportResp.Data.Items) == 0 {
		t.Fatalf("expected im export items after enabling: %s", string(exportBody))
	}

	putJSON(t, mux, "/api/admin/ai-data/im-export-config", adminLoginForTestAs(t, mux, "admin", "admin123"), `{"enabled":false}`, http.StatusOK)
	postAdminJSONWithPermission(t, mux, "/api/admin/ai-data/im-export", "ai:data:export", `{}`, http.StatusForbidden)
}

func TestAIDataAcceptanceFixtureHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	postJSON(t, mux, "/api/admin/ai-data/acceptance-fixture", "", `{}`, http.StatusForbidden)
	body := postAdminJSONWithPermission(t, mux, "/api/admin/ai-data/acceptance-fixture", "ai:data:seed", `{}`, http.StatusOK)
	var resp struct {
		Data struct {
			UsersVerified      int `json:"usersVerified"`
			GamesCreated       int `json:"gamesCreated"`
			BehaviorLogsAdded  int `json:"behaviorLogsAdded"`
			FavoritesAdded     int `json:"favoritesAdded"`
			ReviewsAdded       int `json:"reviewsAdded"`
			AcceptanceSnapshot struct {
				UserCount        int  `json:"userCount"`
				GameCount        int  `json:"gameCount"`
				BehaviorLogCount int  `json:"behaviorLogCount"`
				FavoriteCount    int  `json:"favoriteCount"`
				ReviewCount      int  `json:"reviewCount"`
				AcceptanceReady  bool `json:"acceptanceReady"`
			} `json:"acceptanceSnapshot"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.UsersVerified != 5 || resp.Data.GamesCreated != 3 || resp.Data.BehaviorLogsAdded != 20 || resp.Data.FavoritesAdded != 5 || resp.Data.ReviewsAdded != 5 {
		t.Fatalf("expected E5 acceptance fixture to seed exact gaps: %s", string(body))
	}
	snapshot := resp.Data.AcceptanceSnapshot
	if !snapshot.AcceptanceReady || snapshot.UserCount < 5 || snapshot.GameCount < 3 || snapshot.BehaviorLogCount < 20 || snapshot.FavoriteCount < 5 || snapshot.ReviewCount < 5 {
		t.Fatalf("expected E5 acceptance snapshot ready: %s", string(body))
	}

	repeatBody := postAdminJSONWithPermission(t, mux, "/api/admin/ai-data/acceptance-fixture", "ai:data:seed", `{}`, http.StatusOK)
	var repeatResp struct {
		Data struct {
			GamesCreated       int `json:"gamesCreated"`
			BehaviorLogsAdded  int `json:"behaviorLogsAdded"`
			FavoritesAdded     int `json:"favoritesAdded"`
			ReviewsAdded       int `json:"reviewsAdded"`
			AcceptanceSnapshot struct {
				AcceptanceReady bool `json:"acceptanceReady"`
			} `json:"acceptanceSnapshot"`
		} `json:"data"`
	}
	if err := json.Unmarshal(repeatBody, &repeatResp); err != nil {
		t.Fatal(err)
	}
	if !repeatResp.Data.AcceptanceSnapshot.AcceptanceReady || repeatResp.Data.GamesCreated != 0 || repeatResp.Data.BehaviorLogsAdded != 0 || repeatResp.Data.FavoritesAdded != 0 || repeatResp.Data.ReviewsAdded != 0 {
		t.Fatalf("expected repeat fixture call to be idempotent after ready: %s", string(repeatBody))
	}
}

func TestReportFreezesRevenueAndCanBeHandled(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "report-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "report-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"report game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	extraTokens := approveExtraMembersForHTTP(t, mux, creatorToken, 1, "report-freeze", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/chat/messages", creatorToken, `{"messageType":"text","content":"service started"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/chat/messages", playerToken, `{"messageType":"text","content":"need help"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm", creatorToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/service-confirm-items", playerToken, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	for _, token := range extraTokens {
		postJSON(t, mux, "/api/app/games/1/service-confirm-items", token, `{"confirmItemKeys":["completed","qualified","communicated"]}`, http.StatusOK)
	}
	completeGameReviewsForHTTP(t, mux, 1, append([]string{creatorToken, playerToken}, extraTokens...))
	adminToken := adminLoginForTest(t, mux)
	postAdminJSON(t, mux, "/api/admin/revenue/templates", adminToken, `{"name":"default","gameType":"free","platformBps":1000,"creatorBps":3000,"memberBps":6000}`, http.StatusOK)
	postAdminJSON(t, mux, "/api/admin/revenue/records/generate", adminToken, `{"gameId":1,"amountCent":10000,"templateId":1}`, http.StatusOK)
	postJSON(t, mux, "/api/app/files/upload-token", playerToken, `{"bizType":"chat_file","objectId":1,"fileName":"report-evidence.png","mimeType":"image/png","size":128}`, http.StatusOK)
	postJSON(t, mux, "/api/app/files/upload-token", playerToken, `{"bizType":"report_attachment","objectId":1,"fileName":"report-private.pdf","mimeType":"application/pdf","size":128}`, http.StatusOK)
	getJSON(t, mux, "/api/app/files/2/download-url", playerToken, http.StatusForbidden)

	postJSON(t, mux, "/api/app/reports", playerToken, `{"gameId":1,"targetUserId":1,"reportType":"service_dispute","content":"bad evidence","chatMessageId":999}`, http.StatusUnprocessableEntity)
	reportBody := postJSON(t, mux, "/api/app/reports", playerToken, `{"gameId":1,"targetUserId":1,"reportType":"service_dispute","content":"dispute","chatMessageId":2,"fileId":1,"reviewId":1,"revenueRecordId":1}`, http.StatusOK)
	var reportResp struct {
		Data struct {
			ID            int64 `json:"id"`
			RevenueFrozen bool  `json:"revenueFrozen"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reportBody, &reportResp); err != nil {
		t.Fatal(err)
	}
	if reportResp.Data.ID == 0 || !reportResp.Data.RevenueFrozen {
		t.Fatalf("expected report with frozen revenue: %s", string(reportBody))
	}
	appealFileBody := postJSON(t, mux, "/api/app/files/upload-token", creatorToken, `{"bizType":"report_attachment","objectId":1,"fileName":"appeal-proof.pdf","mimeType":"application/pdf","size":128}`, http.StatusOK)
	var appealFileResp struct {
		Data struct {
			File struct {
				ID int64 `json:"fileId"`
			} `json:"file"`
		} `json:"data"`
	}
	if err := json.Unmarshal(appealFileBody, &appealFileResp); err != nil {
		t.Fatal(err)
	}
	appealBody := postJSON(t, mux, "/api/app/reports/"+strconv.FormatInt(reportResp.Data.ID, 10)+"/appeal", creatorToken, `{"content":"appeal with files","fileIds":[`+strconv.FormatInt(appealFileResp.Data.File.ID, 10)+`]}`, http.StatusOK)
	var appealResp struct {
		Data struct {
			Status        string  `json:"status"`
			HandleResult  string  `json:"handleResult"`
			AppealFileIDs []int64 `json:"appealFileIds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(appealBody, &appealResp); err != nil {
		t.Fatal(err)
	}
	if appealResp.Data.Status != "appealed" || appealResp.Data.HandleResult != "appeal with files" || len(appealResp.Data.AppealFileIDs) != 1 {
		t.Fatalf("expected structured report appeal: %s", string(appealBody))
	}
	postAdminJSON(t, mux, "/api/admin/revenue/records/1/settle", adminToken, `{"method":"offline","proofNo":"P001"}`, http.StatusConflict)

	noticesReq := httptest.NewRequest(http.MethodGet, "/api/app/notifications", nil)
	noticesReq.Header.Set("Authorization", "Bearer "+playerToken)
	noticesRec := httptest.NewRecorder()
	mux.ServeHTTP(noticesRec, noticesReq)
	if noticesRec.Code != http.StatusOK {
		t.Fatalf("expected notifications 200, got %d: %s", noticesRec.Code, noticesRec.Body.String())
	}
	var noticesResp struct {
		Data struct {
			Items []struct {
				ID               int64  `json:"id"`
				NotifyType       string `json:"notifyType"`
				Status           string `json:"status"`
				WechatState      string `json:"wechatState"`
				WechatTemplateID string `json:"wechatTemplateId"`
				WechatTaskID     int64  `json:"wechatTaskId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesRec.Body.Bytes(), &noticesResp); err != nil {
		t.Fatal(err)
	}
	reportNotificationID := int64(0)
	var reportNotification struct {
		ID               int64  `json:"id"`
		NotifyType       string `json:"notifyType"`
		Status           string `json:"status"`
		WechatState      string `json:"wechatState"`
		WechatTemplateID string `json:"wechatTemplateId"`
		WechatTaskID     int64  `json:"wechatTaskId"`
	}
	for _, item := range noticesResp.Data.Items {
		if item.NotifyType == "report_created" && item.Status == "unread" {
			reportNotificationID = item.ID
			reportNotification = item
			break
		}
	}
	if reportNotificationID == 0 {
		t.Fatalf("expected report_created notification: %s", noticesRec.Body.String())
	}
	if reportNotification.WechatState != "pending" || reportNotification.WechatTemplateID == "" || reportNotification.WechatTaskID == 0 {
		t.Fatalf("expected pending wechat subscribe task on notification: %s", noticesRec.Body.String())
	}
	postJSON(t, mux, "/api/app/notifications/"+strconv.FormatInt(reportNotificationID, 10)+"/read", playerToken, `{}`, http.StatusOK)

	templateReq := httptest.NewRequest(http.MethodGet, "/api/admin/notifications/wechat-templates", nil)
	templateReq.Header.Set("Authorization", "Bearer "+adminToken)
	templateRec := httptest.NewRecorder()
	mux.ServeHTTP(templateRec, templateReq)
	if templateRec.Code != http.StatusOK {
		t.Fatalf("expected wechat templates 200, got %d: %s", templateRec.Code, templateRec.Body.String())
	}
	taskReq := httptest.NewRequest(http.MethodGet, "/api/admin/notifications/wechat-tasks", nil)
	taskReq.Header.Set("Authorization", "Bearer "+adminToken)
	taskRec := httptest.NewRecorder()
	mux.ServeHTTP(taskRec, taskReq)
	if taskRec.Code != http.StatusOK {
		t.Fatalf("expected wechat tasks 200, got %d: %s", taskRec.Code, taskRec.Body.String())
	}
	var taskResp struct {
		Data struct {
			Items []struct {
				ID             int64  `json:"id"`
				NotificationID int64  `json:"notificationId"`
				Status         string `json:"status"`
				TemplateID     string `json:"templateId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(taskRec.Body.Bytes(), &taskResp); err != nil {
		t.Fatal(err)
	}
	reportTaskFound := false
	for _, item := range taskResp.Data.Items {
		if item.ID == reportNotification.WechatTaskID && item.NotificationID == reportNotificationID {
			reportTaskFound = item.Status == "pending" && item.TemplateID == reportNotification.WechatTemplateID
			break
		}
	}
	if !reportTaskFound {
		t.Fatalf("expected pending wechat task: %s", taskRec.Body.String())
	}
	postJSON(t, mux, "/api/internal/notifications/wechat-tasks/1/mark-sent", "", `{"resultCode":"0","resultMessage":"mock sent"}`, http.StatusForbidden)
	postJSON(t, mux, "/api/internal/notifications/wechat-tasks/1/mark-sent", adminToken, `{"resultCode":"0","resultMessage":"mock sent"}`, http.StatusOK)

	detailReq := httptest.NewRequest(http.MethodGet, "/api/admin/reports/1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+adminToken)
	detailReq.Header.Set("X-Admin-ID", "99")
	detailRec := httptest.NewRecorder()
	mux.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected report detail 200, got %d: %s", detailRec.Code, detailRec.Body.String())
	}
	var detailResp struct {
		Data struct {
			Report struct {
				ID int64 `json:"id"`
			} `json:"report"`
			Evidence struct {
				Game struct {
					ID int64 `json:"id"`
				} `json:"game"`
				ChatMessages []struct {
					ID      int64  `json:"id"`
					Content string `json:"content"`
				} `json:"chatMessages"`
				Review struct {
					ID      int64  `json:"id"`
					Content string `json:"content"`
				} `json:"review"`
				RevenueRecord struct {
					ID     int64  `json:"id"`
					Status string `json:"status"`
				} `json:"revenueRecord"`
				AppealFileIDs []int64 `json:"appealFileIds"`
				AppealFiles   []struct {
					ID int64 `json:"fileId"`
				} `json:"appealFiles"`
			} `json:"evidence"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailResp); err != nil {
		t.Fatal(err)
	}
	if detailResp.Data.Report.ID != 1 || detailResp.Data.Evidence.Game.ID != 1 || len(detailResp.Data.Evidence.ChatMessages) == 0 || detailResp.Data.Evidence.Review.ID != 1 || detailResp.Data.Evidence.RevenueRecord.Status != "frozen" || len(detailResp.Data.Evidence.AppealFileIDs) != 1 || len(detailResp.Data.Evidence.AppealFiles) != 1 {
		t.Fatalf("expected report evidence detail: %s", detailRec.Body.String())
	}
	reportChatContents := map[string]bool{}
	for _, message := range detailResp.Data.Evidence.ChatMessages {
		reportChatContents[message.Content] = true
	}
	if len(detailResp.Data.Evidence.ChatMessages) < 2 || !reportChatContents["service started"] || !reportChatContents["need help"] {
		t.Fatalf("expected complete report chat evidence messages: %s", detailRec.Body.String())
	}

	noReportHandleBody := postJSON(t, mux, "/api/admin/reports/1/handle", "", `{"adminId":99,"result":"no permission"}`, http.StatusForbidden)
	if len(noReportHandleBody) == 0 {
		t.Fatal("expected forbidden report handle body")
	}

	noPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/im/rooms/1/dispute-messages", nil)
	noPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noPermRec, noPermReq)
	if noPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected dispute messages 403 without permission, got %d: %s", noPermRec.Code, noPermRec.Body.String())
	}

	disputeReq := httptest.NewRequest(http.MethodGet, "/api/admin/im/rooms/1/dispute-messages", nil)
	disputeReq.Header.Set("X-Admin-ID", "99")
	disputeReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	disputeReq.Header.Set("X-Request-ID", "req-dispute-1")
	disputeRec := httptest.NewRecorder()
	mux.ServeHTTP(disputeRec, disputeReq)
	if disputeRec.Code != http.StatusOK {
		t.Fatalf("expected dispute messages 200, got %d: %s", disputeRec.Code, disputeRec.Body.String())
	}
	var disputeResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(disputeRec.Body.Bytes(), &disputeResp); err != nil {
		t.Fatal(err)
	}
	if len(disputeResp.Data.Items) < 2 {
		t.Fatalf("expected dispute messages, got %s", disputeRec.Body.String())
	}

	myReq := httptest.NewRequest(http.MethodGet, "/api/app/reports/my?page=1&pageSize=1", nil)
	myReq.Header.Set("Authorization", "Bearer "+playerToken)
	myRec := httptest.NewRecorder()
	mux.ServeHTTP(myRec, myReq)
	if myRec.Code != http.StatusOK {
		t.Fatalf("expected my reports 200, got %d: %s", myRec.Code, myRec.Body.String())
	}
	var myResp struct {
		Data struct {
			Items    []struct{ ID int64 } `json:"items"`
			Page     int                  `json:"page"`
			PageSize int                  `json:"pageSize"`
			Total    int                  `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(myRec.Body.Bytes(), &myResp); err != nil {
		t.Fatal(err)
	}
	if myResp.Data.Page != 1 || myResp.Data.PageSize != 1 || myResp.Data.Total < 1 || len(myResp.Data.Items) != 1 {
		t.Fatalf("expected paged my reports: %s", myRec.Body.String())
	}

	adminReq := httptest.NewRequest(http.MethodGet, "/api/admin/reports?page=1&pageSize=1", nil)
	adminReq.Header.Set("Authorization", "Bearer "+adminToken)
	adminRec := httptest.NewRecorder()
	mux.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("expected admin reports 200, got %d: %s", adminRec.Code, adminRec.Body.String())
	}
	if !bytes.Contains(adminRec.Body.Bytes(), []byte(`"pageSize":1`)) || !bytes.Contains(adminRec.Body.Bytes(), []byte(`"total":`)) {
		t.Fatalf("expected paged admin reports: %s", adminRec.Body.String())
	}
	assignBody := postAdminJSON(t, mux, "/api/admin/reports/1/assign", adminToken, `{"adminId":98,"handlerAdminId":99}`, http.StatusOK)
	var assignResp struct {
		Data struct {
			Status         string `json:"status"`
			HandlerAdminID int64  `json:"handlerAdminId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(assignBody, &assignResp); err != nil {
		t.Fatal(err)
	}
	if assignResp.Data.Status != "assigned" || assignResp.Data.HandlerAdminID != 99 {
		t.Fatalf("expected assigned report: %s", string(assignBody))
	}
	handleBody := postAdminJSON(t, mux, "/api/admin/reports/1/handle", adminToken, `{"adminId":99,"result":"confirmed and rewarded","outcome":"confirmed","rewardPoints":20,"creditDeduct":10}`, http.StatusOK)
	var handleResp struct {
		Data struct {
			Status             string `json:"status"`
			HandleOutcome      string `json:"handleOutcome"`
			RewardPoints       int    `json:"rewardPoints"`
			CreditChange       int    `json:"creditChange"`
			CreditTargetUserID int64  `json:"creditTargetUserId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(handleBody, &handleResp); err != nil {
		t.Fatal(err)
	}
	if handleResp.Data.Status != "handled" || handleResp.Data.HandleOutcome != "confirmed" || handleResp.Data.RewardPoints != 20 || handleResp.Data.CreditChange != -10 || handleResp.Data.CreditTargetUserID != 1 {
		t.Fatalf("expected handled report: %s", string(handleBody))
	}

	reporterPointsBody := getJSON(t, mux, "/api/app/points/summary", playerToken, http.StatusOK)
	var reporterPointsResp struct {
		Data struct {
			AvailablePoints int `json:"availablePoints"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reporterPointsBody, &reporterPointsResp); err != nil {
		t.Fatal(err)
	}
	if reporterPointsResp.Data.AvailablePoints < 20 {
		t.Fatalf("expected report reward points: %s", string(reporterPointsBody))
	}

	targetCreditBody := getJSON(t, mux, "/api/app/profile/credit-center", creatorToken, http.StatusOK)
	var targetCreditResp struct {
		Data struct {
			Score   int `json:"score"`
			Records []struct {
				Reason string `json:"reason"`
				Score  string `json:"score"`
			} `json:"records"`
		} `json:"data"`
	}
	if err := json.Unmarshal(targetCreditBody, &targetCreditResp); err != nil {
		t.Fatal(err)
	}
	if targetCreditResp.Data.Score != 90 || len(targetCreditResp.Data.Records) == 0 || targetCreditResp.Data.Records[0].Reason != "report_confirmed" {
		t.Fatalf("expected target credit deduction from report handling: %s", string(targetCreditBody))
	}

	noticesAfterHandleReq := httptest.NewRequest(http.MethodGet, "/api/app/notifications", nil)
	noticesAfterHandleReq.Header.Set("Authorization", "Bearer "+playerToken)
	noticesAfterHandleRec := httptest.NewRecorder()
	mux.ServeHTTP(noticesAfterHandleRec, noticesAfterHandleReq)
	if noticesAfterHandleRec.Code != http.StatusOK {
		t.Fatalf("expected notifications after handle 200, got %d: %s", noticesAfterHandleRec.Code, noticesAfterHandleRec.Body.String())
	}
	var noticesAfterHandleResp struct {
		Data struct {
			Items []struct {
				ID               int64  `json:"id"`
				NotifyType       string `json:"notifyType"`
				Status           string `json:"status"`
				WechatState      string `json:"wechatState"`
				WechatTemplateID string `json:"wechatTemplateId"`
				WechatTaskID     int64  `json:"wechatTaskId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesAfterHandleRec.Body.Bytes(), &noticesAfterHandleResp); err != nil {
		t.Fatal(err)
	}
	if !hasNotification(noticesAfterHandleResp.Data.Items, "report_handled", "unread") {
		t.Fatalf("expected report_handled notification: %s", noticesAfterHandleRec.Body.String())
	}
	if !hasNotification(noticesAfterHandleResp.Data.Items, "report_assigned", "unread") {
		t.Fatalf("expected report_assigned notification: %s", noticesAfterHandleRec.Body.String())
	}
	postAdminJSON(t, mux, "/api/admin/reports/1/close", adminToken, `{"adminId":99,"result":"closed"}`, http.StatusOK)

	behaviorReq := httptest.NewRequest(http.MethodGet, "/api/admin/behavior-logs", nil)
	behaviorReq.Header.Set("Authorization", "Bearer "+adminToken)
	behaviorRec := httptest.NewRecorder()
	mux.ServeHTTP(behaviorRec, behaviorReq)
	if behaviorRec.Code != http.StatusOK {
		t.Fatalf("expected behavior logs 200, got %d: %s", behaviorRec.Code, behaviorRec.Body.String())
	}
	var behaviorResp struct {
		Data struct {
			Items []struct {
				EventType string `json:"eventType"`
				TargetID  int64  `json:"targetId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(behaviorRec.Body.Bytes(), &behaviorResp); err != nil {
		t.Fatal(err)
	}
	if !hasBehaviorEvent(behaviorResp.Data.Items, "submit_report", 1) {
		t.Fatalf("expected submit_report behavior log: %s", behaviorRec.Body.String())
	}

	noOperationPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/operation-logs", nil)
	noOperationPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noOperationPermRec, noOperationPermReq)
	if noOperationPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected operation logs 403 without permission, got %d: %s", noOperationPermRec.Code, noOperationPermRec.Body.String())
	}

	operationReq := httptest.NewRequest(http.MethodGet, "/api/admin/operation-logs", nil)
	operationReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	operationRec := httptest.NewRecorder()
	mux.ServeHTTP(operationRec, operationReq)
	if operationRec.Code != http.StatusOK {
		t.Fatalf("expected operation logs 200, got %d: %s", operationRec.Code, operationRec.Body.String())
	}
	var operationResp struct {
		Data struct {
			Items []struct {
				Action      string `json:"action"`
				TargetType  string `json:"targetType"`
				TargetID    string `json:"targetId"`
				AdminUserID int64  `json:"adminUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(operationRec.Body.Bytes(), &operationResp); err != nil {
		t.Fatal(err)
	}
	if !hasOperationLog(operationResp.Data.Items, "im:message:view_dispute", "im_room", "1", 99) {
		t.Fatalf("expected dispute view operation log: %s", operationRec.Body.String())
	}
	if !hasOperationLog(operationResp.Data.Items, "report:handle", "report", "1", 0) {
		t.Fatalf("expected report handle operation log: %s", operationRec.Body.String())
	}
	if !hasOperationLog(operationResp.Data.Items, "report:assign", "report", "1", 0) {
		t.Fatalf("expected report assign operation log: %s", operationRec.Body.String())
	}
	if !hasOperationLog(operationResp.Data.Items, "report:view_detail", "report", "1", 99) {
		t.Fatalf("expected report detail operation log: %s", operationRec.Body.String())
	}

	filteredOperationReq := httptest.NewRequest(http.MethodGet, "/api/admin/operation-logs?action=report&targetType=report&targetId=1&adminUserId=99", nil)
	filteredOperationReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	filteredOperationRec := httptest.NewRecorder()
	mux.ServeHTTP(filteredOperationRec, filteredOperationReq)
	if filteredOperationRec.Code != http.StatusOK {
		t.Fatalf("expected filtered operation logs 200, got %d: %s", filteredOperationRec.Code, filteredOperationRec.Body.String())
	}
	var filteredOperationResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				Action      string `json:"action"`
				TargetType  string `json:"targetType"`
				TargetID    string `json:"targetId"`
				AdminUserID int64  `json:"adminUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(filteredOperationRec.Body.Bytes(), &filteredOperationResp); err != nil {
		t.Fatal(err)
	}
	if filteredOperationResp.Data.Total != 1 || !hasOperationLog(filteredOperationResp.Data.Items, "report:view_detail", "report", "1", 99) {
		t.Fatalf("expected filtered report operation log for admin 99: %s", filteredOperationRec.Body.String())
	}
}

func TestReportAppealCanBeWithdrawn(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)
	reporterToken := loginForTestWithCode(t, mux, "appeal-withdraw-reporter")
	targetToken := loginForTestWithCode(t, mux, "appeal-withdraw-target")

	postJSON(t, mux, "/api/app/games", reporterToken, `{"title":"appeal withdraw game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", reporterToken, `{}`, http.StatusOK)
	applicationBody := postJSON(t, mux, "/api/app/games/1/applications", targetToken, `{"reason":"join"}`, http.StatusOK)
	var applicationResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(applicationBody, &applicationResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/games/applications/"+strconv.FormatInt(applicationResp.Data.ID, 10)+"/review", reporterToken, `{"approve":true}`, http.StatusOK)
	reportBody := postJSON(t, mux, "/api/app/reports", reporterToken, `{"gameId":1,"targetUserId":2,"reportType":"other","content":"need review"}`, http.StatusOK)
	var reportResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reportBody, &reportResp); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/reports/"+strconv.FormatInt(reportResp.Data.ID, 10)+"/appeal", targetToken, `{"content":"appeal reason"}`, http.StatusOK)
	withdrawBody := postJSON(t, mux, "/api/app/reports/"+strconv.FormatInt(reportResp.Data.ID, 10)+"/appeal/withdraw", targetToken, `{}`, http.StatusOK)
	var withdrawResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(withdrawBody, &withdrawResp); err != nil {
		t.Fatal(err)
	}
	if withdrawResp.Data.Status != "appeal_withdrawn" {
		t.Fatalf("expected withdrawn appeal, got: %s", string(withdrawBody))
	}
	postJSON(t, mux, "/api/app/reports/"+strconv.FormatInt(reportResp.Data.ID, 10)+"/appeal/withdraw", targetToken, `{}`, http.StatusUnprocessableEntity)
	appealsBody := getJSON(t, mux, "/api/app/reports/appeals/my", targetToken, http.StatusOK)
	var appealsResp struct {
		Data struct {
			Items []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(appealsBody, &appealsResp); err != nil {
		t.Fatal(err)
	}
	if len(appealsResp.Data.Items) != 1 || appealsResp.Data.Items[0].Status != "appeal_withdrawn" {
		t.Fatalf("expected withdrawn appeal retained in appeal history, got: %s", string(appealsBody))
	}
	homeBody := getJSON(t, mux, "/api/app/profile/home", targetToken, http.StatusOK)
	if strings.Contains(string(homeBody), "1条消息") {
		t.Fatalf("expected withdrawn appeal not counted as profile report todo: %s", string(homeBody))
	}
}

func TestAdminBatchHandleReports(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	userToken := loginForTestWithCode(t, mux, "batch-report-user")
	completeIdentityForTest(t, mux, userToken)
	postJSON(t, mux, "/api/app/games", userToken, `{"title":"batch report game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", userToken, `{}`, http.StatusOK)
	firstReport := postJSON(t, mux, "/api/app/reports", userToken, `{"gameId":1,"reportType":"other","content":"first"}`, http.StatusOK)
	secondReport := postJSON(t, mux, "/api/app/reports", userToken, `{"gameId":1,"reportType":"other","content":"second"}`, http.StatusOK)
	var first, second struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(firstReport, &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(secondReport, &second); err != nil {
		t.Fatal(err)
	}

	analystToken := adminLoginForTestAs(t, mux, "data_analyst", "admin123")
	postAdminJSON(t, mux, "/api/admin/reports/batch-handle", analystToken, `{"reportIds":[1],"action":"handle","adminId":99,"result":"blocked"}`, http.StatusForbidden)

	body := postAdminJSONWithPermission(t, mux, "/api/admin/reports/batch-handle", "report:handle", `{"reportIds":[`+strconv.FormatInt(first.Data.ID, 10)+`,`+strconv.FormatInt(second.Data.ID, 10)+`],"action":"handle","adminId":99,"result":"batch handled"}`, http.StatusOK)
	var resp struct {
		Data struct {
			Success int `json:"success"`
			Failed  int `json:"failed"`
			Items   []struct {
				ID      int64  `json:"id"`
				Success bool   `json:"success"`
				Status  string `json:"status"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Success != 2 || resp.Data.Failed != 0 || len(resp.Data.Items) != 2 || resp.Data.Items[0].Status != "handled" {
		t.Fatalf("expected two handled reports: %s", string(body))
	}

	listBody := getAdminJSONWithPermission(t, mux, "/api/admin/reports", "report:view", http.StatusOK)
	var listResp struct {
		Data struct {
			Items []struct {
				Status         string `json:"status"`
				HandlerAdminID int64  `json:"handlerAdminId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listBody, &listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Data.Items) != 2 || listResp.Data.Items[0].Status != "handled" || listResp.Data.Items[1].Status != "handled" {
		t.Fatalf("expected reports list to be handled: %s", string(listBody))
	}
	if listResp.Data.Items[0].HandlerAdminID != 1 || listResp.Data.Items[1].HandlerAdminID != 1 {
		t.Fatalf("expected batch handler admin id from token, got: %s", string(listBody))
	}
}

func TestReportBatchActionPermissionMatchesAction(t *testing.T) {
	cases := map[string]string{
		"assign": "report:assign",
		"handle": "report:handle",
		"close":  "report:close",
	}
	for action, want := range cases {
		got, ok := reportBatchActionPermission(action)
		if !ok || got != want {
			t.Fatalf("expected %s to require %s, got %s ok=%v", action, want, got, ok)
		}
	}
	if got, ok := reportBatchActionPermission("unknown"); ok || got != "" {
		t.Fatalf("expected unknown action to be rejected, got %s ok=%v", got, ok)
	}
}

func TestProgressFeedbackReminderJob(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "progress-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "progress-player")
	completeIdentityForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"progress game","gameType":"free","minPlayers":5,"maxPlayers":8,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/applications", playerToken, `{"reason":"join"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/applications/1/review", creatorToken, `{"approve":true}`, http.StatusOK)
	approveExtraMembersForHTTP(t, mux, creatorToken, 1, "progress-reminder", 3)
	postJSON(t, mux, "/api/app/games/1/manual-start", creatorToken, `{}`, http.StatusOK)

	body := postJSON(t, mux, "/api/internal/jobs/progress-feedback-remind", "", `{}`, http.StatusOK)
	var resp struct {
		Data struct {
			Created int `json:"created"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Created == 0 {
		t.Fatalf("expected progress remind notification: %s", string(body))
	}

	noticesReq := httptest.NewRequest(http.MethodGet, "/api/app/notifications", nil)
	noticesReq.Header.Set("Authorization", "Bearer "+creatorToken)
	noticesRec := httptest.NewRecorder()
	mux.ServeHTTP(noticesRec, noticesReq)
	if noticesRec.Code != http.StatusOK {
		t.Fatalf("expected notifications 200, got %d: %s", noticesRec.Code, noticesRec.Body.String())
	}
	var noticesResp struct {
		Data struct {
			Items []struct {
				ID               int64  `json:"id"`
				NotifyType       string `json:"notifyType"`
				Status           string `json:"status"`
				WechatState      string `json:"wechatState"`
				WechatTemplateID string `json:"wechatTemplateId"`
				WechatTaskID     int64  `json:"wechatTaskId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(noticesRec.Body.Bytes(), &noticesResp); err != nil {
		t.Fatal(err)
	}
	if !hasNotification(noticesResp.Data.Items, "progress_feedback_remind", "unread") {
		t.Fatalf("expected progress_feedback_remind notification: %s", noticesRec.Body.String())
	}
}

func TestLocationAndNearbyGamesFlow(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{LBS: config.LBSConfig{DefaultRadiusMeter: 1}})
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "loc-user")
	completeIdentityForTest(t, mux, token)
	postJSON(t, mux, "/api/app/locations/current", token, `{"longitude":181,"latitude":39.916527,"accuracyMeter":80,"cityCode":"110100","cityName":"Beijing"}`, http.StatusUnprocessableEntity)
	locationBody := postJSON(t, mux, "/api/app/locations/current", token, `{"longitude":116.397128,"latitude":39.916527,"accuracyMeter":80,"cityCode":"110100","cityName":"Beijing"}`, http.StatusOK)
	var locationResp struct {
		Data struct {
			Longitude       float64 `json:"longitude"`
			Latitude        float64 `json:"latitude"`
			AccuracyWarning bool    `json:"accuracyWarning"`
			Source          string  `json:"source"`
			CityCode        string  `json:"cityCode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(locationBody, &locationResp); err != nil {
		t.Fatal(err)
	}
	if locationResp.Data.Source != "gps" || locationResp.Data.AccuracyWarning || locationResp.Data.CityCode != "110100" {
		t.Fatalf("expected current LocationDTO: %s", string(locationBody))
	}
	manualBody := postJSON(t, mux, "/api/app/locations/manual", token, `{"longitude":116.397128,"latitude":39.916527,"accuracyMeter":500,"cityCode":"110100","cityName":"Beijing","address":"manual point"}`, http.StatusOK)
	if !bytes.Contains(manualBody, []byte(`"source":"manual"`)) || !bytes.Contains(manualBody, []byte(`"accuracyWarning":true`)) {
		t.Fatalf("expected manual LocationDTO with warning: %s", string(manualBody))
	}
	recentReq := httptest.NewRequest(http.MethodGet, "/api/app/locations/my-recent?limit=5", nil)
	recentReq.Header.Set("Authorization", "Bearer "+token)
	recentRec := httptest.NewRecorder()
	mux.ServeHTTP(recentRec, recentReq)
	if recentRec.Code != http.StatusOK {
		t.Fatalf("expected recent locations 200, got %d: %s", recentRec.Code, recentRec.Body.String())
	}
	var recentResp struct {
		Data struct {
			Items []struct {
				Source string `json:"source"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recentRec.Body.Bytes(), &recentResp); err != nil {
		t.Fatal(err)
	}
	if len(recentResp.Data.Items) < 2 || recentResp.Data.Items[0].Source != "manual" {
		t.Fatalf("expected recent locations with manual first: %s", recentRec.Body.String())
	}
	manualRecentReq := httptest.NewRequest(http.MethodGet, "/api/app/locations/my-recent?source=manual", nil)
	manualRecentReq.Header.Set("Authorization", "Bearer "+token)
	manualRecentRec := httptest.NewRecorder()
	mux.ServeHTTP(manualRecentRec, manualRecentReq)
	if manualRecentRec.Code != http.StatusOK {
		t.Fatalf("expected manual recent 200, got %d: %s", manualRecentRec.Code, manualRecentRec.Body.String())
	}
	if err := json.Unmarshal(manualRecentRec.Body.Bytes(), &recentResp); err != nil {
		t.Fatal(err)
	}
	if len(recentResp.Data.Items) != 1 || recentResp.Data.Items[0].Source != "manual" {
		t.Fatalf("expected manual-only recent locations: %s", manualRecentRec.Body.String())
	}
	invalidSourceReq := httptest.NewRequest(http.MethodGet, "/api/app/locations/my-recent?source=unknown", nil)
	invalidSourceReq.Header.Set("Authorization", "Bearer "+token)
	invalidSourceRec := httptest.NewRecorder()
	mux.ServeHTTP(invalidSourceRec, invalidSourceReq)
	if invalidSourceRec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid source 422, got %d: %s", invalidSourceRec.Code, invalidSourceRec.Body.String())
	}
	otherToken := loginForTestWithCode(t, mux, "loc-user-2")
	completeIdentityForTest(t, mux, otherToken)
	otherRecentReq := httptest.NewRequest(http.MethodGet, "/api/app/locations/my-recent", nil)
	otherRecentReq.Header.Set("Authorization", "Bearer "+otherToken)
	otherRecentRec := httptest.NewRecorder()
	mux.ServeHTTP(otherRecentRec, otherRecentReq)
	if otherRecentRec.Code != http.StatusOK {
		t.Fatalf("expected empty recent locations 200, got %d: %s", otherRecentRec.Code, otherRecentRec.Body.String())
	}
	postJSON(t, mux, "/api/app/locations/current", otherToken, `{"longitude":116.39715,"latitude":39.91654,"accuracyMeter":60,"cityCode":"110100","cityName":"Beijing"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games", token, `{"title":"nearby game","gameType":"free","minPlayers":5,"maxPlayers":8,"cityCode":"110100","cityName":"Beijing","longitude":116.3972,"latitude":39.9166,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", token, `{}`, http.StatusOK)

	invalidRadiusReq := httptest.NewRequest(http.MethodGet, "/api/app/games/nearby?radiusMeter=50001", nil)
	invalidRadiusReq.Header.Set("Authorization", "Bearer "+token)
	invalidRadiusRec := httptest.NewRecorder()
	mux.ServeHTTP(invalidRadiusRec, invalidRadiusReq)
	if invalidRadiusRec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid radius 422, got %d: %s", invalidRadiusRec.Code, invalidRadiusRec.Body.String())
	}

	defaultRadiusReq := httptest.NewRequest(http.MethodGet, "/api/app/games/nearby", nil)
	defaultRadiusReq.Header.Set("Authorization", "Bearer "+token)
	defaultRadiusRec := httptest.NewRecorder()
	mux.ServeHTTP(defaultRadiusRec, defaultRadiusReq)
	if defaultRadiusRec.Code != http.StatusOK {
		t.Fatalf("expected default radius request 200, got %d: %s", defaultRadiusRec.Code, defaultRadiusRec.Body.String())
	}
	var defaultRadiusResp struct {
		Data struct {
			Items []struct {
				ID int64 `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(defaultRadiusRec.Body.Bytes(), &defaultRadiusResp); err != nil {
		t.Fatal(err)
	}
	if len(defaultRadiusResp.Data.Items) != 0 {
		t.Fatalf("expected configured 1m default radius to filter nearby game: %s", defaultRadiusRec.Body.String())
	}
	invalidCoordinateReq := httptest.NewRequest(http.MethodGet, "/api/app/games/nearby?longitude=181&latitude=39.9&radiusMeter=1000", nil)
	invalidCoordinateReq.Header.Set("Authorization", "Bearer "+token)
	invalidCoordinateRec := httptest.NewRecorder()
	mux.ServeHTTP(invalidCoordinateRec, invalidCoordinateReq)
	if invalidCoordinateRec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid coordinate 422, got %d: %s", invalidCoordinateRec.Code, invalidCoordinateRec.Body.String())
	}

	noSavedLocationToken := loginForTestWithCode(t, mux, "loc-user-query-only")
	completeIdentityForTest(t, mux, noSavedLocationToken)
	queryLocationReq := httptest.NewRequest(http.MethodGet, "/api/app/games/nearby?longitude=116.39715&latitude=39.91654&radiusMeters=1000", nil)
	queryLocationReq.Header.Set("Authorization", "Bearer "+noSavedLocationToken)
	queryLocationRec := httptest.NewRecorder()
	mux.ServeHTTP(queryLocationRec, queryLocationReq)
	if queryLocationRec.Code != http.StatusOK {
		t.Fatalf("expected query coordinate nearby request 200, got %d: %s", queryLocationRec.Code, queryLocationRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/app/games/nearby?radiusMeter=1000", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var nearbyResp struct {
		Data struct {
			Items []struct {
				ID            int64   `json:"id"`
				DistanceMeter float64 `json:"distanceMeter"`
				DistanceLabel string  `json:"distanceLabel"`
			} `json:"items"`
			OnlinePlayers []struct {
				UserID       int64  `json:"userId"`
				Name         string `json:"name"`
				DistanceText string `json:"distanceText"`
			} `json:"onlinePlayers"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &nearbyResp); err != nil {
		t.Fatal(err)
	}
	if len(nearbyResp.Data.Items) != 1 || nearbyResp.Data.Items[0].DistanceMeter <= 0 || nearbyResp.Data.Items[0].DistanceLabel == "" {
		t.Fatalf("expected nearby distance fields: %s", rec.Body.String())
	}
	if len(nearbyResp.Data.OnlinePlayers) != 1 || nearbyResp.Data.OnlinePlayers[0].UserID == 0 || nearbyResp.Data.OnlinePlayers[0].Name == "" || nearbyResp.Data.OnlinePlayers[0].DistanceText == "" {
		t.Fatalf("expected nearby player marker fields: %s", rec.Body.String())
	}
}

func TestAppHomeAndConnectionNetworkPayloadHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	creatorToken := loginForTestWithCode(t, mux, "home-creator")
	completeIdentityForTest(t, mux, creatorToken)
	playerToken := loginForTestWithCode(t, mux, "home-player")
	completeIdentityForTest(t, mux, playerToken)
	strangerToken := loginForTestWithCode(t, mux, "home-stranger")
	completeIdentityForTest(t, mux, strangerToken)
	creatorID := currentUserIDForTest(t, mux, creatorToken)
	playerID := currentUserIDForTest(t, mux, playerToken)

	postJSON(t, mux, "/api/app/locations/current", creatorToken, `{"longitude":116.397128,"latitude":39.916527,"accuracyMeter":80,"cityCode":"110100","cityName":"Beijing"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games", creatorToken, `{"title":"home dynamic game","gameType":"free","minPlayers":5,"maxPlayers":8,"cityCode":"110100","cityName":"Beijing","longitude":116.3972,"latitude":39.9166,"startAt":"2026-08-01 10:00","endAt":"2026-08-01 12:00"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/games/1/approve-local", creatorToken, `{}`, http.StatusOK)
	server.connections.UpsertPair(creatorID, playerID, "guide_match", "test", 1, 3)

	homeBody := getJSON(t, mux, "/api/app/home", creatorToken, http.StatusOK)
	var homeResp struct {
		Data struct {
			FriendSection struct {
				Route string `json:"route"`
			} `json:"friendSection"`
			NearBySummary struct {
				NearbyGameCount int `json:"nearbyGameCount"`
				RelationCount   int `json:"relationCount"`
			} `json:"nearbySummary"`
			OnlineCard struct {
				Title string   `json:"title"`
				Desc  string   `json:"desc"`
				Tags  []string `json:"tags"`
			} `json:"onlineCard"`
			NearbyGames []struct {
				ID           int64  `json:"id"`
				DistanceText string `json:"distanceText"`
			} `json:"nearbyGames"`
			FriendGames []struct {
				ID int64 `json:"id"`
			} `json:"friendGames"`
			Earth struct {
				Nodes []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"nodes"`
				HeatPoints []struct {
					CityCode string `json:"cityCode"`
				} `json:"heatPoints"`
			} `json:"earth"`
			Visualization struct {
				Network struct {
					Nodes []struct {
						ID string `json:"id"`
					} `json:"nodes"`
					Edges []struct {
						Target string `json:"target"`
					} `json:"edges"`
				} `json:"network"`
			} `json:"visualization"`
		} `json:"data"`
	}
	if err := json.Unmarshal(homeBody, &homeResp); err != nil {
		t.Fatal(err)
	}
	if homeResp.Data.NearBySummary.NearbyGameCount == 0 || homeResp.Data.NearBySummary.RelationCount == 0 {
		t.Fatalf("expected home nearby and relation summary: %s", string(homeBody))
	}
	if homeResp.Data.OnlineCard.Title == "" || homeResp.Data.OnlineCard.Desc == "" || len(homeResp.Data.OnlineCard.Tags) == 0 {
		t.Fatalf("expected role dashboard online card: %s", string(homeBody))
	}
	if len(homeResp.Data.NearbyGames) == 0 || homeResp.Data.NearbyGames[0].ID != 1 || homeResp.Data.NearbyGames[0].DistanceText == "" {
		t.Fatalf("expected dynamic nearby game cards: %s", string(homeBody))
	}
	if len(homeResp.Data.FriendGames) != 0 {
		t.Fatalf("creator's own game must not appear in friend game cards: %s", string(homeBody))
	}
	if homeResp.Data.FriendSection.Route != "pages/game/hall/index" {
		t.Fatalf("expected friend section route: %s", string(homeBody))
	}
	if len(homeResp.Data.Earth.Nodes) < 3 || len(homeResp.Data.Earth.HeatPoints) == 0 {
		t.Fatalf("expected dynamic earth nodes and heat points: %s", string(homeBody))
	}
	if len(homeResp.Data.Visualization.Network.Nodes) < 2 || len(homeResp.Data.Visualization.Network.Edges) == 0 {
		t.Fatalf("expected visualization network nodes and edges: %s", string(homeBody))
	}

	friendHomeBody := getJSON(t, mux, "/api/app/home", playerToken, http.StatusOK)
	var friendHomeResp struct {
		Data struct {
			FriendGames []struct {
				ID int64 `json:"id"`
			} `json:"friendGames"`
		} `json:"data"`
	}
	if err := json.Unmarshal(friendHomeBody, &friendHomeResp); err != nil {
		t.Fatal(err)
	}
	if len(friendHomeResp.Data.FriendGames) != 1 || friendHomeResp.Data.FriendGames[0].ID != 1 {
		t.Fatalf("expected connected user's game in friend game cards: %s", string(friendHomeBody))
	}

	strangerHomeBody := getJSON(t, mux, "/api/app/home", strangerToken, http.StatusOK)
	var strangerHomeResp struct {
		Data struct {
			FriendGames []struct {
				ID int64 `json:"id"`
			} `json:"friendGames"`
		} `json:"data"`
	}
	if err := json.Unmarshal(strangerHomeBody, &strangerHomeResp); err != nil {
		t.Fatal(err)
	}
	if len(strangerHomeResp.Data.FriendGames) != 0 {
		t.Fatalf("unrelated games must not appear in friend game cards: %s", string(strangerHomeBody))
	}

	if err := server.systemConfig.Set(homeDisplayConfigKey, homeDisplayConfigDTO{OnlineBaseCount: 20, OnlineSuffix: "人在线"}); err != nil {
		t.Fatal(err)
	}
	configuredHomeBody := getJSON(t, mux, "/api/app/home", creatorToken, http.StatusOK)
	var configuredHomeResp struct {
		Data struct {
			Hero struct {
				OnlineText string `json:"onlineText"`
			} `json:"hero"`
		} `json:"data"`
	}
	if err := json.Unmarshal(configuredHomeBody, &configuredHomeResp); err != nil {
		t.Fatal(err)
	}
	if configuredHomeResp.Data.Hero.OnlineText != "0人在线" {
		t.Fatalf("expected configured home online text, got %q in %s", configuredHomeResp.Data.Hero.OnlineText, string(configuredHomeBody))
	}
	adminToken := adminLoginForTestAs(t, mux, "admin", "admin123")
	adminConfigBody := putJSON(t, mux, "/api/admin/home/display-config", adminToken, `{"onlineBaseCount":30,"onlineSuffix":"人在线"}`, http.StatusOK)
	var adminConfigResp struct {
		Data struct {
			Config homeDisplayConfigDTO `json:"config"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminConfigBody, &adminConfigResp); err != nil {
		t.Fatal(err)
	}
	if adminConfigResp.Data.Config.OnlineBaseCount != 30 {
		t.Fatalf("expected admin saved home display config: %s", string(adminConfigBody))
	}
	adminConfiguredHomeBody := getJSON(t, mux, "/api/app/home", creatorToken, http.StatusOK)
	var adminConfiguredHomeResp struct {
		Data struct {
			Hero struct {
				OnlineText string `json:"onlineText"`
			} `json:"hero"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminConfiguredHomeBody, &adminConfiguredHomeResp); err != nil {
		t.Fatal(err)
	}
	if adminConfiguredHomeResp.Data.Hero.OnlineText != "0人在线" {
		t.Fatalf("expected admin-configured home online text, got %q in %s", adminConfiguredHomeResp.Data.Hero.OnlineText, string(adminConfiguredHomeBody))
	}

	networkBody := getJSON(t, mux, "/api/app/connections/my", creatorToken, http.StatusOK)
	var networkResp struct {
		Data struct {
			Items []struct {
				ConnectedUserID int64 `json:"connectedUserId"`
			} `json:"items"`
			Header struct {
				Title string `json:"title"`
			} `json:"header"`
			Tabs []struct {
				Key string `json:"key"`
			} `json:"tabs"`
			Network struct {
				Nodes []struct {
					UserID int64  `json:"userId"`
					Type   string `json:"type"`
				} `json:"nodes"`
				Edges []struct {
					Target string `json:"target"`
				} `json:"edges"`
				Stats []struct {
					Key   string `json:"key"`
					Value int    `json:"value"`
				} `json:"stats"`
			} `json:"network"`
		} `json:"data"`
	}
	if err := json.Unmarshal(networkBody, &networkResp); err != nil {
		t.Fatal(err)
	}
	if len(networkResp.Data.Items) != 1 || networkResp.Data.Items[0].ConnectedUserID != playerID {
		t.Fatalf("expected original connection list to remain available: %s", string(networkBody))
	}
	if networkResp.Data.Header.Title == "" || len(networkResp.Data.Tabs) < 2 {
		t.Fatalf("expected network header and tabs: %s", string(networkBody))
	}
	if len(networkResp.Data.Network.Nodes) < 2 || len(networkResp.Data.Network.Edges) == 0 || len(networkResp.Data.Network.Stats) == 0 {
		t.Fatalf("expected network payload nodes edges stats: %s", string(networkBody))
	}
}

func TestHomeExpertSkillStylesComeFromBackend(t *testing.T) {
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	server := newTestAppServer(authService, identity.NewService())
	server.profiles.GrantRole(1, "expert")
	if _, err := server.profiles.UpdateExpertSkill(1, profiles.ExpertSkillRequest{
		SkillTree: []string{"AI工具落地", "创业复盘"},
	}); err != nil {
		t.Fatalf("seed expert skill failed: %v", err)
	}

	items := server.homeExpertSkills(1)
	if len(items) != 3 {
		t.Fatalf("expected three expert skill nodes, got %d", len(items))
	}
	if items[0]["tone"] != "cyan" || items[0]["nodeStyle"] == "" || items[0]["locked"] != false {
		t.Fatalf("expected backend-driven first skill style: %#v", items[0])
	}
	if items[2]["tone"] != "purple" || items[2]["nodeStyle"] == "" || items[2]["locked"] != true {
		t.Fatalf("expected backend-driven locked skill style metadata: %#v", items[2])
	}
}

func TestTencentMapProxyHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	provider := &fakeMapProvider{}
	server.UseMapProvider(provider)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "map-proxy-user")
	completeIdentityForTest(t, mux, token)

	searchBody := getJSON(t, mux, "/api/app/map/search?keyword=%E8%A5%BF%E6%B9%96&city=%E6%9D%AD%E5%B7%9E&longitude=120.15&latitude=30.27&radiusMeter=5000&page=2&pageSize=5", token, http.StatusOK)
	var searchResp struct {
		Data struct {
			Provider string `json:"provider"`
			Total    int    `json:"total"`
			Items    []struct {
				Title         string  `json:"title"`
				DistanceMeter float64 `json:"distanceMeter"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(searchBody, &searchResp); err != nil {
		t.Fatal(err)
	}
	if searchResp.Data.Provider != "tencent" || searchResp.Data.Total != 1 || searchResp.Data.Items[0].Title != "西湖" || provider.searchReq.Page != 2 || provider.searchReq.PageSize != 5 {
		t.Fatalf("unexpected map search response/request: body=%s req=%+v", string(searchBody), provider.searchReq)
	}

	geocodeBody := getJSON(t, mux, "/api/app/map/geocode?address=%E6%9D%AD%E5%B7%9E%E8%A5%BF%E6%B9%96&city=%E6%9D%AD%E5%B7%9E", token, http.StatusOK)
	if !bytes.Contains(geocodeBody, []byte(`"provider":"tencent"`)) || provider.geocodeReq.Address != "杭州西湖" {
		t.Fatalf("unexpected geocode response/request: body=%s req=%+v", string(geocodeBody), provider.geocodeReq)
	}

	reverseBody := getJSON(t, mux, "/api/app/map/reverse-geocode?longitude=120.1551&latitude=30.2741", token, http.StatusOK)
	if !bytes.Contains(reverseBody, []byte(`"address":"杭州市西湖区"`)) || provider.reverseReq.Longitude != 120.1551 {
		t.Fatalf("unexpected reverse geocode response/request: body=%s req=%+v", string(reverseBody), provider.reverseReq)
	}

	routeBody := getJSON(t, mux, "/api/app/map/route?fromLongitude=120.15&fromLatitude=30.27&toLongitude=120.16&toLatitude=30.28&mode=walking", token, http.StatusOK)
	if !bytes.Contains(routeBody, []byte(`"distanceMeter":1500`)) || provider.routeReq.Mode != "walking" {
		t.Fatalf("unexpected route response/request: body=%s req=%+v", string(routeBody), provider.routeReq)
	}

	getJSON(t, mux, "/api/app/map/search?keyword=bad", token, http.StatusUnprocessableEntity)
	getJSON(t, mux, "/api/app/map/reverse-geocode?longitude=bad&latitude=30.2741", token, http.StatusUnprocessableEntity)
}

func TestTencentMapProviderRateLimitHTTP(t *testing.T) {
	server := newTestAppServer(auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()), identity.NewService())
	provider, err := lbs.NewTencentMapClient(lbs.TencentMapConfig{
		KeyServer: "server-map-key",
		APIBase:   "https://apis.map.qq.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	server.UseMapProvider(provider)

	req := httptest.NewRequest(http.MethodGet, "/api/app/map/search?keyword=test", nil)
	first := httptest.NewRecorder()
	if !server.ensureMapProviderRequestAllowed(first, req) {
		t.Fatalf("expected first map provider request to pass, got %d: %s", first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	if server.ensureMapProviderRequestAllowed(second, req) {
		t.Fatal("expected second map provider request to be rate limited")
	}
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d: %s", second.Code, second.Body.String())
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestMapMyCityConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.UseSystemConfigRepository(&appTestSystemConfigRepository{values: map[string]json.RawMessage{}})
	config := defaultMapMyCityConfig()
	config.ProfileName = "测试城市档案"
	config.JourneyDesc = "通过 2 个局，认识了 6 位朋友"
	if err := server.systemConfig.Set(mapMyCityConfigKey, config); err != nil {
		t.Fatal(err)
	}
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "map-my-city-user")
	completeIdentityForTest(t, mux, token)

	body := getJSON(t, mux, "/api/app/map/my-city", token, http.StatusOK)
	var resp struct {
		Data mapMyCityConfigDTO `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ProfileName != "测试城市档案" || resp.Data.JourneyDesc != "通过 2 个局，认识了 6 位朋友" || len(resp.Data.StoryGroups) == 0 || len(resp.Data.TagEmojis) == 0 {
		t.Fatalf("unexpected my city config: %s", string(body))
	}
}

func TestMapIndexConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.UseSystemConfigRepository(&appTestSystemConfigRepository{values: map[string]json.RawMessage{}})
	config := defaultMapIndexConfig()
	config.DefaultRadiusMeters = 8000
	config.MapFilters = []string{"数据库附近局", "数据库热力"}
	config.Texts["nearbySectionTitle"] = "数据库玩法点"
	if err := server.systemConfig.Set(mapIndexConfigKey, config); err != nil {
		t.Fatal(err)
	}
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "map-index-config-user")
	completeIdentityForTest(t, mux, token)

	body := getJSON(t, mux, "/api/app/map/index-config", token, http.StatusOK)
	var resp struct {
		Data mapIndexConfigDTO `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.DefaultRadiusMeters != 8000 || len(resp.Data.MapFilters) != 2 || resp.Data.MapFilters[0] != "数据库附近局" || resp.Data.Texts["nearbySectionTitle"] != "数据库玩法点" {
		t.Fatalf("unexpected map index config: %s", string(body))
	}
}

func TestMapPlayPagesConfigFeedsAppHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.UseSystemConfigRepository(&appTestSystemConfigRepository{values: map[string]json.RawMessage{}})
	config := defaultMapPlayPagesConfig()
	pages := config["pages"].(map[string]interface{})
	page := pages["blind-route"].(map[string]interface{})
	page["title"] = "测试盲盒"
	if err := server.systemConfig.Set(mapPlayPagesConfigKey, config); err != nil {
		t.Fatal(err)
	}
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "map-play-pages-user")
	completeIdentityForTest(t, mux, token)

	body := getJSON(t, mux, "/api/app/map/play-pages?pageKey=blind-route", token, http.StatusOK)
	var resp struct {
		Data struct {
			Title        string                   `json:"title"`
			Description  string                   `json:"description"`
			SectionTitle string                   `json:"sectionTitle"`
			Cards        []map[string]interface{} `json:"cards"`
			RecentRoutes []map[string]interface{} `json:"recentRoutes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Title != "测试盲盒" || resp.Data.Description == "" || resp.Data.SectionTitle == "" || len(resp.Data.Cards) == 0 || len(resp.Data.RecentRoutes) == 0 {
		t.Fatalf("unexpected map play page config: %s", string(body))
	}

	for _, pageKey := range []string{"city-atlas", "footprint-heatmap", "friend-city", "real-checkin"} {
		pageBody := getJSON(t, mux, "/api/app/map/play-pages?pageKey="+pageKey, token, http.StatusOK)
		if !bytes.Contains(pageBody, []byte(`"data"`)) {
			t.Fatalf("expected map play page %s config: %s", pageKey, string(pageBody))
		}
	}

	getJSON(t, mux, "/api/app/map/play-pages?pageKey=missing", token, http.StatusNotFound)
}

func TestMapCheckinSubmit(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	token := loginForTestWithCode(t, mux, "map-checkin-user")
	completeIdentityForTest(t, mux, token)

	postJSON(t, mux, "/api/app/map/checkins", token, `{"pointId":"bund-night","story":"今天外滩夜色很好，适合留下城市足迹。","fileIds":[1]}`, http.StatusOK)
	postJSON(t, mux, "/api/app/map/checkins", token, `{"story":"缺少对象","fileIds":[1]}`, http.StatusUnprocessableEntity)
	postJSON(t, mux, "/api/app/map/checkins", token, `{"pointId":"bund-night","story":"","fileIds":[1]}`, http.StatusUnprocessableEntity)
	postJSON(t, mux, "/api/app/map/checkins", token, `{"pointId":"bund-night","story":"缺少照片","fileIds":[]}`, http.StatusUnprocessableEntity)
}

func TestMapPlayActionsHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	token := loginForTestWithCode(t, mux, "map-play-actions-user")
	completeIdentityForTest(t, mux, token)

	routeBody := postJSON(t, mux, "/api/app/map/blind-routes", token, `{"cardId":"walk","title":"城市漫步盲盒"}`, http.StatusOK)
	var routeResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(routeBody, &routeResp); err != nil {
		t.Fatal(err)
	}
	if routeResp.Data.ID == 0 || routeResp.Data.Status != "opened" {
		t.Fatalf("expected opened blind route: %s", string(routeBody))
	}
	postJSON(t, mux, "/api/app/map/blind-routes/"+strconv.FormatInt(routeResp.Data.ID, 10)+"/complete", token, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/map/blind-routes", token, `{"cardId":"","title":""}`, http.StatusUnprocessableEntity)

	challengeBody := postJSON(t, mux, "/api/app/map/challenges", token, `{"title":"城市挑战","total":3}`, http.StatusOK)
	var challengeResp struct {
		Data struct {
			ID         int64  `json:"id"`
			StatusText string `json:"statusText"`
		} `json:"data"`
	}
	if err := json.Unmarshal(challengeBody, &challengeResp); err != nil {
		t.Fatal(err)
	}
	if challengeResp.Data.ID == 0 || challengeResp.Data.StatusText != "进行中" {
		t.Fatalf("expected challenge: %s", string(challengeBody))
	}
}

func TestMapProxyRequiresConfiguredProvider(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Register(mux)

	token := loginForTestWithCode(t, mux, "map-no-provider-user")
	completeIdentityForTest(t, mux, token)

	getJSON(t, mux, "/api/app/map/search?keyword=%E8%A5%BF%E6%B9%96&city=%E6%9D%AD%E5%B7%9E", token, http.StatusServiceUnavailable)
}

func TestExportTaskFlowRequiresPermissionAndGeneratesDownloadURL(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	noPermReq := httptest.NewRequest(http.MethodGet, "/api/admin/reports/export-templates", nil)
	noPermRec := httptest.NewRecorder()
	mux.ServeHTTP(noPermRec, noPermReq)
	if noPermRec.Code != http.StatusForbidden {
		t.Fatalf("expected export templates 403 without permission, got %d: %s", noPermRec.Code, noPermRec.Body.String())
	}

	templatesReq := httptest.NewRequest(http.MethodGet, "/api/admin/reports/export-templates", nil)
	templatesReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	templatesRec := httptest.NewRecorder()
	mux.ServeHTTP(templatesRec, templatesReq)
	if templatesRec.Code != http.StatusOK {
		t.Fatalf("expected export templates 200, got %d: %s", templatesRec.Code, templatesRec.Body.String())
	}
	var templatesResp struct {
		Data struct {
			Items []struct {
				Code       string `json:"code"`
				ExportType string `json:"exportType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(templatesRec.Body.Bytes(), &templatesResp); err != nil {
		t.Fatal(err)
	}
	if len(templatesResp.Data.Items) == 0 || templatesResp.Data.Items[0].Code == "" {
		t.Fatalf("expected fixed export templates: %s", templatesRec.Body.String())
	}
	if !hasExportTemplate(templatesResp.Data.Items, "reviews_default", "reviews") {
		t.Fatalf("expected reviews export template: %s", templatesRec.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/reports/export", bytes.NewBufferString(`{"templateCode":"reports_default","filters":{"status":"pending"}}`))
	createReq.Header.Set("X-Admin-ID", "88")
	createReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("expected export create 200, got %d: %s", createRec.Code, createRec.Body.String())
	}
	var createResp struct {
		Data struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatal(err)
	}
	if createResp.Data.ID != 1 || createResp.Data.Status != "pending" {
		t.Fatalf("expected pending export task: %s", createRec.Body.String())
	}

	runnerReq := httptest.NewRequest(http.MethodPost, "/api/internal/reports/export-runner", bytes.NewBufferString(`{}`))
	runnerReq.Header.Set("X-Admin-ID", "88")
	runnerReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	runnerRec := httptest.NewRecorder()
	mux.ServeHTTP(runnerRec, runnerReq)
	if runnerRec.Code != http.StatusOK {
		t.Fatalf("expected export runner 200, got %d: %s", runnerRec.Code, runnerRec.Body.String())
	}
	var runnerResp struct {
		Data struct {
			Processed int `json:"processed"`
			Items     []struct {
				Task struct {
					ID     int64  `json:"id"`
					Status string `json:"status"`
					FileID int64  `json:"fileId"`
				} `json:"task"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(runnerRec.Body.Bytes(), &runnerResp); err != nil {
		t.Fatal(err)
	}
	if runnerResp.Data.Processed != 1 || len(runnerResp.Data.Items) != 1 || runnerResp.Data.Items[0].Task.Status != "done" || runnerResp.Data.Items[0].Task.FileID == 0 {
		t.Fatalf("expected generated export file: %s", runnerRec.Body.String())
	}

	filteredTasksReq := httptest.NewRequest(http.MethodGet, "/api/admin/export-tasks?status=done&templateCode=reports_default&exportType=reports&createdBy=88", nil)
	filteredTasksReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	filteredTasksRec := httptest.NewRecorder()
	mux.ServeHTTP(filteredTasksRec, filteredTasksReq)
	if filteredTasksRec.Code != http.StatusOK {
		t.Fatalf("expected filtered export tasks 200, got %d: %s", filteredTasksRec.Code, filteredTasksRec.Body.String())
	}
	var filteredTasksResp struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				ID           int64  `json:"id"`
				Status       string `json:"status"`
				TemplateCode string `json:"templateCode"`
				ExportType   string `json:"exportType"`
				CreatedBy    int64  `json:"createdBy"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(filteredTasksRec.Body.Bytes(), &filteredTasksResp); err != nil {
		t.Fatal(err)
	}
	if filteredTasksResp.Data.Total != 1 || len(filteredTasksResp.Data.Items) != 1 || filteredTasksResp.Data.Items[0].CreatedBy != 88 {
		t.Fatalf("expected filtered export task by status/template/type/createdBy: %s", filteredTasksRec.Body.String())
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/admin/export-tasks/1/download-url", nil)
	downloadReq.Header.Set("X-Admin-ID", "88")
	downloadReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	downloadRec := httptest.NewRecorder()
	mux.ServeHTTP(downloadRec, downloadReq)
	if downloadRec.Code != http.StatusOK {
		t.Fatalf("expected export download-url 200, got %d: %s", downloadRec.Code, downloadRec.Body.String())
	}
	var downloadResp struct {
		Data struct {
			FileID      int64  `json:"fileId"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(downloadRec.Body.Bytes(), &downloadResp); err != nil {
		t.Fatal(err)
	}
	if downloadResp.Data.FileID == 0 || downloadResp.Data.DownloadURL == "" {
		t.Fatalf("expected export download url: %s", downloadRec.Body.String())
	}

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	noFilePermReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/1/download-url", nil)
	noFilePermReq.Header.Set("Authorization", "Bearer "+operatorToken)
	noFilePermRec := httptest.NewRecorder()
	mux.ServeHTTP(noFilePermRec, noFilePermReq)
	if noFilePermRec.Code != http.StatusForbidden {
		t.Fatalf("expected admin file download forbidden without matching permission, got %d: %s", noFilePermRec.Code, noFilePermRec.Body.String())
	}

	coverUploadBody := postAdminJSON(t, mux, "/api/admin/files/upload-token", adminLoginForTest(t, mux), `{"bizType":"game_cover","fileName":"cover.png","mimeType":"image/png","size":128}`, http.StatusOK)
	var coverUploadResp struct {
		Data struct {
			Upload struct {
				FileID    int64  `json:"fileId"`
				UploadURL string `json:"uploadUrl"`
			} `json:"upload"`
			File struct {
				ID      int64  `json:"fileId"`
				BizType string `json:"bizType"`
			} `json:"file"`
			Download struct {
				DownloadURL string `json:"downloadUrl"`
			} `json:"download"`
		} `json:"data"`
	}
	if err := json.Unmarshal(coverUploadBody, &coverUploadResp); err != nil {
		t.Fatal(err)
	}
	if coverUploadResp.Data.Upload.FileID == 0 || coverUploadResp.Data.Upload.UploadURL == "" || coverUploadResp.Data.File.BizType != "game_cover" || coverUploadResp.Data.Download.DownloadURL == "" {
		t.Fatalf("expected admin game cover upload token: %s", string(coverUploadBody))
	}
	postAdminJSON(t, mux, "/api/admin/files/upload-token", operatorToken, `{"bizType":"game_cover","fileName":"cover.png","mimeType":"image/png","size":128}`, http.StatusForbidden)

	server := newTestAppServer(auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore()), identity.NewService())
	server.files.CreateGeneratedFile("realname_material", 1, "id-card.png", "image/png", 128)
	server.files.CreateGeneratedFile("report_attachment", 1, "evidence.pdf", "application/pdf", 128)
	server.files.CreateGeneratedFile("export_file", 1, "export.csv", "text/csv", 128)
	fileMux := http.NewServeMux()
	server.Register(fileMux)
	fileOperatorToken := adminLoginForTestAs(t, fileMux, "operator", "admin123")

	identityReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/1/download-url", nil)
	identityReq.Header.Set("Authorization", "Bearer "+fileOperatorToken)
	identityRec := httptest.NewRecorder()
	fileMux.ServeHTTP(identityRec, identityReq)
	if identityRec.Code != http.StatusForbidden {
		t.Fatalf("expected realname material download to reject missing identity permission, got %d: %s", identityRec.Code, identityRec.Body.String())
	}
	adminIdentityReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/1/download-url", nil)
	adminIdentityReq.Header.Set("Authorization", "Bearer "+adminLoginForTestAs(t, fileMux, "admin", "admin123"))
	adminIdentityRec := httptest.NewRecorder()
	fileMux.ServeHTTP(adminIdentityRec, adminIdentityReq)
	if adminIdentityRec.Code != http.StatusOK {
		t.Fatalf("expected realname material download to need identity permission, got %d: %s", adminIdentityRec.Code, adminIdentityRec.Body.String())
	}

	adminReportReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/2/download-url", nil)
	adminReportReq.Header.Set("Authorization", "Bearer "+adminLoginForTestAs(t, fileMux, "admin", "admin123"))
	adminReportRec := httptest.NewRecorder()
	fileMux.ServeHTTP(adminReportRec, adminReportReq)
	if adminReportRec.Code != http.StatusOK {
		t.Fatalf("expected report attachment download with admin permission, got %d: %s", adminReportRec.Code, adminReportRec.Body.String())
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/3/download-url", nil)
	exportReq.Header.Set("Authorization", "Bearer "+fileOperatorToken)
	exportRec := httptest.NewRecorder()
	fileMux.ServeHTTP(exportRec, exportReq)
	if exportRec.Code != http.StatusForbidden {
		t.Fatalf("expected export file download to need report export permission, got %d: %s", exportRec.Code, exportRec.Body.String())
	}
	adminExportReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/3/download-url", nil)
	adminExportReq.Header.Set("Authorization", "Bearer "+adminLoginForTestAs(t, fileMux, "admin", "admin123"))
	adminExportRec := httptest.NewRecorder()
	fileMux.ServeHTTP(adminExportRec, adminExportReq)
	if adminExportRec.Code != http.StatusOK {
		t.Fatalf("expected super admin export file download, got %d: %s", adminExportRec.Code, adminExportRec.Body.String())
	}

	expiredFile, err := server.files.CreateGeneratedFileWithTTL("export_file", 1, "expired.csv", "text/csv", 128, -1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	expiredReq := httptest.NewRequest(http.MethodGet, "/api/admin/files/"+strconv.FormatInt(expiredFile.ID, 10)+"/download-url", nil)
	expiredReq.Header.Set("Authorization", "Bearer "+adminLoginForTestAs(t, fileMux, "admin", "admin123"))
	expiredRec := httptest.NewRecorder()
	fileMux.ServeHTTP(expiredRec, expiredReq)
	if expiredRec.Code != http.StatusGone {
		t.Fatalf("expected expired export file to return 410, got %d: %s", expiredRec.Code, expiredRec.Body.String())
	}

	logReq := httptest.NewRequest(http.MethodGet, "/api/admin/operation-logs", nil)
	logReq.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	logRec := httptest.NewRecorder()
	mux.ServeHTTP(logRec, logReq)
	if logRec.Code != http.StatusOK {
		t.Fatalf("expected operation logs 200, got %d: %s", logRec.Code, logRec.Body.String())
	}
	var logResp struct {
		Data struct {
			Items []struct {
				Action      string `json:"action"`
				TargetType  string `json:"targetType"`
				TargetID    string `json:"targetId"`
				AdminUserID int64  `json:"adminUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(logRec.Body.Bytes(), &logResp); err != nil {
		t.Fatal(err)
	}
	if !hasOperationLog(logResp.Data.Items, "export:create", "export_task", "1", 88) {
		t.Fatalf("expected export create operation log: %s", logRec.Body.String())
	}
	if !hasOperationLog(logResp.Data.Items, "export:run", "export_task", "1", 88) {
		t.Fatalf("expected export run operation log: %s", logRec.Body.String())
	}
	if !hasOperationLog(logResp.Data.Items, "export:download_url", "export_task", "1", 88) {
		t.Fatalf("expected export download operation log: %s", logRec.Body.String())
	}
}

func TestAdminLoginPermissionsAndTokenPermission(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	postJSON(t, mux, "/api/admin/auth/login", "", `{"username":"admin","password":"bad"}`, http.StatusUnauthorized)

	body := postJSON(t, mux, "/api/admin/auth/login", "", `{"username":"admin","password":"admin123"}`, http.StatusOK)
	var loginResp struct {
		Data struct {
			Token       string   `json:"token"`
			Roles       []string `json:"roles"`
			Permissions []string `json:"permissions"`
			AdminUser   struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			} `json:"adminUser"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		t.Fatal(err)
	}
	if loginResp.Data.Token == "" || loginResp.Data.AdminUser.ID != 1 || !hasString(loginResp.Data.Roles, "super_admin") || !hasString(loginResp.Data.Permissions, "operation_log:view_full") {
		t.Fatalf("expected admin login token and permissions: %s", string(body))
	}

	permReq := httptest.NewRequest(http.MethodGet, "/api/admin/permissions/tree", nil)
	permReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	permRec := httptest.NewRecorder()
	mux.ServeHTTP(permRec, permReq)
	if permRec.Code != http.StatusOK {
		t.Fatalf("expected permissions tree 200, got %d: %s", permRec.Code, permRec.Body.String())
	}
	var permResp struct {
		Data struct {
			Permissions []string `json:"permissions"`
			Menus       []struct {
				Code string `json:"code"`
			} `json:"menus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(permRec.Body.Bytes(), &permResp); err != nil {
		t.Fatal(err)
	}
	if !hasString(permResp.Data.Permissions, "report_export:create") || len(permResp.Data.Menus) == 0 {
		t.Fatalf("expected permission tree: %s", permRec.Body.String())
	}

	logReq := httptest.NewRequest(http.MethodGet, "/api/admin/operation-logs", nil)
	logReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	logRec := httptest.NewRecorder()
	mux.ServeHTTP(logRec, logReq)
	if logRec.Code != http.StatusOK {
		t.Fatalf("expected operation logs with admin token 200, got %d: %s", logRec.Code, logRec.Body.String())
	}

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	operatorBody := getAdminJSON(t, mux, "/api/admin/permissions/tree", operatorToken, http.StatusOK)
	var operatorResp struct {
		Data struct {
			Menus []struct {
				Code string `json:"code"`
			} `json:"menus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(operatorBody, &operatorResp); err != nil {
		t.Fatal(err)
	}
	if !hasMenuCode(operatorResp.Data.Menus, "games") || !hasMenuCode(operatorResp.Data.Menus, "reports") || !hasMenuCode(operatorResp.Data.Menus, "system") || !hasMenuCode(operatorResp.Data.Menus, "delivery") || !hasMenuCode(operatorResp.Data.Menus, "logs") {
		t.Fatalf("expected operator menus for games/reports/system/delivery/logs: %s", string(operatorBody))
	}
	for _, forbidden := range []string{"dashboard", "admins", "revenue"} {
		if hasMenuCode(operatorResp.Data.Menus, forbidden) {
			t.Fatalf("operator must not see menu %s: %s", forbidden, string(operatorBody))
		}
	}

	userManagerToken := adminLoginForTestAs(t, mux, "user_manager", "admin123")
	userManagerBody := getAdminJSON(t, mux, "/api/admin/permissions/tree", userManagerToken, http.StatusOK)
	var userManagerResp struct {
		Data struct {
			Permissions []string `json:"permissions"`
			Menus       []struct {
				Code string `json:"code"`
			} `json:"menus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(userManagerBody, &userManagerResp); err != nil {
		t.Fatal(err)
	}
	if !hasString(userManagerResp.Data.Permissions, "identity:update") {
		t.Fatalf("expected user manager identity update permission: %s", string(userManagerBody))
	}
	if !hasMenuCode(userManagerResp.Data.Menus, "audits") {
		t.Fatalf("expected user manager audits menu: %s", string(userManagerBody))
	}
	if !hasMenuCode(userManagerResp.Data.Menus, "redemption") {
		t.Fatalf("expected user manager redemption menu: %s", string(userManagerBody))
	}
	if !hasMenuCode(userManagerResp.Data.Menus, "members") {
		t.Fatalf("expected user manager members menu: %s", string(userManagerBody))
	}
	if !hasMenuCode(userManagerResp.Data.Menus, "profiles") {
		t.Fatalf("expected user manager profiles menu: %s", string(userManagerBody))
	}
	if !hasMenuCode(userManagerResp.Data.Menus, "growth") {
		t.Fatalf("expected user manager growth menu: %s", string(userManagerBody))
	}

	analystToken := adminLoginForTestAs(t, mux, "data_analyst", "admin123")
	analystBody := getAdminJSON(t, mux, "/api/admin/permissions/tree", analystToken, http.StatusOK)
	var analystResp struct {
		Data struct {
			Menus []struct {
				Code string `json:"code"`
			} `json:"menus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(analystBody, &analystResp); err != nil {
		t.Fatal(err)
	}
	if !hasMenuCode(analystResp.Data.Menus, "analytics") || !hasMenuCode(analystResp.Data.Menus, "delivery") {
		t.Fatalf("expected data analyst analytics and delivery menus: %s", string(analystBody))
	}
}

func TestAdminSystemReadiness(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	server := newTestAppServer(authService, identityService)
	server.Configure(config.Config{
		AppEnv:    "production",
		JWTSecret: "change-me",
	})
	server.Register(mux)

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	getAdminJSON(t, mux, "/api/admin/system/readiness", operatorToken, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	body := getAdminJSON(t, mux, "/api/admin/system/readiness", adminToken, http.StatusOK)
	var resp struct {
		Data struct {
			Production bool `json:"production"`
			Ready      bool `json:"ready"`
			Items      []struct {
				Key      string `json:"key"`
				Status   string `json:"status"`
				Required bool   `json:"required"`
				Message  string `json:"message"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Data.Production || resp.Data.Ready {
		t.Fatalf("expected production readiness to report missing config: %s", string(body))
	}
	if !hasReadinessItem(resp.Data.Items, "jwt_secret", "missing", true) || !hasReadinessItem(resp.Data.Items, "wechat_login", "missing", true) {
		t.Fatalf("expected missing critical readiness items: %s", string(body))
	}
	if strings.Contains(string(body), "change-me") {
		t.Fatalf("readiness response must not leak secret values: %s", string(body))
	}
}

func TestAdminAccountRoleAndPermissionCatalog(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	getAdminJSON(t, mux, "/api/admin/admin-users", operatorToken, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	usersBody := getAdminJSON(t, mux, "/api/admin/admin-users", adminToken, http.StatusOK)
	var usersResp struct {
		Data struct {
			Items []struct {
				ID              int64    `json:"id"`
				Username        string   `json:"username"`
				Roles           []string `json:"roles"`
				PermissionCount int      `json:"permissionCount"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(usersBody, &usersResp); err != nil {
		t.Fatal(err)
	}
	if usersResp.Data.Total < 3 || usersResp.Data.Items[0].Username != "admin" || !hasString(usersResp.Data.Items[0].Roles, "super_admin") || usersResp.Data.Items[0].PermissionCount == 0 {
		t.Fatalf("expected admin account list: %s", string(usersBody))
	}

	rolesBody := getAdminJSON(t, mux, "/api/admin/admin-roles", adminToken, http.StatusOK)
	var rolesResp struct {
		Data struct {
			Items []struct {
				Code            string   `json:"code"`
				Name            string   `json:"name"`
				Permissions     []string `json:"permissions"`
				PermissionCount int      `json:"permissionCount"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rolesBody, &rolesResp); err != nil {
		t.Fatal(err)
	}
	if len(rolesResp.Data.Items) != 7 || !hasRoleCode(rolesResp.Data.Items, "finance_manager") || !hasRoleCode(rolesResp.Data.Items, "game_manager") {
		t.Fatalf("expected seven admin role definitions: %s", string(rolesBody))
	}

	catalogBody := getAdminJSON(t, mux, "/api/admin/admin-permissions/catalog", adminToken, http.StatusOK)
	var catalogResp struct {
		Data struct {
			Items []struct {
				Code   string `json:"code"`
				Module string `json:"module"`
				Action string `json:"action"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(catalogBody, &catalogResp); err != nil {
		t.Fatal(err)
	}
	if !hasPermissionCode(catalogResp.Data.Items, "admin_user:view") || !hasPermissionCode(catalogResp.Data.Items, "game:create_admin") {
		t.Fatalf("expected permission catalog: %s", string(catalogBody))
	}
	if !hasPermissionCode(catalogResp.Data.Items, "admin_user:update") {
		t.Fatalf("expected admin update permission in catalog: %s", string(catalogBody))
	}
}

func TestAdminAccountMutationUpdatesRBAC(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	postAdminJSON(t, mux, "/api/admin/admin-users", operatorToken, `{"username":"blocked","password":"admin123","roles":["finance_manager"],"status":"active"}`, http.StatusForbidden)

	adminToken := adminLoginForTest(t, mux)
	body := postAdminJSON(t, mux, "/api/admin/admin-users", adminToken, `{"username":"finance_ops","password":"admin123","roles":["finance_manager"],"status":"active"}`, http.StatusOK)
	var createResp struct {
		Data struct {
			ID              int64    `json:"id"`
			Username        string   `json:"username"`
			Roles           []string `json:"roles"`
			Status          string   `json:"status"`
			PermissionCount int      `json:"permissionCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &createResp); err != nil {
		t.Fatal(err)
	}
	if createResp.Data.ID == 0 || createResp.Data.Username != "finance_ops" || !hasString(createResp.Data.Roles, "finance_manager") || createResp.Data.PermissionCount == 0 {
		t.Fatalf("expected created finance admin: %s", string(body))
	}

	financeToken := adminLoginForTestAs(t, mux, "finance_ops", "admin123")
	postAdminJSON(t, mux, "/api/admin/revenue/templates", financeToken, `{"name":"finance-managed","gameType":"free","platformBps":1000,"creatorBps":3000,"memberBps":6000}`, http.StatusOK)
	financeLogsBody := getAdminJSON(t, mux, "/api/admin/operation-logs", financeToken, http.StatusOK)
	var financeLogsResp struct {
		Data struct {
			Items []struct {
				AdminUserID int64  `json:"adminUserId"`
				Action      string `json:"action"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(financeLogsBody, &financeLogsResp); err != nil {
		t.Fatal(err)
	}
	if len(financeLogsResp.Data.Items) == 0 {
		t.Fatalf("expected finance admin to view own operation logs: %s", string(financeLogsBody))
	}
	for _, item := range financeLogsResp.Data.Items {
		if item.AdminUserID != createResp.Data.ID {
			t.Fatalf("expected self operation logs only for finance admin: %s", string(financeLogsBody))
		}
	}

	updatePath := "/api/admin/admin-users/" + strconv.FormatInt(createResp.Data.ID, 10)
	putJSON(t, mux, updatePath, adminLoginForTestAs(t, mux, "user_manager", "admin123"), `{"roles":["customer_manager"],"status":"disabled"}`, http.StatusForbidden)
	putJSON(t, mux, updatePath, adminToken, `{"roles":["customer_manager"],"status":"disabled"}`, http.StatusOK)
	getAdminJSON(t, mux, "/api/admin/revenue/templates", financeToken, http.StatusForbidden)
	postJSON(t, mux, "/api/admin/auth/login", "", `{"username":"finance_ops","password":"admin123"}`, http.StatusForbidden)
}

func TestAdminRolePermissionBoundariesHTTP(t *testing.T) {
	mux := http.NewServeMux()
	authService := auth.NewService(users.NewStore(), invites.NewStore(), auth.NewTokenStore())
	identityService := identity.NewService()
	newTestAppServer(authService, identityService).Register(mux)

	userToken := loginForTestWithCode(t, mux, "admin-boundary-user")
	completeIdentityForTest(t, mux, userToken)

	analystToken := adminLoginForTestAs(t, mux, "data_analyst", "admin123")
	getAdminJSON(t, mux, "/api/admin/behavior/events", analystToken, http.StatusOK)
	getAdminJSON(t, mux, "/api/admin/test-cases", analystToken, http.StatusOK)
	postAdminJSON(t, mux, "/api/admin/test-cases", analystToken, `{"module":"E7","caseName":"write denied","priority":"P1","expectedResult":"denied"}`, http.StatusForbidden)
	getAdminJSON(t, mux, "/api/admin/redemption/orders", analystToken, http.StatusForbidden)
	getAdminJSON(t, mux, "/api/admin/operation-logs", analystToken, http.StatusOK)
	postAdminJSON(t, mux, "/api/admin/revenue/templates", analystToken, `{"name":"blocked","platformBps":1000,"expertBps":8000,"guideBps":1000}`, http.StatusForbidden)

	operatorToken := adminLoginForTestAs(t, mux, "operator", "admin123")
	getAdminJSON(t, mux, "/api/admin/users/1", operatorToken, http.StatusOK)
	getAdminJSON(t, mux, "/api/admin/experts/1/skills", operatorToken, http.StatusForbidden)
	getAdminJSON(t, mux, "/api/admin/test-cases", operatorToken, http.StatusForbidden)
	operatorLogsBody := getAdminJSON(t, mux, "/api/admin/operation-logs?adminUserId=1", operatorToken, http.StatusOK)
	var operatorLogsResp struct {
		Data struct {
			Items []struct {
				AdminUserID int64 `json:"adminUserId"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(operatorLogsBody, &operatorLogsResp); err != nil {
		t.Fatal(err)
	}
	for _, item := range operatorLogsResp.Data.Items {
		if item.AdminUserID != 3 {
			t.Fatalf("expected operator to see only own logs even with adminUserId filter: %s", string(operatorLogsBody))
		}
	}
}

func loginForTest(t *testing.T, mux *http.ServeMux) string {
	t.Helper()
	return loginForTestWithCode(t, mux, "u1")
}

func loginForTestWithCode(t *testing.T, mux *http.ServeMux, code string) string {
	t.Helper()
	return loginForTestWithCodeAndInvite(t, mux, code, "TEST2026")
}

func loginForTestWithCodeAndInvite(t *testing.T, mux *http.ServeMux, code string, inviteCode string) string {
	t.Helper()
	body := bytes.NewBufferString(`{"code":"` + code + `","inviteCode":"` + inviteCode + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/app/auth/wechat-login", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", rec.Code, rec.Body.String())
	}
	var login struct {
		Data struct {
			Token        string `json:"token"`
			PreAuthToken string `json:"preAuthToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	if login.Data.Token != "" {
		return login.Data.Token
	}
	if login.Data.PreAuthToken == "" {
		t.Fatalf("expected login token: %s", rec.Body.String())
	}
	return login.Data.PreAuthToken
}

func adminLoginForTest(t *testing.T, mux *http.ServeMux) string {
	t.Helper()
	return adminLoginForTestAs(t, mux, "admin", "admin123")
}

func loginAdminForTest(t *testing.T, mux *http.ServeMux, username ...string) string {
	t.Helper()
	if len(username) > 0 && strings.TrimSpace(username[0]) != "" {
		return adminLoginForTestAs(t, mux, username[0], "admin123")
	}
	return adminLoginForTest(t, mux)
}

func adminLoginForTestAs(t *testing.T, mux *http.ServeMux, username string, password string) string {
	t.Helper()
	body := postJSON(t, mux, "/api/admin/auth/login", "", `{"username":"`+username+`","password":"`+password+`"}`, http.StatusOK)
	var login struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &login); err != nil {
		t.Fatal(err)
	}
	if login.Data.Token == "" {
		t.Fatal("expected admin token")
	}
	return login.Data.Token
}

func getAdminJSON(t *testing.T, mux *http.ServeMux, path string, adminToken string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func completeIdentityForTest(t *testing.T, mux *http.ServeMux, token string) {
	t.Helper()
	identityPhoneSeq++
	phone := fmt.Sprintf("138%08d", identityPhoneSeq)
	postJSON(t, mux, "/api/app/identity/phone/bind", token, `{"phone":"`+phone+`"}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/send-code", token, `{}`, http.StatusOK)
	postJSON(t, mux, "/api/app/sms/verify-code", token, `{"code":"000000"}`, http.StatusOK)
	body := postJSON(t, mux, "/api/app/identity/phone/verify", token, `{"realName":"User","idCard":"110101199001011234"}`, http.StatusOK)
	var submitResp struct {
		Data struct {
			UserID int64 `json:"userId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &submitResp); err != nil {
		t.Fatal(err)
	}
	if submitResp.Data.UserID == 0 {
		t.Fatalf("expected manual realname user id: %s", string(body))
	}
	adminToken := adminLoginForTest(t, mux)
	postAdminJSON(t, mux, "/api/admin/identity-verifications/"+strconv.FormatInt(submitResp.Data.UserID, 10)+"/review", adminToken, `{"approve":true,"reason":"test"}`, http.StatusOK)
	faceResp := postJSON(t, mux, "/api/app/identity/faceid/detect-auth", token, `{}`, http.StatusOK)
	var face struct {
		Data struct {
			FaceToken string `json:"faceToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(faceResp, &face); err != nil {
		t.Fatal(err)
	}
	postJSON(t, mux, "/api/app/identity/faceid/callback", token, `{"faceToken":"`+face.Data.FaceToken+`"}`, http.StatusOK)
}

func approveExtraMembersForHTTP(t *testing.T, mux *http.ServeMux, creatorToken string, gameID int64, prefix string, count int) []string {
	t.Helper()
	tokens := make([]string, 0, count)
	for i := 0; i < count; i++ {
		token := loginForTestWithCode(t, mux, prefix+"-extra-"+strconv.Itoa(i+1))
		completeIdentityForTest(t, mux, token)
		tokens = append(tokens, token)
		appBody := postJSON(t, mux, "/api/app/games/"+strconv.FormatInt(gameID, 10)+"/applications", token, `{"reason":"join"}`, http.StatusOK)
		var appResp struct {
			Data struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(appBody, &appResp); err != nil {
			t.Fatal(err)
		}
		if appResp.Data.ID == 0 {
			t.Fatalf("expected extra member application id: %s", string(appBody))
		}
		postJSON(t, mux, "/api/app/game-applications/"+strconv.FormatInt(appResp.Data.ID, 10)+"/audit", creatorToken, `{"approve":true}`, http.StatusOK)
	}
	return tokens
}

func completeGameReviewsForHTTP(t *testing.T, mux *http.ServeMux, gameID int64, orderedTokens []string) {
	t.Helper()
	userIDs := make([]int64, 0, len(orderedTokens))
	for _, token := range orderedTokens {
		userIDs = append(userIDs, currentUserIDForTest(t, mux, token))
	}
	for reviewerIndex, reviewerToken := range orderedTokens {
		reviewerID := userIDs[reviewerIndex]
		for _, targetID := range userIDs {
			if targetID == reviewerID {
				continue
			}
			postJSON(t, mux, "/api/app/reviews", reviewerToken, `{"gameId":`+strconv.FormatInt(gameID, 10)+`,"targetUserId":`+strconv.FormatInt(targetID, 10)+`,"targetRole":"member","score":5}`, http.StatusOK)
		}
	}
}

func currentUserIDForTest(t *testing.T, mux *http.ServeMux, token string) int64 {
	t.Helper()
	body := getJSON(t, mux, "/api/app/users/me", token, http.StatusOK)
	var resp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ID == 0 {
		t.Fatalf("expected current user id: %s", string(body))
	}
	return resp.Data.ID
}

func postJSON(t *testing.T, mux *http.ServeMux, path string, token string, body string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func postJSONWithIdempotencyKey(t *testing.T, mux *http.ServeMux, path string, token string, key string, body string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", key)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func postAdminJSON(t *testing.T, mux *http.ServeMux, path string, adminToken string, body string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func postAdminJSONWithPermission(t *testing.T, mux *http.ServeMux, path string, permission string, body string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func getJSON(t *testing.T, mux *http.ServeMux, path string, token string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func getAdminJSONWithPermission(t *testing.T, mux *http.ServeMux, path string, permission string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+adminLoginForTest(t, mux))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func putJSON(t *testing.T, mux *http.ServeMux, path string, token string, body string, expected int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != expected {
		t.Fatalf("%s expected %d, got %d: %s", path, expected, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}

func hasFeedbackServiceMessage(messages []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}, keyword string) bool {
	for _, message := range messages {
		if message.Role == "service" && strings.Contains(message.Content, keyword) {
			return true
		}
	}
	return false
}

func faceIDCallbackTestSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func hasBehaviorEvent[T any](items []T, eventType string, targetID int64) bool {
	for _, item := range items {
		value := reflect.ValueOf(item)
		if value.Kind() == reflect.Pointer {
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			continue
		}
		eventField := value.FieldByName("EventType")
		targetField := value.FieldByName("TargetID")
		if eventField.IsValid() && targetField.IsValid() && eventField.Kind() == reflect.String && targetField.CanInt() && eventField.String() == eventType && targetField.Int() == targetID {
			return true
		}
	}
	return false
}

func hasOperationLog(items []struct {
	Action      string `json:"action"`
	TargetType  string `json:"targetType"`
	TargetID    string `json:"targetId"`
	AdminUserID int64  `json:"adminUserId"`
}, action string, targetType string, targetID string, adminUserID int64) bool {
	for _, item := range items {
		if item.Action == action && item.TargetType == targetType && item.TargetID == targetID && (adminUserID == 0 || item.AdminUserID == adminUserID) {
			return true
		}
	}
	return false
}

func hasRevenueItemRole(items []struct {
	Role       string `json:"role"`
	AmountCent int64  `json:"amountCent"`
}, role string, amountCent int64) bool {
	for _, item := range items {
		if item.Role == role && item.AmountCent == amountCent {
			return true
		}
	}
	return false
}

func hasString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func hasMenuCode(items []struct {
	Code string `json:"code"`
}, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}

func hasRoleCode(items []struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Permissions     []string `json:"permissions"`
	PermissionCount int      `json:"permissionCount"`
}, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}

func hasPermissionCode(items []struct {
	Code   string `json:"code"`
	Module string `json:"module"`
	Action string `json:"action"`
}, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}

func hasReadinessItem(items []struct {
	Key      string `json:"key"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Message  string `json:"message"`
}, key string, status string, required bool) bool {
	for _, item := range items {
		if item.Key == key && item.Status == status && item.Required == required {
			return true
		}
	}
	return false
}

func hasExportTemplate(items []struct {
	Code       string `json:"code"`
	ExportType string `json:"exportType"`
}, code string, exportType string) bool {
	for _, item := range items {
		if item.Code == code && item.ExportType == exportType {
			return true
		}
	}
	return false
}

type connectionTestItem struct {
	ID              int64  `json:"id"`
	ConnectedUserID int64  `json:"connectedUserId"`
	RelationType    string `json:"relationType"`
	SourceType      string `json:"sourceType"`
	StrengthScore   int    `json:"strengthScore"`
}

func hasConnectionSource(items []connectionTestItem, sourceType string) bool {
	for _, item := range items {
		if item.SourceType == sourceType {
			return true
		}
	}
	return false
}

func hasConnectionStrength(items []connectionTestItem, connectedUserID int64, sourceType string, minStrength int) bool {
	for _, item := range items {
		if item.ConnectedUserID == connectedUserID && item.SourceType == sourceType && item.StrengthScore >= minStrength {
			return true
		}
	}
	return false
}

func hasFootprintAction(items []struct {
	Action string `json:"action"`
}, action string) bool {
	for _, item := range items {
		if item.Action == action {
			return true
		}
	}
	return false
}

func hasHTTPConfirmItemFileIDs(items []struct {
	UserID  int64   `json:"userId"`
	FileIDs []int64 `json:"fileIds"`
}, userID int64, fileIDs []int64) bool {
	for _, item := range items {
		if item.UserID != userID || len(item.FileIDs) != len(fileIDs) {
			continue
		}
		matched := true
		for index := range fileIDs {
			if item.FileIDs[index] != fileIDs[index] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func hasActionKey(items []struct {
	Key string `json:"key"`
}, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}
	return false
}

func firstConnectionID(items []connectionTestItem) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[0].ID
}

func hasNotification(items []struct {
	ID               int64  `json:"id"`
	NotifyType       string `json:"notifyType"`
	Status           string `json:"status"`
	WechatState      string `json:"wechatState"`
	WechatTemplateID string `json:"wechatTemplateId"`
	WechatTaskID     int64  `json:"wechatTaskId"`
}, notifyType string, status string) bool {
	for _, item := range items {
		if item.NotifyType == notifyType && item.Status == status {
			return true
		}
	}
	return false
}

func countNotificationsByType(t *testing.T, body []byte, notifyType string) int {
	t.Helper()
	var resp struct {
		Data struct {
			Items []struct {
				NotifyType string `json:"notifyType"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, item := range resp.Data.Items {
		if item.NotifyType == notifyType {
			count++
		}
	}
	return count
}

func containsInt64(items []int64, expected int64) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}
