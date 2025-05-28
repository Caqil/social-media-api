// scripts/migrate.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"social-media-api/migrations" // Replace with your actual module path
)

// Configuration
type MigrateConfig struct {
	MongoURI     string
	DatabaseName string
	Timeout      time.Duration
}

// Command-line flags
var (
	mongoURI         = flag.String("mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	databaseName     = flag.String("database", "social_media_db", "Database name")
	command          = flag.String("command", "up", "Migration command: up, down, status, create")
	migrationID      = flag.String("migration", "", "Specific migration ID (for down command)")
	newMigrationName = flag.String("name", "", "Name for new migration (for create command)")
	timeout          = flag.Duration("timeout", 30*time.Second, "Operation timeout")
	verbose          = flag.Bool("verbose", false, "Enable verbose logging")
	dryRun           = flag.Bool("dry-run", false, "Show what would be done without executing")
	force            = flag.Bool("force", false, "Force execution without confirmation")
)

func main() {
	flag.Parse()

	// Setup logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	config := MigrateConfig{
		MongoURI:     *mongoURI,
		DatabaseName: *databaseName,
		Timeout:      *timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Connect to MongoDB
	client, err := connectMongoDB(ctx, config.MongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	db := client.Database(config.DatabaseName)

	// Execute command
	switch *command {
	case "up":
		if err := runMigrationsUp(ctx, db); err != nil {
			log.Fatal("Migration up failed:", err)
		}
	case "down":
		if *migrationID == "" {
			log.Fatal("Migration ID is required for down command. Use -migration flag")
		}
		if err := runMigrationDown(ctx, db, *migrationID); err != nil {
			log.Fatal("Migration down failed:", err)
		}
	case "status":
		if err := showMigrationStatus(ctx, db); err != nil {
			log.Fatal("Failed to show migration status:", err)
		}
	case "create":
		if *newMigrationName == "" {
			log.Fatal("Migration name is required for create command. Use -name flag")
		}
		if err := createNewMigration(*newMigrationName); err != nil {
			log.Fatal("Failed to create migration:", err)
		}
	case "reset":
		if err := resetDatabase(ctx, db); err != nil {
			log.Fatal("Failed to reset database:", err)
		}
	case "validate":
		if err := validateMigrations(ctx, db); err != nil {
			log.Fatal("Migration validation failed:", err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		printUsage()
		os.Exit(1)
	}
}

func connectMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
	log.Printf("Connecting to MongoDB: %s", uri)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB")
	return client, nil
}

func runMigrationsUp(ctx context.Context, db *mongo.Database) error {
	log.Println("Running migrations up...")

	if *dryRun {
		log.Println("DRY RUN: Would execute migrations but not making any changes")
		return showPendingMigrations(ctx, db)
	}

	if !*force {
		if !confirmAction("Run all pending migrations?") {
			log.Println("Migration cancelled by user")
			return nil
		}
	}

	runner := migrations.NewMigrationRunner(db)
	allMigrations := migrations.InitializeMigrations()
	runner.RegisterMigrations(allMigrations)

	startTime := time.Now()
	if err := runner.RunMigrations(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("All migrations completed successfully in %v", duration)
	return nil
}

func runMigrationDown(ctx context.Context, db *mongo.Database, migrationID string) error {
	log.Printf("Rolling back migration: %s", migrationID)

	if *dryRun {
		log.Printf("DRY RUN: Would rollback migration %s", migrationID)
		return nil
	}

	if !*force {
		if !confirmAction(fmt.Sprintf("Rollback migration %s?", migrationID)) {
			log.Println("Rollback cancelled by user")
			return nil
		}
	}

	runner := migrations.NewMigrationRunner(db)
	allMigrations := migrations.InitializeMigrations()
	runner.RegisterMigrations(allMigrations)

	startTime := time.Now()
	if err := runner.RollbackMigration(ctx, migrationID); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Migration %s rolled back successfully in %v", migrationID, duration)
	return nil
}

func showMigrationStatus(ctx context.Context, db *mongo.Database) error {
	log.Println("Checking migration status...")

	runner := migrations.NewMigrationRunner(db)
	allMigrations := migrations.InitializeMigrations()
	runner.RegisterMigrations(allMigrations)

	statuses, err := runner.GetMigrationStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if len(statuses) == 0 {
		log.Println("No migrations found")
		return nil
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("==================")
	fmt.Printf("%-30s %-10s %-20s %s\n", "Migration ID", "Status", "Applied At", "Description")
	fmt.Println(string(make([]byte, 100, 100)[:]))

	appliedCount := 0
	pendingCount := 0

	for _, status := range statuses {
		statusStr := "PENDING"
		appliedAtStr := "-"

		if status.Applied {
			statusStr = "APPLIED"
			appliedCount++
			if status.AppliedAt != nil {
				appliedAtStr = status.AppliedAt.Format("2006-01-02 15:04:05")
			}
		} else {
			pendingCount++
		}

		fmt.Printf("%-30s %-10s %-20s %s\n",
			status.ID,
			statusStr,
			appliedAtStr,
			status.Description,
		)
	}

	fmt.Printf("\nSummary: %d applied, %d pending\n", appliedCount, pendingCount)
	return nil
}

func showPendingMigrations(ctx context.Context, db *mongo.Database) error {
	runner := migrations.NewMigrationRunner(db)
	allMigrations := migrations.InitializeMigrations()
	runner.RegisterMigrations(allMigrations)

	statuses, err := runner.GetMigrationStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	pendingMigrations := make([]migrations.MigrationStatus, 0)
	for _, status := range statuses {
		if !status.Applied {
			pendingMigrations = append(pendingMigrations, status)
		}
	}

	if len(pendingMigrations) == 0 {
		log.Println("No pending migrations")
		return nil
	}

	fmt.Printf("\nPending migrations (%d):\n", len(pendingMigrations))
	for _, migration := range pendingMigrations {
		fmt.Printf("  - %s: %s\n", migration.ID, migration.Description)
	}

	return nil
}

func createNewMigration(name string) error {
	log.Printf("Creating new migration: %s", name)

	// Generate migration ID (timestamp + name)
	timestamp := time.Now().Format("20060102150405")
	migrationID := fmt.Sprintf("%s_%s", timestamp, name)
	filename := fmt.Sprintf("migrations/%s.go", migrationID)

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("migration file already exists: %s", filename)
	}

	// Create migration template
	template := fmt.Sprintf(`// migrations/%s.go
package migrations

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get%sMigration returns the %s migration
func Get%sMigration() Migration {
	return Migration{
		ID:          "%s",
		Description: "%s",
		Up:          %sUp,
		Down:        %sDown,
	}
}

func %sUp(ctx context.Context, db *mongo.Database) error {
	log.Println("Running %s migration up...")
	
	// TODO: Implement migration logic here
	// Example:
	// collection := db.Collection("your_collection")
	// _, err := collection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"new_field": "default_value"}})
	// return err
	
	log.Println("%s migration up completed")
	return nil
}

func %sDown(ctx context.Context, db *mongo.Database) error {
	log.Println("Running %s migration down...")
	
	// TODO: Implement rollback logic here
	// Example:
	// collection := db.Collection("your_collection")
	// _, err := collection.UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"new_field": ""}})
	// return err
	
	log.Println("%s migration down completed")
	return nil
}
`, migrationID, name, toPascalCase(name), migrationID, name, toCamelCase(name), toCamelCase(name), toCamelCase(name), name, name, toCamelCase(name), name, name)

	// Write file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(template); err != nil {
		return fmt.Errorf("failed to write migration template: %w", err)
	}

	log.Printf("Migration created successfully: %s", filename)
	log.Println("Don't forget to:")
	log.Printf("1. Edit %s to implement your migration logic", filename)
	log.Printf("2. Add Get%sMigration() to InitializeMigrations() in migration_runner.go", toPascalCase(name))
	return nil
}

func resetDatabase(ctx context.Context, db *mongo.Database) error {
	log.Println("WARNING: This will drop all collections and reset the database!")

	if !*force {
		if !confirmAction("Are you absolutely sure you want to reset the database? This cannot be undone!") {
			log.Println("Database reset cancelled by user")
			return nil
		}
	}

	if *dryRun {
		log.Println("DRY RUN: Would drop all collections and reset database")
		return nil
	}

	// Get all collection names
	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	log.Printf("Found %d collections to drop", len(collections))

	// Drop all collections
	for _, collectionName := range collections {
		log.Printf("Dropping collection: %s", collectionName)
		if err := db.Collection(collectionName).Drop(ctx); err != nil {
			log.Printf("Warning: Failed to drop collection %s: %v", collectionName, err)
		}
	}

	log.Println("Database reset completed")
	return nil
}

func validateMigrations(ctx context.Context, db *mongo.Database) error {
	log.Println("Validating migrations...")

	runner := migrations.NewMigrationRunner(db)
	allMigrations := migrations.InitializeMigrations()
	runner.RegisterMigrations(allMigrations)

	statuses, err := runner.GetMigrationStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Check for duplicate migration IDs
	seen := make(map[string]bool)
	duplicates := make([]string, 0)

	for _, migration := range allMigrations {
		if seen[migration.ID] {
			duplicates = append(duplicates, migration.ID)
		}
		seen[migration.ID] = true
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate migration IDs found: %v", duplicates)
	}

	// Check migration ordering
	for i := 1; i < len(allMigrations); i++ {
		if allMigrations[i-1].ID >= allMigrations[i].ID {
			return fmt.Errorf("migrations are not in chronological order: %s should come before %s",
				allMigrations[i].ID, allMigrations[i-1].ID)
		}
	}

	// Check for missing Down functions
	missingDown := make([]string, 0)
	for _, migration := range allMigrations {
		if migration.Down == nil {
			missingDown = append(missingDown, migration.ID)
		}
	}

	if len(missingDown) > 0 {
		log.Printf("Warning: Migrations without Down function: %v", missingDown)
	}

	// Summary
	fmt.Printf("\nValidation Results:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total migrations: %d\n", len(allMigrations))
	fmt.Printf("Applied migrations: %d\n", countAppliedMigrations(statuses))
	fmt.Printf("Pending migrations: %d\n", len(allMigrations)-countAppliedMigrations(statuses))
	fmt.Printf("Duplicate IDs: %d\n", len(duplicates))
	fmt.Printf("Missing Down functions: %d\n", len(missingDown))

	if len(duplicates) == 0 {
		log.Println("✅ All migrations are valid")
	} else {
		return fmt.Errorf("❌ Migration validation failed")
	}

	return nil
}

func confirmAction(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "YES"
}

func countAppliedMigrations(statuses []migrations.MigrationStatus) int {
	count := 0
	for _, status := range statuses {
		if status.Applied {
			count++
		}
	}
	return count
}

func toPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Simple implementation - just capitalize first letter
	// In production, you might want a more sophisticated implementation
	result := string(s[0]-32) + s[1:] // Convert first char to uppercase
	return result
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Simple implementation - just lowercase first letter
	result := string(s[0]+32) + s[1:] // Convert first char to lowercase
	return result
}

func printUsage() {
	fmt.Println("MongoDB Migration Tool")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate [options] -command=<command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up      - Run all pending migrations")
	fmt.Println("  down    - Rollback a specific migration (requires -migration flag)")
	fmt.Println("  status  - Show migration status")
	fmt.Println("  create  - Create a new migration file (requires -name flag)")
	fmt.Println("  reset   - Drop all collections and reset database")
	fmt.Println("  validate - Validate migration consistency")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -mongo-uri string     MongoDB connection URI (default: mongodb://localhost:27017)")
	fmt.Println("  -database string      Database name (default: social_media_db)")
	fmt.Println("  -migration string     Migration ID for down command")
	fmt.Println("  -name string          Name for new migration")
	fmt.Println("  -timeout duration     Operation timeout (default: 30s)")
	fmt.Println("  -verbose              Enable verbose logging")
	fmt.Println("  -dry-run              Show what would be done without executing")
	fmt.Println("  -force                Force execution without confirmation")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrate -command=up")
	fmt.Println("  migrate -command=down -migration=001_initial_indexes")
	fmt.Println("  migrate -command=status")
	fmt.Println("  migrate -command=create -name=add_user_preferences")
	fmt.Println("  migrate -command=reset -force")
	fmt.Println("  migrate -command=validate")
}
