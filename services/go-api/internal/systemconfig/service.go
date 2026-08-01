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
	found, _ := s.GetStrict(key, target)
	return found
}

func (s *Service) GetStrict(key string, target interface{}) (bool, error) {
	if s.repository != nil {
		value, err := s.repository.Get(context.Background(), key)
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if err := json.Unmarshal(value, target); err != nil {
			return false, err
		}
		return true, nil
	}
	s.mu.RLock()
	value, ok := s.values[key]
	s.mu.RUnlock()
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(value, target); err != nil {
		return false, err
	}
	return true, nil
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
