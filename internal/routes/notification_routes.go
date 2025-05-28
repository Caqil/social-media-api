// internal/routes/notification_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupNotificationRoutes sets up notification-related routes
func SetupNotificationRoutes(router *gin.Engine, notificationHandler *handlers.NotificationHandler, authMiddleware *middleware.AuthMiddleware) {
	// All notification routes require authentication
	notifications := router.Group("/api/v1/notifications")
	notifications.Use(authMiddleware.RequireAuth())
	{
		// Notification management
		notifications.GET("/", notificationHandler.GetNotifications)
		notifications.GET("/stats", notificationHandler.GetNotificationStats)
		notifications.POST("/mark-read", notificationHandler.MarkAsRead)
		notifications.POST("/mark-all-read", notificationHandler.MarkAllAsRead)
		notifications.DELETE("/", notificationHandler.DeleteNotifications)

		// Notification preferences
		notifications.GET("/preferences", notificationHandler.GetNotificationPreferences)
		notifications.PUT("/preferences", notificationHandler.UpdateNotificationPreferences)

		// Specific notification triggers (usually called internally but exposed for testing)
		notifications.POST("/like", notificationHandler.NotifyLike)
		notifications.POST("/comment", notificationHandler.NotifyComment)
		notifications.POST("/follow", notificationHandler.NotifyFollow)
		notifications.POST("/mention", notificationHandler.NotifyMention)
		notifications.POST("/message", notificationHandler.NotifyMessage)

		// Admin/system notification creation
		notifications.POST("/create", middleware.RequireAdmin(), notificationHandler.CreateNotification)
		notifications.POST("/bulk-create", middleware.RequireAdmin(), notificationHandler.CreateBulkNotifications)

		// Test notification (development/admin use)
		notifications.POST("/test", notificationHandler.TestNotification)
	}
}
