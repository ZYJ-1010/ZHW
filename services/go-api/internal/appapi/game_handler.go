package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/im"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/reviews"
)

type gameService interface {
	Create(userID int64, req games.CreateRequest) (games.Game, error)
	CreateFromAdmin(req games.CreateRequest) (games.Game, error)
	List() []games.Game
	Get(id int64) (games.Game, error)
	ApproveGame(gameID int64) (games.Game, error)
	RejectGame(gameID int64, reason string) (games.Game, error)
	SameCity(cityCode string) []games.Game
	Apply(userID int64, gameID int64, req games.ApplyRequest) (games.Application, error)
	CreateInvitation(inviterID int64, gameID int64, req games.InvitationRequest) (games.Invitation, error)
	RespondInvitation(userID int64, invitationID int64, req games.InvitationRespondRequest) (games.Invitation, games.Application, error)
	ReviewApplication(operatorUserID int64, applicationID int64, approve bool) (games.Application, error)
	ReviewApplicationWithReason(operatorUserID int64, applicationID int64, approve bool, rejectReason string) (games.Application, error)
	CancelApplication(userID int64, applicationID int64) (games.Application, error)
	ApplicationsForUser(userID int64) []games.Application
	ApplicationsForCreator(userID int64) []games.Application
	InvitationsForUser(userID int64) []games.Invitation
	ManualStart(userID int64, gameID int64) (games.Game, error)
	ManualStartWithReason(userID int64, gameID int64, startReason string) (games.Game, error)
	RequestCompletion(userID int64, gameID int64) (games.Game, error)
	Exit(userID int64, gameID int64) (games.ExitResult, error)
	CancelService(gameID int64, reason string) (games.Game, error)
	RecordExitCredit(gameID int64, userID int64, creditLogID int64) error
	ConfirmService(userID int64, gameID int64, note string, fileIDs ...int64) (games.ServiceConfirm, []games.ServiceConfirmItem, games.Game, error)
	ResolveNoExpertPendingConfirm(gameID int64) (games.Game, bool, error)
	ServiceConfirmForGame(gameID int64) (games.ServiceConfirm, []games.ServiceConfirmItem, bool)
	AddProgressFeedback(userID int64, gameID int64, req games.ProgressFeedbackRequest) (games.ProgressFeedback, error)
	ProgressFeedbacks(userID int64, gameID int64) ([]games.ProgressFeedback, error)
	CreateMilestone(userID int64, gameID int64, req games.MilestoneRequest) (games.Milestone, error)
	UpdateMilestone(userID int64, gameID int64, milestoneID int64, req games.MilestoneRequest) (games.Milestone, error)
	Milestones(userID int64, gameID int64) ([]games.Milestone, error)
	AdminMilestones(gameID int64) ([]games.Milestone, error)
	CreateCheckin(userID int64, gameID int64, req games.CheckinRequest) (games.Checkin, error)
	Checkins(userID int64, gameID int64) ([]games.Checkin, error)
	AdminCheckins(gameID int64) ([]games.Checkin, error)
	MarkCheckinInvalid(checkinID int64) (games.Checkin, error)
	CreateRetrospective(userID int64, gameID int64, req games.RetrospectiveRequest) (games.Retrospective, error)
	Retrospectives(userID int64, gameID int64) ([]games.Retrospective, error)
	AdminRetrospectives(gameID int64) ([]games.Retrospective, error)
	ContinueDraft(userID int64, gameID int64, req games.ContinueDraftRequest) (games.ContinueDraft, error)
	CreateReplayGame(userID int64, sourceGameID int64, req games.ReplayGameRequest) (games.ReplayGameResult, error)
	AdminContinueDrafts(gameID int64) ([]games.ContinueDraftRecord, error)
	FavoriteGame(userID int64, gameID int64) (games.Favorite, error)
	UnfavoriteGame(userID int64, gameID int64) error
	FavoriteGames(userID int64) []games.Favorite
	FavoritesForUser(userID int64) []games.Favorite
	AllFavorites() []games.Favorite
	StatsForUser(userID int64) games.UserStats
	Members(gameID int64) []int64
	MemberRoles(gameID int64) []games.MemberRole
	IsMember(gameID int64, userID int64) bool
}

type GameDetailDTO struct {
	games.Game
	MyRelation         GameMyRelationDTO    `json:"myRelation"`
	DetailDisplay      GameDetailDisplayDTO `json:"detailDisplay"`
	MemberIDs          []int64              `json:"memberIds"`
	Members            []GameMemberDTO      `json:"members"`
	IsFavorited        bool                 `json:"isFavorited"`
	FavoriteCount      int                  `json:"favoriteCount"`
	Progress           GameProgressDTO      `json:"progress"`
	IM                 GameIMDTO            `json:"im"`
	Review             GameReviewDTO        `json:"review"`
	ServiceConfirm     *ServiceConfirmDTO   `json:"serviceConfirm,omitempty"`
	PendingApplication *games.Application   `json:"pendingApplication,omitempty"`
	AuditRejectReason  string               `json:"auditRejectReason,omitempty"`
}

type GameListAvatarDTO struct {
	UserID    int64  `json:"userId"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type GameListItemDTO struct {
	games.Game
	PlayerAvatars []GameListAvatarDTO `json:"playerAvatars"`
}

type GameDetailDisplayDTO struct {
	StatusText              string                     `json:"statusText"`
	PendingApplicationCount int                        `json:"pendingApplicationCount"`
	Organizer               GameDetailOrganizerDTO     `json:"organizer"`
	PrimaryAction           GameDetailPrimaryActionDTO `json:"primaryAction"`
}

type GameDetailOrganizerDTO struct {
	UserID      int64  `json:"userId"`
	Name        string `json:"name"`
	AvatarText  string `json:"avatarText"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	Role        string `json:"role"`
	RoleLabel   string `json:"roleLabel"`
	Rating      string `json:"rating,omitempty"`
	RatingCount int    `json:"ratingCount"`
}

type GameDetailPrimaryActionDTO struct {
	Text        string `json:"text"`
	Disabled    bool   `json:"disabled"`
	Action      string `json:"action"`
	Route       string `json:"route,omitempty"`
	ConfirmText string `json:"confirmText,omitempty"`
}

type GameMyRelationDTO struct {
	Role                string `json:"role"`
	IsCreator           bool   `json:"isCreator"`
	IsMember            bool   `json:"isMember"`
	CanApply            bool   `json:"canApply"`
	ApplyDisabledReason string `json:"applyDisabledReason,omitempty"`
	CanAudit            bool   `json:"canAudit"`
	CanStart            bool   `json:"canStart"`
	CanEnterIM          bool   `json:"canEnterIM"`
	CanConfirm          bool   `json:"canConfirm"`
	CanReview           bool   `json:"canReview"`
	ApplicationID       int64  `json:"applicationId,omitempty"`
	ApplicationStatus   string `json:"applicationStatus,omitempty"`
}

type GameProgressDTO struct {
	Feedbacks  []games.ProgressFeedback `json:"feedbacks"`
	Milestones []games.Milestone        `json:"milestones"`
	Checkins   []games.Checkin          `json:"checkins"`
}

type GameIMDTO struct {
	Available     bool   `json:"available"`
	RoomID        int64  `json:"roomId,omitempty"`
	Status        string `json:"status,omitempty"`
	Engine        string `json:"engine,omitempty"`
	OpenIMGroupID string `json:"openIMGroupId,omitempty"`
}

type GameReviewDTO struct {
	Reviewable bool           `json:"reviewable"`
	Complete   bool           `json:"complete"`
	Todos      []reviews.Todo `json:"todos,omitempty"`
}

type GameMemberDTO struct {
	UserID        int64  `json:"userId"`
	Name          string `json:"name,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	RealName      string `json:"realName,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	AvatarURL     string `json:"avatarUrl,omitempty"`
	AvatarText    string `json:"avatarText,omitempty"`
	Role          string `json:"role"`
	RoleLabel     string `json:"roleLabel,omitempty"`
	Position      string `json:"position,omitempty"`
	Topic         string `json:"topic,omitempty"`
	PrimaryTag    string `json:"primaryTag,omitempty"`
	Location      string `json:"location,omitempty"`
	IsCreator     bool   `json:"isCreator"`
	IsCurrentUser bool   `json:"isCurrentUser"`
	Confirmed     bool   `json:"confirmed"`
}

type ServiceConfirmDTO struct {
	Confirm games.ServiceConfirm       `json:"confirm"`
	Items   []games.ServiceConfirmItem `json:"items"`
}

type gameCategoryOptionDTO struct {
	Key        string                  `json:"key"`
	Name       string                  `json:"name"`
	Icon       string                  `json:"icon,omitempty"`
	Visible    bool                    `json:"visible"`
	Order      int                     `json:"order"`
	Children   []gameCategoryOptionDTO `json:"children,omitempty"`
	Selectable bool                    `json:"selectable,omitempty"`
}

type gameCategoryConfigDTO struct {
	PrimaryCategories        []gameCategoryOptionDTO `json:"primaryCategories"`
	TypeFilters              []gameCategoryOptionDTO `json:"typeFilters"`
	LocationFilters          []gameCategoryOptionDTO `json:"locationFilters"`
	SortOptions              []gameHallSortOptionDTO `json:"sortOptions,omitempty"`
	EventActions             []string                `json:"eventActions,omitempty"`
	DefaultPrimaryCategory   string                  `json:"defaultPrimaryCategory"`
	DefaultSecondaryCategory string                  `json:"defaultSecondaryCategory"`
	DefaultType              string                  `json:"defaultType"`
	CreateForm               gameCreateFormConfigDTO `json:"createForm"`
	Version                  string                  `json:"version"`
}

type gameHallSortOptionDTO struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	SortKey   string `json:"sortKey"`
	SortOrder string `json:"sortOrder"`
}

type gameCreateCapacityConfigDTO struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type gameCreateFormOptionDTO struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Active bool   `json:"active,omitempty"`
}

type gameCreateFormConfigDTO struct {
	Capacity            gameCreateCapacityConfigDTO `json:"capacity"`
	CurrentLocationText string                      `json:"currentLocationText"`
	ParticipationModes  []gameCreateFormOptionDTO   `json:"participationModes"`
	Tags                []gameCreateFormOptionDTO   `json:"tags"`
	CompletionRules     []gameCreateFormOptionDTO   `json:"completionRules"`
	FeeTypes            []gameCreateFormOptionDTO   `json:"feeTypes"`
}

const gameCategoryConfigKey = "game.category_config"
const gameApplicationConfigKey = "game.application_config"
const gameAuditConfigKey = "game.audit_config"
const gameConditionRuleConfigKey = "game.condition_rule_config"
const gameCancelConfigKey = "game.cancel_config"
const gameDeliveryPageConfigKey = "game.delivery_page_config"
const gameMyGamesPageConfigKey = "game.my_games_page_config"

type gameApplicationConfigDTO struct {
	AgreementTitle      string                            `json:"agreementTitle"`
	AgreementText       string                            `json:"agreementText"`
	RequireRealname     bool                              `json:"requireRealname"`
	RequireIntro        bool                              `json:"requireIntro"`
	RequireAgreement    bool                              `json:"requireAgreement"`
	AllowDuplicateApply bool                              `json:"allowDuplicateApply"`
	UploadRequired      bool                              `json:"uploadRequired"`
	MaxUploadCount      int                               `json:"maxUploadCount"`
	AllowedUploadTypes  []string                          `json:"allowedUploadTypes"`
	MinIntroLength      int                               `json:"minIntroLength"`
	MaxIntroLength      int                               `json:"maxIntroLength"`
	MaxMessageLength    int                               `json:"maxMessageLength"`
	SearchEnabled       bool                              `json:"searchEnabled"`
	RecommendationHint  string                            `json:"recommendationHint"`
	Texts               map[string]string                 `json:"texts,omitempty"`
	AuditPage           gameApplicationAuditPageConfigDTO `json:"auditPage,omitempty"`
	Version             string                            `json:"version"`
}

type gameApplicationAuditFilterDTO struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type gameApplicationAuditPageConfigDTO struct {
	PageTitle   string                              `json:"pageTitle"`
	Filters     []gameApplicationAuditFilterDTO     `json:"filters"`
	StatusTexts map[string]string                   `json:"statusTexts"`
	RoleNames   map[string]string                   `json:"roleNames"`
	Texts       map[string]string                   `json:"texts"`
	Detail      gameApplicationAuditDetailConfigDTO `json:"detail,omitempty"`
}

type gameApplicationAuditActionDTO struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	IconSrc string `json:"iconSrc,omitempty"`
}

type gameApplicationAuditSessionItemDTO struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	IconText   string `json:"iconText,omitempty"`
	IconSrc    string `json:"iconSrc,omitempty"`
	IconClass  string `json:"iconClass,omitempty"`
	ActionText string `json:"actionText,omitempty"`
}

type gameApplicationAuditDetailConfigDTO struct {
	PageTitle         string                               `json:"pageTitle"`
	ReferralText      string                               `json:"referralText"`
	StatusTitles      map[string]string                    `json:"statusTitles"`
	CountdownTexts    map[string]string                    `json:"countdownTexts"`
	PlayerStatusTexts map[string]string                    `json:"playerStatusTexts"`
	Texts             map[string]string                    `json:"texts"`
	SessionItems      []gameApplicationAuditSessionItemDTO `json:"sessionItems"`
	ConfirmRows       []gameApplicationAuditSessionItemDTO `json:"confirmRows"`
	OptionalActions   []gameApplicationAuditActionDTO      `json:"optionalActions"`
	NoticeBullets     []string                             `json:"noticeBullets"`
}

type gameAuditConfigDTO struct {
	AutoApproveFreeGames         bool     `json:"autoApproveFreeGames"`
	RequireManualAuditTypes      []string `json:"requireManualAuditTypes"`
	RequiredRejectReason         bool     `json:"requiredRejectReason"`
	AllowUserResubmitAfterReject bool     `json:"allowUserResubmitAfterReject"`
	BatchAuditMaxCount           int      `json:"batchAuditMaxCount"`
	ApplicationAuditMode         string   `json:"applicationAuditMode"`
	ReviewerRoles                []string `json:"reviewerRoles"`
	Version                      string   `json:"version"`
}

type conditionRuleItemDTO struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Order       int    `json:"order"`
}

type gameConditionRuleConfigDTO struct {
	Enabled              bool                   `json:"enabled"`
	VisibleInMiniProgram bool                   `json:"visibleInMiniProgram"`
	AdminOnlyCreate      bool                   `json:"adminOnlyCreate"`
	RuleItems            []conditionRuleItemDTO `json:"ruleItems"`
	DefaultVisibility    string                 `json:"defaultVisibility"`
	ReviewRequired       bool                   `json:"reviewRequired"`
	PaymentRequired      bool                   `json:"paymentRequired"`
	Version              string                 `json:"version"`
}

type gameCancelRoleConfigDTO struct {
	ReasonOptions  []map[string]string `json:"reasonOptions"`
	DefaultReason  string              `json:"defaultReason"`
	AgreementText  string              `json:"agreementText"`
	AgreementItems []string            `json:"agreementItems,omitempty"`
}

type gameCancelConfigDTO struct {
	Player  gameCancelRoleConfigDTO `json:"player"`
	Expert  gameCancelRoleConfigDTO `json:"expert"`
	Version string                  `json:"version"`
}

type deliveryStatusConfigDTO struct {
	Theme string `json:"theme"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type deliveryStatePillConfigDTO struct {
	Theme string `json:"theme"`
	Text  string `json:"text"`
}

type deliveryNoticePartDTO struct {
	Text   string `json:"text"`
	Strong bool   `json:"strong,omitempty"`
}

type deliveryNoticeConfigDTO struct {
	IconText string                  `json:"iconText,omitempty"`
	Title    string                  `json:"title,omitempty"`
	Parts    []deliveryNoticePartDTO `json:"parts,omitempty"`
}

type deliveryConfirmItemDTO struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Checked bool   `json:"checked"`
	Locked  bool   `json:"locked,omitempty"`
}

type deliverySubmitHintsDTO struct {
	Ready   string `json:"ready"`
	Pending string `json:"pending"`
}

type deliverySecurityConfigDTO struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type deliveryQuickActionDTO struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	Theme    string `json:"theme"`
	IconText string `json:"iconText,omitempty"`
	IconSrc  string `json:"iconSrc,omitempty"`
}

type deliveryModeConfigDTO struct {
	PageTitle         string                     `json:"pageTitle"`
	Status            deliveryStatusConfigDTO    `json:"status"`
	StatePill         deliveryStatePillConfigDTO `json:"statePill"`
	Notice            deliveryNoticeConfigDTO    `json:"notice"`
	ConfirmItems      []deliveryConfirmItemDTO   `json:"confirmItems"`
	ConfirmNote       string                     `json:"confirmNote"`
	Security          deliverySecurityConfigDTO  `json:"security"`
	SubmitHints       deliverySubmitHintsDTO     `json:"submitHints"`
	SubmitToast       string                     `json:"submitToast"`
	SubmitLoadingText string                     `json:"submitLoadingText"`
	AmountRowLabel    string                     `json:"amountRowLabel"`
}

type gameDeliveryPageConfigDTO struct {
	Paid         deliveryModeConfigDTO    `json:"paid"`
	Free         deliveryModeConfigDTO    `json:"free"`
	QuickActions []deliveryQuickActionDTO `json:"quickActions"`
	Version      string                   `json:"version"`
}

func (s *Server) gameCategoryConfigHandler(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentGameCategoryConfig())
}

func (s *Server) gameApplicationConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentGameApplicationConfig())
}

func (s *Server) gameConditionRuleConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentGameConditionRuleConfig())
}

func (s *Server) gameCancelConfig(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentGameCancelConfig())
}

func defaultGameCategoryConfig() gameCategoryConfigDTO {
	primaryCategories := []gameCategoryOptionDTO{
		{
			Key:     "social",
			Name:    "社交局",
			Icon:    "category-social",
			Visible: true,
			Order:   10,
			Children: []gameCategoryOptionDTO{
				categoryChild("meal", "饭局", 10),
				categoryChild("board_game", "桌游局", 20),
				categoryChild("friend", "交友局", 30),
				categoryChild("walk", "同城散步局", 40),
			},
		},
		{
			Key:     "task",
			Name:    "任务局",
			Icon:    "category-task",
			Visible: true,
			Order:   20,
			Children: []gameCategoryOptionDTO{
				categoryChild("partner", "找合伙人", 10),
				categoryChild("project", "做项目", 20),
				categoryChild("brainstorm", "头脑风暴", 30),
				categoryChild("cowork", "组队共创", 40),
			},
		},
		{
			Key:     "explore",
			Name:    "探索局",
			Icon:    "category-income",
			Visible: true,
			Order:   30,
			Children: []gameCategoryOptionDTO{
				categoryChild("city_explore", "城市探索", 10),
				categoryChild("route_blind_box", "路线盲盒", 20),
				categoryChild("checkin_challenge", "打卡挑战", 30),
				categoryChild("night_walk", "夜游/徒步/骑行", 40),
			},
		},
		{
			Key:     "growth",
			Name:    "成长局",
			Icon:    "category-growth",
			Visible: true,
			Order:   40,
			Children: []gameCategoryOptionDTO{
				categoryChild("reading", "读书局", 10),
				categoryChild("fitness", "健身局", 20),
				categoryChild("checkin", "打卡局", 30),
				categoryChild("deposit_checkin", "押金局", 40),
				categoryChild("study", "学习共修局", 50),
			},
		},
	}
	return gameCategoryConfigDTO{
		PrimaryCategories: primaryCategories,
		TypeFilters: []gameCategoryOptionDTO{
			{Key: "all", Name: "类型", Visible: true, Order: 0, Selectable: true},
			{Key: "free", Name: "免费局", Visible: true, Order: 10, Selectable: true},
			{Key: "standard", Name: "标准局", Visible: true, Order: 20, Selectable: true},
			{Key: "aa", Name: "AA局", Visible: true, Order: 30, Selectable: true},
			{Key: "crowdfund", Name: "众筹局", Visible: true, Order: 40, Selectable: true},
			{Key: "deposit", Name: "押金局", Visible: true, Order: 50, Selectable: true},
			{Key: "public_welfare", Name: "公益局", Visible: true, Order: 60, Selectable: false},
		},
		LocationFilters: []gameCategoryOptionDTO{
			{Key: "all", Name: "全国", Visible: true, Order: 0, Selectable: true},
			{Key: "nearby", Name: "附近(50km)", Visible: true, Order: 10, Selectable: true},
		},
		SortOptions: []gameHallSortOptionDTO{
			{Key: "comprehensive", Name: "综合排序", SortKey: "", SortOrder: "asc"},
			{Key: "latest", Name: "最新发布", SortKey: "time", SortOrder: "desc"},
			{Key: "hot", Name: "热度最高", SortKey: "hot", SortOrder: "desc"},
			{Key: "distance", Name: "距离最近", SortKey: "distance", SortOrder: "asc"},
			{Key: "credit", Name: "信用优先", SortKey: "credit", SortOrder: "desc"},
		},
		EventActions:             []string{"分享", "关注", "引荐", "打招呼"},
		DefaultPrimaryCategory:   "task",
		DefaultSecondaryCategory: "project",
		DefaultType:              "free",
		CreateForm:               defaultGameCreateFormConfig(),
		Version:                  "2026-06-30",
	}
}

func defaultGameCreateFormConfig() gameCreateFormConfigDTO {
	return gameCreateFormConfigDTO{
		Capacity:            gameCreateCapacityConfigDTO{Min: games.MinGamePlayers, Max: games.MaxGamePlayers},
		CurrentLocationText: "当前位置",
		ParticipationModes: []gameCreateFormOptionDTO{
			{Key: "online", Name: "线上"},
			{Key: "offline", Name: "线下"},
			{Key: "hybrid", Name: "混合"},
		},
		Tags: []gameCreateFormOptionDTO{
			{Key: "product", Name: "产品研发"},
			{Key: "startup", Name: "创业"},
			{Key: "city_explore", Name: "城市探索"},
			{Key: "cocreation", Name: "共创"},
		},
		CompletionRules: []gameCreateFormOptionDTO{
			{Key: "time", Name: "时间截止"},
			{Key: "goal", Name: "目标达成", Active: true},
			{Key: "capacity", Name: "人数满额"},
			{Key: "manual", Name: "手动结束", Active: true},
		},
		FeeTypes: []gameCreateFormOptionDTO{
			{Key: "free", Name: "免费局"},
			{Key: "paid", Name: "收费局"},
		},
	}
}

func categoryChild(key string, name string, order int) gameCategoryOptionDTO {
	return gameCategoryOptionDTO{Key: key, Name: name, Visible: true, Order: order, Selectable: true}
}

func (s *Server) adminGameCategoryConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameCategoryConfig()})
	case http.MethodPut:
		var req gameCategoryConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid category config")
			return
		}
		config, err := normalizeGameCategoryConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.setGameCategoryConfig(config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save category config failed")
			return
		}
		s.recordOperation(r, "game_category_config:update", "system_config", "game_category_config", map[string]interface{}{
			"primaryCategoryCount": len(config.PrimaryCategories),
			"typeFilterCount":      len(config.TypeFilters),
			"locationFilterCount":  len(config.LocationFilters),
			"version":              config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameCategoryConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) adminGameApplicationConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameApplicationConfig()})
	case http.MethodPut:
		var req gameApplicationConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid application config")
			return
		}
		config, err := normalizeGameApplicationConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameApplicationConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save application config failed")
			return
		}
		s.recordOperation(r, "game_application_config:update", "system_config", "game_application_config", map[string]interface{}{
			"requireAgreement": config.RequireAgreement,
			"maxUploadCount":   config.MaxUploadCount,
			"version":          config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameApplicationConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) adminGameAuditConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameAuditConfig()})
	case http.MethodPut:
		var req gameAuditConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid audit config")
			return
		}
		config, err := normalizeGameAuditConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameAuditConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save audit config failed")
			return
		}
		s.recordOperation(r, "game_audit_config:update", "system_config", "game_audit_config", map[string]interface{}{
			"autoApproveFreeGames":    config.AutoApproveFreeGames,
			"requireManualAuditTypes": config.RequireManualAuditTypes,
			"version":                 config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameAuditConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) adminGameConditionRuleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameConditionRuleConfig()})
	case http.MethodPut:
		var req gameConditionRuleConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid condition rule config")
			return
		}
		config, err := normalizeGameConditionRuleConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameConditionRuleConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "save condition rule config failed")
			return
		}
		s.recordOperation(r, "game_condition_rule_config:update", "system_config", "game_condition_rule_config", map[string]interface{}{
			"enabled":       config.Enabled,
			"ruleItemCount": len(config.RuleItems),
			"version":       config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameConditionRuleConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "method not allowed")
	}
}

func (s *Server) currentGameCategoryConfig() gameCategoryConfigDTO {
	var stored gameCategoryConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameCategoryConfigKey, &stored) && len(stored.PrimaryCategories) > 0 {
		return cloneGameCategoryConfig(stored)
	}
	s.gameCategoryConfigMu.RLock()
	if len(s.gameCategoryConfig.PrimaryCategories) > 0 {
		config := cloneGameCategoryConfig(s.gameCategoryConfig)
		s.gameCategoryConfigMu.RUnlock()
		return config
	}
	s.gameCategoryConfigMu.RUnlock()
	return cloneGameCategoryConfig(defaultGameCategoryConfig())
}

func (s *Server) currentGameApplicationConfig() gameApplicationConfigDTO {
	var stored gameApplicationConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameApplicationConfigKey, &stored) && strings.TrimSpace(stored.AgreementTitle) != "" {
		return mergeGameApplicationConfigDefaults(stored)
	}
	return defaultGameApplicationConfig()
}

func (s *Server) currentGameAuditConfig() gameAuditConfigDTO {
	var stored gameAuditConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameAuditConfigKey, &stored) && strings.TrimSpace(stored.ApplicationAuditMode) != "" {
		return cloneGameAuditConfig(stored)
	}
	return defaultGameAuditConfig()
}

func (s *Server) currentGameConditionRuleConfig() gameConditionRuleConfigDTO {
	var stored gameConditionRuleConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameConditionRuleConfigKey, &stored) && len(stored.RuleItems) > 0 {
		return cloneGameConditionRuleConfig(stored)
	}
	return defaultGameConditionRuleConfig()
}

func (s *Server) currentGameCancelConfig() gameCancelConfigDTO {
	var stored gameCancelConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameCancelConfigKey, &stored) && len(stored.Player.ReasonOptions) > 0 && len(stored.Expert.ReasonOptions) > 0 {
		return stored
	}
	return defaultGameCancelConfig()
}

func (s *Server) currentGameDeliveryPageConfig() gameDeliveryPageConfigDTO {
	var stored gameDeliveryPageConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameDeliveryPageConfigKey, &stored) && strings.TrimSpace(stored.Paid.PageTitle) != "" && strings.TrimSpace(stored.Free.PageTitle) != "" {
		return stored
	}
	return defaultGameDeliveryPageConfig()
}

func (s *Server) ensureDefaultSystemConfigs() {
	if s.systemConfig == nil {
		return
	}
	var stored gameCategoryConfigDTO
	if !s.systemConfig.Get(gameCategoryConfigKey, &stored) || len(stored.PrimaryCategories) == 0 {
		_ = s.systemConfig.Set(gameCategoryConfigKey, defaultGameCategoryConfig())
	}
	var homeStored homeDisplayConfigDTO
	if !s.systemConfig.Get(homeDisplayConfigKey, &homeStored) {
		_ = s.systemConfig.Set(homeDisplayConfigKey, defaultHomeDisplayConfig())
	}
	var applicationStored gameApplicationConfigDTO
	if !s.systemConfig.Get(gameApplicationConfigKey, &applicationStored) {
		_ = s.systemConfig.Set(gameApplicationConfigKey, defaultGameApplicationConfig())
	}
	var auditStored gameAuditConfigDTO
	if !s.systemConfig.Get(gameAuditConfigKey, &auditStored) {
		_ = s.systemConfig.Set(gameAuditConfigKey, defaultGameAuditConfig())
	}
	var conditionStored gameConditionRuleConfigDTO
	if !s.systemConfig.Get(gameConditionRuleConfigKey, &conditionStored) {
		_ = s.systemConfig.Set(gameConditionRuleConfigKey, defaultGameConditionRuleConfig())
	}
	var cancelStored gameCancelConfigDTO
	if !s.systemConfig.Get(gameCancelConfigKey, &cancelStored) {
		_ = s.systemConfig.Set(gameCancelConfigKey, defaultGameCancelConfig())
	}
	var roleStatusStored map[string]interface{}
	if !s.systemConfig.Get(roleStatusPageConfigKey, &roleStatusStored) || len(roleStatusStored) == 0 {
		_ = s.systemConfig.Set(roleStatusPageConfigKey, defaultRoleStatusPageConfig())
	}
	var roleApplicationPageStored map[string]interface{}
	if !s.systemConfig.Get(roleApplicationPageConfigKey, &roleApplicationPageStored) || len(roleApplicationPageStored) == 0 {
		_ = s.systemConfig.Set(roleApplicationPageConfigKey, defaultRoleApplicationPageConfig())
	}
	var deliveryPageStored gameDeliveryPageConfigDTO
	if !s.systemConfig.Get(gameDeliveryPageConfigKey, &deliveryPageStored) || strings.TrimSpace(deliveryPageStored.Paid.PageTitle) == "" {
		_ = s.systemConfig.Set(gameDeliveryPageConfigKey, defaultGameDeliveryPageConfig())
	}
	var reportStored reportCenterConfigDTO
	if !s.systemConfig.Get(reportCenterConfigKey, &reportStored) {
		_ = s.systemConfig.Set(reportCenterConfigKey, defaultReportCenterConfig())
	}
	var pointsStored pointsPageConfigDTO
	if !s.systemConfig.Get(pointsPageConfigKey, &pointsStored) {
		_ = s.systemConfig.Set(pointsPageConfigKey, defaultPointsPageConfig())
	}
	var assetStored profileAssetManageConfigDTO
	if !s.systemConfig.Get(profileAssetManageConfigKey, &assetStored) {
		_ = s.systemConfig.Set(profileAssetManageConfigKey, defaultProfileAssetManageConfig())
	}
	var redemptionOrderStored redemptionOrderPageConfigDTO
	if !s.systemConfig.Get(redemptionOrderPageConfigKey, &redemptionOrderStored) {
		_ = s.systemConfig.Set(redemptionOrderPageConfigKey, defaultRedemptionOrderPageConfig())
	}
	var reviewPageStored reviewPageConfigDTO
	if !s.systemConfig.Get(reviewPageConfigKey, &reviewPageStored) || strings.TrimSpace(reviewPageStored.NavTitle) == "" {
		_ = s.systemConfig.Set(reviewPageConfigKey, defaultReviewPageConfig())
	}
	var mapMyCityStored mapMyCityConfigDTO
	if !s.systemConfig.Get(mapMyCityConfigKey, &mapMyCityStored) || len(mapMyCityStored.StoryGroups) == 0 {
		_ = s.systemConfig.Set(mapMyCityConfigKey, defaultMapMyCityConfig())
	}
	var mapIndexStored mapIndexConfigDTO
	if !s.systemConfig.Get(mapIndexConfigKey, &mapIndexStored) || len(mapIndexStored.MapFilters) == 0 {
		_ = s.systemConfig.Set(mapIndexConfigKey, defaultMapIndexConfig())
	}
	var systemRecommendationsStored gameSystemRecommendationsConfigDTO
	if !s.systemConfig.Get(gameSystemRecommendationsConfigKey, &systemRecommendationsStored) || strings.TrimSpace(systemRecommendationsStored.Title) == "" {
		_ = s.systemConfig.Set(gameSystemRecommendationsConfigKey, defaultGameSystemRecommendationsConfig())
	}
	var referralRecordsStored map[string]interface{}
	if !s.systemConfig.Get(gameReferralRecordsConfigKey, &referralRecordsStored) || len(referralRecordsStored) == 0 {
		_ = s.systemConfig.Set(gameReferralRecordsConfigKey, defaultGameReferralRecordsConfig())
	}
	var tradeWarningStored map[string]interface{}
	if !s.systemConfig.Get(tradeWarningConfigKey, &tradeWarningStored) || len(tradeWarningStored) == 0 {
		_ = s.systemConfig.Set(tradeWarningConfigKey, defaultTradeWarningConfig())
	}
	var systemNotificationStored map[string]interface{}
	if !s.systemConfig.Get(systemNotificationConfigKey, &systemNotificationStored) || len(systemNotificationStored) == 0 {
		_ = s.systemConfig.Set(systemNotificationConfigKey, defaultSystemNotificationConfig())
	}
	var messageCenterStored map[string]interface{}
	if !s.systemConfig.Get(messageCenterConfigKey, &messageCenterStored) || len(messageCenterStored) == 0 {
		_ = s.systemConfig.Set(messageCenterConfigKey, cloneMap(defaultMessageCenterConfig))
	}
	var messageMyStored map[string]interface{}
	if !s.systemConfig.Get(messageMyConfigKey, &messageMyStored) || len(messageMyStored) == 0 {
		_ = s.systemConfig.Set(messageMyConfigKey, cloneMap(defaultMessageMyConfig))
	}
}

func (s *Server) setGameCategoryConfig(config gameCategoryConfigDTO) error {
	if s.systemConfig != nil {
		if err := s.systemConfig.Set(gameCategoryConfigKey, config); err != nil {
			return err
		}
	}
	s.gameCategoryConfigMu.Lock()
	s.gameCategoryConfig = cloneGameCategoryConfig(config)
	s.gameCategoryConfigMu.Unlock()
	return nil
}

func normalizeGameCategoryConfig(req gameCategoryConfigDTO) (gameCategoryConfigDTO, error) {
	config := req
	config.PrimaryCategories = normalizeCategoryOptions(req.PrimaryCategories, true)
	config.TypeFilters = normalizeCategoryOptions(req.TypeFilters, true)
	config.LocationFilters = normalizeCategoryOptions(req.LocationFilters, false)
	config.DefaultPrimaryCategory = strings.TrimSpace(req.DefaultPrimaryCategory)
	config.DefaultSecondaryCategory = strings.TrimSpace(req.DefaultSecondaryCategory)
	config.DefaultType = strings.TrimSpace(req.DefaultType)
	config.Version = strings.TrimSpace(req.Version)
	if len(config.PrimaryCategories) == 0 {
		return gameCategoryConfigDTO{}, errors.New("primaryCategories required")
	}
	if len(config.TypeFilters) == 0 {
		return gameCategoryConfigDTO{}, errors.New("typeFilters required")
	}
	if key := unsupportedGameTypeFilter(config.TypeFilters); key != "" {
		return gameCategoryConfigDTO{}, errors.New("unsupported game type: " + key)
	}
	if config.DefaultPrimaryCategory == "" {
		config.DefaultPrimaryCategory = config.PrimaryCategories[0].Key
	}
	if config.DefaultType == "" {
		config.DefaultType = firstCreatableGameType(config.TypeFilters)
	}
	config.CreateForm = normalizeGameCreateFormConfig(config.CreateForm)
	if !categoryKeyExists(config.PrimaryCategories, config.DefaultPrimaryCategory) {
		return gameCategoryConfigDTO{}, errors.New("defaultPrimaryCategory not found")
	}
	if config.DefaultSecondaryCategory != "" && !categoryChildKeyExists(config.PrimaryCategories, config.DefaultPrimaryCategory, config.DefaultSecondaryCategory) {
		return gameCategoryConfigDTO{}, errors.New("defaultSecondaryCategory not found")
	}
	if !categoryKeyExists(config.TypeFilters, config.DefaultType) {
		return gameCategoryConfigDTO{}, errors.New("defaultType not found")
	}
	if !validConfigGameType(config.DefaultType) {
		return gameCategoryConfigDTO{}, errors.New("defaultType must be a creatable game type")
	}
	if config.Version == "" {
		config.Version = time.Now().UTC().Format("2006-01-02")
	}
	return config, nil
}

func normalizeGameCreateFormConfig(config gameCreateFormConfigDTO) gameCreateFormConfigDTO {
	defaults := defaultGameCreateFormConfig()
	if config.Capacity.Min <= 0 {
		config.Capacity.Min = defaults.Capacity.Min
	}
	if config.Capacity.Max <= 0 {
		config.Capacity.Max = defaults.Capacity.Max
	}
	if config.Capacity.Min < games.MinGamePlayers {
		config.Capacity.Min = games.MinGamePlayers
	}
	if config.Capacity.Max > games.MaxGamePlayers {
		config.Capacity.Max = games.MaxGamePlayers
	}
	if config.Capacity.Max < config.Capacity.Min {
		config.Capacity.Max = config.Capacity.Min
	}
	config.CurrentLocationText = strings.TrimSpace(config.CurrentLocationText)
	if config.CurrentLocationText == "" {
		config.CurrentLocationText = defaults.CurrentLocationText
	}
	config.ParticipationModes = normalizeCreateFormOptions(config.ParticipationModes, defaults.ParticipationModes)
	config.Tags = normalizeCreateFormOptions(config.Tags, defaults.Tags)
	config.CompletionRules = normalizeCreateFormOptions(config.CompletionRules, defaults.CompletionRules)
	config.FeeTypes = normalizeCreateFormOptions(config.FeeTypes, defaults.FeeTypes)
	return config
}

func normalizeCreateFormOptions(items []gameCreateFormOptionDTO, defaults []gameCreateFormOptionDTO) []gameCreateFormOptionDTO {
	result := make([]gameCreateFormOptionDTO, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		name := strings.TrimSpace(item.Name)
		if key == "" || name == "" || len(key) > 64 || len(name) > 64 {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, gameCreateFormOptionDTO{Key: key, Name: name, Active: item.Active})
	}
	if len(result) == 0 {
		return append([]gameCreateFormOptionDTO(nil), defaults...)
	}
	return result
}

func unsupportedGameTypeFilter(items []gameCategoryOptionDTO) string {
	for _, item := range items {
		if item.Key == "all" {
			continue
		}
		if !validConfigGameType(item.Key) {
			return item.Key
		}
	}
	return ""
}

func firstCreatableGameType(items []gameCategoryOptionDTO) string {
	for _, item := range items {
		if validConfigGameType(item.Key) {
			return item.Key
		}
	}
	if len(items) > 0 {
		return items[0].Key
	}
	return ""
}

func validConfigGameType(value string) bool {
	switch value {
	case "free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition":
		return true
	default:
		return false
	}
}

func normalizeCategoryOptions(items []gameCategoryOptionDTO, selectable bool) []gameCategoryOptionDTO {
	result := make([]gameCategoryOptionDTO, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		name := strings.TrimSpace(item.Name)
		if key == "" || name == "" || len(key) > 64 || len(name) > 64 {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		item.Key = key
		item.Name = name
		item.Icon = strings.TrimSpace(item.Icon)
		if selectable && !item.Selectable {
			item.Selectable = true
		}
		item.Children = normalizeCategoryOptions(item.Children, true)
		result = append(result, item)
	}
	return result
}

func categoryKeyExists(items []gameCategoryOptionDTO, key string) bool {
	for _, item := range items {
		if item.Key == key {
			return true
		}
	}
	return false
}

func categoryChildKeyExists(items []gameCategoryOptionDTO, primaryKey string, childKey string) bool {
	for _, item := range items {
		if item.Key != primaryKey {
			continue
		}
		for _, child := range item.Children {
			if child.Key == childKey {
				return true
			}
		}
	}
	return false
}

func cloneGameCategoryConfig(config gameCategoryConfigDTO) gameCategoryConfigDTO {
	config.PrimaryCategories = cloneGameCategoryOptions(config.PrimaryCategories)
	config.TypeFilters = cloneGameCategoryOptions(config.TypeFilters)
	config.LocationFilters = cloneGameCategoryOptions(config.LocationFilters)
	config.CreateForm = normalizeGameCreateFormConfig(config.CreateForm)
	return config
}

func cloneGameCategoryOptions(items []gameCategoryOptionDTO) []gameCategoryOptionDTO {
	if len(items) == 0 {
		return nil
	}
	result := make([]gameCategoryOptionDTO, len(items))
	for i, item := range items {
		result[i] = item
		result[i].Children = cloneGameCategoryOptions(item.Children)
	}
	return result
}

func defaultGameApplicationConfig() gameApplicationConfigDTO {
	return gameApplicationConfigDTO{
		AgreementTitle:      "入局申请须知",
		AgreementText:       "申请入局前请确认本人已完成实名，了解局的主题、地点、时间和成员规则。申请通过后请按约参与，临时退出可能影响信用分。",
		RequireRealname:     true,
		RequireIntro:        true,
		RequireAgreement:    true,
		AllowDuplicateApply: false,
		UploadRequired:      false,
		MaxUploadCount:      3,
		AllowedUploadTypes:  []string{"jpg", "png", "pdf"},
		MinIntroLength:      5,
		MaxIntroLength:      200,
		MaxMessageLength:    120,
		SearchEnabled:       false,
		RecommendationHint:  "一期优先展示审核通过且人数未满的局。",
		Texts: map[string]string{
			"subtitle":              "你的信息将展示给发起人",
			"wechatTitle":           "微信信息",
			"nicknameLabel":         "昵称",
			"introLabel":            "自我介绍",
			"introPlaceholder":      "介绍你的背景、能力和参与动机",
			"portfolioLabel":        "相关经历/作品",
			"messageLabel":          "申请留言",
			"messagePlaceholder":    "给发起人留一句话",
			"agreementPrefix":       "我已阅读并同意",
			"cancelText":            "取消",
			"submitText":            "提交申请",
			"profileSyncedText":     "资料已同步",
			"profilePendingText":    "资料待同步",
			"profileNameFallback":   "待同步",
			"loadFailedText":        "入局申请配置加载失败",
			"mediaUnsupportedText":  "当前微信版本不支持选择图片",
			"fileUnsupportedText":   "当前微信版本不支持选择文件",
			"imageTypeErrorText":    "仅支持 JPG、PNG、GIF、WEBP 图片",
			"fileTypeErrorText":     "仅支持 PDF 文件",
			"imageSelectedText":     "图片已选择",
			"fileSelectedText":      "文件已选择",
			"chooseFailedText":      "选择失败，请重试",
			"introRequiredText":     "请先填写自我介绍",
			"introMinTemplate":      "自我介绍不少于{min}字",
			"agreementRequiredText": "请先勾选平台协议",
			"uploadRequiredText":    "请先上传相关经历/作品",
			"gameMissingText":       "缺少局信息",
			"submittingText":        "提交中",
			"submitSuccessText":     "申请已提交",
			"submitFailedText":      "提交失败，请重试",
			"maxUploadTemplate":     "最多上传{max}个文件",
			"navUnavailableText":    "当前页暂无左右切换",
		},
		AuditPage: gameApplicationAuditPageConfigDTO{
			PageTitle: "审核申请列表",
			Filters: []gameApplicationAuditFilterDTO{
				{Key: "all", Name: "全部"},
				{Key: "pending", Name: "待审核"},
				{Key: "approved", Name: "已通过"},
				{Key: "rejected", Name: "已拒绝"},
			},
			StatusTexts: map[string]string{
				"pending":  "待审核",
				"approved": "已通过",
				"rejected": "已拒绝",
			},
			RoleNames: map[string]string{
				"expert":     "行家",
				"guide":      "领路人",
				"main_guide": "主行家",
				"player":     "玩家",
				"member":     "玩家",
			},
			Texts: map[string]string{
				"userFallbackTemplate":  "用户{userId}",
				"avatarFallback":        "玩",
				"applyTimeLabel":        "申请时间",
				"approveText":           "通过申请",
				"rejectText":            "拒绝",
				"reviewedText":          "已完成审核",
				"detailText":            "详情",
				"selectAllText":         "全选",
				"batchRejectText":       "批量拒绝",
				"batchApproveText":      "批量通过",
				"loadFailedText":        "申请列表加载失败",
				"approvingText":         "通过中",
				"rejectingText":         "拒绝中",
				"approveSuccessText":    "已通过申请",
				"rejectSuccessText":     "已拒绝申请",
				"reviewFailedText":      "审核失败",
				"emptyPendingText":      "暂无待审核申请",
				"batchApprovingText":    "批量通过中",
				"batchRejectingText":    "批量拒绝中",
				"batchApproveSuccess":   "已批量通过",
				"batchRejectSuccess":    "已批量拒绝",
				"batchReviewFailedText": "批量审核失败",
			},
			Detail: gameApplicationAuditDetailConfigDTO{
				PageTitle:    "审核组局",
				ReferralText: "已撮合双方意向",
				StatusTitles: map[string]string{
					"pending":  "等待你审核",
					"approved": "已通过申请",
					"rejected": "已拒绝申请",
				},
				CountdownTexts: map[string]string{
					"pending":  "待处理",
					"approved": "已处理",
					"rejected": "已处理",
				},
				PlayerStatusTexts: map[string]string{
					"pendingRequirement":  "待确认需求",
					"reviewedRequirement": "已完成审核",
					"pending":             "等待审核",
					"approved":            "申请已通过",
					"rejected":            "申请已拒绝",
				},
				Texts: map[string]string{
					"playerTitle":              "玩家信息",
					"portfolioTitle":           "相关经历/作品",
					"confirmTitle":             "局信息确认",
					"optionTitle":              "可选操作",
					"noticeTitle":              "确认须知",
					"relationTitle":            "组局关系图",
					"expertName":               "我",
					"expertRoleText":           "审核方",
					"expertAvatarText":         "我",
					"guideAvatarFallback":      "领",
					"playerAvatarFallback":     "玩",
					"needPrefix":               "申请说明：",
					"remarkPrefix":             "申请时间：",
					"detailMissingText":        "申请详情不存在",
					"loadFailedText":           "申请详情加载失败",
					"emptyTitle":               "申请详情未加载",
					"emptyText":                "请确认审核入口携带的申请 ID 是否有效，或返回申请列表重新打开。",
					"mediaUnsupportedText":     "当前微信版本不支持选择图片",
					"fileUnsupportedText":      "当前微信版本不支持选择文件",
					"imageTypeErrorText":       "仅支持 JPG、PNG、GIF、WEBP 图片",
					"fileTypeErrorText":        "仅支持 PDF 文件",
					"filePathInvalidText":      "文件路径无效",
					"uploadingText":            "上传中",
					"imageUploadedText":        "图片已上传",
					"fileUploadedText":         "文件已上传",
					"uploadFailedText":         "上传失败，请重试",
					"chooseFailedText":         "选择失败，请重试",
					"detailRequiredActionText": "申请详情加载后才可以沟通",
					"detailRequiredReviewText": "申请详情加载后才可以审核",
					"chatPrefill":              "你好，我想进一步确认本次组局申请。",
					"timePrefill":              "我建议进一步确认本次组局的具体时间，请看是否方便。",
					"unavailableActionText":    "请选择可用操作",
					"approvingText":            "通过中",
					"rejectingText":            "拒绝中",
					"approveSuccessText":       "已确认通过",
					"rejectSuccessText":        "已拒绝申请",
					"reviewFailedText":         "审核失败",
					"actionTip":                "确认后将建立三方连接群并冻结资金",
					"actionLoadingText":        "处理中...",
					"confirmText":              "确认通过",
				},
				SessionItems: []gameApplicationAuditSessionItemDTO{
					{Key: "topic", Label: "组局主题", IconText: "H", IconClass: "topic"},
					{Key: "time", Label: "时间", IconSrc: "/pages/game/detail/assets/icon-clock.png", IconClass: "time"},
					{Key: "location", Label: "地点", ActionText: "地图位置", IconSrc: "/pages/game/detail/assets/icon-location.png", IconClass: "place"},
				},
				ConfirmRows: []gameApplicationAuditSessionItemDTO{
					{Key: "activityType", Label: "活动类型"},
					{Key: "serviceDuration", Label: "服务时长"},
					{Key: "clientBudget", Label: "客户预算"},
					{Key: "platformFee", Label: "平台"},
					{Key: "guideReward", Label: "领路人"},
					{Key: "partnerReward", Label: "生态合伙人"},
					{Key: "expertIncome", Label: "你的收益"},
				},
				OptionalActions: []gameApplicationAuditActionDTO{
					{Key: "time", Name: "提议具体时间", IconSrc: "/pages/game/audit-detail/assets/option-time.png"},
					{Key: "chat", Name: "与玩家沟通", IconSrc: "/pages/game/audit-detail/assets/option-chat.png"},
				},
				NoticeBullets: []string{
					"确认后请准时参加，如需取消请提前通知",
					"双方确认后组局正式生效，领路人将获得积分奖励",
					"请保持专业态度，维护平台信誉",
				},
			},
		},
		Version: "2026-06-30",
	}
}

func defaultGameAuditConfig() gameAuditConfigDTO {
	return gameAuditConfigDTO{
		AutoApproveFreeGames:         false,
		RequireManualAuditTypes:      []string{"free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition"},
		RequiredRejectReason:         true,
		AllowUserResubmitAfterReject: true,
		BatchAuditMaxCount:           50,
		ApplicationAuditMode:         "creator_or_main_guide",
		ReviewerRoles:                []string{"super_admin", "audit_admin"},
		Version:                      "2026-06-30",
	}
}

func defaultGameConditionRuleConfig() gameConditionRuleConfigDTO {
	return gameConditionRuleConfigDTO{
		Enabled:              true,
		VisibleInMiniProgram: true,
		AdminOnlyCreate:      true,
		RuleItems: []conditionRuleItemDTO{
			{Key: "realname_verified", Name: "完成实名认证", Description: "玩家必须完成实名后才能申请条件局", Required: true, Order: 10},
			{Key: "credit_min_80", Name: "信用分不低于 80", Description: "用于测试条件局的信用门槛", Required: true, Order: 20},
			{Key: "profile_complete", Name: "资料完整", Description: "昵称、头像、城市等资料达到基础完整度", Required: false, Order: 30},
		},
		DefaultVisibility: "approved_users",
		ReviewRequired:    true,
		PaymentRequired:   false,
		Version:           "2026-06-30",
	}
}

func defaultGameCancelConfig() gameCancelConfigDTO {
	return gameCancelConfigDTO{
		Player: gameCancelRoleConfigDTO{
			ReasonOptions: []map[string]string{
				{"key": "need_changed", "text": "需求变更，不再需要服务"},
				{"key": "other_solution", "text": "找到其他解决方案"},
				{"key": "service_unexpected", "text": "业务主服务不符合预期"},
				{"key": "budget", "text": "预算问题/资金紧张"},
			},
			DefaultReason: "other_solution",
			AgreementText: "我已阅读并同意上述赔付协议，理解主动取消需承担行家的时间成本损失，并同意按设置比例从托管资金中赔付行家。",
			AgreementItems: []string{
				"我理解主动取消需承担行家的时间成本损失",
				"我同意按设置比例赔付行家，金额从托管资金扣除",
				"剩余金额将在3个工作日内原路退回",
				"此取消记录将影响信用分（-3分）",
			},
		},
		Expert: gameCancelRoleConfigDTO{
			ReasonOptions: []map[string]string{
				{"key": "schedule_conflict", "text": "个人时间冲突，无法交付"},
				{"key": "requirement_mismatch", "text": "需求与描述不符，无法完成"},
				{"key": "emergency", "text": "身体原因/突发状况"},
				{"key": "other", "text": "其他原因"},
			},
			DefaultReason: "schedule_conflict",
			AgreementText: "我已阅读并同意《服务取消协议》，理解主动取消将对我的信用分产生影响（-5分），并同意按设置比例赔付玩家损失。",
		},
		Version: "2026-07-01",
	}
}

func defaultGameDeliveryPageConfig() gameDeliveryPageConfigDTO {
	return gameDeliveryPageConfigDTO{
		Paid: deliveryModeConfigDTO{
			PageTitle: "确认服务完成",
			Status:    deliveryStatusConfigDTO{Theme: "paid", Title: "服务已完成!", Desc: "双方确认后，资金将全额结算"},
			StatePill: deliveryStatePillConfigDTO{Theme: "green", Text: "待确认完成"},
			Notice:    deliveryNoticeConfigDTO{},
			ConfirmItems: []deliveryConfirmItemDTO{
				{ID: "completed", Title: "服务已全部完成", Desc: "约定的2小时咨询服务已完整交付"},
				{ID: "qualified", Title: "服务质量达标", Desc: "需求方对服务内容和质量无异议"},
				{ID: "communicated", Title: "双方已沟通确认", Desc: "已与需求方确认服务完成，对方同意结算"},
			},
			ConfirmNote:       "正常交付无需扣减任何费用，只需双方确认服务已完成，资金将按全额结算。如服务未完全达标，请与玩家沟通后再确认。",
			Security:          deliverySecurityConfigDTO{},
			SubmitHints:       deliverySubmitHintsDTO{Ready: "确认后将通知玩家进行最终确认", Pending: "需勾选上方确认项后方可提交"},
			SubmitToast:       "服务完成确认已提交",
			SubmitLoadingText: "提交中",
			AmountRowLabel:    "合同金额",
		},
		Free: deliveryModeConfigDTO{
			PageTitle: "确认服务完成",
			Status:    deliveryStatusConfigDTO{Theme: "free", Title: "服务已完成!", Desc: "双方确认后，服务正式结束"},
			StatePill: deliveryStatePillConfigDTO{Theme: "blue", Text: "待确认完成"},
			Notice: deliveryNoticeConfigDTO{
				IconText: "🎁",
				Title:    "免费局说明",
				Parts: []deliveryNoticePartDTO{
					{Text: "本局为"},
					{Text: "免费体验局", Strong: true},
					{Text: "不涉及资金结算。双方确认完成后，行家将获得"},
					{Text: "信用积分+5和免费局贡献徽章", Strong: true},
					{Text: "，玩家"},
					{Text: "优先推荐权益", Strong: true},
				},
			},
			ConfirmItems: []deliveryConfirmItemDTO{
				{ID: "completed", Title: "服务已全部完成", Desc: "约定的2小时咨询服务已完整交付"},
				{ID: "qualified", Title: "服务质量达标", Desc: "需求方对服务内容和质量无异议"},
				{ID: "communicated", Title: "双方已沟通确认", Desc: "已与需求方确认服务完成，对方同意归档"},
			},
			ConfirmNote:       "免费局无需扣除任何费用，只需双方确认服务已完成，系统将自动归档。如服务未完全达标，请与玩家沟通后再次确认。",
			Security:          deliverySecurityConfigDTO{Title: "服务保障", Desc: "免费局同样享受平台服务保障，评价真实有效"},
			SubmitHints:       deliverySubmitHintsDTO{Ready: "确认后将通知玩家进行最终确认", Pending: "需勾选上方确认项后方可提交"},
			SubmitToast:       "免费局服务完成确认已提交",
			SubmitLoadingText: "提交中",
			AmountRowLabel:    "服务类型",
		},
		QuickActions: []deliveryQuickActionDTO{
			{Key: "upload", Title: "上传凭证", Theme: "blue", IconText: "📎"},
			{Key: "contact_player", Title: "联系玩家", Theme: "blue", IconSrc: "/pages/game/delivery/assets/i18@3x.png"},
			{Key: "contact_guide", Title: "联系领路人", Theme: "orange", IconText: "👬"},
		},
		Version: "2026-07-01",
	}
}

func normalizeGameApplicationConfig(req gameApplicationConfigDTO) (gameApplicationConfigDTO, error) {
	config := cloneGameApplicationConfig(req)
	defaults := defaultGameApplicationConfig()
	config.AgreementTitle = strings.TrimSpace(config.AgreementTitle)
	config.AgreementText = strings.TrimSpace(config.AgreementText)
	config.RecommendationHint = strings.TrimSpace(config.RecommendationHint)
	config.AllowedUploadTypes = normalizeStringList(config.AllowedUploadTypes, 16)
	config.Texts = mergeStringMap(defaults.Texts, config.Texts)
	config.AuditPage = mergeGameApplicationAuditPageConfig(defaults.AuditPage, config.AuditPage)
	config.Version = strings.TrimSpace(config.Version)
	if config.AgreementTitle == "" {
		return gameApplicationConfigDTO{}, errors.New("agreementTitle required")
	}
	if config.RequireAgreement && config.AgreementText == "" {
		return gameApplicationConfigDTO{}, errors.New("agreementText required")
	}
	if config.MaxUploadCount < 0 || config.MaxUploadCount > 9 {
		return gameApplicationConfigDTO{}, errors.New("maxUploadCount must be between 0 and 9")
	}
	if config.UploadRequired && config.MaxUploadCount == 0 {
		return gameApplicationConfigDTO{}, errors.New("maxUploadCount required when uploadRequired is true")
	}
	if config.MinIntroLength < 0 || config.MaxIntroLength < 0 || config.MinIntroLength > config.MaxIntroLength {
		return gameApplicationConfigDTO{}, errors.New("intro length invalid")
	}
	if config.MaxIntroLength > 500 {
		return gameApplicationConfigDTO{}, errors.New("maxIntroLength must be less than or equal to 500")
	}
	if config.MaxMessageLength < 0 || config.MaxMessageLength > 500 {
		return gameApplicationConfigDTO{}, errors.New("maxMessageLength must be between 0 and 500")
	}
	if len(config.AllowedUploadTypes) == 0 {
		config.AllowedUploadTypes = []string{"jpg", "png", "pdf"}
	}
	if config.Version == "" {
		config.Version = time.Now().UTC().Format("2006-01-02")
	}
	return config, nil
}

func normalizeGameAuditConfig(req gameAuditConfigDTO) (gameAuditConfigDTO, error) {
	config := cloneGameAuditConfig(req)
	config.RequireManualAuditTypes = normalizeStringList(config.RequireManualAuditTypes, 16)
	config.ApplicationAuditMode = strings.TrimSpace(config.ApplicationAuditMode)
	config.ReviewerRoles = normalizeStringList(config.ReviewerRoles, 16)
	config.Version = strings.TrimSpace(config.Version)
	for _, gameType := range config.RequireManualAuditTypes {
		if !validConfigGameType(gameType) {
			return gameAuditConfigDTO{}, errors.New("unsupported audit game type: " + gameType)
		}
	}
	if config.BatchAuditMaxCount <= 0 {
		config.BatchAuditMaxCount = 50
	}
	if config.BatchAuditMaxCount > 200 {
		return gameAuditConfigDTO{}, errors.New("batchAuditMaxCount must be less than or equal to 200")
	}
	switch config.ApplicationAuditMode {
	case "", "creator_or_main_guide":
		config.ApplicationAuditMode = "creator_or_main_guide"
	case "creator_only", "main_guide_only", "admin_only":
	default:
		return gameAuditConfigDTO{}, errors.New("unsupported applicationAuditMode")
	}
	if len(config.ReviewerRoles) == 0 {
		config.ReviewerRoles = []string{"super_admin", "audit_admin"}
	}
	if config.Version == "" {
		config.Version = time.Now().UTC().Format("2006-01-02")
	}
	return config, nil
}

func normalizeGameConditionRuleConfig(req gameConditionRuleConfigDTO) (gameConditionRuleConfigDTO, error) {
	config := cloneGameConditionRuleConfig(req)
	config.DefaultVisibility = strings.TrimSpace(config.DefaultVisibility)
	config.Version = strings.TrimSpace(config.Version)
	config.RuleItems = normalizeConditionRuleItems(config.RuleItems)
	if len(config.RuleItems) == 0 {
		return gameConditionRuleConfigDTO{}, errors.New("ruleItems required")
	}
	switch config.DefaultVisibility {
	case "", "approved_users":
		config.DefaultVisibility = "approved_users"
	case "all_users", "invite_only", "hidden":
	default:
		return gameConditionRuleConfigDTO{}, errors.New("unsupported defaultVisibility")
	}
	if config.Version == "" {
		config.Version = time.Now().UTC().Format("2006-01-02")
	}
	return config, nil
}

func normalizeConditionRuleItems(items []conditionRuleItemDTO) []conditionRuleItemDTO {
	result := make([]conditionRuleItemDTO, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item.Key = strings.TrimSpace(item.Key)
		item.Name = strings.TrimSpace(item.Name)
		item.Description = strings.TrimSpace(item.Description)
		if item.Key == "" || item.Name == "" || len(item.Key) > 64 || len(item.Name) > 64 {
			continue
		}
		if _, ok := seen[item.Key]; ok {
			continue
		}
		seen[item.Key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func normalizeStringList(items []string, max int) []string {
	result := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		value := strings.TrimSpace(strings.ToLower(item))
		if value == "" || len(value) > 64 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if max > 0 && len(result) >= max {
			break
		}
	}
	return result
}

func mergeStringMap(base map[string]string, patch map[string]string) map[string]string {
	result := cloneStringMap(base)
	for key, value := range patch {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}
	return result
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(src))
	for key, value := range src {
		result[key] = value
	}
	return result
}

func cloneGameApplicationConfig(config gameApplicationConfigDTO) gameApplicationConfigDTO {
	config.AllowedUploadTypes = append([]string(nil), config.AllowedUploadTypes...)
	config.Texts = cloneStringMap(config.Texts)
	config.AuditPage = cloneGameApplicationAuditPageConfig(config.AuditPage)
	return config
}

func mergeGameApplicationConfigDefaults(config gameApplicationConfigDTO) gameApplicationConfigDTO {
	defaults := defaultGameApplicationConfig()
	config = cloneGameApplicationConfig(config)
	if len(config.AllowedUploadTypes) == 0 {
		config.AllowedUploadTypes = append([]string(nil), defaults.AllowedUploadTypes...)
	}
	if config.MaxMessageLength == 0 {
		config.MaxMessageLength = defaults.MaxMessageLength
	}
	config.Texts = mergeStringMap(defaults.Texts, config.Texts)
	config.AuditPage = mergeGameApplicationAuditPageConfig(defaults.AuditPage, config.AuditPage)
	return config
}

func cloneGameApplicationAuditPageConfig(config gameApplicationAuditPageConfigDTO) gameApplicationAuditPageConfigDTO {
	config.Filters = append([]gameApplicationAuditFilterDTO(nil), config.Filters...)
	config.StatusTexts = cloneStringMap(config.StatusTexts)
	config.RoleNames = cloneStringMap(config.RoleNames)
	config.Texts = cloneStringMap(config.Texts)
	config.Detail = cloneGameApplicationAuditDetailConfig(config.Detail)
	return config
}

func mergeGameApplicationAuditPageConfig(defaults gameApplicationAuditPageConfigDTO, config gameApplicationAuditPageConfigDTO) gameApplicationAuditPageConfigDTO {
	result := cloneGameApplicationAuditPageConfig(defaults)
	if strings.TrimSpace(config.PageTitle) != "" {
		result.PageTitle = strings.TrimSpace(config.PageTitle)
	}
	if len(config.Filters) > 0 {
		result.Filters = append([]gameApplicationAuditFilterDTO(nil), config.Filters...)
	}
	result.StatusTexts = mergeStringMap(result.StatusTexts, config.StatusTexts)
	result.RoleNames = mergeStringMap(result.RoleNames, config.RoleNames)
	result.Texts = mergeStringMap(result.Texts, config.Texts)
	result.Detail = mergeGameApplicationAuditDetailConfig(result.Detail, config.Detail)
	return result
}

func cloneGameApplicationAuditDetailConfig(config gameApplicationAuditDetailConfigDTO) gameApplicationAuditDetailConfigDTO {
	config.StatusTitles = cloneStringMap(config.StatusTitles)
	config.CountdownTexts = cloneStringMap(config.CountdownTexts)
	config.PlayerStatusTexts = cloneStringMap(config.PlayerStatusTexts)
	config.Texts = cloneStringMap(config.Texts)
	config.SessionItems = append([]gameApplicationAuditSessionItemDTO(nil), config.SessionItems...)
	config.ConfirmRows = append([]gameApplicationAuditSessionItemDTO(nil), config.ConfirmRows...)
	config.OptionalActions = append([]gameApplicationAuditActionDTO(nil), config.OptionalActions...)
	config.NoticeBullets = append([]string(nil), config.NoticeBullets...)
	return config
}

func mergeGameApplicationAuditDetailConfig(defaults gameApplicationAuditDetailConfigDTO, config gameApplicationAuditDetailConfigDTO) gameApplicationAuditDetailConfigDTO {
	result := cloneGameApplicationAuditDetailConfig(defaults)
	if strings.TrimSpace(config.PageTitle) != "" {
		result.PageTitle = strings.TrimSpace(config.PageTitle)
	}
	if strings.TrimSpace(config.ReferralText) != "" {
		result.ReferralText = strings.TrimSpace(config.ReferralText)
	}
	result.StatusTitles = mergeStringMap(result.StatusTitles, config.StatusTitles)
	result.CountdownTexts = mergeStringMap(result.CountdownTexts, config.CountdownTexts)
	result.PlayerStatusTexts = mergeStringMap(result.PlayerStatusTexts, config.PlayerStatusTexts)
	result.Texts = mergeStringMap(result.Texts, config.Texts)
	if len(config.SessionItems) > 0 {
		result.SessionItems = append([]gameApplicationAuditSessionItemDTO(nil), config.SessionItems...)
	}
	if len(config.ConfirmRows) > 0 {
		result.ConfirmRows = append([]gameApplicationAuditSessionItemDTO(nil), config.ConfirmRows...)
	}
	if len(config.OptionalActions) > 0 {
		result.OptionalActions = append([]gameApplicationAuditActionDTO(nil), config.OptionalActions...)
	}
	if len(config.NoticeBullets) > 0 {
		result.NoticeBullets = append([]string(nil), config.NoticeBullets...)
	}
	return result
}

func cloneGameAuditConfig(config gameAuditConfigDTO) gameAuditConfigDTO {
	config.RequireManualAuditTypes = append([]string(nil), config.RequireManualAuditTypes...)
	config.ReviewerRoles = append([]string(nil), config.ReviewerRoles...)
	return config
}

func cloneGameConditionRuleConfig(config gameConditionRuleConfigDTO) gameConditionRuleConfigDTO {
	config.RuleItems = append([]conditionRuleItemDTO(nil), config.RuleItems...)
	return config
}

func (s *Server) createGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var req games.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	categoryConfig := s.currentGameCategoryConfig()
	if strings.TrimSpace(req.PrimaryCategory) != "" && !categoryKeyExists(categoryConfig.PrimaryCategories, strings.TrimSpace(req.PrimaryCategory)) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid primary category")
		return
	}
	if req.CoverFileID > 0 {
		file, err := s.files.Get(req.CoverFileID)
		if err != nil || file.UploaderID != userID || file.BizType != "game_cover" {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid game cover")
			return
		}
		download, err := s.files.DownloadURLForFile(file)
		if err != nil {
			if errors.Is(err, files.ErrStorageNotConfigured) {
				httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "storage base url not configured")
				return
			}
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid game cover")
			return
		}
		req.CoverImage = download.DownloadURL
	}
	game, err := s.games.Create(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, games.ErrRealnameRequired):
			httpx.Error(w, http.StatusForbidden, 40341, "强实名未完成")
		case errors.Is(err, games.ErrInvalidGameType):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "小程序一期只能创建免费局")
		case errors.Is(err, games.ErrInvalidPlayers):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "每局人数必须为 5-8")
		case errors.Is(err, games.ErrInvalidGameInput):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid game input")
		case errors.Is(err, games.ErrDailyLimit):
			httpx.Error(w, http.StatusTooManyRequests, 42921, "每日最多创建 3 局")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "创建局失败")
		}
		return
	}
	if _, err := s.orders.EnsureFreeNoPayOrder(userID, game.ID); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "创建订单占位失败")
		return
	}
	s.recordBehavior(userID, "game_create_submit", "game", game.ID, map[string]interface{}{"status": game.Status, "gameType": game.GameType})
	httpx.OK(w, game)
}

func (s *Server) listGames(w http.ResponseWriter, r *http.Request) {
	items := publicGames(s.games.List())
	if userID, ok := s.currentUserID(r); ok {
		s.recordBehavior(userID, "browse_games", "game", 0, map[string]interface{}{"count": len(items)})
	}
	listItems := make([]GameListItemDTO, 0, len(items))
	for _, game := range items {
		listItems = append(listItems, GameListItemDTO{
			Game:          game,
			PlayerAvatars: s.gameListPlayerAvatars(game),
		})
	}
	httpx.OK(w, map[string]interface{}{"items": listItems})
}

func (s *Server) gameListPlayerAvatars(game games.Game) []GameListAvatarDTO {
	userIDs := make([]int64, 0, 3)
	seen := map[int64]bool{}
	addUserID := func(userID int64) {
		if userID <= 0 || seen[userID] {
			return
		}
		seen[userID] = true
		userIDs = append(userIDs, userID)
	}

	addUserID(game.CreatorUserID)
	for _, userID := range s.games.Members(game.ID) {
		addUserID(userID)
	}

	avatars := make([]GameListAvatarDTO, 0, 3)
	for _, userID := range userIDs {
		avatarURL := strings.TrimSpace(s.imUserAvatarURL(userID))
		if strings.HasPrefix(strings.ToLower(avatarURL), "mock://") {
			avatarURL = ""
		}
		avatars = append(avatars, GameListAvatarDTO{
			UserID:    userID,
			Name:      s.displayName(userID, "玩家"),
			AvatarURL: avatarURL,
		})
		if len(avatars) == 3 {
			break
		}
	}
	return avatars
}

func (s *Server) appHome(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	requestedRoleType := strings.TrimSpace(r.URL.Query().Get("roleType"))
	roleType := normalizeHomeRoleType(requestedRoleType)
	if requestedRoleType != "" && roleType == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid roleType")
		return
	}
	if roleType != "" && roleType != "player" {
		status := s.profiles.RoleSnapshot(userID).RoleStatusMap[roleType]
		if status != "active" && status != "approved" {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "role is not active")
			return
		}
	}
	visibleGames := publicGames(s.games.List())
	s.recordBehavior(userID, "view_home", "home", 0, map[string]interface{}{"gameCount": len(visibleGames), "roleType": roleType})
	httpx.OK(w, s.buildAppHomePayload(userID, visibleGames, roleType))
}

func (s *Server) newbieTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	record := s.identity.Status(userID)
	snapshot := s.profiles.RoleSnapshot(userID)
	stats := s.games.StatsForUser(userID)
	_, err := s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	reviewIntents := s.reviews.MyIntents(userID)
	applications := s.profiles.RoleApplicationsByUser(userID)
	hasRoleApplication := len(applications) > 0
	hasApprovedRole := false
	for _, status := range snapshot.RoleStatusMap {
		if status == "active" || status == "approved" {
			hasApprovedRole = true
			break
		}
	}
	items := []map[string]interface{}{
		{"code": "complete_identity", "title": "完成实名认证", "completed": record.Status == "verified"},
		{"code": "apply_role", "title": "申请行家或领路人", "completed": hasRoleApplication || hasApprovedRole},
		{"code": "join_or_create_game", "title": "创建或参与第一局", "completed": stats.Participated > 0},
		{"code": "complete_game", "title": "完成一局服务", "completed": stats.Completed > 0},
		{"code": "submit_review", "title": "完成评价", "completed": len(reviewIntents) > 0},
	}
	completed := 0
	for _, item := range items {
		if done, _ := item["completed"].(bool); done {
			completed++
		}
	}
	dailyItems := []map[string]interface{}{
		{"code": "daily_join_game", "title": "今日参与 1 次组局", "completed": stats.Participated > 0, "current": stats.Participated, "required": 1},
	}
	activityItems := []map[string]interface{}{
		{"code": "activity_complete_game", "title": "完成一局并提交评价", "completed": stats.Completed > 0 && len(reviewIntents) > 0, "current": stats.Completed, "required": 1},
	}
	completedCodes := map[string]bool{}
	if s.tasks != nil {
		completedCodes = s.tasks.CompletedCodes(userID)
	}
	for _, collection := range [][]map[string]interface{}{items, dailyItems, activityItems} {
		for _, item := range collection {
			code, _ := item["code"].(string)
			if completedCodes[code] {
				item["completed"] = true
			}
			if done, _ := item["completed"].(bool); done && s.tasks != nil {
				_, _ = s.tasks.MarkCompleted(userID, code)
			}
		}
	}
	httpx.OK(w, map[string]interface{}{
		"items":     items,
		"completed": completed,
		"total":     len(items),
		"categories": []map[string]interface{}{
			{"key": "newbie", "title": "新手任务", "items": items},
			{"key": "daily", "title": "每日任务", "items": dailyItems},
			{"key": "activity", "title": "活动任务", "items": activityItems},
		},
	})
}

func (s *Server) myManagedGames(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	items := make([]map[string]interface{}, 0)
	for _, game := range s.games.List() {
		role := s.userRoleForGame(game, userID)
		if game.CreatorUserID == userID || s.userExpertForGame(game, userID) || isManagedGameRole(role) {
			items = append(items, s.buildManagedServiceOrder(userID, game))
		}
	}
	allItems := items
	items, page, pageSize, total := paginateRoleItems(r, allItems, "statusType")
	httpx.OK(w, map[string]interface{}{
		"items":       items,
		"orders":      items,
		"summary":     serviceOrderSummary(allItems, "\u672c\u6708\u670d\u52a1\u6536\u5165"),
		"pageConfig":  s.currentMyGamesPageConfig(),
		"serverTime":  time.Now().Format(time.RFC3339),
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
		"hasPrevious": page > 1,
		"hasMore":     page*pageSize < total,
	})
}

func isManagedGameRole(role string) bool {
	return role == "guide" || role == "main_guide"
}

func (s *Server) myPlayerGames(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	pageConfig := s.currentMyGamesPageConfig()
	invitedGameIDs := make(map[int64]bool)
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if invitation.TargetUserID == userID && (invitation.Status == "pending" || invitation.Status == "accepted") {
			invitedGameIDs[invitation.GameID] = true
		}
	}
	items := make([]map[string]interface{}, 0)
	for _, game := range s.games.List() {
		if game.CreatorUserID != userID && game.MainGuideUserID != userID && !s.userGuideForGame(game, userID) && !s.userExpertForGame(game, userID) && s.userPlayerForGame(game, userID) {
			item := s.buildPlayerServiceOrder(userID, game, pageConfig)
			if invitedGameIDs[game.ID] {
				item["category"] = "invited"
			}
			items = append(items, item)
			continue
		}
		if invitedGameIDs[game.ID] && game.CreatorUserID != userID && game.MainGuideUserID != userID && !s.userGuideForGame(game, userID) && !s.userExpertForGame(game, userID) {
			item := s.buildPlayerServiceOrder(userID, game, pageConfig)
			item["category"] = "invited"
			items = append(items, item)
		}
	}
	allItems := items
	items, page, pageSize, total := paginateRoleItems(r, allItems, "statusType")
	httpx.OK(w, map[string]interface{}{
		"items":       items,
		"orders":      items,
		"summary":     serviceOrderSummary(allItems, "\u672c\u6708\u670d\u52a1\u652f\u51fa"),
		"pageConfig":  pageConfig,
		"serverTime":  time.Now().Format(time.RFC3339),
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
		"hasPrevious": page > 1,
		"hasMore":     page*pageSize < total,
	})
}

func paginateRoleItems(r *http.Request, items []map[string]interface{}, statusField string) ([]map[string]interface{}, int, int, int) {
	page := queryInt(r.URL.Query().Get("page"))
	pageSize := queryInt(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 5
	}
	if pageSize > 20 {
		pageSize = 20
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "refund" {
		status = "canceled"
	}
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		value, _ := item[statusField].(string)
		matchesStatus := status == "" || value == status
		if status == "dispute" && (value == "dispute" || value == "canceled") {
			matchesStatus = true
		}
		if matchesStatus {
			filtered = append(filtered, item)
		}
	}
	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], page, pageSize, total
}

func (s *Server) buildManagedServiceOrder(userID int64, game games.Game) map[string]interface{} {
	targetID := s.firstMemberWithGameRole(game, userID, "member")
	if targetID == 0 && game.CreatorUserID != userID {
		targetID = game.CreatorUserID
	}
	hasReviewTodo := false
	if todo, ok := s.reviewTodoForGame(userID, game.ID); ok {
		targetID = todo.TargetUserID
		hasReviewTodo = true
	}
	hasSubmittedReview := s.userReviewedGame(userID, game.ID)
	statusType, statusText := serviceOrderStatus(game.Status)
	canReview := hasReviewTodo && statusType == "complete" && targetID > 0
	reviewed := hasSubmittedReview && !hasReviewTodo
	amountCent := successFundAmount(game)
	playerName := s.inGameDisplayName(targetID, "\u73a9\u5bb6")
	gameIDText := strconv.FormatInt(game.ID, 10)
	serviceOrderID := serviceOrderID(game.ID)
	reviewRoute := "/pages/game/review/index?gameId=" + gameIDText
	deliveryRoute := "/pages/game/delivery/index?gameId=" + gameIDText + "&mode=paid&note=%E7%A1%AE%E8%AE%A4%E6%9C%8D%E5%8A%A1%E5%AE%8C%E6%88%90"
	completionCanConfirm, _ := s.gameCompletionState(game, userID)["canConfirm"].(bool)
	completedActionText := serviceReviewActionText(reviewed, canReview, statusType, game.Status)
	completedActionRoute := reviewRoute
	completedActionEnabled := canReview
	if completionCanConfirm {
		completedActionText = "确认完成"
		completedActionRoute = deliveryRoute
		completedActionEnabled = true
	}
	expertCancelRoute := "/pages/game/expert-cancel/index?" + serviceCancelQuery(map[string]string{
		"serviceOrderId":  serviceOrderID,
		"gameId":          gameIDText,
		"playerId":        strconv.FormatInt(targetID, 10),
		"playerName":      playerName,
		"avatarText":      avatarTextForName(playerName, targetID),
		"serviceTitle":    game.Title,
		"amount":          "0",
		"amountText":      "免费",
		"gameType":        "free",
		"freeCancel":      "1",
		"platformFeeRate": "0",
		"platformFeeText": "免费",
		"statusText":      statusText,
		"warningTitle":    "免费局取消无需赔付",
		"warningDesc":     "当前没有收费局，本次取消不会产生赔付金额，但会扣减信用分。",
	})
	playerContactRoute := "/pages/im/room/index?gameId=" + gameIDText + "&prefill=%E4%BD%A0%E5%A5%BD%EF%BC%8C%E6%83%B3%E5%92%8C%E4%BD%A0%E7%A1%AE%E8%AE%A4%E4%B8%80%E4%B8%8B%E6%9C%8D%E5%8A%A1%E8%BF%9B%E5%BA%A6%E3%80%82"
	guideContactRoute := "/pages/im/room/index?gameId=" + gameIDText + "&prefill=%E4%BD%A0%E5%A5%BD%EF%BC%8C%E8%BE%9B%E8%8B%A6%E5%90%8C%E6%AD%A5%E4%B8%80%E4%B8%8B%E7%BB%84%E5%B1%80%E8%BF%9B%E5%BA%A6%E3%80%82"
	return map[string]interface{}{
		"id":                  serviceOrderID,
		"game":                game,
		"gameId":              game.ID,
		"serviceOrderId":      serviceOrderID,
		"ref":                 serviceOrderID,
		"title":               game.Title,
		"serviceTitle":        game.Title,
		"statusType":          statusType,
		"statusText":          statusText,
		"amount":              amountCent / 100,
		"amountCent":          amountCent,
		"amountText":          serviceAmountText(amountCent),
		"playerId":            targetID,
		"playerName":          playerName,
		"name":                playerName,
		"player":              serviceOrderPerson(targetID, playerName),
		"guideId":             game.MainGuideUserID,
		"guide":               serviceOrderPerson(game.MainGuideUserID, s.inGameDisplayName(game.MainGuideUserID, "\u9886\u8def\u4eba")),
		"startedAt":           game.CreatedAt.Format(time.RFC3339),
		"expectedDeliveryAt":  game.CreatedAt.Add(72 * time.Hour).Format(time.RFC3339),
		"serverTime":          time.Now().Format(time.RFC3339),
		"reviewStatus":        serviceReviewStatus(reviewed, canReview, statusType, game.Status),
		"reviewed":            reviewed,
		"canReview":           canReview,
		"canReviewBoth":       canReview,
		"reviewActionText":    completedActionText,
		"canReviewAction":     completedActionEnabled,
		"primaryActionText":   "提前结束交付",
		"secondaryActionText": "取消并赔付",
		"playerActionText":    "联系玩家",
		"guideActionText":     "联系领路人",
		"deliveryRoute":       deliveryRoute,
		"expertCancelRoute":   expertCancelRoute,
		"contactPlayerRoute":  playerContactRoute,
		"contactGuideRoute":   guideContactRoute,
		"reviewRoute":         completedActionRoute,
		"actions": map[string]interface{}{
			"canFinishDelivery":         statusType == "active",
			"finishDeliveryRoute":       deliveryRoute,
			"canCancelWithCompensation": statusType == "active",
			"expertCancelRoute":         expertCancelRoute,
			"canContactPlayer":          targetID > 0,
			"contactPlayerRoute":        playerContactRoute,
			"canContactGuide":           game.MainGuideUserID > 0,
			"contactGuideRoute":         guideContactRoute,
			"canReview":                 canReview,
			"canReviewAction":           completedActionEnabled,
			"reviewRoute":               completedActionRoute,
		},
	}
}

func (s *Server) buildPlayerServiceOrder(userID int64, game games.Game, pageConfig map[string]interface{}) map[string]interface{} {
	expertID := s.firstMemberWithGameRole(game, userID, "expert")
	targetID := expertID
	hasReviewTodo := false
	if todo, ok := s.reviewTodoForGame(userID, game.ID); ok {
		targetID = todo.TargetUserID
		hasReviewTodo = true
	}
	hasSubmittedReview := s.userReviewedGame(userID, game.ID)
	statusType, statusText := serviceOrderStatus(game.Status)
	canReview := hasReviewTodo && statusType == "complete" && targetID > 0
	reviewed := hasSubmittedReview && !hasReviewTodo
	amountCent := successFundAmount(game)
	hasExpert := expertID > 0
	expertName := "暂未分配行家"
	expertAvatarText := "局"
	if hasExpert {
		expertName = s.inGameDisplayName(expertID, "\u884c\u5bb6")
		expertAvatarText = avatarTextForName(expertName, expertID)
	}
	gameIDText := strconv.FormatInt(game.ID, 10)
	serviceOrderID := serviceOrderID(game.ID)
	reviewRoute := "/pages/game/review/index?gameId=" + gameIDText
	deliveryRoute := "/pages/game/delivery/index?gameId=" + gameIDText + "&mode=free"
	completionCanConfirm, _ := s.gameCompletionState(game, userID)["canConfirm"].(bool)
	completedActionText := serviceReviewActionText(reviewed, canReview, statusType, game.Status)
	completedActionRoute := reviewRoute
	completedActionEnabled := canReview
	if completionCanConfirm {
		completedActionText = "确认完成"
		completedActionRoute = deliveryRoute
		completedActionEnabled = true
	}
	playerName := s.inGameDisplayName(userID, "玩家")
	playerCancelRoute := "/pages/game/player-cancel/index?" + serviceCancelQuery(map[string]string{
		"serviceOrderId":     serviceOrderID,
		"gameId":             gameIDText,
		"ref":                serviceOrderID,
		"playerName":         playerName,
		"playerAvatarText":   avatarTextForName(playerName, userID),
		"expertId":           strconv.FormatInt(expertID, 10),
		"expertName":         expertName,
		"expertAvatarText":   expertAvatarText,
		"hasExpert":          boolQueryValue(hasExpert),
		"serviceTitle":       game.Title,
		"amount":             "0",
		"contractAmount":     "0",
		"amountText":         "免费",
		"contractAmountText": "免费",
		"gameType":           "free",
		"freeCancel":         "1",
		"minRate":            "0",
		"maxRate":            "0",
		"suggestedRate":      "0",
		"suggestionMinRate":  "0",
		"suggestionMaxRate":  "0",
		"platformFeeRate":    "0",
		"servedDurationText": "待确认",
		"totalDurationText":  "免费局",
		"warningTitle":       "免费局取消无需赔付",
		"warningDesc":        "当前没有收费局，本次取消不会产生赔付金额，但会扣减信用分。",
	})
	contactExpertRoute := ""
	if hasExpert {
		contactExpertRoute = "/pages/message/my/index?mode=private&targetUserId=" + strconv.FormatInt(expertID, 10) + "&sourceGameId=" + gameIDText + "&prefill=" + url.QueryEscape("你好，我这边想确认一下服务内容。")
	}
	primaryActionText := myGamesPageConfigText(pageConfig, "contactExpertText", "联系行家")
	noticeText := myGamesPageConfigText(pageConfig, "cancelNoticeText", "取消需赔付一定比例金额给行家")
	if !hasExpert {
		primaryActionText = "暂无行家"
		noticeText = "本局暂未分配行家，无法联系行家。"
	}
	return map[string]interface{}{
		"id":                  serviceOrderID,
		"game":                game,
		"gameId":              game.ID,
		"serviceOrderId":      serviceOrderID,
		"ref":                 serviceOrderID,
		"title":               game.Title,
		"serviceTitle":        game.Title,
		"statusType":          statusType,
		"statusText":          statusText,
		"amount":              amountCent / 100,
		"amountCent":          amountCent,
		"amountText":          serviceAmountText(amountCent),
		"expertId":            expertID,
		"expertName":          expertName,
		"name":                expertName,
		"hasExpert":           hasExpert,
		"expert":              serviceOrderPerson(expertID, expertName),
		"guideId":             game.MainGuideUserID,
		"guide":               serviceOrderPerson(game.MainGuideUserID, s.inGameDisplayName(game.MainGuideUserID, "\u9886\u8def\u4eba")),
		"startedAt":           game.CreatedAt.Format(time.RFC3339),
		"expectedDeliveryAt":  game.CreatedAt.Add(72 * time.Hour).Format(time.RFC3339),
		"serverTime":          time.Now().Format(time.RFC3339),
		"reviewStatus":        serviceReviewStatus(reviewed, canReview, statusType, game.Status),
		"reviewed":            reviewed,
		"canReview":           canReview,
		"canReviewBoth":       canReview,
		"reviewActionText":    completedActionText,
		"canReviewAction":     completedActionEnabled,
		"primaryActionText":   primaryActionText,
		"secondaryActionText": myGamesPageConfigText(pageConfig, "cancelOrderText", "申请取消"),
		"noticeText":          noticeText,
		"contactExpertRoute":  contactExpertRoute,
		"playerCancelRoute":   playerCancelRoute,
		"reviewRoute":         completedActionRoute,
		"actions": map[string]interface{}{
			"canContactExpert":   hasExpert,
			"contactExpertRoute": contactExpertRoute,
			"canCancelOrder":     statusType == "active",
			"playerCancelRoute":  playerCancelRoute,
			"canReview":          canReview,
			"canReviewAction":    completedActionEnabled,
			"reviewRoute":        completedActionRoute,
		},
	}
}

func (s *Server) reviewTodoForGame(userID int64, gameID int64) (reviews.Todo, bool) {
	todos, err := s.reviews.Todos(userID)
	if err != nil {
		return reviews.Todo{}, false
	}
	for _, todo := range todos {
		if todo.GameID == gameID {
			return todo, true
		}
	}
	return reviews.Todo{}, false
}

func (s *Server) userReviewedGame(userID int64, gameID int64) bool {
	for _, review := range s.reviews.AllReviews() {
		if review.GameID == gameID && review.ReviewerUserID == userID {
			return true
		}
	}
	return false
}

func firstOtherMember(userID int64, memberIDs []int64) int64 {
	for _, memberID := range memberIDs {
		if memberID != userID {
			return memberID
		}
	}
	return 0
}

func (s *Server) firstMemberWithGameRole(game games.Game, excludeUserID int64, expectedRole string) int64 {
	memberRoleItems := s.games.MemberRoles(game.ID)
	roles := gameMemberRoleMap(memberRoleItems)
	for _, item := range memberRoleItems {
		if item.UserID == excludeUserID {
			continue
		}
		if gameMemberRole(game, item.UserID, true, roles) == expectedRole {
			return item.UserID
		}
	}
	return 0
}

func serviceOrderID(gameID int64) string {
	return "GAME-" + strconv.FormatInt(gameID, 10)
}

func serviceCancelQuery(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		values.Set(key, value)
	}
	return values.Encode()
}

func boolQueryValue(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func serviceOrderPerson(userID int64, name string) map[string]interface{} {
	if userID <= 0 {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"id":         userID,
		"userId":     userID,
		"name":       name,
		"avatarText": avatarTextForName(name, userID),
	}
}

func serviceOrderStatus(status string) (string, string) {
	switch status {
	case "pending_confirm":
		return "complete", "\u5df2\u5b8c\u6210"
	case "pending_review":
		return "complete", "\u5f85\u8bc4\u4ef7"
	case "completed":
		return "complete", "\u5df2\u5b8c\u6210"
	case "canceled", "cancelled":
		return "canceled", "\u5df2\u53d6\u6d88"
	case "disputed":
		return "dispute", "\u4e89\u8bae\u4e2d"
	case "in_progress":
		return "active", "\u8fdb\u884c\u4e2d"
	default:
		return "active", homeGameStatusText(status)
	}
}

func serviceReviewStatus(reviewed bool, canReview bool, statusType string, gameStatus string) string {
	if reviewed {
		return "reviewed"
	}
	if canReview {
		return "pending"
	}
	if gameStatus == "pending_confirm" {
		return "waiting_confirmation"
	}
	if statusType == "complete" {
		return "expired"
	}
	return "none"
}

func serviceReviewActionText(reviewed bool, canReview bool, statusType string, gameStatus string) string {
	if reviewed {
		return "\u5df2\u8bc4\u4ef7"
	}
	if canReview {
		return "\u8bc4\u4ef7\u53cc\u65b9"
	}
	if gameStatus == "pending_confirm" {
		return "\u7b49\u5f85\u5b8c\u6210\u786e\u8ba4"
	}
	if statusType == "complete" {
		return "\u8bc4\u4ef7\u5df2\u7ed3\u675f"
	}
	return "\u6682\u4e0d\u53ef\u8bc4\u4ef7"
}

func serviceAmountText(amountCent int64) string {
	if amountCent <= 0 {
		return "\u00a50"
	}
	return "\u00a5" + strconv.FormatInt(amountCent/100, 10)
}

func serviceOrderSummary(items []map[string]interface{}, title string) map[string]interface{} {
	active := 0
	complete := 0
	dispute := 0
	canceled := 0
	var amountCent int64
	for _, item := range items {
		switch item["statusType"] {
		case "active":
			active++
		case "complete":
			complete++
		case "dispute":
			dispute++
		case "canceled":
			canceled++
		}
		if value, ok := item["amountCent"].(int64); ok {
			amountCent += value
		}
	}
	return map[string]interface{}{
		"title":         title,
		"amountCent":    amountCent,
		"amountText":    serviceAmountText(amountCent),
		"activeCount":   active,
		"pendingCount":  complete,
		"completeCount": complete,
		"disputeCount":  dispute + canceled,
		"canceledCount": canceled,
		"refundCount":   canceled,
	}
}

func publicGames(items []games.Game) []games.Game {
	result := make([]games.Game, 0, len(items))
	now := time.Now()
	for _, game := range items {
		if isPublicGameStatus(game.Status) && games.CanApplyWithinSignupWindow(game, now) {
			result = append(result, game)
		}
	}
	return result
}

func isPublicGameStatus(status string) bool {
	return status == "recruiting"
}

func (s *Server) adminGames(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	gameType := strings.TrimSpace(r.URL.Query().Get("gameType"))
	cityCode := strings.TrimSpace(r.URL.Query().Get("cityCode"))
	keyword := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("keyword")))
	allItems := s.games.List()
	items := make([]games.Game, 0, len(allItems))
	for _, game := range allItems {
		if status != "" && game.Status != status {
			continue
		}
		if gameType != "" && game.GameType != gameType {
			continue
		}
		if cityCode != "" && game.CityCode != cityCode {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(game.Title), keyword) && !strings.Contains(strings.ToLower(game.CityName), keyword) {
			continue
		}
		items = append(items, game)
	}
	httpx.OK(w, map[string]interface{}{
		"items":    items,
		"total":    len(items),
		"status":   status,
		"gameType": gameType,
		"cityCode": cityCode,
		"keyword":  keyword,
	})
}

func (s *Server) adminGameApplications(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	gameID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("gameId")), 10, 64)
	userID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("userId")), 10, 64)
	applications := make([]games.Application, 0)
	seen := make(map[int64]bool)
	for _, game := range s.games.List() {
		if gameID > 0 && game.ID != gameID {
			continue
		}
		for _, item := range s.games.ApplicationsForCreator(game.CreatorUserID) {
			if seen[item.ID] || (gameID > 0 && item.GameID != gameID) || (userID > 0 && item.UserID != userID) || (status != "" && item.Status != status) {
				continue
			}
			seen[item.ID] = true
			applications = append(applications, item)
		}
	}
	items := s.buildReceivedApplicationItems(0, applications)
	httpx.OK(w, map[string]interface{}{
		"items":  items,
		"total":  len(items),
		"status": status,
		"gameId": gameID,
		"userId": userID,
	})
}

func (s *Server) adminGameDetail(w http.ResponseWriter, r *http.Request) {
	gameID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/games/", "")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	milestones, err := s.games.AdminMilestones(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	checkins, err := s.games.AdminCheckins(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	retrospectives, err := s.games.AdminRetrospectives(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	continueDrafts, err := s.games.AdminContinueDrafts(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	confirm, confirmItems, hasConfirm := s.games.ServiceConfirmForGame(gameID)
	httpx.OK(w, map[string]interface{}{
		"game":              game,
		"memberIds":         s.games.Members(gameID),
		"milestones":        milestones,
		"checkins":          checkins,
		"retrospectives":    retrospectives,
		"continueDrafts":    continueDrafts,
		"serviceConfirm":    confirm,
		"confirmItems":      confirmItems,
		"hasServiceConfirm": hasConfirm,
	})
}

func (s *Server) adminCreateGame(w http.ResponseWriter, r *http.Request) {
	var req games.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "璇锋眰鍙傛暟閿欒")
		return
	}
	game, err := s.games.CreateFromAdmin(req)
	if err != nil {
		switch {
		case errors.Is(err, games.ErrInvalidGameType):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid game type")
		case errors.Is(err, games.ErrInvalidPlayers):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "姣忓眬浜烘暟蹇呴』涓?5-8")
		case errors.Is(err, games.ErrInvalidGameInput):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid game input")
		case errors.Is(err, games.ErrRealnameRequired):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "main guide realname required")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "鍒涘缓灞€澶辫触")
		}
		return
	}
	s.recordOperation(r, "game:create_admin", "game", strconv.FormatInt(game.ID, 10), map[string]interface{}{
		"creatorUserId": game.CreatorUserID,
		"gameType":      game.GameType,
		"status":        game.Status,
		"source":        game.GameSource,
	})
	httpx.OK(w, game)
}

func (s *Server) adminEnsureGameIMRoom(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/im-room")
	if !ok {
		return
	}
	if _, err := s.games.Get(gameID); err != nil {
		writeGameError(w, err)
		return
	}

	existed := false
	for _, room := range s.im.AdminRooms() {
		if room.GameID == gameID {
			existed = true
			break
		}
	}
	room := s.im.EnsureRoom(gameID)
	s.recordOperation(r, "im:room:ensure", "game", strconv.FormatInt(gameID, 10), map[string]interface{}{
		"roomId":  room.ID,
		"engine":  room.Engine,
		"created": !existed,
	})
	httpx.OK(w, map[string]interface{}{
		"room":    room,
		"created": !existed,
	})
}

func (s *Server) getGame(w http.ResponseWriter, r *http.Request) {
	idText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/app/games/"), "/")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "局 ID 错误")
		return
	}
	game, err := s.games.Get(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, 40421, "局不存在")
		return
	}
	if userID, ok := s.currentUserID(r); ok {
		s.recordBehavior(userID, "view_game_detail", "game", id, nil)
		httpx.OK(w, s.buildGameDetail(userID, game))
		return
	}
	httpx.OK(w, s.buildGameDetail(0, game))
}

func (s *Server) gameMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/members")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.games.IsMember(gameID, userID) && game.CreatorUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权查看成员")
		return
	}
	confirmed, members := s.buildGameMembers(userID, game, true)
	httpx.OK(w, map[string]interface{}{
		"game":    game,
		"items":   members,
		"total":   len(members),
		"confirm": confirmed,
	})
}

func (s *Server) buildGameMembers(userID int64, game games.Game, includeRealName bool) (map[int64]bool, []GameMemberDTO) {
	confirmed := make(map[int64]bool)
	if _, items, ok := s.games.ServiceConfirmForGame(game.ID); ok {
		for _, item := range items {
			confirmed[item.UserID] = true
		}
	}
	memberIDs := s.games.Members(game.ID)
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	members := make([]GameMemberDTO, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		isCreator := memberID == game.CreatorUserID
		isCurrentUser := memberID == userID
		role := gameMemberRole(game, memberID, true, memberRoles)
		roleLabel := gameRoleLabel(role)
		profile := s.inGameIdentity(memberID, roleLabel)
		nickname := ""
		if user, ok := s.auth.UserByID(memberID); ok {
			nickname = strings.TrimSpace(user.Nickname)
		}
		name := nickname
		if name == "" {
			name = profile.DisplayName
		}
		primaryTag := roleLabel
		if isCurrentUser {
			primaryTag = "当前用户"
		}
		member := GameMemberDTO{
			UserID:        memberID,
			Name:          name,
			Nickname:      nickname,
			DisplayName:   name,
			AvatarURL:     s.imUserAvatarURL(memberID),
			AvatarText:    profile.AvatarText,
			Role:          role,
			RoleLabel:     roleLabel,
			Position:      roleLabel,
			Topic:         "参与本次组局",
			PrimaryTag:    primaryTag,
			Location:      game.CityName,
			IsCreator:     isCreator,
			IsCurrentUser: isCurrentUser,
			Confirmed:     confirmed[memberID],
		}
		if includeRealName {
			member.RealName = profile.RealName
		}
		members = append(members, member)
	}
	return confirmed, members
}

func (s *Server) buildGameDetail(userID int64, game games.Game) GameDetailDTO {
	memberIDs := s.games.Members(game.ID)
	_, members := s.buildGameMembers(userID, game, false)
	relation := s.buildGameRelation(userID, game)
	isFavorited, favoriteCount := s.gameFavoriteState(userID, game.ID)
	detail := GameDetailDTO{
		Game:          game,
		MemberIDs:     memberIDs,
		Members:       members,
		MyRelation:    relation,
		IsFavorited:   isFavorited,
		FavoriteCount: favoriteCount,
		Review: GameReviewDTO{
			Complete: s.reviews.GameReviewComplete(game.ID),
		},
	}
	detail.AuditRejectReason = game.RejectReason
	if relation.IsMember {
		if feedbacks, err := s.games.ProgressFeedbacks(userID, game.ID); err == nil {
			detail.Progress.Feedbacks = feedbacks
		}
		if milestones, err := s.games.Milestones(userID, game.ID); err == nil {
			detail.Progress.Milestones = milestones
		}
		if checkins, err := s.games.Checkins(userID, game.ID); err == nil {
			detail.Progress.Checkins = checkins
		}
		if room, err := s.im.RoomForGame(userID, game.ID); err == nil {
			detail.IM = GameIMDTO{
				Available:     true,
				RoomID:        room.ID,
				Status:        room.Status,
				Engine:        room.Engine,
				OpenIMGroupID: room.OpenIMGroupID,
			}
		} else if game.Status == "in_progress" || game.Status == "pending_confirm" || game.Status == "pending_review" || game.Status == "completed" {
			detail.IM.Available = true
		}
		if todos, err := s.reviews.Todos(userID); err == nil {
			for _, todo := range todos {
				if todo.GameID == game.ID {
					detail.Review.Todos = append(detail.Review.Todos, todo)
				}
			}
			detail.Review.Reviewable = len(detail.Review.Todos) > 0
			detail.MyRelation.CanReview = detail.Review.Reviewable
		}
	}
	if confirm, items, ok := s.games.ServiceConfirmForGame(game.ID); ok {
		detail.ServiceConfirm = &ServiceConfirmDTO{Confirm: confirm, Items: items}
	}
	if relation.ApplicationID > 0 {
		app := games.Application{
			ID:     relation.ApplicationID,
			GameID: game.ID,
			UserID: userID,
			Status: relation.ApplicationStatus,
		}
		detail.PendingApplication = &app
	}
	detail.DetailDisplay = s.buildGameDetailDisplay(userID, game, detail.MyRelation)
	return detail
}

func (s *Server) gameFavoriteState(userID int64, gameID int64) (bool, int) {
	isFavorited := false
	count := 0
	for _, favorite := range s.games.AllFavorites() {
		if favorite.GameID != gameID {
			continue
		}
		count++
		if userID > 0 && favorite.UserID == userID {
			isFavorited = true
		}
	}
	return isFavorited, count
}

func (s *Server) buildGameDetailDisplay(userID int64, game games.Game, relation GameMyRelationDTO) GameDetailDisplayDTO {
	pendingCount := 0
	for _, application := range s.games.ApplicationsForCreator(game.CreatorUserID) {
		if application.GameID == game.ID && application.Status == "pending" {
			pendingCount++
		}
	}
	rating, ratingCount := organizerRating(s.reviews.AllReviews(), game.CreatorUserID)
	reviewed := userID > 0 && s.userReviewedGame(userID, game.ID)
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	organizerRole := gameMemberRole(game, game.CreatorUserID, true, memberRoles)
	organizerRoleLabel := gameRoleLabel(organizerRole)
	organizerIdentity := s.inGameIdentity(game.CreatorUserID, organizerRoleLabel)

	return GameDetailDisplayDTO{
		StatusText:              gameDetailStatusText(game.Status),
		PendingApplicationCount: pendingCount,
		Organizer: GameDetailOrganizerDTO{
			UserID:      game.CreatorUserID,
			Name:        s.displayName(game.CreatorUserID, "玩家"),
			AvatarText:  organizerIdentity.AvatarText,
			AvatarURL:   s.imUserAvatarURL(game.CreatorUserID),
			Role:        organizerRole,
			RoleLabel:   organizerRoleLabel,
			Rating:      rating,
			RatingCount: ratingCount,
		},
		PrimaryAction: gameDetailPrimaryAction(game, relation, pendingCount, reviewed),
	}
}

func gameDetailStatusText(status string) string {
	switch status {
	case "pending_confirm", "pending_review", "completed":
		return "已结束"
	default:
		return homeGameStatusText(status)
	}
}

func organizerRating(items []reviews.Review, organizerUserID int64) (string, int) {
	total := 0
	count := 0
	for _, item := range items {
		if item.TargetUserID != organizerUserID || item.Score < 1 || item.Score > 5 {
			continue
		}
		total += item.Score
		count++
	}
	if count == 0 {
		return "", 0
	}
	return strconv.FormatFloat(float64(total)/float64(count), 'f', 1, 64), count
}

func gameDetailPrimaryAction(game games.Game, relation GameMyRelationDTO, pendingCount int, reviewed bool) GameDetailPrimaryActionDTO {
	gameID := strconv.FormatInt(game.ID, 10)
	disabled := func(text string) GameDetailPrimaryActionDTO {
		return GameDetailPrimaryActionDTO{Text: text, Disabled: true, Action: "none"}
	}
	action := func(text string, action string, route string) GameDetailPrimaryActionDTO {
		return GameDetailPrimaryActionDTO{Text: text, Action: action, Route: route}
	}

	switch game.Status {
	case "pending_audit":
		return disabled("后台审核中")
	case "recruiting", "full":
		if relation.CanStart {
			return GameDetailPrimaryActionDTO{
				Text:        "开始组局",
				Action:      "start",
				ConfirmText: "确认开始本局？开始后将进入组局协作。",
			}
		}
		if game.Status == "recruiting" && relation.IsCreator && relation.CanAudit && pendingCount > 0 {
			return action("审核报名（"+strconv.Itoa(pendingCount)+"）", "audit", "/pages/game/audit/index?gameId="+gameID)
		}
		if game.Status == "recruiting" && relation.ApplicationStatus == "pending" {
			return disabled("报名审核中")
		}
		// 人数以已入局成员数（CurrentPlayers）为准。即使状态字段尚未
		// 同步为 full，也不能继续展示“立即报名”或报名时间提示。
		if game.MaxPlayers > 0 && game.CurrentPlayers >= game.MaxPlayers && !relation.IsMember {
			return disabled("该局已满员")
		}
		if game.Status == "recruiting" && relation.CanApply {
			return action("立即报名", "apply", "/pages/game/apply/index?gameId="+gameID)
		}
		if game.Status == "recruiting" && relation.ApplyDisabledReason != "" && !relation.IsCreator && !relation.IsMember {
			return disabled(relation.ApplyDisabledReason)
		}
		if game.Status == "full" {
			if relation.IsMember {
				return disabled("等待开局")
			}
			return disabled("已满员")
		}
		if relation.IsCreator {
			return disabled("等待报名")
		}
		return disabled("等待开局")
	case "in_progress":
		if relation.IsMember {
			return action("进入组局", "collaboration", "/pages/game/collaboration/index?gameId="+gameID)
		}
		return disabled("进行中")
	case "pending_confirm":
		return disabled("已结束")
	case "pending_review":
		if relation.CanReview {
			return action("去评价", "review", "/pages/game/review/index?gameId="+gameID)
		}
		if reviewed {
			return disabled("已评价")
		}
		return disabled("待评价")
	case "completed":
		return disabled("已完成")
	case "draft":
		return disabled("草稿")
	default:
		return disabled(homeGameStatusText(game.Status))
	}
}

func (s *Server) gameSuccessDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/success-detail")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	successInvitation, hasSuccessInvitation := s.successfulExpertInvitation(game.ID, userID)
	if !s.games.IsMember(game.ID, userID) && !hasSuccessInvitation {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	if updatedGame, resolved, err := s.games.ResolveNoExpertPendingConfirm(game.ID); err != nil {
		writeGameError(w, err)
		return
	} else if resolved {
		game = updatedGame
		if !currentGameReviewable(s.games, game.ID) {
			s.reviews.MarkGameReviewable(game.ID)
			s.reviews.AwardCompletedGame(game.ID)
			s.createCoGameConnections(game.ID)
		}
	}
	if r.URL.Query().Get("view") == "service-confirm" && !s.canAccessServiceConfirmation(game, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅组局绑定的行家和玩家可进入确认页")
		return
	}
	relation := s.buildGameRelation(userID, game)
	if hasSuccessInvitation {
		relation.Role = "expert"
	}
	memberIDs := s.games.Members(game.ID)
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	participants := make([]map[string]interface{}, 0, len(memberIDs))
	seenParticipants := make(map[int64]bool)
	for _, memberID := range memberIDs {
		seenParticipants[memberID] = true
		profile := s.inGameIdentity(memberID, "\u6210\u5458")
		name := profile.DisplayName
		participants = append(participants, map[string]interface{}{
			"id":          memberID,
			"userId":      memberID,
			"name":        name,
			"realName":    profile.RealName,
			"displayName": profile.DisplayName,
			"avatar":      profile.AvatarText,
			"avatarUrl":   s.imUserAvatarURL(memberID),
			"avatarText":  profile.AvatarText,
			"role":        gameMemberRole(game, memberID, true, memberRoles),
			"roleLabel":   gameRoleLabel(gameMemberRole(game, memberID, true, memberRoles)),
			"colorClass":  gameRoleAvatarClass(gameMemberRole(game, memberID, true, memberRoles)),
			"isCurrent":   memberID == userID,
		})
	}
	if hasSuccessInvitation {
		playerInvitation, expertInvitation, _ := s.invitationConfirmedParties(successInvitation)
		for _, party := range []struct {
			userID int64
			role   string
			label  string
			color  string
		}{
			{userID: playerInvitation.TargetUserID, role: "player", label: "玩家", color: "pink"},
			{userID: expertInvitation.TargetUserID, role: "expert", label: "行家", color: "blue"},
			{userID: successInvitation.InviterID, role: "guide", label: "领路人", color: "orange"},
		} {
			if party.userID <= 0 || seenParticipants[party.userID] {
				continue
			}
			seenParticipants[party.userID] = true
			profile := s.inGameIdentity(party.userID, party.label)
			name := profile.DisplayName
			participants = append(participants, map[string]interface{}{
				"id": party.userID, "userId": party.userID, "name": name, "realName": profile.RealName, "displayName": profile.DisplayName,
				"avatar": profile.AvatarText, "avatarUrl": s.imUserAvatarURL(party.userID), "avatarText": profile.AvatarText,
				"role": party.role, "roleLabel": party.label, "colorClass": party.color, "isCurrent": party.userID == userID,
			})
		}
	}
	roomID := int64(0)
	if room, err := s.im.RoomForGame(userID, game.ID); err == nil {
		roomID = room.ID
	}
	fundAmount := successFundAmount(game)
	fund := successFundStatus(game, fundAmount)
	deliveryMode := "paid"
	if fundAmount <= 0 {
		deliveryMode = "free"
	}
	deliveryPage := s.currentGameDeliveryPageConfig()
	contactKey, contactText, contactPrefill := deliveryContactForRole(relation.Role)
	confirmTimeline := s.serviceConfirmTimelineSteps(game, memberRoles, userID)
	httpx.OK(w, map[string]interface{}{
		"gameId":    game.ID,
		"pageTexts": successExpertPageTexts(),
		"viewer": map[string]interface{}{
			"userId": userID,
			"role":   relation.Role,
		},
		"group": map[string]interface{}{
			"title":  game.Title + " - \u4e09\u65b9\u7fa4",
			"roles":  "\u884c\u5bb6\u3001\u9886\u8def\u4eba\u3001\u73a9\u5bb6",
			"hint":   "\u9886\u8def\u4eba\u5c06\u6301\u7eed\u8ddf\u8fdb\u6d3b\u52a8\u8fdb\u5ea6\uff0c\u786e\u4fdd\u53cc\u65b9\u987a\u5229\u5bf9\u63a5",
			"roomId": roomID,
			"gameId": game.ID,
		},
		"participants":     participants,
		"participantCount": len(participants),
		"deliveryMode":     deliveryMode,
		"completion":       s.gameCompletionState(game, userID),
		"activityRows":     deliveryActivityRows(game),
		"fund":             fund,
		"nextSteps":        confirmTimeline,
		"deliveryPage":     deliveryPage,
		"deliveryProof": map[string]interface{}{
			"maxCount":                4,
			"emptyText":               "\u53ef\u4e0a\u4f20\u670d\u52a1\u5b8c\u6210\u622a\u56fe\u3001\u4ea4\u4ed8\u6750\u6599\u622a\u56fe\u7b49\u51ed\u8bc1\uff0c\u6700\u591a 4 \u5f20\u3002",
			"fullText":                "\u6700\u591a\u4e0a\u4f20 4 \u5f20\u51ed\u8bc1",
			"selectedTemplate":        "\u5df2\u9009\u62e9 {selected}/{max} \u5f20\u51ed\u8bc1",
			"uploadActionText":        "\u4e0a\u4f20\u51ed\u8bc1",
			"contactPlayerText":       "\u8054\u7cfb\u73a9\u5bb6",
			"contactExpertText":       "联系行家",
			"contactGuideText":        "\u8054\u7cfb\u9886\u8def\u4eba",
			"cancelServiceText":       "\u7533\u8bf7\u53d6\u6d88\u670d\u52a1",
			"unavailableTextTemplate": "{action}\u6682\u4e0d\u53ef\u7528",
			"contactPlayerPrefill":    "\u4f60\u597d\uff0c\u9ebb\u70e6\u786e\u8ba4\u4e00\u4e0b\u670d\u52a1\u5b8c\u6210\u60c5\u51b5\u3002",
			"contactExpertPrefill":    "你好，想和你确认一下本次服务完成情况。",
			"contactGuidePrefill":     "\u4f60\u597d\uff0c\u8f9b\u82e6\u5e2e\u5fd9\u540c\u6b65\u4e00\u4e0b\u670d\u52a1\u5b8c\u6210\u72b6\u6001\u3002",
			"primaryContactKey":       contactKey,
			"primaryContactText":      contactText,
			"primaryContactPrefill":   contactPrefill,
			"timelinePrefill":         "\u4f60\u597d\uff0c\u670d\u52a1\u5df2\u7ecf\u5b8c\u6210\uff0c\u9ebb\u70e6\u786e\u8ba4\u4e00\u4e0b\u3002",
			"idleTimelineText":        "\u5f53\u524d\u8282\u70b9\u65e0\u9700\u989d\u5916\u64cd\u4f5c",
			"timelineActionText":      "\u8054\u7cfb\u73a9\u5bb6",
			"proofNoteTemplate":       "\u4ea4\u4ed8\u51ed\u8bc1ID\uff1a{fileIds}",
		},
	})
}

func deliveryActivityRows(game games.Game) []map[string]interface{} {
	rows := []map[string]interface{}{
		{"label": "服务类型", "value": auditGameTypeText(game), "type": "blue"},
	}
	if duration := deliveryDurationText(game); duration != "" {
		rows = append(rows, map[string]interface{}{"label": "服务时长", "value": duration})
	}
	if startText := deliveryTimeText(game.StartAt, game.CreatedAt); startText != "" {
		rows = append(rows, map[string]interface{}{"label": "开始时间", "value": startText})
	}
	if endText := deliveryTimeText(game.EndAt, time.Time{}); endText != "" {
		rows = append(rows, map[string]interface{}{"label": "完成时间", "value": endText})
	}
	return rows
}

func deliveryDurationText(game games.Game) string {
	start, startOK := parseGameDisplayTime(game.StartAt)
	end, endOK := parseGameDisplayTime(game.EndAt)
	if !startOK || !endOK || !end.After(start) {
		return ""
	}
	hours := end.Sub(start).Hours()
	if hours == float64(int(hours)) {
		return strconv.Itoa(int(hours)) + "小时（已完成）"
	}
	return strconv.FormatFloat(hours, 'f', 1, 64) + "小时（已完成）"
}

func deliveryTimeText(value string, fallback time.Time) string {
	if parsed, ok := parseGameDisplayTime(value); ok {
		return parsed.Format("01-02 15:04")
	}
	if !fallback.IsZero() {
		return fallback.Format("01-02 15:04")
	}
	return ""
}

func parseGameDisplayTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func (s *Server) serviceConfirmTimelineSteps(game games.Game, memberRoles map[int64]string, viewerUserID int64) []map[string]interface{} {
	expertID := int64(0)
	playerID := game.CreatorUserID
	pairedSelected := false
	viewerRole := memberRoles[viewerUserID]
	if viewerRole == "expert" {
		expertID = viewerUserID
		if bound := s.boundPlayersForExpert(game.ID, viewerUserID); len(bound) > 0 {
			playerID = bound[0]
			pairedSelected = true
		}
	} else if viewerRole == "member" || viewerRole == "player" || viewerUserID == game.CreatorUserID {
		playerID = viewerUserID
		for memberID, role := range memberRoles {
			if role == "expert" && containsUserID(s.boundPlayersForExpert(game.ID, memberID), viewerUserID) {
				expertID = memberID
				pairedSelected = true
				break
			}
		}
	}
	for memberID, role := range memberRoles {
		switch role {
		case "expert":
			if expertID <= 0 {
				expertID = memberID
			}
		case "member", "creator":
			if !pairedSelected && (playerID <= 0 || memberID == game.CreatorUserID) {
				playerID = memberID
			}
		}
	}

	confirmed := make(map[int64]games.ServiceConfirmItem)
	if _, items, ok := s.games.ServiceConfirmForGame(game.ID); ok {
		for _, item := range items {
			confirmed[item.UserID] = item
		}
	}

	hasExpert := expertID > 0
	expertConfirmed := !hasExpert || !confirmed[expertID].CreatedAt.IsZero()
	playerConfirmed := playerID > 0 && !confirmed[playerID].CreatedAt.IsZero()
	bothConfirmed := expertConfirmed && playerConfirmed

	expertDesc := "等待行家确认服务完成"
	expertTitle := "等待行家确认"
	expertState := "active"
	expertLineState := "pending"
	if !hasExpert {
		expertTitle = "无需行家确认"
		expertDesc = "本局未配置行家，玩家确认后进入评价"
		expertState = "done"
		expertLineState = "confirmed"
	} else if expertConfirmed {
		expertTitle = "行家已确认完成"
		expertDesc = "行家标记服务已完成 " + confirmed[expertID].CreatedAt.Format("01-02 15:04")
		expertState = "done"
		expertLineState = "confirmed"
	}

	playerDesc := "需玩家确认服务已达标"
	playerTitle := "等待玩家确认"
	playerState := "pending"
	playerLineState := "pending"
	if playerConfirmed {
		playerTitle = "玩家已确认完成"
		playerDesc = s.inGameDisplayName(playerID, "玩家") + "已确认服务已达标"
		playerState = "done"
		playerLineState = "confirmed"
	} else if expertConfirmed {
		playerState = "active"
	}

	archiveDesc := "双方确认后服务自动归档"
	archiveState := "pending"
	if bothConfirmed {
		archiveDesc = "服务已进入归档"
		archiveState = "active"
	}

	return []map[string]interface{}{
		{"index": 1, "key": "expert_confirmed", "title": expertTitle, "desc": expertDesc, "state": expertState, "lineState": expertLineState},
		{"index": 2, "key": "player_confirmed", "title": playerTitle, "desc": playerDesc, "state": playerState, "lineState": playerLineState, "action": "contact_player"},
		{"index": 3, "key": "service_archive", "title": "服务归档", "desc": archiveDesc, "state": archiveState},
	}
}

func deliveryContactForRole(role string) (string, string, string) {
	if role == "member" || role == "player" {
		return "contact_expert", "联系行家", "你好，想和你确认一下本次服务完成情况。"
	}
	return "contact_player", "联系玩家", "你好，麻烦确认一下服务完成情况。"
}

func (s *Server) gameCompletionState(game games.Game, userID int64) map[string]interface{} {
	roles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	role := gameMemberRole(game, userID, s.games.IsMember(game.ID, userID), roles)
	confirmed := make(map[int64]bool)
	if _, items, ok := s.games.ServiceConfirmForGame(game.ID); ok {
		for _, item := range items {
			confirmed[item.UserID] = true
		}
	}
	expertIDs := make([]int64, 0)
	playerIDs := make([]int64, 0)
	for _, memberID := range s.games.Members(game.ID) {
		switch gameMemberRole(game, memberID, true, roles) {
		case "expert":
			expertIDs = append(expertIDs, memberID)
		case "member":
			playerIDs = append(playerIDs, memberID)
		}
	}
	pairedExpertIDs := make([]int64, 0, len(expertIDs))
	pairedPlayerIDs := make([]int64, 0, len(playerIDs))
	seenPlayers := make(map[int64]bool)
	for _, expertID := range expertIDs {
		boundPlayers := s.boundPlayersForExpert(game.ID, expertID)
		if len(boundPlayers) == 0 {
			continue
		}
		pairedExpertIDs = append(pairedExpertIDs, expertID)
		for _, playerID := range boundPlayers {
			if !seenPlayers[playerID] {
				seenPlayers[playerID] = true
				pairedPlayerIDs = append(pairedPlayerIDs, playerID)
			}
		}
	}
	if len(pairedExpertIDs) > 0 && len(pairedPlayerIDs) > 0 {
		expertIDs = pairedExpertIDs
		playerIDs = pairedPlayerIDs
	}
	allExpertsConfirmed := true
	for _, expertID := range expertIDs {
		if !confirmed[expertID] {
			allExpertsConfirmed = false
			break
		}
	}
	hasConfirmed := confirmed[userID]
	canConfirm := game.GameSource != "admin" && game.Status == "pending_confirm" && !hasConfirmed
	if role == "member" {
		canConfirm = canConfirm && allExpertsConfirmed && containsUserID(playerIDs, userID)
	} else if role == "expert" {
		canConfirm = canConfirm && containsUserID(expertIDs, userID)
	} else {
		canConfirm = false
	}
	waitingText := ""
	switch {
	case game.GameSource == "admin" || game.Status == "pending_review":
		waitingText = "组局已结束，可以进入评价"
	case hasConfirmed:
		waitingText = "你已确认完成，等待其他成员确认"
	case role == "member" && !allExpertsConfirmed:
		waitingText = "请等待行家先确认服务完成"
	case role == "guide" || role == "main_guide":
		waitingText = "领路人无需确认，等待行家和玩家完成确认"
	case canConfirm:
		waitingText = "请确认本次服务已经完成"
	}
	return map[string]interface{}{
		"role":                role,
		"canConfirm":          canConfirm,
		"hasConfirmed":        hasConfirmed,
		"allExpertsConfirmed": allExpertsConfirmed,
		"expertCount":         len(expertIDs),
		"playerCount":         len(playerIDs),
		"waitingText":         waitingText,
	}
}

func containsUserID(items []int64, userID int64) bool {
	for _, item := range items {
		if item == userID {
			return true
		}
	}
	return false
}

func (s *Server) canAccessServiceConfirmation(game games.Game, userID int64) bool {
	roles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	role := gameMemberRole(game, userID, s.games.IsMember(game.ID, userID), roles)
	if role != "expert" && role != "member" && role != "player" {
		return false
	}

	hasBoundPair := false
	for memberID, memberRole := range roles {
		if memberRole != "expert" {
			continue
		}
		boundPlayers := s.boundPlayersForExpert(game.ID, memberID)
		if len(boundPlayers) == 0 {
			continue
		}
		hasBoundPair = true
		if userID == memberID || containsUserID(boundPlayers, userID) {
			return true
		}
	}

	return !hasBoundPair
}

func (s *Server) successfulExpertInvitation(gameID int64, userID int64) (games.Invitation, bool) {
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if invitation.GameID != gameID || invitation.TargetUserID != userID || guideProgressInvitationRole(invitation.Role) != "expert" {
			continue
		}
		_, _, allConfirmed := s.invitationConfirmedParties(invitation)
		if allConfirmed {
			return invitation, true
		}
	}
	return games.Invitation{}, false
}

func (s *Server) gameCollaboration(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/collaboration")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.games.IsMember(game.ID, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	messages, _ := s.im.Messages(userID, game.ID)
	now := time.Now()
	httpx.OK(w, map[string]interface{}{
		"gameId":        game.ID,
		"title":         game.Title,
		"status":        game.Status,
		"statusText":    homeGameStatusText(game.Status),
		"dayText":       collaborationDayText(game.CreatedAt, now),
		"progress":      collaborationProgress(game),
		"members":       s.collaborationMembers(userID, game),
		"membersText":   s.collaborationMembersText(game),
		"messages":      s.collaborationMessages(messages),
		"actions":       collaborationActions(game, userID),
		"currentUserId": userID,
		"serverTime":    now.Format(time.RFC3339),
	})
}

func (s *Server) requestGameCompletion(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/completion-request")
	if !ok {
		return
	}
	previousGame, previousErr := s.games.Get(gameID)
	alreadyRequested := previousErr == nil && (previousGame.Status == "pending_confirm" || previousGame.Status == "pending_review")
	wasReviewable := currentGameReviewable(s.games, gameID)
	game, err := s.games.RequestCompletion(userID, gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.im.ReadOnlyRoomsByGameIDs([]int64{gameID}, "game_ended")
	directReview := game.Status == "pending_review"
	if directReview && !wasReviewable {
		s.reviews.MarkGameReviewable(gameID)
		s.reviews.AwardCompletedGame(gameID)
		s.createCoGameConnections(gameID)
	}
	if !alreadyRequested {
		recipients := s.games.Members(gameID)
		if !directReview {
			recipients = make([]int64, 0)
			for _, memberID := range s.games.Members(gameID) {
				if s.userRoleForGame(game, memberID) == "expert" {
					recipients = append(recipients, memberID)
				}
			}
		}
		for _, memberID := range recipients {
			title := "组局进入完成确认"
			content := "《" + game.Title + "》已发起结束，请等待行家先确认完成，随后由玩家确认。"
			notifyType := "game_completion_requested"
			if directReview {
				title = "组局已结束"
				content = "《" + game.Title + "》已结束，现在可以进入评价。"
				notifyType = "game_ended"
			}
			s.notices.Create(notifications.CreateRequest{
				UserID:      memberID,
				NotifyType:  notifyType,
				Title:       title,
				Content:     content,
				BizType:     "game",
				BizID:       game.ID,
				NeedWechat:  true,
				WechatState: "pending",
				WechatData: map[string]string{
					"thing1": game.Title,
					"thing2": title,
				},
			})
		}
		if directReview {
			s.createReviewRemindNotifications(gameID)
		}
	}
	s.ensureCompletionIMReminder(userID, game, directReview)
	nextRoute := "/pages/game/collaboration/index?gameId=" + strconv.FormatInt(gameID, 10)
	if directReview {
		nextRoute = "/pages/game/review/index?gameId=" + strconv.FormatInt(gameID, 10)
	}
	s.recordBehavior(userID, "request_game_completion", "game", gameID, map[string]interface{}{
		"gameSource": game.GameSource,
		"gameStatus": game.Status,
	})
	httpx.OK(w, map[string]interface{}{
		"game":             game,
		"directReview":     directReview,
		"alreadyRequested": alreadyRequested,
		"nextRoute":        nextRoute,
	})
}

func (s *Server) ensureCompletionIMReminder(senderID int64, game games.Game, directReview bool) {
	messages, err := s.im.Messages(senderID, game.ID)
	if err == nil {
		for _, message := range messages {
			if message.Type == "service_confirm_remind" {
				return
			}
		}
	}
	s.sendCompletionIMReminder(senderID, game, directReview)
}

func (s *Server) sendCompletionIMReminder(senderID int64, game games.Game, directReview bool) {
	title := "本局进入完成确认"
	subtitle := "请行家先确认完成，随后由玩家确认；全部确认后开启评价。"
	actionRoute := "pages/game/delivery/index?gameId=" + strconv.FormatInt(game.ID, 10)
	actionText := "去确认"
	if directReview {
		title = "本局已结束"
		subtitle = "本局已结束，现在可以进入评价。"
		actionRoute = "pages/game/review/index?gameId=" + strconv.FormatInt(game.ID, 10)
		actionText = "去评价"
	}
	payload := map[string]interface{}{
		"title":            title,
		"subtitle":         subtitle,
		"guideLabel":       "下一步",
		"guideName":        map[bool]string{true: "评价", false: "完成确认"}[directReview],
		"playerInfoLabel":  "后续操作",
		"memberName":       "行家和玩家",
		"memberDesc":       subtitle,
		"avatarText":       "确",
		"dateText":         "状态已同步到所有局成员",
		"location":         "可从系统通知或本卡片进入下一步",
		"acceptButtonText": actionText,
		"actionRoute":      actionRoute,
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = s.im.SendSystem(senderID, game.ID, im.SendRequest{
		MessageType: "service_confirm_remind",
		Content:     string(content),
	})
}

func collaborationDayText(startedAt time.Time, now time.Time) string {
	elapsedDays := int(now.Sub(startedAt).Hours()/24) + 1
	if startedAt.IsZero() || elapsedDays < 1 {
		elapsedDays = 1
	}
	return "\u7b2c " + strconv.Itoa(elapsedDays) + " \u5929 / \u5171 15 \u5929"
}

func collaborationProgress(game games.Game) map[string]interface{} {
	percent := collaborationProgressPercent(game.Status)
	return map[string]interface{}{
		"percent": percent,
		"title":   "\u8fdb\u5ea6 " + strconv.Itoa(percent) + "%",
		"tasks": []map[string]interface{}{
			{"key": "group_success", "title": "\u5df2\u5b8c\u6210", "desc": "\u7ec4\u5c40\u5df2\u6210\u529f\uff0c\u6210\u5458\u5df2\u8fdb\u5165\u534f\u4f5c", "state": "completed"},
			{"key": "service_active", "title": "\u8fdb\u884c\u4e2d", "desc": collaborationActiveTaskDesc(game.Status), "state": collaborationActiveTaskState(game.Status)},
			{"key": "service_done", "title": "\u5f85\u5b8c\u6210", "desc": "\u786e\u8ba4\u670d\u52a1\u3001\u8bc4\u4ef7\u53cc\u65b9\u5e76\u5f52\u6863", "state": collaborationDoneTaskState(game.Status)},
		},
	}
}

func collaborationProgressPercent(status string) int {
	switch status {
	case "completed":
		return 100
	case "pending_review":
		return 85
	case "pending_confirm":
		return 70
	case "in_progress":
		return 46
	default:
		return 25
	}
}

func collaborationActiveTaskDesc(status string) string {
	switch status {
	case "pending_review":
		return "\u670d\u52a1\u5df2\u786e\u8ba4\uff0c\u7b49\u5f85\u8bc4\u4ef7\u548c\u7ed3\u7b97"
	case "pending_confirm":
		return "\u7b49\u5f85\u6210\u5458\u786e\u8ba4\u670d\u52a1\u5b8c\u6210"
	case "completed":
		return "\u670d\u52a1\u5df2\u5b8c\u6210"
	default:
		return "\u6210\u5458\u6b63\u5728\u534f\u4f5c\u63a8\u8fdb\u4ea4\u4ed8"
	}
}

func collaborationActiveTaskState(status string) string {
	if status == "pending_review" || status == "completed" {
		return "completed"
	}
	return "active"
}

func collaborationDoneTaskState(status string) string {
	if status == "completed" {
		return "completed"
	}
	if status == "pending_review" {
		return "active"
	}
	return "pending"
}

func (s *Server) collaborationMembers(userID int64, game games.Game) []map[string]interface{} {
	memberIDs := s.games.Members(game.ID)
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	items := make([]map[string]interface{}, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		profile := s.inGameIdentity(memberID, "\u6210\u5458")
		name := profile.DisplayName
		role := gameMemberRole(game, memberID, true, memberRoles)
		items = append(items, map[string]interface{}{
			"id":            memberID,
			"userId":        memberID,
			"name":          name,
			"realName":      profile.RealName,
			"displayName":   profile.DisplayName,
			"role":          role,
			"roleText":      collaborationRoleText(game, memberID, role),
			"avatarText":    profile.AvatarText,
			"isCurrentUser": memberID == userID,
		})
	}
	return items
}

func (s *Server) collaborationMembersText(game games.Game) string {
	memberIDs := s.games.Members(game.ID)
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	parts := make([]string, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		name := s.inGameDisplayName(memberID, "\u6210\u5458")
		role := gameMemberRole(game, memberID, true, memberRoles)
		if label := collaborationRoleText(game, memberID, role); label != "" {
			name += " (" + label + ")"
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, " \u00b7 ")
}

func collaborationRoleText(game games.Game, userID int64, role string) string {
	if game.GameSource != "admin" && game.CreatorUserID == userID {
		return "\u53d1\u8d77\u4eba"
	}
	return gameRoleLabel(role)
}

func (s *Server) collaborationMessages(messages []im.Message) []map[string]interface{} {
	start := 0
	if len(messages) > 5 {
		start = len(messages) - 5
	}
	items := make([]map[string]interface{}, 0, len(messages)-start)
	for _, message := range messages[start:] {
		profile := s.inGameIdentity(message.SenderID, "\u6210\u5458")
		name := profile.DisplayName
		items = append(items, map[string]interface{}{
			"id":           message.ID,
			"messageId":    message.ID,
			"senderId":     message.SenderID,
			"senderUserId": message.SenderID,
			"senderName":   name,
			"realName":     profile.RealName,
			"displayName":  profile.DisplayName,
			"avatarText":   profile.AvatarText,
			"name":         name,
			"content":      message.Content,
			"messageType":  message.Type,
			"status":       message.Status,
			"createdAt":    message.CreatedAt.Format(time.RFC3339),
		})
	}
	return items
}

func collaborationActions(game games.Game, userID int64) map[string]interface{} {
	canManage := game.CreatorUserID == userID || game.MainGuideUserID == userID
	canEnd := canManage && game.Status == "in_progress"
	return map[string]interface{}{
		"canManageMembers": canManage,
		"canEndGame":       canEnd,
		"completionMode":   map[bool]string{true: "direct_review", false: "ordered_confirm"}[game.GameSource == "admin"],
		"manageRoute":      "pages/game/participants/index?gameId=" + strconv.FormatInt(game.ID, 10),
		"endConfirmRoute":  "pages/game/collaboration/index?gameId=" + strconv.FormatInt(game.ID, 10),
		"reviewRoute":      "pages/game/review/index?gameId=" + strconv.FormatInt(game.ID, 10),
	}
}

func gameRoleLabel(role string) string {
	switch role {
	case "expert":
		return "\u884c\u5bb6"
	case "guide", "main_guide":
		return "\u9886\u8def\u4eba"
	default:
		return "\u73a9\u5bb6"
	}
}

func gameRoleAvatarClass(role string) string {
	switch role {
	case "guide", "main_guide":
		return "orange"
	case "expert":
		return "blue"
	default:
		return "pink"
	}
}

func successStageText(status string) string {
	switch status {
	case "pending_review", "completed":
		return "\u5f85\u8bc4\u4ef7\u7ed3\u7b97"
	case "in_progress", "pending_confirm":
		return "\u5f85\u4ea4\u4ed8\u670d\u52a1"
	default:
		return "\u5df2\u6210\u5c40"
	}
}

func successFundAmount(game games.Game) int64 {
	if game.GameType == "free" {
		return 0
	}
	return 80000
}

func successFundStatus(game games.Game, amountCent int64) map[string]interface{} {
	if game.GameType == "free" || amountCent <= 0 {
		return map[string]interface{}{
			"title":      "\u4e00\u671f\u514d\u8d39\u5c40",
			"desc":       "\u672c\u5c40\u4e0d\u53d1\u8d77\u771f\u5b9e\u652f\u4ed8\uff0c\u4e5f\u4e0d\u4ea7\u751f\u8d44\u91d1\u6258\u7ba1",
			"amount":     0,
			"amountText": "\u00a50",
			"status":     "free_no_pay",
			"progress":   100,
			"stepLabels": []string{"\u65e0\u9700\u652f\u4ed8", "\u670d\u52a1\u4e2d", "\u5f85\u786e\u8ba4"},
		}
	}

	return map[string]interface{}{
		"title":      "\u8d44\u91d1\u5df2\u6258\u7ba1",
		"desc":       "\u670d\u52a1\u5b8c\u6210\u540e\u81ea\u52a8\u7ed3\u7b97",
		"amount":     amountCent,
		"amountText": "\u00a5" + strconv.FormatInt(amountCent/100, 10),
		"status":     "escrowed",
		"progress":   33,
		"stepLabels": []string{"\u5df2\u6258\u7ba1", "\u670d\u52a1\u4e2d", "\u5df2\u5b8c\u6210"},
	}
}

func successExpertPageTexts() map[string]string {
	return map[string]string{
		"navTitle":          "组局成功",
		"successHeading":    "组局成功!",
		"subtitleFallback":  "成功详情以接口返回为准",
		"groupSectionTitle": "三方连接群",
		"onlineText":        "在线",
		"chatButtonText":    "进入群聊",
		"activityTitle":     "活动信息",
		"fundTitle":         "资金状态",
		"fundEmptyText":     "暂无资金状态",
		"stepsTitle":        "下一步",
		"stepsEmptyText":    "暂无下一步动作",
		"manageButtonText":  "进入局管理",
		"loadFailedText":    "组局成功详情加载失败，请稍后重试",
		"chatPrefill":       "我已进入三方群，准备确认后续服务安排。",
	}
}

func successGuidePageTexts() map[string]string {
	return map[string]string{
		"navTitle":          "组局成功",
		"heroTitle":         "恭喜！组局成功",
		"heroDesc":          "你成功促成了这次连接",
		"timelineTitle":     "成局历程",
		"followTitle":       "后续跟进",
		"followEmptyText":   "暂无后续跟进项",
		"followMissingText": "未找到跟进项",
		"loadFailedText":    "领路人成功详情加载失败，请稍后重试",
		"followFailedText":  "跟进记录失败，继续打开页面",
		"defaultImPrefill":  "我来跟进一下本次组局双方反馈。",
		"shareTitle":        "组局成功",
		"shareButtonText":   "分享成局喜悦",
	}
}

func (s *Server) gameGuideSuccessDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/guide-success-detail")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.canGuideFollowUp(userID, game) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	invitation, hasSuccessInvitation := s.guideSuccessInvitation(game.ID, userID)
	if !hasSuccessInvitation {
		httpx.Error(w, http.StatusConflict, httpx.CodeValidationError, "组局尚未成功")
		return
	}
	playerInvitation, expertInvitation, _ := s.invitationConfirmedParties(invitation)
	playerID := playerInvitation.TargetUserID
	expertID := expertInvitation.TargetUserID
	roomID := int64(0)
	if room, err := s.im.RoomForGame(userID, game.ID); err == nil {
		roomID = room.ID
	}
	httpx.OK(w, map[string]interface{}{
		"gameId":    game.ID,
		"pageTexts": successGuidePageTexts(),
		"viewer": map[string]interface{}{
			"userId": userID,
			"role":   s.userRoleForGame(game, userID),
		},
		"timeline": []map[string]interface{}{
			{"title": "\u53d1\u8d77\u5f15\u8350", "desc": "\u4f60\u5411\u53cc\u65b9\u53d1\u9001\u4e86\u7ec4\u5c40\u9080\u8bf7", "time": game.CreatedAt.Format("01-02 15:04")},
			{"title": "\u73a9\u5bb6\u786e\u8ba4", "desc": s.inGameDisplayName(playerID, "\u73a9\u5bb6") + "\u786e\u8ba4\u53c2\u52a0\u7ec4\u5c40", "time": game.CreatedAt.Format("01-02 15:04")},
			{"title": "\u884c\u5bb6\u786e\u8ba4", "desc": s.inGameDisplayName(expertID, "\u884c\u5bb6") + "\u786e\u8ba4\u53c2\u52a0\u7ec4\u5c40", "time": game.CreatedAt.Format("01-02 15:04")},
			{"title": "\u7ec4\u5c40\u6210\u529f\uff01", "desc": "\u53cc\u65b9\u5df2\u5efa\u7acb\u8fde\u63a5\uff0c\u8fdb\u5165\u4ea4\u4ed8\u9636\u6bb5", "time": game.CreatedAt.Format("01-02 15:04"), "active": true},
		},
		"party": map[string]interface{}{
			"confirmedText": "",
			"cardClass":     "success-guide-party-card",
			"cardStyle":     "width: 684rpx; height: 446rpx; margin: 32rpx auto 0; box-shadow: none;",
			"titleClass":    "regular",
			"player":        guideSuccessPartyMember(s.inGameDisplayName(playerID, "\u73a9\u5bb6"), playerID, "\u73a9\u5bb6", "pink"),
			"expert":        guideSuccessPartyMember(s.inGameDisplayName(expertID, "\u884c\u5bb6"), expertID, "\u884c\u5bb6", "blue"),
		},
		"reward": map[string]interface{}{
			"show":  true,
			"value": "+50",
			"label": "积分奖励已到账",
		},
		"followUps": []map[string]interface{}{
			{"key": "schedule", "title": "\u67e5\u770b\u7ec4\u5c40\u65e5\u7a0b", "desc": "\u67e5\u770b\u8be5\u5c40\u8be6\u60c5\u4e0e\u4ea4\u4ed8\u8fdb\u5ea6", "iconSrc": "https://static.haowan.net.cn/miniprogram/pages/game/success-guide/assets/follow-schedule.png", "iconClass": "schedule", "theme": "blue", "target": map[string]interface{}{"type": "game_detail", "gameId": game.ID}},
			{"key": "feedback", "title": "\u8be2\u95ee\u53cc\u65b9\u53cd\u9988", "desc": "\u8fdb\u5165\u4e09\u65b9\u7fa4\u4e86\u89e3\u4ea4\u6d41\u60c5\u51b5", "iconSrc": "https://static.haowan.net.cn/miniprogram/pages/game/success-guide/assets/follow-feedback.png", "iconClass": "feedback", "theme": "purple", "target": map[string]interface{}{"type": "im_room", "gameId": game.ID, "roomId": roomID, "prefill": "\u6211\u6765\u8ddf\u8fdb\u4e00\u4e0b\u672c\u6b21\u7ec4\u5c40\u53cc\u65b9\u53cd\u9988\u3002"}},
			{"key": "deal", "title": "\u4fc3\u6210\u4ea4\u6613", "desc": "\u8bb0\u5f55\u6216\u8ddf\u8fdb\u53cc\u65b9\u5408\u4f5c\u610f\u5411", "iconSrc": "https://static.haowan.net.cn/miniprogram/pages/game/success-guide/assets/follow-deal.png", "iconClass": "deal", "theme": "orange", "target": map[string]interface{}{"type": "referral_record", "gameId": game.ID}},
		},
	})
}

func guideSuccessExpertID(game games.Game, memberIDs []int64) int64 {
	if game.MainGuideUserID > 0 {
		return game.MainGuideUserID
	}
	for _, memberID := range memberIDs {
		if memberID != game.CreatorUserID {
			return memberID
		}
	}
	return game.CreatorUserID
}

func guideSuccessPartyMember(name string, userID int64, role string, avatarClass string) map[string]interface{} {
	return map[string]interface{}{
		"id":          userID,
		"userId":      userID,
		"avatarClass": avatarClass,
		"name":        name,
		"avatarText":  avatarTextForName(name, userID),
		"role":        role,
		"state":       "\u5df2\u786e\u8ba4",
		"stateClass":  "confirmed",
	}
}

func (s *Server) guideSuccessInvitation(gameID int64, userID int64) (games.Invitation, bool) {
	for _, invitation := range s.games.InvitationsForUser(userID) {
		if invitation.GameID != gameID || invitation.InviterID != userID {
			continue
		}
		if _, _, allConfirmed := s.invitationConfirmedParties(invitation); allConfirmed {
			return invitation, true
		}
	}
	return games.Invitation{}, false
}

func (s *Server) createGuideFollowUp(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/guide-follow-ups")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.canGuideFollowUp(userID, game) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden")
		return
	}
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "action required")
		return
	}
	if !validGuideFollowUpAction(action) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid follow-up action")
		return
	}
	target := guideFollowUpTarget(action, game.ID)
	note := strings.TrimSpace(req.Note)
	s.recordBehavior(userID, "guide_follow_up_"+action, "game", game.ID, map[string]interface{}{"action": action, "note": note})
	s.createGuideFollowUpNotifications(userID, game, action)
	httpx.OK(w, map[string]interface{}{
		"gameId": game.ID,
		"action": action,
		"status": "recorded",
		"target": target,
	})
}

func (s *Server) canGuideFollowUp(userID int64, game games.Game) bool {
	return s.games.IsMember(game.ID, userID) && (game.CreatorUserID == userID || game.MainGuideUserID == userID)
}

func validGuideFollowUpAction(action string) bool {
	switch action {
	case "schedule", "feedback", "deal":
		return true
	default:
		return false
	}
}

func guideFollowUpTarget(action string, gameID int64) map[string]interface{} {
	switch action {
	case "schedule":
		return map[string]interface{}{"type": "game_detail", "gameId": gameID}
	case "feedback":
		return map[string]interface{}{"type": "im_room", "gameId": gameID, "prefill": "\u6211\u6765\u8ddf\u8fdb\u4e00\u4e0b\u672c\u6b21\u7ec4\u5c40\u53cc\u65b9\u53cd\u9988\u3002"}
	default:
		return map[string]interface{}{"type": "referral_record", "gameId": gameID}
	}
}

func (s *Server) createGuideFollowUpNotifications(userID int64, game games.Game, action string) {
	title := "\u9886\u8def\u4eba\u5df2\u8ddf\u8fdb"
	content := map[string]string{
		"schedule": "\u9886\u8def\u4eba\u67e5\u770b\u4e86\u7ec4\u5c40\u65e5\u7a0b",
		"feedback": "\u9886\u8def\u4eba\u6b63\u5728\u8be2\u95ee\u53cc\u65b9\u53cd\u9988",
		"deal":     "\u9886\u8def\u4eba\u6b63\u5728\u4fc3\u6210\u5408\u4f5c\u4ea4\u6613",
	}[action]
	for _, memberID := range s.games.Members(game.ID) {
		if memberID == userID {
			continue
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     memberID,
			NotifyType: "guide_follow_up",
			Title:      title,
			Content:    content,
			BizType:    "game",
			BizID:      game.ID,
		})
	}
}

func (s *Server) buildGameRelation(userID int64, game games.Game) GameMyRelationDTO {
	isCreator := game.CreatorUserID == userID
	isMember := s.games.IsMember(game.ID, userID)
	signupOpen := games.CanApplyWithinSignupWindow(game, time.Now())
	canManageProgress := isCreator || (game.MainGuideUserID > 0 && game.MainGuideUserID == userID)
	startableStatus := game.Status == "recruiting" || game.Status == "full"
	relation := GameMyRelationDTO{
		Role:       s.userRoleForGame(game, userID),
		IsCreator:  isCreator,
		IsMember:   isMember,
		CanApply:   game.Status == "recruiting" && !isMember && signupOpen,
		CanAudit:   isCreator && game.Status == "recruiting",
		CanStart:   canManageProgress && startableStatus && game.CurrentPlayers >= game.MinPlayers,
		CanEnterIM: isMember && (game.Status == "in_progress" || game.Status == "pending_confirm" || game.Status == "pending_review" || game.Status == "completed"),
		CanConfirm: isMember && (game.Status == "in_progress" || game.Status == "pending_confirm"),
	}
	if game.Status == "recruiting" && !isMember && game.MaxPlayers > 0 && game.CurrentPlayers >= game.MaxPlayers {
		relation.CanApply = false
		relation.ApplyDisabledReason = "该局已满员"
	} else if game.Status == "recruiting" && !isMember && !signupOpen {
		relation.ApplyDisabledReason = "不在报名时间内"
	}
	// ApplicationsForUser 在内存实现中来自 map，不能依赖遍历顺序；取该局
	// 最新的一条申请，避免旧的 rejected/pending 记录覆盖当前状态。
	var latest games.Application
	var hasLatest bool
	for _, app := range s.games.ApplicationsForUser(userID) {
		if app.GameID != game.ID {
			continue
		}
		if !hasLatest || app.CreatedAt.After(latest.CreatedAt) || (app.CreatedAt.Equal(latest.CreatedAt) && app.ID > latest.ID) {
			latest = app
			hasLatest = true
		}
	}
	if hasLatest {
		relation.ApplicationID = latest.ID
		relation.ApplicationStatus = latest.Status
		if latest.Status == "pending" {
			relation.CanApply = false
			relation.ApplyDisabledReason = "报名审核中"
		}
	}
	return relation
}

func gameMemberRole(game games.Game, userID int64, isMember bool, memberRoles map[int64]string) string {
	storedRole := normalizeGameMemberRole(memberRoles[userID])
	if !isMember && storedRole == "" {
		return "guest"
	}
	// 后台创建的局不允许出现玩家以外的局内角色。
	if game.GameSource == "admin" {
		return "member"
	}
	if storedRole != "" {
		return storedRole
	}
	if game.MainGuideUserID > 0 && game.MainGuideUserID == userID {
		return "main_guide"
	}
	return "member"
}

func gameMemberRoleMap(items []games.MemberRole) map[int64]string {
	result := make(map[int64]string, len(items))
	for _, item := range items {
		if item.UserID > 0 {
			result[item.UserID] = normalizeGameMemberRole(item.Role)
		}
	}
	return result
}

func normalizeGameMemberRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "expert":
		return "expert"
	case "guide", "leader":
		return "guide"
	case "main_guide":
		return "main_guide"
	case "member", "player", "creator":
		return "member"
	default:
		return ""
	}
}

func (s *Server) userRoleForGame(game games.Game, userID int64) string {
	roles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	return gameMemberRole(game, userID, s.games.IsMember(game.ID, userID), roles)
}

func (s *Server) userPlayerForGame(game games.Game, userID int64) bool {
	return s.canRequestPlayerCancel(game, userID)
}

func (s *Server) userExpertForGame(game games.Game, userID int64) bool {
	return s.userRoleForGame(game, userID) == "expert"
}

func (s *Server) canRequestPlayerCancel(game games.Game, userID int64) bool {
	if game.CreatorUserID == userID {
		return true
	}
	if !s.games.IsMember(game.ID, userID) {
		return false
	}
	if game.MainGuideUserID == userID {
		return false
	}
	return s.userRoleForGame(game, userID) == "member"
}

func (s *Server) userGuideForGame(game games.Game, userID int64) bool {
	role := s.userRoleForGame(game, userID)
	return role == "guide" || role == "main_guide"
}

func (s *Server) approveGameForLocal(w http.ResponseWriter, r *http.Request) {
	if s.productionMode {
		http.NotFound(w, r)
		return
	}
	id, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/approve-local")
	if !ok {
		return
	}
	game, err := s.games.ApproveGame(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, 40421, "局不存在")
		return
	}
	httpx.OK(w, game)
}

func (s *Server) adminAuditGame(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/audit")
	if !ok {
		return
	}
	var req struct {
		Approve bool   `json:"approve"`
		Remark  string `json:"remark"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	remark := strings.TrimSpace(req.Remark)
	if !req.Approve && remark == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "驳回审核必须填写原因")
		return
	}
	game, err := s.reviewGameAudit(gameID, req.Approve, remark)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if req.Approve {
		s.notifyGameApproved(game)
	} else {
		s.notifyGameRejected(game, remark)
	}
	s.recordOperation(r, "game:audit", "game", strconv.FormatInt(game.ID, 10), map[string]interface{}{"status": game.Status, "remark": remark})
	httpx.OK(w, game)
}

func (s *Server) adminBatchAuditGames(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GameIDs []int64 `json:"gameIds"`
		Approve bool    `json:"approve"`
		Remark  string  `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid batch audit request")
		return
	}
	if len(req.GameIDs) == 0 || len(req.GameIDs) > 100 || (!req.Approve && strings.TrimSpace(req.Remark) == "") {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid batch audit request")
		return
	}
	results := make([]batchMutationResult, 0, len(req.GameIDs))
	success := 0
	for _, gameID := range uniquePositiveIDs(req.GameIDs) {
		game, err := s.reviewGameAudit(gameID, req.Approve, strings.TrimSpace(req.Remark))
		result := batchMutationResult{ID: gameID}
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.Status = game.Status
			success++
			if req.Approve {
				s.notifyGameApproved(game)
			} else {
				s.notifyGameRejected(game, strings.TrimSpace(req.Remark))
			}
			s.recordOperation(r, "game:batch_audit", "game", strconv.FormatInt(game.ID, 10), map[string]interface{}{"status": game.Status, "remark": req.Remark})
		}
		results = append(results, result)
	}
	httpx.OK(w, map[string]interface{}{
		"items":   results,
		"success": success,
		"failed":  len(results) - success,
		"total":   len(results),
	})
}

func (s *Server) reviewGameAudit(gameID int64, approve bool, remark string) (games.Game, error) {
	var game games.Game
	var err error
	if approve {
		game, err = s.games.ApproveGame(gameID)
	} else {
		game, err = s.games.RejectGame(gameID, remark)
	}
	return game, err
}

func (s *Server) applyGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	id, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/applications")
	if !ok {
		return
	}
	if game, err := s.games.Get(id); err == nil && game.MaxPlayers > 0 && game.CurrentPlayers >= game.MaxPlayers {
		writeGameError(w, games.ErrFull)
		return
	}
	var req struct {
		Reason   string  `json:"reason"`
		Role     string  `json:"role"`
		RoleType string  `json:"roleType"`
		FileIDs  []int64 `json:"fileIds"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if config := s.currentGameApplicationConfig(); config.RequireRealname && !s.identity.IsVerified(userID) {
		httpx.Error(w, http.StatusForbidden, 40341, "申请入局前请先完成实名认证")
		return
	}
	requestedRole := strings.ToLower(strings.TrimSpace(firstNonEmpty(req.RoleType, req.Role)))
	snapshot := s.profiles.RoleSnapshot(userID)
	switch requestedRole {
	case "guide", "leader", "main_guide":
		if snapshot.RoleStatusMap["guide"] != "approved" {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "当前账号未开通领路人身份")
			return
		}
		req.RoleType = "guide"
	case "expert", "master":
		if snapshot.RoleStatusMap["expert"] != "approved" {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "当前账号未开通行家身份")
			return
		}
		req.RoleType = "expert"
	default:
		req.RoleType = "player"
	}
	app, err := s.games.Apply(userID, id, games.ApplyRequest{Reason: req.Reason, Role: req.Role, RoleType: req.RoleType, FileIDs: req.FileIDs})
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game, gameErr := s.games.Get(app.GameID); gameErr == nil {
		s.notices.Create(notifications.CreateRequest{
			UserID:     game.CreatorUserID,
			NotifyType: "game_apply",
			Title:      "\u6536\u5230\u65b0\u7684\u5165\u5c40\u7533\u8bf7",
			Content:    "\u6709\u4eba\u7533\u8bf7\u52a0\u5165\u300a" + game.Title + "\u300b\uff0c\u8bf7\u53ca\u65f6\u5ba1\u6838\u3002",
			BizType:    "game_application",
			BizID:      app.ID,
		})
	}
	s.recordBehavior(userID, "apply_game", "game", id, map[string]interface{}{"applicationId": app.ID})
	httpx.OK(w, app)
}

func (s *Server) createGameInvitation(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if !s.userCanGenerateInvitations(userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅行家或领路人可生成组局邀请")
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/guide-invitations")
	if !ok {
		return
	}
	var req games.InvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "璇锋眰鍙傛暟閿欒")
		return
	}
	invitation, err := s.games.CreateInvitation(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "create_game_invitation", "game", gameID, map[string]interface{}{"invitationId": invitation.ID, "targetUserId": invitation.TargetUserID})
	s.notices.Create(notifications.CreateRequest{
		UserID:     invitation.TargetUserID,
		NotifyType: "game_invitation",
		Title:      "\u7ec4\u5c40\u9080\u8bf7",
		Content:    invitation.Message,
		BizType:    "game_invitation",
		BizID:      invitation.ID,
	})
	httpx.OK(w, map[string]interface{}{"invitation": invitation})
}

func (s *Server) respondGameInvitation(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	invitationID, ok := invitationIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req games.InvitationRespondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "璇锋眰鍙傛暟閿欒")
		return
	}
	invitation, app, err := s.games.RespondInvitation(userID, invitationID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	successRoute := ""
	if invitation.Status == "accepted" {
		s.connections.UpsertPair(invitation.InviterID, invitation.TargetUserID, "guide_match", "guide_match", invitation.ID, 3)
		if strings.TrimSpace(invitation.InviteGroupID) != "" {
			s.createPairedInvitationProgressNotification(invitation)
		}
		if app.Status == "approved" {
			s.createGameInvitationSuccessNotifications(invitation)
			if _, _, allConfirmed := s.invitationConfirmedParties(invitation); allConfirmed && guideProgressInvitationRole(invitation.Role) == "expert" {
				successRoute, _ = gameInvitationSuccessRoute(invitation.GameID, "expert")
			}
		} else {
			s.createInvitationReviewNotification(invitation, app)
		}
	}
	s.recordBehavior(userID, "respond_game_invitation", "game", invitation.GameID, map[string]interface{}{"invitationId": invitation.ID, "status": invitation.Status, "applicationId": app.ID})
	response := map[string]interface{}{"invitation": invitation, "application": app}
	if successRoute != "" {
		response["successRoute"] = successRoute
	}
	httpx.OK(w, response)
}

func (s *Server) myApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.games.ApplicationsForUser(userID)})
}

func (s *Server) receivedApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	gameID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("gameId")), 10, 64)
	items := filterReceivedApplications(s.games.ApplicationsForCreator(userID), status, gameID)
	items = s.excludeAutoApprovedPairedInvitationApplications(items)
	httpx.OK(w, map[string]interface{}{"items": s.buildReceivedApplicationItems(userID, items), "gameId": gameID})
}

func (s *Server) excludeAutoApprovedPairedInvitationApplications(items []games.Application) []games.Application {
	result := make([]games.Application, 0, len(items))
	for _, app := range items {
		autoApproved := false
		if app.Status == "approved" {
			for _, invitation := range s.games.InvitationsForUser(app.UserID) {
				if invitation.ApplicationID == app.ID && strings.TrimSpace(invitation.InviteGroupID) != "" {
					autoApproved = true
					break
				}
			}
		}
		if !autoApproved {
			result = append(result, app)
		}
	}
	return result
}

func (s *Server) buildReceivedApplicationItems(reviewerID int64, items []games.Application) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, s.buildReceivedApplicationItem(reviewerID, item))
	}
	return result
}

func (s *Server) buildReceivedApplicationItem(reviewerID int64, app games.Application) map[string]interface{} {
	roleKey, roleLabel := auditApplicationRole(app.Role)
	statusTitle, countdown := auditApplicationStatus(app.Status)
	applicantName := s.inGameDisplayName(app.UserID, roleLabel)
	createdAtText := ""
	if !app.CreatedAt.IsZero() {
		createdAtText = app.CreatedAt.Local().Format("2006-01-02 15:04")
	}
	item := map[string]interface{}{
		"id":            app.ID,
		"applicationId": app.ID,
		"gameId":        app.GameID,
		"userId":        app.UserID,
		"status":        app.Status,
		"statusText":    statusTitle,
		"role":          app.Role,
		"roleKey":       roleKey,
		"roleLabel":     roleLabel,
		"nickname":      applicantName,
		"userName":      applicantName,
		"realName":      applicantName,
		"displayName":   applicantName,
		"avatarUrl":     s.imUserAvatarURL(app.UserID),
		"avatarText":    avatarTextForName(applicantName, app.UserID),
		"reason":        app.Reason,
		"rejectReason":  app.RejectReason,
		"fileIds":       app.FileIDs,
		"createdAt":     app.CreatedAt,
		"createdAtText": createdAtText,
		"applyTime":     createdAtText,
	}
	game, err := s.games.Get(app.GameID)
	if err != nil {
		return item
	}
	reviewerUserID := reviewerID
	if reviewerUserID <= 0 {
		reviewerUserID = game.CreatorUserID
	}
	reviewerName := s.inGameDisplayName(reviewerUserID, "\u53d1\u8d77\u4eba")
	timeText, durationText := auditGameSchedule(game)
	locationText := firstNonEmpty(strings.TrimSpace(game.Address), strings.TrimSpace(game.CityName))
	activityType := auditGameTypeText(game)
	budgetText := "\u514d\u8d39"
	if game.Price > 0 {
		budgetText = "\u00a5" + strconv.FormatFloat(game.Price, 'f', 2, 64)
	}
	confirmedCount := 1
	if app.Status == "approved" {
		confirmedCount = 2
	}

	item["gameTitle"] = game.Title
	item["title"] = game.Title
	item["gameType"] = game.GameType
	item["gameTypeText"] = activityType
	item["gameTimeText"] = timeText
	item["locationText"] = locationText
	item["activityType"] = activityType
	item["serviceDurationText"] = durationText
	item["clientBudgetText"] = budgetText
	item["guideName"] = reviewerName
	item["guideAvatarText"] = avatarTextForName(reviewerName, reviewerUserID)
	item["totalCount"] = 2
	item["confirmedCount"] = confirmedCount
	item["scenario"] = "direct_application"
	item["sourceType"] = "self_apply"
	item["viewerRole"] = "organizer"
	item["partyRole"] = roleKey
	item["detailDisplay"] = map[string]interface{}{
		"scenario":           "direct_application",
		"sourceType":         "self_apply",
		"viewerRole":         "organizer",
		"partyRole":          roleKey,
		"pageTitle":          "\u5ba1\u6838\u5165\u5c40\u7533\u8bf7",
		"referralText":       "\u6536\u5230\u65b0\u7684\u5165\u5c40\u7533\u8bf7",
		"applicantTitle":     roleLabel + "\u4fe1\u606f",
		"confirmTitle":       "\u7ec4\u5c40\u4fe1\u606f\u786e\u8ba4",
		"showRelation":       true,
		"showPortfolio":      false,
		"showSession":        true,
		"showConfirm":        true,
		"showRecommend":      false,
		"showOptions":        app.Status == "pending",
		"showNotice":         app.Status == "pending",
		"showActionBar":      app.Status == "pending",
		"reviewReadonlyText": statusTitle,
		"status": map[string]interface{}{
			"title":     statusTitle,
			"quote":     app.Reason,
			"guideName": reviewerName,
			"countdown": countdown,
		},
		"relation": map[string]interface{}{
			"title":          "\u5ba1\u6838\u5173\u7cfb\u56fe",
			"totalCount":     2,
			"confirmedCount": confirmedCount,
			"expert": map[string]interface{}{
				"avatarText": avatarTextForName(reviewerName, reviewerUserID),
				"name":       reviewerName,
				"roleText":   "\u53d1\u8d77\u4eba",
			},
			"guide": map[string]interface{}{
				"iconSrc": "/pages/game/guide-chat/assets/icon-invite.png",
				"name":    "\u5ba1\u6838",
			},
			"player": map[string]interface{}{
				"avatarText": avatarTextForName(applicantName, app.UserID),
				"name":       applicantName,
				"confirmed":  app.Status == "approved",
				"statusText": statusTitle,
			},
		},
		"player": map[string]interface{}{
			"requirementConfirmed":  app.Status != "pending",
			"requirementStatusText": statusTitle,
			"confirmed":             app.Status == "approved",
			"statusText":            statusTitle,
			"name":                  applicantName,
			"avatarText":            avatarTextForName(applicantName, app.UserID),
			"desc":                  roleLabel,
			"roleKey":               roleKey,
			"roleName":              roleLabel,
			"tags":                  []string{roleLabel},
			"needText":              firstNonEmpty(nonEmptyAuditPrefix("\u7533\u8bf7\u8bf4\u660e\uff1a", strings.TrimSpace(app.Reason)), "\u672a\u586b\u5199\u7533\u8bf7\u8bf4\u660e"),
			"remark":                nonEmptyAuditPrefix("\u7533\u8bf7\u65f6\u95f4\uff1a", createdAtText),
		},
		"gameInfo": map[string]interface{}{
			"topic":           game.Title,
			"time":            firstNonEmpty(timeText, "\u672a\u8bbe\u7f6e"),
			"location":        firstNonEmpty(locationText, "\u672a\u8bbe\u7f6e"),
			"activityType":    activityType,
			"serviceDuration": firstNonEmpty(durationText, "\u672a\u8bbe\u7f6e"),
			"clientBudget":    budgetText,
		},
		"sessionInfo": []map[string]interface{}{
			{"key": "topic", "label": "\u7ec4\u5c40\u4e3b\u9898", "value": game.Title, "iconSrc": "/pages/game/detail/assets/icon-calendar.png", "iconClass": "topic"},
			{"key": "time", "label": "\u65f6\u95f4", "value": firstNonEmpty(timeText, "\u672a\u8bbe\u7f6e"), "iconSrc": "/pages/game/detail/assets/icon-clock.png", "iconClass": "time"},
			{"key": "location", "label": "\u5730\u70b9", "value": firstNonEmpty(locationText, "\u672a\u8bbe\u7f6e"), "actionText": "\u5730\u56fe\u4f4d\u7f6e", "iconSrc": "/pages/game/detail/assets/icon-location.png", "iconClass": "place"},
		},
		"confirmRows": []map[string]interface{}{
			{"label": "\u6d3b\u52a8\u7c7b\u578b", "value": activityType},
			{"label": "\u670d\u52a1\u65f6\u957f", "value": firstNonEmpty(durationText, "\u672a\u8bbe\u7f6e")},
			{"label": "\u5ba2\u6237\u9884\u7b97", "value": budgetText},
		},
	}
	return item
}

func auditApplicationRole(role string) (string, string) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "expert":
		return "expert", "\u884c\u5bb6"
	case "guide", "leader", "main_guide":
		return "guide", "\u9886\u8def\u4eba"
	default:
		return "player", "\u73a9\u5bb6"
	}
}

func auditApplicationStatus(status string) (string, string) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "pass", "passed":
		return "\u5df2\u901a\u8fc7\u7533\u8bf7", "\u5df2\u5904\u7406"
	case "rejected", "reject":
		return "\u5df2\u62d2\u7edd\u7533\u8bf7", "\u5df2\u5904\u7406"
	default:
		return "\u7b49\u5f85\u4f60\u5ba1\u6838", "\u5f85\u5904\u7406"
	}
}

func auditGameTypeText(game games.Game) string {
	if value := firstNonEmpty(strings.TrimSpace(game.PrimaryCategoryText), strings.TrimSpace(game.SecondaryCategoryText)); value != "" {
		return value
	}
	switch strings.ToLower(strings.TrimSpace(game.GameType)) {
	case "free":
		return "\u514d\u8d39\u5c40"
	case "paid":
		return "\u4ed8\u8d39\u5c40"
	case "condition":
		return "\u6761\u4ef6\u5c40"
	default:
		return strings.TrimSpace(game.GameType)
	}
}

func auditGameSchedule(game games.Game) (string, string) {
	startText := strings.TrimSpace(game.StartAt)
	endText := strings.TrimSpace(game.EndAt)
	start, startErr := time.Parse(time.RFC3339, startText)
	end, endErr := time.Parse(time.RFC3339, endText)
	if startErr == nil {
		dateText := start.Local().Format("2006-01-02 15:04")
		if endErr == nil {
			dateText += " - " + end.Local().Format("15:04")
			duration := end.Sub(start)
			if duration > 0 {
				return dateText, strconv.FormatFloat(duration.Hours(), 'f', 1, 64) + "\u5c0f\u65f6"
			}
		}
		return dateText, ""
	}
	return firstNonEmpty(startText, endText), ""
}

func nonEmptyAuditPrefix(prefix string, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return prefix + strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func filterReceivedApplications(items []games.Application, status string, gameID int64) []games.Application {
	if status == "" && gameID <= 0 {
		return items
	}
	filtered := make([]games.Application, 0, len(items))
	for _, item := range items {
		if status != "" && item.Status != status {
			continue
		}
		if gameID > 0 && item.GameID != gameID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func (s *Server) reviewApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	id, ok := applicationIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req struct {
		Approve      bool   `json:"approve"`
		RejectReason string `json:"rejectReason"`
		Remark       string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	rejectReason := strings.TrimSpace(firstNonEmpty(req.RejectReason, req.Remark))
	if !req.Approve && rejectReason == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "拒绝申请必须填写原因")
		return
	}
	app, err := s.games.ReviewApplicationWithReason(userID, id, req.Approve, rejectReason)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game, gameErr := s.games.Get(app.GameID); gameErr == nil {
		notifyType := "application_rejected"
		title := "\u5165\u5c40\u7533\u8bf7\u672a\u901a\u8fc7"
		content := "\u4f60\u7533\u8bf7\u52a0\u5165\u7684\u300a" + game.Title + "\u300b\u672a\u901a\u8fc7\u5ba1\u6838\u3002"
		if app.RejectReason != "" {
			content += "原因：" + app.RejectReason
		}
		if req.Approve {
			notifyType = "application_approved"
			title = "\u5165\u5c40\u7533\u8bf7\u5df2\u901a\u8fc7"
			content = "\u4f60\u7533\u8bf7\u52a0\u5165\u7684\u300a" + game.Title + "\u300b\u5df2\u901a\u8fc7\u5ba1\u6838\u3002"
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     app.UserID,
			NotifyType: notifyType,
			Title:      title,
			Content:    content,
			BizType:    "game",
			BizID:      game.ID,
		})
	}
	if req.Approve {
		s.createGameInvitationSuccessNotificationsForApplication(app)
		if game, gameErr := s.games.Get(app.GameID); gameErr == nil {
			s.createGameReadyToStartNotification(game)
		}
	}
	httpx.OK(w, s.buildReceivedApplicationItem(userID, app))
}

func (s *Server) notifyGameApproved(game games.Game) {
	s.notices.Create(notifications.CreateRequest{
		UserID:     game.CreatorUserID,
		NotifyType: "game_approved",
		Title:      "\u7ec4\u5c40\u5ba1\u6838\u901a\u8fc7",
		Content:    "\u4f60\u7684\u7ec4\u5c40\u300a" + game.Title + "\u300b\u5df2\u901a\u8fc7\u5ba1\u6838\uff0c\u73b0\u5df2\u8fdb\u5165\u62db\u52df\u4e2d\u3002",
		BizType:    "game",
		BizID:      game.ID,
	})
}

func (s *Server) notifyGameRejected(game games.Game, reason string) {
	content := "你的组局《" + game.Title + "》未通过审核。"
	if strings.TrimSpace(reason) != "" {
		content += "原因：" + strings.TrimSpace(reason)
	}
	s.notices.Create(notifications.CreateRequest{
		UserID:     game.CreatorUserID,
		NotifyType: "game_rejected",
		Title:      "组局审核未通过",
		Content:    content,
		BizType:    "game",
		BizID:      game.ID,
	})
}

func applicationIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	text := strings.Trim(path, "/")
	parts := strings.Split(text, "/")
	if len(parts) < 3 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "申请 ID 错误")
		return 0, false
	}
	id, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "申请 ID 错误")
		return 0, false
	}
	return id, true
}

func invitationIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	text := strings.Trim(path, "/")
	parts := strings.Split(text, "/")
	if len(parts) < 3 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "閭€绾?ID 閿欒")
		return 0, false
	}
	id, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "閭€绾?ID 閿欒")
		return 0, false
	}
	return id, true
}

func (s *Server) cancelApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	id, ok := applicationIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	app, err := s.games.CancelApplication(userID, id)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, app)
}

func (s *Server) playerCancelRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/player-cancel")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.canRequestPlayerCancel(game, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "no permission to request player cancel")
		return
	}
	if game.Status != "in_progress" {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "current game cannot be canceled")
		return
	}
	var req struct {
		ServiceOrderID         string  `json:"serviceOrderId"`
		Ref                    string  `json:"ref"`
		ReasonKey              string  `json:"reasonKey"`
		ReasonText             string  `json:"reasonText"`
		CompensationRate       float64 `json:"compensationRate"`
		CompensationAmountText string  `json:"compensationAmountText"`
		PlatformFeeText        string  `json:"platformFeeText"`
		PayAmountText          string  `json:"payAmountText"`
		RefundAmountText       string  `json:"refundAmountText"`
		ContractAmount         float64 `json:"contractAmount"`
		ServedDurationText     string  `json:"servedDurationText"`
		TotalDurationText      string  `json:"totalDurationText"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.ReasonKey = strings.TrimSpace(req.ReasonKey)
	req.ReasonText = strings.TrimSpace(req.ReasonText)
	if req.ReasonKey == "" || req.ReasonText == "" || !validCancelReason(req.ReasonKey, s.currentGameCancelConfig().Player.ReasonOptions) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "cancel reason required")
		return
	}
	if req.CompensationRate < 0 || req.CompensationRate > 100 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid compensation rate")
		return
	}
	amountCent := successFundAmount(game)
	compensationCent := amountCent * int64(req.CompensationRate) / 100
	platformFeeCent := compensationCent / 10
	extra := map[string]interface{}{
		"serviceOrderId":         serviceOrderID(gameID),
		"ref":                    serviceOrderID(gameID),
		"reasonKey":              req.ReasonKey,
		"reasonText":             req.ReasonText,
		"compensationRate":       req.CompensationRate,
		"compensationAmountText": serviceAmountText(compensationCent),
		"platformFeeText":        serviceAmountText(platformFeeCent),
		"payAmountText":          serviceAmountText(compensationCent + platformFeeCent),
		"refundAmountText":       serviceAmountText(cancelMaxInt64(0, amountCent-compensationCent-platformFeeCent)),
		"contractAmount":         float64(amountCent) / 100,
		"servedDurationText":     firstNonEmpty(strings.TrimSpace(game.StartAt), "待后端确认"),
		"totalDurationText":      firstNonEmpty(strings.TrimSpace(game.EndAt), "待后端确认"),
	}
	s.recordBehavior(userID, "player_cancel_request", "game", gameID, extra)
	s.createPlayerCancelNotifications(userID, game, req.ReasonText)
	if game.CreatorUserID != userID {
		result, err := s.games.Exit(userID, gameID)
		if err != nil {
			writeGameError(w, err)
			return
		}
		credit := s.reviews.DeductCredit(userID, gameID, "player_cancel_service")
		result.CreditLogID = credit.ID
		if err := s.games.RecordExitCredit(gameID, userID, credit.ID); err != nil {
			writeGameError(w, err)
			return
		}
		httpx.OK(w, map[string]interface{}{
			"game":   result.Game,
			"gameId": gameID,
			"status": "exited",
			"exit":   result,
			"credit": credit,
			"cancelRequest": map[string]interface{}{
				"gameId":                 gameID,
				"userId":                 userID,
				"serviceOrderId":         extra["serviceOrderId"],
				"reasonKey":              req.ReasonKey,
				"reasonText":             req.ReasonText,
				"compensationRate":       req.CompensationRate,
				"compensationAmountText": extra["compensationAmountText"],
				"payAmountText":          extra["payAmountText"],
				"refundAmountText":       extra["refundAmountText"],
			},
		})
		return
	}
	canceledGame, err := s.games.CancelService(gameID, "player_cancel_service")
	if err != nil {
		writeGameError(w, err)
		return
	}
	credit := s.reviews.DeductCredit(userID, gameID, "player_cancel_service")
	httpx.OK(w, map[string]interface{}{
		"game":   canceledGame,
		"gameId": gameID,
		"status": "canceled",
		"credit": credit,
		"cancelRequest": map[string]interface{}{
			"gameId":                 gameID,
			"userId":                 userID,
			"serviceOrderId":         extra["serviceOrderId"],
			"reasonKey":              req.ReasonKey,
			"reasonText":             req.ReasonText,
			"compensationRate":       req.CompensationRate,
			"compensationAmountText": extra["compensationAmountText"],
			"payAmountText":          extra["payAmountText"],
			"refundAmountText":       extra["refundAmountText"],
		},
	})
}

func (s *Server) gameCancelDetail(w http.ResponseWriter, r *http.Request, role string) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	suffix := "/player-cancel-detail"
	if role == "expert" {
		suffix = "/expert-cancel-detail"
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", suffix)
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if role == "player" {
		if !s.canRequestPlayerCancel(game, userID) {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "no permission to request player cancel")
			return
		}
	} else if game.CreatorUserID != userID && game.MainGuideUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "no permission to request expert cancel")
		return
	}

	amountCent := successFundAmount(game)
	isFree := amountCent <= 0
	minRate, maxRate, suggestedRate, platformFeeRate := 0, 100, 20, 10
	if isFree {
		maxRate, suggestedRate, platformFeeRate = 0, 0, 0
	}
	compensationCent := amountCent * int64(suggestedRate) / 100
	platformFeeCent := compensationCent * int64(platformFeeRate) / 100
	statusType, statusText := serviceOrderStatus(game.Status)
	if statusType != "active" {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "current game cannot be canceled")
		return
	}

	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	playerID := s.firstMemberWithGameRole(game, userID, "member")
	if role == "player" {
		playerID = userID
	}
	expertID := s.firstMemberWithGameRole(game, userID, "expert")
	if role == "expert" && expertID == 0 {
		for _, memberID := range s.games.Members(game.ID) {
			if gameMemberRole(game, memberID, true, memberRoles) == "expert" {
				expertID = memberID
				break
			}
		}
	}
	playerName := s.inGameDisplayName(playerID, "玩家")
	hasExpert := expertID > 0
	expertName := "暂未分配行家"
	expertAvatarText := "局"
	if hasExpert {
		expertName = s.inGameDisplayName(expertID, "行家")
		expertAvatarText = avatarTextForName(expertName, expertID)
	}
	warningTitle := "取消将产生赔付"
	warningDesc := "赔付金额由后端根据订单金额和赔付比例计算。"
	if isFree {
		warningTitle = "免费局取消无需赔付"
		warningDesc = "当前为免费局，本次取消不会产生赔付金额，但会扣减信用分。"
	}

	httpx.OK(w, map[string]interface{}{
		"role": role, "gameId": game.ID, "serviceOrderId": serviceOrderID(game.ID), "ref": serviceOrderID(game.ID),
		"serviceTitle": game.Title, "statusText": statusText, "gameType": game.GameType,
		"playerId": playerID, "playerName": playerName, "playerAvatarText": avatarTextForName(playerName, playerID), "avatarText": avatarTextForName(playerName, playerID),
		"expertId": expertID, "expertName": expertName, "expertAvatarText": expertAvatarText, "hasExpert": hasExpert,
		"amount": float64(amountCent) / 100, "contractAmount": float64(amountCent) / 100,
		"amountText": serviceAmountText(amountCent), "contractAmountText": serviceAmountText(amountCent),
		"isFreeCancel": isFree, "minRate": minRate, "maxRate": maxRate, "suggestedRate": suggestedRate,
		"suggestionMinRate": suggestedRate, "suggestionMaxRate": suggestedRate, "platformFeeRate": platformFeeRate,
		"compensationRate": suggestedRate, "compensationAmountText": serviceAmountText(compensationCent),
		"platformFeeText": serviceAmountText(platformFeeCent), "payAmountText": serviceAmountText(compensationCent + platformFeeCent),
		"refundAmountText":   serviceAmountText(cancelMaxInt64(0, amountCent-compensationCent-platformFeeCent)),
		"servedDurationText": firstNonEmpty(strings.TrimSpace(game.StartAt), "待后端确认"),
		"totalDurationText":  firstNonEmpty(strings.TrimSpace(game.EndAt), "待后端确认"),
		"warningTitle":       warningTitle, "warningDesc": warningDesc,
	})
}

func cancelMaxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

func validCancelReason(key string, options []map[string]string) bool {
	for _, option := range options {
		if strings.TrimSpace(option["key"]) == key {
			return true
		}
	}
	return false
}

func serviceOrderStatusText(status string) string {
	_, text := serviceOrderStatus(status)
	return text
}

func (s *Server) createPlayerCancelNotifications(userID int64, game games.Game, reason string) {
	recipients := make(map[int64]bool)
	if game.CreatorUserID > 0 && game.CreatorUserID != userID {
		recipients[game.CreatorUserID] = true
	}
	if game.MainGuideUserID > 0 && game.MainGuideUserID != userID {
		recipients[game.MainGuideUserID] = true
	}
	playerName := s.inGameDisplayName(userID, "玩家")
	content := playerName + "取消了本次组局"
	if reason != "" {
		content += "，原因：" + reason
	}
	for recipientID := range recipients {
		s.notices.Create(notifications.CreateRequest{
			UserID:     recipientID,
			NotifyType: "player_cancel_request",
			Title:      "玩家取消组局",
			Content:    content,
			BizType:    "game",
			BizID:      game.ID,
		})
	}
}

func (s *Server) expertCancelRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/expert-cancel")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if game.CreatorUserID != userID && game.MainGuideUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "no permission to request expert cancel")
		return
	}
	if game.Status != "in_progress" {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "current game cannot be canceled")
		return
	}
	var req struct {
		ServiceOrderID         string  `json:"serviceOrderId"`
		PlayerID               string  `json:"playerId"`
		ReasonKey              string  `json:"reasonKey"`
		ReasonText             string  `json:"reasonText"`
		CompensationRate       float64 `json:"compensationRate"`
		CompensationAmountText string  `json:"compensationAmountText"`
		PlatformFeeText        string  `json:"platformFeeText"`
		PayAmountText          string  `json:"payAmountText"`
		ContractAmount         float64 `json:"contractAmount"`
		StatusText             string  `json:"statusText"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.ReasonKey = strings.TrimSpace(req.ReasonKey)
	req.ReasonText = strings.TrimSpace(req.ReasonText)
	if req.ReasonKey == "" || req.ReasonText == "" || !validCancelReason(req.ReasonKey, s.currentGameCancelConfig().Expert.ReasonOptions) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "cancel reason required")
		return
	}
	if req.CompensationRate < 0 || req.CompensationRate > 100 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid compensation rate")
		return
	}
	amountCent := successFundAmount(game)
	compensationCent := amountCent * int64(req.CompensationRate) / 100
	platformFeeCent := compensationCent / 10
	extra := map[string]interface{}{
		"serviceOrderId":         serviceOrderID(gameID),
		"playerId":               s.firstMemberWithGameRole(game, userID, "member"),
		"reasonKey":              req.ReasonKey,
		"reasonText":             req.ReasonText,
		"compensationRate":       req.CompensationRate,
		"compensationAmountText": serviceAmountText(compensationCent),
		"platformFeeText":        serviceAmountText(platformFeeCent),
		"payAmountText":          serviceAmountText(compensationCent + platformFeeCent),
		"contractAmount":         float64(amountCent) / 100,
		"statusText":             serviceOrderStatusText(game.Status),
	}
	s.recordBehavior(userID, "expert_cancel_request", "game", gameID, extra)
	s.createExpertCancelNotifications(userID, game, req.ReasonText)
	canceledGame, err := s.games.CancelService(gameID, "expert_cancel_service")
	if err != nil {
		writeGameError(w, err)
		return
	}
	credit := s.reviews.DeductCredit(userID, gameID, "expert_cancel_service")
	httpx.OK(w, map[string]interface{}{
		"game":   canceledGame,
		"gameId": gameID,
		"status": "canceled",
		"credit": credit,
		"cancelRequest": map[string]interface{}{
			"gameId":                 gameID,
			"userId":                 userID,
			"serviceOrderId":         extra["serviceOrderId"],
			"playerId":               extra["playerId"],
			"reasonKey":              req.ReasonKey,
			"reasonText":             req.ReasonText,
			"compensationRate":       req.CompensationRate,
			"compensationAmountText": extra["compensationAmountText"],
			"payAmountText":          extra["payAmountText"],
		},
	})
}

func (s *Server) createExpertCancelNotifications(userID int64, game games.Game, reason string) {
	expertName := s.inGameDisplayName(userID, "行家")
	content := expertName + "取消了本次组局"
	if reason != "" {
		content += "，原因：" + reason
	}
	for _, recipientID := range s.games.Members(game.ID) {
		if recipientID == userID {
			continue
		}
		s.notices.Create(notifications.CreateRequest{
			UserID:     recipientID,
			NotifyType: "expert_cancel_request",
			Title:      "行家取消组局",
			Content:    content,
			BizType:    "game",
			BizID:      game.ID,
		})
	}
}

func (s *Server) manualStart(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	id, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/manual-start")
	if !ok {
		return
	}
	var req struct {
		StartReason string `json:"startReason"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	startReason := strings.TrimSpace(req.StartReason)
	if startReason == "" {
		startReason = "发起人手动开始"
	}
	game, err := s.games.ManualStartWithReason(userID, id, startReason)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.im.EnsureRoom(id)
	s.createManualStartChatMessage(userID, game)
	s.createManualStartNotifications(userID, game)
	httpx.OK(w, game)
}

func (s *Server) createManualStartChatMessage(userID int64, game games.Game) {
	content := "组局开始了"
	if title := strings.TrimSpace(game.Title); title != "" {
		content = "《" + title + "》局开始了"
	}
	message, err := s.im.Send(userID, game.ID, im.SendRequest{
		MessageType: "text",
		Content:     content,
	})
	if err != nil {
		return
	}
	if s.imSocketHub != nil {
		s.imSocketHub.broadcast(message.RoomID, imSocketOutgoing{
			Type: "message",
			Data: s.inGameMessageDTO(message),
		})
	}
}

func (s *Server) createManualStartNotifications(userID int64, game games.Game) {
	starterName := s.inGameDisplayName(userID, "组建者")
	content := starterName + "已开始组局，请及时查看并参与"
	if title := strings.TrimSpace(game.Title); title != "" {
		content = "「" + title + "」" + content
	}
	for _, recipientID := range s.games.Members(game.ID) {
		s.notices.Create(notifications.CreateRequest{
			UserID:     recipientID,
			NotifyType: "game_started",
			Title:      "组局已开局",
			Content:    content,
			BizType:    "game",
			BizID:      game.ID,
		})
	}
}

func (s *Server) exitGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	id, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/exit")
	if !ok {
		return
	}
	result, err := s.games.Exit(userID, id)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if result.CreditDeduct {
		credit := s.reviews.DeductCredit(userID, id, result.Reason)
		result.CreditLogID = credit.ID
		if err := s.games.RecordExitCredit(id, userID, credit.ID); err != nil {
			writeGameError(w, err)
			return
		}
		s.createExitNotifications(result, credit.ChangeValue)
		httpx.OK(w, map[string]interface{}{"game": result.Game, "exit": result, "credit": credit})
		return
	}
	s.createExitNotifications(result, 0)
	httpx.OK(w, map[string]interface{}{"game": result.Game, "exit": result})
}

func (s *Server) createExitNotifications(result games.ExitResult, creditChange int) {
	content := "成员已退出局"
	if result.CreditDeduct {
		content = "成员退出局，已扣减信用分 " + strconv.Itoa(-creditChange)
	}
	if result.Game.CreatorUserID > 0 && result.Game.CreatorUserID != result.UserID {
		s.notices.Create(notifications.CreateRequest{
			UserID:     result.Game.CreatorUserID,
			NotifyType: "game_member_quit",
			Title:      "成员退出局",
			Content:    content,
			BizType:    "game",
			BizID:      result.GameID,
		})
	}
	s.notices.Create(notifications.CreateRequest{
		UserID:     result.UserID,
		NotifyType: "game_quit_result",
		Title:      "退出局结果",
		Content:    content,
		BizType:    "game",
		BizID:      result.GameID,
	})
}

func (s *Server) createProgressFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/progress-feedbacks")
	if !ok {
		return
	}
	var req games.ProgressFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	feedback, err := s.games.AddProgressFeedback(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_progress_feedback", "game", gameID, map[string]interface{}{"progress": feedback.Progress})
	httpx.OK(w, feedback)
}

func (s *Server) progressFeedbacks(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/progress-feedbacks")
	if !ok {
		return
	}
	items, err := s.games.ProgressFeedbacks(userID, gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	latestProgress := 0
	if len(items) > 0 {
		latestProgress = items[len(items)-1].Progress
	}
	httpx.OK(w, map[string]interface{}{"items": items, "latestProgress": latestProgress})
}

func (s *Server) createMilestone(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/milestones")
	if !ok {
		return
	}
	var req games.MilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.games.CreateMilestone(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "create_game_milestone", "game", gameID, map[string]interface{}{"milestoneId": item.ID})
	httpx.OK(w, item)
}

func (s *Server) milestones(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/milestones")
	if !ok {
		return
	}
	items, err := s.games.Milestones(userID, gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) updateMilestone(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, milestoneID, ok := gameAndMilestoneIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	var req games.MilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.games.UpdateMilestone(userID, gameID, milestoneID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "update_game_milestone", "game", gameID, map[string]interface{}{"milestoneId": item.ID})
	httpx.OK(w, item)
}

func (s *Server) createCheckin(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/checkins")
	if !ok {
		return
	}
	var req games.CheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.games.CreateCheckin(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_game_checkin", "game", gameID, map[string]interface{}{"checkinId": item.ID})
	httpx.OK(w, item)
}

func (s *Server) checkins(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/checkins")
	if !ok {
		return
	}
	items, err := s.games.Checkins(userID, gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) markCheckinInvalid(w http.ResponseWriter, r *http.Request) {
	checkinIDText := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/admin/game-checkins/"), "/mark-invalid")
	checkinID, err := strconv.ParseInt(strings.Trim(checkinIDText, "/"), 10, 64)
	if err != nil || checkinID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "打卡 ID 错误")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || len(req.Reason) > 300 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid reason")
		return
	}
	checkin, err := s.games.MarkCheckinInvalid(checkinID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordOperation(r, "game:checkin:mark_invalid", "game_checkin", strconv.FormatInt(checkin.ID, 10), map[string]interface{}{"gameId": checkin.GameID, "reason": req.Reason})
	httpx.OK(w, checkin)
}

func (s *Server) createRetrospective(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/retrospectives")
	if !ok {
		return
	}
	var req games.RetrospectiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.games.CreateRetrospective(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "submit_game_retrospective", "game", gameID, map[string]interface{}{"retrospectiveId": item.ID})
	httpx.OK(w, item)
}

func (s *Server) retrospectives(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/retrospectives")
	if !ok {
		return
	}
	items, err := s.games.Retrospectives(userID, gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) continueGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/continue")
	if !ok {
		return
	}
	var req games.ContinueDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	draft, err := s.games.ContinueDraft(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "create_continue_draft", "game", gameID, map[string]interface{}{"draftGameId": draft.Draft.ID})
	httpx.OK(w, draft)
}

func (s *Server) favoriteGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/favorite")
	if !ok {
		return
	}
	favorite, err := s.games.FavoriteGame(userID, gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "favorite_game", "game", gameID, nil)
	httpx.OK(w, favorite)
}

func (s *Server) unfavoriteGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/favorite")
	if !ok {
		return
	}
	if err := s.games.UnfavoriteGame(userID, gameID); err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "unfavorite_game", "game", gameID, nil)
	httpx.OK(w, map[string]interface{}{"gameId": gameID, "favorited": false})
}

func (s *Server) myFavoriteGames(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.games.FavoriteGames(userID), "pageConfig": s.currentMyGamesPageConfig()})
}

func (s *Server) currentMyGamesPageConfig() map[string]interface{} {
	var config map[string]interface{}
	if s.systemConfig != nil && s.systemConfig.Get(gameMyGamesPageConfigKey, &config) && len(config) > 0 {
		// 兼容旧版本将管理局错误放入“我的局”的配置，统一显示为“我受邀的”。
		if tabs, ok := config["categoryTabs"].([]interface{}); ok {
			for _, raw := range tabs {
				if item, ok := raw.(map[string]interface{}); ok && item["key"] == "created" {
					item["key"], item["text"] = "invited", "我受邀的"
				}
			}
		}
		return config
	}
	return map[string]interface{}{
		"pageTitle":     "\u6211\u7684\u5c40",
		"emptyText":     "\u6682\u65e0\u76f8\u5173\u5c40",
		"detailMissing": "\u6682\u65e0\u7ec4\u5c40\u8be6\u60c5",
		"actionMissing": "\u6682\u65e0\u53ef\u6267\u884c\u64cd\u4f5c",
		"categoryTabs": []map[string]interface{}{
			{"key": "joined", "text": "\u6211\u53c2\u4e0e\u7684"},
			{"key": "invited", "text": "\u6211\u53d7\u9080\u7684"},
			{"key": "favorite", "text": "\u6211\u6536\u85cf\u7684"},
		},
		"statusTabs": []map[string]interface{}{
			{"key": "all", "text": "\u5168\u90e8"},
			{"key": "active", "text": "\u8fdb\u884c\u4e2d"},
			{"key": "complete", "text": "\u5df2\u5b8c\u6210"},
			{"key": "overdue", "text": "\u8d85\u65f6"},
			{"key": "canceled", "text": "\u5df2\u53d6\u6d88"},
		},
	}
}

func myGamesPageConfigText(config map[string]interface{}, key string, fallback string) string {
	if value, ok := config[key].(string); ok {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return fallback
}

func (s *Server) adminUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "/favorites")
	if !ok {
		return
	}
	httpx.OK(w, map[string]interface{}{"items": s.games.FavoritesForUser(userID)})
}

func (s *Server) createAdminGameMilestone(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/milestones")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	var req games.MilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	item, err := s.games.CreateMilestone(game.CreatorUserID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordOperation(r, "game:milestone:create", "game", strconv.FormatInt(gameID, 10), map[string]interface{}{"milestoneId": item.ID})
	httpx.OK(w, item)
}

func (s *Server) adminGameMilestones(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/milestones")
	if !ok {
		return
	}
	items, err := s.games.AdminMilestones(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminGameCheckins(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/checkins")
	if !ok {
		return
	}
	items, err := s.games.AdminCheckins(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminGameRetrospectives(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/retrospectives")
	if !ok {
		return
	}
	items, err := s.games.AdminRetrospectives(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) adminGameContinueDrafts(w http.ResponseWriter, r *http.Request) {
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/admin/games/", "/continue-drafts")
	if !ok {
		return
	}
	items, err := s.games.AdminContinueDrafts(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func gameIDFromPath(w http.ResponseWriter, path string, prefix string, suffix string) (int64, bool) {
	idText := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	id, err := strconv.ParseInt(strings.Trim(idText, "/"), 10, 64)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "局 ID 错误")
		return 0, false
	}
	return id, true
}

func gameAndMilestoneIDFromPath(w http.ResponseWriter, path string) (int64, int64, bool) {
	text := strings.Trim(strings.TrimPrefix(path, "/api/app/games/"), "/")
	parts := strings.Split(text, "/")
	if len(parts) != 3 || parts[1] != "milestones" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "里程碑路径错误")
		return 0, 0, false
	}
	gameID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || gameID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "局 ID 错误")
		return 0, 0, false
	}
	milestoneID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || milestoneID <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "里程碑 ID 错误")
		return 0, 0, false
	}
	return gameID, milestoneID, true
}

func writeGameError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, games.ErrRealnameRequired):
		httpx.Error(w, http.StatusForbidden, 40341, "强实名未完成")
	case errors.Is(err, games.ErrGameNotFound), errors.Is(err, games.ErrApplicationNotFound), errors.Is(err, games.ErrInvitationNotFound), errors.Is(err, games.ErrCheckinNotFound), errors.Is(err, games.ErrMilestoneNotFound):
		httpx.Error(w, http.StatusNotFound, 40421, "对象不存在")
	case errors.Is(err, games.ErrDuplicateRetrospective):
		httpx.Error(w, http.StatusConflict, 40941, "已提交复盘")
	case errors.Is(err, games.ErrExpertConfirmRequired):
		httpx.Error(w, http.StatusConflict, 40924, "请等待行家先确认服务完成")
	case errors.Is(err, games.ErrInvitationPlayerPending):
		httpx.Error(w, http.StatusConflict, 40925, "请等待玩家先确认组局")
	case errors.Is(err, games.ErrSignupClosed):
		httpx.Error(w, http.StatusConflict, 40926, "不在报名时间范围内")
	case errors.Is(err, games.ErrGameNotRecruiting), errors.Is(err, games.ErrGameNotStartable), errors.Is(err, games.ErrApplicationNotPending), errors.Is(err, games.ErrInvitationNotPending), errors.Is(err, games.ErrGameNotConfirmable):
		httpx.Error(w, http.StatusConflict, 40921, "当前状态不可操作")
	case errors.Is(err, games.ErrAlreadyApplied), errors.Is(err, games.ErrAlreadyMember), errors.Is(err, games.ErrAlreadyInvited):
		httpx.Error(w, http.StatusConflict, 40923, "重复申请或已是成员")
	case errors.Is(err, games.ErrFull):
		httpx.Error(w, http.StatusConflict, 40922, "人数已满")
	case errors.Is(err, games.ErrInvalidGameInput):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid game input")
	case errors.Is(err, games.ErrInvalidProgress):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "进度参数错误")
	case errors.Is(err, games.ErrInvalidMilestone), errors.Is(err, games.ErrInvalidCheckin), errors.Is(err, games.ErrInvalidRetrospective):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请求参数错误")
	case errors.Is(err, games.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权操作")
	default:
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "操作失败")
	}
}
