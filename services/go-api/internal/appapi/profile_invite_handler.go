package appapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/invites"
)

func (s *Server) profileInviteOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items, err := s.connections.MyStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请关系失败，请稍后重试")
		return
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请记录失败，请稍后重试")
		return
	}
	items = s.inviteConnections(userID, items, relations)
	convertedCount, err := s.countConvertedInviteesStrict(relations)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀约转化数据失败，请稍后重试")
		return
	}
	inviteCode := ""
	availableCodes := make(map[string]string)
	codes, err := s.auth.AdminInviteCodes(invites.CodeFilter{OwnerID: userID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请码失败，请稍后重试")
		return
	}
	for _, item := range codes {
		if item.Status == invites.StatusActive && item.MaxUses == 1 && item.UsedCount == 0 && item.BoundWechatUserID == 0 {
			entryType := invites.NormalizeEntryType(item.EntryType)
			if availableCodes[entryType] == "" {
				availableCodes[entryType] = item.Code
			}
		}
	}
	for _, entryType := range []string{invites.EntryTypeLink, invites.EntryTypeQRCode, invites.EntryTypePoster} {
		if availableCodes[entryType] != "" {
			inviteCode = availableCodes[entryType]
			break
		}
	}
	actions := make([]map[string]interface{}, 0, 4)
	if availableCodes[invites.EntryTypeLink] != "" {
		actions = append(actions, map[string]interface{}{"key": "share_card", "icon": "🔗", "label": "分享邀请码", "inviteCode": availableCodes[invites.EntryTypeLink]})
	}
	if availableCodes[invites.EntryTypeQRCode] != "" {
		actions = append(actions, map[string]interface{}{"key": "qrcode", "icon": "▦", "label": "二维码", "iconClass": "white", "inviteCode": availableCodes[invites.EntryTypeQRCode]})
	}
	if availableCodes[invites.EntryTypePoster] != "" {
		actions = append(actions, map[string]interface{}{"key": "poster", "icon": "▧", "label": "生成海报", "iconClass": "white", "inviteCode": availableCodes[invites.EntryTypePoster]})
	}
	actions = append(actions, map[string]interface{}{"key": "manage_codes", "icon": "⚙", "label": "邀请码管理", "iconClass": "white"})
	s.recordBehavior(userID, "view_profile_invite_overview", "profile_invite", userID, nil)
	httpx.OK(w, map[string]interface{}{
		"profile": map[string]interface{}{
			"level": "V" + strconv.Itoa(maxInt(1, len(items)+1)) + " 探险家",
			"name":  s.displayName(userID, "用户"),
			"desc":  "邀请码: " + inviteCode + " · 已邀请 " + strconv.Itoa(len(relations)) + "人",
		},
		"metrics": []map[string]interface{}{
			{"value": strconv.Itoa(len(relations)), "label": "总邀约数", "trend": "▲", "tone": "up"},
			{"value": strconv.Itoa(convertedCount), "label": "成功转化", "trend": "▲", "tone": "up"},
			{"value": conversionRateText(convertedCount, len(relations)), "label": "转化率", "trend": "▲", "tone": "up"},
		},
		"actions": actions,
		"tabs":    []string{"数据概览", "关系网络", "邀约记录"},
		"trends": []map[string]interface{}{
			{"label": "累计邀约人数", "value": strconv.Itoa(len(relations)), "tone": "cyan"},
			{"label": "累计成功转化", "value": strconv.Itoa(convertedCount), "tone": "green"},
		},
	})
}

func (s *Server) countConvertedInvitees(relations []invites.Relation) int {
	count, _ := s.countConvertedInviteesStrict(relations)
	return count
}

func (s *Server) countConvertedInviteesStrict(relations []invites.Relation) (int, error) {
	seen := make(map[int64]bool, len(relations))
	count := 0
	gameItems, err := s.games.ListStrict()
	if err != nil {
		return 0, err
	}
	for _, relation := range relations {
		if relation.InviteeUserID <= 0 || seen[relation.InviteeUserID] {
			continue
		}
		seen[relation.InviteeUserID] = true
		converted, err := s.inviteeHasCompletedGameStrict(relation.InviteeUserID, gameItems)
		if err != nil {
			return 0, err
		}
		if converted {
			count++
		}
	}
	return count, nil
}

// 邀请转化只能在局正式完成后成立。待评价仅表示局流程结束，评价、争议
// 或结算仍可能改变结果，不能提前作为领路人的成功转化。
func (s *Server) inviteeHasCompletedGame(userID int64) bool {
	gameItems, err := s.games.ListStrict()
	if err != nil {
		return false
	}
	completed, _ := s.inviteeHasCompletedGameStrict(userID, gameItems)
	return completed
}

func (s *Server) inviteeHasCompletedGameStrict(userID int64, gameItems []games.Game) (bool, error) {
	for _, game := range gameItems {
		if game.Status != "completed" {
			continue
		}
		member, err := s.games.IsMemberStrict(game.ID, userID)
		if err != nil {
			return false, err
		}
		if member {
			return true, nil
		}
	}
	return false, nil
}

func (s *Server) profileInviteCodes(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items, err := s.auth.AdminInviteCodes(invites.CodeFilter{OwnerID: userID})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取邀请码失败")
		return
	}
	requests, err := s.auth.InviteQuotaRequests(userID, "")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "获取加量申请失败")
		return
	}
	usedCount, availableCount := 0, 0
	for _, item := range items {
		if item.MaxUses != 1 {
			continue
		}
		if item.UsedCount > 0 || item.BoundWechatUserID > 0 {
			usedCount++
		} else if item.Status == invites.StatusActive {
			availableCount++
		}
	}
	httpx.OK(w, map[string]interface{}{
		"items": items, "requests": requests, "total": len(items),
		"summary": map[string]int{"total": len(items), "used": usedCount, "available": availableCount, "pendingRequests": len(filterInviteQuotaRequests(requests, "pending"))},
		"config":  s.inviteCodeConfig(),
	})
}

func filterInviteQuotaRequests(items []invites.QuotaRequest, status string) []invites.QuotaRequest {
	result := make([]invites.QuotaRequest, 0)
	for _, item := range items {
		if item.Status == status {
			result = append(result, item)
		}
	}
	return result
}

func (s *Server) profileInviteQuotaRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Quantity int    `json:"quantity"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	config, err := s.inviteCodeConfigStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请码配置失败，请稍后重试")
		return
	}
	if req.Quantity > config.MaxRequestCount {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "申请数量超过当前可申请上限")
		return
	}
	request, err := s.auth.CreateInviteQuotaRequest(userID, req.Quantity, req.Reason)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "申请数量应为 1 至 200 个")
		return
	}
	s.recordBehavior(userID, "create_invite_quota_request", "invite_quota_request", request.ID, map[string]interface{}{"quantity": request.Quantity})
	httpx.OK(w, request)
}

func (s *Server) profileInviteNetwork(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireRoleInviteUser(w, r)
	if !ok {
		return
	}
	items, err := s.inviteConnectionsForUser(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请关系失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "view_profile_invite_network", "profile_invite", userID, nil)
	httpx.OK(w, map[string]interface{}{
		"summary": []map[string]interface{}{
			{"value": strconv.Itoa(len(items)), "label": "已服务"},
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
	items, err := s.inviteConnectionsForUser(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请记录失败，请稍后重试")
		return
	}
	records := s.inviteRecords(userID, items)
	timeoutCount := countInviteRecordsByStatus(records, "timeout")
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "all"
	}
	s.recordBehavior(userID, "view_profile_invite_records", "profile_invite", userID, map[string]interface{}{"status": status})
	httpx.OK(w, map[string]interface{}{
		"activeStatus": status,
		"filters": []map[string]interface{}{
			{"key": "all", "label": "全部"},
			{"key": "progress", "label": "进行中(" + strconv.Itoa(countInviteRecordsByStatus(records, "progress")) + ")"},
			{"key": "completed", "label": "已完成(" + strconv.Itoa(countInviteRecordsByStatus(records, "completed")) + ")"},
			{"key": "timeout", "label": "超时(" + strconv.Itoa(timeoutCount) + ")"},
		},
		"allRecords": records,
		"records":    filterInviteRecords(records, "", status),
		"emptyText":  "暂无邀约记录",
		"timeoutWarning": map[string]interface{}{
			"show":  timeoutCount > 0,
			"title": "超时预警",
			"text":  "有" + strconv.Itoa(timeoutCount) + "个组局超过15天无进展",
		},
	})
}

func (s *Server) profileInviteRanking(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireRoleInviteUser(w, r); !ok {
		return
	}
	httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "贡献排行将在二期开放")
}

func (s *Server) profileInviteIncome(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireRoleInviteUser(w, r); !ok {
		return
	}
	httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "收益功能将在后续版本开放")
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
	items, err := s.inviteConnectionsForUser(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请成员失败，请稍后重试")
		return
	}
	var matched connections.Connection
	for _, item := range items {
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
	snapshot, err := s.profiles.RoleSnapshotStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取邀请权限失败，请稍后重试")
		return 0, false
	}
	roles := snapshot.RoleStatusMap
	if roles["guide"] == "approved" || roles["guide"] == "active" {
		return userID, true
	}
	httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅领路人可使用邀请功能")
	return 0, false
}

func (s *Server) inviteConnectionsForUser(userID int64) ([]connections.Connection, error) {
	items, err := s.connections.MyStrict(userID)
	if err != nil {
		return nil, err
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{InviterUserID: userID})
	if err != nil {
		return nil, err
	}
	return s.inviteConnections(userID, items, relations), nil
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
		// 邀请绑定关系的旧表没有绑定时间字段。不能在每次读取时伪造“现在”
		// 作为建立时间，否则邀约记录会不断刷新，15 天超时统计永远失真。
		filtered = append(filtered, connections.Connection{ID: relation.InviteeUserID, UserID: userID, ConnectedUserID: relation.InviteeUserID, RelationType: "invite", SourceType: "invite", SourceID: relation.InviteCodeID})
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
		if status == "completed" {
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
			"gameCount":          len(games),
			"completedGameCount": completed,
			"games":              games,
		})
	}
	return members
}

func (s *Server) inviteRecords(inviterUserID int64, items []connections.Connection) []map[string]interface{} {
	records := make([]map[string]interface{}, 0, len(items))
	inviterName := s.displayName(inviterUserID, "我")
	inviterRole := s.inviteRoleLabel(inviterUserID)
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
		} else if !item.UpdatedAt.IsZero() && time.Since(item.UpdatedAt) >= 15*24*time.Hour {
			status = "timeout"
			statusText = "超时"
			statusClass = "orange"
		}
		createdTime := inviteRecordTime(item.CreatedAt, "绑定时间待补")
		updatedTime := inviteRecordTime(item.UpdatedAt, "状态待更新")
		records = append(records, map[string]interface{}{
			"role":             "referred",
			"statusKey":        status,
			"status":           statusText,
			"statusClass":      statusClass,
			"id":               "REF-" + strconv.FormatInt(item.ID, 10),
			"time":             createdTime,
			"inviterAvatar":    avatarTextForName(inviterName, inviterUserID),
			"inviterName":      inviterName,
			"inviterRoleLabel": inviterRole,
			"inviteeAvatar":    avatarTextForName(name, item.ConnectedUserID),
			"inviteeName":      name,
			"inviteeRoleLabel": "受邀成员",
			"memberId":         strconv.FormatInt(item.ConnectedUserID, 10),
			"title":            "邀请关系服务",
			"route":            "/pages/profile/service-center/invite/member-detail/index?memberId=" + strconv.FormatInt(item.ConnectedUserID, 10),
			"steps": []map[string]interface{}{
				{"label": "邀请关系建立", "time": createdTime},
				{"label": statusText, "time": updatedTime},
			},
			"actions":   []string{"查看详情"},
			"games":     invitedGames,
			"gameCount": len(invitedGames),
		})
	}
	return records
}

func inviteRecordTime(value time.Time, fallback string) string {
	if value.IsZero() {
		return fallback
	}
	return value.Format("01-02 15:04")
}

func (s *Server) inviteRoleLabel(userID int64) string {
	roles := s.profiles.RoleSnapshot(userID).RoleStatusMap
	if roles["guide"] == "approved" || roles["guide"] == "active" {
		return "领路人"
	}
	if roles["expert"] == "approved" || roles["expert"] == "active" {
		return "行家"
	}
	return "邀请人"
}

func (s *Server) inviteRankingMembers(items []connections.Connection) []map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(items))
	for index, item := range items {
		name := s.displayName(item.ConnectedUserID, "成员")
		members = append(members, map[string]interface{}{
			"id":          strconv.FormatInt(item.ConnectedUserID, 10),
			"rank":        index + 1,
			"avatar":      avatarTextForName(name, item.ConnectedUserID),
			"name":        name,
			"level":       "一级成员",
			"activeDays":  s.invitedActiveDays(item.ConnectedUserID),
			"inviteCount": s.inviteCountForUser(item.ConnectedUserID),
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

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
