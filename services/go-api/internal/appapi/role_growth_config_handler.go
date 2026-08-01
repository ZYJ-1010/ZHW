package appapi

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
)

const (
	roleLevelConfigKey         = "growth.role_level_config"
	roleMetricConfigKey        = "growth.role_metric_config"
	creditRestrictionConfigKey = "credit.restriction_config"
	roleGrowthRecalcStatusKey  = "growth.recalculation_status"
)

// roleLevelRuleDTO is deliberately role-specific: player ranks are based on
// experience, while expert and guide ranks are based on a normalized score.
// Keeping both thresholds in the rule lets operations maintain all three in
// one clear, versioned configuration screen.
type roleLevelRuleDTO struct {
	RoleCode                   string `json:"roleCode"`
	LevelCode                  string `json:"levelCode"`
	LevelNo                    int    `json:"levelNo"`
	Title                      string `json:"title"`
	MinExperience              int    `json:"minExperience"`
	MinScore                   int    `json:"minScore"`
	MinDirectCompletedInvitees int    `json:"minDirectCompletedInvitees"`
	Enabled                    bool   `json:"enabled"`
	SortOrder                  int    `json:"sortOrder"`
}

type roleLevelConfigDTO struct {
	Items   []roleLevelRuleDTO `json:"items"`
	Version string             `json:"version"`
}

type roleMetricRuleDTO struct {
	RoleCode    string `json:"roleCode"`
	MetricCode  string `json:"metricCode"`
	MetricGroup string `json:"metricGroup"`
	Title       string `json:"title"`
	MetricType  string `json:"metricType"`
	RawWeight   int    `json:"rawWeight"`
	TargetValue int    `json:"targetValue"`
	WindowDays  int    `json:"windowDays"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int    `json:"sortOrder"`
}

type roleMetricConfigDTO struct {
	Items   []roleMetricRuleDTO `json:"items"`
	Version string              `json:"version"`
}

type creditRestrictionConfigDTO struct {
	InitialScore          int    `json:"initialScore"`
	ScoreCap              int    `json:"scoreCap"`
	CreateRestrictedBelow int    `json:"createRestrictedBelow"`
	JoinRestrictedBelow   int    `json:"joinRestrictedBelow"`
	FrozenBelow           int    `json:"frozenBelow"`
	Version               string `json:"version"`
}

type roleGrowthRecalculationStatusDTO struct {
	State       string `json:"state"`
	Trigger     string `json:"trigger"`
	CompletedAt string `json:"completedAt"`
	Description string `json:"description"`
}

func defaultRoleLevelConfig() roleLevelConfigDTO {
	return roleLevelConfigDTO{Version: "2026-07-25-v1", Items: []roleLevelRuleDTO{
		{RoleCode: "player", LevelCode: "lv0", LevelNo: 0, Title: "Lv0 新玩家", MinExperience: 0, Enabled: true, SortOrder: 10},
		{RoleCode: "player", LevelCode: "lv1", LevelNo: 1, Title: "Lv1 初入局", MinExperience: 100, Enabled: true, SortOrder: 20},
		{RoleCode: "player", LevelCode: "lv2", LevelNo: 2, Title: "Lv2 稳定参与", MinExperience: 300, Enabled: true, SortOrder: 30},
		{RoleCode: "player", LevelCode: "lv3", LevelNo: 3, Title: "Lv3 深度参与", MinExperience: 600, Enabled: true, SortOrder: 40},
		{RoleCode: "player", LevelCode: "lv4", LevelNo: 4, Title: "Lv4 高活跃", MinExperience: 1000, Enabled: true, SortOrder: 50},
		{RoleCode: "expert", LevelCode: "star0", LevelNo: 0, Title: "0 星行家", MinScore: 0, Enabled: true, SortOrder: 10},
		{RoleCode: "expert", LevelCode: "star1", LevelNo: 1, Title: "1 星行家", MinScore: 20, Enabled: true, SortOrder: 20},
		{RoleCode: "expert", LevelCode: "star2", LevelNo: 2, Title: "2 星行家", MinScore: 40, Enabled: true, SortOrder: 30},
		{RoleCode: "expert", LevelCode: "star3", LevelNo: 3, Title: "3 星行家", MinScore: 60, Enabled: true, SortOrder: 40},
		{RoleCode: "expert", LevelCode: "star4", LevelNo: 4, Title: "4 星行家", MinScore: 80, Enabled: true, SortOrder: 50},
		{RoleCode: "expert", LevelCode: "star5", LevelNo: 5, Title: "5 星行家", MinScore: 90, Enabled: true, SortOrder: 60},
		{RoleCode: "guide", LevelCode: "guide0", LevelNo: 0, Title: "0 级领路人", MinScore: 0, Enabled: true, SortOrder: 10},
		{RoleCode: "guide", LevelCode: "guide1", LevelNo: 1, Title: "1 级领路人", MinScore: 20, MinDirectCompletedInvitees: 10, Enabled: true, SortOrder: 20},
		{RoleCode: "guide", LevelCode: "guide2", LevelNo: 2, Title: "2 级领路人", MinScore: 40, MinDirectCompletedInvitees: 50, Enabled: true, SortOrder: 30},
		{RoleCode: "guide", LevelCode: "guide3", LevelNo: 3, Title: "3 级领路人", MinScore: 60, MinDirectCompletedInvitees: 100, Enabled: true, SortOrder: 40},
		{RoleCode: "guide", LevelCode: "guide4", LevelNo: 4, Title: "4 级领路人", MinScore: 80, MinDirectCompletedInvitees: 500, Enabled: true, SortOrder: 50},
		{RoleCode: "guide", LevelCode: "guide5", LevelNo: 5, Title: "5 级领路人", MinScore: 90, MinDirectCompletedInvitees: 1000, Enabled: true, SortOrder: 60},
	}}
}

func defaultRoleMetricConfig() roleMetricConfigDTO {
	return roleMetricConfigDTO{Version: "2026-07-25-v1", Items: []roleMetricRuleDTO{
		{RoleCode: "expert", MetricCode: "newcomer_delivery", MetricGroup: "新人交付", Title: "完成新人交付", MetricType: "数量", RawWeight: 40, TargetValue: 10, WindowDays: 90, Enabled: true, SortOrder: 10},
		{RoleCode: "expert", MetricCode: "completion_rate", MetricGroup: "履约完成率", Title: "局成员完成率", MetricType: "比例", RawWeight: 40, TargetValue: 90, WindowDays: 90, Enabled: true, SortOrder: 20},
		{RoleCode: "expert", MetricCode: "demand_recognition", MetricGroup: "需求认可", Title: "复购与主动选择", MetricType: "组合", RawWeight: 20, TargetValue: 60, WindowDays: 90, Enabled: true, SortOrder: 30},
		{RoleCode: "expert", MetricCode: "activation_effect", MetricGroup: "带动效果", Title: "新局与局后触发", MetricType: "组合", RawWeight: 20, TargetValue: 60, WindowDays: 30, Enabled: true, SortOrder: 40},
		{RoleCode: "guide", MetricCode: "invitee_completed", MetricGroup: "邀请完成", Title: "直属邀请用户完成局", MetricType: "数量", RawWeight: 50, TargetValue: 1000, WindowDays: 0, Enabled: true, SortOrder: 10},
		{RoleCode: "guide", MetricCode: "invitee_created_games", MetricGroup: "产生局数", Title: "直属邀请用户完成发局", MetricType: "数量", RawWeight: 50, TargetValue: 100, WindowDays: 0, Enabled: true, SortOrder: 20},
		{RoleCode: "guide", MetricCode: "invitee_replay_rate", MetricGroup: "再玩率", Title: "直属邀请用户再玩率", MetricType: "比例", RawWeight: 20, TargetValue: 60, WindowDays: 90, Enabled: true, SortOrder: 30},
	}}
}

func defaultCreditRestrictionConfig() creditRestrictionConfigDTO {
	return creditRestrictionConfigDTO{InitialScore: 100, ScoreCap: 100, CreateRestrictedBelow: 60, JoinRestrictedBelow: 40, FrozenBelow: 20, Version: "2026-07-25-v1"}
}

func (s *Server) currentRoleLevelConfig() roleLevelConfigDTO {
	config := defaultRoleLevelConfig()
	if s.systemConfig != nil {
		var stored roleLevelConfigDTO
		if s.systemConfig.Get(roleLevelConfigKey, &stored) {
			if normalized, err := normalizeRoleLevelConfig(stored); err == nil {
				config = normalized
			}
		}
	}
	return config
}

func (s *Server) currentRoleMetricConfig() roleMetricConfigDTO {
	config := defaultRoleMetricConfig()
	if s.systemConfig != nil {
		var stored roleMetricConfigDTO
		if s.systemConfig.Get(roleMetricConfigKey, &stored) {
			if normalized, err := normalizeRoleMetricConfig(stored); err == nil {
				config = normalized
			}
		}
	}
	return config
}

func (s *Server) currentCreditRestrictionConfig() creditRestrictionConfigDTO {
	config, err := s.currentCreditRestrictionConfigStrict()
	if err != nil {
		return defaultCreditRestrictionConfig()
	}
	return config
}

func (s *Server) currentCreditRestrictionConfigStrict() (creditRestrictionConfigDTO, error) {
	config := defaultCreditRestrictionConfig()
	if s.systemConfig != nil {
		var stored creditRestrictionConfigDTO
		found, err := s.systemConfig.GetStrict(creditRestrictionConfigKey, &stored)
		if err != nil {
			return creditRestrictionConfigDTO{}, err
		}
		if found {
			normalized, err := normalizeCreditRestrictionConfig(stored)
			if err != nil {
				return creditRestrictionConfigDTO{}, err
			}
			config = normalized
		}
	}
	return config, nil
}

func (s *Server) currentRoleGrowthRecalculationStatus() roleGrowthRecalculationStatusDTO {
	status := roleGrowthRecalculationStatusDTO{
		State: "已完成", Trigger: "初始配置", Description: "成长档案按实时业务数据计算，当前配置已生效。",
	}
	if s.systemConfig != nil {
		var stored roleGrowthRecalculationStatusDTO
		if s.systemConfig.Get(roleGrowthRecalcStatusKey, &stored) && strings.TrimSpace(stored.State) != "" {
			status = stored
		}
	}
	return status
}

// Role profiles are derived from factual source data at read time, therefore
// a rule update has no stale aggregate to rebuild. Persisting this completion
// record makes the immediately-applied recalculation visible and auditable.
func (s *Server) markRoleGrowthRecalculated(trigger string) roleGrowthRecalculationStatusDTO {
	status := roleGrowthRecalculationStatusDTO{
		State: "已完成", Trigger: trigger, CompletedAt: time.Now().Format(time.RFC3339),
		Description: "已按最新规则刷新三角色成长计算结果。",
	}
	if s.systemConfig != nil {
		_ = s.systemConfig.Set(roleGrowthRecalcStatusKey, status)
	}
	return status
}

// playerLevelForExperience is the only player-level conversion used by the
// application. It reads the enabled administrator-configured ladder every
// time, so a saved threshold adjustment takes effect for existing users as
// well as new experience events without requiring a rebuild or data rewrite.
func (s *Server) playerLevelForExperience(experience int) int {
	if experience < 0 {
		experience = 0
	}
	level := 0
	found := false
	for _, item := range s.currentRoleLevelConfig().Items {
		if item.RoleCode != "player" || !item.Enabled || item.MinExperience > experience {
			continue
		}
		if !found || item.LevelNo > level {
			level = item.LevelNo
			found = true
		}
	}
	return level
}

// canUseCreditAction applies the permanent-credit restrictions at the server
// boundary. The client may use the same thresholds for disabled states, but
// cannot bypass this check by calling the API directly.
func (s *Server) canUseCreditAction(userID int64, action string) (bool, string) {
	score, err := s.reviews.CreditScoreStrict(userID)
	if err != nil {
		return false, "信用数据暂时不可用，请稍后重试"
	}
	config, err := s.currentCreditRestrictionConfigStrict()
	if err != nil {
		return false, "信用规则暂时不可用，请稍后重试"
	}
	state, _ := creditRestrictionState(score, config)
	if state == "frozen" {
		return false, "当前信用分已进入冻结状态，仅可查看与提交申诉"
	}
	switch action {
	case "create_game":
		if score < config.CreateRestrictedBelow {
			return false, "当前信用分低于发局阈值，暂不能发起组局"
		}
	case "join_game", "accept_invitation":
		if score < config.JoinRestrictedBelow {
			return false, "当前信用分低于报名阈值，暂不能报名或接受邀请"
		}
	}
	return true, ""
}

func creditRestrictionState(score int, config creditRestrictionConfigDTO) (string, string) {
	if score < config.FrozenBelow {
		return "frozen", "已冻结"
	}
	if score < config.JoinRestrictedBelow {
		return "restricted_join", "限制报名"
	}
	if score < config.CreateRestrictedBelow {
		return "restricted_create", "限制发局"
	}
	return "normal", "正常"
}

func normalizeRoleLevelConfig(config roleLevelConfigDTO) (roleLevelConfigDTO, error) {
	if len(config.Items) == 0 || len(config.Items) > 60 {
		return roleLevelConfigDTO{}, &validationError{"等级规则数量必须在 1 至 60 条之间"}
	}
	seen := map[string]bool{}
	baseLevelEnabled := map[string]bool{}
	last := map[string]int{"player": -1, "expert": -1, "guide": -1}
	lastPlayerExperience := -1
	lastRoleScore := map[string]int{"expert": -1, "guide": -1}
	lastGuideInvitees := -1
	for i := range config.Items {
		item := &config.Items[i]
		item.RoleCode = strings.TrimSpace(item.RoleCode)
		item.LevelCode = strings.TrimSpace(item.LevelCode)
		item.Title = strings.TrimSpace(item.Title)
		if !validGrowthRole(item.RoleCode) || item.LevelCode == "" || item.Title == "" || len(item.LevelCode) > 32 || len(item.Title) > 32 {
			return roleLevelConfigDTO{}, &validationError{"等级规则中的角色、编码或名称不正确"}
		}
		key := item.RoleCode + ":" + item.LevelCode
		if seen[key] || item.LevelNo < 0 || item.LevelNo > 100 || item.MinExperience < 0 || item.MinScore < 0 || item.MinScore > 100 || item.MinDirectCompletedInvitees < 0 {
			return roleLevelConfigDTO{}, &validationError{"等级规则存在重复或数值不正确"}
		}
		if item.LevelNo <= last[item.RoleCode] {
			return roleLevelConfigDTO{}, &validationError{"同一角色的等级顺序必须递增"}
		}
		switch item.RoleCode {
		case "player":
			if item.MinExperience <= lastPlayerExperience {
				return roleLevelConfigDTO{}, &validationError{"玩家等级的累计经验门槛必须逐级递增"}
			}
			lastPlayerExperience = item.MinExperience
		case "expert":
			if item.MinScore <= lastRoleScore[item.RoleCode] {
				return roleLevelConfigDTO{}, &validationError{"行家星级的综合得分门槛必须逐级递增"}
			}
			lastRoleScore[item.RoleCode] = item.MinScore
		case "guide":
			if item.MinScore <= lastRoleScore[item.RoleCode] {
				return roleLevelConfigDTO{}, &validationError{"领路人等级的综合得分门槛必须逐级递增"}
			}
			if item.MinDirectCompletedInvitees < lastGuideInvitees {
				return roleLevelConfigDTO{}, &validationError{"领路人等级的直属完成用户门槛不能低于上一等级"}
			}
			lastRoleScore[item.RoleCode] = item.MinScore
			lastGuideInvitees = item.MinDirectCompletedInvitees
		}
		last[item.RoleCode] = item.LevelNo
		seen[key] = true
		if item.Enabled && item.LevelNo == 0 {
			switch item.RoleCode {
			case "player":
				baseLevelEnabled[item.RoleCode] = item.MinExperience == 0
			case "expert":
				baseLevelEnabled[item.RoleCode] = item.MinScore == 0
			case "guide":
				baseLevelEnabled[item.RoleCode] = item.MinScore == 0 && item.MinDirectCompletedInvitees == 0
			}
		}
	}
	if !baseLevelEnabled["player"] || !baseLevelEnabled["expert"] || !baseLevelEnabled["guide"] {
		return roleLevelConfigDTO{}, &validationError{"玩家、行家和领路人都必须保留启用的 0 级基础档，且基础门槛为 0"}
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = defaultRoleLevelConfig().Version
	}
	sort.SliceStable(config.Items, func(i, j int) bool {
		if config.Items[i].RoleCode == config.Items[j].RoleCode {
			return config.Items[i].LevelNo < config.Items[j].LevelNo
		}
		return config.Items[i].RoleCode < config.Items[j].RoleCode
	})
	return config, nil
}

func normalizeRoleMetricConfig(config roleMetricConfigDTO) (roleMetricConfigDTO, error) {
	if len(config.Items) == 0 || len(config.Items) > 40 {
		return roleMetricConfigDTO{}, &validationError{"指标规则数量必须在 1 至 40 条之间"}
	}
	seen := map[string]bool{}
	weights := map[string]int{}
	for i := range config.Items {
		item := &config.Items[i]
		item.RoleCode = strings.TrimSpace(item.RoleCode)
		item.MetricCode = strings.TrimSpace(item.MetricCode)
		item.MetricGroup = strings.TrimSpace(item.MetricGroup)
		item.Title = strings.TrimSpace(item.Title)
		if (item.RoleCode != "expert" && item.RoleCode != "guide") || item.MetricCode == "" || item.MetricGroup == "" || item.Title == "" || !validMetricType(item.MetricType) || item.RawWeight < 0 || item.RawWeight > 1000 || item.TargetValue < 0 || item.TargetValue > 1000000 || item.WindowDays < 0 || item.WindowDays > 365 {
			return roleMetricConfigDTO{}, &validationError{"等级指标规则内容不正确"}
		}
		key := item.RoleCode + ":" + item.MetricCode
		if seen[key] {
			return roleMetricConfigDTO{}, &validationError{"同一角色的指标编码不能重复"}
		}
		seen[key] = true
		if item.Enabled {
			if item.RawWeight <= 0 || item.TargetValue <= 0 {
				return roleMetricConfigDTO{}, &validationError{"已启用指标的权重和目标值必须大于 0"}
			}
			weights[item.RoleCode] += item.RawWeight
		}
	}
	if weights["expert"] <= 0 || weights["guide"] <= 0 {
		return roleMetricConfigDTO{}, &validationError{"行家和领路人均至少需要一项启用指标"}
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = defaultRoleMetricConfig().Version
	}
	return config, nil
}

func normalizeCreditRestrictionConfig(config creditRestrictionConfigDTO) (creditRestrictionConfigDTO, error) {
	if config.InitialScore != 100 {
		return creditRestrictionConfigDTO{}, &validationError{"新用户信用初始分固定为 100，不支持修改"}
	}
	if config.ScoreCap < config.InitialScore || config.ScoreCap > 1000 {
		return creditRestrictionConfigDTO{}, &validationError{"信用初始分和上限不正确"}
	}
	if config.FrozenBelow < 0 || config.JoinRestrictedBelow <= config.FrozenBelow || config.CreateRestrictedBelow <= config.JoinRestrictedBelow || config.CreateRestrictedBelow > config.ScoreCap {
		return creditRestrictionConfigDTO{}, &validationError{"信用限制阈值必须满足：冻结 < 报名限制 < 发局限制"}
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = defaultCreditRestrictionConfig().Version
	}
	return config, nil
}

func validGrowthRole(role string) bool {
	return role == "player" || role == "expert" || role == "guide"
}
func validMetricType(kind string) bool {
	return kind == "数量" || kind == "比例" || kind == "组合" || kind == "负向扣减"
}

func (s *Server) adminRoleLevelConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentRoleLevelConfig(), "recalculation": s.currentRoleGrowthRecalculationStatus()})
	case http.MethodPut:
		var req roleLevelConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "等级规则格式不正确")
			return
		}
		config, err := normalizeRoleLevelConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(roleLevelConfigKey, config) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存等级规则失败")
			return
		}
		status := s.markRoleGrowthRecalculated("等级梯度更新")
		s.recordOperation(r, "role_level_config:update", "system_config", roleLevelConfigKey, map[string]interface{}{"count": len(config.Items), "version": config.Version})
		httpx.OK(w, map[string]interface{}{"config": config, "recalculation": status})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) adminRoleMetricConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentRoleMetricConfig(), "recalculation": s.currentRoleGrowthRecalculationStatus()})
	case http.MethodPut:
		var req roleMetricConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "等级指标规则格式不正确")
			return
		}
		config, err := normalizeRoleMetricConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(roleMetricConfigKey, config) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存等级指标规则失败")
			return
		}
		status := s.markRoleGrowthRecalculated("指标规则更新")
		s.recordOperation(r, "role_metric_config:update", "system_config", roleMetricConfigKey, map[string]interface{}{"count": len(config.Items), "version": config.Version})
		httpx.OK(w, map[string]interface{}{"config": config, "recalculation": status})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) adminCreditRestrictionConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentCreditRestrictionConfig(), "recalculation": s.currentRoleGrowthRecalculationStatus()})
	case http.MethodPut:
		var req creditRestrictionConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "信用限制规则格式不正确")
			return
		}
		config, err := normalizeCreditRestrictionConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(creditRestrictionConfigKey, config) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存信用限制规则失败")
			return
		}
		status := s.markRoleGrowthRecalculated("信用限制更新")
		s.recordOperation(r, "credit_restriction_config:update", "system_config", creditRestrictionConfigKey, map[string]interface{}{"version": config.Version})
		httpx.OK(w, map[string]interface{}{"config": config, "recalculation": status})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}
