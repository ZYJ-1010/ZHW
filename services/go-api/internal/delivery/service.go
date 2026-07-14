package delivery

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidDocument = errors.New("invalid delivery document")
	ErrInvalidCase     = errors.New("invalid test case")
	ErrInvalidRun      = errors.New("invalid test run")
	ErrCaseNotFound    = errors.New("test case not found")
)

type Document struct {
	ID        int64     `json:"id"`
	DocType   string    `json:"docType"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type TestCase struct {
	ID             int64     `json:"id"`
	Module         string    `json:"module"`
	CaseName       string    `json:"caseName"`
	Priority       string    `json:"priority"`
	ExpectedResult string    `json:"expectedResult"`
	CreatedAt      time.Time `json:"createdAt"`
}

type TestRun struct {
	ID             int64     `json:"id"`
	CaseID         int64     `json:"caseId"`
	Result         string    `json:"result"`
	ActualResult   string    `json:"actualResult,omitempty"`
	RequestID      string    `json:"requestId,omitempty"`
	EvidenceFileID int64     `json:"evidenceFileId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type DocumentRequest struct {
	DocType string `json:"docType"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type TestCaseRequest struct {
	Module         string `json:"module"`
	CaseName       string `json:"caseName"`
	Priority       string `json:"priority"`
	ExpectedResult string `json:"expectedResult"`
}

type TestRunRequest struct {
	CaseID         int64  `json:"caseId"`
	Result         string `json:"result"`
	ActualResult   string `json:"actualResult"`
	RequestID      string `json:"requestId"`
	EvidenceFileID int64  `json:"evidenceFileId"`
}

type Service struct {
	mu             sync.RWMutex
	nextDocumentID int64
	nextCaseID     int64
	nextRunID      int64
	documents      map[int64]Document
	cases          map[int64]TestCase
	runs           []TestRun
}

func NewService() *Service {
	return &Service{
		nextDocumentID: 1,
		nextCaseID:     1,
		nextRunID:      1,
		documents:      make(map[int64]Document),
		cases:          make(map[int64]TestCase),
		runs:           make([]TestRun, 0),
	}
}

func (s *Service) Documents() []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Document, 0, len(s.documents))
	for _, item := range s.documents {
		result = append(result, item)
	}
	return result
}

func (s *Service) CreateDocument(req DocumentRequest) (Document, error) {
	req.DocType = strings.TrimSpace(req.DocType)
	req.Title = strings.TrimSpace(req.Title)
	req.Status = strings.TrimSpace(req.Status)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.DocType == "" || req.Title == "" {
		return Document{}, ErrInvalidDocument
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if !validDocumentType(req.DocType) || !validDocumentStatus(req.Status) {
		return Document{}, ErrInvalidDocument
	}
	if req.Status != "draft" && req.Reason == "" {
		return Document{}, ErrInvalidDocument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := Document{
		ID:        s.nextDocumentID,
		DocType:   req.DocType,
		Title:     req.Title,
		Status:    req.Status,
		Reason:    req.Reason,
		CreatedAt: time.Now(),
	}
	s.nextDocumentID++
	s.documents[item.ID] = item
	return item, nil
}

func (s *Service) TestCases() []TestCase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]TestCase, 0, len(s.cases))
	for _, item := range s.cases {
		result = append(result, item)
	}
	return result
}

func (s *Service) CreateTestCase(req TestCaseRequest) (TestCase, error) {
	req.Module = strings.TrimSpace(req.Module)
	req.CaseName = strings.TrimSpace(req.CaseName)
	req.Priority = strings.TrimSpace(req.Priority)
	req.ExpectedResult = strings.TrimSpace(req.ExpectedResult)
	if req.Module == "" || req.CaseName == "" || req.ExpectedResult == "" {
		return TestCase{}, ErrInvalidCase
	}
	if req.Priority == "" {
		req.Priority = "P1"
	}
	if !validCasePriority(req.Priority) {
		return TestCase{}, ErrInvalidCase
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := TestCase{
		ID:             s.nextCaseID,
		Module:         req.Module,
		CaseName:       req.CaseName,
		Priority:       req.Priority,
		ExpectedResult: req.ExpectedResult,
		CreatedAt:      time.Now(),
	}
	s.nextCaseID++
	s.cases[item.ID] = item
	return item, nil
}

func (s *Service) TestRuns() []TestRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]TestRun(nil), s.runs...)
}

func (s *Service) CreateTestRun(req TestRunRequest) (TestRun, error) {
	req.RequestID = strings.TrimSpace(req.RequestID)
	if req.CaseID <= 0 || req.Result == "" {
		return TestRun{}, ErrInvalidRun
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cases[req.CaseID]; !ok {
		return TestRun{}, ErrCaseNotFound
	}
	if !validRunResult(req.Result) {
		return TestRun{}, ErrInvalidRun
	}
	if req.RequestID == "" && req.EvidenceFileID <= 0 {
		return TestRun{}, ErrInvalidRun
	}
	item := TestRun{
		ID:             s.nextRunID,
		CaseID:         req.CaseID,
		Result:         req.Result,
		ActualResult:   req.ActualResult,
		RequestID:      req.RequestID,
		EvidenceFileID: req.EvidenceFileID,
		CreatedAt:      time.Now(),
	}
	s.nextRunID++
	s.runs = append(s.runs, item)
	return item, nil
}

func validDocumentStatus(status string) bool {
	return status == "draft" || status == "ready" || status == "archived"
}

func validDocumentType(docType string) bool {
	switch docType {
	case "prd", "openapi", "database", "deployment", "test-report", "release-record", "rollback-plan":
		return true
	default:
		return false
	}
}

func validCasePriority(priority string) bool {
	return priority == "P0" || priority == "P1" || priority == "P2"
}

func validRunResult(result string) bool {
	return result == "passed" || result == "failed" || result == "blocked"
}
