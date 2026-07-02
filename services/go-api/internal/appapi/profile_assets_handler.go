package appapi

import (
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/redemption"
	"zhw-mini/services/go-api/internal/revenue"
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

	orders := s.redemption.OrdersForUser(userID)
	httpx.OK(w, s.buildProfileAssetsPayload(userID, orders, len(todos)))
}

func (s *Server) buildProfileAssetsPayload(userID int64, orders []redemption.Order, reviewTodoCount int) map[string]interface{} {
	income := s.revenue.IncomeSummary(userID)
	pointSummary := s.points.Summary(userID)
	totalAssetCent := income.SettledCent + income.PendingCent
	config := s.currentProfileAssetManageConfig()

	return map[string]interface{}{
		"overview": map[string]interface{}{
			"label":       config.OverviewLabel,
			"value":       moneyYuanText(totalAssetCent),
			"amountCent":  totalAssetCent,
			"points":      pointSummary.AvailablePoints,
			"updatedText": config.OverviewUpdatedText,
		},
		"assetStats":     profileAssetStats(config.AssetStats, income),
		"quickActions":   profileAssetQuickActions(config.QuickActions, income),
		"menuItems":      profileAssetMenuItems(config.MenuItems, config.BankCards, 0),
		"orderStatuses":  profileAssetOrderStatuses(config.OrderStatuses, orders, reviewTodoCount),
		"recentOrders":   profileAssetRecentOrders(orders, 3),
		"balanceRecords": profileAssetBalanceRecords(s.revenue.IncomeLogs(userID, ""), s.points.Logs(userID), 5),
		"bankCards": map[string]interface{}{
			"count":        0,
			"summaryText":  config.BankCards.UnboundText,
			"items":        []map[string]interface{}{},
			"canBind":      config.BankCards.CanBind,
			"needIdentity": false,
		},
		"faqLinks":      profileAssetFAQLinks(config.FAQLinks),
		"pointsSummary": pointSummary,
		"incomeSummary": income,
		"configVersion": config.Version,
	}
}

func (s *Server) currentProfileAssetManageConfig() profileAssetManageConfigDTO {
	var stored profileAssetManageConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(profileAssetManageConfigKey, &stored) && len(stored.AssetStats) > 0 && len(stored.OrderStatuses) > 0 {
		return stored
	}
	return defaultProfileAssetManageConfig()
}

func defaultProfileAssetManageConfig() profileAssetManageConfigDTO {
	return profileAssetManageConfigDTO{
		OverviewLabel:       "总资产（元）",
		OverviewUpdatedText: "实时同步分润与积分账户",
		AssetStats: []profileAssetStatConfigDTO{
			{Key: "totalDealAmount", Label: "总成交额"},
			{Key: "withdrawable", Label: "可提现", Tone: "green"},
			{Key: "pendingSettlement", Label: "待结算", Tone: "yellow"},
		},
		QuickActions: []profileAssetActionConfigDTO{
			{Key: "withdraw", Label: "提现", Tone: "green", IconSrc: "/pages/profile/asset-center/manage/assets/fa/download.svg"},
			{Key: "recharge", Label: "充值", Tone: "blue", IconSrc: "/pages/profile/asset-center/manage/assets/fa/plus.svg", Enabled: false, DisabledReason: "一期未接真实支付充值"},
		},
		MenuItems: []profileAssetMenuConfigDTO{
			{Key: "balance", Title: "余额明细", Desc: "收入支出记录", IconSrc: "/pages/profile/asset-center/manage/assets/fa/list-ul.svg", Tone: "blue"},
			{Key: "bankCards", Title: "银行卡", Desc: "管理收款账户", IconSrc: "/pages/profile/asset-center/manage/assets/fa/credit-card.svg", Tone: "green"},
			{Key: "orders", Title: "我的订单", Desc: "查看全部订单", IconSrc: "/pages/profile/asset-center/manage/assets/fa/bag-shopping.svg", Tone: "purple", Route: "/pages/profile/asset-center/orders/index"},
		},
		OrderStatuses: []profileAssetStatusConfigDTO{
			{Key: "pendingPay", Label: "待付款", IconSrc: "/pages/profile/asset-center/manage/assets/fa/hourglass-half.svg", Tone: "blue", Route: "/pages/profile/asset-center/orders/index?status=pending_pay"},
			{Key: "processing", Label: "进行中", IconSrc: "/pages/profile/asset-center/manage/assets/fa/spinner.svg", Tone: "orange", Route: "/pages/profile/asset-center/orders/index?status=pending"},
			{Key: "completed", Label: "已完成", IconSrc: "/pages/profile/asset-center/manage/assets/fa/check.svg", Tone: "green", Route: "/pages/profile/asset-center/orders/index?status=fulfilled"},
			{Key: "refund", Label: "退款/售后", IconSrc: "/pages/profile/asset-center/manage/assets/fa/rotate-left.svg", Tone: "red", Route: "/pages/profile/asset-center/orders/index?status=canceled"},
			{Key: "review", Label: "待评价", IconSrc: "/pages/profile/asset-center/manage/assets/fa/star.svg", Tone: "gray", Route: "/pages/profile/service-center/manage/review-manage/index"},
		},
		BankCards: profileAssetBankCardConfigDTO{UnboundText: "未绑定", BoundSuffix: "张", CanBind: true},
		FAQLinks: []profileAssetFAQConfigDTO{
			{Key: "withdrawArrival", Label: "提现多久到账？", Answer: "提现需在后台财务审核后处理，具体到账时间以后续支付通道规则为准。"},
			{Key: "bindBankCard", Label: "如何绑定银行卡？", Answer: "银行卡绑定入口已预留，正式资金通道接入后开放。"},
		},
		Version: "2026-07-01",
	}
}

func profileAssetStats(config []profileAssetStatConfigDTO, income revenue.IncomeSummary) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		amountCent := profileAssetAmountByKey(item.Key, income)
		stat := map[string]interface{}{
			"key":        item.Key,
			"label":      item.Label,
			"value":      moneyYuanText(amountCent),
			"amountCent": amountCent,
		}
		if strings.TrimSpace(item.Tone) != "" {
			stat["tone"] = item.Tone
		}
		result = append(result, stat)
	}
	return result
}

func profileAssetAmountByKey(key string, income revenue.IncomeSummary) int64 {
	switch strings.TrimSpace(key) {
	case "totalDealAmount":
		return income.TotalCent
	case "withdrawable":
		return income.SettledCent
	case "pendingSettlement":
		return income.PendingCent
	default:
		return 0
	}
}

func profileAssetQuickActions(config []profileAssetActionConfigDTO, income revenue.IncomeSummary) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(config))
	for _, item := range config {
		enabled := item.Enabled
		if item.Key == "withdraw" {
			enabled = income.SettledCent > 0
		}
		result = append(result, map[string]interface{}{
			"key":            item.Key,
			"label":          item.Label,
			"tone":           item.Tone,
			"iconSrc":        item.IconSrc,
			"enabled":        enabled,
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
		case "pending", "approved":
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

func profileAssetBalanceRecords(incomeLogs []revenue.IncomeLog, pointLogs []points.Log, limit int) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, limit)
	for _, item := range incomeLogs {
		if len(result) >= limit {
			return result
		}
		result = append(result, map[string]interface{}{
			"id":        "income-" + strconv.FormatInt(item.RecordID, 10),
			"title":     "分润收益",
			"desc":      item.RecordNo,
			"amount":    moneyYuanText(item.AmountCent),
			"type":      "income",
			"status":    incomeStatusText(item.Status),
			"createdAt": item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
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
