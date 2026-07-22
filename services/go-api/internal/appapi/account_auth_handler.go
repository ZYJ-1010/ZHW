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
	httpx.OK(w, appLoginResponse{Token: resp.Token, ExpiresAt: resp.ExpiresAt, User: s.buildCurrentUserDTO(resp.User, s.identity.Status(resp.User.ID)), NeedProfile: resp.NeedProfile, IdentityBindStatus: resp.IdentityBindStatus, BoundWechat: resp.BoundWechat, AuthPageMode: resp.AuthPageMode})
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
	httpx.OK(w, s.buildCurrentUserDTO(user, s.identity.Status(userID)))
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
	httpx.OK(w, s.buildCurrentUserDTO(user, s.identity.Status(userID)))
}
