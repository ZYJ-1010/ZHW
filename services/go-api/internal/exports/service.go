package exports

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTemplateNotFound = errors.New("export template not found")
	ErrTaskNotFound     = errors.New("export task not found")
	ErrInvalidRequest   = errors.New("invalid export request")
	ErrTaskNotReady     = errors.New("export task not ready")
)

type Template struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	ExportType  string   `json:"exportType"`
	FileName    string   `json:"fileName"`
	Columns     []string `json:"columns"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled"`
}

type Task struct {
	ID           int64                  `json:"id"`
	TaskNo       string                 `json:"taskNo"`
	TemplateCode string                 `json:"templateCode"`
	ExportType   string                 `json:"exportType"`
	Status       string                 `json:"status"`
	FileID       int64                  `json:"fileId,omitempty"`
	CreatedBy    int64                  `json:"createdBy,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	FailReason   string                 `json:"failReason,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	FinishedAt   string                 `json:"finishedAt,omitempty"`
}

type CreateRequest struct {
	TemplateCode string                 `json:"templateCode"`
	ExportType   string                 `json:"exportType"`
	Filters      map[string]interface{} `json:"filters"`
}

type RunResult struct {
	Task        Task     `json:"task"`
	FileName    string   `json:"fileName"`
	StorageKey  string   `json:"storageKey"`
	ContentType string   `json:"contentType"`
	Columns     []string `json:"columns"`
}

type Service struct {
	mu        sync.RWMutex
	nextID    int64
	templates []Template
	tasks     map[int64]Task
}

func NewService() *Service {
	service := &Service{
		nextID: 1,
		templates: []Template{
			{
				Code:        "reports_default",
				Name:        "举报申诉报表",
				ExportType:  "reports",
				FileName:    "reports-export.csv",
				Columns:     []string{"report_id", "game_id", "reporter_user_id", "report_type", "status", "created_at"},
				Description: "固定格式导出举报申诉列表",
				Enabled:     true,
			},
			{
				Code:        "operation_logs_default",
				Name:        "操作日志报表",
				ExportType:  "operation_logs",
				FileName:    "operation-logs-export.csv",
				Columns:     []string{"log_id", "admin_user_id", "action", "target_type", "target_id", "created_at"},
				Description: "固定格式导出后台操作日志",
				Enabled:     true,
			},
		},
		tasks: make(map[int64]Task),
	}
	service.templates = append(service.templates, reviewExportTemplate())
	return service
}

func reviewExportTemplate() Template {
	return Template{
		Code:        "reviews_default",
		Name:        "reviews export",
		ExportType:  "reviews",
		FileName:    "reviews-export.csv",
		Columns:     []string{"review_id", "game_id", "reviewer_user_id", "target_user_id", "score", "tags", "again_intent", "created_at"},
		Description: "Export review tags and again intent for analytics.",
		Enabled:     true,
	}
}

func (s *Service) Templates() []Template {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Template(nil), s.templates...)
}

func (s *Service) Create(adminID int64, req CreateRequest) (Task, error) {
	tpl, ok := s.templateByRequest(req)
	if !ok {
		return Task{}, ErrTemplateNotFound
	}
	if !tpl.Enabled {
		return Task{}, ErrInvalidRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	task := Task{
		ID:           s.nextID,
		TaskNo:       fmt.Sprintf("EXP%06d", s.nextID),
		TemplateCode: tpl.Code,
		ExportType:   tpl.ExportType,
		Status:       "pending",
		CreatedBy:    adminID,
		Filters:      cloneMap(req.Filters),
		CreatedAt:    time.Now(),
	}
	s.nextID++
	s.tasks[task.ID] = task
	return task, nil
}

func (s *Service) Tasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		result = append(result, task)
	}
	return result
}

func (s *Service) Get(taskID int64) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (s *Service) RunPending(limit int, createFile func(task Task, template Template) (int64, string, error)) ([]RunResult, error) {
	if limit <= 0 {
		limit = 20
	}
	results := make([]RunResult, 0)
	for _, task := range s.pendingTasks(limit) {
		template, _ := s.templateByCode(task.TemplateCode)
		fileID, storageKey, err := createFile(task, template)
		if err != nil {
			s.markFailed(task.ID, err.Error())
			continue
		}
		updated := s.markDone(task.ID, fileID)
		results = append(results, RunResult{
			Task:        updated,
			FileName:    template.FileName,
			StorageKey:  storageKey,
			ContentType: "text/csv",
			Columns:     append([]string(nil), template.Columns...),
		})
	}
	return results, nil
}

func (s *Service) ReadyFileID(taskID int64) (int64, error) {
	task, err := s.Get(taskID)
	if err != nil {
		return 0, err
	}
	if task.Status != "done" || task.FileID <= 0 {
		return 0, ErrTaskNotReady
	}
	return task.FileID, nil
}

func (s *Service) pendingTasks(limit int) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Task, 0)
	for _, task := range s.tasks {
		if task.Status == "pending" {
			result = append(result, task)
			if len(result) >= limit {
				break
			}
		}
	}
	return result
}

func (s *Service) markDone(taskID int64, fileID int64) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	task := s.tasks[taskID]
	task.Status = "done"
	task.FileID = fileID
	task.FinishedAt = time.Now().Format(time.RFC3339)
	s.tasks[taskID] = task
	return task
}

func (s *Service) markFailed(taskID int64, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task := s.tasks[taskID]
	task.Status = "failed"
	task.FailReason = reason
	task.FinishedAt = time.Now().Format(time.RFC3339)
	s.tasks[taskID] = task
}

func (s *Service) templateByRequest(req CreateRequest) (Template, bool) {
	if req.TemplateCode != "" {
		return s.templateByCode(req.TemplateCode)
	}
	if req.ExportType != "" {
		return s.templateByType(req.ExportType)
	}
	return Template{}, false
}

func (s *Service) templateByCode(code string) (Template, bool) {
	for _, template := range s.templates {
		if template.Code == code {
			return template, true
		}
	}
	return Template{}, false
}

func (s *Service) templateByType(exportType string) (Template, bool) {
	for _, template := range s.templates {
		if template.ExportType == exportType {
			return template, true
		}
	}
	return Template{}, false
}

func cloneMap(value map[string]interface{}) map[string]interface{} {
	if len(value) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}
