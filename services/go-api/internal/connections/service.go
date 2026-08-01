package connections

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrConnectionNotFound  = errors.New("connection not found")
	ErrConnectionForbidden = errors.New("connection forbidden")
	ErrInvalidFollowLog    = errors.New("invalid follow log")
	ErrInvalidConnection   = errors.New("invalid connection")
)

type Connection struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"userId"`
	ConnectedUserID int64     `json:"connectedUserId"`
	RelationType    string    `json:"relationType"`
	SourceType      string    `json:"sourceType"`
	SourceID        int64     `json:"sourceId,omitempty"`
	StrengthScore   int       `json:"strengthScore"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type FollowLog struct {
	ID             int64     `json:"id"`
	ConnectionID   int64     `json:"connectionId"`
	OperatorUserID int64     `json:"operatorUserId"`
	FollowType     string    `json:"followType"`
	Content        string    `json:"content"`
	NextFollowAt   string    `json:"nextFollowAt,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type FollowRequest struct {
	FollowType   string `json:"followType"`
	Content      string `json:"content"`
	NextFollowAt string `json:"nextFollowAt"`
}

type Repository interface {
	UpsertConnection(ctx context.Context, connection Connection, strengthDelta int) (Connection, error)
	UpsertConnectionPair(ctx context.Context, first Connection, second Connection, strengthDelta int) (Connection, Connection, error)
	ListConnections(ctx context.Context, userID int64) ([]Connection, error)
	ListAllConnections(ctx context.Context) ([]Connection, error)
	FindConnection(ctx context.Context, connectionID int64) (Connection, bool, error)
	AddFollowLog(ctx context.Context, log FollowLog) (FollowLog, error)
	IncrementStrength(ctx context.Context, connectionID int64, delta int) (Connection, error)
	AddFollowLogAndIncrement(ctx context.Context, log FollowLog, strengthDelta int) (FollowLog, Connection, error)
}

type Service struct {
	mu               sync.RWMutex
	nextConnectionID int64
	nextFollowID     int64
	connections      map[int64]Connection
	byKey            map[string]int64
	followLogs       map[int64][]FollowLog
	repo             Repository
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{
		nextConnectionID: 1,
		nextFollowID:     1,
		connections:      make(map[int64]Connection),
		byKey:            make(map[string]int64),
		followLogs:       make(map[int64][]FollowLog),
		repo:             repo,
	}
}

func (s *Service) UpsertPair(userID int64, connectedUserID int64, relationType string, sourceType string, sourceID int64, strengthDelta int) {
	_ = s.UpsertPairStrict(userID, connectedUserID, relationType, sourceType, sourceID, strengthDelta)
}

func (s *Service) UpsertPairStrict(userID int64, connectedUserID int64, relationType string, sourceType string, sourceID int64, strengthDelta int) error {
	if userID <= 0 || connectedUserID <= 0 || userID == connectedUserID {
		return ErrInvalidConnection
	}
	if s.repo != nil {
		_, _, err := s.repo.UpsertConnectionPair(
			context.Background(),
			newConnection(userID, connectedUserID, relationType, sourceType, sourceID, strengthDelta),
			newConnection(connectedUserID, userID, relationType, sourceType, sourceID, strengthDelta),
			strengthDelta,
		)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upsertLocked(userID, connectedUserID, relationType, sourceType, sourceID, strengthDelta)
	s.upsertLocked(connectedUserID, userID, relationType, sourceType, sourceID, strengthDelta)
	return nil
}

func (s *Service) My(userID int64) []Connection {
	items, _ := s.MyStrict(userID)
	return items
}

func (s *Service) MyStrict(userID int64) ([]Connection, error) {
	if s.repo != nil {
		return s.repo.ListConnections(context.Background(), userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Connection, 0)
	for _, item := range s.connections {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Service) All() []Connection {
	items, _ := s.AllStrict()
	return items
}

func (s *Service) AllStrict() ([]Connection, error) {
	if s.repo != nil {
		return s.repo.ListAllConnections(context.Background())
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Connection, 0, len(s.connections))
	for _, item := range s.connections {
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) AddFollowLog(operatorUserID int64, connectionID int64, req FollowRequest, guideAllowed bool) (FollowLog, error) {
	req.FollowType = strings.TrimSpace(req.FollowType)
	req.Content = strings.TrimSpace(req.Content)
	req.NextFollowAt = strings.TrimSpace(req.NextFollowAt)
	if req.FollowType == "" || req.Content == "" || len(req.FollowType) > 32 || len(req.Content) > 500 || len(req.NextFollowAt) > 64 || !validFollowType(req.FollowType) || !validFollowTime(req.NextFollowAt) {
		return FollowLog{}, ErrInvalidFollowLog
	}
	if s.repo != nil {
		conn, ok, err := s.repo.FindConnection(context.Background(), connectionID)
		if err != nil {
			return FollowLog{}, err
		}
		if !ok {
			return FollowLog{}, ErrConnectionNotFound
		}
		if conn.UserID != operatorUserID && conn.ConnectedUserID != operatorUserID && !guideAllowed {
			return FollowLog{}, ErrConnectionForbidden
		}
		log, _, err := s.repo.AddFollowLogAndIncrement(context.Background(), FollowLog{
			ConnectionID:   connectionID,
			OperatorUserID: operatorUserID,
			FollowType:     req.FollowType,
			Content:        req.Content,
			NextFollowAt:   req.NextFollowAt,
			CreatedAt:      time.Now(),
		}, 1)
		if err != nil {
			return FollowLog{}, err
		}
		return log, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	conn, ok := s.connections[connectionID]
	if !ok {
		return FollowLog{}, ErrConnectionNotFound
	}
	if conn.UserID != operatorUserID && conn.ConnectedUserID != operatorUserID && !guideAllowed {
		return FollowLog{}, ErrConnectionForbidden
	}
	log := FollowLog{
		ID:             s.nextFollowID,
		ConnectionID:   connectionID,
		OperatorUserID: operatorUserID,
		FollowType:     req.FollowType,
		Content:        req.Content,
		NextFollowAt:   req.NextFollowAt,
		CreatedAt:      time.Now(),
	}
	s.nextFollowID++
	s.followLogs[connectionID] = append(s.followLogs[connectionID], log)
	conn.StrengthScore += 1
	conn.UpdatedAt = time.Now()
	s.connections[connectionID] = conn
	return log, nil
}

func validFollowType(value string) bool {
	switch value {
	case "note", "guide", "call", "visit", "task", "other":
		return true
	default:
		return false
	}
}

func validFollowTime(value string) bool {
	if value == "" {
		return true
	}
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func (s *Service) upsertLocked(userID int64, connectedUserID int64, relationType string, sourceType string, sourceID int64, strengthDelta int) Connection {
	if relationType == "" {
		relationType = "user"
	}
	if sourceType == "" {
		sourceType = "manual"
	}
	key := connectionKey(userID, connectedUserID, relationType, sourceType, sourceID)
	now := time.Now()
	if id, ok := s.byKey[key]; ok {
		conn := s.connections[id]
		conn.StrengthScore += strengthDelta
		conn.UpdatedAt = now
		s.connections[id] = conn
		return conn
	}
	conn := Connection{
		ID:              s.nextConnectionID,
		UserID:          userID,
		ConnectedUserID: connectedUserID,
		RelationType:    relationType,
		SourceType:      sourceType,
		SourceID:        sourceID,
		StrengthScore:   maxInt(strengthDelta, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.nextConnectionID++
	s.connections[conn.ID] = conn
	s.byKey[key] = conn.ID
	return conn
}

func connectionKey(userID int64, connectedUserID int64, relationType string, sourceType string, sourceID int64) string {
	return strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(connectedUserID, 10) + ":" + relationType + ":" + sourceType + ":" + strconv.FormatInt(sourceID, 10)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func newConnection(userID int64, connectedUserID int64, relationType string, sourceType string, sourceID int64, strengthDelta int) Connection {
	if relationType == "" {
		relationType = "user"
	}
	if sourceType == "" {
		sourceType = "manual"
	}
	now := time.Now()
	return Connection{
		UserID:          userID,
		ConnectedUserID: connectedUserID,
		RelationType:    relationType,
		SourceType:      sourceType,
		SourceID:        sourceID,
		StrengthScore:   maxInt(strengthDelta, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func nullFollowTime(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{}
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return sql.NullTime{Time: parsed, Valid: true}
	}
	if parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", value); err == nil {
		return sql.NullTime{Time: parsed, Valid: true}
	}
	return sql.NullTime{}
}
