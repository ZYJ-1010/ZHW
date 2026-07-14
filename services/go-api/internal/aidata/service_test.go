package aidata

import (
	"context"
	"testing"

	"zhw-mini/services/go-api/internal/im"
)

func TestIMExportConfigUsesRepository(t *testing.T) {
	repo := &fakeAIDataRepository{}
	service := NewServiceWithRepository(repo)

	if service.IMExportConfig().Enabled {
		t.Fatal("expected repository export config to default disabled")
	}
	if _, err := service.ExportIMMessages([]im.Message{{ID: 1}}); err != ErrIMExportDisabled {
		t.Fatalf("expected disabled export error, got %v", err)
	}

	config := service.SetIMExportEnabled(true)
	if !config.Enabled || !service.IMExportConfig().Enabled {
		t.Fatalf("expected repository config to be enabled: %+v", config)
	}
	items, err := service.ExportIMMessages([]im.Message{{ID: 1, RoomID: 2, GameID: 3, SenderID: 4, Type: "text", Content: "ok"}})
	if err != nil {
		t.Fatalf("expected export success: %v", err)
	}
	if len(items) != 1 || items[0].MessageID != 1 || items[0].Content != "ok" {
		t.Fatalf("unexpected exported items: %+v", items)
	}
}

type fakeAIDataRepository struct {
	enabled bool
}

func (r *fakeAIDataRepository) IMExportConfig(ctx context.Context) (IMExportConfig, error) {
	return IMExportConfig{Enabled: r.enabled, UpdatedAt: "2026-06-17T18:00:00Z"}, nil
}

func (r *fakeAIDataRepository) SetIMExportEnabled(ctx context.Context, enabled bool) (IMExportConfig, error) {
	r.enabled = enabled
	return r.IMExportConfig(ctx)
}
