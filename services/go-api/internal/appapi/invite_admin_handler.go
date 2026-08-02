package appapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/users"
)

func (s *Server) adminInviteCodes(w http.ResponseWriter, r *http.Request) {
	ownerID, _, ok := optionalInt64Query(w, r, "ownerUserId")
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{
		Status:    status,
		EntryType: strings.TrimSpace(r.URL.Query().Get("entryType")),
		OwnerID:   ownerID,
	})
	if err != nil {
		if errors.Is(err, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请码入口类型不正确")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码列表失败")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

// adminInviteOwners returns the platform official account and users who can
// legally own registration invitations. The keyword is matched server-side so
// a full phone number or real name never has to be downloaded before searching.
func (s *Server) adminInviteOwners(w http.ResponseWriter, r *http.Request) {
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	// 邀请人选择器只能通过关键词检索，避免把全部领路人一次性暴露到后台页面。
	if keyword == "" {
		httpx.OK(w, map[string]interface{}{"items": []map[string]interface{}{}, "total": 0})
		return
	}
	_, allowSensitive := s.admins.HasPermission(s.adminToken(r), "identity:sensitive:read")
	usersList, err := s.auth.AdminUsers(users.Filter{Status: "active"})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "获取可选邀请人失败")
		return
	}

	recordByUserID := make(map[int64]identity.Record)
	for _, record := range s.identity.AllRecords() {
		recordByUserID[record.UserID] = record
	}

	const maxResults = 30
	items := make([]map[string]interface{}, 0, maxResults)
	for _, user := range usersList {
		if !s.userCanOwnAdminInviteCodes(user.ID) {
			continue
		}
		record := recordByUserID[user.ID]
		if keyword != "" && !userMatchesInviteOwnerKeyword(user, record, keyword, s.identity) {
			continue
		}

		phoneMasked := strings.TrimSpace(user.PhoneMasked)
		if phoneMasked == "" {
			phoneMasked = strings.TrimSpace(record.PhoneMasked)
		}
		roleNames := make([]string, 0, 2)
		if user.IsPlatformOfficial() {
			roleNames = append(roleNames, "平台官方")
		} else {
			roles := s.profiles.RoleSnapshot(user.ID).RoleStatusMap
			if roles["expert"] == "approved" || roles["expert"] == "active" {
				roleNames = append(roleNames, "行家")
			}
			if roles["guide"] == "approved" || roles["guide"] == "active" {
				roleNames = append(roleNames, "领路人")
			}
		}
		displayName := strings.TrimSpace(user.Nickname)
		if allowSensitive {
			if plain, revealErr := s.identity.RevealRecord(record); revealErr == nil && strings.TrimSpace(plain.RealName) != "" {
				displayName = strings.TrimSpace(plain.RealName)
			}
		}
		items = append(items, map[string]interface{}{
			"id":             user.ID,
			"nickname":       strings.TrimSpace(user.Nickname),
			"displayName":    displayName,
			"realNameMasked": strings.TrimSpace(record.RealNameMasked),
			"phoneMasked":    phoneMasked,
			"roles":          roleNames,
		})
		if len(items) >= maxResults {
			break
		}
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func userMatchesInviteOwnerKeyword(user users.User, record identity.Record, keyword string, identityService identityService) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	if adminSearchContains(strconv.FormatInt(user.ID, 10), keyword) ||
		adminSearchContains(user.Nickname, keyword) ||
		adminSearchContains(user.PhoneMasked, keyword) {
		return true
	}
	return identityRecordMatchesInviteOwnerKeyword(record, keyword, identityService)
}

func identityRecordMatchesInviteOwnerKeyword(record identity.Record, keyword string, identityService identityService) bool {
	if record.UserID <= 0 {
		return false
	}
	if adminSearchContains(record.PhoneMasked, keyword) || adminSearchContains(record.RealNameMasked, keyword) {
		return true
	}
	plain, err := identityService.RevealRecord(record)
	if err != nil {
		return false
	}
	return adminSearchContains(plain.Phone, keyword) || adminSearchContains(plain.RealName, keyword)
}

func (s *Server) createAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prefix      string `json:"prefix"`
		OwnerUserID int64  `json:"ownerUserId"`
		EntryType   string `json:"entryType"`
		BatchCount  int    `json:"batchCount"`
		ExpiresAt   string `json:"expiresAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	config, err := s.inviteCodeConfigStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请码配置失败，请稍后重试")
		return
	}
	var expiresAt time.Time
	if strings.TrimSpace(req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
		if err != nil || !parsed.After(time.Now()) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "有效期必须是未来时间")
			return
		}
		expiresAt = parsed
	} else if config.DefaultValidDays > 0 {
		expiresAt = time.Now().AddDate(0, 0, config.DefaultValidDays)
	}
	if req.OwnerUserID <= 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请选择邀请人")
		return
	}
	if _, found := s.auth.UserByID(req.OwnerUserID); !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "邀请人不存在")
		return
	}
	if !s.userCanOwnAdminInviteCodes(req.OwnerUserID) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请人必须是平台官方账号或已生效的领路人，请更换用户 ID")
		return
	}
	if req.BatchCount < 1 || req.BatchCount > config.MaxBatchCount {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "生成数量不正确")
		return
	}
	prefix, ok := normalizeInviteCodePrefix(req.Prefix)
	if !ok {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请码前缀必须为三位英文字母")
		return
	}
	const maxUses = 1
	startSerial, err := s.nextInviteCodeSerial(prefix)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请码流水失败")
		return
	}
	if startSerial+req.BatchCount-1 > 999999 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "该邀请码前缀的六位流水号已用尽，请更换前缀")
		return
	}
	codes := make([]string, 0, req.BatchCount)
	for offset := 0; offset < req.BatchCount; offset++ {
		codes = append(codes, prefix+fmt.Sprintf("%06d", startSerial+offset))
	}
	items, createErr := s.auth.AdminCreateInviteCodes(codes, req.OwnerUserID, maxUses, req.EntryType, expiresAt)
	if createErr != nil {
		if errors.Is(createErr, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请码入口类型不正确")
			return
		}
		if errors.Is(createErr, invites.ErrInviteCodeExists) {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "邀请码流水已变化，请刷新后重新生成")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "生成邀请码失败")
		return
	}
	action := "invite_code:batch_create"
	if req.BatchCount == 1 {
		action = "invite_code:create"
	}
	s.recordOperation(r, action, "invite_code", "", map[string]interface{}{
		"prefix":      prefix,
		"entryType":   req.EntryType,
		"batchCount":  req.BatchCount,
		"ownerUserId": req.OwnerUserID,
	})
	if req.BatchCount == 1 {
		httpx.OK(w, items[0])
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func normalizeInviteCodePrefix(value string) (string, bool) {
	prefix := strings.ToUpper(strings.TrimSpace(value))
	if len(prefix) != 3 {
		return "", false
	}
	for _, ch := range prefix {
		if ch < 'A' || ch > 'Z' {
			return "", false
		}
	}
	return prefix, true
}

func (s *Server) nextInviteCodeSerial(prefix string) (int, error) {
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{})
	if err != nil {
		return 0, err
	}
	next := 1
	for _, item := range items {
		code := strings.ToUpper(strings.TrimSpace(item.Code))
		if !strings.HasPrefix(code, prefix) || len(code) != len(prefix)+6 {
			continue
		}
		serial, parseErr := strconv.Atoi(code[len(prefix):])
		if parseErr == nil && serial >= next {
			next = serial + 1
		}
	}
	return next, nil
}

func (s *Server) adminInviteCodeDetail(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/admin/invite-codes/"))
	code = strings.Trim(code, "/")
	if strings.HasSuffix(code, "/materials") {
		code = strings.TrimSuffix(code, "/materials")
		code = strings.Trim(code, "/")
		s.adminInviteCodeMaterials(w, r, code)
		return
	}
	if code == "" || strings.Contains(code, "/") {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "缺少邀请码")
		return
	}
	invite, found, err := s.adminInviteByCode(code)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码列表失败")
		return
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "邀请码不存在")
		return
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviteCodeID: invite.ID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码绑定关系失败")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"inviteCode": invite,
		"relations":  relations,
		"ownerUser":  s.adminUserSummary(invite.OwnerID),
		"boundUser":  s.adminUserSummary(invite.BoundWechatUserID),
	})
}

func (s *Server) adminInviteByCode(code string) (invites.InviteCode, bool, error) {
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{})
	if err != nil {
		return invites.InviteCode{}, false, err
	}
	var invite invites.InviteCode
	for _, item := range items {
		if item.Code == code {
			return item, true, nil
		}
	}
	return invite, false, nil
}

func (s *Server) adminInviteCodeMaterials(w http.ResponseWriter, r *http.Request, code string) {
	if code == "" || strings.Contains(code, "/") {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "缺少邀请码")
		return
	}
	invite, found, err := s.adminInviteByCode(code)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码列表失败")
		return
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "邀请码不存在")
		return
	}
	materials := s.inviteMaterialBase(invite)
	if s.cfg.Wechat.AppID == "" || s.cfg.Wechat.AppSecret == "" {
		materials["available"] = false
		materials["ready"] = false
		materials["message"] = "服务器未配置微信 AppID 或 AppSecret，暂不能生成可直接发送给用户的链接和小程序码。"
		materials["missingConfig"] = []string{"WECHAT_APP_ID", "WECHAT_APP_SECRET"}
		materials["copyText"] = inviteMaterialUnavailableText(invite, "服务器未配置微信 AppID 或 AppSecret")
		httpx.OK(w, materials)
		return
	}
	token, err := s.wechatAccessToken(r.Context())
	if err != nil {
		materials["available"] = false
		materials["ready"] = false
		materials["message"] = "获取微信接口凭证失败，暂不能生成可直接发送给用户的链接和小程序码。"
		materials["copyText"] = inviteMaterialUnavailableText(invite, "获取微信接口凭证失败："+err.Error())
		httpx.OK(w, materials)
		return
	}
	urlLink := ""
	wxaCodeDataURL := ""
	if urlLink, err = s.wechatURLLink(r.Context(), token, materials["page"].(string), materials["query"].(string)); err == nil {
		materials["urlLink"] = urlLink
	} else {
		materials["urlLinkError"] = err.Error()
	}
	if image, mimeType, err := s.wechatWxaCode(r.Context(), token, materials["page"].(string), materials["scene"].(string)); err == nil {
		wxaCodeDataURL = "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(image)
		materials["wxaCodeDataUrl"] = wxaCodeDataURL
		materials["wxaCodeMimeType"] = mimeType
	} else {
		materials["wxaCodeError"] = err.Error()
	}
	available := urlLink != "" || wxaCodeDataURL != ""
	materials["available"] = available
	materials["ready"] = available
	materials["urlLinkReady"] = urlLink != ""
	materials["wxaCodeReady"] = wxaCodeDataURL != ""
	if available {
		switch {
		case urlLink != "" && wxaCodeDataURL != "":
			materials["message"] = "已生成可发给用户的邀请入口。链接可直接复制发送；二维码可下载后放入海报或发给用户扫码。"
		case urlLink != "":
			materials["message"] = "已生成可点击邀请链接，可直接复制发送给用户。二维码暂未生成。"
		default:
			materials["message"] = "已生成小程序码，可下载后放入海报或发给体验成员扫码。可点击链接需要小程序正式发布后再生成。"
		}
		materials["copyText"] = inviteMaterialShareText(invite, urlLink, wxaCodeDataURL != "")
	} else {
		materials["message"] = "微信接口已连接，但链接和小程序码都生成失败，暂不能直接发给用户。"
		materials["copyText"] = inviteMaterialUnavailableText(invite, "微信链接和小程序码生成失败")
	}
	httpx.OK(w, materials)
}

func (s *Server) inviteMaterialBase(invite invites.InviteCode) map[string]interface{} {
	entryType := invites.NormalizeEntryType(invite.EntryType)
	query := "inviteCode=" + url.QueryEscape(invite.Code) + "&entryType=" + url.QueryEscape(entryType)
	scene := invite.Code
	page := "pages/login/invite/index"
	return map[string]interface{}{
		"inviteCode": invite.Code,
		"entryType":  entryType,
		"entryLabel": inviteEntryTypeLabel(entryType),
		"page":       page,
		"path":       "/" + page + "?" + query,
		"query":      query,
		"scene":      scene,
		"shareTitle": "真好玩邀请入口",
		"tip":        "用户点击链接或扫码进入小程序后，登录时会自动校验邀请码并绑定微信用户。",
	}
}

func inviteEntryTypeLabel(entryType string) string {
	switch entryType {
	case invites.EntryTypePoster:
		return "小程序卡片"
	case invites.EntryTypeQRCode:
		return "二维码"
	case invites.EntryTypeLink:
		return "链接"
	default:
		return entryType
	}
}

func inviteMaterialShareText(invite invites.InviteCode, urlLink string, hasWxaCode bool) string {
	entryLabel := inviteEntryTypeLabel(invites.NormalizeEntryType(invite.EntryType))
	lines := []string{
		"真好玩邀请入口",
		"邀请码：" + invite.Code,
		"入口方式：" + entryLabel,
	}
	if urlLink != "" {
		lines = append(lines, "点击进入小程序："+urlLink)
	}
	if hasWxaCode {
		if urlLink != "" {
			lines = append(lines, "也可以使用随附的小程序码扫码进入。")
		} else {
			lines = append(lines, "请使用随附的小程序码扫码进入。")
			lines = append(lines, "可点击链接需要小程序正式发布后再生成。")
		}
	}
	lines = append(lines, "进入后请按提示登录，系统会自动绑定该邀请码。")
	return strings.Join(lines, "\n")
}

func inviteMaterialUnavailableText(invite invites.InviteCode, reason string) string {
	entryLabel := inviteEntryTypeLabel(invites.NormalizeEntryType(invite.EntryType))
	return strings.Join([]string{
		"该邀请码暂不能直接发送给用户。",
		"邀请码：" + invite.Code,
		"入口方式：" + entryLabel,
		"原因：" + reason,
		"请先完成微信小程序 AppID/AppSecret 配置，再回到后台重新生成物料。",
	}, "\n")
}

func (s *Server) wechatAccessToken(ctx context.Context) (string, error) {
	endpoint := "https://api.weixin.qq.com/cgi-bin/token"
	values := url.Values{}
	values.Set("grant_type", "client_credential")
	values.Set("appid", s.cfg.Wechat.AppID)
	values.Set("secret", s.cfg.Wechat.AppSecret)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var payload struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New(payload.ErrMsg)
	}
	return payload.AccessToken, nil
}

func (s *Server) wechatURLLink(ctx context.Context, accessToken string, page string, query string) (string, error) {
	body := map[string]interface{}{
		"path":        page,
		"query":       query,
		"env_version": s.wechatMiniProgramEnvVersion(),
		"is_expire":   false,
	}
	var payload struct {
		URLLink string `json:"url_link"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := s.wechatPostJSON(ctx, "/wxa/generate_urllink", accessToken, body, &payload); err != nil {
		return "", err
	}
	if payload.URLLink == "" {
		return "", errors.New(payload.ErrMsg)
	}
	return payload.URLLink, nil
}

func (s *Server) wechatWxaCode(ctx context.Context, accessToken string, page string, scene string) ([]byte, string, error) {
	body := map[string]interface{}{
		"scene":       scene,
		"page":        page,
		"check_path":  false,
		"env_version": s.wechatMiniProgramEnvVersion(),
	}
	data, contentType, err := s.wechatPostBinary(ctx, "/wxa/getwxacodeunlimit", accessToken, body)
	if err != nil {
		return nil, "", err
	}
	if strings.Contains(contentType, "json") {
		var payload struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, "", err
		}
		return nil, "", errors.New(payload.ErrMsg)
	}
	if contentType == "" {
		contentType = "image/png"
	}
	return data, contentType, nil
}

func (s *Server) wechatMiniProgramEnvVersion() string {
	switch strings.ToLower(strings.TrimSpace(s.cfg.Wechat.MiniProgramEnvVersion)) {
	case "develop", "trial", "release":
		return strings.ToLower(strings.TrimSpace(s.cfg.Wechat.MiniProgramEnvVersion))
	default:
		return "release"
	}
}

func (s *Server) wechatPostJSON(ctx context.Context, path string, accessToken string, body interface{}, out interface{}) error {
	data, contentType, err := s.wechatPostBinary(ctx, path, accessToken, body)
	if err != nil {
		return err
	}
	if !strings.Contains(contentType, "json") && len(data) == 0 {
		return errors.New("empty wechat response")
	}
	return json.Unmarshal(data, out)
}

func (s *Server) wechatPostBinary(ctx context.Context, path string, accessToken string, body interface{}) ([]byte, string, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	endpoint := "https://api.weixin.qq.com" + path + "?access_token=" + url.QueryEscape(accessToken)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", errors.New(response.Status)
	}
	return data, response.Header.Get("Content-Type"), nil
}

func (s *Server) routeAdminInviteCodePost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/disable"):
		s.requireAdminPermission("invite_code:manage", s.disableAdminInviteCode)(w, r)
	case strings.HasSuffix(r.URL.Path, "/enable"):
		s.requireAdminPermission("invite_code:manage", s.enableAdminInviteCode)(w, r)
	case strings.HasSuffix(r.URL.Path, "/void"):
		s.requireAdminPermission("invite_code:manage", s.voidAdminInviteCode)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) disableAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/admin/invite-codes/"), "/disable")
	code = strings.TrimSpace(strings.Trim(code, "/"))
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "缺少邀请码")
		return
	}
	invite, err := s.auth.AdminDisableInviteCode(code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, inviteAdminErrorMessage(err))
		return
	}
	s.recordOperation(r, "invite_code:disable", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{"code": invite.Code})
	httpx.OK(w, invite)
}

func (s *Server) enableAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	code := inviteCodeFromAdminActionPath(r.URL.Path, "/enable")
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "缺少邀请码")
		return
	}
	invite, err := s.auth.AdminEnableInviteCode(code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, inviteAdminErrorMessage(err))
		return
	}
	s.recordOperation(r, "invite_code:enable", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{"code": invite.Code})
	httpx.OK(w, invite)
}

func (s *Server) voidAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	code := inviteCodeFromAdminActionPath(r.URL.Path, "/void")
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "缺少邀请码")
		return
	}
	invite, err := s.auth.AdminVoidInviteCode(code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, inviteAdminErrorMessage(err))
		return
	}
	s.recordOperation(r, "invite_code:void", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{"code": invite.Code})
	httpx.OK(w, invite)
}

func inviteCodeFromAdminActionPath(path string, suffix string) string {
	code := strings.TrimSuffix(strings.TrimPrefix(path, "/api/admin/invite-codes/"), suffix)
	return strings.TrimSpace(strings.Trim(code, "/"))
}

func inviteAdminErrorMessage(err error) string {
	switch {
	case errors.Is(err, invites.ErrInvalidEntryType):
		return "邀请码入口类型不正确"
	case strings.Contains(err.Error(), "used invite code"):
		return "邀请码已使用，不能修改状态或作废"
	case strings.Contains(err.Error(), "expired"):
		return "邀请码已过期，不能启用"
	case strings.Contains(err.Error(), "not found"):
		return "邀请码不存在"
	default:
		return "邀请码状态不允许此操作"
	}
}

func (s *Server) updateAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/admin/invite-codes/"))
	code = strings.Trim(code, "/")
	if code == "" || strings.Contains(code, "/") {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "缺少邀请码")
		return
	}
	var req struct {
		OwnerUserID int64  `json:"ownerUserId"`
		EntryType   string `json:"entryType"`
		ExpiresAt   string `json:"expiresAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	if req.OwnerUserID <= 0 || !s.userCanOwnAdminInviteCodes(req.OwnerUserID) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请人必须是平台官方账号或已生效的领路人")
		return
	}
	var expiresAt time.Time
	if strings.TrimSpace(req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "有效期格式不正确")
			return
		}
		expiresAt = parsed
	}
	invite, err := s.auth.AdminUpdateUnusedInviteCode(code, req.OwnerUserID, req.EntryType, expiresAt)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, inviteAdminErrorMessage(err))
		return
	}
	s.recordOperation(r, "invite_code:update", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{"code": invite.Code, "ownerUserId": invite.OwnerID})
	httpx.OK(w, invite)
}

func (s *Server) exportAdminInviteCodes(w http.ResponseWriter, r *http.Request) {
	ownerID, _, ok := optionalInt64Query(w, r, "ownerUserId")
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{Status: status, EntryType: strings.TrimSpace(r.URL.Query().Get("entryType")), OwnerID: ownerID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "导出邀请码失败")
		return
	}
	requestedCodes := strings.Split(strings.TrimSpace(r.URL.Query().Get("codes")), ",")
	selected := map[string]bool{}
	for _, code := range requestedCodes {
		if code = strings.TrimSpace(code); code != "" {
			selected[code] = true
		}
	}
	if len(selected) > 0 {
		filtered := make([]invites.InviteCode, 0, len(selected))
		for _, item := range items {
			if selected[item.Code] {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	buffer := bytes.NewBuffer([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(buffer)
	_ = writer.Write([]string{"邀请码编号", "邀请码", "邀请人用户ID", "邀请人昵称", "邀请人手机号", "入口类型", "状态", "使用状态", "使用次数", "使用用户ID", "使用用户昵称", "使用用户手机号", "有效期", "创建时间", "更新时间"})
	for _, item := range items {
		expiresAt := ""
		if !item.ExpiresAt.IsZero() {
			expiresAt = item.ExpiresAt.Format(time.RFC3339)
		}
		_ = writer.Write([]string{strconv.FormatInt(item.ID, 10), item.Code, strconv.FormatInt(item.OwnerID, 10), item.OwnerNickname, item.OwnerPhoneMasked, item.EntryType, item.DisplayStatus, item.UseStatus, strconv.Itoa(item.UsedCount), strconv.FormatInt(item.BoundWechatUserID, 10), item.BoundWechatNickname, item.BoundWechatPhoneMasked, expiresAt, item.CreatedAt.Format(time.RFC3339), item.UpdatedAt.Format(time.RFC3339)})
	}
	writer.Flush()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="invite-codes.csv"`)
	_, _ = w.Write(buffer.Bytes())
	s.recordOperation(r, "invite_code:export", "invite_code", "", map[string]interface{}{"count": len(items), "batch": len(selected) > 0})
}

func (s *Server) adminInviteRelations(w http.ResponseWriter, r *http.Request) {
	inviterID, _, ok := optionalInt64Query(w, r, "inviterUserId")
	if !ok {
		return
	}
	inviteeID, _, ok := optionalInt64Query(w, r, "inviteeUserId")
	if !ok {
		return
	}
	inviteCodeID, _, ok := optionalInt64Query(w, r, "inviteCodeId")
	if !ok {
		return
	}
	items, err := s.auth.AdminInviteRelations(invites.RelationFilter{
		InviterUserID: inviterID,
		InviteeUserID: inviteeID,
		InviteCodeID:  inviteCodeID,
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码绑定关系失败")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) adminInviteQuotaRequests(w http.ResponseWriter, r *http.Request) {
	ownerID, _, ok := optionalInt64Query(w, r, "ownerUserId")
	if !ok {
		return
	}
	items, err := s.auth.InviteQuotaRequests(ownerID, strings.TrimSpace(r.URL.Query().Get("status")))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码加量申请失败")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) routeAdminInviteQuotaRequestPost(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/audit") {
		http.NotFound(w, r)
		return
	}
	id, err := pathID(r.URL.Path, "/api/admin/invite-quota-requests/", "/audit")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "申请编号不正确")
		return
	}
	var req struct {
		Approve   bool   `json:"approve"`
		Reason    string `json:"reason"`
		EntryType string `json:"entryType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	entryType := strings.TrimSpace(req.EntryType)
	if entryType == "" {
		entryType = invites.EntryTypeLink
	}
	if _, valid := invites.ParseEntryType(entryType); !valid {
		// 先校验生成参数，再改变加量申请状态；否则错误的入口类型会把
		// 申请提前标为“已通过”，但实际没有生成任何邀请码。
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "邀请码入口类型不正确")
		return
	}
	status := "rejected"
	if req.Approve {
		status = "approved"
	}
	adminID := parseInt64Header(r, "X-Admin-ID")
	var request invites.QuotaRequest
	items := []invites.InviteCode{}
	if req.Approve {
		pendingRequests, listErr := s.auth.InviteQuotaRequests(0, "pending")
		if listErr != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取加量申请失败")
			return
		}
		ownerUserID := int64(0)
		for _, pending := range pendingRequests {
			if pending.ID == id {
				ownerUserID = pending.OwnerUserID
				break
			}
		}
		if ownerUserID <= 0 {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "加量申请不存在或已处理")
			return
		}
		if !s.userCanOwnAdminInviteCodes(ownerUserID) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "申请人的领路人身份已失效，不能通过加量申请")
			return
		}
		config, err := s.inviteCodeConfigStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请码配置失败，请稍后重试")
			return
		}
		var expiresAt time.Time
		if config.DefaultValidDays > 0 {
			expiresAt = time.Now().AddDate(0, 0, config.DefaultValidDays)
		}
		var approveErr error
		request, items, approveErr = s.auth.ApproveInviteQuotaRequest(id, strings.TrimSpace(req.Reason), adminID, entryType, expiresAt)
		if approveErr != nil {
			if errors.Is(approveErr, invites.ErrInviteCodeExists) {
				httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "邀请码生成冲突，请重试审核")
				return
			}
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "加量申请不存在、已处理或生成失败")
			return
		}
	} else {
		var reviewErr error
		request, reviewErr = s.auth.ReviewInviteQuotaRequest(id, status, strings.TrimSpace(req.Reason), adminID)
		if reviewErr != nil {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "加量申请不存在或已处理")
			return
		}
	}
	s.recordOperation(r, "invite_quota_request:"+status, "invite_quota_request", strconv.FormatInt(request.ID, 10), map[string]interface{}{"ownerUserId": request.OwnerUserID, "quantity": request.Quantity})
	httpx.OK(w, map[string]interface{}{"request": request, "items": items})
}

func (s *Server) adminUserSummary(userID int64) interface{} {
	if userID <= 0 {
		return nil
	}
	user, ok := s.auth.UserByID(userID)
	if !ok {
		return map[string]interface{}{"id": userID, "missing": true}
	}
	return map[string]interface{}{
		"id":             user.ID,
		"openId":         user.OpenID,
		"nickname":       user.Nickname,
		"realnameStatus": user.RealnameStatus,
		"status":         user.Status,
	}
}
