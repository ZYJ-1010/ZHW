package appapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/revenue"
)

const operationRulesConfigKey = "operation.rules"

type taskRuleDTO struct {
	Code             string `json:"code"`
	Category         string `json:"category"`
	Title            string `json:"title"`
	Required         int    `json:"required"`
	RewardPoints     int    `json:"rewardPoints"`
	RewardExperience int    `json:"rewardExperience"`
	Enabled          bool   `json:"enabled"`
}

type roleOperationRulesDTO struct {
	ExpertCreatedGames     int  `json:"expertCreatedGames"`
	ExpertCreditScore      int  `json:"expertCreditScore"`
	GuideParticipatedGames int  `json:"guideParticipatedGames"`
	GuideInvitedCompleted  int  `json:"guideInvitedCompleted"`
	GuideCreditScore       int  `json:"guideCreditScore"`
	MembershipRequired     bool `json:"membershipRequired"`
}

type operationRulesDTO struct {
	Tasks struct {
		Items []taskRuleDTO `json:"items"`
	} `json:"tasks"`
	Roles     roleOperationRulesDTO `json:"roles"`
	Condition struct {
		Enabled         bool `json:"enabled"`
		CreditMinScore  int  `json:"creditMinScore"`
		ProfileRequired bool `json:"profileRequired"`
	} `json:"condition"`
	Credit struct {
		InitialScore        int `json:"initialScore"`
		ExcellentThreshold  int `json:"excellentThreshold"`
		RestrictedThreshold int `json:"restrictedThreshold"`
		SuspendedThreshold  int `json:"suspendedThreshold"`
		MaxScore            int `json:"maxScore"`
	} `json:"credit"`
	Game struct {
		MinPlayers       int `json:"minPlayers"`
		MaxPlayers       int `json:"maxPlayers"`
		DailyCreateLimit int `json:"dailyCreateLimit"`
	} `json:"game"`
	Invite struct {
		TimeoutMinutes int  `json:"timeoutMinutes"`
		MaxPerGame     int  `json:"maxPerGame"`
		PlayerEnabled  bool `json:"playerEnabled"`
	} `json:"invite"`
	Revenue struct {
		Enabled             bool     `json:"enabled"`
		AllowedStatuses     []string `json:"allowedStatuses"`
		PendingTimeoutHours int      `json:"pendingTimeoutHours"`
		RefundEnabled       bool     `json:"refundEnabled"`
		DisputeEnabled      bool     `json:"disputeEnabled"`
	} `json:"revenue"`
	State struct {
		AuditTimeoutHours      int  `json:"auditTimeoutHours"`
		RecruitingTimeoutHours int  `json:"recruitingTimeoutHours"`
		ReviewTimeoutHours     int  `json:"reviewTimeoutHours"`
		AutoCloseEnabled       bool `json:"autoCloseEnabled"`
	} `json:"state"`
	Points struct {
		ExpireDays    int  `json:"expireDays"`
		ExpireEnabled bool `json:"expireEnabled"`
	} `json:"points"`
	Messages map[string]string `json:"messages"`
	Map      struct {
		DefaultRadiusMeters int `json:"defaultRadiusMeters"`
		MaxRadiusMeters     int `json:"maxRadiusMeters"`
	} `json:"map"`
}

func defaultOperationRules() operationRulesDTO {
	var config operationRulesDTO
	config.Tasks.Items = []taskRuleDTO{
		{Code: "complete_identity", Category: "newbie", Title: "完成实名认证", Required: 1, Enabled: true},
		{Code: "apply_role", Category: "newbie", Title: "申请行家或领路人", Required: 1, Enabled: true},
		{Code: "join_or_create_game", Category: "newbie", Title: "创建或参与第一局", Required: 1, Enabled: true},
		{Code: "complete_game", Category: "newbie", Title: "完成一局服务", Required: 1, Enabled: true},
		{Code: "submit_review", Category: "newbie", Title: "完成评价", Required: 1, Enabled: true},
		{Code: "daily_join_game", Category: "daily", Title: "今日参与 1 次组局", Required: 1, Enabled: true},
		{Code: "activity_complete_game", Category: "activity", Title: "完成一局并提交评价", Required: 1, Enabled: true},
	}
	config.Roles = roleOperationRulesDTO{ExpertCreatedGames: 5, ExpertCreditScore: 90, GuideParticipatedGames: 3, GuideInvitedCompleted: 1, GuideCreditScore: 80}
	config.Condition.Enabled = true
	config.Condition.CreditMinScore = 80
	config.Credit = struct {
		InitialScore        int `json:"initialScore"`
		ExcellentThreshold  int `json:"excellentThreshold"`
		RestrictedThreshold int `json:"restrictedThreshold"`
		SuspendedThreshold  int `json:"suspendedThreshold"`
		MaxScore            int `json:"maxScore"`
	}{InitialScore: 100, ExcellentThreshold: 95, RestrictedThreshold: 80, SuspendedThreshold: 60, MaxScore: 100}
	config.Game.MinPlayers, config.Game.MaxPlayers, config.Game.DailyCreateLimit = 5, 8, 3
	config.Invite.TimeoutMinutes, config.Invite.MaxPerGame, config.Invite.PlayerEnabled = 1440, 1, false
	config.Revenue.Enabled = false
	config.Revenue.AllowedStatuses = []string{"pending", "processing", "succeeded", "failed", "refunded", "disputed", "reversed"}
	config.Revenue.PendingTimeoutHours, config.Revenue.RefundEnabled, config.Revenue.DisputeEnabled = 24, true, true
	config.State.AuditTimeoutHours, config.State.RecruitingTimeoutHours, config.State.ReviewTimeoutHours, config.State.AutoCloseEnabled = 24, 72, 168, true
	config.Points.ExpireDays, config.Points.ExpireEnabled = 365, false
	config.Messages = map[string]string{"creditRestricted": "信用分低于{score}分将限制部分功能", "creditSuspended": "信用分低于{score}分将暂停服务资格"}
	config.Map.DefaultRadiusMeters, config.Map.MaxRadiusMeters = 5000, 20000
	return config
}

func (s *Server) currentOperationRules() operationRulesDTO {
	config := defaultOperationRules()
	if s.systemConfig != nil {
		var stored operationRulesDTO
		if s.systemConfig.Get(operationRulesConfigKey, &stored) {
			config = normalizeOperationRules(stored)
		}
	}
	return config
}

func (s *Server) phaseOneIncomeSummary(userID int64) revenue.IncomeSummary {
	if !s.currentOperationRules().Revenue.Enabled {
		return revenue.IncomeSummary{UserID: userID}
	}
	return s.revenue.IncomeSummary(userID)
}

func normalizeOperationRules(config operationRulesDTO) operationRulesDTO {
	defaults := defaultOperationRules()
	if len(config.Tasks.Items) == 0 {
		config.Tasks.Items = defaults.Tasks.Items
	}
	if config.Roles.ExpertCreatedGames <= 0 {
		config.Roles.ExpertCreatedGames = defaults.Roles.ExpertCreatedGames
	}
	if config.Roles.ExpertCreditScore <= 0 {
		config.Roles.ExpertCreditScore = defaults.Roles.ExpertCreditScore
	}
	if config.Roles.GuideParticipatedGames <= 0 {
		config.Roles.GuideParticipatedGames = defaults.Roles.GuideParticipatedGames
	}
	if config.Roles.GuideInvitedCompleted <= 0 {
		config.Roles.GuideInvitedCompleted = defaults.Roles.GuideInvitedCompleted
	}
	if config.Roles.GuideCreditScore <= 0 {
		config.Roles.GuideCreditScore = defaults.Roles.GuideCreditScore
	}
	// 一期没有会员购买，角色申请不能被运营配置重新打开会员门槛。
	config.Roles.MembershipRequired = false
	if config.Condition.CreditMinScore <= 0 {
		config.Condition.CreditMinScore = defaults.Condition.CreditMinScore
	}
	config.Credit.InitialScore = 100
	if config.Credit.ExcellentThreshold <= 0 {
		config.Credit.ExcellentThreshold = defaults.Credit.ExcellentThreshold
	}
	if config.Credit.RestrictedThreshold <= 0 {
		config.Credit.RestrictedThreshold = defaults.Credit.RestrictedThreshold
	}
	if config.Credit.SuspendedThreshold <= 0 {
		config.Credit.SuspendedThreshold = defaults.Credit.SuspendedThreshold
	}
	if config.Credit.MaxScore <= 0 {
		config.Credit.MaxScore = defaults.Credit.MaxScore
	}
	if config.Game.MinPlayers <= 0 {
		config.Game.MinPlayers = defaults.Game.MinPlayers
	}
	if config.Game.MaxPlayers < config.Game.MinPlayers {
		config.Game.MaxPlayers = defaults.Game.MaxPlayers
	}
	if config.Game.DailyCreateLimit <= 0 {
		config.Game.DailyCreateLimit = defaults.Game.DailyCreateLimit
	}
	if config.Invite.TimeoutMinutes <= 0 {
		config.Invite.TimeoutMinutes = defaults.Invite.TimeoutMinutes
	}
	if config.Invite.MaxPerGame <= 0 {
		config.Invite.MaxPerGame = defaults.Invite.MaxPerGame
	}
	// 玩家没有邀请能力，邀请入口仅由行家、领路人使用。
	config.Invite.PlayerEnabled = false
	if config.Revenue.PendingTimeoutHours <= 0 {
		config.Revenue.PendingTimeoutHours = defaults.Revenue.PendingTimeoutHours
	}
	if len(config.Revenue.AllowedStatuses) == 0 {
		config.Revenue.AllowedStatuses = defaults.Revenue.AllowedStatuses
	}
	// 一期只保留分润数据结构，任何后台配置都不能启用真实分润。
	config.Revenue.Enabled = false
	if config.State.AuditTimeoutHours <= 0 {
		config.State.AuditTimeoutHours = defaults.State.AuditTimeoutHours
	}
	if config.State.RecruitingTimeoutHours <= 0 {
		config.State.RecruitingTimeoutHours = defaults.State.RecruitingTimeoutHours
	}
	if config.State.ReviewTimeoutHours <= 0 {
		config.State.ReviewTimeoutHours = defaults.State.ReviewTimeoutHours
	}
	if config.Points.ExpireDays <= 0 {
		config.Points.ExpireDays = defaults.Points.ExpireDays
	}
	if config.Map.DefaultRadiusMeters <= 0 {
		config.Map.DefaultRadiusMeters = defaults.Map.DefaultRadiusMeters
	}
	if config.Map.MaxRadiusMeters < config.Map.DefaultRadiusMeters {
		config.Map.MaxRadiusMeters = defaults.Map.MaxRadiusMeters
	}
	if len(config.Messages) == 0 {
		config.Messages = defaults.Messages
	}
	return config
}

func (s *Server) adminOperationRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentOperationRules()})
	case http.MethodPut:
		var req operationRulesDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "运营规则格式错误")
			return
		}
		config := normalizeOperationRules(req)
		if err := validateTaskRules(config.Tasks.Items); err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if config.Game.MinPlayers < 2 || config.Game.MaxPlayers > 100 || config.Game.MinPlayers > config.Game.MaxPlayers {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "组局人数范围无效")
			return
		}
		if config.Credit.RestrictedThreshold < config.Credit.SuspendedThreshold {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "信用分阈值顺序无效")
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(operationRulesConfigKey, config) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存运营规则失败")
			return
		}
		if limiter, ok := s.games.(interface {
			SetPlayerLimits(int, int)
			SetDailyCreateLimit(int)
		}); ok {
			limiter.SetPlayerLimits(config.Game.MinPlayers, config.Game.MaxPlayers)
			limiter.SetDailyCreateLimit(config.Game.DailyCreateLimit)
		}
		s.recordOperation(r, "operation_rules:update", "system_config", operationRulesConfigKey, map[string]interface{}{"taskCount": len(config.Tasks.Items)})
		httpx.OK(w, map[string]interface{}{"config": config})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func validateTaskRules(items []taskRuleDTO) error {
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		if code == "" || len(code) > 100 || seen[code] {
			return fmt.Errorf("任务编码为空、过长或重复")
		}
		if item.Category != "newbie" && item.Category != "daily" && item.Category != "activity" {
			return fmt.Errorf("任务分类仅支持 newbie、daily、activity")
		}
		if strings.TrimSpace(item.Title) == "" || len([]rune(item.Title)) > 80 || item.Required <= 0 || item.RewardPoints < 0 || item.RewardExperience < 0 {
			return fmt.Errorf("任务规则参数无效")
		}
		seen[code] = true
	}
	return nil
}
