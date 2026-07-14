package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/orders"
	"zhw-mini/services/go-api/internal/profiles"
)

func (s *Server) appOrderDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	orderIDText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/app/orders/"), "/")
	orderID, err := strconv.ParseInt(orderIDText, 10, 64)
	if err != nil || orderID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "订单 ID 错误")
		return
	}
	order, err := s.orders.GetForUser(userID, orderID)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	httpx.OK(w, order)
}

func (s *Server) paymentPrecreatePlaceholder(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		GameID int64 `json:"gameId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	game, err := s.games.Get(req.GameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game.CreatorUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "只能为自己创建的局预留订单")
		return
	}
	result, err := s.orders.PrecreateFreeNoPay(userID, req.GameID)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	s.recordBehavior(userID, "payment_precreate_placeholder", "order", result.Order.ID, map[string]interface{}{"gameId": req.GameID, "payStatus": result.PayStatus})
	httpx.OK(w, result)
}

func (s *Server) gamePaymentPreview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/payment-preview")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game.CreatorUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "只能查看自己创建的局订单")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"payment": map[string]interface{}{
			"gameId": game.ID, "amount": 0, "amountCent": 0, "currency": "CNY",
			"payStatus": "free_no_pay", "needWechatPay": false,
		},
		"paymentSplits": []interface{}{},
		"rules": []map[string]string{
			{"icon": "✅", "type": "success", "text": "一期免费局不发起真实微信支付"},
			{"icon": "💡", "type": "warning", "text": "当前订单金额与支付状态均以后端返回为准"},
		},
		"pointsDescription": "免费局确认后进入组局流程，积分与成长由服务确认和评价链路沉淀。",
	})
}

func (s *Server) guidePaymentPrecreatePlaceholder(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	result, err := s.orders.PrecreateGuideFeePlaceholder(userID)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	met := true
	qualification, err := s.profiles.UpdateGuideQualification(userID, profiles.UpdateGuideQualificationRequest{PaymentMet: &met})
	if err != nil {
		writeProfileError(w, err)
		return
	}
	s.recordBehavior(userID, "guide_payment_precreate_placeholder", "order", result.Order.ID, map[string]interface{}{"payStatus": result.PayStatus})
	httpx.OK(w, map[string]interface{}{"payment": result, "qualification": qualification})
}

func (s *Server) paymentCallbackPlaceholder(w http.ResponseWriter, r *http.Request) {
	var req orders.CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	result, err := s.orders.RecordCallback(req)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	httpx.OK(w, result)
}

func (s *Server) profitSharingOrderPlaceholder(w http.ResponseWriter, r *http.Request) {
	var req orders.ProfitSharingOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	result, err := s.orders.CreateProfitSharingPlaceholder(req)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	httpx.OK(w, result)
}

func (s *Server) profitSharingOrderQueryPlaceholder(w http.ResponseWriter, r *http.Request) {
	outOrderNo := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/funds/profit-sharing/orders/"), "/")
	result, err := s.orders.ProfitSharingPlaceholder(outOrderNo)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	httpx.OK(w, result)
}

func (s *Server) profitSharingReturnPlaceholder(w http.ResponseWriter, r *http.Request) {
	var req orders.ProfitSharingReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	result, err := s.orders.CreateProfitSharingReturnPlaceholder(req)
	if err != nil {
		writeOrderError(w, err)
		return
	}
	httpx.OK(w, result)
}

func writeOrderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, orders.ErrOrderNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "订单不存在")
	case errors.Is(err, orders.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权查看订单")
	case errors.Is(err, orders.ErrInvalidOrder), errors.Is(err, orders.ErrCallbackInvalid):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "订单参数错误")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "订单操作失败")
	}
}
