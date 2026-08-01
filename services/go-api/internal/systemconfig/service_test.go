package systemconfig

import (
	"context"
	"encoding/json"
	"errors"
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

func TestServiceConfigurationSurvivesServiceRecreation(t *testing.T) {
	repository := &fakeRepository{values: map[string]json.RawMessage{}}
	firstService := NewServiceWithRepository(repository)
	if err := firstService.Set("growth.reward_rules", map[string]int{"experiencePerLevel": 240}); err != nil {
		t.Fatal(err)
	}

	// 新建服务实例模拟 API 重启。配置必须从持久化仓库恢复，而不能依赖内存缓存。
	secondService := NewServiceWithRepository(repository)
	var restored map[string]int
	if !secondService.Get("growth.reward_rules", &restored) {
		t.Fatal("expected persisted growth rules after service recreation")
	}
	if restored["experiencePerLevel"] != 240 {
		t.Fatalf("expected persisted experience per level, got %#v", restored)
	}
}

func TestRepositoryReadFailureDoesNotUseStaleProcessConfig(t *testing.T) {
	repository := &fakeRepository{values: map[string]json.RawMessage{}}
	service := NewServiceWithRepository(repository)
	if err := service.Set("operation.rules", map[string]int{"credit": 80}); err != nil {
		t.Fatal(err)
	}
	repository.getErr = errors.New("configuration database unavailable")
	var value map[string]int
	if service.Get("operation.rules", &value) {
		t.Fatalf("expected repository read failure, got stale config=%+v", value)
	}
}

type fakeRepository struct {
	values map[string]json.RawMessage
	getErr error
}

func (r *fakeRepository) Get(ctx context.Context, key string) (json.RawMessage, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
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
