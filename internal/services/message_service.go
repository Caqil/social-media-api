// internal/services/message_service.go
package services

import (
	"context"
	"errors"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageService struct {
	messageCollection      *mongo.Collection
	conversationCollection *mongo.Collection
	userCollection         *mongo.Collection
	db                     *mongo.Database
}

func NewMessageService() *MessageService {
	return &MessageService{
		messageCollection:      config.DB.Collection("messages"),
		conversationCollection: config.DB.Collection("conversations"),
		userCollection:         config.DB.Collection("users"),
		db:                     config.DB,
	}
}

// SendMessage sends a new message in a conversation
func (ms *MessageService) SendMessage(senderID, conversationID primitive.ObjectID, req models.CreateMessageRequest) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify user is participant in conversation
	if !ms.isUserInConversation(ctx, senderID, conversationID) {
		return nil, errors.New("access denied: user not in conversation")
	}

	// Handle reply to message
	var replyToMessageID *primitive.ObjectID
	if req.ReplyToMessageID != "" {
		if replyID, err := primitive.ObjectIDFromHex(req.ReplyToMessageID); err == nil {
			replyToMessageID = &replyID
		}
	}

	// Create message
	message := &models.Message{
		ConversationID:   conversationID,
		SenderID:         senderID,
		Content:          req.Content,
		ContentType:      req.ContentType,
		Media:            req.Media, // Already []models.MediaInfo
		ReplyToMessageID: replyToMessageID,
		Status:           models.MessageSent,
		Source:           "api",
		ReadBy:           []models.MessageReadReceipt{},
		ReactionsCount:   make(map[models.ReactionType]int64), // Fixed: use ReactionType not string
		Priority:         req.Priority,
		ExpiresAt:        req.ExpiresAt,
		IsEdited:         false,
		IsForwarded:      false,
		ForwardedFrom:    nil,
		IsExpired:        false,
		IsThreadRoot:     false,
		ThreadCount:      0,
	}

	// Set default priority if empty
	if message.Priority == "" {
		message.Priority = "normal"
	}

	message.BeforeCreate()
	now := time.Now()
	message.SentAt = &now

	// Insert message
	result, err := ms.messageCollection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)

	// Update conversation's last message
	go ms.updateConversationLastMessage(conversationID, message)

	// Populate sender information
	ms.populateMessageSender(ctx, message)

	return message, nil
}

// GetConversationMessages retrieves messages from a conversation
func (ms *MessageService) GetConversationMessages(conversationID, userID primitive.ObjectID, limit, skip int) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify user is participant in conversation
	if !ms.isUserInConversation(ctx, userID, conversationID) {
		return nil, errors.New("access denied: user not in conversation")
	}

	filter := bson.M{
		"conversation_id": conversationID,
		"deleted_at":      bson.M{"$exists": false},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ms.messageCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Populate sender information for all messages
	for i := range messages {
		ms.populateMessageSender(ctx, &messages[i])

		// Populate reply to message if exists
		if messages[i].ReplyToMessageID != nil {
			ms.populateReplyToMessage(ctx, &messages[i])
		}
	}

	return messages, nil
}

// GetMessageByID retrieves a specific message
func (ms *MessageService) GetMessageByID(messageID, userID primitive.ObjectID) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message models.Message
	err := ms.messageCollection.FindOne(ctx, bson.M{
		"_id":        messageID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		return nil, err
	}

	// Verify user can access this message
	if !ms.isUserInConversation(ctx, userID, message.ConversationID) {
		return nil, errors.New("access denied")
	}

	// Populate sender information
	ms.populateMessageSender(ctx, &message)

	// Populate reply to message if exists
	if message.ReplyToMessageID != nil {
		ms.populateReplyToMessage(ctx, &message)
	}

	return &message, nil
}

// UpdateMessage updates an existing message
func (ms *MessageService) UpdateMessage(messageID, userID primitive.ObjectID, req models.UpdateMessageRequest) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get existing message
	var message models.Message
	err := ms.messageCollection.FindOne(ctx, bson.M{
		"_id":        messageID,
		"sender_id":  userID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("message not found or access denied")
		}
		return nil, err
	}

	// Check if message can be edited
	if !message.CanEditMessage(userID) {
		return nil, errors.New("message cannot be edited")
	}

	// Build update document
	now := time.Now()
	update := bson.M{"$set": bson.M{"updated_at": now}}

	if req.Content != "" {
		update["$set"].(bson.M)["content"] = req.Content
		update["$set"].(bson.M)["is_edited"] = true
		update["$set"].(bson.M)["edited_at"] = now
	}

	if len(req.Media) > 0 {
		update["$set"].(bson.M)["media"] = req.Media
		update["$set"].(bson.M)["is_edited"] = true
		update["$set"].(bson.M)["edited_at"] = now
	}

	// Update message
	_, err = ms.messageCollection.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	if err != nil {
		return nil, err
	}

	// Get updated message
	return ms.GetMessageByID(messageID, userID)
}

// DeleteMessage soft deletes a message
func (ms *MessageService) DeleteMessage(messageID, userID primitive.ObjectID) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get message to verify ownership and get conversation info
	var message models.Message
	err := ms.messageCollection.FindOne(ctx, bson.M{
		"_id":        messageID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	// Check if user can delete this message (sender or conversation admin)
	if !message.CanDeleteMessage(userID, models.RoleUser, ms.isConversationAdmin(ctx, userID, message.ConversationID)) {
		return nil, errors.New("access denied")
	}

	// Soft delete the message
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = ms.messageCollection.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	if err != nil {
		return nil, err
	}

	message.DeletedAt = &now
	return &message, nil
}

// MarkMessagesAsRead marks messages as read for a user
func (ms *MessageService) MarkMessagesAsRead(conversationID, userID, lastMessageID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify user is in conversation
	if !ms.isUserInConversation(ctx, userID, conversationID) {
		return errors.New("access denied: user not in conversation")
	}

	now := time.Now()
	readReceipt := models.MessageReadReceipt{
		UserID: userID,
		ReadAt: now,
	}

	// Update messages as read
	filter := bson.M{
		"conversation_id": conversationID,
		"_id":             bson.M{"$lte": lastMessageID},
		"sender_id":       bson.M{"$ne": userID}, // Don't mark own messages
		"read_by.user_id": bson.M{"$ne": userID}, // Don't update if already read
		"deleted_at":      bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.MessageRead,
			"updated_at": now,
		},
		"$addToSet": bson.M{
			"read_by": readReceipt,
		},
	}

	_, err := ms.messageCollection.UpdateMany(ctx, filter, update)
	return err
}

// SearchMessages searches for messages
func (ms *MessageService) SearchMessages(userID primitive.ObjectID, query string, conversationID *primitive.ObjectID, limit, skip int) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build search filter
	filter := bson.M{
		"content":    bson.M{"$regex": query, "$options": "i"},
		"deleted_at": bson.M{"$exists": false},
	}

	// Add conversation filter if specified
	if conversationID != nil {
		// Verify user can access the conversation
		if !ms.isUserInConversation(ctx, userID, *conversationID) {
			return nil, errors.New("access denied: user not in conversation")
		}
		filter["conversation_id"] = *conversationID
	} else {
		// Get user's conversations for filtering
		userConversations, err := ms.getUserConversationIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		filter["conversation_id"] = bson.M{"$in": userConversations}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ms.messageCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Populate sender information
	for i := range messages {
		ms.populateMessageSender(ctx, &messages[i])
	}

	return messages, nil
}

// ReactToMessage adds/removes reaction to a message
func (ms *MessageService) ReactToMessage(messageID, userID primitive.ObjectID, reactionType models.ReactionType, action string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get message
	var message models.Message
	err := ms.messageCollection.FindOne(ctx, bson.M{
		"_id":        messageID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&message)

	if err != nil {
		return err
	}

	// Verify user can access this message
	if !ms.isUserInConversation(ctx, userID, message.ConversationID) {
		return errors.New("access denied")
	}

	// Update reaction count
	if action == "add" {
		message.AddReaction(reactionType)
	} else if action == "remove" {
		message.RemoveReaction(reactionType)
	}

	// Save changes
	update := bson.M{
		"$set": bson.M{
			"reactions_count": message.ReactionsCount,
			"updated_at":      time.Now(),
		},
	}

	_, err = ms.messageCollection.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	return err
}

// Helper methods

// isUserInConversation checks if user is participant in conversation
func (ms *MessageService) isUserInConversation(ctx context.Context, userID, conversationID primitive.ObjectID) bool {
	count, err := ms.conversationCollection.CountDocuments(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	})
	return err == nil && count > 0
}

// isConversationAdmin checks if user is admin of conversation
func (ms *MessageService) isConversationAdmin(ctx context.Context, userID, conversationID primitive.ObjectID) bool {
	count, err := ms.conversationCollection.CountDocuments(ctx, bson.M{
		"_id":        conversationID,
		"admin_ids":  userID,
		"deleted_at": bson.M{"$exists": false},
	})
	return err == nil && count > 0
}

// populateMessageSender populates sender information for message
func (ms *MessageService) populateMessageSender(ctx context.Context, message *models.Message) {
	var user models.User
	err := ms.userCollection.FindOne(ctx, bson.M{"_id": message.SenderID}).Decode(&user)
	if err == nil {
		message.Sender = user.ToUserResponse()
	}
}

// populateReplyToMessage populates reply to message information
func (ms *MessageService) populateReplyToMessage(ctx context.Context, message *models.Message) {
	if message.ReplyToMessageID == nil {
		return
	}

	var replyToMessage models.Message
	err := ms.messageCollection.FindOne(ctx, bson.M{
		"_id":        *message.ReplyToMessageID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&replyToMessage)

	if err == nil {
		ms.populateMessageSender(ctx, &replyToMessage)
		response := replyToMessage.ToMessageResponse()
		message.ReplyToMessage = &response
	}
}

// getUserConversationIDs gets all conversation IDs for a user
func (ms *MessageService) getUserConversationIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.M{
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	}

	cursor, err := ms.conversationCollection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	conversationIDs := make([]primitive.ObjectID, len(conversations))
	for i, conv := range conversations {
		conversationIDs[i] = conv.ID
	}

	return conversationIDs, nil
}

// updateConversationLastMessage updates conversation's last message information
func (ms *MessageService) updateConversationLastMessage(conversationID primitive.ObjectID, message *models.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	preview := message.GetMessagePreview()

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

	ms.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
}

// GetMessageStats returns message statistics
func (ms *MessageService) GetMessageStats(conversationID *primitive.ObjectID, userID *primitive.ObjectID) (*models.MessageStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{"deleted_at": bson.M{"$exists": false}}

	if conversationID != nil {
		filter["conversation_id"] = *conversationID
	}

	if userID != nil {
		filter["sender_id"] = *userID
	}

	// Aggregate statistics
	pipeline := []bson.M{
		{"$match": filter},
		{
			"$group": bson.M{
				"_id":            nil,
				"total_messages": bson.M{"$sum": 1},
				"total_media": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$gt": []interface{}{bson.M{"$size": "$media"}, 0}},
							1, 0,
						},
					},
				},
				"content_types": bson.M{"$push": "$content_type"},
				"avg_length":    bson.M{"$avg": bson.M{"$strLenCP": "$content"}},
			},
		},
	}

	cursor, err := ms.messageCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalMessages int64    `bson:"total_messages"`
		TotalMedia    int64    `bson:"total_media"`
		ContentTypes  []string `bson:"content_types"`
		AvgLength     float64  `bson:"avg_length"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &models.MessageStats{}, nil
	}

	result := results[0]

	// Count content types
	contentTypeCounts := make(map[string]int64)
	for _, contentType := range result.ContentTypes {
		contentTypeCounts[contentType]++
	}

	return &models.MessageStats{
		TotalMessages:     result.TotalMessages,
		TotalMediaFiles:   result.TotalMedia,
		AverageLength:     result.AvgLength,
		ContentTypeCounts: contentTypeCounts,
	}, nil
}
