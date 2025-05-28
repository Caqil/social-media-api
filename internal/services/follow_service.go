// internal/services/follow_service.go
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

type FollowService struct {
	followCollection *mongo.Collection
	userCollection   *mongo.Collection
	db               *mongo.Database
}

func NewFollowService() *FollowService {
	return &FollowService{
		followCollection: config.DB.Collection("follows"),
		userCollection:   config.DB.Collection("users"),
		db:               config.DB,
	}
}

// FollowUser creates a follow relationship
func (fs *FollowService) FollowUser(followerID, followeeID primitive.ObjectID) (*models.Follow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if users exist
	if !fs.userExists(ctx, followerID) || !fs.userExists(ctx, followeeID) {
		return nil, errors.New("user not found")
	}

	// Check if already following
	var existingFollow models.Follow
	err := fs.followCollection.FindOne(ctx, bson.M{
		"follower_id": followerID,
		"followee_id": followeeID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&existingFollow)

	if err == nil {
		if existingFollow.Status == models.FollowAccepted {
			return nil, errors.New("already following this user")
		} else if existingFollow.Status == models.FollowPending {
			return nil, errors.New("follow request already pending")
		}
	}

	// Get followee's privacy settings to determine if approval is needed
	var followee models.User
	err = fs.userCollection.FindOne(ctx, bson.M{"_id": followeeID}).Decode(&followee)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Create follow relationship
	follow := &models.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		Status:     models.FollowPending,
	}

	// Auto-approve if user has public profile or no privacy restrictions
	if !followee.PrivacySettings.RequireApprovalToFollow {
		follow.Status = models.FollowAccepted
		follow.AcceptedAt = &[]time.Time{time.Now()}[0]
	}

	follow.BeforeCreate()

	result, err := fs.followCollection.InsertOne(ctx, follow)
	if err != nil {
		return nil, err
	}

	follow.ID = result.InsertedID.(primitive.ObjectID)

	// Update user follow counts if accepted
	if follow.Status == models.FollowAccepted {
		go fs.updateFollowCounts(followerID, followeeID, true)
	}

	return follow, nil
}

// UnfollowUser removes a follow relationship
func (fs *FollowService) UnfollowUser(followerID, followeeID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find and soft delete the follow relationship
	filter := bson.M{
		"follower_id": followerID,
		"followee_id": followeeID,
		"deleted_at":  bson.M{"$exists": false},
	}

	var follow models.Follow
	err := fs.followCollection.FindOne(ctx, filter).Decode(&follow)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("follow relationship not found")
		}
		return err
	}

	// Soft delete
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = fs.followCollection.UpdateOne(ctx, bson.M{"_id": follow.ID}, update)
	if err != nil {
		return err
	}

	// Update follow counts if it was accepted
	if follow.Status == models.FollowAccepted {
		go fs.updateFollowCounts(followerID, followeeID, false)
	}

	return nil
}

// GetFollowers retrieves a user's followers
func (fs *FollowService) GetFollowers(userID primitive.ObjectID, currentUserID *primitive.ObjectID, limit, skip int) ([]models.FollowResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"followee_id": userID,
				"status":      models.FollowAccepted,
				"deleted_at":  bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "follower_id",
				"foreignField": "_id",
				"as":           "follower",
			},
		},
		{
			"$unwind": "$follower",
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

	cursor, err := fs.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Follow `bson:",inline"`
		Follower      models.User `bson:"follower"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to response format
	var followers []models.FollowResponse
	for _, result := range results {
		followResponse := models.FollowResponse{
			ID:        result.Follow.ID,
			User:      result.Follower.ToUserResponse(),
			Status:    result.Follow.Status,
			CreatedAt: result.Follow.CreatedAt,
		}

		// Add current user context if provided
		if currentUserID != nil {
			followResponse.IsFollowedByCurrentUser = fs.isFollowing(ctx, *currentUserID, result.Follower.ID)
		}

		followers = append(followers, followResponse)
	}

	return followers, nil
}

// GetFollowing retrieves users that a user is following
func (fs *FollowService) GetFollowing(userID primitive.ObjectID, currentUserID *primitive.ObjectID, limit, skip int) ([]models.FollowResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"follower_id": userID,
				"status":      models.FollowAccepted,
				"deleted_at":  bson.M{"$exists": false},
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
			"$unwind": "$followee",
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

	cursor, err := fs.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Follow `bson:",inline"`
		Followee      models.User `bson:"followee"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to response format
	var following []models.FollowResponse
	for _, result := range results {
		followResponse := models.FollowResponse{
			ID:        result.Follow.ID,
			User:      result.Followee.ToUserResponse(),
			Status:    result.Follow.Status,
			CreatedAt: result.Follow.CreatedAt,
		}

		// Add current user context if provided
		if currentUserID != nil && *currentUserID != userID {
			followResponse.IsFollowedByCurrentUser = fs.isFollowing(ctx, *currentUserID, result.Followee.ID)
		}

		following = append(following, followResponse)
	}

	return following, nil
}

// GetPendingFollowRequests retrieves pending follow requests for a user
func (fs *FollowService) GetPendingFollowRequests(userID primitive.ObjectID, limit, skip int) ([]models.FollowResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"followee_id": userID,
				"status":      models.FollowPending,
				"deleted_at":  bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "follower_id",
				"foreignField": "_id",
				"as":           "follower",
			},
		},
		{
			"$unwind": "$follower",
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

	cursor, err := fs.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Follow `bson:",inline"`
		Follower      models.User `bson:"follower"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to response format
	var requests []models.FollowResponse
	for _, result := range results {
		requests = append(requests, models.FollowResponse{
			ID:        result.Follow.ID,
			User:      result.Follower.ToUserResponse(),
			Status:    result.Follow.Status,
			CreatedAt: result.Follow.CreatedAt,
		})
	}

	return requests, nil
}

// GetSentFollowRequests retrieves follow requests sent by a user
func (fs *FollowService) GetSentFollowRequests(userID primitive.ObjectID, limit, skip int) ([]models.FollowResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"follower_id": userID,
				"status":      models.FollowPending,
				"deleted_at":  bson.M{"$exists": false},
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
			"$unwind": "$followee",
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

	cursor, err := fs.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		models.Follow `bson:",inline"`
		Followee      models.User `bson:"followee"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to response format
	var requests []models.FollowResponse
	for _, result := range results {
		requests = append(requests, models.FollowResponse{
			ID:        result.Follow.ID,
			User:      result.Followee.ToUserResponse(),
			Status:    result.Follow.Status,
			CreatedAt: result.Follow.CreatedAt,
		})
	}

	return requests, nil
}

// AcceptFollowRequest accepts a follow request
func (fs *FollowService) AcceptFollowRequest(followID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the follow request
	var follow models.Follow
	err := fs.followCollection.FindOne(ctx, bson.M{
		"_id":         followID,
		"followee_id": userID,
		"status":      models.FollowPending,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&follow)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("follow request not found or access denied")
		}
		return err
	}

	// Accept the request
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      models.FollowAccepted,
			"accepted_at": now,
			"updated_at":  now,
		},
	}

	_, err = fs.followCollection.UpdateOne(ctx, bson.M{"_id": followID}, update)
	if err != nil {
		return err
	}

	// Update follow counts
	go fs.updateFollowCounts(follow.FollowerID, follow.FolloweeID, true)

	return nil
}

// RejectFollowRequest rejects a follow request
func (fs *FollowService) RejectFollowRequest(followID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the follow request
	var follow models.Follow
	err := fs.followCollection.FindOne(ctx, bson.M{
		"_id":         followID,
		"followee_id": userID,
		"status":      models.FollowPending,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&follow)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("follow request not found or access denied")
		}
		return err
	}

	// Soft delete the request
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":     models.FollowRejected,
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = fs.followCollection.UpdateOne(ctx, bson.M{"_id": followID}, update)
	return err
}

// CancelFollowRequest cancels a sent follow request
func (fs *FollowService) CancelFollowRequest(followID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the follow request
	var follow models.Follow
	err := fs.followCollection.FindOne(ctx, bson.M{
		"_id":         followID,
		"follower_id": userID,
		"status":      models.FollowPending,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&follow)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("follow request not found or access denied")
		}
		return err
	}

	// Soft delete the request
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = fs.followCollection.UpdateOne(ctx, bson.M{"_id": followID}, update)
	return err
}

// RemoveFollower removes a follower
func (fs *FollowService) RemoveFollower(userID, followerID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the follow relationship
	filter := bson.M{
		"follower_id": followerID,
		"followee_id": userID,
		"status":      models.FollowAccepted,
		"deleted_at":  bson.M{"$exists": false},
	}

	var follow models.Follow
	err := fs.followCollection.FindOne(ctx, filter).Decode(&follow)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("follower relationship not found")
		}
		return err
	}

	// Soft delete
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = fs.followCollection.UpdateOne(ctx, bson.M{"_id": follow.ID}, update)
	if err != nil {
		return err
	}

	// Update follow counts
	go fs.updateFollowCounts(followerID, userID, false)

	return nil
}

// GetFollowStats retrieves follow statistics for a user
func (fs *FollowService) GetFollowStats(userID primitive.ObjectID) (*models.FollowStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get followers count
	followersCount, err := fs.followCollection.CountDocuments(ctx, bson.M{
		"followee_id": userID,
		"status":      models.FollowAccepted,
		"deleted_at":  bson.M{"$exists": false},
	})
	if err != nil {
		return nil, err
	}

	// Get following count
	followingCount, err := fs.followCollection.CountDocuments(ctx, bson.M{
		"follower_id": userID,
		"status":      models.FollowAccepted,
		"deleted_at":  bson.M{"$exists": false},
	})
	if err != nil {
		return nil, err
	}

	// Get pending requests count
	pendingRequestsCount, err := fs.followCollection.CountDocuments(ctx, bson.M{
		"followee_id": userID,
		"status":      models.FollowPending,
		"deleted_at":  bson.M{"$exists": false},
	})
	if err != nil {
		return nil, err
	}

	return &models.FollowStats{
		FollowersCount:       followersCount,
		FollowingCount:       followingCount,
		PendingRequestsCount: pendingRequestsCount,
	}, nil
}

// GetFollowStatus checks the follow status between two users
func (fs *FollowService) GetFollowStatus(followerID, followeeID primitive.ObjectID) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var follow models.Follow
	err := fs.followCollection.FindOne(ctx, bson.M{
		"follower_id": followerID,
		"followee_id": followeeID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&follow)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "not_following", nil
		}
		return "", err
	}

	return string(follow.Status), nil
}

// GetMutualFollows retrieves mutual follows between two users
func (fs *FollowService) GetMutualFollows(userID1, userID2 primitive.ObjectID, limit, skip int) ([]models.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Find users that both users follow
	pipeline := []bson.M{
		// Get user1's following
		{
			"$match": bson.M{
				"follower_id": userID1,
				"status":      models.FollowAccepted,
				"deleted_at":  bson.M{"$exists": false},
			},
		},
		// Look for user2 also following the same users
		{
			"$lookup": bson.M{
				"from": "follows",
				"let":  bson.M{"followee_id": "$followee_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{"$eq": []string{"$follower_id", userID2.Hex()}},
									{"$eq": []string{"$followee_id", "$$followee_id"}},
									{"$eq": []string{"$status", string(models.FollowAccepted)}},
									{"$not": bson.M{"$ifNull": []interface{}{"$deleted_at", false}}},
								},
							},
						},
					},
				},
				"as": "mutual",
			},
		},
		// Only keep records where both users follow the same person
		{
			"$match": bson.M{
				"mutual": bson.M{"$ne": []interface{}{}},
			},
		},
		// Get user details
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "followee_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$unwind": "$user",
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := fs.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		User models.User `bson:"user"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to response format
	var mutualFollows []models.UserResponse
	for _, result := range results {
		mutualFollows = append(mutualFollows, result.User.ToUserResponse())
	}

	return mutualFollows, nil
}

// GetSuggestedUsers retrieves suggested users to follow
func (fs *FollowService) GetSuggestedUsers(userID primitive.ObjectID, limit int) ([]models.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Get users that the user's followers also follow (friends of friends)
	pipeline := []bson.M{
		// Get user's followers
		{
			"$match": bson.M{
				"followee_id": userID,
				"status":      models.FollowAccepted,
				"deleted_at":  bson.M{"$exists": false},
			},
		},
		// Get who these followers are following
		{
			"$lookup": bson.M{
				"from": "follows",
				"let":  bson.M{"follower_id": "$follower_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{"$eq": []string{"$follower_id", "$$follower_id"}},
									{"$eq": []string{"$status", string(models.FollowAccepted)}},
									{"$not": bson.M{"$ifNull": []interface{}{"$deleted_at", false}}},
								},
							},
						},
					},
				},
				"as": "following",
			},
		},
		{
			"$unwind": "$following",
		},
		// Exclude the current user and users already followed
		{
			"$match": bson.M{
				"following.followee_id": bson.M{
					"$ne": userID,
				},
			},
		},
		// Group by suggested user and count occurrences
		{
			"$group": bson.M{
				"_id":   "$following.followee_id",
				"score": bson.M{"$sum": 1},
			},
		},
		// Sort by score
		{
			"$sort": bson.M{"score": -1},
		},
		// Get user details
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$unwind": "$user",
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := fs.followCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		User  models.User `bson:"user"`
		Score int         `bson:"score"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to response format
	var suggestions []models.UserResponse
	for _, result := range results {
		suggestions = append(suggestions, result.User.ToUserResponse())
	}

	return suggestions, nil
}

// BulkFollowUsers follows multiple users at once
func (fs *FollowService) BulkFollowUsers(followerID primitive.ObjectID, userIDStrs []string) (map[string]interface{}, error) {
	results := map[string]interface{}{
		"success": []string{},
		"failed":  []map[string]string{},
	}

	for _, userIDStr := range userIDStrs {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			results["failed"] = append(results["failed"].([]map[string]string), map[string]string{
				"user_id": userIDStr,
				"error":   "invalid user ID format",
			})
			continue
		}

		_, err = fs.FollowUser(followerID, userID)
		if err != nil {
			results["failed"] = append(results["failed"].([]map[string]string), map[string]string{
				"user_id": userIDStr,
				"error":   err.Error(),
			})
		} else {
			results["success"] = append(results["success"].([]string), userIDStr)
		}
	}

	return results, nil
}

// GetFollowActivity retrieves recent follow activity
func (fs *FollowService) GetFollowActivity(userID primitive.ObjectID, activityType string, limit, skip int) ([]models.FollowActivity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	switch activityType {
	case "new_followers":
		filter["followee_id"] = userID
		filter["status"] = models.FollowAccepted
	case "new_following":
		filter["follower_id"] = userID
		filter["status"] = models.FollowAccepted
	case "follow_requests":
		filter["followee_id"] = userID
		filter["status"] = models.FollowPending
	default: // "all"
		filter["$or"] = []bson.M{
			{"followee_id": userID},
			{"follower_id": userID},
		}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := fs.followCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []models.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, err
	}

	// Convert to activity format
	var activities []models.FollowActivity
	for _, follow := range follows {
		var activityType string
		var relatedUserID primitive.ObjectID

		if follow.FolloweeID == userID {
			activityType = "new_follower"
			relatedUserID = follow.FollowerID
		} else {
			activityType = "new_following"
			relatedUserID = follow.FolloweeID
		}

		if follow.Status == models.FollowPending {
			activityType = "follow_request"
		}

		activity := models.FollowActivity{
			ID:            follow.ID,
			Type:          activityType,
			RelatedUserID: relatedUserID,
			Status:        follow.Status,
			CreatedAt:     follow.CreatedAt,
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// Helper methods

// userExists checks if a user exists
func (fs *FollowService) userExists(ctx context.Context, userID primitive.ObjectID) bool {
	count, err := fs.userCollection.CountDocuments(ctx, bson.M{
		"_id":        userID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	})
	return err == nil && count > 0
}

// isFollowing checks if user1 is following user2
func (fs *FollowService) isFollowing(ctx context.Context, followerID, followeeID primitive.ObjectID) bool {
	count, err := fs.followCollection.CountDocuments(ctx, bson.M{
		"follower_id": followerID,
		"followee_id": followeeID,
		"status":      models.FollowAccepted,
		"deleted_at":  bson.M{"$exists": false},
	})
	return err == nil && count > 0
}

// updateFollowCounts updates follower/following counts for users
func (fs *FollowService) updateFollowCounts(followerID, followeeID primitive.ObjectID, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	// Update follower's following count
	fs.userCollection.UpdateOne(ctx, bson.M{"_id": followerID}, bson.M{
		"$inc": bson.M{"following_count": value},
		"$set": bson.M{"updated_at": time.Now()},
	})

	// Update followee's followers count
	fs.userCollection.UpdateOne(ctx, bson.M{"_id": followeeID}, bson.M{
		"$inc": bson.M{"followers_count": value},
		"$set": bson.M{"updated_at": time.Now()},
	})
}
