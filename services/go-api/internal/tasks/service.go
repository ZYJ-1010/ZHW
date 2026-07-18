package tasks

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("task progress not found")

type Progress struct {
	UserID      int64     `json:"userId"`
	TaskCode    string    `json:"taskCode"`
	CompletedAt time.Time `json:"completedAt"`
}

type Repository interface {
	UpsertProgress(ctx context.Context, progress Progress) (Progress, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
}

type Service struct {
	mu       sync.RWMutex
	repo     Repository
	progress map[string]Progress
}

func NewService() *Service { return &Service{progress: make(map[string]Progress)} }

func NewServiceWithRepository(repo Repository) *Service {
	service := NewService()
	service.repo = repo
	return service
}

func progressKey(userID int64, code string) string { return strconv.FormatInt(userID, 10) + ":" + code }

func (s *Service) MarkCompleted(userID int64, code string) (Progress, error) {
	if userID <= 0 || code == "" {
		return Progress{}, ErrTaskNotFound
	}
	progress := Progress{UserID: userID, TaskCode: code, CompletedAt: time.Now()}
	if s.repo != nil {
		return s.repo.UpsertProgress(context.Background(), progress)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := progressKey(userID, code)
	if existing, ok := s.progress[key]; ok {
		return existing, nil
	}
	s.progress[key] = progress
	return progress, nil
}

func (s *Service) CompletedCodes(userID int64) map[string]bool {
	result := make(map[string]bool)
	var items []Progress
	if s.repo != nil {
		items, _ = s.repo.ListProgress(context.Background(), userID)
	} else {
		s.mu.RLock()
		for _, item := range s.progress {
			if item.UserID == userID {
				result[item.TaskCode] = true
			}
		}
		s.mu.RUnlock()
		return result
	}
	for _, item := range items {
		result[item.TaskCode] = true
	}
	return result
}

type SQLRepository struct{ db *sql.DB }

func NewSQLRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) UpsertProgress(ctx context.Context, progress Progress) (Progress, error) {
	var saved Progress
	err := r.db.QueryRowContext(ctx, `
insert into user_task_progress (user_id, task_code, completed_at)
values ($1,$2,$3)
on conflict (user_id, task_code) do nothing
returning user_id, task_code, completed_at
`, progress.UserID, progress.TaskCode, progress.CompletedAt).Scan(&saved.UserID, &saved.TaskCode, &saved.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return r.get(ctx, progress.UserID, progress.TaskCode)
	}
	return saved, err
}

func (r *SQLRepository) get(ctx context.Context, userID int64, code string) (Progress, error) {
	var item Progress
	err := r.db.QueryRowContext(ctx, `select user_id, task_code, completed_at from user_task_progress where user_id=$1 and task_code=$2`, userID, code).Scan(&item.UserID, &item.TaskCode, &item.CompletedAt)
	return item, err
}

func (r *SQLRepository) ListProgress(ctx context.Context, userID int64) ([]Progress, error) {
	rows, err := r.db.QueryContext(ctx, `select user_id, task_code, completed_at from user_task_progress where user_id=$1 order by completed_at asc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Progress, 0)
	for rows.Next() {
		var item Progress
		if err := rows.Scan(&item.UserID, &item.TaskCode, &item.CompletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
