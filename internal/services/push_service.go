// internal/services/push_service.go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PushService struct {
	fcmServerKey    string
	fcmSenderID     string
	apnsKeyID       string
	apnsTeamID      string
	apnsBundleID    string
	apnsKey         []byte
	userCollection  *mongo.Collection
	tokenCollection *mongo.Collection
	db              *mongo.Database
}

type FCMMessage struct {
	To              string                 `json:"to,omitempty"`
	RegistrationIds []string               `json:"registration_ids,omitempty"`
	Notification    FCMNotification        `json:"notification"`
	Data            map[string]interface{} `json:"data,omitempty"`
	Priority        string                 `json:"priority,omitempty"`
	TimeToLive      int                    `json:"time_to_live,omitempty"`
}

type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Sound string `json:"sound,omitempty"`
	Badge int    `json:"badge,omitempty"`
	Tag   string `json:"tag,omitempty"`
	Color string `json:"color,omitempty"`
}

type FCMResponse struct {
	MulticastID  int64       `json:"multicast_id"`
	Success      int         `json:"success"`
	Failure      int         `json:"failure"`
	CanonicalIds int         `json:"canonical_ids"`
	Results      []FCMResult `json:"results"`
}

type FCMResult struct {
	MessageID      string `json:"message_id,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

type PushToken struct {
	models.BaseModel `bson:",inline"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	Token            string             `json:"token" bson:"token"`
	Platform         string             `json:"platform" bson:"platform"` // ios, android, web
	DeviceInfo       string             `json:"device_info" bson:"device_info"`
	IsActive         bool               `json:"is_active" bson:"is_active"`
	LastUsedAt       time.Time          `json:"last_used_at" bson:"last_used_at"`
}

func NewPushService(fcmServerKey, fcmSenderID, apnsKeyID, apnsTeamID, apnsBundleID string, apnsKey []byte) *PushService {
	return &PushService{
		fcmServerKey:    fcmServerKey,
		fcmSenderID:     fcmSenderID,
		apnsKeyID:       apnsKeyID,
		apnsTeamID:      apnsTeamID,
		apnsBundleID:    apnsBundleID,
		apnsKey:         apnsKey,
		userCollection:  config.DB.Collection("users"),
		tokenCollection: config.DB.Collection("push_tokens"),
		db:              config.DB,
	}
}

// SendPushNotification sends a push notification
func (ps *PushService) SendPushNotification(notification *models.Notification) error {
	// Get user's push tokens
	tokens, err := ps.GetUserPushTokens(notification.RecipientID)
	if err != nil {
		log.Printf("Failed to get push tokens for user %s: %v", notification.RecipientID.Hex(), err)
		return err
	}

	if len(tokens) == 0 {
		log.Printf("No push tokens found for user %s", notification.RecipientID.Hex())
		return nil
	}

	// Create push message
	pushData := ps.createPushData(notification)

	// Send to different platforms
	var errors []error
	for _, token := range tokens {
		switch token.Platform {
		case "android", "web":
			err := ps.sendFCMNotification(token.Token, pushData)
			if err != nil {
				errors = append(errors, err)
				// Check if token is invalid and remove it
				if ps.isInvalidTokenError(err) {
					go ps.RemovePushToken(token.Token)
				}
			}
		case "ios":
			err := ps.sendAPNSNotification(token.Token, pushData)
			if err != nil {
				errors = append(errors, err)
				if ps.isInvalidTokenError(err) {
					go ps.RemovePushToken(token.Token)
				}
			}
		}
	}

	// Update last used time for valid tokens
	go ps.updateTokenLastUsed(tokens)

	// Return first error if any
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// SendBulkPushNotification sends push notifications to multiple users
func (ps *PushService) SendBulkPushNotification(userIDs []primitive.ObjectID, title, body string, data map[string]interface{}) error {
	// Get all push tokens for the users
	tokens, err := ps.GetMultipleUsersPushTokens(userIDs)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	// Group tokens by platform
	androidTokens := []string{}
	iosTokens := []string{}
	webTokens := []string{}

	for _, token := range tokens {
		switch token.Platform {
		case "android":
			androidTokens = append(androidTokens, token.Token)
		case "ios":
			iosTokens = append(iosTokens, token.Token)
		case "web":
			webTokens = append(webTokens, token.Token)
		}
	}

	pushData := map[string]interface{}{
		"title": title,
		"body":  body,
		"data":  data,
	}

	// Send to Android/Web via FCM
	if len(androidTokens) > 0 {
		go ps.sendFCMBulkNotification(androidTokens, pushData)
	}
	if len(webTokens) > 0 {
		go ps.sendFCMBulkNotification(webTokens, pushData)
	}

	// Send to iOS via APNS
	if len(iosTokens) > 0 {
		go ps.sendAPNSBulkNotification(iosTokens, pushData)
	}

	return nil
}

// RegisterPushToken registers a new push token for a user
func (ps *PushService) RegisterPushToken(userID primitive.ObjectID, token, platform, deviceInfo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if token already exists
	var existingToken PushToken
	err := ps.tokenCollection.FindOne(ctx, bson.M{
		"token": token,
	}).Decode(&existingToken)

	if err == nil {
		// Token exists, update it
		update := bson.M{
			"$set": bson.M{
				"user_id":      userID,
				"platform":     platform,
				"device_info":  deviceInfo,
				"is_active":    true,
				"last_used_at": time.Now(),
				"updated_at":   time.Now(),
			},
		}
		_, err = ps.tokenCollection.UpdateOne(ctx, bson.M{"_id": existingToken.ID}, update)
		return err
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// Create new token
	pushToken := &PushToken{
		UserID:     userID,
		Token:      token,
		Platform:   platform,
		DeviceInfo: deviceInfo,
		IsActive:   true,
		LastUsedAt: time.Now(),
	}
	pushToken.BeforeCreate()

	_, err = ps.tokenCollection.InsertOne(ctx, pushToken)
	return err
}

// RemovePushToken removes a push token
func (ps *PushService) RemovePushToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ps.tokenCollection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// GetUserPushTokens retrieves all active push tokens for a user
func (ps *PushService) GetUserPushTokens(userID primitive.ObjectID) ([]PushToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":   userID,
		"is_active": true,
	}

	cursor, err := ps.tokenCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []PushToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// GetMultipleUsersPushTokens retrieves push tokens for multiple users
func (ps *PushService) GetMultipleUsersPushTokens(userIDs []primitive.ObjectID) ([]PushToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":   bson.M{"$in": userIDs},
		"is_active": true,
	}

	cursor, err := ps.tokenCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []PushToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// SendTestNotification sends a test notification to a user
func (ps *PushService) SendTestNotification(userID primitive.ObjectID, title, body string) error {
	notification := &models.Notification{
		RecipientID: userID,
		ActorID:     userID,
		Type:        models.NotificationMessage,
		Title:       title,
		Message:     body,
		Priority:    "normal",
	}

	return ps.SendPushNotification(notification)
}

// CleanupInactiveTokens removes tokens that haven't been used recently
func (ps *PushService) CleanupInactiveTokens(daysInactive int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -daysInactive)

	filter := bson.M{
		"last_used_at": bson.M{"$lt": cutoff},
	}

	result, err := ps.tokenCollection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	log.Printf("Cleaned up %d inactive push tokens", result.DeletedCount)
	return nil
}

// Private methods

func (ps *PushService) sendFCMNotification(token string, data map[string]interface{}) error {
	url := "https://fcm.googleapis.com/fcm/send"

	message := FCMMessage{
		To: token,
		Notification: FCMNotification{
			Title: data["title"].(string),
			Body:  data["body"].(string),
			Icon:  "/icon-192x192.png",
			Sound: "default",
		},
		Data:       data["data"].(map[string]interface{}),
		Priority:   "high",
		TimeToLive: 3600, // 1 hour
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "key="+ps.fcmServerKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var fcmResp FCMResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		return err
	}

	if fcmResp.Failure > 0 && len(fcmResp.Results) > 0 {
		for _, result := range fcmResp.Results {
			if result.Error != "" {
				return fmt.Errorf("FCM error: %s", result.Error)
			}
		}
	}

	return nil
}

func (ps *PushService) sendFCMBulkNotification(tokens []string, data map[string]interface{}) error {
	url := "https://fcm.googleapis.com/fcm/send"

	// FCM allows max 1000 tokens per request
	chunkSize := 1000
	for i := 0; i < len(tokens); i += chunkSize {
		end := i + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}

		tokenChunk := tokens[i:end]

		message := FCMMessage{
			RegistrationIds: tokenChunk,
			Notification: FCMNotification{
				Title: data["title"].(string),
				Body:  data["body"].(string),
				Icon:  "/icon-192x192.png",
				Sound: "default",
			},
			Data:       data["data"].(map[string]interface{}),
			Priority:   "high",
			TimeToLive: 3600,
		}

		jsonData, err := json.Marshal(message)
		if err != nil {
			log.Printf("Failed to marshal FCM message: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to create FCM request: %v", err)
			continue
		}

		req.Header.Set("Authorization", "key="+ps.fcmServerKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to send FCM request: %v", err)
			continue
		}

		var fcmResp FCMResponse
		if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
			log.Printf("Failed to decode FCM response: %v", err)
		} else {
			log.Printf("FCM bulk send result: Success=%d, Failure=%d", fcmResp.Success, fcmResp.Failure)
		}

		resp.Body.Close()
	}

	return nil
}

func (ps *PushService) sendAPNSNotification(token string, data map[string]interface{}) error {
	// APNS implementation would go here
	// This is a placeholder - you would use a library like 'github.com/sideshow/apns2'
	log.Printf("APNS notification would be sent to token: %s", token[:10]+"...")
	return nil
}

func (ps *PushService) sendAPNSBulkNotification(tokens []string, data map[string]interface{}) error {
	// APNS bulk implementation would go here
	log.Printf("APNS bulk notification would be sent to %d tokens", len(tokens))
	return nil
}

func (ps *PushService) createPushData(notification *models.Notification) map[string]interface{} {
	data := map[string]interface{}{
		"title": notification.Title,
		"body":  notification.Message,
		"data": map[string]interface{}{
			"notification_id": notification.ID.Hex(),
			"type":            string(notification.Type),
			"target_type":     notification.TargetType,
			"target_url":      notification.TargetURL,
		},
	}

	if notification.TargetID != nil {
		data["data"].(map[string]interface{})["target_id"] = notification.TargetID.Hex()
	}

	// Add custom metadata
	if notification.Metadata != nil {
		for key, value := range notification.Metadata {
			data["data"].(map[string]interface{})[key] = value
		}
	}

	return data
}

func (ps *PushService) isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	return errorStr == "InvalidRegistration" ||
		errorStr == "NotRegistered" ||
		errorStr == "MismatchSenderId"
}

func (ps *PushService) updateTokenLastUsed(tokens []PushToken) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	for _, token := range tokens {
		ps.tokenCollection.UpdateOne(ctx, bson.M{"_id": token.ID}, bson.M{
			"$set": bson.M{
				"last_used_at": now,
				"updated_at":   now,
			},
		})
	}
}

// GetPushTokenStats returns statistics about push tokens
func (ps *PushService) GetPushTokenStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_active": true,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$platform",
				"count": bson.M{"$sum": 1},
				"active_last_week": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$gte": []interface{}{"$last_used_at", time.Now().AddDate(0, 0, -7)}},
							1,
							0,
						},
					},
				},
			},
		},
	}

	cursor, err := ps.tokenCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID             string `bson:"_id"`
		Count          int64  `bson:"count"`
		ActiveLastWeek int64  `bson:"active_last_week"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	totalTokens := int64(0)
	totalActive := int64(0)

	for _, result := range results {
		stats[result.ID] = map[string]int64{
			"total":            result.Count,
			"active_last_week": result.ActiveLastWeek,
		}
		totalTokens += result.Count
		totalActive += result.ActiveLastWeek
	}

	stats["total"] = totalTokens
	stats["active_last_week"] = totalActive

	return stats, nil
}
