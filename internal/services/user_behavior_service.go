// internal/services/user_behavior_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserBehaviorService struct {
	db              *mongo.Database
	sessions        *mongo.Collection
	engagements     *mongo.Collection
	journeys        *mongo.Collection
	recommendations *mongo.Collection
	experiments     *mongo.Collection
	analytics       *mongo.Collection
}

func NewUserBehaviorService(db *mongo.Database) *UserBehaviorService {
	service := &UserBehaviorService{
		db:              db,
		sessions:        db.Collection("user_sessions"),
		engagements:     db.Collection("content_engagements"),
		journeys:        db.Collection("user_journeys"),
		recommendations: db.Collection("recommendation_events"),
		experiments:     db.Collection("experiment_events"),
		analytics:       db.Collection("user_analytics"),
	}

	// Create indexes
	service.createIndexes()
	return service
}

func (s *UserBehaviorService) createIndexes() {
	ctx := context.Background()

	// Session indexes
	s.sessions.CreateIndex(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "session_id", Value: 1}},
	})
	s.sessions.CreateIndex(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "start_time", Value: -1}},
	})

	// Engagement indexes
	s.engagements.CreateIndex(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "view_time", Value: -1}},
	})
	s.engagements.CreateIndex(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "content_id", Value: 1}, {Key: "content_type", Value: 1}},
	})

	// Journey indexes
	s.journeys.CreateIndex(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "session_id", Value: 1}},
	})

	// Recommendation indexes
	s.recommendations.CreateIndex(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "presented", Value: -1}},
	})
}

// SESSION TRACKING

func (s *UserBehaviorService) StartSession(userID primitive.ObjectID, sessionID, deviceInfo, ipAddress, userAgent string) error {
	session := models.UserSession{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		SessionID:    sessionID,
		StartTime:    time.Now(),
		DeviceInfo:   deviceInfo,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		PagesVisited: []models.PageVisit{},
		Actions:      []models.UserAction{},
	}

	_, err := s.sessions.InsertOne(context.Background(), session)
	return err
}

func (s *UserBehaviorService) EndSession(sessionID string) error {
	filter := bson.M{"session_id": sessionID}
	update := bson.M{
		"$set": bson.M{
			"end_time":   time.Now(),
			"updated_at": time.Now(),
		},
	}

	// Calculate duration
	var session models.UserSession
	err := s.sessions.FindOne(context.Background(), filter).Decode(&session)
	if err == nil {
		duration := time.Since(session.StartTime).Seconds()
		update["$set"].(bson.M)["duration"] = int64(duration)
	}

	_, err = s.sessions.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *UserBehaviorService) RecordPageVisit(userID primitive.ObjectID, sessionID string, pageVisit models.PageVisit) error {
	filter := bson.M{"user_id": userID, "session_id": sessionID}
	update := bson.M{
		"$push": bson.M{"pages_visited": pageVisit},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := s.sessions.UpdateOne(context.Background(), filter, update)

	// Also track in journey
	go s.trackJourneyTouchpoint(userID, sessionID, "page_visit", pageVisit.URL, map[string]interface{}{
		"duration":  pageVisit.Duration,
		"timestamp": pageVisit.Timestamp,
	})

	return err
}

func (s *UserBehaviorService) RecordUserAction(userID primitive.ObjectID, sessionID string, action models.UserAction) error {
	filter := bson.M{"user_id": userID, "session_id": sessionID}
	update := bson.M{
		"$push": bson.M{"actions": action},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := s.sessions.UpdateOne(context.Background(), filter, update)

	// Track in journey
	go s.trackJourneyTouchpoint(userID, sessionID, action.Type, action.Target, action.Metadata)

	return err
}

// CONTENT ENGAGEMENT TRACKING

func (s *UserBehaviorService) RecordContentEngagement(engagement models.ContentEngagement) error {
	engagement.ID = primitive.NewObjectID()
	engagement.ViewTime = time.Now()

	_, err := s.engagements.InsertOne(context.Background(), engagement)

	// Update real-time analytics
	go s.updateContentAnalytics(engagement)

	return err
}

func (s *UserBehaviorService) RecordInteraction(userID, contentID primitive.ObjectID, contentType, interactionType, source string, metadata map[string]interface{}) error {
	// Find existing engagement or create new one
	filter := bson.M{
		"user_id":    userID,
		"content_id": contentID,
		"view_time":  bson.M{"$gte": time.Now().Add(-time.Hour)}, // within last hour
	}

	interaction := models.Interaction{
		Type:      interactionType,
		Timestamp: time.Now(),
		Value:     "",
	}

	if val, ok := metadata["value"].(string); ok {
		interaction.Value = val
	}

	update := bson.M{
		"$push": bson.M{"interactions": interaction},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result := s.engagements.UpdateOne(context.Background(), filter, update)

	// If no existing engagement found, create new one
	if result.MatchedCount == 0 {
		engagement := models.ContentEngagement{
			ID:           primitive.NewObjectID(),
			UserID:       userID,
			ContentID:    contentID,
			ContentType:  contentType,
			ViewTime:     time.Now(),
			Source:       source,
			Interactions: []models.Interaction{interaction},
			Context:      metadata,
		}
		_, err := s.engagements.InsertOne(context.Background(), engagement)
		return err
	}

	return result.Err
}

// USER JOURNEY TRACKING

func (s *UserBehaviorService) trackJourneyTouchpoint(userID primitive.ObjectID, sessionID, action, target string, metadata map[string]interface{}) error {
	touchpoint := models.Touchpoint{
		Page:      target,
		Action:    action,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	filter := bson.M{"user_id": userID, "session_id": sessionID, "completed": false}
	update := bson.M{
		"$push": bson.M{"touchpoints": touchpoint},
		"$set":  bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"user_id":    userID,
			"session_id": sessionID,
			"completed":  false,
			"created_at": time.Now(),
		},
	}

	_, err := s.journeys.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	return err
}

func (s *UserBehaviorService) CompleteJourney(userID, sessionID, goal string) error {
	filter := bson.M{"user_id": userID, "session_id": sessionID}

	var journey models.UserJourney
	err := s.journeys.FindOne(context.Background(), filter).Decode(&journey)
	if err != nil {
		return err
	}

	duration := time.Since(journey.Touchpoints[0].Timestamp).Seconds()

	update := bson.M{
		"$set": bson.M{
			"goal":         goal,
			"completed":    true,
			"duration":     int64(duration),
			"completed_at": time.Now(),
		},
	}

	_, err = s.journeys.UpdateOne(context.Background(), filter, update)
	return err
}

// RECOMMENDATION TRACKING

func (s *UserBehaviorService) TrackRecommendation(userID primitive.ObjectID, recommendation models.RecommendationEvent, action string) error {
	switch action {
	case "presented":
		recommendation.ID = primitive.NewObjectID()
		recommendation.Presented = time.Now()
		_, err := s.recommendations.InsertOne(context.Background(), recommendation)
		return err

	case "clicked":
		filter := bson.M{
			"user_id":   userID,
			"item_id":   recommendation.ItemID,
			"algorithm": recommendation.Algorithm,
			"clicked":   nil,
		}
		update := bson.M{"$set": bson.M{"clicked": time.Now()}}
		_, err := s.recommendations.UpdateOne(context.Background(), filter, update)
		return err

	case "converted":
		filter := bson.M{
			"user_id":   userID,
			"item_id":   recommendation.ItemID,
			"algorithm": recommendation.Algorithm,
		}
		update := bson.M{"$set": bson.M{"converted": time.Now()}}
		_, err := s.recommendations.UpdateOne(context.Background(), filter, update)
		return err
	}

	return fmt.Errorf("unknown action: %s", action)
}

// A/B TESTING

func (s *UserBehaviorService) TrackExperiment(userID primitive.ObjectID, experimentID, variantID, event string, value float64) error {
	experiment := models.ExperimentEvent{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		ExperimentID: experimentID,
		VariantID:    variantID,
		Event:        event,
		Value:        value,
		Timestamp:    time.Now(),
	}

	_, err := s.experiments.InsertOne(context.Background(), experiment)
	return err
}

// ANALYTICS AND INSIGHTS

func (s *UserBehaviorService) GetUserBehaviorAnalytics(userID primitive.ObjectID, timeRange string) (*models.UserBehaviorAnalytics, error) {
	var startTime time.Time
	switch timeRange {
	case "day":
		startTime = time.Now().AddDate(0, 0, -1)
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -7)
	}

	analytics := &models.UserBehaviorAnalytics{
		UserID:    userID,
		TimeRange: timeRange,
		StartTime: startTime,
		EndTime:   time.Now(),
	}

	// Session analytics
	sessionStats, err := s.getSessionAnalytics(userID, startTime)
	if err == nil {
		analytics.SessionStats = *sessionStats
	}

	// Engagement analytics
	engagementStats, err := s.getEngagementAnalytics(userID, startTime)
	if err == nil {
		analytics.EngagementStats = *engagementStats
	}

	// Content preferences
	contentPrefs, err := s.getContentPreferences(userID, startTime)
	if err == nil {
		analytics.ContentPreferences = contentPrefs
	}

	// Activity patterns
	activityPatterns, err := s.getActivityPatterns(userID, startTime)
	if err == nil {
		analytics.ActivityPatterns = activityPatterns
	}

	return analytics, nil
}

func (s *UserBehaviorService) getSessionAnalytics(userID primitive.ObjectID, startTime time.Time) (*models.SessionStats, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"start_time": bson.M{"$gte": startTime},
		}},
		{"$group": bson.M{
			"_id":            nil,
			"total_sessions": bson.M{"$sum": 1},
			"total_duration": bson.M{"$sum": "$duration"},
			"avg_duration":   bson.M{"$avg": "$duration"},
			"total_pages":    bson.M{"$sum": bson.M{"$size": "$pages_visited"}},
			"total_actions":  bson.M{"$sum": bson.M{"$size": "$actions"}},
		}},
	}

	cursor, err := s.sessions.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var result struct {
		TotalSessions int64   `bson:"total_sessions"`
		TotalDuration int64   `bson:"total_duration"`
		AvgDuration   float64 `bson:"avg_duration"`
		TotalPages    int64   `bson:"total_pages"`
		TotalActions  int64   `bson:"total_actions"`
	}

	if cursor.Next(context.Background()) {
		cursor.Decode(&result)
	}

	return &models.SessionStats{
		TotalSessions: result.TotalSessions,
		TotalDuration: result.TotalDuration,
		AvgDuration:   result.AvgDuration,
		TotalPages:    result.TotalPages,
		TotalActions:  result.TotalActions,
	}, nil
}

func (s *UserBehaviorService) getEngagementAnalytics(userID primitive.ObjectID, startTime time.Time) (*models.EngagementStats, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":   userID,
			"view_time": bson.M{"$gte": startTime},
		}},
		{"$group": bson.M{
			"_id":                "$content_type",
			"total_views":        bson.M{"$sum": 1},
			"total_duration":     bson.M{"$sum": "$view_duration"},
			"avg_duration":       bson.M{"$avg": "$view_duration"},
			"avg_scroll_depth":   bson.M{"$avg": "$scroll_depth"},
			"total_interactions": bson.M{"$sum": bson.M{"$size": "$interactions"}},
		}},
	}

	cursor, err := s.engagements.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	stats := &models.EngagementStats{
		ByContentType: make(map[string]models.ContentTypeStats),
	}

	for cursor.Next(context.Background()) {
		var result struct {
			ContentType       string  `bson:"_id"`
			TotalViews        int64   `bson:"total_views"`
			TotalDuration     int64   `bson:"total_duration"`
			AvgDuration       float64 `bson:"avg_duration"`
			AvgScrollDepth    float64 `bson:"avg_scroll_depth"`
			TotalInteractions int64   `bson:"total_interactions"`
		}
		cursor.Decode(&result)

		stats.ByContentType[result.ContentType] = models.ContentTypeStats{
			TotalViews:        result.TotalViews,
			TotalDuration:     result.TotalDuration,
			AvgDuration:       result.AvgDuration,
			AvgScrollDepth:    result.AvgScrollDepth,
			TotalInteractions: result.TotalInteractions,
		}
	}

	return stats, nil
}

func (s *UserBehaviorService) getContentPreferences(userID primitive.ObjectID, startTime time.Time) (map[string]float64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":   userID,
			"view_time": bson.M{"$gte": startTime},
		}},
		{"$unwind": "$interactions"},
		{"$group": bson.M{
			"_id":   "$interactions.type",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
	}

	cursor, err := s.engagements.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	preferences := make(map[string]float64)
	var total float64 = 0

	// First pass: get totals
	var results []struct {
		Type  string `bson:"_id"`
		Count int64  `bson:"count"`
	}

	for cursor.Next(context.Background()) {
		var result struct {
			Type  string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		cursor.Decode(&result)
		results = append(results, result)
		total += float64(result.Count)
	}

	// Second pass: calculate percentages
	for _, result := range results {
		if total > 0 {
			preferences[result.Type] = float64(result.Count) / total * 100
		}
	}

	return preferences, nil
}

func (s *UserBehaviorService) getActivityPatterns(userID primitive.ObjectID, startTime time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"start_time": bson.M{"$gte": startTime},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"hour":      bson.M{"$hour": "$start_time"},
				"dayOfWeek": bson.M{"$dayOfWeek": "$start_time"},
			},
			"sessions": bson.M{"$sum": 1},
			"duration": bson.M{"$sum": "$duration"},
		}},
		{"$sort": bson.M{"sessions": -1}},
	}

	cursor, err := s.sessions.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	patterns := make(map[string]interface{})
	hourlyActivity := make(map[int]int64)
	dailyActivity := make(map[int]int64)

	for cursor.Next(context.Background()) {
		var result struct {
			ID struct {
				Hour      int `bson:"hour"`
				DayOfWeek int `bson:"dayOfWeek"`
			} `bson:"_id"`
			Sessions int64 `bson:"sessions"`
			Duration int64 `bson:"duration"`
		}
		cursor.Decode(&result)

		hourlyActivity[result.ID.Hour] += result.Sessions
		dailyActivity[result.ID.DayOfWeek] += result.Sessions
	}

	patterns["hourly_activity"] = hourlyActivity
	patterns["daily_activity"] = dailyActivity

	return patterns, nil
}

// AUTOMATIC TRACKING METHODS

func (s *UserBehaviorService) AutoTrackPostView(userID, postID primitive.ObjectID, source string, duration int64) {
	engagement := models.ContentEngagement{
		UserID:       userID,
		ContentID:    postID,
		ContentType:  "post",
		ViewDuration: duration,
		Source:       source,
		Context:      map[string]interface{}{"auto_tracked": true},
	}

	go s.RecordContentEngagement(engagement)
}

func (s *UserBehaviorService) AutoTrackPostInteraction(userID, postID primitive.ObjectID, interactionType, source string) {
	metadata := map[string]interface{}{
		"auto_tracked": true,
		"timestamp":    time.Now(),
	}

	go s.RecordInteraction(userID, postID, "post", interactionType, source, metadata)
}

func (s *UserBehaviorService) AutoTrackStoryView(userID, storyID primitive.ObjectID, source string, duration int64) {
	engagement := models.ContentEngagement{
		UserID:       userID,
		ContentID:    storyID,
		ContentType:  "story",
		ViewDuration: duration,
		Source:       source,
		Context:      map[string]interface{}{"auto_tracked": true},
	}

	go s.RecordContentEngagement(engagement)
}

func (s *UserBehaviorService) AutoTrackSearch(userID primitive.ObjectID, query, resultType string, resultsCount int) {
	metadata := map[string]interface{}{
		"auto_tracked":  true,
		"query":         query,
		"result_type":   resultType,
		"results_count": resultsCount,
		"timestamp":     time.Now(),
	}

	sessionID := fmt.Sprintf("search_%d", time.Now().Unix())

	action := models.UserAction{
		Type:      "search",
		Target:    query,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	go s.RecordUserAction(userID, sessionID, action)
}

// REAL-TIME ANALYTICS UPDATE

func (s *UserBehaviorService) updateContentAnalytics(engagement models.ContentEngagement) {
	// Update real-time content analytics
	filter := bson.M{
		"content_id":   engagement.ContentID,
		"content_type": engagement.ContentType,
		"date":         time.Now().Format("2006-01-02"),
	}

	update := bson.M{
		"$inc": bson.M{
			"total_views":    1,
			"total_duration": engagement.ViewDuration,
		},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"created_at": time.Now(),
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	s.analytics.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
}

// UTILITY METHODS

func (s *UserBehaviorService) GetRecommendationPerformance(algorithm string, timeRange string) (*models.RecommendationPerformance, error) {
	var startTime time.Time
	switch timeRange {
	case "day":
		startTime = time.Now().AddDate(0, 0, -1)
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -7)
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"algorithm": algorithm,
			"presented": bson.M{"$gte": startTime},
		}},
		{"$group": bson.M{
			"_id":             nil,
			"total_presented": bson.M{"$sum": 1},
			"total_clicked": bson.M{"$sum": bson.M{
				"$cond": bson.A{bson.M{"$ne": bson.A{"$clicked", nil}}, 1, 0},
			}},
			"total_converted": bson.M{"$sum": bson.M{
				"$cond": bson.A{bson.M{"$ne": bson.A{"$converted", nil}}, 1, 0},
			}},
		}},
	}

	cursor, err := s.recommendations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var result struct {
		TotalPresented int64 `bson:"total_presented"`
		TotalClicked   int64 `bson:"total_clicked"`
		TotalConverted int64 `bson:"total_converted"`
	}

	if cursor.Next(context.Background()) {
		cursor.Decode(&result)
	}

	performance := &models.RecommendationPerformance{
		Algorithm:      algorithm,
		TimeRange:      timeRange,
		TotalPresented: result.TotalPresented,
		TotalClicked:   result.TotalClicked,
		TotalConverted: result.TotalConverted,
	}

	if result.TotalPresented > 0 {
		performance.CTR = float64(result.TotalClicked) / float64(result.TotalPresented) * 100
		performance.ConversionRate = float64(result.TotalConverted) / float64(result.TotalPresented) * 100
	}

	return performance, nil
}

func (s *UserBehaviorService) GetContentPopularity(contentType string, timeRange string, limit int) ([]models.ContentPopularity, error) {
	var startTime time.Time
	switch timeRange {
	case "day":
		startTime = time.Now().AddDate(0, 0, -1)
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -7)
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"content_type": contentType,
			"view_time":    bson.M{"$gte": startTime},
		}},
		{"$group": bson.M{
			"_id":                "$content_id",
			"total_views":        bson.M{"$sum": 1},
			"unique_viewers":     bson.M{"$addToSet": "$user_id"},
			"total_duration":     bson.M{"$sum": "$view_duration"},
			"avg_duration":       bson.M{"$avg": "$view_duration"},
			"total_interactions": bson.M{"$sum": bson.M{"$size": "$interactions"}},
		}},
		{"$addFields": bson.M{
			"unique_viewers_count": bson.M{"$size": "$unique_viewers"},
		}},
		{"$sort": bson.M{"total_views": -1}},
		{"$limit": limit},
	}

	cursor, err := s.engagements.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []models.ContentPopularity
	for cursor.Next(context.Background()) {
		var result models.ContentPopularity
		cursor.Decode(&result)
		results = append(results, result)
	}

	return results, nil
}
