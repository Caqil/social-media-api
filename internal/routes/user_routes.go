// internal/routes/user_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes sets up user-related routes
func SetupUserRoutes(router *gin.Engine, userHandler *handlers.UserHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public user routes
	users := router.Group("/api/v1/users")
	{
		// User discovery (public)
		users.GET("/search", userHandler.SearchUsers)
		users.GET("/:id", userHandler.GetUserProfile)
		users.GET("/username/:username", userHandler.GetUserByUsername)
		users.GET("/:id/stats", userHandler.GetUserStats)
	}

	// Protected user routes
	usersProtected := router.Group("/api/v1/users")
	usersProtected.Use(authMiddleware.RequireAuth())
	{
		// User suggestions and discovery
		usersProtected.GET("/suggestions", userHandler.GetSuggestedUsers)

		// Profile management
		usersProtected.PUT("/profile", userHandler.UpdateProfile)
		usersProtected.PUT("/privacy-settings", userHandler.UpdatePrivacySettings)
		usersProtected.PUT("/notification-settings", userHandler.UpdateNotificationSettings)
		usersProtected.PUT("/activity-status", userHandler.UpdateUserActivity)

		// Account management
		usersProtected.POST("/deactivate", userHandler.DeactivateAccount)

		// Blocking functionality
		usersProtected.POST("/:id/block", userHandler.BlockUser)
		usersProtected.DELETE("/:id/block", userHandler.UnblockUser)
		usersProtected.GET("/blocked", userHandler.GetBlockedUsers)
	}

	// Admin-only user routes
	usersAdmin := router.Group("/api/v1/users")
	usersAdmin.Use(authMiddleware.RequireAuth())
	usersAdmin.Use(middleware.RequireAdmin())
	{
		usersAdmin.POST("/:id/verify", userHandler.VerifyUser)
		usersAdmin.POST("/:id/suspend", userHandler.SuspendUser)
		usersAdmin.DELETE("/:id/suspend", userHandler.UnsuspendUser)
	}
}
