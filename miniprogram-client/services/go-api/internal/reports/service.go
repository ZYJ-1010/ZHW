package reports

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/revenue"
)

var (
	ErrReportNotFound = errors.New("report not found")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidReport  = errors.New("invalid report")
)

type RevenueFreezer interface {
	FreezeByGame(gameID int64, reason string) (revenue.Record, bool, error)
}

type Repository interface {
	SaveReport(ctx context.Context, report Report) (Report, error)
	ListReports(ctx context.Context) ([]Report, error)
	ListReportsByUser(ctx context.Context, userID int64) ([]Report, error)
	FindReport(ctx context.Context, reportID int64) (Report, bool, error)
	UpdateReport(ctx context.Context, report Report) (Report, error)
}

type Report struct {
	ID                 int64     `json:"id"`
	GameID             int64     `json:"gameId"`
	ReporterUserID     int64     `json:"reporterUserId"`
	TargetUserID       int64     `json:"targetUserId,omitempty"`
	ReportType         string    `json:"reportType"`
	Content            string    `json:"content"`
	Status             string    `json:"status"`
	ChatMessageID      int64     `json:"chatMessageId,omitempty"`
	FileID             int64     `json:"fileId,omitempty"`
	ReviewID           int64     `json:"reviewId,omitempty"`
	CreditLogID        int64     `json:"creditLogId,omitempty"`
	RevenueRecordID    int64     `json:"revenueRecordId,omitempty"`
	RevenueFrozen      bool      `json:"revenueFrozen"`
	RevenueFreezeNote  string    `json:"revenueFreezeNote,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	HandledAt          string    `json:"handledAt,omitempty"`
	HandlerAdminID     int64     `json:"handlerAdminId,omitempty"`
	HandleResult       string    `json:"handleResult,omitempty"`
	HandleOutcome      string    `json:"handleOutcome,omitempty"`
	RewardPoints       int       `json:"rewardPoints,omitempty"`
	CreditChange       int       `json:"creditChange,omitempty"`
	CreditTargetUserID int64     `json:"creditTargetUserId,omitempty"`
	AppealFileIDs      []int64   `json:"appealFileIds,omitempty"`
}

type CreateRequest struct {
	GameID          int64  `json:"gameId"`
	TargetUserID    int64  `json:"targetUserId"`
	ReportType      string `json:"reportType"`
	Content         string `json:"content"`
	ChatMessageID   int64  `json:"chatMessageId"`
	FileID          int64  `json:"fileId"`
	ReviewID        int64  `json:"reviewId"`
	CreditLogID     int64  `json:"creditLogId"`
	RevenueRecordID int64  `json:"revenueRecordId"`
}

type HandleRequest struct {
	AdminID            int64  `json:"adminId"`
	Result             string `json:"result"`
	Outcome            string `json:"outcome"`
	RewardPoints       int    `json:"rewardPoints"`
	CreditDeduct       int    `json:"creditDeduct"`
	CreditTargetUserID int64  `json:"creditTargetUserId"`
}

type AppealRequest struct {
	Content string  `json:"content"`
	FileID  int64   `json:"fileId"`
	FileIDs []int64 `json:"fileIds"`
}

type CreditAppealRequest struct {
	GameID      int64   `json:"gameId"`
	CreditLogID int64   `json:"creditLogId"`
	Content     string  `json:"content"`
	FileID      int64   `json:"fileId"`
	FileIDs     []int64 `json:"fileIds"`
}

type AssignRequest struct {
	AdminID        int64 `json:"adminId"`
	HandlerAdminID int64 `json:"handlerAdminId"`
}

type Service struct {
	mu      sync.RWMutex
	nextID  int64
	reports map[int64]Report
	freezer RevenueFreezer
	repo    Repository
}

func NewService(freezer RevenueFreezer) *Service {
	return NewServiceWithRepository(freezer, nil)
}

func NewServiceWithRepository(freezer RevenueFreezer, repo Repository) *Service {
	return &Service{
		nextID:  1,
		reports: make(map[int64]Report),
		freezer: freezer,
		repo:    repo,
	}
}

func (s *Service) Create(userID int64, req CreateRequest) (Report, error) {
	req.ReportType = strings.TrimSpace(req.ReportType)
	req.Content = strings.TrimSpace(req.Content)
	if req.GameID <= 0 || !validReportType(req.ReportType) || req.Content == "" || len(req.Content) > 1000 {
		return Report{}, ErrInvalidReport
	}
	if req.TargetUserID < 0 || req.ChatMessageID < 0 || req.FileID < 0 || req.ReviewID < 0 || req.CreditLogID < 0 || req.RevenueRecordID < 0 {
		return Report{}, ErrInvalidReport
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report := Report{
		ID:              s.nextID,
		GameID:          req.GameID,
		ReporterUserID:  userID,
		TargetUserID:    req.TargetUserID,
		ReportType:      req.ReportType,
		Content:         req.Content,
		Status:          "pending",
		ChatMessageID:   req.ChatMessageID,
		FileID:          req.FileID,
		ReviewID:        req.ReviewID,
		CreditLogID:     req.CreditLogID,
		RevenueRecordID: req.RevenueRecordID,
		CreatedAt:       time.Now(),
	}
	s.nextID++
	if s.freezer != nil {
		if _, frozen, err := s.freezer.FreezeByGame(req.GameID, "report_created"); err != nil {
			return Report{}, err
		} else if frozen {
			report.RevenueFrozen = true
			report.RevenueFreezeNote = "report_created"
		}
	}
	if s.repo != nil {
		saved, err := s.repo.SaveReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		report = saved
	}
	s.reports[report.ID] = report
	return report, nil
}

func (s *Service) CreateCreditAppeal(userID int64, req CreditAppealRequest) (Report, error) {
	req.Content = strings.TrimSpace(req.Content)
	if req.CreditLogID <= 0 || req.Content == "" || len(req.Content) > 1000 || req.FileID < 0 || hasNegativeID(req.FileIDs) {
		return Report{}, ErrInvalidReport
	}
	req.FileIDs = normalizeAppealFileIDs(req.FileID, req.FileIDs)
	if len(req.FileIDs) > 9 {
		return Report{}, ErrInvalidReport
	}
	report := Report{
		ID:             s.nextID,
		GameID:         req.GameID,
		ReporterUserID: userID,
		TargetUserID:   userID,
		ReportType:     "credit_appeal",
		Content:        req.Content,
		Status:         "appealed",
		FileID:         firstAppealFileID(req.FileIDs),
		CreditLogID:    req.CreditLogID,
		HandleResult:   req.Content,
		HandleOutcome:  "processing",
		AppealFileIDs:  req.FileIDs,
		CreatedAt:      time.Now(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	if s.repo != nil {
		saved, err := s.repo.SaveReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		report = saved
		report.Status = "appealed"
		report.HandleResult = req.Content
		report.HandleOutcome = "processing"
		report.AppealFileIDs = req.FileIDs
		if updated, err := s.repo.UpdateReport(context.Background(), report); err == nil {
			report = updated
		}
	}
	s.reports[report.ID] = report
	return report, nil
}

func (s *Service) My(userID int64) []Report {
	if s.repo != nil {
		if items, err := s.repo.ListReportsByUser(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Report, 0)
	for _, report := range s.reports {
		if report.ReporterUserID == userID {
			result = append(result, report)
		}
	}
	return result
}

func (s *Service) Appeals(userID int64) []Report {
	items := s.List()
	result := make([]Report, 0)
	for _, report := range items {
		if report.TargetUserID == userID && report.Status == "appealed" {
			result = append(result, report)
		}
	}
	return result
}

func (s *Service) List() []Report {
	if s.repo != nil {
		if items, err := s.repo.ListReports(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Report, 0, len(s.reports))
	for _, report := range s.reports {
		result = append(result, report)
	}
	return result
}

func (s *Service) Get(reportID int64) (Report, error) {
	if s.repo != nil {
		report, ok, err := s.repo.FindReport(context.Background(), reportID)
		if err != nil {
			return Report{}, err
		}
		if !ok {
			return Report{}, ErrReportNotFound
		}
		return report, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	report, ok := s.reports[reportID]
	if !ok {
		return Report{}, ErrReportNotFound
	}
	return report, nil
}

func (s *Service) Assign(reportID int64, req AssignRequest) (Report, error) {
	if req.HandlerAdminID <= 0 {
		return Report{}, ErrInvalidReport
	}
	if s.repo != nil {
		report, err := s.Get(reportID)
		if err != nil {
			return Report{}, err
		}
		report.Status = "assigned"
		report.HandlerAdminID = req.HandlerAdminID
		report.HandleResult = ""
		report.HandledAt = ""
		updated, err := s.repo.UpdateReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		s.mu.Lock()
		s.reports[updated.ID] = updated
		s.mu.Unlock()
		return updated, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report, ok := s.reports[reportID]
	if !ok {
		return Report{}, ErrReportNotFound
	}
	report.Status = "assigned"
	report.HandlerAdminID = req.HandlerAdminID
	report.HandleResult = ""
	report.HandledAt = ""
	s.reports[reportID] = report
	return report, nil
}

func (s *Service) Appeal(userID int64, reportID int64, req AppealRequest) (Report, error) {
	req.Content = strings.TrimSpace(req.Content)
	if hasNegativeID(req.FileIDs) {
		return Report{}, ErrInvalidReport
	}
	req.FileIDs = normalizeAppealFileIDs(req.FileID, req.FileIDs)
	if req.Content == "" || len(req.Content) > 1000 || req.FileID < 0 || len(req.FileIDs) > 9 {
		return Report{}, ErrInvalidReport
	}
	if s.repo != nil {
		report, err := s.Get(reportID)
		if err != nil {
			return Report{}, err
		}
		if report.TargetUserID != userID {
			return Report{}, ErrForbidden
		}
		report.Status = "appealed"
		report.HandleResult = req.Content
		report.AppealFileIDs = req.FileIDs
		report.HandledAt = time.Now().Format(time.RFC3339)
		if len(req.FileIDs) > 0 {
			report.FileID = req.FileIDs[0]
		}
		updated, err := s.repo.UpdateReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		s.mu.Lock()
		s.reports[updated.ID] = updated
		s.mu.Unlock()
		return updated, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report, ok := s.reports[reportID]
	if !ok {
		return Report{}, ErrReportNotFound
	}
	if report.TargetUserID != userID {
		return Report{}, ErrForbidden
	}
	report.Status = "appealed"
	report.HandleResult = req.Content
	report.AppealFileIDs = req.FileIDs
	report.HandledAt = time.Now().Format(time.RFC3339)
	if len(req.FileIDs) > 0 {
		report.FileID = req.FileIDs[0]
	}
	s.reports[reportID] = report
	return report, nil
}

func (s *Service) WithdrawAppeal(userID int64, reportID int64) (Report, error) {
	if s.repo != nil {
		report, err := s.Get(reportID)
		if err != nil {
			return Report{}, err
		}
		if report.TargetUserID != userID {
			return Report{}, ErrForbidden
		}
		if report.Status != "appealed" {
			return Report{}, ErrInvalidReport
		}
		report.Status = "appeal_withdrawn"
		report.HandledAt = time.Now().Format(time.RFC3339)
		updated, err := s.repo.UpdateReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		s.mu.Lock()
		s.reports[updated.ID] = updated
		s.mu.Unlock()
		return updated, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report, ok := s.reports[reportID]
	if !ok {
		return Report{}, ErrReportNotFound
	}
	if report.TargetUserID != userID {
		return Report{}, ErrForbidden
	}
	if report.Status != "appealed" {
		return Report{}, ErrInvalidReport
	}
	report.Status = "appeal_withdrawn"
	report.HandledAt = time.Now().Format(time.RFC3339)
	s.reports[reportID] = report
	return report, nil
}

func normalizeAppealFileIDs(fileID int64, fileIDs []int64) []int64 {
	seen := map[int64]bool{}
	result := make([]int64, 0, len(fileIDs)+1)
	if fileID > 0 {
		seen[fileID] = true
		result = append(result, fileID)
	}
	for _, id := range fileIDs {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

func firstAppealFileID(fileIDs []int64) int64 {
	if len(fileIDs) == 0 {
		return 0
	}
	return fileIDs[0]
}

func hasNegativeID(ids []int64) bool {
	for _, id := range ids {
		if id < 0 {
			return true
		}
	}
	return false
}

func (s *Service) Handle(reportID int64, req HandleRequest) (Report, error) {
	req.Result = normalizeHandleResult(req.Result)
	req.Outcome = normalizeHandleOutcome(req.Outcome)
	if len(req.Result) > 500 || req.Outcome == "invalid" || req.RewardPoints < 0 || req.CreditDeduct < 0 {
		return Report{}, ErrInvalidReport
	}
	if s.repo != nil {
		report, err := s.Get(reportID)
		if err != nil {
			return Report{}, err
		}
		report.Status = "handled"
		report.HandlerAdminID = req.AdminID
		report.HandleResult = defaultResult(req.Result)
		applyHandleMeta(&report, req)
		report.HandledAt = time.Now().Format(time.RFC3339)
		updated, err := s.repo.UpdateReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		applyHandleMeta(&updated, req)
		s.mu.Lock()
		s.reports[updated.ID] = updated
		s.mu.Unlock()
		return updated, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report, ok := s.reports[reportID]
	if !ok {
		return Report{}, ErrReportNotFound
	}
	report.Status = "handled"
	report.HandlerAdminID = req.AdminID
	report.HandleResult = defaultResult(req.Result)
	applyHandleMeta(&report, req)
	report.HandledAt = time.Now().Format(time.RFC3339)
	s.reports[reportID] = report
	return report, nil
}

func (s *Service) Close(reportID int64, req HandleRequest) (Report, error) {
	req.Result = normalizeHandleResult(req.Result)
	if len(req.Result) > 500 {
		return Report{}, ErrInvalidReport
	}
	if s.repo != nil {
		report, err := s.Get(reportID)
		if err != nil {
			return Report{}, err
		}
		report.Status = "closed"
		report.HandlerAdminID = req.AdminID
		report.HandleResult = defaultResult(req.Result)
		report.HandledAt = time.Now().Format(time.RFC3339)
		updated, err := s.repo.UpdateReport(context.Background(), report)
		if err != nil {
			return Report{}, err
		}
		s.mu.Lock()
		s.reports[updated.ID] = updated
		s.mu.Unlock()
		return updated, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report, ok := s.reports[reportID]
	if !ok {
		return Report{}, ErrReportNotFound
	}
	report.Status = "closed"
	report.HandlerAdminID = req.AdminID
	report.HandleResult = defaultResult(req.Result)
	report.HandledAt = time.Now().Format(time.RFC3339)
	s.reports[reportID] = report
	return report, nil
}

func defaultResult(value string) string {
	if value == "" {
		return "processed"
	}
	return value
}

func applyHandleMeta(report *Report, req HandleRequest) {
	report.HandleOutcome = req.Outcome
	report.RewardPoints = req.RewardPoints
	report.CreditChange = -req.CreditDeduct
	if req.CreditTargetUserID > 0 {
		report.CreditTargetUserID = req.CreditTargetUserID
		return
	}
	switch req.Outcome {
	case "confirmed":
		report.CreditTargetUserID = report.TargetUserID
	case "malicious":
		report.CreditTargetUserID = report.ReporterUserID
	case "appeal_approved":
		report.CreditTargetUserID = report.TargetUserID
	}
}

func normalizeHandleResult(value string) string {
	return strings.TrimSpace(value)
}

func normalizeHandleOutcome(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "", "unconfirmed":
		return "unconfirmed"
	case "confirmed", "malicious", "appeal_approved", "appeal_rejected":
		return value
	default:
		return "invalid"
	}
}

func validReportType(value string) bool {
	switch value {
	case "service_dispute", "im_message", "review_dispute", "revenue_dispute", "credit_appeal", "user_complaint", "other":
		return true
	default:
		return false
	}
}
