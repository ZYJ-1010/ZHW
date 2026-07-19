package appapi

import (
	"net/http"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/tasks"
)

func (s *Server) completeTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/app/newbie-tasks/"), "/")
	if code == "" || s.tasks == nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "任务编号无效")
		return
	}
	// 任务必须来自后台启用的任务规则，不能通过任意编码伪造完成记录。
	operationRules := s.currentOperationRules()
	var rule *taskRuleDTO
	for index := range operationRules.Tasks.Items {
		candidate := &operationRules.Tasks.Items[index]
		if candidate.Code == code {
			rule = candidate
			break
		}
	}
	if rule == nil || !rule.Enabled {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "任务不存在或已关闭")
		return
	}
	if rule.Category != "newbie" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "一期仅支持新手任务自动完成")
		return
	}
	if !s.newbieTaskCompletionMet(userID, rule.Code) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "当前尚未满足任务完成条件")
		return
	}
	var progress tasks.Progress
	var err error
	alreadyCompleted := s.tasks.CompletedCodes(userID)[code]
	if rule.Category == "daily" {
		alreadyCompleted = s.tasks.CompletedCodesForDate(userID, time.Now())[code]
	}
	if rule.Category == "daily" {
		progress, err = s.tasks.MarkCompletedForDate(userID, code, time.Now())
	} else {
		progress, err = s.tasks.MarkCompleted(userID, code)
	}
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "任务完成记录失败")
		return
	}
	profile := s.reviews.Profile(userID)
	if !alreadyCompleted {
		profile = s.awardNewbieTaskReward(userID, *rule)
	}
	httpx.OK(w, map[string]interface{}{
		"progress":         progress,
		"claimStatus":      "claimed",
		"taskCode":         rule.Code,
		"category":         rule.Category,
		"rewardPoints":     rule.RewardPoints,
		"rewardExperience": rule.RewardExperience,
		"growthProfile":    profile,
	})
}

func (s *Server) newbieTaskCompletionMet(userID int64, code string) bool {
	record := s.identity.Status(userID)
	snapshot := s.profiles.RoleSnapshot(userID)
	stats := s.games.StatsForUser(userID)
	applications := s.profiles.RoleApplicationsByUser(userID)
	hasApprovedRole := false
	for _, status := range snapshot.RoleStatusMap {
		if status == "active" || status == "approved" {
			hasApprovedRole = true
			break
		}
	}
	switch code {
	case "complete_identity":
		return record.Status == "verified"
	case "apply_role":
		return len(applications) > 0 || hasApprovedRole
	case "join_or_create_game":
		return stats.Participated > 0
	case "complete_game":
		return stats.Completed > 0
	case "submit_review":
		intents := s.reviews.MyIntents(userID)
		return len(intents) > 0
	default:
		return false
	}
}
