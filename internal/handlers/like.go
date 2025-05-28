// internal/handlers/like.go
package handlers

import (
	"strings"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LikeHandler struct {
	likeService *services.LikeService
	validator   *validator.Validate
}

func NewLikeHandler(likeService *services.LikeService) *LikeHandler {
	return &LikeHandler{
		likeService: likeService,
		validator:   validator.New(),
	}
}

// CreateLike handles adding a like/reaction to content
func (h *LikeHandler) CreateLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate reaction type
	if !models.IsValidReactionType(req.ReactionType) {
		utils.BadRequestResponse(c, "Invalid reaction type", nil)
		return
	}

	// Validate target type
	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(req.TargetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	like, err := h.likeService.CreateLike(userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not accessible") {
			utils.NotFoundResponse(c, "Target content not found or not accessible")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to add reaction", err)
		return
	}

	utils.SuccessResponse(c, 201, "Reaction added successfully", gin.H{
		"like":        like.ToLikeResponse(),
		"reaction":    like.ReactionType,
		"emoji":       like.GetReactionEmoji(),
		"action":      "created",
		"target_id":   req.TargetID,
		"target_type": req.TargetType,
	})
}

// UpdateLike handles updating the reaction type of an existing like
func (h *LikeHandler) UpdateLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	likeIDStr := c.Param("id")
	likeID, err := primitive.ObjectIDFromHex(likeIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid like ID format", err)
		return
	}

	var req models.UpdateLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate reaction type
	if !models.IsValidReactionType(req.ReactionType) {
		utils.BadRequestResponse(c, "Invalid reaction type", nil)
		return
	}

	like, err := h.likeService.UpdateLike(likeID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Like not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update reaction", err)
		return
	}

	utils.OkResponse(c, "Reaction updated successfully", gin.H{
		"like":     like.ToLikeResponse(),
		"reaction": like.ReactionType,
		"emoji":    like.GetReactionEmoji(),
		"action":   "updated",
	})
}

// DeleteLike handles removing a like/reaction
func (h *LikeHandler) DeleteLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	targetIDStr := c.Param("targetId")
	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid target ID format", err)
		return
	}

	targetType := c.Param("targetType")
	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(targetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	err = h.likeService.DeleteLike(targetID, userID.(primitive.ObjectID), targetType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Like not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove reaction", err)
		return
	}

	utils.OkResponse(c, "Reaction removed successfully", gin.H{
		"action":      "removed",
		"target_id":   targetIDStr,
		"target_type": targetType,
	})
}

// GetLikes retrieves users who liked/reacted to content
func (h *LikeHandler) GetLikes(c *gin.Context) {
	targetIDStr := c.Param("targetId")
	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid target ID format", err)
		return
	}

	targetType := c.Param("targetType")
	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(targetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Optional reaction type filter
	var reactionType *models.ReactionType
	if reactionStr := c.Query("reaction"); reactionStr != "" {
		reaction := models.ReactionType(reactionStr)
		if models.IsValidReactionType(reaction) {
			reactionType = &reaction
		}
	}

	likes, err := h.likeService.GetLikes(targetID, targetType, reactionType, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get likes", err)
		return
	}

	totalCount := int64(len(likes))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	responseData := gin.H{
		"likes":       likes,
		"target_id":   targetIDStr,
		"target_type": targetType,
	}

	if reactionType != nil {
		responseData["reaction_filter"] = *reactionType
	}

	utils.PaginatedSuccessResponse(c, "Likes retrieved successfully", responseData, paginationMeta, nil)
}

// GetReactionSummary gets aggregated reaction statistics for content
func (h *LikeHandler) GetReactionSummary(c *gin.Context) {
	targetIDStr := c.Param("targetId")
	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid target ID format", err)
		return
	}

	targetType := c.Param("targetType")
	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(targetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	summary, err := h.likeService.GetReactionSummary(targetID, targetType, currentUserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get reaction summary", err)
		return
	}

	utils.OkResponse(c, "Reaction summary retrieved successfully", summary)
}

// GetDetailedReactionStats gets detailed reaction statistics
func (h *LikeHandler) GetDetailedReactionStats(c *gin.Context) {
	targetIDStr := c.Param("targetId")
	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid target ID format", err)
		return
	}

	targetType := c.Param("targetType")
	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(targetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	stats, err := h.likeService.GetDetailedReactionStats(targetID, targetType)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get detailed reaction stats", err)
		return
	}

	utils.OkResponse(c, "Detailed reaction statistics retrieved successfully", gin.H{
		"target_id":   targetIDStr,
		"target_type": targetType,
		"stats":       stats,
	})
}

// GetUserLikes retrieves likes made by a specific user
func (h *LikeHandler) GetUserLikes(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Optional target type filter
	var targetType *string
	if typeStr := c.Query("type"); typeStr != "" {
		validTargetTypes := []string{"post", "comment", "story", "message"}
		if h.isValidTargetType(typeStr, validTargetTypes) {
			targetType = &typeStr
		}
	}

	likes, err := h.likeService.GetUserLikes(userID, targetType, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user likes", err)
		return
	}

	totalCount := int64(len(likes))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	responseData := gin.H{
		"likes":   likes,
		"user_id": userIDStr,
	}

	if targetType != nil {
		responseData["type_filter"] = *targetType
	}

	utils.PaginatedSuccessResponse(c, "User likes retrieved successfully", responseData, paginationMeta, nil)
}

// CheckUserReaction checks if current user has reacted to content
func (h *LikeHandler) CheckUserReaction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	targetIDStr := c.Param("targetId")
	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid target ID format", err)
		return
	}

	targetType := c.Param("targetType")
	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(targetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	reactionType, err := h.likeService.CheckUserReaction(targetID, userID.(primitive.ObjectID), targetType)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to check user reaction", err)
		return
	}

	responseData := gin.H{
		"target_id":   targetIDStr,
		"target_type": targetType,
		"has_reacted": reactionType != nil,
	}

	if reactionType != nil {
		responseData["reaction_type"] = *reactionType
		responseData["emoji"] = models.GetReactionEmoji(*reactionType)
		responseData["name"] = models.GetReactionName(*reactionType)
	}

	utils.OkResponse(c, "User reaction status retrieved successfully", responseData)
}

// GetTrendingReactions gets trending reactions across platform
func (h *LikeHandler) GetTrendingReactions(c *gin.Context) {
	// Optional target type filter
	var targetType *string
	if typeStr := c.Query("type"); typeStr != "" {
		validTargetTypes := []string{"post", "comment", "story", "message"}
		if h.isValidTargetType(typeStr, validTargetTypes) {
			targetType = &typeStr
		}
	}

	// Time range parameter
	timeRangeStr := c.DefaultQuery("time_range", "24h")
	var timeRange time.Duration
	switch timeRangeStr {
	case "1h":
		timeRange = 1 * time.Hour
	case "24h":
		timeRange = 24 * time.Hour
	case "7d":
		timeRange = 7 * 24 * time.Hour
	case "30d":
		timeRange = 30 * 24 * time.Hour
	default:
		timeRange = 24 * time.Hour
	}

	// Limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := utils.StringToInt(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	reactions, err := h.likeService.GetTrendingReactions(targetType, timeRange, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending reactions", err)
		return
	}

	responseData := gin.H{
		"reactions":  reactions,
		"time_range": timeRangeStr,
		"period":     timeRange.String(),
	}

	if targetType != nil {
		responseData["type_filter"] = *targetType
	}

	utils.OkResponse(c, "Trending reactions retrieved successfully", responseData)
}

// GetReactionTypes returns all available reaction types
func (h *LikeHandler) GetReactionTypes(c *gin.Context) {
	reactionTypes := models.GetAllReactionTypes()

	var reactions []gin.H
	for _, reactionType := range reactionTypes {
		reactions = append(reactions, gin.H{
			"type":  reactionType,
			"name":  models.GetReactionName(reactionType),
			"emoji": models.GetReactionEmoji(reactionType),
		})
	}

	utils.OkResponse(c, "Available reaction types retrieved successfully", gin.H{
		"reactions": reactions,
		"total":     len(reactions),
	})
}

// GetMyReactions gets current user's recent reactions
func (h *LikeHandler) GetMyReactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Optional target type filter
	var targetType *string
	if typeStr := c.Query("type"); typeStr != "" {
		validTargetTypes := []string{"post", "comment", "story", "message"}
		if h.isValidTargetType(typeStr, validTargetTypes) {
			targetType = &typeStr
		}
	}

	likes, err := h.likeService.GetUserLikes(userID.(primitive.ObjectID), targetType, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get your reactions", err)
		return
	}

	totalCount := int64(len(likes))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	responseData := gin.H{
		"reactions": likes,
	}

	if targetType != nil {
		responseData["type_filter"] = *targetType
	}

	utils.PaginatedSuccessResponse(c, "Your reactions retrieved successfully", responseData, paginationMeta, nil)
}

// BulkReaction handles bulk reaction operations (admin feature)
func (h *LikeHandler) BulkReaction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user has admin privileges (implement based on your auth system)
	// This is a placeholder - implement actual admin check
	var req struct {
		TargetIDs    []string            `json:"target_ids" binding:"required"`
		TargetType   string              `json:"target_type" binding:"required"`
		ReactionType models.ReactionType `json:"reaction_type" binding:"required"`
		Action       string              `json:"action" binding:"required"` // add, remove
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.TargetIDs) > 100 {
		utils.BadRequestResponse(c, "Too many targets (max 100)", nil)
		return
	}

	validTargetTypes := []string{"post", "comment", "story", "message"}
	if !h.isValidTargetType(req.TargetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	if req.Action != "add" && req.Action != "remove" {
		utils.BadRequestResponse(c, "Invalid action. Must be 'add' or 'remove'", nil)
		return
	}

	var successCount, errorCount int
	var errors []string

	for _, targetIDStr := range req.TargetIDs {
		targetID, err := primitive.ObjectIDFromHex(targetIDStr)
		if err != nil {
			errorCount++
			errors = append(errors, "Invalid target ID: "+targetIDStr)
			continue
		}

		if req.Action == "add" {
			createReq := models.CreateLikeRequest{
				TargetID:     targetIDStr,
				TargetType:   req.TargetType,
				ReactionType: req.ReactionType,
			}
			_, err = h.likeService.CreateLike(userID.(primitive.ObjectID), createReq)
		} else {
			err = h.likeService.DeleteLike(targetID, userID.(primitive.ObjectID), req.TargetType)
		}

		if err != nil {
			errorCount++
			errors = append(errors, "Failed to process "+targetIDStr+": "+err.Error())
		} else {
			successCount++
		}
	}

	utils.OkResponse(c, "Bulk reaction operation completed", gin.H{
		"success_count": successCount,
		"error_count":   errorCount,
		"total":         len(req.TargetIDs),
		"errors":        errors,
		"action":        req.Action,
		"reaction_type": req.ReactionType,
	})
}

// Helper methods

// isValidTargetType checks if target type is valid
func (h *LikeHandler) isValidTargetType(targetType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if targetType == validType {
			return true
		}
	}
	return false
}

// Additional utility endpoints

// GetLikeStats gets overall like statistics for the platform (admin endpoint)
func (h *LikeHandler) GetLikeStats(c *gin.Context) {
	// This would be an admin-only endpoint
	// Implementation would depend on your requirements for platform statistics

	utils.OkResponse(c, "Like statistics endpoint placeholder", gin.H{
		"message": "This endpoint would provide platform-wide like statistics",
		"note":    "Implement based on specific analytics requirements",
	})
}
