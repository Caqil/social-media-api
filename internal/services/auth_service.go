// internal/services/auth_service.go
package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userCollection    *mongo.Collection
	sessionCollection *mongo.Collection
	db                *mongo.Database
	jwtSecret         string
	refreshSecret     string
}

type LoginResponse struct {
	User         models.UserResponse `json:"user"`
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresIn    int64               `json:"expires_in"`
	TokenType    string              `json:"token_type"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Session struct {
	models.BaseModel `bson:",inline"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	SessionID        string             `json:"session_id" bson:"session_id"`
	DeviceInfo       string             `json:"device_info" bson:"device_info"`
	IPAddress        string             `json:"ip_address" bson:"ip_address"`
	UserAgent        string             `json:"user_agent" bson:"user_agent"`
	IsActive         bool               `json:"is_active" bson:"is_active"`
	LastActivityAt   time.Time          `json:"last_activity_at" bson:"last_activity_at"`
	ExpiresAt        time.Time          `json:"expires_at" bson:"expires_at"`
}

func NewAuthService(jwtSecret, refreshSecret string) *AuthService {
	return &AuthService{
		userCollection:    config.DB.Collection("users"),
		sessionCollection: config.DB.Collection("sessions"),
		db:                config.DB,
		jwtSecret:         jwtSecret,
		refreshSecret:     refreshSecret,
	}
}

// Login authenticates user and returns tokens
func (as *AuthService) Login(req models.LoginRequest) (*LoginResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email or username
	var user models.User
	filter := bson.M{
		"$or": []bson.M{
			{"email": req.EmailOrUsername},
			{"username": req.EmailOrUsername},
		},
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	err := as.userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is suspended
	if user.IsSuspended {
		return nil, errors.New("account is suspended")
	}

	// Create session
	sessionID := primitive.NewObjectID().Hex()
	session := &Session{
		UserID:         user.ID,
		SessionID:      sessionID,
		DeviceInfo:     req.DeviceInfo,
		IPAddress:      "", // This would be set by the handler
		IsActive:       true,
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour), // 30 days
	}
	session.BeforeCreate()

	_, err = as.sessionCollection.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := as.GenerateTokens(&user, sessionID, req.DeviceInfo, "")
	if err != nil {
		return nil, err
	}

	// Update user's last login
	as.UpdateUserLogin(user.ID, req.DeviceInfo)

	return &LoginResponse{
		User:         user.ToUserResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 60 * 60, // 24 hours in seconds
		TokenType:    "Bearer",
	}, nil
}

// Register creates a new user account
func (as *AuthService) Register(req models.RegisterRequest) (*LoginResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	exists, err := as.CheckUserExists(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username or email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Username:    req.Username,
		Email:       req.Email,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		Phone:       req.Phone,
	}

	user.BeforeCreate()

	result, err := as.userCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	// Create session
	sessionID := primitive.NewObjectID().Hex()
	session := &Session{
		UserID:         user.ID,
		SessionID:      sessionID,
		IsActive:       true,
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
	}
	session.BeforeCreate()

	_, err = as.sessionCollection.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := as.GenerateTokens(user, sessionID, "", "")
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:         user.ToUserResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 60 * 60,
		TokenType:    "Bearer",
	}, nil
}

// RefreshTokens refreshes access and refresh tokens
func (as *AuthService) RefreshTokens(refreshToken string) (*RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := as.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get session ID from claims
	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return nil, errors.New("invalid session ID in token")
	}

	// Get device info and IP address from claims
	deviceInfo, _ := claims["device_info"].(string)
	ipAddress, _ := claims["ip_address"].(string)

	// Get user
	user, err := as.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if !user.IsActive || user.IsSuspended {
		return nil, errors.New("account is suspended or inactive")
	}

	// Validate session
	session, err := as.GetSession(sessionID)
	if err != nil || !session.IsActive {
		return nil, errors.New("invalid session")
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, err := as.GenerateTokens(user, sessionID, deviceInfo, ipAddress)
	if err != nil {
		return nil, err
	}

	// Update session activity
	as.UpdateSessionActivity(sessionID)

	return &RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    24 * 60 * 60,
		TokenType:    "Bearer",
	}, nil
}

// Logout invalidates user session
func (as *AuthService) Logout(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err := as.sessionCollection.UpdateOne(ctx, bson.M{"session_id": sessionID}, update)
	return err
}

// LogoutAll invalidates all user sessions
func (as *AuthService) LogoutAll(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err := as.sessionCollection.UpdateMany(ctx, bson.M{"user_id": userID}, update)
	return err
}

// ForgotPassword initiates password reset process
func (as *AuthService) ForgotPassword(req models.ForgotPasswordRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email
	var user models.User
	err := as.userCollection.FindOne(ctx, bson.M{
		"email":      req.Email,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil // Don't reveal if email exists
		}
		return err
	}

	// Generate reset token
	bytes := make([]byte, 64)
	_, err = rand.Read(bytes)
	if err != nil {
		return err
	}
	resetToken := base64.URLEncoding.EncodeToString(bytes)[:64]
	expiryTime := time.Now().Add(1 * time.Hour) // 1 hour expiry

	// Update user with reset token
	update := bson.M{
		"$set": bson.M{
			"password_reset_token":  resetToken,
			"password_reset_expiry": expiryTime,
			"updated_at":            time.Now(),
		},
	}

	_, err = as.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	// Send reset email (this would integrate with email service)
	// emailService.SendPasswordResetEmail(user.Email, resetToken)

	return nil
}

// ResetPassword resets user password using token
func (as *AuthService) ResetPassword(req models.ResetPasswordRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate password confirmation
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("passwords do not match")
	}

	// Find user by reset token
	var user models.User
	err := as.userCollection.FindOne(ctx, bson.M{
		"password_reset_token": req.Token,
		"password_reset_expiry": bson.M{
			"$gt": time.Now(),
		},
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("invalid or expired reset token")
		}
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	update := bson.M{
		"$set": bson.M{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		},
		"$unset": bson.M{
			"password_reset_token":  "",
			"password_reset_expiry": "",
		},
	}

	_, err = as.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	// Invalidate all existing sessions
	as.LogoutAll(user.ID)

	return nil
}

// VerifyEmail verifies user's email using token
func (as *AuthService) VerifyEmail(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find user by email verification token
	var user models.User
	err := as.userCollection.FindOne(ctx, bson.M{
		"email_verify_token": token,
		"is_active":          true,
		"deleted_at":         bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("invalid verification token")
		}
		return err
	}

	// Update user as verified
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"email_verified":    true,
			"email_verified_at": now,
			"updated_at":        now,
		},
		"$unset": bson.M{
			"email_verify_token": "",
		},
	}

	_, err = as.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	return err
}

// GenerateTokens generates access and refresh tokens
func (as *AuthService) GenerateTokens(user *models.User, sessionID, deviceInfo, ipAddress string) (string, string, error) {
	now := time.Now()

	// Access token claims
	accessClaims := jwt.MapClaims{
		"user_id":     user.ID.Hex(),
		"username":    user.Username,
		"email":       user.Email,
		"role":        user.Role,
		"session_id":  sessionID,
		"device_info": deviceInfo,
		"ip_address":  ipAddress,
		"token_type":  "access",
		"iat":         now.Unix(),
		"exp":         now.Add(24 * time.Hour).Unix(), // 24 hours
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(as.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token claims
	refreshClaims := jwt.MapClaims{
		"user_id":     user.ID.Hex(),
		"username":    user.Username,
		"email":       user.Email,
		"role":        user.Role,
		"session_id":  sessionID,
		"device_info": deviceInfo,
		"ip_address":  ipAddress,
		"token_type":  "refresh",
		"iat":         now.Unix(),
		"exp":         now.Add(30 * 24 * time.Hour).Unix(), // 30 days
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(as.refreshSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateAccessToken validates access token
func (as *AuthService) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(as.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "access" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates refresh token
func (as *AuthService) ValidateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(as.refreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "refresh" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID gets user by ID
func (as *AuthService) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := as.userCollection.FindOne(ctx, bson.M{
		"_id":        userID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetSession gets session by ID
func (as *AuthService) GetSession(sessionID string) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session Session
	err := as.sessionCollection.FindOne(ctx, bson.M{
		"session_id": sessionID,
		"is_active":  true,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&session)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateUserLogin updates user's last login information
func (as *AuthService) UpdateUserLogin(userID primitive.ObjectID, deviceInfo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"last_login_at":    now,
			"last_active_at":   now,
			"last_device_info": deviceInfo,
			"online_status":    "online",
			"updated_at":       now,
		},
	}

	_, err := as.userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

// UpdateSessionActivity updates session's last activity
func (as *AuthService) UpdateSessionActivity(sessionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"last_activity_at": time.Now(),
			"updated_at":       time.Now(),
		},
	}

	_, err := as.sessionCollection.UpdateOne(ctx, bson.M{"session_id": sessionID}, update)
	return err
}

// GetUserSessions gets all active sessions for a user
func (as *AuthService) GetUserSessions(userID primitive.ObjectID) ([]Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_active":  true,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	cursor, err := as.sessionCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// RevokeSession revokes a specific session
func (as *AuthService) RevokeSession(sessionID string) error {
	return as.Logout(sessionID)
}

// CheckUserExists checks if username or email already exists
func (as *AuthService) CheckUserExists(username, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": email},
		},
		"deleted_at": bson.M{"$exists": false},
	}

	count, err := as.userCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CleanupExpiredSessions removes expired sessions
func (as *AuthService) CleanupExpiredSessions() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"is_active": false},
		},
	}

	_, err := as.sessionCollection.DeleteMany(ctx, filter)
	return err
}
