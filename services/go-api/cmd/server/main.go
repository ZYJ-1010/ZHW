package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"zhw-mini/services/go-api/internal/adminauth"
	"zhw-mini/services/go-api/internal/aidata"
	"zhw-mini/services/go-api/internal/appapi"
	"zhw-mini/services/go-api/internal/audit"
	"zhw-mini/services/go-api/internal/auth"
	"zhw-mini/services/go-api/internal/common/config"
	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/connections"
	"zhw-mini/services/go-api/internal/exports"
	"zhw-mini/services/go-api/internal/files"
	"zhw-mini/services/go-api/internal/gamedrafts"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/identity"
	"zhw-mini/services/go-api/internal/im"
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
	"zhw-mini/services/go-api/internal/tasks"
	"zhw-mini/services/go-api/internal/teams"
	"zhw-mini/services/go-api/internal/users"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateProduction(); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	db := openDatabase(cfg)
	if db != nil {
		defer db.Close()
	}
	userStore := users.NewStoreWithRepository(userRepository(db))
	inviteStore := invites.NewStoreWithRepository(inviteRepository(db))
	tokenStore := auth.NewTokenStoreWithRepository(authRepository(db))
	authService := auth.NewService(userStore, inviteStore, tokenStore)
	authService.UsePhoneLookupSecret(cfg.JWTSecret)
	if cfg.Wechat.AppID != "" && cfg.Wechat.AppSecret != "" {
		authService.UseWechatCodeResolver(auth.NewWechatAPIResolver(cfg.Wechat.AppID, cfg.Wechat.AppSecret))
	}
	identityService := identity.NewServiceWithRepository(identityRepository(db))
	identityDataKey := cfg.IdentityDataKey
	if identityDataKey == "" {
		identityDataKey = cfg.JWTSecret
	}
	identityService.UseDataEncryptionKey(identityDataKey)
	// 腾讯云短信审核通过前，体验版统一使用固定验证码 000000。
	// 恢复真实短信时，再接回 NewHTTPSMSSender。
	if cfg.FaceID.HTTPEndpoint != "" {
		identityService.UseFaceIDStarter(identity.NewHTTPFaceIDStarter(cfg.FaceID.HTTPEndpoint, cfg.FaceID.HTTPSecret))
	}
	gameService := games.NewServiceWithRepositories(identityService, gameRepository(db), favoriteRepository(db))
	if progressRepo := gameProgressRepository(db); progressRepo != nil {
		gameService.UseProgressRepository(progressRepo)
	}
	if confirmRepo := gameServiceConfirmRepository(db); confirmRepo != nil {
		gameService.UseServiceConfirmRepository(confirmRepo)
	}
	lbsService := lbs.NewService()
	imService := im.NewServiceWithOpenIM(gameService, im.OpenIMConfig{
		APIAddr:     cfg.OpenIM.APIAddr,
		Secret:      cfg.OpenIM.Secret,
		AdminUserID: cfg.OpenIM.AdminUserID,
		Enabled:     cfg.OpenIM.Enabled,
	})
	gameService.UseRoomEnsurer(roomEnsurerAdapter{service: imService})
	if db != nil {
		imService.UseRepository(im.NewSQLRepository(db))
	}
	appServer := appapi.New(authService, identityService, gameService, lbsService, imService)
	if db != nil {
		appServer.UseTaskRepository(tasks.NewSQLRepository(db))
		appServer.UseExportRepository(exports.NewSQLRepository(db))
		appServer.UseGameDraftRepository(gamedrafts.NewSQLRepository(db))
	}
	appServer.UseFaceIDCallbackVerifier(cfg.FaceID.CallbackSecret, cfg.FaceID.CallbackRequireSignature)
	appServer.UseAdminRepository(adminRepository(db))
	appServer.UseAIDataRepository(aiDataRepository(db))
	appServer.UseSystemConfigRepository(systemConfigRepository(db))
	appServer.UseRepositories(behaviorRepository(db), operationRepository(db), orderRepository(db), reportRepository(db), notificationRepository(db), fileRepository(db), reviewRepository(db), revenueRepository(db), pointRepository(db), redemptionRepository(db), connectionRepository(db), profileRepository(db), lbsRepository(db), memberReportRepository(db), membershipRepository(db), teamRepository(db))
	appServer.Configure(cfg)
	appServer.Register(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]string{
			"status":  "ok",
			"service": "go-api",
		})
	})

	mux.HandleFunc("GET /api/app/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]string{
			"status":  "ok",
			"service": "app-api",
		})
	})

	mux.HandleFunc("GET /api/admin/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]string{
			"status":  "ok",
			"service": "admin-api",
		})
	})

	log.Printf("go-api listening on %s", cfg.HTTPAddr)
	handler := httpx.AccessLog(httpx.RequestID(mux), nil)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatal(err)
	}
}

// roomEnsurerAdapter keeps the games package independent from the IM return
// type. The games lifecycle only needs the side effect of ensuring a room.
type roomEnsurerAdapter struct{ service *im.Service }

func (a roomEnsurerAdapter) EnsureRoom(gameID int64) {
	if a.service != nil {
		_ = a.service.EnsureRoom(gameID)
	}
}

func openDatabase(cfg config.Config) *sql.DB {
	if cfg.Database.URL == "" || cfg.Database.Driver == "" {
		return nil
	}
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.URL)
	if err != nil {
		log.Printf("database disabled: %v", err)
		return nil
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		log.Printf("database disabled: %v", err)
		return nil
	}
	return db
}

func authRepository(db *sql.DB) auth.SessionRepository {
	if db == nil {
		return nil
	}
	return auth.NewSQLSessionRepository(db)
}

func adminRepository(db *sql.DB) adminauth.Repository {
	if db == nil {
		return nil
	}
	return adminauth.NewSQLRepository(db)
}

func aiDataRepository(db *sql.DB) aidata.Repository {
	if db == nil {
		return nil
	}
	return aidata.NewSQLRepository(db)
}

func systemConfigRepository(db *sql.DB) systemconfig.Repository {
	if db == nil {
		return nil
	}
	return systemconfig.NewSQLRepository(db)
}

func userRepository(db *sql.DB) users.Repository {
	if db == nil {
		return nil
	}
	return users.NewSQLRepository(db)
}

func inviteRepository(db *sql.DB) invites.Repository {
	if db == nil {
		return nil
	}
	return invites.NewSQLRepository(db)
}

func identityRepository(db *sql.DB) identity.Repository {
	if db == nil {
		return nil
	}
	return identity.NewSQLRepository(db)
}

func behaviorRepository(db *sql.DB) audit.BehaviorRepository {
	if db == nil {
		return nil
	}
	return audit.NewSQLBehaviorRepository(db)
}

func operationRepository(db *sql.DB) audit.OperationRepository {
	if db == nil {
		return nil
	}
	return audit.NewSQLOperationRepository(db)
}

func orderRepository(db *sql.DB) orders.Repository {
	if db == nil {
		return nil
	}
	return orders.NewSQLRepository(db)
}

func reportRepository(db *sql.DB) reports.Repository {
	if db == nil {
		return nil
	}
	return reports.NewSQLRepository(db)
}

func notificationRepository(db *sql.DB) notifications.Repository {
	if db == nil {
		return nil
	}
	return notifications.NewSQLRepository(db)
}

func fileRepository(db *sql.DB) files.Repository {
	if db == nil {
		return nil
	}
	return files.NewSQLRepository(db)
}

func reviewRepository(db *sql.DB) reviews.Repository {
	if db == nil {
		return nil
	}
	return reviews.NewSQLRepository(db)
}

func revenueRepository(db *sql.DB) revenue.Repository {
	if db == nil {
		return nil
	}
	return revenue.NewSQLRepository(db)
}

func pointRepository(db *sql.DB) points.Repository {
	if db == nil {
		return nil
	}
	return points.NewSQLRepository(db)
}

func redemptionRepository(db *sql.DB) redemption.Repository {
	if db == nil {
		return nil
	}
	return redemption.NewSQLRepository(db)
}

func connectionRepository(db *sql.DB) connections.Repository {
	if db == nil {
		return nil
	}
	return connections.NewSQLRepository(db)
}

func profileRepository(db *sql.DB) profiles.Repository {
	if db == nil {
		return nil
	}
	return profiles.NewSQLRepository(db)
}

func lbsRepository(db *sql.DB) lbs.Repository {
	if db == nil {
		return nil
	}
	return lbs.NewSQLRepository(db)
}

func memberReportRepository(db *sql.DB) memberreports.Repository {
	if db == nil {
		return nil
	}
	return memberreports.NewSQLRepository(db)
}

func membershipRepository(db *sql.DB) membership.Repository {
	if db == nil {
		return nil
	}
	return membership.NewSQLRepository(db)
}

func teamRepository(db *sql.DB) teams.Repository {
	if db == nil {
		return nil
	}
	return teams.NewSQLRepository(db)
}

func favoriteRepository(db *sql.DB) games.FavoriteRepository {
	if db == nil {
		return nil
	}
	return games.NewSQLFavoriteRepository(db)
}

func gameRepository(db *sql.DB) games.Repository {
	if db == nil {
		return nil
	}
	return games.NewSQLRepository(db)
}

func gameDraftRepository(db *sql.DB) gamedrafts.Repository {
	if db == nil {
		return nil
	}
	return gamedrafts.NewSQLRepository(db)
}

func gameProgressRepository(db *sql.DB) games.ProgressRepository {
	if db == nil {
		return nil
	}
	return games.NewSQLProgressRepository(db)
}

func gameServiceConfirmRepository(db *sql.DB) games.ServiceConfirmRepository {
	if db == nil {
		return nil
	}
	return games.NewSQLServiceConfirmRepository(db)
}
