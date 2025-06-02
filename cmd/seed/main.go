package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataGenerator struct {
	db            *mongo.Database
	users         []models.User
	posts         []models.Post
	comments      []models.Comment
	stories       []models.Story
	groups        []models.Group
	conversations []models.Conversation
	hashtags      []models.Hashtag
	media         []models.Media
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
	CreateMentions      bool
	CreateHashtags      bool
	CreateMedia         bool
	CreateReports       bool
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
		db:            config.DB,
		users:         make([]models.User, 0),
		posts:         make([]models.Post, 0),
		comments:      make([]models.Comment, 0),
		stories:       make([]models.Story, 0),
		groups:        make([]models.Group, 0),
		conversations: make([]models.Conversation, 0),
		hashtags:      make([]models.Hashtag, 0),
		media:         make([]models.Media, 0),
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

	// Generate data in proper synchronized order
	log.Printf("🚀 Starting synchronized data generation with config: %+v", genConfig)
	start := time.Now()

	// Step 1: Generate foundation data (users, hashtags, media)
	if err := generator.generateFoundationData(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate foundation data: %v", err)
	}

	// Step 2: Generate content (posts, stories, groups)
	if err := generator.generateContentData(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate content data: %v", err)
	}

	// Step 3: Generate interactions (follows, likes, comments)
	if err := generator.generateInteractionData(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate interaction data: %v", err)
	}

	// Step 4: Generate messaging and conversations
	if err := generator.generateMessagingData(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate messaging data: %v", err)
	}

	// Step 5: Generate advanced features (notifications, reports, etc.)
	if err := generator.generateAdvancedData(ctx, genConfig); err != nil {
		log.Fatalf("Failed to generate advanced data: %v", err)
	}

	// Step 6: Update all statistics and create admin users
	if err := generator.finalizeData(ctx); err != nil {
		log.Fatalf("Failed to finalize data: %v", err)
	}

	duration := time.Since(start)
	printSummary(generator, genConfig, duration)
}

func parseArgs() GenerationConfig {
	genConfig := GenerationConfig{
		UserCount:           100,
		PostsPerUser:        8,
		MaxFollowsPerUser:   25,
		MaxLikesPerPost:     40,
		MaxCommentsPerPost:  12,
		CommentsPercentage:  0.75,
		LikesPercentage:     0.85,
		FollowsPercentage:   0.7,
		CleanExisting:       false,
		CreateStories:       true,
		CreateGroups:        true,
		CreateConversations: true,
		CreateNotifications: true,
		CreateMentions:      true,
		CreateHashtags:      true,
		CreateMedia:         true,
		CreateReports:       true,
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
		case "--minimal":
			genConfig.CreateStories = false
			genConfig.CreateMentions = false
			genConfig.CreateReports = false
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
Complete Social Media API Data Generator

Usage: go run cmd/seed/main.go [options]

Options:
  -u, --users <count>   Number of users to generate (default: 100)
  -p, --posts <count>   Number of posts per user (default: 8)
  -c, --clean           Clean existing data before generation
  --minimal             Generate minimal data (no stories, events, etc.)
  -v, --verbose         Verbose output
  -h, --help            Show this help message

Examples:
  go run cmd/seed/main.go -u 200 -p 15 -c -v
  go run cmd/seed/main.go --clean --minimal
`)
}

func printBanner() {
	fmt.Println(`
╔══════════════════════════════════════════════════════════════╗
║             SYNCHRONIZED SOCIAL MEDIA DATA GENERATOR         ║
║                                                              ║
║  🎲 Generating properly synchronized social media data...    ║
║  📊 Users → Posts → Comments → Likes → Follows → Messages   ║
║     All data will be properly linked and synchronized!      ║
╚══════════════════════════════════════════════════════════════╝
`)
}

// Step 1: Generate foundation data
func (g *DataGenerator) generateFoundationData(ctx context.Context, genConfig GenerationConfig) error {
	log.Printf("👥 Generating %d users...", genConfig.UserCount)
	if err := g.generateAndFetchUsers(ctx, genConfig); err != nil {
		return err
	}
	log.Printf("✅ Generated and synced %d users", len(g.users))

	if genConfig.CreateHashtags {
		log.Println("🏷️ Generating hashtags...")
		if err := g.generateAndFetchHashtags(ctx, genConfig); err != nil {
			return err
		}
		log.Printf("✅ Generated and synced %d hashtags", len(g.hashtags))
	}

	if genConfig.CreateMedia {
		log.Println("📸 Generating media files...")
		if err := g.generateAndFetchMedia(ctx, genConfig); err != nil {
			return err
		}
		log.Printf("✅ Generated and synced %d media files", len(g.media))
	}

	return nil
}

// Step 2: Generate content data
func (g *DataGenerator) generateContentData(ctx context.Context, genConfig GenerationConfig) error {
	totalPosts := genConfig.UserCount * genConfig.PostsPerUser
	log.Printf("📝 Generating ~%d posts...", totalPosts)
	if err := g.generateAndFetchPosts(ctx, genConfig); err != nil {
		return err
	}
	log.Printf("✅ Generated and synced %d posts", len(g.posts))

	if genConfig.CreateGroups {
		log.Println("👥 Generating groups...")
		if err := g.generateAndFetchGroups(ctx, genConfig); err != nil {
			return err
		}
		log.Printf("✅ Generated and synced %d groups", len(g.groups))

		log.Println("👥 Generating group memberships...")
		if err := g.generateGroupMemberships(ctx, genConfig); err != nil {
			return err
		}
	}

	if genConfig.CreateStories {
		log.Println("📱 Generating stories...")
		if err := g.generateAndFetchStories(ctx, genConfig); err != nil {
			return err
		}
		log.Printf("✅ Generated and synced %d stories", len(g.stories))
	}

	return nil
}

// Step 3: Generate interaction data
func (g *DataGenerator) generateInteractionData(ctx context.Context, genConfig GenerationConfig) error {
	log.Println("🤝 Generating follow relationships...")
	if err := g.generateFollows(ctx, genConfig); err != nil {
		return err
	}

	log.Println("💝 Generating likes and reactions...")
	if err := g.generateLikes(ctx, genConfig); err != nil {
		return err
	}

	log.Println("💬 Generating comments...")
	if err := g.generateAndFetchComments(ctx, genConfig); err != nil {
		return err
	}
	log.Printf("✅ Generated and synced %d comments", len(g.comments))

	log.Println("💬 Generating comment replies...")
	if err := g.generateCommentReplies(ctx, genConfig); err != nil {
		return err
	}

	if genConfig.CreateMentions {
		log.Println("📢 Generating mentions...")
		if err := g.generateMentions(ctx, genConfig); err != nil {
			return err
		}
	}

	return nil
}

// Step 4: Generate messaging data
func (g *DataGenerator) generateMessagingData(ctx context.Context, genConfig GenerationConfig) error {
	if genConfig.CreateConversations {
		log.Println("💬 Generating conversations...")
		if err := g.generateAndFetchConversations(ctx, genConfig); err != nil {
			return err
		}
		log.Printf("✅ Generated and synced %d conversations", len(g.conversations))

		log.Println("📨 Generating messages...")
		if err := g.generateMessages(ctx, genConfig); err != nil {
			return err
		}
	}

	return nil
}

// Step 5: Generate advanced data
func (g *DataGenerator) generateAdvancedData(ctx context.Context, genConfig GenerationConfig) error {
	log.Println("🔄 Generating post shares...")
	if err := g.generatePostShares(ctx, genConfig); err != nil {
		return err
	}

	if genConfig.CreateStories {
		log.Println("👁️ Generating story views...")
		if err := g.generateStoryViews(ctx, genConfig); err != nil {
			return err
		}

		log.Println("⭐ Generating story highlights...")
		if err := g.generateStoryHighlights(ctx, genConfig); err != nil {
			return err
		}
	}

	if genConfig.CreateNotifications {
		log.Println("🔔 Generating notifications...")
		if err := g.generateNotifications(ctx, genConfig); err != nil {
			return err
		}
	}

	if genConfig.CreateReports {
		log.Println("🚨 Generating content reports...")
		if err := g.generateReports(ctx, genConfig); err != nil {
			return err
		}
	}

	log.Println("🚫 Generating user blocks...")
	if err := g.generateUserBlocks(ctx, genConfig); err != nil {
		return err
	}

	return nil
}

// Step 6: Finalize data
func (g *DataGenerator) finalizeData(ctx context.Context) error {
	log.Println("📊 Updating comprehensive statistics...")
	if err := g.updateAllStatistics(ctx); err != nil {
		return err
	}

	log.Println("👑 Creating admin and test users...")
	if err := g.createAdminAndTestUsers(ctx); err != nil {
		return err
	}

	return nil
}

func (g *DataGenerator) cleanExistingData(ctx context.Context) error {
	collections := []string{
		"users", "posts", "comments", "likes", "follows", "stories", "story_views", "story_highlights",
		"groups", "group_members", "group_invites", "conversations", "messages",
		"notifications", "reports", "media", "hashtags", "mentions", "blocks",
	}

	for _, collection := range collections {
		if _, err := g.db.Collection(collection).DeleteMany(ctx, bson.M{}); err != nil {
			log.Printf("Warning: Failed to clean collection %s: %v", collection, err)
		}
	}

	return nil
}

// Synchronized data generation methods
func (g *DataGenerator) generateAndFetchUsers(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("users")
	users := make([]interface{}, 0, genConfig.UserCount)

	for i := 0; i < genConfig.UserCount; i++ {
		user := g.createRandomUser(i + 1)
		users = append(users, user)

		if genConfig.Verbose && (i+1)%25 == 0 {
			log.Printf("Generated %d/%d users", i+1, genConfig.UserCount)
		}
	}

	// Insert users
	if _, err := collection.InsertMany(ctx, users); err != nil {
		return fmt.Errorf("failed to insert users: %w", err)
	}

	// Fetch users back from database to get actual stored data
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	defer cursor.Close(ctx)

	g.users = make([]models.User, 0)
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode user: %v", err)
			continue
		}
		g.users = append(g.users, user)
	}

	return nil
}

func (g *DataGenerator) generateAndFetchHashtags(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("hashtags")

	popularTags := []string{
		"technology", "coding", "javascript", "golang", "react", "nodejs", "ai", "machinelearning",
		"photography", "travel", "food", "fitness", "music", "art", "design", "ux", "ui",
		"business", "startup", "marketing", "entrepreneur", "success", "motivation", "inspiration",
		"nature", "sunset", "beach", "mountains", "city", "life", "love", "family", "friends",
		"education", "learning", "books", "reading", "writing", "productivity", "mindfulness",
		"health", "wellness", "yoga", "meditation", "running", "cycling", "sports", "football",
		"entertainment", "movies", "netflix", "gaming", "streaming", "youtube", "tiktok",
		"fashion", "style", "beauty", "skincare", "makeup", "outfit", "trending", "viral",
	}

	hashtags := make([]interface{}, 0, len(popularTags))

	for _, tag := range popularTags {
		hashtag := models.Hashtag{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(time.Now().AddDate(-1, 0, 0), time.Now()),
			},
			Tag:          tag,
			DisplayTag:   tag,
			Category:     randomHashtagCategory(),
			Language:     "en",
			PostsCount:   int64(rand.Intn(1000) + 10),
			StoriesCount: int64(rand.Intn(500) + 5),
			IsTrending:   rand.Float64() < 0.2,
			IsFeatured:   rand.Float64() < 0.1,
		}

		hashtag.BeforeCreate()
		hashtags = append(hashtags, hashtag)
	}

	// Insert hashtags
	if _, err := collection.InsertMany(ctx, hashtags); err != nil {
		return fmt.Errorf("failed to insert hashtags: %w", err)
	}

	// Fetch hashtags back from database
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to fetch hashtags: %w", err)
	}
	defer cursor.Close(ctx)

	g.hashtags = make([]models.Hashtag, 0)
	for cursor.Next(ctx) {
		var hashtag models.Hashtag
		if err := cursor.Decode(&hashtag); err != nil {
			log.Printf("Failed to decode hashtag: %v", err)
			continue
		}
		g.hashtags = append(g.hashtags, hashtag)
	}

	return nil
}

func (g *DataGenerator) generateAndFetchMedia(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("media")
	media := make([]interface{}, 0)

	// Generate media for about 70% of users
	for _, user := range g.users {
		if rand.Float64() < 0.7 {
			mediaCount := rand.Intn(8) + 2 // 2-10 media files per user

			for i := 0; i < mediaCount; i++ {
				mediaFile := g.createRandomMedia(user)
				media = append(media, mediaFile)
			}
		}
	}

	if len(media) > 0 {
		// Insert media in batches
		batchSize := 100
		for i := 0; i < len(media); i += batchSize {
			end := i + batchSize
			if end > len(media) {
				end = len(media)
			}

			if _, err := collection.InsertMany(ctx, media[i:end]); err != nil {
				return fmt.Errorf("failed to insert media batch: %w", err)
			}
		}

		// Fetch media back from database
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			return fmt.Errorf("failed to fetch media: %w", err)
		}
		defer cursor.Close(ctx)

		g.media = make([]models.Media, 0)
		for cursor.Next(ctx) {
			var mediaItem models.Media
			if err := cursor.Decode(&mediaItem); err != nil {
				log.Printf("Failed to decode media: %v", err)
				continue
			}
			g.media = append(g.media, mediaItem)
		}
	}

	return nil
}

func (g *DataGenerator) generateAndFetchPosts(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("posts")
	posts := make([]interface{}, 0)

	for _, user := range g.users {
		postsCount := rand.Intn(genConfig.PostsPerUser) + 1

		for i := 0; i < postsCount; i++ {
			post := g.createRandomPostWithUserRef(user)
			posts = append(posts, post)
		}
	}

	// Insert posts in batches
	batchSize := 100
	for i := 0; i < len(posts); i += batchSize {
		end := i + batchSize
		if end > len(posts) {
			end = len(posts)
		}

		if _, err := collection.InsertMany(ctx, posts[i:end]); err != nil {
			return fmt.Errorf("failed to insert posts batch: %w", err)
		}
	}

	// Fetch posts back from database
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer cursor.Close(ctx)

	g.posts = make([]models.Post, 0)
	for cursor.Next(ctx) {
		var post models.Post
		if err := cursor.Decode(&post); err != nil {
			log.Printf("Failed to decode post: %v", err)
			continue
		}
		g.posts = append(g.posts, post)
	}

	return nil
}

func (g *DataGenerator) generateAndFetchGroups(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("groups")
	groups := make([]interface{}, 0)

	groupCount := genConfig.UserCount / 8 // One group per 8 users
	if groupCount < 5 {
		groupCount = 5
	}

	categories := []string{
		"technology", "sports", "entertainment", "education", "business",
		"health", "travel", "food", "art", "music", "gaming", "fitness",
		"photography", "books", "movies", "science", "nature", "pets",
	}

	usedSlugs := make(map[string]bool)

	for i := 0; i < groupCount; i++ {
		creator := g.users[rand.Intn(len(g.users))]

		// Generate unique slug
		slug := generateUniqueSlug(usedSlugs, i)
		usedSlugs[slug] = true

		group := models.Group{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(creator.CreatedAt, time.Now()),
			},
			Name:        generateGroupName(),
			Slug:        slug,
			Description: gofakeit.Paragraph(2, 4, 10, " "),
			Privacy:     randomGroupPrivacy(),
			Category:    categories[rand.Intn(len(categories))],
			CreatedBy:   creator.ID,
			Tags:        selectRandomHashtags(g.hashtags),
			ProfilePic:  gofakeit.ImageURL(300, 300),
			CoverPic:    gofakeit.ImageURL(1200, 400),
		}

		group.BeforeCreate()
		groups = append(groups, group)
	}

	// Insert groups
	if _, err := collection.InsertMany(ctx, groups); err != nil {
		return fmt.Errorf("failed to insert groups: %w", err)
	}

	// Fetch groups back from database
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to fetch groups: %w", err)
	}
	defer cursor.Close(ctx)

	g.groups = make([]models.Group, 0)
	for cursor.Next(ctx) {
		var group models.Group
		if err := cursor.Decode(&group); err != nil {
			log.Printf("Failed to decode group: %v", err)
			continue
		}
		g.groups = append(g.groups, group)
	}

	return nil
}

func (g *DataGenerator) generateAndFetchStories(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("stories")
	stories := make([]interface{}, 0)

	for _, user := range g.users {
		if rand.Float64() < 0.4 { // 40% of users have stories
			storyCount := rand.Intn(4) + 1

			for i := 0; i < storyCount; i++ {
				story := models.Story{
					BaseModel: models.BaseModel{
						ID:        primitive.NewObjectID(),
						CreatedAt: gofakeit.DateRange(time.Now().Add(-24*time.Hour), time.Now()),
					},
					UserID:          user.ID,
					Content:         gofakeit.Sentence(rand.Intn(8) + 2),
					ContentType:     randomStoryContentType(),
					Duration:        rand.Intn(25) + 5,
					Visibility:      randomVisibility(),
					AllowReplies:    true,
					AllowReactions:  true,
					AllowSharing:    true,
					AllowScreenshot: rand.Float64() < 0.8,
					BackgroundColor: randomColor(),
					TextColor:       randomColor(),
					FontFamily:      randomFontFamily(),
				}

				// Add media
				story.Media = models.MediaInfo{
					URL:       generateMediaURL(string(story.ContentType)),
					Type:      string(story.ContentType),
					Size:      int64(rand.Intn(5000000) + 500000),
					Width:     1080,
					Height:    1920,
					Duration:  story.Duration,
					Thumbnail: gofakeit.ImageURL(200, 356),
				}

				story.BeforeCreate()
				stories = append(stories, story)
			}
		}
	}

	if len(stories) > 0 {
		// Insert stories
		if _, err := collection.InsertMany(ctx, stories); err != nil {
			return fmt.Errorf("failed to insert stories: %w", err)
		}

		// Fetch stories back from database
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			return fmt.Errorf("failed to fetch stories: %w", err)
		}
		defer cursor.Close(ctx)

		g.stories = make([]models.Story, 0)
		for cursor.Next(ctx) {
			var story models.Story
			if err := cursor.Decode(&story); err != nil {
				log.Printf("Failed to decode story: %v", err)
				continue
			}
			g.stories = append(g.stories, story)
		}
	}

	return nil
}

func (g *DataGenerator) generateAndFetchComments(ctx context.Context, genConfig GenerationConfig) error {
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
				Content:     gofakeit.Sentence(rand.Intn(15) + 3),
				ContentType: models.ContentTypeText,
				Level:       0,
				Source:      randomSource(),
			}

			comment.BeforeCreate()
			comments = append(comments, comment)
		}
	}

	if len(comments) > 0 {
		// Insert comments in batches
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

		// Fetch comments back from database
		cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
		if err != nil {
			return fmt.Errorf("failed to fetch comments: %w", err)
		}
		defer cursor.Close(ctx)

		g.comments = make([]models.Comment, 0)
		for cursor.Next(ctx) {
			var comment models.Comment
			if err := cursor.Decode(&comment); err != nil {
				log.Printf("Failed to decode comment: %v", err)
				continue
			}
			g.comments = append(g.comments, comment)
		}
	}

	return nil
}

func (g *DataGenerator) generateAndFetchConversations(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("conversations")
	conversations := make([]interface{}, 0)

	conversationCount := genConfig.UserCount / 4
	if conversationCount < 10 {
		conversationCount = 10
	}

	for i := 0; i < conversationCount; i++ {
		participants := make([]primitive.ObjectID, 0)

		if rand.Float64() < 0.8 { // 80% direct conversations
			// Direct conversation (2 participants)
			user1 := g.users[rand.Intn(len(g.users))]
			var user2 models.User
			for {
				user2 = g.users[rand.Intn(len(g.users))]
				if user2.ID != user1.ID {
					break
				}
			}
			participants = []primitive.ObjectID{user1.ID, user2.ID}
		} else {
			// Group conversation (3-8 participants)
			participantCount := rand.Intn(6) + 3
			usedUsers := make(map[primitive.ObjectID]bool)

			for len(participants) < participantCount {
				user := g.users[rand.Intn(len(g.users))]
				if !usedUsers[user.ID] {
					participants = append(participants, user.ID)
					usedUsers[user.ID] = true
				}
			}
		}

		convType := "direct"
		if len(participants) > 2 {
			convType = "group"
		}

		conversation := models.Conversation{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(time.Now().AddDate(0, -6, 0), time.Now()),
			},
			Type:         convType,
			Participants: participants,
			CreatedBy:    participants[0],
		}

		if convType == "group" {
			conversation.Title = generateGroupConversationTitle()
			conversation.Description = gofakeit.Sentence(8)
		}

		conversation.BeforeCreate()
		conversations = append(conversations, conversation)
	}

	// Insert conversations
	if _, err := collection.InsertMany(ctx, conversations); err != nil {
		return fmt.Errorf("failed to insert conversations: %w", err)
	}

	// Fetch conversations back from database
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to fetch conversations: %w", err)
	}
	defer cursor.Close(ctx)

	g.conversations = make([]models.Conversation, 0)
	for cursor.Next(ctx) {
		var conversation models.Conversation
		if err := cursor.Decode(&conversation); err != nil {
			log.Printf("Failed to decode conversation: %v", err)
			continue
		}
		g.conversations = append(g.conversations, conversation)
	}

	return nil
}

// Helper method to create posts with proper user references
func (g *DataGenerator) createRandomPostWithUserRef(user models.User) models.Post {
	contentTypes := []models.ContentType{
		models.ContentTypeText,
		models.ContentTypeImage,
		models.ContentTypeVideo,
		models.ContentTypeLink,
		models.ContentTypePoll,
	}

	contentType := contentTypes[rand.Intn(len(contentTypes))]

	post := models.Post{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: gofakeit.DateRange(user.CreatedAt, time.Now()),
		},
		UserID:          user.ID, // Use actual user ID from database
		Content:         generatePostContent(contentType),
		ContentType:     contentType,
		Type:            "post",
		Visibility:      randomVisibility(),
		Language:        "en",
		Hashtags:        selectRandomHashtags(g.hashtags),
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

	// Add poll data for poll posts
	if contentType == models.ContentTypePoll {
		post.PollOptions = generatePollOptions()
		expiry := post.CreatedAt.Add(time.Hour * 24 * time.Duration(rand.Intn(7)+1))
		post.PollExpiresAt = &expiry
		post.PollMultiple = rand.Float64() < 0.3
	}

	// Assign to group occasionally
	if len(g.groups) > 0 && rand.Float64() < 0.2 {
		group := g.groups[rand.Intn(len(g.groups))]
		post.GroupID = &group.ID
	}

	post.BeforeCreate()
	post.UpdatedAt = post.CreatedAt
	publishTime := post.CreatedAt
	post.PublishedAt = &publishTime

	return post
}

func (g *DataGenerator) createRandomUser(index int) models.User {
	hashedPassword, _ := utils.HashPassword("password123")

	user := models.User{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: gofakeit.DateRange(time.Now().AddDate(-2, 0, 0), time.Now()),
		},
		Username:             fmt.Sprintf("user%d", index),
		Email:                fmt.Sprintf("user%d@example.com", index),
		Password:             hashedPassword,
		FirstName:            gofakeit.FirstName(),
		LastName:             gofakeit.LastName(),
		Bio:                  gofakeit.Sentence(rand.Intn(15) + 5),
		ProfilePic:           gofakeit.ImageURL(400, 400),
		CoverPic:             gofakeit.ImageURL(1200, 400),
		Website:              generateWebsite(),
		Location:             gofakeit.City() + ", " + gofakeit.StateAbr(),
		Phone:                gofakeit.Phone(),
		Gender:               randomGender(),
		IsVerified:           rand.Float64() < 0.15, // 15% verified users
		IsActive:             true,
		IsPrivate:            rand.Float64() < 0.25, // 25% private users
		Role:                 models.RoleUser,
		Language:             "en",
		Timezone:             "UTC",
		Theme:                randomTheme(),
		OnlineStatus:         randomOnlineStatus(),
		EmailVerified:        true,
		PrivacySettings:      models.DefaultPrivacySettings(),
		NotificationSettings: models.DefaultNotificationSettings(),
		SocialLinks:          generateSocialLinks(),
		IsPremium:            rand.Float64() < 0.08, // 8% premium users
	}

	user.DisplayName = user.FirstName + " " + user.LastName

	dob := gofakeit.DateRange(time.Now().AddDate(-65, 0, 0), time.Now().AddDate(-18, 0, 0))
	user.DateOfBirth = &dob

	lastLogin := gofakeit.DateRange(user.CreatedAt, time.Now())
	user.LastLoginAt = &lastLogin
	user.LastActiveAt = &lastLogin

	user.BeforeCreate()
	user.UpdatedAt = user.CreatedAt

	return user
}

func (g *DataGenerator) createRandomMedia(user models.User) models.Media {
	mediaTypes := []string{"image", "video", "audio", "document"}
	mediaType := mediaTypes[rand.Intn(len(mediaTypes))]

	media := models.Media{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: gofakeit.DateRange(user.CreatedAt, time.Now()),
		},
		OriginalName:     gofakeit.Word() + getFileExtension(mediaType),
		FileName:         gofakeit.UUID() + getFileExtension(mediaType),
		FilePath:         "/uploads/" + gofakeit.UUID() + getFileExtension(mediaType),
		FileSize:         int64(rand.Intn(10000000) + 100000), // 100KB - 10MB
		MimeType:         getMimeType(mediaType),
		FileExtension:    getFileExtension(mediaType),
		Type:             mediaType,
		Category:         randomMediaCategory(),
		URL:              generateMediaURL(mediaType),
		UploadedBy:       user.ID,              // Use actual user ID
		IsPublic:         rand.Float64() < 0.8, // 80% public
		AccessPolicy:     "public",
		IsProcessed:      true,
		ProcessingStatus: "completed",
		StorageProvider:  "s3",
		StorageKey:       gofakeit.UUID(),
		AltText:          gofakeit.Sentence(3),
		Description:      gofakeit.Sentence(8),
	}

	if mediaType == "image" || mediaType == "video" {
		media.Width = []int{640, 800, 1200, 1920}[rand.Intn(4)]
		media.Height = []int{480, 600, 900, 1080}[rand.Intn(4)]
	}

	if mediaType == "video" || mediaType == "audio" {
		media.Duration = rand.Intn(600) + 30 // 30s - 10min
	}

	// Add thumbnails for images and videos
	if mediaType == "image" || mediaType == "video" {
		media.Thumbnails = []models.MediaVariant{
			{
				Name:      "thumbnail",
				URL:       gofakeit.ImageURL(150, 150),
				Width:     150,
				Height:    150,
				FileSize:  15000,
				Format:    "jpg",
				CreatedAt: media.CreatedAt,
			},
			{
				Name:      "small",
				URL:       gofakeit.ImageURL(300, 300),
				Width:     300,
				Height:    300,
				FileSize:  45000,
				Format:    "jpg",
				CreatedAt: media.CreatedAt,
			},
		}
	}

	media.BeforeCreate()
	processedAt := media.CreatedAt.Add(time.Minute * 2)
	media.ProcessedAt = &processedAt

	return media
}

// Rest of the implementation methods...
func (g *DataGenerator) generateFollows(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("follows")
	follows := make([]interface{}, 0)

	for _, follower := range g.users {
		followCount := rand.Intn(genConfig.MaxFollowsPerUser) + 1
		following := make(map[primitive.ObjectID]bool)

		for i := 0; i < followCount; i++ {
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

			status := models.FollowStatusAccepted
			// Create some pending requests for private accounts
			if followee.IsPrivate && rand.Float64() < 0.3 {
				status = models.FollowStatusPending
			}

			follow := models.Follow{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(maxTime(follower.CreatedAt, followee.CreatedAt), time.Now()),
				},
				FollowerID: follower.ID,
				FolloweeID: followee.ID,
				Status:     status,
				Categories: randomFollowCategories(),
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

func (g *DataGenerator) generateGroupMemberships(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("group_members")
	members := make([]interface{}, 0)

	for _, group := range g.groups {
		// Add creator as admin
		creatorMembership := models.GroupMember{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: group.CreatedAt,
			},
			GroupID: group.ID,
			UserID:  group.CreatedBy,
			Role:    models.GroupRoleAdmin,
			Status:  "active",
		}
		creatorMembership.BeforeCreate()
		members = append(members, creatorMembership)

		// Add random members
		memberCount := rand.Intn(50) + 5
		addedMembers := make(map[primitive.ObjectID]bool)
		addedMembers[group.CreatedBy] = true

		for i := 0; i < memberCount && len(addedMembers) < len(g.users); i++ {
			user := g.users[rand.Intn(len(g.users))]
			if addedMembers[user.ID] {
				continue
			}
			addedMembers[user.ID] = true

			role := models.GroupRoleMember
			if rand.Float64() < 0.1 { // 10% moderators
				role = models.GroupRoleModerator
			}

			member := models.GroupMember{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(group.CreatedAt, time.Now()),
				},
				GroupID: group.ID,
				UserID:  user.ID,
				Role:    role,
				Status:  "active",
			}

			member.BeforeCreate()
			members = append(members, member)
		}
	}

	if len(members) > 0 {
		batchSize := 100
		for i := 0; i < len(members); i += batchSize {
			end := i + batchSize
			if end > len(members) {
				end = len(members)
			}

			if _, err := collection.InsertMany(ctx, members[i:end]); err != nil {
				return fmt.Errorf("failed to insert group members batch: %w", err)
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
		models.ReactionSupport,
	}

	// Likes on posts
	for _, post := range g.posts {
		if rand.Float64() > genConfig.LikesPercentage {
			continue
		}

		likeCount := rand.Intn(genConfig.MaxLikesPerPost) + 1
		likedBy := make(map[primitive.ObjectID]bool)

		for i := 0; i < likeCount && len(likedBy) < len(g.users); i++ {
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

	// Likes on stories
	for _, story := range g.stories {
		if rand.Float64() < 0.6 { // 60% of stories get likes
			likeCount := rand.Intn(15) + 1
			likedBy := make(map[primitive.ObjectID]bool)

			for i := 0; i < likeCount && len(likedBy) < len(g.users); i++ {
				user := g.users[rand.Intn(len(g.users))]
				if likedBy[user.ID] {
					continue
				}
				likedBy[user.ID] = true

				like := models.Like{
					BaseModel: models.BaseModel{
						ID:        primitive.NewObjectID(),
						CreatedAt: gofakeit.DateRange(story.CreatedAt, time.Now()),
					},
					UserID:       user.ID,
					TargetID:     story.ID,
					TargetType:   "story",
					ReactionType: reactionTypes[rand.Intn(len(reactionTypes))],
					Source:       randomSource(),
				}

				like.BeforeCreate()
				likes = append(likes, like)
			}
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

func (g *DataGenerator) generateCommentReplies(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("comments")
	replies := make([]interface{}, 0)

	// Generate replies to existing comments
	for _, comment := range g.comments {
		if rand.Float64() < 0.4 { // 40% of comments get replies
			replyCount := rand.Intn(3) + 1

			for i := 0; i < replyCount; i++ {
				user := g.users[rand.Intn(len(g.users))]

				reply := models.Comment{
					BaseModel: models.BaseModel{
						ID:        primitive.NewObjectID(),
						CreatedAt: gofakeit.DateRange(comment.CreatedAt, time.Now()),
					},
					UserID:          user.ID,
					PostID:          comment.PostID,
					ParentCommentID: &comment.ID,
					RootCommentID:   &comment.ID,
					Content:         gofakeit.Sentence(rand.Intn(10) + 2),
					ContentType:     models.ContentTypeText,
					Level:           1,
					Source:          randomSource(),
				}

				reply.BeforeCreate()
				replies = append(replies, reply)
			}
		}
	}

	if len(replies) > 0 {
		batchSize := 100
		for i := 0; i < len(replies); i += batchSize {
			end := i + batchSize
			if end > len(replies) {
				end = len(replies)
			}

			if _, err := collection.InsertMany(ctx, replies[i:end]); err != nil {
				return fmt.Errorf("failed to insert replies batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) generateMentions(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("mentions")
	mentions := make([]interface{}, 0)

	// Generate mentions in posts
	for _, post := range g.posts {
		if rand.Float64() < 0.3 {
			mentionCount := rand.Intn(3) + 1

			for i := 0; i < mentionCount; i++ {
				mentionedUser := g.users[rand.Intn(len(g.users))]
				if mentionedUser.ID == post.UserID {
					continue
				}

				mention := models.Mention{
					BaseModel: models.BaseModel{
						ID:        primitive.NewObjectID(),
						CreatedAt: post.CreatedAt,
					},
					MentionerID:   post.UserID,
					MentionedID:   mentionedUser.ID,
					ContentType:   "post",
					ContentID:     post.ID,
					MentionText:   "@" + mentionedUser.Username,
					StartPosition: rand.Intn(len(post.Content)),
					EndPosition:   rand.Intn(len(post.Content)),
				}

				mention.BeforeCreate()
				mentions = append(mentions, mention)
			}
		}
	}

	// Generate mentions in comments
	for _, comment := range g.comments {
		if rand.Float64() < 0.2 && len(g.users) > 1 {
			mentionedUser := g.users[rand.Intn(len(g.users))]
			if mentionedUser.ID == comment.UserID {
				continue
			}

			mention := models.Mention{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: comment.CreatedAt,
				},
				MentionerID:   comment.UserID,
				MentionedID:   mentionedUser.ID,
				ContentType:   "comment",
				ContentID:     comment.ID,
				MentionText:   "@" + mentionedUser.Username,
				StartPosition: 0,
				EndPosition:   len(mentionedUser.Username) + 1,
			}

			mention.BeforeCreate()
			mentions = append(mentions, mention)
		}
	}

	if len(mentions) > 0 {
		if _, err := collection.InsertMany(ctx, mentions); err != nil {
			return fmt.Errorf("failed to insert mentions: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateMessages(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("messages")
	messages := make([]interface{}, 0)

	for _, conversation := range g.conversations {
		messageCount := rand.Intn(20) + 5 // 5-25 messages per conversation

		for i := 0; i < messageCount; i++ {
			sender := conversation.Participants[rand.Intn(len(conversation.Participants))]

			contentType := models.ContentTypeText
			if rand.Float64() < 0.1 {
				contentTypes := []models.ContentType{
					models.ContentTypeImage,
					models.ContentTypeVideo,
					models.ContentTypeAudio,
					models.ContentTypeFile,
				}
				contentType = contentTypes[rand.Intn(len(contentTypes))]
			}

			message := models.Message{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(conversation.CreatedAt, time.Now()),
				},
				ConversationID: conversation.ID,
				SenderID:       sender,
				Content:        generateMessageContent(contentType),
				ContentType:    contentType,
				Status:         models.MessageRead,
				Priority:       "normal",
			}

			if contentType != models.ContentTypeText {
				message.Media = generateMediaInfo(contentType)
			}

			message.BeforeCreate()
			sentAt := message.CreatedAt
			deliveredAt := message.CreatedAt.Add(time.Second * 5)
			readAt := message.CreatedAt.Add(time.Minute * 2)

			message.SentAt = &sentAt
			message.DeliveredAt = &deliveredAt
			message.ReadAt = &readAt

			messages = append(messages, message)
		}
	}

	if len(messages) > 0 {
		batchSize := 100
		for i := 0; i < len(messages); i += batchSize {
			end := i + batchSize
			if end > len(messages) {
				end = len(messages)
			}

			if _, err := collection.InsertMany(ctx, messages[i:end]); err != nil {
				return fmt.Errorf("failed to insert messages batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) generatePostShares(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("posts")
	shares := make([]interface{}, 0)

	shareCount := len(g.posts) / 20 // About 5% of posts are shared

	for i := 0; i < shareCount; i++ {
		originalPost := g.posts[rand.Intn(len(g.posts))]
		sharer := g.users[rand.Intn(len(g.users))]

		if sharer.ID == originalPost.UserID {
			continue
		}

		repost := models.Post{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(originalPost.CreatedAt, time.Now()),
			},
			UserID:         sharer.ID,
			Content:        generateRepostComment(),
			ContentType:    models.ContentTypeText,
			Type:           "post",
			Visibility:     randomVisibility(),
			Language:       "en",
			IsRepost:       true,
			OriginalPostID: &originalPost.ID,
			RepostComment:  generateRepostComment(),
			IsPublished:    true,
			Source:         randomSource(),
		}

		repost.BeforeCreate()
		publishTime := repost.CreatedAt
		repost.PublishedAt = &publishTime

		shares = append(shares, repost)
	}

	if len(shares) > 0 {
		if _, err := collection.InsertMany(ctx, shares); err != nil {
			return fmt.Errorf("failed to insert shares: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateStoryViews(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("story_views")
	views := make([]interface{}, 0)

	for _, story := range g.stories {
		viewCount := rand.Intn(30) + 5 // 5-35 views per story
		viewedBy := make(map[primitive.ObjectID]bool)

		for i := 0; i < viewCount && len(viewedBy) < len(g.users); i++ {
			user := g.users[rand.Intn(len(g.users))]
			if viewedBy[user.ID] || user.ID == story.UserID {
				continue
			}
			viewedBy[user.ID] = true

			view := models.StoryView{
				BaseModel: models.BaseModel{
					ID:        primitive.NewObjectID(),
					CreatedAt: gofakeit.DateRange(story.CreatedAt, time.Now()),
				},
				StoryID:      story.ID,
				UserID:       user.ID,
				ViewDuration: float64(rand.Intn(story.Duration) + 1),
				WatchedFully: rand.Float64() < 0.7,
				Source:       randomSource(),
				DeviceType:   randomDeviceType(),
			}

			view.BeforeCreate()
			views = append(views, view)
		}
	}

	if len(views) > 0 {
		batchSize := 100
		for i := 0; i < len(views); i += batchSize {
			end := i + batchSize
			if end > len(views) {
				end = len(views)
			}

			if _, err := collection.InsertMany(ctx, views[i:end]); err != nil {
				return fmt.Errorf("failed to insert story views batch: %w", err)
			}
		}
	}

	return nil
}

func (g *DataGenerator) generateStoryHighlights(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("story_highlights")
	highlights := make([]interface{}, 0)

	// Group stories by user
	userStories := make(map[primitive.ObjectID][]models.Story)
	for _, story := range g.stories {
		userStories[story.UserID] = append(userStories[story.UserID], story)
	}

	for userID, stories := range userStories {
		if len(stories) > 3 && rand.Float64() < 0.4 { // 40% chance if user has 3+ stories
			highlightCount := rand.Intn(3) + 1 // 1-3 highlights per user

			for i := 0; i < highlightCount; i++ {
				storyCount := rand.Intn(len(stories)) + 1
				selectedStories := make([]primitive.ObjectID, 0, storyCount)

				for j := 0; j < storyCount && j < len(stories); j++ {
					selectedStories = append(selectedStories, stories[j].ID)
				}

				highlight := models.StoryHighlight{
					BaseModel: models.BaseModel{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
					},
					UserID:       userID,
					Title:        generateHighlightTitle(),
					CoverImage:   gofakeit.ImageURL(200, 200),
					StoryIDs:     selectedStories,
					StoriesCount: int64(len(selectedStories)),
					IsActive:     true,
					Order:        i + 1,
				}

				highlight.BeforeCreate()
				highlights = append(highlights, highlight)
			}
		}
	}

	if len(highlights) > 0 {
		if _, err := collection.InsertMany(ctx, highlights); err != nil {
			return fmt.Errorf("failed to insert story highlights: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateNotifications(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("notifications")
	notifications := make([]interface{}, 0)

	notificationTypes := []models.NotificationType{
		models.NotificationLike,
		models.NotificationLove,
		models.NotificationComment,
		models.NotificationFollow,
		models.NotificationMention,
		models.NotificationMessage,
		models.NotificationGroupInvite,
		models.NotificationEventInvite,
		models.NotificationPostShare,
		models.NotificationStoryView,
	}

	for _, user := range g.users {
		notificationCount := rand.Intn(15) + 8 // 8-23 notifications per user

		for i := 0; i < notificationCount; i++ {
			actor := g.users[rand.Intn(len(g.users))]
			if actor.ID == user.ID {
				continue
			}

			notifType := notificationTypes[rand.Intn(len(notificationTypes))]
			var targetID *primitive.ObjectID

			// Set appropriate target based on type
			switch notifType {
			case models.NotificationLike, models.NotificationComment, models.NotificationPostShare:
				if len(g.posts) > 0 {
					post := g.posts[rand.Intn(len(g.posts))]
					targetID = &post.ID
				}
			case models.NotificationStoryView:
				if len(g.stories) > 0 {
					story := g.stories[rand.Intn(len(g.stories))]
					targetID = &story.ID
				}
			case models.NotificationGroupInvite:
				if len(g.groups) > 0 {
					group := g.groups[rand.Intn(len(g.groups))]
					targetID = &group.ID
				}
			}

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
				TargetID:    targetID,
				IsRead:      rand.Float64() < 0.65, // 65% read
				Priority:    randomPriority(),
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

func (g *DataGenerator) generateReports(ctx context.Context, genConfig GenerationConfig) error {
	collection := g.db.Collection("reports")
	reports := make([]interface{}, 0)

	reportReasons := []models.ReportReason{
		models.ReportSpam,
		models.ReportHarassment,
		models.ReportHateSpeech,
		models.ReportViolence,
		models.ReportNudity,
		models.ReportFakeNews,
		models.ReportOther,
	}

	// Generate reports for some posts
	reportCount := len(g.posts) / 50 // About 2% of posts get reported

	for i := 0; i < reportCount; i++ {
		post := g.posts[rand.Intn(len(g.posts))]
		reporter := g.users[rand.Intn(len(g.users))]

		if reporter.ID == post.UserID {
			continue
		}

		report := models.Report{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(post.CreatedAt, time.Now()),
			},
			ReporterID:  reporter.ID,
			TargetType:  "post",
			TargetID:    post.ID,
			Reason:      reportReasons[rand.Intn(len(reportReasons))],
			Description: gofakeit.Sentence(rand.Intn(8) + 3),
			Status:      randomReportStatus(),
			Priority:    randomPriority(),
		}

		report.BeforeCreate()
		reports = append(reports, report)
	}

	// Generate reports for some users
	userReportCount := len(g.users) / 100 // About 1% of users get reported

	for i := 0; i < userReportCount; i++ {
		reportedUser := g.users[rand.Intn(len(g.users))]
		reporter := g.users[rand.Intn(len(g.users))]

		if reporter.ID == reportedUser.ID {
			continue
		}

		report := models.Report{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: gofakeit.DateRange(time.Now().AddDate(0, -3, 0), time.Now()),
			},
			ReporterID:  reporter.ID,
			TargetType:  "user",
			TargetID:    reportedUser.ID,
			Reason:      reportReasons[rand.Intn(len(reportReasons))],
			Description: gofakeit.Sentence(rand.Intn(8) + 3),
			Status:      randomReportStatus(),
			Priority:    randomPriority(),
		}

		report.BeforeCreate()
		reports = append(reports, report)
	}

	if len(reports) > 0 {
		if _, err := collection.InsertMany(ctx, reports); err != nil {
			return fmt.Errorf("failed to insert reports: %w", err)
		}
	}

	return nil
}

func (g *DataGenerator) generateUserBlocks(ctx context.Context, genConfig GenerationConfig) error {
	// Update user documents with blocked users
	for _, user := range g.users {
		if rand.Float64() < 0.1 { // 10% of users have blocked someone
			blockCount := rand.Intn(3) + 1 // 1-3 blocked users
			blockedUsers := make([]primitive.ObjectID, 0, blockCount)

			for i := 0; i < blockCount && len(blockedUsers) < len(g.users)-1; i++ {
				blockedUser := g.users[rand.Intn(len(g.users))]
				if blockedUser.ID != user.ID {
					blockedUsers = append(blockedUsers, blockedUser.ID)
				}
			}

			if len(blockedUsers) > 0 {
				_, err := g.db.Collection("users").UpdateOne(
					ctx,
					bson.M{"_id": user.ID},
					bson.M{"$set": bson.M{"blocked_users": blockedUsers}},
				)
				if err != nil {
					log.Printf("Failed to update blocked users for %s: %v", user.Username, err)
				}
			}
		}
	}

	return nil
}

func (g *DataGenerator) updateAllStatistics(ctx context.Context) error {
	// Update user statistics
	if err := g.updateUserStatistics(ctx); err != nil {
		return err
	}

	// Update post statistics
	if err := g.updatePostStatistics(ctx); err != nil {
		return err
	}

	// Update group statistics
	if err := g.updateGroupStatistics(ctx); err != nil {
		return err
	}

	return nil
}

func (g *DataGenerator) updateUserStatistics(ctx context.Context) error {
	// Update follower counts
	pipeline := []bson.M{
		{"$match": bson.M{"status": "accepted"}},
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
	pipeline[1]["$group"].(bson.M)["_id"] = "$follower_id"
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

	// Update users with all statistics
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

func (g *DataGenerator) updatePostStatistics(ctx context.Context) error {
	// Update likes counts
	pipeline := []bson.M{
		{"$match": bson.M{"target_type": "post"}},
		{"$group": bson.M{
			"_id":         "$target_id",
			"likes_count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := g.db.Collection("likes").Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	likeCounts := make(map[primitive.ObjectID]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID         primitive.ObjectID `bson:"_id"`
			LikesCount int64              `bson:"likes_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		likeCounts[result.ID] = result.LikesCount
	}

	// Update comment counts
	pipeline = []bson.M{
		{"$group": bson.M{
			"_id":            "$post_id",
			"comments_count": bson.M{"$sum": 1},
		}},
	}

	cursor, err = g.db.Collection("comments").Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	commentCounts := make(map[primitive.ObjectID]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID            primitive.ObjectID `bson:"_id"`
			CommentsCount int64              `bson:"comments_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		commentCounts[result.ID] = result.CommentsCount
	}

	// Update posts with statistics
	for _, post := range g.posts {
		update := bson.M{
			"$set": bson.M{
				"likes_count":    likeCounts[post.ID],
				"comments_count": commentCounts[post.ID],
				"views_count":    rand.Int63n(1000) + int64(likeCounts[post.ID]*3),
				"updated_at":     time.Now(),
			},
		}

		_, err := g.db.Collection("posts").UpdateOne(
			ctx,
			bson.M{"_id": post.ID},
			update,
		)
		if err != nil {
			log.Printf("Failed to update post statistics: %v", err)
		}
	}

	return nil
}

func (g *DataGenerator) updateGroupStatistics(ctx context.Context) error {
	// Update member counts for groups
	pipeline := []bson.M{
		{"$match": bson.M{"status": "active"}},
		{"$group": bson.M{
			"_id":           "$group_id",
			"members_count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := g.db.Collection("group_members").Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	memberCounts := make(map[primitive.ObjectID]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID           primitive.ObjectID `bson:"_id"`
			MembersCount int64              `bson:"members_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		memberCounts[result.ID] = result.MembersCount
	}

	// Update groups with member counts
	for _, group := range g.groups {
		update := bson.M{
			"$set": bson.M{
				"members_count": memberCounts[group.ID],
				"updated_at":    time.Now(),
			},
		}

		_, err := g.db.Collection("groups").UpdateOne(
			ctx,
			bson.M{"_id": group.ID},
			update,
		)
		if err != nil {
			log.Printf("Failed to update group statistics: %v", err)
		}
	}

	return nil
}

func (g *DataGenerator) createAdminAndTestUsers(ctx context.Context) error {
	users := []models.User{}

	// Create admin user
	hashedPassword, _ := utils.HashPassword("admin123")
	admin := models.User{
		BaseModel: models.BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
		},
		Username:      "admin",
		Email:         "admin@example.com",
		Password:      hashedPassword,
		FirstName:     "Admin",
		LastName:      "User",
		DisplayName:   "Admin User",
		Bio:           "System Administrator",
		IsVerified:    true,
		IsActive:      true,
		EmailVerified: true,
		Language:      "en",
		Timezone:      "UTC",
		OnlineStatus:  "online",
	}
	admin.BeforeCreate()
	// Set role AFTER BeforeCreate to avoid override
	admin.Role = models.RoleAdmin
	users = append(users, admin)

	// Create test users
	testUserData := []struct {
		username, email, firstName, lastName, bio string
		role                                      models.UserRole
		isVerified                                bool
	}{
		{"testuser", "test@example.com", "Test", "User", "Test account for development", models.RoleUser, false},
		{"moderator", "mod@example.com", "Moderator", "User", "Content moderator", models.RoleModerator, true},
		{"premium", "premium@example.com", "Premium", "User", "Premium account holder", models.RoleUser, true},
		{"creator", "creator@example.com", "Content", "Creator", "Content creator with many followers", models.RoleUser, true},
	}

	for _, userData := range testUserData {
		hashedPassword, _ := utils.HashPassword("password123")
		user := models.User{
			BaseModel: models.BaseModel{
				ID:        primitive.NewObjectID(),
				CreatedAt: time.Now(),
			},
			Username:      userData.username,
			Email:         userData.email,
			Password:      hashedPassword,
			FirstName:     userData.firstName,
			LastName:      userData.lastName,
			DisplayName:   userData.firstName + " " + userData.lastName,
			Bio:           userData.bio,
			IsVerified:    userData.isVerified,
			IsActive:      true,
			Role:          userData.role,
			EmailVerified: true,
			Language:      "en",
			Timezone:      "UTC",
			OnlineStatus:  "online",
			IsPremium:     userData.username == "premium",
		}

		if userData.username == "creator" {
			user.FollowersCount = 1000
			user.PostsCount = 50
			user.IsVerified = true
		}

		user.BeforeCreate()
		users = append(users, user)
	}

	// Insert all special users
	if len(users) > 0 {
		userInterfaces := make([]interface{}, len(users))
		for i, user := range users {
			userInterfaces[i] = user
		}

		_, err := g.db.Collection("users").InsertMany(ctx, userInterfaces)
		return err
	}

	return nil
}

// Helper functions
func maxTime(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

func generateUniqueSlug(usedSlugs map[string]bool, index int) string {
	// Generate base slug from fake company/word
	baseSlug := strings.ToLower(strings.ReplaceAll(gofakeit.Company(), " ", "-"))
	baseSlug = strings.ReplaceAll(baseSlug, "'", "")
	baseSlug = strings.ReplaceAll(baseSlug, ".", "")
	baseSlug = strings.ReplaceAll(baseSlug, ",", "")

	// Remove any non-alphanumeric characters except hyphens
	baseSlug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, baseSlug)

	// Ensure it starts with a letter
	if len(baseSlug) > 0 && (baseSlug[0] < 'a' || baseSlug[0] > 'z') {
		baseSlug = "group-" + baseSlug
	}

	// If empty or too short, use fallback
	if len(baseSlug) < 3 {
		baseSlug = fmt.Sprintf("group-%d", index+1)
	}

	// Check if unique, if not add number
	originalSlug := baseSlug
	counter := 1
	for usedSlugs[baseSlug] {
		baseSlug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return baseSlug
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
	weights := []float64{0.7, 0.2, 0.1} // 70% public, 20% friends, 10% private

	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return visibilities[i]
		}
	}
	return visibilities[0]
}

func randomGroupPrivacy() models.GroupPrivacy {
	privacies := []models.GroupPrivacy{
		models.GroupPublic,
		models.GroupPrivate,
		models.GroupSecret,
	}
	weights := []float64{0.6, 0.3, 0.1}

	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return privacies[i]
		}
	}
	return privacies[0]
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
	weights := []float64{0.4, 0.5, 0.1}

	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return sources[i]
		}
	}
	return sources[0]
}

func randomHashtagCategory() string {
	categories := []string{
		"general", "technology", "entertainment", "sports", "business",
		"lifestyle", "travel", "food", "fashion", "music", "art",
	}
	return categories[rand.Intn(len(categories))]
}

func randomMediaCategory() string {
	categories := []string{
		"profile", "post", "story", "message", "group", "event", "general",
	}
	return categories[rand.Intn(len(categories))]
}

func randomFollowCategories() []string {
	categories := []string{
		"close_friends", "family", "work", "school", "interests",
	}

	count := rand.Intn(3)
	if count == 0 {
		return nil
	}

	selected := make([]string, 0, count)
	for i := 0; i < count; i++ {
		selected = append(selected, categories[rand.Intn(len(categories))])
	}
	return selected
}

func randomPriority() string {
	priorities := []string{"low", "medium", "high"}
	weights := []float64{0.3, 0.6, 0.1}

	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return priorities[i]
		}
	}
	return priorities[0]
}

func randomReportStatus() models.ReportStatus {
	statuses := []models.ReportStatus{
		models.ReportPending,
		models.ReportReviewing,
		models.ReportResolved,
		models.ReportRejected,
	}
	weights := []float64{0.4, 0.2, 0.3, 0.1}

	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return statuses[i]
		}
	}
	return statuses[0]
}

func randomDeviceType() string {
	types := []string{"mobile", "desktop", "tablet"}
	weights := []float64{0.7, 0.2, 0.1}

	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return types[i]
		}
	}
	return types[0]
}

func randomColor() string {
	colors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57",
		"#FF9FF3", "#54A0FF", "#5F27CD", "#00D2D3", "#FF9F43",
		"#FFFFFF", "#000000", "#FF3838", "#2ED573", "#3742FA",
	}
	return colors[rand.Intn(len(colors))]
}

func randomFontFamily() string {
	fonts := []string{
		"Arial", "Helvetica", "Georgia", "Times", "Verdana",
		"Roboto", "Open Sans", "Lato", "Montserrat", "Poppins",
	}
	return fonts[rand.Intn(len(fonts))]
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
		platforms := []string{"twitter", "instagram", "linkedin", "github", "youtube"}

		for _, platform := range platforms {
			if rand.Float64() < 0.4 {
				links[platform] = fmt.Sprintf("https://%s.com/%s", platform, gofakeit.Username())
			}
		}

		if len(links) > 0 {
			return links
		}
	}
	return nil
}

func generatePostContent(contentType models.ContentType) string {
	switch contentType {
	case models.ContentTypeImage:
		return gofakeit.Sentence(rand.Intn(12)+3) + " 📸 #photography"
	case models.ContentTypeVideo:
		return gofakeit.Sentence(rand.Intn(10)+2) + " 🎥 #video"
	case models.ContentTypeLink:
		return gofakeit.Sentence(rand.Intn(8)+2) + " " + gofakeit.URL()
	case models.ContentTypePoll:
		return "What do you think? Vote below! 🗳️"
	default:
		sentences := rand.Intn(4) + 1
		return gofakeit.Paragraph(sentences, sentences+2, 20, " ")
	}
}

func generateMessageContent(contentType models.ContentType) string {
	switch contentType {
	case models.ContentTypeImage:
		return "📸 Photo"
	case models.ContentTypeVideo:
		return "🎥 Video"
	case models.ContentTypeAudio:
		return "🎵 Voice message"
	case models.ContentTypeFile:
		return "📎 " + gofakeit.Word() + ".pdf"
	default:
		return gofakeit.Sentence(rand.Intn(15) + 2)
	}
}

func generateRepostComment() string {
	comments := []string{
		"This! 👏",
		"So true!",
		"Couldn't agree more",
		"Sharing this with everyone",
		"This is important",
		"Great point!",
		"Love this perspective",
		"Worth sharing",
		"Exactly what I was thinking",
		"This needs to be seen",
	}
	return comments[rand.Intn(len(comments))]
}

func generateGroupName() string {
	prefixes := []string{
		"Amazing", "Awesome", "Creative", "Digital", "Elite", "Future",
		"Global", "Happy", "Innovative", "Modern", "Professional", "Smart",
	}

	subjects := []string{
		"Developers", "Designers", "Entrepreneurs", "Creators", "Writers",
		"Artists", "Photographers", "Travelers", "Foodies", "Gamers",
		"Fitness", "Music", "Technology", "Business", "Learning",
	}

	suffixes := []string{
		"Community", "Group", "Club", "Network", "Hub", "Society",
		"Alliance", "Circle", "Forum", "Space",
	}

	return fmt.Sprintf("%s %s %s",
		prefixes[rand.Intn(len(prefixes))],
		subjects[rand.Intn(len(subjects))],
		suffixes[rand.Intn(len(suffixes))],
	)
}

func generateGroupConversationTitle() string {
	topics := []string{
		"Project Team", "Study Group", "Friends Chat", "Work Team",
		"Planning Committee", "Creative Crew", "Support Group", "Book Club",
	}
	return topics[rand.Intn(len(topics))]
}

func generateHighlightTitle() string {
	titles := []string{
		"Travel", "Food", "Work", "Friends", "Family", "Moments",
		"Adventures", "Memories", "Life", "Fun", "Events", "Special",
	}
	return titles[rand.Intn(len(titles))]
}

func selectRandomHashtags(hashtags []models.Hashtag) []string {
	if len(hashtags) == 0 {
		return generateFallbackHashtags()
	}

	count := rand.Intn(5) + 1 // 1-5 hashtags
	selected := make([]string, 0, count)

	for i := 0; i < count && i < len(hashtags); i++ {
		hashtag := hashtags[rand.Intn(len(hashtags))]
		selected = append(selected, hashtag.Tag)
	}

	return selected
}

func generateFallbackHashtags() []string {
	fallback := []string{
		"life", "happy", "love", "fun", "amazing", "beautiful",
		"friends", "family", "work", "travel", "food", "music",
	}

	count := rand.Intn(3) + 1
	selected := make([]string, 0, count)

	for i := 0; i < count; i++ {
		selected = append(selected, fallback[rand.Intn(len(fallback))])
	}

	return selected
}

func generatePollOptions() []models.PollOption {
	pollTemplates := [][]string{
		{"Yes", "No"},
		{"Love it", "Like it", "Not sure", "Don't like it"},
		{"Option A", "Option B", "Option C"},
		{"Agree", "Disagree", "Neutral"},
		{"Today", "Tomorrow", "Next week", "Later"},
	}

	template := pollTemplates[rand.Intn(len(pollTemplates))]
	options := make([]models.PollOption, len(template))

	for i, text := range template {
		options[i] = models.PollOption{
			ID:         primitive.NewObjectID(),
			Text:       text,
			VotesCount: int64(rand.Intn(50)),
		}
	}

	return options
}

func generateMediaInfo(contentType models.ContentType) []models.MediaInfo {
	media := models.MediaInfo{
		Type: string(contentType),
		Size: int64(rand.Intn(10000000) + 100000), // 100KB - 10MB
	}

	switch contentType {
	case models.ContentTypeImage:
		media.URL = gofakeit.ImageURL(1200, 800)
		media.Width = 1200
		media.Height = 800
		media.Thumbnail = gofakeit.ImageURL(300, 200)
		media.AltText = gofakeit.Sentence(3)
	case models.ContentTypeVideo:
		media.URL = "https://example.com/video/" + gofakeit.UUID() + ".mp4"
		media.Width = 1920
		media.Height = 1080
		media.Duration = rand.Intn(600) + 30 // 30s - 10min
		media.Thumbnail = gofakeit.ImageURL(480, 270)
	case models.ContentTypeAudio:
		media.URL = "https://example.com/audio/" + gofakeit.UUID() + ".mp3"
		media.Duration = rand.Intn(300) + 10 // 10s - 5min
	case models.ContentTypeFile:
		media.URL = "https://example.com/files/" + gofakeit.UUID() + ".pdf"
	}

	return []models.MediaInfo{media}
}

func generateMediaURL(mediaType string) string {
	switch mediaType {
	case "image":
		return gofakeit.ImageURL(800, 600)
	case "video":
		return "https://example.com/video/" + gofakeit.UUID() + ".mp4"
	case "audio":
		return "https://example.com/audio/" + gofakeit.UUID() + ".mp3"
	default:
		return "https://example.com/file/" + gofakeit.UUID()
	}
}

func getFileExtension(mediaType string) string {
	switch mediaType {
	case "image":
		extensions := []string{".jpg", ".png", ".gif", ".webp"}
		return extensions[rand.Intn(len(extensions))]
	case "video":
		extensions := []string{".mp4", ".mov", ".avi", ".webm"}
		return extensions[rand.Intn(len(extensions))]
	case "audio":
		extensions := []string{".mp3", ".wav", ".ogg", ".m4a"}
		return extensions[rand.Intn(len(extensions))]
	case "document":
		extensions := []string{".pdf", ".doc", ".docx", ".txt", ".xlsx"}
		return extensions[rand.Intn(len(extensions))]
	default:
		return ".bin"
	}
}

func getMimeType(mediaType string) string {
	switch mediaType {
	case "image":
		types := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
		return types[rand.Intn(len(types))]
	case "video":
		types := []string{"video/mp4", "video/quicktime", "video/avi", "video/webm"}
		return types[rand.Intn(len(types))]
	case "audio":
		types := []string{"audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4"}
		return types[rand.Intn(len(types))]
	case "document":
		types := []string{"application/pdf", "application/msword", "text/plain", "application/vnd.ms-excel"}
		return types[rand.Intn(len(types))]
	default:
		return "application/octet-stream"
	}
}

func generateNotificationTitle(notifType models.NotificationType) string {
	switch notifType {
	case models.NotificationLike:
		return "New Like"
	case models.NotificationLove:
		return "Someone loved your post"
	case models.NotificationComment:
		return "New Comment"
	case models.NotificationFollow:
		return "New Follower"
	case models.NotificationMention:
		return "You were mentioned"
	case models.NotificationMessage:
		return "New Message"
	case models.NotificationGroupInvite:
		return "Group Invitation"
	case models.NotificationEventInvite:
		return "Event Invitation"
	case models.NotificationPostShare:
		return "Post Shared"
	case models.NotificationStoryView:
		return "Story View"
	default:
		return "Notification"
	}
}

func generateNotificationMessage(notifType models.NotificationType) string {
	switch notifType {
	case models.NotificationLike:
		return "Someone liked your post"
	case models.NotificationLove:
		return "Someone loved your content"
	case models.NotificationComment:
		return "Someone commented on your post"
	case models.NotificationFollow:
		return "Someone started following you"
	case models.NotificationMention:
		return "Someone mentioned you in a post"
	case models.NotificationMessage:
		return "You have a new message"
	case models.NotificationGroupInvite:
		return "You were invited to join a group"
	case models.NotificationEventInvite:
		return "You were invited to an event"
	case models.NotificationPostShare:
		return "Someone shared your post"
	case models.NotificationStoryView:
		return "Someone viewed your story"
	default:
		return "You have a new notification"
	}
}

func printSummary(generator *DataGenerator, genConfig GenerationConfig, duration time.Duration) {
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║                    GENERATION COMPLETE                       ║
╠══════════════════════════════════════════════════════════════╣
║  👥 Users:              %-8d                        ║
║  📝 Posts:              %-8d                        ║
║  💬 Comments:           %-8d                        ║
║  📱 Stories:            %-8d                        ║
║  👥 Groups:             %-8d                        ║
║  💬 Conversations:      %-8d                        ║
║  🏷️  Hashtags:          %-8d                        ║
║  📸 Media Files:        %-8d                        ║
║  ⏱️  Time Taken:         %-8s                        ║
║                                                              ║
║  ✅ Complete synchronized social media ecosystem generated!  ║
║  🔗 All relationships and interactions properly linked       ║
║  📊 Statistics calculated and updated                       ║
╚══════════════════════════════════════════════════════════════╝

🎉 Synchronized data generation completed successfully!

Admin Credentials:
  Username: admin
  Email:    admin@example.com
  Password: admin123

Test User Credentials:
  Username: testuser, moderator, premium, creator
  Email:    test@example.com, mod@example.com, etc.
  Password: password123

Sample User Credentials:
  Username: user1, user2, user3, ... user%d
  Email:    user1@example.com, user2@example.com, etc.
  Password: password123

Your Social Media API is now ready with a fully synchronized ecosystem! 🚀

Generated Features:
✅ Users with profiles, settings, and relationships
✅ Posts with proper user references and media
✅ Comments with valid user and post references
✅ Stories with correct user associations
✅ Groups with proper member relationships
✅ Direct and group conversations with messages
✅ Follow relationships between actual users
✅ Likes, reactions, and engagement with proper refs
✅ Notifications for all interaction types
✅ User mentions across all content types
✅ Hashtag tracking and trending
✅ Media management with user associations
✅ Content reporting and moderation
✅ User blocking and privacy controls
✅ Comprehensive statistics and analytics

All data is properly synchronized and referenced! 
No more "unknown user" issues! 🎯

`, len(generator.users), len(generator.posts), len(generator.comments),
		len(generator.stories), len(generator.groups),
		len(generator.conversations), len(generator.hashtags), len(generator.media),
		duration.Round(time.Second), genConfig.UserCount)
}
