package games

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

func (r *SQLRepository) CreateGame(ctx context.Context, game Game) (Game, error) {
	tagsJSON, _ := json.Marshal(game.Tags)
	completionRulesJSON, _ := json.Marshal(game.CompletionRules)
	return scanGame(r.db.QueryRowContext(ctx, `
insert into games (
  creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, description, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, updated_at, reject_reason
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$34,$35)
returning id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, description, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason
`, game.CreatorUserID, nullInt64(game.MainGuideUserID), game.Title, game.GameType, game.GameSource, game.Status,
		nullString(game.CoverImage), nullString(game.Description), nullString(game.Highlights), nullString(game.Notice),
		nullString(game.Audience), nullString(game.Participation), game.Price, nullString(game.ProfitTemplate),
		nullString(game.StartAt), nullString(game.EndAt), nullString(game.SignupStartAt), nullString(game.SignupEndAt),
		string(tagsJSON), string(completionRulesJSON),
		nullString(game.PrimaryCategory), nullString(game.PrimaryCategoryText), nullString(game.SecondaryCategory), nullString(game.SecondaryCategoryText), nullString(game.Type),
		game.MinPlayers, game.MaxPlayers, game.CurrentPlayers, game.CityCode, game.CityName,
		nullString(game.Address), game.Longitude, game.Latitude, game.CreatedAt, nullString(game.RejectReason)))
}

func (r *SQLRepository) UpdateGame(ctx context.Context, game Game) (Game, error) {
	tagsJSON, _ := json.Marshal(game.Tags)
	completionRulesJSON, _ := json.Marshal(game.CompletionRules)
	return scanGame(r.db.QueryRowContext(ctx, `
update games set
  main_guide_user_id = $2,
  title = $3,
  game_type = $4,
  game_source = $5,
  status = $6,
  cover_image = $7,
  description = $8,
  highlights = $9,
  notice = $10,
  audience = $11,
  participation = $12,
  price = $13,
  profit_template = $14,
  start_at = $15,
  end_at = $16,
  signup_start_at = $17,
  signup_end_at = $18,
  tags = $19,
  completion_rules = $20,
  primary_category = $21,
  primary_category_text = $22,
  secondary_category = $23,
  secondary_category_text = $24,
  type = $25,
  min_players = $26,
  max_players = $27,
  current_players = $28,
  city_code = $29,
  city_name = $30,
  address = $31,
  longitude = $32,
  latitude = $33,
  reject_reason = $34,
  updated_at = now()
where id = $1
returning id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, description, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason
`, game.ID, nullInt64(game.MainGuideUserID), game.Title, game.GameType, game.GameSource, game.Status,
		nullString(game.CoverImage), nullString(game.Description), nullString(game.Highlights), nullString(game.Notice),
		nullString(game.Audience), nullString(game.Participation), game.Price, nullString(game.ProfitTemplate),
		nullString(game.StartAt), nullString(game.EndAt), nullString(game.SignupStartAt), nullString(game.SignupEndAt),
		string(tagsJSON), string(completionRulesJSON),
		nullString(game.PrimaryCategory), nullString(game.PrimaryCategoryText), nullString(game.SecondaryCategory), nullString(game.SecondaryCategoryText), nullString(game.Type),
		game.MinPlayers, game.MaxPlayers, game.CurrentPlayers, game.CityCode, game.CityName,
		nullString(game.Address), game.Longitude, game.Latitude, nullString(game.RejectReason)))
}

func (r *SQLRepository) GetGame(ctx context.Context, gameID int64) (Game, error) {
	game, err := scanGame(r.db.QueryRowContext(ctx, `
select id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, description, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason
from games
where id = $1
`, gameID))
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrGameNotFound
	}
	return game, err
}

func (r *SQLRepository) ListGames(ctx context.Context) ([]Game, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, description, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason
from games
order by created_at desc, id desc
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGames(rows)
}

func (r *SQLRepository) CountGamesCreatedToday(ctx context.Context, userID int64, now time.Time) (int, error) {
	var count int
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	err := r.db.QueryRowContext(ctx, `
select count(*) from games
where creator_user_id = $1 and created_at >= $2 and created_at < $3
`, userID, start, end).Scan(&count)
	return count, err
}

func (r *SQLRepository) AddMember(ctx context.Context, gameID int64, userID int64, role string) error {
	if role == "" {
		role = "member"
	}
	_, err := r.db.ExecContext(ctx, `
insert into game_members (game_id, user_id, role, status, quit_reason, credit_deducted, credit_log_id, joined_at)
values ($1,$2,$3,'active',null,false,null,now())
on conflict (game_id, user_id) do update set status = 'active', role = excluded.role, quit_reason = null, credit_deducted = false, credit_log_id = null
`, gameID, userID, role)
	return err
}

func (r *SQLRepository) DeleteMember(ctx context.Context, gameID int64, userID int64, status string, reason string) error {
	if status == "" {
		status = "quit"
	}
	_, err := r.db.ExecContext(ctx, `
update game_members
set status = $3,
  quit_reason = $4
where game_id = $1 and user_id = $2
`, gameID, userID, status, reason)
	return err
}

func (r *SQLRepository) UpdateMemberExitCredit(ctx context.Context, gameID int64, userID int64, creditDeducted bool, creditLogID int64) error {
	_, err := r.db.ExecContext(ctx, `
update game_members
set credit_deducted = $3,
  credit_log_id = $4
where game_id = $1 and user_id = $2
`, gameID, userID, creditDeducted, nullInt64(creditLogID))
	return err
}

func (r *SQLRepository) ListMembers(ctx context.Context, gameID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id from game_members
where game_id = $1 and status = 'active'
order by joined_at asc, user_id asc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		items = append(items, userID)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ListMemberRoles(ctx context.Context, gameID int64) ([]MemberRole, error) {
	rows, err := r.db.QueryContext(ctx, `
select user_id, role from game_members
where game_id = $1 and status in ('active', 'canceled')
order by joined_at asc, user_id asc
`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]MemberRole, 0)
	for rows.Next() {
		var item MemberRole
		if err := rows.Scan(&item.UserID, &item.Role); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) CreateApplication(ctx context.Context, application Application) (Application, error) {
	fileIDs, _ := json.Marshal(application.FileIDs)
	return scanApplication(r.db.QueryRowContext(ctx, `
insert into game_applications (game_id, user_id, role, status, reason, reject_reason, file_ids, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
`, application.GameID, application.UserID, application.Role, application.Status, application.Reason, nullString(application.RejectReason), string(fileIDs), application.CreatedAt))
}

func (r *SQLRepository) UpdateApplication(ctx context.Context, application Application) (Application, error) {
	fileIDs, _ := json.Marshal(application.FileIDs)
	return scanApplication(r.db.QueryRowContext(ctx, `
update game_applications set
  status = $2,
  reason = $3,
  reject_reason = $4,
  file_ids = $5,
  reviewed_at = case when $2::varchar = 'pending' then reviewed_at else now() end
where id = $1
returning id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
`, application.ID, application.Status, application.Reason, nullString(application.RejectReason), string(fileIDs)))
}

func (r *SQLRepository) GetApplication(ctx context.Context, applicationID int64) (Application, error) {
	app, err := scanApplication(r.db.QueryRowContext(ctx, `
select id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
from game_applications
where id = $1
`, applicationID))
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrApplicationNotFound
	}
	return app, err
}

func (r *SQLRepository) ListApplicationsByUser(ctx context.Context, userID int64) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
from game_applications
where user_id = $1
order by created_at desc, id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *SQLRepository) ListApplicationsForCreator(ctx context.Context, creatorUserID int64) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
select a.id, a.game_id, a.user_id, a.role, a.status, a.reason, a.reject_reason, a.file_ids, a.created_at
from game_applications a
join games g on g.id = a.game_id
where g.creator_user_id = $1
order by a.created_at desc, a.id desc
`, creatorUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *SQLRepository) ListApplicationsForReviewer(ctx context.Context, reviewerUserID int64) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
select a.id, a.game_id, a.user_id, a.role, a.status, a.reason, a.reject_reason, a.file_ids, a.created_at
from game_applications a
join games g on g.id = a.game_id
where g.creator_user_id = $1 or g.main_guide_user_id = $1
order by a.created_at desc, a.id desc
`, reviewerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApplications(rows)
}

func (r *SQLRepository) PendingApplicationExists(ctx context.Context, gameID int64, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
select exists(
  select 1 from game_applications
  where game_id = $1 and user_id = $2 and status = 'pending'
)
`, gameID, userID).Scan(&exists)
	return exists, err
}

func (r *SQLRepository) CreateInvitation(ctx context.Context, invitation Invitation) (Invitation, error) {
	return scanInvitation(r.db.QueryRowContext(ctx, `
insert into game_invitations (
  invite_group_id, game_id, inviter_user_id, target_user_id, player_user_id, expert_user_id, role, status, message,
  service_type, service_duration_text, demand_detail, budget_amount_cent, expected_time_text,
  application_id, created_at, responded_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
returning id, invite_group_id, game_id, inviter_user_id, target_user_id, player_user_id, expert_user_id, role, status, message,
  service_type, service_duration_text, demand_detail, budget_amount_cent, expected_time_text,
  application_id, created_at, responded_at
`, nullString(invitation.InviteGroupID), invitation.GameID, invitation.InviterID, invitation.TargetUserID, nullInt64(invitation.PlayerUserID), nullInt64(invitation.ExpertUserID), invitation.Role, invitation.Status, nullString(invitation.Message),
		nullString(invitation.ServiceType), nullString(invitation.ServiceDuration), nullString(invitation.DemandDetail), invitation.BudgetAmountCent, nullString(invitation.ExpectedTime),
		nullInt64(invitation.ApplicationID), invitation.CreatedAt, nullTimeString(invitation.RespondedAt)))
}

func (r *SQLRepository) UpdateInvitation(ctx context.Context, invitation Invitation) (Invitation, error) {
	return scanInvitation(r.db.QueryRowContext(ctx, `
update game_invitations set
  status = $2,
  role = $3,
  message = $4,
  service_type = $5,
  service_duration_text = $6,
  demand_detail = $7,
  budget_amount_cent = $8,
  expected_time_text = $9,
  application_id = $10,
  responded_at = $11,
  player_user_id = $12,
  expert_user_id = $13,
  invite_group_id = $14
where id = $1
returning id, invite_group_id, game_id, inviter_user_id, target_user_id, player_user_id, expert_user_id, role, status, message,
  service_type, service_duration_text, demand_detail, budget_amount_cent, expected_time_text,
  application_id, created_at, responded_at
`, invitation.ID, invitation.Status, nullString(invitation.Role), nullString(invitation.Message),
		nullString(invitation.ServiceType), nullString(invitation.ServiceDuration), nullString(invitation.DemandDetail), invitation.BudgetAmountCent, nullString(invitation.ExpectedTime),
		nullInt64(invitation.ApplicationID), nullTimeString(invitation.RespondedAt), nullInt64(invitation.PlayerUserID), nullInt64(invitation.ExpertUserID), nullString(invitation.InviteGroupID)))
}

func (r *SQLRepository) GetInvitation(ctx context.Context, invitationID int64) (Invitation, error) {
	invitation, err := scanInvitation(r.db.QueryRowContext(ctx, `
select id, invite_group_id, game_id, inviter_user_id, target_user_id, player_user_id, expert_user_id, role, status, message,
  service_type, service_duration_text, demand_detail, budget_amount_cent, expected_time_text,
  application_id, created_at, responded_at
from game_invitations
where id = $1
`, invitationID))
	if errors.Is(err, sql.ErrNoRows) {
		return Invitation{}, ErrInvitationNotFound
	}
	return invitation, err
}

func (r *SQLRepository) ListInvitationsForUser(ctx context.Context, userID int64) ([]Invitation, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, invite_group_id, game_id, inviter_user_id, target_user_id, player_user_id, expert_user_id, role, status, message,
  service_type, service_duration_text, demand_detail, budget_amount_cent, expected_time_text,
  application_id, created_at, responded_at
from game_invitations
where inviter_user_id = $1 or target_user_id = $1 or player_user_id = $1 or expert_user_id = $1
order by created_at desc, id desc
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Invitation, 0)
	for rows.Next() {
		item, scanErr := scanInvitation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanGames(rows *sql.Rows) ([]Game, error) {
	items := make([]Game, 0)
	for rows.Next() {
		item, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanGame(row interface {
	Scan(dest ...any) error
}) (Game, error) {
	var game Game
	var mainGuideUserID sql.NullInt64
	var primaryCategory sql.NullString
	var primaryCategoryText sql.NullString
	var secondaryCategory sql.NullString
	var secondaryCategoryText sql.NullString
	var gameTypeDetail sql.NullString
	var coverImage sql.NullString
	var description sql.NullString
	var highlights sql.NullString
	var notice sql.NullString
	var audience sql.NullString
	var participation sql.NullString
	var profitTemplate sql.NullString
	var startAt sql.NullString
	var endAt sql.NullString
	var signupStartAt sql.NullString
	var signupEndAt sql.NullString
	var tagsJSON []byte
	var completionRulesJSON []byte
	var cityCode sql.NullString
	var cityName sql.NullString
	var address sql.NullString
	var longitude sql.NullFloat64
	var latitude sql.NullFloat64
	var rejectReason sql.NullString
	err := row.Scan(
		&game.ID,
		&game.CreatorUserID,
		&mainGuideUserID,
		&game.Title,
		&game.GameType,
		&game.GameSource,
		&game.Status,
		&coverImage,
		&description,
		&highlights,
		&notice,
		&audience,
		&participation,
		&game.Price,
		&profitTemplate,
		&startAt,
		&endAt,
		&signupStartAt,
		&signupEndAt,
		&tagsJSON,
		&completionRulesJSON,
		&primaryCategory,
		&primaryCategoryText,
		&secondaryCategory,
		&secondaryCategoryText,
		&gameTypeDetail,
		&game.MinPlayers,
		&game.MaxPlayers,
		&game.CurrentPlayers,
		&cityCode,
		&cityName,
		&address,
		&longitude,
		&latitude,
		&game.CreatedAt,
		&rejectReason,
	)
	if err != nil {
		return Game{}, err
	}
	game.MainGuideUserID = mainGuideUserID.Int64
	game.RejectReason = rejectReason.String
	game.CoverImage = coverImage.String
	game.Description = description.String
	game.Highlights = highlights.String
	game.Notice = notice.String
	game.Audience = audience.String
	game.Participation = participation.String
	game.ProfitTemplate = profitTemplate.String
	game.StartAt = startAt.String
	game.EndAt = endAt.String
	game.SignupStartAt = signupStartAt.String
	game.SignupEndAt = signupEndAt.String
	if len(tagsJSON) > 0 {
		_ = json.Unmarshal(tagsJSON, &game.Tags)
	}
	if len(completionRulesJSON) > 0 {
		_ = json.Unmarshal(completionRulesJSON, &game.CompletionRules)
	}
	game.PrimaryCategory = primaryCategory.String
	game.PrimaryCategoryText = primaryCategoryText.String
	game.SecondaryCategory = secondaryCategory.String
	game.SecondaryCategoryText = secondaryCategoryText.String
	game.Type = gameTypeDetail.String
	game.CityCode = cityCode.String
	game.CityName = cityName.String
	game.Address = address.String
	game.Longitude = longitude.Float64
	game.Latitude = latitude.Float64
	return game, nil
}

func scanApplications(rows *sql.Rows) ([]Application, error) {
	items := make([]Application, 0)
	for rows.Next() {
		item, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanApplication(row interface {
	Scan(dest ...any) error
}) (Application, error) {
	var application Application
	var reason sql.NullString
	var rejectReason sql.NullString
	var rawFileIDs []byte
	if err := row.Scan(
		&application.ID,
		&application.GameID,
		&application.UserID,
		&application.Role,
		&application.Status,
		&reason,
		&rejectReason,
		&rawFileIDs,
		&application.CreatedAt,
	); err != nil {
		return Application{}, err
	}
	application.Reason = reason.String
	application.RejectReason = rejectReason.String
	if len(rawFileIDs) > 0 {
		_ = json.Unmarshal(rawFileIDs, &application.FileIDs)
	}
	return application, nil
}

func scanInvitation(row interface {
	Scan(dest ...any) error
}) (Invitation, error) {
	var invitation Invitation
	var inviteGroupID sql.NullString
	var playerUserID sql.NullInt64
	var expertUserID sql.NullInt64
	var role sql.NullString
	var message sql.NullString
	var serviceType sql.NullString
	var serviceDuration sql.NullString
	var demandDetail sql.NullString
	var expectedTime sql.NullString
	var applicationID sql.NullInt64
	var respondedAt sql.NullTime
	if err := row.Scan(
		&invitation.ID,
		&inviteGroupID,
		&invitation.GameID,
		&invitation.InviterID,
		&invitation.TargetUserID,
		&playerUserID,
		&expertUserID,
		&role,
		&invitation.Status,
		&message,
		&serviceType,
		&serviceDuration,
		&demandDetail,
		&invitation.BudgetAmountCent,
		&expectedTime,
		&applicationID,
		&invitation.CreatedAt,
		&respondedAt,
	); err != nil {
		return Invitation{}, err
	}
	invitation.InviteGroupID = inviteGroupID.String
	invitation.PlayerUserID = playerUserID.Int64
	invitation.ExpertUserID = expertUserID.Int64
	invitation.Role = role.String
	invitation.Message = message.String
	invitation.ServiceType = serviceType.String
	invitation.ServiceDuration = serviceDuration.String
	invitation.DemandDetail = demandDetail.String
	invitation.ExpectedTime = expectedTime.String
	invitation.ApplicationID = applicationID.Int64
	if respondedAt.Valid {
		invitation.RespondedAt = respondedAt.Time.Format(time.RFC3339)
	}
	return invitation, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullTimeString(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}
