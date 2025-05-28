// internal/services/conversation_service.go
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

type ConversationService struct {
	conversationCollection *mongo.Collection
	messageCollection      *mongo.Collection
	userCollection         *mongo.Collection
	db                     *mongo.Database
}

func NewConversationService() *ConversationService {
	return &ConversationService{
		conversationCollection: config.DB.Collection("conversations"),
		messageCollection:      config.DB.Collection("messages"),
		userCollection:         config.DB.Collection("users"),
		db:                     config.DB,
	}
}

// CreateConversation creates a new conversation
func (cs *ConversationService) CreateConversation(creatorID primitive.ObjectID, req models.CreateConversationRequest) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert participant IDs
	var participants []primitive.ObjectID
	for _, participantIDStr := range req.ParticipantIDs {
		participantID, err := primitive.ObjectIDFromHex(participantIDStr)
		if err != nil {
			return nil, errors.New("invalid participant ID: " + participantIDStr)
		}
		participants = append(participants, participantID)
	}

	// Ensure creator is in participants
	found := false
	for _, participantID := range participants {
		if participantID == creatorID {
			found = true
			break
		}
	}
	if !found {
		participants = append(participants, creatorID)
	}

	// For direct conversations (2 participants), check if conversation already exists
	if len(participants) == 2 && req.Type == models.ConversationTypeDirect {
		existingConv, err := cs.findDirectConversation(ctx, participants[0], participants[1])
		if err == nil {
			return existingConv, nil
		}
	}

	// Validate participants exist
	if err := cs.validateParticipants(ctx, participants); err != nil {
		return nil, err
	}

	// Create conversation
	conversation := &models.Conversation{
		Type:             req.Type,
		Title:            req.Title,
		Description:      req.Description,
		Participants:     participants,
		AdminIDs:         []primitive.ObjectID{creatorID},
		CreatedBy:        creatorID,
		IsActive:         true,
		MessagesCount:    0,
		LastActivityAt:   time.Now(),
		Settings:         models.ConversationSettings{},
		CustomProperties: req.CustomProperties,
	}

	// Set default settings
	conversation.Settings = models.ConversationSettings{
		AllowNewMembers:     true,
		MuteNotifications:   false,
		AllowMediaSharing:   true,
		AllowFileSharing:    true,
		MessageRetention:    0, // No retention limit
		ReadReceiptsEnabled: true,
		TypingIndicators:    true,
	}

	// Override with provided settings
	if req.Settings != nil {
		if req.Settings.AllowNewMembers != nil {
			conversation.Settings.AllowNewMembers = *req.Settings.AllowNewMembers
		}
		if req.Settings.MuteNotifications != nil {
			conversation.Settings.MuteNotifications = *req.Settings.MuteNotifications
		}
		if req.Settings.AllowMediaSharing != nil {
			conversation.Settings.AllowMediaSharing = *req.Settings.AllowMediaSharing
		}
		if req.Settings.AllowFileSharing != nil {
			conversation.Settings.AllowFileSharing = *req.Settings.AllowFileSharing
		}
		if req.Settings.MessageRetention != nil {
			conversation.Settings.MessageRetention = *req.Settings.MessageRetention
		}
		if req.Settings.ReadReceiptsEnabled != nil {
			conversation.Settings.ReadReceiptsEnabled = *req.Settings.ReadReceiptsEnabled
		}
		if req.Settings.TypingIndicators != nil {
			conversation.Settings.TypingIndicators = *req.Settings.TypingIndicators
		}
	}

	conversation.BeforeCreate()

	// Insert conversation
	result, err := cs.conversationCollection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	conversation.ID = result.InsertedID.(primitive.ObjectID)

	// Populate participant information
	cs.populateConversationParticipants(ctx, conversation)

	return conversation, nil
}

// GetUserConversations retrieves conversations for a user
func (cs *ConversationService) GetUserConversations(userID primitive.ObjectID, limit, skip int) ([]models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"participants": userID,
		"is_active":    true,
		"deleted_at":   bson.M{"$exists": false},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"last_activity_at": -1})

	cursor, err := cs.conversationCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []models.Conversation
	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	// Populate participant information for all conversations
	for i := range conversations {
		cs.populateConversationParticipants(ctx, &conversations[i])
		cs.setUnreadCount(ctx, &conversations[i], userID)
	}

	return conversations, nil
}

// GetConversationByID retrieves a specific conversation
func (cs *ConversationService) GetConversationByID(conversationID, userID primitive.ObjectID) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"is_active":    true,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("conversation not found or access denied")
		}
		return nil, err
	}

	// Populate participant information
	cs.populateConversationParticipants(ctx, &conversation)
	cs.setUnreadCount(ctx, &conversation, userID)

	return &conversation, nil
}

// UpdateConversation updates conversation details
func (cs *ConversationService) UpdateConversation(conversationID, userID primitive.ObjectID, req models.UpdateConversationRequest) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get existing conversation and verify permissions
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("conversation not found or access denied")
		}
		return nil, err
	}

	// Check if user is admin for certain operations
	isAdmin := cs.isUserAdmin(userID, conversation.AdminIDs)
	if !isAdmin && (req.Title != nil || req.Description != nil || req.Settings != nil) {
		return nil, errors.New("admin privileges required")
	}

	// Build update document
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if req.Title != nil {
		update["$set"].(bson.M)["title"] = *req.Title
	}

	if req.Description != nil {
		update["$set"].(bson.M)["description"] = *req.Description
	}

	if req.Settings != nil {
		if req.Settings.AllowNewMembers != nil {
			update["$set"].(bson.M)["settings.allow_new_members"] = *req.Settings.AllowNewMembers
		}
		if req.Settings.MuteNotifications != nil {
			update["$set"].(bson.M)["settings.mute_notifications"] = *req.Settings.MuteNotifications
		}
		if req.Settings.AllowMediaSharing != nil {
			update["$set"].(bson.M)["settings.allow_media_sharing"] = *req.Settings.AllowMediaSharing
		}
		if req.Settings.AllowFileSharing != nil {
			update["$set"].(bson.M)["settings.allow_file_sharing"] = *req.Settings.AllowFileSharing
		}
		if req.Settings.MessageRetention != nil {
			update["$set"].(bson.M)["settings.message_retention"] = *req.Settings.MessageRetention
		}
		if req.Settings.ReadReceiptsEnabled != nil {
			update["$set"].(bson.M)["settings.read_receipts_enabled"] = *req.Settings.ReadReceiptsEnabled
		}
		if req.Settings.TypingIndicators != nil {
			update["$set"].(bson.M)["settings.typing_indicators"] = *req.Settings.TypingIndicators
		}
	}

	// Update conversation
	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	if err != nil {
		return nil, err
	}

	// Get updated conversation
	return cs.GetConversationByID(conversationID, userID)
}

// AddParticipants adds participants to a conversation
func (cs *ConversationService) AddParticipants(conversationID, userID primitive.ObjectID, participantIDStrs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get conversation and verify permissions
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("conversation not found or access denied")
		}
		return err
	}

	// Check permissions
	if !conversation.Settings.AllowNewMembers && !cs.isUserAdmin(userID, conversation.AdminIDs) {
		return errors.New("only admins can add new members")
	}

	// Convert and validate participant IDs
	var newParticipants []primitive.ObjectID
	for _, participantIDStr := range participantIDStrs {
		participantID, err := primitive.ObjectIDFromHex(participantIDStr)
		if err != nil {
			return errors.New("invalid participant ID: " + participantIDStr)
		}

		// Check if already a participant
		alreadyParticipant := false
		for _, existingID := range conversation.Participants {
			if existingID == participantID {
				alreadyParticipant = true
				break
			}
		}

		if !alreadyParticipant {
			newParticipants = append(newParticipants, participantID)
		}
	}

	if len(newParticipants) == 0 {
		return errors.New("no new participants to add")
	}

	// Validate participants exist
	if err := cs.validateParticipants(ctx, newParticipants); err != nil {
		return err
	}

	// Add participants
	update := bson.M{
		"$addToSet": bson.M{
			"participants": bson.M{"$each": newParticipants},
		},
		"$set": bson.M{
			"updated_at":       time.Now(),
			"last_activity_at": time.Now(),
		},
	}

	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	return err
}

// RemoveParticipant removes a participant from a conversation
func (cs *ConversationService) RemoveParticipant(conversationID, userID, participantID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get conversation and verify permissions
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("conversation not found or access denied")
		}
		return err
	}

	// Check permissions - admins can remove anyone, users can only remove themselves
	if userID != participantID && !cs.isUserAdmin(userID, conversation.AdminIDs) {
		return errors.New("admin privileges required to remove other participants")
	}

	// Don't allow removing the last admin
	if cs.isUserAdmin(participantID, conversation.AdminIDs) && len(conversation.AdminIDs) == 1 {
		return errors.New("cannot remove the last admin")
	}

	// Remove participant
	update := bson.M{
		"$pull": bson.M{
			"participants": participantID,
			"admin_ids":    participantID, // Also remove from admins if they were admin
		},
		"$set": bson.M{
			"updated_at":       time.Now(),
			"last_activity_at": time.Now(),
		},
	}

	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	return err
}

// LeaveConversation allows a user to leave a conversation
func (cs *ConversationService) LeaveConversation(conversationID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get conversation
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("conversation not found or access denied")
		}
		return err
	}

	// Don't allow leaving direct conversations
	if conversation.Type == models.ConversationTypeDirect {
		return errors.New("cannot leave direct conversations")
	}

	// Don't allow the last admin to leave
	if cs.isUserAdmin(userID, conversation.AdminIDs) && len(conversation.AdminIDs) == 1 && len(conversation.Participants) > 1 {
		return errors.New("cannot leave as the last admin. Transfer admin rights first")
	}

	return cs.RemoveParticipant(conversationID, userID, userID)
}

// Helper methods

// findDirectConversation finds existing direct conversation between two users
func (cs *ConversationService) findDirectConversation(ctx context.Context, user1ID, user2ID primitive.ObjectID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"type":         models.ConversationTypeDirect,
		"participants": bson.M{"$all": []primitive.ObjectID{user1ID, user2ID}, "$size": 2},
		"is_active":    true,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		return nil, err
	}

	cs.populateConversationParticipants(ctx, &conversation)
	return &conversation, nil
}

// validateParticipants validates that all participant IDs exist and are active users
func (cs *ConversationService) validateParticipants(ctx context.Context, participantIDs []primitive.ObjectID) error {
	count, err := cs.userCollection.CountDocuments(ctx, bson.M{
		"_id":        bson.M{"$in": participantIDs},
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	})

	if err != nil {
		return err
	}

	if count != int64(len(participantIDs)) {
		return errors.New("one or more participants not found or inactive")
	}

	return nil
}

// isUserAdmin checks if user is admin of conversation
func (cs *ConversationService) isUserAdmin(userID primitive.ObjectID, adminIDs []primitive.ObjectID) bool {
	for _, adminID := range adminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// populateConversationParticipants populates participant information
func (cs *ConversationService) populateConversationParticipants(ctx context.Context, conversation *models.Conversation) {
	// Get participant details
	cursor, err := cs.userCollection.Find(ctx, bson.M{
		"_id": bson.M{"$in": conversation.Participants},
	}, options.Find().SetProjection(bson.M{
		"password": 0, // Exclude sensitive fields
	}))

	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return
	}

	// Convert to response format
	var participantResponses []models.UserResponse
	for _, user := range users {
		participantResponses = append(participantResponses, user.ToUserResponse())
	}

	conversation.ParticipantDetails = participantResponses
}

// setUnreadCount sets unread message count for user
func (cs *ConversationService) setUnreadCount(ctx context.Context, conversation *models.Conversation, userID primitive.ObjectID) {
	// Count unread messages
	count, err := cs.messageCollection.CountDocuments(ctx, bson.M{
		"conversation_id": conversation.ID,
		"sender_id":       bson.M{"$ne": userID},
		"read_by.user_id": bson.M{"$ne": userID},
		"deleted_at":      bson.M{"$exists": false},
	})

	if err == nil {
		conversation.UnreadCount = count
	}
}

// GetConversationStats returns conversation statistics
func (cs *ConversationService) GetConversationStats(conversationID primitive.ObjectID, userID primitive.ObjectID) (*models.ConversationStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Verify user access
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		return nil, errors.New("conversation not found or access denied")
	}

	// Aggregate message statistics
	pipeline := []bson.M{
		{"$match": bson.M{
			"conversation_id": conversationID,
			"deleted_at":      bson.M{"$exists": false},
		}},
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
				"message_senders": bson.M{"$addToSet": "$sender_id"},
				"first_message":   bson.M{"$min": "$created_at"},
				"last_message":    bson.M{"$max": "$created_at"},
			},
		},
	}

	cursor, err := cs.messageCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalMessages  int64                `bson:"total_messages"`
		TotalMedia     int64                `bson:"total_media"`
		MessageSenders []primitive.ObjectID `bson:"message_senders"`
		FirstMessage   time.Time            `bson:"first_message"`
		LastMessage    time.Time            `bson:"last_message"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := &models.ConversationStats{
		ConversationID:     conversationID,
		ParticipantCount:   int64(len(conversation.Participants)),
		AdminCount:         int64(len(conversation.AdminIDs)),
		TotalMessages:      0,
		TotalMediaFiles:    0,
		ActiveParticipants: 0,
		CreatedAt:          conversation.CreatedAt,
	}

	if len(results) > 0 {
		result := results[0]
		stats.TotalMessages = result.TotalMessages
		stats.TotalMediaFiles = result.TotalMedia
		stats.ActiveParticipants = int64(len(result.MessageSenders))
		stats.FirstMessageAt = &result.FirstMessage
		stats.LastMessageAt = &result.LastMessage
	}

	return stats, nil
}

// MuteConversation mutes/unmutes conversation for user
func (cs *ConversationService) MuteConversation(conversationID, userID primitive.ObjectID, muted bool, muteUntil *time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This would typically be stored in a separate user_conversation_settings collection
	// For now, we'll update the conversation settings
	// In a real implementation, you'd want per-user mute settings

	filter := bson.M{
		"_id":          conversationID,
		"participants": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"settings.mute_notifications": muted,
			"updated_at":                  time.Now(),
		},
	}

	if muteUntil != nil {
		update["$set"].(bson.M)["settings.mute_until"] = *muteUntil
	}

	_, err := cs.conversationCollection.UpdateOne(ctx, filter, update)
	return err
}
