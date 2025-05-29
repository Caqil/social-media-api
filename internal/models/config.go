// internal/models/config.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DynamicConfig represents all application configuration stored in database
type DynamicConfig struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Version   int                `json:"version" bson:"version"` // For config versioning
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy primitive.ObjectID `json:"updated_by" bson:"updated_by"`

	// Server Configuration
	Server ServerConfig `json:"server" bson:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database" bson:"database"`

	// JWT Configuration
	JWT JWTConfig `json:"jwt" bson:"jwt"`

	// Email Configuration
	Email EmailConfig `json:"email" bson:"email"`

	// Storage Configuration
	Storage StorageConfig `json:"storage" bson:"storage"`

	// External Services
	External ExternalConfig `json:"external" bson:"external"`

	// Security Configuration
	Security SecurityConfig `json:"security" bson:"security"`

	// Rate Limiting
	RateLimit RateLimitConfig `json:"rate_limit" bson:"rate_limit"`

	// Features
	Features FeatureConfig `json:"features" bson:"features"`

	// Content Policies
	Content ContentConfig `json:"content" bson:"content"`

	// Monitoring
	Monitoring MonitoringConfig `json:"monitoring" bson:"monitoring"`

	// Cache Configuration
	Cache CacheConfig `json:"cache" bson:"cache"`

	// Backup Configuration
	Backup BackupConfig `json:"backup" bson:"backup"`
}

type ServerConfig struct {
	Host            string        `json:"host" bson:"host"`
	Port            int           `json:"port" bson:"port"`
	Mode            string        `json:"mode" bson:"mode"` // debug, release, test
	ReadTimeout     time.Duration `json:"read_timeout" bson:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" bson:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" bson:"shutdown_timeout"`
	TrustedProxies  []string      `json:"trusted_proxies" bson:"trusted_proxies"`
	EnableHTTPS     bool          `json:"enable_https" bson:"enable_https"`
	CertFile        string        `json:"cert_file" bson:"cert_file"`
	KeyFile         string        `json:"key_file" bson:"key_file"`
	EnableGzip      bool          `json:"enable_gzip" bson:"enable_gzip"`
	MaxRequestSize  int64         `json:"max_request_size" bson:"max_request_size"`
	CORSOrigins     []string      `json:"cors_origins" bson:"cors_origins"`
	CORSMethods     []string      `json:"cors_methods" bson:"cors_methods"`
	CORSHeaders     []string      `json:"cors_headers" bson:"cors_headers"`
}

type DatabaseConfig struct {
	URI                string `json:"uri" bson:"uri"`
	DatabaseName       string `json:"database_name" bson:"database_name"`
	MaxPoolSize        int    `json:"max_pool_size" bson:"max_pool_size"`
	MinPoolSize        int    `json:"min_pool_size" bson:"min_pool_size"`
	ConnectTimeout     int    `json:"connect_timeout" bson:"connect_timeout"`
	EnableSSL          bool   `json:"enable_ssl" bson:"enable_ssl"`
	ReplicaSet         string `json:"replica_set" bson:"replica_set"`
	EnableLogging      bool   `json:"enable_logging" bson:"enable_logging"`
	SlowQueryThreshold int    `json:"slow_query_threshold" bson:"slow_query_threshold"`
}

type JWTConfig struct {
	SecretKey          string        `json:"secret_key" bson:"secret_key"`
	RefreshSecretKey   string        `json:"refresh_secret_key" bson:"refresh_secret_key"`
	AccessTokenExpiry  time.Duration `json:"access_token_expiry" bson:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry" bson:"refresh_token_expiry"`
	Issuer             string        `json:"issuer" bson:"issuer"`
	Audience           string        `json:"audience" bson:"audience"`
	EnableTokenRefresh bool          `json:"enable_token_refresh" bson:"enable_token_refresh"`
	MaxRefreshCount    int           `json:"max_refresh_count" bson:"max_refresh_count"`
	BlacklistEnabled   bool          `json:"blacklist_enabled" bson:"blacklist_enabled"`
}

type EmailConfig struct {
	SMTPHost       string `json:"smtp_host" bson:"smtp_host"`
	SMTPPort       int    `json:"smtp_port" bson:"smtp_port"`
	SMTPUser       string `json:"smtp_user" bson:"smtp_user"`
	SMTPPassword   string `json:"smtp_password" bson:"smtp_password"`
	FromEmail      string `json:"from_email" bson:"from_email"`
	FromName       string `json:"from_name" bson:"from_name"`
	ReplyToEmail   string `json:"reply_to_email" bson:"reply_to_email"`
	EnableTLS      bool   `json:"enable_tls" bson:"enable_tls"`
	EnableSSL      bool   `json:"enable_ssl" bson:"enable_ssl"`
	TemplateDir    string `json:"template_dir" bson:"template_dir"`
	MaxRetries     int    `json:"max_retries" bson:"max_retries"`
	RetryDelay     int    `json:"retry_delay" bson:"retry_delay"`
	EnableTracking bool   `json:"enable_tracking" bson:"enable_tracking"`
}

type StorageConfig struct {
	Provider     string   `json:"provider" bson:"provider"` // local, s3, gcs, azure
	LocalPath    string   `json:"local_path" bson:"local_path"`
	PublicURL    string   `json:"public_url" bson:"public_url"`
	MaxFileSize  int64    `json:"max_file_size" bson:"max_file_size"`
	AllowedTypes []string `json:"allowed_types" bson:"allowed_types"`
	EnableCDN    bool     `json:"enable_cdn" bson:"enable_cdn"`
	CDNBaseURL   string   `json:"cdn_base_url" bson:"cdn_base_url"`

	// AWS S3 Config
	S3Config S3Config `json:"s3_config" bson:"s3_config"`

	// Google Cloud Storage Config
	GCSConfig GCSConfig `json:"gcs_config" bson:"gcs_config"`

	// Azure Blob Config
	AzureConfig AzureConfig `json:"azure_config" bson:"azure_config"`
}

type S3Config struct {
	AccessKeyID     string `json:"access_key_id" bson:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" bson:"secret_access_key"`
	Region          string `json:"region" bson:"region"`
	Bucket          string `json:"bucket" bson:"bucket"`
	Endpoint        string `json:"endpoint" bson:"endpoint"`
	UseSSL          bool   `json:"use_ssl" bson:"use_ssl"`
}

type GCSConfig struct {
	ProjectID       string `json:"project_id" bson:"project_id"`
	Bucket          string `json:"bucket" bson:"bucket"`
	CredentialsFile string `json:"credentials_file" bson:"credentials_file"`
}

type AzureConfig struct {
	AccountName   string `json:"account_name" bson:"account_name"`
	AccountKey    string `json:"account_key" bson:"account_key"`
	ContainerName string `json:"container_name" bson:"container_name"`
}

type ExternalConfig struct {
	// Push Notifications
	FirebaseServerKey string `json:"firebase_server_key" bson:"firebase_server_key"`
	FCMSenderID       string `json:"fcm_sender_id" bson:"fcm_sender_id"`

	// Apple Push Notifications
	APNSKeyID      string `json:"apns_key_id" bson:"apns_key_id"`
	APNSTeamID     string `json:"apns_team_id" bson:"apns_team_id"`
	APNSBundleID   string `json:"apns_bundle_id" bson:"apns_bundle_id"`
	APNSKeyFile    string `json:"apns_key_file" bson:"apns_key_file"`
	APNSProduction bool   `json:"apns_production" bson:"apns_production"`

	// Social Media APIs
	TwitterAPIKey      string `json:"twitter_api_key" bson:"twitter_api_key"`
	TwitterAPISecret   string `json:"twitter_api_secret" bson:"twitter_api_secret"`
	FacebookAppID      string `json:"facebook_app_id" bson:"facebook_app_id"`
	FacebookAppSecret  string `json:"facebook_app_secret" bson:"facebook_app_secret"`
	GoogleClientID     string `json:"google_client_id" bson:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret" bson:"google_client_secret"`

	// Payment Processing
	StripePublishableKey string `json:"stripe_publishable_key" bson:"stripe_publishable_key"`
	StripeSecretKey      string `json:"stripe_secret_key" bson:"stripe_secret_key"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret" bson:"stripe_webhook_secret"`

	// Analytics
	GoogleAnalyticsID string `json:"google_analytics_id" bson:"google_analytics_id"`
	MixpanelToken     string `json:"mixpanel_token" bson:"mixpanel_token"`

	// Monitoring
	SentryDSN     string `json:"sentry_dsn" bson:"sentry_dsn"`
	DatadogAPIKey string `json:"datadog_api_key" bson:"datadog_api_key"`

	// AI/ML Services
	OpenAIAPIKey     string `json:"openai_api_key" bson:"openai_api_key"`
	ModerationAPIKey string `json:"moderation_api_key" bson:"moderation_api_key"`
}

type SecurityConfig struct {
	EnableHTTPS                bool     `json:"enable_https" bson:"enable_https"`
	HSTSMaxAge                 int      `json:"hsts_max_age" bson:"hsts_max_age"`
	EnableCSRF                 bool     `json:"enable_csrf" bson:"enable_csrf"`
	CSRFSecret                 string   `json:"csrf_secret" bson:"csrf_secret"`
	PasswordMinLength          int      `json:"password_min_length" bson:"password_min_length"`
	PasswordRequireUpper       bool     `json:"password_require_upper" bson:"password_require_upper"`
	PasswordRequireLower       bool     `json:"password_require_lower" bson:"password_require_lower"`
	PasswordRequireDigit       bool     `json:"password_require_digit" bson:"password_require_digit"`
	PasswordRequireSymbol      bool     `json:"password_require_symbol" bson:"password_require_symbol"`
	MaxLoginAttempts           int      `json:"max_login_attempts" bson:"max_login_attempts"`
	LockoutDuration            int      `json:"lockout_duration" bson:"lockout_duration"`
	SessionTimeout             int      `json:"session_timeout" bson:"session_timeout"`
	IPWhitelist                []string `json:"ip_whitelist" bson:"ip_whitelist"`
	IPBlacklist                []string `json:"ip_blacklist" bson:"ip_blacklist"`
	EnableBruteForceProtection bool     `json:"enable_brute_force_protection" bson:"enable_brute_force_protection"`
	Enable2FA                  bool     `json:"enable_2fa" bson:"enable_2fa"`
	Require2FA                 bool     `json:"require_2fa" bson:"require_2fa"`
}

type RateLimitConfig struct {
	Enabled            bool           `json:"enabled" bson:"enabled"`
	DefaultLimit       int            `json:"default_limit" bson:"default_limit"`
	DefaultWindow      time.Duration  `json:"default_window" bson:"default_window"`
	LoginLimit         int            `json:"login_limit" bson:"login_limit"`
	LoginWindow        time.Duration  `json:"login_window" bson:"login_window"`
	RegistrationLimit  int            `json:"registration_limit" bson:"registration_limit"`
	RegistrationWindow time.Duration  `json:"registration_window" bson:"registration_window"`
	PostLimit          int            `json:"post_limit" bson:"post_limit"`
	PostWindow         time.Duration  `json:"post_window" bson:"post_window"`
	CommentLimit       int            `json:"comment_limit" bson:"comment_limit"`
	CommentWindow      time.Duration  `json:"comment_window" bson:"comment_window"`
	MessageLimit       int            `json:"message_limit" bson:"message_limit"`
	MessageWindow      time.Duration  `json:"message_window" bson:"message_window"`
	FollowLimit        int            `json:"follow_limit" bson:"follow_limit"`
	FollowWindow       time.Duration  `json:"follow_window" bson:"follow_window"`
	LikeLimit          int            `json:"like_limit" bson:"like_limit"`
	LikeWindow         time.Duration  `json:"like_window" bson:"like_window"`
	SearchLimit        int            `json:"search_limit" bson:"search_limit"`
	SearchWindow       time.Duration  `json:"search_window" bson:"search_window"`
	UploadLimit        int            `json:"upload_limit" bson:"upload_limit"`
	UploadWindow       time.Duration  `json:"upload_window" bson:"upload_window"`
	IPBasedLimiting    bool           `json:"ip_based_limiting" bson:"ip_based_limiting"`
	UserBasedLimiting  bool           `json:"user_based_limiting" bson:"user_based_limiting"`
	CustomLimits       map[string]int `json:"custom_limits" bson:"custom_limits"`
	WhitelistIPs       []string       `json:"whitelist_ips" bson:"whitelist_ips"`
	WhitelistUsers     []string       `json:"whitelist_users" bson:"whitelist_users"`
}

type FeatureConfig struct {
	EnableRegistration       bool   `json:"enable_registration" bson:"enable_registration"`
	EnablePasswordReset      bool   `json:"enable_password_reset" bson:"enable_password_reset"`
	EnableEmailVerification  bool   `json:"enable_email_verification" bson:"enable_email_verification"`
	EnableUserProfiles       bool   `json:"enable_user_profiles" bson:"enable_user_profiles"`
	EnablePosts              bool   `json:"enable_posts" bson:"enable_posts"`
	EnableComments           bool   `json:"enable_comments" bson:"enable_comments"`
	EnableStories            bool   `json:"enable_stories" bson:"enable_stories"`
	EnableGroups             bool   `json:"enable_groups" bson:"enable_groups"`
	EnableMessaging          bool   `json:"enable_messaging" bson:"enable_messaging"`
	EnableNotifications      bool   `json:"enable_notifications" bson:"enable_notifications"`
	EnableLikes              bool   `json:"enable_likes" bson:"enable_likes"`
	EnableSharing            bool   `json:"enable_sharing" bson:"enable_sharing"`
	EnableFollowing          bool   `json:"enable_following" bson:"enable_following"`
	EnableSearch             bool   `json:"enable_search" bson:"enable_search"`
	EnableReporting          bool   `json:"enable_reporting" bson:"enable_reporting"`
	EnableModeration         bool   `json:"enable_moderation" bson:"enable_moderation"`
	EnableAIModeration       bool   `json:"enable_ai_moderation" bson:"enable_ai_moderation"`
	EnableAnalytics          bool   `json:"enable_analytics" bson:"enable_analytics"`
	EnableBehaviorTracking   bool   `json:"enable_behavior_tracking" bson:"enable_behavior_tracking"`
	EnableRecommendations    bool   `json:"enable_recommendations" bson:"enable_recommendations"`
	EnablePushNotifications  bool   `json:"enable_push_notifications" bson:"enable_push_notifications"`
	EnableEmailNotifications bool   `json:"enable_email_notifications" bson:"enable_email_notifications"`
	EnableSMSNotifications   bool   `json:"enable_sms_notifications" bson:"enable_sms_notifications"`
	EnableWebhooks           bool   `json:"enable_webhooks" bson:"enable_webhooks"`
	EnableAPI                bool   `json:"enable_api" bson:"enable_api"`
	EnableGraphQL            bool   `json:"enable_graphql" bson:"enable_graphql"`
	EnableWebSocket          bool   `json:"enable_websocket" bson:"enable_websocket"`
	EnableMobileApp          bool   `json:"enable_mobile_app" bson:"enable_mobile_app"`
	EnableDesktopApp         bool   `json:"enable_desktop_app" bson:"enable_desktop_app"`
	EnablePWA                bool   `json:"enable_pwa" bson:"enable_pwa"`
	EnableDarkMode           bool   `json:"enable_dark_mode" bson:"enable_dark_mode"`
	EnableMultiLanguage      bool   `json:"enable_multi_language" bson:"enable_multi_language"`
	EnableGuestMode          bool   `json:"enable_guest_mode" bson:"enable_guest_mode"`
	EnableInviteOnly         bool   `json:"enable_invite_only" bson:"enable_invite_only"`
	EnablePrivateMode        bool   `json:"enable_private_mode" bson:"enable_private_mode"`
	EnableMaintenanceMode    bool   `json:"enable_maintenance_mode" bson:"enable_maintenance_mode"`
	MaintenanceMessage       string `json:"maintenance_message" bson:"maintenance_message"`
}

type ContentConfig struct {
	MaxPostLength           int      `json:"max_post_length" bson:"max_post_length"`
	MaxCommentLength        int      `json:"max_comment_length" bson:"max_comment_length"`
	MaxStoryLength          int      `json:"max_story_length" bson:"max_story_length"`
	MaxBioLength            int      `json:"max_bio_length" bson:"max_bio_length"`
	MaxUsernameLength       int      `json:"max_username_length" bson:"max_username_length"`
	MinUsernameLength       int      `json:"min_username_length" bson:"min_username_length"`
	MaxHashtagsPerPost      int      `json:"max_hashtags_per_post" bson:"max_hashtags_per_post"`
	MaxMentionsPerPost      int      `json:"max_mentions_per_post" bson:"max_mentions_per_post"`
	MaxLinksPerPost         int      `json:"max_links_per_post" bson:"max_links_per_post"`
	AllowHTML               bool     `json:"allow_html" bson:"allow_html"`
	AllowMarkdown           bool     `json:"allow_markdown" bson:"allow_markdown"`
	AutoLinkURLs            bool     `json:"auto_link_urls" bson:"auto_link_urls"`
	AutoLinkHashtags        bool     `json:"auto_link_hashtags" bson:"auto_link_hashtags"`
	AutoLinkMentions        bool     `json:"auto_link_mentions" bson:"auto_link_mentions"`
	EnableTextFormatting    bool     `json:"enable_text_formatting" bson:"enable_text_formatting"`
	EnableEmojis            bool     `json:"enable_emojis" bson:"enable_emojis"`
	EnableCustomEmojis      bool     `json:"enable_custom_emojis" bson:"enable_custom_emojis"`
	EnableGIFs              bool     `json:"enable_gifs" bson:"enable_gifs"`
	EnableStickers          bool     `json:"enable_stickers" bson:"enable_stickers"`
	EnablePolls             bool     `json:"enable_polls" bson:"enable_polls"`
	EnableScheduledPosts    bool     `json:"enable_scheduled_posts" bson:"enable_scheduled_posts"`
	EnableDraftPosts        bool     `json:"enable_draft_posts" bson:"enable_draft_posts"`
	EnableContentWarnings   bool     `json:"enable_content_warnings" bson:"enable_content_warnings"`
	EnableSpoilerTags       bool     `json:"enable_spoiler_tags" bson:"enable_spoiler_tags"`
	RequireContentWarnings  bool     `json:"require_content_warnings" bson:"require_content_warnings"`
	ModerationLevel         string   `json:"moderation_level" bson:"moderation_level"` // none, low, medium, high, strict
	AutoModerationEnabled   bool     `json:"auto_moderation_enabled" bson:"auto_moderation_enabled"`
	ProfanityFilterEnabled  bool     `json:"profanity_filter_enabled" bson:"profanity_filter_enabled"`
	SpamDetectionEnabled    bool     `json:"spam_detection_enabled" bson:"spam_detection_enabled"`
	LinkPreviewEnabled      bool     `json:"link_preview_enabled" bson:"link_preview_enabled"`
	ImageCompressionEnabled bool     `json:"image_compression_enabled" bson:"image_compression_enabled"`
	VideoCompressionEnabled bool     `json:"video_compression_enabled" bson:"video_compression_enabled"`
	ThumbnailGeneration     bool     `json:"thumbnail_generation" bson:"thumbnail_generation"`
	AllowedDomains          []string `json:"allowed_domains" bson:"allowed_domains"`
	BlockedDomains          []string `json:"blocked_domains" bson:"blocked_domains"`
	BannedWords             []string `json:"banned_words" bson:"banned_words"`
	ContentRetentionDays    int      `json:"content_retention_days" bson:"content_retention_days"`
	EnableContentBackup     bool     `json:"enable_content_backup" bson:"enable_content_backup"`
	EnableContentVersioning bool     `json:"enable_content_versioning" bson:"enable_content_versioning"`
}

type MonitoringConfig struct {
	EnableRequestLogging     bool    `json:"enable_request_logging" bson:"enable_request_logging"`
	EnableErrorLogging       bool    `json:"enable_error_logging" bson:"enable_error_logging"`
	EnablePerformanceLogging bool    `json:"enable_performance_logging" bson:"enable_performance_logging"`
	EnableSecurityLogging    bool    `json:"enable_security_logging" bson:"enable_security_logging"`
	LogLevel                 string  `json:"log_level" bson:"log_level"`   // debug, info, warn, error
	LogFormat                string  `json:"log_format" bson:"log_format"` // json, text
	LogRotation              bool    `json:"log_rotation" bson:"log_rotation"`
	LogRetentionDays         int     `json:"log_retention_days" bson:"log_retention_days"`
	EnableMetrics            bool    `json:"enable_metrics" bson:"enable_metrics"`
	MetricsEndpoint          string  `json:"metrics_endpoint" bson:"metrics_endpoint"`
	EnableHealthChecks       bool    `json:"enable_health_checks" bson:"enable_health_checks"`
	HealthCheckInterval      int     `json:"health_check_interval" bson:"health_check_interval"`
	EnableAlerts             bool    `json:"enable_alerts" bson:"enable_alerts"`
	AlertWebhookURL          string  `json:"alert_webhook_url" bson:"alert_webhook_url"`
	SlowQueryThreshold       int     `json:"slow_query_threshold" bson:"slow_query_threshold"`
	ErrorRateThreshold       float64 `json:"error_rate_threshold" bson:"error_rate_threshold"`
	ResponseTimeThreshold    int     `json:"response_time_threshold" bson:"response_time_threshold"`
}

type CacheConfig struct {
	Enabled            bool          `json:"enabled" bson:"enabled"`
	Provider           string        `json:"provider" bson:"provider"` // redis, memory, memcached
	RedisURL           string        `json:"redis_url" bson:"redis_url"`
	RedisPassword      string        `json:"redis_password" bson:"redis_password"`
	RedisDB            int           `json:"redis_db" bson:"redis_db"`
	DefaultTTL         time.Duration `json:"default_ttl" bson:"default_ttl"`
	UserCacheTTL       time.Duration `json:"user_cache_ttl" bson:"user_cache_ttl"`
	PostCacheTTL       time.Duration `json:"post_cache_ttl" bson:"post_cache_ttl"`
	FeedCacheTTL       time.Duration `json:"feed_cache_ttl" bson:"feed_cache_ttl"`
	SearchCacheTTL     time.Duration `json:"search_cache_ttl" bson:"search_cache_ttl"`
	MediaCacheTTL      time.Duration `json:"media_cache_ttl" bson:"media_cache_ttl"`
	SessionCacheTTL    time.Duration `json:"session_cache_ttl" bson:"session_cache_ttl"`
	MaxMemorySize      int64         `json:"max_memory_size" bson:"max_memory_size"`
	EvictionPolicy     string        `json:"eviction_policy" bson:"eviction_policy"`
	EnableCompression  bool          `json:"enable_compression" bson:"enable_compression"`
	EnableEncryption   bool          `json:"enable_encryption" bson:"enable_encryption"`
	ConnectionPoolSize int           `json:"connection_pool_size" bson:"connection_pool_size"`
	ConnectionTimeout  time.Duration `json:"connection_timeout" bson:"connection_timeout"`
	ReadTimeout        time.Duration `json:"read_timeout" bson:"read_timeout"`
	WriteTimeout       time.Duration `json:"write_timeout" bson:"write_timeout"`
}

type BackupConfig struct {
	Enabled             bool          `json:"enabled" bson:"enabled"`
	Provider            string        `json:"provider" bson:"provider"` // local, s3, gcs, azure
	Schedule            string        `json:"schedule" bson:"schedule"` // cron expression
	RetentionDays       int           `json:"retention_days" bson:"retention_days"`
	BackupLocation      string        `json:"backup_location" bson:"backup_location"`
	IncludeUserData     bool          `json:"include_user_data" bson:"include_user_data"`
	IncludeMedia        bool          `json:"include_media" bson:"include_media"`
	IncludeSystemData   bool          `json:"include_system_data" bson:"include_system_data"`
	EnableCompression   bool          `json:"enable_compression" bson:"enable_compression"`
	EnableEncryption    bool          `json:"enable_encryption" bson:"enable_encryption"`
	EncryptionKey       string        `json:"encryption_key" bson:"encryption_key"`
	MaxBackupSize       int64         `json:"max_backup_size" bson:"max_backup_size"`
	NotificationEmail   string        `json:"notification_email" bson:"notification_email"`
	EnableNotifications bool          `json:"enable_notifications" bson:"enable_notifications"`
	BackupTimeout       time.Duration `json:"backup_timeout" bson:"backup_timeout"`
	ParallelBackups     int           `json:"parallel_backups" bson:"parallel_backups"`
	IncrementalBackups  bool          `json:"incremental_backups" bson:"incremental_backups"`
	BackupVerification  bool          `json:"backup_verification" bson:"backup_verification"`
}

// ConfigVersion tracks configuration changes
type ConfigVersion struct {
	ID           primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Version      int                    `json:"version" bson:"version"`
	Config       DynamicConfig          `json:"config" bson:"config"`
	ChangedBy    primitive.ObjectID     `json:"changed_by" bson:"changed_by"`
	ChangedAt    time.Time              `json:"changed_at" bson:"changed_at"`
	ChangeReason string                 `json:"change_reason" bson:"change_reason"`
	Changes      map[string]interface{} `json:"changes" bson:"changes"`
	RollbackData map[string]interface{} `json:"rollback_data" bson:"rollback_data"`
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *DynamicConfig {
	return &DynamicConfig{
		ID:        primitive.NewObjectID(),
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),

		Server: ServerConfig{
			Host:            "localhost",
			Port:            8080,
			Mode:            "debug",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			TrustedProxies:  []string{},
			EnableHTTPS:     false,
			EnableGzip:      true,
			MaxRequestSize:  32 << 20, // 32MB
			CORSOrigins:     []string{"*"},
			CORSMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CORSHeaders:     []string{"*"},
		},

		Database: DatabaseConfig{
			URI:                "mongodb://localhost:27017",
			DatabaseName:       "social_media",
			MaxPoolSize:        100,
			MinPoolSize:        10,
			ConnectTimeout:     10,
			EnableSSL:          false,
			EnableLogging:      true,
			SlowQueryThreshold: 1000,
		},

		JWT: JWTConfig{
			SecretKey:          "your-secret-key-change-this",
			RefreshSecretKey:   "your-refresh-secret-key-change-this",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "social-media-api",
			Audience:           "social-media-users",
			EnableTokenRefresh: true,
			MaxRefreshCount:    5,
			BlacklistEnabled:   true,
		},

		Email: EmailConfig{
			SMTPHost:       "smtp.gmail.com",
			SMTPPort:       587,
			FromEmail:      "noreply@example.com",
			FromName:       "Social Media API",
			EnableTLS:      true,
			EnableSSL:      false,
			MaxRetries:     3,
			RetryDelay:     5,
			EnableTracking: false,
		},

		Storage: StorageConfig{
			Provider:     "local",
			LocalPath:    "./uploads",
			PublicURL:    "http://localhost:8080/uploads",
			MaxFileSize:  10 << 20, // 10MB
			AllowedTypes: []string{"jpg", "jpeg", "png", "gif", "mp4", "mov", "pdf"},
			EnableCDN:    false,
		},

		Security: SecurityConfig{
			EnableHTTPS:                false,
			HSTSMaxAge:                 31536000,
			EnableCSRF:                 true,
			CSRFSecret:                 "csrf-secret-change-this",
			PasswordMinLength:          8,
			PasswordRequireUpper:       true,
			PasswordRequireLower:       true,
			PasswordRequireDigit:       true,
			PasswordRequireSymbol:      false,
			MaxLoginAttempts:           5,
			LockoutDuration:            300,
			SessionTimeout:             3600,
			EnableBruteForceProtection: true,
			Enable2FA:                  false,
			Require2FA:                 false,
		},

		RateLimit: RateLimitConfig{
			Enabled:            true,
			DefaultLimit:       100,
			DefaultWindow:      time.Hour,
			LoginLimit:         5,
			LoginWindow:        5 * time.Minute,
			RegistrationLimit:  3,
			RegistrationWindow: time.Hour,
			PostLimit:          10,
			PostWindow:         time.Hour,
			CommentLimit:       50,
			CommentWindow:      time.Hour,
			MessageLimit:       100,
			MessageWindow:      time.Hour,
			FollowLimit:        20,
			FollowWindow:       time.Hour,
			LikeLimit:          100,
			LikeWindow:         time.Hour,
			SearchLimit:        50,
			SearchWindow:       time.Hour,
			UploadLimit:        5,
			UploadWindow:       time.Hour,
			IPBasedLimiting:    true,
			UserBasedLimiting:  true,
		},

		Features: FeatureConfig{
			EnableRegistration:       true,
			EnablePasswordReset:      true,
			EnableEmailVerification:  true,
			EnableUserProfiles:       true,
			EnablePosts:              true,
			EnableComments:           true,
			EnableStories:            true,
			EnableGroups:             true,
			EnableMessaging:          true,
			EnableNotifications:      true,
			EnableLikes:              true,
			EnableSharing:            true,
			EnableFollowing:          true,
			EnableSearch:             true,
			EnableReporting:          true,
			EnableModeration:         true,
			EnableAIModeration:       false,
			EnableAnalytics:          true,
			EnableBehaviorTracking:   true,
			EnableRecommendations:    true,
			EnablePushNotifications:  true,
			EnableEmailNotifications: true,
			EnableSMSNotifications:   false,
			EnableWebhooks:           false,
			EnableAPI:                true,
			EnableGraphQL:            false,
			EnableWebSocket:          true,
			EnableMobileApp:          true,
			EnableDesktopApp:         false,
			EnablePWA:                true,
			EnableDarkMode:           true,
			EnableMultiLanguage:      false,
			EnableGuestMode:          true,
			EnableInviteOnly:         false,
			EnablePrivateMode:        false,
			EnableMaintenanceMode:    false,
			MaintenanceMessage:       "System maintenance in progress. Please try again later.",
		},

		Content: ContentConfig{
			MaxPostLength:           280,
			MaxCommentLength:        1000,
			MaxStoryLength:          500,
			MaxBioLength:            500,
			MaxUsernameLength:       30,
			MinUsernameLength:       3,
			MaxHashtagsPerPost:      10,
			MaxMentionsPerPost:      10,
			MaxLinksPerPost:         5,
			AllowHTML:               false,
			AllowMarkdown:           true,
			AutoLinkURLs:            true,
			AutoLinkHashtags:        true,
			AutoLinkMentions:        true,
			EnableTextFormatting:    true,
			EnableEmojis:            true,
			EnableCustomEmojis:      false,
			EnableGIFs:              true,
			EnableStickers:          false,
			EnablePolls:             true,
			EnableScheduledPosts:    false,
			EnableDraftPosts:        true,
			EnableContentWarnings:   false,
			EnableSpoilerTags:       false,
			RequireContentWarnings:  false,
			ModerationLevel:         "medium",
			AutoModerationEnabled:   true,
			ProfanityFilterEnabled:  true,
			SpamDetectionEnabled:    true,
			LinkPreviewEnabled:      true,
			ImageCompressionEnabled: true,
			VideoCompressionEnabled: true,
			ThumbnailGeneration:     true,
			ContentRetentionDays:    365,
			EnableContentBackup:     true,
			EnableContentVersioning: false,
		},

		Monitoring: MonitoringConfig{
			EnableRequestLogging:     true,
			EnableErrorLogging:       true,
			EnablePerformanceLogging: true,
			EnableSecurityLogging:    true,
			LogLevel:                 "info",
			LogFormat:                "json",
			LogRotation:              true,
			LogRetentionDays:         30,
			EnableMetrics:            true,
			MetricsEndpoint:          "/metrics",
			EnableHealthChecks:       true,
			HealthCheckInterval:      60,
			EnableAlerts:             false,
			SlowQueryThreshold:       1000,
			ErrorRateThreshold:       5.0,
			ResponseTimeThreshold:    1000,
		},

		Cache: CacheConfig{
			Enabled:            true,
			Provider:           "memory",
			DefaultTTL:         time.Hour,
			UserCacheTTL:       30 * time.Minute,
			PostCacheTTL:       15 * time.Minute,
			FeedCacheTTL:       5 * time.Minute,
			SearchCacheTTL:     10 * time.Minute,
			MediaCacheTTL:      24 * time.Hour,
			SessionCacheTTL:    30 * time.Minute,
			MaxMemorySize:      100 << 20, // 100MB
			EvictionPolicy:     "lru",
			EnableCompression:  false,
			EnableEncryption:   false,
			ConnectionPoolSize: 10,
			ConnectionTimeout:  5 * time.Second,
			ReadTimeout:        3 * time.Second,
			WriteTimeout:       3 * time.Second,
		},

		Backup: BackupConfig{
			Enabled:             false,
			Provider:            "local",
			Schedule:            "0 2 * * *", // Daily at 2 AM
			RetentionDays:       30,
			BackupLocation:      "./backups",
			IncludeUserData:     true,
			IncludeMedia:        true,
			IncludeSystemData:   true,
			EnableCompression:   true,
			EnableEncryption:    false,
			MaxBackupSize:       1 << 30, // 1GB
			EnableNotifications: false,
			BackupTimeout:       2 * time.Hour,
			ParallelBackups:     1,
			IncrementalBackups:  false,
			BackupVerification:  true,
		},
	}
}
