// migrations/002_add_social_features.go
package migrations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetSocialFeaturesMigration returns the social features migration
func GetSocialFeaturesMigration() Migration {
	return Migration{
		ID:          "002_add_social_features",
		Description: "Add indexes and collections for advanced social features",
		Up:          addSocialFeatures,
		Down:        removeSocialFeatures,
	}
}

func addSocialFeatures(ctx context.Context, db *mongo.Database) error {
	log.Println("Adding social features...")

	// Create additional collections
	if err := createGroupMembersCollection(ctx, db); err != nil {
		return err
	}

	if err := createEventRSVPCollection(ctx, db); err != nil {
		return err
	}

	if err := createStoryViewsCollection(ctx, db); err != nil {
		return err
	}

	if err := createStoryHighlightsCollection(ctx, db); err != nil {
		return err
	}

	if err := createBlockedUsersCollection(ctx, db); err != nil {
		return err
	}

	if err := createUserSessionsCollection(ctx, db); err != nil {
		return err
	}

	if err := createFeedCollection(ctx, db); err != nil {
		return err
	}

	if err := createAnalyticsCollection(ctx, db); err != nil {
		return err
	}

	// Add compound indexes for better performance
	if err := addAdvancedPostIndexes(ctx, db); err != nil {
		return err
	}

	if err := addAdvancedUserIndexes(ctx, db); err != nil {
		return err
	}

	if err := addAdvancedMessageIndexes(ctx, db); err != nil {
		return err
	}

	// Create capped collections for real-time features
	if err := createRealtimeCollections(ctx, db); err != nil {
		return err
	}

	log.Println("Social features added successfully")
	return nil
}

func createGroupMembersCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("group_members")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"group_id", 1}, {"user_id", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"group_id", 1}}},
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"role", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"joined_at", -1}}},
		{Keys: bson.D{{"last_active_at", -1}}},
		{Keys: bson.D{{"is_muted", 1}}},
		{Keys: bson.D{{"notifications_enabled", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"group_id", 1}, {"status", 1}, {"joined_at", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"status", 1}, {"joined_at", -1}}},
		{Keys: bson.D{{"group_id", 1}, {"role", 1}}},
		{Keys: bson.D{{"group_id", 1}, {"last_active_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Group members collection indexes created")
	return nil
}

func createEventRSVPCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("event_rsvps")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"event_id", 1}, {"user_id", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"event_id", 1}}},
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"responded_at", -1}}},
		{Keys: bson.D{{"checked_in", 1}}},
		{Keys: bson.D{{"checked_in_at", -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"no_show", 1}}},
		{Keys: bson.D{{"guest_count", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"event_id", 1}, {"status", 1}, {"responded_at", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"status", 1}, {"responded_at", -1}}},
		{Keys: bson.D{{"event_id", 1}, {"checked_in", 1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Event RSVPs collection indexes created")
	return nil
}

func createStoryViewsCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("story_views")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"story_id", 1}, {"user_id", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"story_id", 1}}},
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"view_duration", -1}}},
		{Keys: bson.D{{"watched_fully", 1}}},
		{Keys: bson.D{{"source", 1}}},
		{Keys: bson.D{{"device_type", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"story_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"story_id", 1}, {"watched_fully", 1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Story views collection indexes created")
	return nil
}

func createStoryHighlightsCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("story_highlights")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"title", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"updated_at", -1}}},
		{Keys: bson.D{{"is_active", 1}}},
		{Keys: bson.D{{"order", 1}}},
		{Keys: bson.D{{"stories_count", -1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"user_id", 1}, {"order", 1}}},
		{Keys: bson.D{{"user_id", 1}, {"is_active", 1}, {"order", 1}}},
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Story highlights collection indexes created")
	return nil
}

func createBlockedUsersCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("blocked_users")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"blocker_id", 1}, {"blocked_id", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"blocker_id", 1}}},
		{Keys: bson.D{{"blocked_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"reason", 1}}},
		{Keys: bson.D{{"is_active", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"blocker_id", 1}, {"is_active", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"blocked_id", 1}, {"is_active", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Blocked users collection indexes created")
	return nil
}

func createUserSessionsCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("user_sessions")

	// Create unique index for session_token first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"session_token", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"last_activity", -1}}},
		{Keys: bson.D{{"is_active", 1}}},
		{Keys: bson.D{{"device_type", 1}}},
		{Keys: bson.D{{"ip_address", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"user_id", 1}, {"is_active", 1}, {"last_activity", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"device_type", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	// Create TTL index for expired sessions
	if err := EnsureTTLIndex(ctx, collection, "expires_at", 0); err != nil {
		return err
	}

	log.Println("User sessions collection indexes created")
	return nil
}

func createFeedCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("user_feeds")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"user_id", 1}, {"post_id", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"post_id", 1}}},
		{Keys: bson.D{{"score", -1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"feed_type", 1}}},
		{Keys: bson.D{{"is_seen", 1}}},
		{Keys: bson.D{{"interaction_score", -1}}},
		{Keys: bson.D{{"author_id", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"user_id", 1}, {"feed_type", 1}, {"score", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"is_seen", 1}, {"score", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"author_id", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("User feeds collection indexes created")
	return nil
}

func createAnalyticsCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("analytics_events")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"event_type", 1}}},
		{Keys: bson.D{{"timestamp", -1}}},
		{Keys: bson.D{{"session_id", 1}}},
		{Keys: bson.D{{"resource_type", 1}}},
		{Keys: bson.D{{"resource_id", 1}}},
		{Keys: bson.D{{"platform", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		// Compound indexes
		{Keys: bson.D{{"event_type", 1}, {"timestamp", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"event_type", 1}, {"timestamp", -1}}},
		{Keys: bson.D{{"resource_type", 1}, {"resource_id", 1}, {"timestamp", -1}}},
		{Keys: bson.D{{"session_id", 1}, {"timestamp", -1}}},
		{Keys: bson.D{{"platform", 1}, {"event_type", 1}, {"timestamp", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	// Create TTL index for old analytics data (keep for 90 days)
	if err := EnsureTTLIndex(ctx, collection, "created_at", 90*24*60*60); err != nil {
		return err
	}

	log.Println("Analytics collection indexes created")
	return nil
}

func addAdvancedPostIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("posts")

	additionalIndexes := []mongo.IndexModel{
		// Performance indexes for feed generation
		{Keys: bson.D{{"user_id", 1}, {"visibility", 1}, {"is_published", 1}, {"engagement_rate", -1}}},
		{Keys: bson.D{{"visibility", 1}, {"is_published", 1}, {"is_hidden", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"hashtags", 1}, {"visibility", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"mentions", 1}, {"visibility", 1}, {"created_at", -1}}},
		// Trending/popular content indexes
		{Keys: bson.D{{"likes_count", -1}, {"created_at", -1}}},
		{Keys: bson.D{{"comments_count", -1}, {"created_at", -1}}},
		{Keys: bson.D{{"shares_count", -1}, {"created_at", -1}}},
		// Content moderation indexes
		{Keys: bson.D{{"is_reported", 1}, {"reports_count", -1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, additionalIndexes); err != nil {
		return err
	}

	log.Println("Advanced post indexes created")
	return nil
}

func addAdvancedUserIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	additionalIndexes := []mongo.IndexModel{
		// User discovery and search indexes
		{Keys: bson.D{{"is_verified", 1}, {"followers_count", -1}}},
		{Keys: bson.D{{"is_private", 1}, {"is_active", 1}, {"last_active_at", -1}}},
		{Keys: bson.D{{"location", 1}, {"is_private", 1}}},
		// Activity and engagement indexes
		{Keys: bson.D{{"posts_count", -1}, {"followers_count", -1}}},
		{Keys: bson.D{{"total_likes_received", -1}, {"followers_count", -1}}},
		// Moderation indexes
		{Keys: bson.D{{"is_suspended", 1}, {"reported_by_count", -1}}},
		{Keys: bson.D{{"role", 1}, {"is_active", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, additionalIndexes); err != nil {
		return err
	}

	log.Println("Advanced user indexes created")
	return nil
}

func addAdvancedMessageIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("messages")

	additionalIndexes := []mongo.IndexModel{
		// Message search and filtering
		{Keys: bson.D{{"conversation_id", 1}, {"sender_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"conversation_id", 1}, {"content_type", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"conversation_id", 1}, {"status", 1}, {"created_at", -1}}},
		// Thread and reply indexes
		{Keys: bson.D{{"thread_id", 1}, {"is_thread_root", 1}, {"created_at", 1}}},
		{Keys: bson.D{{"reply_to_message_id", 1}, {"created_at", 1}}},
		// Media and file message indexes
		{Keys: bson.D{{"conversation_id", 1}, {"content_type", 1}, {"duration", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, additionalIndexes); err != nil {
		return err
	}

	log.Println("Advanced message indexes created")
	return nil
}

func createRealtimeCollections(ctx context.Context, db *mongo.Database) error {
	// Handle existing capped collections by dropping and recreating them as regular collections

	// Check and handle typing_indicators collection
	if err := recreateCollectionAsRegular(ctx, db, "typing_indicators"); err != nil {
		return err
	}

	// Check and handle user_presence collection
	if err := recreateCollectionAsRegular(ctx, db, "user_presence"); err != nil {
		return err
	}

	// Now create indexes on the regular collections

	// Typing indicators (regular collection with TTL)
	typingCollection := db.Collection("typing_indicators")

	typingIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"conversation_id", 1}}},
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, typingCollection, typingIndexes); err != nil {
		return err
	}

	// Create TTL index to auto-expire typing indicators after 30 seconds
	if err := EnsureTTLIndex(ctx, typingCollection, "created_at", 30); err != nil {
		return err
	}

	// Online presence (regular collection with TTL)
	presenceCollection := db.Collection("user_presence")

	presenceIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"last_seen", -1}}},
	}

	if err := CreateIndexesSafely(ctx, presenceCollection, presenceIndexes); err != nil {
		return err
	}

	// Create TTL index to auto-expire presence data after 5 minutes
	if err := EnsureTTLIndex(ctx, presenceCollection, "last_seen", 300); err != nil {
		return err
	}

	log.Println("Realtime collections created with TTL indexes")
	return nil
}

// Alternative: createRealtimeCollectionsCapped - Use this if you specifically need capped collections
func createRealtimeCollectionsCapped(ctx context.Context, db *mongo.Database) error {
	// Typing indicators (capped collection - documents expire by size limit)
	typingOptions := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024 * 1024) // 1MB

	// Check if collection already exists
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": "typing_indicators_capped"})
	if err != nil {
		return err
	}

	if len(collections) == 0 {
		if err := db.CreateCollection(ctx, "typing_indicators_capped", typingOptions); err != nil {
			log.Printf("Typing indicators capped collection creation warning: %v", err)
		} else {
			log.Println("Created capped collection: typing_indicators_capped (1MB)")
		}
	}

	typingCollection := db.Collection("typing_indicators_capped")
	typingIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"conversation_id", 1}}},
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, typingCollection, typingIndexes); err != nil {
		return err
	}
	// Note: No TTL index for capped collections - they auto-expire by size

	// Online presence (capped collection)
	presenceOptions := options.CreateCollection().SetCapped(true).SetSizeInBytes(2 * 1024 * 1024) // 2MB

	// Check if collection already exists
	collections, err = db.ListCollectionNames(ctx, bson.M{"name": "user_presence_capped"})
	if err != nil {
		return err
	}

	if len(collections) == 0 {
		if err := db.CreateCollection(ctx, "user_presence_capped", presenceOptions); err != nil {
			log.Printf("User presence capped collection creation warning: %v", err)
		} else {
			log.Println("Created capped collection: user_presence_capped (2MB)")
		}
	}

	presenceCollection := db.Collection("user_presence_capped")
	presenceIndexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"last_seen", -1}}},
	}

	if err := CreateIndexesSafely(ctx, presenceCollection, presenceIndexes); err != nil {
		return err
	}
	// Note: No TTL index for capped collections - they auto-expire by size

	log.Println("Realtime capped collections created")
	return nil
}

func removeSocialFeatures(ctx context.Context, db *mongo.Database) error {
	log.Println("Removing social features...")

	// Collections to drop
	collectionsToRemove := []string{
		"group_members",
		"event_rsvps",
		"story_views",
		"story_highlights",
		"blocked_users",
		"user_sessions",
		"user_feeds",
		"analytics_events",
		"typing_indicators",
		"user_presence",
	}

	for _, collectionName := range collectionsToRemove {
		if err := db.Collection(collectionName).Drop(ctx); err != nil {
			log.Printf("Warning: Failed to drop collection %s: %v", collectionName, err)
		}
	}

	log.Println("Social features removed")
	return nil
}
