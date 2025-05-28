// internal/routes/comment_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupCommentRoutes sets up comment-related routes
func SetupCommentRoutes(router *gin.Engine, commentHandler *handlers.CommentHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public comment routes
	comments := router.Group("/api/v1/comments")
	{
		// Comment viewing (public/optional auth)
		comments.GET("/:id", authMiddleware.OptionalAuth(), commentHandler.GetComment)
		comments.GET("/:id/thread", authMiddleware.OptionalAuth(), commentHandler.GetCommentThread)
		comments.GET("/:id/replies", authMiddleware.OptionalAuth(), commentHandler.GetCommentReplies)
		comments.GET("/:id/likes", authMiddleware.OptionalAuth(), commentHandler.GetCommentLikes)
	}

	// Protected comment routes
	commentsProtected := router.Group("/api/v1/comments")
	commentsProtected.Use(authMiddleware.RequireAuth())
	{
		// Comment creation and management
		commentsProtected.POST("/", middleware.CommentRateLimit(), commentHandler.CreateComment)
		commentsProtected.PUT("/:id", commentHandler.UpdateComment)
		commentsProtected.DELETE("/:id", commentHandler.DeleteComment)

		// Comment interactions
		commentsProtected.POST("/:id/like", middleware.LikeRateLimit(), commentHandler.LikeComment)
		commentsProtected.DELETE("/:id/like", commentHandler.UnlikeComment)
		commentsProtected.POST("/:id/report", commentHandler.ReportComment)

		// Comment moderation (post author only)
		commentsProtected.POST("/:id/pin", commentHandler.PinComment)
		commentsProtected.DELETE("/:id/pin", commentHandler.UnpinComment)

		// User-specific comment endpoints
		commentsProtected.GET("/user/:userId", commentHandler.GetUserComments)
	}

	// Post-specific comment routes - FIXED: Changed :postId to :id to avoid conflict
	postComments := router.Group("/api/v1/posts/:id/comments")
	{
		// Public comment viewing
		postComments.GET("/", authMiddleware.OptionalAuth(), commentHandler.GetPostComments)
	}
}
