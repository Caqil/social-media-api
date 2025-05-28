// internal/routes/admin_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes sets up admin and moderation routes
func SetupAdminRoutes(router *gin.Engine, reportHandler *handlers.ReportHandler, authMiddleware *middleware.AuthMiddleware) {
	// Admin routes - require admin authentication
	admin := router.Group("/api/v1/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(middleware.RequireAdmin())
	{
		// Report management
		reports := admin.Group("/reports")
		{
			// Report CRUD and management
			reports.GET("/", reportHandler.GetReports)
			reports.GET("/:id", reportHandler.GetReport)
			reports.PUT("/:id", reportHandler.UpdateReport)
			reports.POST("/:id/assign", reportHandler.AssignReport)
			reports.POST("/:id/resolve", reportHandler.ResolveReport)
			reports.POST("/:id/reject", reportHandler.RejectReport)

			// Bulk operations
			reports.POST("/bulk-update", reportHandler.BulkUpdateReports)

			// Report analytics
			reports.GET("/stats", reportHandler.GetReportStats)
			reports.GET("/summary", reportHandler.GetReportSummary)

			// Reports by target
			reports.GET("/target/:type/:id", reportHandler.GetReportsByTarget)
			reports.GET("/user/:userId", reportHandler.GetUserReports)

			// Report configuration
			reports.GET("/reasons", reportHandler.GetReportReasons)
			reports.GET("/categories", reportHandler.GetReportCategories)
		}

		// Admin rate limiting with higher limits
		admin.Use(middleware.AdminRateLimit())
	}

	// Moderator routes - require moderator authentication
	moderator := router.Group("/api/v1/moderator")
	moderator.Use(authMiddleware.RequireAuth())
	moderator.Use(middleware.RequireModerator())
	{
		// Moderator-specific report management
		moderatorReports := moderator.Group("/reports")
		{
			// Reports assigned to current moderator
			moderatorReports.GET("/pending", reportHandler.GetPendingReports)
			moderatorReports.GET("/", reportHandler.GetReports)
			moderatorReports.GET("/:id", reportHandler.GetReport)
			moderatorReports.PUT("/:id", reportHandler.UpdateReport)
			moderatorReports.POST("/:id/resolve", reportHandler.ResolveReport)
			moderatorReports.POST("/:id/reject", reportHandler.RejectReport)
		}
	}

	// Public report routes (for users to create reports)
	reports := router.Group("/api/v1/reports")
	reports.Use(authMiddleware.RequireAuth())
	{
		// User report creation and management
		reports.POST("/", reportHandler.CreateReport)
		reports.GET("/my-reports", reportHandler.GetMyReports)

		// Public report configuration
		reports.GET("/reasons", reportHandler.GetReportReasons)
		reports.GET("/categories", reportHandler.GetReportCategories)
	}
}
