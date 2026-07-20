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
		"game:read", "game:update_status", "identity:read", "identity:update",
		"role:view", "role:update", "report:view", "report:handle", "redemption:manage",
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

// adminPendingCounts exposes one source for the admin work-queue red dots.
// Individual modules still own their records; this endpoint only aggregates
// counts and never changes their state.
func (s *Server) adminPendingCounts(w http.ResponseWriter, r *http.Request) {
	counts := map[string]int{}
	seenApplications := map[int64]bool{}
	for _, game := range s.games.List() {
		if game.Status == "pending_audit" {
			counts["games"]++
		}
		for _, application := range s.games.ApplicationsForCreator(game.CreatorUserID) {
			if application.Status == "pending" && !seenApplications[application.ID] {
				seenApplications[application.ID] = true
				counts["gameApplications"]++
			}
		}
	}
	for _, record := range s.identity.AllRecords() {
		if string(record.Status) == "pending" || string(record.Status) == "submitted" {
			counts["identity"]++
		}
	}
	for _, item := range s.profiles.AllEnterpriseCertifications("pending") {
		if item.Status == "pending" {
			counts["enterprise"]++
		}
	}
	for _, application := range s.profiles.AllRoleApplications() {
		if application.Status == "pending" {
			counts["roles"]++
		}
	}
	for _, report := range s.reports.List() {
		if report.Status == "pending" || report.Status == "assigned" {
			counts["reports"]++
		}
	}
	for _, order := range s.redemption.AdminOrders() {
		if strings.EqualFold(order.Status, "pending") || strings.EqualFold(order.Status, "pending_review") {
			counts["redemption"]++
		}
	}
	for _, config := range s.profiles.SystemManagementConfigs("profile-info") {
		personal, ok := objectField(config.Value, "personalInfo")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(stringField(personal, "avatarAuditStatus")), "pending") || parseFlexibleInt64(personal["pendingAvatarFileId"]) > 0 {
			counts["avatars"]++
		}
	}
	total := 0
	for _, value := range counts {
		total += value
	}
	httpx.OK(w, map[string]interface{}{"counts": counts, "total": total})
}
