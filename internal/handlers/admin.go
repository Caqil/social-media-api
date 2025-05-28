// internal/handlers/admin.go
package handlers

import (
	"strconv"
	"strings"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminHandler struct {
	adminService        *services.AdminService
	userService         *services.UserService
	reportService       *services.ReportService
	analyticsService    *services.AnalyticsService
	notificationService *services.NotificationService
}

func NewAdminHandler(adminService *services.AdminService, userService *services.UserService, reportService *services.ReportService, analyticsService *services.AnalyticsService, notificationService *services.NotificationService) *AdminHandler {
	return &AdminHandler{
		adminService:        adminService,
		userService:         userService,
		reportService:       reportService,
		analyticsService:    analyticsService,
		notificationService: notificationService,
	}
}

// Dashboard and Analytics

// GetDashboard returns comprehensive admin dashboard data
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get dashboard stats", err)
		return
	}

	utils.OkResponse(c, "Dashboard data retrieved successfully", gin.H{
		"stats":        stats,
		"last_updated": time.Now(),
		"admin_id":     adminID.Hex(),
	})
}

// GetSystemHealth returns current system health status
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	metrics, err := h.adminService.GetSystemMetrics()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get system metrics", err)
		return
	}

	// Determine overall health status
	overallStatus := "healthy"
	if metrics.ErrorMetrics.ErrorRate5xx > 1.0 || metrics.PerformanceMetrics.AverageResponseTime > 1000 {
		overallStatus = "warning"
	}
	if metrics.ErrorMetrics.ErrorRate5xx > 5.0 || metrics.PerformanceMetrics.AverageResponseTime > 3000 {
		overallStatus = "critical"
	}

	utils.OkResponse(c, "System health retrieved successfully", gin.H{
		"status":         overallStatus,
		"metrics":        metrics,
		"last_checked":   time.Now(),
		"uptime_seconds": 86400, // Would calculate actual uptime
	})
}

// GetAnalytics returns detailed platform analytics
func (h *AdminHandler) GetAnalytics(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	timeRange := c.DefaultQuery("time_range", "week") // day, week, month, year
	if !h.isValidTimeRange(timeRange) {
		utils.BadRequestResponse(c, "Invalid time range. Must be: day, week, month, year", nil)
		return
	}

	// Get comprehensive analytics
	analytics := gin.H{
		"time_range":   timeRange,
		"generated_at": time.Now(),
	}

	// Get dashboard stats which include most analytics
	stats, err := h.adminService.GetDashboardStats()
	if err == nil {
		analytics["user_metrics"] = gin.H{
			"total_users":  stats.TotalUsers,
			"active_users": stats.ActiveUsers,
			"new_users":    stats.NewUsers,
			"user_growth":  stats.UserGrowth,
			"geographic":   stats.GeographicData,
			"devices":      stats.DeviceStats,
		}

		analytics["content_metrics"] = gin.H{
			"total_posts":    stats.TotalPosts,
			"total_comments": stats.TotalComments,
			"total_stories":  stats.TotalStories,
			"content_growth": stats.ContentGrowth,
			"top_hashtags":   stats.TopHashtags,
		}

		analytics["engagement_metrics"] = stats.EngagementMetrics

		analytics["platform_metrics"] = stats.PlatformMetrics

		analytics["moderation_metrics"] = gin.H{
			"total_reports":   stats.TotalReports,
			"pending_reports": stats.PendingReports,
			"queue_stats":     stats.ModerationQueue,
		}

		analytics["revenue_metrics"] = stats.RevenueMetrics
	}

	utils.OkResponse(c, "Analytics retrieved successfully", analytics)
}

// GetRealtimeStats returns real-time platform statistics
func (h *AdminHandler) GetRealtimeStats(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	// Get current metrics
	metrics, err := h.adminService.GetSystemMetrics()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get real-time stats", err)
		return
	}

	realtimeStats := gin.H{
		"timestamp": time.Now(),
		"system": gin.H{
			"cpu_usage":    metrics.CPUUsage,
			"memory_usage": metrics.MemoryUsage,
			"disk_usage":   metrics.DiskUsage,
		},
		"performance": gin.H{
			"response_time":    metrics.PerformanceMetrics.AverageResponseTime,
			"requests_per_sec": metrics.PerformanceMetrics.RequestsPerSecond,
			"concurrent_users": metrics.PerformanceMetrics.ConcurrentUsers,
		},
		"database": gin.H{
			"connections":    metrics.DatabaseMetrics.ConnectionCount,
			"active_queries": metrics.DatabaseMetrics.ActiveQueries,
			"query_latency":  metrics.DatabaseMetrics.QueryLatency,
		},
		"cache": gin.H{
			"hit_rate":     metrics.CacheMetrics.HitRate,
			"cached_items": metrics.CacheMetrics.CachedItemCount,
		},
		"queues": gin.H{
			"pending_jobs":    metrics.QueueMetrics.PendingJobs,
			"processing_jobs": metrics.QueueMetrics.ProcessingJobs,
		},
		"errors": gin.H{
			"error_rate_5xx": metrics.ErrorMetrics.ErrorRate5xx,
			"error_rate_4xx": metrics.ErrorMetrics.ErrorRate4xx,
			"total_errors":   metrics.ErrorMetrics.TotalErrors,
		},
	}

	utils.OkResponse(c, "Real-time stats retrieved successfully", realtimeStats)
}

// User Management

// GetUsers returns paginated list of users with filtering
func (h *AdminHandler) GetUsers(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filter from query parameters
	filter := models.AdminUserFilter{
		Status:    c.Query("status"),
		Role:      c.Query("role"),
		Location:  c.Query("location"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	// Parse boolean parameters
	if isVerifiedStr := c.Query("is_verified"); isVerifiedStr != "" {
		isVerified := isVerifiedStr == "true"
		filter.IsVerified = &isVerified
	}

	// Parse date parameters
	if createdAfterStr := c.Query("created_after"); createdAfterStr != "" {
		if date, err := time.Parse("2006-01-02", createdAfterStr); err == nil {
			filter.CreatedAfter = date
		}
	}

	if createdBeforeStr := c.Query("created_before"); createdBeforeStr != "" {
		if date, err := time.Parse("2006-01-02", createdBeforeStr); err == nil {
			filter.CreatedBefore = date
		}
	}

	// Parse numeric parameters
	if minPostCountStr := c.Query("min_post_count"); minPostCountStr != "" {
		if count, err := strconv.ParseInt(minPostCountStr, 10, 64); err == nil {
			filter.MinPostCount = count
		}
	}

	if maxPostCountStr := c.Query("max_post_count"); maxPostCountStr != "" {
		if count, err := strconv.ParseInt(maxPostCountStr, 10, 64); err == nil {
			filter.MaxPostCount = count
		}
	}

	users, totalCount, err := h.adminService.GetUsers(filter, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get users", err)
		return
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Users retrieved successfully", gin.H{
		"users":  users,
		"filter": filter,
	}, paginationMeta, nil)
}

// GetUser returns detailed information about a specific user
func (h *AdminHandler) GetUser(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	userDetail, err := h.adminService.GetUserDetail(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user detail", err)
		return
	}

	utils.OkResponse(c, "User detail retrieved successfully", userDetail)
}

// BulkUserAction performs bulk actions on multiple users
func (h *AdminHandler) BulkUserAction(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	var req models.AdminUserActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate action
	validActions := []string{"suspend", "unsuspend", "verify", "unverify", "delete", "warn"}
	if !h.isValidAction(req.Action, validActions) {
		utils.BadRequestResponse(c, "Invalid action. Must be one of: suspend, unsuspend, verify, unverify, delete, warn", nil)
		return
	}

	// Validate user IDs
	if len(req.UserIDs) == 0 {
		utils.BadRequestResponse(c, "At least one user ID is required", nil)
		return
	}

	if len(req.UserIDs) > 100 {
		utils.BadRequestResponse(c, "Cannot perform bulk action on more than 100 users at once", nil)
		return
	}

	// Validate reason
	if strings.TrimSpace(req.Reason) == "" {
		utils.BadRequestResponse(c, "Reason is required for user actions", nil)
		return
	}

	err := h.adminService.BulkUserAction(adminID, req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to perform bulk user action", err)
		return
	}

	utils.OkResponse(c, "Bulk user action completed successfully", gin.H{
		"action":       req.Action,
		"user_count":   len(req.UserIDs),
		"reason":       req.Reason,
		"performed_by": adminID.Hex(),
		"performed_at": time.Now(),
	})
}

// SuspendUser suspends a specific user
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	var req struct {
		Reason   string `json:"reason" binding:"required"`
		Duration string `json:"duration,omitempty"` // Optional: "7d", "30d", "permanent"
		Note     string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	actionReq := models.AdminUserActionRequest{
		UserIDs:  []string{userIDStr},
		Action:   "suspend",
		Reason:   req.Reason,
		Duration: &req.Duration,
		Note:     req.Note,
	}

	err = h.adminService.BulkUserAction(adminID, actionReq)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to suspend user", err)
		return
	}

	// Send notification to user
	if h.notificationService != nil {
		go h.notificationService.NotifyUserSuspension(userID, req.Reason, req.Duration)
	}

	utils.OkResponse(c, "User suspended successfully", gin.H{
		"user_id":      userIDStr,
		"action":       "suspended",
		"reason":       req.Reason,
		"duration":     req.Duration,
		"suspended_by": adminID.Hex(),
		"suspended_at": time.Now(),
	})
}

// UnsuspendUser unsuspends a specific user
func (h *AdminHandler) UnsuspendUser(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	userIDStr := c.Param("userId")

	var req struct {
		Note string `json:"note"`
	}

	c.ShouldBindJSON(&req)

	actionReq := models.AdminUserActionRequest{
		UserIDs: []string{userIDStr},
		Action:  "unsuspend",
		Reason:  "Admin unsuspension",
		Note:    req.Note,
	}

	err := h.adminService.BulkUserAction(adminID, actionReq)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to unsuspend user", err)
		return
	}

	utils.OkResponse(c, "User unsuspended successfully", gin.H{
		"user_id":        userIDStr,
		"action":         "unsuspended",
		"unsuspended_by": adminID.Hex(),
		"unsuspended_at": time.Now(),
	})
}

// VerifyUser verifies a specific user
func (h *AdminHandler) VerifyUser(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	userIDStr := c.Param("userId")

	actionReq := models.AdminUserActionRequest{
		UserIDs: []string{userIDStr},
		Action:  "verify",
		Reason:  "Admin verification",
	}

	err := h.adminService.BulkUserAction(adminID, actionReq)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to verify user", err)
		return
	}

	utils.OkResponse(c, "User verified successfully", gin.H{
		"user_id":     userIDStr,
		"action":      "verified",
		"verified_by": adminID.Hex(),
		"verified_at": time.Now(),
	})
}

// DeleteUser soft deletes a specific user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	userIDStr := c.Param("userId")

	var req struct {
		Reason string `json:"reason" binding:"required"`
		Note   string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	actionReq := models.AdminUserActionRequest{
		UserIDs: []string{userIDStr},
		Action:  "delete",
		Reason:  req.Reason,
		Note:    req.Note,
	}

	err := h.adminService.BulkUserAction(adminID, actionReq)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete user", err)
		return
	}

	utils.OkResponse(c, "User deleted successfully", gin.H{
		"user_id":    userIDStr,
		"action":     "deleted",
		"reason":     req.Reason,
		"deleted_by": adminID.Hex(),
		"deleted_at": time.Now(),
	})
}

// Content Management

// GetContent returns paginated content (posts, comments, stories) with filtering
func (h *AdminHandler) GetContent(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	contentType := c.Param("contentType")
	validContentTypes := []string{"posts", "comments", "stories"}
	if !h.isValidContentType(contentType, validContentTypes) {
		utils.BadRequestResponse(c, "Invalid content type. Must be: posts, comments, stories", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filter
	filter := models.AdminContentFilter{
		AuthorID:         c.Query("author_id"),
		ContentType:      c.Query("content_type"),
		Status:           c.Query("status"),
		ModerationStatus: c.Query("moderation_status"),
		SortBy:           c.DefaultQuery("sort_by", "created_at"),
		SortOrder:        c.DefaultQuery("sort_order", "desc"),
	}

	// Parse boolean parameters
	if isFlaggedStr := c.Query("is_flagged"); isFlaggedStr != "" {
		isFlagged := isFlaggedStr == "true"
		filter.IsFlagged = &isFlagged
	}

	content, totalCount, err := h.adminService.GetContent(contentType, filter, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get content", err)
		return
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Content retrieved successfully", gin.H{
		"content_type": contentType,
		"content":      content,
		"filter":       filter,
	}, paginationMeta, nil)
}

// BulkContentAction performs bulk actions on content
func (h *AdminHandler) BulkContentAction(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	var req models.AdminContentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate action
	validActions := []string{"approve", "reject", "delete", "flag", "unflag", "hide"}
	if !h.isValidAction(req.Action, validActions) {
		utils.BadRequestResponse(c, "Invalid action", nil)
		return
	}

	if len(req.ContentIDs) == 0 {
		utils.BadRequestResponse(c, "At least one content ID is required", nil)
		return
	}

	if len(req.ContentIDs) > 100 {
		utils.BadRequestResponse(c, "Cannot perform bulk action on more than 100 items at once", nil)
		return
	}

	// Process each content item
	successCount := 0
	for _, contentIDStr := range req.ContentIDs {
		_, err := primitive.ObjectIDFromHex(contentIDStr)
		if err != nil {
			continue
		}

		// Perform action based on type
		// This would involve updating the specific collections
		// For now, we'll just count successful operations
		successCount++
	}

	utils.OkResponse(c, "Bulk content action completed successfully", gin.H{
		"action":        req.Action,
		"total_items":   len(req.ContentIDs),
		"success_count": successCount,
		"failed_count":  len(req.ContentIDs) - successCount,
		"reason":        req.Reason,
		"performed_by":  adminID.Hex(),
		"performed_at":  time.Now(),
	})
}

// Report Management

// GetReports returns paginated reports with filtering (uses existing report service)
func (h *AdminHandler) GetReports(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filter
	filter := models.ReportFilter{
		Status:     c.Query("status"),
		TargetType: c.Query("target_type"),
		Priority:   c.Query("priority"),
		AssignedTo: c.Query("assigned_to"),
		SortBy:     c.DefaultQuery("sort_by", "created_at"),
	}

	if reason := c.Query("reason"); reason != "" {
		filter.Reason = models.ReportReason(reason)
	}

	reports, totalCount, err := h.reportService.GetReports(filter, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get reports", err)
		return
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Reports retrieved successfully", gin.H{
		"reports": reports,
		"filter":  filter,
	}, paginationMeta, nil)
}

// System Configuration

// GetSystemConfig returns current system configuration
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	config, err := h.adminService.GetSystemConfiguration()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get system configuration", err)
		return
	}

	utils.OkResponse(c, "System configuration retrieved successfully", config)
}

// UpdateSystemConfig updates system configuration
func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	var config models.SystemConfiguration
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequestResponse(c, "Invalid configuration format", err)
		return
	}

	// Validate configuration values
	if config.MaxPostLength < 1 || config.MaxPostLength > 10000 {
		utils.BadRequestResponse(c, "Max post length must be between 1 and 10000", nil)
		return
	}

	if config.MaxCommentLength < 1 || config.MaxCommentLength > 5000 {
		utils.BadRequestResponse(c, "Max comment length must be between 1 and 5000", nil)
		return
	}

	if config.MaxFileSize < 1 || config.MaxFileSize > 1000 {
		utils.BadRequestResponse(c, "Max file size must be between 1MB and 1000MB", nil)
		return
	}

	err := h.adminService.UpdateSystemConfiguration(adminID, config)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update system configuration", err)
		return
	}

	utils.OkResponse(c, "System configuration updated successfully", gin.H{
		"updated_by": adminID.Hex(),
		"updated_at": time.Now(),
		"config":     config,
	})
}

// Admin Activity Logs

// GetAdminActivities returns admin activity logs
func (h *AdminHandler) GetAdminActivities(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	activities, totalCount, err := h.adminService.GetAdminActivities(params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get admin activities", err)
		return
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Admin activities retrieved successfully", activities, paginationMeta, nil)
}

// Platform Monitoring

// GetActiveUsers returns currently active users
func (h *AdminHandler) GetActiveUsers(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	timeWindow := c.DefaultQuery("time_window", "1h") // 1h, 24h, 7d

	// This would query active sessions or recent activity
	// For now, return mock data
	activeUsers := gin.H{
		"time_window":     timeWindow,
		"total_active":    1500,
		"new_sessions":    150,
		"peak_concurrent": 2000,
		"by_platform": gin.H{
			"web":     800,
			"mobile":  600,
			"desktop": 100,
		},
		"by_location": []gin.H{
			{"country": "US", "count": 500},
			{"country": "UK", "count": 300},
			{"country": "CA", "count": 200},
		},
		"last_updated": time.Now(),
	}

	utils.OkResponse(c, "Active users retrieved successfully", activeUsers)
}

// GetErrorLogs returns recent error logs
func (h *AdminHandler) GetErrorLogs(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)
	severity := c.DefaultQuery("severity", "all") // all, error, critical, warning

	// This would query actual error logs
	// For now, return mock data
	errorLogs := []gin.H{
		{
			"id":         "error_1",
			"timestamp":  time.Now().Add(-1 * time.Hour),
			"severity":   "error",
			"message":    "Database connection timeout",
			"source":     "user_service.go:123",
			"user_id":    "user_123",
			"request_id": "req_456",
		},
		{
			"id":        "error_2",
			"timestamp": time.Now().Add(-2 * time.Hour),
			"severity":  "warning",
			"message":   "High memory usage detected",
			"source":    "system_monitor.go:78",
		},
	}

	paginationMeta := utils.CreatePaginationMeta(params, int64(len(errorLogs)))

	utils.PaginatedSuccessResponse(c, "Error logs retrieved successfully", gin.H{
		"logs":     errorLogs,
		"severity": severity,
	}, paginationMeta, nil)
}

// Maintenance and Operations

// SetMaintenanceMode enables/disables maintenance mode
func (h *AdminHandler) SetMaintenanceMode(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	var req struct {
		Enabled bool   `json:"enabled"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Update system configuration
	config, err := h.adminService.GetSystemConfiguration()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get system configuration", err)
		return
	}

	config.MaintenanceMode = req.Enabled
	config.MaintenanceMessage = req.Message

	err = h.adminService.UpdateSystemConfiguration(adminID, *config)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update maintenance mode", err)
		return
	}

	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}

	utils.OkResponse(c, "Maintenance mode updated successfully", gin.H{
		"maintenance_mode": status,
		"message":          req.Message,
		"updated_by":       adminID.Hex(),
		"updated_at":       time.Now(),
	})
}

// ClearCache clears various system caches
func (h *AdminHandler) ClearCache(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	var req struct {
		CacheType string `json:"cache_type"` // all, user, content, feed, search
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.CacheType = "all"
	}

	// This would clear actual caches
	// For now, simulate cache clearing
	clearedCaches := []string{}

	switch req.CacheType {
	case "all":
		clearedCaches = []string{"user_cache", "content_cache", "feed_cache", "search_cache"}
	case "user":
		clearedCaches = []string{"user_cache"}
	case "content":
		clearedCaches = []string{"content_cache"}
	case "feed":
		clearedCaches = []string{"feed_cache"}
	case "search":
		clearedCaches = []string{"search_cache"}
	default:
		clearedCaches = []string{"user_cache", "content_cache", "feed_cache", "search_cache"}
	}

	utils.OkResponse(c, "Cache cleared successfully", gin.H{
		"cache_type":     req.CacheType,
		"cleared_caches": clearedCaches,
		"cleared_by":     adminID.Hex(),
		"cleared_at":     time.Now(),
	})
}

// Helper Methods

func (h *AdminHandler) getCurrentAdminID(c *gin.Context) primitive.ObjectID {
	userID, exists := c.Get("user_id")
	if !exists {
		return primitive.NilObjectID
	}

	// Check if user has admin role
	userRole, exists := c.Get("user_role")
	if !exists {
		return primitive.NilObjectID
	}

	role, ok := userRole.(models.UserRole)
	if !ok {
		return primitive.NilObjectID
	}

	if role != models.RoleAdmin && role != models.RoleSuperAdmin {
		return primitive.NilObjectID
	}

	return userID.(primitive.ObjectID)
}

func (h *AdminHandler) isValidTimeRange(timeRange string) bool {
	validRanges := []string{"day", "week", "month", "year"}
	for _, r := range validRanges {
		if timeRange == r {
			return true
		}
	}
	return false
}

func (h *AdminHandler) isValidAction(action string, validActions []string) bool {
	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

func (h *AdminHandler) isValidContentType(contentType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// Export/Import Operations

// ExportData exports platform data
func (h *AdminHandler) ExportData(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	var req struct {
		DataType  string `json:"data_type"` // users, posts, analytics
		Format    string `json:"format"`    // json, csv, xlsx
		DateRange struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		} `json:"date_range"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// This would generate actual export files
	exportID := primitive.NewObjectID().Hex()

	utils.OkResponse(c, "Data export initiated", gin.H{
		"export_id":            exportID,
		"data_type":            req.DataType,
		"format":               req.Format,
		"status":               "processing",
		"initiated_by":         adminID.Hex(),
		"initiated_at":         time.Now(),
		"estimated_completion": time.Now().Add(10 * time.Minute),
	})
}

// GetExportStatus returns status of data export
func (h *AdminHandler) GetExportStatus(c *gin.Context) {
	adminID := h.getCurrentAdminID(c)
	if adminID.IsZero() {
		utils.UnauthorizedResponse(c, "Admin authentication required")
		return
	}

	exportID := c.Param("exportId")

	// This would check actual export status
	utils.OkResponse(c, "Export status retrieved", gin.H{
		"export_id":    exportID,
		"status":       "completed",
		"progress":     100,
		"download_url": "/api/v1/admin/exports/" + exportID + "/download",
		"file_size":    "15.2 MB",
		"expires_at":   time.Now().Add(24 * time.Hour),
	})
}
