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
		UserID       int64    `json:"userId"`
		UserIDs      []int64  `json:"userIds"`
		PhoneNumbers []string `json:"phoneNumbers"`
		RoleCode     string   `json:"roleCode"`
		Reason       string   `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "身份白名单开通请求参数错误")
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
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "身份类型仅支持行家或领路人")
		return
	}
	ids := append([]int64{}, req.UserIDs...)
	if req.UserID > 0 {
		ids = append(ids, req.UserID)
	}
	phoneInputs := uniqueRoleGrantPhones(req.PhoneNumbers)
	if len(ids)+len(phoneInputs) == 0 || len(ids)+len(phoneInputs) > 500 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请填写 1 至 500 个用户编号或手机号")
		return
	}
	phoneUserIDs, phoneFailures := s.roleGrantUserIDsByPhone(phoneInputs)
	ids = uniquePositiveIDs(append(ids, phoneUserIDs...))
	adminID := parseInt64Header(r, "X-Admin-ID")
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "一期后台白名单开通"
	}
	batchID := fmt.Sprintf("role-grant-%d", time.Now().UnixNano())
	items := make([]map[string]interface{}, 0, len(ids)+len(phoneFailures))
	successCount := 0
	skippedCount := 0
	failedCount := len(phoneFailures)
	items = append(items, phoneFailures...)
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
		verified, verifyErr := s.identity.IsRealnameVerifiedStrict(userID)
		if verifyErr != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，身份开通已中止")
			return
		}
		if !verified {
			item["status"] = "failed"
			item["reason"] = "必须先完成实名认证"
			item["nickname"] = user.Nickname
			failedCount++
			items = append(items, item)
			continue
		}
		snapshot, snapshotErr := s.profiles.RoleSnapshotStrict(userID)
		if snapshotErr != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取当前身份状态失败，身份开通已中止")
			return
		}
		before := snapshot.RoleStatusMap[roleCode]
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
		_, _ = s.createCriticalNotification(w, "role_whitelist_granted", notifications.CreateRequest{
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

func uniqueRoleGrantPhones(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		phone := strings.TrimSpace(value)
		if phone == "" || seen[phone] {
			continue
		}
		seen[phone] = true
		result = append(result, phone)
	}
	return result
}

// roleGrantUserIDsByPhone keeps the import path inside the existing real-name
// store: white-list grants already require real-name approval, so unresolved
// or unverified phone numbers are returned as per-row failures instead of
// silently granting a role to a different account.
func (s *Server) roleGrantUserIDsByPhone(phones []string) ([]int64, []map[string]interface{}) {
	if len(phones) == 0 {
		return nil, nil
	}
	lookup := make(map[string]int64, len(phones))
	for _, record := range s.identity.AllRecords() {
		plain, err := s.identity.RevealRecord(record)
		if err != nil || plain.Phone == "" {
			continue
		}
		lookup[plain.Phone] = record.UserID
	}
	ids := make([]int64, 0, len(phones))
	failures := make([]map[string]interface{}, 0)
	for _, phone := range phones {
		userID := lookup[phone]
		if userID > 0 {
			ids = append(ids, userID)
			continue
		}
		failures = append(failures, map[string]interface{}{
			"status": "failed", "reason": "手机号未找到已实名用户", "phoneMasked": maskRoleGrantPhone(phone),
		})
	}
	return ids, failures
}

func maskRoleGrantPhone(phone string) string {
	if len(phone) != 11 {
		return "***"
	}
	return phone[:3] + "****" + phone[7:]
}
