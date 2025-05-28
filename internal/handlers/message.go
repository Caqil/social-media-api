// internal/handlers/message.go
package handlers

import (
	"strings"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"
	"social-media-api/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageHandler struct {
	messageService      *services.MessageService
	conversationService *services.ConversationService
	hub                 *websocket.Hub
	validator           *validator.Validate
}

func NewMessageHandler(messageService *services.MessageService, conversationService *services.ConversationService, hub *websocket.Hub) *MessageHandler {
	return &MessageHandler{
		messageService:      messageService,
		conversationService: conversationService,
		hub:                 hub,
		validator:           validator.New(),
	}
}

// CreateConversation creates a new conversation
func (h *MessageHandler) CreateConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate participants
	if len(req.ParticipantIDs) == 0 {
		utils.BadRequestResponse(c, "At least one participant is required", nil)
		return
	}

	// Add current user to participants if not already included
	currentUserIDStr := userID.(primitive.ObjectID).Hex()
	found := false
	for _, participantID := range req.ParticipantIDs {
		if participantID == currentUserIDStr {
			found = true
			break
		}
	}
	if !found {
		req.ParticipantIDs = append(req.ParticipantIDs, currentUserIDStr)
	}

	conversation, err := h.conversationService.CreateConversation(userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.ConflictResponse(c, "Conversation already exists", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create conversation", err)
		return
	}

	utils.CreatedResponse(c, "Conversation created successfully", conversation.ToConversationResponse())
}

// GetConversations retrieves user's conversations
func (h *MessageHandler) GetConversations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	conversations, err := h.conversationService.GetUserConversations(userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get conversations", err)
		return
	}

	// Convert to response format
	var conversationResponses []models.ConversationResponse
	for _, conversation := range conversations {
		conversationResponses = append(conversationResponses, conversation.ToConversationResponse())
	}

	totalCount := int64(len(conversationResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Conversations retrieved successfully", conversationResponses, paginationMeta, nil)
}

// GetConversation retrieves a specific conversation
func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	conversation, err := h.conversationService.GetConversationByID(conversationID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get conversation", err)
		return
	}

	utils.OkResponse(c, "Conversation retrieved successfully", conversation.ToConversationResponse())
}

// SendMessage sends a message in a conversation
func (h *MessageHandler) SendMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("conversationId")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate content
	if strings.TrimSpace(req.Content) == "" && len(req.Media) == 0 {
		utils.BadRequestResponse(c, "Message content or media is required", nil)
		return
	}

	if len(req.Content) > utils.MaxMessageContentLength {
		utils.BadRequestResponse(c, "Message content exceeds maximum length", nil)
		return
	}

	message, err := h.messageService.SendMessage(userID.(primitive.ObjectID), conversationID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to send message", err)
		return
	}

	// Broadcast message via WebSocket
	go h.broadcastMessage(message)

	utils.CreatedResponse(c, "Message sent successfully", message.ToMessageResponse())
}

// GetMessages retrieves messages from a conversation
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("conversationId")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	messages, err := h.messageService.GetConversationMessages(conversationID, userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get messages", err)
		return
	}

	// Convert to response format
	var messageResponses []models.MessageResponse
	for _, message := range messages {
		messageResponses = append(messageResponses, message.ToMessageResponse())
	}

	totalCount := int64(len(messageResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Messages retrieved successfully", messageResponses, paginationMeta, nil)
}

// UpdateMessage updates a message
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID format", err)
		return
	}

	var req models.UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate content length if provided
	if req.Content != nil && len(*req.Content) > utils.MaxMessageContentLength {
		utils.BadRequestResponse(c, "Message content exceeds maximum length", nil)
		return
	}

	message, err := h.messageService.UpdateMessage(messageID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Message not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update message", err)
		return
	}

	// Broadcast message update via WebSocket
	go h.broadcastMessageUpdate(message)

	utils.OkResponse(c, "Message updated successfully", message.ToMessageResponse())
}

// DeleteMessage deletes a message
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID format", err)
		return
	}

	message, err := h.messageService.DeleteMessage(messageID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Message not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete message", err)
		return
	}

	// Broadcast message deletion via WebSocket
	go h.broadcastMessageDeletion(message)

	utils.OkResponse(c, "Message deleted successfully", nil)
}

// MarkMessagesAsRead marks messages as read
func (h *MessageHandler) MarkMessagesAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("conversationId")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	var req struct {
		LastMessageID string `json:"last_message_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	lastMessageID, err := primitive.ObjectIDFromHex(req.LastMessageID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID format", err)
		return
	}

	err = h.messageService.MarkMessagesAsRead(conversationID, userID.(primitive.ObjectID), lastMessageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to mark messages as read", err)
		return
	}

	// Broadcast read receipt via WebSocket
	go h.broadcastReadReceipt(conversationID, userID.(primitive.ObjectID), lastMessageID)

	utils.OkResponse(c, "Messages marked as read successfully", gin.H{
		"conversation_id": conversationIDStr,
		"last_message_id": req.LastMessageID,
	})
}

// SearchMessages searches messages in conversations
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	if len(strings.TrimSpace(query)) < 2 {
		utils.BadRequestResponse(c, "Search query must be at least 2 characters", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get conversation ID filter if provided
	var conversationID *primitive.ObjectID
	if conversationIDStr := c.Query("conversation_id"); conversationIDStr != "" {
		if id, err := primitive.ObjectIDFromHex(conversationIDStr); err == nil {
			conversationID = &id
		}
	}

	messages, err := h.messageService.SearchMessages(userID.(primitive.ObjectID), query, conversationID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search messages", err)
		return
	}

	// Convert to response format
	var messageResponses []models.MessageResponse
	for _, message := range messages {
		messageResponses = append(messageResponses, message.ToMessageResponse())
	}

	totalCount := int64(len(messageResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Message search completed successfully", messageResponses, paginationMeta, gin.H{
		"query": query,
	})
}

// LeaveConversation allows user to leave a group conversation
func (h *MessageHandler) LeaveConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	err = h.conversationService.LeaveConversation(conversationID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		if strings.Contains(err.Error(), "cannot leave") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to leave conversation", err)
		return
	}

	utils.OkResponse(c, "Left conversation successfully", nil)
}

// AddParticipants adds participants to a group conversation
func (h *MessageHandler) AddParticipants(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	var req struct {
		ParticipantIDs []string `json:"participant_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.ParticipantIDs) == 0 {
		utils.BadRequestResponse(c, "At least one participant ID is required", nil)
		return
	}

	err = h.conversationService.AddParticipants(conversationID, userID.(primitive.ObjectID), req.ParticipantIDs)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to add participants", err)
		return
	}

	utils.OkResponse(c, "Participants added successfully", gin.H{
		"added_count": len(req.ParticipantIDs),
	})
}

// RemoveParticipant removes a participant from a group conversation
func (h *MessageHandler) RemoveParticipant(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	participantIDStr := c.Param("participantId")
	participantID, err := primitive.ObjectIDFromHex(participantIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid participant ID format", err)
		return
	}

	err = h.conversationService.RemoveParticipant(conversationID, userID.(primitive.ObjectID), participantID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove participant", err)
		return
	}

	utils.OkResponse(c, "Participant removed successfully", nil)
}

// UpdateConversation updates conversation details
func (h *MessageHandler) UpdateConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := primitive.ObjectIDFromHex(conversationIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID format", err)
		return
	}

	var req models.UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	conversation, err := h.conversationService.UpdateConversation(conversationID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Conversation not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update conversation", err)
		return
	}

	utils.OkResponse(c, "Conversation updated successfully", conversation.ToConversationResponse())
}

// WebSocket broadcasting helper methods

func (h *MessageHandler) broadcastMessage(message *models.Message) {
	if h.hub == nil {
		return
	}

	wsMessage := websocket.WebSocketMessage{
		Type:   "message",
		Action: "new",
		Data: map[string]interface{}{
			"id":              message.ID.Hex(),
			"conversation_id": message.ConversationID.Hex(),
			"sender_id":       message.SenderID.Hex(),
			"content":         message.Content,
			"content_type":    message.ContentType,
			"media":           message.Media,
			"created_at":      message.CreatedAt,
		},
	}

	// Broadcast to conversation channel
	channel := "conversation:" + message.ConversationID.Hex()
	h.hub.BroadcastToChannel(channel, wsMessage, message.SenderID)
}

func (h *MessageHandler) broadcastMessageUpdate(message *models.Message) {
	if h.hub == nil {
		return
	}

	wsMessage := websocket.WebSocketMessage{
		Type:   "message",
		Action: "updated",
		Data: map[string]interface{}{
			"id":              message.ID.Hex(),
			"conversation_id": message.ConversationID.Hex(),
			"content":         message.Content,
			"is_edited":       message.IsEdited,
			"edited_at":       message.EditedAt,
		},
	}

	channel := "conversation:" + message.ConversationID.Hex()
	h.hub.BroadcastToChannel(channel, wsMessage, primitive.NilObjectID)
}

func (h *MessageHandler) broadcastMessageDeletion(message *models.Message) {
	if h.hub == nil {
		return
	}

	wsMessage := websocket.WebSocketMessage{
		Type:   "message",
		Action: "deleted",
		Data: map[string]interface{}{
			"id":              message.ID.Hex(),
			"conversation_id": message.ConversationID.Hex(),
			"deleted_at":      message.DeletedAt,
		},
	}

	channel := "conversation:" + message.ConversationID.Hex()
	h.hub.BroadcastToChannel(channel, wsMessage, primitive.NilObjectID)
}

func (h *MessageHandler) broadcastReadReceipt(conversationID, userID, lastMessageID primitive.ObjectID) {
	if h.hub == nil {
		return
	}

	wsMessage := websocket.WebSocketMessage{
		Type:   "message",
		Action: "read",
		Data: map[string]interface{}{
			"conversation_id": conversationID.Hex(),
			"user_id":         userID.Hex(),
			"last_message_id": lastMessageID.Hex(),
			"read_at":         time.Now(),
		},
	}

	channel := "conversation:" + conversationID.Hex()
	h.hub.BroadcastToChannel(channel, wsMessage, userID)
}
