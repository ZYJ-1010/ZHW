package appapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"zhw-mini/services/go-api/internal/adminauth"
	"zhw-mini/services/go-api/internal/aidata"
	"zhw-mini/services/go-api/internal/audit"
	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/common/config"
	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/delivery"
	"zhw-mini/services/go-api/internal/exports"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/invites"
	"zhw-mini/services/go-api/internal/lbs"
	"zhw-mini/services/go-api/internal/memberreports"
	"zhw-mini/services/go-api/internal/membership"
	"zhw-mini/services/go-api/internal/notifications"
	"zhw-mini/services/go-api/internal/orders"
	"zhw-mini/services/go-api/internal/points"
	"zhw-mini/services/go-api/internal/profiles"
	"zhw-mini/services/go-api/internal/redemption"
	"zhw-mini/services/go-api/internal/reports"
	"zhw-mini/services/go-api/internal/revenue"
	"zhw-mini/services/go-api/internal/reviews"
	"zhw-mini/services/go-api/internal/systemconfig"
	"zhw-mini/services/go-api/internal/teams"
	"zhw-mini/services/go-api/internal/users"
)

type CurrentUserDTO struct {
	users.User
	Identity      identity.Record       `json:"identity"`
	Roles         []string              `json:"roles"`
	RoleStatusMap map[string]string     `json:"roleStatusMap"`
	Membership    membership.Membership `json:"membership"`
	Growth        GrowthDTO             `json:"growth"`
	InviteCode    string                `json:"inviteCode"`
	IncomeSummary revenue.IncomeSummary `json:"incomeSummary"`
	Points        points.Account        `json:"pointsSummary"`
}

type GrowthDTO struct {
	Level            int `json:"level"`
	ExperienceValue  int `json:"experienceValue"`
	CreditScore      int `json:"creditScore"`
	TodayCreditScore int `json:"todayCreditScore"`
	Points           int `json:"points"`
}

type appLoginResponse struct {
	Token                   string            `json:"token,omitempty"`
	PreAuthToken            string            `json:"preAuthToken,omitempty"`
	ExpiresAt               string            `json:"expiresAt"`
	User                    CurrentUserDTO    `json:"user"`
	InviteRelation          *invites.Relation `json:"inviteRelation,omitempty"`
	NeedProfile             bool              `json:"needProfile"`
	RequiresIdentityBinding bool              `json:"requiresIdentityBinding"`
	IdentityBindStatus      string            `json:"identityBindStatus,omitempty"`
	Roles                   []string          `json:"roles"`
	RoleStatusMap           map[string]string `json:"roleStatusMap"`
	EntryType               string            `json:"entryType"`
	AuthPageMode            string            `json:"authPageMode"`
	BoundWechat             bool              `json:"boundWechat"`
	InviteBindingStatus     string            `json:"inviteBindingStatus,omitempty"`
	InviteBindingMessage    string            `json:"inviteBindingMessage,omitempty"`
}

type inviteEntryResponse struct {
	InviteCode     string      `json:"inviteCode"`
	EntryType      string      `json:"entryType"`
	Path           string      `json:"path"`
	Title          string      `json:"title,omitempty"`
	UrlLink        string      `json:"urlLink,omitempty"`
	UrlLinkError   string      `json:"urlLinkError,omitempty"`
	Scene          string      `json:"scene,omitempty"`
	WxaCodeDataURL string      `json:"wxaCodeDataUrl,omitempty"`
	Game           interface{} `json:"game,omitempty"`
}

type CurrentUserSummaryDTO struct {
	User                    CurrentUserDTO        `json:"user"`
	ReviewTodoCount         int                   `json:"reviewTodoCount"`
	UnreadNotificationCount int                   `json:"unreadNotificationCount"`
	IncomeSummary           revenue.IncomeSummary `json:"incomeSummary"`
	PointsSummary           points.Account        `json:"pointsSummary"`
	RecentFootprints        []reviews.Footprint   `json:"recentFootprints"`
}

type Server struct {
	auth                 *auth.Service
	identity             identityService
	games                gameService
	lbs                  lbsService
	mapProvider          lbs.MapProvider
	im                   imService
	reviews              reviewService
	revenue              revenueService
	reports              reportService
	memberReports        *memberreports.Service
	membership           *membership.Service
	teams                *teams.Service
	orders               *orders.Service
	points               *points.Service
	redemption           *redemption.Service
	connections          *connections.Service
	profiles             *profiles.Service
	files                *files.Service
	audit                *audit.Service
	notices              *notifications.Service
	exports              *exports.Service
	admins               *adminauth.Service
	delivery             *delivery.Service
	aidata               *aidata.Service
	systemConfig         *systemconfig.Service
	imSocketHub          *imSocketHub
	idempotency          *idempotencyStore
	cfg                  config.Config
	reviewReplies        map[int64]profileReviewReply
	reviewLikes          map[int64]map[int64]bool
	reviewReplyMu        sync.RWMutex
	gameCategoryConfig   gameCategoryConfigDTO
	gameCategoryConfigMu sync.RWMutex
	mapPlayMu            sync.RWMutex
	mapBlindRoutes       map[int64]mapBlindRouteDTO
	mapChallenges        map[int64]mapChallengeDTO
	mapProviderLimitMu   sync.Mutex
	mapProviderLastSeen  map[string]time.Time
	gameAuditRejectMu    sync.RWMutex
	gameAuditRejects     map[int64]string

	faceIDCallbackSecret           string
	faceIDCallbackRequireSignature bool
	nearbyDefaultRadiusMeter       float64
	productionMode                 bool
	inviteURLLinkBaseURL           string
}

func New(authService *auth.Service, identityService identityService, gameService gameService, lbsService lbsService, imService imService) *Server {
	reviewService := reviews.NewService(gameService)
	revenueService := revenue.NewService(reviewService)
	pointsService := points.NewService()
	server := &Server{auth: authService, identity: identityService, games: gameService, lbs: lbsService, im: imService, reviews: reviewService, revenue: revenueService, reports: reports.NewService(revenueService), memberReports: memberreports.NewService(gameService, revenueService), membership: membership.NewService(), teams: teams.NewService(revenueService), orders: orders.NewService(), points: pointsService, redemption: redemption.NewService(pointsService), connections: connections.NewService(), profiles: profiles.NewService(), files: files.NewService(), audit: audit.NewService(), notices: notifications.NewService(), exports: exports.NewService(), admins: adminauth.NewService(), delivery: delivery.NewService(), aidata: aidata.NewService(), systemConfig: systemconfig.NewService(), reviewReplies: make(map[int64]profileReviewReply), reviewLikes: make(map[int64]map[int64]bool), mapBlindRoutes: make(map[int64]mapBlindRouteDTO), mapChallenges: make(map[int64]mapChallengeDTO), mapProviderLastSeen: make(map[string]time.Time), gameAuditRejects: make(map[int64]string), nearbyDefaultRadiusMeter: 5000}
	server.imSocketHub = newIMSocketHub(server)
	return server
}

func (s *Server) Configure(cfg config.Config) {
	s.cfg = cfg
	s.productionMode = appAPIProductionEnv(cfg.AppEnv)
	s.files.RequirePublicBaseURLs(s.productionMode)
	if cfg.AppLimits.DailyGameLimit > 0 {
		if service, ok := s.games.(*games.Service); ok {
			service.SetDailyCreateLimit(cfg.AppLimits.DailyGameLimit)
		}
	}
	if cfg.LBS.DefaultRadiusMeter > 0 {
		s.nearbyDefaultRadiusMeter = cfg.LBS.DefaultRadiusMeter
	}
	if cfg.TencentMap.Enabled && strings.TrimSpace(cfg.TencentMap.KeyServer) != "" {
		timeoutMS := cfg.TencentMap.RequestTimeoutMS
		if timeoutMS <= 0 {
			timeoutMS = 3000
		}
		provider, err := lbs.NewTencentMapClient(lbs.TencentMapConfig{
			KeyServer: cfg.TencentMap.KeyServer,
			SK:        cfg.TencentMap.SK,
			APIBase:   cfg.TencentMap.APIBase,
			Timeout:   time.Duration(timeoutMS) * time.Millisecond,
		})
		if err == nil {
			s.mapProvider = provider
		}
	}
	s.files.UsePublicBaseURLs(cfg.Storage.UploadBaseURL, cfg.Storage.DownloadBaseURL)
	if strings.EqualFold(strings.TrimSpace(cfg.Storage.Provider), "cos") {
		s.files.UseCOSPostSigner(files.NewCOSPostSigner(cfg.Storage.COSSecretID, cfg.Storage.COSSecretKey))
	}
	s.inviteURLLinkBaseURL = strings.TrimSpace(cfg.Wechat.URLLinkBaseURL)
	s.notices.UseOpenIDResolver(func(userID int64) (string, bool) {
		user, ok := s.auth.UserByID(userID)
		if !ok || strings.TrimSpace(user.OpenID) == "" {
			return "", false
		}
		return user.OpenID, true
	})
	if cfg.Wechat.AppID != "" && cfg.Wechat.AppSecret != "" {
		s.notices.UseWechatSubscribeSender(notifications.NewWechatSubscribeSender(cfg.Wechat.AppID, cfg.Wechat.AppSecret))
	}
}

func (s *Server) UseMapProvider(provider lbs.MapProvider) {
	s.mapProvider = provider
}

func appAPIProductionEnv(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "prod" || value == "production"
}

func (s *Server) UseFaceIDCallbackVerifier(secret string, requireSignature bool) {
	s.faceIDCallbackSecret = strings.TrimSpace(secret)
	s.faceIDCallbackRequireSignature = requireSignature
}

func (s *Server) UseAdminRepository(repository adminauth.Repository) {
	if repository != nil {
		s.admins = adminauth.NewServiceWithRepository(repository)
	}
}

func (s *Server) UseAIDataRepository(repository aidata.Repository) {
	if repository != nil {
		s.aidata = aidata.NewServiceWithRepository(repository)
	}
}

func (s *Server) UseSystemConfigRepository(repository systemconfig.Repository) {
	if repository != nil {
		s.systemConfig = systemconfig.NewServiceWithRepository(repository)
		s.ensureDefaultSystemConfigs()
	}
}

func (s *Server) UseRepositories(behaviorRepo audit.BehaviorRepository, operationRepo audit.OperationRepository, orderRepo orders.Repository, reportRepo reports.Repository, notificationRepo notifications.Repository, fileRepo files.Repository, reviewRepo reviews.Repository, revenueRepo revenue.Repository, pointRepo points.Repository, redemptionRepo redemption.Repository, connectionRepo connections.Repository, profileRepo profiles.Repository, lbsRepo lbs.Repository, memberReportRepo memberreports.Repository, membershipRepo membership.Repository, teamRepo teams.Repository) {
	if behaviorRepo != nil || operationRepo != nil {
		s.audit = audit.NewServiceWithRepositories(behaviorRepo, operationRepo)
	}
	if reviewRepo != nil {
		s.reviews = reviews.NewServiceWithRepository(s.games, reviewRepo)
		if revenueRepo != nil {
			s.revenue = revenue.NewServiceWithRepository(s.reviews, revenueRepo)
		} else {
			s.revenue = revenue.NewService(s.reviews)
		}
		s.reports = reports.NewService(s.revenue)
		s.memberReports = memberreports.NewService(s.games, s.revenue)
		s.teams = teams.NewService(s.revenue)
	}
	if reviewRepo == nil && revenueRepo != nil {
		s.revenue = revenue.NewServiceWithRepository(s.reviews, revenueRepo)
		s.reports = reports.NewService(s.revenue)
		s.memberReports = memberreports.NewService(s.games, s.revenue)
		s.teams = teams.NewService(s.revenue)
	}
	if orderRepo != nil {
		s.orders = orders.NewServiceWithRepository(orderRepo)
	}
	if reportRepo != nil {
		s.reports = reports.NewServiceWithRepository(s.revenue, reportRepo)
	}
	if notificationRepo != nil {
		s.notices = notifications.NewServiceWithRepository(notificationRepo)
	}
	if fileRepo != nil {
		s.files = files.NewServiceWithRepository(fileRepo)
	}
	if pointRepo != nil {
		s.points = points.NewServiceWithRepository(pointRepo)
		s.redemption = redemption.NewService(s.points)
	}
	if redemptionRepo != nil {
		s.redemption = redemption.NewServiceWithRepository(s.points, redemptionRepo)
	}
	if connectionRepo != nil {
		s.connections = connections.NewServiceWithRepository(connectionRepo)
	}
	if profileRepo != nil {
		s.profiles = profiles.NewServiceWithRepository(profileRepo)
	}
	if lbsRepo != nil {
		s.lbs = lbs.NewServiceWithRepository(lbsRepo)
	}
	if memberReportRepo != nil {
		s.memberReports = memberreports.NewServiceWithRepository(s.games, s.revenue, memberReportRepo)
	}
	if membershipRepo != nil {
		s.membership = membership.NewServiceWithRepository(membershipRepo)
	}
	if teamRepo != nil {
		s.teams = teams.NewServiceWithRepository(s.revenue, teamRepo)
	}
}

func (s *Server) Register(mux *http.ServeMux) {
	handle := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, httpx.RequestID(handler))
	}
	handle("POST /api/app/invites/precheck", s.invitePrecheck)
	handle("POST /api/app/invites/entries", s.AppAuthMiddleware(s.createInviteEntry))
	handle("POST /api/app/auth/wechat-login", s.wechatLogin)
	handle("POST /api/app/auth/phone-login", s.phoneLogin)
	handle("POST /api/app/auth/issue-token-after-identity", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.issueTokenAfterIdentity)))
	handle("GET /api/app/home", s.AppAuthMiddleware(s.appHome))
	handle("GET /api/app/newbie-tasks", s.AppAuthMiddleware(s.newbieTasks))
	handle("GET /api/app/users/me", s.AppAuthMiddleware(s.currentUser))
	handle("GET /api/app/users/me/summary", s.AppAuthMiddleware(s.currentUserSummary))
	handle("GET /api/app/profile/home", s.AppAuthMiddleware(s.profileHome))
	handle("GET /api/app/profile/assets", s.AppAuthMiddleware(s.profileAssets))
	handle("GET /api/app/profile/credit-center", s.AppAuthMiddleware(s.profileCreditCenter))
	handle("PUT /api/app/users/me/profile", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.updateCurrentUserProfile)))
	handle("POST /api/app/identity/phone/bind", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.bindPhone)))
	handle("POST /api/app/sms/send-code", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.sendSMSCode)))
	handle("POST /api/app/sms/verify-code", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.verifySMSCode)))
	handle("POST /api/app/identity/phone/verify", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.verifyPhone)))
	handle("POST /api/app/identity/realname/restart", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.restartRealname)))
	handle("POST /api/app/identity/faceid/detect-auth", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.startFaceID)))
	handle("POST /api/app/identity/faceid/callback", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.faceIDCallback)))
	handle("POST /api/app/identity/faceid/result", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.faceIDCallback)))
	handle("GET /api/app/identity/status", s.AppAuthMiddleware(s.identityStatus))
	handle("GET /api/app/im/ws", s.imSocket)
	handle("POST /api/app/games", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createGame)))
	handle("GET /api/app/games", s.AppAuthMiddleware(s.listGames))
	handle("GET /api/app/games/my/manage", s.AppAuthMiddleware(s.myManagedGames))
	handle("GET /api/app/games/player/manage", s.AppAuthMiddleware(s.myPlayerGames))
	handle("GET /api/app/games/profit-templates", s.AppAuthMiddleware(s.appRevenueTemplates))
	handle("GET /api/app/games/category-config", s.AppAuthMiddleware(s.gameCategoryConfigHandler))
	handle("GET /api/app/games/application-config", s.AppAuthMiddleware(s.gameApplicationConfig))
	handle("GET /api/app/games/condition-rule-config", s.AppAuthMiddleware(s.gameConditionRuleConfig))
	handle("GET /api/app/games/cancel-config", s.AppAuthMiddleware(s.gameCancelConfig))
	handle("GET /api/app/games/applications/my", s.AppAuthMiddleware(s.myApplications))
	handle("GET /api/app/game-applications/my", s.AppAuthMiddleware(s.myApplications))
	handle("GET /api/app/game-applications/received", s.AppAuthMiddleware(s.receivedApplications))
	handle("GET /api/app/games/favorites/my", s.AppAuthMiddleware(s.myFavoriteGames))
	handle("GET /api/app/game-invites/player-config", s.AppAuthMiddleware(s.gameInvitePlayerConfig))
	handle("GET /api/app/game-invites/permission", s.AppAuthMiddleware(s.gameInvitePermission))
	handle("GET /api/app/game-invites/recent-players", s.AppAuthMiddleware(s.gameInviteRecentPlayers))
	handle("GET /api/app/game-invites/players", s.AppAuthMiddleware(s.gameInvitePlayers))
	handle("GET /api/app/game-invites/replay-context", s.AppAuthMiddleware(s.gameInviteReplayContext))
	handle("GET /api/app/game-invites/system-recommendations", s.AppAuthMiddleware(s.gameInviteSystemRecommendations))
	handle("POST /api/app/game-invites/current", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createCurrentGameInvite)))
	handle("POST /api/app/game-invites/replay", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createReplayGameInvite)))
	handle("POST /api/app/game-invites/reminders", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createGameInviteReminder)))
	handle("GET /api/app/game-invites/guide-progress", s.AppAuthMiddleware(s.gameInviteGuideProgress))
	handle("GET /api/app/game-invites/guide-cancel-detail", s.AppAuthMiddleware(s.gameInviteCancelDetail))
	handle("GET /api/app/game-invites/referral-records", s.AppAuthMiddleware(s.gameInviteReferralRecords))
	handle("GET /api/app/games/city", s.AppAuthMiddleware(s.sameCityGames))
	handle("GET /api/app/games/nearby", s.AppAuthMiddleware(s.nearbyGames))
	handle("GET /api/app/map/search", s.AppAuthMiddleware(s.mapSearch))
	handle("GET /api/app/map/geocode", s.AppAuthMiddleware(s.mapGeocode))
	handle("GET /api/app/map/reverse-geocode", s.AppAuthMiddleware(s.mapReverseGeocode))
	handle("GET /api/app/map/route", s.AppAuthMiddleware(s.mapRoute))
	handle("GET /api/app/map/index-config", s.AppAuthMiddleware(s.mapIndexConfig))
	handle("GET /api/app/map/my-city", s.AppAuthMiddleware(s.mapMyCity))
	handle("GET /api/app/map/play-pages", s.AppAuthMiddleware(s.mapPlayPages))
	handle("POST /api/app/map/checkins", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.submitMapCheckin)))
	handle("POST /api/app/map/blind-routes", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createMapBlindRoute)))
	handle("POST /api/app/map/blind-routes/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeMapBlindRoutePost)))
	handle("POST /api/app/map/challenges", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createMapChallenge)))
	handle("POST /api/app/games/invitations/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.respondGameInvitation)))
	handle("POST /api/app/game-invitations/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.respondGameInvitation)))
	handle("GET /api/app/games/", s.AppAuthMiddleware(s.routeGameGet))
	handle("PUT /api/app/games/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeGamePut)))
	handle("GET /api/app/chat/rooms/", s.AppAuthMiddleware(s.routeChatRoomGet))
	handle("POST /api/app/chat/rooms/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeChatRoomPost)))
	handle("POST /api/app/files/upload-token", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createUploadToken)))
	handle("GET /api/app/files/", s.AppAuthMiddleware(s.downloadFileURL))
	handle("GET /api/app/reviews/todos", s.AppAuthMiddleware(s.reviewTodos))
	handle("GET /api/app/reviews/available", s.AppAuthMiddleware(s.reviewTodos))
	handle("GET /api/app/reviews/complete-config", s.AppAuthMiddleware(s.reviewCompleteConfig))
	handle("POST /api/app/reviews", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.submitReview)))
	handle("GET /api/app/reviews/my-intents", s.AppAuthMiddleware(s.myReviewIntents))
	handle("GET /api/app/users/me/growth", s.AppAuthMiddleware(s.reviewProfile))
	handle("GET /api/app/growth/my", s.AppAuthMiddleware(s.reviewProfile))
	handle("GET /api/app/footprints/my", s.AppAuthMiddleware(s.myFootprints))
	handle("GET /api/app/profile/system-management/profile-info", s.AppAuthMiddleware(s.getSystemProfileInfo))
	handle("PUT /api/app/profile/system-management/profile-info", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.saveSystemProfileInfo)))
	handle("GET /api/app/profile/system-management/skill-config", s.AppAuthMiddleware(s.getSystemSkillConfig))
	handle("GET /api/app/profile/system-management/service-cases/", s.AppAuthMiddleware(s.getSystemServiceCaseDetail))
	handle("PUT /api/app/profile/system-management/skill-config", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.saveSystemSkillConfig)))
	handle("GET /api/app/profile/system-management/feedback", s.AppAuthMiddleware(s.getSystemFeedbackHome))
	handle("POST /api/app/profile/system-management/feedback", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.submitSystemFeedback)))
	handle("GET /api/app/profile/system-management/feedback-records", s.AppAuthMiddleware(s.getSystemFeedbackRecords))
	handle("GET /api/app/profile/system-management/feedback-records/", s.AppAuthMiddleware(s.getSystemFeedbackDetail))
	handle("POST /api/app/profile/system-management/feedback-records/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.appendSystemFeedbackMessage)))
	handle("GET /api/app/profile/system-management/block-settings", s.AppAuthMiddleware(s.getSystemBlockSettings))
	handle("PUT /api/app/profile/system-management/block-settings", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.saveSystemBlockSettings)))
	handle("GET /api/app/profile/settings", s.AppAuthMiddleware(s.getProfileSettings))
	handle("PUT /api/app/profile/settings", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.saveProfileSettings)))
	handle("GET /api/app/profile/agreements", s.AppAuthMiddleware(s.listProfileAgreements))
	handle("GET /api/app/profile/agreements/", s.AppAuthMiddleware(s.routeProfileAgreementGet))
	handle("POST /api/app/profile/agreements/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeProfileAgreementPost)))
	handle("GET /api/app/profile/service-center/reviews", s.AppAuthMiddleware(s.profileReviewList))
	handle("GET /api/app/profile/service-center/reviews/", s.AppAuthMiddleware(s.profileReviewDetail))
	handle("POST /api/app/profile/service-center/reviews/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeProfileReviewPost)))
	handle("GET /api/app/profile/service-center/invite/overview", s.AppAuthMiddleware(s.profileInviteOverview))
	handle("GET /api/app/profile/service-center/invite/network", s.AppAuthMiddleware(s.profileInviteNetwork))
	handle("GET /api/app/profile/service-center/invite/records", s.AppAuthMiddleware(s.profileInviteRecords))
	handle("GET /api/app/profile/service-center/invite/ranking", s.AppAuthMiddleware(s.profileInviteRanking))
	handle("GET /api/app/profile/service-center/invite/income", s.AppAuthMiddleware(s.profileInviteIncome))
	handle("GET /api/app/profile/service-center/invite/member-detail", s.AppAuthMiddleware(s.profileInviteMemberDetail))
	handle("GET /api/app/incomes/account", s.AppAuthMiddleware(s.incomeAccount))
	handle("GET /api/app/incomes/summary", s.AppAuthMiddleware(s.incomeSummary))
	handle("GET /api/app/incomes/logs", s.AppAuthMiddleware(s.incomeLogs))
	handle("GET /api/app/income/account", s.AppAuthMiddleware(s.incomeAccount))
	handle("GET /api/app/income/summary", s.AppAuthMiddleware(s.incomeSummary))
	handle("GET /api/app/income/logs", s.AppAuthMiddleware(s.incomeLogs))
	handle("POST /api/app/revenues/simulate", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.appRevenueSimulate)))
	handle("GET /api/app/membership/plans", s.AppAuthMiddleware(s.membershipPlans))
	handle("GET /api/app/membership/radar-config", s.AppAuthMiddleware(s.membershipRadarConfig))
	handle("POST /api/app/membership/radar/actions", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.membershipRadarAction)))
	handle("GET /api/app/membership/my", s.AppAuthMiddleware(s.membershipMy))
	handle("GET /api/app/orders/", s.AppAuthMiddleware(s.appOrderDetail))
	handle("POST /api/app/payment/precreate-placeholder", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.paymentPrecreatePlaceholder)))
	handle("POST /api/app/guides/payment/precreate-placeholder", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.guidePaymentPrecreatePlaceholder)))
	handle("GET /api/app/points/summary", s.AppAuthMiddleware(s.pointsSummary))
	handle("GET /api/app/points/logs", s.AppAuthMiddleware(s.pointsLogs))
	handle("GET /api/app/redemption/items", s.AppAuthMiddleware(s.redemptionItems))
	handle("POST /api/app/redemption/orders", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createRedemptionOrder)))
	handle("GET /api/app/redemption/orders/my", s.AppAuthMiddleware(s.myRedemptionOrders))
	handle("GET /api/app/redemption/orders/", s.AppAuthMiddleware(s.redemptionOrderDetail))
	handle("POST /api/app/redemption/orders/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeRedemptionOrderPost)))
	handle("GET /api/app/profile/points/orders/", s.AppAuthMiddleware(s.redemptionOrderLogistics))
	handle("POST /api/app/behavior/events", s.IdempotencyMiddleware(s.createBehaviorEvent))
	handle("GET /api/app/connections/my", s.AppAuthMiddleware(s.myConnections))
	handle("POST /api/app/connections/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeConnectionPost)))
	handle("GET /api/app/experts/me/skills", s.AppAuthMiddleware(s.expertSkill))
	handle("POST /api/app/experts/me/skills", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.updateExpertSkill)))
	handle("PUT /api/app/experts/me/skills", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.updateExpertSkill)))
	handle("GET /api/app/guides/me/resources", s.AppAuthMiddleware(s.guideResource))
	handle("POST /api/app/guides/me/resources", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.updateGuideResource)))
	handle("PUT /api/app/guides/me/resources", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.updateGuideResource)))
	handle("GET /api/app/roles/my", s.AppAuthMiddleware(s.myRoles))
	handle("GET /api/app/role-applications/expert/config", s.AppAuthMiddleware(s.expertApplyConfig))
	handle("GET /api/app/role-applications/guide/config", s.AppAuthMiddleware(s.guideApplyConfig))
	handle("GET /api/app/role-applications/status-config", s.AppAuthMiddleware(s.roleStatusPageConfig))
	handle("GET /api/app/role-applications/benefit-config", s.AppAuthMiddleware(s.roleBenefitConfig))
	handle("POST /api/app/role-applications", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.submitRoleApplication)))
	handle("GET /api/app/role-applications/my", s.AppAuthMiddleware(s.myRoleApplications))
	handle("GET /api/app/guides/qualification/me", s.AppAuthMiddleware(s.guideQualificationMe))
	handle("POST /api/app/guides/apply", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.applyGuide)))
	handle("GET /api/app/member-reports/me", s.AppAuthMiddleware(s.memberReportMe))
	handle("GET /api/app/teams/my", s.AppAuthMiddleware(s.myTeam))
	handle("GET /api/app/teams/my/members", s.AppAuthMiddleware(s.myTeamMembers))
	handle("GET /api/app/teams/my/revenue-summary", s.AppAuthMiddleware(s.myTeamRevenueSummary))
	handle("GET /api/app/notifications", s.AppAuthMiddleware(s.myNotifications))
	handle("GET /api/app/messages/center", s.AppAuthMiddleware(s.myNotifications))
	handle("GET /api/app/messages/my-config", s.AppAuthMiddleware(s.messageMyConfig))
	handle("GET /api/app/private-chat/messages", s.AppAuthMiddleware(s.privateChatMessages))
	handle("POST /api/app/private-chat/messages", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.sendPrivateChatMessage)))
	handle("GET /api/app/messages/trade-warning", s.AppAuthMiddleware(s.tradeWarningDetail))
	handle("POST /api/app/messages/trade-warning/actions", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.handleTradeWarningAction)))
	handle("GET /api/app/messages/system-notification", s.AppAuthMiddleware(s.systemNotificationDetail))
	handle("POST /api/app/messages/system-notification/feedback", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.submitSystemNotificationFeedback)))
	handle("POST /api/app/notifications/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeNotificationPost)))
	handle("GET /api/app/reports/config", s.AppAuthMiddleware(s.reportConfig))
	handle("GET /api/app/reports/my", s.AppAuthMiddleware(s.myReports))
	handle("GET /api/app/reports/appeals/my", s.AppAuthMiddleware(s.myAppeals))
	handle("POST /api/app/profile/credit-appeals", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createCreditAppeal)))
	handle("GET /api/app/reports/", s.AppAuthMiddleware(s.myReportDetail))
	handle("POST /api/app/reports", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.createReport)))
	handle("POST /api/app/reports/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeAppReportPost)))
	handle("POST /api/admin/auth/login", s.adminLogin)
	handle("POST /api/admin/auth/applications", s.submitAdminApplication)
	handle("GET /api/admin/auth/permissions", s.adminPermissions)
	handle("GET /api/admin/permissions/tree", s.adminPermissions)
	handle("GET /api/admin/admin-users", s.requireAdminPermission("admin_user:view", s.adminAccountUsers))
	handle("GET /api/admin/admin-applications", s.requireAdminPermission("admin_user:view", s.adminApplications))
	handle("POST /api/admin/admin-users", s.requireAdminPermission("admin_user:create", s.createAdminAccountUser))
	handle("PUT /api/admin/admin-users/", s.requireAdminPermission("admin_user:update", s.routeAdminAccountUserPut))
	handle("GET /api/admin/admin-roles", s.requireAdminPermission("admin_user:view", s.adminRoles))
	handle("GET /api/admin/admin-permissions/catalog", s.requireAdminPermission("admin_user:view", s.adminPermissionCatalog))
	handle("GET /api/admin/invite-codes", s.requireAdminPermission("invite_code:read", s.adminInviteCodes))
	handle("POST /api/admin/invite-codes", s.requireAdminPermission("invite_code:manage", s.createAdminInviteCode))
	handle("GET /api/admin/invite-codes/", s.requireAdminPermission("invite_code:read", s.adminInviteCodeDetail))
	handle("POST /api/admin/invite-codes/", s.routeAdminInviteCodePost)
	handle("GET /api/admin/invite-relations", s.requireAdminPermission("invite_code:read", s.adminInviteRelations))
	handle("GET /api/admin/users", s.requireAdminPermission("user:view", s.adminUsers))
	handle("GET /api/admin/users/options", s.requireAdminPermission("game:create_admin", s.adminUserOptions))
	handle("PUT /api/admin/users/", s.routeAdminUsersPut)
	handle("GET /api/admin/users/", s.routeAdminUsersGet)
	handle("GET /api/admin/games", s.requireAdminPermission("game:read", s.adminGames))
	handle("GET /api/admin/game-applications", s.requireAdminPermission("game:read", s.adminGameApplications))
	handle("POST /api/admin/games", s.requireAdminPermission("game:create_admin", s.adminCreateGame))
	handle("POST /api/admin/games/batch-audit", s.requireAdminPermission("game:update_status", s.adminBatchAuditGames))
	handle("GET /api/admin/games/category-config", s.requireAdminPermission("system_config:read", s.adminGameCategoryConfig))
	handle("PUT /api/admin/games/category-config", s.requireAdminPermission("system_config:update", s.adminGameCategoryConfig))
	handle("GET /api/admin/games/application-config", s.requireAdminPermission("system_config:read", s.adminGameApplicationConfig))
	handle("PUT /api/admin/games/application-config", s.requireAdminPermission("system_config:update", s.adminGameApplicationConfig))
	handle("GET /api/admin/games/audit-config", s.requireAdminPermission("system_config:read", s.adminGameAuditConfig))
	handle("PUT /api/admin/games/audit-config", s.requireAdminPermission("system_config:update", s.adminGameAuditConfig))
	handle("GET /api/admin/games/condition-rule-config", s.requireAdminPermission("system_config:read", s.adminGameConditionRuleConfig))
	handle("PUT /api/admin/games/condition-rule-config", s.requireAdminPermission("system_config:update", s.adminGameConditionRuleConfig))
	handle("GET /api/admin/roles/benefit-config", s.requireAdminPermission("system_config:read", s.adminRoleBenefitConfig))
	handle("PUT /api/admin/roles/benefit-config", s.requireAdminPermission("system_config:update", s.adminRoleBenefitConfig))
	handle("POST /api/admin/roles/grant", s.requireAdminPermission("role:update", s.adminGrantRole))
	handle("GET /api/admin/home/display-config", s.requireAdminPermission("system_config:read", s.adminHomeDisplayConfig))
	handle("PUT /api/admin/home/display-config", s.requireAdminPermission("system_config:update", s.adminHomeDisplayConfig))
	handle("GET /api/admin/reviews/complete-config", s.requireAdminPermission("system_config:read", s.adminReviewCompleteConfig))
	handle("PUT /api/admin/reviews/complete-config", s.requireAdminPermission("system_config:update", s.adminReviewCompleteConfig))
	handle("GET /api/admin/credit-deduction-rules", s.requireAdminPermission("system_config:read", s.adminCreditDeductionRules))
	handle("PUT /api/admin/credit-deduction-rules", s.requireAdminPermission("system_config:update", s.adminCreditDeductionRules))
	handle("GET /api/admin/games/", s.routeAdminGamesGet)
	handle("POST /api/admin/games/", s.routeAdminGamesPost)
	handle("POST /api/admin/game-checkins/", s.requireAdminPermission("game:progress:manage", s.routeAdminGameCheckinPost))
	handle("GET /api/admin/connections", s.requireAdminPermission("connection:read", s.adminConnections))
	handle("GET /api/admin/profile-users", s.requireAdminPermission("profile:read", s.adminProfileUsers))
	handle("GET /api/admin/experts/", s.requireAdminPermission("profile:read", s.adminExpertSkill))
	handle("GET /api/admin/guides/", s.requireAdminPermission("profile:read", s.adminGuideResource))
	handle("GET /api/admin/points/logs", s.requireAdminPermission("points:read", s.adminPointsLogs))
	handle("GET /api/admin/redemption/items", s.requireAdminPermission("redemption:manage", s.adminRedemptionItems))
	handle("POST /api/admin/redemption/items", s.requireAdminPermission("redemption:manage", s.createAdminRedemptionItem))
	handle("PUT /api/admin/redemption/items/", s.requireAdminPermission("redemption:manage", s.updateAdminRedemptionItem))
	handle("GET /api/admin/redemption/orders", s.requireAdminPermission("redemption:manage", s.adminRedemptionOrders))
	handle("POST /api/admin/redemption/orders/", s.routeAdminRedemptionOrderPost)
	handle("GET /api/admin/teams", s.requireAdminPermission("team:read", s.adminTeams))
	handle("GET /api/admin/teams/", s.requireAdminPermission("team:read", s.adminTeamDetail))
	handle("GET /api/admin/member-reports", s.requireAdminPermission("member_report:read", s.adminMemberReports))
	handle("GET /api/admin/identity-verifications", s.requireAdminPermission("identity:read", s.adminIdentityVerifications))
	handle("GET /api/admin/identity-verifications/", s.requireAdminPermission("identity:read", s.adminIdentityVerificationDetail))
	handle("POST /api/admin/identity-verifications/", s.routeAdminIdentityVerificationPost)
	handle("GET /api/admin/avatar-audits", s.requireAdminPermission("identity:read", s.adminAvatarAudits))
	handle("POST /api/admin/avatar-audits/", s.routeAdminAvatarAuditPost)
	handle("POST /api/admin/files/upload-token", s.requireAdminPermission("game:create_admin", s.adminCreateUploadToken))
	handle("GET /api/admin/files/", s.adminDownloadFileURL)
	handle("GET /api/admin/delivery-documents", s.requireAdminPermission("delivery:manage", s.adminDeliveryDocuments))
	handle("POST /api/admin/delivery-documents", s.requireAdminPermission("delivery:manage", s.createAdminDeliveryDocument))
	handle("GET /api/admin/audits/role-applications", s.requireAdminPermission("role:view", s.adminRoleApplications))
	handle("POST /api/admin/audits/role-applications/", s.routeAdminRoleApplicationPost)
	handle("GET /api/admin/guides/qualification-rules", s.requireAdminPermission("role:view", s.adminGuideQualificationRules))
	handle("PUT /api/admin/guides/qualification-rules/", s.routeAdminGuideQualificationRulePut)
	handle("GET /api/admin/guide-qualification-rules", s.requireAdminPermission("role:view", s.adminGuideQualification))
	handle("POST /api/admin/guide-qualification-rules", s.requireAdminPermission("role:update", s.updateGuideQualification))
	handle("GET /api/admin/test-cases", s.requireAdminPermission("testcase:read", s.adminTestCases))
	handle("POST /api/admin/test-cases", s.requireAdminPermission("testcase:manage", s.createAdminTestCase))
	handle("GET /api/admin/test-runs", s.requireAdminPermission("testcase:read", s.adminTestRuns))
	handle("POST /api/admin/test-runs", s.requireAdminPermission("testcase:manage", s.createAdminTestRun))
	handle("GET /api/admin/im/rooms", s.requireAdminPermission("im:room:read", s.adminIMRooms))
	handle("GET /api/admin/im/rooms/", s.routeAdminIMRoomGet)
	handle("POST /api/admin/im/rooms/", s.routeAdminIMRoomPost)
	handle("POST /api/admin/im/messages/", s.routeAdminIMMessagePost)
	handle("GET /api/admin/sensitive-words", s.requireAdminPermission("content:sensitive_word:view", s.adminSensitiveWords))
	handle("POST /api/admin/sensitive-words", s.requireAdminPermission("content:sensitive_word:create", s.createAdminSensitiveWord))
	handle("PUT /api/admin/sensitive-words/", s.requireAdminPermission("content:sensitive_word:update", s.updateAdminSensitiveWord))
	handle("POST /api/admin/sensitive-words/import", s.requireAdminPermission("content:sensitive_word:import", s.importAdminSensitiveWords))
	handle("GET /api/admin/content-risk/logs", s.requireAdminPermission("content:risk_log:view", s.adminContentRiskLogs))
	handle("GET /api/admin/system/readiness", s.requireAdminPermission("system_config:read", s.adminSystemReadiness))
	handle("GET /api/admin/dashboard", s.requireAdminPermission("analytics:funnel:view", s.adminDashboard))
	handle("GET /api/admin/analytics/funnel", s.requireAdminPermission("analytics:funnel:view", s.adminFunnel))
	handle("GET /api/admin/analytics/retention", s.requireAdminPermission("analytics:retention:view", s.adminRetention))
	handle("GET /api/admin/behavior/events", s.requireAdminPermission("data:behavior:read", s.adminBehaviorEvents))
	handle("GET /api/admin/ai-data/snapshot", s.requireAdminPermission("ai:data:read", s.adminAIDataSnapshot))
	handle("POST /api/admin/ai-data/acceptance-fixture", s.requireAdminPermission("ai:data:seed", s.adminAIDataAcceptanceFixture))
	handle("POST /api/admin/ai-data/im-export", s.requireAdminPermission("ai:data:export", s.adminAIDataIMExport))
	handle("GET /api/admin/ai-data/im-export-config", s.requireAdminPermission("ai:data:read", s.adminAIDataIMExportConfig))
	handle("PUT /api/admin/ai-data/im-export-config", s.requireAdminPermission("system_config:update", s.adminAIDataIMExportConfig))
	handle("GET /api/admin/behavior-logs", s.requireAdminPermission("analytics:timeline:view", s.adminBehaviorLogs))
	handle("GET /api/admin/operation-logs", s.adminOperationLogs)
	handle("GET /api/admin/export-tasks", s.requireAdminPermission("report_export:create", s.adminExportTasks))
	handle("GET /api/admin/export-tasks/", s.requireAdminPermission("report_export:create", s.adminExportTaskDownloadURL))
	handle("GET /api/admin/notifications/wechat-tasks", s.requireAdminPermission("notification:wechat:view", s.adminWechatSubscribeTasks))
	handle("GET /api/admin/notifications/wechat-templates", s.requireAdminPermission("notification:wechat:view", s.adminWechatSubscribeTemplates))
	handle("GET /api/admin/reports", s.requireAdminPermission("report:view", s.adminReports))
	handle("GET /api/admin/feedback-records", s.requireAdminPermission("feedback:view", s.adminSystemFeedbackRecords))
	handle("GET /api/admin/reports/config", s.requireAdminPermission("system_config:read", s.adminReportConfig))
	handle("PUT /api/admin/reports/config", s.requireAdminPermission("system_config:update", s.adminReportConfig))
	handle("POST /api/admin/reports/batch-handle", s.batchHandleReports)
	handle("GET /api/admin/reports/export-templates", s.requireAdminPermission("report_export:create", s.adminExportTemplates))
	handle("POST /api/admin/reports/export", s.requireAdminPermission("report_export:create", s.createExportTask))
	handle("POST /api/admin/feedback-records/", s.requireAdminPermission("feedback:reply", s.adminSystemFeedbackReply))
	handle("GET /api/admin/reports/", s.requireAdminPermission("report:view", s.adminReportDetail))
	handle("POST /api/admin/reports/", s.routeAdminReportPost)
	handle("GET /api/admin/revenue/templates", s.requireAdminPermission("revenue:template:view", s.revenueTemplates))
	handle("POST /api/admin/revenue/templates", s.requireAdminPermission("revenue:template:update", s.createRevenueTemplate))
	handle("GET /api/admin/revenue/profit-template-config", s.requireAdminPermission("system_config:read", s.adminProfitTemplateConfig))
	handle("PUT /api/admin/revenue/profit-template-config", s.requireAdminPermission("system_config:update", s.adminProfitTemplateConfig))
	handle("GET /api/admin/revenue/rules", s.requireAdminPermission("revenue:template:view", s.revenueRules))
	handle("POST /api/admin/revenue/rules", s.requireAdminPermission("revenue:template:update", s.upsertRevenueRule))
	handle("POST /api/admin/revenue/preview", s.requireAdminPermission("revenue:simulate", s.adminRevenuePreview))
	handle("POST /api/admin/revenue/calculate", s.requireAdminPermission("revenue:simulate", s.adminRevenuePreview))
	handle("POST /api/admin/revenue/records/generate", s.requireAdminPermission("revenue:generate", s.generateRevenueRecord))
	handle("GET /api/admin/revenue/records", s.requireAdminPermission("revenue:record:view", s.revenueRecords))
	handle("GET /api/admin/revenue/settlements", s.requireAdminPermission("settlement:offline:create", s.revenueSettlements))
	handle("POST /api/admin/revenue/records/", s.routeRevenueRecordPost)
	handle("POST /api/internal/openim/webhooks", s.openIMWebhook)
	handle("POST /api/internal/im/archive-expired-rooms", s.archiveExpiredIMRooms)
	handle("POST /api/internal/ai/content-risk/check-placeholder", s.aiContentRiskPlaceholder)
	handle("POST /api/internal/pay/callback-placeholder", s.paymentCallbackPlaceholder)
	handle("POST /api/funds/profit-sharing/orders", s.profitSharingOrderPlaceholder)
	handle("GET /api/funds/profit-sharing/orders/", s.profitSharingOrderQueryPlaceholder)
	handle("POST /api/funds/profit-sharing/return-orders", s.profitSharingReturnPlaceholder)
	handle("POST /api/internal/notifications/wechat-tasks/", s.routeWechatTaskPost)
	handle("POST /api/internal/jobs/review-remind", s.runReviewRemindJob)
	handle("POST /api/internal/jobs/progress-feedback-remind", s.runProgressFeedbackRemindJob)
	handle("POST /api/internal/reports/export-runner", s.requireAdminPermission("report_export:create", s.runExportTasks))
	handle("POST /api/app/games/applications/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.reviewApplication)))
	handle("POST /api/app/game-applications/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeGameApplicationPost)))
	handle("POST /api/app/games/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeGamePost)))
	handle("DELETE /api/app/games/", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.routeGameDelete)))
	handle("POST /api/app/locations/current", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.saveCurrentLocation)))
	handle("POST /api/app/locations/manual", s.AppAuthMiddleware(s.IdempotencyMiddleware(s.saveManualLocation)))
	handle("GET /api/app/locations/my-recent", s.AppAuthMiddleware(s.myRecentLocations))
	handle("GET /api/admin/map/search", s.requireAdminPermission("game:create_admin", s.mapSearch))
}

func (s *Server) routeGameGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/progress-feedbacks"):
		s.progressFeedbacks(w, r)
	case strings.HasSuffix(path, "/members"):
		s.gameMembers(w, r)
	case strings.HasSuffix(path, "/milestones"):
		s.milestones(w, r)
	case strings.HasSuffix(path, "/checkins"):
		s.checkins(w, r)
	case strings.HasSuffix(path, "/retrospectives"):
		s.retrospectives(w, r)
	case strings.HasSuffix(path, "/chat-room"):
		s.chatRoom(w, r)
	case strings.HasSuffix(path, "/chat-session"):
		s.chatSession(w, r)
	case strings.HasSuffix(path, "/chat/messages"):
		s.historyMessages(w, r)
	case strings.HasSuffix(path, "/revenue-preview"):
		s.appRevenuePreview(w, r)
	case strings.HasSuffix(path, "/success-detail"):
		s.gameSuccessDetail(w, r)
	case strings.HasSuffix(path, "/guide-success-detail"):
		s.gameGuideSuccessDetail(w, r)
	case strings.HasSuffix(path, "/collaboration"):
		s.gameCollaboration(w, r)
	case strings.HasSuffix(path, "/player-cancel-detail"):
		s.gameCancelDetail(w, r, "player")
	case strings.HasSuffix(path, "/expert-cancel-detail"):
		s.gameCancelDetail(w, r, "expert")
	case strings.HasSuffix(path, "/payment-preview"):
		s.gamePaymentPreview(w, r)
	default:
		s.getGame(w, r)
	}
}

func (s *Server) routeChatRoomGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/messages"):
		s.historyMessagesByRoom(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeChatRoomPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/messages"):
		s.sendMessageToRoom(w, r)
	case strings.HasSuffix(path, "/ack"):
		s.ackMessage(w, r)
	case strings.HasSuffix(path, "/read"):
		s.readMessage(w, r)
	case strings.HasSuffix(path, "/archive"):
		s.archiveRoom(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeGamePost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/approve-local"):
		s.approveGameForLocal(w, r)
	case strings.HasSuffix(path, "/applications"):
		s.applyGame(w, r)
	case strings.HasSuffix(path, "/guide-invitations"):
		s.createGameInvitation(w, r)
	case strings.HasSuffix(path, "/manual-start"):
		s.manualStart(w, r)
	case strings.HasSuffix(path, "/exit"):
		s.exitGame(w, r)
	case strings.HasSuffix(path, "/chat/messages"):
		s.sendMessage(w, r)
	case strings.HasSuffix(path, "/service-confirm"):
		s.serviceConfirm(w, r)
	case strings.HasSuffix(path, "/completion-request"):
		s.requestGameCompletion(w, r)
	case strings.HasSuffix(path, "/service-confirm-items"):
		s.serviceConfirmItem(w, r)
	case strings.HasSuffix(path, "/player-cancel"):
		s.playerCancelRequest(w, r)
	case strings.HasSuffix(path, "/expert-cancel"):
		s.expertCancelRequest(w, r)
	case strings.HasSuffix(path, "/progress-feedbacks"):
		s.createProgressFeedback(w, r)
	case strings.HasSuffix(path, "/milestones"):
		s.createMilestone(w, r)
	case strings.HasSuffix(path, "/checkins"):
		s.createCheckin(w, r)
	case strings.HasSuffix(path, "/retrospectives"):
		s.createRetrospective(w, r)
	case strings.HasSuffix(path, "/continue"):
		s.continueGame(w, r)
	case strings.HasSuffix(path, "/favorite"):
		s.favoriteGame(w, r)
	case strings.HasSuffix(path, "/guide-follow-ups"):
		s.createGuideFollowUp(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeGameApplicationPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/audit"):
		s.reviewApplication(w, r)
	case strings.HasSuffix(path, "/cancel"):
		s.cancelApplication(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeGamePut(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.Contains(path, "/milestones/"):
		s.updateMilestone(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeGameDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/favorite"):
		s.unfavoriteGame(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeAdminGameCheckinPost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/mark-invalid") {
		s.markCheckinInvalid(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) routeAdminUsersGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/growth"):
		s.requireAdminPermission("user:view", s.adminUserGrowth)(w, r)
	case strings.HasSuffix(path, "/favorites"):
		s.requireAdminPermission("user:read", s.adminUserFavorites)(w, r)
	case strings.HasSuffix(path, "/connections"):
		s.requireAdminPermission("connection:read", s.adminUserConnections)(w, r)
	default:
		s.requireAdminPermission("user:view", s.adminUserDetail)(w, r)
	}
}

func (s *Server) routeAdminUsersPut(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/invite-relation"):
		s.requireAdminPermission("invite_code:manage", s.adminUpdateUserInviteRelation)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) adminUsers(w http.ResponseWriter, r *http.Request) {
	filter := users.Filter{
		Status:         r.URL.Query().Get("status"),
		RealnameStatus: r.URL.Query().Get("realnameStatus"),
		Keyword:        r.URL.Query().Get("keyword"),
	}
	items, err := s.auth.AdminUsers(filter)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list users failed")
		return
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		items, err = s.appendIdentityMatchedAdminUsers(filter, items, keyword)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list users failed")
			return
		}
	}
	relations, err := s.auth.AdminInviteRelations(invites.RelationFilter{})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list invite relations failed")
		return
	}
	relationByInvitee := make(map[int64]invites.Relation, len(relations))
	for _, relation := range relations {
		if relation.InviteeUserID <= 0 {
			continue
		}
		existing, exists := relationByInvitee[relation.InviteeUserID]
		if !exists || (existing.InviterUserID <= 0 && relation.InviterUserID > 0) || relation.InviteCodeID > existing.InviteCodeID {
			relationByInvitee[relation.InviteeUserID] = relation
		}
	}
	payload := make([]map[string]interface{}, 0, len(items))
	for _, user := range items {
		payload = append(payload, s.adminUserPayload(user, relationByInvitee[user.ID]))
	}
	httpx.OK(w, map[string]interface{}{
		"items": payload,
		"total": len(payload),
	})
}

func (s *Server) appendIdentityMatchedAdminUsers(filter users.Filter, items []users.User, keyword string) ([]users.User, error) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return items, nil
	}
	records := s.identity.AllRecords()
	if len(records) == 0 {
		return items, nil
	}
	recordByUserID := make(map[int64]identity.Record, len(records))
	for _, record := range records {
		recordByUserID[record.UserID] = record
	}
	seen := make(map[int64]bool, len(items))
	for _, user := range items {
		seen[user.ID] = true
	}
	filter.Keyword = ""
	candidates, err := s.auth.AdminUsers(filter)
	if err != nil {
		return nil, err
	}
	for _, user := range candidates {
		if seen[user.ID] {
			continue
		}
		if s.identityRecordMatchesKeyword(recordByUserID[user.ID], keyword) {
			items = append(items, user)
			seen[user.ID] = true
		}
	}
	return items, nil
}

func (s *Server) identityRecordMatchesKeyword(record identity.Record, keyword string) bool {
	if record.UserID <= 0 || keyword == "" {
		return false
	}
	if adminSearchContains(record.PhoneMasked, keyword) ||
		adminSearchContains(record.RealNameMasked, keyword) ||
		adminSearchContains(record.IDCardMasked, keyword) {
		return true
	}
	plain, err := s.identity.RevealRecord(record)
	if err != nil {
		return false
	}
	return adminSearchContains(plain.Phone, keyword) ||
		adminSearchContains(plain.RealName, keyword) ||
		adminSearchContains(plain.IDCard, keyword)
}

func adminSearchContains(value string, keyword string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value != "" && strings.Contains(value, keyword)
}

func (s *Server) adminUserOptions(w http.ResponseWriter, r *http.Request) {
	items, err := s.auth.AdminUsers(users.Filter{
		Status:  "active",
		Keyword: r.URL.Query().Get("keyword"),
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternalError, "list users failed")
		return
	}
	const maxOptions = 200
	options := make([]map[string]interface{}, 0, len(items))
	for index, user := range items {
		if index >= maxOptions {
			break
		}
		nickname := strings.TrimSpace(user.Nickname)
		label := "用户 " + strconv.FormatInt(user.ID, 10)
		if nickname != "" {
			label += " - " + nickname
		}
		options = append(options, map[string]interface{}{
			"id":       user.ID,
			"nickname": nickname,
			"label":    label,
			"status":   user.Status,
		})
	}
	httpx.OK(w, map[string]interface{}{
		"items": options,
		"total": len(options),
	})
}

func (s *Server) adminUserDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "")
	if !ok {
		return
	}
	user, found := s.auth.UserByID(userID)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "user not found")
		return
	}
	record := s.identity.Status(userID)
	relation, _, _ := s.auth.InviteRelationForUser(userID)
	inviter := s.adminUserSummary(relation.InviterUserID)
	httpx.OK(w, map[string]interface{}{
		"user":           s.adminUserPayload(user, relation),
		"identity":       record,
		"inviteRelation": relation,
		"inviter":        inviter,
		"growth":         s.reviews.Profile(userID),
		"roles":          s.profiles.RoleSnapshot(userID),
		"favorites":      s.games.FavoritesForUser(userID),
		"connections":    s.connections.My(userID),
		"incomeSummary":  s.revenue.IncomeSummary(userID),
	})
}

func (s *Server) adminUpdateUserInviteRelation(w http.ResponseWriter, r *http.Request) {
	userID, ok := idFromAdminPath(w, r.URL.Path, "/api/admin/users/", "/invite-relation")
	if !ok {
		return
	}
	var req struct {
		InviterUserID int64 `json:"inviterUserId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	if req.InviterUserID <= 0 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "inviter user id required")
		return
	}
	if req.InviterUserID == userID {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "inviter cannot be self")
		return
	}
	if _, found := s.auth.UserByID(userID); !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "user not found")
		return
	}
	inviter, found := s.auth.UserByID(req.InviterUserID)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "inviter not found")
		return
	}
	relation, err := s.auth.SetInviteRelationInviter(userID, inviter.ID, "admin_manual")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "update invite relation failed")
		return
	}
	s.recordOperation(r, "invite_relation:update", "user", strconv.FormatInt(userID, 10), map[string]interface{}{
		"userId":        userID,
		"inviterUserId": inviter.ID,
		"inviteCodeId":  relation.InviteCodeID,
		"bindSource":    relation.BindSource,
	})
	httpx.OK(w, map[string]interface{}{
		"inviteRelation": relation,
		"inviter": map[string]interface{}{
			"id":       inviter.ID,
			"nickname": inviter.Nickname,
			"label":    "用户 " + strconv.FormatInt(inviter.ID, 10) + inviteNicknameSuffix(inviter.Nickname),
		},
	})
}

func (s *Server) adminUserPayload(user users.User, relation invites.Relation) map[string]interface{} {
	inviteCode, _ := s.auth.InviteCodeForUser(user.ID)
	payload := map[string]interface{}{
		"id":             user.ID,
		"openId":         user.OpenID,
		"phoneMasked":    user.PhoneMasked,
		"nickname":       user.Nickname,
		"avatarUrl":      user.AvatarURL,
		"avatarFileId":   user.AvatarFileID,
		"realnameStatus": user.RealnameStatus,
		"status":         user.Status,
		"createdAt":      user.CreatedAt,
		"inviteCode":     inviteCode,
	}
	if relation.InviteeUserID > 0 {
		payload["inviteRelation"] = relation
		payload["inviterUserId"] = relation.InviterUserID
		payload["inviteCodeId"] = relation.InviteCodeID
		payload["bindSource"] = relation.BindSource
		if inviter, ok := s.auth.UserByID(relation.InviterUserID); ok {
			payload["inviter"] = map[string]interface{}{
				"id":       inviter.ID,
				"nickname": inviter.Nickname,
				"label":    "用户 " + strconv.FormatInt(inviter.ID, 10) + inviteNicknameSuffix(inviter.Nickname),
			}
			payload["inviterNickname"] = inviter.Nickname
		}
	}
	return payload
}

func inviteNicknameSuffix(nickname string) string {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return ""
	}
	return " - " + nickname
}

func (s *Server) routeAdminGamesGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/api/admin/games/":
		s.requireAdminPermission("game:read", s.adminGames)(w, r)
	case strings.HasSuffix(path, "/milestones"):
		s.requireAdminPermission("game:read", s.adminGameMilestones)(w, r)
	case strings.HasSuffix(path, "/checkins"):
		s.requireAdminPermission("game:read", s.adminGameCheckins)(w, r)
	case strings.HasSuffix(path, "/retrospectives"):
		s.requireAdminPermission("game:read", s.adminGameRetrospectives)(w, r)
	case strings.HasSuffix(path, "/continue-drafts"):
		s.requireAdminPermission("game:read", s.adminGameContinueDrafts)(w, r)
	case strings.HasSuffix(path, "/review-trace"):
		s.requireAdminPermission("game:view", s.adminGameReviewTrace)(w, r)
	default:
		s.requireAdminPermission("game:read", s.adminGameDetail)(w, r)
	}
}

func (s *Server) routeAdminGamesPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/api/admin/games/":
		s.requireAdminPermission("game:create_admin", s.adminCreateGame)(w, r)
	case strings.HasSuffix(path, "/audit"):
		s.requireAdminPermission("game:update_status", s.adminAuditGame)(w, r)
	case strings.HasSuffix(path, "/im-room"):
		s.requireAdminPermission("game:update_status", s.adminEnsureGameIMRoom)(w, r)
	case strings.HasSuffix(path, "/milestones"):
		s.requireAdminPermission("game:progress:manage", s.createAdminGameMilestone)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeRevenueRecordPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/freeze"):
		s.requireAdminPermission("revenue:freeze", s.freezeRevenueRecord)(w, r)
	case strings.HasSuffix(path, "/settle"):
		s.requireAdminPermission("settlement:offline:create", s.settleRevenueRecord)(w, r)
	case strings.HasSuffix(path, "/settle-offline"):
		s.requireAdminPermission("settlement:offline:create", s.settleRevenueRecord)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeAdminRedemptionOrderPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/review"):
		s.requireAdminPermission("redemption:manage", s.reviewAdminRedemptionOrder)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeRedemptionOrderPost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/cancel"):
		s.cancelRedemptionOrder(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeProfileAgreementGet(w http.ResponseWriter, r *http.Request) {
	s.getProfileAgreementDetail(w, r)
}

func (s *Server) routeProfileAgreementPost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/sign"):
		s.signProfileAgreement(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeAdminReportPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/assign"):
		s.requireAdminPermission("report:assign", s.assignReport)(w, r)
	case strings.HasSuffix(path, "/handle"):
		s.requireAdminPermission("report:handle", s.handleReport)(w, r)
	case strings.HasSuffix(path, "/close"):
		s.requireAdminPermission("report:close", s.closeReport)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) routeNotificationPost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/read"):
		s.markNotificationRead(w, r)
		return
	case strings.HasSuffix(r.URL.Path, "/actions"):
		s.handleNotificationAction(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) routeWechatTaskPost(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/send-pending") {
		s.requireAdminPermission("notification:wechat:view", s.sendPendingWechatSubscribeTasks)(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/send") {
		s.requireAdminPermission("notification:wechat:view", s.sendWechatSubscribeTask)(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/mark-sent") {
		s.requireAdminPermission("notification:wechat:view", s.markWechatSubscribeTaskSent)(w, r)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) invitePrecheck(w http.ResponseWriter, r *http.Request) {
	var req auth.InvitePrecheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	result, err := s.auth.InvitePrecheck(req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInviteRequired):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invite code required")
		case errors.Is(err, auth.ErrInvalidInvite):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invalid invite code")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "invite precheck failed")
		}
		return
	}
	httpx.OK(w, result)
}

func (s *Server) createInviteEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		EntryType string `json:"entryType"`
		Title     string `json:"title"`
		GameID    int64  `json:"gameId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	// 个人邀请码、海报和二维码只属于行家/领路人；组局内的成员邀请仍由组局权限控制。
	if req.GameID <= 0 {
		roles := s.profiles.RoleSnapshot(userID).RoleStatusMap
		if roles["expert"] != "approved" && roles["expert"] != "active" && roles["guide"] != "approved" && roles["guide"] != "active" {
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "仅行家或领路人可生成邀请入口")
			return
		}
	}
	invite, err := s.auth.IssueInviteEntry(userID, req.EntryType)
	if err != nil {
		if errors.Is(err, invites.ErrInvalidEntryType) {
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid entry type")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "create invite entry failed")
		return
	}
	title := strings.TrimSpace(req.Title)
	path := "/pages/login/invite/index?inviteCode=" + invite.Code + "&entryType=" + invite.EntryType
	scene := "inviteCode=" + invite.Code
	var gameInfo interface{}
	wxaCodeDataURL := ""
	if req.GameID > 0 {
		game, err := s.games.Get(req.GameID)
		if err != nil {
			httpx.Error(w, http.StatusNotFound, 40421, "局不存在")
			return
		}
		if title == "" {
			title = shareTitleForGame(game)
		}
		path = "/pages/game/detail/index?id=" + strconv.FormatInt(game.ID, 10) + "&inviteCode=" + invite.Code + "&entryType=" + invite.EntryType
		scene = "gameId=" + strconv.FormatInt(game.ID, 10) + "&inviteCode=" + invite.Code
		gameInfo = map[string]interface{}{
			"id":             game.ID,
			"title":          game.Title,
			"gameType":       game.GameType,
			"status":         game.Status,
			"cityName":       game.CityName,
			"minPlayers":     game.MinPlayers,
			"maxPlayers":     game.MaxPlayers,
			"currentPlayers": game.CurrentPlayers,
		}
	}
	if invite.EntryType == invites.EntryTypeQRCode && s.cfg.Wechat.AppID != "" && s.cfg.Wechat.AppSecret != "" {
		if token, tokenErr := s.wechatAccessToken(r.Context()); tokenErr == nil {
			if image, mimeType, codeErr := s.wechatWxaCode(r.Context(), token, strings.TrimPrefix(strings.Split(path, "?")[0], "/"), scene); codeErr == nil {
				wxaCodeDataURL = "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(image)
			}
		}
	}
	urlLink, urlLinkError := s.createInviteEntryURLLink(r.Context(), path)
	if urlLink == "" {
		urlLink = path
	}
	httpx.OK(w, inviteEntryResponse{
		InviteCode:     invite.Code,
		EntryType:      invite.EntryType,
		Path:           path,
		Title:          title,
		UrlLink:        urlLink,
		UrlLinkError:   urlLinkError,
		Scene:          scene,
		WxaCodeDataURL: wxaCodeDataURL,
		Game:           gameInfo,
	})
}

func (s *Server) createInviteEntryURLLink(ctx context.Context, path string) (string, string) {
	if s.cfg.Wechat.AppID == "" || s.cfg.Wechat.AppSecret == "" {
		return "", "服务器未配置微信 AppID 或 AppSecret，暂不能生成正式 URL Link"
	}
	token, err := s.wechatAccessToken(ctx)
	if err != nil {
		return "", "获取微信接口凭证失败：" + err.Error()
	}
	page, query := miniProgramPageAndQuery(path)
	urlLink, err := s.wechatURLLink(ctx, token, page, query)
	if err != nil {
		return "", "微信 URL Link 生成失败：" + err.Error()
	}
	return urlLink, ""
}

func miniProgramPageAndQuery(path string) (string, string) {
	value := strings.TrimPrefix(strings.TrimSpace(path), "/")
	parts := strings.SplitN(value, "?", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func shareTitleForGame(game games.Game) string {
	title := strings.TrimSpace(game.Title)
	if title == "" {
		title = "真好玩组局"
	}
	cityName := strings.TrimSpace(game.CityName)
	if cityName != "" {
		return cityName + " · " + title
	}
	return title
}

func (s *Server) inviteURLLink(inviteCode string, entryType string, path string) string {
	base := strings.TrimSpace(s.inviteURLLinkBaseURL)
	if base == "" {
		return "wechat://mini-program?path=" + path
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "wechat://mini-program?path=" + path
	}
	q := parsed.Query()
	q.Set("path", path)
	q.Set("query", "inviteCode="+inviteCode+"&entryType="+entryType)
	q.Set("inviteCode", inviteCode)
	q.Set("entryType", entryType)
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func (s *Server) wechatLogin(w http.ResponseWriter, r *http.Request) {
	var req auth.WechatLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}

	resp, err := s.auth.WechatLogin(req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInviteRequired):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invite code required")
		case errors.Is(err, auth.ErrInvalidInvite):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invalid invite code")
		case errors.Is(err, auth.ErrInviteAlreadyBound):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invite code already bound")
		case errors.Is(err, auth.ErrWechatCodeInvalid):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "wechat login code invalid")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "login failed")
		}
		return
	}

	record := s.identity.Status(resp.User.ID)
	resp.IdentityBindStatus = string(record.Status)
	resp.RequiresIdentityBinding = !s.identity.IsVerified(resp.User.ID)
	if !resp.RequiresIdentityBinding {
		session, err := s.auth.IssueAppToken(resp.User.ID)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "login failed")
			return
		}
		resp.Token = session.Token
		resp.ExpiresAt = session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}
	s.recordBehavior(resp.User.ID, "login", "user", resp.User.ID, map[string]interface{}{"needProfile": resp.NeedProfile})
	if resp.InviteRelation != nil && resp.InviteRelation.InviterUserID > 0 {
		s.connections.UpsertPair(resp.InviteRelation.InviterUserID, resp.User.ID, "invite", "invite", resp.InviteRelation.InviteCodeID, 1)
	}
	current := s.buildCurrentUserDTO(resp.User, record)
	httpx.OK(w, appLoginResponse{
		Token:                   resp.Token,
		PreAuthToken:            resp.PreAuthToken,
		ExpiresAt:               resp.ExpiresAt,
		User:                    current,
		InviteRelation:          resp.InviteRelation,
		NeedProfile:             resp.NeedProfile,
		RequiresIdentityBinding: resp.RequiresIdentityBinding,
		IdentityBindStatus:      resp.IdentityBindStatus,
		Roles:                   current.Roles,
		RoleStatusMap:           current.RoleStatusMap,
		EntryType:               resp.EntryType,
		AuthPageMode:            resp.AuthPageMode,
		BoundWechat:             resp.BoundWechat,
		InviteBindingStatus:     resp.InviteBindingStatus,
		InviteBindingMessage:    resp.InviteBindingMessage,
	})
}

func (s *Server) phoneLogin(w http.ResponseWriter, r *http.Request) {
	var req auth.PhoneLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}

	resp, err := s.auth.PhoneLogin(req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrPhoneRequired), errors.Is(err, auth.ErrPhoneInvalid):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid phone")
		case errors.Is(err, auth.ErrPhoneCodeInvalid):
			httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid sms code")
		case errors.Is(err, auth.ErrInviteRequired):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invite code required")
		case errors.Is(err, auth.ErrInvalidInvite):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invalid invite code")
		case errors.Is(err, auth.ErrInviteAlreadyBound):
			httpx.Error(w, http.StatusForbidden, httpx.CodeInviteRequired, "invite code already bound")
		default:
			httpx.Error(w, http.StatusInternalServerError, httpx.CodeSystemError, "phone login failed")
		}
		return
	}

	record, err := s.identity.BindPhone(resp.User.ID, req.Phone)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid phone")
		return
	}
	record, err = s.identity.VerifySMSCode(resp.User.ID, req.Code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid sms code")
		return
	}
	resp.IdentityBindStatus = string(record.Status)
	resp.RequiresIdentityBinding = false
	s.recordBehavior(resp.User.ID, "phone_login", "user", resp.User.ID, map[string]interface{}{
		"authPageMode": resp.AuthPageMode,
		"entryType":    resp.EntryType,
	})
	if resp.InviteRelation != nil && resp.InviteRelation.InviterUserID > 0 {
		s.connections.UpsertPair(resp.InviteRelation.InviterUserID, resp.User.ID, "invite", "invite", resp.InviteRelation.InviteCodeID, 1)
	}
	current := s.buildCurrentUserDTO(resp.User, record)
	httpx.OK(w, appLoginResponse{
		Token:                   resp.Token,
		ExpiresAt:               resp.ExpiresAt,
		User:                    current,
		InviteRelation:          resp.InviteRelation,
		NeedProfile:             resp.NeedProfile,
		RequiresIdentityBinding: resp.RequiresIdentityBinding,
		IdentityBindStatus:      resp.IdentityBindStatus,
		Roles:                   current.Roles,
		RoleStatusMap:           current.RoleStatusMap,
		EntryType:               resp.EntryType,
		AuthPageMode:            resp.AuthPageMode,
		BoundWechat:             resp.BoundWechat,
	})
}

func (s *Server) currentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	user, _ := s.auth.CurrentUser(bearerToken(r.Header.Get("Authorization")))
	record := s.identity.Status(userID)
	httpx.OK(w, s.buildCurrentUserDTO(user, record))
}

func (s *Server) currentUserSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	user, _ := s.auth.CurrentUser(bearerToken(r.Header.Get("Authorization")))
	record := s.identity.Status(userID)
	todos, err := s.reviews.Todos(userID)
	if err != nil {
		writeReviewError(w, err)
		return
	}
	unreadCount := unreadNotificationCount(s.notices.List(userID))
	footprints := s.reviews.Footprints(userID)
	if len(footprints) > 5 {
		footprints = footprints[:5]
	}
	httpx.OK(w, CurrentUserSummaryDTO{
		User:                    s.buildCurrentUserDTO(user, record),
		ReviewTodoCount:         len(todos),
		UnreadNotificationCount: unreadCount,
		IncomeSummary:           s.revenue.IncomeSummary(userID),
		PointsSummary:           s.points.Summary(userID),
		RecentFootprints:        footprints,
	})
}

func unreadNotificationCount(items []notifications.Notification) int {
	count := 0
	for _, item := range items {
		if item.Status == "unread" {
			count++
		}
	}
	return count
}

func (s *Server) buildCurrentUserDTO(user users.User, record identity.Record) CurrentUserDTO {
	snapshot := s.profiles.RoleSnapshot(user.ID)
	growth := s.reviews.Profile(user.ID)
	pointsSummary := s.points.Summary(user.ID)
	membership := s.membership.My(user.ID)
	inviteCode, _ := s.auth.InviteCodeForUser(user.ID)
	if record.Status != "" {
		user.RealnameStatus = string(record.Status)
	} else if user.RealnameStatus == "" {
		user.RealnameStatus = string(record.Status)
	}
	return CurrentUserDTO{
		User:          user,
		Identity:      record,
		Roles:         snapshot.Roles,
		RoleStatusMap: snapshot.RoleStatusMap,
		Membership:    membership,
		Growth: GrowthDTO{
			Level:            growth.Level,
			ExperienceValue:  growth.Experience,
			CreditScore:      growth.CreditScore,
			TodayCreditScore: growth.TodayCreditScore,
			Points:           pointsSummary.AvailablePoints,
		},
		InviteCode:    inviteCode,
		IncomeSummary: s.revenue.IncomeSummary(user.ID),
		Points:        pointsSummary,
	}
}

func (s *Server) updateCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	var req struct {
		Nickname     string `json:"nickname"`
		AvatarURL    string `json:"avatarUrl"`
		AvatarFileID int64  `json:"avatarFileId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "invalid request")
		return
	}
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname != "" && s.rejectSensitiveNickname(w, req.Nickname) {
		return
	}
	if req.AvatarFileID > 0 || strings.TrimSpace(req.AvatarURL) != "" {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "头像需在我的资料提交后台审核")
		return
	}
	current, _ := s.auth.UserByID(userID)
	user, err := s.auth.UpdateProfile(userID, req.Nickname, current.AvatarURL, current.AvatarFileID)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "invalid profile")
		return
	}
	s.recordBehavior(userID, "update_user_profile", "user", userID, map[string]interface{}{"needProfile": user.Nickname == ""})
	httpx.OK(w, user)
}

func bearerToken(value string) string {
	if strings.HasPrefix(value, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
	}
	return ""
}
