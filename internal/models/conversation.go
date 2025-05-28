// models/conversation.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversation represents a conversation between users
type Conversation struct {
	BaseModel `bson:",inline"`

	// Conversation details
	Type        string `json:"type" bson:"type"`                       // "direct", "group"
	Title       string `json:"title,omitempty" bson:"title,omitempty"` // For group conversations
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty" bson:"avatar_url,omitempty"`

	// Participants
	Participants    []primitive.ObjectID      `json:"participants" bson:"participants" validate:"required,min=2"`
	ParticipantInfo []ConversationParticipant `json:"participant_info,omitempty" bson:"participant_info,omitempty"`
	AdminIDs        []primitive.ObjectID      `json:"admin_ids,omitempty" bson:"admin_ids,omitempty"` // For group conversations
	CreatedBy       primitive.ObjectID        `json:"created_by" bson:"created_by"`

	// Last message info (for conversation list)
	LastMessageID      *primitive.ObjectID `json:"last_message_id,omitempty" bson:"last_message_id,omitempty"`
	LastMessage        *MessageResponse    `json:"last_message,omitempty" bson:"-"` // Populated when querying
	LastMessageAt      *time.Time          `json:"last_message_at,omitempty" bson:"last_message_at,omitempty"`
	LastMessagePreview string              `json:"last_message_preview,omitempty" bson:"last_message_preview,omitempty"`
	LastActivityAt     *time.Time          `json:"last_activity_at,omitempty" bson:"last_activity_at,omitempty"`

	// Conversation settings
	IsArchived        bool `json:"is_archived" bson:"is_archived"`
	IsMuted           bool `json:"is_muted" bson:"is_muted"`
	IsLocked          bool `json:"is_locked" bson:"is_locked"` // Only admins can send messages
	AllowInvites      bool `json:"allow_invites" bson:"allow_invites"`
	AllowMediaSharing bool `json:"allow_media_sharing" bson:"allow_media_sharing"`

	// Privacy and moderation
	IsPrivate      bool       `json:"is_private" bson:"is_private"`
	JoinCode       string     `json:"join_code,omitempty" bson:"join_code,omitempty"` // For group invites
	JoinCodeExpiry *time.Time `json:"join_code_expiry,omitempty" bson:"join_code_expiry,omitempty"`

	// Statistics
	MessagesCount      int64 `json:"messages_count" bson:"messages_count"`
	ActiveMembersCount int64 `json:"active_members_count" bson:"active_members_count"`

	// Features
	HasPinnedMessages bool                 `json:"has_pinned_messages" bson:"has_pinned_messages"`
	PinnedMessages    []primitive.ObjectID `json:"pinned_messages,omitempty" bson:"pinned_messages,omitempty"`

	// Encryption (for future implementation)
	IsEncrypted   bool   `json:"is_encrypted" bson:"is_encrypted"`
	EncryptionKey string `json:"-" bson:"encryption_key,omitempty"`

	// Conversation metadata
	Tags         []string               `json:"tags,omitempty" bson:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`

	// Auto-moderation
	AutoDeleteAfter  *time.Duration `json:"auto_delete_after,omitempty" bson:"auto_delete_after,omitempty"`
	MessageRetention int64          `json:"message_retention,omitempty" bson:"message_retention,omitempty"` // Days

	// Group-specific features (when type = "group")
	MaxParticipants int64  `json:"max_participants,omitempty" bson:"max_participants,omitempty"`
	IsPublic        bool   `json:"is_public" bson:"is_public"`
	Category        string `json:"category,omitempty" bson:"category,omitempty"`

	// Conversation state
	IsActive      bool                `json:"is_active" bson:"is_active"`
	DeactivatedAt *time.Time          `json:"deactivated_at,omitempty" bson:"deactivated_at,omitempty"`
	DeactivatedBy *primitive.ObjectID `json:"deactivated_by,omitempty" bson:"deactivated_by,omitempty"`
}

// ConversationParticipant represents participant-specific conversation settings
type ConversationParticipant struct {
	UserID               primitive.ObjectID  `json:"user_id" bson:"user_id"`
	User                 UserResponse        `json:"user,omitempty" bson:"-"` // Populated when querying
	JoinedAt             time.Time           `json:"joined_at" bson:"joined_at"`
	LeftAt               *time.Time          `json:"left_at,omitempty" bson:"left_at,omitempty"`
	Role                 string              `json:"role" bson:"role"` // admin, moderator, member
	IsAdmin              bool                `json:"is_admin" bson:"is_admin"`
	IsMuted              bool                `json:"is_muted" bson:"is_muted"`
	LastReadMessageID    *primitive.ObjectID `json:"last_read_message_id,omitempty" bson:"last_read_message_id,omitempty"`
	LastReadAt           *time.Time          `json:"last_read_at,omitempty" bson:"last_read_at,omitempty"`
	UnreadCount          int64               `json:"unread_count" bson:"unread_count"`
	NotificationsEnabled bool                `json:"notifications_enabled" bson:"notifications_enabled"`
	Nickname             string              `json:"nickname,omitempty" bson:"nickname,omitempty"` // Custom name in this conversation

	// Participant permissions
	CanSendMessages  bool `json:"can_send_messages" bson:"can_send_messages"`
	CanSendMedia     bool `json:"can_send_media" bson:"can_send_media"`
	CanAddMembers    bool `json:"can_add_members" bson:"can_add_members"`
	CanRemoveMembers bool `json:"can_remove_members" bson:"can_remove_members"`
	CanChangeInfo    bool `json:"can_change_info" bson:"can_change_info"`
	CanPinMessages   bool `json:"can_pin_messages" bson:"can_pin_messages"`

	// Activity tracking
	LastActiveAt    *time.Time `json:"last_active_at,omitempty" bson:"last_active_at,omitempty"`
	MessagesCount   int64      `json:"messages_count" bson:"messages_count"`
	IsTyping        bool       `json:"is_typing" bson:"is_typing"`
	TypingStartedAt *time.Time `json:"typing_started_at,omitempty" bson:"typing_started_at,omitempty"`

	// Invitation info
	InvitedBy  *primitive.ObjectID `json:"invited_by,omitempty" bson:"invited_by,omitempty"`
	InvitedAt  *time.Time          `json:"invited_at,omitempty" bson:"invited_at,omitempty"`
	JoinMethod string              `json:"join_method,omitempty" bson:"join_method,omitempty"` // invited, joined, added
}

// ConversationResponse represents the conversation data returned in API responses
type ConversationResponse struct {
	ID                 string                    `json:"id"`
	Type               string                    `json:"type"`
	Title              string                    `json:"title,omitempty"`
	Description        string                    `json:"description,omitempty"`
	AvatarURL          string                    `json:"avatar_url,omitempty"`
	Participants       []UserResponse            `json:"participants"`
	ParticipantInfo    []ConversationParticipant `json:"participant_info,omitempty"`
	AdminIDs           []string                  `json:"admin_ids,omitempty"`
	CreatedBy          string                    `json:"created_by"`
	LastMessage        *MessageResponse          `json:"last_message,omitempty"`
	LastMessageAt      *time.Time                `json:"last_message_at,omitempty"`
	LastMessagePreview string                    `json:"last_message_preview,omitempty"`
	LastActivityAt     *time.Time                `json:"last_activity_at,omitempty"`
	IsArchived         bool                      `json:"is_archived"`
	IsMuted            bool                      `json:"is_muted"`
	IsLocked           bool                      `json:"is_locked"`
	AllowInvites       bool                      `json:"allow_invites"`
	AllowMediaSharing  bool                      `json:"allow_media_sharing"`
	IsPrivate          bool                      `json:"is_private"`
	MessagesCount      int64                     `json:"messages_count"`
	ActiveMembersCount int64                     `json:"active_members_count"`
	HasPinnedMessages  bool                      `json:"has_pinned_messages"`
	IsEncrypted        bool                      `json:"is_encrypted"`
	Tags               []string                  `json:"tags,omitempty"`
	MaxParticipants    int64                     `json:"max_participants,omitempty"`
	IsPublic           bool                      `json:"is_public"`
	Category           string                    `json:"category,omitempty"`
	IsActive           bool                      `json:"is_active"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`

	// User-specific context
	UnreadCount       int64          `json:"unread_count,omitempty"`
	LastReadAt        *time.Time     `json:"last_read_at,omitempty"`
	IsUserAdmin       bool           `json:"is_user_admin,omitempty"`
	UserNotifications bool           `json:"user_notifications,omitempty"`
	UserRole          string         `json:"user_role,omitempty"`
	CanSendMessages   bool           `json:"can_send_messages,omitempty"`
	CanAddMembers     bool           `json:"can_add_members,omitempty"`
	TypingUsers       []UserResponse `json:"typing_users,omitempty"`
}

// CreateConversationRequest represents the request to create a conversation
type CreateConversationRequest struct {
	Type              string   `json:"type" validate:"required,oneof=direct group"`
	Title             string   `json:"title,omitempty" validate:"max=100"`
	Description       string   `json:"description,omitempty" validate:"max=500"`
	ParticipantIDs    []string `json:"participant_ids" validate:"required,min=1,max=50"`
	IsPrivate         bool     `json:"is_private"`
	AllowInvites      bool     `json:"allow_invites"`
	AllowMediaSharing bool     `json:"allow_media_sharing"`
	InitialMessage    string   `json:"initial_message,omitempty" validate:"max=5000"`
	MaxParticipants   int64    `json:"max_participants,omitempty"`
	Category          string   `json:"category,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

// UpdateConversationRequest represents the request to update a conversation
type UpdateConversationRequest struct {
	Title             *string  `json:"title,omitempty" validate:"omitempty,max=100"`
	Description       *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	AvatarURL         *string  `json:"avatar_url,omitempty"`
	AllowInvites      *bool    `json:"allow_invites,omitempty"`
	AllowMediaSharing *bool    `json:"allow_media_sharing,omitempty"`
	IsLocked          *bool    `json:"is_locked,omitempty"`
	IsPrivate         *bool    `json:"is_private,omitempty"`
	MaxParticipants   *int64   `json:"max_participants,omitempty"`
	Category          *string  `json:"category,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

// AddParticipantsRequest represents the request to add participants to a conversation
type AddParticipantsRequest struct {
	ParticipantIDs []string `json:"participant_ids" validate:"required,min=1,max=20"`
}

// UpdateParticipantRequest represents the request to update participant settings
type UpdateParticipantRequest struct {
	Role                 *string `json:"role,omitempty" validate:"omitempty,oneof=admin moderator member"`
	NotificationsEnabled *bool   `json:"notifications_enabled,omitempty"`
	Nickname             *string `json:"nickname,omitempty" validate:"omitempty,max=50"`
	IsMuted              *bool   `json:"is_muted,omitempty"`
}

// TypingIndicatorRequest represents typing indicator updates
type TypingIndicatorRequest struct {
	ConversationID string `json:"conversation_id" validate:"required"`
	IsTyping       bool   `json:"is_typing"`
}

// ConversationStatsResponse represents conversation statistics
type ConversationStatsResponse struct {
	ConversationID      string `json:"conversation_id"`
	MessagesCount       int64  `json:"messages_count"`
	ActiveMembersCount  int64  `json:"active_members_count"`
	TotalMembersCount   int64  `json:"total_members_count"`
	UnreadMessagesCount int64  `json:"unread_messages_count"`
	LastActivityHours   int64  `json:"last_activity_hours"`
}

// Methods for Conversation model

// BeforeCreate sets default values before creating conversation
func (c *Conversation) BeforeCreate() {
	c.BaseModel.BeforeCreate()

	// Set default values
	c.IsArchived = false
	c.IsMuted = false
	c.IsLocked = false
	c.IsPrivate = false
	c.AllowInvites = true
	c.AllowMediaSharing = true
	c.MessagesCount = 0
	c.ActiveMembersCount = int64(len(c.Participants))
	c.IsEncrypted = false
	c.HasPinnedMessages = false
	c.IsActive = true
	c.IsPublic = false

	// Initialize participant info
	if len(c.ParticipantInfo) == 0 {
		c.ParticipantInfo = make([]ConversationParticipant, len(c.Participants))
		now := time.Now()

		for i, participantID := range c.Participants {
			c.ParticipantInfo[i] = ConversationParticipant{
				UserID:               participantID,
				JoinedAt:             now,
				Role:                 "member",
				IsAdmin:              participantID == c.CreatedBy,
				IsMuted:              false,
				UnreadCount:          0,
				NotificationsEnabled: true,
				CanSendMessages:      true,
				CanSendMedia:         true,
				CanAddMembers:        false,
				CanRemoveMembers:     false,
				CanChangeInfo:        false,
				CanPinMessages:       false,
				MessagesCount:        0,
				IsTyping:             false,
				JoinMethod:           "created",
			}
		}
	}

	// Set creator as admin for group conversations
	if c.Type == "group" {
		c.AdminIDs = []primitive.ObjectID{c.CreatedBy}
		// Update creator's permissions
		for i, info := range c.ParticipantInfo {
			if info.UserID == c.CreatedBy {
				c.ParticipantInfo[i].Role = "admin"
				c.ParticipantInfo[i].IsAdmin = true
				c.ParticipantInfo[i].CanAddMembers = true
				c.ParticipantInfo[i].CanRemoveMembers = true
				c.ParticipantInfo[i].CanChangeInfo = true
				c.ParticipantInfo[i].CanPinMessages = true
				break
			}
		}
	}

	// Set default max participants for groups
	if c.Type == "group" && c.MaxParticipants == 0 {
		c.MaxParticipants = 500 // Default max group size
	}
}

// ToConversationResponse converts Conversation model to ConversationResponse
func (c *Conversation) ToConversationResponse() ConversationResponse {
	response := ConversationResponse{
		ID:                 c.ID.Hex(),
		Type:               c.Type,
		Title:              c.Title,
		Description:        c.Description,
		AvatarURL:          c.AvatarURL,
		ParticipantInfo:    c.ParticipantInfo,
		CreatedBy:          c.CreatedBy.Hex(),
		LastMessageAt:      c.LastMessageAt,
		LastMessagePreview: c.LastMessagePreview,
		LastActivityAt:     c.LastActivityAt,
		IsArchived:         c.IsArchived,
		IsMuted:            c.IsMuted,
		IsLocked:           c.IsLocked,
		AllowInvites:       c.AllowInvites,
		AllowMediaSharing:  c.AllowMediaSharing,
		IsPrivate:          c.IsPrivate,
		MessagesCount:      c.MessagesCount,
		ActiveMembersCount: c.ActiveMembersCount,
		HasPinnedMessages:  c.HasPinnedMessages,
		IsEncrypted:        c.IsEncrypted,
		Tags:               c.Tags,
		MaxParticipants:    c.MaxParticipants,
		IsPublic:           c.IsPublic,
		Category:           c.Category,
		IsActive:           c.IsActive,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}

	// Convert ObjectIDs to strings
	if len(c.AdminIDs) > 0 {
		response.AdminIDs = make([]string, len(c.AdminIDs))
		for i, adminID := range c.AdminIDs {
			response.AdminIDs[i] = adminID.Hex()
		}
	}

	return response
}

// AddParticipant adds a participant to the conversation
func (c *Conversation) AddParticipant(userID primitive.ObjectID, invitedBy *primitive.ObjectID) {
	// Check if user is already a participant
	for _, participantID := range c.Participants {
		if participantID == userID {
			return
		}
	}

	// Add to participants list
	c.Participants = append(c.Participants, userID)
	c.ActiveMembersCount++

	// Add participant info
	now := time.Now()
	participantInfo := ConversationParticipant{
		UserID:               userID,
		JoinedAt:             now,
		Role:                 "member",
		IsAdmin:              false,
		IsMuted:              false,
		UnreadCount:          0,
		NotificationsEnabled: true,
		CanSendMessages:      true,
		CanSendMedia:         true,
		CanAddMembers:        false,
		CanRemoveMembers:     false,
		CanChangeInfo:        false,
		CanPinMessages:       false,
		MessagesCount:        0,
		IsTyping:             false,
	}

	if invitedBy != nil {
		participantInfo.InvitedBy = invitedBy
		participantInfo.InvitedAt = &now
		participantInfo.JoinMethod = "invited"
	} else {
		participantInfo.JoinMethod = "joined"
	}

	c.ParticipantInfo = append(c.ParticipantInfo, participantInfo)
	c.BeforeUpdate()
}

// RemoveParticipant removes a participant from the conversation
func (c *Conversation) RemoveParticipant(userID primitive.ObjectID) {
	// Remove from participants list
	for i, participantID := range c.Participants {
		if participantID == userID {
			c.Participants = append(c.Participants[:i], c.Participants[i+1:]...)
			c.ActiveMembersCount--
			break
		}
	}

	// Update participant info
	for i, info := range c.ParticipantInfo {
		if info.UserID == userID {
			now := time.Now()
			c.ParticipantInfo[i].LeftAt = &now
			break
		}
	}

	// Remove from admins if they were an admin
	for i, adminID := range c.AdminIDs {
		if adminID == userID {
			c.AdminIDs = append(c.AdminIDs[:i], c.AdminIDs[i+1:]...)
			break
		}
	}

	c.BeforeUpdate()
}

// IsParticipant checks if a user is a participant in the conversation
func (c *Conversation) IsParticipant(userID primitive.ObjectID) bool {
	for _, participantID := range c.Participants {
		if participantID == userID {
			return true
		}
	}
	return false
}

// IsAdmin checks if a user is an admin of the conversation
func (c *Conversation) IsAdmin(userID primitive.ObjectID) bool {
	for _, adminID := range c.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// UpdateLastMessage updates the last message information
func (c *Conversation) UpdateLastMessage(messageID primitive.ObjectID, messagePreview string) {
	c.LastMessageID = &messageID
	c.LastMessagePreview = messagePreview
	now := time.Now()
	c.LastMessageAt = &now
	c.LastActivityAt = &now
	c.MessagesCount++
	c.BeforeUpdate()
}

// GetUnreadCount gets unread message count for a specific user
func (c *Conversation) GetUnreadCount(userID primitive.ObjectID) int64 {
	for _, info := range c.ParticipantInfo {
		if info.UserID == userID {
			return info.UnreadCount
		}
	}
	return 0
}

// UpdateUnreadCount updates unread message count for a specific user
func (c *Conversation) UpdateUnreadCount(userID primitive.ObjectID, count int64) {
	for i, info := range c.ParticipantInfo {
		if info.UserID == userID {
			c.ParticipantInfo[i].UnreadCount = count
			break
		}
	}
	c.BeforeUpdate()
}

// MarkAsRead marks messages as read for a specific user
func (c *Conversation) MarkAsRead(userID primitive.ObjectID, lastReadMessageID primitive.ObjectID) {
	for i, info := range c.ParticipantInfo {
		if info.UserID == userID {
			c.ParticipantInfo[i].LastReadMessageID = &lastReadMessageID
			now := time.Now()
			c.ParticipantInfo[i].LastReadAt = &now
			c.ParticipantInfo[i].UnreadCount = 0
			break
		}
	}
	c.BeforeUpdate()
}

// SetTypingStatus sets typing status for a user
func (c *Conversation) SetTypingStatus(userID primitive.ObjectID, isTyping bool) {
	for i, info := range c.ParticipantInfo {
		if info.UserID == userID {
			c.ParticipantInfo[i].IsTyping = isTyping
			if isTyping {
				now := time.Now()
				c.ParticipantInfo[i].TypingStartedAt = &now
			} else {
				c.ParticipantInfo[i].TypingStartedAt = nil
			}
			break
		}
	}
	c.BeforeUpdate()
}

// GetTypingUsers returns list of users currently typing
func (c *Conversation) GetTypingUsers() []primitive.ObjectID {
	var typingUsers []primitive.ObjectID
	cutoff := time.Now().Add(-30 * time.Second) // Consider typing expired after 30 seconds

	for _, info := range c.ParticipantInfo {
		if info.IsTyping && info.TypingStartedAt != nil && info.TypingStartedAt.After(cutoff) {
			typingUsers = append(typingUsers, info.UserID)
		}
	}

	return typingUsers
}

// PinMessage pins a message in the conversation
func (c *Conversation) PinMessage(messageID primitive.ObjectID) {
	// Check if already pinned
	for _, pinnedID := range c.PinnedMessages {
		if pinnedID == messageID {
			return
		}
	}

	c.PinnedMessages = append(c.PinnedMessages, messageID)
	c.HasPinnedMessages = true
	c.BeforeUpdate()
}

// UnpinMessage unpins a message from the conversation
func (c *Conversation) UnpinMessage(messageID primitive.ObjectID) {
	for i, pinnedID := range c.PinnedMessages {
		if pinnedID == messageID {
			c.PinnedMessages = append(c.PinnedMessages[:i], c.PinnedMessages[i+1:]...)
			c.HasPinnedMessages = len(c.PinnedMessages) > 0
			c.BeforeUpdate()
			return
		}
	}
}

// GenerateDirectConversationTitle generates a title for direct conversations
func (c *Conversation) GenerateDirectConversationTitle(currentUserID primitive.ObjectID, users map[primitive.ObjectID]UserResponse) string {
	if c.Type != "direct" {
		return c.Title
	}

	for participantID := range users {
		if participantID != currentUserID {
			if user, exists := users[participantID]; exists {
				return user.DisplayName
			}
		}
	}

	return "Direct Message"
}

// CanSendMessages checks if a user can send messages in this conversation
func (c *Conversation) CanSendMessages(userID primitive.ObjectID) bool {
	if !c.IsActive || c.IsLocked {
		return c.IsAdmin(userID)
	}

	for _, info := range c.ParticipantInfo {
		if info.UserID == userID {
			return info.CanSendMessages && info.LeftAt == nil
		}
	}

	return false
}

// CanAddMembers checks if a user can add members to this conversation
func (c *Conversation) CanAddMembers(userID primitive.ObjectID) bool {
	if !c.AllowInvites {
		return c.IsAdmin(userID)
	}

	for _, info := range c.ParticipantInfo {
		if info.UserID == userID {
			return info.CanAddMembers || info.IsAdmin
		}
	}

	return false
}

// Deactivate deactivates the conversation
func (c *Conversation) Deactivate(deactivatedBy primitive.ObjectID) {
	c.IsActive = false
	now := time.Now()
	c.DeactivatedAt = &now
	c.DeactivatedBy = &deactivatedBy
	c.BeforeUpdate()
}

// Reactivate reactivates the conversation
func (c *Conversation) Reactivate() {
	c.IsActive = true
	c.DeactivatedAt = nil
	c.DeactivatedBy = nil
	c.BeforeUpdate()
}

// Archive archives the conversation for a specific user
func (c *Conversation) Archive() {
	c.IsArchived = true
	c.BeforeUpdate()
}

// Unarchive unarchives the conversation for a specific user
func (c *Conversation) Unarchive() {
	c.IsArchived = false
	c.BeforeUpdate()
}

// Mute mutes the conversation for a specific user
func (c *Conversation) Mute() {
	c.IsMuted = true
	c.BeforeUpdate()
}

// Unmute unmutes the conversation for a specific user
func (c *Conversation) Unmute() {
	c.IsMuted = false
	c.BeforeUpdate()
}

// GetParticipantRole returns the role of a specific participant
func (c *Conversation) GetParticipantRole(userID primitive.ObjectID) string {
	for _, info := range c.ParticipantInfo {
		if info.UserID == userID {
			return info.Role
		}
	}
	return ""
}

// UpdateParticipantRole updates the role of a specific participant
func (c *Conversation) UpdateParticipantRole(userID primitive.ObjectID, newRole string) {
	for i, info := range c.ParticipantInfo {
		if info.UserID == userID {
			c.ParticipantInfo[i].Role = newRole
			c.ParticipantInfo[i].IsAdmin = (newRole == "admin")

			// Update permissions based on role
			if newRole == "admin" {
				c.ParticipantInfo[i].CanAddMembers = true
				c.ParticipantInfo[i].CanRemoveMembers = true
				c.ParticipantInfo[i].CanChangeInfo = true
				c.ParticipantInfo[i].CanPinMessages = true
			} else if newRole == "moderator" {
				c.ParticipantInfo[i].CanAddMembers = true
				c.ParticipantInfo[i].CanPinMessages = true
			}

			c.BeforeUpdate()
			break
		}
	}
}
