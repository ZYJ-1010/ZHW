package gamedrafts

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound  = errors.New("game draft not found")
	ErrForbidden = errors.New("game draft forbidden")
	ErrInvalid   = errors.New("invalid game draft")
)

type Draft struct {
	ID            int64           `json:"id"`
	CreatorUserID int64           `json:"creatorUserId"`
	Title         string          `json:"title"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type Repository interface {
	Create(ctx context.Context, draft Draft) (Draft, error)
	Update(ctx context.Context, draft Draft) (Draft, error)
	GetByUser(ctx context.Context, userID int64, draftID int64) (Draft, error)
	ListByUser(ctx context.Context, userID int64) ([]Draft, error)
	DeleteByUser(ctx context.Context, userID int64, draftID int64) error
}

type Service struct {
	mu     sync.RWMutex
	nextID int64
	repo   Repository
	drafts map[int64]Draft
}

func NewService() *Service {
	return &Service{nextID: 1, drafts: make(map[int64]Draft)}
}

func NewServiceWithRepository(repository Repository) *Service {
	service := NewService()
	service.repo = repository
	return service
}

func (s *Service) Save(userID int64, draftID int64, title string, payload json.RawMessage) (Draft, error) {
	if userID <= 0 || len(payload) == 0 || !json.Valid(payload) {
		return Draft{}, ErrInvalid
	}
	title = strings.TrimSpace(title)
	if len([]rune(title)) > 100 {
		return Draft{}, ErrInvalid
	}
	if title == "" {
		title = "未命名组局"
	}
	if len(payload) > 128*1024 {
		return Draft{}, ErrInvalid
	}

	now := time.Now()
	if s.repo != nil {
		if draftID > 0 {
			existing, err := s.repo.GetByUser(context.Background(), userID, draftID)
			if err != nil {
				return Draft{}, err
			}
			existing.Title = title
			existing.Payload = append(json.RawMessage(nil), payload...)
			existing.UpdatedAt = now
			return s.repo.Update(context.Background(), existing)
		}
		return s.repo.Create(context.Background(), Draft{CreatorUserID: userID, Title: title, Payload: append(json.RawMessage(nil), payload...), CreatedAt: now, UpdatedAt: now})
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if draftID > 0 {
		existing, ok := s.drafts[draftID]
		if !ok {
			return Draft{}, ErrNotFound
		}
		if existing.CreatorUserID != userID {
			return Draft{}, ErrForbidden
		}
		existing.Title = title
		existing.Payload = append(json.RawMessage(nil), payload...)
		existing.UpdatedAt = now
		s.drafts[draftID] = existing
		return existing, nil
	}

	draft := Draft{ID: s.nextID, CreatorUserID: userID, Title: title, Payload: append(json.RawMessage(nil), payload...), CreatedAt: now, UpdatedAt: now}
	s.nextID++
	s.drafts[draft.ID] = draft
	return draft, nil
}

func (s *Service) Get(userID int64, draftID int64) (Draft, error) {
	if userID <= 0 || draftID <= 0 {
		return Draft{}, ErrNotFound
	}
	if s.repo != nil {
		return s.repo.GetByUser(context.Background(), userID, draftID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	draft, ok := s.drafts[draftID]
	if !ok {
		return Draft{}, ErrNotFound
	}
	if draft.CreatorUserID != userID {
		return Draft{}, ErrForbidden
	}
	return draft, nil
}

func (s *Service) List(userID int64) ([]Draft, error) {
	if userID <= 0 {
		return []Draft{}, nil
	}
	if s.repo != nil {
		return s.repo.ListByUser(context.Background(), userID)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Draft, 0)
	for _, draft := range s.drafts {
		if draft.CreatorUserID == userID {
			items = append(items, draft)
		}
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].UpdatedAt.After(items[right].UpdatedAt)
	})
	return items, nil
}

func (s *Service) Delete(userID int64, draftID int64) error {
	if userID <= 0 || draftID <= 0 {
		return ErrNotFound
	}
	if s.repo != nil {
		return s.repo.DeleteByUser(context.Background(), userID, draftID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	draft, ok := s.drafts[draftID]
	if !ok {
		return ErrNotFound
	}
	if draft.CreatorUserID != userID {
		return ErrForbidden
	}
	delete(s.drafts, draftID)
	return nil
}
