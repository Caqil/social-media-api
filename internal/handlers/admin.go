package handlers

import (
	"strings"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminHandler struct {
	userService         *services.UserService
	postService         *services.PostService
	commentService      *services.CommentService
	storyService        *services.StoryService
	groupService        *services.GroupService
	messageService      *services.MessageService
	conversationService *services.ConversationService
	followService       *services.FollowService
	notificationService *services.NotificationService
	mediaService        *services.MediaService
	reportService       *services.ReportService
	searchService       *services.SearchService
	feedService         *services.FeedService
	likeService         *services.LikeService
	behaviorService     *services.UserBehaviorService
	validator           *validator.Validate
}

func NewAdminHandler(
	userService *services.UserService,
	postService *services.PostService,
	commentService *services.CommentService,
	storyService *services.StoryService,
	groupService *services.GroupService,
	messageService *services.MessageService,
	conversationService *services.ConversationService,
	followService *services.FollowService,
	notificationService *services.NotificationService,
	mediaService *services.MediaService,
	reportService *services.ReportService,
	searchService *services.SearchService,
	feedService *services.FeedService,
	likeService *services.LikeService,
	behaviorService *services.UserBehaviorService,
) *AdminHandler {
	return &AdminHandler{
		userService:         userService,
		postService:         postService,
		commentService:      commentService,
		storyService:        storyService,
		groupService:        groupService,
		messageService:      messageService,
		conversationService: conversationService,
		followService:       followService,
		notificationService: notificationService,
		mediaService:        mediaService,
		reportService:       reportService,
		searchService:       searchService,
		feedService:         feedService,
		likeService:         likeService,
		behaviorService:     behaviorService,
		validator:           validator.New(),
	}
}

// ==================== DASHBOARD & OVERVIEW ====================

// GetDashboardStats provides comprehensive dashboard statistics
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	timeRange := c.DefaultQuery("time_range", "week")

	// Get all statistics in parallel
	stats := gin.H{
		"overview":     h.getOverviewStats(),
		"users":        h.getUserStats(timeRange),
		"content":      h.getContentStats(timeRange),
		"engagement":   h.getEngagementStats(timeRange),
		"moderation":   h.getModerationStats(timeRange),
		"system":       h.getSystemStats(),
		"growth":       h.getGrowthStats(timeRange),
		"performance":  h.getPerformanceStats(),
		"generated_at": time.Now(),
		"time_range":   timeRange,
	}

	utils.OkResponse(c, "Dashboard statistics retrieved successfully", stats)
}

// GetSystemHealth provides system health monitoring
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": gin.H{
			"database": h.getDatabaseHealth(),
			"cache":    h.getCacheHealth(),
			"storage":  h.getStorageHealth(),
			"email":    h.getEmailHealth(),
			"push":     h.getPushHealth(),
			"search":   h.getSearchHealth(),
		},
		"metrics": gin.H{
			"uptime":            h.getUptime(),
			"memory_usage":      h.getMemoryUsage(),
			"cpu_usage":         h.getCPUUsage(),
			"disk_usage":        h.getDiskUsage(),
			"active_sessions":   h.getActiveSessionsCount(),
			"api_response_time": h.getAPIResponseTime(),
		},
	}

	utils.OkResponse(c, "System health retrieved successfully", health)
}

// GetRealTimeStats provides real-time statistics
func (h *AdminHandler) GetRealTimeStats(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	realTimeStats := gin.H{
		"timestamp":            time.Now(),
		"active_users":         h.getActiveUsersCount(),
		"online_users":         h.getOnlineUsersCount(),
		"recent_posts":         h.getRecentPostsCount(15), // Last 15 minutes
		"recent_comments":      h.getRecentCommentsCount(15),
		"recent_registrations": h.getRecentRegistrationsCount(60), // Last hour
		"active_conversations": h.getActiveConversationsCount(),
		"pending_reports":      h.getPendingReportsCount(),
		"system_alerts":        h.getSystemAlerts(),
		"api_requests":         h.getAPIRequestsCount(5), // Last 5 minutes
	}

	utils.OkResponse(c, "Real-time statistics retrieved successfully", realTimeStats)
}

// ==================== USER MANAGEMENT ====================

// GetUsers retrieves users with comprehensive filtering
func (h *AdminHandler) GetUsers(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filters
	filters := h.buildUserFilters(c)

	users, totalCount, err := h.getUsersWithFilters(filters, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get users", err)
		return
	}

	// Convert to admin response format
	var userResponses []gin.H
	for _, user := range users {
		userResponses = append(userResponses, h.userToAdminResponse(user))
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	response := gin.H{
		"users":   userResponses,
		"filters": filters,
		"summary": gin.H{
			"total_users":     totalCount,
			"active_users":    h.getActiveUsersInResults(users),
			"suspended_users": h.getSuspendedUsersInResults(users),
			"verified_users":  h.getVerifiedUsersInResults(users),
		},
	}

	utils.PaginatedSuccessResponse(c, "Users retrieved successfully", response, paginationMeta, nil)
}

// GetUserDetails provides comprehensive user details
func (h *AdminHandler) GetUserDetails(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	userDetails := gin.H{
		"user":          h.userToAdminResponse(*user),
		"statistics":    h.getUserStatistics(userID),
		"activity":      h.getUserActivity(userID),
		"content":       h.getUserContent(userID),
		"relationships": h.getUserRelationships(userID),
		"moderation":    h.getUserModerationHistory(userID),
		"sessions":      h.getUserSessions(userID),
		"devices":       h.getUserDevices(userID),
		"preferences":   h.getUserPreferences(userID),
	}

	utils.OkResponse(c, "User details retrieved successfully", userDetails)
}

// GetUserActivity retrieves user activity timeline
func (h *AdminHandler) GetUserActivity(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	timeRange := c.DefaultQuery("time_range", "week")
	activityType := c.Query("activity_type") // posts, comments, likes, etc.

	params := utils.GetPaginationParams(c)

	activities := h.getUserActivityTimeline(userID, timeRange, activityType, params.Limit, params.Offset)

	response := gin.H{
		"user_id":       userIDStr,
		"time_range":    timeRange,
		"activity_type": activityType,
		"activities":    activities,
		"summary":       h.getUserActivitySummary(userID, timeRange),
	}

	paginationMeta := utils.CreatePaginationMeta(params, int64(len(activities)))
	utils.PaginatedSuccessResponse(c, "User activity retrieved successfully", response, paginationMeta, nil)
}

// BulkUserAction performs bulk actions on users
func (h *AdminHandler) BulkUserAction(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	var req struct {
		UserIDs   []string `json:"user_ids" binding:"required"`
		Action    string   `json:"action" binding:"required"`
		Reason    string   `json:"reason,omitempty"`
		Duration  string   `json:"duration,omitempty"`
		Note      string   `json:"note,omitempty"`
		SendEmail bool     `json:"send_email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.UserIDs) == 0 {
		utils.BadRequestResponse(c, "At least one user ID is required", nil)
		return
	}

	if len(req.UserIDs) > 100 {
		utils.BadRequestResponse(c, "Cannot process more than 100 users at once", nil)
		return
	}

	if !h.isValidBulkAction(req.Action) {
		utils.BadRequestResponse(c, "Invalid bulk action", nil)
		return
	}

	currentUserID := c.GetString("user_id")
	results := h.performBulkUserAction(req.UserIDs, req.Action, req.Reason, req.Duration, req.Note, currentUserID, req.SendEmail)

	response := gin.H{
		"action":       req.Action,
		"total_users":  len(req.UserIDs),
		"successful":   results["successful"],
		"failed":       results["failed"],
		"results":      results["details"],
		"processed_at": time.Now(),
		"processed_by": currentUserID,
	}

	utils.OkResponse(c, "Bulk user action completed", response)
}

// ==================== CONTENT MANAGEMENT ====================

// GetContent retrieves content with filtering
func (h *AdminHandler) GetContent(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	contentType := c.DefaultQuery("content_type", "all") // posts, comments, stories
	status := c.Query("status")                          // published, deleted, reported
	authorID := c.Query("author_id")
	timeRange := c.DefaultQuery("time_range", "week")

	params := utils.GetPaginationParams(c)

	var content []gin.H
	var totalCount int64
	var err error

	switch contentType {
	case "posts":
		content, totalCount, err = h.getPostsForAdmin(status, authorID, timeRange, params.Limit, params.Offset)
	case "comments":
		content, totalCount, err = h.getCommentsForAdmin(status, authorID, timeRange, params.Limit, params.Offset)
	case "stories":
		content, totalCount, err = h.getStoriesForAdmin(status, authorID, timeRange, params.Limit, params.Offset)
	default:
		content, totalCount, err = h.getAllContentForAdmin(status, authorID, timeRange, params.Limit, params.Offset)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get content", err)
		return
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	response := gin.H{
		"content":      content,
		"content_type": contentType,
		"filters": gin.H{
			"status":     status,
			"author_id":  authorID,
			"time_range": timeRange,
		},
		"summary": h.getContentSummary(contentType, timeRange),
	}

	utils.PaginatedSuccessResponse(c, "Content retrieved successfully", response, paginationMeta, nil)
}

// GetContentDetails provides detailed content information
func (h *AdminHandler) GetContentDetails(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	contentType := c.Param("contentType")
	contentID := c.Param("contentId")

	if !h.isValidContentType(contentType) {
		utils.BadRequestResponse(c, "Invalid content type", nil)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(contentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid content ID format", err)
		return
	}

	details := h.getContentDetailsForAdmin(contentType, objectID)
	if details == nil {
		utils.NotFoundResponse(c, "Content not found")
		return
	}

	utils.OkResponse(c, "Content details retrieved successfully", details)
}

// BulkContentAction performs bulk actions on content
func (h *AdminHandler) BulkContentAction(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	var req struct {
		ContentIDs   []string `json:"content_ids" binding:"required"`
		ContentType  string   `json:"content_type" binding:"required"`
		Action       string   `json:"action" binding:"required"`
		Reason       string   `json:"reason,omitempty"`
		NotifyAuthor bool     `json:"notify_author"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if !h.isValidContentType(req.ContentType) {
		utils.BadRequestResponse(c, "Invalid content type", nil)
		return
	}

	if !h.isValidContentAction(req.Action) {
		utils.BadRequestResponse(c, "Invalid content action", nil)
		return
	}

	currentUserID := c.GetString("user_id")
	results := h.performBulkContentAction(req.ContentIDs, req.ContentType, req.Action, req.Reason, currentUserID, req.NotifyAuthor)

	response := gin.H{
		"action":       req.Action,
		"content_type": req.ContentType,
		"total_items":  len(req.ContentIDs),
		"results":      results,
		"processed_at": time.Now(),
		"processed_by": currentUserID,
	}

	utils.OkResponse(c, "Bulk content action completed", response)
}

// ==================== REPORTS & MODERATION ====================

// GetReports retrieves reports with comprehensive filtering
func (h *AdminHandler) GetReports(c *gin.Context) {
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Moderator access required")
		return
	}

	params := utils.GetPaginationParams(c)
	filters := h.buildReportFilters(c)

	reports, totalCount, err := h.reportService.GetReports(filters, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get reports", err)
		return
	}

	// Convert to admin response format
	var reportResponses []gin.H
	for _, report := range reports {
		reportResponses = append(reportResponses, h.reportToAdminResponse(report))
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	response := gin.H{
		"reports": reportResponses,
		"filters": filters,
		"summary": gin.H{
			"total_reports":   totalCount,
			"pending_reports": h.getPendingReportsCount(),
			"urgent_reports":  h.getUrgentReportsCount(),
			"my_assignments":  h.getMyAssignedReportsCount(c.GetString("user_id")),
		},
	}

	utils.PaginatedSuccessResponse(c, "Reports retrieved successfully", response, paginationMeta, nil)
}

// GetModerationQueue retrieves moderation queue
func (h *AdminHandler) GetModerationQueue(c *gin.Context) {
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Moderator access required")
		return
	}

	queueType := c.DefaultQuery("queue_type", "all")
	priority := c.Query("priority")
	params := utils.GetPaginationParams(c)

	queue := h.getModerationQueue(queueType, priority, params.Limit, params.Offset)

	response := gin.H{
		"queue":      queue,
		"queue_type": queueType,
		"priority":   priority,
		"summary":    h.getModerationQueueSummary(),
	}

	paginationMeta := utils.CreatePaginationMeta(params, int64(len(queue)))
	utils.PaginatedSuccessResponse(c, "Moderation queue retrieved successfully", response, paginationMeta, nil)
}

// GetModerationStats provides moderation statistics
func (h *AdminHandler) GetModerationStats(c *gin.Context) {
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Moderator access required")
		return
	}

	timeRange := c.DefaultQuery("time_range", "week")

	stats := gin.H{
		"time_range":         timeRange,
		"reports":            h.getReportStats(timeRange),
		"content_actions":    h.getContentActionStats(timeRange),
		"user_actions":       h.getUserActionStats(timeRange),
		"moderator_activity": h.getModeratorActivityStats(timeRange),
		"queue_health":       h.getQueueHealthStats(),
		"response_times":     h.getModerationResponseTimes(timeRange),
	}

	utils.OkResponse(c, "Moderation statistics retrieved successfully", stats)
}

// ==================== ANALYTICS & REPORTING ====================

// GetAnalytics provides comprehensive analytics
func (h *AdminHandler) GetAnalytics(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	analyticsType := c.DefaultQuery("type", "overview")
	timeRange := c.DefaultQuery("time_range", "week")
	granularity := c.DefaultQuery("granularity", "day") // hour, day, week, month

	var analytics gin.H
	var err error

	switch analyticsType {
	case "users":
		analytics = h.getUserAnalytics(timeRange, granularity)
	case "content":
		analytics = h.getContentAnalytics(timeRange, granularity)
	case "engagement":
		analytics = h.getEngagementAnalytics(timeRange, granularity)
	case "financial":
		analytics = h.getFinancialAnalytics(timeRange, granularity)
	case "geographical":
		analytics = h.getGeographicalAnalytics(timeRange)
	case "technical":
		analytics = h.getTechnicalAnalytics(timeRange, granularity)
	default:
		analytics = h.getOverviewAnalytics(timeRange, granularity)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get analytics", err)
		return
	}

	response := gin.H{
		"analytics":    analytics,
		"type":         analyticsType,
		"time_range":   timeRange,
		"granularity":  granularity,
		"generated_at": time.Now(),
	}

	utils.OkResponse(c, "Analytics retrieved successfully", response)
}

// ExportData exports data for reporting
func (h *AdminHandler) ExportData(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	var req struct {
		DataType   string    `json:"data_type" binding:"required"`
		Format     string    `json:"format" binding:"required"`
		TimeRange  string    `json:"time_range"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
		Filters    gin.H     `json:"filters"`
		IncludeRaw bool      `json:"include_raw"`
		Email      string    `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if !h.isValidExportType(req.DataType) {
		utils.BadRequestResponse(c, "Invalid data type", nil)
		return
	}

	if !h.isValidExportFormat(req.Format) {
		utils.BadRequestResponse(c, "Invalid export format", nil)
		return
	}

	// Create export job
	exportID := primitive.NewObjectID().Hex()
	h.createExportJob(exportID, req, c.GetString("user_id"))

	response := gin.H{
		"export_id":      exportID,
		"status":         "queued",
		"estimated_time": h.getEstimatedExportTime(req.DataType),
		"download_url":   "/api/v1/admin/exports/" + exportID + "/download",
		"created_at":     time.Now(),
	}

	utils.OkResponse(c, "Export job created successfully", response)
}

// ==================== SYSTEM MANAGEMENT ====================

// GetSystemConfig retrieves system configuration
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	config := gin.H{
		"general": gin.H{
			"app_name":             "Social Media Platform",
			"app_version":          "1.0.0",
			"maintenance_mode":     false,
			"registration_open":    true,
			"require_email_verify": true,
			"allow_guest_access":   false,
		},
		"content": gin.H{
			"max_post_length":     2000,
			"max_comment_length":  500,
			"max_story_duration":  30,
			"max_media_size_mb":   50,
			"allowed_media_types": []string{"jpg", "jpeg", "png", "gif", "mp4", "mov"},
			"auto_moderation":     true,
		},
		"user": gin.H{
			"max_follows_per_day":  100,
			"max_posts_per_day":    50,
			"max_comments_per_day": 200,
			"account_verification": true,
			"two_factor_auth":      true,
		},
		"notifications": gin.H{
			"email_enabled":    true,
			"push_enabled":     true,
			"sms_enabled":      false,
			"digest_frequency": "daily",
			"batch_size":       100,
		},
		"privacy": gin.H{
			"data_retention_days": 365,
			"cookie_consent":      true,
			"gdpr_compliance":     true,
			"data_export_enabled": true,
		},
		"security": gin.H{
			"session_timeout":         3600,
			"max_login_attempts":      5,
			"password_min_length":     8,
			"require_strong_password": true,
			"rate_limiting":           true,
		},
	}

	utils.OkResponse(c, "System configuration retrieved successfully", config)
}

// UpdateSystemConfig updates system configuration
func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate configuration changes
	if err := h.validateConfigChanges(req); err != nil {
		utils.BadRequestResponse(c, err.Error(), err)
		return
	}

	// Apply configuration changes
	h.applyConfigChanges(req, c.GetString("user_id"))

	response := gin.H{
		"updated_config":   req,
		"updated_at":       time.Now(),
		"updated_by":       c.GetString("user_id"),
		"requires_restart": h.requiresRestart(req),
	}

	utils.OkResponse(c, "System configuration updated successfully", response)
}

// GetSystemLogs retrieves system logs
func (h *AdminHandler) GetSystemLogs(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	logType := c.DefaultQuery("log_type", "all")
	level := c.DefaultQuery("level", "all")
	timeRange := c.DefaultQuery("time_range", "hour")

	params := utils.GetPaginationParams(c)

	logs := h.getSystemLogs(logType, level, timeRange, params.Limit, params.Offset)

	response := gin.H{
		"logs":       logs,
		"log_type":   logType,
		"level":      level,
		"time_range": timeRange,
		"summary":    h.getLogsSummary(timeRange),
	}

	paginationMeta := utils.CreatePaginationMeta(params, int64(len(logs)))
	utils.PaginatedSuccessResponse(c, "System logs retrieved successfully", response, paginationMeta, nil)
}

// ==================== ACTIVITY MONITORING ====================

// GetActivityLogs retrieves activity logs
func (h *AdminHandler) GetActivityLogs(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	activityType := c.Query("activity_type")
	userID := c.Query("user_id")
	timeRange := c.DefaultQuery("time_range", "day")

	params := utils.GetPaginationParams(c)

	activities := h.getActivityLogs(activityType, userID, timeRange, params.Limit, params.Offset)

	response := gin.H{
		"activities":    activities,
		"activity_type": activityType,
		"user_id":       userID,
		"time_range":    timeRange,
		"summary":       h.getActivitySummary(timeRange),
	}

	paginationMeta := utils.CreatePaginationMeta(params, int64(len(activities)))
	utils.PaginatedSuccessResponse(c, "Activity logs retrieved successfully", response, paginationMeta, nil)
}

// GetAuditTrail retrieves audit trail
func (h *AdminHandler) GetAuditTrail(c *gin.Context) {
	if !h.isAdmin(c) {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	action := c.Query("action")
	performedBy := c.Query("performed_by")
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	timeRange := c.DefaultQuery("time_range", "week")

	params := utils.GetPaginationParams(c)

	auditLogs := h.getAuditTrail(action, performedBy, targetType, targetID, timeRange, params.Limit, params.Offset)

	response := gin.H{
		"audit_logs": auditLogs,
		"filters": gin.H{
			"action":       action,
			"performed_by": performedBy,
			"target_type":  targetType,
			"target_id":    targetID,
			"time_range":   timeRange,
		},
		"summary": h.getAuditSummary(timeRange),
	}

	paginationMeta := utils.CreatePaginationMeta(params, int64(len(auditLogs)))
	utils.PaginatedSuccessResponse(c, "Audit trail retrieved successfully", response, paginationMeta, nil)
}

// ==================== HELPER METHODS ====================

// Authentication and authorization helpers
func (h *AdminHandler) isAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(models.UserRole)
	if !ok {
		return false
	}

	return role == models.RoleAdmin || role == models.RoleSuperAdmin
}

func (h *AdminHandler) isModerator(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(models.UserRole)
	if !ok {
		return false
	}

	return role == models.RoleModerator || role == models.RoleAdmin || role == models.RoleSuperAdmin
}

// Statistics helper methods
func (h *AdminHandler) getOverviewStats() gin.H {
	return gin.H{
		"total_users":         h.getTotalUsersCount(),
		"active_users":        h.getActiveUsersCount(),
		"total_posts":         h.getTotalPostsCount(),
		"total_comments":      h.getTotalCommentsCount(),
		"total_stories":       h.getTotalStoriesCount(),
		"total_groups":        h.getTotalGroupsCount(),
		"total_conversations": h.getTotalConversationsCount(),
		"total_reports":       h.getTotalReportsCount(),
		"storage_used":        h.getStorageUsed(),
		"bandwidth_used":      h.getBandwidthUsed(),
	}
}

func (h *AdminHandler) getUserStats(timeRange string) gin.H {
	return gin.H{
		"new_registrations": h.getNewRegistrationsCount(timeRange),
		"active_users":      h.getActiveUsersCount(),
		"verified_users":    h.getVerifiedUsersCount(),
		"suspended_users":   h.getSuspendedUsersCount(),
		"deleted_users":     h.getDeletedUsersCount(),
		"top_active_users":  h.getTopActiveUsers(10),
		"user_growth_rate":  h.getUserGrowthRate(timeRange),
		"retention_rate":    h.getUserRetentionRate(timeRange),
	}
}

func (h *AdminHandler) getContentStats(timeRange string) gin.H {
	return gin.H{
		"new_posts":        h.getNewPostsCount(timeRange),
		"new_comments":     h.getNewCommentsCount(timeRange),
		"new_stories":      h.getNewStoriesCount(timeRange),
		"deleted_content":  h.getDeletedContentCount(timeRange),
		"reported_content": h.getReportedContentCount(timeRange),
		"top_content":      h.getTopContent(10),
		"content_by_type":  h.getContentByType(timeRange),
		"viral_content":    h.getViralContent(timeRange),
	}
}

func (h *AdminHandler) getEngagementStats(timeRange string) gin.H {
	return gin.H{
		"total_likes":          h.getTotalLikesCount(timeRange),
		"total_comments":       h.getTotalCommentsCount(),
		"total_shares":         h.getTotalSharesCount(timeRange),
		"total_views":          h.getTotalViewsCount(timeRange),
		"engagement_rate":      h.getEngagementRate(timeRange),
		"avg_session_duration": h.getAvgSessionDuration(timeRange),
		"bounce_rate":          h.getBounceRate(timeRange),
		"top_hashtags":         h.getTopHashtags(10),
	}
}

func (h *AdminHandler) getModerationStats(timeRange string) gin.H {
	return gin.H{
		"new_reports":         h.getNewReportsCount(timeRange),
		"resolved_reports":    h.getResolvedReportsCount(timeRange),
		"pending_reports":     h.getPendingReportsCount(),
		"content_removed":     h.getContentRemovedCount(timeRange),
		"users_suspended":     h.getUsersSuspendedCount(timeRange),
		"false_reports":       h.getFalseReportsCount(timeRange),
		"avg_resolution_time": h.getAvgResolutionTime(timeRange),
		"moderator_activity":  h.getModeratorActivity(timeRange),
	}
}

func (h *AdminHandler) getSystemStats() gin.H {
	return gin.H{
		"server_uptime":     h.getUptime(),
		"memory_usage":      h.getMemoryUsage(),
		"cpu_usage":         h.getCPUUsage(),
		"disk_usage":        h.getDiskUsage(),
		"database_size":     h.getDatabaseSize(),
		"cache_hit_rate":    h.getCacheHitRate(),
		"api_response_time": h.getAPIResponseTime(),
		"error_rate":        h.getErrorRate(),
	}
}

func (h *AdminHandler) getGrowthStats(timeRange string) gin.H {
	return gin.H{
		"user_growth":       h.getUserGrowthData(timeRange),
		"content_growth":    h.getContentGrowthData(timeRange),
		"engagement_growth": h.getEngagementGrowthData(timeRange),
		"revenue_growth":    h.getRevenueGrowthData(timeRange),
		"geographic_growth": h.getGeographicGrowthData(timeRange),
	}
}

func (h *AdminHandler) getPerformanceStats() gin.H {
	return gin.H{
		"avg_response_time":    h.getAPIResponseTime(),
		"requests_per_second":  h.getRequestsPerSecond(),
		"error_rate":           h.getErrorRate(),
		"cache_performance":    h.getCachePerformance(),
		"database_performance": h.getDatabasePerformance(),
		"cdn_performance":      h.getCDNPerformance(),
	}
}

// Data retrieval helper methods (these would integrate with actual services)
func (h *AdminHandler) getTotalUsersCount() int64 {
	// This would call the actual user service
	return 10000 // Placeholder
}

func (h *AdminHandler) getActiveUsersCount() int64 {
	// This would call the actual user service
	return 2500 // Placeholder
}

func (h *AdminHandler) getTotalPostsCount() int64 {
	// This would call the actual post service
	return 50000 // Placeholder
}

func (h *AdminHandler) getTotalCommentsCount() int64 {
	// This would call the actual comment service
	return 150000 // Placeholder
}

func (h *AdminHandler) getTotalStoriesCount() int64 {
	// This would call the actual story service
	return 25000 // Placeholder
}

func (h *AdminHandler) getTotalGroupsCount() int64 {
	// This would call the actual group service
	return 500 // Placeholder
}

func (h *AdminHandler) getTotalConversationsCount() int64 {
	// This would call the actual conversation service
	return 75000 // Placeholder
}

func (h *AdminHandler) getTotalReportsCount() int64 {
	// This would call the actual report service
	return 1250 // Placeholder
}

// Additional helper methods would be implemented similarly...
// For brevity, I'm showing the structure but not implementing every single method

// These methods would interface with the actual services:
func (h *AdminHandler) getStorageUsed() string                        { return "125.5 GB" }
func (h *AdminHandler) getBandwidthUsed() string                      { return "2.3 TB" }
func (h *AdminHandler) getUptime() string                             { return "99.95%" }
func (h *AdminHandler) getMemoryUsage() string                        { return "68%" }
func (h *AdminHandler) getCPUUsage() string                           { return "45%" }
func (h *AdminHandler) getDiskUsage() string                          { return "72%" }
func (h *AdminHandler) getActiveSessionsCount() int64                 { return 1875 }
func (h *AdminHandler) getAPIResponseTime() string                    { return "120ms" }
func (h *AdminHandler) getOnlineUsersCount() int64                    { return 892 }
func (h *AdminHandler) getRecentPostsCount(minutes int) int64         { return 45 }
func (h *AdminHandler) getRecentCommentsCount(minutes int) int64      { return 128 }
func (h *AdminHandler) getRecentRegistrationsCount(minutes int) int64 { return 12 }
func (h *AdminHandler) getActiveConversationsCount() int64            { return 234 }
func (h *AdminHandler) getPendingReportsCount() int64                 { return 23 }
func (h *AdminHandler) getSystemAlerts() []gin.H                      { return []gin.H{} }
func (h *AdminHandler) getAPIRequestsCount(minutes int) int64         { return 5680 }

// Health check methods
func (h *AdminHandler) getDatabaseHealth() gin.H {
	return gin.H{
		"status":        "healthy",
		"connections":   45,
		"response_time": "5ms",
		"last_backup":   "2024-01-15T02:00:00Z",
	}
}

func (h *AdminHandler) getCacheHealth() gin.H {
	return gin.H{
		"status":       "healthy",
		"hit_rate":     "94%",
		"memory_usage": "68%",
		"connections":  12,
	}
}

func (h *AdminHandler) getStorageHealth() gin.H {
	return gin.H{
		"status":     "healthy",
		"disk_usage": "72%",
		"free_space": "340 GB",
		"iops":       "2.5K",
	}
}

func (h *AdminHandler) getEmailHealth() gin.H {
	return gin.H{
		"status":      "healthy",
		"queue_size":  23,
		"sent_today":  1247,
		"bounce_rate": "2.1%",
	}
}

func (h *AdminHandler) getPushHealth() gin.H {
	return gin.H{
		"status":        "healthy",
		"active_tokens": 8945,
		"sent_today":    5674,
		"delivery_rate": "97.8%",
	}
}

func (h *AdminHandler) getSearchHealth() gin.H {
	return gin.H{
		"status":       "healthy",
		"index_size":   "1.2 GB",
		"query_time":   "23ms",
		"indexed_docs": 125000,
	}
}

// Validation methods
func (h *AdminHandler) isValidBulkAction(action string) bool {
	validActions := []string{"suspend", "unsuspend", "verify", "unverify", "delete", "warn", "reset_password"}
	for _, a := range validActions {
		if action == a {
			return true
		}
	}
	return false
}

func (h *AdminHandler) isValidContentType(contentType string) bool {
	validTypes := []string{"post", "comment", "story", "group", "message"}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func (h *AdminHandler) isValidContentAction(action string) bool {
	validActions := []string{"delete", "hide", "feature", "unfeature", "approve", "reject"}
	for _, a := range validActions {
		if action == a {
			return true
		}
	}
	return false
}

func (h *AdminHandler) isValidExportType(dataType string) bool {
	validTypes := []string{"users", "posts", "comments", "reports", "analytics"}
	for _, t := range validTypes {
		if dataType == t {
			return true
		}
	}
	return false
}

func (h *AdminHandler) isValidExportFormat(format string) bool {
	validFormats := []string{"csv", "json", "xlsx"}
	for _, f := range validFormats {
		if format == f {
			return true
		}
	}
	return false
}

// Placeholder methods for complex operations that would integrate with actual services
func (h *AdminHandler) buildUserFilters(c *gin.Context) models.UserFilter {
	return models.UserFilter{
		Status:   c.Query("status"),
		Role:     c.Query("role"),
		Verified: c.Query("verified") == "true",
		CreatedAfter: func() time.Time {
			if t := c.Query("created_after"); t != "" {
				if parsed, err := time.Parse(time.RFC3339, t); err == nil {
					return parsed
				}
			}
			return time.Time{}
		}(),
		// Add more filters as needed
	}
}

// These methods would implement the actual business logic
func (h *AdminHandler) getUsersWithFilters(filters models.UserFilter, limit, offset int) ([]models.User, int64, error) {
	// Implementation would use actual services
	return []models.User{}, 0, nil
}

func (h *AdminHandler) userToAdminResponse(user models.User) gin.H {
	return gin.H{
		"id":              user.ID.Hex(),
		"username":        user.Username,
		"email":           user.Email,
		"display_name":    user.DisplayName,
		"status":          user.Status,
		"role":            user.Role,
		"verified":        user.IsVerified,
		"created_at":      user.CreatedAt,
		"last_login":      user.LastLoginAt,
		"posts_count":     user.PostsCount,
		"followers_count": user.FollowersCount,
		"following_count": user.FollowingCount,
	}
}

func (h *AdminHandler) getActiveUsersInResults(users []models.User) int64 {
	count := int64(0)
	for _, user := range users {
		if user.Status == "active" {
			count++
		}
	}
	return count
}

func (h *AdminHandler) getSuspendedUsersInResults(users []models.User) int64 {
	count := int64(0)
	for _, user := range users {
		if user.Status == "suspended" {
			count++
		}
	}
	return count
}

func (h *AdminHandler) getVerifiedUsersInResults(users []models.User) int64 {
	count := int64(0)
	for _, user := range users {
		if user.IsVerified {
			count++
		}
	}
	return count
}

// More helper methods would be implemented here...
// The key is that this provides a comprehensive admin interface
// that can be consumed by a frontend admin panel

func (h *AdminHandler) getUserStatistics(userID primitive.ObjectID) gin.H {
	return gin.H{
		"posts_count":     100,
		"comments_count":  250,
		"likes_given":     1500,
		"likes_received":  850,
		"followers_count": 120,
		"following_count": 180,
		"stories_count":   45,
		"groups_joined":   12,
		"reports_made":    2,
		"reports_against": 0,
	}
}

func (h *AdminHandler) getUserActivity(userID primitive.ObjectID) gin.H {
	return gin.H{
		"last_login":           "2024-01-15T10:30:00Z",
		"login_frequency":      "daily",
		"avg_session_duration": "45 minutes",
		"posts_per_week":       7,
		"comments_per_week":    25,
		"most_active_hours":    []int{9, 13, 18, 21},
		"device_usage":         gin.H{"mobile": 70, "desktop": 30},
	}
}

func (h *AdminHandler) getUserContent(userID primitive.ObjectID) gin.H {
	return gin.H{
		"recent_posts":    []gin.H{},
		"top_posts":       []gin.H{},
		"recent_comments": []gin.H{},
		"content_types":   gin.H{"text": 60, "image": 30, "video": 10},
	}
}

func (h *AdminHandler) getUserRelationships(userID primitive.ObjectID) gin.H {
	return gin.H{
		"mutual_connections":  45,
		"connection_strength": "high",
		"influence_score":     78,
		"network_reach":       5000,
	}
}

func (h *AdminHandler) getUserModerationHistory(userID primitive.ObjectID) gin.H {
	return gin.H{
		"warnings":         0,
		"suspensions":      0,
		"content_removed":  0,
		"reports_filed":    2,
		"reports_resolved": 2,
		"last_moderation":  nil,
	}
}

func (h *AdminHandler) getUserSessions(userID primitive.ObjectID) []gin.H {
	return []gin.H{
		{
			"session_id":  "session_123",
			"device":      "iPhone 13",
			"location":    "New York, NY",
			"ip_address":  "192.168.1.100",
			"last_active": "2024-01-15T10:30:00Z",
			"duration":    "2h 15m",
		},
	}
}

func (h *AdminHandler) getUserDevices(userID primitive.ObjectID) []gin.H {
	return []gin.H{
		{
			"device_type": "mobile",
			"device_name": "iPhone 13",
			"os":          "iOS 15.2",
			"browser":     "Safari",
			"first_seen":  "2024-01-01T00:00:00Z",
			"last_seen":   "2024-01-15T10:30:00Z",
			"trusted":     true,
		},
	}
}

func (h *AdminHandler) getUserPreferences(userID primitive.ObjectID) gin.H {
	return gin.H{
		"privacy_level":       "medium",
		"notifications":       gin.H{"email": true, "push": true, "sms": false},
		"content_preferences": gin.H{"mature_content": false, "personalization": true},
		"language":            "en",
		"timezone":            "America/New_York",
	}
}

// Placeholder methods for data operations
func (h *AdminHandler) performBulkUserAction(userIDs []string, action, reason, duration, note, performedBy string, sendEmail bool) gin.H {
	successful := 0
	failed := 0
	details := []gin.H{}

	for _, userID := range userIDs {
		// Simulate processing each user
		successful++
		details = append(details, gin.H{
			"user_id": userID,
			"status":  "success",
			"action":  action,
		})
	}

	return gin.H{
		"successful": successful,
		"failed":     failed,
		"details":    details,
	}
}

// More methods would be implemented following the same pattern...
// This provides a solid foundation for a comprehensive admin system
