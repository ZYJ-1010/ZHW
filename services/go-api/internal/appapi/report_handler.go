package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/reports"
	"zhw-mini/services/go-api/internal/reviews"
)

const reportCenterConfigKey = "report.center_config"

type reportTypeConfigDTO struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	ReportType string `json:"reportType"`
	Order      int    `json:"order"`
	Visible    bool   `json:"visible"`
}

type reportCenterConfigDTO struct {
	Types                        []reportTypeConfigDTO `json:"types"`
	DefaultType                  string                `json:"defaultType"`
	MaxEvidenceCount             int                   `json:"maxEvidenceCount"`
	AllowedUploadTypes           []string              `json:"allowedUploadTypes"`
	Tips                         []string              `json:"tips"`
	AppealReasons                []reportTypeConfigDTO `json:"appealReasons"`
	AppealPlaceholder            string                `json:"appealPlaceholder"`
	AppealUploadNote             string                `json:"appealUploadNote"`
	AppealFileMaxCount           int                   `json:"appealFileMaxCount"`
	AppealUploadFullText         string                `json:"appealUploadFullText"`
	AppealUploadSelectedTemplate string                `json:"appealUploadSelectedTemplate"`
	AppealReviewTitle            string                `json:"appealReviewTitle"`
	AppealReviewRules            []string              `json:"appealReviewRules"`
	Version                      string                `json:"version"`
}

type reportService interface {
	Create(userID int64, req reports.CreateRequest) (reports.Report, error)
	CreateCreditAppeal(userID int64, req reports.CreditAppealRequest) (reports.Report, error)
	My(userID int64) []reports.Report
	MyStrict(userID int64) ([]reports.Report, error)
	Appeals(userID int64) []reports.Report
	AppealsStrict(userID int64) ([]reports.Report, error)
	List() []reports.Report
	ListStrict() ([]reports.Report, error)
	Get(reportID int64) (reports.Report, error)
	Appeal(userID int64, reportID int64, req reports.AppealRequest) (reports.Report, error)
	WithdrawAppeal(userID int64, reportID int64) (reports.Report, error)
	Assign(reportID int64, req reports.AssignRequest) (reports.Report, error)
	Handle(reportID int64, req reports.HandleRequest) (reports.Report, error)
	Close(reportID int64, req reports.HandleRequest) (reports.Report, error)
}

func (s *Server) createReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req reports.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	game, err := s.games.Get(req.GameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !reportGameParticipant(s, game, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "只有本局参与者可以举报")
		return
	}
	if req.TargetUserID > 0 {
		if req.TargetUserID == userID {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "不能举报自己")
			return
		}
		if _, exists := s.auth.UserByID(req.TargetUserID); !exists {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "被举报用户不存在")
			return
		}
		if !reportGameParticipant(s, game, req.TargetUserID) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "被举报用户不是本局参与者")
			return
		}
		items, listErr := s.reports.MyStrict(userID)
		if listErr != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报记录失败，请稍后重试")
			return
		}
		for _, item := range items {
			if item.GameID == req.GameID && item.TargetUserID == req.TargetUserID && item.ReportType == req.ReportType && item.Status != "handled" && item.Status != "closed" {
				httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "相同举报正在处理中，请勿重复提交")
				return
			}
		}
	}
	if !s.validateReportEvidence(w, req) {
		return
	}
	report, err := s.reports.Create(userID, req)
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_report", "report", report.ID, map[string]interface{}{"gameId": report.GameID, "revenueFrozen": report.RevenueFrozen})
	_, _ = s.createCriticalNotification(w, "report_created", notifications.CreateRequest{
		UserID:      userID,
		NotifyType:  "report_created",
		Title:       "举报申诉已提交",
		Content:     "你的举报申诉已提交，后台将尽快处理。",
		BizType:     "report",
		BizID:       report.ID,
		NeedWechat:  true,
		WechatState: "pending",
		WechatData: map[string]string{
			"thing1": "举报申诉已提交",
			"thing2": report.Status,
		},
	})
	httpx.OK(w, report)
}

func reportGameParticipant(s *Server, game games.Game, userID int64) bool {
	return userID > 0 && (s.games.IsMember(game.ID, userID) || game.CreatorUserID == userID || game.MainGuideUserID == userID)
}

func (s *Server) createCreditAppeal(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req reports.CreditAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "申诉参数错误")
		return
	}
	if !s.validateAppealFiles(w, req.FileID, req.FileIDs) {
		return
	}
	log, found, readErr := s.creditLogForUser(userID, req.CreditLogID)
	if readErr != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取信用记录失败，请稍后重试")
		return
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "信用记录不存在")
		return
	}
	if log.ChangeValue >= 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "仅可对扣分记录提交信用申诉")
		return
	}
	req.GameID = log.GameID
	report, err := s.reports.CreateCreditAppeal(userID, req)
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_credit_appeal", "credit_log", log.ID, map[string]interface{}{"gameId": log.GameID, "reportId": report.ID})
	_, _ = s.createCriticalNotification(w, "credit_appeal_created", notifications.CreateRequest{
		UserID:     userID,
		NotifyType: "credit_appeal_submitted",
		Title:      "信用申诉已提交",
		Content:    "你的信用申诉已提交，平台将尽快复核。",
		BizType:    "report",
		BizID:      report.ID,
	})
	httpx.OK(w, mergeReportAppealFiles(report, req.FileID, req.FileIDs))
}

func mergeReportAppealFiles(report reports.Report, fileID int64, fileIDs []int64) reports.Report {
	if len(report.AppealFileIDs) > 0 {
		return report
	}
	ids := append([]int64{}, fileIDs...)
	if fileID > 0 {
		ids = append([]int64{fileID}, ids...)
	}
	if len(ids) == 0 {
		return report
	}
	seen := map[int64]bool{}
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	report.AppealFileIDs = result
	if report.FileID == 0 && len(result) > 0 {
		report.FileID = result[0]
	}
	return report
}

func (s *Server) creditLogForUser(userID int64, creditLogID int64) (reviews.CreditLog, bool, error) {
	trace, err := s.reviews.TraceByUserStrict(userID)
	if err != nil {
		return reviews.CreditLog{}, false, err
	}
	for _, log := range trace.CreditLogs {
		if log.ID == creditLogID {
			return log, true, nil
		}
	}
	return reviews.CreditLog{}, false, nil
}

func (s *Server) reportConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentReportCenterConfig())
}

func (s *Server) adminReportConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentReportCenterConfig()})
	case http.MethodPut:
		var req reportCenterConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "举报中心配置参数错误")
			return
		}
		config, err := normalizeReportCenterConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(reportCenterConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存举报中心配置失败")
			return
		}
		s.recordOperation(r, "system_config:update", "system_config", reportCenterConfigKey, map[string]interface{}{
			"version":          config.Version,
			"maxEvidenceCount": config.MaxEvidenceCount,
			"typeCount":        len(config.Types),
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentReportCenterConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) myReports(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.reports.MyStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报记录失败，请稍后重试")
		return
	}
	httpx.OK(w, pagedReports(items, r))
}

func (s *Server) myAppeals(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items, err := s.reports.AppealsStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取申诉记录失败，请稍后重试")
		return
	}
	httpx.OK(w, pagedReports(items, r))
}

func (s *Server) myReportDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	reportID, parseOK := idFromAdminPath(w, r.URL.Path, "/api/app/reports/", "")
	if !parseOK {
		return
	}
	report, err := s.reports.Get(reportID)
	if err != nil {
		writeReportError(w, err)
		return
	}
	if report.ReporterUserID != userID && report.TargetUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权查看举报详情")
		return
	}
	report.CanAppeal = report.TargetUserID == userID && report.ReportType != "credit_appeal" && report.Status != "appealed"
	httpx.OK(w, report)
}

func (s *Server) appealReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	reportID, parseOK := reportIDFromAppPath(w, r.URL.Path, "/appeal")
	if !parseOK {
		return
	}
	var req reports.AppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "申诉参数错误")
		return
	}
	if !s.validateAppealFiles(w, req.FileID, req.FileIDs) {
		return
	}
	report, err := s.reports.Appeal(userID, reportID, req)
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_report_appeal", "report", report.ID, map[string]interface{}{"gameId": report.GameID})
	_, _ = s.createCriticalNotification(w, "report_appeal_created", notifications.CreateRequest{
		UserID:      report.ReporterUserID,
		NotifyType:  "report_assigned",
		Title:       "举报申诉已提交",
		Content:     "被举报人已提交申诉，平台将重新核实处理。",
		BizType:     "report",
		BizID:       report.ID,
		NeedWechat:  true,
		WechatState: "pending",
	})
	httpx.OK(w, report)
}

func (s *Server) routeAppReportPost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/appeal/withdraw"):
		s.withdrawReportAppeal(w, r)
	case strings.HasSuffix(r.URL.Path, "/appeal"):
		s.appealReport(w, r)
	default:
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "举报操作不存在")
	}
}

func (s *Server) withdrawReportAppeal(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	reportID, parseOK := reportIDFromAppPath(w, r.URL.Path, "/appeal/withdraw")
	if !parseOK {
		return
	}
	report, err := s.reports.WithdrawAppeal(userID, reportID)
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordBehavior(userID, "withdraw_report_appeal", "report", report.ID, map[string]interface{}{"gameId": report.GameID})
	_, _ = s.createCriticalNotification(w, "report_appeal_withdrawn", notifications.CreateRequest{
		UserID:     report.ReporterUserID,
		NotifyType: "report_appeal_withdrawn",
		Title:      "举报申诉已撤回",
		Content:    "被举报人已撤回申诉，平台将按当前举报状态继续处理。",
		BizType:    "report",
		BizID:      report.ID,
	})
	httpx.OK(w, report)
}

func (s *Server) adminReports(w http.ResponseWriter, r *http.Request) {
	items, err := s.reports.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取举报列表失败，请稍后重试")
		return
	}
	httpx.OK(w, pagedReports(items, r))
}

func (s *Server) adminReportDetail(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportIDFromPath(w, r.URL.Path, "")
	if !ok {
		return
	}
	report, err := s.reports.Get(reportID)
	if err != nil {
		writeReportError(w, err)
		return
	}
	evidence := s.reportEvidence(report)
	s.recordOperation(r, "report:view_detail", "report", strconv.FormatInt(report.ID, 10), map[string]interface{}{"gameId": report.GameID})
	httpx.OK(w, map[string]interface{}{"report": report, "evidence": evidence})
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportIDFromPath(w, r.URL.Path, "/handle")
	if !ok {
		return
	}
	var req reports.HandleRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if reportHandleReasonRequired(req.Outcome) && strings.TrimSpace(req.Result) == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "驳回审核必须填写原因")
		return
	}
	req.AdminID = parseInt64Header(r, "X-Admin-ID")
	before, err := s.reports.Get(reportID)
	if err != nil {
		writeReportError(w, err)
		return
	}
	if before.Status == "handled" {
		httpx.OK(w, before)
		return
	}
	var restoredCredit *reviews.CreditLog
	if before.ReportType == "credit_appeal" && strings.TrimSpace(req.Outcome) == "appeal_approved" {
		amount := req.CreditDeduct
		if amount <= 0 {
			amount, err = s.creditAppealRestoreAmount(before.TargetUserID, before.CreditLogID)
			if err != nil {
				httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取原信用扣分记录失败，请稍后重试")
				return
			}
		}
		if amount <= 0 {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "原扣分记录无可恢复分值")
			return
		}
		credit, _, restoreErr := s.reviews.RestoreCreditForAppeal(before.TargetUserID, before.GameID, before.CreditLogID, before.ID, amount)
		if restoreErr != nil {
			writeCreditAppealRestoreError(w, restoreErr)
			return
		}
		restoredCredit = &credit
		req.CreditDeduct = credit.ChangeValue
		req.CreditTargetUserID = before.TargetUserID
	}
	prepared := before
	prepared.HandleOutcome = normalizedReportOutcome(req.Outcome)
	if restoredCredit != nil {
		if restoreErr := s.restoreReportRevenue(&prepared); restoreErr != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "举报关联收益恢复失败，请稍后重试")
			return
		}
		prepared.CreditChange = restoredCredit.ChangeValue
		prepared.CreditTargetUserID = prepared.TargetUserID
	} else if outcomeErr := s.applyReportHandleOutcome(&prepared, req); outcomeErr != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "举报处理结果落账失败，请稍后重试")
		return
	}
	creditAmount := -prepared.CreditChange
	if prepared.HandleOutcome == "appeal_approved" {
		creditAmount = prepared.CreditChange
	}
	report, err := s.reports.Handle(reportID, reports.HandleRequest{
		AdminID:            req.AdminID,
		Result:             req.Result,
		Outcome:            prepared.HandleOutcome,
		RewardPoints:       prepared.RewardPoints,
		CreditDeduct:       creditAmount,
		CreditTargetUserID: prepared.CreditTargetUserID,
	})
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordOperation(r, "report:handle", "report", strconv.FormatInt(reportID, 10), map[string]interface{}{"status": report.Status, "adminId": req.AdminID})
	_, _ = s.createCriticalNotification(w, "report_handled", notifications.CreateRequest{
		UserID:      report.ReporterUserID,
		NotifyType:  "report_handled",
		Title:       "举报申诉已处理",
		Content:     reportNotificationContent(report),
		BizType:     "report",
		BizID:       report.ID,
		NeedWechat:  true,
		WechatState: "pending",
		WechatData: map[string]string{
			"thing1": "举报申诉已处理",
			"thing2": report.Status,
		},
	})
	httpx.OK(w, report)
}

func (s *Server) applyReportHandleOutcome(report *reports.Report, req reports.HandleRequest) error {
	report.RewardPoints = 0
	report.CreditChange = 0
	report.CreditTargetUserID = 0
	report.HandleOutcome = normalizedReportOutcome(req.Outcome)
	if report.ReportType == "credit_appeal" {
		return s.applyCreditAppealHandleOutcome(report, req)
	}
	switch report.HandleOutcome {
	case "confirmed":
		if report.TargetUserID > 0 && req.CreditDeduct > 0 {
			credit, _, err := s.reviews.DeductCreditOnceStrict(report.TargetUserID, report.GameID, "report_confirmed", "report_confirmed:"+strconv.FormatInt(report.ID, 10))
			if err != nil {
				return err
			}
			report.CreditChange = credit.ChangeValue
			report.CreditTargetUserID = report.TargetUserID
		}
		if report.ReporterUserID > 0 && req.RewardPoints > 0 {
			_, log, _, err := s.points.GrantOnce(report.ReporterUserID, req.RewardPoints, "report_reward", report.ID, "举报核实奖励")
			if err != nil {
				return err
			}
			report.RewardPoints = log.ChangeValue
		}
	case "malicious":
		if report.ReporterUserID > 0 && req.CreditDeduct > 0 {
			credit, _, err := s.reviews.DeductCreditOnceStrict(report.ReporterUserID, report.GameID, "malicious_report", "malicious_report:"+strconv.FormatInt(report.ID, 10))
			if err != nil {
				return err
			}
			report.CreditChange = credit.ChangeValue
			report.CreditTargetUserID = report.ReporterUserID
		}
	case "appeal_approved":
		return s.restoreReportRevenue(report)
	}
	return nil
}

func (s *Server) applyCreditAppealHandleOutcome(report *reports.Report, _ reports.HandleRequest) error {
	if err := s.restoreReportRevenue(report); err != nil {
		return err
	}
	report.CreditTargetUserID = report.TargetUserID
	return nil
}

func writeCreditAppealRestoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, reviews.ErrCreditLogNotFound):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "原扣分记录不存在或不可申诉")
	case errors.Is(err, reviews.ErrCreditAppealConflict):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "该扣分记录已关联其他申诉")
	case errors.Is(err, reviews.ErrInvalidReview):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "信用恢复参数错误")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "恢复信用分失败，请稍后重试")
	}
}

func (s *Server) restoreReportRevenue(report *reports.Report) error {
	if report == nil || (!report.RevenueFrozen && report.RevenueRecordID <= 0) {
		return nil
	}
	restored, changed, err := s.revenue.RestoreFrozenByGame(report.GameID, "appeal_approved")
	if err != nil {
		return err
	}
	if changed {
		report.RevenueFrozen = restored.Status == "frozen"
		report.RevenueFreezeNote = "appeal_approved"
	}
	return nil
}

func normalizedReportOutcome(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unconfirmed"
	}
	return value
}

func (s *Server) creditAppealRestoreAmount(userID int64, creditLogID int64) (int, error) {
	trace, err := s.reviews.TraceByUserStrict(userID)
	if err != nil {
		return 0, err
	}
	for _, log := range trace.CreditLogs {
		if log.ID == creditLogID && log.ChangeValue < 0 {
			return -log.ChangeValue, nil
		}
	}
	return 0, nil
}

func (s *Server) assignReport(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportIDFromPath(w, r.URL.Path, "/assign")
	if !ok {
		return
	}
	var req reports.AssignRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	req.AdminID = parseInt64Header(r, "X-Admin-ID")
	if req.HandlerAdminID <= 0 {
		req.HandlerAdminID = req.AdminID
	}
	report, err := s.reports.Assign(reportID, req)
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordOperation(r, "report:assign", "report", strconv.FormatInt(reportID, 10), map[string]interface{}{"status": report.Status, "adminId": req.AdminID, "handlerAdminId": report.HandlerAdminID})
	_, _ = s.createCriticalNotification(w, "report_assigned", notifications.CreateRequest{
		UserID:      report.ReporterUserID,
		NotifyType:  "report_assigned",
		Title:       "举报申诉已分配处理人",
		Content:     "你的举报申诉已进入专人处理流程。",
		BizType:     "report",
		BizID:       report.ID,
		NeedWechat:  true,
		WechatState: "pending",
		WechatData: map[string]string{
			"thing1": "举报申诉已分配",
			"thing2": report.Status,
		},
	})
	httpx.OK(w, report)
}

func (s *Server) reportEvidence(report reports.Report) map[string]interface{} {
	evidence := map[string]interface{}{
		"ids": map[string]int64{
			"chatMessageId":   report.ChatMessageID,
			"fileId":          report.FileID,
			"reviewId":        report.ReviewID,
			"revenueRecordId": report.RevenueRecordID,
		},
	}
	if game, err := s.games.Get(report.GameID); err == nil {
		evidence["game"] = game
	}
	if messages, room, err := s.messagesForReport(report); err == nil {
		evidence["chatRoom"] = room
		evidence["chatMessages"] = messages
	}
	if report.FileID > 0 {
		if file, err := s.files.Get(report.FileID); err == nil {
			evidence["file"] = file
		}
	}
	if len(report.AppealFileIDs) > 0 {
		files := make([]interface{}, 0, len(report.AppealFileIDs))
		for _, fileID := range report.AppealFileIDs {
			if file, err := s.files.Get(fileID); err == nil {
				files = append(files, file)
			}
		}
		evidence["appealFileIds"] = report.AppealFileIDs
		evidence["appealFiles"] = files
	}
	if report.ReviewID > 0 {
		trace := s.reviews.TraceByGame(report.GameID)
		for _, review := range trace.Reviews {
			if review.ID == report.ReviewID {
				evidence["review"] = review
				break
			}
		}
	}
	if report.RevenueRecordID > 0 {
		if record, err := s.revenue.Record(report.RevenueRecordID); err == nil {
			evidence["revenueRecord"] = record
		}
	}
	return evidence
}

func (s *Server) validateAppealFiles(w http.ResponseWriter, fileID int64, fileIDs []int64) bool {
	for _, id := range append([]int64{fileID}, fileIDs...) {
		if id <= 0 {
			continue
		}
		file, err := s.files.Get(id)
		if err != nil || file.BizType != "report_attachment" {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "申诉附件无效")
			return false
		}
	}
	return true
}

func (s *Server) messagesForReport(report reports.Report) ([]interface{}, interface{}, error) {
	room := s.im.EnsureRoom(report.GameID)
	messages, room, err := s.im.AdminMessagesByRoom(room.ID)
	if err != nil {
		return nil, nil, err
	}
	result := make([]interface{}, 0)
	for _, message := range messages {
		result = append(result, message)
	}
	return result, room, nil
}

func (s *Server) validateReportEvidence(w http.ResponseWriter, req reports.CreateRequest) bool {
	if req.ChatMessageID > 0 {
		found := false
		for _, message := range s.im.AllMessages() {
			if message.ID == req.ChatMessageID && message.GameID == req.GameID {
				found = true
				break
			}
		}
		if !found {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "举报聊天证据无效")
			return false
		}
	}
	if req.FileID > 0 {
		file, err := s.files.Get(req.FileID)
		if err != nil || file.ObjectID != req.GameID {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "举报文件证据无效")
			return false
		}
	}
	if req.ReviewID > 0 {
		found := false
		trace := s.reviews.TraceByGame(req.GameID)
		for _, review := range trace.Reviews {
			if review.ID == req.ReviewID {
				found = true
				break
			}
		}
		if !found {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "举报评价证据无效")
			return false
		}
	}
	if req.RevenueRecordID > 0 {
		record, err := s.revenue.Record(req.RevenueRecordID)
		if err != nil || record.GameID != req.GameID {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "举报收益证据无效")
			return false
		}
	}
	return true
}

func (s *Server) closeReport(w http.ResponseWriter, r *http.Request) {
	reportID, ok := reportIDFromPath(w, r.URL.Path, "/close")
	if !ok {
		return
	}
	var req reports.HandleRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	req.AdminID = parseInt64Header(r, "X-Admin-ID")
	report, err := s.reports.Close(reportID, req)
	if err != nil {
		writeReportError(w, err)
		return
	}
	s.recordOperation(r, "report:close", "report", strconv.FormatInt(reportID, 10), map[string]interface{}{"status": report.Status, "adminId": req.AdminID})
	_, _ = s.createCriticalNotification(w, "report_closed", notifications.CreateRequest{
		UserID:      report.ReporterUserID,
		NotifyType:  "report_closed",
		Title:       "举报申诉已关闭",
		Content:     "你的举报申诉已关闭。",
		BizType:     "report",
		BizID:       report.ID,
		NeedWechat:  true,
		WechatState: "pending",
		WechatData: map[string]string{
			"thing1": "举报申诉已关闭",
			"thing2": report.Status,
		},
	})
	httpx.OK(w, report)
}

func (s *Server) batchHandleReports(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ReportIDs      []int64 `json:"reportIds"`
		Action         string  `json:"action"`
		AdminID        int64   `json:"adminId"`
		HandlerAdminID int64   `json:"handlerAdminId"`
		Result         string  `json:"result"`
		Outcome        string  `json:"outcome"`
		RewardPoints   int     `json:"rewardPoints"`
		CreditDeduct   int     `json:"creditDeduct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "批量举报处理请求参数错误")
		return
	}
	permission, ok := reportBatchActionPermission(req.Action)
	if !ok {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "批量举报处理动作无效")
		return
	}
	adminID, ok := s.requireAdminPermissionID(w, r, permission)
	if !ok {
		return
	}
	req.AdminID = adminID
	if req.Action == "handle" && reportHandleReasonRequired(req.Outcome) && strings.TrimSpace(req.Result) == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "驳回审核必须填写原因")
		return
	}
	if req.HandlerAdminID <= 0 {
		req.HandlerAdminID = adminID
	}
	reportIDs := uniquePositiveIDs(req.ReportIDs)
	if len(reportIDs) == 0 || len(reportIDs) > 100 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "举报编号列表错误")
		return
	}
	results := make([]batchMutationResult, 0, len(reportIDs))
	success := 0
	for _, reportID := range reportIDs {
		report, err := s.applyReportBatchAction(reportID, req.Action, req.AdminID, req.HandlerAdminID, req.Result, req.Outcome, req.RewardPoints, req.CreditDeduct)
		result := batchMutationResult{ID: reportID}
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Status = report.Status
			success++
			s.recordOperation(r, "report:batch_"+req.Action, "report", strconv.FormatInt(reportID, 10), map[string]interface{}{"status": report.Status, "adminId": req.AdminID})
			if req.Action == "handle" {
				_, _ = s.createCriticalNotification(w, "report_batch_handled", notifications.CreateRequest{UserID: report.ReporterUserID, NotifyType: "report_handled", Title: "举报申诉已处理", Content: reportNotificationContent(report), BizType: "report", BizID: report.ID})
			}
		}
		results = append(results, result)
	}
	httpx.OK(w, map[string]interface{}{
		"items":   results,
		"success": success,
		"failed":  len(results) - success,
		"total":   len(results),
	})
}

func reportHandleReasonRequired(outcome string) bool {
	switch strings.TrimSpace(outcome) {
	case "appeal_rejected", "malicious":
		return true
	default:
		return false
	}
}

func reportNotificationContent(report reports.Report) string {
	content := "你的举报申诉已有处理结果，请进入小程序查看。"
	if strings.TrimSpace(report.HandleResult) != "" {
		content += "处理说明：" + strings.TrimSpace(report.HandleResult)
	}
	return content
}

func reportBatchActionPermission(action string) (string, bool) {
	switch strings.TrimSpace(action) {
	case "assign":
		return "report:assign", true
	case "handle":
		return "report:handle", true
	case "close":
		return "report:close", true
	default:
		return "", false
	}
}

func pagedReports(items []reports.Report, r *http.Request) map[string]interface{} {
	page := queryInt(r.URL.Query().Get("page"))
	pageSize := queryInt(r.URL.Query().Get("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = len(items)
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return map[string]interface{}{
		"items":    items[start:end],
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"hasMore":  end < total,
	}
}

func (s *Server) applyReportBatchAction(reportID int64, action string, adminID int64, handlerAdminID int64, result string, outcome string, rewardPoints int, creditDeduct int) (reports.Report, error) {
	switch strings.TrimSpace(action) {
	case "assign":
		if handlerAdminID <= 0 {
			handlerAdminID = adminID
		}
		return s.reports.Assign(reportID, reports.AssignRequest{AdminID: adminID, HandlerAdminID: handlerAdminID})
	case "handle":
		before, err := s.reports.Get(reportID)
		if err != nil {
			return reports.Report{}, err
		}
		if before.Status == "handled" {
			return before, nil
		}
		req := reports.HandleRequest{AdminID: adminID, Result: result, Outcome: outcome, RewardPoints: rewardPoints, CreditDeduct: creditDeduct}
		prepared := before
		prepared.HandleOutcome = normalizedReportOutcome(outcome)
		if before.ReportType == "credit_appeal" && prepared.HandleOutcome == "appeal_approved" {
			amount := creditDeduct
			if amount <= 0 {
				amount, err = s.creditAppealRestoreAmount(before.TargetUserID, before.CreditLogID)
				if err != nil {
					return reports.Report{}, err
				}
			}
			if amount <= 0 {
				return reports.Report{}, reviews.ErrCreditLogNotFound
			}
			credit, _, restoreErr := s.reviews.RestoreCreditForAppeal(before.TargetUserID, before.GameID, before.CreditLogID, before.ID, amount)
			if restoreErr != nil {
				return reports.Report{}, restoreErr
			}
			if restoreErr = s.restoreReportRevenue(&prepared); restoreErr != nil {
				return reports.Report{}, restoreErr
			}
			prepared.CreditChange = credit.ChangeValue
			prepared.CreditTargetUserID = before.TargetUserID
		} else if outcomeErr := s.applyReportHandleOutcome(&prepared, req); outcomeErr != nil {
			return reports.Report{}, outcomeErr
		}
		creditAmount := -prepared.CreditChange
		if prepared.HandleOutcome == "appeal_approved" {
			creditAmount = prepared.CreditChange
		}
		return s.reports.Handle(reportID, reports.HandleRequest{AdminID: adminID, Result: result, Outcome: prepared.HandleOutcome, RewardPoints: prepared.RewardPoints, CreditDeduct: creditAmount, CreditTargetUserID: prepared.CreditTargetUserID})
	case "close":
		return s.reports.Close(reportID, reports.HandleRequest{AdminID: adminID, Result: result})
	default:
		return reports.Report{}, reports.ErrInvalidReport
	}
}

func reportIDFromPath(w http.ResponseWriter, path string, suffix string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/admin/reports/"), suffix)
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "举报 ID 错误")
		return 0, false
	}
	return id, true
}

func reportIDFromAppPath(w http.ResponseWriter, path string, suffix string) (int64, bool) {
	text := strings.TrimSuffix(strings.TrimPrefix(path, "/api/app/reports/"), suffix)
	id, err := strconv.ParseInt(strings.Trim(text, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "举报 ID 错误")
		return 0, false
	}
	return id, true
}

func writeReportError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, reports.ErrCreditAppealExists):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "该扣分记录已提交过申诉，请勿重复提交")
	case errors.Is(err, reports.ErrInvalidReport):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "举报参数错误")
	case errors.Is(err, reports.ErrReportNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "举报不存在")
	case errors.Is(err, reports.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权处理举报")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "举报操作失败")
	}
}

func (s *Server) currentReportCenterConfig() reportCenterConfigDTO {
	var stored reportCenterConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(reportCenterConfigKey, &stored) && len(stored.Types) > 0 {
		config, err := normalizeReportCenterConfig(stored)
		if err == nil {
			return config
		}
	}
	return defaultReportCenterConfig()
}

func defaultReportCenterConfig() reportCenterConfigDTO {
	return reportCenterConfigDTO{
		Types: []reportTypeConfigDTO{
			{Key: "private-guide", Label: "诱导私下交易", ReportType: "revenue_dispute", Order: 10, Visible: true},
			{Key: "private-done", Label: "私下交易已完成", ReportType: "revenue_dispute", Order: 20, Visible: true},
			{Key: "harassment", Label: "言语骚扰", ReportType: "user_complaint", Order: 30, Visible: true},
			{Key: "fake", Label: "虚假信息", ReportType: "user_complaint", Order: 40, Visible: true},
			{Key: "cancel", Label: "恶意取消", ReportType: "service_dispute", Order: 50, Visible: true},
			{Key: "other", Label: "其他违规", ReportType: "other", Order: 60, Visible: true},
		},
		DefaultType:        "private-guide",
		MaxEvidenceCount:   9,
		AllowedUploadTypes: []string{"jpg", "png", "pdf"},
		Tips: []string{
			"举报属实且能核实金额：罚款20% (50%奖励举报人)",
			"属实但无法核实：按后台处理规则发放奖励并记录信用变化",
			"不属实扣除举报人信用分2分，多次恶意举报封号",
		},
		AppealReasons: []reportTypeConfigDTO{
			{Key: "misjudge", Label: "误判扣分", Order: 10, Visible: true},
			{Key: "system", Label: "系统错误", Order: 20, Visible: true},
			{Key: "special", Label: "特殊情况", Order: 30, Visible: true},
			{Key: "other", Label: "其他", Order: 40, Visible: true},
		},
		AppealPlaceholder:            "请详细说明申诉原因，包括但不限于事件经过、时间、涉及人员等信息...",
		AppealUploadNote:             "支持 JPG、PNG 格式，单张不超过 5MB，最多 4 张证明材料",
		AppealFileMaxCount:           4,
		AppealUploadFullText:         "最多上传 4 张证明材料",
		AppealUploadSelectedTemplate: "已选择 {selected}/{max} 张证明材料",
		AppealReviewTitle:            "处理时效",
		AppealReviewRules: []string{
			"提交后24小时内初审",
			"复杂情况48小时内复核",
			"结果将通过站内消息通知",
		},
		Version: "2026-07-01",
	}
}

func normalizeReportCenterConfig(req reportCenterConfigDTO) (reportCenterConfigDTO, error) {
	req.DefaultType = strings.TrimSpace(req.DefaultType)
	req.Version = strings.TrimSpace(req.Version)
	req.AppealPlaceholder = strings.TrimSpace(req.AppealPlaceholder)
	req.AppealUploadNote = strings.TrimSpace(req.AppealUploadNote)
	req.AppealUploadFullText = strings.TrimSpace(req.AppealUploadFullText)
	req.AppealUploadSelectedTemplate = strings.TrimSpace(req.AppealUploadSelectedTemplate)
	req.AppealReviewTitle = strings.TrimSpace(req.AppealReviewTitle)
	req.AllowedUploadTypes = normalizeStringList(req.AllowedUploadTypes, 12)
	req.Tips = normalizeStringList(req.Tips, 8)
	req.AppealReviewRules = normalizeStringList(req.AppealReviewRules, 8)
	if req.MaxEvidenceCount <= 0 {
		req.MaxEvidenceCount = 9
	}
	if req.MaxEvidenceCount > 9 {
		return reportCenterConfigDTO{}, errors.New("maxEvidenceCount must be less than or equal to 9")
	}
	if req.AppealFileMaxCount <= 0 {
		req.AppealFileMaxCount = req.MaxEvidenceCount
	}
	if req.AppealFileMaxCount > req.MaxEvidenceCount {
		req.AppealFileMaxCount = req.MaxEvidenceCount
	}
	types := make([]reportTypeConfigDTO, 0, len(req.Types))
	seen := map[string]bool{}
	for _, item := range req.Types {
		item.Key = strings.TrimSpace(item.Key)
		item.Label = strings.TrimSpace(item.Label)
		item.ReportType = strings.TrimSpace(item.ReportType)
		if item.Key == "" || item.Label == "" || !validReportTypeForConfig(item.ReportType) || seen[item.Key] {
			continue
		}
		seen[item.Key] = true
		types = append(types, item)
	}
	if len(types) == 0 {
		return reportCenterConfigDTO{}, errors.New("report types required")
	}
	if req.DefaultType == "" || !seen[req.DefaultType] {
		req.DefaultType = types[0].Key
	}
	if len(req.AllowedUploadTypes) == 0 {
		req.AllowedUploadTypes = []string{"jpg", "png", "pdf"}
	}
	if len(req.Tips) == 0 {
		req.Tips = defaultReportCenterConfig().Tips
	}
	req.AppealReasons = normalizeAppealReasonConfig(req.AppealReasons)
	defaultConfig := defaultReportCenterConfig()
	if len(req.AppealReasons) == 0 {
		req.AppealReasons = defaultConfig.AppealReasons
	}
	if req.AppealPlaceholder == "" {
		req.AppealPlaceholder = defaultConfig.AppealPlaceholder
	}
	if req.AppealUploadNote == "" {
		req.AppealUploadNote = defaultConfig.AppealUploadNote
	}
	if req.AppealUploadFullText == "" {
		req.AppealUploadFullText = defaultConfig.AppealUploadFullText
	}
	if req.AppealUploadSelectedTemplate == "" {
		req.AppealUploadSelectedTemplate = defaultConfig.AppealUploadSelectedTemplate
	}
	if req.AppealReviewTitle == "" {
		req.AppealReviewTitle = defaultConfig.AppealReviewTitle
	}
	if len(req.AppealReviewRules) == 0 {
		req.AppealReviewRules = defaultConfig.AppealReviewRules
	}
	if req.Version == "" {
		req.Version = "2026-07-01"
	}
	req.Types = types
	return req, nil
}

func normalizeAppealReasonConfig(items []reportTypeConfigDTO) []reportTypeConfigDTO {
	reasons := make([]reportTypeConfigDTO, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		item.Key = strings.TrimSpace(item.Key)
		item.Label = strings.TrimSpace(item.Label)
		if item.Key == "" || item.Label == "" || seen[item.Key] {
			continue
		}
		seen[item.Key] = true
		item.ReportType = ""
		reasons = append(reasons, item)
	}
	return reasons
}

func validReportTypeForConfig(value string) bool {
	switch value {
	case "service_dispute", "im_message", "review_dispute", "revenue_dispute", "user_complaint", "other":
		return true
	default:
		return false
	}
}
