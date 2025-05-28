// models/like.go
package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Like represents a like/reaction on a post, comment, or story
type Like struct {
	BaseModel `bson:",inline"`

	// User who liked/reacted
	UserID primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	User   UserResponse       `json:"user,omitempty" bson:"-"` // Populated when querying

	// Target of like/reaction (can be post, comment, story, etc.)
	TargetID   primitive.ObjectID `json:"target_id" bson:"target_id" validate:"required"`
	TargetType string             `json:"target_type" bson:"target_type" validate:"required,oneof=post comment story message"`

	// Reaction type (like, love, haha, wow, sad, angry, support)
	ReactionType ReactionType `json:"reaction_type" bson:"reaction_type" validate:"required"`

	// Additional metadata
	Source    string `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api
	IPAddress string `json:"-" bson:"ip_address,omitempty"`
}

// LikeResponse represents the like data returned in API responses
type LikeResponse struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	User         UserResponse `json:"user,omitempty"`
	TargetID     string       `json:"target_id"`
	TargetType   string       `json:"target_type"`
	ReactionType ReactionType `json:"reaction_type"`
	CreatedAt    string       `json:"created_at"`
}

// CreateLikeRequest represents the request to create a like/reaction
type CreateLikeRequest struct {
	TargetID     string       `json:"target_id" validate:"required"`
	TargetType   string       `json:"target_type" validate:"required,oneof=post comment story message"`
	ReactionType ReactionType `json:"reaction_type" validate:"required"`
}

// UpdateLikeRequest represents the request to update a reaction
type UpdateLikeRequest struct {
	ReactionType ReactionType `json:"reaction_type" validate:"required"`
}

// ReactionSummary represents aggregated reaction counts for a target
type ReactionSummary struct {
	TargetID     string                 `json:"target_id"`
	TargetType   string                 `json:"target_type"`
	TotalCount   int64                  `json:"total_count"`
	Reactions    map[ReactionType]int64 `json:"reactions"`
	TopReactors  []UserResponse         `json:"top_reactors,omitempty"`
	UserReaction ReactionType           `json:"user_reaction,omitempty"` // Current user's reaction
}

// ReactionStats represents detailed reaction statistics
type ReactionStats struct {
	Like    ReactionCount `json:"like"`
	Love    ReactionCount `json:"love"`
	Haha    ReactionCount `json:"haha"`
	Wow     ReactionCount `json:"wow"`
	Sad     ReactionCount `json:"sad"`
	Angry   ReactionCount `json:"angry"`
	Support ReactionCount `json:"support"`
}

// ReactionCount represents count and sample users for a reaction type
type ReactionCount struct {
	Count       int64          `json:"count"`
	SampleUsers []UserResponse `json:"sample_users,omitempty"` // First few users who reacted
}

// ReactionUsersResponse represents users who reacted with a specific reaction
type ReactionUsersResponse struct {
	ReactionType ReactionType   `json:"reaction_type"`
	Users        []UserResponse `json:"users"`
	TotalCount   int64          `json:"total_count"`
	HasMore      bool           `json:"has_more"`
}

// Methods for Like model

// BeforeCreate sets default values before creating like
func (l *Like) BeforeCreate() {
	l.BaseModel.BeforeCreate()

	// Set default reaction type if not specified
	if l.ReactionType == "" {
		l.ReactionType = ReactionLike
	}
}

// ToLikeResponse converts Like model to LikeResponse
func (l *Like) ToLikeResponse() LikeResponse {
	return LikeResponse{
		ID:           l.ID.Hex(),
		UserID:       l.UserID.Hex(),
		TargetID:     l.TargetID.Hex(),
		TargetType:   l.TargetType,
		ReactionType: l.ReactionType,
		CreatedAt:    l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CanUpdateReaction checks if a user can update this reaction
func (l *Like) CanUpdateReaction(currentUserID primitive.ObjectID) bool {
	return l.UserID == currentUserID && !l.IsDeleted()
}

// CanDeleteReaction checks if a user can delete this reaction
func (l *Like) CanDeleteReaction(currentUserID primitive.ObjectID, userRole UserRole) bool {
	// User can delete their own reaction
	if l.UserID == currentUserID {
		return true
	}

	// Moderators and admins can delete any reaction
	return userRole == RoleModerator || userRole == RoleAdmin || userRole == RoleSuperAdmin
}

// GetReactionEmoji returns the emoji representation of the reaction
func (l *Like) GetReactionEmoji() string {
	switch l.ReactionType {
	case ReactionLike:
		return "👍"
	case ReactionLove:
		return "❤️"
	case ReactionHaha:
		return "😂"
	case ReactionWow:
		return "😮"
	case ReactionSad:
		return "😢"
	case ReactionAngry:
		return "😠"
	case ReactionSupport:
		return "🤗"
	default:
		return "👍"
	}
}

// Utility functions for reaction handling

// GetReactionEmoji returns emoji for a reaction type
func GetReactionEmoji(reactionType ReactionType) string {
	switch reactionType {
	case ReactionLike:
		return "👍"
	case ReactionLove:
		return "❤️"
	case ReactionHaha:
		return "😂"
	case ReactionWow:
		return "😮"
	case ReactionSad:
		return "😢"
	case ReactionAngry:
		return "😠"
	case ReactionSupport:
		return "🤗"
	default:
		return "👍"
	}
}

// GetReactionName returns human-readable name for a reaction type
func GetReactionName(reactionType ReactionType) string {
	switch reactionType {
	case ReactionLike:
		return "Like"
	case ReactionLove:
		return "Love"
	case ReactionHaha:
		return "Haha"
	case ReactionWow:
		return "Wow"
	case ReactionSad:
		return "Sad"
	case ReactionAngry:
		return "Angry"
	case ReactionSupport:
		return "Support"
	default:
		return "Like"
	}
}

// IsValidReactionType checks if a reaction type is valid
func IsValidReactionType(reactionType ReactionType) bool {
	validTypes := []ReactionType{
		ReactionLike,
		ReactionLove,
		ReactionHaha,
		ReactionWow,
		ReactionSad,
		ReactionAngry,
		ReactionSupport,
	}

	for _, validType := range validTypes {
		if reactionType == validType {
			return true
		}
	}
	return false
}

// GetAllReactionTypes returns all available reaction types
func GetAllReactionTypes() []ReactionType {
	return []ReactionType{
		ReactionLike,
		ReactionLove,
		ReactionHaha,
		ReactionWow,
		ReactionSad,
		ReactionAngry,
		ReactionSupport,
	}
}

// CreateReactionSummary creates a reaction summary from aggregated data
func CreateReactionSummary(targetID, targetType string, reactions map[ReactionType]int64, topReactors []UserResponse, userReaction ReactionType) ReactionSummary {
	var totalCount int64
	for _, count := range reactions {
		totalCount += count
	}

	return ReactionSummary{
		TargetID:     targetID,
		TargetType:   targetType,
		TotalCount:   totalCount,
		Reactions:    reactions,
		TopReactors:  topReactors,
		UserReaction: userReaction,
	}
}

// CreateReactionStats creates detailed reaction statistics
func CreateReactionStats(reactions map[ReactionType]ReactionCount) ReactionStats {
	// Initialize with zero counts
	stats := ReactionStats{
		Like:    ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
		Love:    ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
		Haha:    ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
		Wow:     ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
		Sad:     ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
		Angry:   ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
		Support: ReactionCount{Count: 0, SampleUsers: []UserResponse{}},
	}

	// Fill in actual data
	if count, exists := reactions[ReactionLike]; exists {
		stats.Like = count
	}
	if count, exists := reactions[ReactionLove]; exists {
		stats.Love = count
	}
	if count, exists := reactions[ReactionHaha]; exists {
		stats.Haha = count
	}
	if count, exists := reactions[ReactionWow]; exists {
		stats.Wow = count
	}
	if count, exists := reactions[ReactionSad]; exists {
		stats.Sad = count
	}
	if count, exists := reactions[ReactionAngry]; exists {
		stats.Angry = count
	}
	if count, exists := reactions[ReactionSupport]; exists {
		stats.Support = count
	}

	return stats
}


// GetTopReactions returns the top N reaction types by count
func GetTopReactions(reactions map[ReactionType]int64, limit int) []struct {
	Type  ReactionType `json:"type"`
	Count int64        `json:"count"`
	Emoji string       `json:"emoji"`
	Name  string       `json:"name"`
} {
	type reactionInfo struct {
		Type  ReactionType `json:"type"`
		Count int64        `json:"count"`
		Emoji string       `json:"emoji"`
		Name  string       `json:"name"`
	}

	var sortedReactions []reactionInfo

	// Convert map to slice for sorting
	for reactionType, count := range reactions {
		if count > 0 {
			sortedReactions = append(sortedReactions, reactionInfo{
				Type:  reactionType,
				Count: count,
				Emoji: GetReactionEmoji(reactionType),
				Name:  GetReactionName(reactionType),
			})
		}
	}

	// Sort by count (descending)
	for i := 0; i < len(sortedReactions)-1; i++ {
		for j := i + 1; j < len(sortedReactions); j++ {
			if sortedReactions[i].Count < sortedReactions[j].Count {
				sortedReactions[i], sortedReactions[j] = sortedReactions[j], sortedReactions[i]
			}
		}
	}

	// Return top N reactions
	if len(sortedReactions) > limit {
		return sortedReactions[:limit]
	}
	return sortedReactions
}

// FormatReactionText formats reaction text for display (e.g., "You and 5 others reacted")
func FormatReactionText(totalCount int64, userReaction ReactionType, currentUserID string) string {
	if totalCount == 0 {
		return ""
	}

	if userReaction != "" {
		if totalCount == 1 {
			return "You reacted"
		}
		othersCount := totalCount - 1
		if othersCount == 1 {
			return "You and 1 other reacted"
		}
		return fmt.Sprintf("You and %d others reacted", othersCount)
	}

	if totalCount == 1 {
		return "1 person reacted"
	}
	return fmt.Sprintf("%d people reacted", totalCount)
}

// Like aggregation helper functions

// AggregateReactionsByTarget aggregates reactions for multiple targets
func AggregateReactionsByTarget(likes []Like) map[string]map[ReactionType]int64 {
	result := make(map[string]map[ReactionType]int64)

	for _, like := range likes {
		targetKey := like.TargetID.Hex()

		if result[targetKey] == nil {
			result[targetKey] = make(map[ReactionType]int64)
		}

		result[targetKey][like.ReactionType]++
	}

	return result
}

// GetUserReactions gets user's reactions for multiple targets
func GetUserReactions(likes []Like, userID primitive.ObjectID) map[string]ReactionType {
	result := make(map[string]ReactionType)

	for _, like := range likes {
		if like.UserID == userID {
			result[like.TargetID.Hex()] = like.ReactionType
		}
	}

	return result
}
