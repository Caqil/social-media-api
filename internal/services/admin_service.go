// internal/services/admin_service.go
package services

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminService struct {
	userCollection          *mongo.Collection
	postCollection          *mongo.Collection
	commentCollection       *mongo.Collection
	storyCollection         *mongo.Collection
	groupCollection         *mongo.Collection
	reportCollection        *mongo.Collection
	sessionCollection       *mongo.Collection
	activityCollection      *mongo.Collection
	configCollection        *mongo.Collection
	adminActivityCollection *mongo.Collection
	notificationCollection  *mongo.Collection
	moderationCollection    *mongo.Collection
	db                      *mongo.Database
}

func NewAdminService() *AdminService {
	return &AdminService{
		userCollection:          config.DB.Collection("users"),
		postCollection:          config.DB.Collection("posts"),
		commentCollection:       config.DB.Collection("comments"),
		storyCollection:         config.DB.Collection("stories"),
		groupCollection:         config.DB.Collection("groups"),
		reportCollection:        config.DB.Collection("reports"),
		sessionCollection:       config.DB.Collection("user_sessions"),
		activityCollection:      config.DB.Collection("user_activities"),
		configCollection:        config.DB.Collection("system_config"),
		adminActivityCollection: config.DB.Collection("admin_activities"),
		notificationCollection:  config.DB.Collection("notifications"),
		moderationCollection:    config.DB.Collection("moderation_actions"),
		db:                      config.DB,
	}
}

// Dashboard and Analytics

func (as *AdminService) GetDashboardStats() (*models.AdminDashboardStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats := &models.AdminDashboardStats{}

	// Get basic counts in parallel
	errChan := make(chan error, 10)

	go func() {
		count, err := as.userCollection.CountDocuments(ctx, bson.M{})
		if err == nil {
			stats.TotalUsers = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.userCollection.CountDocuments(ctx, bson.M{
			"last_login_at": bson.M{"$gte": time.Now().Add(-24 * time.Hour)},
		})
		if err == nil {
			stats.ActiveUsers = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.userCollection.CountDocuments(ctx, bson.M{
			"created_at": bson.M{"$gte": time.Now().Add(-24 * time.Hour)},
		})
		if err == nil {
			stats.NewUsers = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.postCollection.CountDocuments(ctx, bson.M{})
		if err == nil {
			stats.TotalPosts = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.commentCollection.CountDocuments(ctx, bson.M{})
		if err == nil {
			stats.TotalComments = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.storyCollection.CountDocuments(ctx, bson.M{})
		if err == nil {
			stats.TotalStories = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.groupCollection.CountDocuments(ctx, bson.M{})
		if err == nil {
			stats.TotalGroups = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.reportCollection.CountDocuments(ctx, bson.M{})
		if err == nil {
			stats.TotalReports = count
		}
		errChan <- err
	}()

	go func() {
		count, err := as.reportCollection.CountDocuments(ctx, bson.M{"status": "pending"})
		if err == nil {
			stats.PendingReports = count
		}
		errChan <- err
	}()

	go func() {
		systemHealth, err := as.getSystemHealth(ctx)
		if err == nil {
			stats.SystemHealth = *systemHealth
		}
		errChan <- err
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		if err := <-errChan; err != nil {
			// Log error but continue
			fmt.Printf("Error getting dashboard stat: %v\n", err)
		}
	}

	// Get additional stats
	var err error
	stats.TopHashtags, _ = as.getTopHashtags(ctx, 10)
	stats.UserGrowth, _ = as.getUserGrowthData(ctx, 30)
	stats.ContentGrowth, _ = as.getContentGrowthData(ctx, 30)
	stats.EngagementMetrics, _ = as.getEngagementMetrics(ctx)
	stats.GeographicData, _ = as.getGeographicData(ctx)
	stats.DeviceStats, _ = as.getDeviceStats(ctx)
	stats.PlatformMetrics, _ = as.getPlatformMetrics(ctx)
	stats.ModerationQueue, _ = as.getModerationQueueStats(ctx)
	stats.RevenueMetrics, _ = as.getRevenueMetrics(ctx)

	return stats, err
}

func (as *AdminService) getSystemHealth(ctx context.Context) (*models.SystemHealthStatus, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get basic system metrics
	health := &models.SystemHealthStatus{
		Status:            "healthy",
		CPUUsage:          0.0,                            // Would need external library for actual CPU usage
		MemoryUsage:       float64(m.Alloc) / 1024 / 1024, // MB
		DiskUsage:         0.0,                            // Would need external library for disk usage
		ActiveConnections: 0,                              // Would get from connection pool
		ResponseTime:      0.0,                            // Would calculate from metrics
		ErrorRate:         0.0,                            // Would calculate from error logs
		DatabaseHealth:    "healthy",
		CacheHealth:       "healthy",
		LastHealthCheck:   time.Now(),
	}

	// Test database connection
	err := as.db.Client().Ping(ctx, nil)
	if err != nil {
		health.DatabaseHealth = "unhealthy"
		health.Status = "warning"
	}

	return health, nil
}

func (as *AdminService) getTopHashtags(ctx context.Context, limit int) ([]models.HashtagStats, error) {
	pipeline := []bson.M{
		{"$unwind": "$hashtags"},
		{"$group": bson.M{
			"_id":        "$hashtags",
			"post_count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"post_count": -1}},
		{"$limit": limit},
	}

	cursor, err := as.postCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.HashtagStats
	for cursor.Next(ctx) {
		var stat models.HashtagStats
		cursor.Decode(&stat)
		stat.Hashtag = cursor.Current.Lookup("_id").StringValue()
		stat.PostCount = cursor.Current.Lookup("post_count").AsInt64()
		results = append(results, stat)
	}

	return results, nil
}

func (as *AdminService) getUserGrowthData(ctx context.Context, days int) ([]models.UserGrowthData, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"created_at": bson.M{"$gte": time.Now().AddDate(0, 0, -days)},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$created_at"},
				"month": bson.M{"$month": "$created_at"},
				"day":   bson.M{"$dayOfMonth": "$created_at"},
			},
			"new_users": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := as.userCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.UserGrowthData
	for cursor.Next(ctx) {
		var data models.UserGrowthData
		// Parse the date from _id
		idDoc := cursor.Current.Lookup("_id").Document()
		year := idDoc.Lookup("year").AsInt64()
		month := idDoc.Lookup("month").AsInt64()
		day := idDoc.Lookup("day").AsInt64()

		date := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
		data.Date = date.Format("2006-01-02")
		data.NewUsers = cursor.Current.Lookup("new_users").AsInt64()
		data.Active = 0      // Would calculate active users for that day
		data.Retention = 0.0 // Would calculate retention rate

		results = append(results, data)
	}

	return results, nil
}

func (as *AdminService) getContentGrowthData(ctx context.Context, days int) ([]models.ContentGrowthData, error) {
	// Similar to user growth but for content
	since := time.Now().AddDate(0, 0, -days)

	// Get posts by day
	pipeline := []bson.M{
		{"$match": bson.M{"created_at": bson.M{"$gte": since}}},
		{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$created_at"},
				"month": bson.M{"$month": "$created_at"},
				"day":   bson.M{"$dayOfMonth": "$created_at"},
			},
			"posts": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := as.postCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.ContentGrowthData
	for cursor.Next(ctx) {
		var data models.ContentGrowthData
		idDoc := cursor.Current.Lookup("_id").Document()
		year := idDoc.Lookup("year").AsInt64()
		month := idDoc.Lookup("month").AsInt64()
		day := idDoc.Lookup("day").AsInt64()

		date := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
		data.Date = date.Format("2006-01-02")
		data.Posts = cursor.Current.Lookup("posts").AsInt64()
		// Would similarly get comments, stories, groups
		results = append(results, data)
	}

	return results, nil
}
func (as *AdminService) getEngagementMetrics(ctx context.Context) (models.EngagementMetrics, error) {
	metrics := models.EngagementMetrics{}

	// Get average session duration
	pipeline := []bson.M{
		{"$match": bson.M{"duration": bson.M{"$gt": 0}}},
		{"$group": bson.M{
			"_id":          nil,
			"avg_duration": bson.M{"$avg": "$duration"},
		}},
	}

	cursor, err := as.sessionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return metrics, fmt.Errorf("failed to execute aggregation pipeline: %w", err)
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		avgDurationRaw := cursor.Current.Lookup("avg_duration")
		var avgDuration float64
		switch avgDurationRaw.Type {
		case bson.TypeDouble:
			avgDuration = avgDurationRaw.Double()
		case bson.TypeInt32:
			avgDuration = float64(avgDurationRaw.Int32())
		case bson.TypeInt64:
			avgDuration = float64(avgDurationRaw.Int64())
		case bson.TypeNull:
			avgDuration = 0 // Handle null case (e.g., no sessions)
		default:
			return metrics, fmt.Errorf("unexpected type for avg_duration: %v", avgDurationRaw.Type)
		}
		metrics.AverageSessionDuration = avgDuration / 60000 // Convert to minutes
	} else {
		// No sessions found, set default value
		metrics.AverageSessionDuration = 0
	}

	// Get DAU, WAU, MAU
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)
	monthAgo := now.Add(-30 * 24 * time.Hour)

	metrics.DailyActiveUsers, err = as.userCollection.CountDocuments(ctx, bson.M{
		"last_login_at": bson.M{"$gte": dayAgo},
	})
	if err != nil {
		return metrics, fmt.Errorf("failed to count daily active users: %w", err)
	}

	metrics.WeeklyActiveUsers, err = as.userCollection.CountDocuments(ctx, bson.M{
		"last_login_at": bson.M{"$gte": weekAgo},
	})
	if err != nil {
		return metrics, fmt.Errorf("failed to count weekly active users: %w", err)
	}

	metrics.MonthlyActiveUsers, err = as.userCollection.CountDocuments(ctx, bson.M{
		"last_login_at": bson.M{"$gte": monthAgo},
	})
	if err != nil {
		return metrics, fmt.Errorf("failed to count monthly active users: %w", err)
	}

	return metrics, nil
}

func (as *AdminService) getGeographicData(ctx context.Context) ([]models.GeographicStats, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"location.country": bson.M{"$ne": ""}}},
		{"$group": bson.M{
			"_id":        "$location.country",
			"user_count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"user_count": -1}},
		{"$limit": 20},
	}

	cursor, err := as.userCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.GeographicStats
	for cursor.Next(ctx) {
		var stat models.GeographicStats
		stat.Country = cursor.Current.Lookup("_id").StringValue()
		stat.UserCount = cursor.Current.Lookup("user_count").AsInt64()
		results = append(results, stat)
	}

	return results, nil
}

func (as *AdminService) getDeviceStats(ctx context.Context) ([]models.DeviceStats, error) {
	// Would analyze user agent strings to get device types
	// For now, return mock data
	return []models.DeviceStats{
		{DeviceType: "Mobile", UserCount: 1500, Percentage: 60.0},
		{DeviceType: "Desktop", UserCount: 750, Percentage: 30.0},
		{DeviceType: "Tablet", UserCount: 250, Percentage: 10.0},
	}, nil
}

func (as *AdminService) getPlatformMetrics(ctx context.Context) (models.PlatformMetrics, error) {
	// Would get actual system metrics
	return models.PlatformMetrics{
		StorageUsed:          100, // GB
		BandwidthUsed:        500, // GB
		APIRequestsPerDay:    50000,
		AverageResponseTime:  150.0, // ms
		UptimePercentage:     99.9,
		ErrorRatePercentage:  0.1,
		ActiveWebSocketConns: 1200,
		QueuedJobs:           50,
		ProcessedJobs:        10000,
	}, nil
}

func (as *AdminService) getModerationQueueStats(ctx context.Context) (models.ModerationQueueStats, error) {
	stats := models.ModerationQueueStats{}

	stats.PendingReports, _ = as.reportCollection.CountDocuments(ctx, bson.M{"status": "pending"})
	// Would get other pending content counts

	return stats, nil
}

func (as *AdminService) getRevenueMetrics(ctx context.Context) (models.RevenueMetrics, error) {
	// Would integrate with payment/billing system
	return models.RevenueMetrics{
		TotalRevenue:        50000.0,
		MonthlyRevenue:      5000.0,
		AdvertisingRevenue:  3000.0,
		SubscriptionRevenue: 2000.0,
		RevenueGrowthRate:   15.5,
	}, nil
}

// User Management

func (as *AdminService) GetUsers(filter models.AdminUserFilter, limit, skip int) ([]models.AdminUserOverview, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Build filter query
	query := bson.M{}

	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.Role != "" {
		query["role"] = filter.Role
	}
	if filter.IsVerified != nil {
		query["is_verified"] = *filter.IsVerified
	}
	if !filter.CreatedAfter.IsZero() {
		query["created_at"] = bson.M{"$gte": filter.CreatedAfter}
	}
	if !filter.CreatedBefore.IsZero() {
		if createdAt, exists := query["created_at"]; exists {
			query["created_at"] = bson.M{
				"$gte": createdAt.(bson.M)["$gte"],
				"$lte": filter.CreatedBefore,
			}
		} else {
			query["created_at"] = bson.M{"$lte": filter.CreatedBefore}
		}
	}
	if filter.Location != "" {
		query["location.city"] = bson.M{"$regex": filter.Location, "$options": "i"}
	}

	// Get total count
	totalCount, err := as.userCollection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Build sort
	sortBy := "created_at"
	sortOrder := -1
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortOrder == "asc" {
		sortOrder = 1
	}

	// Get users with aggregation to include stats
	pipeline := []bson.M{
		{"$match": query},
		{"$sort": bson.M{sortBy: sortOrder}},
		{"$skip": skip},
		{"$limit": limit},
		// Join with posts to get post count
		{"$lookup": bson.M{
			"from":         "posts",
			"localField":   "_id",
			"foreignField": "user_id",
			"as":           "posts",
		}},
		{"$lookup": bson.M{
			"from":         "follows",
			"localField":   "_id",
			"foreignField": "followee_id",
			"as":           "followers",
		}},
		{"$lookup": bson.M{
			"from":         "follows",
			"localField":   "_id",
			"foreignField": "follower_id",
			"as":           "following",
		}},
		{"$addFields": bson.M{
			"post_count":      bson.M{"$size": "$posts"},
			"follower_count":  bson.M{"$size": "$followers"},
			"following_count": bson.M{"$size": "$following"},
		}},
		{"$project": bson.M{
			"posts":     0,
			"followers": 0,
			"following": 0,
			"password":  0,
		}},
	}

	cursor, err := as.userCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []models.AdminUserOverview
	for cursor.Next(ctx) {
		var user models.AdminUserOverview
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, totalCount, nil
}

func (as *AdminService) GetUserDetail(userID primitive.ObjectID) (*models.AdminUserDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get user
	var user models.User
	err := as.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	detail := &models.AdminUserDetail{
		User: models.AdminUserOverview{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Status:      user.UserStatus,
			Role:        user.Role,
			IsVerified:  user.IsVerified,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: user.LastLoginAt,
		},
	}

	// Get activity summary, recent posts, etc. in parallel
	errChan := make(chan error, 8)

	go func() {
		summary, err := as.getUserActivitySummary(ctx, userID)
		if err == nil {
			detail.ActivitySummary = *summary
		}
		errChan <- err
	}()

	go func() {
		posts, err := as.getUserRecentPosts(ctx, userID, 10)
		if err == nil {
			detail.RecentPosts = posts
		}
		errChan <- err
	}()

	go func() {
		comments, err := as.getUserRecentComments(ctx, userID, 10)
		if err == nil {
			detail.RecentComments = comments
		}
		errChan <- err
	}()

	go func() {
		reports, err := as.getReportsAgainstUser(ctx, userID, 10)
		if err == nil {
			detail.ReportsAgainst = reports
		}
		errChan <- err
	}()

	go func() {
		reports, err := as.getReportsByUser(ctx, userID, 10)
		if err == nil {
			detail.ReportsMade = reports
		}
		errChan <- err
	}()

	go func() {
		history, err := as.getModerationHistory(ctx, userID, 10)
		if err == nil {
			detail.ModerationHistory = history
		}
		errChan <- err
	}()

	go func() {
		events, err := as.getSecurityEvents(ctx, userID, 10)
		if err == nil {
			detail.SecurityEvents = events
		}
		errChan <- err
	}()

	go func() {
		logins, err := as.getLoginHistory(ctx, userID, 10)
		if err == nil {
			detail.LoginHistory = logins
		}
		errChan <- err
	}()

	// Wait for all goroutines
	for i := 0; i < 8; i++ {
		if err := <-errChan; err != nil {
			// Log error but continue
			fmt.Printf("Error getting user detail: %v\n", err)
		}
	}

	return detail, nil
}
func (as *AdminService) getUserActivitySummary(ctx context.Context, userID primitive.ObjectID) (*models.UserActivitySummary, error) {
	summary := &models.UserActivitySummary{}

	// Get session stats
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":            nil,
			"total_sessions": bson.M{"$sum": 1},
			"avg_duration":   bson.M{"$avg": "$duration"},
		}},
	}

	cursor, err := as.sessionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return summary, fmt.Errorf("failed to execute aggregation pipeline: %w", err)
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		summary.TotalSessions = cursor.Current.Lookup("total_sessions").AsInt64()

		// Look up the avg_duration field
		avgDurationRaw := cursor.Current.Lookup("avg_duration")
		var avgDuration float64
		switch avgDurationRaw.Type {
		case bson.TypeDouble:
			avgDuration = avgDurationRaw.Double()
		case bson.TypeInt32:
			avgDuration = float64(avgDurationRaw.Int32())
		case bson.TypeInt64:
			avgDuration = float64(avgDurationRaw.Int64())
		case bson.TypeNull:
			avgDuration = 0 // Handle null case (e.g., no sessions)
		default:
			return summary, fmt.Errorf("unexpected type for avg_duration: %v", avgDurationRaw.Type)
		}
		summary.AverageSessionTime = avgDuration / 60000 // Convert to minutes
	} else {
		// No results from aggregation
		summary.TotalSessions = 0
		summary.AverageSessionTime = 0
	}

	// Get recent activity counts
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)

	postsCount, err := as.postCollection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"created_at": bson.M{"$gte": weekAgo},
	})
	if err != nil {
		return summary, fmt.Errorf("failed to count posts: %w", err)
	}
	summary.PostsThisWeek = postsCount

	commentsCount, err := as.commentCollection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"created_at": bson.M{"$gte": weekAgo},
	})
	if err != nil {
		return summary, fmt.Errorf("failed to count comments: %w", err)
	}
	summary.CommentsThisWeek = commentsCount

	return summary, nil
}

func (as *AdminService) getUserRecentPosts(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.AdminPostOverview, error) {
	cursor, err := as.postCollection.Find(ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.AdminPostOverview
	for cursor.Next(ctx) {
		var post models.AdminPostOverview
		cursor.Decode(&post)
		posts = append(posts, post)
	}

	return posts, nil
}

func (as *AdminService) getUserRecentComments(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.AdminCommentOverview, error) {
	cursor, err := as.commentCollection.Find(ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.AdminCommentOverview
	for cursor.Next(ctx) {
		var comment models.AdminCommentOverview
		cursor.Decode(&comment)
		comments = append(comments, comment)
	}

	return comments, nil
}

func (as *AdminService) getReportsAgainstUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.AdminReportOverview, error) {
	cursor, err := as.reportCollection.Find(ctx,
		bson.M{"target_id": userID, "target_type": "user"},
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.AdminReportOverview
	for cursor.Next(ctx) {
		var report models.AdminReportOverview
		cursor.Decode(&report)
		reports = append(reports, report)
	}

	return reports, nil
}

func (as *AdminService) getReportsByUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.AdminReportOverview, error) {
	cursor, err := as.reportCollection.Find(ctx,
		bson.M{"reporter_id": userID},
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.AdminReportOverview
	for cursor.Next(ctx) {
		var report models.AdminReportOverview
		cursor.Decode(&report)
		reports = append(reports, report)
	}

	return reports, nil
}

func (as *AdminService) getModerationHistory(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.ModerationAction, error) {
	cursor, err := as.moderationCollection.Find(ctx,
		bson.M{"target_id": userID},
		options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var actions []models.ModerationAction
	for cursor.Next(ctx) {
		var action models.ModerationAction
		cursor.Decode(&action)
		actions = append(actions, action)
	}

	return actions, nil
}

func (as *AdminService) getSecurityEvents(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.SecurityEvent, error) {
	// Would query security events collection
	return []models.SecurityEvent{}, nil
}

func (as *AdminService) getLoginHistory(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.LoginEvent, error) {
	// Would query login events from sessions or separate collection
	return []models.LoginEvent{}, nil
}

// User Actions

func (as *AdminService) BulkUserAction(adminID primitive.ObjectID, request models.AdminUserActionRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, userIDStr := range request.UserIDs {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			continue
		}

		switch request.Action {
		case "suspend":
			err = as.suspendUser(ctx, userID, request.Reason, request.Duration)
		case "unsuspend":
			err = as.unsuspendUser(ctx, userID)
		case "verify":
			err = as.verifyUser(ctx, userID)
		case "unverify":
			err = as.unverifyUser(ctx, userID)
		case "delete":
			err = as.deleteUser(ctx, userID, request.Reason)
		case "warn":
			err = as.warnUser(ctx, userID, request.Reason, request.Note)
		}

		if err != nil {
			fmt.Printf("Error performing action %s on user %s: %v\n", request.Action, userIDStr, err)
		}

		// Log admin activity
		as.logAdminActivity(ctx, adminID, "user_action", userID,
			fmt.Sprintf("Performed %s action on user", request.Action),
			map[string]interface{}{
				"action": request.Action,
				"reason": request.Reason,
				"note":   request.Note,
			})
	}

	return nil
}

func (as *AdminService) suspendUser(ctx context.Context, userID primitive.ObjectID, reason string, duration *string) error {
	update := bson.M{
		"$set": bson.M{
			"status":            models.UserStatusSuspended,
			"suspended_at":      time.Now(),
			"suspension_reason": reason,
		},
	}

	if duration != nil && *duration != "" {
		// Parse duration and set unsuspend_at
		// This would require parsing duration string like "7d", "30d", etc.
	}

	_, err := as.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (as *AdminService) unsuspendUser(ctx context.Context, userID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"status": models.UserStatusActive,
		},
		"$unset": bson.M{
			"suspended_at":      "",
			"suspension_reason": "",
			"unsuspend_at":      "",
		},
	}

	_, err := as.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (as *AdminService) verifyUser(ctx context.Context, userID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_verified": true,
			"verified_at": time.Now(),
		},
	}

	_, err := as.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (as *AdminService) unverifyUser(ctx context.Context, userID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_verified": false,
		},
		"$unset": bson.M{
			"verified_at": "",
		},
	}

	_, err := as.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (as *AdminService) deleteUser(ctx context.Context, userID primitive.ObjectID, reason string) error {
	// Soft delete - mark as deleted
	update := bson.M{
		"$set": bson.M{
			"status":          models.UserStatusDeleted,
			"deleted_at":      time.Now(),
			"deletion_reason": reason,
		},
	}

	_, err := as.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (as *AdminService) warnUser(ctx context.Context, userID primitive.ObjectID, reason, note string) error {
	// Create warning record
	warning := bson.M{
		"_id":        primitive.NewObjectID(),
		"user_id":    userID,
		"reason":     reason,
		"note":       note,
		"created_at": time.Now(),
		"type":       "warning",
	}

	_, err := as.moderationCollection.InsertOne(ctx, warning)
	return err
}

// Content Management

func (as *AdminService) GetContent(contentType string, filter models.AdminContentFilter, limit, skip int) (interface{}, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var collection *mongo.Collection
	switch contentType {
	case "posts":
		collection = as.postCollection
	case "comments":
		collection = as.commentCollection
	case "stories":
		collection = as.storyCollection
	default:
		return nil, 0, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Build filter query
	query := bson.M{}
	if filter.AuthorID != "" {
		authorID, err := primitive.ObjectIDFromHex(filter.AuthorID)
		if err == nil {
			query["user_id"] = authorID
		}
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.IsFlagged != nil {
		if *filter.IsFlagged {
			query["flagged_at"] = bson.M{"$exists": true}
		} else {
			query["flagged_at"] = bson.M{"$exists": false}
		}
	}

	// Get total count
	totalCount, err := collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Build sort
	sortBy := "created_at"
	sortOrder := -1
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortOrder == "asc" {
		sortOrder = 1
	}

	// Get content
	cursor, err := collection.Find(ctx, query,
		options.Find().
			SetSort(bson.M{sortBy: sortOrder}).
			SetSkip(int64(skip)).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		var item interface{}
		cursor.Decode(&item)
		results = append(results, item)
	}

	return results, totalCount, nil
}

// System Configuration

func (as *AdminService) GetSystemConfiguration() (*models.SystemConfiguration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var config models.SystemConfiguration
	err := as.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default configuration
			return as.getDefaultConfiguration(), nil
		}
		return nil, err
	}

	return &config, nil
}

func (as *AdminService) UpdateSystemConfiguration(adminID primitive.ObjectID, config models.SystemConfiguration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config.UpdatedAt = time.Now()
	config.UpdatedBy = adminID

	opts := options.Replace().SetUpsert(true)
	_, err := as.configCollection.ReplaceOne(ctx, bson.M{}, config, opts)

	if err == nil {
		// Log admin activity
		as.logAdminActivity(ctx, adminID, "system_config", primitive.NilObjectID,
			"Updated system configuration",
			map[string]interface{}{"config_updated": true})
	}

	return err
}

func (as *AdminService) getDefaultConfiguration() *models.SystemConfiguration {
	return &models.SystemConfiguration{
		MaxPostLength:          280,
		MaxCommentLength:       1000,
		MaxFileSize:            10, // MB
		AllowedFileTypes:       []string{"jpg", "jpeg", "png", "gif", "mp4", "mov"},
		RateLimitPosts:         10,
		RateLimitComments:      50,
		RateLimitMessages:      100,
		AutoModeration:         true,
		RequireEmailVerify:     true,
		AllowGuestViewing:      true,
		MaintenanceMode:        false,
		RegistrationEnabled:    true,
		InviteOnlyMode:         false,
		ContentModerationLevel: "moderate",
		SpamDetectionEnabled:   true,
		AIContentFilter:        false,
		DataRetentionDays:      365,
		BackupEnabled:          true,
		BackupFrequency:        "daily",
		CDNEnabled:             true,
		CacheEnabled:           true,
		AnalyticsEnabled:       true,
		UpdatedAt:              time.Now(),
	}
}

// System Metrics

func (as *AdminService) GetSystemMetrics() (*models.SystemMetricsResponse, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := &models.SystemMetricsResponse{
		CPUUsage:    0.0,                            // Would need external library
		MemoryUsage: float64(m.Alloc) / 1024 / 1024, // MB
		DiskUsage:   0.0,                            // Would need external library
		DatabaseMetrics: models.DatabaseMetrics{
			ConnectionCount: 10, // Would get from connection pool
			ActiveQueries:   5,
			QueryLatency:    50.0,
			DatabaseSize:    1024, // MB
			CollectionCount: 20,
			IndexCount:      50,
		},
		CacheMetrics: models.CacheMetrics{
			HitRate:         85.5,
			MissRate:        14.5,
			CachedItemCount: 10000,
			CacheSize:       256, // MB
			EvictionCount:   100,
		},
		QueueMetrics: models.QueueMetrics{
			PendingJobs:        50,
			ProcessingJobs:     10,
			CompletedJobs:      1000,
			FailedJobs:         5,
			AverageProcessTime: 150.0,
		},
		ErrorMetrics: models.ErrorMetrics{
			ErrorRate5xx:   0.1,
			ErrorRate4xx:   2.5,
			TotalErrors:    100,
			CriticalErrors: 2,
			DatabaseErrors: 1,
			AuthErrors:     10,
		},
		PerformanceMetrics: models.PerformanceMetrics{
			AverageResponseTime: 150.0,
			P50ResponseTime:     120.0,
			P95ResponseTime:     300.0,
			P99ResponseTime:     500.0,
			RequestsPerSecond:   100.0,
			ConcurrentUsers:     1000,
		},
	}

	return metrics, nil
}

// Admin Activity Logging

func (as *AdminService) logAdminActivity(ctx context.Context, adminID primitive.ObjectID, targetType string, targetID primitive.ObjectID, description string, metadata map[string]interface{}) {
	activity := models.AdminActivity{
		AdminID:     adminID,
		Action:      targetType,
		TargetType:  targetType,
		TargetID:    targetID,
		Description: description,
		Timestamp:   time.Now(),
		Metadata:    metadata,
		Severity:    "info",
		Category:    "admin_action",
	}

	// Insert activity log (don't block on error)
	go func() {
		as.adminActivityCollection.InsertOne(context.Background(), activity)
	}()
}

func (as *AdminService) GetAdminActivities(limit, skip int) ([]models.AdminActivity, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get total count
	totalCount, err := as.adminActivityCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Get activities
	cursor, err := as.adminActivityCollection.Find(ctx, bson.M{},
		options.Find().
			SetSort(bson.M{"timestamp": -1}).
			SetSkip(int64(skip)).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var activities []models.AdminActivity
	for cursor.Next(ctx) {
		var activity models.AdminActivity
		cursor.Decode(&activity)
		activities = append(activities, activity)
	}

	return activities, totalCount, nil
}
