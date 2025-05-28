// models/message.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message represents a message in a conversation
type Message struct {
	BaseModel `bson:",inline"`

	// Message details
	ConversationID primitive.ObjectID `json:"conversation_id" bson:"conversation_id" validate:"required"`
	SenderID       primitive.ObjectID `json:"sender_id" bson:"sender_id" validate:"required"`
	Sender         UserResponse       `json:"sender,omitempty" bson:"-"` // Populated when querying

	// Content
	Content     string      `json:"content" bson:"content" validate:"max=5000"`
	ContentType ContentType `json:"content_type" bson:"content_type"`
	Media       []MediaInfo `json:"media,omitempty" bson:"media,omitempty"`

	// Message status
	Status      MessageStatus `json:"status" bson:"status"`
	SentAt      *time.Time    `json:"sent_at,omitempty" bson:"sent_at,omitempty"`
	DeliveredAt *time.Time    `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
	ReadAt      *time.Time    `json:"read_at,omitempty" bson:"read_at,omitempty"`

	// Message features
	IsEdited      bool                `json:"is_edited" bson:"is_edited"`
	EditedAt      *time.Time          `json:"edited_at,omitempty" bson:"edited_at,omitempty"`
	IsForwarded   bool                `json:"is_forwarded" bson:"is_forwarded"`
	ForwardedFrom *primitive.ObjectID `json:"forwarded_from,omitempty" bson:"forwarded_from,omitempty"`

	// Reply to another message
	ReplyToMessageID *primitive.ObjectID `json:"reply_to_message_id,omitempty" bson:"reply_to_message_id,omitempty"`
	ReplyToMessage   *MessageResponse    `json:"reply_to_message,omitempty" bson:"-"` // Populated when querying

	// Reactions on messages
	ReactionsCount map[ReactionType]int64 `json:"reactions_count,omitempty" bson:"reactions_count,omitempty"`

	// Message metadata
	Source    string `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	IPAddress string `json:"-" bson:"ip_address,omitempty"`
	UserAgent string `json:"-" bson:"user_agent,omitempty"`

	// Delivery tracking
	ReadBy []MessageReadReceipt `json:"read_by,omitempty" bson:"read_by,omitempty"` // For group conversations

	// Message priority and features
	Priority  string     `json:"priority,omitempty" bson:"priority,omitempty"`     // normal, high, urgent
	ExpiresAt *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"` // For disappearing messages
	IsExpired bool       `json:"is_expired" bson:"is_expired"`

	// Voice/Video messages
	Duration   int    `json:"duration,omitempty" bson:"duration,omitempty"`     // Duration in seconds for voice/video
	Transcript string `json:"transcript,omitempty" bson:"transcript,omitempty"` // AI transcription

	// Message threading
	ThreadID     *primitive.ObjectID `json:"thread_id,omitempty" bson:"thread_id,omitempty"`
	IsThreadRoot bool                `json:"is_thread_root" bson:"is_thread_root"`
	ThreadCount  int64               `json:"thread_count" bson:"thread_count"`
}

// MessageReadReceipt tracks when a user read a message
type MessageReadReceipt struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	ReadAt   time.Time          `json:"read_at" bson:"read_at"`
	Platform string             `json:"platform,omitempty" bson:"platform,omitempty"` // web, mobile, desktop
}

// MessageResponse represents the message data returned in API responses
type MessageResponse struct {
	ID               string                 `json:"id"`
	ConversationID   string                 `json:"conversation_id"`
	SenderID         string                 `json:"sender_id"`
	Sender           UserResponse           `json:"sender"`
	Content          string                 `json:"content"`
	ContentType      ContentType            `json:"content_type"`
	Media            []MediaInfo            `json:"media,omitempty"`
	Status           MessageStatus          `json:"status"`
	SentAt           *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt      *time.Time             `json:"delivered_at,omitempty"`
	ReadAt           *time.Time             `json:"read_at,omitempty"`
	IsEdited         bool                   `json:"is_edited"`
	EditedAt         *time.Time             `json:"edited_at,omitempty"`
	IsForwarded      bool                   `json:"is_forwarded"`
	ReplyToMessageID string                 `json:"reply_to_message_id,omitempty"`
	ReplyToMessage   *MessageResponse       `json:"reply_to_message,omitempty"`
	ReactionsCount   map[ReactionType]int64 `json:"reactions_count,omitempty"`
	ReadBy           []MessageReadReceipt   `json:"read_by,omitempty"`
	Priority         string                 `json:"priority,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	IsExpired        bool                   `json:"is_expired"`
	Duration         int                    `json:"duration,omitempty"`
	Transcript       string                 `json:"transcript,omitempty"`
	ThreadID         string                 `json:"thread_id,omitempty"`
	IsThreadRoot     bool                   `json:"is_thread_root"`
	ThreadCount      int64                  `json:"thread_count"`
	CreatedAt        time.Time              `json:"created_at"`
	TimeAgo          string                 `json:"time_ago,omitempty"`

	// User-specific context
	UserReaction ReactionType `json:"user_reaction,omitempty"`
	CanEdit      bool         `json:"can_edit,omitempty"`
	CanDelete    bool         `json:"can_delete,omitempty"`
	CanReact     bool         `json:"can_react,omitempty"`
	CanReply     bool         `json:"can_reply,omitempty"`
}

// CreateMessageRequest represents the request to send a message
type CreateMessageRequest struct {
	ConversationID   string      `json:"conversation_id" validate:"required"`
	Content          string      `json:"content" validate:"max=5000"`
	ContentType      ContentType `json:"content_type" validate:"required,oneof=text image video audio file gif"`
	Media            []MediaInfo `json:"media,omitempty"`
	ReplyToMessageID string      `json:"reply_to_message_id,omitempty"`
	Priority         string      `json:"priority,omitempty" validate:"omitempty,oneof=normal high urgent"`
	ExpiresAt        *time.Time  `json:"expires_at,omitempty"`
}

// UpdateMessageRequest represents the request to update a message
type UpdateMessageRequest struct {
	Content string      `json:"content" validate:"required,max=5000"`
	Media   []MediaInfo `json:"media,omitempty"`
}

// MessageSearchRequest represents message search parameters
type MessageSearchRequest struct {
	Query          string     `json:"query" validate:"required,min=1"`
	ConversationID string     `json:"conversation_id,omitempty"`
	SenderID       string     `json:"sender_id,omitempty"`
	ContentType    string     `json:"content_type,omitempty"`
	Before         *time.Time `json:"before,omitempty"`
	After          *time.Time `json:"after,omitempty"`
	Page           int        `json:"page,omitempty" validate:"min=1"`
	Limit          int        `json:"limit,omitempty" validate:"min=1,max=50"`
}

// MessageReactionRequest represents adding/removing reaction to a message
type MessageReactionRequest struct {
	ReactionType ReactionType `json:"reaction_type" validate:"required"`
	Action       string       `json:"action" validate:"required,oneof=add remove"`
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages     int64            `json:"total_messages"`
	TotalMediaFiles   int64            `json:"total_media_files"`
	AverageLength     float64          `json:"average_length"`
	ContentTypeCounts map[string]int64 `json:"content_type_counts"`
	ReactionCounts    map[string]int64 `json:"reaction_counts,omitempty"`
	MostActiveHour    int              `json:"most_active_hour,omitempty"`
	MostActiveDay     string           `json:"most_active_day,omitempty"`
}

// BeforeCreate sets default values before creating message
func (m *Message) BeforeCreate() {
	m.BaseModel.BeforeCreate()

	// Set default values
	m.Status = MessageSent
	now := time.Now()
	m.SentAt = &now
	m.IsEdited = false
	m.IsForwarded = false
	m.IsExpired = false
	m.IsThreadRoot = false
	m.ThreadCount = 0

	// Set default content type
	if m.ContentType == "" {
		m.ContentType = ContentTypeText
	}

	// Set default priority
	if m.Priority == "" {
		m.Priority = "normal"
	}

	// Initialize reactions count
	if m.ReactionsCount == nil {
		m.ReactionsCount = make(map[ReactionType]int64)
	}
}

// ToMessageResponse converts Message model to MessageResponse
func (m *Message) ToMessageResponse() MessageResponse {
	response := MessageResponse{
		ID:             m.ID.Hex(),
		ConversationID: m.ConversationID.Hex(),
		SenderID:       m.SenderID.Hex(),
		Content:        m.Content,
		ContentType:    m.ContentType,
		Media:          m.Media,
		Status:         m.Status,
		SentAt:         m.SentAt,
		DeliveredAt:    m.DeliveredAt,
		ReadAt:         m.ReadAt,
		IsEdited:       m.IsEdited,
		EditedAt:       m.EditedAt,
		IsForwarded:    m.IsForwarded,
		ReactionsCount: m.ReactionsCount,
		ReadBy:         m.ReadBy,
		Priority:       m.Priority,
		ExpiresAt:      m.ExpiresAt,
		IsExpired:      m.IsExpired,
		Duration:       m.Duration,
		Transcript:     m.Transcript,
		IsThreadRoot:   m.IsThreadRoot,
		ThreadCount:    m.ThreadCount,
		CreatedAt:      m.CreatedAt,
	}

	if m.ReplyToMessageID != nil {
		response.ReplyToMessageID = m.ReplyToMessageID.Hex()
	}

	if m.ThreadID != nil {
		response.ThreadID = m.ThreadID.Hex()
	}

	return response
}

// MarkAsDelivered marks the message as delivered
func (m *Message) MarkAsDelivered() {
	if m.Status == MessageSent {
		m.Status = MessageDelivered
		now := time.Now()
		m.DeliveredAt = &now
		m.BeforeUpdate()
	}
}

// MarkAsRead marks the message as read
func (m *Message) MarkAsRead(userID primitive.ObjectID) {
	if m.Status != MessageRead {
		m.Status = MessageRead
		now := time.Now()
		m.ReadAt = &now

		// Add to read receipts
		receipt := MessageReadReceipt{
			UserID: userID,
			ReadAt: now,
		}

		// Check if user already has a read receipt
		found := false
		for i, existing := range m.ReadBy {
			if existing.UserID == userID {
				m.ReadBy[i] = receipt
				found = true
				break
			}
		}

		if !found {
			m.ReadBy = append(m.ReadBy, receipt)
		}

		m.BeforeUpdate()
	}
}

// MarkAsEdited marks the message as edited
func (m *Message) MarkAsEdited() {
	m.IsEdited = true
	now := time.Now()
	m.EditedAt = &now
	m.BeforeUpdate()
}

// CheckExpiration checks and updates expiration status
func (m *Message) CheckExpiration() {
	if !m.IsExpired && m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		m.IsExpired = true
		m.BeforeUpdate()
	}
}

// CanEditMessage checks if a user can edit this message
func (m *Message) CanEditMessage(currentUserID primitive.ObjectID) bool {
	return m.SenderID == currentUserID && !m.IsDeleted() && !m.IsExpired
}

// CanDeleteMessage checks if a user can delete this message
func (m *Message) CanDeleteMessage(currentUserID primitive.ObjectID, userRole UserRole, isConversationAdmin bool) bool {
	// Sender can delete their own message
	if m.SenderID == currentUserID {
		return true
	}

	// Conversation admins can delete any message in the conversation
	if isConversationAdmin {
		return true
	}

	// Platform moderators and admins can delete any message
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// GetMessagePreview generates a preview text for the message
func (m *Message) GetMessagePreview() string {
	switch m.ContentType {
	case ContentTypeText:
		if len(m.Content) > 100 {
			return m.Content[:97] + "..."
		}
		return m.Content
	case ContentTypeImage:
		return "📷 Image"
	case ContentTypeVideo:
		return "🎥 Video"
	case ContentTypeGif:
		return "🎬 GIF"
	case ContentTypeFile:
		return "📎 File"
	case ContentTypeAudio:
		return "🎵 Audio"
	default:
		return "Message"
	}
}

// IsReply checks if this message is a reply to another message
func (m *Message) IsReply() bool {
	return m.ReplyToMessageID != nil
}

// IsInThread checks if this message is part of a thread
func (m *Message) IsInThread() bool {
	return m.ThreadID != nil
}

// AddReaction adds a reaction to the message
func (m *Message) AddReaction(reactionType ReactionType) {
	if m.ReactionsCount == nil {
		m.ReactionsCount = make(map[ReactionType]int64)
	}
	m.ReactionsCount[reactionType]++
	m.BeforeUpdate()
}

// RemoveReaction removes a reaction from the message
func (m *Message) RemoveReaction(reactionType ReactionType) {
	if m.ReactionsCount != nil {
		if count, exists := m.ReactionsCount[reactionType]; exists && count > 0 {
			m.ReactionsCount[reactionType]--
			if m.ReactionsCount[reactionType] == 0 {
				delete(m.ReactionsCount, reactionType)
			}
			m.BeforeUpdate()
		}
	}
}

// GetTotalReactions returns total number of reactions
func (m *Message) GetTotalReactions() int64 {
	var total int64
	for _, count := range m.ReactionsCount {
		total += count
	}
	return total
}

// MarkAsForwarded marks the message as forwarded
func (m *Message) MarkAsForwarded(originalMessageID primitive.ObjectID) {
	m.IsForwarded = true
	m.ForwardedFrom = &originalMessageID
	m.BeforeUpdate()
}

// CreateThread creates a new thread from this message
func (m *Message) CreateThread() {
	if !m.IsThreadRoot {
		m.IsThreadRoot = true
		m.ThreadID = &m.ID // Thread ID is the same as message ID for root messages
		m.ThreadCount = 0
		m.BeforeUpdate()
	}
}

// IncrementThreadCount increments the thread reply count
func (m *Message) IncrementThreadCount() {
	if m.IsThreadRoot {
		m.ThreadCount++
		m.BeforeUpdate()
	}
}

// DecrementThreadCount decrements the thread reply count
func (m *Message) DecrementThreadCount() {
	if m.IsThreadRoot && m.ThreadCount > 0 {
		m.ThreadCount--
		m.BeforeUpdate()
	}
}
