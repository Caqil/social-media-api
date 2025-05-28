// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server Configuration
	Server ServerConfig `json:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database"`

	// Redis Configuration
	Redis RedisConfig `json:"redis"`

	// JWT Configuration
	JWT JWTConfig `json:"jwt"`

	// Email Configuration
	Email EmailConfig `json:"email"`

	// File Upload Configuration
	Upload UploadConfig `json:"upload"`

	// AWS Configuration
	AWS AWSConfig `json:"aws"`

	// Rate Limiting Configuration
	RateLimit RateLimitConfig `json:"rate_limit"`

	// Security Configuration
	Security SecurityConfig `json:"security"`

	// Feature Flags
	Features FeatureFlags `json:"features"`

	// External Services
	External ExternalConfig `json:"external"`

	// Monitoring
	Monitoring MonitoringConfig `json:"monitoring"`

	// Environment
	Environment string `json:"environment"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port            string        `json:"port"`
	Host            string        `json:"host"`
	Mode            string        `json:"mode"` // debug, release, test
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	MaxRequestSize  int64         `json:"max_request_size"`
	TrustedProxies  []string      `json:"trusted_proxies"`
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	MongoURI        string        `json:"mongo_uri"`
	DatabaseName    string        `json:"database_name"`
	MaxPoolSize     uint64        `json:"max_pool_size"`
	MinPoolSize     uint64        `json:"min_pool_size"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
	ServerTimeout   time.Duration `json:"server_timeout"`
}

// RedisConfig contains Redis-related configuration
type RedisConfig struct {
	URL              string        `json:"url"`
	Host             string        `json:"host"`
	Port             string        `json:"port"`
	Password         string        `json:"password"`
	Database         int           `json:"database"`
	MaxRetries       int           `json:"max_retries"`
	MinRetryBackoff  time.Duration `json:"min_retry_backoff"`
	MaxRetryBackoff  time.Duration `json:"max_retry_backoff"`
	DialTimeout      time.Duration `json:"dial_timeout"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	PoolSize         int           `json:"pool_size"`
	MinIdleConns     int           `json:"min_idle_conns"`
	MaxConnAge       time.Duration `json:"max_conn_age"`
	PoolTimeout      time.Duration `json:"pool_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	IdleCheckFreq    time.Duration `json:"idle_check_freq"`
	EnableCluster    bool          `json:"enable_cluster"`
	ClusterAddresses []string      `json:"cluster_addresses"`
}

// JWTConfig contains JWT-related configuration
type JWTConfig struct {
	SecretKey            string        `json:"secret_key"`
	RefreshSecretKey     string        `json:"refresh_secret_key"`
	AccessTokenDuration  time.Duration `json:"access_token_duration"`
	RefreshTokenDuration time.Duration `json:"refresh_token_duration"`
	Issuer               string        `json:"issuer"`
	Algorithm            string        `json:"algorithm"`
}

// EmailConfig contains email-related configuration
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     string `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	ReplyTo      string `json:"reply_to"`
	UseTLS       bool   `json:"use_tls"`
	UseSSL       bool   `json:"use_ssl"`
}

// UploadConfig contains file upload configuration
type UploadConfig struct {
	MaxFileSize     int64    `json:"max_file_size"`
	MaxImageSize    int64    `json:"max_image_size"`
	MaxVideoSize    int64    `json:"max_video_size"`
	MaxAudioSize    int64    `json:"max_audio_size"`
	MaxDocumentSize int64    `json:"max_document_size"`
	AllowedTypes    []string `json:"allowed_types"`
	UploadPath      string   `json:"upload_path"`
	TempPath        string   `json:"temp_path"`
	UseS3           bool     `json:"use_s3"`
	LocalURL        string   `json:"local_url"`
}

// AWSConfig contains AWS-related configuration
type AWSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	S3Bucket        string `json:"s3_bucket"`
	S3Endpoint      string `json:"s3_endpoint"`
	CloudFrontURL   string `json:"cloudfront_url"`
	SESRegion       string `json:"ses_region"`
	SNSRegion       string `json:"sns_region"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled             bool          `json:"enabled"`
	DefaultLimit        int           `json:"default_limit"`
	DefaultWindow       time.Duration `json:"default_window"`
	AuthLimit           int           `json:"auth_limit"`
	AuthWindow          time.Duration `json:"auth_window"`
	PostLimit           int           `json:"post_limit"`
	PostWindow          time.Duration `json:"post_window"`
	CommentLimit        int           `json:"comment_limit"`
	CommentWindow       time.Duration `json:"comment_window"`
	MessageLimit        int           `json:"message_limit"`
	MessageWindow       time.Duration `json:"message_window"`
	PasswordResetLimit  int           `json:"password_reset_limit"`
	PasswordResetWindow time.Duration `json:"password_reset_window"`
	EmailVerifyLimit    int           `json:"email_verify_limit"`
	EmailVerifyWindow   time.Duration `json:"email_verify_window"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	PasswordMinLength    int      `json:"password_min_length"`
	PasswordRequireUpper bool     `json:"password_require_upper"`
	PasswordRequireLower bool     `json:"password_require_lower"`
	PasswordRequireDigit bool     `json:"password_require_digit"`
	PasswordRequireSpec  bool     `json:"password_require_special"`
	AllowedOrigins       []string `json:"allowed_origins"`
	AllowedMethods       []string `json:"allowed_methods"`
	AllowedHeaders       []string `json:"allowed_headers"`
	EnableCSRF           bool     `json:"enable_csrf"`
	CSRFSecret           string   `json:"csrf_secret"`
	EnableHTTPS          bool     `json:"enable_https"`
	HSTSEnabled          bool     `json:"hsts_enabled"`
	HSTSMaxAge           int      `json:"hsts_max_age"`
}

// FeatureFlags contains feature toggle configuration
type FeatureFlags struct {
	EnableStories            bool `json:"enable_stories"`
	EnableGroups             bool `json:"enable_groups"`
	EnableEvents             bool `json:"enable_events"`
	EnableLiveChat           bool `json:"enable_live_chat"`
	EnablePushNotifications  bool `json:"enable_push_notifications"`
	EnableEmailNotifications bool `json:"enable_email_notifications"`
	EnableSMSNotifications   bool `json:"enable_sms_notifications"`
	EnableContentModeration  bool `json:"enable_content_moderation"`
	EnableAnalytics          bool `json:"enable_analytics"`
	EnableSearch             bool `json:"enable_search"`
	EnableFeedAlgorithm      bool `json:"enable_feed_algorithm"`
	EnableFileUploads        bool `json:"enable_file_uploads"`
	EnableVideoUploads       bool `json:"enable_video_uploads"`
	EnableAudioUploads       bool `json:"enable_audio_uploads"`
}

// ExternalConfig contains external service configuration
type ExternalConfig struct {
	FrontendURL        string `json:"frontend_url"`
	APIURL             string `json:"api_url"`
	WebhookSecret      string `json:"webhook_secret"`
	GoogleClientID     string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
	FacebookAppID      string `json:"facebook_app_id"`
	FacebookAppSecret  string `json:"facebook_app_secret"`
	TwitterAPIKey      string `json:"twitter_api_key"`
	TwitterAPISecret   string `json:"twitter_api_secret"`
	FirebaseServerKey  string `json:"firebase_server_key"`
	TwilioAccountSID   string `json:"twilio_account_sid"`
	TwilioAuthToken    string `json:"twilio_auth_token"`
	TwilioPhoneNumber  string `json:"twilio_phone_number"`
}

// MonitoringConfig contains monitoring and logging configuration
type MonitoringConfig struct {
	EnableMetrics     bool    `json:"enable_metrics"`
	MetricsPort       string  `json:"metrics_port"`
	LogLevel          string  `json:"log_level"`
	LogFormat         string  `json:"log_format"` // json, text
	LogFile           string  `json:"log_file"`
	EnableRequestLog  bool    `json:"enable_request_log"`
	EnableErrorLog    bool    `json:"enable_error_log"`
	SentryDSN         string  `json:"sentry_dsn"`
	EnableProfiling   bool    `json:"enable_profiling"`
	ProfilingPort     string  `json:"profiling_port"`
	HealthCheckPath   string  `json:"health_check_path"`
	MetricsPath       string  `json:"metrics_path"`
	EnableTracing     bool    `json:"enable_tracing"`
	JaegerEndpoint    string  `json:"jaeger_endpoint"`
	TracingSampleRate float64 `json:"tracing_sample_rate"`
}

// Global config instance
var AppConfig *Config

// Load loads configuration from environment variables
func Load() *Config {
	config := &Config{
		Server:      loadServerConfig(),
		Database:    loadDatabaseConfig(),
		Redis:       loadRedisConfig(),
		JWT:         loadJWTConfig(),
		Email:       loadEmailConfig(),
		Upload:      loadUploadConfig(),
		AWS:         loadAWSConfig(),
		RateLimit:   loadRateLimitConfig(),
		Security:    loadSecurityConfig(),
		Features:    loadFeatureFlags(),
		External:    loadExternalConfig(),
		Monitoring:  loadMonitoringConfig(),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	AppConfig = config
	return config
}

// loadServerConfig loads server configuration
func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port:            getEnv("PORT", "8080"),
		Host:            getEnv("HOST", "0.0.0.0"),
		Mode:            getEnv("GIN_MODE", "debug"),
		ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
		MaxRequestSize:  getEnvInt64("MAX_REQUEST_SIZE", 32<<20), // 32MB
		TrustedProxies:  getEnvStringSlice("TRUSTED_PROXIES", []string{}),
	}
}

// loadDatabaseConfig loads database configuration
func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:    getEnv("DB_NAME", "social_media"),
		MaxPoolSize:     getEnvUint64("MONGO_MAX_POOL_SIZE", 100),
		MinPoolSize:     getEnvUint64("MONGO_MIN_POOL_SIZE", 5),
		MaxConnIdleTime: getEnvDuration("MONGO_MAX_CONN_IDLE_TIME", 30*time.Minute),
		ConnectTimeout:  getEnvDuration("MONGO_CONNECT_TIMEOUT", 10*time.Second),
		ServerTimeout:   getEnvDuration("MONGO_SERVER_TIMEOUT", 10*time.Second),
	}
}

// loadRedisConfig loads Redis configuration
func loadRedisConfig() RedisConfig {
	return RedisConfig{
		URL:              getEnv("REDIS_URL", ""),
		Host:             getEnv("REDIS_HOST", "localhost"),
		Port:             getEnv("REDIS_PORT", "6379"),
		Password:         getEnv("REDIS_PASSWORD", ""),
		Database:         getEnvInt("REDIS_DB", 0),
		MaxRetries:       getEnvInt("REDIS_MAX_RETRIES", 3),
		MinRetryBackoff:  getEnvDuration("REDIS_MIN_RETRY_BACKOFF", 8*time.Millisecond),
		MaxRetryBackoff:  getEnvDuration("REDIS_MAX_RETRY_BACKOFF", 512*time.Millisecond),
		DialTimeout:      getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:      getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout:     getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		PoolSize:         getEnvInt("REDIS_POOL_SIZE", 20),
		MinIdleConns:     getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
		MaxConnAge:       getEnvDuration("REDIS_MAX_CONN_AGE", 0),
		PoolTimeout:      getEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
		IdleTimeout:      getEnvDuration("REDIS_IDLE_TIMEOUT", 5*time.Minute),
		IdleCheckFreq:    getEnvDuration("REDIS_IDLE_CHECK_FREQ", 1*time.Minute),
		EnableCluster:    getEnvBool("REDIS_ENABLE_CLUSTER", false),
		ClusterAddresses: getEnvStringSlice("REDIS_CLUSTER_ADDRESSES", []string{}),
	}
}

// loadJWTConfig loads JWT configuration
func loadJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:            getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		RefreshSecretKey:     getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
		AccessTokenDuration:  getEnvDuration("JWT_ACCESS_DURATION", 24*time.Hour),
		RefreshTokenDuration: getEnvDuration("JWT_REFRESH_DURATION", 30*24*time.Hour),
		Issuer:               getEnv("JWT_ISSUER", "social-media-api"),
		Algorithm:            getEnv("JWT_ALGORITHM", "HS256"),
	}
}

// loadEmailConfig loads email configuration
func loadEmailConfig() EmailConfig {
	return EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASS", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@socialmedia.com"),
		FromName:     getEnv("FROM_NAME", "Social Media App"),
		ReplyTo:      getEnv("REPLY_TO_EMAIL", ""),
		UseTLS:       getEnvBool("SMTP_USE_TLS", true),
		UseSSL:       getEnvBool("SMTP_USE_SSL", false),
	}
}

// loadUploadConfig loads upload configuration
func loadUploadConfig() UploadConfig {
	return UploadConfig{
		MaxFileSize:     getEnvInt64("MAX_FILE_SIZE", 100<<20),    // 100MB
		MaxImageSize:    getEnvInt64("MAX_IMAGE_SIZE", 10<<20),    // 10MB
		MaxVideoSize:    getEnvInt64("MAX_VIDEO_SIZE", 100<<20),   // 100MB
		MaxAudioSize:    getEnvInt64("MAX_AUDIO_SIZE", 50<<20),    // 50MB
		MaxDocumentSize: getEnvInt64("MAX_DOCUMENT_SIZE", 25<<20), // 25MB
		AllowedTypes:    getEnvStringSlice("ALLOWED_FILE_TYPES", []string{"image/jpeg", "image/png", "image/gif", "video/mp4"}),
		UploadPath:      getEnv("UPLOAD_PATH", "./uploads"),
		TempPath:        getEnv("TEMP_PATH", "./temp"),
		UseS3:           getEnvBool("USE_S3", false),
		LocalURL:        getEnv("LOCAL_UPLOAD_URL", "http://localhost:8080/uploads"),
	}
}

// loadAWSConfig loads AWS configuration
func loadAWSConfig() AWSConfig {
	return AWSConfig{
		AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		Region:          getEnv("AWS_REGION", "us-east-1"),
		S3Bucket:        getEnv("S3_BUCKET", ""),
		S3Endpoint:      getEnv("S3_ENDPOINT", ""),
		CloudFrontURL:   getEnv("CLOUDFRONT_URL", ""),
		SESRegion:       getEnv("SES_REGION", "us-east-1"),
		SNSRegion:       getEnv("SNS_REGION", "us-east-1"),
	}
}

// loadRateLimitConfig loads rate limiting configuration
func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:             getEnvBool("RATE_LIMIT_ENABLED", true),
		DefaultLimit:        getEnvInt("RATE_LIMIT_DEFAULT", 100),
		DefaultWindow:       getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		AuthLimit:           getEnvInt("RATE_LIMIT_AUTH", 5),
		AuthWindow:          getEnvDuration("RATE_LIMIT_AUTH_WINDOW", 1*time.Minute),
		PostLimit:           getEnvInt("RATE_LIMIT_POST", 10),
		PostWindow:          getEnvDuration("RATE_LIMIT_POST_WINDOW", 1*time.Hour),
		CommentLimit:        getEnvInt("RATE_LIMIT_COMMENT", 30),
		CommentWindow:       getEnvDuration("RATE_LIMIT_COMMENT_WINDOW", 1*time.Minute),
		MessageLimit:        getEnvInt("RATE_LIMIT_MESSAGE", 60),
		MessageWindow:       getEnvDuration("RATE_LIMIT_MESSAGE_WINDOW", 1*time.Minute),
		PasswordResetLimit:  getEnvInt("RATE_LIMIT_PASSWORD_RESET", 3),
		PasswordResetWindow: getEnvDuration("RATE_LIMIT_PASSWORD_RESET_WINDOW", 1*time.Hour),
		EmailVerifyLimit:    getEnvInt("RATE_LIMIT_EMAIL_VERIFY", 3),
		EmailVerifyWindow:   getEnvDuration("RATE_LIMIT_EMAIL_VERIFY_WINDOW", 1*time.Hour),
	}
}

// loadSecurityConfig loads security configuration
func loadSecurityConfig() SecurityConfig {
	return SecurityConfig{
		PasswordMinLength:    getEnvInt("PASSWORD_MIN_LENGTH", 8),
		PasswordRequireUpper: getEnvBool("PASSWORD_REQUIRE_UPPER", true),
		PasswordRequireLower: getEnvBool("PASSWORD_REQUIRE_LOWER", true),
		PasswordRequireDigit: getEnvBool("PASSWORD_REQUIRE_DIGIT", true),
		PasswordRequireSpec:  getEnvBool("PASSWORD_REQUIRE_SPECIAL", false),
		AllowedOrigins:       getEnvStringSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		AllowedMethods:       getEnvStringSlice("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowedHeaders:       getEnvStringSlice("ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		EnableCSRF:           getEnvBool("ENABLE_CSRF", false),
		CSRFSecret:           getEnv("CSRF_SECRET", "csrf-secret-key"),
		EnableHTTPS:          getEnvBool("ENABLE_HTTPS", false),
		HSTSEnabled:          getEnvBool("HSTS_ENABLED", false),
		HSTSMaxAge:           getEnvInt("HSTS_MAX_AGE", 31536000), // 1 year
	}
}

// loadFeatureFlags loads feature flags
func loadFeatureFlags() FeatureFlags {
	return FeatureFlags{
		EnableStories:            getEnvBool("ENABLE_STORIES", true),
		EnableGroups:             getEnvBool("ENABLE_GROUPS", true),
		EnableEvents:             getEnvBool("ENABLE_EVENTS", true),
		EnableLiveChat:           getEnvBool("ENABLE_LIVE_CHAT", true),
		EnablePushNotifications:  getEnvBool("ENABLE_PUSH_NOTIFICATIONS", true),
		EnableEmailNotifications: getEnvBool("ENABLE_EMAIL_NOTIFICATIONS", true),
		EnableSMSNotifications:   getEnvBool("ENABLE_SMS_NOTIFICATIONS", false),
		EnableContentModeration:  getEnvBool("ENABLE_CONTENT_MODERATION", true),
		EnableAnalytics:          getEnvBool("ENABLE_ANALYTICS", true),
		EnableSearch:             getEnvBool("ENABLE_SEARCH", true),
		EnableFeedAlgorithm:      getEnvBool("ENABLE_FEED_ALGORITHM", true),
		EnableFileUploads:        getEnvBool("ENABLE_FILE_UPLOADS", true),
		EnableVideoUploads:       getEnvBool("ENABLE_VIDEO_UPLOADS", true),
		EnableAudioUploads:       getEnvBool("ENABLE_AUDIO_UPLOADS", true),
	}
}

// loadExternalConfig loads external service configuration
func loadExternalConfig() ExternalConfig {
	return ExternalConfig{
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		APIURL:             getEnv("API_URL", "http://localhost:8080"),
		WebhookSecret:      getEnv("WEBHOOK_SECRET", "webhook-secret"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		FacebookAppID:      getEnv("FACEBOOK_APP_ID", ""),
		FacebookAppSecret:  getEnv("FACEBOOK_APP_SECRET", ""),
		TwitterAPIKey:      getEnv("TWITTER_API_KEY", ""),
		TwitterAPISecret:   getEnv("TWITTER_API_SECRET", ""),
		FirebaseServerKey:  getEnv("FIREBASE_SERVER_KEY", ""),
		TwilioAccountSID:   getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:    getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber:  getEnv("TWILIO_PHONE_NUMBER", ""),
	}
}

// loadMonitoringConfig loads monitoring configuration
func loadMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		EnableMetrics:     getEnvBool("ENABLE_METRICS", true),
		MetricsPort:       getEnv("METRICS_PORT", "2112"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
		LogFile:           getEnv("LOG_FILE", ""),
		EnableRequestLog:  getEnvBool("ENABLE_REQUEST_LOG", true),
		EnableErrorLog:    getEnvBool("ENABLE_ERROR_LOG", true),
		SentryDSN:         getEnv("SENTRY_DSN", ""),
		EnableProfiling:   getEnvBool("ENABLE_PROFILING", false),
		ProfilingPort:     getEnv("PROFILING_PORT", "6060"),
		HealthCheckPath:   getEnv("HEALTH_CHECK_PATH", "/health"),
		MetricsPath:       getEnv("METRICS_PATH", "/metrics"),
		EnableTracing:     getEnvBool("ENABLE_TRACING", false),
		JaegerEndpoint:    getEnv("JAEGER_ENDPOINT", ""),
		TracingSampleRate: getEnvFloat64("TRACING_SAMPLE_RATE", 0.1),
	}
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvInt64 gets environment variable as int64 with default value
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid int64 value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvUint64 gets environment variable as uint64 with default value
func getEnvUint64(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid uint64 value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvFloat64 gets environment variable as float64 with default value
func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
		log.Printf("Warning: Invalid float64 value for %s: %s, using default: %f", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvBool gets environment variable as boolean with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Warning: Invalid boolean value for %s: %s, using default: %v", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvDuration gets environment variable as duration with default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: Invalid duration value for %s: %s, using default: %v", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvStringSlice gets environment variable as string slice with default value
func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.SecretKey == "your-secret-key-change-in-production" {
		log.Println("Warning: Using default JWT secret key. Please change in production!")
	}

	if c.JWT.RefreshSecretKey == "your-refresh-secret-key-change-in-production" {
		log.Println("Warning: Using default JWT refresh secret key. Please change in production!")
	}

	if c.Database.MongoURI == "" {
		return fmt.Errorf("database URI is required")
	}

	if c.Environment == "production" {
		if c.JWT.SecretKey == "your-secret-key-change-in-production" {
			return fmt.Errorf("JWT secret key must be set in production")
		}
		if c.Server.Mode != "release" {
			log.Println("Warning: Server should be in release mode for production")
		}
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return c.Environment == "test" || c.Environment == "testing"
}

// GetRedisAddr returns Redis address in host:port format
func (c *Config) GetRedisAddr() string {
	if c.Redis.URL != "" {
		return c.Redis.URL
	}
	return c.Redis.Host + ":" + c.Redis.Port
}

// GetServerAddr returns server address in host:port format
func (c *Config) GetServerAddr() string {
	return c.Server.Host + ":" + c.Server.Port
}

// GetDatabaseURI returns the complete database URI
func (c *Config) GetDatabaseURI() string {
	return c.Database.MongoURI
}

// PrintConfig prints configuration (excluding sensitive data)
func (c *Config) PrintConfig() {
	log.Printf("=== Application Configuration ===")
	log.Printf("Environment: %s", c.Environment)
	log.Printf("Server: %s (mode: %s)", c.GetServerAddr(), c.Server.Mode)
	log.Printf("Database: %s", c.Database.DatabaseName)
	log.Printf("Redis: %s (DB: %d)", c.GetRedisAddr(), c.Redis.Database)
	log.Printf("Features: Stories=%v, Groups=%v, Events=%v, LiveChat=%v",
		c.Features.EnableStories, c.Features.EnableGroups, c.Features.EnableEvents, c.Features.EnableLiveChat)
	log.Printf("Upload: UseS3=%v, MaxSize=%d MB", c.Upload.UseS3, c.Upload.MaxFileSize/(1024*1024))
	log.Printf("Rate Limiting: Enabled=%v, Default=%d req/min", c.RateLimit.Enabled, c.RateLimit.DefaultLimit)
	log.Printf("Monitoring: Metrics=%v, LogLevel=%s", c.Monitoring.EnableMetrics, c.Monitoring.LogLevel)
	log.Printf("================================")
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if AppConfig == nil {
		log.Println("Configuration not loaded, loading now...")
		return Load()
	}
	return AppConfig
}

// MustLoad loads configuration and panics if validation fails
func MustLoad() *Config {
	config := Load()
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	return config
}

// ReloadConfig reloads configuration from environment
func ReloadConfig() *Config {
	log.Println("Reloading configuration...")
	return Load()
}
