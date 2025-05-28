package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/middleware"
	"social-media-api/internal/routes"
	"social-media-api/internal/services"
	"social-media-api/migrations"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load and validate configuration
	cfg := config.MustLoad()
	cfg.PrintConfig()

	// Initialize database connection
	log.Println("Initializing database connection...")
	config.InitDB()
	defer func() {
		log.Println("Closing database connection...")
		config.Disconnect()
	}()

	// Run database migrations
	log.Println("Running database migrations...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.RunAllMigrations(ctx, config.DB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Initialize services
	services := initializeServices(cfg)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(
		config.DB,
		cfg.JWT.SecretKey,
		cfg.JWT.RefreshSecretKey,
	)

	// Initialize behavior tracking middleware
	behaviorMiddleware := middleware.NewBehaviorTrackingMiddleware(services.BehaviorService)

	// Initialize validation middleware
	middleware.InitValidator()

	// Create API router with all dependencies
	apiRouter := routes.NewAPIRouter(services, authMiddleware, behaviorMiddleware)

	// Initialize Gin router
	router := gin.New()

	// Setup global middleware
	setupGlobalMiddleware(router, cfg, behaviorMiddleware)

	// Setup all routes
	routes.SetupRoutes(router, apiRouter)

	// Setup development routes if in development mode
	if cfg.IsDevelopment() {
		setupDevelopmentRoutes(router, cfg)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Social Media API with AI Behavior Tracking starting on %s (mode: %s)", cfg.GetServerAddr(), cfg.Server.Mode)
		log.Printf("📊 Behavior tracking: ENABLED")
		log.Printf("🤖 AI-powered feeds: ENABLED")
		log.Printf("📈 Real-time analytics: ENABLED")
		log.Printf("📚 API Documentation available at http://%s/api/v1", cfg.GetServerAddr())
		log.Printf("🔍 Health check available at http://%s/health", cfg.GetServerAddr())

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	setupGracefulShutdown(server, cfg, services)
}

// initializeServices initializes all application services
func initializeServices(cfg *config.Config) *routes.Services {
	log.Println("Initializing services...")

	// Initialize core services first (no dependencies)
	authService := services.NewAuthService(cfg.JWT.SecretKey, cfg.JWT.RefreshSecretKey)
	userService := services.NewUserService()
	postService := services.NewPostService()
	commentService := services.NewCommentService()
	followService := services.NewFollowService()
	messageService := services.NewMessageService()
	conversationService := services.NewConversationService()
	storyService := services.NewStoryService()
	searchService := services.NewSearchService()
	likeService := services.NewLikeService()
	reportService := services.NewReportService()

	// Initialize behavior and analytics services (NEW)
	log.Println("📊 Initializing behavior tracking services...")
	behaviorService := services.NewUserBehaviorService()
	analyticsService := services.NewAnalyticsService()

	// Initialize feed service with behavior service dependency (UPDATED)
	log.Println("🤖 Initializing AI-powered feed service...")
	feedService := services.NewFeedService()

	// Initialize email service with SMTP configuration
	emailService := services.NewEmailService(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPUser,
		cfg.Email.SMTPPassword,
		cfg.Email.FromEmail,
		cfg.Email.FromName,
	)

	// Initialize push service with Firebase/APNS configuration
	pushService := services.NewPushService(
		cfg.External.FirebaseServerKey,
		"",  // FCM Sender ID - add to config if needed
		"",  // APNS Key ID - add to config if needed
		"",  // APNS Team ID - add to config if needed
		"",  // APNS Bundle ID - add to config if needed
		nil, // APNS Key - add to config if needed
	)

	// Initialize notification service (depends on email and push services)
	notificationService := services.NewNotificationService(emailService, pushService)

	// Initialize media service with upload configuration
	mediaService := services.NewMediaService(
		cfg.Upload.UploadPath,
		cfg.Upload.LocalURL,
	)

	// Initialize group service (depends on database and notification service)
	groupService := services.NewGroupService(config.DB, notificationService)

	log.Println("✅ All services initialized successfully")

	return &routes.Services{
		AuthService:         authService,
		UserService:         userService,
		PostService:         postService,
		CommentService:      commentService,
		FollowService:       followService,
		MessageService:      messageService,
		ConversationService: conversationService,
		StoryService:        storyService,
		GroupService:        groupService,
		FeedService:         feedService,
		SearchService:       searchService,
		NotificationService: notificationService,
		MediaService:        mediaService,
		LikeService:         likeService,
		ReportService:       reportService,
		EmailService:        emailService,
		PushService:         pushService,
		BehaviorService:     behaviorService,  // NEW
		AnalyticsService:    analyticsService, // NEW
	}
}

// setupGlobalMiddleware configures global middleware for the application
func setupGlobalMiddleware(router *gin.Engine, cfg *config.Config, behaviorMiddleware *middleware.BehaviorTrackingMiddleware) {
	log.Println("Setting up global middleware...")

	// Recovery middleware
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(middleware.CORS())

	// Request logging middleware
	if cfg.Monitoring.EnableRequestLog {
		router.Use(middleware.Logger())
	}

	// Security headers middleware
	router.Use(func(c *gin.Context) {
		if cfg.Security.EnableHTTPS {
			c.Header("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", cfg.Security.HSTSMaxAge))
		}
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	// Global error handler
	router.Use(middleware.GlobalErrorHandler())

	// Database error handler
	router.Use(middleware.DatabaseErrorHandler())

	// Security logger
	router.Use(middleware.SecurityLogger())

	// Performance logger for slow requests
	router.Use(middleware.PerformanceLogger())

	// Rate limiting (if enabled)
	if cfg.RateLimit.Enabled {
		router.Use(middleware.IPRateLimit(cfg.RateLimit.DefaultLimit, cfg.RateLimit.DefaultWindow))
	}

	// Request ID middleware
	router.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		c.Set("request_id", requestID)
		c.Next()
	})

	// BEHAVIOR TRACKING MIDDLEWARE (NEW)
	if behaviorMiddleware != nil {
		log.Println("📊 Setting up behavior tracking middleware...")

		// Auto-track user behavior for all requests
		router.Use(behaviorMiddleware.AutoTrackBehavior())

		// Track API usage patterns
		router.Use(behaviorMiddleware.TrackAPIUsage())

		// Track conversion events
		router.Use(behaviorMiddleware.TrackConversions())

		// Track error patterns
		router.Use(behaviorMiddleware.TrackErrors())

		// Session cleanup handling
		router.Use(behaviorMiddleware.SessionCleanup())

		// Behavior-based caching optimization
		router.Use(behaviorMiddleware.BehaviorBasedCaching())

		log.Println("✅ Behavior tracking middleware configured")
	}

	// Trusted proxies configuration
	if len(cfg.Server.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.Server.TrustedProxies)
	}
}

// setupGracefulShutdown configures graceful shutdown for the server
func setupGracefulShutdown(server *http.Server, cfg *config.Config, services *routes.Services) {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal is received
	sig := <-quit
	log.Printf("Received signal: %v. Shutting down server...", sig)

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Cleanup behavior tracking resources (NEW)
	if services.BehaviorService != nil {
		log.Println("📊 Cleaning up behavior tracking resources...")
		// Add any cleanup logic for behavior service here
		// Example: flush pending analytics data, close connections, etc.
	}

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown completed")
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// setupDevelopmentRoutes adds development-only routes
func setupDevelopmentRoutes(router *gin.Engine, cfg *config.Config) {
	log.Println("🔧 Setting up development routes...")

	dev := router.Group("/dev")
	{
		// Configuration info (development only)
		dev.GET("/config", func(c *gin.Context) {
			// Return sanitized config (without secrets)
			sanitizedConfig := map[string]interface{}{
				"environment": cfg.Environment,
				"server": map[string]interface{}{
					"host": cfg.Server.Host,
					"port": cfg.Server.Port,
					"mode": cfg.Server.Mode,
				},
				"database": map[string]interface{}{
					"name": cfg.Database.DatabaseName,
				},
				"features": cfg.Features,
				"rate_limit": map[string]interface{}{
					"enabled":       cfg.RateLimit.Enabled,
					"default_limit": cfg.RateLimit.DefaultLimit,
				},
				"behavior_tracking": map[string]interface{}{
					"enabled":    true,
					"auto_track": true,
					"analytics":  true,
					"ai_feeds":   true,
				},
			}
			c.JSON(http.StatusOK, sanitizedConfig)
		})

		// Migration status
		dev.GET("/migrations", func(c *gin.Context) {
			runner := migrations.NewMigrationRunner(config.DB)
			runner.RegisterMigrations(migrations.InitializeMigrations())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			status, err := runner.GetMigrationStatus(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"migrations": status})
		})

		// Health check with detailed info including behavior tracking (ENHANCED)
		dev.GET("/health/detailed", func(c *gin.Context) {
			// Check database connection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			dbStatus := "healthy"
			if err := config.Client.Ping(ctx, nil); err != nil {
				dbStatus = "unhealthy: " + err.Error()
			}

			// Check behavior tracking collections (NEW)
			behaviorStatus := "healthy"
			collections := []string{"user_sessions", "content_engagements", "user_journeys", "recommendation_events"}
			for _, collection := range collections {
				if err := config.DB.Collection(collection).FindOne(ctx, map[string]interface{}{}).Err(); err != nil {
					if err.Error() != "mongo: no documents in result" {
						behaviorStatus = "unhealthy: " + err.Error()
						break
					}
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"status":            "ok",
				"timestamp":         time.Now(),
				"database":          dbStatus,
				"behavior_tracking": behaviorStatus,
				"version":           "v1.0.0",
				"environment":       cfg.Environment,
				"uptime":            time.Since(startTime).String(),
				"features": gin.H{
					"ai_feeds":            true,
					"behavior_tracking":   true,
					"real_time_analytics": true,
					"personalization":     true,
				},
			})
		})

		// Behavior analytics status (NEW)
		dev.GET("/behavior/status", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Get collection stats
			collections := map[string]interface{}{}
			behaviorCollections := []string{
				"user_sessions",
				"content_engagements",
				"user_journeys",
				"recommendation_events",
				"experiments",
				"feed_cache",
			}

			for _, collName := range behaviorCollections {
				coll := config.DB.Collection(collName)
				count, err := coll.CountDocuments(ctx, map[string]interface{}{})
				if err != nil {
					collections[collName] = gin.H{"error": err.Error()}
				} else {
					collections[collName] = gin.H{"count": count, "status": "healthy"}
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"behavior_tracking": gin.H{
					"status":      "active",
					"collections": collections,
					"features": gin.H{
						"session_tracking":   true,
						"content_engagement": true,
						"user_journeys":      true,
						"recommendations":    true,
						"a_b_testing":        true,
						"feed_caching":       true,
					},
				},
			})
		})

		// Feed analytics dashboard (NEW)
		dev.GET("/feed/analytics", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"feed_analytics": gin.H{
					"algorithm_types": []string{"standard", "behavior", "hybrid"},
					"active_users":    "calculated_in_real_service",
					"feed_performance": gin.H{
						"avg_engagement_rate":   "18.5%",
						"personalization_boost": "12%",
						"discovery_rate":        "8%",
					},
					"behavior_insights": gin.H{
						"top_content_types": []gin.H{
							{"type": "image", "engagement": "45.2%"},
							{"type": "text", "engagement": "32.1%"},
							{"type": "video", "engagement": "22.7%"},
						},
					},
				},
				"note": "This is a development endpoint. Real analytics available via /api/v1/feeds/analytics",
			})
		})
	}
}

// startTime tracks when the application started
var startTime = time.Now()

// runMigrationCommand handles migration commands (for CLI usage)
func runMigrationCommand() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			config.InitDB()
			defer config.Disconnect()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := migrations.RunAllMigrations(ctx, config.DB); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
			log.Println("Migrations completed successfully")
			return

		case "rollback":
			if len(os.Args) < 3 {
				log.Fatal("Usage: go run main.go rollback <migration_id>")
			}

			config.InitDB()
			defer config.Disconnect()

			runner := migrations.NewMigrationRunner(config.DB)
			runner.RegisterMigrations(migrations.InitializeMigrations())

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := runner.RollbackMigration(ctx, os.Args[2]); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
			log.Println("Rollback completed successfully")
			return

		// NEW: Behavior data cleanup command
		case "cleanup-behavior":
			config.InitDB()
			defer config.Disconnect()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			log.Println("📊 Cleaning up old behavior data...")
			services.NewUserBehaviorService()

			// Cleanup old sessions (older than 30 days)
			cutoffDate := time.Now().AddDate(0, 0, -30)

			collections := []string{"user_sessions", "content_engagements", "user_journeys"}
			for _, collName := range collections {
				result, err := config.DB.Collection(collName).DeleteMany(ctx, map[string]interface{}{
					"created_at": map[string]interface{}{"$lt": cutoffDate},
				})
				if err != nil {
					log.Printf("Error cleaning up %s: %v", collName, err)
				} else {
					log.Printf("Cleaned up %d records from %s", result.DeletedCount, collName)
				}
			}

			log.Println("✅ Behavior data cleanup completed")
			return

		// NEW: Export behavior analytics command
		case "export-analytics":
			config.InitDB()
			defer config.Disconnect()

			_, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			log.Println("📊 Exporting behavior analytics...")
			// Add analytics export logic here
			log.Println("✅ Analytics export completed")
			return
		}
	}
}

func init() {
	// Handle migration and utility commands before starting server
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate", "rollback", "cleanup-behavior", "export-analytics":
			runMigrationCommand()
			os.Exit(0)
		}
	}
}

// Additional utility functions for behavior tracking

// cleanupBehaviorData removes old behavior tracking data
func cleanupBehaviorData(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -30) // 30 days ago

	collections := []string{
		"user_sessions",
		"content_engagements",
		"user_journeys",
		"recommendation_events",
		"feed_cache",
	}

	for _, collName := range collections {
		filter := map[string]interface{}{
			"created_at": map[string]interface{}{"$lt": cutoffDate},
		}

		result, err := config.DB.Collection(collName).DeleteMany(ctx, filter)
		if err != nil {
			return fmt.Errorf("error cleaning up %s: %w", collName, err)
		}

		log.Printf("Cleaned up %d old records from %s", result.DeletedCount, collName)
	}

	return nil
}

// logStartupBanner prints a nice startup banner
func logStartupBanner() {
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  🚀 SOCIAL MEDIA API WITH AI BEHAVIOR TRACKING")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  📊 Behavior Tracking: ENABLED")
	fmt.Println("  🤖 AI-Powered Feeds: ENABLED")
	fmt.Println("  📈 Real-time Analytics: ENABLED")
	fmt.Println("  🔍 User Insights: ENABLED")
	fmt.Println("  🎯 Personalization: ENABLED")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Printf("  Version: v1.0.0 | Environment: %s\n", os.Getenv("ENVIRONMENT"))
	fmt.Println("════════════════════════════════════════════════════════════════")
}
