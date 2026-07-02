package reviews

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) MarkReviewable(ctx context.Context, gameID int64, userID int64, deadline time.Time) error {
	_, err := r.db.ExecContext(ctx, `
insert into review_reminders (game_id, user_id, status, deadline_at, created_at)
values ($1,$2,'pending',$3,now())
on conflict (game_id, user_id) do update set deadline_at = excluded.deadline_at
`, gameID, userID, deadline)
	return err
}

func (r *SQLRepository) ListReviewable(ctx context.Context, userID int64) (map[int64]time.Time, error) {
	rows, err := r.db.QueryContext(ctx, `
select game_id, deadline_at
from review_reminders
where user_id = $1 and status = 'pending'
order by created_at desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[int64]time.Time)
	for rows.Next() {
		var gameID int64
		var deadline time.Time
		if err := rows.Scan(&gameID, &deadline); err != nil {
			return nil, err
		}
		result[gameID] = deadline
	}
	return result, rows.Err()
}

func (r *SQLRepository) ReviewExists(ctx context.Context, gameID int64, reviewerID int64, targetID int64, targetRole string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
select exists(
  select 1 from reviews
  where game_id = $1 and reviewer_user_id = $2 and target_user_id = $3 and target_role = $4
)
`, gameID, reviewerID, targetID, targetRole).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) SaveReview(ctx context.Context, review Review) (Review, error) {
	tags, _ := json.Marshal(review.Tags)
	return scanReview(r.db.QueryRowContext(ctx, `
insert into reviews (game_id, reviewer_user_id, target_user_id, target_role, score, content, tags, again_intent, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9)
returning id, game_id, reviewer_user_id, target_user_id, target_role, score, content, tags, again_intent, created_at
`, review.GameID, review.ReviewerUserID, review.TargetUserID, review.TargetRole, review.Score, review.Content, string(tags), review.AgainIntent, review.CreatedAt))
}

func (r *SQLRepository) ListReviews(ctx context.Context) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx, reviewSelect()+` order by created_at desc, id desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReviews(rows)
}

func (r *SQLRepository) ListReviewsByUser(ctx context.Context, userID int64) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx, reviewSelect()+`
where reviewer_user_id = $1 or target_user_id = $1
order by created_at desc, id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReviews(rows)
}

func (r *SQLRepository) ListReviewsByGame(ctx context.Context, gameID int64) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx, reviewSelect()+`
where game_id = $1
order by created_at desc, id desc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReviews(rows)
}

func (r *SQLRepository) GetGrowthProfile(ctx context.Context, userID int64) (GrowthProfile, bool, error) {
	var profile GrowthProfile
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, `
select user_id, level, experience, review_count, updated_at
from user_growth_profiles
where user_id = $1
`, userID).Scan(&profile.UserID, &profile.Level, &profile.Experience, &profile.ReviewCount, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return GrowthProfile{}, false, nil
	}
	if err != nil {
		return GrowthProfile{}, false, err
	}
	profile.UpdatedAt = updatedAt.Format(time.RFC3339)
	points := 0
	_ = r.db.QueryRowContext(ctx, `select available_points from points_accounts where user_id = $1`, userID).Scan(&points)
	profile.AvailablePoints = points
	profile.CreditScore = 100
	profile.TodayCreditScore = 100
	return profile, true, nil
}

func (r *SQLRepository) SaveGrowthProfile(ctx context.Context, profile GrowthProfile) (GrowthProfile, error) {
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, `
insert into user_growth_profiles (user_id, level, experience, review_count, updated_at)
values ($1,$2,$3,$4,now())
on conflict (user_id) do update set
  level = excluded.level,
  experience = excluded.experience,
  review_count = excluded.review_count,
  updated_at = now()
returning updated_at
`, profile.UserID, profile.Level, profile.Experience, profile.ReviewCount).Scan(&updatedAt)
	if err != nil {
		return GrowthProfile{}, err
	}
	_, err = r.db.ExecContext(ctx, `
insert into points_accounts (user_id, available_points, total_earned_points, updated_at)
values ($1,$2,$2,now())
on conflict (user_id) do update set
  available_points = excluded.available_points,
  total_earned_points = greatest(points_accounts.total_earned_points, excluded.total_earned_points),
  updated_at = now()
`, profile.UserID, profile.AvailablePoints)
	if err != nil {
		return GrowthProfile{}, err
	}
	profile.UpdatedAt = updatedAt.Format(time.RFC3339)
	return profile, nil
}

func (r *SQLRepository) AddExperienceLog(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, createdAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
insert into experience_logs (user_id, game_id, change_value, reason, created_at)
values ($1,$2,$3,$4,$5)
`, userID, nullInt64(gameID), changeValue, reason, createdAt)
	return err
}

func (r *SQLRepository) AddPointsLog(ctx context.Context, userID int64, gameID int64, changeValue int, reason string, createdAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
insert into points_logs (user_id, game_id, change_value, reason, created_at)
values ($1,$2,$3,$4,$5)
`, userID, nullInt64(gameID), changeValue, reason, createdAt)
	return err
}

func (r *SQLRepository) AddFootprint(ctx context.Context, footprint Footprint) (Footprint, error) {
	return scanFootprint(r.db.QueryRowContext(ctx, `
insert into user_footprints (user_id, game_id, action, created_at)
values ($1,$2,$3,$4)
returning user_id, game_id, action, created_at
`, footprint.UserID, footprint.GameID, footprint.Action, footprint.CreatedAt))
}

func (r *SQLRepository) ListFootprintsByUser(ctx context.Context, userID int64) ([]Footprint, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, game_id, action, created_at
from user_footprints
where user_id = $1
order by created_at desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFootprints(rows)
}

func (r *SQLRepository) ListFootprints(ctx context.Context) ([]Footprint, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, game_id, action, created_at
from user_footprints
order by created_at desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFootprints(rows)
}

func (r *SQLRepository) GetTodayCredit(ctx context.Context, userID int64, now time.Time) (int, error) {
	scoreDate := dateOnly(now)
	var score int
	err := r.db.QueryRowContext(ctx, `
insert into daily_credit_scores (user_id, score_date, current_score, created_at, updated_at)
values ($1,$2,100,now(),now())
on conflict (user_id, score_date) do update set current_score = daily_credit_scores.current_score
returning current_score
`, userID, scoreDate).Scan(&score)
	return score, err
}

func (r *SQLRepository) SaveTodayCredit(ctx context.Context, userID int64, now time.Time, score int) error {
	_, err := r.db.ExecContext(ctx, `
insert into daily_credit_scores (user_id, score_date, current_score, created_at, updated_at)
values ($1,$2,$3,now(),now())
on conflict (user_id, score_date) do update set current_score = excluded.current_score, updated_at = now()
`, userID, dateOnly(now), score)
	return err
}

func (r *SQLRepository) CreditDeductionValue(ctx context.Context, ruleCode string) (int, bool, error) {
	var change int
	err := r.db.QueryRowContext(ctx, `
select change_value
from credit_deduction_rules
where rule_code = $1 and enabled = true
`, ruleCode).Scan(&change)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return change, true, nil
}

func (r *SQLRepository) ListCreditDeductionRules(ctx context.Context) ([]CreditDeductionRule, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, rule_code, change_value, enabled, coalesce(description, ''), created_at, updated_at
from credit_deduction_rules
order by id asc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]CreditDeductionRule, 0)
	for rows.Next() {
		var item CreditDeductionRule
		if err := rows.Scan(&item.ID, &item.RuleCode, &item.ChangeValue, &item.Enabled, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *SQLRepository) UpsertCreditDeductionRule(ctx context.Context, rule CreditDeductionRule) (CreditDeductionRule, error) {
	return scanCreditDeductionRule(r.db.QueryRowContext(ctx, `
insert into credit_deduction_rules (rule_code, change_value, enabled, description, created_at, updated_at)
values ($1,$2,$3,$4,now(),now())
on conflict (rule_code) do update set
  change_value = excluded.change_value,
  enabled = excluded.enabled,
  description = excluded.description,
  updated_at = now()
returning id, rule_code, change_value, enabled, coalesce(description, ''), created_at, updated_at
`, rule.RuleCode, rule.ChangeValue, rule.Enabled, rule.Description))
}

func (r *SQLRepository) AddCreditLog(ctx context.Context, log CreditLog) (CreditLog, error) {
	return scanCreditLog(r.db.QueryRowContext(ctx, `
insert into credit_logs (user_id, game_id, change_value, before_score, after_score, reason, created_at)
values ($1,$2,$3,$4,$5,$6,$7)
returning id, user_id, game_id, change_value, before_score, after_score, reason, created_at
`, log.UserID, nullInt64(log.GameID), log.ChangeValue, log.BeforeScore, log.AfterScore, log.Reason, log.CreatedAt))
}

func (r *SQLRepository) ListCreditLogsByUser(ctx context.Context, userID int64) ([]CreditLog, error) {
	rows, err := r.db.QueryContext(ctx, creditLogSelect()+` where user_id = $1 order by created_at desc, id desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCreditLogs(rows)
}

func (r *SQLRepository) ListCreditLogsByGame(ctx context.Context, gameID int64) ([]CreditLog, error) {
	rows, err := r.db.QueryContext(ctx, creditLogSelect()+` where game_id = $1 order by created_at desc, id desc`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCreditLogs(rows)
}

func (r *SQLRepository) EnsureAchievement(ctx context.Context, achievement Achievement) (Achievement, bool, error) {
	_, err := r.db.ExecContext(ctx, `
insert into achievements (code, title, created_at)
values ($1,$2,now())
on conflict (code) do update set title = excluded.title
`, achievement.Code, achievement.Title)
	if err != nil {
		return Achievement{}, false, err
	}
	var created bool
	err = r.db.QueryRowContext(ctx, `
insert into user_achievements (user_id, achievement_code, achieved_at)
values ($1,$2,$3)
on conflict (user_id, achievement_code) do nothing
returning true
`, achievement.UserID, achievement.Code, achievement.AchievedAt).Scan(&created)
	if errors.Is(err, sql.ErrNoRows) {
		return achievement, false, nil
	}
	return achievement, created, err
}

func (r *SQLRepository) ListAchievementsByUser(ctx context.Context, userID int64) ([]Achievement, error) {
	rows, err := r.db.QueryContext(ctx, `
select ua.user_id, ua.achievement_code, coalesce(a.title, ua.achievement_code), ua.achieved_at
from user_achievements ua
left join achievements a on a.code = ua.achievement_code
where ua.user_id = $1
order by ua.achieved_at desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAchievements(rows)
}

func (r *SQLRepository) ListAchievements(ctx context.Context) ([]Achievement, error) {
	rows, err := r.db.QueryContext(ctx, `
select ua.user_id, ua.achievement_code, coalesce(a.title, ua.achievement_code), ua.achieved_at
from user_achievements ua
left join achievements a on a.code = ua.achievement_code
order by ua.achieved_at desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAchievements(rows)
}

func reviewSelect() string {
	return `select id, game_id, reviewer_user_id, target_user_id, target_role, score, content, coalesce(tags, '[]'::jsonb), again_intent, created_at from reviews `
}

func creditLogSelect() string {
	return `select id, user_id, game_id, change_value, before_score, after_score, reason, created_at from credit_logs`
}

func scanReviews(rows *sql.Rows) ([]Review, error) {
	items := make([]Review, 0)
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanReview(row interface {
	Scan(dest ...any) error
}) (Review, error) {
	var item Review
	var content sql.NullString
	var againIntent sql.NullString
	var tags []byte
	err := row.Scan(&item.ID, &item.GameID, &item.ReviewerUserID, &item.TargetUserID, &item.TargetRole, &item.Score, &content, &tags, &againIntent, &item.CreatedAt)
	item.Content = content.String
	_ = json.Unmarshal(tags, &item.Tags)
	item.AgainIntent = againIntent.String
	return item, err
}

func scanFootprints(rows *sql.Rows) ([]Footprint, error) {
	items := make([]Footprint, 0)
	for rows.Next() {
		item, err := scanFootprint(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanFootprint(row interface {
	Scan(dest ...any) error
}) (Footprint, error) {
	var item Footprint
	err := row.Scan(&item.UserID, &item.GameID, &item.Action, &item.CreatedAt)
	return item, err
}

func scanCreditLogs(rows *sql.Rows) ([]CreditLog, error) {
	items := make([]CreditLog, 0)
	for rows.Next() {
		item, err := scanCreditLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanCreditLog(row interface {
	Scan(dest ...any) error
}) (CreditLog, error) {
	var item CreditLog
	var gameID sql.NullInt64
	err := row.Scan(&item.ID, &item.UserID, &gameID, &item.ChangeValue, &item.BeforeScore, &item.AfterScore, &item.Reason, &item.CreatedAt)
	item.GameID = gameID.Int64
	return item, err
}

func scanCreditDeductionRule(row interface {
	Scan(dest ...any) error
}) (CreditDeductionRule, error) {
	var item CreditDeductionRule
	err := row.Scan(&item.ID, &item.RuleCode, &item.ChangeValue, &item.Enabled, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func scanAchievements(rows *sql.Rows) ([]Achievement, error) {
	items := make([]Achievement, 0)
	for rows.Next() {
		var item Achievement
		if err := rows.Scan(&item.UserID, &item.Code, &item.Title, &item.AchievedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}
