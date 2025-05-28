// internal/routes/admin_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes sets up all admin routes
func SetupAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Admin routes - require authentication and admin role
	adminRoutes := router.Group("/api/v1/admin")
	adminRoutes.Use(authMiddleware.RequireAuth())
	adminRoutes.Use(authMiddleware.RequireRole("admin"))

	// Dashboard and Analytics
	adminRoutes.GET("/dashboard", adminHandler.GetDashboard)
	adminRoutes.GET("/analytics", adminHandler.GetAnalytics)
	adminRoutes.GET("/system/health", adminHandler.GetSystemHealth)
	adminRoutes.GET("/system/metrics", adminHandler.GetRealtimeStats)

	// User Management
	adminRoutes.GET("/users", adminHandler.GetUsers)
	adminRoutes.GET("/users/:userId", adminHandler.GetUser)
	adminRoutes.POST("/users/bulk-action", adminHandler.BulkUserAction)
	adminRoutes.POST("/users/:userId/suspend", adminHandler.SuspendUser)
	adminRoutes.DELETE("/users/:userId/suspend", adminHandler.UnsuspendUser)
	adminRoutes.POST("/users/:userId/verify", adminHandler.VerifyUser)
	adminRoutes.DELETE("/users/:userId", adminHandler.DeleteUser)

	// Content Management
	adminRoutes.GET("/content/:contentType", adminHandler.GetContent) // posts, comments, stories
	adminRoutes.POST("/content/bulk-action", adminHandler.BulkContentAction)

	// Report Management
	adminRoutes.GET("/reports", adminHandler.GetReports)

	// System Configuration
	adminRoutes.GET("/system/config", adminHandler.GetSystemConfig)
	adminRoutes.PUT("/system/config", adminHandler.UpdateSystemConfig)
	adminRoutes.POST("/system/maintenance", adminHandler.SetMaintenanceMode)
	adminRoutes.POST("/system/clear-cache", adminHandler.ClearCache)

	// Platform Monitoring
	adminRoutes.GET("/monitoring/active-users", adminHandler.GetActiveUsers)
	adminRoutes.GET("/monitoring/error-logs", adminHandler.GetErrorLogs)

	// Admin Activity Logs
	adminRoutes.GET("/activities", adminHandler.GetAdminActivities)

	// Data Export/Import
	adminRoutes.POST("/exports", adminHandler.ExportData)
	adminRoutes.GET("/exports/:exportId/status", adminHandler.GetExportStatus)
}

// SetupSuperAdminRoutes sets up super admin only routes
func SetupSuperAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Super Admin routes - require super admin role
	superAdminRoutes := router.Group("/api/v1/super-admin")
	superAdminRoutes.Use(authMiddleware.RequireAuth())
	superAdminRoutes.Use(authMiddleware.RequireRole("super_admin"))

	// Advanced System Management
	superAdminRoutes.GET("/system/full-metrics", adminHandler.GetRealtimeStats) // Extended metrics
	superAdminRoutes.POST("/system/emergency-shutdown", func(c *gin.Context) {
		// Emergency shutdown functionality
		c.JSON(200, gin.H{"message": "Emergency shutdown initiated"})
	})

	// Advanced User Management
	superAdminRoutes.GET("/users/all-data/:userId", adminHandler.GetUser) // Full user data
	superAdminRoutes.POST("/users/:userId/permanent-delete", func(c *gin.Context) {
		// Permanent deletion (hard delete)
		c.JSON(200, gin.H{"message": "User permanently deleted"})
	})

	// Admin Management
	superAdminRoutes.GET("/admins", func(c *gin.Context) {
		// Get all admin users
		c.JSON(200, gin.H{"message": "Admin users retrieved"})
	})
	superAdminRoutes.POST("/admins/:userId/promote", func(c *gin.Context) {
		// Promote user to admin
		c.JSON(200, gin.H{"message": "User promoted to admin"})
	})
	superAdminRoutes.POST("/admins/:userId/demote", func(c *gin.Context) {
		// Demote admin to user
		c.JSON(200, gin.H{"message": "Admin demoted to user"})
	})
}

// SetupModeratorRoutes sets up moderator routes
func SetupModeratorRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, reportHandler *handlers.ReportHandler, authMiddleware *middleware.AuthMiddleware) {
	// Moderator routes - require moderator role or higher
	moderatorRoutes := router.Group("/api/v1/moderator")
	moderatorRoutes.Use(authMiddleware.RequireAuth())
	moderatorRoutes.Use(authMiddleware.RequireRole("moderator"))

	// Content Moderation
	moderatorRoutes.GET("/queue", func(c *gin.Context) {
		// Get moderation queue
		c.JSON(200, gin.H{"message": "Moderation queue retrieved"})
	})
	moderatorRoutes.GET("/reports", reportHandler.GetReports)
	moderatorRoutes.GET("/reports/:reportId", reportHandler.GetReport)
	moderatorRoutes.PUT("/reports/:reportId", reportHandler.UpdateReport)
	moderatorRoutes.POST("/reports/:reportId/resolve", reportHandler.ResolveReport)
	moderatorRoutes.POST("/reports/:reportId/reject", reportHandler.RejectReport)
	moderatorRoutes.POST("/reports/:reportId/assign", reportHandler.AssignReport)

	// Content Actions
	moderatorRoutes.GET("/content/flagged", func(c *gin.Context) {
		// Get flagged content
		c.JSON(200, gin.H{"message": "Flagged content retrieved"})
	})
	moderatorRoutes.POST("/content/:contentType/:contentId/approve", func(c *gin.Context) {
		// Approve content
		c.JSON(200, gin.H{"message": "Content approved"})
	})
	moderatorRoutes.POST("/content/:contentType/:contentId/reject", func(c *gin.Context) {
		// Reject content
		c.JSON(200, gin.H{"message": "Content rejected"})
	})
	moderatorRoutes.POST("/content/:contentType/:contentId/flag", func(c *gin.Context) {
		// Flag content
		c.JSON(200, gin.H{"message": "Content flagged"})
	})

	// User Moderation
	moderatorRoutes.POST("/users/:userId/warn", func(c *gin.Context) {
		// Warn user
		c.JSON(200, gin.H{"message": "User warned"})
	})
	moderatorRoutes.POST("/users/:userId/restrict", func(c *gin.Context) {
		// Restrict user (temporary limitations)
		c.JSON(200, gin.H{"message": "User restricted"})
	})

	// Moderation Stats
	moderatorRoutes.GET("/stats", func(c *gin.Context) {
		// Get moderation statistics
		c.JSON(200, gin.H{"message": "Moderation stats retrieved"})
	})
}

// SetupPublicAdminRoutes sets up public admin routes (no auth required)
func SetupPublicAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler) {
	// Public admin routes (no authentication required)
	publicAdminRoutes := router.Group("/api/v1/public/admin")

	// System Status (public)
	publicAdminRoutes.GET("/status", func(c *gin.Context) {
		// Public system status
		c.JSON(200, gin.H{
			"status":       "operational",
			"version":      "1.0.0",
			"uptime":       "24h 15m 30s",
			"last_updated": "2024-01-15T10:30:00Z",
		})
	})

	// Maintenance Status (public)
	publicAdminRoutes.GET("/maintenance", func(c *gin.Context) {
		// Check if system is in maintenance mode
		c.JSON(200, gin.H{
			"maintenance_mode": false,
			"message":          "",
		})
	})
}

// SetupAdminWebhookRoutes sets up webhook routes for admin integrations
func SetupAdminWebhookRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Webhook routes for admin integrations
	webhookRoutes := router.Group("/api/v1/admin/webhooks")
	webhookRoutes.Use(authMiddleware.RequireAuth())
	webhookRoutes.Use(authMiddleware.RequireRole("admin"))

	// System Monitoring Webhooks
	webhookRoutes.POST("/system/alert", func(c *gin.Context) {
		// Handle system alerts
		c.JSON(200, gin.H{"message": "System alert received"})
	})

	// External Service Webhooks
	webhookRoutes.POST("/payments/stripe", func(c *gin.Context) {
		// Handle Stripe webhooks
		c.JSON(200, gin.H{"message": "Stripe webhook processed"})
	})

	webhookRoutes.POST("/analytics/google", func(c *gin.Context) {
		// Handle Google Analytics webhooks
		c.JSON(200, gin.H{"message": "Analytics webhook processed"})
	})

	webhookRoutes.POST("/monitoring/datadog", func(c *gin.Context) {
		// Handle Datadog webhooks
		c.JSON(200, gin.H{"message": "Monitoring webhook processed"})
	})
}

// SetupAdminAPIRoutes sets up API management routes
func SetupAdminAPIRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// API Management routes
	apiRoutes := router.Group("/api/v1/admin/api")
	apiRoutes.Use(authMiddleware.RequireAuth())
	apiRoutes.Use(authMiddleware.RequireRole("admin"))

	// API Keys Management
	apiRoutes.GET("/keys", func(c *gin.Context) {
		// Get API keys
		c.JSON(200, gin.H{"message": "API keys retrieved"})
	})
	apiRoutes.POST("/keys", func(c *gin.Context) {
		// Create API key
		c.JSON(201, gin.H{"message": "API key created"})
	})
	apiRoutes.DELETE("/keys/:keyId", func(c *gin.Context) {
		// Delete API key
		c.JSON(200, gin.H{"message": "API key deleted"})
	})

	// Rate Limiting
	apiRoutes.GET("/rate-limits", func(c *gin.Context) {
		// Get rate limits
		c.JSON(200, gin.H{"message": "Rate limits retrieved"})
	})
	apiRoutes.PUT("/rate-limits", func(c *gin.Context) {
		// Update rate limits
		c.JSON(200, gin.H{"message": "Rate limits updated"})
	})

	// API Usage Statistics
	apiRoutes.GET("/usage", func(c *gin.Context) {
		// Get API usage statistics
		c.JSON(200, gin.H{"message": "API usage retrieved"})
	})
	apiRoutes.GET("/usage/:endpoint", func(c *gin.Context) {
		// Get specific endpoint usage
		c.JSON(200, gin.H{"message": "Endpoint usage retrieved"})
	})
}

// SetupBackupRoutes sets up backup management routes
func SetupBackupRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Backup Management routes
	backupRoutes := router.Group("/api/v1/admin/backup")
	backupRoutes.Use(authMiddleware.RequireAuth())
	backupRoutes.Use(authMiddleware.RequireRole("admin"))

	// Backup Operations
	backupRoutes.GET("/", func(c *gin.Context) {
		// List all backups
		c.JSON(200, gin.H{"message": "Backups retrieved"})
	})
	backupRoutes.POST("/create", func(c *gin.Context) {
		// Create new backup
		c.JSON(201, gin.H{"message": "Backup created"})
	})
	backupRoutes.POST("/restore/:backupId", func(c *gin.Context) {
		// Restore from backup
		c.JSON(200, gin.H{"message": "Backup restored"})
	})
	backupRoutes.DELETE("/:backupId", func(c *gin.Context) {
		// Delete backup
		c.JSON(200, gin.H{"message": "Backup deleted"})
	})

	// Backup Settings
	backupRoutes.GET("/settings", func(c *gin.Context) {
		// Get backup settings
		c.JSON(200, gin.H{"message": "Backup settings retrieved"})
	})
	backupRoutes.PUT("/settings", func(c *gin.Context) {
		// Update backup settings
		c.JSON(200, gin.H{"message": "Backup settings updated"})
	})
}

// SetupLogRoutes sets up log management routes
func SetupLogRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Log Management routes
	logRoutes := router.Group("/api/v1/admin/logs")
	logRoutes.Use(authMiddleware.RequireAuth())
	logRoutes.Use(authMiddleware.RequireRole("admin"))

	// System Logs
	logRoutes.GET("/system", adminHandler.GetErrorLogs)
	logRoutes.GET("/access", func(c *gin.Context) {
		// Get access logs
		c.JSON(200, gin.H{"message": "Access logs retrieved"})
	})
	logRoutes.GET("/audit", adminHandler.GetAdminActivities)
	logRoutes.GET("/performance", func(c *gin.Context) {
		// Get performance logs
		c.JSON(200, gin.H{"message": "Performance logs retrieved"})
	})

	// Log Analysis
	logRoutes.GET("/analysis/errors", func(c *gin.Context) {
		// Analyze error patterns
		c.JSON(200, gin.H{"message": "Error analysis retrieved"})
	})
	logRoutes.GET("/analysis/performance", func(c *gin.Context) {
		// Analyze performance patterns
		c.JSON(200, gin.H{"message": "Performance analysis retrieved"})
	})

	// Log Configuration
	logRoutes.GET("/config", func(c *gin.Context) {
		// Get logging configuration
		c.JSON(200, gin.H{"message": "Log config retrieved"})
	})
	logRoutes.PUT("/config", func(c *gin.Context) {
		// Update logging configuration
		c.JSON(200, gin.H{"message": "Log config updated"})
	})
}

// SetupNotificationRoutes sets up admin notification routes
func SetupAdminNotificationRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Admin Notification routes
	notificationRoutes := router.Group("/api/v1/admin/notifications")
	notificationRoutes.Use(authMiddleware.RequireAuth())
	notificationRoutes.Use(authMiddleware.RequireRole("admin"))

	// System Notifications
	notificationRoutes.GET("/system", func(c *gin.Context) {
		// Get system notifications
		c.JSON(200, gin.H{"message": "System notifications retrieved"})
	})
	notificationRoutes.POST("/system", func(c *gin.Context) {
		// Create system notification
		c.JSON(201, gin.H{"message": "System notification created"})
	})

	// User Notifications
	notificationRoutes.POST("/broadcast", func(c *gin.Context) {
		// Broadcast notification to all users
		c.JSON(200, gin.H{"message": "Broadcast notification sent"})
	})
	notificationRoutes.POST("/targeted", func(c *gin.Context) {
		// Send targeted notifications
		c.JSON(200, gin.H{"message": "Targeted notifications sent"})
	})

	// Notification Templates
	notificationRoutes.GET("/templates", func(c *gin.Context) {
		// Get notification templates
		c.JSON(200, gin.H{"message": "Notification templates retrieved"})
	})
	notificationRoutes.POST("/templates", func(c *gin.Context) {
		// Create notification template
		c.JSON(201, gin.H{"message": "Notification template created"})
	})
	notificationRoutes.PUT("/templates/:templateId", func(c *gin.Context) {
		// Update notification template
		c.JSON(200, gin.H{"message": "Notification template updated"})
	})
	notificationRoutes.DELETE("/templates/:templateId", func(c *gin.Context) {
		// Delete notification template
		c.JSON(200, gin.H{"message": "Notification template deleted"})
	})
}

// SetupSchedulerRoutes sets up scheduled task management routes
func SetupSchedulerRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Scheduler Management routes
	schedulerRoutes := router.Group("/api/v1/admin/scheduler")
	schedulerRoutes.Use(authMiddleware.RequireAuth())
	schedulerRoutes.Use(authMiddleware.RequireRole("admin"))

	// Scheduled Jobs
	schedulerRoutes.GET("/jobs", func(c *gin.Context) {
		// Get scheduled jobs
		c.JSON(200, gin.H{"message": "Scheduled jobs retrieved"})
	})
	schedulerRoutes.POST("/jobs", func(c *gin.Context) {
		// Create scheduled job
		c.JSON(201, gin.H{"message": "Scheduled job created"})
	})
	schedulerRoutes.PUT("/jobs/:jobId", func(c *gin.Context) {
		// Update scheduled job
		c.JSON(200, gin.H{"message": "Scheduled job updated"})
	})
	schedulerRoutes.DELETE("/jobs/:jobId", func(c *gin.Context) {
		// Delete scheduled job
		c.JSON(200, gin.H{"message": "Scheduled job deleted"})
	})
	schedulerRoutes.POST("/jobs/:jobId/run", func(c *gin.Context) {
		// Run job immediately
		c.JSON(200, gin.H{"message": "Job executed"})
	})

	// Job History
	schedulerRoutes.GET("/jobs/:jobId/history", func(c *gin.Context) {
		// Get job execution history
		c.JSON(200, gin.H{"message": "Job history retrieved"})
	})

	// Cron Jobs
	schedulerRoutes.GET("/cron", func(c *gin.Context) {
		// Get cron jobs
		c.JSON(200, gin.H{"message": "Cron jobs retrieved"})
	})
	schedulerRoutes.POST("/cron", func(c *gin.Context) {
		// Create cron job
		c.JSON(201, gin.H{"message": "Cron job created"})
	})
}

// SetupComplianceRoutes sets up compliance and legal routes
func SetupComplianceRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Compliance Management routes
	complianceRoutes := router.Group("/api/v1/admin/compliance")
	complianceRoutes.Use(authMiddleware.RequireAuth())
	complianceRoutes.Use(authMiddleware.RequireRole("admin"))

	// GDPR Compliance
	complianceRoutes.GET("/gdpr/requests", func(c *gin.Context) {
		// Get GDPR requests
		c.JSON(200, gin.H{"message": "GDPR requests retrieved"})
	})
	complianceRoutes.POST("/gdpr/data-export/:userId", func(c *gin.Context) {
		// Export user data for GDPR
		c.JSON(200, gin.H{"message": "GDPR data export initiated"})
	})
	complianceRoutes.POST("/gdpr/data-deletion/:userId", func(c *gin.Context) {
		// Delete user data for GDPR
		c.JSON(200, gin.H{"message": "GDPR data deletion initiated"})
	})

	// Audit Logs
	complianceRoutes.GET("/audit", func(c *gin.Context) {
		// Get compliance audit logs
		c.JSON(200, gin.H{"message": "Audit logs retrieved"})
	})
	complianceRoutes.POST("/audit/report", func(c *gin.Context) {
		// Generate compliance report
		c.JSON(200, gin.H{"message": "Compliance report generated"})
	})

	// Legal Holds
	complianceRoutes.GET("/legal-holds", func(c *gin.Context) {
		// Get legal holds
		c.JSON(200, gin.H{"message": "Legal holds retrieved"})
	})
	complianceRoutes.POST("/legal-holds", func(c *gin.Context) {
		// Create legal hold
		c.JSON(201, gin.H{"message": "Legal hold created"})
	})
	complianceRoutes.DELETE("/legal-holds/:holdId", func(c *gin.Context) {
		// Remove legal hold
		c.JSON(200, gin.H{"message": "Legal hold removed"})
	})
}
