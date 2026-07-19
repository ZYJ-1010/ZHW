package appapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/notifications"
)

// adminGrantRole supports the一期白名单/导入场景。调用方可以提交单个 userId
// 或 userIds 批量赋予已审核的行家、领路人身份。
func (s *Server) adminGrantRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   int64   `json:"userId"`
		UserIDs  []int64 `json:"userIds"`
		RoleCode string  `json:"roleCode"`
		Reason   string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid role grant request")
		return
	}
	roleCode := strings.ToLower(strings.TrimSpace(req.RoleCode))
	if roleCode == "master" || roleCode == "行家" {
		roleCode = "expert"
	}
	if roleCode == "leader" || roleCode == "领路人" {
		roleCode = "guide"
	}
	if roleCode != "expert" && roleCode != "guide" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "roleCode must be expert or guide")
		return
	}
	ids := append([]int64{}, req.UserIDs...)
	if req.UserID > 0 {
		ids = append(ids, req.UserID)
	}
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 || len(ids) > 500 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "userIds required")
		return
	}
	adminID := parseInt64Header(r, "X-Admin-ID")
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "一期后台白名单开通"
	}
	batchID := fmt.Sprintf("role-grant-%d", time.Now().UnixNano())
	items := make([]map[string]interface{}, 0, len(ids))
	successCount := 0
	skippedCount := 0
	failedCount := 0
	for _, userID := range ids {
		item := map[string]interface{}{"userId": userID, "roleCode": roleCode, "batchId": batchID}
		user, found := s.auth.UserByID(userID)
		if !found {
			item["status"] = "failed"
			item["reason"] = "用户不存在"
			failedCount++
			items = append(items, item)
			continue
		}
		if !s.identity.IsVerified(userID) {
			item["status"] = "failed"
			item["reason"] = "必须先完成实名认证"
			item["nickname"] = user.Nickname
			failedCount++
			items = append(items, item)
			continue
		}
		before := s.profiles.RoleSnapshot(userID).RoleStatusMap[roleCode]
		if err := s.profiles.GrantRoleWhitelist(userID, roleCode, adminID, reason); err != nil {
			item["status"] = "failed"
			item["reason"] = "身份开通失败"
			item["nickname"] = user.Nickname
			failedCount++
			items = append(items, item)
			continue
		}
		item["nickname"] = user.Nickname
		item["status"] = "approved"
		item["source"] = "admin_whitelist"
		item["reason"] = reason
		if before == "approved" || before == "active" {
			item["status"] = "already_active"
			skippedCount++
		} else {
			successCount++
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     userID,
			NotifyType: "role_granted",
			Title:      "身份已开通",
			Content:    "后台已为你开通" + map[string]string{"expert": "行家", "guide": "领路人"}[roleCode] + "身份。",
			BizType:    "role",
			BizID:      userID,
		})
		s.recordOperation(r, "role:grant", "user", strconv.FormatInt(userID, 10), map[string]interface{}{"roleCode": roleCode, "source": "admin_whitelist", "reason": reason, "batchId": batchID})
		items = append(items, item)
	}
	httpx.OK(w, map[string]interface{}{"items": items, "total": len(items), "success": successCount, "alreadyActive": skippedCount, "failed": failedCount, "batchId": batchID})
}
