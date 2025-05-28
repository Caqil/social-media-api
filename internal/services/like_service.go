// internal/services/like_service.go
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
)

type LikeService struct {
	collection        *mongo.Collection
	postCollection    *mongo.Collection
	commentCollection *mongo.Collection
	storyCollection   *mongo.Collection
	messageCollection *mongo.Collection
	userCollection    *mongo.Collection
	db                *mongo.Database
}

func NewLikeService() *LikeService {
	return &LikeService{
		collection:        config.DB.Collection("likes"),
		postCollection:    config.DB.Collection("posts"),
		commentCollection: config.DB.Collection("comments"),
		storyCollection:   config.DB.Collection("stories"),
		messageCollection: config.DB.Collection("messages"),
		userCollection:    config.DB.Collection("users"),
		db:                config.DB,
	}
}

// CreateLike adds a like/reaction to a target
func (ls *LikeService) CreateLike(userID primitive.ObjectID, req models.CreateLikeRequest) (*models.Like, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert target ID
	targetID, err := primitive.ObjectIDFromHex(req.TargetID)
	if err != nil {
		return nil, errors.New("invalid target ID")
	}

	// Validate target exists and user can interact with it
	if err := ls.validateTarget(targetID, req.TargetType, userID); err != nil {
		return nil, err
	}

	// Check if user already liked this target
	var existingLike models.Like
	err = ls.collection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   targetID,
		"target_type": req.TargetType,
	}).Decode(&existingLike)

	if err == nil {
		// Update existing like with new reaction
		update := bson.M{
			"$set": bson.M{
				"reaction_type": req.ReactionType,
				"updated_at":    time.Now(),
			},
		}
		_, err = ls.collection.UpdateOne(ctx, bson.M{"_id": existingLike.ID}, update)
		if err != nil {
			return nil, err
		}

		existingLike.ReactionType = req.ReactionType
		existingLike.UpdatedAt = time.Now()

		// Populate user information
		ls.populateLikeUser(&existingLike)

		return &existingLike, nil
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Create new like
	like := &models.Like{
		UserID:       userID,
		TargetID:     targetID,
		TargetType:   req.TargetType,
		ReactionType: req.ReactionType,
	}

	like.BeforeCreate()

	result, err := ls.collection.InsertOne(ctx, like)
	if err != nil {
		return nil, err
	}

	like.ID = result.InsertedID.(primitive.ObjectID)

	// Update target engagement counts
	go ls.updateTargetCounts(targetID, req.TargetType, true)

	// Update user engagement stats
	go ls.updateUserEngagementStats(userID, req.TargetType, true)

	// Send notification if not self-like
	go ls.sendLikeNotification(userID, targetID, req.TargetType, req.ReactionType)

	// Populate user information
	ls.populateLikeUser(like)

	return like, nil
}

// UpdateLike changes the reaction type of an existing like
func (ls *LikeService) UpdateLike(likeID, userID primitive.ObjectID, req models.UpdateLikeRequest) (*models.Like, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the like
	var like models.Like
	err := ls.collection.FindOne(ctx, bson.M{
		"_id":     likeID,
		"user_id": userID,
	}).Decode(&like)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("like not found or access denied")
		}
		return nil, err
	}

	// Update reaction type
	update := bson.M{
		"$set": bson.M{
			"reaction_type": req.ReactionType,
			"updated_at":    time.Now(),
		},
	}

	_, err = ls.collection.UpdateOne(ctx, bson.M{"_id": likeID}, update)
	if err != nil {
		return nil, err
	}

	like.ReactionType = req.ReactionType
	like.UpdatedAt = time.Now()

	// Populate user information
	ls.populateLikeUser(&like)

	return &like, nil
}

// DeleteLike removes a like
func (ls *LikeService) DeleteLike(targetID, userID primitive.ObjectID, targetType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find and delete the like
	result, err := ls.collection.DeleteOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   targetID,
		"target_type": targetType,
	})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("like not found")
	}

	// Update target engagement counts
	go ls.updateTargetCounts(targetID, targetType, false)

	// Update user engagement stats
	go ls.updateUserEngagementStats(userID, targetType, false)

	return nil
}

// GetLikes retrieves users who liked a target
func (ls *LikeService) GetLikes(targetID primitive.ObjectID, targetType string, reactionType *models.ReactionType, limit, skip int) ([]models.LikeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchFilter := bson.M{
		"target_id":   targetID,
		"target_type": targetType,
	}

	// Filter by reaction type if specified
	if reactionType != nil {
		matchFilter["reaction_type"] = *reactionType
	}

	pipeline := []bson.M{
		{
			"$match": matchFilter,
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
		{
			"$project": bson.M{
				"_id":           1,
				"user_id":       1,
				"target_id":     1,
				"target_type":   1,
				"reaction_type": 1,
				"created_at":    1,
				"user": bson.M{
					"_id":          1,
					"username":     1,
					"first_name":   1,
					"last_name":    1,
					"display_name": 1,
					"profile_pic":  1,
					"is_verified":  1,
				},
			},
		},
	}

	cursor, err := ls.collection.Aggregate(ctx, pipeline)
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

// GetReactionSummary gets aggregated reaction statistics for a target
func (ls *LikeService) GetReactionSummary(targetID primitive.ObjectID, targetType string, currentUserID *primitive.ObjectID) (*models.ReactionSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id":   targetID,
				"target_type": targetType,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$reaction_type",
				"count": bson.M{"$sum": 1},
				"users": bson.M{"$push": "$user_id"},
			},
		},
	}

	cursor, err := ls.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    models.ReactionType  `bson:"_id"`
		Count int64                `bson:"count"`
		Users []primitive.ObjectID `bson:"users"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Build reaction summary
	reactions := make(map[models.ReactionType]int64)
	var totalCount int64
	var userReaction models.ReactionType

	for _, result := range results {
		reactions[result.ID] = result.Count
		totalCount += result.Count

		// Check if current user reacted
		if currentUserID != nil {
			for _, userID := range result.Users {
				if userID == *currentUserID {
					userReaction = result.ID
					break
				}
			}
		}
	}

	// Get top reactors (first few users)
	topReactors, _ := ls.getTopReactors(targetID, targetType, 5)

	summary := models.CreateReactionSummary(
		targetID.Hex(),
		targetType,
		reactions,
		topReactors,
		userReaction,
	)

	return &summary, nil
}

// GetUserLikes retrieves likes made by a specific user
func (ls *LikeService) GetUserLikes(userID primitive.ObjectID, targetType *string, limit, skip int) ([]models.LikeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchFilter := bson.M{
		"user_id": userID,
	}

	// Filter by target type if specified
	if targetType != nil {
		matchFilter["target_type"] = *targetType
	}

	pipeline := []bson.M{
		{
			"$match": matchFilter,
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

	cursor, err := ls.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var likes []models.Like
	if err := cursor.All(ctx, &likes); err != nil {
		return nil, err
	}

	// Convert to response format
	var likeResponses []models.LikeResponse
	for _, like := range likes {
		ls.populateLikeUser(&like)
		likeResponses = append(likeResponses, like.ToLikeResponse())
	}

	return likeResponses, nil
}

// GetDetailedReactionStats gets detailed reaction statistics
func (ls *LikeService) GetDetailedReactionStats(targetID primitive.ObjectID, targetType string) (*models.ReactionStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id":   targetID,
				"target_type": targetType,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$reaction_type",
				"count": bson.M{"$sum": 1},
				"users": bson.M{
					"$push": bson.M{
						"user_id":    "$user_id",
						"created_at": "$created_at",
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"userIds": "$users.user_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{"$in": []interface{}{"$_id", "$$userIds"}},
						},
					},
					{
						"$limit": 3, // Get first 3 users for sample
					},
					{
						"$project": bson.M{
							"_id":          1,
							"username":     1,
							"first_name":   1,
							"last_name":    1,
							"display_name": 1,
							"profile_pic":  1,
							"is_verified":  1,
						},
					},
				},
				"as": "sample_users",
			},
		},
	}

	cursor, err := ls.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID          models.ReactionType   `bson:"_id"`
		Count       int64                 `bson:"count"`
		SampleUsers []models.UserResponse `bson:"sample_users"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Build detailed stats
	reactionCounts := make(map[models.ReactionType]models.ReactionCount)

	for _, result := range results {
		reactionCounts[result.ID] = models.ReactionCount{
			Count:       result.Count,
			SampleUsers: result.SampleUsers,
		}
	}

	stats := models.CreateReactionStats(reactionCounts)
	return &stats, nil
}

// CheckUserReaction checks if a user has reacted to a target
func (ls *LikeService) CheckUserReaction(targetID, userID primitive.ObjectID, targetType string) (*models.ReactionType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var like models.Like
	err := ls.collection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   targetID,
		"target_type": targetType,
	}).Decode(&like)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No reaction found
		}
		return nil, err
	}

	return &like.ReactionType, nil
}

// GetTrendingReactions gets trending reactions across all targets
func (ls *LikeService) GetTrendingReactions(targetType *string, timeRange time.Duration, limit int) ([]models.ReactionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	matchFilter := bson.M{
		"created_at": bson.M{
			"$gte": time.Now().Add(-timeRange),
		},
	}

	if targetType != nil {
		matchFilter["target_type"] = *targetType
	}

	pipeline := []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$group": bson.M{
				"_id":   "$reaction_type",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := ls.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    models.ReactionType `bson:"_id"`
		Count int64               `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to reaction info
	var reactions []models.ReactionInfo
	for _, result := range results {
		reactions = append(reactions, models.ReactionInfo{
			Type:  result.ID,
			Count: result.Count,
			Emoji: models.GetReactionEmoji(result.ID),
			Name:  models.GetReactionName(result.ID),
		})
	}

	return reactions, nil
}

// Helper methods

// validateTarget validates that target exists and user can interact with it
func (ls *LikeService) validateTarget(targetID primitive.ObjectID, targetType string, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var collection *mongo.Collection
	var filter bson.M

	switch targetType {
	case "post":
		collection = ls.postCollection
		filter = bson.M{
			"_id":           targetID,
			"is_published":  true,
			"likes_enabled": true,
			"deleted_at":    bson.M{"$exists": false},
		}
	case "comment":
		collection = ls.commentCollection
		filter = bson.M{
			"_id":         targetID,
			"is_approved": true,
			"is_hidden":   false,
			"deleted_at":  bson.M{"$exists": false},
		}
	case "story":
		collection = ls.storyCollection
		filter = bson.M{
			"_id":             targetID,
			"is_expired":      false,
			"allow_reactions": true,
			"deleted_at":      bson.M{"$exists": false},
		}
	case "message":
		collection = ls.messageCollection
		filter = bson.M{
			"_id":        targetID,
			"deleted_at": bson.M{"$exists": false},
		}
	default:
		return errors.New("invalid target type")
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("target not found or not accessible")
	}

	return nil
}

// updateTargetCounts updates engagement counts on the target
func (ls *LikeService) updateTargetCounts(targetID primitive.ObjectID, targetType string, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	var collection *mongo.Collection
	switch targetType {
	case "post":
		collection = ls.postCollection
	case "comment":
		collection = ls.commentCollection
	case "story":
		collection = ls.storyCollection
	case "message":
		collection = ls.messageCollection
	default:
		return
	}

	update := bson.M{
		"$inc": bson.M{"likes_count": value},
		"$set": bson.M{"updated_at": time.Now()},
	}

	collection.UpdateOne(ctx, bson.M{"_id": targetID}, update)
}

// updateUserEngagementStats updates user engagement statistics
func (ls *LikeService) updateUserEngagementStats(userID primitive.ObjectID, targetType string, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	update := bson.M{
		"$inc": bson.M{"likes_given": value},
		"$set": bson.M{"updated_at": time.Now()},
	}

	ls.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
}

// populateLikeUser populates user information for a like
func (ls *LikeService) populateLikeUser(like *models.Like) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := ls.userCollection.FindOne(ctx, bson.M{"_id": like.UserID}).Decode(&user)
	if err != nil {
		return err
	}

	like.User = user.ToUserResponse()
	return nil
}

// getTopReactors gets top reactors for a target
func (ls *LikeService) getTopReactors(targetID primitive.ObjectID, targetType string, limit int) ([]models.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id":   targetID,
				"target_type": targetType,
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
			"$limit": limit,
		},
		{
			"$project": bson.M{
				"user._id":          1,
				"user.username":     1,
				"user.first_name":   1,
				"user.last_name":    1,
				"user.display_name": 1,
				"user.profile_pic":  1,
				"user.is_verified":  1,
			},
		},
	}

	cursor, err := ls.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		User models.UserResponse `bson:"user"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var users []models.UserResponse
	for _, result := range results {
		users = append(users, result.User)
	}

	return users, nil
}

// sendLikeNotification sends notification for like action
func (ls *LikeService) sendLikeNotification(userID, targetID primitive.ObjectID, targetType string, reactionType models.ReactionType) {
	// This would integrate with notification service
	// Implementation depends on notification system requirements
}
