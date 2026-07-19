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
	UpsertRule(req revenue.RuleRequest) (revenue.Rule, error)
	Rules(templateID int64) []revenue.Rule
	Preview(req revenue.CalculateRequest) (revenue.Preview, error)
	Generate(req revenue.CalculateRequest) (revenue.Record, error)
	Records() []revenue.Record
	Record(id int64) (revenue.Record, error)
	Freeze(recordID int64, reason string) (revenue.Record, error)
	FreezeByGame(gameID int64, reason string) (revenue.Record, bool, error)
	RestoreFrozenByGame(gameID int64, reason string) (revenue.Record, bool, error)
	Settle(recordID int64, method string, proofNo string) (revenue.Record, revenue.Settlement, error)
	Settlements() []revenue.Settlement
	IncomeAccount(userID int64) revenue.IncomeAccount
	IncomeSummary(userID int64) revenue.IncomeSummary
	IncomeLogs(userID int64, status string) []revenue.IncomeLog
}

func (s *Server) revenueTemplates(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.revenue.Templates()})
}

func (s *Server) appRevenueTemplates(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireUser(w, r); !ok {
		return
	}
	config := s.currentProfitTemplateConfig()
	httpx.OK(w, map[string]interface{}{
		"items":             s.revenue.Templates(),
		"templates":         s.revenue.Templates(),
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

func (s *Server) adminProfitTemplateConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentProfitTemplateConfig()})
	case http.MethodPut:
		var req profitTemplateConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid profit template config")
			return
		}
		config, err := normalizeProfitTemplateConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if s.systemConfig != nil {
			if err := s.systemConfig.Set(profitTemplateConfigKey, config); err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save profit template config failed")
				return
			}
		}
		s.recordOperation(r, "profit_template_config:update", "system_config", profitTemplateConfigKey, map[string]interface{}{"version": config.Version})
		httpx.OK(w, map[string]interface{}{"config": s.currentProfitTemplateConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
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
		return profitTemplateConfigDTO{}, errors.New("deposit config text too long")
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
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
	httpx.OK(w, map[string]interface{}{"items": s.revenue.Rules(templateID)})
}

func (s *Server) upsertRevenueRule(w http.ResponseWriter, r *http.Request) {
	var req revenue.RuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	templateID := parseIntQuery(r, "templateId")
	if templateID == 0 {
		templateID = firstTemplateID(s.revenue.Templates())
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
	preview = s.applyRevenuePreviewBlocks(preview)
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	if req.TemplateID == 0 {
		req.TemplateID = firstTemplateID(s.revenue.Templates())
	}
	req.MemberIDs = s.games.Members(req.GameID)
	req.CreatorID = game.CreatorUserID
	preview, err := s.revenue.Preview(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	preview = s.applyRevenuePreviewBlocks(preview)
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
	preview = s.applyRevenuePreviewBlocks(preview)
	httpx.OK(w, preview)
}

func (s *Server) generateRevenueRecord(w http.ResponseWriter, r *http.Request) {
	req, ok := s.revenueRequest(w, r)
	if !ok {
		return
	}
	record, err := s.revenue.Generate(req)
	if err != nil {
		writeRevenueError(w, err)
		return
	}
	if s.gameHasOpenReport(req.GameID) {
		record, err = s.revenue.Freeze(record.ID, "report_open")
		if err != nil {
			writeRevenueError(w, err)
			return
		}
	}
	httpx.OK(w, record)
}

func (s *Server) revenueRecords(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.revenue.Records()})
}

func (s *Server) revenueSettlements(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.revenue.Settlements()})
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
	httpx.OK(w, map[string]interface{}{"record": record, "settlement": settlement})
}

func (s *Server) incomeSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	s.recordBehavior(userID, "view_income_summary", "income", userID, nil)
	httpx.OK(w, s.revenue.IncomeSummary(userID))
}

func (s *Server) incomeAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	s.recordBehavior(userID, "view_income_account", "income", userID, nil)
	httpx.OK(w, s.revenue.IncomeAccount(userID))
}

func (s *Server) incomeLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	s.recordBehavior(userID, "view_income_logs", "income", userID, map[string]interface{}{"status": status})
	httpx.OK(w, map[string]interface{}{"items": s.revenue.IncomeLogs(userID, status)})
}

func (s *Server) revenueRequest(w http.ResponseWriter, r *http.Request) (revenue.CalculateRequest, bool) {
	var req revenue.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid revenue record id")
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

func (s *Server) gameHasOpenReport(gameID int64) bool {
	for _, report := range s.reports.List() {
		if report.GameID == gameID && report.Status != "closed" {
			return true
		}
	}
	return false
}

func (s *Server) applyRevenuePreviewBlocks(preview revenue.Preview) revenue.Preview {
	if preview.GameID > 0 && s.gameHasOpenReport(preview.GameID) && !hasRevenueBlockReason(preview.BlockReasons, "disputed") {
		preview.BlockReasons = append(preview.BlockReasons, "disputed")
	}
	preview.CanGenerateRecord = len(preview.BlockReasons) == 0
	return preview
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
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid revenue template")
	case errors.Is(err, revenue.ErrTemplateNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "revenue template not found")
	case errors.Is(err, revenue.ErrReviewIncomplete):
		httpx.Error(w, http.StatusConflict, 40941, "review incomplete")
	case errors.Is(err, revenue.ErrDuplicateRecord):
		httpx.Error(w, http.StatusConflict, 40942, "duplicate revenue record")
	case errors.Is(err, revenue.ErrRecordFrozen):
		httpx.Error(w, http.StatusConflict, 40943, "revenue record frozen")
	case errors.Is(err, revenue.ErrRecordNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "revenue record not found")
	case errors.Is(err, revenue.ErrRecordNotSettleable):
		httpx.Error(w, http.StatusConflict, 40944, "revenue record not settleable")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "revenue operation failed")
	}
}
