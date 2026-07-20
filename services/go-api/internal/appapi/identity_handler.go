package appapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/notifications"
)

type identityService interface {
	Ensure(userID int64) identity.Record
	BindPhone(userID int64, phone string) (identity.Record, error)
	RestartRealname(userID int64, phone string) (identity.Record, error)
	SendSMSCode(userID int64) (identity.SMSDispatchResult, error)
	VerifySMSCode(userID int64, code string) (identity.Record, error)
	MarkSMSVerified(userID int64) (identity.Record, error)
	VerifyPhone(userID int64, realName string, idCard string) (identity.Record, error)
	SubmitManualRealname(userID int64, realName string, idCard string) (identity.Record, error)
	ReviewManualRealname(userID int64, approve bool, reason string) (identity.Record, error)
	StartFaceID(userID int64) (string, error)
	CompleteFaceID(userID int64, token string) (identity.Record, error)
	Status(userID int64) identity.Record
	InGameIdentity(userID int64) (identity.InGameIdentity, bool)
	IsVerified(userID int64) bool
	AllRecords() []identity.Record
	RevealRecord(record identity.Record) (identity.PlainIdentity, error)
}

func (s *Server) bindPhone(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if _, err := s.auth.BindPhoneAuth(userID, req.Phone); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "手机号已被使用或格式错误")
		return
	}
	record, err := s.identity.BindPhone(userID, req.Phone)
	if err != nil {
		if errors.Is(err, identity.ErrPhoneInvalid) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid phone")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "手机号不能为空")
		return
	}
	httpx.OK(w, record)
}

func (s *Server) restartRealname(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Code          string `json:"code"`
		Phone         string `json:"phone"`
		EncryptedData string `json:"encryptedData"`
		IV            string `json:"iv"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		var err error
		phone, err = s.wechatPhoneNumber(r.Context(), req.Code)
		if err != nil {
			httpx.Error(w, http.StatusBadGateway, httpx.CodeSystemError, "获取微信手机号失败")
			return
		}
	}
	if _, err := s.auth.BindPhoneAuth(userID, phone); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "手机号已被使用或格式错误")
		return
	}
	record, err := s.identity.RestartRealname(userID, phone)
	if err != nil {
		if errors.Is(err, identity.ErrPhoneInvalid) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid phone")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "手机号不能为空")
		return
	}
	if _, err := s.auth.UpdateRealnameStatus(userID, string(record.Status)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "sync user realname status failed")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"status":      record.Status,
		"phoneMasked": record.PhoneMasked,
		"next":        "/pages/login/realname/index?mode=reverify",
	})
}

func (s *Server) wechatPhoneNumber(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" || strings.TrimSpace(s.cfg.Wechat.AppID) == "" || strings.TrimSpace(s.cfg.Wechat.AppSecret) == "" {
		return "", errors.New("wechat phone code unavailable")
	}
	token, err := s.wechatAccessToken(ctx)
	if err != nil {
		return "", err
	}
	var payload struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
			CountryCode     string `json:"countryCode"`
		} `json:"phone_info"`
	}
	body := map[string]string{"code": code}
	if err := s.wechatPostJSON(ctx, "/wxa/business/getuserphonenumber", token, body, &payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		if strings.TrimSpace(payload.ErrMsg) != "" {
			return "", errors.New(payload.ErrMsg)
		}
		return "", errors.New("wechat phone number failed")
	}
	phone := strings.TrimSpace(payload.PhoneInfo.PurePhoneNumber)
	if phone == "" {
		phone = strings.TrimSpace(payload.PhoneInfo.PhoneNumber)
	}
	if phone == "" {
		return "", errors.New("wechat phone number empty")
	}
	return phone, nil
}

func (s *Server) sendSMSCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Scene string `json:"scene"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if strings.TrimSpace(req.Phone) != "" {
		result, err := s.auth.SendPhoneCode(r.Context(), req.Phone, req.Scene)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrPhoneInvalid):
				httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid phone")
			case errors.Is(err, auth.ErrPhoneCodeRateLimit), errors.Is(err, auth.ErrPhoneCodeDailyLimit):
				httpx.Error(w, http.StatusTooManyRequests, httpx.CodeValidationError, "sms code send too frequently")
			default:
				httpx.Error(w, http.StatusBadGateway, httpx.CodeSystemError, "sms code send failed")
			}
			return
		}
		httpx.OK(w, smsDispatchPayload(result))
		return
	}
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	result, err := s.identity.SendSMSCode(userID)
	if err != nil {
		if errors.Is(err, identity.ErrSMSRateLimited) {
			httpx.Error(w, http.StatusTooManyRequests, httpx.CodeValidationError, "sms code send too frequently")
			return
		}
		if errors.Is(err, identity.ErrSMSDailyLimited) {
			httpx.Error(w, http.StatusTooManyRequests, httpx.CodeValidationError, "sms code daily limit exceeded")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "sms code send failed")
		return
	}
	httpx.OK(w, smsDispatchPayload(result))
}

func smsDispatchPayload(result identity.SMSDispatchResult) map[string]string {
	data := map[string]string{
		"provider": result.Provider,
		"message":  "sms code sent",
	}
	if result.MessageID != "" {
		data["messageId"] = result.MessageID
	}
	if result.MockCode != "" {
		data["mockCode"] = result.MockCode
		data["message"] = "本地环境固定验证码为 000000"
	}
	return data
}

func (s *Server) verifySMSCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if strings.TrimSpace(req.Phone) != "" {
		if err := s.auth.VerifyPhoneCode(req.Phone, req.Code); err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "短信验证码错误")
			return
		}
		httpx.OK(w, map[string]string{"message": "sms code verified"})
		return
	}
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	record, err := s.identity.VerifySMSCode(userID, req.Code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "短信验证码错误")
		return
	}
	httpx.OK(w, record)
}

func (s *Server) verifyPhone(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		RealName string `json:"realName"`
		IDCard   string `json:"idCard"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	record, err := s.identity.SubmitManualRealname(userID, req.RealName, req.IDCard)
	if err != nil {
		status := http.StatusUnprocessableEntity
		message := "实名认证提交失败"
		if errors.Is(err, identity.ErrCodeInvalid) {
			message = "请先完成短信验证码校验"
		}
		if errors.Is(err, identity.ErrRealnameRequired) {
			message = "姓名和身份证号不能为空"
		}
		if errors.Is(err, identity.ErrIDCardInvalid) {
			message = "invalid id card"
		}
		httpx.Error(w, status, httpx.CodeValidationError, message)
		return
	}
	if _, err := s.auth.UpdateRealnameStatus(userID, string(record.Status)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "sync user realname status failed")
		return
	}
	httpx.OK(w, record)
}

func (s *Server) startFaceID(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	token, err := s.identity.StartFaceID(userID)
	if err != nil {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "请先完成手机号核验")
		return
	}
	httpx.OK(w, map[string]string{"faceToken": token})
}

func (s *Server) faceIDCallback(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if !s.verifyFaceIDCallbackSignature(r, payload) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "invalid faceid callback signature")
		return
	}
	var req struct {
		FaceToken string `json:"faceToken"`
	}
	if err := json.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	record, err := s.identity.CompleteFaceID(userID, req.FaceToken)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "人脸核身流程未开始或 token 不匹配")
		return
	}
	httpx.OK(w, record)
}

func (s *Server) identityStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, s.identity.Status(userID))
}

func (s *Server) verifyFaceIDCallbackSignature(r *http.Request, payload []byte) bool {
	if !s.faceIDCallbackRequireSignature && s.faceIDCallbackSecret == "" {
		return true
	}
	if s.faceIDCallbackSecret == "" {
		return false
	}
	signature := strings.TrimSpace(r.Header.Get("X-FaceID-Signature"))
	if signature == "" {
		signature = strings.TrimSpace(r.Header.Get("X-Tencent-FaceID-Signature"))
	}
	signature = strings.TrimPrefix(strings.ToLower(signature), "sha256=")
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.faceIDCallbackSecret))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

func (s *Server) requireIdentityUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "未登录")
		return 0, false
	}
	user, ok := s.auth.CurrentUser(token)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "登录已失效")
		return 0, false
	}
	return user.ID, true
}

func (s *Server) issueTokenAfterIdentity(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	if !s.identity.IsVerified(userID) {
		httpx.Error(w, http.StatusForbidden, 40341, "strong identity required")
		return
	}
	session, err := s.auth.IssueAppToken(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "issue token failed")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"token":              session.Token,
		"expiresAt":          session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"identityBindStatus": string(s.identity.Status(userID).Status),
	})
}

func (s *Server) adminIdentityVerifications(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	userID := parseInt64Query(r, "userId")
	items := s.identity.AllRecords()
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if status != "" && string(item.Status) != status {
			continue
		}
		if userID > 0 && item.UserID != userID {
			continue
		}
		filtered = append(filtered, s.adminIdentityPayloadForRequest(r, item))
	}
	httpx.OK(w, map[string]interface{}{"items": filtered})
}

func (s *Server) adminIdentityVerificationDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/identity-verifications/", "")
	if !ok {
		return
	}
	record := s.identity.Status(userID)
	s.recordOperation(r, "identity:verification:view", "identity_verification", strconv.FormatInt(userID, 10), map[string]interface{}{
		"status": record.Status,
	})
	httpx.OK(w, s.adminIdentityPayloadForRequest(r, record))
}

func (s *Server) routeAdminIdentityVerificationPost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/review") {
		s.requireAdminPermission("identity:update", s.reviewIdentityVerification)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) reviewIdentityVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/identity-verifications/", "/review")
	if !ok {
		return
	}
	var req struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
		Remark  string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	reason := strings.TrimSpace(firstNonEmpty(req.Reason, req.Remark))
	record, err := s.identity.ReviewManualRealname(userID, req.Approve, reason)
	if err != nil {
		switch {
		case errors.Is(err, identity.ErrRecordNotFound):
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "identity record not found")
		case errors.Is(err, identity.ErrRealnameRequired):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "identity material missing")
		case errors.Is(err, identity.ErrReviewReasonRequired):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "驳回审核必须填写原因")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "review identity failed")
		}
		return
	}
	if _, err := s.auth.UpdateRealnameStatus(userID, string(record.Status)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "sync user realname status failed")
		return
	}
	status := "rejected"
	title := "实名认证未通过"
	content := "你的实名认证未通过，请核对姓名和身份证号后重新提交。"
	if req.Approve {
		status = "verified"
		title = "实名认证已通过"
		content = "你的实名认证已通过，可以继续使用平台身份相关功能。"
	}
	if !req.Approve && reason != "" {
		content += " 原因：" + reason
	}
	s.notices.Create(notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "identity_review",
		Title:      title,
		Content:    content,
		BizType:    "identity_verification",
		BizID:      userID,
	})
	s.recordOperation(r, "identity:verification:review", "identity_verification", strconv.FormatInt(userID, 10), map[string]interface{}{
		"status": status,
		"reason": reason,
	})
	httpx.OK(w, s.adminIdentityPayloadForRequest(r, record))
}

func (s *Server) adminIdentityPayloadForRequest(r *http.Request, record identity.Record) map[string]interface{} {
	_, allowSensitive := s.admins.HasPermission(s.adminToken(r), "identity:sensitive:read")
	return s.adminIdentityPayload(record, allowSensitive)
}

func (s *Server) adminIdentityPayload(record identity.Record, allowSensitive bool) map[string]interface{} {
	payload := map[string]interface{}{
		"userId":                    record.UserID,
		"status":                    record.Status,
		"phoneMasked":               record.PhoneMasked,
		"realNameMasked":            record.RealNameMasked,
		"idCardMasked":              record.IDCardMasked,
		"smsVerified":               record.SMSVerified,
		"phoneVerified":             record.PhoneVerified,
		"faceVerified":              record.FaceVerified,
		"wechatRealnameConsistency": record.WechatRealnameConsistency,
		"failureReason":             record.FailureReason,
		"updatedAt":                 record.UpdatedAt,
		"idCardFullAvailable":       false,
		"phoneFullAvailable":        false,
		"realNameFullAvailable":     false,
	}
	plain, err := s.identity.RevealRecord(record)
	if allowSensitive && err == nil {
		if strings.TrimSpace(plain.Phone) != "" {
			payload["phone"] = plain.Phone
			payload["phoneFull"] = plain.Phone
			payload["phoneFullAvailable"] = true
		}
		if strings.TrimSpace(plain.RealName) != "" {
			payload["realNameFull"] = plain.RealName
			payload["realNameFullAvailable"] = true
		}
		if strings.TrimSpace(plain.IDCard) != "" {
			payload["idCardFull"] = plain.IDCard
			payload["idCardFullAvailable"] = true
		}
	}
	return payload
}

func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return s.requireIdentityUser(w, r)
}
