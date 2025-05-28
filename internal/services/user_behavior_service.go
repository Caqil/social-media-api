// internal/services/user_behavior_service.go
package services

import (
	"context"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserBehaviorService struct {
	sessionCollection        *mongo.Collection
	engagementCollection     *mongo.Collection
	journeyCollection        *mongo.Collection
	recommendationCollection *mongo.Collection
	experimentCollection     *mongo.Collection
	db                       *mongo.Database
}

func NewUserBehaviorService() *UserBehaviorService {
	return &UserBehaviorService{
		sessionCollection:        config.DB.Collection("user_sessions"),
		engagementCollection:     config.DB.Collection("content_engagements"),
		journeyCollection:        config.DB.Collection("user_journeys"),
		recommendationCollection: config.DB.Collection("recommendation_events"),
		experimentCollection:     config.DB.Collection("experiments"),
		db:                       config.DB,
	}
}

// Session Management
func (ubs *UserBehaviorService) StartSession(userID primitive.ObjectID, sessionID, deviceInfo, ipAddress, userAgent string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session := models.UserSession{
		UserID:       userID,
		SessionID:    sessionID,
		StartTime:    time.Now(),
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		PagesVisited: []models.PageVisit{},
		Actions:      []models.UserAction{},
	}

	_, err := ubs.sessionCollection.InsertOne(ctx, session)
	return err
}

func (ubs *UserBehaviorService) EndSession(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	endTime := time.Now()

	// Get session to calculate duration
	var session models.UserSession
	err := ubs.sessionCollection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&session)
	if err != nil {
		return err
	}

	duration := endTime.Sub(session.StartTime).Milliseconds()

	update := bson.M{
		"$set": bson.M{
			"end_time": endTime,
			"duration": duration,
		},
	}

	_, err = ubs.sessionCollection.UpdateOne(ctx, bson.M{"session_id": sessionID}, update)
	return err
}

// Page Visit Tracking
func (ubs *UserBehaviorService) RecordPageVisit(userID primitive.ObjectID, sessionID string, pageVisit models.PageVisit) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{
			"pages_visited": pageVisit,
		},
	}

	_, err := ubs.sessionCollection.UpdateOne(ctx,
		bson.M{"user_id": userID, "session_id": sessionID},
		update,
	)
	return err
}

// User Action Tracking
func (ubs *UserBehaviorService) RecordUserAction(userID primitive.ObjectID, sessionID string, action models.UserAction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{
			"actions": action,
		},
	}

	_, err := ubs.sessionCollection.UpdateOne(ctx,
		bson.M{"user_id": userID, "session_id": sessionID},
		update,
	)
	return err
}

// Content Engagement Tracking
func (ubs *UserBehaviorService) RecordContentEngagement(engagement models.ContentEngagement) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engagement.ID = primitive.NewObjectID()

	_, err := ubs.engagementCollection.InsertOne(ctx, engagement)
	return err
}

// Post Interaction Tracking
func (ubs *UserBehaviorService) AutoTrackPostInteraction(userID, postID primitive.ObjectID, interactionType, source string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engagement := models.ContentEngagement{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		ContentID:   postID,
		ContentType: "post",
		ViewTime:    time.Now(),
		Source:      source,
		Interactions: []models.Interaction{
			{
				Type:      interactionType,
				Timestamp: time.Now(),
			},
		},
	}

	_, err := ubs.engagementCollection.InsertOne(ctx, engagement)
	return err
}

// Story View Tracking
func (ubs *UserBehaviorService) AutoTrackStoryView(userID, storyID primitive.ObjectID, source string, duration int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engagement := models.ContentEngagement{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		ContentID:    storyID,
		ContentType:  "story",
		ViewTime:     time.Now(),
		ViewDuration: duration,
		Source:       source,
		Interactions: []models.Interaction{
			{
				Type:      "view",
				Timestamp: time.Now(),
			},
		},
	}

	_, err := ubs.engagementCollection.InsertOne(ctx, engagement)
	return err
}

// Search Tracking
func (ubs *UserBehaviorService) AutoTrackSearch(userID primitive.ObjectID, query, searchType string, resultsCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	action := models.UserAction{
		Type:      "search",
		Target:    query,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"search_type":   searchType,
			"results_count": resultsCount,
			"auto_tracked":  true,
		},
	}

	// Find the most recent session for this user
	var session models.UserSession
	err := ubs.sessionCollection.FindOne(ctx,
		bson.M{"user_id": userID},
		options.FindOne().SetSort(bson.M{"start_time": -1}),
	).Decode(&session)

	if err != nil {
		// Create a new engagement record instead
		engagement := models.ContentEngagement{
			ID:          primitive.NewObjectID(),
			UserID:      userID,
			ContentType: "search",
			ViewTime:    time.Now(),
			Source:      "search",
			Context: map[string]interface{}{
				"query":         query,
				"search_type":   searchType,
				"results_count": resultsCount,
			},
		}
		_, err = ubs.engagementCollection.InsertOne(ctx, engagement)
		return err
	}

	return ubs.RecordUserAction(userID, session.SessionID, action)
}

// Generic Interaction Recording
func (ubs *UserBehaviorService) RecordInteraction(userID, contentID primitive.ObjectID, contentType, interactionType, source string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engagement := models.ContentEngagement{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		ContentID:   contentID,
		ContentType: contentType,
		ViewTime:    time.Now(),
		Source:      source,
		Context:     metadata,
		Interactions: []models.Interaction{
			{
				Type:      interactionType,
				Timestamp: time.Now(),
			},
		},
	}

	_, err := ubs.engagementCollection.InsertOne(ctx, engagement)
	return err
}

// Recommendation Tracking
func (ubs *UserBehaviorService) TrackRecommendation(userID primitive.ObjectID, event models.RecommendationEvent, action string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event.ID = primitive.NewObjectID()

	switch action {
	case "clicked":
		now := time.Now()
		event.Clicked = &now
	case "converted":
		now := time.Now()
		event.Converted = &now
	}

	_, err := ubs.recommendationCollection.InsertOne(ctx, event)
	return err
}

// Experiment Tracking
func (ubs *UserBehaviorService) TrackExperiment(userID primitive.ObjectID, experimentID, variantID, event string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	experiment := bson.M{
		"_id":           primitive.NewObjectID(),
		"user_id":       userID,
		"experiment_id": experimentID,
		"variant_id":    variantID,
		"event":         event,
		"value":         value,
		"timestamp":     time.Now(),
	}

	_, err := ubs.experimentCollection.InsertOne(ctx, experiment)
	return err
}

// Analytics and Insights
func (ubs *UserBehaviorService) GetUserBehaviorAnalytics(userID primitive.ObjectID, timeRange string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Calculate time filter
	var since time.Time
	switch timeRange {
	case "day":
		since = time.Now().Add(-24 * time.Hour)
	case "week":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "month":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default:
		since = time.Now().Add(-7 * 24 * time.Hour)
	}

	analytics := make(map[string]interface{})

	// Get session stats
	sessionStats, err := ubs.getSessionStats(ctx, userID, since)
	if err == nil {
		analytics["sessions"] = sessionStats
	}

	// Get engagement stats
	engagementStats, err := ubs.getEngagementStats(ctx, userID, since)
	if err == nil {
		analytics["engagement"] = engagementStats
	}

	// Get content preferences
	contentPrefs, err := ubs.getContentPreferences(ctx, userID, since)
	if err == nil {
		analytics["content_preferences"] = contentPrefs
	}

	// Get interaction patterns
	interactionPatterns, err := ubs.getInteractionPatterns(ctx, userID, since)
	if err == nil {
		analytics["interaction_patterns"] = interactionPatterns
	}

	return analytics, nil
}

// Helper methods for analytics
func (ubs *UserBehaviorService) getSessionStats(ctx context.Context, userID primitive.ObjectID, since time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":    userID,
				"start_time": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id":              nil,
				"total_sessions":   bson.M{"$sum": 1},
				"avg_duration":     bson.M{"$avg": "$duration"},
				"total_page_views": bson.M{"$sum": bson.M{"$size": "$pages_visited"}},
				"total_actions":    bson.M{"$sum": bson.M{"$size": "$actions"}},
			},
		},
	}

	cursor, err := ubs.sessionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"total_sessions":   0,
			"avg_duration":     0,
			"total_page_views": 0,
			"total_actions":    0,
		}, nil
	}

	return results[0], nil
}

func (ubs *UserBehaviorService) getEngagementStats(ctx context.Context, userID primitive.ObjectID, since time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"view_time": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id":                "$content_type",
				"total_views":        bson.M{"$sum": 1},
				"avg_view_duration":  bson.M{"$avg": "$view_duration"},
				"total_interactions": bson.M{"$sum": bson.M{"$size": "$interactions"}},
			},
		},
	}

	cursor, err := ubs.engagementCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"by_content_type": results,
	}, nil
}

func (ubs *UserBehaviorService) getContentPreferences(ctx context.Context, userID primitive.ObjectID, since time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"view_time": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id":              "$source",
				"engagement_count": bson.M{"$sum": 1},
				"avg_duration":     bson.M{"$avg": "$view_duration"},
			},
		},
		{
			"$sort": bson.M{"engagement_count": -1},
		},
		{
			"$limit": 10,
		},
	}

	cursor, err := ubs.engagementCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"preferred_sources": results,
	}, nil
}

func (ubs *UserBehaviorService) getInteractionPatterns(ctx context.Context, userID primitive.ObjectID, since time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"view_time": bson.M{"$gte": since},
			},
		},
		{
			"$unwind": "$interactions",
		},
		{
			"$group": bson.M{
				"_id":   "$interactions.type",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	cursor, err := ubs.engagementCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"interaction_types": results,
	}, nil
}

// Get User Interest Score for Content
func (ubs *UserBehaviorService) GetUserInterestScore(userID, contentID primitive.ObjectID) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":    userID,
				"content_id": contentID,
			},
		},
		{
			"$group": bson.M{
				"_id":                 nil,
				"total_view_duration": bson.M{"$sum": "$view_duration"},
				"interaction_count":   bson.M{"$sum": bson.M{"$size": "$interactions"}},
				"view_count":          bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := ubs.engagementCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	result := results[0]

	// Calculate interest score based on engagement metrics
	viewDuration := getFloat64(result["total_view_duration"])
	interactionCount := getFloat64(result["interaction_count"])
	viewCount := getFloat64(result["view_count"])

	// Simple scoring algorithm (can be made more sophisticated)
	score := (viewDuration/1000)*0.3 + interactionCount*0.5 + viewCount*0.2

	return score, nil
}

// Helper function to safely convert interface{} to float64
func getFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// Get User Content Preferences
func (ubs *UserBehaviorService) GetUserContentPreferences(userID primitive.ObjectID) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's engagement patterns from last 30 days
	since := time.Now().Add(-30 * 24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"view_time": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id":               "$content_type",
				"total_engagement":  bson.M{"$sum": bson.M{"$size": "$interactions"}},
				"avg_view_duration": bson.M{"$avg": "$view_duration"},
				"view_count":        bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := ubs.engagementCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	preferences := make(map[string]float64)

	for _, result := range results {
		contentType := result["_id"].(string)
		engagement := getFloat64(result["total_engagement"])
		avgDuration := getFloat64(result["avg_view_duration"])
		viewCount := getFloat64(result["view_count"])

		// Calculate preference score
		score := engagement*0.4 + (avgDuration/1000)*0.3 + viewCount*0.3
		preferences[contentType] = score
	}

	return preferences, nil
}

// Get Similar Users based on behavior
func (ubs *UserBehaviorService) GetSimilarUsers(userID primitive.ObjectID, limit int) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get user's interests
	userPrefs, err := ubs.GetUserContentPreferences(userID)
	if err != nil {
		return nil, err
	}

	// Find users with similar preferences (simplified approach)
	// In a real implementation, you'd use more sophisticated similarity algorithms
	since := time.Now().Add(-30 * 24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   bson.M{"$ne": userID},
				"view_time": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id":               "$user_id",
				"content_types":     bson.M{"$addToSet": "$content_type"},
				"interaction_count": bson.M{"$sum": bson.M{"$size": "$interactions"}},
			},
		},
		{
			"$match": bson.M{
				"interaction_count": bson.M{"$gte": 5}, // Minimum activity threshold
			},
		},
		{
			"$limit": limit * 2, // Get more candidates to filter
		},
	}

	cursor, err := ubs.engagementCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var candidates []map[string]interface{}
	if err := cursor.All(ctx, &candidates); err != nil {
		return nil, err
	}

	// Simple similarity scoring based on common content types
	var similarUsers []primitive.ObjectID

	for _, candidate := range candidates {
		candidateID := candidate["_id"].(primitive.ObjectID)
		contentTypes := candidate["content_types"].(primitive.A)

		// Check for common interests
		commonInterests := 0
		for _, ct := range contentTypes {
			if _, exists := userPrefs[ct.(string)]; exists {
				commonInterests++
			}
		}

		// If they have at least 2 common content types, consider them similar
		if commonInterests >= 2 {
			similarUsers = append(similarUsers, candidateID)
			if len(similarUsers) >= limit {
				break
			}
		}
	}

	return similarUsers, nil
}
