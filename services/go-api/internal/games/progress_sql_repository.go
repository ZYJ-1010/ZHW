package games

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type SQLProgressRepository struct {
	db *sql.DB
}

func NewSQLProgressRepository(db *sql.DB) *SQLProgressRepository {
	return &SQLProgressRepository{db: db}
}

func (r *SQLProgressRepository) CreateProgressFeedback(ctx context.Context, feedback ProgressFeedback) (ProgressFeedback, error) {
	fileIDs, err := json.Marshal(feedback.FileIDs)
	if err != nil {
		return ProgressFeedback{}, err
	}
	return scanProgressFeedback(r.db.QueryRowContext(ctx, `
insert into game_progress_feedbacks (game_id, user_id, progress, content, file_ids, created_at)
values ($1,$2,$3,$4,$5,$6)
returning id, game_id, user_id, progress, content, file_ids, created_at
`, feedback.GameID, feedback.UserID, feedback.Progress, nullString(feedback.Content), string(fileIDs), feedback.CreatedAt))
}

func (r *SQLProgressRepository) ListProgressFeedbacks(ctx context.Context, gameID int64) ([]ProgressFeedback, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, game_id, user_id, progress, content, file_ids, created_at
from game_progress_feedbacks
where game_id = $1
order by created_at asc, id asc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ProgressFeedback, 0)
	for rows.Next() {
		item, err := scanProgressFeedback(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLProgressRepository) CreateMilestone(ctx context.Context, milestone Milestone) (Milestone, error) {
	return scanMilestone(r.db.QueryRowContext(ctx, `
insert into game_milestones (game_id, title, status, created_at, updated_at)
values ($1,$2,$3,$4,$4)
returning id, game_id, title, status, created_at
`, milestone.GameID, milestone.Title, milestone.Status, milestone.CreatedAt))
}

func (r *SQLProgressRepository) UpdateMilestone(ctx context.Context, milestone Milestone) (Milestone, error) {
	return scanMilestone(r.db.QueryRowContext(ctx, `
update game_milestones set title = $2, status = $3, updated_at = now()
where id = $1
returning id, game_id, title, status, created_at
`, milestone.ID, milestone.Title, milestone.Status))
}

func (r *SQLProgressRepository) ListMilestones(ctx context.Context, gameID int64) ([]Milestone, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, game_id, title, status, created_at
from game_milestones
where game_id = $1
order by created_at asc, id asc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Milestone, 0)
	for rows.Next() {
		item, err := scanMilestone(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLProgressRepository) CreateCheckin(ctx context.Context, checkin Checkin) (Checkin, error) {
	fileIDs, err := json.Marshal(checkin.FileIDs)
	if err != nil {
		return Checkin{}, err
	}
	return scanCheckin(r.db.QueryRowContext(ctx, `
insert into game_checkins (game_id, user_id, milestone_id, checkin_type, content, file_ids, status, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id, game_id, user_id, milestone_id, checkin_type, content, file_ids, status, created_at
`, checkin.GameID, checkin.UserID, nullInt64(checkin.MilestoneID), checkin.CheckinType, checkin.Content, string(fileIDs), checkin.Status, checkin.CreatedAt))
}

func (r *SQLProgressRepository) UpdateCheckinStatus(ctx context.Context, checkinID int64, status string) (Checkin, error) {
	item, err := scanCheckin(r.db.QueryRowContext(ctx, `
update game_checkins set status = $2
where id = $1
returning id, game_id, user_id, milestone_id, checkin_type, content, file_ids, status, created_at
`, checkinID, status))
	if errors.Is(err, sql.ErrNoRows) {
		return Checkin{}, ErrCheckinNotFound
	}
	return item, err
}

func (r *SQLProgressRepository) ListCheckins(ctx context.Context, gameID int64, includeInvalid bool) ([]Checkin, error) {
	query := `
select id, game_id, user_id, milestone_id, checkin_type, content, file_ids, status, created_at
from game_checkins
where game_id = $1`
	if !includeInvalid {
		query += ` and status <> 'invalid'`
	}
	query += ` order by created_at asc, id asc`
	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Checkin, 0)
	for rows.Next() {
		item, err := scanCheckin(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLProgressRepository) CreateRetrospective(ctx context.Context, retrospective Retrospective) (Retrospective, error) {
	return scanRetrospective(r.db.QueryRowContext(ctx, `
insert into game_retrospectives (game_id, user_id, content, again_intent, created_at)
values ($1,$2,$3,$4,$5)
returning id, game_id, user_id, content, again_intent, created_at
`, retrospective.GameID, retrospective.UserID, retrospective.Content, retrospective.AgainIntent, retrospective.CreatedAt))
}

func (r *SQLProgressRepository) RetrospectiveExists(ctx context.Context, gameID int64, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
select exists(select 1 from game_retrospectives where game_id = $1 and user_id = $2)
`, gameID, userID).Scan(&exists)
	return exists, err
}

func (r *SQLProgressRepository) ListRetrospectives(ctx context.Context, gameID int64) ([]Retrospective, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, game_id, user_id, content, again_intent, created_at
from game_retrospectives
where game_id = $1
order by created_at asc, id asc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Retrospective, 0)
	for rows.Next() {
		item, err := scanRetrospective(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLProgressRepository) CreateContinueDraft(ctx context.Context, record ContinueDraftRecord) (ContinueDraftRecord, error) {
	return scanContinueDraft(r.db.QueryRowContext(ctx, `
insert into game_continue_drafts (original_game_id, draft_game_id, creator_user_id, title, status, created_at)
values ($1,$2,$3,$4,$5,$6)
returning original_game_id, draft_game_id, creator_user_id, title, status, created_at
`, record.OriginalGameID, record.DraftGameID, record.CreatorUserID, record.Title, record.Status, record.CreatedAt))
}

func (r *SQLProgressRepository) ListContinueDrafts(ctx context.Context, gameID int64) ([]ContinueDraftRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
select original_game_id, draft_game_id, creator_user_id, title, status, created_at
from game_continue_drafts
where original_game_id = $1
order by created_at desc, draft_game_id desc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ContinueDraftRecord, 0)
	for rows.Next() {
		item, err := scanContinueDraft(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanMilestone(row interface {
	Scan(dest ...any) error
}) (Milestone, error) {
	var item Milestone
	err := row.Scan(&item.ID, &item.GameID, &item.Title, &item.Status, &item.CreatedAt)
	return item, err
}

func scanProgressFeedback(row interface {
	Scan(dest ...any) error
}) (ProgressFeedback, error) {
	var item ProgressFeedback
	var content sql.NullString
	var fileIDs []byte
	err := row.Scan(&item.ID, &item.GameID, &item.UserID, &item.Progress, &content, &fileIDs, &item.CreatedAt)
	if err != nil {
		return ProgressFeedback{}, err
	}
	item.Content = content.String
	if len(fileIDs) > 0 {
		_ = json.Unmarshal(fileIDs, &item.FileIDs)
	}
	return item, nil
}

func scanCheckin(row interface {
	Scan(dest ...any) error
}) (Checkin, error) {
	var item Checkin
	var milestoneID sql.NullInt64
	var content sql.NullString
	var rawFileIDs []byte
	err := row.Scan(&item.ID, &item.GameID, &item.UserID, &milestoneID, &item.CheckinType, &content, &rawFileIDs, &item.Status, &item.CreatedAt)
	if err != nil {
		return Checkin{}, err
	}
	item.MilestoneID = milestoneID.Int64
	item.Content = content.String
	if len(rawFileIDs) > 0 {
		_ = json.Unmarshal(rawFileIDs, &item.FileIDs)
	}
	return item, nil
}

func scanRetrospective(row interface {
	Scan(dest ...any) error
}) (Retrospective, error) {
	var item Retrospective
	var againIntent sql.NullString
	err := row.Scan(&item.ID, &item.GameID, &item.UserID, &item.Content, &againIntent, &item.CreatedAt)
	if err != nil {
		return Retrospective{}, err
	}
	item.AgainIntent = againIntent.String
	return item, nil
}

func scanContinueDraft(row interface {
	Scan(dest ...any) error
}) (ContinueDraftRecord, error) {
	var item ContinueDraftRecord
	err := row.Scan(&item.OriginalGameID, &item.DraftGameID, &item.CreatorUserID, &item.Title, &item.Status, &item.CreatedAt)
	return item, err
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}
