package appapi

import (
	"encoding/json"
	"net/http"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/reviews"
)

const growthRewardRulesConfigKey = "growth.reward_rules"

func (s *Server) adminGrowthRewardRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.reviews.GrowthRules()})
	case http.MethodPut:
		var req reviews.GrowthRules
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "成长积分规则格式错误")
			return
		}
		config, err := normalizeGrowthRewardRules(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig == nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "系统配置服务不可用")
			return
		}
		if err := s.systemConfig.Set(growthRewardRulesConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存成长积分规则失败")
			return
		}
		s.recordOperation(r, "growth_reward_rules:update", "system_config", growthRewardRulesConfigKey, map[string]interface{}{
			"completedGameExperience":   config.CompletedGameExperience,
			"submittedReviewExperience": config.SubmittedReviewExperience,
			"receivedReviewExperience":  config.ReceivedReviewExperience,
			"experiencePerLevel":        config.ExperiencePerLevel,
		})
		httpx.OK(w, map[string]interface{}{"config": config})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func normalizeGrowthRewardRules(config reviews.GrowthRules) (reviews.GrowthRules, error) {
	defaults := reviews.DefaultGrowthRules()
	if config.CompletedGameExperience <= 0 || config.CompletedGameExperience > 10000 {
		return reviews.GrowthRules{}, &validationError{"完成组局经验必须在 1-10000 之间"}
	}
	if config.SubmittedReviewExperience <= 0 || config.SubmittedReviewExperience > 10000 {
		return reviews.GrowthRules{}, &validationError{"提交评价经验必须在 1-10000 之间"}
	}
	if config.ReceivedReviewExperience <= 0 || config.ReceivedReviewExperience > 10000 {
		return reviews.GrowthRules{}, &validationError{"收到评价经验必须在 1-10000 之间"}
	}
	if config.ExperiencePerLevel <= 0 || config.ExperiencePerLevel > 1000000 {
		return reviews.GrowthRules{}, &validationError{"升级所需经验必须在 1-1000000 之间"}
	}
	if config.InitialLevel <= 0 {
		config.InitialLevel = defaults.InitialLevel
	}
	config.InitialCreditScore = defaults.InitialCreditScore
	if config.CreditScoreCap <= 0 {
		config.CreditScoreCap = defaults.CreditScoreCap
	}
	if config.InitialCreditScore > config.CreditScoreCap {
		return reviews.GrowthRules{}, &validationError{"信用初始分不能高于信用分上限"}
	}
	return config, nil
}

type validationError struct{ message string }

func (e *validationError) Error() string { return e.message }
