package points

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidPoints      = errors.New("invalid points")
	ErrInsufficientPoints = errors.New("insufficient points")
)

type Account struct {
	UserID            int64     `json:"userId"`
	AvailablePoints   int       `json:"availablePoints"`
	FrozenPoints      int       `json:"frozenPoints"`
	TotalEarnedPoints int       `json:"totalEarnedPoints"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Log struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	ChangeValue  int       `json:"changeValue"`
	BeforePoints int       `json:"beforePoints"`
	AfterPoints  int       `json:"afterPoints"`
	BizType      string    `json:"bizType"`
	BizID        int64     `json:"bizId,omitempty"`
	Reason       string    `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Repository interface {
	GetAccount(ctx context.Context, userID int64) (Account, bool, error)
	SaveAccount(ctx context.Context, account Account) (Account, error)
	ListLogsByUser(ctx context.Context, userID int64) ([]Log, error)
	ListLogs(ctx context.Context) ([]Log, error)
	AddLog(ctx context.Context, log Log) (Log, error)
}

// atomicRepository is implemented by the PostgreSQL adapter. It keeps the
// balance and its ledger entry in one transaction.
type atomicRepository interface {
	ApplyChange(ctx context.Context, userID int64, changeValue int, bizType string, bizID int64, reason string) (Account, Log, error)
	ApplyChangeOnce(ctx context.Context, userID int64, changeValue int, bizType string, bizID int64, reason string) (Account, Log, bool, error)
}

type Service struct {
	mu       sync.RWMutex
	nextID   int64
	accounts map[int64]Account
	logs     []Log
	repo     Repository
}

func NewService() *Service {
	return NewServiceWithRepository(nil)
}

func NewServiceWithRepository(repo Repository) *Service {
	return &Service{
		nextID:   1,
		accounts: make(map[int64]Account),
		logs:     make([]Log, 0),
		repo:     repo,
	}
}

func (s *Service) Summary(userID int64) Account {
	if s.repo != nil {
		if account, ok, err := s.repo.GetAccount(context.Background(), userID); err == nil && ok {
			return account
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ensureLocked(userID)
}

func (s *Service) Logs(userID int64) []Log {
	if s.repo != nil {
		if items, err := s.repo.ListLogsByUser(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Log, 0)
	for _, item := range s.logs {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) AllLogs() []Log {
	if s.repo != nil {
		if items, err := s.repo.ListLogs(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Log(nil), s.logs...)
}

func (s *Service) Grant(userID int64, value int, bizType string, bizID int64, reason string) (Account, Log, error) {
	if value <= 0 {
		return Account{}, Log{}, ErrInvalidPoints
	}
	if s.repo != nil {
		return s.applyRepositoryChange(userID, value, bizType, bizID, reason)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	account := s.ensureLocked(userID)
	before := account.AvailablePoints
	account.AvailablePoints += value
	account.TotalEarnedPoints += value
	account.UpdatedAt = time.Now()
	s.accounts[userID] = account
	log := s.appendLogLocked(userID, value, before, account.AvailablePoints, bizType, bizID, reason)
	return account, log, nil
}

// GrantOnce is used for automatic business rewards. The same user, business
// type and business id can only create one reward ledger record, including in
// a multi-instance deployment.
func (s *Service) GrantOnce(userID int64, value int, bizType string, bizID int64, reason string) (Account, Log, bool, error) {
	if value <= 0 || bizID <= 0 || bizType == "" {
		return Account{}, Log{}, false, ErrInvalidPoints
	}
	if repository, ok := s.repo.(atomicRepository); ok {
		return repository.ApplyChangeOnce(context.Background(), userID, value, bizType, bizID, reason)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.logs {
		if item.UserID == userID && item.BizType == bizType && item.BizID == bizID {
			return s.ensureLocked(userID), item, false, nil
		}
	}
	account := s.ensureLocked(userID)
	before := account.AvailablePoints
	account.AvailablePoints += value
	account.TotalEarnedPoints += value
	account.UpdatedAt = time.Now()
	s.accounts[userID] = account
	log := s.appendLogLocked(userID, value, before, account.AvailablePoints, bizType, bizID, reason)
	return account, log, true, nil
}

func (s *Service) Deduct(userID int64, value int, bizType string, bizID int64, reason string) (Account, Log, error) {
	if value <= 0 {
		return Account{}, Log{}, ErrInvalidPoints
	}
	if s.repo != nil {
		return s.applyRepositoryChange(userID, -value, bizType, bizID, reason)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	account := s.ensureLocked(userID)
	if account.AvailablePoints < value {
		return account, Log{}, ErrInsufficientPoints
	}
	before := account.AvailablePoints
	account.AvailablePoints -= value
	account.UpdatedAt = time.Now()
	s.accounts[userID] = account
	log := s.appendLogLocked(userID, -value, before, account.AvailablePoints, bizType, bizID, reason)
	return account, log, nil
}

// Expire removes points earned before cutoff. Expiry is idempotent: previous
// points_expire logs are excluded from the next calculation, so a scheduled
// job can safely run repeatedly.
func (s *Service) Expire(userID int64, cutoff time.Time) (Account, Log, error) {
	if userID <= 0 {
		return Account{}, Log{}, ErrInvalidPoints
	}
	logs := s.Logs(userID)
	eligible, expired := 0, 0
	for _, item := range logs {
		if item.BizType == "points_expire" && item.ChangeValue < 0 {
			expired += -item.ChangeValue
			continue
		}
		if item.ChangeValue > 0 && item.CreatedAt.Before(cutoff) {
			eligible += item.ChangeValue
		}
	}
	amount := eligible - expired
	account := s.Summary(userID)
	if amount <= 0 || account.AvailablePoints <= 0 {
		return account, Log{}, nil
	}
	if amount > account.AvailablePoints {
		amount = account.AvailablePoints
	}
	return s.Deduct(userID, amount, "points_expire", 0, "积分到期扣减")
}

func (s *Service) ensureLocked(userID int64) Account {
	if account, ok := s.accounts[userID]; ok {
		return account
	}
	account := Account{UserID: userID, UpdatedAt: time.Now()}
	s.accounts[userID] = account
	return account
}

func (s *Service) appendLogLocked(userID int64, changeValue int, before int, after int, bizType string, bizID int64, reason string) Log {
	log := Log{
		ID:           s.nextID,
		UserID:       userID,
		ChangeValue:  changeValue,
		BeforePoints: before,
		AfterPoints:  after,
		BizType:      bizType,
		BizID:        bizID,
		Reason:       reason,
		CreatedAt:    time.Now(),
	}
	s.nextID++
	s.logs = append(s.logs, log)
	return log
}

func (s *Service) applyRepositoryChange(userID int64, changeValue int, bizType string, bizID int64, reason string) (Account, Log, error) {
	if repository, ok := s.repo.(atomicRepository); ok {
		return repository.ApplyChange(context.Background(), userID, changeValue, bizType, bizID, reason)
	}
	account, ok, err := s.repo.GetAccount(context.Background(), userID)
	if err != nil {
		return Account{}, Log{}, err
	}
	if !ok {
		account = Account{UserID: userID, UpdatedAt: time.Now()}
	}
	before := account.AvailablePoints
	after := before + changeValue
	if after < 0 {
		return account, Log{}, ErrInsufficientPoints
	}
	account.AvailablePoints = after
	if changeValue > 0 {
		account.TotalEarnedPoints += changeValue
	}
	account.UpdatedAt = time.Now()
	account, err = s.repo.SaveAccount(context.Background(), account)
	if err != nil {
		return Account{}, Log{}, err
	}
	log := Log{
		UserID:       userID,
		ChangeValue:  changeValue,
		BeforePoints: before,
		AfterPoints:  account.AvailablePoints,
		BizType:      bizType,
		BizID:        bizID,
		Reason:       reason,
		CreatedAt:    time.Now(),
	}
	log, err = s.repo.AddLog(context.Background(), log)
	if err != nil {
		return Account{}, Log{}, err
	}
	return account, log, nil
}
