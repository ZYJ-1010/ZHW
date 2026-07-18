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
	ErrRealnameRequired        = errors.New("realname required")
	ErrInvalidGameInput        = errors.New("invalid game input")
	ErrInvalidGameType         = errors.New("invalid game type")
	ErrInvalidPlayers          = errors.New("invalid players")
	ErrDailyLimit              = errors.New("daily limit reached")
	ErrGameNotFound            = errors.New("game not found")
	ErrGameNotRecruiting       = errors.New("game not recruiting")
	ErrSignupClosed            = errors.New("signup closed")
	ErrGameNotStartable        = errors.New("game not startable")
	ErrAlreadyApplied          = errors.New("already applied")
	ErrAlreadyMember           = errors.New("already member")
	ErrApplicationNotFound     = errors.New("application not found")
	ErrApplicationNotPending   = errors.New("application not pending")
	ErrFull                    = errors.New("game full")
	ErrForbidden               = errors.New("forbidden")
	ErrGameNotConfirmable      = errors.New("game not confirmable")
	ErrExpertConfirmRequired   = errors.New("expert confirmation required")
	ErrInvalidProgress         = errors.New("invalid progress")
	ErrInvalidMilestone        = errors.New("invalid milestone")
	ErrInvalidCheckin          = errors.New("invalid checkin")
	ErrInvalidRetrospective    = errors.New("invalid retrospective")
	ErrDuplicateRetrospective  = errors.New("duplicate retrospective")
	ErrCheckinNotFound         = errors.New("checkin not found")
	ErrMilestoneNotFound       = errors.New("milestone not found")
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationNotPending    = errors.New("invitation not pending")
	ErrInvitationPlayerPending = errors.New("player invitation pending")
	ErrAlreadyInvited          = errors.New("already invited")
)

const (
	MinGamePlayers = 5
	MaxGamePlayers = 8
)

var appTimeLocation = loadAppTimeLocation()

func loadAppTimeLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return location
	}
	return time.FixedZone("Asia/Shanghai", 8*60*60)
}

type IdentityChecker interface {
	IsVerified(userID int64) bool
}

// RoomEnsurer is notified after a game becomes full so the IM layer can
// create (or reconcile) the single room for that game. It is intentionally a
// small interface to keep the games package independent from the IM package.
type RoomEnsurer interface {
	EnsureRoom(gameID int64)
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
	RejectReason          string    `json:"rejectReason,omitempty"`
	StartReason           string    `json:"startReason,omitempty"`
	StartedByUserID       int64     `json:"startedByUserId,omitempty"`
	StartedAt             string    `json:"startedAt,omitempty"`
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
	ID           int64     `json:"id"`
	GameID       int64     `json:"gameId"`
	UserID       int64     `json:"userId"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Reason       string    `json:"reason,omitempty"`
	RejectReason string    `json:"rejectReason,omitempty"`
	FileIDs      []int64   `json:"fileIds,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Invitation struct {
	ID               int64     `json:"id"`
	InviteGroupID    string    `json:"inviteGroupId,omitempty"`
	GameID           int64     `json:"gameId"`
	InviterID        int64     `json:"inviterUserId"`
	TargetUserID     int64     `json:"targetUserId"`
	PlayerUserID     int64     `json:"playerUserId,omitempty"`
	ExpertUserID     int64     `json:"expertUserId,omitempty"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	Message          string    `json:"message,omitempty"`
	ServiceType      string    `json:"serviceType,omitempty"`
	ServiceDuration  string    `json:"serviceDuration,omitempty"`
	DemandDetail     string    `json:"demandDetail,omitempty"`
	BudgetAmountCent int64     `json:"budgetAmountCent,omitempty"`
	ExpectedTime     string    `json:"expectedTime,omitempty"`
	ApplicationID    int64     `json:"applicationId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	RespondedAt      string    `json:"respondedAt,omitempty"`
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

type ReplayGameOverrides struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       *float64 `json:"price,omitempty"`
	StartAt     string   `json:"startAt"`
	EndAt       string   `json:"endAt"`
}

type ReplayGameRequest struct {
	PlayerUserIDs    []int64             `json:"playerUserIds"`
	ExpertUserIDs    []int64             `json:"expertUserIds"`
	GuideUserIDs     []int64             `json:"guideUserIds"`
	Message          string              `json:"message"`
	InviteGroupID    string              `json:"inviteGroupId"`
	ServiceType      string              `json:"serviceType"`
	ServiceDuration  string              `json:"serviceDuration"`
	DemandDetail     string              `json:"demandDetail"`
	BudgetAmountCent int64               `json:"budgetAmountCent"`
	ExpectedTime     string              `json:"expectedTime"`
	Overrides        ReplayGameOverrides `json:"overrides"`
}

type ReplayGameResult struct {
	SourceGameID        int64        `json:"sourceGameId"`
	ReplayGameID        int64        `json:"replayGameId"`
	PrimaryInvitationID int64        `json:"primaryInvitationId"`
	InvitationCount     int          `json:"invitationCount"`
	Game                Game         `json:"replayGame"`
	Invitations         []Invitation `json:"invitations"`
	PlayerUserIDs       []int64      `json:"playerUserIds"`
	ExpertUserIDs       []int64      `json:"expertUserIds"`
	GuideUserIDs        []int64      `json:"guideUserIds"`
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

type memberRoleRepository interface {
	ListMemberRoles(ctx context.Context, gameID int64) ([]MemberRole, error)
}

type invitationListRepository interface {
	ListInvitationsForUser(ctx context.Context, userID int64) ([]Invitation, error)
}

type applicationReviewerRepository interface {
	ListApplicationsForReviewer(ctx context.Context, userID int64) ([]Application, error)
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

type MemberRole struct {
	UserID int64  `json:"userId"`
	Role   string `json:"role"`
}

type CreateRequest struct {
	Title                 string   `json:"title"`
	GameType              string   `json:"gameType"`
	CoverFileID           int64    `json:"coverFileId,omitempty"`
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
	TargetUserID     int64  `json:"targetUserId"`
	PlayerUserID     int64  `json:"playerUserId"`
	ExpertUserID     int64  `json:"expertUserId"`
	InviteGroupID    string `json:"inviteGroupId"`
	Role             string `json:"role"`
	RoleType         string `json:"roleType"`
	TargetRole       string `json:"targetRole"`
	Message          string `json:"message"`
	ServiceType      string `json:"serviceType"`
	ServiceDuration  string `json:"serviceDuration"`
	DemandDetail     string `json:"demandDetail"`
	BudgetAmountCent int64  `json:"budgetAmountCent"`
	ExpectedTime     string `json:"expectedTime"`
}

type InvitationRespondRequest struct {
	Accept bool   `json:"accept"`
	Reason string `json:"reason"`
}

func normalizeApplicationRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "expert":
		return "expert"
	case "guide", "leader", "main_guide":
		return "guide"
	default:
		return "player"
	}
}

func normalizeMemberRole(role string) string {
	switch normalizeApplicationRole(role) {
	case "expert":
		return "expert"
	case "guide":
		return "guide"
	default:
		return "member"
	}
}

func gameEntryRole(game Game, requestedRole string) string {
	if game.GameSource == "admin" {
		return "member"
	}
	return normalizeMemberRole(requestedRole)
}

func normalizeInvitationRole(role string) string {
	if strings.TrimSpace(role) == "" {
		return "guide"
	}
	return normalizeApplicationRole(role)
}

type ApplyRequest struct {
	Reason   string  `json:"reason"`
	Role     string  `json:"role"`
	RoleType string  `json:"roleType"`
	FileIDs  []int64 `json:"fileIds"`
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
	memberRoles       map[int64]map[int64]string
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
	roomEnsurer       RoomEnsurer
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
		memberRoles:       make(map[int64]map[int64]string),
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

// UseRoomEnsurer wires the IM room lifecycle into the games service. The
// callback is invoked only after a successful approval makes the game full.
func (s *Service) UseRoomEnsurer(ensurer RoomEnsurer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roomEnsurer = ensurer
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
		if err := s.repo.AddMember(context.Background(), game.ID, userID, "member"); err != nil {
			return Game{}, err
		}
	}
	s.nextID++
	s.games[game.ID] = game
	s.members[game.ID] = map[int64]bool{userID: true}
	s.memberRoles[game.ID] = map[int64]string{userID: "member"}
	return game, nil
}

func (s *Service) CreateFromAdmin(req CreateRequest) (Game, error) {
	req = normalizeCreateRequest(req)
	if req.CreatorUserID <= 0 {
		return Game{}, ErrInvalidGameInput
	}
	// 后台创建的局只有玩家身份，不能预设领路人或行家。
	if req.MainGuideUserID > 0 {
		return Game{}, ErrInvalidGameInput
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
	if !validAdminSignupTimeRange(req.SignupStartAt, req.SignupEndAt, req.StartAt) {
		return Game{}, ErrInvalidGameInput
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
		if err := s.repo.AddMember(context.Background(), game.ID, req.CreatorUserID, "member"); err != nil {
			return Game{}, err
		}
	}
	s.nextID++
	s.games[game.ID] = game
	s.members[game.ID] = map[int64]bool{req.CreatorUserID: true}
	s.memberRoles[game.ID] = map[int64]string{req.CreatorUserID: "member"}
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

func parseAppGameTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04"} {
		parsed, err := time.ParseInLocation(layout, value, appTimeLocation)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func validAppGameTimeRange(startText string, endText string) bool {
	startAt, hasStart := parseAppGameTime(startText)
	endAt, hasEnd := parseAppGameTime(endText)
	return hasStart && hasEnd && endAt.After(startAt)
}

func validAdminSignupTimeRange(signupStartText string, signupEndText string, gameStartText string) bool {
	signupStartAt, hasSignupStart := parseAppGameTime(signupStartText)
	signupEndAt, hasSignupEnd := parseAppGameTime(signupEndText)
	gameStartAt, hasGameStart := parseAppGameTime(gameStartText)
	return hasSignupStart && hasSignupEnd && hasGameStart &&
		signupEndAt.After(signupStartAt) && !signupEndAt.After(gameStartAt)
}

func CanApplyWithinSignupWindow(game Game, now time.Time) bool {
	return signupWindowError(game, now) == nil
}

func signupWindowError(game Game, now time.Time) error {
	if signupStartAt, ok := parseAppGameTime(game.SignupStartAt); ok && now.Before(signupStartAt) {
		return ErrSignupClosed
	}
	if signupEndAt, ok := parseAppGameTime(game.SignupEndAt); ok && now.After(signupEndAt) {
		return ErrSignupClosed
	}
	return nil
}

func validateCreateRequest(req CreateRequest) error {
	if req.MinPlayers < MinGamePlayers || req.MaxPlayers > MaxGamePlayers || req.MinPlayers > req.MaxPlayers {
		return ErrInvalidPlayers
	}
	if req.Title == "" || len(req.Title) > 80 || len(req.CityCode) > 32 || len(req.CityName) > 64 || len(req.Address) > 255 {
		return ErrInvalidGameInput
	}
	if req.StartAt == "" || req.EndAt == "" || !validAppGameTimeRange(req.StartAt, req.EndAt) {
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
	game.RejectReason = ""
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

func (s *Service) RejectGame(gameID int64, reason string) (Game, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return Game{}, ErrInvalidGameInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Game{}, err
	}
	if !ok {
		return Game{}, ErrGameNotFound
	}
	game.Status = "rejected"
	game.RejectReason = reason
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
	requestedRole := req.RoleType
	if strings.TrimSpace(requestedRole) == "" {
		requestedRole = req.Role
	}
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
	// 满员校验必须在服务层完成，避免绕过 HTTP handler 直接调用 API 时超额报名。
	if game.CurrentPlayers >= game.MaxPlayers {
		return Application{}, ErrFull
	}
	if err := signupWindowError(game, time.Now()); err != nil {
		return Application{}, err
	}
	if s.memberLocked(gameID, userID) {
		return Application{}, ErrAlreadyMember
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
		Role:      gameEntryRole(game, requestedRole),
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
	req.InviteGroupID = strings.TrimSpace(req.InviteGroupID)
	req.ServiceType = strings.TrimSpace(req.ServiceType)
	req.ServiceDuration = strings.TrimSpace(req.ServiceDuration)
	req.DemandDetail = strings.TrimSpace(req.DemandDetail)
	req.ExpectedTime = strings.TrimSpace(req.ExpectedTime)
	role := normalizeApplicationRole(req.RoleType)
	if strings.TrimSpace(req.RoleType) == "" {
		role = normalizeApplicationRole(req.Role)
	}
	if strings.TrimSpace(req.RoleType) == "" && strings.TrimSpace(req.Role) == "" {
		role = normalizeApplicationRole(req.TargetRole)
	}
	if strings.TrimSpace(req.RoleType) == "" && strings.TrimSpace(req.Role) == "" && strings.TrimSpace(req.TargetRole) == "" {
		role = "guide"
	}
	if role == "expert" {
		req.ExpertUserID = req.TargetUserID
	} else {
		req.PlayerUserID = req.TargetUserID
	}
	if req.TargetUserID <= 0 || req.TargetUserID == inviterID || len(req.Message) > 300 || len(req.ServiceType) > 120 || len(req.ServiceDuration) > 64 || len(req.DemandDetail) > 500 || len(req.ExpectedTime) > 64 || req.BudgetAmountCent < 0 {
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
	role = gameEntryRole(game, role)
	if game.Status != "recruiting" {
		return Invitation{}, ErrGameNotRecruiting
	}
	if !canInviteGuide(game, inviterID) {
		return Invitation{}, ErrForbidden
	}
	if s.memberLocked(gameID, req.TargetUserID) {
		return Invitation{}, ErrAlreadyMember
	}
	pendingExists, err := s.pendingInvitationExistsLocked(gameID, req.TargetUserID)
	if err != nil {
		return Invitation{}, err
	}
	if pendingExists {
		return Invitation{}, ErrAlreadyInvited
	}
	invitation := Invitation{
		ID:               s.nextInvitationID,
		InviteGroupID:    req.InviteGroupID,
		GameID:           gameID,
		InviterID:        inviterID,
		TargetUserID:     req.TargetUserID,
		PlayerUserID:     req.PlayerUserID,
		ExpertUserID:     req.ExpertUserID,
		Role:             role,
		Status:           "pending",
		Message:          req.Message,
		ServiceType:      req.ServiceType,
		ServiceDuration:  req.ServiceDuration,
		DemandDetail:     req.DemandDetail,
		BudgetAmountCent: req.BudgetAmountCent,
		ExpectedTime:     req.ExpectedTime,
		CreatedAt:        time.Now(),
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
	pairedInvitation := strings.TrimSpace(invitation.InviteGroupID) != "" && (normalizeApplicationRole(invitation.Role) == "player" || normalizeApplicationRole(invitation.Role) == "expert")
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
	if pairedInvitation {
		if err := s.requirePlayerInvitationAcceptedLocked(invitation); err != nil {
			return Invitation{}, Application{}, err
		}
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
	if err := signupWindowError(game, time.Now()); err != nil {
		return Invitation{}, Application{}, err
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
		Role:      normalizeInvitationRole(invitation.Role),
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
	return s.ReviewApplicationWithReason(operatorUserID, applicationID, approve, "")
}

func (s *Service) ReviewApplicationWithReason(operatorUserID int64, applicationID int64, approve bool, rejectReason string) (Application, error) {
	rejectReason = strings.TrimSpace(rejectReason)
	s.mu.Lock()
	var ensureRoomGameID int64
	roomEnsurer := s.roomEnsurer
	// Unlock before calling into the IM service. The IM service reads the game
	// membership map through this service, so invoking it while the mutex is
	// held would deadlock.
	defer func() {
		s.mu.Unlock()
		if ensureRoomGameID > 0 && roomEnsurer != nil {
			roomEnsurer.EnsureRoom(ensureRoomGameID)
		}
	}()
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
	if game.CreatorUserID != operatorUserID && game.MainGuideUserID != operatorUserID {
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
		memberRole := gameEntryRole(game, app.Role)
		if memberRole == "main_guide" && game.MainGuideUserID == 0 {
			game.MainGuideUserID = app.UserID
		} else if memberRole == "main_guide" {
			memberRole = "guide"
		}
		if memberRole == "guide" && game.GameSource != "admin" && game.MainGuideUserID == 0 {
			game.MainGuideUserID = app.UserID
			memberRole = "main_guide"
		}
		if s.memberRoles[game.ID] == nil {
			s.memberRoles[game.ID] = make(map[int64]string)
		}
		s.memberRoles[game.ID][app.UserID] = memberRole
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
		app.RejectReason = rejectReason
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
	if approve && game.Status == "full" {
		ensureRoomGameID = game.ID
	}
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
	return s.ManualStartWithReason(userID, gameID, "发起人手动开始")
}

func (s *Service) ManualStartWithReason(userID int64, gameID int64, startReason string) (Game, error) {
	startReason = strings.TrimSpace(startReason)
	if startReason == "" || len([]rune(startReason)) > 300 {
		return Game{}, ErrInvalidGameInput
	}
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
	game.StartReason = startReason
	game.StartedByUserID = userID
	game.StartedAt = time.Now().Format(time.RFC3339)
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

// RequestCompletion ends admin-created and no-expert games immediately, while
// app-created games with experts enter the ordered expert-then-player flow.
func (s *Service) RequestCompletion(userID int64, gameID int64) (Game, error) {
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
	if game.Status == "pending_confirm" || game.Status == "pending_review" {
		return game, nil
	}
	if game.Status != "in_progress" {
		return Game{}, ErrGameNotConfirmable
	}
	expertIDs, _, _, err := s.confirmationParticipantsLocked(game)
	if err != nil {
		return Game{}, err
	}
	if game.GameSource == "admin" || len(expertIDs) == 0 {
		game.Status = "pending_review"
	} else {
		game.Status = "pending_confirm"
	}
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
	if game.GameSource == "admin" {
		return ServiceConfirm{}, nil, Game{}, ErrGameNotConfirmable
	}
	expertIDs, playerIDs, roleByUser, err := s.confirmationParticipantsLocked(game)
	if err != nil {
		return ServiceConfirm{}, nil, Game{}, err
	}
	role := roleByUser[userID]
	if role != "expert" && role != "member" {
		return ServiceConfirm{}, nil, Game{}, ErrForbidden
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
	if game.Status == "pending_review" {
		if _, exists := s.confirmItems[gameID][userID]; exists {
			items := make([]ServiceConfirmItem, 0, len(s.confirmItems[gameID]))
			for _, item := range s.confirmItems[gameID] {
				items = append(items, item)
			}
			return confirm, items, game, nil
		}
		return ServiceConfirm{}, nil, Game{}, ErrGameNotConfirmable
	}
	if role == "member" && !allConfirmationUsersConfirmed(expertIDs, s.confirmItems[gameID]) {
		return ServiceConfirm{}, nil, Game{}, ErrExpertConfirmRequired
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
	allExpertsConfirmed := allConfirmationUsersConfirmed(expertIDs, s.confirmItems[gameID])
	allPlayersConfirmed := allConfirmationUsersConfirmed(playerIDs, s.confirmItems[gameID])
	if len(playerIDs) > 0 && allExpertsConfirmed && allPlayersConfirmed {
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

func (s *Service) ResolveNoExpertPendingConfirm(gameID int64) (Game, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Game{}, false, err
	}
	if !ok {
		return Game{}, false, ErrGameNotFound
	}
	if game.Status != "pending_confirm" {
		return game, false, nil
	}
	expertIDs, _, _, err := s.confirmationParticipantsLocked(game)
	if err != nil {
		return Game{}, false, err
	}
	if len(expertIDs) > 0 {
		return game, false, nil
	}
	game.Status = "pending_review"
	if confirm, ok := s.confirms[gameID]; ok {
		confirm.Status = "completed"
		if confirm.CompletedAt == "" {
			confirm.CompletedAt = time.Now().Format(time.RFC3339)
		}
		if s.confirmRepo != nil {
			saved, err := s.confirmRepo.SaveConfirm(context.Background(), confirm)
			if err != nil {
				return Game{}, false, err
			}
			confirm = saved
		}
		s.confirms[gameID] = confirm
	}
	if s.repo != nil {
		saved, err := s.repo.UpdateGame(context.Background(), game)
		if err != nil {
			return Game{}, false, err
		}
		game = saved
	}
	s.games[gameID] = game
	return game, true, nil
}

func (s *Service) confirmationParticipantsLocked(game Game) ([]int64, []int64, map[int64]string, error) {
	memberIDs := make([]int64, 0)
	roleByUser := make(map[int64]string)
	if s.repo != nil {
		items, err := s.repo.ListMembers(context.Background(), game.ID)
		if err != nil {
			return nil, nil, nil, err
		}
		memberIDs = append(memberIDs, items...)
		if repo, ok := s.repo.(memberRoleRepository); ok {
			roles, err := repo.ListMemberRoles(context.Background(), game.ID)
			if err != nil {
				return nil, nil, nil, err
			}
			for _, item := range roles {
				roleByUser[item.UserID] = normalizeConfirmationRole(item.Role)
			}
		}
	} else {
		for memberID := range s.members[game.ID] {
			memberIDs = append(memberIDs, memberID)
		}
		for memberID, role := range s.memberRoles[game.ID] {
			roleByUser[memberID] = normalizeConfirmationRole(role)
		}
	}
	expertIDs := make([]int64, 0)
	playerIDs := make([]int64, 0)
	for _, memberID := range memberIDs {
		role := roleByUser[memberID]
		if role == "" {
			if memberID == game.MainGuideUserID {
				role = "guide"
			} else {
				role = "member"
			}
			roleByUser[memberID] = role
		}
		switch role {
		case "expert":
			expertIDs = append(expertIDs, memberID)
		case "member":
			playerIDs = append(playerIDs, memberID)
		}
	}

	// A grouped invitation explicitly binds an expert to the player receiving
	// that service. When such bindings exist, unrelated players in the same
	// game must neither block completion nor be allowed to confirm it.
	pairedExperts := make([]int64, 0, len(expertIDs))
	pairedPlayers := make([]int64, 0, len(playerIDs))
	pairedRoles := make(map[int64]string)
	seenExperts := make(map[int64]bool)
	seenPlayers := make(map[int64]bool)
	for _, expertID := range expertIDs {
		invitations, err := s.confirmationInvitationsForUserLocked(expertID)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, invitation := range invitations {
			playerID := invitation.PlayerUserID
			if invitation.GameID != game.ID || invitation.TargetUserID != expertID || invitation.Status != "accepted" || normalizeApplicationRole(invitation.Role) != "expert" || playerID <= 0 || roleByUser[playerID] != "member" {
				continue
			}
			if !seenExperts[expertID] {
				seenExperts[expertID] = true
				pairedExperts = append(pairedExperts, expertID)
				pairedRoles[expertID] = "expert"
			}
			if !seenPlayers[playerID] {
				seenPlayers[playerID] = true
				pairedPlayers = append(pairedPlayers, playerID)
				pairedRoles[playerID] = "member"
			}
		}
	}
	if len(pairedExperts) > 0 && len(pairedPlayers) > 0 {
		return pairedExperts, pairedPlayers, pairedRoles, nil
	}
	return expertIDs, playerIDs, roleByUser, nil
}

func (s *Service) confirmationInvitationsForUserLocked(userID int64) ([]Invitation, error) {
	itemsByID := make(map[int64]Invitation)
	for id, invitation := range s.invitations {
		if invitation.InviterID == userID || invitation.TargetUserID == userID {
			itemsByID[id] = invitation
		}
	}
	if s.repo != nil {
		if repo, ok := s.repo.(invitationListRepository); ok {
			items, err := repo.ListInvitationsForUser(context.Background(), userID)
			if err != nil {
				return nil, err
			}
			for _, invitation := range items {
				s.invitations[invitation.ID] = invitation
				itemsByID[invitation.ID] = invitation
			}
		}
	}
	items := make([]Invitation, 0, len(itemsByID))
	for _, invitation := range itemsByID {
		items = append(items, invitation)
	}
	return items, nil
}

func normalizeConfirmationRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "expert":
		return "expert"
	case "guide", "leader", "main_guide":
		return "guide"
	case "member", "player", "creator":
		return "member"
	default:
		return ""
	}
}

func allConfirmationUsersConfirmed(userIDs []int64, items map[int64]ServiceConfirmItem) bool {
	for _, userID := range userIDs {
		if _, ok := items[userID]; !ok {
			return false
		}
	}
	return true
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

func (s *Service) MemberRoles(gameID int64) []MemberRole {
	if s.repo != nil {
		if repo, ok := s.repo.(memberRoleRepository); ok {
			if items, err := repo.ListMemberRoles(context.Background(), gameID); err == nil {
				return items
			}
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	game := s.games[gameID]
	roleMap := s.memberRoles[gameID]
	userIDs := make(map[int64]bool, len(s.members[gameID])+len(roleMap))
	for userID := range s.members[gameID] {
		userIDs[userID] = true
	}
	for userID := range roleMap {
		userIDs[userID] = true
	}
	result := make([]MemberRole, 0, len(userIDs))
	for userID := range userIDs {
		role := normalizeMemberRole(roleMap[userID])
		if game.GameSource == "admin" {
			role = "member"
		} else if role == "guide" && game.MainGuideUserID == userID {
			role = "main_guide"
		}
		result = append(result, MemberRole{UserID: userID, Role: role})
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

func (s *Service) IsIMReadOnly(gameID int64) bool {
	game, err := s.Get(gameID)
	if err != nil {
		return false
	}
	switch game.Status {
	case "pending_confirm", "pending_review", "completed", "cancelled":
		return true
	default:
		return false
	}
}

// IsIMRoomReady reports whether a game has reached a lifecycle state where an
// IM room may exist. Recruiting and pre-approval games deliberately return
// false so merely opening their detail or IM endpoint cannot create a room.
func (s *Service) IsIMRoomReady(gameID int64) bool {
	game, err := s.Get(gameID)
	if err != nil {
		return false
	}
	switch game.Status {
	case "full", "in_progress", "pending_confirm", "pending_review", "completed":
		return true
	default:
		return false
	}
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

func (s *Service) pendingInvitationExistsLocked(gameID int64, targetUserID int64) (bool, error) {
	for _, invitation := range s.invitations {
		if invitation.GameID == gameID && invitation.TargetUserID == targetUserID && invitation.Status == "pending" {
			return true, nil
		}
	}
	if s.repo == nil {
		return false, nil
	}
	repo, ok := s.repo.(invitationListRepository)
	if !ok {
		return false, nil
	}
	items, err := repo.ListInvitationsForUser(context.Background(), targetUserID)
	if err != nil {
		return false, err
	}
	for _, invitation := range items {
		if invitation.GameID == gameID && invitation.TargetUserID == targetUserID && invitation.Status == "pending" {
			s.invitations[invitation.ID] = invitation
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) requirePlayerInvitationAcceptedLocked(invitation Invitation) error {
	if normalizeApplicationRole(invitation.Role) != "expert" {
		return nil
	}
	playerUserID := invitation.PlayerUserID
	if playerUserID <= 0 {
		return ErrInvitationPlayerPending
	}
	for _, candidate := range s.invitations {
		if matchingPlayerInvitation(candidate, invitation, playerUserID) {
			if candidate.Status == "accepted" {
				return nil
			}
			return ErrInvitationPlayerPending
		}
	}
	if s.repo == nil {
		return ErrInvitationPlayerPending
	}
	repo, ok := s.repo.(invitationListRepository)
	if !ok {
		return ErrInvitationPlayerPending
	}
	items, err := repo.ListInvitationsForUser(context.Background(), playerUserID)
	if err != nil {
		return err
	}
	for _, candidate := range items {
		if !matchingPlayerInvitation(candidate, invitation, playerUserID) {
			continue
		}
		s.invitations[candidate.ID] = candidate
		if candidate.Status == "accepted" {
			return nil
		}
		return ErrInvitationPlayerPending
	}
	return ErrInvitationPlayerPending
}

func matchingPlayerInvitation(candidate Invitation, expertInvitation Invitation, playerUserID int64) bool {
	if candidate.ID == expertInvitation.ID || candidate.GameID != expertInvitation.GameID || candidate.InviterID != expertInvitation.InviterID || candidate.TargetUserID != playerUserID || normalizeApplicationRole(candidate.Role) == "expert" {
		return false
	}
	if expertInvitation.InviteGroupID != "" && candidate.InviteGroupID != "" && candidate.InviteGroupID != expertInvitation.InviteGroupID {
		return false
	}
	if candidate.PlayerUserID > 0 && candidate.PlayerUserID != playerUserID {
		return false
	}
	return true
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
	if s.memberRoles[gameID] != nil {
		delete(s.memberRoles[gameID], userID)
	}
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

func (s *Service) CancelService(gameID int64, reason string) (Game, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "service_canceled"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok, err := s.gameLocked(gameID)
	if err != nil {
		return Game{}, err
	}
	if !ok {
		return Game{}, ErrGameNotFound
	}
	if game.Status != "in_progress" {
		return Game{}, ErrGameNotConfirmable
	}

	memberIDs := make([]int64, 0)
	if s.repo != nil {
		items, err := s.repo.ListMembers(context.Background(), gameID)
		if err != nil {
			return Game{}, err
		}
		memberIDs = append(memberIDs, items...)
	} else {
		for memberID := range s.members[gameID] {
			memberIDs = append(memberIDs, memberID)
		}
	}

	if s.repo != nil {
		for _, memberID := range memberIDs {
			if err := s.repo.DeleteMember(context.Background(), gameID, memberID, "canceled", reason); err != nil {
				return Game{}, err
			}
		}
	}
	for _, memberID := range memberIDs {
		delete(s.members[gameID], memberID)
	}

	game.Status = "canceled"
	game.CurrentPlayers = 0
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
		if err := s.repo.AddMember(context.Background(), draft.ID, userID, "member"); err != nil {
			return ContinueDraft{}, err
		}
	}
	s.nextID++
	s.games[draft.ID] = draft
	s.members[draft.ID] = map[int64]bool{userID: true}
	s.memberRoles[draft.ID] = map[int64]string{userID: "member"}
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

func (s *Service) CreateReplayGame(userID int64, sourceGameID int64, req ReplayGameRequest) (ReplayGameResult, error) {
	req.Message = strings.TrimSpace(req.Message)
	req.InviteGroupID = strings.TrimSpace(req.InviteGroupID)
	req.ServiceType = strings.TrimSpace(req.ServiceType)
	req.ServiceDuration = strings.TrimSpace(req.ServiceDuration)
	req.DemandDetail = strings.TrimSpace(req.DemandDetail)
	req.ExpectedTime = strings.TrimSpace(req.ExpectedTime)
	req.Overrides.Title = strings.TrimSpace(req.Overrides.Title)
	req.Overrides.Description = strings.TrimSpace(req.Overrides.Description)
	req.Overrides.StartAt = strings.TrimSpace(req.Overrides.StartAt)
	req.Overrides.EndAt = strings.TrimSpace(req.Overrides.EndAt)
	if sourceGameID <= 0 || len(req.Message) > 300 || len(req.ServiceType) > 120 || len(req.ServiceDuration) > 64 || len(req.DemandDetail) > 500 || len(req.ExpectedTime) > 64 || req.BudgetAmountCent < 0 || len(req.Overrides.Title) > 80 || len(req.Overrides.Description) > 2000 || len(req.Overrides.StartAt) > 64 || len(req.Overrides.EndAt) > 64 {
		return ReplayGameResult{}, ErrInvalidGameInput
	}
	if req.Overrides.Price != nil && *req.Overrides.Price < 0 {
		return ReplayGameResult{}, ErrInvalidGameInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	source, ok, err := s.gameLocked(sourceGameID)
	if err != nil {
		return ReplayGameResult{}, err
	}
	if !ok {
		return ReplayGameResult{}, ErrGameNotFound
	}
	if !canInviteGuide(source, userID) {
		return ReplayGameResult{}, ErrForbidden
	}
	if s.createdTodayLocked(userID) >= s.dailyCreateLimit {
		return ReplayGameResult{}, ErrDailyLimit
	}

	type replayTarget struct {
		userID int64
		role   string
	}
	targets := make([]replayTarget, 0, len(req.PlayerUserIDs)+len(req.ExpertUserIDs)+len(req.GuideUserIDs))
	seen := map[int64]bool{userID: true}
	appendTargets := func(userIDs []int64, role string) {
		for _, targetUserID := range userIDs {
			if targetUserID <= 0 || seen[targetUserID] {
				continue
			}
			seen[targetUserID] = true
			targets = append(targets, replayTarget{userID: targetUserID, role: role})
		}
	}
	appendTargets(req.PlayerUserIDs, "player")
	appendTargets(req.ExpertUserIDs, "expert")
	appendTargets(req.GuideUserIDs, "guide")
	if len(targets) == 0 {
		return ReplayGameResult{}, ErrInvalidGameInput
	}

	title := req.Overrides.Title
	if title == "" {
		title = source.Title + " 续局"
	}
	description := source.Description
	if req.Overrides.Description != "" {
		description = req.Overrides.Description
	}
	price := source.Price
	if req.Overrides.Price != nil {
		price = *req.Overrides.Price
	}
	startAt := req.Overrides.StartAt
	if startAt == "" {
		startAt = source.StartAt
	}
	endAt := req.Overrides.EndAt
	if endAt == "" {
		endAt = source.EndAt
	}
	now := time.Now()
	replayGame := Game{
		ID:                    s.nextID,
		CreatorUserID:         userID,
		Title:                 title,
		GameType:              source.GameType,
		CoverImage:            source.CoverImage,
		Description:           description,
		Highlights:            source.Highlights,
		Notice:                source.Notice,
		Audience:              source.Audience,
		Participation:         source.Participation,
		Price:                 price,
		ProfitTemplate:        source.ProfitTemplate,
		StartAt:               startAt,
		EndAt:                 endAt,
		Tags:                  append([]string{}, source.Tags...),
		CompletionRules:       append([]string{}, source.CompletionRules...),
		PrimaryCategory:       source.PrimaryCategory,
		PrimaryCategoryText:   source.PrimaryCategoryText,
		SecondaryCategory:     source.SecondaryCategory,
		SecondaryCategoryText: source.SecondaryCategoryText,
		Type:                  source.Type,
		GameSource:            "replay",
		Status:                "recruiting",
		MinPlayers:            source.MinPlayers,
		MaxPlayers:            max(source.MaxPlayers, len(targets)+1),
		CurrentPlayers:        1,
		CityCode:              source.CityCode,
		CityName:              source.CityName,
		Address:               source.Address,
		Longitude:             source.Longitude,
		Latitude:              source.Latitude,
		CreatedAt:             now,
	}
	if s.repo != nil {
		saved, createErr := s.repo.CreateGame(context.Background(), replayGame)
		if createErr != nil {
			return ReplayGameResult{}, createErr
		}
		replayGame = saved
		if addErr := s.repo.AddMember(context.Background(), replayGame.ID, userID, "member"); addErr != nil {
			return ReplayGameResult{}, addErr
		}
	}
	s.nextID++
	s.games[replayGame.ID] = replayGame
	s.members[replayGame.ID] = map[int64]bool{userID: true}
	s.memberRoles[replayGame.ID] = map[int64]string{userID: "creator"}

	invitations := make([]Invitation, 0, len(targets))
	playerUserIDs := make([]int64, 0)
	expertUserIDs := make([]int64, 0)
	guideUserIDs := make([]int64, 0)
	firstExpertUserID := int64(0)
	if len(req.ExpertUserIDs) > 0 {
		firstExpertUserID = req.ExpertUserIDs[0]
	}
	firstPlayerUserID := int64(0)
	if len(req.PlayerUserIDs) > 0 {
		firstPlayerUserID = req.PlayerUserIDs[0]
	}
	for _, target := range targets {
		role := normalizeApplicationRole(target.role)
		invitation := Invitation{
			ID:               s.nextInvitationID,
			InviteGroupID:    req.InviteGroupID,
			GameID:           replayGame.ID,
			InviterID:        userID,
			TargetUserID:     target.userID,
			Role:             role,
			Status:           "pending",
			Message:          req.Message,
			ServiceType:      req.ServiceType,
			ServiceDuration:  req.ServiceDuration,
			DemandDetail:     req.DemandDetail,
			BudgetAmountCent: req.BudgetAmountCent,
			ExpectedTime:     req.ExpectedTime,
			CreatedAt:        now,
		}
		if role == "expert" {
			invitation.ExpertUserID = target.userID
			invitation.PlayerUserID = firstPlayerUserID
		} else if role == "player" {
			invitation.PlayerUserID = target.userID
			invitation.ExpertUserID = firstExpertUserID
		}
		if s.repo != nil {
			saved, createErr := s.repo.CreateInvitation(context.Background(), invitation)
			if createErr != nil {
				return ReplayGameResult{}, createErr
			}
			invitation = saved
		}
		s.nextInvitationID++
		s.invitations[invitation.ID] = invitation
		invitations = append(invitations, invitation)
		switch role {
		case "expert":
			expertUserIDs = append(expertUserIDs, target.userID)
		case "guide":
			guideUserIDs = append(guideUserIDs, target.userID)
		default:
			playerUserIDs = append(playerUserIDs, target.userID)
		}
	}

	record := ContinueDraftRecord{
		OriginalGameID: sourceGameID,
		DraftGameID:    replayGame.ID,
		CreatorUserID:  userID,
		Title:          replayGame.Title,
		Status:         replayGame.Status,
		CreatedAt:      replayGame.CreatedAt,
	}
	if s.progressRepo != nil {
		saved, createErr := s.progressRepo.CreateContinueDraft(context.Background(), record)
		if createErr != nil {
			return ReplayGameResult{}, createErr
		}
		record = saved
	}
	s.continueDrafts[sourceGameID] = append(s.continueDrafts[sourceGameID], record)

	primaryInvitationID := int64(0)
	if len(invitations) > 0 {
		primaryInvitationID = invitations[0].ID
	}
	return ReplayGameResult{
		SourceGameID:        sourceGameID,
		ReplayGameID:        replayGame.ID,
		PrimaryInvitationID: primaryInvitationID,
		InvitationCount:     len(invitations),
		Game:                replayGame,
		Invitations:         invitations,
		PlayerUserIDs:       playerUserIDs,
		ExpertUserIDs:       expertUserIDs,
		GuideUserIDs:        guideUserIDs,
	}, nil
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
	if repository, ok := s.repo.(applicationReviewerRepository); ok {
		if items, err := repository.ListApplicationsForReviewer(context.Background(), userID); err == nil {
			return items
		}
	}
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
		if ok && (game.CreatorUserID == userID || game.MainGuideUserID == userID) {
			result = append(result, app)
		}
	}
	return result
}

func (s *Service) InvitationsForUser(userID int64) []Invitation {
	if repo, ok := s.repo.(invitationListRepository); ok {
		if items, err := repo.ListInvitationsForUser(context.Background(), userID); err == nil {
			return items
		}
	}
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
