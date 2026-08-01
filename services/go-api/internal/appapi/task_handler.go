package appapi

import (
	"net/http"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/tasks"
	"zhw-mini/services/go-api/internal/users"
)

func (s *Server) recordProfileGuideReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	state, err := s.newbieGuideStateStrict(userID, nil)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取新手引导状态失败，请稍后重试")
		return
	}
	if complete, _ := state["profileComplete"].(bool); complete {
		httpx.OK(w, map[string]interface{}{"guide": state})
		return
	}
	if visible, _ := state["profileReminderVisible"].(bool); visible && s.tasks != nil {
		if _, err := s.tasks.RecordProfileReminder(userID, time.Now()); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "记录资料提醒失败")
			return
		}
	}
	state, err = s.newbieGuideStateStrict(userID, nil)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取新手引导状态失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"guide": state})
}

// newbieGuideState returns only user-facing guide state. A task is never
// completed by this state machine; it merely controls how often the user sees
// the profile reminder before it moves to the persistent profile entry.
func (s *Server) newbieGuideState(userID int64, items []map[string]interface{}) map[string]interface{} {
	state, _ := s.newbieGuideStateStrict(userID, items)
	return state
}

func (s *Server) newbieGuideStateStrict(userID int64, items []map[string]interface{}) (map[string]interface{}, error) {
	operationRules, err := s.currentOperationRulesStrict()
	if err != nil {
		return nil, err
	}
	rules := operationRules.NewbieGuide
	user, _ := s.auth.UserByID(userID)
	profileComplete := profileBasicsComplete(user)
	progress := tasks.GuideProgress{UserID: userID}
	if s.tasks != nil {
		var err error
		progress, err = s.tasks.GuideProgressStrict(userID)
		if err != nil {
			return nil, err
		}
	}
	completed := 0
	currentTaskCode := ""
	for _, item := range items {
		done, _ := item["completed"].(bool)
		if done {
			completed++
			continue
		}
		if currentTaskCode == "" {
			currentTaskCode, _ = item["code"].(string)
		}
	}
	pending := len(items) > completed
	canRemind := rules.Enabled && !profileComplete && progress.ProfileReminderCount < rules.ProfileReminderLimit
	if canRemind && !progress.LastProfileReminderAt.IsZero() {
		canRemind = time.Since(progress.LastProfileReminderAt) >= time.Duration(rules.ProfileReminderIntervalHours)*time.Hour
	}
	return map[string]interface{}{
		"taskState":                 guideTaskState(pending, completed),
		"currentTaskCode":           currentTaskCode,
		"profileComplete":           profileComplete,
		"profileReminderVisible":    canRemind,
		"profileReminderCount":      progress.ProfileReminderCount,
		"profileReminderRemaining":  maxGuideInt(0, rules.ProfileReminderLimit-progress.ProfileReminderCount),
		"persistentProfileReminder": !profileComplete && progress.ProfileReminderCount >= rules.ProfileReminderLimit,
	}, nil
}

func profileBasicsComplete(user users.User) bool {
	return strings.TrimSpace(user.Nickname) != "" && (strings.TrimSpace(user.AvatarURL) != "" || user.AvatarFileID > 0)
}

func guideTaskState(pending bool, completed int) string {
	if !pending {
		return ""
	}
	if completed == 0 {
		return "!"
	}
	return "?"
}

func maxGuideInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func (s *Server) completeTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/app/newbie-tasks/"), "/")
	code = strings.TrimSuffix(code, "/complete")
	if code == "" || s.tasks == nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "任务编号无效")
		return
	}
	// 任务必须来自后台启用的任务规则，不能通过任意编码伪造完成记录。
	operationRules, err := s.currentOperationRulesStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取任务规则失败，请稍后重试")
		return
	}
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
	completed, err := s.newbieTaskCompletionMet(userID, rule.Code)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取任务完成状态失败，请稍后重试")
		return
	}
	if !completed {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "当前尚未满足任务完成条件")
		return
	}
	var progress tasks.Progress
	if rule.Category == "daily" {
		progress, _, err = s.tasks.MarkCompletedForDateOnce(userID, code, time.Now())
	} else {
		progress, _, err = s.tasks.MarkCompletedOnce(userID, code)
	}
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "任务完成记录失败")
		return
	}
	profile, err := s.reviews.ProfileStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取成长档案失败，请稍后重试")
		return
	}
	rewardKey := rule.Code
	if rule.Category == "daily" {
		rewardKey += "@" + time.Now().In(appDisplayLocation).Format("2006-01-02")
	}
	if rewardedProfile, rewardErr := s.awardTaskReward(userID, *rule, rewardKey); rewardErr != nil {
		markGrowthPersistenceDegraded(w, "task_complete", userID, 0, rewardErr)
	} else {
		profile = rewardedProfile
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

func (s *Server) newbieTaskCompletionMet(userID int64, code string) (bool, error) {
	record, err := s.identity.StatusStrict(userID)
	if err != nil {
		return false, err
	}
	snapshot, err := s.profiles.RoleSnapshotStrict(userID)
	if err != nil {
		return false, err
	}
	stats, err := s.games.StatsForUserStrict(userID)
	if err != nil {
		return false, err
	}
	applications, err := s.profiles.RoleApplicationsByUserStrict(userID)
	if err != nil {
		return false, err
	}
	hasApprovedRole := hasApprovedNonPlayerRole(snapshot)
	switch code {
	case "complete_identity":
		return record.Status == "verified", nil
	case "apply_role":
		return len(applications) > 0 || hasApprovedRole, nil
	case "join_or_create_game":
		return stats.Participated > 0, nil
	case "complete_game":
		return stats.Completed > 0, nil
	case "daily_join_game":
		now := time.Now().In(appDisplayLocation)
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, appDisplayLocation)
		return s.games.ParticipatedBetweenStrict(userID, start, start.AddDate(0, 0, 1))
	case "submit_review":
		intents, err := s.reviews.MyIntentsStrict(userID)
		return len(intents) > 0, err
	default:
		return false, nil
	}
}

func hasApprovedNonPlayerRole(snapshot profiles.RoleSnapshot) bool {
	for roleCode, status := range snapshot.RoleStatusMap {
		if roleCode != "player" && (status == "active" || status == "approved") {
			return true
		}
	}
	return false
}
