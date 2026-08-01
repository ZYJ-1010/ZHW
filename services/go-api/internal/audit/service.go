package audit

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"
)

type BehaviorLog struct {
	ID           int64           `json:"id"`
	UserID       int64           `json:"userId,omitempty"`
	EventType    string          `json:"eventType"`
	EventCode    string          `json:"eventCode"`
	TargetType   string          `json:"targetType"`
	BusinessType string          `json:"businessType"`
	TargetID     int64           `json:"targetId,omitempty"`
	BusinessID   int64           `json:"businessId,omitempty"`
	PagePath     string          `json:"pagePath,omitempty"`
	Keyword      string          `json:"keyword,omitempty"`
	Source       string          `json:"source"`
	Device       string          `json:"device,omitempty"`
	IP           string          `json:"ip,omitempty"`
	Extra        json.RawMessage `json:"extra,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	OccurredAt   time.Time       `json:"occurredAt"`
}

type OperationLog struct {
	ID          int64           `json:"id"`
	AdminUserID int64           `json:"adminUserId,omitempty"`
	Action      string          `json:"action"`
	TargetType  string          `json:"targetType,omitempty"`
	TargetID    string          `json:"targetId,omitempty"`
	RequestID   string          `json:"requestId,omitempty"`
	IP          string          `json:"ip,omitempty"`
	Detail      json.RawMessage `json:"detail,omitempty"`
	Before      json.RawMessage `json:"before,omitempty"`
	After       json.RawMessage `json:"after,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type BehaviorRequest struct {
	UserID       int64
	EventType    string
	EventCode    string
	TargetType   string
	BusinessType string
	TargetID     int64
	BusinessID   int64
	PagePath     string
	Keyword      string
	Source       string
	Device       string
	IP           string
	Extra        map[string]interface{}
	OccurredAt   time.Time
}

type BehaviorQuery struct {
	UserID    int64
	EventType string
	EventCode string
}

type FunnelSnapshot struct {
	Steps     []FunnelStep `json:"steps"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type FunnelStep struct {
	EventCode      string  `json:"eventCode"`
	UserCount      int     `json:"userCount"`
	ConversionRate float64 `json:"conversionRate"`
	DropOffRate    float64 `json:"dropOffRate"`
}

type RetentionSnapshot struct {
	Buckets   []RetentionBucket `json:"buckets"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type RetentionBucket struct {
	CohortDate    string  `json:"cohortDate"`
	NewUsers      int     `json:"newUsers"`
	Day1Retained  int     `json:"day1Retained"`
	Day7Retained  int     `json:"day7Retained"`
	Day30Retained int     `json:"day30Retained"`
	Day1Rate      float64 `json:"day1Rate"`
	Day7Rate      float64 `json:"day7Rate"`
	Day30Rate     float64 `json:"day30Rate"`
}

type DashboardSnapshot struct {
	Funnel    FunnelSnapshot    `json:"funnel"`
	Retention RetentionSnapshot `json:"retention"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type OperationRequest struct {
	AdminUserID int64
	Action      string
	TargetType  string
	TargetID    string
	RequestID   string
	IP          string
	Detail      map[string]interface{}
	Before      map[string]interface{}
	After       map[string]interface{}
}

type BehaviorRepository interface {
	SaveBehavior(ctx context.Context, log BehaviorLog) error
	ListBehavior(ctx context.Context) ([]BehaviorLog, error)
	QueryBehavior(ctx context.Context, query BehaviorQuery) ([]BehaviorLog, error)
}

type OperationRepository interface {
	SaveOperation(ctx context.Context, log OperationLog) error
	ListOperations(ctx context.Context) ([]OperationLog, error)
}

type Service struct {
	mu              sync.RWMutex
	nextBehaviorID  int64
	nextOperationID int64
	behaviorLogs    []BehaviorLog
	operationLogs   []OperationLog
	behaviorRepo    BehaviorRepository
	operationRepo   OperationRepository
}

func NewService() *Service {
	return NewServiceWithRepositories(nil, nil)
}

func NewServiceWithBehaviorRepository(repo BehaviorRepository) *Service {
	return NewServiceWithRepositories(repo, nil)
}

func NewServiceWithRepositories(behaviorRepo BehaviorRepository, operationRepo OperationRepository) *Service {
	return &Service{
		nextBehaviorID:  1,
		nextOperationID: 1,
		behaviorLogs:    make([]BehaviorLog, 0),
		operationLogs:   make([]OperationLog, 0),
		behaviorRepo:    behaviorRepo,
		operationRepo:   operationRepo,
	}
}

func (s *Service) RecordBehavior(req BehaviorRequest) BehaviorLog {
	log, _ := s.RecordBehaviorStrict(req)
	return log
}

func (s *Service) RecordBehaviorStrict(req BehaviorRequest) (BehaviorLog, error) {
	eventCode := firstNonEmpty(req.EventCode, req.EventType, "unknown")
	eventType := firstNonEmpty(req.EventType, eventCode)
	businessType := firstNonEmpty(req.BusinessType, req.TargetType, "unknown")
	targetType := firstNonEmpty(req.TargetType, businessType)
	businessID := req.BusinessID
	if businessID == 0 {
		businessID = req.TargetID
	}
	targetID := req.TargetID
	if targetID == 0 {
		targetID = businessID
	}
	source := firstNonEmpty(req.Source, "app")
	occurredAt := req.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	log := BehaviorLog{
		UserID:       req.UserID,
		EventType:    eventType,
		EventCode:    eventCode,
		TargetType:   targetType,
		BusinessType: businessType,
		TargetID:     targetID,
		BusinessID:   businessID,
		PagePath:     req.PagePath,
		Keyword:      req.Keyword,
		Source:       source,
		Device:       req.Device,
		IP:           req.IP,
		Extra:        marshalMap(req.Extra),
		CreatedAt:    time.Now(),
		OccurredAt:   occurredAt,
	}
	if s.behaviorRepo != nil {
		if err := s.behaviorRepo.SaveBehavior(context.Background(), log); err != nil {
			return BehaviorLog{}, err
		}
	}
	s.mu.Lock()
	log.ID = s.nextBehaviorID
	s.nextBehaviorID++
	s.behaviorLogs = append(s.behaviorLogs, log)
	s.mu.Unlock()
	return log, nil
}

func (s *Service) RecordOperation(req OperationRequest) OperationLog {
	log, _ := s.RecordOperationStrict(req)
	return log
}

func (s *Service) RecordOperationStrict(req OperationRequest) (OperationLog, error) {
	if req.Action == "" {
		req.Action = "unknown"
	}
	log := OperationLog{
		AdminUserID: req.AdminUserID,
		Action:      req.Action,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		RequestID:   req.RequestID,
		IP:          req.IP,
		Detail:      marshalMap(req.Detail),
		Before:      marshalMap(req.Before),
		After:       marshalMap(req.After),
		CreatedAt:   time.Now(),
	}
	if len(log.After) == 0 {
		log.After = append(log.After, log.Detail...)
	}
	if s.operationRepo != nil {
		if err := s.operationRepo.SaveOperation(context.Background(), log); err != nil {
			return OperationLog{}, err
		}
	}
	s.mu.Lock()
	log.ID = s.nextOperationID
	s.nextOperationID++
	s.operationLogs = append(s.operationLogs, log)
	s.mu.Unlock()
	return log, nil
}

func (s *Service) BehaviorLogs() []BehaviorLog {
	items, _ := s.BehaviorLogsStrict()
	return items
}

func (s *Service) BehaviorLogsStrict() ([]BehaviorLog, error) {
	if s.behaviorRepo != nil {
		return s.behaviorRepo.ListBehavior(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]BehaviorLog(nil), s.behaviorLogs...), nil
}

func (s *Service) QueryBehaviorLogs(query BehaviorQuery) []BehaviorLog {
	items, _ := s.QueryBehaviorLogsStrict(query)
	return items
}

func (s *Service) QueryBehaviorLogsStrict(query BehaviorQuery) ([]BehaviorLog, error) {
	if s.behaviorRepo != nil {
		return s.behaviorRepo.QueryBehavior(context.Background(), query)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BehaviorLog, 0)
	for _, item := range s.behaviorLogs {
		if query.UserID > 0 && item.UserID != query.UserID {
			continue
		}
		if query.EventType != "" && item.EventType != query.EventType {
			continue
		}
		if query.EventCode != "" && item.EventCode != query.EventCode {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) Funnel(eventCodes []string) FunnelSnapshot {
	snapshot, _ := s.FunnelStrict(eventCodes)
	return snapshot
}

func (s *Service) FunnelStrict(eventCodes []string) (FunnelSnapshot, error) {
	logs, err := s.BehaviorLogsStrict()
	if err != nil {
		return FunnelSnapshot{}, err
	}
	return funnelFromLogs(logs, eventCodes), nil
}

func funnelFromLogs(logs []BehaviorLog, eventCodes []string) FunnelSnapshot {
	if len(eventCodes) == 0 {
		eventCodes = []string{"login", "browse_games", "view_game_detail", "apply_game", "enter_im", "send_message", "service_confirm", "submit_review", "view_income_summary"}
	}
	userSets := make([]map[int64]struct{}, len(eventCodes))
	for i := range userSets {
		userSets[i] = make(map[int64]struct{})
	}
	stepIndex := make(map[string]int, len(eventCodes))
	for i, eventCode := range eventCodes {
		stepIndex[eventCode] = i
	}
	for _, log := range logs {
		if log.UserID <= 0 {
			continue
		}
		if index, ok := stepIndex[log.EventCode]; ok {
			userSets[index][log.UserID] = struct{}{}
		}
	}
	steps := make([]FunnelStep, 0, len(eventCodes))
	firstCount := len(userSets[0])
	previousCount := 0
	for i, eventCode := range eventCodes {
		userCount := len(userSets[i])
		step := FunnelStep{EventCode: eventCode, UserCount: userCount}
		if firstCount > 0 {
			step.ConversionRate = ratio(userCount, firstCount)
		}
		if i > 0 && previousCount > 0 {
			step.DropOffRate = 1 - ratio(userCount, previousCount)
		}
		steps = append(steps, step)
		previousCount = userCount
	}
	return FunnelSnapshot{Steps: steps, UpdatedAt: time.Now()}
}

func (s *Service) Retention() RetentionSnapshot {
	snapshot, _ := s.RetentionStrict()
	return snapshot
}

func (s *Service) RetentionStrict() (RetentionSnapshot, error) {
	logs, err := s.BehaviorLogsStrict()
	if err != nil {
		return RetentionSnapshot{}, err
	}
	return retentionFromLogs(logs), nil
}

func retentionFromLogs(logs []BehaviorLog) RetentionSnapshot {
	firstSeen := make(map[int64]time.Time)
	activeDates := make(map[int64]map[string]struct{})
	for _, log := range logs {
		if log.UserID <= 0 {
			continue
		}
		occurredAt := log.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = log.CreatedAt
		}
		if occurredAt.IsZero() {
			continue
		}
		if current, ok := firstSeen[log.UserID]; !ok || occurredAt.Before(current) {
			firstSeen[log.UserID] = occurredAt
		}
		if activeDates[log.UserID] == nil {
			activeDates[log.UserID] = make(map[string]struct{})
		}
		activeDates[log.UserID][dateKey(occurredAt)] = struct{}{}
	}
	cohorts := make(map[string][]int64)
	for userID, firstAt := range firstSeen {
		cohorts[dateKey(firstAt)] = append(cohorts[dateKey(firstAt)], userID)
	}
	dates := make([]string, 0, len(cohorts))
	for date := range cohorts {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	buckets := make([]RetentionBucket, 0, len(dates))
	for _, cohortDate := range dates {
		users := cohorts[cohortDate]
		day1 := shiftedDateKey(cohortDate, 1)
		day7 := shiftedDateKey(cohortDate, 7)
		day30 := shiftedDateKey(cohortDate, 30)
		bucket := RetentionBucket{CohortDate: cohortDate, NewUsers: len(users)}
		for _, userID := range users {
			if _, ok := activeDates[userID][day1]; ok {
				bucket.Day1Retained++
			}
			if _, ok := activeDates[userID][day7]; ok {
				bucket.Day7Retained++
			}
			if _, ok := activeDates[userID][day30]; ok {
				bucket.Day30Retained++
			}
		}
		bucket.Day1Rate = ratio(bucket.Day1Retained, bucket.NewUsers)
		bucket.Day7Rate = ratio(bucket.Day7Retained, bucket.NewUsers)
		bucket.Day30Rate = ratio(bucket.Day30Retained, bucket.NewUsers)
		buckets = append(buckets, bucket)
	}
	return RetentionSnapshot{Buckets: buckets, UpdatedAt: time.Now()}
}

func (s *Service) Dashboard() DashboardSnapshot {
	snapshot, _ := s.DashboardStrict()
	return snapshot
}

func (s *Service) DashboardStrict() (DashboardSnapshot, error) {
	logs, err := s.BehaviorLogsStrict()
	if err != nil {
		return DashboardSnapshot{}, err
	}
	return DashboardSnapshot{Funnel: funnelFromLogs(logs, nil), Retention: retentionFromLogs(logs), UpdatedAt: time.Now()}, nil
}

func (s *Service) OperationLogs() []OperationLog {
	items, _ := s.OperationLogsStrict()
	return items
}

func (s *Service) OperationLogsStrict() ([]OperationLog, error) {
	if s.operationRepo != nil {
		return s.operationRepo.ListOperations(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]OperationLog(nil), s.operationLogs...), nil
}

func ratio(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func dateKey(value time.Time) string {
	return value.UTC().Format("2006-01-02")
}

func shiftedDateKey(cohortDate string, days int) string {
	value, err := time.Parse("2006-01-02", cohortDate)
	if err != nil {
		return cohortDate
	}
	return value.AddDate(0, 0, days).Format("2006-01-02")
}

func marshalMap(value map[string]interface{}) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return data
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
