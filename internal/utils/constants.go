// utils/constants.go
package utils

import "time"

// Application constants
const (
	// Application Info
	AppName    = "Social Media API"
	AppVersion = "1.0.0"
	APIVersion = "v1"

	// Default pagination
	DefaultPageSize               = 20
	MaxPageSize                   = 100
	MinPageSize                   = 1
	MaxBulkNotificationRecipients = 10
	MaxMessageContentLength       = 5000
	MaxPostContentLength          = 5000
	MaxCommentContentLength       = 2000
	MaxStoryContentLength         = 2000
	MaxBioLength                  = 500
	MaxUsernameLength             = 50
	MinUsernameLength             = 3
	MaxDisplayNameLength          = 100
	MaxHashtagLength              = 100
	MaxGroupNameLength            = 100
	MaxEventTitleLength           = 200

	// Media constraints
	MaxImageSizeMB            = 10  // 10MB
	MaxVideoSizeMB            = 100 // 100MB
	MaxAudioSizeMB            = 50  // 50MB
	MaxDocumentSizeMB         = 25  // 25MB
	MaxImageWidth             = 4096
	MaxImageHeight            = 4096
	MaxBulkUploadFiles        = 20
	MaxAltTextLength          = 250
	MaxMediaDescriptionLength = 1000
	MaxMediaCaptionLength     = 500
	MaxThumbnailSize          = 500 // pixels
	DefaultThumbnailQuality   = 85  // JPEG quality
	MaxMediaVariants          = 10
	// Rate limiting
	DefaultRateLimit       = 100 // requests per minute
	AuthRateLimit          = 5   // login attempts per minute
	PasswordResetRateLimit = 3   // password reset attempts per hour
	EmailVerifyRateLimit   = 3   // email verification attempts per hour
	PostCreationRateLimit  = 10  // posts per hour
	CommentRateLimit       = 30  // comments per minute
	MessageRateLimit       = 60  // messages per minute

	// Token expiration
	AccessTokenExpiry       = 24 * time.Hour      // 24 hours
	RefreshTokenExpiry      = 30 * 24 * time.Hour // 30 days
	PasswordResetExpiry     = 1 * time.Hour       // 1 hour
	EmailVerificationExpiry = 24 * time.Hour      // 24 hours
	SessionExpiry           = 7 * 24 * time.Hour  // 7 days

	// Story and temporary content
	StoryExpiry        = 24 * time.Hour      // 24 hours
	MessageExpiry      = 7 * 24 * time.Hour  // 7 days (for disappearing messages)
	NotificationExpiry = 30 * 24 * time.Hour // 30 days

	// Cache TTL
	UserCacheTTL     = 15 * time.Minute
	PostCacheTTL     = 5 * time.Minute
	FeedCacheTTL     = 2 * time.Minute
	TrendingCacheTTL = 10 * time.Minute
	SearchCacheTTL   = 5 * time.Minute

	// Database
	MongoTimeout     = 10 * time.Second
	MongoMaxPoolSize = 100
	MongoMinPoolSize = 5

	// File upload paths
	UserUploadsPath    = "uploads/users"
	PostUploadsPath    = "uploads/posts"
	StoryUploadsPath   = "uploads/stories"
	GroupUploadsPath   = "uploads/groups"
	EventUploadsPath   = "uploads/events"
	MessageUploadsPath = "uploads/messages"
	TempUploadsPath    = "uploads/temp"

	// Image sizes for thumbnails
	ThumbnailSmallSize  = 150
	ThumbnailMediumSize = 300
	ThumbnailLargeSize  = 600

	// Search limits
	MaxSearchResults = 50
	MinSearchLength  = 2
	MaxSearchLength  = 100

	// Notification batch sizes
	NotificationBatchSize = 100
	EmailBatchSize        = 50
	PushBatchSize         = 1000

	// Analytics retention
	AnalyticsRetentionDays = 90
	LogRetentionDays       = 30

	// Feature flags
	EnableStories            = true
	EnableGroups             = true
	EnableEvents             = true
	EnableLiveChat           = true
	EnablePushNotifications  = true
	EnableEmailNotifications = true
	EnableSMSNotifications   = false
	EnableContentModeration  = true
	EnableAnalytics          = true
)

// HTTP Status Messages
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFail    = "fail"
)

// Error Messages
const (
	ErrInvalidCredentials    = "Invalid email or password"
	ErrUserNotFound          = "User not found"
	ErrUserAlreadyExists     = "User already exists"
	ErrEmailAlreadyExists    = "Email already registered"
	ErrUsernameAlreadyExists = "Username already taken"
	ErrUnauthorized          = "Unauthorized access"
	ErrForbidden             = "Access forbidden"
	ErrTokenExpired          = "Token has expired"
	ErrInvalidToken          = "Invalid token"
	ErrAccountSuspended      = "Account has been suspended"
	ErrAccountNotVerified    = "Email verification required"
	ErrInvalidRequest        = "Invalid request format"
	ErrValidationFailed      = "Validation failed"
	ErrInternalError         = "Internal server error"
	ErrNotFound              = "Resource not found"
	ErrContentTooLarge       = "Content exceeds maximum length"
	ErrFileTooLarge          = "File size exceeds limit"
	ErrUnsupportedFileType   = "Unsupported file type"
	ErrRateLimitExceeded     = "Rate limit exceeded"
	ErrServiceUnavailable    = "Service temporarily unavailable"
)

// Success Messages
const (
	MsgUserCreated       = "User registered successfully"
	MsgLoginSuccess      = "Login successful"
	MsgLogoutSuccess     = "Logout successful"
	MsgPasswordChanged   = "Password changed successfully"
	MsgPasswordResetSent = "Password reset link sent to email"
	MsgEmailVerified     = "Email verified successfully"
	MsgProfileUpdated    = "Profile updated successfully"
	MsgPostCreated       = "Post created successfully"
	MsgPostUpdated       = "Post updated successfully"
	MsgPostDeleted       = "Post deleted successfully"
	MsgCommentCreated    = "Comment added successfully"
	MsgCommentUpdated    = "Comment updated successfully"
	MsgCommentDeleted    = "Comment deleted successfully"
	MsgLikeAdded         = "Reaction added successfully"
	MsgLikeRemoved       = "Reaction removed successfully"
	MsgFollowSuccess     = "User followed successfully"
	MsgUnfollowSuccess   = "User unfollowed successfully"
	MsgMessageSent       = "Message sent successfully"
	MsgStoryCreated      = "Story created successfully"
	MsgGroupCreated      = "Group created successfully"
	MsgGroupJoined       = "Joined group successfully"
	MsgEventCreated      = "Event created successfully"
	MsgEventJoined       = "RSVP updated successfully"
)

// Context Keys
const (
	ContextUserID    = "user_id"
	ContextUser      = "user"
	ContextUserRole  = "user_role"
	ContextSessionID = "session_id"
)

// Default Values
var (
	DefaultUserRoles = []string{"user", "moderator", "admin", "super_admin"}

	SupportedImageTypes = []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/bmp", "image/tiff",
	}

	SupportedVideoTypes = []string{
		"video/mp4", "video/mov", "video/avi", "video/mkv",
		"video/webm", "video/flv", "video/wmv",
	}

	SupportedAudioTypes = []string{
		"audio/mp3", "audio/wav", "audio/ogg", "audio/aac",
		"audio/flac", "audio/m4a",
	}

	SupportedDocumentTypes = []string{
		"application/pdf", "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain", "application/rtf", "application/vnd.oasis.opendocument.text",
	}

	DefaultPrivacyLevels = []string{"public", "friends", "private"}

	DefaultReactionTypes = []string{
		"like", "love", "haha", "wow", "sad", "angry", "support",
	}

	DefaultNotificationTypes = []string{
		"like", "comment", "follow", "message", "mention",
		"group_invite", "event_invite", "friend_request",
		"post_share", "story_view", "group_post", "event_reminder",
	}

	DefaultHashtagCategories = []string{
		"general", "entertainment", "sports", "news", "technology",
		"business", "lifestyle", "travel", "food", "fashion", "art",
		"music", "health", "education", "politics", "science", "nature",
		"photography", "fitness", "gaming",
	}

	DefaultGroupCategories = []string{
		"general", "technology", "business", "health", "education",
		"entertainment", "sports", "travel", "food", "art", "music",
		"books", "gaming", "fitness", "parenting", "pets", "hobbies",
		"local", "support", "professional",
	}

	DefaultEventCategories = []string{
		"business", "technology", "education", "health", "arts", "music",
		"sports", "entertainment", "food", "travel", "lifestyle",
		"community", "charity", "networking", "workshop", "conference",
		"meetup", "party", "festival", "other",
	}
)

// Environment variables keys
const (
	EnvMongoURI     = "MONGO_URI"
	EnvDBName       = "DB_NAME"
	EnvPort         = "PORT"
	EnvJWTSecret    = "JWT_SECRET"
	EnvJWTExpiry    = "JWT_EXPIRY"
	EnvGinMode      = "GIN_MODE"
	EnvRedisURL     = "REDIS_URL"
	EnvAWSAccessKey = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretKey = "AWS_SECRET_ACCESS_KEY"
	EnvAWSRegion    = "AWS_REGION"
	EnvS3Bucket     = "S3_BUCKET"
	EnvSMTPHost     = "SMTP_HOST"
	EnvSMTPPort     = "SMTP_PORT"
	EnvSMTPUser     = "SMTP_USER"
	EnvSMTPPass     = "SMTP_PASS"
	EnvFrontendURL  = "FRONTEND_URL"
	EnvAPIURL       = "API_URL"
)

// Regular expressions
const (
	EmailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	UsernameRegex = `^[a-zA-Z0-9_]+$`
	PhoneRegex    = `^\+?[1-9]\d{1,14}$`
	URLRegex      = `^https?://[^\s/$.?#].[^\s]*$`
	HashtagRegex  = `#[a-zA-Z0-9_]+`
	MentionRegex  = `@[a-zA-Z0-9_]+`
)
