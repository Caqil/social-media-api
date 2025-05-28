// internal/routes/social_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupSocialRoutes sets up social feed and interaction routes
func SetupSocialRoutes(router *gin.Engine, feedHandler *handlers.FeedHandler, searchHandler *handlers.SearchHandler, likeHandler *handlers.LikeHandler, authMiddleware *middleware.AuthMiddleware) {
	// Feed routes
	feeds := router.Group("/api/v1/feeds")
	feeds.Use(authMiddleware.RequireAuth())
	{
		// Different feed types
		feeds.GET("/personalized", feedHandler.GetPersonalizedFeed)
		feeds.GET("/following", feedHandler.GetFollowingFeed)
		feeds.GET("/trending", feedHandler.GetTrendingFeed)
		feeds.GET("/discover", feedHandler.GetDiscoverFeed)
		
		// Feed interactions
		feeds.POST("/interactions", feedHandler.RecordInteraction)
		feeds.POST("/refresh", feedHandler.RefreshFeed)
		feeds.POST("/posts/:postId/hide", feedHandler.HidePost)
		feeds.POST("/report-issue", feedHandler.ReportFeedIssue)
		
		// Feed preferences
		feeds.GET("/preferences", feedHandler.GetFeedPreferences)
		feeds.PUT("/preferences", feedHandler.UpdateFeedPreferences)
		feeds.GET("/analytics", feedHandler.GetFeedAnalytics)
	}

	// Search routes
	search := router.Group("/api/v1/search")
	{
		// Public search (optional auth for personalization)
		search.GET("/", authMiddleware.OptionalAuth(), searchHandler.Search)
		search.GET("/posts", authMiddleware.OptionalAuth(), searchHandler.SearchPosts)
		search.GET("/users", authMiddleware.OptionalAuth(), searchHandler.SearchUsers)
		search.GET("/hashtags", searchHandler.SearchHashtags)
		search.GET("/suggestions", authMiddleware.OptionalAuth(), searchHandler.GetSearchSuggestions)
		
		// Trending and popular content
		search.GET("/trending/hashtags", searchHandler.GetTrendingHashtags)
		search.GET("/popular", searchHandler.GetPopularSearches)
	}

	// Protected search routes
	searchProtected := router.Group("/api/v1/search")
	searchProtected.Use(authMiddleware.RequireAuth())
	{
		// User-specific search features
		searchProtected.GET("/history", searchHandler.GetSearchHistory)
		searchProtected.DELETE("/history", searchHandler.ClearSearchHistory)
	}

	// Admin search routes
	searchAdmin := router.Group("/api/v1/search")
	searchAdmin.Use(authMiddleware.RequireAuth())
	searchAdmin.Use(middleware.RequireAdmin())
	{
		searchAdmin.POST("/index", searchHandler.IndexContent)
		searchAdmin.PUT("/hashtags", searchHandler.UpdateHashtagInfo)
	}

	// Like/Reaction routes
	reactions := router.Group("/api/v1/reactions")
	{
		// Public reaction viewing
		reactions.GET("/types", likeHandler.GetReactionTypes)
		reactions.GET("/:targetType/:targetId", likeHandler.GetLikes)
		reactions.GET("/:targetType/:targetId/summary", likeHandler.GetReactionSummary)
		reactions.GET("/:targetType/:targetId/stats", likeHandler.GetDetailedReactionStats)
		reactions.GET("/trending", likeHandler.GetTrendingReactions)
	}

	reactionsProtected := router.Group("/api/v1/reactions")
	reactionsProtected.Use(authMiddleware.RequireAuth())
	{
		// Reaction management
		reactionsProtected.POST("/", middleware.LikeRateLimit(), likeHandler.CreateLike)
		reactionsProtected.PUT("/:id", likeHandler.UpdateLike)
		reactionsProtected.DELETE("/:targetType/:targetId", likeHandler.DeleteLike)
		
		// User reaction queries
		reactionsProtected.GET("/:targetType/:targetId/check", likeHandler.CheckUserReaction)
		reactionsProtected.GET("/my-reactions", likeHandler.GetMyReactions)
		reactionsProtected.GET("/users/:userId", likeHandler.GetUserLikes)
	}

	// Admin reaction routes
	reactionsAdmin := router.Group("/api/v1/reactions")
	reactionsAdmin.Use(authMiddleware.RequireAuth())
	reactionsAdmin.Use(middleware.RequireAdmin())
	{
		reactionsAdmin.POST("/bulk", likeHandler.BulkReaction)
		reactionsAdmin.GET("/stats", likeHandler.GetLikeStats)
	}
}