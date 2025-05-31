// migrations/001_initial_indexes.go
package migrations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetInitialIndexesMigration returns the initial indexes migration
func GetInitialIndexesMigration() Migration {
	return Migration{
		ID:          "001_initial_indexes",
		Description: "Create initial database indexes for all collections",
		Up:          createInitialIndexes,
		Down:        dropInitialIndexes,
	}
}

func createInitialIndexes(ctx context.Context, db *mongo.Database) error {
	log.Println("Creating initial database indexes...")

	// Users collection indexes
	if err := createUsersIndexes(ctx, db); err != nil {
		return err
	}

	// Posts collection indexes
	if err := createPostsIndexes(ctx, db); err != nil {
		return err
	}

	// Comments collection indexes
	if err := createCommentsIndexes(ctx, db); err != nil {
		return err
	}

	// Stories collection indexes
	if err := createStoriesIndexes(ctx, db); err != nil {
		return err
	}

	// Messages collection indexes
	if err := createMessagesIndexes(ctx, db); err != nil {
		return err
	}

	// Conversations collection indexes
	if err := createConversationsIndexes(ctx, db); err != nil {
		return err
	}

	// Follows collection indexes
	if err := createFollowsIndexes(ctx, db); err != nil {
		return err
	}

	// Likes collection indexes
	if err := createLikesIndexes(ctx, db); err != nil {
		return err
	}

	// Notifications collection indexes
	if err := createNotificationsIndexes(ctx, db); err != nil {
		return err
	}

	// Groups collection indexes
	if err := createGroupsIndexes(ctx, db); err != nil {
		return err
	}

	// // Events collection indexes
	// if err := createEventsIndexes(ctx, db); err != nil {
	// 	return err
	// }

	// Reports collection indexes
	if err := createReportsIndexes(ctx, db); err != nil {
		return err
	}

	// Hashtags collection indexes
	if err := createHashtagsIndexes(ctx, db); err != nil {
		return err
	}

	// Mentions collection indexes
	if err := createMentionsIndexes(ctx, db); err != nil {
		return err
	}

	// Media collection indexes
	if err := createMediaIndexes(ctx, db); err != nil {
		return err
	}

	log.Println("Initial indexes created successfully")
	return nil
}

func createUsersIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Create unique indexes first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"username", 1}}); err != nil {
		return err
	}

	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"email", 1}}); err != nil {
		return err
	}

	// Regular indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"last_active_at", -1}}},
		{Keys: bson.D{{"is_verified", 1}}},
		{Keys: bson.D{{"role", 1}}},
		{Keys: bson.D{{"is_suspended", 1}}},
		{Keys: bson.D{{"followers_count", -1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Text search index
		{
			Keys: bson.D{
				{"username", "text"},
				{"first_name", "text"},
				{"last_name", "text"},
				{"display_name", "text"},
				{"bio", "text"},
			},
		},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Users indexes created")
	return nil
}

func createPostsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("posts")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"updated_at", -1}}},
		{Keys: bson.D{{"visibility", 1}}},
		{Keys: bson.D{{"is_published", 1}}},
		{Keys: bson.D{{"is_hidden", 1}}},
		{Keys: bson.D{{"is_reported", 1}}},
		{Keys: bson.D{{"type", 1}}},
		{Keys: bson.D{{"content_type", 1}}},
		{Keys: bson.D{{"hashtags", 1}}},
		{Keys: bson.D{{"mentions", 1}}},
		{Keys: bson.D{{"group_id", 1}}, Options: options.Index().SetSparse(true)},
		//{Keys: bson.D{{"event_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"original_post_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"scheduled_for", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"likes_count", -1}}},
		{Keys: bson.D{{"engagement_rate", -1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"visibility", 1}, {"is_published", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"group_id", 1}, {"created_at", -1}}},
		// Text search index
		{Keys: bson.D{{"content", "text"}, {"hashtags", "text"}}},
		// Geospatial index for location
		{Keys: bson.D{{"location", "2dsphere"}}, Options: options.Index().SetSparse(true)},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Posts indexes created")
	return nil
}

func createCommentsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("comments")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"post_id", 1}}},
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"parent_comment_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"root_comment_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"level", 1}}},
		{Keys: bson.D{{"is_hidden", 1}}},
		{Keys: bson.D{{"is_approved", 1}}},
		{Keys: bson.D{{"quality_score", -1}}},
		{Keys: bson.D{{"vote_score", -1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"post_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"post_id", 1}, {"level", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		// Text search index
		{Keys: bson.D{{"content", "text"}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Comments indexes created")
	return nil
}

func createStoriesIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("stories")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"is_expired", 1}}},
		{Keys: bson.D{{"is_highlighted", 1}}},
		{Keys: bson.D{{"highlight_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"visibility", 1}}},
		{Keys: bson.D{{"content_type", 1}}},
		{Keys: bson.D{{"is_hidden", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"visibility", 1}, {"is_expired", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	// Create TTL index for expired stories
	if err := EnsureTTLIndex(ctx, collection, "expires_at", 0); err != nil {
		return err
	}

	log.Println("Stories indexes created")
	return nil
}

func createMessagesIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("messages")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"conversation_id", 1}}},
		{Keys: bson.D{{"sender_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"content_type", 1}}},
		{Keys: bson.D{{"reply_to_message_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"thread_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"is_thread_root", 1}}},
		{Keys: bson.D{{"is_expired", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"conversation_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"sender_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"conversation_id", 1}, {"thread_id", 1}, {"created_at", -1}}},
		// Text search index
		{Keys: bson.D{{"content", "text"}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	// Create TTL index for disappearing messages
	if err := EnsureTTLIndex(ctx, collection, "expires_at", 0); err != nil {
		return err
	}

	log.Println("Messages indexes created")
	return nil
}

func createConversationsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("conversations")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"participants", 1}}},
		{Keys: bson.D{{"created_by", 1}}},
		{Keys: bson.D{{"type", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"last_message_at", -1}}},
		{Keys: bson.D{{"last_activity_at", -1}}},
		{Keys: bson.D{{"is_active", 1}}},
		{Keys: bson.D{{"is_private", 1}}},
		{Keys: bson.D{{"is_public", 1}}},
		{Keys: bson.D{{"admin_ids", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"type", 1}, {"last_activity_at", -1}}},
		{Keys: bson.D{{"participants", 1}, {"last_message_at", -1}}},
		// Text search index
		{Keys: bson.D{{"title", "text"}, {"description", "text"}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Conversations indexes created")
	return nil
}

func createFollowsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("follows")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"follower_id", 1}, {"followee_id", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"follower_id", 1}}},
		{Keys: bson.D{{"followee_id", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"accepted_at", -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"categories", 1}}},
		{Keys: bson.D{{"notifications_enabled", 1}}},
		{Keys: bson.D{{"show_in_feed", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"follower_id", 1}, {"status", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"followee_id", 1}, {"status", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"status", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Follows indexes created")
	return nil
}

func createLikesIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("likes")

	// Create unique compound index first
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"user_id", 1}, {"target_id", 1}, {"target_type", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"target_id", 1}}},
		{Keys: bson.D{{"target_type", 1}}},
		{Keys: bson.D{{"reaction_type", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"target_id", 1}, {"target_type", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"target_id", 1}, {"target_type", 1}, {"reaction_type", 1}}},
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Likes indexes created")
	return nil
}

func createNotificationsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("notifications")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"recipient_id", 1}}},
		{Keys: bson.D{{"actor_id", 1}}},
		{Keys: bson.D{{"type", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"is_read", 1}}},
		{Keys: bson.D{{"is_delivered", 1}}},
		{Keys: bson.D{{"target_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"target_type", 1}}},
		{Keys: bson.D{{"priority", 1}}},
		{Keys: bson.D{{"group_key", 1}}},
		{Keys: bson.D{{"scheduled_at", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"recipient_id", 1}, {"is_read", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"recipient_id", 1}, {"type", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"group_key", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	// Create TTL index for expired notifications
	if err := EnsureTTLIndex(ctx, collection, "expires_at", 0); err != nil {
		return err
	}

	log.Println("Notifications indexes created")
	return nil
}

func createGroupsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("groups")

	// Create unique index for slug
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"slug", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"created_by", 1}}},
		{Keys: bson.D{{"privacy", 1}}},
		{Keys: bson.D{{"category", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"members_count", -1}}},
		{Keys: bson.D{{"activity_score", -1}}},
		{Keys: bson.D{{"is_verified", 1}}},
		{Keys: bson.D{{"is_active", 1}}},
		{Keys: bson.D{{"is_suspended", 1}}},
		{Keys: bson.D{{"tags", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"privacy", 1}, {"is_active", 1}, {"members_count", -1}}},
		{Keys: bson.D{{"category", 1}, {"members_count", -1}}},
		// Text search index
		{Keys: bson.D{{"name", "text"}, {"description", "text"}, {"tags", "text"}}},
		// Geospatial index for location
		{Keys: bson.D{{"location", "2dsphere"}}, Options: options.Index().SetSparse(true)},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Groups indexes created")
	return nil
}

// func createEventsIndexes(ctx context.Context, db *mongo.Database) error {
// 	collection := db.Collection("events")

// 	// Create unique index for slug
// 	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"slug", 1}}); err != nil {
// 		return err
// 	}

// 	indexes := []mongo.IndexModel{
// 		{Keys: bson.D{{"created_by", 1}}},
// 		{Keys: bson.D{{"group_id", 1}}, Options: options.Index().SetSparse(true)},
// 		{Keys: bson.D{{"status", 1}}},
// 		{Keys: bson.D{{"privacy", 1}}},
// 		{Keys: bson.D{{"category", 1}}},
// 		{Keys: bson.D{{"type", 1}}},
// 		{Keys: bson.D{{"start_time", 1}}},
// 		{Keys: bson.D{{"end_time", 1}}},
// 		{Keys: bson.D{{"created_at", -1}}},
// 		{Keys: bson.D{{"attendees_count", -1}}},
// 		{Keys: bson.D{{"is_recurring", 1}}},
// 		{Keys: bson.D{{"is_hidden", 1}}},
// 		{Keys: bson.D{{"tags", 1}}},
// 		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
// 		// Compound indexes
// 		{Keys: bson.D{{"privacy", 1}, {"status", 1}, {"start_time", 1}}},
// 		{Keys: bson.D{{"group_id", 1}, {"start_time", 1}}},
// 		{Keys: bson.D{{"category", 1}, {"start_time", 1}}},
// 		{Keys: bson.D{{"start_time", 1}, {"end_time", 1}}},
// 		// Text search index
// 		{Keys: bson.D{{"title", "text"}, {"description", "text"}, {"tags", "text"}}},
// 		// Geospatial index for location
// 		{Keys: bson.D{{"location", "2dsphere"}}, Options: options.Index().SetSparse(true)},
// 	}

// 	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
// 		return err
// 	}

// 	log.Println("Events indexes created")
// 	return nil
// }

func createReportsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("reports")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"reporter_id", 1}}},
		{Keys: bson.D{{"target_id", 1}}},
		{Keys: bson.D{{"target_type", 1}}},
		{Keys: bson.D{{"reason", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"priority", 1}}},
		{Keys: bson.D{{"assigned_to", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"resolved_by", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"resolved_at", -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"auto_detected", 1}}},
		{Keys: bson.D{{"requires_follow_up", 1}}},
		{Keys: bson.D{{"follow_up_date", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"status", 1}, {"priority", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"target_id", 1}, {"target_type", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"assigned_to", 1}, {"status", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"reason", 1}, {"status", 1}}},
		// Text search index
		{Keys: bson.D{{"description", "text"}, {"resolution_note", "text"}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Reports indexes created")
	return nil
}

func createHashtagsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("hashtags")

	// Create unique index for normalized_tag
	if err := EnsureUniqueIndex(ctx, collection, bson.D{{"normalized_tag", 1}}); err != nil {
		return err
	}

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"tag", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"total_usage", -1}}},
		{Keys: bson.D{{"trending_score", -1}}},
		{Keys: bson.D{{"is_trending", 1}}},
		{Keys: bson.D{{"category", 1}}},
		{Keys: bson.D{{"language", 1}}},
		{Keys: bson.D{{"is_blocked", 1}}},
		{Keys: bson.D{{"is_featured", 1}}},
		{Keys: bson.D{{"first_used_at", -1}}},
		{Keys: bson.D{{"search_count", -1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"is_trending", 1}, {"trending_score", -1}}},
		{Keys: bson.D{{"category", 1}, {"total_usage", -1}}},
		{Keys: bson.D{{"language", 1}, {"total_usage", -1}}},
		// Text search index
		{Keys: bson.D{{"tag", "text"}, {"description", "text"}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Hashtags indexes created")
	return nil
}

func createMentionsIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("mentions")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"mentioner_id", 1}}},
		{Keys: bson.D{{"mentioned_id", 1}}},
		{Keys: bson.D{{"content_id", 1}}},
		{Keys: bson.D{{"content_type", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"is_read", 1}}},
		{Keys: bson.D{{"is_notified", 1}}},
		{Keys: bson.D{{"is_active", 1}}},
		{Keys: bson.D{{"is_visible", 1}}},
		{Keys: bson.D{{"group_id", 1}}, Options: options.Index().SetSparse(true)},
		//{Keys: bson.D{{"event_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"mentioned_id", 1}, {"is_read", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"mentioner_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"content_id", 1}, {"content_type", 1}}},
		{Keys: bson.D{{"mentioned_id", 1}, {"content_type", 1}, {"created_at", -1}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	log.Println("Mentions indexes created")
	return nil
}

func createMediaIndexes(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("media")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"uploaded_by", 1}}},
		{Keys: bson.D{{"type", 1}}},
		{Keys: bson.D{{"category", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"file_size", -1}}},
		{Keys: bson.D{{"mime_type", 1}}},
		{Keys: bson.D{{"is_public", 1}}},
		{Keys: bson.D{{"access_policy", 1}}},
		{Keys: bson.D{{"is_processed", 1}}},
		{Keys: bson.D{{"processing_status", 1}}},
		{Keys: bson.D{{"storage_provider", 1}}},
		{Keys: bson.D{{"related_to", 1}}},
		{Keys: bson.D{{"related_id", 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{"is_expired", 1}}},
		{Keys: bson.D{{"tags", 1}}},
		{Keys: bson.D{{"moderation_status", 1}}},
		{Keys: bson.D{{"deleted_at", 1}}, Options: options.Index().SetSparse(true)},
		// Compound indexes
		{Keys: bson.D{{"uploaded_by", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"type", 1}, {"is_public", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"related_to", 1}, {"related_id", 1}}},
		{Keys: bson.D{{"storage_provider", 1}, {"is_processed", 1}}},
		// Text search index
		{Keys: bson.D{{"original_name", "text"}, {"description", "text"}, {"alt_text", "text"}, {"tags", "text"}}},
	}

	if err := CreateIndexesSafely(ctx, collection, indexes); err != nil {
		return err
	}

	// Create TTL index for expired media
	if err := EnsureTTLIndex(ctx, collection, "expires_at", 0); err != nil {
		return err
	}

	log.Println("Media indexes created")
	return nil
}

func dropInitialIndexes(ctx context.Context, db *mongo.Database) error {
	log.Println("Dropping initial database indexes...")

	collections := []string{
		"users", "posts", "comments", "stories", "messages", "conversations",
		"follows", "likes", "notifications", "groups", //"events",
		"reports",
		"hashtags", "mentions", "media",
	}

	for _, collectionName := range collections {
		collection := db.Collection(collectionName)
		result, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			log.Printf("Warning: Failed to drop indexes for collection %s: %v", collectionName, err)
			continue
		}
		log.Printf("Dropped indexes for collection %s: %v", collectionName, result)
	}

	log.Println("Initial indexes dropped")
	return nil
}
