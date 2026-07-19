package revenue

import (
	"context"
	"testing"
	"time"
)

func TestServiceUsesRepositoryForRevenueFlow(t *testing.T) {
	repo := newFakeRevenueRepository()
	service := NewServiceWithRepository(fakeReviewChecker{complete: true}, repo)

	template, err := service.CreateTemplate(TemplateRequest{Name: "default", GameType: "free", PlatformBps: 1000, CreatorBps: 3000, MemberBps: 6000})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.savedTemplate || template.ID == 0 {
		t.Fatalf("expected template saved in repository, template=%+v repo=%+v", template, repo)
	}

	preview, err := service.Preview(CalculateRequest{GameID: 1, AmountCent: 10000, TemplateID: template.ID, CreatorID: 10, MemberIDs: []int64{10, 20}})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Items) != 3 {
		t.Fatalf("expected platform creator member preview, got %+v", preview)
	}

	record, err := service.Generate(CalculateRequest{GameID: 1, AmountCent: 10000, TemplateID: template.ID, CreatorID: 10, MemberIDs: []int64{10, 20}})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.savedRecord || record.ID == 0 || len(record.Items) != 3 {
		t.Fatalf("expected record saved in repository, record=%+v repo=%+v", record, repo)
	}
	if len(repo.incomeLogs) != 2 {
		t.Fatalf("expected income logs for payable record items, got %+v", repo.incomeLogs)
	}
	if _, err := service.Generate(CalculateRequest{GameID: 1, AmountCent: 10000, TemplateID: template.ID, CreatorID: 10, MemberIDs: []int64{10, 20}}); err != ErrDuplicateRecord {
		t.Fatalf("expected duplicate from repository, got %v", err)
	}

	frozen, err := service.Freeze(record.ID, "report")
	if err != nil {
		t.Fatal(err)
	}
	if frozen.Status != "frozen" || !repo.updatedRecord {
		t.Fatalf("expected frozen record from repository, got %+v repo=%+v", frozen, repo)
	}
	if len(repo.incomeLogs) != 4 {
		t.Fatalf("expected frozen income logs appended, got %+v", repo.incomeLogs)
	}
	if _, _, err := service.Settle(record.ID, "offline", "P001"); err != ErrRecordFrozen {
		t.Fatalf("expected frozen record block settlement, got %v", err)
	}
	restored, changed, err := service.RestoreFrozenByGame(record.GameID, "appeal_approved")
	if err != nil || !changed || restored.Status != "pending_settlement" {
		t.Fatalf("expected frozen record restored after appeal, changed=%v record=%+v err=%v", changed, restored, err)
	}

	record2, err := service.Generate(CalculateRequest{GameID: 2, AmountCent: 10000, TemplateID: template.ID, CreatorID: 10, MemberIDs: []int64{10, 20}})
	if err != nil {
		t.Fatal(err)
	}
	settled, settlement, err := service.Settle(record2.ID, "offline", "P002")
	if err != nil {
		t.Fatal(err)
	}
	if settled.Status != "settled" || settlement.ID == 0 || !repo.savedSettlement {
		t.Fatalf("expected settlement persisted, record=%+v settlement=%+v repo=%+v", settled, settlement, repo)
	}
	if len(repo.incomeLogs) != 10 {
		t.Fatalf("expected generated and settled income logs, got %+v", repo.incomeLogs)
	}
	if summary := service.IncomeSummary(20); summary.TotalCent == 0 || summary.SettledCent == 0 {
		t.Fatalf("expected income summary from repository, got %+v", summary)
	}
	if account := service.IncomeAccount(20); account.UserID != 20 || account.TotalCent == 0 || account.SettledCent == 0 {
		t.Fatalf("expected income account from repository, got %+v", account)
	}
	if logs := service.IncomeLogs(20, "settled"); len(logs) != 1 || logs[0].Status != "settled" {
		t.Fatalf("expected settled income logs from repository, got %+v", logs)
	}
}

func TestPreviewReturnsGenerateBlockReasons(t *testing.T) {
	service := NewService(fakeReviewChecker{complete: false})
	template, err := service.CreateTemplate(TemplateRequest{Name: "default", GameType: "free", PlatformBps: 1000, CreatorBps: 3000, MemberBps: 6000})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := service.Preview(CalculateRequest{GameID: 1, AmountCent: 10000, TemplateID: template.ID, CreatorID: 10, MemberIDs: []int64{10, 20}})
	if err != nil {
		t.Fatal(err)
	}
	if preview.CanGenerateRecord || !hasBlockReason(preview.BlockReasons, "review_incomplete") {
		t.Fatalf("expected review_incomplete preview block, got %+v", preview)
	}
	invalidAmount, err := service.Preview(CalculateRequest{GameID: 1, AmountCent: 0, TemplateID: template.ID})
	if err != nil {
		t.Fatal(err)
	}
	if invalidAmount.CanGenerateRecord || !hasBlockReason(invalidAmount.BlockReasons, "invalid_amount") {
		t.Fatalf("expected invalid_amount preview block, got %+v", invalidAmount)
	}
}

func TestPreviewAppliesRoleBpsRules(t *testing.T) {
	service := NewService(fakeReviewChecker{complete: true})
	template, err := service.CreateTemplate(TemplateRequest{Name: "default", GameType: "free", PlatformBps: 1000, CreatorBps: 3000, MemberBps: 6000})
	if err != nil {
		t.Fatal(err)
	}
	mustUpsertRule(t, service, RuleRequest{TemplateID: template.ID, RuleCode: "expert_bps", RuleValue: "500"})
	mustUpsertRule(t, service, RuleRequest{TemplateID: template.ID, RuleCode: "guide_bps", RuleValue: "700"})
	mustUpsertRule(t, service, RuleRequest{TemplateID: template.ID, RuleCode: "system_guide_bps", RuleValue: "300"})

	preview, err := service.Preview(CalculateRequest{
		GameID:        1,
		AmountCent:    10000,
		TemplateID:    template.ID,
		CreatorID:     10,
		MemberIDs:     []int64{10, 20, 30},
		ExpertID:      40,
		GuideID:       50,
		SystemGuideID: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !preview.CanGenerateRecord {
		t.Fatalf("expected preview generatable, got %+v", preview)
	}
	if amountForRole(preview.Items, "expert") != 500 || amountForRole(preview.Items, "guide") != 700 || amountForRole(preview.Items, "system_guide") != 300 {
		t.Fatalf("expected role rule items, got %+v", preview.Items)
	}
	if sumRoleAmount(preview.Items, "member") != 4500 {
		t.Fatalf("expected member pool after role rules to start with rounding remainder, got %+v", preview.Items)
	}
	if preview.RoundingDiffCent != 0 || sumPreviewItemAmounts(preview.Items) != preview.AmountCent {
		t.Fatalf("expected zero rounding diff and closed total, got %+v", preview)
	}
}

func TestPreviewAddsRoundingAdjustmentForUnallocatedPool(t *testing.T) {
	service := NewService(fakeReviewChecker{complete: true})
	template, err := service.CreateTemplate(TemplateRequest{Name: "default", GameType: "free", PlatformBps: 1000, CreatorBps: 3000, MemberBps: 6000})
	if err != nil {
		t.Fatal(err)
	}

	preview, err := service.Preview(CalculateRequest{
		GameID:     1,
		AmountCent: 9999,
		TemplateID: template.ID,
		CreatorID:  10,
		MemberIDs:  []int64{10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.RoundingDiffCent != 6001 {
		t.Fatalf("expected unallocated member pool as rounding diff, got %+v", preview)
	}
	if amountForRole(preview.Items, "rounding_adjustment") != preview.RoundingDiffCent {
		t.Fatalf("expected rounding adjustment item, got %+v", preview.Items)
	}
	if sumPreviewItemAmounts(preview.Items) != preview.AmountCent {
		t.Fatalf("expected item amounts to equal amountCent, got %+v", preview)
	}
}

func mustUpsertRule(t *testing.T, service *Service, req RuleRequest) {
	t.Helper()
	if _, err := service.UpsertRule(req); err != nil {
		t.Fatal(err)
	}
}

func amountForRole(items []Item, role string) int64 {
	for _, item := range items {
		if item.Role == role {
			return item.AmountCent
		}
	}
	return 0
}

func sumRoleAmount(items []Item, role string) int64 {
	var total int64
	for _, item := range items {
		if item.Role == role {
			total += item.AmountCent
		}
	}
	return total
}

func sumPreviewItemAmounts(items []Item) int64 {
	var total int64
	for _, item := range items {
		total += item.AmountCent
	}
	return total
}

func hasBlockReason(items []string, reason string) bool {
	for _, item := range items {
		if item == reason {
			return true
		}
	}
	return false
}

type fakeReviewChecker struct {
	complete bool
}

func (f fakeReviewChecker) GameReviewComplete(gameID int64) bool {
	return f.complete
}

type fakeRevenueRepository struct {
	nextTemplateID  int64
	nextRecordID    int64
	nextSettlement  int64
	templates       map[int64]Template
	rules           map[int64][]Rule
	records         map[int64]Record
	settlements     []Settlement
	incomeLogs      []IncomeLogEntry
	savedTemplate   bool
	savedRecord     bool
	updatedRecord   bool
	savedSettlement bool
}

func newFakeRevenueRepository() *fakeRevenueRepository {
	return &fakeRevenueRepository{
		nextTemplateID: 1,
		nextRecordID:   1,
		nextSettlement: 1,
		templates:      make(map[int64]Template),
		rules:          make(map[int64][]Rule),
		records:        make(map[int64]Record),
		settlements:    make([]Settlement, 0),
		incomeLogs:     make([]IncomeLogEntry, 0),
	}
}

func (r *fakeRevenueRepository) SaveTemplate(ctx context.Context, template Template) (Template, error) {
	r.savedTemplate = true
	template.ID = r.nextTemplateID
	r.nextTemplateID++
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	r.templates[template.ID] = template
	return template, nil
}

func (r *fakeRevenueRepository) ListTemplates(ctx context.Context) ([]Template, error) {
	result := make([]Template, 0, len(r.templates))
	for _, item := range r.templates {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeRevenueRepository) FindTemplate(ctx context.Context, templateID int64) (Template, bool, error) {
	item, ok := r.templates[templateID]
	return item, ok, nil
}

func (r *fakeRevenueRepository) SaveRule(ctx context.Context, rule Rule) (Rule, error) {
	rule.ID = int64(len(r.rules[rule.TemplateID]) + 1)
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	items := r.rules[rule.TemplateID]
	replaced := false
	for i := range items {
		if items[i].RuleCode == rule.RuleCode {
			items[i] = rule
			replaced = true
			break
		}
	}
	if !replaced {
		items = append(items, rule)
	}
	r.rules[rule.TemplateID] = items
	return rule, nil
}

func (r *fakeRevenueRepository) ListRules(ctx context.Context, templateID int64) ([]Rule, error) {
	return append([]Rule(nil), r.rules[templateID]...), nil
}

func (r *fakeRevenueRepository) SaveRecord(ctx context.Context, record Record) (Record, error) {
	r.savedRecord = true
	record.ID = r.nextRecordID
	record.RecordNo = "REV-FAKE"
	r.nextRecordID++
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	r.records[record.ID] = record
	return record, nil
}

func (r *fakeRevenueRepository) ListRecords(ctx context.Context) ([]Record, error) {
	result := make([]Record, 0, len(r.records))
	for _, item := range r.records {
		result = append(result, item)
	}
	return result, nil
}

func (r *fakeRevenueRepository) FindRecord(ctx context.Context, recordID int64) (Record, bool, error) {
	item, ok := r.records[recordID]
	return item, ok, nil
}

func (r *fakeRevenueRepository) FindRecordByGame(ctx context.Context, gameID int64) (Record, bool, error) {
	for _, item := range r.records {
		if item.GameID == gameID {
			return item, true, nil
		}
	}
	return Record{}, false, nil
}

func (r *fakeRevenueRepository) UpdateRecord(ctx context.Context, record Record) (Record, error) {
	r.updatedRecord = true
	r.records[record.ID] = record
	return record, nil
}

func (r *fakeRevenueRepository) SaveSettlement(ctx context.Context, settlement Settlement) (Settlement, error) {
	r.savedSettlement = true
	settlement.ID = r.nextSettlement
	r.nextSettlement++
	r.settlements = append(r.settlements, settlement)
	return settlement, nil
}

func (r *fakeRevenueRepository) ListSettlements(ctx context.Context) ([]Settlement, error) {
	return append([]Settlement(nil), r.settlements...), nil
}

func (r *fakeRevenueRepository) UpsertIncomeAccount(ctx context.Context, account IncomeAccount) (IncomeAccount, error) {
	account.ID = account.UserID
	if account.UpdatedAt.IsZero() {
		account.UpdatedAt = time.Now()
	}
	return account, nil
}

func (r *fakeRevenueRepository) AppendIncomeLog(ctx context.Context, log IncomeLogEntry) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	r.incomeLogs = append(r.incomeLogs, log)
	return nil
}

func (r *fakeRevenueRepository) IncomeSummary(ctx context.Context, userID int64) (IncomeSummary, error) {
	summary := IncomeSummary{UserID: userID}
	for _, record := range r.records {
		for _, item := range record.Items {
			if item.UserID != userID {
				continue
			}
			summary.TotalCent += item.AmountCent
			if record.Status == "settled" {
				summary.SettledCent += item.AmountCent
			} else {
				summary.PendingCent += item.AmountCent
			}
		}
	}
	return summary, nil
}

func (r *fakeRevenueRepository) IncomeLogs(ctx context.Context, userID int64, status string) ([]IncomeLog, error) {
	result := make([]IncomeLog, 0)
	for _, record := range r.records {
		if status != "" && record.Status != status {
			continue
		}
		for _, item := range record.Items {
			if item.UserID != userID {
				continue
			}
			result = append(result, IncomeLog{RecordID: record.ID, RecordNo: record.RecordNo, GameID: record.GameID, Role: item.Role, AmountCent: item.AmountCent, Status: record.Status, CreatedAt: record.CreatedAt, SettledAt: record.SettledAt})
		}
	}
	return result, nil
}
