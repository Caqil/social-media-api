// internal/routes/post_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupPostRoutes sets up post-related routes
func SetupPostRoutes(router *gin.Engine, postHandler *handlers.PostHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public post routes
	posts := router.Group("/api/v1/posts")
	{
		// Post discovery (public/optional auth)
		posts.GET("/search", authMiddleware.OptionalAuth(), postHandler.SearchPosts)
		posts.GET("/trending", authMiddleware.OptionalAuth(), postHandler.GetTrendingPosts)
		posts.GET("/:id", authMiddleware.OptionalAuth(), postHandler.GetPost)
		posts.GET("/:id/stats", authMiddleware.OptionalAuth(), postHandler.GetPostStats)
		posts.GET("/:id/likes", authMiddleware.OptionalAuth(), postHandler.GetPostLikes)
	}

	// Protected post routes
	postsProtected := router.Group("/api/v1/posts")
	postsProtected.Use(authMiddleware.RequireAuth())
	{
		// Post creation and management
		postsProtected.POST("/", middleware.PostRateLimit(), postHandler.CreatePost)
		postsProtected.PUT("/:id", postHandler.UpdatePost)
		postsProtected.DELETE("/:id", postHandler.DeletePost)

		// Post interactions
		postsProtected.POST("/:id/like", middleware.LikeRateLimit(), postHandler.LikePost)
		postsProtected.DELETE("/:id/like", postHandler.UnlikePost)
		postsProtected.POST("/:id/report", postHandler.ReportPost)

		// Post management
		postsProtected.POST("/:id/pin", postHandler.PinPost)
		postsProtected.DELETE("/:id/pin", postHandler.UnpinPost)

		// User-specific post endpoints
		postsProtected.GET("/feed", postHandler.GetFeed)
		postsProtected.GET("/user/:userId", postHandler.GetUserPosts)
	}
}
