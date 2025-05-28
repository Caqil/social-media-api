// models/post.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a post in the social media platform
type Post struct {
	BaseModel `bson:",inline"`

	// Author Information
	UserID primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	Author UserResponse       `json:"author,omitempty" bson:"-"` // Populated when querying

	// Content
	Content     string      `json:"content" bson:"content" validate:"max=5000"`
	ContentType ContentType `json:"content_type" bson:"content_type"`
	Media       []MediaInfo `json:"media,omitempty" bson:"media,omitempty"`

	// Post Metadata
	Type       string       `json:"type" bson:"type"` // post, story, reel, poll
	Visibility PrivacyLevel `json:"visibility" bson:"visibility"`
	Language   string       `json:"language,omitempty" bson:"language,omitempty"`
	Location   *Location    `json:"location,omitempty" bson:"location,omitempty"`

	// Engagement Statistics
	LikesCount    int64 `json:"likes_count" bson:"likes_count"`
	CommentsCount int64 `json:"comments_count" bson:"comments_count"`
	SharesCount   int64 `json:"shares_count" bson:"shares_count"`
	ViewsCount    int64 `json:"views_count" bson:"views_count"`
	SavesCount    int64 `json:"saves_count" bson:"saves_count"`

	// Social Features
	Hashtags     []string             `json:"hashtags,omitempty" bson:"hashtags,omitempty"`
	Mentions     []primitive.ObjectID `json:"mentions,omitempty" bson:"mentions,omitempty"`
	MentionUsers []UserResponse       `json:"mention_users,omitempty" bson:"-"` // Populated when querying

	// Post Options
	IsEdited        bool       `json:"is_edited" bson:"is_edited"`
	EditedAt        *time.Time `json:"edited_at,omitempty" bson:"edited_at,omitempty"`
	CommentsEnabled bool       `json:"comments_enabled" bson:"comments_enabled"`
	LikesEnabled    bool       `json:"likes_enabled" bson:"likes_enabled"`
	SharesEnabled   bool       `json:"shares_enabled" bson:"shares_enabled"`
	IsPinned        bool       `json:"is_pinned" bson:"is_pinned"`
	IsPromoted      bool       `json:"is_promoted" bson:"is_promoted"`

	// Content Moderation
	IsReported     bool   `json:"is_reported" bson:"is_reported"`
	ReportsCount   int64  `json:"reports_count" bson:"reports_count"`
	IsHidden       bool   `json:"is_hidden" bson:"is_hidden"`
	IsApproved     bool   `json:"is_approved" bson:"is_approved"`
	ModerationNote string `json:"moderation_note,omitempty" bson:"moderation_note,omitempty"`

	// Sharing and Reposting
	OriginalPostID *primitive.ObjectID `json:"original_post_id,omitempty" bson:"original_post_id,omitempty"`
	OriginalPost   *PostResponse       `json:"original_post,omitempty" bson:"-"` // Populated when querying
	IsRepost       bool                `json:"is_repost" bson:"is_repost"`
	RepostComment  string              `json:"repost_comment,omitempty" bson:"repost_comment,omitempty"`

	// Group/Event Association
	GroupID *primitive.ObjectID `json:"group_id,omitempty" bson:"group_id,omitempty"`
	EventID *primitive.ObjectID `json:"event_id,omitempty" bson:"event_id,omitempty"`

	// Scheduled Posts
	IsScheduled  bool       `json:"is_scheduled" bson:"is_scheduled"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty" bson:"scheduled_for,omitempty"`
	IsPublished  bool       `json:"is_published" bson:"is_published"`
	PublishedAt  *time.Time `json:"published_at,omitempty" bson:"published_at,omitempty"`

	// Poll Data (if post type is poll)
	PollOptions   []PollOption `json:"poll_options,omitempty" bson:"poll_options,omitempty"`
	PollExpiresAt *time.Time   `json:"poll_expires_at,omitempty" bson:"poll_expires_at,omitempty"`
	PollMultiple  bool         `json:"poll_multiple,omitempty" bson:"poll_multiple,omitempty"`
	TotalVotes    int64        `json:"total_votes,omitempty" bson:"total_votes,omitempty"`

	// Analytics
	EngagementRate  float64 `json:"engagement_rate" bson:"engagement_rate"`
	ReachCount      int64   `json:"reach_count" bson:"reach_count"`
	ImpressionCount int64   `json:"impression_count" bson:"impression_count"`

	// Additional Metadata
	Source       string                 `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	IPAddress    string                 `json:"-" bson:"ip_address,omitempty"`
	UserAgent    string                 `json:"-" bson:"user_agent,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
}

// PollOption represents an option in a poll post
type PollOption struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Text       string             `json:"text" bson:"text" validate:"required,max=100"`
	VotesCount int64              `json:"votes_count" bson:"votes_count"`
	Percentage float64            `json:"percentage" bson:"percentage"`
}

// PostResponse represents the post data returned in API responses
type PostResponse struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Author          UserResponse   `json:"author"`
	Content         string         `json:"content"`
	ContentType     ContentType    `json:"content_type"`
	Media           []MediaInfo    `json:"media,omitempty"`
	Type            string         `json:"type"`
	Visibility      PrivacyLevel   `json:"visibility"`
	Language        string         `json:"language,omitempty"`
	Location        *Location      `json:"location,omitempty"`
	LikesCount      int64          `json:"likes_count"`
	CommentsCount   int64          `json:"comments_count"`
	SharesCount     int64          `json:"shares_count"`
	ViewsCount      int64          `json:"views_count"`
	SavesCount      int64          `json:"saves_count"`
	Hashtags        []string       `json:"hashtags,omitempty"`
	Mentions        []string       `json:"mentions,omitempty"` // User IDs as strings
	MentionUsers    []UserResponse `json:"mention_users,omitempty"`
	IsEdited        bool           `json:"is_edited"`
	EditedAt        *time.Time     `json:"edited_at,omitempty"`
	CommentsEnabled bool           `json:"comments_enabled"`
	LikesEnabled    bool           `json:"likes_enabled"`
	SharesEnabled   bool           `json:"shares_enabled"`
	IsPinned        bool           `json:"is_pinned"`
	IsRepost        bool           `json:"is_repost"`
	RepostComment   string         `json:"repost_comment,omitempty"`
	OriginalPost    *PostResponse  `json:"original_post,omitempty"`
	GroupID         string         `json:"group_id,omitempty"`
	EventID         string         `json:"event_id,omitempty"`
	IsScheduled     bool           `json:"is_scheduled"`
	ScheduledFor    *time.Time     `json:"scheduled_for,omitempty"`
	PublishedAt     *time.Time     `json:"published_at,omitempty"`
	PollOptions     []PollOption   `json:"poll_options,omitempty"`
	PollExpiresAt   *time.Time     `json:"poll_expires_at,omitempty"`
	TotalVotes      int64          `json:"total_votes,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// User-specific context (set based on current user)
	IsLiked       bool         `json:"is_liked,omitempty"`
	IsSaved       bool         `json:"is_saved,omitempty"`
	UserReaction  ReactionType `json:"user_reaction,omitempty"`
	UserPollVotes []string     `json:"user_poll_votes,omitempty"` // Poll option IDs user voted for
}

// CreatePostRequest represents the request to create a new post
type CreatePostRequest struct {
	Content         string                 `json:"content" validate:"max=5000"`
	ContentType     ContentType            `json:"content_type" validate:"required,oneof=text image video link gif poll"`
	Media           []MediaInfo            `json:"media,omitempty"`
	Type            string                 `json:"type" validate:"oneof=post story reel poll"`
	Visibility      PrivacyLevel           `json:"visibility" validate:"required,oneof=public friends private"`
	Language        string                 `json:"language,omitempty"`
	Location        *Location              `json:"location,omitempty"`
	Hashtags        []string               `json:"hashtags,omitempty"`
	Mentions        []string               `json:"mentions,omitempty"` // User IDs as strings
	CommentsEnabled bool                   `json:"comments_enabled"`
	LikesEnabled    bool                   `json:"likes_enabled"`
	SharesEnabled   bool                   `json:"shares_enabled"`
	GroupID         string                 `json:"group_id,omitempty"`
	EventID         string                 `json:"event_id,omitempty"`
	ScheduledFor    *time.Time             `json:"scheduled_for,omitempty"`
	PollOptions     []CreatePollOption     `json:"poll_options,omitempty"`
	PollExpiresAt   *time.Time             `json:"poll_expires_at,omitempty"`
	PollMultiple    bool                   `json:"poll_multiple,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty"`
}

// CreatePollOption represents a poll option in create request
type CreatePollOption struct {
	Text string `json:"text" validate:"required,max=100"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Content         *string       `json:"content,omitempty" validate:"omitempty,max=5000"`
	Visibility      *PrivacyLevel `json:"visibility,omitempty" validate:"omitempty,oneof=public friends private"`
	Language        *string       `json:"language,omitempty"`
	Location        *Location     `json:"location,omitempty"`
	Hashtags        []string      `json:"hashtags,omitempty"`
	Mentions        []string      `json:"mentions,omitempty"`
	CommentsEnabled *bool         `json:"comments_enabled,omitempty"`
	LikesEnabled    *bool         `json:"likes_enabled,omitempty"`
	SharesEnabled   *bool         `json:"shares_enabled,omitempty"`
	IsPinned        *bool         `json:"is_pinned,omitempty"`
}

// RepostRequest represents the request to repost/share a post
type RepostRequest struct {
	PostID     string       `json:"post_id" validate:"required"`
	Comment    string       `json:"comment,omitempty" validate:"max=500"`
	Visibility PrivacyLevel `json:"visibility" validate:"required,oneof=public friends private"`
	GroupID    string       `json:"group_id,omitempty"`
}

// PostStatsResponse represents post statistics
type PostStatsResponse struct {
	PostID          string  `json:"post_id"`
	LikesCount      int64   `json:"likes_count"`
	CommentsCount   int64   `json:"comments_count"`
	SharesCount     int64   `json:"shares_count"`
	ViewsCount      int64   `json:"views_count"`
	SavesCount      int64   `json:"saves_count"`
	EngagementRate  float64 `json:"engagement_rate"`
	ReachCount      int64   `json:"reach_count"`
	ImpressionCount int64   `json:"impression_count"`
}

// PostFeedResponse represents posts in feed with additional context
type PostFeedResponse struct {
	PostResponse     `json:",inline"`
	TimeAgo          string  `json:"time_ago"`
	IsSponsored      bool    `json:"is_sponsored,omitempty"`
	InteractionScore float64 `json:"-"` // Used for feed ranking, not exposed to client
}

// Methods for Post model

// BeforeCreate sets default values before creating post
func (p *Post) BeforeCreate() {
	p.BaseModel.BeforeCreate()

	// Set default values
	p.LikesCount = 0
	p.CommentsCount = 0
	p.SharesCount = 0
	p.ViewsCount = 0
	p.SavesCount = 0
	p.ReportsCount = 0
	p.TotalVotes = 0
	p.EngagementRate = 0.0
	p.ReachCount = 0
	p.ImpressionCount = 0

	// Set default permissions
	p.CommentsEnabled = true
	p.LikesEnabled = true
	p.SharesEnabled = true

	// Set default content type
	if p.ContentType == "" {
		p.ContentType = ContentTypeText
	}

	// Set default type
	if p.Type == "" {
		p.Type = "post"
	}

	// Set default visibility
	if p.Visibility == "" {
		p.Visibility = PrivacyPublic
	}

	// Set publication status
	if p.ScheduledFor == nil || p.ScheduledFor.Before(time.Now()) {
		p.IsPublished = true
		now := time.Now()
		p.PublishedAt = &now
	} else {
		p.IsScheduled = true
		p.IsPublished = false
	}

	// Set moderation status
	p.IsApproved = true // Auto-approve by default, can be changed based on moderation settings
	p.IsHidden = false
	p.IsReported = false

	// Initialize poll vote counts
	for i := range p.PollOptions {
		p.PollOptions[i].ID = primitive.NewObjectID()
		p.PollOptions[i].VotesCount = 0
		p.PollOptions[i].Percentage = 0.0
	}
}

// ToPostResponse converts Post model to PostResponse
func (p *Post) ToPostResponse() PostResponse {
	response := PostResponse{
		ID:              p.ID.Hex(),
		UserID:          p.UserID.Hex(),
		Content:         p.Content,
		ContentType:     p.ContentType,
		Media:           p.Media,
		Type:            p.Type,
		Visibility:      p.Visibility,
		Language:        p.Language,
		Location:        p.Location,
		LikesCount:      p.LikesCount,
		CommentsCount:   p.CommentsCount,
		SharesCount:     p.SharesCount,
		ViewsCount:      p.ViewsCount,
		SavesCount:      p.SavesCount,
		Hashtags:        p.Hashtags,
		IsEdited:        p.IsEdited,
		EditedAt:        p.EditedAt,
		CommentsEnabled: p.CommentsEnabled,
		LikesEnabled:    p.LikesEnabled,
		SharesEnabled:   p.SharesEnabled,
		IsPinned:        p.IsPinned,
		IsRepost:        p.IsRepost,
		RepostComment:   p.RepostComment,
		IsScheduled:     p.IsScheduled,
		ScheduledFor:    p.ScheduledFor,
		PublishedAt:     p.PublishedAt,
		PollOptions:     p.PollOptions,
		PollExpiresAt:   p.PollExpiresAt,
		TotalVotes:      p.TotalVotes,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}

	// Convert ObjectIDs to strings
	if len(p.Mentions) > 0 {
		response.Mentions = make([]string, len(p.Mentions))
		for i, mention := range p.Mentions {
			response.Mentions[i] = mention.Hex()
		}
	}

	if p.GroupID != nil {
		response.GroupID = p.GroupID.Hex()
	}

	if p.EventID != nil {
		response.EventID = p.EventID.Hex()
	}

	return response
}

// ToPostStatsResponse converts Post model to PostStatsResponse
func (p *Post) ToPostStatsResponse() PostStatsResponse {
	return PostStatsResponse{
		PostID:          p.ID.Hex(),
		LikesCount:      p.LikesCount,
		CommentsCount:   p.CommentsCount,
		SharesCount:     p.SharesCount,
		ViewsCount:      p.ViewsCount,
		SavesCount:      p.SavesCount,
		EngagementRate:  p.EngagementRate,
		ReachCount:      p.ReachCount,
		ImpressionCount: p.ImpressionCount,
	}
}

// IncrementLikesCount increments the likes count
func (p *Post) IncrementLikesCount() {
	p.LikesCount++
	p.UpdateEngagementRate()
	p.BeforeUpdate()
}

// DecrementLikesCount decrements the likes count
func (p *Post) DecrementLikesCount() {
	if p.LikesCount > 0 {
		p.LikesCount--
	}
	p.UpdateEngagementRate()
	p.BeforeUpdate()
}

// IncrementCommentsCount increments the comments count
func (p *Post) IncrementCommentsCount() {
	p.CommentsCount++
	p.UpdateEngagementRate()
	p.BeforeUpdate()
}

// DecrementCommentsCount decrements the comments count
func (p *Post) DecrementCommentsCount() {
	if p.CommentsCount > 0 {
		p.CommentsCount--
	}
	p.UpdateEngagementRate()
	p.BeforeUpdate()
}

// IncrementSharesCount increments the shares count
func (p *Post) IncrementSharesCount() {
	p.SharesCount++
	p.UpdateEngagementRate()
	p.BeforeUpdate()
}

// DecrementSharesCount decrements the shares count
func (p *Post) DecrementSharesCount() {
	if p.SharesCount > 0 {
		p.SharesCount--
	}
	p.UpdateEngagementRate()
	p.BeforeUpdate()
}

// IncrementViewsCount increments the views count
func (p *Post) IncrementViewsCount() {
	p.ViewsCount++
	p.BeforeUpdate()
}

// IncrementSavesCount increments the saves count
func (p *Post) IncrementSavesCount() {
	p.SavesCount++
	p.BeforeUpdate()
}

// DecrementSavesCount decrements the saves count
func (p *Post) DecrementSavesCount() {
	if p.SavesCount > 0 {
		p.SavesCount--
	}
	p.BeforeUpdate()
}

// UpdateEngagementRate calculates and updates the engagement rate
func (p *Post) UpdateEngagementRate() {
	if p.ViewsCount == 0 {
		p.EngagementRate = 0.0
		return
	}

	totalEngagements := p.LikesCount + p.CommentsCount + p.SharesCount
	p.EngagementRate = (float64(totalEngagements) / float64(p.ViewsCount)) * 100
}

// CanViewPost checks if a user can view this post based on privacy settings and relationships
func (p *Post) CanViewPost(currentUserID primitive.ObjectID, isFollowing bool, isAuthor bool) bool {
	// Author can always view their own post
	if isAuthor {
		return true
	}

	// Check if post is hidden or not approved
	if p.IsHidden || !p.IsApproved {
		return false
	}

	// Check if post is published (for scheduled posts)
	if !p.IsPublished {
		return false
	}

	// Check visibility settings
	switch p.Visibility {
	case PrivacyPublic:
		return true
	case PrivacyFriends:
		return isFollowing
	case PrivacyPrivate:
		return false
	default:
		return false
	}
}

// CanEditPost checks if a user can edit this post
func (p *Post) CanEditPost(currentUserID primitive.ObjectID) bool {
	return p.UserID == currentUserID && !p.IsDeleted()
}

// CanDeletePost checks if a user can delete this post
func (p *Post) CanDeletePost(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// Author can delete their own post
	if p.UserID == currentUserID {
		return true
	}

	// Moderators and admins can delete any post
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// IsExpiredPoll checks if the poll has expired
func (p *Post) IsExpiredPoll() bool {
	return p.ContentType == ContentTypePoll && p.PollExpiresAt != nil && p.PollExpiresAt.Before(time.Now())
}

// UpdatePollVotes updates poll option vote counts and percentages
func (p *Post) UpdatePollVotes() {
	if len(p.PollOptions) == 0 {
		return
	}

	// Calculate total votes
	var totalVotes int64 = 0
	for _, option := range p.PollOptions {
		totalVotes += option.VotesCount
	}
	p.TotalVotes = totalVotes

	// Calculate percentages
	for i := range p.PollOptions {
		if totalVotes > 0 {
			p.PollOptions[i].Percentage = (float64(p.PollOptions[i].VotesCount) / float64(totalVotes)) * 100
		} else {
			p.PollOptions[i].Percentage = 0.0
		}
	}

	p.BeforeUpdate()
}

// AddHashtag adds a hashtag to the post
func (p *Post) AddHashtag(hashtag string) {
	// Check if hashtag already exists
	for _, existing := range p.Hashtags {
		if existing == hashtag {
			return
		}
	}
	p.Hashtags = append(p.Hashtags, hashtag)
}

// AddMention adds a user mention to the post
func (p *Post) AddMention(userID primitive.ObjectID) {
	// Check if mention already exists
	for _, existing := range p.Mentions {
		if existing == userID {
			return
		}
	}
	p.Mentions = append(p.Mentions, userID)
}

// MarkAsEdited marks the post as edited
func (p *Post) MarkAsEdited() {
	p.IsEdited = true
	now := time.Now()
	p.EditedAt = &now
	p.BeforeUpdate()
}
