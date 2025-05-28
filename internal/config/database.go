package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB     *mongo.Database
	Client *mongo.Client
)

func InitDB() {
	// Get MongoDB connection string from environment
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("DB_NAME", "social_media")

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	// Set global variables
	Client = client
	DB = client.Database(dbName)

	log.Println("MongoDB connected successfully")

	// Create indexes
	createIndexes()
}

func createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create unique index for username
	usernameIndex := mongo.IndexModel{
		Keys:    map[string]int{"username": 1},
		Options: options.Index().SetUnique(true),
	}

	// Create unique index for email
	emailIndex := mongo.IndexModel{
		Keys:    map[string]int{"email": 1},
		Options: options.Index().SetUnique(true),
	}

	// Create indexes on users collection
	userCollection := DB.Collection("users")
	_, err := userCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{usernameIndex, emailIndex})
	if err != nil {
		log.Printf("Failed to create indexes: %v", err)
	} else {
		log.Println("Database indexes created successfully")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		Client.Disconnect(ctx)
	}
}
