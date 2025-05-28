// internal/services/story_service.go
package services

import (
	"context"
	"errors"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StoryService struct {
	collection          *mongo.Collection
	viewCollection      *mongo.Collection
	highlightCollection *mongo.Collection
	userCollection      *mongo.Collection
	followCollection    *mongo.Collection
	likeCollection      *mongo.Collection
	db                  *mongo.Database
}

func NewStoryService() *StoryService {
	return &StoryService{
		collection:          config.DB.Collection("stories"),
		viewCollection:      config.DB.Collection("story_views"),
		highlightCollection: config.DB.Collection("story_highlights"),
		userCollection:      config.DB.Collection("users"),
		followCollection:    config.DB.Collection("follows"),
		likeCollection:      config.DB.Collection("likes"),
		db:                  config.DB,
	}
}

// CreateStory creates a new story
func (ss *StoryService) CreateStory(userID primitive.ObjectID, req models.CreateStoryRequest) (*models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate content type
	if !models.IsValidStoryContentType(req.ContentType) {
		return nil, errors.New("invalid content type for story")
	}

	// Convert allowed and blocked viewers
	var allowedViewers []primitive.ObjectID
	for _, viewerID := range req.AllowedViewers {
		if id, err := primitive.ObjectIDFromHex(viewerID); err == nil {
			allowedViewers = append(allowedViewers, id)
		}
	}

	var blockedViewers []primitive.ObjectID
	for _, viewerID := range req.BlockedViewers {
		if id, err := primitive.ObjectIDFromHex(viewerID); err == nil {
			blockedViewers = append(blockedViewers, id)
		}
	}

	// Create story
	story := &models.Story{
		UserID:          userID,
		Content:         req.Content,
		ContentType:     req.ContentType,
		Media:           req.Media,
		Duration:        req.Duration,
		Visibility:      req.Visibility,
		AllowedViewers:  allowedViewers,
		BlockedViewers:  blockedViewers,
		AllowReplies:    req.AllowReplies,
		AllowReactions:  req.AllowReactions,
		AllowSharing:    req.AllowSharing,
		AllowScreenshot: req.AllowScreenshot,
		BackgroundColor: req.BackgroundColor,
		TextColor:       req.TextColor,
		FontFamily:      req.FontFamily,
		Stickers:        req.Stickers,
		Mentions:        req.Mentions,
		Hashtags:        req.Hashtags,
		Location:        req.Location,
		Music:           req.Music,
	}

	story.BeforeCreate()

	result, err := ss.collection.InsertOne(ctx, story)
	if err != nil {
		return nil, err
	}

	story.ID = result.InsertedID.(primitive.ObjectID)

	// Populate author information
	if err := ss.populateStoryAuthor(story); err != nil {
		return nil, err
	}

	return story, nil
}

// GetStoryByID retrieves a story by ID with access control
func (ss *StoryService) GetStoryByID(storyID primitive.ObjectID, currentUserID *primitive.ObjectID) (*models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var story models.Story
	err := ss.collection.FindOne(ctx, bson.M{
		"_id":        storyID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&story)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("story not found")
		}
		return nil, err
	}

	// Check expiration
	story.CheckExpiration()
	if story.IsExpired && !story.IsHighlighted {
		return nil, errors.New("story has expired")
	}

	// Check access permissions
	var isFollowing bool
	var isAuthor bool

	if currentUserID != nil {
		isAuthor = story.UserID == *currentUserID
		if !isAuthor {
			isFollowing = ss.isUserFollowing(*currentUserID, story.UserID)
		}
	}

	if !story.CanViewStory(*currentUserID, isFollowing, isAuthor) {
		return nil, errors.New("access denied")
	}

	// Populate author information
	if err := ss.populateStoryAuthor(&story); err != nil {
		return nil, err
	}

	return &story, nil
}

// GetUserStories retrieves stories from a specific user
func (ss *StoryService) GetUserStories(userID primitive.ObjectID, currentUserID *primitive.ObjectID) ([]models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if current user can view stories from this user
	var isFollowing bool
	var isAuthor bool

	if currentUserID != nil {
		isAuthor = userID == *currentUserID
		if !isAuthor {
			isFollowing = ss.isUserFollowing(*currentUserID, userID)
		}
	}

	filter := bson.M{
		"user_id":    userID,
		"deleted_at": bson.M{"$exists": false},
		"is_hidden":  false,
	}

	// Add visibility filter based on relationship
	if !isAuthor {
		visibilityFilter := []bson.M{
			{"visibility": models.PrivacyPublic},
		}

		if isFollowing {
			visibilityFilter = append(visibilityFilter, bson.M{"visibility": models.PrivacyFriends})
		}

		if currentUserID != nil {
			visibilityFilter = append(visibilityFilter, bson.M{
				"visibility":      models.PrivacyPrivate,
				"allowed_viewers": bson.M{"$in": []primitive.ObjectID{*currentUserID}},
			})
		}

		filter["$or"] = visibilityFilter

		// Exclude blocked viewers
		if currentUserID != nil {
			filter["blocked_viewers"] = bson.M{"$nin": []primitive.ObjectID{*currentUserID}}
		}
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(50) // Reasonable limit for stories

	cursor, err := ss.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		return nil, err
	}

	// Filter expired non-highlighted stories and populate authors
	var activeStories []models.Story
	for i := range stories {
		stories[i].CheckExpiration()
		if !stories[i].IsExpired || stories[i].IsHighlighted {
			ss.populateStoryAuthor(&stories[i])
			activeStories = append(activeStories, stories[i])
		}
	}

	return activeStories, nil
}

// GetFollowingStories retrieves stories from users that current user follows
func (ss *StoryService) GetFollowingStories(userID primitive.ObjectID) ([]models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get list of users that current user follows
	followingPipeline := []bson.M{
		{
			"$match": bson.M{
				"follower_id": userID,
				"status":      "accepted",
				"deleted_at":  bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":           nil,
				"following_ids": bson.M{"$push": "$followee_id"},
			},
		},
	}

	cursor, err := ss.followCollection.Aggregate(ctx, followingPipeline)
	if err != nil {
		return nil, err
	}

	var followingResult []struct {
		FollowingIDs []primitive.ObjectID `bson:"following_ids"`
	}

	if err := cursor.All(ctx, &followingResult); err != nil {
		return nil, err
	}

	var followingIDs []primitive.ObjectID
	if len(followingResult) > 0 {
		followingIDs = followingResult[0].FollowingIDs
	}

	// Add current user to see their own stories
	followingIDs = append(followingIDs, userID)

	// Get stories from followed users
	filter := bson.M{
		"user_id":    bson.M{"$in": followingIDs},
		"deleted_at": bson.M{"$exists": false},
		"is_hidden":  false,
		"$or": []bson.M{
			{"visibility": models.PrivacyPublic},
			{"visibility": models.PrivacyFriends},
			{
				"visibility":      models.PrivacyPrivate,
				"allowed_viewers": bson.M{"$in": []primitive.ObjectID{userID}},
			},
		},
		"blocked_viewers": bson.M{"$nin": []primitive.ObjectID{userID}},
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(100)

	storyCursor, err := ss.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer storyCursor.Close(ctx)

	var stories []models.Story
	if err := storyCursor.All(ctx, &stories); err != nil {
		return nil, err
	}

	// Filter expired stories and populate authors
	var activeStories []models.Story
	for i := range stories {
		stories[i].CheckExpiration()
		if !stories[i].IsExpired || stories[i].IsHighlighted {
			ss.populateStoryAuthor(&stories[i])
			activeStories = append(activeStories, stories[i])
		}
	}

	return activeStories, nil
}

// UpdateStory updates an existing story (limited fields can be updated)
func (ss *StoryService) UpdateStory(storyID, userID primitive.ObjectID, req map[string]interface{}) (*models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if story exists and user owns it
	story, err := ss.GetStoryByID(storyID, &userID)
	if err != nil {
		return nil, err
	}

	if story.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Stories have limited update capabilities (mostly settings)
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}
	setFields := update["$set"].(bson.M)

	// Allow updating certain fields
	if visibility, ok := req["visibility"]; ok {
		if vis, ok := visibility.(models.PrivacyLevel); ok {
			setFields["visibility"] = vis
		}
	}

	if allowReplies, ok := req["allow_replies"]; ok {
		if allow, ok := allowReplies.(bool); ok {
			setFields["allow_replies"] = allow
		}
	}

	if allowReactions, ok := req["allow_reactions"]; ok {
		if allow, ok := allowReactions.(bool); ok {
			setFields["allow_reactions"] = allow
		}
	}

	if allowSharing, ok := req["allow_sharing"]; ok {
		if allow, ok := allowSharing.(bool); ok {
			setFields["allow_sharing"] = allow
		}
	}

	if allowScreenshot, ok := req["allow_screenshot"]; ok {
		if allow, ok := allowScreenshot.(bool); ok {
			setFields["allow_screenshot"] = allow
		}
	}

	_, err = ss.collection.UpdateOne(ctx, bson.M{"_id": storyID}, update)
	if err != nil {
		return nil, err
	}

	return ss.GetStoryByID(storyID, &userID)
}

// DeleteStory soft deletes a story
func (ss *StoryService) DeleteStory(storyID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user role for permission check
	var user models.User
	err := ss.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	// Check if story exists
	var story models.Story
	err = ss.collection.FindOne(ctx, bson.M{
		"_id":        storyID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&story)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("story not found")
		}
		return err
	}

	// Check permissions
	if !story.CanDeleteStory(userID, user.Role) {
		return errors.New("access denied")
	}

	// Soft delete
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = ss.collection.UpdateOne(ctx, bson.M{"_id": storyID}, update)
	return err
}

// ViewStory records a view for a story
func (ss *StoryService) ViewStory(storyID, viewerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if story exists and user can view it
	story, err := ss.GetStoryByID(storyID, &viewerID)
	if err != nil {
		return err
	}

	// Don't record view for own story
	if story.UserID == viewerID {
		return nil
	}

	// Check if user already viewed this story
	var existingView models.StoryView
	err = ss.viewCollection.FindOne(ctx, bson.M{
		"story_id": storyID,
		"user_id":  viewerID,
	}).Decode(&existingView)

	if err == mongo.ErrNoDocuments {
		// Create new view record
		view := &models.StoryView{
			StoryID:      storyID,
			UserID:       viewerID,
			ViewDuration: float64(story.Duration), // Assume full view by default
			WatchedFully: true,
			Source:       "feed",
			DeviceType:   "mobile",
		}

		view.BeforeCreate()
		_, err = ss.viewCollection.InsertOne(ctx, view)
		if err != nil {
			return err
		}

		// Increment story views count
		ss.collection.UpdateOne(ctx, bson.M{"_id": storyID}, bson.M{
			"$inc": bson.M{"views_count": 1, "unique_views_count": 1},
		})
	}

	return nil
}

// GetStoryViews retrieves viewers of a story
func (ss *StoryService) GetStoryViews(storyID, userID primitive.ObjectID, limit, skip int) ([]models.StoryViewResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if story exists and user owns it
	story, err := ss.GetStoryByID(storyID, &userID)
	if err != nil {
		return nil, err
	}

	if story.UserID != userID {
		return nil, errors.New("access denied")
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"story_id": storyID,
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

	cursor, err := ss.viewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.StoryView `bson:",inline"`
		User             models.User `bson:"user"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var views []models.StoryViewResponse
	for _, result := range results {
		view := result.StoryView.ToStoryViewResponse()
		view.User = result.User.ToUserResponse()
		views = append(views, view)
	}

	return views, nil
}

// ReactToStory adds a reaction to a story
func (ss *StoryService) ReactToStory(storyID, userID primitive.ObjectID, reactionType models.ReactionType) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if story exists and allows reactions
	story, err := ss.GetStoryByID(storyID, &userID)
	if err != nil {
		return err
	}

	if !story.AllowReactions {
		return errors.New("reactions not allowed on this story")
	}

	// Check if user already reacted
	var existingLike models.Like
	err = ss.likeCollection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   storyID,
		"target_type": "story",
	}).Decode(&existingLike)

	if err == nil {
		// Update existing reaction
		_, err = ss.likeCollection.UpdateOne(ctx, bson.M{"_id": existingLike.ID}, bson.M{
			"$set": bson.M{
				"reaction_type": reactionType,
				"updated_at":    time.Now(),
			},
		})
	} else if err == mongo.ErrNoDocuments {
		// Create new reaction
		like := &models.Like{
			UserID:       userID,
			TargetID:     storyID,
			TargetType:   "story",
			ReactionType: reactionType,
		}
		like.BeforeCreate()

		_, err = ss.likeCollection.InsertOne(ctx, like)
		if err != nil {
			return err
		}

		// Increment story likes count
		ss.collection.UpdateOne(ctx, bson.M{"_id": storyID}, bson.M{
			"$inc": bson.M{"likes_count": 1},
		})
	}

	return err
}

// UnreactToStory removes a reaction from a story
func (ss *StoryService) UnreactToStory(storyID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Remove the reaction
	result, err := ss.likeCollection.DeleteOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   storyID,
		"target_type": "story",
	})

	if err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		// Decrement story likes count
		ss.collection.UpdateOne(ctx, bson.M{"_id": storyID}, bson.M{
			"$inc": bson.M{"likes_count": -1},
		})
	}

	return nil
}

// GetStoryReactions retrieves reactions to a story
func (ss *StoryService) GetStoryReactions(storyID, userID primitive.ObjectID, limit, skip int) ([]models.LikeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if story exists and user can view it
	_, err := ss.GetStoryByID(storyID, &userID)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id":   storyID,
				"target_type": "story",
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

	cursor, err := ss.likeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Like `bson:",inline"`
		User        models.User `bson:"user"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var reactions []models.LikeResponse
	for _, result := range results {
		reaction := result.Like.ToLikeResponse()
		reaction.User = result.User.ToUserResponse()
		reactions = append(reactions, reaction)
	}

	return reactions, nil
}

// GetStoryStats retrieves story statistics
func (ss *StoryService) GetStoryStats(storyID, userID primitive.ObjectID) (map[string]interface{}, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if story exists and user owns it
	story, err := ss.GetStoryByID(storyID, &userID)
	if err != nil {
		return nil, err
	}

	if story.UserID != userID {
		return nil, errors.New("access denied")
	}

	stats := map[string]interface{}{
		"story_id":              storyID.Hex(),
		"views_count":           story.ViewsCount,
		"unique_views_count":    story.UniqueViewsCount,
		"likes_count":           story.LikesCount,
		"replies_count":         story.RepliesCount,
		"shares_count":          story.SharesCount,
		"average_view_duration": story.AverageViewDuration,
		"completion_rate":       story.CompletionRate,
		"engagement_rate":       story.EngagementRate,
		"is_highlighted":        story.IsHighlighted,
		"created_at":            story.CreatedAt,
		"expires_at":            story.ExpiresAt,
	}

	return stats, nil
}

// GetActiveStories retrieves currently active stories from all users
func (ss *StoryService) GetActiveStories(currentUserID *primitive.ObjectID, limit, skip int) ([]models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{
		"deleted_at": bson.M{"$exists": false},
		"is_hidden":  false,
		"$or": []bson.M{
			{"is_expired": false},
			{"is_highlighted": true},
		},
	}

	// If user is authenticated, apply privacy filters
	if currentUserID != nil {
		filter["$and"] = []bson.M{
			{
				"$or": []bson.M{
					{"visibility": models.PrivacyPublic},
					{
						"visibility": models.PrivacyFriends,
						"user_id":    bson.M{"$in": ss.getFollowingUserIDs(*currentUserID)},
					},
					{
						"visibility":      models.PrivacyPrivate,
						"allowed_viewers": bson.M{"$in": []primitive.ObjectID{*currentUserID}},
					},
					{"user_id": *currentUserID}, // Own stories
				},
			},
			{
				"blocked_viewers": bson.M{"$nin": []primitive.ObjectID{*currentUserID}},
			},
		}
	} else {
		// Only public stories for unauthenticated users
		filter["visibility"] = models.PrivacyPublic
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ss.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		return nil, err
	}

	// Populate authors and check expiration
	var activeStories []models.Story
	for i := range stories {
		stories[i].CheckExpiration()
		if !stories[i].IsExpired || stories[i].IsHighlighted {
			ss.populateStoryAuthor(&stories[i])
			activeStories = append(activeStories, stories[i])
		}
	}

	return activeStories, nil
}

// ArchiveStory marks a story as archived (not implemented in model, but adding for compatibility)
func (ss *StoryService) ArchiveStory(storyID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if story exists and user owns it
	story, err := ss.GetStoryByID(storyID, &userID)
	if err != nil {
		return err
	}

	if story.UserID != userID {
		return errors.New("access denied")
	}

	// Mark as hidden (closest to archived)
	_, err = ss.collection.UpdateOne(ctx, bson.M{"_id": storyID}, bson.M{
		"$set": bson.M{
			"is_hidden":  true,
			"updated_at": time.Now(),
		},
	})

	return err
}

// GetArchivedStories retrieves user's archived stories
func (ss *StoryService) GetArchivedStories(userID primitive.ObjectID, limit, skip int) ([]models.Story, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_hidden":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ss.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err := cursor.All(ctx, &stories); err != nil {
		return nil, err
	}

	// Populate authors
	for i := range stories {
		ss.populateStoryAuthor(&stories[i])
	}

	return stories, nil
}

// Helper methods

// populateStoryAuthor populates the author information for a story
func (ss *StoryService) populateStoryAuthor(story *models.Story) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := ss.userCollection.FindOne(ctx, bson.M{"_id": story.UserID}).Decode(&user)
	if err != nil {
		return err
	}

	story.Author = user.ToUserResponse()
	return nil
}

// isUserFollowing checks if one user is following another
func (ss *StoryService) isUserFollowing(followerID, followeeID primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := ss.followCollection.CountDocuments(ctx, bson.M{
		"follower_id": followerID,
		"followee_id": followeeID,
		"status":      "accepted",
		"deleted_at":  bson.M{"$exists": false},
	})

	return err == nil && count > 0
}

// getFollowingUserIDs gets the list of user IDs that the current user follows
func (ss *StoryService) getFollowingUserIDs(userID primitive.ObjectID) []primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := ss.followCollection.Find(ctx, bson.M{
		"follower_id": userID,
		"status":      "accepted",
		"deleted_at":  bson.M{"$exists": false},
	})

	if err != nil {
		return []primitive.ObjectID{}
	}
	defer cursor.Close(ctx)

	var follows []models.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return []primitive.ObjectID{}
	}

	var userIDs []primitive.ObjectID
	for _, follow := range follows {
		userIDs = append(userIDs, follow.FolloweeID)
	}

	return userIDs
}

// Story Highlights methods

// CreateStoryHighlight creates a new story highlight
func (ss *StoryService) CreateStoryHighlight(userID primitive.ObjectID, req models.CreateStoryHighlightRequest) (*models.StoryHighlight, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert story IDs
	var storyIDs []primitive.ObjectID
	for _, storyIDStr := range req.StoryIDs {
		if id, err := primitive.ObjectIDFromHex(storyIDStr); err == nil {
			storyIDs = append(storyIDs, id)
		}
	}

	if len(storyIDs) == 0 {
		return nil, errors.New("at least one valid story ID is required")
	}

	// Verify user owns all the stories
	count, err := ss.collection.CountDocuments(ctx, bson.M{
		"_id":        bson.M{"$in": storyIDs},
		"user_id":    userID,
		"deleted_at": bson.M{"$exists": false},
	})

	if err != nil {
		return nil, err
	}

	if count != int64(len(storyIDs)) {
		return nil, errors.New("some stories not found or access denied")
	}

	// Create highlight
	highlight := &models.StoryHighlight{
		UserID:     userID,
		Title:      req.Title,
		CoverImage: req.CoverImage,
		StoryIDs:   storyIDs,
	}

	highlight.BeforeCreate()

	result, err := ss.highlightCollection.InsertOne(ctx, highlight)
	if err != nil {
		return nil, err
	}

	highlight.ID = result.InsertedID.(primitive.ObjectID)

	// Mark stories as highlighted
	ss.collection.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": storyIDs}}, bson.M{
		"$set": bson.M{
			"is_highlighted": true,
			"highlight_id":   highlight.ID,
		},
	})

	return highlight, nil
}

// GetUserStoryHighlights retrieves story highlights for a user
func (ss *StoryService) GetUserStoryHighlights(userID primitive.ObjectID) ([]models.StoryHighlightResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().SetSort(bson.M{"order": 1, "created_at": 1})

	cursor, err := ss.highlightCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var highlights []models.StoryHighlight
	if err := cursor.All(ctx, &highlights); err != nil {
		return nil, err
	}

	var responses []models.StoryHighlightResponse
	for _, highlight := range highlights {
		responses = append(responses, highlight.ToStoryHighlightResponse())
	}

	return responses, nil
}
