package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
)

const expertBlueBadgeProfileConfigKey = "expert-blue-badge"

// ExpertBlueBadgeDTO is a manual platform endorsement. It is deliberately
// separate from growth level and credit score.
type ExpertBlueBadgeDTO struct {
	Enabled     bool   `json:"enabled"`
	Label       string `json:"label"`
	RuleText    string `json:"ruleText"`
	ContactText string `json:"contactText"`
}

func defaultExpertBlueBadgeDTO() ExpertBlueBadgeDTO {
	return ExpertBlueBadgeDTO{
		Label:       "蓝标认证",
		RuleText:    "业内 TOP10 可申请，蓝标由平台人工评估点亮，不影响行家等级晋升。",
		ContactText: "人工审核通道：请联系平台运营提交认证材料。",
	}
}

func (s *Server) expertBlueBadgeForUser(userID int64) ExpertBlueBadgeDTO {
	badge := defaultExpertBlueBadgeDTO()
	if userID <= 0 || s.profiles == nil {
		return badge
	}
	roles := s.profiles.RoleSnapshot(userID).RoleStatusMap
	if roles["expert"] != "approved" && roles["expert"] != "active" {
		return badge
	}
	profileConfig := s.profiles.SystemManagementConfig(userID, expertBlueBadgeProfileConfigKey, map[string]interface{}{})
	badge.Enabled, _ = profileConfig["enabled"].(bool)
	return badge
}

func (s *Server) adminUpdateExpertBlueBadge(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "/expert-blue-badge")
	if !ok {
		return
	}
	if _, found := s.auth.UserByID(userID); !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "用户不存在")
		return
	}
	snapshot, snapshotErr := s.profiles.RoleSnapshotStrict(userID)
	if snapshotErr != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取用户角色身份失败，请稍后重试")
		return
	}
	roles := snapshot.RoleStatusMap
	if roles["expert"] != "approved" && roles["expert"] != "active" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "仅已生效行家可设置蓝标认证")
		return
	}
	var req struct {
		Enabled bool   `json:"enabled"`
		Remark  string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "蓝标认证参数错误")
		return
	}
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Remark == "" || len([]rune(req.Remark)) > 200 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请填写不超过 200 字的人工评估说明")
		return
	}
	if _, saved := s.saveProfileConfig(w, userID, expertBlueBadgeProfileConfigKey, map[string]interface{}{
		"enabled":   req.Enabled,
		"remark":    req.Remark,
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"updatedBy": parseInt64Header(r, "X-Admin-ID"),
	}); !saved {
		return
	}
	s.recordOperation(r, "expert_blue_badge:update", "user", strconv.FormatInt(userID, 10), map[string]interface{}{
		"enabled": req.Enabled,
		"remark":  req.Remark,
	})
	httpx.OK(w, map[string]interface{}{"blueBadge": s.expertBlueBadgeForUser(userID)})
}
