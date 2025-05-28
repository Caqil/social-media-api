// middleware/auth.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"social-media-api/internal/models"
	"social-media-api/internal/utils"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID     string          `json:"user_id"`
	Username   string          `json:"username"`
	Email      string          `json:"email"`
	Role       models.UserRole `json:"role"`
	SessionID  string          `json:"session_id"`
	DeviceInfo string          `json:"device_info,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
	IssuedAt   int64           `json:"iat"`
	ExpiresAt  int64           `json:"exp"`
	TokenType  string          `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	db            *mongo.Database
	jwtSecret     []byte
	refreshSecret []byte
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(db *mongo.Database, jwtSecret, refreshSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		db:            db,
		jwtSecret:     []byte(jwtSecret),
		refreshSecret: []byte(refreshSecret),
	}
}

// RequireAuth middleware that requires valid JWT token
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		claims, err := am.validateToken(token, am.jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Check if token type is access token
		if claims.TokenType != "access" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token type", nil)
			c.Abort()
			return
		}

		// Get user from database to ensure account is still active
		user, err := am.getUserFromDB(claims.UserID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", nil)
			c.Abort()
			return
		}

		// Check if user account is active
		if !user.IsActive || user.IsSuspended {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Account suspended or inactive", nil)
			c.Abort()
			return
		}

		// Update user's last active time
		go am.updateUserActivity(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Set("user_role", user.Role)
		c.Set("session_id", claims.SessionID)

		c.Next()
	})
}

// OptionalAuth middleware that allows both authenticated and unauthenticated requests
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		claims, err := am.validateToken(token, am.jwtSecret)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Check if token type is access token
		if claims.TokenType != "access" {
			c.Next()
			return
		}

		// Get user from database
		user, err := am.getUserFromDB(claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		// Check if user account is active
		if !user.IsActive || user.IsSuspended {
			c.Next()
			return
		}

		// Update user's last active time
		go am.updateUserActivity(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Set("user_role", user.Role)
		c.Set("session_id", claims.SessionID)

		c.Next()
	})
}

// RequireRole middleware that requires specific user roles
func (am *AuthMiddleware) RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First require authentication
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Role information not found", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)

		// Check if user has required role
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", nil)
		c.Abort()
	})
}

// RequireVerified middleware that requires verified users
func (am *AuthMiddleware) RequireVerified() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First require authentication
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		user, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "User information not found", nil)
			c.Abort()
			return
		}

		u := user.(*models.User)
		if !u.IsVerified {
			utils.ErrorResponse(c, http.StatusForbidden, "Email verification required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireEmailVerified middleware that requires email verification
func (am *AuthMiddleware) RequireEmailVerified() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First require authentication
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		user, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "User information not found", nil)
			c.Abort()
			return
		}

		u := user.(*models.User)
		if !u.EmailVerified {
			utils.ErrorResponse(c, http.StatusForbidden, "Email verification required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RefreshToken middleware for handling refresh token requests
func (am *AuthMiddleware) RefreshToken() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		refreshToken := am.extractToken(c)
		if refreshToken == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Refresh token required", nil)
			c.Abort()
			return
		}

		claims, err := am.validateToken(refreshToken, am.refreshSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token", err.Error())
			c.Abort()
			return
		}

		// Check if token type is refresh token
		if claims.TokenType != "refresh" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token type", nil)
			c.Abort()
			return
		}

		// Get user from database
		user, err := am.getUserFromDB(claims.UserID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", nil)
			c.Abort()
			return
		}

		// Check if user account is active
		if !user.IsActive || user.IsSuspended {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Account suspended or inactive", nil)
			c.Abort()
			return
		}

		// Set user info in context for token generation
		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Set("user_role", user.Role)
		c.Set("session_id", claims.SessionID)

		c.Next()
	})
}

// BlockSuspended middleware that blocks suspended users
func (am *AuthMiddleware) BlockSuspended() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}

		u := user.(*models.User)
		if u.IsSuspended {
			utils.ErrorResponse(c, http.StatusForbidden, "Account is suspended", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RateLimitByUser middleware that applies rate limiting per user
func (am *AuthMiddleware) RateLimitByUser(maxRequests int, window time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userObjID := userID.(primitive.ObjectID)
		key := fmt.Sprintf("rate_limit_user_%s", userObjID.Hex())

		// Implement rate limiting logic here (Redis recommended)
		// For now, continue without blocking
		c.Next()
	})
}

// Helper methods

// extractToken extracts JWT token from request headers
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try X-Access-Token header
	if token := c.GetHeader("X-Access-Token"); token != "" {
		return token
	}

	// Try query parameter
	if token := c.Query("token"); token != "" {
		return token
	}

	// Try cookie
	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		return token
	}

	return ""
}

// validateToken validates JWT token and returns claims
func (am *AuthMiddleware) validateToken(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if token is expired
		if time.Now().Unix() > claims.ExpiresAt {
			return nil, fmt.Errorf("token is expired")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// getUserFromDB retrieves user from database
func (am *AuthMiddleware) getUserFromDB(userID string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = am.db.Collection("users").FindOne(context.Background(), bson.M{
		"_id":        objID,
		"deleted_at": nil,
	}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// updateUserActivity updates user's last activity
func (am *AuthMiddleware) updateUserActivity(userID primitive.ObjectID, ipAddress, userAgent string) {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"last_active_at":   now,
			"last_device_info": userAgent,
			"online_status":    "online",
			"updated_at":       now,
		},
	}

	am.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		update,
	)
}

// GenerateTokens generates access and refresh tokens
func (am *AuthMiddleware) GenerateTokens(user *models.User, sessionID, deviceInfo, ipAddress string) (string, string, error) {
	now := time.Now()

	// Access token (short-lived)
	accessClaims := &JWTClaims{
		UserID:     user.ID.Hex(),
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		SessionID:  sessionID,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(24 * time.Hour).Unix(), // 24 hours
		TokenType:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   user.ID.Hex(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(am.jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh token (long-lived)
	refreshClaims := &JWTClaims{
		UserID:     user.ID.Hex(),
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		SessionID:  sessionID,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(30 * 24 * time.Hour).Unix(), // 30 days
		TokenType:  "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID + "_refresh",
			Subject:   user.ID.Hex(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(am.refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateTokenString validates a token string and returns claims
func (am *AuthMiddleware) ValidateTokenString(tokenString string) (*JWTClaims, error) {
	return am.validateToken(tokenString, am.jwtSecret)
}

// GetCurrentUser gets current user from context
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	return user.(*models.User), true
}

// GetCurrentUserID gets current user ID from context
func GetCurrentUserID(c *gin.Context) (primitive.ObjectID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return primitive.NilObjectID, false
	}
	return userID.(primitive.ObjectID), true
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// HasRole checks if current user has specific role
func HasRole(c *gin.Context, role models.UserRole) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return userRole.(models.UserRole) == role
}

// HasAnyRole checks if current user has any of the specified roles
func HasAnyRole(c *gin.Context, roles ...models.UserRole) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	currentRole := userRole.(models.UserRole)
	for _, role := range roles {
		if currentRole == role {
			return true
		}
	}
	return false
}

// IsOwnerOrAdmin checks if current user is owner of resource or admin
func IsOwnerOrAdmin(c *gin.Context, resourceOwnerID primitive.ObjectID) bool {
	currentUserID, exists := GetCurrentUserID(c)
	if !exists {
		return false
	}

	// Check if user is owner
	if currentUserID == resourceOwnerID {
		return true
	}

	// Check if user is admin or super admin
	return HasAnyRole(c, models.RoleAdmin, models.RoleSuperAdmin)
}
