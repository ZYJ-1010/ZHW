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
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

const homeDisplayConfigKey = "home.display_config"

type homeDisplayConfigDTO struct {
	OnlineBaseCount int    `json:"onlineBaseCount"`
	OnlineSuffix    string `json:"onlineSuffix"`
}

func (s *Server) buildAppHomePayload(userID int64, visibleGames []games.Game, requestedRoleType string) map[string]interface{} {
	if len(visibleGames) > 10 {
		visibleGames = visibleGames[:10]
	}
	roleType := normalizeHomeRoleType(requestedRoleType)
	if roleType == "" {
		roleType = s.homeRoleType(userID)
	}
	user, _ := s.auth.UserByID(userID)
	growth := s.reviews.Profile(userID)
	stats := s.games.StatsForUser(userID)
	conns := s.connections.My(userID)
	nearbyGames := s.homeNearbyGames(userID, visibleGames, 6)
	friendGames := s.homeFriendGames(userID, visibleGames, conns, 4)
	nearbySummary := s.homeNearbySummary(userID, nearbyGames, conns)
	rankingBoards := s.homeRankingBoards(userID, user, stats, growth, conns)

	payload := map[string]interface{}{
		"hero":               s.homeHero(userID, user, growth, len(visibleGames), len(conns), roleType),
		"user":               s.buildCurrentUserDTO(user, s.identity.Status(userID)),
		"playerSummary":      s.homePlayerSummary(userID, user, stats, growth, roleType),
		"nearbySummary":      nearbySummary,
		"onlineCard":         s.homeOnlineCard(nearbySummary),
		"nearbySection":      s.homeNearbySection(),
		"nearbyGames":        nearbyGames,
		"recommendedGames":   s.homeGameCards(visibleGames, "city", 6),
		"friendSection":      s.homeFriendSection(len(friendGames)),
		"friendGames":        friendGames,
		"rankingSection":     s.homeRankingSection(),
		"rankingBoards":      rankingBoards,
		"achievementSection": s.homeAchievementSection(),
		"achievements":       s.homeAchievements(stats, growth, nearbyGames),
		"metaverseEntry":     s.homeMetaverseEntry(len(conns), len(visibleGames)),
		"earth":              s.homeEarth(userID, visibleGames, nearbyGames, conns),
		"visualization":      s.homeVisualization(userID, visibleGames, nearbyGames, conns),
		"games":              visibleGames,
		"userStats":          stats,
		"pointsSummary":      s.points.Summary(userID),
		"growth":             growth,
		"notifications":      map[string]interface{}{"unreadCount": unreadNotificationCount(s.notices.List(userID))},
	}
	if roleType == "expert" {
		payload["skills"] = []map[string]interface{}{}
	}
	if roleType == "guide" {
		payload["network"] = s.homeRoleNetwork(conns)
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

func (s *Server) homeRoleType(userID int64) string {
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

func normalizeHomeRoleType(roleType string) string {
	switch strings.ToLower(strings.TrimSpace(roleType)) {
	case "player", "expert", "guide":
		return strings.ToLower(strings.TrimSpace(roleType))
	default:
		return ""
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
		progress = int(math.Round(float64(growth.Experience%100) / 100 * 100))
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
		"xpText":              strconv.Itoa(growth.Experience) + " XP",
		"scoreText":           strconv.Itoa(growth.CreditScore) + "\u4fe1\u7528\u5206",
		"nextLevelExperience": nextLevelExperience,
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

func (s *Server) homeRoleNetwork(conns []connections.Connection) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(conns))
	for index, conn := range conns {
		name := s.displayName(conn.ConnectedUserID, "用户")
		items = append(items, map[string]interface{}{
			"id":   strconv.FormatInt(conn.ConnectedUserID, 10),
			"key":  "connection-" + strconv.FormatInt(conn.ConnectedUserID, 10),
			"name": name,
			"desc": "关系强度 " + strconv.Itoa(conn.StrengthScore),
			"icon": avatarTextForName(name, conn.ConnectedUserID),
		})
		if index >= 5 {
			break
		}
	}
	return map[string]interface{}{
		"title":   "人脉网络",
		"items":   items,
		"buttons": []map[string]interface{}{},
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
		return s.homeGameCards(visibleGames, "city", limit)
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
	return s.homeGameCards(items, "nearby", limit)
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
		"icon":     "\u2605",
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

func (s *Server) homeAchievementSection() map[string]interface{} {
	return map[string]interface{}{
		"icon":  "\u25c6",
		"title": "\u6211\u7684\u6210\u5c31",
	}
}

func (s *Server) homeAchievements(stats games.UserStats, growth reviews.GrowthProfile, nearbyGames []map[string]interface{}) []map[string]interface{} {
	earthProgress := maxInt(0, minInt(100, len(nearbyGames)*12))
	return []map[string]interface{}{
		{"id": "first_game", "title": "\u9996\u5c40\u8fbe\u6210", "status": "\u8fdb\u5ea6" + strconv.Itoa(minInt(100, stats.Participated*100)) + "%", "unlocked": stats.Participated > 0, "progressPercent": minInt(100, stats.Participated*100)},
		{"id": "credit_keeper", "title": "\u4fe1\u7528\u5b88\u62a4", "status": strconv.Itoa(growth.CreditScore) + "\u5206", "unlocked": growth.CreditScore >= 80},
		{"id": "earth", "title": "\u5730\u7403\u6f2b\u6e38\u8005", "status": "\u8fdb\u5ea6" + strconv.Itoa(earthProgress) + "%", "unlocked": earthProgress > 0, "progressPercent": earthProgress},
	}
}

func (s *Server) homeMetaverseEntry(connectionCount int, gameCount int) map[string]interface{} {
	return map[string]interface{}{
		"title":       "\u8fdb\u5165\u5143\u5b87\u5b99",
		"description": "\u57fa\u4e8e\u7ec4\u5c40\u548c\u4eba\u8109\u7684\u52a8\u6001\u5173\u7cfb\u7f51",
		"tags":        []string{"\u5173\u7cfb\u7f51", "\u5730\u56fe\u8282\u70b9"},
		"onlineCount": maxInt(1, connectionCount+gameCount),
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
