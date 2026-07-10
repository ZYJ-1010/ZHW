package appapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/identity"
)

type identityService interface {
	Ensure(userID int64) identity.Record
	BindPhone(userID int64, phone string) (identity.Record, error)
	SendSMSCode(userID int64) (identity.SMSDispatchResult, error)
	VerifySMSCode(userID int64, code string) (identity.Record, error)
	VerifyPhone(userID int64, realName string, idCard string) (identity.Record, error)
	StartFaceID(userID int64) (string, error)
	CompleteFaceID(userID int64, token string) (identity.Record, error)
	Status(userID int64) identity.Record
	IsVerified(userID int64) bool
	AllRecords() []identity.Record
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

func (s *Server) sendSMSCode(w http.ResponseWriter, r *http.Request) {
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
	httpx.OK(w, data)
}

func (s *Server) verifySMSCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
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
	record, err := s.identity.VerifyPhone(userID, req.RealName, req.IDCard)
	if err != nil {
		status := http.StatusUnprocessableEntity
		message := "手机号核验失败"
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
	filtered := make([]identity.Record, 0, len(items))
	for _, item := range items {
		if status != "" && string(item.Status) != status {
			continue
		}
		if userID > 0 && item.UserID != userID {
			continue
		}
		filtered = append(filtered, item)
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
	httpx.OK(w, record)
}

func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return s.requireIdentityUser(w, r)
}
