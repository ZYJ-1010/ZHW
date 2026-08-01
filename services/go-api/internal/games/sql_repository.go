package games

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/lib/pq"
)

type SQLRepository struct {
	db *sql.DB
}

var _ statusTransitionRepository = (*SQLRepository)(nil)
var _ gameCreatorWithFreeOrderRepository = (*SQLRepository)(nil)
var _ cancelGameTransitionRepository = (*SQLRepository)(nil)

func (r *SQLRepository) SaveStatusLog(ctx context.Context, log StatusLog) (StatusLog, error) {
	return saveStatusLogWithQueryer(ctx, r.db, log)
}

func saveStatusLogWithQueryer(ctx context.Context, queryer gameQueryer, log StatusLog) (StatusLog, error) {
	var saved StatusLog
	var operatorID sql.NullInt64
	var reason sql.NullString
	err := queryer.QueryRowContext(ctx, `
insert into game_status_logs (game_id, from_status, to_status, operator_user_id, reason, created_at)
values ($1,$2,$3,$4,$5,$6)
returning id, game_id, from_status, to_status, operator_user_id, reason, created_at
	`, log.GameID, log.FromStatus, log.ToStatus, nullInt64(log.OperatorID), nullString(log.Reason), log.CreatedAt).Scan(&saved.ID, &saved.GameID, &saved.FromStatus, &saved.ToStatus, &operatorID, &reason, &saved.CreatedAt)
	saved.OperatorID = operatorID.Int64
	saved.Reason = reason.String
	return saved, err
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

type gameQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// CreateGameWithCreator keeps the game and its mandatory creator membership
// atomic. A game without that member cannot be managed or started correctly.
func (r *SQLRepository) CreateGameWithCreator(ctx context.Context, game Game, creatorRole string) (Game, error) {
	return r.createGameWithCreator(ctx, game, creatorRole, false)
}

// CreateGameWithCreatorAndFreeOrder is used by the app creation flow. The
// phase-one order is free and never invokes WeChat Pay, but it is still part of
// the creation contract exposed by the order APIs.
func (r *SQLRepository) CreateGameWithCreatorAndFreeOrder(ctx context.Context, game Game, creatorRole string) (Game, error) {
	return r.createGameWithCreator(ctx, game, creatorRole, true)
}

func (r *SQLRepository) createGameWithCreator(ctx context.Context, game Game, creatorRole string, createFreeOrder bool) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, err
	}
	defer tx.Rollback()
	saved, err := createGameWithQueryer(ctx, tx, game)
	if err != nil {
		return Game{}, err
	}
	if creatorRole == "" {
		creatorRole = "member"
	}
	if _, err := tx.ExecContext(ctx, `
insert into game_members (game_id, user_id, role, status, quit_reason, credit_deducted, credit_log_id, joined_at)
values ($1,$2,$3,'active',null,false,null,now())
on conflict (game_id, user_id) do update set status = 'active', role = excluded.role, quit_reason = null, credit_deducted = false, credit_log_id = null
	`, saved.ID, saved.CreatorUserID, creatorRole); err != nil {
		return Game{}, err
	}
	if createFreeOrder {
		orderNo := "FREE-" + strconv.FormatInt(saved.ID, 10)
		result, err := tx.ExecContext(ctx, `
insert into payment_orders (order_no, user_id, game_id, amount_cent, pay_status)
values ($1,$2,$3,0,'free_no_pay')
on conflict (order_no) do update set
  updated_at = payment_orders.updated_at
where payment_orders.user_id = excluded.user_id
  and payment_orders.game_id = excluded.game_id
		`, orderNo, saved.CreatorUserID, saved.ID)
		if err != nil {
			return Game{}, err
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			if err != nil {
				return Game{}, err
			}
			return Game{}, sql.ErrNoRows
		}
	}
	if err := tx.Commit(); err != nil {
		return Game{}, err
	}
	return saved, nil
}

func (r *SQLRepository) CreateGame(ctx context.Context, game Game) (Game, error) {
	return createGameWithQueryer(ctx, r.db, game)
}

func createGameWithQueryer(ctx context.Context, queryer gameQueryer, game Game) (Game, error) {
	tagsJSON, _ := json.Marshal(game.Tags)
	completionRulesJSON, _ := json.Marshal(game.CompletionRules)
	allowedRolesJSON, _ := json.Marshal(normalizedAllowedRoles(game.AllowedRoles))
	descriptionMediaJSON, _ := json.Marshal(game.DescriptionMedia)
	return scanGame(queryer.QueryRowContext(ctx, `
insert into games (
  creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, updated_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$37,$38,$39,$40,$41,$42,$43,$44)
returning id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
`, game.CreatorUserID, nullInt64(game.MainGuideUserID), game.Title, game.GameType, game.GameSource, game.Status,
		nullString(game.CoverImage), nullString(game.Introduction), nullString(game.Description), string(descriptionMediaJSON), nullString(game.Highlights), nullString(game.Notice),
		nullString(game.Audience), nullString(game.Participation), game.Price, nullString(game.ProfitTemplate),
		nullString(game.StartAt), nullString(game.EndAt), nullString(game.SignupStartAt), nullString(game.SignupEndAt),
		string(tagsJSON), string(completionRulesJSON), string(allowedRolesJSON),
		nullString(game.PrimaryCategory), nullString(game.PrimaryCategoryText), nullString(game.SecondaryCategory), nullString(game.SecondaryCategoryText), nullString(game.Type),
		game.MinPlayers, game.MaxPlayers, game.CurrentPlayers, game.CityCode, game.CityName,
		nullString(game.Address), game.Longitude, game.Latitude, game.CreatedAt, nullString(game.RejectReason), nullString(game.StartReason), nullInt64(game.StartedByUserID), nullTimeString(game.StartedAt), game.AllowGuideEscort, game.DistributionMethod, game.PaymentStatus))
}

func (r *SQLRepository) UpdateGame(ctx context.Context, game Game) (Game, error) {
	return updateGameWithQueryer(ctx, r.db, game, "")
}

func updateGameWithQueryer(ctx context.Context, queryer gameQueryer, game Game, expectedStatus string) (Game, error) {
	tagsJSON, _ := json.Marshal(game.Tags)
	completionRulesJSON, _ := json.Marshal(game.CompletionRules)
	allowedRolesJSON, _ := json.Marshal(normalizedAllowedRoles(game.AllowedRoles))
	descriptionMediaJSON, _ := json.Marshal(game.DescriptionMedia)
	return scanGame(queryer.QueryRowContext(ctx, `
update games set
  main_guide_user_id = $2,
  title = $3,
  game_type = $4,
  game_source = $5,
  status = $6,
  cover_image = $7,
  introduction = $8,
  description = $9,
  description_media = $10,
  highlights = $11,
  notice = $12,
  audience = $13,
  participation = $14,
  price = $15,
  profit_template = $16,
  start_at = $17,
  end_at = $18,
  signup_start_at = $19,
  signup_end_at = $20,
  tags = $21,
  completion_rules = $22,
  allowed_roles = $23,
  primary_category = $24,
  primary_category_text = $25,
  secondary_category = $26,
  secondary_category_text = $27,
  type = $28,
  min_players = $29,
  max_players = $30,
  current_players = $31,
  city_code = $32,
  city_name = $33,
  address = $34,
  longitude = $35,
  latitude = $36,
  reject_reason = $37,
	start_reason = $38,
	started_by_user_id = $39,
	started_at = $40,
	allow_guide_escort = $41,
	distribution_method = $42,
	payment_status = $43,
  updated_at = now()
where id = $1 and ($44 = '' or status = $44)
returning id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
`, game.ID, nullInt64(game.MainGuideUserID), game.Title, game.GameType, game.GameSource, game.Status,
		nullString(game.CoverImage), nullString(game.Introduction), nullString(game.Description), string(descriptionMediaJSON), nullString(game.Highlights), nullString(game.Notice),
		nullString(game.Audience), nullString(game.Participation), game.Price, nullString(game.ProfitTemplate),
		nullString(game.StartAt), nullString(game.EndAt), nullString(game.SignupStartAt), nullString(game.SignupEndAt),
		string(tagsJSON), string(completionRulesJSON), string(allowedRolesJSON),
		nullString(game.PrimaryCategory), nullString(game.PrimaryCategoryText), nullString(game.SecondaryCategory), nullString(game.SecondaryCategoryText), nullString(game.Type),
		game.MinPlayers, game.MaxPlayers, game.CurrentPlayers, game.CityCode, game.CityName,
		nullString(game.Address), game.Longitude, game.Latitude, nullString(game.RejectReason), nullString(game.StartReason), nullInt64(game.StartedByUserID), nullTimeString(game.StartedAt), game.AllowGuideEscort, game.DistributionMethod, game.PaymentStatus, expectedStatus))
}

// UpdateGameWithStatusLog commits the lifecycle row and its audit log as one
// transaction. The expected source status also prevents a stale process from
// recording a transition after another process has already moved the game.
func (r *SQLRepository) UpdateGameWithStatusLog(ctx context.Context, game Game, statusLog StatusLog) (Game, StatusLog, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, StatusLog{}, err
	}
	defer tx.Rollback()
	savedGame, err := updateGameWithQueryer(ctx, tx, game, statusLog.FromStatus)
	if err != nil {
		return Game{}, StatusLog{}, err
	}
	savedLog, err := saveStatusLogWithQueryer(ctx, tx, statusLog)
	if err != nil {
		return Game{}, StatusLog{}, err
	}
	if err := tx.Commit(); err != nil {
		return Game{}, StatusLog{}, err
	}
	return savedGame, savedLog, nil
}

// CancelGameWithStatusLog commits every database mutation caused by service
// cancellation together. Updating members first is safe because a stale game
// status or a log failure rolls the entire transaction back.
func (r *SQLRepository) CancelGameWithStatusLog(ctx context.Context, game Game, reason string, statusLog StatusLog) (Game, StatusLog, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, StatusLog{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
update game_members
set status = 'canceled', quit_reason = $2
where game_id = $1 and status = 'active'
	`, game.ID, reason); err != nil {
		return Game{}, StatusLog{}, err
	}
	savedGame, err := updateGameWithQueryer(ctx, tx, game, statusLog.FromStatus)
	if err != nil {
		return Game{}, StatusLog{}, err
	}
	savedLog, err := saveStatusLogWithQueryer(ctx, tx, statusLog)
	if err != nil {
		return Game{}, StatusLog{}, err
	}
	if err := tx.Commit(); err != nil {
		return Game{}, StatusLog{}, err
	}
	return savedGame, savedLog, nil
}

// UpdateGameAfterApproval atomically reserves one player slot. The expected
// count predicate makes concurrent reviewers fail instead of overwriting the
// same capacity value.
func (r *SQLRepository) UpdateGameAfterApproval(ctx context.Context, game Game, expectedCurrentPlayers int) (Game, error) {
	result, err := r.db.ExecContext(ctx, `
update games
set main_guide_user_id = $2,
    current_players = $3,
    status = $4,
    updated_at = now()
where id = $1 and status = 'recruiting' and current_players = $5
`, game.ID, nullInt64(game.MainGuideUserID), game.CurrentPlayers, game.Status, expectedCurrentPlayers)
	if err != nil {
		return Game{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Game{}, sql.ErrNoRows
	}
	return r.GetGame(ctx, game.ID)
}

// ApproveApplication commits the game capacity change, member admission,
// application result and optional full-status log in one transaction. These
// records describe one approval event and must not become visible separately.
func (r *SQLRepository) ApproveApplication(ctx context.Context, game Game, expectedCurrentPlayers int, application Application, memberRole string, statusLog *StatusLog) (Game, Application, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, Application{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
update games
set main_guide_user_id = $2,
    current_players = $3,
    status = $4,
    updated_at = now()
where id = $1 and status = 'recruiting' and current_players = $5
`, game.ID, nullInt64(game.MainGuideUserID), game.CurrentPlayers, game.Status, expectedCurrentPlayers)
	if err != nil {
		return Game{}, Application{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Game{}, Application{}, sql.ErrNoRows
	}

	if memberRole == "" {
		memberRole = "member"
	}
	if _, err := tx.ExecContext(ctx, `
insert into game_members (game_id, user_id, role, status, quit_reason, credit_deducted, credit_log_id, joined_at)
values ($1,$2,$3,'active',null,false,null,now())
on conflict (game_id, user_id) do update set status = 'active', role = excluded.role, quit_reason = null, credit_deducted = false, credit_log_id = null
`, game.ID, application.UserID, memberRole); err != nil {
		return Game{}, Application{}, err
	}

	fileIDs, _ := json.Marshal(application.FileIDs)
	savedApplication, err := scanApplication(tx.QueryRowContext(ctx, `
update game_applications set
  status = $2,
  reason = $3,
  reject_reason = $4,
  file_ids = $5,
  reviewed_at = now()
where id = $1
  and status = 'pending'
returning id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
`, application.ID, application.Status, application.Reason, nullString(application.RejectReason), string(fileIDs)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Game{}, Application{}, ErrApplicationNotPending
		}
		return Game{}, Application{}, err
	}

	if statusLog != nil {
		if _, err := tx.ExecContext(ctx, `
insert into game_status_logs (game_id, from_status, to_status, operator_user_id, reason, created_at)
values ($1,$2,$3,$4,$5,$6)
`, statusLog.GameID, statusLog.FromStatus, statusLog.ToStatus, nullInt64(statusLog.OperatorID), nullString(statusLog.Reason), statusLog.CreatedAt); err != nil {
			return Game{}, Application{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Game{}, Application{}, err
	}
	// game is the exact row state written by the transaction. Returning it
	// directly avoids reporting a false failure if a follow-up SELECT has a
	// transient error after the transaction has already committed.
	return game, savedApplication, nil
}

func (r *SQLRepository) ExitGame(ctx context.Context, gameID int64, userID int64, memberStatus string, reason string) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, err
	}
	defer tx.Rollback()
	var currentPlayers int
	var status string
	if err := tx.QueryRowContext(ctx, `select current_players, status from games where id = $1 for update`, gameID).Scan(&currentPlayers, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Game{}, ErrGameNotFound
		}
		return Game{}, err
	}
	result, err := tx.ExecContext(ctx, `
update game_members set status = $3, quit_reason = $4
where game_id = $1 and user_id = $2 and status = 'active'
`, gameID, userID, memberStatus, nullString(reason))
	if err != nil {
		return Game{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Game{}, ErrForbidden
	}
	if currentPlayers > 0 {
		currentPlayers--
	}
	if _, err := tx.ExecContext(ctx, `update games set current_players = $2, updated_at = now() where id = $1`, gameID, currentPlayers); err != nil {
		return Game{}, err
	}
	if err := tx.Commit(); err != nil {
		return Game{}, err
	}
	return r.GetGame(ctx, gameID)
}

// ExitGameWithCredit commits the member exit, capacity decrement, permanent
// credit account and credit ledger in one PostgreSQL transaction. The game row
// and the user's credit account are both locked before any mutation is applied.
func (r *SQLRepository) ExitGameWithCredit(ctx context.Context, gameID int64, userID int64, memberStatus string, reason string, mutation ExitCreditMutation) (ExitCreditResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ExitCreditResult{}, err
	}
	defer tx.Rollback()
	if memberStatus == "" {
		memberStatus = "quit"
	}
	if reason == "" {
		reason = "quit_after_started"
	}
	now := time.Now()
	var currentPlayers int
	var ignoredStatus string
	if err := tx.QueryRowContext(ctx, `select current_players, status from games where id = $1 for update`, gameID).Scan(&currentPlayers, &ignoredStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExitCreditResult{}, ErrGameNotFound
		}
		return ExitCreditResult{}, err
	}
	var currentMemberStatus string
	if err := tx.QueryRowContext(ctx, `select status from game_members where game_id = $1 and user_id = $2 for update`, gameID, userID).Scan(&currentMemberStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExitCreditResult{}, ErrForbidden
		}
		return ExitCreditResult{}, err
	}
	if currentMemberStatus != "active" {
		return ExitCreditResult{}, ErrForbidden
	}
	if mutation.Reason != "" {
		reason = mutation.Reason
	}
	if mutation.ChangeValue > 0 {
		mutation.ChangeValue = 0
	}
	if mutation.InitialScore <= 0 {
		mutation.InitialScore = 100
	}
	var beforeScore int
	if err := tx.QueryRowContext(ctx, `
insert into credit_accounts (user_id, current_score, updated_at)
values ($1,$2,now())
on conflict (user_id) do update set current_score = credit_accounts.current_score
returning current_score
`, userID, mutation.InitialScore).Scan(&beforeScore); err != nil {
		return ExitCreditResult{}, err
	}
	// The INSERT ... RETURNING above locks the conflicting row for the
	// statement; explicitly lock it again for clarity and concurrent updates.
	if err := tx.QueryRowContext(ctx, `select current_score from credit_accounts where user_id = $1 for update`, userID).Scan(&beforeScore); err != nil {
		return ExitCreditResult{}, err
	}
	afterScore := beforeScore + mutation.ChangeValue
	if afterScore < 0 {
		afterScore = 0
	}
	if _, err := tx.ExecContext(ctx, `
update credit_accounts set current_score = $2, updated_at = now()
where user_id = $1
`, userID, afterScore); err != nil {
		return ExitCreditResult{}, err
	}
	var creditLogID int64
	creditDeducted := afterScore < beforeScore
	if creditDeducted {
		if err := tx.QueryRowContext(ctx, `
insert into credit_logs (user_id, game_id, change_value, before_score, after_score, reason, created_at)
values ($1,$2,$3,$4,$5,$6,$7)
returning id
`, userID, nullInt64(gameID), afterScore-beforeScore, beforeScore, afterScore, reason, now).Scan(&creditLogID); err != nil {
			return ExitCreditResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `
insert into user_footprints (user_id, game_id, action, created_at)
values ($1,$2,'credit_deducted',$3)
`, userID, nullInt64(gameID), now); err != nil {
			return ExitCreditResult{}, err
		}
	}
	result, err := tx.ExecContext(ctx, `
update game_members
set status = $3, quit_reason = $4, credit_deducted = $5, credit_log_id = $6
where game_id = $1 and user_id = $2 and status = 'active'
`, gameID, userID, memberStatus, reason, creditDeducted, nullInt64(creditLogID))
	if err != nil {
		return ExitCreditResult{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ExitCreditResult{}, ErrForbidden
	}
	if currentPlayers > 0 {
		currentPlayers--
	}
	if _, err := tx.ExecContext(ctx, `update games set current_players = $2, updated_at = now() where id = $1`, gameID, currentPlayers); err != nil {
		return ExitCreditResult{}, err
	}
	game, err := scanGame(tx.QueryRowContext(ctx, gameSelectSQL()+` where id = $1`, gameID))
	if err != nil {
		return ExitCreditResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExitCreditResult{}, err
	}
	return ExitCreditResult{
		ExitResult:  ExitResult{Game: game, GameID: gameID, UserID: userID, Reason: reason, CreditDeduct: creditDeducted, CreditDeducted: creditDeducted, CreditLogID: creditLogID, MemberStatus: memberStatus},
		CreditLogID: creditLogID,
		BeforeScore: beforeScore,
		AfterScore:  afterScore,
		ChangeValue: afterScore - beforeScore,
		CreatedAt:   now,
	}, nil
}

func gameSelectSQL() string {
	return `select id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
from games`
}

func (r *SQLRepository) GetGame(ctx context.Context, gameID int64) (Game, error) {
	game, err := scanGame(r.db.QueryRowContext(ctx, `
select id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
from games
where id = $1
`, gameID))
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrGameNotFound
	}
	return game, err
}

// LockGame is used by approval paths that must serialize capacity checks with
// other reviewers. The caller keeps the surrounding database transaction open
// when composing multiple mutations.
func (r *SQLRepository) LockGame(ctx context.Context, gameID int64) (Game, error) {
	game, err := scanGame(r.db.QueryRowContext(ctx, `
select id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
from games
where id = $1
for update
`, gameID))
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrGameNotFound
	}
	return game, err
}

func (r *SQLRepository) ListGames(ctx context.Context) ([]Game, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, creator_user_id, main_guide_user_id, title, game_type, game_source, status,
  cover_image, introduction, description, description_media, highlights, notice, audience, participation, price, profit_template,
  start_at, end_at, signup_start_at, signup_end_at, tags, completion_rules, allowed_roles,
  primary_category, primary_category_text, secondary_category, secondary_category_text, type,
  min_players, max_players, current_players, city_code, city_name, address,
  longitude, latitude, created_at, reject_reason, start_reason, started_by_user_id, started_at, allow_guide_escort, distribution_method, payment_status
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

// StatsForUser reads participation from PostgreSQL instead of the service's
// process-local cache. Role growth and invitation conversion therefore remain
// stable after an API restart and after a game enters a later terminal state.
func (r *SQLRepository) StatsForUser(ctx context.Context, userID int64) (UserStats, error) {
	stats := UserStats{UserID: userID}
	err := r.db.QueryRowContext(ctx, `
select
  count(*)::int,
  count(*) filter (where g.status in ('pending_review','completed','settling','closed'))::int
from game_members gm
join games g on g.id = gm.game_id
where gm.user_id = $1 and gm.status = 'active'
`, userID).Scan(&stats.Participated, &stats.Completed)
	return stats, err
}

func (r *SQLRepository) ParticipatedBetween(ctx context.Context, userID int64, start time.Time, end time.Time) (bool, error) {
	var participated bool
	err := r.db.QueryRowContext(ctx, `
select exists(
  select 1
  from game_members
  where user_id = $1 and joined_at >= $2 and joined_at < $3
)
`, userID, start, end).Scan(&participated)
	return participated, err
}

func (r *SQLRepository) CreateApplication(ctx context.Context, application Application) (Application, error) {
	fileIDs, _ := json.Marshal(application.FileIDs)
	saved, err := scanApplication(r.db.QueryRowContext(ctx, `
insert into game_applications (game_id, user_id, role, status, reason, reject_reason, file_ids, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
`, application.GameID, application.UserID, application.Role, application.Status, application.Reason, nullString(application.RejectReason), string(fileIDs), application.CreatedAt))
	if isUniqueViolation(err) {
		return Application{}, ErrAlreadyApplied
	}
	return saved, err
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
	and status = 'pending'
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
	saved, err := scanInvitation(r.db.QueryRowContext(ctx, `
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
	if isUniqueViolation(err) {
		return Invitation{}, ErrAlreadyInvited
	}
	return saved, err
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

// AcceptInvitation creates the pending application and marks the invitation
// accepted in one transaction so neither record can get ahead of the other.
func (r *SQLRepository) AcceptInvitation(ctx context.Context, invitation Invitation, application Application) (Invitation, Application, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Invitation{}, Application{}, err
	}
	defer tx.Rollback()
	fileIDs, _ := json.Marshal(application.FileIDs)
	savedApplication, err := scanApplication(tx.QueryRowContext(ctx, `
insert into game_applications (game_id, user_id, role, status, reason, reject_reason, file_ids, created_at)
values ($1,$2,$3,$4,$5,$6,$7,$8)
returning id, game_id, user_id, role, status, reason, reject_reason, file_ids, created_at
`, application.GameID, application.UserID, application.Role, application.Status, application.Reason, nullString(application.RejectReason), string(fileIDs), application.CreatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return Invitation{}, Application{}, ErrAlreadyApplied
		}
		return Invitation{}, Application{}, err
	}
	invitation.ApplicationID = savedApplication.ID
	savedInvitation, err := scanInvitation(tx.QueryRowContext(ctx, `
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
where id = $1 and status = 'pending'
returning id, invite_group_id, game_id, inviter_user_id, target_user_id, player_user_id, expert_user_id, role, status, message,
  service_type, service_duration_text, demand_detail, budget_amount_cent, expected_time_text,
  application_id, created_at, responded_at
`, invitation.ID, invitation.Status, nullString(invitation.Role), nullString(invitation.Message),
		nullString(invitation.ServiceType), nullString(invitation.ServiceDuration), nullString(invitation.DemandDetail), invitation.BudgetAmountCent, nullString(invitation.ExpectedTime),
		nullInt64(invitation.ApplicationID), nullTimeString(invitation.RespondedAt), nullInt64(invitation.PlayerUserID), nullInt64(invitation.ExpertUserID), nullString(invitation.InviteGroupID)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invitation{}, Application{}, ErrInvitationNotPending
		}
		return Invitation{}, Application{}, err
	}
	if err := tx.Commit(); err != nil {
		return Invitation{}, Application{}, err
	}
	return savedInvitation, savedApplication, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
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
	var introduction sql.NullString
	var description sql.NullString
	var descriptionMediaJSON []byte
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
	var allowedRolesJSON []byte
	var cityCode sql.NullString
	var cityName sql.NullString
	var address sql.NullString
	var longitude sql.NullFloat64
	var latitude sql.NullFloat64
	var rejectReason sql.NullString
	var startReason sql.NullString
	var startedByUserID sql.NullInt64
	var startedAt sql.NullTime
	err := row.Scan(
		&game.ID,
		&game.CreatorUserID,
		&mainGuideUserID,
		&game.Title,
		&game.GameType,
		&game.GameSource,
		&game.Status,
		&coverImage,
		&introduction,
		&description,
		&descriptionMediaJSON,
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
		&allowedRolesJSON,
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
		&startReason,
		&startedByUserID,
		&startedAt,
		&game.AllowGuideEscort,
		&game.DistributionMethod,
		&game.PaymentStatus,
	)
	if err != nil {
		return Game{}, err
	}
	game.MainGuideUserID = mainGuideUserID.Int64
	game.RejectReason = rejectReason.String
	game.StartReason = startReason.String
	game.StartedByUserID = startedByUserID.Int64
	if startedAt.Valid {
		game.StartedAt = startedAt.Time.Format(time.RFC3339)
	}
	game.CoverImage = coverImage.String
	game.Introduction = introduction.String
	game.Description = description.String
	if len(descriptionMediaJSON) > 0 {
		_ = json.Unmarshal(descriptionMediaJSON, &game.DescriptionMedia)
	}
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
	if len(allowedRolesJSON) > 0 {
		_ = json.Unmarshal(allowedRolesJSON, &game.AllowedRoles)
	}
	game.AllowedRoles = normalizedAllowedRoles(game.AllowedRoles)
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
