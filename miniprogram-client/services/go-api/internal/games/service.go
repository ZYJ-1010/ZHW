package games

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	ErrRealnameRequired       = errors.New("realname required")
	ErrInvalidGameInput       = errors.New("invalid game input")
	ErrInvalidGameType        = errors.New("invalid game type")
	ErrInvalidPlayers         = errors.New("invalid players")
	ErrDailyLimit             = errors.New("daily limit reached")
	ErrGameNotFound           = errors.New("game not found")
	ErrGameNotRecruiting      = errors.New("game not recruiting")
	ErrGameNotStartable       = errors.New("game not startable")
	ErrAlreadyApplied         = errors.New("already applied")
	ErrAlreadyMember          = errors.New("already member")
	ErrApplicationNotFound    = errors.New("application not found")
	ErrApplicationNotPending  = errors.New("application not pending")
	ErrFull                   = errors.New("game full")
	ErrForbidden              = errors.New("forbidden")
	ErrGameNotConfirmable     = errors.New("game not confirmable")
	ErrInvalidProgress        = errors.New("invalid progress")
	ErrInvalidMilestone       = errors.New("invalid milestone")
	ErrInvalidCheckin         = errors.New("invalid checkin")
	ErrInvalidRetrospective   = errors.New("invalid retrospective")
	ErrDuplicateRetrospective = errors.New("duplicate retrospective")
	ErrCheckinNotFound        = errors.New("checkin not found")
	ErrMilestoneNotFound      = errors.New("milestone not found")
	ErrInvitationNotFound     = errors.New("invitation not found")
	ErrInvitationNotPending   = errors.New("invitation not pending")
)

const (
	MinGamePlayers = 5
	MaxGamePlayers = 8
)

type IdentityChecker interface {
	IsVerified(userID int64) bool
}

type Game struct {
	ID                    int64     `json:"id"`
	CreatorUserID         int64     `json:"creatorUserId"`
	MainGuideUserID       int64     `json:"mainGuideUserId,omitempty"`
	Title                 string    `json:"title"`
	GameType              string    `json:"gameType"`
	CoverImage            string    `json:"coverImage,omitempty"`
	Description           string    `json:"description,omitempty"`
	Highlights            string    `json:"highlights,omitempty"`
	Notice                string    `json:"notice,omitempty"`
	Audience              string    `json:"audience,omitempty"`
	Participation         string    `json:"participation,omitempty"`
	Price                 float64   `json:"price,omitempty"`
	ProfitTemplate        string    `json:"profitTemplate,omitempty"`
	StartAt               string    `json:"startAt,omitempty"`
	EndAt                 string    `json:"endAt,omitempty"`
	SignupStartAt         string    `json:"signupStartAt,omitempty"`
	SignupEndAt           string    `json:"signupEndAt,omitempty"`
	Tags                  []string  `json:"tags,omitempty"`
	CompletionRules       []string  `json:"completionRules,omitempty"`
	PrimaryCategory       string    `json:"primaryCategory,omitempty"`
	PrimaryCategoryText   string    `json:"primaryCategoryText,omitempty"`
	SecondaryCategory     string    `json:"secondaryCategory,omitempty"`
	SecondaryCategoryText string    `json:"secondaryCategoryText,omitempty"`
	Type                  string    `json:"type,omitempty"`
	GameSource            string    `json:"gameSource"`
	Status                string    `json:"status"`
	MinPlayers            int       `json:"minPlayers"`
	MaxPlayers            int       `json:"maxPlayers"`
	CurrentPlayers        int       `json:"currentPlayers"`
	CityCode              string    `json:"cityCode,omitempty"`
	CityName              string    `json:"cityName,omitempty"`
	Address               string    `json:"address,omitempty"`
	Longitude             float64   `json:"longitude,omitempty"`
	Latitude              float64   `json:"latitude,omitempty"`
	DistanceMeter         float64   `json:"distanceMeter,omitempty"`
	DistanceLabel         string    `json:"distanceLabel,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

type Application struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"gameId"`
	UserID    int64     `json:"userId"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	FileIDs   []int64   `json:"fileIds,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Invitation struct {
	ID            int64     `json:"id"`
	GameID        int64     `json:"gameId"`
	InviterID     int64     `json:"inviterUserId"`
	TargetUserID  int64     `json:"targetUserId"`
	Status        string    `json:"status"`
	Message       string    `json:"message,omitempty"`
	ApplicationID int64     `json:"applicationId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	RespondedAt   string    `json:"respondedAt,omitempty"`
}

type ServiceConfirm struct {
	ID          int64     `json:"id"`
	GameID      int64     `json:"gameId"`
	Status      string    `json:"status"`
	ConfirmedBy []int64   `json:"confirmedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt string    `json:"completedAt,omitempty"`
}

type ServiceConfirmItem struct {
	ID        int64     `json:"id"`
	ConfirmID int64     `json:"confirmId"`
	GameID    int64     `json:"gameId"`
	UserID    int64     `json:"userId"`
	Note      string    `json:"note,omitempty"`
	FileIDs   []int64   `json:"fileIds,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type ExitResult struct {
	Game           Game   `json:"game"`
	GameID         int64  `json:"gameId"`
	UserID         int64  `json:"userId"`
	Reason         string `json:"reason"`
	CreditDeduct   bool   `json:"creditDeduct"`
	CreditDeducted bool   `json:"creditDeducted"`
	CreditLogID    int64  `json:"creditLogId,omitempty"`
	MemberStatus   string `json:"memberStatus"`
}

type ProgressFeedback struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"gameId"`
	UserID    int64     `json:"userId"`
	Progress  int       `json:"progress"`
	Content   string    `json:"content,omitempty"`
	FileIDs   []int64   `json:"fileIds,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type ProgressFeedbackRequest struct {
	Progress int     `json:"progress"`
	Content  string  `json:"content"`
	FileIDs  []int64 `json:"fileIds"`
}

type Milestone struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"gameId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type MilestoneRequest struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

type Checkin struct {
	ID          int64     `json:"id"`
	GameID      int64     `json:"gameId"`
	UserID      int64     `json:"userId"`
	MilestoneID int64     `json:"milestoneId,omitempty"`
	CheckinType string    `json:"checkinType"`
	Content     string    `json:"content,omitempty"`
	FileIDs     []int64   `json:"fileIds,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CheckinRequest struct {
	MilestoneID int64   `json:"milestoneId"`
	CheckinType string  `json:"checkinType"`
	Content     string  `json:"content"`
	FileIDs     []int64 `json:"fileIds"`
}

type Retrospective struct {
	ID          int64     `json:"id"`
	GameID      int64     `json:"gameId"`
	UserID      int64     `json:"userId"`
	Content     string    `json:"content"`
	AgainIntent string    `json:"againIntent,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RetrospectiveRequest struct {
	Content     string `json:"content"`
	AgainIntent string `json:"againIntent"`
}

type ContinueDraftRequest struct {
	Title string `json:"title"`
}

type Favorite struct {
	UserID    int64     `json:"userId"`
	GameID    int64     `json:"gameId"`
	Game      Game      `json:"game"`
	CreatedAt time.Time `json:"createdAt"`
}

type FavoriteRepository interface {
	SaveFavorite(ctx context.Context, favorite Favorite) (Favorite, error)
	DeleteFavorite(ctx context.Context, userID int64, gameID int64) error
	ListFavoritesByUser(ctx context.Context, userID int64) ([]Favorite, error)
	ListAllFavorites(ctx context.Context) ([]Favorite, error)
}

type Repository interface {
	CreateGame(ctx context.Context, game Game) (Game, error)
	UpdateGame(ctx context.Context, game Game) (Game, error)
	GetGame(ctx context.Context, gameID int64) (Game, error)
	ListGames(ctx context.Context) ([]Game, error)
	CountGamesCreatedToday(ctx context.Context, userID int64, now time.Time) (int, error)
	AddMember(ctx context.Context, gameID int64, userID int64, role string) error
	DeleteMember(ctx context.Context, gameID int64, userID int64, status string, reason string) error
	UpdateMemberExitCredit(ctx context.Context, gameID int64, userID int64, creditDeducted bool, creditLogID int64) error
	ListMembers(ctx context.Context, gameID int64) ([]int64, error)
	CreateApplication(ctx context.Context, application Application) (Application, error)
	UpdateApplication(ctx context.Context, application Application) (Application, error)
	GetApplication(ctx context.Context, applicationID int64) (Application, error)
	ListApplicationsByUser(ctx context.Context, userID int64) ([]Application, error)
	ListApplicationsForCreator(ctx context.Context, creatorUserID int64) ([]Application, error)
	PendingApplicationExists(ctx context.Context, gameID int64, userID int64) (bool, error)
	CreateInvitation(ctx context.Context, invitation Invitation) (Invitation, error)
	UpdateInvitation(ctx context.Context, invitation Invitation) (Invitation, error)
	GetInvitation(ctx context.Context, invitationID int64) (Invitation, error)
}

type ProgressRepository interface {
	CreateMilestone(ctx context.Context, milestone Milestone) (Milestone, error)
	UpdateMilestone(ctx context.Context, milestone Milestone) (Milestone, error)
	ListMilestones(ctx context.Context, gameID int64) ([]Milestone, error)
	CreateCheckin(ctx context.Context, checkin Checkin) (Checkin, error)
	UpdateCheckinStatus(ctx context.Context, checkinID int64, status string) (Checkin, error)
	ListCheckins(ctx context.Context, gameID int64, includeInvalid bool) ([]Checkin, error)
	CreateRetrospective(ctx context.Context, retrospective Retrospective) (Retrospective, error)
	RetrospectiveExists(ctx context.Context, gameID int64, userID int64) (bool, error)
	ListRetrospectives(ctx context.Context, gameID int64) ([]Retrospective, error)
	CreateContinueDraft(ctx context.Context, record ContinueDraftRecord) (ContinueDraftRecord, error)
	ListContinueDrafts(ctx context.Context, gameID int64) ([]ContinueDraftRecord, error)
}

type ServiceConfirmRepository interface {
	GetConfirm(ctx context.Context, gameID int64) (ServiceConfirm, []ServiceConfirmItem, bool, error)
	SaveConfirm(ctx context.Context, confirm ServiceConfirm) (ServiceConfirm, error)
	SaveConfirmItem(ctx context.Context, item ServiceConfirmItem) (ServiceConfirmItem, error)
	ListConfirmItems(ctx context.Context, gameID int64) ([]ServiceConfirmItem, error)
}

type ContinueDraft struct {
	OriginalGameID int64 `json:"originalGameId"`
	Draft          Game  `json:"draft"`
}

type ContinueDraftRecord struct {
	OriginalGameID int64     `json:"originalGameId"`
	DraftGameID    int64     `json:"draftGameId"`
	CreatorUserID  int64     `json:"creatorUserId"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type UserStats struct {
	UserID       int64   `json:"userId"`
	Participated int     `json:"participatedGames"`
	Completed    int     `json:"completedGames"`
	AverageScore float64 `json:"averageScore"`
}

type CreateRequest struct {
	Title                 string   `json:"title"`
	GameType              string   `json:"gameType"`
	CoverImage            string   `json:"coverImage"`
	Description           string   `json:"description"`
	Highlights            string   `json:"highlights"`
	Notice                string   `json:"notice"`
	Audience              string   `json:"audience"`
	Participation         string   `json:"participation"`
	Price                 float64  `json:"price"`
	ProfitTemplate        string   `json:"profitTemplate"`
	StartAt               string   `json:"startAt"`
	EndAt                 string   `json:"endAt"`
	SignupStartAt         string   `json:"signupStartAt"`
	SignupEndAt           string   `json:"signupEndAt"`
	Tags                  []string `json:"tags"`
	CompletionRules       []string `json:"completionRules"`
	PrimaryCategory       string   `json:"primaryCategory"`
	PrimaryCategoryText   string   `json:"primaryCategoryText"`
	SecondaryCategory     string   `json:"secondaryCategory"`
	SecondaryCategoryText string   `json:"secondaryCategoryText"`
	Type                  string   `json:"type"`
	CreatorUserID         int64    `json:"creatorUserId,omitempty"`
	MainGuideUserID       int64    `json:"mainGuideUserId,omitempty"`
	MinPlayers            int      `json:"minPlayers"`
	MaxPlayers            int      `json:"maxPlayers"`
	CityCode              string   `json:"cityCode"`
	CityName              string   `json:"cityName"`
	Address               string   `json:"address"`
	Longitude             float64  `json:"longitude"`
	Latitude              float64  `json:"latitude"`
}

type InvitationRequest struct {
	TargetUserID int64  `json:"targetUserId"`
	Message      string `json:"message"`
}

type InvitationRespondRequest struct {
	Accept bool   `json:"accept"`
	Reason string `json:"reason"`
}

type ApplyRequest struct {
	Reason  string  `json:"reason"`
	FileIDs []int64 `json:"fileIds"`
}

type Service struct {
	mu                sync.RWMutex
	nextID            int64
	nextApplicationID int64
	nextInvitationID  int64
	games             map[int64]Game
	applications      map[int64]Application
	invitations       map[int64]Invitation
	members           map[int64]map[int64]bool
	nextConfirmID     int64
	nextConfirmItemID int64
	confirms          map[int64]ServiceConfirm
	confirmItems      map[int64]map[int64]ServiceConfirmItem
	nextProgressID    int64
	progressFeedbacks map[int64][]ProgressFeedback
	nextMilestoneID   int64
	milestones        map[int64][]Milestone
	nextCheckinID     int64
	checkins          map[int64][]Checkin
	nextRetroID       int64
	retrospectives    map[int64][]Retrospective
	continueDrafts    map[int64][]ContinueDraftRecord
	favorites         map[int64]map[int64]Favorite
	repo              Repository
	progressRepo      ProgressRepository
	confirmRepo       ServiceConfirmRepository
	favoriteRepo      FavoriteRepository
	identity          IdentityChecker
	dailyCreateLimit  int
}

func NewService(identity IdentityChecker) *Service {
	return NewServiceWithRepositories(identity, nil, nil)
}

func NewServiceWithFavoriteRepository(identity IdentityChecker, favoriteRepo FavoriteRepository) *Service {
	return NewServiceWithRepositories(identity, nil, favoriteRepo)
}

func NewServiceWithRepositories(identity IdentityChecker, repo Repository, favoriteRepo FavoriteRepository) *Service {
	return &Service{
		nextID:            1,
		nextApplicationID: 1,
		nextInvitationID:  1,
		games:             make(map[int64]Game),
		applications:      make(map[int64]Application),
		invitations:       make(map[int64]Invitation),
		members:           make(map[int64]map[int64]bool),
		nextConfirmID:     1,
		nextConfirmItemID: 1,
		confirms:          make(map[int64]ServiceConfirm),
		confirmItems:      make(map[int64]map[int64]ServiceConfirmItem),
		nextProgressID:    1,
		progressFeedbacks: make(map[int64][]ProgressFeedback),
		nextMilestoneID:   1,
		milestones:        make(map[int64][]Milestone),
		nextCheckinID:     1,
		checkins:          make(map[int64][]Checkin),
		nextRetroID:       1,
		retrospectives:    make(map[int64][]Retrospective),
		continueDrafts:    make(map[int64][]ContinueDraftRecord),
		favorites:         make(map[int64]map[int64]Favorite),
		repo:              repo,
		favoriteRepo:      favoriteRepo,
		identity:          identity,
		dailyCreateLimit:  3,
	}
}

func (s *Service) SetDailyCreateLimit(limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 {
		limit = 3
	}
	s.dailyCreateLimit = limit
}

func (s *Service) UseProgressRepository(progressRepo ProgressRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progressRepo = progressRepo
}

func (s *Service) UseServiceConfirmRepository(confirmRepo ServiceConfirmRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.confirmRepo = confirmRepo
}

func (s *Service) Create(userID int64, req CreateRequest) (Game, error) {
	req = normalizeCreateRequest(req)
	req.MainGuideUserID = 0
	if req.GameType != "" && req.GameType != "free" {
		return Game{}, ErrInvalidGameType
	}
	if err := validateCreateRequest(req); err != nil {
		return Game{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.createdTodayLocked(userID) >= s.dailyCreateLimit {
		return Game{}, ErrDailyLimit
	}

	game := newGameFromRequest(s.nextID, userID, req, "free", "app", "pending_audit")
	if s.repo != nil {
		saved, err := s.repo.CreateGame(context.Background(), game)
		if err != nil {
			return Game{}, err
		}
		game = saved
		if err := s.repo.AddMember(context.Background(), game.ID, userID, "creator"); err != nil {
			return Game{}, err
		}
	}
	s.nextID++
	s.games[game.ID] = game
	s.members[game.ID] = map[int64]bool{userID: true}
	return game, nil
}

func (s *Service) CreateFromAdmin(req CreateRequest) (Game, error) {
	req = normalizeCreateRequest(req)
	if req.CreatorUserID <= 0 {
		return Game{}, ErrInvalidGameInput
	}
	if req.MainGuideUserID > 0 && req.MainGuideUserID == req.CreatorUserID {
		return Game{}, ErrInvalidGameInput
	}
	if req.MainGuideUserID > 0 && !s.identity.IsVerified(req.MainGuideUserID) {
		return Game{}, ErrRealnameRequired
	}
	gameType := req.GameType
	if gameType == "" {
		gameType = "free"
	}
	if !validAdminGameType(gameType) {
		return Game{}, ErrInvalidGameType
	}
	if err := validateCreateRequest(req); err != nil {
		return Game{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	game := newGameFromRequest(s.nextID, req.CreatorUserID, req, gameType, "admin", "recruiting")
	if s.repo != nil {
		saved, err := s.repo.CreateGame(context.Background(), game)
		if err != nil {
			return Game{}, err
		}
		game = saved
		if err := s.repo.AddMember(context.Background(), game.ID, req.CreatorUserID, "creator"); err != nil {
			return Game{}, err
		}
		if req.MainGuideUserID > 0 && req.MainGuideUserID != req.CreatorUserID {
			if err := s.repo.AddMember(context.Background(), game.ID, req.MainGuideUserID, "main_guide"); err != nil {
				return Game{}, err
			}
		}
	}
	s.nextID++
	s.games[game.ID] = game
	s.members[game.ID] = map[int64]bool{req.CreatorUserID: true}
	if req.MainGuideUserID > 0 {
		s.members[game.ID][req.MainGuideUserID] = true
	}
	return game, nil
}

func normalizeCreateRequest(req CreateRequest) CreateRequest {
	req.Title = strings.TrimSpace(req.Title)
	req.GameType = strings.TrimSpace(req.GameType)
	req.CoverImage = strings.TrimSpace(req.CoverImage)
	req.Description = strings.TrimSpace(req.Description)
	req.Highlights = strings.TrimSpace(req.Highlights)
	req.Notice = strings.TrimSpace(req.Notice)
	req.Audience = strings.TrimSpace(req.Audience)
	req.Participation = strings.TrimSpace(req.Participation)
	req.ProfitTemplate = strings.TrimSpace(req.ProfitTemplate)
	req.StartAt = strings.TrimSpace(req.StartAt)
	req.EndAt = strings.TrimSpace(req.EndAt)
	req.SignupStartAt = strings.TrimSpace(req.SignupStartAt)
	req.SignupEndAt = strings.TrimSpace(req.SignupEndAt)
	req.Tags = cleanStringList(req.Tags, 20)
	req.CompletionRules = cleanStringList(req.CompletionRules, 20)
	req.PrimaryCategory = strings.TrimSpace(req.PrimaryCategory)
	req.PrimaryCategoryText = strings.TrimSpace(req.PrimaryCategoryText)
	req.SecondaryCategory = strings.TrimSpace(req.SecondaryCategory)
	req.SecondaryCategoryText = strings.TrimSpace(req.SecondaryCategoryText)
	req.Type = strings.TrimSpace(req.Type)
	req.CityCode = strings.TrimSpace(req.CityCode)
	req.CityName = strings.TrimSpace(req.CityName)
	req.Address = strings.TrimSpace(req.Address)
	return req
}

func validateCreateRequest(req CreateRequest) error {
	if req.MinPlayers < MinGamePlayers || req.MaxPlayers > MaxGamePlayers || req.MinPlayers > req.MaxPlayers {
		return ErrInvalidPlayers
	}
	if req.Title == "" || len(req.Title) > 80 || len(req.CityCode) > 32 || len(req.CityName) > 64 || len(req.Address) > 255 {
		return ErrInvalidGameInput
	}
	if len(req.PrimaryCategory) > 32 || len(req.PrimaryCategoryText) > 32 || len(req.SecondaryCategory) > 32 || len(req.SecondaryCategoryText) > 32 || len(req.Type) > 32 {
		return ErrInvalidGameInput
	}
	if len(req.CoverImage) > 500 || len(req.Description) > 1000 || len(req.Highlights) > 500 || len(req.Notice) > 500 || len(req.Audience) > 200 || len(req.Participation) > 64 || len(req.ProfitTemplate) > 64 {
		return ErrInvalidGameInput
	}
	if len(req.StartAt) > 32 || len(req.EndAt) > 32 || len(req.SignupStartAt) > 32 || len(req.SignupEndAt) > 32 {
		return ErrInvalidGameInput
	}
	if req.Price < 0 || len(req.Tags) > 20 || len(req.CompletionRules) > 20 {
		return ErrInvalidGameInput
	}
	if req.MainGuideUserID < 0 {
		return ErrInvalidGameInput
	}
	if !validCoordinate(req.Longitude, req.Latitude) {
		return ErrInvalidGameInput
	}
	return nil
}

func validAdminGameType(value string) bool {
	switch value {
	case "free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition":
		return true
	default:
		return false
	}
}

func newGameFromRequest(id int64, creatorUserID int64, req CreateRequest, gameType string, source string, status string) Game {
	currentPlayers := 1
	if req.MainGuideUserID > 0 && req.MainGuideUserID != creatorUserID {
		currentPlayers = 2
	}
	return Game{
		ID:                    id,
		CreatorUserID:         creatorUserID,
		MainGuideUserID:       req.MainGuideUserID,
		Title:                 req.Title,
		GameType:              gameType,
		CoverImage:            req.CoverImage,
		Description:           req.Description,
		Highlights:            req.Highlights,
		Notice:                req.Notice,
		Audience:              req.Audience,
		Participation:         req.Participation,
		Price:                 req.Price,
		ProfitTemplate:        req.ProfitTemplate,
		StartAt:               req.StartAt,
		EndAt:                 req.EndAt,
		SignupStartAt:         req.SignupStartAt,
		SignupEndAt:           req.SignupEndAt,
		Tags:                  append([]string(nil), req.Tags...),
		CompletionRules:       append([]string(nil), req.CompletionRules...),
		PrimaryCategory:       req.PrimaryCategory,
		PrimaryCategoryText:   req.PrimaryCategoryText,
		SecondaryCategory:     req.SecondaryCategory,
		SecondaryCategoryText: req.SecondaryCategoryText,
		Type:                  req.Type,
		GameSource:            source,
		Status:                status,
		MinPlayers:            req.MinPlayers,
		MaxPlayers:            req.MaxPlayers,
		CurrentPlayers:        currentPlayers,
		CityCode:              req.CityCode,
		CityName:              req.CityName,
		Address:               req.Address,
		Longitude:             req.Longitude,
		Latitude:              req.Latitude,
		CreatedAt:             time.Now(),
	}
}

func cleanStringList(values []string, limit int) []string {
	if limit <= 0 || len(values) == 0 {
		return nil
	}
	capacity := len(values)
	if capacity > limit {
		capacity = limit
	}
	result := make([]string, 0, capacity)
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > 64 || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func validCoordinate(longitude float64, latitude float64) bool {
	if longitude == 0 && latitude == 0 {
		return true
	}
	if math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
		return false
	}
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 {
		return false
	}
	return true
}

func validFileIDs(fileIDs []int64, maxCount int) bool {
	if len(fileIDs) > maxCount {
		return false
	}
	for _, fileID := range fileIDs {
		if fileID <= 0 {
			return false
		}
	}
	return true
}

func validCheckinType(value string) bool {
	switch value {
	case "progress", "complete", "arrival", "proof":
		return true
	default:
		return false
	}
}

func validMilestoneStatus(value string) bool {
	switch value {
	case "pending", "in_progress", "completed", "cancelled":
		return true
	default:
		return false
	}
}

func validAgainIntent(value string) bool {
	switch value {
	case "", "yes", "no", "maybe":
		return true
	default:
		return false
	}
}

func (s *Service) List() []Game {
	if s.repo != nil {
		if items, err := s.repo.ListGames(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Game, 0, len(s.games))
	for _, game := range s.games {
		result = append(result, game)
	}
	return result
}

func (s *Service) SameCity(cityCode string) []Game {
	if s.repo != nil {
		if items, err := s.repo.ListGames(context.Background()); err == nil {
			result := make([]Game, 0)
			for _, game := range items {
				if game.CityCode == cityCode {
					result = append(result, game)
				}
			}
			return result
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Game, 0)
	for _, game := range s.games {
		if game.CityCode == cityCode {
			result = append(result, game)
		}
	}
	return result
}

func (s *Service) ApproveGame(gameID int64) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Game{}, err
	}
	if !ok {
		return Game{}, ErrGameNotFound
	}
	game.Status = "recruiting"
	if s.repo != nil {
		saved, err := s.repo.UpdateGame(context.Background(), game)
		if err != nil {
			return Game{}, err
		}
		game = saved
	}
	s.games[gameID] = game
	return game, nil
}

func (s *Service) Apply(userID int64, gameID int64, req ApplyRequest) (Application, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	if len(req.Reason) > 300 || !validFileIDs(req.FileIDs, 9) {
		return Application{}, ErrInvalidGameInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Application{}, err
	}
	if !ok {
		return Application{}, ErrGameNotFound
	}
	if game.Status != "recruiting" {
		return Application{}, ErrGameNotRecruiting
	}
	if s.memberLocked(gameID, userID) {
		return Application{}, ErrAlreadyMember
	}
	if game.CurrentPlayers >= game.MaxPlayers {
		return Application{}, ErrFull
	}
	for _, app := range s.applications {
		if app.GameID == gameID && app.UserID == userID && app.Status == "pending" {
			return Application{}, ErrAlreadyApplied
		}
	}
	if s.repo != nil {
		exists, err := s.repo.PendingApplicationExists(context.Background(), gameID, userID)
		if err != nil {
			return Application{}, err
		}
		if exists {
			return Application{}, ErrAlreadyApplied
		}
	}
	app := Application{
		ID:        s.nextApplicationID,
		GameID:    gameID,
		UserID:    userID,
		Status:    "pending",
		Reason:    req.Reason,
		FileIDs:   append([]int64(nil), req.FileIDs...),
		CreatedAt: time.Now(),
	}
	if s.repo != nil {
		saved, err := s.repo.CreateApplication(context.Background(), app)
		if err != nil {
			return Application{}, err
		}
		app = saved
	}
	s.nextApplicationID++
	s.applications[app.ID] = app
	return app, nil
}

func (s *Service) CreateInvitation(inviterID int64, gameID int64, req InvitationRequest) (Invitation, error) {
	req.Message = strings.TrimSpace(req.Message)
	if req.TargetUserID <= 0 || req.TargetUserID == inviterID || len(req.Message) > 300 {
		return Invitation{}, ErrInvalidGameInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Invitation{}, err
	}
	if !ok {
		return Invitation{}, ErrGameNotFound
	}
	if game.Status != "recruiting" {
		return Invitation{}, ErrGameNotRecruiting
	}
	if !canInviteGuide(game, inviterID) {
		return Invitation{}, ErrForbidden
	}
	if s.memberLocked(gameID, req.TargetUserID) {
		return Invitation{}, ErrAlreadyMember
	}
	invitation := Invitation{
		ID:           s.nextInvitationID,
		GameID:       gameID,
		InviterID:    inviterID,
		TargetUserID: req.TargetUserID,
		Status:       "pending",
		Message:      req.Message,
		CreatedAt:    time.Now(),
	}
	if s.repo != nil {
		saved, err := s.repo.CreateInvitation(context.Background(), invitation)
		if err != nil {
			return Invitation{}, err
		}
		invitation = saved
	}
	s.nextInvitationID++
	s.invitations[invitation.ID] = invitation
	return invitation, nil
}

func (s *Service) RespondInvitation(userID int64, invitationID int64, req InvitationRespondRequest) (Invitation, Application, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	if invitationID <= 0 || len(req.Reason) > 300 {
		return Invitation{}, Application{}, ErrInvalidGameInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	invitation, ok := s.invitations[invitationID]
	if !ok {
		if s.repo == nil {
			return Invitation{}, Application{}, ErrInvitationNotFound
		}
		saved, err := s.repo.GetInvitation(context.Background(), invitationID)
		if err != nil {
			return Invitation{}, Application{}, err
		}
		invitation = saved
	}
	if invitation.TargetUserID != userID {
		return Invitation{}, Application{}, ErrForbidden
	}
	if invitation.Status != "pending" {
		return Invitation{}, Application{}, ErrInvitationNotPending
	}
	if !req.Accept {
		invitation.Status = "rejected"
		invitation.RespondedAt = time.Now().Format(time.RFC3339)
		if s.repo != nil {
			saved, err := s.repo.UpdateInvitation(context.Background(), invitation)
			if err != nil {
				return Invitation{}, Application{}, err
			}
			invitation = saved
		}
		s.invitations[invitation.ID] = invitation
		return invitation, Application{}, nil
	}
	if !s.identity.IsVerified(userID) {
		return Invitation{}, Application{}, ErrRealnameRequired
	}
	game, ok, err := s.gameLocked(invitation.GameID)
	if err != nil {
		return Invitation{}, Application{}, err
	}
	if !ok {
		return Invitation{}, Application{}, ErrGameNotFound
	}
	if game.Status != "recruiting" {
		return Invitation{}, Application{}, ErrGameNotRecruiting
	}
	if s.memberLocked(invitation.GameID, userID) {
		return Invitation{}, Application{}, ErrAlreadyMember
	}
	if game.CurrentPlayers >= game.MaxPlayers {
		return Invitation{}, Application{}, ErrFull
	}
	for _, app := range s.applications {
		if app.GameID == invitation.GameID && app.UserID == userID && app.Status == "pending" {
			return Invitation{}, Application{}, ErrAlreadyApplied
		}
	}
	if s.repo != nil {
		exists, err := s.repo.PendingApplicationExists(context.Background(), invitation.GameID, userID)
		if err != nil {
			return Invitation{}, Application{}, err
		}
		if exists {
			return Invitation{}, Application{}, ErrAlreadyApplied
		}
	}
	app := Application{
		ID:        s.nextApplicationID,
		GameID:    invitation.GameID,
		UserID:    userID,
		Status:    "pending",
		Reason:    req.Reason,
		CreatedAt: time.Now(),
	}
	if app.Reason == "" {
		app.Reason = "accepted_invitation"
	}
	if s.repo != nil {
		saved, err := s.repo.CreateApplication(context.Background(), app)
		if err != nil {
			return Invitation{}, Application{}, err
		}
		app = saved
	}
	s.nextApplicationID++
	s.applications[app.ID] = app

	invitation.Status = "accepted"
	invitation.ApplicationID = app.ID
	invitation.RespondedAt = time.Now().Format(time.RFC3339)
	if s.repo != nil {
		saved, err := s.repo.UpdateInvitation(context.Background(), invitation)
		if err != nil {
			return Invitation{}, Application{}, err
		}
		invitation = saved
	}
	s.invitations[invitation.ID] = invitation
	return invitation, app, nil
}

func (s *Service) ReviewApplication(operatorUserID int64, applicationID int64, approve bool) (Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.applications[applicationID]
	if !ok {
		if s.repo == nil {
			return Application{}, ErrApplicationNotFound
		}
		saved, err := s.repo.GetApplication(context.Background(), applicationID)
		if err != nil {
			return Application{}, err
		}
		app = saved
	}
	if app.Status != "pending" {
		return Application{}, ErrApplicationNotPending
	}
	game, ok, err := s.gameLocked(app.GameID)
	if err != nil {
		return Application{}, err
	}
	if !ok {
		return Application{}, ErrGameNotFound
	}
	if game.CreatorUserID != operatorUserID {
		return Application{}, ErrForbidden
	}
	if game.Status != "recruiting" {
		return Application{}, ErrGameNotRecruiting
	}
	if approve {
		if game.CurrentPlayers >= game.MaxPlayers {
			return Application{}, ErrFull
		}
		app.Status = "approved"
		if s.members[game.ID] == nil {
			s.members[game.ID] = make(map[int64]bool)
		}
		s.members[game.ID][app.UserID] = true
		memberRole := "member"
		if game.MainGuideUserID == 0 && s.invitationAcceptedByApplicationLocked(app.ID) {
			game.MainGuideUserID = app.UserID
			memberRole = "main_guide"
		}
		if s.repo != nil {
			if err := s.repo.AddMember(context.Background(), game.ID, app.UserID, memberRole); err != nil {
				return Application{}, err
			}
		}
		game.CurrentPlayers++
		if game.CurrentPlayers >= game.MaxPlayers {
			game.Status = "full"
		}
		s.games[game.ID] = game
	} else {
		app.Status = "rejected"
	}
	if s.repo != nil {
		if saved, err := s.repo.UpdateApplication(context.Background(), app); err != nil {
			return Application{}, err
		} else {
			app = saved
		}
		if saved, err := s.repo.UpdateGame(context.Background(), game); err != nil {
			return Application{}, err
		} else {
			game = saved
		}
	}
	s.applications[app.ID] = app
	return app, nil
}

func (s *Service) CancelApplication(userID int64, applicationID int64) (Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.applications[applicationID]
	if !ok {
		if s.repo == nil {
			return Application{}, ErrApplicationNotFound
		}
		saved, err := s.repo.GetApplication(context.Background(), applicationID)
		if err != nil {
			return Application{}, err
		}
		app = saved
	}
	if app.UserID != userID {
		return Application{}, ErrForbidden
	}
	if app.Status != "pending" {
		return Application{}, ErrApplicationNotPending
	}
	app.Status = "cancelled"
	if s.repo != nil {
		saved, err := s.repo.UpdateApplication(context.Background(), app)
		if err != nil {
			return Application{}, err
		}
		app = saved
	}
	s.applications[app.ID] = app
	return app, nil
}

func (s *Service) ManualStart(userID int64, gameID int64) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Game{}, err
	}
	if !ok {
		return Game{}, ErrGameNotFound
	}
	if game.CreatorUserID != userID && game.MainGuideUserID != userID {
		return Game{}, ErrForbidden
	}
	if game.Status != "recruiting" && game.Status != "full" {
		return Game{}, ErrGameNotStartable
	}
	if game.CurrentPlayers < game.MinPlayers {
		return Game{}, ErrGameNotStartable
	}
	game.Status = "in_progress"
	if s.repo != nil {
		saved, err := s.repo.UpdateGame(context.Background(), game)
		if err != nil {
			return Game{}, err
		}
		game = saved
	}
	s.games[gameID] = game
	return game, nil
}

func (s *Service) ConfirmService(userID int64, gameID int64, note string, fileIDs ...int64) (ServiceConfirm, []ServiceConfirmItem, Game, error) {
	note = strings.TrimSpace(note)
	if len(note) > 300 || !validFileIDs(fileIDs, 9) {
		return ServiceConfirm{}, nil, Game{}, ErrInvalidGameInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return ServiceConfirm{}, nil, Game{}, err
	}
	if !ok {
		return ServiceConfirm{}, nil, Game{}, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return ServiceConfirm{}, nil, Game{}, ErrForbidden
	}
	if game.Status != "in_progress" && game.Status != "pending_confirm" && game.Status != "pending_review" {
		return ServiceConfirm{}, nil, Game{}, ErrGameNotConfirmable
	}

	confirm, ok := s.confirms[gameID]
	if !ok && s.confirmRepo != nil {
		saved, savedItems, found, err := s.confirmRepo.GetConfirm(context.Background(), gameID)
		if err != nil {
			return ServiceConfirm{}, nil, Game{}, err
		}
		if found {
			confirm = saved
			ok = true
			if s.confirmItems[gameID] == nil {
				s.confirmItems[gameID] = make(map[int64]ServiceConfirmItem)
			}
			for _, item := range savedItems {
				s.confirmItems[gameID][item.UserID] = item
			}
		}
	}
	if !ok {
		confirm = ServiceConfirm{
			ID:        s.nextConfirmID,
			GameID:    gameID,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		if s.confirmRepo != nil {
			saved, err := s.confirmRepo.SaveConfirm(context.Background(), confirm)
			if err != nil {
				return ServiceConfirm{}, nil, Game{}, err
			}
			confirm = saved
		}
		s.nextConfirmID++
		s.confirms[gameID] = confirm
	}
	if s.confirmItems[gameID] == nil {
		s.confirmItems[gameID] = make(map[int64]ServiceConfirmItem)
	}
	if _, exists := s.confirmItems[gameID][userID]; !exists {
		item := ServiceConfirmItem{
			ID:        s.nextConfirmItemID,
			ConfirmID: confirm.ID,
			GameID:    gameID,
			UserID:    userID,
			Note:      note,
			FileIDs:   append([]int64(nil), fileIDs...),
			CreatedAt: time.Now(),
		}
		if s.confirmRepo != nil {
			saved, err := s.confirmRepo.SaveConfirmItem(context.Background(), item)
			if err != nil {
				return ServiceConfirm{}, nil, Game{}, err
			}
			item = saved
		}
		s.nextConfirmItemID++
		s.confirmItems[gameID][userID] = item
	}

	confirmedBy := make([]int64, 0, len(s.confirmItems[gameID]))
	items := make([]ServiceConfirmItem, 0, len(s.confirmItems[gameID]))
	for memberID, item := range s.confirmItems[gameID] {
		confirmedBy = append(confirmedBy, memberID)
		items = append(items, item)
	}
	confirm.ConfirmedBy = confirmedBy
	if len(s.members[gameID]) > 0 && len(s.confirmItems[gameID]) >= len(s.members[gameID]) {
		confirm.Status = "completed"
		confirm.CompletedAt = time.Now().Format(time.RFC3339)
		game.Status = "pending_review"
	} else {
		confirm.Status = "pending"
		game.Status = "pending_confirm"
	}
	if s.confirmRepo != nil {
		saved, err := s.confirmRepo.SaveConfirm(context.Background(), confirm)
		if err != nil {
			return ServiceConfirm{}, nil, Game{}, err
		}
		confirm = saved
	}
	if s.repo != nil {
		saved, err := s.repo.UpdateGame(context.Background(), game)
		if err != nil {
			return ServiceConfirm{}, nil, Game{}, err
		}
		game = saved
	}
	s.confirms[gameID] = confirm
	s.games[gameID] = game
	return confirm, items, game, nil
}

func (s *Service) ServiceConfirmForGame(gameID int64) (ServiceConfirm, []ServiceConfirmItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.confirmRepo != nil {
		confirm, items, ok, err := s.confirmRepo.GetConfirm(context.Background(), gameID)
		if err == nil && ok {
			return confirm, items, true
		}
	}
	confirm, ok := s.confirms[gameID]
	if !ok {
		return ServiceConfirm{}, nil, false
	}
	items := make([]ServiceConfirmItem, 0, len(s.confirmItems[gameID]))
	for _, item := range s.confirmItems[gameID] {
		items = append(items, item)
	}
	return confirm, items, true
}

func (s *Service) Members(gameID int64) []int64 {
	if s.repo != nil {
		if items, err := s.repo.ListMembers(context.Background(), gameID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]int64, 0)
	for userID := range s.members[gameID] {
		result = append(result, userID)
	}
	return result
}

func (s *Service) IsMember(gameID int64, userID int64) bool {
	if s.repo != nil {
		if items, err := s.repo.ListMembers(context.Background(), gameID); err == nil {
			for _, item := range items {
				if item == userID {
					return true
				}
			}
			return false
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.members[gameID][userID]
}

func (s *Service) invitationAcceptedByApplicationLocked(applicationID int64) bool {
	if applicationID <= 0 {
		return false
	}
	for _, invitation := range s.invitations {
		if invitation.ApplicationID == applicationID && invitation.Status == "accepted" {
			return true
		}
	}
	return false
}

func canManageGameProgress(game Game, userID int64) bool {
	return game.CreatorUserID == userID || (game.MainGuideUserID > 0 && game.MainGuideUserID == userID)
}

func canInviteGuide(game Game, userID int64) bool {
	return game.CreatorUserID == userID || (game.MainGuideUserID > 0 && game.MainGuideUserID == userID)
}

func (s *Service) Exit(userID int64, gameID int64) (ExitResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return ExitResult{}, err
	}
	if !ok {
		return ExitResult{}, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return ExitResult{}, ErrForbidden
	}
	reason := exitReason(game.Status)
	creditDeduct := reason == "quit_after_confirm" || reason == "quit_after_started"
	memberStatus := exitMemberStatus(game.Status)
	if s.repo != nil {
		if err := s.repo.DeleteMember(context.Background(), gameID, userID, memberStatus, reason); err != nil {
			return ExitResult{}, err
		}
	}
	delete(s.members[gameID], userID)
	if game.CurrentPlayers > 0 {
		game.CurrentPlayers--
	}
	if s.repo != nil {
		saved, err := s.repo.UpdateGame(context.Background(), game)
		if err != nil {
			return ExitResult{}, err
		}
		game = saved
	}
	s.games[gameID] = game
	return ExitResult{Game: game, GameID: gameID, UserID: userID, Reason: reason, CreditDeduct: creditDeduct, CreditDeducted: creditDeduct, MemberStatus: memberStatus}, nil
}

func (s *Service) RecordExitCredit(gameID int64, userID int64, creditLogID int64) error {
	if gameID <= 0 || userID <= 0 || creditLogID <= 0 {
		return ErrInvalidGameInput
	}
	if s.repo != nil {
		return s.repo.UpdateMemberExitCredit(context.Background(), gameID, userID, true, creditLogID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

func (s *Service) AddProgressFeedback(userID int64, gameID int64, req ProgressFeedbackRequest) (ProgressFeedback, error) {
	req.Content = strings.TrimSpace(req.Content)
	if req.Progress < 0 || req.Progress > 100 || len(req.Content) > 500 || !validFileIDs(req.FileIDs, 9) {
		return ProgressFeedback{}, ErrInvalidProgress
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok := s.games[gameID]
	if !ok {
		return ProgressFeedback{}, ErrGameNotFound
	}
	if !canManageGameProgress(game, userID) {
		return ProgressFeedback{}, ErrForbidden
	}
	if game.Status != "in_progress" && game.Status != "pending_confirm" {
		return ProgressFeedback{}, ErrGameNotStartable
	}
	items := s.progressFeedbacks[gameID]
	if len(items) > 0 && req.Progress < items[len(items)-1].Progress {
		return ProgressFeedback{}, ErrInvalidProgress
	}
	feedback := ProgressFeedback{
		ID:        s.nextProgressID,
		GameID:    gameID,
		UserID:    userID,
		Progress:  req.Progress,
		Content:   req.Content,
		FileIDs:   append([]int64(nil), req.FileIDs...),
		CreatedAt: time.Now(),
	}
	s.nextProgressID++
	s.progressFeedbacks[gameID] = append(items, feedback)
	return feedback, nil
}

func (s *Service) ProgressFeedbacks(userID int64, gameID int64) ([]ProgressFeedback, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.games[gameID]; !ok {
		return nil, ErrGameNotFound
	}
	if !s.members[gameID][userID] {
		return nil, ErrForbidden
	}
	items := s.progressFeedbacks[gameID]
	result := make([]ProgressFeedback, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) CreateMilestone(userID int64, gameID int64, req MilestoneRequest) (Milestone, error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || len(req.Title) > 80 {
		return Milestone{}, ErrInvalidMilestone
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Milestone{}, err
	}
	if !ok {
		return Milestone{}, ErrGameNotFound
	}
	if !canManageGameProgress(game, userID) {
		return Milestone{}, ErrForbidden
	}
	item := Milestone{
		ID:        s.nextMilestoneID,
		GameID:    gameID,
		Title:     req.Title,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	if s.progressRepo != nil {
		saved, err := s.progressRepo.CreateMilestone(context.Background(), item)
		if err != nil {
			return Milestone{}, err
		}
		item = saved
	}
	s.nextMilestoneID++
	s.milestones[gameID] = append(s.milestones[gameID], item)
	return item, nil
}

func (s *Service) UpdateMilestone(userID int64, gameID int64, milestoneID int64, req MilestoneRequest) (Milestone, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Status = strings.TrimSpace(req.Status)
	if milestoneID <= 0 || (req.Title == "" && req.Status == "") {
		return Milestone{}, ErrInvalidMilestone
	}
	if len(req.Title) > 80 {
		return Milestone{}, ErrInvalidMilestone
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Milestone{}, err
	}
	if !ok {
		return Milestone{}, ErrGameNotFound
	}
	if !canManageGameProgress(game, userID) {
		return Milestone{}, ErrForbidden
	}
	items := s.milestones[gameID]
	for index, item := range items {
		if item.ID != milestoneID {
			continue
		}
		if req.Title != "" {
			item.Title = req.Title
		}
		if req.Status != "" {
			if !validMilestoneStatus(req.Status) {
				return Milestone{}, ErrInvalidMilestone
			}
			item.Status = req.Status
		}
		if s.progressRepo != nil {
			saved, err := s.progressRepo.UpdateMilestone(context.Background(), item)
			if err != nil {
				return Milestone{}, err
			}
			item = saved
		}
		items[index] = item
		s.milestones[gameID] = items
		return item, nil
	}
	return Milestone{}, ErrMilestoneNotFound
}

func (s *Service) Milestones(userID int64, gameID int64) ([]Milestone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return nil, ErrForbidden
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListMilestones(context.Background(), gameID)
	}
	items := s.milestones[gameID]
	result := make([]Milestone, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) AdminMilestones(gameID int64) ([]Milestone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListMilestones(context.Background(), gameID)
	}
	items := s.milestones[gameID]
	result := make([]Milestone, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) CreateCheckin(userID int64, gameID int64, req CheckinRequest) (Checkin, error) {
	req.CheckinType = strings.TrimSpace(req.CheckinType)
	req.Content = strings.TrimSpace(req.Content)
	if !validCheckinType(req.CheckinType) || len(req.Content) > 500 || req.MilestoneID < 0 || !validFileIDs(req.FileIDs, 9) {
		return Checkin{}, ErrInvalidCheckin
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return Checkin{}, err
	} else if !ok {
		return Checkin{}, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return Checkin{}, ErrForbidden
	}
	item := Checkin{
		ID:          s.nextCheckinID,
		GameID:      gameID,
		UserID:      userID,
		MilestoneID: req.MilestoneID,
		CheckinType: req.CheckinType,
		Content:     req.Content,
		FileIDs:     append([]int64(nil), req.FileIDs...),
		Status:      "valid",
		CreatedAt:   time.Now(),
	}
	if s.progressRepo != nil {
		saved, err := s.progressRepo.CreateCheckin(context.Background(), item)
		if err != nil {
			return Checkin{}, err
		}
		item = saved
	}
	s.nextCheckinID++
	s.checkins[gameID] = append(s.checkins[gameID], item)
	return item, nil
}

func (s *Service) Checkins(userID int64, gameID int64) ([]Checkin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return nil, ErrForbidden
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListCheckins(context.Background(), gameID, false)
	}
	items := s.checkins[gameID]
	result := make([]Checkin, 0, len(items))
	for _, item := range items {
		if item.Status == "invalid" {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) AdminCheckins(gameID int64) ([]Checkin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListCheckins(context.Background(), gameID, true)
	}
	items := s.checkins[gameID]
	result := make([]Checkin, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) MarkCheckinInvalid(checkinID int64) (Checkin, error) {
	if checkinID <= 0 {
		return Checkin{}, ErrInvalidCheckin
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.progressRepo != nil {
		item, err := s.progressRepo.UpdateCheckinStatus(context.Background(), checkinID, "invalid")
		if err != nil {
			return Checkin{}, err
		}
		for gameID, items := range s.checkins {
			for index, existing := range items {
				if existing.ID == checkinID {
					items[index] = item
					s.checkins[gameID] = items
					break
				}
			}
		}
		return item, nil
	}
	for gameID, items := range s.checkins {
		for index, item := range items {
			if item.ID != checkinID {
				continue
			}
			item.Status = "invalid"
			if s.progressRepo != nil {
				saved, err := s.progressRepo.UpdateCheckinStatus(context.Background(), checkinID, "invalid")
				if err != nil {
					return Checkin{}, err
				}
				item = saved
			}
			items[index] = item
			s.checkins[gameID] = items
			return item, nil
		}
	}
	return Checkin{}, ErrCheckinNotFound
}

func (s *Service) CreateRetrospective(userID int64, gameID int64, req RetrospectiveRequest) (Retrospective, error) {
	req.Content = strings.TrimSpace(req.Content)
	req.AgainIntent = strings.TrimSpace(req.AgainIntent)
	if req.Content == "" || len(req.Content) > 1000 || !validAgainIntent(req.AgainIntent) {
		return Retrospective{}, ErrInvalidRetrospective
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Retrospective{}, err
	}
	if !ok {
		return Retrospective{}, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return Retrospective{}, ErrForbidden
	}
	if game.Status != "pending_review" && game.Status != "completed" {
		return Retrospective{}, ErrGameNotConfirmable
	}
	for _, existing := range s.retrospectives[gameID] {
		if existing.UserID == userID {
			return Retrospective{}, ErrDuplicateRetrospective
		}
	}
	if s.progressRepo != nil {
		exists, err := s.progressRepo.RetrospectiveExists(context.Background(), gameID, userID)
		if err != nil {
			return Retrospective{}, err
		}
		if exists {
			return Retrospective{}, ErrDuplicateRetrospective
		}
	}
	item := Retrospective{
		ID:          s.nextRetroID,
		GameID:      gameID,
		UserID:      userID,
		Content:     req.Content,
		AgainIntent: req.AgainIntent,
		CreatedAt:   time.Now(),
	}
	if s.progressRepo != nil {
		saved, err := s.progressRepo.CreateRetrospective(context.Background(), item)
		if err != nil {
			return Retrospective{}, err
		}
		item = saved
	}
	s.nextRetroID++
	s.retrospectives[gameID] = append(s.retrospectives[gameID], item)
	return item, nil
}

func (s *Service) Retrospectives(userID int64, gameID int64) ([]Retrospective, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if !s.memberLocked(gameID, userID) {
		return nil, ErrForbidden
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListRetrospectives(context.Background(), gameID)
	}
	items := s.retrospectives[gameID]
	result := make([]Retrospective, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) AdminRetrospectives(gameID int64) ([]Retrospective, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListRetrospectives(context.Background(), gameID)
	}
	items := s.retrospectives[gameID]
	result := make([]Retrospective, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) ContinueDraft(userID int64, gameID int64, req ContinueDraftRequest) (ContinueDraft, error) {
	req.Title = strings.TrimSpace(req.Title)
	if len(req.Title) > 80 {
		return ContinueDraft{}, ErrInvalidGameInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return ContinueDraft{}, err
	}
	if !ok {
		return ContinueDraft{}, ErrGameNotFound
	}
	if game.CreatorUserID != userID {
		return ContinueDraft{}, ErrForbidden
	}
	if game.Status != "pending_review" && game.Status != "completed" {
		return ContinueDraft{}, ErrGameNotConfirmable
	}
	title := req.Title
	if title == "" {
		title = game.Title + " 续局"
	}
	draft := Game{
		ID:                    s.nextID,
		CreatorUserID:         userID,
		Title:                 title,
		GameType:              game.GameType,
		PrimaryCategory:       game.PrimaryCategory,
		PrimaryCategoryText:   game.PrimaryCategoryText,
		SecondaryCategory:     game.SecondaryCategory,
		SecondaryCategoryText: game.SecondaryCategoryText,
		Type:                  game.Type,
		GameSource:            "continue",
		Status:                "draft",
		MinPlayers:            game.MinPlayers,
		MaxPlayers:            game.MaxPlayers,
		CurrentPlayers:        1,
		CityCode:              game.CityCode,
		CityName:              game.CityName,
		Longitude:             game.Longitude,
		Latitude:              game.Latitude,
		CreatedAt:             time.Now(),
	}
	if s.repo != nil {
		saved, err := s.repo.CreateGame(context.Background(), draft)
		if err != nil {
			return ContinueDraft{}, err
		}
		draft = saved
		if err := s.repo.AddMember(context.Background(), draft.ID, userID, "creator"); err != nil {
			return ContinueDraft{}, err
		}
	}
	s.nextID++
	s.games[draft.ID] = draft
	s.members[draft.ID] = map[int64]bool{userID: true}
	record := ContinueDraftRecord{
		OriginalGameID: gameID,
		DraftGameID:    draft.ID,
		CreatorUserID:  userID,
		Title:          draft.Title,
		Status:         draft.Status,
		CreatedAt:      draft.CreatedAt,
	}
	if s.progressRepo != nil {
		saved, err := s.progressRepo.CreateContinueDraft(context.Background(), record)
		if err != nil {
			return ContinueDraft{}, err
		}
		record = saved
	}
	s.continueDrafts[gameID] = append(s.continueDrafts[gameID], record)
	return ContinueDraft{OriginalGameID: gameID, Draft: draft}, nil
}

func (s *Service) AdminContinueDrafts(gameID int64) ([]ContinueDraftRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok, err := s.gameLocked(gameID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrGameNotFound
	}
	if s.progressRepo != nil {
		return s.progressRepo.ListContinueDrafts(context.Background(), gameID)
	}
	items := s.continueDrafts[gameID]
	result := make([]ContinueDraftRecord, len(items))
	copy(result, items)
	return result, nil
}

func (s *Service) FavoriteGame(userID int64, gameID int64) (Favorite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok := s.games[gameID]
	if !ok {
		return Favorite{}, ErrGameNotFound
	}
	if s.favorites[userID] == nil {
		s.favorites[userID] = make(map[int64]Favorite)
	}
	if favorite, ok := s.favorites[userID][gameID]; ok {
		return favorite, nil
	}
	favorite := Favorite{UserID: userID, GameID: gameID, Game: game, CreatedAt: time.Now()}
	if s.favoriteRepo != nil {
		saved, err := s.favoriteRepo.SaveFavorite(context.Background(), favorite)
		if err != nil {
			return Favorite{}, err
		}
		favorite.CreatedAt = saved.CreatedAt
	}
	s.favorites[userID][gameID] = favorite
	return favorite, nil
}

func (s *Service) UnfavoriteGame(userID int64, gameID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[gameID]; !ok {
		return ErrGameNotFound
	}
	if s.favoriteRepo != nil {
		if err := s.favoriteRepo.DeleteFavorite(context.Background(), userID, gameID); err != nil {
			return err
		}
	}
	if s.favorites[userID] != nil {
		delete(s.favorites[userID], gameID)
	}
	return nil
}

func (s *Service) FavoriteGames(userID int64) []Favorite {
	if s.favoriteRepo != nil {
		if items, err := s.favoriteRepo.ListFavoritesByUser(context.Background(), userID); err == nil {
			return s.hydrateFavorites(items)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Favorite, 0, len(s.favorites[userID]))
	for gameID, favorite := range s.favorites[userID] {
		game, ok := s.games[gameID]
		if !ok || game.Status == "pending_audit" {
			continue
		}
		favorite.Game = game
		result = append(result, favorite)
	}
	sortFavorites(result)
	return result
}

func (s *Service) FavoritesForUser(userID int64) []Favorite {
	return s.FavoriteGames(userID)
}

func (s *Service) AllFavorites() []Favorite {
	if s.favoriteRepo != nil {
		if items, err := s.favoriteRepo.ListAllFavorites(context.Background()); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Favorite, 0)
	for _, userFavorites := range s.favorites {
		for _, favorite := range userFavorites {
			result = append(result, favorite)
		}
	}
	sortFavorites(result)
	return result
}

func (s *Service) hydrateFavorites(items []Favorite) []Favorite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Favorite, 0, len(items))
	for _, favorite := range items {
		game, ok := s.games[favorite.GameID]
		if !ok || game.Status == "pending_audit" {
			continue
		}
		favorite.Game = game
		result = append(result, favorite)
	}
	sortFavorites(result)
	return result
}

func (s *Service) StatsForUser(userID int64) UserStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := UserStats{UserID: userID}
	for gameID := range s.members {
		if !s.members[gameID][userID] {
			continue
		}
		stats.Participated++
		status := s.games[gameID].Status
		if status == "pending_review" || status == "completed" {
			stats.Completed++
		}
	}
	return stats
}

func exitReason(status string) string {
	switch status {
	case "pending_confirm":
		return "quit_after_confirm"
	case "in_progress", "pending_review", "completed":
		return "quit_after_started"
	default:
		return "quit_before_confirm"
	}
}

func exitMemberStatus(status string) string {
	switch exitReason(status) {
	case "quit_after_confirm":
		return "quit_after_confirm"
	case "quit_after_started":
		return "quit_after_started"
	default:
		return "quit_before_confirm"
	}
}

func (s *Service) ApplicationsForUser(userID int64) []Application {
	if s.repo != nil {
		if items, err := s.repo.ListApplicationsByUser(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Application, 0)
	for _, app := range s.applications {
		if app.UserID == userID {
			result = append(result, app)
		}
	}
	return result
}

func (s *Service) ApplicationsForCreator(userID int64) []Application {
	if s.repo != nil {
		if items, err := s.repo.ListApplicationsForCreator(context.Background(), userID); err == nil {
			return items
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Application, 0)
	for _, app := range s.applications {
		game, ok := s.games[app.GameID]
		if ok && game.CreatorUserID == userID {
			result = append(result, app)
		}
	}
	return result
}

func (s *Service) InvitationsForUser(userID int64) []Invitation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Invitation, 0)
	for _, invitation := range s.invitations {
		if invitation.InviterID == userID || invitation.TargetUserID == userID {
			result = append(result, invitation)
		}
	}
	return result
}

func (s *Service) Get(id int64) (Game, error) {
	if s.repo != nil {
		if game, err := s.repo.GetGame(context.Background(), id); err == nil {
			return game, nil
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, ok := s.games[id]
	if !ok {
		return Game{}, ErrGameNotFound
	}
	return game, nil
}

func (s *Service) gameLocked(gameID int64) (Game, bool, error) {
	game, ok := s.games[gameID]
	if ok {
		return game, true, nil
	}
	if s.repo == nil {
		return Game{}, false, nil
	}
	game, err := s.repo.GetGame(context.Background(), gameID)
	if err != nil {
		if errors.Is(err, ErrGameNotFound) {
			return Game{}, false, nil
		}
		return Game{}, false, err
	}
	s.games[game.ID] = game
	s.loadMembersLocked(game.ID)
	return game, true, nil
}

func (s *Service) memberLocked(gameID int64, userID int64) bool {
	if s.members[gameID] != nil && s.members[gameID][userID] {
		return true
	}
	s.loadMembersLocked(gameID)
	return s.members[gameID] != nil && s.members[gameID][userID]
}

func (s *Service) loadMembersLocked(gameID int64) {
	if s.repo == nil {
		return
	}
	items, err := s.repo.ListMembers(context.Background(), gameID)
	if err != nil {
		return
	}
	if s.members[gameID] == nil {
		s.members[gameID] = make(map[int64]bool)
	}
	for _, userID := range items {
		s.members[gameID][userID] = true
	}
}

func (s *Service) createdTodayLocked(userID int64) int {
	if s.repo != nil {
		if count, err := s.repo.CountGamesCreatedToday(context.Background(), userID, time.Now()); err == nil {
			return count
		}
	}
	count := 0
	now := time.Now()
	for _, game := range s.games {
		if game.CreatorUserID == userID && sameDay(now, game.CreatedAt) {
			count++
		}
	}
	return count
}

func sameDay(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func sortFavorites(items []Favorite) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].CreatedAt.After(items[i].CreatedAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
