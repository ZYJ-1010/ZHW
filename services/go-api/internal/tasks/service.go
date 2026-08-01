package tasks

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("task progress not found")

type Progress struct {
	UserID      int64     `json:"userId"`
	TaskCode    string    `json:"taskCode"`
	CompletedAt time.Time `json:"completedAt"`
}

// GuideProgress records delivery of the profile-completion reminder. It is
// intentionally independent from task completion: dismissing a reminder must
// never mark a newcomer task as completed.
type GuideProgress struct {
	UserID                int64     `json:"userId"`
	ProfileReminderCount  int       `json:"profileReminderCount"`
	LastProfileReminderAt time.Time `json:"lastProfileReminderAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type Repository interface {
	UpsertProgress(ctx context.Context, progress Progress) (Progress, bool, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
}

// GuideProgressRepository is optional so existing task repositories continue
// to work while the API server persists the guide state when SQL is available.
type GuideProgressRepository interface {
	GetGuideProgress(ctx context.Context, userID int64) (GuideProgress, error)
	RecordProfileReminder(ctx context.Context, userID int64, occurredAt time.Time) (GuideProgress, error)
}

type Service struct {
	mu       sync.RWMutex
	repo     Repository
	progress map[string]Progress
	guide    map[int64]GuideProgress
}

func NewService() *Service {
	return &Service{progress: make(map[string]Progress), guide: make(map[int64]GuideProgress)}
}

func NewServiceWithRepository(repo Repository) *Service {
	service := NewService()
	service.repo = repo
	return service
}

func (s *Service) GuideProgress(userID int64) GuideProgress {
	item, _ := s.GuideProgressStrict(userID)
	return item
}

// GuideProgressStrict never treats a repository read failure as an empty guide
// state. Callers that render a user-facing reminder must surface that failure
// instead of incorrectly showing a fresh-user flow.
func (s *Service) GuideProgressStrict(userID int64) (GuideProgress, error) {
	if userID <= 0 {
		return GuideProgress{}, nil
	}
	if repo, ok := s.repo.(GuideProgressRepository); ok {
		item, err := repo.GetGuideProgress(context.Background(), userID)
		if err != nil {
			return GuideProgress{}, err
		}
		return item, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.guide[userID], nil
}

func (s *Service) RecordProfileReminder(userID int64, occurredAt time.Time) (GuideProgress, error) {
	if userID <= 0 {
		return GuideProgress{}, ErrTaskNotFound
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	if repo, ok := s.repo.(GuideProgressRepository); ok {
		return repo.RecordProfileReminder(context.Background(), userID, occurredAt)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.guide[userID]
	item.UserID = userID
	item.ProfileReminderCount++
	item.LastProfileReminderAt = occurredAt
	item.UpdatedAt = occurredAt
	s.guide[userID] = item
	return item, nil
}

func progressKey(userID int64, code string) string { return strconv.FormatInt(userID, 10) + ":" + code }

func (s *Service) MarkCompleted(userID int64, code string) (Progress, error) {
	progress, _, err := s.MarkCompletedOnce(userID, code)
	return progress, err
}

func (s *Service) MarkCompletedForDate(userID int64, code string, date time.Time) (Progress, error) {
	progress, _, err := s.MarkCompletedForDateOnce(userID, code, date)
	return progress, err
}

// MarkCompletedOnce reports whether this call created the completion record.
// Reward callers must use this result instead of a separate read-before-write
// check, which races when the task center is opened concurrently.
func (s *Service) MarkCompletedOnce(userID int64, code string) (Progress, bool, error) {
	return s.markCompleted(userID, code)
}

func (s *Service) MarkCompletedForDateOnce(userID int64, code string, date time.Time) (Progress, bool, error) {
	day := date.Format("2006-01-02")
	if day == "" {
		return Progress{}, false, ErrTaskNotFound
	}
	return s.markCompleted(userID, code+"@"+day)
}

func (s *Service) markCompleted(userID int64, code string) (Progress, bool, error) {
	if userID <= 0 || code == "" {
		return Progress{}, false, ErrTaskNotFound
	}
	progress := Progress{UserID: userID, TaskCode: code, CompletedAt: time.Now()}
	if s.repo != nil {
		return s.repo.UpsertProgress(context.Background(), progress)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := progressKey(userID, code)
	if existing, ok := s.progress[key]; ok {
		return existing, false, nil
	}
	s.progress[key] = progress
	return progress, true, nil
}

func (s *Service) CompletedCodes(userID int64) map[string]bool {
	result, _ := s.CompletedCodesStrict(userID)
	return result
}

// CompletedCodesStrict keeps a persistence outage distinguishable from a user
// with no completed tasks. UI callers must use this form to avoid showing an
// already claimed reward as claimable again.
func (s *Service) CompletedCodesStrict(userID int64) (map[string]bool, error) {
	result := make(map[string]bool)
	var items []Progress
	if s.repo != nil {
		var err error
		items, err = s.repo.ListProgress(context.Background(), userID)
		if err != nil {
			return nil, err
		}
	} else {
		s.mu.RLock()
		for _, item := range s.progress {
			if item.UserID == userID {
				result[item.TaskCode] = true
			}
		}
		s.mu.RUnlock()
		return result, nil
	}
	for _, item := range items {
		result[item.TaskCode] = true
	}
	return result, nil
}

func (s *Service) CompletedCodesForDate(userID int64, date time.Time) map[string]bool {
	result, _ := s.CompletedCodesForDateStrict(userID, date)
	return result
}

func (s *Service) CompletedCodesForDateStrict(userID int64, date time.Time) (map[string]bool, error) {
	day := date.Format("2006-01-02")
	result := make(map[string]bool)
	codes, err := s.CompletedCodesStrict(userID)
	if err != nil {
		return nil, err
	}
	for code := range codes {
		if strings.HasSuffix(code, "@"+day) {
			result[strings.TrimSuffix(code, "@"+day)] = true
		}
	}
	return result, nil
}

type SQLRepository struct{ db *sql.DB }

func NewSQLRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) UpsertProgress(ctx context.Context, progress Progress) (Progress, bool, error) {
	var saved Progress
	err := r.db.QueryRowContext(ctx, `
insert into user_task_progress (user_id, task_code, completed_at)
values ($1,$2,$3)
on conflict (user_id, task_code) do nothing
returning user_id, task_code, completed_at
`, progress.UserID, progress.TaskCode, progress.CompletedAt).Scan(&saved.UserID, &saved.TaskCode, &saved.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		existing, getErr := r.get(ctx, progress.UserID, progress.TaskCode)
		return existing, false, getErr
	}
	return saved, err == nil, err
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

func (r *SQLRepository) GetGuideProgress(ctx context.Context, userID int64) (GuideProgress, error) {
	var item GuideProgress
	err := r.db.QueryRowContext(ctx, `
select user_id, profile_reminder_count, coalesce(last_profile_reminder_at, to_timestamp(0)), updated_at
from user_newbie_guide_progress where user_id=$1
`, userID).Scan(&item.UserID, &item.ProfileReminderCount, &item.LastProfileReminderAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return GuideProgress{UserID: userID}, nil
	}
	return item, err
}

func (r *SQLRepository) RecordProfileReminder(ctx context.Context, userID int64, occurredAt time.Time) (GuideProgress, error) {
	var item GuideProgress
	err := r.db.QueryRowContext(ctx, `
insert into user_newbie_guide_progress (user_id, profile_reminder_count, last_profile_reminder_at, updated_at)
values ($1, 1, $2, $2)
on conflict (user_id) do update set
  profile_reminder_count = user_newbie_guide_progress.profile_reminder_count + 1,
  last_profile_reminder_at = excluded.last_profile_reminder_at,
  updated_at = excluded.updated_at
returning user_id, profile_reminder_count, last_profile_reminder_at, updated_at
`, userID, occurredAt).Scan(&item.UserID, &item.ProfileReminderCount, &item.LastProfileReminderAt, &item.UpdatedAt)
	return item, err
}
