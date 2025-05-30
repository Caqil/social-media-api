package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DB     *mongo.Database
	Client *mongo.Client
)

// InitDB initializes MongoDB Atlas connection with optimized settings
func InitDB() {
	log.Println("🔄 Initializing MongoDB Atlas connection...")

	// Get MongoDB connection details from environment
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("DB_NAME", "social_media")

	// Validate connection string
	if mongoURI == "mongodb://localhost:27017" {
		log.Println("⚠️  Warning: Using default local MongoDB URI. Make sure to set MONGO_URI for Atlas connection.")
	}

	// Create client options with MongoDB Atlas optimizations
	clientOptions := createAtlasClientOptions(mongoURI)

	// Create connection context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to MongoDB Atlas
	log.Println("🌐 Connecting to MongoDB Atlas...")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB Atlas: %v", err)
	}

	// Test the connection with Atlas-specific ping
	log.Println("🔄 Testing MongoDB Atlas connection...")
	if err := testAtlasConnection(ctx, client); err != nil {
		log.Fatalf("❌ Failed to ping MongoDB Atlas: %v", err)
	}

	// Set global variables
	Client = client
	DB = client.Database(dbName)

	log.Printf("✅ MongoDB Atlas connected successfully!")
	log.Printf("📍 Database: %s", dbName)

}

// createAtlasClientOptions creates optimized client options for MongoDB Atlas
func createAtlasClientOptions(mongoURI string) *options.ClientOptions {
	// Use MongoDB Atlas Stable API (recommended for Atlas)
	//serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	// Create base client options
	clientOptions := options.Client().
		ApplyURI(mongoURI)
		//.
		//SetServerAPIOptions(serverAPI)

	// Connection Pool Settings (optimized for Atlas)
	maxPoolSize := getEnvUint64("MONGO_MAX_POOL_SIZE", 50)
	minPoolSize := getEnvUint64("MONGO_MIN_POOL_SIZE", 5)
	maxConnIdleTime := getEnvDuration("MONGO_MAX_CONN_IDLE_TIME", 30*time.Minute)

	clientOptions.SetMaxPoolSize(maxPoolSize)
	clientOptions.SetMinPoolSize(minPoolSize)
	clientOptions.SetMaxConnIdleTime(maxConnIdleTime)

	// Timeout Settings (optimized for Atlas)
	connectTimeout := getEnvDuration("MONGO_CONNECT_TIMEOUT", 20*time.Second)
	serverSelectionTimeout := getEnvDuration("MONGO_SERVER_TIMEOUT", 20*time.Second)
	heartbeatInterval := getEnvDuration("MONGO_HEARTBEAT_INTERVAL", 10*time.Second)

	clientOptions.SetConnectTimeout(connectTimeout)
	clientOptions.SetServerSelectionTimeout(serverSelectionTimeout)
	clientOptions.SetHeartbeatInterval(heartbeatInterval)

	// Retry Settings (important for Atlas)
	clientOptions.SetRetryWrites(true)
	clientOptions.SetRetryReads(true)

	// Read Preference (optimized for Atlas)
	readPreference, err := readpref.New(readpref.PrimaryMode)
	if err == nil {
		clientOptions.SetReadPreference(readPreference)
	}

	// Compression (helps with Atlas performance)
	clientOptions.SetCompressors([]string{"snappy", "zlib", "zstd"})

	// Application Name (helps with Atlas monitoring)
	appName := getEnv("MONGO_APP_NAME", "social-media-api")
	clientOptions.SetAppName(appName)

	log.Printf("📊 Connection Pool: Min=%d, Max=%d", minPoolSize, maxPoolSize)
	log.Printf("⏱️  Timeouts: Connect=%v, Selection=%v", connectTimeout, serverSelectionTimeout)

	return clientOptions
}

// testAtlasConnection performs comprehensive connection testing for Atlas
func testAtlasConnection(ctx context.Context, client *mongo.Client) error {
	// Test 1: Basic ping to admin database (Atlas standard)
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		return fmt.Errorf("admin ping failed: %w", err)
	}

	// Test 2: Check if we can list databases (validates permissions)
	if _, err := client.ListDatabaseNames(ctx, bson.D{}); err != nil {
		log.Printf("⚠️  Warning: Cannot list databases (limited permissions): %v", err)
		// This is not necessarily fatal for application databases
	}

	// Test 3: Test target database accessibility
	dbName := getEnv("DB_NAME", "social_media")
	targetDB := client.Database(dbName)

	// Try to run a simple command on target database
	var result bson.M
	err := targetDB.RunCommand(ctx, bson.D{{"dbStats", 1}}).Decode(&result)
	if err != nil {
		log.Printf("⚠️  Warning: Cannot access database stats: %v", err)
		// Try a simpler test - just list collections
		if _, err := targetDB.ListCollectionNames(ctx, bson.D{}); err != nil {
			return fmt.Errorf("cannot access target database '%s': %w", dbName, err)
		}
	}

	return nil
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
