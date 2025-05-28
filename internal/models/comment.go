
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a comment on a post or reply to another comment
type Comment struct {
	BaseModel `bson:",inline"`

	// Author Information
	UserID primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	Author UserResponse       `json:"author,omitempty" bson:"-"` // Populated when querying

	// Content
	Content     string      `json:"content" bson:"content" validate:"required,max=2000"`
	ContentType ContentType `json:"content_type" bson:"content_type"`
	Media       []MediaInfo `json:"media,omitempty" bson:"media,omitempty"`

	// Comment Hierarchy
	PostID          primitive.ObjectID  `json:"post_id" bson:"post_id" validate:"required"`
	ParentCommentID *primitive.ObjectID `json:"parent_comment_id,omitempty" bson:"parent_comment_id,omitempty"`
	RootCommentID   *primitive.ObjectID `json:"root_comment_id,omitempty" bson:"root_comment_id,omitempty"`
	Level           int                 `json:"level" bson:"level"`             // 0 for top-level, 1 for replies, 2 for nested replies
	ThreadPath      string              `json:"thread_path" bson:"thread_path"` // For efficient nested comments querying

	// Engagement Statistics
	LikesCount   int64 `json:"likes_count" bson:"likes_count"`
	RepliesCount int64 `json:"replies_count" bson:"replies_count"`

	// Social Features
	Mentions     []primitive.ObjectID `json:"mentions,omitempty" bson:"mentions,omitempty"`
	MentionUsers []UserResponse       `json:"mention_users,omitempty" bson:"-"` // Populated when querying

	// Comment Status
	IsEdited      bool       `json:"is_edited" bson:"is_edited"`
	EditedAt      *time.Time `json:"edited_at,omitempty" bson:"edited_at,omitempty"`
	IsPinned      bool       `json:"is_pinned" bson:"is_pinned"`
	IsHighlighted bool       `json:"is_highlighted" bson:"is_highlighted"` // Highlighted by post author

	// Content Moderation
	IsReported   bool  `json:"is_reported" bson:"is_reported"`
	ReportsCount int64 `json:"reports_count" bson:"reports_count"`
	IsHidden     bool  `json:"is_hidden" bson:"is_hidden"`
	IsApproved   bool  `json:"is_approved" bson:"is_approved"`

	// Additional Metadata
	Source    string `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	IPAddress string `json:"-" bson:"ip_address,omitempty"`
	UserAgent string `json:"-" bson:"user_agent,omitempty"`

	// Voting System (for community-driven sorting)
	UpvotesCount   int64 `json:"upvotes_count" bson:"upvotes_count"`
	DownvotesCount int64 `json:"downvotes_count" bson:"downvotes_count"`
	VoteScore      int64 `json:"vote_score" bson:"vote_score"` // upvotes - downvotes

	// Comment Quality Metrics
	QualityScore    float64 `json:"quality_score" bson:"quality_score"`
	IsVerifiedReply bool    `json:"is_verified_reply" bson:"is_verified_reply"` // Reply from verified accounts
	IsAuthorReply   bool    `json:"is_author_reply" bson:"is_author_reply"`     // Reply from post author

	// Time-based Features
	IsLatestReply bool       `json:"is_latest_reply" bson:"is_latest_reply"`
	LastReplyAt   *time.Time `json:"last_reply_at,omitempty" bson:"last_reply_at,omitempty"`

	// Awards/Recognition
	Awards      []CommentAward `json:"awards,omitempty" bson:"awards,omitempty"`
	AwardsCount int64          `json:"awards_count" bson:"awards_count"`
}

// CommentAward represents awards given to comments
type CommentAward struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type        string             `json:"type" bson:"type"` // helpful, funny, insightful, etc.
	GivenBy     primitive.ObjectID `json:"given_by" bson:"given_by"`
	GivenAt     time.Time          `json:"given_at" bson:"given_at"`
	Message     string             `json:"message,omitempty" bson:"message,omitempty"`
	IsAnonymous bool               `json:"is_anonymous" bson:"is_anonymous"`
}

// CommentResponse represents the comment data returned in API responses
type CommentResponse struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Author          UserResponse   `json:"author"`
	Content         string         `json:"content"`
	ContentType     ContentType    `json:"content_type"`
	Media           []MediaInfo    `json:"media,omitempty"`
	PostID          string         `json:"post_id"`
	ParentCommentID string         `json:"parent_comment_id,omitempty"`
	RootCommentID   string         `json:"root_comment_id,omitempty"`
	Level           int            `json:"level"`
	LikesCount      int64          `json:"likes_count"`
	RepliesCount    int64          `json:"replies_count"`
	Mentions        []string       `json:"mentions,omitempty"`
	MentionUsers    []UserResponse `json:"mention_users,omitempty"`
	IsEdited        bool           `json:"is_edited"`
	EditedAt        *time.Time     `json:"edited_at,omitempty"`
	IsPinned        bool           `json:"is_pinned"`
	IsHighlighted   bool           `json:"is_highlighted"`
	UpvotesCount    int64          `json:"upvotes_count"`
	DownvotesCount  int64          `json:"downvotes_count"`
	VoteScore       int64          `json:"vote_score"`
	QualityScore    float64        `json:"quality_score"`
	IsVerifiedReply bool           `json:"is_verified_reply"`
	IsAuthorReply   bool           `json:"is_author_reply"`
	Awards          []CommentAward `json:"awards,omitempty"`
	AwardsCount     int64          `json:"awards_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// User-specific context
	IsLiked      bool              `json:"is_liked,omitempty"`
	UserReaction ReactionType      `json:"user_reaction,omitempty"`
	UserVote     string            `json:"user_vote,omitempty"` // upvote, downvote, none
	Replies      []CommentResponse `json:"replies,omitempty"`   // Nested replies
	TimeAgo      string            `json:"time_ago,omitempty"`
	CanEdit      bool              `json:"can_edit,omitempty"`
	CanDelete    bool              `json:"can_delete,omitempty"`
	CanReply     bool              `json:"can_reply,omitempty"`
}

// CreateCommentRequest represents the request to create a new comment
type CreateCommentRequest struct {
	PostID          string      `json:"post_id" validate:"required"`
	ParentCommentID string      `json:"parent_comment_id,omitempty"`
	Content         string      `json:"content" validate:"required,max=2000"`
	ContentType     ContentType `json:"content_type" validate:"required,oneof=text image gif"`
	Media           []MediaInfo `json:"media,omitempty"`
	Mentions        []string    `json:"mentions,omitempty"` // User IDs as strings
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Content  *string     `json:"content,omitempty" validate:"omitempty,max=2000"`
	Media    []MediaInfo `json:"media,omitempty"`
	Mentions []string    `json:"mentions,omitempty"`
}

// CommentVoteRequest represents a vote on a comment
type CommentVoteRequest struct {
	VoteType string `json:"vote_type" validate:"required,oneof=upvote downvote remove"`
}

// CommentAwardRequest represents giving an award to a comment
type CommentAwardRequest struct {
	Type        string `json:"type" validate:"required"`
	Message     string `json:"message,omitempty" validate:"max=200"`
	IsAnonymous bool   `json:"is_anonymous"`
}

// CommentStatsResponse represents comment statistics
type CommentStatsResponse struct {
	CommentID    string  `json:"comment_id"`
	LikesCount   int64   `json:"likes_count"`
	RepliesCount int64   `json:"replies_count"`
	VoteScore    int64   `json:"vote_score"`
	QualityScore float64 `json:"quality_score"`
}

// CommentTreeResponse represents a comment with its nested replies
type CommentTreeResponse struct {
	CommentResponse `json:",inline"`
	Children        []CommentTreeResponse `json:"children,omitempty"`
	HasMoreReplies  bool                  `json:"has_more_replies,omitempty"`
	TotalReplies    int64                 `json:"total_replies,omitempty"`
	LoadMoreURL     string                `json:"load_more_url,omitempty"`
}

// CommentModerationRequest represents comment moderation actions
type CommentModerationRequest struct {
	Action string `json:"action" validate:"required,oneof=approve hide pin unpin highlight unhighlight"`
	Reason string `json:"reason,omitempty" validate:"max=500"`
}

// CommentSearchRequest represents comment search parameters
type CommentSearchRequest struct {
	Query  string `json:"query" validate:"required,min=1"`
	PostID string `json:"post_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	SortBy string `json:"sort_by,omitempty" validate:"omitempty,oneof=newest oldest top controversial"`
	Page   int    `json:"page,omitempty" validate:"min=1"`
	Limit  int    `json:"limit,omitempty" validate:"min=1,max=50"`
}

// Methods for Comment model

// BeforeCreate sets default values before creating comment
func (c *Comment) BeforeCreate() {
	c.BaseModel.BeforeCreate()

	// Set default values
	c.LikesCount = 0
	c.RepliesCount = 0
	c.ReportsCount = 0
	c.UpvotesCount = 0
	c.DownvotesCount = 0
	c.VoteScore = 0
	c.QualityScore = 0.0
	c.AwardsCount = 0
	c.IsEdited = false
	c.IsPinned = false
	c.IsHighlighted = false
	c.IsReported = false
	c.IsHidden = false
	c.IsApproved = true // Auto-approve by default
	c.IsVerifiedReply = false
	c.IsAuthorReply = false
	c.IsLatestReply = false

	// Set default content type
	if c.ContentType == "" {
		c.ContentType = ContentTypeText
	}

	// Set source if not provided
	if c.Source == "" {
		c.Source = "web"
	}

	// Set thread path for nested comments
	c.updateThreadPath()
}

// updateThreadPath generates a thread path for efficient nested comment queries
func (c *Comment) updateThreadPath() {
	if c.ParentCommentID == nil {
		// Top-level comment
		c.Level = 0
		c.ThreadPath = c.ID.Hex()
	} else {
		// This would typically be set by the service layer with parent's thread path
		// Format: "parent_path/comment_id"
		c.Level = 1 // This would be calculated based on parent's level + 1
	}
}

// ToCommentResponse converts Comment model to CommentResponse
func (c *Comment) ToCommentResponse() CommentResponse {
	response := CommentResponse{
		ID:              c.ID.Hex(),
		UserID:          c.UserID.Hex(),
		Content:         c.Content,
		ContentType:     c.ContentType,
		Media:           c.Media,
		PostID:          c.PostID.Hex(),
		Level:           c.Level,
		LikesCount:      c.LikesCount,
		RepliesCount:    c.RepliesCount,
		IsEdited:        c.IsEdited,
		EditedAt:        c.EditedAt,
		IsPinned:        c.IsPinned,
		IsHighlighted:   c.IsHighlighted,
		UpvotesCount:    c.UpvotesCount,
		DownvotesCount:  c.DownvotesCount,
		VoteScore:       c.VoteScore,
		QualityScore:    c.QualityScore,
		IsVerifiedReply: c.IsVerifiedReply,
		IsAuthorReply:   c.IsAuthorReply,
		Awards:          c.Awards,
		AwardsCount:     c.AwardsCount,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}

	// Convert ObjectIDs to strings
	if c.ParentCommentID != nil {
		response.ParentCommentID = c.ParentCommentID.Hex()
	}

	if c.RootCommentID != nil {
		response.RootCommentID = c.RootCommentID.Hex()
	}

	if len(c.Mentions) > 0 {
		response.Mentions = make([]string, len(c.Mentions))
		for i, mention := range c.Mentions {
			response.Mentions[i] = mention.Hex()
		}
	}

	return response
}

// ToCommentStatsResponse converts Comment model to CommentStatsResponse
func (c *Comment) ToCommentStatsResponse() CommentStatsResponse {
	return CommentStatsResponse{
		CommentID:    c.ID.Hex(),
		LikesCount:   c.LikesCount,
		RepliesCount: c.RepliesCount,
		VoteScore:    c.VoteScore,
		QualityScore: c.QualityScore,
	}
}

// IncrementLikesCount increments the likes count
func (c *Comment) IncrementLikesCount() {
	c.LikesCount++
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// DecrementLikesCount decrements the likes count
func (c *Comment) DecrementLikesCount() {
	if c.LikesCount > 0 {
		c.LikesCount--
	}
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// IncrementRepliesCount increments the replies count
func (c *Comment) IncrementRepliesCount() {
	c.RepliesCount++
	c.IsLatestReply = false // No longer the latest if it has replies
	now := time.Now()
	c.LastReplyAt = &now
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// DecrementRepliesCount decrements the replies count
func (c *Comment) DecrementRepliesCount() {
	if c.RepliesCount > 0 {
		c.RepliesCount--
	}
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// Upvote adds an upvote to the comment
func (c *Comment) Upvote() {
	c.UpvotesCount++
	c.VoteScore = c.UpvotesCount - c.DownvotesCount
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// Downvote adds a downvote to the comment
func (c *Comment) Downvote() {
	c.DownvotesCount++
	c.VoteScore = c.UpvotesCount - c.DownvotesCount
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// RemoveUpvote removes an upvote from the comment
func (c *Comment) RemoveUpvote() {
	if c.UpvotesCount > 0 {
		c.UpvotesCount--
		c.VoteScore = c.UpvotesCount - c.DownvotesCount
		c.UpdateQualityScore()
		c.BeforeUpdate()
	}
}

// RemoveDownvote removes a downvote from the comment
func (c *Comment) RemoveDownvote() {
	if c.DownvotesCount > 0 {
		c.DownvotesCount--
		c.VoteScore = c.UpvotesCount - c.DownvotesCount
		c.UpdateQualityScore()
		c.BeforeUpdate()
	}
}

// UpdateQualityScore calculates and updates the quality score
func (c *Comment) UpdateQualityScore() {
	// Simple quality scoring algorithm
	// In practice, this would be more sophisticated
	engagementScore := float64(c.LikesCount + c.RepliesCount)
	voteScore := float64(c.VoteScore)
	ageScore := float64(time.Since(c.CreatedAt).Hours())

	// Newer comments get slight boost, engagement matters most
	c.QualityScore = (engagementScore + voteScore) / (1 + ageScore/24)

	// Boost for verified replies and author replies
	if c.IsVerifiedReply {
		c.QualityScore *= 1.2
	}
	if c.IsAuthorReply {
		c.QualityScore *= 1.5
	}

	// Awards boost
	c.QualityScore += float64(c.AwardsCount) * 0.5
}

// CanEditComment checks if a user can edit this comment
func (c *Comment) CanEditComment(currentUserID primitive.ObjectID) bool {
	return c.UserID == currentUserID && !c.IsDeleted()
}

// CanDeleteComment checks if a user can delete this comment
func (c *Comment) CanDeleteComment(currentUserID primitive.ObjectID, userRole UserRole, isPostAuthor bool) bool {
	// Comment author can delete their own comment
	if c.UserID == currentUserID {
		return true
	}

	// Post author can delete comments on their post
	if isPostAuthor {
		return true
	}

	// Moderators and admins can delete any comment
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// CanReplyToComment checks if a user can reply to this comment
func (c *Comment) CanReplyToComment(maxNestingLevel int) bool {
	// Check if nesting level allows replies
	return c.Level < maxNestingLevel && !c.IsDeleted() && !c.IsHidden
}

// CanModerateComment checks if a user can moderate this comment
func (c *Comment) CanModerateComment(currentUserID primitive.ObjectID, userRole UserRole, isPostAuthor bool) bool {
	// Post author can moderate comments on their post
	if isPostAuthor {
		return true
	}

	// Moderators and admins can moderate any comment
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// MarkAsEdited marks the comment as edited
func (c *Comment) MarkAsEdited() {
	c.IsEdited = true
	now := time.Now()
	c.EditedAt = &now
	c.BeforeUpdate()
}

// HighlightComment highlights the comment (usually by post author)
func (c *Comment) HighlightComment() {
	c.IsHighlighted = true
	c.BeforeUpdate()
}

// UnhighlightComment removes highlight from the comment
func (c *Comment) UnhighlightComment() {
	c.IsHighlighted = false
	c.BeforeUpdate()
}

// PinComment pins the comment
func (c *Comment) PinComment() {
	c.IsPinned = true
	c.BeforeUpdate()
}

// UnpinComment unpins the comment
func (c *Comment) UnpinComment() {
	c.IsPinned = false
	c.BeforeUpdate()
}

// HideComment hides the comment from public view
func (c *Comment) HideComment() {
	c.IsHidden = true
	c.BeforeUpdate()
}

// ShowComment shows the comment in public view
func (c *Comment) ShowComment() {
	c.IsHidden = false
	c.BeforeUpdate()
}

// ApproveComment approves the comment for display
func (c *Comment) ApproveComment() {
	c.IsApproved = true
	c.BeforeUpdate()
}

// RejectComment rejects the comment from display
func (c *Comment) RejectComment() {
	c.IsApproved = false
	c.BeforeUpdate()
}

// AddAward adds an award to the comment
func (c *Comment) AddAward(awardType string, givenBy primitive.ObjectID, message string, isAnonymous bool) {
	award := CommentAward{
		ID:          primitive.NewObjectID(),
		Type:        awardType,
		GivenBy:     givenBy,
		GivenAt:     time.Now(),
		Message:     message,
		IsAnonymous: isAnonymous,
	}

	c.Awards = append(c.Awards, award)
	c.AwardsCount++
	c.UpdateQualityScore()
	c.BeforeUpdate()
}

// AddMention adds a user mention to the comment
func (c *Comment) AddMention(userID primitive.ObjectID) {
	// Check if mention already exists
	for _, existing := range c.Mentions {
		if existing == userID {
			return
		}
	}
	c.Mentions = append(c.Mentions, userID)
}

// IsTopLevel checks if this is a top-level comment (not a reply)
func (c *Comment) IsTopLevel() bool {
	return c.ParentCommentID == nil && c.Level == 0
}

// IsReply checks if this comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentCommentID != nil
}

// GetThreadDepth returns the depth of this comment in the thread
func (c *Comment) GetThreadDepth() int {
	return c.Level
}

// SetThreadInfo sets the thread information for nested comments
func (c *Comment) SetThreadInfo(parentComment *Comment) {
	if parentComment == nil {
		// Top-level comment
		c.Level = 0
		c.RootCommentID = nil
		c.ThreadPath = c.ID.Hex()
	} else {
		// Reply comment
		c.Level = parentComment.Level + 1
		c.ParentCommentID = &parentComment.ID

		// Set root comment ID
		if parentComment.RootCommentID != nil {
			c.RootCommentID = parentComment.RootCommentID
		} else {
			c.RootCommentID = &parentComment.ID
		}

		// Build thread path
		c.ThreadPath = parentComment.ThreadPath + "/" + c.ID.Hex()
	}
}

// CanViewComment checks if a user can view this comment
func (c *Comment) CanViewComment() bool {
	return !c.IsDeleted() && !c.IsHidden && c.IsApproved
}

// GetEngagementRate calculates the engagement rate of the comment
func (c *Comment) GetEngagementRate() float64 {
	totalEngagements := c.LikesCount + c.RepliesCount + c.UpvotesCount + c.DownvotesCount
	if totalEngagements == 0 {
		return 0.0
	}

	// Simple engagement rate based on time since creation
	hoursSinceCreation := time.Since(c.CreatedAt).Hours()
	if hoursSinceCreation == 0 {
		return float64(totalEngagements)
	}

	return float64(totalEngagements) / hoursSinceCreation
}

// GetMentionedUsernames extracts mentioned usernames from content
func (c *Comment) GetMentionedUsernames() []string {
	return ExtractMentionsFromText(c.Content)
}

// BuildCommentTree builds a tree structure from a flat list of comments
func BuildCommentTree(comments []Comment, maxDepth int) []CommentTreeResponse {
	commentMap := make(map[string]CommentTreeResponse)
	var rootComments []CommentTreeResponse

	// First pass: create all comment responses
	for _, comment := range comments {
		response := CommentTreeResponse{
			CommentResponse: comment.ToCommentResponse(),
			Children:        []CommentTreeResponse{},
		}
		commentMap[comment.ID.Hex()] = response

		if comment.IsTopLevel() {
			rootComments = append(rootComments, response)
		}
	}

	// Second pass: build the tree structure
	for _, comment := range comments {
		if comment.ParentCommentID != nil && comment.Level <= maxDepth {
			parentID := comment.ParentCommentID.Hex()
			if parent, exists := commentMap[parentID]; exists {
				parent.Children = append(parent.Children, commentMap[comment.ID.Hex()])
				commentMap[parentID] = parent
			}
		}
	}

	return rootComments
}

// SortCommentsByQuality sorts comments by quality score
func SortCommentsByQuality(comments []Comment) []Comment {
	// Simple bubble sort by quality score (descending)
	for i := 0; i < len(comments)-1; i++ {
		for j := i + 1; j < len(comments); j++ {
			if comments[i].QualityScore < comments[j].QualityScore {
				comments[i], comments[j] = comments[j], comments[i]
			}
		}
	}
	return comments
}

// SortCommentsByTime sorts comments by creation time
func SortCommentsByTime(comments []Comment, ascending bool) []Comment {
	for i := 0; i < len(comments)-1; i++ {
		for j := i + 1; j < len(comments); j++ {
			var shouldSwap bool
			if ascending {
				shouldSwap = comments[i].CreatedAt.After(comments[j].CreatedAt)
			} else {
				shouldSwap = comments[i].CreatedAt.Before(comments[j].CreatedAt)
			}

			if shouldSwap {
				comments[i], comments[j] = comments[j], comments[i]
			}
		}
	}
	return comments
}

// FilterCommentsByLevel filters comments by nesting level
func FilterCommentsByLevel(comments []Comment, maxLevel int) []Comment {
	var filtered []Comment
	for _, comment := range comments {
		if comment.Level <= maxLevel {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

// GetCommentAwardTypes returns available comment award types
func GetCommentAwardTypes() []string {
	return []string{
		"helpful",
		"funny",
		"insightful",
		"informative",
		"thoughtful",
		"creative",
		"wholesome",
		"gold",
		"silver",
		"bronze",
	}
}

// IsValidCommentAwardType checks if an award type is valid
func IsValidCommentAwardType(awardType string) bool {
	validTypes := GetCommentAwardTypes()
	for _, validType := range validTypes {
		if awardType == validType {
			return true
		}
	}
	return false
}
