// common.go - Auto-generated placeholder
// models/common.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// BeforeCreate sets timestamps before creating a document
func (b *BaseModel) BeforeCreate() {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
}

// BeforeUpdate sets updated timestamp
func (b *BaseModel) BeforeUpdate() {
	b.UpdatedAt = time.Now()
}

// SoftDelete marks the document as deleted
func (b *BaseModel) SoftDelete() {
	now := time.Now()
	b.DeletedAt = &now
	b.UpdatedAt = now
}

// IsDeleted checks if the document is soft deleted
func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt != nil
}

// Privacy settings enum
type PrivacyLevel string

const (
	PrivacyPublic  PrivacyLevel = "public"
	PrivacyFriends PrivacyLevel = "friends"
	PrivacyPrivate PrivacyLevel = "private"
)

// Content type enum
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeVideo ContentType = "video"
	ContentTypeAudio ContentType = "audio"
	ContentTypeFile  ContentType = "file"
	ContentTypeLink  ContentType = "link"
	ContentTypeGif   ContentType = "gif"
	ContentTypePoll  ContentType = "poll"
)

// Reaction type enum
type ReactionType string

const (
	ReactionLike    ReactionType = "like"
	ReactionLove    ReactionType = "love"
	ReactionHaha    ReactionType = "haha"
	ReactionWow     ReactionType = "wow"
	ReactionSad     ReactionType = "sad"
	ReactionAngry   ReactionType = "angry"
	ReactionSupport ReactionType = "support"
)

// Notification type enum
type NotificationType string

const (
	NotificationLike          NotificationType = "like"
	NotificationLove          NotificationType = "love"
	NotificationComment       NotificationType = "comment"
	NotificationFollow        NotificationType = "follow"
	NotificationMessage       NotificationType = "message"
	NotificationMention       NotificationType = "mention"
	NotificationGroupInvite   NotificationType = "group_invite"
	NotificationEventInvite   NotificationType = "event_invite"
	NotificationFriendRequest NotificationType = "friend_request"
	NotificationPostShare     NotificationType = "post_share"
	NotificationStoryView     NotificationType = "story_view"
	NotificationGroupPost     NotificationType = "group_post"
	NotificationEventReminder NotificationType = "event_reminder"
)

// User role enum
type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleModerator  UserRole = "moderator"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

// Report status enum
type ReportStatus string

const (
	ReportPending   ReportStatus = "pending"
	ReportReviewing ReportStatus = "reviewing"
	ReportResolved  ReportStatus = "resolved"
	ReportRejected  ReportStatus = "rejected"
)

// Report reason enum
type ReportReason string

const (
	ReportSpam       ReportReason = "spam"
	ReportHarassment ReportReason = "harassment"
	ReportHateSpeech ReportReason = "hate_speech"
	ReportViolence   ReportReason = "violence"
	ReportNudity     ReportReason = "nudity"
	ReportFakeNews   ReportReason = "fake_news"
	ReportCopyright  ReportReason = "copyright"
	ReportOther      ReportReason = "other"
)

// Message status enum
type MessageStatus string

const (
	MessageSent      MessageStatus = "sent"
	MessageDelivered MessageStatus = "delivered"
	MessageRead      MessageStatus = "read"
)

// Event status enum
type EventStatus string

const (
	EventDraft     EventStatus = "draft"
	EventPublished EventStatus = "published"
	EventCancelled EventStatus = "cancelled"
	EventCompleted EventStatus = "completed"
)

// RSVP status enum
type RSVPStatus string

const (
	RSVPGoing    RSVPStatus = "going"
	RSVPMaybe    RSVPStatus = "maybe"
	RSVPNotGoing RSVPStatus = "not_going"
)

// Group role enum
type GroupRole string

const (
	GroupRoleMember    GroupRole = "member"
	GroupRoleModerator GroupRole = "moderator"
	GroupRoleAdmin     GroupRole = "admin"
	GroupRoleOwner     GroupRole = "owner"
)

// Group privacy enum
type GroupPrivacy string

const (
	GroupPublic  GroupPrivacy = "public"
	GroupPrivate GroupPrivacy = "private"
	GroupSecret  GroupPrivacy = "secret"
)

// Location struct for geo-tagging
type Location struct {
	Name      string  `json:"name" bson:"name"`
	Address   string  `json:"address" bson:"address"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	PlaceID   string  `json:"place_id,omitempty" bson:"place_id,omitempty"`
}

// MediaInfo struct for media metadata
type MediaInfo struct {
	URL       string `json:"url" bson:"url"`
	Type      string `json:"type" bson:"type"` // image, video, audio
	Size      int64  `json:"size" bson:"size"`
	Width     int    `json:"width,omitempty" bson:"width,omitempty"`
	Height    int    `json:"height,omitempty" bson:"height,omitempty"`
	Duration  int    `json:"duration,omitempty" bson:"duration,omitempty"` // for videos/audio in seconds
	Thumbnail string `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
	AltText   string `json:"alt_text,omitempty" bson:"alt_text,omitempty"`
}

// PaginationInfo for API responses
type PaginationInfo struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// PaginatedResponse generic structure for paginated responses
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// GenericCountResponse for count endpoints
type GenericCountResponse struct {
	Count int64 `json:"count"`
}

// Stats struct for various statistics
type Stats struct {
	PostsCount     int64 `json:"posts_count" bson:"posts_count"`
	FollowersCount int64 `json:"followers_count" bson:"followers_count"`
	FollowingCount int64 `json:"following_count" bson:"following_count"`
	LikesCount     int64 `json:"likes_count" bson:"likes_count"`
	CommentsCount  int64 `json:"comments_count" bson:"comments_count"`
	SharesCount    int64 `json:"shares_count" bson:"shares_count"`
	ViewsCount     int64 `json:"views_count" bson:"views_count"`
}

// PrivacySettings struct for user privacy configuration
type PrivacySettings struct {
	ProfileVisibility   PrivacyLevel `json:"profile_visibility" bson:"profile_visibility"`
	PostsVisibility     PrivacyLevel `json:"posts_visibility" bson:"posts_visibility"`
	FollowersVisibility PrivacyLevel `json:"followers_visibility" bson:"followers_visibility"`
	FollowingVisibility PrivacyLevel `json:"following_visibility" bson:"following_visibility"`
	EmailVisibility     PrivacyLevel `json:"email_visibility" bson:"email_visibility"`
	PhoneVisibility     PrivacyLevel `json:"phone_visibility" bson:"phone_visibility"`
	AllowMessages       bool         `json:"allow_messages" bson:"allow_messages"`
	AllowTagging        bool         `json:"allow_tagging" bson:"allow_tagging"`
	AllowFollowRequests bool         `json:"allow_follow_requests" bson:"allow_follow_requests"`
	ShowOnlineStatus    bool         `json:"show_online_status" bson:"show_online_status"`
	AllowStoryViews     bool         `json:"allow_story_views" bson:"allow_story_views"`
}

// NotificationSettings struct for user notification preferences
type NotificationSettings struct {
	EmailNotifications bool `json:"email_notifications" bson:"email_notifications"`
	PushNotifications  bool `json:"push_notifications" bson:"push_notifications"`
	SMSNotifications   bool `json:"sms_notifications" bson:"sms_notifications"`

	// Specific notification types
	LikeNotifications    bool `json:"like_notifications" bson:"like_notifications"`
	CommentNotifications bool `json:"comment_notifications" bson:"comment_notifications"`
	FollowNotifications  bool `json:"follow_notifications" bson:"follow_notifications"`
	MessageNotifications bool `json:"message_notifications" bson:"message_notifications"`
	MentionNotifications bool `json:"mention_notifications" bson:"mention_notifications"`
	GroupNotifications   bool `json:"group_notifications" bson:"group_notifications"`
	EventNotifications   bool `json:"event_notifications" bson:"event_notifications"`
}

// DefaultPrivacySettings returns default privacy settings for new users
func DefaultPrivacySettings() PrivacySettings {
	return PrivacySettings{
		ProfileVisibility:   PrivacyPublic,
		PostsVisibility:     PrivacyPublic,
		FollowersVisibility: PrivacyPublic,
		FollowingVisibility: PrivacyPublic,
		EmailVisibility:     PrivacyPrivate,
		PhoneVisibility:     PrivacyPrivate,
		AllowMessages:       true,
		AllowTagging:        true,
		AllowFollowRequests: true,
		ShowOnlineStatus:    true,
		AllowStoryViews:     true,
	}
}

// DefaultNotificationSettings returns default notification settings for new users
func DefaultNotificationSettings() NotificationSettings {
	return NotificationSettings{
		EmailNotifications:   true,
		PushNotifications:    true,
		SMSNotifications:     false,
		LikeNotifications:    true,
		CommentNotifications: true,
		FollowNotifications:  true,
		MessageNotifications: true,
		MentionNotifications: true,
		GroupNotifications:   true,
		EventNotifications:   true,
	}
}
