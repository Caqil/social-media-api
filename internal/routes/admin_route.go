// internal/routes/admin_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	// Admin routes group with authentication and authorization middleware
	admin := router.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware())    // Verify JWT token
	admin.Use(middleware.AdminMiddleware())   // Verify admin role
	admin.Use(middleware.RateLimit())         // Rate limiting
	admin.Use(middleware.LoggingMiddleware()) // Request logging

	// Dashboard
	admin.GET("/dashboard", adminHandler.GetDashboard)
	admin.GET("/dashboard/stats", adminHandler.GetDashboard)

	// User Management
	users := admin.Group("/users")
	{
		users.GET("", adminHandler.GetAllUsers)
		users.GET("/search", adminHandler.SearchUsers)
		users.GET("/:id", adminHandler.GetUser)
		users.GET("/:id/stats", adminHandler.GetUserStats)
		users.PUT("/:id/status", adminHandler.UpdateUserStatus)
		users.PUT("/:id/verify", adminHandler.VerifyUser)
		users.DELETE("/:id", adminHandler.DeleteUser)

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
		posts.GET("/:id/stats", adminHandler.GetPostStats)
		posts.PUT("/:id/hide", adminHandler.HidePost)
		posts.DELETE("/:id", adminHandler.DeletePost)

		// Bulk operations
		posts.POST("/bulk/actions", adminHandler.BulkPostAction)

		// Export
		posts.GET("/export", adminHandler.ExportPosts)
	}

	// Comment Management
	comments := admin.Group("/comments")
	{
		comments.GET("", adminHandler.GetAllComments)
		comments.GET("/:id", adminHandler.GetComment)
		comments.PUT("/:id/hide", adminHandler.HideComment)
		comments.DELETE("/:id", adminHandler.DeleteComment)
		comments.POST("/bulk/actions", adminHandler.BulkCommentAction)
	}

	// Group Management
	groups := admin.Group("/groups")
	{
		groups.GET("", adminHandler.GetAllGroups)
		groups.GET("/:id", adminHandler.GetGroup)
		groups.GET("/:id/members", adminHandler.GetGroupMembers)
		groups.PUT("/:id/status", adminHandler.UpdateGroupStatus)
		groups.DELETE("/:id", adminHandler.DeleteGroup)
		groups.POST("/bulk/actions", adminHandler.BulkGroupAction)
	}

	// Event Management
	events := admin.Group("/events")
	{
		events.GET("", adminHandler.GetAllEvents)
		events.GET("/:id", adminHandler.GetEvent)
		events.GET("/:id/attendees", adminHandler.GetEventAttendees)
		events.PUT("/:id/status", adminHandler.UpdateEventStatus)
		events.DELETE("/:id", adminHandler.DeleteEvent)
		events.POST("/bulk/actions", adminHandler.BulkEventAction)
	}

	// Story Management
	stories := admin.Group("/stories")
	{
		stories.GET("", adminHandler.GetAllStories)
		stories.GET("/:id", adminHandler.GetStory)
		stories.PUT("/:id/hide", adminHandler.HideStory)
		stories.DELETE("/:id", adminHandler.DeleteStory)
		stories.POST("/bulk/actions", adminHandler.BulkStoryAction)
	}

	// Message Management
	messages := admin.Group("/messages")
	{
		messages.GET("", adminHandler.GetAllMessages)
		messages.GET("/:id", adminHandler.GetMessage)
		messages.GET("/conversations", adminHandler.GetAllConversations)
		messages.GET("/conversations/:id", adminHandler.GetConversation)
		messages.DELETE("/:id", adminHandler.DeleteMessage)
		messages.POST("/bulk/actions", adminHandler.BulkMessageAction)
	}

	// Report Management
	reports := admin.Group("/reports")
	{
		reports.GET("", adminHandler.GetAllReports)
		reports.GET("/:id", adminHandler.GetReport)
		reports.PUT("/:id/status", adminHandler.UpdateReportStatus)
		reports.PUT("/:id/assign", adminHandler.AssignReport)
		reports.POST("/:id/resolve", adminHandler.ResolveReport)
		reports.POST("/:id/reject", adminHandler.RejectReport)
		reports.POST("/bulk/actions", adminHandler.BulkReportAction)

		// Report statistics
		reports.GET("/stats", adminHandler.GetReportStats)
		reports.GET("/stats/summary", adminHandler.GetReportSummary)
	}

	// Follow/Relationship Management
	follows := admin.Group("/follows")
	{
		follows.GET("", adminHandler.GetAllFollows)
		follows.GET("/:id", adminHandler.GetFollow)
		follows.DELETE("/:id", adminHandler.DeleteFollow)
		follows.GET("/relationships", adminHandler.GetRelationships)
		follows.POST("/bulk/actions", adminHandler.BulkFollowAction)
	}

	// Like/Reaction Management
	likes := admin.Group("/likes")
	{
		likes.GET("", adminHandler.GetAllLikes)
		likes.GET("/stats", adminHandler.GetLikeStats)
		likes.DELETE("/:id", adminHandler.DeleteLike)
		likes.POST("/bulk/actions", adminHandler.BulkLikeAction)
	}

	// Hashtag Management
	hashtags := admin.Group("/hashtags")
	{
		hashtags.GET("", adminHandler.GetAllHashtags)
		hashtags.GET("/:id", adminHandler.GetHashtag)
		hashtags.GET("/trending", adminHandler.GetTrendingHashtags)
		hashtags.PUT("/:id/block", adminHandler.BlockHashtag)
		hashtags.PUT("/:id/unblock", adminHandler.UnblockHashtag)
		hashtags.DELETE("/:id", adminHandler.DeleteHashtag)
		hashtags.POST("/bulk/actions", adminHandler.BulkHashtagAction)
	}

	// Mention Management
	mentions := admin.Group("/mentions")
	{
		mentions.GET("", adminHandler.GetAllMentions)
		mentions.GET("/:id", adminHandler.GetMention)
		mentions.DELETE("/:id", adminHandler.DeleteMention)
		mentions.POST("/bulk/actions", adminHandler.BulkMentionAction)
	}

	// Media Management
	media := admin.Group("/media")
	{
		media.GET("", adminHandler.GetAllMedia)
		media.GET("/:id", adminHandler.GetMedia)
		media.GET("/stats", adminHandler.GetMediaStats)
		media.PUT("/:id/moderate", adminHandler.ModerateMedia)
		media.DELETE("/:id", adminHandler.DeleteMedia)
		media.POST("/bulk/actions", adminHandler.BulkMediaAction)

		// Storage management
		media.GET("/storage/stats", adminHandler.GetStorageStats)
		media.POST("/storage/cleanup", adminHandler.CleanupStorage)
	}

	// Notification Management
	notifications := admin.Group("/notifications")
	{
		notifications.GET("", adminHandler.GetAllNotifications)
		notifications.GET("/:id", adminHandler.GetNotification)
		notifications.POST("/send", adminHandler.SendNotificationToUsers)
		notifications.POST("/broadcast", adminHandler.BroadcastNotification)
		notifications.GET("/stats", adminHandler.GetNotificationStats)
		notifications.DELETE("/:id", adminHandler.DeleteNotification)
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

	// System Management
	system := admin.Group("/system")
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

	// Configuration Management
	config := admin.Group("/config")
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

	// // Security Management
	// security := admin.Group("/security")
	// {
	// 	security.GET("/blocked-ips", adminHandler.GetBlockedIPs)
	// 	security.POST("/blocked-ips", adminHandler.BlockIP)
	// 	security.DELETE("/blocked-ips/:ip", adminHandler.UnblockIP)

	// 	security.GET("/suspicious-activities", adminHandler.GetSuspiciousActivities)
	// 	security.GET("/login-attempts", adminHandler.GetLoginAttempts)
	// 	security.GET("/security-events", adminHandler.GetSecurityEvents)

	// 	// API keys and tokens
	// 	security.GET("/api-keys", adminHandler.GetAPIKeys)
	// 	security.POST("/api-keys", adminHandler.CreateAPIKey)
	// 	security.DELETE("/api-keys/:id", adminHandler.DeleteAPIKey)
	// 	security.PUT("/api-keys/:id/regenerate", adminHandler.RegenerateAPIKey)

	// 	// Two-factor authentication
	// 	security.GET("/2fa/stats", adminHandler.Get2FAStats)
	// 	security.POST("/2fa/reset", adminHandler.Reset2FA)
	// }

	// // Admin User Management
	// admins := admin.Group("/admins")
	// {
	// 	admins.GET("", adminHandler.GetAllAdmins)
	// 	admins.GET("/:id", adminHandler.GetAdmin)
	// 	admins.POST("", adminHandler.CreateAdmin)
	// 	admins.PUT("/:id", adminHandler.UpdateAdmin)
	// 	admins.PUT("/:id/role", adminHandler.UpdateAdminRole)
	// 	admins.PUT("/:id/permissions", adminHandler.UpdateAdminPermissions)
	// 	admins.DELETE("/:id", adminHandler.DeleteAdmin)

	// 	// Admin activities
	// 	admins.GET("/activities", adminHandler.GetAdminActivities)
	// 	admins.GET("/:id/activities", adminHandler.GetAdminActivitiesByAdmin)
	// }

	// // Audit and Logging
	// audit := admin.Group("/audit")
	// {
	// 	audit.GET("/logs", adminHandler.GetAuditLogs)
	// 	audit.GET("/logs/:id", adminHandler.GetAuditLog)
	// 	audit.GET("/trail", adminHandler.GetAuditTrail)
	// 	audit.POST("/logs/export", adminHandler.ExportAuditLogs)

	// 	// User activities
	// 	audit.GET("/users/:id/activities", adminHandler.GetUserActivities)
	// 	audit.GET("/content/:type/:id/history", adminHandler.GetContentHistory)
	// }

	// // Moderation Tools
	// moderation := admin.Group("/moderation")
	// {
	// 	moderation.GET("/queue", adminHandler.GetModerationQueue)
	// 	moderation.GET("/queue/priority", adminHandler.GetPriorityModerationQueue)
	// 	moderation.POST("/content/:id/approve", adminHandler.ApproveContent)
	// 	moderation.POST("/content/:id/reject", adminHandler.RejectContent)
	// 	moderation.POST("/content/:id/flag", adminHandler.FlagContent)

	// 	// AI moderation
	// 	moderation.GET("/ai/stats", adminHandler.GetAIModerationStats)
	// 	moderation.POST("/ai/retrain", adminHandler.RetrainAIModeration)
	// 	moderation.GET("/ai/confidence", adminHandler.GetAIModerationConfidence)

	// 	// Moderation rules
	// 	moderation.GET("/rules", adminHandler.GetModerationRules)
	// 	moderation.POST("/rules", adminHandler.CreateModerationRule)
	// 	moderation.PUT("/rules/:id", adminHandler.UpdateModerationRule)
	// 	moderation.DELETE("/rules/:id", adminHandler.DeleteModerationRule)
	// }

	// // Communication Tools
	// communication := admin.Group("/communication")
	// {
	// 	// Email management
	// 	communication.GET("/emails", adminHandler.GetEmailCampaigns)
	// 	communication.POST("/emails/send", adminHandler.SendEmail)
	// 	communication.POST("/emails/broadcast", adminHandler.BroadcastEmail)
	// 	communication.GET("/emails/templates", adminHandler.GetEmailTemplates)
	// 	communication.POST("/emails/templates", adminHandler.CreateEmailTemplate)

	// 	// Push notifications
	// 	communication.GET("/push/stats", adminHandler.GetPushNotificationStats)
	// 	communication.POST("/push/send", adminHandler.SendPushNotification)
	// 	communication.POST("/push/broadcast", adminHandler.BroadcastPushNotification)

	// 	// Announcements
	// 	communication.GET("/announcements", adminHandler.GetAnnouncements)
	// 	communication.POST("/announcements", adminHandler.CreateAnnouncement)
	// 	communication.PUT("/announcements/:id", adminHandler.UpdateAnnouncement)
	// 	communication.DELETE("/announcements/:id", adminHandler.DeleteAnnouncement)
	// }

	// // Integration Management
	// integrations := admin.Group("/integrations")
	// {
	// 	integrations.GET("", adminHandler.GetIntegrations)
	// 	integrations.GET("/:id", adminHandler.GetIntegration)
	// 	integrations.POST("", adminHandler.CreateIntegration)
	// 	integrations.PUT("/:id", adminHandler.UpdateIntegration)
	// 	integrations.DELETE("/:id", adminHandler.DeleteIntegration)
	// 	integrations.POST("/:id/test", adminHandler.TestIntegration)

	// 	// Webhooks
	// 	integrations.GET("/webhooks", adminHandler.GetWebhooks)
	// 	integrations.POST("/webhooks", adminHandler.CreateWebhook)
	// 	integrations.PUT("/webhooks/:id", adminHandler.UpdateWebhook)
	// 	integrations.DELETE("/webhooks/:id", adminHandler.DeleteWebhook)
	// 	integrations.POST("/webhooks/:id/test", adminHandler.TestWebhook)
	// }

	// // Business Intelligence
	// bi := admin.Group("/business-intelligence")
	// {
	// 	bi.GET("/kpis", adminHandler.GetKPIs)
	// 	bi.GET("/metrics", adminHandler.GetBusinessMetrics)
	// 	bi.GET("/forecasts", adminHandler.GetForecasts)
	// 	bi.GET("/trends", adminHandler.GetTrends)
	// 	bi.GET("/cohorts", adminHandler.GetCohortAnalysis)
	// 	bi.GET("/funnel", adminHandler.GetFunnelAnalysis)
	// 	bi.GET("/retention", adminHandler.GetRetentionAnalysis)

	// 	// Custom dashboards
	// 	bi.GET("/dashboards", adminHandler.GetCustomDashboards)
	// 	bi.POST("/dashboards", adminHandler.CreateCustomDashboard)
	// 	bi.PUT("/dashboards/:id", adminHandler.UpdateCustomDashboard)
	// 	bi.DELETE("/dashboards/:id", adminHandler.DeleteCustomDashboard)
	// }

	// // Content Management
	// content := admin.Group("/content")
	// {
	// 	// Content categories
	// 	content.GET("/categories", adminHandler.GetContentCategories)
	// 	content.POST("/categories", adminHandler.CreateContentCategory)
	// 	content.PUT("/categories/:id", adminHandler.UpdateContentCategory)
	// 	content.DELETE("/categories/:id", adminHandler.DeleteContentCategory)

	// 	// Content moderation
	// 	content.GET("/moderation/pending", adminHandler.GetPendingContent)
	// 	content.POST("/moderation/batch", adminHandler.BatchModerateContent)
	// 	content.GET("/moderation/history", adminHandler.GetModerationHistory)

	// 	// Content performance
	// 	content.GET("/performance", adminHandler.GetContentPerformance)
	// 	content.GET("/viral", adminHandler.GetViralContent)
	// 	content.GET("/trending", adminHandler.getTrendingContent)
	// }

	// // Data Management
	// data := admin.Group("/data")
	// {
	// 	// Import/Export
	// 	data.POST("/import", adminHandler.ImportData)
	// 	data.GET("/export/status/:id", adminHandler.GetExportStatus)
	// 	data.GET("/export/download/:id", adminHandler.DownloadExport)

	// 	// Data quality
	// 	data.GET("/quality/report", adminHandler.GetDataQualityReport)
	// 	data.POST("/quality/fix", adminHandler.FixDataQuality)
	// 	data.GET("/duplicates", adminHandler.GetDuplicateData)
	// 	data.POST("/duplicates/merge", adminHandler.MergeDuplicateData)

	// 	// Data retention
	// 	data.GET("/retention/policies", adminHandler.GetDataRetentionPolicies)
	// 	data.POST("/retention/policies", adminHandler.CreateDataRetentionPolicy)
	// 	data.PUT("/retention/policies/:id", adminHandler.UpdateDataRetentionPolicy)
	// 	data.POST("/retention/execute", adminHandler.ExecuteDataRetention)
	// }

	// // API Management
	// api := admin.Group("/api-management")
	// {
	// 	api.GET("/endpoints", adminHandler.GetAPIEndpoints)
	// 	api.GET("/usage", adminHandler.GetAPIUsage)
	// 	api.GET("/performance", adminHandler.GetAPIPerformance)
	// 	api.GET("/errors", adminHandler.GetAPIErrors)
	// 	api.GET("/rate-limits/usage", adminHandler.GetRateLimitUsage)

	// 	// API documentation
	// 	api.GET("/documentation", adminHandler.GetAPIDocumentation)
	// 	api.PUT("/documentation", adminHandler.UpdateAPIDocumentation)

	// 	// API versions
	// 	api.GET("/versions", adminHandler.GetAPIVersions)
	// 	api.POST("/versions", adminHandler.CreateAPIVersion)
	// 	api.PUT("/versions/:version/deprecate", adminHandler.DeprecateAPIVersion)
	// }

	// // Testing and Development
	// dev := admin.Group("/development")
	// {
	// 	dev.GET("/sandbox", adminHandler.GetSandboxEnvironment)
	// 	dev.POST("/sandbox/reset", adminHandler.ResetSandboxEnvironment)
	// 	dev.GET("/test-data", adminHandler.GetTestData)
	// 	dev.POST("/test-data/generate", adminHandler.GenerateTestData)
	// 	dev.POST("/test-data/cleanup", adminHandler.CleanupTestData)

	// 	// Feature testing
	// 	dev.GET("/ab-tests", adminHandler.GetABTests)
	// 	dev.POST("/ab-tests", adminHandler.CreateABTest)
	// 	dev.PUT("/ab-tests/:id", adminHandler.UpdateABTest)
	// 	dev.POST("/ab-tests/:id/start", adminHandler.StartABTest)
	// 	dev.POST("/ab-tests/:id/stop", adminHandler.StopABTest)
	// 	dev.GET("/ab-tests/:id/results", adminHandler.GetABTestResults)
	// }

	// // Help and Documentation
	// help := admin.Group("/help")
	// {
	// 	help.GET("/docs", adminHandler.GetDocumentation)
	// 	help.GET("/faq", adminHandler.GetFAQ)
	// 	help.GET("/tutorials", adminHandler.GetTutorials)
	// 	help.GET("/support", adminHandler.GetSupportInfo)
	// 	help.POST("/feedback", adminHandler.SubmitFeedback)
	// }
}

// Additional middleware that might be needed
func SetupAdminMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(),
		middleware.AdminMiddleware(),
		middleware.RateLimitMiddleware(),
		middleware.LoggingMiddleware(),
		middleware.CORSMiddleware(),
		middleware.SecurityHeadersMiddleware(),
	}
}

// Public admin routes (no authentication required)
func SetupPublicAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	public := router.Group("/api/admin/public")

	// System status (for monitoring)
	public.GET("/status", adminHandler.GetPublicSystemStatus)
	public.GET("/health", adminHandler.GetPublicHealthCheck)

	// Login
	public.POST("/login", adminHandler.AdminLogin)
	public.POST("/logout", adminHandler.AdminLogout)
	public.POST("/refresh", adminHandler.RefreshAdminToken)

	// Password reset
	public.POST("/forgot-password", adminHandler.AdminForgotPassword)
	public.POST("/reset-password", adminHandler.AdminResetPassword)
}

// WebSocket routes for real-time admin features
func SetupAdminWebSocketRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	ws := router.Group("/api/admin/ws")
	ws.Use(middleware.AuthMiddleware())
	ws.Use(middleware.AdminMiddleware())
	ws.Use(middleware.WebSocketUpgradeMiddleware())

	// Real-time dashboard updates
	ws.GET("/dashboard", adminHandler.DashboardWebSocket)

	// Real-time system monitoring
	ws.GET("/monitoring", adminHandler.MonitoringWebSocket)

	// Real-time moderation queue
	ws.GET("/moderation", adminHandler.ModerationWebSocket)

	// Real-time user activities
	ws.GET("/activities", adminHandler.ActivitiesWebSocket)
}
