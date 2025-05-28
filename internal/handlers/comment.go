// internal/handlers/comment.go
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

type CommentHandler struct {
	commentService *services.CommentService
	validator      *validator.Validate
}

func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		validator:      validator.New(),
	}
}

// CreateComment creates a new comment
func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate content length
	if len(strings.TrimSpace(req.Content)) == 0 {
		utils.BadRequestResponse(c, "Comment content is required", nil)
		return
	}

	if len(req.Content) > utils.MaxCommentContentLength {
		utils.BadRequestResponse(c, "Comment content exceeds maximum length", nil)
		return
	}

	comment, err := h.commentService.CreateComment(userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Target post not found")
			return
		}
		if strings.Contains(err.Error(), "disabled") {
			utils.BadRequestResponse(c, "Comments are disabled for this post", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create comment", err)
		return
	}

	utils.CreatedResponse(c, "Comment created successfully", comment.ToCommentResponse())
}

// GetComment retrieves a single comment by ID
func (h *CommentHandler) GetComment(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	comment, err := h.commentService.GetCommentByID(commentID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get comment", err)
		return
	}

	utils.OkResponse(c, "Comment retrieved successfully", comment.ToCommentResponse())
}

// GetPostComments retrieves comments for a specific post
func (h *CommentHandler) GetPostComments(c *gin.Context) {
	postIDStr := c.Param("postId")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get sort parameter
	sortBy := c.DefaultQuery("sort", "newest")
	if !h.isValidSortBy(sortBy) {
		sortBy = "newest"
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	comments, err := h.commentService.GetPostComments(postID, currentUserID, sortBy, params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Post not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get comments", err)
		return
	}

	// Convert to response format
	var commentResponses []models.CommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, comment.ToCommentResponse())
	}

	totalCount := int64(len(commentResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Comments retrieved successfully", commentResponses, paginationMeta, nil)
}

// GetCommentReplies retrieves replies to a specific comment
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	replies, err := h.commentService.GetCommentReplies(commentID, currentUserID, params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get comment replies", err)
		return
	}

	// Convert to response format
	var replyResponses []models.CommentResponse
	for _, reply := range replies {
		replyResponses = append(replyResponses, reply.ToCommentResponse())
	}

	totalCount := int64(len(replyResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Comment replies retrieved successfully", replyResponses, paginationMeta, nil)
}

// UpdateComment updates an existing comment
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate content length if provided
	if req.Content != nil {
		if len(strings.TrimSpace(*req.Content)) == 0 {
			utils.BadRequestResponse(c, "Comment content cannot be empty", nil)
			return
		}
		if len(*req.Content) > utils.MaxCommentContentLength {
			utils.BadRequestResponse(c, "Comment content exceeds maximum length", nil)
			return
		}
	}

	comment, err := h.commentService.UpdateComment(commentID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Comment not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update comment", err)
		return
	}

	utils.OkResponse(c, "Comment updated successfully", comment.ToCommentResponse())
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	err = h.commentService.DeleteComment(commentID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Comment not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete comment", err)
		return
	}

	utils.OkResponse(c, "Comment deleted successfully", nil)
}

// LikeComment adds or removes a like from a comment
func (h *CommentHandler) LikeComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	var req struct {
		ReactionType models.ReactionType `json:"reaction_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to "like" if no reaction type provided
		req.ReactionType = models.ReactionLike
	}

	err = h.commentService.LikeComment(commentID, userID.(primitive.ObjectID), req.ReactionType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to like comment", err)
		return
	}

	utils.OkResponse(c, "Comment reaction added successfully", gin.H{
		"reaction_type": req.ReactionType,
		"action":        "added",
	})
}

// UnlikeComment removes a like from a comment
func (h *CommentHandler) UnlikeComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	err = h.commentService.UnlikeComment(commentID, userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to unlike comment", err)
		return
	}

	utils.OkResponse(c, "Comment reaction removed successfully", gin.H{
		"action": "removed",
	})
}

// GetCommentLikes retrieves users who liked a comment
func (h *CommentHandler) GetCommentLikes(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	likes, err := h.commentService.GetCommentLikes(commentID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get comment likes", err)
		return
	}

	totalCount := int64(len(likes))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Comment likes retrieved successfully", likes, paginationMeta, nil)
}

// ReportComment reports a comment
func (h *CommentHandler) ReportComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
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

	err = h.commentService.ReportComment(commentID, userID.(primitive.ObjectID), req.Reason, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to report comment", err)
		return
	}

	utils.OkResponse(c, "Comment reported successfully", gin.H{
		"reported": true,
		"reason":   req.Reason,
	})
}

// PinComment pins a comment (post author only)
func (h *CommentHandler) PinComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	err = h.commentService.PinComment(commentID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Only post authors can pin comments")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to pin comment", err)
		return
	}

	utils.OkResponse(c, "Comment pinned successfully", gin.H{
		"pinned": true,
	})
}

// UnpinComment unpins a comment (post author only)
func (h *CommentHandler) UnpinComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	err = h.commentService.UnpinComment(commentID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Only post authors can unpin comments")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to unpin comment", err)
		return
	}

	utils.OkResponse(c, "Comment unpinned successfully", gin.H{
		"pinned": false,
	})
}

// GetUserComments retrieves comments made by a specific user
func (h *CommentHandler) GetUserComments(c *gin.Context) {
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

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	comments, err := h.commentService.GetUserComments(userID, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user comments", err)
		return
	}

	// Convert to response format
	var commentResponses []models.CommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, comment.ToCommentResponse())
	}

	totalCount := int64(len(commentResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "User comments retrieved successfully", commentResponses, paginationMeta, nil)
}

// GetCommentThread retrieves a complete comment thread
func (h *CommentHandler) GetCommentThread(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	thread, err := h.commentService.GetCommentThread(commentID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Comment not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get comment thread", err)
		return
	}

	// Convert to response format
	var threadResponses []models.CommentResponse
	for _, comment := range thread {
		threadResponses = append(threadResponses, comment.ToCommentResponse())
	}

	utils.OkResponse(c, "Comment thread retrieved successfully", gin.H{
		"root_comment_id": commentIDStr,
		"thread":          threadResponses,
		"count":           len(threadResponses),
	})
}

// Helper methods for validation

func (h *CommentHandler) isValidSortBy(sortBy string) bool {
	validSorts := []string{"newest", "oldest", "popular", "controversial"}
	for _, s := range validSorts {
		if sortBy == s {
			return true
		}
	}
	return false
}
