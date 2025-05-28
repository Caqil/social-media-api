// internal/services/user_service.go
package services

import (
	"context"
	"errors"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewUserService() *UserService {
	return &UserService{
		collection: config.DB.Collection("users"),
		db:         config.DB,
	}
}

// CreateUser creates a new user
func (us *UserService) CreateUser(req models.RegisterRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if username or email already exists
	exists, err := us.CheckUserExists(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username or email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		Phone:       req.Phone,
	}

	user.BeforeCreate()

	result, err := us.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

// GetUserByID retrieves user by ID
func (us *UserService) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := us.collection.FindOne(ctx, bson.M{
		"_id":        userID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername retrieves user by username
func (us *UserService) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := us.collection.FindOne(ctx, bson.M{
		"username":   username,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves user by email
func (us *UserService) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := us.collection.FindOne(ctx, bson.M{
		"email":      email,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates user information
func (us *UserService) UpdateUser(userID primitive.ObjectID, req models.UpdateProfileRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if req.FirstName != nil {
		update["$set"].(bson.M)["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		update["$set"].(bson.M)["last_name"] = *req.LastName
	}
	if req.DisplayName != nil {
		update["$set"].(bson.M)["display_name"] = *req.DisplayName
	}
	if req.Bio != nil {
		update["$set"].(bson.M)["bio"] = *req.Bio
	}
	if req.Website != nil {
		update["$set"].(bson.M)["website"] = *req.Website
	}
	if req.Location != nil {
		update["$set"].(bson.M)["location"] = *req.Location
	}
	if req.DateOfBirth != nil {
		update["$set"].(bson.M)["date_of_birth"] = *req.DateOfBirth
	}
	if req.Gender != nil {
		update["$set"].(bson.M)["gender"] = *req.Gender
	}
	if req.Phone != nil {
		update["$set"].(bson.M)["phone"] = *req.Phone
	}
	if req.SocialLinks != nil {
		update["$set"].(bson.M)["social_links"] = req.SocialLinks
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return nil, err
	}

	return us.GetUserByID(userID)
}

// UpdateUserPrivacySettings updates user privacy settings
func (us *UserService) UpdateUserPrivacySettings(userID primitive.ObjectID, settings models.PrivacySettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"privacy_settings": settings,
			"updated_at":       time.Now(),
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// UpdateNotificationSettings updates user notification settings
func (us *UserService) UpdateNotificationSettings(userID primitive.ObjectID, settings models.NotificationSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"notification_settings": settings,
			"updated_at":            time.Now(),
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// ChangePassword changes user password
func (us *UserService) ChangePassword(userID primitive.ObjectID, req models.ChangePasswordRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get current user
	user, err := us.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	// Validate new password confirmation
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("new password and confirmation do not match")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	update := bson.M{
		"$set": bson.M{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		},
	}

	_, err = us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// SearchUsers searches for users
func (us *UserService) SearchUsers(query string, limit, skip int) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"username": bson.M{"$regex": query, "$options": "i"}},
					{"first_name": bson.M{"$regex": query, "$options": "i"}},
					{"last_name": bson.M{"$regex": query, "$options": "i"}},
					{"display_name": bson.M{"$regex": query, "$options": "i"}},
				},
			},
			{"is_active": true},
			{"deleted_at": bson.M{"$exists": false}},
		},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"followers_count": -1})

	cursor, err := us.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserStats retrieves user statistics
func (us *UserService) GetUserStats(userID primitive.ObjectID) (*models.Stats, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := us.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	stats := &models.Stats{
		PostsCount:     user.PostsCount,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
		LikesCount:     user.TotalLikesReceived,
		CommentsCount:  user.TotalCommentsReceived,
		SharesCount:    user.TotalSharesReceived,
		ViewsCount:     user.ProfileViews,
	}

	return stats, nil
}

// GetSuggestedUsers gets suggested users for a user
func (us *UserService) GetSuggestedUsers(userID primitive.ObjectID, limit int) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get users that the current user is not following
	// and have high engagement or mutual connections
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        bson.M{"$ne": userID},
				"is_active":  true,
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "follows",
				"localField":   "_id",
				"foreignField": "followee_id",
				"as":           "followers",
				"pipeline": []bson.M{
					{"$match": bson.M{"follower_id": userID}},
				},
			},
		},
		{
			"$match": bson.M{
				"followers": bson.M{"$size": 0}, // Not already following
			},
		},
		{
			"$sort": bson.M{
				"followers_count": -1,
				"posts_count":     -1,
			},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := us.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserActivity updates user's last activity
func (us *UserService) UpdateUserActivity(userID primitive.ObjectID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"last_active_at": time.Now(),
			"online_status":  status,
			"updated_at":     time.Now(),
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// UpdateUserCounts updates user's various counts
func (us *UserService) UpdateUserCounts(userID primitive.ObjectID, countType string, increment bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	var update bson.M
	switch countType {
	case "posts":
		update = bson.M{"$inc": bson.M{"posts_count": value}}
	case "followers":
		update = bson.M{"$inc": bson.M{"followers_count": value}}
	case "following":
		update = bson.M{"$inc": bson.M{"following_count": value}}
	case "likes":
		update = bson.M{"$inc": bson.M{"total_likes_received": value}}
	case "comments":
		update = bson.M{"$inc": bson.M{"total_comments_received": value}}
	case "shares":
		update = bson.M{"$inc": bson.M{"total_shares_received": value}}
	case "profile_views":
		update = bson.M{"$inc": bson.M{"profile_views": value}}
	default:
		return errors.New("invalid count type")
	}

	update["$set"] = bson.M{"updated_at": time.Now()}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// BlockUser blocks a user
func (us *UserService) BlockUser(userID, blockedUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$addToSet": bson.M{"blocked_users": blockedUserID},
		"$set":      bson.M{"updated_at": time.Now()},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// UnblockUser unblocks a user
func (us *UserService) UnblockUser(userID, blockedUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$pull": bson.M{"blocked_users": blockedUserID},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// GetBlockedUsers gets list of blocked users
func (us *UserService) GetBlockedUsers(userID primitive.ObjectID) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := us.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if len(user.BlockedUsers) == 0 {
		return []models.User{}, nil
	}

	filter := bson.M{
		"_id": bson.M{"$in": user.BlockedUsers},
	}

	cursor, err := us.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// SuspendUser suspends a user account
func (us *UserService) SuspendUser(userID primitive.ObjectID, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_suspended": true,
			"updated_at":   time.Now(),
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// UnsuspendUser unsuspends a user account
func (us *UserService) UnsuspendUser(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_suspended": false,
			"updated_at":   time.Now(),
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// VerifyUser verifies a user account
func (us *UserService) VerifyUser(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_verified": true,
			"updated_at":  time.Now(),
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// VerifyEmail verifies user's email
func (us *UserService) VerifyEmail(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"email_verified":    true,
			"email_verified_at": now,
			"updated_at":        now,
		},
		"$unset": bson.M{
			"email_verify_token": "",
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// CheckUserExists checks if username or email already exists
func (us *UserService) CheckUserExists(username, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": email},
		},
		"deleted_at": bson.M{"$exists": false},
	}

	count, err := us.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// SoftDeleteUser soft deletes a user
func (us *UserService) SoftDeleteUser(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"is_active":  false,
			"updated_at": now,
			"username":   "deleted_" + userID.Hex(),
			"email":      "deleted_" + userID.Hex() + "@deleted.com",
		},
	}

	_, err := us.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// GetUserProfile gets complete user profile with context
func (us *UserService) GetUserProfile(userID, currentUserID primitive.ObjectID) (*models.ProfileResponse, error) {
	user, err := us.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Get relationship context if different users
	var isFollowing, isFollowedBy, isFriend, isBlocked bool
	var mutualFriends int64

	if userID != currentUserID {
		// This would typically involve checking follows, blocks, etc.
		// For now, we'll set defaults
		isFollowing = false
		isFollowedBy = false
		isFriend = false
		isBlocked = false
		mutualFriends = 0
	}

	userResponse := user.ToUserResponseWithContext(currentUserID, isFollowing, isFollowedBy, isFriend, isBlocked, mutualFriends)

	profile := &models.ProfileResponse{
		UserResponse:          userResponse,
		TotalLikesReceived:    user.TotalLikesReceived,
		TotalCommentsReceived: user.TotalCommentsReceived,
		ProfileViews:          user.ProfileViews,
	}

	return profile, nil
}
