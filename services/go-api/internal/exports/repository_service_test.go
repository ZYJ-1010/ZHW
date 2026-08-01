package exports

import (
	"context"
	"errors"
	"testing"
)

func TestRepositoryReadFailureDoesNotLookLikeEmptyExportQueue(t *testing.T) {
	repo := &failingExportRepository{err: errors.New("export database unavailable")}
	service := NewServiceWithRepository(repo)
	service.tasks[1] = Task{ID: 1, Status: "pending", TemplateCode: "reports_default"}

	if _, err := service.TasksStrict(); !errors.Is(err, repo.err) {
		t.Fatalf("expected task list error, got %v", err)
	}
	if _, err := service.RunPending(20, func(Task, Template) (int64, string, error) {
		return 1, "unused", nil
	}); !errors.Is(err, repo.err) {
		t.Fatalf("expected run pending read error, got %v", err)
	}
}

func TestRunPendingDoesNotReportDoneWhenTaskPersistenceFails(t *testing.T) {
	persistErr := errors.New("export task update unavailable")
	repo := &updateFailingExportRepository{task: Task{ID: 1, TaskNo: "EXP000001", TemplateCode: "reports_default", ExportType: "reports", Status: "pending"}, err: persistErr}
	service := NewServiceWithRepository(repo)
	results, err := service.RunPending(20, func(Task, Template) (int64, string, error) {
		return 99, "exports/reports.csv", nil
	})
	if !errors.Is(err, persistErr) {
		t.Fatalf("expected task update failure, got %v", err)
	}
	if results != nil {
		t.Fatalf("must not report a downloadable result when done status was not persisted: %+v", results)
	}
}

type failingExportRepository struct{ err error }

func (r *failingExportRepository) CreateTask(context.Context, Task) (Task, error) {
	return Task{}, r.err
}

func (r *failingExportRepository) ListTasks(context.Context) ([]Task, error) {
	return nil, r.err
}

func (r *failingExportRepository) GetTask(context.Context, int64) (Task, bool, error) {
	return Task{}, false, r.err
}

func (r *failingExportRepository) UpdateTask(context.Context, Task) (Task, error) {
	return Task{}, r.err
}

func (r *failingExportRepository) SaveContent(context.Context, int64, []byte) error {
	return r.err
}

func (r *failingExportRepository) LoadContent(context.Context, int64) ([]byte, bool, error) {
	return nil, false, r.err
}

type updateFailingExportRepository struct {
	task Task
	err  error
}

func (r *updateFailingExportRepository) CreateTask(context.Context, Task) (Task, error) {
	return Task{}, r.err
}

func (r *updateFailingExportRepository) ListTasks(context.Context) ([]Task, error) {
	return []Task{r.task}, nil
}

func (r *updateFailingExportRepository) GetTask(_ context.Context, taskID int64) (Task, bool, error) {
	return r.task, r.task.ID == taskID, nil
}

func (r *updateFailingExportRepository) UpdateTask(context.Context, Task) (Task, error) {
	return Task{}, r.err
}

func (r *updateFailingExportRepository) SaveContent(context.Context, int64, []byte) error {
	return r.err
}

func (r *updateFailingExportRepository) LoadContent(context.Context, int64) ([]byte, bool, error) {
	return nil, false, r.err
}
