// models/follow.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FollowStatus represents the status of a follow relationship
type FollowStatus string

const (
	FollowStatusPending  FollowStatus = "pending"  // Follow request sent but not accepted
	FollowStatusAccepted FollowStatus = "accepted" // Follow request accepted
	FollowStatusBlocked  FollowStatus = "blocked"  // User is blocked
	FollowStatusMuted    FollowStatus = "muted"    // User is muted (still following but posts hidden)
)

// Follow represents a follow relationship between two users
type Follow struct {
	BaseModel `bson:",inline"`

	// Follow relationship
	FollowerID primitive.ObjectID `json:"follower_id" bson:"follower_id" validate:"required"` // User who follows
	FolloweeID primitive.ObjectID `json:"followee_id" bson:"followee_id" validate:"required"` // User being followed

	// User information (populated when querying)
	Follower UserResponse `json:"follower,omitempty" bson:"-"`
	Followee UserResponse `json:"followee,omitempty" bson:"-"`

	// Follow status
	Status FollowStatus `json:"status" bson:"status" validate:"required"`

	// Timestamps for different states
	RequestedAt *time.Time `json:"requested_at,omitempty" bson:"requested_at,omitempty"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	BlockedAt   *time.Time `json:"blocked_at,omitempty" bson:"blocked_at,omitempty"`
	MutedAt     *time.Time `json:"muted_at,omitempty" bson:"muted_at,omitempty"`

	// Follow preferences
	NotificationsEnabled bool `json:"notifications_enabled" bson:"notifications_enabled"`
	ShowInFeed           bool `json:"show_in_feed" bson:"show_in_feed"`

	// Categories/Lists (for organizing follows)
	Categories []string `json:"categories,omitempty" bson:"categories,omitempty"` // e.g., "close_friends", "family", "work"

	// Interaction tracking
	InteractionScore  float64    `json:"interaction_score" bson:"interaction_score"` // For feed ranking
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty" bson:"last_interaction_at,omitempty"`

	// Additional metadata
	Source    string `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	IPAddress string `json:"-" bson:"ip_address,omitempty"`
}

// FollowResponse represents the follow data returned in API responses
type FollowResponse struct {
	ID                   string       `json:"id"`
	FollowerID           string       `json:"follower_id"`
	FolloweeID           string       `json:"followee_id"`
	Follower             UserResponse `json:"follower,omitempty"`
	Followee             UserResponse `json:"followee,omitempty"`
	Status               FollowStatus `json:"status"`
	RequestedAt          *time.Time   `json:"requested_at,omitempty"`
	AcceptedAt           *time.Time   `json:"accepted_at,omitempty"`
	NotificationsEnabled bool         `json:"notifications_enabled"`
	ShowInFeed           bool         `json:"show_in_feed"`
	Categories           []string     `json:"categories,omitempty"`
	CreatedAt            time.Time    `json:"created_at"`
}

// CreateFollowRequest represents the request to follow a user
type CreateFollowRequest struct {
	FolloweeID           string   `json:"followee_id" validate:"required"`
	NotificationsEnabled bool     `json:"notifications_enabled"`
	Categories           []string `json:"categories,omitempty"`
}

// UpdateFollowRequest represents the request to update follow settings
type UpdateFollowRequest struct {
	NotificationsEnabled *bool    `json:"notifications_enabled,omitempty"`
	ShowInFeed           *bool    `json:"show_in_feed,omitempty"`
	Categories           []string `json:"categories,omitempty"`
}

// FollowStatsResponse represents follow statistics for a user
type FollowStatsResponse struct {
	UserID         string `json:"user_id"`
	FollowersCount int64  `json:"followers_count"`
	FollowingCount int64  `json:"following_count"`
	MutualFollows  int64  `json:"mutual_follows,omitempty"`
}

// FollowSuggestionResponse represents a follow suggestion
type FollowSuggestionResponse struct {
	User            UserResponse `json:"user"`
	MutualFollows   int64        `json:"mutual_follows"`
	ReasonType      string       `json:"reason_type"` // "mutual_friends", "contacts", "activity", "interests"
	ReasonText      string       `json:"reason_text"`
	ConfidenceScore float64      `json:"confidence_score"`
}

// MutualFollowsResponse represents mutual follows between users
type MutualFollowsResponse struct {
	UserID        string         `json:"user_id"`
	MutualFollows []UserResponse `json:"mutual_follows"`
	TotalCount    int64          `json:"total_count"`
	HasMore       bool           `json:"has_more"`
}

// FollowRequestResponse represents a follow request
type FollowRequestResponse struct {
	ID          string       `json:"id"`
	Requester   UserResponse `json:"requester"`
	RequestedAt time.Time    `json:"requested_at"`
	Message     string       `json:"message,omitempty"`
}

// BatchFollowRequest represents a request to follow multiple users
type BatchFollowRequest struct {
	UserIDs              []string `json:"user_ids" validate:"required,max=50"`
	NotificationsEnabled bool     `json:"notifications_enabled"`
	Categories           []string `json:"categories,omitempty"`
}

// FollowActivity represents a follow activity for activity feed
type FollowActivity struct {
	ID            primitive.ObjectID `json:"id"`
	Type          string             `json:"type"` // "new_follower", "new_following", "follow_request"
	RelatedUserID primitive.ObjectID `json:"related_user_id"`
	RelatedUser   UserResponse       `json:"related_user,omitempty"`
	Status        FollowStatus       `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
}

// BeforeCreate sets default values before creating follow
func (f *Follow) BeforeCreate() {
	f.BaseModel.BeforeCreate()

	// Set default status based on target user's privacy settings
	// This would typically be determined by the service layer
	if f.Status == "" {
		f.Status = FollowStatusPending
	}

	// Set timestamps based on status
	now := time.Now()
	switch f.Status {
	case FollowStatusPending:
		f.RequestedAt = &now
	case FollowStatusAccepted:
		f.RequestedAt = &now
		f.AcceptedAt = &now
	case FollowStatusBlocked:
		f.BlockedAt = &now
	case FollowStatusMuted:
		f.MutedAt = &now
	}

	// Set default preferences
	f.NotificationsEnabled = true
	f.ShowInFeed = true
	f.InteractionScore = 0.0
}

// ToFollowResponse converts Follow model to FollowResponse
func (f *Follow) ToFollowResponse() FollowResponse {
	return FollowResponse{
		ID:                   f.ID.Hex(),
		FollowerID:           f.FollowerID.Hex(),
		FolloweeID:           f.FolloweeID.Hex(),
		Status:               f.Status,
		RequestedAt:          f.RequestedAt,
		AcceptedAt:           f.AcceptedAt,
		NotificationsEnabled: f.NotificationsEnabled,
		ShowInFeed:           f.ShowInFeed,
		Categories:           f.Categories,
		CreatedAt:            f.CreatedAt,
	}
}

// Accept accepts a pending follow request
func (f *Follow) Accept() {
	if f.Status == FollowStatusPending {
		f.Status = FollowStatusAccepted
		now := time.Now()
		f.AcceptedAt = &now
		f.BeforeUpdate()
	}
}

// Reject rejects a pending follow request (soft delete)
func (f *Follow) Reject() {
	if f.Status == FollowStatusPending {
		f.SoftDelete()
	}
}

// Block blocks the follower
func (f *Follow) Block() {
	f.Status = FollowStatusBlocked
	now := time.Now()
	f.BlockedAt = &now
	f.BeforeUpdate()
}

// Unblock unblocks the follower
func (f *Follow) Unblock() {
	if f.Status == FollowStatusBlocked {
		f.Status = FollowStatusAccepted
		f.BlockedAt = nil
		f.BeforeUpdate()
	}
}

// Mute mutes the user (still following but posts hidden from feed)
func (f *Follow) Mute() {
	if f.Status == FollowStatusAccepted {
		f.Status = FollowStatusMuted
		now := time.Now()
		f.MutedAt = &now
		f.BeforeUpdate()
	}
}

// Unmute unmutes the user
func (f *Follow) Unmute() {
	if f.Status == FollowStatusMuted {
		f.Status = FollowStatusAccepted
		f.MutedAt = nil
		f.BeforeUpdate()
	}
}

// UpdateInteractionScore updates the interaction score for feed ranking
func (f *Follow) UpdateInteractionScore(score float64) {
	f.InteractionScore = score
	now := time.Now()
	f.LastInteractionAt = &now
	f.BeforeUpdate()
}

// IsActive returns true if the follow relationship is active (accepted)
func (f *Follow) IsActive() bool {
	return f.Status == FollowStatusAccepted && !f.IsDeleted()
}

// IsPending returns true if the follow request is pending
func (f *Follow) IsPending() bool {
	return f.Status == FollowStatusPending && !f.IsDeleted()
}

// IsBlocked returns true if the relationship is blocked
func (f *Follow) IsBlocked() bool {
	return f.Status == FollowStatusBlocked && !f.IsDeleted()
}

// IsMuted returns true if the user is muted
func (f *Follow) IsMuted() bool {
	return f.Status == FollowStatusMuted && !f.IsDeleted()
}

// CanReceiveNotifications returns true if notifications are enabled
func (f *Follow) CanReceiveNotifications() bool {
	return f.NotificationsEnabled && f.IsActive()
}

// ShouldShowInFeed returns true if posts should show in feed
func (f *Follow) ShouldShowInFeed() bool {
	return f.ShowInFeed && f.IsActive() && !f.IsMuted()
}

// AddCategory adds a category to the follow relationship
func (f *Follow) AddCategory(category string) {
	// Check if category already exists
	for _, existing := range f.Categories {
		if existing == category {
			return
		}
	}
	f.Categories = append(f.Categories, category)
	f.BeforeUpdate()
}

// RemoveCategory removes a category from the follow relationship
func (f *Follow) RemoveCategory(category string) {
	for i, existing := range f.Categories {
		if existing == category {
			f.Categories = append(f.Categories[:i], f.Categories[i+1:]...)
			f.BeforeUpdate()
			return
		}
	}
}

// HasCategory checks if the follow relationship has a specific category
func (f *Follow) HasCategory(category string) bool {
	for _, existing := range f.Categories {
		if existing == category {
			return true
		}
	}
	return false
}

// CanUnfollow checks if the follower can unfollow
func (f *Follow) CanUnfollow(currentUserID primitive.ObjectID) bool {
	return f.FollowerID == currentUserID && !f.IsDeleted()
}

// CanAcceptReject checks if the followee can accept/reject the request
func (f *Follow) CanAcceptReject(currentUserID primitive.ObjectID) bool {
	return f.FolloweeID == currentUserID && f.IsPending()
}

// CanBlock checks if the user can block this relationship
func (f *Follow) CanBlock(currentUserID primitive.ObjectID) bool {
	return (f.FollowerID == currentUserID || f.FolloweeID == currentUserID) && !f.IsDeleted()
}

// Utility functions for follow relationships

// GetRelationshipStatus determines the relationship status between two users
func GetRelationshipStatus(followerID, followeeID primitive.ObjectID, follows []Follow) (isFollowing, isFollowedBy, isPending, isBlocked bool) {
	for _, follow := range follows {
		if follow.FollowerID == followerID && follow.FolloweeID == followeeID {
			switch follow.Status {
			case FollowStatusAccepted:
				isFollowing = true
			case FollowStatusPending:
				isPending = true
			case FollowStatusBlocked:
				isBlocked = true
			}
		}

		if follow.FollowerID == followeeID && follow.FolloweeID == followerID {
			if follow.Status == FollowStatusAccepted {
				isFollowedBy = true
			}
		}
	}

	return
}

// AreMutualFollows checks if two users follow each other
func AreMutualFollows(userID1, userID2 primitive.ObjectID, follows []Follow) bool {
	var user1FollowsUser2, user2FollowsUser1 bool

	for _, follow := range follows {
		if follow.FollowerID == userID1 && follow.FolloweeID == userID2 && follow.IsActive() {
			user1FollowsUser2 = true
		}
		if follow.FollowerID == userID2 && follow.FolloweeID == userID1 && follow.IsActive() {
			user2FollowsUser1 = true
		}
	}

	return user1FollowsUser2 && user2FollowsUser1
}

// GetMutualFollowsCount counts mutual follows between two users
func GetMutualFollowsCount(userID1, userID2 primitive.ObjectID, allFollows map[string][]Follow) int64 {
	// This would typically be implemented with a more efficient database query
	// For now, it's a placeholder for the concept
	var count int64 = 0

	// Implementation would involve finding users who follow both userID1 and userID2
	// This is a complex query that would be better handled by the service layer

	return count
}

// CreateFollowSuggestion creates a follow suggestion response
func CreateFollowSuggestion(user UserResponse, mutualFollows int64, reasonType, reasonText string, confidenceScore float64) FollowSuggestionResponse {
	return FollowSuggestionResponse{
		User:            user,
		MutualFollows:   mutualFollows,
		ReasonType:      reasonType,
		ReasonText:      reasonText,
		ConfidenceScore: confidenceScore,
	}
}

// GetFollowCategories returns predefined follow categories
func GetFollowCategories() []string {
	return []string{
		"close_friends",
		"family",
		"work",
		"school",
		"interests",
		"celebrities",
		"brands",
		"news",
		"sports",
		"entertainment",
	}
}

// IsValidFollowCategory checks if a category is valid
func IsValidFollowCategory(category string) bool {
	validCategories := GetFollowCategories()
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// GetFollowStatusDisplayText returns display text for follow status
func GetFollowStatusDisplayText(status FollowStatus) string {
	switch status {
	case FollowStatusPending:
		return "Requested"
	case FollowStatusAccepted:
		return "Following"
	case FollowStatusBlocked:
		return "Blocked"
	case FollowStatusMuted:
		return "Muted"
	default:
		return "Unknown"
	}
}
