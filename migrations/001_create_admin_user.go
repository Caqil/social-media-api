// migrations/001_create_admin_user.go
package migrations

import (
	"context"

	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateAdminUser001() Migration {
	return Migration{
		ID:          "001_create_admin_user",
		Description: "Create default admin user",
		Up: func(ctx context.Context, db *mongo.Database) error {
			collection := db.Collection("users")

			// Check if admin user already exists
			count, err := collection.CountDocuments(ctx, bson.M{
				"role": bson.M{"$in": []string{"admin", "super_admin"}},
			})
			if err != nil {
				return err
			}

			if count > 0 {
				// Admin user already exists
				return nil
			}

			// Hash the default password
			hashedPassword, err := utils.HashPassword("admin123!@#")
			if err != nil {
				return err
			}

			// Create admin user
			adminUser := models.User{
				Username:      "admin",
				Email:         "admin@example.com",
				Password:      hashedPassword,
				FirstName:     "System",
				LastName:      "Admin",
				DisplayName:   "System Administrator",
				Role:          models.RoleSuperAdmin, // Give super admin role
				IsActive:      true,
				IsVerified:    true,
				EmailVerified: true,
			}

			adminUser.BeforeCreate()

			_, err = collection.InsertOne(ctx, adminUser)
			if err != nil {
				return err
			}

			return nil
		},
		Down: func(ctx context.Context, db *mongo.Database) error {
			collection := db.Collection("users")
			_, err := collection.DeleteOne(ctx, bson.M{
				"email": "admin@example.com",
			})
			return err
		},
	}
}
