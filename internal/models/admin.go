// internal/models/admin.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==================== FILTER MODELS ====================

// UserFilter represents filters for user queries
type UserFilter struct {
	Status          string    `json:"status"`   // active, suspended, deleted, pending
	Role            string    `json:"role"`     // user, moderator, admin, super_admin
	Verified        bool      `json:"verified"` // true, false
	EmailVerified   bool      `json:"email_verified"`
	CreatedAfter    time.Time `json:"created_after"`
	CreatedBefore   time.Time `json:"created_before"`
	LastLoginAfter  time.Time `json:"last_login_after"`
	LastLoginBefore time.Time `json:"last_login_before"`
	MinPostsCount   int       `json:"min_posts_count"`
	MaxPostsCount   int       `json:"max_posts_count"`
	MinFollowers    int       `json:"min_followers"`
	MaxFollowers    int       `json:"max_followers"`
	Location        string    `json:"location"`
	Language        string    `json:"language"`
	AgeMin          int       `json:"age_min"`
	AgeMax          int       `json:"age_max"`
	Gender          string    `json:"gender"`
	HasReports      bool      `json:"has_reports"`
	SortBy          string    `json:"sort_by"`    // created_at, last_login, posts_count, followers_count
	SortOrder       string    `json:"sort_order"` // asc, desc
}

// ContentFilter represents filters for content queries
type ContentFilter struct {
	ContentType   string    `json:"content_type"` // post, comment, story, all
	Status        string    `json:"status"`       // published, deleted, hidden, reported
	AuthorID      string    `json:"author_id"`
	CreatedAfter  time.Time `json:"created_after"`
	CreatedBefore time.Time `json:"created_before"`
	MinLikes      int       `json:"min_likes"`
	MaxLikes      int       `json:"max_likes"`
	MinComments   int       `json:"min_comments"`
	MaxComments   int       `json:"max_comments"`
	MinViews      int       `json:"min_views"`
	MaxViews      int       `json:"max_views"`
	HasMedia      bool      `json:"has_media"`
	Language      string    `json:"language"`
	Location      string    `json:"location"`
	Hashtags      []string  `json:"hashtags"`
	IsViral       bool      `json:"is_viral"`
	IsFeatured    bool      `json:"is_featured"`
	HasReports    bool      `json:"has_reports"`
	SortBy        string    `json:"sort_by"`    // created_at, likes_count, comments_count, views_count
	SortOrder     string    `json:"sort_order"` // asc, desc
}


// ActivityFilter represents filters for activity queries
type ActivityFilter struct {
	ActivityType  string    `json:"activity_type"` // login, post_created, comment_created, etc.
	UserID        string    `json:"user_id"`
	TargetType    string    `json:"target_type"` // user, post, comment, etc.
	TargetID      string    `json:"target_id"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	CreatedAfter  time.Time `json:"created_after"`
	CreatedBefore time.Time `json:"created_before"`
	SortBy        string    `json:"sort_by"`    // created_at, activity_type
	SortOrder     string    `json:"sort_order"` // asc, desc
}

// ==================== RESPONSE MODELS ====================

// AdminDashboardStats represents the main dashboard statistics
type AdminDashboardStats struct {
	Overview    DashboardOverview `json:"overview"`
	Users       UserStats         `json:"users"`
	Content     ContentStats      `json:"content"`
	Engagement  EngagementStats   `json:"engagement"`
	Moderation  ModerationStats   `json:"moderation"`
	System      SystemStats       `json:"system"`
	Growth      GrowthStats       `json:"growth"`
	Performance PerformanceStats  `json:"performance"`
	GeneratedAt time.Time         `json:"generated_at"`
	TimeRange   string            `json:"time_range"`
}

// DashboardOverview provides high-level platform statistics
type DashboardOverview struct {
	TotalUsers         int64  `json:"total_users"`
	ActiveUsers        int64  `json:"active_users"`
	TotalPosts         int64  `json:"total_posts"`
	TotalComments      int64  `json:"total_comments"`
	TotalStories       int64  `json:"total_stories"`
	TotalGroups        int64  `json:"total_groups"`
	TotalConversations int64  `json:"total_conversations"`
	TotalReports       int64  `json:"total_reports"`
	StorageUsed        string `json:"storage_used"`
	BandwidthUsed      string `json:"bandwidth_used"`
	PlatformHealth     string `json:"platform_health"`
}

// UserStats provides user-related statistics
type UserStats struct {
	NewRegistrations int64             `json:"new_registrations"`
	ActiveUsers      int64             `json:"active_users"`
	VerifiedUsers    int64             `json:"verified_users"`
	SuspendedUsers   int64             `json:"suspended_users"`
	DeletedUsers     int64             `json:"deleted_users"`
	TopActiveUsers   []TopUser         `json:"top_active_users"`
	UserGrowthRate   float64           `json:"user_growth_rate"`
	RetentionRate    float64           `json:"retention_rate"`
	Demographics     UserDemographics  `json:"demographics"`
	GrowthData       []GrowthDataPoint `json:"growth_data"`
}

// ContentStats provides content-related statistics
type ContentStats struct {
	NewPosts        int64             `json:"new_posts"`
	NewComments     int64             `json:"new_comments"`
	NewStories      int64             `json:"new_stories"`
	DeletedContent  int64             `json:"deleted_content"`
	ReportedContent int64             `json:"reported_content"`
	TopContent      []TopContent      `json:"top_content"`
	ContentByType   map[string]int64  `json:"content_by_type"`
	ViralContent    []ViralContent    `json:"viral_content"`
	GrowthData      []GrowthDataPoint `json:"growth_data"`
}


// ModerationStats provides moderation-related statistics
type ModerationStats struct {
	NewReports        int64               `json:"new_reports"`
	ResolvedReports   int64               `json:"resolved_reports"`
	PendingReports    int64               `json:"pending_reports"`
	ContentRemoved    int64               `json:"content_removed"`
	UsersSuspended    int64               `json:"users_suspended"`
	FalseReports      int64               `json:"false_reports"`
	AvgResolutionTime string              `json:"avg_resolution_time"`
	ModeratorActivity []ModeratorActivity `json:"moderator_activity"`
	ReportsByReason   map[string]int64    `json:"reports_by_reason"`
	QueueHealth       QueueHealth         `json:"queue_health"`
}

// SystemStats provides system-related statistics
type SystemStats struct {
	ServerUptime      string  `json:"server_uptime"`
	MemoryUsage       string  `json:"memory_usage"`
	CPUUsage          string  `json:"cpu_usage"`
	DiskUsage         string  `json:"disk_usage"`
	DatabaseSize      string  `json:"database_size"`
	CacheHitRate      string  `json:"cache_hit_rate"`
	APIResponseTime   string  `json:"api_response_time"`
	ErrorRate         string  `json:"error_rate"`
	ActiveSessions    int64   `json:"active_sessions"`
	RequestsPerSecond float64 `json:"requests_per_second"`
}

// GrowthStats provides growth-related statistics
type GrowthStats struct {
	UserGrowth       []GrowthDataPoint       `json:"user_growth"`
	ContentGrowth    []GrowthDataPoint       `json:"content_growth"`
	EngagementGrowth []GrowthDataPoint       `json:"engagement_growth"`
	RevenueGrowth    []GrowthDataPoint       `json:"revenue_growth"`
	GeographicGrowth []GeographicGrowthPoint `json:"geographic_growth"`
	RetentionCohorts []RetentionCohort       `json:"retention_cohorts"`
}

// PerformanceStats provides performance-related statistics
type PerformanceStats struct {
	AvgResponseTime     string              `json:"avg_response_time"`
	RequestsPerSecond   float64             `json:"requests_per_second"`
	ErrorRate           float64             `json:"error_rate"`
	CachePerformance    CachePerformance    `json:"cache_performance"`
	DatabasePerformance DatabasePerformance `json:"database_performance"`
	CDNPerformance      CDNPerformance      `json:"cdn_performance"`
	APIEndpointStats    []APIEndpointStat   `json:"api_endpoint_stats"`
}

// ==================== SUPPORTING MODELS ====================

// TopUser represents a top active user
type TopUser struct {
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	DisplayName     string    `json:"display_name"`
	PostsCount      int64     `json:"posts_count"`
	LikesCount      int64     `json:"likes_count"`
	EngagementScore float64   `json:"engagement_score"`
	LastActive      time.Time `json:"last_active"`
}

// TopContent represents top performing content
type TopContent struct {
	ContentID     string    `json:"content_id"`
	ContentType   string    `json:"content_type"`
	Title         string    `json:"title"`
	AuthorID      string    `json:"author_id"`
	AuthorName    string    `json:"author_name"`
	LikesCount    int64     `json:"likes_count"`
	CommentsCount int64     `json:"comments_count"`
	ViewsCount    int64     `json:"views_count"`
	SharesCount   int64     `json:"shares_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// ViralContent represents viral content
type ViralContent struct {
	ContentID      string    `json:"content_id"`
	ContentType    string    `json:"content_type"`
	Title          string    `json:"title"`
	AuthorID       string    `json:"author_id"`
	AuthorName     string    `json:"author_name"`
	ViralScore     float64   `json:"viral_score"`
	GrowthRate     float64   `json:"growth_rate"`
	EngagementRate float64   `json:"engagement_rate"`
	ReachCount     int64     `json:"reach_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// TopHashtag represents trending hashtags
type TopHashtag struct {
	Hashtag     string  `json:"hashtag"`
	UsageCount  int64   `json:"usage_count"`
	GrowthRate  float64 `json:"growth_rate"`
	PostsCount  int64   `json:"posts_count"`
	UniqueUsers int64   `json:"unique_users"`
}

// UserDemographics represents user demographic information
type UserDemographics struct {
	AgeGroups    map[string]int64 `json:"age_groups"`
	GenderSplit  map[string]int64 `json:"gender_split"`
	TopCountries map[string]int64 `json:"top_countries"`
	TopCities    map[string]int64 `json:"top_cities"`
	Languages    map[string]int64 `json:"languages"`
}

// GrowthDataPoint represents a point in growth data
type GrowthDataPoint struct {
	Date          time.Time `json:"date"`
	Value         int64     `json:"value"`
	Change        int64     `json:"change"`
	ChangePercent float64   `json:"change_percent"`
}

// GeographicGrowthPoint represents geographic growth data
type GeographicGrowthPoint struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	UserCount   int64   `json:"user_count"`
	GrowthRate  float64 `json:"growth_rate"`
}

// RetentionCohort represents user retention cohort data
type RetentionCohort struct {
	CohortMonth   string             `json:"cohort_month"`
	UserCount     int64              `json:"user_count"`
	RetentionData map[string]float64 `json:"retention_data"` // month -> retention percentage
}

// HourlyActivityPoint represents hourly activity data
type HourlyActivityPoint struct {
	Hour      int   `json:"hour"`
	UserCount int64 `json:"user_count"`
	PostCount int64 `json:"post_count"`
	Activity  int64 `json:"activity"`
}

// ModeratorActivity represents moderator activity data
type ModeratorActivity struct {
	ModeratorID      string    `json:"moderator_id"`
	ModeratorName    string    `json:"moderator_name"`
	ReportsResolved  int64     `json:"reports_resolved"`
	ContentModerated int64     `json:"content_moderated"`
	UsersModerated   int64     `json:"users_moderated"`
	AvgResponseTime  string    `json:"avg_response_time"`
	LastActive       time.Time `json:"last_active"`
}

// QueueHealth represents moderation queue health
type QueueHealth struct {
	PendingCount   int64   `json:"pending_count"`
	UrgentCount    int64   `json:"urgent_count"`
	OverdueCount   int64   `json:"overdue_count"`
	AvgWaitTime    string  `json:"avg_wait_time"`
	ProcessingRate float64 `json:"processing_rate"`
	BacklogTrend   string  `json:"backlog_trend"` // increasing, decreasing, stable
}

// CachePerformance represents cache performance metrics
type CachePerformance struct {
	HitRate         float64 `json:"hit_rate"`
	MissRate        float64 `json:"miss_rate"`
	EvictionRate    float64 `json:"eviction_rate"`
	MemoryUsage     string  `json:"memory_usage"`
	ConnectionsUsed int     `json:"connections_used"`
	AvgResponseTime string  `json:"avg_response_time"`
}

// DatabasePerformance represents database performance metrics
type DatabasePerformance struct {
	ConnectionsActive int     `json:"connections_active"`
	ConnectionsUsed   int     `json:"connections_used"`
	QueriesPerSecond  float64 `json:"queries_per_second"`
	AvgQueryTime      string  `json:"avg_query_time"`
	SlowQueriesCount  int64   `json:"slow_queries_count"`
	IndexEfficiency   float64 `json:"index_efficiency"`
	LockWaitTime      string  `json:"lock_wait_time"`
}

// CDNPerformance represents CDN performance metrics
type CDNPerformance struct {
	HitRatio        float64 `json:"hit_ratio"`
	BandwidthUsed   string  `json:"bandwidth_used"`
	RequestsServed  int64   `json:"requests_served"`
	AvgResponseTime string  `json:"avg_response_time"`
	ErrorRate       float64 `json:"error_rate"`
	CacheEfficiency float64 `json:"cache_efficiency"`
}

// APIEndpointStat represents API endpoint statistics
type APIEndpointStat struct {
	Endpoint        string  `json:"endpoint"`
	Method          string  `json:"method"`
	RequestCount    int64   `json:"request_count"`
	AvgResponseTime string  `json:"avg_response_time"`
	ErrorRate       float64 `json:"error_rate"`
	P95ResponseTime string  `json:"p95_response_time"`
	P99ResponseTime string  `json:"p99_response_time"`
}

// ==================== ACTIVITY & AUDIT MODELS ====================

// AdminActivity represents admin activity logging
type AdminActivity struct {
	BaseModel `bson:",inline"`

	AdminID      primitive.ObjectID     `json:"admin_id" bson:"admin_id"`
	Action       string                 `json:"action" bson:"action"`
	TargetType   string                 `json:"target_type" bson:"target_type"`
	TargetID     primitive.ObjectID     `json:"target_id" bson:"target_id"`
	Description  string                 `json:"description" bson:"description"`
	Metadata     map[string]interface{} `json:"metadata" bson:"metadata"`
	IPAddress    string                 `json:"ip_address" bson:"ip_address"`
	UserAgent    string                 `json:"user_agent" bson:"user_agent"`
	Success      bool                   `json:"success" bson:"success"`
	ErrorMessage string                 `json:"error_message,omitempty" bson:"error_message,omitempty"`
}

// SystemLog represents system logging
type SystemLog struct {
	ID        primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time              `json:"timestamp" bson:"timestamp"`
	Level     string                 `json:"level" bson:"level"` // debug, info, warn, error, fatal
	Service   string                 `json:"service" bson:"service"`
	Message   string                 `json:"message" bson:"message"`
	Data      map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty" bson:"trace_id,omitempty"`
	UserID    primitive.ObjectID     `json:"user_id,omitempty" bson:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty" bson:"request_id,omitempty"`
}

// AuditLog represents audit trail logging
type AuditLog struct {
	BaseModel `bson:",inline"`

	Action     string                 `json:"action" bson:"action"`
	ActorID    primitive.ObjectID     `json:"actor_id" bson:"actor_id"`
	ActorType  string                 `json:"actor_type" bson:"actor_type"` // user, admin, system
	TargetType string                 `json:"target_type" bson:"target_type"`
	TargetID   primitive.ObjectID     `json:"target_id" bson:"target_id"`
	Changes    map[string]interface{} `json:"changes,omitempty" bson:"changes,omitempty"`
	OldValues  map[string]interface{} `json:"old_values,omitempty" bson:"old_values,omitempty"`
	NewValues  map[string]interface{} `json:"new_values,omitempty" bson:"new_values,omitempty"`
	Reason     string                 `json:"reason,omitempty" bson:"reason,omitempty"`
	IPAddress  string                 `json:"ip_address" bson:"ip_address"`
	UserAgent  string                 `json:"user_agent" bson:"user_agent"`
	SessionID  string                 `json:"session_id,omitempty" bson:"session_id,omitempty"`
}

// ==================== EXPORT MODELS ====================

// ExportJob represents a data export job
type ExportJob struct {
	BaseModel `bson:",inline"`

	ExportID     string                 `json:"export_id" bson:"export_id"`
	RequestedBy  primitive.ObjectID     `json:"requested_by" bson:"requested_by"`
	DataType     string                 `json:"data_type" bson:"data_type"`
	Format       string                 `json:"format" bson:"format"`
	Parameters   map[string]interface{} `json:"parameters" bson:"parameters"`
	Status       string                 `json:"status" bson:"status"` // queued, processing, completed, failed
	Progress     float64                `json:"progress" bson:"progress"`
	FileURL      string                 `json:"file_url,omitempty" bson:"file_url,omitempty"`
	FileSize     int64                  `json:"file_size,omitempty" bson:"file_size,omitempty"`
	RecordCount  int64                  `json:"record_count" bson:"record_count"`
	ErrorMessage string                 `json:"error_message,omitempty" bson:"error_message,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty" bson:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	ExpiresAt    time.Time              `json:"expires_at" bson:"expires_at"`
}

// ==================== ALERT MODELS ====================

// SystemAlert represents system alerts
type SystemAlert struct {
	BaseModel `bson:",inline"`

	AlertID        string                 `json:"alert_id" bson:"alert_id"`
	Type           string                 `json:"type" bson:"type"`         // error, warning, info
	Category       string                 `json:"category" bson:"category"` // performance, security, moderation, etc.
	Title          string                 `json:"title" bson:"title"`
	Description    string                 `json:"description" bson:"description"`
	Severity       int                    `json:"severity" bson:"severity"` // 1-5, 5 being most severe
	Source         string                 `json:"source" bson:"source"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Acknowledged   bool                   `json:"acknowledged" bson:"acknowledged"`
	AcknowledgedBy primitive.ObjectID     `json:"acknowledged_by,omitempty" bson:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty" bson:"acknowledged_at,omitempty"`
	Resolved       bool                   `json:"resolved" bson:"resolved"`
	ResolvedBy     primitive.ObjectID     `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

// ==================== REQUEST MODELS ====================

// BulkUserActionRequest represents bulk user action request
type BulkUserActionRequest struct {
	UserIDs   []string `json:"user_ids" binding:"required"`
	Action    string   `json:"action" binding:"required"`
	Reason    string   `json:"reason,omitempty"`
	Duration  string   `json:"duration,omitempty"`
	Note      string   `json:"note,omitempty"`
	SendEmail bool     `json:"send_email"`
}

// BulkContentActionRequest represents bulk content action request
type BulkContentActionRequest struct {
	ContentIDs   []string `json:"content_ids" binding:"required"`
	ContentType  string   `json:"content_type" binding:"required"`
	Action       string   `json:"action" binding:"required"`
	Reason       string   `json:"reason,omitempty"`
	NotifyAuthor bool     `json:"notify_author"`
}

// ExportDataRequest represents export data request
type ExportDataRequest struct {
	DataType   string                 `json:"data_type" binding:"required"`
	Format     string                 `json:"format" binding:"required"`
	TimeRange  string                 `json:"time_range,omitempty"`
	StartDate  time.Time              `json:"start_date,omitempty"`
	EndDate    time.Time              `json:"end_date,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	IncludeRaw bool                   `json:"include_raw"`
	Email      string                 `json:"email,omitempty"`
}

// SystemConfigUpdateRequest represents system configuration update
type SystemConfigUpdateRequest struct {
	General       map[string]interface{} `json:"general,omitempty"`
	Content       map[string]interface{} `json:"content,omitempty"`
	User          map[string]interface{} `json:"user,omitempty"`
	Notifications map[string]interface{} `json:"notifications,omitempty"`
	Privacy       map[string]interface{} `json:"privacy,omitempty"`
	Security      map[string]interface{} `json:"security,omitempty"`
}

// NotificationBroadcastRequest represents notification broadcast
type NotificationBroadcastRequest struct {
	Title       string                 `json:"title" binding:"required"`
	Message     string                 `json:"message" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	TargetUsers []string               `json:"target_users,omitempty"`
	UserFilter  map[string]interface{} `json:"user_filter,omitempty"`
	Channels    []string               `json:"channels"` // email, push, sms
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// ==================== RESPONSE MODELS ====================

// AdminUserResponse represents user data for admin panel
type AdminUserResponse struct {
	ID               string     `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	DisplayName      string     `json:"display_name"`
	Status           string     `json:"status"`
	Role             string     `json:"role"`
	Verified         bool       `json:"verified"`
	EmailVerified    bool       `json:"email_verified"`
	CreatedAt        time.Time  `json:"created_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	PostsCount       int64      `json:"posts_count"`
	FollowersCount   int64      `json:"followers_count"`
	FollowingCount   int64      `json:"following_count"`
	ReportsCount     int64      `json:"reports_count"`
	WarningsCount    int64      `json:"warnings_count"`
	SuspensionsCount int64      `json:"suspensions_count"`
	LastActivity     *time.Time `json:"last_activity"`
	RegistrationIP   string     `json:"registration_ip"`
	Location         string     `json:"location"`
	Language         string     `json:"language"`
}

// AdminContentResponse represents content data for admin panel
type AdminContentResponse struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	AuthorID       string    `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	Title          string    `json:"title,omitempty"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LikesCount     int64     `json:"likes_count"`
	CommentsCount  int64     `json:"comments_count"`
	ViewsCount     int64     `json:"views_count"`
	SharesCount    int64     `json:"shares_count"`
	ReportsCount   int64     `json:"reports_count"`
	IsFeatured     bool      `json:"is_featured"`
	IsHidden       bool      `json:"is_hidden"`
	Language       string    `json:"language"`
	Location       string    `json:"location"`
	Hashtags       []string  `json:"hashtags"`
	HasMedia       bool      `json:"has_media"`
}

// AdminReportResponse represents report data for admin panel
type AdminReportResponse struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	Priority         string     `json:"priority"`
	TargetType       string     `json:"target_type"`
	TargetID         string     `json:"target_id"`
	Reason           string     `json:"reason"`
	Description      string     `json:"description"`
	ReporterID       string     `json:"reporter_id"`
	ReporterUsername string     `json:"reporter_username"`
	AssignedTo       string     `json:"assigned_to,omitempty"`
	AssignedToName   string     `json:"assigned_to_name,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	Resolution       string     `json:"resolution,omitempty"`
	ResolutionNote   string     `json:"resolution_note,omitempty"`
	AutoDetected     bool       `json:"auto_detected"`
	RequiresFollowUp bool       `json:"requires_follow_up"`
	Category         string     `json:"category"`
	Severity         int        `json:"severity"`
}

// AdminActivityResponse represents activity data for admin panel
type AdminActivityResponse struct {
	ID           string                 `json:"id"`
	ActivityType string                 `json:"activity_type"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	TargetType   string                 `json:"target_type"`
	TargetID     string                 `json:"target_id"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	CreatedAt    time.Time              `json:"created_at"`
	Success      bool                   `json:"success"`
}

// BulkOperationResult represents the result of bulk operations
type BulkOperationResult struct {
	Total      int                       `json:"total"`
	Successful int                       `json:"successful"`
	Failed     int                       `json:"failed"`
	Results    []BulkOperationItemResult `json:"results"`
}

// BulkOperationItemResult represents individual item result in bulk operations
type BulkOperationItemResult struct {
	ID      string `json:"id"`
	Status  string `json:"status"` // success, failed
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
