package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

func (s *Server) adminAccountUsers(w http.ResponseWriter, r *http.Request) {
	items, err := s.admins.AdminUsers()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list admin users failed")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
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
	item, err := s.admins.UpdateAdminUser(id, req)
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
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list admin roles failed")
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
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list admin permissions failed")
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
	case errors.Is(err, adminauth.ErrInvalidAdminInput):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "管理员参数无效")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "管理员操作失败")
	}
}

func (s *Server) adminToken(r *http.Request) string {
	return bearerToken(r.Header.Get("Authorization"))
}
