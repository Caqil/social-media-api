// scripts/seed_data.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Configuration for seeding
type SeedConfig struct {
	UsersCount         int
	PostsPerUser       int
	CommentsPerPost    int
	StoriesPerUser     int
	GroupsCount        int
	EventsCount        int
	FollowsPerUser     int
	HashtagsCount      int
	NotificationsCount int
}

// Default seeding configuration
func DefaultSeedConfig() SeedConfig {
	return SeedConfig{
		UsersCount:         50,
		PostsPerUser:       10,
		CommentsPerPost:    5,
		StoriesPerUser:     3,
		GroupsCount:        10,
		EventsCount:        20,
		FollowsPerUser:     15,
		HashtagsCount:      100,
		NotificationsCount: 200,
	}
}

// SeedData contains the seeded data for reference
type SeedData struct {
	Users    []primitive.ObjectID
	Posts    []primitive.ObjectID
	Comments []primitive.ObjectID
	Stories  []primitive.ObjectID
	Groups   []primitive.ObjectID
	Events   []primitive.ObjectID
	Hashtags []primitive.ObjectID
}

// DataSeeder handles database seeding
type DataSeeder struct {
	db     *mongo.Database
	config SeedConfig
	data   SeedData
}

// NewDataSeeder creates a new data seeder
func NewDataSeeder(db *mongo.Database, config SeedConfig) *DataSeeder {
	return &DataSeeder{
		db:     db,
		config: config,
		data:   SeedData{},
	}
}

// SeedAll seeds all data types
func (ds *DataSeeder) SeedAll(ctx context.Context) error {
	log.Println("Starting database seeding...")

	// Seed in dependency order
	if err := ds.seedUsers(ctx); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	if err := ds.seedHashtags(ctx); err != nil {
		return fmt.Errorf("failed to seed hashtags: %w", err)
	}

	if err := ds.seedPosts(ctx); err != nil {
		return fmt.Errorf("failed to seed posts: %w", err)
	}

	if err := ds.seedComments(ctx); err != nil {
		return fmt.Errorf("failed to seed comments: %w", err)
	}

	if err := ds.seedStories(ctx); err != nil {
		return fmt.Errorf("failed to seed stories: %w", err)
	}

	if err := ds.seedGroups(ctx); err != nil {
		return fmt.Errorf("failed to seed groups: %w", err)
	}

	if err := ds.seedEvents(ctx); err != nil {
		return fmt.Errorf("failed to seed events: %w", err)
	}

	if err := ds.seedFollows(ctx); err != nil {
		return fmt.Errorf("failed to seed follows: %w", err)
	}

	if err := ds.seedLikes(ctx); err != nil {
		return fmt.Errorf("failed to seed likes: %w", err)
	}

	if err := ds.seedMentions(ctx); err != nil {
		return fmt.Errorf("failed to seed mentions: %w", err)
	}

	if err := ds.seedNotifications(ctx); err != nil {
		return fmt.Errorf("failed to seed notifications: %w", err)
	}

	if err := ds.seedConversationsAndMessages(ctx); err != nil {
		return fmt.Errorf("failed to seed conversations and messages: %w", err)
	}

	log.Println("Database seeding completed successfully!")
	ds.printSeedingSummary()
	return nil
}

func (ds *DataSeeder) seedUsers(ctx context.Context) error {
	log.Println("Seeding users...")
	collection := ds.db.Collection("users")

	users := make([]interface{}, ds.config.UsersCount)
	userIDs := make([]primitive.ObjectID, ds.config.UsersCount)

	for i := 0; i < ds.config.UsersCount; i++ {
		userID := primitive.NewObjectID()
		userIDs[i] = userID

		now := time.Now()
		user := map[string]interface{}{
			"_id":             userID,
			"username":        fmt.Sprintf("user%d", i+1),
			"email":           fmt.Sprintf("user%d@example.com", i+1),
			"password":        "$2a$10$hashedpasswordexample", // In real app, properly hash passwords
			"first_name":      fmt.Sprintf("FirstName%d", i+1),
			"last_name":       fmt.Sprintf("LastName%d", i+1),
			"display_name":    fmt.Sprintf("User %d", i+1),
			"bio":             fmt.Sprintf("This is the bio for user %d. Passionate about technology and social media.", i+1),
			"profile_pic":     fmt.Sprintf("https://via.placeholder.com/150?text=User%d", i+1),
			"cover_pic":       fmt.Sprintf("https://via.placeholder.com/800x200?text=Cover%d", i+1),
			"is_verified":     i < 5, // First 5 users are verified
			"is_active":       true,
			"is_private":      i%3 == 0, // Every third user is private
			"is_suspended":    false,
			"role":            getUserRole(i),
			"followers_count": rand.Intn(1000),
			"following_count": rand.Intn(500),
			"posts_count":     ds.config.PostsPerUser,
			"friends_count":   rand.Intn(100),
			"language":        "en",
			"timezone":        "UTC",
			"theme":           "light",
			"online_status":   getRandomOnlineStatus(),
			"last_active_at":  getRandomRecentTime(),
			"created_at":      now,
			"updated_at":      now,
			"privacy_settings": map[string]interface{}{
				"profile_visibility":    "public",
				"posts_visibility":      getRandomPrivacyLevel(),
				"followers_visibility":  "public",
				"following_visibility":  "public",
				"email_visibility":      "private",
				"phone_visibility":      "private",
				"allow_messages":        true,
				"allow_tagging":         true,
				"allow_follow_requests": true,
				"show_online_status":    true,
				"allow_story_views":     true,
			},
			"notification_settings": map[string]interface{}{
				"email_notifications":   true,
				"push_notifications":    true,
				"sms_notifications":     false,
				"like_notifications":    true,
				"comment_notifications": true,
				"follow_notifications":  true,
				"message_notifications": true,
				"mention_notifications": true,
				"group_notifications":   true,
				"event_notifications":   true,
			},
		}

		users[i] = user
	}

	if _, err := collection.InsertMany(ctx, users); err != nil {
		return err
	}

	ds.data.Users = userIDs
	log.Printf("Seeded %d users", len(users))
	return nil
}

func (ds *DataSeeder) seedHashtags(ctx context.Context) error {
	log.Println("Seeding hashtags...")
	collection := ds.db.Collection("hashtags")

	popularHashtags := []string{
		"technology", "programming", "design", "photography", "travel",
		"food", "fitness", "music", "art", "nature", "business", "startup",
		"lifestyle", "fashion", "health", "education", "gaming", "sports",
		"science", "books", "movies", "tv", "culture", "news", "politics",
		"climate", "sustainability", "innovation", "ai", "blockchain",
		"webdev", "mobile", "ux", "ui", "marketing", "socialmedia",
		"productivity", "motivation", "inspiration", "success", "goals",
		"mindfulness", "meditation", "yoga", "cooking", "recipe",
		"adventure", "hiking", "beach", "sunset", "coffee", "weekend",
	}

	hashtags := make([]interface{}, 0)
	hashtagIDs := make([]primitive.ObjectID, 0)

	for i := 0; i < ds.config.HashtagsCount && i < len(popularHashtags)*2; i++ {
		hashtagID := primitive.NewObjectID()
		hashtagIDs = append(hashtagIDs, hashtagID)

		tag := popularHashtags[i%len(popularHashtags)]
		if i >= len(popularHashtags) {
			tag = fmt.Sprintf("%s%d", tag, i-len(popularHashtags)+1)
		}

		now := time.Now()
		hashtag := map[string]interface{}{
			"_id":            hashtagID,
			"tag":            tag,
			"normalized_tag": strings.ToLower(tag),
			"display_tag":    tag,
			"posts_count":    rand.Intn(500),
			"stories_count":  rand.Intn(100),
			"total_usage":    rand.Intn(600),
			"is_trending":    i < 10, // First 10 are trending
			"trending_score": float64(rand.Intn(100)),
			"category":       getRandomHashtagCategory(),
			"language":       "en",
			"is_blocked":     false,
			"is_featured":    i < 5, // First 5 are featured
			"search_count":   rand.Intn(1000),
			"click_count":    rand.Intn(500),
			"first_used_at":  getRandomPastTime(),
			"created_at":     now,
			"updated_at":     now,
		}

		hashtags = append(hashtags, hashtag)
	}

	if len(hashtags) > 0 {
		if _, err := collection.InsertMany(ctx, hashtags); err != nil {
			return err
		}
	}

	ds.data.Hashtags = hashtagIDs
	log.Printf("Seeded %d hashtags", len(hashtags))
	return nil
}

func (ds *DataSeeder) seedPosts(ctx context.Context) error {
	log.Println("Seeding posts...")
	collection := ds.db.Collection("posts")

	posts := make([]interface{}, 0)
	postIDs := make([]primitive.ObjectID, 0)

	postContents := []string{
		"Just had an amazing day exploring the city! 🌆",
		"Working on some exciting new projects. Can't wait to share! 💻",
		"Beautiful sunset today. Nature never fails to amaze me 🌅",
		"Coffee and code - the perfect combination ☕️",
		"Grateful for another day of learning and growing 🌱",
		"Weekend vibes are the best! Time to relax and recharge 😌",
		"Trying out a new recipe today. Wish me luck! 👨‍🍳",
		"The power of community is incredible. Thank you all! 🙏",
		"Sometimes the best moments are the quiet ones 📚",
		"Innovation happens when we dare to think differently 💡",
	}

	for userIndex, userID := range ds.data.Users {
		for i := 0; i < ds.config.PostsPerUser; i++ {
			postID := primitive.NewObjectID()
			postIDs = append(postIDs, postID)

			now := time.Now()
			post := map[string]interface{}{
				"_id":              postID,
				"user_id":          userID,
				"content":          postContents[rand.Intn(len(postContents))],
				"content_type":     getRandomContentType(),
				"type":             "post",
				"visibility":       getRandomPrivacyLevel(),
				"language":         "en",
				"hashtags":         getRandomHashtags(ds.data.Hashtags, 3),
				"mentions":         getRandomMentions(ds.data.Users, userIndex, 2),
				"is_edited":        false,
				"comments_enabled": true,
				"likes_enabled":    true,
				"shares_enabled":   true,
				"is_pinned":        i == 0 && userIndex < 5, // Pin first post for first 5 users
				"is_promoted":      false,
				"is_repost":        rand.Float32() < 0.1, // 10% chance of being a repost
				"is_reported":      false,
				"is_hidden":        false,
				"is_approved":      true,
				"is_scheduled":     false,
				"is_published":     true,
				"published_at":     getRandomPastTime(),
				"likes_count":      rand.Intn(100),
				"comments_count":   ds.config.CommentsPerPost,
				"shares_count":     rand.Intn(20),
				"views_count":      rand.Intn(1000),
				"saves_count":      rand.Intn(50),
				"engagement_rate":  rand.Float64() * 10,
				"reach_count":      rand.Intn(500),
				"impression_count": rand.Intn(2000),
				"created_at":       getRandomPastTime(),
				"updated_at":       now,
			}

			// Add media for some posts
			if rand.Float32() < 0.3 { // 30% chance of having media
				post["media"] = []map[string]interface{}{
					{
						"url":       fmt.Sprintf("https://picsum.photos/800/600?random=%d", rand.Intn(1000)),
						"type":      "image",
						"size":      rand.Int63n(1024*1024) + 100000, // 100KB to 1MB
						"width":     800,
						"height":    600,
						"thumbnail": fmt.Sprintf("https://picsum.photos/200/150?random=%d", rand.Intn(1000)),
					},
				}
			}

			posts = append(posts, post)
		}
	}

	if len(posts) > 0 {
		if _, err := collection.InsertMany(ctx, posts); err != nil {
			return err
		}
	}

	ds.data.Posts = postIDs
	log.Printf("Seeded %d posts", len(posts))
	return nil
}

func (ds *DataSeeder) seedComments(ctx context.Context) error {
	log.Println("Seeding comments...")
	collection := ds.db.Collection("comments")

	comments := make([]interface{}, 0)
	commentIDs := make([]primitive.ObjectID, 0)

	commentTexts := []string{
		"Great post! Thanks for sharing.",
		"This is really insightful!",
		"I completely agree with this.",
		"Thanks for the inspiration!",
		"Love this! 😍",
		"So true! Well said.",
		"This made my day!",
		"Absolutely amazing!",
		"Can't wait to try this out!",
		"You're so right about this.",
	}

	for _, postID := range ds.data.Posts {
		for i := 0; i < ds.config.CommentsPerPost; i++ {
			commentID := primitive.NewObjectID()
			commentIDs = append(commentIDs, commentID)

			now := time.Now()
			comment := map[string]interface{}{
				"_id":               commentID,
				"post_id":           postID,
				"user_id":           ds.data.Users[rand.Intn(len(ds.data.Users))],
				"content":           commentTexts[rand.Intn(len(commentTexts))],
				"content_type":      "text",
				"level":             0, // All top-level for simplicity
				"likes_count":       rand.Intn(20),
				"replies_count":     0,
				"is_edited":         false,
				"is_pinned":         false,
				"is_highlighted":    i == 0 && rand.Float32() < 0.1, // 10% chance first comment is highlighted
				"is_reported":       false,
				"is_hidden":         false,
				"is_approved":       true,
				"upvotes_count":     rand.Intn(15),
				"downvotes_count":   rand.Intn(3),
				"vote_score":        rand.Intn(15) - rand.Intn(3),
				"quality_score":     rand.Float64() * 10,
				"is_verified_reply": false,
				"is_author_reply":   false,
				"awards_count":      rand.Intn(3),
				"created_at":        getRandomRecentTime(),
				"updated_at":        now,
			}

			comments = append(comments, comment)
		}
	}

	if len(comments) > 0 {
		if _, err := collection.InsertMany(ctx, comments); err != nil {
			return err
		}
	}

	ds.data.Comments = commentIDs
	log.Printf("Seeded %d comments", len(comments))
	return nil
}

func (ds *DataSeeder) seedStories(ctx context.Context) error {
	log.Println("Seeding stories...")
	collection := ds.db.Collection("stories")

	stories := make([]interface{}, 0)
	storyIDs := make([]primitive.ObjectID, 0)

	for _, userID := range ds.data.Users {
		for i := 0; i < ds.config.StoriesPerUser; i++ {
			storyID := primitive.NewObjectID()
			storyIDs = append(storyIDs, storyID)

			now := time.Now()
			expiresAt := now.Add(24 * time.Hour) // Stories expire in 24 hours

			story := map[string]interface{}{
				"_id":          storyID,
				"user_id":      userID,
				"content":      fmt.Sprintf("Story content %d", i+1),
				"content_type": getRandomStoryContentType(),
				"media": map[string]interface{}{
					"url":       fmt.Sprintf("https://picsum.photos/400/800?random=%d", rand.Intn(1000)),
					"type":      "image",
					"size":      rand.Int63n(512*1024) + 50000, // 50KB to 512KB
					"width":     400,
					"height":    800,
					"thumbnail": fmt.Sprintf("https://picsum.photos/200/400?random=%d", rand.Intn(1000)),
				},
				"duration":           15,
				"expires_at":         expiresAt,
				"is_expired":         false,
				"visibility":         getRandomPrivacyLevel(),
				"views_count":        rand.Intn(50),
				"likes_count":        rand.Intn(20),
				"replies_count":      rand.Intn(5),
				"shares_count":       rand.Intn(3),
				"allow_replies":      true,
				"allow_reactions":    true,
				"allow_sharing":      true,
				"allow_screenshot":   true,
				"is_highlighted":     rand.Float32() < 0.2, // 20% chance of being highlighted
				"unique_views_count": rand.Intn(40),
				"completion_rate":    rand.Float64() * 100,
				"engagement_rate":    rand.Float64() * 5,
				"is_reported":        false,
				"is_hidden":          false,
				"created_at":         getRandomRecentTime(),
				"updated_at":         now,
			}

			stories = append(stories, story)
		}
	}

	if len(stories) > 0 {
		if _, err := collection.InsertMany(ctx, stories); err != nil {
			return err
		}
	}

	ds.data.Stories = storyIDs
	log.Printf("Seeded %d stories", len(stories))
	return nil
}

func (ds *DataSeeder) seedGroups(ctx context.Context) error {
	log.Println("Seeding groups...")
	collection := ds.db.Collection("groups")

	groups := make([]interface{}, ds.config.GroupsCount)
	groupIDs := make([]primitive.ObjectID, ds.config.GroupsCount)

	groupNames := []string{
		"Tech Enthusiasts", "Photography Club", "Fitness Motivation",
		"Book Lovers", "Food Adventures", "Travel Buddies",
		"Startup Founders", "Design Inspiration", "Music Lovers",
		"Gaming Community", "Art & Creativity", "Local Events",
	}

	for i := 0; i < ds.config.GroupsCount; i++ {
		groupID := primitive.NewObjectID()
		groupIDs[i] = groupID

		name := groupNames[i%len(groupNames)]
		if i >= len(groupNames) {
			name = fmt.Sprintf("%s %d", name, i/len(groupNames)+1)
		}

		now := time.Now()
		group := map[string]interface{}{
			"_id":                      groupID,
			"name":                     name,
			"slug":                     fmt.Sprintf("%s-%d", strings.ToLower(strings.ReplaceAll(name, " ", "-")), i+1),
			"description":              fmt.Sprintf("Welcome to %s! This is a community for people who are passionate about our shared interests.", name),
			"profile_pic":              fmt.Sprintf("https://via.placeholder.com/150?text=%s", strings.ReplaceAll(name, " ", "")),
			"cover_pic":                fmt.Sprintf("https://via.placeholder.com/800x200?text=%s", strings.ReplaceAll(name, " ", "")),
			"privacy":                  getRandomGroupPrivacy(),
			"category":                 getRandomGroupCategory(),
			"tags":                     getRandomTags(3),
			"created_by":               ds.data.Users[rand.Intn(len(ds.data.Users))],
			"members_count":            rand.Intn(1000) + 10,
			"admins_count":             rand.Intn(5) + 1,
			"mods_count":               rand.Intn(3),
			"posts_count":              rand.Intn(500),
			"events_count":             rand.Intn(20),
			"post_approval_required":   rand.Float32() < 0.3, // 30% require approval
			"member_approval_required": rand.Float32() < 0.4, // 40% require approval
			"allow_member_invites":     true,
			"allow_external_sharing":   true,
			"allow_polls":              true,
			"allow_events":             true,
			"allow_discussions":        true,
			"last_activity_at":         getRandomRecentTime(),
			"last_post_at":             getRandomRecentTime(),
			"weekly_growth_rate":       rand.Float64() * 10,
			"engagement_score":         rand.Float64() * 100,
			"activity_score":           rand.Float64() * 100,
			"is_verified":              i < 3, // First 3 groups are verified
			"is_active":                true,
			"is_suspended":             false,
			"is_premium":               i < 2, // First 2 groups are premium
			"created_at":               getRandomPastTime(),
			"updated_at":               now,
		}

		groups[i] = group
	}

	if _, err := collection.InsertMany(ctx, groups); err != nil {
		return err
	}

	ds.data.Groups = groupIDs
	log.Printf("Seeded %d groups", len(groups))
	return nil
}

func (ds *DataSeeder) seedEvents(ctx context.Context) error {
	log.Println("Seeding events...")
	collection := ds.db.Collection("events")

	events := make([]interface{}, ds.config.EventsCount)
	eventIDs := make([]primitive.ObjectID, ds.config.EventsCount)

	eventTitles := []string{
		"Tech Meetup", "Photography Workshop", "Fitness Bootcamp",
		"Book Club Meeting", "Food Festival", "Travel Planning Session",
		"Startup Pitch Night", "Design Conference", "Music Concert",
		"Gaming Tournament", "Art Exhibition", "Community Cleanup",
	}

	for i := 0; i < ds.config.EventsCount; i++ {
		eventID := primitive.NewObjectID()
		eventIDs[i] = eventID

		title := eventTitles[i%len(eventTitles)]
		if i >= len(eventTitles) {
			title = fmt.Sprintf("%s %d", title, i/len(eventTitles)+1)
		}

		startTime := getRandomFutureTime()
		endTime := startTime.Add(time.Hour * time.Duration(rand.Intn(6)+1)) // 1-6 hours duration

		now := time.Now()
		event := map[string]interface{}{
			"_id":                 eventID,
			"title":               title,
			"description":         fmt.Sprintf("Join us for an amazing %s! This will be a great opportunity to connect with like-minded people.", title),
			"slug":                fmt.Sprintf("%s-%d", strings.ToLower(strings.ReplaceAll(title, " ", "-")), i+1),
			"cover_image":         fmt.Sprintf("https://via.placeholder.com/800x400?text=%s", strings.ReplaceAll(title, " ", "")),
			"category":            getRandomEventCategory(),
			"tags":                getRandomTags(3),
			"type":                getRandomEventType(),
			"start_time":          startTime,
			"end_time":            endTime,
			"timezone":            "UTC",
			"is_all_day":          rand.Float32() < 0.1, // 10% are all-day events
			"is_recurring":        rand.Float32() < 0.2, // 20% are recurring
			"created_by":          ds.data.Users[rand.Intn(len(ds.data.Users))],
			"group_id":            getRandomGroupID(ds.data.Groups),
			"status":              "published",
			"privacy":             getRandomPrivacyLevel(),
			"max_attendees":       rand.Intn(500) + 10,
			"require_approval":    rand.Float32() < 0.3, // 30% require approval
			"allow_guest_invites": true,
			"allow_comments":      true,
			"allow_photos":        true,
			"attendees_count":     rand.Intn(100),
			"interested_count":    rand.Intn(50),
			"going_count":         rand.Intn(80),
			"maybe_count":         rand.Intn(30),
			"not_going_count":     rand.Intn(10),
			"invited_count":       rand.Intn(200),
			"views_count":         rand.Intn(1000),
			"shares_count":        rand.Intn(20),
			"comments_count":      rand.Intn(50),
			"is_hidden":           false,
			"created_at":          getRandomPastTime(),
			"updated_at":          now,
		}

		// Add location for offline/hybrid events
		if event["type"] != "online" {
			event["location"] = map[string]interface{}{
				"name":      fmt.Sprintf("Venue for %s", title),
				"address":   fmt.Sprintf("%d Main St, City, State", rand.Intn(9999)+1),
				"latitude":  rand.Float64()*180 - 90,  // -90 to 90
				"longitude": rand.Float64()*360 - 180, // -180 to 180
			}
		}

		events[i] = event
	}

	if _, err := collection.InsertMany(ctx, events); err != nil {
		return err
	}

	ds.data.Events = eventIDs
	log.Printf("Seeded %d events", len(events))
	return nil
}

func (ds *DataSeeder) seedFollows(ctx context.Context) error {
	log.Println("Seeding follows...")
	collection := ds.db.Collection("follows")

	follows := make([]interface{}, 0)
	followMap := make(map[string]bool) // To avoid duplicates

	for _, followerID := range ds.data.Users {
		// Each user follows a random number of other users
		followCount := rand.Intn(ds.config.FollowsPerUser) + 1

		for i := 0; i < followCount; i++ {
			followeeID := ds.data.Users[rand.Intn(len(ds.data.Users))]

			// Don't follow yourself and avoid duplicates
			if followerID == followeeID {
				continue
			}

			key := followerID.Hex() + ":" + followeeID.Hex()
			if followMap[key] {
				continue
			}
			followMap[key] = true

			now := time.Now()
			requestedAt := getRandomPastTime()

			follow := map[string]interface{}{
				"_id":                   primitive.NewObjectID(),
				"follower_id":           followerID,
				"followee_id":           followeeID,
				"status":                getRandomFollowStatus(),
				"requested_at":          requestedAt,
				"notifications_enabled": true,
				"show_in_feed":          true,
				"categories":            getRandomFollowCategories(),
				"interaction_score":     rand.Float64() * 100,
				"last_interaction_at":   getRandomRecentTime(),
				"source":                "web",
				"created_at":            requestedAt,
				"updated_at":            now,
			}

			// Set accepted_at for accepted follows
			if follow["status"] == "accepted" {
				follow["accepted_at"] = requestedAt.Add(time.Hour * time.Duration(rand.Intn(48)))
			}

			follows = append(follows, follow)
		}
	}

	if len(follows) > 0 {
		if _, err := collection.InsertMany(ctx, follows); err != nil {
			return err
		}
	}

	log.Printf("Seeded %d follows", len(follows))
	return nil
}

func (ds *DataSeeder) seedLikes(ctx context.Context) error {
	log.Println("Seeding likes...")
	collection := ds.db.Collection("likes")

	likes := make([]interface{}, 0)
	likeMap := make(map[string]bool) // To avoid duplicates

	// Like posts
	for _, postID := range ds.data.Posts {
		likeCount := rand.Intn(20) + 1 // 1-20 likes per post

		for i := 0; i < likeCount; i++ {
			userID := ds.data.Users[rand.Intn(len(ds.data.Users))]
			key := userID.Hex() + ":post:" + postID.Hex()

			if likeMap[key] {
				continue
			}
			likeMap[key] = true

			like := map[string]interface{}{
				"_id":           primitive.NewObjectID(),
				"user_id":       userID,
				"target_id":     postID,
				"target_type":   "post",
				"reaction_type": getRandomReactionType(),
				"source":        "web",
				"created_at":    getRandomRecentTime(),
			}

			likes = append(likes, like)
		}
	}

	// Like comments
	for _, commentID := range ds.data.Comments {
		if rand.Float32() < 0.5 { // 50% chance of getting likes
			likeCount := rand.Intn(5) + 1 // 1-5 likes per comment

			for i := 0; i < likeCount; i++ {
				userID := ds.data.Users[rand.Intn(len(ds.data.Users))]
				key := userID.Hex() + ":comment:" + commentID.Hex()

				if likeMap[key] {
					continue
				}
				likeMap[key] = true

				like := map[string]interface{}{
					"_id":           primitive.NewObjectID(),
					"user_id":       userID,
					"target_id":     commentID,
					"target_type":   "comment",
					"reaction_type": getRandomReactionType(),
					"source":        "web",
					"created_at":    getRandomRecentTime(),
				}

				likes = append(likes, like)
			}
		}
	}

	if len(likes) > 0 {
		if _, err := collection.InsertMany(ctx, likes); err != nil {
			return err
		}
	}

	log.Printf("Seeded %d likes", len(likes))
	return nil
}

func (ds *DataSeeder) seedMentions(ctx context.Context) error {
	log.Println("Seeding mentions...")
	collection := ds.db.Collection("mentions")

	mentions := make([]interface{}, 0)

	// Create mentions for posts
	for _, postID := range ds.data.Posts {
		if rand.Float32() < 0.3 { // 30% chance of having mentions
			mentionCount := rand.Intn(3) + 1 // 1-3 mentions per post

			for i := 0; i < mentionCount; i++ {
				mentionerID := ds.data.Users[rand.Intn(len(ds.data.Users))]
				mentionedID := ds.data.Users[rand.Intn(len(ds.data.Users))]

				if mentionerID == mentionedID {
					continue
				}

				now := time.Now()
				mention := map[string]interface{}{
					"_id":          primitive.NewObjectID(),
					"mentioner_id": mentionerID,
					"mentioned_id": mentionedID,
					"content_type": "post",
					"content_id":   postID,
					"mention_text": fmt.Sprintf("@user%d", rand.Intn(ds.config.UsersCount)+1),
					"is_active":    true,
					"is_notified":  rand.Float32() < 0.8, // 80% are notified
					"is_read":      rand.Float32() < 0.6, // 60% are read
					"is_visible":   true,
					"is_blocked":   false,
					"click_count":  rand.Intn(10),
					"view_count":   rand.Intn(50),
					"source":       "web",
					"created_at":   getRandomRecentTime(),
					"updated_at":   now,
				}

				if mention["is_notified"].(bool) {
					mention["notified_at"] = getRandomRecentTime()
				}

				if mention["is_read"].(bool) {
					mention["read_at"] = getRandomRecentTime()
				}

				mentions = append(mentions, mention)
			}
		}
	}

	if len(mentions) > 0 {
		if _, err := collection.InsertMany(ctx, mentions); err != nil {
			return err
		}
	}

	log.Printf("Seeded %d mentions", len(mentions))
	return nil
}

func (ds *DataSeeder) seedNotifications(ctx context.Context) error {
	log.Println("Seeding notifications...")
	collection := ds.db.Collection("notifications")

	notifications := make([]interface{}, ds.config.NotificationsCount)

	notificationTypes := []string{
		"like", "comment", "follow", "mention", "message",
		"group_invite", "event_invite", "post_share",
	}

	for i := 0; i < ds.config.NotificationsCount; i++ {
		recipientID := ds.data.Users[rand.Intn(len(ds.data.Users))]
		actorID := ds.data.Users[rand.Intn(len(ds.data.Users))]

		if recipientID == actorID {
			actorID = ds.data.Users[(rand.Intn(len(ds.data.Users))+1)%len(ds.data.Users)]
		}

		notifType := notificationTypes[rand.Intn(len(notificationTypes))]
		title, message := getNotificationContent(notifType)

		now := time.Now()
		createdAt := getRandomRecentTime()

		notification := map[string]interface{}{
			"_id":            primitive.NewObjectID(),
			"recipient_id":   recipientID,
			"actor_id":       actorID,
			"type":           notifType,
			"title":          title,
			"message":        message,
			"action_text":    "View",
			"target_type":    getTargetType(notifType),
			"is_read":        rand.Float32() < 0.4, // 40% are read
			"is_delivered":   rand.Float32() < 0.9, // 90% are delivered
			"sent_via_email": rand.Float32() < 0.6, // 60% sent via email
			"sent_via_push":  rand.Float32() < 0.8, // 80% sent via push
			"sent_via_sms":   rand.Float32() < 0.1, // 10% sent via SMS
			"group_count":    1,
			"is_grouped":     false,
			"priority":       getRandomPriority(),
			"created_at":     createdAt,
			"updated_at":     now,
		}

		if rand.Float32() < 0.8 { // 80% have target ID
			notification["target_id"] = getRandomTargetID(ds.data, notifType)
		}

		if notification["is_read"].(bool) {
			notification["read_at"] = createdAt.Add(time.Hour * time.Duration(rand.Intn(48)))
		}

		if notification["is_delivered"].(bool) {
			notification["delivered_at"] = createdAt.Add(time.Minute * time.Duration(rand.Intn(60)))
		}

		notifications[i] = notification
	}

	if _, err := collection.InsertMany(ctx, notifications); err != nil {
		return err
	}

	log.Printf("Seeded %d notifications", len(notifications))
	return nil
}

func (ds *DataSeeder) seedConversationsAndMessages(ctx context.Context) error {
	log.Println("Seeding conversations and messages...")

	// Create conversations
	conversationsCollection := ds.db.Collection("conversations")
	messagesCollection := ds.db.Collection("messages")

	conversations := make([]interface{}, 0)
	conversationIDs := make([]primitive.ObjectID, 0)

	// Create direct conversations
	conversationCount := len(ds.data.Users) / 2 // Half as many conversations as users

	for i := 0; i < conversationCount; i++ {
		conversationID := primitive.NewObjectID()
		conversationIDs = append(conversationIDs, conversationID)

		participant1 := ds.data.Users[rand.Intn(len(ds.data.Users))]
		participant2 := ds.data.Users[rand.Intn(len(ds.data.Users))]

		if participant1 == participant2 {
			participant2 = ds.data.Users[(rand.Intn(len(ds.data.Users))+1)%len(ds.data.Users)]
		}

		now := time.Now()
		createdAt := getRandomPastTime()
		lastMessageAt := getRandomRecentTime()

		conversation := map[string]interface{}{
			"_id":                  conversationID,
			"type":                 "direct",
			"participants":         []primitive.ObjectID{participant1, participant2},
			"created_by":           participant1,
			"last_message_at":      lastMessageAt,
			"last_message_preview": "Hey, how are you doing?",
			"last_activity_at":     lastMessageAt,
			"is_archived":          false,
			"is_muted":             false,
			"is_locked":            false,
			"is_private":           true,
			"allow_invites":        false,
			"allow_media_sharing":  true,
			"messages_count":       rand.Intn(20) + 5, // 5-24 messages
			"active_members_count": 2,
			"is_encrypted":         false,
			"has_pinned_messages":  false,
			"is_active":            true,
			"created_at":           createdAt,
			"updated_at":           now,
		}

		conversations = append(conversations, conversation)
	}

	if len(conversations) > 0 {
		if _, err := conversationsCollection.InsertMany(ctx, conversations); err != nil {
			return err
		}
	}

	// Create messages for conversations
	messages := make([]interface{}, 0)
	messageTexts := []string{
		"Hey, how are you doing?",
		"Good morning!",
		"Thanks for sharing that!",
		"Let's catch up soon.",
		"That sounds great!",
		"I'll talk to you later.",
		"Have a great day!",
		"What's up?",
		"How was your weekend?",
		"See you soon!",
	}

	for _, conversationID := range conversationIDs {
		messageCount := rand.Intn(15) + 5 // 5-19 messages per conversation

		for j := 0; j < messageCount; j++ {
			messageID := primitive.NewObjectID()
			senderID := ds.data.Users[rand.Intn(len(ds.data.Users))]

			now := time.Now()
			createdAt := getRandomRecentTime()

			message := map[string]interface{}{
				"_id":             messageID,
				"conversation_id": conversationID,
				"sender_id":       senderID,
				"content":         messageTexts[rand.Intn(len(messageTexts))],
				"content_type":    "text",
				"status":          getRandomMessageStatus(),
				"sent_at":         createdAt,
				"is_edited":       false,
				"is_forwarded":    false,
				"is_thread_root":  false,
				"thread_count":    0,
				"is_expired":      false,
				"priority":        "normal",
				"created_at":      createdAt,
				"updated_at":      now,
			}

			if message["status"] == "delivered" || message["status"] == "read" {
				message["delivered_at"] = createdAt.Add(time.Second * time.Duration(rand.Intn(300)))
			}

			if message["status"] == "read" {
				message["read_at"] = createdAt.Add(time.Minute * time.Duration(rand.Intn(60)))
			}

			messages = append(messages, message)
		}
	}

	if len(messages) > 0 {
		if _, err := messagesCollection.InsertMany(ctx, messages); err != nil {
			return err
		}
	}

	log.Printf("Seeded %d conversations and %d messages", len(conversations), len(messages))
	return nil
}

func (ds *DataSeeder) printSeedingSummary() {
	log.Println("\n=== SEEDING SUMMARY ===")
	log.Printf("Users: %d", len(ds.data.Users))
	log.Printf("Posts: %d", len(ds.data.Posts))
	log.Printf("Comments: %d", len(ds.data.Comments))
	log.Printf("Stories: %d", len(ds.data.Stories))
	log.Printf("Groups: %d", len(ds.data.Groups))
	log.Printf("Events: %d", len(ds.data.Events))
	log.Printf("Hashtags: %d", len(ds.data.Hashtags))
	log.Println("======================")
}

// Helper functions

func getUserRole(index int) string {
	if index == 0 {
		return "super_admin"
	} else if index < 3 {
		return "admin"
	} else if index < 8 {
		return "moderator"
	}
	return "user"
}

func getRandomOnlineStatus() string {
	statuses := []string{"online", "offline", "away"}
	return statuses[rand.Intn(len(statuses))]
}

func getRandomPrivacyLevel() string {
	levels := []string{"public", "friends", "private"}
	return levels[rand.Intn(len(levels))]
}

func getRandomContentType() string {
	types := []string{"text", "image", "video", "link"}
	return types[rand.Intn(len(types))]
}

func getRandomStoryContentType() string {
	types := []string{"image", "video"}
	return types[rand.Intn(len(types))]
}

func getRandomHashtagCategory() string {
	categories := []string{
		"general", "entertainment", "sports", "news", "technology",
		"business", "lifestyle", "travel", "food", "fashion",
	}
	return categories[rand.Intn(len(categories))]
}

func getRandomGroupPrivacy() string {
	privacies := []string{"public", "private", "secret"}
	return privacies[rand.Intn(len(privacies))]
}

func getRandomGroupCategory() string {
	categories := []string{
		"general", "technology", "business", "health", "education",
		"entertainment", "sports", "travel", "food", "art",
	}
	return categories[rand.Intn(len(categories))]
}

func getRandomEventCategory() string {
	categories := []string{
		"business", "technology", "education", "health", "arts",
		"music", "sports", "entertainment", "food", "travel",
	}
	return categories[rand.Intn(len(categories))]
}

func getRandomEventType() string {
	types := []string{"online", "offline", "hybrid"}
	return types[rand.Intn(len(types))]
}

func getRandomFollowStatus() string {
	weights := []int{20, 70, 5, 5} // 70% accepted, 20% pending, 5% blocked, 5% muted

	total := 0
	for _, w := range weights {
		total += w
	}

	r := rand.Intn(total)
	cumulative := 0

	statuses_list := []string{"pending", "accepted", "blocked", "muted"}
	for i, w := range weights {
		cumulative += w
		if r < cumulative {
			return statuses_list[i]
		}
	}

	return "accepted"
}

func getRandomReactionType() string {
	reactions := []string{"like", "love", "haha", "wow", "sad", "angry", "support"}
	return reactions[rand.Intn(len(reactions))]
}

func getRandomMessageStatus() string {
	statuses := []string{"sent", "delivered", "read"}
	weights := []int{10, 30, 60} // 60% read, 30% delivered, 10% sent

	total := 0
	for _, w := range weights {
		total += w
	}

	r := rand.Intn(total)
	cumulative := 0

	for i, w := range weights {
		cumulative += w
		if r < cumulative {
			return statuses[i]
		}
	}

	return "read"
}

func getRandomPriority() string {
	priorities := []string{"low", "medium", "high"}
	weights := []int{20, 70, 10} // 70% medium, 20% low, 10% high

	total := 0
	for _, w := range weights {
		total += w
	}

	r := rand.Intn(total)
	cumulative := 0

	for i, w := range weights {
		cumulative += w
		if r < cumulative {
			return priorities[i]
		}
	}

	return "medium"
}

func getRandomHashtags(allHashtags []primitive.ObjectID, maxCount int) []string {
	if len(allHashtags) == 0 {
		return []string{}
	}

	count := rand.Intn(maxCount) + 1
	if count > len(allHashtags) {
		count = len(allHashtags)
	}

	hashtags := make([]string, count)
	for i := 0; i < count; i++ {
		hashtags[i] = fmt.Sprintf("hashtag%d", rand.Intn(len(allHashtags))+1)
	}

	return hashtags
}

func getRandomMentions(allUsers []primitive.ObjectID, excludeIndex int, maxCount int) []primitive.ObjectID {
	if len(allUsers) <= 1 {
		return []primitive.ObjectID{}
	}

	count := rand.Intn(maxCount + 1) // 0 to maxCount
	if count == 0 {
		return []primitive.ObjectID{}
	}

	mentions := make([]primitive.ObjectID, 0, count)
	used := make(map[int]bool)
	used[excludeIndex] = true

	for len(mentions) < count && len(used) < len(allUsers) {
		index := rand.Intn(len(allUsers))
		if !used[index] {
			mentions = append(mentions, allUsers[index])
			used[index] = true
		}
	}

	return mentions
}

func getRandomTags(count int) []string {
	allTags := []string{
		"trending", "popular", "new", "featured", "community",
		"local", "global", "discussion", "help", "tips",
	}

	if count > len(allTags) {
		count = len(allTags)
	}

	tags := make([]string, count)
	for i := 0; i < count; i++ {
		tags[i] = allTags[rand.Intn(len(allTags))]
	}

	return tags
}

func getRandomFollowCategories() []string {
	categories := []string{"close_friends", "family", "work", "interests"}
	count := rand.Intn(3) // 0-2 categories

	if count == 0 {
		return []string{}
	}

	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = categories[rand.Intn(len(categories))]
	}

	return result
}

func getRandomGroupID(groups []primitive.ObjectID) *primitive.ObjectID {
	if len(groups) == 0 || rand.Float32() < 0.5 { // 50% chance of no group
		return nil
	}

	groupID := groups[rand.Intn(len(groups))]
	return &groupID
}

func getRandomRecentTime() time.Time {
	now := time.Now()
	hours := rand.Intn(72) // Last 72 hours
	return now.Add(-time.Hour * time.Duration(hours))
}

func getRandomPastTime() time.Time {
	now := time.Now()
	days := rand.Intn(365) + 1 // Last 1-365 days
	return now.Add(-time.Hour * 24 * time.Duration(days))
}

func getRandomFutureTime() time.Time {
	now := time.Now()
	days := rand.Intn(90) + 1 // Next 1-90 days
	return now.Add(time.Hour * 24 * time.Duration(days))
}

func getNotificationContent(notifType string) (string, string) {
	switch notifType {
	case "like":
		return "New Like", "Someone liked your post"
	case "comment":
		return "New Comment", "Someone commented on your post"
	case "follow":
		return "New Follower", "Someone started following you"
	case "mention":
		return "You were mentioned", "Someone mentioned you in a post"
	case "message":
		return "New Message", "You have a new message"
	case "group_invite":
		return "Group Invitation", "You were invited to join a group"
	case "event_invite":
		return "Event Invitation", "You were invited to an event"
	case "post_share":
		return "Post Shared", "Someone shared your post"
	default:
		return "Notification", "You have a new notification"
	}
}

func getTargetType(notifType string) string {
	switch notifType {
	case "like", "comment", "post_share", "mention":
		return "post"
	case "follow":
		return "user"
	case "message":
		return "conversation"
	case "group_invite":
		return "group"
	case "event_invite":
		return "event"
	default:
		return "unknown"
	}
}

func getRandomTargetID(data SeedData, notifType string) primitive.ObjectID {
	switch getTargetType(notifType) {
	case "post":
		if len(data.Posts) > 0 {
			return data.Posts[rand.Intn(len(data.Posts))]
		}
	case "user":
		if len(data.Users) > 0 {
			return data.Users[rand.Intn(len(data.Users))]
		}
	case "group":
		if len(data.Groups) > 0 {
			return data.Groups[rand.Intn(len(data.Groups))]
		}
	case "event":
		if len(data.Events) > 0 {
			return data.Events[rand.Intn(len(data.Events))]
		}
	}

	// Fallback to a user ID
	if len(data.Users) > 0 {
		return data.Users[rand.Intn(len(data.Users))]
	}

	return primitive.NewObjectID()
}
