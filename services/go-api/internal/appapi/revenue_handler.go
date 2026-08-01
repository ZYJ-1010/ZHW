package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/revenue"
)

type revenueService interface {
	CreateTemplate(req revenue.TemplateRequest) (revenue.Template, error)
	Templates() []revenue.Template
	TemplatesStrict() ([]revenue.Template, error)
	UpsertRule(req revenue.RuleRequest) (revenue.Rule, error)
	Rules(templateID int64) []revenue.Rule
	RulesStrict(templateID int64) ([]revenue.Rule, error)
	Preview(req revenue.CalculateRequest) (revenue.Preview, error)
	Generate(req revenue.CalculateRequest) (revenue.Record, error)
	Records() []revenue.Record
	RecordsStrict() ([]revenue.Record, error)
	Record(id int64) (revenue.Record, error)
	Freeze(recordID int64, reason string) (revenue.Record, error)
	FreezeByGame(gameID int64, reason string) (revenue.Record, bool, error)
	RestoreFrozenByGame(gameID int64, reason string) (revenue.Record, bool, error)
	Settle(recordID int64, method string, proofNo string) (revenue.Record, revenue.Settlement, error)
	Settlements() []revenue.Settlement
	SettlementsStrict() ([]revenue.Settlement, error)
	IncomeAccount(userID int64) revenue.IncomeAccount
	IncomeAccountStrict(userID int64) (revenue.IncomeAccount, error)
	IncomeSummary(userID int64) revenue.IncomeSummary
	IncomeSummaryStrict(userID int64) (revenue.IncomeSummary, error)
	IncomeLogs(userID int64, status string) []revenue.IncomeLog
	IncomeLogsStrict(userID int64, status string) ([]revenue.IncomeLog, error)
}

func (s *Server) revenueTemplates(w http.ResponseWriter, r *http.Request) {
	items, err := s.revenue.TemplatesStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取分润模板失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) appRevenueTemplates(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireUser(w, r); !ok {
		return
	}
	templates, err := s.revenue.TemplatesStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取分润模板失败，请稍后重试")
		return
	}
	config := s.currentProfitTemplateConfig()
	httpx.OK(w, map[string]interface{}{
		"items":             templates,
		"templates":         templates,
		"depositRuleText":   config.DepositRuleText,
		"depositNoticeText": config.DepositNoticeText,
		"deposit": map[string]string{
			"ruleText":   config.DepositRuleText,
			"noticeText": config.DepositNoticeText,
		},
	})
}

type profitTemplateConfigDTO struct {
	DepositRuleText   string `json:"depositRuleText"`
	DepositNoticeText string `json:"depositNoticeText"`
	Version           string `json:"version"`
}

const profitTemplateConfigKey = "game.profit_template_config"

// revenuePointsRewardConfig controls the only automatic source of redeemable
// points: a user's own, successfully settled paid-game revenue share.
// PointsPerRevenueBPS uses the same basis-point convention as revenue rules:
// 1000 means a 10% conversion. One yuan of settled revenue is the base unit.
type revenuePointsRewardConfigDTO struct {
	Enabled                bool `json:"enabled"`
	PointsPerRevenueBPS    int  `json:"pointsPerRevenueBps"`
	MaxPointsPerSettlement int  `json:"maxPointsPerSettlement"`
}

const revenuePointsRewardConfigKey = "revenue.points_reward_config"

func (s *Server) adminRevenuePointsRewardConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentRevenuePointsRewardConfig()})
	case http.MethodPut:
		var req revenuePointsRewardConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "分润积分规则格式错误")
			return
		}
		config, err := normalizeRevenuePointsRewardConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig == nil || s.systemConfig.Set(revenuePointsRewardConfigKey, config) != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存分润积分规则失败")
			return
		}
		s.recordOperation(r, "revenue_points_reward_config:update", "system_config", revenuePointsRewardConfigKey, map[string]interface{}{
			"enabled": config.Enabled, "pointsPerRevenueBps": config.PointsPerRevenueBPS, "maxPointsPerSettlement": config.MaxPointsPerSettlement,
		})
		httpx.OK(w, map[string]interface{}{"config": config})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "不支持当前请求方式")
	}
}

func (s *Server) currentRevenuePointsRewardConfig() revenuePointsRewardConfigDTO {
	var config revenuePointsRewardConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(revenuePointsRewardConfigKey, &config) {
		if normalized, err := normalizeRevenuePointsRewardConfig(config); err == nil {
			return normalized
		}
	}
	return defaultRevenuePointsRewardConfig()
}

func defaultRevenuePointsRewardConfig() revenuePointsRewardConfigDTO {
	return revenuePointsRewardConfigDTO{Enabled: true, PointsPerRevenueBPS: 1000}
}

func normalizeRevenuePointsRewardConfig(config revenuePointsRewardConfigDTO) (revenuePointsRewardConfigDTO, error) {
	if config.PointsPerRevenueBPS < 0 || config.PointsPerRevenueBPS > 10000 {
		return revenuePointsRewardConfigDTO{}, errors.New("分润积分比例必须在 0% 到 100% 之间")
	}
	if config.MaxPointsPerSettlement < 0 || config.MaxPointsPerSettlement > 1000000 {
		return revenuePointsRewardConfigDTO{}, errors.New("单次结算积分上限必须在 0 到 1000000 之间")
	}
	return config, nil
}

func (s *Server) adminProfitTemplateConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentProfitTemplateConfig()})
	case http.MethodPut:
		var req profitTemplateConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "分润模板配置格式错误")
			return
		}
		config, err := normalizeProfitTemplateConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig != nil {
			if err := s.systemConfig.Set(profitTemplateConfigKey, config); err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存分润模板配置失败")
				return
			}
		}
		s.recordOperation(r, "profit_template_config:update", "system_config", profitTemplateConfigKey, map[string]interface{}{"version": config.Version})
		httpx.OK(w, map[string]interface{}{"config": s.currentProfitTemplateConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "不支持当前请求方式")
	}
}

func (s *Server) currentProfitTemplateConfig() profitTemplateConfigDTO {
	var config profitTemplateConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(profitTemplateConfigKey, &config) {
		if normalized, err := normalizeProfitTemplateConfig(config); err == nil {
			return normalized
		}
	}
	return defaultProfitTemplateConfig()
}

func normalizeProfitTemplateConfig(config profitTemplateConfigDTO) (profitTemplateConfigDTO, error) {
	config.DepositRuleText = strings.TrimSpace(config.DepositRuleText)
	config.DepositNoticeText = strings.TrimSpace(config.DepositNoticeText)
	config.Version = strings.TrimSpace(config.Version)
	if config.DepositRuleText == "" {
		config.DepositRuleText = defaultProfitTemplateConfig().DepositRuleText
	}
	if config.DepositNoticeText == "" {
		config.DepositNoticeText = defaultProfitTemplateConfig().DepositNoticeText
	}
	if len(config.DepositRuleText) > 300 || len(config.DepositNoticeText) > 300 {
		return profitTemplateConfigDTO{}, errors.New("押金规则文案最多 300 个字")
	}
	if config.Version == "" {
		config.Version = "2026-06-30"
	}
	return config, nil
}

func defaultProfitTemplateConfig() profitTemplateConfigDTO {
	return profitTemplateConfigDTO{
		DepositRuleText:   "连续打卡 7 天即完成。完成者拿回押金池金额，未完成者押金由完成者平分。",
		DepositNoticeText: "支付金额：100元 = 服务费10元 + 押金池90元。服务费不退，押金池按完成情况结算。",
		Version:           "2026-06-30",
	}
}

func (s *Server) createRevenueTemplate(w http.ResponseWriter, r *http.Request) {
	var req revenue.TemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	template, err := s.revenue.CreateTemplate(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	s.recordOperation(r, "revenue:template:create", "revenue_template", strconv.FormatInt(template.ID, 10), map[string]interface{}{"gameType": template.GameType})
	httpx.OK(w, template)
}

func (s *Server) revenueRules(w http.ResponseWriter, r *http.Request) {
	templateID := parseIntQuery(r, "templateId")
	items, err := s.revenue.RulesStrict(templateID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取分润规则失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) upsertRevenueRule(w http.ResponseWriter, r *http.Request) {
	var req revenue.RuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	rule, err := s.revenue.UpsertRule(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	s.recordOperation(r, "revenue:rule:upsert", "revenue_rule", strconv.FormatInt(rule.ID, 10), map[string]interface{}{"templateId": rule.TemplateID, "ruleCode": rule.RuleCode, "ruleValue": rule.RuleValue})
	httpx.OK(w, rule)
}

func (s *Server) appRevenuePreview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/revenue-preview")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game.CreatorUserID != userID && !s.games.IsMember(gameID, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅局内成员可查看分润预览")
		return
	}
	templateID := parseIntQuery(r, "templateId")
	if templateID == 0 {
		templates, err := s.revenue.TemplatesStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取分润模板失败，请稍后重试")
			return
		}
		templateID = firstTemplateID(templates)
	}
	preview, err := s.revenue.Preview(revenue.CalculateRequest{
		GameID:     gameID,
		AmountCent: parseIntQuery(r, "amountCent"),
		TemplateID: templateID,
		MemberIDs:  s.games.Members(gameID),
		CreatorID:  game.CreatorUserID,
	})
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	preview, err = s.applyRevenuePreviewBlocks(preview)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报状态失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "view_revenue", "game", gameID, map[string]interface{}{"amountCent": preview.AmountCent})
	httpx.OK(w, preview)
}

func (s *Server) appRevenueSimulate(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	req, ok := s.revenueRequest(w, r)
	if !ok {
		return
	}
	game, err := s.games.Get(req.GameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game.CreatorUserID != userID && !s.games.IsMember(req.GameID, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅局内成员可试算分润")
		return
	}
	if req.TemplateID == 0 {
		templates, err := s.revenue.TemplatesStrict()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取分润模板失败，请稍后重试")
			return
		}
		req.TemplateID = firstTemplateID(templates)
	}
	req.MemberIDs = s.games.Members(req.GameID)
	req.CreatorID = game.CreatorUserID
	preview, err := s.revenue.Preview(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	preview, err = s.applyRevenuePreviewBlocks(preview)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报状态失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "simulate_revenue", "game", req.GameID, map[string]interface{}{"amountCent": preview.AmountCent})
	httpx.OK(w, preview)
}

func (s *Server) adminRevenuePreview(w http.ResponseWriter, r *http.Request) {
	req, ok := s.revenueRequest(w, r)
	if !ok {
		return
	}
	preview, err := s.revenue.Preview(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	preview, err = s.applyRevenuePreviewBlocks(preview)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报状态失败，请稍后重试")
		return
	}
	httpx.OK(w, preview)
}

func (s *Server) generateRevenueRecord(w http.ResponseWriter, r *http.Request) {
	req, ok := s.revenueRequest(w, r)
	if !ok {
		return
	}
	hasOpenReport, err := s.gameHasOpenReport(req.GameID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报状态失败，请稍后重试")
		return
	}
	record, err := s.revenue.Generate(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	if hasOpenReport {
		record, err = s.revenue.Freeze(record.ID, "report_open")
		if err != nil {
			writeRevenueError(w, err)
			return
		}
	}
	httpx.OK(w, record)
}

func (s *Server) revenueRecords(w http.ResponseWriter, r *http.Request) {
	items, err := s.revenue.RecordsStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取分润记录失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) revenueSettlements(w http.ResponseWriter, r *http.Request) {
	items, err := s.revenue.SettlementsStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取结算记录失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) freezeRevenueRecord(w http.ResponseWriter, r *http.Request) {
	recordID, ok := revenueRecordIDFromPath(w, r.URL.Path, "/freeze")
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	record, err := s.revenue.Freeze(recordID, req.Reason)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	httpx.OK(w, record)
}

func (s *Server) settleRevenueRecord(w http.ResponseWriter, r *http.Request) {
	suffix := "/settle"
	if strings.HasSuffix(r.URL.Path, "/settle-offline") {
		suffix = "/settle-offline"
	}
	recordID, ok := revenueRecordIDFromPath(w, r.URL.Path, suffix)
	if !ok {
		return
	}
	var req struct {
		Method  string `json:"method"`
		ProofNo string `json:"proofNo"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	record, settlement, err := s.revenue.Settle(recordID, req.Method, req.ProofNo)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	pointRewards := s.awardRevenueSettlementPoints(record, settlement)
	httpx.OK(w, map[string]interface{}{"record": record, "settlement": settlement, "pointRewards": pointRewards})
}

type revenueSettlementPointRewardDTO struct {
	UserID int64 `json:"userId"`
	Points int   `json:"points"`
}

// awardRevenueSettlementPoints is intentionally called only after revenue has
// been settled. It never rewards free games, platform/rounding entries, or a
// revenue share without a recipient. GrantOnce makes retrying a successful
// settlement safe.
func (s *Server) awardRevenueSettlementPoints(record revenue.Record, settlement revenue.Settlement) []revenueSettlementPointRewardDTO {
	config := s.currentRevenuePointsRewardConfig()
	if !config.Enabled || settlement.ID <= 0 || config.PointsPerRevenueBPS == 0 {
		return []revenueSettlementPointRewardDTO{}
	}
	game, err := s.games.Get(record.GameID)
	if err != nil || game.Price <= 0 {
		return []revenueSettlementPointRewardDTO{}
	}
	amountsByUser := make(map[int64]int64)
	for _, item := range record.Items {
		if item.UserID <= 0 || item.AmountCent <= 0 {
			continue
		}
		amountsByUser[item.UserID] += item.AmountCent
	}
	rewards := make([]revenueSettlementPointRewardDTO, 0, len(amountsByUser))
	for userID, amountCent := range amountsByUser {
		// amount is stored in cents. The denominator converts cents to yuan and
		// basis points to a percentage, so 100 yuan at 10% earns 10 points.
		pointsValue := int((amountCent * int64(config.PointsPerRevenueBPS)) / 1000000)
		if config.MaxPointsPerSettlement > 0 && pointsValue > config.MaxPointsPerSettlement {
			pointsValue = config.MaxPointsPerSettlement
		}
		if pointsValue <= 0 {
			continue
		}
		_, _, granted, grantErr := s.points.GrantOnce(userID, pointsValue, "revenue_settlement_reward", settlement.ID, "有偿局分润结算积分")
		if grantErr == nil && granted {
			rewards = append(rewards, revenueSettlementPointRewardDTO{UserID: userID, Points: pointsValue})
		}
	}
	return rewards
}

func (s *Server) incomeSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	summary, err := s.revenue.IncomeSummaryStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取收益概览失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "view_income_summary", "income", userID, nil)
	httpx.OK(w, summary)
}

func (s *Server) incomeAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	account, err := s.revenue.IncomeAccountStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取收益账户失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "view_income_account", "income", userID, nil)
	httpx.OK(w, account)
}

func (s *Server) incomeLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	items, err := s.revenue.IncomeLogsStrict(userID, status)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取收益明细失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "view_income_logs", "income", userID, map[string]interface{}{"status": status})
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) revenueRequest(w http.ResponseWriter, r *http.Request) (revenue.CalculateRequest, bool) {
	var req revenue.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return revenue.CalculateRequest{}, false
	}
	if len(req.MemberIDs) == 0 && req.GameID > 0 {
		req.MemberIDs = s.games.Members(req.GameID)
	}
	if req.CreatorID == 0 && req.GameID > 0 {
		if game, err := s.games.Get(req.GameID); err == nil {
			req.CreatorID = game.CreatorUserID
		}
	}
	return req, true
}

func revenueRecordIDFromPath(w http.ResponseWriter, path string, suffix string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/admin/revenue/records/"), suffix)
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "分润记录编号错误")
		return 0, false
	}
	return id, true
}

func parseIntQuery(r *http.Request, key string) int64 {
	value, _ := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	return value
}

func firstTemplateID(templates []revenue.Template) int64 {
	for _, template := range templates {
		if template.Status == "active" {
			return template.ID
		}
	}
	return 0
}

func (s *Server) gameHasOpenReport(gameID int64) (bool, error) {
	items, err := s.reports.ListStrict()
	if err != nil {
		return false, err
	}
	for _, report := range items {
		if report.GameID == gameID && report.Status != "closed" {
			return true, nil
		}
	}
	return false, nil
}

func (s *Server) applyRevenuePreviewBlocks(preview revenue.Preview) (revenue.Preview, error) {
	hasOpenReport := false
	var err error
	if preview.GameID > 0 {
		hasOpenReport, err = s.gameHasOpenReport(preview.GameID)
		if err != nil {
			return revenue.Preview{}, err
		}
	}
	if hasOpenReport && !hasRevenueBlockReason(preview.BlockReasons, "disputed") {
		preview.BlockReasons = append(preview.BlockReasons, "disputed")
	}
	preview.CanGenerateRecord = len(preview.BlockReasons) == 0
	return preview, nil
}

func hasRevenueBlockReason(items []string, reason string) bool {
	for _, item := range items {
		if item == reason {
			return true
		}
	}
	return false
}

func writeRevenueError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, revenue.ErrInvalidTemplate):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "分润模板参数无效")
	case errors.Is(err, revenue.ErrTemplateNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "分润模板不存在")
	case errors.Is(err, revenue.ErrReviewIncomplete):
		httpx.Error(w, http.StatusConflict, 40941, "局内评价尚未完成")
	case errors.Is(err, revenue.ErrDuplicateRecord):
		httpx.Error(w, http.StatusConflict, 40942, "该局已生成分润记录")
	case errors.Is(err, revenue.ErrRecordFrozen):
		httpx.Error(w, http.StatusConflict, 40943, "分润记录已冻结")
	case errors.Is(err, revenue.ErrRecordNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "分润记录不存在")
	case errors.Is(err, revenue.ErrRecordNotSettleable):
		httpx.Error(w, http.StatusConflict, 40944, "当前分润记录不可结算")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "分润操作失败")
	}
}
