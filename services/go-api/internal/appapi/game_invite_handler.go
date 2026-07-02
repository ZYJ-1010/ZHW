package appapi

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/users"
)

type inviteCandidateDTO struct {
	ID          int64    `json:"id"`
	UserID      int64    `json:"userId"`
	Name        string   `json:"name"`
	Nickname    string   `json:"nickname"`
	AvatarText  string   `json:"avatarText"`
	Tag         string   `json:"tag,omitempty"`
	Desc        string   `json:"desc,omitempty"`
	Meta        string   `json:"meta,omitempty"`
	RoleType    string   `json:"roleType,omitempty"`
	RoleLabel   string   `json:"roleLabel,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Match       int      `json:"match,omitempty"`
	ReviewCount int      `json:"reviewCount,omitempty"`
	Selected    bool     `json:"selected,omitempty"`
}

const gameInviteConfigKey = "game.invite_config"
const gameReplayQuickActionsConfigKey = "game.replay_quick_actions"
const gameReplayQuickMessagesConfigKey = "game.replay_quick_messages"
const gameSystemRecommendationsConfigKey = "game.system_recommendations_config"
const gameReferralRecordsConfigKey = "game.referral_records_config"

type inviteActivityTypeDTO struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type gameSystemRecommendationCategoryDTO struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type gameSystemRecommendationsConfigDTO struct {
	Title           string                                `json:"title"`
	Desc            string                                `json:"desc"`
	LoadingText     string                                `json:"loadingText"`
	EmptyText       string                                `json:"emptyText"`
	SummaryTemplate string                                `json:"summaryTemplate"`
	SummaryDesc     string                                `json:"summaryDesc"`
	CancelText      string                                `json:"cancelText"`
	ConfirmText     string                                `json:"confirmText"`
	MinSelectToast  string                                `json:"minSelectToast"`
	ConfirmingText  string                                `json:"confirmingText"`
	Categories      []gameSystemRecommendationCategoryDTO `json:"categories"`
	DefaultCategory string                                `json:"defaultCategory"`
	Version         string                                `json:"version"`
}

type inviteRewardRateConfigDTO struct {
	PlatformServiceRate   int `json:"platformServiceRate"`
	SystemGuideRewardRate int `json:"systemGuideRewardRate"`
	InviteRewardRate      int `json:"inviteRewardRate"`
}

type gameInviteConfigDTO struct {
	MinPlayerCount      int                       `json:"minPlayerCount"`
	MaxPlayerCount      int                       `json:"maxPlayerCount"`
	BudgetMaxAmount     int                       `json:"budgetMaxAmount"`
	DefaultBudget       string                    `json:"defaultBudget"`
	DefaultTitle        string                    `json:"defaultTitle"`
	DefaultDetail       string                    `json:"defaultDetail"`
	PlayerIntroTemplate string                    `json:"playerIntroTemplate"`
	ActivityTypes       []inviteActivityTypeDTO   `json:"activityTypes"`
	RewardRateConfig    inviteRewardRateConfigDTO `json:"rewardRateConfig"`
}

type replayQuickActionDTO struct {
	ID       string `json:"id"`
	Theme    string `json:"theme"`
	IconText string `json:"iconText,omitempty"`
	IconType string `json:"iconType,omitempty"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Route    string `json:"route"`
	Order    int    `json:"order"`
	Visible  bool   `json:"visible"`
}

type replayQuickMessageDTO struct {
	Text    string `json:"text"`
	Order   int    `json:"order"`
	Visible bool   `json:"visible"`
}

func (s *Server) gameInvitePlayerConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	expert := s.firstExpertCandidate(userID)
	config := s.currentGameInviteConfig()
	resp := map[string]interface{}{}
	raw, _ := json.Marshal(config)
	_ = json.Unmarshal(raw, &resp)
	resp["expert"] = expert
	resp["playerIntroTemplate"] = strings.ReplaceAll(config.PlayerIntroTemplate, "{expertName}", expert.Name)
	httpx.OK(w, resp)
}

func (s *Server) currentGameInviteConfig() gameInviteConfigDTO {
	var stored gameInviteConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameInviteConfigKey, &stored) {
		return normalizeGameInviteConfig(stored)
	}
	return defaultGameInviteConfig()
}

func (s *Server) currentReplayQuickActionsConfig() []replayQuickActionDTO {
	var stored []replayQuickActionDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameReplayQuickActionsConfigKey, &stored) {
		return normalizeReplayQuickActions(stored)
	}
	return defaultReplayQuickActions()
}

func (s *Server) currentReplayQuickMessagesConfig() []string {
	var stored []replayQuickMessageDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameReplayQuickMessagesConfigKey, &stored) {
		return normalizeReplayQuickMessages(stored)
	}
	return defaultReplayQuickMessages()
}

func defaultGameInviteConfig() gameInviteConfigDTO {
	return gameInviteConfigDTO{
		MinPlayerCount:      1,
		MaxPlayerCount:      1,
		BudgetMaxAmount:     99999999,
		DefaultBudget:       "800",
		DefaultTitle:        "产品架构梳理咨询",
		DefaultDetail:       "需要资深产品经理帮忙梳理B端产品架构，预计咨询时长2小时，涉及模块划分和数据流转设计。",
		PlayerIntroTemplate: "我帮你邀请了{expertName}，可以一起确认需求、预算和服务节奏。",
		ActivityTypes: []inviteActivityTypeDTO{
			{Key: "product", Name: "产品咨询"},
			{Key: "design", Name: "设计服务"},
			{Key: "tech", Name: "技术开发"},
		},
		RewardRateConfig: inviteRewardRateConfigDTO{
			PlatformServiceRate:   10,
			SystemGuideRewardRate: 10,
			InviteRewardRate:      40,
		},
	}
}

func normalizeGameInviteConfig(config gameInviteConfigDTO) gameInviteConfigDTO {
	fallback := defaultGameInviteConfig()
	if config.MinPlayerCount <= 0 {
		config.MinPlayerCount = fallback.MinPlayerCount
	}
	if config.MaxPlayerCount < config.MinPlayerCount {
		config.MaxPlayerCount = config.MinPlayerCount
	}
	if config.BudgetMaxAmount <= 0 {
		config.BudgetMaxAmount = fallback.BudgetMaxAmount
	}
	if strings.TrimSpace(config.DefaultBudget) == "" {
		config.DefaultBudget = fallback.DefaultBudget
	}
	if strings.TrimSpace(config.DefaultTitle) == "" {
		config.DefaultTitle = fallback.DefaultTitle
	}
	if strings.TrimSpace(config.DefaultDetail) == "" {
		config.DefaultDetail = fallback.DefaultDetail
	}
	if strings.TrimSpace(config.PlayerIntroTemplate) == "" {
		config.PlayerIntroTemplate = fallback.PlayerIntroTemplate
	}
	if len(config.ActivityTypes) == 0 {
		config.ActivityTypes = fallback.ActivityTypes
	}
	if config.RewardRateConfig.PlatformServiceRate <= 0 && config.RewardRateConfig.SystemGuideRewardRate <= 0 && config.RewardRateConfig.InviteRewardRate <= 0 {
		config.RewardRateConfig = fallback.RewardRateConfig
	}
	return config
}

func defaultReplayQuickActions() []replayQuickActionDTO {
	return []replayQuickActionDTO{
		{
			ID:       "same-friends",
			Theme:    "green",
			IconText: "\U0001f46b",
			Title:    "\u540c\u5c40\u597d\u53cb\u518d\u73a9\u4e00\u5c40",
			Desc:     "\u7acb\u5373\u9080\u8bf7\u4e0a\u4e00\u5c40\u6210\u5458",
			Route:    "confirm",
			Order:    10,
			Visible:  true,
		},
		{
			ID:       "smart-match",
			Theme:    "blue",
			IconText: "\U0001f916",
			Title:    "\u7cfb\u7edf\u63a8\u8350\u9002\u914d\u7ec4\u5c40",
			Desc:     "\u6839\u636e\u4f60\u7684\u504f\u597d\u8fd4\u56de\u9ad8\u5339\u914d\u5ea6\u5019\u9009\u5c40",
			Route:    "system_recommend",
			Order:    20,
			Visible:  true,
		},
		{
			ID:       "create-new",
			Theme:    "pink",
			IconType: "plus",
			Title:    "\u73a9\u5bb6\u521b\u5efa\u65b0\u5c40",
			Desc:     "\u81ea\u5b9a\u4e49\u9700\u6c42\uff0c\u5f00\u542f\u5168\u65b0\u7ec4\u5c40",
			Route:    "create",
			Order:    30,
			Visible:  true,
		},
	}
}

func normalizeReplayQuickActions(items []replayQuickActionDTO) []replayQuickActionDTO {
	if len(items) == 0 {
		return defaultReplayQuickActions()
	}
	allowedRoutes := map[string]bool{
		"confirm":          true,
		"system_recommend": true,
		"create":           true,
	}
	result := make([]replayQuickActionDTO, 0, len(items))
	for _, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Theme = strings.TrimSpace(item.Theme)
		item.Title = strings.TrimSpace(item.Title)
		item.Desc = strings.TrimSpace(item.Desc)
		item.Route = strings.TrimSpace(item.Route)
		if item.ID == "" || item.Title == "" || !allowedRoutes[item.Route] {
			continue
		}
		if !item.Visible && item.Order != 0 {
			continue
		}
		if item.Theme == "" {
			item.Theme = "green"
		}
		item.Visible = item.Visible || item.Order == 0
		result = append(result, item)
	}
	if len(result) == 0 {
		return defaultReplayQuickActions()
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})
	return result
}

func defaultReplayQuickMessages() []string {
	return []string{
		"\u518d\u6765\u4e00\u5c40\uff1f",
		"\u4e0a\u6b21\u5408\u4f5c\u5f88\u6109\u5feb\uff0c\u7ee7\u7eed\uff01",
		"\u6709\u4e2a\u65b0\u9700\u6c42\u60f3\u804a\u804a",
		"\u6709\u7a7a\u518d\u7ea6\u4e00\u5c40",
	}
}

func normalizeReplayQuickMessages(items []replayQuickMessageDTO) []string {
	if len(items) == 0 {
		return defaultReplayQuickMessages()
	}
	resultItems := make([]replayQuickMessageDTO, 0, len(items))
	for _, item := range items {
		item.Text = strings.TrimSpace(item.Text)
		if item.Text == "" {
			continue
		}
		if !item.Visible && item.Order != 0 {
			continue
		}
		item.Visible = item.Visible || item.Order == 0
		resultItems = append(resultItems, item)
	}
	if len(resultItems) == 0 {
		return defaultReplayQuickMessages()
	}
	sort.SliceStable(resultItems, func(i, j int) bool {
		return resultItems[i].Order < resultItems[j].Order
	})
	result := make([]string, 0, len(resultItems))
	for _, item := range resultItems {
		result = append(result, item.Text)
	}
	return result
}

func (s *Server) firstExpertCandidate(userID int64) inviteCandidateDTO {
	candidates := s.expertCandidates(userID, "")
	if len(candidates) > 0 {
		return candidates[0]
	}
	return inviteCandidateDTO{
		ID:         userID,
		UserID:     userID,
		Name:       s.displayName(userID, "行家"),
		Nickname:   s.displayName(userID, "行家"),
		AvatarText: avatarTextForName(s.displayName(userID, "行家"), userID),
		RoleType:   "expert",
		RoleLabel:  "行家",
		Desc:       "待完善行家资料",
		Tags:       []string{},
	}
}

func (s *Server) gameInviteRecentPlayers(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	items := s.inviteCandidates(userID, keyword, false)
	httpx.OK(w, map[string]interface{}{"list": items, "items": items, "total": len(items)})
}

func (s *Server) gameInvitePlayers(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	items := s.inviteCandidates(userID, keyword, true)
	httpx.OK(w, map[string]interface{}{"list": items, "items": items, "total": len(items)})
}

func (s *Server) gameInviteReplayContext(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	sourceGameID := parseFlexibleInt64(r.URL.Query().Get("sourceGameId"))
	game, ok := s.gameForInviteContext(sourceGameID, userID)
	if !ok {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "\u4e0a\u5c40\u7ec4\u5c40\u4e0d\u5b58\u5728\u6216\u65e0\u6743\u8bbf\u95ee")
		return
	}
	invitees := s.replayInvitees(game, userID)
	httpx.OK(w, map[string]interface{}{
		"sourceGameId":   game.ID,
		"serviceOrderId": "GAME-" + strconv.FormatInt(game.ID, 10),
		"title":          "\u592a\u68d2\u4e86\uff01\u4f60\u60f3\u600e\u4e48\u5f00\u542f\u4e0b\u4e00\u5c40\uff1f",
		"desc":           "\u6839\u636e\u4e0a\u4e00\u5c40\u548c\u540e\u53f0\u914d\u7f6e\u8fd4\u56de\u53ef\u7528\u65b9\u5f0f",
		"inviter": map[string]interface{}{
			"id":        userID,
			"name":      s.displayName(userID, "\u9080\u8bf7\u4eba"),
			"roleType":  "guide",
			"roleLabel": "\u9886\u8def\u4eba",
		},
		"previousSession": map[string]interface{}{
			"serviceType":     game.Title,
			"completedAtText": inviteTimeText(game),
			"participantText": strconv.Itoa(len(s.games.Members(game.ID))) + "\u4eba",
		},
		"invitees":      invitees,
		"quickActions":  s.currentReplayQuickActionsConfig(),
		"quickMessages": s.currentReplayQuickMessagesConfig(),
	})
}

func (s *Server) gameInviteSystemRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	config := s.currentGameSystemRecommendationsConfig()
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category == "" {
		category = config.DefaultCategory
	}
	experts := s.expertCandidates(userID, category)
	defaultSelected := []string{}
	if len(experts) > 0 {
		defaultSelected = append(defaultSelected, strconv.FormatInt(experts[0].UserID, 10))
		experts[0].Selected = true
	}
	httpx.OK(w, map[string]interface{}{
		"recommendationId":         "REC-" + strconv.FormatInt(time.Now().Unix(), 10),
		"title":                    config.Title,
		"desc":                     config.Desc,
		"loadingText":              config.LoadingText,
		"emptyText":                config.EmptyText,
		"summaryTemplate":          config.SummaryTemplate,
		"summaryDesc":              config.SummaryDesc,
		"cancelText":               config.CancelText,
		"confirmText":              config.ConfirmText,
		"minSelectToast":           config.MinSelectToast,
		"confirmingText":           config.ConfirmingText,
		"categories":               config.Categories,
		"activeCategory":           category,
		"defaultSelectedExpertIds": defaultSelected,
		"experts":                  experts,
	})
}

func (s *Server) currentGameSystemRecommendationsConfig() gameSystemRecommendationsConfigDTO {
	var stored gameSystemRecommendationsConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameSystemRecommendationsConfigKey, &stored) && strings.TrimSpace(stored.Title) != "" {
		return normalizeGameSystemRecommendationsConfig(stored)
	}
	return defaultGameSystemRecommendationsConfig()
}

func (s *Server) currentGameReferralRecordsConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(gameReferralRecordsConfigKey, &config) && len(config) > 0 {
		return config
	}
	return defaultGameReferralRecordsConfig()
}

func defaultGameReferralRecordsConfig() map[string]interface{} {
	return map[string]interface{}{
		"pageTitle": "我的引荐记录",
		"summary": map[string]interface{}{
			"label":      "本月引荐收益",
			"background": "linear-gradient(135deg, #ffb347 0%, #ff7b00 100%)",
			"iconSrc":    "/pages/game/referral-record/assets/wallet.png",
			"statTemplates": map[string]string{
				"success":    "成功 {count}单",
				"processing": "进行中 {count}单",
				"review":     "待评价 {count}单",
			},
		},
		"tabs": []map[string]string{
			{"key": "processing", "label": "进行中"},
			{"key": "completed", "label": "已完成"},
			{"key": "canceled", "label": "已取消"},
		},
		"texts": map[string]string{
			"emptyText":             "暂无引荐记录",
			"loadFailedText":        "引荐记录加载失败",
			"expertRoleText":        "行家",
			"playerRoleText":        "玩家",
			"selfLabel":             "我",
			"reviewedTagText":       "已评价",
			"pendingReviewTagText":  "待评价",
			"reviewedActionText":    "已评价",
			"reviewActionText":      "评价双方",
			"remindActionText":      "提醒交付",
			"chatActionText":        "查看群聊",
			"chatPrefill":           "你好，我想查看本次引荐服务的群聊进度。",
			"unavailableText":       "该操作暂不可用",
			"remindMessageTemplate": "请及时确认交付：{serviceTitle}",
			"remindServiceFallback": "引荐服务",
			"remindSuccessText":     "已提醒交付",
			"remindFailedText":      "提醒交付失败",
			"rewardPrefix":          "¥",
		},
	}
}

func stringFromMap(values map[string]interface{}, key string) string {
	value, _ := values[key].(string)
	return value
}

func nestedStringMap(values map[string]interface{}, key string) map[string]string {
	result := map[string]string{}
	switch raw := values[key].(type) {
	case map[string]string:
		for itemKey, itemValue := range raw {
			result[itemKey] = itemValue
		}
	case map[string]interface{}:
		for itemKey, itemValue := range raw {
			if text, ok := itemValue.(string); ok {
				result[itemKey] = text
			}
		}
	}
	return result
}

func referralConfigTexts(config map[string]interface{}) map[string]string {
	return nestedStringMap(config, "texts")
}

func referralSummaryConfig(config map[string]interface{}) map[string]interface{} {
	switch summary := config["summary"].(type) {
	case map[string]interface{}:
		return summary
	case map[string]string:
		result := map[string]interface{}{}
		for key, value := range summary {
			result[key] = value
		}
		return result
	default:
		return map[string]interface{}{}
	}
}

func referralTabsConfig(config map[string]interface{}) []map[string]string {
	tabs := []map[string]string{}
	switch rawTabs := config["tabs"].(type) {
	case []map[string]string:
		tabs = append(tabs, rawTabs...)
	case []interface{}:
		tabs = make([]map[string]string, 0, len(rawTabs))
		for _, rawTab := range rawTabs {
			tabMap, _ := rawTab.(map[string]interface{})
			key, _ := tabMap["key"].(string)
			label, _ := tabMap["label"].(string)
			if key != "" {
				tabs = append(tabs, map[string]string{"key": key, "label": label})
			}
		}
	}
	if len(tabs) > 0 {
		return tabs
	}
	return []map[string]string{{"key": "processing", "label": "进行中"}, {"key": "completed", "label": "已完成"}, {"key": "canceled", "label": "已取消"}}
}

func applyCountTemplate(template string, count int) string {
	return strings.ReplaceAll(template, "{count}", strconv.Itoa(count))
}

func normalizeGameSystemRecommendationsConfig(config gameSystemRecommendationsConfigDTO) gameSystemRecommendationsConfigDTO {
	defaults := defaultGameSystemRecommendationsConfig()
	if strings.TrimSpace(config.Title) == "" {
		config.Title = defaults.Title
	}
	if strings.TrimSpace(config.Desc) == "" {
		config.Desc = defaults.Desc
	}
	if strings.TrimSpace(config.LoadingText) == "" {
		config.LoadingText = defaults.LoadingText
	}
	if strings.TrimSpace(config.EmptyText) == "" {
		config.EmptyText = defaults.EmptyText
	}
	if strings.TrimSpace(config.SummaryTemplate) == "" {
		config.SummaryTemplate = defaults.SummaryTemplate
	}
	if strings.TrimSpace(config.SummaryDesc) == "" {
		config.SummaryDesc = defaults.SummaryDesc
	}
	if strings.TrimSpace(config.CancelText) == "" {
		config.CancelText = defaults.CancelText
	}
	if strings.TrimSpace(config.ConfirmText) == "" {
		config.ConfirmText = defaults.ConfirmText
	}
	if strings.TrimSpace(config.MinSelectToast) == "" {
		config.MinSelectToast = defaults.MinSelectToast
	}
	if strings.TrimSpace(config.ConfirmingText) == "" {
		config.ConfirmingText = defaults.ConfirmingText
	}
	if strings.TrimSpace(config.DefaultCategory) == "" {
		config.DefaultCategory = defaults.DefaultCategory
	}
	if len(config.Categories) == 0 {
		config.Categories = defaults.Categories
	}
	if strings.TrimSpace(config.Version) == "" {
		config.Version = defaults.Version
	}
	return config
}

func defaultGameSystemRecommendationsConfig() gameSystemRecommendationsConfigDTO {
	return gameSystemRecommendationsConfigDTO{
		Title:           "系统推荐适配局",
		Desc:            "暂无推荐结果，请完善资料或稍后重试",
		LoadingText:     "推荐数据加载中",
		EmptyText:       "暂无匹配行家",
		SummaryTemplate: "已选择 {count} 位行家",
		SummaryDesc:     "还可以选择多位行家组成顾问团，或搭配玩家共同组局",
		CancelText:      "取消",
		ConfirmText:     "确认组局",
		MinSelectToast:  "请选择至少一位行家",
		ConfirmingText:  "正在进入组局",
		Categories: []gameSystemRecommendationCategoryDTO{
			{Key: "all", Name: "全部"},
			{Key: "product", Name: "产品架构"},
			{Key: "tech", Name: "技术咨询"},
			{Key: "operation", Name: "运营策略"},
		},
		DefaultCategory: "all",
		Version:         "2026-07-01",
	}
}

func (s *Server) createReplayGameInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		SourceGameID interface{} `json:"sourceGameId"`
		Message      string      `json:"message"`
		Invitees     []struct {
			ID       interface{} `json:"id"`
			RoleType string      `json:"roleType"`
		} `json:"invitees"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	sourceGameID := parseFlexibleInt64(req.SourceGameID)
	if sourceGameID <= 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "source game required")
		return
	}
	game, err := s.games.Get(sourceGameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.canViewInviteGame(userID, game) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	targetIDs := make([]int64, 0, len(req.Invitees))
	for _, invitee := range req.Invitees {
		targetID := parseFlexibleInt64(invitee.ID)
		if targetID > 0 && targetID != userID {
			targetIDs = append(targetIDs, targetID)
		}
	}
	if len(targetIDs) == 0 {
		for _, candidate := range s.inviteCandidates(userID, "", false) {
			targetIDs = append(targetIDs, candidate.UserID)
			break
		}
	}
	items := make([]games.Invitation, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		invitation, err := s.games.CreateInvitation(userID, sourceGameID, games.InvitationRequest{TargetUserID: targetID, Message: req.Message})
		if err != nil {
			writeGameError(w, err)
			return
		}
		items = append(items, invitation)
	}
	s.recordBehavior(userID, "create_replay_invitation", "game", sourceGameID, map[string]interface{}{"count": len(items)})
	httpx.OK(w, map[string]interface{}{
		"replayInvitationId": firstInvitationID(items),
		"invitationId":       firstInvitationID(items),
		"items":              items,
		"count":              len(items),
	})
}

func (s *Server) createGameInviteReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		InvitationID interface{} `json:"invitationId"`
		GameID       interface{} `json:"gameId"`
		RemindTarget string      `json:"remindTarget"`
		Message      string      `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	invitationID := parseFlexibleInt64(req.InvitationID)
	gameID := parseFlexibleInt64(req.GameID)
	invitation := s.findInviteForReminder(userID, invitationID, gameID)
	if invitation.ID == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invitation not found")
		return
	}
	targetUserID := invitation.TargetUserID
	if targetUserID == userID {
		targetUserID = invitation.InviterID
	}
	if targetUserID <= 0 || targetUserID == userID {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid reminder target")
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = "\u8bf7\u5c3d\u5feb\u786e\u8ba4\u672c\u6b21\u7ec4\u5c40\u9080\u8bf7"
	}
	if len(message) > 300 {
		message = message[:300]
	}
	notice := s.notices.Create(notifications.CreateRequest{
		UserID:     targetUserID,
		NotifyType: "game_invitation_remind",
		Title:      "\u7ec4\u5c40\u786e\u8ba4\u63d0\u9192",
		Content:    message,
		BizType:    "game_invitation",
		BizID:      invitation.ID,
		NeedWechat: true,
		WechatData: map[string]string{
			"thing1": message,
			"page":   "pages/game/guide-progress-detail/index?invitationId=" + strconv.FormatInt(invitation.ID, 10),
		},
	})
	s.recordBehavior(userID, "game_invitation_remind", "game_invitation", invitation.ID, map[string]interface{}{
		"gameId":       invitation.GameID,
		"targetUserId": targetUserID,
		"remindTarget": strings.TrimSpace(req.RemindTarget),
	})
	httpx.OK(w, map[string]interface{}{
		"invitationId": invitation.ID,
		"gameId":       invitation.GameID,
		"targetUserId": targetUserID,
		"notification": notice,
	})
}

func (s *Server) gameInviteGuideProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	active := make([]map[string]interface{}, 0)
	completed := make([]map[string]interface{}, 0)
	for _, invitation := range s.games.InvitationsForUser(userID) {
		item := s.invitationProgressItem(invitation)
		switch invitation.Status {
		case "pending":
			active = append(active, item)
		default:
			completed = append(completed, item)
		}
	}
	httpx.OK(w, map[string]interface{}{
		"activeCount":      len(active),
		"activeParties":    active,
		"completedParties": completed,
	})
}

func (s *Server) gameInviteReferralRecords(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	pageConfig := s.currentGameReferralRecordsConfig()
	records := make([]map[string]interface{}, 0)
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if invitation.InviterID != userID {
			continue
		}
		records = append(records, s.referralRecordItem(invitation, referralConfigTexts(pageConfig)))
	}
	counts := map[string]int{"processing": 0, "completed": 0, "canceled": 0}
	totalReward := 0
	for _, record := range records {
		state, _ := record["state"].(string)
		counts[state]++
		if reward, ok := record["reward"].(int); ok {
			totalReward += reward
		}
	}
	summaryConfig := referralSummaryConfig(pageConfig)
	summaryTemplates := nestedStringMap(summaryConfig, "statTemplates")
	tabsConfig := referralTabsConfig(pageConfig)
	tabs := make([]map[string]interface{}, 0, len(tabsConfig))
	for _, tab := range tabsConfig {
		tabs = append(tabs, map[string]interface{}{
			"key":   tab["key"],
			"label": tab["label"],
			"count": counts[tab["key"]],
		})
	}
	httpx.OK(w, map[string]interface{}{
		"summary": map[string]interface{}{
			"label":      stringFromMap(summaryConfig, "label"),
			"amount":     totalReward,
			"background": stringFromMap(summaryConfig, "background"),
			"iconSrc":    stringFromMap(summaryConfig, "iconSrc"),
			"stats": []map[string]interface{}{
				{"key": "success", "text": applyCountTemplate(summaryTemplates["success"], counts["completed"])},
				{"key": "processing", "text": applyCountTemplate(summaryTemplates["processing"], counts["processing"])},
				{"key": "review", "text": applyCountTemplate(summaryTemplates["review"], counts["completed"])},
			},
		},
		"tabs":       tabs,
		"records":    records,
		"total":      len(records),
		"pageConfig": pageConfig,
	})
}

func (s *Server) gameInviteCancelDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	invitation := s.findInviteForCancelDetail(userID, r)
	if invitation.ID == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invitation not found")
		return
	}
	game, _ := s.games.Get(invitation.GameID)
	cancelUserID := invitation.TargetUserID
	if invitation.Status != "rejected" {
		cancelUserID = invitation.InviterID
	}
	httpx.OK(w, map[string]interface{}{
		"id":          invitation.ID,
		"statusTitle": "\u7ec4\u5c40\u5df2\u53d6\u6d88",
		"statusDesc":  "\u672c\u6b21\u7ec4\u5c40\u9080\u8bf7\u5df2\u53d6\u6d88",
		"canceledBy": map[string]interface{}{
			"id":        cancelUserID,
			"name":      s.displayName(cancelUserID, "\u6210\u5458"),
			"roleType":  "member",
			"roleLabel": "\u6210\u5458",
		},
		"reason":          map[string]string{"title": "\u9080\u8bf7\u672a\u5b8c\u6210", "desc": invitation.Message},
		"message":         invitation.Message,
		"messageTimeText": inviteRespondedText(invitation),
		"timeline": []map[string]interface{}{
			{"key": "invite", "title": "\u53d1\u8d77\u9080\u8bf7", "desc": game.Title, "timeText": inviteTimeText(game), "state": "active"},
			{"key": "cancel", "title": "\u7ec4\u5c40\u53d6\u6d88", "desc": "\u9080\u8bf7\u72b6\u6001\uff1a" + invitation.Status, "timeText": inviteRespondedText(invitation), "state": "error"},
		},
	})
}

func (s *Server) inviteCandidates(userID int64, keyword string, includeAllUsers bool) []inviteCandidateDTO {
	seen := map[int64]bool{userID: true}
	items := make([]inviteCandidateDTO, 0)
	add := func(candidateID int64, tag string, score int) {
		if candidateID <= 0 || seen[candidateID] {
			return
		}
		name := s.displayName(candidateID, "\u73a9\u5bb6")
		if keyword != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) && !strings.Contains(strconv.FormatInt(candidateID, 10), keyword) {
			return
		}
		seen[candidateID] = true
		items = append(items, inviteCandidateDTO{
			ID:         candidateID,
			UserID:     candidateID,
			Name:       name,
			Nickname:   name,
			AvatarText: avatarTextForName(name, candidateID),
			Tag:        tag,
			Desc:       "\u6765\u81ea\u771f\u5b9e\u5173\u7cfb\u6216\u7ec4\u5c40\u8bb0\u5f55",
			Meta:       "\u5173\u7cfb\u5f3a\u5ea6 " + strconv.Itoa(score),
			RoleType:   "player",
			RoleLabel:  "\u73a9\u5bb6",
		})
	}
	for _, conn := range s.connections.My(userID) {
		add(conn.ConnectedUserID, conn.RelationType, conn.StrengthScore)
	}
	for _, game := range s.games.List() {
		if !s.canViewInviteGame(userID, game) {
			continue
		}
		for _, memberID := range s.games.Members(game.ID) {
			add(memberID, "\u540c\u5c40\u73a9\u5bb6", 2)
		}
	}
	if includeAllUsers {
		if allUsers, err := s.auth.AdminUsers(users.Filter{}); err == nil {
			for _, user := range allUsers {
				add(user.ID, "\u5e73\u53f0\u7528\u6237", 1)
			}
		}
	}
	return items
}

func (s *Server) expertCandidates(userID int64, category string) []inviteCandidateDTO {
	items := make([]inviteCandidateDTO, 0)
	for _, profile := range s.profiles.AllExpertSkills() {
		if profile.UserID == userID {
			continue
		}
		tags := append([]string{}, profile.SkillTree...)
		tags = append(tags, profile.ServiceTags...)
		cat := categoryForTags(tags)
		if category != "" && category != "all" && cat != category {
			continue
		}
		name := s.displayName(profile.UserID, "\u884c\u5bb6")
		match := 70 + profile.Completeness/4
		if match > 98 {
			match = 98
		}
		items = append(items, inviteCandidateDTO{
			ID:          profile.UserID,
			UserID:      profile.UserID,
			Name:        name,
			Nickname:    name,
			AvatarText:  avatarTextForName(name, profile.UserID),
			RoleType:    "expert",
			RoleLabel:   "\u884c\u5bb6",
			Category:    cat,
			Tags:        tags,
			Desc:        strings.Join(tags, " / "),
			Match:       match,
			ReviewCount: s.games.StatsForUser(profile.UserID).Completed,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Match > items[j].Match
	})
	return items
}

func (s *Server) replayInvitees(game games.Game, userID int64) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, memberID := range s.games.Members(game.ID) {
		if memberID == userID {
			continue
		}
		name := s.displayName(memberID, "\u6210\u5458")
		roleType := "player"
		roleLabel := "\u73a9\u5bb6"
		if game.MainGuideUserID == memberID {
			roleType = "expert"
			roleLabel = "\u884c\u5bb6"
		}
		result = append(result, map[string]interface{}{
			"id":        memberID,
			"userId":    memberID,
			"name":      name,
			"roleType":  roleType,
			"roleLabel": roleLabel,
			"desc":      game.Title,
			"selected":  true,
		})
	}
	if len(result) == 0 {
		for _, candidate := range s.inviteCandidates(userID, "", false) {
			result = append(result, map[string]interface{}{
				"id":        candidate.UserID,
				"userId":    candidate.UserID,
				"name":      candidate.Name,
				"roleType":  "player",
				"roleLabel": "\u73a9\u5bb6",
				"desc":      candidate.Desc,
				"selected":  true,
			})
			break
		}
	}
	return result
}

func (s *Server) invitationProgressItem(invitation games.Invitation) map[string]interface{} {
	game, _ := s.games.Get(invitation.GameID)
	targetName := s.displayName(invitation.TargetUserID, "\u73a9\u5bb6")
	inviterName := s.displayName(invitation.InviterID, "\u9080\u8bf7\u4eba")
	statusText := map[string]string{"pending": "\u5f85\u786e\u8ba4", "accepted": "\u5df2\u786e\u8ba4", "rejected": "\u5df2\u62d2\u7edd"}[invitation.Status]
	if statusText == "" {
		statusText = invitation.Status
	}
	return map[string]interface{}{
		"id":              invitation.ID,
		"invitationId":    invitation.ID,
		"status":          invitation.Status,
		"statusText":      statusText,
		"timeText":        invitation.CreatedAt.Format("01-02 15:04"),
		"gameTitle":       game.Title,
		"players":         []map[string]interface{}{{"id": invitation.TargetUserID, "name": targetName, "roleType": "player", "status": invitation.Status}},
		"experts":         []map[string]interface{}{{"id": invitation.InviterID, "name": inviterName, "roleType": "expert", "status": "accepted"}},
		"progressPercent": inviteProgressPercent(invitation.Status),
		"progressText":    statusText,
		"title":           "\u7ec4\u5c40" + statusText,
		"memberText":      targetName + " \u4e0e " + inviterName,
		"reason":          invitation.Message,
	}
}

func (s *Server) referralRecordItem(invitation games.Invitation, texts map[string]string) map[string]interface{} {
	game, _ := s.games.Get(invitation.GameID)
	state := "processing"
	stateText := "\u670d\u52a1\u8fdb\u884c\u4e2d"
	stateClass := "blue"
	rewardClass := "orange"
	actions := []map[string]interface{}{
		{"key": "remind", "text": texts["remindActionText"], "theme": "primary"},
		{"key": "chat", "text": texts["chatActionText"], "theme": "plain"},
	}
	if invitation.Status == "accepted" || game.Status == "pending_review" || game.Status == "completed" {
		state = "completed"
		stateText = "\u5df2\u5b8c\u6210"
		stateClass = "green"
		rewardClass = "green"
		actions = []map[string]interface{}{{"key": "review", "text": texts["reviewActionText"], "theme": "highlight"}}
	}
	if invitation.Status == "rejected" {
		state = "canceled"
		stateText = "\u5df2\u53d6\u6d88"
		stateClass = "gray"
		rewardClass = "gray"
		actions = nil
	}
	targetName := s.displayName(invitation.TargetUserID, "\u73a9\u5bb6")
	inviterName := s.displayName(invitation.InviterID, "\u884c\u5bb6")
	reward := 0
	if state != "canceled" {
		reward = int(successFundAmount(game) / 1000)
	}
	return map[string]interface{}{
		"id":               "REF-" + strconv.FormatInt(invitation.ID, 10),
		"referralId":       invitation.ID,
		"invitationId":     invitation.ID,
		"gameId":           invitation.GameID,
		"state":            state,
		"stateText":        stateText,
		"stateClass":       stateClass,
		"timeText":         invitation.CreatedAt.Format("01-02 15:04"),
		"expertUserId":     invitation.InviterID,
		"playerUserId":     invitation.TargetUserID,
		"expert":           map[string]interface{}{"id": invitation.InviterID, "userId": invitation.InviterID, "name": inviterName, "avatarText": avatarTextForName(inviterName, invitation.InviterID), "avatarClass": "blue"},
		"player":           map[string]interface{}{"id": invitation.TargetUserID, "userId": invitation.TargetUserID, "name": targetName, "avatarText": avatarTextForName(targetName, invitation.TargetUserID), "avatarClass": "pink"},
		"serviceTitle":     game.Title,
		"reward":           reward,
		"rewardClass":      rewardClass,
		"noticeText":       referralNoticeText(state, invitation),
		"reviewStatus":     referralReviewStatus(state),
		"reviewActionText": texts["reviewActionText"],
		"actions":          actions,
	}
}

func referralNoticeText(state string, invitation games.Invitation) string {
	switch state {
	case "processing":
		return "\u9884\u8ba1\u4ea4\u4ed8\u65f6\u95f4\uff1a" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	case "canceled":
		return "\u53d6\u6d88\u539f\u56e0\uff1a" + invitation.Message
	default:
		return ""
	}
}

func referralReviewStatus(state string) string {
	if state == "completed" {
		return "pending"
	}
	return ""
}

func (s *Server) findInviteForCancelDetail(userID int64, r *http.Request) games.Invitation {
	id := parseFlexibleInt64(r.URL.Query().Get("id"))
	if id <= 0 {
		id = parseFlexibleInt64(r.URL.Query().Get("invitationId"))
	}
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if id <= 0 || invitation.ID == id {
			if invitation.Status != "accepted" {
				return invitation
			}
		}
	}
	return games.Invitation{}
}

func (s *Server) findInviteForReminder(userID int64, invitationID int64, gameID int64) games.Invitation {
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if invitationID > 0 && invitation.ID != invitationID {
			continue
		}
		if gameID > 0 && invitation.GameID != gameID {
			continue
		}
		return invitation
	}
	return games.Invitation{}
}

func (s *Server) gameForInviteContext(gameID int64, userID int64) (games.Game, bool) {
	if gameID > 0 {
		if game, err := s.games.Get(gameID); err == nil && s.canViewInviteGame(userID, game) {
			return game, true
		}
	}
	return games.Game{}, false
}

func (s *Server) canViewInviteGame(userID int64, game games.Game) bool {
	return game.CreatorUserID == userID || game.MainGuideUserID == userID || s.games.IsMember(game.ID, userID)
}

func (s *Server) displayName(userID int64, fallback string) string {
	if user, ok := s.auth.UserByID(userID); ok {
		if strings.TrimSpace(user.Nickname) != "" {
			return strings.TrimSpace(user.Nickname)
		}
	}
	if fallback == "" {
		fallback = "\u7528\u6237"
	}
	return fallback + strconv.FormatInt(userID, 10)
}

func parseFlexibleInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		id, _ := v.Int64()
		return id
	case string:
		if id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return id
		}
	}
	return 0
}

func avatarTextForName(name string, id int64) string {
	name = strings.TrimSpace(name)
	if name != "" {
		runes := []rune(name)
		if len(runes) >= 2 {
			return string(runes[:2])
		}
		return string(runes)
	}
	return "U" + strconv.FormatInt(id, 10)
}

func categoryForTags(tags []string) string {
	joined := strings.ToLower(strings.Join(tags, " "))
	switch {
	case strings.Contains(joined, "tech") || strings.Contains(joined, "\u6280\u672f"):
		return "tech"
	case strings.Contains(joined, "operation") || strings.Contains(joined, "\u8fd0\u8425"):
		return "operation"
	default:
		return "product"
	}
}

func inviteProgressPercent(status string) int {
	switch status {
	case "accepted":
		return 100
	case "rejected":
		return 100
	default:
		return 50
	}
}

func inviteTimeText(game games.Game) string {
	if !game.CreatedAt.IsZero() {
		return game.CreatedAt.Format("2006-01-02 15:04")
	}
	return time.Now().Format("2006-01-02 15:04")
}

func inviteRespondedText(invitation games.Invitation) string {
	if invitation.RespondedAt != "" {
		return invitation.RespondedAt
	}
	return invitation.CreatedAt.Format("01-02 15:04")
}

func firstInvitationID(items []games.Invitation) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[0].ID
}
