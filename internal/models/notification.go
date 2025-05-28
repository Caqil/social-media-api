// models/notification.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a notification sent to a user
type Notification struct {
	BaseModel `bson:",inline"`

	// Recipient
	RecipientID primitive.ObjectID `json:"recipient_id" bson:"recipient_id" validate:"required"`
	Recipient   UserResponse       `json:"recipient,omitempty" bson:"-"` // Populated when querying

	// Sender/Actor (who triggered the notification)
	ActorID primitive.ObjectID `json:"actor_id" bson:"actor_id" validate:"required"`
	Actor   UserResponse       `json:"actor,omitempty" bson:"-"` // Populated when querying

	// Notification details
	Type       NotificationType `json:"type" bson:"type" validate:"required"`
	Title      string           `json:"title" bson:"title" validate:"required,max=200"`
	Message    string           `json:"message" bson:"message" validate:"required,max=500"`
	ActionText string           `json:"action_text,omitempty" bson:"action_text,omitempty" validate:"max=50"`

	// Target object (what the notification is about)
	TargetID   *primitive.ObjectID `json:"target_id,omitempty" bson:"target_id,omitempty"`
	TargetType string              `json:"target_type,omitempty" bson:"target_type,omitempty"` // post, comment, user, group, event
	TargetURL  string              `json:"target_url,omitempty" bson:"target_url,omitempty"`

	// Additional context data
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Notification status
	IsRead      bool       `json:"is_read" bson:"is_read"`
	ReadAt      *time.Time `json:"read_at,omitempty" bson:"read_at,omitempty"`
	IsDelivered bool       `json:"is_delivered" bson:"is_delivered"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`

	// Delivery channels
	SentViaEmail bool `json:"sent_via_email" bson:"sent_via_email"`
	SentViaPush  bool `json:"sent_via_push" bson:"sent_via_push"`
	SentViaSMS   bool `json:"sent_via_sms" bson:"sent_via_sms"`

	// Grouping (for bundling similar notifications)
	GroupKey   string `json:"group_key,omitempty" bson:"group_key,omitempty"`
	GroupCount int64  `json:"group_count" bson:"group_count"`
	IsGrouped  bool   `json:"is_grouped" bson:"is_grouped"`

	// Priority and scheduling
	Priority    string     `json:"priority" bson:"priority"` // high, medium, low
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" bson:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`

	// Interaction tracking
	ClickedAt       *time.Time             `json:"clicked_at,omitempty" bson:"clicked_at,omitempty"`
	DismissedAt     *time.Time             `json:"dismissed_at,omitempty" bson:"dismissed_at,omitempty"`
	InteractionData map[string]interface{} `json:"interaction_data,omitempty" bson:"interaction_data,omitempty"`
}

// NotificationResponse represents the notification data returned in API responses
type NotificationResponse struct {
	ID          string                 `json:"id"`
	RecipientID string                 `json:"recipient_id"`
	ActorID     string                 `json:"actor_id"`
	Actor       UserResponse           `json:"actor"`
	Type        NotificationType       `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	ActionText  string                 `json:"action_text,omitempty"`
	TargetID    string                 `json:"target_id,omitempty"`
	TargetType  string                 `json:"target_type,omitempty"`
	TargetURL   string                 `json:"target_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsRead      bool                   `json:"is_read"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	GroupCount  int64                  `json:"group_count"`
	IsGrouped   bool                   `json:"is_grouped"`
	Priority    string                 `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	TimeAgo     string                 `json:"time_ago,omitempty"`

	// Additional display info
	Icon         string `json:"icon,omitempty"`
	Color        string `json:"color,omitempty"`
	PreviewImage string `json:"preview_image,omitempty"`
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	RecipientID  string                 `json:"recipient_id" validate:"required"`
	ActorID      string                 `json:"actor_id" validate:"required"`
	Type         NotificationType       `json:"type" validate:"required"`
	Title        string                 `json:"title" validate:"required,max=200"`
	Message      string                 `json:"message" validate:"required,max=500"`
	ActionText   string                 `json:"action_text,omitempty" validate:"max=50"`
	TargetID     string                 `json:"target_id,omitempty"`
	TargetType   string                 `json:"target_type,omitempty"`
	TargetURL    string                 `json:"target_url,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Priority     string                 `json:"priority,omitempty" validate:"omitempty,oneof=high medium low"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	SendViaEmail bool                   `json:"send_via_email"`
	SendViaPush  bool                   `json:"send_via_push"`
	SendViaSMS   bool                   `json:"send_via_sms"`
}

// BulkCreateNotificationRequest represents the request to create multiple notifications
type BulkCreateNotificationRequest struct {
	RecipientIDs []string               `json:"recipient_ids" validate:"required,min=1,max=1000"`
	ActorID      string                 `json:"actor_id" validate:"required"`
	Type         NotificationType       `json:"type" validate:"required"`
	Title        string                 `json:"title" validate:"required,max=200"`
	Message      string                 `json:"message" validate:"required,max=500"`
	ActionText   string                 `json:"action_text,omitempty" validate:"max=50"`
	TargetID     string                 `json:"target_id,omitempty"`
	TargetType   string                 `json:"target_type,omitempty"`
	TargetURL    string                 `json:"target_url,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Priority     string                 `json:"priority,omitempty" validate:"omitempty,oneof=high medium low"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
	SendViaEmail bool                   `json:"send_via_email"`
	SendViaPush  bool                   `json:"send_via_push"`
	SendViaSMS   bool                   `json:"send_via_sms"`
}

// NotificationStatsResponse represents notification statistics
type NotificationStatsResponse struct {
	TotalCount     int64 `json:"total_count"`
	UnreadCount    int64 `json:"unread_count"`
	ReadCount      int64 `json:"read_count"`
	DeliveredCount int64 `json:"delivered_count"`
	ClickedCount   int64 `json:"clicked_count"`
	DismissedCount int64 `json:"dismissed_count"`
	RecentCount    int64 `json:"recent_count"` // Last 24 hours
}

// NotificationPreferences represents user's notification preferences
type NotificationPreferences struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// Channel preferences
	EmailEnabled bool `json:"email_enabled" bson:"email_enabled"`
	PushEnabled  bool `json:"push_enabled" bson:"push_enabled"`
	SMSEnabled   bool `json:"sms_enabled" bson:"sms_enabled"`

	// Type-specific preferences
	LikeNotifications          bool `json:"like_notifications" bson:"like_notifications"`
	CommentNotifications       bool `json:"comment_notifications" bson:"comment_notifications"`
	FollowNotifications        bool `json:"follow_notifications" bson:"follow_notifications"`
	MessageNotifications       bool `json:"message_notifications" bson:"message_notifications"`
	MentionNotifications       bool `json:"mention_notifications" bson:"mention_notifications"`
	GroupNotifications         bool `json:"group_notifications" bson:"group_notifications"`
	EventNotifications         bool `json:"event_notifications" bson:"event_notifications"`
	PostShareNotifications     bool `json:"post_share_notifications" bson:"post_share_notifications"`
	StoryViewNotifications     bool `json:"story_view_notifications" bson:"story_view_notifications"`
	FriendRequestNotifications bool `json:"friend_request_notifications" bson:"friend_request_notifications"`

	// Timing preferences
	QuietHoursEnabled bool      `json:"quiet_hours_enabled" bson:"quiet_hours_enabled"`
	QuietHoursStart   time.Time `json:"quiet_hours_start" bson:"quiet_hours_start"`
	QuietHoursEnd     time.Time `json:"quiet_hours_end" bson:"quiet_hours_end"`

	// Grouping preferences
	GroupSimilarNotifications bool   `json:"group_similar_notifications" bson:"group_similar_notifications"`
	DigestFrequency           string `json:"digest_frequency" bson:"digest_frequency"` // immediate, hourly, daily, weekly
}

// Methods for Notification model

// BeforeCreate sets default values before creating notification
func (n *Notification) BeforeCreate() {
	n.BaseModel.BeforeCreate()

	// Set default values
	n.IsRead = false
	n.IsDelivered = false
	n.GroupCount = 1
	n.IsGrouped = false
	n.SentViaEmail = false
	n.SentViaPush = false
	n.SentViaSMS = false

	// Set default priority
	if n.Priority == "" {
		n.Priority = "medium"
	}

	// Generate group key for similar notifications
	if n.GroupKey == "" {
		n.GroupKey = n.generateGroupKey()
	}

	// Set expiration if not specified
	if n.ExpiresAt == nil {
		// Default expiration: 30 days from creation
		expiry := n.CreatedAt.Add(30 * 24 * time.Hour)
		n.ExpiresAt = &expiry
	}
}

// generateGroupKey generates a key for grouping similar notifications
func (n *Notification) generateGroupKey() string {
	// Group by type, actor, and target
	key := string(n.Type) + "_" + n.ActorID.Hex()
	if n.TargetID != nil {
		key += "_" + n.TargetID.Hex()
	}
	return key
}

// ToNotificationResponse converts Notification model to NotificationResponse
func (n *Notification) ToNotificationResponse() NotificationResponse {
	response := NotificationResponse{
		ID:          n.ID.Hex(),
		RecipientID: n.RecipientID.Hex(),
		ActorID:     n.ActorID.Hex(),
		Type:        n.Type,
		Title:       n.Title,
		Message:     n.Message,
		ActionText:  n.ActionText,
		TargetType:  n.TargetType,
		TargetURL:   n.TargetURL,
		Metadata:    n.Metadata,
		IsRead:      n.IsRead,
		ReadAt:      n.ReadAt,
		GroupCount:  n.GroupCount,
		IsGrouped:   n.IsGrouped,
		Priority:    n.Priority,
		CreatedAt:   n.CreatedAt,
	}

	if n.TargetID != nil {
		response.TargetID = n.TargetID.Hex()
	}

	// Set display properties based on notification type
	response.Icon, response.Color = n.getDisplayProperties()

	return response
}

// getDisplayProperties returns icon and color based on notification type
func (n *Notification) getDisplayProperties() (string, string) {
	switch n.Type {
	case NotificationLike:
		return "👍", "#1877F2"
	case NotificationLove:
		return "❤️", "#E41E3F"
	case NotificationComment:
		return "💬", "#42B883"
	case NotificationFollow:
		return "👤", "#8B5CF6"
	case NotificationMessage:
		return "✉️", "#06B6D4"
	case NotificationMention:
		return "📢", "#F59E0B"
	case NotificationGroupInvite:
		return "👥", "#10B981"
	case NotificationEventInvite:
		return "📅", "#EF4444"
	case NotificationFriendRequest:
		return "🤝", "#6366F1"
	case NotificationPostShare:
		return "🔄", "#84CC16"
	case NotificationStoryView:
		return "👁️", "#EC4899"
	case NotificationGroupPost:
		return "📝", "#F97316"
	case NotificationEventReminder:
		return "⏰", "#D97706"
	default:
		return "🔔", "#6B7280"
	}
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	if !n.IsRead {
		n.IsRead = true
		now := time.Now()
		n.ReadAt = &now
		n.BeforeUpdate()
	}
}

// MarkAsDelivered marks the notification as delivered
func (n *Notification) MarkAsDelivered() {
	if !n.IsDelivered {
		n.IsDelivered = true
		now := time.Now()
		n.DeliveredAt = &now
		n.BeforeUpdate()
	}
}

// MarkAsClicked tracks when the notification was clicked
func (n *Notification) MarkAsClicked(interactionData map[string]interface{}) {
	now := time.Now()
	n.ClickedAt = &now
	n.InteractionData = interactionData
	n.MarkAsRead() // Also mark as read when clicked
}

// MarkAsDismissed tracks when the notification was dismissed
func (n *Notification) MarkAsDismissed() {
	now := time.Now()
	n.DismissedAt = &now
	n.MarkAsRead() // Also mark as read when dismissed
}

// IsExpired checks if the notification has expired
func (n *Notification) IsExpired() bool {
	return n.ExpiresAt != nil && n.ExpiresAt.Before(time.Now())
}

// ShouldSendViaEmail checks if notification should be sent via email
func (n *Notification) ShouldSendViaEmail(userPrefs NotificationPreferences) bool {
	if !userPrefs.EmailEnabled {
		return false
	}

	return n.isTypeEnabled(userPrefs) && !n.isInQuietHours(userPrefs)
}

// ShouldSendViaPush checks if notification should be sent via push
func (n *Notification) ShouldSendViaPush(userPrefs NotificationPreferences) bool {
	if !userPrefs.PushEnabled {
		return false
	}

	return n.isTypeEnabled(userPrefs) && !n.isInQuietHours(userPrefs)
}

// ShouldSendViaSMS checks if notification should be sent via SMS
func (n *Notification) ShouldSendViaSMS(userPrefs NotificationPreferences) bool {
	if !userPrefs.SMSEnabled {
		return false
	}

	// Only send high priority notifications via SMS
	return n.Priority == "high" && n.isTypeEnabled(userPrefs)
}

// isTypeEnabled checks if the notification type is enabled in user preferences
func (n *Notification) isTypeEnabled(prefs NotificationPreferences) bool {
	switch n.Type {
	case NotificationLike:
		return prefs.LikeNotifications
	case NotificationComment:
		return prefs.CommentNotifications
	case NotificationFollow:
		return prefs.FollowNotifications
	case NotificationMessage:
		return prefs.MessageNotifications
	case NotificationMention:
		return prefs.MentionNotifications
	case NotificationGroupInvite, NotificationGroupPost:
		return prefs.GroupNotifications
	case NotificationEventInvite, NotificationEventReminder:
		return prefs.EventNotifications
	case NotificationPostShare:
		return prefs.PostShareNotifications
	case NotificationStoryView:
		return prefs.StoryViewNotifications
	case NotificationFriendRequest:
		return prefs.FriendRequestNotifications
	default:
		return true
	}
}

// isInQuietHours checks if current time is within user's quiet hours
func (n *Notification) isInQuietHours(prefs NotificationPreferences) bool {
	if !prefs.QuietHoursEnabled {
		return false
	}

	now := time.Now()
	currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)

	start := time.Date(0, 1, 1, prefs.QuietHoursStart.Hour(), prefs.QuietHoursStart.Minute(), 0, 0, time.UTC)
	end := time.Date(0, 1, 1, prefs.QuietHoursEnd.Hour(), prefs.QuietHoursEnd.Minute(), 0, 0, time.UTC)

	if start.Before(end) {
		// Same day quiet hours (e.g., 22:00 - 08:00 next day)
		return currentTime.After(start) && currentTime.Before(end)
	} else {
		// Quiet hours cross midnight
		return currentTime.After(start) || currentTime.Before(end)
	}
}

// CanDelete checks if the notification can be deleted
func (n *Notification) CanDelete(currentUserID primitive.ObjectID) bool {
	return n.RecipientID == currentUserID && !n.IsDeleted()
}

// Utility functions for notifications

// CreateNotificationFromTemplate creates a notification from a predefined template
func CreateNotificationFromTemplate(notifType NotificationType, recipientID, actorID primitive.ObjectID, targetID *primitive.ObjectID, metadata map[string]interface{}) *Notification {
	notification := &Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        notifType,
		TargetID:    targetID,
		Metadata:    metadata,
		Priority:    "medium",
	}

	// Set title and message based on type
	notification.Title, notification.Message, notification.ActionText = getNotificationContent(notifType, metadata)

	// Set target type and URL
	if targetID != nil {
		notification.TargetType, notification.TargetURL = getTargetInfo(notifType, *targetID)
	}

	return notification
}

// getNotificationContent returns title, message, and action text for notification types
func getNotificationContent(notifType NotificationType, metadata map[string]interface{}) (string, string, string) {
	switch notifType {
	case NotificationLike:
		return "New Like", "Someone liked your post", "View Post"
	case NotificationComment:
		return "New Comment", "Someone commented on your post", "View Comment"
	case NotificationFollow:
		return "New Follower", "Someone started following you", "View Profile"
	case NotificationMessage:
		return "New Message", "You have a new message", "Read Message"
	case NotificationMention:
		return "You were mentioned", "Someone mentioned you in a post", "View Post"
	case NotificationGroupInvite:
		return "Group Invitation", "You were invited to join a group", "View Group"
	case NotificationEventInvite:
		return "Event Invitation", "You were invited to an event", "View Event"
	case NotificationFriendRequest:
		return "Friend Request", "Someone sent you a friend request", "View Request"
	case NotificationPostShare:
		return "Post Shared", "Someone shared your post", "View Post"
	case NotificationStoryView:
		return "Story View", "Someone viewed your story", "View Story"
	case NotificationGroupPost:
		return "New Group Post", "New post in your group", "View Post"
	case NotificationEventReminder:
		return "Event Reminder", "You have an upcoming event", "View Event"
	default:
		return "Notification", "You have a new notification", "View"
	}
}

// getTargetInfo returns target type and URL based on notification type and target ID
func getTargetInfo(notifType NotificationType, targetID primitive.ObjectID) (string, string) {
	targetIDStr := targetID.Hex()

	switch notifType {
	case NotificationLike, NotificationComment, NotificationPostShare, NotificationMention:
		return "post", "/posts/" + targetIDStr
	case NotificationFollow, NotificationFriendRequest:
		return "user", "/users/" + targetIDStr
	case NotificationMessage:
		return "conversation", "/conversations/" + targetIDStr
	case NotificationGroupInvite, NotificationGroupPost:
		return "group", "/groups/" + targetIDStr
	case NotificationEventInvite, NotificationEventReminder:
		return "event", "/events/" + targetIDStr
	case NotificationStoryView:
		return "story", "/stories/" + targetIDStr
	default:
		return "unknown", "/"
	}
}

// DefaultNotificationPreferences returns default notification preferences
func DefaultNotificationPreferences(userID primitive.ObjectID) NotificationPreferences {
	return NotificationPreferences{
		UserID:                     userID,
		EmailEnabled:               true,
		PushEnabled:                true,
		SMSEnabled:                 false,
		LikeNotifications:          true,
		CommentNotifications:       true,
		FollowNotifications:        true,
		MessageNotifications:       true,
		MentionNotifications:       true,
		GroupNotifications:         true,
		EventNotifications:         true,
		PostShareNotifications:     true,
		StoryViewNotifications:     false,
		FriendRequestNotifications: true,
		QuietHoursEnabled:          false,
		QuietHoursStart:            time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
		QuietHoursEnd:              time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		GroupSimilarNotifications:  true,
		DigestFrequency:            "immediate",
	}
}
