// internal/services/feed_service.go
package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedService struct {
	postCollection        *mongo.Collection
	userCollection        *mongo.Collection
	followCollection      *mongo.Collection
	interactionCollection *mongo.Collection
	feedCacheCollection   *mongo.Collection
	db                    *mongo.Database
}

type FeedItem struct {
	Post          models.Post    `json:"post" bson:"post"`
	Score         float64        `json:"score" bson:"score"`
	Reason        string         `json:"reason" bson:"reason"` // "following", "suggested", "trending", etc.
	TimeAgo       string         `json:"time_ago" bson:"time_ago"`
	IsPromoted    bool           `json:"is_promoted" bson:"is_promoted"`
	PromotionInfo *PromotionInfo `json:"promotion_info,omitempty" bson:"promotion_info,omitempty"`
}

type PromotionInfo struct {
	Type         string    `json:"type" bson:"type"` // "sponsored", "boosted"
	Advertiser   string    `json:"advertiser" bson:"advertiser"`
	TargetAge    []int     `json:"target_age" bson:"target_age"`
	TargetGender string    `json:"target_gender" bson:"target_gender"`
	Budget       float64   `json:"budget" bson:"budget"`
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
}

type UserInteraction struct {
	models.BaseModel `bson:",inline"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	PostID           primitive.ObjectID `json:"post_id" bson:"post_id"`
	InteractionType  string             `json:"interaction_type" bson:"interaction_type"` // "view", "like", "comment", "share", "save"
	InteractionScore float64            `json:"interaction_score" bson:"interaction_score"`
	TimeSpent        int64              `json:"time_spent" bson:"time_spent"` // in seconds
	Source           string             `json:"source" bson:"source"`         // "feed", "profile", "search", etc.
}

type FeedCache struct {
	models.BaseModel `bson:",inline"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	FeedType         string             `json:"feed_type" bson:"feed_type"` // "home", "trending", "following"
	Posts            []FeedItem         `json:"posts" bson:"posts"`
	LastRefreshed    time.Time          `json:"last_refreshed" bson:"last_refreshed"`
	ExpiresAt        time.Time          `json:"expires_at" bson:"expires_at"`
}

type FeedAlgorithmWeights struct {
	RecencyWeight      float64 `json:"recency_weight"`
	EngagementWeight   float64 `json:"engagement_weight"`
	RelationshipWeight float64 `json:"relationship_weight"`
	ContentTypeWeight  float64 `json:"content_type_weight"`
	UserInterestWeight float64 `json:"user_interest_weight"`
	DiversityWeight    float64 `json:"diversity_weight"`
}

func NewFeedService() *FeedService {
	return &FeedService{
		postCollection:        config.DB.Collection("posts"),
		userCollection:        config.DB.Collection("users"),
		followCollection:      config.DB.Collection("follows"),
		interactionCollection: config.DB.Collection("user_interactions"),
		feedCacheCollection:   config.DB.Collection("feed_cache"),
		db:                    config.DB,
	}
}

// GetUserFeed generates and returns personalized feed for a user
func (fs *FeedService) GetUserFeed(userID primitive.ObjectID, feedType string, limit, skip int, refresh bool) ([]FeedItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check cache first if not forcing refresh
	if !refresh {
		cachedFeed, err := fs.getCachedFeed(ctx, userID, feedType)
		if err == nil && cachedFeed != nil && !fs.isCacheExpired(cachedFeed) {
			start := skip
			end := skip + limit
			if end > len(cachedFeed.Posts) {
				end = len(cachedFeed.Posts)
			}

			if start < len(cachedFeed.Posts) {
				return cachedFeed.Posts[start:end], nil
			}
		}
	}

	// Generate fresh feed
	var feedItems []FeedItem
	var err error

	switch feedType {
	case "home", "personal":
		feedItems, err = fs.generatePersonalizedFeed(ctx, userID, limit*3) // Get more for better selection
	case "following":
		feedItems, err = fs.generateFollowingFeed(ctx, userID, limit*2)
	case "trending":
		feedItems, err = fs.generateTrendingFeed(ctx, userID, limit*2)
	case "discover":
		feedItems, err = fs.generateDiscoverFeed(ctx, userID, limit*2)
	default:
		feedItems, err = fs.generatePersonalizedFeed(ctx, userID, limit*2)
	}

	if err != nil {
		return nil, err
	}

	// Apply diversity and ranking
	rankedFeed := fs.applyFinalRanking(feedItems, userID)

	// Cache the feed
	go fs.cacheFeed(userID, feedType, rankedFeed)

	// Return requested page
	start := skip
	end := skip + limit
	if end > len(rankedFeed) {
		end = len(rankedFeed)
	}

	if start < len(rankedFeed) {
		return rankedFeed[start:end], nil
	}

	return []FeedItem{}, nil
}

// generatePersonalizedFeed creates a personalized feed using ML-like algorithm
func (fs *FeedService) generatePersonalizedFeed(ctx context.Context, userID primitive.ObjectID, limit int) ([]FeedItem, error) {
	weights := FeedAlgorithmWeights{
		RecencyWeight:      0.3,
		EngagementWeight:   0.25,
		RelationshipWeight: 0.2,
		ContentTypeWeight:  0.1,
		UserInterestWeight: 0.1,
		DiversityWeight:    0.05,
	}

	// Get user's following list
	following, err := fs.getUserFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user interests and interaction history
	userInterests, err := fs.getUserInterests(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create aggregation pipeline for scoring posts
	pipeline := []bson.M{
		// Match eligible posts
		{
			"$match": bson.M{
				"is_published": true,
				"deleted_at":   bson.M{"$exists": false},
				"created_at":   bson.M{"$gte": time.Now().Add(-7 * 24 * time.Hour)}, // Last 7 days
				"$or": []bson.M{
					{"visibility": "public"},
					{
						"$and": []bson.M{
							{"visibility": "friends"},
							{"user_id": bson.M{"$in": following}},
						},
					},
					{"user_id": userID}, // User's own posts
				},
			},
		},
		// Lookup author information
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "author",
			},
		},
		{
			"$unwind": "$author",
		},
		// Calculate engagement score
		{
			"$addFields": bson.M{
				"engagement_score": bson.M{
					"$add": []interface{}{
						"$likes_count",
						bson.M{"$multiply": []interface{}{"$comments_count", 2}},
						bson.M{"$multiply": []interface{}{"$shares_count", 3}},
					},
				},
				"hours_since_posted": bson.M{
					"$divide": []interface{}{
						bson.M{"$subtract": []interface{}{time.Now(), "$created_at"}},
						1000 * 60 * 60, // Convert to hours
					},
				},
			},
		},
		// Calculate recency score (higher for newer posts)
		{
			"$addFields": bson.M{
				"recency_score": bson.M{
					"$divide": []interface{}{
						1,
						bson.M{
							"$add": []interface{}{
								1,
								bson.M{"$divide": []interface{}{"$hours_since_posted", 24}},
							},
						},
					},
				},
			},
		},
		// Calculate relationship score
		{
			"$addFields": bson.M{
				"relationship_score": bson.M{
					"$cond": []interface{}{
						bson.M{"$in": []interface{}{"$user_id", following}},
						1.0,
						bson.M{
							"$cond": []interface{}{
								bson.M{"$eq": []interface{}{"$user_id", userID}},
								0.8, // Own posts get high but not highest score
								0.3, // Public posts from non-following users
							},
						},
					},
				},
			},
		},
		// Calculate final score
		{
			"$addFields": bson.M{
				"final_score": bson.M{
					"$add": []interface{}{
						bson.M{"$multiply": []interface{}{"$recency_score", weights.RecencyWeight}},
						bson.M{"$multiply": []interface{}{
							bson.M{"$divide": []interface{}{"$engagement_score", 100}}, // Normalize
							weights.EngagementWeight,
						}},
						bson.M{"$multiply": []interface{}{"$relationship_score", weights.RelationshipWeight}},
					},
				},
			},
		},
		// Sort by score
		{
			"$sort": bson.M{
				"final_score": -1,
				"created_at":  -1,
			},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := fs.postCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Post `bson:",inline"`
		FinalScore  float64     `bson:"final_score"`
		Author      models.User `bson:"author"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to FeedItems
	var feedItems []FeedItem
	for _, result := range results {
		// Set author information
		result.Post.Author = result.Author.ToUserResponse()

		feedItem := FeedItem{
			Post:       result.Post,
			Score:      result.FinalScore,
			Reason:     fs.determineFeedReason(result.Post.UserID, userID, following),
			TimeAgo:    fs.calculateTimeAgo(result.Post.CreatedAt),
			IsPromoted: result.Post.IsPromoted,
		}

		feedItems = append(feedItems, feedItem)
	}

	// Apply interest-based filtering and boosting
	feedItems = fs.applyInterestFiltering(feedItems, userInterests)

	return feedItems, nil
}

// generateFollowingFeed creates feed from followed users only
func (fs *FeedService) generateFollowingFeed(ctx context.Context, userID primitive.ObjectID, limit int) ([]FeedItem, error) {
	following, err := fs.getUserFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(following) == 0 {
		return []FeedItem{}, nil
	}

	filter := bson.M{
		"user_id":      bson.M{"$in": append(following, userID)}, // Include user's own posts
		"is_published": true,
		"deleted_at":   bson.M{"$exists": false},
		"created_at":   bson.M{"$gte": time.Now().Add(-3 * 24 * time.Hour)}, // Last 3 days
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := fs.postCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Convert to feed items and populate author info
	var feedItems []FeedItem
	for _, post := range posts {
		// Populate author info
		fs.populatePostAuthor(ctx, &post)

		feedItem := FeedItem{
			Post:    post,
			Score:   fs.calculateEngagementScore(post),
			Reason:  "following",
			TimeAgo: fs.calculateTimeAgo(post.CreatedAt),
		}

		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nil
}

// generateTrendingFeed creates feed of trending content
func (fs *FeedService) generateTrendingFeed(ctx context.Context, userID primitive.ObjectID, limit int) ([]FeedItem, error) {
	// Get posts with high engagement in last 24 hours
	timeThreshold := time.Now().Add(-24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_published": true,
				"visibility":   "public",
				"deleted_at":   bson.M{"$exists": false},
				"created_at":   bson.M{"$gte": timeThreshold},
			},
		},
		{
			"$addFields": bson.M{
				"trending_score": bson.M{
					"$add": []interface{}{
						bson.M{"$multiply": []interface{}{"$likes_count", 1}},
						bson.M{"$multiply": []interface{}{"$comments_count", 2}},
						bson.M{"$multiply": []interface{}{"$shares_count", 3}},
						bson.M{"$multiply": []interface{}{"$views_count", 0.1}},
					},
				},
			},
		},
		{
			"$match": bson.M{
				"trending_score": bson.M{"$gte": 10}, // Minimum engagement threshold
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "author",
			},
		},
		{
			"$unwind": "$author",
		},
		{
			"$sort": bson.M{
				"trending_score": -1,
				"created_at":     -1,
			},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := fs.postCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Post   `bson:",inline"`
		TrendingScore float64     `bson:"trending_score"`
		Author        models.User `bson:"author"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var feedItems []FeedItem
	for _, result := range results {
		result.Post.Author = result.Author.ToUserResponse()

		feedItem := FeedItem{
			Post:    result.Post,
			Score:   result.TrendingScore,
			Reason:  "trending",
			TimeAgo: fs.calculateTimeAgo(result.Post.CreatedAt),
		}

		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nil
}

// generateDiscoverFeed creates discovery feed with new content
func (fs *FeedService) generateDiscoverFeed(ctx context.Context, userID primitive.ObjectID, limit int) ([]FeedItem, error) {
	// Get users that current user is NOT following
	following, _ := fs.getUserFollowing(ctx, userID)
	userInterests, _ := fs.getUserInterests(ctx, userID)

	filter := bson.M{
		"user_id":      bson.M{"$nin": append(following, userID)}, // Exclude following and self
		"is_published": true,
		"visibility":   "public",
		"deleted_at":   bson.M{"$exists": false},
		"created_at":   bson.M{"$gte": time.Now().Add(-2 * 24 * time.Hour)}, // Last 2 days
	}

	// Add hashtag filter based on user interests
	if len(userInterests) > 0 {
		filter["$or"] = []bson.M{
			{"hashtags": bson.M{"$in": userInterests}},
			{"likes_count": bson.M{"$gte": 5}}, // Or posts with decent engagement
		}
	}

	opts := options.Find().
		SetLimit(int64(limit * 2)). // Get more for better selection
		SetSort(bson.M{"engagement_rate": -1, "created_at": -1})

	cursor, err := fs.postCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	var feedItems []FeedItem
	for _, post := range posts {
		fs.populatePostAuthor(ctx, &post)

		score := fs.calculateDiscoveryScore(post, userInterests)

		feedItem := FeedItem{
			Post:    post,
			Score:   score,
			Reason:  "discover",
			TimeAgo: fs.calculateTimeAgo(post.CreatedAt),
		}

		feedItems = append(feedItems, feedItem)
	}

	// Sort by discovery score
	sort.Slice(feedItems, func(i, j int) bool {
		return feedItems[i].Score > feedItems[j].Score
	})

	// Return top items
	if len(feedItems) > limit {
		feedItems = feedItems[:limit]
	}

	return feedItems, nil
}

// RecordInteraction records user interaction with content
func (fs *FeedService) RecordInteraction(userID, postID primitive.ObjectID, interactionType, source string, timeSpent int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Calculate interaction score based on type
	var score float64
	switch interactionType {
	case "view":
		score = 1.0
	case "like":
		score = 5.0
	case "comment":
		score = 10.0
	case "share":
		score = 15.0
	case "save":
		score = 8.0
	default:
		score = 1.0
	}

	// Adjust score based on time spent
	if timeSpent > 0 {
		timeBonus := math.Min(float64(timeSpent)/30.0, 2.0) // Cap at 2x bonus for 30+ seconds
		score *= (1.0 + timeBonus)
	}

	interaction := &UserInteraction{
		UserID:           userID,
		PostID:           postID,
		InteractionType:  interactionType,
		InteractionScore: score,
		TimeSpent:        timeSpent,
		Source:           source,
	}
	interaction.BeforeCreate()

	_, err := fs.interactionCollection.InsertOne(ctx, interaction)
	if err != nil {
		return err
	}

	// Invalidate feed cache for this user
	go fs.invalidateFeedCache(userID)

	return nil
}

// RefreshUserFeed forces refresh of user's cached feed
func (fs *FeedService) RefreshUserFeed(userID primitive.ObjectID, feedType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete cached feed
	filter := bson.M{
		"user_id":   userID,
		"feed_type": feedType,
	}

	_, err := fs.feedCacheCollection.DeleteMany(ctx, filter)
	return err
}

// Helper methods

func (fs *FeedService) getUserFollowing(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	cursor, err := fs.followCollection.Find(ctx, bson.M{
		"follower_id": userID,
		"status":      "accepted",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []struct {
		FolloweeID primitive.ObjectID `bson:"followee_id"`
	}

	if err := cursor.All(ctx, &follows); err != nil {
		return nil, err
	}

	var following []primitive.ObjectID
	for _, follow := range follows {
		following = append(following, follow.FolloweeID)
	}

	return following, nil
}

func (fs *FeedService) getUserInterests(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	// Get user's most interacted hashtags
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id": userID,
				"created_at": bson.M{
					"$gte": time.Now().Add(-30 * 24 * time.Hour), // Last 30 days
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "posts",
				"localField":   "post_id",
				"foreignField": "_id",
				"as":           "post",
			},
		},
		{
			"$unwind": "$post",
		},
		{
			"$unwind": "$post.hashtags",
		},
		{
			"$group": bson.M{
				"_id":          "$post.hashtags",
				"interactions": bson.M{"$sum": "$interaction_score"},
			},
		},
		{
			"$sort": bson.M{"interactions": -1},
		},
		{
			"$limit": 10,
		},
	}

	cursor, err := fs.interactionCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return []string{}, nil // Return empty if error
	}
	defer cursor.Close(ctx)

	var results []struct {
		Hashtag string `bson:"_id"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return []string{}, nil
	}

	var interests []string
	for _, result := range results {
		interests = append(interests, result.Hashtag)
	}

	return interests, nil
}

func (fs *FeedService) populatePostAuthor(ctx context.Context, post *models.Post) {
	var user models.User
	err := fs.userCollection.FindOne(ctx, bson.M{"_id": post.UserID}).Decode(&user)
	if err == nil {
		post.Author = user.ToUserResponse()
	}
}

func (fs *FeedService) calculateEngagementScore(post models.Post) float64 {
	return float64(post.LikesCount + post.CommentsCount*2 + post.SharesCount*3)
}

func (fs *FeedService) calculateDiscoveryScore(post models.Post, userInterests []string) float64 {
	baseScore := fs.calculateEngagementScore(post)

	// Boost score if post has hashtags matching user interests
	interestBoost := 0.0
	for _, hashtag := range post.Hashtags {
		for _, interest := range userInterests {
			if hashtag == interest {
				interestBoost += 10.0
				break
			}
		}
	}

	// Age penalty (prefer newer content in discovery)
	hoursSincePosted := time.Since(post.CreatedAt).Hours()
	agePenalty := math.Max(0, hoursSincePosted-24) * 0.5

	return baseScore + interestBoost - agePenalty
}

func (fs *FeedService) determineFeedReason(postUserID, currentUserID primitive.ObjectID, following []primitive.ObjectID) string {
	if postUserID == currentUserID {
		return "your_post"
	}

	for _, followedID := range following {
		if postUserID == followedID {
			return "following"
		}
	}

	return "suggested"
}

func (fs *FeedService) calculateTimeAgo(createdAt time.Time) string {
	duration := time.Since(createdAt)

	if duration.Hours() < 1 {
		minutes := int(duration.Minutes())
		if minutes <= 1 {
			return "1m"
		}
		return fmt.Sprintf("%dm", minutes)
	} else if duration.Hours() < 24 {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1d"
		} else if days < 7 {
			return fmt.Sprintf("%dd", days)
		} else if days < 30 {
			weeks := days / 7
			return fmt.Sprintf("%dw", weeks)
		} else {
			return createdAt.Format("Jan 2")
		}
	}
}

func (fs *FeedService) applyInterestFiltering(feedItems []FeedItem, userInterests []string) []FeedItem {
	// Boost posts that match user interests
	for i := range feedItems {
		for _, hashtag := range feedItems[i].Post.Hashtags {
			for _, interest := range userInterests {
				if hashtag == interest {
					feedItems[i].Score *= 1.2 // 20% boost
					break
				}
			}
		}
	}

	return feedItems
}

func (fs *FeedService) applyFinalRanking(feedItems []FeedItem, userID primitive.ObjectID) []FeedItem {
	// Apply diversity: avoid too many posts from same author
	authorPostCount := make(map[primitive.ObjectID]int)
	var finalFeed []FeedItem

	// Sort by score first
	sort.Slice(feedItems, func(i, j int) bool {
		return feedItems[i].Score > feedItems[j].Score
	})

	for _, item := range feedItems {
		authorID := item.Post.UserID

		// Limit to 3 posts per author in feed
		if authorPostCount[authorID] < 3 {
			finalFeed = append(finalFeed, item)
			authorPostCount[authorID]++
		}
	}

	return finalFeed
}

func (fs *FeedService) getCachedFeed(ctx context.Context, userID primitive.ObjectID, feedType string) (*FeedCache, error) {
	var cache FeedCache
	err := fs.feedCacheCollection.FindOne(ctx, bson.M{
		"user_id":   userID,
		"feed_type": feedType,
	}).Decode(&cache)

	if err != nil {
		return nil, err
	}

	return &cache, nil
}

func (fs *FeedService) isCacheExpired(cache *FeedCache) bool {
	return time.Now().After(cache.ExpiresAt)
}

func (fs *FeedService) cacheFeed(userID primitive.ObjectID, feedType string, feedItems []FeedItem) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cache := &FeedCache{
		UserID:        userID,
		FeedType:      feedType,
		Posts:         feedItems,
		LastRefreshed: time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour), // Cache for 1 hour
	}
	cache.BeforeCreate()

	// Upsert cache
	filter := bson.M{
		"user_id":   userID,
		"feed_type": feedType,
	}

	opts := options.Replace().SetUpsert(true)
	fs.feedCacheCollection.ReplaceOne(ctx, filter, cache, opts)
}

func (fs *FeedService) invalidateFeedCache(userID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fs.feedCacheCollection.DeleteMany(ctx, bson.M{"user_id": userID})
}

// CleanupOldCaches removes expired feed caches
func (fs *FeedService) CleanupOldCaches() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	}

	result, err := fs.feedCacheCollection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	fmt.Printf("Cleaned up %d expired feed caches\n", result.DeletedCount)
	return nil
}
