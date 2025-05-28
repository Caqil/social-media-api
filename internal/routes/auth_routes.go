// internal/routes/auth_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up authentication and user profile routes
func SetupAuthRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public auth routes (no authentication required)
	auth := router.Group("/api/v1/auth")
	{
		// Rate limiting for auth endpoints
		auth.Use(middleware.LoginRateLimit())

		// Authentication endpoints
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.GET("/verify-email", authHandler.VerifyEmail)
	}

	// Protected auth routes (require authentication)
	authProtected := router.Group("/api/v1/auth")
	authProtected.Use(authMiddleware.RequireAuth())
	{
		// Profile management
		authProtected.GET("/profile", authHandler.GetProfile)
		authProtected.PUT("/profile", authHandler.UpdateProfile)
		authProtected.POST("/change-password", authHandler.ChangePassword)

		// Session management
		authProtected.GET("/sessions", authHandler.GetSessions)
		authProtected.DELETE("/sessions/:sessionId", authHandler.RevokeSession)
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.POST("/logout-all", authHandler.LogoutAll)
	}
}
