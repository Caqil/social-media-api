// internal/handlers/post.go
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

type PostHandler struct {
	postService *services.PostService
	validator   *validator.Validate
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
		validator:   validator.New(),
	}
}

// CreatePost handles post creation
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate content length
	if len(req.Content) > utils.MaxPostContentLength {
		utils.BadRequestResponse(c, "Post content exceeds maximum length", nil)
		return
	}

	post, err := h.postService.CreatePost(userID.(primitive.ObjectID), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create post", err)
		return
	}

	utils.CreatedResponse(c, "Post created successfully", post.ToPostResponse())
}

// GetPost retrieves a single post by ID
func (h *PostHandler) GetPost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	post, err := h.postService.GetPostByID(postID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Post not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get post", err)
		return
	}

	utils.OkResponse(c, "Post retrieved successfully", post.ToPostResponse())
}

// GetUserPosts retrieves posts by a specific user
func (h *PostHandler) GetUserPosts(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		currentUserID = &id
	}

	posts, err := h.postService.GetUserPosts(userID, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user posts", err)
		return
	}

	// Convert to response format
	var postResponses []models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToPostResponse())
	}

	// Create pagination meta (you'd need to get total count for accurate pagination)
	totalCount := int64(len(postResponses)) // This is a simplified approach
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "User posts retrieved successfully", postResponses, paginationMeta, nil)
}

// GetFeed retrieves user's personalized feed
func (h *PostHandler) GetFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	posts, err := h.postService.GetFeedPosts(userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get feed", err)
		return
	}

	totalCount := int64(len(posts))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Feed retrieved successfully", posts, paginationMeta, nil)
}

// UpdatePost handles post updates
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate content length if provided
	if req.Content != nil && len(*req.Content) > utils.MaxPostContentLength {
		utils.BadRequestResponse(c, "Post content exceeds maximum length", nil)
		return
	}

	post, err := h.postService.UpdatePost(postID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Post not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update post", err)
		return
	}

	utils.OkResponse(c, "Post updated successfully", post.ToPostResponse())
}

// DeletePost handles post deletion
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	err = h.postService.DeletePost(postID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Post not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete post", err)
		return
	}

	utils.OkResponse(c, "Post deleted successfully", nil)
}

// LikePost handles post likes
func (h *PostHandler) LikePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	var req struct {
		ReactionType models.ReactionType `json:"reaction_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to "like" if no reaction type provided
		req.ReactionType = models.ReactionLike
	}

	err = h.postService.LikePost(postID, userID.(primitive.ObjectID), req.ReactionType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Post not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to like post", err)
		return
	}

	utils.OkResponse(c, "Post reaction added successfully", gin.H{
		"reaction_type": req.ReactionType,
		"action":        "added",
	})
}

// UnlikePost handles post unlikes
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	err = h.postService.UnlikePost(postID, userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to unlike post", err)
		return
	}

	utils.OkResponse(c, "Post reaction removed successfully", gin.H{
		"action": "removed",
	})
}

// GetPostLikes retrieves users who liked a post
func (h *PostHandler) GetPostLikes(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	likes, err := h.postService.GetPostLikes(postID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get post likes", err)
		return
	}

	totalCount := int64(len(likes))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Post likes retrieved successfully", likes, paginationMeta, nil)
}

// ReportPost handles post reporting
func (h *PostHandler) ReportPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	var req struct {
		Reason      models.ReportReason `json:"reason" binding:"required"`
		Description string              `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	err = h.postService.ReportPost(postID, userID.(primitive.ObjectID), req.Reason, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Post not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to report post", err)
		return
	}

	utils.OkResponse(c, "Post reported successfully", gin.H{
		"reported": true,
		"reason":   req.Reason,
	})
}

// GetPostStats retrieves post statistics
func (h *PostHandler) GetPostStats(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	stats, err := h.postService.GetPostStats(postID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Post not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get post stats", err)
		return
	}

	utils.OkResponse(c, "Post statistics retrieved successfully", stats)
}

// SearchPosts handles post search
func (h *PostHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	posts, err := h.postService.SearchPosts(query, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search posts", err)
		return
	}

	// Convert to response format
	var postResponses []models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToPostResponse())
	}

	totalCount := int64(len(postResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Search results retrieved successfully", postResponses, paginationMeta, nil)
}

// GetTrendingPosts retrieves trending posts
func (h *PostHandler) GetTrendingPosts(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "day")
	if timeRange != "hour" && timeRange != "day" && timeRange != "week" {
		timeRange = "day"
	}

	posts, err := h.postService.GetTrendingPosts(params.Limit, params.Offset, timeRange)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending posts", err)
		return
	}

	// Convert to response format
	var postResponses []models.PostResponse
	for _, post := range posts {
		postResponses = append(postResponses, post.ToPostResponse())
	}

	totalCount := int64(len(postResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Trending posts retrieved successfully", postResponses, paginationMeta, nil)
}

// PinPost pins a post to user's profile
func (h *PostHandler) PinPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	// Update post to set as pinned
	isPinned := true
	req := models.UpdatePostRequest{
		IsPinned: &isPinned,
	}

	_, err = h.postService.UpdatePost(postID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Post not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to pin post", err)
		return
	}

	utils.OkResponse(c, "Post pinned successfully", gin.H{
		"pinned": true,
	})
}

// UnpinPost unpins a post from user's profile
func (h *PostHandler) UnpinPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	// Update post to unset as pinned
	isPinned := false
	req := models.UpdatePostRequest{
		IsPinned: &isPinned,
	}

	_, err = h.postService.UpdatePost(postID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Post not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to unpin post", err)
		return
	}

	utils.OkResponse(c, "Post unpinned successfully", gin.H{
		"pinned": false,
	})
}
