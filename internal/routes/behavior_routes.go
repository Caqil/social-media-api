// internal/routes/behavior_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupBehaviorRoutes sets up behavior tracking routes
func SetupBehaviorRoutes(router *gin.Engine, behaviorHandler *handlers.UserBehaviorHandler, authMiddleware middleware.AuthMiddleware, behaviorMiddleware *middleware.BehaviorTrackingMiddleware) {

	// Public behavior routes (minimal tracking)
	publicBehavior := router.Group("/api/v1/behavior")
	{
		// Anonymous session tracking (for non-authenticated users)
		publicBehavior.POST("/anonymous-session", behaviorHandler.StartSession)
	}

	// Protected behavior routes
	behaviorRoutes := router.Group("/api/v1/behavior")
	behaviorRoutes.Use(authMiddleware.RequireAuth())
	{
		// Session Management
		behaviorRoutes.POST("/sessions/start", behaviorHandler.StartSession)
		behaviorRoutes.POST("/sessions/end", behaviorHandler.EndSession)

		// Page and Action Tracking
		behaviorRoutes.POST("/page-views", behaviorHandler.TrackPageView)
		behaviorRoutes.POST("/actions", behaviorHandler.TrackUserAction)
		behaviorRoutes.POST("/interactions", behaviorHandler.RecordInteraction)

		// Content Engagement
		behaviorRoutes.POST("/content-engagement", behaviorHandler.TrackContentEngagement)
		behaviorRoutes.GET("/content/:contentId/interest-score", behaviorHandler.GetInterestScore)

		// Analytics and Insights
		behaviorRoutes.GET("/analytics", behaviorHandler.GetUserBehaviorAnalytics)
		behaviorRoutes.GET("/insights", behaviorHandler.GetBehaviorInsights)
		behaviorRoutes.GET("/preferences", behaviorHandler.GetUserContentPreferences)

		// Social Intelligence
		behaviorRoutes.GET("/similar-users", behaviorHandler.GetSimilarUsers)

		// Recommendation Tracking
		behaviorRoutes.POST("/recommendations", behaviorHandler.TrackRecommendation)

		// A/B Testing
		behaviorRoutes.POST("/experiments", behaviorHandler.TrackExperiment)
	}

	// Admin behavior routes (for platform analytics)
	adminBehaviorRoutes := router.Group("/api/v1/admin/behavior")
	adminBehaviorRoutes.Use(authMiddleware.RequireAuth())
	adminBehaviorRoutes.Use(authMiddleware.RequireRole("admin"))
	{
		// Platform-wide behavior analytics would go here
		// adminBehaviorRoutes.GET("/platform-analytics", behaviorHandler.GetPlatformBehaviorAnalytics)
		// adminBehaviorRoutes.GET("/user-segments", behaviorHandler.GetUserSegments)
		// adminBehaviorRoutes.GET("/behavior-patterns", behaviorHandler.GetBehaviorPatterns)
	}
}

// SetupEnhancedFeedRoutes sets up behavior-enhanced feed routes using regular FeedHandler
func SetupEnhancedFeedRoutes(router *gin.Engine, feedHandler *handlers.FeedHandler, authMiddleware middleware.AuthMiddleware, behaviorMiddleware *middleware.BehaviorTrackingMiddleware) {

	// Enhanced feed routes with behavior tracking
	feedRoutes := router.Group("/api/v1/feeds")
	feedRoutes.Use(authMiddleware.RequireAuth())
	feedRoutes.Use(behaviorMiddleware.TrackContentInteraction())
	{
		// Behavior-driven feeds (using regular FeedHandler with behavior algorithm)
		feedRoutes.GET("/smart-personalized", func(c *gin.Context) {
			c.Set("algorithm", "behavior")
			feedHandler.GetPersonalizedFeed(c)
		})
		feedRoutes.GET("/behavior-following", func(c *gin.Context) {
			c.Set("algorithm", "behavior")
			feedHandler.GetFollowingFeed(c)
		})
		feedRoutes.GET("/smart-trending", func(c *gin.Context) {
			c.Set("algorithm", "behavior")
			feedHandler.GetTrendingFeed(c)
		})
		feedRoutes.GET("/intelligent-discover", func(c *gin.Context) {
			c.Set("algorithm", "behavior")
			feedHandler.GetDiscoverFeed(c)
		})

		// Enhanced feed interactions
		feedRoutes.POST("/interactions/enhanced", feedHandler.RecordInteraction)
		feedRoutes.GET("/analytics/enhanced", feedHandler.GetFeedAnalytics)

		// Feed optimization
		feedRoutes.POST("/optimize", feedHandler.RefreshFeed)
		feedRoutes.GET("/recommendations/explain", feedHandler.GetBehaviorInsights)
	}

	// Original feed routes (still available)
	originalFeedRoutes := router.Group("/api/v1/feeds/original")
	originalFeedRoutes.Use(authMiddleware.RequireAuth())
	originalFeedRoutes.Use(behaviorMiddleware.TrackContentInteraction())
	{
		originalFeedRoutes.GET("/personalized", feedHandler.GetPersonalizedFeed)
		originalFeedRoutes.GET("/following", feedHandler.GetFollowingFeed)
		originalFeedRoutes.GET("/trending", feedHandler.GetTrendingFeed)
		originalFeedRoutes.GET("/discover", feedHandler.GetDiscoverFeed)
	}
}