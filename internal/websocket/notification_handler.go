// notification_handler.go
package websocket

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NotificationHandlerInterface defines the interface for notification handling
type NotificationHandlerInterface interface {
	HandleNotification(client *Client, message WebSocketMessage) error
	SendNotificationToUser(userID string, notification models.Notification) error
	BroadcastSystemNotification(notification models.Notification) error
}

// NotificationHandler handles real-time notifications
type NotificationHandler struct {
	db                *mongo.Database
	notificationsColl *mongo.Collection
	usersColl         *mongo.Collection
	hub               *Hub

	// Notification delivery tracking
	deliveryStats map[string]*NotificationDeliveryStats
}

// NotificationDeliveryStats tracks notification delivery statistics
type NotificationDeliveryStats struct {
	TotalSent      int64     `json:"total_sent"`
	TotalDelivered int64     `json:"total_delivered"`
	TotalRead      int64     `json:"total_read"`
	LastSent       time.Time `json:"last_sent"`
	FailureCount   int64     `json:"failure_count"`
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(db *mongo.Database, hub *Hub) *NotificationHandler {
	return &NotificationHandler{
		db:                db,
		notificationsColl: db.Collection("notifications"),
		usersColl:         db.Collection("users"),
		hub:               hub,
		deliveryStats:     make(map[string]*NotificationDeliveryStats),
	}
}

// HandleNotification handles incoming WebSocket messages related to notifications
func (h *NotificationHandler) HandleNotification(client *Client, message WebSocketMessage) error {
	switch message.Action {
	case "mark_read":
		return h.handleMarkAsRead(client, message)
	case "mark_all_read":
		return h.handleMarkAllAsRead(client, message)
	case "get_notifications":
		return h.handleGetNotifications(client, message)
	case "get_unread_count":
		return h.handleGetUnreadCount(client, message)
	case "dismiss":
		return h.handleDismissNotification(client, message)
	case "subscribe_to_type":
		return h.handleSubscribeToType(client, message)
	case "update_preferences":
		return h.handleUpdatePreferences(client, message)
	default:
		return fmt.Errorf("unknown notification action: %s", message.Action)
	}
}

// handleMarkAsRead handles marking a notification as read
func (h *NotificationHandler) handleMarkAsRead(client *Client, wsMessage WebSocketMessage) error {
	notificationID, ok := wsMessage.Data["notification_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing notification_id")
	}

	// Validate notification ID
	notificationObjectID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_NOTIFICATION", "Invalid notification ID")
	}

	// Update notification as read
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		},
	}

	filter := bson.M{
		"_id":          notificationObjectID,
		"recipient_id": client.UserID,
	}

	result, err := h.notificationsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to mark notification as read")
	}

	if result.MatchedCount == 0 {
		return h.sendError(client, wsMessage.RequestID, "NOTIFICATION_NOT_FOUND", "Notification not found or unauthorized")
	}

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "marked_read",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"notification_id": notificationID,
			"read_at":         now,
		},
	}

	return client.SendMessage(confirmMessage)
}

// handleMarkAllAsRead handles marking all notifications as read for a user
func (h *NotificationHandler) handleMarkAllAsRead(client *Client, wsMessage WebSocketMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		},
	}

	filter := bson.M{
		"recipient_id": client.UserID,
		"is_read":      false,
	}

	result, err := h.notificationsColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to mark notifications as read")
	}

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "all_marked_read",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"marked_count": result.ModifiedCount,
			"read_at":      now,
		},
	}

	return client.SendMessage(confirmMessage)
}

// handleGetNotifications handles fetching notifications for a user
func (h *NotificationHandler) handleGetNotifications(client *Client, wsMessage WebSocketMessage) error {
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

	// Filter by type if specified
	var typeFilter interface{}
	if notifType, exists := wsMessage.Data["type"]; exists {
		typeFilter = notifType
	}

	// Fetch notifications
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"recipient_id": client.UserID,
	}

	if typeFilter != nil {
		filter["type"] = typeFilter
	}

	// Check if only unread notifications are requested
	if unreadOnly, exists := wsMessage.Data["unread_only"]; exists {
		if unreadBool, ok := unreadOnly.(bool); ok && unreadBool {
			filter["is_read"] = false
		}
	}

	skip := (page - 1) * limit
	cursor, err := h.notificationsColl.Find(ctx, filter, &options.FindOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Skip:  &skip,
		Limit: &limit,
	})

	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to fetch notifications")
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to decode notifications")
	}

	// Convert to response format
	notificationResponses := make([]map[string]interface{}, len(notifications))
	for i, notif := range notifications {
		notificationResponses[i] = h.notificationToMap(notif)
	}

	// Send response
	responseMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "notifications",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"notifications": notificationResponses,
			"page":          page,
			"limit":         limit,
			"has_more":      len(notifications) == int(limit),
		},
	}

	return client.SendMessage(responseMessage)
}

// handleGetUnreadCount handles getting unread notification count
func (h *NotificationHandler) handleGetUnreadCount(client *Client, wsMessage WebSocketMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count unread notifications
	unreadCount, err := h.notificationsColl.CountDocuments(ctx, bson.M{
		"recipient_id": client.UserID,
		"is_read":      false,
	})

	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to count notifications")
	}

	// Count by type if needed
	typeCounts := make(map[string]int64)
	if includeTypes, exists := wsMessage.Data["include_type_counts"]; exists {
		if includeBool, ok := includeTypes.(bool); ok && includeBool {
			// Aggregate by type
			pipeline := []bson.M{
				{"$match": bson.M{
					"recipient_id": client.UserID,
					"is_read":      false,
				}},
				{"$group": bson.M{
					"_id":   "$type",
					"count": bson.M{"$sum": 1},
				}},
			}

			cursor, err := h.notificationsColl.Aggregate(ctx, pipeline)
			if err == nil {
				defer cursor.Close(ctx)

				var results []struct {
					Type  models.NotificationType `bson:"_id"`
					Count int64                   `bson:"count"`
				}

				if cursor.All(ctx, &results) == nil {
					for _, result := range results {
						typeCounts[string(result.Type)] = result.Count
					}
				}
			}
		}
	}

	// Send response
	responseMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "unread_count",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"total_unread": unreadCount,
			"by_type":      typeCounts,
		},
	}

	return client.SendMessage(responseMessage)
}

// handleDismissNotification handles dismissing a notification
func (h *NotificationHandler) handleDismissNotification(client *Client, wsMessage WebSocketMessage) error {
	notificationID, ok := wsMessage.Data["notification_id"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing notification_id")
	}

	// Validate notification ID
	notificationObjectID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "INVALID_NOTIFICATION", "Invalid notification ID")
	}

	// Update notification as dismissed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"dismissed_at": now,
			"updated_at":   now,
		},
	}

	filter := bson.M{
		"_id":          notificationObjectID,
		"recipient_id": client.UserID,
	}

	result, err := h.notificationsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to dismiss notification")
	}

	if result.MatchedCount == 0 {
		return h.sendError(client, wsMessage.RequestID, "NOTIFICATION_NOT_FOUND", "Notification not found or unauthorized")
	}

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "dismissed",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"notification_id": notificationID,
			"dismissed_at":    now,
		},
	}

	return client.SendMessage(confirmMessage)
}

// handleSubscribeToType handles subscribing to specific notification types
func (h *NotificationHandler) handleSubscribeToType(client *Client, wsMessage WebSocketMessage) error {
	notificationType, ok := wsMessage.Data["type"].(string)
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing notification type")
	}

	// Subscribe client to notification type channel
	channel := "notifications:" + notificationType
	client.Subscribe(channel)

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "subscribed",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"type":    notificationType,
			"channel": channel,
		},
	}

	return client.SendMessage(confirmMessage)
}

// handleUpdatePreferences handles updating notification preferences
func (h *NotificationHandler) handleUpdatePreferences(client *Client, wsMessage WebSocketMessage) error {
	preferences, ok := wsMessage.Data["preferences"].(map[string]interface{})
	if !ok {
		return h.sendError(client, wsMessage.RequestID, "INVALID_DATA", "Missing preferences")
	}

	// Update user notification preferences in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convert preferences to BSON
	updateDoc := bson.M{}
	for key, value := range preferences {
		updateDoc["notification_settings."+key] = value
	}
	updateDoc["updated_at"] = time.Now()

	_, err := h.usersColl.UpdateOne(ctx, bson.M{"_id": client.UserID}, bson.M{"$set": updateDoc})
	if err != nil {
		return h.sendError(client, wsMessage.RequestID, "DATABASE_ERROR", "Failed to update preferences")
	}

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:      "notification",
		Action:    "preferences_updated",
		RequestID: wsMessage.RequestID,
		Data: map[string]interface{}{
			"preferences": preferences,
		},
	}

	return client.SendMessage(confirmMessage)
}

// SendNotificationToUser sends a notification to a specific user
func (h *NotificationHandler) SendNotificationToUser(userID string, notification models.Notification) error {
	// Convert notification to WebSocket message
	notificationMessage := WebSocketMessage{
		Type:   "notification",
		Action: "new",
		Data:   h.notificationToMap(notification),
	}

	// Check if user is online
	if h.hub.IsUserOnline(userID) {
		// Send immediately to online user
		h.hub.BroadcastToUser(userID, notificationMessage)

		// Update delivery stats
		h.updateDeliveryStats(string(notification.Type), true)

		// Mark as delivered in database
		go h.markNotificationAsDelivered(notification.ID)

		return nil
	}

	// User is offline, notification will be delivered when they come online
	h.updateDeliveryStats(string(notification.Type), false)
	return nil
}

// BroadcastSystemNotification broadcasts a system notification to all users
func (h *NotificationHandler) BroadcastSystemNotification(notification models.Notification) error {
	systemMessage := WebSocketMessage{
		Type:   "notification",
		Action: "system",
		Data:   h.notificationToMap(notification),
	}

	h.hub.BroadcastToAll(systemMessage)
	return nil
}

// SendNotificationsByType sends notifications to users subscribed to a specific type
func (h *NotificationHandler) SendNotificationsByType(notificationType models.NotificationType, notification models.Notification) error {
	typeMessage := WebSocketMessage{
		Type:   "notification",
		Action: "type_broadcast",
		Data:   h.notificationToMap(notification),
	}

	channel := "notifications:" + string(notificationType)
	h.hub.BroadcastToChannel(channel, typeMessage, primitive.NilObjectID)
	return nil
}

// SendBulkNotifications sends notifications to multiple users
func (h *NotificationHandler) SendBulkNotifications(userIDs []string, notification models.Notification) error {
	bulkMessage := WebSocketMessage{
		Type:   "notification",
		Action: "bulk",
		Data:   h.notificationToMap(notification),
	}

	h.hub.BroadcastToUsers(userIDs, bulkMessage, "")

	// Update delivery stats for each user
	for _, userID := range userIDs {
		isOnline := h.hub.IsUserOnline(userID)
		h.updateDeliveryStats(string(notification.Type), isOnline)
	}

	return nil
}

// SendLiveUpdate sends live updates for activities (likes, comments, etc.)
func (h *NotificationHandler) SendLiveUpdate(targetUserID string, updateType string, data map[string]interface{}) error {
	updateMessage := WebSocketMessage{
		Type:   "live_update",
		Action: updateType,
		Data:   data,
	}

	if h.hub.IsUserOnline(targetUserID) {
		h.hub.BroadcastToUser(targetUserID, updateMessage)
	}

	return nil
}

// SendTypingIndicator sends typing indicators for conversations
func (h *NotificationHandler) SendTypingIndicator(conversationID string, userID, username string, isTyping bool) error {
	typingMessage := WebSocketMessage{
		Type:   "typing",
		Action: "indicator",
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"user_id":         userID,
			"username":        username,
			"is_typing":       isTyping,
			"timestamp":       time.Now(),
		},
	}

	channel := "conversation:" + conversationID
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	h.hub.BroadcastToChannel(channel, typingMessage, userObjectID)

	return nil
}

// SendPresenceUpdate sends user presence/online status updates
func (h *NotificationHandler) SendPresenceUpdate(userID string, status string, lastSeen *time.Time) error {
	presenceMessage := WebSocketMessage{
		Type:   "presence",
		Action: "update",
		Data: map[string]interface{}{
			"user_id":   userID,
			"status":    status,
			"last_seen": lastSeen,
			"timestamp": time.Now(),
		},
	}

	// Broadcast to users who follow this user or are in conversations with them
	// This would require additional logic to determine relevant users
	h.hub.BroadcastToChannel("presence:"+userID, presenceMessage, primitive.NilObjectID)

	return nil
}

// SendStoryViewNotification sends story view notifications
func (h *NotificationHandler) SendStoryViewNotification(storyOwnerID string, viewerID, viewerUsername string, storyID string) error {
	storyViewMessage := WebSocketMessage{
		Type:   "story",
		Action: "viewed",
		Data: map[string]interface{}{
			"story_id": storyID,
			"viewer": map[string]interface{}{
				"user_id":  viewerID,
				"username": viewerUsername,
			},
			"timestamp": time.Now(),
		},
	}

	if h.hub.IsUserOnline(storyOwnerID) {
		h.hub.BroadcastToUser(storyOwnerID, storyViewMessage)
	}

	return nil
}

// SendGroupActivityNotification sends group activity notifications
func (h *NotificationHandler) SendGroupActivityNotification(groupID string, activityType string, data map[string]interface{}) error {
	groupMessage := WebSocketMessage{
		Type:   "group",
		Action: activityType,
		Data:   data,
	}

	channel := "group:" + groupID
	h.hub.BroadcastToChannel(channel, groupMessage, primitive.NilObjectID)

	return nil
}

// SendEventReminder sends event reminder notifications
func (h *NotificationHandler) SendEventReminder(attendeeIDs []string, eventData map[string]interface{}) error {
	reminderMessage := WebSocketMessage{
		Type:   "event",
		Action: "reminder",
		Data:   eventData,
	}

	h.hub.BroadcastToUsers(attendeeIDs, reminderMessage, "")
	return nil
}

// Helper methods

// sendError sends an error message to the client
func (h *NotificationHandler) sendError(client *Client, requestID, errorCode, message string) error {
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

// notificationToMap converts a notification model to a map for WebSocket transmission
func (h *NotificationHandler) notificationToMap(notification models.Notification) map[string]interface{} {
	data := map[string]interface{}{
		"id":          notification.ID.Hex(),
		"type":        notification.Type,
		"title":       notification.Title,
		"message":     notification.Message,
		"action_text": notification.ActionText,
		"is_read":     notification.IsRead,
		"read_at":     notification.ReadAt,
		"created_at":  notification.CreatedAt,
		"priority":    notification.Priority,
		"group_count": notification.GroupCount,
		"is_grouped":  notification.IsGrouped,
	}

	// Add actor information
	if notification.ActorID != primitive.NilObjectID {
		data["actor_id"] = notification.ActorID.Hex()
	}

	// Add target information
	if notification.TargetID != nil {
		data["target_id"] = notification.TargetID.Hex()
		data["target_type"] = notification.TargetType
		data["target_url"] = notification.TargetURL
	}

	// Add metadata
	if notification.Metadata != nil {
		data["metadata"] = notification.Metadata
	}

	// Add display properties
	icon, color := h.getDisplayProperties(notification.Type)
	data["icon"] = icon
	data["color"] = color

	return data
}

// getDisplayProperties returns display properties for notification types
func (h *NotificationHandler) getDisplayProperties(notifType models.NotificationType) (string, string) {
	switch notifType {
	case models.NotificationLike:
		return "👍", "#1877F2"
	case models.NotificationLove:
		return "❤️", "#E41E3F"
	case models.NotificationComment:
		return "💬", "#42B883"
	case models.NotificationFollow:
		return "👤", "#8B5CF6"
	case models.NotificationMessage:
		return "✉️", "#06B6D4"
	case models.NotificationMention:
		return "📢", "#F59E0B"
	case models.NotificationGroupInvite:
		return "👥", "#10B981"
	case models.NotificationEventInvite:
		return "📅", "#EF4444"
	case models.NotificationFriendRequest:
		return "🤝", "#6366F1"
	case models.NotificationPostShare:
		return "🔄", "#84CC16"
	case models.NotificationStoryView:
		return "👁️", "#EC4899"
	case models.NotificationGroupPost:
		return "📝", "#F97316"
	case models.NotificationEventReminder:
		return "⏰", "#D97706"
	default:
		return "🔔", "#6B7280"
	}
}

// markNotificationAsDelivered marks a notification as delivered in the database
func (h *NotificationHandler) markNotificationAsDelivered(notificationID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_delivered": true,
			"delivered_at": now,
			"updated_at":   now,
		},
	}

	_, err := h.notificationsColl.UpdateOne(ctx, bson.M{"_id": notificationID}, update)
	if err != nil {
		log.Printf("Failed to mark notification as delivered: %v", err)
	}
}

// updateDeliveryStats updates notification delivery statistics
func (h *NotificationHandler) updateDeliveryStats(notificationType string, delivered bool) {
	if h.deliveryStats[notificationType] == nil {
		h.deliveryStats[notificationType] = &NotificationDeliveryStats{}
	}

	stats := h.deliveryStats[notificationType]
	stats.TotalSent++
	stats.LastSent = time.Now()

	if delivered {
		stats.TotalDelivered++
	} else {
		stats.FailureCount++
	}
}

// GetDeliveryStats returns notification delivery statistics
func (h *NotificationHandler) GetDeliveryStats() map[string]*NotificationDeliveryStats {
	statsCopy := make(map[string]*NotificationDeliveryStats)
	for key, stats := range h.deliveryStats {
		statsCopy[key] = &NotificationDeliveryStats{
			TotalSent:      stats.TotalSent,
			TotalDelivered: stats.TotalDelivered,
			TotalRead:      stats.TotalRead,
			LastSent:       stats.LastSent,
			FailureCount:   stats.FailureCount,
		}
	}
	return statsCopy
}

// ProcessNotificationQueue processes queued notifications for offline users
func (h *NotificationHandler) ProcessNotificationQueue() {
	// This would typically run as a background job
	// to process notifications for users who come online

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find undelivered notifications for online users
	onlineUsers := h.hub.GetConnectedUsers()
	if len(onlineUsers) == 0 {
		return
	}

	// Convert to ObjectIDs
	userObjectIDs := make([]primitive.ObjectID, 0, len(onlineUsers))
	for _, userID := range onlineUsers {
		if objectID, err := primitive.ObjectIDFromHex(userID); err == nil {
			userObjectIDs = append(userObjectIDs, objectID)
		}
	}

	// Find undelivered notifications
	cursor, err := h.notificationsColl.Find(ctx, bson.M{
		"recipient_id": bson.M{"$in": userObjectIDs},
		"is_delivered": false,
		"created_at":   bson.M{"$gte": time.Now().Add(-24 * time.Hour)}, // Last 24 hours
	})

	if err != nil {
		log.Printf("Failed to fetch undelivered notifications: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		log.Printf("Failed to decode notifications: %v", err)
		return
	}

	// Send notifications to online users
	for _, notification := range notifications {
		userID := notification.RecipientID.Hex()
		if h.hub.IsUserOnline(userID) {
			if err := h.SendNotificationToUser(userID, notification); err != nil {
				log.Printf("Failed to send queued notification: %v", err)
			}
		}
	}
}

// CreateNotificationFromActivity creates and sends a notification from an activity
func (h *NotificationHandler) CreateNotificationFromActivity(activityType string, actorID, recipientID primitive.ObjectID, targetID *primitive.ObjectID, metadata map[string]interface{}) error {
	// Determine notification type
	var notifType models.NotificationType
	switch activityType {
	case "like":
		notifType = models.NotificationLike
	case "comment":
		notifType = models.NotificationComment
	case "follow":
		notifType = models.NotificationFollow
	case "mention":
		notifType = models.NotificationMention
	case "share":
		notifType = models.NotificationPostShare
	default:
		return fmt.Errorf("unknown activity type: %s", activityType)
	}

	// Create notification
	notification := models.CreateNotificationFromTemplate(notifType, recipientID, actorID, targetID, metadata)
	notification.BeforeCreate()

	// Save to database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.notificationsColl.InsertOne(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)

	// Send to user
	return h.SendNotificationToUser(recipientID.Hex(), *notification)
}

// CleanupOldNotifications removes old notifications to keep the database clean
func (h *NotificationHandler) CleanupOldNotifications(olderThan time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Delete old notifications
	filter := bson.M{
		"created_at": bson.M{"$lt": olderThan},
		"is_read":    true, // Only delete read notifications
	}

	result, err := h.notificationsColl.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to cleanup old notifications: %w", err)
	}

	log.Printf("Cleaned up %d old notifications", result.DeletedCount)
	return nil
}

// GetNotificationPreferences gets user notification preferences
func (h *NotificationHandler) GetNotificationPreferences(userID primitive.ObjectID) (*models.NotificationPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := h.usersColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Convert to notification preferences
	prefs := models.DefaultNotificationPreferences(userID)
	// Map user.NotificationSettings to prefs
	// This would require implementing the mapping logic

	return &prefs, nil
}

// ValidateNotificationPermissions checks if a user can send notifications to another user
func (h *NotificationHandler) ValidateNotificationPermissions(senderID, recipientID primitive.ObjectID, notifType models.NotificationType) bool {
	// Check if users are connected (following, friends, etc.)
	// Check user's notification preferences
	// Check if sender is blocked
	// This would require implementing the permission logic

	return true // Simplified for now
}

// FormatNotificationText formats notification text with user mentions and links
func (h *NotificationHandler) FormatNotificationText(text string, mentions []string) string {
	// Replace @username mentions with clickable links
	for _, mention := range mentions {
		text = strings.ReplaceAll(text, "@"+mention, fmt.Sprintf(`<a href="/users/%s">@%s</a>`, mention, mention))
	}

	return text
}

// ScheduleNotification schedules a notification to be sent at a specific time
func (h *NotificationHandler) ScheduleNotification(notification models.Notification, sendAt time.Time) error {
	// Set scheduled time
	notification.ScheduledAt = &sendAt

	// Save to database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.notificationsColl.InsertOne(ctx, notification)
	return err
}

// ProcessScheduledNotifications processes notifications scheduled to be sent
func (h *NotificationHandler) ProcessScheduledNotifications() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find notifications scheduled to be sent now
	now := time.Now()
	cursor, err := h.notificationsColl.Find(ctx, bson.M{
		"scheduled_at": bson.M{"$lte": now},
		"is_delivered": false,
	})

	if err != nil {
		log.Printf("Failed to fetch scheduled notifications: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		log.Printf("Failed to decode scheduled notifications: %v", err)
		return
	}

	// Send scheduled notifications
	for _, notification := range notifications {
		userID := notification.RecipientID.Hex()
		if err := h.SendNotificationToUser(userID, notification); err != nil {
			log.Printf("Failed to send scheduled notification: %v", err)
		} else {
			// Mark as processed
			h.notificationsColl.UpdateOne(ctx, bson.M{"_id": notification.ID}, bson.M{
				"$unset": bson.M{"scheduled_at": ""},
			})
		}
	}
}
