// internal/routes/media_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMediaRoutes sets up media upload and management routes
func SetupMediaRoutes(router *gin.Engine, mediaHandler *handlers.MediaHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public media routes
	media := router.Group("/api/v1/media")
	{
		// Media viewing (public/optional auth)
		media.GET("/:id", authMiddleware.OptionalAuth(), mediaHandler.GetMedia)
		media.GET("/:id/download", authMiddleware.OptionalAuth(), mediaHandler.DownloadMedia)
		media.GET("/:id/variants/:variant", authMiddleware.OptionalAuth(), mediaHandler.GetMediaVariant)
		media.GET("/:id/metadata", authMiddleware.OptionalAuth(), mediaHandler.GetMediaMetadata)
		media.GET("/search", authMiddleware.OptionalAuth(), mediaHandler.SearchMedia)
		media.GET("/user/:userId", authMiddleware.OptionalAuth(), mediaHandler.GetUserMedia)
	}

	// Protected media routes
	mediaProtected := router.Group("/api/v1/media")
	mediaProtected.Use(authMiddleware.RequireAuth())
	{
		// Media upload and management
		mediaProtected.POST("/upload", mediaHandler.UploadMedia)
		mediaProtected.POST("/bulk-upload", mediaHandler.BulkUploadMedia)
		mediaProtected.PUT("/:id", mediaHandler.UpdateMedia)
		mediaProtected.DELETE("/:id", mediaHandler.DeleteMedia)

		// Media statistics
		mediaProtected.GET("/stats", mediaHandler.GetMediaStats)
	}
}
