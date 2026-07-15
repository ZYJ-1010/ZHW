package appapi

import (
	"net/http"
	"strconv"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/reports"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/users"
)

func (s *Server) profileHome(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}

	user, _ := s.auth.CurrentUser(bearerToken(r.Header.Get("Authorization")))
	record := s.identity.Status(userID)
	current := s.buildCurrentUserDTO(user, record)
	todos, err := s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}

	httpx.OK(w, s.buildProfileHomePayload(userID, user, current, len(todos)))
}

func (s *Server) buildProfileHomePayload(userID int64, user users.User, current CurrentUserDTO, reviewTodoCount int) map[string]interface{} {
	conns := s.connections.My(userID)
	inviteRelations, _ := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	invitedConns := connectionsForInvitees(conns, inviteRelations)
	income := s.revenue.IncomeSummary(userID)
	points := s.points.Summary(userID)
	membership := s.membership.My(userID)
	inProgressGameCount := countInProgressUserGames(userID, s.games.List(), s.games.IsMember)
	managedServiceCount := countManagedUserGames(userID, s.games.List())
	notifications := unreadNotificationCount(s.notices.List(userID))
	reportMessageCount := s.profileReportMessageCount(userID)
	growthLevel := "V" + strconv.Itoa(maxInt(1, current.Growth.Level)) + " 探险家"
	role := homeRoleName(s.homeRoleType(userID))
	memberLevel := membership.PlanName
	if memberLevel == "" || memberLevel == "none" {
		memberLevel = "基础会员"
	}
	name := homeDisplayName(user, s.displayName(userID, "用户"))
	avatarFileID, avatarURL := s.currentUserAvatar(userID)

	return map[string]interface{}{
		"onlineText": s.homeOnlineText(0, 0),
		"user": map[string]interface{}{
			"nickname":     name,
			"avatarText":   avatarTextForName(name, userID),
			"avatarUrl":    avatarURL,
			"avatarFileId": avatarFileID,
			"memberLevel":  memberLevel,
			"roleLevel":    growthLevel,
			"growthLevel":  growthLevel,
			"role":         role,
			"inviteCode":   current.InviteCode,
		},
		"stats": []map[string]interface{}{
			{"key": "referrals", "label": "引荐数", "value": strconv.Itoa(len(inviteRelations))},
			{"key": "successes", "label": "成功数", "value": strconv.Itoa(countStrongConnections(invitedConns))},
			{"key": "dealAmount", "label": "成交总额", "value": moneyYuanText(income.TotalCent)},
			{"key": "credit", "label": "信用度", "value": strconv.Itoa(current.Growth.CreditScore)},
		},
		"assets": []map[string]interface{}{
			{"key": "totalDealAmount", "label": "总成交额", "value": moneyYuanText(income.TotalCent)},
			{"key": "withdrawable", "label": "可提现", "value": moneyYuanText(income.SettledCent), "tone": "green"},
			{"key": "pendingSettlement", "label": "待结算", "value": moneyYuanText(income.PendingCent), "tone": "orange"},
		},
		"assetSummary": map[string]interface{}{
			"totalDealAmountCent":   income.TotalCent,
			"withdrawableCent":      income.SettledCent,
			"pendingSettlementCent": income.PendingCent,
			"points":                points.AvailablePoints,
		},
		"vipBanner": map[string]interface{}{
			"text":       "升级会员，认证您的角色",
			"actionText": "增购会员 >",
			"route":      "/pages/profile/member/index",
		},
		"serviceSections": s.profileHomeSections(reviewTodoCount, reportMessageCount, len(inviteRelations), inProgressGameCount, managedServiceCount),
		"badges": map[string]interface{}{
			"reviewTodoCount":         reviewTodoCount,
			"unreadNotificationCount": notifications,
			"reportMessageCount":      reportMessageCount,
			"connectionCount":         len(conns),
		},
		"summary": CurrentUserSummaryDTO{
			User:                    current,
			ReviewTodoCount:         reviewTodoCount,
			UnreadNotificationCount: notifications,
			IncomeSummary:           income,
			PointsSummary:           points,
			RecentFootprints:        limitedFootprints(s.reviews.Footprints(userID), 5),
		},
	}
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

func (s *Server) profileHomeSections(reviewTodoCount int, unreadCount int, connectionCount int, inProgressGameCount int, managedServiceCount int) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"title": "服务中心",
			"items": []map[string]interface{}{
				profileHomeItem("myGames", "我的局", "/pages/profile/assets/i66@3x.png", "purple-blue", badgeText(inProgressGameCount, "进行中"), "pink", "/pages/profile/service-center/my-games/index"),
				profileHomeItem("gameManage", "组局管理", "/pages/profile/assets/i69@3x.png", "purple-blue", badgeText(managedServiceCount, "个服务"), "blue", "/pages/game/player-manage/index"),
				profileHomeItem("invite", "我的邀请", "/pages/profile/assets/i68@3x.png", "purple-blue", badgeText(connectionCount, "个关系"), "orange", "/pages/profile/service-center/invite/overview/index"),
				profileHomeItem("reviews", "评价管理", "/pages/profile/assets/i69@3x.png", "purple-blue", badgeText(reviewTodoCount, "待评价"), "pink", "/pages/profile/service-center/manage/review-manage/index"),
			},
		},
		{
			"title": "资产中心",
			"items": []map[string]interface{}{
				profileHomeItem("assets", "我的资产", "/pages/profile/assets/i70@3x.png", "orange", "", "", "/pages/profile/asset-center/manage/index"),
				profileHomeDisabledItem("deposit", "我的押金", "/pages/profile/assets/i71@3x.png", "orange", "", "", "一期未接真实支付，押金账户暂不可查看"),
				profileHomeItem("mall", "积分商城", "/pages/profile/assets/i72@3x.png", "orange", "", "", "/pages/profile/asset-center/mall/index"),
				profileHomeItem("points", "我的积分", "/pages/profile/assets/i73@3x.png", "orange", "", "", "/pages/profile/asset-center/points/index"),
				profileHomeDisabledItem("invoice", "开票中心", "/pages/profile/assets/i74@3x.png", "orange", "", "", "发票能力需上线资质确认后开放"),
			},
		},
		{
			"title": "足迹中心",
			"items": []map[string]interface{}{
				profileHomeItem("footprints", "我的足迹", "/pages/profile/assets/i75@3x.png", "teal", "", "", "/pages/profile/footprint/achievements/index"),
				profileHomeItem("cityStories", "我的城市故事", "/pages/profile/assets/i76@3x.png", "teal", "", "", "/pages/map/my-city/index"),
				profileHomeItem("achievements", "我的成就墙", "/pages/profile/assets/i77@3x.png", "teal", "", "", "/pages/profile/footprint/achievements/index"),
			},
		},
		{
			"title": "系统管理",
			"items": []map[string]interface{}{
				profileHomeItem("profileInfo", "我的资料", "/pages/profile/assets/i78@3x.png", "blue-purple", "", "", "/pages/profile/system-management/profile-info/index"),
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

func (s *Server) profileReportMessageCount(userID int64) int {
	if s.reports == nil {
		return 0
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
	add(s.reports.My(userID))
	add(s.reports.Appeals(userID))
	return count
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
