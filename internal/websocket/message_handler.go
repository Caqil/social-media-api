// message_handler.go
package websocket

import (
	"context"
	"fmt"
	"log"
	"time"

	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MessageHandlerInterface defines the interface for message handling
type MessageHandlerInterface interface {
	HandleMessage(client *Client, message WebSocketMessage) error
}

// MessageHandler handles real-time messaging functionality
type MessageHandler struct {
	db                *mongo.Database
	conversationsColl *mongo.Collection
	messagesColl      *mongo.Collection
	hub               *Hub
	typingIndicators  map[string]map[primitive.ObjectID]time.Time // conversationID -> userID -> lastTypingTime
	typingMutex       map[string]*typingMutex
}

type typingMutex struct {
	indicators map[primitive.ObjectID]time.Time
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(db *mongo.Database, hub *Hub) *MessageHandler {
	return &MessageHandler{
		db:                db,
		conversationsColl: db.Collection("conversations"),
		messagesColl:      db.Collection("messages"),
		hub:               hub,
		typingIndicators:  make(map[string]map[primitive.ObjectID]time.Time),
		typingMutex:       make(map[string]*typingMutex),
	}
}

// HandleMessage handles incoming WebSocket messages related to messaging
func (h *MessageHandler) HandleMessage(client *Client, message WebSocketMessage) error {
	switch message.Action {
	case "send":
		return h.handleSendMessage(client, message)
	case "edit":
		return h.handleEditMessage(client, message)
	case "delete":
		return h.handleDeleteMessage(client, message)
	case "react":
		return h.handleReactToMessage(client, message)
	case "mark_read":
		return h.handleMarkAsRead(client, message)
	case "typing_start":
		return h.handleTypingStart(client, message)
	case "typing_stop":
		return h.handleTypingStop(client, message)
	case "get_messages":
		return h.handleGetMessages(client, message)
	case "search":
		return h.handleSearchMessages(client, message)
	case "forward":
		return h.handleForwardMessage(client, message)
	default:
		return fmt.Errorf("unknown message action: %s", message.Action)
	}
}

// handleSendMessage handles sending a new message
func (h *MessageHandler) handleSendMessage(client *Client, wsMessage WebSocketMessage) error {
	// Parse message data
	conversationID, ok := wsMessage.Data["conversation_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing conversation_id")
	}

	content, ok := wsMessage.Data["content"].(string)
	if !ok || content == "" {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing content")
	}

	contentType, ok := wsMessage.Data["content_type"].(string)
	if !ok {
		contentType = "text"
	}

	// Validate conversation ID
	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_CONVERSATION", "Invalid conversation ID")
	}

	// Check if user is participant in the conversation
	if !h.isUserInConversation(client.UserID, conversationObjectID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "User not in conversation")
	}

	// Create message
	newMessage := models.Message{
		ConversationID: conversationObjectID,
		SenderID:       client.UserID,
		Content:        content,
		ContentType:    models.ContentType(contentType),
		Status:         models.MessageSent,
		Source:         "websocket",
	}

	// Handle media if present
	if mediaData, exists := wsMessage.Data["media"]; exists {
		if mediaArray, ok := mediaData.([]interface{}); ok {
			media := make([]models.MediaInfo, 0, len(mediaArray))
			for _, item := range mediaArray {
				if mediaMap, ok := item.(map[string]interface{}); ok {
					mediaInfo := models.MediaInfo{
						URL:  getStringFromMap(mediaMap, "url"),
						Type: getStringFromMap(mediaMap, "type"),
					}
					media = append(media, mediaInfo)
				}
			}
			newMessage.Media = media
		}
	}

	// Handle reply
	if replyToID, exists := wsMessage.Data["reply_to_message_id"].(string); exists && replyToID != "" {
		if replyObjectID, err := primitive.ObjectIDFromHex(replyToID); err == nil {
			newMessage.ReplyToMessageID = &replyObjectID
		}
	}

	// Set timestamps
	newMessage.BeforeCreate()
	now := time.Now()
	newMessage.SentAt = &now

	// Insert message into database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.messagesColl.InsertOne(ctx, newMessage)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to save message")
	}

	newMessage.ID = result.InsertedID.(primitive.ObjectID)

	// Update conversation last message
	if err := h.updateConversationLastMessage(conversationObjectID, newMessage); err != nil {
		log.Printf("Failed to update conversation last message: %v", err)
	}

	// Create WebSocket message for broadcast
	broadcastMessage := WebSocketMessage{
		Type:    "message",
		Action:  "new",
		Channel: "conversation:" + conversationID,
		Data: map[string]interface{}{
			"id":              newMessage.ID.Hex(),
			"conversation_id": conversationID,
			"sender_id":       client.UserID.Hex(),
			"sender": map[string]interface{}{
				"user_id":  client.UserID.Hex(),
				"username": client.Username,
			},
			"content":      newMessage.Content,
			"content_type": newMessage.ContentType,
			"media":        newMessage.Media,
			"status":       newMessage.Status,
			"sent_at":      newMessage.SentAt,
			"created_at":   newMessage.CreatedAt,
		},
	}

	// Add reply info if present
	if newMessage.ReplyToMessageID != nil {
		broadcastMessage.Data["reply_to_message_id"] = newMessage.ReplyToMessageID.Hex()
	}

	// Broadcast to conversation participants
	h.broadcastToConversationParticipants(conversationObjectID, broadcastMessage, client.UserID)

	// Send confirmation to sender
	confirmMessage := WebSocketMessage{
		Type:      "message",
		Action:    "sent",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"message_id":      newMessage.ID.Hex(),
			"conversation_id": conversationID,
			"status":          "sent",
			"sent_at":         newMessage.SentAt,
		},
	}
	client.SendMessage(confirmMessage)

	// Clear typing indicator for sender
	h.clearTypingIndicator(conversationID, client.UserID)

	return nil
}

// handleEditMessage handles editing an existing message
func (h *MessageHandler) handleEditMessage(client *Client, wsMessage WebSocketMessage) error {
	messageID, ok := wsMessage.Data["message_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing message_id")
	}

	newContent, ok := wsMessage.Data["content"].(string)
	if !ok || newContent == "" {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing content")
	}

	// Validate message ID
	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_MESSAGE", "Invalid message ID")
	}

	// Get message from database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message models.Message
	err = h.messagesColl.FindOne(ctx, bson.M{
		"_id":        messageObjectID,
		"sender_id":  client.UserID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return h.sendError(client, wsMessage.RequestID, "MESSAGE_NOT_FOUND", "Message not found or unauthorized")
		}
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to fetch message")
	}

	// Update message
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"content":    newContent,
			"is_edited":  true,
			"edited_at":  now,
			"updated_at": now,
		},
	}

	_, err = h.messagesColl.UpdateOne(ctx, bson.M{"_id": messageObjectID}, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to update message")
	}

	// Broadcast edit to conversation participants
	editMessage := WebSocketMessage{
		Type:    "message",
		Action:  "edited",
		Channel: "conversation:" + message.ConversationID.Hex(),
		Data: map[string]interface{}{
			"message_id":      messageID,
			"conversation_id": message.ConversationID.Hex(),
			"content":         newContent,
			"is_edited":       true,
			"edited_at":       now,
		},
	}

	h.broadcastToConversationParticipants(message.ConversationID, editMessage, primitive.NilObjectID)

	return nil
}

// handleDeleteMessage handles deleting a message
func (h *MessageHandler) handleDeleteMessage(client *Client, wsMessage WebSocketMessage) error {
	messageID, ok := wsMessage.Data["message_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing message_id")
	}

	// Validate message ID
	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_MESSAGE", "Invalid message ID")
	}

	// Get message from database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message models.Message
	err = h.messagesColl.FindOne(ctx, bson.M{
		"_id":        messageObjectID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return h.sendError(client, wsMessage.RequestID, "MESSAGE_NOT_FOUND", "Message not found")
		}
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to fetch message")
	}

	// Check permissions (sender or conversation admin)
	if message.SenderID != client.UserID && !h.isConversationAdmin(client.UserID, message.ConversationID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "Not authorized to delete this message")
	}

	// Soft delete message
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = h.messagesColl.UpdateOne(ctx, bson.M{"_id": messageObjectID}, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to delete message")
	}

	// Broadcast deletion to conversation participants
	deleteMessage := WebSocketMessage{
		Type:    "message",
		Action:  "deleted",
		Channel: "conversation:" + message.ConversationID.Hex(),
		Data: map[string]interface{}{
			"message_id":      messageID,
			"conversation_id": message.ConversationID.Hex(),
			"deleted_at":      now,
		},
	}

	h.broadcastToConversationParticipants(message.ConversationID, deleteMessage, primitive.NilObjectID)

	return nil
}

// handleReactToMessage handles adding/removing reactions to messages
func (h *MessageHandler) handleReactToMessage(client *Client, wsMessage WebSocketMessage) error {
	messageID, ok := wsMessage.Data["message_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing message_id")
	}

	reactionType, ok := wsMessage.Data["reaction_type"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing reaction_type")
	}

	action, ok := wsMessage.Data["action"].(string)
	if !ok {
		action = "add"
	}

	// Validate message ID
	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_MESSAGE", "Invalid message ID")
	}

	// Get message from database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message models.Message
	err = h.messagesColl.FindOne(ctx, bson.M{
		"_id":        messageObjectID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return h.sendError(client, wsMessage.RequestID, "MESSAGE_NOT_FOUND", "Message not found")
		}
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to fetch message")
	}

	// Check if user is in conversation
	if !h.isUserInConversation(client.UserID, message.ConversationID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "User not in conversation")
	}

	// Update reactions
	reactionKey := fmt.Sprintf("reactions_count.%s", reactionType)
	var update bson.M

	if action == "add" {
		update = bson.M{
			"$inc": bson.M{reactionKey: 1},
			"$set": bson.M{"updated_at": time.Now()},
		}
	} else {
		update = bson.M{
			"$inc": bson.M{reactionKey: -1},
			"$set": bson.M{"updated_at": time.Now()},
		}
	}

	_, err = h.messagesColl.UpdateOne(ctx, bson.M{"_id": messageObjectID}, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to update reaction")
	}

	// Broadcast reaction to conversation participants
	reactionMessage := WebSocketMessage{
		Type:    "message",
		Action:  "reaction",
		Channel: "conversation:" + message.ConversationID.Hex(),
		Data: map[string]interface{}{
			"message_id":      messageID,
			"conversation_id": message.ConversationID.Hex(),
			"user_id":         client.UserID.Hex(),
			"username":        client.Username,
			"reaction_type":   reactionType,
			"action":          action,
		},
	}

	h.broadcastToConversationParticipants(message.ConversationID, reactionMessage, primitive.NilObjectID)

	return nil
}

// handleMarkAsRead handles marking messages as read
func (h *MessageHandler) handleMarkAsRead(client *Client, wsMessage WebSocketMessage) error {
	conversationID, ok := wsMessage.Data["conversation_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing conversation_id")
	}

	lastMessageID, ok := wsMessage.Data["last_message_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing last_message_id")
	}

	// Validate IDs
	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_CONVERSATION", "Invalid conversation ID")
	}

	lastMessageObjectID, err := primitive.ObjectIDFromHex(lastMessageID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_MESSAGE", "Invalid message ID")
	}

	// Check if user is in conversation
	if !h.isUserInConversation(client.UserID, conversationObjectID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "User not in conversation")
	}

	// Mark messages as read
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	readReceipt := models.MessageReadReceipt{
		UserID: client.UserID,
		ReadAt: now,
	}

	// Update messages with read receipt
	filter := bson.M{
		"conversation_id": conversationObjectID,
		"_id":             bson.M{"$lte": lastMessageObjectID},
		"sender_id":       bson.M{"$ne": client.UserID}, // Don't mark own messages
		"status":          bson.M{"$ne": models.MessageRead},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.MessageRead,
			"read_at":    now,
			"updated_at": now,
		},
		"$addToSet": bson.M{
			"read_by": readReceipt,
		},
	}

	_, err = h.messagesColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to mark messages as read")
	}

	// Broadcast read receipt to conversation participants
	readMessage := WebSocketMessage{
		Type:    "message",
		Action:  "read",
		Channel: "conversation:" + conversationID,
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"last_message_id": lastMessageID,
			"user_id":         client.UserID.Hex(),
			"username":        client.Username,
			"read_at":         now,
		},
	}

	h.broadcastToConversationParticipants(conversationObjectID, readMessage, client.UserID)

	return nil
}

// handleTypingStart handles typing start indicator
func (h *MessageHandler) handleTypingStart(client *Client, wsMessage WebSocketMessage) error {
	conversationID, ok := wsMessage.Data["conversation_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing conversation_id")
	}

	// Validate conversation ID
	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_CONVERSATION", "Invalid conversation ID")
	}

	// Check if user is in conversation
	if !h.isUserInConversation(client.UserID, conversationObjectID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "User not in conversation")
	}

	// Update typing indicator
	h.updateTypingIndicator(conversationID, client.UserID, true)

	// Broadcast typing indicator
	typingMessage := WebSocketMessage{
		Type:    "typing",
		Action:  "start",
		Channel: "conversation:" + conversationID,
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"user_id":         client.UserID.Hex(),
			"username":        client.Username,
			"is_typing":       true,
		},
	}

	h.broadcastToConversationParticipants(conversationObjectID, typingMessage, client.UserID)

	return nil
}

// handleTypingStop handles typing stop indicator
func (h *MessageHandler) handleTypingStop(client *Client, wsMessage WebSocketMessage) error {
	conversationID, ok := wsMessage.Data["conversation_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing conversation_id")
	}

	// Clear typing indicator
	h.clearTypingIndicator(conversationID, client.UserID)

	// Validate conversation ID
	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_CONVERSATION", "Invalid conversation ID")
	}

	// Broadcast typing stop
	typingMessage := WebSocketMessage{
		Type:    "typing",
		Action:  "stop",
		Channel: "conversation:" + conversationID,
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"user_id":         client.UserID.Hex(),
			"username":        client.Username,
			"is_typing":       false,
		},
	}

	h.broadcastToConversationParticipants(conversationObjectID, typingMessage, client.UserID)

	return nil
}

// handleGetMessages handles fetching messages for a conversation
// handleGetMessages handles fetching messages for a conversation
func (h *MessageHandler) handleGetMessages(client *Client, wsMessage WebSocketMessage) error {
	conversationID, ok := wsMessage.Data["conversation_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing conversation_id")
	}

	// Validate conversation ID
	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_CONVERSATION", "Invalid conversation ID")
	}

	// Check if user is in conversation
	if !h.isUserInConversation(client.UserID, conversationObjectID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "User not in conversation")
	}

	// Get pagination parameters
	page := int64(1)
	if pageData, exists := wsMessage.Data["page"]; exists {
		if pageFloat, ok := pageData.(float64); ok {
			page = int64(pageFloat)
		}
	}

	limit := int64(20)
	if limitData, exists := wsMessage.Data["limit"]; exists {
		if limitFloat, ok := limitData.(float64); ok {
			limit = int64(limitFloat)
		}
	}

	// Fetch messages
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := (page - 1) * limit
	cursor, err := h.messagesColl.Find(ctx, bson.M{
		"conversation_id": conversationObjectID,
		"deleted_at":      bson.M{"$exists": false},
	}, &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Skip:  &skip,
		Limit: &limit,
	})

	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to fetch messages")
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to decode messages")
	}

	// Convert to response format
	messageResponses := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		messageResponses[i] = map[string]interface{}{
			"id":              msg.ID.Hex(),
			"conversation_id": msg.ConversationID.Hex(),
			"sender_id":       msg.SenderID.Hex(),
			"content":         msg.Content,
			"content_type":    msg.ContentType,
			"media":           msg.Media,
			"status":          msg.Status,
			"sent_at":         msg.SentAt,
			"created_at":      msg.CreatedAt,
			"is_edited":       msg.IsEdited,
			"edited_at":       msg.EditedAt,
		}

		if msg.ReplyToMessageID != nil {
			messageResponses[i]["reply_to_message_id"] = msg.ReplyToMessageID.Hex()
		}
	}

	// Send response
	responseMessage := WebSocketMessage{
		Type:      "message",
		Action:    "messages",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"messages":        messageResponses,
			"page":            page,
			"limit":           limit,
			"has_more":        len(messages) == int(limit),
		},
	}

	return client.SendMessage(responseMessage)
}

// handleSearchMessages handles searching messages in a conversation
// handleSearchMessages handles searching messages in a conversation
func (h *MessageHandler) handleSearchMessages(client *Client, wsMessage WebSocketMessage) error {
	conversationID, ok := wsMessage.Data["conversation_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing conversation_id")
	}

	query, ok := wsMessage.Data["query"].(string)
	if !ok || query == "" {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing search query")
	}

	// Validate conversation ID
	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_CONVERSATION", "Invalid conversation ID")
	}

	// Check if user is in conversation
	if !h.isUserInConversation(client.UserID, conversationObjectID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "User not in conversation")
	}

	// Search messages
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"content":         bson.M{"$regex": query, "$options": "i"},
		" respondent_at":  bson.M{"$exists": false},
	}

	cursor, err := h.messagesColl.Find(ctx, filter, &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: func() *int64 { l := int64(50); return &l }(),
	})

	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to search messages")
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to decode search results")
	}

	// Convert to response format
	searchResults := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		searchResults[i] = map[string]interface{}{
			"id":         msg.ID.Hex(),
			"content":    msg.Content,
			"sender_id":  msg.SenderID.Hex(),
			"created_at": msg.CreatedAt,
		}
	}

	// Send response
	responseMessage := WebSocketMessage{
		Type:      "message",
		Action:    "search_results",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"query":           query,
			"results":         searchResults,
			"count":           len(searchResults),
		},
	}

	return client.SendMessage(responseMessage)
}

// handleForwardMessage handles forwarding a message to other conversations
func (h *MessageHandler) handleForwardMessage(client *Client, wsMessage WebSocketMessage) error {
	messageID, ok := wsMessage.Data["message_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing message_id")
	}

	targetConversationIDs, ok := wsMessage.Data["target_conversations"].([]interface{})
	if !ok || len(targetConversationIDs) == 0 {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing target conversations")
	}

	// Validate message ID
	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_MESSAGE", "Invalid message ID")
	}

	// Get original message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var originalMessage models.Message
	err = h.messagesColl.FindOne(ctx, bson.M{
		"_id":        messageObjectID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&originalMessage)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return h.sendError(client, wsMessage.RequestID, "MESSAGE_NOT_FOUND", "Message not found")
		}
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to fetch message")
	}

	// Check if user can access original message
	if !h.isUserInConversation(client.UserID, originalMessage.ConversationID) {
		return h.sendError(client, wsMessage.RequestID, "UNAUTHORIZED", "Cannot access original message")
	}

	// Forward to each target conversation
	successCount := 0
	for _, targetID := range targetConversationIDs {
		if targetIDStr, ok := targetID.(string); ok {
			if targetObjectID, err := primitive.ObjectIDFromHex(targetIDStr); err == nil {
				// Check if user is in target conversation
				if h.isUserInConversation(client.UserID, targetObjectID) {
					// Create forwarded message
					forwardedMessage := models.Message{
						ConversationID: targetObjectID,
						SenderID:       client.UserID,
						Content:        originalMessage.Content,
						ContentType:    originalMessage.ContentType,
						Media:          originalMessage.Media,
						IsForwarded:    true,
						ForwardedFrom:  &originalMessage.ID,
						Status:         models.MessageSent,
						Source:         "websocket",
					}

					forwardedMessage.BeforeCreate()
					now := time.Now()
					forwardedMessage.SentAt = &now

					// Insert forwarded message
					if result, err := h.messagesColl.InsertOne(ctx, forwardedMessage); err == nil {
						forwardedMessage.ID = result.InsertedID.(primitive.ObjectID)
						successCount++

						// Broadcast to target conversation
						broadcastMessage := WebSocketMessage{
							Type:    "message",
							Action:  "new",
							Channel: "conversation:" + targetIDStr,
							Data: map[string]interface{}{
								"id":              forwardedMessage.ID.Hex(),
								"conversation_id": targetIDStr,
								"sender_id":       client.UserID.Hex(),
								"content":         forwardedMessage.Content,
								"content_type":    forwardedMessage.ContentType,
								"media":           forwardedMessage.Media,
								"is_forwarded":    true,
								"forwarded_from":  originalMessage.ID.Hex(),
								"sent_at":         forwardedMessage.SentAt,
								"created_at":      forwardedMessage.CreatedAt,
							},
						}

						h.broadcastToConversationParticipants(targetObjectID, broadcastMessage, client.UserID)
					}
				}
			}
		}
	}

	// Send response
	responseMessage := WebSocketMessage{
		Type:      "message",
		Action:    "forwarded",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"message_id":    messageID,
			"success_count": successCount,
			"total_targets": len(targetConversationIDs),
		},
	}

	return client.SendMessage(responseMessage)
}

// Helper methods

// sendError sends an error message to the client
func (h *MessageHandler) sendError(client *Client, requestID, errorCode, message string) error {
	errorMessage := WebSocketMessage{
		Type:      "error",
		RequestID: requestID,
		Data: map[string]interface{}{
			"error_code": errorCode,
			"message":    message,
		},
	}
	return client.SendMessage(errorMessage)
}

// isUserInConversation checks if a user is a participant in a conversation
func (h *MessageHandler) isUserInConversation(userID, conversationID primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count, err := h.conversationsColl.CountDocuments(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	})

	return err == nil && count > 0
}

// isConversationAdmin checks if a user is an admin of a conversation
func (h *MessageHandler) isConversationAdmin(userID, conversationID primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count, err := h.conversationsColl.CountDocuments(ctx, bson.M{
		"_id":        conversationID,
		"admin_ids":  userID,
		"deleted_at": bson.M{"$exists": false},
	})

	return err == nil && count > 0
}

// getConversationParticipants gets all participants of a conversation
func (h *MessageHandler) getConversationParticipants(conversationID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var conversation models.Conversation
	err := h.conversationsColl.FindOne(ctx, bson.M{
		"_id":        conversationID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		return nil, err
	}

	return conversation.Participants, nil
}

// broadcastToConversationParticipants broadcasts a message to all conversation participants
func (h *MessageHandler) broadcastToConversationParticipants(conversationID primitive.ObjectID, message WebSocketMessage, excludeUserID primitive.ObjectID) {
	participants, err := h.getConversationParticipants(conversationID)
	if err != nil {
		log.Printf("Failed to get conversation participants: %v", err)
		return
	}

	userIDs := make([]string, 0, len(participants))
	for _, userID := range participants {
		if userID != excludeUserID {
			userIDs = append(userIDs, userID.Hex())
		}
	}

	if len(userIDs) > 0 {
		h.hub.BroadcastToUsers(userIDs, message, excludeUserID.Hex())
	}
}

// updateConversationLastMessage updates the last message info in conversation
func (h *MessageHandler) updateConversationLastMessage(conversationID primitive.ObjectID, message models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	preview := message.Content
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}

	update := bson.M{
		"$set": bson.M{
			"last_message_id":      message.ID,
			"last_message_at":      message.CreatedAt,
			"last_message_preview": preview,
			"last_activity_at":     message.CreatedAt,
			"updated_at":           time.Now(),
		},
		"$inc": bson.M{
			"messages_count": 1,
		},
	}

	_, err := h.conversationsColl.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	return err
}

// Typing indicator management

// updateTypingIndicator updates typing indicator for a user in a conversation
func (h *MessageHandler) updateTypingIndicator(conversationID string, userID primitive.ObjectID, isTyping bool) {
	if h.typingIndicators[conversationID] == nil {
		h.typingIndicators[conversationID] = make(map[primitive.ObjectID]time.Time)
	}

	if isTyping {
		h.typingIndicators[conversationID][userID] = time.Now()
	} else {
		delete(h.typingIndicators[conversationID], userID)
		if len(h.typingIndicators[conversationID]) == 0 {
			delete(h.typingIndicators, conversationID)
		}
	}
}

// clearTypingIndicator clears typing indicator for a user
func (h *MessageHandler) clearTypingIndicator(conversationID string, userID primitive.ObjectID) {
	if indicators := h.typingIndicators[conversationID]; indicators != nil {
		delete(indicators, userID)
		if len(indicators) == 0 {
			delete(h.typingIndicators, conversationID)
		}
	}
}

// CleanupExpiredTypingIndicators removes expired typing indicators
func (h *MessageHandler) CleanupExpiredTypingIndicators() {
	now := time.Now()
	timeout := 10 * time.Second

	for conversationID, indicators := range h.typingIndicators {
		for userID, lastTyping := range indicators {
			if now.Sub(lastTyping) > timeout {
				delete(indicators, userID)
			}
		}
		if len(indicators) == 0 {
			delete(h.typingIndicators, conversationID)
		}
	}
}

// Utility functions

// getStringFromMap safely gets a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
