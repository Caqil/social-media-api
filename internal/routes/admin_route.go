// internal/routes/admin_routes.go
package routes

import (
	"log"
	"net/http"

	"social-media-api/internal/config"
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Simple admin middleware - just check if user has admin role
func requireAdminRole() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get user role from context (set by regular auth middleware)
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// Simple super admin middleware
func requireSuperAdminRole() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Super admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

func SetupAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	admin := router.Group("/api/v1/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(middleware.RequireAdmin()) // Same auth that works for users
	admin.Use(requireAdminRole())        // Simple role check
	admin.Use(middleware.Logger())       // Request logging

	// Dashboard
	admin.GET("/dashboard", adminHandler.GetDashboard)
	admin.GET("/dashboard/stats", adminHandler.GetDashboard)

	// User Management
	users := admin.Group("/users")
	{
		users.GET("", adminHandler.GetAllUsers)
		users.GET("/search", adminHandler.SearchUsers)
		users.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetUser)
		users.GET("/:id/stats", middleware.ValidateObjectID("id"), adminHandler.GetUserStats)
		users.PUT("/:id/status", middleware.ValidateObjectID("id"), adminHandler.UpdateUserStatus)
		users.PUT("/:id/verify", middleware.ValidateObjectID("id"), adminHandler.VerifyUser)
		users.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteUser)

		// Bulk operations
		users.POST("/bulk/actions", adminHandler.BulkUserAction)

		// Export
		users.GET("/export", adminHandler.ExportUsers)
	}

	// Post Management
	posts := admin.Group("/posts")
	{
		posts.GET("", adminHandler.GetAllPosts)
		posts.GET("/search", adminHandler.SearchPosts)
		posts.GET("/:id/stats", middleware.ValidateObjectID("id"), adminHandler.GetPostStats)
		posts.PUT("/:id/hide", middleware.ValidateObjectID("id"), adminHandler.HidePost)
		posts.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeletePost)

		// Bulk operations
		posts.POST("/bulk/actions", adminHandler.BulkPostAction)

		// Export
		posts.GET("/export", adminHandler.ExportPosts)
	}

	// Comment Management
	comments := admin.Group("/comments")
	{
		comments.GET("", adminHandler.GetAllComments)
		comments.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetComment)
		comments.PUT("/:id/hide", middleware.ValidateObjectID("id"), adminHandler.HideComment)
		comments.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteComment)
		comments.POST("/bulk/actions", adminHandler.BulkCommentAction)
		
		comments.PUT("/:id", middleware.ValidateObjectID("id"), adminHandler.UpdateComment)
		comments.PUT("/:id/show", middleware.ValidateObjectID("id"), adminHandler.ShowComment)
	}

	// Group Management
	groups := admin.Group("/groups")
	{
		groups.GET("", adminHandler.GetAllGroups)
		groups.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetGroup)
		groups.GET("/:id/members", middleware.ValidateObjectID("id"), adminHandler.GetGroupMembers)
		groups.PUT("/:id/status", middleware.ValidateObjectID("id"), adminHandler.UpdateGroupStatus)
		groups.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteGroup)
		groups.POST("/bulk/actions", adminHandler.BulkGroupAction)
	}

	// Event Management
	events := admin.Group("/events")
	{
		events.GET("", adminHandler.GetAllEvents)
		events.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetEvent)
		events.GET("/:id/attendees", middleware.ValidateObjectID("id"), adminHandler.GetEventAttendees)
		events.PUT("/:id/status", middleware.ValidateObjectID("id"), adminHandler.UpdateEventStatus)
		events.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteEvent)
		events.POST("/bulk/actions", adminHandler.BulkEventAction)
	}

	// Story Management
	stories := admin.Group("/stories")
	{
		stories.GET("", adminHandler.GetAllStories)
		stories.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetStory)
		stories.PUT("/:id/hide", middleware.ValidateObjectID("id"), adminHandler.HideStory)
		stories.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteStory)
		stories.POST("/bulk/actions", adminHandler.BulkStoryAction)
	}

	// Message Management
	messages := admin.Group("/messages")
	{
		messages.GET("", adminHandler.GetAllMessages)
		messages.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetMessage)
		messages.GET("/conversations", adminHandler.GetAllConversations)
		messages.GET("/conversations/:id", middleware.ValidateObjectID("id"), adminHandler.GetConversation)
		messages.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteMessage)
		messages.POST("/bulk/actions", adminHandler.BulkMessageAction)
	}

	// Report Management
	reports := admin.Group("/reports")
	{
		reports.GET("", adminHandler.GetAllReports)
		reports.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetReport)
		reports.PUT("/:id/status", middleware.ValidateObjectID("id"), adminHandler.UpdateReportStatus)
		reports.PUT("/:id/assign", middleware.ValidateObjectID("id"), adminHandler.AssignReport)
		reports.POST("/:id/resolve", middleware.ValidateObjectID("id"), adminHandler.ResolveReport)
		reports.POST("/:id/reject", middleware.ValidateObjectID("id"), adminHandler.RejectReport)
		reports.POST("/bulk/actions", adminHandler.BulkReportAction)

		// Report statistics
		reports.GET("/stats", adminHandler.GetReportStats)
		reports.GET("/stats/summary", adminHandler.GetReportSummary)
	}

	// Follow/Relationship Management
	follows := admin.Group("/follows")
	{
		follows.GET("", adminHandler.GetAllFollows)
		follows.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetFollow)
		follows.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteFollow)
		follows.GET("/relationships", adminHandler.GetRelationships)
		follows.POST("/bulk/actions", adminHandler.BulkFollowAction)
	}

	// Like/Reaction Management
	likes := admin.Group("/likes")
	{
		likes.GET("", adminHandler.GetAllLikes)
		likes.GET("/stats", adminHandler.GetLikeStats)
		likes.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteLike)
		likes.POST("/bulk/actions", adminHandler.BulkLikeAction)
	}

	// Hashtag Management
	hashtags := admin.Group("/hashtags")
	{
		hashtags.GET("", adminHandler.GetAllHashtags)
		hashtags.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetHashtag)
		hashtags.GET("/trending", adminHandler.GetTrendingHashtags)
		hashtags.PUT("/:id/block", middleware.ValidateObjectID("id"), adminHandler.BlockHashtag)
		hashtags.PUT("/:id/unblock", middleware.ValidateObjectID("id"), adminHandler.UnblockHashtag)
		hashtags.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteHashtag)
		hashtags.POST("/bulk/actions", adminHandler.BulkHashtagAction)
	}

	// Mention Management
	mentions := admin.Group("/mentions")
	{
		mentions.GET("", adminHandler.GetAllMentions)
		mentions.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetMention)
		mentions.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteMention)
		mentions.POST("/bulk/actions", adminHandler.BulkMentionAction)
	}

	// Media Management
	media := admin.Group("/media")
	{
		media.GET("", adminHandler.GetAllMedia)
		media.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetMedia)
		media.GET("/stats", adminHandler.GetMediaStats)
		media.PUT("/:id/moderate", middleware.ValidateObjectID("id"), adminHandler.ModerateMedia)
		media.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteMedia)
		media.POST("/bulk/actions", adminHandler.BulkMediaAction)

		// Storage management
		media.GET("/storage/stats", adminHandler.GetStorageStats)
		media.POST("/storage/cleanup", adminHandler.CleanupStorage)
	}

	// Notification Management
	notifications := admin.Group("/notifications")
	{
		notifications.GET("", adminHandler.GetAllNotifications)
		notifications.GET("/:id", middleware.ValidateObjectID("id"), adminHandler.GetNotification)
		notifications.POST("/send", adminHandler.SendNotificationToUsers)
		notifications.POST("/broadcast", adminHandler.BroadcastNotification)
		notifications.GET("/stats", adminHandler.GetNotificationStats)
		notifications.DELETE("/:id", middleware.ValidateObjectID("id"), adminHandler.DeleteNotification)
		notifications.POST("/bulk/actions", adminHandler.BulkNotificationAction)
	}

	// Analytics
	analytics := admin.Group("/analytics")
	{
		analytics.GET("/users", adminHandler.GetUserAnalytics)
		analytics.GET("/content", adminHandler.GetContentAnalytics)
		analytics.GET("/engagement", adminHandler.GetEngagementAnalytics)
		analytics.GET("/growth", adminHandler.GetGrowthAnalytics)
		analytics.GET("/demographics", adminHandler.GetDemographicAnalytics)
		analytics.GET("/revenue", adminHandler.GetRevenueAnalytics)
		analytics.GET("/reports/custom", adminHandler.GetCustomReport)

		// Real-time analytics
		analytics.GET("/realtime", adminHandler.GetRealtimeAnalytics)
		analytics.GET("/live-stats", adminHandler.GetLiveStats)
	}

	// System Management (Super Admin only)
	system := admin.Group("/system")
	system.Use(requireSuperAdminRole()) // Additional check for super admin
	{
		system.GET("/health", adminHandler.GetSystemHealth)
		system.GET("/info", adminHandler.GetSystemInfo)
		system.GET("/logs", adminHandler.GetSystemLogs)
		system.GET("/performance", adminHandler.GetPerformanceMetrics)
		system.GET("/database/stats", adminHandler.GetDatabaseStats)
		system.GET("/cache/stats", adminHandler.GetCacheStats)

		// System operations
		system.POST("/cache/clear", adminHandler.ClearCache)
		system.POST("/cache/warm", adminHandler.WarmCache)
		system.POST("/maintenance/enable", adminHandler.EnableMaintenanceMode)
		system.POST("/maintenance/disable", adminHandler.DisableMaintenanceMode)

		// Database operations
		system.POST("/database/backup", adminHandler.BackupDatabase)
		system.GET("/database/backups", adminHandler.GetDatabaseBackups)
		system.POST("/database/restore", adminHandler.RestoreDatabase)
		system.POST("/database/optimize", adminHandler.OptimizeDatabase)
	}

	// Configuration Management (Super Admin only)
	config := admin.Group("/config")
	config.Use(requireSuperAdminRole())
	{
		config.GET("", adminHandler.GetConfiguration)
		config.PUT("", adminHandler.UpdateConfiguration)
		config.GET("/history", adminHandler.GetConfigurationHistory)
		config.POST("/rollback", adminHandler.RollbackConfiguration)
		config.GET("/validate", adminHandler.ValidateConfiguration)

		// Feature flags
		config.GET("/features", adminHandler.GetFeatureFlags)
		config.PUT("/features", adminHandler.UpdateFeatureFlags)
		config.PUT("/features/:feature/toggle", adminHandler.ToggleFeature)

		// Rate limits
		config.GET("/rate-limits", adminHandler.GetRateLimits)
		config.PUT("/rate-limits", adminHandler.UpdateRateLimits)
	}
}

// Public admin routes (no authentication required)
func SetupPublicAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	public := router.Group("/api/v1/admin/public")
	public.Use(middleware.CORS())
	public.GET("/status", adminHandler.GetPublicSystemStatus)
	public.GET("/health", adminHandler.GetPublicHealthCheck)

	// Authentication routes
	auth := public.Group("/auth")

	cfg := config.GetConfig()
	if cfg.IsDevelopment() {
		log.Println("⚠️  ALL RATE LIMITING DISABLED FOR DEVELOPMENT")
		log.Println("🔑 Login endpoint: POST /api/v1/admin/public/auth/login")
		log.Println("🔄 Refresh endpoint: POST /api/v1/admin/public/auth/refresh")
		// NO MIDDLEWARE APPLIED IN DEVELOPMENT
	} else {
		// Only apply rate limiting in production
		//auth.Use(middleware.LoginRateLimit())
		log.Println("🛡️  Rate limiting ENABLED for production")
	}
	{
		auth.POST("/login", adminHandler.AdminLogin)
		auth.POST("/logout", adminHandler.AdminLogout)
		auth.POST("/refresh", adminHandler.RefreshAdminToken)
		auth.POST("/forgot-password", adminHandler.AdminForgotPassword)
		auth.POST("/reset-password", adminHandler.AdminResetPassword)
	}
}

// WebSocket routes for real-time admin features
func SetupAdminWebSocketRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, db *mongo.Database, jwtSecret, refreshSecret string) {
	authMiddleware := middleware.NewAuthMiddleware(db, jwtSecret, refreshSecret)

	ws := router.Group("/api/v1/admin/ws")
	ws.Use(authMiddleware.RequireAuth())
	ws.Use(requireAdminRole())

	// Real-time dashboard updates
	ws.GET("/dashboard", adminHandler.DashboardWebSocket)

	// Real-time system monitoring
	ws.GET("/monitoring", adminHandler.MonitoringWebSocket)

	// Real-time moderation queue
	ws.GET("/moderation", adminHandler.ModerationWebSocket)

	// Real-time user activities
	ws.GET("/activities", adminHandler.ActivitiesWebSocket)
}

// SetupSuperAdminRoutes sets up routes that require super admin access
func SetupSuperAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, db *mongo.Database, jwtSecret, refreshSecret string) {
	authMiddleware := middleware.NewAuthMiddleware(db, jwtSecret, refreshSecret)

	superAdmin := router.Group("/api/v1/super-admin")
	superAdmin.Use(authMiddleware.RequireAuth())
	superAdmin.Use(requireSuperAdminRole())

	// Additional super admin only routes can be added here
	// These would be for extremely sensitive operations
}
