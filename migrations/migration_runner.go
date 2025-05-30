// migrations/migration_runner.go
package migrations

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration represents a database migration
type Migration struct {
	ID          string
	Description string
	Up          func(ctx context.Context, db *mongo.Database) error
	Down        func(ctx context.Context, db *mongo.Database) error
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Version   string             `bson:"version"`
	Applied   bool               `bson:"applied"`
	AppliedAt time.Time          `bson:"applied_at"`
	CreatedAt time.Time          `bson:"created_at"`
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db         *mongo.Database
	migrations []Migration
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *mongo.Database) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrations: []Migration{},
	}
}

// RegisterMigration registers a migration
func (mr *MigrationRunner) RegisterMigration(migration Migration) {
	mr.migrations = append(mr.migrations, migration)
}

// RegisterMigrations registers multiple migrations
func (mr *MigrationRunner) RegisterMigrations(migrations []Migration) {
	for _, migration := range migrations {
		mr.RegisterMigration(migration)
	}
}

// RunMigrations executes all pending migrations
func (mr *MigrationRunner) RunMigrations(ctx context.Context) error {
	log.Println("Starting database migrations...")

	// Ensure migrations collection exists
	if err := mr.ensureMigrationsCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations collection: %w", err)
	}

	// Sort migrations by ID
	sort.Slice(mr.migrations, func(i, j int) bool {
		return mr.migrations[i].ID < mr.migrations[j].ID
	})

	// Get applied migrations
	appliedMigrations, err := mr.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Execute pending migrations
	for _, migration := range mr.migrations {
		if _, applied := appliedMigrations[migration.ID]; applied {
			log.Printf("Migration %s already applied, skipping...", migration.ID)
			continue
		}

		log.Printf("Running migration %s: %s", migration.ID, migration.Description)

		if err := mr.runMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.ID, err)
		}

		log.Printf("Migration %s completed successfully", migration.ID)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// RollbackMigration rolls back a specific migration
func (mr *MigrationRunner) RollbackMigration(ctx context.Context, migrationID string) error {
	log.Printf("Rolling back migration %s...", migrationID)

	// Find the migration
	var targetMigration *Migration
	for _, migration := range mr.migrations {
		if migration.ID == migrationID {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found", migrationID)
	}

	// Check if migration is applied
	appliedMigrations, err := mr.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if _, applied := appliedMigrations[migrationID]; !applied {
		return fmt.Errorf("migration %s is not applied", migrationID)
	}

	// Execute rollback
	if targetMigration.Down == nil {
		return fmt.Errorf("migration %s does not support rollback", migrationID)
	}

	if err := targetMigration.Down(ctx, mr.db); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	// Remove migration record
	if err := mr.removeMigrationRecord(ctx, migrationID); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	log.Printf("Migration %s rolled back successfully", migrationID)
	return nil
}

// GetMigrationStatus returns the status of all migrations
func (mr *MigrationRunner) GetMigrationStatus(ctx context.Context) ([]MigrationStatus, error) {
	appliedMigrations, err := mr.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var statuses []MigrationStatus
	for _, migration := range mr.migrations {
		status := MigrationStatus{
			ID:          migration.ID,
			Description: migration.Description,
			Applied:     false,
		}

		if record, applied := appliedMigrations[migration.ID]; applied {
			status.Applied = true
			status.AppliedAt = &record.AppliedAt
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Applied     bool       `json:"applied"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
}

// Private helper methods

func (mr *MigrationRunner) ensureMigrationsCollection(ctx context.Context) error {
	// Create index on version field
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"version", 1}},
		Options: options.Index().SetUnique(true),
	}

	collection := mr.db.Collection("migrations")
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

func (mr *MigrationRunner) getAppliedMigrations(ctx context.Context) (map[string]MigrationRecord, error) {
	collection := mr.db.Collection("migrations")
	cursor, err := collection.Find(ctx, bson.M{"applied": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	appliedMigrations := make(map[string]MigrationRecord)
	for cursor.Next(ctx) {
		var record MigrationRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		appliedMigrations[record.Version] = record
	}

	return appliedMigrations, cursor.Err()
}

func (mr *MigrationRunner) runMigration(ctx context.Context, migration Migration) error {
	// Execute migration
	if err := migration.Up(ctx, mr.db); err != nil {
		return err
	}

	// Record migration as applied
	return mr.recordMigration(ctx, migration.ID)
}

func (mr *MigrationRunner) recordMigration(ctx context.Context, migrationID string) error {
	collection := mr.db.Collection("migrations")

	record := MigrationRecord{
		Version:   migrationID,
		Applied:   true,
		AppliedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx, record)
	return err
}

func (mr *MigrationRunner) removeMigrationRecord(ctx context.Context, migrationID string) error {
	collection := mr.db.Collection("migrations")
	_, err := collection.DeleteOne(ctx, bson.M{"version": migrationID})
	return err
}

// InitializeMigrations initializes and returns all available migrations
func InitializeMigrations() []Migration {
	return []Migration{
		GetInitialIndexesMigration(),
		GetSocialFeaturesMigration(),
		CreateAdminUser001(),
	}
}

// RunAllMigrations is a convenience function to run all migrations
func RunAllMigrations(ctx context.Context, db *mongo.Database) error {
	runner := NewMigrationRunner(db)
	migrations := InitializeMigrations()
	runner.RegisterMigrations(migrations)
	return runner.RunMigrations(ctx)
}

// CreateMigrationTemplate creates a template for a new migration
func CreateMigrationTemplate(id, description string) Migration {
	return Migration{
		ID:          id,
		Description: description,
		Up: func(ctx context.Context, db *mongo.Database) error {
			// TODO: Implement migration logic
			log.Printf("Running migration %s", id)
			return nil
		},
		Down: func(ctx context.Context, db *mongo.Database) error {
			// TODO: Implement rollback logic
			log.Printf("Rolling back migration %s", id)
			return nil
		},
	}
}
