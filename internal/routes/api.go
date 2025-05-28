// internal/routes/api.go
package routes

import (
	"net/http"

	"social-media-api/internal/config"
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"
	"social-media-api/internal/services"

	"github.com/gin-gonic/gin"
)

// APIRouter holds all route handlers and services
type APIRouter struct {
	// Handlers
	AuthHandler         *handlers.AuthHandler
	AdminHandler        *handlers.AdminHandler
	UserHandler         *handlers.UserHandler
	PostHandler         *handlers.PostHandler
	CommentHandler      *handlers.CommentHandler
	FollowHandler       *handlers.FollowHandler
	MessageHandler      *handlers.MessageHandler
	ConversationHandler *handlers.ConversationHandler
	StoryHandler        *handlers.StoryHandler
	GroupHandler        *handlers.GroupHandler
	FeedHandler         *handlers.FeedHandler
	SearchHandler       *handlers.SearchHandler
	NotificationHandler *handlers.NotificationHandler
	MediaHandler        *handlers.MediaHandler
	LikeHandler         *handlers.LikeHandler
	ReportHandler       *handlers.ReportHandler
	BehaviorHandler     *handlers.UserBehaviorHandler
	// Middleware
	AuthMiddleware     *middleware.AuthMiddleware
	BehaviorMiddleware *middleware.BehaviorTrackingMiddleware
	// Services (for dependency injection)
	Services *Services
}

// Services holds all service instances
type Services struct {
	AuthService         *services.AuthService
	AdminService        *services.AdminService
	UserService         *services.UserService
	PostService         *services.PostService
	CommentService      *services.CommentService
	FollowService       *services.FollowService
	MessageService      *services.MessageService
	ConversationService *services.ConversationService
	StoryService        *services.StoryService
	GroupService        *services.GroupService
	FeedService         *services.FeedService
	SearchService       *services.SearchService
	NotificationService *services.NotificationService
	MediaService        *services.MediaService
	LikeService         *services.LikeService
	ReportService       *services.ReportService
	EmailService        *services.EmailService
	PushService         *services.PushService
	BehaviorService     *services.UserBehaviorService // Added behavior service
	AnalyticsService    *services.AnalyticsService
}

// SetupRoutes initializes all routes for the API
func SetupRoutes(router *gin.Engine, apiRouter *APIRouter) {
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.GlobalErrorHandler())

	// Security middleware
	if config.GetConfig().Security.EnableCSRF {
		// Add CSRF protection if enabled
	}

	// Health check endpoint
	router.GET("/health", healthCheck)
	router.GET("/api/v1/health", healthCheck)

	// API version info
	router.GET("/api/v1", apiInfo)

	// Setup all route groups
	SetupAuthRoutes(router, apiRouter.AuthHandler, apiRouter.AuthMiddleware)
	SetupUserRoutes(router, apiRouter.UserHandler, apiRouter.AuthMiddleware)
	SetupPostRoutes(router, apiRouter.PostHandler, apiRouter.AuthMiddleware)
	SetupCommentRoutes(router, apiRouter.CommentHandler, apiRouter.AuthMiddleware)
	SetupFollowRoutes(router, apiRouter.FollowHandler, apiRouter.AuthMiddleware)
	SetupMessagingRoutes(router, apiRouter.MessageHandler, apiRouter.ConversationHandler, apiRouter.AuthMiddleware)
	SetupStoryRoutes(router, apiRouter.StoryHandler, apiRouter.AuthMiddleware)
	SetupGroupRoutes(router, apiRouter.GroupHandler, apiRouter.AuthMiddleware)
	SetupSocialRoutes(router, apiRouter.FeedHandler, apiRouter.SearchHandler, apiRouter.LikeHandler, apiRouter.AuthMiddleware)
	SetupNotificationRoutes(router, apiRouter.NotificationHandler, apiRouter.AuthMiddleware)
	SetupMediaRoutes(router, apiRouter.MediaHandler, apiRouter.AuthMiddleware)
	SetupAdminRoutes(router, apiRouter.AdminHandler, apiRouter.AuthMiddleware)

	// 404 handler
	router.NoRoute(middleware.NotFoundHandler())

	// 405 handler
	router.NoMethod(middleware.MethodNotAllowedHandler())
}

// healthCheck returns the health status of the API
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": gin.H{
			"unix": gin.H{
				"seconds": gin.H{
					"now": nil,
				},
			},
		},
		"service": "social-media-api",
		"version": "v1.0.0",
	})
}

// apiInfo returns API information
func apiInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "Social Media API",
		"version":     "v1.0.0",
		"description": "A comprehensive social media platform API",
		"endpoints": gin.H{
			"auth":          "/api/v1/auth",
			"users":         "/api/v1/users",
			"posts":         "/api/v1/posts",
			"comments":      "/api/v1/comments",
			"stories":       "/api/v1/stories",
			"groups":        "/api/v1/groups",
			"messaging":     "/api/v1/messaging",
			"feeds":         "/api/v1/feeds",
			"search":        "/api/v1/search",
			"notifications": "/api/v1/notifications",
			"media":         "/api/v1/media",
			"reactions":     "/api/v1/reactions",
			"reports":       "/api/v1/reports",
			"admin":         "/api/v1/admin",
		},
		"features": []string{
			"User authentication and authorization",
			"Social posts and comments",
			"Stories with reactions",
			"Group management",
			"Real-time messaging",
			"Personalized feeds",
			"Search functionality",
			"Media upload and management",
			"Notification system",
			"Content moderation",
			"Admin panel",
		},
	})
}

// NewAPIRouter creates a new API router with all dependencies
func NewAPIRouter(services *Services, authMiddleware *middleware.AuthMiddleware, behaviorMiddleware *middleware.BehaviorTrackingMiddleware) *APIRouter {
	return &APIRouter{
		// Initialize handlers with their respective services
		AuthHandler:         handlers.NewAuthHandler(services.AuthService, services.UserService),
		UserHandler:         handlers.NewUserHandler(services.UserService),
		PostHandler:         handlers.NewPostHandler(services.PostService),
		CommentHandler:      handlers.NewCommentHandler(services.CommentService),
		FollowHandler:       handlers.NewFollowHandler(services.FollowService),
		MessageHandler:      handlers.NewMessageHandler(services.MessageService, services.ConversationService, nil), // WebSocket hub would be injected here
		ConversationHandler: handlers.NewConversationHandler(services.ConversationService, services.MessageService, services.NotificationService),
		StoryHandler:        handlers.NewStoryHandler(services.StoryService),
		GroupHandler:        handlers.NewGroupHandler(services.GroupService),
		FeedHandler:         handlers.NewFeedHandler(services.FeedService, services.BehaviorService),
		SearchHandler:       handlers.NewSearchHandler(services.SearchService),
		NotificationHandler: handlers.NewNotificationHandler(services.NotificationService),
		MediaHandler:        handlers.NewMediaHandler(services.MediaService),
		LikeHandler:         handlers.NewLikeHandler(services.LikeService),
		ReportHandler:       handlers.NewReportHandler(services.ReportService),
		BehaviorHandler:     handlers.NewUserBehaviorHandler(services.BehaviorService, services.AnalyticsService),
		// Middleware
		AuthMiddleware:     authMiddleware,
		BehaviorMiddleware: behaviorMiddleware,
		// Services
		Services: services,
	}
}
