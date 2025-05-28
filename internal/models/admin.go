// internal/models/admin.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Admin Dashboard Models

type AdminDashboardStats struct {
	TotalUsers        int64                `json:"total_users"`
	ActiveUsers       int64                `json:"active_users"`
	NewUsers          int64                `json:"new_users"`
	TotalPosts        int64                `json:"total_posts"`
	TotalComments     int64                `json:"total_comments"`
	TotalStories      int64                `json:"total_stories"`
	TotalGroups       int64                `json:"total_groups"`
	TotalReports      int64                `json:"total_reports"`
	PendingReports    int64                `json:"pending_reports"`
	SystemHealth      SystemHealthStatus   `json:"system_health"`
	TopHashtags       []HashtagStats       `json:"top_hashtags"`
	UserGrowth        []UserGrowthData     `json:"user_growth"`
	ContentGrowth     []ContentGrowthData  `json:"content_growth"`
	EngagementMetrics EngagementMetrics    `json:"engagement_metrics"`
	GeographicData    []GeographicStats    `json:"geographic_data"`
	DeviceStats       []DeviceStats        `json:"device_stats"`
	PlatformMetrics   PlatformMetrics      `json:"platform_metrics"`
	ModerationQueue   ModerationQueueStats `json:"moderation_queue"`
	RevenueMetrics    RevenueMetrics       `json:"revenue_metrics"`
}

type SystemHealthStatus struct {
	Status            string    `json:"status"` // healthy, warning, critical
	CPUUsage          float64   `json:"cpu_usage"`
	MemoryUsage       float64   `json:"memory_usage"`
	DiskUsage         float64   `json:"disk_usage"`
	ActiveConnections int64     `json:"active_connections"`
	ResponseTime      float64   `json:"response_time_ms"`
	ErrorRate         float64   `json:"error_rate"`
	DatabaseHealth    string    `json:"database_health"`
	CacheHealth       string    `json:"cache_health"`
	LastHealthCheck   time.Time `json:"last_health_check"`
}

type UserGrowthData struct {
	Date      string  `json:"date"`
	NewUsers  int64   `json:"new_users"`
	Active    int64   `json:"active_users"`
	Retention float64 `json:"retention_rate"`
}

type ContentGrowthData struct {
	Date     string `json:"date"`
	Posts    int64  `json:"posts"`
	Comments int64  `json:"comments"`
	Stories  int64  `json:"stories"`
	Groups   int64  `json:"groups"`
}

type EngagementMetrics struct {
	AverageSessionDuration float64 `json:"avg_session_duration"`
	AveragePostsPerUser    float64 `json:"avg_posts_per_user"`
	AverageCommentsPerPost float64 `json:"avg_comments_per_post"`
	AverageLikesPerPost    float64 `json:"avg_likes_per_post"`
	AverageSharesPerPost   float64 `json:"avg_shares_per_post"`
	DailyActiveUsers       int64   `json:"daily_active_users"`
	WeeklyActiveUsers      int64   `json:"weekly_active_users"`
	MonthlyActiveUsers     int64   `json:"monthly_active_users"`
	UserRetentionRate      float64 `json:"user_retention_rate"`
	ContentEngagementRate  float64 `json:"content_engagement_rate"`
}

type GeographicStats struct {
	Country   string `json:"country"`
	UserCount int64  `json:"user_count"`
	PostCount int64  `json:"post_count"`
}

type DeviceStats struct {
	DeviceType string  `json:"device_type"`
	UserCount  int64   `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

type HashtagStats struct {
	Hashtag   string `json:"hashtag"`
	PostCount int64  `json:"post_count"`
	UserCount int64  `json:"user_count"`
}

type PlatformMetrics struct {
	StorageUsed          int64   `json:"storage_used_gb"`
	BandwidthUsed        int64   `json:"bandwidth_used_gb"`
	APIRequestsPerDay    int64   `json:"api_requests_per_day"`
	AverageResponseTime  float64 `json:"avg_response_time_ms"`
	UptimePercentage     float64 `json:"uptime_percentage"`
	ErrorRatePercentage  float64 `json:"error_rate_percentage"`
	ActiveWebSocketConns int64   `json:"active_websocket_connections"`
	QueuedJobs           int64   `json:"queued_jobs"`
	ProcessedJobs        int64   `json:"processed_jobs_today"`
}

type ModerationQueueStats struct {
	PendingReports     int64 `json:"pending_reports"`
	PendingPosts       int64 `json:"pending_posts"`
	PendingComments    int64 `json:"pending_comments"`
	PendingUsers       int64 `json:"pending_users"`
	AutoFlaggedContent int64 `json:"auto_flagged_content"`
	EscalatedIssues    int64 `json:"escalated_issues"`
}

type RevenueMetrics struct {
	TotalRevenue        float64 `json:"total_revenue"`
	MonthlyRevenue      float64 `json:"monthly_revenue"`
	AdvertisingRevenue  float64 `json:"advertising_revenue"`
	SubscriptionRevenue float64 `json:"subscription_revenue"`
	RevenueGrowthRate   float64 `json:"revenue_growth_rate"`
}

// User Management Models

type AdminUserOverview struct {
	ID             primitive.ObjectID `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	FirstName      string             `json:"first_name"`
	LastName       string             `json:"last_name"`
	Status         UserStatus         `json:"status"`
	Role           UserRole           `json:"role"`
	IsVerified     bool               `json:"is_verified"`
	CreatedAt      time.Time          `json:"created_at"`
	LastLoginAt    *time.Time         `json:"last_login_at"`
	PostCount      int64              `json:"post_count"`
	FollowerCount  int64              `json:"follower_count"`
	FollowingCount int64              `json:"following_count"`
	ReportCount    int64              `json:"report_count"`
	ViolationCount int64              `json:"violation_count"`
	Location       string             `json:"location"`
	RegistrationIP string             `json:"registration_ip"`
	LastActiveIP   string             `json:"last_active_ip"`
	DeviceInfo     string             `json:"device_info"`
	RiskScore      float64            `json:"risk_score"`
	AccountValue   float64            `json:"account_value"`
}

type AdminUserDetail struct {
	User              AdminUserOverview      `json:"user"`
	ActivitySummary   UserActivitySummary    `json:"activity_summary"`
	RecentPosts       []AdminPostOverview    `json:"recent_posts"`
	RecentComments    []AdminCommentOverview `json:"recent_comments"`
	SecurityEvents    []SecurityEvent        `json:"security_events"`
	ReportsAgainst    []AdminReportOverview  `json:"reports_against"`
	ReportsMade       []AdminReportOverview  `json:"reports_made"`
	ModerationHistory []ModerationAction     `json:"moderation_history"`
	LoginHistory      []LoginEvent           `json:"login_history"`
	DeviceHistory     []DeviceEvent          `json:"device_history"`
	EngagementMetrics UserEngagementMetrics  `json:"engagement_metrics"`
	RelationshipStats UserRelationshipStats  `json:"relationship_stats"`
	ContentStats      UserContentStats       `json:"content_stats"`
}

type UserActivitySummary struct {
	TotalSessions      int64     `json:"total_sessions"`
	AverageSessionTime float64   `json:"avg_session_time_minutes"`
	LastActiveAt       time.Time `json:"last_active_at"`
	PostsThisWeek      int64     `json:"posts_this_week"`
	CommentsThisWeek   int64     `json:"comments_this_week"`
	LikesGivenThisWeek int64     `json:"likes_given_this_week"`
	SharesThisWeek     int64     `json:"shares_this_week"`
	ReportsThisWeek    int64     `json:"reports_this_week"`
	LoginStreakDays    int64     `json:"login_streak_days"`
}

type SecurityEvent struct {
	ID          primitive.ObjectID `json:"id"`
	EventType   string             `json:"event_type"`
	Description string             `json:"description"`
	IPAddress   string             `json:"ip_address"`
	UserAgent   string             `json:"user_agent"`
	Timestamp   time.Time          `json:"timestamp"`
	Severity    string             `json:"severity"`
	Action      string             `json:"action"`
}

type LoginEvent struct {
	ID        primitive.ObjectID `json:"id"`
	IPAddress string             `json:"ip_address"`
	UserAgent string             `json:"user_agent"`
	Location  string             `json:"location"`
	Success   bool               `json:"success"`
	Timestamp time.Time          `json:"timestamp"`
	DeviceID  string             `json:"device_id"`
}

type DeviceEvent struct {
	ID         primitive.ObjectID `json:"id"`
	DeviceType string             `json:"device_type"`
	DeviceID   string             `json:"device_id"`
	Browser    string             `json:"browser"`
	OS         string             `json:"os"`
	FirstSeen  time.Time          `json:"first_seen"`
	LastSeen   time.Time          `json:"last_seen"`
	IsActive   bool               `json:"is_active"`
}

type UserEngagementMetrics struct {
	EngagementScore       float64 `json:"engagement_score"`
	ContentQualityScore   float64 `json:"content_quality_score"`
	CommunityContribution float64 `json:"community_contribution"`
	InfluenceScore        float64 `json:"influence_score"`
	TrustScore            float64 `json:"trust_score"`
}

type UserRelationshipStats struct {
	MutualConnections   int64   `json:"mutual_connections"`
	FollowBackRate      float64 `json:"follow_back_rate"`
	BlockedByUsers      int64   `json:"blocked_by_users"`
	BlockedUsers        int64   `json:"blocked_users"`
	ReportedByUsers     int64   `json:"reported_by_users"`
	NetworkReachability int64   `json:"network_reachability"`
}

type UserContentStats struct {
	TotalViralPosts      int64   `json:"total_viral_posts"`
	AverageEngagement    float64 `json:"average_engagement"`
	TopPerformingPost    string  `json:"top_performing_post_id"`
	ContentDiversity     float64 `json:"content_diversity_score"`
	OriginalContentRatio float64 `json:"original_content_ratio"`
}

// Content Management Models

type AdminPostOverview struct {
	ID               primitive.ObjectID `json:"id"`
	AuthorID         primitive.ObjectID `json:"author_id"`
	AuthorUsername   string             `json:"author_username"`
	Content          string             `json:"content"`
	ContentType      string             `json:"content_type"`
	Status           string             `json:"status"`
	Visibility       string             `json:"visibility"`
	CreatedAt        time.Time          `json:"created_at"`
	LikesCount       int64              `json:"likes_count"`
	CommentsCount    int64              `json:"comments_count"`
	SharesCount      int64              `json:"shares_count"`
	ViewsCount       int64              `json:"views_count"`
	ReportsCount     int64              `json:"reports_count"`
	FlaggedAt        *time.Time         `json:"flagged_at"`
	FlagReason       string             `json:"flag_reason"`
	ModerationStatus string             `json:"moderation_status"`
	EngagementRate   float64            `json:"engagement_rate"`
	ViralityScore    float64            `json:"virality_score"`
	QualityScore     float64            `json:"quality_score"`
	RiskScore        float64            `json:"risk_score"`
}

type AdminCommentOverview struct {
	ID               primitive.ObjectID `json:"id"`
	PostID           primitive.ObjectID `json:"post_id"`
	AuthorID         primitive.ObjectID `json:"author_id"`
	AuthorUsername   string             `json:"author_username"`
	Content          string             `json:"content"`
	CreatedAt        time.Time          `json:"created_at"`
	LikesCount       int64              `json:"likes_count"`
	RepliesCount     int64              `json:"replies_count"`
	ReportsCount     int64              `json:"reports_count"`
	FlaggedAt        *time.Time         `json:"flagged_at"`
	FlagReason       string             `json:"flag_reason"`
	ModerationStatus string             `json:"moderation_status"`
	IsDeleted        bool               `json:"is_deleted"`
	RiskScore        float64            `json:"risk_score"`
}

type AdminReportOverview struct {
	ID          primitive.ObjectID  `json:"id"`
	ReporterID  primitive.ObjectID  `json:"reporter_id"`
	TargetID    primitive.ObjectID  `json:"target_id"`
	TargetType  string              `json:"target_type"`
	Reason      ReportReason        `json:"reason"`
	Status      string              `json:"status"`
	Priority    string              `json:"priority"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	AssignedTo  *primitive.ObjectID `json:"assigned_to"`
	ResolvedAt  *time.Time          `json:"resolved_at"`
	Resolution  string              `json:"resolution"`
	Category    string              `json:"category"`
	Severity    string              `json:"severity"`
	Description string              `json:"description"`
}

type ModerationAction struct {
	ID          primitive.ObjectID  `json:"id"`
	ModeratorID primitive.ObjectID  `json:"moderator_id"`
	TargetID    primitive.ObjectID  `json:"target_id"`
	TargetType  string              `json:"target_type"`
	Action      string              `json:"action"`
	Reason      string              `json:"reason"`
	Timestamp   time.Time           `json:"timestamp"`
	Duration    *time.Duration      `json:"duration,omitempty"`
	Note        string              `json:"note"`
	Status      string              `json:"status"`
	AppealID    *primitive.ObjectID `json:"appeal_id,omitempty"`
}

// System Management Models

type SystemConfiguration struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MaxPostLength          int                `json:"max_post_length" bson:"max_post_length"`
	MaxCommentLength       int                `json:"max_comment_length" bson:"max_comment_length"`
	MaxFileSize            int64              `json:"max_file_size_mb" bson:"max_file_size_mb"`
	AllowedFileTypes       []string           `json:"allowed_file_types" bson:"allowed_file_types"`
	RateLimitPosts         int                `json:"rate_limit_posts_per_hour" bson:"rate_limit_posts_per_hour"`
	RateLimitComments      int                `json:"rate_limit_comments_per_hour" bson:"rate_limit_comments_per_hour"`
	RateLimitMessages      int                `json:"rate_limit_messages_per_hour" bson:"rate_limit_messages_per_hour"`
	AutoModeration         bool               `json:"auto_moderation_enabled" bson:"auto_moderation_enabled"`
	RequireEmailVerify     bool               `json:"require_email_verification" bson:"require_email_verification"`
	AllowGuestViewing      bool               `json:"allow_guest_viewing" bson:"allow_guest_viewing"`
	MaintenanceMode        bool               `json:"maintenance_mode" bson:"maintenance_mode"`
	MaintenanceMessage     string             `json:"maintenance_message" bson:"maintenance_message"`
	RegistrationEnabled    bool               `json:"registration_enabled" bson:"registration_enabled"`
	InviteOnlyMode         bool               `json:"invite_only_mode" bson:"invite_only_mode"`
	ContentModerationLevel string             `json:"content_moderation_level" bson:"content_moderation_level"`
	SpamDetectionEnabled   bool               `json:"spam_detection_enabled" bson:"spam_detection_enabled"`
	AIContentFilter        bool               `json:"ai_content_filter_enabled" bson:"ai_content_filter_enabled"`
	DataRetentionDays      int                `json:"data_retention_days" bson:"data_retention_days"`
	BackupEnabled          bool               `json:"backup_enabled" bson:"backup_enabled"`
	BackupFrequency        string             `json:"backup_frequency" bson:"backup_frequency"`
	CDNEnabled             bool               `json:"cdn_enabled" bson:"cdn_enabled"`
	CacheEnabled           bool               `json:"cache_enabled" bson:"cache_enabled"`
	AnalyticsEnabled       bool               `json:"analytics_enabled" bson:"analytics_enabled"`
	UpdatedAt              time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy              primitive.ObjectID `json:"updated_by" bson:"updated_by"`
}

// Admin Activity Models

type AdminActivity struct {
	ID          primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	AdminID     primitive.ObjectID     `json:"admin_id" bson:"admin_id"`
	AdminName   string                 `json:"admin_name" bson:"admin_name"`
	Action      string                 `json:"action" bson:"action"`
	TargetType  string                 `json:"target_type" bson:"target_type"`
	TargetID    primitive.ObjectID     `json:"target_id" bson:"target_id"`
	Description string                 `json:"description" bson:"description"`
	IPAddress   string                 `json:"ip_address" bson:"ip_address"`
	UserAgent   string                 `json:"user_agent" bson:"user_agent"`
	Timestamp   time.Time              `json:"timestamp" bson:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata" bson:"metadata"`
	Severity    string                 `json:"severity" bson:"severity"`
	Category    string                 `json:"category" bson:"category"`
}

// Request/Response Models

type AdminUserFilter struct {
	Status         string    `json:"status"`
	Role           string    `json:"role"`
	IsVerified     *bool     `json:"is_verified"`
	CreatedAfter   time.Time `json:"created_after"`
	CreatedBefore  time.Time `json:"created_before"`
	LastLoginAfter time.Time `json:"last_login_after"`
	MinPostCount   int64     `json:"min_post_count"`
	MaxPostCount   int64     `json:"max_post_count"`
	MinRiskScore   float64   `json:"min_risk_score"`
	MaxRiskScore   float64   `json:"max_risk_score"`
	Location       string    `json:"location"`
	SortBy         string    `json:"sort_by"`
	SortOrder      string    `json:"sort_order"`
}

type AdminUserActionRequest struct {
	UserIDs  []string `json:"user_ids" binding:"required"`
	Action   string   `json:"action" binding:"required"` // suspend, unsuspend, verify, unverify, delete, warn
	Reason   string   `json:"reason" binding:"required"`
	Duration *string  `json:"duration,omitempty"` // for temporary actions
	Note     string   `json:"note"`
}

type AdminContentFilter struct {
	AuthorID         string    `json:"author_id"`
	ContentType      string    `json:"content_type"`
	Status           string    `json:"status"`
	ModerationStatus string    `json:"moderation_status"`
	CreatedAfter     time.Time `json:"created_after"`
	CreatedBefore    time.Time `json:"created_before"`
	MinReports       int64     `json:"min_reports"`
	MinEngagement    float64   `json:"min_engagement"`
	MaxEngagement    float64   `json:"max_engagement"`
	IsFlagged        *bool     `json:"is_flagged"`
	SortBy           string    `json:"sort_by"`
	SortOrder        string    `json:"sort_order"`
}

type AdminContentActionRequest struct {
	ContentIDs []string `json:"content_ids" binding:"required"`
	Action     string   `json:"action" binding:"required"` // approve, reject, delete, flag, unflag
	Reason     string   `json:"reason" binding:"required"`
	Note       string   `json:"note"`
}

type SystemMetricsResponse struct {
	CPUUsage           float64            `json:"cpu_usage"`
	MemoryUsage        float64            `json:"memory_usage"`
	DiskUsage          float64            `json:"disk_usage"`
	NetworkIO          NetworkIOStats     `json:"network_io"`
	DatabaseMetrics    DatabaseMetrics    `json:"database_metrics"`
	CacheMetrics       CacheMetrics       `json:"cache_metrics"`
	QueueMetrics       QueueMetrics       `json:"queue_metrics"`
	ErrorMetrics       ErrorMetrics       `json:"error_metrics"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
}

type NetworkIOStats struct {
	BytesReceived   int64 `json:"bytes_received"`
	BytesSent       int64 `json:"bytes_sent"`
	PacketsReceived int64 `json:"packets_received"`
	PacketsSent     int64 `json:"packets_sent"`
}

type DatabaseMetrics struct {
	ConnectionCount int64   `json:"connection_count"`
	ActiveQueries   int64   `json:"active_queries"`
	QueryLatency    float64 `json:"query_latency_ms"`
	DatabaseSize    int64   `json:"database_size_mb"`
	CollectionCount int64   `json:"collection_count"`
	IndexCount      int64   `json:"index_count"`
}

type CacheMetrics struct {
	HitRate         float64 `json:"hit_rate"`
	MissRate        float64 `json:"miss_rate"`
	CachedItemCount int64   `json:"cached_item_count"`
	CacheSize       int64   `json:"cache_size_mb"`
	EvictionCount   int64   `json:"eviction_count"`
}

type QueueMetrics struct {
	PendingJobs        int64   `json:"pending_jobs"`
	ProcessingJobs     int64   `json:"processing_jobs"`
	CompletedJobs      int64   `json:"completed_jobs"`
	FailedJobs         int64   `json:"failed_jobs"`
	AverageProcessTime float64 `json:"avg_process_time_ms"`
}

type ErrorMetrics struct {
	ErrorRate5xx   float64 `json:"error_rate_5xx"`
	ErrorRate4xx   float64 `json:"error_rate_4xx"`
	TotalErrors    int64   `json:"total_errors"`
	CriticalErrors int64   `json:"critical_errors"`
	DatabaseErrors int64   `json:"database_errors"`
	AuthErrors     int64   `json:"auth_errors"`
}

type PerformanceMetrics struct {
	AverageResponseTime float64 `json:"avg_response_time_ms"`
	P50ResponseTime     float64 `json:"p50_response_time_ms"`
	P95ResponseTime     float64 `json:"p95_response_time_ms"`
	P99ResponseTime     float64 `json:"p99_response_time_ms"`
	RequestsPerSecond   float64 `json:"requests_per_second"`
	ConcurrentUsers     int64   `json:"concurrent_users"`
}
