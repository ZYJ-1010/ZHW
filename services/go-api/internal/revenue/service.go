package revenue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidTemplate     = errors.New("invalid revenue template")
	ErrTemplateNotFound    = errors.New("revenue template not found")
	ErrReviewIncomplete    = errors.New("review incomplete")
	ErrDuplicateRecord     = errors.New("duplicate revenue record")
	ErrRecordNotFound      = errors.New("revenue record not found")
	ErrRecordFrozen        = errors.New("revenue record frozen")
	ErrRecordNotSettleable = errors.New("revenue record not settleable")
)

type ReviewChecker interface {
	GameReviewComplete(gameID int64) bool
}

type Template struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	GameType    string    `json:"gameType"`
	PlatformBps int       `json:"platformBps"`
	CreatorBps  int       `json:"creatorBps"`
	MemberBps   int       `json:"memberBps"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Rule struct {
	ID         int64     `json:"id"`
	TemplateID int64     `json:"templateId"`
	RuleCode   string    `json:"ruleCode"`
	RuleValue  string    `json:"ruleValue"`
	CreatedAt  time.Time `json:"createdAt"`
}

type TemplateRequest struct {
	Name        string `json:"name"`
	GameType    string `json:"gameType"`
	PlatformBps int    `json:"platformBps"`
	CreatorBps  int    `json:"creatorBps"`
	MemberBps   int    `json:"memberBps"`
}

type RuleRequest struct {
	TemplateID int64  `json:"templateId"`
	RuleCode   string `json:"ruleCode"`
	RuleValue  string `json:"ruleValue"`
}

type CalculateRequest struct {
	GameID        int64   `json:"gameId"`
	AmountCent    int64   `json:"amountCent"`
	TemplateID    int64   `json:"templateId"`
	MemberIDs     []int64 `json:"memberIds"`
	CreatorID     int64   `json:"creatorUserId"`
	ExpertID      int64   `json:"expertUserId"`
	GuideID       int64   `json:"guideUserId"`
	SystemGuideID int64   `json:"systemGuideUserId"`
}

type Item struct {
	UserID     int64  `json:"userId,omitempty"`
	Role       string `json:"role"`
	AmountCent int64  `json:"amountCent"`
}

type Preview struct {
	GameID            int64    `json:"gameId"`
	TemplateID        int64    `json:"templateId"`
	AmountCent        int64    `json:"amountCent"`
	RoundingDiffCent  int64    `json:"roundingDiffCent"`
	Items             []Item   `json:"items"`
	Rules             []Rule   `json:"rules"`
	CanGenerateRecord bool     `json:"canGenerateRecord"`
	BlockReasons      []string `json:"blockReasons"`
}

type Record struct {
	ID           int64     `json:"id"`
	RecordNo     string    `json:"recordNo"`
	GameID       int64     `json:"gameId"`
	TemplateID   int64     `json:"templateId"`
	AmountCent   int64     `json:"amountCent"`
	Status       string    `json:"status"`
	Items        []Item    `json:"items"`
	FrozenReason string    `json:"frozenReason,omitempty"`
	SettledAt    string    `json:"settledAt,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Settlement struct {
	ID         int64     `json:"id"`
	RecordID   int64     `json:"recordId"`
	Method     string    `json:"method"`
	ProofNo    string    `json:"proofNo"`
	AmountCent int64     `json:"amountCent"`
	CreatedAt  time.Time `json:"createdAt"`
}

type IncomeSummary struct {
	UserID      int64 `json:"userId"`
	TotalCent   int64 `json:"totalCent"`
	PendingCent int64 `json:"pendingCent"`
	SettledCent int64 `json:"settledCent"`
}

type IncomeAccount struct {
	ID          int64     `json:"id,omitempty"`
	UserID      int64     `json:"userId"`
	TotalCent   int64     `json:"totalCent"`
	PendingCent int64     `json:"pendingCent"`
	SettledCent int64     `json:"settledCent"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type IncomeLogEntry struct {
	UserID          int64
	RevenueRecordID int64
	ChangeValueCent int64
	Reason          string
	CreatedAt       time.Time
}

type IncomeLog struct {
	RecordID   int64     `json:"recordId"`
	RecordNo   string    `json:"recordNo"`
	GameID     int64     `json:"gameId"`
	Role       string    `json:"role"`
	AmountCent int64     `json:"amountCent"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	SettledAt  string    `json:"settledAt,omitempty"`
}

type Repository interface {
	SaveTemplate(ctx context.Context, template Template) (Template, error)
	ListTemplates(ctx context.Context) ([]Template, error)
	FindTemplate(ctx context.Context, templateID int64) (Template, bool, error)
	SaveRule(ctx context.Context, rule Rule) (Rule, error)
	ListRules(ctx context.Context, templateID int64) ([]Rule, error)
	SaveRecord(ctx context.Context, record Record) (Record, error)
	SaveRecordWithIncome(ctx context.Context, record Record) (Record, error)
	ListRecords(ctx context.Context) ([]Record, error)
	FindRecord(ctx context.Context, recordID int64) (Record, bool, error)
	FindRecordByGame(ctx context.Context, gameID int64) (Record, bool, error)
	UpdateRecord(ctx context.Context, record Record) (Record, error)
	SaveSettlement(ctx context.Context, settlement Settlement) (Settlement, error)
	SettleRecord(ctx context.Context, recordID int64, settlement Settlement) (Record, Settlement, error)
	TransitionRecord(ctx context.Context, recordID int64, expectedStatus string, nextStatus string, frozenReason string, logReason string, changedAt time.Time) (Record, bool, error)
	ListSettlements(ctx context.Context) ([]Settlement, error)
	UpsertIncomeAccount(ctx context.Context, account IncomeAccount) (IncomeAccount, error)
	AppendIncomeLog(ctx context.Context, log IncomeLogEntry) error
	IncomeSummary(ctx context.Context, userID int64) (IncomeSummary, error)
	IncomeLogs(ctx context.Context, userID int64, status string) ([]IncomeLog, error)
}

type Service struct {
	mu               sync.RWMutex
	nextTemplateID   int64
	nextRecordID     int64
	nextSettlementID int64
	reviews          ReviewChecker
	templates        map[int64]Template
	records          map[int64]Record
	recordsByGame    map[int64]int64
	settlements      []Settlement
	accounts         map[int64]IncomeAccount
	rules            map[int64][]Rule
	repo             Repository
}

func NewService(reviews ReviewChecker) *Service {
	return NewServiceWithRepository(reviews, nil)
}

func NewServiceWithRepository(reviews ReviewChecker, repo Repository) *Service {
	return &Service{
		nextTemplateID:   1,
		nextRecordID:     1,
		nextSettlementID: 1,
		reviews:          reviews,
		templates:        make(map[int64]Template),
		records:          make(map[int64]Record),
		recordsByGame:    make(map[int64]int64),
		settlements:      make([]Settlement, 0),
		accounts:         make(map[int64]IncomeAccount),
		rules:            make(map[int64][]Rule),
		repo:             repo,
	}
}

func (s *Service) CreateTemplate(req TemplateRequest) (Template, error) {
	if req.Name == "" {
		req.Name = "default"
	}
	if req.GameType == "" {
		req.GameType = "free"
	}
	if req.PlatformBps < 0 || req.CreatorBps < 0 || req.MemberBps < 0 || req.PlatformBps+req.CreatorBps+req.MemberBps != 10000 {
		return Template{}, ErrInvalidTemplate
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	template := Template{
		ID:          s.nextTemplateID,
		Name:        req.Name,
		GameType:    req.GameType,
		PlatformBps: req.PlatformBps,
		CreatorBps:  req.CreatorBps,
		MemberBps:   req.MemberBps,
		Status:      "active",
		CreatedAt:   time.Now(),
	}
	s.nextTemplateID++
	if s.repo != nil {
		saved, err := s.repo.SaveTemplate(context.Background(), template)
		if err != nil {
			return Template{}, err
		}
		template = saved
	}
	s.templates[template.ID] = template
	return template, nil
}

func (s *Service) Templates() []Template {
	items, _ := s.TemplatesStrict()
	return items
}

func (s *Service) TemplatesStrict() ([]Template, error) {
	if s.repo != nil {
		return s.repo.ListTemplates(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Template, 0, len(s.templates))
	for _, item := range s.templates {
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) UpsertRule(req RuleRequest) (Rule, error) {
	if req.TemplateID <= 0 || req.RuleCode == "" || req.RuleValue == "" {
		return Rule{}, ErrInvalidTemplate
	}
	if _, err := s.Template(req.TemplateID); err != nil {
		return Rule{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rule := Rule{ID: int64(len(s.rules[req.TemplateID]) + 1), TemplateID: req.TemplateID, RuleCode: req.RuleCode, RuleValue: req.RuleValue, CreatedAt: time.Now()}
	if s.repo != nil {
		saved, err := s.repo.SaveRule(context.Background(), rule)
		if err != nil {
			return Rule{}, err
		}
		rule = saved
	}
	rules := s.rules[req.TemplateID]
	replaced := false
	for i := range rules {
		if rules[i].RuleCode == req.RuleCode {
			rules[i] = rule
			replaced = true
			break
		}
	}
	if !replaced {
		rules = append(rules, rule)
	}
	s.rules[req.TemplateID] = rules
	return rule, nil
}

func (s *Service) Rules(templateID int64) []Rule {
	items, _ := s.RulesStrict(templateID)
	return items
}

func (s *Service) RulesStrict(templateID int64) ([]Rule, error) {
	if s.repo != nil {
		return s.repo.ListRules(context.Background(), templateID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Rule(nil), s.rules[templateID]...), nil
}

func (s *Service) Template(templateID int64) (Template, error) {
	if s.repo != nil {
		template, found, err := s.repo.FindTemplate(context.Background(), templateID)
		if err != nil {
			return Template{}, err
		}
		if found {
			return template, nil
		}
		return Template{}, ErrTemplateNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	template, ok := s.templates[templateID]
	if ok {
		return template, nil
	}
	return Template{}, ErrTemplateNotFound
}

func (s *Service) Preview(req CalculateRequest) (Preview, error) {
	blockReasons := make([]string, 0)
	if req.GameID <= 0 {
		blockReasons = append(blockReasons, "game_missing")
	}
	if req.AmountCent <= 0 {
		blockReasons = append(blockReasons, "invalid_amount")
	}
	if req.GameID > 0 && !s.reviews.GameReviewComplete(req.GameID) {
		blockReasons = append(blockReasons, "review_incomplete")
	}
	template, err := s.Template(req.TemplateID)
	if err != nil || template.Status != "active" {
		if err != nil && !errors.Is(err, ErrTemplateNotFound) {
			return Preview{}, err
		}
		return Preview{}, ErrTemplateNotFound
	}
	if req.AmountCent < 0 {
		req.AmountCent = 0
	}
	rules, err := s.RulesStrict(req.TemplateID)
	if err != nil {
		return Preview{}, err
	}
	items := make([]Item, 0)
	platformAmount := bpsAmount(req.AmountCent, template.PlatformBps)
	creatorAmount := bpsAmount(req.AmountCent, template.CreatorBps)
	memberPool := req.AmountCent - platformAmount - creatorAmount
	items = append(items, Item{Role: "platform", AmountCent: platformAmount})
	if req.CreatorID > 0 {
		items = append(items, Item{UserID: req.CreatorID, Role: "creator", AmountCent: creatorAmount})
	}
	roleItems, roleTotal := roleItemsFromRules(req.AmountCent, rules, req)
	memberPool -= roleTotal
	if memberPool < 0 {
		blockReasons = append(blockReasons, "invalid_rule_bps")
		memberPool = 0
	}
	items = append(items, roleItems...)
	memberIDs := nonCreatorMembers(req.MemberIDs, req.CreatorID)
	if len(memberIDs) > 0 {
		each := memberPool / int64(len(memberIDs))
		remain := memberPool - each*int64(len(memberIDs))
		for index, memberID := range memberIDs {
			amount := each
			if index == 0 {
				amount += remain
			}
			items = append(items, Item{UserID: memberID, Role: "member", AmountCent: amount})
		}
	}
	roundingDiffCent := req.AmountCent - sumItemAmounts(items)
	if roundingDiffCent != 0 {
		items = append(items, Item{Role: "rounding_adjustment", AmountCent: roundingDiffCent})
	}
	return Preview{
		GameID:            req.GameID,
		TemplateID:        req.TemplateID,
		AmountCent:        req.AmountCent,
		RoundingDiffCent:  roundingDiffCent,
		Items:             items,
		Rules:             rules,
		CanGenerateRecord: len(blockReasons) == 0,
		BlockReasons:      blockReasons,
	}, nil
}

func (s *Service) Generate(req CalculateRequest) (Record, error) {
	if !s.reviews.GameReviewComplete(req.GameID) {
		return Record{}, ErrReviewIncomplete
	}
	if s.repo != nil {
		if record, found, err := s.repo.FindRecordByGame(context.Background(), req.GameID); err != nil {
			return Record{}, err
		} else if found {
			return record, ErrDuplicateRecord
		}
	}
	if s.repo == nil {
		s.mu.RLock()
		if recordID, ok := s.recordsByGame[req.GameID]; ok {
			record := s.records[recordID]
			s.mu.RUnlock()
			return record, ErrDuplicateRecord
		}
		s.mu.RUnlock()
	}
	preview, err := s.Preview(req)
	if err != nil {
		return Record{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := Record{
		ID:         s.nextRecordID,
		RecordNo:   fmt.Sprintf("REV-%d-%d", req.GameID, s.nextRecordID),
		GameID:     req.GameID,
		TemplateID: req.TemplateID,
		AmountCent: req.AmountCent,
		Status:     "pending_settlement",
		Items:      preview.Items,
		CreatedAt:  time.Now(),
	}
	s.nextRecordID++
	if s.repo != nil {
		saved, err := s.repo.SaveRecordWithIncome(context.Background(), record)
		if err != nil {
			return Record{}, err
		}
		record = saved
	}
	s.records[record.ID] = record
	s.recordsByGame[record.GameID] = record.ID
	if s.repo == nil {
		s.syncIncomeAccountsForItemsLocked(record.Items)
	}
	return record, nil
}

func (s *Service) Records() []Record {
	items, _ := s.RecordsStrict()
	return items
}

func (s *Service) RecordsStrict() ([]Record, error) {
	if s.repo != nil {
		return s.repo.ListRecords(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Record, 0, len(s.records))
	for _, item := range s.records {
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) Record(id int64) (Record, error) {
	if s.repo != nil {
		record, found, err := s.repo.FindRecord(context.Background(), id)
		if err != nil {
			return Record{}, err
		}
		if !found {
			return Record{}, ErrRecordNotFound
		}
		return record, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[id]
	if !ok {
		return Record{}, ErrRecordNotFound
	}
	return record, nil
}

func (s *Service) Freeze(recordID int64, reason string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var record Record
	var ok bool
	if s.repo != nil {
		saved, found, err := s.repo.FindRecord(context.Background(), recordID)
		if err != nil {
			return Record{}, err
		}
		if found {
			record = saved
			ok = true
		}
	} else {
		record, ok = s.records[recordID]
	}
	if !ok {
		return Record{}, ErrRecordNotFound
	}
	record.Status = "frozen"
	record.FrozenReason = reason
	if s.repo != nil {
		saved, _, err := s.repo.TransitionRecord(context.Background(), recordID, "", "frozen", reason, "record_frozen", time.Now())
		if err != nil {
			return Record{}, err
		}
		record = saved
	}
	s.records[recordID] = record
	if s.repo == nil {
		s.syncIncomeAccountsForItemsLocked(record.Items)
	}
	return record, nil
}

func (s *Service) FreezeByGame(gameID int64, reason string) (Record, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.repo != nil {
		record, found, err := s.repo.FindRecordByGame(context.Background(), gameID)
		if err != nil || !found {
			return Record{}, false, err
		}
		saved, changed, err := s.repo.TransitionRecord(context.Background(), record.ID, "", "frozen", reason, "record_frozen", time.Now())
		if err != nil {
			return Record{}, false, err
		}
		s.records[saved.ID] = saved
		s.recordsByGame[gameID] = saved.ID
		return saved, changed, nil
	}
	recordID, ok := s.recordsByGame[gameID]
	if !ok {
		return Record{}, false, nil
	}
	record := s.records[recordID]
	if record.Status == "frozen" {
		return record, false, nil
	}
	record.Status = "frozen"
	record.FrozenReason = reason
	s.records[recordID] = record
	s.syncIncomeAccountsForItemsLocked(record.Items)
	return record, true, nil
}

// RestoreFrozenByGame reopens a frozen revenue record after an appeal is
// approved. It is idempotent and keeps the freeze/unfreeze trail in income
// logs for reconciliation.
func (s *Service) RestoreFrozenByGame(gameID int64, reason string) (Record, bool, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "appeal_approved"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var record Record
	var ok bool
	if s.repo != nil {
		saved, found, err := s.repo.FindRecordByGame(context.Background(), gameID)
		if err != nil || !found {
			return Record{}, false, err
		}
		record, ok = saved, true
	} else if recordID, found := s.recordsByGame[gameID]; found {
		record, ok = s.records[recordID]
	}
	if !ok {
		return Record{}, false, nil
	}
	if record.Status != "frozen" {
		return record, false, nil
	}
	record.Status = "pending_settlement"
	record.FrozenReason = ""
	if s.repo != nil {
		saved, changed, err := s.repo.TransitionRecord(context.Background(), record.ID, "frozen", "pending_settlement", "", "record_unfrozen:"+reason, time.Now())
		if err != nil {
			return Record{}, false, err
		}
		record = saved
		if !changed {
			return record, false, nil
		}
	}
	s.records[record.ID] = record
	s.recordsByGame[gameID] = record.ID
	if s.repo == nil {
		s.syncIncomeAccountsForItemsLocked(record.Items)
	}
	return record, true, nil
}

func (s *Service) HasFrozenRecordForGame(gameID int64) bool {
	if s.repo != nil {
		record, found, err := s.repo.FindRecordByGame(context.Background(), gameID)
		return err == nil && found && record.Status == "frozen"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	recordID, ok := s.recordsByGame[gameID]
	if !ok {
		return false
	}
	return s.records[recordID].Status == "frozen"
}

func (s *Service) Settle(recordID int64, method string, proofNo string) (Record, Settlement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var record Record
	var ok bool
	if s.repo != nil {
		saved, found, err := s.repo.FindRecord(context.Background(), recordID)
		if err != nil {
			return Record{}, Settlement{}, err
		}
		if found {
			record = saved
			ok = true
		}
	} else {
		record, ok = s.records[recordID]
	}
	if !ok {
		return Record{}, Settlement{}, ErrRecordNotFound
	}
	if record.Status == "frozen" {
		return Record{}, Settlement{}, ErrRecordFrozen
	}
	if record.Status != "pending_settlement" {
		return Record{}, Settlement{}, ErrRecordNotSettleable
	}
	settlement := Settlement{
		ID:         s.nextSettlementID,
		RecordID:   recordID,
		Method:     defaultMethod(method),
		ProofNo:    proofNo,
		AmountCent: record.AmountCent,
		CreatedAt:  time.Now(),
	}
	record.Status = "settled"
	record.SettledAt = settlement.CreatedAt.Format(time.RFC3339)
	if s.repo != nil {
		savedRecord, savedSettlement, err := s.repo.SettleRecord(context.Background(), recordID, settlement)
		if err != nil {
			return Record{}, Settlement{}, err
		}
		record = savedRecord
		settlement = savedSettlement
	} else {
		s.nextSettlementID++
	}
	s.records[recordID] = record
	s.settlements = append(s.settlements, settlement)
	if s.repo == nil {
		s.syncIncomeAccountsForItemsLocked(record.Items)
	}
	return record, settlement, nil
}

func (s *Service) Settlements() []Settlement {
	items, _ := s.SettlementsStrict()
	return items
}

func (s *Service) SettlementsStrict() ([]Settlement, error) {
	if s.repo != nil {
		return s.repo.ListSettlements(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Settlement(nil), s.settlements...), nil
}

func (s *Service) IncomeSummary(userID int64) IncomeSummary {
	summary, _ := s.IncomeSummaryStrict(userID)
	return summary
}

func (s *Service) IncomeSummaryStrict(userID int64) (IncomeSummary, error) {
	if s.repo != nil {
		return s.repo.IncomeSummary(context.Background(), userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.incomeSummaryLocked(userID), nil
}

func (s *Service) IncomeAccount(userID int64) IncomeAccount {
	account, _ := s.IncomeAccountStrict(userID)
	return account
}

func (s *Service) IncomeAccountStrict(userID int64) (IncomeAccount, error) {
	summary, err := s.IncomeSummaryStrict(userID)
	if err != nil {
		return IncomeAccount{}, err
	}
	if s.repo != nil {
		account, err := s.repo.UpsertIncomeAccount(context.Background(), accountFromSummary(summary))
		if err != nil {
			return IncomeAccount{}, err
		}
		s.mu.Lock()
		s.accounts[userID] = account
		s.mu.Unlock()
		return account, nil
	}
	return s.syncIncomeAccountFromSummary(summary), nil
}

func (s *Service) IncomeLogs(userID int64, status string) []IncomeLog {
	items, _ := s.IncomeLogsStrict(userID, status)
	return items
}

func (s *Service) IncomeLogsStrict(userID int64, status string) ([]IncomeLog, error) {
	if s.repo != nil {
		return s.repo.IncomeLogs(context.Background(), userID, status)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]IncomeLog, 0)
	for _, record := range s.records {
		if status != "" && record.Status != status {
			continue
		}
		for _, item := range record.Items {
			if item.UserID != userID {
				continue
			}
			result = append(result, IncomeLog{
				RecordID:   record.ID,
				RecordNo:   record.RecordNo,
				GameID:     record.GameID,
				Role:       item.Role,
				AmountCent: item.AmountCent,
				Status:     record.Status,
				CreatedAt:  record.CreatedAt,
				SettledAt:  record.SettledAt,
			})
		}
	}
	return result, nil
}

func (s *Service) incomeSummaryLocked(userID int64) IncomeSummary {
	summary := IncomeSummary{UserID: userID}
	for _, record := range s.records {
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
	return summary
}

func (s *Service) syncIncomeAccountsForItemsLocked(items []Item) {
	for _, userID := range affectedUserIDs(items) {
		if s.repo != nil {
			summary, err := s.repo.IncomeSummary(context.Background(), userID)
			if err != nil {
				continue
			}
			if account, err := s.repo.UpsertIncomeAccount(context.Background(), accountFromSummary(summary)); err == nil {
				s.accounts[userID] = account
			}
			continue
		}
		summary := s.incomeSummaryLocked(userID)
		account := accountFromSummary(summary)
		if existing, ok := s.accounts[userID]; ok {
			account.ID = existing.ID
		} else {
			account.ID = int64(len(s.accounts) + 1)
		}
		s.accounts[userID] = account
	}
}

func (s *Service) appendIncomeLogsForRecordLocked(record Record, reason string) {
	if s.repo == nil {
		return
	}
	for _, item := range record.Items {
		if item.UserID <= 0 {
			continue
		}
		_ = s.repo.AppendIncomeLog(context.Background(), IncomeLogEntry{
			UserID:          item.UserID,
			RevenueRecordID: record.ID,
			ChangeValueCent: item.AmountCent,
			Reason:          reason,
			CreatedAt:       time.Now(),
		})
	}
}

func (s *Service) syncIncomeAccountFromSummary(summary IncomeSummary) IncomeAccount {
	if s.repo != nil {
		if account, err := s.repo.UpsertIncomeAccount(context.Background(), accountFromSummary(summary)); err == nil {
			s.mu.Lock()
			s.accounts[summary.UserID] = account
			s.mu.Unlock()
			return account
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	account := accountFromSummary(summary)
	if existing, ok := s.accounts[summary.UserID]; ok {
		account.ID = existing.ID
	} else {
		account.ID = int64(len(s.accounts) + 1)
	}
	s.accounts[summary.UserID] = account
	return account
}

func affectedUserIDs(items []Item) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0)
	for _, item := range items {
		if item.UserID <= 0 {
			continue
		}
		if _, ok := seen[item.UserID]; ok {
			continue
		}
		seen[item.UserID] = struct{}{}
		result = append(result, item.UserID)
	}
	return result
}

func accountFromSummary(summary IncomeSummary) IncomeAccount {
	return IncomeAccount{
		UserID:      summary.UserID,
		TotalCent:   summary.TotalCent,
		PendingCent: summary.PendingCent,
		SettledCent: summary.SettledCent,
		UpdatedAt:   time.Now(),
	}
}

func bpsAmount(amountCent int64, bps int) int64 {
	return amountCent * int64(bps) / 10000
}

func roleItemsFromRules(amountCent int64, rules []Rule, req CalculateRequest) ([]Item, int64) {
	items := make([]Item, 0)
	var total int64
	for _, rule := range rules {
		bps := parseRuleBps(rule.RuleValue)
		if bps <= 0 {
			continue
		}
		amount := bpsAmount(amountCent, bps)
		switch rule.RuleCode {
		case "expert_bps":
			items = append(items, Item{UserID: req.ExpertID, Role: "expert", AmountCent: amount})
			total += amount
		case "guide_bps":
			items = append(items, Item{UserID: req.GuideID, Role: "guide", AmountCent: amount})
			total += amount
		case "system_guide_bps":
			items = append(items, Item{UserID: req.SystemGuideID, Role: "system_guide", AmountCent: amount})
			total += amount
		}
	}
	return items, total
}

func sumItemAmounts(items []Item) int64 {
	var total int64
	for _, item := range items {
		total += item.AmountCent
	}
	return total
}

func parseRuleBps(value string) int {
	var bps int
	_, _ = fmt.Sscanf(value, "%d", &bps)
	if bps < 0 {
		return 0
	}
	return bps
}

func nonCreatorMembers(memberIDs []int64, creatorID int64) []int64 {
	result := make([]int64, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		if memberID != creatorID {
			result = append(result, memberID)
		}
	}
	return result
}

func defaultMethod(method string) string {
	if method == "" {
		return "offline"
	}
	return method
}
