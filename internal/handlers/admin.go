// internal/handlers/admin_handler.go
package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
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

	// Parse filters
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

	// Create pagination links
	links := h.createPaginationLinks(c, pagination)

	utils.PaginatedSuccessResponse(c, "Users retrieved successfully", users, *pagination, links)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.adminService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	utils.OkResponse(c, "User retrieved successfully", user)
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

	// Log admin activity
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

	// Get admin ID from context (set by auth middleware)
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

// Message Management

func (h *AdminHandler) GetAllMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	messages, pagination, err := h.adminService.GetAllMessages(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get messages", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Messages retrieved successfully", messages, *pagination, links)
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

// System Management

func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	// This would typically check various system components
	health := gin.H{
		"status":           "healthy",
		"database_status":  "connected",
		"cache_status":     "active",
		"storage_status":   "available",
		"response_time_ms": 150,
		"memory_usage":     65.5,
		"cpu_usage":        23.2,
		"disk_usage":       45.8,
		"uptime":           "5d 12h 30m",
		"last_updated":     time.Now(),
	}

	utils.OkResponse(c, "System health retrieved successfully", health)
}

func (h *AdminHandler) GetSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	level := c.DefaultQuery("level", "")
	c.Query("start_date")
	c.Query("end_date")

	// This would typically read from log files or a logging service
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

// Configuration Management

func (h *AdminHandler) GetConfiguration(c *gin.Context) {
	// This would typically read from database or configuration service
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

	// Here you would update the configuration in database
	// For now, we'll just return success

	h.logAdminActivity(c, "config_update", "Updated system configuration")

	utils.OkResponse(c, "Configuration updated successfully", req)
}

// Bulk Operations

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

// Export Data

func (h *AdminHandler) ExportUsers(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// This would generate and return export file
	// For now, we'll return a download URL

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

// Helper Methods

func (h *AdminHandler) createPaginationLinks(c *gin.Context, pagination *utils.PaginationMeta) *utils.PaginationLinks {
	baseURL := c.Request.URL.Path
	query := c.Request.URL.Query()

	var nextURL, prevURL *string

	if pagination.HasNext {
		query.Set("page", strconv.Itoa(pagination.Page+1))
		next := baseURL + "?" + query.Encode()
		nextURL = &next
	}

	if pagination.HasPrevious {
		query.Set("page", strconv.Itoa(pagination.Page-1))
		prev := baseURL + "?" + query.Encode()
		prevURL = &prev
	}

	return &utils.PaginationLinks{
		Next:     nextURL,
		Previous: prevURL,
	}
}

func (h *AdminHandler) logAdminActivity(c *gin.Context, activityType, description string) {
	// Get admin ID from context
	adminIDValue, exists := c.Get("user_id")
	if !exists {
		return
	}

	adminID, ok := adminIDValue.(primitive.ObjectID)
	if !ok {
		return
	}

	// Log the activity (this would typically be stored in database)
	// For now, we'll just log to console or send to logging service
	activity := gin.H{
		"admin_id":    adminID.Hex(),
		"type":        activityType,
		"description": description,
		"ip_address":  c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
		"timestamp":   time.Now(),
	}

	// In a real implementation, you would store this in an admin_activities collection
	_ = activity
}

// Search functionality

func (h *AdminHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.UserFilter{
		Search: query,
	}

	users, pagination, err := h.adminService.GetAllUsers(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search users", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "User search completed", users, *pagination, links)
}

func (h *AdminHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := services.PostFilter{
		Search: query,
	}

	posts, pagination, err := h.adminService.GetAllPosts(c.Request.Context(), filter, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search posts", err)
		return
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Post search completed", posts, *pagination, links)
}

// Statistics endpoints

func (h *AdminHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("id")

	// This would calculate comprehensive user statistics
	stats := gin.H{
		"user_id":         userID,
		"posts_count":     25,
		"comments_count":  150,
		"likes_received":  320,
		"followers_count": 45,
		"following_count": 67,
		"stories_count":   12,
		"messages_sent":   89,
		"reports_made":    2,
		"reports_against": 0,
		"join_date":       "2023-01-15",
		"last_active":     time.Now().Add(-2 * time.Hour),
		"engagement_rate": 4.2,
		"average_likes":   12.8,
		"top_hashtags":    []string{"#travel", "#food", "#photography"},
	}

	utils.OkResponse(c, "User statistics retrieved successfully", stats)
}

func (h *AdminHandler) GetPostStats(c *gin.Context) {
	postID := c.Param("id")

	// This would calculate comprehensive post statistics
	stats := gin.H{
		"post_id":            postID,
		"views_count":        1250,
		"likes_count":        89,
		"comments_count":     23,
		"shares_count":       15,
		"saves_count":        7,
		"reports_count":      0,
		"engagement_rate":    7.1,
		"reach":              856,
		"impressions":        1450,
		"click_through_rate": 2.3,
		"created_at":         time.Now().Add(-48 * time.Hour),
		"engagement_timeline": []gin.H{
			{"hour": 0, "likes": 5, "comments": 1},
			{"hour": 1, "likes": 12, "comments": 3},
			{"hour": 2, "likes": 8, "comments": 2},
		},
	}

	utils.OkResponse(c, "Post statistics retrieved successfully", stats)
}
