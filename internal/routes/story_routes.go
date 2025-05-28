// internal/routes/story_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupStoryRoutes sets up story-related routes
func SetupStoryRoutes(router *gin.Engine, storyHandler *handlers.StoryHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public story routes
	stories := router.Group("/api/v1/stories")
	{
		// Story discovery (public/optional auth)
		stories.GET("/active", authMiddleware.OptionalAuth(), storyHandler.GetActiveStories)
		stories.GET("/:id", authMiddleware.OptionalAuth(), storyHandler.GetStory)
		stories.GET("/:id/views", authMiddleware.OptionalAuth(), storyHandler.GetStoryViews)
		stories.GET("/:id/reactions", authMiddleware.OptionalAuth(), storyHandler.GetStoryReactions)
		stories.GET("/:id/stats", authMiddleware.OptionalAuth(), storyHandler.GetStoryStats)
		stories.GET("/user/:userId", authMiddleware.OptionalAuth(), storyHandler.GetUserStories)
	}

	// Protected story routes
	storiesProtected := router.Group("/api/v1/stories")
	storiesProtected.Use(authMiddleware.RequireAuth())
	{
		// Story creation and management
		storiesProtected.POST("/", storyHandler.CreateStory)
		storiesProtected.PUT("/:id", storyHandler.UpdateStory)
		storiesProtected.DELETE("/:id", storyHandler.DeleteStory)

		// Story interactions
		storiesProtected.POST("/:id/view", storyHandler.ViewStory)
		storiesProtected.POST("/:id/react", storyHandler.ReactToStory)
		storiesProtected.DELETE("/:id/react", storyHandler.UnreactToStory)

		// Story management
		storiesProtected.POST("/:id/archive", storyHandler.ArchiveStory)
		storiesProtected.GET("/archived", storyHandler.GetArchivedStories)

		// Story feeds
		storiesProtected.GET("/following", storyHandler.GetFollowingStories)
	}

	// Story highlights
	highlights := router.Group("/api/v1/story-highlights")
	{
		// Public highlight viewing
		highlights.GET("/user/:userId", storyHandler.GetUserStoryHighlights)
	}

	highlightsProtected := router.Group("/api/v1/story-highlights")
	highlightsProtected.Use(authMiddleware.RequireAuth())
	{
		// Highlight management
		highlightsProtected.POST("/", storyHandler.CreateStoryHighlight)
		highlightsProtected.PUT("/:id", storyHandler.UpdateStoryHighlight)
		highlightsProtected.DELETE("/:id", storyHandler.DeleteStoryHighlight)
	}
}
