package appapi

import (
	"math"
	"sort"
	"strconv"
	"time"

	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/reviews"
)

// roleGrowthProfileDTO is deliberately derived from source business records
// at read time in phase one. This keeps the values real and immediately
// consistent with completed games, evaluations and invitation relationships;
// the later event-ledger/recalculation job can reuse this same output shape.
type roleGrowthProfileDTO struct {
	RoleCode    string                `json:"roleCode"`
	RoleName    string                `json:"roleName"`
	Active      bool                  `json:"active"`
	LevelNo     int                   `json:"levelNo"`
	LevelCode   string                `json:"levelCode"`
	LevelTitle  string                `json:"levelTitle"`
	Score       int                   `json:"score"`
	ScoreLabel  string                `json:"scoreLabel"`
	Progress    int                   `json:"progressPercent"`
	Description string                `json:"description"`
	Metrics     []roleGrowthMetricDTO `json:"metrics"`
}

type roleGrowthMetricDTO struct {
	MetricCode  string  `json:"metricCode"`
	Title       string  `json:"title"`
	RawValue    float64 `json:"rawValue"`
	TargetValue int     `json:"targetValue"`
	Score       int     `json:"score"`
	Weight      int     `json:"weight"`
	Unit        string  `json:"unit"`
}

func (s *Server) roleGrowthProfilesForUser(userID int64, player reviews.GrowthProfile) []roleGrowthProfileDTO {
	levels := s.currentRoleLevelConfig()
	metrics := s.currentRoleMetricConfig()
	snapshot := s.profiles.RoleSnapshot(userID)
	active := map[string]bool{"player": true}
	for _, role := range snapshot.Roles {
		active[role] = true
	}

	result := []roleGrowthProfileDTO{s.playerGrowthProfile(player, levels)}
	for _, role := range []string{"expert", "guide"} {
		profile := s.roleScoreProfile(userID, role, active[role], levels, metrics)
		result = append(result, profile)
	}
	return result
}

func (s *Server) playerGrowthProfile(player reviews.GrowthProfile, config roleLevelConfigDTO) roleGrowthProfileDTO {
	level := roleLevelForPlayerExperience(player.Experience, config)
	_, _, progress := s.playerLevelProgress(player.Experience, level.LevelNo)
	nextExperience := 0
	for _, item := range config.Items {
		if item.RoleCode == "player" && item.Enabled && item.MinExperience > player.Experience && (nextExperience == 0 || item.MinExperience < nextExperience) {
			nextExperience = item.MinExperience
		}
	}
	description := "累计经验实时升级"
	if nextExperience > 0 {
		description = "距离下一等级还差 " + intText(nextExperience-player.Experience) + " 经验"
	} else {
		description = "已达到当前已配置的最高等级"
	}
	return roleGrowthProfileDTO{
		RoleCode: "player", RoleName: "玩家", Active: true,
		LevelNo: level.LevelNo, LevelCode: level.LevelCode, LevelTitle: level.Title,
		Score: player.Experience, ScoreLabel: "累计经验", Progress: progress, Description: description,
		Metrics: []roleGrowthMetricDTO{{MetricCode: "experience", Title: "累计经验", RawValue: float64(player.Experience), TargetValue: nextExperience, Score: progress, Unit: "经验"}},
	}
}

func (s *Server) roleScoreProfile(userID int64, role string, active bool, levels roleLevelConfigDTO, metricConfig roleMetricConfigDTO) roleGrowthProfileDTO {
	roleName := map[string]string{"expert": "行家", "guide": "领路人"}[role]
	if !active {
		return roleGrowthProfileDTO{
			RoleCode: role, RoleName: roleName, Active: false, LevelTitle: "身份未开通", ScoreLabel: "综合得分",
			Description: "开通" + roleName + "身份后，将按真实业务记录计算等级。", Metrics: []roleGrowthMetricDTO{},
		}
	}
	items := make([]roleMetricRuleDTO, 0)
	for _, item := range metricConfig.Items {
		if item.RoleCode == role && item.Enabled {
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].SortOrder < items[j].SortOrder })
	raw, directCompleted := s.roleGrowthRawValues(userID, role, items)
	weightTotal := 0
	for _, item := range items {
		weightTotal += item.RawWeight
	}
	metrics := make([]roleGrowthMetricDTO, 0, len(items))
	score := 0.0
	for _, item := range items {
		rawValue := raw[item.MetricCode]
		metricScore := normalizedMetricScore(rawValue, item.TargetValue)
		if weightTotal > 0 {
			score += metricScore * float64(item.RawWeight) / float64(weightTotal)
		}
		metrics = append(metrics, roleGrowthMetricDTO{
			MetricCode: item.MetricCode, Title: item.Title, RawValue: rawValue, TargetValue: item.TargetValue,
			Score: int(math.Round(metricScore)), Weight: item.RawWeight, Unit: metricUnit(item.MetricType),
		})
	}
	finalScore := int(math.Round(score))
	level := roleLevelForScore(role, finalScore, directCompleted, levels)
	description := "指标按已启用权重归一化计算；无可审计事件的数据暂不计分"
	if role == "guide" {
		description = "已带动 " + intText(directCompleted) + " 位直属邀请用户完成局"
	}
	return roleGrowthProfileDTO{
		RoleCode: role, RoleName: roleName, Active: true,
		LevelNo: level.LevelNo, LevelCode: level.LevelCode, LevelTitle: level.Title,
		Score: finalScore, ScoreLabel: "综合得分", Progress: finalScore, Description: description, Metrics: metrics,
	}
}

func (s *Server) roleGrowthRawValues(userID int64, role string, items []roleMetricRuleDTO) (map[string]float64, int) {
	values := map[string]float64{}
	windowStarts := roleMetricWindowStarts(items, time.Now())
	allGames := s.games.List()
	if role == "guide" {
		relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
		if err != nil {
			return values, 0
		}
		directCompleted := 0
		completedInvitees := 0
		replayUsers := 0
		createdGames := 0
		for _, relation := range relations {
			completedLifetime := s.completedGamesForUserAfter(allGames, relation.InviteeUserID, time.Time{})
			if completedLifetime > 0 {
				directCompleted++
			}
			if s.completedGamesForUserAfter(allGames, relation.InviteeUserID, windowStarts["invitee_completed"]) > 0 {
				completedInvitees++
			}
			if s.completedGamesForUserAfter(allGames, relation.InviteeUserID, windowStarts["invitee_replay_rate"]) >= 2 {
				replayUsers++
			}
			for _, game := range allGames {
				if game.CreatorUserID == relation.InviteeUserID && isGrowthCompletedGame(game) && gameWithinGrowthWindow(game, windowStarts["invitee_created_games"]) {
					createdGames++
				}
			}
		}
		values["invitee_completed"] = float64(completedInvitees)
		values["invitee_created_games"] = float64(createdGames)
		if len(relations) > 0 {
			values["invitee_replay_rate"] = float64(replayUsers) * 100 / float64(len(relations))
		}
		return values, directCompleted
	}

	completed := 0
	total := 0
	for _, game := range allGames {
		if !gameWithinGrowthWindow(game, windowStarts["completion_rate"]) {
			continue
		}
		for _, memberRole := range s.games.MemberRoles(game.ID) {
			if memberRole.UserID != userID || memberRole.Role != "expert" {
				continue
			}
			total++
			if isGrowthCompletedGame(game) {
				completed++
			}
		}
	}
	// “新人交付”“复购/主动选择”“局后触发”在现有数据表中尚无可审计的
	// 业务标记，不能用完成局、好评或发局数量冒充，暂按 0 展示并等待
	// 对应事件账本接入。履约完成率可由局状态直接、真实地计算。
	values["newcomer_delivery"] = 0
	values["demand_recognition"] = 0
	values["activation_effect"] = 0
	if total > 0 {
		values["completion_rate"] = float64(completed) * 100 / float64(total)
	}
	return values, 0
}

// WindowDays is an operator-controlled score window. Game completion has no
// separate completion timestamp in the phase-one schema, so the auditable
// game creation time is consistently used as the window boundary.
func roleMetricWindowStarts(items []roleMetricRuleDTO, now time.Time) map[string]time.Time {
	starts := make(map[string]time.Time, len(items))
	for _, item := range items {
		if item.WindowDays > 0 {
			starts[item.MetricCode] = now.AddDate(0, 0, -item.WindowDays)
		}
	}
	return starts
}

func gameWithinGrowthWindow(game games.Game, start time.Time) bool {
	return start.IsZero() || !game.CreatedAt.Before(start)
}

func (s *Server) completedGamesForUserAfter(items []games.Game, userID int64, start time.Time) int {
	completed := 0
	for _, game := range items {
		if !isGrowthCompletedGame(game) || !gameWithinGrowthWindow(game, start) {
			continue
		}
		if game.CreatorUserID == userID || containsUserID(s.games.Members(game.ID), userID) {
			completed++
		}
	}
	return completed
}

func isGrowthCompletedGame(game games.Game) bool {
	return game.Status == games.StatusPendingReview || game.Status == games.StatusCompleted || game.Status == games.StatusSettling || game.Status == games.StatusClosed
}

func roleLevelForPlayerExperience(experience int, config roleLevelConfigDTO) roleLevelRuleDTO {
	best := roleLevelRuleDTO{RoleCode: "player", LevelCode: "lv0", Title: "Lv0 新玩家"}
	for _, item := range config.Items {
		if item.RoleCode == "player" && item.Enabled && item.MinExperience <= experience && item.LevelNo >= best.LevelNo {
			best = item
		}
	}
	return best
}

func roleLevelForScore(role string, score int, directCompleted int, config roleLevelConfigDTO) roleLevelRuleDTO {
	best := roleLevelRuleDTO{RoleCode: role, Title: "0 级"}
	for _, item := range config.Items {
		if item.RoleCode != role || !item.Enabled || item.MinScore > score || (role == "guide" && item.MinDirectCompletedInvitees > directCompleted) {
			continue
		}
		if item.LevelNo >= best.LevelNo {
			best = item
		}
	}
	return best
}

func normalizedMetricScore(raw float64, target int) float64 {
	if target <= 0 || raw <= 0 {
		return 0
	}
	return math.Min(100, raw*100/float64(target))
}

func metricUnit(metricType string) string {
	if metricType == "比例" || metricType == "组合" {
		return "%"
	}
	return "次"
}

func intText(value int) string {
	if value < 0 {
		return "0"
	}
	return strconv.Itoa(value)
}
