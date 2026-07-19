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
	var progress tasks.Progress
	var err error
	if rule.Category == "daily" {
		progress, err = s.tasks.MarkCompletedForDate(userID, code, time.Now())
	} else {
		progress, err = s.tasks.MarkCompleted(userID, code)
	}
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "任务完成记录失败")
		return
	}
	profile := s.reviews.AwardTaskReward(userID, rule.Code, rule.RewardPoints, rule.RewardExperience)
	httpx.OK(w, map[string]interface{}{
		"progress":         progress,
		"taskCode":         rule.Code,
		"category":         rule.Category,
		"rewardPoints":     rule.RewardPoints,
		"rewardExperience": rule.RewardExperience,
		"growthProfile":    profile,
	})
}
