package appapi

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/lbs"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

const homeDisplayConfigKey = "home.display_config"
const homeRoleDashboardConfigKey = "home.role_dashboard_config"
const homeNearbyDistanceLimitMeter = 10000

type homeDisplayConfigDTO struct {
	OnlineBaseCount int    `json:"onlineBaseCount"`
	OnlineSuffix    string `json:"onlineSuffix"`
}

type homeRoleDashboardConfigDTO struct {
	Version string                              `json:"version"`
	Roles   map[string]homeRoleDashboardRoleDTO `json:"roles"`
}

type homeRoleDashboardRoleDTO struct {
	QuickActions   []homeQuickActionDTO      `json:"quickActions,omitempty"`
	Network        homeRoleNetworkDTO        `json:"network,omitempty"`
	Recommendation homeRoleRecommendationDTO `json:"recommendation,omitempty"`
}

type homeQuickActionDTO struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title"`
	Desc      string `json:"desc,omitempty"`
	Icon      string `json:"icon,omitempty"`
	IconSrc   string `json:"iconSrc,omitempty"`
	Theme     string `json:"theme,omitempty"`
	Route     string `json:"route,omitempty"`
	RouteIcon bool   `json:"routeIcon,omitempty"`
}

type homeRoleNetworkDTO struct {
	Title        string                     `json:"title,omitempty"`
	Status       string                     `json:"status,omitempty"`
	LocationText string                     `json:"locationText,omitempty"`
	Items        []homeRoleNetworkItemDTO   `json:"items,omitempty"`
	Buttons      []homeRoleNetworkButtonDTO `json:"buttons,omitempty"`
}

type homeRoleNetworkItemDTO struct {
	ID     string `json:"id,omitempty"`
	Key    string `json:"key,omitempty"`
	Icon   string `json:"icon,omitempty"`
	Name   string `json:"name,omitempty"`
	Desc   string `json:"desc,omitempty"`
	Dashed bool   `json:"dashed,omitempty"`
}

type homeRoleNetworkButtonDTO struct {
	Text    string `json:"text,omitempty"`
	Route   string `json:"route,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type homeRoleRecommendationDTO struct {
	Title string `json:"title,omitempty"`
}

func (s *Server) buildAppHomePayload(userID int64, visibleGames []games.Game, roleType string) map[string]interface{} {
	if len(visibleGames) > 10 {
		visibleGames = visibleGames[:10]
	}
	roleType = normalizeHomeRoleType(roleType)
	if roleType == "" {
		roleType = s.defaultHomeRoleType(userID)
	}
	user, _ := s.auth.UserByID(userID)
	growth := s.reviews.Profile(userID)
	stats := s.games.StatsForUser(userID)
	conns := s.connections.My(userID)
	nearbyGames := s.homeNearbyGames(userID, visibleGames, 6)
	cityGames := s.homeCityGames(userID, visibleGames, nearbyGames, 6)
	friendGames := s.homeFriendGames(userID, visibleGames, conns, 4)
	nearbySummary := s.homeNearbySummary(userID, nearbyGames, conns)
	rankingBoards := s.homeRankingBoards(userID, user, stats, growth, conns)
	visualization := s.homeVisualization(userID, visibleGames, nearbyGames, conns)
	currentUser := s.buildCurrentUserDTO(user, s.identity.Status(userID))
	roleHomeConfig := s.currentHomeRoleDashboardConfig().Roles[roleType]

	payload := map[string]interface{}{
		"hero":               s.homeHero(userID, user, growth, len(visibleGames), len(conns), roleType),
		"user":               currentUser,
		"playerSummary":      s.homePlayerSummary(userID, user, stats, growth, roleType),
		"nearbySummary":      nearbySummary,
		"onlineCard":         s.homeOnlineCard(nearbySummary),
		"nearbySection":      s.homeNearbySection(),
		"nearbyGames":        nearbyGames,
		"recommendedGames":   cityGames,
		"friendSection":      s.homeFriendSection(len(friendGames)),
		"friendGames":        friendGames,
		"rankingSection":     s.homeRankingSection(),
		"rankingBoards":      rankingBoards,
		"achievementSection": s.homeAchievementSection(),
		"achievements":       s.homeAchievements(stats, growth, nearbyGames),
		"metaverseEntry":     s.homeMetaverseEntry(len(conns), len(visibleGames), s.homeMetaverseAvatars(userID, visibleGames, conns)),
		"earth":              s.homeEarth(userID, visibleGames, nearbyGames, conns),
		"visualization":      visualization,
		"games":              visibleGames,
		"userStats":          stats,
		"pointsSummary":      s.points.Summary(userID),
		"growth":             growth,
		"notifications":      map[string]interface{}{"unreadCount": unreadNotificationCount(s.notices.List(userID))},
	}

	if len(roleHomeConfig.QuickActions) > 0 {
		payload["quickActions"] = roleHomeConfig.QuickActions
	}
	if roleType == "expert" {
		payload["skills"] = s.homeExpertSkills(userID)
	}
	if roleType == "guide" {
		payload["network"] = s.homeRoleNetwork(roleHomeConfig.Network, conns, currentUser.IncomeSummary.PendingCent, currentUser.IncomeSummary.TotalCent)
		if recommendation := s.homeRoleRecommendation(roleHomeConfig.Recommendation, rankingBoards); recommendation != nil {
			payload["recommendation"] = recommendation
		}
	}

	return payload
}

func (s *Server) homeOnlineCard(nearbySummary map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"title": "地球online",
		"desc":  "探索城市副本 · 解锁地图成就",
		"tags": []string{
			"附近 " + strconv.Itoa(homeSummaryInt(nearbySummary, "nearbyGameCount")) + " 个组局",
			"已打卡 " + strconv.Itoa(homeSummaryInt(nearbySummary, "checkedInCount")) + " 处",
		},
	}
}

func homeSummaryInt(summary map[string]interface{}, key string) int {
	if summary == nil {
		return 0
	}
	switch value := summary[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func (s *Server) homeHero(userID int64, user users.User, growth reviews.GrowthProfile, gameCount int, connectionCount int, roleType string) map[string]interface{} {
	return map[string]interface{}{
		"onlineText":  s.homeOnlineText(gameCount, connectionCount),
		"roleType":    roleType,
		"roleName":    homeRoleName(roleType),
		"currentRole": roleType,
		"dateLabel":   time.Now().Format("2006.01.02"),
		"subtitle":    "\u4eca\u65e5\u63a8\u8350 " + strconv.Itoa(gameCount) + " \u4e2a\u7ec4\u5c40",
		"nickname":    homeDisplayName(user, s.displayName(userID, "\u7528\u6237")),
		"level":       growth.Level,
	}
}

func (s *Server) homeOnlineText(gameCount int, connectionCount int) string {
	config := s.currentHomeDisplayConfig()
	suffix := strings.TrimSpace(config.OnlineSuffix)
	if suffix == "" {
		suffix = "\u4eba\u5728\u7ebf"
	}
	return strconv.Itoa(maxInt(1, config.OnlineBaseCount+gameCount+connectionCount+1)) + suffix
}

func (s *Server) currentHomeDisplayConfig() homeDisplayConfigDTO {
	config := defaultHomeDisplayConfig()
	var stored homeDisplayConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(homeDisplayConfigKey, &stored) {
		if stored.OnlineBaseCount >= 0 {
			config.OnlineBaseCount = stored.OnlineBaseCount
		}
		if strings.TrimSpace(stored.OnlineSuffix) != "" {
			config.OnlineSuffix = strings.TrimSpace(stored.OnlineSuffix)
		}
	}
	return config
}

func (s *Server) currentHomeRoleDashboardConfig() homeRoleDashboardConfigDTO {
	config := defaultHomeRoleDashboardConfig()
	var stored homeRoleDashboardConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(homeRoleDashboardConfigKey, &stored) && len(stored.Roles) > 0 {
		return mergeHomeRoleDashboardConfig(config, stored)
	}
	return config
}

func mergeHomeRoleDashboardConfig(defaultConfig, stored homeRoleDashboardConfigDTO) homeRoleDashboardConfigDTO {
	if strings.TrimSpace(stored.Version) != "" {
		defaultConfig.Version = strings.TrimSpace(stored.Version)
	}
	if defaultConfig.Roles == nil {
		defaultConfig.Roles = map[string]homeRoleDashboardRoleDTO{}
	}
	for key, roleConfig := range stored.Roles {
		roleType := normalizeHomeRoleType(key)
		if roleType == "" {
			roleType = strings.TrimSpace(key)
		}
		if roleType == "" {
			continue
		}
		defaultConfig.Roles[roleType] = mergeHomeRoleConfig(defaultConfig.Roles[roleType], roleConfig)
	}
	return defaultConfig
}

func mergeHomeRoleConfig(defaultConfig, stored homeRoleDashboardRoleDTO) homeRoleDashboardRoleDTO {
	if len(stored.QuickActions) > 0 {
		defaultConfig.QuickActions = stored.QuickActions
	}
	defaultConfig.Network = mergeHomeRoleNetworkConfig(defaultConfig.Network, stored.Network)
	if strings.TrimSpace(stored.Recommendation.Title) != "" {
		defaultConfig.Recommendation = stored.Recommendation
	}
	return defaultConfig
}

func mergeHomeRoleNetworkConfig(defaultConfig, stored homeRoleNetworkDTO) homeRoleNetworkDTO {
	if strings.TrimSpace(stored.Title) != "" {
		defaultConfig.Title = stored.Title
	}
	if strings.TrimSpace(stored.Status) != "" {
		defaultConfig.Status = stored.Status
	}
	if strings.TrimSpace(stored.LocationText) != "" {
		defaultConfig.LocationText = stored.LocationText
	}
	if len(stored.Items) > 0 {
		defaultConfig.Items = stored.Items
	}
	if len(stored.Buttons) > 0 {
		defaultConfig.Buttons = stored.Buttons
	}
	return defaultConfig
}

func (s *Server) adminHomeDisplayConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentHomeDisplayConfig()})
	case http.MethodPut:
		var req homeDisplayConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid home display config")
			return
		}
		config, err := normalizeHomeDisplayConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig != nil {
			if err := s.systemConfig.Set(homeDisplayConfigKey, config); err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save home display config failed")
				return
			}
		}
		s.recordOperation(r, "home_display_config:update", "system_config", "home_display_config", map[string]interface{}{
			"onlineBaseCount": config.OnlineBaseCount,
			"onlineSuffix":    config.OnlineSuffix,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentHomeDisplayConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func normalizeHomeDisplayConfig(req homeDisplayConfigDTO) (homeDisplayConfigDTO, error) {
	if req.OnlineBaseCount < 0 {
		return homeDisplayConfigDTO{}, errors.New("onlineBaseCount must be greater than or equal to 0")
	}
	config := req
	config.OnlineSuffix = strings.TrimSpace(req.OnlineSuffix)
	if config.OnlineSuffix == "" {
		config.OnlineSuffix = "\u4eba\u5728\u7ebf"
	}
	if len([]rune(config.OnlineSuffix)) > 12 {
		return homeDisplayConfigDTO{}, errors.New("onlineSuffix too long")
	}
	return config, nil
}

func defaultHomeDisplayConfig() homeDisplayConfigDTO {
	return homeDisplayConfigDTO{
		OnlineBaseCount: 0,
		OnlineSuffix:    "\u4eba\u5728\u7ebf",
	}
}

func defaultHomeRoleDashboardConfig() homeRoleDashboardConfigDTO {
	return homeRoleDashboardConfigDTO{
		Version: "2026-07-04",
		Roles: map[string]homeRoleDashboardRoleDTO{
			"player": {
				QuickActions: []homeQuickActionDTO{
					{ID: "create", Title: "发起组局", Desc: "创建你的带局房间", Icon: "📍", Theme: "pink", Route: "pages/game/create/index"},
					{ID: "lobby", Title: "局前大厅", Desc: "准备就绪加入一局", Theme: "cyan", Route: "pages/game/hall/index", RouteIcon: true},
				},
			},
			"guide": {
				QuickActions: []homeQuickActionDTO{
					{ID: "lobby", Title: "局前大厅", Desc: "准备加入一局", Theme: "pink", Route: "pages/game/hall/index", RouteIcon: true},
					{ID: "invite", Title: "我的邀约", Desc: "管理连接的玩家", Icon: "📍", Theme: "cyan", Route: "pages/profile/service-center/invite/overview/index", RouteIcon: true},
				},
				Network: homeRoleNetworkDTO{
					Title:        "我的关系网络",
					Status:       "实时连接中",
					LocationText: "核心区",
					Items: []homeRoleNetworkItemDTO{
						{ID: "relations", Key: "relations", Icon: "👑", Name: "累计连接"},
						{ID: "strong", Key: "strongRelations", Icon: "🎓", Name: "强关系"},
						{ID: "nodes", Key: "onlineNodes", Icon: "👶", Name: "动态节点"},
						{ID: "income", Key: "income", Icon: "🏛️", Name: "本周收益"},
						{ID: "more", Key: "more", Icon: "+", Name: "更多", Desc: "待加入", Dashed: true},
					},
					Buttons: []homeRoleNetworkButtonDTO{
						{Text: "管理我的连接", Route: "pages/relation/network/index", Primary: true},
						{Text: "查看分润", Route: "pages/profile/service-center/invite/income/index"},
					},
				},
				Recommendation: homeRoleRecommendationDTO{Title: "推荐行家"},
			},
		},
	}
}

func (s *Server) homeRoleNetwork(config homeRoleNetworkDTO, conns []connections.Connection, pendingCent int64, totalCent int64) map[string]interface{} {
	if strings.TrimSpace(config.Title) == "" && len(config.Items) == 0 && len(config.Buttons) == 0 {
		return nil
	}
	relationCount := len(conns)
	strongCount := countStrongConnections(conns)
	nodeCount := maxInt(1, relationCount+1)
	incomeCent := pendingCent
	if incomeCent == 0 {
		incomeCent = totalCent
	}
	items := make([]map[string]interface{}, 0, len(config.Items))
	for _, item := range config.Items {
		items = append(items, map[string]interface{}{
			"id":     firstNonEmpty(item.ID, item.Key),
			"key":    item.Key,
			"icon":   item.Icon,
			"name":   item.Name,
			"desc":   homeRoleNetworkItemDesc(item, relationCount, strongCount, nodeCount, incomeCent),
			"dashed": item.Dashed,
		})
	}
	buttons := make([]map[string]interface{}, 0, len(config.Buttons))
	for _, item := range config.Buttons {
		if strings.TrimSpace(item.Text) == "" {
			continue
		}
		buttons = append(buttons, map[string]interface{}{
			"text":    item.Text,
			"route":   item.Route,
			"primary": item.Primary,
		})
	}
	return map[string]interface{}{
		"title":        config.Title,
		"status":       config.Status,
		"locationText": config.LocationText,
		"summary":      "已连接 " + strconv.Itoa(relationCount) + " 位玩家",
		"income":       "本周收益 " + moneyYuanText(incomeCent),
		"items":        items,
		"buttons":      buttons,
	}
}

func homeRoleNetworkItemDesc(item homeRoleNetworkItemDTO, relationCount int, strongCount int, nodeCount int, incomeCent int64) string {
	if strings.TrimSpace(item.Desc) != "" {
		return item.Desc
	}
	switch item.Key {
	case "relations":
		return strconv.Itoa(relationCount) + "人"
	case "strongRelations":
		return strconv.Itoa(strongCount) + "人"
	case "onlineNodes":
		return strconv.Itoa(nodeCount) + "个"
	case "income":
		return moneyYuanText(incomeCent)
	default:
		return ""
	}
}

func (s *Server) homeRoleRecommendation(config homeRoleRecommendationDTO, rankingBoards map[string]interface{}) map[string]interface{} {
	if strings.TrimSpace(config.Title) == "" {
		return nil
	}
	items := []map[string]interface{}{}
	if board, ok := rankingBoards["expert"].(map[string]interface{}); ok {
		if list, ok := board["list"].([]map[string]interface{}); ok {
			for _, item := range list {
				if len(items) >= 3 {
					break
				}
				if isMe, _ := item["isMe"].(bool); isMe {
					continue
				}
				items = append(items, map[string]interface{}{
					"id":             item["userId"],
					"name":           item["name"],
					"avatarUrl":      item["avatarUrl"],
					"avatarFallback": item["avatarFallback"],
				})
			}
		}
	}
	return map[string]interface{}{"title": config.Title, "items": items}
}

func (s *Server) homeExpertSkills(userID int64) []map[string]interface{} {
	tags := make([]string, 0, 3)
	source := "expert_skill_profile"
	if profile, err := s.profiles.ExpertSkill(userID); err == nil {
		tags = appendHomeSkillTags(tags, profile.SkillTree...)
		tags = appendHomeSkillTags(tags, profile.ServiceTags...)
	}
	if len(tags) == 0 {
		source = "role_application"
		tags = appendHomeSkillTags(tags, s.latestApprovedExpertApplicationSkills(userID)...)
	}
	return buildHomeExpertSkillNodes(tags, source)
}

func (s *Server) latestApprovedExpertApplicationSkills(userID int64) []string {
	apps := s.profiles.RoleApplicationsByUser(userID)
	sort.SliceStable(apps, func(i, j int) bool {
		if apps[i].UpdatedAt.Equal(apps[j].UpdatedAt) {
			return apps[i].ID > apps[j].ID
		}
		return apps[i].UpdatedAt.After(apps[j].UpdatedAt)
	})
	for _, app := range apps {
		if app.RoleCode == "expert" && app.Status == "approved" {
			return expertApplicationSkills(app)
		}
	}
	return nil
}

func expertApplicationSkills(app profiles.RoleApplication) []string {
	fields := parseRoleApplicationAbilityFields(app.AbilityDescription)
	tags := make([]string, 0, 3)
	tags = appendHomeSkillTags(tags, fields["技能领域"], app.Reason)
	tags = appendHomeSkillTags(tags, splitHomeSkillTags(fields["技能标签"])...)
	return tags
}

func parseRoleApplicationAbilityFields(text string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "：", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(line, ":", 2)
		}
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" && value != "" {
			fields[key] = value
		}
	}
	return fields
}

func splitHomeSkillTags(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case ',', '，', '、', '/', '／', '|', '｜', ';', '；', ' ', '\t', '\r', '\n':
			return true
		default:
			return false
		}
	})
}

func appendHomeSkillTags(tags []string, values ...string) []string {
	seen := map[string]bool{}
	for _, item := range tags {
		seen[item] = true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		tags = append(tags, value)
		seen[value] = true
		if len(tags) >= 3 {
			break
		}
	}
	return tags
}

func buildHomeExpertSkillNodes(tags []string, source string) []map[string]interface{} {
	tones := []string{"green", "orange", "blue"}
	icons := []string{"🎯", "★", "◆"}
	nodes := make([]map[string]interface{}, 0, 3)
	for index := 0; index < 3; index++ {
		if index < len(tags) {
			nodes = append(nodes, map[string]interface{}{
				"id":       "expert-skill-" + strconv.Itoa(index+1),
				"title":    tags[index],
				"icon":     icons[index],
				"tone":     tones[index],
				"locked":   false,
				"unlocked": true,
				"source":   source,
			})
			continue
		}
		nodes = append(nodes, map[string]interface{}{
			"id":       "expert-skill-locked-" + strconv.Itoa(index+1),
			"title":    "待解锁",
			"icon":     "🔒",
			"tone":     "locked",
			"locked":   true,
			"unlocked": false,
			"source":   source,
		})
	}
	return nodes
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizeHomeRoleType(roleType string) string {
	switch strings.TrimSpace(roleType) {
	case "player", "\u73a9\u5bb6":
		return "player"
	case "expert", "master", "\u884c\u5bb6":
		return "expert"
	case "guide", "leader", "\u9886\u8def\u4eba":
		return "guide"
	default:
		return ""
	}
}

func (s *Server) resolveHomeRoleType(userID int64, requestedRoleType string) (string, bool) {
	trimmed := strings.TrimSpace(requestedRoleType)
	roleType := normalizeHomeRoleType(trimmed)
	if trimmed != "" && roleType == "" {
		return "", false
	}
	if roleType == "" {
		return s.defaultHomeRoleType(userID), true
	}
	if roleType == "player" {
		return roleType, true
	}
	snapshot := s.profiles.RoleSnapshot(userID)
	return roleType, snapshot.RoleStatusMap[roleType] == "approved"
}

func (s *Server) defaultHomeRoleType(userID int64) string {
	snapshot := s.profiles.RoleSnapshot(userID)
	if snapshot.RoleStatusMap["guide"] == "approved" {
		return "guide"
	}
	if snapshot.RoleStatusMap["expert"] == "approved" {
		return "expert"
	}
	return "player"
}

func homeRoleName(roleType string) string {
	switch roleType {
	case "guide":
		return "\u9886\u8def\u4eba"
	case "expert":
		return "\u884c\u5bb6"
	default:
		return "\u73a9\u5bb6"
	}
}

func (s *Server) homePlayerSummary(userID int64, user users.User, stats games.UserStats, growth reviews.GrowthProfile, roleType string) map[string]interface{} {
	nextLevelExperience := maxInt(growth.Level*100, 100)
	expToNext := nextLevelExperience - growth.Experience
	if expToNext < 0 {
		expToNext = 0
	}
	progress := 0
	if nextLevelExperience > 0 {
		progress = int(math.Round(float64(growth.Experience) / float64(nextLevelExperience) * 100))
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
	}
	joinCount := stats.Participated
	participationRate := "0%"
	if joinCount > 0 {
		participationRate = "100%"
	}
	return map[string]interface{}{
		"currentRole":         roleType,
		"roleType":            roleType,
		"roleLabel":           homeRoleName(roleType) + " Lv." + strconv.Itoa(growth.Level),
		"displayName":         homeDisplayName(user, s.displayName(userID, "\u7528\u6237")),
		"level":               growth.Level,
		"experience":          growth.Experience,
		"nextLevelExperience": nextLevelExperience,
		"scoreText":           strconv.Itoa(growth.Experience) + "/" + strconv.Itoa(nextLevelExperience) + " XP",
		"nextLevelText":       "\u8ddd\u79bb\u4e0b\u4e00\u7b49\u7ea7\u8fd8\u9700 " + strconv.Itoa(expToNext) + " \u7ecf\u9a8c\u503c",
		"expToNextLevel":      expToNext,
		"progressPercent":     progress,
		"joinCount":           joinCount,
		"monthlyMvpCount":     len(growth.Achievements),
		"participationRate":   participationRate,
		"stats": []map[string]interface{}{
			{"value": joinCount, "label": "\u53c2\u4e0e\u5c40\u6570"},
			{"value": stats.Completed, "label": "\u5b8c\u6210\u5c40\u6570"},
			{"value": growth.CreditScore, "label": "\u4fe1\u7528\u5206"},
		},
	}
}

func (s *Server) homeNearbySummary(userID int64, nearbyGames []map[string]interface{}, conns []connections.Connection) map[string]interface{} {
	recent := s.lbs.Recent(userID, 20)
	return map[string]interface{}{
		"nearbyGameCount": len(nearbyGames),
		"checkedInCount":  len(recent),
		"relationCount":   len(conns),
		"onlineNodeCount": maxInt(1, len(nearbyGames)+len(conns)+1),
	}
}

func (s *Server) homeNearbySection() map[string]interface{} {
	return map[string]interface{}{
		"title": "\u9644\u8fd1\u6b63\u5728\u53d1\u751f",
		"tabs": []map[string]string{
			{"key": "all", "name": "\u5168\u90e8"},
			{"key": "nearby", "name": "\u9644\u8fd1"},
			{"key": "city", "name": "\u540c\u57ce"},
		},
	}
}

func (s *Server) homeFriendSection(count int) map[string]interface{} {
	return map[string]interface{}{
		"icon":     "\u2726",
		"title":    "\u670b\u53cb\u5728\u73a9",
		"count":    count,
		"moreText": "\u67e5\u770b\u5168\u90e8",
	}
}

func (s *Server) homeNearbyGames(userID int64, visibleGames []games.Game, limit int) []map[string]interface{} {
	location, ok := s.lbs.Current(userID)
	if !ok {
		return []map[string]interface{}{}
	}
	items := make([]games.Game, 0, len(visibleGames))
	for _, game := range visibleGames {
		if game.Longitude == 0 && game.Latitude == 0 {
			continue
		}
		distance := lbs.DistanceMeter(location.Longitude, location.Latitude, game.Longitude, game.Latitude)
		if distance > homeNearbyDistanceLimitMeter {
			continue
		}
		game.DistanceMeter = math.Round(distance)
		game.DistanceLabel = lbs.FormatDistanceLabel(game.DistanceMeter)
		items = append(items, game)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].DistanceMeter < items[j].DistanceMeter
	})
	return s.homeGameCards(items, "nearby", limit)
}

func (s *Server) homeCityGames(userID int64, visibleGames []games.Game, nearbyGames []map[string]interface{}, limit int) []map[string]interface{} {
	nearbyIDs := make(map[int64]bool, len(nearbyGames))
	for _, item := range nearbyGames {
		id := mapInt64(item, "id")
		if id > 0 {
			nearbyIDs[id] = true
		}
	}

	location, hasLocation := s.lbs.Current(userID)
	items := make([]games.Game, 0, len(visibleGames))
	for _, game := range visibleGames {
		if nearbyIDs[game.ID] {
			continue
		}
		if hasLocation && !sameHomeCity(game, location) {
			continue
		}
		items = append(items, game)
	}
	return s.homeGameCards(items, "city", limit)
}

func sameHomeCity(game games.Game, location lbs.Location) bool {
	if strings.TrimSpace(location.CityCode) != "" && strings.TrimSpace(game.CityCode) != "" {
		return strings.TrimSpace(location.CityCode) == strings.TrimSpace(game.CityCode)
	}
	if strings.TrimSpace(location.CityName) != "" && strings.TrimSpace(game.CityName) != "" {
		return strings.TrimSpace(location.CityName) == strings.TrimSpace(game.CityName)
	}
	return true
}

func mapInt64(item map[string]interface{}, key string) int64 {
	value, ok := item[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed
	default:
		return 0
	}
}

func (s *Server) homeFriendGames(userID int64, visibleGames []games.Game, conns []connections.Connection, limit int) []map[string]interface{} {
	connected := make(map[int64]bool, len(conns))
	for _, conn := range conns {
		connected[conn.ConnectedUserID] = true
	}
	items := make([]games.Game, 0)
	for _, game := range visibleGames {
		if connected[game.CreatorUserID] || connected[game.MainGuideUserID] {
			items = append(items, game)
			continue
		}
		for _, memberID := range s.games.Members(game.ID) {
			if connected[memberID] {
				items = append(items, game)
				break
			}
		}
	}
	if len(items) == 0 {
		for _, game := range visibleGames {
			if game.CreatorUserID != userID {
				items = append(items, game)
			}
		}
	}
	return s.homeGameCards(items, "friend", limit)
}

func (s *Server) homeGameCards(items []games.Game, scope string, limit int) []map[string]interface{} {
	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}
	cards := make([]map[string]interface{}, 0, limit)
	for index := 0; index < limit; index++ {
		game := items[index]
		cards = append(cards, map[string]interface{}{
			"id":            game.ID,
			"route":         "pages/game/detail/index?id=" + strconv.FormatInt(game.ID, 10),
			"scope":         scope,
			"title":         game.Title,
			"typeText":      homeGameTypeText(game.GameType),
			"statusText":    homeGameStatusText(game.Status),
			"coverSrc":      homeGameCover(index),
			"priceText":     homeGamePriceText(game.GameType),
			"actionText":    homeGameActionText(game.Status),
			"cityName":      game.CityName,
			"address":       game.Address,
			"distanceText":  game.DistanceLabel,
			"memberText":    strconv.Itoa(game.CurrentPlayers) + "/" + strconv.Itoa(game.MaxPlayers) + "\u4eba",
			"timeText":      game.CreatedAt.Format("2006-01-02 15:04"),
			"joinedCount":   game.CurrentPlayers,
			"joinedText":    "+" + strconv.Itoa(game.CurrentPlayers) + "\u4f4d\u73a9\u5bb6\u5df2\u5165\u5c40",
			"playerAvatars": s.homeGamePlayerAvatars(game),
			"actions":       []string{"share", "follow", "refer", "greet"},
			"longitude":     game.Longitude,
			"latitude":      game.Latitude,
			"creatorUserId": game.CreatorUserID,
		})
	}
	return cards
}

func homeGameTypeText(gameType string) string {
	switch gameType {
	case "standard":
		return "\u6807\u51c6\u5c40"
	case "aa":
		return "AA\u5c40"
	case "crowdfund":
		return "\u4f17\u7b79\u5c40"
	case "deposit":
		return "\u4fdd\u8bc1\u91d1\u5c40"
	case "condition":
		return "\u6761\u4ef6\u5c40"
	case "public_welfare":
		return "\u516c\u76ca\u5c40"
	default:
		return "\u666e\u901a\u5c40"
	}
}

func homeGameStatusText(status string) string {
	switch status {
	case "full":
		return "\u5df2\u6ee1\u5458"
	case "in_progress":
		return "\u8fdb\u884c\u4e2d"
	case "pending_confirm":
		return "\u5f85\u786e\u8ba4"
	case "pending_review":
		return "\u5f85\u8bc4\u4ef7"
	case "completed":
		return "\u5df2\u5b8c\u6210"
	default:
		return "\u62db\u52df\u4e2d"
	}
}

func homeGamePriceText(gameType string) string {
	if gameType == "" || gameType == "free" {
		return "\u514d\u8d39"
	}
	return "\u540e\u53f0\u5f00\u5c40"
}

func homeGameActionText(status string) string {
	if status == "recruiting" {
		return "\u52a0\u5165"
	}
	return "\u67e5\u770b"
}

func homeGameCover(index int) string {
	covers := []string{
		"/components/game-card/assets/cover-city.png",
		"/components/game-card/assets/cover-sunset.png",
	}
	return covers[index%len(covers)]
}

func (s *Server) homeRankingSection() map[string]interface{} {
	return map[string]interface{}{
		"icon":     "\U0001F3C6",
		"title":    "\u672c\u5468\u73a9\u9738\u699c",
		"moreText": "\u67e5\u770b\u5168\u90e8\u699c\u5355",
		"tabs": []map[string]string{
			{"key": "player", "name": "\u73a9\u5bb6"},
			{"key": "expert", "name": "\u884c\u5bb6"},
			{"key": "guide", "name": "\u9886\u8def\u4eba"},
		},
	}
}

func (s *Server) homeRankingBoards(userID int64, user users.User, stats games.UserStats, growth reviews.GrowthProfile, conns []connections.Connection) map[string]interface{} {
	me := s.homeRankingItem(1, userID, homeDisplayName(user, s.displayName(userID, "\u7528\u6237")), stats, growth, true)
	networkItems := make([]map[string]interface{}, 0, len(conns))
	for index, conn := range conns {
		connUser, _ := s.auth.UserByID(conn.ConnectedUserID)
		connStats := s.games.StatsForUser(conn.ConnectedUserID)
		connGrowth := s.reviews.Profile(conn.ConnectedUserID)
		networkItems = append(networkItems, s.homeRankingItem(index+1, conn.ConnectedUserID, homeDisplayName(connUser, s.displayName(conn.ConnectedUserID, "\u7528\u6237")), connStats, connGrowth, false))
	}
	if len(networkItems) == 0 {
		networkItems = append(networkItems, me)
	}
	board := map[string]interface{}{"list": networkItems, "myRank": me}
	return map[string]interface{}{
		"player": board,
		"expert": board,
		"guide":  board,
	}
}

func (s *Server) homeRankingItem(rank int, userID int64, name string, stats games.UserStats, growth reviews.GrowthProfile, isMe bool) map[string]interface{} {
	return map[string]interface{}{
		"userId":          userID,
		"rank":            rank,
		"avatarUrl":       s.userAvatarURL(userID),
		"avatarFallback":  avatarTextForName(name, userID),
		"name":            name,
		"gameCount":       stats.Participated,
		"weeklyGameCount": stats.Participated,
		"weeklyMvpCount":  len(growth.Achievements),
		"experience":      growth.Experience,
		"xp":              growth.Experience,
		"isMe":            isMe,
	}
}

func (s *Server) homeGamePlayerAvatars(game games.Game) []map[string]interface{} {
	userIDs := make([]int64, 0, 3)
	seen := map[int64]bool{}
	addUserID := func(userID int64) {
		if userID <= 0 || seen[userID] || len(userIDs) >= 3 {
			return
		}
		seen[userID] = true
		userIDs = append(userIDs, userID)
	}

	addUserID(game.CreatorUserID)
	for _, userID := range s.games.Members(game.ID) {
		addUserID(userID)
	}

	avatars := make([]map[string]interface{}, 0, len(userIDs))
	for _, userID := range userIDs {
		name := s.displayName(userID, "\u7528\u6237")
		if user, ok := s.auth.UserByID(userID); ok {
			name = homeDisplayName(user, name)
			avatars = append(avatars, map[string]interface{}{
				"userId":         userID,
				"name":           name,
				"avatarUrl":      strings.TrimSpace(user.AvatarURL),
				"avatarFallback": avatarTextForName(name, userID),
			})
			continue
		}
		avatars = append(avatars, map[string]interface{}{
			"userId":         userID,
			"name":           name,
			"avatarFallback": avatarTextForName(name, userID),
		})
	}
	return avatars
}

func (s *Server) userAvatarURL(userID int64) string {
	if user, ok := s.auth.UserByID(userID); ok {
		return strings.TrimSpace(user.AvatarURL)
	}
	return ""
}

func (s *Server) homeAchievementSection() map[string]interface{} {
	return map[string]interface{}{
		"icon":  "\U0001F48E",
		"title": "\u6211\u7684\u6210\u5c31",
	}
}

func (s *Server) homeAchievements(stats games.UserStats, growth reviews.GrowthProfile, nearbyGames []map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{
		{"id": "hundred_king", "title": "\u767e\u573a\u738b\u8005", "icon": "\U0001F3C6", "status": "\u7b49\u7ea7", "unlocked": true, "locked": false},
		{"id": "pilot_king", "title": "\u5f15\u822a\u738b\u8005", "icon": "\U0001F3C6", "status": "\u7b49\u7ea7", "unlocked": true, "locked": false},
		{"id": "earth_roamer", "title": "\u5730\u7403\u6f2b\u6e38\u8005", "icon": "\U0001F30D", "status": "\u8fdb\u5ea620%", "unlocked": true, "locked": false, "progressPercent": 20},
		{"id": "hidden_badge", "title": "", "icon": "\U0001F512", "status": "\u672a\u89e3\u9501", "unlocked": false, "locked": true},
	}
}

func (s *Server) homeMetaverseEntry(connectionCount int, gameCount int, avatars []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"title":       "\u8fdb\u5165\u5143\u5b87\u5b99",
		"description": "\u57fa\u4e8e\u7ec4\u5c40\u548c\u4eba\u8109\u7684\u52a8\u6001\u5173\u7cfb\u7f51",
		"tags":        []string{"\u5173\u7cfb\u7f51", "\u5730\u56fe\u8282\u70b9"},
		"avatars":     avatars,
		"onlineCount": maxInt(1, connectionCount+gameCount),
	}
}

func (s *Server) homeMetaverseAvatars(userID int64, visibleGames []games.Game, conns []connections.Connection) []map[string]interface{} {
	userIDs := make([]int64, 0, 3)
	seen := map[int64]bool{userID: true}
	addUserID := func(id int64) {
		if id <= 0 || seen[id] || len(userIDs) >= 3 {
			return
		}
		seen[id] = true
		userIDs = append(userIDs, id)
	}

	for _, conn := range conns {
		addUserID(conn.ConnectedUserID)
	}
	for _, game := range visibleGames {
		for _, id := range s.games.Members(game.ID) {
			addUserID(id)
		}
		addUserID(game.CreatorUserID)
	}

	avatars := make([]map[string]interface{}, 0, len(userIDs))
	for _, id := range userIDs {
		avatars = append(avatars, s.homeUserAvatar(id))
	}
	return avatars
}

func (s *Server) homeUserAvatar(userID int64) map[string]interface{} {
	name := s.displayName(userID, "\u7528\u6237")
	avatarURL := ""
	if user, ok := s.auth.UserByID(userID); ok {
		name = homeDisplayName(user, name)
		avatarURL = strings.TrimSpace(user.AvatarURL)
	}
	return map[string]interface{}{
		"userId":         userID,
		"name":           name,
		"imageUrl":       avatarURL,
		"avatarUrl":      avatarURL,
		"avatarFallback": avatarTextForName(name, userID),
	}
}

func (s *Server) homeEarth(userID int64, visibleGames []games.Game, nearbyGames []map[string]interface{}, conns []connections.Connection) map[string]interface{} {
	nodes := []map[string]interface{}{
		{"id": "user-" + strconv.FormatInt(userID, 10), "type": "self", "label": "\u6211", "weight": 5},
	}
	heat := make([]map[string]interface{}, 0)
	for _, game := range visibleGames {
		if game.Longitude == 0 && game.Latitude == 0 {
			continue
		}
		nodes = append(nodes, map[string]interface{}{
			"id":        "game-" + strconv.FormatInt(game.ID, 10),
			"type":      "game",
			"label":     game.Title,
			"longitude": game.Longitude,
			"latitude":  game.Latitude,
			"weight":    maxInt(1, game.CurrentPlayers),
		})
		heat = append(heat, map[string]interface{}{
			"cityCode":  game.CityCode,
			"cityName":  game.CityName,
			"address":   game.Address,
			"longitude": game.Longitude,
			"latitude":  game.Latitude,
			"weight":    maxInt(1, game.CurrentPlayers),
		})
	}
	for _, conn := range conns {
		name := s.displayName(conn.ConnectedUserID, "\u7528\u6237")
		nodes = append(nodes, map[string]interface{}{
			"id":       "user-" + strconv.FormatInt(conn.ConnectedUserID, 10),
			"type":     "relation",
			"label":    name,
			"avatar":   avatarTextForName(name, conn.ConnectedUserID),
			"strength": conn.StrengthScore,
		})
	}
	return map[string]interface{}{
		"onlineCount": len(nodes),
		"nearbyCount": len(nearbyGames),
		"nodes":       nodes,
		"heatPoints":  heat,
		"edges":       s.connectionEdges(userID, conns),
	}
}

func (s *Server) homeVisualization(userID int64, visibleGames []games.Game, nearbyGames []map[string]interface{}, conns []connections.Connection) map[string]interface{} {
	earth := s.homeEarth(userID, visibleGames, nearbyGames, conns)
	return map[string]interface{}{
		"earth":   earth,
		"network": s.connectionNetworkPayload(userID, conns),
	}
}

func homeDisplayName(user users.User, fallback string) string {
	if user.Nickname != "" {
		return user.Nickname
	}
	return fallback
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
