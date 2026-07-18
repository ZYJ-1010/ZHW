package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/profiles"
)

func (s *Server) appEnterpriseCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	item, found := s.profiles.EnterpriseCertification(userID)
	if !found {
		httpx.OK(w, map[string]interface{}{"status": "none", "items": []interface{}{}})
		return
	}
	httpx.OK(w, item)
}

func (s *Server) submitEnterpriseCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req profiles.SubmitEnterpriseCertificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if err := s.validateEnterpriseFiles(userID, req.BusinessLicenseFileID, req.PublicAccountFileID); err != nil {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "企业认证材料无效")
		return
	}
	item, err := s.profiles.SubmitEnterpriseCertification(userID, req)
	if err != nil {
		if errors.Is(err, profiles.ErrEnterpriseCertificationPending) {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "已有企业认证申请正在审核")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "企业认证信息不完整或格式错误")
		return
	}
	s.recordBehavior(userID, "submit_enterprise_certification", "enterprise_certification", item.ID, nil)
	httpx.OK(w, item)
}

func (s *Server) validateEnterpriseFiles(userID int64, businessLicenseFileID int64, publicAccountFileID int64) error {
	if businessLicenseFileID <= 0 || publicAccountFileID <= 0 || businessLicenseFileID == publicAccountFileID {
		return files.ErrFileNotFound
	}
	for _, fileID := range []int64{businessLicenseFileID, publicAccountFileID} {
		item, err := s.files.Get(fileID)
		if err != nil || item.UploaderID != userID || item.BizType != "enterprise_material" {
			return files.ErrFileNotFound
		}
	}
	return nil
}

func (s *Server) adminEnterpriseCertifications(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	items := s.profiles.AllEnterpriseCertifications(status)
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items)})
}

func (s *Server) adminEnterpriseCertificationDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/enterprise-certifications/", "")
	if !ok {
		return
	}
	item, found := s.profiles.EnterpriseCertification(userID)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "企业认证记录不存在")
		return
	}
	httpx.OK(w, item)
}

func (s *Server) routeAdminEnterpriseCertificationPost(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/review") {
		http.NotFound(w, r)
		return
	}
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/enterprise-certifications/", "/review")
	if !ok {
		return
	}
	var req profiles.ReviewEnterpriseCertificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	adminID := parseInt64Header(r, "X-Admin-ID")
	item, err := s.profiles.ReviewEnterpriseCertification(adminID, userID, req)
	if err != nil {
		if errors.Is(err, profiles.ErrEnterpriseCertificationNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "企业认证记录不存在")
			return
		}
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "企业认证审核参数错误")
		return
	}
	s.notices.Create(notifications.CreateRequest{
		UserID: userID, NotifyType: "enterprise_certification_reviewed", Title: "企业认证审核结果",
		Content: map[bool]string{true: "企业认证已通过", false: "企业认证未通过：" + item.RejectReason}[item.Status == "approved"],
		BizType: "enterprise_certification", BizID: item.ID,
	})
	s.recordOperation(r, "enterprise_certification:review", "enterprise_certification", strconv.FormatInt(item.ID, 10), map[string]interface{}{
		"userId": userID, "status": item.Status, "remark": item.ReviewRemark,
	})
	httpx.OK(w, item)
}
