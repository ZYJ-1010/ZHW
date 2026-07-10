package systemconfig

import (
	"context"
	"encoding/json"
	"testing"
)

func TestServiceReadsAndWritesRepository(t *testing.T) {
	repository := &fakeRepository{values: map[string]json.RawMessage{}}
	service := NewServiceWithRepository(repository)

	if service.Get("missing", &map[string]string{}) {
		t.Fatal("expected missing config")
	}
	if err := service.Set("game.category_config", map[string]string{"version": "repo-test"}); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if !service.Get("game.category_config", &got) {
		t.Fatal("expected config")
	}
	if got["version"] != "repo-test" {
		t.Fatalf("unexpected config: %#v", got)
	}
}

type fakeRepository struct {
	values map[string]json.RawMessage
}

func (r *fakeRepository) Get(ctx context.Context, key string) (json.RawMessage, error) {
	value, ok := r.values[key]
	if !ok {
		return nil, ErrNotFound
	}
	return value, nil
}

func (r *fakeRepository) Set(ctx context.Context, key string, value json.RawMessage) error {
	r.values[key] = append(json.RawMessage(nil), value...)
	return nil
}
