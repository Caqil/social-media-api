// internal/handlers/follow.go
package handlers

import (
	"strings"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FollowHandler struct {
	followService *services.FollowService
	validator     *validator.Validate
}

func NewFollowHandler(followService *services.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
		validator:     validator.New(),
	}
}

// GetFollowers retrieves user's followers
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		currentUserID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	followers, err := h.followService.GetFollowers(userID, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get followers", err)
		return
	}

	totalCount := int64(len(followers))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Followers retrieved successfully", followers, paginationMeta, nil)
}

// GetFollowing retrieves users that a user is following
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	userIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		currentUserID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	following, err := h.followService.GetFollowing(userID, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get following", err)
		return
	}

	totalCount := int64(len(following))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Following retrieved successfully", following, paginationMeta, nil)
}

// GetFollowStats retrieves follow statistics for a user
func (h *FollowHandler) GetFollowStats(c *gin.Context) {
	userIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	stats, err := h.followService.GetFollowStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get follow statistics", err)
		return
	}

	utils.OkResponse(c, "Follow statistics retrieved successfully", stats)
}

// GetMutualFollows retrieves mutual follows between current user and another user
func (h *FollowHandler) GetMutualFollows(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	targetUserIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	targetUserID, err := primitive.ObjectIDFromHex(targetUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	mutualFollows, err := h.followService.GetMutualFollows(userID.(primitive.ObjectID), targetUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get mutual follows", err)
		return
	}

	totalCount := int64(len(mutualFollows))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Mutual follows retrieved successfully", mutualFollows, paginationMeta, nil)
}

// CheckFollowStatus checks if current user follows another user
func (h *FollowHandler) CheckFollowStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	targetUserIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	targetUserID, err := primitive.ObjectIDFromHex(targetUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	status, err := h.followService.GetFollowStatus(userID.(primitive.ObjectID), targetUserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to check follow status", err)
		return
	}

	utils.OkResponse(c, "Follow status retrieved successfully", gin.H{
		"target_user_id": targetUserIDStr,
		"status":         status,
		"is_following":   status == "accepted",
		"is_pending":     status == "pending",
	})
}

// FollowUser follows another user
func (h *FollowHandler) FollowUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	followeeIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	followeeID, err := primitive.ObjectIDFromHex(followeeIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Check if user is trying to follow themselves
	if userID.(primitive.ObjectID) == followeeID {
		utils.BadRequestResponse(c, "Cannot follow yourself", nil)
		return
	}

	follow, err := h.followService.FollowUser(userID.(primitive.ObjectID), followeeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		if strings.Contains(err.Error(), "already following") {
			utils.ConflictResponse(c, "Already following this user", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to follow user", err)
		return
	}

	utils.CreatedResponse(c, "User followed successfully", gin.H{
		"follow_id":   follow.ID.Hex(),
		"followee_id": followeeIDStr,
		"status":      follow.Status,
		"followed_at": follow.CreatedAt,
	})
}

// UnfollowUser unfollows another user
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	followeeIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	followeeID, err := primitive.ObjectIDFromHex(followeeIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.followService.UnfollowUser(userID.(primitive.ObjectID), followeeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Follow relationship not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to unfollow user", err)
		return
	}

	utils.OkResponse(c, "User unfollowed successfully", gin.H{
		"followee_id": followeeIDStr,
		"unfollowed":  true,
	})
}

// RemoveFollower removes a follower
func (h *FollowHandler) RemoveFollower(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	followerIDStr := c.Param("id") // ✅ Changed from "userId" to "id"
	followerID, err := primitive.ObjectIDFromHex(followerIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.followService.RemoveFollower(userID.(primitive.ObjectID), followerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Follower relationship not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove follower", err)
		return
	}

	utils.OkResponse(c, "Follower removed successfully", gin.H{
		"follower_id": followerIDStr,
		"removed":     true,
	})
}

// GetFollowRequests retrieves pending follow requests
func (h *FollowHandler) GetFollowRequests(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get type parameter (received or sent)
	requestType := c.DefaultQuery("type", "received")
	if requestType != "received" && requestType != "sent" {
		requestType = "received"
	}

	var requests []models.FollowResponse
	var err error

	if requestType == "received" {
		requests, err = h.followService.GetPendingFollowRequests(userID.(primitive.ObjectID), params.Limit, params.Offset)
	} else {
		requests, err = h.followService.GetSentFollowRequests(userID.(primitive.ObjectID), params.Limit, params.Offset)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get follow requests", err)
		return
	}

	totalCount := int64(len(requests))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Follow requests retrieved successfully", requests, paginationMeta, nil)
}

// AcceptFollowRequest accepts a follow request
func (h *FollowHandler) AcceptFollowRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	followIDStr := c.Param("followId")
	followID, err := primitive.ObjectIDFromHex(followIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid follow ID format", err)
		return
	}

	err = h.followService.AcceptFollowRequest(followID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Follow request not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to accept follow request", err)
		return
	}

	utils.OkResponse(c, "Follow request accepted successfully", gin.H{
		"follow_id": followIDStr,
		"status":    "accepted",
	})
}

// RejectFollowRequest rejects a follow request
func (h *FollowHandler) RejectFollowRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	followIDStr := c.Param("followId")
	followID, err := primitive.ObjectIDFromHex(followIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid follow ID format", err)
		return
	}

	err = h.followService.RejectFollowRequest(followID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Follow request not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to reject follow request", err)
		return
	}

	utils.OkResponse(c, "Follow request rejected successfully", gin.H{
		"follow_id": followIDStr,
		"status":    "rejected",
	})
}

// CancelFollowRequest cancels a sent follow request
func (h *FollowHandler) CancelFollowRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	followIDStr := c.Param("followId")
	followID, err := primitive.ObjectIDFromHex(followIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid follow ID format", err)
		return
	}

	err = h.followService.CancelFollowRequest(followID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Follow request not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to cancel follow request", err)
		return
	}

	utils.OkResponse(c, "Follow request cancelled successfully", gin.H{
		"follow_id": followIDStr,
		"status":    "cancelled",
	})
}

// GetSuggestedUsers retrieves suggested users to follow
func (h *FollowHandler) GetSuggestedUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := utils.StringToInt(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	suggestions, err := h.followService.GetSuggestedUsers(userID.(primitive.ObjectID), limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get suggested users", err)
		return
	}

	utils.OkResponse(c, "Suggested users retrieved successfully", gin.H{
		"suggestions": suggestions,
		"count":       len(suggestions),
	})
}

// BulkFollowUsers follows multiple users at once
func (h *FollowHandler) BulkFollowUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.UserIDs) == 0 {
		utils.BadRequestResponse(c, "At least one user ID is required", nil)
		return
	}

	if len(req.UserIDs) > 50 {
		utils.BadRequestResponse(c, "Cannot follow more than 50 users at once", nil)
		return
	}

	results, err := h.followService.BulkFollowUsers(userID.(primitive.ObjectID), req.UserIDs)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to follow users", err)
		return
	}

	utils.OkResponse(c, "Bulk follow completed", results)
}

// GetFollowActivity retrieves recent follow activity
func (h *FollowHandler) GetFollowActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get activity type parameter
	activityType := c.DefaultQuery("type", "all")
	validTypes := []string{"all", "new_followers", "new_following", "follow_requests"}
	isValidType := false
	for _, validType := range validTypes {
		if activityType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		activityType = "all"
	}

	activity, err := h.followService.GetFollowActivity(userID.(primitive.ObjectID), activityType, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get follow activity", err)
		return
	}

	totalCount := int64(len(activity))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Follow activity retrieved successfully", activity, paginationMeta, nil)
}
