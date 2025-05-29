// internal/services/comment_service.go
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

type CommentService struct {
	collection     *mongo.Collection
	postCollection *mongo.Collection
	userCollection *mongo.Collection
	likeCollection *mongo.Collection
	db             *mongo.Database
}

func NewCommentService() *CommentService {
	return &CommentService{
		collection:     config.DB.Collection("comments"),
		postCollection: config.DB.Collection("posts"),
		userCollection: config.DB.Collection("users"),
		likeCollection: config.DB.Collection("likes"),
		db:             config.DB,
	}
}

// CreateComment creates a new comment
func (cs *CommentService) CreateComment(userID primitive.ObjectID, req models.CreateCommentRequest) (*models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert post ID
	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		return nil, errors.New("invalid post ID")
	}

	// Check if post exists and comments are enabled
	var post models.Post
	err = cs.postCollection.FindOne(ctx, bson.M{
		"_id":              postID,
		"is_published":     true,
		"comments_enabled": true,
		"deleted_at":       bson.M{"$exists": false},
	}).Decode(&post)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("post not found or comments disabled")
		}
		return nil, err
	}

	// Convert parent comment ID if provided
	var parentCommentID *primitive.ObjectID
	if req.ParentCommentID != "" {
		if pID, err := primitive.ObjectIDFromHex(req.ParentCommentID); err == nil {
			parentCommentID = &pID
		}
	}

	// Convert mentioned user IDs
	var mentions []primitive.ObjectID
	for _, mentionStr := range req.Mentions {
		if mentionID, err := primitive.ObjectIDFromHex(mentionStr); err == nil {
			mentions = append(mentions, mentionID)
		}
	}

	// Create comment
	comment := &models.Comment{
		UserID:          userID,
		PostID:          postID,
		ParentCommentID: parentCommentID,
		Content:         req.Content,
		ContentType:     req.ContentType,
		Media:           req.Media,
		Mentions:        mentions,
		IsApproved:      true, // Auto-approve by default
	}

	// Set thread information
	if parentCommentID != nil {
		parentComment, err := cs.GetCommentByID(*parentCommentID, &userID)
		if err != nil {
			return nil, errors.New("parent comment not found")
		}
		comment.SetThreadInfo(parentComment)
	} else {
		comment.SetThreadInfo(nil)
	}

	comment.BeforeCreate()

	// Extract mentions from content if not provided
	if len(comment.Mentions) == 0 {
		mentionedUsernames := comment.GetMentionedUsernames()
		if len(mentionedUsernames) > 0 {
			userIDs, _ := cs.getUserIDsByUsernames(mentionedUsernames)
			comment.Mentions = userIDs
		}
	}

	result, err := cs.collection.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}

	comment.ID = result.InsertedID.(primitive.ObjectID)

	// Update post comments count
	cs.postCollection.UpdateOne(ctx, bson.M{"_id": postID}, bson.M{
		"$inc": bson.M{"comments_count": 1},
	})

	// Update parent comment replies count if this is a reply
	if parentCommentID != nil {
		cs.collection.UpdateOne(ctx, bson.M{"_id": parentCommentID}, bson.M{
			"$inc": bson.M{"replies_count": 1},
		})
	}

	// Update user's comments count
	go cs.updateUserCommentsCount(userID, true)

	// Create mention notifications
	if len(comment.Mentions) > 0 {
		go cs.createMentionNotifications(userID, comment.ID, comment.Mentions)
	}

	// Populate author information
	cs.populateCommentAuthor(comment)

	return comment, nil
}

// GetCommentByID retrieves a comment by ID
func (cs *CommentService) GetCommentByID(commentID primitive.ObjectID, currentUserID *primitive.ObjectID) (*models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var comment models.Comment
	err := cs.collection.FindOne(ctx, bson.M{
		"_id":        commentID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&comment)

	if err != nil {
		return nil, err
	}

	// Check if user can view this comment
	if !comment.CanViewComment() {
		return nil, errors.New("comment not accessible")
	}

	// Populate author information
	if err := cs.populateCommentAuthor(&comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

// GetPostComments retrieves comments for a specific post
func (cs *CommentService) GetPostComments(postID primitive.ObjectID, currentUserID *primitive.ObjectID, sortBy string, limit, skip int) ([]models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if post exists
	var post models.Post
	err := cs.postCollection.FindOne(ctx, bson.M{
		"_id":        postID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&post)

	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"post_id":     postID,
		"level":       0, // Only top-level comments
		"deleted_at":  bson.M{"$exists": false},
		"is_hidden":   false,
		"is_approved": true,
	}

	// Set sort order
	var sortOption bson.M
	switch sortBy {
	case "oldest":
		sortOption = bson.M{"created_at": 1}
	case "popular":
		sortOption = bson.M{"likes_count": -1, "created_at": -1}
	case "controversial":
		sortOption = bson.M{"vote_score": 1, "created_at": -1}
	default: // newest
		sortOption = bson.M{"created_at": -1}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(sortOption)

	cursor, err := cs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	// Populate author information for all comments
	for i := range comments {
		cs.populateCommentAuthor(&comments[i])
	}

	return comments, nil
}

// GetCommentReplies retrieves replies to a specific comment
func (cs *CommentService) GetCommentReplies(commentID primitive.ObjectID, currentUserID *primitive.ObjectID, limit, skip int) ([]models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if parent comment exists
	_, err := cs.GetCommentByID(commentID, currentUserID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"parent_comment_id": commentID,
		"deleted_at":        bson.M{"$exists": false},
		"is_hidden":         false,
		"is_approved":       true,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": 1}) // Oldest first for replies

	cursor, err := cs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var replies []models.Comment
	if err := cursor.All(ctx, &replies); err != nil {
		return nil, err
	}

	// Populate author information for all replies
	for i := range replies {
		cs.populateCommentAuthor(&replies[i])
	}

	return replies, nil
}

// UpdateComment updates an existing comment
func (cs *CommentService) UpdateComment(commentID, userID primitive.ObjectID, req models.UpdateCommentRequest) (*models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if comment exists and user owns it
	comment, err := cs.GetCommentByID(commentID, &userID)
	if err != nil {
		return nil, err
	}

	if !comment.CanEditComment(userID) {
		return nil, errors.New("access denied")
	}

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	// Update fields if provided
	if req.Content != nil {
		update["$set"].(bson.M)["content"] = *req.Content
		// Re-extract mentions if content changed
		mentionedUsernames := models.ExtractMentionsFromText(*req.Content)
		if len(mentionedUsernames) > 0 {
			userIDs, _ := cs.getUserIDsByUsernames(mentionedUsernames)
			update["$set"].(bson.M)["mentions"] = userIDs
		}
	}
	if req.Media != nil {
		update["$set"].(bson.M)["media"] = req.Media
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

	// Mark as edited
	update["$set"].(bson.M)["is_edited"] = true
	update["$set"].(bson.M)["edited_at"] = time.Now()

	_, err = cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, update)
	if err != nil {
		return nil, err
	}

	return cs.GetCommentByID(commentID, &userID)
}

// DeleteComment soft deletes a comment
func (cs *CommentService) DeleteComment(commentID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user role for permission check
	var user models.User
	err := cs.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	// Check if comment exists
	var comment models.Comment
	err = cs.collection.FindOne(ctx, bson.M{
		"_id":        commentID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&comment)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("comment not found")
		}
		return err
	}

	// Check if user can delete this comment
	var post models.Post
	cs.postCollection.FindOne(ctx, bson.M{"_id": comment.PostID}).Decode(&post)
	isPostAuthor := post.UserID == userID

	if !comment.CanDeleteComment(userID, user.Role, isPostAuthor) {
		return errors.New("access denied")
	}

	// Soft delete the comment
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at":  now,
			"updated_at":  now,
			"is_hidden":   true,
			"is_approved": false,
		},
	}

	_, err = cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, update)
	if err != nil {
		return err
	}

	// Update post comments count
	cs.postCollection.UpdateOne(ctx, bson.M{"_id": comment.PostID}, bson.M{
		"$inc": bson.M{"comments_count": -1},
	})

	// Update parent comment replies count if this is a reply
	if comment.ParentCommentID != nil {
		cs.collection.UpdateOne(ctx, bson.M{"_id": comment.ParentCommentID}, bson.M{
			"$inc": bson.M{"replies_count": -1},
		})
	}

	// Update user's comments count
	go cs.updateUserCommentsCount(comment.UserID, false)

	return nil
}

// LikeComment adds or updates a like on a comment
func (cs *CommentService) LikeComment(commentID, userID primitive.ObjectID, reactionType models.ReactionType) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if comment exists
	var comment models.Comment
	err := cs.collection.FindOne(ctx, bson.M{
		"_id":         commentID,
		"deleted_at":  bson.M{"$exists": false},
		"is_hidden":   false,
		"is_approved": true,
	}).Decode(&comment)

	if err != nil {
		return err
	}

	// Check if user already liked this comment
	var existingLike models.Like
	err = cs.likeCollection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   commentID,
		"target_type": "comment",
	}).Decode(&existingLike)

	if err == nil {
		// Update existing like
		update := bson.M{
			"$set": bson.M{
				"reaction_type": reactionType,
				"updated_at":    time.Now(),
			},
		}
		_, err = cs.likeCollection.UpdateOne(ctx, bson.M{"_id": existingLike.ID}, update)
	} else if err == mongo.ErrNoDocuments {
		// Create new like
		like := &models.Like{
			UserID:       userID,
			TargetID:     commentID,
			TargetType:   "comment",
			ReactionType: reactionType,
		}
		like.BeforeCreate()

		_, err = cs.likeCollection.InsertOne(ctx, like)
		if err != nil {
			return err
		}

		// Increment comment like count
		cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
			"$inc": bson.M{"likes_count": 1},
		})

		// Update comment quality score
		go cs.updateCommentQualityScore(commentID)
	}

	return err
}

// UnlikeComment removes a like from a comment
func (cs *CommentService) UnlikeComment(commentID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find and delete the like
	result, err := cs.likeCollection.DeleteOne(ctx, bson.M{
		"user_id":     userID,
		"target_id":   commentID,
		"target_type": "comment",
	})

	if err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		// Decrement comment like count
		cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
			"$inc": bson.M{"likes_count": -1},
		})

		// Update comment quality score
		go cs.updateCommentQualityScore(commentID)
	}

	return nil
}

// GetCommentLikes retrieves users who liked a comment
func (cs *CommentService) GetCommentLikes(commentID primitive.ObjectID, limit, skip int) ([]models.LikeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id":   commentID,
				"target_type": "comment",
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

	cursor, err := cs.likeCollection.Aggregate(ctx, pipeline)
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

// ReportComment reports a comment
func (cs *CommentService) ReportComment(commentID, reporterID primitive.ObjectID, reason models.ReportReason, description string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if comment exists
	var comment models.Comment
	err := cs.collection.FindOne(ctx, bson.M{
		"_id":        commentID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&comment)

	if err != nil {
		return err
	}

	// Create report
	report := &models.Report{
		ReporterID:  reporterID,
		TargetType:  "comment",
		TargetID:    commentID,
		Reason:      reason,
		Description: description,
	}
	report.BeforeCreate()

	reportCollection := cs.db.Collection("reports")
	_, err = reportCollection.InsertOne(ctx, report)
	if err != nil {
		return err
	}

	// Update comment report count
	_, err = cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
		"$inc": bson.M{"reports_count": 1},
		"$set": bson.M{"is_reported": true},
	})

	return err
}

// PinComment pins a comment (post author only)
func (cs *CommentService) PinComment(commentID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get comment and check if user is post author
	var comment models.Comment
	err := cs.collection.FindOne(ctx, bson.M{
		"_id":        commentID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&comment)

	if err != nil {
		return err
	}

	// Check if user is post author
	var post models.Post
	err = cs.postCollection.FindOne(ctx, bson.M{"_id": comment.PostID}).Decode(&post)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return errors.New("access denied: only post authors can pin comments")
	}

	// Pin the comment
	_, err = cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
		"$set": bson.M{
			"is_pinned":  true,
			"updated_at": time.Now(),
		},
	})

	return err
}

// UnpinComment unpins a comment (post author only)
func (cs *CommentService) UnpinComment(commentID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get comment and check if user is post author
	var comment models.Comment
	err := cs.collection.FindOne(ctx, bson.M{
		"_id":        commentID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&comment)

	if err != nil {
		return err
	}

	// Check if user is post author
	var post models.Post
	err = cs.postCollection.FindOne(ctx, bson.M{"_id": comment.PostID}).Decode(&post)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return errors.New("access denied: only post authors can unpin comments")
	}

	// Unpin the comment
	_, err = cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
		"$set": bson.M{
			"is_pinned":  false,
			"updated_at": time.Now(),
		},
	})

	return err
}

// GetUserComments retrieves comments made by a specific user
func (cs *CommentService) GetUserComments(userID primitive.ObjectID, currentUserID *primitive.ObjectID, limit, skip int) ([]models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":     userID,
		"deleted_at":  bson.M{"$exists": false},
		"is_hidden":   false,
		"is_approved": true,
	}

	// If not viewing own comments, only show public comments
	if currentUserID == nil || *currentUserID != userID {
		// Add privacy checks based on post visibility
		filter["$lookup"] = bson.M{
			"from":         "posts",
			"localField":   "post_id",
			"foreignField": "_id",
			"as":           "post",
		}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := cs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	// Populate author information for all comments
	for i := range comments {
		cs.populateCommentAuthor(&comments[i])
	}

	return comments, nil
}

// GetCommentThread retrieves a complete comment thread
func (cs *CommentService) GetCommentThread(commentID primitive.ObjectID, currentUserID *primitive.ObjectID) ([]models.Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the root comment
	rootComment, err := cs.GetCommentByID(commentID, currentUserID)
	if err != nil {
		return nil, err
	}

	// Get all comments in the same thread
	filter := bson.M{
		"$or": []bson.M{
			{"_id": commentID},
			{"root_comment_id": rootComment.ID},
		},
		"deleted_at":  bson.M{"$exists": false},
		"is_hidden":   false,
		"is_approved": true,
	}

	// If this is already a reply, get the root and all its replies
	if rootComment.RootCommentID != nil {
		filter = bson.M{
			"$or": []bson.M{
				{"_id": *rootComment.RootCommentID},
				{"root_comment_id": *rootComment.RootCommentID},
			},
			"deleted_at":  bson.M{"$exists": false},
			"is_hidden":   false,
			"is_approved": true,
		}
	}

	opts := options.Find().
		SetSort(bson.M{"level": 1, "created_at": 1})

	cursor, err := cs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	// Populate author information for all comments
	for i := range comments {
		cs.populateCommentAuthor(&comments[i])
	}

	return comments, nil
}

// Helper methods

func (cs *CommentService) populateCommentAuthor(comment *models.Comment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := cs.userCollection.FindOne(ctx, bson.M{"_id": comment.UserID}).Decode(&user)
	if err != nil {
		return err
	}

	comment.Author = user.ToUserResponse()
	return nil
}

func (cs *CommentService) updateUserCommentsCount(userID primitive.ObjectID, increment bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 1
	if !increment {
		value = -1
	}

	cs.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$inc": bson.M{"comments_count": value},
		"$set": bson.M{"updated_at": time.Now()},
	})
}

func (cs *CommentService) updateCommentQualityScore(commentID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var comment models.Comment
	if err := cs.collection.FindOne(ctx, bson.M{"_id": commentID}).Decode(&comment); err != nil {
		return
	}

	comment.UpdateQualityScore()

	cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
		"$set": bson.M{
			"quality_score": comment.QualityScore,
			"updated_at":    time.Now(),
		},
	})
}

func (cs *CommentService) getUserIDsByUsernames(usernames []string) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := cs.userCollection.Find(ctx, bson.M{
		"username": bson.M{"$in": usernames},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	var userIDs []primitive.ObjectID
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	return userIDs, nil
}

func (cs *CommentService) createMentionNotifications(authorID, commentID primitive.ObjectID, mentionedUsers []primitive.ObjectID) {
	// This would integrate with notification service to create mention notifications
	// Implementation depends on notification system
}

// VoteComment adds or updates a vote on a comment
func (cs *CommentService) VoteComment(commentID, userID primitive.ObjectID, voteType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if comment exists
	var comment models.Comment
	err := cs.collection.FindOne(ctx, bson.M{
		"_id":        commentID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&comment)

	if err != nil {
		return err
	}

	// Handle vote logic
	var updateFields bson.M
	switch voteType {
	case "upvote":
		updateFields = bson.M{
			"$inc": bson.M{"upvotes_count": 1},
			"$set": bson.M{"updated_at": time.Now()},
		}
	case "downvote":
		updateFields = bson.M{
			"$inc": bson.M{"downvotes_count": 1},
			"$set": bson.M{"updated_at": time.Now()},
		}
	case "remove":
		// Remove vote logic would need to track previous vote type
		updateFields = bson.M{
			"$set": bson.M{"updated_at": time.Now()},
		}
	default:
		return errors.New("invalid vote type")
	}

	_, err = cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, updateFields)
	if err != nil {
		return err
	}

	// Update vote score
	go cs.updateCommentVoteScore(commentID)

	return nil
}

func (cs *CommentService) updateCommentVoteScore(commentID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var comment models.Comment
	if err := cs.collection.FindOne(ctx, bson.M{"_id": commentID}).Decode(&comment); err != nil {
		return
	}

	voteScore := comment.UpvotesCount - comment.DownvotesCount

	cs.collection.UpdateOne(ctx, bson.M{"_id": commentID}, bson.M{
		"$set": bson.M{
			"vote_score": voteScore,
			"updated_at": time.Now(),
		},
	})
}
func (us *CommentService) GetCollection() *mongo.Collection {
	return us.collection
}
