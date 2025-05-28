// scripts/cleanup.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Cleanup configuration
type CleanupConfig struct {
	MongoURI             string
	DatabaseName         string
	Timeout              time.Duration
	ExpiredStoriesAge    time.Duration
	ExpiredSessionsAge   time.Duration
	ExpiredNotifications time.Duration
	ExpiredAnalytics     time.Duration
	ExpiredMessages      time.Duration
	InactiveUsersAge     time.Duration
	BatchSize            int64
}

// Cleanup statistics
type CleanupStats struct {
	ExpiredStories       int64
	ExpiredSessions      int64
	ExpiredNotifications int64
	ExpiredAnalytics     int64
	ExpiredMessages      int64
	InactiveUsers        int64
	OrphanedData         int64
	OptimizedCollections int64
	Duration             time.Duration
}

func main() {
	flag.Parse()

	// Setup logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	config := CleanupConfig{
		MongoURI:             *mongoURI,
		DatabaseName:         *databaseName,
		Timeout:              *timeout,
		ExpiredStoriesAge:    *expiredStoriesAge,
		ExpiredSessionsAge:   *expiredSessionsAge,
		ExpiredNotifications: *expiredNotifications,
		ExpiredAnalytics:     *expiredAnalytics,
		ExpiredMessages:      *expiredMessages,
		InactiveUsersAge:     *inactiveUsersAge,
		BatchSize:            *batchSize,
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Connect to MongoDB
	client, err := connectMongoDB(ctx, config.MongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	db := client.Database(config.DatabaseName)

	// Execute cleanup operation
	cleaner := &DatabaseCleaner{
		db:     db,
		config: config,
		stats:  &CleanupStats{},
	}

	startTime := time.Now()

	switch *operation {
	case "all":
		err = cleaner.CleanupAll(ctx)
	case "stories":
		err = cleaner.CleanupExpiredStories(ctx)
	case "sessions":
		err = cleaner.CleanupExpiredSessions(ctx)
	case "notifications":
		err = cleaner.CleanupExpiredNotifications(ctx)
	case "analytics":
		err = cleaner.CleanupExpiredAnalytics(ctx)
	case "messages":
		err = cleaner.CleanupExpiredMessages(ctx)
	case "users":
		err = cleaner.CleanupInactiveUsers(ctx)
	case "optimize":
		err = cleaner.OptimizeCollections(ctx)
	case "stats":
		err = cleaner.ShowDatabaseStats(ctx)
	default:
		log.Fatalf("Unknown operation: %s", *operation)
	}

	if err != nil {
		log.Fatal("Cleanup failed:", err)
	}

	cleaner.stats.Duration = time.Since(startTime)
	cleaner.PrintCleanupSummary()
}

// DatabaseCleaner handles database cleanup operations
type DatabaseCleaner struct {
	db     *mongo.Database
	config CleanupConfig
	stats  *CleanupStats
}

func connectMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
	log.Printf("Connecting to MongoDB: %s", uri)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB")
	return client, nil
}

func (dc *DatabaseCleaner) CleanupAll(ctx context.Context) error {
	log.Println("Starting comprehensive database cleanup...")

	if *dryRun {
		log.Println("DRY RUN: Showing what would be cleaned")
		return dc.performDryRun(ctx)
	}

	if !*force {
		if !confirmAction("Perform comprehensive database cleanup?") {
			log.Println("Cleanup cancelled by user")
			return nil
		}
	}

	// Run all cleanup operations
	operations := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"Expired Stories", dc.CleanupExpiredStories},
		{"Expired Sessions", dc.CleanupExpiredSessions},
		{"Expired Notifications", dc.CleanupExpiredNotifications},
		{"Expired Analytics", dc.CleanupExpiredAnalytics},
		{"Expired Messages", dc.CleanupExpiredMessages},
		{"Inactive Users", dc.CleanupInactiveUsers},
		{"Orphaned Data", dc.CleanupOrphanedData},
		{"Collection Optimization", dc.OptimizeCollections},
	}

	for _, op := range operations {
		log.Printf("Running: %s", op.name)
		if err := op.fn(ctx); err != nil {
			log.Printf("Warning: %s failed: %v", op.name, err)
		}
	}

	log.Println("Comprehensive cleanup completed")
	return nil
}

func (dc *DatabaseCleaner) CleanupExpiredStories(ctx context.Context) error {
	log.Println("Cleaning up expired stories...")

	collection := dc.db.Collection("stories")
	cutoffTime := time.Now().Add(-dc.config.ExpiredStoriesAge)

	filter := bson.M{
		"$or": []bson.M{
			{"is_expired": true},
			{"expires_at": bson.M{"$lt": cutoffTime}},
		},
		"is_highlighted": false, // Don't delete highlighted stories
	}

	if *dryRun {
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		log.Printf("DRY RUN: Would delete %d expired stories", count)
		return nil
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired stories: %w", err)
	}

	dc.stats.ExpiredStories = result.DeletedCount
	log.Printf("Deleted %d expired stories", result.DeletedCount)
	return nil
}

func (dc *DatabaseCleaner) CleanupExpiredSessions(ctx context.Context) error {
	log.Println("Cleaning up expired sessions...")

	collection := dc.db.Collection("user_sessions")
	cutoffTime := time.Now().Add(-dc.config.ExpiredSessionsAge)

	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"last_activity": bson.M{"$lt": cutoffTime}},
			{"is_active": false},
		},
	}

	if *dryRun {
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		log.Printf("DRY RUN: Would delete %d expired sessions", count)
		return nil
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	dc.stats.ExpiredSessions = result.DeletedCount
	log.Printf("Deleted %d expired sessions", result.DeletedCount)
	return nil
}

func (dc *DatabaseCleaner) CleanupExpiredNotifications(ctx context.Context) error {
	log.Println("Cleaning up expired notifications...")

	collection := dc.db.Collection("notifications")
	cutoffTime := time.Now().Add(-dc.config.ExpiredNotifications)

	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"created_at": bson.M{"$lt": cutoffTime}},
		},
		"is_read": true, // Only delete read notifications
	}

	if *dryRun {
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		log.Printf("DRY RUN: Would delete %d expired notifications", count)
		return nil
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired notifications: %w", err)
	}

	dc.stats.ExpiredNotifications = result.DeletedCount
	log.Printf("Deleted %d expired notifications", result.DeletedCount)
	return nil
}

func (dc *DatabaseCleaner) CleanupExpiredAnalytics(ctx context.Context) error {
	log.Println("Cleaning up expired analytics...")

	collection := dc.db.Collection("analytics_events")
	cutoffTime := time.Now().Add(-dc.config.ExpiredAnalytics)

	filter := bson.M{
		"created_at": bson.M{"$lt": cutoffTime},
	}

	if *dryRun {
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		log.Printf("DRY RUN: Would delete %d expired analytics events", count)
		return nil
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired analytics: %w", err)
	}

	dc.stats.ExpiredAnalytics = result.DeletedCount
	log.Printf("Deleted %d expired analytics events", result.DeletedCount)
	return nil
}

func (dc *DatabaseCleaner) CleanupExpiredMessages(ctx context.Context) error {
	log.Println("Cleaning up expired messages...")

	collection := dc.db.Collection("messages")
	filter := bson.M{
		"is_expired": true,
		"expires_at": bson.M{"$lt": time.Now()},
	}

	if *dryRun {
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		log.Printf("DRY RUN: Would delete %d expired messages", count)
		return nil
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired messages: %w", err)
	}

	dc.stats.ExpiredMessages = result.DeletedCount
	log.Printf("Deleted %d expired messages", result.DeletedCount)
	return nil
}

func (dc *DatabaseCleaner) CleanupInactiveUsers(ctx context.Context) error {
	log.Println("Cleaning up inactive users...")

	collection := dc.db.Collection("users")
	cutoffTime := time.Now().Add(-dc.config.InactiveUsersAge)

	filter := bson.M{
		"last_active_at":  bson.M{"$lt": cutoffTime},
		"is_active":       false,
		"posts_count":     0,
		"followers_count": 0,
		"following_count": 0,
	}

	if *dryRun {
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		log.Printf("DRY RUN: Would delete %d inactive users", count)
		return nil
	}

	// First, get the user IDs to delete
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return fmt.Errorf("failed to find inactive users: %w", err)
	}
	defer cursor.Close(ctx)

	var userIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var user struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		userIDs = append(userIDs, user.ID)
	}

	if len(userIDs) == 0 {
		log.Println("No inactive users to delete")
		return nil
	}

	// Clean up related data first
	if err := dc.cleanupUserRelatedData(ctx, userIDs); err != nil {
		log.Printf("Warning: Failed to cleanup user related data: %v", err)
	}

	// Delete users
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete inactive users: %w", err)
	}

	dc.stats.InactiveUsers = result.DeletedCount
	log.Printf("Deleted %d inactive users", result.DeletedCount)
	return nil
}

func (dc *DatabaseCleaner) CleanupOrphanedData(ctx context.Context) error {
	log.Println("Cleaning up orphaned data...")

	totalCleaned := int64(0)

	// Clean orphaned comments (posts that no longer exist)
	if count, err := dc.cleanupOrphanedComments(ctx); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned comments: %v", err)
	} else {
		totalCleaned += count
	}

	// Clean orphaned likes (targets that no longer exist)
	if count, err := dc.cleanupOrphanedLikes(ctx); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned likes: %v", err)
	} else {
		totalCleaned += count
	}

	// Clean orphaned notifications (actors/recipients that no longer exist)
	if count, err := dc.cleanupOrphanedNotifications(ctx); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned notifications: %v", err)
	} else {
		totalCleaned += count
	}

	// Clean orphaned follows (users that no longer exist)
	if count, err := dc.cleanupOrphanedFollows(ctx); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned follows: %v", err)
	} else {
		totalCleaned += count
	}

	dc.stats.OrphanedData = totalCleaned
	log.Printf("Cleaned up %d orphaned records", totalCleaned)
	return nil
}

func (dc *DatabaseCleaner) OptimizeCollections(ctx context.Context) error {
	log.Println("Optimizing collections...")

	collections := []string{
		"users", "posts", "comments", "stories", "messages", "conversations",
		"follows", "likes", "notifications", "groups", "events", "hashtags",
	}

	optimized := int64(0)
	for _, collectionName := range collections {
		if err := dc.optimizeCollection(ctx, collectionName); err != nil {
			log.Printf("Warning: Failed to optimize collection %s: %v", collectionName, err)
		} else {
			optimized++
		}
	}

	dc.stats.OptimizedCollections = optimized
	log.Printf("Optimized %d collections", optimized)
	return nil
}

func (dc *DatabaseCleaner) ShowDatabaseStats(ctx context.Context) error {
	log.Println("Gathering database statistics...")

	stats, err := dc.db.RunCommand(ctx, bson.D{{"dbStats", 1}}).DecodeBytes()
	if err != nil {
		return fmt.Errorf("failed to get database stats: %w", err)
	}

	fmt.Println("\nDatabase Statistics:")
	fmt.Println("===================")

	// Extract key metrics from stats
	if dataSize, err := stats.LookupErr("dataSize"); err == nil {
		fmt.Printf("Data Size: %.2f MB\n", float64(dataSize.AsInt64())/1024/1024)
	}

	if storageSize, err := stats.LookupErr("storageSize"); err == nil {
		fmt.Printf("Storage Size: %.2f MB\n", float64(storageSize.AsInt64())/1024/1024)
	}

	if indexSize, err := stats.LookupErr("indexSize"); err == nil {
		fmt.Printf("Index Size: %.2f MB\n", float64(indexSize.AsInt64())/1024/1024)
	}

	if collections, err := stats.LookupErr("collections"); err == nil {
		fmt.Printf("Collections: %d\n", collections.AsInt32())
	}

	if indexes, err := stats.LookupErr("indexes"); err == nil {
		fmt.Printf("Indexes: %d\n", indexes.AsInt32())
	}

	// Show collection-specific stats
	fmt.Println("\nCollection Statistics:")
	fmt.Println("=====================")

	collections := []string{
		"users", "posts", "comments", "stories", "messages",
		"notifications", "follows", "likes", "groups", "events",
	}

	for _, collectionName := range collections {
		if err := dc.showCollectionStats(ctx, collectionName); err != nil {
			log.Printf("Warning: Failed to get stats for %s: %v", collectionName, err)
		}
	}

	return nil
}

// Helper methods

func (dc *DatabaseCleaner) performDryRun(ctx context.Context) error {
	log.Println("=== DRY RUN RESULTS ===")

	operations := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"Expired Stories", dc.CleanupExpiredStories},
		{"Expired Sessions", dc.CleanupExpiredSessions},
		{"Expired Notifications", dc.CleanupExpiredNotifications},
		{"Expired Analytics", dc.CleanupExpiredAnalytics},
		{"Expired Messages", dc.CleanupExpiredMessages},
		{"Inactive Users", dc.CleanupInactiveUsers},
	}

	for _, op := range operations {
		if err := op.fn(ctx); err != nil {
			log.Printf("Error checking %s: %v", op.name, err)
		}
	}

	log.Println("=== END DRY RUN ===")
	return nil
}

func (dc *DatabaseCleaner) cleanupUserRelatedData(ctx context.Context, userIDs []primitive.ObjectID) error {
	collections := []string{"posts", "comments", "stories", "follows", "likes", "notifications"}

	for _, collectionName := range collections {
		collection := dc.db.Collection(collectionName)

		var filter bson.M
		switch collectionName {
		case "follows":
			filter = bson.M{
				"$or": []bson.M{
					{"follower_id": bson.M{"$in": userIDs}},
					{"followee_id": bson.M{"$in": userIDs}},
				},
			}
		case "notifications":
			filter = bson.M{
				"$or": []bson.M{
					{"recipient_id": bson.M{"$in": userIDs}},
					{"actor_id": bson.M{"$in": userIDs}},
				},
			}
		default:
			filter = bson.M{"user_id": bson.M{"$in": userIDs}}
		}

		if _, err := collection.DeleteMany(ctx, filter); err != nil {
			return fmt.Errorf("failed to cleanup %s: %w", collectionName, err)
		}
	}

	return nil
}

func (dc *DatabaseCleaner) cleanupOrphanedComments(ctx context.Context) (int64, error) {
	// Find comments with non-existent posts
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "posts",
				"localField":   "post_id",
				"foreignField": "_id",
				"as":           "post",
			},
		},
		{
			"$match": bson.M{
				"post": bson.M{"$size": 0},
			},
		},
	}

	if *dryRun {
		cursor, err := dc.db.Collection("comments").Aggregate(ctx, pipeline)
		if err != nil {
			return 0, err
		}
		defer cursor.Close(ctx)

		count := int64(0)
		for cursor.Next(ctx) {
			count++
		}
		log.Printf("DRY RUN: Would delete %d orphaned comments", count)
		return count, nil
	}

	// Get orphaned comment IDs
	cursor, err := dc.db.Collection("comments").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var orphanedIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var comment struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&comment); err != nil {
			continue
		}
		orphanedIDs = append(orphanedIDs, comment.ID)
	}

	if len(orphanedIDs) == 0 {
		return 0, nil
	}

	result, err := dc.db.Collection("comments").DeleteMany(ctx, bson.M{
		"_id": bson.M{"$in": orphanedIDs},
	})
	if err != nil {
		return 0, err
	}

	log.Printf("Deleted %d orphaned comments", result.DeletedCount)
	return result.DeletedCount, nil
}

func (dc *DatabaseCleaner) cleanupOrphanedLikes(ctx context.Context) (int64, error) {
	// This is a simplified implementation
	// In a real scenario, you'd need to check against multiple target collections
	totalDeleted := int64(0)

	// Check likes for non-existent posts
	pipeline := []bson.M{
		{
			"$match": bson.M{"target_type": "post"},
		},
		{
			"$lookup": bson.M{
				"from":         "posts",
				"localField":   "target_id",
				"foreignField": "_id",
				"as":           "target",
			},
		},
		{
			"$match": bson.M{
				"target": bson.M{"$size": 0},
			},
		},
	}

	cursor, err := dc.db.Collection("likes").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var orphanedIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var like struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&like); err != nil {
			continue
		}
		orphanedIDs = append(orphanedIDs, like.ID)
	}

	if len(orphanedIDs) > 0 {
		if *dryRun {
			log.Printf("DRY RUN: Would delete %d orphaned likes", len(orphanedIDs))
			return int64(len(orphanedIDs)), nil
		}

		result, err := dc.db.Collection("likes").DeleteMany(ctx, bson.M{
			"_id": bson.M{"$in": orphanedIDs},
		})
		if err != nil {
			return 0, err
		}
		totalDeleted += result.DeletedCount
		log.Printf("Deleted %d orphaned likes", result.DeletedCount)
	}

	return totalDeleted, nil
}

func (dc *DatabaseCleaner) cleanupOrphanedNotifications(ctx context.Context) (int64, error) {
	// Check notifications for non-existent users
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "recipient_id",
				"foreignField": "_id",
				"as":           "recipient",
			},
		},
		{
			"$match": bson.M{
				"recipient": bson.M{"$size": 0},
			},
		},
	}

	cursor, err := dc.db.Collection("notifications").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var orphanedIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var notification struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&notification); err != nil {
			continue
		}
		orphanedIDs = append(orphanedIDs, notification.ID)
	}

	if len(orphanedIDs) == 0 {
		return 0, nil
	}

	if *dryRun {
		log.Printf("DRY RUN: Would delete %d orphaned notifications", len(orphanedIDs))
		return int64(len(orphanedIDs)), nil
	}

	result, err := dc.db.Collection("notifications").DeleteMany(ctx, bson.M{
		"_id": bson.M{"$in": orphanedIDs},
	})
	if err != nil {
		return 0, err
	}

	log.Printf("Deleted %d orphaned notifications", result.DeletedCount)
	return result.DeletedCount, nil
}

func (dc *DatabaseCleaner) cleanupOrphanedFollows(ctx context.Context) (int64, error) {
	// Check follows for non-existent users
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "follower_id",
				"foreignField": "_id",
				"as":           "follower",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "followee_id",
				"foreignField": "_id",
				"as":           "followee",
			},
		},
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"follower": bson.M{"$size": 0}},
					{"followee": bson.M{"$size": 0}},
				},
			},
		},
	}

	cursor, err := dc.db.Collection("follows").Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var orphanedIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var follow struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&follow); err != nil {
			continue
		}
		orphanedIDs = append(orphanedIDs, follow.ID)
	}

	if len(orphanedIDs) == 0 {
		return 0, nil
	}

	if *dryRun {
		log.Printf("DRY RUN: Would delete %d orphaned follows", len(orphanedIDs))
		return int64(len(orphanedIDs)), nil
	}

	result, err := dc.db.Collection("follows").DeleteMany(ctx, bson.M{
		"_id": bson.M{"$in": orphanedIDs},
	})
	if err != nil {
		return 0, err
	}

	log.Printf("Deleted %d orphaned follows", result.DeletedCount)
	return result.DeletedCount, nil
}

func (dc *DatabaseCleaner) optimizeCollection(ctx context.Context, collectionName string) error {
	collection := dc.db.Collection(collectionName)

	// Reindex the collection
	if err := collection.Indexes().DropAll(ctx); err != nil {
		return fmt.Errorf("failed to drop indexes: %w", err)
	}

	// The indexes will be recreated by the application or migration system
	log.Printf("Optimized collection: %s", collectionName)
	return nil
}

func (dc *DatabaseCleaner) showCollectionStats(ctx context.Context, collectionName string) error {
	collection := dc.db.Collection(collectionName)

	count, err := collection.EstimatedDocumentCount(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%-15s: %10d documents\n", collectionName, count)
	return nil
}

func (dc *DatabaseCleaner) PrintCleanupSummary() {
	fmt.Println("\n=== CLEANUP SUMMARY ===")
	fmt.Printf("Expired Stories:       %d\n", dc.stats.ExpiredStories)
	fmt.Printf("Expired Sessions:      %d\n", dc.stats.ExpiredSessions)
	fmt.Printf("Expired Notifications: %d\n", dc.stats.ExpiredNotifications)
	fmt.Printf("Expired Analytics:     %d\n", dc.stats.ExpiredAnalytics)
	fmt.Printf("Expired Messages:      %d\n", dc.stats.ExpiredMessages)
	fmt.Printf("Inactive Users:        %d\n", dc.stats.InactiveUsers)
	fmt.Printf("Orphaned Data:         %d\n", dc.stats.OrphanedData)
	fmt.Printf("Optimized Collections: %d\n", dc.stats.OptimizedCollections)
	fmt.Printf("Total Duration:        %v\n", dc.stats.Duration)
	fmt.Println("========================")
}

func confirmAction(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "YES"
}
