package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DataGenerator struct {
	db    *mongo.Database
	users []models.User
	posts []models.Post
}

type GenerationConfig struct {
	UserCount           int
	PostsPerUser        int
	MaxFollowsPerUser   int
	MaxLikesPerPost     int
	MaxCommentsPerPost  int
	CommentsPercentage  float64
	LikesPercentage     float64
	FollowsPercentage   float64
	CleanExisting       bool
	CreateStories       bool
	CreateGroups        bool
	CreateConversations bool
	CreateNotifications bool
	Verbose             bool
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse command line arguments
	genConfig := parseArgs()

	// Initialize database configuration
	config.MustLoad()
	config.InitDB()
	defer config.Disconnect()

	// Initialize data generator
	generator := &DataGenerator{
		db:    config.DB,
		users: make([]models.User, 0),
		posts: make([]models.Post, 0),
	}

	// Initialize faker with seed for consistent data
	gofakeit.Seed(time.Now().UnixNano())

	ctx := context.Background()

	printBanner()

	// Clean existing data if requested
	if genConfig.CleanExisting {
		log.Println("🧹 Cleaning existing data...")
		if err := generator.cleanExistingData(ctx); err != nil {
			log.Fatalf("Failed to clean existing data: %v", err)
		}
		log.Println("✅ Existing data cleaned")
	}

	// Generate data
	log.Printf("🚀 Starting data generation with config: %+v", genConfig)

	start := time.Now()

	// Step 1: Generate users
	log.Printf("👥 Generating %d users...", genConfig.UserCount)
	if err := generator.generateUsers(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate users: %v", err)
	}
	log.Printf("✅ Generated %d users", len(generator.users))

	// Step 2: Generate posts
	totalPosts := genConfig.UserCount * genConfig.PostsPerUser
	log.Printf("📝 Generating ~%d posts...", totalPosts)
	if err := generator.generatePosts(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate posts: %v", err)
	}
	log.Printf("✅ Generated %d posts", len(generator.posts))

	// Step 3: Generate relationships (follows)
	log.Println("🤝 Generating follow relationships...")
	if err := generator.generateFollows(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate follows: %v", err)
	}
	log.Println("✅ Generated follow relationships")

	// Step 4: Generate interactions (likes, comments)
	log.Println("💝 Generating likes...")
	if err := generator.generateLikes(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate likes: %v", err)
	}
	log.Println("✅ Generated likes")

	log.Println("💬 Generating comments...")
	if err := generator.generateComments(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate comments: %v", err)
	}
	log.Println("✅ Generated comments")

	// Step 5: Generate optional content
	if genConfig.CreateStories {
		log.Println("📱 Generating stories...")
		if err := generator.generateStories(ctx, genConfig); err != nil {
			log.Fatalf("Failed to generate stories: %v", err)
		}
		log.Println("✅ Generated stories")
	}

	if genConfig.CreateGroups {
		log.Println("👥 Generating groups...")
		if err := generator.generateGroups(ctx, genConfig); err != nil {
			log.Fatalf("Failed to generate groups: %v", err)
		}
		log.Println("✅ Generated groups")
	}

	if genConfig.CreateConversations {
		log.Println("💬 Generating conversations...")
		if err := generator.generateConversations(ctx, genConfig); err != nil {
			log.Fatalf("Failed to generate conversations: %v", err)
		}
		log.Println("✅ Generated conversations")
	}

	if genConfig.CreateNotifications {
		log.Println("🔔 Generating notifications...")
		if err := generator.generateNotifications(ctx, genConfig); err != nil {
			log.Fatalf("Failed to generate notifications: %v", err)
		}
		log.Println("✅ Generated notifications")
	}

	// Step 6: Update user statistics
	log.Println("📊 Updating user statistics...")
	if err := generator.updateUserStatistics(ctx); err != nil {
		log.Fatalf("Failed to update user statistics: %v", err)
	}
	log.Println("✅ Updated user statistics")

	// Step 7: Create admin user
	log.Println("👑 Creating admin user...")
	if err := generator.createAdminUser(ctx); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	log.Println("✅ Created admin user")

	duration := time.Since(start)

	// Print summary
	printSummary(generator, genConfig, duration)
}

func parseArgs() GenerationConfig {
	genConfig := GenerationConfig{
		UserCount:           50,
		PostsPerUser:        5,
		MaxFollowsPerUser:   15,
		MaxLikesPerPost:     25,
		MaxCommentsPerPost:  8,
		CommentsPercentage:  0.7,
		LikesPercentage:     0.8,
		FollowsPercentage:   0.6,
		CleanExisting:       false,
		CreateStories:       true,
		CreateGroups:        true,
		CreateConversations: true,
		CreateNotifications: true,
		Verbose:             false,
	}

	// Parse command line arguments
	args := os.Args[1:]
	for i, arg := range args {
		switch arg {
		case "--users", "-u":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil {
					genConfig.UserCount = count
				}
			}
		case "--posts", "-p":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil {
					genConfig.PostsPerUser = count
				}
			}
		case "--clean", "-c":
			genConfig.CleanExisting = true
		case "--no-stories":
			genConfig.CreateStories = false
		case "--no-groups":
			genConfig.CreateGroups = false
		case "--no-conversations":
			genConfig.CreateConversations = false
		case "--no-notifications":
			genConfig.CreateNotifications = false
		case "--verbose", "-v":
			genConfig.Verbose = true
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	return genConfig
}

func printHelp() {
	fmt.Println(`
Social Media API Data Generator

Usage: go run cmd/seed/main.go [options]

Options:
  -u, --users <count>        Number of users to generate (default: 50)
  -p, --posts <count>        Number of posts per user (default: 5)
  -c, --clean               Clean existing data before generation
  --no-stories              Skip story generation
  --no-groups               Skip group generation
  --no-conversations        Skip conversation generation
  --no-notifications        Skip notification generation
  -v, --verbose             Verbose output
  -h, --help                Show this help message

Examples:
  go run cmd/seed/main.go -u 100 -p 10 -c
  go run cmd/seed/main.go --users 25 --posts 3 --clean --verbose
`)
}

func printBanner() {
	fmt.Println(`
╔══════════════════════════════════════════════════════════════╗
║                 SOCIAL MEDIA DATA GENERATOR                  ║
║                                                              ║
║  🎲 Generating realistic social media data...               ║
║  📊 Users, Posts, Comments, Likes, Follows & More!          ║
╚══════════════════════════════════════════════════════════════╝
`)
}

func printSummary(generator *DataGenerator, genConfig GenerationConfig, duration time.Duration) {
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║                      GENERATION COMPLETE                     ║
╠══════════════════════════════════════════════════════════════╣
║  👥 Users Generated:        %-8d                      ║
║  📝 Posts Generated:        %-8d                      ║
║  ⏱️  Time Taken:           %-8s                      ║
║                                                              ║
║  🔗 Database: Connected and populated                        ║
║  📊 Statistics: Updated                                      ║
║  👑 Admin User: Created                                      ║
╚══════════════════════════════════════════════════════════════╝

🎉 Data generation completed successfully!

Sample Admin Credentials:
  Username: admin
  Email:    admin@example.com
  Password: admin123

Sample User Credentials:
  Username: user1, user2, user3, etc.
  Email:    user1@example.com, user2@example.com, etc.
  Password: password123

Ready to test your Social Media API! 🚀
`, len(generator.users), len(generator.posts), duration.Round(time.Second))
}

func (g *DataGenerator) cleanExistingData(ctx context.Context) error {
	collections := []string{
		"users", "posts", "comments", "likes", "follows", "stories",
		"groups", "group_members", "conversations", "messages",
		"notifications", "reports", "media", "hashtags", "mentions",
		"user_sessions", "content_engagements", "user_journeys",
		"recommendation_events", "user_behavior_profiles",
	}

	for _, collection := range collections {
		if _, err := g.db.Collection(collection).DeleteMany(ctx, bson.M{}); err != nil {
			log.Printf("Warning: Failed to clean collection %s: %v", collection, err)
		}
	}

	return nil
}

func (g *DataGenerator) generateUsers(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("users")
	users := make([]interface{}, 0, genConfig.UserCount)

	for i := 0; i < genConfig.UserCount; i++ {
		user := g.createRandomUser(i + 1)
		users = append(users, user)
		g.users = append(g.users, user)

		if genConfig.Verbose && (i+1)%10 == 0 {
			log.Printf("Generated %d/%d users", i+1, genConfig.UserCount)
		}
	}

	if _, err := collection.InsertMany(ctx, users); err != nil {
		return fmt.Errorf("failed to insert users: %w", err)
	}

	return nil
}
func (g *DataGenerator) createRandomUser(index int) models.User {
	// Call HashPassword and handle both return values
	hashedPassword, err := utils.HashPassword("password123")
	if err != nil {
		// Handle the error appropriately (e.g., log or panic for data generation)
		panic(fmt.Sprintf("Failed to hash password: %v", err)) // Adjust based on your error handling strategy
	}

	user := models.User{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: gofakeit.DateRange(time.Now().AddDate(-2, 0, 0), time.Now()),
		},
		Username:             fmt.Sprintf("user%d", index),
		Email:                fmt.Sprintf("user%d@example.com", index),
		Password:             hashedPassword, // Use the hashed password
		FirstName:            gofakeit.FirstName(),
		LastName:             gofakeit.LastName(),
		Bio:                  gofakeit.Sentence(rand.Intn(15) + 5),
		ProfilePic:           gofakeit.ImageURL(400, 400),
		CoverPic:             gofakeit.ImageURL(1200, 400),
		Website:              generateWebsite(),
		Location:             gofakeit.City() + ", " + gofakeit.StateAbr(),
		Phone:                gofakeit.Phone(),
		Gender:               randomGender(),
		IsVerified:           rand.Float64() < 0.1, // 10% verified users
		IsActive:             true,
		IsPrivate:            rand.Float64() < 0.2, // 20% private users
		Role:                 models.RoleUser,
		Language:             "en",
		Timezone:             "UTC",
		Theme:                randomTheme(),
		OnlineStatus:         randomOnlineStatus(),
		EmailVerified:        true,
		PrivacySettings:      models.DefaultPrivacySettings(),
		NotificationSettings: models.DefaultNotificationSettings(),
		SocialLinks:          generateSocialLinks(),
		IsPremium:            rand.Float64() < 0.05, // 5% premium users
	}

	// Set display name
	user.DisplayName = user.FirstName + " " + user.LastName

	// Set date of birth (18-65 years old)
	user.DateOfBirth = &[]time.Time{gofakeit.DateRange(
		time.Now().AddDate(-65, 0, 0),
		time.Now().AddDate(-18, 0, 0),
	)}[0]

	// Set last login and activity
	lastLogin := gofakeit.DateRange(user.CreatedAt, time.Now())
	user.LastLoginAt = &lastLogin
	user.LastActiveAt = &lastLogin

	user.BeforeCreate()
	user.UpdatedAt = user.CreatedAt

	return user
}

func (g *DataGenerator) generatePosts(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("posts")
	posts := make([]interface{}, 0)

	for _, user := range g.users {
		postsCount := rand.Intn(genConfig.PostsPerUser) + 1

		for i := 0; i < postsCount; i++ {
			post := g.createRandomPost(user)
			posts = append(posts, post)
			g.posts = append(g.posts, post)
		}
	}

	// Insert in batches to avoid memory issues
	batchSize := 100
	for i := 0; i < len(posts); i += batchSize {
		end := i + batchSize
		if end > len(posts) {
			end = len(posts)
		}

		if _, err := collection.InsertMany(ctx, posts[i:end]); err != nil {
			return fmt.Errorf("failed to insert posts batch: %w", err)
		}

		if genConfig.Verbose {
			log.Printf("Inserted posts batch %d-%d", i+1, end)
		}
	}

	return nil
}

func (g *DataGenerator) createRandomPost(user models.User) models.Post {
	contentTypes := []models.ContentType{
		models.ContentTypeText,
		models.ContentTypeImage,
		models.ContentTypeVideo,
		models.ContentTypeLink,
	}

	contentType := contentTypes[rand.Intn(len(contentTypes))]

	post := models.Post{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: gofakeit.DateRange(user.CreatedAt, time.Now()),
		},
		UserID:          user.ID,
		Content:         generatePostContent(contentType),
		ContentType:     contentType,
		Type:            "post",
		Visibility:      randomVisibility(),
		Language:        "en",
		Hashtags:        generateHashtags(),
		CommentsEnabled: true,
		LikesEnabled:    true,
		SharesEnabled:   true,
		IsPublished:     true,
		Source:          randomSource(),
	}

	// Add media if needed
	if contentType == models.ContentTypeImage || contentType == models.ContentTypeVideo {
		post.Media = generateMediaInfo(contentType)
	}

	// Add location sometimes
	if rand.Float64() < 0.3 {
		post.Location = &models.Location{
			Name:      gofakeit.City(),
			Address:   gofakeit.Address().Address,
			Latitude:  gofakeit.Latitude(),
			Longitude: gofakeit.Longitude(),
		}
	}

	post.BeforeCreate()
	post.UpdatedAt = post.CreatedAt
	publishTime := post.CreatedAt
	post.PublishedAt = &publishTime

	return post
}

func (g *DataGenerator) generateFollows(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("follows")
	follows := make([]interface{}, 0)

	for _, follower := range g.users {
		followCount := rand.Intn(genConfig.MaxFollowsPerUser) + 1
		following := make(map[primitive.ObjectID]bool)

		for i := 0; i < followCount; i++ {
			// Pick random user to follow (not self)
			var followee models.User
			maxAttempts := 10
			for attempts := 0; attempts < maxAttempts; attempts++ {
				followee = g.users[rand.Intn(len(g.users))]
				if followee.ID != follower.ID && !following[followee.ID] {
					break
				}
			}

			if followee.ID == follower.ID || following[followee.ID] {
				continue
			}

			following[followee.ID] = true

			follow := models.Follow{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(maxTime(follower.CreatedAt, followee.CreatedAt), time.Now()),
				},
				FollowerID: follower.ID,
				FolloweeID: followee.ID,
				Status:     models.FollowStatusAccepted,
			}

			follow.BeforeCreate()
			follows = append(follows, follow)
		}
	}

	if len(follows) > 0 {
		batchSize := 100
		for i := 0; i < len(follows); i += batchSize {
			end := i + batchSize
			if end > len(follows) {
				end = len(follows)
			}

			if _, err := collection.InsertMany(ctx, follows[i:end]); err != nil {
				return fmt.Errorf("failed to insert follows batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) generateLikes(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("likes")
	likes := make([]interface{}, 0)

	reactionTypes := []models.ReactionType{
		models.ReactionLike,
		models.ReactionLove,
		models.ReactionHaha,
		models.ReactionWow,
		models.ReactionSad,
		models.ReactionAngry,
	}

	for _, post := range g.posts {
		if rand.Float64() > genConfig.LikesPercentage {
			continue
		}

		likeCount := rand.Intn(genConfig.MaxLikesPerPost) + 1
		likedBy := make(map[primitive.ObjectID]bool)

		for i := 0; i < likeCount; i++ {
			user := g.users[rand.Intn(len(g.users))]
			if likedBy[user.ID] {
				continue
			}
			likedBy[user.ID] = true

			like := models.Like{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(post.CreatedAt, time.Now()),
				},
				UserID:       user.ID,
				TargetID:     post.ID,
				TargetType:   "post",
				ReactionType: reactionTypes[rand.Intn(len(reactionTypes))],
				Source:       randomSource(),
			}

			like.BeforeCreate()
			likes = append(likes, like)
		}
	}

	if len(likes) > 0 {
		batchSize := 100
		for i := 0; i < len(likes); i += batchSize {
			end := i + batchSize
			if end > len(likes) {
				end = len(likes)
			}

			if _, err := collection.InsertMany(ctx, likes[i:end]); err != nil {
				return fmt.Errorf("failed to insert likes batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) generateComments(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("comments")
	comments := make([]interface{}, 0)

	for _, post := range g.posts {
		if rand.Float64() > genConfig.CommentsPercentage {
			continue
		}

		commentCount := rand.Intn(genConfig.MaxCommentsPerPost) + 1

		for i := 0; i < commentCount; i++ {
			user := g.users[rand.Intn(len(g.users))]

			comment := models.Comment{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(post.CreatedAt, time.Now()),
				},
				UserID:      user.ID,
				PostID:      post.ID,
				Content:     gofakeit.Sentence(rand.Intn(10) + 3),
				ContentType: models.ContentTypeText,
				Level:       0,
				Source:      randomSource(),
			}

			comment.BeforeCreate()
			comments = append(comments, comment)
		}
	}

	if len(comments) > 0 {
		batchSize := 100
		for i := 0; i < len(comments); i += batchSize {
			end := i + batchSize
			if end > len(comments) {
				end = len(comments)
			}

			if _, err := collection.InsertMany(ctx, comments[i:end]); err != nil {
				return fmt.Errorf("failed to insert comments batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) generateStories(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("stories")
	stories := make([]interface{}, 0)

	for _, user := range g.users {
		if rand.Float64() < 0.3 { // 30% of users have stories
			storyCount := rand.Intn(3) + 1

			for i := 0; i < storyCount; i++ {
				story := models.Story{
					BaseModel: models.BaseModel{
						ID:        primitive.NewObjectID(),
						CreatedAt: gofakeit.DateRange(time.Now().Add(-24*time.Hour), time.Now()),
					},
					UserID:          user.ID,
					Content:         gofakeit.Sentence(rand.Intn(5) + 2),
					ContentType:     randomStoryContentType(),
					Duration:        rand.Intn(25) + 5, // 5-30 seconds
					Visibility:      randomVisibility(),
					AllowReplies:    true,
					AllowReactions:  true,
					AllowSharing:    true,
					AllowScreenshot: true,
				}

				story.BeforeCreate()
				stories = append(stories, story)
			}
		}
	}

	if len(stories) > 0 {
		if _, err := collection.InsertMany(ctx, stories); err != nil {
			return fmt.Errorf("failed to insert stories: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateGroups(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("groups")
	groups := make([]interface{}, 0)

	groupCount := genConfig.UserCount / 10 // One group per 10 users
	if groupCount < 1 {
		groupCount = 1
	}

	categories := []string{
		"technology", "sports", "entertainment", "education", "business",
		"health", "travel", "food", "art", "music",
	}

	for i := 0; i < groupCount; i++ {
		creator := g.users[rand.Intn(len(g.users))]

		group := models.Group{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(creator.CreatedAt, time.Now()),
			},
			Name:        gofakeit.Company() + " " + gofakeit.Word(),
			Description: gofakeit.Paragraph(2, 3, 8, " "),
			Privacy:     randomGroupPrivacy(),
			Category:    categories[rand.Intn(len(categories))],
			CreatedBy:   creator.ID,
			Tags:        generateHashtags(),
		}

		group.BeforeCreate()
		groups = append(groups, group)
	}

	if len(groups) > 0 {
		if _, err := collection.InsertMany(ctx, groups); err != nil {
			return fmt.Errorf("failed to insert groups: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateConversations(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("conversations")
	conversations := make([]interface{}, 0)

	conversationCount := genConfig.UserCount / 5 // One conversation per 5 users
	if conversationCount < 1 {
		conversationCount = 1
	}

	for i := 0; i < conversationCount; i++ {
		participants := make([]primitive.ObjectID, 2)
		participants[0] = g.users[rand.Intn(len(g.users))].ID

		// Find different user
		for {
			user := g.users[rand.Intn(len(g.users))]
			if user.ID != participants[0] {
				participants[1] = user.ID
				break
			}
		}

		conversation := models.Conversation{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(time.Now().AddDate(0, -6, 0), time.Now()),
			},
			Type:         "direct",
			Participants: participants,
			CreatedBy:    participants[0],
		}

		conversation.BeforeCreate()
		conversations = append(conversations, conversation)
	}

	if len(conversations) > 0 {
		if _, err := collection.InsertMany(ctx, conversations); err != nil {
			return fmt.Errorf("failed to insert conversations: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateNotifications(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("notifications")
	notifications := make([]interface{}, 0)

	notificationTypes := []models.NotificationType{
		models.NotificationLike,
		models.NotificationComment,
		models.NotificationFollow,
		models.NotificationMention,
	}

	for _, user := range g.users {
		notificationCount := rand.Intn(10) + 5 // 5-15 notifications per user

		for i := 0; i < notificationCount; i++ {
			actor := g.users[rand.Intn(len(g.users))]
			if actor.ID == user.ID {
				continue
			}

			notifType := notificationTypes[rand.Intn(len(notificationTypes))]

			notification := models.Notification{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(user.CreatedAt, time.Now()),
				},
				RecipientID: user.ID,
				ActorID:     actor.ID,
				Type:        notifType,
				Title:       generateNotificationTitle(notifType),
				Message:     generateNotificationMessage(notifType),
				IsRead:      rand.Float64() < 0.6, // 60% read
				Priority:    "medium",
			}

			if notification.IsRead {
				readTime := gofakeit.DateRange(notification.CreatedAt, time.Now())
				notification.ReadAt = &readTime
			}

			notification.BeforeCreate()
			notifications = append(notifications, notification)
		}
	}

	if len(notifications) > 0 {
		batchSize := 100
		for i := 0; i < len(notifications); i += batchSize {
			end := i + batchSize
			if end > len(notifications) {
				end = len(notifications)
			}

			if _, err := collection.InsertMany(ctx, notifications[i:end]); err != nil {
				return fmt.Errorf("failed to insert notifications batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) updateUserStatistics(ctx context.Context) error {
	// Update follower/following counts
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":             "$followee_id",
			"followers_count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := g.db.Collection("follows").Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	followerCounts := make(map[primitive.ObjectID]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID             primitive.ObjectID `bson:"_id"`
			FollowersCount int64              `bson:"followers_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		followerCounts[result.ID] = result.FollowersCount
	}

	// Update following counts
	pipeline[0]["$group"].(bson.M)["_id"] = "$follower_id"
	cursor, err = g.db.Collection("follows").Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	followingCounts := make(map[primitive.ObjectID]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID             primitive.ObjectID `bson:"_id"`
			FollowingCount int64              `bson:"followers_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		followingCounts[result.ID] = result.FollowingCount
	}

	// Update post counts
	pipeline = []bson.M{
		{"$group": bson.M{
			"_id":         "$user_id",
			"posts_count": bson.M{"$sum": 1},
		}},
	}

	cursor, err = g.db.Collection("posts").Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	postCounts := make(map[primitive.ObjectID]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID         primitive.ObjectID `bson:"_id"`
			PostsCount int64              `bson:"posts_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		postCounts[result.ID] = result.PostsCount
	}

	// Update users
	for _, user := range g.users {
		update := bson.M{
			"$set": bson.M{
				"followers_count": followerCounts[user.ID],
				"following_count": followingCounts[user.ID],
				"posts_count":     postCounts[user.ID],
				"updated_at":      time.Now(),
			},
		}

		_, err := g.db.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": user.ID},
			update,
		)
		if err != nil {
			log.Printf("Failed to update user %s statistics: %v", user.Username, err)
		}
	}

	return nil
}
func (g *DataGenerator) createAdminUser(ctx context.Context) error {
	// Hash the password and handle both return values
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin := models.User{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
		},
		Username:      "admin",
		Email:         "admin@example.com",
		Password:      hashedPassword, // Use the hashed password
		FirstName:     "Admin",
		LastName:      "User",
		DisplayName:   "Admin User",
		Bio:           "System Administrator",
		IsVerified:    true,
		IsActive:      true,
		Role:          models.RoleAdmin,
		EmailVerified: true,
		Language:      "en",
		Timezone:      "UTC",
		OnlineStatus:  "online",
	}

	admin.BeforeCreate()

	_, err = g.db.Collection("users").InsertOne(ctx, admin)
	return err
}

// Helper functions

func maxTime(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

func randomGender() string {
	genders := []string{"male", "female", "other", "prefer_not_to_say"}
	return genders[rand.Intn(len(genders))]
}

func randomTheme() string {
	themes := []string{"light", "dark", "auto"}
	return themes[rand.Intn(len(themes))]
}

func randomOnlineStatus() string {
	statuses := []string{"online", "offline", "away"}
	return statuses[rand.Intn(len(statuses))]
}

func randomVisibility() models.PrivacyLevel {
	visibilities := []models.PrivacyLevel{
		models.PrivacyPublic,
		models.PrivacyFriends,
		models.PrivacyPrivate,
	}
	return visibilities[rand.Intn(len(visibilities))]
}

func randomGroupPrivacy() models.GroupPrivacy {
	privacies := []models.GroupPrivacy{
		models.GroupPublic,
		models.GroupPrivate,
		models.GroupSecret,
	}
	return privacies[rand.Intn(len(privacies))]
}

func randomStoryContentType() models.ContentType {
	types := []models.ContentType{
		models.ContentTypeImage,
		models.ContentTypeVideo,
	}
	return types[rand.Intn(len(types))]
}

func randomSource() string {
	sources := []string{"web", "mobile", "api"}
	return sources[rand.Intn(len(sources))]
}

func generateWebsite() string {
	if rand.Float64() < 0.3 {
		return gofakeit.URL()
	}
	return ""
}

func generateSocialLinks() map[string]string {
	if rand.Float64() < 0.4 {
		links := make(map[string]string)
		platforms := []string{"twitter", "instagram", "linkedin", "github"}

		for _, platform := range platforms {
			if rand.Float64() < 0.5 {
				links[platform] = fmt.Sprintf("https://%s.com/%s", platform, gofakeit.Username())
			}
		}

		return links
	}
	return nil
}

func generatePostContent(contentType models.ContentType) string {
	switch contentType {
	case models.ContentTypeImage:
		return gofakeit.Sentence(rand.Intn(10)+3) + " 📸"
	case models.ContentTypeVideo:
		return gofakeit.Sentence(rand.Intn(8)+2) + " 🎥"
	case models.ContentTypeLink:
		return gofakeit.Sentence(rand.Intn(6)+2) + " " + gofakeit.URL()
	default:
		return gofakeit.Paragraph(1, 3, 15, " ")
	}
}

func generateHashtags() []string {
	hashtags := []string{
		"technology", "coding", "javascript", "golang", "react", "nodejs",
		"photography", "travel", "food", "fitness", "music", "art",
		"business", "startup", "marketing", "design", "ux", "ui",
		"nature", "sunset", "beach", "mountains", "city", "life",
		"motivation", "success", "learning", "education", "books",
	}

	count := rand.Intn(4) + 1 // 1-5 hashtags
	selected := make([]string, 0, count)
	used := make(map[string]bool)

	for i := 0; i < count; i++ {
		tag := hashtags[rand.Intn(len(hashtags))]
		if !used[tag] {
			selected = append(selected, tag)
			used[tag] = true
		}
	}

	return selected
}

func generateMediaInfo(contentType models.ContentType) []models.MediaInfo {
	media := models.MediaInfo{
		Type: string(contentType),
		Size: int64(rand.Intn(5000000) + 100000), // 100KB - 5MB
	}

	switch contentType {
	case models.ContentTypeImage:
		media.URL = gofakeit.ImageURL(800, 600)
		media.Width = 800
		media.Height = 600
		media.Thumbnail = gofakeit.ImageURL(200, 150)
	case models.ContentTypeVideo:
		media.URL = "https://example.com/video/" + gofakeit.UUID() + ".mp4"
		media.Width = 1920
		media.Height = 1080
		media.Duration = rand.Intn(300) + 10 // 10-310 seconds
		media.Thumbnail = gofakeit.ImageURL(320, 180)
	}

	return []models.MediaInfo{media}
}

func generateNotificationTitle(notifType models.NotificationType) string {
	switch notifType {
	case models.NotificationLike:
		return "New Like"
	case models.NotificationComment:
		return "New Comment"
	case models.NotificationFollow:
		return "New Follower"
	case models.NotificationMention:
		return "You were mentioned"
	default:
		return "Notification"
	}
}

func generateNotificationMessage(notifType models.NotificationType) string {
	switch notifType {
	case models.NotificationLike:
		return "Someone liked your post"
	case models.NotificationComment:
		return "Someone commented on your post"
	case models.NotificationFollow:
		return "Someone started following you"
	case models.NotificationMention:
		return "Someone mentioned you in a post"
	default:
		return "You have a new notification"
	}
}
