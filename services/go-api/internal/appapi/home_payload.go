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
	"zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

const homeDisplayConfigKey = "home.display_config"

type homeDisplayConfigDTO struct {
	OnlineBaseCount int    `json:"onlineBaseCount"`
	OnlineSuffix    string `json:"onlineSuffix"`
}

func (s *Server) buildAppHomePayload(userID int64, visibleGames []games.Game, requestedRoleType string, pointSummary points.Account, unreadNotificationCountValue int, onlineCount int) (map[string]interface{}, error) {
	if len(visibleGames) > 10 {
		visibleGames = visibleGames[:10]
	}
	roleType := normalizeHomeRoleType(requestedRoleType)
	if roleType == "" {
		roleType = s.homeRoleType(userID)
	}
	user, _ := s.auth.UserByID(userID)
	identityRecord, err := s.identity.StatusStrict(userID)
	if err != nil {
		return nil, err
	}
	currentUser, err := s.buildCurrentUserDTOStrict(user, identityRecord)
	if err != nil {
		return nil, err
	}
	growth, err := s.reviews.ProfileStrict(userID)
	if err != nil {
		return nil, err
	}
	roleGrowthProfiles := s.roleGrowthProfilesForUser(userID, growth)
	stats := s.games.StatsForUser(userID)
	conns, err := s.connections.MyStrict(userID)
	if err != nil {
		return nil, err
	}
	nearbyGames, err := s.homeNearbyGames(userID, visibleGames, 6)
	if err != nil {
		return nil, err
	}
	friendGames := s.homeFriendGames(userID, visibleGames, conns, 4)
	recommendedGames := s.homeGameCards(visibleGames, "city", 6)
	if len(recommendedGames) > 6 {
		recommendedGames = recommendedGames[:6]
	}
	nearbySummary, err := s.homeNearbySummary(userID, nearbyGames, conns)
	if err != nil {
		return nil, err
	}
	rankingBoards := s.homeRankingBoards(userID, user, stats, growth, conns)

	payload := map[string]interface{}{
		"hero":               s.homeHero(userID, user, growth, len(visibleGames), len(conns), roleType, onlineCount),
		"user":               currentUser,
		"playerSummary":      s.homePlayerSummary(userID, user, stats, growth, roleType, roleGrowthProfiles),
		"nearbySummary":      nearbySummary,
		"onlineCard":         s.homeOnlineCard(nearbySummary),
		"nearbySection":      s.homeNearbySection(),
		"nearbyGames":        nearbyGames,
		"recommendedGames":   recommendedGames,
		"friendSection":      s.homeFriendSection(len(friendGames)),
		"friendGames":        friendGames,
		"rankingSection":     s.homeRankingSection(),
		"rankingBoards":      rankingBoards,
		"achievementSection": s.homeAchievementSection(),
		"achievements":       s.homeAchievements(stats, growth, nearbyGames),
		"metaverseEntry":     s.homeMetaverseEntry(len(conns), len(visibleGames), onlineCount),
		"earth":              s.homeEarth(userID, visibleGames, nearbyGames, conns, onlineCount),
		"visualization":      s.homeVisualization(userID, visibleGames, nearbyGames, conns, onlineCount),
		"games":              visibleGames,
		"userStats":          stats,
		"pointsSummary":      pointSummary,
		"growth":             growth,
		"notifications":      map[string]interface{}{"unreadCount": unreadNotificationCountValue},
		"expertBlueBadge":    s.expertBlueBadgeForUser(userID),
	}
	if roleType == "expert" {
		payload["skills"] = s.homeExpertSkills(userID)
		payload["skillDisplay"] = s.currentExpertSkillDisplayConfig()
	}
	if roleType == "guide" {
		payload["network"] = s.homeRoleNetwork(userID, conns)
	}
	return payload, nil
}

func (s *Server) homePendingGameCards(userID int64, limit int) []map[string]interface{} {
	allGames := s.games.List()
	items := make([]games.Game, 0)
	for index := len(allGames) - 1; index >= 0; index-- {
		game := allGames[index]
		if game.CreatorUserID == userID && game.Status == "pending_audit" {
			items = append(items, game)
			if limit > 0 && len(items) >= limit {
				break
			}
		}
	}
	return s.homeGameCards(items, "mine", limit)
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

func (s *Server) homeHero(userID int64, user users.User, growth reviews.GrowthProfile, gameCount int, connectionCount int, roleType string, onlineCount int) map[string]interface{} {
	return map[string]interface{}{
		"onlineText":  s.homeOnlineText(gameCount, connectionCount, onlineCount),
		"roleType":    roleType,
		"roleName":    homeRoleName(roleType),
		"currentRole": roleType,
		"dateLabel":   time.Now().Format("2006.01.02"),
		"subtitle":    "\u4eca\u65e5\u63a8\u8350 " + strconv.Itoa(gameCount) + " \u4e2a\u7ec4\u5c40",
		"nickname":    homeDisplayName(user, s.displayName(userID, "\u7528\u6237")),
		"level":       growth.Level,
	}
}

func (s *Server) homeOnlineText(gameCount int, connectionCount int, activeCount int) string {
	config := s.currentHomeDisplayConfig()
	suffix := strings.TrimSpace(config.OnlineSuffix)
	if suffix == "" {
		suffix = "\u4eba\u5728\u7ebf"
	}
	return strconv.Itoa(maxInt(0, activeCount)) + suffix
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

func (s *Server) adminHomeDisplayConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentHomeDisplayConfig()})
	case http.MethodPut:
		var req homeDisplayConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "首页展示配置格式错误")
			return
		}
		config, err := normalizeHomeDisplayConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig != nil {
			if err := s.systemConfig.Set(homeDisplayConfigKey, config); err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存首页展示配置失败")
				return
			}
		}
		s.recordOperation(r, "home_display_config:update", "system_config", "home_display_config", map[string]interface{}{
			"onlineBaseCount": config.OnlineBaseCount,
			"onlineSuffix":    config.OnlineSuffix,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentHomeDisplayConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "不支持当前请求方式")
	}
}

func normalizeHomeDisplayConfig(req homeDisplayConfigDTO) (homeDisplayConfigDTO, error) {
	if req.OnlineBaseCount < 0 {
		return homeDisplayConfigDTO{}, errors.New("在线人数基础值不能小于 0")
	}
	config := req
	config.OnlineSuffix = strings.TrimSpace(req.OnlineSuffix)
	if config.OnlineSuffix == "" {
		config.OnlineSuffix = "\u4eba\u5728\u7ebf"
	}
	if len([]rune(config.OnlineSuffix)) > 12 {
		return homeDisplayConfigDTO{}, errors.New("在线人数后缀最多 12 个字")
	}
	return config, nil
}

func defaultHomeDisplayConfig() homeDisplayConfigDTO {
	return homeDisplayConfigDTO{
		OnlineBaseCount: 0,
		OnlineSuffix:    "\u4eba\u5728\u7ebf",
	}
}

func (s *Server) homeRoleType(userID int64) string {
	if s.userHasActiveRole(userID, "guide") {
		return "guide"
	}
	if s.userHasActiveRole(userID, "expert") {
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

func normalizeHomeRoleType(roleType string) string {
	switch strings.ToLower(strings.TrimSpace(roleType)) {
	case "player", "expert", "guide":
		return strings.ToLower(strings.TrimSpace(roleType))
	default:
		return ""
	}
}

func homeRoleGrowthProfile(roleType string, profiles []roleGrowthProfileDTO) roleGrowthProfileDTO {
	for _, item := range profiles {
		if item.RoleCode == roleType {
			return item
		}
	}
	return roleGrowthProfileDTO{RoleCode: roleType, RoleName: homeRoleName(roleType)}
}

func (s *Server) homePlayerSummary(userID int64, user users.User, stats games.UserStats, growth reviews.GrowthProfile, roleType string, profiles []roleGrowthProfileDTO) map[string]interface{} {
	roleGrowth := homeRoleGrowthProfile(roleType, profiles)
	roleName := homeRoleName(roleType)
	levelTitle := strings.TrimSpace(roleGrowth.LevelTitle)
	if levelTitle == "" {
		levelTitle = roleName
	}
	levelBadge, levelName := roleGrowthLevelDisplay(roleType, roleGrowth.LevelNo, levelTitle)

	// 玩家按累计经验升级；行家和领路人按后台配置的真实指标得分和等级
	// 升级。首页不能再套用统一的 Lv.X / XP 口径。
	if roleType != "player" {
		progress := roleGrowth.Score
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
		return map[string]interface{}{
			"currentRole":       roleType,
			"roleType":          roleType,
			"roleLabel":         roleName + " · " + levelTitle,
			"levelBadge":        levelBadge,
			"levelTitle":        levelName,
			"displayName":       homeDisplayName(user, s.displayName(userID, "用户")),
			"level":             roleGrowth.LevelNo,
			"experience":        roleGrowth.Score,
			"xpText":            roleGrowth.ScoreLabel + " " + strconv.Itoa(roleGrowth.Score) + "分",
			"scoreText":         strconv.Itoa(growth.CreditScore) + "信用分",
			"nextLevelText":     roleGrowth.Description,
			"progressPercent":   progress,
			"joinCount":         stats.Participated,
			"monthlyMvpCount":   len(growth.Achievements),
			"participationRate": "--",
			"stats": []map[string]interface{}{
				{"value": roleGrowth.Score, "label": roleGrowth.ScoreLabel},
				{"value": growth.CreditScore, "label": "信用分"},
				{"value": stats.Completed, "label": "完成局数"},
			},
		}
	}
	// 玩家等级门槛由后台维护，首页必须使用同一套梯度，不能再按每级
	// 固定 100 经验推算，否则修改配置后首页会显示错误的下一等级进度。
	playerExperience := roleGrowth.Score
	nextLevelExperience, expToNext, progress := s.playerLevelProgress(playerExperience, roleGrowth.LevelNo)
	joinCount := stats.Participated
	participationRate := "0%"
	if joinCount > 0 {
		participationRate = "100%"
	}
	return map[string]interface{}{
		"currentRole":         roleType,
		"roleType":            roleType,
		"roleLabel":           roleName + " · " + levelTitle,
		"levelBadge":          levelBadge,
		"levelTitle":          levelName,
		"displayName":         homeDisplayName(user, s.displayName(userID, "\u7528\u6237")),
		"level":               roleGrowth.LevelNo,
		"experience":          playerExperience,
		"xpText":              strconv.Itoa(playerExperience) + "经验",
		"scoreText":           strconv.Itoa(growth.CreditScore) + "\u4fe1\u7528\u5206",
		"nextLevelExperience": nextLevelExperience,
		"expToNextLevel":      expToNext,
		"nextLevelText":       roleGrowth.Description,
		"progressPercent":     progress,
		"joinCount":           joinCount,
		"monthlyMvpCount":     len(growth.Achievements),
		"participationRate":   participationRate,
		"stats": []map[string]interface{}{
			{"value": joinCount, "label": "\u53c2\u4e0e\u5c40\u6570"},
			{"value": stats.Completed, "label": "\u5b8c\u6210\u5c40\u6570"},
			{"value": playerExperience, "label": "\u7ecf\u9a8c"},
		},
	}
}

// roleGrowthLevelDisplay keeps the operator-configured level title as the
// source of truth while separating the compact home-card badge from its
// descriptive title.
func roleGrowthLevelDisplay(roleType string, levelNo int, configuredTitle string) (string, string) {
	title := strings.TrimSpace(configuredTitle)
	badge := ""
	switch roleType {
	case "expert":
		badge = strconv.Itoa(levelNo) + "星"
	case "guide":
		badge = strconv.Itoa(levelNo) + "级"
	default:
		badge = "Lv" + strconv.Itoa(levelNo)
	}
	for _, prefix := range []string{badge + " ", badge + "　", badge} {
		if strings.HasPrefix(title, prefix) {
			if name := strings.TrimSpace(strings.TrimPrefix(title, prefix)); name != "" {
				return badge, name
			}
			break
		}
	}
	return badge, title
}

func (s *Server) playerLevelProgress(experience int, levelNo int) (nextExperience int, remaining int, percent int) {
	if experience < 0 {
		experience = 0
	}
	currentThreshold := 0
	for _, item := range s.currentRoleLevelConfig().Items {
		if item.RoleCode != "player" || !item.Enabled {
			continue
		}
		if item.LevelNo == levelNo {
			currentThreshold = item.MinExperience
		}
		if item.MinExperience > experience && (nextExperience == 0 || item.MinExperience < nextExperience) {
			nextExperience = item.MinExperience
		}
	}
	if nextExperience == 0 {
		return 0, 0, 100
	}
	remaining = nextExperience - experience
	if remaining < 0 {
		remaining = 0
	}
	span := nextExperience - currentThreshold
	if span <= 0 {
		return nextExperience, remaining, 0
	}
	percent = int(math.Round(float64(experience-currentThreshold) / float64(span) * 100))
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return nextExperience, remaining, percent
}

func (s *Server) homeExpertSkills(userID int64) []map[string]interface{} {
	config := s.systemSkillConfig(userID)
	groups, _ := config["skillGroups"].(map[string]interface{})
	visible := systemSkillItemList(groups["visible"])
	limit := s.currentExpertSkillDisplayConfig().VisibleSkillLimit
	items := make([]map[string]interface{}, 0, minInt(limit, len(visible)))
	tones := []string{"cyan", "green", "purple", "orange", "blue"}
	icons := []string{"✓", "★", "◆", "✦", "✚"}
	for index, skill := range visible {
		if index >= limit {
			break
		}
		title := strings.TrimSpace(firstNonEmptyString(skill, "title", "name", "label"))
		if title == "" {
			continue
		}
		tone := strings.TrimSpace(firstNonEmptyString(skill, "tone"))
		if tone == "" {
			tone = tones[index%len(tones)]
		}
		icon := strings.TrimSpace(firstNonEmptyString(skill, "iconText", "icon"))
		if icon == "" {
			icon = icons[index%len(icons)]
		}
		items = append(items, map[string]interface{}{
			"key":       firstNonEmptyString(skill, "id", "key"),
			"title":     title,
			"icon":      icon,
			"tone":      tone,
			"nodeStyle": firstNonEmptyString(skill, "nodeStyle"),
			"locked":    false,
		})
	}
	return items
}

func (s *Server) homeRoleNetwork(userID int64, conns []connections.Connection) map[string]interface{} {
	relationCount := len(conns)
	items := s.homeGuideIndustryItems(userID, relationCount)
	return map[string]interface{}{
		"title":        "我的关系领域",
		"status":       "实时连接中",
		"hubTitle":     s.displayName(userID, "领路人"),
		"hubDesc":      "领路人",
		"summary":      "已连接 " + strconv.Itoa(relationCount) + " 位玩家",
		"location":     "核心区",
		"locationText": "核心区",
		"items":        items,
		"buttons": []map[string]interface{}{
			{"text": "管理我的连接", "primary": true, "route": "pages/relation/network/index"},
		},
		"userId": userID,
	}
}

func (s *Server) homeGuideIndustryItems(userID int64, relationCount int) []map[string]interface{} {
	profile := s.profiles.AdminGuideResource(userID)
	industries := append([]string(nil), profile.IndustryTags...)
	audiences := append([]string(nil), profile.ResourceTags...)
	city := ""
	if len(profile.CityCodes) > 0 {
		city = profile.CityCodes[0]
	}
	if len(industries) == 0 {
		applications := s.profiles.RoleApplicationsByUser(userID)
		for index := len(applications) - 1; index >= 0; index-- {
			application := applications[index]
			if application.RoleCode != "guide" {
				continue
			}
			industries, audiences, city = parseGuideApplicationIndustries(application.AbilityDescription)
			if len(industries) > 0 {
				break
			}
		}
	}
	items := make([]map[string]interface{}, 0, minInt(4, len(industries))+1)
	for index, industry := range industries {
		if index >= 4 {
			break
		}
		desc := "服务领域"
		if len(audiences) > 0 {
			desc = audiences[index%len(audiences)]
		} else if city != "" {
			desc = city
		}
		items = append(items, map[string]interface{}{
			"id": "industry-" + strconv.Itoa(index+1), "icon": guideIndustryIcon(industry), "name": industry, "desc": desc,
		})
	}
	if len(items) == 0 {
		items = append(items, map[string]interface{}{
			"id": "industry-pending", "icon": "🌐", "name": "领域待完善", "desc": "请完善领路人行业",
		})
	}
	items = append(items, map[string]interface{}{
		"id": "all", "icon": "+", "name": "查看全部", "desc": strconv.Itoa(relationCount) + "人", "dashed": true,
	})
	return items
}

func parseGuideApplicationIndustries(description string) ([]string, []string, string) {
	industries := []string{}
	audiences := []string{}
	city := ""
	for _, line := range strings.Split(description, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "所在城市："):
			city = strings.TrimSpace(strings.TrimPrefix(line, "所在城市："))
		case strings.HasPrefix(line, "可推荐人群："):
			audiences = splitGuideApplicationValues(strings.TrimPrefix(line, "可推荐人群："), "、")
		case strings.HasPrefix(line, "业务说明："):
			for _, service := range strings.Split(strings.TrimPrefix(line, "业务说明："), "；") {
				if separator := strings.Index(service, "："); separator >= 0 {
					service = service[separator+len("："):]
				}
				if separator := strings.Index(service, "，"); separator >= 0 {
					service = service[:separator]
				}
				if service = strings.TrimSpace(service); service != "" && service != "未命名" {
					industries = append(industries, service)
				}
			}
		}
	}
	return industries, audiences, city
}

func splitGuideApplicationValues(value string, separator string) []string {
	items := []string{}
	for _, item := range strings.Split(value, separator) {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func guideIndustryIcon(industry string) string {
	switch {
	case strings.Contains(industry, "剧本"), strings.Contains(industry, "桌游"), strings.Contains(industry, "游戏"):
		return "👑"
	case strings.Contains(industry, "教育"), strings.Contains(industry, "培训"), strings.Contains(industry, "学生"), strings.Contains(strings.ToUpper(industry), "DM"):
		return "🎓"
	case strings.Contains(industry, "母婴"), strings.Contains(industry, "亲子"), strings.Contains(industry, "宝妈"), strings.Contains(industry, "儿童"):
		return "👶"
	case strings.Contains(industry, "科研"), strings.Contains(industry, "研究"), strings.Contains(industry, "技术"), strings.Contains(industry, "AI"):
		return "🔬"
	case strings.Contains(industry, "金融"), strings.Contains(industry, "投资"):
		return "💰"
	case strings.Contains(industry, "健康"), strings.Contains(industry, "医疗"):
		return "⚕️"
	default:
		return "🌐"
	}
}

func (s *Server) homeNearbySummary(userID int64, nearbyGames []map[string]interface{}, conns []connections.Connection) (map[string]interface{}, error) {
	recent, err := s.lbs.RecentStrict(userID, 20)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"nearbyGameCount": len(nearbyGames),
		"checkedInCount":  len(recent),
		"relationCount":   len(conns),
		"onlineNodeCount": maxInt(1, len(nearbyGames)+len(conns)+1),
	}, nil
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
		"route":    "pages/game/hall/index",
	}
}

func (s *Server) homeNearbyGames(userID int64, visibleGames []games.Game, limit int) ([]map[string]interface{}, error) {
	location, ok, err := s.lbs.CurrentStrict(userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		// 首页加载不触发设备定位。没有保存定位时只给同城/发布时间
		// 兜底卡片，真实“附近”查询由用户点击后再发起。
		return nil, nil
	}
	items := make([]games.Game, 0, len(visibleGames))
	for _, game := range visibleGames {
		if game.Longitude == 0 && game.Latitude == 0 {
			continue
		}
		distance := lbs.DistanceMeter(location.Longitude, location.Latitude, game.Longitude, game.Latitude)
		game.DistanceMeter = math.Round(distance)
		game.DistanceLabel = lbs.FormatDistanceLabel(game.DistanceMeter)
		items = append(items, game)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].DistanceMeter < items[j].DistanceMeter
	})
	return s.homeGameCards(items, "nearby", limit), nil
}

func (s *Server) homeFriendGames(userID int64, visibleGames []games.Game, conns []connections.Connection, limit int) []map[string]interface{} {
	connected := make(map[int64]bool, len(conns))
	for _, conn := range conns {
		connected[conn.ConnectedUserID] = true
	}
	items := make([]games.Game, 0)
	for _, game := range visibleGames {
		if game.CreatorUserID == userID || game.MainGuideUserID == userID || s.games.IsMember(game.ID, userID) {
			continue
		}
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
			"id":             game.ID,
			"route":          "pages/game/detail/index?id=" + strconv.FormatInt(game.ID, 10),
			"scope":          scope,
			"title":          game.Title,
			"typeText":       homeGameCategoryText(game),
			"statusText":     homeGameStatusText(game.Status),
			"coverSrc":       homeGameCover(game, index),
			"priceText":      homeGamePriceText(game.GameType),
			"actionText":     homeGameActionText(game.Status),
			"cityName":       game.CityName,
			"address":        game.Address,
			"distanceText":   game.DistanceLabel,
			"memberText":     strconv.Itoa(game.CurrentPlayers) + "/" + strconv.Itoa(game.MaxPlayers) + "\u4eba",
			"timeText":       game.CreatedAt.Format("2006-01-02 15:04"),
			"joinedCount":    game.CurrentPlayers,
			"joinedText":     "+" + strconv.Itoa(game.CurrentPlayers) + "\u4f4d\u73a9\u5bb6\u5df2\u5165\u5c40",
			"playerAvatars":  s.gameListPlayerAvatars(game),
			"actions":        s.currentGameCategoryConfig().EventActions,
			"shareComponent": s.currentGameCategoryConfig().ShareComponent,
			"longitude":      game.Longitude,
			"latitude":       game.Latitude,
			"creatorUserId":  game.CreatorUserID,
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

func homeGameCategoryText(game games.Game) string {
	if text := strings.TrimSpace(game.PrimaryCategoryText); text != "" {
		return text
	}
	switch strings.TrimSpace(game.PrimaryCategory) {
	case "social":
		return "社交局"
	case "task":
		return "任务局"
	case "explore":
		return "探索局"
	case "growth":
		return "成长局"
	default:
		return homeGameTypeText(game.GameType)
	}
}

func homeGameStatusText(status string) string {
	return games.StatusText(status)
}

func homeGamePriceText(gameType string) string {
	if gameType == "" || gameType == "free" {
		return "\u514d\u8d39"
	}
	return "\u540e\u53f0\u5f00\u5c40"
}

func homeGameActionText(status string) string {
	switch status {
	case "recruiting":
		return "\u62db\u52df\u4e2d"
	case "full":
		return "\u5df2\u6ee1\u5458"
	case "in_progress":
		return "\u5df2\u5f00\u5c40"
	default:
		return homeGameStatusText(status)
	}
}

func homeGameCover(game games.Game, index int) string {
	if coverURL := strings.TrimSpace(game.CoverImage); coverURL != "" {
		return coverURL
	}
	covers := []string{
		"/components/game-card/assets/cover-city.png",
		"/components/game-card/assets/cover-sunset.png",
	}
	return covers[index%len(covers)]
}

func (s *Server) homeRankingSection() map[string]interface{} {
	return map[string]interface{}{
		"icon":     "\U0001f3c6",
		"title":    "\u672c\u5468\u73a9\u9738\u699c",
		"desc":     "\u6309\u53c2\u4e0e\u5c40\u6570\u3001\u5b8c\u6210\u5c40\u6570\u548c\u7ecf\u9a8c\u503c\u7efc\u5408\u6392\u5e8f",
		"moreText": "\u67e5\u770b\u5168\u90e8\u699c\u5355",
		"route":    "pages/home/ranking/index",
		"tabs": []map[string]string{
			{"key": "player", "name": "\u73a9\u5bb6"},
			{"key": "expert", "name": "\u884c\u5bb6"},
			{"key": "guide", "name": "\u9886\u8def\u4eba"},
		},
	}
}

func (s *Server) homeRankingBoards(userID int64, user users.User, stats games.UserStats, growth reviews.GrowthProfile, conns []connections.Connection) map[string]interface{} {
	usersForRanking, err := s.auth.AdminUsers(users.Filter{})
	if err != nil || len(usersForRanking) == 0 {
		usersForRanking = []users.User{user}
		for _, conn := range conns {
			if connUser, ok := s.auth.UserByID(conn.ConnectedUserID); ok {
				usersForRanking = append(usersForRanking, connUser)
			}
		}
	}

	return map[string]interface{}{
		"player": s.homeRankingBoardForRole(userID, usersForRanking, ""),
		"expert": s.homeRankingBoardForRole(userID, usersForRanking, "expert"),
		"guide":  s.homeRankingBoardForRole(userID, usersForRanking, "guide"),
	}
}

func (s *Server) homeRankingBoardForRole(currentUserID int64, usersForRanking []users.User, roleType string) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(usersForRanking))
	myRank := map[string]interface{}{}
	rankInput := make([]struct {
		user   users.User
		stats  games.UserStats
		growth reviews.GrowthProfile
		score  int
	}, 0, len(usersForRanking))

	for _, item := range usersForRanking {
		if item.ID <= 0 {
			continue
		}
		if roleType != "" && !s.homeUserHasActiveRole(item.ID, roleType) {
			continue
		}
		itemStats := s.games.StatsForUser(item.ID)
		itemGrowth := s.reviews.Profile(item.ID)
		rankInput = append(rankInput, struct {
			user   users.User
			stats  games.UserStats
			growth reviews.GrowthProfile
			score  int
		}{
			user:   item,
			stats:  itemStats,
			growth: itemGrowth,
			score:  homeRankingScore(itemStats, itemGrowth),
		})
	}

	sort.Slice(rankInput, func(i, j int) bool {
		if rankInput[i].score == rankInput[j].score {
			return rankInput[i].user.ID < rankInput[j].user.ID
		}
		return rankInput[i].score > rankInput[j].score
	})

	for index, item := range rankInput {
		rankingItem := s.homeRankingItem(index+1, item.user.ID, homeDisplayName(item.user, s.displayName(item.user.ID, "\u7528\u6237")), item.stats, item.growth, item.user.ID == currentUserID)
		items = append(items, rankingItem)
		if item.user.ID == currentUserID {
			myRank = rankingItem
		}
	}

	return map[string]interface{}{"list": items, "myRank": myRank}
}

func (s *Server) homeUserHasActiveRole(userID int64, roleType string) bool {
	status := s.profiles.RoleSnapshot(userID).RoleStatusMap[roleType]
	return status == "active" || status == "approved"
}

func homeRankingScore(stats games.UserStats, growth reviews.GrowthProfile) int {
	return growth.Experience + stats.Completed*30 + stats.Participated*10
}

func (s *Server) homeRankingItem(rank int, userID int64, name string, stats games.UserStats, growth reviews.GrowthProfile, isMe bool) map[string]interface{} {
	score := homeRankingScore(stats, growth)
	return map[string]interface{}{
		"userId":          userID,
		"rank":            rank,
		"avatarFallback":  avatarTextForName(name, userID),
		"name":            name,
		"desc":            "\u53c2\u4e0e" + strconv.Itoa(stats.Participated) + "\u5c40 \u00b7 \u5b8c\u6210" + strconv.Itoa(stats.Completed) + "\u5c40",
		"gameCount":       stats.Participated,
		"weeklyGameCount": stats.Participated,
		"weeklyMvpCount":  len(growth.Achievements),
		"experience":      growth.Experience,
		"xp":              score,
		"xpUnit":          "\u5206",
		"isMe":            isMe,
	}
}

func (s *Server) homeAchievementSection() map[string]interface{} {
	return map[string]interface{}{
		"icon":  "\u25c6",
		"title": "\u6211\u7684\u6210\u5c31",
	}
}

func (s *Server) homeAchievements(stats games.UserStats, growth reviews.GrowthProfile, nearbyGames []map[string]interface{}) []map[string]interface{} {
	earthProgress := maxInt(0, minInt(100, len(nearbyGames)*12))
	config := s.currentGrowthAchievementConfig()
	configured := append(append([]growthAchievementItemDTO{}, config.Catalog...), config.Locked...)
	if len(configured) > 0 {
		unlockedCodes := make(map[string]bool, len(growth.Achievements))
		for _, code := range growth.Achievements {
			unlockedCodes[code] = true
		}
		items := make([]map[string]interface{}, 0, len(configured))
		for _, item := range configured {
			if !item.Visible {
				continue
			}
			code := strings.TrimSpace(item.Code)
			unlocked := unlockedCodes[code]
			progress := item.ProgressPercent
			switch code {
			case "first_game":
				progress = minInt(100, stats.Participated*100)
				unlocked = unlocked || stats.Participated > 0
			case "credit_keeper":
				progress = minInt(100, growth.CreditScore)
				unlocked = unlocked || growth.CreditScore >= 80
			case "earth":
				progress = earthProgress
				unlocked = unlocked || earthProgress > 0
			}
			status := item.StatusText
			if status == "" {
				status = map[bool]string{true: "已解锁", false: "进行中"}[unlocked]
			}
			items = append(items, map[string]interface{}{
				"id": item.ID, "title": item.Title, "status": status,
				"unlocked": unlocked, "progressPercent": progress,
			})
		}
		if len(items) > 0 {
			return items
		}
	}
	return []map[string]interface{}{
		{"id": "first_game", "title": "\u9996\u5c40\u8fbe\u6210", "status": "\u8fdb\u5ea6" + strconv.Itoa(minInt(100, stats.Participated*100)) + "%", "unlocked": stats.Participated > 0, "progressPercent": minInt(100, stats.Participated*100)},
		{"id": "credit_keeper", "title": "\u4fe1\u7528\u5b88\u62a4", "status": strconv.Itoa(growth.CreditScore) + "\u5206", "unlocked": growth.CreditScore >= 80},
		{"id": "earth", "title": "\u5730\u7403\u6f2b\u6e38\u8005", "status": "\u8fdb\u5ea6" + strconv.Itoa(earthProgress) + "%", "unlocked": earthProgress > 0, "progressPercent": earthProgress},
	}
}

func (s *Server) homeMetaverseEntry(connectionCount int, gameCount int, onlineCount int) map[string]interface{} {
	return map[string]interface{}{
		"title":             "\u8fdb\u5165\u5143\u5b87\u5b99",
		"description":       "\u57fa\u4e8e\u7ec4\u5c40\u548c\u4eba\u8109\u7684\u52a8\u6001\u5173\u7cfb\u7f51",
		"tags":              []string{"\u5173\u7cfb\u7f51", "\u5730\u56fe\u8282\u70b9"},
		"onlineCount":       maxInt(0, onlineCount),
		"relationNodeCount": connectionCount + gameCount,
	}
}

func (s *Server) homeEarth(userID int64, visibleGames []games.Game, nearbyGames []map[string]interface{}, conns []connections.Connection, onlineCount int) map[string]interface{} {
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
		"onlineCount": maxInt(0, onlineCount),
		"nodeCount":   len(nodes),
		"nearbyCount": len(nearbyGames),
		"nodes":       nodes,
		"heatPoints":  heat,
		"edges":       s.connectionEdges(userID, conns),
	}
}

func (s *Server) homeVisualization(userID int64, visibleGames []games.Game, nearbyGames []map[string]interface{}, conns []connections.Connection, onlineCount int) map[string]interface{} {
	earth := s.homeEarth(userID, visibleGames, nearbyGames, conns, onlineCount)
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
