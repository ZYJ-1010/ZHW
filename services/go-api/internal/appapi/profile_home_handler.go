package appapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/membership"
	pointspkg "zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/reports"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

type profileHomeData struct {
	Connections      []connections.Connection
	InviteRelations  []invites.Relation
	Membership       membership.Membership
	InProgressGames  int
	ManagedGames     int
	Stats            games.UserStats
	RecentFootprints []reviews.Footprint
	GuideState       map[string]interface{}
}

func (s *Server) profileHome(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}

	user, _ := s.auth.CurrentUser(bearerToken(r.Header.Get("Authorization")))
	record, err := s.identity.StatusStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，请稍后重试")
		return
	}
	current, err := s.buildCurrentUserDTOStrict(user, record)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取个人资料失败，请稍后重试")
		return
	}
	todos, err := s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	pointSummary, _, loaded := s.loadPointsData(w, userID, false)
	if !loaded {
		return
	}
	notificationItems, err := s.notices.ListStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取个人中心消息提醒失败，请稍后重试")
		return
	}

	reportMessageCount, err := s.profileReportMessageCountStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报提醒失败，请稍后重试")
		return
	}
	onlineCount, err := s.auth.ActiveAppSessionCountStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取在线人数失败，请稍后重试")
		return
	}
	data, err := s.profileHomeDataStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取个人中心数据失败，请稍后重试")
		return
	}
	httpx.OK(w, s.buildProfileHomePayloadWithPoints(userID, user, current, len(todos), pointSummary, unreadNotificationCount(notificationItems), reportMessageCount, onlineCount, r.URL.Query().Get("roleType"), data))
}

func (s *Server) profileHomeDataStrict(userID int64) (profileHomeData, error) {
	conns, err := s.connections.MyStrict(userID)
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取关系网络失败: %w", err)
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取邀请关系失败: %w", err)
	}
	membershipInfo, err := s.membership.MyStrict(userID)
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取会员信息失败: %w", err)
	}
	gameItems, err := s.games.ListStrict()
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取组局数据失败: %w", err)
	}
	inProgress := 0
	for _, game := range gameItems {
		if game.Status != "in_progress" {
			continue
		}
		member, err := s.games.IsMemberStrict(game.ID, userID)
		if err != nil {
			return profileHomeData{}, fmt.Errorf("读取组局成员失败: %w", err)
		}
		if game.CreatorUserID == userID || game.MainGuideUserID == userID || member {
			inProgress++
		}
	}
	stats, err := s.games.StatsForUserStrict(userID)
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取组局统计失败: %w", err)
	}
	footprints, err := s.reviews.FootprintsStrict(userID)
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取成长足迹失败: %w", err)
	}
	guideState, err := s.newbieGuideStateStrict(userID, nil)
	if err != nil {
		return profileHomeData{}, fmt.Errorf("读取新手引导状态失败: %w", err)
	}
	return profileHomeData{
		Connections: conns, InviteRelations: relations, Membership: membershipInfo,
		InProgressGames: inProgress, ManagedGames: countManagedUserGames(userID, gameItems),
		Stats: stats, RecentFootprints: limitedFootprints(footprints, 5), GuideState: guideState,
	}, nil
}

func (s *Server) buildProfileHomePayload(userID int64, user users.User, current CurrentUserDTO, reviewTodoCount int, requestedRoles ...string) map[string]interface{} {
	requestedRole := ""
	if len(requestedRoles) > 0 {
		requestedRole = requestedRoles[0]
	}
	return s.buildProfileHomePayloadWithPoints(userID, user, current, reviewTodoCount, s.points.Summary(userID), unreadNotificationCount(s.notices.List(userID)), s.profileReportMessageCount(userID), s.auth.ActiveAppSessionCount(), requestedRole)
}

func (s *Server) buildProfileHomePayloadWithPoints(userID int64, user users.User, current CurrentUserDTO, reviewTodoCount int, pointSummary pointspkg.Account, notifications int, reportMessageCount int, onlineCount int, roleRequest string, dataItems ...profileHomeData) map[string]interface{} {
	var data profileHomeData
	if len(dataItems) > 0 {
		data = dataItems[0]
	} else {
		data = profileHomeData{
			Connections: s.connections.My(userID), Membership: s.membership.My(userID),
			InProgressGames: countInProgressUserGames(userID, s.games.List(), s.games.IsMember),
			ManagedGames:    countManagedUserGames(userID, s.games.List()), Stats: s.games.StatsForUser(userID),
			RecentFootprints: limitedFootprints(s.reviews.Footprints(userID), 5),
		}
		data.InviteRelations, _ = s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	}
	conns := data.Connections
	inviteRelations := data.InviteRelations
	invitedConns := connectionsForInvitees(conns, inviteRelations)
	membership := data.Membership
	inProgressGameCount := data.InProgressGames
	managedGameCount := data.ManagedGames
	roleGrowth := s.roleGrowthProfilesForUser(userID, reviews.GrowthProfile{
		UserID: userID, Level: current.Growth.Level, Experience: current.Growth.ExperienceValue,
		CreditScore: current.Growth.CreditScore, TodayCreditScore: current.Growth.TodayCreditScore,
	})
	roleType := ""
	if roleRequest != "" {
		roleType = normalizeHomeRoleType(roleRequest)
	}
	if roleType == "" || (roleType != "player" && !s.userHasActiveRole(userID, roleType)) {
		roleType = s.homeRoleType(userID)
	}
	role := homeRoleName(roleType)
	growthLevel := "Lv0 新玩家"
	for _, profile := range roleGrowth {
		if profile.RoleCode == roleType && profile.Active && strings.TrimSpace(profile.LevelTitle) != "" {
			growthLevel = profile.LevelTitle
			break
		}
	}
	memberLevel := membership.PlanName
	memberStatus := strings.ToLower(strings.TrimSpace(membership.Status))
	if memberLevel == "" || memberLevel == "none" || memberStatus != "active" {
		memberLevel = ""
	}
	name := homeDisplayName(user, s.displayName(userID, "用户"))
	avatarFileID, avatarURL := s.currentUserAvatar(userID)

	guideState := data.GuideState
	if guideState == nil {
		guideState = s.newbieGuideState(userID, nil)
	}
	persistentProfileReminder, _ := guideState["persistentProfileReminder"].(bool)
	serviceSections := s.profileHomeSections(reviewTodoCount, reportMessageCount, len(inviteRelations), inProgressGameCount, managedGameCount, persistentProfileReminder)
	if !s.userCanGenerateRegistrationInvitations(userID) {
		serviceSections = removeProfileHomeItem(serviceSections, "invite")
	}

	assetItems := []map[string]interface{}{
		{"key": "points", "label": "可用积分", "value": strconv.Itoa(pointSummary.AvailablePoints), "tone": "green"},
		{"key": "experience", "label": "累计经验", "value": strconv.Itoa(current.Growth.ExperienceValue), "tone": "orange"},
		{"key": "credit", "label": "信用分", "value": strconv.Itoa(current.Growth.CreditScore)},
	}

	return map[string]interface{}{
		"onlineText": s.homeOnlineText(0, 0, onlineCount),
		"user": map[string]interface{}{
			"nickname":     name,
			"avatarText":   avatarTextForName(name, userID),
			"avatarUrl":    avatarURL,
			"avatarFileId": avatarFileID,
			"memberLevel":  memberLevel,
			"memberStatus": memberStatus,
			"roleLevel":    growthLevel,
			"growthLevel":  growthLevel,
			"role":         role,
			"inviteCode":   current.InviteCode,
		},
		"stats": []map[string]interface{}{
			{"key": "referrals", "label": "引荐数", "value": strconv.Itoa(len(inviteRelations))},
			{"key": "successes", "label": "成功数", "value": strconv.Itoa(countStrongConnections(invitedConns))},
			{"key": "completedGames", "label": "完成局数", "value": strconv.Itoa(data.Stats.Completed)},
			{"key": "credit", "label": "信用度", "value": strconv.Itoa(current.Growth.CreditScore)},
		},
		"assets": assetItems,
		"assetSummary": map[string]interface{}{
			"points":      pointSummary.AvailablePoints,
			"experience":  current.Growth.ExperienceValue,
			"creditScore": current.Growth.CreditScore,
		},
		// 一期不展示会员购买或增值入口；角色申请由服务区的身份入口承接。
		"vipBanner":       map[string]interface{}{"visible": false},
		"serviceSections": serviceSections,
		"badges": map[string]interface{}{
			"reviewTodoCount":         reviewTodoCount,
			"unreadNotificationCount": notifications,
			"reportMessageCount":      reportMessageCount,
			"connectionCount":         len(conns),
		},
		"newbieGuide": guideState,
		"summary": CurrentUserSummaryDTO{
			User:                    current,
			ReviewTodoCount:         reviewTodoCount,
			UnreadNotificationCount: notifications,
			PointsSummary:           pointSummary,
			RecentFootprints:        data.RecentFootprints,
		},
	}
}

func removeProfileHomeItem(sections []map[string]interface{}, key string) []map[string]interface{} {
	for index := range sections {
		items, _ := sections[index]["items"].([]map[string]interface{})
		filtered := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			if item["key"] != key {
				filtered = append(filtered, item)
			}
		}
		sections[index]["items"] = filtered
	}
	return sections
}

func connectionsForInvitees(items []connections.Connection, relations []invites.Relation) []connections.Connection {
	invitees := make(map[int64]struct{}, len(relations))
	for _, relation := range relations {
		invitees[relation.InviteeUserID] = struct{}{}
	}
	result := make([]connections.Connection, 0, len(invitees))
	for _, item := range items {
		if _, ok := invitees[item.ConnectedUserID]; ok {
			result = append(result, item)
		}
	}
	return result
}

func (s *Server) profileHomeSections(reviewTodoCount int, unreadCount int, connectionCount int, inProgressGameCount int, managedServiceCount int, persistentProfileReminder bool) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"title": "服务中心",
			"items": []map[string]interface{}{
				profileHomeItem("myGames", "我的局", "/pages/profile/assets/i66@3x.png", "purple-blue", badgeText(inProgressGameCount, "进行中"), "pink", "/pages/profile/service-center/my-games/index"),
				profileHomeItem("gameManage", "发起与管理", "/pages/profile/assets/i69@3x.png", "purple-blue", badgeText(managedServiceCount, "个局"), "blue", "/pages/profile/service-center/my-games/index?category=created"),
				profileHomeItem("invite", "我的邀请", "/pages/profile/assets/i68@3x.png", "purple-blue", badgeText(connectionCount, "个关系"), "orange", "/pages/profile/service-center/invite/overview/index"),
				profileHomeItem("reviews", "评价管理", "/pages/profile/assets/i69@3x.png", "purple-blue", badgeText(reviewTodoCount, "待评价"), "pink", "/pages/profile/service-center/manage/review-manage/index"),
			},
		},
		{
			"title": "成长中心",
			"items": []map[string]interface{}{
				profileHomeItem("mall", "积分商城", "/pages/profile/assets/i72@3x.png", "orange", "", "", "/pages/profile/asset-center/mall/index"),
				profileHomeItem("points", "积分明细", "/pages/profile/assets/i73@3x.png", "orange", "", "", "/pages/profile/asset-center/points/index"),
				profileHomeItem("taskCenter", "任务中心", "/pages/profile/assets/i73@3x.png", "orange", "", "", "/pages/profile/task-center/index"),
			},
		},
		{
			"title": "足迹中心",
			"items": []map[string]interface{}{
				profileHomeItem("footprints", "我的足迹", "/pages/profile/assets/i75@3x.png", "teal", "", "", "/pages/profile/footprint/achievements/index"),
				profileHomeDisabledItem("cityStories", "我的城市故事", "/pages/profile/assets/i76@3x.png", "teal", "", "", "地图与城市探索将在后续版本开放"),
				profileHomeItem("achievements", "我的成就墙", "/pages/profile/assets/i77@3x.png", "teal", "", "", "/pages/profile/footprint/achievements/index"),
			},
		},
		{
			"title": "系统管理",
			"items": []map[string]interface{}{
				profileHomeItem("profileInfo", "我的资料", "/pages/profile/assets/i78@3x.png", "blue-purple", profileReminderBadge(persistentProfileReminder), "orange", "/pages/profile/system-management/profile-info/index"),
				profileHomeItem("skillConfig", "技能配置", "/pages/profile/assets/i79@3x.png", "blue-purple", "", "", "/pages/profile/system-management/skill-config/index"),
				profileHomeItem("blockSettings", "屏蔽设置", "/pages/profile/assets/i80@3x.png", "blue-purple", "", "", "/pages/profile/system-management/block-settings/index"),
				profileHomeItem("credit", "信用中心", "/pages/profile/assets/i81@3x.png", "blue-purple", "", "", "/pages/profile/credit-center/index"),
				profileHomeItem("reports", "举报中心", "/pages/profile/assets/i82@3x.png", "blue-purple", badgeText(unreadCount, "条消息"), "orange", "/pages/profile/system-management/report-center/index"),
				profileHomeItem("agreement", "签署协议", "/pages/profile/assets/i83@3x.png", "blue-purple", "", "", "/pages/profile/system-management/agreement-sign/index"),
				profileHomeItem("feedback", "建议反馈", "/pages/profile/assets/i84@3x.png", "blue-purple", "", "", "/pages/profile/system-management/feedback/index"),
				profileHomeItem("settings", "系统设置", "/pages/profile/assets/i85@3x.png", "blue-purple", "", "", "/pages/profile/settings/index"),
			},
		},
	}
}

func profileReminderBadge(visible bool) string {
	if visible {
		return "待完善"
	}
	return ""
}

func (s *Server) profileReportMessageCount(userID int64) int {
	count, _ := s.profileReportMessageCountStrict(userID)
	return count
}

func (s *Server) profileReportMessageCountStrict(userID int64) (int, error) {
	if s.reports == nil {
		return 0, nil
	}
	seen := map[int64]bool{}
	count := 0
	add := func(items []reports.Report) {
		for _, item := range items {
			if item.ID <= 0 || seen[item.ID] || !activeReportStatus(item.Status) {
				continue
			}
			seen[item.ID] = true
			count++
		}
	}
	myReports, err := s.reports.MyStrict(userID)
	if err != nil {
		return 0, err
	}
	appeals, err := s.reports.AppealsStrict(userID)
	if err != nil {
		return 0, err
	}
	add(myReports)
	add(appeals)
	return count, nil
}

func activeReportStatus(status string) bool {
	switch status {
	case "", "pending", "assigned", "appealed":
		return true
	default:
		return false
	}
}

func countManagedUserGames(userID int64, items []games.Game) int {
	count := 0
	for _, game := range items {
		if game.Status == "canceled" || game.Status == "cancelled" {
			continue
		}
		if game.CreatorUserID == userID || game.MainGuideUserID == userID {
			count++
		}
	}
	return count
}

func countInProgressUserGames(userID int64, items []games.Game, isMember func(int64, int64) bool) int {
	count := 0
	for _, game := range items {
		if game.Status != "in_progress" {
			continue
		}
		if game.CreatorUserID == userID || game.MainGuideUserID == userID || isMember(game.ID, userID) {
			count++
		}
	}
	return count
}

func profileHomeItem(key string, title string, iconSrc string, iconClass string, badge string, badgeClass string, route string) map[string]interface{} {
	return map[string]interface{}{
		"key":        key,
		"title":      title,
		"iconSrc":    iconSrc,
		"iconClass":  iconClass,
		"badge":      badge,
		"badgeClass": badgeClass,
		"route":      route,
		"enabled":    route != "",
	}
}

func profileHomeDisabledItem(key string, title string, iconSrc string, iconClass string, badge string, badgeClass string, reason string) map[string]interface{} {
	item := profileHomeItem(key, title, iconSrc, iconClass, badge, badgeClass, "")
	item["disabledReason"] = reason
	return item
}

func badgeText(count int, suffix string) string {
	if count <= 0 {
		return ""
	}
	return strconv.Itoa(count) + suffix
}

func limitedFootprints(items []reviews.Footprint, limit int) []reviews.Footprint {
	if len(items) > limit {
		return items[:limit]
	}
	return items
}
