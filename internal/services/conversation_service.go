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
	if len(participants) == 2 && req.Type == "direct" {
		if existingConv, err := cs.findDirectConversation(ctx, participants[0], participants[1]); err == nil {
			return existingConv, nil
		}
	}

	// Validate participants exist
	if err := cs.validateParticipants(ctx, participants); err != nil {
		return nil, err
	}

	// Create conversation using model
	conversation := &models.Conversation{
		Type:              req.Type,
		Title:             req.Title,
		Description:       req.Description,
		AvatarURL:         "",
		Participants:      participants,
		CreatedBy:         creatorID,
		IsPrivate:         req.IsPrivate,
		AllowInvites:      req.AllowInvites,
		AllowMediaSharing: req.AllowMediaSharing,
		MaxParticipants:   req.MaxParticipants,
		Category:          req.Category,
		Tags:              req.Tags,
	}

	// Use model's BeforeCreate method to set defaults
	conversation.BeforeCreate()

	// Insert conversation
	result, err := cs.conversationCollection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	conversation.ID = result.InsertedID.(primitive.ObjectID)

	// Populate participant information
	cs.populateConversationUsers(ctx, conversation)

	// Send initial message if provided
	if req.InitialMessage != "" {
		go cs.sendInitialMessage(conversation.ID, creatorID, req.InitialMessage)
	}

	return conversation, nil
}

// GetUserConversations retrieves conversations for a user
func (cs *ConversationService) GetUserConversations(userID primitive.ObjectID, limit, skip int) ([]models.ConversationResponse, error) {
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

	// Convert to response format
	var responses []models.ConversationResponse
	for _, conv := range conversations {
		// Populate participant information
		cs.populateConversationUsers(ctx, &conv)

		// Convert to response
		response := conv.ToConversationResponse()

		// Set user-specific context
		response.UnreadCount = cs.getUnreadCount(ctx, conv.ID, userID)
		response.IsUserAdmin = conv.IsAdmin(userID)
		response.UserRole = conv.GetParticipantRole(userID)
		response.CanSendMessages = conv.CanSendMessages(userID)
		response.CanAddMembers = conv.CanAddMembers(userID)

		// Get typing users
		response.TypingUsers = cs.getTypingUsers(ctx, conv.ID, userID)

		responses = append(responses, response)
	}

	return responses, nil
}

// GetConversationByID retrieves a specific conversation
func (cs *ConversationService) GetConversationByID(conversationID, userID primitive.ObjectID) (*models.ConversationResponse, error) {
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
	cs.populateConversationUsers(ctx, &conversation)

	// Convert to response
	response := conversation.ToConversationResponse()

	// Set user-specific context
	response.UnreadCount = cs.getUnreadCount(ctx, conversation.ID, userID)
	response.IsUserAdmin = conversation.IsAdmin(userID)
	response.UserRole = conversation.GetParticipantRole(userID)
	response.CanSendMessages = conversation.CanSendMessages(userID)
	response.CanAddMembers = conversation.CanAddMembers(userID)
	response.TypingUsers = cs.getTypingUsers(ctx, conversation.ID, userID)

	return &response, nil
}

// UpdateConversation updates conversation details
func (cs *ConversationService) UpdateConversation(conversationID, userID primitive.ObjectID, req models.UpdateConversationRequest) (*models.ConversationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get existing conversation
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
	if !conversation.IsAdmin(userID) {
		return nil, errors.New("admin privileges required")
	}

	// Build update document
	update := bson.M{"$set": bson.M{}}

	if req.Title != nil {
		update["$set"].(bson.M)["title"] = *req.Title
	}

	if req.Description != nil {
		update["$set"].(bson.M)["description"] = *req.Description
	}

	if req.AvatarURL != nil {
		update["$set"].(bson.M)["avatar_url"] = *req.AvatarURL
	}

	if req.AllowInvites != nil {
		update["$set"].(bson.M)["allow_invites"] = *req.AllowInvites
	}

	if req.AllowMediaSharing != nil {
		update["$set"].(bson.M)["allow_media_sharing"] = *req.AllowMediaSharing
	}

	if req.IsLocked != nil {
		update["$set"].(bson.M)["is_locked"] = *req.IsLocked
	}

	if req.IsPrivate != nil {
		update["$set"].(bson.M)["is_private"] = *req.IsPrivate
	}

	if req.MaxParticipants != nil {
		update["$set"].(bson.M)["max_participants"] = *req.MaxParticipants
	}

	if req.Category != nil {
		update["$set"].(bson.M)["category"] = *req.Category
	}

	if req.Tags != nil {
		update["$set"].(bson.M)["tags"] = req.Tags
	}

	// Always update timestamp
	update["$set"].(bson.M)["updated_at"] = time.Now()

	// Update conversation
	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	if err != nil {
		return nil, err
	}

	// Return updated conversation
	return cs.GetConversationByID(conversationID, userID)
}

// AddParticipants adds participants to a conversation
func (cs *ConversationService) AddParticipants(conversationID, userID primitive.ObjectID, req models.AddParticipantsRequest) error {
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

	// Check permissions
	if !conversation.CanAddMembers(userID) {
		return errors.New("insufficient permissions to add members")
	}

	// Convert and validate participant IDs
	var newParticipants []primitive.ObjectID
	for _, participantIDStr := range req.ParticipantIDs {
		participantID, err := primitive.ObjectIDFromHex(participantIDStr)
		if err != nil {
			return errors.New("invalid participant ID: " + participantIDStr)
		}

		// Check if already a participant
		if !conversation.IsParticipant(participantID) {
			newParticipants = append(newParticipants, participantID)
		}
	}

	if len(newParticipants) == 0 {
		return errors.New("no new participants to add")
	}

	// Check max participants limit
	if conversation.MaxParticipants > 0 && int64(len(conversation.Participants)+len(newParticipants)) > conversation.MaxParticipants {
		return errors.New("would exceed maximum participants limit")
	}

	// Validate participants exist
	if err := cs.validateParticipants(ctx, newParticipants); err != nil {
		return err
	}

	// Add participants using model method
	for _, participantID := range newParticipants {
		conversation.AddParticipant(participantID, &userID)
	}

	// Update in database
	update := bson.M{
		"$set": bson.M{
			"participants":         conversation.Participants,
			"participant_info":     conversation.ParticipantInfo,
			"active_members_count": conversation.ActiveMembersCount,
			"updated_at":           time.Now(),
			"last_activity_at":     time.Now(),
		},
	}

	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	return err
}

// RemoveParticipant removes a participant from a conversation
func (cs *ConversationService) RemoveParticipant(conversationID, userID, participantID primitive.ObjectID) error {
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

	// Check permissions - admins can remove anyone, users can only remove themselves
	if userID != participantID && !conversation.IsAdmin(userID) {
		return errors.New("admin privileges required to remove other participants")
	}

	// Don't allow removing the last admin
	if conversation.IsAdmin(participantID) && len(conversation.AdminIDs) == 1 {
		return errors.New("cannot remove the last admin")
	}

	// Remove participant using model method
	conversation.RemoveParticipant(participantID)

	// Update in database
	update := bson.M{
		"$set": bson.M{
			"participants":         conversation.Participants,
			"participant_info":     conversation.ParticipantInfo,
			"admin_ids":            conversation.AdminIDs,
			"active_members_count": conversation.ActiveMembersCount,
			"updated_at":           time.Now(),
			"last_activity_at":     time.Now(),
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
	if conversation.Type == "direct" {
		return errors.New("cannot leave direct conversations")
	}

	// Don't allow the last admin to leave
	if conversation.IsAdmin(userID) && len(conversation.AdminIDs) == 1 && len(conversation.Participants) > 1 {
		return errors.New("cannot leave as the last admin. Transfer admin rights first")
	}

	return cs.RemoveParticipant(conversationID, userID, userID)
}

// UpdateParticipantRole updates a participant's role
func (cs *ConversationService) UpdateParticipantRole(conversationID, adminID, participantID primitive.ObjectID, req models.UpdateParticipantRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get conversation
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"_id":          conversationID,
		"participants": adminID,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("conversation not found or access denied")
		}
		return err
	}

	// Check if requester is admin
	if !conversation.IsAdmin(adminID) {
		return errors.New("admin privileges required")
	}

	// Check if target is participant
	if !conversation.IsParticipant(participantID) {
		return errors.New("user is not a participant")
	}

	// Update participant role using model method
	if req.Role != nil {
		conversation.UpdateParticipantRole(participantID, *req.Role)

		// Update admin list if promoting/demoting admin
		if *req.Role == "admin" {
			found := false
			for _, adminID := range conversation.AdminIDs {
				if adminID == participantID {
					found = true
					break
				}
			}
			if !found {
				conversation.AdminIDs = append(conversation.AdminIDs, participantID)
			}
		} else {
			// Remove from admin list if demoting
			for i, adminID := range conversation.AdminIDs {
				if adminID == participantID {
					conversation.AdminIDs = append(conversation.AdminIDs[:i], conversation.AdminIDs[i+1:]...)
					break
				}
			}
		}
	}

	// Update other settings
	for i, info := range conversation.ParticipantInfo {
		if info.UserID == participantID {
			if req.NotificationsEnabled != nil {
				conversation.ParticipantInfo[i].NotificationsEnabled = *req.NotificationsEnabled
			}
			if req.Nickname != nil {
				conversation.ParticipantInfo[i].Nickname = *req.Nickname
			}
			if req.IsMuted != nil {
				conversation.ParticipantInfo[i].IsMuted = *req.IsMuted
			}
			break
		}
	}

	// Update in database
	update := bson.M{
		"$set": bson.M{
			"participant_info": conversation.ParticipantInfo,
			"admin_ids":        conversation.AdminIDs,
			"updated_at":       time.Now(),
		},
	}

	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	return err
}

// ArchiveConversation archives/unarchives a conversation for a user
func (cs *ConversationService) ArchiveConversation(conversationID, userID primitive.ObjectID, archived bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify user is participant
	count, err := cs.conversationCollection.CountDocuments(ctx, bson.M{
		"_id":          conversationID,
		"participants": userID,
		"deleted_at":   bson.M{"$exists": false},
	})

	if err != nil || count == 0 {
		return errors.New("conversation not found or access denied")
	}

	// In a real implementation, this would be per-user setting
	// For now, we'll update the conversation's archived status
	update := bson.M{
		"$set": bson.M{
			"is_archived": archived,
			"updated_at":  time.Now(),
		},
	}

	_, err = cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
	return err
}

// Helper methods

// findDirectConversation finds existing direct conversation between two users
func (cs *ConversationService) findDirectConversation(ctx context.Context, user1ID, user2ID primitive.ObjectID) (*models.Conversation, error) {
	var conversation models.Conversation
	err := cs.conversationCollection.FindOne(ctx, bson.M{
		"type":         "direct",
		"participants": bson.M{"$all": []primitive.ObjectID{user1ID, user2ID}, "$size": 2},
		"is_active":    true,
		"deleted_at":   bson.M{"$exists": false},
	}).Decode(&conversation)

	if err != nil {
		return nil, err
	}

	cs.populateConversationUsers(ctx, &conversation)
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

// populateConversationUsers populates participant information from users collection
func (cs *ConversationService) populateConversationUsers(ctx context.Context, conversation *models.Conversation) {
	// Get participant details
	cursor, err := cs.userCollection.Find(ctx, bson.M{
		"_id": bson.M{"$in": conversation.Participants},
	}, options.Find().SetProjection(bson.M{
		"password":       0, // Exclude sensitive fields
		"refresh_tokens": 0,
		"reset_tokens":   0,
	}))

	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return
	}

	// Convert to response format and populate in participant info
	for i, info := range conversation.ParticipantInfo {
		for _, user := range users {
			if user.ID == info.UserID {
				conversation.ParticipantInfo[i].User = user.ToUserResponse()
				break
			}
		}
	}
}

// getUnreadCount gets unread message count for user in conversation
func (cs *ConversationService) getUnreadCount(ctx context.Context, conversationID, userID primitive.ObjectID) int64 {
	count, err := cs.messageCollection.CountDocuments(ctx, bson.M{
		"conversation_id": conversationID,
		"sender_id":       bson.M{"$ne": userID},
		"read_by.user_id": bson.M{"$ne": userID},
		"deleted_at":      bson.M{"$exists": false},
	})

	if err != nil {
		return 0
	}
	return count
}

// getTypingUsers gets users currently typing in conversation
func (cs *ConversationService) getTypingUsers(ctx context.Context, conversationID, excludeUserID primitive.ObjectID) []models.UserResponse {
	// In a real implementation, this would check a typing indicators cache/collection
	// For now, return empty slice
	return []models.UserResponse{}
}

// sendInitialMessage sends an initial message when conversation is created
func (cs *ConversationService) sendInitialMessage(conversationID, senderID primitive.ObjectID, content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		ContentType:    models.ContentTypeText,
		Status:         models.MessageSent,
		Source:         "system",
	}

	message.BeforeCreate()
	now := time.Now()
	message.SentAt = &now

	result, err := cs.messageCollection.InsertOne(ctx, message)
	if err != nil {
		return
	}

	message.ID = result.InsertedID.(primitive.ObjectID)

	// Update conversation's last message
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
			"messages_count":       1,
		},
	}

	cs.conversationCollection.UpdateOne(ctx, bson.M{"_id": conversationID}, update)
}

// GetConversationStats returns conversation statistics
func (cs *ConversationService) GetConversationStats(conversationID, userID primitive.ObjectID) (*models.ConversationStatsResponse, error) {
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

	// Count messages
	messageCount, _ := cs.messageCollection.CountDocuments(ctx, bson.M{
		"conversation_id": conversationID,
		"deleted_at":      bson.M{"$exists": false},
	})

	// Count unread messages for user
	unreadCount, _ := cs.messageCollection.CountDocuments(ctx, bson.M{
		"conversation_id": conversationID,
		"sender_id":       bson.M{"$ne": userID},
		"read_by.user_id": bson.M{"$ne": userID},
		"deleted_at":      bson.M{"$exists": false},
	})

	return &models.ConversationStatsResponse{
		ConversationID:      conversationID.Hex(),
		MessagesCount:       messageCount,
		ActiveMembersCount:  conversation.ActiveMembersCount,
		TotalMembersCount:   int64(len(conversation.Participants)),
		UnreadMessagesCount: unreadCount,
		LastActivityHours:   0, // Calculate based on last_activity_at
	}, nil
}
