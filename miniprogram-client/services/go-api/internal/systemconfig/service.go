package systemconfig

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("system config not found")

type Repository interface {
	Get(ctx context.Context, key string) (json.RawMessage, error)
	Set(ctx context.Context, key string, value json.RawMessage) error
}

type Service struct {
	mu         sync.RWMutex
	values     map[string]json.RawMessage
	repository Repository
}

func NewService() *Service {
	return &Service{values: map[string]json.RawMessage{}}
}

func NewServiceWithRepository(repository Repository) *Service {
	return &Service{values: map[string]json.RawMessage{}, repository: repository}
}

func (s *Service) Get(key string, target interface{}) bool {
	if s.repository != nil {
		value, err := s.repository.Get(context.Background(), key)
		if err == nil {
			return json.Unmarshal(value, target) == nil
		}
	}
	s.mu.RLock()
	value, ok := s.values[key]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return json.Unmarshal(value, target) == nil
}

func (s *Service) Set(key string, value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if s.repository != nil {
		if err := s.repository.Set(context.Background(), key, raw); err != nil {
			return err
		}
	}
	s.mu.Lock()
	s.values[key] = append(json.RawMessage(nil), raw...)
	s.mu.Unlock()
	return nil
}
