package appapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/revenue"
)

func (s *Server) profileInviteOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items := s.connections.My(userID)
	relations, _ := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	items = s.inviteConnections(userID, items, relations)
	income := s.phaseOneIncomeSummary(userID)
	revenueEnabled := s.currentOperationRules().Revenue.Enabled
	inviteCode, _ := s.auth.InviteCodeForUser(userID)
	s.recordBehavior(userID, "view_profile_invite_overview", "profile_invite", userID, nil)
	httpx.OK(w, map[string]interface{}{
		"profile": map[string]interface{}{
			"level": "V" + strconv.Itoa(maxInt(1, len(items)+1)) + " 探险家",
			"name":  s.displayName(userID, "用户"),
			"desc":  "邀请码: " + inviteCode + " · 已邀请 " + strconv.Itoa(len(relations)) + "人",
		},
		"metrics": []map[string]interface{}{
			{"value": strconv.Itoa(len(relations)), "label": "总邀约数", "trend": "▲", "tone": "up"},
			{"value": strconv.Itoa(countStrongConnections(items)), "label": "成功转化", "trend": "▲", "tone": "up"},
			{"value": conversionRateText(countStrongConnections(items), len(relations)), "label": "转化率", "trend": "▲", "tone": "up"},
			{"value": incomeDisplayText(revenueEnabled, income.TotalCent), "label": "分润收益", "trend": "▲", "tone": "up"},
		},
		"actions": []map[string]interface{}{
			{"key": "share_card", "icon": "🔗", "label": "分享邀请码", "inviteCode": inviteCode},
			{"key": "qrcode", "icon": "▦", "label": "二维码", "iconClass": "white", "inviteCode": inviteCode},
			{"key": "poster", "icon": "▧", "label": "生成海报", "iconClass": "white", "inviteCode": inviteCode},
		},
		"tabs": []string{"数据概览", "关系网络", "邀约记录", "贡献排行", "收益明细"},
		"trends": []map[string]interface{}{
			{"label": "本周新增邀约", "value": "+" + strconv.Itoa(len(relations)), "tone": "cyan"},
			{"label": "本周新增转化", "value": "+" + strconv.Itoa(countStrongConnections(items)), "tone": "green"},
			{"label": "本周分润", "value": incomeDisplayText(revenueEnabled, income.SettledCent), "tone": "cyan"},
		},
	})
}

func (s *Server) profileInviteNetwork(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items := s.inviteConnectionsForUser(userID)
	income := s.phaseOneIncomeSummary(userID)
	revenueEnabled := s.currentOperationRules().Revenue.Enabled
	s.recordBehavior(userID, "view_profile_invite_network", "profile_invite", userID, nil)
	httpx.OK(w, map[string]interface{}{
		"summary": []map[string]interface{}{
			{"value": strconv.Itoa(len(items)), "label": "已服务\n位玩家"},
			{"value": incomeDisplayText(revenueEnabled, income.SettledCent), "label": "本周收益"},
		},
		"networkNodes":       inviteNetworkNodes(s, items),
		"networkNodeDetails": s.inviteNetworkNodeDetails(items),
		"networkEdges":       s.inviteNetworkEdges(userID, items),
		"invitedGameSummary": s.invitedGameSummary(items),
		"avatars":            inviteAvatarList(s, items),
		"members":            s.inviteMembers(items),
	})
}

func (s *Server) profileInviteRecords(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items := s.inviteConnectionsForUser(userID)
	records := s.inviteRecords(items)
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if role == "" {
		role = "referred"
	}
	if status == "" {
		status = "all"
	}
	s.recordBehavior(userID, "view_profile_invite_records", "profile_invite", userID, map[string]interface{}{"role": role, "status": status})
	httpx.OK(w, map[string]interface{}{
		"activeRole":   role,
		"activeStatus": status,
		"roleTabs": []map[string]interface{}{
			{"key": "referred", "label": "我引荐的", "count": len(records)},
			{"key": "created", "label": "我发起的", "count": 0},
		},
		"filters": []map[string]interface{}{
			{"key": "all", "label": "全部"},
			{"key": "progress", "label": "进行中(" + strconv.Itoa(countInviteRecordsByStatus(records, "progress")) + ")"},
			{"key": "completed", "label": "已完成(" + strconv.Itoa(countInviteRecordsByStatus(records, "completed")) + ")"},
			{"key": "timeout", "label": "超时(0)"},
			{"key": "cancelled", "label": "已取消(0)"},
		},
		"allRecords": records,
		"records":    filterInviteRecords(records, role, status),
		"emptyText":  "暂无邀约记录",
	})
}

func (s *Server) profileInviteRanking(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items := s.inviteConnectionsForUser(userID)
	s.recordBehavior(userID, "view_profile_invite_ranking", "profile_invite", userID, nil)
	httpx.OK(w, map[string]interface{}{
		"activePeriodIndex": 0,
		"activeType":        defaultString(r.URL.Query().Get("type"), "inviteCount"),
		"periods": []map[string]string{
			{"key": "week", "label": "本周"},
			{"key": "month", "label": "本月"},
			{"key": "quarter", "label": "本季"},
			{"key": "year", "label": "本年"},
			{"key": "all", "label": "全部"},
		},
		"rankTypes": []map[string]string{
			{"key": "inviteCount", "label": "邀约数排行"},
			{"key": "profitContribution", "label": "分润贡献排行"},
		},
		"members": s.inviteRankingMembers(items),
	})
}

func (s *Server) profileInviteIncome(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	summary := s.phaseOneIncomeSummary(userID)
	revenueEnabled := s.currentOperationRules().Revenue.Enabled
	logs := []revenue.IncomeLog{}
	if s.currentOperationRules().Revenue.Enabled {
		logs = s.revenue.IncomeLogs(userID, "")
	}
	s.recordBehavior(userID, "view_profile_invite_income", "profile_invite", userID, nil)
	httpx.OK(w, map[string]interface{}{
		"trendSeries": inviteIncomeTrendSeries(logs),
		"metrics": []map[string]interface{}{
			{"label": "本月分润", "value": incomeDisplayText(revenueEnabled, summary.SettledCent), "desc": "已结算收益"},
			{"label": "累计分润", "value": incomeDisplayText(revenueEnabled, summary.TotalCent), "desc": "含待结算收益"},
			{"label": "活跃成员", "value": strconv.Itoa(len(s.inviteConnectionsForUser(userID))), "desc": "当前关系数"},
			{"label": "产生分润局数", "value": strconv.Itoa(len(logs)), "desc": "累计流水"},
		},
		"flows": inviteIncomeFlows(logs),
	})
}

func (s *Server) profileInviteMemberDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	memberID := parseInviteMemberID(r.URL.Query().Get("memberId"))
	if memberID == 0 {
		memberID = parseInviteMemberID(r.URL.Query().Get("id"))
	}
	if memberID == 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "成员 ID 错误")
		return
	}
	var matched connections.Connection
	for _, item := range s.inviteConnectionsForUser(userID) {
		if item.ConnectedUserID == memberID || item.ID == memberID {
			matched = item
			break
		}
	}
	if matched.ID == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "成员不存在")
		return
	}
	name := s.displayName(matched.ConnectedUserID, "成员")
	invitedGames := s.invitedGames(matched.ConnectedUserID)
	completedGameCount := countCompletedInvitedGames(invitedGames)
	s.recordBehavior(userID, "view_profile_invite_member", "profile_invite_member", matched.ConnectedUserID, nil)
	httpx.OK(w, map[string]interface{}{
		"memberId": strconv.FormatInt(matched.ConnectedUserID, 10),
		"member": map[string]interface{}{
			"avatar": avatarTextForName(name, matched.ConnectedUserID),
			"name":   name,
			"level":  "一级成员",
		},
		"stats": []map[string]interface{}{
			{"value": strconv.Itoa(len(invitedGames)), "label": "参与组局"},
			{"value": strconv.Itoa(completedGameCount), "label": "完成组局"},
			{"value": conversionRateText(completedGameCount, len(invitedGames)), "label": "完成率"},
		},
		"income": []map[string]interface{}{
			{"label": "直接贡献收益", "value": "一期未启用"},
			{"label": "团队贡献收益", "value": "一期未启用"},
			{"label": "合计贡献", "value": "一期未启用", "highlight": true},
		},
		"activities": []map[string]interface{}{
			{"icon": "🎯", "title": "邀请关系建立", "time": matched.CreatedAt.Format("01-02 15:04"), "amount": "+"},
			{"icon": "📈", "title": "关系强度更新", "time": matched.UpdatedAt.Format("01-02 15:04"), "amount": "+"},
		},
		"games":              invitedGames,
		"gameCount":          len(invitedGames),
		"completedGameCount": completedGameCount,
	})
}

func (s *Server) requireRoleInviteUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return 0, false
	}
	roles := s.profiles.RoleSnapshot(userID).RoleStatusMap
	if roles["expert"] == "approved" || roles["expert"] == "active" || roles["guide"] == "approved" || roles["guide"] == "active" {
		return userID, true
	}
	httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅行家或领路人可使用邀请功能")
	return 0, false
}

func (s *Server) inviteConnectionsForUser(userID int64) []connections.Connection {
	items := s.connections.My(userID)
	relations, _ := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	return s.inviteConnections(userID, items, relations)
}

func (s *Server) inviteConnections(userID int64, items []connections.Connection, relations []invites.Relation) []connections.Connection {
	filtered := connectionsForInvitees(items, relations)
	if len(relations) == 0 {
		for _, item := range items {
			if item.RelationType == "invite" || item.SourceType == "invite" {
				filtered = append(filtered, item)
			}
		}
	}
	seen := make(map[int64]bool, len(filtered))
	for _, item := range filtered {
		seen[item.ConnectedUserID] = true
	}
	for _, relation := range relations {
		if relation.InviteeUserID <= 0 || seen[relation.InviteeUserID] {
			continue
		}
		now := time.Now()
		filtered = append(filtered, connections.Connection{ID: relation.InviteeUserID, UserID: userID, ConnectedUserID: relation.InviteeUserID, RelationType: "invite", SourceType: "invite", SourceID: relation.InviteCodeID, CreatedAt: now, UpdatedAt: now})
	}
	return filtered
}

func (s *Server) invitedGames(userID int64) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, game := range s.games.List() {
		if game.CreatorUserID != userID && !s.games.IsMember(game.ID, userID) {
			continue
		}
		result = append(result, map[string]interface{}{"id": game.ID, "title": game.Title, "status": game.Status, "statusText": homeGameStatusText(game.Status), "createdAt": game.CreatedAt.Format(time.RFC3339)})
	}
	return result
}

func countCompletedInvitedGames(items []map[string]interface{}) int {
	count := 0
	for _, item := range items {
		status, _ := item["status"].(string)
		if status == "pending_review" || status == "completed" {
			count++
		}
	}
	return count
}

func (s *Server) inviteMembers(items []connections.Connection) []map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		name := s.displayName(item.ConnectedUserID, "成员")
		games := s.invitedGames(item.ConnectedUserID)
		completed := countCompletedInvitedGames(games)
		members = append(members, map[string]interface{}{
			"id":                 strconv.FormatInt(item.ConnectedUserID, 10),
			"avatar":             avatarTextForName(name, item.ConnectedUserID),
			"name":               name,
			"desc":               "邀约 1 · 转化 " + strconv.Itoa(completed) + " · 参与 " + strconv.Itoa(len(games)) + "局",
			"direct":             "+",
			"team":               s.inviteContributionText(item.ConnectedUserID),
			"gameCount":          len(games),
			"completedGameCount": completed,
			"games":              games,
		})
	}
	return members
}

func (s *Server) inviteRecords(items []connections.Connection) []map[string]interface{} {
	records := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		name := s.displayName(item.ConnectedUserID, "成员")
		invitedGames := s.invitedGames(item.ConnectedUserID)
		status := "progress"
		statusText := "进行中"
		statusClass := "blue"
		if countCompletedInvitedGames(invitedGames) > 0 {
			status = "completed"
			statusText = "已完成"
			statusClass = "green"
		}
		records = append(records, map[string]interface{}{
			"role":         "referred",
			"statusKey":    status,
			"status":       statusText,
			"statusClass":  statusClass,
			"id":           "REF-" + strconv.FormatInt(item.ID, 10),
			"time":         item.CreatedAt.Format("01-02 15:04"),
			"expertAvatar": avatarTextForName(name, item.ConnectedUserID),
			"expert":       name,
			"memberId":     strconv.FormatInt(item.ConnectedUserID, 10),
			"playerAvatar": "我",
			"player":       "我",
			"title":        "邀请关系服务",
			"budget":       "你的奖励：" + s.inviteContributionText(item.ConnectedUserID),
			"income":       incomeStatusText(status),
			"route":        "/pages/profile/service-center/invite/member-detail/index?memberId=" + strconv.FormatInt(item.ConnectedUserID, 10),
			"steps": []map[string]interface{}{
				{"label": "邀请关系建立", "time": item.CreatedAt.Format("01-02 15:04")},
				{"label": statusText, "time": item.UpdatedAt.Format("01-02 15:04")},
			},
			"actions":   []string{"查看详情"},
			"games":     invitedGames,
			"gameCount": len(invitedGames),
		})
	}
	return records
}

func (s *Server) inviteRankingMembers(items []connections.Connection) []map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(items))
	for index, item := range items {
		name := s.displayName(item.ConnectedUserID, "成员")
		members = append(members, map[string]interface{}{
			"id":                 strconv.FormatInt(item.ConnectedUserID, 10),
			"rank":               index + 1,
			"avatar":             avatarTextForName(name, item.ConnectedUserID),
			"name":               name,
			"level":              "一级成员",
			"activeDays":         s.invitedActiveDays(item.ConnectedUserID),
			"inviteCount":        s.inviteCountForUser(item.ConnectedUserID),
			"profitContribution": s.inviteContributionText(item.ConnectedUserID),
		})
	}
	return members
}

func inviteNetworkNodes(s *Server, items []connections.Connection) []string {
	nodes := make([]string, 0, minInt(8, len(items)))
	for i, item := range items {
		if i >= 8 {
			break
		}
		name := s.displayName(item.ConnectedUserID, "成员")
		nodes = append(nodes, avatarTextForName(name, item.ConnectedUserID))
	}
	return nodes
}

func (s *Server) inviteNetworkNodeDetails(items []connections.Connection) []map[string]interface{} {
	details := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		name := s.displayName(item.ConnectedUserID, "成员")
		games := s.invitedGames(item.ConnectedUserID)
		details = append(details, map[string]interface{}{
			"id":                 strconv.FormatInt(item.ConnectedUserID, 10),
			"userId":             item.ConnectedUserID,
			"name":               name,
			"avatar":             avatarTextForName(name, item.ConnectedUserID),
			"relation":           "direct_invitee",
			"level":              1,
			"gameCount":          len(games),
			"completedGameCount": countCompletedInvitedGames(games),
		})
	}
	return details
}

func (s *Server) inviteNetworkEdges(userID int64, items []connections.Connection) []map[string]interface{} {
	edges := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		edges = append(edges, map[string]interface{}{
			"source":   strconv.FormatInt(userID, 10),
			"target":   strconv.FormatInt(item.ConnectedUserID, 10),
			"relation": "invite",
		})
	}
	return edges
}

func (s *Server) invitedGameSummary(items []connections.Connection) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		games := s.invitedGames(item.ConnectedUserID)
		result = append(result, map[string]interface{}{
			"memberId":           strconv.FormatInt(item.ConnectedUserID, 10),
			"gameCount":          len(games),
			"completedGameCount": countCompletedInvitedGames(games),
			"games":              games,
		})
	}
	return result
}

func (s *Server) inviteCountForUser(userID int64) int {
	if userID <= 0 || s.auth == nil {
		return 0
	}
	relations, _ := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	return len(relations)
}

func (s *Server) invitedActiveDays(userID int64) int {
	games := s.invitedGames(userID)
	seen := make(map[string]struct{}, len(games))
	for _, game := range games {
		if value, ok := game["createdAt"].(string); ok && len(value) >= 10 {
			seen[value[:10]] = struct{}{}
		}
	}
	return len(seen)
}

func (s *Server) inviteContributionText(userID int64) string {
	if !s.currentOperationRules().Revenue.Enabled {
		return "一期未启用"
	}
	return "按实际结算"
}

func inviteAvatarList(s *Server, items []connections.Connection) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		name := s.displayName(item.ConnectedUserID, "成员")
		result = append(result, avatarTextForName(name, item.ConnectedUserID))
	}
	return result
}

func filterInviteRecords(records []map[string]interface{}, role string, status string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		if role != "" && record["role"] != role {
			continue
		}
		if status != "" && status != "all" && record["statusKey"] != status {
			continue
		}
		result = append(result, record)
	}
	return result
}

func countInviteRecordsByStatus(records []map[string]interface{}, status string) int {
	count := 0
	for _, record := range records {
		if record["statusKey"] == status {
			count++
		}
	}
	return count
}

func inviteIncomeTrendSeries(logs []revenue.IncomeLog) []map[string]interface{} {
	if len(logs) == 0 {
		return []map[string]interface{}{}
	}
	monthly := make(map[string]int64)
	order := make([]string, 0, len(logs))
	for _, log := range logs {
		month := log.CreatedAt.Format("2006-01")
		if _, ok := monthly[month]; !ok {
			order = append(order, month)
		}
		monthly[month] += log.AmountCent
	}
	result := make([]map[string]interface{}, 0, len(order))
	for _, month := range order {
		parsed, _ := time.Parse("2006-01", month)
		result = append(result, map[string]interface{}{"month": parsed.Format("1月"), "amount": monthly[month]})
	}
	return result
}

func inviteIncomeFlows(logs []revenue.IncomeLog) []map[string]interface{} {
	flows := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		flows = append(flows, map[string]interface{}{
			"icon":   "🎯",
			"title":  "组局分润 · " + log.Role,
			"time":   log.CreatedAt.Format("01-02 15:04"),
			"amount": "+" + moneyYuanText(log.AmountCent),
		})
	}
	return flows
}

func parseInviteMemberID(value string) int64 {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "member-")
	id, _ := strconv.ParseInt(value, 10, 64)
	return id
}

func conversionRateText(success int, total int) string {
	if total <= 0 {
		return "0%"
	}
	return strconv.Itoa(success*100/total) + "%"
}

func moneyYuanText(cent int64) string {
	return "¥" + strconv.FormatFloat(float64(cent)/100, 'f', 2, 64)
}

func incomeDisplayText(enabled bool, cent int64) string {
	if !enabled {
		return "一期未启用"
	}
	return moneyYuanText(cent)
}

func incomeStatusText(status string) string {
	if status == "completed" {
		return "已到账"
	}
	return ""
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
