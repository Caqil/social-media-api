package handlers

import (
	"context"
	"net/http"
	"social-media-api/config"
	"social-media-api/models"
	"social-media-api/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var validate = validator.New()

// RegisterUser handles user registration
func RegisterUser(c *gin.Context) {
	var req models.RegisterRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Validate request data
	if err := validate.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get users collection
	userCollection := config.DB.Collection("users")

	// Check if username already exists
	var existingUser models.User
	err := userCollection.FindOne(ctx, bson.M{
		"username":   req.Username,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Username already exists", "")
		return
	} else if err != mongo.ErrNoDocuments {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	// Check if email already exists
	err = userCollection.FindOne(ctx, bson.M{
		"email":      req.Email,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email already exists", "")
		return
	} else if err != mongo.ErrNoDocuments {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process password", "")
		return
	}

	// Create new user
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Bio:       req.Bio,
	}

	// Set default values
	user.BeforeCreate()

	// Insert user into database
	result, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	// Set the ID from the insert result
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Return success response
	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", user.ToUserResponse())
}

// GetUserProfile handles getting user profile by ID
func GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("id")

	// Convert string ID to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get users collection
	userCollection := config.DB.Collection("users")

	// Find user by ID (excluding soft deleted users)
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id":        userID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found", "")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User profile retrieved successfully", user.ToUserResponse())
}

// GetUserByUsername handles getting user profile by username
func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get users collection
	userCollection := config.DB.Collection("users")

	// Find user by username (excluding soft deleted users)
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{
		"username":   username,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found", "")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User profile retrieved successfully", user.ToUserResponse())
}
