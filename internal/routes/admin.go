// internal/routes/admin_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	// Admin routes group with authentication and admin role middleware
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired()) // Custom middleware to check admin role

	// ==================== DASHBOARD & OVERVIEW ====================

	// Dashboard statistics
	admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)
	admin.GET("/dashboard/real-time", adminHandler.GetRealTimeStats)

	// System health monitoring
	admin.GET("/system/health", adminHandler.GetSystemHealth)
	admin.GET("/system/status", adminHandler.GetSystemHealth) // Alias

	// ==================== USER MANAGEMENT ====================

	// User listing and management
	admin.GET("/users", adminHandler.GetUsers)
	admin.GET("/users/:userId", adminHandler.GetUserDetails)
	admin.GET("/users/:userId/activity", adminHandler.GetUserActivity)
	admin.POST("/users/bulk-action", adminHandler.BulkUserAction)

	// User statistics and analytics
	admin.GET("/users/stats", adminHandler.GetUsers)               // With stats=true query param
	admin.GET("/users/:userId/stats", adminHandler.GetUserDetails) // Stats section

	// ==================== CONTENT MANAGEMENT ====================

	// Content listing and management
	admin.GET("/content", adminHandler.GetContent)
	admin.GET("/content/:contentType/:contentId", adminHandler.GetContentDetails)
	admin.POST("/content/bulk-action", adminHandler.BulkContentAction)

	// Content statistics
	admin.GET("/content/stats", adminHandler.GetContent) // With stats=true query param

	// ==================== MODERATION ====================

	// Reports management (requires moderator role)
	moderator := admin.Group("/moderation")
	moderator.Use(middleware.ModeratorRequired()) // Custom middleware for moderator+ roles

	moderator.GET("/reports", adminHandler.GetReports)
	moderator.GET("/reports/:reportId", adminHandler.GetReports)           // Single report
	moderator.POST("/reports/bulk-action", adminHandler.BulkContentAction) // Reuse bulk action

	// Moderation queue
	moderator.GET("/queue", adminHandler.GetModerationQueue)
	moderator.GET("/queue/stats", adminHandler.GetModerationStats)

	// ==================== ANALYTICS & REPORTING ====================

	// Analytics endpoints
	admin.GET("/analytics", adminHandler.GetAnalytics)
	admin.GET("/analytics/users", func(c *gin.Context) {
		c.Set("analytics_type", "users")
		adminHandler.GetAnalytics(c)
	})
	admin.GET("/analytics/content", func(c *gin.Context) {
		c.Set("analytics_type", "content")
		adminHandler.GetAnalytics(c)
	})
	admin.GET("/analytics/engagement", func(c *gin.Context) {
		c.Set("analytics_type", "engagement")
		adminHandler.GetAnalytics(c)
	})
	admin.GET("/analytics/financial", func(c *gin.Context) {
		c.Set("analytics_type", "financial")
		adminHandler.GetAnalytics(c)
	})
	admin.GET("/analytics/geographical", func(c *gin.Context) {
		c.Set("analytics_type", "geographical")
		adminHandler.GetAnalytics(c)
	})
	admin.GET("/analytics/technical", func(c *gin.Context) {
		c.Set("analytics_type", "technical")
		adminHandler.GetAnalytics(c)
	})

	// Data export
	admin.POST("/export", adminHandler.ExportData)
	admin.GET("/exports/:exportId/status", func(c *gin.Context) {
		// Implementation for checking export status
		c.JSON(200, gin.H{"status": "processing"})
	})
	admin.GET("/exports/:exportId/download", func(c *gin.Context) {
		// Implementation for downloading export file
		c.JSON(200, gin.H{"download_url": "https://example.com/download/export.csv"})
	})

	// ==================== SYSTEM MANAGEMENT ====================

	// System configuration
	admin.GET("/system/config", adminHandler.GetSystemConfig)
	admin.PUT("/system/config", adminHandler.UpdateSystemConfig)

	// System logs
	admin.GET("/system/logs", adminHandler.GetSystemLogs)
	admin.GET("/system/logs/download", func(c *gin.Context) {
		// Implementation for downloading log files
		c.JSON(200, gin.H{"download_url": "https://example.com/download/logs.zip"})
	})

	// System operations
	admin.POST("/system/cache/clear", func(c *gin.Context) {
		// Implementation for clearing cache
		c.JSON(200, gin.H{"message": "Cache cleared successfully"})
	})
	admin.POST("/system/maintenance", func(c *gin.Context) {
		// Implementation for maintenance mode toggle
		c.JSON(200, gin.H{"maintenance_mode": true})
	})

	// ==================== ACTIVITY MONITORING ====================

	// Activity logs
	admin.GET("/activity", adminHandler.GetActivityLogs)
	admin.GET("/activity/audit", adminHandler.GetAuditTrail)

	// Security monitoring
	admin.GET("/security/sessions", func(c *gin.Context) {
		// Implementation for security session monitoring
		c.JSON(200, gin.H{"active_sessions": 1000})
	})
	admin.GET("/security/alerts", func(c *gin.Context) {
		// Implementation for security alerts
		c.JSON(200, gin.H{"alerts": []gin.H{}})
	})

	// ==================== SPECIALIZED ENDPOINTS ====================

	// User-specific admin actions
	userAdmin := admin.Group("/users/:userId")
	userAdmin.POST("/suspend", func(c *gin.Context) {
		// Implementation for user suspension
		c.JSON(200, gin.H{"message": "User suspended successfully"})
	})
	userAdmin.DELETE("/suspend", func(c *gin.Context) {
		// Implementation for user unsuspension
		c.JSON(200, gin.H{"message": "User unsuspended successfully"})
	})
	userAdmin.POST("/verify", func(c *gin.Context) {
		// Implementation for user verification
		c.JSON(200, gin.H{"message": "User verified successfully"})
	})
	userAdmin.DELETE("/verify", func(c *gin.Context) {
		// Implementation for removing user verification
		c.JSON(200, gin.H{"message": "User verification removed successfully"})
	})
	userAdmin.POST("/warn", func(c *gin.Context) {
		// Implementation for user warning
		c.JSON(200, gin.H{"message": "User warned successfully"})
	})
	userAdmin.DELETE("", func(c *gin.Context) {
		// Implementation for user deletion
		c.JSON(200, gin.H{"message": "User deleted successfully"})
	})

	// Content-specific admin actions
	contentAdmin := admin.Group("/content/:contentType/:contentId")
	contentAdmin.POST("/feature", func(c *gin.Context) {
		// Implementation for featuring content
		c.JSON(200, gin.H{"message": "Content featured successfully"})
	})
	contentAdmin.DELETE("/feature", func(c *gin.Context) {
		// Implementation for unfeaturing content
		c.JSON(200, gin.H{"message": "Content unfeatured successfully"})
	})
	contentAdmin.POST("/hide", func(c *gin.Context) {
		// Implementation for hiding content
		c.JSON(200, gin.H{"message": "Content hidden successfully"})
	})
	contentAdmin.DELETE("/hide", func(c *gin.Context) {
		// Implementation for unhiding content
		c.JSON(200, gin.H{"message": "Content unhidden successfully"})
	})
	contentAdmin.DELETE("", func(c *gin.Context) {
		// Implementation for content deletion
		c.JSON(200, gin.H{"message": "Content deleted successfully"})
	})

	// ==================== MONITORING ENDPOINTS ====================

	// Real-time monitoring
	admin.GET("/monitor/users/online", func(c *gin.Context) {
		// Implementation for online users monitoring
		c.JSON(200, gin.H{"online_users": 250})
	})
	admin.GET("/monitor/system/performance", func(c *gin.Context) {
		// Implementation for system performance monitoring
		c.JSON(200, gin.H{
			"cpu_usage":    45.2,
			"memory_usage": 68.5,
			"disk_usage":   72.1,
		})
	})
	admin.GET("/monitor/api/requests", func(c *gin.Context) {
		// Implementation for API request monitoring
		c.JSON(200, gin.H{"requests_per_minute": 1200})
	})

	// Database monitoring
	admin.GET("/monitor/database", func(c *gin.Context) {
		// Implementation for database monitoring
		c.JSON(200, gin.H{
			"connections":        45,
			"queries_per_second": 150,
			"avg_query_time":     "12ms",
		})
	})

	// ==================== UTILITY ENDPOINTS ====================

	// Search and filtering utilities
	admin.GET("/search/users", func(c *gin.Context) {
		// Quick user search for admin panel
		adminHandler.GetUsers(c)
	})
	admin.GET("/search/content", func(c *gin.Context) {
		// Quick content search for admin panel
		adminHandler.GetContent(c)
	})

	// Bulk operations
	admin.POST("/bulk/users", adminHandler.BulkUserAction)
	admin.POST("/bulk/content", adminHandler.BulkContentAction)

	// Data management
	admin.POST("/data/cleanup", func(c *gin.Context) {
		// Implementation for data cleanup
		c.JSON(200, gin.H{"message": "Data cleanup initiated"})
	})
	admin.POST("/data/backup", func(c *gin.Context) {
		// Implementation for data backup
		c.JSON(200, gin.H{"message": "Backup initiated"})
	})

	// ==================== NOTIFICATION MANAGEMENT ====================

	// Admin notifications
	admin.GET("/notifications", func(c *gin.Context) {
		// Implementation for admin notifications
		c.JSON(200, gin.H{"notifications": []gin.H{}})
	})
	admin.POST("/notifications/broadcast", func(c *gin.Context) {
		// Implementation for broadcasting notifications
		c.JSON(200, gin.H{"message": "Notification broadcasted successfully"})
	})

	// ==================== SETTINGS & CONFIGURATION ====================

	// Feature flags
	admin.GET("/features", func(c *gin.Context) {
		// Implementation for feature flags
		c.JSON(200, gin.H{
			"features": gin.H{
				"new_ui":          true,
				"advanced_search": true,
				"video_calling":   false,
				"ai_moderation":   true,
			},
		})
	})
	admin.PUT("/features", func(c *gin.Context) {
		// Implementation for updating feature flags
		c.JSON(200, gin.H{"message": "Feature flags updated successfully"})
	})

	// API rate limiting
	admin.GET("/rate-limits", func(c *gin.Context) {
		// Implementation for rate limit configuration
		c.JSON(200, gin.H{
			"rate_limits": gin.H{
				"posts_per_hour":    10,
				"comments_per_hour": 60,
				"follows_per_hour":  50,
			},
		})
	})
	admin.PUT("/rate-limits", func(c *gin.Context) {
		// Implementation for updating rate limits
		c.JSON(200, gin.H{"message": "Rate limits updated successfully"})
	})

	// ==================== INTEGRATION ENDPOINTS ====================

	// Third-party integrations
	admin.GET("/integrations", func(c *gin.Context) {
		// Implementation for integration status
		c.JSON(200, gin.H{
			"integrations": gin.H{
				"email_service":   "active",
				"push_service":    "active",
				"storage_service": "active",
				"cdn_service":     "active",
			},
		})
	})

	// Webhook management
	admin.GET("/webhooks", func(c *gin.Context) {
		// Implementation for webhook management
		c.JSON(200, gin.H{"webhooks": []gin.H{}})
	})
	admin.POST("/webhooks", func(c *gin.Context) {
		// Implementation for creating webhooks
		c.JSON(200, gin.H{"message": "Webhook created successfully"})
	})

	// ==================== TESTING & DEVELOPMENT ====================

	// Testing endpoints (only in development)
	if gin.Mode() == gin.DebugMode {
		test := admin.Group("/test")
		test.POST("/email", func(c *gin.Context) {
			// Implementation for testing email
			c.JSON(200, gin.H{"message": "Test email sent"})
		})
		test.POST("/push", func(c *gin.Context) {
			// Implementation for testing push notifications
			c.JSON(200, gin.H{"message": "Test push notification sent"})
		})
		test.POST("/sms", func(c *gin.Context) {
			// Implementation for testing SMS
			c.JSON(200, gin.H{"message": "Test SMS sent"})
		})
		test.GET("/performance", func(c *gin.Context) {
			// Implementation for performance testing
			c.JSON(200, gin.H{"response_time": "15ms"})
		})
	}
}

// Helper function to setup super admin routes (most restrictive)
func SetupSuperAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	superAdmin := router.Group("/api/v1/super-admin")
	superAdmin.Use(middleware.AuthRequired())
	superAdmin.Use(middleware.SuperAdminRequired()) // Custom middleware for super admin only

	// Super admin only operations
	superAdmin.POST("/users/:userId/delete-permanently", func(c *gin.Context) {
		// Implementation for permanent user deletion
		c.JSON(200, gin.H{"message": "User permanently deleted"})
	})

	superAdmin.POST("/system/reset", func(c *gin.Context) {
		// Implementation for system reset (very dangerous)
		c.JSON(200, gin.H{"message": "System reset initiated"})
	})

	superAdmin.GET("/system/secrets", func(c *gin.Context) {
		// Implementation for viewing system secrets/config
		c.JSON(200, gin.H{"secrets": "encrypted_data"})
	})

	superAdmin.POST("/admin/create", func(c *gin.Context) {
		// Implementation for creating new admin users
		c.JSON(200, gin.H{"message": "Admin user created"})
	})

	superAdmin.DELETE("/admin/:userId", func(c *gin.Context) {
		// Implementation for removing admin privileges
		c.JSON(200, gin.H{"message": "Admin privileges revoked"})
	})
}

// Helper function to setup moderator routes
func SetupModeratorRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	moderator := router.Group("/api/v1/moderator")
	moderator.Use(middleware.AuthRequired())
	moderator.Use(middleware.ModeratorRequired())

	// Moderator-specific operations
	moderator.GET("/queue", adminHandler.GetModerationQueue)
	moderator.GET("/reports", adminHandler.GetReports)
	moderator.POST("/reports/:reportId/resolve", func(c *gin.Context) {
		// Implementation for resolving reports
		c.JSON(200, gin.H{"message": "Report resolved"})
	})
	moderator.POST("/reports/:reportId/reject", func(c *gin.Context) {
		// Implementation for rejecting reports
		c.JSON(200, gin.H{"message": "Report rejected"})
	})

	// Content moderation
	moderator.POST("/content/:contentType/:contentId/approve", func(c *gin.Context) {
		// Implementation for approving content
		c.JSON(200, gin.H{"message": "Content approved"})
	})
	moderator.POST("/content/:contentType/:contentId/remove", func(c *gin.Context) {
		// Implementation for removing content
		c.JSON(200, gin.H{"message": "Content removed"})
	})

	// User moderation
	moderator.POST("/users/:userId/warn", func(c *gin.Context) {
		// Implementation for warning users
		c.JSON(200, gin.H{"message": "User warned"})
	})
	moderator.POST("/users/:userId/timeout", func(c *gin.Context) {
		// Implementation for timing out users
		c.JSON(200, gin.H{"message": "User timed out"})
	})
}
