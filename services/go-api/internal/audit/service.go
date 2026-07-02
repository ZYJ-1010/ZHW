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
	s.mu.Lock()
	log := BehaviorLog{
		ID:           s.nextBehaviorID,
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
	s.nextBehaviorID++
	s.behaviorLogs = append(s.behaviorLogs, log)
	repo := s.behaviorRepo
	s.mu.Unlock()
	if repo != nil {
		_ = repo.SaveBehavior(context.Background(), log)
	}
	return log
}

func (s *Service) RecordOperation(req OperationRequest) OperationLog {
	if req.Action == "" {
		req.Action = "unknown"
	}
	s.mu.Lock()
	log := OperationLog{
		ID:          s.nextOperationID,
		AdminUserID: req.AdminUserID,
		Action:      req.Action,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		RequestID:   req.RequestID,
		IP:          req.IP,
		Detail:      marshalMap(req.Detail),
		CreatedAt:   time.Now(),
	}
	s.nextOperationID++
	s.operationLogs = append(s.operationLogs, log)
	repo := s.operationRepo
	s.mu.Unlock()
	if repo != nil {
		_ = repo.SaveOperation(context.Background(), log)
	}
	return log
}

func (s *Service) BehaviorLogs() []BehaviorLog {
	if s.behaviorRepo != nil {
		if items, err := s.behaviorRepo.ListBehavior(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]BehaviorLog(nil), s.behaviorLogs...)
}

func (s *Service) QueryBehaviorLogs(query BehaviorQuery) []BehaviorLog {
	if s.behaviorRepo != nil {
		if items, err := s.behaviorRepo.QueryBehavior(context.Background(), query); err == nil {
			return items
		}
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
	return result
}

func (s *Service) Funnel(eventCodes []string) FunnelSnapshot {
	if len(eventCodes) == 0 {
		eventCodes = []string{"login", "browse_games", "view_game_detail", "apply_game", "enter_im", "send_message", "service_confirm", "submit_review", "view_income_summary"}
	}
	logs := s.BehaviorLogs()
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
	logs := s.BehaviorLogs()
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
	return DashboardSnapshot{Funnel: s.Funnel(nil), Retention: s.Retention(), UpdatedAt: time.Now()}
}

func (s *Service) OperationLogs() []OperationLog {
	if s.operationRepo != nil {
		if items, err := s.operationRepo.ListOperations(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]OperationLog(nil), s.operationLogs...)
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
