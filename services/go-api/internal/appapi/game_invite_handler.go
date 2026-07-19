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
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/users"
)

type inviteCandidateDTO struct {
	ID          int64    `json:"id"`
	UserID      int64    `json:"userId"`
	Name        string   `json:"name"`
	Nickname    string   `json:"nickname"`
	RealName    string   `json:"realName"`
	DisplayName string   `json:"displayName"`
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
	InGame      bool     `json:"inGame,omitempty"`
	CanSelect   bool     `json:"canSelect,omitempty"`
}

const gameInviteConfigKey = "game.invite_config"
const gameReplayQuickActionsConfigKey = "game.replay_quick_actions"
const gameReplayQuickMessagesConfigKey = "game.replay_quick_messages"
const gameSystemRecommendationsConfigKey = "game.system_recommendations_config"
const gameReferralRecordsConfigKey = "game.referral_records_config"

const guideProgressRemindIconURL = "https://static.haowan.net.cn/miniprogram/pages/game/referral-record/assets/action-bell-blue.svg"
const guideProgressConfirmedIconURL = "https://static.haowan.net.cn/miniprogram/components/game-detail/party-card/icon-confirmed.svg"
const guideProgressWaitingIconURL = "https://static.haowan.net.cn/miniprogram/pages/game/guide-progress-detail/assets/participant-waiting.svg"

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
	MinPlayerCount           int                       `json:"minPlayerCount"`
	MaxPlayerCount           int                       `json:"maxPlayerCount"`
	BudgetMaxAmount          int                       `json:"budgetMaxAmount"`
	InvitationTimeoutMinutes int                       `json:"invitationTimeoutMinutes"`
	DefaultBudget            string                    `json:"defaultBudget"`
	DefaultTitle             string                    `json:"defaultTitle"`
	DefaultDetail            string                    `json:"defaultDetail"`
	PlayerIntroTemplate      string                    `json:"playerIntroTemplate"`
	ActivityTypes            []inviteActivityTypeDTO   `json:"activityTypes"`
	RewardRateConfig         inviteRewardRateConfigDTO `json:"rewardRateConfig"`
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
	sourceGameID := parseFlexibleInt64(r.URL.Query().Get("sourceGameId"))
	if sourceGameID <= 0 {
		sourceGameID = parseFlexibleInt64(r.URL.Query().Get("gameId"))
	}
	if sourceGameID > 0 {
		game, found := s.gameForInviteContext(sourceGameID, userID)
		if !found {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "\u4e0a\u5c40\u7ec4\u5c40\u4e0d\u5b58\u5728\u6216\u65e0\u6743\u8bbf\u95ee")
			return
		}
		for key, value := range inviteSourceGameConfig(game) {
			resp[key] = value
		}
	}
	resp["expert"] = expert
	resp["experts"] = s.expertCandidates(userID, "")
	resp["playerIntroTemplate"] = strings.ReplaceAll(config.PlayerIntroTemplate, "{expertName}", expert.Name)
	httpx.OK(w, resp)
}

func inviteSourceGameConfig(game games.Game) map[string]interface{} {
	categoryText := strings.TrimSpace(game.PrimaryCategoryText)
	if categoryText == "" {
		categoryText = strings.TrimSpace(game.SecondaryCategoryText)
	}
	if categoryText == "" {
		categoryText = strings.TrimSpace(game.GameType)
	}
	locationText := strings.TrimSpace(game.Address)
	if locationText == "" {
		locationText = strings.TrimSpace(game.CityName)
	}
	timeText, durationText := invitationGameSchedule(game)
	amountCent := int64(game.Price*100 + 0.5)
	defaultBudget := strconv.FormatInt(amountCent/100, 10)
	feeType := "付费局"
	if game.GameType == "free" || amountCent <= 0 {
		feeType = "免费局"
	}
	rows := make([]map[string]string, 0, 5)
	appendRow := func(label string, value string) {
		if strings.TrimSpace(value) != "" {
			rows = append(rows, map[string]string{"label": label, "value": value})
		}
	}
	appendRow("主题", game.Title)
	appendRow("类型", categoryText)
	appendRow("时间", timeText)
	appendRow("地点", locationText)
	appendRow("费用", feeType)
	pricingRows := []map[string]string{{"label": "费用类型", "value": feeType}}
	totalLabel := ""
	totalValue := ""
	if amountCent > 0 && game.GameType != "free" {
		amountText := "¥" + defaultBudget
		pricingRows = append(pricingRows, map[string]string{"label": "预算金额", "value": amountText})
		totalLabel = "行家实收"
		totalValue = amountText
	}
	return map[string]interface{}{
		"defaultTitle":  game.Title,
		"defaultDetail": game.Description,
		"defaultBudget": defaultBudget,
		"sourceGame": map[string]interface{}{
			"id":               game.ID,
			"title":            game.Title,
			"description":      game.Description,
			"price":            game.Price,
			"startAt":          game.StartAt,
			"endAt":            game.EndAt,
			"serviceDuration":  durationText,
			"durationText":     durationText,
			"expectedTime":     timeText,
			"expectedTimeText": timeText,
			"locationText":     locationText,
		},
		"sourceGameRows": rows,
		"pricingDisplay": map[string]interface{}{
			"enabled":    true,
			"free":       amountCent <= 0 || game.GameType == "free",
			"title":      "费用信息",
			"rows":       pricingRows,
			"totalLabel": totalLabel,
			"totalValue": totalValue,
		},
	}
}

func parseInvitationGameTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04"} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func invitationDurationText(duration time.Duration) string {
	totalMinutes := int(duration / time.Minute)
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if hours <= 0 {
		return strconv.Itoa(minutes) + "分钟"
	}
	if minutes <= 0 {
		return strconv.Itoa(hours) + "小时"
	}
	return strconv.Itoa(hours) + "小时" + strconv.Itoa(minutes) + "分钟"
}

func invitationGameSchedule(game games.Game) (string, string) {
	startAt, hasStart := parseInvitationGameTime(game.StartAt)
	endAt, hasEnd := parseInvitationGameTime(game.EndAt)
	if !hasStart || !hasEnd || !endAt.After(startAt) {
		return firstNonEmpty(strings.TrimSpace(game.StartAt), strings.TrimSpace(game.EndAt)), ""
	}
	timeText := startAt.Format("2006-01-02 15:04")
	if startAt.Format("2006-01-02") == endAt.Format("2006-01-02") {
		timeText += " - " + endAt.Format("15:04")
	} else {
		timeText += " - " + endAt.Format("2006-01-02 15:04")
	}
	return timeText, invitationDurationText(endAt.Sub(startAt))
}

func (s *Server) currentGameInviteConfig() gameInviteConfigDTO {
	var stored gameInviteConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameInviteConfigKey, &stored) {
		config := normalizeGameInviteConfig(stored)
		rules := s.currentOperationRules().Invite
		if rules.TimeoutMinutes > 0 {
			config.InvitationTimeoutMinutes = rules.TimeoutMinutes
		}
		return config
	}
	config := defaultGameInviteConfig()
	config.InvitationTimeoutMinutes = s.currentOperationRules().Invite.TimeoutMinutes
	return config
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
		MinPlayerCount:           1,
		MaxPlayerCount:           1,
		BudgetMaxAmount:          99999999,
		InvitationTimeoutMinutes: 24 * 60,
		DefaultBudget:            "0",
		DefaultTitle:             "产品架构梳理咨询",
		DefaultDetail:            "需要资深产品经理帮忙梳理B端产品架构，预计咨询时长2小时，涉及模块划分和数据流转设计。",
		PlayerIntroTemplate:      "我帮你邀请了{expertName}，可以一起确认需求、预算和服务节奏。",
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
	if config.InvitationTimeoutMinutes <= 0 {
		config.InvitationTimeoutMinutes = fallback.InvitationTimeoutMinutes
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
	result := make([]replayQuickActionDTO, 0, 3)
	seenRoutes := make(map[string]bool, 3)
	for _, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Theme = strings.TrimSpace(item.Theme)
		item.Title = strings.TrimSpace(item.Title)
		item.Desc = strings.TrimSpace(item.Desc)
		item.Route = strings.TrimSpace(item.Route)
		if item.ID == "" || item.Title == "" || !allowedRoutes[item.Route] {
			continue
		}
		if seenRoutes[item.Route] {
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
		seenRoutes[item.Route] = true
	}
	if len(result) == 0 {
		return defaultReplayQuickActions()
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})
	// 一期固定提供三个入口；后台隐藏或误配某一项时，用默认文案补齐，
	// 避免用户只看到一张卡片或把流程误判为没有下一步。
	for _, fallback := range defaultReplayQuickActions() {
		if len(result) >= 3 {
			break
		}
		if seenRoutes[fallback.Route] {
			continue
		}
		result = append(result, fallback)
		seenRoutes[fallback.Route] = true
	}
	if len(result) > 3 {
		result = result[:3]
	}
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
		ID:          userID,
		UserID:      userID,
		Name:        s.inGameDisplayName(userID, "行家"),
		Nickname:    s.inGameDisplayName(userID, "行家"),
		RealName:    s.inGameDisplayName(userID, "行家"),
		DisplayName: s.inGameDisplayName(userID, "行家"),
		AvatarText:  avatarTextForName(s.inGameDisplayName(userID, "行家"), userID),
		RoleType:    "expert",
		RoleLabel:   "行家",
		Desc:        "待完善行家资料",
		Tags:        []string{},
	}
}

func (s *Server) gameInviteRecentPlayers(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	gameID := parseFlexibleInt64(r.URL.Query().Get("sourceGameId"))
	if gameID <= 0 {
		gameID = parseFlexibleInt64(r.URL.Query().Get("gameId"))
	}
	excludedUserID := parseFlexibleInt64(r.URL.Query().Get("excludeUserId"))
	if excludedUserID <= 0 {
		excludedUserID = parseFlexibleInt64(r.URL.Query().Get("expertUserId"))
	}
	items := s.inviteCandidates(userID, keyword, false, gameID, excludedUserID)
	httpx.OK(w, map[string]interface{}{"list": items, "items": items, "total": len(items)})
}

func (s *Server) gameInvitePermission(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID := parseFlexibleInt64(r.URL.Query().Get("sourceGameId"))
	if gameID <= 0 {
		gameID = parseFlexibleInt64(r.URL.Query().Get("gameId"))
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	allowed := game.CreatorUserID == userID || (game.MainGuideUserID > 0 && game.MainGuideUserID == userID)
	reason := ""
	if !allowed {
		reason = "\u4ec5\u5c40\u521b\u5efa\u8005\u6216\u4e3b\u9886\u8def\u4eba\u53ef\u53d1\u8d77\u5f15\u8350"
	}
	httpx.OK(w, map[string]interface{}{"allowed": allowed, "reason": reason, "gameId": gameID})
}

func (s *Server) gameInvitePlayers(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	gameID := parseFlexibleInt64(r.URL.Query().Get("sourceGameId"))
	if gameID <= 0 {
		gameID = parseFlexibleInt64(r.URL.Query().Get("gameId"))
	}
	excludedUserID := parseFlexibleInt64(r.URL.Query().Get("excludeUserId"))
	if excludedUserID <= 0 {
		excludedUserID = parseFlexibleInt64(r.URL.Query().Get("expertUserId"))
	}
	items := s.inviteCandidates(userID, keyword, true, gameID, excludedUserID)
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
		"sourceGameId": game.ID,
		"sourceGame": map[string]interface{}{
			"id":          game.ID,
			"title":       game.Title,
			"description": game.Description,
			"price":       game.Price,
			"startAt":     game.StartAt,
			"endAt":       game.EndAt,
		},
		"title": "\u592a\u68d2\u4e86\uff01\u4f60\u60f3\u600e\u4e48\u5f00\u542f\u4e0b\u4e00\u5c40\uff1f",
		"desc":  "\u6839\u636e\u4e0a\u4e00\u5c40\u548c\u540e\u53f0\u914d\u7f6e\u8fd4\u56de\u53ef\u7528\u65b9\u5f0f",
		"inviter": map[string]interface{}{
			"id":        userID,
			"name":      s.inGameDisplayName(userID, "\u9080\u8bf7\u4eba"),
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
	if !s.userCanGenerateInvitations(userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅行家或领路人可生成再玩一局邀请")
		return
	}
	var req struct {
		SourceGameID     interface{} `json:"sourceGameId"`
		ExpertUserID     interface{} `json:"expertUserId"`
		PlayerUserIDs    []int64     `json:"playerUserIds"`
		ExpertUserIDs    []int64     `json:"expertUserIds"`
		GuideUserIDs     []int64     `json:"guideUserIds"`
		Message          string      `json:"message"`
		ServiceType      string      `json:"serviceType"`
		ServiceDuration  string      `json:"serviceDuration"`
		DemandDetail     string      `json:"demandDetail"`
		BudgetAmount     interface{} `json:"budgetAmount"`
		BudgetAmountCent interface{} `json:"budgetAmountCent"`
		ExpectedTime     string      `json:"expectedTime"`
		Invitees         []struct {
			ID       interface{} `json:"id"`
			UserID   interface{} `json:"userId"`
			Role     string      `json:"role"`
			RoleType string      `json:"roleType"`
		} `json:"invitees"`
		Overrides games.ReplayGameOverrides `json:"overrides"`
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
	sourceGame, err := s.games.Get(sourceGameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	scheduleText, sourceDurationText := invitationGameSchedule(sourceGame)
	if strings.TrimSpace(req.Overrides.StartAt) == "" {
		req.Overrides.StartAt = sourceGame.StartAt
	}
	if strings.TrimSpace(req.Overrides.EndAt) == "" {
		req.Overrides.EndAt = sourceGame.EndAt
	}
	appendUnique := func(values []int64, id int64) []int64 {
		if id <= 0 || id == userID {
			return values
		}
		for _, item := range values {
			if item == id {
				return values
			}
		}
		return append(values, id)
	}
	req.ExpertUserIDs = appendUnique(req.ExpertUserIDs, parseFlexibleInt64(req.ExpertUserID))
	for _, invitee := range req.Invitees {
		id := parseFlexibleInt64(invitee.UserID)
		if id <= 0 {
			id = parseFlexibleInt64(invitee.ID)
		}
		role := strings.ToLower(strings.TrimSpace(invitee.RoleType))
		if role == "" {
			role = strings.ToLower(strings.TrimSpace(invitee.Role))
		}
		if role == "expert" || role == "行家" {
			req.ExpertUserIDs = appendUnique(req.ExpertUserIDs, id)
		} else {
			req.PlayerUserIDs = appendUnique(req.PlayerUserIDs, id)
		}
	}
	// 续局是新建一局，上一局成员正是本流程的邀请对象，不能按旧局成员拦截。
	// 仅由 appendUnique 排除当前发起人，避免把自己作为受邀人重复加入。
	budgetAmountCent := parseFlexibleInt64(req.BudgetAmountCent)
	if budgetAmountCent <= 0 {
		budgetAmountCent = parseFlexibleInt64(req.BudgetAmount) * 100
	}
	if budgetAmountCent <= 0 && req.Overrides.Price != nil {
		budgetAmountCent = int64(*req.Overrides.Price*100 + 0.5)
	}
	serviceType := strings.TrimSpace(req.ServiceType)
	if serviceType == "" {
		serviceType = strings.TrimSpace(req.Overrides.Title)
	}
	serviceDuration := firstNonEmpty(strings.TrimSpace(req.ServiceDuration), sourceDurationText)
	expectedTime := firstNonEmpty(strings.TrimSpace(req.ExpectedTime), scheduleText)
	demandDetail := strings.TrimSpace(req.DemandDetail)
	if demandDetail == "" {
		demandDetail = strings.TrimSpace(req.Overrides.Description)
	}
	inviteGroupID := "replay-" + strconv.FormatInt(sourceGameID, 10) + "-" + strconv.FormatInt(userID, 10) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	result, err := s.games.CreateReplayGame(userID, sourceGameID, games.ReplayGameRequest{
		PlayerUserIDs:    req.PlayerUserIDs,
		ExpertUserIDs:    req.ExpertUserIDs,
		GuideUserIDs:     req.GuideUserIDs,
		Message:          req.Message,
		InviteGroupID:    inviteGroupID,
		ServiceType:      serviceType,
		ServiceDuration:  serviceDuration,
		DemandDetail:     demandDetail,
		BudgetAmountCent: budgetAmountCent,
		ExpectedTime:     expectedTime,
		Overrides:        req.Overrides,
	})
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "create_replay_game", "game", result.ReplayGameID, map[string]interface{}{
		"sourceGameId":    sourceGameID,
		"invitationCount": result.InvitationCount,
	})
	for _, invitation := range result.Invitations {
		content := strings.TrimSpace(invitation.Message)
		if content == "" {
			content = "你收到了一条组局引荐邀请"
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     invitation.TargetUserID,
			NotifyType: "game_invitation",
			Title:      "组局引荐",
			Content:    content,
			BizType:    "game_invitation",
			BizID:      invitation.ID,
			NeedWechat: true,
			WechatData: map[string]string{
				"thing1": content,
				"page":   "pages/game/guide-progress-detail/index?invitationId=" + strconv.FormatInt(invitation.ID, 10) + "&gameId=" + strconv.FormatInt(result.ReplayGameID, 10),
			},
		})
	}
	if result.PrimaryInvitationID > 0 {
		content := "已发起组局引荐，请关注玩家和行家的确认及审核进度"
		page := "pages/game/guide-progress-detail/index?invitationId=" + strconv.FormatInt(result.PrimaryInvitationID, 10) + "&gameId=" + strconv.FormatInt(result.ReplayGameID, 10)
		s.notices.Create(notifications.CreateRequest{
			UserID:     userID,
			NotifyType: "game_invitation",
			Title:      "组局动态",
			Content:    content,
			BizType:    "game_invitation",
			BizID:      result.PrimaryInvitationID,
			NeedWechat: true,
			WechatData: map[string]string{
				"thing1": content,
				"page":   page,
			},
		})
	}
	httpx.OK(w, map[string]interface{}{
		"sourceGameId":        result.SourceGameID,
		"replayGameId":        result.ReplayGameID,
		"primaryInvitationId": result.PrimaryInvitationID,
		"replayInvitationId":  result.PrimaryInvitationID,
		"invitationId":        result.PrimaryInvitationID,
		"inviteGroupId":       inviteGroupID,
		"invitationCount":     result.InvitationCount,
		"invitations":         result.Invitations,
		"playerUserIds":       result.PlayerUserIDs,
		"expertUserIds":       result.ExpertUserIDs,
		"guideUserIds":        result.GuideUserIDs,
		"replayGame":          result.Game,
	})
}

func (s *Server) createCurrentGameInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if !s.userCanGenerateInvitations(userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅行家或领路人可生成组局邀请")
		return
	}
	var req struct {
		SourceGameID     interface{} `json:"sourceGameId"`
		ExpertUserID     interface{} `json:"expertUserId"`
		PlayerUserIDs    []int64     `json:"playerUserIds"`
		ExpertUserIDs    []int64     `json:"expertUserIds"`
		Message          string      `json:"message"`
		ServiceType      string      `json:"serviceType"`
		ServiceDuration  string      `json:"serviceDuration"`
		DemandDetail     string      `json:"demandDetail"`
		BudgetAmountCent interface{} `json:"budgetAmountCent"`
		ExpectedTime     string      `json:"expectedTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	gameID := parseFlexibleInt64(req.SourceGameID)
	if gameID <= 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "game required")
		return
	}
	appendUnique := func(values []int64, id int64) []int64 {
		if id <= 0 || id == userID {
			return values
		}
		for _, item := range values {
			if item == id {
				return values
			}
		}
		return append(values, id)
	}
	expertUserIDs := make([]int64, 0, len(req.ExpertUserIDs)+1)
	for _, id := range req.ExpertUserIDs {
		expertUserIDs = appendUnique(expertUserIDs, id)
	}
	expertUserIDs = appendUnique(expertUserIDs, parseFlexibleInt64(req.ExpertUserID))
	playerUserIDs := make([]int64, 0, len(req.PlayerUserIDs))
	for _, id := range req.PlayerUserIDs {
		playerUserIDs = appendUnique(playerUserIDs, id)
	}
	if len(playerUserIDs) == 0 || len(expertUserIDs) == 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "player and expert required")
		return
	}

	inviteGroupID := "game-" + strconv.FormatInt(gameID, 10) + "-" + strconv.FormatInt(userID, 10) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	firstPlayerUserID := playerUserIDs[0]
	firstExpertUserID := expertUserIDs[0]
	invitations := make([]games.Invitation, 0, len(playerUserIDs)+len(expertUserIDs))
	createInvitation := func(targetUserID int64, role string, playerUserID int64, expertUserID int64) error {
		invitation, err := s.games.CreateInvitation(userID, gameID, games.InvitationRequest{
			TargetUserID:     targetUserID,
			PlayerUserID:     playerUserID,
			ExpertUserID:     expertUserID,
			InviteGroupID:    inviteGroupID,
			RoleType:         role,
			Message:          req.Message,
			ServiceType:      req.ServiceType,
			ServiceDuration:  req.ServiceDuration,
			DemandDetail:     req.DemandDetail,
			BudgetAmountCent: parseFlexibleInt64(req.BudgetAmountCent),
			ExpectedTime:     req.ExpectedTime,
		})
		if err != nil {
			return err
		}
		invitations = append(invitations, invitation)
		content := strings.TrimSpace(invitation.Message)
		if content == "" {
			content = "\u4f60\u6536\u5230\u4e86\u4e00\u6761\u7ec4\u5c40\u5f15\u8350\u9080\u8bf7"
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     invitation.TargetUserID,
			NotifyType: "game_invitation",
			Title:      "\u7ec4\u5c40\u5f15\u8350",
			Content:    content,
			BizType:    "game_invitation",
			BizID:      invitation.ID,
			NeedWechat: true,
			WechatData: map[string]string{
				"thing1": content,
				"page":   "pages/game/guide-progress-detail/index?invitationId=" + strconv.FormatInt(invitation.ID, 10) + "&gameId=" + strconv.FormatInt(gameID, 10),
			},
		})
		return nil
	}
	for _, playerUserID := range playerUserIDs {
		if err := createInvitation(playerUserID, "player", playerUserID, firstExpertUserID); err != nil {
			writeGameError(w, err)
			return
		}
	}
	for _, expertUserID := range expertUserIDs {
		if err := createInvitation(expertUserID, "expert", firstPlayerUserID, expertUserID); err != nil {
			writeGameError(w, err)
			return
		}
	}
	primaryInvitationID := invitations[0].ID
	s.recordBehavior(userID, "create_current_game_invitation", "game", gameID, map[string]interface{}{
		"invitationCount": len(invitations),
		"inviteGroupId":   inviteGroupID,
	})
	httpx.OK(w, map[string]interface{}{
		"gameId":              gameID,
		"primaryInvitationId": primaryInvitationID,
		"invitationId":        primaryInvitationID,
		"inviteGroupId":       inviteGroupID,
		"invitationCount":     len(invitations),
		"invitations":         invitations,
		"playerUserIds":       playerUserIDs,
		"expertUserIds":       expertUserIDs,
	})
}

func (s *Server) userCanGenerateInvitations(userID int64) bool {
	if userID <= 0 || s.profiles == nil {
		return false
	}
	roles := s.profiles.RoleSnapshot(userID).RoleStatusMap
	return roles["expert"] == "approved" || roles["expert"] == "active" || roles["guide"] == "approved" || roles["guide"] == "active"
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
	filterInvitationID := parseFlexibleInt64(r.URL.Query().Get("invitationId"))
	if filterInvitationID <= 0 {
		filterInvitationID = parseFlexibleInt64(r.URL.Query().Get("id"))
	}
	filterGameID := parseFlexibleInt64(r.URL.Query().Get("gameId"))
	if filterGameID <= 0 {
		filterGameID = parseFlexibleInt64(r.URL.Query().Get("sourceGameId"))
	}
	active := make([]map[string]interface{}, 0)
	completed := make([]map[string]interface{}, 0)
	matchedInvitation := false
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if filterInvitationID > 0 && invitation.ID != filterInvitationID {
			continue
		}
		if filterGameID > 0 && invitation.GameID != filterGameID {
			continue
		}
		matchedInvitation = true
		item := s.invitationGuideProgressItem(invitation)
		isCanceled, _ := item["isCanceled"].(bool)
		statusTitle, _ := item["statusTitle"].(string)
		if isCanceled || statusTitle == "组局成功" {
			completed = append(completed, item)
		} else {
			active = append(active, item)
		}
	}
	if filterInvitationID > 0 && !matchedInvitation {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "invitation not found")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"activeCount":      len(active),
		"activeParties":    active,
		"completedParties": completed,
		"pageTexts":        guideProgressPageTexts(),
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
	allRecords := records
	records, page, pageSize, total := paginateRoleItems(r, allRecords, "state")
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
		"tabs":        tabs,
		"records":     records,
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
		"hasPrevious": page > 1,
		"hasMore":     page*pageSize < total,
		"pageConfig":  pageConfig,
	})
}

func (s *Server) gameInviteCancelDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if notificationID := strings.TrimSpace(r.URL.Query().Get("notificationId")); notificationID != "" {
		notice, found := s.findNotificationForUser(userID, notificationID)
		if !found || !isServiceCancelNotification(notice.NotifyType) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "cancel notification not found")
			return
		}
		detail, found := s.serviceCancelNotificationDetail(notice)
		if !found {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "canceled game not found")
			return
		}
		httpx.OK(w, detail)
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
	cancelProfile := s.inGameIdentity(cancelUserID, "成员")
	httpx.OK(w, map[string]interface{}{
		"id":          invitation.ID,
		"statusTitle": "\u7ec4\u5c40\u5df2\u53d6\u6d88",
		"statusDesc":  "\u672c\u6b21\u7ec4\u5c40\u9080\u8bf7\u5df2\u53d6\u6d88",
		"canceledBy": map[string]interface{}{
			"id":          cancelUserID,
			"name":        cancelProfile.DisplayName,
			"realName":    cancelProfile.RealName,
			"displayName": cancelProfile.DisplayName,
			"avatarText":  cancelProfile.AvatarText,
			"roleType":    "member",
			"roleLabel":   "\u6210\u5458",
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

func isServiceCancelNotification(notifyType string) bool {
	switch strings.TrimSpace(notifyType) {
	case "player_cancel_request", "expert_cancel_request", "guide_cancel_request":
		return true
	default:
		return false
	}
}

func (s *Server) serviceCancelNotificationDetail(notice notifications.Notification) (map[string]interface{}, bool) {
	if notice.BizType != "game" || notice.BizID <= 0 {
		return nil, false
	}
	game, err := s.games.Get(notice.BizID)
	if err != nil {
		return nil, false
	}

	cancelUserID := int64(0)
	roleType := "member"
	roleLabel := "成员"
	switch notice.NotifyType {
	case "player_cancel_request":
		cancelUserID = s.firstMemberWithGameRole(game, notice.UserID, "member")
		roleType = "player"
		roleLabel = "玩家"
	case "expert_cancel_request":
		cancelUserID = game.CreatorUserID
		if cancelUserID == notice.UserID && game.MainGuideUserID != notice.UserID {
			cancelUserID = game.MainGuideUserID
		}
		roleType = "expert"
		roleLabel = "行家"
	case "guide_cancel_request":
		cancelUserID = game.MainGuideUserID
		roleType = "guide"
		roleLabel = "领路人"
	}

	cancelProfile := identity.InGameIdentity{
		RealName:    roleLabel,
		DisplayName: roleLabel,
		AvatarText:  roleLabel,
	}
	if cancelUserID > 0 {
		cancelProfile = s.inGameIdentity(cancelUserID, roleLabel)
	}
	reason := serviceCancelReasonFromContent(notice.Content)
	timeText := notice.CreatedAt.Format("01-02 15:04")
	return map[string]interface{}{
		"id":              notice.ID,
		"notificationId":  notice.ID,
		"gameId":          game.ID,
		"statusTitle":     "组局已取消",
		"statusDesc":      roleLabel + "已取消本次组局",
		"cancelRole":      roleType,
		"cancelRoleLabel": roleLabel,
		"canceledBy": map[string]interface{}{
			"id":          cancelUserID,
			"name":        cancelProfile.DisplayName,
			"realName":    cancelProfile.RealName,
			"displayName": cancelProfile.DisplayName,
			"avatarText":  cancelProfile.AvatarText,
			"roleType":    roleType,
			"roleLabel":   roleLabel,
		},
		"reason":          map[string]string{"title": roleLabel + "主动取消", "desc": reason},
		"message":         notice.Content,
		"messageTimeText": timeText,
		"timeline": []map[string]interface{}{
			{"key": "game", "title": "组局进行中", "desc": game.Title, "timeText": inviteTimeText(game), "state": "active"},
			{"key": "cancel", "title": roleLabel + "取消", "desc": notice.Content, "timeText": timeText, "state": "error"},
		},
	}, true
}

func serviceCancelReasonFromContent(content string) string {
	content = strings.TrimSpace(content)
	if index := strings.LastIndex(content, "原因："); index >= 0 {
		if reason := strings.TrimSpace(content[index+len("原因："):]); reason != "" {
			return reason
		}
	}
	if index := strings.LastIndex(content, ":"); index >= 0 {
		if reason := strings.TrimSpace(content[index+1:]); reason != "" {
			return reason
		}
	}
	if content != "" {
		return content
	}
	return "未填写取消原因"
}

func (s *Server) inviteCandidates(userID int64, keyword string, includeAllUsers bool, excludedGameID int64, excludedUserID int64) []inviteCandidateDTO {
	seen := map[int64]bool{userID: true}
	if excludedUserID > 0 {
		seen[excludedUserID] = true
	}
	inGame := map[int64]bool{}
	if excludedGameID > 0 {
		if game, err := s.games.Get(excludedGameID); err == nil {
			if game.CreatorUserID > 0 {
				inGame[game.CreatorUserID] = true
			}
			if game.MainGuideUserID > 0 {
				inGame[game.MainGuideUserID] = true
			}
		}
		for _, memberID := range s.games.Members(excludedGameID) {
			inGame[memberID] = true
		}
	}
	items := make([]inviteCandidateDTO, 0)
	add := func(candidateID int64, tag string, score int) {
		if candidateID <= 0 || seen[candidateID] {
			return
		}
		name := s.inGameDisplayName(candidateID, "\u73a9\u5bb6")
		if keyword != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) && !strings.Contains(strconv.FormatInt(candidateID, 10), keyword) {
			return
		}
		seen[candidateID] = true
		candidateInGame := inGame[candidateID]
		if candidateInGame {
			tag = "\u5df2\u5728\u5c40\u5185"
		}
		items = append(items, inviteCandidateDTO{
			ID:          candidateID,
			UserID:      candidateID,
			Name:        name,
			Nickname:    name,
			RealName:    name,
			DisplayName: name,
			AvatarText:  avatarTextForName(name, candidateID),
			Tag:         tag,
			Desc:        "\u6765\u81ea\u771f\u5b9e\u5173\u7cfb\u6216\u7ec4\u5c40\u8bb0\u5f55",
			Meta:        "\u5173\u7cfb\u5f3a\u5ea6 " + strconv.Itoa(score),
			RoleType:    "player",
			RoleLabel:   "\u73a9\u5bb6",
			InGame:      candidateInGame,
			CanSelect:   !candidateInGame,
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
		name := s.inGameDisplayName(profile.UserID, "\u884c\u5bb6")
		match := 70 + profile.Completeness/4
		if match > 98 {
			match = 98
		}
		items = append(items, inviteCandidateDTO{
			ID:          profile.UserID,
			UserID:      profile.UserID,
			Name:        name,
			Nickname:    name,
			RealName:    name,
			DisplayName: name,
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
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	for _, memberID := range s.games.Members(game.ID) {
		if memberID == userID {
			continue
		}
		name := s.inGameDisplayName(memberID, "\u6210\u5458")
		roleType := gameMemberRole(game, memberID, true, memberRoles)
		roleLabel := "\u73a9\u5bb6"
		switch roleType {
		case "expert":
			roleType = "expert"
			roleLabel = "\u884c\u5bb6"
		case "guide", "main_guide":
			roleType = "guide"
			roleLabel = "\u9886\u8def\u4eba"
		default:
			roleType = "player"
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
		for _, candidate := range s.inviteCandidates(userID, "", false, game.ID, 0) {
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

func (s *Server) invitationProgressItem(invitation games.Invitation, viewerID int64) map[string]interface{} {
	game, _ := s.games.Get(invitation.GameID)
	targetName := s.inGameDisplayName(invitation.TargetUserID, "\u73a9\u5bb6")
	inviterName := s.inGameDisplayName(invitation.InviterID, "\u9080\u8bf7\u4eba")
	targetRole, targetRoleLabel := auditApplicationRole(invitation.Role)
	isTargetViewer := viewerID == invitation.TargetUserID
	displayName := targetName
	displayRole := targetRoleLabel
	if isTargetViewer {
		displayName = inviterName
		displayRole = "\u9080\u8bf7\u4eba"
	}
	statusText := map[string]string{"pending": "\u5f85\u786e\u8ba4", "accepted": "\u5df2\u786e\u8ba4", "rejected": "\u5df2\u62d2\u7edd"}[invitation.Status]
	if statusText == "" {
		statusText = invitation.Status
	}
	pageTitle := "\u9080\u8bf7\u8fdb\u5ea6"
	applicantTitle := targetRoleLabel + "\u4fe1\u606f"
	confirmText := "\u786e\u8ba4\u52a0\u5165"
	referralText := "\u5df2\u53d1\u51fa\u7ec4\u5c40\u9080\u8bf7"
	if isTargetViewer {
		referralText = "\u9080\u8bf7\u4f60\u53c2\u4e0e\u7ec4\u5c40"
		switch targetRole {
		case "expert":
			pageTitle = "\u884c\u5bb6\u5ba1\u6838\u7ec4\u5c40"
			applicantTitle = "\u73a9\u5bb6\u4e0e\u9700\u6c42\u4fe1\u606f"
			confirmText = "\u786e\u8ba4\u901a\u8fc7"
		case "player":
			pageTitle = "\u73a9\u5bb6\u786e\u8ba4\u7ec4\u5c40"
			applicantTitle = "\u884c\u5bb6 / \u9080\u8bf7\u4eba\u4fe1\u606f"
			confirmText = "\u786e\u8ba4\u53c2\u52a0"
		case "guide":
			pageTitle = "\u9886\u8def\u4eba\u786e\u8ba4\u52a0\u5165"
			applicantTitle = "\u53d1\u8d77\u4eba\u4fe1\u606f"
		}
	}
	timeText, durationText := invitationGameSchedule(game)
	locationText := firstNonEmpty(strings.TrimSpace(game.Address), strings.TrimSpace(game.CityName))
	activityType := auditGameTypeText(game)
	budgetText := "\u514d\u8d39"
	if game.Price > 0 {
		budgetText = "\u00a5" + strconv.FormatFloat(game.Price, 'f', 2, 64)
	}
	confirmedCount := 1
	if invitation.Status == "accepted" {
		confirmedCount = 2
	}
	detailDisplay := map[string]interface{}{
		"scenario":           "invitation_" + targetRole,
		"pageTitle":          pageTitle,
		"referralText":       referralText,
		"applicantTitle":     applicantTitle,
		"confirmTitle":       "\u7ec4\u5c40\u4fe1\u606f",
		"confirmText":        confirmText,
		"showRelation":       true,
		"showPortfolio":      targetRole == "expert" && isTargetViewer,
		"showSession":        true,
		"showConfirm":        true,
		"showRecommend":      false,
		"showOptions":        false,
		"showNotice":         invitation.Status == "pending" && isTargetViewer,
		"showActionBar":      invitation.Status == "pending" && isTargetViewer,
		"reviewReadonlyText": statusText,
		"status":             map[string]interface{}{"title": statusText, "quote": invitation.Message, "guideName": inviterName, "countdown": statusText},
		"player": map[string]interface{}{
			"name": displayName, "realName": displayName, "displayName": displayName, "avatarText": avatarTextForName(displayName, invitation.InviterID), "desc": displayRole,
			"roleKey": targetRole, "roleName": displayRole, "tags": []string{displayRole},
			"needText": firstNonEmpty(invitation.Message, "\u672a\u586b\u5199\u9080\u8bf7\u8bf4\u660e"),
			"remark":   nonEmptyAuditPrefix("\u9080\u8bf7\u65f6\u95f4\uff1a", invitation.CreatedAt.Local().Format("2006-01-02 15:04")),
		},
		"relation": map[string]interface{}{
			"title": "\u7ec4\u5c40\u786e\u8ba4\u5173\u7cfb", "totalCount": 2, "confirmedCount": confirmedCount,
			"expert": map[string]interface{}{"avatarText": avatarTextForName(inviterName, invitation.InviterID), "name": inviterName, "roleText": "\u9080\u8bf7\u4eba"},
			"guide":  map[string]interface{}{"iconSrc": "/pages/game/guide-chat/assets/icon-invite.png", "name": "\u9080\u8bf7"},
			"player": map[string]interface{}{"avatarText": avatarTextForName(targetName, invitation.TargetUserID), "name": targetName, "confirmed": invitation.Status == "accepted", "statusText": statusText},
		},
		"gameInfo": map[string]interface{}{
			"topic": game.Title, "time": firstNonEmpty(timeText, "\u672a\u8bbe\u7f6e"), "location": firstNonEmpty(locationText, "\u672a\u8bbe\u7f6e"),
			"activityType": activityType, "serviceDuration": firstNonEmpty(durationText, "\u672a\u8bbe\u7f6e"), "clientBudget": budgetText,
		},
		"sessionInfo": []map[string]interface{}{
			{"key": "topic", "label": "\u7ec4\u5c40\u4e3b\u9898", "value": game.Title, "iconSrc": "/pages/game/detail/assets/icon-calendar.png", "iconClass": "topic"},
			{"key": "time", "label": "\u65f6\u95f4", "value": firstNonEmpty(timeText, "\u672a\u8bbe\u7f6e"), "iconSrc": "/pages/game/detail/assets/icon-clock.png", "iconClass": "time"},
			{"key": "location", "label": "\u5730\u70b9", "value": firstNonEmpty(locationText, "\u672a\u8bbe\u7f6e"), "actionText": "\u5730\u56fe\u4f4d\u7f6e", "iconSrc": "/pages/game/detail/assets/icon-location.png", "iconClass": "place"},
		},
		"confirmRows": []map[string]interface{}{
			{"label": "\u6d3b\u52a8\u7c7b\u578b", "value": activityType}, {"label": "\u670d\u52a1\u65f6\u957f", "value": firstNonEmpty(durationText, "\u672a\u8bbe\u7f6e")}, {"label": "\u5ba2\u6237\u9884\u7b97", "value": budgetText},
		},
		"noticeBullets": []string{"\u8bf7\u786e\u8ba4\u7ec4\u5c40\u4fe1\u606f\u540e\u518d\u505a\u51b3\u5b9a", "\u786e\u8ba4\u540e\u5c06\u6309\u5bf9\u5e94\u8eab\u4efd\u52a0\u5165\u672c\u5c40"},
	}
	return map[string]interface{}{
		"id":              invitation.ID,
		"invitationId":    invitation.ID,
		"status":          invitation.Status,
		"statusText":      statusText,
		"targetRole":      targetRole,
		"viewerRole":      map[bool]string{true: targetRole, false: "inviter"}[isTargetViewer],
		"detailDisplay":   detailDisplay,
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

func guideProgressPageTexts() map[string]string {
	return map[string]string{
		"pageTitle":             "组局消息",
		"loadingText":           "加载中...",
		"retryText":             "点击重试",
		"activeTitleTemplate":   "进行中的组局 {count}",
		"activeEmptyText":       "暂无进行中的组局",
		"completedTitle":        "最近完成",
		"completedEmptyText":    "暂无完成记录",
		"loadFailedText":        "组局消息加载失败",
		"progressDetailTitle":   "组局进度详情",
		"detailLoadingText":     "加载中...",
		"detailRetryText":       "点击重试",
		"detailEmptyText":       "暂无组局进度详情",
		"countdownLabel":        "确认剩余时间",
		"timelineTitle":         "进度追踪",
		"participantsTitle":     "参与双方",
		"cancelActionText":      "取消组局",
		"chatPageTitle":         "小程序通知",
		"chatCardTitle":         "收到组局邀请",
		"chatGuideLabel":        "领路人",
		"chatAssistantName":     "组局助手",
		"chatPlayerInfoLabel":   "玩家信息",
		"chatAcceptButtonText":  "确认参加",
		"chatDeclineButtonText": "婉拒",
		"chatEmptyText":         "暂无引荐上下文，请从组局进度页进入",
		"chatAcceptText":        "已确认参加",
		"chatDeclineText":       "已婉拒",
		"chatSendFailedText":    "发送失败",
		"remindFailedText":      "提醒记录失败，将进入聊天",
	}
}

func guideProgressInvitationRole(role string) string {
	if normalized, _ := auditApplicationRole(role); normalized == "expert" {
		return "expert"
	}
	return "player"
}

func guideProgressStatusText(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved":
		return "已审核通过"
	case "review_pending":
		return "待审核"
	case "review_rejected":
		return "审核未通过"
	case "accepted":
		return "已确认"
	case "rejected":
		return "已拒绝"
	default:
		return "待确认"
	}
}

func guideProgressStateClass(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved":
		return "confirmed"
	case "rejected", "review_rejected":
		return "canceled"
	default:
		return "waiting"
	}
}

func (s *Server) guideProgressInvitationForRole(base games.Invitation, role string) (games.Invitation, bool) {
	var selected games.Invitation
	for _, candidate := range s.games.InvitationsForUser(base.InviterID) {
		if candidate.GameID != base.GameID || candidate.InviterID != base.InviterID || !sameInvitationGroup(base, candidate) || guideProgressInvitationRole(candidate.Role) != role {
			continue
		}
		if candidate.ID == base.ID {
			return candidate, true
		}
		if selected.ID == 0 || candidate.CreatedAt.After(selected.CreatedAt) {
			selected = candidate
		}
	}
	return selected, selected.ID > 0
}

func sameInvitationGroup(base games.Invitation, candidate games.Invitation) bool {
	return strings.TrimSpace(base.InviteGroupID) == strings.TrimSpace(candidate.InviteGroupID)
}

func guideProgressFallbackPartyID(game games.Game, memberIDs []int64, inviterID int64, excludedID int64) int64 {
	if game.CreatorUserID > 0 && game.CreatorUserID != inviterID && game.CreatorUserID != excludedID {
		return game.CreatorUserID
	}
	for _, memberID := range memberIDs {
		if memberID > 0 && memberID != inviterID && memberID != excludedID {
			return memberID
		}
	}
	return 0
}

func (s *Server) guideProgressMember(userID int64, role string, status string, gameTitle string) map[string]interface{} {
	roleLabel := "玩家"
	avatarClass := "blue"
	if role == "expert" {
		roleLabel = "行家"
		avatarClass = "pink"
	}
	name := s.inGameDisplayName(userID, roleLabel)
	stateText := guideProgressStatusText(status)
	if strings.EqualFold(strings.TrimSpace(status), "approved") {
		stateText = "已确认"
	}
	stateClass := guideProgressStateClass(status)
	badgeIcon := ""
	if stateClass == "confirmed" {
		badgeIcon = guideProgressConfirmedIconURL
	} else if stateClass == "waiting" {
		badgeIcon = guideProgressWaitingIconURL
	}
	return map[string]interface{}{
		"id":          userID,
		"userId":      userID,
		"name":        name,
		"roleType":    role,
		"roleLabel":   roleLabel,
		"roleClass":   role,
		"desc":        gameTitle,
		"status":      status,
		"statusText":  stateText,
		"state":       stateText,
		"stateClass":  stateClass,
		"cardClass":   stateClass,
		"badgeClass":  stateClass,
		"avatarText":  avatarTextForName(name, userID),
		"avatarClass": avatarClass,
		"badgeIcon":   badgeIcon,
	}
}

func guideProgressStep(role string, name string, status string, targetRole string) map[string]interface{} {
	roleLabel := "玩家"
	if role == "expert" {
		roleLabel = "行家"
	}
	title := "等待" + roleLabel + "确认"
	desc := "等待" + name + "确认参加组局"
	state := "pending"
	actionText := ""
	if status == "approved" {
		title = roleLabel + "已审核通过"
		desc = name + "已正式加入组局"
		state = "confirmed"
	} else if status == "review_pending" {
		title = roleLabel + "已接受，待审核"
		desc = name + "已接受邀请，请完成入局审核"
		state = "active"
	} else if status == "rejected" || status == "review_rejected" {
		title = roleLabel + "已拒绝"
		if status == "review_rejected" {
			desc = name + "的入局申请审核未通过"
		} else {
			desc = name + "已拒绝邀请"
		}
		state = "pending"
	} else {
		if role == targetRole {
			state = "active"
		}
		actionText = "再次提醒"
	}
	return map[string]interface{}{
		"key":        role,
		"title":      title,
		"desc":       desc,
		"state":      state,
		"actionText": actionText,
		"actionKey":  "remind" + strings.ToUpper(role[:1]) + role[1:],
		"iconSrc":    guideProgressRemindIconURL,
		"hasLine":    true,
		"lineState":  map[bool]string{true: "confirmed", false: "pending"}[status == "approved"],
	}
}

func invitationViewerRole(invitation games.Invitation, viewerID int64) string {
	if viewerID == invitation.InviterID {
		return "guide"
	}
	if viewerID == invitation.TargetUserID {
		return guideProgressInvitationRole(invitation.Role)
	}
	return ""
}

func (s *Server) gameInvitationConfirmRoute(invitation games.Invitation, viewerID int64) (string, string) {
	routePath, routeKey := "pages/game/audit-detail/index", "gameAuditDetail"
	switch invitationViewerRole(invitation, viewerID) {
	case "guide":
		routePath, routeKey = "pages/game/guide-progress-detail/index", "gameGuideProgressDetail"
	case "player":
		routePath, routeKey = "pages/game/player-confirm/index", "gamePlayerConfirm"
	}
	route := routePath + "?invitationId=" + strconv.FormatInt(invitation.ID, 10)
	if invitation.GameID > 0 {
		route += "&gameId=" + strconv.FormatInt(invitation.GameID, 10)
	}
	return route, routeKey
}

func gameInvitationSuccessRoute(gameID int64, role string) (string, string) {
	gameIDText := strconv.FormatInt(gameID, 10)
	switch strings.TrimSpace(role) {
	case "expert":
		return "pages/game/success-expert/index?gameId=" + gameIDText, "gameSuccessExpert"
	case "guide", "main_guide":
		return "pages/game/success-guide/index?gameId=" + gameIDText, "gameSuccessGuide"
	default:
		return "", ""
	}
}

func (s *Server) invitationConfirmedParties(base games.Invitation) (games.Invitation, games.Invitation, bool) {
	var playerInvitation games.Invitation
	var expertInvitation games.Invitation
	for _, candidate := range s.games.InvitationsForUser(base.InviterID) {
		if candidate.GameID != base.GameID || candidate.InviterID != base.InviterID || !sameInvitationGroup(base, candidate) {
			continue
		}
		switch guideProgressInvitationRole(candidate.Role) {
		case "player":
			if playerInvitation.ID == 0 || candidate.CreatedAt.After(playerInvitation.CreatedAt) {
				playerInvitation = candidate
			}
		case "expert":
			if expertInvitation.ID == 0 || candidate.CreatedAt.After(expertInvitation.CreatedAt) {
				expertInvitation = candidate
			}
		}
	}
	return playerInvitation, expertInvitation,
		playerInvitation.ID > 0 && expertInvitation.ID > 0 &&
			s.invitationApplicationStatus(playerInvitation) == "approved" &&
			s.invitationApplicationStatus(expertInvitation) == "approved"
}

func (s *Server) invitationApplicationStatus(invitation games.Invitation) string {
	if invitation.ApplicationID <= 0 || invitation.InviterID <= 0 {
		return ""
	}
	for _, application := range s.games.ApplicationsForUser(invitation.TargetUserID) {
		if application.ID == invitation.ApplicationID {
			return strings.ToLower(strings.TrimSpace(application.Status))
		}
	}
	for _, application := range s.games.ApplicationsForCreator(invitation.InviterID) {
		if application.ID == invitation.ApplicationID {
			return strings.ToLower(strings.TrimSpace(application.Status))
		}
	}
	return ""
}

func (s *Server) invitationReviewStatus(invitation games.Invitation) string {
	status := strings.ToLower(strings.TrimSpace(invitation.Status))
	if status != "accepted" {
		return status
	}
	switch s.invitationApplicationStatus(invitation) {
	case "approved":
		return "approved"
	case "rejected", "canceled", "cancelled":
		return "review_rejected"
	default:
		return "review_pending"
	}
}

func (s *Server) createInvitationReviewNotification(invitation games.Invitation, application games.Application) {
	if invitation.InviterID <= 0 || application.ID <= 0 {
		return
	}
	roleLabel := "玩家"
	if guideProgressInvitationRole(invitation.Role) == "expert" {
		roleLabel = "行家"
	}
	name := s.inGameDisplayName(invitation.TargetUserID, roleLabel)
	s.notices.Create(notifications.CreateRequest{
		UserID:     invitation.InviterID,
		NotifyType: "game_application",
		Title:      "邀请已接受，待你审核",
		Content:    name + "已接受" + roleLabel + "邀请，请完成入局审核。",
		BizType:    "game_application",
		BizID:      application.ID,
		NeedWechat: true,
	})
}

func (s *Server) createGameInvitationSuccessNotificationsForApplication(application games.Application) {
	if application.ID <= 0 || application.UserID <= 0 {
		return
	}
	for _, invitation := range s.games.InvitationsForUser(application.UserID) {
		if invitation.ApplicationID == application.ID {
			s.createGameInvitationSuccessNotifications(invitation)
			return
		}
	}
}

func (s *Server) createPairedInvitationProgressNotification(invitation games.Invitation) {
	if strings.TrimSpace(invitation.InviteGroupID) == "" || invitation.GameID <= 0 || invitation.TargetUserID <= 0 {
		return
	}
	role := guideProgressInvitationRole(invitation.Role)
	if role == "player" {
		s.createExpertConfirmationReadyNotification(invitation)
	}
	game, err := s.games.Get(invitation.GameID)
	if err != nil || game.CreatorUserID <= 0 || game.CreatorUserID == invitation.InviterID {
		return
	}
	roleLabel := "玩家"
	title := "领路人邀请的玩家待审核"
	if role == "expert" {
		roleLabel = "行家"
		title = "领路人邀请的行家待审核"
	}
	for _, item := range s.notices.List(game.CreatorUserID) {
		if item.NotifyType == "game_invitation_progress" && item.BizType == "game" && item.BizID == game.ID && item.Title == title {
			return
		}
	}
	targetName := s.inGameDisplayName(invitation.TargetUserID, roleLabel)
	content := targetName + "已接受领路人的邀请，等待发起人或行家最终审核。"
	s.notices.Create(notifications.CreateRequest{
		UserID:     game.CreatorUserID,
		NotifyType: "game_invitation_progress",
		Title:      title,
		Content:    content,
		BizType:    "game",
		BizID:      game.ID,
		NeedWechat: true,
		WechatData: map[string]string{
			"thing1": title,
			"page":   "pages/game/detail/index?id=" + strconv.FormatInt(game.ID, 10),
		},
	})
}

func (s *Server) createExpertConfirmationReadyNotification(playerInvitation games.Invitation) {
	expertInvitation, ok := s.guideProgressInvitationForRole(playerInvitation, "expert")
	if !ok || expertInvitation.ID <= 0 || expertInvitation.TargetUserID <= 0 || expertInvitation.Status != "pending" {
		return
	}
	const title = "玩家已确认，等待你的确认"
	for _, item := range s.notices.List(expertInvitation.TargetUserID) {
		if item.NotifyType == "game_invitation_progress" && item.BizType == "game_invitation" && item.BizID == expertInvitation.ID && item.Title == title {
			return
		}
	}
	playerName := s.inGameDisplayName(playerInvitation.TargetUserID, "玩家")
	page := "pages/game/audit-detail/index?invitationId=" + strconv.FormatInt(expertInvitation.ID, 10) + "&gameId=" + strconv.FormatInt(expertInvitation.GameID, 10)
	s.notices.Create(notifications.CreateRequest{
		UserID:     expertInvitation.TargetUserID,
		NotifyType: "game_invitation_progress",
		Title:      title,
		Content:    playerName + "已确认参加本次组局，请你完成最终确认。",
		BizType:    "game_invitation",
		BizID:      expertInvitation.ID,
		NeedWechat: true,
		WechatData: map[string]string{"thing1": title, "page": page},
	})
}

func (s *Server) hasInvitationSuccessNotification(userID int64, invitationID int64) bool {
	for _, item := range s.notices.List(userID) {
		if item.NotifyType == "game_invitation_success" && item.BizType == "game_invitation" && item.BizID == invitationID {
			return true
		}
	}
	return false
}

func (s *Server) createGameReadyToStartNotification(game games.Game) {
	if game.ID <= 0 || game.CreatorUserID <= 0 || game.MaxPlayers <= 0 ||
		game.Status != "full" || game.CurrentPlayers < game.MaxPlayers ||
		!s.gameInvitationGroupSucceeded(game) {
		return
	}
	for _, item := range s.notices.List(game.CreatorUserID) {
		if item.NotifyType == "game_ready_to_start" && item.BizType == "game" && item.BizID == game.ID {
			return
		}
	}
	title := "组局成功，可以开局"
	content := "玩家和行家均已确认，本局已组局成功，现在可以开局。"
	if gameTitle := strings.TrimSpace(game.Title); gameTitle != "" {
		content = "《" + gameTitle + "》玩家和行家均已确认，现在可以开局。"
	}
	s.notices.Create(notifications.CreateRequest{
		UserID:     game.CreatorUserID,
		NotifyType: "game_ready_to_start",
		Title:      title,
		Content:    content,
		BizType:    "game",
		BizID:      game.ID,
		NeedWechat: true,
		WechatData: map[string]string{
			"thing1": title,
			"page":   "pages/game/detail/index?id=" + strconv.FormatInt(game.ID, 10),
		},
	})
}

func (s *Server) gameInvitationGroupSucceeded(game games.Game) bool {
	for _, invitation := range s.games.InvitationsForUser(game.CreatorUserID) {
		if invitation.GameID != game.ID {
			continue
		}
		if _, _, allConfirmed := s.invitationConfirmedParties(invitation); allConfirmed {
			return true
		}
	}
	return false
}

func (s *Server) createGameInvitationSuccessNotifications(invitation games.Invitation) {
	_, expertInvitation, allConfirmed := s.invitationConfirmedParties(invitation)
	if !allConfirmed {
		return
	}
	game, _ := s.games.Get(invitation.GameID)
	s.createGameReadyToStartNotification(game)
	content := "组局成功，玩家和行家申请均已审核通过，接下来可进入组局详情安排后续事项。"
	if strings.TrimSpace(game.Title) != "" {
		content = "「" + strings.TrimSpace(game.Title) + "」" + content
	}
	recipients := []struct {
		userID     int64
		role       string
		invitation games.Invitation
	}{
		{userID: invitation.InviterID, role: "guide", invitation: invitation},
		{userID: expertInvitation.TargetUserID, role: "expert", invitation: expertInvitation},
	}
	for _, recipient := range recipients {
		if recipient.userID <= 0 || recipient.invitation.ID <= 0 || s.hasInvitationSuccessNotification(recipient.userID, recipient.invitation.ID) {
			continue
		}
		page, _ := gameInvitationSuccessRoute(invitation.GameID, recipient.role)
		s.notices.Create(notifications.CreateRequest{
			UserID:     recipient.userID,
			NotifyType: "game_invitation_success",
			Title:      "组局成功",
			Content:    content,
			BizType:    "game_invitation",
			BizID:      recipient.invitation.ID,
			NeedWechat: true,
			WechatData: map[string]string{"thing1": "组局成功", "page": page},
		})
	}
}

type guideProgressCountdownDTO struct {
	TimeoutSeconds   int64
	TimeoutAt        string
	RemainingSeconds int64
	Text             string
	Percent          int
}

func (s *Server) invitationTimeoutDuration() time.Duration {
	minutes := s.currentGameInviteConfig().InvitationTimeoutMinutes
	if minutes <= 0 {
		minutes = defaultGameInviteConfig().InvitationTimeoutMinutes
	}
	return time.Duration(minutes) * time.Minute
}

func (s *Server) guideProgressCountdown(invitation games.Invitation, statusText string) guideProgressCountdownDTO {
	timeout := s.invitationTimeoutDuration()
	totalSeconds := int64(timeout / time.Second)
	if totalSeconds <= 0 {
		totalSeconds = int64((24 * time.Hour) / time.Second)
		timeout = 24 * time.Hour
	}
	timeoutAt := invitation.CreatedAt.Add(timeout)
	remainingSeconds := int64(0)
	if invitation.Status == "pending" {
		remainingDuration := time.Until(timeoutAt)
		if remainingDuration > 0 {
			remainingSeconds = int64(remainingDuration / time.Second)
			if remainingDuration%time.Second != 0 {
				remainingSeconds++
			}
		}
	}
	percent := 0
	if totalSeconds > 0 {
		percent = int((remainingSeconds*100 + totalSeconds/2) / totalSeconds)
	}
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}
	text := statusText
	if invitation.Status == "pending" {
		text = guideProgressCountdownText(remainingSeconds)
	}
	return guideProgressCountdownDTO{
		TimeoutSeconds:   totalSeconds,
		TimeoutAt:        timeoutAt.Format(time.RFC3339),
		RemainingSeconds: remainingSeconds,
		Text:             text,
		Percent:          percent,
	}
}

func guideProgressCountdownText(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainSeconds := seconds % 60
	return guideProgressTwoDigit(hours) + ":" + guideProgressTwoDigit(minutes) + ":" + guideProgressTwoDigit(remainSeconds)
}

func guideProgressTwoDigit(value int64) string {
	text := strconv.FormatInt(value, 10)
	if value >= 0 && value < 10 {
		return "0" + text
	}
	return text
}

func guideProgressBudgetText(invitation games.Invitation, game games.Game) string {
	if invitation.BudgetAmountCent > 0 {
		return "¥" + strconv.FormatFloat(float64(invitation.BudgetAmountCent)/100, 'f', 2, 64)
	}
	if game.Price > 0 {
		return "¥" + strconv.FormatFloat(game.Price, 'f', 2, 64)
	}
	return "免费"
}

func guideProgressInfoRows(game games.Game, serviceType string, durationText string, budgetText string, expectedTimeText string, locationText string) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, 6)
	appendRow := func(label string, value string, highlight bool) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		row := map[string]interface{}{"label": label, "value": value}
		if highlight {
			row["highlight"] = true
		}
		rows = append(rows, row)
	}
	appendRow("主题", game.Title, false)
	appendRow("服务类型", serviceType, false)
	appendRow("咨询时长", durationText, false)
	appendRow("预算金额", budgetText, true)
	appendRow("预计时间", expectedTimeText, false)
	appendRow("地点", locationText, false)
	return rows
}

func (s *Server) invitationGuideProgressItem(invitation games.Invitation) map[string]interface{} {
	game, _ := s.games.Get(invitation.GameID)
	legacyTargetView := s.invitationProgressItem(invitation, invitation.TargetUserID)
	detailDisplay, _ := legacyTargetView["detailDisplay"].(map[string]interface{})
	targetRole := guideProgressInvitationRole(invitation.Role)
	playerInvitation, hasPlayerInvitation := s.guideProgressInvitationForRole(invitation, "player")
	expertInvitation, hasExpertInvitation := s.guideProgressInvitationForRole(invitation, "expert")
	playerID, playerStatus := int64(0), "pending"
	expertID, expertStatus := int64(0), "pending"
	if hasPlayerInvitation {
		playerID, playerStatus = playerInvitation.TargetUserID, s.invitationReviewStatus(playerInvitation)
	}
	if hasExpertInvitation {
		expertID, expertStatus = expertInvitation.TargetUserID, s.invitationReviewStatus(expertInvitation)
	}
	if targetRole == "player" {
		playerID, playerStatus = invitation.TargetUserID, s.invitationReviewStatus(invitation)
	} else {
		expertID, expertStatus = invitation.TargetUserID, s.invitationReviewStatus(invitation)
	}
	if playerID <= 0 {
		playerID = guideProgressFallbackPartyID(game, s.games.Members(game.ID), invitation.InviterID, expertID)
	}
	if expertID <= 0 {
		expertID = guideProgressFallbackPartyID(game, s.games.Members(game.ID), invitation.InviterID, playerID)
	}
	player := s.guideProgressMember(playerID, "player", playerStatus, game.Title)
	expert := s.guideProgressMember(expertID, "expert", expertStatus, game.Title)
	playerName, _ := player["name"].(string)
	expertName, _ := expert["name"].(string)
	allConfirmed := playerStatus == "approved" && expertStatus == "approved"
	isRejected := playerStatus == "rejected" || playerStatus == "review_rejected" || expertStatus == "rejected" || expertStatus == "review_rejected"
	statusText := "待确认"
	if playerStatus == "review_pending" || expertStatus == "review_pending" || playerStatus == "approved" || expertStatus == "approved" {
		statusText = "待审核"
	}
	title := "组局" + statusText
	statusTitle := statusText
	if allConfirmed {
		statusText = "组局成功"
		title, statusTitle = "组局成功", "组局成功"
	} else if isRejected {
		statusText = "组局已取消"
		title, statusTitle = "组局已取消", "组局已取消"
	}
	primaryActionText := "提醒行家"
	primaryActionRoute := ""
	remindTarget := targetRole
	if targetRole == "player" {
		primaryActionText = "提醒玩家"
	}
	if playerStatus == "review_pending" || expertStatus == "review_pending" {
		primaryActionText = "前往审核"
		primaryActionRoute = "/pages/game/audit/index?gameId=" + strconv.FormatInt(invitation.GameID, 10)
		remindTarget = ""
	} else if allConfirmed {
		primaryActionText = "查看组局"
		primaryActionRoute, _ = gameInvitationSuccessRoute(invitation.GameID, "guide")
		remindTarget = ""
	} else if isRejected {
		primaryActionText = ""
		remindTarget = ""
	}
	detailRoute := "/pages/game/guide-progress-detail/index?invitationId=" + strconv.FormatInt(invitation.ID, 10) + "&gameId=" + strconv.FormatInt(invitation.GameID, 10)
	cancelRoute := "/pages/game/guide-cancel/index?invitationId=" + strconv.FormatInt(invitation.ID, 10) + "&gameId=" + strconv.FormatInt(invitation.GameID, 10)
	steps := []map[string]interface{}{
		{"key": "launch", "title": "发起引荐", "desc": "已向" + playerName + "和" + expertName + "发送组局邀请", "timeText": invitation.CreatedAt.Format("01-02 15:04"), "state": "confirmed", "hasLine": true, "lineState": "confirmed"},
		guideProgressStep("player", playerName, playerStatus, targetRole),
		guideProgressStep("expert", expertName, expertStatus, targetRole),
		{"key": "success", "title": "组局成功", "desc": "玩家确认、行家审核通过后自动成局", "state": map[bool]string{true: "confirmed", false: "pending"}[allConfirmed], "hasLine": false, "lineState": "pending"},
	}
	progressPercent := 50
	if playerStatus == "review_pending" || expertStatus == "review_pending" || playerStatus == "approved" || expertStatus == "approved" {
		progressPercent = 66
	}
	if allConfirmed || isRejected {
		progressPercent = 100
	}
	rejectName, rejectRole := "", ""
	if isRejected {
		if targetRole == "expert" {
			rejectName, rejectRole = expertName, "行家"
		} else {
			rejectName, rejectRole = playerName, "玩家"
		}
	}
	scheduleText, gameDurationText := invitationGameSchedule(game)
	serviceType := firstNonEmpty(strings.TrimSpace(invitation.ServiceType), auditGameTypeText(game))
	durationText := firstNonEmpty(strings.TrimSpace(invitation.ServiceDuration), gameDurationText)
	budgetText := guideProgressBudgetText(invitation, game)
	expectedTimeText := firstNonEmpty(strings.TrimSpace(invitation.ExpectedTime), scheduleText)
	locationText := firstNonEmpty(strings.TrimSpace(game.Address), strings.TrimSpace(game.CityName))
	gameTimeText := firstNonEmpty(inviteTimeText(game), scheduleText)
	gameInfo := map[string]interface{}{
		"id": game.ID, "title": game.Title, "topic": game.Title,
		"time": gameTimeText, "timeText": gameTimeText,
		"location": locationText, "locationText": locationText,
		"activityType": serviceType, "serviceType": serviceType,
		"serviceDuration": durationText, "durationText": durationText,
		"clientBudget": budgetText, "budgetAmountText": budgetText,
		"expectedTime": expectedTimeText, "expectedTimeText": expectedTimeText,
		"demandDetail": strings.TrimSpace(invitation.DemandDetail),
	}
	infoRows := guideProgressInfoRows(game, serviceType, durationText, budgetText, expectedTimeText, locationText)
	countdown := s.guideProgressCountdown(invitation, statusText)
	countdownText := countdown.Text
	canConfirm := invitation.Status == "pending"
	confirmDisabledReason := ""
	if !canConfirm {
		confirmDisabledReason = "该邀请已处理"
	} else if targetRole == "expert" && playerStatus != "approved" {
		canConfirm = false
		confirmDisabledReason = "请等待玩家先确认组局"
	}
	if detailDisplay != nil {
		detailDisplay["canConfirm"] = canConfirm
		detailDisplay["confirmDisabledReason"] = confirmDisabledReason
		detailDisplay["declineText"] = "婉拒"
		detailDisplay["showActionBar"] = invitation.Status == "pending"
		detailDisplay["actionTip"] = map[bool]string{true: "确认后将通知行家进行最终审核", false: "确认后将建立三方连接并完成组局"}[targetRole == "player"]
		detailDisplay["optionTitle"] = "可选操作"
		detailDisplay["noticeTitle"] = "确认须知"
		detailDisplay["showOptions"] = true
		detailDisplay["optionalActions"] = []map[string]interface{}{
			{"key": "chat", "name": map[bool]string{true: "与行家沟通", false: "与玩家沟通"}[targetRole == "player"], "iconSrc": "/pages/game/audit-detail/assets/option-chat.svg"},
			{"key": "time", "name": "提议具体时间", "iconSrc": "/pages/game/audit-detail/assets/option-time.svg"},
		}
		displayProfile := player
		if targetRole == "player" {
			displayProfile = expert
			detailDisplay["applicantTitle"] = "行家信息"
			detailDisplay["confirmText"] = "确认参加"
			detailDisplay["showPortfolio"] = false
		} else {
			detailDisplay["applicantTitle"] = "玩家信息"
			detailDisplay["confirmText"] = "确认通过"
			detailDisplay["showPortfolio"] = true
		}
		displayProfile["compactCard"] = true
		displayProfile["compactProfile"] = true
		displayProfile["hideTags"] = false
		displayProfile["profileSections"] = []map[string]interface{}{
			{"title": "组局需求", "text": firstNonEmpty(strings.TrimSpace(invitation.DemandDetail), strings.TrimSpace(invitation.Message)), "tone": "need"},
			{"title": "预计时间", "text": expectedTimeText, "tone": "time", "iconSrc": "/pages/game/audit-detail/assets/player-time.svg"},
		}
		player["role"] = player["roleLabel"]
		player["confirmed"] = playerStatus == "approved"
		player["requirementConfirmed"] = playerStatus == "approved"
		player["requirementStatusText"] = player["statusText"]
		expert["role"] = expert["roleLabel"]
		expert["confirmed"] = expertStatus == "approved"
		expert["requirementConfirmed"] = expertStatus == "approved"
		expert["requirementStatusText"] = expert["statusText"]
		detailDisplay["player"] = displayProfile
		detailDisplay["relation"] = map[string]interface{}{
			"variant": "parties", "title": "成局双方",
			"expert": expert, "player": player,
			"confirmedText": map[bool]string{true: "双方已确认", false: "等待双方确认"}[allConfirmed],
		}
		if status, ok := detailDisplay["status"].(map[string]interface{}); ok {
			status["countdown"] = countdownText
		}
	}
	successRoute := ""
	if allConfirmed {
		successRoute, _ = gameInvitationSuccessRoute(invitation.GameID, "guide")
	}
	return map[string]interface{}{
		"id": invitation.ID, "invitationId": invitation.ID, "gameId": invitation.GameID,
		"route": detailRoute, "detailRoute": detailRoute, "cancelRoute": cancelRoute, "successRoute": successRoute,
		"status": invitation.Status, "statusText": statusText, "statusTitle": statusTitle,
		"targetRole": targetRole, "remindTarget": remindTarget, "primaryActionText": primaryActionText, "primaryActionRoute": primaryActionRoute,
		"canConfirm": canConfirm, "confirmDisabledReason": confirmDisabledReason,
		"timeoutSeconds": countdown.TimeoutSeconds, "timeLimitSeconds": countdown.TimeoutSeconds, "timeoutAt": countdown.TimeoutAt, "expiresAt": countdown.TimeoutAt,
		"remainingSeconds": countdown.RemainingSeconds, "countdownText": countdownText, "remainingText": countdownText,
		"countdownProgressPercent": countdown.Percent,
		"timeText":                 invitation.CreatedAt.Format("01-02 15:04"), "time": invitation.CreatedAt.Format("01-02 15:04"),
		"gameTitle": game.Title, "gameInfo": gameInfo, "game": gameInfo,
		"detailDisplay": detailDisplay,
		"infoRows":      infoRows,
		"players":       []map[string]interface{}{player}, "experts": []map[string]interface{}{expert}, "player": player, "expert": expert,
		"inviter":         map[string]interface{}{"id": invitation.InviterID, "userId": invitation.InviterID, "name": s.inGameDisplayName(invitation.InviterID, "领路人"), "roleType": "guide", "roleLabel": "领路人"},
		"progressPercent": progressPercent, "progressText": statusText, "theme": map[bool]string{true: "blue", false: "orange"}[allConfirmed], "tag": statusText,
		"steps": steps, "title": title, "memberText": playerName + " 与 " + expertName,
		"reason": invitation.Message, "icon": map[bool]string{true: "/pages/game/guide-progress/assets/result-canceled.png", false: "/pages/game/guide-progress/assets/result-success.png"}[isRejected],
		"isCanceled": isRejected, "muted": isRejected, "summaryPrefix": "", "summarySuffix": statusText,
		"rejectName": rejectName, "rejectRole": rejectRole, "rejectText": map[bool]string{true: "已拒绝邀请", false: ""}[isRejected],
		"chatPageTitle": "小程序通知", "chatCardTitle": "收到组局邀请", "chatGuideLabel": "领路人", "chatAssistantName": "组局助手",
		"chatPlayerInfoLabel":   map[bool]string{true: "行家信息", false: "玩家信息"}[targetRole == "player"],
		"chatAcceptButtonText":  map[bool]string{true: "确认参加", false: "确认通过"}[targetRole == "player"],
		"chatDeclineButtonText": "婉拒", "chatMessageText": invitation.Message,
		"chatDemandActionText": map[bool]string{true: "确认组局", false: "审核组局"}[targetRole == "player"],
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
	if game.Status == "pending_review" || game.Status == "completed" {
		state = "completed"
		stateText = "\u5df2\u5b8c\u6210"
		stateClass = "green"
		rewardClass = "green"
		actions = []map[string]interface{}{{"key": "review", "text": texts["reviewActionText"], "theme": "highlight"}}
	}
	if invitation.Status == "rejected" || game.Status == "canceled" || game.Status == "cancelled" {
		state = "canceled"
		stateText = "\u5df2\u53d6\u6d88"
		stateClass = "gray"
		rewardClass = "gray"
		actions = nil
	}
	targetName := s.inGameDisplayName(invitation.TargetUserID, "\u73a9\u5bb6")
	inviterName := s.inGameDisplayName(invitation.InviterID, "\u884c\u5bb6")
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

func (s *Server) inGameIdentity(userID int64, fallback string) identity.InGameIdentity {
	if value, ok := s.identity.InGameIdentity(userID); ok {
		return value
	}
	if fallback == "" {
		fallback = "成员"
	}
	name := s.displayName(userID, fallback)
	return identity.InGameIdentity{DisplayName: name, AvatarText: avatarTextForName(name, userID)}
}

func (s *Server) inGameDisplayName(userID int64, fallback string) string {
	return s.inGameIdentity(userID, fallback).DisplayName
}

func (s *Server) inGameAvatarText(userID int64, fallback string) string {
	return s.inGameIdentity(userID, fallback).AvatarText
}

func realnameAvatarType(hasRealName bool) string {
	if hasRealName {
		return "realname_initials"
	}
	return ""
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
	if initials := identity.RealNameInitials(name); initials != "" {
		return initials
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
