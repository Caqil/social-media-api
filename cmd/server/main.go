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

	// Initialize validation middleware
	middleware.InitValidator()

	// Create API router with all dependencies
	apiRouter := routes.NewAPIRouter(services, authMiddleware)

	// Initialize Gin router
	router := gin.New()

	// Setup global middleware
	setupGlobalMiddleware(router, cfg)

	// Setup all routes
	routes.SetupRoutes(router, apiRouter)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s (mode: %s)", cfg.GetServerAddr(), cfg.Server.Mode)
		log.Printf("API Documentation available at http://%s/api/v1", cfg.GetServerAddr())

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	setupGracefulShutdown(server, cfg)
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
	feedService := services.NewFeedService()
	searchService := services.NewSearchService()
	likeService := services.NewLikeService()
	reportService := services.NewReportService()

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
	}
}

// setupGlobalMiddleware configures global middleware for the application
func setupGlobalMiddleware(router *gin.Engine, cfg *config.Config) {
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

	// Trusted proxies configuration
	if len(cfg.Server.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.Server.TrustedProxies)
	}
}

// setupGracefulShutdown configures graceful shutdown for the server
func setupGracefulShutdown(server *http.Server, cfg *config.Config) {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal is received
	sig := <-quit
	log.Printf("Received signal: %v. Shutting down server...", sig)

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

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

// Additional helper functions for development/debugging

// setupDevelopmentRoutes adds development-only routes
func setupDevelopmentRoutes(router *gin.Engine, cfg *config.Config) {
	if cfg.IsDevelopment() {
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

			// Health check with detailed info
			dev.GET("/health/detailed", func(c *gin.Context) {
				// Check database connection
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				dbStatus := "healthy"
				if err := config.Client.Ping(ctx, nil); err != nil {
					dbStatus = "unhealthy: " + err.Error()
				}

				c.JSON(http.StatusOK, gin.H{
					"status":      "ok",
					"timestamp":   time.Now(),
					"database":    dbStatus,
					"version":     "v1.0.0",
					"environment": cfg.Environment,
					"uptime":      time.Since(startTime).String(),
				})
			})
		}
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
		}
	}
}

func init() {
	// Handle migration commands before starting server
	if len(os.Args) > 1 && (os.Args[1] == "migrate" || os.Args[1] == "rollback") {
		runMigrationCommand()
		os.Exit(0)
	}
}
