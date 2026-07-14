package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr        string
	AppEnv          string
	JWTSecret       string
	IdentityDataKey string
	Storage         StorageConfig
	Database        DatabaseConfig
	OpenIM          OpenIMConfig
	Wechat          WechatConfig
	SMS             SMSConfig
	FaceID          FaceIDConfig
	TencentMap      TencentMapConfig
	FundsService    FundsServiceConfig
	AppLimits       AppLimitsConfig
	LBS             LBSConfig
}

type DatabaseConfig struct {
	URL    string
	Driver string
}

type StorageConfig struct {
	Provider        string
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	COSSecretID     string
	COSSecretKey    string
	UploadBaseURL   string
	DownloadBaseURL string
}

type OpenIMConfig struct {
	APIAddr     string
	Secret      string
	AdminUserID string
	Enabled     bool
}

type WechatConfig struct {
	AppID                 string
	AppSecret             string
	URLLinkBaseURL        string
	MiniProgramEnvVersion string
}

type SMSConfig struct {
	HTTPEndpoint string
	HTTPSecret   string
}

type FaceIDConfig struct {
	HTTPEndpoint             string
	HTTPSecret               string
	CallbackSecret           string
	CallbackRequireSignature bool
}

type TencentMapConfig struct {
	Enabled          bool
	KeyServer        string
	SK               string
	APIBase          string
	RequestTimeoutMS int
}

type FundsServiceConfig struct {
	Addr    string
	Enabled bool
}

type AppLimitsConfig struct {
	DailyGameLimit int
}

type LBSConfig struct {
	DefaultRadiusMeter float64
}

type ReadinessReport struct {
	Environment string          `json:"environment"`
	Production  bool            `json:"production"`
	Ready       bool            `json:"ready"`
	Items       []ReadinessItem `json:"items"`
}

type ReadinessItem struct {
	Key      string            `json:"key"`
	Name     string            `json:"name"`
	Status   string            `json:"status"`
	Ready    bool              `json:"ready"`
	Required bool              `json:"required"`
	Critical bool              `json:"critical"`
	Message  string            `json:"message"`
	Details  map[string]string `json:"details,omitempty"`
}

func Load() Config {
	addr := os.Getenv("GO_API_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	database := DatabaseConfig{
		URL:    os.Getenv("DATABASE_URL"),
		Driver: os.Getenv("DATABASE_DRIVER"),
	}
	if database.Driver == "" && database.URL != "" {
		database.Driver = "postgres"
	}
	openIM := OpenIMConfig{
		APIAddr:     os.Getenv("OPENIM_API_ADDR"),
		Secret:      os.Getenv("OPENIM_SECRET"),
		AdminUserID: os.Getenv("OPENIM_ADMIN_USER_ID"),
	}
	if openIM.AdminUserID == "" {
		openIM.AdminUserID = "imAdmin"
	}
	openIM.Enabled = openIM.APIAddr != "" && openIM.Secret != ""
	wechat := WechatConfig{
		AppID:                 strings.TrimSpace(os.Getenv("WECHAT_APP_ID")),
		AppSecret:             strings.TrimSpace(os.Getenv("WECHAT_APP_SECRET")),
		URLLinkBaseURL:        strings.TrimSpace(os.Getenv("WECHAT_URL_LINK_BASE_URL")),
		MiniProgramEnvVersion: strings.TrimSpace(os.Getenv("WECHAT_MINIPROGRAM_ENV_VERSION")),
	}
	sms := SMSConfig{
		HTTPEndpoint: strings.TrimSpace(os.Getenv("SMS_HTTP_ENDPOINT")),
		HTTPSecret:   strings.TrimSpace(os.Getenv("SMS_HTTP_SECRET")),
	}
	faceID := FaceIDConfig{
		HTTPEndpoint:   strings.TrimSpace(os.Getenv("FACEID_HTTP_ENDPOINT")),
		HTTPSecret:     strings.TrimSpace(os.Getenv("FACEID_HTTP_SECRET")),
		CallbackSecret: strings.TrimSpace(os.Getenv("TENCENT_FACEID_CALLBACK_SECRET")),
	}
	faceID.CallbackRequireSignature = envBool(os.Getenv("TENCENT_FACEID_CALLBACK_REQUIRE_SIGNATURE")) || isProduction(appEnv)
	tencentMap := TencentMapConfig{
		Enabled:          envBool(os.Getenv("TENCENT_MAP_ENABLED")),
		KeyServer:        strings.TrimSpace(os.Getenv("TENCENT_MAP_KEY_SERVER")),
		SK:               strings.TrimSpace(os.Getenv("TENCENT_MAP_SK")),
		APIBase:          strings.TrimSpace(os.Getenv("TENCENT_MAP_API_BASE")),
		RequestTimeoutMS: envInt("TENCENT_MAP_REQUEST_TIMEOUT_MS", 3000),
	}
	if tencentMap.APIBase == "" {
		tencentMap.APIBase = "https://apis.map.qq.com"
	}
	if tencentMap.KeyServer != "" {
		tencentMap.Enabled = true
	}
	funds := FundsServiceConfig{Addr: os.Getenv("FUNDS_SERVICE_ADDR")}
	if funds.Addr == "" {
		funds.Addr = "http://127.0.0.1:8081"
	}
	funds.Enabled = funds.Addr != ""
	storage := StorageConfig{
		Provider:        os.Getenv("STORAGE_PROVIDER"),
		MinIOEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		COSSecretID:     strings.TrimSpace(os.Getenv("COS_SECRET_ID")),
		COSSecretKey:    strings.TrimSpace(os.Getenv("COS_SECRET_KEY")),
		UploadBaseURL:   strings.TrimSpace(os.Getenv("STORAGE_UPLOAD_BASE_URL")),
		DownloadBaseURL: strings.TrimSpace(os.Getenv("STORAGE_DOWNLOAD_BASE_URL")),
	}
	return Config{
		HTTPAddr:        addr,
		AppEnv:          appEnv,
		JWTSecret:       os.Getenv("JWT_SECRET"),
		IdentityDataKey: strings.TrimSpace(os.Getenv("IDENTITY_DATA_KEY")),
		Storage:         storage,
		Database:        database,
		OpenIM:          openIM,
		Wechat:          wechat,
		SMS:             sms,
		FaceID:          faceID,
		TencentMap:      tencentMap,
		FundsService:    funds,
		AppLimits:       AppLimitsConfig{DailyGameLimit: envInt("APP_DAILY_GAME_LIMIT", 3)},
		LBS:             LBSConfig{DefaultRadiusMeter: envFloat("LBS_DEFAULT_RADIUS_METER", 5000)},
	}
}

func (c Config) ValidateProduction() error {
	if !isProduction(c.AppEnv) {
		return nil
	}
	if c.Database.URL == "" {
		return errors.New("DATABASE_URL is required in production")
	}
	if isPlaceholderSecret(c.JWTSecret) {
		return errors.New("JWT_SECRET must be set from a secure secret in production")
	}
	if c.OpenIM.Enabled && isPlaceholderSecret(c.OpenIM.Secret) {
		return errors.New("OPENIM_SECRET must be set from a secure secret in production")
	}
	if c.Wechat.AppID == "" || isPlaceholderSecret(c.Wechat.AppSecret) {
		return errors.New("WECHAT_APP_ID and WECHAT_APP_SECRET must be set in production")
	}
	if !isHTTPSURL(c.SMS.HTTPEndpoint) || isPlaceholderSecret(c.SMS.HTTPSecret) {
		return errors.New("SMS_HTTP_ENDPOINT and SMS_HTTP_SECRET must be set from secure values in production")
	}
	if !isHTTPSURL(c.FaceID.HTTPEndpoint) || isPlaceholderSecret(c.FaceID.HTTPSecret) {
		return errors.New("FACEID_HTTP_ENDPOINT and FACEID_HTTP_SECRET must be set from secure values in production")
	}
	if c.FaceID.CallbackRequireSignature && isPlaceholderSecret(c.FaceID.CallbackSecret) {
		return errors.New("TENCENT_FACEID_CALLBACK_SECRET must be set from a secure secret when callback signature is required")
	}
	if c.TencentMap.Enabled {
		if isPlaceholderSecret(c.TencentMap.KeyServer) {
			return errors.New("TENCENT_MAP_KEY_SERVER must be set from a secure server-side key when Tencent map is enabled")
		}
		if !isHTTPSURL(c.TencentMap.APIBase) {
			return errors.New("TENCENT_MAP_API_BASE must be a valid HTTPS URL when Tencent map is enabled")
		}
	}
	if strings.EqualFold(c.Storage.Provider, "minio") && isPlaceholderSecret(c.Storage.MinIOSecretKey) {
		return errors.New("MINIO_SECRET_KEY must be set from a secure secret in production")
	}
	if strings.EqualFold(c.Storage.Provider, "cos") && (isPlaceholderSecret(c.Storage.COSSecretID) || isPlaceholderSecret(c.Storage.COSSecretKey)) {
		return errors.New("COS_SECRET_ID and COS_SECRET_KEY must be set from secure values when STORAGE_PROVIDER=cos")
	}
	if !isHTTPSURL(c.Storage.UploadBaseURL) || !isHTTPSURL(c.Storage.DownloadBaseURL) {
		return errors.New("STORAGE_UPLOAD_BASE_URL and STORAGE_DOWNLOAD_BASE_URL must be valid HTTPS URLs in production")
	}
	return nil
}

func (c Config) ReadinessReport() ReadinessReport {
	production := isProduction(c.AppEnv)
	report := ReadinessReport{
		Environment: strings.TrimSpace(c.AppEnv),
		Production:  production,
		Ready:       true,
	}
	if report.Environment == "" {
		report.Environment = "local"
	}
	report.Items = append(report.Items,
		readinessItem("database", production, c.Database.URL != "" && c.Database.Driver != "", "DATABASE_URL and DATABASE_DRIVER are configured"),
		readinessItem("jwt_secret", production, !isPlaceholderSecret(c.JWTSecret), "JWT_SECRET is configured with a non-placeholder value"),
		readinessItem("wechat_login", production, c.Wechat.AppID != "" && !isPlaceholderSecret(c.Wechat.AppSecret), "WECHAT_APP_ID and WECHAT_APP_SECRET are configured"),
		readinessItem("wechat_url_link", production, isHTTPSURL(c.Wechat.URLLinkBaseURL), "WECHAT_URL_LINK_BASE_URL is HTTPS for invite link entry"),
		readinessItem("sms", production, isHTTPSURL(c.SMS.HTTPEndpoint) && !isPlaceholderSecret(c.SMS.HTTPSecret), "SMS_HTTP_ENDPOINT is HTTPS and SMS_HTTP_SECRET is configured"),
		readinessItem("faceid", production, isHTTPSURL(c.FaceID.HTTPEndpoint) && !isPlaceholderSecret(c.FaceID.HTTPSecret), "FACEID_HTTP_ENDPOINT is HTTPS and FACEID_HTTP_SECRET is configured"),
		readinessItem("faceid_callback", production && c.FaceID.CallbackRequireSignature, !c.FaceID.CallbackRequireSignature || !isPlaceholderSecret(c.FaceID.CallbackSecret), "TENCENT_FACEID_CALLBACK_SECRET is configured when callback signature is required"),
		readinessItemWithDetails("tencent_map_server_key", c.TencentMap.Enabled, !isPlaceholderSecret(c.TencentMap.KeyServer) && isHTTPSURL(c.TencentMap.APIBase), "TENCENT_MAP_KEY_SERVER is server-side only; mini program frontend does not bind AppID", map[string]string{
			"apiBase": c.TencentMap.APIBase,
			"mode":    "server_proxy",
		}),
		readinessItem("storage_urls", production, isHTTPSURL(c.Storage.UploadBaseURL) && isHTTPSURL(c.Storage.DownloadBaseURL), "STORAGE_UPLOAD_BASE_URL and STORAGE_DOWNLOAD_BASE_URL are HTTPS"),
		readinessItem("storage_secret", production && storageProviderNeedsSecret(c.Storage.Provider), storageSecretReady(c.Storage), "storage provider secret is configured"),
		readinessItem("openim", production, c.OpenIM.Enabled && c.OpenIM.APIAddr != "" && !isPlaceholderSecret(c.OpenIM.Secret), "OPENIM_API_ADDR and OPENIM_SECRET are configured"),
		readinessItem("funds_service", false, strings.TrimSpace(c.FundsService.Addr) != "", "FUNDS_SERVICE_ADDR is configured for revenue integration"),
	)
	for _, item := range report.Items {
		if item.Required && item.Status != "ok" {
			report.Ready = false
			break
		}
	}
	return report
}

func storageProviderNeedsSecret(provider string) bool {
	provider = strings.ToLower(strings.TrimSpace(provider))
	return provider == "minio" || provider == "cos"
}

func storageSecretReady(storage StorageConfig) bool {
	switch strings.ToLower(strings.TrimSpace(storage.Provider)) {
	case "minio":
		return !isPlaceholderSecret(storage.MinIOSecretKey)
	case "cos":
		return !isPlaceholderSecret(storage.COSSecretID) && !isPlaceholderSecret(storage.COSSecretKey)
	default:
		return true
	}
}

func readinessItem(key string, required bool, ok bool, message string) ReadinessItem {
	return readinessItemWithDetails(key, required, ok, message, nil)
}

func readinessItemWithDetails(key string, required bool, ok bool, message string, details map[string]string) ReadinessItem {
	status := "ok"
	if !ok {
		status = "warn"
		if required {
			status = "missing"
		}
	}
	return ReadinessItem{
		Key:      key,
		Name:     readinessItemName(key),
		Status:   status,
		Ready:    ok,
		Required: required,
		Critical: required,
		Message:  message,
		Details:  cleanReadinessDetails(details),
	}
}

func readinessItemName(key string) string {
	switch key {
	case "database":
		return "Database"
	case "jwt_secret":
		return "JWT secret"
	case "wechat_login":
		return "WeChat login"
	case "wechat_url_link":
		return "Invite URL link"
	case "sms":
		return "SMS verify"
	case "faceid":
		return "Realname FaceID"
	case "faceid_callback":
		return "FaceID callback"
	case "tencent_map_server_key":
		return "Tencent map server proxy"
	case "storage_urls":
		return "File upload/download"
	case "storage_secret":
		return "Storage secret"
	case "openim":
		return "OpenIM"
	case "funds_service":
		return "Funds service"
	default:
		return key
	}
}

func cleanReadinessDetails(details map[string]string) map[string]string {
	if len(details) == 0 {
		return nil
	}
	cleaned := make(map[string]string, len(details))
	for key, value := range details {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" || isPlaceholderSecret(value) {
			continue
		}
		cleaned[key] = value
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func envBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envFloat(name string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(name)), 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func isProduction(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "prod" || value == "production"
}

func isHTTPSURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func isPlaceholderSecret(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	lower := strings.ToLower(value)
	return strings.Contains(lower, "change-me") || strings.Contains(lower, "openim123") || strings.Contains(lower, "password") || strings.Contains(lower, "secret")
}
