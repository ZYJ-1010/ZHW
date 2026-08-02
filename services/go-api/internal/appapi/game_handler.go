package appapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
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
	ListStrict() ([]games.Game, error)
	Members(gameID int64) []int64
	MembersStrict(gameID int64) ([]int64, error)
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
	ApplicationsForUserStrict(userID int64) ([]games.Application, error)
	ApplicationsForCreator(userID int64) []games.Application
	ApplicationsForCreatorStrict(userID int64) ([]games.Application, error)
	InvitationsForUser(userID int64) []games.Invitation
	ManualStart(userID int64, gameID int64) (games.Game, error)
	ManualStartWithReason(userID int64, gameID int64, startReason string) (games.Game, error)
	RequestCompletion(userID int64, gameID int64) (games.Game, error)
	Exit(userID int64, gameID int64) (games.ExitResult, error)
	ExitWithCredit(userID int64, gameID int64, creditLogID int64) (games.ExitResult, error)
	RestoreMemberAfterExit(userID int64, gameID int64) error
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
	FavoriteGamesStrict(userID int64) ([]games.Favorite, error)
	FavoritesForUser(userID int64) []games.Favorite
	AllFavorites() []games.Favorite
	AllFavoritesStrict() ([]games.Favorite, error)
	StatsForUser(userID int64) games.UserStats
	StatsForUserStrict(userID int64) (games.UserStats, error)
	ParticipatedBetween(userID int64, start time.Time, end time.Time) bool
	ParticipatedBetweenStrict(userID int64, start time.Time, end time.Time) (bool, error)
	MemberRoles(gameID int64) []games.MemberRole
	IsMember(gameID int64, userID int64) bool
	IsMemberStrict(gameID int64, userID int64) (bool, error)
}

type GameDetailDTO struct {
	games.Game
	MyRelation         GameMyRelationDTO           `json:"myRelation"`
	DetailDisplay      GameDetailDisplayDTO        `json:"detailDisplay"`
	MemberIDs          []int64                     `json:"memberIds"`
	Members            []GameMemberDTO             `json:"members"`
	IsFavorited        bool                        `json:"isFavorited"`
	FavoriteCount      int                         `json:"favoriteCount"`
	Progress           GameProgressDTO             `json:"progress"`
	IM                 GameIMDTO                   `json:"im"`
	Review             GameReviewDTO               `json:"review"`
	ServiceConfirm     *ServiceConfirmDTO          `json:"serviceConfirm,omitempty"`
	PendingApplication *games.Application          `json:"pendingApplication,omitempty"`
	AuditRejectReason  string                      `json:"auditRejectReason,omitempty"`
	ShareComponent     gameShareComponentConfigDTO `json:"shareComponent"`
}

type GameListAvatarDTO struct {
	UserID    int64  `json:"userId"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type GameListItemDTO struct {
	games.Game
	PlayerAvatars  []GameListAvatarDTO         `json:"playerAvatars"`
	ShareComponent gameShareComponentConfigDTO `json:"shareComponent"`
}

type GameDetailDisplayDTO struct {
	StatusText              string                     `json:"statusText"`
	PendingApplicationCount int                        `json:"pendingApplicationCount"`
	Organizer               GameDetailOrganizerDTO     `json:"organizer"`
	PrimaryAction           GameDetailPrimaryActionDTO `json:"primaryAction"`
}

type GameDetailOrganizerDTO struct {
	UserID          int64              `json:"userId"`
	Name            string             `json:"name"`
	AvatarText      string             `json:"avatarText"`
	AvatarURL       string             `json:"avatarUrl,omitempty"`
	Role            string             `json:"role"`
	RoleLabel       string             `json:"roleLabel"`
	Rating          string             `json:"rating,omitempty"`
	RatingCount     int                `json:"ratingCount"`
	ExpertBlueBadge ExpertBlueBadgeDTO `json:"expertBlueBadge"`
}

type GameDetailPrimaryActionDTO struct {
	Text           string `json:"text"`
	Disabled       bool   `json:"disabled"`
	Action         string `json:"action"`
	Route          string `json:"route,omitempty"`
	ConfirmText    string `json:"confirmText,omitempty"`
	DisabledReason string `json:"disabledReason,omitempty"`
}

type GameMyRelationDTO struct {
	Role                string                   `json:"role"`
	AllowedRoles        []string                 `json:"allowedRoles"`
	ApplyRoleOptions    []GameApplyRoleOptionDTO `json:"applyRoleOptions"`
	IsCreator           bool                     `json:"isCreator"`
	IsMember            bool                     `json:"isMember"`
	CanApply            bool                     `json:"canApply"`
	ApplyDisabledReason string                   `json:"applyDisabledReason,omitempty"`
	CanAudit            bool                     `json:"canAudit"`
	CanStart            bool                     `json:"canStart"`
	CanEnterIM          bool                     `json:"canEnterIM"`
	CanConfirm          bool                     `json:"canConfirm"`
	CanReview           bool                     `json:"canReview"`
	ApplicationID       int64                    `json:"applicationId,omitempty"`
	ApplicationStatus   string                   `json:"applicationStatus,omitempty"`
}

type GameApplyRoleOptionDTO struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Enabled        bool   `json:"enabled"`
	DisabledReason string `json:"disabledReason,omitempty"`
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
	UserID          int64              `json:"userId"`
	Name            string             `json:"name,omitempty"`
	Nickname        string             `json:"nickname,omitempty"`
	RealName        string             `json:"realName,omitempty"`
	DisplayName     string             `json:"displayName,omitempty"`
	AvatarURL       string             `json:"avatarUrl,omitempty"`
	AvatarText      string             `json:"avatarText,omitempty"`
	Role            string             `json:"role"`
	RoleLabel       string             `json:"roleLabel,omitempty"`
	Position        string             `json:"position,omitempty"`
	Topic           string             `json:"topic,omitempty"`
	PrimaryTag      string             `json:"primaryTag,omitempty"`
	Location        string             `json:"location,omitempty"`
	IsCreator       bool               `json:"isCreator"`
	IsCurrentUser   bool               `json:"isCurrentUser"`
	Confirmed       bool               `json:"confirmed"`
	ExpertBlueBadge ExpertBlueBadgeDTO `json:"expertBlueBadge"`
}

type ServiceConfirmDTO struct {
	Confirm games.ServiceConfirm       `json:"confirm"`
	Items   []games.ServiceConfirmItem `json:"items"`
}

type gameCategoryOptionDTO struct {
	Key        string                    `json:"key"`
	Name       string                    `json:"name"`
	Icon       string                    `json:"icon,omitempty"`
	Visible    bool                      `json:"visible"`
	Order      int                       `json:"order"`
	Children   []gameCategoryOptionDTO   `json:"children,omitempty"`
	Tags       []gameCreateFormOptionDTO `json:"tags,omitempty"`
	Selectable bool                      `json:"selectable,omitempty"`
}

type gameCategoryConfigDTO struct {
	PrimaryCategories        []gameCategoryOptionDTO     `json:"primaryCategories"`
	TypeFilters              []gameCategoryOptionDTO     `json:"typeFilters"`
	LocationFilters          []gameCategoryOptionDTO     `json:"locationFilters"`
	SortOptions              []gameHallSortOptionDTO     `json:"sortOptions,omitempty"`
	EventActions             []string                    `json:"eventActions,omitempty"`
	ShareComponent           gameShareComponentConfigDTO `json:"shareComponent"`
	DefaultPrimaryCategory   string                      `json:"defaultPrimaryCategory"`
	DefaultSecondaryCategory string                      `json:"defaultSecondaryCategory"`
	DefaultType              string                      `json:"defaultType"`
	CreateForm               gameCreateFormConfigDTO     `json:"createForm"`
	Version                  string                      `json:"version"`
}

type gameShareComponentConfigDTO struct {
	Enabled        bool   `json:"enabled"`
	Variant        string `json:"variant"`
	Label          string `json:"label"`
	EnableInternal bool   `json:"enableInternal"`
	EnableWechat   bool   `json:"enableWechat"`
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
}

// gameCreateTemplateDTO is an operation-maintained starting point for the
// creation form. It deliberately stores only form defaults, never a real
// user's title, location, media or other business data.
type gameCreateTemplateDTO struct {
	Key               string   `json:"key"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	PrimaryCategory   string   `json:"primaryCategory"`
	SecondaryCategory string   `json:"secondaryCategory"`
	Participation     string   `json:"participation,omitempty"`
	Capacity          int      `json:"capacity,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	CompletionRules   []string `json:"completionRules,omitempty"`
	Visible           bool     `json:"visible"`
	Order             int      `json:"order"`
}

type gameCreateTemplateConfigDTO struct {
	Items   []gameCreateTemplateDTO `json:"items"`
	Version string                  `json:"version"`
}

const gameCategoryConfigKey = "game.category_config"
const gameCreateTemplateConfigKey = "game.create_template_config"
const gameApplicationConfigKey = "game.application_config"
const gameAuditConfigKey = "game.audit_config"
const gameConditionRuleConfigKey = "game.condition_rule_config"
const gameCancelConfigKey = "game.cancel_config"
const gameDeliveryPageConfigKey = "game.delivery_page_config"
const gameMyGamesPageConfigKey = "game.my_games_page_config"

type gameApplicationConfigDTO struct {
	AgreementTitle       string                            `json:"agreementTitle"`
	AgreementText        string                            `json:"agreementText"`
	RequireRealname      bool                              `json:"requireRealname"`
	RequireIntro         bool                              `json:"requireIntro"`
	RequireAgreement     bool                              `json:"requireAgreement"`
	AllowDuplicateApply  bool                              `json:"allowDuplicateApply"`
	UploadRequired       bool                              `json:"uploadRequired"`
	MaxUploadCount       int                               `json:"maxUploadCount"`
	AllowedUploadTypes   []string                          `json:"allowedUploadTypes"`
	MinIntroLength       int                               `json:"minIntroLength"`
	MaxIntroLength       int                               `json:"maxIntroLength"`
	MaxMessageLength     int                               `json:"maxMessageLength"`
	SearchEnabled        bool                              `json:"searchEnabled"`
	RecommendationHint   string                            `json:"recommendationHint"`
	SubscribeTemplateIDs []string                          `json:"subscribeTemplateIds,omitempty"`
	Texts                map[string]string                 `json:"texts,omitempty"`
	AuditPage            gameApplicationAuditPageConfigDTO `json:"auditPage,omitempty"`
	Version              string                            `json:"version"`
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

func (s *Server) gameCreateTemplateConfigHandler(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.currentGameCreateTemplateConfig())
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
			Tags:    categoryTags("轻松社交", "同城搭子", "桌游", "美食"),
		},
		{
			Key:     "task",
			Name:    "任务局",
			Icon:    "category-task",
			Visible: true,
			Order:   20,
			Tags:    categoryTags("找搭子", "项目协作", "创业", "资源对接"),
		},
		{
			Key:     "explore",
			Name:    "探索局",
			Icon:    "category-income",
			Visible: true,
			Order:   30,
			Tags:    categoryTags("城市探索", "户外", "打卡", "周末"),
		},
		{
			Key:     "growth",
			Name:    "成长局",
			Icon:    "category-growth",
			Visible: true,
			Order:   40,
			Tags:    categoryTags("学习", "健康", "习惯养成", "自我提升"),
		},
	}
	return gameCategoryConfigDTO{
		PrimaryCategories: primaryCategories,
		TypeFilters: []gameCategoryOptionDTO{
			{Key: "all", Name: "全部", Visible: true, Order: 0, Selectable: true},
		},
		LocationFilters: []gameCategoryOptionDTO{
			{Key: "all", Name: "全国", Visible: true, Order: 0, Selectable: true},
			{Key: "nearby", Name: "附近(50km)", Visible: true, Order: 10, Selectable: true},
		},
		SortOptions: []gameHallSortOptionDTO{
			{Key: "latest", Name: "最新发布", SortKey: "time", SortOrder: "desc"},
			{Key: "hot", Name: "热度最高", SortKey: "hot", SortOrder: "desc"},
			{Key: "distance", Name: "距离最近", SortKey: "distance", SortOrder: "asc"},
			{Key: "credit", Name: "信用优先", SortKey: "credit", SortOrder: "desc"},
		},
		EventActions:             []string{"分享", "关注", "打招呼"},
		ShareComponent:           gameShareComponentConfigDTO{Enabled: true, Variant: "channel_sheet", Label: "分享", EnableInternal: true, EnableWechat: true},
		DefaultPrimaryCategory:   "task",
		DefaultSecondaryCategory: "",
		DefaultType:              "all",
		CreateForm:               defaultGameCreateFormConfig(),
		Version:                  "2026-07-20-category-v2",
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
	}
}

func defaultGameCreateTemplateConfig() gameCreateTemplateConfigDTO {
	return gameCreateTemplateConfigDTO{
		Items: []gameCreateTemplateDTO{
			{Key: "social_board_game", Name: "同城桌游", Description: "适合线下轻松社交", PrimaryCategory: "social", Participation: "offline", Capacity: 5, Tags: []string{"同城搭子", "桌游"}, CompletionRules: []string{"time", "manual"}, Visible: true, Order: 10},
			{Key: "task_cocreation", Name: "项目共创", Description: "适合共同推进一个明确目标", PrimaryCategory: "task", Participation: "hybrid", Capacity: 5, Tags: []string{"项目协作", "资源对接"}, CompletionRules: []string{"goal", "manual"}, Visible: true, Order: 20},
			{Key: "explore_weekend", Name: "周末探索", Description: "适合城市探索和线下打卡", PrimaryCategory: "explore", Participation: "offline", Capacity: 5, Tags: []string{"城市探索", "周末"}, CompletionRules: []string{"time", "capacity"}, Visible: true, Order: 30},
			{Key: "growth_reading", Name: "读书共修", Description: "适合学习、习惯养成和共同复盘", PrimaryCategory: "growth", Participation: "hybrid", Capacity: 5, Tags: []string{"学习", "习惯养成"}, CompletionRules: []string{"time", "goal"}, Visible: true, Order: 40},
		},
		Version: "2026-07-20-create-templates-v1",
	}
}

func (s *Server) currentGameCreateTemplateConfig() gameCreateTemplateConfigDTO {
	categoryConfig := s.currentGameCategoryConfig()
	var stored gameCreateTemplateConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameCreateTemplateConfigKey, &stored) {
		if config, err := normalizeGameCreateTemplateConfig(stored, categoryConfig); err == nil {
			return config
		}
	}
	defaults := defaultGameCreateTemplateConfig()
	if config, err := normalizeGameCreateTemplateConfig(defaults, categoryConfig); err == nil {
		return config
	}
	// 分类已被运营替换、但模板尚未同步维护时，宁可不展示模板，也不能
	// 返回指向无效分类的按钮。
	return gameCreateTemplateConfigDTO{Items: []gameCreateTemplateDTO{}, Version: defaults.Version}
}

func normalizeGameCreateTemplateConfig(req gameCreateTemplateConfigDTO, categoryConfig gameCategoryConfigDTO) (gameCreateTemplateConfigDTO, error) {
	seen := map[string]bool{}
	primaryCategories := map[string]bool{}
	for _, primary := range categoryConfig.PrimaryCategories {
		primaryCategories[strings.TrimSpace(primary.Key)] = true
	}
	items := make([]gameCreateTemplateDTO, 0, len(req.Items))
	for _, item := range req.Items {
		item.Key = strings.TrimSpace(item.Key)
		item.Name = strings.TrimSpace(item.Name)
		item.Description = strings.TrimSpace(item.Description)
		item.PrimaryCategory = strings.TrimSpace(item.PrimaryCategory)
		// 小类别已从一期创建局及后台模板配置中移除；保留字段仅用于
		// 兼容历史局的读取，新增和编辑模板不再写入。
		item.SecondaryCategory = ""
		item.Participation = strings.TrimSpace(item.Participation)
		if item.Key == "" || item.Name == "" || seen[item.Key] {
			return gameCreateTemplateConfigDTO{}, errors.New("模板编码和名称不能为空，且编码不能重复")
		}
		if !primaryCategories[item.PrimaryCategory] {
			return gameCreateTemplateConfigDTO{}, errors.New("模板局类型不存在：" + item.PrimaryCategory)
		}
		if item.Participation != "" && item.Participation != "online" && item.Participation != "offline" && item.Participation != "hybrid" {
			return gameCreateTemplateConfigDTO{}, errors.New("模板参与方式不正确")
		}
		minCapacity := categoryConfig.CreateForm.Capacity.Min
		maxCapacity := categoryConfig.CreateForm.Capacity.Max
		if item.Capacity < 0 || item.Capacity > maxCapacity || (item.Capacity > 0 && item.Capacity < minCapacity) {
			return gameCreateTemplateConfigDTO{}, errors.New("模板人数不在允许范围内")
		}
		item.Tags = cleanStringSlice(item.Tags)
		item.CompletionRules = cleanStringSlice(item.CompletionRules)
		seen[item.Key] = true
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Order < items[j].Order })
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "2026-07-20-create-templates-v1"
	}
	return gameCreateTemplateConfigDTO{Items: items, Version: version}, nil
}

func cleanStringSlice(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func categoryTags(names ...string) []gameCreateFormOptionDTO {
	items := make([]gameCreateFormOptionDTO, 0, len(names))
	for _, name := range names {
		key := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "_"))
		if key != "" {
			items = append(items, gameCreateFormOptionDTO{Key: key, Name: name})
		}
	}
	return items
}

func (s *Server) adminGameCategoryConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameCategoryConfig()})
	case http.MethodPut:
		var req gameCategoryConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "局类型配置格式不正确")
			return
		}
		config, err := normalizeGameCategoryConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.setGameCategoryConfig(config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存局类型配置失败")
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
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) adminGameCreateTemplateConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameCreateTemplateConfig()})
	case http.MethodPut:
		var req gameCreateTemplateConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "局模板配置格式不正确")
			return
		}
		config, err := normalizeGameCreateTemplateConfig(req, s.currentGameCategoryConfig())
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameCreateTemplateConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存局模板配置失败")
			return
		}
		s.recordOperation(r, "game_create_template_config:update", "system_config", gameCreateTemplateConfigKey, map[string]interface{}{
			"templateCount": len(config.Items),
			"version":       config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameCreateTemplateConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) adminGameApplicationConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameApplicationConfig()})
	case http.MethodPut:
		var req gameApplicationConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "入局申请配置格式不正确")
			return
		}
		config, err := normalizeGameApplicationConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameApplicationConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存入局申请配置失败")
			return
		}
		s.recordOperation(r, "game_application_config:update", "system_config", "game_application_config", map[string]interface{}{
			"requireAgreement": config.RequireAgreement,
			"maxUploadCount":   config.MaxUploadCount,
			"version":          config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameApplicationConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) adminGameAuditConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameAuditConfig()})
	case http.MethodPut:
		var req gameAuditConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "组局审核配置格式不正确")
			return
		}
		config, err := normalizeGameAuditConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameAuditConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存组局审核配置失败")
			return
		}
		s.recordOperation(r, "game_audit_config:update", "system_config", "game_audit_config", map[string]interface{}{
			"autoApproveFreeGames":    config.AutoApproveFreeGames,
			"requireManualAuditTypes": config.RequireManualAuditTypes,
			"version":                 config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameAuditConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) adminGameConditionRuleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		httpx.OK(w, map[string]interface{}{"config": s.currentGameConditionRuleConfig()})
	case http.MethodPut:
		var req gameConditionRuleConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "条件规则配置格式不正确")
			return
		}
		config, err := normalizeGameConditionRuleConfig(req)
		if err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, err.Error())
			return
		}
		if err := s.systemConfig.Set(gameConditionRuleConfigKey, config); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "保存条件规则配置失败")
			return
		}
		s.recordOperation(r, "game_condition_rule_config:update", "system_config", "game_condition_rule_config", map[string]interface{}{
			"enabled":       config.Enabled,
			"ruleItemCount": len(config.RuleItems),
			"version":       config.Version,
		})
		httpx.OK(w, map[string]interface{}{"config": s.currentGameConditionRuleConfig()})
	default:
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeValidationError, "请求方式不支持")
	}
}

func (s *Server) currentGameCategoryConfig() gameCategoryConfigDTO {
	var stored gameCategoryConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameCategoryConfigKey, &stored) && len(stored.PrimaryCategories) > 0 {
		config := cloneGameCategoryConfig(upgradeLegacyGameCategoryConfig(stored))
		config.ShareComponent = normalizeGameShareComponentConfig(config.ShareComponent)
		config.EventActions = normalizeGameCardActions(config.EventActions, config.ShareComponent)
		s.applyOperationGameLimits(&config)
		return config
	}
	s.gameCategoryConfigMu.RLock()
	if len(s.gameCategoryConfig.PrimaryCategories) > 0 {
		config := cloneGameCategoryConfig(s.gameCategoryConfig)
		config.ShareComponent = normalizeGameShareComponentConfig(config.ShareComponent)
		config.EventActions = normalizeGameCardActions(config.EventActions, config.ShareComponent)
		s.gameCategoryConfigMu.RUnlock()
		s.applyOperationGameLimits(&config)
		return config
	}
	s.gameCategoryConfigMu.RUnlock()
	config := cloneGameCategoryConfig(defaultGameCategoryConfig())
	config.ShareComponent = normalizeGameShareComponentConfig(config.ShareComponent)
	config.EventActions = normalizeGameCardActions(config.EventActions, config.ShareComponent)
	s.applyOperationGameLimits(&config)
	return config
}

// upgradeLegacyGameCategoryConfig keeps operational settings while converting
// old taxonomy records to the four supported primary categories.
func upgradeLegacyGameCategoryConfig(stored gameCategoryConfigDTO) gameCategoryConfigDTO {
	if hasRequiredPrimaryCategories(stored.PrimaryCategories) {
		for index := range stored.PrimaryCategories {
			stored.PrimaryCategories[index].Children = nil
		}
		stored.TypeFilters = []gameCategoryOptionDTO{{Key: "all", Name: "全部", Visible: true, Order: 0, Selectable: true}}
		stored.DefaultType = "all"
		stored.CreateForm = normalizeGameCreateFormConfig(stored.CreateForm)
		return stored
	}
	defaults := defaultGameCategoryConfig()
	defaults.TypeFilters = []gameCategoryOptionDTO{{Key: "all", Name: "全部", Visible: true, Order: 0, Selectable: true}}
	defaults.DefaultType = "all"
	if len(stored.LocationFilters) > 0 {
		defaults.LocationFilters = stored.LocationFilters
	}
	if len(stored.SortOptions) > 0 {
		defaults.SortOptions = stored.SortOptions
	}
	if len(stored.EventActions) > 0 {
		defaults.EventActions = stored.EventActions
	}
	if len(stored.CreateForm.ParticipationModes) > 0 || len(stored.CreateForm.Tags) > 0 || len(stored.CreateForm.CompletionRules) > 0 {
		defaults.CreateForm = normalizeGameCreateFormConfig(stored.CreateForm)
	}
	return defaults
}

func hasRequiredPrimaryCategories(items []gameCategoryOptionDTO) bool {
	required := map[string]bool{"social": false, "task": false, "explore": false, "growth": false}
	if len(items) != len(required) {
		return false
	}
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		present, ok := required[key]
		if !ok || present || !item.Visible {
			return false
		}
		required[key] = true
	}
	for _, present := range required {
		if !present {
			return false
		}
	}
	return true
}

func (s *Server) applyOperationGameLimits(config *gameCategoryConfigDTO) {
	if config == nil {
		return
	}
	rules := s.currentOperationRules()
	config.CreateForm.Capacity.Min = rules.Game.MinPlayers
	config.CreateForm.Capacity.Max = rules.Game.MaxPlayers
}

func (s *Server) currentGameApplicationConfig() gameApplicationConfigDTO {
	config, err := s.currentGameApplicationConfigStrict()
	if err != nil {
		return s.withApplicationResultSubscribeTemplates(defaultGameApplicationConfig())
	}
	return config
}

// currentGameApplicationConfigStrict is used before accepting an application.
// A storage outage must not silently drop operator-required introductions,
// agreements, or supporting-material checks.
func (s *Server) currentGameApplicationConfigStrict() (gameApplicationConfigDTO, error) {
	var stored gameApplicationConfigDTO
	if s.systemConfig != nil {
		found, err := s.systemConfig.GetStrict(gameApplicationConfigKey, &stored)
		if err != nil {
			return gameApplicationConfigDTO{}, err
		}
		if found && strings.TrimSpace(stored.AgreementTitle) != "" {
			config := mergeGameApplicationConfigDefaults(stored)
			config.RequireRealname = false
			return s.withApplicationResultSubscribeTemplates(config), nil
		}
	}
	return s.withApplicationResultSubscribeTemplates(defaultGameApplicationConfig()), nil
}

// withApplicationResultSubscribeTemplates exposes the active production
// templates needed to ask for subscription permission before a player applies.
func (s *Server) withApplicationResultSubscribeTemplates(config gameApplicationConfigDTO) gameApplicationConfigDTO {
	config.SubscribeTemplateIDs = s.applicationResultSubscribeTemplateIDs()
	return config
}

func (s *Server) applicationResultSubscribeTemplateIDs() []string {
	ids := make([]string, 0, 2)
	for _, template := range s.notices.WechatTemplates() {
		if template.Status != "active" || strings.HasPrefix(template.TemplateID, "mock_") {
			continue
		}
		switch template.Scene {
		case "application_approved", "application_rejected":
			if strings.TrimSpace(template.TemplateID) != "" {
				ids = append(ids, template.TemplateID)
			}
		}
	}
	sort.Strings(ids)
	return ids
}

func (s *Server) applicationResultSubscribeTemplateID(scene string) string {
	for _, template := range s.notices.WechatTemplates() {
		if template.Scene == scene && template.Status == "active" && strings.TrimSpace(template.TemplateID) != "" && !strings.HasPrefix(template.TemplateID, "mock_") {
			return template.TemplateID
		}
	}
	return ""
}

func (s *Server) currentGameAuditConfig() gameAuditConfigDTO {
	var stored gameAuditConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameAuditConfigKey, &stored) && strings.TrimSpace(stored.ApplicationAuditMode) != "" {
		config, err := normalizeGameAuditConfig(stored)
		if err == nil {
			return config
		}
	}
	return defaultGameAuditConfig()
}

func (s *Server) currentGameConditionRuleConfig() gameConditionRuleConfigDTO {
	var stored gameConditionRuleConfigDTO
	if s.systemConfig != nil && s.systemConfig.Get(gameConditionRuleConfigKey, &stored) && len(stored.RuleItems) > 0 {
		config := cloneGameConditionRuleConfig(stored)
		config.PaymentRequired = false
		return config
	}
	return defaultGameConditionRuleConfig()
}

func (s *Server) currentGameCancelConfig() gameCancelConfigDTO {
	// 当前仅开放免费局。旧配置带有赔付、托管资金等商业化规则，不能继续
	// 作用于前台取消页；后续支付能力上线时再恢复后台可配置版本。
	return defaultGameCancelConfig()
}

func (s *Server) currentGameDeliveryPageConfig() gameDeliveryPageConfigDTO {
	// 一期只有免费局。旧配置中的收费、结算文案不能重新下发到小程序，
	// 完成确认统一采用本局完成配置；二期接入支付后再扩展独立模式。
	return defaultGameDeliveryPageConfig()
}

func (s *Server) ensureDefaultSystemConfigs() {
	if s.systemConfig == nil {
		return
	}
	var stored gameCategoryConfigDTO
	if !s.systemConfig.Get(gameCategoryConfigKey, &stored) || len(stored.PrimaryCategories) == 0 {
		_ = s.systemConfig.Set(gameCategoryConfigKey, defaultGameCategoryConfig())
	} else if !hasRequiredPrimaryCategories(stored.PrimaryCategories) {
		_ = s.systemConfig.Set(gameCategoryConfigKey, upgradeLegacyGameCategoryConfig(stored))
	}
	var templateStored gameCreateTemplateConfigDTO
	if !s.systemConfig.Get(gameCreateTemplateConfigKey, &templateStored) {
		_ = s.systemConfig.Set(gameCreateTemplateConfigKey, defaultGameCreateTemplateConfig())
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
	if !s.systemConfig.Get(gameDeliveryPageConfigKey, &deliveryPageStored) || strings.TrimSpace(deliveryPageStored.Free.PageTitle) == "" {
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
	var growthRulesStored reviews.GrowthRules
	if !s.systemConfig.Get(growthRewardRulesConfigKey, &growthRulesStored) {
		_ = s.systemConfig.Set(growthRewardRulesConfigKey, reviews.DefaultGrowthRules())
	}
	var expertSkillDisplayStored expertSkillDisplayConfigDTO
	if !s.systemConfig.Get(expertSkillDisplayConfigKey, &expertSkillDisplayStored) {
		_ = s.systemConfig.Set(expertSkillDisplayConfigKey, defaultExpertSkillDisplayConfig())
	}
	var operationRulesStored operationRulesDTO
	if !s.systemConfig.Get(operationRulesConfigKey, &operationRulesStored) {
		_ = s.systemConfig.Set(operationRulesConfigKey, defaultOperationRules())
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
	// 一期不再维护小类别，历史字段只用于读取旧局。
	config.DefaultSecondaryCategory = ""
	config.DefaultType = strings.TrimSpace(req.DefaultType)
	config.Version = strings.TrimSpace(req.Version)
	if !hasRequiredPrimaryCategories(config.PrimaryCategories) {
		return gameCategoryConfigDTO{}, errors.New("局类型必须保留并展示社交局、任务局、探索局和成长局四类")
	}
	config.TypeFilters = []gameCategoryOptionDTO{{Key: "all", Name: "全部", Visible: true, Order: 0, Selectable: true}}
	if config.DefaultPrimaryCategory == "" {
		config.DefaultPrimaryCategory = config.PrimaryCategories[0].Key
	}
	config.DefaultType = "all"
	config.CreateForm = normalizeGameCreateFormConfig(config.CreateForm)
	config.ShareComponent = normalizeGameShareComponentConfig(config.ShareComponent)
	if config.ShareComponent.Enabled && !config.ShareComponent.EnableInternal && !config.ShareComponent.EnableWechat {
		return gameCategoryConfigDTO{}, errors.New("开启分享时至少选择一种分享渠道")
	}
	config.EventActions = normalizeGameCardActions(config.EventActions, config.ShareComponent)
	if !categoryKeyExists(config.PrimaryCategories, config.DefaultPrimaryCategory) {
		return gameCategoryConfigDTO{}, errors.New("默认局类型不存在")
	}
	if config.Version == "" {
		config.Version = time.Now().UTC().Format("2006-01-02")
	}
	return config, nil
}

func normalizeGameShareComponentConfig(config gameShareComponentConfigDTO) gameShareComponentConfigDTO {
	defaults := defaultGameCategoryConfig().ShareComponent
	// Existing category configurations predate the share component. Treat a
	// missing object as the default rather than silently hiding sharing.
	if config == (gameShareComponentConfigDTO{}) {
		return defaults
	}
	config.Variant = strings.TrimSpace(config.Variant)
	if config.Variant == "" {
		config.Variant = defaults.Variant
	}
	if config.Variant != "channel_sheet" && config.Variant != "icon_button" {
		config.Variant = defaults.Variant
	}
	config.Label = strings.TrimSpace(config.Label)
	if config.Label == "" {
		config.Label = defaults.Label
	}
	return config
}

func normalizeGameCardActions(actions []string, share gameShareComponentConfigDTO) []string {
	result := make([]string, 0, len(actions)+1)
	seenShare := false
	for _, action := range actions {
		action = strings.TrimSpace(action)
		if action == "" || action == "引荐" || action == "邀请（站内）" || action == "站内邀请" {
			continue
		}
		if action == "分享" || action == share.Label {
			if share.Enabled && !seenShare {
				result = append(result, share.Label)
				seenShare = true
			}
			continue
		}
		result = append(result, action)
	}
	if share.Enabled && !seenShare {
		result = append([]string{share.Label}, result...)
	}
	return result
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
		item.Tags = normalizeCreateFormOptions(item.Tags, nil)
		if selectable && !item.Selectable {
			item.Selectable = true
		}
		// 小类别不再是一期的组局维度。历史数据可以读取，但后台配置
		// 和小程序创建表单都不再下发或保存该选项。
		item.Children = nil
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
		result[i].Tags = append([]gameCreateFormOptionDTO(nil), item.Tags...)
	}
	return result
}

func defaultGameApplicationConfig() gameApplicationConfigDTO {
	return gameApplicationConfigDTO{
		AgreementTitle:      "入局申请须知",
		AgreementText:       "申请入局前请确认本人已完成实名，了解局的主题、地点、时间和成员规则。申请通过后请按约参与，临时退出可能影响信用分。",
		RequireRealname:     false,
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
			"subscribeDeclinedText": "审核结果将同步到消息中心",
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
					"actionTip":                "确认后将按对应身份加入本局；局开始后可进入局内群聊。",
					"actionLoadingText":        "处理中...",
					"confirmText":              "确认通过",
				},
				SessionItems: []gameApplicationAuditSessionItemDTO{
					{Key: "topic", Label: "组局主题", IconText: "H", IconClass: "topic"},
					{Key: "time", Label: "时间", IconSrc: "/pages/game/detail/assets/icon-clock.png", IconClass: "time"},
					{Key: "location", Label: "地点", ActionText: "地图位置", IconSrc: "/pages/game/detail/assets/icon-location.png", IconClass: "place"},
				},
				ConfirmRows: []gameApplicationAuditSessionItemDTO{
					{Key: "activityType", Label: "局分类"},
					{Key: "serviceDuration", Label: "活动时长"},
				},
				OptionalActions: []gameApplicationAuditActionDTO{
					{Key: "time", Name: "提议具体时间", IconSrc: "/pages/game/audit-detail/assets/option-time.png"},
				},
				NoticeBullets: []string{
					"确认后请准时参加，如需取消请提前通知",
					"确认后将按对应身份加入本局，局开始后可进行局内沟通",
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
		RequireManualAuditTypes:      []string{"free"},
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
				{"key": "schedule_conflict", "text": "临时有事，无法参加"},
				{"key": "time_or_place", "text": "时间或地点不合适"},
				{"key": "content_mismatch", "text": "本局内容与预期不符"},
				{"key": "other", "text": "其他原因"},
			},
			DefaultReason: "schedule_conflict",
			AgreementText: "我已了解本局取消规则，确认取消将按规则扣减信用分。",
			AgreementItems: []string{
				"我已了解本局为免费局，不涉及资金赔付",
				"我已了解本次取消将按规则扣减信用分",
			},
		},
		Expert: gameCancelRoleConfigDTO{
			ReasonOptions: []map[string]string{
				{"key": "schedule_conflict", "text": "个人时间冲突，无法参加"},
				{"key": "requirement_mismatch", "text": "本局信息与实际情况不符"},
				{"key": "emergency", "text": "身体原因/突发状况"},
				{"key": "other", "text": "其他原因"},
			},
			DefaultReason: "schedule_conflict",
			AgreementText: "我已了解本局取消规则，确认取消将按规则扣减信用分。",
		},
		Version: "2026-07-01",
	}
}

func defaultGameDeliveryPageConfig() gameDeliveryPageConfigDTO {
	completion := deliveryModeConfigDTO{
		PageTitle: "确认本局完成",
		Status:    deliveryStatusConfigDTO{Theme: "free", Title: "本局已结束", Desc: "请确认本局流程和体验已完成"},
		StatePill: deliveryStatePillConfigDTO{Theme: "blue", Text: "待成员确认"},
		Notice: deliveryNoticeConfigDTO{
			IconText: "✓",
			Title:    "完成说明",
			Parts:    []deliveryNoticePartDTO{{Text: "本局为免费局，不涉及资金结算。成员确认完成后可进入评价。"}},
		},
		ConfirmItems: []deliveryConfirmItemDTO{
			{ID: "completed", Title: "本局流程已完成", Desc: "本次组局已按约定完成"},
			{ID: "qualified", Title: "参与体验已确认", Desc: "如有问题，请先在局内沟通"},
			{ID: "communicated", Title: "已通知局内成员", Desc: "成员可继续完成确认和评价"},
		},
		ConfirmNote:       "确认后将通知局内成员完成后续确认；如本局仍有问题，请先在局内沟通。",
		Security:          deliverySecurityConfigDTO{Title: "组局确认", Desc: "确认记录将用于后续评价和信用规则处理"},
		SubmitHints:       deliverySubmitHintsDTO{Ready: "确认后将通知局内成员", Pending: "请完成上方确认项后提交"},
		SubmitToast:       "本局完成确认已提交",
		SubmitLoadingText: "提交中",
		AmountRowLabel:    "",
	}
	return gameDeliveryPageConfigDTO{
		Paid: completion,
		Free: completion,
		QuickActions: []deliveryQuickActionDTO{
			{Key: "contact_player", Title: "联系局内成员", Theme: "blue", IconSrc: "/pages/game/delivery/assets/i18@3x.png"},
			{Key: "contact_guide", Title: "联系领路人", Theme: "orange", IconText: "👬"},
		},
		Version: "2026-07-29",
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
	// 一期普通玩家申请入局不强制实名；实名是行家、领路人身份申请的前置条件。
	config.RequireRealname = false
	if config.AgreementTitle == "" {
		return gameApplicationConfigDTO{}, errors.New("申请协议标题不能为空")
	}
	if config.RequireAgreement && config.AgreementText == "" {
		return gameApplicationConfigDTO{}, errors.New("启用申请协议时，协议内容不能为空")
	}
	if config.MaxUploadCount < 0 || config.MaxUploadCount > 9 {
		return gameApplicationConfigDTO{}, errors.New("最多上传材料数须为 0-9")
	}
	if config.UploadRequired && config.MaxUploadCount == 0 {
		return gameApplicationConfigDTO{}, errors.New("要求上传材料时，最多上传材料数不能为 0")
	}
	if config.MinIntroLength < 0 || config.MaxIntroLength < 0 || config.MinIntroLength > config.MaxIntroLength {
		return gameApplicationConfigDTO{}, errors.New("自我介绍字数范围不正确")
	}
	if config.MaxIntroLength > 500 {
		return gameApplicationConfigDTO{}, errors.New("自我介绍最多不能超过 500 字")
	}
	if config.MaxMessageLength < 0 || config.MaxMessageLength > 500 {
		return gameApplicationConfigDTO{}, errors.New("申请留言最多字数须为 0-500")
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
	// 一期所有局均为免费局，旧收费类别不能通过配置重新启用。
	config.RequireManualAuditTypes = []string{"free"}
	config.ApplicationAuditMode = strings.TrimSpace(config.ApplicationAuditMode)
	config.ReviewerRoles = normalizeStringList(config.ReviewerRoles, 16)
	config.Version = strings.TrimSpace(config.Version)
	if config.BatchAuditMaxCount <= 0 {
		config.BatchAuditMaxCount = 50
	}
	if config.BatchAuditMaxCount > 200 {
		return gameAuditConfigDTO{}, errors.New("批量审核上限不能超过 200 条")
	}
	switch config.ApplicationAuditMode {
	case "", "creator_or_main_guide":
		config.ApplicationAuditMode = "creator_or_main_guide"
	case "creator_only", "main_guide_only", "admin_only":
	default:
		return gameAuditConfigDTO{}, errors.New("入局审核方式不支持")
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
	// 一期没有支付与押金能力，条件局仅使用资格与可见性规则。
	config.PaymentRequired = false
	if len(config.RuleItems) == 0 {
		return gameConditionRuleConfigDTO{}, errors.New("至少保留一条条件规则")
	}
	switch config.DefaultVisibility {
	case "", "approved_users":
		config.DefaultVisibility = "approved_users"
	case "all_users", "invite_only", "hidden":
	default:
		return gameConditionRuleConfigDTO{}, errors.New("默认可见范围不支持")
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
	config.SubscribeTemplateIDs = append([]string(nil), config.SubscribeTemplateIDs...)
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
	// 一期审核页不允许历史配置重新带回资金冻结、收益分配或开局前私聊。
	config.AuditPage.Detail.ConfirmRows = append([]gameApplicationAuditSessionItemDTO(nil), defaults.AuditPage.Detail.ConfirmRows...)
	config.AuditPage.Detail.OptionalActions = append([]gameApplicationAuditActionDTO(nil), defaults.AuditPage.Detail.OptionalActions...)
	config.AuditPage.Detail.Texts["actionTip"] = defaults.AuditPage.Detail.Texts["actionTip"]
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
	if allowed, message := s.canUseCreditAction(userID, "create_game"); !allowed {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, message)
		return
	}
	var req games.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if s.rejectSensitiveGameCreateRequest(w, req) {
		return
	}
	categoryConfig := s.currentGameCategoryConfig()
	operationRules, err := s.currentOperationRulesStrict()
	if err != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取组局规则失败，请稍后重试")
		return
	}
	// 直属领路人护航已从一期范围移除；忽略遗留客户端字段，避免创建后重新开放。
	req.AllowGuideEscort = false
	// 小类别已经下线，不能再让旧客户端或手工请求写回新局。
	req.SecondaryCategory = ""
	req.SecondaryCategoryText = ""
	if strings.TrimSpace(req.Introduction) == "" || strings.TrimSpace(req.Highlights) == "" || strings.TrimSpace(req.Description) == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请填写局介绍、亮点和局描述")
		return
	}
	if req.CoverFileID <= 0 && strings.TrimSpace(req.CoverImage) == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先上传局封面")
		return
	}
	if !games.ValidSignupTimeRange(req.SignupStartAt, req.SignupEndAt, req.StartAt) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "报名时间需完整填写，且报名结束须早于局开始")
		return
	}
	if !hasGameCreateLocation(req) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先选择组局地址")
		return
	}
	if req.MinPlayers < operationRules.Game.MinPlayers || req.MaxPlayers > operationRules.Game.MaxPlayers || req.MinPlayers > req.MaxPlayers {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, fmt.Sprintf("每局人数必须为 %d-%d", operationRules.Game.MinPlayers, operationRules.Game.MaxPlayers))
		return
	}
	if service, ok := s.games.(*games.Service); ok {
		service.SetDailyCreateLimit(operationRules.Game.DailyCreateLimit)
	}
	if strings.TrimSpace(req.PrimaryCategory) != "" && !categoryKeyExists(categoryConfig.PrimaryCategories, strings.TrimSpace(req.PrimaryCategory)) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "局类型无效，请重新选择")
		return
	}
	if req.CoverFileID > 0 {
		file, err := s.files.Get(req.CoverFileID)
		if err != nil || file.UploaderID != userID || file.BizType != "game_cover" {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "封面文件无效，请重新上传")
			return
		}
		download, err := s.files.DownloadURLForFile(file)
		if err != nil {
			if errors.Is(err, files.ErrStorageNotConfigured) {
				httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "文件存储服务暂不可用，请稍后重试")
				return
			}
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "封面文件无效，请重新上传")
			return
		}
		req.CoverImage = download.DownloadURL
	}
	if len(req.DescriptionMedia) > 0 {
		media, mediaErr := s.resolveGameDescriptionMedia(userID, req.DescriptionMedia)
		if mediaErr != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "局详情图片或视频无效，请重新上传")
			return
		}
		req.DescriptionMedia = media
	}
	game, err := s.games.Create(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, games.ErrRealnameRequired):
			httpx.Error(w, http.StatusForbidden, 40341, "强实名未完成")
		case errors.Is(err, games.ErrInvalidGameType):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "小程序一期只能创建免费局")
		case errors.Is(err, games.ErrInvalidPlayers):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, fmt.Sprintf("每局人数必须为 %d-%d", operationRules.Game.MinPlayers, operationRules.Game.MaxPlayers))
		case errors.Is(err, games.ErrInvalidGameInput):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "组局信息不完整或格式不正确，请检查后重试")
		case errors.Is(err, games.ErrDailyLimit):
			httpx.Error(w, http.StatusTooManyRequests, 42921, fmt.Sprintf("每日最多创建 %d 局", operationRules.Game.DailyCreateLimit))
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "创建局失败")
		}
		return
	}
	// SQL-backed creation has already committed the free order in the same
	// transaction. This idempotent call keeps the in-memory test/development
	// implementation aligned and repairs a missing placeholder if needed. A
	// recheck failure must not turn an already committed game into a false 5xx.
	orderPlaceholderReady := true
	if _, err := s.orders.EnsureFreeNoPayOrder(userID, game.ID); err != nil {
		orderPlaceholderReady = false
	}
	s.recordBehavior(userID, "game_create_submit", "game", game.ID, map[string]interface{}{
		"status": game.Status, "gameType": game.GameType, "orderPlaceholderReady": orderPlaceholderReady,
	})
	httpx.OK(w, game)
}

func hasGameCreateLocation(req games.CreateRequest) bool {
	if strings.TrimSpace(req.Address) == "" || (req.Longitude == 0 && req.Latitude == 0) {
		return false
	}
	if math.IsNaN(req.Longitude) || math.IsNaN(req.Latitude) || math.IsInf(req.Longitude, 0) || math.IsInf(req.Latitude, 0) {
		return false
	}
	return req.Longitude >= -180 && req.Longitude <= 180 && req.Latitude >= -90 && req.Latitude <= 90
}

func (s *Server) resolveGameDescriptionMedia(userID int64, items []games.DescriptionMedia) ([]games.DescriptionMedia, error) {
	result := make([]games.DescriptionMedia, 0, len(items))
	imageCount, videoCount := 0, 0
	seen := make(map[int64]bool, len(items))
	for _, item := range items {
		if item.FileID <= 0 || seen[item.FileID] {
			return nil, errors.New("invalid description media")
		}
		seen[item.FileID] = true
		file, err := s.files.Get(item.FileID)
		if err != nil || file.UploaderID != userID || file.BizType != "game_description" {
			return nil, errors.New("invalid description media")
		}
		item.Type = strings.ToLower(strings.TrimSpace(item.Type))
		switch item.Type {
		case "image":
			if !strings.HasPrefix(strings.ToLower(file.MimeType), "image/") {
				return nil, errors.New("invalid description media")
			}
			imageCount++
		case "video":
			if file.MimeType != "video/mp4" && file.MimeType != "video/quicktime" || item.Duration < 0 || item.Duration > 60 {
				return nil, errors.New("invalid description media")
			}
			videoCount++
		default:
			return nil, errors.New("invalid description media")
		}
		download, err := s.files.DownloadURLForFile(file)
		if err != nil {
			return nil, err
		}
		item.URL = download.DownloadURL
		result = append(result, item)
	}
	if imageCount > 9 || videoCount > 1 || len(result) > 10 {
		return nil, errors.New("invalid description media")
	}
	return result, nil
}

func (s *Server) listGames(w http.ResponseWriter, r *http.Request) {
	items := publicGames(s.games.List())
	if userID, ok := s.currentUserID(r); ok {
		s.recordBehavior(userID, "browse_games", "game", 0, map[string]interface{}{"count": len(items)})
	}
	listItems := make([]GameListItemDTO, 0, len(items))
	for _, game := range items {
		listItems = append(listItems, GameListItemDTO{
			Game:           game,
			PlayerAvatars:  s.gameListPlayerAvatars(game),
			ShareComponent: s.currentGameCategoryConfig().ShareComponent,
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
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "身份参数不正确")
		return
	}
	if roleType != "" && roleType != "player" {
		snapshot, err := s.profiles.RoleSnapshotStrict(userID)
		if err != nil {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取角色身份失败，请稍后重试")
			return
		}
		status := snapshot.RoleStatusMap[roleType]
		if status != "active" && status != "approved" {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "当前身份尚未生效")
			return
		}
	}
	allGames, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取首页组局失败，请稍后重试")
		return
	}
	visibleGames := publicGames(allGames)
	pointSummary, _, loaded := s.loadPointsData(w, userID, false)
	if !loaded {
		return
	}
	notificationItems, err := s.notices.ListStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取首页消息提醒失败，请稍后重试")
		return
	}
	onlineCount, err := s.auth.ActiveAppSessionCountStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取在线人数失败，请稍后重试")
		return
	}
	payload, err := s.buildAppHomePayload(userID, visibleGames, roleType, pointSummary, unreadNotificationCount(notificationItems), onlineCount)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取首页数据失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "view_home", "home", 0, map[string]interface{}{"gameCount": len(visibleGames), "roleType": roleType})
	httpx.OK(w, payload)
}

func (s *Server) newbieTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	record, err := s.identity.StatusStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取实名认证状态失败，请稍后重试")
		return
	}
	snapshot, err := s.profiles.RoleSnapshotStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取角色状态失败，请稍后重试")
		return
	}
	stats, err := s.games.StatsForUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局任务状态失败，请稍后重试")
		return
	}
	_, err = s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	reviewIntents, err := s.reviews.MyIntentsStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取评价任务状态失败，请稍后重试")
		return
	}
	applications, err := s.profiles.RoleApplicationsByUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取身份任务状态失败，请稍后重试")
		return
	}
	hasRoleApplication := len(applications) > 0
	hasApprovedRole := hasApprovedNonPlayerRole(snapshot)
	rules := s.currentOperationRules()
	today := time.Now().In(appDisplayLocation)
	dayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, appDisplayLocation)
	dailyJoinedToday, err := s.games.ParticipatedBetweenStrict(userID, dayStart, dayStart.AddDate(0, 0, 1))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取每日任务参与状态失败，请稍后重试")
		return
	}
	completion := map[string]bool{
		"complete_identity":   record.Status == "verified",
		"apply_role":          hasRoleApplication || hasApprovedRole,
		"join_or_create_game": stats.Participated > 0,
		"complete_game":       stats.Completed > 0,
		"submit_review":       len(reviewIntents) > 0,
		"daily_join_game":     dailyJoinedToday,
	}
	items := make([]map[string]interface{}, 0)
	dailyItems := make([]map[string]interface{}, 0)
	activityItems := make([]map[string]interface{}, 0)
	for _, rule := range rules.Tasks.Items {
		if !rule.Enabled {
			continue
		}
		item := map[string]interface{}{"code": rule.Code, "title": rule.Title, "completed": completion[rule.Code], "required": rule.Required, "rewardPoints": rule.RewardPoints, "rewardExperience": rule.RewardExperience, "claimStatus": "available"}
		switch rule.Category {
		case "daily":
			dailyItems = append(dailyItems, item)
		case "activity":
			activityItems = append(activityItems, item)
		default:
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		items = []map[string]interface{}{{"code": "complete_identity", "title": "完成实名认证", "completed": record.Status == "verified"}}
	}
	if len(dailyItems) == 0 {
		current := 0
		if dailyJoinedToday {
			current = 1
		}
		dailyItems = []map[string]interface{}{{"code": "daily_join_game", "title": "今日参与 1 次组局", "completed": dailyJoinedToday, "current": current, "required": 1}}
	}
	if len(activityItems) == 0 {
		activityItems = []map[string]interface{}{{"code": "activity_complete_game", "title": "完成一局并提交评价", "completed": stats.Completed > 0 && len(reviewIntents) > 0, "current": stats.Completed, "required": 1}}
	}
	completed := 0
	for _, item := range items {
		if done, _ := item["completed"].(bool); done {
			completed++
		}
	}
	completedCodes := map[string]bool{}
	dailyCompletedCodes := map[string]bool{}
	if s.tasks != nil {
		completedCodes, err = s.tasks.CompletedCodesStrict(userID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取任务领取状态失败，请稍后重试")
			return
		}
		dailyCompletedCodes, err = s.tasks.CompletedCodesForDateStrict(userID, time.Now())
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取每日任务领取状态失败，请稍后重试")
			return
		}
	}
	for collectionIndex, collection := range [][]map[string]interface{}{items, dailyItems, activityItems} {
		for _, item := range collection {
			code, _ := item["code"].(string)
			isCompleted := completedCodes[code]
			if collectionIndex == 1 {
				isCompleted = dailyCompletedCodes[code]
			}
			if isCompleted {
				item["completed"] = true
				item["claimStatus"] = "claimed"
			}
			if done, _ := item["completed"].(bool); done && s.tasks != nil && !isCompleted {
				var markErr error
				if collectionIndex == 1 {
					_, _, markErr = s.tasks.MarkCompletedForDateOnce(userID, code, time.Now())
				} else {
					_, _, markErr = s.tasks.MarkCompletedOnce(userID, code)
				}
				if markErr == nil {
					item["claimStatus"] = "claimed"
				}
			}
			if done, _ := item["completed"].(bool); done {
				for _, taskRule := range rules.Tasks.Items {
					if taskRule.Code != code || !taskRule.Enabled {
						continue
					}
					rewardKey := taskRule.Code
					if collectionIndex == 1 {
						rewardKey += "@" + today.Format("2006-01-02")
					}
					if _, rewardErr := s.awardTaskReward(userID, taskRule, rewardKey); rewardErr != nil {
						markGrowthPersistenceDegraded(w, "task_center_repair", userID, 0, rewardErr)
					}
					break
				}
			}
		}
	}
	guideState, err := s.newbieGuideStateStrict(userID, items)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取新手引导状态失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{
		"items":     items,
		"completed": completed,
		"total":     len(items),
		"guide":     guideState,
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
		"summary":     gameListSummary(allItems, "我管理的局"),
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
	listedGameIDs := make(map[int64]bool)
	for _, game := range s.games.List() {
		if game.CreatorUserID != userID && game.MainGuideUserID != userID && !s.userGuideForGame(game, userID) && !s.userExpertForGame(game, userID) && s.userPlayerForGame(game, userID) {
			item := s.buildPlayerServiceOrder(userID, game, pageConfig)
			if invitedGameIDs[game.ID] {
				item["category"] = "invited"
			}
			items = append(items, item)
			listedGameIDs[game.ID] = true
			continue
		}
		if invitedGameIDs[game.ID] && game.CreatorUserID != userID && game.MainGuideUserID != userID && !s.userGuideForGame(game, userID) && !s.userExpertForGame(game, userID) {
			item := s.buildPlayerServiceOrder(userID, game, pageConfig)
			item["category"] = "invited"
			items = append(items, item)
			listedGameIDs[game.ID] = true
		}
	}
	// 报名记录在通过前并不是局成员，旧实现会导致“报名审核中”或
	// “报名未通过”在我的局中消失。每个局只展示最新一条申请，且不和
	// 已入局/已受邀的卡片重复。
	latestApplications := make(map[int64]games.Application)
	for _, application := range s.games.ApplicationsForUser(userID) {
		current, exists := latestApplications[application.GameID]
		if !exists || application.CreatedAt.After(current.CreatedAt) || (application.CreatedAt.Equal(current.CreatedAt) && application.ID > current.ID) {
			latestApplications[application.GameID] = application
		}
	}
	for gameID, application := range latestApplications {
		if listedGameIDs[gameID] {
			continue
		}
		// 审核通过后应由成员关系承载；这里只补足尚未入局的待审、驳回和
		// 主动取消记录，避免行家/领路人角色页面被旧申请重复占位。
		if application.Status == "approved" {
			continue
		}
		game, err := s.games.Get(gameID)
		if err != nil {
			continue
		}
		items = append(items, s.buildPlayerApplicationOrder(game, application))
	}
	allItems := items
	items, page, pageSize, total := paginateRoleItems(r, allItems, "statusType")
	httpx.OK(w, map[string]interface{}{
		"items":       items,
		"orders":      items,
		"summary":     gameListSummary(allItems, "我参与的局"),
		"pageConfig":  pageConfig,
		"serverTime":  time.Now().Format(time.RFC3339),
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
		"hasPrevious": page > 1,
		"hasMore":     page*pageSize < total,
	})
}

func (s *Server) buildPlayerApplicationOrder(game games.Game, application games.Application) map[string]interface{} {
	statusType := "active"
	if application.Status == "rejected" || application.Status == "cancelled" {
		statusType = "canceled"
	}
	statusText := games.ApplicationStatusText(application.Status)
	return map[string]interface{}{
		"id":                      "APPLICATION-" + strconv.FormatInt(application.ID, 10),
		"game":                    game,
		"gameId":                  game.ID,
		"serviceOrderId":          "APPLICATION-" + strconv.FormatInt(application.ID, 10),
		"ref":                     "APPLICATION-" + strconv.FormatInt(application.ID, 10),
		"title":                   game.Title,
		"serviceTitle":            game.Title,
		"category":                "joined",
		"statusType":              statusType,
		"status":                  application.Status,
		"statusText":              statusText,
		"applicationStatus":       application.Status,
		"applicationRejectReason": application.RejectReason,
		"reason":                  application.RejectReason,
		"reasonLabel":             "报名未通过原因",
		"resultText":              statusText,
		"createdAt":               application.CreatedAt.Format(time.RFC3339),
		"categoryText":            homeGameCategoryText(game),
		"startAt":                 game.StartAt,
		"addressText":             game.Address,
		"memberText":              gameMemberCountText(game),
		// 该构建器也被轻量单元测试直接调用；此时 Server 未初始化用户资料服务，
		// 因而使用稳定的角色文案，详情页仍可展示真实发起人资料。
		"creatorName":    "发起人",
		"viewerRoleText": "报名用户",
		"actions": map[string]interface{}{
			"canContactExpert": false,
			"canCancelOrder":   application.Status == "pending",
		},
	}
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
	playerName := s.inGameDisplayName(targetID, "\u73a9\u5bb6")
	gameIDText := strconv.FormatInt(game.ID, 10)
	serviceOrderID := serviceOrderID(game.ID)
	reviewRoute := "/pages/game/review/index?gameId=" + gameIDText
	deliveryRoute := "/pages/game/delivery/index?gameId=" + gameIDText + "&mode=free&note=%E7%A1%AE%E8%AE%A4%E6%9C%AC%E5%B1%80%E5%B7%B2%E5%AE%8C%E6%88%90"
	completionCanConfirm, _ := s.gameCompletionState(game, userID)["canConfirm"].(bool)
	completedActionText := serviceReviewActionText(reviewed, canReview, statusType, game.Status)
	completedActionRoute := reviewRoute
	completedActionEnabled := canReview
	if completionCanConfirm {
		completedActionText = "确认完成"
		completedActionRoute = deliveryRoute
		completedActionEnabled = true
	}
	// 取消页只接收局 ID 并从后端读取本局状态，不能再通过路由传入金额、
	// 赔付比例等二期字段，避免旧页面参数被当作真实结算数据展示。
	expertCancelRoute := "/pages/game/expert-cancel/index?gameId=" + gameIDText
	playerContactRoute := ""
	guideContactRoute := ""
	if game.Status == games.StatusInProgress {
		playerContactRoute = "/pages/im/room/index?gameId=" + gameIDText + "&prefill=%E4%BD%A0%E5%A5%BD%EF%BC%8C%E6%88%91%E6%83%B3%E7%A1%AE%E8%AE%A4%E4%B8%80%E4%B8%8B%E6%9C%AC%E5%B1%80%E8%BF%9B%E5%BA%A6%E3%80%82"
		guideContactRoute = "/pages/im/room/index?gameId=" + gameIDText + "&prefill=%E4%BD%A0%E5%A5%BD%EF%BC%8C%E8%AF%B7%E5%90%8C%E6%AD%A5%E4%B8%80%E4%B8%8B%E6%9C%AC%E5%B1%80%E8%BF%9B%E5%BA%A6%E3%80%82"
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
		"categoryText":        homeGameCategoryText(game),
		"startAt":             game.StartAt,
		"addressText":         game.Address,
		"memberText":          gameMemberCountText(game),
		"creatorName":         s.inGameDisplayName(game.CreatorUserID, "发起人"),
		"viewerRoleText":      managedGameViewerRoleText(game, userID, s.games.MemberRoles(game.ID)),
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
		"primaryActionText":   "确认本局完成",
		"secondaryActionText": "申请取消",
		"playerActionText":    "联系局内玩家",
		"guideActionText":     "联系局内领路人",
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
			"canContactPlayer":          targetID > 0 && playerContactRoute != "",
			"contactPlayerRoute":        playerContactRoute,
			"canContactGuide":           game.MainGuideUserID > 0 && guideContactRoute != "",
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
	applicationStatus := ""
	applicationStatusText := ""
	if application, ok := s.latestApplicationForGameUser(userID, game.ID); ok {
		applicationStatus = application.Status
		applicationStatusText = games.ApplicationStatusText(application.Status)
		if application.Status == "approved" && (game.Status == games.StatusRecruiting || game.Status == games.StatusFull) {
			statusText = applicationStatusText
		}
	}
	canReview := hasReviewTodo && statusType == "complete" && targetID > 0
	reviewed := hasSubmittedReview && !hasReviewTodo
	hasExpert := expertID > 0
	expertName := "暂未分配行家"
	if hasExpert {
		expertName = s.inGameDisplayName(expertID, "\u884c\u5bb6")
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
	// 同上：取消规则、信用扣减和当前局状态以详情接口为准。
	playerCancelRoute := "/pages/game/player-cancel/index?gameId=" + gameIDText
	contactExpertRoute := ""
	if hasExpert && game.Status == games.StatusInProgress {
		contactExpertRoute = "/pages/im/room/index?gameId=" + gameIDText + "&prefill=" + url.QueryEscape("你好，我想确认一下本次组局进度。")
	}
	primaryActionText := "查看局内沟通"
	noticeText := "本局开始后可在局内群聊沟通。"
	return map[string]interface{}{
		"id":                    serviceOrderID,
		"game":                  game,
		"gameId":                game.ID,
		"serviceOrderId":        serviceOrderID,
		"ref":                   serviceOrderID,
		"title":                 game.Title,
		"serviceTitle":          game.Title,
		"statusType":            statusType,
		"statusText":            statusText,
		"applicationStatus":     applicationStatus,
		"applicationStatusText": applicationStatusText,
		"categoryText":          homeGameCategoryText(game),
		"startAt":               game.StartAt,
		"addressText":           game.Address,
		"memberText":            gameMemberCountText(game),
		"creatorName":           s.inGameDisplayName(game.CreatorUserID, "发起人"),
		"viewerRoleText":        "参与者",
		"expertId":              expertID,
		"expertName":            expertName,
		"name":                  expertName,
		"hasExpert":             hasExpert,
		"expert":                serviceOrderPerson(expertID, expertName),
		"guideId":               game.MainGuideUserID,
		"guide":                 serviceOrderPerson(game.MainGuideUserID, s.inGameDisplayName(game.MainGuideUserID, "\u9886\u8def\u4eba")),
		"startedAt":             game.CreatedAt.Format(time.RFC3339),
		"expectedDeliveryAt":    game.CreatedAt.Add(72 * time.Hour).Format(time.RFC3339),
		"serverTime":            time.Now().Format(time.RFC3339),
		"reviewStatus":          serviceReviewStatus(reviewed, canReview, statusType, game.Status),
		"reviewed":              reviewed,
		"canReview":             canReview,
		"canReviewBoth":         canReview,
		"reviewActionText":      completedActionText,
		"canReviewAction":       completedActionEnabled,
		"primaryActionText":     primaryActionText,
		"secondaryActionText":   myGamesPageConfigText(pageConfig, "cancelOrderText", "申请取消"),
		"noticeText":            noticeText,
		"contactExpertRoute":    contactExpertRoute,
		"playerCancelRoute":     playerCancelRoute,
		"reviewRoute":           completedActionRoute,
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

func (s *Server) latestApplicationForGameUser(userID int64, gameID int64) (games.Application, bool) {
	var latest games.Application
	hasLatest := false
	for _, application := range s.games.ApplicationsForUser(userID) {
		if application.GameID != gameID {
			continue
		}
		if !hasLatest || application.CreatedAt.After(latest.CreatedAt) || (application.CreatedAt.Equal(latest.CreatedAt) && application.ID > latest.ID) {
			latest = application
			hasLatest = true
		}
	}
	return latest, hasLatest
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

// gameListSummary is the first-phase summary for "我的局". 组局目前没有
// 收费、分润或结算入口，因此此处只返回局数量，避免旧的服务订单字段被
// 新页面误用成资金数据。
func gameListSummary(items []map[string]interface{}, title string) map[string]interface{} {
	active := 0
	completed := 0
	canceled := 0
	disputed := 0
	for _, item := range items {
		switch item["statusType"] {
		case "active":
			active++
		case "complete":
			completed++
		case "canceled":
			canceled++
		case "dispute":
			disputed++
		}
	}
	return map[string]interface{}{
		"title":         title,
		"amountText":    strconv.Itoa(len(items)) + " 局",
		"activeCount":   active,
		"pendingCount":  completed,
		"completeCount": completed,
		"disputeCount":  disputed + canceled,
		"canceledCount": canceled,
	}
}

func gameMemberCountText(game games.Game) string {
	if game.MaxPlayers <= 0 {
		return strconv.Itoa(game.CurrentPlayers) + " 人已加入"
	}
	return strconv.Itoa(game.CurrentPlayers) + "/" + strconv.Itoa(game.MaxPlayers) + " 人"
}

func managedGameViewerRoleText(game games.Game, userID int64, memberRoles []games.MemberRole) string {
	if game.CreatorUserID == userID {
		return "发起人"
	}
	return collaborationRoleText(game, userID, gameMemberRole(game, userID, true, gameMemberRoleMap(memberRoles)))
}

func publicGames(items []games.Game) []games.Game {
	result := make([]games.Game, 0, len(items))
	now := time.Now()
	for _, game := range items {
		if isPublicJoinableGame(game, now) {
			result = append(result, game)
		}
	}
	return result
}

func isPublicGameStatus(status string) bool {
	return status == "recruiting" || status == "full" || status == "in_progress"
}

func isPublicJoinableGame(game games.Game, now time.Time) bool {
	if !isPublicGameStatus(game.Status) {
		return false
	}
	// 满员只停止继续报名，仍须由发起人手动开局。满员和进行中的局都只在当天保留为
	// 状态展示卡片，次日从首页移除，且不再允许报名。
	if game.Status == "full" {
		return sameAppDay(game.CreatedAt, now)
	}
	if game.Status == "in_progress" {
		if game.CurrentPlayers < game.MaxPlayers || game.MaxPlayers <= 0 {
			return false
		}
		startedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(game.StartedAt))
		return err == nil && sameAppDay(startedAt, now)
	}
	if !games.CanApplyWithinSignupWindow(game, now) {
		return false
	}
	return game.MaxPlayers <= 0 || game.CurrentPlayers < game.MaxPlayers
}

func sameAppDay(left time.Time, right time.Time) bool {
	if left.IsZero() || right.IsZero() {
		return false
	}
	return left.In(appDisplayLocation).Format("2006-01-02") == right.In(appDisplayLocation).Format("2006-01-02")
}

func (s *Server) adminGames(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	gameType := strings.TrimSpace(r.URL.Query().Get("gameType"))
	primaryCategory := strings.TrimSpace(r.URL.Query().Get("primaryCategory"))
	cityCode := strings.TrimSpace(r.URL.Query().Get("cityCode"))
	keyword := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("keyword")))
	allItems, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局管理列表失败，请稍后重试")
		return
	}
	items := make([]games.Game, 0, len(allItems))
	for _, game := range allItems {
		if status != "" && game.Status != status {
			continue
		}
		if gameType != "" && game.GameType != gameType {
			continue
		}
		if primaryCategory != "" && game.PrimaryCategory != primaryCategory {
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
		"items":           items,
		"total":           len(items),
		"status":          status,
		"gameType":        gameType,
		"primaryCategory": primaryCategory,
		"cityCode":        cityCode,
		"keyword":         keyword,
	})
}

func (s *Server) adminGameApplications(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	gameID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("gameId")), 10, 64)
	userID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("userId")), 10, 64)
	applications := make([]games.Application, 0)
	seen := make(map[int64]bool)
	allGames, err := s.games.ListStrict()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局申请列表失败，请稍后重试")
		return
	}
	seenCreators := make(map[int64]bool)
	for _, game := range allGames {
		if gameID > 0 && game.ID != gameID {
			continue
		}
		if seenCreators[game.CreatorUserID] {
			continue
		}
		seenCreators[game.CreatorUserID] = true
		creatorApplications, listErr := s.games.ApplicationsForCreatorStrict(game.CreatorUserID)
		if listErr != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取组局申请列表失败，请稍后重试")
			return
		}
		for _, item := range creatorApplications {
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	if s.rejectSensitiveGameCreateRequest(w, req) {
		return
	}
	// 一期后台与小程序保持同一规则：只可创建免费局。
	// 忽略历史管理端或手工请求提交的商业化字段。
	req.GameType = "free"
	req.Type = "free"
	req.Price = 0
	req.ProfitTemplate = ""
	req.DistributionMethod = "none"
	req.PaymentStatus = "not_required"
	if !hasGameCreateLocation(req) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先选择组局地址")
		return
	}
	game, err := s.games.CreateFromAdmin(req)
	if err != nil {
		switch {
		case errors.Is(err, games.ErrInvalidGameType):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "局类型不支持")
		case errors.Is(err, games.ErrInvalidPlayers):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "每局人数必须为 5-8 人")
		case errors.Is(err, games.ErrInvalidGameInput):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "组局信息填写不完整或格式不正确")
		case errors.Is(err, games.ErrRealnameRequired):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "主领路人须先完成实名认证")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "创建组局失败，请稍后重试")
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
			UserID:          memberID,
			Name:            name,
			Nickname:        nickname,
			DisplayName:     name,
			AvatarURL:       s.imUserAvatarURL(memberID),
			AvatarText:      profile.AvatarText,
			Role:            role,
			RoleLabel:       roleLabel,
			Position:        roleLabel,
			Topic:           "参与本次组局",
			PrimaryTag:      primaryTag,
			Location:        game.CityName,
			IsCreator:       isCreator,
			IsCurrentUser:   isCurrentUser,
			Confirmed:       confirmed[memberID],
			ExpertBlueBadge: s.expertBlueBadgeForUser(memberID),
		}
		if includeRealName {
			member.RealName = profile.RealName
		}
		members = append(members, member)
	}
	return confirmed, members
}

func (s *Server) buildGameDetail(userID int64, game games.Game) GameDetailDTO {
	game.AllowedRoles = games.AllowedRolesForGame(game)
	memberIDs := s.games.Members(game.ID)
	_, members := s.buildGameMembers(userID, game, false)
	relation := s.buildGameRelation(userID, game)
	isFavorited, favoriteCount := s.gameFavoriteState(userID, game.ID)
	detail := GameDetailDTO{
		Game:           game,
		MemberIDs:      memberIDs,
		Members:        members,
		MyRelation:     relation,
		IsFavorited:    isFavorited,
		FavoriteCount:  favoriteCount,
		ShareComponent: s.currentGameCategoryConfig().ShareComponent,
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
			UserID:          game.CreatorUserID,
			Name:            s.displayName(game.CreatorUserID, "玩家"),
			AvatarText:      organizerIdentity.AvatarText,
			AvatarURL:       s.imUserAvatarURL(game.CreatorUserID),
			Role:            organizerRole,
			RoleLabel:       organizerRoleLabel,
			Rating:          rating,
			RatingCount:     ratingCount,
			ExpertBlueBadge: s.expertBlueBadgeForUser(game.CreatorUserID),
		},
		PrimaryAction: gameDetailPrimaryAction(game, relation, pendingCount, reviewed),
	}
}

func gameDetailStatusText(status string) string {
	return games.StatusText(status)
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
		return GameDetailPrimaryActionDTO{Text: text, Disabled: true, Action: "none", DisabledReason: text}
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
		if game.Status == "recruiting" && relation.CanAudit && pendingCount > 0 {
			return action("审核报名（"+strconv.Itoa(pendingCount)+"）", "audit", "/pages/game/audit/index?gameId="+gameID)
		}
		if game.Status == "recruiting" && relation.ApplicationStatus == "pending" {
			return disabled(games.ApplicationStatusText(relation.ApplicationStatus))
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
				return disabled("已入局")
			}
			return disabled("已满员")
		}
		if relation.IsMember {
			return disabled("已入局")
		}
		if relation.IsCreator {
			return disabled("等待报名")
		}
		return disabled("等待开局")
	case "in_progress":
		if relation.IsMember {
			return action("进入组局", "collaboration", "/pages/game/collaboration/index?gameId="+gameID)
		}
		if relation.ApplyDisabledReason != "" {
			return disabled(relation.ApplyDisabledReason)
		}
		return disabled("进行中")
	case "pending_confirm":
		if relation.CanConfirm {
			return action("确认完成", "confirm", "/pages/game/delivery/index?gameId="+gameID+"&mode=free")
		}
		if relation.ApplyDisabledReason != "" {
			return disabled(relation.ApplyDisabledReason)
		}
		return disabled("已结束")
	case "pending_review":
		if relation.CanReview {
			return action("去评价", "review", "/pages/game/review/index?gameId="+gameID)
		}
		if reviewed {
			return disabled("已评价")
		}
		if relation.ApplyDisabledReason != "" {
			return disabled(relation.ApplyDisabledReason)
		}
		return disabled("待评价")
	case "completed":
		if relation.ApplyDisabledReason != "" {
			return disabled(relation.ApplyDisabledReason)
		}
		return disabled("已完成")
	case "canceled", "cancelled":
		if relation.ApplyDisabledReason != "" {
			return disabled(relation.ApplyDisabledReason)
		}
		return disabled("本局已取消")
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅局内成员可查看本局完成详情")
		return
	}
	if updatedGame, resolved, err := s.games.ResolveNoExpertPendingConfirm(game.ID); err != nil {
		writeGameError(w, err)
		return
	} else if resolved {
		game = updatedGame
		if !currentGameReviewable(s.games, game.ID) {
			if reviewErr := s.reviews.MarkGameReviewableStrict(game.ID); reviewErr != nil {
				markReviewPersistenceDegraded(w, "resolve_pending_confirm", game.ID, reviewErr)
			}
			if growthErr := s.awardCompletedGameRewards(game.ID); growthErr != nil {
				markGrowthPersistenceDegraded(w, "resolve_pending_confirm", userID, game.ID, growthErr)
			}
			if connectionErr := s.createCoGameConnections(game.ID); connectionErr != nil {
				markConnectionPersistenceDegraded(w, "resolve_pending_confirm", userID, 0, connectionErr)
			}
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
	// 一期所有前台组局均按免费局处理；支付、资金和结算字段只作为二期
	// 架构预留，不能影响完成确认页面。
	fund := map[string]interface{}{"status": "free_no_pay"}
	deliveryMode := "free"
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
			"emptyText":               "可上传本局完成截图、活动照片等凭证，最多 4 张。",
			"fullText":                "\u6700\u591a\u4e0a\u4f20 4 \u5f20\u51ed\u8bc1",
			"selectedTemplate":        "\u5df2\u9009\u62e9 {selected}/{max} \u5f20\u51ed\u8bc1",
			"uploadActionText":        "\u4e0a\u4f20\u51ed\u8bc1",
			"contactPlayerText":       "\u8054\u7cfb\u73a9\u5bb6",
			"contactExpertText":       "联系行家",
			"contactGuideText":        "\u8054\u7cfb\u9886\u8def\u4eba",
			"cancelServiceText":       "申请取消本局",
			"unavailableTextTemplate": "{action}\u6682\u4e0d\u53ef\u7528",
			"contactPlayerPrefill":    "你好，麻烦确认一下本局完成情况。",
			"contactExpertPrefill":    "你好，想和你确认一下本局完成情况。",
			"contactGuidePrefill":     "你好，辛苦同步一下本局完成状态。",
			"primaryContactKey":       contactKey,
			"primaryContactText":      contactText,
			"primaryContactPrefill":   contactPrefill,
			"timelinePrefill":         "你好，本局已经结束，麻烦确认一下。",
			"idleTimelineText":        "\u5f53\u524d\u8282\u70b9\u65e0\u9700\u989d\u5916\u64cd\u4f5c",
			"timelineActionText":      "\u8054\u7cfb\u73a9\u5bb6",
			"proofNoteTemplate":       "完成凭证 ID：{fileIds}",
		},
	})
}

func deliveryActivityRows(game games.Game) []map[string]interface{} {
	rows := []map[string]interface{}{
		{"label": "局分类", "value": homeGameCategoryText(game), "type": "blue"},
	}
	if duration := deliveryDurationText(game); duration != "" {
		rows = append(rows, map[string]interface{}{"label": "活动时长", "value": duration})
	}
	if startText := deliveryTimeText(game.StartAt, game.CreatedAt); startText != "" {
		rows = append(rows, map[string]interface{}{"label": "开始时间", "value": startText})
	}
	if endText := deliveryTimeText(game.EndAt, time.Time{}); endText != "" {
		rows = append(rows, map[string]interface{}{"label": "结束时间", "value": endText})
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

	expertDesc := "等待行家确认本局完成"
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
		expertDesc = "行家已确认本局完成 " + confirmed[expertID].CreatedAt.Format("01-02 15:04")
		expertState = "done"
		expertLineState = "confirmed"
	}

	playerDesc := "需玩家确认本局体验已完成"
	playerTitle := "等待玩家确认"
	playerState := "pending"
	playerLineState := "pending"
	if playerConfirmed {
		playerTitle = "玩家已确认完成"
		playerDesc = s.inGameDisplayName(playerID, "玩家") + "已确认本局体验完成"
		playerState = "done"
		playerLineState = "confirmed"
	} else if expertConfirmed {
		playerState = "active"
	}

	archiveDesc := "成员确认后本局自动归档"
	archiveState := "pending"
	if bothConfirmed {
		archiveDesc = "本局已进入归档"
		archiveState = "active"
	}

	return []map[string]interface{}{
		{"index": 1, "key": "expert_confirmed", "title": expertTitle, "desc": expertDesc, "state": expertState, "lineState": expertLineState},
		{"index": 2, "key": "player_confirmed", "title": playerTitle, "desc": playerDesc, "state": playerState, "lineState": playerLineState, "action": "contact_player"},
		{"index": 3, "key": "game_archive", "title": "本局归档", "desc": archiveDesc, "state": archiveState},
	}
}

func deliveryContactForRole(role string) (string, string, string) {
	if role == "member" || role == "player" {
		return "contact_expert", "联系行家", "你好，想和你确认一下本局完成情况。"
	}
	return "contact_player", "联系玩家", "你好，麻烦确认一下本局完成情况。"
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
		waitingText = "请等待行家先确认本局完成"
	case role == "guide" || role == "main_guide":
		waitingText = "领路人无需确认，等待行家和玩家完成确认"
	case role == "guide_escort":
		waitingText = "护航领路人无需交付确认，可保持观察"
	case canConfirm:
		waitingText = "请确认本局已经完成"
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅局内成员可查看局协作进度")
		return
	}
	messages, _ := s.im.Messages(userID, game.ID)
	feedbacks, _ := s.games.ProgressFeedbacks(userID, game.ID)
	milestones, milestoneErr := s.games.Milestones(userID, game.ID)
	if milestoneErr != nil {
		writeGameError(w, milestoneErr)
		return
	}
	checkins, checkinErr := s.games.Checkins(userID, game.ID)
	if checkinErr != nil {
		writeGameError(w, checkinErr)
		return
	}
	now := time.Now()
	httpx.OK(w, map[string]interface{}{
		"gameId":        game.ID,
		"title":         game.Title,
		"status":        game.Status,
		"statusText":    homeGameStatusText(game.Status),
		"dayText":       collaborationDayText(game.CreatedAt, now),
		"progress":      collaborationProgressWithFeedback(game, feedbacks),
		"milestones":    milestones,
		"checkins":      s.collaborationCheckins(checkins),
		"members":       s.collaborationMembers(userID, game),
		"membersText":   s.collaborationMembersText(game),
		"messages":      s.collaborationMessages(messages),
		"actions":       s.collaborationActions(game, userID),
		"currentUserId": userID,
		"serverTime":    now.Format(time.RFC3339),
	})
}

// collaborationCheckins supplies the small amount of identity and time context
// needed by the member-facing collaboration timeline. Raw check-in records are
// still available from the dedicated endpoint and the administrative audit view.
func (s *Server) collaborationCheckins(checkins []games.Checkin) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(checkins))
	for _, checkin := range checkins {
		items = append(items, map[string]interface{}{
			"id":            checkin.ID,
			"userId":        checkin.UserID,
			"milestoneId":   checkin.MilestoneID,
			"checkinType":   checkin.CheckinType,
			"content":       checkin.Content,
			"fileIds":       checkin.FileIDs,
			"status":        checkin.Status,
			"createdAt":     checkin.CreatedAt,
			"userName":      s.inGameDisplayName(checkin.UserID, "成员"),
			"createdAtText": checkin.CreatedAt.Local().Format("01-02 15:04"),
		})
	}
	return items
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
	if _, imErr := s.im.ReadOnlyRoomsByGameIDsStrict([]int64{gameID}, "game_ended"); imErr != nil {
		markIMPersistenceDegraded(w, "request_completion", gameID, imErr)
	}
	directReview := game.Status == "pending_review"
	if directReview && !wasReviewable {
		if reviewErr := s.reviews.MarkGameReviewableStrict(gameID); reviewErr != nil {
			markReviewPersistenceDegraded(w, "request_completion", gameID, reviewErr)
		}
		if growthErr := s.awardCompletedGameRewards(gameID); growthErr != nil {
			markGrowthPersistenceDegraded(w, "request_completion", userID, gameID, growthErr)
		}
		if connectionErr := s.createCoGameConnections(gameID); connectionErr != nil {
			markConnectionPersistenceDegraded(w, "request_completion", userID, 0, connectionErr)
		}
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
			_, _ = s.createCriticalNotification(w, "request_game_completion", notifications.CreateRequest{
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
			s.createReviewRemindNotifications(w, gameID)
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
	if game.Status == "cancelled" || game.Status == "canceled" {
		return map[string]interface{}{
			"percent": 0,
			"title":   "组局已取消",
			"tasks": []map[string]interface{}{
				{"key": "group_success", "title": "组局已取消", "desc": "本局已取消，协作流程已结束", "state": "cancelled"},
				{"key": "service_active", "title": "已停止", "desc": "不会再产生新的进度反馈", "state": "cancelled"},
				{"key": "service_done", "title": "已结束", "desc": "组局取消后无需继续确认或评价", "state": "cancelled"},
			},
		}
	}
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

func collaborationProgressWithFeedback(game games.Game, feedbacks []games.ProgressFeedback) map[string]interface{} {
	progress := collaborationProgress(game)
	if len(feedbacks) == 0 || (game.Status != "in_progress" && game.Status != "pending_confirm") {
		return progress
	}
	latest := feedbacks[len(feedbacks)-1]
	if latest.Progress < 0 || latest.Progress > 100 {
		return progress
	}
	progress["percent"] = latest.Progress
	progress["title"] = "进度 " + strconv.Itoa(latest.Progress) + "%"
	return progress
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

func (s *Server) collaborationActions(game games.Game, userID int64) map[string]interface{} {
	memberRoles := gameMemberRoleMap(s.games.MemberRoles(game.ID))
	canManage := game.CreatorUserID == userID || game.MainGuideUserID == userID
	canManageProgress := canManage || memberRoles[userID] == "expert"
	canEnd := canManage && game.Status == "in_progress"
	canCheckin := s.games.IsMember(game.ID, userID) && game.Status == "in_progress"
	return map[string]interface{}{
		"canManageMembers":    canManage,
		"canManageProgress":   canManageProgress,
		"canManageMilestones": canManageProgress,
		"canCheckin":          canCheckin,
		"canEndGame":          canEnd,
		"completionMode":      map[bool]string{true: "direct_review", false: "ordered_confirm"}[game.GameSource == "admin"],
		"manageRoute":         "pages/game/participants/index?gameId=" + strconv.FormatInt(game.ID, 10),
		"endConfirmRoute":     "pages/game/collaboration/index?gameId=" + strconv.FormatInt(game.ID, 10),
		"reviewRoute":         "pages/game/review/index?gameId=" + strconv.FormatInt(game.ID, 10),
	}
}

func gameRoleLabel(role string) string {
	switch role {
	case "expert":
		return "\u884c\u5bb6"
	case "guide", "main_guide":
		return "\u9886\u8def\u4eba"
	case "guide_escort":
		return "\u62a4\u822a\u9886\u8def\u4eba"
	default:
		return "\u73a9\u5bb6"
	}
}

func gameRoleAvatarClass(role string) string {
	switch role {
	case "guide", "main_guide", "guide_escort":
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
		return "待评价"
	case "in_progress", "pending_confirm":
		return "进行中"
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
		"groupSectionTitle": "局内成员",
		"onlineText":        "局内",
		"chatButtonText":    "进入局内群聊",
		"activityTitle":     "本局信息",
		"fundTitle":         "",
		"fundEmptyText":     "",
		"stepsTitle":        "下一步",
		"stepsEmptyText":    "暂无下一步动作",
		"manageButtonText":  "进入局管理",
		"loadFailedText":    "组局成功详情加载失败，请稍后重试",
		"chatPrefill":       "你好，想和大家确认一下本局安排。",
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
		"defaultImPrefill":  "我来跟进一下本局成员反馈。",
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权查看该局的领路人跟进信息")
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
			{"title": "组局成功！", "desc": "成员已确认参加，等待本局按计划开始", "time": game.CreatedAt.Format("01-02 15:04"), "active": true},
		},
		"party": map[string]interface{}{
			"confirmedText": "",
			"cardClass":     "success-guide-party-card",
			"cardStyle":     "width: 684rpx; height: 446rpx; margin: 32rpx auto 0; box-shadow: none;",
			"titleClass":    "regular",
			"player":        guideSuccessPartyMember(s.inGameDisplayName(playerID, "\u73a9\u5bb6"), playerID, "\u73a9\u5bb6", "pink"),
			"expert":        guideSuccessPartyMember(s.inGameDisplayName(expertID, "\u884c\u5bb6"), expertID, "\u884c\u5bb6", "blue"),
		},
		"reward": map[string]interface{}{"show": false},
		"followUps": []map[string]interface{}{
			{"key": "schedule", "title": "查看局详情", "desc": "查看本局时间、地点和协作进度", "iconSrc": "https://static.haowan.net.cn/miniprogram/pages/game/success-guide/assets/follow-schedule.png", "iconClass": "schedule", "theme": "blue", "target": map[string]interface{}{"type": "game_detail", "gameId": game.ID}},
			{"key": "feedback", "title": "查看局内沟通", "desc": "本局开始后可进入局内群聊跟进情况", "iconSrc": "https://static.haowan.net.cn/miniprogram/pages/game/success-guide/assets/follow-feedback.png", "iconClass": "feedback", "theme": "purple", "target": map[string]interface{}{"type": "im_room", "gameId": game.ID, "roomId": roomID, "prefill": "我来跟进一下本局成员反馈。"}},
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅发起人或主领路人可执行此操作")
		return
	}
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请选择跟进操作")
		return
	}
	if !validGuideFollowUpAction(action) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "跟进操作不支持")
		return
	}
	target := guideFollowUpTarget(action, game.ID)
	note := strings.TrimSpace(req.Note)
	s.recordBehavior(userID, "guide_follow_up_"+action, "game", game.ID, map[string]interface{}{"action": action, "note": note})
	s.createGuideFollowUpNotifications(w, userID, game, action)
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

func (s *Server) createGuideFollowUpNotifications(w http.ResponseWriter, userID int64, game games.Game, action string) {
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
		_, _ = s.createCriticalNotification(w, "guide_follow_up", notifications.CreateRequest{
			UserID:     memberID,
			NotifyType: "guide_follow_up",
			Title:      title,
			Content:    content,
			BizType:    "game",
			BizID:      game.ID,
		})
	}
}

func (s *Server) gameApplyRoleOptions(userID int64, game games.Game) []GameApplyRoleOptionDTO {
	snapshot := s.profiles.RoleSnapshot(userID)
	status := snapshot.RoleStatusMap
	items := []GameApplyRoleOptionDTO{
		{Key: "player", Label: "玩家"},
		{Key: "expert", Label: "行家"},
		{Key: "guide", Label: "领路人"},
	}
	for index := range items {
		item := &items[index]
		if !games.GameAllowsRole(game, item.Key) {
			item.DisabledReason = "本局未开放" + item.Label + "身份入局"
			continue
		}
		if item.Key == "expert" && status["expert"] != "approved" && status["expert"] != "active" {
			item.DisabledReason = "当前账号未开通行家身份"
			continue
		}
		if item.Key == "guide" && status["guide"] != "approved" && status["guide"] != "active" {
			item.DisabledReason = "当前账号未开通领路人身份"
			continue
		}
		if games.RoleOccupiesSeat(game, item.Key) && game.MaxPlayers > 0 && game.CurrentPlayers >= game.MaxPlayers {
			item.DisabledReason = "该局已满员"
			continue
		}
		item.Enabled = true
	}
	return items
}

func (s *Server) buildGameRelation(userID int64, game games.Game) GameMyRelationDTO {
	isCreator := game.CreatorUserID == userID
	isMember := s.games.IsMember(game.ID, userID)
	signupOpen := games.CanApplyWithinSignupWindow(game, time.Now())
	canManageProgress := isCreator || (game.MainGuideUserID > 0 && game.MainGuideUserID == userID)
	startableStatus := game.Status == "recruiting" || game.Status == "full"
	relationRole := s.userRoleForGame(game, userID)
	completionState := s.gameCompletionState(game, userID)
	canConfirm, _ := completionState["canConfirm"].(bool)
	applyRoleOptions := s.gameApplyRoleOptions(userID, game)
	hasEnabledApplyRole := false
	firstRoleReason := ""
	for _, option := range applyRoleOptions {
		if option.Enabled {
			hasEnabledApplyRole = true
		}
		if firstRoleReason == "" && option.DisabledReason != "" {
			firstRoleReason = option.DisabledReason
		}
	}
	relation := GameMyRelationDTO{
		Role:             relationRole,
		AllowedRoles:     games.AllowedRolesForGame(game),
		ApplyRoleOptions: applyRoleOptions,
		IsCreator:        isCreator,
		IsMember:         isMember,
		CanApply:         game.Status == "recruiting" && !isMember && signupOpen && hasEnabledApplyRole,
		CanAudit:         (isCreator || game.MainGuideUserID == userID) && game.Status == "recruiting",
		CanStart:         canManageProgress && startableStatus && game.CurrentPlayers >= game.MinPlayers,
		CanEnterIM:       isMember && (game.Status == "in_progress" || game.Status == "pending_confirm" || game.Status == "pending_review" || game.Status == "completed"),
		// 交付确认还要受行家优先、角色和已确认状态约束，不能只按“局内成员”
		// 粗略放开，否则详情页会把无法提交的成员带到确认流程。
		CanConfirm: canConfirm,
	}
	if !isCreator && !isMember {
		if allowed, message := s.canUseCreditAction(userID, "join_game"); !allowed {
			relation.CanApply = false
			relation.ApplyDisabledReason = message
		}
	}
	if game.Status == "recruiting" && !isMember && !hasEnabledApplyRole {
		if relation.ApplyDisabledReason == "" {
			relation.ApplyDisabledReason = firstRoleReason
		}
	} else if game.Status == "recruiting" && !isMember && !signupOpen {
		if relation.ApplyDisabledReason == "" {
			switch games.SignupWindowStateAt(game, time.Now()) {
			case "not_started":
				relation.ApplyDisabledReason = "报名尚未开始"
			case "ended":
				relation.ApplyDisabledReason = "报名已截止"
			default:
				relation.ApplyDisabledReason = "当前不在报名时间内"
			}
		}
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
			relation.ApplyDisabledReason = games.ApplicationStatusText(latest.Status)
		} else if latest.Status == "rejected" && !s.currentGameApplicationConfig().AllowDuplicateApply {
			relation.CanApply = false
			relation.ApplyDisabledReason = games.ApplicationStatusText(latest.Status)
		}
	}
	if !relation.IsMember && !relation.IsCreator && relation.ApplyDisabledReason == "" {
		switch game.Status {
		case "in_progress":
			relation.ApplyDisabledReason = "组局进行中，暂不可报名"
		case "pending_confirm", "pending_review", "completed":
			relation.ApplyDisabledReason = "本局已结束，暂不可报名"
		case "canceled", "cancelled":
			relation.ApplyDisabledReason = "本局已取消，暂不可报名"
		case "pending_audit":
			relation.ApplyDisabledReason = "组局尚在后台审核，暂不可报名"
		case "draft":
			relation.ApplyDisabledReason = "草稿局不可报名"
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
	case "guide_escort", "guide-escort", "escort", "observer":
		return "guide_escort"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
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
		s.notifyGameApproved(w, game)
	} else {
		s.notifyGameRejected(w, game, remark)
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "批量审核请求格式不正确")
		return
	}
	if len(req.GameIDs) == 0 || len(req.GameIDs) > 100 || (!req.Approve && strings.TrimSpace(req.Remark) == "") {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请选择 1-100 条组局；驳回时必须填写原因")
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
				s.notifyGameApproved(w, game)
			} else {
				s.notifyGameRejected(w, game, strings.TrimSpace(req.Remark))
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
	if allowed, message := s.canUseCreditAction(userID, "join_game"); !allowed {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, message)
		return
	}
	id, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/applications")
	if !ok {
		return
	}
	var req struct {
		Reason   string  `json:"reason"`
		Role     string  `json:"role"`
		RoleType string  `json:"roleType"`
		FileIDs  []int64 `json:"fileIds"`
		Agreed   bool    `json:"agreed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	// 局状态、报名窗口和满员是整个申请表单的前置条件。应先返回
	// “当前状态不可报名”，避免用户在已结束的局上被误导去补自我介绍。
	game, gameErr := s.games.Get(id)
	if gameErr != nil {
		writeGameError(w, gameErr)
		return
	}
	if game.Status != games.StatusRecruiting {
		writeGameError(w, games.ErrGameNotRecruiting)
		return
	}
	if !games.CanApplyWithinSignupWindow(game, time.Now()) {
		writeGameError(w, games.ErrSignupClosed)
		return
	}
	if game.CurrentPlayers >= game.MaxPlayers {
		writeGameError(w, games.ErrFull)
		return
	}
	config, configErr := s.currentGameApplicationConfigStrict()
	if configErr != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取入局申请规则失败，请稍后重试")
		return
	}
	if config.RequireRealname && !s.identity.IsRealnameVerified(userID) {
		httpx.Error(w, http.StatusForbidden, 40341, "申请入局前请先完成实名认证")
		return
	}
	reasonLength := len([]rune(strings.TrimSpace(req.Reason)))
	if config.RequireIntro && reasonLength == 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先填写自我介绍")
		return
	}
	if config.MinIntroLength > 0 && reasonLength > 0 && reasonLength < config.MinIntroLength {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, fmt.Sprintf("自我介绍不少于%d字", config.MinIntroLength))
		return
	}
	if config.RequireAgreement && !req.Agreed {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先阅读并同意入局申请须知")
		return
	}
	if config.MaxUploadCount > 0 && len(req.FileIDs) > config.MaxUploadCount {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, fmt.Sprintf("最多上传%d个申请材料", config.MaxUploadCount))
		return
	}
	if config.UploadRequired && len(req.FileIDs) == 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请先上传相关经历或作品")
		return
	}
	requestedRole := strings.ToLower(strings.TrimSpace(firstNonEmpty(req.RoleType, req.Role)))
	snapshot, snapshotErr := s.profiles.RoleSnapshotStrict(userID)
	if snapshotErr != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取角色身份失败，请稍后重试")
		return
	}
	switch requestedRole {
	case "guide", "leader", "main_guide":
		if snapshot.RoleStatusMap["guide"] != "approved" && snapshot.RoleStatusMap["guide"] != "active" {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "当前账号未开通领路人身份")
			return
		}
		req.RoleType = "guide"
	case "guide_escort", "guide-escort", "escort", "observer":
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "领路人护航功能暂未开放")
		return
	case "expert", "master":
		if snapshot.RoleStatusMap["expert"] != "approved" && snapshot.RoleStatusMap["expert"] != "active" {
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
		_, _ = s.createCriticalNotification(w, "game_application_created", notifications.CreateRequest{
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
	gameID, ok := gameIDFromPath(w, r.URL.Path, "/api/app/games/", "/guide-invitations")
	if !ok {
		return
	}
	game, err := s.games.Get(gameID)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if allowed, message := s.currentGameInvitePermission(userID, game); !allowed {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, message)
		return
	}
	var req games.InvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	invitation, err := s.games.CreateInvitation(userID, gameID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	s.recordBehavior(userID, "create_game_invitation", "game", gameID, map[string]interface{}{"invitationId": invitation.ID, "targetUserId": invitation.TargetUserID})
	_, _ = s.createCriticalNotification(w, "create_game_invitation", notifications.CreateRequest{
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	if req.Accept {
		if allowed, message := s.canUseCreditAction(userID, "accept_invitation"); !allowed {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, message)
			return
		}
	}
	invitation, app, err := s.games.RespondInvitation(userID, invitationID, req)
	if err != nil {
		writeGameError(w, err)
		return
	}
	successRoute := ""
	if invitation.Status == "accepted" {
		if connectionErr := s.connections.UpsertPairStrict(invitation.InviterID, invitation.TargetUserID, "guide_match", "guide_match", invitation.ID, 3); connectionErr != nil {
			markConnectionPersistenceDegraded(w, "invitation_accept", invitation.InviterID, invitation.TargetUserID, connectionErr)
		}
		if strings.TrimSpace(invitation.InviteGroupID) != "" {
			s.createPairedInvitationProgressNotification(w, invitation)
		}
		if app.Status == "approved" {
			s.createGameInvitationSuccessNotifications(w, invitation)
			if _, _, allConfirmed := s.invitationConfirmedParties(invitation); allConfirmed && guideProgressInvitationRole(invitation.Role) == "expert" {
				successRoute, _ = gameInvitationSuccessRoute(invitation.GameID, "expert")
			}
		} else {
			s.createInvitationReviewNotification(w, invitation, app)
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
	items, err := s.games.ApplicationsForUserStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取我的报名申请失败，请稍后重试")
		return
	}
	httpx.OK(w, map[string]interface{}{"items": items})
}

func (s *Server) receivedApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	gameID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("gameId")), 10, 64)
	applications, err := s.games.ApplicationsForCreatorStrict(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "读取待审核报名申请失败，请稍后重试")
		return
	}
	items := filterReceivedApplications(applications, status, gameID)
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
			{"label": "局分类", "value": homeGameCategoryText(game)},
			{"label": "活动时长", "value": firstNonEmpty(durationText, "未设置")},
			{"label": "参与人数", "value": gameMemberCountText(game)},
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
	case "guide_escort", "guide-escort", "escort", "observer":
		return "guide_escort", "\u62a4\u822a\u9886\u8def\u4eba"
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
		reviewedAt := time.Now()
		resultText := games.ApplicationStatusText("rejected")
		if req.Approve {
			resultText = games.ApplicationStatusText("approved")
		}
		templateID := s.applicationResultSubscribeTemplateID(notifyType)
		content = applicationReviewNotificationContent(game.Title, resultText, reviewedAt, app.RejectReason)
		notice, notificationErr := s.createCriticalNotification(w, "game_application_reviewed", notifications.CreateRequest{
			UserID:           app.UserID,
			NotifyType:       notifyType,
			Title:            title,
			Content:          content,
			BizType:          "game",
			BizID:            game.ID,
			NeedWechat:       templateID != "",
			WechatTemplateID: templateID,
			WechatData: map[string]string{
				"thing1": game.Title,
				"thing2": resultText,
				"time3":  reviewedAt.Local().Format("2006-01-02 15:04"),
				"page":   "pages/game/detail/index?id=" + strconv.FormatInt(game.ID, 10),
			},
		})
		if notificationErr == nil && notice.NeedWechat && notice.WechatTaskID > 0 {
			// 授权成功后立即尝试发送；发送失败时通知已保留在消息中心，
			// 订阅任务保持待重试状态，不影响审核结果返回。
			_, _ = s.notices.SendWechatTask(notice.WechatTaskID)
		}
	}
	if req.Approve {
		s.createGameInvitationSuccessNotificationsForApplication(w, app)
		if game, gameErr := s.games.Get(app.GameID); gameErr == nil {
			s.createGameReadyToStartNotification(w, game)
		}
	}
	httpx.OK(w, s.buildReceivedApplicationItem(userID, app))
}

func applicationReviewNotificationContent(gameTitle string, resultText string, reviewedAt time.Time, rejectReason string) string {
	gameTitle = strings.TrimSpace(gameTitle)
	if gameTitle == "" {
		gameTitle = "未命名组局"
	}
	content := "你申请加入的《" + gameTitle + "》审核结果：" + resultText + "。审核时间：" + reviewedAt.Local().Format("2006-01-02 15:04") + "。"
	if strings.TrimSpace(rejectReason) != "" {
		content += "拒绝原因：" + strings.TrimSpace(rejectReason)
	}
	return content
}

func (s *Server) notifyGameApproved(w http.ResponseWriter, game games.Game) {
	_, _ = s.createCriticalNotification(w, "game_audit_approved", notifications.CreateRequest{
		UserID:     game.CreatorUserID,
		NotifyType: "game_approved",
		Title:      "\u7ec4\u5c40\u5ba1\u6838\u901a\u8fc7",
		Content:    "\u4f60\u7684\u7ec4\u5c40\u300a" + game.Title + "\u300b\u5df2\u901a\u8fc7\u5ba1\u6838\uff0c\u73b0\u5df2\u8fdb\u5165\u62db\u52df\u4e2d\u3002",
		BizType:    "game",
		BizID:      game.ID,
	})
}

func (s *Server) notifyGameRejected(w http.ResponseWriter, game games.Game, reason string) {
	content := "你的组局《" + game.Title + "》未通过审核。"
	if strings.TrimSpace(reason) != "" {
		content += "原因：" + strings.TrimSpace(reason)
	}
	_, _ = s.createCriticalNotification(w, "game_audit_rejected", notifications.CreateRequest{
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "邀约 ID 错误")
		return 0, false
	}
	id, err := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "邀约 ID 错误")
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权申请取消该局")
		return
	}
	if game.Status != "in_progress" {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "当前局状态不可取消")
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	req.ReasonKey = strings.TrimSpace(req.ReasonKey)
	req.ReasonText = strings.TrimSpace(req.ReasonText)
	if req.ReasonKey == "" || req.ReasonText == "" || !validCancelReason(req.ReasonKey, s.currentGameCancelConfig().Player.ReasonOptions) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请选择并填写取消原因")
		return
	}
	if req.CompensationRate < 0 || req.CompensationRate > 100 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "赔付比例不合法")
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
	membersBeforeCancel, membersErr := s.games.MembersStrict(gameID)
	if membersErr != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取局成员失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "player_cancel_request", "game", gameID, extra)
	if game.CreatorUserID != userID {
		result, err := s.games.Exit(userID, gameID)
		if err != nil {
			writeGameError(w, err)
			return
		}
		credit, creditErr := s.reviews.DeductCreditStrict(userID, gameID, "player_cancel_service")
		if creditErr != nil {
			markCreditPersistenceDegraded(w, "player_cancel_after_exit", userID, gameID, creditErr)
		}
		result.CreditLogID = credit.ID
		if err := s.games.RecordExitCredit(gameID, userID, credit.ID); err != nil {
			writeGameError(w, err)
			return
		}
		s.createPlayerCancelNotifications(w, userID, game, req.ReasonText, false, nil)
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
	credit, creditErr := s.reviews.DeductCreditStrict(userID, gameID, "player_cancel_service")
	if creditErr != nil {
		markCreditPersistenceDegraded(w, "player_cancel_after_cancel", userID, gameID, creditErr)
	}
	s.createPlayerCancelNotifications(w, userID, game, req.ReasonText, true, membersBeforeCancel)
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
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权申请取消该局")
			return
		}
	} else if game.CreatorUserID != userID && game.MainGuideUserID != userID {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权取消该局")
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
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "当前局状态不可取消")
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
		warningTitle = "取消本局提醒"
		warningDesc = "本局为免费局，取消后将按规则扣减信用分。"
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

func (s *Server) createPlayerCancelNotifications(w http.ResponseWriter, userID int64, game games.Game, reason string, gameCanceled bool, memberIDs []int64) {
	recipients := make(map[int64]bool)
	if gameCanceled {
		for _, memberID := range memberIDs {
			if memberID > 0 && memberID != userID {
				recipients[memberID] = true
			}
		}
	}
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
		_, _ = s.createCriticalNotification(w, "player_cancel_request", notifications.CreateRequest{
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
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "无权取消该局")
		return
	}
	if game.Status != "in_progress" {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "当前局状态不可取消")
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	req.ReasonKey = strings.TrimSpace(req.ReasonKey)
	req.ReasonText = strings.TrimSpace(req.ReasonText)
	if req.ReasonKey == "" || req.ReasonText == "" || !validCancelReason(req.ReasonKey, s.currentGameCancelConfig().Expert.ReasonOptions) {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "请选择并填写取消原因")
		return
	}
	if req.CompensationRate < 0 || req.CompensationRate > 100 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "赔付比例不合法")
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
	membersBeforeCancel, membersErr := s.games.MembersStrict(gameID)
	if membersErr != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取局成员失败，请稍后重试")
		return
	}
	s.recordBehavior(userID, "expert_cancel_request", "game", gameID, extra)
	canceledGame, err := s.games.CancelService(gameID, "expert_cancel_service")
	if err != nil {
		writeGameError(w, err)
		return
	}
	credit, creditErr := s.reviews.DeductCreditStrict(userID, gameID, "expert_cancel_service")
	if creditErr != nil {
		markCreditPersistenceDegraded(w, "expert_cancel_after_cancel", userID, gameID, creditErr)
	}
	s.createExpertCancelNotifications(w, userID, game, req.ReasonText, membersBeforeCancel)
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

func (s *Server) createExpertCancelNotifications(w http.ResponseWriter, userID int64, game games.Game, reason string, memberIDs []int64) {
	expertName := s.inGameDisplayName(userID, "行家")
	content := expertName + "取消了本次组局"
	if reason != "" {
		content += "，原因：" + reason
	}
	for _, recipientID := range memberIDs {
		if recipientID == userID {
			continue
		}
		_, _ = s.createCriticalNotification(w, "expert_cancel_request", notifications.CreateRequest{
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
			return
		}
	}
	startReason := strings.TrimSpace(req.StartReason)
	if startReason == "" {
		startReason = "发起人手动开始"
	}
	members, membersErr := s.games.MembersStrict(id)
	if membersErr != nil {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "读取局成员失败，请稍后重试")
		return
	}
	game, err := s.games.ManualStartWithReason(userID, id, startReason)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if _, roomErr := s.im.EnsureRoomStrict(id); roomErr != nil {
		markIMPersistenceDegraded(w, "manual_start", id, roomErr)
	} else {
		s.createManualStartChatMessage(userID, game)
	}
	s.createManualStartNotifications(w, userID, game, members)
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
	s.broadcastIMMessage(message)
}

func (s *Server) createManualStartNotifications(w http.ResponseWriter, userID int64, game games.Game, memberIDs []int64) {
	starterName := s.inGameDisplayName(userID, "组建者")
	content := starterName + "已开始组局，请及时查看并参与"
	if title := strings.TrimSpace(game.Title); title != "" {
		content = "「" + title + "」" + content
	}
	for _, recipientID := range memberIDs {
		_, _ = s.createCriticalNotification(w, "manual_start", notifications.CreateRequest{
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
	game, err := s.games.Get(id)
	if err != nil {
		writeGameError(w, err)
		return
	}
	if !s.games.IsMember(id, userID) {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅局内成员可退出")
		return
	}
	if atomic, ok := s.games.(interface {
		SupportsExitWithCreditMutation() bool
		ExitWithCreditMutation(userID int64, gameID int64, memberStatus string, reason string, mutation games.ExitCreditMutation) (games.ExitCreditResult, error)
	}); ok && atomic.SupportsExitWithCreditMutation() && gameNeedsExitCredit(game.Status) {
		reason := exitCreditReason(game.Status)
		changeValue, creditRuleErr := s.reviews.CreditDeductionValueStrict(reason)
		if creditRuleErr != nil {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "信用规则暂时不可用，请稍后重试")
			return
		}
		creditConfig, configErr := s.currentCreditRestrictionConfigStrict()
		if configErr != nil {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeSystemError, "信用规则暂时不可用，请稍后重试")
			return
		}
		atomicResult, err := atomic.ExitWithCreditMutation(userID, id, games.ExitMemberStatusForGame(game.Status), reason, games.ExitCreditMutation{
			ChangeValue:  changeValue,
			Reason:       reason,
			InitialScore: creditConfig.InitialScore,
		})
		if err != nil {
			writeGameError(w, err)
			return
		}
		credit := reviews.CreditLog{
			ID:          atomicResult.CreditLogID,
			UserID:      userID,
			GameID:      id,
			ChangeValue: atomicResult.ChangeValue,
			BeforeScore: atomicResult.BeforeScore,
			AfterScore:  atomicResult.AfterScore,
			Reason:      reason,
			CreatedAt:   atomicResult.CreatedAt,
		}
		if atomicResult.CreditDeducted {
			s.createExitNotifications(w, atomicResult.ExitResult, credit.ChangeValue)
		}
		httpx.OK(w, map[string]interface{}{"game": atomicResult.Game, "exit": atomicResult.ExitResult, "credit": credit})
		return
	}
	var credit reviews.CreditLog
	if gameNeedsExitCredit(game.Status) {
		var creditErr error
		credit, creditErr = s.reviews.DeductCreditStrict(userID, id, exitCreditReason(game.Status))
		if creditErr != nil {
			writeGameError(w, creditErr)
			return
		}
	}
	result, err := s.games.ExitWithCredit(userID, id, credit.ID)
	if err != nil {
		if credit.ID > 0 && credit.ChangeValue < 0 {
			if _, restoreErr := s.reviews.RestoreCreditStrict(userID, id, "exit_rollback", -credit.ChangeValue); restoreErr != nil {
				markCreditPersistenceDegraded(w, "exit_rollback", userID, id, restoreErr)
			}
			_ = s.games.RestoreMemberAfterExit(userID, id)
		}
		writeGameError(w, err)
		return
	}
	if result.CreditDeduct {
		s.createExitNotifications(w, result, credit.ChangeValue)
		httpx.OK(w, map[string]interface{}{"game": result.Game, "exit": result, "credit": credit})
		return
	}
	s.createExitNotifications(w, result, 0)
	httpx.OK(w, map[string]interface{}{"game": result.Game, "exit": result})
}

func gameNeedsExitCredit(status string) bool {
	return status == "recruiting" || status == "full" || status == "pending_confirm" || status == "in_progress" || status == "pending_review" || status == "completed"
}

func exitCreditReason(status string) string {
	if status == "recruiting" || status == "full" {
		return "quit_after_admitted"
	}
	if status == "pending_confirm" {
		return "quit_after_confirm"
	}
	return "quit_after_started"
}

func (s *Server) createExitNotifications(w http.ResponseWriter, result games.ExitResult, creditChange int) {
	content := "成员已退出局"
	if result.CreditDeduct {
		content = "成员退出局，已扣减信用分 " + strconv.Itoa(-creditChange)
	}
	if result.Game.CreatorUserID > 0 && result.Game.CreatorUserID != result.UserID {
		_, _ = s.createCriticalNotification(w, "exit_game", notifications.CreateRequest{
			UserID:     result.Game.CreatorUserID,
			NotifyType: "game_member_quit",
			Title:      "成员退出局",
			Content:    content,
			BizType:    "game",
			BizID:      result.GameID,
		})
	}
	_, _ = s.createCriticalNotification(w, "exit_game", notifications.CreateRequest{
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
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数不正确")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || len(req.Reason) > 300 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "异常原因不能为空且不能超过 300 字")
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
	if allowed, message := s.canUseCreditAction(userID, "create_game"); !allowed {
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, message)
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
	if s.rejectSensitiveGameContent(w, req.Title) {
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
		config["categoryTabs"] = normalizeMyGamesCategoryTabs(config["categoryTabs"])
		return config
	}
	return map[string]interface{}{
		"pageTitle":     "\u6211\u7684\u5c40",
		"emptyText":     "\u6682\u65e0\u76f8\u5173\u5c40",
		"detailMissing": "\u6682\u65e0\u7ec4\u5c40\u8be6\u60c5",
		"actionMissing": "\u6682\u65e0\u53ef\u6267\u884c\u64cd\u4f5c",
		"categoryTabs":  normalizeMyGamesCategoryTabs(nil),
		"statusTabs": []map[string]interface{}{
			{"key": "all", "text": "\u5168\u90e8"},
			{"key": "active", "text": "\u8fdb\u884c\u4e2d"},
			{"key": "complete", "text": "\u5df2\u5b8c\u6210"},
			{"key": "overdue", "text": "\u8d85\u65f6"},
			{"key": "canceled", "text": "\u5df2\u53d6\u6d88"},
			{"key": "dispute", "text": "\u4e89\u8bae\u4e2d"},
		},
	}
}

// normalizeMyGamesCategoryTabs keeps the two distinct business views that
// were previously conflated: a user can both participate in games and manage
// games they created or serve. Existing operation configuration is retained,
// while missing canonical tabs are restored so management cards never become
// inaccessible after an older configuration is loaded.
func normalizeMyGamesCategoryTabs(raw interface{}) []map[string]interface{} {
	defaults := []map[string]interface{}{
		{"key": "joined", "text": "我参与的"},
		{"key": "created", "text": "我发起/管理的"},
		{"key": "invited", "text": "我受邀的"},
		{"key": "favorite", "text": "我收藏的"},
	}
	configured := map[string]map[string]interface{}{}
	items, _ := raw.([]interface{})
	for _, item := range items {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		key := strings.TrimSpace(stringValue(row["key"]))
		if key == "joined" || key == "created" || key == "invited" || key == "favorite" {
			configured[key] = row
		}
	}
	result := make([]map[string]interface{}, 0, len(defaults))
	for _, fallback := range defaults {
		key := fallback["key"].(string)
		if item, ok := configured[key]; ok {
			text := strings.TrimSpace(stringValue(item["text"]))
			if text != "" && !(key == "created" && text == "我受邀的") {
				result = append(result, map[string]interface{}{"key": key, "text": text})
				continue
			}
		}
		result = append(result, fallback)
	}
	return result
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
		httpx.Error(w, http.StatusConflict, 40922, "该局已满员")
	case errors.Is(err, games.ErrRoleNotAllowed):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "本局未开放当前身份入局")
	case errors.Is(err, games.ErrInvalidGameInput):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "组局信息不完整或格式不正确")
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
