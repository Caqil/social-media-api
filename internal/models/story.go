// models/story.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Story represents a 24-hour temporary post (like Instagram/Snapchat stories)
type Story struct {
	BaseModel `bson:",inline"`

	// Author Information
	UserID primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	Author UserResponse       `json:"author,omitempty" bson:"-"` // Populated when querying

	// Content
	Content     string      `json:"content,omitempty" bson:"content,omitempty" validate:"max=2000"`
	ContentType ContentType `json:"content_type" bson:"content_type" validate:"required"`
	Media       MediaInfo   `json:"media" bson:"media" validate:"required"` // Stories always have media

	// Story-specific properties
	Duration  int       `json:"duration" bson:"duration"`     // Duration in seconds (default: 15)
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"` // 24 hours from creation
	IsExpired bool      `json:"is_expired" bson:"is_expired"`

	// Privacy and visibility
	Visibility     PrivacyLevel         `json:"visibility" bson:"visibility"`
	AllowedViewers []primitive.ObjectID `json:"allowed_viewers,omitempty" bson:"allowed_viewers,omitempty"` // For custom audience
	BlockedViewers []primitive.ObjectID `json:"blocked_viewers,omitempty" bson:"blocked_viewers,omitempty"` // Users who can't see this story

	// Engagement Statistics
	ViewsCount   int64 `json:"views_count" bson:"views_count"`
	LikesCount   int64 `json:"likes_count" bson:"likes_count"`
	RepliesCount int64 `json:"replies_count" bson:"replies_count"`
	SharesCount  int64 `json:"shares_count" bson:"shares_count"`

	// Story interactions
	AllowReplies    bool `json:"allow_replies" bson:"allow_replies"`
	AllowReactions  bool `json:"allow_reactions" bson:"allow_reactions"`
	AllowSharing    bool `json:"allow_sharing" bson:"allow_sharing"`
	AllowScreenshot bool `json:"allow_screenshot" bson:"allow_screenshot"`

	// Story features
	BackgroundColor string         `json:"background_color,omitempty" bson:"background_color,omitempty"`
	TextColor       string         `json:"text_color,omitempty" bson:"text_color,omitempty"`
	FontFamily      string         `json:"font_family,omitempty" bson:"font_family,omitempty"`
	Stickers        []StorySticker `json:"stickers,omitempty" bson:"stickers,omitempty"`
	Mentions        []StoryMention `json:"mentions,omitempty" bson:"mentions,omitempty"`
	Hashtags        []StoryHashtag `json:"hashtags,omitempty" bson:"hashtags,omitempty"`
	Location        *Location      `json:"location,omitempty" bson:"location,omitempty"`
	Music           *StoryMusic    `json:"music,omitempty" bson:"music,omitempty"`

	// Highlights (permanent stories)
	IsHighlighted bool                `json:"is_highlighted" bson:"is_highlighted"`
	HighlightID   *primitive.ObjectID `json:"highlight_id,omitempty" bson:"highlight_id,omitempty"`

	// Analytics
	UniqueViewsCount    int64   `json:"unique_views_count" bson:"unique_views_count"`
	AverageViewDuration float64 `json:"average_view_duration" bson:"average_view_duration"`
	CompletionRate      float64 `json:"completion_rate" bson:"completion_rate"`
	EngagementRate      float64 `json:"engagement_rate" bson:"engagement_rate"`

	// Content moderation
	IsReported     bool   `json:"is_reported" bson:"is_reported"`
	ReportsCount   int64  `json:"reports_count" bson:"reports_count"`
	IsHidden       bool   `json:"is_hidden" bson:"is_hidden"`
	ModerationNote string `json:"moderation_note,omitempty" bson:"moderation_note,omitempty"`
}

// StorySticker represents a sticker placed on a story
type StorySticker struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type     string             `json:"type" bson:"type"` // emoji, gif, poll, question, location, time
	Content  string             `json:"content" bson:"content"`
	X        float64            `json:"x" bson:"x"`               // Position X (0-1)
	Y        float64            `json:"y" bson:"y"`               // Position Y (0-1)
	Width    float64            `json:"width" bson:"width"`       // Width (0-1)
	Height   float64            `json:"height" bson:"height"`     // Height (0-1)
	Rotation float64            `json:"rotation" bson:"rotation"` // Rotation in degrees
	Scale    float64            `json:"scale" bson:"scale"`       // Scale factor

	// Interactive sticker data
	PollOptions     []string             `json:"poll_options,omitempty" bson:"poll_options,omitempty"`
	PollVotes       map[string]int64     `json:"poll_votes,omitempty" bson:"poll_votes,omitempty"`
	QuestionText    string               `json:"question_text,omitempty" bson:"question_text,omitempty"`
	QuestionReplies []StoryQuestionReply `json:"question_replies,omitempty" bson:"question_replies,omitempty"`
}

// StoryMention represents a user mention in a story
type StoryMention struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	User     UserResponse       `json:"user,omitempty" bson:"-"` // Populated when querying
	Username string             `json:"username" bson:"username"`
	X        float64            `json:"x" bson:"x"` // Position X (0-1)
	Y        float64            `json:"y" bson:"y"` // Position Y (0-1)
	Width    float64            `json:"width" bson:"width"`
	Height   float64            `json:"height" bson:"height"`
}

// StoryHashtag represents a hashtag in a story
type StoryHashtag struct {
	Tag    string  `json:"tag" bson:"tag"`
	X      float64 `json:"x" bson:"x"`
	Y      float64 `json:"y" bson:"y"`
	Width  float64 `json:"width" bson:"width"`
	Height float64 `json:"height" bson:"height"`
}

// StoryMusic represents music added to a story
type StoryMusic struct {
	Title      string `json:"title" bson:"title"`
	Artist     string `json:"artist" bson:"artist"`
	PreviewURL string `json:"preview_url" bson:"preview_url"`
	StartTime  int    `json:"start_time" bson:"start_time"`   // Start time in seconds
	Duration   int    `json:"duration" bson:"duration"`       // Duration in seconds
	ExternalID string `json:"external_id" bson:"external_id"` // Spotify/Apple Music ID
}

// StoryQuestionReply represents a reply to a question sticker
type StoryQuestionReply struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	User      UserResponse       `json:"user,omitempty" bson:"-"`
	Reply     string             `json:"reply" bson:"reply"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// StoryView represents a view of a story by a user
type StoryView struct {
	BaseModel `bson:",inline"`

	StoryID primitive.ObjectID `json:"story_id" bson:"story_id" validate:"required"`
	UserID  primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	User    UserResponse       `json:"user,omitempty" bson:"-"` // Populated when querying

	// View details
	ViewDuration float64 `json:"view_duration" bson:"view_duration"` // Duration in seconds
	WatchedFully bool    `json:"watched_fully" bson:"watched_fully"`
	Source       string  `json:"source" bson:"source"`           // feed, profile, search, direct
	DeviceType   string  `json:"device_type" bson:"device_type"` // mobile, desktop, tablet

	// Interaction during view
	Liked      bool `json:"liked" bson:"liked"`
	Replied    bool `json:"replied" bson:"replied"`
	Shared     bool `json:"shared" bson:"shared"`
	Screenshot bool `json:"screenshot" bson:"screenshot"`

	// Location and metadata
	IPAddress string `json:"-" bson:"ip_address,omitempty"`
	UserAgent string `json:"-" bson:"user_agent,omitempty"`
}

// StoryHighlight represents a collection of highlighted stories
type StoryHighlight struct {
	BaseModel `bson:",inline"`

	UserID       primitive.ObjectID   `json:"user_id" bson:"user_id" validate:"required"`
	Title        string               `json:"title" bson:"title" validate:"required,max=50"`
	CoverImage   string               `json:"cover_image" bson:"cover_image"`
	StoryIDs     []primitive.ObjectID `json:"story_ids" bson:"story_ids"`
	StoriesCount int64                `json:"stories_count" bson:"stories_count"`
	IsActive     bool                 `json:"is_active" bson:"is_active"`
	Order        int                  `json:"order" bson:"order"` // Display order
}

// Story Response Models

// StoryResponse represents the story data returned in API responses
type StoryResponse struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Author          UserResponse   `json:"author"`
	Content         string         `json:"content,omitempty"`
	ContentType     ContentType    `json:"content_type"`
	Media           MediaInfo      `json:"media"`
	Duration        int            `json:"duration"`
	ExpiresAt       time.Time      `json:"expires_at"`
	IsExpired       bool           `json:"is_expired"`
	Visibility      PrivacyLevel   `json:"visibility"`
	ViewsCount      int64          `json:"views_count"`
	LikesCount      int64          `json:"likes_count"`
	RepliesCount    int64          `json:"replies_count"`
	SharesCount     int64          `json:"shares_count"`
	AllowReplies    bool           `json:"allow_replies"`
	AllowReactions  bool           `json:"allow_reactions"`
	AllowSharing    bool           `json:"allow_sharing"`
	AllowScreenshot bool           `json:"allow_screenshot"`
	BackgroundColor string         `json:"background_color,omitempty"`
	TextColor       string         `json:"text_color,omitempty"`
	FontFamily      string         `json:"font_family,omitempty"`
	Stickers        []StorySticker `json:"stickers,omitempty"`
	Mentions        []StoryMention `json:"mentions,omitempty"`
	Hashtags        []StoryHashtag `json:"hashtags,omitempty"`
	Location        *Location      `json:"location,omitempty"`
	Music           *StoryMusic    `json:"music,omitempty"`
	IsHighlighted   bool           `json:"is_highlighted"`
	HighlightID     string         `json:"highlight_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	TimeAgo         string         `json:"time_ago,omitempty"`

	// User-specific context
	HasViewed    bool         `json:"has_viewed,omitempty"`
	UserReaction ReactionType `json:"user_reaction,omitempty"`
	CanView      bool         `json:"can_view,omitempty"`
}

// StoryViewResponse represents story view data
type StoryViewResponse struct {
	ID           string       `json:"id"`
	StoryID      string       `json:"story_id"`
	UserID       string       `json:"user_id"`
	User         UserResponse `json:"user"`
	ViewDuration float64      `json:"view_duration"`
	WatchedFully bool         `json:"watched_fully"`
	CreatedAt    time.Time    `json:"created_at"`
	TimeAgo      string       `json:"time_ago,omitempty"`
}

// StoryHighlightResponse represents story highlight data
type StoryHighlightResponse struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Title        string          `json:"title"`
	CoverImage   string          `json:"cover_image"`
	StoriesCount int64           `json:"stories_count"`
	IsActive     bool            `json:"is_active"`
	Order        int             `json:"order"`
	Stories      []StoryResponse `json:"stories,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Request Models

// CreateStoryRequest represents the request to create a story
type CreateStoryRequest struct {
	Content         string         `json:"content,omitempty" validate:"max=2000"`
	ContentType     ContentType    `json:"content_type" validate:"required,oneof=image video"`
	Media           MediaInfo      `json:"media" validate:"required"`
	Duration        int            `json:"duration,omitempty" validate:"min=1,max=30"`
	Visibility      PrivacyLevel   `json:"visibility" validate:"required,oneof=public friends private"`
	AllowedViewers  []string       `json:"allowed_viewers,omitempty"`
	BlockedViewers  []string       `json:"blocked_viewers,omitempty"`
	AllowReplies    bool           `json:"allow_replies"`
	AllowReactions  bool           `json:"allow_reactions"`
	AllowSharing    bool           `json:"allow_sharing"`
	AllowScreenshot bool           `json:"allow_screenshot"`
	BackgroundColor string         `json:"background_color,omitempty"`
	TextColor       string         `json:"text_color,omitempty"`
	FontFamily      string         `json:"font_family,omitempty"`
	Stickers        []StorySticker `json:"stickers,omitempty"`
	Mentions        []StoryMention `json:"mentions,omitempty"`
	Hashtags        []StoryHashtag `json:"hashtags,omitempty"`
	Location        *Location      `json:"location,omitempty"`
	Music           *StoryMusic    `json:"music,omitempty"`
}

// CreateStoryHighlightRequest represents the request to create a story highlight
type CreateStoryHighlightRequest struct {
	Title      string   `json:"title" validate:"required,max=50"`
	CoverImage string   `json:"cover_image,omitempty"`
	StoryIDs   []string `json:"story_ids" validate:"required,min=1"`
}

// UpdateStoryHighlightRequest represents the request to update a story highlight
type UpdateStoryHighlightRequest struct {
	Title      *string  `json:"title,omitempty" validate:"omitempty,max=50"`
	CoverImage *string  `json:"cover_image,omitempty"`
	StoryIDs   []string `json:"story_ids,omitempty"`
	IsActive   *bool    `json:"is_active,omitempty"`
	Order      *int     `json:"order,omitempty"`
}

// Methods for Story model

// BeforeCreate sets default values before creating story
func (s *Story) BeforeCreate() {
	s.BaseModel.BeforeCreate()

	// Set default values
	s.ViewsCount = 0
	s.LikesCount = 0
	s.RepliesCount = 0
	s.SharesCount = 0
	s.UniqueViewsCount = 0
	s.AverageViewDuration = 0.0
	s.CompletionRate = 0.0
	s.EngagementRate = 0.0
	s.ReportsCount = 0
	s.IsExpired = false
	s.IsHighlighted = false
	s.IsReported = false
	s.IsHidden = false

	// Set default duration
	if s.Duration == 0 {
		if s.ContentType == ContentTypeImage {
			s.Duration = 5 // 5 seconds for images
		} else {
			s.Duration = 15 // 15 seconds for videos
		}
	}

	// Set default permissions
	s.AllowReplies = true
	s.AllowReactions = true
	s.AllowSharing = true
	s.AllowScreenshot = true

	// Set default visibility
	if s.Visibility == "" {
		s.Visibility = PrivacyPublic
	}

	// Set expiration time (24 hours from now)
	s.ExpiresAt = s.CreatedAt.Add(24 * time.Hour)
}

// ToStoryResponse converts Story model to StoryResponse
func (s *Story) ToStoryResponse() StoryResponse {
	response := StoryResponse{
		ID:              s.ID.Hex(),
		UserID:          s.UserID.Hex(),
		Content:         s.Content,
		ContentType:     s.ContentType,
		Media:           s.Media,
		Duration:        s.Duration,
		ExpiresAt:       s.ExpiresAt,
		IsExpired:       s.IsExpired,
		Visibility:      s.Visibility,
		ViewsCount:      s.ViewsCount,
		LikesCount:      s.LikesCount,
		RepliesCount:    s.RepliesCount,
		SharesCount:     s.SharesCount,
		AllowReplies:    s.AllowReplies,
		AllowReactions:  s.AllowReactions,
		AllowSharing:    s.AllowSharing,
		AllowScreenshot: s.AllowScreenshot,
		BackgroundColor: s.BackgroundColor,
		TextColor:       s.TextColor,
		FontFamily:      s.FontFamily,
		Stickers:        s.Stickers,
		Mentions:        s.Mentions,
		Hashtags:        s.Hashtags,
		Location:        s.Location,
		Music:           s.Music,
		IsHighlighted:   s.IsHighlighted,
		CreatedAt:       s.CreatedAt,
	}

	if s.HighlightID != nil {
		response.HighlightID = s.HighlightID.Hex()
	}

	return response
}

// CheckExpiration checks and updates expiration status
func (s *Story) CheckExpiration() {
	if !s.IsExpired && time.Now().After(s.ExpiresAt) {
		s.IsExpired = true
		s.BeforeUpdate()
	}
}

// IncrementViewsCount increments the views count
func (s *Story) IncrementViewsCount() {
	s.ViewsCount++
	s.BeforeUpdate()
}

// IncrementLikesCount increments the likes count
func (s *Story) IncrementLikesCount() {
	s.LikesCount++
	s.UpdateEngagementRate()
	s.BeforeUpdate()
}

// DecrementLikesCount decrements the likes count
func (s *Story) DecrementLikesCount() {
	if s.LikesCount > 0 {
		s.LikesCount--
	}
	s.UpdateEngagementRate()
	s.BeforeUpdate()
}

// IncrementRepliesCount increments the replies count
func (s *Story) IncrementRepliesCount() {
	s.RepliesCount++
	s.UpdateEngagementRate()
	s.BeforeUpdate()
}

// IncrementSharesCount increments the shares count
func (s *Story) IncrementSharesCount() {
	s.SharesCount++
	s.UpdateEngagementRate()
	s.BeforeUpdate()
}

// UpdateEngagementRate calculates and updates engagement rate
func (s *Story) UpdateEngagementRate() {
	if s.ViewsCount == 0 {
		s.EngagementRate = 0.0
		return
	}

	totalEngagements := s.LikesCount + s.RepliesCount + s.SharesCount
	s.EngagementRate = (float64(totalEngagements) / float64(s.ViewsCount)) * 100
}

// CanViewStory checks if a user can view this story
func (s *Story) CanViewStory(currentUserID primitive.ObjectID, isFollowing bool, isAuthor bool) bool {
	// Author can always view their own story
	if isAuthor {
		return true
	}

	// Check if story is expired and not highlighted
	if s.IsExpired && !s.IsHighlighted {
		return false
	}

	// Check if story is hidden
	if s.IsHidden {
		return false
	}

	// Check if user is blocked
	for _, blockedID := range s.BlockedViewers {
		if blockedID == currentUserID {
			return false
		}
	}

	// Check visibility settings
	switch s.Visibility {
	case PrivacyPublic:
		return true
	case PrivacyFriends:
		return isFollowing
	case PrivacyPrivate:
		// Check if user is in allowed viewers list
		for _, allowedID := range s.AllowedViewers {
			if allowedID == currentUserID {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// CanDeleteStory checks if a user can delete this story
func (s *Story) CanDeleteStory(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// Author can delete their own story
	if s.UserID == currentUserID {
		return true
	}

	// Moderators and admins can delete any story
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// AddToHighlight adds the story to a highlight
func (s *Story) AddToHighlight(highlightID primitive.ObjectID) {
	s.IsHighlighted = true
	s.HighlightID = &highlightID
	s.BeforeUpdate()
}

// RemoveFromHighlight removes the story from highlight
func (s *Story) RemoveFromHighlight() {
	s.IsHighlighted = false
	s.HighlightID = nil
	s.BeforeUpdate()
}

// Methods for StoryView model

// BeforeCreate sets default values before creating story view
func (sv *StoryView) BeforeCreate() {
	sv.BaseModel.BeforeCreate()

	// Set default values
	sv.WatchedFully = false
	sv.Liked = false
	sv.Replied = false
	sv.Shared = false
	sv.Screenshot = false

	// Set default view duration based on story duration
	if sv.ViewDuration == 0 {
		sv.ViewDuration = 1.0 // Default 1 second view
	}
}

// ToStoryViewResponse converts StoryView to StoryViewResponse
func (sv *StoryView) ToStoryViewResponse() StoryViewResponse {
	return StoryViewResponse{
		ID:           sv.ID.Hex(),
		StoryID:      sv.StoryID.Hex(),
		UserID:       sv.UserID.Hex(),
		ViewDuration: sv.ViewDuration,
		WatchedFully: sv.WatchedFully,
		CreatedAt:    sv.CreatedAt,
	}
}

// Methods for StoryHighlight model

// BeforeCreate sets default values before creating story highlight
func (sh *StoryHighlight) BeforeCreate() {
	sh.BaseModel.BeforeCreate()

	// Set default values
	sh.IsActive = true
	sh.StoriesCount = int64(len(sh.StoryIDs))

	// Set default order
	if sh.Order == 0 {
		sh.Order = 1
	}
}

// ToStoryHighlightResponse converts StoryHighlight to StoryHighlightResponse
func (sh *StoryHighlight) ToStoryHighlightResponse() StoryHighlightResponse {
	return StoryHighlightResponse{
		ID:           sh.ID.Hex(),
		UserID:       sh.UserID.Hex(),
		Title:        sh.Title,
		CoverImage:   sh.CoverImage,
		StoriesCount: sh.StoriesCount,
		IsActive:     sh.IsActive,
		Order:        sh.Order,
		CreatedAt:    sh.CreatedAt,
		UpdatedAt:    sh.UpdatedAt,
	}
}

// AddStory adds a story to the highlight
func (sh *StoryHighlight) AddStory(storyID primitive.ObjectID) {
	// Check if story already exists
	for _, existingID := range sh.StoryIDs {
		if existingID == storyID {
			return
		}
	}

	sh.StoryIDs = append(sh.StoryIDs, storyID)
	sh.StoriesCount = int64(len(sh.StoryIDs))
	sh.BeforeUpdate()
}

// RemoveStory removes a story from the highlight
func (sh *StoryHighlight) RemoveStory(storyID primitive.ObjectID) {
	for i, existingID := range sh.StoryIDs {
		if existingID == storyID {
			sh.StoryIDs = append(sh.StoryIDs[:i], sh.StoryIDs[i+1:]...)
			sh.StoriesCount = int64(len(sh.StoryIDs))
			sh.BeforeUpdate()
			return
		}
	}
}

// CanEditHighlight checks if a user can edit this highlight
func (sh *StoryHighlight) CanEditHighlight(currentUserID primitive.ObjectID) bool {
	return sh.UserID == currentUserID && !sh.IsDeleted()
}

// Utility functions

// GetStoryDuration returns default duration based on content type
func GetStoryDuration(contentType ContentType) int {
	switch contentType {
	case ContentTypeImage:
		return 5
	case ContentTypeVideo:
		return 15
	default:
		return 10
	}
}

// IsValidStoryContentType checks if content type is valid for stories
func IsValidStoryContentType(contentType ContentType) bool {
	validTypes := []ContentType{ContentTypeImage, ContentTypeVideo}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}
