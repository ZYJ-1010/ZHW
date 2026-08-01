package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/adminauth"
	"zhw-mini/services/go-api/internal/common/httpx"
)

func (s *Server) adminLogin(w http.ResponseWriter, r *http.Request) {
	var req adminauth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	resp, err := s.admins.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, adminauth.ErrAdminDisabled):
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "管理员已禁用")
		default:
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "账号或密码错误")
		}
		return
	}
	httpx.OK(w, resp)
}

func (s *Server) adminPermissions(w http.ResponseWriter, r *http.Request) {
	tree, err := s.admins.Permissions(bearerToken(r.Header.Get("Authorization")))
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "后台登录已失效")
		return
	}
	httpx.OK(w, tree)
}

func (s *Server) submitAdminApplication(w http.ResponseWriter, r *http.Request) {
	var req adminauth.SubmitAdminApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.admins.SubmitAdminApplication(req)
	if err != nil {
		if errors.Is(err, adminauth.ErrInvalidAdminInput) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "管理员申请参数无效")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "提交管理员申请失败")
		return
	}
	httpx.OK(w, item)
}

func (s *Server) adminAccountUsers(w http.ResponseWriter, r *http.Request) {
	items, err := s.admins.AdminUsers()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "获取管理员账号失败")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func (s *Server) adminApplications(w http.ResponseWriter, r *http.Request) {
	items, err := s.admins.AdminApplications()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "获取管理员申请失败")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func (s *Server) routeAdminApplicationPost(w http.ResponseWriter, r *http.Request) {
	const prefix = "/api/admin/admin-applications/"
	const suffix = "/review"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "管理员申请操作不存在")
		return
	}
	idText := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "管理员申请编号错误")
		return
	}
	var req adminauth.ReviewAdminApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.admins.ReviewAdminApplication(id, parseInt64Header(r, "X-Admin-ID"), req)
	if err != nil {
		writeAdminApplicationReviewError(w, err)
		return
	}
	action := "admin_application:close"
	switch item.Status {
	case "approved":
		action = "admin_application:approve"
	case "rejected":
		action = "admin_application:reject"
	}
	s.recordOperation(r, action, "admin_application", strconv.FormatInt(item.ID, 10), map[string]interface{}{
		"status":              item.Status,
		"reviewRemark":        item.ReviewRemark,
		"linkedAdminUserId":   item.LinkedAdminUserID,
		"linkedAdminUsername": item.LinkedAdminUsername,
		"desiredRole":         item.DesiredRole,
	})
	httpx.OK(w, item)
}

func writeAdminApplicationReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, adminauth.ErrApplicationNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "管理员申请不存在")
	case errors.Is(err, adminauth.ErrApplicationProcessed):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "该管理员申请已处理，请勿重复操作")
	case errors.Is(err, adminauth.ErrAdminExists):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "管理员账号已存在，请更换账号名或关联已有账号")
	case errors.Is(err, adminauth.ErrAdminNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "要关联的管理员账号不存在")
	case errors.Is(err, adminauth.ErrAdminDisabled):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "不能关联已禁用的管理员账号")
	case errors.Is(err, adminauth.ErrAdminSelfMutation):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "审核人不能通过申请给自己追加管理员角色")
	case errors.Is(err, adminauth.ErrInvalidAdminInput):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "审核参数无效；驳回或关闭时必须填写备注")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "处理管理员申请失败")
	}
}

func (s *Server) createAdminAccountUser(w http.ResponseWriter, r *http.Request) {
	var req adminauth.CreateAdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.admins.CreateAdminUser(req)
	if err != nil {
		writeAdminUserMutationError(w, err)
		return
	}
	s.recordOperation(r, "admin_user:create", "admin_user", strconv.FormatInt(item.ID, 10), map[string]interface{}{
		"username": item.Username,
		"roles":    item.Roles,
		"status":   item.Status,
	})
	httpx.OK(w, item)
}

func (s *Server) routeAdminAccountUserPut(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/admin-users/", "")
	if !ok {
		return
	}
	var req adminauth.UpdateAdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.admins.UpdateAdminUserByActor(parseInt64Header(r, "X-Admin-ID"), id, req)
	if err != nil {
		writeAdminUserMutationError(w, err)
		return
	}
	s.recordOperation(r, "admin_user:update", "admin_user", strconv.FormatInt(item.ID, 10), map[string]interface{}{
		"username": item.Username,
		"roles":    item.Roles,
		"status":   item.Status,
	})
	httpx.OK(w, item)
}

func (s *Server) adminRoles(w http.ResponseWriter, r *http.Request) {
	items, err := s.admins.Roles()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "获取管理员角色失败")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func (s *Server) adminPermissionCatalog(w http.ResponseWriter, r *http.Request) {
	items, err := s.admins.PermissionCatalog()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "获取管理员权限失败")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func writeAdminUserMutationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, adminauth.ErrAdminExists):
		httpx.Error(w, http.StatusConflict, httpx.CodeValidationError, "管理员账号已存在")
	case errors.Is(err, adminauth.ErrAdminNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "管理员不存在")
	case errors.Is(err, adminauth.ErrAdminSelfMutation):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "不能修改当前登录账号自身的状态或角色")
	case errors.Is(err, adminauth.ErrLastSuperAdmin):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "必须至少保留一名正常状态的超级管理员")
	case errors.Is(err, adminauth.ErrAdminDisabled):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "禁用的管理员账号不能用于此操作")
	case errors.Is(err, adminauth.ErrInvalidAdminInput):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "管理员参数无效")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "管理员操作失败")
	}
}

func (s *Server) adminToken(r *http.Request) string {
	return bearerToken(r.Header.Get("Authorization"))
}
