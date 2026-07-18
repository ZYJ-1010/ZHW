package appapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/redemption"
)

type appRedemptionOrderDTO struct {
	ID          int64                    `json:"id"`
	OrderID     int64                    `json:"orderId"`
	OrderNo     string                   `json:"orderNo"`
	ItemID      int64                    `json:"itemId"`
	ItemName    string                   `json:"itemName"`
	ImageURL    string                   `json:"imageUrl,omitempty"`
	Title       string                   `json:"title"`
	PointsCost  int                      `json:"pointsCost"`
	PointsText  string                   `json:"pointsText"`
	Status      string                   `json:"status"`
	StatusKey   string                   `json:"statusKey"`
	StatusText  string                   `json:"statusText"`
	StatusTone  string                   `json:"statusTone"`
	CreatedAt   string                   `json:"createdAt"`
	CreatedText string                   `json:"createdAtText"`
	Actions     []map[string]interface{} `json:"actions"`
}

const (
	pointsPageConfigKey          = "points.page_config"
	redemptionOrderPageConfigKey = "redemption.order_page_config"
)

type pointsPageStatDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type pointsPageRuleDTO struct {
	Text   string `json:"text"`
	Strong string `json:"strong"`
	Suffix string `json:"suffix"`
}

type pointsPageConfigDTO struct {
	Stats        []pointsPageStatDTO      `json:"stats"`
	Rules        []pointsPageRuleDTO      `json:"rules"`
	EarnExample  map[string]interface{}   `json:"earnExample"`
	RoleExamples []map[string]interface{} `json:"roleExamples"`
	Filters      []map[string]string      `json:"filters"`
	NoteText     string                   `json:"noteText"`
	Version      string                   `json:"version"`
}

type redemptionOrderPageConfigDTO struct {
	Tabs               []map[string]interface{} `json:"tabs"`
	EmptyText          string                   `json:"emptyText"`
	LogisticsEmptyText string                   `json:"logisticsEmptyText"`
	DetailEmptyText    string                   `json:"detailEmptyText"`
	CancelConfirm      map[string]string        `json:"cancelConfirm"`
	Actions            map[string]string        `json:"actions"`
	Version            string                   `json:"version"`
}

func (s *Server) pointsSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	s.recordBehavior(userID, "view_points_summary", "points", userID, nil)
	httpx.OK(w, s.buildPointsSummaryPayload(userID))
}

func (s *Server) pointsLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	s.recordBehavior(userID, "view_points_logs", "points", userID, nil)
	httpx.OK(w, map[string]interface{}{"items": s.points.Logs(userID)})
}

func (s *Server) adminPointsLogs(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.points.AllLogs()})
}

func (s *Server) buildPointsSummaryPayload(userID int64) map[string]interface{} {
	account := s.points.Summary(userID)
	logs := s.points.Logs(userID)
	config := s.currentPointsPageConfig()
	operationRules := s.currentOperationRules()
	redeemedPoints := 0
	expiredPoints := 0
	expiryCutoff := time.Now().Add(-time.Duration(operationRules.Points.ExpireDays) * 24 * time.Hour)
	for _, item := range logs {
		if item.ChangeValue < 0 {
			redeemedPoints += -item.ChangeValue
		}
		if operationRules.Points.ExpireEnabled && item.ChangeValue > 0 && item.CreatedAt.Before(expiryCutoff) {
			expiredPoints += item.ChangeValue
		}
	}

	return map[string]interface{}{
		"userId":            account.UserID,
		"availablePoints":   account.AvailablePoints,
		"frozenPoints":      account.FrozenPoints,
		"totalEarnedPoints": account.TotalEarnedPoints,
		"redeemedPoints":    redeemedPoints,
		"expiredPoints":     expiredPoints,
		"updatedAt":         account.UpdatedAt,
		"stats":             config.Stats,
		"rules":             config.Rules,
		"earnExample":       config.EarnExample,
		"roleExamples":      config.RoleExamples,
		"filters":           config.Filters,
		"noteText":          config.NoteText,
		"version":           config.Version,
	}
}

func (s *Server) currentPointsPageConfig() pointsPageConfigDTO {
	var stored pointsPageConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(pointsPageConfigKey, &stored) && len(stored.Stats) > 0 && len(stored.Filters) > 0 {
		return stored
	}
	return defaultPointsPageConfig()
}

func (s *Server) currentRedemptionOrderPageConfig() redemptionOrderPageConfigDTO {
	var stored redemptionOrderPageConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(redemptionOrderPageConfigKey, &stored) && len(stored.Tabs) > 0 {
		return stored
	}
	return defaultRedemptionOrderPageConfig()
}

func defaultPointsPageConfig() pointsPageConfigDTO {
	return pointsPageConfigDTO{
		Stats: []pointsPageStatDTO{
			{Key: "total", Label: "累计积分"},
			{Key: "redeemed", Label: "已兑换"},
			{Key: "expired", Label: "过期积分"},
		},
		Rules: []pointsPageRuleDTO{
			{Text: "服务完成、举报核实、平台活动等行为可产生积分，具体比例以后台配置为准", Strong: "后台规则"},
			{Text: "积分有效期按平台规则执行，到期后由后台任务处理", Strong: "有效期规则"},
			{Text: "积分仅可兑换", Strong: "平台限定商品", Suffix: "，不可提现或抵扣付费局"},
		},
		EarnExample: map[string]interface{}{
			"title":    "可获得积分的行为",
			"subtitle": "服务分润、举报核实、活动奖励",
			"points":   "+20",
			"rows": []map[string]string{
				{"label": "举报核实奖励", "value": "后台确认后发放"},
				{"label": "服务分润积分", "value": "按后台比例生成"},
			},
			"result": "积分以后台流水为准",
		},
		RoleExamples: []map[string]interface{}{
			{"key": "expert", "role": "行家服务完成", "amount": "按分润金额", "points": "+积分", "iconText": "行"},
			{"key": "guide", "role": "领路人引荐成功", "amount": "按引荐收益", "points": "+积分", "iconText": "领"},
			{"key": "platform", "role": "平台核实奖励", "amount": "后台配置", "points": "+积分", "iconText": "奖"},
		},
		Filters: []map[string]string{
			{"key": "all", "label": "全部", "tone": "all"},
			{"key": "income", "label": "收入", "tone": "income"},
			{"key": "expense", "label": "支出", "tone": "expense"},
		},
		NoteText: "积分规则、比例、有效期和兑换限制均以后端后台配置为准。积分不可提现，不可支付付费局，仅可兑换平台限定商品。",
		Version:  "2026-07-01",
	}
}

func defaultRedemptionOrderPageConfig() redemptionOrderPageConfigDTO {
	return redemptionOrderPageConfigDTO{
		Tabs: []map[string]interface{}{
			{"key": "all", "label": "\u5168\u90e8"},
			{"key": "pending", "label": "\u5f85\u5ba1\u6838"},
			{"key": "approved", "label": "\u5f85\u53d1\u653e"},
			{"key": "fulfilled", "label": "\u5df2\u5b8c\u6210"},
			{"key": "rejected", "label": "\u5df2\u9a73\u56de"},
			{"key": "canceled", "label": "\u5df2\u53d6\u6d88"},
		},
		EmptyText:          "\u6682\u65e0\u5151\u6362\u8ba2\u5355",
		LogisticsEmptyText: "\u6682\u65e0\u7269\u6d41\u4fe1\u606f",
		DetailEmptyText:    "\u6682\u65e0\u8ba2\u5355\u8be6\u60c5",
		CancelConfirm: map[string]string{
			"title":       "\u53d6\u6d88\u8ba2\u5355",
			"content":     "\u53d6\u6d88\u540e\u79ef\u5206\u5c06\u9000\u56de\u5230\u8d26\u6237\uff0c\u786e\u8ba4\u53d6\u6d88\u8fd9\u4e2a\u5151\u6362\u8ba2\u5355\u5417\uff1f",
			"confirmText": "\u786e\u8ba4\u53d6\u6d88",
			"cancelText":  "\u518d\u60f3\u60f3",
			"reason":      "\u7528\u6237\u4e3b\u52a8\u53d6\u6d88",
		},
		Actions: map[string]string{
			"detail":    "\u67e5\u770b\u8be6\u60c5",
			"cancel":    "\u53d6\u6d88\u8ba2\u5355",
			"logistics": "\u67e5\u770b\u7269\u6d41",
			"again":     "\u518d\u6b21\u5151\u6362",
		},
		Version: "2026-07-01",
	}
}

func (s *Server) redemptionItems(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireUser(w, r); !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.redemption.Items()})
}

func (s *Server) adminRedemptionItems(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.redemption.AdminItems()})
}

func (s *Server) createAdminRedemptionItem(w http.ResponseWriter, r *http.Request) {
	var req redemption.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.redemption.CreateItem(req)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	s.recordOperation(r, "redemption:item:create", "redemption_item", strconv.FormatInt(item.ID, 10), map[string]interface{}{"pointsCost": item.PointsCost, "stock": item.Stock})
	httpx.OK(w, item)
}

func (s *Server) updateAdminRedemptionItem(w http.ResponseWriter, r *http.Request) {
	itemID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/redemption/items/", "")
	if !ok {
		return
	}
	var req redemption.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.redemption.UpdateItem(itemID, req)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	s.recordOperation(r, "redemption:item:update", "redemption_item", strconv.FormatInt(item.ID, 10), map[string]interface{}{"status": item.Status, "stock": item.Stock})
	httpx.OK(w, item)
}

func (s *Server) createRedemptionOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req redemption.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	order, err := s.redemption.CreateOrder(userID, req)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	s.recordBehavior(userID, "create_redemption_order", "redemption_order", order.ID, map[string]interface{}{"itemId": order.ItemID, "pointsCost": order.PointsCost})
	httpx.OK(w, order)
}

func (s *Server) myRedemptionOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	config := s.currentRedemptionOrderPageConfig()
	orders := s.redemption.OrdersForUser(userID)
	items := make([]appRedemptionOrderDTO, 0, len(orders))
	for _, order := range orders {
		if !matchAppRedemptionStatus(status, order.Status) {
			continue
		}
		items = append(items, buildAppRedemptionOrderDTO(order, config))
	}
	httpx.OK(w, map[string]interface{}{
		"items":      items,
		"orders":     items,
		"tabs":       config.Tabs,
		"emptyText":  config.EmptyText,
		"pageConfig": config,
	})
}

func (s *Server) redemptionOrderLogistics(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	orderID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/profile/points/orders/", "/logistics")
	if !ok {
		return
	}
	order, err := s.redemption.OrderForUser(userID, orderID)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	httpx.OK(w, s.buildRedemptionLogisticsDTO(order))
}

func (s *Server) redemptionOrderDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	orderID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/redemption/orders/", "")
	if !ok {
		return
	}
	order, err := s.redemption.OrderForUser(userID, orderID)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	s.recordBehavior(userID, "view_redemption_order", "redemption_order", order.ID, nil)
	httpx.OK(w, s.buildAppRedemptionOrderDetailDTO(order))
}

func (s *Server) cancelRedemptionOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	orderID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/redemption/orders/", "/cancel")
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
			return
		}
	}
	order, err := s.redemption.CancelOrder(userID, orderID, req.Reason)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	s.recordBehavior(userID, "cancel_redemption_order", "redemption_order", order.ID, map[string]interface{}{"reason": order.ReviewReason})
	httpx.OK(w, map[string]interface{}{
		"order":         buildAppRedemptionOrderDTO(order, s.currentRedemptionOrderPageConfig()),
		"pointsSummary": s.points.Summary(userID),
		"message":       "订单已取消，积分已退回",
	})
}

func (s *Server) adminRedemptionOrders(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.redemption.AdminOrders()})
}

func (s *Server) reviewAdminRedemptionOrder(w http.ResponseWriter, r *http.Request) {
	orderID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/redemption/orders/", "/review")
	if !ok {
		return
	}
	var req redemption.ReviewOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	order, err := s.redemption.ReviewOrder(orderID, parseInt64Header(r, "X-Admin-ID"), req)
	if err != nil {
		writeRedemptionError(w, err)
		return
	}
	s.recordOperation(r, "redemption:order:review", "redemption_order", strconv.FormatInt(order.ID, 10), map[string]interface{}{"status": order.Status, "reason": order.ReviewReason})
	httpx.OK(w, order)
}

func buildAppRedemptionOrderDTO(order redemption.Order, config redemptionOrderPageConfigDTO) appRedemptionOrderDTO {
	statusKey, statusText, statusTone := appRedemptionStatus(order.Status)
	actions := make([]map[string]interface{}, 0, 3)
	actions = append(actions, map[string]interface{}{"key": "detail", "label": redemptionOrderActionLabel(config, "detail"), "type": "ghost"})
	if order.Status == "pending" {
		actions = append(actions, map[string]interface{}{"key": "cancel", "label": redemptionOrderActionLabel(config, "cancel"), "type": "ghost"})
	}
	if order.Status == "approved" || order.Status == "fulfilled" {
		actions = append(actions, map[string]interface{}{"key": "logistics", "label": redemptionOrderActionLabel(config, "logistics"), "type": "ghost"})
	}
	createdAt := order.CreatedAt.Format("2006-01-02 15:04:05")
	return appRedemptionOrderDTO{
		ID:          order.ID,
		OrderID:     order.ID,
		OrderNo:     order.OrderNo,
		ItemID:      order.ItemID,
		ItemName:    order.ItemName,
		ImageURL:    order.ItemImageURL,
		Title:       order.ItemName,
		PointsCost:  order.PointsCost,
		PointsText:  fmt.Sprintf("%d\u79ef\u5206", order.PointsCost),
		Status:      order.Status,
		StatusKey:   statusKey,
		StatusText:  statusText,
		StatusTone:  statusTone,
		CreatedAt:   createdAt,
		CreatedText: "\u5151\u6362\u65f6\u95f4: " + createdAt,
		Actions:     actions,
	}
}

func (s *Server) buildAppRedemptionOrderDetailDTO(order redemption.Order) map[string]interface{} {
	config := s.currentRedemptionOrderPageConfig()
	dto := buildAppRedemptionOrderDTO(order, config)
	updatedAt := order.UpdatedAt.Format("2006-01-02 15:04:05")
	rows := []map[string]interface{}{
		{"label": "订单编号", "value": order.OrderNo},
		{"label": "兑换商品", "value": order.ItemName},
		{"label": "消耗积分", "value": fmt.Sprintf("%d积分", order.PointsCost)},
		{"label": "订单状态", "value": dto.StatusText},
		{"label": "提交时间", "value": dto.CreatedAt},
		{"label": "更新时间", "value": updatedAt},
	}
	if strings.TrimSpace(order.ReviewReason) != "" {
		rows = append(rows, map[string]interface{}{"label": "处理备注", "value": order.ReviewReason})
	}
	return map[string]interface{}{
		"order":      dto,
		"detailRows": rows,
		"emptyText":  config.DetailEmptyText,
		"logistics":  s.buildRedemptionLogisticsDTO(order),
	}
}

func redemptionOrderActionLabel(config redemptionOrderPageConfigDTO, key string) string {
	if config.Actions != nil && strings.TrimSpace(config.Actions[key]) != "" {
		return config.Actions[key]
	}
	return key
}

func appRedemptionStatus(status string) (string, string, string) {
	switch status {
	case "pending":
		return "pending", "\u5f85\u5ba1\u6838", "orange"
	case "approved":
		return "approved", "\u5f85\u53d1\u653e", "blue"
	case "fulfilled":
		return "fulfilled", "\u5df2\u5b8c\u6210", "success"
	case "rejected":
		return "rejected", "\u5df2\u9a73\u56de", "muted"
	case "canceled":
		return "canceled", "\u5df2\u53d6\u6d88", "muted"
	default:
		return status, status, "blue"
	}
}

func matchAppRedemptionStatus(query string, status string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" || query == "all" {
		return true
	}
	if query == strings.ToLower(status) {
		return true
	}
	switch query {
	case "pending_ship", "pendingship":
		return status == "approved"
	case "shipping", "delivering":
		return status == "approved"
	case "completed", "complete":
		return status == "fulfilled"
	case "cancelled", "canceled":
		return status == "canceled"
	default:
		return false
	}
}

func (s *Server) buildRedemptionLogisticsDTO(order redemption.Order) map[string]interface{} {
	statusKey, statusText, _ := appRedemptionStatus(order.Status)
	config := s.currentRedemptionOrderPageConfig()
	updatedAt := order.UpdatedAt.Format("2006-01-02 15:04:05")
	createdAt := order.CreatedAt.Format("2006-01-02 15:04:05")
	timeline := []map[string]interface{}{
		{"id": "created", "desc": "\u5151\u6362\u8ba2\u5355\u5df2\u63d0\u4ea4", "time": createdAt, "active": order.Status == "pending"},
	}
	switch order.Status {
	case "approved":
		timeline = append([]map[string]interface{}{{"id": "approved", "desc": "\u8ba2\u5355\u5df2\u5ba1\u6838\u901a\u8fc7\uff0c\u7b49\u5f85\u540e\u53f0\u53d1\u653e", "time": updatedAt, "active": true}}, timeline...)
	case "fulfilled":
		timeline = append([]map[string]interface{}{{"id": "fulfilled", "desc": "\u5151\u6362\u6743\u76ca\u5df2\u5b8c\u6210\u53d1\u653e", "time": updatedAt, "active": true}}, timeline...)
	case "rejected":
		timeline = append([]map[string]interface{}{{"id": "rejected", "desc": "\u8ba2\u5355\u5df2\u9a73\u56de\uff0c\u79ef\u5206\u5df2\u6309\u89c4\u5219\u9000\u56de", "time": updatedAt, "active": true}}, timeline...)
	case "canceled":
		timeline = append([]map[string]interface{}{{"id": "canceled", "desc": "\u8ba2\u5355\u5df2\u53d6\u6d88\uff0c\u79ef\u5206\u5df2\u9000\u56de", "time": updatedAt, "active": true}}, timeline...)
	default:
		timeline[0]["active"] = true
	}
	return map[string]interface{}{
		"orderId":    order.ID,
		"orderNo":    order.OrderNo,
		"status":     statusKey,
		"statusText": statusText,
		"courier": map[string]interface{}{
			"name":       "\u5e73\u53f0\u53d1\u653e",
			"trackingNo": "",
		},
		"timeline":  timeline,
		"emptyText": config.LogisticsEmptyText,
	}
}

func writeRedemptionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, redemption.ErrInvalidOrder), errors.Is(err, points.ErrInvalidPoints):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "兑换参数错误")
	case errors.Is(err, redemption.ErrItemNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "兑换项目不存在")
	case errors.Is(err, redemption.ErrOrderNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "兑换订单不存在")
	case errors.Is(err, redemption.ErrItemInactive):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "兑换项目已下架")
	case errors.Is(err, redemption.ErrInvalidStatus):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "兑换状态不允许")
	case errors.Is(err, redemption.ErrInsufficientStock):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "库存不足")
	case errors.Is(err, redemption.ErrInsufficientPoints), errors.Is(err, points.ErrInsufficientPoints):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "积分不足")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "兑换操作失败")
	}
}
