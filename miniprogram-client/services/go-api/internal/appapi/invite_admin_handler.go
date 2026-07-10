package appapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/invites"
)

func (s *Server) adminInviteCodes(w http.ResponseWriter, r *http.Request) {
	ownerID, _, ok := optionalInt64Query(w, r, "ownerUserId")
	if !ok {
		return
	}
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		EntryType: strings.TrimSpace(r.URL.Query().Get("entryType")),
		OwnerID:   ownerID,
	})
	if err != nil {
		if errors.Is(err, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite codes failed")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) createAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code       string `json:"code"`
		EntryType  string `json:"entryType"`
		BatchCount int    `json:"batchCount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if req.BatchCount < 0 || req.BatchCount > 200 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid batch count")
		return
	}
	ownerUserID := parseInt64Header(r, "X-Admin-ID")
	const maxUses = 1
	if req.BatchCount > 1 {
		if strings.TrimSpace(req.Code) != "" {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "batch code must be auto generated")
			return
		}
		items := make([]invites.InviteCode, 0, req.BatchCount)
		for i := 0; i < req.BatchCount; i++ {
			invite, err := s.auth.AdminCreateInviteCode("", ownerUserID, maxUses, req.EntryType)
			if err != nil {
				if errors.Is(err, invites.ErrInvalidEntryType) {
					httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
					return
				}
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "create invite code failed")
				return
			}
			items = append(items, invite)
		}
		s.recordOperation(r, "invite_code:batch_create", "invite_code", "", map[string]interface{}{
			"entryType":  req.EntryType,
			"batchCount": req.BatchCount,
		})
		httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
		return
	}
	invite, err := s.auth.AdminCreateInviteCode(req.Code, ownerUserID, maxUses, req.EntryType)
	if err != nil {
		if errors.Is(err, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "create invite code failed")
		return
	}
	s.recordOperation(r, "invite_code:create", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{
		"code":        invite.Code,
		"entryType":   invite.EntryType,
		"ownerUserId": invite.OwnerID,
	})
	httpx.OK(w, invite)
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invite code required")
		return
	}
	invite, found, err := s.adminInviteByCode(code)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite codes failed")
		return
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invite code not found")
		return
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviteCodeID: invite.ID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite relations failed")
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invite code required")
		return
	}
	invite, found, err := s.adminInviteByCode(code)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite codes failed")
		return
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invite code not found")
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
	if available {
		materials["message"] = "已生成可发给用户的邀请入口。链接可直接复制发送；二维码可下载后放入海报或发给用户扫码。"
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
	scene := "i=" + invite.Code + "&t=" + inviteEntryTypeShort(entryType)
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

func inviteEntryTypeShort(entryType string) string {
	switch entryType {
	case invites.EntryTypePoster:
		return "p"
	case invites.EntryTypeQRCode:
		return "q"
	default:
		return "l"
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
		lines = append(lines, "也可以使用随附的小程序码扫码进入。")
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
	if strings.HasSuffix(r.URL.Path, "/disable") {
		s.requireAdminPermission("invite_code:manage", s.disableAdminInviteCode)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) disableAdminInviteCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/admin/invite-codes/"), "/disable")
	code = strings.TrimSpace(strings.Trim(code, "/"))
	if code == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invite code required")
		return
	}
	invite, err := s.auth.AdminDisableInviteCode(code)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invite code not found")
		return
	}
	s.recordOperation(r, "invite_code:disable", "invite_code", strconv.FormatInt(invite.ID, 10), map[string]interface{}{"code": invite.Code})
	httpx.OK(w, invite)
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
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "list invite relations failed")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
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
