package appapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) passwordLogin(w http.ResponseWriter, r *http.Request) {
	var req auth.PasswordLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	resp, err := s.auth.PasswordLogin(req)
	if err != nil {
		message := "手机号或密码错误"
		if errors.Is(err, auth.ErrPasswordNotSet) {
			message = "该账号尚未设置密码，请使用短信验证码登录"
		}
		if errors.Is(err, auth.ErrPhoneInvalid) {
			message = "手机号格式不正确"
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, message)
		return
	}
	record, err := s.identity.StatusStrict(resp.User.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，请稍后重试")
		return
	}
	current, err := s.buildCurrentUserDTOStrict(resp.User, record)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取登录用户资料失败，请稍后重试")
		return
	}
	httpx.OK(w, appLoginResponse{Token: resp.Token, ExpiresAt: resp.ExpiresAt, User: current, NeedProfile: resp.NeedProfile, IdentityBindStatus: resp.IdentityBindStatus, BoundWechat: resp.BoundWechat, AuthPageMode: resp.AuthPageMode})
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone    string `json:"phone"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	user, err := s.auth.ResetPasswordByPhone(req.Phone, req.Code, req.Password)
	if err != nil {
		message := "密码重置失败，请稍后重试"
		switch {
		case errors.Is(err, auth.ErrPhoneInvalid):
			message = "手机号格式不正确"
		case errors.Is(err, auth.ErrPhoneCodeInvalid):
			message = "短信验证码错误或已失效"
		case errors.Is(err, auth.ErrPasswordInvalid):
			message = "账号不存在，或密码须为 8-64 位字母和数字"
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, message)
		return
	}
	s.recordBehavior(user.ID, "reset_login_password", "user", user.ID, nil)
	record, err := s.identity.StatusStrict(user.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，请稍后重试")
		return
	}
	current, err := s.buildCurrentUserDTOStrict(user, record)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取用户资料失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"user": current})
}

func (s *Server) bindWechatAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
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
	user, err := s.auth.BindWechat(userID, req.Code)
	if err != nil {
		message := "微信绑定失败，请重试"
		if errors.Is(err, auth.ErrWechatAlreadyBound) {
			message = "该微信已绑定其他账号"
		}
		if errors.Is(err, auth.ErrWechatCodeInvalid) {
			message = "微信授权已失效，请重试"
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, message)
		return
	}
	record, err := s.identity.StatusStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，请稍后重试")
		return
	}
	current, err := s.buildCurrentUserDTOStrict(user, record)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取用户资料失败，请稍后重试")
		return
	}
	httpx.OK(w, current)
}

func (s *Server) setLoginPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	user, err := s.auth.SetPassword(userID, req.Password)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "密码须为 8-64 位字母和数字")
		return
	}
	record, err := s.identity.StatusStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，请稍后重试")
		return
	}
	current, err := s.buildCurrentUserDTOStrict(user, record)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取用户资料失败，请稍后重试")
		return
	}
	httpx.OK(w, current)
}
