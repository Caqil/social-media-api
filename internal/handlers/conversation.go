// internal/handlers/conversation_handler.go
package handlers

import (
	"strings"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationHandler struct {
	conversationService *services.ConversationService
	messageService      *services.MessageService
	notificationService *services.NotificationService
}

func NewConversationHandler(conversationService *services.ConversationService, messageService *services.MessageService, notificationService *services.NotificationService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		messageService:      messageService,
		notificationService: notificationService,
	}
}

// CreateConversation creates a new conversation
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req models.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Create conversation
	conversation, err := h.conversationService.CreateConversation(userObjectID, req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create conversation", err)
		return
	}

	// Convert to response
	response := conversation.ToConversationResponse()
	utils.CreatedResponse(c, "Conversation created successfully", response)
}

// GetUserConversations retrieves conversations for the authenticated user
func (h *ConversationHandler) GetUserConversations(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Get pagination parameters
	paginationParams := utils.GetPaginationParams(c)

	// Try to use the method with total count if available
	// Otherwise fall back to the basic method
	conversations, total, err := h.conversationService.GetUserConversationsWithTotal(userObjectID, paginationParams.Limit, paginationParams.Offset)
	if err != nil {
		// Fallback to basic method
		conversationsBasic, errBasic := h.conversationService.GetUserConversations(userObjectID, paginationParams.Limit, paginationParams.Offset)
		if errBasic != nil {
			utils.InternalServerErrorResponse(c, "Failed to get conversations", errBasic)
			return
		}
		conversations = conversationsBasic
		total = int64(len(conversations))
	}

	// Create paginated response
	pagination := utils.CreatePaginationMeta(paginationParams, total)
	utils.PaginatedSuccessResponse(c, "Conversations retrieved successfully", conversations, pagination, nil)
}

// GetConversation retrieves a specific conversation
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Get conversation - service returns *models.ConversationResponse, error
	conversation, err := h.conversationService.GetConversationByID(conversationID, userObjectID)
	if err != nil {
		if err.Error() == "conversation not found or access denied" {
			utils.NotFoundResponse(c, "Conversation not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get conversation", err)
		return
	}

	utils.OkResponse(c, "Conversation retrieved successfully", conversation)
}

// UpdateConversation updates conversation details
func (h *ConversationHandler) UpdateConversation(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	var req models.UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Update conversation - service returns *models.ConversationResponse, error
	conversation, err := h.conversationService.UpdateConversation(conversationID, userObjectID, req)
	if err != nil {
		if err.Error() == "admin privileges required" {
			utils.ForbiddenResponse(c, "Admin privileges required")
			return
		}
		if err.Error() == "conversation not found or access denied" {
			utils.NotFoundResponse(c, "Conversation not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update conversation", err)
		return
	}

	utils.OkResponse(c, "Conversation updated successfully", conversation)
}

// AddParticipants adds participants to a conversation
func (h *ConversationHandler) AddParticipants(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	var req models.AddParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Add participants - service expects models.AddParticipantsRequest
	err = h.conversationService.AddParticipants(conversationID, userObjectID, req)
	if err != nil {
		if err.Error() == "insufficient permissions to add members" {
			utils.ForbiddenResponse(c, "Insufficient permissions to add members")
			return
		}
		if err.Error() == "conversation not found or access denied" {
			utils.NotFoundResponse(c, "Conversation not found")
			return
		}
		utils.BadRequestResponse(c, "Failed to add participants", err)
		return
	}

	utils.OkResponse(c, "Participants added successfully", nil)
}

// RemoveParticipant removes a participant from a conversation
func (h *ConversationHandler) RemoveParticipant(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	// Get participant ID from URL parameter
	participantIDStr := c.Param("participant_id")
	participantID, err := primitive.ObjectIDFromHex(participantIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid participant ID", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Remove participant
	err = h.conversationService.RemoveParticipant(conversationID, userObjectID, participantID)
	if err != nil {
		if err.Error() == "admin privileges required to remove other participants" {
			utils.ForbiddenResponse(c, "Admin privileges required")
			return
		}
		if err.Error() == "cannot remove the last admin" {
			utils.BadRequestResponse(c, "Cannot remove the last admin", nil)
			return
		}
		utils.BadRequestResponse(c, "Failed to remove participant", err)
		return
	}

	utils.OkResponse(c, "Participant removed successfully", nil)
}

// LeaveConversation allows a user to leave a conversation
func (h *ConversationHandler) LeaveConversation(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Leave conversation
	err = h.conversationService.LeaveConversation(conversationID, userObjectID)
	if err != nil {
		if err.Error() == "cannot leave direct conversations" {
			utils.BadRequestResponse(c, "Cannot leave direct conversations", nil)
			return
		}
		if err.Error() == "cannot leave as the last admin. Transfer admin rights first" {
			utils.BadRequestResponse(c, "Cannot leave as the last admin", nil)
			return
		}
		utils.BadRequestResponse(c, "Failed to leave conversation", err)
		return
	}

	utils.OkResponse(c, "Left conversation successfully", nil)
}

// GetConversationMessages retrieves messages from a conversation
func (h *ConversationHandler) GetConversationMessages(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Get pagination parameters
	paginationParams := utils.GetPaginationParams(c)

	// Get messages - service returns []models.Message, error (not total count)
	messages, err := h.messageService.GetConversationMessages(conversationID, userObjectID, paginationParams.Limit, paginationParams.Offset)
	if err != nil {
		if err.Error() == "access denied: user not in conversation" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get messages", err)
		return
	}

	// Convert to response format
	var responses []models.MessageResponse
	for _, msg := range messages {
		responses = append(responses, msg.ToMessageResponse())
	}

	// Create paginated response - using returned count as total
	total := int64(len(responses))
	pagination := utils.CreatePaginationMeta(paginationParams, total)
	utils.PaginatedSuccessResponse(c, "Messages retrieved successfully", responses, pagination, nil)
}

// SendMessage sends a message in a conversation
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Set conversation ID from URL parameter
	req.ConversationID = conversationIDStr

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Send message - service returns *models.Message, error
	message, err := h.messageService.SendMessage(userObjectID, conversationID, req)
	if err != nil {
		if err.Error() == "access denied: user not in conversation" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.BadRequestResponse(c, "Failed to send message", err)
		return
	}

	// Send notifications to other participants
	go h.notifyConversationParticipants(conversationID, userObjectID, "message")

	response := message.ToMessageResponse()
	utils.CreatedResponse(c, "Message sent successfully", response)
}

// MarkAsRead marks messages as read
func (h *ConversationHandler) MarkAsRead(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	var req struct {
		LastMessageID string `json:"last_message_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	lastMessageID, err := primitive.ObjectIDFromHex(req.LastMessageID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Mark messages as read
	err = h.messageService.MarkMessagesAsRead(conversationID, userObjectID, lastMessageID)
	if err != nil {
		if err.Error() == "access denied: user not in conversation" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to mark messages as read", err)
		return
	}

	utils.OkResponse(c, "Messages marked as read", nil)
}

// GetConversationStats returns conversation statistics
func (h *ConversationHandler) GetConversationStats(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Get conversation stats
	stats, err := h.conversationService.GetConversationStats(conversationID, userObjectID)
	if err != nil {
		if err.Error() == "conversation not found or access denied" {
			utils.NotFoundResponse(c, "Conversation not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get conversation stats", err)
		return
	}

	utils.OkResponse(c, "Conversation stats retrieved successfully", stats)
}

// SearchConversations searches user's conversations
func (h *ConversationHandler) SearchConversations(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Get pagination parameters
	paginationParams := utils.GetPaginationParams(c)

	// Try to use the search method if available
	conversations, total, err := h.conversationService.SearchUserConversations(userObjectID, query, paginationParams.Limit, paginationParams.Offset)
	if err != nil {
		// Fallback to basic search with client-side filtering
		allConversations, errBasic := h.conversationService.GetUserConversations(userObjectID, paginationParams.Limit, paginationParams.Offset)
		if errBasic != nil {
			utils.InternalServerErrorResponse(c, "Failed to search conversations", errBasic)
			return
		}

		// Simple client-side filtering (should be done in service for better performance)
		var filteredConversations []models.ConversationResponse
		for _, conv := range allConversations {
			if conv.Title != "" && len(conv.Title) > 0 {
				// Simple case-insensitive search
				if len(conv.Title) > 0 && strings.Contains(strings.ToLower(conv.Title), strings.ToLower(query)) {
					filteredConversations = append(filteredConversations, conv)
				}
			}
		}
		conversations = filteredConversations
		total = int64(len(filteredConversations))
	}

	// Create paginated response
	pagination := utils.CreatePaginationMeta(paginationParams, total)
	utils.PaginatedSuccessResponse(c, "Conversations found", conversations, pagination, nil)
}

// MuteConversation mutes/unmutes a conversation
func (h *ConversationHandler) MuteConversation(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	var req struct {
		Muted     bool       `json:"muted"`
		MuteUntil *time.Time `json:"mute_until,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Mute/unmute conversation
	err = h.conversationService.MuteConversation(conversationID, userObjectID, req.Muted, req.MuteUntil)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update mute status", err)
		return
	}

	message := "Conversation muted successfully"
	if !req.Muted {
		message = "Conversation unmuted successfully"
	}

	utils.OkResponse(c, message, nil)
}

// UpdateParticipantRole updates a participant's role in the conversation
func (h *ConversationHandler) UpdateParticipantRole(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	// Get participant ID from URL parameter
	participantIDStr := c.Param("participant_id")
	participantID, err := primitive.ObjectIDFromHex(participantIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid participant ID", err)
		return
	}

	var req models.UpdateParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Update participant role
	err = h.conversationService.UpdateParticipantRole(conversationID, userObjectID, participantID, req)
	if err != nil {
		if err.Error() == "admin privileges required" {
			utils.ForbiddenResponse(c, "Admin privileges required")
			return
		}
		if err.Error() == "conversation not found or access denied" {
			utils.NotFoundResponse(c, "Conversation not found")
			return
		}
		utils.BadRequestResponse(c, "Failed to update participant role", err)
		return
	}

	utils.OkResponse(c, "Participant role updated successfully", nil)
}

// GetUnreadCounts returns unread message counts for all conversations
func (h *ConversationHandler) GetUnreadCounts(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Get unread counts
	counts, err := h.conversationService.GetUnreadCounts(userObjectID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get unread counts", err)
		return
	}

	utils.OkResponse(c, "Unread counts retrieved successfully", counts)
}

// ArchiveConversation archives/unarchives a conversation
func (h *ConversationHandler) ArchiveConversation(c *gin.Context) {
	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", err)
		return
	}

	var req struct {
		Archived bool `json:"archived"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userObjectID := userID.(primitive.ObjectID)

	// Archive/unarchive conversation
	err = h.conversationService.ArchiveConversation(conversationID, userObjectID, req.Archived)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update archive status", err)
		return
	}

	message := "Conversation archived successfully"
	if !req.Archived {
		message = "Conversation unarchived successfully"
	}

	utils.OkResponse(c, message, nil)
}

// Helper method to notify conversation participants
func (h *ConversationHandler) notifyConversationParticipants(conversationID, senderID primitive.ObjectID, notificationType string) {
	// Get conversation to find participants
	conversation, err := h.conversationService.GetConversationByID(conversationID, senderID)
	if err != nil {
		return
	}

	// Send notifications to all participants except sender
	for _, participant := range conversation.ParticipantInfo {
		if participant.UserID != senderID {
			switch notificationType {
			case "message":
				h.notificationService.NotifyMessage(senderID, participant.UserID, conversationID)
			}
		}
	}
}
