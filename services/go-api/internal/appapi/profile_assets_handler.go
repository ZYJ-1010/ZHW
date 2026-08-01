package appapi

import (
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/redemption"
)

const profileAssetManageConfigKey = "profile.asset_manage_config"

type profileAssetManageConfigDTO struct {
	OverviewLabel       string                        `json:"overviewLabel"`
	OverviewUpdatedText string                        `json:"overviewUpdatedText"`
	AssetStats          []profileAssetStatConfigDTO   `json:"assetStats"`
	QuickActions        []profileAssetActionConfigDTO `json:"quickActions"`
	MenuItems           []profileAssetMenuConfigDTO   `json:"menuItems"`
	OrderStatuses       []profileAssetStatusConfigDTO `json:"orderStatuses"`
	BankCards           profileAssetBankCardConfigDTO `json:"bankCards"`
	FAQLinks            []profileAssetFAQConfigDTO    `json:"faqLinks"`
	Version             string                        `json:"version"`
}

type profileAssetStatConfigDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Tone  string `json:"tone,omitempty"`
}

type profileAssetActionConfigDTO struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Tone           string `json:"tone,omitempty"`
	IconSrc        string `json:"iconSrc,omitempty"`
	Enabled        bool   `json:"enabled,omitempty"`
	DisabledReason string `json:"disabledReason,omitempty"`
}

type profileAssetMenuConfigDTO struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Value   string `json:"value,omitempty"`
	IconSrc string `json:"iconSrc,omitempty"`
	Tone    string `json:"tone,omitempty"`
	Route   string `json:"route,omitempty"`
}

type profileAssetStatusConfigDTO struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	IconSrc string `json:"iconSrc,omitempty"`
	Tone    string `json:"tone,omitempty"`
	Route   string `json:"route,omitempty"`
}

type profileAssetBankCardConfigDTO struct {
	UnboundText string `json:"unboundText"`
	BoundSuffix string `json:"boundSuffix"`
	CanBind     bool   `json:"canBind"`
}

type profileAssetFAQConfigDTO struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Answer string `json:"answer"`
}

func (s *Server) profileAssets(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}

	todos, err := s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}

	orders, err := s.redemption.OrdersForUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取兑换订单失败，请稍后重试")
		return
	}
	pointSummary, pointLogs, loaded := s.loadPointsData(w, userID, true)
	if !loaded {
		return
	}
	httpx.OK(w, s.buildProfileAssetsPayload(userID, orders, len(todos), pointSummary, pointLogs))
}

func (s *Server) buildProfileAssetsPayload(userID int64, orders []redemption.Order, reviewTodoCount int, pointSummary points.Account, pointLogs []points.Log) map[string]interface{} {
	config := s.currentProfileAssetManageConfig()

	return map[string]interface{}{
		"overview": map[string]interface{}{
			"label":       config.OverviewLabel,
			"value":       strconv.Itoa(pointSummary.AvailablePoints) + " 积分",
			"points":      pointSummary.AvailablePoints,
			"updatedText": config.OverviewUpdatedText,
		},
		"assetStats":     profileAssetStats(config.AssetStats, pointSummary, len(orders), reviewTodoCount),
		"quickActions":   profileAssetQuickActions(config.QuickActions),
		"menuItems":      profileAssetMenuItems(config.MenuItems, config.BankCards, 0),
		"orderStatuses":  profileAssetOrderStatuses(config.OrderStatuses, orders, reviewTodoCount),
		"recentOrders":   profileAssetRecentOrders(orders, 3),
		"balanceRecords": profileAssetBalanceRecords(pointLogs, 5),
		"bankCards": map[string]interface{}{
			"count":        0,
			"summaryText":  "一期不提供银行卡功能",
			"items":        []map[string]interface{}{},
			"canBind":      false,
			"needIdentity": false,
		},
		"faqLinks":      profileAssetFAQLinks(config.FAQLinks),
		"pointsSummary": pointSummary,
		"configVersion": config.Version,
	}
}

func (s *Server) currentProfileAssetManageConfig() profileAssetManageConfigDTO {
	// 资金、银行卡与提现均不在一期范围。历史配置只保留在服务端兼容，
	// 小程序统一使用积分资产视图，避免旧配置重新展示资金入口。
	return defaultProfileAssetManageConfig()
}

func defaultProfileAssetManageConfig() profileAssetManageConfigDTO {
	return profileAssetManageConfigDTO{
		OverviewLabel:       "积分资产",
		OverviewUpdatedText: "实时同步积分、订单与评价记录",
		AssetStats: []profileAssetStatConfigDTO{
			{Key: "availablePoints", Label: "可用积分", Tone: "green"},
			{Key: "orderCount", Label: "兑换订单", Tone: "yellow"},
			{Key: "reviewTodo", Label: "待评价", Tone: "orange"},
		},
		QuickActions: []profileAssetActionConfigDTO{},
		MenuItems: []profileAssetMenuConfigDTO{
			{Key: "points", Title: "积分明细", Desc: "查看积分获取和使用记录", IconSrc: "/pages/profile/asset-center/manage/assets/fa/list-ul.svg", Tone: "blue", Route: "/pages/profile/asset-center/points/index"},
			{Key: "mall", Title: "积分商城", Desc: "使用积分兑换权益", IconSrc: "/pages/profile/asset-center/manage/assets/fa/bag-shopping.svg", Tone: "green", Route: "/pages/profile/asset-center/mall/index"},
			{Key: "orders", Title: "我的订单", Desc: "查看全部订单", IconSrc: "/pages/profile/asset-center/manage/assets/fa/bag-shopping.svg", Tone: "purple", Route: "/pages/profile/asset-center/orders/index"},
		},
		OrderStatuses: []profileAssetStatusConfigDTO{
			{Key: "pending", Label: "待处理", IconSrc: "/pages/profile/asset-center/manage/assets/fa/hourglass-half.svg", Tone: "blue", Route: "/pages/profile/asset-center/orders/index?status=pending"},
			{Key: "processing", Label: "进行中", IconSrc: "/pages/profile/asset-center/manage/assets/fa/spinner.svg", Tone: "orange", Route: "/pages/profile/asset-center/orders/index?status=pending"},
			{Key: "completed", Label: "已完成", IconSrc: "/pages/profile/asset-center/manage/assets/fa/check.svg", Tone: "green", Route: "/pages/profile/asset-center/orders/index?status=fulfilled"},
			{Key: "refund", Label: "已取消", IconSrc: "/pages/profile/asset-center/manage/assets/fa/rotate-left.svg", Tone: "red", Route: "/pages/profile/asset-center/orders/index?status=canceled"},
			{Key: "review", Label: "待评价", IconSrc: "/pages/profile/asset-center/manage/assets/fa/star.svg", Tone: "gray", Route: "/pages/profile/service-center/manage/review-manage/index"},
		},
		BankCards: profileAssetBankCardConfigDTO{UnboundText: "一期不提供银行卡功能", BoundSuffix: "张", CanBind: false},
		FAQLinks: []profileAssetFAQConfigDTO{
			{Key: "pointsUse", Label: "积分有什么用？", Answer: "积分可用于积分商城兑换；具体商品以商城展示为准。"},
			{Key: "pointsRecord", Label: "如何查看积分记录？", Answer: "可在积分明细中查看积分获取和使用记录。"},
		},
		Version: "2026-07-01",
	}
}

func profileAssetStats(config []profileAssetStatConfigDTO, pointSummary points.Account, orderCount int, reviewTodoCount int) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		value := 0
		switch strings.TrimSpace(item.Key) {
		case "availablePoints":
			value = int(pointSummary.AvailablePoints)
		case "orderCount":
			value = orderCount
		case "reviewTodo":
			value = reviewTodoCount
		}
		stat := map[string]interface{}{
			"key":   item.Key,
			"label": item.Label,
			"value": strconv.Itoa(value),
		}
		if strings.TrimSpace(item.Tone) != "" {
			stat["tone"] = item.Tone
		}
		result = append(result, stat)
	}
	return result
}

func profileAssetQuickActions(config []profileAssetActionConfigDTO) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		result = append(result, map[string]interface{}{
			"key":            item.Key,
			"label":          item.Label,
			"tone":           item.Tone,
			"iconSrc":        item.IconSrc,
			"enabled":        item.Enabled,
			"disabledReason": item.DisabledReason,
		})
	}
	return result
}

func profileAssetMenuItems(config []profileAssetMenuConfigDTO, bankCards profileAssetBankCardConfigDTO, bankCardCount int) []map[string]interface{} {
	cardText := strings.TrimSpace(bankCards.UnboundText)
	if bankCardCount > 0 {
		cardText = "已绑定" + strconv.Itoa(bankCardCount) + strings.TrimSpace(bankCards.BoundSuffix)
	}
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		value := item.Value
		if item.Key == "bankCards" {
			value = cardText
		}
		result = append(result, map[string]interface{}{
			"key":     item.Key,
			"title":   item.Title,
			"desc":    item.Desc,
			"value":   value,
			"iconSrc": item.IconSrc,
			"tone":    item.Tone,
			"route":   item.Route,
		})
	}
	return result
}

func profileAssetOrderStatuses(config []profileAssetStatusConfigDTO, orders []redemption.Order, reviewTodoCount int) []map[string]interface{} {
	counts := map[string]int{}
	for _, order := range orders {
		switch order.Status {
		case "pending":
			counts["pending"]++
		case "approved":
			counts["processing"]++
		case "fulfilled":
			counts["completed"]++
		case "rejected", "canceled":
			counts["refund"]++
		}
	}
	counts["review"] = reviewTodoCount
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		count := counts[item.Key]
		result = append(result, map[string]interface{}{
			"key":       item.Key,
			"label":     item.Label,
			"count":     count,
			"countText": strconv.Itoa(count),
			"iconSrc":   item.IconSrc,
			"tone":      item.Tone,
			"route":     item.Route,
		})
	}
	return result
}

func profileAssetFAQLinks(config []profileAssetFAQConfigDTO) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		result = append(result, map[string]interface{}{
			"key":    item.Key,
			"label":  item.Label,
			"answer": item.Answer,
		})
	}
	return result
}

func profileAssetRecentOrders(orders []redemption.Order, limit int) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, minInt(len(orders), limit))
	config := defaultRedemptionOrderPageConfig()
	for i := len(orders) - 1; i >= 0 && len(result) < limit; i-- {
		dto := buildAppRedemptionOrderDTO(orders[i], config)
		result = append(result, map[string]interface{}{
			"id":         orders[i].OrderNo,
			"orderId":    orders[i].ID,
			"title":      dto.Title,
			"imageUrl":   dto.ImageURL,
			"status":     dto.StatusText,
			"statusKey":  dto.StatusKey,
			"statusTone": dto.StatusTone,
			"time":       dto.CreatedAt,
			"amount":     dto.PointsText,
			"route":      "/pages/profile/asset-center/orders/index?orderId=" + strconv.FormatInt(orders[i].ID, 10),
		})
	}
	return result
}

func profileAssetBalanceRecords(pointLogs []points.Log, limit int) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, limit)
	for _, item := range pointLogs {
		if len(result) >= limit {
			return result
		}
		result = append(result, map[string]interface{}{
			"id":        "points-" + strconv.FormatInt(item.ID, 10),
			"title":     "积分变动",
			"desc":      defaultString(item.Reason, item.BizType),
			"amount":    strconv.Itoa(item.ChangeValue) + "积分",
			"type":      "points",
			"status":    "已记录",
			"createdAt": item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}
