// models/mention.go
package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mention represents a user mention in content (posts, comments, stories, etc.)
type Mention struct {
	BaseModel `bson:",inline"`

	// Mention Details
	MentionerID primitive.ObjectID `json:"mentioner_id" bson:"mentioner_id" validate:"required"` // User who made the mention
	MentionedID primitive.ObjectID `json:"mentioned_id" bson:"mentioned_id" validate:"required"` // User who was mentioned

	// Context of Mention
	ContentType string             `json:"content_type" bson:"content_type" validate:"required"` // post, comment, story, message
	ContentID   primitive.ObjectID `json:"content_id" bson:"content_id" validate:"required"`

	// Position in Content (for rich text mentions)
	StartPosition int    `json:"start_position,omitempty" bson:"start_position,omitempty"`
	EndPosition   int    `json:"end_position,omitempty" bson:"end_position,omitempty"`
	MentionText   string `json:"mention_text" bson:"mention_text"` // The actual text (e.g., "@username")

	// Users (populated when querying)
	Mentioner UserResponse `json:"mentioner,omitempty" bson:"-"`
	Mentioned UserResponse `json:"mentioned,omitempty" bson:"-"`

	// Mention Status
	IsActive   bool       `json:"is_active" bson:"is_active"`
	IsNotified bool       `json:"is_notified" bson:"is_notified"`
	NotifiedAt *time.Time `json:"notified_at,omitempty" bson:"notified_at,omitempty"`
	IsRead     bool       `json:"is_read" bson:"is_read"`
	ReadAt     *time.Time `json:"read_at,omitempty" bson:"read_at,omitempty"`

	// Privacy and Blocking
	IsVisible bool `json:"is_visible" bson:"is_visible"` // Can the mentioned user see this mention
	IsBlocked bool `json:"is_blocked" bson:"is_blocked"` // Blocked by mentioned user

	// Context Information
	ParentContentID *primitive.ObjectID `json:"parent_content_id,omitempty" bson:"parent_content_id,omitempty"` // For mentions in replies
	GroupID         *primitive.ObjectID `json:"group_id,omitempty" bson:"group_id,omitempty"`                   // If mention is in group content
	EventID         *primitive.ObjectID `json:"event_id,omitempty" bson:"event_id,omitempty"`                   // If mention is in event content

	// Analytics
	ClickCount       int64   `json:"click_count" bson:"click_count"`
	ViewCount        int64   `json:"view_count" bson:"view_count"`
	InteractionScore float64 `json:"interaction_score" bson:"interaction_score"`

	// Additional Context
	Source    string                 `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	IPAddress string                 `json:"-" bson:"ip_address,omitempty"`
	UserAgent string                 `json:"-" bson:"user_agent,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// MentionResponse represents mention data returned in API responses
type MentionResponse struct {
	ID            string       `json:"id"`
	MentionerID   string       `json:"mentioner_id"`
	MentionedID   string       `json:"mentioned_id"`
	Mentioner     UserResponse `json:"mentioner"`
	Mentioned     UserResponse `json:"mentioned"`
	ContentType   string       `json:"content_type"`
	ContentID     string       `json:"content_id"`
	StartPosition int          `json:"start_position,omitempty"`
	EndPosition   int          `json:"end_position,omitempty"`
	MentionText   string       `json:"mention_text"`
	IsActive      bool         `json:"is_active"`
	IsNotified    bool         `json:"is_notified"`
	NotifiedAt    *time.Time   `json:"notified_at,omitempty"`
	IsRead        bool         `json:"is_read"`
	ReadAt        *time.Time   `json:"read_at,omitempty"`
	IsVisible     bool         `json:"is_visible"`
	GroupID       string       `json:"group_id,omitempty"`
	EventID       string       `json:"event_id,omitempty"`
	ClickCount    int64        `json:"click_count"`
	ViewCount     int64        `json:"view_count"`
	CreatedAt     time.Time    `json:"created_at"`
	TimeAgo       string       `json:"time_ago,omitempty"`

	// Context Data (populated based on content type)
	PostContent    *PostResponse    `json:"post_content,omitempty"`
	CommentContent *CommentResponse `json:"comment_content,omitempty"`
	StoryContent   *StoryResponse   `json:"story_content,omitempty"`
}

// CreateMentionRequest represents request to create a mention
type CreateMentionRequest struct {
	MentionedID     string `json:"mentioned_id" validate:"required"`
	ContentType     string `json:"content_type" validate:"required,oneof=post comment story message"`
	ContentID       string `json:"content_id" validate:"required"`
	StartPosition   int    `json:"start_position,omitempty"`
	EndPosition     int    `json:"end_position,omitempty"`
	MentionText     string `json:"mention_text" validate:"required"`
	ParentContentID string `json:"parent_content_id,omitempty"`
	GroupID         string `json:"group_id,omitempty"`
	EventID         string `json:"event_id,omitempty"`
}

// UpdateMentionRequest represents request to update a mention
type UpdateMentionRequest struct {
	IsVisible *bool `json:"is_visible,omitempty"`
	IsBlocked *bool `json:"is_blocked,omitempty"`
}

// MentionStatsResponse represents mention statistics
type MentionStatsResponse struct {
	TotalMentions  int64              `json:"total_mentions"`
	UnreadMentions int64              `json:"unread_mentions"`
	TodayMentions  int64              `json:"today_mentions"`
	WeekMentions   int64              `json:"week_mentions"`
	MonthMentions  int64              `json:"month_mentions"`
	TopMentioners  []UserMentionStats `json:"top_mentioners,omitempty"`
	MentionsByType map[string]int64   `json:"mentions_by_type,omitempty"`
}

// UserMentionStats represents mention statistics for a user
type UserMentionStats struct {
	UserID       string       `json:"user_id"`
	User         UserResponse `json:"user"`
	MentionCount int64        `json:"mention_count"`
	LastMention  time.Time    `json:"last_mention"`
	ContentTypes []string     `json:"content_types,omitempty"`
}

// MentionNotificationResponse represents mentions for notifications
type MentionNotificationResponse struct {
	ID             string       `json:"id"`
	Mentioner      UserResponse `json:"mentioner"`
	ContentType    string       `json:"content_type"`
	ContentPreview string       `json:"content_preview"`
	IsRead         bool         `json:"is_read"`
	CreatedAt      time.Time    `json:"created_at"`
	TimeAgo        string       `json:"time_ago"`
}

// BulkMentionOperation represents bulk operations on mentions
type BulkMentionOperation struct {
	MentionIDs []string `json:"mention_ids" validate:"required,min=1,max=100"`
	Operation  string   `json:"operation" validate:"required,oneof=mark_read mark_unread delete block unblock"`
}

// BeforeCreate sets default values before creating mention
func (m *Mention) BeforeCreate() {
	m.BaseModel.BeforeCreate()

	// Set default values
	m.IsActive = true
	m.IsNotified = false
	m.IsRead = false
	m.IsVisible = true
	m.IsBlocked = false
	m.ClickCount = 0
	m.ViewCount = 0
	m.InteractionScore = 0.0

	// Set source if not provided
	if m.Source == "" {
		m.Source = "web"
	}
}

// ToMentionResponse converts Mention to MentionResponse
func (m *Mention) ToMentionResponse() MentionResponse {
	response := MentionResponse{
		ID:            m.ID.Hex(),
		MentionerID:   m.MentionerID.Hex(),
		MentionedID:   m.MentionedID.Hex(),
		ContentType:   m.ContentType,
		ContentID:     m.ContentID.Hex(),
		StartPosition: m.StartPosition,
		EndPosition:   m.EndPosition,
		MentionText:   m.MentionText,
		IsActive:      m.IsActive,
		IsNotified:    m.IsNotified,
		NotifiedAt:    m.NotifiedAt,
		IsRead:        m.IsRead,
		ReadAt:        m.ReadAt,
		IsVisible:     m.IsVisible,
		ClickCount:    m.ClickCount,
		ViewCount:     m.ViewCount,
		CreatedAt:     m.CreatedAt,
	}

	if m.GroupID != nil {
		response.GroupID = m.GroupID.Hex()
	}

	if m.EventID != nil {
		response.EventID = m.EventID.Hex()
	}

	return response
}

// ToMentionNotificationResponse converts to notification format
func (m *Mention) ToMentionNotificationResponse(contentPreview string) MentionNotificationResponse {
	return MentionNotificationResponse{
		ID:             m.ID.Hex(),
		ContentType:    m.ContentType,
		ContentPreview: contentPreview,
		IsRead:         m.IsRead,
		CreatedAt:      m.CreatedAt,
	}
}

// MarkAsNotified marks the mention as notified
func (m *Mention) MarkAsNotified() {
	m.IsNotified = true
	now := time.Now()
	m.NotifiedAt = &now
	m.BeforeUpdate()
}

// MarkAsRead marks the mention as read
func (m *Mention) MarkAsRead() {
	m.IsRead = true
	now := time.Now()
	m.ReadAt = &now
	m.BeforeUpdate()
}

// MarkAsUnread marks the mention as unread
func (m *Mention) MarkAsUnread() {
	m.IsRead = false
	m.ReadAt = nil
	m.BeforeUpdate()
}


// Hide hides this mention from the mentioned user
func (m *Mention) Hide() {
	m.IsVisible = false
	m.BeforeUpdate()
}

// Show shows this mention to the mentioned user
func (m *Mention) Show() {
	if !m.IsBlocked {
		m.IsVisible = true
		m.BeforeUpdate()
	}
}

// Deactivate deactivates the mention (soft disable)
func (m *Mention) Deactivate() {
	m.IsActive = false
	m.BeforeUpdate()
}

// Activate activates the mention
func (m *Mention) Activate() {
	m.IsActive = true
	m.BeforeUpdate()
}

// CanViewMention checks if a user can view this mention
func (m *Mention) CanViewMention(currentUserID primitive.ObjectID) bool {
	// Mentioner and mentioned user can view
	if m.MentionerID == currentUserID || m.MentionedID == currentUserID {
		return true
	}

	// Check if mention is visible and active
	return m.IsVisible && m.IsActive && !m.IsBlocked
}

// CanDeleteMention checks if a user can delete this mention
func (m *Mention) CanDeleteMention(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// Mentioner can delete their mention
	if m.MentionerID == currentUserID {
		return true
	}

	// Mentioned user can delete mentions of them
	if m.MentionedID == currentUserID {
		return true
	}

	// Moderators and admins can delete any mention
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// CanEditMention checks if a user can edit this mention
func (m *Mention) CanEditMention(currentUserID primitive.ObjectID) bool {
	// Only the mentioned user can edit their mention settings
	return m.MentionedID == currentUserID
}

// IsFromGroup checks if the mention is from group content
func (m *Mention) IsFromGroup() bool {
	return m.GroupID != nil
}

// IsFromEvent checks if the mention is from event content
func (m *Mention) IsFromEvent() bool {
	return m.EventID != nil
}

// IsReply checks if the mention is in a reply/nested content
func (m *Mention) IsReply() bool {
	return m.ParentContentID != nil
}

// GetMentionContext returns context information about the mention
func (m *Mention) GetMentionContext() map[string]interface{} {
	context := make(map[string]interface{})

	context["content_type"] = m.ContentType
	context["is_group_mention"] = m.IsFromGroup()
	context["is_event_mention"] = m.IsFromEvent()
	context["is_reply_mention"] = m.IsReply()
	context["has_position"] = m.StartPosition > 0 || m.EndPosition > 0

	return context
}

// GetReadStatus returns detailed read status information
func (m *Mention) GetReadStatus() map[string]interface{} {
	status := make(map[string]interface{})

	status["is_read"] = m.IsRead
	status["is_notified"] = m.IsNotified
	status["read_at"] = m.ReadAt
	status["notified_at"] = m.NotifiedAt

	if m.ReadAt != nil && m.NotifiedAt != nil {
		status["time_to_read"] = m.ReadAt.Sub(*m.NotifiedAt).Minutes()
	}

	return status
}

// Utility functions for mentions

// ExtractMentionsFromText extracts user mentions from text content
func ExtractMentionsFromText(text string) []string {
	var mentions []string
	words := strings.Fields(text)

	for _, word := range words {
		if strings.HasPrefix(word, "@") && len(word) > 1 {
			// Clean the mention
			mention := strings.TrimPrefix(word, "@")
			// Remove trailing punctuation
			mention = strings.TrimRight(mention, ".,!?;:")
			if len(mention) > 0 {
				mentions = append(mentions, mention)
			}
		}
	}

	return mentions
}

// ExtractMentionsWithPositions extracts mentions with their positions in text
func ExtractMentionsWithPositions(text string) []MentionPosition {
	var mentions []MentionPosition

	// Simple implementation - in practice would use regex
	for i := 0; i < len(text)-1; i++ {
		if text[i] == '@' {
			start := i
			end := i + 1

			// Find end of mention
			for end < len(text) && (isAlphanumeric(text[end]) || text[end] == '_') {
				end++
			}

			if end > start+1 {
				mentionText := text[start:end]
				username := mentionText[1:] // Remove @

				mentions = append(mentions, MentionPosition{
					Username:      username,
					MentionText:   mentionText,
					StartPosition: start,
					EndPosition:   end,
				})
			}
		}
	}

	return mentions
}

// MentionPosition represents a mention with its position in text
type MentionPosition struct {
	Username      string `json:"username"`
	MentionText   string `json:"mention_text"`
	StartPosition int    `json:"start_position"`
	EndPosition   int    `json:"end_position"`
}

// isAlphanumeric checks if a character is alphanumeric
func isAlphanumeric(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

// IsValidMention checks if a mention string is valid
func IsValidMention(mention string) bool {
	// Remove @ if present
	mention = strings.TrimPrefix(mention, "@")

	// Check length
	if len(mention) < 1 || len(mention) > 50 {
		return false
	}

	// Check for invalid characters (simplified)
	// In practice, would validate against username rules
	return len(strings.TrimSpace(mention)) > 0
}

// ValidateMentionText validates mention text format
func ValidateMentionText(text string) bool {
	if !strings.HasPrefix(text, "@") {
		return false
	}

	username := strings.TrimPrefix(text, "@")
	return IsValidMention(username)
}

// GetMentionContentTypes returns valid content types that can have mentions
func GetMentionContentTypes() []string {
	return []string{"post", "comment", "story", "message"}
}

// IsValidMentionContentType checks if content type can have mentions
func IsValidMentionContentType(contentType string) bool {
	validTypes := GetMentionContentTypes()
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// CreateMentionsFromText creates mention objects from text content
func CreateMentionsFromText(text string, mentionerID primitive.ObjectID, contentType string, contentID primitive.ObjectID) []CreateMentionRequest {
	var mentions []CreateMentionRequest

	mentionPositions := ExtractMentionsWithPositions(text)

	for _, pos := range mentionPositions {
		mentions = append(mentions, CreateMentionRequest{
			MentionedID:   pos.Username, // This would need to be resolved to user ID
			ContentType:   contentType,
			ContentID:     contentID.Hex(),
			StartPosition: pos.StartPosition,
			EndPosition:   pos.EndPosition,
			MentionText:   pos.MentionText,
		})
	}

	return mentions
}
