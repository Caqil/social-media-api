// internal/services/analytics_service.go
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

type AnalyticsService struct {
	eventsCollection *mongo.Collection
	userCollection   *mongo.Collection
	postCollection   *mongo.Collection
	db               *mongo.Database
}

type AnalyticsEvent struct {
	models.BaseModel `bson:",inline"`
	UserID           *primitive.ObjectID    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	SessionID        string                 `json:"session_id" bson:"session_id"`
	EventType        string                 `json:"event_type" bson:"event_type"`
	EventName        string                 `json:"event_name" bson:"event_name"`
	Properties       map[string]interface{} `json:"properties" bson:"properties"`
	IPAddress        string                 `json:"ip_address" bson:"ip_address"`
	UserAgent        string                 `json:"user_agent" bson:"user_agent"`
	Platform         string                 `json:"platform" bson:"platform"` // web, ios, android
	Source           string                 `json:"source" bson:"source"`     // direct, social, search, etc.
	Referrer         string                 `json:"referrer" bson:"referrer"`
	Timestamp        time.Time              `json:"timestamp" bson:"timestamp"`
}

type UserAnalytics struct {
	UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
	TotalSessions   int64              `json:"total_sessions" bson:"total_sessions"`
	TotalPageViews  int64              `json:"total_page_views" bson:"total_page_views"`
	TotalTimeSpent  int64              `json:"total_time_spent" bson:"total_time_spent"` // in seconds
	PostsCreated    int64              `json:"posts_created" bson:"posts_created"`
	CommentsCreated int64              `json:"comments_created" bson:"comments_created"`
	LikesGiven      int64              `json:"likes_given" bson:"likes_given"`
	SharesGiven     int64              `json:"shares_given" bson:"shares_given"`
	FollowsGiven    int64              `json:"follows_given" bson:"follows_given"`
	ProfileViews    int64              `json:"profile_views" bson:"profile_views"`
	LastActiveAt    time.Time          `json:"last_active_at" bson:"last_active_at"`
	TopDevices      []DeviceInfo       `json:"top_devices" bson:"top_devices"`
	TopLocations    []LocationInfo     `json:"top_locations" bson:"top_locations"`
}

type PostAnalytics struct {
	PostID           primitive.ObjectID `json:"post_id" bson:"post_id"`
	TotalViews       int64              `json:"total_views" bson:"total_views"`
	UniqueViews      int64              `json:"unique_views" bson:"unique_views"`
	TotalLikes       int64              `json:"total_likes" bson:"total_likes"`
	TotalComments    int64              `json:"total_comments" bson:"total_comments"`
	TotalShares      int64              `json:"total_shares" bson:"total_shares"`
	EngagementRate   float64            `json:"engagement_rate" bson:"engagement_rate"`
	ViewDuration     float64            `json:"view_duration" bson:"view_duration"`
	ClickThroughRate float64            `json:"click_through_rate" bson:"click_through_rate"`
	Demographics     DemographicData    `json:"demographics" bson:"demographics"`
	TopReferrers     []ReferrerInfo     `json:"top_referrers" bson:"top_referrers"`
	TimeDistribution []TimeData         `json:"time_distribution" bson:"time_distribution"`
}

type DemographicData struct {
	AgeGroups    map[string]int64 `json:"age_groups" bson:"age_groups"`
	GenderSplit  map[string]int64 `json:"gender_split" bson:"gender_split"`
	TopCountries []CountryData    `json:"top_countries" bson:"top_countries"`
	TopCities    []CityData       `json:"top_cities" bson:"top_cities"`
}

type DeviceInfo struct {
	Platform string `json:"platform" bson:"platform"`
	Count    int64  `json:"count" bson:"count"`
}

type LocationInfo struct {
	Country string `json:"country" bson:"country"`
	City    string `json:"city" bson:"city"`
	Count   int64  `json:"count" bson:"count"`
}

type ReferrerInfo struct {
	Source string `json:"source" bson:"source"`
	Count  int64  `json:"count" bson:"count"`
}

type TimeData struct {
	Hour  int   `json:"hour" bson:"hour"`
	Count int64 `json:"count" bson:"count"`
}

type CountryData struct {
	Country string `json:"country" bson:"country"`
	Count   int64  `json:"count" bson:"count"`
}

type CityData struct {
	City  string `json:"city" bson:"city"`
	Count int64  `json:"count" bson:"count"`
}

type PlatformStats struct {
	TotalUsers     int64                `json:"total_users"`
	ActiveUsers    int64                `json:"active_users"`
	NewUsers       int64                `json:"new_users"`
	TotalPosts     int64                `json:"total_posts"`
	TotalComments  int64                `json:"total_comments"`
	TotalLikes     int64                `json:"total_likes"`
	UserGrowthRate float64              `json:"user_growth_rate"`
	EngagementRate float64              `json:"engagement_rate"`
	TopContent     []ContentPerformance `json:"top_content"`
	UserActivity   map[string]int64     `json:"user_activity"`
	PlatformSplit  map[string]int64     `json:"platform_split"`
	GrowthTrends   []GrowthData         `json:"growth_trends"`
}

type ContentPerformance struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"` // post, story, etc.
	Title          string  `json:"title"`
	Views          int64   `json:"views"`
	Engagement     int64   `json:"engagement"`
	EngagementRate float64 `json:"engagement_rate"`
}

type GrowthData struct {
	Date       time.Time `json:"date"`
	Users      int64     `json:"users"`
	Posts      int64     `json:"posts"`
	Engagement int64     `json:"engagement"`
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		eventsCollection: config.DB.Collection("analytics_events"),
		userCollection:   config.DB.Collection("users"),
		postCollection:   config.DB.Collection("posts"),
		db:               config.DB,
	}
}

// TrackEvent records an analytics event
func (as *AnalyticsService) TrackEvent(event AnalyticsEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event.Timestamp = time.Now()
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := as.eventsCollection.InsertOne(ctx, event)
	return err
}

// TrackPageView tracks a page view event
func (as *AnalyticsService) TrackPageView(userID *primitive.ObjectID, sessionID, page, referrer, userAgent, platform string) error {
	event := AnalyticsEvent{
		UserID:    userID,
		SessionID: sessionID,
		EventType: "page_view",
		EventName: "page_viewed",
		Properties: map[string]interface{}{
			"page":     page,
			"referrer": referrer,
		},
		UserAgent: userAgent,
		Platform:  platform,
		Referrer:  referrer,
	}

	return as.TrackEvent(event)
}

// TrackUserAction tracks user interactions
func (as *AnalyticsService) TrackUserAction(userID primitive.ObjectID, action, targetType string, targetID *primitive.ObjectID, properties map[string]interface{}) error {
	if properties == nil {
		properties = make(map[string]interface{})
	}

	properties["target_type"] = targetType
	if targetID != nil {
		properties["target_id"] = targetID.Hex()
	}

	event := AnalyticsEvent{
		UserID:     &userID,
		EventType:  "user_action",
		EventName:  action,
		Properties: properties,
	}

	return as.TrackEvent(event)
}

// TrackPostView tracks when a post is viewed
func (as *AnalyticsService) TrackPostView(userID *primitive.ObjectID, postID primitive.ObjectID, sessionID, source string, duration float64) error {
	properties := map[string]interface{}{
		"post_id":  postID.Hex(),
		"source":   source,
		"duration": duration,
	}

	event := AnalyticsEvent{
		UserID:     userID,
		SessionID:  sessionID,
		EventType:  "content_interaction",
		EventName:  "post_viewed",
		Properties: properties,
	}

	return as.TrackEvent(event)
}

// GetUserAnalytics retrieves analytics for a specific user
func (as *AnalyticsService) GetUserAnalytics(userID primitive.ObjectID, timeRange string) (*UserAnalytics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Calculate time filter
	timeFilter := as.getTimeFilter(timeRange)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"timestamp": bson.M{"$gte": timeFilter},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"total_sessions": bson.M{
					"$addToSet": "$session_id",
				},
				"total_page_views": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_type", "page_view"}},
							1, 0,
						},
					},
				},
				"posts_created": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_created"}},
							1, 0,
						},
					},
				},
				"likes_given": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_liked"}},
							1, 0,
						},
					},
				},
				"platforms": bson.M{
					"$push": "$platform",
				},
				"last_active": bson.M{
					"$max": "$timestamp",
				},
			},
		},
		{
			"$project": bson.M{
				"total_sessions":   bson.M{"$size": "$total_sessions"},
				"total_page_views": 1,
				"posts_created":    1,
				"likes_given":      1,
				"last_active":      1,
				"platforms":        1,
			},
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []UserAnalytics
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &UserAnalytics{
			UserID: userID,
		}, nil
	}

	return &results[0], nil
}

// GetPostAnalytics retrieves analytics for a specific post
func (as *AnalyticsService) GetPostAnalytics(postID primitive.ObjectID, timeRange string) (*PostAnalytics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	timeFilter := as.getTimeFilter(timeRange)

	// Get post engagement data
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"properties.post_id": postID.Hex(),
				"timestamp":          bson.M{"$gte": timeFilter},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"total_views": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_viewed"}},
							1, 0,
						},
					},
				},
				"unique_viewers": bson.M{
					"$addToSet": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_viewed"}},
							"$user_id", nil,
						},
					},
				},
				"total_likes": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_liked"}},
							1, 0,
						},
					},
				},
				"total_shares": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_shared"}},
							1, 0,
						},
					},
				},
				"avg_duration": bson.M{
					"$avg": bson.M{
						"$toDouble": "$properties.duration",
					},
				},
				"referrers": bson.M{
					"$push": "$properties.source",
				},
			},
		},
		{
			"$project": bson.M{
				"total_views":   1,
				"unique_views":  bson.M{"$size": "$unique_viewers"},
				"total_likes":   1,
				"total_shares":  1,
				"view_duration": "$avg_duration",
				"referrers":     1,
			},
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []PostAnalytics
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &PostAnalytics{
			PostID: postID,
		}, nil
	}

	result := results[0]
	result.PostID = postID

	// Calculate engagement rate
	if result.TotalViews > 0 {
		totalEngagement := result.TotalLikes + result.TotalComments + result.TotalShares
		result.EngagementRate = (float64(totalEngagement) / float64(result.TotalViews)) * 100
	}

	return &result, nil
}

// GetPlatformStats retrieves overall platform statistics
func (as *AnalyticsService) GetPlatformStats(timeRange string) (*PlatformStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timeFilter := as.getTimeFilter(timeRange)

	// Aggregate various metrics
	stats := &PlatformStats{}

	// Get user stats
	userStats, err := as.getUserStats(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = userStats["total"].(int64)
	stats.ActiveUsers = userStats["active"].(int64)
	stats.NewUsers = userStats["new"].(int64)

	// Get content stats
	contentStats, err := as.getContentStats(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.TotalPosts = contentStats["posts"].(int64)
	stats.TotalComments = contentStats["comments"].(int64)
	stats.TotalLikes = contentStats["likes"].(int64)

	// Calculate engagement rate
	if stats.ActiveUsers > 0 {
		totalEngagement := stats.TotalLikes + stats.TotalComments
		stats.EngagementRate = (float64(totalEngagement) / float64(stats.ActiveUsers)) * 100
	}

	// Get platform distribution
	platformDist, err := as.getPlatformDistribution(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.PlatformSplit = platformDist

	// Get growth trends
	growthTrends, err := as.getGrowthTrends(ctx, timeFilter)
	if err != nil {
		return nil, err
	}
	stats.GrowthTrends = growthTrends

	return stats, nil
}

// GetTopContent retrieves top performing content
func (as *AnalyticsService) GetTopContent(contentType, timeRange string, limit int) ([]ContentPerformance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	timeFilter := as.getTimeFilter(timeRange)

	var collection *mongo.Collection
	var matchField string

	switch contentType {
	case "posts":
		collection = as.postCollection
		matchField = "properties.post_id"
	default:
		collection = as.postCollection
		matchField = "properties.post_id"
	}

	// Get engagement data from analytics events
	pipeline := []bson.M{
		{
			"$match": bson.M{
				matchField:  bson.M{"$exists": true},
				"timestamp": bson.M{"$gte": timeFilter},
			},
		},
		{
			"$group": bson.M{
				"_id": "$" + matchField,
				"views": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", contentType[:len(contentType)-1] + "_viewed"}},
							1, 0,
						},
					},
				},
				"engagement": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$in": []interface{}{"$event_name", []string{contentType[:len(contentType)-1] + "_liked", contentType[:len(contentType)-1] + "_shared", contentType[:len(contentType)-1] + "_commented"}}},
							1, 0,
						},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"engagement_rate": bson.M{
					"$cond": []interface{}{
						bson.M{"$gt": []interface{}{"$views", 0}},
						bson.M{"$multiply": []interface{}{
							bson.M{"$divide": []interface{}{"$engagement", "$views"}},
							100,
						}},
						0,
					},
				},
			},
		},
		{
			"$sort": bson.M{"engagement_rate": -1, "views": -1},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID             string  `bson:"_id"`
		Views          int64   `bson:"views"`
		Engagement     int64   `bson:"engagement"`
		EngagementRate float64 `bson:"engagement_rate"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to ContentPerformance and get additional details
	var content []ContentPerformance
	for _, result := range results {
		performance := ContentPerformance{
			ID:             result.ID,
			Type:           contentType,
			Views:          result.Views,
			Engagement:     result.Engagement,
			EngagementRate: result.EngagementRate,
		}

		// Get content title/details from the respective collection
		if objID, err := primitive.ObjectIDFromHex(result.ID); err == nil {
			if title := as.getContentTitle(ctx, collection, objID); title != "" {
				performance.Title = title
			}
		}

		content = append(content, performance)
	}

	return content, nil
}

// GetUserEngagement calculates user engagement metrics
func (as *AnalyticsService) GetUserEngagement(userID primitive.ObjectID, timeRange string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	timeFilter := as.getTimeFilter(timeRange)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"timestamp": bson.M{"$gte": timeFilter},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date": bson.M{
						"$dateToString": bson.M{
							"format": "%Y-%m-%d",
							"date":   "$timestamp",
						},
					},
				},
				"sessions": bson.M{"$addToSet": "$session_id"},
				"page_views": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_type", "page_view"}},
							1, 0,
						},
					},
				},
				"actions": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_type", "user_action"}},
							1, 0,
						},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"date":       "$_id.date",
				"sessions":   bson.M{"$size": "$sessions"},
				"page_views": 1,
				"actions":    1,
			},
		},
		{
			"$sort": bson.M{"date": 1},
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dailyEngagement []map[string]interface{}
	if err := cursor.All(ctx, &dailyEngagement); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"daily_engagement": dailyEngagement,
		"user_id":          userID.Hex(),
		"time_range":       timeRange,
	}

	return result, nil
}

// Helper methods

func (as *AnalyticsService) getTimeFilter(timeRange string) time.Time {
	now := time.Now()
	switch timeRange {
	case "hour":
		return now.Add(-1 * time.Hour)
	case "day":
		return now.Add(-24 * time.Hour)
	case "week":
		return now.Add(-7 * 24 * time.Hour)
	case "month":
		return now.Add(-30 * 24 * time.Hour)
	case "year":
		return now.Add(-365 * 24 * time.Hour)
	default:
		return now.Add(-24 * time.Hour)
	}
}

func (as *AnalyticsService) getUserStats(ctx context.Context, timeFilter time.Time) (map[string]interface{}, error) {
	// Get total users
	totalUsers, err := as.userCollection.CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		return nil, err
	}

	// Get active users (users with events in time range)
	activeUsersPipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{"$gte": timeFilter},
				"user_id":   bson.M{"$exists": true},
			},
		},
		{
			"$group": bson.M{
				"_id": "$user_id",
			},
		},
		{
			"$count": "active_users",
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, activeUsersPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activeResult []struct {
		ActiveUsers int64 `bson:"active_users"`
	}
	cursor.All(ctx, &activeResult)

	activeUsers := int64(0)
	if len(activeResult) > 0 {
		activeUsers = activeResult[0].ActiveUsers
	}

	// Get new users
	newUsers, err := as.userCollection.CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": timeFilter},
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":  totalUsers,
		"active": activeUsers,
		"new":    newUsers,
	}, nil
}

func (as *AnalyticsService) getContentStats(ctx context.Context, timeFilter time.Time) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{"$gte": timeFilter},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"posts": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "post_created"}},
							1, 0,
						},
					},
				},
				"comments": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$event_name", "comment_created"}},
							1, 0,
						},
					},
				},
				"likes": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$regex": []interface{}{"$event_name", ".*_liked"}},
							1, 0,
						},
					},
				},
			},
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, pipeline)
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
			"posts":    int64(0),
			"comments": int64(0),
			"likes":    int64(0),
		}, nil
	}

	return results[0], nil
}

func (as *AnalyticsService) getPlatformDistribution(ctx context.Context, timeFilter time.Time) (map[string]int64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{"$gte": timeFilter},
				"platform":  bson.M{"$exists": true, "$ne": ""},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$platform",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := as.eventsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Platform string `bson:"_id"`
		Count    int64  `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Platform] = result.Count
	}

	return distribution, nil
}

func (as *AnalyticsService) getGrowthTrends(ctx context.Context, timeFilter time.Time) ([]GrowthData, error) {
	// Implement growth trends calculation
	// This would typically involve daily/weekly aggregations
	return []GrowthData{}, nil
}

func (as *AnalyticsService) getContentTitle(ctx context.Context, collection *mongo.Collection, objID primitive.ObjectID) string {
	var result struct {
		Title   string `bson:"title"`
		Content string `bson:"content"`
	}

	err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return ""
	}

	if result.Title != "" {
		return result.Title
	}

	// If no title, use truncated content
	if len(result.Content) > 50 {
		return result.Content[:47] + "..."
	}

	return result.Content
}

// CreateIndexes creates necessary indexes for analytics collection
func (as *AnalyticsService) CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "event_type", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "properties.post_id", Value: 1}, {Key: "timestamp", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "session_id", Value: 1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "timestamp", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
	}

	_, err := as.eventsCollection.Indexes().CreateMany(ctx, indexes)
	return err
}
