// internal/routes/follow_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupFollowRoutes sets up follow/social relationship routes
func SetupFollowRoutes(router *gin.Engine, followHandler *handlers.FollowHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public follow routes (viewing relationships) - FIXED: Changed :userId to :id
	follows := router.Group("/api/v1")
	{
		// User relationship viewing
		follows.GET("/users/:id/followers", authMiddleware.OptionalAuth(), followHandler.GetFollowers)
		follows.GET("/users/:id/following", authMiddleware.OptionalAuth(), followHandler.GetFollowing)
		follows.GET("/users/:id/follow-stats", authMiddleware.OptionalAuth(), followHandler.GetFollowStats)
		follows.GET("/users/:id/mutual-follows", authMiddleware.RequireAuth(), followHandler.GetMutualFollows)
	}

	// Protected follow routes - FIXED: Changed :userId to :id
	followsProtected := router.Group("/api/v1")
	followsProtected.Use(authMiddleware.RequireAuth())
	{
		// Follow actions
		followsProtected.POST("/users/:id/follow", middleware.FollowRateLimit(), followHandler.FollowUser)
		followsProtected.DELETE("/users/:id/follow", followHandler.UnfollowUser)
		followsProtected.GET("/users/:id/follow-status", followHandler.CheckFollowStatus)

		// Follow management
		followsProtected.GET("/follow-requests", followHandler.GetFollowRequests)
		followsProtected.POST("/follow-requests/:followId/accept", followHandler.AcceptFollowRequest)
		followsProtected.POST("/follow-requests/:followId/reject", followHandler.RejectFollowRequest)
		followsProtected.DELETE("/follow-requests/:followId", followHandler.CancelFollowRequest)

		// Follower management - FIXED: Changed :userId to :id
		followsProtected.DELETE("/followers/:id", followHandler.RemoveFollower)

		// Follow discovery and suggestions
		followsProtected.GET("/suggested-users", followHandler.GetSuggestedUsers)
		followsProtected.POST("/bulk-follow", followHandler.BulkFollowUsers)

		// Follow activity
		followsProtected.GET("/follow-activity", followHandler.GetFollowActivity)
	}
}
