package config

import "testing"

func TestValidateProductionRejectsPlaceholderSecrets(t *testing.T) {
	cfg := Config{
		AppEnv:    "production",
		JWTSecret: "change-me",
		Database:  DatabaseConfig{URL: "postgres://user:pass@db/app"},
		OpenIM:    OpenIMConfig{Enabled: true, Secret: "openIM123"},
		Storage:   StorageConfig{Provider: "minio", MinIOSecretKey: "change-me-123456"},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected placeholder secrets to be rejected in production")
	}
}

func TestValidateProductionAllowsSecureEnvironmentSecrets(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		OpenIM:    OpenIMConfig{Enabled: true, Secret: "openim-prod-32-random-bytes-value"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		TencentMap: TencentMapConfig{
			Enabled:   true,
			KeyServer: "tencent-map-prod-server-key",
			APIBase:   "https://apis.map.qq.com",
		},
		Storage: StorageConfig{
			Provider:        "cos",
			COSSecretID:     "AKID-prod-value",
			COSSecretKey:    "cos-prod-32-random-bytes-value",
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	if err := cfg.ValidateProduction(); err != nil {
		t.Fatalf("expected production config to pass: %v", err)
	}
}

func TestValidateProductionRequiresTencentMapServerKeyWhenEnabled(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		TencentMap: TencentMapConfig{
			Enabled: true,
			APIBase: "https://apis.map.qq.com",
		},
		Storage: StorageConfig{
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected missing Tencent map server key to be rejected")
	}
}

func TestValidateProductionSkipsLocalDefaults(t *testing.T) {
	cfg := Config{
		AppEnv:    "local",
		JWTSecret: "change-me",
	}

	if err := cfg.ValidateProduction(); err != nil {
		t.Fatalf("expected local config to pass: %v", err)
	}
}

func TestValidateProductionRequiresFaceIDCallbackSecret(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected missing faceid callback secret to be rejected")
	}
}

func TestValidateProductionRequiresWechatCredentials(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected missing wechat credentials to be rejected")
	}
}

func TestValidateProductionRequiresSMSConfig(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		Storage: StorageConfig{
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected missing sms config to be rejected")
	}
}

func TestValidateProductionRequiresFaceIDHTTPConfig(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		Storage: StorageConfig{
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected missing faceid http config to be rejected")
	}
}

func TestValidateProductionRequiresHTTPSStorageURLs(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		Storage: StorageConfig{
			Provider:        "minio",
			MinIOSecretKey:  "minio-prod-32-random-bytes-value",
			UploadBaseURL:   "http://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected non-https storage upload url to be rejected")
	}
}

func TestValidateProductionRequiresCOSSecrets(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app"},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		Storage: StorageConfig{
			Provider:        "cos",
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected missing COS secrets to be rejected")
	}
}

func TestReadinessReportMarksMissingProductionDependencies(t *testing.T) {
	cfg := Config{
		AppEnv:    "production",
		JWTSecret: "change-me",
	}

	report := cfg.ReadinessReport()
	if report.Ready {
		t.Fatalf("expected missing production dependencies to make readiness false: %+v", report)
	}
	if !report.Production || report.Environment != "production" {
		t.Fatalf("expected production report metadata: %+v", report)
	}
	if !hasReadinessStatus(report, "jwt_secret", "missing", true) || !hasReadinessStatus(report, "wechat_login", "missing", true) {
		t.Fatalf("expected missing required items: %+v", report.Items)
	}
}

func TestReadinessReportAllowsCompleteProductionConfig(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app", Driver: "postgres"},
		OpenIM:    OpenIMConfig{APIAddr: "https://openim.example.com", Secret: "openim-prod-32-random-bytes-value", AdminUserID: "imAdmin", Enabled: true},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		Storage: StorageConfig{
			Provider:        "minio",
			MinIOSecretKey:  "minio-prod-32-random-bytes-value",
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	report := cfg.ReadinessReport()
	if !report.Ready {
		t.Fatalf("expected complete production config to be ready: %+v", report.Items)
	}
	if !hasReadinessStatus(report, "openim", "ok", true) || !hasReadinessStatus(report, "faceid_callback", "ok", true) {
		t.Fatalf("expected critical items ok: %+v", report.Items)
	}
}

func TestReadinessReportIncludesSafeIntegrationFields(t *testing.T) {
	cfg := Config{
		AppEnv:    "prod",
		JWTSecret: "jwt-prod-32-random-bytes-value",
		Database:  DatabaseConfig{URL: "postgres://user:strong-pass@db/app", Driver: "postgres"},
		OpenIM:    OpenIMConfig{APIAddr: "https://openim.example.com", Secret: "openim-prod-32-random-bytes-value", AdminUserID: "imAdmin", Enabled: true},
		Wechat:    WechatConfig{AppID: "wx-prod-appid", AppSecret: "wechat-prod-32-random-bytes-value", URLLinkBaseURL: "https://wxaurl.example.com/invite"},
		SMS:       SMSConfig{HTTPEndpoint: "https://sms.example.com/send", HTTPSecret: "sms-prod-32-random-bytes-value"},
		FaceID:    FaceIDConfig{HTTPEndpoint: "https://faceid.example.com/detect-auth", HTTPSecret: "faceid-http-32-random-bytes-value", CallbackRequireSignature: true, CallbackSecret: "faceid-prod-32-random-bytes-value"},
		TencentMap: TencentMapConfig{
			Enabled:   true,
			KeyServer: "tencent-map-prod-server-key",
			APIBase:   "https://apis.map.qq.com",
		},
		Storage: StorageConfig{
			UploadBaseURL:   "https://upload.example.com",
			DownloadBaseURL: "https://download.example.com",
		},
	}

	report := cfg.ReadinessReport()
	if !hasReadinessStatus(report, "wechat_url_link", "ok", true) || !hasReadinessStatus(report, "tencent_map_server_key", "ok", true) {
		t.Fatalf("expected invite url link and map proxy readiness: %+v", report.Items)
	}
	for _, item := range report.Items {
		if item.Key == "tencent_map_server_key" {
			if !item.Ready || !item.Critical || item.Name == "" || item.Details["mode"] != "server_proxy" {
				t.Fatalf("expected safe map readiness metadata, got %+v", item)
			}
			if item.Details["keyServer"] != "" {
				t.Fatalf("readiness details must not expose map key: %+v", item.Details)
			}
		}
	}
}

func TestLoadReadsBusinessLimits(t *testing.T) {
	t.Setenv("APP_DAILY_GAME_LIMIT", "2")
	t.Setenv("LBS_DEFAULT_RADIUS_METER", "1500")
	t.Setenv("WECHAT_APP_ID", "wx-test")
	t.Setenv("WECHAT_APP_SECRET", "wechat-secret")
	t.Setenv("WECHAT_URL_LINK_BASE_URL", "https://wxaurl.example.com/invite")
	t.Setenv("SMS_HTTP_ENDPOINT", "https://sms.example.com/send")
	t.Setenv("SMS_HTTP_SECRET", "sms-secret")
	t.Setenv("FACEID_HTTP_ENDPOINT", "https://faceid.example.com/detect-auth")
	t.Setenv("FACEID_HTTP_SECRET", "faceid-secret")
	t.Setenv("TENCENT_MAP_KEY_SERVER", "server-map-key")
	t.Setenv("TENCENT_MAP_SK", "server-map-sk")
	t.Setenv("TENCENT_MAP_API_BASE", "https://apis.map.qq.com")
	t.Setenv("TENCENT_MAP_REQUEST_TIMEOUT_MS", "2500")
	t.Setenv("STORAGE_UPLOAD_BASE_URL", "https://upload.example.com")
	t.Setenv("STORAGE_DOWNLOAD_BASE_URL", "https://download.example.com")
	t.Setenv("COS_SECRET_ID", "AKID-load-test")
	t.Setenv("COS_SECRET_KEY", "cos-load-secret")

	cfg := Load()

	if cfg.AppLimits.DailyGameLimit != 2 {
		t.Fatalf("expected daily game limit from env, got %d", cfg.AppLimits.DailyGameLimit)
	}
	if cfg.LBS.DefaultRadiusMeter != 1500 {
		t.Fatalf("expected default radius from env, got %f", cfg.LBS.DefaultRadiusMeter)
	}
	if cfg.Wechat.AppID != "wx-test" || cfg.Wechat.AppSecret != "wechat-secret" {
		t.Fatalf("expected wechat config from env, got %+v", cfg.Wechat)
	}
	if cfg.Wechat.URLLinkBaseURL != "https://wxaurl.example.com/invite" {
		t.Fatalf("expected wechat url link base from env, got %+v", cfg.Wechat)
	}
	if cfg.SMS.HTTPEndpoint != "https://sms.example.com/send" || cfg.SMS.HTTPSecret != "sms-secret" {
		t.Fatalf("expected sms config from env, got %+v", cfg.SMS)
	}
	if cfg.FaceID.HTTPEndpoint != "https://faceid.example.com/detect-auth" || cfg.FaceID.HTTPSecret != "faceid-secret" {
		t.Fatalf("expected faceid config from env, got %+v", cfg.FaceID)
	}
	if !cfg.TencentMap.Enabled || cfg.TencentMap.KeyServer != "server-map-key" || cfg.TencentMap.SK != "server-map-sk" || cfg.TencentMap.RequestTimeoutMS != 2500 {
		t.Fatalf("expected Tencent map server config from env, got %+v", cfg.TencentMap)
	}
	if cfg.Storage.UploadBaseURL != "https://upload.example.com" || cfg.Storage.DownloadBaseURL != "https://download.example.com" {
		t.Fatalf("expected storage base urls from env, got %+v", cfg.Storage)
	}
	if cfg.Storage.COSSecretID != "AKID-load-test" || cfg.Storage.COSSecretKey != "cos-load-secret" {
		t.Fatalf("expected COS config from env, got %+v", cfg.Storage)
	}
}

func hasReadinessStatus(report ReadinessReport, key string, status string, required bool) bool {
	for _, item := range report.Items {
		if item.Key == key && item.Status == status && item.Required == required {
			return true
		}
	}
	return false
}
