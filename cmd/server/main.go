package main

import (
	"log"
	"social-media-api/config"
	"social-media-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	config.InitDB()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router)

	// Start server
	log.Println("Server starting on port 8080...")
	log.Fatal(router.Run(":8080"))
}
