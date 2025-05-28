// migrations/helpers.go
package migrations

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexConflictError represents an index conflict error
type IndexConflictError struct {
	Collection string
	IndexName  string
	Message    string
}

func (e *IndexConflictError) Error() string {
	return e.Message
}

// CreateIndexesSafely creates indexes while handling conflicts
func CreateIndexesSafely(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel) error {
	collectionName := collection.Name()

	for _, index := range indexes {
		if err := createSingleIndexSafely(ctx, collection, index); err != nil {
			log.Printf("Failed to create index on collection %s: %v", collectionName, err)
			return err
		}
	}

	return nil
}

// createSingleIndexSafely creates a single index while handling conflicts
func createSingleIndexSafely(ctx context.Context, collection *mongo.Collection, index mongo.IndexModel) error {
	collectionName := collection.Name()

	// Try to create the index normally first
	_, err := collection.Indexes().CreateOne(ctx, index)
	if err == nil {
		return nil // Success
	}

	// Check if it's an index conflict error
	if strings.Contains(err.Error(), "IndexOptionsConflict") || strings.Contains(err.Error(), "already exists") {
		log.Printf("Index conflict detected on collection %s, attempting to resolve...", collectionName)

		// Get the index name
		indexName := getIndexName(index)
		if indexName == "" {
			return err // Can't resolve without index name
		}

		// Drop the existing conflicting index
		// Drop the existing conflicting index
		_, dropErr := collection.Indexes().DropOne(ctx, indexName)
		if dropErr != nil {
			log.Printf("Failed to drop conflicting index %s on collection %s: %v", indexName, collectionName, dropErr)
			return err // Return original error
		}

		log.Printf("Dropped conflicting index %s on collection %s", indexName, collectionName)

		// Try to create the index again
		_, createErr := collection.Indexes().CreateOne(ctx, index)
		if createErr != nil {
			log.Printf("Failed to recreate index %s on collection %s: %v", indexName, collectionName, createErr)
			return createErr
		}

		log.Printf("Successfully recreated index %s on collection %s", indexName, collectionName)
		return nil
	}

	// For other errors, return as-is
	return err
}

// getIndexName extracts the index name from IndexModel
func getIndexName(index mongo.IndexModel) string {
	if index.Options != nil && index.Options.Name != nil {
		return *index.Options.Name
	}

	// Generate default name from keys
	var parts []string
	if index.Keys != nil {
		doc, ok := index.Keys.(bson.D)
		if ok {
			for _, elem := range doc {
				parts = append(parts, elem.Key+"_"+getDirectionString(elem.Value))
			}
		}
	}

	if len(parts) > 0 {
		return strings.Join(parts, "_")
	}

	return ""
}

// getDirectionString converts index direction to string
func getDirectionString(direction interface{}) string {
	switch v := direction.(type) {
	case int:
		if v == 1 {
			return "1"
		} else if v == -1 {
			return "-1"
		}
	case int32:
		if v == 1 {
			return "1"
		} else if v == -1 {
			return "-1"
		}
	case int64:
		if v == 1 {
			return "1"
		} else if v == -1 {
			return "-1"
		}
	case string:
		return v
	}
	return "1"
}

// DropIndexIfExists drops an index if it exists
func DropIndexIfExists(ctx context.Context, collection *mongo.Collection, indexName string) error {
	// List existing indexes
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var exists bool
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			continue
		}

		if name, ok := index["name"].(string); ok && name == indexName {
			exists = true
			break
		}
	}

	if exists {
		log.Printf("Dropping existing index: %s", indexName)
		_, err := collection.Indexes().DropOne(ctx, indexName)
		return err
	}

	return nil
}

// EnsureTTLIndex ensures a TTL index exists with correct expiration
func EnsureTTLIndex(ctx context.Context, collection *mongo.Collection, field string, expireAfterSeconds int32) error {
	indexName := field + "_1"

	// Drop existing index if it exists
	if err := DropIndexIfExists(ctx, collection, indexName); err != nil {
		log.Printf("Warning: Could not drop existing index %s: %v", indexName, err)
	}

	// Create the TTL index
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{field, 1}},
		Options: options.Index().SetExpireAfterSeconds(expireAfterSeconds),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	log.Printf("Created TTL index on %s.%s with expiration %d seconds", collection.Name(), field, expireAfterSeconds)
	return nil
}

// EnsureUniqueIndex ensures a unique index exists
func EnsureUniqueIndex(ctx context.Context, collection *mongo.Collection, keys bson.D) error {
	// Generate index name
	var parts []string
	for _, elem := range keys {
		parts = append(parts, elem.Key+"_"+getDirectionString(elem.Value))
	}
	indexName := strings.Join(parts, "_")

	// Check if index already exists with correct properties
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var indexExists bool
	var isUnique bool

	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			continue
		}

		if name, ok := index["name"].(string); ok && name == indexName {
			indexExists = true
			if unique, ok := index["unique"].(bool); ok {
				isUnique = unique
			}
			break
		}
	}

	// If index exists and is unique, we're good
	if indexExists && isUnique {
		return nil
	}

	// If index exists but is not unique, drop it
	if indexExists && !isUnique {
		_, err := collection.Indexes().DropOne(ctx, indexName)
		return err
	}

	// Create the unique index
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// RecreateCollectionAsRegular drops a collection if it exists and ensures it's recreated as a regular collection
func RecreateCollectionAsRegular(ctx context.Context, db *mongo.Database, collectionName string) error {
	// Check if collection exists
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if err != nil {
		return err
	}

	// If collection exists, check if it's capped and drop it
	if len(collections) > 0 {
		// Get collection stats to check if it's capped
		var result bson.M
		err := db.RunCommand(ctx, bson.D{
			{"collStats", collectionName},
		}).Decode(&result)

		if err == nil {
			// Check if collection is capped
			if capped, exists := result["capped"]; exists && capped.(bool) {
				log.Printf("Collection %s is capped, dropping it to recreate as regular collection", collectionName)
				if err := db.Collection(collectionName).Drop(ctx); err != nil {
					return fmt.Errorf("failed to drop capped collection %s: %w", collectionName, err)
				}
				log.Printf("Dropped capped collection: %s", collectionName)
			} else {
				log.Printf("Collection %s already exists as regular collection", collectionName)
			}
		} else {
			// If we can't get stats, assume it might be capped and drop it to be safe
			log.Printf("Could not get stats for collection %s, dropping to recreate: %v", collectionName, err)
			if err := db.Collection(collectionName).Drop(ctx); err != nil {
				return fmt.Errorf("failed to drop collection %s: %w", collectionName, err)
			}
			log.Printf("Dropped collection: %s", collectionName)
		}
	}

	// Collection either didn't exist or has been dropped
	// MongoDB will create it as a regular collection when first document is inserted
	// or when first index is created
	log.Printf("Collection %s is ready as regular collection", collectionName)
	return nil
}

// recreateCollectionAsRegular is a wrapper function for the migration
func recreateCollectionAsRegular(ctx context.Context, db *mongo.Database, collectionName string) error {
	return RecreateCollectionAsRegular(ctx, db, collectionName)
}
