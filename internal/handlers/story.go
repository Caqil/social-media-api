// internal/handlers/story.go
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

type StoryHandler struct {
	storyService *services.StoryService
	validator    *validator.Validate
}

func NewStoryHandler(storyService *services.StoryService) *StoryHandler {
	return &StoryHandler{
		storyService: storyService,
		validator:    validator.New(),
	}
}

// CreateStory creates a new story
func (h *StoryHandler) CreateStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate content
	if len(req.Media) == 0 && strings.TrimSpace(req.Content) == "" {
		utils.BadRequestResponse(c, "Story must have content or media", nil)
		return
	}

	story, err := h.storyService.CreateStory(userID.(primitive.ObjectID), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create story", err)
		return
	}

	utils.CreatedResponse(c, "Story created successfully", story.ToStoryResponse())
}

// GetStory retrieves a specific story
func (h *StoryHandler) GetStory(c *gin.Context) {
	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	story, err := h.storyService.GetStoryByID(storyID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get story", err)
		return
	}

	utils.OkResponse(c, "Story retrieved successfully", story.ToStoryResponse())
}

// GetUserStories retrieves stories from a specific user
func (h *StoryHandler) GetUserStories(c *gin.Context) {
	userIDStr := c.Param("userId")
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

	stories, err := h.storyService.GetUserStories(userID, currentUserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user stories", err)
		return
	}

	// Convert to response format
	var storyResponses []models.StoryResponse
	for _, story := range stories {
		storyResponses = append(storyResponses, story.ToStoryResponse())
	}

	utils.OkResponse(c, "User stories retrieved successfully", gin.H{
		"user_id": userIDStr,
		"stories": storyResponses,
		"count":   len(storyResponses),
	})
}

// GetFollowingStories retrieves stories from users that current user follows
func (h *StoryHandler) GetFollowingStories(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	stories, err := h.storyService.GetFollowingStories(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get following stories", err)
		return
	}

	// Group stories by user
	storyGroups := make(map[string][]models.StoryResponse)
	var userOrder []string

	for _, story := range stories {
		storyResponse := story.ToStoryResponse()
		userIDStr := story.UserID.Hex()

		if _, exists := storyGroups[userIDStr]; !exists {
			userOrder = append(userOrder, userIDStr)
			storyGroups[userIDStr] = []models.StoryResponse{}
		}

		storyGroups[userIDStr] = append(storyGroups[userIDStr], storyResponse)
	}

	// Create ordered response
	var orderedStories []gin.H
	for _, userIDStr := range userOrder {
		orderedStories = append(orderedStories, gin.H{
			"user_id": userIDStr,
			"stories": storyGroups[userIDStr],
			"count":   len(storyGroups[userIDStr]),
		})
	}

	utils.OkResponse(c, "Following stories retrieved successfully", gin.H{
		"story_groups": orderedStories,
		"total_count":  len(stories),
	})
}

// UpdateStory updates an existing story
func (h *StoryHandler) UpdateStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	var req models.UpdateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	story, err := h.storyService.UpdateStory(storyID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update story", err)
		return
	}

	utils.OkResponse(c, "Story updated successfully", story.ToStoryResponse())
}

// DeleteStory deletes a story
func (h *StoryHandler) DeleteStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	err = h.storyService.DeleteStory(storyID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete story", err)
		return
	}

	utils.OkResponse(c, "Story deleted successfully", nil)
}

// ViewStory marks a story as viewed by the current user
func (h *StoryHandler) ViewStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	err = h.storyService.ViewStory(storyID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Story not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to mark story as viewed", err)
		return
	}

	utils.OkResponse(c, "Story marked as viewed", gin.H{
		"story_id": storyIDStr,
		"viewed":   true,
	})
}

// GetStoryViews retrieves viewers of a story
func (h *StoryHandler) GetStoryViews(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	views, err := h.storyService.GetStoryViews(storyID, userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get story views", err)
		return
	}

	totalCount := int64(len(views))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Story views retrieved successfully", views, paginationMeta, gin.H{
		"story_id": storyIDStr,
	})
}

// ReactToStory adds a reaction to a story
func (h *StoryHandler) ReactToStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	var req struct {
		ReactionType models.ReactionType `json:"reaction_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	err = h.storyService.ReactToStory(storyID, userID.(primitive.ObjectID), req.ReactionType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Story not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to react to story", err)
		return
	}

	utils.OkResponse(c, "Reaction added to story successfully", gin.H{
		"story_id":      storyIDStr,
		"reaction_type": req.ReactionType,
	})
}

// UnreactToStory removes a reaction from a story
func (h *StoryHandler) UnreactToStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	err = h.storyService.UnreactToStory(storyID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Story not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove reaction from story", err)
		return
	}

	utils.OkResponse(c, "Reaction removed from story successfully", gin.H{
		"story_id": storyIDStr,
		"removed":  true,
	})
}

// GetStoryReactions retrieves reactions to a story
func (h *StoryHandler) GetStoryReactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	reactions, err := h.storyService.GetStoryReactions(storyID, userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get story reactions", err)
		return
	}

	totalCount := int64(len(reactions))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Story reactions retrieved successfully", reactions, paginationMeta, gin.H{
		"story_id": storyIDStr,
	})
}

// ReplyToStory creates a reply to a story
func (h *StoryHandler) ReplyToStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	var req models.CreateStoryReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	reply, err := h.storyService.ReplyToStory(storyID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Story not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to reply to story", err)
		return
	}

	utils.CreatedResponse(c, "Story reply created successfully", reply.ToStoryReplyResponse())
}

// GetStoryReplies retrieves replies to a story
func (h *StoryHandler) GetStoryReplies(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	replies, err := h.storyService.GetStoryReplies(storyID, userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get story replies", err)
		return
	}

	// Convert to response format
	var replyResponses []models.StoryReplyResponse
	for _, reply := range replies {
		replyResponses = append(replyResponses, reply.ToStoryReplyResponse())
	}

	totalCount := int64(len(replyResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Story replies retrieved successfully", replyResponses, paginationMeta, gin.H{
		"story_id": storyIDStr,
	})
}

// GetStoryStats retrieves story statistics
func (h *StoryHandler) GetStoryStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	stats, err := h.storyService.GetStoryStats(storyID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get story statistics", err)
		return
	}

	utils.OkResponse(c, "Story statistics retrieved successfully", stats)
}

// GetActiveStories retrieves currently active stories from all users
func (h *StoryHandler) GetActiveStories(c *gin.Context) {
	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	stories, err := h.storyService.GetActiveStories(currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get active stories", err)
		return
	}

	// Group stories by user
	storyGroups := make(map[string][]models.StoryResponse)
	userOrder := []string{}

	for _, story := range stories {
		storyResponse := story.ToStoryResponse()
		userIDStr := story.UserID.Hex()

		if _, exists := storyGroups[userIDStr]; !exists {
			userOrder = append(userOrder, userIDStr)
			storyGroups[userIDStr] = []models.StoryResponse{}
		}

		storyGroups[userIDStr] = append(storyGroups[userIDStr], storyResponse)
	}

	// Create ordered response
	var orderedStories []gin.H
	for _, userIDStr := range userOrder {
		orderedStories = append(orderedStories, gin.H{
			"user_id": userIDStr,
			"stories": storyGroups[userIDStr],
			"count":   len(storyGroups[userIDStr]),
		})
	}

	totalCount := int64(len(orderedStories))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Active stories retrieved successfully", orderedStories, paginationMeta, nil)
}

// ArchiveStory archives a story
func (h *StoryHandler) ArchiveStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	storyIDStr := c.Param("id")
	storyID, err := primitive.ObjectIDFromHex(storyIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID format", err)
		return
	}

	err = h.storyService.ArchiveStory(storyID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Story not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to archive story", err)
		return
	}

	utils.OkResponse(c, "Story archived successfully", gin.H{
		"story_id": storyIDStr,
		"archived": true,
	})
}

// GetArchivedStories retrieves user's archived stories
func (h *StoryHandler) GetArchivedStories(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	stories, err := h.storyService.GetArchivedStories(userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get archived stories", err)
		return
	}

	// Convert to response format
	var storyResponses []models.StoryResponse
	for _, story := range stories {
		storyResponses = append(storyResponses, story.ToStoryResponse())
	}

	totalCount := int64(len(storyResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Archived stories retrieved successfully", storyResponses, paginationMeta, nil)
}
