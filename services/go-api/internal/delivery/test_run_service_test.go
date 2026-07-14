package delivery

import "testing"

func TestCreateTestCaseAndRun(t *testing.T) {
	service := NewService()

	testCase, err := service.CreateTestCase(TestCaseRequest{
		Module:         "E7",
		CaseName:       "delivery test run",
		Priority:       "P0",
		ExpectedResult: "passed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if testCase.ID != 1 || len(service.TestCases()) != 1 {
		t.Fatalf("expected test case registered, got %+v", testCase)
	}

	run, err := service.CreateTestRun(TestRunRequest{
		CaseID:       testCase.ID,
		Result:       "passed",
		ActualResult: "ok",
		RequestID:    "req-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.ID != 1 || run.CaseID != testCase.ID || len(service.TestRuns()) != 1 {
		t.Fatalf("expected test run registered, got %+v", run)
	}

	evidenceRun, err := service.CreateTestRun(TestRunRequest{
		CaseID:         testCase.ID,
		Result:         "blocked",
		ActualResult:   "see evidence",
		EvidenceFileID: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if evidenceRun.ID != 2 || evidenceRun.EvidenceFileID != 12 || evidenceRun.RequestID != "" {
		t.Fatalf("expected test run with evidence file, got %+v", evidenceRun)
	}
}

func TestCreateDeliveryDocumentValidatesTypeAndReason(t *testing.T) {
	service := NewService()

	if _, err := service.CreateDocument(DocumentRequest{DocType: "handoff", Title: "bad", Status: "draft"}); err != ErrInvalidDocument {
		t.Fatalf("expected ErrInvalidDocument for invalid type, got %v", err)
	}
	if _, err := service.CreateDocument(DocumentRequest{DocType: "openapi", Title: "OpenAPI", Status: "ready"}); err != ErrInvalidDocument {
		t.Fatalf("expected ErrInvalidDocument when ready reason is missing, got %v", err)
	}
	doc, err := service.CreateDocument(DocumentRequest{DocType: "openapi", Title: "OpenAPI", Status: "ready", Reason: "reviewed"})
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID != 1 || doc.Status != "ready" || doc.Reason != "reviewed" {
		t.Fatalf("expected delivery document with reason, got %+v", doc)
	}
}

func TestCreateTestRunRequiresKnownCaseAndValidResult(t *testing.T) {
	service := NewService()

	if _, err := service.CreateTestRun(TestRunRequest{CaseID: 1, Result: "passed"}); err != ErrCaseNotFound {
		t.Fatalf("expected ErrCaseNotFound, got %v", err)
	}
	if _, err := service.CreateTestCase(TestCaseRequest{Module: "E7", CaseName: "bad priority", Priority: "P3", ExpectedResult: "blocked"}); err != ErrInvalidCase {
		t.Fatalf("expected ErrInvalidCase for invalid priority, got %v", err)
	}
	testCase, err := service.CreateTestCase(TestCaseRequest{Module: "E7", CaseName: "blocked case", ExpectedResult: "blocked"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreateTestRun(TestRunRequest{CaseID: testCase.ID, Result: "unknown"}); err != ErrInvalidRun {
		t.Fatalf("expected ErrInvalidRun, got %v", err)
	}
	if _, err = service.CreateTestRun(TestRunRequest{CaseID: testCase.ID, Result: "passed"}); err != ErrInvalidRun {
		t.Fatalf("expected ErrInvalidRun when request id and evidence file are missing, got %v", err)
	}
}
