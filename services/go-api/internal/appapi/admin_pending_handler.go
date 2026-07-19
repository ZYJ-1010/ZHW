package appapi

import (
	"net/http"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
)

// adminPendingCounts exposes one source for the admin work-queue red dots.
// Individual modules still own their records; this endpoint only aggregates
// counts and never changes their state.
func (s *Server) adminPendingCounts(w http.ResponseWriter, r *http.Request) {
	counts := map[string]int{}
	for _, game := range s.games.List() {
		if game.Status == "pending_audit" {
			counts["games"]++
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
