package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/delivery"
)

func (s *Server) adminDeliveryDocuments(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.delivery.Documents()})
}

func (s *Server) createAdminDeliveryDocument(w http.ResponseWriter, r *http.Request) {
	var req delivery.DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.delivery.CreateDocument(req)
	if err != nil {
		writeDeliveryError(w, err)
		return
	}
	s.recordOperation(r, "delivery:document:create", "delivery_document", strconv.FormatInt(item.ID, 10), map[string]interface{}{"status": item.Status})
	httpx.OK(w, item)
}

func (s *Server) adminTestCases(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.delivery.TestCases()})
}

func (s *Server) createAdminTestCase(w http.ResponseWriter, r *http.Request) {
	var req delivery.TestCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.delivery.CreateTestCase(req)
	if err != nil {
		writeDeliveryError(w, err)
		return
	}
	s.recordOperation(r, "testcase:create", "test_case", strconv.FormatInt(item.ID, 10), map[string]interface{}{"module": item.Module})
	httpx.OK(w, item)
}

func (s *Server) adminTestRuns(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]interface{}{"items": s.delivery.TestRuns()})
}

func (s *Server) createAdminTestRun(w http.ResponseWriter, r *http.Request) {
	var req delivery.TestRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.delivery.CreateTestRun(req)
	if err != nil {
		writeDeliveryError(w, err)
		return
	}
	s.recordOperation(r, "testcase:run:create", "test_run", strconv.FormatInt(item.ID, 10), map[string]interface{}{"caseId": item.CaseID, "result": item.Result, "evidenceFileId": item.EvidenceFileID})
	httpx.OK(w, item)
}

func writeDeliveryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, delivery.ErrCaseNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "测试用例不存在")
	case errors.Is(err, delivery.ErrInvalidDocument), errors.Is(err, delivery.ErrInvalidCase), errors.Is(err, delivery.ErrInvalidRun):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "交付测试参数错误")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "交付测试操作失败")
	}
}
