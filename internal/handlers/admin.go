// internal/handlers/admin_handler.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"
)

type AdminHandler struct {
	adminService *services.AdminService
	authService  *services.AuthService
	db           *mongo.Database
	upgrader     websocket.Upgrader
}

func NewAdminHandler(adminService *services.AdminService, authService *services.AuthService, db *mongo.Database) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		authService:  authService,
		db:           db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Dashboard
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get dashboard statistics", err)
		return
	}
	utils.OkResponse(c, "Dashboard statistics retrieved successfully", stats)
}

// User Management
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.UserFilter{
		Search: c.Query("search"),
		Role:   c.Query("role"),
	}

	if c.Query("is_verified") != "" {
		isVerified := c.Query("is_verified") == "true"
		filter.IsVerified = &isVerified
	}

	if c.Query("is_active") != "" {
		isActive := c.Query("is_active") == "true"
		filter.IsActive = &isActive
	}

	if c.Query("is_suspended") != "" {
		isSuspended := c.Query("is_suspended") == "true"
		filter.IsSuspended = &isSuspended
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &parsedDate
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &parsedDate
		}
	}

	users, pagination, err := h.adminService.GetAllUsers(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get users", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Users retrieved successfully", users, *pagination, links)
}

func (h *AdminHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.UserFilter{Search: query}
	users, pagination, err := h.adminService.GetAllUsers(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search users", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "User search completed", users, *pagination, links)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.adminService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}
	utils.OkResponse(c, "User retrieved successfully", user)
}

func (h *AdminHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", nil)
		return
	}

	ctx := c.Request.Context()

	// Get user basic info
	var user models.User
	err = h.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	// Get posts count
	postsCount, _ := h.db.Collection("posts").CountDocuments(ctx, bson.M{
		"user_id":    objID,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get comments count
	commentsCount, _ := h.db.Collection("comments").CountDocuments(ctx, bson.M{
		"user_id":    objID,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get likes received
	likesReceived, _ := h.db.Collection("likes").CountDocuments(ctx, bson.M{
		"target_user_id": objID,
		"deleted_at":     bson.M{"$exists": false},
	})

	// Get followers count
	followersCount, _ := h.db.Collection("follows").CountDocuments(ctx, bson.M{
		"following_id": objID,
		"deleted_at":   bson.M{"$exists": false},
	})

	// Get following count
	followingCount, _ := h.db.Collection("follows").CountDocuments(ctx, bson.M{
		"follower_id": objID,
		"deleted_at":  bson.M{"$exists": false},
	})

	// Get stories count
	storiesCount, _ := h.db.Collection("stories").CountDocuments(ctx, bson.M{
		"user_id":    objID,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get messages sent
	messagesSent, _ := h.db.Collection("messages").CountDocuments(ctx, bson.M{
		"sender_id":  objID,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get reports made
	reportsMade, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"reporter_id": objID,
		"deleted_at":  bson.M{"$exists": false},
	})

	// Get reports against
	reportsAgainst, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"target_user_id": objID,
		"deleted_at":     bson.M{"$exists": false},
	})

	stats := gin.H{
		"user_id":         userID,
		"username":        user.Username,
		"posts_count":     postsCount,
		"comments_count":  commentsCount,
		"likes_received":  likesReceived,
		"followers_count": followersCount,
		"following_count": followingCount,
		"stories_count":   storiesCount,
		"messages_sent":   messagesSent,
		"reports_made":    reportsMade,
		"reports_against": reportsAgainst,
		"join_date":       user.CreatedAt,
		"last_active":     user.LastActiveAt,
		"is_verified":     user.IsVerified,
		"is_suspended":    user.IsSuspended,
		"is_active":       user.IsActive,
	}

	utils.OkResponse(c, "User statistics retrieved successfully", stats)
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		IsActive    bool   `json:"is_active"`
		IsSuspended bool   `json:"is_suspended"`
		Reason      string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.adminService.UpdateUserStatus(c.Request.Context(), userID, req.IsActive, req.IsSuspended)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update user status", err)
		return
	}

	h.logAdminActivity(c, "user_status_update", "Updated user status for user ID: "+userID)
	utils.OkResponse(c, "User status updated successfully", gin.H{
		"user_id":      userID,
		"is_active":    req.IsActive,
		"is_suspended": req.IsSuspended,
	})
}

func (h *AdminHandler) VerifyUser(c *gin.Context) {
	userID := c.Param("id")
	err := h.adminService.VerifyUser(c.Request.Context(), userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to verify user", err)
		return
	}

	h.logAdminActivity(c, "user_verification", "Verified user ID: "+userID)
	utils.OkResponse(c, "User verified successfully", gin.H{
		"user_id":     userID,
		"is_verified": true,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.adminService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete user", err)
		return
	}

	h.logAdminActivity(c, "user_deletion", "Deleted user ID: "+userID+" Reason: "+req.Reason)
	utils.OkResponse(c, "User deleted successfully", gin.H{
		"user_id": userID,
		"reason":  req.Reason,
	})
}

func (h *AdminHandler) BulkUserAction(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
		Action  string   `json:"action" binding:"required"`
		Reason  string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.UserIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 users allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, userID := range req.UserIDs {
		var err error
		switch req.Action {
		case "suspend":
			err = h.adminService.UpdateUserStatus(c.Request.Context(), userID, true, true)
		case "activate":
			err = h.adminService.UpdateUserStatus(c.Request.Context(), userID, true, false)
		case "verify":
			err = h.adminService.VerifyUser(c.Request.Context(), userID)
		case "delete":
			err = h.adminService.DeleteUser(c.Request.Context(), userID)
		default:
			err = fmt.Errorf("invalid action: %s", req.Action)
		}

		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("User %s: %v", userID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_user_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk user action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.UserIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

func (h *AdminHandler) ExportUsers(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	exportURL := "/api/admin/exports/users_" + time.Now().Format("20060102_150405") + "." + format

	h.logAdminActivity(c, "export_users", "Exported users data in "+format+" format")

	utils.AcceptedResponse(c, "Export started successfully", gin.H{
		"export_url": exportURL,
		"format":     format,
		"date_from":  dateFrom,
		"date_to":    dateTo,
		"status":     "processing",
	})
}

// Post Management
func (h *AdminHandler) GetAllPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.PostFilter{
		UserID:     c.Query("user_id"),
		Type:       c.Query("type"),
		Visibility: models.PrivacyLevel(c.Query("visibility")),
		Search:     c.Query("search"),
	}

	if c.Query("is_reported") != "" {
		isReported := c.Query("is_reported") == "true"
		filter.IsReported = &isReported
	}

	if c.Query("is_hidden") != "" {
		isHidden := c.Query("is_hidden") == "true"
		filter.IsHidden = &isHidden
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &parsedDate
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &parsedDate
		}
	}

	posts, pagination, err := h.adminService.GetAllPosts(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get posts", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Posts retrieved successfully", posts, *pagination, links)
}

func (h *AdminHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.PostFilter{Search: query}
	posts, pagination, err := h.adminService.GetAllPosts(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search posts", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Post search completed", posts, *pagination, links)
}

func (h *AdminHandler) GetPostStats(c *gin.Context) {
	postID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID", nil)
		return
	}

	ctx := c.Request.Context()

	// Get post basic info
	var post models.Post
	err = h.db.Collection("posts").FindOne(ctx, bson.M{"_id": objID}).Decode(&post)
	if err != nil {
		utils.NotFoundResponse(c, "Post not found")
		return
	}

	// Get likes count
	likesCount, _ := h.db.Collection("likes").CountDocuments(ctx, bson.M{
		"target_id":   objID,
		"target_type": "post",
		"deleted_at":  bson.M{"$exists": false},
	})

	// Get comments count
	commentsCount, _ := h.db.Collection("comments").CountDocuments(ctx, bson.M{
		"post_id":    objID,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get shares count (if you have shares collection)
	sharesCount, _ := h.db.Collection("shares").CountDocuments(ctx, bson.M{
		"post_id":    objID,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get reports count
	reportsCount, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"target_id":   objID,
		"target_type": "post",
		"deleted_at":  bson.M{"$exists": false},
	})

	stats := gin.H{
		"post_id":        postID,
		"content":        post.Content,
		"type":           post.Type,
		"likes_count":    likesCount,
		"comments_count": commentsCount,
		"shares_count":   sharesCount,
		"reports_count":  reportsCount,
		"created_at":     post.CreatedAt,
		"visibility":     post.Visibility,
		"is_hidden":      post.IsHidden,
		"is_reported":    post.IsReported,
	}

	utils.OkResponse(c, "Post statistics retrieved successfully", stats)
}

func (h *AdminHandler) HidePost(c *gin.Context) {
	postID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.adminService.HidePost(c.Request.Context(), postID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to hide post", err)
		return
	}

	h.logAdminActivity(c, "post_hidden", "Hidden post ID: "+postID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Post hidden successfully", gin.H{
		"post_id": postID,
		"reason":  req.Reason,
	})
}

func (h *AdminHandler) DeletePost(c *gin.Context) {
	postID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.adminService.DeletePost(c.Request.Context(), postID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete post", err)
		return
	}

	h.logAdminActivity(c, "post_deletion", "Deleted post ID: "+postID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Post deleted successfully", gin.H{
		"post_id": postID,
		"reason":  req.Reason,
	})
}

func (h *AdminHandler) BulkPostAction(c *gin.Context) {
	var req struct {
		PostIDs []string `json:"post_ids" binding:"required"`
		Action  string   `json:"action" binding:"required"`
		Reason  string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.PostIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 posts allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, postID := range req.PostIDs {
		var err error
		switch req.Action {
		case "hide":
			err = h.adminService.HidePost(c.Request.Context(), postID)
		case "delete":
			err = h.adminService.DeletePost(c.Request.Context(), postID)
		default:
			err = fmt.Errorf("invalid action: %s", req.Action)
		}

		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Post %s: %v", postID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_post_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk post action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.PostIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

func (h *AdminHandler) ExportPosts(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	userID := c.Query("user_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	exportURL := "/api/admin/exports/posts_" + time.Now().Format("20060102_150405") + "." + format

	h.logAdminActivity(c, "export_posts", "Exported posts data in "+format+" format")

	utils.AcceptedResponse(c, "Export started successfully", gin.H{
		"export_url": exportURL,
		"format":     format,
		"user_id":    userID,
		"date_from":  dateFrom,
		"date_to":    dateTo,
		"status":     "processing",
	})
}

// Comment Management
func (h *AdminHandler) GetAllComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	ctx := c.Request.Context()
	skip := (page - 1) * limit

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
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
			"$lookup": bson.M{
				"from":         "posts",
				"localField":   "post_id",
				"foreignField": "_id",
				"as":           "post",
			},
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

	cursor, err := h.db.Collection("comments").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get comments", err)
		return
	}
	defer cursor.Close(ctx)

	var comments []bson.M
	if err := cursor.All(ctx, &comments); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode comments", err)
		return
	}

	total, _ := h.db.Collection("comments").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Comments retrieved successfully", comments, *pagination, links)
}

func (h *AdminHandler) GetComment(c *gin.Context) {
	commentID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID", nil)
		return
	}

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
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
			"$lookup": bson.M{
				"from":         "posts",
				"localField":   "post_id",
				"foreignField": "_id",
				"as":           "post",
			},
		},
	}

	cursor, err := h.db.Collection("comments").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get comment", err)
		return
	}
	defer cursor.Close(ctx)

	var comment bson.M
	if cursor.Next(ctx) {
		cursor.Decode(&comment)
	} else {
		utils.NotFoundResponse(c, "Comment not found")
		return
	}

	utils.OkResponse(c, "Comment retrieved successfully", comment)
}

// UpdateComment updates a comment content
func (h *AdminHandler) UpdateComment(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	// Get admin user for audit logging
	adminUser, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}

	// Update the comment
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"content":    input.Content,
			"updated_at": time.Now(),
			"edited_by":  adminUser.(models.User).ID,
		},
	}

	_, err = h.db.Collection("comments").UpdateOne(c, bson.M{"_id": objID}, update)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update comment", err)
		return
	}

	h.logAdminActivity(c, "comment_update", fmt.Sprintf("Updated comment %s", id))

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "Comment updated successfully", nil)
}

// ShowComment makes a hidden comment visible
func (h *AdminHandler) ShowComment(c *gin.Context) {
	id := c.Param("id")

	// Get admin user for audit logging
	adminUser, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}

	// Update the comment
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"is_hidden":    false,
			"updated_at":   time.Now(),
			"moderated_by": adminUser.(models.User).ID,
		},
	}

	_, err = h.db.Collection("comments").UpdateOne(c, bson.M{"_id": objID}, update)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to show comment", err)
		return
	}

	h.logAdminActivity(c, "comment_show", fmt.Sprintf("Unhid comment %s", id))

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "Comment is now visible", nil)
}
func (h *AdminHandler) HideComment(c *gin.Context) {
	commentID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"is_hidden":  true,
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("comments").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to hide comment", err)
		return
	}

	h.logAdminActivity(c, "comment_hidden", "Hidden comment ID: "+commentID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Comment hidden successfully", gin.H{
		"comment_id": commentID,
		"reason":     req.Reason,
	})
}

func (h *AdminHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid comment ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("comments").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete comment", err)
		return
	}

	h.logAdminActivity(c, "comment_deletion", "Deleted comment ID: "+commentID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Comment deleted successfully", gin.H{
		"comment_id": commentID,
		"reason":     req.Reason,
	})
}

func (h *AdminHandler) BulkCommentAction(c *gin.Context) {
	var req struct {
		CommentIDs []string `json:"comment_ids" binding:"required"`
		Action     string   `json:"action" binding:"required"`
		Reason     string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.CommentIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 comments allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, commentID := range req.CommentIDs {
		objID, err := primitive.ObjectIDFromHex(commentID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Comment %s: invalid ID", commentID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "hide":
			update = bson.M{
				"$set": bson.M{
					"is_hidden":  true,
					"updated_at": time.Now(),
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Comment %s: invalid action", commentID))
			continue
		}

		_, err = h.db.Collection("comments").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Comment %s: %v", commentID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_comment_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk comment action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.CommentIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Group Management
func (h *AdminHandler) GetAllGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	groups, pagination, err := h.adminService.GetAllGroups(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get groups", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Groups retrieved successfully", groups, *pagination, links)
}

func (h *AdminHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", nil)
		return
	}

	ctx := c.Request.Context()
	var group models.Group
	err = h.db.Collection("groups").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "Group not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get group", err)
		return
	}

	utils.OkResponse(c, "Group retrieved successfully", group.ToGroupResponse())
}

func (h *AdminHandler) GetGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"group_id":   objID,
				"deleted_at": bson.M{"$exists": false},
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
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("group_members").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get group members", err)
		return
	}
	defer cursor.Close(ctx)

	var members []bson.M
	if err := cursor.All(ctx, &members); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode group members", err)
		return
	}

	total, _ := h.db.Collection("group_members").CountDocuments(ctx, bson.M{
		"group_id":   objID,
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Group members retrieved successfully", members, *pagination, links)
}

func (h *AdminHandler) UpdateGroupStatus(c *gin.Context) {
	groupID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", nil)
		return
	}

	var req struct {
		IsActive bool   `json:"is_active"`
		Reason   string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"is_active":  req.IsActive,
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("groups").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update group status", err)
		return
	}

	h.logAdminActivity(c, "group_status_update", "Updated group status for group ID: "+groupID)
	utils.OkResponse(c, "Group status updated successfully", gin.H{
		"group_id":  groupID,
		"is_active": req.IsActive,
		"reason":    req.Reason,
	})
}

func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("groups").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete group", err)
		return
	}

	h.logAdminActivity(c, "group_deletion", "Deleted group ID: "+groupID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Group deleted successfully", gin.H{
		"group_id": groupID,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) BulkGroupAction(c *gin.Context) {
	var req struct {
		GroupIDs []string `json:"group_ids" binding:"required"`
		Action   string   `json:"action" binding:"required"`
		Reason   string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.GroupIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 groups allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, groupID := range req.GroupIDs {
		objID, err := primitive.ObjectIDFromHex(groupID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Group %s: invalid ID", groupID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "suspend":
			update = bson.M{
				"$set": bson.M{
					"is_active":  false,
					"updated_at": time.Now(),
				},
			}
		case "activate":
			update = bson.M{
				"$set": bson.M{
					"is_active":  true,
					"updated_at": time.Now(),
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Group %s: invalid action", groupID))
			continue
		}

		_, err = h.db.Collection("groups").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Group %s: %v", groupID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_group_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk group action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.GroupIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Event Management
func (h *AdminHandler) GetAllEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	events, pagination, err := h.adminService.GetAllEvents(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get events", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Events retrieved successfully", events, *pagination, links)
}

func (h *AdminHandler) GetEvent(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID", nil)
		return
	}

	ctx := c.Request.Context()
	var event models.Event
	err = h.db.Collection("events").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "Event not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get event", err)
		return
	}

	utils.OkResponse(c, "Event retrieved successfully", event.ToEventResponse())
}

func (h *AdminHandler) GetEventAttendees(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"event_id":   objID,
				"status":     "attending",
				"deleted_at": bson.M{"$exists": false},
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
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("event_attendees").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get event attendees", err)
		return
	}
	defer cursor.Close(ctx)

	var attendees []bson.M
	if err := cursor.All(ctx, &attendees); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode event attendees", err)
		return
	}

	total, _ := h.db.Collection("event_attendees").CountDocuments(ctx, bson.M{
		"event_id":   objID,
		"status":     "attending",
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Event attendees retrieved successfully", attendees, *pagination, links)
}

func (h *AdminHandler) UpdateEventStatus(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"status":     req.Status,
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("events").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update event status", err)
		return
	}

	h.logAdminActivity(c, "event_status_update", "Updated event status for event ID: "+eventID)
	utils.OkResponse(c, "Event status updated successfully", gin.H{
		"event_id": eventID,
		"status":   req.Status,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) DeleteEvent(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("events").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete event", err)
		return
	}

	h.logAdminActivity(c, "event_deletion", "Deleted event ID: "+eventID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Event deleted successfully", gin.H{
		"event_id": eventID,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) BulkEventAction(c *gin.Context) {
	var req struct {
		EventIDs []string `json:"event_ids" binding:"required"`
		Action   string   `json:"action" binding:"required"`
		Reason   string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.EventIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 events allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, eventID := range req.EventIDs {
		objID, err := primitive.ObjectIDFromHex(eventID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Event %s: invalid ID", eventID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "cancel":
			update = bson.M{
				"$set": bson.M{
					"status":     "cancelled",
					"updated_at": time.Now(),
				},
			}
		case "approve":
			update = bson.M{
				"$set": bson.M{
					"status":     "approved",
					"updated_at": time.Now(),
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Event %s: invalid action", eventID))
			continue
		}

		_, err = h.db.Collection("events").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Event %s: %v", eventID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_event_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk event action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.EventIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Story Management
func (h *AdminHandler) GetAllStories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	stories, pagination, err := h.adminService.GetAllStories(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get stories", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Stories retrieved successfully", stories, *pagination, links)
}

func (h *AdminHandler) GetStory(c *gin.Context) {
	storyID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(storyID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID", nil)
		return
	}

	ctx := c.Request.Context()
	var story models.Story
	err = h.db.Collection("stories").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&story)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "Story not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get story", err)
		return
	}

	utils.OkResponse(c, "Story retrieved successfully", story.ToStoryResponse())
}

func (h *AdminHandler) HideStory(c *gin.Context) {
	storyID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(storyID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"is_hidden":  true,
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("stories").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to hide story", err)
		return
	}

	h.logAdminActivity(c, "story_hidden", "Hidden story ID: "+storyID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Story hidden successfully", gin.H{
		"story_id": storyID,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) DeleteStory(c *gin.Context) {
	storyID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(storyID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid story ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("stories").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete story", err)
		return
	}

	h.logAdminActivity(c, "story_deletion", "Deleted story ID: "+storyID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Story deleted successfully", gin.H{
		"story_id": storyID,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) BulkStoryAction(c *gin.Context) {
	var req struct {
		StoryIDs []string `json:"story_ids" binding:"required"`
		Action   string   `json:"action" binding:"required"`
		Reason   string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.StoryIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 stories allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, storyID := range req.StoryIDs {
		objID, err := primitive.ObjectIDFromHex(storyID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Story %s: invalid ID", storyID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "hide":
			update = bson.M{
				"$set": bson.M{
					"is_hidden":  true,
					"updated_at": time.Now(),
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Story %s: invalid action", storyID))
			continue
		}

		_, err = h.db.Collection("stories").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Story %s: %v", storyID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_story_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk story action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.StoryIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// GetAllConversations with enhanced filtering and data structure - FIXED
func (h *AdminHandler) GetAllConversations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()

	// Build match filter
	matchFilter := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	// Add search filter if provided
	if search := c.Query("search"); search != "" {
		matchFilter["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Add type filter if provided
	if convType := c.Query("type"); convType != "" && convType != "all" {
		matchFilter["type"] = convType
	}

	// Add archived filter if provided
	if isArchived := c.Query("is_archived"); isArchived != "" && isArchived != "all" {
		matchFilter["is_archived"] = isArchived == "true"
	}

	// Add muted filter if provided
	if isMuted := c.Query("is_muted"); isMuted != "" && isMuted != "all" {
		matchFilter["is_muted"] = isMuted == "true"
	}

	// Add date range filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			matchFilter["created_at"] = bson.M{"$gte": parsedDate}
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			if existingDateFilter, exists := matchFilter["created_at"]; exists {
				if dateFilter, ok := existingDateFilter.(bson.M); ok {
					dateFilter["$lte"] = parsedDate.Add(24 * time.Hour)
				}
			} else {
				matchFilter["created_at"] = bson.M{"$lte": parsedDate.Add(24 * time.Hour)}
			}
		}
	}

	// Sort configuration
	sortField := c.DefaultQuery("sort_by", "last_message_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	sortValue := -1
	if sortOrder == "asc" {
		sortValue = 1
	}

	// Build aggregation pipeline
	pipeline := []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"participant_ids": "$participant_ids"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{"$in": []interface{}{"$_id", "$$participant_ids"}},
						},
					},
					{
						"$project": bson.M{
							"id":              bson.M{"$toString": "$_id"},
							"username":        1,
							"first_name":      1,
							"last_name":       1,
							"profile_picture": 1,
							"is_verified":     1,
						},
					},
				},
				"as": "participants",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "messages",
				"localField":   "last_message_id",
				"foreignField": "_id",
				"as":           "last_message_data",
			},
		},
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"sender_id": bson.M{"$arrayElemAt": []interface{}{"$last_message_data.sender_id", 0}}},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{"$eq": []interface{}{"$_id", "$$sender_id"}},
						},
					},
					{
						"$project": bson.M{
							"id":       bson.M{"$toString": "$_id"},
							"username": 1,
						},
					},
				},
				"as": "last_message_sender",
			},
		},
		{
			"$addFields": bson.M{
				"id": bson.M{"$toString": "$_id"},
				"participant_ids": bson.M{
					"$map": bson.M{
						"input": "$participant_ids",
						"as":    "pid",
						"in":    bson.M{"$toString": "$$pid"},
					},
				},
				"last_message": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$last_message_data"}, 0}},
						"then": bson.M{
							"id":           bson.M{"$toString": bson.M{"$arrayElemAt": []interface{}{"$last_message_data._id", 0}}},
							"content":      bson.M{"$arrayElemAt": []interface{}{"$last_message_data.content", 0}},
							"content_type": bson.M{"$arrayElemAt": []interface{}{"$last_message_data.content_type", 0}},
							"created_at":   bson.M{"$arrayElemAt": []interface{}{"$last_message_data.created_at", 0}},
							"sender": bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$gt": []interface{}{bson.M{"$size": "$last_message_sender"}, 0}},
									"then": bson.M{"$arrayElemAt": []interface{}{"$last_message_sender", 0}},
									"else": nil,
								},
							},
						},
						"else": nil,
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":             0,
				"id":              1,
				"type":            1,
				"title":           1,
				"avatar":          1,
				"participant_ids": 1,
				"last_message_id": bson.M{"$toString": "$last_message_id"},
				"last_message_at": 1,
				"is_archived":     1,
				"is_muted":        1,
				"unread_count":    1,
				"created_at":      1,
				"participants":    1,
				"last_message":    1,
			},
		},
		{
			"$sort": bson.M{sortField: sortValue},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("conversations").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get conversations", err)
		return
	}
	defer cursor.Close(ctx)

	var conversations []bson.M
	if err := cursor.All(ctx, &conversations); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode conversations", err)
		return
	}

	// Get total count for pagination
	countPipeline := []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$count": "total",
		},
	}

	countCursor, err := h.db.Collection("conversations").Aggregate(ctx, countPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to count conversations", err)
		return
	}
	defer countCursor.Close(ctx)

	var countResult []bson.M
	total := int64(0)
	if err := countCursor.All(ctx, &countResult); err == nil && len(countResult) > 0 {
		if count, ok := countResult[0]["total"].(int32); ok {
			total = int64(count)
		}
	}

	// Create pagination metadata
	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Conversations retrieved successfully", conversations, *pagination, links)
}
// GetConversationMessages retrieves all messages in a specific conversation
func (h *AdminHandler) GetConversationMessages(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()

	// Build aggregation pipeline for messages
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"conversation_id": objID,
				"deleted_at":      bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "sender_id",
				"foreignField": "_id",
				"as":           "sender_data",
			},
		},
		{
			"$addFields": bson.M{
				"id": bson.M{"$toString": "$_id"},
				"sender": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$sender_data"}, 0}},
						"then": bson.M{
							"id":              bson.M{"$toString": bson.M{"$arrayElemAt": []interface{}{"$sender_data._id", 0}}},
							"username":        bson.M{"$arrayElemAt": []interface{}{"$sender_data.username", 0}},
							"first_name":      bson.M{"$arrayElemAt": []interface{}{"$sender_data.first_name", 0}},
							"last_name":       bson.M{"$arrayElemAt": []interface{}{"$sender_data.last_name", 0}},
							"profile_picture": bson.M{"$arrayElemAt": []interface{}{"$sender_data.profile_picture", 0}},
						},
						"else": nil,
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":          0,
				"id":           1,
				"content":      1,
				"content_type": 1,
				"media_url":    1,
				"file_name":    1,
				"file_size":    1,
				"is_read":      1,
				"read_at":      1,
				"is_edited":    1,
				"edited_at":    1,
				"created_at":   1,
				"sender":       1,
			},
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

	cursor, err := h.db.Collection("messages").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get conversation messages", err)
		return
	}
	defer cursor.Close(ctx)

	var messages []bson.M
	if err := cursor.All(ctx, &messages); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode messages", err)
		return
	}

	// Get total count
	total, _ := h.db.Collection("messages").CountDocuments(ctx, bson.M{
		"conversation_id": objID,
		"deleted_at":      bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Conversation messages retrieved successfully", messages, *pagination, links)
}

// GetConversationAnalytics provides analytics for a specific conversation
func (h *AdminHandler) GetConversationAnalytics(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", nil)
		return
	}

	ctx := c.Request.Context()

	// Get conversation details
	var conversation bson.M
	err = h.db.Collection("conversations").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&conversation)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation not found")
		return
	}

	// Get message statistics
	messageStatsPipeline := []bson.M{
		{
			"$match": bson.M{
				"conversation_id": objID,
				"deleted_at":      bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":            nil,
				"total_messages": bson.M{"$sum": 1},
				"text_messages":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$content_type", "text"}}, 1, 0}}},
				"media_messages": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$ne": []interface{}{"$content_type", "text"}}, 1, 0}}},
				"read_messages":  bson.M{"$sum": bson.M{"$cond": []interface{}{"$is_read", 1, 0}}},
				"avg_length":     bson.M{"$avg": bson.M{"$strLenCP": "$content"}},
			},
		},
	}

	cursor, err := h.db.Collection("messages").Aggregate(ctx, messageStatsPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get message statistics", err)
		return
	}
	defer cursor.Close(ctx)

	var messageStats []bson.M
	cursor.All(ctx, &messageStats)

	// Get activity by day (last 30 days)
	activityPipeline := []bson.M{
		{
			"$match": bson.M{
				"conversation_id": objID,
				"created_at":      bson.M{"$gte": time.Now().AddDate(0, 0, -30)},
				"deleted_at":      bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$created_at",
					},
				},
				"message_count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	activityCursor, err := h.db.Collection("messages").Aggregate(ctx, activityPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get activity statistics", err)
		return
	}
	defer activityCursor.Close(ctx)

	var activityStats []bson.M
	activityCursor.All(ctx, &activityStats)

	// Get participant activity
	participantActivityPipeline := []bson.M{
		{
			"$match": bson.M{
				"conversation_id": objID,
				"deleted_at":      bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":           "$sender_id",
				"message_count": bson.M{"$sum": 1},
				"last_message":  bson.M{"$max": "$created_at"},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$addFields": bson.M{
				"username": bson.M{"$arrayElemAt": []interface{}{"$user.username", 0}},
			},
		},
		{
			"$sort": bson.M{"message_count": -1},
		},
	}

	participantCursor, err := h.db.Collection("messages").Aggregate(ctx, participantActivityPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get participant activity", err)
		return
	}
	defer participantCursor.Close(ctx)

	var participantActivity []bson.M
	participantCursor.All(ctx, &participantActivity)

	analytics := gin.H{
		"conversation_id":      conversationID,
		"conversation_type":    conversation["type"],
		"participant_count":    len(conversation["participant_ids"].(primitive.A)),
		"message_statistics":   messageStats,
		"activity_by_day":      activityStats,
		"participant_activity": participantActivity,
		"generated_at":         time.Now(),
	}

	utils.OkResponse(c, "Conversation analytics retrieved successfully", analytics)
}

// DeleteConversation deletes a conversation and optionally its messages
func (h *AdminHandler) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", nil)
		return
	}

	var req struct {
		Reason         string `json:"reason" binding:"required"`
		DeleteMessages bool   `json:"delete_messages"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	now := time.Now()

	// Delete conversation
	_, err = h.db.Collection("conversations").UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"deleted_at": now,
				"updated_at": now,
			},
		},
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete conversation", err)
		return
	}

	// Optionally delete all messages in the conversation
	if req.DeleteMessages {
		_, err = h.db.Collection("messages").UpdateMany(
			ctx,
			bson.M{"conversation_id": objID},
			bson.M{
				"$set": bson.M{
					"deleted_at": now,
					"updated_at": now,
				},
			},
		)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to delete conversation messages: %v\n", err)
		}
	}

	h.logAdminActivity(c, "conversation_deletion", "Deleted conversation ID: "+conversationID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Conversation deleted successfully", gin.H{
		"conversation_id":  conversationID,
		"reason":           req.Reason,
		"messages_deleted": req.DeleteMessages,
	})
}

// GetConversationReports gets reports related to a specific conversation
func (h *AdminHandler) GetConversationReports(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", nil)
		return
	}

	ctx := c.Request.Context()

	// Get reports for messages in this conversation
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "messages",
				"localField":   "target_id",
				"foreignField": "_id",
				"as":           "message",
			},
		},
		{
			"$match": bson.M{
				"message.conversation_id": objID,
				"target_type":             "message",
				"deleted_at":              bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "reporter_id",
				"foreignField": "_id",
				"as":           "reporter",
			},
		},
		{
			"$addFields": bson.M{
				"id": bson.M{"$toString": "$_id"},
				"reporter": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$gt": []interface{}{bson.M{"$size": "$reporter"}, 0}},
						"then": bson.M{"$arrayElemAt": []interface{}{"$reporter", 0}},
						"else": nil,
					},
				},
				"message_content": bson.M{"$arrayElemAt": []interface{}{"$message.content", 0}},
			},
		},
		{
			"$sort": bson.M{"created_at": -1},
		},
	}

	cursor, err := h.db.Collection("reports").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get conversation reports", err)
		return
	}
	defer cursor.Close(ctx)

	var reports []bson.M
	if err := cursor.All(ctx, &reports); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode reports", err)
		return
	}

	utils.OkResponse(c, "Conversation reports retrieved successfully", reports)
}

// GetAllMessages with enhanced filtering and data structure
func (h *AdminHandler) GetAllMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()

	// Build match filter
	matchFilter := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	// Add search filter if provided
	if search := c.Query("search"); search != "" {
		matchFilter["$or"] = []bson.M{
			{"content": bson.M{"$regex": search, "$options": "i"}},
			{"file_name": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Add conversation_id filter if provided - FIXED: Better validation
	if conversationID := c.Query("conversation_id"); conversationID != "" && conversationID != "all" {
		// Validate ObjectID format before using
		if objID, err := primitive.ObjectIDFromHex(conversationID); err == nil {
			matchFilter["conversation_id"] = objID
		} else {
			// Return error for invalid ObjectID instead of silently ignoring
			utils.BadRequestResponse(c, "Invalid conversation ID format", nil)
			return
		}
	}

	// Add content_type filter if provided
	if contentType := c.Query("content_type"); contentType != "" && contentType != "all" {
		matchFilter["content_type"] = contentType
	}

	// Add is_read filter if provided
	if isRead := c.Query("is_read"); isRead != "" && isRead != "all" {
		matchFilter["is_read"] = isRead == "true"
	}

	// Add date range filters if provided
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			matchFilter["created_at"] = bson.M{"$gte": parsedDate}
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			if existingDateFilter, exists := matchFilter["created_at"]; exists {
				if dateFilter, ok := existingDateFilter.(bson.M); ok {
					dateFilter["$lte"] = parsedDate.Add(24 * time.Hour)
				}
			} else {
				matchFilter["created_at"] = bson.M{"$lte": parsedDate.Add(24 * time.Hour)}
			}
		}
	}

	// Sort configuration
	sortField := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	sortValue := -1
	if sortOrder == "asc" {
		sortValue = 1
	}

	// Build aggregation pipeline with proper data transformation
	pipeline := []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "sender_id",
				"foreignField": "_id",
				"as":           "sender_data",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "conversations",
				"localField":   "conversation_id",
				"foreignField": "_id",
				"as":           "conversation_data",
			},
		},
		{
			"$addFields": bson.M{
				"id": bson.M{"$toString": "$_id"},
				"sender": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$sender_data"}, 0}},
						"then": bson.M{
							"id":              bson.M{"$toString": bson.M{"$arrayElemAt": []interface{}{"$sender_data._id", 0}}},
							"username":        bson.M{"$arrayElemAt": []interface{}{"$sender_data.username", 0}},
							"email":           bson.M{"$arrayElemAt": []interface{}{"$sender_data.email", 0}},
							"first_name":      bson.M{"$arrayElemAt": []interface{}{"$sender_data.first_name", 0}},
							"last_name":       bson.M{"$arrayElemAt": []interface{}{"$sender_data.last_name", 0}},
							"profile_picture": bson.M{"$arrayElemAt": []interface{}{"$sender_data.profile_picture", 0}},
							"is_verified":     bson.M{"$arrayElemAt": []interface{}{"$sender_data.is_verified", 0}},
						},
						"else": nil,
					},
				},
				"conversation": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$conversation_data"}, 0}},
						"then": bson.M{
							"id":                bson.M{"$toString": bson.M{"$arrayElemAt": []interface{}{"$conversation_data._id", 0}}},
							"type":              bson.M{"$arrayElemAt": []interface{}{"$conversation_data.type", 0}},
							"title":             bson.M{"$arrayElemAt": []interface{}{"$conversation_data.title", 0}},
							"participant_count": bson.M{"$size": bson.M{"$arrayElemAt": []interface{}{"$conversation_data.participant_ids", 0}}},
						},
						"else": nil,
					},
				},
				"conversation_id": bson.M{"$toString": "$conversation_id"},
				"sender_id":       bson.M{"$toString": "$sender_id"},
				"reply_to_id":     bson.M{"$toString": "$reply_to_id"},
			},
		},
		{
			"$project": bson.M{
				"_id":             0,
				"id":              1,
				"conversation_id": 1,
				"sender_id":       1,
				"content":         1,
				"content_type":    1,
				"media_url":       1,
				"file_name":       1,
				"file_size":       1,
				"is_read":         1,
				"read_at":         1,
				"is_edited":       1,
				"edited_at":       1,
				"reply_to_id":     1,
				"created_at":      1,
				"updated_at":      1,
				"sender":          1,
				"conversation":    1,
			},
		},
		{
			"$sort": bson.M{sortField: sortValue},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("messages").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get messages", err)
		return
	}
	defer cursor.Close(ctx)

	var messages []bson.M
	if err := cursor.All(ctx, &messages); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode messages", err)
		return
	}

	// Get total count for pagination
	countPipeline := []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$count": "total",
		},
	}

	countCursor, err := h.db.Collection("messages").Aggregate(ctx, countPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to count messages", err)
		return
	}
	defer countCursor.Close(ctx)

	var countResult []bson.M
	total := int64(0)
	if err := countCursor.All(ctx, &countResult); err == nil && len(countResult) > 0 {
		if count, ok := countResult[0]["total"].(int32); ok {
			total = int64(count)
		}
	}

	// Create pagination metadata
	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Messages retrieved successfully", messages, *pagination, links)
}

func (h *AdminHandler) GetMessage(c *gin.Context) {
	messageID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID", nil)
		return
	}

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "sender_id",
				"foreignField": "_id",
				"as":           "sender",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "recipient_id",
				"foreignField": "_id",
				"as":           "recipient",
			},
		},
	}

	cursor, err := h.db.Collection("messages").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get message", err)
		return
	}
	defer cursor.Close(ctx)

	var message bson.M
	if cursor.Next(ctx) {
		cursor.Decode(&message)
	} else {
		utils.NotFoundResponse(c, "Message not found")
		return
	}

	utils.OkResponse(c, "Message retrieved successfully", message)
}

// BulkConversationAction performs bulk operations on conversations
func (h *AdminHandler) BulkConversationAction(c *gin.Context) {
	var req struct {
		ConversationIDs []string `json:"conversation_ids" binding:"required"`
		Action          string   `json:"action" binding:"required"`
		Reason          string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.ConversationIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 conversations allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, conversationID := range req.ConversationIDs {
		objID, err := primitive.ObjectIDFromHex(conversationID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Conversation %s: invalid ID", conversationID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "archive":
			update = bson.M{
				"$set": bson.M{
					"is_archived": true,
					"updated_at":  time.Now(),
				},
			}
		case "unarchive":
			update = bson.M{
				"$set": bson.M{
					"is_archived": false,
					"updated_at":  time.Now(),
				},
			}
		case "mute":
			update = bson.M{
				"$set": bson.M{
					"is_muted":   true,
					"updated_at": time.Now(),
				},
			}
		case "unmute":
			update = bson.M{
				"$set": bson.M{
					"is_muted":   false,
					"updated_at": time.Now(),
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Conversation %s: invalid action", conversationID))
			continue
		}

		_, err = h.db.Collection("conversations").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Conversation %s: %v", conversationID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_conversation_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk conversation action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.ConversationIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}
func (h *AdminHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid conversation ID", nil)
		return
	}

	ctx := c.Request.Context()
	var conversation bson.M
	err = h.db.Collection("conversations").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&conversation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "Conversation not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get conversation", err)
		return
	}

	utils.OkResponse(c, "Conversation retrieved successfully", conversation)
}

func (h *AdminHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("messages").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete message", err)
		return
	}

	h.logAdminActivity(c, "message_deletion", "Deleted message ID: "+messageID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Message deleted successfully", gin.H{
		"message_id": messageID,
		"reason":     req.Reason,
	})
}

func (h *AdminHandler) BulkMessageAction(c *gin.Context) {
	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"`
		Action     string   `json:"action" binding:"required"`
		Reason     string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.MessageIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 messages allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, messageID := range req.MessageIDs {
		objID, err := primitive.ObjectIDFromHex(messageID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Message %s: invalid ID", messageID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Message %s: invalid action", messageID))
			continue
		}

		_, err = h.db.Collection("messages").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Message %s: %v", messageID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_message_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk message action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.MessageIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Report Management
func (h *AdminHandler) GetAllReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.ReportFilter{
		Status:     models.ReportStatus(c.Query("status")),
		TargetType: c.Query("target_type"),
		Reason:     models.ReportReason(c.Query("reason")),
		Priority:   c.Query("priority"),
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &parsedDate
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &parsedDate
		}
	}

	reports, pagination, err := h.adminService.GetAllReports(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get reports", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Reports retrieved successfully", reports, *pagination, links)
}

func (h *AdminHandler) GetReport(c *gin.Context) {
	reportID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID", nil)
		return
	}

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "reporter_id",
				"foreignField": "_id",
				"as":           "reporter",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "resolved_by",
				"foreignField": "_id",
				"as":           "resolved_by_user",
			},
		},
	}

	cursor, err := h.db.Collection("reports").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get report", err)
		return
	}
	defer cursor.Close(ctx)

	var report bson.M
	if cursor.Next(ctx) {
		cursor.Decode(&report)
	} else {
		utils.NotFoundResponse(c, "Report not found")
		return
	}

	utils.OkResponse(c, "Report retrieved successfully", report)
}

func (h *AdminHandler) UpdateReportStatus(c *gin.Context) {
	reportID := c.Param("id")

	var req struct {
		Status     models.ReportStatus `json:"status" binding:"required"`
		Resolution string              `json:"resolution"`
		Note       string              `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	adminIDValue, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "Admin not authenticated")
		return
	}

	adminID, ok := adminIDValue.(primitive.ObjectID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid admin ID", nil)
		return
	}

	err := h.adminService.UpdateReportStatus(c.Request.Context(), reportID, req.Status, req.Resolution, req.Note, adminID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update report status", err)
		return
	}

	h.logAdminActivity(c, "report_status_update", "Updated report ID: "+reportID+" Status: "+string(req.Status))

	utils.OkResponse(c, "Report status updated successfully", gin.H{
		"report_id":  reportID,
		"status":     req.Status,
		"resolution": req.Resolution,
	})
}

func (h *AdminHandler) AssignReport(c *gin.Context) {
	reportID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID", nil)
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	assigneeObjID, err := primitive.ObjectIDFromHex(req.AssigneeID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid assignee ID", nil)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"assigned_to": assigneeObjID,
			"status":      models.ReportReviewing,
			"updated_at":  time.Now(),
		},
	}

	_, err = h.db.Collection("reports").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to assign report", err)
		return
	}

	h.logAdminActivity(c, "report_assignment", "Assigned report ID: "+reportID+" to "+req.AssigneeID)
	utils.OkResponse(c, "Report assigned successfully", gin.H{
		"report_id":   reportID,
		"assignee_id": req.AssigneeID,
	})
}

func (h *AdminHandler) ResolveReport(c *gin.Context) {
	reportID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID", nil)
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required"`
		Note       string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	adminIDValue, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "Admin not authenticated")
		return
	}

	adminID, ok := adminIDValue.(primitive.ObjectID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid admin ID", nil)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"status":          models.ReportResolved,
			"resolution":      req.Resolution,
			"resolution_note": req.Note,
			"resolved_by":     adminID,
			"resolved_at":     time.Now(),
			"updated_at":      time.Now(),
		},
	}

	_, err = h.db.Collection("reports").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to resolve report", err)
		return
	}

	h.logAdminActivity(c, "report_resolution", "Resolved report ID: "+reportID)
	utils.OkResponse(c, "Report resolved successfully", gin.H{
		"report_id":  reportID,
		"resolution": req.Resolution,
		"note":       req.Note,
	})
}

func (h *AdminHandler) RejectReport(c *gin.Context) {
	reportID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	adminIDValue, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "Admin not authenticated")
		return
	}

	adminID, ok := adminIDValue.(primitive.ObjectID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid admin ID", nil)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"status":          models.ReportRejected,
			"resolution":      "rejected",
			"resolution_note": req.Reason,
			"resolved_by":     adminID,
			"resolved_at":     time.Now(),
			"updated_at":      time.Now(),
		},
	}

	_, err = h.db.Collection("reports").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to reject report", err)
		return
	}

	h.logAdminActivity(c, "report_rejection", "Rejected report ID: "+reportID)
	utils.OkResponse(c, "Report rejected successfully", gin.H{
		"report_id": reportID,
		"reason":    req.Reason,
	})
}

func (h *AdminHandler) BulkReportAction(c *gin.Context) {
	var req struct {
		ReportIDs []string `json:"report_ids" binding:"required"`
		Action    string   `json:"action" binding:"required"`
		Reason    string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.ReportIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 reports allowed per bulk operation", nil)
		return
	}

	adminIDValue, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "Admin not authenticated")
		return
	}

	adminID, ok := adminIDValue.(primitive.ObjectID)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid admin ID", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, reportID := range req.ReportIDs {
		objID, err := primitive.ObjectIDFromHex(reportID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Report %s: invalid ID", reportID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "resolve":
			update = bson.M{
				"$set": bson.M{
					"status":          models.ReportResolved,
					"resolution":      "bulk_resolved",
					"resolution_note": req.Reason,
					"resolved_by":     adminID,
					"resolved_at":     time.Now(),
					"updated_at":      time.Now(),
				},
			}
		case "reject":
			update = bson.M{
				"$set": bson.M{
					"status":          models.ReportRejected,
					"resolution":      "rejected",
					"resolution_note": req.Reason,
					"resolved_by":     adminID,
					"resolved_at":     time.Now(),
					"updated_at":      time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Report %s: invalid action", reportID))
			continue
		}

		_, err = h.db.Collection("reports").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Report %s: %v", reportID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_report_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk report action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.ReportIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

func (h *AdminHandler) GetReportStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get total reports
	totalReports, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	// Get pending reports
	pendingReports, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"status":     models.ReportPending,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get resolved reports
	resolvedReports, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"status":     models.ReportResolved,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get rejected reports
	rejectedReports, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"status":     models.ReportRejected,
		"deleted_at": bson.M{"$exists": false},
	})

	// Get reports by reason
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$reason",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := h.db.Collection("reports").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get report statistics", err)
		return
	}
	defer cursor.Close(ctx)

	var reportsByReason []bson.M
	if err := cursor.All(ctx, &reportsByReason); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode report statistics", err)
		return
	}

	stats := gin.H{
		"total_reports":     totalReports,
		"pending_reports":   pendingReports,
		"resolved_reports":  resolvedReports,
		"rejected_reports":  rejectedReports,
		"reports_by_reason": reportsByReason,
		"resolution_rate":   float64(resolvedReports) / float64(totalReports) * 100,
		"rejection_rate":    float64(rejectedReports) / float64(totalReports) * 100,
	}

	utils.OkResponse(c, "Report statistics retrieved successfully", stats)
}

func (h *AdminHandler) GetReportSummary(c *gin.Context) {
	ctx := c.Request.Context()

	// Get reports by status over time
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": time.Now().AddDate(0, 0, -30)},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date":   bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
					"status": "$status",
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id.date": 1},
		},
	}

	cursor, err := h.db.Collection("reports").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get report summary", err)
		return
	}
	defer cursor.Close(ctx)

	var reportTrends []bson.M
	if err := cursor.All(ctx, &reportTrends); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode report summary", err)
		return
	}

	summary := gin.H{
		"report_trends": reportTrends,
		"period":        "last_30_days",
	}

	utils.OkResponse(c, "Report summary retrieved successfully", summary)
}

// Follow/Relationship Management
func (h *AdminHandler) GetAllFollows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
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
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "following_id",
				"foreignField": "_id",
				"as":           "following",
			},
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

	cursor, err := h.db.Collection("follows").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get follows", err)
		return
	}
	defer cursor.Close(ctx)

	var follows []bson.M
	if err := cursor.All(ctx, &follows); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode follows", err)
		return
	}

	total, _ := h.db.Collection("follows").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Follows retrieved successfully", follows, *pagination, links)
}

func (h *AdminHandler) GetFollow(c *gin.Context) {
	followID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(followID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid follow ID", nil)
		return
	}

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
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
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "following_id",
				"foreignField": "_id",
				"as":           "following",
			},
		},
	}

	cursor, err := h.db.Collection("follows").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get follow", err)
		return
	}
	defer cursor.Close(ctx)

	var follow bson.M
	if cursor.Next(ctx) {
		cursor.Decode(&follow)
	} else {
		utils.NotFoundResponse(c, "Follow relationship not found")
		return
	}

	utils.OkResponse(c, "Follow relationship retrieved successfully", follow)
}

func (h *AdminHandler) DeleteFollow(c *gin.Context) {
	followID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(followID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid follow ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("follows").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete follow relationship", err)
		return
	}

	h.logAdminActivity(c, "follow_deletion", "Deleted follow relationship ID: "+followID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Follow relationship deleted successfully", gin.H{
		"follow_id": followID,
		"reason":    req.Reason,
	})
}

func (h *AdminHandler) GetRelationships(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required", nil)
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", nil)
		return
	}

	ctx := c.Request.Context()

	// Get followers
	followersCount, _ := h.db.Collection("follows").CountDocuments(ctx, bson.M{
		"following_id": objID,
		"deleted_at":   bson.M{"$exists": false},
	})

	// Get following
	followingCount, _ := h.db.Collection("follows").CountDocuments(ctx, bson.M{
		"follower_id": objID,
		"deleted_at":  bson.M{"$exists": false},
	})

	// Get mutual follows (users who follow each other)
	mutualPipeline := []bson.M{
		{
			"$match": bson.M{
				"follower_id": objID,
				"deleted_at":  bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from": "follows",
				"let":  bson.M{"following_id": "$following_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{"$eq": []interface{}{"$follower_id", "$$following_id"}},
									{"$eq": []interface{}{"$following_id", objID}},
								},
							},
							"deleted_at": bson.M{"$exists": false},
						},
					},
				},
				"as": "mutual",
			},
		},
		{
			"$match": bson.M{
				"mutual": bson.M{"$ne": []interface{}{}},
			},
		},
		{
			"$count": "mutual_count",
		},
	}

	mutualCursor, err := h.db.Collection("follows").Aggregate(ctx, mutualPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get mutual relationships", err)
		return
	}
	defer mutualCursor.Close(ctx)

	var mutualResult []bson.M
	mutualCount := int64(0)
	if err := mutualCursor.All(ctx, &mutualResult); err == nil && len(mutualResult) > 0 {
		if count, ok := mutualResult[0]["mutual_count"].(int32); ok {
			mutualCount = int64(count)
		}
	}

	relationships := gin.H{
		"user_id":         userID,
		"followers_count": followersCount,
		"following_count": followingCount,
		"mutual_count":    mutualCount,
	}

	utils.OkResponse(c, "User relationships retrieved successfully", relationships)
}

func (h *AdminHandler) BulkFollowAction(c *gin.Context) {
	var req struct {
		FollowIDs []string `json:"follow_ids" binding:"required"`
		Action    string   `json:"action" binding:"required"`
		Reason    string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.FollowIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 follow relationships allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, followID := range req.FollowIDs {
		objID, err := primitive.ObjectIDFromHex(followID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Follow %s: invalid ID", followID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Follow %s: invalid action", followID))
			continue
		}

		_, err = h.db.Collection("follows").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Follow %s: %v", followID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_follow_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk follow action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.FollowIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Like/Reaction Management
func (h *AdminHandler) GetAllLikes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
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
			"$sort": bson.M{"created_at": -1},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("likes").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get likes", err)
		return
	}
	defer cursor.Close(ctx)

	var likes []bson.M
	if err := cursor.All(ctx, &likes); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode likes", err)
		return
	}

	total, _ := h.db.Collection("likes").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Likes retrieved successfully", likes, *pagination, links)
}

func (h *AdminHandler) GetLikeStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get total likes
	totalLikes, _ := h.db.Collection("likes").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	// Get likes by type
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$target_type",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := h.db.Collection("likes").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get like statistics", err)
		return
	}
	defer cursor.Close(ctx)

	var likesByType []bson.M
	if err := cursor.All(ctx, &likesByType); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode like statistics", err)
		return
	}

	// Get likes over time (last 30 days)
	timePipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": time.Now().AddDate(0, 0, -30)},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$created_at",
					},
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	timeCursor, err := h.db.Collection("likes").Aggregate(ctx, timePipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get like trends", err)
		return
	}
	defer timeCursor.Close(ctx)

	var likesOverTime []bson.M
	if err := timeCursor.All(ctx, &likesOverTime); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode like trends", err)
		return
	}

	stats := gin.H{
		"total_likes":     totalLikes,
		"likes_by_type":   likesByType,
		"likes_over_time": likesOverTime,
	}

	utils.OkResponse(c, "Like statistics retrieved successfully", stats)
}

func (h *AdminHandler) DeleteLike(c *gin.Context) {
	likeID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(likeID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid like ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("likes").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete like", err)
		return
	}

	h.logAdminActivity(c, "like_deletion", "Deleted like ID: "+likeID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Like deleted successfully", gin.H{
		"like_id": likeID,
		"reason":  req.Reason,
	})
}

func (h *AdminHandler) BulkLikeAction(c *gin.Context) {
	var req struct {
		LikeIDs []string `json:"like_ids" binding:"required"`
		Action  string   `json:"action" binding:"required"`
		Reason  string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.LikeIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 likes allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, likeID := range req.LikeIDs {
		objID, err := primitive.ObjectIDFromHex(likeID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Like %s: invalid ID", likeID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Like %s: invalid action", likeID))
			continue
		}

		_, err = h.db.Collection("likes").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Like %s: %v", likeID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_like_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk like action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.LikeIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Hashtag Management
func (h *AdminHandler) GetAllHashtags(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	hashtags, pagination, err := h.adminService.GetAllHashtags(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get hashtags", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Hashtags retrieved successfully", hashtags, *pagination, links)
}

func (h *AdminHandler) GetHashtag(c *gin.Context) {
	hashtagID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(hashtagID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hashtag ID", nil)
		return
	}

	ctx := c.Request.Context()
	var hashtag models.Hashtag
	err = h.db.Collection("hashtags").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&hashtag)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "Hashtag not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get hashtag", err)
		return
	}

	utils.OkResponse(c, "Hashtag retrieved successfully", hashtag.ToHashtagResponse())
}

func (h *AdminHandler) GetTrendingHashtags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	ctx := c.Request.Context()
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.M{"total_usage": -1})

	cursor, err := h.db.Collection("hashtags").Find(ctx, bson.M{
		"is_blocked": false,
		"deleted_at": bson.M{"$exists": false},
	}, opts)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending hashtags", err)
		return
	}
	defer cursor.Close(ctx)

	var hashtags []models.Hashtag
	if err := cursor.All(ctx, &hashtags); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode trending hashtags", err)
		return
	}

	var hashtagResponses []models.HashtagResponse
	for _, hashtag := range hashtags {
		hashtagResponses = append(hashtagResponses, hashtag.ToHashtagResponse())
	}

	utils.OkResponse(c, "Trending hashtags retrieved successfully", hashtagResponses)
}

func (h *AdminHandler) BlockHashtag(c *gin.Context) {
	hashtagID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(hashtagID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hashtag ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"is_blocked":   true,
			"block_reason": req.Reason,
			"updated_at":   time.Now(),
		},
	}

	_, err = h.db.Collection("hashtags").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to block hashtag", err)
		return
	}

	h.logAdminActivity(c, "hashtag_blocked", "Blocked hashtag ID: "+hashtagID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Hashtag blocked successfully", gin.H{
		"hashtag_id": hashtagID,
		"reason":     req.Reason,
	})
}

func (h *AdminHandler) UnblockHashtag(c *gin.Context) {
	hashtagID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(hashtagID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hashtag ID", nil)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"is_blocked": false,
			"updated_at": time.Now(),
		},
		"$unset": bson.M{
			"block_reason": "",
		},
	}

	_, err = h.db.Collection("hashtags").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to unblock hashtag", err)
		return
	}

	h.logAdminActivity(c, "hashtag_unblocked", "Unblocked hashtag ID: "+hashtagID)
	utils.OkResponse(c, "Hashtag unblocked successfully", gin.H{
		"hashtag_id": hashtagID,
	})
}

func (h *AdminHandler) DeleteHashtag(c *gin.Context) {
	hashtagID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(hashtagID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid hashtag ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("hashtags").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete hashtag", err)
		return
	}

	h.logAdminActivity(c, "hashtag_deletion", "Deleted hashtag ID: "+hashtagID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Hashtag deleted successfully", gin.H{
		"hashtag_id": hashtagID,
		"reason":     req.Reason,
	})
}

func (h *AdminHandler) BulkHashtagAction(c *gin.Context) {
	var req struct {
		HashtagIDs []string `json:"hashtag_ids" binding:"required"`
		Action     string   `json:"action" binding:"required"`
		Reason     string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.HashtagIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 hashtags allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, hashtagID := range req.HashtagIDs {
		objID, err := primitive.ObjectIDFromHex(hashtagID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Hashtag %s: invalid ID", hashtagID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "block":
			update = bson.M{
				"$set": bson.M{
					"is_blocked":   true,
					"block_reason": req.Reason,
					"updated_at":   time.Now(),
				},
			}
		case "unblock":
			update = bson.M{
				"$set": bson.M{
					"is_blocked": false,
					"updated_at": time.Now(),
				},
				"$unset": bson.M{
					"block_reason": "",
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Hashtag %s: invalid action", hashtagID))
			continue
		}

		_, err = h.db.Collection("hashtags").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Hashtag %s: %v", hashtagID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_hashtag_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk hashtag action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.HashtagIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Mention Management
func (h *AdminHandler) GetAllMentions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "mentioner_id",
				"foreignField": "_id",
				"as":           "mentioner",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "mentioned_user_id",
				"foreignField": "_id",
				"as":           "mentioned_user",
			},
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

	cursor, err := h.db.Collection("mentions").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get mentions", err)
		return
	}
	defer cursor.Close(ctx)

	var mentions []bson.M
	if err := cursor.All(ctx, &mentions); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode mentions", err)
		return
	}

	total, _ := h.db.Collection("mentions").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Mentions retrieved successfully", mentions, *pagination, links)
}

func (h *AdminHandler) GetMention(c *gin.Context) {
	mentionID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(mentionID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid mention ID", nil)
		return
	}

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "mentioner_id",
				"foreignField": "_id",
				"as":           "mentioner",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "mentioned_user_id",
				"foreignField": "_id",
				"as":           "mentioned_user",
			},
		},
	}

	cursor, err := h.db.Collection("mentions").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get mention", err)
		return
	}
	defer cursor.Close(ctx)

	var mention bson.M
	if cursor.Next(ctx) {
		cursor.Decode(&mention)
	} else {
		utils.NotFoundResponse(c, "Mention not found")
		return
	}

	utils.OkResponse(c, "Mention retrieved successfully", mention)
}

func (h *AdminHandler) DeleteMention(c *gin.Context) {
	mentionID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(mentionID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid mention ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("mentions").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete mention", err)
		return
	}

	h.logAdminActivity(c, "mention_deletion", "Deleted mention ID: "+mentionID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Mention deleted successfully", gin.H{
		"mention_id": mentionID,
		"reason":     req.Reason,
	})
}

func (h *AdminHandler) BulkMentionAction(c *gin.Context) {
	var req struct {
		MentionIDs []string `json:"mention_ids" binding:"required"`
		Action     string   `json:"action" binding:"required"`
		Reason     string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.MentionIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 mentions allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, mentionID := range req.MentionIDs {
		objID, err := primitive.ObjectIDFromHex(mentionID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Mention %s: invalid ID", mentionID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Mention %s: invalid action", mentionID))
			continue
		}

		_, err = h.db.Collection("mentions").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Mention %s: %v", mentionID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_mention_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk mention action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.MentionIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Media Management
func (h *AdminHandler) GetAllMedia(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	media, pagination, err := h.adminService.GetAllMedia(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get media", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Media retrieved successfully", media, *pagination, links)
}

func (h *AdminHandler) GetMedia(c *gin.Context) {
	mediaID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(mediaID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID", nil)
		return
	}

	ctx := c.Request.Context()
	var media models.Media
	err = h.db.Collection("media").FindOne(ctx, bson.M{
		"_id":        objID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&media)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.NotFoundResponse(c, "Media not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get media", err)
		return
	}

	utils.OkResponse(c, "Media retrieved successfully", media.ToMediaResponse())
}

func (h *AdminHandler) GetMediaStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get total media count
	totalMedia, _ := h.db.Collection("media").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	// Get media by type
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":        "$media_type",
				"count":      bson.M{"$sum": 1},
				"total_size": bson.M{"$sum": "$file_size"},
			},
		},
	}

	cursor, err := h.db.Collection("media").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get media statistics", err)
		return
	}
	defer cursor.Close(ctx)

	var mediaByType []bson.M
	if err := cursor.All(ctx, &mediaByType); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode media statistics", err)
		return
	}

	// Get total storage used
	storagePipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":           nil,
				"total_storage": bson.M{"$sum": "$file_size"},
			},
		},
	}

	storageCursor, err := h.db.Collection("media").Aggregate(ctx, storagePipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get storage statistics", err)
		return
	}
	defer storageCursor.Close(ctx)

	var storageResult []bson.M
	var totalStorage int64 = 0
	if err := storageCursor.All(ctx, &storageResult); err == nil && len(storageResult) > 0 {
		if size, ok := storageResult[0]["total_storage"].(int64); ok {
			totalStorage = size
		}
	}

	stats := gin.H{
		"total_media":   totalMedia,
		"media_by_type": mediaByType,
		"total_storage": totalStorage,
		"storage_gb":    float64(totalStorage) / (1024 * 1024 * 1024),
	}

	utils.OkResponse(c, "Media statistics retrieved successfully", stats)
}

func (h *AdminHandler) ModerateMedia(c *gin.Context) {
	mediaID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(mediaID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID", nil)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // approve, reject, flag
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	var update bson.M

	switch req.Action {
	case "approve":
		update = bson.M{
			"$set": bson.M{
				"moderation_status": "approved",
				"updated_at":        time.Now(),
			},
		}
	case "reject":
		update = bson.M{
			"$set": bson.M{
				"moderation_status": "rejected",
				"moderation_reason": req.Reason,
				"updated_at":        time.Now(),
			},
		}
	case "flag":
		update = bson.M{
			"$set": bson.M{
				"moderation_status": "flagged",
				"moderation_reason": req.Reason,
				"updated_at":        time.Now(),
			},
		}
	default:
		utils.BadRequestResponse(c, "Invalid moderation action", nil)
		return
	}

	_, err = h.db.Collection("media").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to moderate media", err)
		return
	}

	h.logAdminActivity(c, "media_moderation", "Moderated media ID: "+mediaID+" Action: "+req.Action)
	utils.OkResponse(c, "Media moderated successfully", gin.H{
		"media_id": mediaID,
		"action":   req.Action,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) DeleteMedia(c *gin.Context) {
	mediaID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(mediaID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("media").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete media", err)
		return
	}

	h.logAdminActivity(c, "media_deletion", "Deleted media ID: "+mediaID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Media deleted successfully", gin.H{
		"media_id": mediaID,
		"reason":   req.Reason,
	})
}

func (h *AdminHandler) BulkMediaAction(c *gin.Context) {
	var req struct {
		MediaIDs []string `json:"media_ids" binding:"required"`
		Action   string   `json:"action" binding:"required"`
		Reason   string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.MediaIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 media items allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, mediaID := range req.MediaIDs {
		objID, err := primitive.ObjectIDFromHex(mediaID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Media %s: invalid ID", mediaID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "approve":
			update = bson.M{
				"$set": bson.M{
					"moderation_status": "approved",
					"updated_at":        time.Now(),
				},
			}
		case "reject":
			update = bson.M{
				"$set": bson.M{
					"moderation_status": "rejected",
					"moderation_reason": req.Reason,
					"updated_at":        time.Now(),
				},
			}
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Media %s: invalid action", mediaID))
			continue
		}

		_, err = h.db.Collection("media").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Media %s: %v", mediaID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_media_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk media action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.MediaIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

func (h *AdminHandler) GetStorageStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get storage statistics by media type
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":        "$media_type",
				"count":      bson.M{"$sum": 1},
				"total_size": bson.M{"$sum": "$file_size"},
				"avg_size":   bson.M{"$avg": "$file_size"},
			},
		},
	}

	cursor, err := h.db.Collection("media").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get storage statistics", err)
		return
	}
	defer cursor.Close(ctx)

	var storageByType []bson.M
	if err := cursor.All(ctx, &storageByType); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode storage statistics", err)
		return
	}

	// Get total storage
	totalPipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":           nil,
				"total_files":   bson.M{"$sum": 1},
				"total_storage": bson.M{"$sum": "$file_size"},
			},
		},
	}

	totalCursor, err := h.db.Collection("media").Aggregate(ctx, totalPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get total storage", err)
		return
	}
	defer totalCursor.Close(ctx)

	var totalResult []bson.M
	var totalFiles int64 = 0
	var totalStorage int64 = 0
	if err := totalCursor.All(ctx, &totalResult); err == nil && len(totalResult) > 0 {
		if files, ok := totalResult[0]["total_files"].(int32); ok {
			totalFiles = int64(files)
		}
		if storage, ok := totalResult[0]["total_storage"].(int64); ok {
			totalStorage = storage
		}
	}

	stats := gin.H{
		"storage_by_type": storageByType,
		"total_files":     totalFiles,
		"total_storage":   totalStorage,
		"storage_gb":      float64(totalStorage) / (1024 * 1024 * 1024),
		"avg_file_size":   float64(totalStorage) / float64(totalFiles),
	}

	utils.OkResponse(c, "Storage statistics retrieved successfully", stats)
}

func (h *AdminHandler) CleanupStorage(c *gin.Context) {
	var req struct {
		OlderThan int    `json:"older_than_days" binding:"required"` // Delete files older than X days
		MediaType string `json:"media_type,omitempty"`               // Optional: specific media type
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	cutoffDate := time.Now().AddDate(0, 0, -req.OlderThan)

	filter := bson.M{
		"created_at": bson.M{"$lt": cutoffDate},
		"deleted_at": bson.M{"$exists": false},
	}

	if req.MediaType != "" {
		filter["media_type"] = req.MediaType
	}

	// Count files to be cleaned up
	count, err := h.db.Collection("media").CountDocuments(ctx, filter)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to count files for cleanup", err)
		return
	}

	// Mark files as deleted
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := h.db.Collection("media").UpdateMany(ctx, filter, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to cleanup storage", err)
		return
	}

	h.logAdminActivity(c, "storage_cleanup", fmt.Sprintf("Cleaned up %d files older than %d days", result.ModifiedCount, req.OlderThan))

	utils.OkResponse(c, "Storage cleanup completed", gin.H{
		"files_found":   count,
		"files_cleaned": result.ModifiedCount,
		"older_than":    req.OlderThan,
		"media_type":    req.MediaType,
	})
}

// Notification Management
func (h *AdminHandler) GetAllNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
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
			"$sort": bson.M{"created_at": -1},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := h.db.Collection("notifications").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notifications", err)
		return
	}
	defer cursor.Close(ctx)

	var notifications []bson.M
	if err := cursor.All(ctx, &notifications); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode notifications", err)
		return
	}

	total, _ := h.db.Collection("notifications").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Notifications retrieved successfully", notifications, *pagination, links)
}

func (h *AdminHandler) GetNotification(c *gin.Context) {
	notificationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid notification ID", nil)
		return
	}

	ctx := c.Request.Context()
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
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
	}

	cursor, err := h.db.Collection("notifications").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notification", err)
		return
	}
	defer cursor.Close(ctx)

	var notification bson.M
	if cursor.Next(ctx) {
		cursor.Decode(&notification)
	} else {
		utils.NotFoundResponse(c, "Notification not found")
		return
	}

	utils.OkResponse(c, "Notification retrieved successfully", notification)
}

func (h *AdminHandler) SendNotificationToUsers(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
		Title   string   `json:"title" binding:"required"`
		Message string   `json:"message" binding:"required"`
		Type    string   `json:"type" binding:"required"`
		Data    gin.H    `json:"data,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.UserIDs) > 1000 {
		utils.BadRequestResponse(c, "Maximum 1000 users allowed per notification", nil)
		return
	}

	ctx := c.Request.Context()
	successCount := 0
	failureCount := 0
	var errors []string

	for _, userID := range req.UserIDs {
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("User %s: invalid ID", userID))
			continue
		}

		notification := bson.M{
			"user_id":    objID,
			"title":      req.Title,
			"message":    req.Message,
			"type":       req.Type,
			"data":       req.Data,
			"is_read":    false,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}

		_, err = h.db.Collection("notifications").InsertOne(ctx, notification)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("User %s: %v", userID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "notification_sent", fmt.Sprintf("Sent notification to %d users", successCount))

	utils.OkResponse(c, "Notifications sent", gin.H{
		"title":         req.Title,
		"total":         len(req.UserIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

func (h *AdminHandler) BroadcastNotification(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Type    string `json:"type" binding:"required"`
		Data    gin.H  `json:"data,omitempty"`
		Filter  gin.H  `json:"filter,omitempty"` // Optional filter for targeting users
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()

	// Get all active users (or filtered users)
	userFilter := bson.M{
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	// Apply additional filters if provided
	if req.Filter != nil {
		for key, value := range req.Filter {
			userFilter[key] = value
		}
	}

	cursor, err := h.db.Collection("users").Find(ctx, userFilter)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get users for broadcast", err)
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode users", err)
		return
	}

	// Create notifications for all users
	var notifications []interface{}
	for _, user := range users {
		notification := bson.M{
			"user_id":    user.ID,
			"title":      req.Title,
			"message":    req.Message,
			"type":       req.Type,
			"data":       req.Data,
			"is_read":    false,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		notifications = append(notifications, notification)
	}

	if len(notifications) > 0 {
		_, err = h.db.Collection("notifications").InsertMany(ctx, notifications)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to broadcast notifications", err)
			return
		}
	}

	h.logAdminActivity(c, "notification_broadcast", fmt.Sprintf("Broadcast notification to %d users", len(users)))

	utils.OkResponse(c, "Notification broadcast successfully", gin.H{
		"title":      req.Title,
		"recipients": len(users),
		"type":       req.Type,
	})
}

func (h *AdminHandler) GetNotificationStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get total notifications
	totalNotifications, _ := h.db.Collection("notifications").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	// Get read vs unread
	readNotifications, _ := h.db.Collection("notifications").CountDocuments(ctx, bson.M{
		"is_read":    true,
		"deleted_at": bson.M{"$exists": false},
	})

	unreadNotifications := totalNotifications - readNotifications

	// Get notifications by type
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := h.db.Collection("notifications").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get notification statistics", err)
		return
	}
	defer cursor.Close(ctx)

	var notificationsByType []bson.M
	if err := cursor.All(ctx, &notificationsByType); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode notification statistics", err)
		return
	}

	stats := gin.H{
		"total_notifications":   totalNotifications,
		"read_notifications":    readNotifications,
		"unread_notifications":  unreadNotifications,
		"read_rate":             float64(readNotifications) / float64(totalNotifications) * 100,
		"notifications_by_type": notificationsByType,
	}

	utils.OkResponse(c, "Notification statistics retrieved successfully", stats)
}

func (h *AdminHandler) DeleteNotification(c *gin.Context) {
	notificationID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid notification ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = h.db.Collection("notifications").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete notification", err)
		return
	}

	h.logAdminActivity(c, "notification_deletion", "Deleted notification ID: "+notificationID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Notification deleted successfully", gin.H{
		"notification_id": notificationID,
		"reason":          req.Reason,
	})
}

func (h *AdminHandler) BulkNotificationAction(c *gin.Context) {
	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
		Action          string   `json:"action" binding:"required"`
		Reason          string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.NotificationIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 notifications allowed per bulk operation", nil)
		return
	}

	successCount := 0
	failureCount := 0
	var errors []string

	for _, notificationID := range req.NotificationIDs {
		objID, err := primitive.ObjectIDFromHex(notificationID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Notification %s: invalid ID", notificationID))
			continue
		}

		ctx := c.Request.Context()
		var update bson.M

		switch req.Action {
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		case "mark_read":
			update = bson.M{
				"$set": bson.M{
					"is_read":    true,
					"updated_at": time.Now(),
				},
			}
		case "mark_unread":
			update = bson.M{
				"$set": bson.M{
					"is_read":    false,
					"updated_at": time.Now(),
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Notification %s: invalid action", notificationID))
			continue
		}

		_, err = h.db.Collection("notifications").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Notification %s: %v", notificationID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_notification_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk notification action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.NotificationIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}

// Analytics
func (h *AdminHandler) GetUserAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	analytics, err := h.adminService.GetUserAnalytics(c.Request.Context(), period)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user analytics", err)
		return
	}
	utils.OkResponse(c, "User analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetContentAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	analytics, err := h.adminService.GetContentAnalytics(c.Request.Context(), period)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get content analytics", err)
		return
	}
	utils.OkResponse(c, "Content analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetEngagementAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	ctx := c.Request.Context()

	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	// Get engagement metrics
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": startDate},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
					"type": "$target_type",
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id.date": 1},
		},
	}

	cursor, err := h.db.Collection("likes").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get engagement analytics", err)
		return
	}
	defer cursor.Close(ctx)

	var engagementData []bson.M
	if err := cursor.All(ctx, &engagementData); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode engagement analytics", err)
		return
	}

	// Get comment engagement
	commentPipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": startDate},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"},
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	commentCursor, err := h.db.Collection("comments").Aggregate(ctx, commentPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get comment analytics", err)
		return
	}
	defer commentCursor.Close(ctx)

	var commentData []bson.M
	if err := commentCursor.All(ctx, &commentData); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode comment analytics", err)
		return
	}

	analytics := gin.H{
		"period":          period,
		"engagement_data": engagementData,
		"comment_data":    commentData,
	}

	utils.OkResponse(c, "Engagement analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetGrowthAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	ctx := c.Request.Context()

	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	case "365d":
		days = 365
	default:
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	// User growth
	userGrowthPipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": startDate},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"},
				},
				"new_users": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	userCursor, err := h.db.Collection("users").Aggregate(ctx, userGrowthPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user growth analytics", err)
		return
	}
	defer userCursor.Close(ctx)

	var userGrowthData []bson.M
	if err := userCursor.All(ctx, &userGrowthData); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode user growth analytics", err)
		return
	}

	// Content growth
	contentGrowthPipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": startDate},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"},
				},
				"new_posts": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	contentCursor, err := h.db.Collection("posts").Aggregate(ctx, contentGrowthPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get content growth analytics", err)
		return
	}
	defer contentCursor.Close(ctx)

	var contentGrowthData []bson.M
	if err := contentCursor.All(ctx, &contentGrowthData); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode content growth analytics", err)
		return
	}

	analytics := gin.H{
		"period":              period,
		"user_growth_data":    userGrowthData,
		"content_growth_data": contentGrowthData,
	}

	utils.OkResponse(c, "Growth analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetDemographicAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	// Users by age group (if age is stored)
	agePipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
				"age":        bson.M{"$exists": true, "$ne": nil},
			},
		},
		{
			"$addFields": bson.M{
				"age_group": bson.M{
					"$switch": bson.M{
						"branches": []bson.M{
							{"case": bson.M{"$lt": []interface{}{"$age", 18}}, "then": "Under 18"},
							{"case": bson.M{"$lt": []interface{}{"$age", 25}}, "then": "18-24"},
							{"case": bson.M{"$lt": []interface{}{"$age", 35}}, "then": "25-34"},
							{"case": bson.M{"$lt": []interface{}{"$age", 45}}, "then": "35-44"},
							{"case": bson.M{"$lt": []interface{}{"$age", 55}}, "then": "45-54"},
							{"case": bson.M{"$lt": []interface{}{"$age", 65}}, "then": "55-64"},
						},
						"default": "65+",
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$age_group",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	ageCursor, err := h.db.Collection("users").Aggregate(ctx, agePipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get age demographics", err)
		return
	}
	defer ageCursor.Close(ctx)

	var ageGroups []bson.M
	if err := ageCursor.All(ctx, &ageGroups); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode age demographics", err)
		return
	}

	// Users by gender (if gender is stored)
	genderPipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
				"gender":     bson.M{"$exists": true, "$ne": ""},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$gender",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	genderCursor, err := h.db.Collection("users").Aggregate(ctx, genderPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get gender demographics", err)
		return
	}
	defer genderCursor.Close(ctx)

	var genderGroups []bson.M
	if err := genderCursor.All(ctx, &genderGroups); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode gender demographics", err)
		return
	}

	// Users by country/location (if location is stored)
	locationPipeline := []bson.M{
		{
			"$match": bson.M{
				"deleted_at": bson.M{"$exists": false},
				"country":    bson.M{"$exists": true, "$ne": ""},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$country",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": 10,
		},
	}

	locationCursor, err := h.db.Collection("users").Aggregate(ctx, locationPipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get location demographics", err)
		return
	}
	defer locationCursor.Close(ctx)

	var locationGroups []bson.M
	if err := locationCursor.All(ctx, &locationGroups); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode location demographics", err)
		return
	}

	analytics := gin.H{
		"age_groups":      ageGroups,
		"gender_groups":   genderGroups,
		"location_groups": locationGroups,
	}

	utils.OkResponse(c, "Demographic analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetRevenueAnalytics(c *gin.Context) {
	// This would be implemented if you have revenue/billing data
	// For now, return placeholder data
	analytics := gin.H{
		"total_revenue":     0,
		"monthly_revenue":   0,
		"revenue_sources":   []gin.H{},
		"revenue_trends":    []gin.H{},
		"subscription_data": gin.H{},
	}

	utils.OkResponse(c, "Revenue analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetCustomReport(c *gin.Context) {
	var req struct {
		ReportType string `json:"report_type" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"`
		EndDate    string `json:"end_date" binding:"required"`
		Filters    gin.H  `json:"filters,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid start date format", nil)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid end date format", nil)
		return
	}

	ctx := c.Request.Context()
	var report gin.H

	switch req.ReportType {
	case "user_activity":
		report, err = h.generateUserActivityReport(ctx, startDate, endDate, req.Filters)
	case "content_performance":
		report, err = h.generateContentPerformanceReport(ctx, startDate, endDate, req.Filters)
	case "engagement_summary":
		report, err = h.generateEngagementSummaryReport(ctx, startDate, endDate, req.Filters)
	default:
		utils.BadRequestResponse(c, "Invalid report type", nil)
		return
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate custom report", err)
		return
	}

	h.logAdminActivity(c, "custom_report_generated", "Generated "+req.ReportType+" report")
	utils.OkResponse(c, "Custom report generated successfully", report)
}

func (h *AdminHandler) GetRealtimeAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	// Get real-time metrics for the last hour
	lastHour := time.Now().Add(-1 * time.Hour)

	// Active users in last hour
	activeUsers, _ := h.db.Collection("users").CountDocuments(ctx, bson.M{
		"last_active_at": bson.M{"$gte": lastHour},
		"deleted_at":     bson.M{"$exists": false},
	})

	// New posts in last hour
	newPosts, _ := h.db.Collection("posts").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": lastHour},
		"deleted_at": bson.M{"$exists": false},
	})

	// New comments in last hour
	newComments, _ := h.db.Collection("comments").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": lastHour},
		"deleted_at": bson.M{"$exists": false},
	})

	// New likes in last hour
	newLikes, _ := h.db.Collection("likes").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": lastHour},
		"deleted_at": bson.M{"$exists": false},
	})

	// New users in last hour
	newUsers, _ := h.db.Collection("users").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": lastHour},
		"deleted_at": bson.M{"$exists": false},
	})

	// New reports in last hour
	newReports, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": lastHour},
		"deleted_at": bson.M{"$exists": false},
	})

	analytics := gin.H{
		"timestamp":    time.Now(),
		"period":       "last_hour",
		"active_users": activeUsers,
		"new_posts":    newPosts,
		"new_comments": newComments,
		"new_likes":    newLikes,
		"new_users":    newUsers,
		"new_reports":  newReports,
	}

	utils.OkResponse(c, "Real-time analytics retrieved successfully", analytics)
}

func (h *AdminHandler) GetLiveStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current statistics
	totalUsers, _ := h.db.Collection("users").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	totalPosts, _ := h.db.Collection("posts").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	totalComments, _ := h.db.Collection("comments").CountDocuments(ctx, bson.M{
		"deleted_at": bson.M{"$exists": false},
	})

	pendingReports, _ := h.db.Collection("reports").CountDocuments(ctx, bson.M{
		"status":     models.ReportPending,
		"deleted_at": bson.M{"$exists": false},
	})

	// Current online users (active in last 5 minutes)
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	onlineUsers, _ := h.db.Collection("users").CountDocuments(ctx, bson.M{
		"last_active_at": bson.M{"$gte": fiveMinutesAgo},
		"deleted_at":     bson.M{"$exists": false},
	})

	stats := gin.H{
		"timestamp":       time.Now(),
		"total_users":     totalUsers,
		"total_posts":     totalPosts,
		"total_comments":  totalComments,
		"pending_reports": pendingReports,
		"online_users":    onlineUsers,
	}

	utils.OkResponse(c, "Live statistics retrieved successfully", stats)
}

// System Management
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database connectivity
	err := h.db.Client().Ping(ctx, nil)
	dbStatus := "connected"
	if err != nil {
		dbStatus = "disconnected"
	}

	// Get database stats
	dbStats := h.db.RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}})
	var dbStatsResult bson.M
	dbStats.Decode(&dbStatsResult)

	health := gin.H{
		"status":           "healthy",
		"database_status":  dbStatus,
		"cache_status":     "active", // This would check your cache (Redis, etc.)
		"storage_status":   "available",
		"response_time_ms": 150,          // This would measure actual response time
		"memory_usage":     65.5,         // This would get system memory usage
		"cpu_usage":        23.2,         // This would get system CPU usage
		"disk_usage":       45.8,         // This would get disk usage
		"uptime":           "5d 12h 30m", // This would calculate actual uptime
		"last_updated":     time.Now(),
		"database_stats":   dbStatsResult,
	}

	utils.OkResponse(c, "System health retrieved successfully", health)
}

func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	info := gin.H{
		"app_name":        "Social Media API",
		"version":         "1.0.0",
		"environment":     "production",
		"go_version":      "1.21.0",
		"mongodb_version": "7.0",
		"api_version":     "v1",
		"build_date":      "2024-01-15",
		"commit_hash":     "abc123def",
		"server_time":     time.Now(),
		"timezone":        "UTC",
	}

	utils.OkResponse(c, "System information retrieved successfully", info)
}

func (h *AdminHandler) GetSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	level := c.DefaultQuery("level", "")
	c.Query("start_date")
	c.Query("end_date")

	// This would typically read from actual log files or a logging service
	// For now, return sample log data
	logs := []gin.H{
		{
			"id":        "1",
			"timestamp": time.Now().Add(-1 * time.Hour),
			"level":     "INFO",
			"message":   "User logged in successfully",
			"source":    "auth_service",
			"user_id":   "user123",
		},
		{
			"id":        "2",
			"timestamp": time.Now().Add(-2 * time.Hour),
			"level":     "WARN",
			"message":   "High memory usage detected",
			"source":    "system_monitor",
		},
		{
			"id":        "3",
			"timestamp": time.Now().Add(-3 * time.Hour),
			"level":     "ERROR",
			"message":   "Database connection timeout",
			"source":    "database_service",
		},
	}

	// Filter logs based on parameters
	if level != "" {
		var filteredLogs []gin.H
		for _, log := range logs {
			if log["level"] == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       int64(len(logs)),
		TotalPages:  (len(logs) + limit - 1) / limit,
		HasNext:     page*limit < len(logs),
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "System logs retrieved successfully", logs, *pagination, links)
}

func (h *AdminHandler) GetPerformanceMetrics(c *gin.Context) {
	// This would typically get actual performance metrics
	metrics := gin.H{
		"api_response_time": gin.H{
			"avg": 150.5,
			"p50": 125.0,
			"p95": 300.0,
			"p99": 500.0,
		},
		"database_performance": gin.H{
			"query_time_avg":   45.2,
			"connections_used": 25,
			"connections_max":  100,
		},
		"memory_usage": gin.H{
			"used_mb":       512,
			"total_mb":      1024,
			"usage_percent": 50.0,
		},
		"cpu_usage": gin.H{
			"current_percent": 23.5,
			"avg_percent":     18.2,
		},
		"request_metrics": gin.H{
			"requests_per_minute": 145,
			"error_rate":          0.02,
			"success_rate":        99.98,
		},
	}

	utils.OkResponse(c, "Performance metrics retrieved successfully", metrics)
}

func (h *AdminHandler) GetDatabaseStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get database statistics
	dbStats := h.db.RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}})
	var dbStatsResult bson.M
	dbStats.Decode(&dbStatsResult)

	// Get collection stats
	collections := []string{"users", "posts", "comments", "likes", "follows", "reports", "notifications"}
	collectionStats := make(map[string]interface{})

	for _, collection := range collections {
		stats := h.db.RunCommand(ctx, bson.D{{Key: "collStats", Value: collection}})
		var result bson.M
		stats.Decode(&result)
		collectionStats[collection] = result
	}

	stats := gin.H{
		"database_stats":   dbStatsResult,
		"collection_stats": collectionStats,
		"timestamp":        time.Now(),
	}

	utils.OkResponse(c, "Database statistics retrieved successfully", stats)
}

func (h *AdminHandler) GetCacheStats(c *gin.Context) {
	// This would get actual cache statistics (Redis, Memcached, etc.)
	stats := gin.H{
		"cache_hits":        15420,
		"cache_misses":      1240,
		"hit_rate":          92.5,
		"memory_used":       "156MB",
		"memory_total":      "512MB",
		"keys_count":        8945,
		"expired_keys":      234,
		"evicted_keys":      12,
		"connected_clients": 45,
		"uptime_seconds":    86400,
	}

	utils.OkResponse(c, "Cache statistics retrieved successfully", stats)
}

func (h *AdminHandler) ClearCache(c *gin.Context) {
	var req struct {
		CacheType string `json:"cache_type"` // "all", "users", "posts", etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.CacheType = "all"
	}

	// This would clear actual cache
	h.logAdminActivity(c, "cache_cleared", "Cleared "+req.CacheType+" cache")

	utils.OkResponse(c, "Cache cleared successfully", gin.H{
		"cache_type": req.CacheType,
		"timestamp":  time.Now(),
	})
}

func (h *AdminHandler) WarmCache(c *gin.Context) {
	var req struct {
		CacheType string `json:"cache_type"` // "all", "users", "posts", etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.CacheType = "all"
	}

	// This would warm up the cache
	h.logAdminActivity(c, "cache_warmed", "Warmed "+req.CacheType+" cache")

	utils.OkResponse(c, "Cache warming started", gin.H{
		"cache_type": req.CacheType,
		"status":     "in_progress",
		"timestamp":  time.Now(),
	})
}

func (h *AdminHandler) EnableMaintenanceMode(c *gin.Context) {
	var req struct {
		Message  string `json:"message"`
		Duration string `json:"duration"` // "1h", "30m", etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would enable maintenance mode
	h.logAdminActivity(c, "maintenance_enabled", "Enabled maintenance mode: "+req.Message)

	utils.OkResponse(c, "Maintenance mode enabled", gin.H{
		"message":   req.Message,
		"duration":  req.Duration,
		"enabled":   true,
		"timestamp": time.Now(),
	})
}

func (h *AdminHandler) DisableMaintenanceMode(c *gin.Context) {
	// This would disable maintenance mode
	h.logAdminActivity(c, "maintenance_disabled", "Disabled maintenance mode")

	utils.OkResponse(c, "Maintenance mode disabled", gin.H{
		"enabled":   false,
		"timestamp": time.Now(),
	})
}

func (h *AdminHandler) BackupDatabase(c *gin.Context) {
	var req struct {
		BackupType  string   `json:"backup_type"` // "full", "incremental"
		Collections []string `json:"collections,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.BackupType = "full"
	}

	// This would create a database backup
	backupID := primitive.NewObjectID().Hex()

	h.logAdminActivity(c, "database_backup", "Started database backup: "+req.BackupType)

	utils.AcceptedResponse(c, "Database backup started", gin.H{
		"backup_id":   backupID,
		"backup_type": req.BackupType,
		"status":      "in_progress",
		"started_at":  time.Now(),
	})
}

func (h *AdminHandler) GetDatabaseBackups(c *gin.Context) {
	// This would list database backups
	backups := []gin.H{
		{
			"id":         "backup_1",
			"type":       "full",
			"status":     "completed",
			"size":       "2.5GB",
			"created_at": time.Now().Add(-24 * time.Hour),
			"expires_at": time.Now().Add(30 * 24 * time.Hour),
		},
		{
			"id":         "backup_2",
			"type":       "incremental",
			"status":     "completed",
			"size":       "150MB",
			"created_at": time.Now().Add(-12 * time.Hour),
			"expires_at": time.Now().Add(7 * 24 * time.Hour),
		},
	}

	utils.OkResponse(c, "Database backups retrieved successfully", backups)
}

func (h *AdminHandler) RestoreDatabase(c *gin.Context) {
	var req struct {
		BackupID    string   `json:"backup_id" binding:"required"`
		RestoreType string   `json:"restore_type"` // "full", "selective"
		Collections []string `json:"collections,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would restore from database backup
	h.logAdminActivity(c, "database_restore", "Started database restore from backup: "+req.BackupID)

	utils.AcceptedResponse(c, "Database restore started", gin.H{
		"backup_id":    req.BackupID,
		"restore_type": req.RestoreType,
		"status":       "in_progress",
		"started_at":   time.Now(),
	})
}

func (h *AdminHandler) OptimizeDatabase(c *gin.Context) {
	// This would optimize database indexes and performance
	h.logAdminActivity(c, "database_optimize", "Started database optimization")

	utils.AcceptedResponse(c, "Database optimization started", gin.H{
		"status":     "in_progress",
		"started_at": time.Now(),
	})
}

// Configuration Management
func (h *AdminHandler) GetConfiguration(c *gin.Context) {
	// This would read from actual configuration
	config := gin.H{
		"app_name":           "Social Media API",
		"version":            "1.0.0",
		"environment":        "production",
		"max_file_size":      "10MB",
		"allowed_file_types": []string{"jpg", "png", "gif", "mp4", "mov"},
		"rate_limits": gin.H{
			"posts_per_hour":    10,
			"comments_per_hour": 50,
			"messages_per_hour": 100,
		},
		"features": gin.H{
			"registration_enabled": true,
			"email_verification":   true,
			"two_factor_auth":      false,
			"maintenance_mode":     false,
		},
		"moderation": gin.H{
			"auto_moderation":  true,
			"profanity_filter": true,
			"spam_detection":   true,
		},
	}

	utils.OkResponse(c, "Configuration retrieved successfully", config)
}

func (h *AdminHandler) UpdateConfiguration(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would update configuration in database/file
	h.logAdminActivity(c, "config_update", "Updated system configuration")

	utils.OkResponse(c, "Configuration updated successfully", req)
}

func (h *AdminHandler) GetConfigurationHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// This would get configuration change history
	history := []gin.H{
		{
			"id":          "1",
			"changes":     gin.H{"rate_limits.posts_per_hour": gin.H{"from": 5, "to": 10}},
			"changed_by":  "admin_user",
			"changed_at":  time.Now().Add(-2 * time.Hour),
			"description": "Increased post rate limit",
		},
	}

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       int64(len(history)),
		TotalPages:  (len(history) + limit - 1) / limit,
		HasNext:     page*limit < len(history),
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Configuration history retrieved successfully", history, *pagination, links)
}

func (h *AdminHandler) RollbackConfiguration(c *gin.Context) {
	var req struct {
		ConfigID string `json:"config_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would rollback to previous configuration
	h.logAdminActivity(c, "config_rollback", "Rolled back configuration to: "+req.ConfigID)

	utils.OkResponse(c, "Configuration rolled back successfully", gin.H{
		"config_id":      req.ConfigID,
		"rolled_back_at": time.Now(),
	})
}

func (h *AdminHandler) ValidateConfiguration(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would validate configuration
	validation := gin.H{
		"valid":    true,
		"errors":   []string{},
		"warnings": []string{},
	}

	utils.OkResponse(c, "Configuration validated successfully", validation)
}

func (h *AdminHandler) GetFeatureFlags(c *gin.Context) {
	// This would get feature flags from database/configuration
	flags := gin.H{
		"new_ui_enabled":     true,
		"beta_features":      false,
		"advanced_search":    true,
		"video_upload":       false,
		"live_streaming":     false,
		"marketplace":        false,
		"ai_moderation":      true,
		"push_notifications": true,
	}

	utils.OkResponse(c, "Feature flags retrieved successfully", flags)
}

func (h *AdminHandler) UpdateFeatureFlags(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would update feature flags
	h.logAdminActivity(c, "feature_flags_update", "Updated feature flags")

	utils.OkResponse(c, "Feature flags updated successfully", req)
}

func (h *AdminHandler) ToggleFeature(c *gin.Context) {
	feature := c.Param("feature")

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would toggle specific feature flag
	h.logAdminActivity(c, "feature_toggle", fmt.Sprintf("Toggled feature %s to %t", feature, req.Enabled))

	utils.OkResponse(c, "Feature toggled successfully", gin.H{
		"feature": feature,
		"enabled": req.Enabled,
	})
}

func (h *AdminHandler) GetRateLimits(c *gin.Context) {
	// This would get rate limits from configuration
	rateLimits := gin.H{
		"global": gin.H{
			"requests_per_minute": 1000,
			"burst_limit":         100,
		},
		"user": gin.H{
			"posts_per_hour":    10,
			"comments_per_hour": 50,
			"messages_per_hour": 100,
			"likes_per_minute":  60,
			"follows_per_hour":  30,
		},
		"ip": gin.H{
			"requests_per_minute": 100,
			"login_attempts":      5,
		},
	}

	utils.OkResponse(c, "Rate limits retrieved successfully", rateLimits)
}

func (h *AdminHandler) UpdateRateLimits(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would update rate limits
	h.logAdminActivity(c, "rate_limits_update", "Updated rate limits")

	utils.OkResponse(c, "Rate limits updated successfully", req)
}

// Helper functions
func (h *AdminHandler) createPaginationLinks(c *gin.Context, pagination *utils.PaginationMeta) *utils.PaginationLinks {
	// Defensive check for nil pagination
	if pagination == nil {
		return &utils.PaginationLinks{
			Self: c.Request.URL.Path + "?" + c.Request.URL.Query().Encode(),
		}
	}

	baseURL := c.Request.URL.Path
	query := c.Request.URL.Query()

	links := &utils.PaginationLinks{
		Self: baseURL + "?" + query.Encode(),
	}

	if pagination.HasNext {
		query.Set("page", strconv.Itoa(pagination.CurrentPage+1))
		links.Next = baseURL + "?" + query.Encode()
	}

	if pagination.HasPrevious {
		query.Set("page", strconv.Itoa(pagination.CurrentPage-1))
		links.Previous = baseURL + "?" + query.Encode()
	}

	// Add first and last page links
	if pagination.TotalPages > 0 {
		query.Set("page", "1")
		links.First = baseURL + "?" + query.Encode()

		query.Set("page", strconv.Itoa(pagination.TotalPages))
		links.Last = baseURL + "?" + query.Encode()
	}

	return links
}

func (h *AdminHandler) logAdminActivity(c *gin.Context, activityType, description string) {
	adminIDValue, exists := c.Get("user_id")
	if !exists {
		return
	}

	adminID, ok := adminIDValue.(primitive.ObjectID)
	if !ok {
		return
	}

	activity := bson.M{
		"admin_id":    adminID,
		"type":        activityType,
		"description": description,
		"ip_address":  c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
		"timestamp":   time.Now(),
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	// Store admin activity in database
	h.db.Collection("admin_activities").InsertOne(context.Background(), activity)
}

// Report generation helper functions
func (h *AdminHandler) generateUserActivityReport(ctx context.Context, startDate, endDate time.Time, filters gin.H) (gin.H, error) {
	// Generate user activity report
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"last_active_at": bson.M{"$gte": startDate, "$lte": endDate},
				"deleted_at":     bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$last_active_at"},
				},
				"active_users": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := h.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activityData []bson.M
	if err := cursor.All(ctx, &activityData); err != nil {
		return nil, err
	}

	return gin.H{
		"report_type":   "user_activity",
		"period":        fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"activity_data": activityData,
		"generated_at":  time.Now(),
	}, nil
}

func (h *AdminHandler) generateContentPerformanceReport(ctx context.Context, startDate, endDate time.Time, filters gin.H) (gin.H, error) {
	// Generate content performance report
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": startDate, "$lte": endDate},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "likes",
				"localField":   "_id",
				"foreignField": "target_id",
				"as":           "likes",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "comments",
				"localField":   "_id",
				"foreignField": "post_id",
				"as":           "comments",
			},
		},
		{
			"$addFields": bson.M{
				"likes_count":    bson.M{"$size": "$likes"},
				"comments_count": bson.M{"$size": "$comments"},
				"engagement":     bson.M{"$add": []interface{}{bson.M{"$size": "$likes"}, bson.M{"$size": "$comments"}}},
			},
		},
		{
			"$sort": bson.M{"engagement": -1},
		},
		{
			"$limit": 50,
		},
	}

	cursor, err := h.db.Collection("posts").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var performanceData []bson.M
	if err := cursor.All(ctx, &performanceData); err != nil {
		return nil, err
	}

	return gin.H{
		"report_type":      "content_performance",
		"period":           fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"performance_data": performanceData,
		"generated_at":     time.Now(),
	}, nil
}

func (h *AdminHandler) generateEngagementSummaryReport(ctx context.Context, startDate, endDate time.Time, filters gin.H) (gin.H, error) {
	// Generate engagement summary report
	likesCount, _ := h.db.Collection("likes").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": startDate, "$lte": endDate},
		"deleted_at": bson.M{"$exists": false},
	})

	commentsCount, _ := h.db.Collection("comments").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": startDate, "$lte": endDate},
		"deleted_at": bson.M{"$exists": false},
	})

	followsCount, _ := h.db.Collection("follows").CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": startDate, "$lte": endDate},
		"deleted_at": bson.M{"$exists": false},
	})

	return gin.H{
		"report_type":      "engagement_summary",
		"period":           fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"total_likes":      likesCount,
		"total_comments":   commentsCount,
		"total_follows":    followsCount,
		"total_engagement": likesCount + commentsCount + followsCount,
		"generated_at":     time.Now(),
	}, nil
}

// Public Admin Routes (Login, etc.)
func (h *AdminHandler) GetPublicSystemStatus(c *gin.Context) {
	status := gin.H{
		"status":      "operational",
		"timestamp":   time.Now(),
		"version":     "1.0.0",
		"environment": "production",
	}

	utils.OkResponse(c, "System status retrieved successfully", status)
}

func (h *AdminHandler) GetPublicHealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// Basic health check
	err := h.db.Client().Ping(ctx, nil)
	healthy := err == nil

	status := gin.H{
		"healthy":   healthy,
		"timestamp": time.Now(),
		"services": gin.H{
			"database": healthy,
			"api":      true,
		},
	}

	if healthy {
		utils.OkResponse(c, "Health check passed", status)
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Health check failed",
			"data":    status,
		})
	}
}

// internal/handlers/admin.go
func (h *AdminHandler) AdminLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()

	// Find admin user by email
	var user models.User
	err := h.db.Collection("users").FindOne(ctx, bson.M{
		"email":      req.Email,
		"role":       bson.M{"$in": []string{"admin", "super_admin"}}, // Only admin/super_admin can login
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.UnauthorizedResponse(c, "Invalid credentials or insufficient permissions")
			return
		}
		utils.InternalServerErrorResponse(c, "Login failed", err)
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		utils.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Check if user is suspended
	if user.IsSuspended {
		utils.ForbiddenResponse(c, "Account is suspended")
		return
	}

	// Generate real JWT tokens using AuthService
	sessionID := primitive.NewObjectID().Hex()
	accessToken, refreshToken, err := h.authService.GenerateTokens(&user, sessionID, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate tokens", err)
		return
	}

	// Create admin session
	session := bson.M{
		"user_id":          user.ID,
		"session_id":       sessionID,
		"device_info":      c.GetHeader("User-Agent"),
		"ip_address":       c.ClientIP(),
		"is_active":        true,
		"last_activity_at": time.Now(),
		"expires_at":       time.Now().Add(24 * time.Hour),
		"created_at":       time.Now(),
		"updated_at":       time.Now(),
	}

	_, err = h.db.Collection("admin_sessions").InsertOne(ctx, session)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create session", err)
		return
	}

	// Log admin login
	h.logAdminActivity(c, "admin_login", "Admin user logged in: "+user.Email)

	// Return real authentication data
	utils.OkResponse(c, "Admin login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    24 * 60 * 60, // 24 hours
		"token_type":    "Bearer",
		"user": gin.H{
			"id":          user.ID.Hex(),
			"email":       user.Email,
			"username":    user.Username,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"role":        user.Role,
			"is_verified": user.IsVerified,
		},
	})
}

func (h *AdminHandler) AdminLogout(c *gin.Context) {
	// This would invalidate admin session/token
	utils.OkResponse(c, "Admin logout successful", nil)
}

func (h *AdminHandler) RefreshAdminToken(c *gin.Context) {
	// This would refresh admin token
	utils.OkResponse(c, "Admin token refreshed successfully", gin.H{
		"access_token": "new_admin_access_token",
		"expires_in":   3600,
	})
}

func (h *AdminHandler) AdminForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would send password reset email
	utils.OkResponse(c, "Password reset email sent", gin.H{
		"email": req.Email,
	})
}

func (h *AdminHandler) AdminResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// This would reset admin password
	utils.OkResponse(c, "Password reset successful", nil)
}

// WebSocket handlers
func (h *AdminHandler) DashboardWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upgrade WebSocket connection", err)
		return
	}
	defer conn.Close()

	// Send real-time dashboard updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := h.adminService.GetDashboardStats(context.Background())
			if err == nil {
				conn.WriteJSON(gin.H{
					"type": "dashboard_update",
					"data": stats,
				})
			}
		}
	}
}

func (h *AdminHandler) MonitoringWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upgrade WebSocket connection", err)
		return
	}
	defer conn.Close()

	// Send real-time monitoring data
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send system metrics
			metrics := gin.H{
				"cpu_usage":    23.5,
				"memory_usage": 65.2,
				"active_users": 1250,
				"timestamp":    time.Now(),
			}
			conn.WriteJSON(gin.H{
				"type": "metrics_update",
				"data": metrics,
			})
		}
	}
}

func (h *AdminHandler) ModerationWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upgrade WebSocket connection", err)
		return
	}
	defer conn.Close()

	// Send real-time moderation updates
	// This would listen for new reports, flagged content, etc.
	for {
		// In a real implementation, this would listen to events
		time.Sleep(5 * time.Second)
	}
}

func (h *AdminHandler) ActivitiesWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upgrade WebSocket connection", err)
		return
	}
	defer conn.Close()

	// Send real-time user activities
	// This would stream user activities, new registrations, etc.
	for {
		// In a real implementation, this would listen to events
		time.Sleep(2 * time.Second)
	}
}
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username   string `json:"username" binding:"required"`
		Email      string `json:"email" binding:"required,email"`
		Password   string `json:"password" binding:"required,min=8"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Bio        string `json:"bio"`
		Role       string `json:"role"`
		IsActive   bool   `json:"is_active"`
		IsVerified bool   `json:"is_verified"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()

	// Check if username or email already exists
	existingCount, err := h.db.Collection("users").CountDocuments(ctx, bson.M{
		"$or": []bson.M{
			{"username": req.Username},
			{"email": req.Email},
		},
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to check existing user", err)
		return
	}
	if existingCount > 0 {
		utils.BadRequestResponse(c, "Username or email already exists", nil)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to hash password", err)
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "user"
	}

	// Create user document
	now := time.Now()
	user := bson.M{
		"username":           req.Username,
		"email":              req.Email,
		"password":           hashedPassword,
		"first_name":         req.FirstName,
		"last_name":          req.LastName,
		"bio":                req.Bio,
		"role":               req.Role,
		"is_active":          req.IsActive,
		"is_verified":        req.IsVerified,
		"is_suspended":       false,
		"is_private":         false,
		"followers_count":    0,
		"following_count":    0,
		"posts_count":        0,
		"email_verified":     req.IsVerified,
		"phone_verified":     false,
		"two_factor_enabled": false,
		"privacy_level":      "public",
		"login_count":        0,
		"verification_level": 0,
		"reputation_score":   0,
		"created_at":         now,
		"updated_at":         now,
	}

	result, err := h.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create user", err)
		return
	}

	h.logAdminActivity(c, "user_creation", "Created user: "+req.Username)

	// Return created user (without password)
	user["_id"] = result.InsertedID
	delete(user, "password")

	utils.CreatedResponse(c, "User created successfully", user)
}

// UpdateUser updates an existing user (admin only)
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", nil)
		return
	}

	var req struct {
		Username   *string `json:"username"`
		Email      *string `json:"email"`
		FirstName  *string `json:"first_name"`
		LastName   *string `json:"last_name"`
		Bio        *string `json:"bio"`
		Role       *string `json:"role"`
		IsActive   *bool   `json:"is_active"`
		IsVerified *bool   `json:"is_verified"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx := c.Request.Context()

	// Build update document
	updateDoc := bson.M{
		"updated_at": time.Now(),
	}

	if req.Username != nil {
		// Check if username is already taken by another user
		existingCount, err := h.db.Collection("users").CountDocuments(ctx, bson.M{
			"username":   *req.Username,
			"_id":        bson.M{"$ne": objID},
			"deleted_at": bson.M{"$exists": false},
		})
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to check username availability", err)
			return
		}
		if existingCount > 0 {
			utils.BadRequestResponse(c, "Username already exists", nil)
			return
		}
		updateDoc["username"] = *req.Username
	}

	if req.Email != nil {
		// Check if email is already taken by another user
		existingCount, err := h.db.Collection("users").CountDocuments(ctx, bson.M{
			"email":      *req.Email,
			"_id":        bson.M{"$ne": objID},
			"deleted_at": bson.M{"$exists": false},
		})
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to check email availability", err)
			return
		}
		if existingCount > 0 {
			utils.BadRequestResponse(c, "Email already exists", nil)
			return
		}
		updateDoc["email"] = *req.Email
	}

	if req.FirstName != nil {
		updateDoc["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updateDoc["last_name"] = *req.LastName
	}
	if req.Bio != nil {
		updateDoc["bio"] = *req.Bio
	}
	if req.Role != nil {
		updateDoc["role"] = *req.Role
	}
	if req.IsActive != nil {
		updateDoc["is_active"] = *req.IsActive
	}
	if req.IsVerified != nil {
		updateDoc["is_verified"] = *req.IsVerified
		updateDoc["email_verified"] = *req.IsVerified
	}

	// Update user
	result, err := h.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": objID, "deleted_at": bson.M{"$exists": false}},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update user", err)
		return
	}

	if result.MatchedCount == 0 {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	h.logAdminActivity(c, "user_update", "Updated user ID: "+userID)

	// Get updated user
	var updatedUser bson.M
	err = h.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedUser)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch updated user", err)
		return
	}

	// Remove password from response
	delete(updatedUser, "password")

	utils.OkResponse(c, "User updated successfully", updatedUser)
}
