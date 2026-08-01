package appapi

import (
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
)

// adminPendingCountsAuthorized intentionally accepts any administrator who
// owns at least one queue. The older implementation required analytics access,
// causing audit/game operators to miss the red-dot work reminders that they
// were actually responsible for.
func (s *Server) adminPendingCountsAuthorized(w http.ResponseWriter, r *http.Request) {
	permissions := []string{
		"game:update_status", "identity:update", "role:update", "report:handle", "redemption:manage", "invite_code:manage",
		"feedback:reply", "notification:wechat:view", "admin_user:view",
	}
	for _, permission := range permissions {
		if adminID, ok := s.admins.HasPermission(s.adminToken(r), permission); ok {
			if r.Header.Get("X-Admin-ID") == "" {
				r.Header.Set("X-Admin-ID", strconv.FormatInt(adminID, 10))
			}
			s.adminPendingCounts(w, r)
			return
		}
	}
	httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "缺少后台接口权限")
}

func (s *Server) adminDashboardAuthorized(w http.ResponseWriter, r *http.Request) {
	permissions := []string{
		"analytics:funnel:view", "user:view", "game:read", "identity:read", "role:view", "report:view",
		"redemption:manage", "invite_code:read", "feedback:view", "notification:wechat:view", "admin_user:view",
	}
	for _, permission := range permissions {
		if adminID, ok := s.admins.HasPermission(s.adminToken(r), permission); ok {
			if r.Header.Get("X-Admin-ID") == "" {
				r.Header.Set("X-Admin-ID", strconv.FormatInt(adminID, 10))
			}
			s.adminDashboard(w, r)
			return
		}
	}
	httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "缺少后台接口权限")
}

// adminPendingCounts exposes one source for the admin work-queue red dots.
// Individual modules still own their records; this endpoint only aggregates
// counts and never changes their state.
func (s *Server) adminPendingCounts(w http.ResponseWriter, r *http.Request) {
	counts := map[string]int{}
	canHandle := func(permission string) bool {
		_, ok := s.admins.HasPermission(s.adminToken(r), permission)
		return ok
	}
	seenApplications := map[int64]bool{}
	if canHandle("game:update_status") {
		allGames, err := s.games.ListStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局待办失败，请稍后重试")
			return
		}
		seenCreators := map[int64]bool{}
		for _, game := range allGames {
			if game.Status == "pending_audit" {
				counts["games"]++
			}
			if seenCreators[game.CreatorUserID] {
				continue
			}
			seenCreators[game.CreatorUserID] = true
			applications, listErr := s.games.ApplicationsForCreatorStrict(game.CreatorUserID)
			if listErr != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取报名审核待办失败，请稍后重试")
				return
			}
			for _, application := range applications {
				if application.Status == "pending" && !seenApplications[application.ID] {
					seenApplications[application.ID] = true
					counts["gameApplications"]++
				}
			}
		}
	}
	if canHandle("identity:update") {
		identityRecords, err := s.identity.AllRecordsStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证待办失败，请稍后重试")
			return
		}
		for _, record := range identityRecords {
			if string(record.Status) == "pending" || string(record.Status) == "submitted" {
				counts["identity"]++
			}
		}
		enterpriseItems, err := s.profiles.AllEnterpriseCertificationsStrict("pending")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取企业认证待办失败，请稍后重试")
			return
		}
		for _, item := range enterpriseItems {
			if item.Status == "pending" {
				counts["enterprise"]++
			}
		}
		profileConfigs, err := s.profiles.SystemManagementConfigsStrict("profile-info")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取头像审核待办失败，请稍后重试")
			return
		}
		for _, config := range profileConfigs {
			personal, ok := objectField(config.Value, "personalInfo")
			if !ok {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(stringField(personal, "avatarAuditStatus")), "pending") || parseFlexibleInt64(personal["pendingAvatarFileId"]) > 0 {
				counts["avatars"]++
			}
		}
	}
	if canHandle("role:update") {
		roleApplications, err := s.profiles.AllRoleApplicationsStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取身份申请待办失败，请稍后重试")
			return
		}
		for _, application := range roleApplications {
			if application.Status == "pending" {
				counts["roles"]++
			}
		}
	}
	if canHandle("report:handle") {
		reports, err := s.reports.ListStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报待办失败，请稍后重试")
			return
		}
		for _, report := range reports {
			if report.Status == "pending" || report.Status == "assigned" || report.Status == "appealed" || report.Status == "appeal_withdrawn" {
				counts["reports"]++
			}
		}
	}
	if canHandle("redemption:manage") {
		orders, err := s.redemption.AdminOrdersStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取兑换待办失败，请稍后重试")
			return
		}
		for _, order := range orders {
			if strings.EqualFold(order.Status, "pending") || strings.EqualFold(order.Status, "pending_review") {
				counts["redemption"]++
			}
		}
	}
	if canHandle("invite_code:manage") {
		if requests, err := s.auth.InviteQuotaRequests(0, "pending"); err == nil {
			counts["invites"] = len(requests)
		}
	}
	if canHandle("feedback:reply") {
		feedbackConfigs, err := s.profiles.SystemManagementConfigsStrict("feedback-records")
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取反馈待办失败，请稍后重试")
			return
		}
		for _, config := range feedbackConfigs {
			for _, record := range systemFeedbackRecordsFromPayload(config.Value) {
				if status := strings.ToLower(strings.TrimSpace(stringField(record, "statusClass"))); status != "resolved" {
					counts["feedback"]++
				}
			}
		}
	}
	if canHandle("notification:wechat:view") {
		tasks, err := s.notices.WechatTasksStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取订阅消息待办失败，请稍后重试")
			return
		}
		for _, task := range tasks {
			if strings.EqualFold(task.Status, "pending") {
				counts["wechatTasks"]++
			}
		}
	}
	if canHandle("admin_user:view") {
		if applications, err := s.admins.AdminApplications(); err == nil {
			for _, application := range applications {
				if strings.EqualFold(application.Status, "pending") {
					counts["adminApplications"]++
				}
			}
		}
	}
	total := 0
	for _, value := range counts {
		total += value
	}
	httpx.OK(w, map[string]interface{}{"counts": counts, "total": total})
}
