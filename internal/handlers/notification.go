// internal/handlers/notification.go
package handlers

import (
	"strconv"
	"strings"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
	validator           *validator.Validate
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		validator:           validator.New(),
	}
}

// GetNotifications retrieves user's notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get unread only parameter
	unreadOnly := c.Query("unread_only") == "true"

	notifications, err := h.notificationService.GetUserNotifications(
		userID.(primitive.ObjectID),
		params.Limit,
		params.Offset,
		unreadOnly,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notifications", err)
		return
	}

	totalCount := int64(len(notifications))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Notifications retrieved successfully", notifications, paginationMeta, gin.H{
		"unread_only": unreadOnly,
	})
}

// GetNotificationStats retrieves notification statistics
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	stats, err := h.notificationService.GetNotificationStats(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notification statistics", err)
		return
	}

	utils.OkResponse(c, "Notification statistics retrieved successfully", stats)
}

// MarkAsRead marks notifications as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.NotificationIDs) == 0 {
		utils.BadRequestResponse(c, "At least one notification ID is required", nil)
		return
	}

	err := h.notificationService.MarkAsRead(userID.(primitive.ObjectID), req.NotificationIDs)
	if err != nil {
		if strings.Contains(err.Error(), "no valid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to mark notifications as read", err)
		return
	}

	utils.OkResponse(c, "Notifications marked as read successfully", gin.H{
		"marked_count": len(req.NotificationIDs),
	})
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	err := h.notificationService.MarkAllAsRead(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to mark all notifications as read", err)
		return
	}

	utils.OkResponse(c, "All notifications marked as read successfully", nil)
}

// DeleteNotifications deletes notifications
func (h *NotificationHandler) DeleteNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.NotificationIDs) == 0 {
		utils.BadRequestResponse(c, "At least one notification ID is required", nil)
		return
	}

	err := h.notificationService.DeleteNotifications(userID.(primitive.ObjectID), req.NotificationIDs)
	if err != nil {
		if strings.Contains(err.Error(), "no valid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete notifications", err)
		return
	}

	utils.OkResponse(c, "Notifications deleted successfully", gin.H{
		"deleted_count": len(req.NotificationIDs),
	})
}

// CreateNotification creates a new notification (admin/system use)
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Set actor ID to current user if not provided
	if req.ActorID == "" {
		req.ActorID = userID.(primitive.ObjectID).Hex()
	}

	notification, err := h.notificationService.CreateNotification(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create notification", err)
		return
	}

	utils.CreatedResponse(c, "Notification created successfully", notification.ToNotificationResponse())
}

// CreateBulkNotifications creates notifications for multiple users
func (h *NotificationHandler) CreateBulkNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.BulkCreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Set actor ID to current user if not provided
	if req.ActorID == "" {
		req.ActorID = userID.(primitive.ObjectID).Hex()
	}

	// Validate recipient count
	if len(req.RecipientIDs) == 0 {
		utils.BadRequestResponse(c, "At least one recipient ID is required", nil)
		return
	}

	if len(req.RecipientIDs) > utils.MaxBulkNotificationRecipients {
		utils.BadRequestResponse(c, "Too many recipients. Maximum allowed: "+strconv.Itoa(utils.MaxBulkNotificationRecipients), nil)
		return
	}

	err := h.notificationService.CreateBulkNotifications(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "no valid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create bulk notifications", err)
		return
	}

	utils.CreatedResponse(c, "Bulk notifications created successfully", gin.H{
		"recipient_count":   len(req.RecipientIDs),
		"notification_type": req.Type,
	})
}

// GetNotificationPreferences retrieves user's notification preferences
func (h *NotificationHandler) GetNotificationPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	preferences, err := h.notificationService.GetUserPreferences(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notification preferences", err)
		return
	}

	utils.OkResponse(c, "Notification preferences retrieved successfully", preferences)
}

// UpdateNotificationPreferences updates user's notification preferences
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var preferences models.NotificationPreferences
	if err := c.ShouldBindJSON(&preferences); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	err := h.notificationService.UpdateUserPreferences(userID.(primitive.ObjectID), preferences)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update notification preferences", err)
		return
	}

	utils.OkResponse(c, "Notification preferences updated successfully", preferences)
}

// NotifyLike creates a like notification
func (h *NotificationHandler) NotifyLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		RecipientID string `json:"recipient_id" binding:"required"`
		PostID      string `json:"post_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	recipientID, err := primitive.ObjectIDFromHex(req.RecipientID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid recipient ID format", err)
		return
	}

	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	err = h.notificationService.NotifyLike(userID.(primitive.ObjectID), recipientID, postID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create like notification", err)
		return
	}

	utils.OkResponse(c, "Like notification created successfully", nil)
}

// NotifyComment creates a comment notification
func (h *NotificationHandler) NotifyComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		RecipientID string `json:"recipient_id" binding:"required"`
		PostID      string `json:"post_id" binding:"required"`
		CommentID   string `json:"comment_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	recipientID, err := primitive.ObjectIDFromHex(req.RecipientID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid recipient ID format", err)
		return
	}

	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	commentID, err := primitive.ObjectIDFromHex(req.CommentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID format", err)
		return
	}

	err = h.notificationService.NotifyComment(userID.(primitive.ObjectID), recipientID, postID, commentID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create comment notification", err)
		return
	}

	utils.OkResponse(c, "Comment notification created successfully", nil)
}

// NotifyFollow creates a follow notification
func (h *NotificationHandler) NotifyFollow(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		RecipientID string `json:"recipient_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	recipientID, err := primitive.ObjectIDFromHex(req.RecipientID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid recipient ID format", err)
		return
	}

	err = h.notificationService.NotifyFollow(userID.(primitive.ObjectID), recipientID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create follow notification", err)
		return
	}

	utils.OkResponse(c, "Follow notification created successfully", nil)
}

// NotifyMention creates a mention notification
func (h *NotificationHandler) NotifyMention(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		RecipientID string `json:"recipient_id" binding:"required"`
		ContentID   string `json:"content_id" binding:"required"`
		ContentType string `json:"content_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	recipientID, err := primitive.ObjectIDFromHex(req.RecipientID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid recipient ID format", err)
		return
	}

	contentID, err := primitive.ObjectIDFromHex(req.ContentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid content ID format", err)
		return
	}

	// Validate content type
	if !h.isValidContentType(req.ContentType) {
		utils.BadRequestResponse(c, "Invalid content type", nil)
		return
	}

	err = h.notificationService.NotifyMention(userID.(primitive.ObjectID), recipientID, contentID, req.ContentType)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create mention notification", err)
		return
	}

	utils.OkResponse(c, "Mention notification created successfully", nil)
}

// NotifyMessage creates a message notification
func (h *NotificationHandler) NotifyMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		RecipientID    string `json:"recipient_id" binding:"required"`
		ConversationID string `json:"conversation_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	recipientID, err := primitive.ObjectIDFromHex(req.RecipientID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid recipient ID format", err)
		return
	}

	conversationID, err := primitive.ObjectIDFromHex(req.ConversationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	err = h.notificationService.NotifyMessage(userID.(primitive.ObjectID), recipientID, conversationID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create message notification", err)
		return
	}

	utils.OkResponse(c, "Message notification created successfully", nil)
}

// TestNotification sends a test notification (development/admin use)
func (h *NotificationHandler) TestNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	notificationReq := models.CreateNotificationRequest{
		RecipientID: userID.(primitive.ObjectID).Hex(),
		ActorID:     userID.(primitive.ObjectID).Hex(),
		Type:        models.NotificationType(req.Type),
		Title:       req.Title,
		Message:     req.Message,
		Priority:    "normal",
		SendViaPush: true,
	}

	if req.Type == "" {
		notificationReq.Type = models.NotificationMessage
	}

	notification, err := h.notificationService.CreateNotification(notificationReq)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create test notification", err)
		return
	}

	utils.OkResponse(c, "Test notification sent successfully", notification.ToNotificationResponse())
}

// Helper methods for validation

func (h *NotificationHandler) isValidContentType(contentType string) bool {
	validTypes := []string{"post", "comment", "story", "message"}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}
