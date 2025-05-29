// internal/services/post_service.go
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostService struct {
	collection     *mongo.Collection
	userCollection *mongo.Collection
	likeCollection *mongo.Collection
	db             *mongo.Database
}

func NewPostService() *PostService {
	return &PostService{
		collection:     config.DB.Collection("posts"),
		userCollection: config.DB.Collection("users"),
		likeCollection: config.DB.Collection("likes"),
		db:             config.DB,
	}
}

// CreatePost creates a new post
func (ps *PostService) CreatePost(userID primitive.ObjectID, req models.CreatePostRequest) (*models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert group and event IDs if provided
	var groupID, eventID *primitive.ObjectID
	if req.GroupID != "" {
		if gID, err := primitive.ObjectIDFromHex(req.GroupID); err == nil {
			groupID = &gID
		}
	}
	if req.EventID != "" {
		if eID, err := primitive.ObjectIDFromHex(req.EventID); err == nil {
			eventID = &eID
		}
	}

	// Convert mentioned user IDs
	var mentions []primitive.ObjectID
	for _, mentionStr := range req.Mentions {
		if mentionID, err := primitive.ObjectIDFromHex(mentionStr); err == nil {
			mentions = append(mentions, mentionID)
		}
	}

	// Create post
	post := &models.Post{
		UserID:          userID,
		Content:         req.Content,
		ContentType:     req.ContentType,
		Media:           req.Media,
		Type:            req.Type,
		Visibility:      req.Visibility,
		Language:        req.Language,
		Location:        req.Location,
		Hashtags:        req.Hashtags,
		Mentions:        mentions,
		CommentsEnabled: req.CommentsEnabled,
		LikesEnabled:    req.LikesEnabled,
		SharesEnabled:   req.SharesEnabled,
		GroupID:         groupID,
		EventID:         eventID,
		ScheduledFor:    req.ScheduledFor,
		PollOptions:     convertPollOptions(req.PollOptions),
		PollExpiresAt:   req.PollExpiresAt,
		PollMultiple:    req.PollMultiple,
		CustomFields:    req.CustomFields,
	}

	post.BeforeCreate()

	// Handle scheduled posts
	if req.ScheduledFor != nil && req.ScheduledFor.After(time.Now()) {
		post.IsScheduled = true
		post.IsPublished = false
		post.PublishedAt = nil
	}

	// Extract hashtags from content if not provided
	if len(post.Hashtags) == 0 {
		extractedHashtags := extractHashtagsFromText(post.Content)
		post.Hashtags = extractedHashtags
	}

	result, err := ps.collection.InsertOne(ctx, post)
	if err != nil {
		return nil, err
	}

	post.ID = result.InsertedID.(primitive.ObjectID)

	// Update user's post count if published
	if post.IsPublished {
		ps.updateUserPostCount(userID, true)
	}

	// Create hashtag entries
	if len(post.Hashtags) > 0 {
		go ps.createHashtagEntries(post.Hashtags, post.ID)
	}

	// Create mention notifications
	if len(post.Mentions) > 0 {
		go ps.createMentionNotifications(userID, post.ID, post.Mentions)
	}

	return post, nil
}

// GetPostByID retrieves a post by ID
func (ps *PostService) GetPostByID(postID primitive.ObjectID, currentUserID *primitive.ObjectID) (*models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var post models.Post
	err := ps.collection.FindOne(ctx, bson.M{
		"_id":        postID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&post)

	if err != nil {
		return nil, err
	}

	// Check if user can view this post
	if currentUserID != nil && !ps.canUserViewPost(&post, *currentUserID) {
		return nil, errors.New("access denied")
	}

	// Populate author information
	if err := ps.populatePostAuthor(&post); err != nil {
		return nil, err
	}

	// Increment view count
	if currentUserID != nil && *currentUserID != post.UserID {
		go ps.incrementViewCount(postID)
	}

	return &post, nil
}

// GetUserPosts retrieves posts by a specific user
func (ps *PostService) GetUserPosts(userID primitive.ObjectID, currentUserID *primitive.ObjectID, limit, skip int) ([]models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":      userID,
		"is_published": true,
		"deleted_at":   bson.M{"$exists": false},
	}

	// Apply privacy filter if not viewing own posts
	if currentUserID == nil || *currentUserID != userID {
		filter["visibility"] = bson.M{"$in": []string{"public", "friends"}}
		// Add additional privacy logic here based on follow relationship
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ps.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Populate author information for all posts
	for i := range posts {
		ps.populatePostAuthor(&posts[i])
	}

	return posts, nil
}

// GetFeedPosts retrieves posts for user's feed
func (ps *PostService) GetFeedPosts(userID primitive.ObjectID, limit, skip int) ([]models.PostFeedResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Complex aggregation pipeline for feed algorithm
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_published": true,
				"deleted_at":   bson.M{"$exists": false},
				"$or": []bson.M{
					{"visibility": "public"},
					{
						"$and": []bson.M{
							{"visibility": "friends"},
							// Add follow relationship check here
						},
					},
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
		// Calculate engagement score for ranking
		{
			"$addFields": bson.M{
				"engagement_score": bson.M{
					"$add": []interface{}{
						"$likes_count",
						bson.M{"$multiply": []interface{}{"$comments_count", 2}},
						bson.M{"$multiply": []interface{}{"$shares_count", 3}},
					},
				},
				"time_decay": bson.M{
					"$divide": []interface{}{
						1,
						bson.M{
							"$add": []interface{}{
								1,
								bson.M{
									"$divide": []interface{}{
										bson.M{"$subtract": []interface{}{time.Now(), "$created_at"}},
										1000 * 60 * 60, // Convert to hours
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"feed_score": bson.M{
					"$multiply": []interface{}{"$engagement_score", "$time_decay"},
				},
			},
		},
		{
			"$sort": bson.M{
				"feed_score": -1,
				"created_at": -1,
			},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := ps.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.PostFeedResponse
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

// UpdatePost updates an existing post
func (ps *PostService) UpdatePost(postID, userID primitive.ObjectID, req models.UpdatePostRequest) (*models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if post exists and user owns it
	post, err := ps.GetPostByID(postID, &userID)
	if err != nil {
		return nil, err
	}

	if post.UserID != userID {
		return nil, errors.New("access denied")
	}

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	// Update fields if provided
	if req.Content != nil {
		update["$set"].(bson.M)["content"] = *req.Content
		// Re-extract hashtags if content changed
		if req.Hashtags == nil {
			update["$set"].(bson.M)["hashtags"] = extractHashtagsFromText(*req.Content)
		}
	}
	if req.Visibility != nil {
		update["$set"].(bson.M)["visibility"] = *req.Visibility
	}
	if req.Language != nil {
		update["$set"].(bson.M)["language"] = *req.Language
	}
	if req.Location != nil {
		update["$set"].(bson.M)["location"] = *req.Location
	}
	if req.Hashtags != nil {
		update["$set"].(bson.M)["hashtags"] = req.Hashtags
	}
	if req.Mentions != nil {
		var mentions []primitive.ObjectID
		for _, mentionStr := range req.Mentions {
			if mentionID, err := primitive.ObjectIDFromHex(mentionStr); err == nil {
				mentions = append(mentions, mentionID)
			}
		}
		update["$set"].(bson.M)["mentions"] = mentions
	}
	if req.CommentsEnabled != nil {
		update["$set"].(bson.M)["comments_enabled"] = *req.CommentsEnabled
	}
	if req.LikesEnabled != nil {
		update["$set"].(bson.M)["likes_enabled"] = *req.LikesEnabled
	}
	if req.SharesEnabled != nil {
		update["$set"].(bson.M)["shares_enabled"] = *req.SharesEnabled
	}
	if req.IsPinned != nil {
		update["$set"].(bson.M)["is_pinned"] = *req.IsPinned
	}

	// Mark as edited
	update["$set"].(bson.M)["is_edited"] = true
	update["$set"].(bson.M)["edited_at"] = time.Now()

	_, err = ps.collection.UpdateOne(ctx, bson.M{"_id": postID}, update)
	if err != nil {
		return nil, err
	}

	return ps.GetPostByID(postID, &userID)
}

// DeletePost soft deletes a post
func (ps *PostService) DeletePost(postID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if post exists and user owns it
	var post models.Post
	err := ps.collection.FindOne(ctx, bson.M{
		"_id":     postID,
		"user_id": userID,
	}).Decode(&post)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("post not found or access denied")
		}
		return err
	}

	// Soft delete the post
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at":  now,
			"updated_at":  now,
			"is_hidden":   true,
			"is_approved": false,
		},
	}

	_, err = ps.collection.UpdateOne(ctx, bson.M{"_id": postID}, update)
	if err != nil {
		return err
	}

	// Update user's post count
	go ps.updateUserPostCount(userID, false)

	return nil
}

// LikePost adds or removes a like from a post
func (ps *PostService) LikePost(postID, userID primitive.ObjectID, reactionType models.ReactionType) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if post exists
	var post models.Post
	err := ps.collection.FindOne(ctx, bson.M{
		"_id":           postID,
		"is_published":  true,
		"likes_enabled": true,
		"deleted_at":    bson.M{"$exists": false},
	}).Decode(&post)

	if err != nil {
		return err
	}

	// Check if user already liked this post
	var existingLike models.Like
	err = ps.likeCollection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   postID,
		"target_type": "post",
	}).Decode(&existingLike)

	if err == nil {
		// Update existing like
		update := bson.M{
			"$set": bson.M{
				"reaction_type": reactionType,
				"updated_at":    time.Now(),
			},
		}
		_, err = ps.likeCollection.UpdateOne(ctx, bson.M{"_id": existingLike.ID}, update)
	} else if err == mongo.ErrNoDocuments {
		// Create new like
		like := &models.Like{
			UserID:       userID,
			TargetID:     postID,
			TargetType:   "post",
			ReactionType: reactionType,
		}
		like.BeforeCreate()

		_, err = ps.likeCollection.InsertOne(ctx, like)
		if err != nil {
			return err
		}

		// Increment post like count
		ps.collection.UpdateOne(ctx, bson.M{"_id": postID}, bson.M{
			"$inc": bson.M{"likes_count": 1},
		})

		// Update user's total likes received
		go ps.updateUserLikesCount(post.UserID, true)
	}

	return err
}

// UnlikePost removes a like from a post
func (ps *PostService) UnlikePost(postID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find and delete the like
	result, err := ps.likeCollection.DeleteOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   postID,
		"target_type": "post",
	})

	if err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		// Decrement post like count
		ps.collection.UpdateOne(ctx, bson.M{"_id": postID}, bson.M{
			"$inc": bson.M{"likes_count": -1},
		})

		// Get post owner for updating their likes count
		var post models.Post
		if err := ps.collection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post); err == nil {
			go ps.updateUserLikesCount(post.UserID, false)
		}
	}

	return nil
}

// ReportPost reports a post
func (ps *PostService) ReportPost(postID, reporterID primitive.ObjectID, reason models.ReportReason, description string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if post exists
	var post models.Post
	err := ps.collection.FindOne(ctx, bson.M{
		"_id":        postID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&post)

	if err != nil {
		return err
	}

	// Create report
	report := &models.Report{
		ReporterID:  reporterID,
		TargetType:  "post",
		TargetID:    postID,
		Reason:      reason,
		Description: description,
	}
	report.BeforeCreate()

	reportCollection := ps.db.Collection("reports")
	_, err = reportCollection.InsertOne(ctx, report)
	if err != nil {
		return err
	}

	// Update post report count
	_, err = ps.collection.UpdateOne(ctx, bson.M{"_id": postID}, bson.M{
		"$inc": bson.M{"reports_count": 1},
		"$set": bson.M{"is_reported": true},
	})

	return err
}

// GetPostLikes retrieves users who liked a post
func (ps *PostService) GetPostLikes(postID primitive.ObjectID, limit, skip int) ([]models.LikeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id":   postID,
				"target_type": "post",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$unwind": "$user",
		},
		{
			"$sort": bson.M{"created_at": -1},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := ps.likeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var likes []models.LikeResponse
	if err := cursor.All(ctx, &likes); err != nil {
		return nil, err
	}

	return likes, nil
}

// GetPostStats retrieves post statistics
func (ps *PostService) GetPostStats(postID primitive.ObjectID) (*models.PostStatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var post models.Post
	err := ps.collection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
	if err != nil {
		return nil, err
	}

	return &models.PostStatsResponse{
		PostID:          post.ID.Hex(),
		LikesCount:      post.LikesCount,
		CommentsCount:   post.CommentsCount,
		SharesCount:     post.SharesCount,
		ViewsCount:      post.ViewsCount,
		SavesCount:      post.SavesCount,
		EngagementRate:  post.EngagementRate,
		ReachCount:      post.ReachCount,
		ImpressionCount: post.ImpressionCount,
	}, nil
}

// SearchPosts searches for posts
func (ps *PostService) SearchPosts(query string, currentUserID *primitive.ObjectID, limit, skip int) ([]models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"content": bson.M{"$regex": query, "$options": "i"}},
					{"hashtags": bson.M{"$in": []string{query}}},
				},
			},
			{"is_published": true},
			{"deleted_at": bson.M{"$exists": false}},
			{"visibility": "public"}, // Only search public posts for now
		},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"engagement_rate": -1, "created_at": -1})

	cursor, err := ps.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Populate author information
	for i := range posts {
		ps.populatePostAuthor(&posts[i])
	}

	return posts, nil
}

// GetTrendingPosts retrieves trending posts
func (ps *PostService) GetTrendingPosts(limit, skip int, timeRange string) ([]models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Calculate time range filter
	var timeFilter time.Time
	switch timeRange {
	case "hour":
		timeFilter = time.Now().Add(-1 * time.Hour)
	case "day":
		timeFilter = time.Now().Add(-24 * time.Hour)
	case "week":
		timeFilter = time.Now().Add(-7 * 24 * time.Hour)
	default:
		timeFilter = time.Now().Add(-24 * time.Hour)
	}

	filter := bson.M{
		"is_published": true,
		"visibility":   "public",
		"deleted_at":   bson.M{"$exists": false},
		"created_at":   bson.M{"$gte": timeFilter},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{
			"engagement_rate": -1,
			"likes_count":     -1,
			"comments_count":  -1,
		})

	cursor, err := ps.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	// Populate author information
	for i := range posts {
		ps.populatePostAuthor(&posts[i])
	}

	return posts, nil
}

// Helper methods

func (ps *PostService) canUserViewPost(post *models.Post, userID primitive.ObjectID) bool {
	// Post author can always view
	if post.UserID == userID {
		return true
	}

	// Check if post is published and not hidden
	if !post.IsPublished || post.IsHidden {
		return false
	}

	// Check visibility
	switch post.Visibility {
	case models.PrivacyPublic:
		return true
	case models.PrivacyFriends:
		// Check if users are following each other
		return ps.areUsersFriends(post.UserID, userID)
	case models.PrivacyPrivate:
		return false
	default:
		return false
	}
}

func (ps *PostService) areUsersFriends(userID1, userID2 primitive.ObjectID) bool {
	// Check follow relationship - simplified implementation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	followCollection := ps.db.Collection("follows")
	count, err := followCollection.CountDocuments(ctx, bson.M{
		"follower_id": userID1,
		"followee_id": userID2,
		"status":      "accepted",
	})

	return err == nil && count > 0
}

func (ps *PostService) populatePostAuthor(post *models.Post) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := ps.userCollection.FindOne(ctx, bson.M{"_id": post.UserID}).Decode(&user)
	if err != nil {
		return err
	}

	post.Author = user.ToUserResponse()
	return nil
}

func (ps *PostService) updateUserPostCount(userID primitive.ObjectID, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	ps.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$inc": bson.M{"posts_count": value},
		"$set": bson.M{"updated_at": time.Now()},
	})
}

func (ps *PostService) updateUserLikesCount(userID primitive.ObjectID, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	ps.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$inc": bson.M{"total_likes_received": value},
		"$set": bson.M{"updated_at": time.Now()},
	})
}

func (ps *PostService) incrementViewCount(postID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps.collection.UpdateOne(ctx, bson.M{"_id": postID}, bson.M{
		"$inc": bson.M{"views_count": 1},
	})
}

func (ps *PostService) createHashtagEntries(hashtags []string, postID primitive.ObjectID) {
	// This would integrate with hashtag service to create/update hashtag entries
	// Implementation depends on hashtag tracking requirements
}

func (ps *PostService) createMentionNotifications(authorID, postID primitive.ObjectID, mentionedUsers []primitive.ObjectID) {
	// This would integrate with notification service to create mention notifications
	// Implementation depends on notification system
}

func extractHashtagsFromText(text string) []string {
	var hashtags []string
	words := strings.Fields(text)

	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			hashtag := strings.TrimPrefix(word, "#")
			hashtag = strings.TrimRight(hashtag, ".,!?;:")
			if len(hashtag) > 0 {
				hashtags = append(hashtags, hashtag)
			}
		}
	}

	return hashtags
}

func convertPollOptions(reqOptions []models.CreatePollOption) []models.PollOption {
	var options []models.PollOption
	for _, opt := range reqOptions {
		options = append(options, models.PollOption{
			ID:         primitive.NewObjectID(),
			Text:       opt.Text,
			VotesCount: 0,
			Percentage: 0.0,
		})
	}
	return options
}
func (us *PostService) GetCollection() *mongo.Collection {
    return us.collection
}