// internal/services/notification_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationService struct {
	collection            *mongo.Collection
	userCollection        *mongo.Collection
	preferencesCollection *mongo.Collection
	db                    *mongo.Database
	emailService          *EmailService
	pushService           *PushService
}

func NewNotificationService(emailService *EmailService, pushService *PushService) *NotificationService {
	return &NotificationService{
		collection:            config.DB.Collection("notifications"),
		userCollection:        config.DB.Collection("users"),
		preferencesCollection: config.DB.Collection("notification_preferences"),
		db:                    config.DB,
		emailService:          emailService,
		pushService:           pushService,
	}
}

// CreateNotification creates a new notification
func (ns *NotificationService) CreateNotification(req models.CreateNotificationRequest) (*models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert IDs
	recipientID, err := primitive.ObjectIDFromHex(req.RecipientID)
	if err != nil {
		return nil, errors.New("invalid recipient ID")
	}

	actorID, err := primitive.ObjectIDFromHex(req.ActorID)
	if err != nil {
		return nil, errors.New("invalid actor ID")
	}

	var targetID *primitive.ObjectID
	if req.TargetID != "" {
		if tID, err := primitive.ObjectIDFromHex(req.TargetID); err == nil {
			targetID = &tID
		}
	}

	// Create notification
	notification := &models.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        req.Type,
		Title:       req.Title,
		Message:     req.Message,
		ActionText:  req.ActionText,
		TargetID:    targetID,
		TargetType:  req.TargetType,
		TargetURL:   req.TargetURL,
		Metadata:    req.Metadata,
		Priority:    req.Priority,
		ScheduledAt: req.ScheduledAt,
		ExpiresAt:   req.ExpiresAt,
	}

	notification.BeforeCreate()

	// Insert notification
	result, err := ns.collection.InsertOne(ctx, notification)
	if err != nil {
		return nil, err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)

	// Get user preferences
	prefs, err := ns.GetUserPreferences(recipientID)
	if err != nil {
		// Use default preferences if not found
		prefs = models.DefaultNotificationPreferences(recipientID)
	}

	// Send notification through various channels
	go ns.sendNotificationChannels(notification, prefs, req.SendViaEmail, req.SendViaPush, req.SendViaSMS)

	return notification, nil
}

// CreateBulkNotifications creates multiple notifications
func (ns *NotificationService) CreateBulkNotifications(req models.BulkCreateNotificationRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	actorID, err := primitive.ObjectIDFromHex(req.ActorID)
	if err != nil {
		return errors.New("invalid actor ID")
	}

	var targetID *primitive.ObjectID
	if req.TargetID != "" {
		if tID, err := primitive.ObjectIDFromHex(req.TargetID); err == nil {
			targetID = &tID
		}
	}

	var notifications []interface{}
	for _, recipientIDStr := range req.RecipientIDs {
		recipientID, err := primitive.ObjectIDFromHex(recipientIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}

		notification := &models.Notification{
			RecipientID: recipientID,
			ActorID:     actorID,
			Type:        req.Type,
			Title:       req.Title,
			Message:     req.Message,
			ActionText:  req.ActionText,
			TargetID:    targetID,
			TargetType:  req.TargetType,
			TargetURL:   req.TargetURL,
			Metadata:    req.Metadata,
			Priority:    req.Priority,
			ScheduledAt: req.ScheduledAt,
		}

		notification.BeforeCreate()
		notifications = append(notifications, notification)
	}

	if len(notifications) == 0 {
		return errors.New("no valid recipient IDs")
	}

	// Insert all notifications
	_, err = ns.collection.InsertMany(ctx, notifications)
	if err != nil {
		return err
	}

	// Send notifications asynchronously
	go ns.sendBulkNotifications(notifications, req.SendViaEmail, req.SendViaPush, req.SendViaSMS)

	return nil
}

// GetUserNotifications retrieves notifications for a user
func (ns *NotificationService) GetUserNotifications(userID primitive.ObjectID, limit, skip int, unreadOnly bool) ([]models.NotificationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"recipient_id": userID,
		"$or": []bson.M{
			{"expires_at": bson.M{"$exists": false}},
			{"expires_at": bson.M{"$gt": time.Now()}},
		},
	}

	if unreadOnly {
		filter["is_read"] = false
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ns.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	// Convert to response format and populate actor information
	var responses []models.NotificationResponse
	for _, notif := range notifications {
		response := notif.ToNotificationResponse()

		// Populate actor information
		if actor, err := ns.getUserByID(notif.ActorID); err == nil {
			response.Actor = actor.ToUserResponse()
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// MarkAsRead marks notifications as read
func (ns *NotificationService) MarkAsRead(userID primitive.ObjectID, notificationIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, idStr := range notificationIDs {
		if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	if len(objectIDs) == 0 {
		return errors.New("no valid notification IDs")
	}

	filter := bson.M{
		"_id":          bson.M{"$in": objectIDs},
		"recipient_id": userID,
		"is_read":      false,
	}

	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"read_at":    time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err := ns.collection.UpdateMany(ctx, filter, update)
	return err
}

// MarkAllAsRead marks all notifications as read for a user
func (ns *NotificationService) MarkAllAsRead(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"recipient_id": userID,
		"is_read":      false,
	}

	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"read_at":    time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err := ns.collection.UpdateMany(ctx, filter, update)
	return err
}

// DeleteNotifications deletes notifications
func (ns *NotificationService) DeleteNotifications(userID primitive.ObjectID, notificationIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, idStr := range notificationIDs {
		if objID, err := primitive.ObjectIDFromHex(idStr); err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	if len(objectIDs) == 0 {
		return errors.New("no valid notification IDs")
	}

	filter := bson.M{
		"_id":          bson.M{"$in": objectIDs},
		"recipient_id": userID,
	}

	_, err := ns.collection.DeleteMany(ctx, filter)
	return err
}

// GetNotificationStats retrieves notification statistics for a user
func (ns *NotificationService) GetNotificationStats(userID primitive.ObjectID) (*models.NotificationStatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"recipient_id": userID,
				"$or": []bson.M{
					{"expires_at": bson.M{"$exists": false}},
					{"expires_at": bson.M{"$gt": time.Now()}},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":         nil,
				"total_count": bson.M{"$sum": 1},
				"unread_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							"$is_read", 0, 1,
						},
					},
				},
				"delivered_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							"$is_delivered", 1, 0,
						},
					},
				},
				"clicked_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$ne": []interface{}{"$clicked_at", nil}}, 1, 0,
						},
					},
				},
				"dismissed_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$ne": []interface{}{"$dismissed_at", nil}}, 1, 0,
						},
					},
				},
				"recent_count": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$gte": []interface{}{"$created_at", time.Now().Add(-24 * time.Hour)}}, 1, 0,
						},
					},
				},
			},
		},
	}

	cursor, err := ns.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalCount     int64 `bson:"total_count"`
		UnreadCount    int64 `bson:"unread_count"`
		DeliveredCount int64 `bson:"delivered_count"`
		ClickedCount   int64 `bson:"clicked_count"`
		DismissedCount int64 `bson:"dismissed_count"`
		RecentCount    int64 `bson:"recent_count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &models.NotificationStatsResponse{}, nil
	}

	result := results[0]
	return &models.NotificationStatsResponse{
		TotalCount:     result.TotalCount,
		UnreadCount:    result.UnreadCount,
		ReadCount:      result.TotalCount - result.UnreadCount,
		DeliveredCount: result.DeliveredCount,
		ClickedCount:   result.ClickedCount,
		DismissedCount: result.DismissedCount,
		RecentCount:    result.RecentCount,
	}, nil
}

// GetUserPreferences retrieves user notification preferences
func (ns *NotificationService) GetUserPreferences(userID primitive.ObjectID) (models.NotificationPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var prefs models.NotificationPreferences
	err := ns.preferencesCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&prefs)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.DefaultNotificationPreferences(userID), nil
		}
		return prefs, err
	}

	return prefs, nil
}

// UpdateUserPreferences updates user notification preferences
func (ns *NotificationService) UpdateUserPreferences(userID primitive.ObjectID, prefs models.NotificationPreferences) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prefs.UserID = userID

	opts := options.Replace().SetUpsert(true)
	_, err := ns.preferencesCollection.ReplaceOne(ctx, bson.M{"user_id": userID}, prefs, opts)
	return err
}

// SendRealTimeNotification sends a real-time notification (WebSocket)
func (ns *NotificationService) SendRealTimeNotification(userID primitive.ObjectID, notification *models.Notification) error {
	// This would integrate with WebSocket service to send real-time notifications
	// Implementation depends on WebSocket architecture
	return nil
}

// Convenience methods for creating common notification types

// NotifyLike creates a like notification
func (ns *NotificationService) NotifyLike(actorID, recipientID, postID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil // Don't notify users about their own actions
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationLike,
		Title:       "New Like",
		Message:     "Someone liked your post",
		ActionText:  "View Post",
		TargetID:    postID.Hex(),
		TargetType:  "post",
		TargetURL:   "/posts/" + postID.Hex(),
		Priority:    "medium",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyComment creates a comment notification
func (ns *NotificationService) NotifyComment(actorID, recipientID, postID, commentID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationComment,
		Title:       "New Comment",
		Message:     "Someone commented on your post",
		ActionText:  "View Comment",
		TargetID:    commentID.Hex(),
		TargetType:  "comment",
		TargetURL:   "/posts/" + postID.Hex() + "#comment-" + commentID.Hex(),
		Priority:    "medium",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyFollow creates a follow notification
func (ns *NotificationService) NotifyFollow(actorID, recipientID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationFollow,
		Title:       "New Follower",
		Message:     "Someone started following you",
		ActionText:  "View Profile",
		TargetID:    actorID.Hex(),
		TargetType:  "user",
		TargetURL:   "/users/" + actorID.Hex(),
		Priority:    "medium",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// Add these methods to the existing NotificationService in internal/services/notification_service.go
// Also add "fmt" import if not already present

// NotifyGroupInvite creates a group invitation notification
func (ns *NotificationService) NotifyGroupInvite(actorID, recipientID, groupID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID:  recipientID.Hex(),
		ActorID:      actorID.Hex(),
		Type:         models.NotificationGroupInvite,
		Title:        "Group Invitation",
		Message:      "You've been invited to join a group",
		ActionText:   "View Invitation",
		TargetID:     groupID.Hex(),
		TargetType:   "group",
		TargetURL:    "/groups/" + groupID.Hex(),
		Priority:     "medium",
		SendViaPush:  true,
		SendViaEmail: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyGroupJoinRequest creates a join request notification for group admins
func (ns *NotificationService) NotifyGroupJoinRequest(actorID, recipientID, groupID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationGroupPost, // Using existing type or create new one
		Title:       "New Join Request",
		Message:     "Someone wants to join your group",
		ActionText:  "Review Request",
		TargetID:    groupID.Hex(),
		TargetType:  "group",
		TargetURL:   "/groups/" + groupID.Hex() + "/members/pending",
		Priority:    "medium",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyGroupAccepted creates a notification when join request is approved
func (ns *NotificationService) NotifyGroupAccepted(actorID, recipientID, groupID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationGroupPost,
		Title:       "Join Request Approved",
		Message:     "Your request to join the group has been approved",
		ActionText:  "View Group",
		TargetID:    groupID.Hex(),
		TargetType:  "group",
		TargetURL:   "/groups/" + groupID.Hex(),
		Priority:    "medium",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyGroupRoleChanged creates a notification when member role is changed
func (ns *NotificationService) NotifyGroupRoleChanged(actorID, recipientID, groupID primitive.ObjectID, newRole string) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationGroupPost,
		Title:       "Role Updated",
		Message:     fmt.Sprintf("Your role in the group has been changed to %s", newRole),
		ActionText:  "View Group",
		TargetID:    groupID.Hex(),
		TargetType:  "group",
		TargetURL:   "/groups/" + groupID.Hex(),
		Priority:    "medium",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyGroupPostCreated notifies group members about new posts
func (ns *NotificationService) NotifyGroupPostCreated(actorID, groupID, postID primitive.ObjectID, memberIDs []primitive.ObjectID) error {
	recipientIDs := make([]string, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		if memberID != actorID { // Don't notify the author
			recipientIDs = append(recipientIDs, memberID.Hex())
		}
	}

	if len(recipientIDs) == 0 {
		return nil
	}

	req := models.BulkCreateNotificationRequest{
		RecipientIDs: recipientIDs,
		ActorID:      actorID.Hex(),
		Type:         models.NotificationGroupPost,
		Title:        "New Group Post",
		Message:      "Someone posted in your group",
		ActionText:   "View Post",
		TargetID:     postID.Hex(),
		TargetType:   "post",
		TargetURL:    "/posts/" + postID.Hex(),
		Priority:     "low",
		SendViaPush:  true,
	}

	return ns.CreateBulkNotifications(req)
}

// NotifyMention creates a mention notification
func (ns *NotificationService) NotifyMention(actorID, recipientID, contentID primitive.ObjectID, contentType string) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID:  recipientID.Hex(),
		ActorID:      actorID.Hex(),
		Type:         models.NotificationMention,
		Title:        "You were mentioned",
		Message:      "Someone mentioned you in a " + contentType,
		ActionText:   "View " + contentType,
		TargetID:     contentID.Hex(),
		TargetType:   contentType,
		TargetURL:    "/" + contentType + "s/" + contentID.Hex(),
		Priority:     "high",
		SendViaPush:  true,
		SendViaEmail: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// NotifyMessage creates a message notification
func (ns *NotificationService) NotifyMessage(actorID, recipientID, conversationID primitive.ObjectID) error {
	if actorID == recipientID {
		return nil
	}

	req := models.CreateNotificationRequest{
		RecipientID: recipientID.Hex(),
		ActorID:     actorID.Hex(),
		Type:        models.NotificationMessage,
		Title:       "New Message",
		Message:     "You have a new message",
		ActionText:  "Read Message",
		TargetID:    conversationID.Hex(),
		TargetType:  "conversation",
		TargetURL:   "/messages/" + conversationID.Hex(),
		Priority:    "high",
		SendViaPush: true,
	}

	_, err := ns.CreateNotification(req)
	return err
}

// CleanupExpiredNotifications removes expired notifications
func (ns *NotificationService) CleanupExpiredNotifications() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	}

	_, err := ns.collection.DeleteMany(ctx, filter)
	return err
}

// Helper methods

func (ns *NotificationService) sendNotificationChannels(notification *models.Notification, prefs models.NotificationPreferences, sendEmail, sendPush, sendSMS bool) {
	// Send via email
	if sendEmail && notification.ShouldSendViaEmail(prefs) && ns.emailService != nil {
		ns.emailService.SendNotificationEmail(notification)
		ns.markAsSent(notification.ID, "email")
	}

	// Send via push
	if sendPush && notification.ShouldSendViaPush(prefs) && ns.pushService != nil {
		ns.pushService.SendPushNotification(notification)
		ns.markAsSent(notification.ID, "push")
	}

	// Send via SMS
	if sendSMS && notification.ShouldSendViaSMS(prefs) {
		// SMS implementation would go here
		ns.markAsSent(notification.ID, "sms")
	}

	// Send real-time notification
	ns.SendRealTimeNotification(notification.RecipientID, notification)

	// Mark as delivered
	ns.markAsDelivered(notification.ID)
}

func (ns *NotificationService) sendBulkNotifications(notifications []interface{}, sendEmail, sendPush, sendSMS bool) {
	for _, notifInterface := range notifications {
		notification := notifInterface.(*models.Notification)

		// Get user preferences
		prefs, err := ns.GetUserPreferences(notification.RecipientID)
		if err != nil {
			prefs = models.DefaultNotificationPreferences(notification.RecipientID)
		}

		ns.sendNotificationChannels(notification, prefs, sendEmail, sendPush, sendSMS)
	}
}

func (ns *NotificationService) markAsSent(notificationID primitive.ObjectID, channel string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"sent_via_" + channel: true,
			"updated_at":          time.Now(),
		},
	}

	ns.collection.UpdateOne(ctx, bson.M{"_id": notificationID}, update)
}

func (ns *NotificationService) markAsDelivered(notificationID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_delivered": true,
			"delivered_at": now,
			"updated_at":   now,
		},
	}

	ns.collection.UpdateOne(ctx, bson.M{"_id": notificationID}, update)
}

func (ns *NotificationService) getUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := ns.userCollection.FindOne(ctx, bson.M{
		"_id":        userID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	return &user, err
}
